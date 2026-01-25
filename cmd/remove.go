package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/spf13/cobra"
)

var removeForceFlag bool

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:     "remove <pattern>...",
	Aliases: []string{"rm"},
	Short:   "Remove resources from the repository",
	Long: `Remove one or more resources from the aimgr repository using type/name patterns.

This permanently deletes resources from the repository.

Patterns support wildcards (* and ?) to match multiple resources.
Format: type/name or just name (searches all types)

Examples:
  aimgr repo remove skill/pdf-processing
  aimgr repo remove command/test-*
  aimgr repo remove agent/deprecated
  aimgr repo remove skill/old command/legacy agent/unused
  aimgr repo rm */temp-* --force`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create repo manager
		manager, err := repo.NewManager()
		if err != nil {
			return err
		}

		// Expand all patterns and collect matches
		var toRemove []string
		for _, pattern := range args {
			matches, err := ExpandPattern(manager, pattern)
			if err != nil {
				return err
			}
			toRemove = append(toRemove, matches...)
		}

		if len(toRemove) == 0 {
			return fmt.Errorf("no resources found matching patterns")
		}

		// Remove duplicates
		toRemove = uniqueStrings(toRemove)

		// Confirmation prompt (unless --force)
		if !removeForceFlag {
			fmt.Println("The following resources will be removed:")
			for _, resourceArg := range toRemove {
				resType, name, err := ParseResourceArg(resourceArg)
				if err != nil {
					return err
				}
				res, err := manager.Get(name, resType)
				if err != nil {
					// Resource might have been removed already in batch, skip
					continue
				}
				desc := res.Description
				if desc == "" {
					desc = "(no description)"
				}
				fmt.Printf("  %-8s %-30s %s\n", resType, name, desc)
			}
			fmt.Printf("\nRemove %d resource(s)? [y/N] ", len(toRemove))

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

		// Remove each resource
		successCount := 0
		for _, resourceArg := range toRemove {
			resType, name, err := ParseResourceArg(resourceArg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "✗ Failed to parse %s: %v\n", resourceArg, err)
				continue
			}

			if err := manager.Remove(name, resType); err != nil {
				fmt.Fprintf(os.Stderr, "✗ Failed to remove %s: %v\n", resourceArg, err)
			} else {
				fmt.Printf("✓ Removed %s\n", resourceArg)
				successCount++
			}
		}

		if successCount == 0 {
			return fmt.Errorf("failed to remove any resources")
		}

		if successCount < len(toRemove) {
			fmt.Fprintf(os.Stderr, "\nWarning: %d of %d resource(s) failed to remove\n", len(toRemove)-successCount, len(toRemove))
		}

		return nil
	},
}

// uniqueStrings removes duplicate strings from a slice while preserving order
func uniqueStrings(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}

func init() {
	repoCmd.AddCommand(removeCmd)

	// Add --force flag to main command
	removeCmd.Flags().BoolVarP(&removeForceFlag, "force", "f", false, "Skip confirmation prompt")
}
