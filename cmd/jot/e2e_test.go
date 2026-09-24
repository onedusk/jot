package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// These tests run a full build from a temporary directory outside the
// repository, the way an installed binary is used, and load jot.yml through
// the same discovery path as the CLI.

// writeFixture creates files under root from a map of relative path to content.
func writeFixture(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("mkdir %s: %v", rel, err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
}

// enterFixture changes into dir for the rest of the test and loads its jot.yml
// (if any) using the CLI's config discovery.
func enterFixture(t *testing.T, dir string) {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(prev)
		viper.Reset()
	})
	viper.Reset()
	cfgFile = ""
	initConfig()
}

// newTestBuildCmd returns a command carrying the build flags with their defaults.
func newTestBuildCmd() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().StringP("output", "o", "", "output directory")
	cmd.Flags().BoolP("clean", "c", false, "clean output directory")
	cmd.Flags().Bool("skip-llms-txt", false, "skip llms.txt generation")
	return cmd
}

// listFiles returns every file under root as sorted slash-separated relative paths.
func listFiles(t *testing.T, root string) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			rel, _ := filepath.Rel(root, path)
			files = append(files, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	sort.Strings(files)
	return files
}

func TestE2EBuildOutOfTree(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, map[string]string{
		"jot.yml":         "input:\n  paths: [\"docs\"]\noutput:\n  path: dist\n",
		"docs/index.md":   "# Home\n\nWelcome.\n",
		"docs/guide.md":   "# Guide\n\nSteps.\n",
		"docs/sub/ref.md": "# Reference\n\nDetails.\n",
	})
	enterFixture(t, dir)

	if err := runBuild(newTestBuildCmd(), nil); err != nil {
		t.Fatalf("build failed: %v", err)
	}

	for _, page := range []string{"index.html", "guide.html", "sub/ref.html"} {
		if _, err := os.Stat(filepath.Join(dir, "dist", page)); err != nil {
			t.Errorf("expected %s: %v", page, err)
		}
	}
}

func TestE2EBuildWritesAssets(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, map[string]string{
		"jot.yml":       "input:\n  paths: [\"docs\"]\noutput:\n  path: dist\n",
		"docs/index.md": "# Home\n",
	})
	enterFixture(t, dir)

	if err := runBuild(newTestBuildCmd(), nil); err != nil {
		t.Fatalf("build failed: %v", err)
	}

	for _, asset := range []string{"style.css", "search.js", "highlight.js", "syntax-highlighting.css"} {
		info, err := os.Stat(filepath.Join(dir, "dist", "assets", asset))
		if err != nil {
			t.Errorf("missing asset %s: %v", asset, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("asset %s is empty", asset)
		}
	}
}

func TestE2ERebuildDoesNotIngestOutput(t *testing.T) {
	// No jot.yml: defaults are input "." and output "./dist", without clean.
	dir := t.TempDir()
	writeFixture(t, dir, map[string]string{
		"README.md":       "# Home\n",
		"guides/intro.md": "# Intro\n",
	})
	enterFixture(t, dir)

	if err := runBuild(newTestBuildCmd(), nil); err != nil {
		t.Fatalf("first build failed: %v", err)
	}
	first := listFiles(t, filepath.Join(dir, "dist"))

	if err := runBuild(newTestBuildCmd(), nil); err != nil {
		t.Fatalf("second build failed: %v", err)
	}
	second := listFiles(t, filepath.Join(dir, "dist"))

	if strings.Join(first, "\n") != strings.Join(second, "\n") {
		t.Errorf("second build changed the output file set:\nfirst:  %v\nsecond: %v", first, second)
	}
}

