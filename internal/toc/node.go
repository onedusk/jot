// Package toc provides functionality for building a Table of Contents (TOC)
// from a collection of documents. It organizes documents into a hierarchical
// structure based on their file paths.
package toc

import "time"

// TOCNode represents a single entry in the table of contents. It can be either
// a directory (a node with children but no path) or a document (a leaf node with a path).
type TOCNode struct {
	ID       string     // A unique, URL-friendly identifier for the node.
	Title    string     // The display title, derived from the file/directory name or document title.
	Path     string     // The relative file path for document nodes; empty for directory nodes.
	Weight   int        // An optional weight for custom sorting of sibling nodes.
	Children []*TOCNode // Child nodes, representing files and subdirectories.

	// Enhanced metadata for searchability and richer display.
	Metadata NodeMetadata
}

// NodeMetadata contains supplementary information about a TOC node, primarily
// for document nodes. This data is used for features like search and display enhancements.
type NodeMetadata struct {
	Modified    time.Time // The last modification time of the source file.
	Size        int64     // The file size in bytes.
	WordCount   int       // The approximate word count of the document.
	ReadTime    string    // An estimated reading time (e.g., "5min").
	Tags        []string  // A list of tags or categories from frontmatter.
	Summary     string    // A brief summary of the document content.
	Keywords    []string  // A list of automatically extracted keywords.
	ContentHash string    // A hash of the file content for change detection.
}

// AddChild appends a new child node to the current node's list of children.
func (n *TOCNode) AddChild(child *TOCNode) {
	n.Children = append(n.Children, child)
}

// FindChildByTitle searches the immediate children of the node for one with a
// matching title and returns it. If no match is found, it returns nil.
func (n *TOCNode) FindChildByTitle(title string) *TOCNode {
	for _, child := range n.Children {
		if child.Title == title {
			return child
		}
	}
	return nil
}

// IsLeaf returns true if the node represents a document (i.e., it has a non-empty path).
func (n *TOCNode) IsLeaf() bool {
	return n.Path != ""
}

// SortChildren sorts the node's children based on their weight, and then alphabetically
// by title for nodes with the same weight.
// TODO: Implement the sorting logic.
func (n *TOCNode) SortChildren() {
	// TODO: Implement sorting logic
	// For now, children remain in insertion order
}
