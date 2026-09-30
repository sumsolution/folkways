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
	for j < len(s) && strings.IndexByte(" \t\n)'\"", s[j]) < 0 {
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

// splitCommands splits a command line on && || ; | & and newlines outside quotes and $(...).
func splitCommands(s string) []string {
	var out []string
	last := 0
	flush := func(end, next int) {
		if seg := strings.TrimSpace(s[last:end]); seg != "" {
			out = append(out, seg)
		}
		last = next
	}
	for i := 0; i < len(s); {
		switch {
		case s[i] == '\\':
			i += 2
		case s[i] == '\'':
			i = skipSingle(s, i)
		case s[i] == '"':
			i = skipDouble(s, i)
		case strings.HasPrefix(s[i:], "$("):
			i = skipSubst(s, i)
		case strings.HasPrefix(s[i:], "&&") || strings.HasPrefix(s[i:], "||"):
			flush(i, i+2)
			i += 2
		case strings.IndexByte(";|&\n", s[i]) >= 0:
			flush(i, i+1)
			i++
		default:
			i++
		}
	}
	if last < len(s) {
		flush(len(s), len(s))
	}
	return out
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
