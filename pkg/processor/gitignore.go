// File: pkg/processor/gitignore.go
package processor

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// GitIgnore holds patterns from a .gitignore file
type GitIgnore struct {
	patterns []string
	rootDir  string
}

// NewGitIgnore creates a new GitIgnore processor
func NewGitIgnore(rootDir string) (*GitIgnore, error) {
	gitignorePath := filepath.Join(rootDir, ".gitignore")
	gi := &GitIgnore{
		rootDir: rootDir,
	}

	file, err := os.Open(gitignorePath)
	if err != nil {
		// Return an empty gitignore if file doesn't exist
		if os.IsNotExist(err) {
			return gi, nil
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		gi.patterns = append(gi.patterns, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return gi, nil
}

// ShouldIgnore checks if a file should be ignored based on .gitignore patterns
func (gi *GitIgnore) ShouldIgnore(path string) bool {
	// Get relative path from root directory
	relPath, err := filepath.Rel(gi.rootDir, path)
	if err != nil {
		return false
	}

	// Normalize path separators to forward slashes
	relPath = filepath.ToSlash(relPath)

	// First pass: find if the file is ignored by any pattern
	isIgnored := false
	for _, pattern := range gi.patterns {
		// Skip negated patterns in the first pass
		if strings.HasPrefix(pattern, "!") {
			continue
		}

		// Check if file matches any non-negated pattern
		if matchGitIgnorePattern(relPath, pattern) {
			isIgnored = true
			break
		}
	}

	// If the file is not ignored by any pattern, no need to check negations
	if !isIgnored {
		return false
	}

	// Second pass: check if the file is explicitly included by a negated pattern
	for _, pattern := range gi.patterns {
		// Only process negated patterns
		if !strings.HasPrefix(pattern, "!") {
			continue
		}

		// Remove the leading ! for matching
		negatedPattern := pattern[1:]
		if matchGitIgnorePattern(relPath, negatedPattern) {
			// This file is explicitly included
			return false
		}
	}

	// File is ignored by at least one pattern and not explicitly included
	return true
}

// matchGitIgnorePattern checks if a path matches a gitignore pattern
func matchGitIgnorePattern(path, pattern string) bool {
	// Handle root pattern (starting with /)
	if strings.HasPrefix(pattern, "/") {
		// Remove the leading / for matching against relative paths
		pattern = pattern[1:]
		// For root patterns, the path must match from the root
		return path == pattern || strings.HasPrefix(path, pattern+"/")
	}

	// Handle directory pattern (ending with /)
	if strings.HasSuffix(pattern, "/") {
		// Remove the trailing / for exact directory name matching
		dirPattern := pattern[:len(pattern)-1]
		
		// Match exact directory name or path starting with the directory
		return path == dirPattern || 
			   strings.HasPrefix(path, pattern) || 
			   strings.HasSuffix(path, "/"+dirPattern) ||
			   strings.Contains(path, "/"+dirPattern+"/")
	}

	// Handle negated pattern (patterns beginning with !)
	if strings.HasPrefix(pattern, "!") {
		// This is handled at a higher level in ShouldIgnore
		return false
	}

	// Handle wildcard patterns
	if strings.Contains(pattern, "*") {
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			// Simple case: pattern is something like "*.log" or "dir/*.js"
			prefix, suffix := parts[0], parts[1]
			
			// For "*.ext" pattern
			if prefix == "" && suffix != "" {
				return strings.HasSuffix(path, suffix)
			}
			
			// For "prefix*" pattern
			if prefix != "" && suffix == "" {
				return strings.HasPrefix(path, prefix)
			}

			// For "prefix*suffix" pattern
			if prefix != "" && suffix != "" {
				// If the pattern starts with dir/, make sure we're matching the whole directory path
				if strings.Contains(prefix, "/") {
					return strings.HasPrefix(path, prefix) && strings.HasSuffix(path, suffix)
				} else {
					// Otherwise, match the pattern anywhere in the path components
					filename := filepath.Base(path)
					return strings.HasPrefix(filename, prefix) && strings.HasSuffix(filename, suffix)
				}
			}
		}
	}

	// Exact match
	if strings.Contains(pattern, "/") {
		// If pattern has a slash, it's a specific path
		return path == pattern || strings.HasPrefix(path, pattern+"/")
	} else {
		// Otherwise, look for the pattern in any part of the path
		pathParts := strings.Split(path, "/")
		for _, part := range pathParts {
			if part == pattern {
				return true
			}
		}
	}

	return false
}