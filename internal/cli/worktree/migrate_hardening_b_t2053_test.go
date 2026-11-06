package worktree

// migrate_hardening_b_t2053_test.go — card t2053 consolidated audit, group B:
// the forward and reverse move, the identity rule, the rollback preconditions.
//
//	C3   an apply-time reverse never reports `reversed` with stale metadata
//	C4   ANY forward-move error after the rename goes through the reverse order
//	C5   a truncated registration file is repaired like a stale one
//	C6   a relative gitdir pointer resolves against the tree, not the cwd
//	C13  a forward move that failed before any rename releases MovedBy
//	C14  the manifest precondition covers every non-pending step
//	C15  an unknown manifest_version is refused by name
//	G8   the apply-time reverse and the rollback share one identity rule
//	G10  a step without an identity is refused with a reason
//	C16  the refusal reason text, and --json for --rollback

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMigrateApply_ReverseWithFailingInverseMetadata_IsNotReversed — C3: the
// state file is unreadable, so the forward re-key fails and so does its
// inverse. The reverse MOVE succeeds, but the metadata still names a path the
// tree does not occupy: the tree is `failed` with a pending line, never
// `reversed`.
func TestMigrateApply_ReverseWithFailingInverseMetadata_IsNotReversed(t *testing.T) {
	w := t2053TwoTreeWorld(t)
	stateFile := filepath.Join(w.repo, ".moai", "state", "worktrees.json")
	if err := os.MkdirAll(filepath.Dir(stateFile), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(stateFile, []byte("{this is not json"), 0o644); err != nil {
		t.Fatalf("corrupt the state file: %v", err)
	}
	report, err := t2053EngineApply(t, w.repo)
	if err != nil {
		t.Fatalf("engine: %v", err)
	}
	if report.reversed != 0 {
		t.Errorf("reversed = %d, want 0: the inverse re-key failed so the metadata is stale: %+v", report.reversed, report.rows)
	}
	if report.failed != 2 {
		t.Errorf("failed = %d, want 2: %+v", report.failed, report.rows)
	}
	raw, _ := os.ReadFile(moveListPath(w.repo))
	if strings.Contains(string(raw), `"result":"reversed"`) {
		t.Errorf("a `reversed` line released the protection of a tree with stale metadata:\n%s", raw)
	}
}

// TestMigrateApply_ForwardMoveNonTimeoutErrorAfterRename_Reverses — C4: the
// move errors with something other than a timeout after the directory was
// renamed; the tree must not be left stranded at the destination.
func TestMigrateApply_ForwardMoveNonTimeoutErrorAfterRename_Reverses(t *testing.T) {
	w := t2053TwoTreeWorld(t)
	orig := forwardGitMove
	forwardGitMove = func(root, from, to string) error {
		if err := os.Rename(from, to); err != nil {
			return err
		}
		return errors.New("fatal: could not update the registration")
	}
	t.Cleanup(func() { forwardGitMove = orig })

	report, err := t2053EngineApply(t, w.repo)
	if err != nil {
		t.Fatalf("engine: %v", err)
	}
	if report.reversed != 2 || report.failed != 0 {
		t.Fatalf("reversed=%d failed=%d, want 2 reversed and 0 failed: %+v", report.reversed, report.failed, report.rows)
	}
	for _, src := range []string{w.movedTree, w.nonWTTree} {
		if !m6Registered(t, w.repo, src) {
			t.Errorf("%s is not back and registered at its source", src)
		}
	}
}

// TestRollback_EmptiedRegistrationFile_RepairsAndRestores — C5: the admin
// gitdir file is empty (git drops such a worktree from its listing), so the
// source is NOT listed either. The repair must still be tried.
func TestRollback_EmptiedRegistrationFile_RepairsAndRestores(t *testing.T) {
	w, p, id, mlPath := t2053HalfMovedWorld(t)
	admin, err := gitDirPointer(p.destination)
	if err != nil {
		t.Fatalf("admin dir: %v", err)
	}
	if werr := os.WriteFile(filepath.Join(admin, "gitdir"), nil, 0o644); werr != nil {
		t.Fatalf("empty the registration file: %v", werr)
	}
	if m6Registered(t, w.repo, p.source) {
		t.Skip("this git still lists a worktree with an empty gitdir file; C5 not reproducible here")
	}
	body, _ := json.Marshal(map[string]any{"steps": []journalStep{{
		OldPath: p.source, NewPath: p.destination, Outcome: "pending", Ident: id.Ident(),
	}}})
	journal := filepath.Join(t.TempDir(), "journal.json")
	if werr := os.WriteFile(journal, body, 0o644); werr != nil {
		t.Fatalf("write journal: %v", werr)
	}
	out, rerr := m4RunMigrate(t, "--rollback", journal)
	if rerr != nil {
		t.Fatalf("rollback of a tree with an emptied registration file failed: %v\n%s", rerr, out)
	}
	if got := t2053LastMoveListResult(t, mlPath); got != "restored" {
		t.Errorf("last move-list result = %q, want restored", got)
	}
	if !m6Registered(t, w.repo, p.source) {
		t.Errorf("the source is not registered after the rollback")
	}
}

// TestTreeIdent_RelativeGitdirPointer — C6: with worktree.useRelativePaths the
// .git file carries a path relative to the tree; the identity must resolve it
// against the tree, whatever the process working directory is.
func TestTreeIdent_RelativeGitdirPointer(t *testing.T) {
	base := m6Real(t, t.TempDir())
	repo := filepath.Join(base, "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	m6Git(t, repo, "init", "-q", "-b", "main")
	m6Git(t, repo, "config", "user.email", "t@example.com")
	m6Git(t, repo, "config", "user.name", "T")
	m6Git(t, repo, "commit", "-q", "--allow-empty", "-m", "seed")
	tree := filepath.Join(base, "tree")
	m6Git(t, repo, "-c", "worktree.useRelativePaths=true", "worktree", "add", "-q", "-b", "rel", tree)
	pointer, err := os.ReadFile(filepath.Join(tree, ".git"))
	if err != nil {
		t.Fatalf("read the pointer: %v", err)
	}
	if rest := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(string(pointer)), "gitdir:")); filepath.IsAbs(rest) {
		t.Skipf("this git wrote an absolute pointer (%q): relative paths unsupported", rest)
	}
	elsewhere := t.TempDir()
	prev, _ := os.Getwd()
	if cerr := os.Chdir(elsewhere); cerr != nil {
		t.Fatalf("chdir: %v", cerr)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })

	id, err := EnsureInstanceToken(tree)
	if err != nil {
		t.Fatalf("EnsureInstanceToken with a relative pointer: %v", err)
	}
	if id.Token == "" {
		t.Errorf("no token was written")
	}
	again, err := TreeIdent(tree)
	if err != nil || again.Ident() != id.Ident() {
		t.Errorf("TreeIdent = %v / %v, want %s", again, err, id.Ident())
	}
	if _, serr := os.Stat(filepath.Join(elsewhere, "moai-instance")); serr == nil {
		t.Errorf("the token landed in the working directory instead of the admin directory")
	}
}

