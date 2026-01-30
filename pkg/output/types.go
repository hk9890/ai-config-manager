package output

// Renderable interface for custom complex layouts
type Renderable interface {
	Render(format Format) error
}

// TableOptions configures table rendering behavior
type TableOptions struct {
	ShowBorders bool
	AutoWrap    bool
}
