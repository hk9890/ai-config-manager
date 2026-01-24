package metadata

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hk9890/ai-config-manager/pkg/resource"
)

// MigrationResult tracks the results of a metadata migration
type MigrationResult struct {
	TotalFiles   int
	MovedFiles   int
	SkippedFiles int
	Errors       []error
}

// MigrationOptions configures migration behavior
type MigrationOptions struct {
	DryRun bool // Preview changes without moving files
}

// MigrateMetadataFiles migrates all existing metadata files from their current locations
// to the new .metadata/ directory structure.
//
// Migration process:
//  1. Scans each resource type directory (commands/, skills/, agents/)
//  2. Finds all *-metadata.json files
//  3. Parses resource name from filename (removes type prefix and -metadata.json suffix)
//  4. Creates .metadata/<type>s/ directory if it doesn't exist
//  5. Moves file to new location with simplified name: <name>-metadata.json
//  6. Verifies file was moved successfully
//
// Old pattern: /repo/<type>s/<type>-<name>-metadata.json
// New pattern: /repo/.metadata/<type>s/<name>-metadata.json
//
// Example:
//   - Old: /repo/skills/skill-pdf-processor-metadata.json
//   - New: /repo/.metadata/skills/pdf-processor-metadata.json
func MigrateMetadataFiles(repoPath string) (*MigrationResult, error) {
	return MigrateMetadataFilesWithOptions(repoPath, MigrationOptions{DryRun: false})
}

// MigrateMetadataFilesWithOptions migrates metadata files with custom options
func MigrateMetadataFilesWithOptions(repoPath string, opts MigrationOptions) (*MigrationResult, error) {
	if repoPath == "" {
		return nil, fmt.Errorf("repoPath cannot be empty")
	}

	// Verify repo path exists
	if _, err := os.Stat(repoPath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("repository path does not exist: %s", repoPath)
		}
		return nil, fmt.Errorf("failed to access repository path: %w", err)
	}

	result := &MigrationResult{
		Errors: make([]error, 0),
	}

	// Resource types to migrate
	resourceTypes := []resource.ResourceType{
		resource.Command,
		resource.Skill,
		resource.Agent,
	}

	// Migrate each resource type
	for _, resType := range resourceTypes {
		if err := migrateResourceType(repoPath, resType, result, opts.DryRun); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to migrate %s metadata: %w", resType, err))
		}
	}

	return result, nil
}

// migrateResourceType migrates metadata files for a single resource type
func migrateResourceType(repoPath string, resType resource.ResourceType, result *MigrationResult, dryRun bool) error {
	// Old directory: /repo/<type>s/
	oldDir := filepath.Join(repoPath, string(resType)+"s")

	// Check if old directory exists
	if _, err := os.Stat(oldDir); err != nil {
		if os.IsNotExist(err) {
			// Directory doesn't exist, nothing to migrate
			return nil
		}
		return fmt.Errorf("failed to access directory %s: %w", oldDir, err)
	}

	// Read directory contents
	entries, err := os.ReadDir(oldDir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", oldDir, err)
	}

	// Process each file
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()

		// Check if this is a metadata file
		if !isOldMetadataFile(filename, resType) {
			continue
		}

		result.TotalFiles++

		// Parse resource name from filename
		name, err := parseResourceName(filename, resType)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to parse resource name from %s: %w", filename, err))
			continue
		}

		// Build old and new paths
		oldPath := filepath.Join(oldDir, filename)
		newPath := getNewMetadataPath(repoPath, name, resType)

		// Check if target file already exists
		if _, err := os.Stat(newPath); err == nil {
			// File already exists at new location, skip migration
			result.SkippedFiles++
			if !dryRun {
				fmt.Printf("Skipping %s (already exists at new location)\n", filename)
			}
			continue
		}

		// Migrate the file
		if err := migrateFile(oldPath, newPath, dryRun); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to migrate %s: %w", filename, err))
			continue
		}

		result.MovedFiles++
		if !dryRun {
			fmt.Printf("Migrated: %s -> %s\n", oldPath, newPath)
		}
	}

	return nil
}

// isOldMetadataFile checks if a filename matches the old metadata pattern
// Pattern: <type>-<name>-metadata.json
func isOldMetadataFile(filename string, resType resource.ResourceType) bool {
	// Must end with -metadata.json
	if !strings.HasSuffix(filename, "-metadata.json") {
		return false
	}

	// Must start with resource type prefix
	prefix := string(resType) + "-"
	return strings.HasPrefix(filename, prefix)
}

// parseResourceName extracts the resource name from an old metadata filename
// Pattern: <type>-<name>-metadata.json -> <name>
func parseResourceName(filename string, resType resource.ResourceType) (string, error) {
	// Remove type prefix
	prefix := string(resType) + "-"
	if !strings.HasPrefix(filename, prefix) {
		return "", fmt.Errorf("filename does not start with expected prefix %s", prefix)
	}
	withoutPrefix := strings.TrimPrefix(filename, prefix)

	// Remove -metadata.json suffix
	suffix := "-metadata.json"
	if !strings.HasSuffix(withoutPrefix, suffix) {
		return "", fmt.Errorf("filename does not end with expected suffix %s", suffix)
	}
	name := strings.TrimSuffix(withoutPrefix, suffix)

	if name == "" {
		return "", fmt.Errorf("parsed name is empty")
	}

	return name, nil
}

// getNewMetadataPath returns the new path for a metadata file
// Pattern: <repoPath>/.metadata/<type>s/<name>-metadata.json
func getNewMetadataPath(repoPath, name string, resType resource.ResourceType) string {
	// Build filename: <name>-metadata.json (no type prefix)
	filename := fmt.Sprintf("%s-metadata.json", name)

	// Build path: /repo/.metadata/<type>s/<name>-metadata.json
	return filepath.Join(repoPath, ".metadata", string(resType)+"s", filename)
}

// migrateFile moves a file from oldPath to newPath
// If dryRun is true, only checks are performed without actually moving files
func migrateFile(oldPath, newPath string, dryRun bool) error {
	// Verify old file exists and is readable
	if _, err := os.Stat(oldPath); err != nil {
		return fmt.Errorf("failed to stat old file: %w", err)
	}

	// In dry-run mode, only verify paths and permissions
	if dryRun {
		// Check if parent directory can be created (without actually creating it)
		newDir := filepath.Dir(newPath)
		if _, err := os.Stat(filepath.Dir(newDir)); err != nil {
			if os.IsNotExist(err) {
				// Parent of new directory doesn't exist, this would fail
				return fmt.Errorf("parent directory %s does not exist", filepath.Dir(newDir))
			}
		}
		return nil
	}

	// Create parent directory for new location
	newDir := filepath.Dir(newPath)
	if err := os.MkdirAll(newDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", newDir, err)
	}

	// Read old file contents
	data, err := os.ReadFile(oldPath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", oldPath, err)
	}

	// Write to new location
	if err := os.WriteFile(newPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", newPath, err)
	}

	// Verify new file exists and has same size
	oldInfo, err := os.Stat(oldPath)
	if err != nil {
		return fmt.Errorf("failed to stat old file: %w", err)
	}

	newInfo, err := os.Stat(newPath)
	if err != nil {
		return fmt.Errorf("failed to verify new file: %w", err)
	}

	if oldInfo.Size() != newInfo.Size() {
		return fmt.Errorf("file size mismatch after migration (old: %d, new: %d)", oldInfo.Size(), newInfo.Size())
	}

	// Delete old file
	if err := os.Remove(oldPath); err != nil {
		return fmt.Errorf("failed to remove old file: %w", err)
	}

	return nil
}
