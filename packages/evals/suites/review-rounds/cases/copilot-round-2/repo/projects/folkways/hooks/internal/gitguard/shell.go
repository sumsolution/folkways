package gitguard

import "strings"

// A deliberately small shell lexer: enough to split chains and read quoted words, not to run a shell.

// skipSingle returns the index after the single-quoted string starting at s[i].
func skipSingle(s string, i int) int {
	if j := strings.IndexByte(s[i+1:], '\''); j >= 0 {
		return i + 1 + j + 1
	}
	return len(s)
}

// skipDouble returns the index after the double-quoted string starting at s[i].
func skipDouble(s string, i int) int {
	for j := i + 1; j < len(s); {
		switch {
		case s[j] == '\\':
			j += 2
		case s[j] == '"':
			return j + 1
		case strings.HasPrefix(s[j:], "$("):
			j = skipSubst(s, j)
		default:
			j++
		}
	}
	return len(s)
}

// skipSubst returns the index after the $( ... ) starting at s[i], honoring quotes, nesting and heredocs.
func skipSubst(s string, i int) int {
	depth := 1
	for j := i + 2; j < len(s); {
		switch {
		case s[j] == '\\':
			j += 2
		case s[j] == '\'':
			j = skipSingle(s, j)
		case s[j] == '"':
			j = skipDouble(s, j)
		case strings.HasPrefix(s[j:], "$("):
			j = skipSubst(s, j)
		case strings.HasPrefix(s[j:], "<<") && !strings.HasPrefix(s[j:], "<<<"):
			j = skipHeredoc(s, j)
		case s[j] == '(':
			depth++
			j++
		case s[j] == ')':
			if depth--; depth == 0 {
				return j + 1
			}
			j++
		default:
			j++
		}
	}
	return len(s)
}

// heredocStart parses `<<[-]['"]WORD['"]` at s[i] and returns WORD and the index after it ("" if malformed).
func heredocStart(s string, i int) (string, int) {
	j := i + 2
	if j < len(s) && s[j] == '-' {
		j++
	}
	for j < len(s) && s[j] == ' ' {
		j++
	}
	if j < len(s) && (s[j] == '\'' || s[j] == '"') {
		j++ // opening quote
	}
	start := j
	for j < len(s) && strings.IndexByte(" \t\n)'\";|&<>", s[j]) < 0 {
		j++
	}
	word := s[start:j]
	if j < len(s) && (s[j] == '\'' || s[j] == '"') {
		j++ // closing quote
	}
	return word, j
}

// skipHeredoc returns the index after the heredoc body whose `<<WORD` starts at s[i].
func skipHeredoc(s string, i int) int {
	word, j := heredocStart(s, i)
	if word == "" {
		return i + 2
	}
	nl := strings.IndexByte(s[j:], '\n')
	if nl < 0 {
		return len(s)
	}
	for pos := j + nl + 1; pos <= len(s); {
		end := strings.IndexByte(s[pos:], '\n')
		line := s[pos:]
		if end >= 0 {
			line = s[pos : pos+end]
		}
		if strings.TrimSpace(line) == word {
			return pos + len(line)
		}
		if end < 0 {
			break
		}
		pos += end + 1
	}
	return len(s)
}

// elem is one command of a sequence: its text, or a ( ) subshell, and the operator that follows it
// ("&&", "||", ";", "|", "&", "\n", or "" for the last).
type elem struct {
	text string
	sub  []elem
	op   string
}

// parseSeq splits s from index i into commands on && || ; | & and newlines outside quotes and substitutions.
// A ( at the start of a command opens a subshell, parsed recursively; nested reports that a ) ends the sequence.
// It returns the elements and the index after the closing ) (or len(s)).
func parseSeq(s string, i int, nested bool) ([]elem, int) {
	var out []elem
	var sub []elem
	hasSub := false
	last := i
	flush := func(end int, op string) {
		text := strings.TrimSpace(s[last:end])
		if hasSub {
			out = append(out, elem{sub: sub, op: op})
		} else if text != "" {
			out = append(out, elem{text: text, op: op})
		} else if op != "" && len(out) > 0 && out[len(out)-1].op == "" {
			out[len(out)-1].op = op
		}
		sub, hasSub = nil, false
	}
	for i < len(s) {
		switch {
		case s[i] == '\\':
			i += 2
		case s[i] == '\'':
			i = skipSingle(s, i)
		case s[i] == '"':
			i = skipDouble(s, i)
		case s[i] == '`':
			i = skipBacktick(s, i)
		case startsSubst(s, i):
			i = skipSubst(s, i)
		case s[i] == '(' && !hasSub && strings.TrimSpace(s[last:i]) == "":
			sub, i = parseSeq(s, i+1, true)
			hasSub = true
			last = i
		case s[i] == ')' && nested:
			flush(i, "")
			return out, i + 1
		case strings.HasPrefix(s[i:], "&&") || strings.HasPrefix(s[i:], "||"):
			flush(i, s[i:i+2])
			i += 2
			last = i
		case strings.IndexByte(";|&\n", s[i]) >= 0:
			flush(i, s[i:i+1])
			i++
			last = i
		default:
			i++
		}
	}
	if last < len(s) || hasSub {
		flush(len(s), "")
	}
	return out, len(s)
}

