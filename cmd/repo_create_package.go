package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/hk9890/ai-config-manager/pkg/metadata"
	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
	"github.com/spf13/cobra"
)

var (
	pkgDescriptionFlag string
	pkgResourcesFlag   string
	pkgForceFlag       bool
)

// repoCreatePackageCmd represents the repo create-package command
var repoCreatePackageCmd = &cobra.Command{
	Use:   "create-package <name>",
	Short: "Create a package from existing resources",
	Long: `Create a package from existing resources in the repository.

A package is a collection of resources (commands, skills, agents) that can be
installed together as a unit. Packages are useful for grouping related resources
or distributing collections of tools.

Resource Format:
  Resources are specified in type/name format:
    - command/name
    - skill/name
    - agent/name

Examples:
  # Create package with description and resources
  aimgr repo create-package my-tools \
    --description="My development tools" \
    --resources="command/test,skill/pdf-processing,agent/reviewer"

  # Create package (overwrite if exists)
  aimgr repo create-package my-tools \
    --description="Updated tools" \
    --resources="command/test,command/lint" \
    --force

Package files are stored in:
  - Package: ~/.local/share/ai-config/repo/packages/<name>.package.json
  - Metadata: ~/.local/share/ai-config/repo/.metadata/packages/<name>-metadata.json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		packageName := args[0]

		// Validate package name
		if err := resource.ValidateName(packageName); err != nil {
			return fmt.Errorf("invalid package name: %w", err)
		}

		// Create manager
		manager, err := repo.NewManager()
		if err != nil {
			return err
		}

		// Ensure repo is initialized
		if err := manager.Init(); err != nil {
			return err
		}

		// Check for required flags in non-interactive mode
		if pkgDescriptionFlag == "" {
			return fmt.Errorf("--description flag is required")
		}
		if pkgResourcesFlag == "" {
			return fmt.Errorf("--resources flag is required (comma-separated list in type/name format)")
		}

		// Parse resources flag
		resourceRefs := parseResourcesFlag(pkgResourcesFlag)
		if len(resourceRefs) == 0 {
			return fmt.Errorf("at least one resource must be specified")
		}

		// Validate all resources exist in repository and parse references
		validatedResources := []string{}
		for _, ref := range resourceRefs {
			resType, resName, err := resource.ParseResourceReference(ref)
			if err != nil {
				return fmt.Errorf("invalid resource reference '%s': %w", ref, err)
			}

			// Check if resource exists in repository
			_, err = manager.Get(resName, resType)
			if err != nil {
				return fmt.Errorf("resource '%s' not found in repository", ref)
			}

			// Add to validated list
			validatedResources = append(validatedResources, ref)
		}

		// Check if package already exists
		packagePath := resource.GetPackagePath(packageName, manager.GetRepoPath())
		if !pkgForceFlag {
			if _, err := resource.LoadPackage(packagePath); err == nil {
				return fmt.Errorf("package '%s' already exists (use --force to overwrite)", packageName)
			}
		}

		// Create package struct
		pkg := &resource.Package{
			Name:        packageName,
			Description: pkgDescriptionFlag,
			Resources:   validatedResources,
		}

		// Save package
		if err := resource.SavePackage(pkg, manager.GetRepoPath()); err != nil {
			return fmt.Errorf("failed to save package: %w", err)
		}

		// Create and save metadata
		now := time.Now()
		pkgMetadata := &metadata.PackageMetadata{
			Name:          packageName,
			SourceType:    "manual",
			SourceURL:     "",
			FirstAdded:    now,
			LastUpdated:   now,
			ResourceCount: len(validatedResources),
		}

		if err := metadata.SavePackageMetadata(pkgMetadata, manager.GetRepoPath()); err != nil {
			return fmt.Errorf("failed to save package metadata: %w", err)
		}

		// Print success message
		fmt.Printf("✓ Created package: %s (%d resources)\n", packageName, len(validatedResources))
		if pkgDescriptionFlag != "" {
			fmt.Printf("  Description: %s\n", pkgDescriptionFlag)
		}
		fmt.Println("\nResources:")
		for _, ref := range validatedResources {
			fmt.Printf("  - %s\n", ref)
		}

		return nil
	},
}

// parseResourcesFlag parses a comma-separated list of resource references
func parseResourcesFlag(resourcesStr string) []string {
	if resourcesStr == "" {
		return nil
	}

	// Split by comma
	parts := strings.Split(resourcesStr, ",")

	// Trim whitespace from each part
	var resources []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			resources = append(resources, trimmed)
		}
	}

	return resources
}

func init() {
	repoCmd.AddCommand(repoCreatePackageCmd)

	// Add flags
	repoCreatePackageCmd.Flags().StringVar(&pkgDescriptionFlag, "description", "", "Package description (required)")
	repoCreatePackageCmd.Flags().StringVar(&pkgResourcesFlag, "resources", "", "Comma-separated list of resources (type/name format)")
	repoCreatePackageCmd.Flags().BoolVarP(&pkgForceFlag, "force", "f", false, "Overwrite existing package")
}
