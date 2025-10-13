// Package renderer provides functionality for converting markdown documents into HTML.
// It uses the blackfriday library for markdown processing and includes features
// like syntax highlighting, task lists, and template-based page rendering.
package renderer

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/russross/blackfriday/v2"
	"github.com/onedusk/jot/internal/scanner"
	"github.com/onedusk/jot/internal/toc"
)

// HTMLRenderer is responsible for converting markdown documents into final HTML pages.
// It manages templates, markdown-to-HTML conversion, and generation of navigation elements.
type HTMLRenderer struct {
	templates *template.Template
}

// NewHTMLRenderer creates and returns a new HTMLRenderer instance.
func NewHTMLRenderer() *HTMLRenderer {
	return &HTMLRenderer{}
}

// getRelativePrefix calculates the relative path prefix (e.g., "../") needed to
// access root-level assets from a nested document.
func (r *HTMLRenderer) getRelativePrefix(path string) string {
	// Clean the path
	path = filepath.ToSlash(path)

	// Get the directory part
	dir := filepath.Dir(path)

	// If at root, no prefix needed
	if dir == "." || dir == "/" || dir == "" {
		return ""
	}

	// Count the depth by splitting on /
	parts := strings.Split(dir, "/")
	depth := 0
	for _, part := range parts {
		if part != "" && part != "." {
			depth++
		}
	}

	// Generate relative prefix (../ for each level)
	if depth == 0 {
		return ""
	}

	prefix := ""
	for i := 0; i < depth; i++ {
		prefix += "../"
	}
	return prefix
}

// RenderDocument converts the markdown content of a document to an HTML string.
// It enables several markdown extensions for features like tables, code blocks, and footnotes.
func (r *HTMLRenderer) RenderDocument(doc scanner.Document) (string, error) {
	// Convert markdown to HTML using blackfriday with all extensions
	extensions := blackfriday.CommonExtensions |
		blackfriday.AutoHeadingIDs |
		blackfriday.Tables |
		blackfriday.FencedCode |
		blackfriday.Autolink |
		blackfriday.Strikethrough |
		blackfriday.SpaceHeadings |
		blackfriday.BackslashLineBreak |
		blackfriday.DefinitionLists |
		blackfriday.Footnotes

	renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		Flags: blackfriday.CommonHTMLFlags |
			blackfriday.FootnoteReturnLinks,
	})

	html := blackfriday.Run(doc.Content,
		blackfriday.WithExtensions(extensions),
		blackfriday.WithRenderer(renderer))

	// Post-process HTML for enhanced features
	htmlStr := string(html)

	// Add language classes to code blocks for Prism.js
	htmlStr = r.enhanceCodeBlocks(htmlStr)

	// Resolve internal links (convert .md to .html)
	htmlStr = r.ResolveInternalLinks(htmlStr)

	// Process task lists (GitHub-style checkboxes)
	htmlStr = r.processTaskLists(htmlStr)

	return htmlStr, nil
}

// enhanceCodeBlocks adds language classes to <pre> and <code> tags in the HTML
// to enable syntax highlighting with libraries like Prism.js or highlight.js.
func (r *HTMLRenderer) enhanceCodeBlocks(html string) string {
	// Regular expression to find code blocks with language hints
	codeBlockRegex := regexp.MustCompile(`<pre><code class="language-(\w+)">`)
	html = codeBlockRegex.ReplaceAllString(html, `<pre class="language-$1"><code class="language-$1">`)

	// Also handle code blocks without language specification
	plainCodeRegex := regexp.MustCompile(`<pre><code>`)
	html = plainCodeRegex.ReplaceAllString(html, `<pre><code>`)

	return html
}

// processTaskLists converts GitHub-style task list markdown (e.g., "- [x] Task")
// into disabled HTML checkboxes.
func (r *HTMLRenderer) processTaskLists(html string) string {
	// Convert [ ] to unchecked checkbox
	html = strings.ReplaceAll(html, `<li>[ ]`, `<li class="task-list-item"><input type="checkbox" disabled>`)
	// Convert [x] or [X] to checked checkbox
	html = strings.ReplaceAll(html, `<li>[x]`, `<li class="task-list-item"><input type="checkbox" disabled checked>`)
	html = strings.ReplaceAll(html, `<li>[X]`, `<li class="task-list-item"><input type="checkbox" disabled checked>`)

	// Add task-list class to ul elements containing task items
	taskListRegex := regexp.MustCompile(`<ul>\s*<li class="task-list-item">`)
	html = taskListRegex.ReplaceAllString(html, `<ul class="task-list"><li class="task-list-item">`)

	return html
}

// RenderPage renders a full HTML page for a given document, including layout,
// navigation, breadcrumbs, and the document's content.
func (r *HTMLRenderer) RenderPage(doc scanner.Document, tableOfContents *toc.TableOfContents) (string, error) {
	// Render the document content
	content, err := r.RenderDocument(doc)
	if err != nil {
		return "", err
	}

	// Calculate relative path prefix based on document depth
	relativePrefix := r.getRelativePrefix(doc.RelativePath)

	// Generate breadcrumb
	breadcrumb := GenerateBreadcrumb(doc.RelativePath, relativePrefix)

	// Generate navigation
	nav := r.GenerateNavigation(tableOfContents.Root, doc.RelativePath, relativePrefix)

	// Create page data
	data := PageData{
		Title:          doc.Title,
		Content:        template.HTML(content),
		Navigation:     template.HTML(nav),
		Breadcrumb:     breadcrumb,
		RelativePrefix: relativePrefix,
	}

	// Render using template
	return r.renderTemplate(data)
}

