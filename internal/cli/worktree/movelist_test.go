package worktree

// movelist_test.go - SPEC-WORKTREE-HOME-ROOT-001 M1 RED commit (card t1640).
//
// The tree-identity reader and writer and the move-list reader
// (internal/cli/worktree/movelist*.go) do not exist yet: this package fails to
// compile until the M1 production commits, which is the RED (plan.md section
// C, test compilation order).
//
// Acceptance mapping: AC-WHR-012b reader clauses (steps 7 to 14 and the
// mutant probe of the lookup), AC-WHR-014b witness (MovedBy, the "moved by the
// migration" reverse lookup of design.md D-11). The move-list lines here are
// hand-written fixtures, not output of the M6 writer, which is built later.
//
// Surface fixed by this file (the GREEN stage implements exactly this):
//
//	type TreeIdentity struct{ AdminID, Token string } // immutable part only
//	func (i TreeIdentity) Ident() string              // "<admin id>:<token>"
//	func TreeIdent(path string) (TreeIdentity, error) // read-only; Token "" when absent
//	func EnsureInstanceToken(path string) (TreeIdentity, error) // one-time writer
//	type MoveRequester struct{ Card, Tree, Ident string }
//	func MovedTo(primary, oldPath string, requester MoveRequester, cardPath string) (string, bool)
//	func MovedBy(primary, path string) bool
//
// The admin id is the leaf of the gitdir pointer in the tree's .git file; the
// token lives at <common-dir>/worktrees/<admin id>/moai-instance (git's
// administrative directory, never the working tree). HEAD and branch are
// display-only and are never part of the identity. MovedTo returns (path,
// true) or ("", false) for "no answer". The move list is
// <primary>/.moai/state/worktree-migrate/move-list.jsonl.
//
// Every fixture is a scratch repository under t.TempDir(). The linked trees
// made with `git worktree add` / `git worktree move` / `git worktree remove`
// are scratch fixtures, the sanctioned exception to AGENTS.md section 3.

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// ---------------------------------------------------------------- fixtures

func mlReal(t *testing.T, p string) string {
	t.Helper()
	r, err := filepath.EvalSymlinks(p)
	if err != nil {
		t.Fatalf("evalsymlinks %s: %v", p, err)
	}
	return r
}

func mlGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

// mlFix is one scratch repository (the "primary") with a directory for the
// pre-move trees ("old"), one for their destinations ("new") and counters that
// keep branch names unique.
type mlFix struct {
	t       *testing.T
	primary string
	base    string
	seq     int
}

func mlNew(t *testing.T) *mlFix {
	t.Helper()
	base := mlReal(t, t.TempDir())
	primary := filepath.Join(base, "primary")
	if err := os.MkdirAll(primary, 0o755); err != nil {
		t.Fatalf("mkdir primary: %v", err)
	}
	mlGit(t, primary, "init", "-q", "--template=", "-b", "main")
	mlGit(t, primary, "config", "user.email", "ml-test@example.com")
	mlGit(t, primary, "config", "user.name", "ML Test")
	mlGit(t, primary, "config", "commit.gpgsign", "false")
	if err := os.WriteFile(filepath.Join(primary, "README.md"), []byte("seed\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	mlGit(t, primary, "add", "README.md")
	mlGit(t, primary, "commit", "-q", "-m", "seed")
	return &mlFix{t: t, primary: primary, base: base}
}

// old and nw are the pre-move and destination spellings of one tree name; both
// share the leaf, as a migration's source and destination do, so git's admin
// id (named after the leaf at creation) is the one a replacement tree reuses.
func (f *mlFix) old(name string) string { return filepath.Join(f.base, "old", name) }
func (f *mlFix) nw(name string) string  { return filepath.Join(f.base, "new", name) }

// add creates a linked worktree at path on a fresh branch and returns path.
func (f *mlFix) add(path string) string {
	f.t.Helper()
	f.seq++
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		f.t.Fatalf("mkdir parent: %v", err)
	}
	mlGit(f.t, f.primary, "worktree", "add", "-q", "-b", fmt.Sprintf("WT-%s-%d", filepath.Base(path), f.seq), "--", path)
	return path
}

func (f *mlFix) mv(from, to string) {
	f.t.Helper()
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		f.t.Fatalf("mkdir parent: %v", err)
	}
	mlGit(f.t, f.primary, "worktree", "move", from, to)
}

