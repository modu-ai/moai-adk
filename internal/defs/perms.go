package defs

import "os"

// Standard file-system permissions used during project scaffolding.
const (
	// DirPerm is the default permission for directories (rwxr-xr-x).
	DirPerm os.FileMode = 0o755

	// FilePerm is the default permission for regular files (rw-r--r--).
	FilePerm os.FileMode = 0o644

	// PrivateFilePerm keeps unpublished payloads accessible only to their owner.
	PrivateFilePerm os.FileMode = 0o600
)