func TestE2EInPlaceBuildKeepsSources(t *testing.T) {
	// Output directory equal to the input root: sources must still be scanned
	// and must not be overwritten by the markdown copies.
	source := "---\ntitle: Guide\n---\n\n# Guide\n\nSteps.\n"
	dir := t.TempDir()
	writeFixture(t, dir, map[string]string{
		"jot.yml":  "input:\n  paths: [\".\"]\noutput:\n  path: .\n",
		"index.md": "# Home\n",
		"guide.md": source,
	})
	enterFixture(t, dir)

	for i := 1; i <= 2; i++ {
		if err := runBuild(newTestBuildCmd(), nil); err != nil {
			t.Fatalf("build %d failed: %v", i, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "guide.html")); err != nil {
		t.Errorf("expected guide.html: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "guide.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != source {
		t.Errorf("source guide.md was modified:\n%s", got)
	}
}

func TestE2ECleanRefusesToDeleteSources(t *testing.T) {
	source := "# Guide\n"
	dir := t.TempDir()
	writeFixture(t, dir, map[string]string{
		"jot.yml":       "input:\n  paths: [\"docs\"]\noutput:\n  path: docs\n  clean: true\n",
		"docs/guide.md": source,
	})
	enterFixture(t, dir)

	err := runBuild(newTestBuildCmd(), nil)
	if err == nil || !strings.Contains(err.Error(), "refusing to clean") {
		t.Fatalf("expected a refusal to clean the source directory, got %v", err)
	}
	got, readErr := os.ReadFile(filepath.Join(dir, "docs", "guide.md"))
	if readErr != nil || string(got) != source {
		t.Fatalf("source file was removed or changed: %v", readErr)
	}
}

func TestE2EDuplicateOutputPathsFail(t *testing.T) {
	// This is the input layout and clean setting `jot init` generates.
	dir := t.TempDir()
	writeFixture(t, dir, map[string]string{
		"jot.yml":   "input:\n  paths: [\"docs\", \"README.md\"]\noutput:\n  path: dist\n  clean: true\n",
		"README.md": "# Root\n",
	})
	enterFixture(t, dir)

	if err := runBuild(newTestBuildCmd(), nil); err != nil {
		t.Fatalf("first build failed: %v", err)
	}

	writeFixture(t, dir, map[string]string{"docs/README.md": "# Docs\n"})
	err := runBuild(newTestBuildCmd(), nil)
	if err == nil {
		t.Fatal("expected an error for two documents mapping to README.html")
	}
	msg := err.Error()
	if !strings.Contains(msg, filepath.Join("docs", "README.md")) {
		t.Errorf("error should name docs/README.md: %v", err)
	}
	page, readErr := os.ReadFile(filepath.Join(dir, "dist", "README.html"))
	if readErr != nil || !strings.Contains(string(page), "Root") {
		t.Errorf("failed build should leave the previous output in place: %v", readErr)
	}
}

func TestE2EOverlappingInputPathsBuild(t *testing.T) {
	// One file reached through two input paths is not a collision.
	dir := t.TempDir()
	writeFixture(t, dir, map[string]string{
		"jot.yml":   "input:\n  paths: [\".\", \"README.md\", \"./docs/\", \"docs\"]\noutput:\n  path: dist\n",
		"README.md": "# Root\n",
		"docs/a.md": "# A\n",
	})
	enterFixture(t, dir)

	if err := runBuild(newTestBuildCmd(), nil); err != nil {
		t.Fatalf("build failed: %v", err)
	}
}

// captureStdout runs fn with os.Stdout redirected and returns what it wrote.
func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	orig := os.Stdout
	os.Stdout = w
	done := make(chan []byte)
	go func() {
		out, _ := io.ReadAll(r)
		done <- out
	}()
	runErr := fn()
	w.Close()
	os.Stdout = orig
	return string(<-done), runErr
}

func TestE2EIndexNameCasing(t *testing.T) {
	tests := []struct {
		name      string
		files     map[string]string
		wantIndex string
	}{
		{
			name:      "capitalized Index.md wins over README.md",
			files:     map[string]string{"docs/Index.md": "# My Index\n", "docs/README.md": "# My Readme\n"},
			wantIndex: "My Index",
		},
		{
			name:      "lowercase readme.md becomes the index",
			files:     map[string]string{"docs/readme.md": "# My Readme\n", "docs/guide.md": "# Guide\n"},
			wantIndex: "My Readme",
		},
		{
			name:      "README.md in a subdirectory is not the index",
			files:     map[string]string{"docs/sub/README.md": "# Sub Readme\n", "docs/guide.md": "# Guide\n"},
			wantIndex: "Documentation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.files["jot.yml"] = "input:\n  paths: [\"docs\"]\noutput:\n  path: dist\n"
			writeFixture(t, dir, tt.files)
			enterFixture(t, dir)

			if err := runBuild(newTestBuildCmd(), nil); err != nil {
				t.Fatalf("build failed: %v", err)
			}
			index, err := os.ReadFile(filepath.Join(dir, "dist", "index.html"))
			if err != nil {
				t.Fatalf("expected dist/index.html: %v", err)
			}
			// Every page lists every title in its sidebar, so check the page title
			if want := "<title>" + tt.wantIndex + " |"; !strings.Contains(string(index), want) {
				t.Errorf("index.html should be the %q page", tt.wantIndex)
			}
		})
	}
}

func TestE2EExportToStdoutIsPureJSON(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, map[string]string{
		"jot.yml":       "input:\n  paths: [\"docs\"]\n",
		"docs/index.md": "# Home\n\nWelcome.\n",
	})
	enterFixture(t, dir)

	out, err := captureStdout(t, func() error { return runExport(exportCmd, nil) })
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(out), &decoded); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout:\n%s", err, out)
	}
}

