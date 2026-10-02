// Command cisort sorts #include directives in C and C++ source files.
//
// Install it with:
//
//	go install github.com/pavel-cpp/cisort/cmd/cisort@latest
//
// Run "cisort -h" for usage.
package main

import (
	"os"

	"github.com/pavel-cpp/cisort/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
