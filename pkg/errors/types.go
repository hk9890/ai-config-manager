package errors

import (
	"errors"
	"fmt"
)

// ErrorCategory represents the category of an error
type ErrorCategory int

const (
	// CategoryFatal represents fatal errors that should stop processing immediately
	// Examples: internal bugs, system failures, out of memory
	CategoryFatal ErrorCategory = iota

	// CategoryValidation represents validation errors that should be collected and processing should continue
	// Examples: invalid YAML, missing fields, already exists, invalid name format
	CategoryValidation

	// CategoryResource represents resource access errors that should be collected and processing should continue
	// Examples: file not found, permission denied, network timeout
	CategoryResource
)

// String returns the string representation of an error category
func (c ErrorCategory) String() string {
	switch c {
	case CategoryFatal:
		return "fatal"
	case CategoryValidation:
		return "validation"
	case CategoryResource:
		return "resource"
	default:
		return "unknown"
	}
}

// TypedError is an error with a category
type TypedError struct {
	Category ErrorCategory
	Err      error
	Context  string
}

// Error implements the error interface
func (e *TypedError) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("%s: %v", e.Context, e.Err)
	}
	return e.Err.Error()
}

// Unwrap returns the wrapped error
func (e *TypedError) Unwrap() error {
	return e.Err
}

// Fatal creates a fatal error
func Fatal(err error, context string) error {
	return &TypedError{
		Category: CategoryFatal,
		Err:      err,
		Context:  context,
	}
}

// Validation creates a validation error
func Validation(err error, context string) error {
	return &TypedError{
		Category: CategoryValidation,
		Err:      err,
		Context:  context,
	}
}

// Resource creates a resource error
func Resource(err error, context string) error {
	return &TypedError{
		Category: CategoryResource,
		Err:      err,
		Context:  context,
	}
}

// GetCategory returns the category of an error
// Returns CategoryValidation if the error is not a TypedError
func GetCategory(err error) ErrorCategory {
	if err == nil {
		return CategoryValidation // Shouldn't happen, but safe default
	}

	var te *TypedError
	if errors.As(err, &te) {
		return te.Category
	}

	// Default to validation for untyped errors (safest for bulk operations)
	return CategoryValidation
}

// IsFatal returns true if the error is a fatal error
func IsFatal(err error) bool {
	return GetCategory(err) == CategoryFatal
}

// IsValidation returns true if the error is a validation error
func IsValidation(err error) bool {
	return GetCategory(err) == CategoryValidation
}

// IsResource returns true if the error is a resource error
func IsResource(err error) bool {
	return GetCategory(err) == CategoryResource
}
