package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/onedusk/jot/pkg/scanner"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// TestBuildWithLLMSTxt verifies that llms.txt and llms-full.txt are generated during build
func TestBuildWithLLMSTxt(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "jot-build-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test markdown files
	docsDir := filepath.Join(tmpDir, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatalf("Failed to create docs dir: %v", err)
	}

	// Create README.md
	readmeContent := `# Test Documentation

This is a test documentation project.

## Introduction

This is the introduction section.
`
	if err := os.WriteFile(filepath.Join(docsDir, "README.md"), []byte(readmeContent), 0644); err != nil {
		t.Fatalf("Failed to write README.md: %v", err)
	}

	// Create another test file
	guideContent := `# User Guide

Learn how to use this amazing tool.

This guide will help you get started.
`
	if err := os.WriteFile(filepath.Join(docsDir, "guide.md"), []byte(guideContent), 0644); err != nil {
		t.Fatalf("Failed to write guide.md: %v", err)
	}

	// Create output directory
	outputDir := filepath.Join(tmpDir, "dist")

	// Create test config
	configContent := `version: 1.0
project:
  name: "Test Project"
  description: "Test documentation project"

input:
  paths:
    - "` + docsDir + `"

output:
  path: "` + outputDir + `"

features:
  llm_export: true
`
	configPath := filepath.Join(tmpDir, "jot.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Initialize viper with test config
	viper.Reset()
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	// Create build command and run
	cmd := &cobra.Command{}
	cmd.Flags().StringP("output", "o", "", "output directory")
	cmd.Flags().BoolP("clean", "c", false, "clean output directory")
	cmd.Flags().Bool("skip-llms-txt", false, "skip llms.txt generation")

	// Run build
	if err := runBuild(cmd, []string{}); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verify llms.txt exists
	llmsTxtPath := filepath.Join(outputDir, "llms.txt")
	if _, err := os.Stat(llmsTxtPath); os.IsNotExist(err) {
		t.Errorf("llms.txt was not created at %s", llmsTxtPath)
	}

	// Verify llms-full.txt exists
	llmsFullTxtPath := filepath.Join(outputDir, "llms-full.txt")
	if _, err := os.Stat(llmsFullTxtPath); os.IsNotExist(err) {
		t.Errorf("llms-full.txt was not created at %s", llmsFullTxtPath)
	}

	// Verify llms.txt content
	llmsTxtContent, err := os.ReadFile(llmsTxtPath)
	if err != nil {
		t.Fatalf("Failed to read llms.txt: %v", err)
	}

	llmsTxtStr := string(llmsTxtContent)

	// Check for required elements in llms.txt
	if !strings.Contains(llmsTxtStr, "# Test Project") {
		t.Errorf("llms.txt missing project name header")
	}
	if !strings.Contains(llmsTxtStr, "> Test documentation project") {
		t.Errorf("llms.txt missing project description")
	}
	if !strings.Contains(llmsTxtStr, "Test Documentation") {
		t.Errorf("llms.txt missing document title")
	}
	if !strings.Contains(llmsTxtStr, "User Guide") {
		t.Errorf("llms.txt missing guide title")
	}

	// Verify llms-full.txt content
	llmsFullTxtContent, err := os.ReadFile(llmsFullTxtPath)
	if err != nil {
		t.Fatalf("Failed to read llms-full.txt: %v", err)
	}

	llmsFullTxtStr := string(llmsFullTxtContent)

	// Check for required elements in llms-full.txt
	if !strings.Contains(llmsFullTxtStr, "# Test Project") {
		t.Errorf("llms-full.txt missing project name header")
	}
	if !strings.Contains(llmsFullTxtStr, "> Test documentation project") {
		t.Errorf("llms-full.txt missing project description")
	}
	if !strings.Contains(llmsFullTxtStr, "This is a test documentation project") {
		t.Errorf("llms-full.txt missing full content")
	}
	if !strings.Contains(llmsFullTxtStr, "---") {
		t.Errorf("llms-full.txt missing document separator")
	}

	// Verify that llms-full.txt is larger than llms.txt
	if len(llmsFullTxtContent) <= len(llmsTxtContent) {
		t.Errorf("llms-full.txt should be larger than llms.txt, got %d vs %d bytes",
			len(llmsFullTxtContent), len(llmsTxtContent))
	}
}

// TestBuildWithSkipLLMSTxt verifies that --skip-llms-txt flag prevents generation
func TestBuildWithSkipLLMSTxt(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "jot-build-skip-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test markdown files
	docsDir := filepath.Join(tmpDir, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatalf("Failed to create docs dir: %v", err)
	}

	readmeContent := `# Test

Test content.
`
	if err := os.WriteFile(filepath.Join(docsDir, "README.md"), []byte(readmeContent), 0644); err != nil {
		t.Fatalf("Failed to write README.md: %v", err)
	}

	// Create output directory
	outputDir := filepath.Join(tmpDir, "dist")

	// Create test config
	configContent := `version: 1.0
project:
  name: "Test Project"
  description: "Test project"

input:
  paths:
    - "` + docsDir + `"

output:
  path: "` + outputDir + `"

features:
  llm_export: true
`
	configPath := filepath.Join(tmpDir, "jot.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Initialize viper with test config
	viper.Reset()
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	// Create build command with --skip-llms-txt flag
	cmd := &cobra.Command{}
	cmd.Flags().StringP("output", "o", "", "output directory")
	cmd.Flags().BoolP("clean", "c", false, "clean output directory")
	cmd.Flags().Bool("skip-llms-txt", false, "skip llms.txt generation")
	cmd.Flags().Set("skip-llms-txt", "true")

	// Run build
	if err := runBuild(cmd, []string{}); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verify llms.txt does NOT exist
	llmsTxtPath := filepath.Join(outputDir, "llms.txt")
	if _, err := os.Stat(llmsTxtPath); !os.IsNotExist(err) {
		t.Errorf("llms.txt should not be created when --skip-llms-txt is set")
	}

	// Verify llms-full.txt does NOT exist
	llmsFullTxtPath := filepath.Join(outputDir, "llms-full.txt")
	if _, err := os.Stat(llmsFullTxtPath); !os.IsNotExist(err) {
		t.Errorf("llms-full.txt should not be created when --skip-llms-txt is set")
	}
}

// TestDedupeDocuments verifies duplicate source and output path handling.
func TestDedupeDocuments(t *testing.T) {
	doc := func(path, rel string) scanner.Document {
		return scanner.Document{Path: path, RelativePath: rel}
	}

	tests := []struct {
		name         string
		docs         []scanner.Document
		projectPaths bool
		wantCount    int
		wantErr      string
		dontWant     string
	}{
		{
			name:      "same basename in different directories",
			docs:      []scanner.Document{doc("/p/README.md", "README.md"), doc("/p/docs/guide/README.md", "guide/README.md")},
			wantCount: 2,
		},
		{
			name:      "same file scanned through two input paths",
			docs:      []scanner.Document{doc("/p/README.md", "README.md"), doc("/p/README.md", "README.md")},
			wantCount: 1,
		},
		{
			name:    "different files with the same output path",
			docs:    []scanner.Document{doc("/p/README.md", "README.md"), doc("/p/docs/README.md", "README.md")},
			wantErr: "both map to README.md",
		},
		{
			name:    "output paths differing only in case, from different roots",
			docs:    []scanner.Document{doc("/p/README.md", "README.md"), doc("/p/docs/readme.md", "readme.md")},
			wantErr: "differ only in case",
		},
		{
			name:    "case-only collision across roots suggests project-relative paths",
			docs:    []scanner.Document{doc("/p/README.md", "README.md"), doc("/p/docs/readme.md", "readme.md")},
			wantErr: "or set output.structure: project",
		},
		{
			name:     "case-only collision within one root does not suggest the setting",
			docs:     []scanner.Document{doc("/p/docs/Guide.md", "Guide.md"), doc("/p/docs/guide.md", "guide.md")},
			wantErr:  "differ only in case",
			dontWant: "output.structure",
		},
		{
			name:    "collision suggests project-relative paths",
			docs:    []scanner.Document{doc("/p/README.md", "README.md"), doc("/p/docs/README.md", "README.md")},
			wantErr: "or set output.structure: project",
		},
		{
			name:         "project-relative collision does not suggest the setting",
			docs:         []scanner.Document{doc("/p/Docs/a.md", "Docs/a.md"), doc("/p/docs/a.md", "docs/a.md")},
			projectPaths: true,
			wantErr:      "rename one of them or remove one of the input paths",
			dontWant:     "output.structure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dedupeDocuments(tt.docs, tt.projectPaths)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("dedupeDocuments() error = %v, want it to contain %q", err, tt.wantErr)
				}
				if tt.dontWant != "" && strings.Contains(err.Error(), tt.dontWant) {
					t.Errorf("dedupeDocuments() error = %v, should not contain %q", err, tt.dontWant)
				}
				return
			}
			if err != nil {
				t.Fatalf("dedupeDocuments() unexpected error: %v", err)
			}
			if len(got) != tt.wantCount {
				t.Errorf("dedupeDocuments() returned %d documents, want %d", len(got), tt.wantCount)
			}
		})
	}
}