// TestMigrateApply_ForwardMoveFailsBeforeRename_ReleasesMovedBy — C13: the
// forward move fails before touching anything. The intent line the engine
// wrote must not keep protecting a tree that never moved.
func TestMigrateApply_ForwardMoveFailsBeforeRename_ReleasesMovedBy(t *testing.T) {
	w := t2053TwoTreeWorld(t)
	orig := forwardGitMove
	forwardGitMove = func(root, from, to string) error { return errors.New("fatal: refused before any rename") }
	t.Cleanup(func() { forwardGitMove = orig })

	report, err := t2053EngineApply(t, w.repo)
	if err != nil {
		t.Fatalf("engine: %v", err)
	}
	if report.failed != 2 {
		t.Fatalf("failed = %d, want 2: %+v", report.failed, report.rows)
	}
	if MovedBy(w.repo, w.movedTree) {
		t.Errorf("MovedBy still protects a tree whose move never started")
	}
}

// TestRollback_ConvergedStep_WritesRestored — C13: a converged step (tree at
// the source, registered there) records a `restored` line, so MovedBy and
// --verify stop treating the step as unfinished.
func TestRollback_ConvergedStep_WritesRestored(t *testing.T) {
	w, p, id, mlPath := t2053ReverseWorld(t)
	m6Git(t, w.repo, "worktree", "move", p.destination, p.source)
	journal := t2053RollbackJournal(t, p, id)

	out, rerr := m4RunMigrate(t, "--rollback", journal)
	if rerr != nil {
		t.Fatalf("rollback of a converged step failed: %v\n%s", rerr, out)
	}
	if got := t2053LastMoveListResult(t, mlPath); got != "restored" {
		t.Errorf("last move-list result = %q, want restored", got)
	}
}

