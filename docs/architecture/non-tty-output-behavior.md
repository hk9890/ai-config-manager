# Non-TTY Output Behavior

**Status:** Approved Design  
**Created:** 2026-02-13  
**Related Issues:** ai-config-manager-aiv, ai-config-manager-6ju, ai-config-manager-815

## Overview

This document defines the expected behavior of `aimgr` output when running in non-TTY environments (piped, redirected, or CI/CD contexts).

## Design Decision

**Responsive mode is DISABLED for non-TTY contexts.**

When `IsTTY() == false`, tables use fixed column widths regardless of the `Responsive` setting. This ensures predictable, consistent output for scripts, pipelines, and automation.

## Rationale

### 1. Predictability for Scripts
Scripts that parse `aimgr` output need consistent formatting:
```bash
# These should produce identical table output
aimgr list > output.txt
aimgr list | tee output.txt
aimgr list | grep "skill"
```

If responsive mode activated in piped contexts, the output width would depend on the terminal where the pipe is displayed, creating inconsistent results.

### 2. CI/CD Compatibility
CI environments (GitHub Actions, GitLab CI, Jenkins) typically don't provide a TTY. Responsive mode would be meaningless and could cause issues:
- No terminal width available
- Logs displayed in various viewers with different widths
- Output needs to be copy-pasteable and readable in logs

### 3. Structured Output Isolation
JSON and YAML formats must be completely independent of terminal width:
```bash
# These MUST be identical regardless of terminal size
aimgr list --format=json
aimgr list --format=json | jq
aimgr list --format=json > data.json
```

Structured formats bypass table rendering entirely, so terminal width is irrelevant.

### 4. Separation of Concerns
- **Interactive use (TTY):** Responsive tables that adapt to user's terminal
- **Automation use (non-TTY):** Fixed-width tables with consistent layout
- **Structured output:** Terminal-independent JSON/YAML

## Behavior Specification

### TTY Detection

```go
func IsTTY() bool {
    return term.IsTerminal(int(os.Stdout.Fd()))
}
```

### Responsive Mode Activation

Responsive mode activates **ONLY** when **ALL** conditions are met:
1. `TableOptions.Responsive == true` (explicitly enabled)
2. `IsTTY() == true` (stdout is a terminal)
3. `Format == Table` (not JSON/YAML)

### Non-TTY Behavior

When `IsTTY() == false`:
- **Fixed Width:** Tables use default column widths (tablewriter defaults)
- **No Terminal Width Detection:** Terminal size queries are skipped
- **Consistent Output:** Same layout regardless of where output is displayed
- **Default Width:** tablewriter automatically wraps at reasonable defaults (~80-100 cols)

### Format Interactions

| Format | TTY Mode | Non-TTY Mode |
|--------|----------|--------------|
| `table` | Responsive sizing if enabled | Fixed width (tablewriter defaults) |
| `json`  | Terminal-independent | Terminal-independent |
| `yaml`  | Terminal-independent | Terminal-independent |

## Implementation Requirements

### Current Code (pkg/output/table.go)

```go
// renderTable renders TableData as a human-readable table
func renderTable(data *TableData) error {
    table := tablewriter.NewWriter(os.Stdout)
    
    // ... header setup ...
    
    // Responsive sizing only in TTY mode
    if data.Options.Responsive && IsTTY() {
        term, err := NewTerminal()
        if err == nil && term.Width() > 0 {
            // Apply responsive column sizing
        }
    }
    // Non-TTY: Use tablewriter defaults (no special handling)
    
    // ... render table ...
}
```

**Status:** ✅ Already implemented correctly (lines 94-101)

### Required Enhancements for ai-config-manager-6ju

When implementing responsive sizing (ai-config-manager-6ju), ensure:

```go
// renderTable renders TableData as a human-readable table
func renderTable(data *TableData) error {
    table := tablewriter.NewWriter(os.Stdout)
    
    // ... header setup ...
    
    // Responsive sizing ONLY when:
    // 1. Responsive option is enabled
    // 2. Running in a TTY
    if data.Options.Responsive && IsTTY() {
        term, err := NewTerminal()
        if err == nil && term.Width() > 0 {
            // Calculate responsive column widths
            columnWidths := calculateColumnWidths(term.Width(), data.Headers, data.Rows)
            
            // Apply widths using tablewriter v1.1.3 API
            table.WithColumnWidths(columnWidths)
            table.WithMaxWidth(term.Width())
            table.WithRowAutoWrap(tw.WrapTruncate)
        }
    }
    // Non-TTY: No width constraints, use tablewriter defaults
    
    // ... render table ...
}
```

### Terminal Width Defaults (NewTerminal)

```go
// NewTerminal creates a new Terminal with current dimensions
func NewTerminal() (*Terminal, error) {
    width, height, err := term.GetSize(int(os.Stdout.Fd()))
    if err != nil {
        // Not a TTY or unable to get size - use safe default
        return &Terminal{width: 80, height: 24}, nil
    }
    
    return &Terminal{width: width, height: height}, nil
}
```

**Status:** ✅ Already returns sensible defaults (line 20)

**Note:** These defaults are only used as fallback. In non-TTY contexts, responsive mode won't activate, so these defaults won't affect output.

## Testing Requirements (ai-config-manager-815)

### Non-TTY Test Cases

```go
func TestTableBuilder_NonTTY_FixedWidth(t *testing.T) {
    // Simulate non-TTY environment
    // Even with Responsive=true, table should use fixed widths
    
    table := NewTable("Name", "Type", "Description").
        WithResponsive() // Enable responsive (but won't activate)
    
    table.AddRow("test-command", "command", "A very long description that would normally be truncated in responsive mode")
    table.AddRow("test-skill", "skill", "Another long description")
    
    // Capture output
    output := captureTableOutput(table)
    
    // Verify:
    // 1. Output has consistent width (not dependent on terminal)
    // 2. Uses tablewriter default wrapping behavior
    // 3. Matches expected non-responsive layout
}

func TestTableBuilder_PipedOutput(t *testing.T) {
    // Simulate: aimgr list | grep "skill"
    // Responsive mode should NOT activate
    
    // Mock IsTTY() to return false
    // Verify table uses fixed widths
}

func TestTableBuilder_RedirectedOutput(t *testing.T) {
    // Simulate: aimgr list > output.txt
    // Responsive mode should NOT activate
    
    // Mock IsTTY() to return false
    // Verify output is consistent across runs
}

func TestTableBuilder_CIEnvironment(t *testing.T) {
    // Simulate CI environment (no TTY)
    // Verify responsive mode disabled
    
    // Mock IsTTY() to return false
    // Verify tables render with fixed widths
}

func TestTableBuilder_JSONFormat_TerminalIndependent(t *testing.T) {
    // Verify JSON output is identical at different terminal widths
    
    table := NewTable("Name", "Type").WithResponsive()
    table.AddRow("test", "command")
    
    // Capture JSON output at "different" terminal widths
    // (terminal width should have NO effect on JSON)
    output1 := captureJSONOutput(table, termWidth: 80)
    output2 := captureJSONOutput(table, termWidth: 120)
    output3 := captureJSONOutput(table, termWidth: 200)
    
    // Assert all outputs are identical
    assert.Equal(t, output1, output2)
    assert.Equal(t, output2, output3)
}

func TestTableBuilder_YAMLFormat_TerminalIndependent(t *testing.T) {
    // Same as JSON test, but for YAML format
}
```

### Test Helpers Needed

```go
// Mock IsTTY for testing
func mockIsTTY(isTTY bool) func() {
    // Implementation to override IsTTY() for tests
}

// Capture table output
func captureTableOutput(tb *TableBuilder) string {
    // Capture stdout during table.Format(Table)
}

// Capture JSON output (with mocked terminal width)
func captureJSONOutput(tb *TableBuilder, termWidth int) string {
    // Capture JSON output
}
```

