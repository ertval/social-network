package gates

import (
	"os"
	"path/filepath"
)

// isFeatureSlice checks if a directory under internalDir is a feature slice.
// It matches if the directory contains a `.feature-slice` marker file,
// OR if it is a directory under internalDir that is not in the skipDirs list.
func isFeatureSlice(internalDir, name string) bool {
	// Marker file takes precedence
	marker := filepath.Join(internalDir, name, ".feature-slice")
	if _, err := os.Stat(marker); err == nil {
		return true
	}

	// Restrict skipDirs check to top-level directories of internalDir
	if skipDirs[name] {
		return false
	}

	// By default, if it's a directory and not in skipDirs, treat it as a feature slice
	dirPath := filepath.Join(internalDir, name)
	info, err := os.Stat(dirPath)
	if err == nil && info.IsDir() {
		return true
	}

	return false
}
