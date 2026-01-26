package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestFatal(t *testing.T) {
	baseErr := fmt.Errorf("out of memory")
	err := Fatal(baseErr, "failed to allocate buffer")

	if err == nil {
		t.Fatal("Fatal() returned nil")
	}

	te, ok := err.(*TypedError)
	if !ok {
		t.Fatal("Fatal() did not return *TypedError")
	}

	if te.Category != CategoryFatal {
		t.Errorf("Category = %v, want %v", te.Category, CategoryFatal)
	}

	if te.Context != "failed to allocate buffer" {
		t.Errorf("Context = %q, want %q", te.Context, "failed to allocate buffer")
	}

	if !errors.Is(err, baseErr) {
		t.Error("Fatal() error does not wrap original error")
	}
}

func TestValidation(t *testing.T) {
	baseErr := fmt.Errorf("missing required field")
	err := Validation(baseErr, "invalid YAML frontmatter")

	if err == nil {
		t.Fatal("Validation() returned nil")
	}

	te, ok := err.(*TypedError)
	if !ok {
		t.Fatal("Validation() did not return *TypedError")
	}

	if te.Category != CategoryValidation {
		t.Errorf("Category = %v, want %v", te.Category, CategoryValidation)
	}

	if te.Context != "invalid YAML frontmatter" {
		t.Errorf("Context = %q, want %q", te.Context, "invalid YAML frontmatter")
	}

	if !errors.Is(err, baseErr) {
		t.Error("Validation() error does not wrap original error")
	}
}

func TestResource(t *testing.T) {
	baseErr := fmt.Errorf("file not found")
	err := Resource(baseErr, "failed to read skill")

	if err == nil {
		t.Fatal("Resource() returned nil")
	}

	te, ok := err.(*TypedError)
	if !ok {
		t.Fatal("Resource() did not return *TypedError")
	}

	if te.Category != CategoryResource {
		t.Errorf("Category = %v, want %v", te.Category, CategoryResource)
	}

	if te.Context != "failed to read skill" {
		t.Errorf("Context = %q, want %q", te.Context, "failed to read skill")
	}

	if !errors.Is(err, baseErr) {
		t.Error("Resource() error does not wrap original error")
	}
}

func TestGetCategory(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCat  ErrorCategory
	}{
		{
			name:    "fatal error",
			err:     Fatal(errors.New("test"), "context"),
			wantCat: CategoryFatal,
		},
		{
			name:    "validation error",
			err:     Validation(errors.New("test"), "context"),
			wantCat: CategoryValidation,
		},
		{
			name:    "resource error",
			err:     Resource(errors.New("test"), "context"),
			wantCat: CategoryResource,
		},
		{
			name:    "untyped error defaults to validation",
			err:     errors.New("plain error"),
			wantCat: CategoryValidation,
		},
		{
			name:    "wrapped untyped error defaults to validation",
			err:     fmt.Errorf("wrapper: %w", errors.New("plain")),
			wantCat: CategoryValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetCategory(tt.err)
			if got != tt.wantCat {
				t.Errorf("GetCategory() = %v, want %v", got, tt.wantCat)
			}
		})
	}
}

func TestIsFatal(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "fatal error",
			err:  Fatal(errors.New("test"), "context"),
			want: true,
		},
		{
			name: "validation error",
			err:  Validation(errors.New("test"), "context"),
			want: false,
		},
		{
			name: "resource error",
			err:  Resource(errors.New("test"), "context"),
			want: false,
		},
		{
			name: "untyped error",
			err:  errors.New("plain"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsFatal(tt.err)
			if got != tt.want {
				t.Errorf("IsFatal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidation(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "validation error",
			err:  Validation(errors.New("test"), "context"),
			want: true,
		},
		{
			name: "fatal error",
			err:  Fatal(errors.New("test"), "context"),
			want: false,
		},
		{
			name: "resource error",
			err:  Resource(errors.New("test"), "context"),
			want: false,
		},
		{
			name: "untyped error defaults to validation",
			err:  errors.New("plain"),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidation(tt.err)
			if got != tt.want {
				t.Errorf("IsValidation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsResource(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "resource error",
			err:  Resource(errors.New("test"), "context"),
			want: true,
		},
		{
			name: "validation error",
			err:  Validation(errors.New("test"), "context"),
			want: false,
		},
		{
			name: "fatal error",
			err:  Fatal(errors.New("test"), "context"),
			want: false,
		},
		{
			name: "untyped error",
			err:  errors.New("plain"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsResource(tt.err)
			if got != tt.want {
				t.Errorf("IsResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTypedErrorMessage(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		want    string
	}{
		{
			name: "with context",
			err:  Validation(errors.New("missing field"), "invalid YAML"),
			want: "invalid YAML: missing field",
		},
		{
			name: "without context",
			err:  &TypedError{Category: CategoryValidation, Err: errors.New("error message"), Context: ""},
			want: "error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTypedErrorUnwrap(t *testing.T) {
	baseErr := errors.New("base error")
	wrappedErr := Validation(baseErr, "context")

	unwrapped := errors.Unwrap(wrappedErr)
	if unwrapped != baseErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, baseErr)
	}

	// Test errors.Is works
	if !errors.Is(wrappedErr, baseErr) {
		t.Error("errors.Is() should work with TypedError")
	}
}

func TestCategoryString(t *testing.T) {
	tests := []struct {
		category ErrorCategory
		want     string
	}{
		{CategoryFatal, "fatal"},
		{CategoryValidation, "validation"},
		{CategoryResource, "resource"},
		{ErrorCategory(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.category.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWrappedTypedErrors(t *testing.T) {
	// Test that we can detect categories even when wrapped with fmt.Errorf
	baseValidation := Validation(errors.New("invalid field"), "")
	wrappedOnce := fmt.Errorf("first wrap: %w", baseValidation)
	wrappedTwice := fmt.Errorf("second wrap: %w", wrappedOnce)

	if !IsValidation(wrappedTwice) {
		t.Error("IsValidation() should work through multiple fmt.Errorf wraps")
	}

	if GetCategory(wrappedTwice) != CategoryValidation {
		t.Error("GetCategory() should work through multiple fmt.Errorf wraps")
	}
}
