// Package compiler provides functionality for compiling documentation from markdown files into HTML.
package compiler

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/onedusk/jot/internal/renderer"
	"github.com/onedusk/jot/internal/search"
	"github.com/onedusk/jot/internal/toc"
	"github.com/onedusk/jot/pkg/scanner"
	"github.com/onedusk/jot/web"
)

// Compiler orchestrates the documentation build process. It handles file processing,
// HTML rendering, asset copying, and search index generation.
type Compiler struct {
	outputPath string
	renderer   *renderer.HTMLRenderer
	siteConfig renderer.SiteConfig
}

// NewCompiler creates a new documentation compiler. It takes the output path
// where the compiled documentation will be stored and a site configuration.
func NewCompiler(outputPath string, siteConfig renderer.SiteConfig) *Compiler {
	return &Compiler{
		outputPath: outputPath,
		renderer:   renderer.NewHTMLRenderer(),
		siteConfig: siteConfig,
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

	// Static hosts serve index.html at the site root, so make sure it exists
	switch {
	case c.hasDocument(documents, "index.md"):
		// Already written as index.html
	case c.hasDocument(documents, "README.md"):
		if err := c.copyPage("README.html", "index.html"); err != nil {
			return fmt.Errorf("failed to write index page: %w", err)
		}
	default:
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
	html, err := c.renderer.RenderPage(doc, toc, c.siteConfig)
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
	if err := os.WriteFile(outputPath, []byte(html), 0644); err != nil {
		return err
	}

	// Copy source markdown file for "Open Markdown" feature. When building into
	// the source directory the copy would overwrite the source itself (with its
	// frontmatter stripped), so leave the source in place instead.
	mdOutputPath := filepath.Join(c.outputPath, doc.RelativePath)
	if isSameFile(doc.Path, mdOutputPath) {
		return nil
	}
	mdOutputDir := filepath.Dir(mdOutputPath)
	if err := os.MkdirAll(mdOutputDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(mdOutputPath, doc.Content, 0644)
}

// isSameFile reports whether a and b both exist and are the same file.
func isSameFile(a, b string) bool {
	if a == "" {
		return false
	}
	aInfo, err := os.Stat(a)
	if err != nil {
		return false
	}
	bInfo, err := os.Stat(b)
	if err != nil {
		return false
	}
	return os.SameFile(aInfo, bInfo)
}

// getOutputPath converts a markdown file's relative path to its corresponding HTML output path.
func (c *Compiler) getOutputPath(relativePath string) string {
	// Replace .md with .html
	htmlPath := strings.Replace(relativePath, ".md", ".html", 1)

	// Join with output directory
	return filepath.Join(c.outputPath, htmlPath)
}

// hasDocument checks if the list of documents includes one at the given relative path.
func (c *Compiler) hasDocument(documents []scanner.Document, relativePath string) bool {
	for _, doc := range documents {
		if doc.RelativePath == relativePath {
			return true
		}
	}
	return false
}

// copyPage copies an already-written page to another path within the output directory.
func (c *Compiler) copyPage(from, to string) error {
	content, err := os.ReadFile(filepath.Join(c.outputPath, from))
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(c.outputPath, to), content, 0644)
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

// copyAssets writes the static assets (CSS, JS) embedded in the binary to the output directory.
func (c *Compiler) copyAssets() error {
	assetsDir := filepath.Join(c.outputPath, "assets")
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		return err
	}

	entries, err := fs.ReadDir(web.Assets, web.AssetsDir)
	if err != nil {
		return fmt.Errorf("failed to read embedded assets: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		content, err := fs.ReadFile(web.Assets, path.Join(web.AssetsDir, entry.Name()))
		if err != nil {
			return fmt.Errorf("failed to read embedded asset %s: %w", entry.Name(), err)
		}
		if err := os.WriteFile(filepath.Join(assetsDir, entry.Name()), content, 0644); err != nil {
			return fmt.Errorf("failed to write asset %s: %w", entry.Name(), err)
		}
	}

	return nil
}
