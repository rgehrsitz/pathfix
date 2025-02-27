// File: pkg/models/models.go
package models

// CommentStyle defines how comments are formatted for a specific file type
type CommentStyle struct {
	LineComment       string // For single line comments (e.g. // for C-style, # for Python)
	BlockCommentStart string // For block comments start (e.g. /* for C-style)
	BlockCommentEnd   string // For block comments end (e.g. */ for C-style)
	Preferred         string // Preferred comment style: "line" or "block"
}

// Config holds the application configuration
type Config struct {
	FileTypes            map[string]CommentStyle // Map of file extension to comment style
	AdditionalIgnores    []string                // Additional file/directory patterns to ignore
	IncludeGitIgnored    bool                    // Whether to process files ignored by .gitignore
	IncludeHidden        bool                    // Whether to process hidden files/directories
	DryRun               bool                    // If true, don't modify files
	CommentPrefix        string                  // Text to prepend before the file path (default: "File: ")
	UpdateExistingPrefix string                  // If not empty, only update comments starting with this prefix
}

// Stats tracks processing statistics
type Stats struct {
	Processed int // Total number of files processed
	Updated   int // Number of files updated
	Skipped   int // Number of files skipped
	Errors    int // Number of files with errors
}