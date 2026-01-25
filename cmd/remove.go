package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
	"github.com/spf13/cobra"
)

var removeForceFlag bool
var removeWithResourcesFlag bool

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:     "remove <pattern>...",
	Aliases: []string{"rm"},
	Short:   "Remove resources from the repository",
	Long: `Remove one or more resources or packages from the aimgr repository using type/name patterns.

This permanently deletes resources from the repository.

Patterns support wildcards (* and ?) to match multiple resources.
Format: type/name or just name (searches all types)

For packages: Use package/name format to remove a package.
- By default, only the package definition is removed (resources remain)
- Use --with-resources flag to also remove all referenced resources
- Confirmation is required unless --force is used

Examples:
  # Remove resources
  aimgr repo remove skill/pdf-processing
  aimgr repo remove command/test-*
  aimgr repo remove agent/deprecated
  aimgr repo remove skill/old command/legacy agent/unused
  aimgr repo rm */temp-* --force

  # Remove packages
  aimgr repo remove package/my-package
  aimgr repo remove package/my-package --with-resources
  aimgr repo remove package/my-package --with-resources --force`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create repo manager
		manager, err := repo.NewManager()
		if err != nil {
			return err
		}

		// Check if removing a package
		if len(args) == 1 && strings.HasPrefix(args[0], "package/") {
			packageName := strings.TrimPrefix(args[0], "package/")
			return removePackage(packageName, manager)
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
	removeCmd.Flags().BoolVar(&removeWithResourcesFlag, "with-resources", false, "Also remove all resources referenced by the package")
}

// removePackage removes a package from the repository
func removePackage(packageName string, manager *repo.Manager) error {
	repoPath := manager.GetRepoPath()
	pkgPath := resource.GetPackagePath(packageName, repoPath)

	// Load package
	pkg, err := resource.LoadPackage(pkgPath)
	if err != nil {
		return fmt.Errorf("package '%s' not found in repository: %w", packageName, err)
	}

	fmt.Printf("Package: %s\n", pkg.Name)
	fmt.Printf("Description: %s\n\n", pkg.Description)

	// Check if we should remove resources too
	if removeWithResourcesFlag {
		// Show warning with list of resources
		fmt.Println("The following resources will also be removed:")
		for _, ref := range pkg.Resources {
			fmt.Printf("  - %s\n", ref)
		}
		fmt.Println()

		// Require confirmation unless --force
		if !removeForceFlag {
			fmt.Printf("Remove package '%s' and %d resource(s)? [y/N] ", packageName, len(pkg.Resources))

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
		removedCount := 0
		notFoundCount := 0
		errorCount := 0

		for _, ref := range pkg.Resources {
			// Parse type/name format
			resType, resName, err := resource.ParseResourceReference(ref)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  ✗ %s - invalid format: %v\n", ref, err)
				errorCount++
				continue
			}

			// Check if resource exists
			_, err = manager.Get(resName, resType)
			if err != nil {
				fmt.Printf("  ⊘ %s - not found in repo\n", ref)
				notFoundCount++
				continue
			}

			// Remove resource
			if err := manager.Remove(resName, resType); err != nil {
				fmt.Fprintf(os.Stderr, "  ✗ %s - failed to remove: %v\n", ref, err)
				errorCount++
			} else {
				fmt.Printf("  ✓ %s\n", ref)
				removedCount++
			}
		}

		fmt.Println()
		fmt.Printf("Removed %d of %d resources\n", removedCount, len(pkg.Resources))
		if notFoundCount > 0 {
			fmt.Printf("⚠ %d resource(s) not found\n", notFoundCount)
		}
		if errorCount > 0 {
			fmt.Printf("✗ %d resource(s) failed to remove\n", errorCount)
		}
	} else {
		// Just removing package, require confirmation unless --force
		if !removeForceFlag {
			fmt.Printf("Remove package '%s'? [y/N] ", packageName)

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
	}

	// Remove package file
	if err := os.Remove(pkgPath); err != nil {
		return fmt.Errorf("failed to remove package file: %w", err)
	}

	// Remove metadata file
	metadataPath := filepath.Join(repoPath, ".metadata", "packages", fmt.Sprintf("%s-metadata.json", packageName))
	if _, err := os.Stat(metadataPath); err == nil {
		if err := os.Remove(metadataPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to remove metadata file: %v\n", err)
		}
	}

	fmt.Printf("\n✓ Package '%s' removed successfully\n", packageName)
	return nil
}