// TestDedupeDocumentsSameFileSpellings verifies that one file reached through
// two spellings is kept once when both map to the same page, and kept as two
// pages, like any alias, when they map to different pages.
func TestDedupeDocumentsSameFileSpellings(t *testing.T) {
	dir := t.TempDir()
	guide := filepath.Join(dir, "docs", "guide.md")
	if err := os.MkdirAll(filepath.Dir(guide), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(guide, []byte("# Guide"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(dir, "docs"), filepath.Join(dir, "docslink")); err != nil {
		t.Skipf("symlinks not supported: %v", err)
	}
	linked := filepath.Join(dir, "docslink", "guide.md")

	t.Run("symlink spelling with the same page", func(t *testing.T) {
		docs := []scanner.Document{
			{Path: linked, RelativePath: "guide.md"},
			{Path: guide, RelativePath: "guide.md"},
		}
		got, err := dedupeDocuments(docs, false)
		if err != nil || len(got) != 1 {
			t.Fatalf("dedupeDocuments() = %d documents, %v; want 1, nil", len(got), err)
		}
	})

	t.Run("symlink spelling with a different page", func(t *testing.T) {
		docs := []scanner.Document{
			{Path: linked, RelativePath: "docslink/guide.md"},
			{Path: guide, RelativePath: "docs/guide.md"},
		}
		got, err := dedupeDocuments(docs, true)
		if err != nil || len(got) != 2 {
			t.Fatalf("dedupeDocuments() = %d documents, %v; want 2, nil", len(got), err)
		}
	})

	t.Run("differently-cased directory with the same page", func(t *testing.T) {
		upper := filepath.Join(dir, "Docs", "guide.md")
		if _, err := os.Stat(upper); err != nil {
			t.Skip("filesystem is case-sensitive")
		}
		docs := []scanner.Document{
			{Path: upper, RelativePath: "Docs/guide.md"},
			{Path: guide, RelativePath: "docs/guide.md"},
		}
		got, err := dedupeDocuments(docs, true)
		if err != nil || len(got) != 1 {
			t.Fatalf("dedupeDocuments() = %d documents, %v; want 1, nil", len(got), err)
		}
	})
}

// TestHumanizeBytes verifies the humanizeBytes function
func TestHumanizeBytes(t *testing.T) {
	tests := []struct {
		bytes    int
		expected string
	}{
		{100, "100B"},
		{1024, "1.0KB"},
		{1536, "1.5KB"},
		{1048576, "1.0MB"},
		{1572864, "1.5MB"},
		{1073741824, "1.0GB"},
	}

	for _, tt := range tests {
		result := humanizeBytes(tt.bytes)
		if result != tt.expected {
			t.Errorf("humanizeBytes(%d) = %s, expected %s", tt.bytes, result, tt.expected)
		}
	}
}
