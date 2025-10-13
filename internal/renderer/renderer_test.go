// Package renderer_test contains tests for the renderer package.
package renderer

import (
	"strings"
	"testing"

	"github.com/onedusk/jot/internal/scanner"
	"github.com/onedusk/jot/internal/toc"
)

// TestNewHTMLRenderer tests the creation of a new HTMLRenderer.
func TestNewHTMLRenderer(t *testing.T) {
	renderer := NewHTMLRenderer()
	if renderer == nil {
		t.Fatal("NewHTMLRenderer() returned nil")
	}
}

// TestHTMLRenderer_RenderDocument tests the conversion of markdown content to HTML.
func TestHTMLRenderer_RenderDocument(t *testing.T) {
	tests := []struct {
		name        string
		doc         scanner.Document
		wantContent []string // HTML snippets that should be present
	}{
		{
			name: "simple markdown",
			doc: scanner.Document{
				Title:   "Test Document",
				Content: []byte("# Heading\n\nThis is a paragraph with **bold** text."),
			},
			wantContent: []string{
				"<h1",
				"Heading</h1>",
				"<p>This is a paragraph with <strong>bold</strong> text.</p>",
			},
		},
		{
			name: "code block",
			doc: scanner.Document{
				Title:   "Code Example",
				Content: []byte("```go\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```"),
			},
			wantContent: []string{
				`<pre class="language-go"><code class="language-go">`,
				`func main()`,
				`Println`,
				`</code></pre>`,
			},
		},
		{
			name: "links",
			doc: scanner.Document{
				Title:   "Links Test",
				Content: []byte("[Internal Link](./other.md)\n[External Link](https://example.com)"),
			},
			wantContent: []string{
				`<a href="./other.html">Internal Link</a>`,
				`<a href="https://example.com">External Link</a>`,
			},
		},
		{
			name: "lists",
			doc: scanner.Document{
				Title:   "Lists",
				Content: []byte("- Item 1\n- Item 2\n\n1. First\n2. Second"),
			},
			wantContent: []string{
				"<ul>",
				"<li>Item 1</li>",
				"<li>Item 2</li>",
				"</ul>",
				"<ol>",
				"<li>First</li>",
				"<li>Second</li>",
				"</ol>",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewHTMLRenderer()
			html, err := renderer.RenderDocument(tt.doc)
			if err != nil {
				t.Fatalf("RenderDocument() error = %v", err)
			}

			for _, want := range tt.wantContent {
				if !strings.Contains(html, want) {
					t.Errorf("RenderDocument() missing expected content:\nwant: %s\ngot:\n%s", want, html)
				}
			}
		})
	}
}

// TestHTMLRenderer_RenderPage tests the rendering of a full HTML page.
func TestHTMLRenderer_RenderPage(t *testing.T) {
	doc := scanner.Document{
		Title:        "Test Page",
		RelativePath: "docs/test.md",
		Content:      []byte("# Test Page\n\nContent here"),
	}

	tocRoot := &toc.TOCNode{
		ID:    "root",
		Title: "Table of Contents",
		Children: []*toc.TOCNode{
			{
				ID:    "docs",
				Title: "Documentation",
				Children: []*toc.TOCNode{
					{
						ID:    "test",
						Title: "Test Page",
						Path:  "docs/test.md",
					},
				},
			},
		},
	}

	renderer := NewHTMLRenderer()
	page, err := renderer.RenderPage(doc, &toc.TableOfContents{Root: tocRoot})
	if err != nil {
		t.Fatalf("RenderPage() error = %v", err)
	}

	// Check for expected page elements
	expectedElements := []string{
		"<!DOCTYPE html>",
		"<html",
		"<head>",
		"<title>Test Page | Jot Documentation</title>",
		"<body>",
		"<nav", // Navigation
		"<main",
		"<h1",
		"Test Page</h1>",
		"</main>",
		"</body>",
		"</html>",
	}

	for _, elem := range expectedElements {
		if !strings.Contains(page, elem) {
			t.Errorf("RenderPage() missing expected element: %s", elem)
		}
	}
}

