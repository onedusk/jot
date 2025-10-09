// Package toc provides functionality for building a Table of Contents (TOC)
// from a collection of documents. It organizes documents into a hierarchical
// structure based on their file paths.
package toc

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// TableOfContents represents the entire hierarchical structure of the documentation,
// including the root node and a fast-lookup index of all nodes.
type TableOfContents struct {
	Version string
	Root    *TOCNode
	Index   map[string]*TOCNode // Fast lookup by ID for all nodes in the tree.
}

// ToXML serializes the TableOfContents into an XML format, including rich
// metadata for consumption by other tools or systems.
func (t *TableOfContents) ToXML() string {
	var builder strings.Builder

	// XML header with metadata
	builder.WriteString(`<toc version="`)
	builder.WriteString(t.Version)
	builder.WriteString(`" llm-optimized="true">`)
	builder.WriteString("\n")

	// Add metadata section
	t.writeMetadata(&builder)

	// Add main sections
	builder.WriteString("  <sections>\n")
	for _, child := range t.Root.Children {
		t.nodeToXML(&builder, child, 2)
	}
	builder.WriteString("  </sections>\n")

	builder.WriteString("</toc>")

	return builder.String()
}

// writeMetadata is a helper function to write the <metadata> section of the XML output.
func (t *TableOfContents) writeMetadata(builder *strings.Builder) {
	totalDocs := t.countDocuments(t.Root)

	builder.WriteString("  <metadata>\n")
	builder.WriteString("    <totalDocs>")
	builder.WriteString(fmt.Sprintf("%d", totalDocs))
	builder.WriteString("</totalDocs>\n")
	builder.WriteString("  </metadata>\n")
}

// countDocuments recursively counts the number of document (leaf) nodes in the tree.
func (t *TableOfContents) countDocuments(node *TOCNode) int {
	count := 0
	if node.IsLeaf() {
		count = 1
	}
	for _, child := range node.Children {
		count += t.countDocuments(child)
	}
	return count
}

// nodeToXML is a recursive helper function that serializes a TOCNode and its
// children to XML.
func (t *TableOfContents) nodeToXML(builder *strings.Builder, node *TOCNode, depth int) {
	indent := strings.Repeat("  ", depth)

	if node.IsLeaf() {
		// Chapter node (has a path) with enhanced metadata
		builder.WriteString(indent)
		builder.WriteString(`<chapter id="`)
		builder.WriteString(node.ID)
		builder.WriteString(`" path="`)
		builder.WriteString(node.Path)
		builder.WriteString(`"`)

		// Add metadata attributes
		if !node.Metadata.Modified.IsZero() {
			builder.WriteString(` modified="`)
			builder.WriteString(node.Metadata.Modified.Format("2006-01-02T15:04:05Z"))
			builder.WriteString(`"`)
		}
		if node.Metadata.Size > 0 {
			builder.WriteString(` size="`)
			builder.WriteString(fmt.Sprintf("%d", node.Metadata.Size))
			builder.WriteString(`"`)
		}
		if node.Metadata.WordCount > 0 {
			builder.WriteString(` words="`)
			builder.WriteString(fmt.Sprintf("%d", node.Metadata.WordCount))
			builder.WriteString(`"`)
		}
		if node.Metadata.ReadTime != "" {
			builder.WriteString(` readTime="`)
			builder.WriteString(node.Metadata.ReadTime)
			builder.WriteString(`"`)
		}
		if node.Metadata.ContentHash != "" {
			builder.WriteString(` hash="`)
			builder.WriteString(node.Metadata.ContentHash)
			builder.WriteString(`"`)
		}
		if len(node.Metadata.Tags) > 0 {
			builder.WriteString(` tags="`)
			builder.WriteString(escapeXML(strings.Join(node.Metadata.Tags, ",")))
			builder.WriteString(`"`)
		}

		builder.WriteString(">\n")

		builder.WriteString(indent + "  ")
		builder.WriteString("<title>")
		builder.WriteString(escapeXML(node.Title))
		builder.WriteString("</title>\n")

		// Add summary if available
		if node.Metadata.Summary != "" {
			builder.WriteString(indent + "  ")
			builder.WriteString("<summary>")
			builder.WriteString(escapeXML(node.Metadata.Summary))
			builder.WriteString("</summary>\n")
		}

		// Add keywords if available
		if len(node.Metadata.Keywords) > 0 {
			builder.WriteString(indent + "  ")
			builder.WriteString("<keywords>")
			builder.WriteString(escapeXML(strings.Join(node.Metadata.Keywords, ", ")))
			builder.WriteString("</keywords>\n")
		}

		builder.WriteString(indent)
		builder.WriteString("</chapter>\n")
	} else {
		// Section node (directory)
		builder.WriteString(indent)
		builder.WriteString(`<section id="`)
		builder.WriteString(node.ID)
		builder.WriteString("\">\n")

		builder.WriteString(indent + "  ")
		builder.WriteString("<title>")
		builder.WriteString(escapeXML(node.Title))
		builder.WriteString("</title>\n")

		// Process children
		for _, child := range node.Children {
			t.nodeToXML(builder, child, depth+1)
		}

		builder.WriteString(indent)
		builder.WriteString("</section>\n")
	}
}

// GetNodeByID retrieves a specific TOCNode from the TOC's index using its unique ID.
// It returns nil if no node with that ID is found.
func (t *TableOfContents) GetNodeByID(id string) *TOCNode {
	if t.Index == nil {
		t.buildIndex()
	}
	return t.Index[id]
}

// buildIndex creates a map of all nodes in the TOC, indexed by their ID, to allow
// for fast lookups.
func (t *TableOfContents) buildIndex() {
	t.Index = make(map[string]*TOCNode)
	t.indexNode(t.Root)
}

// indexNode is a recursive helper function that populates the TOC's node index.
func (t *TableOfContents) indexNode(node *TOCNode) {
	t.Index[node.ID] = node
	for _, child := range node.Children {
		t.indexNode(child)
	}
}

// escapeXML escapes characters that have special meaning in XML.
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// MarshalXML provides a custom XML marshaling implementation for the TableOfContents.
// This allows it to be easily encoded into XML format.
func (t *TableOfContents) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// For now, just use the string representation
	return e.EncodeElement(t.ToXML(), start)
}
