package sorter

import (
	"strings"
	"testing"

	"github.com/pavel-cpp/cisort/internal/config"
)

func preset(t *testing.T, name string, adjust ...func(*config.Config)) *config.Config {
	t.Helper()
	cfg, err := config.Preset(name)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range adjust {
		f(cfg)
	}
	return cfg
}

// lines joins its arguments with "\n" and terminates the result with "\n".
func lines(s ...string) string {
	return strings.Join(s, "\n") + "\n"
}

func TestSort(t *testing.T) {
	tests := []struct {
		name     string
		preset   string
		adjust   func(*config.Config)
		filename string
		in, want string
	}{
		{
			name:   "readme example",
			preset: "alpha",
			in: lines(
				"#include <iostream>", "#include <vector>", "#include <stdio.h>", "#include <algorithm>",
				"",
				`#include "mylib.h"`, `#include "b_lib.h"`, `#include "a_first_lib.h"`,
				"",
				"#include <external/lib/main.hpp>", "#include <external/lib/abuse.hpp>", "#include <external/lib/func.hpp>",
				"",
				"int main(){}",
			),
			want: lines(
				"#include <algorithm>", "#include <iostream>", "#include <stdio.h>", "#include <vector>",
				"",
				`#include "a_first_lib.h"`, `#include "b_lib.h"`, `#include "mylib.h"`,
				"",
				"#include <external/lib/abuse.hpp>", "#include <external/lib/func.hpp>", "#include <external/lib/main.hpp>",
				"",
				"int main(){}",
			),
		},
		{
			name: "groups order a block",
			in:   lines(`#include "app.h"`, "#include <vector>", "#include <boost/any.hpp>", "#include <stdio.h>", "#include <unistd.h>"),
			want: lines("#include <stdio.h>", "#include <unistd.h>", "#include <vector>", "#include <boost/any.hpp>", `#include "app.h"`),
		},
		{
			name: "keeps CRLF",
			in:   "#include <b>\r\n#include <a>\r\nint x;\r\n",
			want: "#include <a>\r\n#include <b>\r\nint x;\r\n",
		},
		{
			name: "no newline at end of file",
			in:   "#include <b>\n#include <a>",
			want: "#include <a>\n#include <b>",
		},
		{
			name: "keeps byte order mark",
			in:   "\xEF\xBB\xBF#include <b>\n#include <a>\n",
			want: "\xEF\xBB\xBF#include <a>\n#include <b>\n",
		},
		{
			name: "ignores includes in block comments",
			in:   lines("/*", "#include <b>", "#include <a>", "*/", "#include <d>", "#include <c>"),
			want: lines("/*", "#include <b>", "#include <a>", "*/", "#include <c>", "#include <d>"),
		},
		{
			name: "ignores includes between off and on markers",
			in:   lines("// cisort: off", "#include <b>", "#include <a>", "// cisort: on", "#include <d>", "#include <c>"),
			want: lines("// cisort: off", "#include <b>", "#include <a>", "// cisort: on", "#include <c>", "#include <d>"),
		},
		{
			name: "honours clang-format markers",
			in:   lines("// clang-format off", "#include <b>", "#include <a>", "// clang-format on"),
			want: lines("// clang-format off", "#include <b>", "#include <a>", "// clang-format on"),
		},
		{
			name: "keep marker pins a line",
			in:   lines("#include <d>", "#include <c>", "#include <winsock2.h> // cisort: keep", "#include <b>", "#include <a>"),
			want: lines("#include <c>", "#include <d>", "#include <winsock2.h> // cisort: keep", "#include <a>", "#include <b>"),
		},
		{
			name: "directives split runs",
			in:   lines("#include <b>", "#ifdef X", "#include <z>", "#include <y>", "#endif", "#include <a>"),
			want: lines("#include <b>", "#ifdef X", "#include <y>", "#include <z>", "#endif", "#include <a>"),
		},
		{
			name: "comments between includes travel",
			in:   lines("#include <c>", "// for b", "#include <b>", "#include <a> // trailing"),
			want: lines("#include <a> // trailing", "// for b", "#include <b>", "#include <c>"),
		},
		{
			name: "comments above a block stay",
			in:   lines("// Copyright", "#include <b>", "#include <a>", "", "// Section", "#include <d>", "#include <c>"),
			want: lines("// Copyright", "#include <a>", "#include <b>", "", "// Section", "#include <c>", "#include <d>"),
		},
		{
			name: "code after an include is not touched",
			in:   lines("#include <b> int x;", "#include <a>"),
			want: lines("#include <b> int x;", "#include <a>"),
		},
		{
			name: "removes duplicates",
			in:   lines("#include <b>", "#include <a>", "#include <b>", "#import <b>"),
			want: lines("#include <a>", "#import <b>", "#include <b>"),
		},
		{
			name: "keeps commented duplicates",
			in:   lines("#include <a>", "#include <a> // again"),
			want: lines("#include <a>", "#include <a> // again"),
		},
		{
			name:   "dedup can be disabled",
			adjust: func(c *config.Config) { c.Dedup = false },
			in:     lines("#include <a>", "#include <a>"),
			want:   lines("#include <a>", "#include <a>"),
		},
		{
			name:     "main header first",
			filename: "src/widget.cpp",
			in:       lines("#include <vector>", `#include "app.h"`, `#include "ui/widget.h"`),
			want:     lines(`#include "ui/widget.h"`, "#include <vector>", `#include "app.h"`),
		},
		{
			name:     "main header of a test",
			filename: "widget_test.cc",
			in:       lines(`#include "app.h"`, `#include "widget.h"`),
			want:     lines(`#include "widget.h"`, `#include "app.h"`),
		},
		{
			name:     "headers have no main header",
			filename: "widget.h",
			in:       lines(`#include "widget.h"`, `#include "app.h"`),
			want:     lines(`#include "app.h"`, `#include "widget.h"`),
		},
		{
			name:     "main header in angle brackets",
			filename: "src/widget.cpp",
			in:       lines("#include <vector>", "#include <mylib/ui/Widget.hpp>"),
			want:     lines("#include <mylib/ui/Widget.hpp>", "#include <vector>"),
		},
		{
			name:     "standard headers are never main",
			filename: "string.cpp",
			in:       lines(`#include "app.h"`, "#include <string>", "#include <string.h>"),
			want:     lines("#include <string.h>", "#include <string>", `#include "app.h"`),
		},
		{
			name:     "main header of an implementation file",
			filename: "widget_impl.cpp",
			in:       lines("#include <vector>", `#include "widget.h"`),
			want:     lines(`#include "widget.h"`, "#include <vector>"),
		},
		{
			name:     "main header of an inline file",
			filename: "widget.inl",
			in:       lines("#include <vector>", `#include "widget.hpp"`),
			want:     lines(`#include "widget.hpp"`, "#include <vector>"),
		},
		{
			name:     "main header of an -inl header",
			filename: "widget-inl.h",
			in:       lines("#include <vector>", `#include "widget.h"`),
			want:     lines(`#include "widget.h"`, "#include <vector>"),
		},
		{
			name:     "main header of a private header",
			filename: "qwidget_p.h",
			in:       lines("#include <QtCore/qobject.h>", `#include "qwidget.h"`),
			want:     lines(`#include "qwidget.h"`, "#include <QtCore/qobject.h>"),
		},
		{
			name:     "main header with the alpha preset",
			preset:   "alpha",
			filename: "widget.cc",
			in:       lines(`#include "app.h"`, `#include "widget.h"`),
			want:     lines(`#include "widget.h"`, `#include "app.h"`),
		},
		{
			name:     "main header detection can be narrowed",
			adjust:   func(c *config.Config) { c.ImplExtensions = []string{"cpp"} },
			filename: "widget.cc",
			in:       lines(`#include "widget.h"`, "#include <vector>"),
			want:     lines("#include <vector>", `#include "widget.h"`),
		},
		{
			name: "precompiled header first",
			in:   lines("#include <vector>", `#include "StdAfx.h"`),
			want: lines(`#include "StdAfx.h"`, "#include <vector>"),
		},
		{
			name: "keeps directive spelling",
			in:   lines("  #  include <b>", "#import <a>", "#include_next <c>"),
			want: lines("#import <a>", "  #  include <b>", "#include_next <c>"),
		},
		{
			name:   "regroup",
			preset: "google",
			in: lines(
				"#include <vector>",
				`#include "base/util.h"`,
				"",
				"#include <stdio.h>",
				"#include <absl/strings/str_cat.h>",
				"",
				"",
				"int main() {}",
			),
			want: lines(
				"#include <absl/strings/str_cat.h>",
				"#include <stdio.h>",
				"",
				"#include <vector>",
				"",
				`#include "base/util.h"`,
				"",
				"",
				"int main() {}",
			),
		},
		{
			name:   "regroup with comments",
			preset: "default",
			adjust: func(c *config.Config) { c.Blocks, c.Comments = config.Regroup, true },
			in:     lines("#include <vector>", `#include "a.h"`, "#include <map>"),
			want:   lines("// C++ standard library", "#include <map>", "#include <vector>", "", "// Project", `#include "a.h"`),
		},
		{
			name:   "label above an attached comment",
			preset: "local-first",
			adjust: func(c *config.Config) { c.Comments = true },
			in:     lines("#include <vector>", "// why", `#include "a.h"`),
			want:   lines("// Project", "// why", `#include "a.h"`, "", "// C++ standard library", "#include <vector>"),
		},
		{
			name:   "no labels for a single group",
			preset: "default",
			adjust: func(c *config.Config) { c.Blocks, c.Comments = config.Regroup, true },
			in:     lines("// C++ standard library", "#include <vector>", "#include <map>"),
			want:   lines("#include <map>", "#include <vector>"),
		},
		{
			name:   "regroup keeps a user comment as a boundary",
			preset: "google",
			in:     lines("#include <vector>", "", "// Third party", "#include <zlib.h>", "#include <map>"),
			want:   lines("#include <vector>", "", "// Third party", "#include <zlib.h>", "", "#include <map>"),
		},
		{
			name:   "merge",
			preset: "llvm",
			in:     lines("#include <vector>", "", `#include "llvm/ADT/APInt.h"`, "", `#include "Local.h"`),
			want:   lines(`#include "Local.h"`, `#include "llvm/ADT/APInt.h"`, "#include <vector>"),
		},
		{
			name:   "case insensitive",
			preset: "alpha",
			adjust: func(c *config.Config) { c.Sort = config.CaseInsensitive },
			in:     lines(`#include "b.h"`, `#include "A.h"`, `#include "a.h"`),
			want:   lines(`#include "A.h"`, `#include "a.h"`, `#include "b.h"`),
		},
		{
			name:   "natural",
			preset: "alpha",
			adjust: func(c *config.Config) { c.Sort = config.Natural },
			in:     lines(`#include "v10.h"`, `#include "v2.h"`, `#include "V1.h"`),
			want:   lines(`#include "V1.h"`, `#include "v2.h"`, `#include "v10.h"`),
		},
		{
			name: "digit separators and strings do not open comments",
			in:   lines(`int x = 1'000; const char* s = "/*";`, "#include <b>", "#include <a>"),
			want: lines(`int x = 1'000; const char* s = "/*";`, "#include <a>", "#include <b>"),
		},
		{
			name: "no includes",
			in:   lines("int main() {}"),
			want: lines("int main() {}"),
		},
		{
			name: "empty",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name := tt.preset
			if name == "" {
				name = config.DefaultPreset
			}
			var adjust []func(*config.Config)
			if tt.adjust != nil {
				adjust = append(adjust, tt.adjust)
			}
			cfg := preset(t, name, adjust...)

			got := string(Sort([]byte(tt.in), tt.filename, cfg))
			if got != tt.want {
				t.Errorf("Sort() =\n%s\nwant\n%s", got, tt.want)
			}
			if again := string(Sort([]byte(got), tt.filename, cfg)); again != got {
				t.Errorf("Sort() is not idempotent, second run =\n%s", again)
			}
		})
	}
}
