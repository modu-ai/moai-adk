package shell

import "errors"

var (
	// ErrUnsupportedShell is returned when the shell type is not supported.
	ErrUnsupportedShell = errors.New("shell: unsupported shell type")

	// ErrConfigNotFound is returned when the config file cannot be found.
	ErrConfigNotFound = errors.New("shell: config file not found")

	// ErrPermissionDenied is returned when permission is denied to modify the config file.
	ErrPermissionDenied = errors.New("shell: permission denied")

	// ErrAlreadyConfigured is returned when the configuration already exists.
	ErrAlreadyConfigured = errors.New("shell: already configured")
)
