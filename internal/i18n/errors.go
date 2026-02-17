// Package i18n provides multilingual comment generation for GitHub issue
// automation. It supports Korean, English, Japanese, and Chinese with
// automatic fallback to English for unsupported languages.
package i18n

import "errors"

// Sentinel errors for i18n operations.
// All errors can be checked with errors.Is().
var (
	// ErrTemplateNotFound indicates the requested template does not exist.
	ErrTemplateNotFound = errors.New("i18n: template not found")

	// ErrTemplateExecution indicates the template failed to render.
	ErrTemplateExecution = errors.New("i18n: template execution failed")

	// ErrInvalidData indicates the provided template data is invalid.
	ErrInvalidData = errors.New("i18n: invalid template data")
)
