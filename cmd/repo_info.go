package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
	"github.com/spf13/cobra"
)

// repoInfoCmd represents the repo info command
var repoInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Display repository information and statistics",
	Long: `Display comprehensive information about the aimgr repository.

Shows repository location, total resource counts, breakdown by type,
and other useful statistics.

Example:
  aimgr repo info`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create a new repo manager
		manager, err := repo.NewManager()
		if err != nil {
			return err
		}

		// Get repository path
		repoPath := manager.GetRepoPath()

		// Check if repository exists
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			fmt.Println("Repository not initialized")
			fmt.Printf("Expected location: %s\n", repoPath)
			fmt.Println()
			fmt.Println("Run 'aimgr repo add <resource>' to add resources and initialize the repository.")
			return nil
		}

		// List all resources to get counts
		allResources, err := manager.List(nil)
		if err != nil {
			return fmt.Errorf("failed to list resources: %w", err)
		}

		// Count by type
		commandCount := 0
		skillCount := 0
		agentCount := 0

		for _, res := range allResources {
			switch res.Type {
			case resource.Command:
				commandCount++
			case resource.Skill:
				skillCount++
			case resource.Agent:
				agentCount++
			}
		}

		// Display the information
		fmt.Println("Repository Information")
		fmt.Println("======================")
		fmt.Println()
		fmt.Printf("Location: %s\n", repoPath)
		fmt.Println()
		fmt.Printf("Total Resources: %d\n", len(allResources))
		fmt.Printf("  Commands:      %d\n", commandCount)
		fmt.Printf("  Skills:        %d\n", skillCount)
		fmt.Printf("  Agents:        %d\n", agentCount)
		fmt.Println()

		// Calculate and display disk usage
		size, err := calculateDirSize(repoPath)
		if err == nil {
			fmt.Printf("Disk Usage:    %s\n", formatBytes(size))
			fmt.Println()
		}

		return nil
	},
}

// calculateDirSize calculates the total size of a directory
func calculateDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// formatBytes formats bytes as human-readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func init() {
	repoCmd.AddCommand(repoInfoCmd)
}
