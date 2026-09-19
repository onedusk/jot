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

func TestE2EDuplicateOutputPathsFail(t *testing.T) {
	// This is the input layout `jot init` generates.
	dir := t.TempDir()
	writeFixture(t, dir, map[string]string{
		"jot.yml":        "input:\n  paths: [\"docs\", \"README.md\"]\noutput:\n  path: dist\n",
		"README.md":      "# Root\n",
		"docs/README.md": "# Docs\n",
	})
	enterFixture(t, dir)

	err := runBuild(newTestBuildCmd(), nil)
	if err == nil {
		t.Fatal("expected an error for two documents mapping to README.html")
	}
	msg := err.Error()
	if !strings.Contains(msg, filepath.Join("docs", "README.md")) {
		t.Errorf("error should name docs/README.md: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(dir, "dist", "README.html")); statErr == nil {
		t.Error("build should fail before writing any pages")
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