// startsSubst reports whether s[i:] opens $( ... ), <( ... ) or >( ... ).
func startsSubst(s string, i int) bool {
	return strings.HasPrefix(s[i:], "$(") || strings.HasPrefix(s[i:], "<(") || strings.HasPrefix(s[i:], ">(")
}

// skipBacktick returns the index after the `...` starting at s[i].
func skipBacktick(s string, i int) int {
	for j := i + 1; j < len(s); {
		switch s[j] {
		case '\\':
			j += 2
		case '`':
			return j + 1
		default:
			j++
		}
	}
	return len(s)
}

// substitutions returns the bodies of the command substitutions ($( ), backticks, <( ), >( )) in s that the
// shell would run: outside single quotes, including inside double quotes. Nested ones stay inside their body.
func substitutions(s string) []string {
	var out []string
	body := func(start, end int, closer byte) string {
		if end > start && end <= len(s) && s[end-1] == closer {
			end--
		}
		return s[start:end]
	}
	for i := 0; i < len(s); {
		switch {
		case s[i] == '\\':
			i += 2
		case s[i] == '\'':
			i = skipSingle(s, i)
		case s[i] == '"':
			i++
			for i < len(s) && s[i] != '"' {
				switch {
				case s[i] == '\\':
					i += 2
				case s[i] == '`':
					end := skipBacktick(s, i)
					out = append(out, body(i+1, end, '`'))
					i = end
				case strings.HasPrefix(s[i:], "$("):
					end := skipSubst(s, i)
					out = append(out, body(i+2, end, ')'))
					i = end
				default:
					i++
				}
			}
			i++
		case s[i] == '`':
			end := skipBacktick(s, i)
			out = append(out, body(i+1, end, '`'))
			i = end
		case startsSubst(s, i):
			end := skipSubst(s, i)
			out = append(out, body(i+2, end, ')'))
			i = end
		default:
			i++
		}
	}
	return out
}

// stripHeredocs removes the bodies (and terminator lines) of top-level heredocs: they are data, not commands.
// Heredocs inside quotes or substitutions are left alone; those belong to their enclosing word.
func stripHeredocs(s string) string {
	type pending struct {
		word string
		dash bool
	}
	var out strings.Builder
	var queue []pending
	for i := 0; i < len(s); {
		from := i
		switch {
		case s[i] == '\\':
			i = min(i+2, len(s))
		case s[i] == '\'':
			i = skipSingle(s, i)
		case s[i] == '"':
			i = skipDouble(s, i)
		case s[i] == '`':
			i = skipBacktick(s, i)
		case startsSubst(s, i):
			i = skipSubst(s, i)
		case strings.HasPrefix(s[i:], "<<") && !strings.HasPrefix(s[i:], "<<<"):
			word, j := heredocStart(s, i)
			if word == "" {
				j = i + 2
			} else {
				queue = append(queue, pending{word, strings.HasPrefix(s[i:], "<<-")})
			}
			i = j
		case s[i] == '\n' && len(queue) > 0:
			out.WriteByte('\n')
			i++
			for _, p := range queue {
				i = heredocBodyEnd(s, i, p.word, p.dash)
			}
			queue = nil
			continue
		default:
			i++
		}
		out.WriteString(s[from:i])
	}
	return out.String()
}

// heredocBodyEnd returns the index after the terminator line of the heredoc body starting at pos.
func heredocBodyEnd(s string, pos int, word string, dash bool) int {
	for pos < len(s) {
		end := strings.IndexByte(s[pos:], '\n')
		line, next := s[pos:], len(s)
		if end >= 0 {
			line, next = s[pos:pos+end], pos+end+1
		}
		line = strings.TrimSuffix(line, "\r")
		if dash {
			line = strings.TrimLeft(line, "\t")
		}
		if line == word {
			return next
		}
		pos = next
	}
	return len(s)
}

// words splits one command into arguments with quotes removed. A $(...) is kept verbatim inside its word.
func words(s string) []string {
	var out []string
	for i := 0; i < len(s); {
		if strings.IndexByte(" \t\n", s[i]) >= 0 {
			i++
			continue
		}
		var w strings.Builder
		for i < len(s) && strings.IndexByte(" \t\n", s[i]) < 0 {
			switch {
			case s[i] == '\\' && i+1 < len(s):
				w.WriteByte(s[i+1])
				i += 2
			case s[i] == '\'':
				end := skipSingle(s, i)
				w.WriteString(s[i+1 : max(i+1, end-1)])
				i = end
			case s[i] == '"':
				i = appendDouble(&w, s, i)
			case strings.HasPrefix(s[i:], "$("):
				end := skipSubst(s, i)
				w.WriteString(s[i:end])
				i = end
			default:
				w.WriteByte(s[i])
				i++
			}
		}
		out = append(out, w.String())
	}
	return out
}

// appendDouble writes the content of the double-quoted string at s[i] to w and returns the index after it.
func appendDouble(w *strings.Builder, s string, i int) int {
	j := i + 1
	for j < len(s) && s[j] != '"' {
		switch {
		case s[j] == '\\' && j+1 < len(s) && strings.IndexByte("\"\\$`\n", s[j+1]) >= 0:
			w.WriteByte(s[j+1])
			j += 2
		case strings.HasPrefix(s[j:], "$("):
			end := skipSubst(s, j)
			w.WriteString(s[j:end])
			j = end
		default:
			w.WriteByte(s[j])
			j++
		}
	}
	return min(j+1, len(s))
}
