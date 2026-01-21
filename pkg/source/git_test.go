package source

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func TestCleanupTempDir(t *testing.T) {
	tests := []struct {
		name      string
		setupDir  func() string
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid temp directory",
			setupDir: func() string {
				dir, err := os.MkdirTemp("", "test-cleanup-*")
				if err != nil {
					t.Fatalf("failed to create temp dir: %v", err)
				}
				return dir
			},
			wantError: false,
		},
		{
			name: "empty path",
			setupDir: func() string {
				return ""
			},
			wantError: true,
			errorMsg:  "cannot be empty",
		},
		{
			name: "directory outside temp",
			setupDir: func() string {
				return "/etc"
			},
			wantError: true,
			errorMsg:  "outside temp directory",
		},
		{
			name: "temp directory root",
			setupDir: func() string {
				return os.TempDir()
			},
			wantError: true,
			errorMsg:  "temp directory root",
		},
		{
			name: "non-existent directory in temp",
			setupDir: func() string {
				return filepath.Join(os.TempDir(), "non-existent-dir-12345")
			},
			wantError: false, // os.RemoveAll doesn't error on non-existent paths
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setupDir()

			err := CleanupTempDir(dir)

			if tt.wantError {
				if err == nil {
					t.Errorf("CleanupTempDir() expected error but got none")
					return
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("CleanupTempDir() error = %v, want error containing %q", err, tt.errorMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("CleanupTempDir() unexpected error: %v", err)
			}
		})
	}
}

func TestCleanupTempDir_SecurityValidation(t *testing.T) {
	// Test that we can't delete arbitrary directories
	dangerousPaths := []string{
		"/",
		"/home",
		"/usr",
		"/var",
		filepath.Join(os.Getenv("HOME"), "Documents"),
	}

	for _, path := range dangerousPaths {
		t.Run("refuse to delete "+path, func(t *testing.T) {
			err := CleanupTempDir(path)
			if err == nil {
				t.Errorf("CleanupTempDir(%s) should have failed but succeeded", path)
			}
			if !strings.Contains(err.Error(), "outside temp directory") &&
				!strings.Contains(err.Error(), "temp directory root") {
				t.Errorf("CleanupTempDir(%s) error should mention security check, got: %v", path, err)
			}
		})
	}
}

func TestCheckGitAvailable(t *testing.T) {
	err := checkGitAvailable()

	if !isGitAvailable() {
		// If git is not available, we should get an error
		if err == nil {
			t.Errorf("checkGitAvailable() expected error when git is not available")
		}
	} else {
		// If git is available, we should not get an error
		if err != nil {
			t.Errorf("checkGitAvailable() unexpected error: %v", err)
		}
	}
}

func TestGetCloneURL(t *testing.T) {
	tests := []struct {
		name      string
		source    *ParsedSource
		wantURL   string
		wantError bool
	}{
		{
			name: "GitHub without ref",
			source: &ParsedSource{
				Type: GitHub,
				URL:  "https://github.com/owner/repo",
				Ref:  "",
			},
			wantURL:   "https://github.com/owner/repo",
			wantError: false,
		},
		{
			name: "GitHub with ref",
			source: &ParsedSource{
				Type: GitHub,
				URL:  "https://github.com/owner/repo/tree/main",
				Ref:  "main",
			},
			wantURL:   "https://github.com/owner/repo",
			wantError: false,
		},
		{
			name: "GitLab",
			source: &ParsedSource{
				Type: GitLab,
				URL:  "https://gitlab.com/group/project",
			},
			wantURL:   "https://gitlab.com/group/project",
			wantError: false,
		},
		{
			name: "Generic git URL",
			source: &ParsedSource{
				Type: GitURL,
				URL:  "https://git.example.com/repo.git",
			},
			wantURL:   "https://git.example.com/repo.git",
			wantError: false,
		},
		{
			name: "Local source",
			source: &ParsedSource{
				Type:      Local,
				LocalPath: "/path/to/local",
			},
			wantError: true,
		},
		{
			name:      "nil source",
			source:    nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotURL, err := GetCloneURL(tt.source)

			if tt.wantError {
				if err == nil {
					t.Errorf("GetCloneURL() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("GetCloneURL() unexpected error: %v", err)
				return
			}

			if gotURL != tt.wantURL {
				t.Errorf("GetCloneURL() = %v, want %v", gotURL, tt.wantURL)
			}
		})
	}
}

// isGitAvailable checks if git is available for testing
func isGitAvailable() bool {
	cmd := exec.Command("git", "--version")
	return cmd.Run() == nil
}
