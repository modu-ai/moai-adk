package atomicfile

import (
	"bytes"
	"os"
	"path/filepath"
)

// WriteFileIfChanged makes path hold data, touching the disk only when the
// bytes on disk differ. When path already holds exactly data it returns
// (false, nil) after a read and performs ZERO writes: mtime and inode stay as
// they were, so an idempotent re-record of a wait, verdict, or progress
// section leaves no trace. When the content differs (or path is absent) it
// writes a temp file beside path and Replaces it in, returning (true, nil).
//
// The mode of an existing destination is kept; perm applies to a new file.
// The write is a replace, not an in-place truncate, so a reader never sees a
// partial file.
func WriteFileIfChanged(path string, data []byte, perm os.FileMode) (changed bool, err error) {
	mode := perm
	if cur, rerr := ReadFile(path); rerr == nil {
		if bytes.Equal(cur, data) {
			return false, nil
		}
		if st, serr := os.Stat(path); serr == nil {
			mode = st.Mode().Perm()
		}
	} else if !os.IsNotExist(rerr) {
		return false, rerr
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".write_if_changed_*")
	if err != nil {
		return false, err
	}
	tmpName := tmp.Name()
	fail := func(err error) (bool, error) {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return false, err
	}
	if _, err := tmp.Write(data); err != nil {
		return fail(err)
	}
	if err := tmp.Chmod(mode); err != nil {
		return fail(err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return false, err
	}
	if err := Replace(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return false, err
	}
	return true, nil
}
