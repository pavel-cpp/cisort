package finder

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestFind(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{
		"main.cpp", "README.md", "include/app/App.HPP", "src/gen/table.cc",
		"build/out.cpp", "cmake-build-debug/x.cpp", ".git/hook.c", "third_party/zlib.h",
	} {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	rules := func(string) (Rules, error) {
		return Rules{
			Extensions: []string{".cpp", "hpp", ".cc", ".h"},
			Exclude:    []string{".*", "build", "cmake-build-*", "third_party", "src/gen"},
		}, nil
	}

	readme := filepath.Join(root, "README.md")
	got, err := Find([]string{root, readme, readme}, rules)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		filepath.Join(root, "include", "app", "App.HPP"),
		filepath.Join(root, "main.cpp"),
		readme,
	}
	if !slices.Equal(got, want) {
		t.Errorf("Find() = %v, want %v", got, want)
	}
}

func TestFindErrors(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "a.cpp"), nil, 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Find([]string{filepath.Join(root, "missing")}, nil)
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("missing path: err = %v, want ErrNotExist", err)
	}

	bad := errors.New("bad config")
	files, err := Find([]string{root}, func(string) (Rules, error) { return Rules{}, bad })
	if !errors.Is(err, bad) || len(files) != 0 {
		t.Errorf("rules error: files = %v, err = %v", files, err)
	}

	_, err = Find([]string{root}, func(string) (Rules, error) { return Rules{Exclude: []string{"["}}, nil })
	if err == nil {
		t.Error("bad pattern: want an error")
	}
}
