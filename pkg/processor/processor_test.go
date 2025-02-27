// File: pkg/processor/processor_test.go
package processor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCommentStyles(t *testing.T) {
	p := &Processor{}
	p.initializeFileTypes()

	tests := []struct {
		extension          string
		expectedStyle      string
		expectedLinePrefix string
		expectedPresent    bool
	}{
		{".go", "line", "//", true},
		{".cs", "line", "//", true},
		{".js", "line", "//", true},
		{".py", "line", "#", true},
		{".html", "block", "", true},
		{".yml", "line", "#", true},
		{".unknown", "", "", false},
	}

	for _, test := range tests {
		style, ok := p.fileTypes[test.extension]
		
		if ok != test.expectedPresent {
			t.Errorf("Expected presence of %s: %v, got: %v", test.extension, test.expectedPresent, ok)
			continue
		}
		
		if !ok {
			continue
		}
		
		if style.Preferred != test.expectedStyle {
			t.Errorf("Expected style for %s: %s, got: %s", test.extension, test.expectedStyle, style.Preferred)
		}
		
		if style.LineComment != test.expectedLinePrefix {
			t.Errorf("Expected line prefix for %s: %s, got: %s", 
				test.extension, test.expectedLinePrefix, style.LineComment)
		}
	}
}

func TestFileProcessing(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "process-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test Go file
	goFilePath := filepath.Join(tempDir, "test.go")
	goContent := `package main

func main() {
	println("Hello World")
}
`
	if err := os.WriteFile(goFilePath, []byte(goContent), 0644); err != nil {
		t.Fatalf("Failed to write Go file: %v", err)
	}

	// Create a Go file that already has a comment
	goFileWithCommentPath := filepath.Join(tempDir, "with_comment.go")
	goWithCommentContent := `// File: old/path/with_comment.go
package main

func main() {
	println("Hello World with comment")
}
`
	if err := os.WriteFile(goFileWithCommentPath, []byte(goWithCommentContent), 0644); err != nil {
		t.Fatalf("Failed to write Go file with comment: %v", err)
	}

	// Create processor
	processor := NewProcessor(tempDir, &Options{
		Verbose: true,
	})

	// Test processing for new file
	relPath := "test.go"
	updated, err := processor.processFile(goFilePath, relPath)
	if err != nil {
		t.Fatalf("processFile failed: %v", err)
	}
	if !updated {
		t.Errorf("Expected file to be updated, but it wasn't")
	}

	// Read the updated file
	content, err := os.ReadFile(goFilePath)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}
	expectedComment := "// File: test.go\n"
	if string(content[:len(expectedComment)]) != expectedComment {
		t.Errorf("Expected comment not found in processed file")
	}

	// Test processing for file with existing comment
	relPath = "with_comment.go"
	updated, err = processor.processFile(goFileWithCommentPath, relPath)
	if err != nil {
		t.Fatalf("processFile failed: %v", err)
	}
	if !updated {
		t.Errorf("Expected file with comment to be updated, but it wasn't")
	}

	// Read the updated file
	content, err = os.ReadFile(goFileWithCommentPath)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}
	expectedComment = "// File: with_comment.go\n"
	if string(content[:len(expectedComment)]) != expectedComment {
		t.Errorf("Expected comment update not found in processed file, got: %s", 
			string(content[:len(expectedComment)]))
	}
}

func TestConfigLoading(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test config file
	configPath := filepath.Join(tempDir, "config.json")
	configContent := `{
		"CommentPrefix": "MyPrefix: ",
		"IncludeGitIgnored": true,
		"FileTypes": {
			".custom": {
				"LineComment": "##",
				"Preferred": "line"
			}
		}
	}`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Load the config
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify config values
	if config.CommentPrefix != "MyPrefix: " {
		t.Errorf("Expected CommentPrefix: MyPrefix: , got: %s", config.CommentPrefix)
	}
	if !config.IncludeGitIgnored {
		t.Errorf("Expected IncludeGitIgnored: true, got: false")
	}
	if customStyle, ok := config.FileTypes[".custom"]; !ok {
		t.Errorf("Custom file type .custom not found in config")
	} else if customStyle.LineComment != "##" {
		t.Errorf("Expected LineComment: ##, got: %s", customStyle.LineComment)
	}
}

func TestHiddenFileDetection(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{".hidden", true},
		{".git", true},
		{".gitignore", true},
		{"normal.txt", false},
		{"file", false},
		{"..", true}, // Parent directory is considered hidden
	}

	for _, test := range tests {
		result := isHidden(test.filename)
		if result != test.expected {
			t.Errorf("isHidden(%s) = %v, expected %v", test.filename, result, test.expected)
		}
	}
}