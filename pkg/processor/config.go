// File: pkg/processor/config.go
package processor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yourusername/pathfix/pkg/models"
)

// LoadConfig loads configuration from the specified file
func LoadConfig(configPath string) (*models.Config, error) {
	// Default configuration
	config := &models.Config{
		CommentPrefix:     "File: ",
		FileTypes:         make(map[string]models.CommentStyle),
		AdditionalIgnores: []string{},
	}

	// If no config file specified, return defaults
	if configPath == "" {
		return config, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Parse JSON
	err = json.Unmarshal(data, config)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	return config, nil
}

// SaveConfig saves configuration to the specified file
func SaveConfig(config *models.Config, configPath string) error {
	// Marshal to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}

// MergeConfig merges the provided configuration with the default file types
func MergeConfig(config *models.Config, defaultFileTypes map[string]models.CommentStyle) *models.Config {
	// If config file types is empty, use defaults
	if len(config.FileTypes) == 0 {
		config.FileTypes = defaultFileTypes
	} else {
		// Otherwise, merge with defaults
		for ext, style := range defaultFileTypes {
			if _, ok := config.FileTypes[ext]; !ok {
				config.FileTypes[ext] = style
			}
		}
	}

	return config
}