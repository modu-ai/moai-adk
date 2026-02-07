package defs

import "os"

// Standard file-system permissions used during project scaffolding.
const (
	// DirPerm is the default permission for directories (rwxr-xr-x).
	DirPerm os.FileMode = 0o755

	// FilePerm is the default permission for regular files (rw-r--r--).
	FilePerm os.FileMode = 0o644

	// CredDirPerm is the permission for credential directories (rwx------).
	CredDirPerm os.FileMode = 0o700

	// CredFilePerm is the permission for credential files (rw-------).
	CredFilePerm os.FileMode = 0o600
)
