//go:build !windows

package sqliteguard

import "os"

// @MX:ANCHOR: [AUTO] Preserve private Unix namespaces for managed state writers.
// @MX:REASON: Home/project layout, factory opening and profile lease opening share this boundary.
func EnsureDirectory(path string) error {
	if err := os.MkdirAll(path, 0o700); err != nil {
		return err
	}
	return os.Chmod(path, 0o700)
}

func SecureArtifact(path string) error { return os.Chmod(path, 0o600) }
