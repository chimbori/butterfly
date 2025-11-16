package core

import (
	"os"
	"path/filepath"
)

// CreateFile creates a file at the specified relative path, & returns a file handle.
func CreateFile(relPath string) (*os.File, error) {
	absPath, err := filepath.Abs(relPath)
	if err != nil {
		return nil, err
	}

	absDir := filepath.Dir(absPath)
	err = os.MkdirAll(absDir, 0o755)
	if err != nil {
		return nil, err
	}

	f, err := os.Create(absPath)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// FileExists checks if a file exists and is not a directory.
func FileExists(filename string) (bool, error) {
	info, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return !info.IsDir(), nil
}