func (f *mlFix) rm(path string) {
	f.t.Helper()
	mlGit(f.t, f.primary, "worktree", "remove", "--force", path)
}

// ident gives the tree its instance token (the M1 one-time writer) and
// returns the tree's ident string.
func (f *mlFix) ident(path string) string {
	f.t.Helper()
	id, err := EnsureInstanceToken(path)
	if err != nil {
		f.t.Fatalf("EnsureInstanceToken(%s): %v", path, err)
	}
	return id.Ident()
}

func (f *mlFix) listPath() string {
	return filepath.Join(f.primary, ".moai", "state", "worktree-migrate", "move-list.jsonl")
}

// mlLine mirrors the fields of design.md D-11; a hand-written fixture line.
type mlLine struct {
	Schema  int    `json:"schema"`
	TS      string `json:"ts"`
	Run     string `json:"run"`
	Trigger string `json:"trigger"`
	Tree    string `json:"tree"`
	Card    string `json:"card"`
	OldPath string `json:"old_path"`
	NewPath string `json:"new_path"`
	Result  string `json:"result"`
	Reason  string `json:"reason"`
	Ident   string `json:"ident,omitempty"`
	Head    string `json:"head,omitempty"`
	Branch  string `json:"branch,omitempty"`
}

func mlMk(result, tree, card, oldPath, newPath, ident string) mlLine {
	return mlLine{
		Schema: 1, TS: "2026-10-19T12:00:00Z", Run: "20261019T120000Z", Trigger: "migrate",
		Tree: tree, Card: card, OldPath: oldPath, NewPath: newPath, Result: result, Ident: ident,
	}
}

