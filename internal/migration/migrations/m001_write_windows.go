//go:build windows

package migrations

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
)

// Native owner/group/DACL policy is observed and recreated before stage data.
// os.Stat's Windows mode bits cannot represent the original execute policy.
func writeM001Wrapper(path string, content []byte, mode os.FileMode) (result error) {
	f, err := atomicfile.CreatePolicyTemp(path, filepath.Base(path)+".*.tmp", mode.Perm(), 0)
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
