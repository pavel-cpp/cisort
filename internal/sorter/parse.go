package sorter

import (
	"regexp"
	"strings"
)

// kind classifies a source line.
type kind int

const (
	kindOther   kind = iota // code, other directives, anything cisort must not move
	kindBlank               // whitespace only
	kindComment             // a line holding nothing but a comment
	kindLabel               // a group label written by cisort, e.g. "// Project"
	kindInclude             // a sortable #include, #import or #include_next
)

type line struct {
	text string // without the line terminator
	eol  string // "\n", "\r\n" or "" for the last line of a file
	kind kind
	inc  include // set when kind == kindInclude
}

type include struct {
	directive string // include, import or include_next
	spelling  string // path with its delimiters, e.g. <vector>
	tail      string // trailing comment, if any
}

var (
	includeRE = regexp.MustCompile(`^[ \t]*#[ \t]*(include_next|include|import)[ \t]*(<[^<>]*>|"[^"]*")[ \t]*(.*)$`)
	markerRE  = regexp.MustCompile(`(?i)^(?:cisort|clang-format):?[ \t]*(off|on)\b`)
	keepRE    = regexp.MustCompile(`(?i)\bcisort:?[ \t]*keep\b`)
)

// splitLines splits src into lines, keeping each line's terminator.
func splitLines(src string) []line {
	var lines []line
	for src != "" {
		i := strings.IndexByte(src, '\n')
		if i < 0 {
			lines = append(lines, line{text: src})
			break
		}
		text, eol := src[:i], "\n"
		if strings.HasSuffix(text, "\r") {
			text, eol = text[:len(text)-1], "\r\n"
		}
		lines = append(lines, line{text: text, eol: eol})
		src = src[i+1:]
	}
	return lines
}

// classify sets the kind of every line. Lines whose trimmed text is in labels
// are classified as kindLabel.
func classify(lines []line, labels map[string]bool) {
	inComment, off := false, false
	for i := range lines {
		l := &lines[i]
		startsInComment := inComment
		inComment = scanComments(l.text, inComment)
		trimmed := strings.TrimSpace(l.text)

		switch {
		case startsInComment:
			l.kind = kindOther
		case trimmed == "":
			l.kind = kindBlank
		case !inComment && isCommentOnly(trimmed):
			switch marker(trimmed) {
			case "off":
				off, l.kind = true, kindOther
			case "on":
				off, l.kind = false, kindOther
			default:
				switch {
				case off:
					l.kind = kindOther
				case labels[trimmed]:
					l.kind = kindLabel
				default:
					l.kind = kindComment
				}
			}
		case off || inComment:
			l.kind = kindOther
		default:
			if inc, ok := parseInclude(l.text); ok {
				l.kind, l.inc = kindInclude, inc
			} else {
				l.kind = kindOther
			}
		}
	}
}

// parseInclude parses a sortable include directive. Directives with code or an
// unterminated comment after the path, line continuations and lines marked
// "cisort: keep" are not sortable.
func parseInclude(text string) (include, bool) {
	m := includeRE.FindStringSubmatch(text)
	if m == nil || strings.HasSuffix(text, `\`) {
		return include{}, false
	}
	inc := include{directive: m[1], spelling: m[2], tail: strings.TrimRight(m[3], " \t")}
	if (inc.tail != "" && !isCommentOnly(inc.tail)) || keepRE.MatchString(inc.tail) {
		return include{}, false
	}
	return inc, true
}

// isCommentOnly reports whether s, already trimmed, is a single complete
// comment.
func isCommentOnly(s string) bool {
	if strings.HasPrefix(s, "//") {
		return !strings.HasSuffix(s, `\`)
	}
	if strings.HasPrefix(s, "/*") {
		end := strings.Index(s[2:], "*/")
		return end >= 0 && end+4 == len(s)
	}
	return false
}

// marker returns "off" or "on" for "cisort: off/on" and "clang-format off/on"
// comments, and "" otherwise.
func marker(comment string) string {
	body := strings.TrimPrefix(comment, "//")
	if strings.HasPrefix(comment, "/*") {
		body = strings.TrimSuffix(comment[2:], "*/")
	}
	m := markerRE.FindStringSubmatch(strings.TrimSpace(body))
	if m == nil {
		return ""
	}
	return strings.ToLower(m[1])
}

// scanComments reports whether a block comment is still open at the end of s,
// given whether one was open at its start. String and character literals are
// skipped so that "/*" inside them is not mistaken for a comment.
func scanComments(s string, inComment bool) bool {
	for i := 0; i < len(s); i++ {
		if inComment {
			if s[i] == '*' && i+1 < len(s) && s[i+1] == '/' {
				inComment = false
				i++
			}
			continue
		}
		switch s[i] {
		case '/':
			if i+1 < len(s) {
				switch s[i+1] {
				case '/':
					return false
				case '*':
					inComment = true
					i++
				}
			}
		case '\'':
			// A quote between hex digits is a digit separator: 1'000'000.
			if i > 0 && i+1 < len(s) && isHexDigit(s[i-1]) && isHexDigit(s[i+1]) {
				continue
			}
			i = skipLiteral(s, i)
		case '"':
			i = skipLiteral(s, i)
		}
	}
	return inComment
}

// skipLiteral returns the index of the quote closing the literal opened at i,
// or the last index of s if the literal is not closed.
func skipLiteral(s string, i int) int {
	quote := s[i]
	for j := i + 1; j < len(s); j++ {
		switch s[j] {
		case '\\':
			j++
		case quote:
			return j
		}
	}
	return len(s) - 1
}

func isHexDigit(c byte) bool {
	return '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F'
}
