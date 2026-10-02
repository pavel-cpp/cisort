// Package config describes how cisort orders #include directives and loads
// that description from built-in presets and .cisort.json files.
package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/pavel-cpp/cisort/internal/headers"
)

// FileName is the name of the configuration file looked up next to sources.
const FileName = ".cisort.json"

// OtherGroup names the implicit group of includes that match no group.
const OtherGroup = "Other"

// BuiltinMain is the builtin matching the main header of a source file,
// e.g. "foo.h" for foo.cpp.
const BuiltinMain = "main"

// Blocks controls what happens to blank lines between includes.
type Blocks string

const (
	// Preserve sorts every blank-line separated block on its own.
	Preserve Blocks = "preserve"
	// Merge joins adjacent blocks into a single sorted block.
	Merge Blocks = "merge"
	// Regroup joins adjacent blocks and splits them again by group.
	Regroup Blocks = "regroup"
)

// SortOrder controls how includes are compared within a group.
type SortOrder string

const (
	// Lexical compares bytes, so "Z.h" sorts before "a.h".
	Lexical SortOrder = "lexical"
	// CaseInsensitive ignores letter case.
	CaseInsensitive SortOrder = "case-insensitive"
	// Natural ignores letter case and compares digit runs as numbers,
	// so "v2.h" sorts before "v10.h".
	Natural SortOrder = "natural"
)

// Group is a category of includes. An include belongs to the first group
// it matches; groups are emitted in ascending priority.
type Group struct {
	// Name labels the group in comments.
	Name string `json:"name"`
	// Builtin is one of "main", "c", "cpp" or "posix".
	Builtin string `json:"builtin,omitempty"`
	// Match is a regular expression applied to the include spelling with its
	// delimiters, e.g. `<vector>` or `"foo/bar.h"`.
	Match string `json:"match,omitempty"`
	// Headers lists exact spellings with delimiters, e.g. "<vector>".
	Headers []string `json:"headers,omitempty"`
	// Priority orders groups. It defaults to the group's index.
	Priority *int `json:"priority,omitempty"`

	re      *regexp.Regexp
	headers map[string]struct{}
}

// Config is a complete sorting style.
type Config struct {
	// Preset is the built-in preset this configuration extends.
	Preset string `json:"preset,omitempty"`
	// Blocks controls blank lines between includes.
	Blocks Blocks `json:"blocks"`
	// Comments labels groups with a "// <name>" line when regrouping.
	Comments bool `json:"comments"`
	// Sort controls ordering within a group.
	Sort SortOrder `json:"sort"`
	// Dedup removes repeated includes.
	Dedup bool `json:"dedup"`
	// ImplExtensions lists the extensions of implementation files. Their main
	// header goes first: widget.cpp and widget.inl pair with widget.h.
	ImplExtensions []string `json:"impl_extensions"`
	// MainSuffixes are stripped from a file name when looking for its main
	// header, so widget_test.cpp and widget_impl.h still pair with widget.h.
	// A file of any extension with such a suffix has a main header.
	MainSuffixes []string `json:"main_suffixes"`
	// Groups are the include categories.
	Groups []Group `json:"groups"`
	// Extensions lists the file extensions processed when walking directories.
	Extensions []string `json:"extensions"`
	// Exclude lists glob patterns of files and directories skipped while walking.
	Exclude []string `json:"exclude"`

	labels        map[int]string
	otherPriority int
	usesMain      bool
}

// Category is the group an include falls into.
type Category struct {
	Priority int
	Name     string
}

// Categorize returns the category of an include spelled with its delimiters.
// isMain reports whether the include is the main header of the file.
func (c *Config) Categorize(spelling string, isMain bool) Category {
	for i := range c.Groups {
		g := &c.Groups[i]
		if g.matches(spelling, isMain) {
			p := g.priority(i)
			return Category{Priority: p, Name: c.labels[p]}
		}
	}
	return Category{Priority: c.otherPriority, Name: OtherGroup}
}

// UsesMain reports whether some group matches the main header.
func (c *Config) UsesMain() bool {
	return c.usesMain
}

// Labels returns the names of all groups, including OtherGroup.
func (c *Config) Labels() []string {
	names := []string{OtherGroup}
	for _, g := range c.Groups {
		names = append(names, g.Name)
	}
	return names
}

func (g *Group) matches(spelling string, isMain bool) bool {
	switch g.Builtin {
	case "":
	case BuiltinMain:
		if isMain {
			return true
		}
	default:
		if strings.HasPrefix(spelling, "<") && headers.Builtin(g.Builtin).Contains(spelling[1:len(spelling)-1]) {
			return true
		}
	}
	if g.re != nil && g.re.MatchString(spelling) {
		return true
	}
	_, ok := g.headers[spelling]
	return ok
}

func (g *Group) priority(index int) int {
	if g.Priority != nil {
		return *g.Priority
	}
	return index
}

// compile validates c and prepares it for Categorize.
func (c *Config) compile() error {
	switch c.Blocks {
	case Preserve, Merge, Regroup:
	default:
		return fmt.Errorf("blocks: unknown value %q (want %q, %q or %q)", c.Blocks, Preserve, Merge, Regroup)
	}
	switch c.Sort {
	case Lexical, CaseInsensitive, Natural:
	default:
		return fmt.Errorf("sort: unknown value %q (want %q, %q or %q)", c.Sort, Lexical, CaseInsensitive, Natural)
	}

	c.labels = make(map[int]string)
	c.otherPriority = 0
	c.usesMain = false
	for i := range c.Groups {
		g := &c.Groups[i]
		if g.Name == "" {
			return fmt.Errorf("groups[%d]: name is required", i)
		}
		if g.Builtin == "" && g.Match == "" && len(g.Headers) == 0 {
			return fmt.Errorf("group %q: set builtin, match or headers", g.Name)
		}
		if g.Builtin == BuiltinMain {
			c.usesMain = true
		} else if g.Builtin != "" && !headers.Builtin(g.Builtin).Valid() {
			return fmt.Errorf("group %q: unknown builtin %q (want main, c, cpp or posix)", g.Name, g.Builtin)
		}
		g.re = nil
		if g.Match != "" {
			re, err := regexp.Compile(g.Match)
			if err != nil {
				return fmt.Errorf("group %q: %w", g.Name, err)
			}
			g.re = re
		}
		g.headers = make(map[string]struct{}, len(g.Headers))
		for _, h := range g.Headers {
			g.headers[h] = struct{}{}
		}

		p := g.priority(i)
		if _, ok := c.labels[p]; !ok {
			c.labels[p] = g.Name
		}
		c.otherPriority = max(c.otherPriority, p+1)
	}
	return nil
}

// overlay decodes the JSON object data over c. Fields missing from data keep
// their values; a present "groups" replaces the groups entirely.
func (c *Config) overlay(data []byte) error {
	groups := c.Groups
	c.Groups = nil
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(c); err != nil {
		return err
	}
	if c.Groups == nil {
		c.Groups = groups
	}
	return nil
}

// Parse decodes a configuration file. The style starts from the preset named
// by its "preset" field, or from the default preset.
func Parse(data []byte) (*Config, error) {
	var head struct {
		Preset string `json:"preset"`
	}
	if err := json.Unmarshal(data, &head); err != nil {
		return nil, err
	}
	if head.Preset == "" {
		head.Preset = DefaultPreset
	}
	cfg, err := Preset(head.Preset)
	if err != nil {
		return nil, err
	}
	if err := cfg.overlay(data); err != nil {
		return nil, err
	}
	if err := cfg.compile(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Load reads and parses the configuration file at path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg, err := Parse(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return cfg, nil
}
