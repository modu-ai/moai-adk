// Package project provides project initialization, detection, and validation
// for the MoAI-ADK Go Edition. It implements the core domain logic for the
// "moai init" CLI command, including language/framework detection, methodology
// recommendation, directory scaffolding, and configuration generation.
package project

import "errors"

// Sentinel errors for the project package.
var (
	// ErrProjectExists indicates the project root already contains a .moai/ directory.
	ErrProjectExists = errors.New("project already initialized")

	// ErrNoLanguageFound indicates no programming language was detected in the project.
	ErrNoLanguageFound = errors.New("no programming language detected")

	// ErrInvalidRoot indicates the given project root path is invalid or inaccessible.
	ErrInvalidRoot = errors.New("invalid project root path")

	// ErrInitFailed indicates a project initialization step failed.
	ErrInitFailed = errors.New("initialization failed")

	// ErrInvalidDevelopmentMode indicates an unrecognized development mode value.
	ErrInvalidDevelopmentMode = errors.New("invalid development mode: must be ddd or tdd")

	// ErrMethodologyDetection indicates methodology detection failed.
	ErrMethodologyDetection = errors.New("methodology detection failed")
)
