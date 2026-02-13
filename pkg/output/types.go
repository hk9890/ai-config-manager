package output

// Renderable interface for custom complex layouts
type Renderable interface {
	Render(format Format) error
}

// TableOptions configures table rendering behavior
type TableOptions struct {
	ShowBorders      bool
	AutoWrap         bool
	Responsive       bool  // Enable terminal-aware sizing
	TerminalWidth    int   // Explicit terminal width (0 = auto-detect)
	DynamicColumn    int   // Index of column to stretch (-1 = rightmost visible)
	MinColumnWidths  []int // Minimum width for each column (0 = use default)
	MinTerminalWidth int   // Minimum terminal width for responsive mode (default: 15)
}
