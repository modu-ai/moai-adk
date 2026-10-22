//go:build !windows

package migrations

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// writeM001Wrapper stages bytes privately, then restores the original exact
// permissions on the same descriptor. A creation mode alone is umask-filtered.
func writeM001Wrapper(path string, content []byte, mode os.FileMode) (result error) {
	f, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("wrapper 쓰기 실패 %s: %w", path, err)
	}
	tempPath := f.Name()
	closed, published := false, false
	defer func() {
		if !closed {
			if err := f.Close(); err != nil {
				result = errors.Join(result, fmt.Errorf("wrapper close 실패 %s: %w", path, err))
			}
		}
		if !published {
			if err := os.Remove(tempPath); err != nil && !os.IsNotExist(err) {
				result = errors.Join(result, fmt.Errorf("wrapper temporary cleanup 실패 %s: %w", path, err))
			}
		}
	}()
	if _, err := f.Write(content); err != nil {
		return fmt.Errorf("wrapper 쓰기 실패 %s: %w", path, err)
	}
	if err := f.Chmod(mode.Perm()); err != nil {
		return fmt.Errorf("wrapper mode 보존 실패 %s: %w", path, err)
	}
	closed = true
	if err := f.Close(); err != nil {
		return fmt.Errorf("wrapper close 실패 %s: %w", path, err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("wrapper rename 실패 %s: %w", path, err)
	}
	published = true
	return nil
}
