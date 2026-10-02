// Package sorter orders #include directives in C and C++ source code.
//
// Sorting works on runs of include lines. Blank lines separate blocks within
// a run; comment lines between two includes travel with the include below
// them. Any other line (code, #if, #pragma, a comment after a blank line, an
// include marked "cisort: keep", ...) ends the run, so the code around
// includes is never touched.
package sorter

import (
	"bytes"
	"cmp"
	"path/filepath"
	"slices"
	"strings"

	"github.com/pavel-cpp/cisort/internal/config"
	"github.com/pavel-cpp/cisort/internal/headers"
)

var utf8BOM = []byte("\xEF\xBB\xBF")

// Sort returns src with its include directives ordered according to cfg.
// filename is used to find the main header and may be empty. Line endings,
// the byte order mark and every line that is not moved are kept as is.
func Sort(src []byte, filename string, cfg *config.Config) []byte {
	bom := bytes.HasPrefix(src, utf8BOM)
	if bom {
		src = src[len(utf8BOM):]
	}
	lines := splitLines(string(src))

	var labels map[string]bool
	if cfg.Blocks == config.Regroup {
		labels = make(map[string]bool)
		for _, name := range cfg.Labels() {
			labels[label(name)] = true
		}
	}
	classify(lines, labels)

	s := sorter{cfg: cfg, eol: dominantEOL(lines), mainStems: mainStems(filename, cfg)}

	var out bytes.Buffer
	out.Grow(len(src) + len(utf8BOM))
	if bom {
		out.Write(utf8BOM)
	}
	pos := 0
	for _, u := range s.units(lines) {
		writeLines(&out, lines[pos:u.start])
		writeLines(&out, s.render(lines[u.start:u.end]))
		pos = u.end
	}
	writeLines(&out, lines[pos:])
	return out.Bytes()
}

type sorter struct {
	cfg       *config.Config
	eol       string   // terminator for generated lines
	mainStems []string // base names the main header may have
}

// unit is a range of lines that is sorted as a whole.
type unit struct{ start, end int }

// entry is an include together with the comments attached above it.
type entry struct {
	comments []line
	line     line
	category config.Category
	key      string
}

// units returns the line ranges to sort: every block in preserve mode, every
// run of blocks otherwise.
func (s *sorter) units(lines []line) []unit {
	var units []unit
	for i := 0; i < len(lines); {
		start := i
		if lines[i].kind == kindLabel {
			// A label starts a run together with the comments below it.
			i = skipComments(lines, i)
		}
		if i == len(lines) || lines[i].kind != kindInclude {
			i = start + 1
			continue
		}
		end := runEnd(lines, i)
		if s.cfg.Blocks == config.Preserve {
			units = append(units, blocks(lines, start, end)...)
		} else {
			units = append(units, unit{start, end})
		}
		i = end
	}
	return units
}

// runEnd returns the end of the run of includes starting at the include at i.
// The run ends after its last include line.
func runEnd(lines []line, i int) int {
	end := i + 1
	for j := i; j < len(lines); {
		switch lines[j].kind {
		case kindInclude:
			j++
			end = j
		case kindBlank:
			j++
		case kindComment, kindLabel:
			// Comments belong to the run when they sit between two includes
			// or below a label, which may follow a blank line.
			r := skipComments(lines, j)
			if r == len(lines) || lines[r].kind != kindInclude ||
				(lines[j-1].kind != kindInclude && lines[j].kind != kindLabel) {
				return end
			}
			j = r
		default:
			return end
		}
	}
	return end
}

// skipComments returns the index of the first line at or after i that is
// neither a comment nor a label.
func skipComments(lines []line, i int) int {
	for i < len(lines) && (lines[i].kind == kindComment || lines[i].kind == kindLabel) {
		i++
	}
	return i
}

// blocks splits lines[start:end] at blank lines.
func blocks(lines []line, start, end int) []unit {
	var units []unit
	for i := start; i < end; {
		if lines[i].kind == kindBlank {
			i++
			continue
		}
		j := i
		for j < end && lines[j].kind != kindBlank {
			j++
		}
		units = append(units, unit{i, j})
		i = j
	}
	return units
}

