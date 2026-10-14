package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/spf13/cobra"
)

// writeStateFile creates a file under .moai/state with the given content and
// sets its modification time back by age.
func writeStateFile(t *testing.T, root, rel, content string, age time.Duration) {
	t.Helper()
	path := filepath.Join(root, ".moai", "state", rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	past := time.Now().Add(-age)
	if err := os.Chtimes(path, past, past); err != nil {
		t.Fatal(err)
	}
}

func TestRemoveStateResidue_RemovesStubsAndRetiredLockOnly(t *testing.T) {
	root := t.TempDir()
	const old = 48 * time.Hour
	// Residue: zero-byte databases and the retired legacy lock name.
	writeStateFile(t, root, "factory.db", "", old)
	writeStateFile(t, root, "factory/factory.db", "", old)
	writeStateFile(t, root, "todo/backlog.db", "", old)
	writeStateFile(t, root, "todo/backlog.json.lock", "", old)
	// Must survive: a populated database, the live lock name, and the
	// rollback snapshots that the queue export route depends on.
	writeStateFile(t, root, "factory/live.db", "sqlite-bytes", old)
	writeStateFile(t, root, "todo/backlog.lock", "", old)
	writeStateFile(t, root, "todo/backlog.json", `{"items":[]}`, old)
	writeStateFile(t, root, "kanban/backlog.json", `{"items":[]}`, old)
	writeStateFile(t, root, "kanban/backlog.lock", "", old)

	want := []string{
		".moai/state/factory.db",
		".moai/state/factory/factory.db",
		".moai/state/todo/backlog.db",
		".moai/state/todo/backlog.json.lock",
	}
	got := removeStateResidue(root)
	sort.Strings(got)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("removed %v, want %v", got, want)
	}
	for _, kept := range []string{"factory/live.db", "todo/backlog.lock", "todo/backlog.json", "kanban/backlog.json", "kanban/backlog.lock"} {
		if _, err := os.Stat(filepath.Join(root, ".moai", "state", kept)); err != nil {
			t.Fatalf("%s must survive: %v", kept, err)
		}
	}
	// Idempotent: a second pass finds nothing left to remove.
	if again := removeStateResidue(root); len(again) != 0 {
		t.Fatalf("second pass removed %v; want none", again)
	}
}

// A zero-byte database written moments ago may still be open in a live
// session; removing it could discard that session's first write.
func TestRemoveStateResidue_KeepsRecentStub(t *testing.T) {
	root := t.TempDir()
	writeStateFile(t, root, "todo/backlog.db", "", 0)
	if got := removeStateResidue(root); len(got) != 0 {
		t.Fatalf("removed a just-written stub %v; want none", got)
	}
	if _, err := os.Stat(filepath.Join(root, ".moai", "state", "todo", "backlog.db")); err != nil {
		t.Fatalf("recent stub must survive: %v", err)
	}
}

