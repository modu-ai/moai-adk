//go:build !windows

package core

// GetLongPathName is a no-op on non-Windows platforms.
func GetLongPathName(shortPath string) (string, error) {
	return shortPath, nil
}
