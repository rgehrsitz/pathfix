# PathFix

PathFix is a Go CLI application that adds or updates file header comments with their relative paths. For instance, it adds comments like `// File: src/main.go` to the beginning of your source files.

## Features

- Adds or updates the first line of files with a comment containing the file's relative path
- Supports multiple comment styles based on file extensions
- Skips binary files automatically
- Respects .gitignore files
- Configurable via JSON configuration files
- Cross-platform (works on Linux, macOS, Windows)
- Dry-run mode to preview changes without modifying files

## Installation

### From Source

```bash
go install github.com/yourusername/pathfix@latest
```

Or clone and build:

```bash
git clone https://github.com/yourusername/pathfix.git
cd pathfix
go build
```

## Usage

Basic usage:

```bash
pathfix --dir /path/to/your/project
```

With options:

```bash
pathfix --dir /path/to/your/project --dry-run --verbose --config /path/to/config.json
```

### Available Options

- `--dir`: Target directory to process (default: current directory)
- `--dry-run`: Preview changes without modifying files
- `--config`: Path to custom configuration file
- `--verbose`: Enable verbose output
- `--include-hidden`: Process hidden files and directories

## Configuration

PathFix can be configured via a JSON file. Here's an example:

```json
{
  "CommentPrefix": "File: ",
  "IncludeGitIgnored": false,
  "IncludeHidden": false,
  "AdditionalIgnores": [
    "vendor/",
    "node_modules/"
  ],
  "FileTypes": {
    ".rs": {
      "LineComment": "//",
      "BlockCommentStart": "/*",
      "BlockCommentEnd": "*/",
      "Preferred": "line"
    },
    ".md": {
      "LineComment": "",
      "BlockCommentStart": "<!--",
      "BlockCommentEnd": "-->",
      "Preferred": "block"
    }
  }
}
```

### Configuration Options

- `CommentPrefix`: Text to prepend before the file path (default: "File: ")
- `IncludeGitIgnored`: Whether to process files ignored by .gitignore
- `IncludeHidden`: Whether to process hidden files/directories
- `AdditionalIgnores`: Additional file/directory patterns to ignore
- `FileTypes`: Map of file extensions to comment styles

### Supported Languages

PathFix supports many languages and file types, including:

- C# (.cs)
- Go (.go)
- C/C++ (.c, .cpp, .h, .hpp)
- Java (.java)
- JavaScript/TypeScript (.js, .ts, .jsx, .tsx)
- Shell scripts (.sh, .ps1)
- Python (.py)
- Ruby (.rb)
- Web languages (.html, .xml, .css)
- Config files (.yaml, .yml, .toml, .ini, .conf)
- And many more

## Extending for New File Types

Adding support for new file types is easy. You can either:

1. Add them to the configuration file
2. Update the `initializeFileTypes` function in `processor.go`

## Testing

PathFix includes comprehensive unit and integration tests. Run the tests with:

```bash
go test ./...
```

The test suite includes:

- Binary file detection
- GitIgnore pattern matching and file exclusion
- Comment style detection for various file types
- End-to-end file processing
- Configuration loading and validation

## License

MIT License