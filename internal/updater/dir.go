package updater

import (
	"os"
	"path/filepath"
)

// sourceDir is the canonical resolver for the bot's source/binary directory.
// Override in tests: updater.sourceDir = func() string { return t.TempDir() }
var sourceDir = func() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

// Dir returns the source/binary directory for this process.
func Dir() string { return sourceDir() }