// TestRollback_FailedStepWithManifest_RefusesAChangedTree — C14: the manifest
// precondition is not limited to `moved` steps; a `failed` step that carries a
// manifest and whose tree changed since is refused.
func TestRollback_FailedStepWithManifest_RefusesAChangedTree(t *testing.T) {
	w, p, id, mlPath := t2053ReverseWorld(t)
	manifest := captureManifest(p.destination)
	if werr := os.WriteFile(filepath.Join(p.destination, "later.txt"), []byte("x"), 0o644); werr != nil {
		t.Fatalf("change the tree: %v", werr)
	}
	body, _ := json.Marshal(map[string]any{"steps": []map[string]any{{
		"old_path": p.source, "new_path": p.destination, "outcome": "failed",
		"ident": id.Ident(), "manifest": manifest, "manifest_version": manifestFormatVersion,
	}}})
	journal := filepath.Join(t.TempDir(), "journal.json")
	if werr := os.WriteFile(journal, body, 0o644); werr != nil {
		t.Fatalf("write journal: %v", werr)
	}
	out, rerr := m4RunMigrate(t, "--rollback", journal)
	if rerr == nil {
		t.Fatalf("rollback moved a changed tree back:\n%s", out)
	}
	if got := t2053LastMoveListResult(t, mlPath); got != "refused" {
		t.Errorf("last move-list result = %q, want refused", got)
	}
	_ = w
}

// TestRollback_UnknownManifestVersion_IsRefusedByName — C15: a journal from a
// newer build is refused as such, not as a content mismatch.
func TestRollback_UnknownManifestVersion_IsRefusedByName(t *testing.T) {
	w, p, id, mlPath := t2053ReverseWorld(t)
	body, _ := json.Marshal(map[string]any{"steps": []map[string]any{{
		"old_path": p.source, "new_path": p.destination, "outcome": "moved",
		"ident": id.Ident(), "manifest": "abc", "manifest_version": manifestFormatVersion + 7,
	}}})
	journal := filepath.Join(t.TempDir(), "journal.json")
	if werr := os.WriteFile(journal, body, 0o644); werr != nil {
		t.Fatalf("write journal: %v", werr)
	}
	out, rerr := m4RunMigrate(t, "--rollback", journal)
	if rerr == nil {
		t.Fatalf("rollback accepted a journal of an unknown manifest version:\n%s", out)
	}
	want := fmt.Sprintf("unknown manifest_version %d", manifestFormatVersion+7)
	if !strings.Contains(rerr.Error(), want) {
		t.Errorf("error %q does not name %q", rerr, want)
	}
	if strings.Contains(rerr.Error(), "manifest mismatch") {
		t.Errorf("error %q blames a content mismatch", rerr)
	}
	if got := t2053LastMoveListResult(t, mlPath); got != "refused" {
		t.Errorf("last move-list result = %q, want refused", got)
	}
	_ = w
}

