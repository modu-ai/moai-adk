package worktree

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/hook"
)

func TestMigrationAutomaticReversePropagatesMetadataFailure(t *testing.T) {
	w := sfxWorld(t)
	dest := filepath.Join(w.homeRoot, "legacy6")
	if err := hook.RegisterLauncherWorktreeE(w.repo, w.movedTree, "WT-legacy6", "test"); err != nil {
		t.Fatal(err)
	}
	id, err := EnsureInstanceToken(w.movedTree)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := listGitOutput(w.repo, "worktree", "move", w.movedTree, dest); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(w.repo, ".moai", "state", "worktrees.json"), []byte("broken registry\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	mlPath := moveListPath(w.repo)
	if err := os.MkdirAll(filepath.Dir(mlPath), 0o755); err != nil {
		t.Fatal(err)
	}
	_, err = m6ReverseMove(w.repo, migratePlan{source: w.movedTree, destination: dest}, id, mlPath, "migrate")
	if err == nil {
		t.Error("automatic reverse swallowed its inverse registry re-key failure")
	}
	raw, err := os.ReadFile(mlPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), `"result":"reversed"`) || !strings.Contains(string(raw), `"result":"pending"`) {
		t.Fatalf("failed inverse metadata must remain pending, never settled reversed:\n%s", raw)
	}
}

func TestMigrationRollbackRetriesMetadataAfterTreeMovedBack(t *testing.T) {
	w := sfxWorld(t)
	m4InRepo(t, w.repo)
	if err := hook.RegisterLauncherWorktreeE(w.repo, w.movedTree, "WT-legacy6", "test"); err != nil {
		t.Fatal(err)
	}
	db, err := homestate.OpenFactory(w.repo)
	if err != nil {
		t.Fatal(err)
	}
	oldPath := w.movedTree
	if _, err := db.RecordPicked(context.Background(), "review-run", "t1", homestate.CardFields{WorktreePath: &oldPath}, "test", time.Now().UTC()); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	_ = db.Close()
	if out, err := m4RunMigrate(t, "--apply"); err != nil {
		t.Fatalf("apply: %v\n%s", err, out)
	}
	journal := sfxLatestJournal(t, w.repo)
	statePath := filepath.Join(w.repo, ".moai", "state", "worktrees.json")
	goodState, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(statePath, []byte("broken registry\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if out, err := m4RunMigrate(t, "--rollback", journal); err == nil {
		t.Fatalf("inverse metadata fault must fail rollback:\n%s", out)
	}
	if !m6Registered(t, w.repo, oldPath) {
		t.Fatal("fixture requires git reverse to have succeeded before metadata failed")
	}
	if err := os.WriteFile(statePath, goodState, 0o600); err != nil {
		t.Fatal(err)
	}
	if out, err := m4RunMigrate(t, "--rollback", journal); err != nil {
		t.Fatalf("retry after registry recovery: %v\n%s", err, out)
	}
	db, err = homestate.OpenFactory(w.repo)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	card, err := db.LoadCard(context.Background(), "review-run", "t1")
	if err != nil || card.WorktreePath != oldPath {
		t.Fatalf("retry did not restore the card's actual tree binding: %q, %v", card.WorktreePath, err)
	}
	if MovedBy(w.repo, oldPath) {
		t.Fatal("fully restored tree still carries pending migration protection")
	}
}

func TestMigrateVerifyRejectsManifestMutationAndRollbackKeepsProtection(t *testing.T) {
	w := sfxWorld(t)
	m4InRepo(t, w.repo)
	if out, err := m4RunMigrate(t, "--apply"); err != nil {
		t.Fatalf("apply: %v\n%s", err, out)
	}
	dest := filepath.Join(w.homeRoot, "legacy6")
	if !MovedBy(w.repo, dest) {
		t.Fatal("applied migration must protect its destination")
	}
	if err := os.WriteFile(filepath.Join(dest, "operator-evidence.txt"), []byte("new evidence\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if out, err := m4RunMigrate(t, "--verify"); err == nil || !strings.Contains(out, "mismatch") {
		t.Errorf("verify accepted changed destination contents: err=%v\n%s", err, out)
	}
	if out, err := m4RunMigrate(t, "--rollback", sfxLatestJournal(t, w.repo)); err == nil || !strings.Contains(out, "manifest mismatch") {
		t.Fatalf("rollback must refuse changed destination contents: err=%v\n%s", err, out)
	}
	if !MovedBy(w.repo, dest) {
		t.Error("rollback refusal removed cleanup protection from the still-registered migrated tree")
	}
	if body, err := os.ReadFile(filepath.Join(dest, "operator-evidence.txt")); err != nil || string(body) != "new evidence\n" {
		t.Fatalf("rollback refusal lost operator evidence: %q, %v", body, err)
	}
}

func TestMoveListDedupeUsesLatestRelatedSemanticRecord(t *testing.T) {
	path := filepath.Join(t.TempDir(), "move-list.jsonl")
	first := m6MoveLineTriggerReason("migrate", "moved", "/old", "/new", "tree:token", "same reason")
	unrelated := m6MoveLine("moved", "/another-old", "/another-new", "other:token")
	if err := appendMoveListLine(path, first); err != nil {
		t.Fatal(err)
	}
	if err := appendMoveListLine(path, unrelated); err != nil {
		t.Fatal(err)
	}
	var retry map[string]any
	if err := json.Unmarshal([]byte(first), &retry); err != nil {
		t.Fatal(err)
	}
	retry["ts"] = "2099-01-01T00:00:00Z"
	body, err := json.Marshal(retry)
	if err != nil {
		t.Fatal(err)
	}
	if err := appendMoveListLine(path, string(body)); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if count := len(strings.Split(strings.TrimSpace(string(raw)), "\n")); count != 2 {
		t.Fatalf("semantic retry after unrelated entry added audit noise: %d records\n%s", count, raw)
	}
	if err := appendMoveListLine(path, m6MoveLine("restored", "/new", "/old", "tree:token")); err != nil {
		t.Fatal(err)
	}
	if err := appendMoveListLine(path, first); err != nil {
		t.Fatal(err)
	}
	raw, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if count := len(strings.Split(strings.TrimSpace(string(raw)), "\n")); count != 4 {
		t.Fatalf("move re-applied after rollback must remain in audit history: %d records\n%s", count, raw)
	}
}
