// Package scanner provides types and functions for discovering, reading, and parsing
// markdown documents from the filesystem.
package scanner

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Scanner is used to discover and read markdown files from a specified root directory,
// applying ignore patterns and parsing files into Document structs.
type Scanner struct {
	rootPath string
	filter   *IgnoreFilter
}

// NewScanner creates a new Scanner for the given root path and ignore patterns.
// It returns an error if the root path is empty or does not exist.
func NewScanner(rootPath string, ignorePatterns []string) (*Scanner, error) {
	if rootPath == "" {
		return nil, errors.New("root path cannot be empty")
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}

	// Check if path exists
	if _, err := os.Stat(absPath); err != nil {
		return nil, err
	}

	return &Scanner{
		rootPath: absPath,
		filter:   NewIgnoreFilter(ignorePatterns),
	}, nil
}

// Scan walks the configured root path, discovers all markdown files that are not
// ignored, and returns them as a slice of parsed Document structs.
func (s *Scanner) Scan() ([]Document, error) {
	var documents []Document

	err := filepath.WalkDir(s.rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Only process markdown files
		if !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(s.rootPath, path)
		if err != nil {
			return err
		}

		// Check if should ignore
		if s.filter.ShouldIgnore(relPath) {
			return nil
		}

		// Read file
		doc, err := s.readDocument(path, relPath)
		if err != nil {
			// Log error but continue scanning
			return nil
		}

		documents = append(documents, doc)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return documents, nil
}

// readDocument reads and parses a single markdown file from the given path.
func (s *Scanner) readDocument(path, relPath string) (Document, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Document{}, err
	}

	// Get file info
	info, err := os.Stat(path)
	if err != nil {
		return Document{}, err
	}

	// Fix relative path if it's just "."
	if relPath == "." {
		relPath = filepath.Base(path)
	}

	// Extract frontmatter
	metadata, cleanContent := ExtractFrontmatter(content)

	// Create document
	doc := Document{
		ID:           generateDocumentID(relPath),
		Path:         path,
		RelativePath: filepath.ToSlash(relPath), // Normalize to forward slashes
		Content:      cleanContent,
		Metadata:     metadata,
		ModTime:      info.ModTime(),
	}

	// Extract title
	doc.Title = doc.ExtractTitle()

	// Extract sections, links, and code blocks
	doc.Sections = doc.ExtractSections()
	doc.Links = doc.ExtractLinks()
	doc.CodeBlocks = doc.ExtractCodeBlocks()

	return doc, nil
}

// generateDocumentID creates a stable, unique identifier for a document by
// hashing its relative path.
func generateDocumentID(relPath string) string {
	// Use MD5 hash of relative path for consistent IDs
	hash := md5.Sum([]byte(relPath))
	return hex.EncodeToString(hash[:])
}

// ScanSingle reads and parses a single file specified by its path.
func (s *Scanner) ScanSingle(path string) (Document, error) {
	// Get relative path
	relPath, err := filepath.Rel(s.rootPath, path)
	if err != nil {
		return Document{}, err
	}

	return s.readDocument(path, relPath)
}