func TestE2EErrorOutput(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		want      []string
		wantUsage bool
	}{
		{"runtime error", []string{"build"}, []string{"Error: no markdown files found"}, false},
		{"unknown flag", []string{"build", "--bogus"}, []string{"Error: unknown flag: --bogus"}, true},
		{"unknown command", []string{"biuld"}, []string{`Error: unknown command "biuld"`, "Run 'jot --help' for usage."}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enterFixture(t, t.TempDir())

			var cobraOut strings.Builder
			rootCmd.SetOut(&cobraOut)
			rootCmd.SetErr(&cobraOut)
			rootCmd.SetArgs(tt.args)
			t.Cleanup(func() {
				rootCmd.SetOut(nil)
				rootCmd.SetErr(nil)
				rootCmd.SetArgs(nil)
				// PersistentPreRun sets this on the command that ran; reset for the next case
				for _, cmd := range rootCmd.Commands() {
					cmd.SilenceUsage = false
				}
			})

			if err := rootCmd.Execute(); err == nil {
				t.Fatal("expected an error")
			}
			got := cobraOut.String()
			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Errorf("output missing %q:\n%s", want, got)
				}
			}
			if n := strings.Count(got, "Error:"); n != 1 {
				t.Errorf("error printed %d times, want once:\n%s", n, got)
			}
			if hasUsage := strings.Contains(got, "Usage:"); hasUsage != tt.wantUsage {
				t.Errorf("usage shown = %v, want %v:\n%s", hasUsage, tt.wantUsage, got)
			}
		})
	}
}

func TestE2EVerboseConfigMessageNotOnStdout(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, map[string]string{"jot.yml": "input:\n  paths: [\"docs\"]\n"})
	verbose = true
	t.Cleanup(func() { verbose = false })

	out, _ := captureStdout(t, func() error {
		enterFixture(t, dir)
		return nil
	})
	if out != "" {
		t.Errorf("config discovery wrote to stdout with --verbose: %q", out)
	}
}

func TestE2EExportToStdoutEndsWithOneNewline(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, map[string]string{
		"jot.yml":       "input:\n  paths: [\"docs\"]\n",
		"docs/index.md": "# Home\n\nWelcome.\n\n\n",
	})
	enterFixture(t, dir)
	if err := exportCmd.Flags().Set("format", "llms-full"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { exportCmd.Flags().Set("format", "json") })

	out, err := captureStdout(t, func() error { return runExport(exportCmd, nil) })
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}
	if !strings.HasSuffix(out, "\n") || strings.HasSuffix(out, "\n\n") {
		t.Errorf("stdout should end with exactly one newline, got %q", out[max(0, len(out)-10):])
	}
}

