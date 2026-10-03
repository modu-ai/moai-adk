package cli

// managed_codex_tui_logalias_test.go — card t1408: the session log open must
// not write through any existing directory entry (hard link, symlinked parent,
// symlinked .moai), because a cloned repository can track such links.

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func logAliasVictim(t *testing.T) (path string) {
	t.Helper()
	path = filepath.Join(t.TempDir(), "victim.txt")
	if err := os.WriteFile(path, []byte("keep\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func assertVictimUntouched(t *testing.T, path string) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil || string(b) != "keep\n" {
		t.Errorf("the outside file changed: content %q (err %v)", b, err)
	}
	if info, err := os.Stat(path); err != nil || info.Mode().Perm() != 0o644 {
		t.Errorf("the outside file's mode changed: %v (err %v)", info.Mode().Perm(), err)
	}
}

func TestManagedTUILogOpenRefusesAliases(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX links")
	}
	t.Run("hard_link_at_log_path", func(t *testing.T) {
		root := t.TempDir()
		logs := filepath.Join(root, ".moai", "logs")
		if err := os.MkdirAll(logs, 0o755); err != nil {
			t.Fatal(err)
		}
		victim := logAliasVictim(t)
		if err := os.Link(victim, filepath.Join(logs, "factory-managed-run1-lane-1.log")); err != nil {
			t.Skipf("hard links unavailable: %v", err)
		}
		f, _, err := openManagedTUILog(root, "run1", "lane-1")
		if err == nil {
			_, _ = f.WriteString("probe\n")
			_ = f.Close()
		}
		assertVictimUntouched(t, victim)
	})
	t.Run("symlinked_logs_dir", func(t *testing.T) {
		root := t.TempDir()
		outside := t.TempDir()
		if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, filepath.Join(root, ".moai", "logs")); err != nil {
			t.Fatal(err)
		}
		f, _, err := openManagedTUILog(root, "run2", "lane-1")
		if err == nil {
			_ = f.Close()
			t.Errorf("a symlinked .moai/logs was accepted")
		}
		if entries, _ := os.ReadDir(outside); len(entries) != 0 {
			t.Errorf("something was written outside the project through the symlinked logs dir: %d entries", len(entries))
		}
	})
	t.Run("symlinked_moai_dir", func(t *testing.T) {
		root := t.TempDir()
		outside := t.TempDir()
		if err := os.Symlink(outside, filepath.Join(root, ".moai")); err != nil {
			t.Fatal(err)
		}
		f, _, err := openManagedTUILog(root, "run3", "lane-1")
		if err == nil {
			_ = f.Close()
			t.Errorf("a symlinked .moai was accepted")
		}
		entries, _ := os.ReadDir(outside)
		if len(entries) != 0 {
			t.Errorf("something was created outside the project through the symlinked .moai: %d entries", len(entries))
		}
	})
}

// TestManagedTUILogRefusalFallsBackHeadless pins the refusal contract: an
// unopenable session log is not a session error; the owner prints the one
// "not attached" notice and continues headless (REQ-MT-004).
func TestManagedTUILogRefusalFallsBackHeadless(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX links")
	}
	f := newTUIFake(t, "tui-log-refused")
	logs := filepath.Join(f.root, ".moai", "logs")
	if err := os.RemoveAll(logs); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(logs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(t.TempDir(), logs); err != nil {
		t.Fatal(err)
	}
	sess := f.newSession()
	if sess.tui != nil {
		t.Errorf("a TUI was planned although the session log location is a symlink")
	}
	term := f.term.snapshot()
	if n := strings.Count(term, "operator TUI not attached (session log file unavailable"); n != 1 {
		t.Errorf("want exactly one fallback notice, got %d:\n%s", n, term)
	}
}
