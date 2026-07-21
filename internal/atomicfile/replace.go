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

// Replace renames oldpath onto newpath, replacing newpath if it exists.
//
// The returned error is the underlying *os.LinkError from os.Rename, so
// callers may keep matching on os.IsNotExist / errors.Is as before.
func Replace(oldpath, newpath string) error {
	return replace(oldpath, newpath)
}
