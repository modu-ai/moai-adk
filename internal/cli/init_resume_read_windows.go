//go:build windows

package cli

// init_resume_read_windows.go — the windows half of the resume path's
// manifest read: the unix FIFO open-hang hazard has no windows-path
// equivalent, so Lstat + ReadFile suffices. The type judgment still fails
// closed on non-regular files.

import (
	"os"
)

func readManifestRecord(path string) (data []byte, ok bool) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, false
	}
	if !info.Mode().IsRegular() {
		return nil, false
	}
	data, readErr := os.ReadFile(path)
	if readErr != nil {
		return nil, false
	}
	return data, true
}