## Use Case Examples

### Interactive Terminal (TTY)
```bash
$ aimgr list
# → Responsive mode active
# → Table adapts to terminal width
# → Description column stretches/shrinks
```

### Piped Output (non-TTY)
```bash
$ aimgr list | less
# → Responsive mode disabled
# → Fixed width tables
# → Consistent layout for scrolling
```

### Redirected Output (non-TTY)
```bash
$ aimgr list > output.txt
# → Responsive mode disabled
# → Fixed width tables
# → File has consistent format
```

### CI/CD Pipeline (non-TTY)
```bash
# GitHub Actions workflow
- run: aimgr list
# → Responsive mode disabled
# → Output appears in workflow logs
# → Readable and consistent
```

### Structured Output (always terminal-independent)
```bash
$ aimgr list --format=json | jq '.rows[] | select(.type == "skill")'
# → JSON output unaffected by terminal width
# → Same output in TTY and non-TTY

$ aimgr list --format=yaml > data.yaml
# → YAML output unaffected by terminal width
# → Structured format independent of display context
```

## Edge Cases

### Narrow Terminal (TTY)
```bash
# Terminal width: 60 columns
$ aimgr list
# → Responsive mode active
# → Columns may hide if too narrow
# → Text truncates with "..."
```

### Narrow Terminal (non-TTY)
```bash
# Terminal width: 60 columns (but piped)
$ aimgr list | cat
# → Responsive mode disabled
# → Uses fixed widths (may exceed 60 cols)
# → Output wraps in display viewer, not truncated
```

### No Terminal Size Available
```bash
# SSH session with broken TTY
$ aimgr list
# → term.GetSize() fails
# → NewTerminal() returns default (80x24)
# → Responsive mode may activate with default width
```

**Design Decision:** If TTY detection succeeds but size detection fails, use default 80-column width for responsive calculations. This is better than disabling responsive mode entirely.

## Documentation Updates Needed

### User Guide (docs/user-guide/output-formats.md)

Add section:

```markdown
### Terminal Output Behavior

#### Interactive Terminal (TTY)
When running `aimgr` in an interactive terminal, table output uses responsive sizing to adapt to your terminal width. Columns dynamically adjust to use available space.

#### Piped or Redirected Output (non-TTY)
When output is piped (`aimgr list | less`) or redirected (`aimgr list > file.txt`), responsive mode is automatically disabled. Tables use fixed widths for consistent, predictable output.

This ensures:
- Scripts get consistent table formatting
- CI/CD logs have readable output
- Redirected files have stable layouts

#### Structured Formats
JSON and YAML formats (`--format=json`, `--format=yaml`) are always independent of terminal width. They produce identical output in all contexts.
```

### CLI Help Text

Consider adding brief mention in relevant command help:

```go
// Example: aimgr list --help
Long: `List all installed resources in the current directory.

Table output adapts to terminal width in interactive mode. Piped or redirected
output uses fixed widths for consistency. Use --format=json for machine-readable output.`,
```

**Decision:** Help text update is optional (low priority). Most users won't need this detail.

## Summary

| Context | TTY? | Responsive? | Width |
|---------|------|-------------|-------|
| Interactive terminal | ✅ Yes | ✅ Yes (if enabled) | Dynamic (terminal width) |
| Piped output | ❌ No | ❌ No | Fixed (tablewriter defaults) |
| Redirected output | ❌ No | ❌ No | Fixed (tablewriter defaults) |
| CI/CD environment | ❌ No | ❌ No | Fixed (tablewriter defaults) |
| JSON format | N/A | N/A | N/A (structured) |
| YAML format | N/A | N/A | N/A (structured) |

**Key Principle:** Responsive mode is a TTY-only feature for interactive user experience. Non-TTY contexts prioritize predictability and consistency.
