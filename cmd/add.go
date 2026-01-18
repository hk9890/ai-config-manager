package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hans-m-leitner/ai-config-manager/pkg/repo"
	"github.com/hans-m-leitner/ai-config-manager/pkg/resource"
	"github.com/spf13/cobra"
)

var forceFlag bool

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add [command|skill]",
	Short: "Add a resource to the repository",
	Long: `Add a command or skill resource to the ai-repo repository.

Commands are single .md files with YAML frontmatter.
Skills are directories containing a SKILL.md file.`,
}

// addCommandCmd represents the add command subcommand
var addCommandCmd = &cobra.Command{
	Use:   "command <path>",
	Short: "Add a command resource",
	Long: `Add a command resource to the repository.

A command is a single .md file with YAML frontmatter containing at minimum
a description field.

Example:
  ai-repo add command ~/.claude/commands/my-command.md
  ai-repo add command ./my-command.md --force`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourcePath := args[0]

		// Validate path exists
		if _, err := os.Stat(sourcePath); err != nil {
			return fmt.Errorf("path does not exist: %s", sourcePath)
		}

		// Validate it's a file
		info, err := os.Stat(sourcePath)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return fmt.Errorf("path is a directory, expected a .md file: %s", sourcePath)
		}

		// Validate it's a .md file
		if filepath.Ext(sourcePath) != ".md" {
			return fmt.Errorf("file must have .md extension: %s", sourcePath)
		}

		// Try to load and validate the command
		res, err := resource.LoadCommand(sourcePath)
		if err != nil {
			return fmt.Errorf("invalid command resource: %w", err)
		}

		// Create manager and add command
		manager, err := repo.NewManager()
		if err != nil {
			return err
		}

		// Check if already exists (if not force mode)
		if !forceFlag {
			existing, _ := manager.Get(res.Name, resource.Command)
			if existing != nil {
				return fmt.Errorf("command '%s' already exists in repository (use --force to overwrite)", res.Name)
			}
		} else {
			// Remove existing if force mode
			_ = manager.Remove(res.Name, resource.Command)
		}

		// Add the command
		if err := manager.AddCommand(sourcePath); err != nil {
			return fmt.Errorf("failed to add command: %w", err)
		}

		// Success message
		fmt.Printf("✓ Added command '%s' to repository\n", res.Name)
		if res.Description != "" {
			fmt.Printf("  Description: %s\n", res.Description)
		}

		return nil
	},
}

// addSkillCmd represents the add skill subcommand
var addSkillCmd = &cobra.Command{
	Use:   "skill <path>",
	Short: "Add a skill resource",
	Long: `Add a skill resource to the repository.

A skill is a directory containing a SKILL.md file with YAML frontmatter.
The directory name must match the 'name' field in SKILL.md.

Example:
  ai-repo add skill ~/my-skills/pdf-processing
  ai-repo add skill ./my-skill --force`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourcePath := args[0]

		// Validate path exists
		if _, err := os.Stat(sourcePath); err != nil {
			return fmt.Errorf("path does not exist: %s", sourcePath)
		}

		// Validate it's a directory
		info, err := os.Stat(sourcePath)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return fmt.Errorf("path is a file, expected a directory: %s", sourcePath)
		}

		// Check for SKILL.md
		skillMdPath := filepath.Join(sourcePath, "SKILL.md")
		if _, err := os.Stat(skillMdPath); err != nil {
			return fmt.Errorf("directory must contain SKILL.md: %s", sourcePath)
		}

		// Try to load and validate the skill
		res, err := resource.LoadSkill(sourcePath)
		if err != nil {
			return fmt.Errorf("invalid skill resource: %w", err)
		}

		// Validate folder name matches frontmatter name
		dirName := filepath.Base(sourcePath)
		if res.Name != dirName {
			return fmt.Errorf("folder name '%s' does not match skill name '%s' in SKILL.md", dirName, res.Name)
		}

		// Create manager and add skill
		manager, err := repo.NewManager()
		if err != nil {
			return err
		}

		// Check if already exists (if not force mode)
		if !forceFlag {
			existing, _ := manager.Get(res.Name, resource.Skill)
			if existing != nil {
				return fmt.Errorf("skill '%s' already exists in repository (use --force to overwrite)", res.Name)
			}
		} else {
			// Remove existing if force mode
			_ = manager.Remove(res.Name, resource.Skill)
		}

		// Add the skill
		if err := manager.AddSkill(sourcePath); err != nil {
			return fmt.Errorf("failed to add skill: %w", err)
		}

		// Success message
		fmt.Printf("✓ Added skill '%s' to repository\n", res.Name)
		if res.Version != "" {
			fmt.Printf("  Version: %s\n", res.Version)
		}
		if res.Description != "" {
			fmt.Printf("  Description: %s\n", res.Description)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.AddCommand(addCommandCmd)
	addCmd.AddCommand(addSkillCmd)

	// Add --force flag to both subcommands
	addCommandCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Overwrite existing resource")
	addSkillCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Overwrite existing resource")
}
