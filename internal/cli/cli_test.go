package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pavel-cpp/cisort/internal/config"
)

const (
	unsorted = "#include <vector>\n#include <algorithm>\n"
	sorted   = "#include <algorithm>\n#include <vector>\n"
)

func run(t *testing.T, stdin string, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errOut bytes.Buffer
	code = Run(args, strings.NewReader(stdin), &out, &errOut)
	return code, out.String(), errOut.String()
}

// project creates files from name → content and returns the root directory.
func project(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, content := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func read(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestSortsFilesInPlace(t *testing.T) {
	root := project(t, map[string]string{
		"a.cpp":       unsorted,
		"sub/b.h":     sorted,
		"notes.txt":   unsorted,
		"build/c.cpp": unsorted,
	})

	code, stdout, stderr := run(t, "", "-v", root)
	if code != ExitOK {
		t.Fatalf("exit code = %d, stderr = %s", code, stderr)
	}
	if got := read(t, filepath.Join(root, "a.cpp")); got != sorted {
		t.Errorf("a.cpp =\n%s", got)
	}
	for _, name := range []string{"notes.txt", "build/c.cpp"} {
		if got := read(t, filepath.Join(root, name)); got != unsorted {
			t.Errorf("%s was modified", name)
		}
	}
	if !strings.Contains(stdout, "a.cpp") || strings.Contains(stdout, "b.h") {
		t.Errorf("stdout = %q, want only a.cpp", stdout)
	}
	if !strings.Contains(stderr, "2 files checked, 1 sorted") {
		t.Errorf("stderr = %q, want a summary", stderr)
	}
}

func TestCheck(t *testing.T) {
	root := project(t, map[string]string{"a.cpp": unsorted, "b.cpp": sorted})

	code, stdout, _ := run(t, "", "-check", root)
	if code != ExitUnsorted {
		t.Errorf("exit code = %d, want %d", code, ExitUnsorted)
	}
	if want := filepath.Join(root, "a.cpp") + "\n"; stdout != want {
		t.Errorf("stdout = %q, want %q", stdout, want)
	}
	if got := read(t, filepath.Join(root, "a.cpp")); got != unsorted {
		t.Error("-check modified a file")
	}

	if code, _, _ := run(t, "", "-check", filepath.Join(root, "b.cpp")); code != ExitOK {
		t.Errorf("exit code for a sorted file = %d, want %d", code, ExitOK)
	}
}

func TestDiff(t *testing.T) {
	root := project(t, map[string]string{"a.cpp": unsorted})

	code, stdout, _ := run(t, "", root, "-diff", "-check")
	if code != ExitUnsorted {
		t.Errorf("exit code = %d, want %d", code, ExitUnsorted)
	}
	for _, want := range []string{"--- a/", "+++ b/", "-#include <vector>", "+#include <vector>"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("diff does not contain %q:\n%s", want, stdout)
		}
	}
	if got := read(t, filepath.Join(root, "a.cpp")); got != unsorted {
		t.Error("-diff modified a file")
	}
}

func TestStdin(t *testing.T) {
	code, stdout, stderr := run(t, unsorted, "-")
	if code != ExitOK || stdout != sorted {
		t.Errorf("exit code = %d, stdout = %q, stderr = %q", code, stdout, stderr)
	}

	code, stdout, _ = run(t, "#include <vector>\n#include \"app.h\"\n", "-preset", "google", "-c", "-")
	want := "// C++ standard library\n#include <vector>\n\n// Project headers\n#include \"app.h\"\n"
	if code != ExitOK || stdout != want {
		t.Errorf("regroup: exit code = %d, stdout =\n%s\nwant\n%s", code, stdout, want)
	}

	if code, stdout, _ := run(t, unsorted, "-check", "-"); code != ExitUnsorted || stdout != "<stdin>\n" {
		t.Errorf("-check: exit code = %d, stdout = %q", code, stdout)
	}
}

func TestStdinFilenameFindsConfig(t *testing.T) {
	root := project(t, map[string]string{
		"src/" + config.FileName: `{"preset": "google"}`,
	})
	in := "#include <vector>\n#include \"widget.h\"\n"
	want := "#include \"widget.h\"\n\n#include <vector>\n"

	code, stdout, stderr := run(t, in, "-stdin-filename", filepath.Join(root, "src", "widget.cc"), "-")
	if code != ExitOK || stdout != want {
		t.Errorf("exit code = %d, stdout =\n%s\nwant\n%s\nstderr = %s", code, stdout, want, stderr)
	}
}

func TestConfigFiles(t *testing.T) {
	root := project(t, map[string]string{
		config.FileName:             `{"extensions": [".cxx"], "exclude": ["skip"]}`,
		"a.cxx":                     unsorted,
		"a.cpp":                     unsorted,
		"skip/b.cxx":                unsorted,
		"broken/" + "x.h":           unsorted,
		"broken/" + config.FileName: `{"blocks": "sideways"}`,
	})

	code, _, stderr := run(t, "", root)
	if code != ExitError || !strings.Contains(stderr, "sideways") {
		t.Errorf("exit code = %d, stderr = %q, want a config error", code, stderr)
	}
	if got := read(t, filepath.Join(root, "a.cxx")); got != sorted {
		t.Error("a.cxx was not sorted")
	}
	for _, name := range []string{"a.cpp", "skip/b.cxx", "broken/x.h"} {
		if got := read(t, filepath.Join(root, name)); got != unsorted {
			t.Errorf("%s was modified", name)
		}
	}
}

func TestPrintConfigRoundTrip(t *testing.T) {
	code, stdout, stderr := run(t, "", "-preset", "topics", "-print-config")
	if code != ExitOK {
		t.Fatalf("exit code = %d, stderr = %s", code, stderr)
	}
	cfg, err := config.Parse([]byte(stdout))
	if err != nil {
		t.Fatalf("printed config does not parse: %v\n%s", err, stdout)
	}
	if cfg.Preset != "topics" || !cfg.Comments {
		t.Errorf("round trip lost settings: %+v", cfg)
	}
}

func TestInfoFlags(t *testing.T) {
	if code, stdout, _ := run(t, "", "-version"); code != ExitOK || !strings.HasPrefix(stdout, "cisort ") {
		t.Errorf("-version: %d %q", code, stdout)
	}
	if code, stdout, _ := run(t, "", "-list-presets"); code != ExitOK || !strings.Contains(stdout, "google\n") {
		t.Errorf("-list-presets: %d %q", code, stdout)
	}
	if code, _, stderr := run(t, "", "-h"); code != ExitOK || !strings.Contains(stderr, "Usage: cisort") {
		t.Errorf("-h: %d %q", code, stderr)
	}
}

func TestUsageErrors(t *testing.T) {
	tests := [][]string{
		{"-unknown"},
		{"-preset", "nope"},
		{"-preset", "google", "-config", "x.json"},
		{"-j", "0"},
		{"-", "."},
		{"does-not-exist"},
	}
	for _, args := range tests {
		if code, _, _ := run(t, "", args...); code != ExitError {
			t.Errorf("cisort %v: exit code = %d, want %d", args, code, ExitError)
		}
	}
}

func TestDoubleDash(t *testing.T) {
	root := project(t, map[string]string{"-odd.cpp": unsorted})
	t.Chdir(root)
	if code, _, stderr := run(t, "", "--", "-odd.cpp"); code != ExitOK {
		t.Fatalf("exit code = %d, stderr = %s", code, stderr)
	}
	if got := read(t, filepath.Join(root, "-odd.cpp")); got != sorted {
		t.Error("-odd.cpp was not sorted")
	}
}
