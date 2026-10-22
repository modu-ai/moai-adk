//go:build windows

package migrations

import (
	"fmt"
	"os"
)

// Retain the existing Windows publication behavior while its native policy
// tests run in official CI. The Unix umask repair does not implement that policy.
func writeM001Wrapper(path string, content []byte, mode os.FileMode) error {
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, content, mode); err != nil {
		return fmt.Errorf("wrapper 쓰기 실패 %s: %w", path, err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("wrapper rename 실패 %s: %w", path, err)
	}
	return nil
}
