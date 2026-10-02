package cli

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
	"github.com/schollz/progressbar/v3"
)

// paint wraps its arguments in terminal color codes, or formats them plainly
// when color is off.
type paint func(a ...any) string

// palette holds the styles of one output stream.
type palette struct {
	errLabel paint // the "cisort:" prefix of errors
	file     paint // a file that was changed
	unsorted paint // a file that needs sorting
	header   paint // the file names heading a diff
	hunk     paint // a diff hunk header
	added    paint // an added diff line
	removed  paint // a removed diff line
	note     paint // "\ No newline at end of file"
	count    paint // numbers in the summary
}

func newPalette(enabled bool) palette {
	style := func(attrs ...color.Attribute) paint {
		c := color.New(attrs...)
		if enabled {
			c.EnableColor()
		} else {
			c.DisableColor()
		}
		return c.SprintFunc()
	}
	return palette{
		errLabel: style(color.FgRed, color.Bold),
		file:     style(color.FgGreen),
		unsorted: style(color.FgYellow),
		header:   style(color.Bold),
		hunk:     style(color.FgCyan),
		added:    style(color.FgGreen),
		removed:  style(color.FgRed),
		note:     style(color.Faint),
		count:    style(color.Bold),
	}
}

// terminal reports whether w is a terminal.
func terminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	return ok && (isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd()))
}

// colorWriter returns w ready for color codes and whether color is on.
// Color is on for terminals unless disabled by -no-color or NO_COLOR.
func colorWriter(w io.Writer, noColor bool) (io.Writer, bool) {
	if noColor || os.Getenv("NO_COLOR") != "" || !terminal(w) {
		return w, false
	}
	// Translates color codes on Windows consoles without ANSI support.
	return colorable.NewColorable(w.(*os.File)), true
}

// colorDiff colors a unified diff of one file.
func (p palette) colorDiff(diff string) string {
	lines := strings.SplitAfter(diff, "\n")
	for i, l := range lines {
		text, eol := strings.CutSuffix(l, "\n")
		if text == "" {
			continue
		}
		switch {
		case i < 2: // "--- a/..." and "+++ b/..."
			text = p.header(text)
		case strings.HasPrefix(text, "@@"):
			text = p.hunk(text)
		case text[0] == '+':
			text = p.added(text)
		case text[0] == '-':
			text = p.removed(text)
		case text[0] == '\\':
			text = p.note(text)
		}
		if eol {
			text += "\n"
		}
		lines[i] = text
	}
	return strings.Join(lines, "")
}

// progress reports how many files are done.
type progress interface {
	Add(n int) error
	Finish() error
}

type noProgress struct{}

func (noProgress) Add(int) error { return nil }
func (noProgress) Finish() error { return nil }

// newProgress returns a progress bar over total files drawn on w, which must
// be a terminal, and cleared when finished.
func newProgress(w io.Writer, total int, description string, colored bool) progress {
	theme := progressbar.ThemeASCII
	theme.SaucerPadding = " "
	if colored {
		theme.Saucer = "[green]=[reset]"
		theme.SaucerHead = "[green]>[reset]"
	}
	return progressbar.NewOptions(total,
		progressbar.OptionSetWriter(w),
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetTheme(theme),
		progressbar.OptionEnableColorCodes(colored),
		progressbar.OptionSetWidth(30),
		progressbar.OptionShowCount(),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionThrottle(50*time.Millisecond),
		progressbar.OptionClearOnFinish(),
	)
}
