// Package atomicfile provides a cross-platform atomic file-replacement
// primitive for the write-temp-then-rename idiom used across MoAI state files.
//
// On POSIX, rename(2) atomically replaces the destination even while other
// processes hold it open, so Replace is a direct os.Rename passthrough. On
// Windows the same call fails when any handle is still open on the
// destination — see replace_windows.go for the platform-specific handling.
//
// Callers keep owning temp-file creation and cleanup; Replace covers only the
// final rename step so it can be dropped into existing atomic writers without
// changing their error-handling shape.
package atomicfile

import "os"

// Replace renames oldpath onto newpath, replacing newpath if it exists.
//
// The returned error is the underlying *os.LinkError from os.Rename, so
// callers may keep matching on os.IsNotExist / errors.Is as before.
func Replace(oldpath, newpath string) error {
	return replace(oldpath, newpath)
}

// ReadFile reads path, absorbing the transient window in which a concurrent
// Replace makes the file briefly unopenable on Windows. On POSIX it is a plain
// os.ReadFile. A missing file returns immediately in both cases, so callers may
// keep matching on os.IsNotExist.
//
// Use it for any read that can race a writer of the same file; a read fully
// serialized against writers does not need it.
func ReadFile(path string) ([]byte, error) {
	return readFile(path)
}

// Claim exclusively creates path, returning an error wrapping fs.ErrExist when
// path already exists. Exactly one concurrent caller can succeed for a given
// path: O_CREATE|O_EXCL is atomic on POSIX and maps to CREATE_NEW on Windows,
// so the exclusivity is guaranteed by both platforms rather than inferred from
// a particular failure errno.
//
// Claim is deliberately the opposite of Replace, and the two must not be
// substituted for each other:
//
//   - Replace answers "make newpath be this content", so replacing an existing
//     destination is success, and on Windows it retries a transient sharing
//     violation until the destination is writable.
//   - Claim answers "am I the one who gets to proceed", so an existing path is
//     failure, and it NEVER retries — a retry would hand the same claim to a
//     second caller and destroy the mutual exclusion it exists to provide.
//
// Callers own removing the claim once the guarded work is done.
func Claim(path string, perm os.FileMode) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	return f.Close()
}
