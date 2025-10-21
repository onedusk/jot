package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

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
