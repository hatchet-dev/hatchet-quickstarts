package quickstarts_test

import (
	"io/fs"
	"path"
	"slices"
	"sort"
	"strings"
	"testing"
	"text/template"

	quickstarts "github.com/hatchet-dev/hatchet-quickstarts"
)

// The tests render the embedded templates with the same composition the
// Hatchet CLI templater uses (shared overlaid by the package-manager
// directory, Go read directly), so they prove the layout contract the CLI
// discovers at runtime.

type combo struct {
	useCase        string // empty for the default simple templates
	language       string
	packageManager string
}

func (c combo) String() string {
	useCase := c.useCase
	if useCase == "" {
		useCase = "simple"
	}
	return useCase + "/" + c.language + "/" + c.packageManager
}

var packageManagersByLanguage = map[string][]string{
	"python":     {"poetry", "uv", "pip"},
	"typescript": {"npm", "pnpm", "yarn", "bun"},
	"go":         {"go"},
}

func allCombos() []combo {
	var combos []combo
	for _, useCase := range []string{"", "scheduled"} {
		for _, language := range []string{"python", "typescript", "go"} {
			for _, pm := range packageManagersByLanguage[language] {
				combos = append(combos, combo{useCase: useCase, language: language, packageManager: pm})
			}
		}
	}
	return combos
}

func templateRoot(useCase string) string {
	if useCase == "" {
		return "templates"
	}
	return path.Join("templates", "use-cases", useCase)
}

// render composes a combo the way the CLI does and returns rendered file
// contents keyed by output path (with the .embed suffix stripped).
func render(t *testing.T, c combo) map[string]string {
	t.Helper()
	fsys := quickstarts.TemplatesFS()
	root := templateRoot(c.useCase)

	dirs := []string{path.Join(root, "go")}
	if c.language != "go" {
		dirs = []string{
			path.Join(root, c.language, "shared"),
			path.Join(root, c.language, c.packageManager),
		}
	}

	data := struct {
		Name           string
		PackageManager string
	}{Name: "render-test", PackageManager: c.packageManager}

	out := map[string]string{}
	for _, dir := range dirs {
		err := fs.WalkDir(fsys, dir, func(srcPath string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || path.Base(srcPath) == "POST_QUICKSTART.md" {
				return nil
			}
			content, err := fs.ReadFile(fsys, srcPath)
			if err != nil {
				return err
			}
			tmpl, err := template.New(srcPath).Parse(string(content))
			if err != nil {
				return err
			}
			var rendered strings.Builder
			if err := tmpl.Execute(&rendered, data); err != nil {
				return err
			}
			rel := strings.TrimSuffix(strings.TrimPrefix(srcPath, dir+"/"), ".embed")
			out[rel] = rendered.String()
			return nil
		})
		if err != nil {
			t.Fatalf("%s: rendering %s: %v", c, dir, err)
		}
	}
	return out
}

// TestScheduledMatchesSimplePackageManagerMatrix reads the package-manager
// directory names from the embedded tree, so a package manager added under
// the simple templates without a scheduled counterpart fails here even
// before the hardcoded combination lists in these tests are updated.
func TestScheduledMatchesSimplePackageManagerMatrix(t *testing.T) {
	fsys := quickstarts.TemplatesFS()

	packageManagerDirs := func(dir string) []string {
		entries, err := fs.ReadDir(fsys, dir)
		if err != nil {
			t.Fatalf("reading %s: %v", dir, err)
		}
		var names []string
		for _, entry := range entries {
			if entry.IsDir() && entry.Name() != "shared" {
				names = append(names, entry.Name())
			}
		}
		sort.Strings(names)
		return names
	}

	for _, language := range []string{"python", "typescript"} {
		simple := packageManagerDirs(path.Join("templates", language))
		scheduled := packageManagerDirs(path.Join("templates", "use-cases", "scheduled", language))
		if !slices.Equal(simple, scheduled) {
			t.Errorf("%s package managers differ:\n  simple:    %v\n  scheduled: %v",
				language, simple, scheduled)
		}
	}
}

