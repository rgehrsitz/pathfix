// File: main.go
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yourusername/pathfix/pkg/processor"
)

func main() {
	var (
		targetDir      string
		dryRun         bool
		configFilePath string
		verbose        bool
		includeHidden  bool
	)

	// Parse command line arguments
	flag.StringVar(&targetDir, "dir", ".", "Target directory to process")
	flag.BoolVar(&dryRun, "dry-run", false, "Preview changes without modifying files")
	flag.StringVar(&configFilePath, "config", "", "Path to custom configuration file")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&includeHidden, "include-hidden", false, "Process hidden files and directories")
	flag.Parse()

	// Convert to absolute path
	absPath, err := filepath.Abs(targetDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving path %s: %v\n", targetDir, err)
		os.Exit(1)
	}

	// Check if the directory exists
	info, err := os.Stat(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error accessing directory %s: %v\n", absPath, err)
		os.Exit(1)
	}

	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "%s is not a directory\n", absPath)
		os.Exit(1)
	}

	// Create processor with options
	p := processor.NewProcessor(absPath, &processor.Options{
		DryRun:        dryRun,
		ConfigFile:    configFilePath,
		Verbose:       verbose,
		IncludeHidden: includeHidden,
	})

	// Process the directory
	stats, err := p.Process()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing directory: %v\n", err)
		os.Exit(1)
	}

	// Print summary
	fmt.Printf("Processed %d files (%d updated, %d skipped, %d errors)\n",
		stats.Processed, stats.Updated, stats.Skipped, stats.Errors)

	if dryRun {
		fmt.Println("This was a dry run. No files were modified.")
	}
}