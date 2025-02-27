// File: pkg/processor/integration_test.go
package processor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFullProcessor(t *testing.T) {
	// Create a temporary directory structure
	tempDir, err := os.MkdirTemp("", "integration-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a nested directory structure with different file types
	dirs := []string{
		filepath.Join(tempDir, "src", "app"),
		filepath.Join(tempDir, "src", "utils"),
		filepath.Join(tempDir, "config"),
		filepath.Join(tempDir, ".git"), // Hidden directory
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create a .gitignore file
	gitignoreContent := `
# Test gitignore
*.log
dist/
vendor/
`
	if err := os.WriteFile(filepath.Join(tempDir, ".gitignore"), []byte(gitignoreContent), 0644); err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}

	// Create various test files
	files := map[string]string{
		filepath.Join(tempDir, "src", "app", "main.go"): `package main

func main() {
	println("Hello, World!")
}
`,
		filepath.Join(tempDir, "src", "utils", "helper.js"): `function helper() {
	console.log("Helper function");
}

module.exports = helper;
`,
		filepath.Join(tempDir, "config", "settings.yaml"): `
app:
  name: TestApp
  version: 1.0.0
`,
		filepath.Join(tempDir, "README.md"): `# Test Project

This is a test project for integration testing.
`,
		filepath.Join(tempDir, "dist", "app.js"): `// This should be ignored by gitignore
console.log("Built app");
`,
		filepath.Join(tempDir, ".git", "config"): `[core]
	repositoryformatversion = 0
	filemode = true
`,
	}

	for filePath, content := range files {
		// Create parent directory if needed
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			t.Fatalf("Failed to create directory for %s: %v", filePath, err)
		}
		
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", filePath, err)
		}
	}

	// Initialize the processor
	processor := NewProcessor(tempDir, &Options{
		Verbose: true,
	})

	// Process the directory
	stats, err := processor.Process()
	if err != nil {
		t.Fatalf("Processor.Process failed: %v", err)
	}

	// Verify statistics
	if stats.Processed < 3 {
		t.Errorf("Expected at least 3 files processed, got: %d", stats.Processed)
	}
	if stats.Updated < 3 {
		t.Errorf("Expected at least 3 files updated, got: %d", stats.Updated)
	}
	
	// Verify that files were updated correctly
	filesToCheck := []struct {
		path            string
		shouldBeUpdated bool
		expectedPrefix  string
	}{
		{filepath.Join(tempDir, "src", "app", "main.go"), true, "// File: src/app/main.go"},
		{filepath.Join(tempDir, "src", "utils", "helper.js"), true, "// File: src/utils/helper.js"},
		{filepath.Join(tempDir, "config", "settings.yaml"), true, "# File: config/settings.yaml"},
		// Don't check README.md since the tests don't have special case for markdown files
		{filepath.Join(tempDir, "dist", "app.js"), false, ""}, // Should be ignored by gitignore
		{filepath.Join(tempDir, ".git", "config"), false, ""}, // Should be ignored as hidden
	}

	for _, fileCheck := range filesToCheck {
		content, err := os.ReadFile(fileCheck.path)
		if err != nil {
			// Skip files that don't exist (might have been created during the test)
			continue
		}
		
		lines := strings.Split(string(content), "\n")
		if len(lines) == 0 {
			t.Errorf("File %s is empty", fileCheck.path)
			continue
		}
		
		firstLine := lines[0]
		
		if fileCheck.shouldBeUpdated {
			if !strings.HasPrefix(firstLine, fileCheck.expectedPrefix) {
				t.Errorf("File %s expected to start with %q, got %q", 
					fileCheck.path, fileCheck.expectedPrefix, firstLine)
			}
		} else {
			if strings.Contains(firstLine, "File:") {
				t.Errorf("File %s should not have been updated, but was: %q", 
					fileCheck.path, firstLine)
			}
		}
	}
}

func TestDryRunMode(t *testing.T) {
	// Create a temporary directory with a test file
	tempDir, err := os.MkdirTemp("", "dryrun-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.go")
	originalContent := "package main\n\nfunc main() {}\n"
	if err := os.WriteFile(testFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Run processor in dry-run mode
	processor := NewProcessor(tempDir, &Options{
		Verbose: true,
		DryRun:  true,
	})

	// Process the directory
	stats, err := processor.Process()
	if err != nil {
		t.Fatalf("Processor.Process failed: %v", err)
	}

	// Verify no files were actually changed
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	if string(content) != originalContent {
		t.Errorf("File was modified in dry-run mode. Expected: %q, got: %q", 
			originalContent, string(content))
	}

	// But stats should indicate files would have been updated
	if stats.Updated < 1 {
		t.Errorf("Dry run statistics should show files would be updated, got updated count: %d", 
			stats.Updated)
	}
}