// TestReverse_UnreadableDestinationIdentity_IsRefused — G8: the apply-time
// reverse follows the rollback's identity rule: an identity that cannot be
// read is not a match, so the reverse does not run. Since the sixth review the
// verdict is `pending` (unknown, error in the reason), not `refused`, which is
// reserved for an identity that was read and differs (TestReverseAfterFailure_
// IdentityRefusal_ReasonSaysRefused still pins that case).
func TestReverse_UnreadableDestinationIdentity_IsRefused(t *testing.T) {
	w, p, id, mlPath := t2053ReverseWorld(t)
	if err := os.WriteFile(filepath.Join(p.destination, ".git"), []byte("not a gitdir pointer\n"), 0o644); err != nil {
		t.Fatalf("break the pointer: %v", err)
	}
	movedBack, _ := m6ReverseMove(w.repo, p, id, mlPath, "migrate")
	if movedBack {
		t.Fatalf("reversed a tree whose identity could not be read")
	}
	if got := t2053LastMoveListResult(t, mlPath); got != "pending" {
		t.Errorf("last move-list result = %q, want pending", got)
	}
}

// TestReverseAfterFailure_IdentityRefusal_ReasonSaysRefused — C16: when the
// identity check refuses the reverse, the outcome reason says so instead of
// claiming the reverse move failed.
func TestReverseAfterFailure_IdentityRefusal_ReasonSaysRefused(t *testing.T) {
	w, p, _, mlPath := t2053ReverseWorld(t)
	other := TreeIdentity{AdminID: "other", Token: "other"}
	out := reverseAfterFailure(w.repo, p, other, mlPath, "migrate", "digest", "manifest mismatch after the move")
	if out.Outcome != "failed" {
		t.Fatalf("outcome = %q, want failed", out.Outcome)
	}
	if !strings.Contains(out.Reason, "refused") || !strings.Contains(out.Reason, "identity mismatch") {
		t.Errorf("reason does not name the identity refusal: %q", out.Reason)
	}
	if strings.Contains(out.Reason, "reverse move failed") {
		t.Errorf("reason claims a failed reverse move that never ran: %q", out.Reason)
	}
}

// TestRollback_EmptyIdentStep_IsRefusedWithReason — G10: a journal step that
// carries no identity cannot be matched to a tree; both the destination-present
// and the destination-absent forms are refused with a reason.
func TestRollback_EmptyIdentStep_IsRefusedWithReason(t *testing.T) {
	w, p, _, mlPath := t2053ReverseWorld(t)
	body, _ := json.Marshal(map[string]any{"steps": []journalStep{{
		OldPath: p.source, NewPath: p.destination, Outcome: "moved",
	}}})
	journal := filepath.Join(t.TempDir(), "journal.json")
	if err := os.WriteFile(journal, body, 0o644); err != nil {
		t.Fatalf("write journal: %v", err)
	}
	out, rerr := m4RunMigrate(t, "--rollback", journal)
	if rerr == nil || !strings.Contains(rerr.Error(), "no identity") {
		t.Fatalf("rollback of an identity-less step: err=%v\n%s", rerr, out)
	}
	if got := t2053LastMoveListResult(t, mlPath); got != "refused" {
		t.Errorf("last move-list result = %q, want refused", got)
	}
	if _, serr := os.Stat(p.destination); serr != nil {
		t.Errorf("the tree was moved although the step could not be matched: %v", serr)
	}
	_ = w
}

// TestRollback_JSON — C16: --rollback --json emits one JSON array of rows.
func TestRollback_JSON(t *testing.T) {
	_, p, id, _ := t2053ReverseWorld(t)
	journal := t2053RollbackJournal(t, p, id)
	out, err := m4RunMigrate(t, "--rollback", journal, "--json")
	if err != nil {
		t.Fatalf("rollback --json: %v\n%s", err, out)
	}
	var rows []map[string]any
	if jerr := json.NewDecoder(strings.NewReader(out)).Decode(&rows); jerr != nil {
		t.Fatalf("--rollback --json is not a JSON array: %v\n%s", jerr, out)
	}
	if len(rows) != 1 || rows[0]["outcome"] != "restored" {
		t.Errorf("rows: %v", rows)
	}
}
