// Package template provides template deployment, rendering, settings generation,
// and deployment validation for MoAI-ADK.
//
// It implements ADR-011 (Zero Runtime Template Expansion) by generating
// configuration files through Go struct serialization rather than string
// concatenation, and ADR-003 (go:embed for Template Distribution) by
// bundling templates into the binary.
package template

import "errors"

// Sentinel errors for the template package.
var (
	// ErrTemplateNotFound indicates a template does not exist in the embedded filesystem.
	ErrTemplateNotFound = errors.New("template: not found in embedded filesystem")

	// ErrPathTraversal indicates an attempt to access a path outside the project root.
	ErrPathTraversal = errors.New("template: path traversal detected")

	// ErrInvalidJSON indicates generated JSON is not valid.
	ErrInvalidJSON = errors.New("template: generated JSON is invalid")

	// ErrUnexpandedToken indicates unexpanded dynamic tokens remain in output.
	ErrUnexpandedToken = errors.New("template: unexpanded dynamic token detected")

	// ErrMissingTemplateKey indicates a required key is missing from template data.
	ErrMissingTemplateKey = errors.New("template: missing key in template data")
)
