//go:build windows

// proc_info_windows.go — ancestry lookup via a Toolhelp32 process-table
// snapshot.
package session

import (
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Snapshot retry bounds for CreateToolhelp32Snapshot's transient failure mode.
const (
	procInfoSnapshotAttempts   = 3
	procInfoSnapshotRetryDelay = 5 * time.Millisecond
)

// platformProcInfo reports the parent PID and executable name of pid by
// scanning a whole-process Toolhelp32 snapshot. A PID absent from the snapshot
// reports ok=false — that is how the ancestry walk detects a dead ancestor on
// Windows, where isProcessAlive is conservative (always true) and snapshot
// absence is the only dead signal the walk has.
//
// Cost: each call takes one fresh whole-table snapshot. The walk makes ~2
// calls per ancestor at depth <= maxAncestryDepth (8), which is acceptable for
// a hook subprocess; there is deliberately no caching layer.
func platformProcInfo(pid int) (ppid int, comm string, ok bool) {
	if pid <= 0 {
		return 0, "", false
	}
	snap, err := takeProcessSnapshot()
	if err != nil {
		return 0, "", false
	}
	// CloseHandle's error has no recovery path in a read-only lookup.
	defer func() { _ = windows.CloseHandle(snap) }()

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	if err := windows.Process32First(snap, &entry); err != nil {
		return 0, "", false
	}
	for {
		if int(entry.ProcessID) == pid {
			return int(entry.ParentProcessID),
				normalizeWindowsComm(windows.UTF16ToString(entry.ExeFile[:])), true
		}
		// ERROR_NO_MORE_FILES is the normal end of the scan: pid is not in
		// the table.
		if err := windows.Process32Next(snap, &entry); err != nil {
			return 0, "", false
		}
	}
}

// takeProcessSnapshot opens a snapshot of the whole process table.
//
// @MX:NOTE: [AUTO] while the table churns, CreateToolhelp32Snapshot can fail
// transiently with ERROR_BAD_LENGTH; the retry (3 attempts, 5ms apart) is
// safety rather than polish because a spurious snapshot failure is the
// dangerous direction for the integration lock — it makes a live holder's
// ancestry read as unresolvable, and the holder then look reclaimable.
func takeProcessSnapshot() (windows.Handle, error) {
	var lastErr error
	for attempt := 0; attempt < procInfoSnapshotAttempts; attempt++ {
		snap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
		if err == nil {
			return snap, nil
		}
		lastErr = err
		if err != windows.ERROR_BAD_LENGTH {
			break
		}
		time.Sleep(procInfoSnapshotRetryDelay)
	}
	return 0, lastErr
}

// normalizeWindowsComm lowercases the OS-reported executable name and strips
// one trailing ".exe" so it matches the lowercase bare-name keys of
// wrapperProcessNames ("Moai.EXE" -> "moai", "cmd.exe" -> "cmd"). Lowercasing
// first makes the suffix strip case-insensitive.
func normalizeWindowsComm(exe string) string {
	return strings.TrimSuffix(strings.ToLower(exe), ".exe")
}