// render returns the sorted replacement for the lines of one unit.
func (s *sorter) render(lines []line) []line {
	entries := s.entries(lines)
	if s.cfg.Dedup {
		entries = dedup(entries)
	}
	slices.SortStableFunc(entries, func(a, b entry) int {
		if c := cmp.Compare(a.category.Priority, b.category.Priority); c != 0 {
			return c
		}
		if c := compareKeys(a.key, b.key, s.cfg.Sort); c != 0 {
			return c
		}
		return strings.Compare(a.line.text, b.line.text)
	})

	// Labels only pay off when they tell groups apart.
	labeled := s.cfg.Comments && entries[0].category.Priority != entries[len(entries)-1].category.Priority

	out := make([]line, 0, len(lines))
	for i, e := range entries {
		if s.cfg.Blocks == config.Regroup && (i == 0 || e.category.Priority != entries[i-1].category.Priority) {
			if i > 0 {
				out = append(out, line{})
			}
			if labeled {
				out = append(out, line{text: label(e.category.Name)})
			}
		}
		out = append(out, e.comments...)
		out = append(out, e.line)
	}

	for i := range out {
		if out[i].eol == "" {
			out[i].eol = s.eol
		}
	}
	out[len(out)-1].eol = lines[len(lines)-1].eol
	return out
}

// entries collects the includes of a unit. Blank lines and old labels are
// dropped; they are regenerated by render.
func (s *sorter) entries(lines []line) []entry {
	var entries []entry
	var comments []line
	for _, l := range lines {
		switch l.kind {
		case kindComment:
			comments = append(comments, l)
		case kindInclude:
			isMain := s.isMain(l.inc.spelling)
			entries = append(entries, entry{
				comments: comments,
				line:     l,
				category: s.cfg.Categorize(l.inc.spelling, isMain),
				key:      sortKey(l.inc.spelling, s.cfg.Sort),
			})
			comments = nil
		}
	}
	return entries
}

// dedup drops repeated includes unless dropping would lose a comment.
func dedup(entries []entry) []entry {
	tails := make(map[[2]string]string) // directive and spelling → tail of the first occurrence
	return slices.DeleteFunc(entries, func(e entry) bool {
		id := [2]string{e.line.inc.directive, e.line.inc.spelling}
		tail, seen := tails[id]
		if !seen {
			tails[id] = e.line.inc.tail
			return false
		}
		return len(e.comments) == 0 && (e.line.inc.tail == "" || e.line.inc.tail == tail)
	})
}

// isMain reports whether the include is the main header of the file: the
// header declaring what the file implements. Standard headers never are, so
// string.cpp does not pair with <string>.
func (s *sorter) isMain(spelling string) bool {
	name := spelling[1 : len(spelling)-1]
	if len(s.mainStems) == 0 || (spelling[0] == '<' && headers.Standard(name)) {
		return false
	}
	stem := stripExt(lastElem(name))
	for _, m := range s.mainStems {
		if strings.EqualFold(stem, m) {
			return true
		}
	}
	return false
}

// mainStems returns the base names the main header of filename may have:
// "widget" for widget.cpp, widget.inl and widget_impl.h, and both
// "widget_test" and "widget" for widget_test.cpp.
func mainStems(filename string, cfg *config.Config) []string {
	if filename == "" || !cfg.UsesMain() {
		return nil
	}
	base := filepath.Base(filename)
	stem := stripExt(base)
	var stems []string
	if hasExt(cfg.ImplExtensions, filepath.Ext(base)) {
		stems = append(stems, stem)
	}
	for _, suffix := range cfg.MainSuffixes {
		if s, ok := strings.CutSuffix(stem, suffix); ok && s != "" {
			stems = append(stems, s)
			break
		}
	}
	return stems
}

func hasExt(exts []string, ext string) bool {
	for _, e := range exts {
		if strings.EqualFold(ext, "."+strings.TrimPrefix(e, ".")) {
			return true
		}
	}
	return false
}

func lastElem(path string) string {
	return path[strings.LastIndexAny(path, `/\`)+1:]
}

func stripExt(name string) string {
	if i := strings.LastIndexByte(name, '.'); i > 0 {
		return name[:i]
	}
	return name
}

func label(name string) string {
	return "// " + name
}

func dominantEOL(lines []line) string {
	for _, l := range lines {
		if l.eol != "" {
			return l.eol
		}
	}
	return "\n"
}

func writeLines(buf *bytes.Buffer, lines []line) {
	for _, l := range lines {
		buf.WriteString(l.text)
		buf.WriteString(l.eol)
	}
}
