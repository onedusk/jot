// Package scanner provides types and functions for discovering, reading, and parsing
// markdown documents from the filesystem.
package scanner

import (
	"path/filepath"
	"strings"
)

// IgnoreFilter provides a mechanism to filter out files and directories based on
// a list of gitignore-style patterns.
type IgnoreFilter struct {
	patterns []string
}

// NewIgnoreFilter creates a new IgnoreFilter with the given set of patterns.
func NewIgnoreFilter(patterns []string) *IgnoreFilter {
	return &IgnoreFilter{
		patterns: patterns,
	}
}

// ShouldIgnore determines if a given file path should be ignored by checking it
// against all of the filter's patterns.
func (f *IgnoreFilter) ShouldIgnore(path string) bool {
	// Normalize path separators
	path = filepath.ToSlash(path)

	for _, pattern := range f.patterns {
		pattern = filepath.ToSlash(pattern)

		// Handle different pattern types
		if matched := f.matchPattern(pattern, path); matched {
			return true
		}
	}

	return false
}

// matchPattern checks if a path matches a single gitignore-style pattern.
// This is a helper function for ShouldIgnore.
func (f *IgnoreFilter) matchPattern(pattern, path string) bool {
	// Exact match
	if pattern == path {
		return true
	}

	// Directory prefix (e.g., "docs/")
	if strings.HasSuffix(pattern, "/") {
		return strings.HasPrefix(path, pattern)
	}

	// Wildcard patterns
	if strings.Contains(pattern, "*") {
		// Convert gitignore pattern to filepath pattern
		if strings.HasPrefix(pattern, "**/") {
			// Match anywhere in path
			suffix := pattern[3:]

			// Special handling for hidden directories pattern
			if suffix == ".*/**" {
				// Check if any part of the path contains a hidden directory
				parts := strings.Split(path, "/")
				for _, part := range parts {
					if strings.HasPrefix(part, ".") && part != "." && part != ".." {
						return true
					}
				}
			}

			return f.matchWildcard(suffix, path) ||
				f.matchInSubpath(suffix, path)
		} else if strings.HasSuffix(pattern, "/**") {
			// Match directory and all contents
			prefix := pattern[:len(pattern)-3]
			return strings.HasPrefix(path, prefix+"/") || path == prefix
		} else {
			// Simple wildcard
			return f.matchWildcard(pattern, path)
		}
	}

	// Check if pattern matches any parent directory
	parts := strings.Split(path, "/")
	for i := range parts {
		subpath := strings.Join(parts[:i+1], "/")
		if pattern == subpath {
			return true
		}
	}

	return false
}

// matchWildcard performs a simple wildcard match.
// This is a helper function for matchPattern.
func (f *IgnoreFilter) matchWildcard(pattern, path string) bool {
	// Simple implementation - just check prefix and suffix
	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
		// *something*
		contains := pattern[1 : len(pattern)-1]
		return strings.Contains(path, contains)
	} else if strings.HasPrefix(pattern, "*") {
		// *suffix
		suffix := pattern[1:]
		return strings.HasSuffix(path, suffix)
	} else if strings.HasSuffix(pattern, "*") {
		// prefix*
		prefix := pattern[:len(pattern)-1]
		return strings.HasPrefix(path, prefix)
	}

	// Use filepath.Match for more complex patterns
	matched, _ := filepath.Match(pattern, filepath.Base(path))
	return matched
}

// matchInSubpath checks if a pattern matches any component of the path.
// This is a helper function for matchPattern.
func (f *IgnoreFilter) matchInSubpath(pattern, path string) bool {
	parts := strings.Split(path, "/")

	// Check if pattern matches any part of the path
	for i := 0; i < len(parts); i++ {
		for j := i; j <= len(parts); j++ {
			subpath := strings.Join(parts[i:j], "/")
			if f.matchWildcard(pattern, subpath) {
				return true
			}
			// Also check without leading/trailing slashes
			if matched, _ := filepath.Match(pattern, subpath); matched {
				return true
			}
		}
	}

	return false
}

// LoadIgnoreFile reads a .jotignore file from the given path and returns a slice
// of patterns.
// TODO: Implement the file reading logic.
func LoadIgnoreFile(path string) ([]string, error) {
	// TODO: Implement reading from file
	// For now, return empty patterns
	return []string{}, nil
}
