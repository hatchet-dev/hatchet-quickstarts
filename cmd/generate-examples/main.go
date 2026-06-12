// Command generate-examples writes the cloneable examples under examples/ from
// this repo's templates/ tree.
//
// It mirrors the current Hatchet CLI quickstart templater so generated examples
// match `hatchet quickstart` output.
//
// Usage:
//
//	go run ./cmd/generate-examples
//	go run ./cmd/generate-examples --check
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	quickstarts "github.com/hatchet-dev/hatchet-quickstarts"
)

// data matches templater.Data in the main hatchet repo.
type data struct {
	Name           string
	PackageManager string
}

type variant struct {
	name           string // output directory name, e.g. "simple-go"
	language       string
	packageManager string
}

// poetry and pnpm are the generated variants because their templates carry lockfiles.
var variants = []variant{
	{name: "simple-go", language: "go", packageManager: "go"},
	{name: "simple-python", language: "python", packageManager: "poetry"},
	{name: "simple-typescript", language: "typescript", packageManager: "pnpm"},
}

const examplesDir = "examples"

func main() {
	check := flag.Bool("check", false, "verify generated examples are up to date instead of writing them")
	flag.Parse()

	if err := run(*check); err != nil {
		fmt.Fprintln(os.Stderr, "generate-examples:", err)
		os.Exit(1)
	}
}

func run(check bool) error {
	fsys := quickstarts.TemplatesFS()

	if check {
		return runCheck(fsys)
	}
	return runGenerate(fsys)
}

// runGenerate rewrites examples/ in place, clearing stale output first.
func runGenerate(fsys fs.FS) error {
	for _, v := range variants {
		dst := filepath.Join(examplesDir, v.name)
		if err := generateVariant(fsys, v, dst); err != nil {
			return fmt.Errorf("generating %s: %w", v.name, err)
		}
	}
	return nil
}

// runCheck is the drift check: it regenerates into a temp directory and reports
// any difference from the committed examples, leaving the working tree untouched.
func runCheck(fsys fs.FS) error {
	tmpRoot, err := os.MkdirTemp("", "generate-examples-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpRoot)

	var diffs []string
	for _, v := range variants {
		want := filepath.Join(tmpRoot, v.name)
		if err := generateVariant(fsys, v, want); err != nil {
			return fmt.Errorf("generating %s: %w", v.name, err)
		}
		got := filepath.Join(examplesDir, v.name)
		diffs = append(diffs, compareTrees(want, got)...)
	}

	if len(diffs) > 0 {
		sort.Strings(diffs)
		return fmt.Errorf("examples are out of date:\n  %s\nrun `go run ./cmd/generate-examples` to regenerate",
			strings.Join(diffs, "\n  "))
	}
	return nil
}

func generateVariant(fsys fs.FS, v variant, dst string) error {
	if err := os.RemoveAll(dst); err != nil {
		return err
	}
	d := data{Name: v.name, PackageManager: v.packageManager}
	return processMultiSource(fsys, v.language, v.packageManager, dst, d)
}

// processMultiSource mirrors templater.ProcessMultiSource. Go reads its language
// directory directly. Python and TypeScript overlay the package-manager
// directory on shared.
func processMultiSource(fsys fs.FS, language, packageManager, dstDir string, d data) error {
	if language == "go" {
		return process(fsys, "templates/go", dstDir, d)
	}

	sharedDir := path.Join("templates", language, "shared")
	pkgMgrDir := path.Join("templates", language, packageManager)

	if err := process(fsys, sharedDir, dstDir, d); err != nil {
		return err
	}
	return process(fsys, pkgMgrDir, dstDir, d)
}

// process mirrors templater.Process: it renders each file under srcDir as a
// text/template, skips POST_QUICKSTART.md, and strips the .embed suffix.
func process(fsys fs.FS, srcDir, dstDir string, d data) error {
	subFS, err := fs.Sub(fsys, srcDir)
	if err != nil {
		return err
	}

	return fs.WalkDir(subFS, ".", func(srcPath string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !entry.IsDir() && path.Base(srcPath) == "POST_QUICKSTART.md" {
			return nil
		}

		dstPath := filepath.Join(dstDir, srcPath)
		dstPath = strings.TrimSuffix(dstPath, ".embed")

		if entry.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		content, err := fs.ReadFile(subFS, srcPath)
		if err != nil {
			return err
		}

		tmpl, err := template.New(srcPath).Parse(string(content))
		if err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}

		outFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}

		if err := tmpl.Execute(outFile, d); err != nil {
			outFile.Close()
			return err
		}
		return outFile.Close()
	})
}

// compareTrees lists differences between the freshly generated tree (want) and
// the committed tree (got). Empty means identical.
func compareTrees(want, got string) []string {
	wantFiles, err := snapshot(want)
	if err != nil {
		return []string{fmt.Sprintf("%s: could not read generated output: %v", want, err)}
	}
	gotFiles, err := snapshot(got)
	if err != nil {
		return []string{fmt.Sprintf("%s: could not read committed example: %v", got, err)}
	}

	var diffs []string
	for rel, wantContent := range wantFiles {
		gotContent, ok := gotFiles[rel]
		switch {
		case !ok:
			diffs = append(diffs, fmt.Sprintf("%s: missing", filepath.Join(got, rel)))
		case !bytes.Equal(wantContent, gotContent):
			diffs = append(diffs, fmt.Sprintf("%s: differs", filepath.Join(got, rel)))
		}
	}
	for rel := range gotFiles {
		if _, ok := wantFiles[rel]; !ok {
			diffs = append(diffs, fmt.Sprintf("%s: unexpected (not produced by templates)", filepath.Join(got, rel)))
		}
	}
	return diffs
}

// snapshot maps every file under root to its contents, keyed by relative path.
// A missing root returns an error.
func snapshot(root string) (map[string][]byte, error) {
	files := map[string][]byte{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[rel] = content
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}