// TestHTMLRenderer_ResolveInternalLinks tests the resolution of internal markdown links.
func TestHTMLRenderer_ResolveInternalLinks(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "markdown to html extension",
			input: `<a href="./doc.md">Link</a>`,
			want:  `<a href="./doc.html">Link</a>`,
		},
		{
			name:  "preserve external links",
			input: `<a href="https://example.com">External</a>`,
			want:  `<a href="https://example.com">External</a>`,
		},
		{
			name:  "handle relative paths",
			input: `<a href="../other/file.md">Other</a>`,
			want:  `<a href="../other/file.html">Other</a>`,
		},
		{
			name:  "preserve anchors",
			input: `<a href="./doc.md#section">Section</a>`,
			want:  `<a href="./doc.html#section">Section</a>`,
		},
	}

	renderer := NewHTMLRenderer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderer.ResolveInternalLinks(tt.input)
			if got != tt.want {
				t.Errorf("ResolveInternalLinks() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGenerateBreadcrumb tests the breadcrumb generation logic.
func TestGenerateBreadcrumb(t *testing.T) {
	tests := []struct {
		name string
		path string
		want []BreadcrumbItem
	}{
		{
			name: "root level",
			path: "index.md",
			want: []BreadcrumbItem{
				{Title: "Home", Path: "/"},
			},
		},
		{
			name: "one level deep",
			path: "docs/getting-started.md",
			want: []BreadcrumbItem{
				{Title: "Home", Path: "/"},
				{Title: "Docs", Path: "/docs/"},
				{Title: "Getting Started", Path: "/docs/getting-started.html"},
			},
		},
		{
			name: "multiple levels",
			path: "docs/api/reference/endpoints.md",
			want: []BreadcrumbItem{
				{Title: "Home", Path: "/"},
				{Title: "Docs", Path: "/docs/"},
				{Title: "Api", Path: "/docs/api/"},
				{Title: "Reference", Path: "/docs/api/reference/"},
				{Title: "Endpoints", Path: "/docs/api/reference/endpoints.html"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateBreadcrumb(tt.path, "")
			if len(got) != len(tt.want) {
				t.Errorf("GenerateBreadcrumb() returned %d items, want %d", len(got), len(tt.want))
				return
			}
			for i, item := range got {
				if item.Title != tt.want[i].Title || item.Path != tt.want[i].Path {
					t.Errorf("GenerateBreadcrumb()[%d] = {%s, %s}, want {%s, %s}",
						i, item.Title, item.Path, tt.want[i].Title, tt.want[i].Path)
				}
			}
		})
	}
}

// TestHTMLRenderer_GenerateNavigation tests the navigation tree generation.
func TestHTMLRenderer_GenerateNavigation(t *testing.T) {
	tocRoot := &toc.TOCNode{
		ID:    "root",
		Title: "Root",
		Children: []*toc.TOCNode{
			{
				ID:    "getting-started",
				Title: "Getting Started",
				Path:  "getting-started.md",
			},
			{
				ID:    "guides",
				Title: "Guides",
				Children: []*toc.TOCNode{
					{
						ID:    "advanced",
						Title: "Advanced",
						Path:  "guides/advanced.md",
					},
				},
			},
		},
	}

	renderer := NewHTMLRenderer()
	nav := renderer.GenerateNavigation(tocRoot, "guides/advanced.md", "")

	// Check for expected navigation structure
	if !strings.Contains(nav, `class="nav-tree"`) {
		t.Error("GenerateNavigation() missing nav-tree class")
	}
	if !strings.Contains(nav, "Getting Started") {
		t.Error("GenerateNavigation() missing Getting Started link")
	}
	if !strings.Contains(nav, "Advanced") {
		t.Error("GenerateNavigation() missing Advanced link")
	}
	if !strings.Contains(nav, `active`) {
		t.Error("GenerateNavigation() missing active class for current page")
	}
}
