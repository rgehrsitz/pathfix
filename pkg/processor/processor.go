// File: pkg/processor/processor.go
package processor

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/yourusername/pathfix/pkg/models"
)

// Options represents processor options
type Options struct {
	DryRun        bool
	ConfigFile    string
	Verbose       bool
	IncludeHidden bool
}

// Processor handles the file processing logic
type Processor struct {
	rootDir    string
	options    *Options
	config     *models.Config
	fileTypes  map[string]models.CommentStyle
	statistics models.Stats
}

// NewProcessor creates a new processor
func NewProcessor(rootDir string, options *Options) *Processor {
	p := &Processor{
		rootDir: rootDir,
		options: options,
	}

	// Initialize default file types
	p.initializeFileTypes()

	// Load config file if specified
	var config *models.Config
	var err error

	if options.ConfigFile != "" {
		config, err = LoadConfig(options.ConfigFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Error loading config file: %v\n", err)
			config = &models.Config{
				CommentPrefix: "File: ",
				DryRun:        options.DryRun,
				IncludeHidden: options.IncludeHidden,
			}
		}
	} else {
		// Default configuration
		config = &models.Config{
			CommentPrefix: "File: ",
			DryRun:        options.DryRun,
			IncludeHidden: options.IncludeHidden,
		}
	}

	// Merge with default file types
	p.config = MergeConfig(config, p.fileTypes)
	
	// Override with command line options
	p.config.DryRun = options.DryRun
	p.config.IncludeHidden = options.IncludeHidden

	return p
}

// initializeFileTypes sets up supported file types and their comment styles
func (p *Processor) initializeFileTypes() {
	p.fileTypes = map[string]models.CommentStyle{
		// C-style languages
		".cs":   {LineComment: "//", BlockCommentStart: "/*", BlockCommentEnd: "*/", Preferred: "line"},
		".go":   {LineComment: "//", BlockCommentStart: "/*", BlockCommentEnd: "*/", Preferred: "line"},
		".c":    {LineComment: "//", BlockCommentStart: "/*", BlockCommentEnd: "*/", Preferred: "line"},
		".cpp":  {LineComment: "//", BlockCommentStart: "/*", BlockCommentEnd: "*/", Preferred: "line"},
		".h":    {LineComment: "//", BlockCommentStart: "/*", BlockCommentEnd: "*/", Preferred: "line"},
		".hpp":  {LineComment: "//", BlockCommentStart: "/*", BlockCommentEnd: "*/", Preferred: "line"},
		".java": {LineComment: "//", BlockCommentStart: "/*", BlockCommentEnd: "*/", Preferred: "line"},
		".js":   {LineComment: "//", BlockCommentStart: "/*", BlockCommentEnd: "*/", Preferred: "line"},
		".ts":   {LineComment: "//", BlockCommentStart: "/*", BlockCommentEnd: "*/", Preferred: "line"},
		".jsx":  {LineComment: "//", BlockCommentStart: "/*", BlockCommentEnd: "*/", Preferred: "line"},
		".tsx":  {LineComment: "//", BlockCommentStart: "/*", BlockCommentEnd: "*/", Preferred: "line"},

		// Shell/script languages
		".sh":  {LineComment: "#", Preferred: "line"},
		".ps1": {LineComment: "#", Preferred: "line"},
		".py":  {LineComment: "#", BlockCommentStart: "'''", BlockCommentEnd: "'''", Preferred: "line"},
		".rb":  {LineComment: "#", BlockCommentStart: "=begin", BlockCommentEnd: "=end", Preferred: "line"},

		// Web languages
		".html": {LineComment: "", BlockCommentStart: "<!--", BlockCommentEnd: "-->", Preferred: "block"},
		".xml":  {LineComment: "", BlockCommentStart: "<!--", BlockCommentEnd: "-->", Preferred: "block"},
		".css":  {LineComment: "", BlockCommentStart: "/*", BlockCommentEnd: "*/", Preferred: "block"},

		// Config files
		".yaml": {LineComment: "#", Preferred: "line"},
		".yml":  {LineComment: "#", Preferred: "line"},
		".toml": {LineComment: "#", Preferred: "line"},
		".ini":  {LineComment: ";", Preferred: "line"},
		".conf": {LineComment: "#", Preferred: "line"},

		// Other languages
		".rs":   {LineComment: "//", BlockCommentStart: "/*", BlockCommentEnd: "*/", Preferred: "line"},
		".swift": {LineComment: "//", BlockCommentStart: "/*", BlockCommentEnd: "*/", Preferred: "line"},
		".kt":    {LineComment: "//", BlockCommentStart: "/*", BlockCommentEnd: "*/", Preferred: "line"},
		".lua":   {LineComment: "--", BlockCommentStart: "--[[", BlockCommentEnd: "--]]", Preferred: "line"},
		".pl":    {LineComment: "#", Preferred: "line"},
		".php":   {LineComment: "//", BlockCommentStart: "/*", BlockCommentEnd: "*/", Preferred: "line"},
	}
}

