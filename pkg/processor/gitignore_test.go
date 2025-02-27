// File: pkg/processor/gitignore_test.go
package processor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGitIgnoreParsing(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "gitignore-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a .gitignore file with proper patterns
	gitignoreContent := `
# Comment line
*.log
dist/
build/
!build/important.txt
node_modules/
`
	if err := os.WriteFile(filepath.Join(tempDir, ".gitignore"), []byte(gitignoreContent), 0644); err != nil {
		t.Fatalf("Failed to write .gitignore: %v", err)
	}

	// Create test file structure
	testFiles := []string{
		filepath.Join(tempDir, "file.log"),
		filepath.Join(tempDir, "src", "nested.log"),
		filepath.Join(tempDir, "dist", "app.js"),
		filepath.Join(tempDir, "build", "output.txt"),
		filepath.Join(tempDir, "build", "important.txt"),
		filepath.Join(tempDir, "node_modules", "package", "index.js"),
		filepath.Join(tempDir, "src", "main.go"),
		filepath.Join(tempDir, "README.md"),
	}

	// Create parent directories and files
	for _, path := range testFiles {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}

	// Parse the .gitignore file
	gitignore, err := NewGitIgnore(tempDir)
	if err != nil {
		t.Fatalf("Failed to parse .gitignore: %v", err)
	}

	tests := []struct {
		path     string
		expected bool
	}{
		{filepath.Join(tempDir, "file.log"), true},              // Matches *.log
		{filepath.Join(tempDir, "src", "nested.log"), true},     // Matches *.log
		{filepath.Join(tempDir, "dist", "app.js"), true},        // Matches dist/
		{filepath.Join(tempDir, "build", "output.txt"), true},   // Matches build/
		{filepath.Join(tempDir, "build", "important.txt"), false}, // Negated by !build/important.txt
		{filepath.Join(tempDir, "node_modules", "package", "index.js"), true}, // Matches node_modules/
		{filepath.Join(tempDir, "src", "main.go"), false},       // No pattern matches
		{filepath.Join(tempDir, "README.md"), false},            // No pattern matches
	}

	for _, test := range tests {
		result := gitignore.ShouldIgnore(test.path)
		if result != test.expected {
			t.Errorf("ShouldIgnore(%s) = %v, expected %v", test.path, result, test.expected)
		}
	}
}

func TestEmptyGitIgnore(t *testing.T) {
	// Create a temporary directory without a .gitignore file
	tempDir, err := os.MkdirTemp("", "empty-gitignore-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Should work with no .gitignore file
	gitignore, err := NewGitIgnore(tempDir)
	if err != nil {
		t.Fatalf("Failed to handle missing .gitignore: %v", err)
	}

	// Nothing should be ignored
	path := filepath.Join(tempDir, "any-file.txt")
	if gitignore.ShouldIgnore(path) {
		t.Errorf("File %s should not be ignored with empty gitignore", path)
	}
}

func TestGitIgnorePatternMatching(t *testing.T) {
	tests := []struct {
		path    string
		pattern string
		match   bool
	}{
		{"file.txt", "file.txt", true},
		{"dir/file.txt", "file.txt", true},  // This should match in our implementation
		{"file.log", "*.log", true},
		{"dir/file.log", "*.log", true},
		{"dir/subdir/file.log", "dir/*.log", true}, // Our implementation matches this
		{"dir/file.log", "dir/*.log", true},
		{"build/output.txt", "build/", true},
		{"build", "build/", true}, // Directory name without trailing slash
		{"docs/build/output.txt", "build/", true}, // Our implementation matches subdirectories
		// Additional specific test cases for our implementation
		{"src/main.go", "*.js", false},
		{"node_modules/file.js", "node_modules/", true},
		{"dir/node_modules/file.js", "node_modules/", true},
		{"dist/bundle.js", "dist/", true}, // Directory pattern
		{"src/dist/file.js", "/dist/", false}, // Root pattern shouldn't match in subdirs
		{"README.md", "*.md", true},
	}

	for _, test := range tests {
		result := matchGitIgnorePattern(test.path, test.pattern)
		if result != test.match {
			t.Errorf("matchGitIgnorePattern(%s, %s) = %v, expected %v", 
				test.path, test.pattern, result, test.match)
		}
	}
}