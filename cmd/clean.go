package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dynatrace-oss/ai-config-manager/v3/pkg/manifest"
	"github.com/dynatrace-oss/ai-config-manager/v3/pkg/output"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove all content from owned project resource directories",
	Long: `Remove all entries inside aimgr-owned project resource directories.

For detected tools (for example .claude, .opencode), aimgr owns the contents of
commands/skills/agents directories. This command removes every entry inside
those owned directories, including symlinks, broken symlinks, regular files,
and nested subdirectories. Owned root directories are kept in place.

This command does not modify ai.package.yaml or non-resource tool config files.

Examples:
  aimgr clean
  aimgr clean --project-path ~/myproject
  aimgr clean --format json

  # Common workflow: wipe and restore from manifest
  aimgr clean && aimgr repair
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath := cleanProjectPath
		if projectPath == "" {
			var err error
			projectPath, err = os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
		}

		parsedFormat, err := parseCleanFormat(cleanFormatFlag)
		if err != nil {
			return err
		}

		warnings := collectCleanWarnings(projectPath)
		printCleanWarnings(warnings)

		ownedDirs, err := detectOwnedResourceDirs(projectPath)
		if err != nil {
			return fmt.Errorf("failed to detect owned resource directories: %w", err)
		}

		result := CleanResult{
			Warnings: warnings,
			Removed:  make([]CleanRemovedEntry, 0),
			Failed:   make([]CleanFailedEntry, 0),
			Summary: CleanSummary{
				OwnedDirsDetected: len(ownedDirs),
			},
		}

		if len(ownedDirs) == 0 {
			return displayCleanResult(result, parsedFormat)
		}

		result.Removed, result.Failed = cleanOwnedResourceDirs(ownedDirs)
		result.Summary = summarizeCleanResult(ownedDirs, result.Removed, result.Failed)

		return displayCleanResult(result, parsedFormat)
	},
}

type CleanResult struct {
	Warnings []string            `json:"warnings"`
	Removed  []CleanRemovedEntry `json:"removed"`
	Failed   []CleanFailedEntry  `json:"failed"`
	Summary  CleanSummary        `json:"summary"`
}

type CleanRemovedEntry struct {
	Tool         string `json:"tool"`
	ResourceType string `json:"resource_type"`
	Path         string `json:"path"`
	EntryType    string `json:"entry_type"`
}

type CleanFailedEntry struct {
	Tool         string `json:"tool"`
	ResourceType string `json:"resource_type"`
	Path         string `json:"path"`
	EntryType    string `json:"entry_type"`
	Error        string `json:"error"`
}

type CleanSummary struct {
	OwnedDirsDetected int `json:"owned_dirs_detected"`
	OwnedDirsExisting int `json:"owned_dirs_existing"`
	Removed           int `json:"removed"`
	RemovedFiles      int `json:"removed_files"`
	RemovedSymlinks   int `json:"removed_symlinks"`
	RemovedDirs       int `json:"removed_directories"`
	Failed            int `json:"failed"`
}

func parseCleanFormat(raw string) (output.Format, error) {
	switch strings.ToLower(raw) {
	case "", "table":
		return output.Table, nil
	case "json":
		return output.JSON, nil
	default:
		return "", fmt.Errorf("invalid format: %s (valid: table, json)", raw)
	}
}

func collectCleanWarnings(projectPath string) []string {
	manifestPath := filepath.Join(projectPath, manifest.ManifestFileName)
	if _, err := os.Stat(manifestPath); err != nil {
		if os.IsNotExist(err) {
			return []string{fmt.Sprintf("Warning: %s not found — 'aimgr repair' will not be able to restore resources.", manifest.ManifestFileName)}
		}
		return []string{fmt.Sprintf("Warning: failed to check %s: %v", manifest.ManifestFileName, err)}
	}
	return nil
}

func cleanOwnedResourceDirs(ownedDirs []OwnedResourceDir) ([]CleanRemovedEntry, []CleanFailedEntry) {
	removed := make([]CleanRemovedEntry, 0)
	failed := make([]CleanFailedEntry, 0)

	for _, owned := range ownedDirs {
		if _, err := os.Stat(owned.Path); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			failed = append(failed, CleanFailedEntry{
				Tool:         owned.Tool.String(),
				ResourceType: string(owned.ResourceType),
				Path:         owned.Path,
				EntryType:    "directory",
				Error:        err.Error(),
			})
			continue
		}

		entries, err := os.ReadDir(owned.Path)
		if err != nil {
			failed = append(failed, CleanFailedEntry{
				Tool:         owned.Tool.String(),
				ResourceType: string(owned.ResourceType),
				Path:         owned.Path,
				EntryType:    "directory",
				Error:        err.Error(),
			})
			continue
		}

		for _, entry := range entries {
			entryPath := filepath.Join(owned.Path, entry.Name())
			entryType := cleanEntryTypeFromDirEntry(entry)

			if err := os.RemoveAll(entryPath); err != nil {
				failed = append(failed, CleanFailedEntry{
					Tool:         owned.Tool.String(),
					ResourceType: string(owned.ResourceType),
					Path:         entryPath,
					EntryType:    entryType,
					Error:        err.Error(),
				})
				continue
			}

			removed = append(removed, CleanRemovedEntry{
				Tool:         owned.Tool.String(),
				ResourceType: string(owned.ResourceType),
				Path:         entryPath,
				EntryType:    entryType,
			})
		}
	}

	sort.Slice(removed, func(i, j int) bool { return removed[i].Path < removed[j].Path })
	sort.Slice(failed, func(i, j int) bool { return failed[i].Path < failed[j].Path })

	return removed, failed
}

func cleanEntryTypeFromDirEntry(entry os.DirEntry) string {
	if entry.Type()&os.ModeSymlink != 0 {
		return "symlink"
	}
	if entry.IsDir() {
		return "directory"
	}
	return "file"
}

func summarizeCleanResult(ownedDirs []OwnedResourceDir, removed []CleanRemovedEntry, failed []CleanFailedEntry) CleanSummary {
	summary := CleanSummary{
		OwnedDirsDetected: len(ownedDirs),
		Removed:           len(removed),
		Failed:            len(failed),
	}

	for _, owned := range ownedDirs {
		if info, err := os.Stat(owned.Path); err == nil && info.IsDir() {
			summary.OwnedDirsExisting++
		}
	}

	for _, entry := range removed {
		switch entry.EntryType {
		case "symlink":
			summary.RemovedSymlinks++
		case "directory":
			summary.RemovedDirs++
		default:
			summary.RemovedFiles++
		}
	}

	return summary
}

func displayCleanResult(result CleanResult, format output.Format) error {
	switch format {
	case output.JSON:
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	case output.Table:
		return displayCleanTable(result)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func displayCleanTable(result CleanResult) error {
	if len(result.Removed) > 0 {
		table := output.NewTable("Path", "Type", "Tool", "Resource")
		table.WithResponsive().
			WithDynamicColumn(0).
			WithMinColumnWidths(40, 10, 12, 10)
		for _, entry := range result.Removed {
			table.AddRow(entry.Path, entry.EntryType, entry.Tool, entry.ResourceType)
		}
		if err := table.Format(output.Table); err != nil {
			return err
		}
	} else if result.Summary.OwnedDirsDetected == 0 {
		fmt.Println("No tool directories found in this project. Nothing to clean.")
	} else {
		fmt.Println("No entries found in owned resource directories. Nothing to clean.")
	}

	if len(result.Failed) > 0 {
		fmt.Println("\nFailed removals:")
		for _, failure := range result.Failed {
			fmt.Printf("  - %s (%s): %s\n", failure.Path, failure.EntryType, failure.Error)
		}
	}

	fmt.Printf("\nSummary: owned dirs detected=%d, existing=%d, removed=%d (files=%d, symlinks=%d, directories=%d), failures=%d\n",
		result.Summary.OwnedDirsDetected,
		result.Summary.OwnedDirsExisting,
		result.Summary.Removed,
		result.Summary.RemovedFiles,
		result.Summary.RemovedSymlinks,
		result.Summary.RemovedDirs,
		result.Summary.Failed,
	)

	if result.Summary.OwnedDirsDetected > 0 {
		fmt.Println("To restore declared resources from ai.package.yaml, run: aimgr repair")
	}

	return nil
}

func printCleanWarnings(warnings []string) {
	for _, warning := range warnings {
		fmt.Fprintln(os.Stderr, warning)
	}
}

var (
	cleanProjectPath string
	cleanFormatFlag  string
)

func init() {
	rootCmd.AddCommand(cleanCmd)
	cleanCmd.Flags().StringVar(&cleanProjectPath, "project-path", "", "Project directory path (default: current directory)")
	cleanCmd.Flags().StringVar(&cleanFormatFlag, "format", "table", "Output format (table|json)")
	_ = cleanCmd.RegisterFlagCompletionFunc("format", completeTableJSONFormat)
}

func completeTableJSONFormat(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"table", "json"}, cobra.ShellCompDirectiveNoFileComp
}
