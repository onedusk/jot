// Package scanner provides types and functions for discovering, reading, and parsing
// markdown documents from the filesystem.
package scanner

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Scanner is used to discover and read markdown files from a specified root directory,
// applying ignore patterns and parsing files into Document structs.
type Scanner struct {
	rootPath string
	prefix   string // rootPath relative to the directory given to RelativeTo
	filter   *IgnoreFilter
	excluded []os.FileInfo
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

// Exclude prevents the scanner from descending into the given directories,
// regardless of ignore patterns. Use it to keep a build's output directory
// from being read back in as source.
//
// Directories are matched by file identity, so symlinks and other spellings of
// the same directory are excluded too. Directories that do not exist are
// ignored, and the scan root itself is never excluded, so building into an
// input directory still scans it.
func (s *Scanner) Exclude(dirs ...string) error {
	for _, dir := range dirs {
		info, err := os.Stat(dir)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		s.excluded = append(s.excluded, info)
	}
	return nil
}

// RelativeTo makes document paths, and the IDs derived from them, relative to
// dir instead of the scan root. With dir as the project root, scanning docs
// yields docs/guide.md rather than guide.md, and scanning the file
// docs/README.md yields docs/README.md rather than README.md. Ignore patterns
// still match paths relative to the scan root. Call it before Scan.
//
// The scan root must be dir or inside it. As with Exclude, dir is matched by
// file identity, so it may be spelled through a symlink.
func (s *Scanner) RelativeTo(dir string) error {
	dirInfo, err := os.Stat(dir)
	if err != nil {
		return err
	}
	for p := s.rootPath; ; p = filepath.Dir(p) {
		if info, err := os.Stat(p); err == nil && os.SameFile(info, dirInfo) {
			s.prefix, err = filepath.Rel(p, s.rootPath)
			return err
		}
		if filepath.Dir(p) == p {
			return fmt.Errorf("%s is outside %s", s.rootPath, dir)
		}
	}
}

// isExcluded reports whether a directory entry is one of the excluded directories.
func (s *Scanner) isExcluded(d fs.DirEntry) bool {
	if len(s.excluded) == 0 {
		return false
	}
	info, err := d.Info()
	if err != nil {
		return false
	}
	for _, excluded := range s.excluded {
		if os.SameFile(info, excluded) {
			return true
		}
	}
	return false
}

// Scan walks the configured root path, discovers all markdown files that are not
// ignored, and returns them as a slice of parsed Document structs.
func (s *Scanner) Scan() ([]Document, error) {
	var documents []Document

	err := filepath.WalkDir(s.rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories, and do not descend into excluded ones
		if d.IsDir() {
			if path != s.rootPath && s.isExcluded(d) {
				return filepath.SkipDir
			}
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
		doc, err := s.readDocument(path, filepath.Join(s.prefix, relPath))
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
