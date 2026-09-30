//go:build windows

// sweep_cwd_windows_test.go — AC-WS-007 cell (c): the Windows cwd-probe
// stub. Runs only on the Windows side of the CI matrix; the cross-platform
// build gate (GOOS=windows go build) covers the compile half.

package worktree

import "testing"

// TestSweepProcessCWDs_WindowsStubUnanswerable pins REQ-WS-006's Windows
// behavior: the probe reports UNANSWERABLE (an error), never an empty
// negative — an empty list here would read as "no process inside" and the
// sweep would dispose trees it could not judge.
func TestSweepProcessCWDs_WindowsStubUnanswerable(t *testing.T) {
	dirs, err := sweepProcessCWDs()
	if err == nil {
		t.Fatal("the Windows cwd probe must return an error (unanswerable), got nil")
	}
	if len(dirs) != 0 {
		t.Fatalf("the Windows cwd probe must return no directories, got %v", dirs)
	}
}
