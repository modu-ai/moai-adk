//go:build !darwin && !linux

package runtime

// hardLinked cannot detect shared inodes on this platform; the rename
// replace path applies (a hardlinked progress.md is not a supported
// posture here). Flagged alongside the windows seeder posture.
func hardLinked(path string) bool {
	return false
}