// appendRaw appends text to the move list, creating it (and its directory).
func (f *mlFix) appendRaw(text string) {
	f.t.Helper()
	if err := os.MkdirAll(filepath.Dir(f.listPath()), 0o755); err != nil {
		f.t.Fatalf("mkdir state: %v", err)
	}
	fh, err := os.OpenFile(f.listPath(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		f.t.Fatalf("open move list: %v", err)
	}
	defer func() { _ = fh.Close() }()
	if _, err := fh.WriteString(text); err != nil {
		f.t.Fatalf("append move list: %v", err)
	}
}

func (f *mlFix) append(lines ...mlLine) {
	f.t.Helper()
	for _, l := range lines {
		b, err := json.Marshal(l)
		if err != nil {
			f.t.Fatalf("marshal line: %v", err)
		}
		f.appendRaw(string(b) + "\n")
	}
}

// wantTo asserts one MovedTo answer; want "" means "no answer".
func (f *mlFix) wantTo(label, oldPath string, req MoveRequester, cardPath, want string) {
	f.t.Helper()
	got, ok := MovedTo(f.primary, oldPath, req, cardPath)
	if want == "" {
		if ok {
			f.t.Errorf("%s: MovedTo = %q, want no answer", label, got)
		}
		return
	}
	if !ok || got != want {
		f.t.Errorf("%s: MovedTo = (%q, %v), want %q", label, got, ok, want)
	}
}

// wantBy asserts one MovedBy answer.
func (f *mlFix) wantBy(label, path string, want bool) {
	f.t.Helper()
	if got := MovedBy(f.primary, path); got != want {
		f.t.Errorf("%s: MovedBy(%s) = %v, want %v", label, path, got, want)
	}
}

func mlReq(card string) MoveRequester { return MoveRequester{Card: card, Tree: card} }

// ------------------------------------------------- tree identity (TreeIdent)

// TestTreeIdent_ImmutablePartSurvivesCommitBranchRenameAndMove pins the
// immutable identity of D-11: admin id (leaf of the gitdir pointer) plus
// instance token. A commit, a branch rename and a `git worktree move` change
// HEAD, branch and path, and none of those changes the identity.
func TestTreeIdent_ImmutablePartSurvivesCommitBranchRenameAndMove(t *testing.T) {
	f := mlNew(t)
	a := f.add(f.old("t7"))

	raw, err := os.ReadFile(filepath.Join(a, ".git"))
	if err != nil {
		t.Fatalf("read .git file: %v", err)
	}
	gitdir := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(string(raw)), "gitdir:"))
	wantAdmin := filepath.Base(gitdir)
	tokenFile := filepath.Join(gitdir, "moai-instance")

	before, err := TreeIdent(a)
	if err != nil {
		t.Fatalf("TreeIdent: %v", err)
	}
	if before.AdminID != wantAdmin {
		t.Errorf("AdminID = %q, want the leaf of the gitdir pointer %q", before.AdminID, wantAdmin)
	}
	if before.Token != "" {
		t.Errorf("Token = %q for a tree nobody tokened, want empty", before.Token)
	}
	// TreeIdent is read-only: it must not create the token.
	if _, err := os.Stat(tokenFile); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("TreeIdent created %s (stat err=%v); the reader must be read-only", tokenFile, err)
	}

	tokened, err := EnsureInstanceToken(a)
	if err != nil {
		t.Fatalf("EnsureInstanceToken: %v", err)
	}
	if tokened.AdminID != wantAdmin || tokened.Token == "" {
		t.Fatalf("EnsureInstanceToken = %+v, want admin id %q and a non-empty token", tokened, wantAdmin)
	}
	if got, want := tokened.Ident(), wantAdmin+":"+tokened.Token; got != want {
		t.Errorf("Ident() = %q, want %q", got, want)
	}
	if b, err := os.ReadFile(tokenFile); err != nil || !strings.Contains(string(b), tokened.Token) {
		t.Errorf("token file %s = %q (err=%v), want it to hold the token", tokenFile, b, err)
	}

	// Display-only changes: a commit, a branch rename, a move.
	if err := os.WriteFile(filepath.Join(a, "after.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	mlGit(t, a, "add", "after.txt")
	mlGit(t, a, "commit", "-q", "-m", "after")
	mlGit(t, a, "branch", "-m", "WT-renamed")
	moved := f.nw("t7")
	f.mv(a, moved)

	after, err := TreeIdent(moved)
	if err != nil {
		t.Fatalf("TreeIdent after the move: %v", err)
	}
	if after != tokened {
		t.Errorf("identity changed across commit, branch rename and move: before %+v, after %+v", tokened, after)
	}
}

// TestEnsureInstanceToken_WritesOnceInTheGitAdminDirectoryOnly pins the
// one-time writer: it writes the token once, never again, never in the
// working tree, and gives two trees two different tokens.
func TestEnsureInstanceToken_WritesOnceInTheGitAdminDirectoryOnly(t *testing.T) {
	f := mlNew(t)
	a := f.add(f.old("t7"))
	b := f.add(f.old("t8"))

	first, err := EnsureInstanceToken(a)
	if err != nil {
		t.Fatalf("first EnsureInstanceToken: %v", err)
	}
	tokenFile := filepath.Join(f.primary, ".git", "worktrees", first.AdminID, "moai-instance")
	body1, err := os.ReadFile(tokenFile)
	if err != nil {
		t.Fatalf("token file: %v", err)
	}
	second, err := EnsureInstanceToken(a)
	if err != nil {
		t.Fatalf("second EnsureInstanceToken: %v", err)
	}
	if second != first {
		t.Errorf("second call = %+v, want the unchanged %+v (the token is written once, never again)", second, first)
	}
	if body2, _ := os.ReadFile(tokenFile); string(body2) != string(body1) {
		t.Errorf("the token file was rewritten: %q then %q", body1, body2)
	}
	if got, err := TreeIdent(a); err != nil || got != first {
		t.Errorf("TreeIdent = %+v, %v, want the writer's %+v", got, err, first)
	}

	other, err := EnsureInstanceToken(b)
	if err != nil {
		t.Fatalf("EnsureInstanceToken(b): %v", err)
	}
	if other.Token == first.Token || other.AdminID == first.AdminID {
		t.Errorf("two trees share identity parts: %+v and %+v", first, other)
	}

	// Nothing is written into the working tree.
	if out := mlGit(t, a, "status", "--porcelain", "--ignored"); out != "" {
		t.Errorf("the writer left traces in the working tree:\n%s", out)
	}
}

func TestEnsureInstanceToken_FailsWhenTheAdminDirectoryCannotBeWritten(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("directory permissions are not enforced for this user or platform")
	}
	f := mlNew(t)
	a := f.add(f.old("t7"))
	admin := filepath.Join(f.primary, ".git", "worktrees", "t7")
	if err := os.Chmod(admin, 0o555); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(admin, 0o755) })

	if _, err := EnsureInstanceToken(a); err == nil {
		t.Errorf("EnsureInstanceToken succeeded over an unwritable admin directory")
	}
	if _, err := os.Stat(filepath.Join(admin, "moai-instance")); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("a token file exists after the failed write (stat err=%v)", err)
	}
}

