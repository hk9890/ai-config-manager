package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hk9890/ai-config-manager/pkg/metadata"
	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
	"github.com/spf13/cobra"
)

// VerifyResult contains the results of repository verification
type VerifyResult struct {
	ResourcesWithoutMetadata []ResourceIssue `json:"resources_without_metadata,omitempty"`
	OrphanedMetadata         []MetadataIssue `json:"orphaned_metadata,omitempty"`
	MissingSourcePaths       []MetadataIssue `json:"missing_source_paths,omitempty"`
	TypeMismatches           []TypeMismatch  `json:"type_mismatches,omitempty"`
	PackagesWithMissingRefs  []PackageIssue  `json:"packages_with_missing_refs,omitempty"`
	HasErrors                bool            `json:"has_errors"`
	HasWarnings              bool            `json:"has_warnings"`
}

// ResourceIssue represents a resource with an issue
type ResourceIssue struct {
	Name string                `json:"name"`
	Type resource.ResourceType `json:"type"`
	Path string                `json:"path"`
}

// MetadataIssue represents a metadata file with an issue
type MetadataIssue struct {
	Name       string                `json:"name"`
	Type       resource.ResourceType `json:"type"`
	Path       string                `json:"path"`
	SourcePath string                `json:"source_path,omitempty"`
}

// TypeMismatch represents a mismatch between resource and metadata types
type TypeMismatch struct {
	Name         string                `json:"name"`
	ResourceType resource.ResourceType `json:"resource_type"`
	MetadataType resource.ResourceType `json:"metadata_type"`
	ResourcePath string                `json:"resource_path"`
	MetadataPath string                `json:"metadata_path"`
}

// PackageIssue represents a package with missing resource references
type PackageIssue struct {
	Name             string   `json:"name"`
	Path             string   `json:"path"`
	MissingResources []string `json:"missing_resources"`
}

var (
	verifyFix  bool
	verifyJSON bool
)

// repoVerifyCmd represents the repo verify command
var repoVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Check repository metadata and package integrity",
	Long: `Check for consistency issues between resources and metadata, and validate package references.

This command performs the following checks:
  - Resources without metadata (warning)
  - Orphaned metadata files with missing resources (error)
  - Metadata with non-existent source paths (warning)
  - Type mismatches between resource and metadata (error)
  - Packages with missing resource references (error)

Use --fix to automatically resolve issues:
  - Create missing metadata for resources
  - Remove orphaned metadata files

Exit status:
  0 - No errors found
  1 - Errors found (orphaned metadata, type mismatches, or broken package references)

Examples:
  aimgr repo verify              # Check for issues
  aimgr repo verify --fix        # Fix issues automatically
  aimgr repo verify --json       # Machine-readable output`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create a new repo manager
		manager, err := repo.NewManager()
		if err != nil {
			return err
		}

		repoPath := manager.GetRepoPath()

		// Check if repository exists
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			if verifyJSON {
				result := VerifyResult{HasErrors: false, HasWarnings: false}
				output, _ := json.MarshalIndent(result, "", "  ")
				fmt.Println(string(output))
				return nil
			}
			fmt.Println("Repository not initialized")
			return nil
		}

		// Run verification
		result, err := verifyRepository(manager, verifyFix)
		if err != nil {
			return fmt.Errorf("verification failed: %w", err)
		}

		// Output results
		if verifyJSON {
			output, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON output: %w", err)
			}
			fmt.Println(string(output))
		} else {
			displayVerifyResults(result, verifyFix)
		}

		// Exit with non-zero status if errors found
		if result.HasErrors {
			os.Exit(1)
		}

		return nil
	},
}

