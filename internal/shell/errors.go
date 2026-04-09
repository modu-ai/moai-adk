package shell

import "errors"

var (
	// ErrPermissionDenied is returned when permission is denied to modify the config file.
	ErrPermissionDenied = errors.New("shell: permission denied")

	// ErrAlreadyConfigured is returned when the configuration already exists.
	ErrAlreadyConfigured = errors.New("shell: already configured")
)
