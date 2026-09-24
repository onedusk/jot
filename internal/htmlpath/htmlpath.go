// Package htmlpath maps markdown source paths to the paths of their rendered pages.
package htmlpath

import "strings"

// FromMarkdown returns the page path for a markdown path by replacing its
// ".md" extension (in any case) with ".html". Only the extension is replaced,
// so a directory such as "site.mdocs/" is left alone. Paths without a ".md"
// extension are returned unchanged.
func FromMarkdown(path string) string {
	if len(path) >= 3 && strings.EqualFold(path[len(path)-3:], ".md") {
		return path[:len(path)-3] + ".html"
	}
	return path
}
