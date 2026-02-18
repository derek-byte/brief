package store

import (
	"fmt"
	"os"
	"path/filepath"
)

// AppDataDir returns the application data directory path for macOS
func AppDataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, "Library", "Application Support", "branchbrief"), nil
}

// EnsureDir creates the directory if it doesn't exist
func EnsureDir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}
	return nil
}
