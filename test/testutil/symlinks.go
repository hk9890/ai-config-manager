package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// CreateSymlinkedDir creates a source directory and symlinks it to target location.
// This helper is used to test that code properly handles symlinked directories
// in addition to real directories (SYMLINK mode vs COPY mode).
//
// Returns:
//   - sourceDir: The real source directory path
//   - symlinkPath: The symlink path pointing to sourceDir
//   - cleanup: Function to clean up both directories
//
// Example:
//
//	sourceDir, symlinkPath, cleanup := CreateSymlinkedDir(t, "my-skill")
//	defer cleanup()
//
//	// Write to source
//	os.WriteFile(filepath.Join(sourceDir, "SKILL.md"), []byte("content"), 0644)
//
//	// Read from symlink (should work identically)
//	data, err := os.ReadFile(filepath.Join(symlinkPath, "SKILL.md"))
func CreateSymlinkedDir(t *testing.T, name string) (string, string, func()) {
	t.Helper()

	// Create source directory
	sourceDir := filepath.Join(t.TempDir(), "source", name)
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("failed to create source directory: %v", err)
	}

	// Create symlink directory
	symlinkBase := filepath.Join(t.TempDir(), "symlinked")
	if err := os.MkdirAll(symlinkBase, 0755); err != nil {
		t.Fatalf("failed to create symlink base directory: %v", err)
	}

	symlinkPath := filepath.Join(symlinkBase, name)
	if err := os.Symlink(sourceDir, symlinkPath); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	cleanup := func() {
		// Note: t.TempDir() handles cleanup automatically,
		// but we provide this for explicit cleanup if needed
		os.RemoveAll(filepath.Dir(sourceDir))
		os.RemoveAll(symlinkBase)
	}

	return sourceDir, symlinkPath, cleanup
}
