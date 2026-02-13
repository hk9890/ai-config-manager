package output

// Renderable interface for custom complex layouts
type Renderable interface {
	Render(format Format) error
}

// TableOptions configures table rendering behavior
type TableOptions struct {
	ShowBorders   bool
	AutoWrap      bool
	Responsive    bool // Enable terminal-aware sizing
	TerminalWidth int  // Explicit terminal width (0 = auto-detect)
}