// ResolveInternalLinks converts relative links to markdown files (.md) into
// links to the corresponding HTML files (.html) within the generated HTML.
func (r *HTMLRenderer) ResolveInternalLinks(html string) string {
	// Regular expression to find href attributes with .md files
	linkRegex := regexp.MustCompile(`href="([^"]+\.md(?:#[^"]*)?)"`)

	return linkRegex.ReplaceAllStringFunc(html, func(match string) string {
		// Extract the URL
		urlMatch := linkRegex.FindStringSubmatch(match)
		if len(urlMatch) < 2 {
			return match
		}

		url := urlMatch[1]

		// Skip external links
		if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "//") {
			return match
		}

		// Replace .md with .html
		newURL := strings.Replace(url, ".md", ".html", 1)
		return fmt.Sprintf(`href="%s"`, newURL)
	})
}

// GenerateNavigation creates the HTML for the sidebar navigation tree based on the
// table of contents, highlighting the current page.
func (r *HTMLRenderer) GenerateNavigation(root *toc.TOCNode, currentPath string, relativePrefix string) string {
	var buf bytes.Buffer
	r.renderNavSection(&buf, root, currentPath, relativePrefix, 0)
	return buf.String()
}

// renderNavSection is a recursive helper function that builds the HTML for the navigation menu.
func (r *HTMLRenderer) renderNavSection(buf *bytes.Buffer, node *toc.TOCNode, currentPath string, relativePrefix string, depth int) {
	// Skip the root node at depth 0
	if depth == 0 {
		for _, child := range node.Children {
			r.renderNavSection(buf, child, currentPath, relativePrefix, depth+1)
		}
		return
	}

	// Check if this is a directory or file
	if node.Path != "" {
		// This is a file - render as a nav item
		htmlPath := strings.Replace(node.Path, ".md", ".html", 1)
		activeClass := ""
		if node.Path == currentPath {
			activeClass = " active"
		}

		// Start nav item
		if depth == 1 {
			// Top level - might need section title
			buf.WriteString(`<div class="nav-section">`)
			buf.WriteString(`<ul class="nav-list">`)
		}

		buf.WriteString(fmt.Sprintf(`<li class="nav-item"><a href="%s%s" class="nav-link%s">%s</a></li>`,
			relativePrefix, htmlPath, activeClass, node.Title))

		if depth == 1 {
			buf.WriteString(`</ul></div>`)
		}
	} else {
		// This is a directory - render as a section
		buf.WriteString(`<div class="nav-section">`)
		buf.WriteString(fmt.Sprintf(`<div class="nav-section-title">%s</div>`, node.Title))
		buf.WriteString(`<ul class="nav-list">`)

		// Render children
		for _, child := range node.Children {
			if child.Path != "" {
				htmlPath := strings.Replace(child.Path, ".md", ".html", 1)
				activeClass := ""
				if child.Path == currentPath {
					activeClass = " active"
				}
				buf.WriteString(fmt.Sprintf(`<li class="nav-item"><a href="%s%s" class="nav-link%s">%s</a></li>`,
					relativePrefix, htmlPath, activeClass, child.Title))
			} else {
				// Nested directory
				r.renderNavSection(buf, child, currentPath, relativePrefix, depth+1)
			}
		}

		buf.WriteString(`</ul></div>`)
	}
}

// containsActivePage recursively checks if a TOC node or any of its children
// corresponds to the currently active page path.
func (r *HTMLRenderer) containsActivePage(node *toc.TOCNode, currentPath string) bool {
	if node.Path == currentPath {
		return true
	}
	for _, child := range node.Children {
		if r.containsActivePage(child, currentPath) {
			return true
		}
	}
	return false
}

// renderTemplate executes the HTML template with the provided page data.
func (r *HTMLRenderer) renderTemplate(data PageData) (string, error) {
	// Use the exact template design from proposal_template_001.html
	tmplStr := htmlTemplate

	// Parse template
	tmpl, err := template.New("page").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// PageData holds the data passed to the HTML template for rendering a single page.
type PageData struct {
	Title          string
	Content        template.HTML
	Navigation     template.HTML
	Breadcrumb     []BreadcrumbItem
	RelativePrefix string
}

// BreadcrumbItem represents a single item in a breadcrumb navigation trail.
type BreadcrumbItem struct {
	Title string
	Path  string
}

// GenerateBreadcrumb creates a slice of BreadcrumbItem for a given document path,
// which can be used to render a breadcrumb navigation menu.
func GenerateBreadcrumb(path string, relativePrefix string) []BreadcrumbItem {
	// Clean and split the path
	path = filepath.ToSlash(path)
	parts := strings.Split(path, "/")

	breadcrumbs := []BreadcrumbItem{
		{Title: "Home", Path: relativePrefix + "index.html"},
	}

	// Build breadcrumb path
	currentPath := ""
	for i, part := range parts {
		if part == "" || part == "." {
			continue
		}

		// For the last part, remove .md extension
		if i == len(parts)-1 {
			part = strings.TrimSuffix(part, ".md")
		}

		// Add path separator
		if currentPath != "" {
			currentPath += "/"
		}
		currentPath += part

		// Create breadcrumb item
		title := strings.Title(strings.ReplaceAll(part, "_", " "))
		href := relativePrefix + currentPath
		if !strings.HasSuffix(href, ".html") {
			href += ".html"
		}

		breadcrumbs = append(breadcrumbs, BreadcrumbItem{
			Title: title,
			Path:  href,
		})
	}

	return breadcrumbs
}