// TestTreeIdent_RemovedTreeTakesItsTokenWithIt pins the premise of the
// dispose-and-recreate cases (AC-WHR-012b step 9d): git reuses the admin id of
// a removed tree, the token directory goes with it, so a tree created later at
// the same path has the same admin id and no token.
func TestTreeIdent_RemovedTreeTakesItsTokenWithIt(t *testing.T) {
	f := mlNew(t)
	a := f.add(f.old("t7"))
	x, err := EnsureInstanceToken(a)
	if err != nil {
		t.Fatalf("EnsureInstanceToken: %v", err)
	}
	f.rm(a)
	z := f.add(f.old("t7"))

	zi, err := TreeIdent(z)
	if err != nil {
		t.Fatalf("TreeIdent of the replacement: %v", err)
	}
	if zi.AdminID != x.AdminID {
		t.Errorf("fixture premise: git did not reuse admin id %q (got %q); the dispose-and-recreate cases below would not exercise reuse", x.AdminID, zi.AdminID)
	}
	if zi.Token != "" {
		t.Errorf("the replacement tree carries token %q; the removed tree's token must go with its admin directory", zi.Token)
	}
	if zi.Ident() == x.Ident() {
		t.Errorf("the replacement tree has the removed tree's ident %q", x.Ident())
	}
}