func TestE2EInitThenBuild(t *testing.T) {
	// The project `jot init` creates must build without errors.
	dir := t.TempDir()
	enterFixture(t, dir)

	if err := runInit(initCmd, nil); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	viper.Reset()
	initConfig()

	if err := runBuild(newTestBuildCmd(), nil); err != nil {
		t.Fatalf("build of a freshly initialized project failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "dist", "README.html")); err != nil {
		t.Errorf("expected dist/README.html: %v", err)
	}
}

func TestE2EIndexFromRootReadme(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, map[string]string{
		"jot.yml":        "input:\n  paths: [\"docs\"]\noutput:\n  path: dist\n",
		"docs/README.md": "# Home\n",
		"docs/guide.md":  "# Guide\n",
	})
	enterFixture(t, dir)

	if err := runBuild(newTestBuildCmd(), nil); err != nil {
		t.Fatalf("build failed: %v", err)
	}
	index, err := os.ReadFile(filepath.Join(dir, "dist", "index.html"))
	if err != nil {
		t.Fatalf("expected dist/index.html: %v", err)
	}
	readme, err := os.ReadFile(filepath.Join(dir, "dist", "README.html"))
	if err != nil {
		t.Fatalf("expected dist/README.html to remain: %v", err)
	}
	if string(index) != string(readme) {
		t.Error("index.html should be the rendered README page")
	}
}

func TestE2EProjectStructureKeepsInputDirectories(t *testing.T) {
	// The input paths `jot init` generates, with output.structure: project
	dir := t.TempDir()
	writeFixture(t, dir, map[string]string{
		"jot.yml":         "input:\n  paths: [\"docs\", \"README.md\"]\noutput:\n  path: dist\n  structure: project\n",
		"README.md":       "# Root\n",
		"docs/README.md":  "# Docs\n",
		"docs/sub/ref.md": "# Ref\n",
	})
	enterFixture(t, dir)

	if err := runBuild(newTestBuildCmd(), nil); err != nil {
		t.Fatalf("build failed: %v", err)
	}
	for _, want := range []string{"README.html", "docs/README.html", "docs/README.md", "docs/sub/ref.html", "docs/sub/ref.md"} {
		if _, err := os.Stat(filepath.Join(dir, "dist", want)); err != nil {
			t.Errorf("expected dist/%s: %v", want, err)
		}
	}
	checks := map[string]string{
		"index.html":               "<title>Root |",
		"docs/sub/ref.html":        `href="../../assets/style.css"`,
		"assets/search-index.json": `"docs/sub/ref.html"`,
		"llms.txt":                 "(docs/README.md)",
	}
	for file, want := range checks {
		content, err := os.ReadFile(filepath.Join(dir, "dist", file))
		if err != nil || !strings.Contains(string(content), want) {
			t.Errorf("dist/%s should contain %q (read error: %v)", file, want, err)
		}
	}
}

func TestE2EProjectStructureGeneratedIndex(t *testing.T) {
	// Only a document at the project root can become index.html
	dir := t.TempDir()
	writeFixture(t, dir, map[string]string{
		"jot.yml":       "input:\n  paths: [\"docs\"]\noutput:\n  path: dist\n  structure: project\n",
		"docs/index.md": "# Docs Home\n",
		"docs/guide.md": "# Guide\n",
	})
	enterFixture(t, dir)

	if err := runBuild(newTestBuildCmd(), nil); err != nil {
		t.Fatalf("build failed: %v", err)
	}
	index, err := os.ReadFile(filepath.Join(dir, "dist", "index.html"))
	if err != nil || !strings.Contains(string(index), "<title>Documentation |") {
		t.Errorf("index.html should be the generated contents page (read error: %v)", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "dist", "docs", "index.html")); err != nil {
		t.Errorf("expected dist/docs/index.html: %v", err)
	}
}

