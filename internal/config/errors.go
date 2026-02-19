// Package config provides configuration management for MoAI-ADK Go Edition.
// It loads YAML section files, applies defaults, validates, and provides
// thread-safe access to configuration values.
package config

import (
	"errors"
	"fmt"
	"strings"
)

// Sentinel errors for configuration operations.
var (
	// ErrConfigNotFound indicates the configuration directory was not found.
	ErrConfigNotFound = errors.New("config: configuration directory not found")

	// ErrInvalidConfig indicates the configuration is invalid.
	ErrInvalidConfig = errors.New("config: invalid configuration")

	// ErrSectionNotFound indicates the requested section does not exist.
	ErrSectionNotFound = errors.New("config: section not found")

	// ErrInvalidDevelopmentMode indicates an invalid development mode value.
	ErrInvalidDevelopmentMode = errors.New("config: invalid development_mode, must be one of: ddd, tdd")

	// ErrNotInitialized indicates the ConfigManager has not been initialized via Load().
	ErrNotInitialized = errors.New("config: manager not initialized, call Load() first")

	// ErrSectionTypeMismatch indicates the section value type does not match expected type.
	ErrSectionTypeMismatch = errors.New("config: section type mismatch")

	// ErrDynamicToken indicates an unexpanded dynamic token was detected in a config value.
	ErrDynamicToken = errors.New("config: unexpanded dynamic token detected")

	// ErrInvalidYAML indicates invalid YAML syntax in a configuration file.
	ErrInvalidYAML = errors.New("config: invalid YAML syntax")
)

// ValidationError represents a single validation error with field context.
type ValidationError struct {
	Field   string
	Message string
	Value   any
	Wrapped error // underlying sentinel error for errors.Is support
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if e.Value != nil {
		return fmt.Sprintf("validation error: field %q: %s (got: %v)", e.Field, e.Message, e.Value)
	}
	return fmt.Sprintf("validation error: field %q: %s", e.Field, e.Message)
}

// Unwrap returns the underlying sentinel error.
func (e *ValidationError) Unwrap() error {
	return e.Wrapped
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors struct {
	Errors []ValidationError
}

// Error implements the error interface.
func (e *ValidationErrors) Error() string {
	if len(e.Errors) == 0 {
		return "validation: no errors"
	}
	msgs := make([]string, len(e.Errors))
	for i, err := range e.Errors {
		msgs[i] = err.Error()
	}
	return fmt.Sprintf("validation failed with %d error(s): %s", len(e.Errors), strings.Join(msgs, "; "))
}

// Is supports errors.Is by checking contained validation errors against the target.
func (e *ValidationErrors) Is(target error) bool {
	if target == ErrInvalidConfig {
		return true
	}
	for _, ve := range e.Errors {
		if ve.Wrapped != nil && errors.Is(ve.Wrapped, target) {
			return true
		}
	}
	return false
}
