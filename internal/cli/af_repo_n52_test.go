package cli

// N52 (audit of 1c05083f71): the repository moved on disk and nothing finishes the card, because the ledger is bound to
// an absolute git directory and no verb changes that. `moai af move-repository` records the move, and only when it is
// one: the old git directory is gone from disk, the new repository holds every commit the rows name, and the caller
// gave a reason. Real ledger, real git, the real verbs. The refusals are asserted against a snapshot of everything the
// ledger holds (rows, current identity, recorded moves), and each has a control in the same test: the same call goes
// through once the one thing it was refused for is put right.

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/mergeq"
)

// afN52 is the landed card of N28 whose repository was renamed on disk: the ledger is bound to the old git directory and
// the project directory is the new location.
type afN52 struct {
	afN3Card
	moved string // the new location of the repository
	oldID string // the identity the ledger was bound to
}

func afN52Renamed(t *testing.T) afN52 {
	t.Helper()
	a := afN52{afN3Card: afN28Landed(t)}
	a.oldID = a.bound()
	a.moved = a.repo + "-moved"
	if err := os.Rename(a.repo, a.moved); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLAUDE_PROJECT_DIR", a.moved)
	return a
}

// move runs `moai af move-repository` and reports a refusal as a value: the refusals here are what the tests are about.
func (a afN52) move(args ...string) (string, error) {
	a.t.Helper()
	return a.run(append([]string{"move-repository"}, args...)...)
}

// afN52State is everything the ledger holds that a refused call must leave as it was.
type afN52State struct {
	Rows  afLedgerSnapshot
	Ident string
	Moves []ledger.RepositoryMove
}

func (a afN52) state() afN52State {
	a.t.Helper()
	moves, err := a.st.RepositoryMoves(context.Background())
	if err != nil {
		a.t.Fatal(err)
	}
	return afN52State{Rows: afSnapshot(a.t, a.st), Ident: a.bound(), Moves: moves}
}

// wantRefused checks that the call was refused with the words, and that the ledger is exactly as it was.
func (a afN52) wantRefused(before afN52State, out string, err error, words ...string) {
	a.t.Helper()
	if err == nil {
		a.t.Fatalf("the verb was accepted, want a refusal that says %q: %.300s", words, out)
	}
	text := err.Error() + " " + out
	for _, w := range words {
		if !strings.Contains(text, w) {
			a.t.Errorf("the refusal does not say %q: %.500s", w, text)
		}
	}
	if after := a.state(); !reflect.DeepEqual(before, after) {
		a.t.Errorf("the refused call changed the ledger:\nbefore %+v\nafter  %+v", before, after)
	}
}

func (a afN52) wantMoved(out string, err error, newRepo string) {
	a.t.Helper()
	if err != nil {
		a.t.Fatalf("the move was refused: %v\n%.500s", err, out)
	}
	if got, want := a.bound(), afN38GitDir(a.t, newRepo); got != want {
		a.t.Fatalf("the ledger is bound to %q after the move, want %q", got, want)
	}
}

func (a afN52) inTrunkOf(repo, sha string) bool {
	return exec.Command("git", "-C", repo, "merge-base", "--is-ancestor", sha, mergeq.DefaultTrunk).Run() == nil
}

