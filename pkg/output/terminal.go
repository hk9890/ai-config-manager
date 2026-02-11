package output

import (
	"os"

	"golang.org/x/term"
)

// Terminal provides terminal size information
type Terminal struct {
	width  int
	height int
}

// NewTerminal creates a new Terminal with current dimensions
func NewTerminal() (*Terminal, error) {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		// Not a TTY or unable to get size - use default
		return &Terminal{width: 80, height: 24}, nil
	}

	return &Terminal{width: width, height: height}, nil
}

// Width returns the terminal width in columns
func (t *Terminal) Width() int {
	return t.width
}

// Height returns the terminal height in rows
func (t *Terminal) Height() int {
	return t.height
}

// IsTTY checks if stdout is a terminal
func IsTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}
