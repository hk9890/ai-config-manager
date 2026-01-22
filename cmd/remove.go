package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
	"github.com/spf13/cobra"
)

var removeForceFlag bool

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:     "remove [command|skill]",
	Aliases: []string{"rm"},
	Short:   "Remove a resource from the repository",
	Long: `Remove a command or skill resource from the aimgr repository.

This permanently deletes the resource from the repository.`,
}

// removeCommandCmd represents the remove command subcommand
var removeCommandCmd = &cobra.Command{
	Use:   "command <name>",
	Short: "Remove a command resource",
	Long: `Remove a command resource from the repository.

This permanently deletes the command file from the repository.

Example:
  aimgr remove command my-command
  aimgr rm command test --force`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Create repo manager
		manager, err := repo.NewManager()
		if err != nil {
			return err
		}

		// Verify command exists
		res, err := manager.Get(name, resource.Command)
		if err != nil {
			return fmt.Errorf("command '%s' not found in repository. Use 'aimgr list' to see available resources.", name)
		}

		// Confirmation prompt (unless --force)
		if !removeForceFlag {
			fmt.Printf("Remove command '%s'?\n", name)
			if res.Description != "" {
				fmt.Printf("  Description: %s\n", res.Description)
			}
			fmt.Print("[y/N] ")

			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}

			response = strings.ToLower(strings.TrimSpace(response))
			if response != "y" && response != "yes" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		// Remove command
		if err := manager.Remove(name, resource.Command); err != nil {
			return fmt.Errorf("failed to remove command: %w", err)
		}

		// Success message
		fmt.Printf("✓ Removed command '%s' from repository\n", name)

		return nil
	},
}

// removeSkillCmd represents the remove skill subcommand
var removeSkillCmd = &cobra.Command{
	Use:   "skill <name>",
	Short: "Remove a skill resource",
	Long: `Remove a skill resource from the repository.

This permanently deletes the skill folder with all scripts, references, and assets.

Example:
  aimgr remove skill pdf-processing
  aimgr rm skill my-skill --force`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Create repo manager
		manager, err := repo.NewManager()
		if err != nil {
			return err
		}

		// Verify skill exists
		res, err := manager.Get(name, resource.Skill)
		if err != nil {
			return fmt.Errorf("skill '%s' not found in repository. Use 'aimgr list' to see available resources.", name)
		}

		// Confirmation prompt (unless --force)
		if !removeForceFlag {
			fmt.Printf("Remove skill '%s'", name)
			if res.Version != "" {
				fmt.Printf(" (v%s)", res.Version)
			}
			fmt.Println("?")
			if res.Description != "" {
				fmt.Printf("  Description: %s\n", res.Description)
			}
			fmt.Println("  This will delete the folder with all scripts, references, and assets.")
			fmt.Print("[y/N] ")

			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}

			response = strings.ToLower(strings.TrimSpace(response))
			if response != "y" && response != "yes" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		// Remove skill
		if err := manager.Remove(name, resource.Skill); err != nil {
			return fmt.Errorf("failed to remove skill: %w", err)
		}

		// Success message
		fmt.Printf("✓ Removed skill '%s' from repository\n", name)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
	removeCmd.AddCommand(removeCommandCmd)
	removeCmd.AddCommand(removeSkillCmd)

	// Add --force flag to both subcommands
	removeCommandCmd.Flags().BoolVarP(&removeForceFlag, "force", "f", false, "Skip confirmation prompt")
	removeSkillCmd.Flags().BoolVarP(&removeForceFlag, "force", "f", false, "Skip confirmation prompt")
}
