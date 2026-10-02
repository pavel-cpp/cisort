# cisort

`cisort` sorts `#include` directives in C and C++ projects. It is a single
static binary with no dependencies, fast enough to run on every save and
strict enough to run in CI.

```c++
#include <iostream>                         #include <stdio.h>
#include <vector>                           #include <algorithm>
#include <stdio.h>                          #include <iostream>
#include <algorithm>                        #include <vector>
#include <vector>
                                    →       #include "a_first_lib.h"
#include "mylib.h"                          #include "b_lib.h"
#include "b_lib.h"                          #include "mylib.h"
#include "a_first_lib.h"
```

## Installation

```shell
go install github.com/pavel-cpp/cisort/cmd/cisort@latest
```

This needs Go 1.27 or newer and puts `cisort` into `$(go env GOPATH)/bin`.
Make sure that directory is on your `PATH`.

## Usage

```shell
cisort                      # sort every C/C++ file under the current directory
cisort src include main.cpp # sort the given directories and files
cisort -check -diff .       # change nothing; show the diff, exit 1 if unsorted
cisort -preset google -c .  # regroup in Google style with group comments
cisort - < a.cpp > b.cpp    # filter stdin to stdout
```

| Flag | Meaning |
|---|---|
| `-check` | Do not write files; list unsorted files and exit with status 1 if there are any. |
| `-diff` | Do not write files; print a unified diff. |
| `-g` | Regroup includes (see [Blocks](#blocks)). |
| `-c` | Label groups with comments; implies `-g`. |
| `-preset name` | Use a built-in preset and ignore config files. |
| `-config file` | Use this config file instead of looking up `.cisort.json`. |
| `-ext .cpp,.h` | File extensions to process when walking directories. |
| `-exclude glob` | Skip matching files and directories; repeatable. |
| `-stdin-filename path` | Name of the file read from stdin, used to find its main header and config. |
| `-j n` | Number of files processed in parallel (default: number of CPUs). |
| `-v` | Print changed files and a summary. |
| `-no-color` | Disable colored output. |
| `-no-progress` | Do not show the progress bar. |
| `-list-presets` | Print the built-in presets. |
| `-print-config` | Print the effective configuration as JSON. |
| `-version` | Print the version. |

Flags may appear before or after paths. Everything after `--` is a path.

In a terminal, diffs and file lists are colored and a progress bar shows how
many files are done. The bar is cleared when the run ends. Output that goes to
a pipe or a file is never colored, and the bar is drawn only when stderr is a
terminal, so scripts and CI logs stay clean. Setting the
[`NO_COLOR`](https://no-color.org) environment variable works like
`-no-color`.

Exit status is 0 on success, 1 when `-check` found unsorted files, and 2 on
errors.

Directories are walked recursively. Hidden directories, `build`, `build-*`,
`cmake-build-*`, `out`, `node_modules`, `testdata`, `third_party`, `3rdparty` and
`vendor` are skipped. Files named explicitly on the command line are always
processed, whatever their extension.

## How it works

`cisort` only moves include lines and the comments that belong to them.
Everything else stays byte for byte as it was. That includes line endings
(LF and CRLF), the UTF-8 byte order mark, any encoding, indentation and the
spelling of directives (`#  include`, `#import`, `#include_next`). Files are
rewritten atomically, and only when something changed.

**Runs and blocks.** A *run* is a sequence of include lines, possibly
separated by blank lines. Blank lines split a run into *blocks*. Any other
line ends the run: code, `#if`/`#ifdef`, `#pragma`, `#define`, and so on. So
includes never jump across conditional compilation:

```c++
#include <b.h>        #include <b.h>
#ifdef _WIN32         #ifdef _WIN32
#include <z.h>   →    #include <y.h>
#include <y.h>        #include <z.h>
#endif                #endif
```

**Comments.** A comment line between two includes moves with the include
below it. A trailing comment (`#include <x> // why`) moves with its line. A
comment above the first include of a block, such as a license header or a
section title, stays where it is.

```c++
#include <c>                     #include <a> // trailing
// needed for b                  // needed for b
#include <b>               →     #include <b>
#include <a> // trailing         #include <c>
```

**Order.** Every include belongs to a *group* (C++ standard library,
third-party, project, …). Includes are sorted by group, then by their path
spelled with delimiters, so `"x.h"` and `<x.h>` differ.

**Main header.** A file that implements a class includes the header that
declares it first, so that the header is proven to compile on its own. Every
preset puts this *main header* first. The include whose file name (without
directory and extension, ignoring case) matches the file being sorted is the
main header:

| File being sorted | Main header |
|---|---|
| `widget.cpp`, `widget.cc`, `widget.c`, `widget.mm`, … | `"widget.h"`, `"ui/widget.hpp"`, `<mylib/widget.h>` |
| `widget.inl`, `widget.ipp`, `widget.tpp`, `widget.tcc` (template implementations) | `"widget.h"` |
| `widget_impl.cpp`, `widgetImpl.cpp`, `widget_test.cc`, `widget_unittest.cpp`, `WidgetTest.cpp` | `"widget.h"` (and `"widget_test.h"` for the test) |
| `widget_impl.h`, `widget-inl.h`, `widget_p.h`, `widget_private.h` | `"widget.h"` |
| `widget.h` | none |

Standard headers are never main headers, so `string.cpp` does not pair with
`<string>`. Configure the file kinds with `impl_extensions` and
`main_suffixes`.

**Precompiled headers.** `"pch.h"`, `"stdafx.h"` and similar always come
first.

**Duplicates.** A repeated include is removed, unless removing it would lose a
comment.

**Markers.**

| Marker | Effect |
|---|---|
| `// cisort: off` … `// cisort: on` | Leave everything between the markers alone. `// clang-format off/on` works too. |
| `#include <x> // cisort: keep` | Pin this line. It splits the run in two. |

Use `keep` for headers whose order matters, for example
`<winsock2.h>` before `<windows.h>`.

**Comments and strings.** Includes inside `/* … */` comments are ignored.
String and character literals are understood, including digit separators such
as `1'000`.

### Blocks

The `blocks` setting decides what happens to blank lines:

| Value | Effect |
|---|---|
| `preserve` | Sort each block on its own and keep the blank lines. |
| `merge` | Join the blocks of a run into one sorted block. |
| `regroup` | Join the blocks, then split them again by group with one blank line between groups. With `comments`, each group gets a `// <name>` label. Labels are recognised on the next run, so the output is stable. |

## Presets

`cisort -list-presets` lists them. `cisort -preset <name> -print-config`
shows any of them in full.

| Preset | Blocks | Group order |
|---|---|---|
| `default` | preserve | precompiled · main · C system (C, POSIX) · C++ std · third-party `<…>` · project `"…"` |
| `alpha` | preserve | precompiled · main · everything else alphabetically (close to the original Python cisort) |
| `google` | regroup | [Google style](https://google.github.io/styleguide/cppguide.html#Names_and_Order_of_Includes): related header · C system (`<*.h>`) · C++ std · other libraries · project |
| `llvm` | merge | [LLVM style](https://llvm.org/docs/CodingStandards.html#include-style): main module · local `"…"` · LLVM project · system |
| `qt` | regroup | main · project · Qt (`<QString>`, `<QtCore/…>`) · third-party · C++ std · C/POSIX |
| `local-first` | regroup | main · project · third-party · C++ std · C system (Lakos: "local to global", which catches headers that are not self-contained) |
| `topics` | regroup + comments | main · C++ std by topic (Containers, Strings, Streams and I/O, Concurrency, Memory, Utilities, Algorithms, Types, Time, Errors) · C library · POSIX · third-party · project |

What `topics` produces:

```c++
// Main header
#include "widget.h"

// Containers
#include <map>
#include <vector>

// Concurrency
#include <mutex>

// Third-party
#include <QString>
#include <boost/asio.hpp>

// Project
#include "core/model.h"
```

All presets are tested against the same input. The expected output of each
lives in [`internal/sorter/testdata/golden`](internal/sorter/testdata/golden).

## Configuration

`cisort` uses the nearest `.cisort.json` in the file's directory or one of its
parents, so different parts of a repository can use different styles. To
start a config from a preset:

```shell
cisort -preset google -print-config > .cisort.json
```

A config file holds only what differs from its preset:

```json
{
  "preset": "google",
  "comments": true,
  "sort": "natural",
  "exclude": [".*", "build", "external/*"]
}
```

| Key | Default | Meaning |
|---|---|---|
| `preset` | `"default"` | Preset to start from. |
| `blocks` | from preset | `"preserve"`, `"merge"` or `"regroup"`. |
| `comments` | `false` | Label groups with `// <name>` when regrouping. |
| `sort` | `"lexical"` | `"lexical"` (byte order), `"case-insensitive"`, or `"natural"` (case-insensitive, with `v2` before `v10`). |
| `dedup` | `true` | Remove repeated includes. |
| `impl_extensions` | `.c .cc .cpp .cxx .cu .m .mm .inl .ipp .tpp .tcc …` | Extensions of implementation files, which put their main header first. |
| `main_suffixes` | `_test`, `_unittest`, `Test`, `_impl`, `Impl`, `-inl`, `_p`, … | Suffixes stripped from a file name when looking for its main header. A file of any extension with such a suffix has a main header. Only the first matching suffix is stripped. |
| `groups` | from preset | Include groups (see below). When set, it replaces the preset's groups entirely. |
| `extensions` | `.c .cc .cpp .cxx .h .hh .hpp .hxx .inl .ipp .tpp .cu .m .mm …` | Extensions processed when walking directories. |
| `exclude` | `.*`, `build`, `cmake-build-*`, … | Globs to skip. A pattern without `/` matches a base name; a pattern with `/` matches the path relative to the walked directory. |

### Groups

```json
{
  "groups": [
    { "name": "Main header", "builtin": "main" },
    { "name": "C++ standard library", "builtin": "cpp" },
    { "name": "Our libraries", "match": "^<(core|net|ui)/" },
    { "name": "Logging", "headers": ["<spdlog/spdlog.h>", "<fmt/format.h>"] },
    { "name": "Third-party", "match": "^<", "priority": 10 },
    { "name": "Project", "match": "^\"" }
  ]
}
```

An include matches a group if any of these matches it:

- `builtin`: `main` (the main header), `c` (C standard library), `cpp` (C++
  standard library up to C++26, including `<cstdio>` and friends), or `posix`.
- `match`: a regular expression ([RE2 syntax](https://github.com/google/re2/wiki/Syntax))
  applied to the spelling with delimiters, e.g. `<vector>` or `"app/app.h"`.
- `headers`: exact spellings with delimiters.

Groups are tried **in list order** and the first match wins, so put catch-all
patterns such as `^<` last. Groups are **emitted by `priority`**, which
defaults to the group's index. Give a catch-all group a smaller priority to
emit it earlier while still matching it last. Groups that share a priority
form one block, labelled with the name of the first of them. Includes that
match no group go to an implicit `Other` group at the end.

## Integrations

### pre-commit

```yaml
repos:
  - repo: https://github.com/pavel-cpp/cisort
    rev: v1.0.0
    hooks:
      - id: cisort        # sorts files; use cisort-check to only report
```

### GitHub Actions

```yaml
- uses: actions/setup-go@v5
  with:
    go-version: stable
- run: go install github.com/pavel-cpp/cisort/cmd/cisort@latest
- run: cisort -check -diff .
```

### Editors

Any editor that can pipe a buffer through a command works. Pass the file name
so that the main header and `.cisort.json` are found:

- **CLion / other JetBrains IDEs**: *Settings → Tools → External Tools*, add
  Program `cisort`, Arguments `$FilePath$`. Or use *File Watchers* to run it
  on save.
- **Vim**: `:%!cisort -stdin-filename % -`

## Development

```shell
go test ./...                              # run the tests
go test ./internal/sorter -run Golden -update  # rewrite the preset golden files
```

```
cmd/cisort           entry point (package main)
internal/cli         flags, file processing, output
internal/sorter      parsing and sorting of include runs
internal/config      settings, built-in presets (internal/config/presets/*.json), .cisort.json lookup
internal/headers     C, C++ and POSIX standard header lists
internal/finder      directory walking and filtering
internal/diff        unified diff output
```

## Authors

[Anton Zemtsov](https://github.com/antonata-c/),
[Pavel Remdenok](https://github.com/pavel-cpp/)
