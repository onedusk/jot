// Package compiler provides functionality for compiling documentation from markdown files into HTML.
package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/thrive/jot/internal/renderer"
	"github.com/thrive/jot/internal/scanner"
	"github.com/thrive/jot/internal/search"
	"github.com/thrive/jot/internal/toc"
)

// Compiler orchestrates the documentation build process. It handles file processing,
// HTML rendering, asset copying, and search index generation.
type Compiler struct {
	outputPath string
	renderer   *renderer.HTMLRenderer
}

// NewCompiler creates a new documentation compiler. It takes the output path
// where the compiled documentation will be stored.
func NewCompiler(outputPath string) *Compiler {
	return &Compiler{
		outputPath: outputPath,
		renderer:   renderer.NewHTMLRenderer(),
	}
}

// Compile processes a slice of documents, generates HTML output, and creates a search index.
// It also ensures that an index page is created if one doesn't exist.
func (c *Compiler) Compile(documents []scanner.Document, tableOfContents *toc.TableOfContents) error {
	// Ensure output directory exists
	if err := os.MkdirAll(c.outputPath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Process each document
	for _, doc := range documents {
		if err := c.compileDocument(doc, tableOfContents); err != nil {
			return fmt.Errorf("failed to compile %s: %w", doc.RelativePath, err)
		}
	}

	// Generate index page if not present
	if !c.hasIndexPage(documents) {
		if err := c.generateIndexPage(tableOfContents); err != nil {
			return fmt.Errorf("failed to generate index page: %w", err)
		}
	}

	// Generate search index
	if err := c.generateSearchIndex(documents); err != nil {
		return fmt.Errorf("failed to generate search index: %w", err)
	}

	// Copy assets
	if err := c.copyAssets(); err != nil {
		return fmt.Errorf("failed to copy assets: %w", err)
	}

	return nil
}

// compileDocument compiles a single document to HTML. It renders the document
// using the HTML renderer and writes the output to the appropriate file.
func (c *Compiler) compileDocument(doc scanner.Document, toc *toc.TableOfContents) error {
	// Render the page
	html, err := c.renderer.RenderPage(doc, toc)
	if err != nil {
		return err
	}

	// Determine output path
	outputPath := c.getOutputPath(doc.RelativePath)

	// Create directory if needed
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	// Write HTML file
	return os.WriteFile(outputPath, []byte(html), 0644)
}

// getOutputPath converts a markdown file's relative path to its corresponding HTML output path.
func (c *Compiler) getOutputPath(relativePath string) string {
	// Replace .md with .html
	htmlPath := strings.Replace(relativePath, ".md", ".html", 1)

	// Join with output directory
	return filepath.Join(c.outputPath, htmlPath)
}

// hasIndexPage checks if the list of documents includes an index page (index.md or README.md).
func (c *Compiler) hasIndexPage(documents []scanner.Document) bool {
	for _, doc := range documents {
		if doc.RelativePath == "index.md" || doc.RelativePath == "README.md" {
			return true
		}
	}
	return false
}

// generateIndexPage creates a default index page if one is not found in the documents.
// It includes a table of contents.
func (c *Compiler) generateIndexPage(tableOfContents *toc.TableOfContents) error {
	// Create a synthetic index document
	indexDoc := scanner.Document{
		Title:        "Documentation",
		RelativePath: "index.md",
		Content:      []byte("# Documentation\n\nWelcome to the documentation.\n\n## Table of Contents\n\n" + c.generateTOCMarkdown(tableOfContents.Root)),
	}

	// Compile it
	return c.compileDocument(indexDoc, tableOfContents)
}

// generateTOCMarkdown creates a markdown representation of the Table of Contents.
func (c *Compiler) generateTOCMarkdown(node *toc.TOCNode) string {
	var sb strings.Builder
	c.writeTOCNode(&sb, node, 0)
	return sb.String()
}

// writeTOCNode recursively writes TOC nodes as a markdown list to the provided strings.Builder.
func (c *Compiler) writeTOCNode(sb *strings.Builder, node *toc.TOCNode, depth int) {
	// Skip root node
	if depth == 0 {
		for _, child := range node.Children {
			c.writeTOCNode(sb, child, depth+1)
		}
		return
	}

	indent := strings.Repeat("  ", depth-1)

	if node.Path != "" {
		// Document link
		htmlPath := strings.Replace(node.Path, ".md", ".html", 1)
		sb.WriteString(fmt.Sprintf("%s- [%s](%s)\n", indent, node.Title, htmlPath))
	} else {
		// Section header
		sb.WriteString(fmt.Sprintf("%s- **%s**\n", indent, node.Title))
		for _, child := range node.Children {
			c.writeTOCNode(sb, child, depth+1)
		}
	}
}

// generateSearchIndex creates the search index JSON file by building and saving an index of the documents.
func (c *Compiler) generateSearchIndex(documents []scanner.Document) error {
	indexer := search.NewIndexer(c.outputPath)

	// Build index
	index, err := indexer.BuildIndex(documents)
	if err != nil {
		return err
	}

	// Save index
	return indexer.SaveIndex(index)
}

// copyAssets copies static assets (CSS, JS) required for the documentation to the output directory.
func (c *Compiler) copyAssets() error {
	assetsDir := filepath.Join(c.outputPath, "assets")
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		return err
	}

	// Copy syntax highlighting CSS if it exists
	syntaxCSSPath := filepath.Join("web", "templates", "assets", "syntax-highlighting.css")
	if syntaxContent, err := os.ReadFile(syntaxCSSPath); err == nil {
		syntaxOutputPath := filepath.Join(assetsDir, "syntax-highlighting.css")
		if err := os.WriteFile(syntaxOutputPath, syntaxContent, 0644); err != nil {
			return fmt.Errorf("failed to copy syntax highlighting CSS: %w", err)
		}
	}

	// Copy search.js if it exists
	searchJSPath := filepath.Join("web", "templates", "assets", "search.js")
	if searchContent, err := os.ReadFile(searchJSPath); err == nil {
		searchOutputPath := filepath.Join(assetsDir, "search.js")
		if err := os.WriteFile(searchOutputPath, searchContent, 0644); err != nil {
			return fmt.Errorf("failed to copy search JS: %w", err)
		}
	}

	// Copy style.css if it exists
	styleCSSPath := filepath.Join("web", "templates", "assets", "style.css")
	if styleContent, err := os.ReadFile(styleCSSPath); err == nil {
		styleOutputPath := filepath.Join(assetsDir, "style.css")
		if err := os.WriteFile(styleOutputPath, styleContent, 0644); err != nil {
			return fmt.Errorf("failed to copy style.css: %w", err)
		}
	}

	// Copy highlight.js if it exists
	highlightJSPath := filepath.Join("web", "templates", "assets", "highlight.js")
	if highlightContent, err := os.ReadFile(highlightJSPath); err == nil {
		highlightOutputPath := filepath.Join(assetsDir, "highlight.js")
		if err := os.WriteFile(highlightOutputPath, highlightContent, 0644); err != nil {
			return fmt.Errorf("failed to copy highlight.js: %w", err)
		}
	}

	return nil
}
