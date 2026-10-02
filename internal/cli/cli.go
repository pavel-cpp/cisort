// Package cli implements the cisort command.
package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"slices"
	"strings"
	"sync"

	"github.com/pavel-cpp/cisort/internal/config"
	"github.com/pavel-cpp/cisort/internal/diff"
	"github.com/pavel-cpp/cisort/internal/finder"
	"github.com/pavel-cpp/cisort/internal/sorter"
)

// Exit codes returned by Run.
const (
	ExitOK       = 0 // success; with -check, every file is sorted
	ExitUnsorted = 1 // with -check, some file is not sorted
	ExitError    = 2 // bad usage or a file could not be processed
)

// version is set at link time with
// -ldflags "-X github.com/pavel-cpp/cisort/internal/cli.version=v1.2.3".
// When empty, the module version recorded by go install is used.
var version string

const usage = `Usage: cisort [flags] [path ...]

cisort sorts #include directives in C and C++ files. Directories are walked
recursively; with no path the current directory is processed. The path "-"
reads standard input and writes the result to standard output.

The style comes from the nearest %s in the file's directory or its
parents, or from the %q preset. See -list-presets and -print-config.

Flags:
`

const examples = `
Examples:
  cisort                          sort every C/C++ file under the current directory
  cisort -check -diff src         show what is unsorted and exit with status 1
  cisort -preset google -c .      regroup the google way, with group comments
  cisort -preset qt -print-config > .cisort.json
                                  start a config file from a preset
`

type options struct {
	check, diff       bool
	regroup, comments bool
	preset, config    string
	ext               string
	exclude           []string
	stdinFilename     string
	jobs              int
	verbose           bool
	listPresets       bool
	printConfig       bool
	version           bool
	noColor           bool
	noProgress        bool
}

// Run runs cisort with args, which exclude the program name, and returns the
// exit code.
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	var o options
	fs := flag.NewFlagSet("cisort", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintf(stderr, usage, config.FileName, config.DefaultPreset)
		fs.PrintDefaults()
		fmt.Fprint(stderr, examples)
	}
	fs.BoolVar(&o.check, "check", false, "do not write files; list unsorted files and exit with status 1 if there are any")
	fs.BoolVar(&o.diff, "diff", false, "do not write files; print a unified diff of the changes")
	fs.BoolVar(&o.regroup, "g", false, "regroup includes: merge blocks and separate groups with blank lines")
	fs.BoolVar(&o.comments, "c", false, "label groups with comments (implies -g)")
	fs.StringVar(&o.preset, "preset", "", "use the built-in preset `name` and ignore config files")
	fs.StringVar(&o.config, "config", "", "use the config `file` instead of looking up "+config.FileName)
	fs.StringVar(&o.ext, "ext", "", "comma-separated file `extensions` to process, e.g. .cpp,.h")
	fs.Func("exclude", "skip files and directories matching `glob` (repeatable)", func(s string) error {
		o.exclude = append(o.exclude, s)
		return nil
	})
	fs.StringVar(&o.stdinFilename, "stdin-filename", "", "`path` of the file read from stdin, used to find its main header and config")
	fs.IntVar(&o.jobs, "j", runtime.NumCPU(), "number of files processed in parallel")
	fs.BoolVar(&o.verbose, "v", false, "print changed files and a summary")
	fs.BoolVar(&o.listPresets, "list-presets", false, "print the built-in presets and exit")
	fs.BoolVar(&o.printConfig, "print-config", false, "print the configuration for the first path as JSON and exit")
	fs.BoolVar(&o.version, "version", false, "print the version and exit")
	fs.BoolVar(&o.noColor, "no-color", false, "disable colored output (NO_COLOR in the environment does the same)")
	fs.BoolVar(&o.noProgress, "no-progress", false, "do not show the progress bar")

	paths, err := parseInterspersed(fs, args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitError
	}

	r := newRunner(o, stdin, stdout, stderr)
	switch {
	case o.version:
		fmt.Fprintln(stdout, "cisort", Version())
		return ExitOK
	case o.listPresets:
		fmt.Fprintln(stdout, strings.Join(config.Presets(), "\n"))
		return ExitOK
	}
	if err := r.setup(); err != nil {
		return r.fail(err)
	}

	if len(paths) == 0 {
		paths = []string{"."}
	}
	switch {
	case o.printConfig:
		return r.printConfig(paths[0])
	case len(paths) == 1 && paths[0] == "-":
		return r.runStdin()
	case slices.Contains(paths, "-"):
		return r.fail(errors.New(`"-" cannot be combined with other paths`))
	}
	return r.runFiles(paths)
}

