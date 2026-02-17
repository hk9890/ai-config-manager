package repo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func (m *Manager) isGitRepo() bool {
	gitDir := filepath.Join(m.repoPath, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

// CommitChanges commits all changes in the repository with a message
// Returns nil if successful, or an error if the commit fails
// If not a git repo, returns nil (non-fatal - operations work without git)
func (m *Manager) CommitChanges(message string) error {
	if !m.isGitRepo() {
		// Not a git repo - this is not an error, just skip
		return nil
	}

	// Stage all changes (respecting .gitignore)
	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = m.repoPath
	if output, err := addCmd.CombinedOutput(); err != nil {
		if m.logger != nil {
			m.logger.Error("failed to stage changes",
				"path", m.repoPath,
				"error", err.Error(),
				"output", string(output),
			)
		}
		return fmt.Errorf("failed to stage changes: %w\nOutput: %s", err, output)
	}

	// Check if there are changes to commit
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusCmd.Dir = m.repoPath
	output, err := statusCmd.CombinedOutput()
	if err != nil {
		if m.logger != nil {
			m.logger.Error("failed to check git status",
				"path", m.repoPath,
				"error", err.Error(),
				"output", string(output),
			)
		}
		return fmt.Errorf("failed to check git status: %w\nOutput: %s", err, output)
	}

	// If no changes, nothing to commit
	if len(output) == 0 {
		return nil
	}

	// Create commit
	commitCmd := exec.Command("git", "commit", "-m", message)
	commitCmd.Dir = m.repoPath
	if output, err := commitCmd.CombinedOutput(); err != nil {
		if m.logger != nil {
			m.logger.Error("failed to commit changes",
				"path", m.repoPath,
				"message", message,
				"error", err.Error(),
				"output", string(output),
			)
		}
		return fmt.Errorf("failed to commit changes: %w\nOutput: %s", err, output)
	}

	return nil
}

// copyFile copies a single file from src to dst
