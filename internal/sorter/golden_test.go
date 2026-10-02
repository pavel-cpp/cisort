package sorter

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/pavel-cpp/cisort/internal/config"
)

var update = flag.Bool("update", false, "rewrite the golden files in testdata/golden")

// TestPresetsGolden sorts testdata/widget.cpp with every preset and compares
// the result with testdata/golden/<preset>.cpp.
func TestPresetsGolden(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("testdata", "widget.cpp"))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range config.Presets() {
		t.Run(name, func(t *testing.T) {
			cfg := preset(t, name)
			got := Sort(src, "widget.cpp", cfg)
			if again := Sort(got, "widget.cpp", cfg); string(again) != string(got) {
				t.Errorf("not idempotent, second run:\n%s", again)
			}

			golden := filepath.Join("testdata", "golden", name+".cpp")
			if *update {
				if err := os.WriteFile(golden, got, 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("%v (run go test -update to create it)", err)
			}
			if string(got) != string(want) {
				t.Errorf("Sort() =\n%s\nwant\n%s", got, want)
			}
		})
	}
}
