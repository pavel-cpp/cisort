package cli

import (
	"bytes"
	"strings"
	"testing"
)

const sampleDiff = "--- a/x.cpp\n+++ b/x.cpp\n@@ -1,2 +1,2 @@\n-#include <b>\n #include <a>\n+#include <b>\n\\ No newline at end of file\n"

func TestColorDiff(t *testing.T) {
	got := newPalette(true).colorDiff(sampleDiff)
	for _, want := range []string{
		"\x1b[1m--- a/x.cpp\x1b[",
		"\x1b[1m+++ b/x.cpp\x1b[",
		"\x1b[36m@@ -1,2 +1,2 @@\x1b[",
		"\x1b[31m-#include <b>\x1b[",
		"\n #include <a>\n",
		"\x1b[32m+#include <b>\x1b[",
		"\x1b[2m\\ No newline at end of file\x1b[",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("colored diff does not contain %q:\n%q", want, got)
		}
	}
}

func TestColorDiffDisabled(t *testing.T) {
	if got := newPalette(false).colorDiff(sampleDiff); got != sampleDiff {
		t.Errorf("colorDiff() without color changed the diff:\n%q", got)
	}
}

func TestNoColorOutsideTerminals(t *testing.T) {
	var buf bytes.Buffer
	if w, ok := colorWriter(&buf, false); ok || w != &buf {
		t.Error("colorWriter() enabled color for a buffer")
	}

	root := project(t, map[string]string{"a.cpp": unsorted})
	_, stdout, _ := run(t, "", "-diff", root)
	if strings.Contains(stdout, "\x1b[") {
		t.Errorf("diff written to a pipe contains color codes:\n%q", stdout)
	}
}

func TestNoColorAndNoProgressFlags(t *testing.T) {
	root := project(t, map[string]string{"a.cpp": unsorted})
	code, _, stderr := run(t, "", "-no-color", "-no-progress", "-check", root)
	if code != ExitUnsorted {
		t.Errorf("exit code = %d, want %d; stderr = %s", code, ExitUnsorted, stderr)
	}
}

func TestProgress(t *testing.T) {
	var buf bytes.Buffer
	bar := newProgress(&buf, 2, "sorting", false)
	bar.Add(1)
	bar.Add(1)
	bar.Finish()
	if !strings.Contains(buf.String(), "sorting") {
		t.Errorf("progress output = %q, want the description", buf.String())
	}
}
