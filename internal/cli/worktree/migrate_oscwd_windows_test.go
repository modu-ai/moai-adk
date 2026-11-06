//go:build windows

// migrate_oscwd_windows_test.go — cards t2075 and t2094: the premise behind
// the Windows probe and the scan's contract. The migration no longer skips
// trees for want of a cwd scan because the operating system refuses to rename
// a directory a process holds as its working directory; t2094 replaces the
// always-guarded stub with a real PEB scan, so the contract tests pin the
// never-empty guarantee and the premise tests measure, on the CI Windows
// runner, which shells the OS guard actually protects. These tests were not
// run on the authoring host; each assertion matches the documented premise
// analysis (see card t2094's report), not a guess.

package worktree

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestWindowsProbeContract — the t2094 scan contract: total failure (not a
// single working directory observable) reports errProcessCWDOSGuarded so the
// plan never skips every tree; success returns a non-empty list. The one
// forbidden shape is an empty list with a nil error, which would read as "no
// process inside".
func TestWindowsProbeContract(t *testing.T) {
	dirs, err := defaultActiveProcessCWDs()
	if err != nil {
		if !errors.Is(err, errProcessCWDOSGuarded) {
			t.Fatalf("err = %v, want errProcessCWDOSGuarded on total scan failure", err)
		}
		if len(dirs) != 0 {
			t.Fatalf("dirs = %v, want none beside the total-failure error", dirs)
		}
		return
	}
	if len(dirs) == 0 {
		t.Fatal("the scan returned an empty list with a nil error: an empty list reads as no process inside")
	}
}

// TestWindowsProbeObservesAChildWorkingDirectory — the scan reads a real
// process's working directory out of its PEB: a child settled inside a fresh
// directory must appear in the returned set.
func TestWindowsProbeObservesAChildWorkingDirectory(t *testing.T) {
	inside := t.TempDir()
	child := exec.Command("ping", "-n", "60", "127.0.0.1")
	child.Dir = inside
	if err := child.Start(); err != nil {
		t.Fatalf("start a process inside the tree: %v", err)
	}
	t.Cleanup(func() {
		_ = child.Process.Kill()
		_ = child.Wait()
	})
	time.Sleep(500 * time.Millisecond) // let the child settle in its directory

	dirs, err := defaultActiveProcessCWDs()
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	for _, d := range dirs {
		if strings.EqualFold(filepath.Clean(d), filepath.Clean(inside)) {
			return
		}
	}
	t.Errorf("the scan does not report a process working in %s:\n%v", inside, dirs)
}

// TestWindowsRefusesRenamingAProcessWorkingDirectory — t2075's premise for
// every process that holds the Win32 working directory (a plain native
// process here): the rename is refused.
func TestWindowsRefusesRenamingAProcessWorkingDirectory(t *testing.T) {
	tree := filepath.Join(t.TempDir(), "tree")
	if err := os.MkdirAll(tree, 0o700); err != nil {
		t.Fatal(err)
	}
	child := exec.Command("ping", "-n", "30", "127.0.0.1")
	child.Dir = tree
	if err := child.Start(); err != nil {
		t.Fatalf("start a process inside the tree: %v", err)
	}
	t.Cleanup(func() {
		_ = child.Process.Kill()
		_ = child.Wait()
	})
	time.Sleep(500 * time.Millisecond) // let the child settle in its directory

	moved := tree + "-moved"
	if err := os.Rename(tree, moved); err == nil {
		t.Fatalf("renaming a directory that a process holds as its working directory succeeded; the OS guard the migration relies on does not hold")
	}
	if _, err := os.Stat(tree); err != nil {
		t.Fatalf("the tree must stay in place after the refused rename: %v", err)
	}
}

// t2094WaitForMarker polls for a file the child writes once it has settled in
// the tree, so the rename attempt below it happens after the cd took effect.
func t2094WaitForMarker(t *testing.T, marker string) {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for {
		if _, err := os.Stat(marker); err == nil {
			time.Sleep(200 * time.Millisecond) // let the cd settle past the marker write
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("the child never wrote %s", marker)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// TestWindowsPowerShellSetLocationDoesNotHoldTheWin32CWD — premise analysis
// (card t2094, F2): PowerShell's Set-Location updates only the provider
// location. It does not call the Win32 SetCurrentDirectory, so the process's
// Win32 working directory — the one the kernel holds in the per-process DOS
// device map, the one the rename guard and the PEB scan both see — stays where
// the process was launched. This is the same mechanism as the documented .NET
// gotcha that [Environment]::CurrentDirectory does not follow cd. The test
// asserts the investigated behavior: the marker records the Win32 cwd staying
// at the launch directory while the provider location is the tree, and the
// rename SUCCEEDS — neither the OS guard nor the scan can see a PowerShell
// Set-Location, which is why this shell is a documented residual gap.
func TestWindowsPowerShellSetLocationDoesNotHoldTheWin32CWD(t *testing.T) {
	if _, err := exec.LookPath("powershell.exe"); err != nil {
		t.Skip("no powershell.exe on PATH")
	}
	outside := t.TempDir()
	tree := filepath.Join(t.TempDir(), "tree")
	if err := os.MkdirAll(tree, 0o700); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(outside, "cwds.txt")
	script := fmt.Sprintf(
		"Set-Location -LiteralPath '%s'; \"prov=$($PWD.Path)`nwin32=[Environment]::CurrentDirectory\" | Set-Content -LiteralPath '%s'; Start-Sleep -Seconds 60",
		tree, marker)
	child := exec.Command("powershell.exe", "-NoProfile", "-Command", script)
	child.Dir = outside
	if err := child.Start(); err != nil {
		t.Fatalf("start powershell outside the tree: %v", err)
	}
	t.Cleanup(func() {
		_ = child.Process.Kill()
		_ = child.Wait()
	})
	t2094WaitForMarker(t, marker)

	body, err := os.ReadFile(marker)
	if err != nil {
		t.Fatalf("read marker: %v", err)
	}
	win32 := ""
	for _, line := range strings.Split(string(body), "\n") {
		if v, ok := strings.CutPrefix(strings.TrimSpace(line), "win32="); ok {
			win32 = v
		}
	}
	if want := filepath.Clean(outside); !strings.EqualFold(filepath.Clean(win32), want) {
		t.Errorf("powershell's Win32 cwd = %q, want the launch directory %q (Set-Location must not move it)", win32, want)
	}

	moved := tree + "-moved"
	if err := os.Rename(tree, moved); err != nil {
		t.Fatalf("rename failed (%v); the premise analysis says a PowerShell Set-Location does not hold the Win32 cwd and must not block it", err)
	}
	_ = os.Rename(moved, tree) // put the tree back for the cleanup removal
}

// TestWindowsCmdCDHoldsTheWin32CWD — premise analysis (card t2094, F2):
// cmd.exe's cd calls the Win32 SetCurrentDirectory, so the process's Win32
// working directory follows into the tree and the rename must be refused.
func TestWindowsCmdCDHoldsTheWin32CWD(t *testing.T) {
	outside := t.TempDir()
	tree := filepath.Join(t.TempDir(), "tree")
	if err := os.MkdirAll(tree, 0o700); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(outside, "cwd.txt")
	child := exec.Command("cmd.exe", "/c",
		fmt.Sprintf("cd /d \"%s\" && cd > \"%s\" && ping -n 60 127.0.0.1 > nul", tree, marker))
	child.Dir = outside
	if err := child.Start(); err != nil {
		t.Fatalf("start cmd outside the tree: %v", err)
	}
	t.Cleanup(func() {
		_ = child.Process.Kill()
		_ = child.Wait()
	})
	t2094WaitForMarker(t, marker)

	if err := os.Rename(tree, tree+"-moved"); err == nil {
		t.Fatal("renaming a directory cmd.exe cd'd into succeeded; the premise analysis says cmd keeps the Win32 cwd in sync")
	}
}

// TestWindowsMSYSBashCDHoldsTheWin32CWD — premise analysis (card t2094, F2):
// Git-for-Windows/MSYS2 bash tracks its own logical cwd, but msys-2.0.dll's
// chdir keeps the process's Win32 working directory synchronized (it calls
// SetCurrentDirectory, which native children and Win32 path resolution rely
// on), so the rename must be refused. Only bash builtins run inside: an msys
// child like sleep would hold the Win32 cwd itself and mask what this test
// measures.
func TestWindowsMSYSBashCDHoldsTheWin32CWD(t *testing.T) {
	bash, err := exec.LookPath("bash.exe")
	if err != nil {
		t.Skip("no bash.exe on PATH (Git for Windows absent)")
	}
	outside := t.TempDir()
	tree := filepath.Join(t.TempDir(), "tree")
	if err := os.MkdirAll(tree, 0o700); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(outside, "cwd.txt")
	// read -t is a builtin: bash blocks in-process for 60s without spawning.
	script := fmt.Sprintf("cd '%s' && printf '%%s\\n' \"$PWD\" > '%s' && IFS= read -t 60 line",
		filepath.ToSlash(tree), filepath.ToSlash(marker))
	child := exec.Command(bash, "-c", script)
	child.Dir = outside
	if err := child.Start(); err != nil {
		t.Fatalf("start bash outside the tree: %v", err)
	}
	t.Cleanup(func() {
		_ = child.Process.Kill()
		_ = child.Wait()
	})
	t2094WaitForMarker(t, marker)

	if err := os.Rename(tree, tree+"-moved"); err == nil {
		t.Fatal("renaming a directory bash cd'd into succeeded; the premise analysis says msys keeps the Win32 cwd in sync")
	}
}