// afN52Fresh is a repository with one base commit and nothing of the card's.
func afN52Fresh(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	afGit(t, dir, "init", "-q", "-b", "work")
	if err := os.WriteFile(filepath.Join(dir, "other.txt"), []byte("other\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	afGit(t, dir, "add", "other.txt")
	afGit(t, dir, "commit", "-q", "-m", "other base")
	return dir
}

// (1) The whole path: bound to A, A renamed, every verb refused, the verb run, the card goes on to DONE with its commit in
// the trunk of the new location.
func TestAFVerbs_N52ARenamedRepositoryIsRecordedAndTheCardGoesOnToDone(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN52Renamed(t)

	before := a.state()
	for _, args := range [][]string{
		{"move", "t1", "--epoch", a.e2, "--stage", "REVIEWING"},
		{"release", "t1", "--epoch", a.e2},
		{"reap"},
		{"add", "t9"},
	} {
		out, err := a.run(args...)
		a.wantRefused(before, out, err, "bound to the repository "+a.oldID)
	}
	if err := a.tryFinish(); err == nil {
		t.Fatal("the card finished from the new location before the move was recorded")
	}
	if got := a.stage(); got == core.StageDone {
		t.Fatal("the card is DONE before the move was recorded")
	}

	out, err := a.move("--reason", "renamed the checkout")
	a.wantMoved(out, err, a.moved)
	var res struct {
		From, To, Reason string
		CommitsChecked   int `json:"commits_checked"`
	}
	if jerr := json.Unmarshal([]byte(out), &res); jerr != nil {
		t.Fatalf("stdout %.300q is not the verb's JSON line: %v", out, jerr)
	}
	if res.From != a.oldID || res.To != afN38GitDir(t, a.moved) || res.Reason != "renamed the checkout" || res.CommitsChecked < 1 {
		t.Errorf("the verb printed %+v, want from %q to %q with the reason and at least one commit checked", res, a.oldID, afN38GitDir(t, a.moved))
	}

	if err := a.tryFinish(); err != nil {
		t.Fatalf("after the move the card is still refused: %v", err)
	}
	if got := a.stage(); got != core.StageDone {
		t.Errorf("stage %s, want DONE", got)
	}
	if !a.inTrunkOf(a.moved, a.c1) {
		t.Error("the card is DONE and c1 is not in the trunk of the new location")
	}
}

// (2) The old git directory still exists: this is a second repository, not a move.
func TestAFVerbs_N52TheMoveIsRefusedWhileTheOldRepositoryStillExists(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN52Renamed(t)
	// the repository the ledger was bound to comes back at its old path as a clone of the new one (it has every commit)
	if out, err := exec.Command("git", "clone", "-q", "--no-hardlinks", a.moved, a.repo).CombinedOutput(); err != nil {
		t.Fatalf("clone: %v %s", err, out)
	}
	before := a.state()
	out, err := a.move("--reason", "a second checkout", "--bind-repo", a.moved)
	a.wantRefused(before, out, err, "still exists", a.oldID)

	// the control: the old path is gone, and the very same call goes through
	if err := os.RemoveAll(a.repo); err != nil {
		t.Fatal(err)
	}
	out, err = a.move("--reason", "a second checkout", "--bind-repo", a.moved)
	a.wantMoved(out, err, a.moved)
}

// (3) The new repository lacks a commit the rows name: it is not the repository the ledger belongs to.
func TestAFVerbs_N52TheMoveIsRefusedWhenTheNewRepositoryLacksACommitTheRowsName(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN52Renamed(t)
	// A ledger bound before the identity marker existed has no marker to compare, so the commits are its only evidence
	// (N61: a marked ledger is refused by the marker instead, see af_repo_identity_test.go).
	afIDForget(t, a.db)
	other := afN52Fresh(t)
	t.Setenv("CLAUDE_PROJECT_DIR", other)
	before := a.state()
	out, err := a.move("--reason", "moved")
	a.wantRefused(before, out, err, a.c1, "does not have")

	// the control: the repository that has the commits, declared, and the same call goes through
	out, err = a.move("--reason", "moved", "--bind-repo", a.moved)
	a.wantMoved(out, err, a.moved)
}

// (4) A reason is required: absent, empty, blank.
func TestAFVerbs_N52TheMoveNeedsAReason(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN52Renamed(t)
	before := a.state()
	for _, args := range [][]string{{}, {"--reason", ""}, {"--reason", "  \t "}} {
		out, err := a.move(args...)
		a.wantRefused(before, out, err, "--reason")
	}
	// a ledger that does not exist is not created by a call that was refused before it was opened
	missing := filepath.Join(t.TempDir(), "af", "none.db")
	if out, _, err := afRunBoth("move-repository", "--db", missing); err == nil {
		t.Errorf("a call without a reason was accepted: %.200s", out)
	}
	if _, err := os.Stat(missing); err == nil {
		t.Error("a call refused for its missing reason created a ledger file")
	}
	// the control: with a reason the call goes through
	out, err := a.move("--reason", "renamed")
	a.wantMoved(out, err, a.moved)
}

// (5) After the move the other verbs compare with the CURRENT identity: the new one passes, the old one is refused.
func TestAFVerbs_N52AfterTheMoveTheOtherVerbsReadTheCurrentIdentity(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN52Renamed(t)
	out, err := a.move("--reason", "renamed the checkout")
	a.wantMoved(out, err, a.moved)
	newID := afN38GitDir(t, a.moved)
	ctx := context.Background()

	if err := a.st.BindRepository(ctx, a.oldID); !errors.Is(err, ledger.ErrWrongRepository) {
		t.Errorf("binding the old identity after the move: %v, want ErrWrongRepository", err)
	}
	if err := a.st.BindRepository(ctx, newID); err != nil {
		t.Errorf("binding the new identity after the move: %v", err)
	}

	// a verb run from the old location is refused: a repository at the old path is the identity the ledger left
	if err := os.MkdirAll(a.repo, 0o755); err != nil {
		t.Fatal(err)
	}
	afGit(t, a.repo, "init", "-q", "-b", "work")
	t.Setenv("CLAUDE_PROJECT_DIR", a.repo)
	before := a.state()
	out, err = a.run("add", "t2")
	a.wantRefused(before, out, err, "bound to the repository "+newID, a.oldID)

	// and from the new project it passes
	t.Setenv("CLAUDE_PROJECT_DIR", a.moved)
	if out, err = a.run("add", "t2"); err != nil {
		t.Fatalf("add from the new project after the move: %v\n%.300s", err, out)
	}
	if cards, err := a.st.Cards(ctx); err != nil || len(cards) != 2 {
		t.Errorf("cards %d %v, want the added card beside t1", len(cards), err)
	}
	// a verb that looks a commit up passes from the new project as well
	a.verb("evidence", "t1", "--epoch", a.e2, "--kind", "green", "--subject-sha", a.c1, "--result", "pass", "--repo", a.moved)
}

// (6) The move is listed with its reason, from the identity it left to the one it went to.
func TestAFVerbs_N52TheRecordedMoveIsListedWithItsReason(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN52Renamed(t)
	if moves := a.state().Moves; len(moves) != 0 {
		t.Fatalf("moves before the verb: %+v", moves)
	}
	out, err := a.move("--reason", "moved the checkout to a bigger disk")
	a.wantMoved(out, err, a.moved)
	moves := a.state().Moves
	if len(moves) != 1 {
		t.Fatalf("moves %+v, want exactly one", moves)
	}
	if m := moves[0]; m.From != a.oldID || m.To != afN38GitDir(t, a.moved) || m.Reason != "moved the checkout to a bigger disk" || m.At.IsZero() {
		t.Errorf("the recorded move is %+v, want from %q to %q with the reason as given", m, a.oldID, afN38GitDir(t, a.moved))
	}
}

// (7) The help says what the verb checks.
func TestAFVerbs_N52TheHelpSaysWhatTheMoveChecks(t *testing.T) {
	flat := afHelpOf(t, "move-repository")
	for _, want := range []string{
		"no longer exists on disk", // the old git directory
		"holds every commit",       // the new repository
		"--reason",
		"--bind-repo",
		"ledger unchanged",
		"does not bind",
	} {
		if !strings.Contains(flat, want) {
			t.Errorf("`moai af move-repository --help` does not say %q:\n%s", want, flat)
		}
	}
}

// A ledger bound to nothing has nothing to move: the verb does not bind it to the project's repository.
func TestAFVerbs_N52AnUnboundLedgerHasNothingToMove(t *testing.T) {
	afSkipWithoutSh(t)
	ctx := context.Background()
	db := filepath.Join(t.TempDir(), "af", "ledger.db")
	st, err := ledger.Open(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if err := st.AddCard(ctx, ledger.Card{ID: "t1", Title: "t1", Position: core.Position{Stage: core.StageReady}, Spec: ledger.Spec{Class: "A", Paths: []string{"a.txt"}}}); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLAUDE_PROJECT_DIR", afN52Fresh(t))
	_, _, rerr := afRunBoth("move-repository", "--db", db, "--reason", "moved")
	if rerr == nil || !strings.Contains(rerr.Error(), "no repository to move") {
		t.Errorf("error %v, want a refusal that says there is no repository to move", rerr)
	}
	if got, err := st.Repository(ctx); err != nil || got != "" {
		t.Errorf("the ledger is bound to %q (%v) after a refused move, want unbound", got, err)
	}
	if moves, err := st.RepositoryMoves(ctx); err != nil || len(moves) != 0 {
		t.Errorf("moves %+v %v", moves, err)
	}
}

// A new location that is no git repository holds no commits.
func TestAFVerbs_N52TheNewLocationMustBeAGitRepository(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN52Renamed(t)
	before := a.state()
	out, err := a.move("--reason", "moved", "--bind-repo", t.TempDir())
	a.wantRefused(before, out, err, "--bind-repo", "git repository")
	out, err = a.move("--reason", "moved")
	a.wantMoved(out, err, a.moved)
}

// --bind-repo is the new repository: it wins over the project directory, which may be anything.
func TestAFVerbs_N52TheDeclaredRepositoryIsTheNewOneNotTheProjectsOwn(t *testing.T) {
	afSkipWithoutSh(t)
	a := afN52Renamed(t)
	other := afN52Fresh(t)
	t.Setenv("CLAUDE_PROJECT_DIR", other)
	out, err := a.move("--reason", "moved", "--bind-repo", a.moved)
	a.wantMoved(out, err, a.moved)
	if got := a.bound(); got == afN38GitDir(t, other) {
		t.Errorf("the ledger is bound to the project's own repository %q", got)
	}
}