func TestTreeIdent_ErrorsForPathsThatAreNotLinkedWorktrees(t *testing.T) {
	f := mlNew(t)
	plain := filepath.Join(f.base, "plain")
	if err := os.MkdirAll(plain, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	for name, p := range map[string]string{
		"missing path":    filepath.Join(f.base, "no-such-tree"),
		"plain directory": plain,
	} {
		if _, err := TreeIdent(p); err == nil {
			t.Errorf("TreeIdent(%s) succeeded for a %s", p, name)
		}
		if _, err := EnsureInstanceToken(p); err == nil {
			t.Errorf("EnsureInstanceToken(%s) succeeded for a %s", p, name)
		}
	}
}

// ------------------------------------------------------ MovedTo (the lookup)

// TestMovedTo_ForwardChainRollbackAndReapply pins AC-WHR-012b steps 3, 4 and
// 7: the answer follows a move, a rollback and a re-apply, and the card
// record's rebound path is consulted first.
func TestMovedTo_ForwardChainRollbackAndReapply(t *testing.T) {
	f := mlNew(t)
	a := f.add(f.old("t7"))
	b := f.nw("t7")
	x := f.ident(a)
	req := mlReq("t7")

	f.mv(a, b)
	f.append(mlMk("pending", "t7", "t7", a, b, x), mlMk("moved", "t7", "t7", a, b, x))
	f.wantTo("settled move, no card pointer", a, req, "", b)
	f.wantTo("settled move, stale card pointer naming the old path", a, req, a, b)
	f.wantTo("settled move, rebound card pointer", a, req, b, b)

	f.mv(b, a)
	f.append(mlMk("restored", "t7", "t7", b, a, x))
	f.wantTo("after the rollback", a, req, "", a)

	f.mv(a, b)
	f.append(mlMk("pending", "t7", "t7", a, b, x), mlMk("moved", "t7", "t7", a, b, x))
	f.wantTo("after the re-apply", a, req, "", b)
}

func TestMovedTo_TwoMoveChainFollowsEveryLaterEntry(t *testing.T) {
	f := mlNew(t)
	a := f.add(f.old("t7"))
	b := f.nw("t7")
	c := filepath.Join(f.base, "newer", "t7")
	x := f.ident(a)
	f.mv(a, b)
	f.mv(b, c)
	f.append(mlMk("moved", "t7", "t7", a, b, x), mlMk("moved", "t7", "t7", b, c, x))
	f.wantTo("from the first path", a, mlReq("t7"), "", c)
	f.wantTo("from the intermediate path", b, mlReq("t7"), "", c)
}

// TestMovedTo_PathReusedByAnotherTreeAnswersEachItsOwn pins step 5: a tree
// created at a path another tree left is looked up under its own requester.
func TestMovedTo_PathReusedByAnotherTreeAnswersEachItsOwn(t *testing.T) {
	f := mlNew(t)
	a := f.add(f.old("t7"))
	b := f.nw("t7")
	x := f.ident(a)
	f.mv(a, b)
	f.append(mlMk("pending", "t7", "t7", a, b, x), mlMk("moved", "t7", "t7", a, b, x))

	y := f.add(f.old("t7"))
	yi := f.ident(y)
	c := f.nw("t8")
	f.mv(y, c)
	f.append(mlMk("pending", "t8", "t8", a, c, yi), mlMk("moved", "t8", "t8", a, c, yi))

	f.wantTo("first tree's card", a, mlReq("t7"), "", b)
	f.wantTo("second tree's card", a, mlReq("t8"), "", c)
}

// TestMovedTo_SkipsUnparseableLines pins step 8: an unparseable line in the
// middle, and a torn tail without its newline, never end the lookup.
func TestMovedTo_SkipsUnparseableLines(t *testing.T) {
	t.Run("in the middle and at the end", func(t *testing.T) {
		f := mlNew(t)
		a := f.add(f.old("t7"))
		b := f.nw("t7")
		c := filepath.Join(f.base, "newer", "t7")
		x := f.ident(a)
		f.mv(a, b)
		f.mv(b, c)
		f.append(mlMk("moved", "t7", "t7", a, b, x))
		f.appendRaw("{ this line is not json\n")
		f.append(mlMk("moved", "t7", "t7", b, c, x))
		f.appendRaw(`{"schema":1,"ts":`)
		f.wantTo("chain across an unparseable line", a, mlReq("t7"), "", c)
	})
	t.Run("torn tail only", func(t *testing.T) {
		f := mlNew(t)
		a := f.add(f.old("t7"))
		b := f.nw("t7")
		x := f.ident(a)
		f.mv(a, b)
		f.append(mlMk("moved", "t7", "t7", a, b, x))
		f.appendRaw(`{"schema":1,"ts":`)
		f.wantTo("settled line before a torn tail", a, mlReq("t7"), "", b)
	})
}

// TestMovedTo_UnsettledPendingKillWindow pins step 9: a kill after the intent
// line. With the tree already at the destination the lookup answers the
// destination (the card record still names the old path, which is no longer
// registered); with the tree still at the source it answers the source.
func TestMovedTo_UnsettledPendingKillWindow(t *testing.T) {
	t.Run("kill after the move", func(t *testing.T) {
		f := mlNew(t)
		a := f.add(f.old("t7"))
		b := f.nw("t7")
		x := f.ident(a)
		f.mv(a, b)
		f.append(mlMk("pending", "t7", "t7", a, b, x))
		f.wantTo("card pointer on the vanished old path", a, mlReq("t7"), a, b)
		f.wantTo("no card pointer", a, mlReq("t7"), "", b)
	})
	t.Run("kill before the move", func(t *testing.T) {
		f := mlNew(t)
		a := f.add(f.old("t7"))
		x := f.ident(a)
		f.append(mlMk("pending", "t7", "t7", a, f.nw("t7"), x))
		f.wantTo("card pointer on the source", a, mlReq("t7"), a, a)
		f.wantTo("no card pointer", a, mlReq("t7"), "", a)
	})
}

// TestMovedTo_StaleCardPointerAtAnOldPathNowHeldByAnotherTree pins step 9b:
// the old path is registered again, by a different tree, and the card record
// still names it. The expected identity comes from the requester's own
// pending line, so neither the card pointer nor the pending line's old path
// is accepted; the destination, which carries that identity, is.
func TestMovedTo_StaleCardPointerAtAnOldPathNowHeldByAnotherTree(t *testing.T) {
	for _, tokened := range []bool{false, true} {
		t.Run(fmt.Sprintf("replacement tokened=%v", tokened), func(t *testing.T) {
			f := mlNew(t)
			a := f.add(f.old("t7"))
			b := f.nw("t7")
			x := f.ident(a)
			f.mv(a, b)
			f.append(mlMk("pending", "t7", "t7", a, b, x))
			y := f.add(f.old("t7"))
			if tokened {
				f.ident(y)
			}
			f.wantTo("card pointer naming the old path held by a stranger", a, mlReq("t7"), a, b)
		})
	}
}

// TestMovedTo_CommitAndBranchRenameInsideTheKillWindow pins step 9c: HEAD and
// branch are display-only; a commit and a `git branch -m` after the kill do
// not change the answer.
func TestMovedTo_CommitAndBranchRenameInsideTheKillWindow(t *testing.T) {
	f := mlNew(t)
	a := f.add(f.old("t7"))
	b := f.nw("t7")
	x := f.ident(a)
	f.mv(a, b)
	line := mlMk("pending", "t7", "t7", a, b, x)
	line.Head = "0000000000000000000000000000000000000001"
	line.Branch = "WT-t7-1"
	f.append(line)

	if err := os.WriteFile(filepath.Join(b, "after-kill.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	mlGit(t, b, "add", "after-kill.txt")
	mlGit(t, b, "commit", "-q", "-m", "after the kill")
	mlGit(t, b, "branch", "-m", "WT-renamed-after-kill")

	f.wantTo("HEAD and branch differ from the pending line", a, mlReq("t7"), a, b)
}

// TestMovedTo_DestinationDisposedAndReplaced pins step 9d: a settled move whose
// destination later holds a different tree (git reuses the admin id; the new
// tree has no token or its own) is answered to nobody.
func TestMovedTo_DestinationDisposedAndReplaced(t *testing.T) {
	for _, tokened := range []bool{false, true} {
		t.Run(fmt.Sprintf("replacement tokened=%v", tokened), func(t *testing.T) {
			f := mlNew(t)
			a := f.add(f.old("t7"))
			b := f.nw("t7")
			x := f.ident(a)
			f.mv(a, b)
			f.append(mlMk("pending", "t7", "t7", a, b, x), mlMk("moved", "t7", "t7", a, b, x))
			f.rm(b)
			z := f.add(b)
			if tokened {
				f.ident(z)
			}
			f.wantTo("by card id from the pre-move path", a, mlReq("t7"), "", "")
			f.wantTo("card record naming the destination", a, mlReq("t7"), b, "")
		})
	}
}

// TestMovedTo_TwoMigrationsThroughOnePath pins step 12: the stale pointer of
// the first tree's card never receives the second tree, because the expected
// identity is fixed from the requester's own lines, never from the latest line
// naming the destination.
func TestMovedTo_TwoMigrationsThroughOnePath(t *testing.T) {
	f := mlNew(t)
	a := f.add(f.old("t7"))
	b := f.nw("t7")
	x := f.ident(a)
	f.mv(a, b)
	f.append(mlMk("pending", "t7", "t7", a, b, x), mlMk("moved", "t7", "t7", a, b, x))
	f.rm(b)

	y := f.add(f.old("t7"))
	yi := f.ident(y)
	f.mv(y, b)
	f.append(mlMk("pending", "t8", "t8", a, b, yi), mlMk("moved", "t8", "t8", a, b, yi))

	f.wantTo("first tree's card, from the old path", a, mlReq("t7"), "", "")
	f.wantTo("first tree's stale card record naming the destination", a, mlReq("t7"), b, "")
	f.wantTo("second tree's card, from the old path", a, mlReq("t8"), "", b)
}

// TestMovedTo_TwoTreesOnOneCardNeedASuppliedIdent pins step 13: with two
// idents behind one (card, path) there is no expected identity unless the
// caller supplies one.
func TestMovedTo_TwoTreesOnOneCardNeedASuppliedIdent(t *testing.T) {
	f := mlNew(t)
	a := f.add(f.old("t7"))
	b := f.nw("t7")
	x := f.ident(a)
	f.mv(a, b)
	f.append(mlMk("pending", "t7", "t7", a, b, x), mlMk("moved", "t7", "t7", a, b, x))
	f.rm(b)

	y := f.add(f.old("t7"))
	yi := f.ident(y)
	f.mv(y, b)
	f.append(mlMk("pending", "t7", "t7", a, b, yi), mlMk("moved", "t7", "t7", a, b, yi))

	f.wantTo("no ident supplied", a, mlReq("t7"), "", "")
	f.wantTo("first tree's ident supplied", a, MoveRequester{Card: "t7", Tree: "t7", Ident: x}, "", "")
	f.wantTo("second tree's ident supplied", a, MoveRequester{Card: "t7", Tree: "t7", Ident: yi}, "", b)
}

// TestMovedTo_UnregisteredPathsAreNeverReturned pins step 10: a chain whose
// final path is not a registered worktree, and a card record naming an
// unregistered path, both yield no answer.
func TestMovedTo_UnregisteredPathsAreNeverReturned(t *testing.T) {
	f := mlNew(t)
	a := f.add(f.old("t7"))
	b := f.nw("t7")
	x := f.ident(a)
	f.rm(a)
	f.append(mlMk("moved", "t7", "t7", a, b, x))
	f.wantTo("chain ending at an unregistered path", a, mlReq("t7"), "", "")

	plain := filepath.Join(f.base, "plain-directory")
	if err := os.MkdirAll(plain, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	f.wantTo("card record naming a directory git does not register", a, mlReq("t7"), plain, "")
}

// TestMovedTo_SkippedLineCarriesNoIdentAndNoPath: a skipped line has no ident
// and is not a location line; it never produces a destination, and a
// registered card pointer stays valid because no ident-bearing line names it.
func TestMovedTo_SkippedLineCarriesNoIdentAndNoPath(t *testing.T) {
	f := mlNew(t)
	a := f.add(f.old("t7"))
	skipped := mlMk("skipped", "t7", "t7", a, f.nw("t7"), "")
	skipped.Reason = "locked"
	f.append(skipped)

	f.wantTo("registered card pointer", a, mlReq("t7"), a, a)
	f.wantTo("no card pointer", a, mlReq("t7"), "", "")
}

// TestMovedTo_RequesterWithoutAnIdent pins step 14: a tree made before this
// SPEC (no token, no line) is answered on registration alone; and when the
// expected identity is empty a candidate named by an ident-bearing line is not
// answered at all.
func TestMovedTo_RequesterWithoutAnIdent(t *testing.T) {
	t.Run("tokenless tree, no lines, no move list", func(t *testing.T) {
		f := mlNew(t)
		a := f.add(f.old("t7"))
		f.wantTo("registered card record", a, mlReq("t7"), a, a)
		f.wantTo("no card record", a, mlReq("t7"), "", "")
	})
	t.Run("a stranger registered at the card record's path is not verified", func(t *testing.T) {
		f := mlNew(t)
		stranger := f.add(f.old("t9"))
		f.ident(stranger)
		f.wantTo("card record naming a registered stranger", f.old("t7"), mlReq("t7"), stranger, stranger)
	})
	t.Run("a path an ident-bearing line names is not answered to an empty identity", func(t *testing.T) {
		f := mlNew(t)
		a := f.add(f.old("t7"))
		b := f.nw("t7")
		x := f.ident(a)
		f.mv(a, b)
		f.append(mlMk("moved", "t7", "t7", a, b, x))
		f.wantTo("tokenless requester whose card record names the moved tree's destination", f.old("t9"), mlReq("t9"), b, "")
	})
}

// ------------------------------------------------------ MovedBy (the witness)

// mlMovedTree builds the common fixture: tree X tokened at a, moved to b, with
// its pending and moved lines.
func mlMovedTree(t *testing.T) (f *mlFix, a, b, x string) {
	t.Helper()
	f = mlNew(t)
	a = f.add(f.old("t7"))
	x = f.ident(a)
	b = f.nw("t7")
	f.mv(a, b)
	f.append(mlMk("pending", "t7", "t7", a, b, x), mlMk("moved", "t7", "t7", a, b, x))
	return f, a, b, x
}

// TestMovedBy_IdentKeyedAcrossAllLines pins the reverse lookup of REQ-WHR-014:
// keyed on the tree's own ident across every line; the tree's latest location
// line must be a moved, pending or refused whose new_path is the path, and only
// a later restored or reversed releases it. Other idents, skipped, failed and
// unparseable lines never enter the decision.
func TestMovedBy_IdentKeyedAcrossAllLines(t *testing.T) {
	t.Run("moved tree standing at its destination", func(t *testing.T) {
		f, _, b, _ := mlMovedTree(t)
		f.wantBy("destination", b, true)
	})

	t.Run("a path that is not registered", func(t *testing.T) {
		f, _, _, _ := mlMovedTree(t)
		f.wantBy("unregistered path", filepath.Join(f.base, "nowhere"), false)
	})

	t.Run("a tree that was never moved", func(t *testing.T) {
		f, _, _, _ := mlMovedTree(t)
		tokenless := f.add(f.old("t5"))
		tokened := f.add(f.old("t6"))
		f.ident(tokened)
		f.wantBy("tree with no token", tokenless, false)
		f.wantBy("tokened tree with no line", tokened, false)
	})

	t.Run("no move list at all", func(t *testing.T) {
		f := mlNew(t)
		a := f.add(f.old("t7"))
		f.ident(a)
		f.wantBy("tokened tree, no move list", a, false)
	})

	for _, release := range []string{"restored", "reversed"} {
		t.Run(release+" releases the protection", func(t *testing.T) {
			f, a, b, x := mlMovedTree(t)
			f.mv(b, a)
			f.append(mlMk(release, "t7", "t7", b, a, x))
			f.wantBy("tree back at its source", a, false)
		})
	}

	t.Run("reversed releases even when the tree still stands at the destination", func(t *testing.T) {
		f, a, b, x := mlMovedTree(t)
		f.append(mlMk("reversed", "t7", "t7", b, a, x))
		f.wantBy("destination after a reversed line", b, false)
	})

	t.Run("moved again after a rollback is protected again", func(t *testing.T) {
		f, a, b, x := mlMovedTree(t)
		f.mv(b, a)
		f.append(mlMk("restored", "t7", "t7", b, a, x))
		f.mv(a, b)
		f.append(mlMk("pending", "t7", "t7", a, b, x), mlMk("moved", "t7", "t7", a, b, x))
		f.wantBy("destination after the re-apply", b, true)
	})

	t.Run("a worker created at the destination after the rollback is not protected", func(t *testing.T) {
		f, a, b, x := mlMovedTree(t)
		f.mv(b, a)
		f.append(mlMk("restored", "t7", "t7", b, a, x))
		w := f.add(b)
		f.wantBy("stranger at the old destination", w, false)
	})

	t.Run("a later skipped line of another tree naming the same path", func(t *testing.T) {
		f, a, b, _ := mlMovedTree(t)
		skipped := mlMk("skipped", "t8", "t8", filepath.Join(f.base, "old", "t8"), b, "")
		skipped.Reason = "locked"
		f.append(skipped)
		f.wantBy("destination after a foreign skipped line", b, true)
		_ = a
	})

	t.Run("a later moved line of another tree naming the same path", func(t *testing.T) {
		f, _, b, _ := mlMovedTree(t)
		f.append(mlMk("moved", "t8", "t8", filepath.Join(f.base, "old", "t8"), b, "other-admin:other-token"))
		f.wantBy("destination after a foreign moved line", b, true)
	})

	t.Run("latest pending line, tree at its new path (kill after the move)", func(t *testing.T) {
		f := mlNew(t)
		a := f.add(f.old("t7"))
		x := f.ident(a)
		b := f.nw("t7")
		f.mv(a, b)
		f.append(mlMk("pending", "t7", "t7", a, b, x))
		f.wantBy("destination", b, true)
	})

	t.Run("latest pending line, tree at its old path (kill before the move)", func(t *testing.T) {
		f := mlNew(t)
		a := f.add(f.old("t7"))
		x := f.ident(a)
		f.append(mlMk("pending", "t7", "t7", a, f.nw("t7"), x))
		// M8S (card t1640 review F5): a pending line is an UNSETTLED move —
		// the tree may stand at either of its two paths, and the cleanup
		// protection covers whichever end is registered until a later
		// restored or reversed line releases it. A kill BEFORE the move
		// leaves the tree at its old path, still protected.
		f.wantBy("source", a, true)
	})

	t.Run("the tree's own later pending line does not drop the protection", func(t *testing.T) {
		f, a, b, x := mlMovedTree(t)
		f.append(mlMk("pending", "t7", "t7", a, b, x))
		f.wantBy("destination after an own pending line", b, true)
	})

	for _, reason := range []string{"source path occupied", "identity mismatch"} {
		t.Run("the tree's own refused line ("+reason+") does not drop the protection", func(t *testing.T) {
			f, a, b, x := mlMovedTree(t)
			refused := mlMk("refused", "t7", "t7", a, b, x)
			refused.Reason = reason
			f.append(refused)
			f.wantBy("destination after an own refused line", b, true)
		})
	}

	t.Run("failed and unparseable lines are ignored", func(t *testing.T) {
		f, a, b, x := mlMovedTree(t)
		failed := mlMk("failed", "t7", "t7", a, b, x)
		failed.Reason = "state re-key failed"
		f.append(failed)
		f.appendRaw("{ not json\n")
		f.appendRaw(`{"schema":1,"result":`)
		f.wantBy("destination after failed and unparseable lines", b, true)
	})
}
