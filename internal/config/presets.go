package config

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"
)

// DefaultPreset is used when neither a flag nor a config file names a preset.
const DefaultPreset = "default"

//go:embed presets/*.json
var presetFS embed.FS

// Presets returns the names of the built-in presets in alphabetical order.
func Presets() []string {
	files, _ := fs.Glob(presetFS, "presets/*.json")
	names := make([]string, 0, len(files))
	for _, f := range files {
		names = append(names, strings.TrimSuffix(path.Base(f), ".json"))
	}
	slices.Sort(names)
	return names
}

// Preset returns the built-in preset with the given name.
func Preset(name string) (*Config, error) {
	data, err := presetFS.ReadFile("presets/" + name + ".json")
	if err != nil {
		return nil, fmt.Errorf("unknown preset %q (available: %s)", name, strings.Join(Presets(), ", "))
	}
	cfg := defaults()
	if err := cfg.overlay(data); err != nil {
		return nil, fmt.Errorf("preset %s: %w", name, err)
	}
	cfg.Preset = name
	if err := cfg.compile(); err != nil {
		return nil, fmt.Errorf("preset %s: %w", name, err)
	}
	return cfg, nil
}

// defaults returns the settings shared by all presets.
func defaults() *Config {
	return &Config{
		Blocks: Preserve,
		Sort:   Lexical,
		Dedup:  true,
		ImplExtensions: []string{
			".c", ".cc", ".cpp", ".cxx", ".c++", ".cp", ".cu", ".m", ".mm",
			".inl", ".ipp", ".tpp", ".tcc",
		},
		MainSuffixes: []string{
			"_test", "_tests", "_unittest", "Test", "Tests", "_bench", "_benchmark",
			"_impl", "Impl", "-inl", "_inl", "_private", "_p",
		},
		Extensions: []string{
			".c", ".cc", ".cpp", ".cxx", ".c++", ".cppm", ".ixx",
			".h", ".hh", ".hpp", ".hxx", ".h++", ".inl", ".ipp", ".tpp",
			".cu", ".cuh", ".m", ".mm",
		},
		Exclude: []string{
			".*", "build", "build-*", "cmake-build-*", "out",
			"node_modules", "testdata", "third_party", "3rdparty", "vendor",
		},
	}
}