func TestE2EProjectStructureOverlappingInputs(t *testing.T) {
	// A file reached through two input paths gets the same project-relative
	// path from both, whichever is listed first.
	dir := t.TempDir()
	writeFixture(t, dir, map[string]string{
		"jot.yml":   "input:\n  paths: [\"docs\", \".\"]\noutput:\n  path: dist\n  structure: project\n",
		"README.md": "# Root\n",
		"docs/a.md": "# A\n",
	})
	enterFixture(t, dir)

	if err := runBuild(newTestBuildCmd(), nil); err != nil {
		t.Fatalf("build failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "dist", "docs", "a.html")); err != nil {
		t.Errorf("expected dist/docs/a.html: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "dist", "a.html")); err == nil {
		t.Error("dist/a.html should not exist")
	}
}

func TestE2EProjectStructureAbsoluteInputThroughSymlink(t *testing.T) {
	// The project root is matched by file identity, so an absolute input path
	// spelled differently from the working directory is still inside it.
	base := t.TempDir()
	if err := os.Symlink(filepath.Join(base, "proj"), filepath.Join(base, "alias")); err != nil {
		t.Skipf("symlinks not supported: %v", err)
	}
	input := filepath.ToSlash(filepath.Join(base, "alias", "docs"))
	writeFixture(t, base, map[string]string{
		"proj/jot.yml":       "input:\n  paths: [\"" + input + "\"]\noutput:\n  path: dist\n  structure: project\n",
		"proj/docs/guide.md": "# Guide\n",
	})
	enterFixture(t, filepath.Join(base, "proj"))

	if err := runBuild(newTestBuildCmd(), nil); err != nil {
		t.Fatalf("build failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(base, "proj", "dist", "docs", "guide.html")); err != nil {
		t.Errorf("expected dist/docs/guide.html: %v", err)
	}
}

func TestE2EProjectStructureInputOutsideRoot(t *testing.T) {
	base := t.TempDir()
	writeFixture(t, base, map[string]string{
		"site/jot.yml": "input:\n  paths: [\"../shared\"]\noutput:\n  path: dist\n  structure: project\n",
		"shared/x.md":  "# X\n",
	})
	enterFixture(t, filepath.Join(base, "site"))

	err := runBuild(newTestBuildCmd(), nil)
	if err == nil {
		t.Fatal("expected an error for an input path outside the project root")
	}
	for _, want := range []string{"input path ../shared is outside the project root", "set output.structure: input"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error should contain %q: %v", want, err)
		}
	}
	if _, statErr := os.Stat(filepath.Join(base, "site", "dist")); statErr == nil {
		t.Error("a rejected build should not create the output directory")
	}
}

func TestE2EUnsupportedOutputStructure(t *testing.T) {
	// The value is checked before scanning, so it is reported even when no
	// input path exists. A YAML list must not fall back to the default.
	tests := []struct{ value, want string }{
		{"flat", `unsupported output.structure: "flat" (supported: input, project)`},
		{"[project]", `unsupported output.structure: "[project]" (supported: input, project)`},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			dir := t.TempDir()
			writeFixture(t, dir, map[string]string{
				"jot.yml": "input:\n  paths: [\"missing\"]\noutput:\n  path: dist\n  structure: " + tt.value + "\n",
			})
			enterFixture(t, dir)

			err := runBuild(newTestBuildCmd(), nil)
			if err == nil || err.Error() != tt.want {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestE2EProjectStructureInEveryCommand(t *testing.T) {
	// export, toc, and debug name documents the same way build does.
	files := map[string]string{
		"jot.yml":        "input:\n  paths: [\"docs\", \"README.md\"]\noutput:\n  path: dist\n  structure: project\n",
		"README.md":      "# Root\n",
		"docs/README.md": "# Docs\n",
		"docs/guide.md":  "# Guide\n",
	}
	tests := []struct {
		name  string
		run   func() error
		check func(t *testing.T, dir, stdout string)
	}{
		{"export", func() error { return runExport(exportCmd, nil) }, func(t *testing.T, dir, stdout string) {
			for _, want := range []string{`"path": "README.md"`, `"path": "docs/README.md"`, `"path": "docs/guide.md"`} {
				if !strings.Contains(stdout, want) {
					t.Errorf("export should contain %s", want)
				}
			}
		}},
		{"toc", func() error { return runTOC(tocCmd, nil) }, func(t *testing.T, dir, stdout string) {
			toc, err := os.ReadFile(filepath.Join(dir, "docs", "toc.xml"))
			if err != nil || !strings.Contains(string(toc), `path="docs/guide.md"`) {
				t.Errorf("docs/toc.xml should contain path=\"docs/guide.md\" (read error: %v)", err)
			}
		}},
		{"debug", func() error { return runDebug(debugCmd, nil) }, func(t *testing.T, dir, stdout string) {
			if !strings.Contains(stdout, `Found: "docs/guide.md"`) {
				t.Errorf("debug output should list docs/guide.md:\n%s", stdout)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFixture(t, dir, files)
			enterFixture(t, dir)

			stdout, err := captureStdout(t, tt.run)
			if err != nil {
				t.Fatalf("%s failed: %v", tt.name, err)
			}
			tt.check(t, dir, stdout)
		})
	}
}

func TestE2EOutputInputDirectoryRejectedWhenCopiesMove(t *testing.T) {
	// Building into an input path is only safe when every markdown copy lands
	// on its own source. Otherwise a copy overwrites another source or is read
	// back as a new source by the next build.
	tests := []struct {
		name      string
		structure string
		files     map[string]string
	}{
		{
			name:      "project mode, output is a non-root input",
			structure: "project",
			files:     map[string]string{"README.md": "# Root\n", "docs/README.md": "# Docs\n", "docs/guide.md": "# Guide\n"},
		},
		{
			name:      "input mode, output is one of several inputs",
			structure: "input",
			files:     map[string]string{"README.md": "# Root\n", "docs/guide.md": "# Guide\n"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.files["jot.yml"] = "input:\n  paths: [\"docs\", \"README.md\"]\noutput:\n  path: dist\n  structure: " + tt.structure + "\n"
			writeFixture(t, dir, tt.files)
			enterFixture(t, dir)

			cmd := newTestBuildCmd()
			if err := cmd.Flags().Set("output", "docs"); err != nil {
				t.Fatal(err)
			}
			err := runBuild(cmd, nil)
			if err == nil || !strings.Contains(err.Error(), "is also input path docs") {
				t.Fatalf("expected the output directory to be rejected, got %v", err)
			}
			for rel, content := range tt.files {
				if rel == "jot.yml" {
					continue
				}
				got, readErr := os.ReadFile(filepath.Join(dir, rel))
				if readErr != nil || string(got) != content {
					t.Errorf("source %s was changed: %q (%v)", rel, got, readErr)
				}
			}
			if _, statErr := os.Stat(filepath.Join(dir, "docs", "docs")); statErr == nil {
				t.Error("docs/docs should not be created")
			}
			if tt.structure == "input" {
				if _, statErr := os.Stat(filepath.Join(dir, "docs", "README.md")); statErr == nil {
					t.Error("the root README should not be copied into docs/")
				}
			}
		})
	}
}

func TestE2EProjectStructureInPlaceBuild(t *testing.T) {
	// Building into the project root puts every copy on its own source.
	source := "---\ntitle: Guide\n---\n\n# Guide\n"
	dir := t.TempDir()
	writeFixture(t, dir, map[string]string{
		"jot.yml":       "input:\n  paths: [\"docs\", \"README.md\"]\noutput:\n  path: .\n  structure: project\n",
		"README.md":     "# Root\n",
		"docs/guide.md": source,
	})
	enterFixture(t, dir)

	for i := 1; i <= 2; i++ {
		if err := runBuild(newTestBuildCmd(), nil); err != nil {
			t.Fatalf("build %d failed: %v", i, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "docs", "guide.html")); err != nil {
		t.Errorf("expected docs/guide.html: %v", err)
	}
	if got, err := os.ReadFile(filepath.Join(dir, "docs", "guide.md")); err != nil || string(got) != source {
		t.Errorf("source docs/guide.md was modified: %q (%v)", got, err)
	}
}
