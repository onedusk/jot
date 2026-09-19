// Package scanner_test contains tests for the scanner package.
package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNewScanner tests the creation of a new Scanner.
func TestNewScanner(t *testing.T) {
	tests := []struct {
		name     string
		rootPath string
		ignore   []string
		wantErr  bool
	}{
		{
			name:     "valid root path",
			rootPath: ".",
			ignore:   []string{"*.tmp"},
			wantErr:  false,
		},
		{
			name:     "empty root path",
			rootPath: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewScanner(tt.rootPath, tt.ignore)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewScanner() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestScanner_Scan tests the file scanning functionality.
func TestScanner_Scan(t *testing.T) {
	// Create temporary test directory
	tmpDir := t.TempDir()

	// Create test file structure
	testFiles := map[string]string{
		"README.md":               "# Test Project",
		"docs/getting-started.md": "# Getting Started",
		"docs/api/reference.md":   "# API Reference",
		"docs/drafts/wip.md":      "# Work in Progress",
		"docs/.hidden/secret.md":  "# Secret",
		"src/main.go":             "package main",
		"test.txt":                "not markdown",
	}

	for path, content := range testFiles {
		fullPath := filepath.Join(tmpDir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name           string
		ignorePatterns []string
		wantCount      int
		wantPaths      []string
		notWantPaths   []string
	}{
		{
			name:           "scan all markdown files",
			ignorePatterns: []string{},
			wantCount:      5,
			wantPaths:      []string{"README.md", "docs/getting-started.md", "docs/api/reference.md"},
		},
		{
			name:           "ignore drafts directory",
			ignorePatterns: []string{"**/drafts/**"},
			wantCount:      4,
			notWantPaths:   []string{"docs/drafts/wip.md"},
		},
		{
			name:           "ignore hidden directories",
			ignorePatterns: []string{"**/.*/**"},
			wantCount:      4,
			notWantPaths:   []string{"docs/.hidden/secret.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner, err := NewScanner(tmpDir, tt.ignorePatterns)
			if err != nil {
				t.Fatal(err)
			}

			docs, err := scanner.Scan()
			if err != nil {
				t.Fatal(err)
			}

			if len(docs) != tt.wantCount {
				t.Errorf("Scan() returned %d documents, want %d", len(docs), tt.wantCount)
			}

			// Check wanted paths
			for _, wantPath := range tt.wantPaths {
				found := false
				for _, doc := range docs {
					if doc.RelativePath == wantPath {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Scan() missing expected path: %s", wantPath)
				}
			}

			// Check not wanted paths
			for _, notWantPath := range tt.notWantPaths {
				for _, doc := range docs {
					if doc.RelativePath == notWantPath {
						t.Errorf("Scan() included ignored path: %s", notWantPath)
					}
				}
			}
		})
	}
}

// TestScanner_Exclude verifies that excluded directories are not scanned.
func TestScanner_Exclude(t *testing.T) {
	tmpDir := t.TempDir()
	testFiles := map[string]string{
		"README.md":            "# Home",
		"guide.md":             "# Guide",
		"dist/README.md":       "# Copied home",
		"dist/nested/guide.md": "# Copied guide",
	}
	for path, content := range testFiles {
		fullPath := filepath.Join(tmpDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	scanner, err := NewScanner(tmpDir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := scanner.Exclude(filepath.Join(tmpDir, "dist")); err != nil {
		t.Fatal(err)
	}

	docs, err := scanner.Scan()
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 2 {
		var paths []string
		for _, doc := range docs {
			paths = append(paths, doc.RelativePath)
		}
		t.Errorf("Scan() returned %v, want only README.md and guide.md", paths)
	}
}

// TestScanner_ExcludeMatchesByIdentity verifies that an excluded directory is
// matched through a symlink, and that excluding the root does not empty the scan.
func TestScanner_ExcludeMatchesByIdentity(t *testing.T) {
	tmpDir := t.TempDir()
	for path, content := range map[string]string{
		"README.md":          "# Home",
		"build/site/copy.md": "# Copied home",
	} {
		fullPath := filepath.Join(tmpDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	// public -> build/site, a common deploy layout
	if err := os.Symlink(filepath.Join(tmpDir, "build", "site"), filepath.Join(tmpDir, "public")); err != nil {
		t.Skipf("symlinks not supported: %v", err)
	}

	tests := []struct {
		name    string
		exclude string
		want    []string
	}{
		{"symlinked output directory", filepath.Join(tmpDir, "public"), []string{"README.md"}},
		{"output directory is the scan root", tmpDir, []string{"README.md", "build/site/copy.md"}},
		{"output directory does not exist", filepath.Join(tmpDir, "missing"), []string{"README.md", "build/site/copy.md"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner, err := NewScanner(tmpDir, nil)
			if err != nil {
				t.Fatal(err)
			}
			if err := scanner.Exclude(tt.exclude); err != nil {
				t.Fatal(err)
			}
			docs, err := scanner.Scan()
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			for _, doc := range docs {
				got = append(got, doc.RelativePath)
			}
			if strings.Join(got, ",") != strings.Join(tt.want, ",") {
				t.Errorf("Scan() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDocument_ExtractTitle tests the title extraction logic.
func TestDocument_ExtractTitle(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		metadata map[string]interface{}
		want     string
	}{
		{
			name:    "extract from H1",
			content: "# My Document Title\n\nSome content",
			want:    "My Document Title",
		},
		{
			name:    "extract from frontmatter",
			content: "---\ntitle: Frontmatter Title\n---\n# Different Title",
			metadata: map[string]interface{}{
				"title": "Frontmatter Title",
			},
			want: "Frontmatter Title",
		},
		{
			name:    "no title found",
			content: "Just some text without headers",
			want:    "Untitled",
		},
		{
			name:    "multiple H1s",
			content: "# First Title\n\n# Second Title",
			want:    "First Title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := &Document{
				Content:  []byte(tt.content),
				Metadata: tt.metadata,
			}

			title := doc.ExtractTitle()
			if title != tt.want {
				t.Errorf("ExtractTitle() = %v, want %v", title, tt.want)
			}
		})
	}
}

// TestIgnoreFilter tests the file ignore filtering logic.
func TestIgnoreFilter(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		path     string
		want     bool
	}{
		{
			name:     "match exact filename",
			patterns: []string{"README.md"},
			path:     "README.md",
			want:     true,
		},
		{
			name:     "match wildcard",
			patterns: []string{"*.tmp"},
			path:     "test.tmp",
			want:     true,
		},
		{
			name:     "match directory",
			patterns: []string{"docs/**"},
			path:     "docs/guide.md",
			want:     true,
		},
		{
			name:     "no match",
			patterns: []string{"*.tmp", "drafts/**"},
			path:     "guide.md",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewIgnoreFilter(tt.patterns)
			if got := filter.ShouldIgnore(tt.path); got != tt.want {
				t.Errorf("ShouldIgnore() = %v, want %v", got, tt.want)
			}
		})
	}
}