// parseInterspersed parses flags that may appear before, between or after
// paths, and returns the paths. Everything after "--" is a path.
func parseInterspersed(fs *flag.FlagSet, args []string) ([]string, error) {
	var paths []string
	for {
		if err := fs.Parse(args); err != nil {
			return nil, err
		}
		rest := fs.Args()
		if consumed := len(args) - len(rest); consumed > 0 && args[consumed-1] == "--" {
			return append(paths, rest...), nil
		}
		if len(rest) == 0 {
			return paths, nil
		}
		paths = append(paths, rest[0])
		args = rest[1:]
	}
}

// Version returns the version of cisort.
func Version() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return info.Main.Version
	}
	return "(devel)"
}

type runner struct {
	options
	stdin          io.Reader
	stdout, stderr io.Writer
	outPaint       palette // styles for stdout
	errPaint       palette // styles for stderr
	errColor       bool    // whether stderr takes color codes
	progressBar    bool    // whether to draw a progress bar on stderr
	configFor      func(dir string) (*config.Config, error)
}

func newRunner(o options, stdin io.Reader, stdout, stderr io.Writer) *runner {
	out, outColor := colorWriter(stdout, o.noColor)
	errOut, errColor := colorWriter(stderr, o.noColor)
	return &runner{
		options:     o,
		stdin:       stdin,
		stdout:      out,
		stderr:      errOut,
		outPaint:    newPalette(outColor),
		errPaint:    newPalette(errColor),
		errColor:    errColor,
		progressBar: !o.noProgress && terminal(stderr),
	}
}

// fail reports err and returns ExitError. Every line of a multi-line error
// gets the "cisort:" prefix.
func (r *runner) fail(err error) int {
	for _, line := range strings.Split(err.Error(), "\n") {
		fmt.Fprintln(r.stderr, r.errPaint.errLabel("cisort:"), line)
	}
	return ExitError
}

// setup validates the options and prepares the config lookup.
func (r *runner) setup() error {
	if r.jobs < 1 {
		return errors.New("-j must be at least 1")
	}
	if r.preset != "" && r.config != "" {
		return errors.New("-preset and -config are mutually exclusive")
	}

	var fixed *config.Config
	var err error
	switch {
	case r.preset != "":
		fixed, err = config.Preset(r.preset)
	case r.config != "":
		fixed, err = config.Load(r.config)
	default:
		var fallback *config.Config
		if fallback, err = config.Preset(config.DefaultPreset); err != nil {
			return err
		}
		r.configFor = config.NewResolver(fallback, r.adjust).For
		return nil
	}
	if err != nil {
		return err
	}
	r.adjust(fixed)
	r.configFor = func(string) (*config.Config, error) { return fixed, nil }
	return nil
}

// adjust applies the command line overrides to cfg.
func (r *runner) adjust(cfg *config.Config) {
	if r.regroup || r.comments {
		cfg.Blocks = config.Regroup
	}
	if r.comments {
		cfg.Comments = true
	}
	if r.ext != "" {
		cfg.Extensions = strings.Split(r.ext, ",")
	}
	cfg.Exclude = append(cfg.Exclude, r.exclude...)
}

func (r *runner) rules(dir string) (finder.Rules, error) {
	cfg, err := r.configFor(dir)
	if err != nil {
		return finder.Rules{}, err
	}
	return finder.Rules{Extensions: cfg.Extensions, Exclude: cfg.Exclude}, nil
}

type result struct {
	changed bool
	diff    string
	err     error
}