// Process walks through the directory and processes files
func (p *Processor) Process() (models.Stats, error) {
	// Load gitignore if it exists
	gitignore, err := NewGitIgnore(p.rootDir)
	if err != nil {
		return p.statistics, fmt.Errorf("error loading .gitignore: %w", err)
	}

	err = filepath.WalkDir(p.rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			// Skip hidden directories unless explicitly included
			if !p.options.IncludeHidden && isHidden(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip hidden files unless explicitly included
		if !p.options.IncludeHidden && isHidden(d.Name()) {
			p.statistics.Skipped++
			return nil
		}

		// Skip files ignored by gitignore unless explicitly included
		if !p.config.IncludeGitIgnored && gitignore.ShouldIgnore(path) {
			if p.options.Verbose {
				fmt.Printf("Skipping gitignored file: %s\n", path)
			}
			p.statistics.Skipped++
			return nil
		}

		// Get relative path from root directory
		relPath, err := filepath.Rel(p.rootDir, path)
		if err != nil {
			if p.options.Verbose {
				fmt.Fprintf(os.Stderr, "Error getting relative path for %s: %v\n", path, err)
			}
			p.statistics.Errors++
			return nil
		}

		// Skip files based on extension
		ext := strings.ToLower(filepath.Ext(path))
		if _, ok := p.fileTypes[ext]; !ok {
			if p.options.Verbose {
				fmt.Printf("Skipping unsupported file type: %s\n", path)
			}
			p.statistics.Skipped++
			return nil
		}

		// Process the file
		p.statistics.Processed++
		updated, err := p.processFile(path, relPath)
		if err != nil {
			if p.options.Verbose {
				fmt.Fprintf(os.Stderr, "Error processing file %s: %v\n", path, err)
			}
			p.statistics.Errors++
		} else if updated {
			p.statistics.Updated++
		} else {
			p.statistics.Skipped++
		}

		return nil
	})

	return p.statistics, err
}

// processFile adds or updates the file header comment
func (p *Processor) processFile(filePath, relPath string) (bool, error) {
	// Check if file is binary
	if isBinaryFile(filePath) {
		if p.options.Verbose {
			fmt.Printf("Skipping binary file: %s\n", filePath)
		}
		return false, nil
	}

	// Normalize path separators for comments
	relPath = filepath.ToSlash(relPath)

	// Get file extension and comment style
	ext := strings.ToLower(filepath.Ext(filePath))
	commentStyle, ok := p.fileTypes[ext]
	if !ok {
		return false, fmt.Errorf("unsupported file type: %s", ext)
	}

	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	var newContent []byte
	var updated bool

	// Format the comment
	var commentText string
	commentPrefix := p.config.CommentPrefix
	filePathComment := fmt.Sprintf("%s%s", commentPrefix, relPath)

	if commentStyle.Preferred == "line" && commentStyle.LineComment != "" {
		commentText = fmt.Sprintf("%s %s\n", commentStyle.LineComment, filePathComment)
	} else if commentStyle.BlockCommentStart != "" && commentStyle.BlockCommentEnd != "" {
		commentText = fmt.Sprintf("%s %s %s\n", commentStyle.BlockCommentStart, filePathComment, commentStyle.BlockCommentEnd)
	} else {
		return false, fmt.Errorf("no valid comment style for file type: %s", ext)
	}

	// Check if file already has a comment header and update it if necessary
	scanner := bufio.NewScanner(bytes.NewReader(content))
	if scanner.Scan() {
		firstLine := scanner.Text()
		
		// Check if there's already a file path comment
		if strings.Contains(firstLine, commentStyle.LineComment) && strings.Contains(firstLine, commentPrefix) ||
			(strings.Contains(firstLine, commentStyle.BlockCommentStart) && strings.Contains(firstLine, commentPrefix)) {
			
			// Replace the existing comment
			newContent = bytes.Replace(content, []byte(firstLine+"\n"), []byte(commentText), 1)
			updated = true
		} else {
			// Add comment to the beginning
			newContent = append([]byte(commentText), content...)
			updated = true
		}
	} else {
		// Empty file, just add the comment
		newContent = []byte(commentText)
		updated = true
	}

	// Write back if updated
	if updated && !p.options.DryRun {
		err = os.WriteFile(filePath, newContent, 0644)
		if err != nil {
			return false, err
		}
	}

	if p.options.Verbose {
		if updated {
			if p.options.DryRun {
				fmt.Printf("Would update: %s\n", filePath)
			} else {
				fmt.Printf("Updated: %s\n", filePath)
			}
		} else {
			fmt.Printf("No changes needed: %s\n", filePath)
		}
	}

	return updated, nil
}

// isBinaryFile checks if a file is likely to be binary
func isBinaryFile(path string) bool {
	// Open file
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	// Read first 512 bytes
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return false
	}
	buf = buf[:n]

	// Check for null bytes which would indicate binary
	return bytes.IndexByte(buf, 0) != -1
}

// isHidden checks if a file or directory is hidden
func isHidden(name string) bool {
	return strings.HasPrefix(name, ".")
}