// verifyRepository performs repository verification checks
func verifyRepository(manager *repo.Manager, fix bool) (*VerifyResult, error) {
	result := &VerifyResult{}
	repoPath := manager.GetRepoPath()

	// Get all resources in the repository
	allResources, err := manager.List(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}

	// Build resource map for quick lookup
	resourceMap := make(map[string]resource.Resource)
	for _, res := range allResources {
		key := fmt.Sprintf("%s:%s", res.Type, res.Name)
		resourceMap[key] = res
	}

	// Check 1: Resources without metadata
	for _, res := range allResources {
		// Skip packages - they use PackageMetadata and are validated separately
		if res.Type == resource.PackageType {
			continue
		}

		meta, err := manager.GetMetadata(res.Name, res.Type)
		if err != nil {
			// Metadata doesn't exist
			issue := ResourceIssue{
				Name: res.Name,
				Type: res.Type,
				Path: manager.GetPath(res.Name, res.Type),
			}
			result.ResourcesWithoutMetadata = append(result.ResourcesWithoutMetadata, issue)
			result.HasWarnings = true

			// Fix: Create metadata
			if fix {
				if err := createMetadataForResource(manager, res); err != nil {
					return nil, fmt.Errorf("failed to create metadata for %s: %w", res.Name, err)
				}
			}
			continue
		}

		// Check 2: Type mismatches
		if meta.Type != res.Type {
			mismatch := TypeMismatch{
				Name:         res.Name,
				ResourceType: res.Type,
				MetadataType: meta.Type,
				ResourcePath: manager.GetPath(res.Name, res.Type),
				MetadataPath: metadata.GetMetadataPath(res.Name, meta.Type, repoPath),
			}
			result.TypeMismatches = append(result.TypeMismatches, mismatch)
			result.HasErrors = true
		}

		// Check 3: Metadata with missing source paths
		if meta.SourceURL != "" && strings.HasPrefix(meta.SourceURL, "file://") {
			sourcePath := strings.TrimPrefix(meta.SourceURL, "file://")
			if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
				issue := MetadataIssue{
					Name:       meta.Name,
					Type:       meta.Type,
					Path:       metadata.GetMetadataPath(meta.Name, meta.Type, repoPath),
					SourcePath: sourcePath,
				}
				result.MissingSourcePaths = append(result.MissingSourcePaths, issue)
				result.HasWarnings = true
			}
		}
	}

	// Check 4: Orphaned metadata (metadata without resources)
	metadataDir := filepath.Join(repoPath, ".metadata")
	if _, err := os.Stat(metadataDir); err == nil {
		orphanedMeta, err := findVerifyOrphanedMetadata(manager, resourceMap, metadataDir, fix)
		if err != nil {
			return nil, fmt.Errorf("failed to check for orphaned metadata: %w", err)
		}
		result.OrphanedMetadata = orphanedMeta
		if len(orphanedMeta) > 0 {
			result.HasErrors = true
		}
	}

	// Check 5: Packages with missing resource references
	packageInfos, err := manager.ListPackages()
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}

	for _, pkgInfo := range packageInfos {
		// Load the full package
		pkg, err := manager.GetPackage(pkgInfo.Name)
		if err != nil {
			// Skip packages that cannot be loaded
			continue
		}

		missingRefs := manager.ValidatePackageResources(pkg)
		if len(missingRefs) > 0 {
			issue := PackageIssue{
				Name:             pkg.Name,
				Path:             resource.GetPackagePath(pkg.Name, repoPath),
				MissingResources: missingRefs,
			}
			result.PackagesWithMissingRefs = append(result.PackagesWithMissingRefs, issue)
			result.HasErrors = true
		}
	}

	return result, nil
}

// findVerifyOrphanedMetadata finds metadata files without corresponding resources
func findVerifyOrphanedMetadata(manager *repo.Manager, resourceMap map[string]resource.Resource, metadataDir string, fix bool) ([]MetadataIssue, error) {
	var orphaned []MetadataIssue

	// Check each resource type
	for _, resType := range []resource.ResourceType{resource.Command, resource.Skill, resource.Agent} {
		typeDir := filepath.Join(metadataDir, string(resType)+"s")
		if _, err := os.Stat(typeDir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(typeDir)
		if err != nil {
			return nil, fmt.Errorf("failed to read metadata directory %s: %w", typeDir, err)
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), "-metadata.json") {
				continue
			}

			// Read metadata file to get actual resource name
			// This handles nested paths correctly (e.g., "dt/cluster/overview")
			// Filename uses hyphens but metadata content preserves slashes
			metaPath := filepath.Join(typeDir, entry.Name())
			data, err := os.ReadFile(metaPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read metadata file %s: %w", metaPath, err)
			}
			var meta metadata.ResourceMetadata
			if err := json.Unmarshal(data, &meta); err != nil {
				return nil, fmt.Errorf("failed to parse metadata file %s: %w", metaPath, err)
			}

			name := meta.Name
			key := fmt.Sprintf("%s:%s", resType, name)

			// Check if corresponding resource exists
			if _, exists := resourceMap[key]; !exists {
				issue := MetadataIssue{
					Name: name,
					Type: resType,
					Path: metaPath,
				}
				orphaned = append(orphaned, issue)

				// Fix: Remove orphaned metadata
				if fix {
					if err := os.Remove(metaPath); err != nil {
						return nil, fmt.Errorf("failed to remove orphaned metadata %s: %w", metaPath, err)
					}
				}
			}
		}
	}

	return orphaned, nil
}

