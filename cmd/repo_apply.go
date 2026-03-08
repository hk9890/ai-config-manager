package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dynatrace-oss/ai-config-manager/v3/pkg/repomanifest"
)

var (
	repoApplyDryRunFlag      bool
	repoApplyIncludeModeFlag string
)

// repoApplyCmd represents the apply command.
var repoApplyCmd = &cobra.Command{
	Use:   "apply <path-or-url>",
	Short: "Apply and merge sources from a shared ai.repo.yaml",
	Long: `Load and merge sources from a shared ai.repo.yaml into the local repository manifest.

Accepted inputs in v1:
  - Local path to ai.repo.yaml
  - HTTP(S) URL that points directly to ai.repo.yaml

Merge behavior:
  - New source name: added
  - Same source name + identical definition: no-op
  - Same source name + different location: conflict (never overwritten)
  - Same location with different include filters: replace or preserve (configurable)

Fresh repository behavior:
  - apply auto-initializes the local repository (same bootstrap as repo init)
  - in --dry-run mode, bootstrap is previewed and not persisted

Relationship to repo init:
  - repo init bootstraps local repository structure only
  - repo apply <path-or-url> bootstraps if needed, then merges shared sources`,
	Example: `  # Apply a manifest from local disk
  aimgr repo apply ./ai.repo.yaml
  aimgr repo apply /tmp/team/ai.repo.yaml

  # Apply a shared manifest from URL
  aimgr repo apply https://example.com/platform/ai.repo.yaml

  # Preview merge actions without writing
  aimgr repo apply ./ai.repo.yaml --dry-run

  # Preserve existing include filters when source location matches
  aimgr repo apply ./ai.repo.yaml --include-mode preserve`,
	Args: cobra.ExactArgs(1),
	RunE: runApply,
}

func runApply(cmd *cobra.Command, args []string) error {
	input := args[0]

	mgr, err := NewManagerWithLogLevel()
	if err != nil {
		return err
	}

	if !repoApplyDryRunFlag {
		if err := mgr.Init(); err != nil {
			return fmt.Errorf("failed to initialize repository for apply: %w", err)
		}
	}

	incoming, err := repomanifest.LoadForApply(input)
	if err != nil {
		return err
	}

	current, err := repomanifest.Load(mgr.GetRepoPath())
	if err != nil {
		return fmt.Errorf("failed to load local manifest: %w", err)
	}

	merged, report, err := repomanifest.MergeForApply(current, incoming, repomanifest.ApplyMergeOptions{
		IncludeMode: repomanifest.IncludeMergeMode(repoApplyIncludeModeFlag),
	})
	if err != nil {
		return err
	}

	printApplyReport(report, repoApplyDryRunFlag)

	if report.HasConflicts() {
		return fmt.Errorf("manifest apply has %d conflict(s); resolve conflicts and retry", report.Conflicts())
	}

	if repoApplyDryRunFlag {
		fmt.Println("\nDry-run complete: no changes were written")
		return nil
	}

	if report.Added() == 0 && report.Updated() == 0 {
		fmt.Println("\nNo changes to apply")
		return nil
	}

	if err := merged.Save(mgr.GetRepoPath()); err != nil {
		return fmt.Errorf("failed to save merged manifest: %w", err)
	}

	if err := mgr.CommitChanges("aimgr: apply manifest sources"); err != nil {
		fmt.Printf("Warning: Failed to commit manifest: %v\n", err)
	}

	fmt.Println("\n✓ Applied manifest successfully")

	return nil
}

func printApplyReport(report *repomanifest.ApplyMergeReport, dryRun bool) {
	header := "Manifest apply results"
	if dryRun {
		header = "Manifest apply dry-run results"
	}
	fmt.Println(header + ":")

	for _, change := range report.Changes {
		fmt.Printf("  - %s: %s (%s)\n", change.Action, change.Name, change.Message)
	}

	fmt.Printf("\nSummary: added=%d updated=%d noop=%d conflicts=%d\n",
		report.Added(), report.Updated(), report.NoOp(), report.Conflicts())
}

func init() {
	repoCmd.AddCommand(repoApplyCmd)

	repoApplyCmd.Flags().BoolVar(&repoApplyDryRunFlag, "dry-run", false, "Preview merge actions without writing ai.repo.yaml")
	repoApplyCmd.Flags().StringVar(&repoApplyIncludeModeFlag, "include-mode", string(repomanifest.IncludeMergeReplace), "Include handling for same-location sources: replace or preserve")
}
