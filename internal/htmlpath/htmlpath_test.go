package htmlpath

import "testing"

func TestFromMarkdown(t *testing.T) {
	tests := []struct{ in, want string }{
		{"guide.md", "guide.html"},
		{"docs/guide.md", "docs/guide.html"},
		{"site.mdocs/a.md", "site.mdocs/a.html"},
		{"notes.md.md", "notes.md.html"},
		{"Guide.MD", "Guide.html"},
		{"README.Md", "README.html"},
		{"no-extension", "no-extension"},
		{"page.html", "page.html"},
	}
	for _, tt := range tests {
		if got := FromMarkdown(tt.in); got != tt.want {
			t.Errorf("FromMarkdown(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
