// Package compiler provides functionality for compiling documentation from markdown files into HTML.
package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/thrive/jot/internal/scanner"
	"github.com/thrive/jot/internal/toc"
)

// MarkdownCompiler generates markdown output with enhanced navigation features
// like breadcrumbs and consolidated files for LLM consumption.
type MarkdownCompiler struct {
	outputPath string
}

// NewMarkdownCompiler creates a new markdown compiler with the specified output path.
func NewMarkdownCompiler(outputPath string) *MarkdownCompiler {
	return &MarkdownCompiler{
		outputPath: outputPath,
	}
}

// Compile processes documents and generates enhanced markdown output. This includes
// individual files with navigation, a consolidated file, and a markdown TOC.
func (m *MarkdownCompiler) Compile(documents []scanner.Document, tableOfContents *toc.TableOfContents) error {
	// Create markdown output directory
	markdownDir := filepath.Join(m.outputPath, "markdown")
	if err := os.MkdirAll(markdownDir, 0755); err != nil {
		return fmt.Errorf("failed to create markdown directory: %w", err)
	}

	// Process each document
	for _, doc := range documents {
		if err := m.compileDocument(doc, tableOfContents, markdownDir); err != nil {
			return fmt.Errorf("failed to compile markdown %s: %w", doc.RelativePath, err)
		}
	}

	// Generate consolidated markdown for LLMs
	if err := m.generateConsolidatedMarkdown(documents, markdownDir); err != nil {
		return fmt.Errorf("failed to generate consolidated markdown: %w", err)
	}

	// Generate TOC in markdown format
	if err := m.generateMarkdownTOC(tableOfContents, markdownDir); err != nil {
		return fmt.Errorf("failed to generate markdown TOC: %w", err)
	}

	return nil
}

// compileDocument copies and enhances a single markdown document with navigation metadata.
func (m *MarkdownCompiler) compileDocument(doc scanner.Document, toc *toc.TableOfContents, outputDir string) error {
	// Determine output path
	outputPath := filepath.Join(outputDir, doc.RelativePath)

	// Create directory if needed
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Enhanced markdown with navigation
	var enhanced strings.Builder

	// Add navigation header
	enhanced.WriteString("---\n")
	enhanced.WriteString(fmt.Sprintf("title: %s\n", doc.Title))
	enhanced.WriteString(fmt.Sprintf("path: %s\n", doc.RelativePath))
	enhanced.WriteString(fmt.Sprintf("modified: %s\n", doc.ModTime.Format("2006-01-02")))
	enhanced.WriteString("---\n\n")

	// Add breadcrumb navigation
	enhanced.WriteString(m.generateBreadcrumb(doc.RelativePath))
	enhanced.WriteString("\n\n")

	// Add original content
	enhanced.Write(doc.Content)

	// Add footer navigation
	enhanced.WriteString("\n\n---\n")
	enhanced.WriteString("[← Back to TOC](./index.md) | ")
	enhanced.WriteString("[View All Docs](./all-docs.md)\n")

	// Write enhanced markdown file
	return os.WriteFile(outputPath, []byte(enhanced.String()), 0644)
}

// generateBreadcrumb creates breadcrumb navigation as a string for a given file path.
func (m *MarkdownCompiler) generateBreadcrumb(path string) string {
	parts := strings.Split(filepath.ToSlash(path), "/")
	breadcrumbs := []string{"[Home](index.md)"}

	for i, part := range parts[:len(parts)-1] {
		// Build relative path to directory index
		relPath := strings.Repeat("../", len(parts)-i-2)
		breadcrumbs = append(breadcrumbs, fmt.Sprintf("[%s](%sindex.md)", part, relPath))
	}

	// Add current file
	if len(parts) > 0 {
		breadcrumbs = append(breadcrumbs, parts[len(parts)-1])
	}

	return "📍 " + strings.Join(breadcrumbs, " / ")
}

// generateConsolidatedMarkdown creates a single markdown file containing all documents.
// This is useful for consumption by Large Language Models (LLMs).
func (m *MarkdownCompiler) generateConsolidatedMarkdown(documents []scanner.Document, outputDir string) error {
	outputPath := filepath.Join(outputDir, "all-docs.md")

	var consolidated strings.Builder
	consolidated.WriteString("# Complete Documentation Archive\n\n")
	consolidated.WriteString("This file contains all documentation in a single markdown file for easy LLM consumption.\n\n")
	consolidated.WriteString("## Table of Contents\n\n")

	// Generate TOC
	for i, doc := range documents {
		consolidated.WriteString(fmt.Sprintf("%d. [%s](#doc-%d)\n", i+1, doc.Title, i))
	}
	consolidated.WriteString("\n---\n\n")

	// Add all documents
	for i, doc := range documents {
		consolidated.WriteString(fmt.Sprintf("<a name=\"doc-%d\"></a>\n\n", i))
		consolidated.WriteString(fmt.Sprintf("## %s\n\n", doc.Title))
		consolidated.WriteString(fmt.Sprintf("**Path:** `%s`\n", doc.RelativePath))
		consolidated.WriteString(fmt.Sprintf("**Modified:** %s\n\n", doc.ModTime.Format("2006-01-02")))
		consolidated.Write(doc.Content)
		consolidated.WriteString("\n\n---\n\n")
	}

	return os.WriteFile(outputPath, []byte(consolidated.String()), 0644)
}

// generateMarkdownTOC creates a markdown version of the Table of Contents.
func (m *MarkdownCompiler) generateMarkdownTOC(toc *toc.TableOfContents, outputDir string) error {
	outputPath := filepath.Join(outputDir, "index.md")

	var tocMarkdown strings.Builder
	tocMarkdown.WriteString("# Documentation Index\n\n")

	// Process root children
	for _, child := range toc.Root.Children {
		m.nodeToMarkdown(&tocMarkdown, child, 0)
	}

	tocMarkdown.WriteString("\n---\n\n")
	tocMarkdown.WriteString("[View All Documentation](all-docs.md) | ")
	tocMarkdown.WriteString("[XML TOC](../toc.xml) | ")
	tocMarkdown.WriteString("[HTML Version](../index.html)\n")

	return os.WriteFile(outputPath, []byte(tocMarkdown.String()), 0644)
}

// nodeToMarkdown converts a TOC node and its children to a markdown list representation.
func (m *MarkdownCompiler) nodeToMarkdown(builder *strings.Builder, node *toc.TOCNode, depth int) {
	indent := strings.Repeat("  ", depth)

	if node.IsLeaf() {
		// Document link
		builder.WriteString(fmt.Sprintf("%s- [%s](%s)\n", indent, node.Title, node.Path))
	} else {
		// Section header
		if depth == 0 {
			builder.WriteString(fmt.Sprintf("\n## %s\n\n", node.Title))
		} else {
			builder.WriteString(fmt.Sprintf("%s- **%s**\n", indent, node.Title))
		}

		// Process children
		for _, child := range node.Children {
			if depth == 0 && !child.IsLeaf() {
				m.nodeToMarkdown(builder, child, depth)
			} else {
				m.nodeToMarkdown(builder, child, depth+1)
			}
		}
	}
}
