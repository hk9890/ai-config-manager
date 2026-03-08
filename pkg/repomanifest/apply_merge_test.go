package repomanifest

import (
	"strings"
	"testing"
)

func TestMergeForApply_AddAndIdempotent(t *testing.T) {
	current := &Manifest{Version: 1, Sources: []*Source{}}
	incoming := &Manifest{Version: 1, Sources: []*Source{{
		Name:    "team-tools",
		URL:     "https://github.com/example/tools",
		Include: []string{"skill/*"},
	}}}

	merged, report, err := MergeForApply(current, incoming, ApplyMergeOptions{})
	if err != nil {
		t.Fatalf("MergeForApply() error = %v", err)
	}
	if len(merged.Sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(merged.Sources))
	}
	if report.Added() != 1 || report.NoOp() != 0 || report.Conflicts() != 0 || report.Updated() != 0 {
		t.Fatalf("unexpected report counts: add=%d update=%d noop=%d conflict=%d", report.Added(), report.Updated(), report.NoOp(), report.Conflicts())
	}

	mergedAgain, reportAgain, err := MergeForApply(merged, incoming, ApplyMergeOptions{})
	if err != nil {
		t.Fatalf("MergeForApply() second run error = %v", err)
	}
	if len(mergedAgain.Sources) != 1 {
		t.Fatalf("expected 1 source after re-apply, got %d", len(mergedAgain.Sources))
	}
	if reportAgain.NoOp() != 1 || reportAgain.Added() != 0 || reportAgain.Updated() != 0 || reportAgain.Conflicts() != 0 {
		t.Fatalf("unexpected re-apply counts: add=%d update=%d noop=%d conflict=%d", reportAgain.Added(), reportAgain.Updated(), reportAgain.NoOp(), reportAgain.Conflicts())
	}
}

func TestMergeForApply_ConflictOnDifferentLocation(t *testing.T) {
	current := &Manifest{Version: 1, Sources: []*Source{{
		Name: "team-tools",
		URL:  "https://github.com/example/tools",
	}}}
	incoming := &Manifest{Version: 1, Sources: []*Source{{
		Name: "team-tools",
		URL:  "https://github.com/example/other-tools",
	}}}

	merged, report, err := MergeForApply(current, incoming, ApplyMergeOptions{})
	if err != nil {
		t.Fatalf("MergeForApply() error = %v", err)
	}
	if !report.HasConflicts() || report.Conflicts() != 1 {
		t.Fatalf("expected 1 conflict, got %d", report.Conflicts())
	}
	if len(merged.Sources) != 1 || merged.Sources[0].URL != "https://github.com/example/tools" {
		t.Fatalf("conflict must not overwrite existing source")
	}
}

func TestMergeForApply_IncludeReplaceAndPreserve(t *testing.T) {
	current := &Manifest{Version: 1, Sources: []*Source{{
		Name:    "team-tools",
		URL:     "https://github.com/example/tools",
		Include: []string{"skill/pdf*"},
	}}}
	incoming := &Manifest{Version: 1, Sources: []*Source{{
		Name:    "team-tools",
		URL:     "https://github.com/example/tools",
		Include: []string{"command/lint-*"},
	}}}

	mergedReplace, reportReplace, err := MergeForApply(current, incoming, ApplyMergeOptions{IncludeMode: IncludeMergeReplace})
	if err != nil {
		t.Fatalf("replace merge error = %v", err)
	}
	if reportReplace.Updated() != 1 {
		t.Fatalf("expected 1 update in replace mode, got %d", reportReplace.Updated())
	}
	if got := strings.Join(mergedReplace.Sources[0].Include, ","); got != "command/lint-*" {
		t.Fatalf("replace mode include mismatch: %s", got)
	}

	mergedPreserve, reportPreserve, err := MergeForApply(current, incoming, ApplyMergeOptions{IncludeMode: IncludeMergePreserve})
	if err != nil {
		t.Fatalf("preserve merge error = %v", err)
	}
	if reportPreserve.NoOp() != 1 || reportPreserve.Updated() != 0 {
		t.Fatalf("expected noop in preserve mode, got noop=%d update=%d", reportPreserve.NoOp(), reportPreserve.Updated())
	}
	if got := strings.Join(mergedPreserve.Sources[0].Include, ","); got != "skill/pdf*" {
		t.Fatalf("preserve mode include mismatch: %s", got)
	}
}

func TestMergeForApply_InvalidIncludeMode(t *testing.T) {
	current := &Manifest{Version: 1, Sources: []*Source{}}
	incoming := &Manifest{Version: 1, Sources: []*Source{}}

	_, _, err := MergeForApply(current, incoming, ApplyMergeOptions{IncludeMode: IncludeMergeMode("invalid")})
	if err == nil {
		t.Fatal("expected error for invalid include mode")
	}
	if !strings.Contains(err.Error(), "invalid include merge mode") {
		t.Fatalf("unexpected error: %v", err)
	}
}