// writeNamespaceBackup creates one update-* entry under .moai/backups.
func writeNamespaceBackup(t *testing.T, root, name string, files map[string]string) {
	t.Helper()
	base := filepath.Join(root, defs.MoAIDir, defs.NamespaceBackupsSubdir, name)
	for rel, content := range files {
		path := filepath.Join(base, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// markProject writes the positive project marker that checkProjectMarker requires.
func markProject(t *testing.T, root string) {
	t.Helper()
	marker := filepath.Join(root, ".moai", "config", "sections", "system.yaml")
	if err := os.MkdirAll(filepath.Dir(marker), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(marker, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// Each change in a genuine project is recorded as a terminal reference row.
func TestRunUpdateHousekeeping_ReportsChangesInProject(t *testing.T) {
	root := t.TempDir()
	markProject(t, root)
	writeStateFile(t, root, "todo/backlog.db", "", 48*time.Hour)
	// Six namespace entries; the oldest holds only content the newest five also hold.
	writeNamespaceBackup(t, root, "update-2026-01-01T00-00-00Z", map[string]string{"a.yml": "shared"})
	for _, name := range []string{
		"update-2026-01-02T00-00-00Z", "update-2026-01-03T00-00-00Z", "update-2026-01-04T00-00-00Z",
		"update-2026-01-05T00-00-00Z", "update-2026-01-06T00-00-00Z",
	} {
		writeNamespaceBackup(t, root, name, map[string]string{"a.yml": "shared"})
	}

	updateLedger.reset()
	runUpdateHousekeeping(root)

	var msgs []string
	for _, row := range updateLedger.reference {
		msgs = append(msgs, row.msg)
	}
	want := []string{
		"removed 1 leftover state file(s): .moai/state/todo/backlog.db",
		"removed namespace backup update-2026-01-01T00-00-00Z: it holds no superseded version within the per-file budget",
	}
	if !reflect.DeepEqual(msgs, want) {
		t.Fatalf("reference rows = %q, want %q", msgs, want)
	}
}

// Outside a genuine project the housekeeping touches nothing and records nothing.
func TestRunUpdateHousekeeping_NoOpOutsideProject(t *testing.T) {
	root := t.TempDir()
	writeStateFile(t, root, "todo/backlog.db", "", 48*time.Hour)

	updateLedger.reset()
	runUpdateHousekeeping(root)

	if len(updateLedger.reference) != 0 {
		t.Fatalf("recorded %d row(s) outside a project; want none", len(updateLedger.reference))
	}
	if _, err := os.Stat(filepath.Join(root, ".moai", "state", "todo", "backlog.db")); err != nil {
		t.Fatalf("residue removed outside a project: %v", err)
	}
}

// A busy update lock means another update is running and will housekeep, so the
// locked variant skips and leaves the residue for that run.
func TestRunUpdateHousekeepingLocked_SkipsWhileLockHeld(t *testing.T) {
	root := t.TempDir()
	markProject(t, root)
	writeStateFile(t, root, "todo/backlog.db", "", 48*time.Hour)
	release, err := acquireUpdateLock(root)
	if err != nil {
		t.Fatalf("acquire lock: %v", err)
	}
	defer release()

	updateLedger.reset()
	runUpdateHousekeepingLocked(root)

	if len(updateLedger.reference) != 0 {
		t.Fatalf("housekeeping ran under a held lock: %d row(s)", len(updateLedger.reference))
	}
	if _, err := os.Stat(filepath.Join(root, ".moai", "state", "todo", "backlog.db")); err != nil {
		t.Fatalf("residue removed while the lock was held: %v", err)
	}
}

// Outside a project the locked variant must not create .moai/ through the lock.
func TestRunUpdateHousekeepingLocked_NoLockDirOutsideProject(t *testing.T) {
	root := t.TempDir()
	runUpdateHousekeepingLocked(root)
	if _, err := os.Stat(filepath.Join(root, defs.MoAIDir)); !os.IsNotExist(err) {
		t.Fatalf(".moai/ created outside a project: err=%v", err)
	}
}

// --restore reads the backup directory it names, so it must never run
// housekeeping first: the cap would otherwise remove the redundant oldest entry,
// which here is the very directory being restored from. Excluded by leader
// decision d-20261014T041814Z-e23a.
func TestRunUpdate_RestoreDoesNotRunHousekeeping(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	markProject(t, root)
	writeStateFile(t, root, "todo/backlog.db", "", 48*time.Hour)
	restoreSource := "update-2026-01-01T00-00-00Z"
	writeNamespaceBackup(t, root, restoreSource, map[string]string{"a.yml": "shared"})
	for _, name := range []string{
		"update-2026-01-02T00-00-00Z", "update-2026-01-03T00-00-00Z", "update-2026-01-04T00-00-00Z",
		"update-2026-01-05T00-00-00Z", "update-2026-01-06T00-00-00Z",
	} {
		writeNamespaceBackup(t, root, name, map[string]string{"a.yml": "shared"})
	}
	restoreDir := filepath.Join(root, defs.MoAIDir, defs.NamespaceBackupsSubdir, restoreSource)

	origDeps := deps
	deps = &Dependencies{}
	defer func() { deps = origDeps }()

	cmd := &cobra.Command{Use: "update-test"}
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.Flags().String("version", "", "")
	cmd.Flags().Bool("check", false, "")
	cmd.Flags().Bool("shell-env", false, "")
	cmd.Flags().Bool("config", false, "")
	cmd.Flags().Bool("binary", false, "")
	cmd.Flags().Bool("templates-only", false, "")
	cmd.Flags().Bool("verbose", false, "")
	cmd.Flags().String("restore", restoreDir, "")

	updateLedger.reset()
	restoreErr := runUpdate(cmd, nil)

	// Vacuity guard: the restore branch must have run, either to its success
	// line or to its own refusal ("restore: ..."), or the pins below prove nothing.
	reached := strings.Contains(buf.String(), "Restored .moai/config from") ||
		(restoreErr != nil && strings.Contains(restoreErr.Error(), "restore:"))
	if !reached {
		t.Fatalf("restore branch not reached: err=%v output=%q", restoreErr, buf.String())
	}
	if _, err := os.Stat(filepath.Join(root, ".moai", "state", "todo", "backlog.db")); err != nil {
		t.Fatalf("--restore ran housekeeping: residue removed: %v", err)
	}
	if _, err := os.Stat(restoreDir); err != nil {
		t.Fatalf("--restore ran the namespace cap: its source was removed: %v", err)
	}
}

// Skill-folder residue is repaired on the same housekeeping pass: an empty folder
// of a skill the template ships is removed, and the report row names it. The
// template is the real embedded one, which ships moai-lane-watchdog.
func TestRunUpdateHousekeeping_RepairsSkillResidue(t *testing.T) {
	root := t.TempDir()
	markProject(t, root)
	empty := filepath.Join(root, ".claude", "skills", "moai-lane-watchdog")
	if err := os.MkdirAll(empty, 0o755); err != nil {
		t.Fatal(err)
	}

	updateLedger.reset()
	runUpdateHousekeeping(root)

	if _, err := os.Stat(empty); !os.IsNotExist(err) {
		t.Fatalf("empty folder of a shipped skill survived housekeeping: err=%v", err)
	}
	found := false
	for _, row := range updateLedger.reference {
		if strings.Contains(row.msg, "removed empty skill folder .claude/skills/moai-lane-watchdog") {
			found = true
		}
	}
	if !found {
		t.Fatalf("no report row for the removed skill folder: %+v", updateLedger.reference)
	}
}
