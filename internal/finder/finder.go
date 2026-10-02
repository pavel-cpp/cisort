// Package finder discovers the C and C++ files cisort should process.
package finder

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Rules select the files processed inside one directory.
type Rules struct {
	// Extensions lists accepted file extensions, e.g. ".cpp".
	Extensions []string
	// Exclude lists glob patterns. A pattern without a slash matches the base
	// name of a file or directory; a pattern with a slash matches its path
	// relative to the walked root.
	Exclude []string
}

// RulesFunc returns the rules for the entries of dir.
type RulesFunc func(dir string) (Rules, error)

// Find returns the files named by paths, without duplicates. A file is
// returned as is, whatever its extension. A directory is walked recursively
// and yields the files accepted by the rules of their directory. Find skips
// what it cannot read and reports it in the returned error.
func Find(paths []string, rules RulesFunc) ([]string, error) {
	f := finder{rules: rules, seen: make(map[string]bool), cache: make(map[string]cached)}
	for _, p := range paths {
		info, err := os.Stat(p)
		switch {
		case err != nil:
			f.errs = append(f.errs, err)
		case info.IsDir():
			f.walk(p)
		default:
			f.add(p)
		}
	}
	return f.files, errors.Join(f.errs...)
}

type finder struct {
	rules RulesFunc
	files []string
	errs  []error
	seen  map[string]bool
	cache map[string]cached
}

type cached struct {
	rules Rules
	err   error
}

func (f *finder) add(p string) {
	if !f.seen[p] {
		f.seen[p] = true
		f.files = append(f.files, p)
	}
}

func (f *finder) walk(root string) {
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			f.errs = append(f.errs, err)
			return nil
		}
		if p == root {
			return nil
		}
		r, ok := f.rulesFor(filepath.Dir(p))
		if !ok {
			return fs.SkipDir
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		skip, err := excluded(r.Exclude, d.Name(), filepath.ToSlash(rel))
		if err != nil {
			return err
		}
		switch {
		case skip && d.IsDir():
			return fs.SkipDir
		case skip, d.IsDir(), !d.Type().IsRegular():
			return nil
		case hasExt(r.Extensions, d.Name()):
			f.add(p)
		}
		return nil
	})
	if err != nil {
		f.errs = append(f.errs, err)
	}
}

// rulesFor returns the rules of dir, reporting a failure only once.
func (f *finder) rulesFor(dir string) (Rules, bool) {
	c, ok := f.cache[dir]
	if !ok {
		c.rules, c.err = f.rules(dir)
		f.cache[dir] = c
		if c.err != nil {
			f.errs = append(f.errs, c.err)
		}
	}
	return c.rules, c.err == nil
}

func excluded(patterns []string, name, rel string) (bool, error) {
	for _, pattern := range patterns {
		subject := name
		if strings.Contains(pattern, "/") {
			subject = rel
		}
		ok, err := path.Match(pattern, subject)
		if err != nil {
			return false, fmt.Errorf("exclude pattern %q: %w", pattern, err)
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

func hasExt(exts []string, name string) bool {
	ext := filepath.Ext(name)
	for _, e := range exts {
		if strings.EqualFold(ext, "."+strings.TrimPrefix(e, ".")) {
			return true
		}
	}
	return false
}
