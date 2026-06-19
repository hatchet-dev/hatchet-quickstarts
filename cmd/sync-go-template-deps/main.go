// Command sync-go-template-deps keeps the embedded Go template dependency files
// in sync with their generated Go examples.
//
// A Go template cannot hold a real go.mod, since it sits under the embedded
// template tree, so it uses go.mod.embed and go.sum. This command keeps each of
// those identical to its generated example's go.mod and go.sum. It does not
// parse versions and does not run go commands.
//
// Usage:
//
//	go run ./cmd/sync-go-template-deps
//	go run ./cmd/sync-go-template-deps --check
//
// Run it from the repository root. Paths are relative to the working directory.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"
)

type syncPair struct {
	src string
	dst string
}

var pairs = []syncPair{
	{src: "examples/simple-go/go.mod", dst: "templates/go/go.mod.embed"},
	{src: "examples/simple-go/go.sum", dst: "templates/go/go.sum"},
	{src: "examples/use-cases/scheduled/go/go.mod", dst: "templates/use-cases/scheduled/go/go.mod.embed"},
	{src: "examples/use-cases/scheduled/go/go.sum", dst: "templates/use-cases/scheduled/go/go.sum"},
}

func main() {
	check := flag.Bool("check", false, "verify the template files match the generated Go example instead of writing them")
	flag.Parse()

	if err := run(*check); err != nil {
		fmt.Fprintln(os.Stderr, "sync-go-template-deps:", err)
		os.Exit(1)
	}
}

func run(check bool) error {
	if check {
		return runCheck()
	}
	return runSync()
}

// runSync writes each example file onto its template counterpart, preserving
// the destination's existing file mode.
func runSync() error {
	for _, p := range pairs {
		content, err := os.ReadFile(p.src)
		if err != nil {
			return fmt.Errorf("reading %s: %w", p.src, err)
		}

		mode := os.FileMode(0644)
		if info, err := os.Stat(p.dst); err == nil {
			mode = info.Mode().Perm()
		}

		if err := os.WriteFile(p.dst, content, mode); err != nil {
			return fmt.Errorf("writing %s: %w", p.dst, err)
		}
	}
	return nil
}

// runCheck reports any template file that does not match its example
// counterpart byte-for-byte, leaving the working tree untouched.
func runCheck() error {
	var stale []string
	for _, p := range pairs {
		srcContent, err := os.ReadFile(p.src)
		if err != nil {
			return fmt.Errorf("reading %s: %w", p.src, err)
		}
		dstContent, err := os.ReadFile(p.dst)
		if err != nil {
			return fmt.Errorf("reading %s: %w", p.dst, err)
		}
		if !bytes.Equal(srcContent, dstContent) {
			stale = append(stale, fmt.Sprintf("%s is out of sync with %s", p.dst, p.src))
		}
	}

	if len(stale) > 0 {
		return fmt.Errorf("Go template dependency files are out of sync:\n  %s\nrun `go run ./cmd/sync-go-template-deps` to update",
			strings.Join(stale, "\n  "))
	}
	return nil
}
