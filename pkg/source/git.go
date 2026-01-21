package source

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CloneRepo clones a Git repository to a temporary directory
// Returns the path to the temporary directory containing the cloned repository
func CloneRepo(url string, ref string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("repository URL cannot be empty")
	}

	// Check if git is available
	if err := checkGitAvailable(); err != nil {
		return "", err
	}

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "ai-repo-clone-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Build git clone command with shallow clone
	args := []string{"clone", "--depth", "1"}

	// Add branch/tag reference if specified
	if ref != "" {
		args = append(args, "--branch", ref)
	}

	args = append(args, url, tempDir)

	// Execute git clone
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Clean up temp directory on failure
		_ = os.RemoveAll(tempDir)
		return "", fmt.Errorf("git clone failed: %w\nOutput: %s", err, string(output))
	}

	return tempDir, nil
}

// CleanupTempDir removes a temporary directory with security validation
// Ensures the directory is within the system's temp directory before removal
func CleanupTempDir(dir string) error {
	if dir == "" {
		return fmt.Errorf("directory path cannot be empty")
	}

	// Get absolute path
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Security check: ensure the directory is within temp directory
	tempDir := os.TempDir()
	absTempDir, err := filepath.Abs(tempDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute temp directory path: %w", err)
	}

	// Normalize paths for comparison
	absDir = filepath.Clean(absDir)
	absTempDir = filepath.Clean(absTempDir)

	// Check if the directory is within temp directory
	if !strings.HasPrefix(absDir, absTempDir) {
		return fmt.Errorf("refusing to delete directory outside temp directory: %s", absDir)
	}

	// Additional safety check: don't allow deleting the temp directory itself
	if absDir == absTempDir {
		return fmt.Errorf("refusing to delete temp directory root: %s", absDir)
	}

	// Remove the directory
	if err := os.RemoveAll(absDir); err != nil {
		return fmt.Errorf("failed to remove directory: %w", err)
	}

	return nil
}

// checkGitAvailable checks if git is available in PATH
func checkGitAvailable() error {
	cmd := exec.Command("git", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git is not available: %w (please install git)", err)
	}
	return nil
}

// GetCloneURL converts a ParsedSource to a git clone URL
// This is useful for getting the clone URL from a parsed source
func GetCloneURL(ps *ParsedSource) (string, error) {
	if ps == nil {
		return "", fmt.Errorf("parsed source cannot be nil")
	}

	switch ps.Type {
	case GitHub:
		// Convert GitHub URL to git clone URL
		// https://github.com/owner/repo/tree/branch -> https://github.com/owner/repo
		url := ps.URL
		// Remove /tree/ref suffix if present
		if ps.Ref != "" {
			url = strings.TrimSuffix(url, fmt.Sprintf("/tree/%s", ps.Ref))
		}
		return url, nil

	case GitLab:
		return ps.URL, nil

	case GitURL:
		return ps.URL, nil

	case Local:
		return "", fmt.Errorf("local sources cannot be cloned")

	default:
		return "", fmt.Errorf("unsupported source type: %s", ps.Type)
	}
}
