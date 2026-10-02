// Package diff renders unified diffs of text files.
package diff

import (
	"fmt"
	"strings"
)

// contextLines is the number of unchanged lines shown around a change.
const contextLines = 3

// maxCells bounds the size of the LCS table; larger changes are shown as a
// whole-block replacement.
const maxCells = 1 << 22

type op struct {
	kind byte // ' ', '-' or '+'
	line string
}

// Unified returns the unified diff turning a into b, or "" if they are equal.
func Unified(oldName, newName string, a, b []byte) string {
	ops := edits(splitLines(string(a)), splitLines(string(b)))

	var sb strings.Builder
	oldLine, newLine := make([]int, len(ops)+1), make([]int, len(ops)+1)
	for i, o := range ops {
		oldLine[i+1], newLine[i+1] = oldLine[i], newLine[i]
		if o.kind != '+' {
			oldLine[i+1]++
		}
		if o.kind != '-' {
			newLine[i+1]++
		}
	}

	prevEnd := 0
	for i := 0; i < len(ops); {
		for i < len(ops) && ops[i].kind == ' ' {
			i++
		}
		if i == len(ops) {
			break
		}
		start := max(prevEnd, i-contextLines)
		end := i
		for {
			for end < len(ops) && ops[end].kind != ' ' {
				end++
			}
			next := end
			for next < len(ops) && ops[next].kind == ' ' {
				next++
			}
			if next < len(ops) && next-end <= 2*contextLines {
				end = next
				continue
			}
			end = min(len(ops), end+contextLines)
			break
		}

		if sb.Len() == 0 {
			fmt.Fprintf(&sb, "--- %s\n+++ %s\n", oldName, newName)
		}
		fmt.Fprintf(&sb, "@@ -%s +%s @@\n",
			span(oldLine[start], oldLine[end]-oldLine[start]),
			span(newLine[start], newLine[end]-newLine[start]))
		for _, o := range ops[start:end] {
			text, hasEOL := strings.CutSuffix(o.line, "\n")
			sb.WriteByte(o.kind)
			sb.WriteString(strings.TrimSuffix(text, "\r"))
			sb.WriteByte('\n')
			if !hasEOL {
				sb.WriteString("\\ No newline at end of file\n")
			}
		}
		prevEnd, i = end, end
	}
	return sb.String()
}

func span(before, count int) string {
	if count == 0 {
		return fmt.Sprintf("%d,0", before)
	}
	if count == 1 {
		return fmt.Sprint(before + 1)
	}
	return fmt.Sprintf("%d,%d", before+1, count)
}

// edits returns an edit script turning a into b.
func edits(a, b []string) []op {
	prefix := 0
	for prefix < len(a) && prefix < len(b) && a[prefix] == b[prefix] {
		prefix++
	}
	suffix := 0
	for suffix < len(a)-prefix && suffix < len(b)-prefix && a[len(a)-1-suffix] == b[len(b)-1-suffix] {
		suffix++
	}

	ops := make([]op, 0, len(a)+len(b))
	for _, l := range a[:prefix] {
		ops = append(ops, op{' ', l})
	}
	ops = append(ops, lcs(a[prefix:len(a)-suffix], b[prefix:len(b)-suffix])...)
	for _, l := range a[len(a)-suffix:] {
		ops = append(ops, op{' ', l})
	}
	return ops
}

// lcs diffs a and b through their longest common subsequence.
func lcs(a, b []string) []op {
	var ops []op
	if (len(a)+1)*(len(b)+1) > maxCells {
		for _, l := range a {
			ops = append(ops, op{'-', l})
		}
		for _, l := range b {
			ops = append(ops, op{'+', l})
		}
		return ops
	}

	// length[i][j] is the LCS length of a[i:] and b[j:].
	width := len(b) + 1
	length := make([]int, (len(a)+1)*width)
	for i := len(a) - 1; i >= 0; i-- {
		for j := len(b) - 1; j >= 0; j-- {
			if a[i] == b[j] {
				length[i*width+j] = length[(i+1)*width+j+1] + 1
			} else {
				length[i*width+j] = max(length[(i+1)*width+j], length[i*width+j+1])
			}
		}
	}

	i, j := 0, 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i] == b[j]:
			ops = append(ops, op{' ', a[i]})
			i++
			j++
		case length[(i+1)*width+j] >= length[i*width+j+1]:
			ops = append(ops, op{'-', a[i]})
			i++
		default:
			ops = append(ops, op{'+', b[j]})
			j++
		}
	}
	for ; i < len(a); i++ {
		ops = append(ops, op{'-', a[i]})
	}
	for ; j < len(b); j++ {
		ops = append(ops, op{'+', b[j]})
	}
	return ops
}

// splitLines splits s after every newline.
func splitLines(s string) []string {
	lines := strings.SplitAfter(s, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}
