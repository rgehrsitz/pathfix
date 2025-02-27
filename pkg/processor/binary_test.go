// File: pkg/processor/binary_test.go
package processor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBinaryFileDetection(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "binary-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a text file
	textFile := filepath.Join(tempDir, "text.txt")
	textContent := "This is a plain text file.\nIt has multiple lines.\nNo binary content here."
	if err := os.WriteFile(textFile, []byte(textContent), 0644); err != nil {
		t.Fatalf("Failed to write text file: %v", err)
	}

	// Create a binary file with null bytes
	binaryFile := filepath.Join(tempDir, "binary.bin")
	binaryContent := []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x00, 0x57, 0x6F, 0x72, 0x6C, 0x64} // "Hello\0World"
	if err := os.WriteFile(binaryFile, binaryContent, 0644); err != nil {
		t.Fatalf("Failed to write binary file: %v", err)
	}

	// Create an empty file
	emptyFile := filepath.Join(tempDir, "empty.txt")
	if err := os.WriteFile(emptyFile, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to write empty file: %v", err)
	}

	tests := []struct {
		path     string
		expected bool
	}{
		{textFile, false},
		{binaryFile, true},
		{emptyFile, false},
	}

	for _, test := range tests {
		result := isBinaryFile(test.path)
		if result != test.expected {
			t.Errorf("isBinaryFile(%s) = %v, expected %v", test.path, result, test.expected)
		}
	}
}

func TestNonExistentFile(t *testing.T) {
	// Test with a non-existent file
	result := isBinaryFile("/path/to/nonexistent/file.txt")
	
	// Should return false for non-existent files
	if result {
		t.Errorf("isBinaryFile should return false for non-existent files")
	}
}