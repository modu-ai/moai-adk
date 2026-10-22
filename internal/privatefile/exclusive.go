package privatefile

import (
	"fmt"
	"os"
)

// CreateExclusive atomically creates a new owner-only file, open for reading
// and writing. Existing names, including links, are never replaced. Windows
// retains READ_DATA access without DELETE sharing until the caller closes it.
// The callback owns the file immediately, before validation can fail. Callers
// must close every acquired file exactly once, including on validation errors.
func CreateExclusive(path string, acquired func(*os.File)) error {
	if acquired == nil {
		return fmt.Errorf("private exclusive file: owner callback is required")
	}
	return createExclusiveNative(path, acquired)
}
