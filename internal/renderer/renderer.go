// Package renderer provides functionality for converting markdown documents into HTML.
// It uses the blackfriday library for markdown processing and includes features
// like syntax highlighting, task lists, and template-based page rendering.
package renderer

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html"
	"html/template"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/onedusk/jot/internal/toc"
	"github.com/onedusk/jot/pkg/scanner"
	"github.com/russross/blackfriday/v2"
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

	// Convert blockquotes with Note:/Warning: to callout boxes
	htmlStr = r.processCallouts(htmlStr)

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

// processCallouts converts blockquotes starting with Note: or Warning: into styled callout boxes.
func (r *HTMLRenderer) processCallouts(html string) string {
	// Match blockquotes that start with <strong>Note:</strong> or <strong>Warning:</strong>
	noteRegex := regexp.MustCompile(`<blockquote>\s*<p>\s*<strong>Note:?</strong>`)
	html = noteRegex.ReplaceAllString(html, `<div class="callout callout-info"><p><strong>Note:</strong>`)

	warningRegex := regexp.MustCompile(`<blockquote>\s*<p>\s*<strong>Warning:?</strong>`)
	html = warningRegex.ReplaceAllString(html, `<div class="callout callout-warning"><p><strong>Warning:</strong>`)

	// Close the callout divs: replace </blockquote> after callout divs with </div>
	html = regexp.MustCompile(`(<div class="callout callout-(?:info|warning)">(?:.*?))</blockquote>`).ReplaceAllString(html, `$1</div>`)

	return html
}

// RenderPage renders a full HTML page for a given document, including layout,
// navigation, breadcrumbs, and the document's content.
func (r *HTMLRenderer) RenderPage(doc scanner.Document, tableOfContents *toc.TableOfContents, config SiteConfig) (string, error) {
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

	// Compute prev/next page links
	prevPage, nextPage := r.computeAdjacentPages(tableOfContents.Root, doc.RelativePath, relativePrefix)

	// Create page data
	data := PageData{
		Title:          doc.Title,
		Content:        template.HTML(content),
		Navigation:     template.HTML(nav),
		Breadcrumb:     breadcrumb,
		RelativePrefix: relativePrefix,
		ProjectName:    config.ProjectName,
		NavLinks:       config.NavLinks,
		PrevPage:       prevPage,
		NextPage:       nextPage,
		MarkdownSource: base64.StdEncoding.EncodeToString(doc.Content),
		MarkdownPath:   doc.RelativePath,
	}

	// Render using template
	return r.renderTemplate(data)
}

// flattenTOC collects all leaf nodes (documents) from the TOC tree in order.
func flattenTOC(node *toc.TOCNode) []toc.TOCNode {
	var result []toc.TOCNode
	for _, child := range node.Children {
		if child.Path != "" {
			result = append(result, *child)
		}
		result = append(result, flattenTOC(child)...)
	}
	return result
}

// computeAdjacentPages finds the previous and next documents relative to the current page.
func (r *HTMLRenderer) computeAdjacentPages(root *toc.TOCNode, currentPath string, relativePrefix string) (*PageLink, *PageLink) {
	leaves := flattenTOC(root)

	// Deduplicate (a node with children and a path would appear twice)
	seen := make(map[string]bool)
	var unique []toc.TOCNode
	for _, leaf := range leaves {
		if !seen[leaf.Path] {
			seen[leaf.Path] = true
			unique = append(unique, leaf)
		}
	}

	var prevPage, nextPage *PageLink
	for i, leaf := range unique {
		if leaf.Path == currentPath {
			if i > 0 {
				htmlPath := strings.Replace(unique[i-1].Path, ".md", ".html", 1)
				prevPage = &PageLink{
					Title: unique[i-1].Title,
					Path:  relativePrefix + htmlPath,
				}
			}
			if i < len(unique)-1 {
				htmlPath := strings.Replace(unique[i+1].Path, ".md", ".html", 1)
				nextPage = &PageLink{
					Title: unique[i+1].Title,
					Path:  relativePrefix + htmlPath,
				}
			}
			break
		}
	}
	return prevPage, nextPage
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
	buf.WriteString(`<div class="nav-tree">`)
	r.renderNavSection(&buf, root, currentPath, relativePrefix, 0)
	buf.WriteString(`</div>`)
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
			relativePrefix, html.EscapeString(htmlPath), activeClass, html.EscapeString(node.Title)))

		if depth == 1 {
			buf.WriteString(`</ul></div>`)
		}
	} else {
		// This is a directory - render as a section
		buf.WriteString(`<div class="nav-section">`)
		buf.WriteString(fmt.Sprintf(`<div class="nav-section-title">%s</div>`, html.EscapeString(node.Title)))
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
					relativePrefix, html.EscapeString(htmlPath), activeClass, html.EscapeString(child.Title)))
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

// SiteConfig holds site-wide configuration passed from jot.yml to the renderer.
type SiteConfig struct {
	ProjectName string
	NavLinks    []NavLink
}

// NavLink represents a navigation link in the header.
type NavLink struct {
	Label string
	Href  string
}

// PageLink represents a link to an adjacent page (previous or next).
type PageLink struct {
	Title string
	Path  string
}

// PageData holds the data passed to the HTML template for rendering a single page.
type PageData struct {
	Title          string
	Content        template.HTML
	Navigation     template.HTML
	Breadcrumb     []BreadcrumbItem
	RelativePrefix string
	ProjectName    string
	NavLinks       []NavLink
	PrevPage       *PageLink
	NextPage       *PageLink
	MarkdownSource string // base64-encoded raw markdown for clipboard copy
	MarkdownPath   string // relative .md file path for "Open Markdown" link
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

	// Special case for index
	if path == "index.md" {
		return []BreadcrumbItem{
			{Title: "Home", Path: "/"},
		}
	}

	parts := strings.Split(path, "/")

	breadcrumbs := []BreadcrumbItem{
		{Title: "Home", Path: "/"},
	}

	// Build breadcrumb path
	currentPath := ""
	for i, part := range parts {
		if part == "" || part == "." {
			continue
		}

		// Add path separator
		if currentPath != "" {
			currentPath += "/"
		}
		currentPath += part

		// Create breadcrumb item
		title := strings.Title(strings.ReplaceAll(strings.TrimSuffix(part, ".md"), "-", " "))

		// For directories (not the last part or doesn't end with .md), append /
		// For files (last part and ends with .md), replace .md with .html
		var href string
		if i == len(parts)-1 && strings.HasSuffix(part, ".md") {
			// This is a file
			href = "/" + strings.Replace(currentPath, ".md", ".html", 1)
		} else {
			// This is a directory
			href = "/" + strings.TrimSuffix(currentPath, ".md") + "/"
		}

		breadcrumbs = append(breadcrumbs, BreadcrumbItem{
			Title: title,
			Path:  href,
		})
	}

	return breadcrumbs
}
