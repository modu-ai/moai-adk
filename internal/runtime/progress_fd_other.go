//go:build !darwin && !linux

package runtime

import "os"

// fdMatchesName cannot compare device/inode on this platform (the syscall
// Stat_t shape differs) — it reports a match so the replace proceeds; the
// temp-swap defense this provides on darwin/linux is a documented gap here
// (flagged for leader review alongside the windows seeder posture).
func fdMatchesName(f *os.File, name string) bool {
	_ = f
	_ = name
	return true
}
