// regular_file.go — card t1632 (SPEC-MCP-NEUTRALITY-001, audit N6/N9): the
// shared reader for a path a caller supplied (or a stored caller path).
//
// A path-based check ("Lstat says regular") followed by a second open BY NAME
// is two path resolutions: the path can be replaced between them, so a symlink
// makes its target the evidence and a FIFO blocks the open before any size cap
// applies. openRegularFile checks the OPENED FILE instead; its two platform
// variants (regular_open_nofollow.go, regular_open_stat.go) differ only in how
// they refuse a link.
package runtime

import (
	"fmt"
	"io"
	"io/fs"
)

// beforeRegularOpen is a test seam: it runs with the path immediately before
// the file is opened, i.e. inside the window in which a path-based check could
// be raced. Production leaves it a no-op.
var beforeRegularOpen = func(path string) {}

// notRegularError is the one refusal text for a path that is not a regular
// file, whichever check noticed it.
func notRegularError(path string, mode fs.FileMode) error {
	return fmt.Errorf("%s is not a regular file (mode %s)", path, mode.Type())
}

// readRegularFile reads the whole of the regular file at path, refusing a file
// larger than maxBytes. A missing path returns the open error unchanged, so
// callers can still test it with os.IsNotExist.
func readRegularFile(path string, maxBytes int64) ([]byte, error) {
	f, err := openRegularFile(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	data, err := io.ReadAll(io.LimitReader(f, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("%s exceeds the %d-byte read limit", path, maxBytes)
	}
	return data, nil
}

// openRegularFile (defined per platform) returns path opened read-only, only
// when the OPENED file is a regular file; any other file is closed unread and
// refused with notRegularError.
