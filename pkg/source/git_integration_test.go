//go:build integration

package source

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCloneRepo(t *testing.T) {
	// Skip if git is not available
	if !isGitAvailable() {
		t.Skip("git is not available, skipping clone tests")
	}

	tests := []struct {
		name      string
		url       string
		ref       string
		wantError bool
	}{
		{
			name:      "clone public repo without ref",
			url:       "https://github.com/github/gitignore",
			ref:       "",
			wantError: false,
		},
		{
			name:      "clone public repo with branch ref",
			url:       "https://github.com/github/gitignore",
			ref:       "main",
			wantError: false,
		},
		{
			name:      "empty URL",
			url:       "",
			ref:       "",
			wantError: true,
		},
		{
			name:      "invalid repo URL",
			url:       "https://github.com/nonexistent/repo-that-does-not-exist-12345",
			ref:       "",
			wantError: true,
		},
		{
			name:      "invalid branch ref",
			url:       "https://github.com/github/gitignore",
			ref:       "nonexistent-branch-12345",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, err := CloneRepo(tt.url, tt.ref)

			if tt.wantError {
				if err == nil {
					t.Errorf("CloneRepo() expected error but got none")
					// Clean up if we accidentally succeeded
					if tempDir != "" {
						_ = CleanupTempDir(tempDir)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("CloneRepo() unexpected error: %v", err)
				return
			}

			if tempDir == "" {
				t.Errorf("CloneRepo() returned empty temp directory")
				return
			}

			// Verify the directory exists
			if _, err := os.Stat(tempDir); os.IsNotExist(err) {
				t.Errorf("CloneRepo() temp directory does not exist: %s", tempDir)
				return
			}

			// Verify it's a git repository
			gitDir := filepath.Join(tempDir, ".git")
			if _, err := os.Stat(gitDir); os.IsNotExist(err) {
				t.Errorf("CloneRepo() directory is not a git repository: %s", tempDir)
			}

			// Clean up
			if err := CleanupTempDir(tempDir); err != nil {
				t.Errorf("CleanupTempDir() failed: %v", err)
			}

			// Verify cleanup worked
			if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
				t.Errorf("CleanupTempDir() directory still exists after cleanup: %s", tempDir)
			}
		})
	}
}
