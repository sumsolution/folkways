package settings

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"strings"
)

// Object is a JSON object that keeps its key order, as a JS object does (JSON.parse, then JSON.stringify).
// Values are nil, bool, float64, string, []any or *Object.
type Object struct {
	Keys []string
	vals map[string]any
}

// Get returns the value at key and whether the key is set.
func (o *Object) Get(key string) (any, bool) {
	v, ok := o.vals[key]
	return v, ok
}

// Set sets key, appending it if it is new; an existing key keeps its place.
func (o *Object) Set(key string, v any) {
	if o.vals == nil {
		o.vals = map[string]any{}
	}
	if _, ok := o.vals[key]; !ok {
		o.Keys = append(o.Keys, key)
	}
	o.vals[key] = v
}

// Delete removes key.
func (o *Object) Delete(key string) {
	if _, ok := o.vals[key]; !ok {
		return
	}
	delete(o.vals, key)
	for i, k := range o.Keys {
		if k == key {
			o.Keys = append(o.Keys[:i:i], o.Keys[i+1:]...)
			break
		}
	}
}

func (o *Object) clone() *Object { return clone(o).(*Object) }

func clone(v any) any {
	switch v := v.(type) {
	case *Object:
		out := &Object{}
		for _, k := range v.Keys {
			out.Set(k, clone(v.vals[k]))
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, x := range v {
			out[i] = clone(x)
		}
		return out
	}
	return v
}

// Parse decodes JSON like JSON.parse: objects as *Object in key order, numbers as float64. A repeated key
// keeps its first place and its last value.
func Parse(data []byte) (any, error) {
	if !json.Valid(data) {
		return nil, errors.New("invalid JSON")
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	return decode(dec)
}

func decode(dec *json.Decoder) (any, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	switch t := tok.(type) {
	case json.Delim:
		if t == '[' {
			out := []any{}
			for dec.More() {
				v, err := decode(dec)
				if err != nil {
					return nil, err
				}
				out = append(out, v)
			}
			_, err := dec.Token()
			return out, err
		}
		out := &Object{}
		for dec.More() {
			k, err := dec.Token()
			if err != nil {
				return nil, err
			}
			v, err := decode(dec)
			if err != nil {
				return nil, err
			}
			out.Set(k.(string), v)
		}
		_, err := dec.Token()
		return out, err
	case json.Number:
		f, _ := strconv.ParseFloat(string(t), 64) // out of range is ±Inf, as in JS
		return f, nil
	}
	return tok, nil // nil, bool, string
}

// Stringify encodes v like JSON.stringify(v, null, 2).
func Stringify(v any) string {
	var b strings.Builder
	write(&b, v, "")
	return b.String()
}

func write(b *strings.Builder, v any, indent string) {
	inner := indent + "  "
	switch v := v.(type) {
	case nil:
		b.WriteString("null")
	case bool:
		b.WriteString(strconv.FormatBool(v))
	case float64:
		b.WriteString(jsNumber(v))
	case string:
		quote(b, v)
	case []any:
		if len(v) == 0 {
			b.WriteString("[]")
			return
		}
		b.WriteString("[\n")
		for i, x := range v {
			if i > 0 {
				b.WriteString(",\n")
			}
			b.WriteString(inner)
			write(b, x, inner)
		}
		b.WriteString("\n" + indent + "]")
	case *Object:
		if len(v.Keys) == 0 {
			b.WriteString("{}")
			return
		}
		b.WriteString("{\n")
		for i, k := range v.Keys {
			if i > 0 {
				b.WriteString(",\n")
			}
			b.WriteString(inner)
			quote(b, k)
			b.WriteString(": ")
			write(b, v.vals[k], inner)
		}
		b.WriteString("\n" + indent + "}")
	}
}

// jsNumber formats f as JS's Number#toString does.
func jsNumber(f float64) string {
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return "null"
	}
	if f == 0 {
		return "0"
	}
	if a := math.Abs(f); a >= 1e-6 && a < 1e21 {
		return strconv.FormatFloat(f, 'f', -1, 64)
	}
	s := strconv.FormatFloat(f, 'e', -1, 64) // 1e+21, 1.5e-07
	mant, exp, _ := strings.Cut(s, "e")
	sign, digits := exp[:1], strings.TrimLeft(exp[1:], "0")
	return mant + "e" + sign + digits
}

// quote writes s as JSON.stringify does: only ", \ and control characters are escaped.
func quote(b *strings.Builder, s string) {
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 0x20 {
				b.WriteString(`\u00`)
				b.WriteString(strconv.FormatInt(int64(r>>4), 16))
				b.WriteString(strconv.FormatInt(int64(r&0xf), 16))
			} else {
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
}