func (r *runner) runFiles(paths []string) int {
	code := ExitOK
	files, err := finder.Find(paths, r.rules)
	if err != nil {
		code = r.fail(err)
	}

	bar := progress(noProgress{})
	if r.progressBar && len(files) > 0 {
		description := "sorting"
		if r.check || r.diff {
			description = "checking"
		}
		bar = newProgress(r.stderr, len(files), description, r.errColor)
	}
	results := make([]result, len(files))
	next := make(chan int)
	var wg sync.WaitGroup
	for range min(r.jobs, len(files)) {
		wg.Go(func() {
			for i := range next {
				results[i] = r.processFile(files[i])
				bar.Add(1)
			}
		})
	}
	for i := range files {
		next <- i
	}
	close(next)
	wg.Wait()
	bar.Finish()

	changed := 0
	for i, res := range results {
		switch {
		case res.err != nil:
			code = r.fail(res.err)
			continue
		case !res.changed:
			continue
		case r.diff:
			fmt.Fprint(r.stdout, r.outPaint.colorDiff(res.diff))
		case r.check:
			fmt.Fprintln(r.stdout, r.outPaint.unsorted(files[i]))
		case r.verbose:
			fmt.Fprintln(r.stdout, r.outPaint.file(files[i]))
		}
		changed++
	}
	if r.verbose {
		verb := "sorted"
		if r.check || r.diff {
			verb = "unsorted"
		}
		p := r.errPaint
		fmt.Fprintf(r.stderr, "cisort: %s files checked, %s %s\n", p.count(len(files)), p.count(changed), verb)
	}
	if code == ExitOK && r.check && changed > 0 {
		code = ExitUnsorted
	}
	return code
}

func (r *runner) processFile(path string) result {
	cfg, err := r.configFor(filepath.Dir(path))
	if err != nil {
		return result{err: err}
	}
	src, err := os.ReadFile(path)
	if err != nil {
		return result{err: err}
	}
	out := sorter.Sort(src, path, cfg)
	if bytes.Equal(src, out) {
		return result{}
	}
	res := result{changed: true}
	switch {
	case r.diff:
		name := filepath.ToSlash(path)
		res.diff = diff.Unified("a/"+name, "b/"+name, src, out)
	case !r.check:
		res.err = writeFile(path, out)
	}
	return res
}

func (r *runner) runStdin() int {
	src, err := io.ReadAll(r.stdin)
	if err != nil {
		return r.fail(err)
	}
	name, dir := r.stdinFilename, "."
	if name != "" {
		dir = filepath.Dir(name)
	}
	cfg, err := r.configFor(dir)
	if err != nil {
		return r.fail(err)
	}

	out := sorter.Sort(src, name, cfg)
	changed := !bytes.Equal(src, out)
	if name == "" {
		name = "<stdin>"
	}
	switch {
	case r.diff:
		fmt.Fprint(r.stdout, r.outPaint.colorDiff(diff.Unified(name, name, src, out)))
	case r.check:
		if changed {
			fmt.Fprintln(r.stdout, r.outPaint.unsorted(name))
		}
	default:
		r.stdout.Write(out)
	}
	if r.check && changed {
		return ExitUnsorted
	}
	return ExitOK
}

func (r *runner) printConfig(path string) int {
	dir := path
	switch info, err := os.Stat(path); {
	case path == "-":
		dir = filepath.Dir(r.stdinFilename)
	case err != nil:
		return r.fail(err)
	case !info.IsDir():
		dir = filepath.Dir(path)
	}
	cfg, err := r.configFor(dir)
	if err != nil {
		return r.fail(err)
	}
	enc := json.NewEncoder(r.stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(cfg); err != nil {
		return r.fail(err)
	}
	return ExitOK
}

// writeFile replaces the contents of path atomically, keeping its permissions.
func writeFile(path string, data []byte) (err error) {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".cisort-*")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			os.Remove(tmp.Name())
		}
	}()
	if _, err = tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	if err = os.Chmod(tmp.Name(), info.Mode().Perm()); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
