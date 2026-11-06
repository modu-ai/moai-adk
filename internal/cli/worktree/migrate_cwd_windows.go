//go:build windows

// migrate_cwd_windows.go — the migration's process-cwd scan on Windows (card
// t2075). There is no lsof and no supported call that lists every process's
// working directory. Windows refuses to rename a directory that a process holds
// as its working directory, so the move step's own failure handling is the
// guard: the scan reports errProcessCWDOSGuarded instead of an error that would
// skip every tree. Never return an empty list here — an empty list reads as
// "no process inside".

package worktree

// defaultActiveProcessCWDs reports that the operating system guards the move.
func defaultActiveProcessCWDs() ([]string, error) {
	return nil, errProcessCWDOSGuarded
}