// createMetadataForResource creates metadata for a resource that's missing it
func createMetadataForResource(manager *repo.Manager, res resource.Resource) error {
	repoPath := manager.GetRepoPath()
	resourcePath := manager.GetPath(res.Name, res.Type)

	// Get absolute path for source URL
	absPath, err := filepath.Abs(resourcePath)
	if err != nil {
		absPath = resourcePath
	}

	// Get file modification time
	fileInfo, err := os.Stat(resourcePath)
	if err != nil {
		return fmt.Errorf("failed to stat resource file: %w", err)
	}

	meta := &metadata.ResourceMetadata{
		Name:           res.Name,
		Type:           res.Type,
		SourceType:     "local",
		SourceURL:      "file://" + absPath,
		FirstInstalled: fileInfo.ModTime(),
		LastUpdated:    fileInfo.ModTime(),
	}

	return metadata.Save(meta, repoPath)
}

// displayVerifyResults displays verification results in human-readable format
func displayVerifyResults(result *VerifyResult, fixed bool) {
	fmt.Println("Repository Verification")
	fmt.Println("=======================")
	fmt.Println()

	hasIssues := false

	// Display resources without metadata (warning)
	if len(result.ResourcesWithoutMetadata) > 0 {
		hasIssues = true
		fmt.Printf("⚠ Resources without metadata: %d\n", len(result.ResourcesWithoutMetadata))
		for _, issue := range result.ResourcesWithoutMetadata {
			if fixed {
				fmt.Printf("  ✓ Created metadata for %s (%s)\n", issue.Name, issue.Type)
			} else {
				fmt.Printf("  - %s (%s) at %s\n", issue.Name, issue.Type, issue.Path)
			}
		}
		fmt.Println()
	}

	// Display orphaned metadata (error)
	if len(result.OrphanedMetadata) > 0 {
		hasIssues = true
		fmt.Printf("✗ Orphaned metadata (resource missing): %d\n", len(result.OrphanedMetadata))
		for _, issue := range result.OrphanedMetadata {
			if fixed {
				fmt.Printf("  ✓ Removed orphaned metadata for %s (%s)\n", issue.Name, issue.Type)
			} else {
				fmt.Printf("  - %s (%s) at %s\n", issue.Name, issue.Type, issue.Path)
			}
		}
		fmt.Println()
	}

	// Display metadata with missing source paths (warning)
	if len(result.MissingSourcePaths) > 0 {
		hasIssues = true
		fmt.Printf("⚠ Metadata with missing source paths: %d\n", len(result.MissingSourcePaths))
		for _, issue := range result.MissingSourcePaths {
			fmt.Printf("  - %s (%s)\n", issue.Name, issue.Type)
			fmt.Printf("    Source: %s\n", issue.SourcePath)
		}
		fmt.Println()
	}

	// Display type mismatches (error)
	if len(result.TypeMismatches) > 0 {
		hasIssues = true
		fmt.Printf("✗ Type mismatches: %d\n", len(result.TypeMismatches))
		for _, mismatch := range result.TypeMismatches {
			metaTypeStr := string(mismatch.MetadataType)
			if metaTypeStr == "" {
				metaTypeStr = "(empty - metadata may be corrupted)"
			}
			fmt.Printf("  - %s: resource is %s, metadata says %s\n",
				mismatch.Name, mismatch.ResourceType, metaTypeStr)
			fmt.Printf("    Resource: %s\n", mismatch.ResourcePath)
			fmt.Printf("    Metadata: %s\n", mismatch.MetadataPath)
		}
		fmt.Println()
	}

	// Display packages with missing resource references (error)
	if len(result.PackagesWithMissingRefs) > 0 {
		hasIssues = true
		fmt.Printf("✗ Packages with missing resource references: %d\n", len(result.PackagesWithMissingRefs))
		for _, issue := range result.PackagesWithMissingRefs {
			fmt.Printf("  - %s (%d missing):\n", issue.Name, len(issue.MissingResources))
			for _, ref := range issue.MissingResources {
				fmt.Printf("    - %s\n", ref)
			}
		}
		fmt.Println()
	}

	// Summary
	if !hasIssues {
		fmt.Println("✓ No issues found. Repository is healthy!")
	} else {
		if result.HasErrors {
			fmt.Println("Status: ERRORS found (exit code 1)")
		} else if result.HasWarnings {
			fmt.Println("Status: Warnings only (exit code 0)")
		}

		if !fixed && (len(result.ResourcesWithoutMetadata) > 0 || len(result.OrphanedMetadata) > 0) {
			fmt.Println()
			fmt.Println("Run 'aimgr repo verify --fix' to automatically resolve these issues.")
		}
	}
}

func init() {
	repoCmd.AddCommand(repoVerifyCmd)
	repoVerifyCmd.Flags().BoolVar(&verifyFix, "fix", false, "Automatically fix issues (create missing metadata, remove orphaned)")
	repoVerifyCmd.Flags().BoolVar(&verifyJSON, "json", false, "Output results in JSON format")
}