func TestAllCombinationsRender(t *testing.T) {
	expectedSources := map[string]map[string][]string{
		"": {
			"go":         {"cmd/worker/main.go", "cmd/run/main.go", "workflows/first_workflow.go", "hatchet.yaml", "go.mod"},
			"python":     {"src/worker.py", "src/run.py", "src/workflows/first_workflow.py", "hatchet.yaml"},
			"typescript": {"src/worker.ts", "src/run.ts", "src/workflows/first-workflow.ts", "tsconfig.json", "hatchet.yaml", "package.json"},
		},
		"scheduled": {
			"go":         {"cmd/worker/main.go", "cmd/run/main.go", "workflows/scheduled_workflow.go", "hatchet.yaml", "go.mod"},
			"python":     {"src/worker.py", "src/run.py", "src/workflows/scheduled_workflow.py", "hatchet.yaml"},
			"typescript": {"src/worker.ts", "src/run.ts", "src/workflows/scheduled-workflow.ts", "tsconfig.json", "hatchet.yaml", "package.json"},
		},
	}
	manifestByPackageManager := map[string]string{
		"poetry": "pyproject.toml",
		"uv":     "pyproject.toml",
		"pip":    "requirements.txt",
		"npm":    "package.json",
		"pnpm":   "package.json",
		"yarn":   "package.json",
		"bun":    "package.json",
		"go":     "go.mod",
	}
	sdkDependencyByLanguage := map[string]string{
		"python":     "hatchet-sdk",
		"typescript": "@hatchet-dev/typescript-sdk",
		"go":         "github.com/hatchet-dev/hatchet",
	}

	for _, c := range allCombos() {
		t.Run(c.String(), func(t *testing.T) {
			files := render(t, c)

			for _, want := range expectedSources[c.useCase][c.language] {
				if _, ok := files[want]; !ok {
					t.Errorf("missing expected file %s", want)
				}
			}

			manifest := files[manifestByPackageManager[c.packageManager]]
			if !strings.Contains(manifest, sdkDependencyByLanguage[c.language]) {
				t.Errorf("manifest does not declare the SDK dependency %s", sdkDependencyByLanguage[c.language])
			}
			// Every TypeScript variant except bun runs its scripts through
			// npx ts-node, and npx falls back to fetching ts-node at run
			// time when the project does not declare it.
			if c.language == "typescript" && c.packageManager != "bun" {
				if !strings.Contains(manifest, "ts-node") {
					t.Error("package.json does not declare ts-node")
				}
			}

			if readme, ok := files["README.md"]; !ok {
				t.Error("missing README.md")
			} else {
				if !strings.Contains(readme, "hatchet quickstart") {
					t.Error("README does not use the hatchet quickstart command")
				}
				if strings.Contains(readme, "git clone") {
					t.Error("README still contains retired clone instructions")
				}
			}

			if c.useCase == "scheduled" {
				workflow := files[expectedSources["scheduled"][c.language][2]]
				if !strings.Contains(workflow, "*/5 * * * *") {
					t.Error("scheduled workflow does not register the cron schedule")
				}
				if !strings.Contains(files["hatchet.yaml"], "manual-run") {
					t.Error("hatchet.yaml does not define the manual-run trigger")
				}
				for name := range files {
					if strings.Contains(name, "first_workflow") || strings.Contains(name, "first-workflow") {
						t.Errorf("simple-only file leaked into scheduled output: %s", name)
					}
				}
			}
		})
	}
}

func TestPostQuickstartGuidance(t *testing.T) {
	fsys := quickstarts.TemplatesFS()

	for _, c := range allCombos() {
		t.Run(c.String(), func(t *testing.T) {
			root := templateRoot(c.useCase)
			postPath := path.Join(root, "go", "POST_QUICKSTART.md")
			if c.language != "go" {
				postPath = path.Join(root, c.language, c.packageManager, "POST_QUICKSTART.md")
			}

			content, err := fs.ReadFile(fsys, postPath)
			if err != nil {
				t.Fatalf("reading %s: %v", postPath, err)
			}

			trigger := "hatchet trigger simple"
			if c.useCase == "scheduled" {
				trigger = "hatchet trigger manual-run"
			}
			if !strings.Contains(string(content), trigger) {
				t.Errorf("%s does not contain %q", postPath, trigger)
			}
		})
	}
}
