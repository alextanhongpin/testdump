package internal

import (
	"errors"
	"os"
	"path/filepath"
)

// WriteFile writes a file to the filesystem.
func WriteFile(name string, body []byte, overwrite bool) (bool, error) {
	f, err := os.OpenFile(name, os.O_RDONLY, 0644)
	if errors.Is(err, os.ErrNotExist) || overwrite {
		dir := filepath.Dir(name)

		// Create the directory first
		if err := os.MkdirAll(dir, 0700); err != nil && !os.IsExist(err) {
			return false, err
		}

		// Write the file.
		if err := os.WriteFile(name, body, 0644); err != nil {
			return false, err
		}

		return true, nil
	}
	if err != nil {
		return false, err
	}

	defer f.Close()

	return false, nil
}
