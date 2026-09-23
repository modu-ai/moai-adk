package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/codexwiring"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

func wiringSHA(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// wiredProject returns a project wired through `moai tool enable codex`.
func wiredProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	var out, warn bytes.Buffer
	if err := runToolEnableCodexAt(root, &out, &warn, false); err != nil {
		t.Fatalf("enable: %v (%s)", err, warn.String())
	}
	return root
}

const (
	stagedTemp = ".codexwiring-stagedentry"
	orphanTemp = ".codexwiring-orphanentry"
)

// interruptWiring leaves the project as a crash between staging the temp file
// and renaming it would: one staged journal entry whose temp file exists, plus
// one temp file no entry references.
func interruptWiring(t *testing.T, root string) {
	t.Helper()
	hooks := filepath.Join(root, codexwiring.HooksRelPath)
	cur, err := os.ReadFile(hooks)
	if err != nil {
		t.Fatal(err)
	}
	staged := []byte("{\"hooks\": {}}\n")
	entry := codexwiring.JournalEntry{
		ID: "interrupted", Op: codexwiring.OpWrite, Path: codexwiring.HooksRelPath,
		PreHash: wiringSHA(cur), PostHash: wiringSHA(staged), Temp: stagedTemp, State: codexwiring.JournalStaged,
	}
	raw, err := json.Marshal(map[string]any{"entries": []codexwiring.JournalEntry{entry}})
	if err != nil {
		t.Fatal(err)
	}
	for path, body := range map[string][]byte{
		filepath.Join(root, filepath.FromSlash(codexwiring.JournalRelPath)): raw,
		filepath.Join(root, ".codex", stagedTemp):                           staged,
		filepath.Join(root, ".codex", orphanTemp):                           []byte("left by an older binary"),
	} {
		if err := os.WriteFile(path, body, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if r := codexwiring.InspectJournal(root); len(r.Incomplete) != 1 || len(r.OrphanTemps) != 1 {
		t.Fatalf("fixture not interrupted: %+v", r)
	}
}

func assertRecovered(t *testing.T, root string) {
	t.Helper()
	r := codexwiring.InspectJournal(root)
	if r.Err != nil || len(r.Incomplete) != 0 || len(r.OrphanTemps) != 0 {
		t.Fatalf("recovery did not run before the write: %+v", r)
	}
	left, _ := filepath.Glob(filepath.Join(root, ".codex", ".codexwiring-*"))
	if len(left) != 0 {
		t.Fatalf("temp files survived: %v", left)
	}
}

func treeHashes(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		out[rel] = wiringSHA(b)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// TestCodexWiringRecoveryEntryPoints covers AC-DHR-021: enable, disable, and
// the update-path refresh recover an interrupted wiring change before any new
// write; doctor reports it and the command that recovers it without changing
// a single file.
func TestCodexWiringRecoveryEntryPoints(t *testing.T) {
	t.Run("enable", func(t *testing.T) {
		root := wiredProject(t)
		interruptWiring(t, root)
		var out, warn bytes.Buffer
		if err := runToolEnableCodexAt(root, &out, &warn, false); err != nil {
			t.Fatalf("enable: %v (%s)", err, warn.String())
		}
		assertRecovered(t, root)
	})
	t.Run("disable", func(t *testing.T) {
		root := wiredProject(t)
		interruptWiring(t, root)
		var out, warn bytes.Buffer
		if err := runToolDisableCodexAt(root, &out, &warn, false); err != nil {
			t.Fatalf("disable: %v (%s)", err, warn.String())
		}
		assertRecovered(t, root)
		if _, err := os.Stat(filepath.Join(root, codexwiring.HooksRelPath)); !os.IsNotExist(err) {
			t.Fatalf("disable did not unwire after recovering: %v\n%s", err, out.String())
		}
	})
	t.Run("update", func(t *testing.T) {
		root := wiredProject(t)
		interruptWiring(t, root)
		var out, warn bytes.Buffer
		refreshCodexWiringBestEffortAt(root, &out, &warn)
		if strings.Contains(warn.String(), "failed") || strings.Contains(warn.String(), "not refreshed") {
			t.Fatalf("update refresh warned: %s", warn.String())
		}
		assertRecovered(t, root)
	})
	t.Run("doctor_readonly", func(t *testing.T) {
		root := wiredProject(t)
		interruptWiring(t, root)
		before := treeHashes(t, root)
		check := checkCodexWiring(root, true)
		after := treeHashes(t, root)
		if fmt.Sprint(before) != fmt.Sprint(after) {
			t.Fatalf("doctor changed the tree:\nbefore=%v\nafter =%v", before, after)
		}
		report := check.Message + "\n" + check.Detail
		for _, want := range []string{codexwiring.HooksRelPath, orphanTemp, codexwiring.RecoverCommand, "incomplete"} {
			if !strings.Contains(report, want) {
				t.Errorf("doctor report lacks %q:\n%s", want, report)
			}
		}
	})
}

// writeWiringLock plants a wiring lock owned by pid.
func writeWiringLock(t *testing.T, root string, pid int, start string) {
	t.Helper()
	raw, err := json.Marshal(codexwiring.WiringLockPayload{PID: pid, ProcessStart: start})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, filepath.FromSlash(codexwiring.WiringLockRelPath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(path) })
}

// staleHooks replaces hooks.json with content the refresh must rewrite.
func staleHooks(t *testing.T, root string) []byte {
	t.Helper()
	stale := []byte("{\n  \"hooks\": {}\n}\n")
	if err := os.WriteFile(filepath.Join(root, codexwiring.HooksRelPath), stale, 0o644); err != nil {
		t.Fatal(err)
	}
	return stale
}

func hooksRefreshed(t *testing.T, root string, stale []byte) bool {
	t.Helper()
	got, err := os.ReadFile(filepath.Join(root, codexwiring.HooksRelPath))
	if err != nil {
		t.Fatal(err)
	}
	want, err := codexwiring.RenderHooks(stale)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		return false
	}
	entries, err := codexwiring.LoadJournal(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Path == codexwiring.HooksRelPath && e.State == codexwiring.JournalComplete && e.PostHash == wiringSHA(got) {
			return true
		}
	}
	return false
}

// TestCodexWiringLockUnderUpdateLock covers AC-DHR-022: the wiring lock is
// the wiring package's own, so the update path refreshes while it holds the
// update lock, enable wires standalone, a live owner of the wiring lock makes
// the update path change nothing and say so, and a dead owner's lock is
// cleared.
func TestCodexWiringLockUnderUpdateLock(t *testing.T) {
	t.Run("update_holds_update_lock", func(t *testing.T) {
		root := wiredProject(t)
		stale := staleHooks(t, root)
		release, err := acquireUpdateLock(root)
		if err != nil {
			t.Fatal(err)
		}
		defer release()
		var out, warn bytes.Buffer
		refreshCodexWiringBestEffortAt(root, &out, &warn)
		if !hooksRefreshed(t, root, stale) {
			t.Fatalf("refresh under the update lock did not refresh (warn=%s)", warn.String())
		}
	})
	t.Run("standalone_enable", func(t *testing.T) {
		root := t.TempDir()
		var out, warn bytes.Buffer
		if err := runToolEnableCodexAt(root, &out, &warn, false); err != nil {
			t.Fatalf("enable: %v", err)
		}
		entries, err := codexwiring.LoadJournal(root)
		if err != nil {
			t.Fatal(err)
		}
		complete := map[string]bool{}
		for _, e := range entries {
			if e.State == codexwiring.JournalComplete {
				complete[e.Path] = true
			}
		}
		if !complete[codexwiring.HooksRelPath] || !complete[codexwiring.ConfigRelPath] {
			t.Fatalf("standalone enable journal=%+v", entries)
		}
	})
	t.Run("lock_held_by_live_owner", func(t *testing.T) {
		root := wiredProject(t)
		stale := staleHooks(t, root)
		owner := os.Getppid()
		start, state := homestate.ProbeProcessIdentity(owner)
		if state != homestate.ProcessIdentityLive {
			t.Fatalf("parent process %d not observable as live (%s)", owner, state)
		}
		writeWiringLock(t, root, owner, start)
		var out, warn bytes.Buffer
		refreshCodexWiringBestEffortAt(root, &out, &warn)
		got, err := os.ReadFile(filepath.Join(root, codexwiring.HooksRelPath))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, stale) {
			t.Fatalf("wiring changed while another process held the lock: %q", got)
		}
		if !strings.Contains(warn.String(), "Codex wiring not refreshed") {
			t.Fatalf("update path did not report the held lock: %q", warn.String())
		}
	})
	t.Run("dead_owner_lock", func(t *testing.T) {
		root := wiredProject(t)
		stale := staleHooks(t, root)
		const deadPID = 999999999
		if _, state := homestate.ProbeProcessIdentity(deadPID); state != homestate.ProcessIdentityDead {
			t.Fatalf("pid %d not dead on this host (%s)", deadPID, state)
		}
		writeWiringLock(t, root, deadPID, "gone")
		var out, warn bytes.Buffer
		refreshCodexWiringBestEffortAt(root, &out, &warn)
		if !hooksRefreshed(t, root, stale) {
			t.Fatalf("dead owner's lock blocked the refresh (warn=%s)", warn.String())
		}
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(codexwiring.WiringLockRelPath))); !os.IsNotExist(err) {
			t.Fatalf("wiring lock left behind: %v", err)
		}
	})
}
