package config

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/pavel-cpp/cisort/internal/headers"
)

func TestPresets(t *testing.T) {
	want := []string{"alpha", "default", "google", "llvm", "local-first", "qt", "topics"}
	if got := Presets(); !slices.Equal(got, want) {
		t.Fatalf("Presets() = %v, want %v", got, want)
	}
	for _, name := range want {
		cfg, err := Preset(name)
		if err != nil {
			t.Fatalf("Preset(%q): %v", name, err)
		}
		if cfg.Preset != name {
			t.Errorf("Preset(%q).Preset = %q", name, cfg.Preset)
		}
	}
	if _, err := Preset("nope"); err == nil || !strings.Contains(err.Error(), "available") {
		t.Errorf("Preset(nope) error = %v, want a list of presets", err)
	}
}

func TestTopicsListStandardHeaders(t *testing.T) {
	cfg, err := Preset("topics")
	if err != nil {
		t.Fatal(err)
	}
	for _, g := range cfg.Groups {
		for _, h := range g.Headers {
			if !strings.HasPrefix(h, "<") || !headers.Cpp.Contains(strings.Trim(h, "<>")) {
				t.Errorf("group %q: %s is not a C++ standard header", g.Name, h)
			}
		}
	}
}

func TestCategorize(t *testing.T) {
	tests := []struct {
		preset   string
		spelling string
		isMain   bool
		want     string
	}{
		{"default", `"pch.h"`, true, "Precompiled header"},
		{"default", `"StdAfx.h"`, false, "Precompiled header"},
		{"default", `"widget.h"`, true, "Main header"},
		{"default", "<stdio.h>", false, "C system headers"},
		{"default", "<sys/socket.h>", false, "C system headers"},
		{"default", "<cstdio>", false, "C++ standard library"},
		{"default", "<boost/any.hpp>", false, "Third-party"},
		{"default", `"app.h"`, false, "Project"},
		{"alpha", `"app.h"`, false, OtherGroup},
		{"google", "<zlib.h>", false, "C system headers"},
		{"google", "<vector>", false, "C++ standard library"},
		{"llvm", `"llvm/ADT/APInt.h"`, false, "LLVM project headers"},
		{"llvm", `"gtest/gtest.h"`, false, "System headers"},
		{"llvm", `"Local.h"`, false, "Local headers"},
		{"qt", "<QString>", false, "Qt"},
		{"qt", "<QtCore/qstring.h>", false, "Qt"},
		{"qt", "<QtWidgets>", false, "Qt"},
		{"qt", "<unistd.h>", false, "C and POSIX"},
		{"qt", "<boost/any.hpp>", false, "Third-party"},
		{"topics", "<vector>", false, "Containers"},
		{"topics", "<cmath>", false, "C library"},
		{"topics", "<math.h>", false, "C library"},
	}
	for _, tt := range tests {
		cfg, err := Preset(tt.preset)
		if err != nil {
			t.Fatal(err)
		}
		if got := cfg.Categorize(tt.spelling, tt.isMain).Name; got != tt.want {
			t.Errorf("%s: Categorize(%s, %v) = %q, want %q", tt.preset, tt.spelling, tt.isMain, got, tt.want)
		}
	}
}

func TestCategorizePriority(t *testing.T) {
	cfg, err := Preset("llvm")
	if err != nil {
		t.Fatal(err)
	}
	local := cfg.Categorize(`"Local.h"`, false)
	project := cfg.Categorize(`"llvm/IR/Function.h"`, false)
	system := cfg.Categorize("<vector>", false)
	if !(local.Priority < project.Priority && project.Priority < system.Priority) {
		t.Errorf("priorities local=%d project=%d system=%d, want ascending", local.Priority, project.Priority, system.Priority)
	}
}

func TestParse(t *testing.T) {
	cfg, err := Parse([]byte(`{"preset": "google", "comments": true, "sort": "natural"}`))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Preset != "google" || cfg.Blocks != Regroup || !cfg.Comments || cfg.Sort != Natural {
		t.Errorf("Parse() = %+v, want google with comments and natural sort", cfg)
	}
	if len(cfg.Groups) == 0 || len(cfg.Extensions) == 0 {
		t.Error("Parse() lost the preset groups or the default extensions")
	}

	cfg, err = Parse([]byte(`{"groups": [{"name": "Mine", "match": "^\"mine/"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Preset != DefaultPreset || len(cfg.Groups) != 1 || cfg.Groups[0].Builtin != "" {
		t.Errorf("groups were not replaced: %+v", cfg.Groups)
	}
	if got := cfg.Categorize(`"mine/a.h"`, false).Name; got != "Mine" {
		t.Errorf("Categorize = %q, want Mine", got)
	}
}

func TestParseErrors(t *testing.T) {
	tests := map[string]string{
		`{"blocks": "sideways"}`:                         "blocks",
		`{"sort": "random"}`:                             "sort",
		`{"comment": true}`:                              "unknown field",
		`{"preset": "nope"}`:                             "unknown preset",
		`{"groups": [{"match": "^<"}]}`:                  "name is required",
		`{"groups": [{"name": "X"}]}`:                    "set builtin, match or headers",
		`{"groups": [{"name": "X", "builtin": "rust"}]}`: "unknown builtin",
		`{"groups": [{"name": "X", "match": "("}]}`:      "missing closing",
		`{} {}`: "after top-level value",
		`[`:     "unexpected end",
	}
	for in, want := range tests {
		_, err := Parse([]byte(in))
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("Parse(%s) error = %v, want it to mention %q", in, err, want)
		}
	}
}

func TestResolver(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "a", FileName), []byte(`{"preset": "qt"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	broken := filepath.Join(root, "broken")
	if err := os.MkdirAll(broken, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(broken, FileName), []byte(`{`), 0o644); err != nil {
		t.Fatal(err)
	}

	fallback, err := Preset("alpha")
	if err != nil {
		t.Fatal(err)
	}
	r := NewResolver(fallback, func(c *Config) { c.Comments = true })

	cfg, err := r.For(sub)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Preset != "qt" || !cfg.Comments {
		t.Errorf("For(a/b) = preset %q comments %v, want qt with comments", cfg.Preset, cfg.Comments)
	}
	if again, _ := r.For(filepath.Join(root, "a")); again != cfg {
		t.Error("For(a) did not reuse the config loaded for a/b")
	}
	if cfg, _ := r.For(root); cfg != fallback || !cfg.Comments {
		t.Error("For(root) did not return the adjusted fallback")
	}
	if _, err := r.For(broken); err == nil || !strings.Contains(err.Error(), FileName) {
		t.Errorf("For(broken) error = %v, want it to name the file", err)
	}
}
