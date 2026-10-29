package gitref

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

var _ claim.ControlRef = (*Ref)(nil)

// hermetic keeps the user's git configuration out of the fixture.
var hermetic = []string{
	"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1",
	"GIT_AUTHOR_NAME=fixture", "GIT_AUTHOR_EMAIL=fixture@example.invalid",
	"GIT_COMMITTER_NAME=fixture", "GIT_COMMITTER_EMAIL=fixture@example.invalid",
}

func run(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), hermetic...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// fixture is a local bare repository holding a genesis commit. Nothing here
// reaches a real remote: every URL is a path under t.TempDir().
type fixture struct {
	t      *testing.T
	remote string
}

func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
}

func newFixture(t *testing.T, withGenesis bool) *fixture {
	t.Helper()
	requireGit(t)
	root := t.TempDir()
	remote := filepath.Join(root, "remote.git")
	run(t, root, "init", "--bare", "--initial-branch=main", remote)
	f := &fixture{t: t, remote: remote}
	if withGenesis {
		work := filepath.Join(root, "genesis")
		run(t, root, "init", "--initial-branch=main", work)
		if err := os.MkdirAll(filepath.Join(work, "control"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(work, "control", "schema.json"), []byte(`{"schema":"moai-control","schema_version":1}`+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		run(t, work, "add", "control/schema.json")
		run(t, work, "commit", "-m", "genesis")
		run(t, work, "push", remote, "HEAD:refs/heads/moai-control")
	}
	return f
}

func (f *fixture) ref() *Ref {
	f.t.Helper()
	dir := filepath.Join(f.t.TempDir(), "clone")
	run(f.t, filepath.Dir(dir), "init", "--initial-branch=main", dir)
	if !strings.HasPrefix(f.remote, os.TempDir()) && !strings.Contains(f.remote, "/T/") && !strings.Contains(f.remote, "/tmp") {
		f.t.Fatalf("fixture remote %q is not a temporary path", f.remote)
	}
	r := New(dir, f.remote)
	r.Env = hermetic
	return r
}

func (f *fixture) head() string {
	return run(f.t, f.remote, "rev-parse", "refs/heads/moai-control")
}

var bg = context.Background()

func TestContractMatchesTheAuthorityInterface(t *testing.T) {
	r := newFixture(t, true).ref()
	if _, _, err := r.Read(bg, "issue/7"); !errors.Is(err, core.ErrNotHeld) {
		t.Errorf("missing key: %v", err)
	}
	if _, err := r.CompareAndSwap(bg, "issue/7", 3, []byte(`{}`)); !errors.Is(err, core.ErrRevisionConflict) {
		t.Errorf("create with a wrong expectation: %v", err)
	}
	rev, err := r.CompareAndSwap(bg, "issue/7", 0, []byte(`{"state":"claimed"}`))
	if err != nil || rev != 1 {
		t.Fatalf("create: %d %v", rev, err)
	}
	if _, err := r.CompareAndSwap(bg, "issue/7", 0, []byte(`{}`)); !errors.Is(err, core.ErrRevisionConflict) {
		t.Errorf("second create: %v", err)
	}
	data, rev, err := r.Read(bg, "issue/7")
	if err != nil || rev != 1 || string(data) != `{"state":"claimed"}` {
		t.Errorf("read: %q %d %v", data, rev, err)
	}
	if rev, err = r.CompareAndSwap(bg, "issue/7", 1, []byte(`{"state":"implementing"}`)); err != nil || rev != 2 {
		t.Fatalf("update: %d %v", rev, err)
	}
	_, _ = r.CompareAndSwap(bg, "request/req-1", 0, []byte(`{}`))
	_, _ = r.CompareAndSwap(bg, "agent/1", 0, []byte(`{}`))
	keys, err := r.Keys(bg, "")
	if err != nil || strings.Join(keys, ",") != "agent/1,issue/7,request/req-1" {
		t.Errorf("keys: %v %v", keys, err)
	}
	if keys, _ := r.Keys(bg, "issue/"); len(keys) != 1 || keys[0] != "issue/7" {
		t.Errorf("prefix keys: %v", keys)
	}
	if err := r.Delete(bg, "issue/7", 9); !errors.Is(err, core.ErrRevisionConflict) {
		t.Errorf("delete with a wrong revision: %v", err)
	}
	if err := r.Delete(bg, "issue/7", 2); err != nil {
		t.Errorf("delete: %v", err)
	}
	if err := r.Delete(bg, "issue/7", 2); !errors.Is(err, core.ErrNotHeld) {
		t.Errorf("delete of a missing key: %v", err)
	}
	if _, _, err := r.Read(bg, "issue/7"); !errors.Is(err, core.ErrNotHeld) {
		t.Errorf("deleted key still readable: %v", err)
	}
}

func TestKeysAreValidated(t *testing.T) {
	r := newFixture(t, true).ref()
	for _, bad := range []string{"", "issue", "issue/", "issue/../x", "other/1", "issue/A", "issue/1/2", "../issue/1", "issue/ 1"} {
		if _, _, err := r.Read(bg, bad); !errors.Is(err, core.ErrIDInvalid) {
			t.Errorf("read %q: %v", bad, err)
		}
		if _, err := r.CompareAndSwap(bg, bad, 0, []byte(`{}`)); !errors.Is(err, core.ErrIDInvalid) {
			t.Errorf("write %q: %v", bad, err)
		}
	}
}

// Two clones racing to create the same key: exactly one wins.
func TestRacingCreatesHaveExactlyOneWinner(t *testing.T) {
	fx := newFixture(t, true)
	refs := []*Ref{fx.ref(), fx.ref()}
	var wins, conflicts atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := refs[i%2].CompareAndSwap(bg, "issue/7", 0, []byte(fmt.Sprintf(`{"by":%d}`, i)))
			switch {
			case err == nil:
				wins.Add(1)
			case errors.Is(err, core.ErrRevisionConflict):
				conflicts.Add(1)
			default:
				t.Errorf("unexpected error: %v", err)
			}
		}(i)
	}
	wg.Wait()
	if wins.Load() != 1 || conflicts.Load() != 7 {
		t.Errorf("wins=%d conflicts=%d", wins.Load(), conflicts.Load())
	}
}

// Writes to different keys from different clones all land: a moved branch is
// retried, not reported as a conflict, and the history stays linear.
func TestDifferentKeysFromTwoClonesAllLandInALinearHistory(t *testing.T) {
	fx := newFixture(t, true)
	refs := []*Ref{fx.ref(), fx.ref()}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if _, err := refs[i%2].CompareAndSwap(bg, fmt.Sprintf("issue/%d", i+1), 0, []byte(`{}`)); err != nil {
				t.Errorf("write %d: %v", i, err)
			}
		}(i)
	}
	wg.Wait()
	keys, _ := refs[0].Keys(bg, "issue/")
	if len(keys) != 8 {
		t.Errorf("keys = %d", len(keys))
	}
	if n := run(t, fx.remote, "rev-list", "--count", "refs/heads/moai-control"); n != "9" {
		t.Errorf("commits = %s, want genesis plus one per write", n)
	}
	if merges := run(t, fx.remote, "rev-list", "--merges", "--count", "refs/heads/moai-control"); merges != "0" {
		t.Errorf("merge commits = %s", merges)
	}
}

func TestEveryWriteIsAnAppendOnlyCommitOnTheGenesis(t *testing.T) {
	fx := newFixture(t, true)
	genesis := fx.head()
	r := fx.ref()
	_, _ = r.CompareAndSwap(bg, "issue/1", 0, []byte(`{}`))
	_, _ = r.CompareAndSwap(bg, "issue/1", 1, []byte(`{"n":2}`))
	if parent := run(t, fx.remote, "rev-parse", "refs/heads/moai-control~2"); parent != genesis {
		t.Errorf("history does not descend from the genesis: %s vs %s", parent, genesis)
	}
	if got := run(t, fx.remote, "show", "refs/heads/moai-control~2:control/schema.json"); !strings.Contains(got, "moai-control") {
		t.Errorf("genesis file lost: %s", got)
	}
	if got := run(t, fx.remote, "show", "refs/heads/moai-control:control/schema.json"); !strings.Contains(got, "moai-control") {
		t.Errorf("writes must keep the schema file: %s", got)
	}
}

// A writer with push rights can rewrite history on an unprotected branch. The
// ref cannot prevent it; it must notice and refuse to write on top of it.
func TestRewrittenHistoryIsDetectedAndWritesAreRefused(t *testing.T) {
	fx := newFixture(t, true)
	a := fx.ref()
	if _, err := a.CompareAndSwap(bg, "issue/1", 0, []byte(`{}`)); err != nil {
		t.Fatal(err)
	}
	// Another clone force-pushes a divergent history over the branch.
	evil := filepath.Join(t.TempDir(), "evil")
	run(t, filepath.Dir(evil), "init", "--initial-branch=main", evil)
	run(t, evil, "commit", "--allow-empty", "-m", "divergent")
	run(t, evil, "push", "--force", fx.remote, "HEAD:refs/heads/moai-control")
	before := fx.head()
	if _, err := a.CompareAndSwap(bg, "issue/2", 0, []byte(`{}`)); !errors.Is(err, core.ErrHistoryRewritten) {
		t.Fatalf("want ErrHistoryRewritten, got %v", err)
	}
	if _, _, err := a.Read(bg, "issue/1"); !errors.Is(err, core.ErrHistoryRewritten) {
		t.Errorf("reads must also refuse a rewritten history: %v", err)
	}
	if fx.head() != before {
		t.Error("a write was pushed on top of a rewritten history")
	}
}

// The code never creates the branch: without a genesis commit it refuses.
func TestMissingBranchIsNeverCreated(t *testing.T) {
	fx := newFixture(t, false)
	r := fx.ref()
	if _, err := r.CompareAndSwap(bg, "issue/1", 0, []byte(`{}`)); !errors.Is(err, core.ErrStoreUnavailable) {
		t.Errorf("write without a genesis: %v", err)
	}
	if _, _, err := r.Read(bg, "issue/1"); !errors.Is(err, core.ErrStoreUnavailable) {
		t.Errorf("read without a genesis: %v", err)
	}
	if refs := run(t, fx.remote, "for-each-ref"); refs != "" {
		t.Errorf("the remote gained refs: %q", refs)
	}
}

func TestUnreachableRemoteIsStoreUnavailable(t *testing.T) {
	requireGit(t)
	r := New(filepath.Join(t.TempDir(), "clone"), filepath.Join(t.TempDir(), "does-not-exist.git"))
	r.Env = hermetic
	run(t, filepath.Dir(r.Dir), "init", "--initial-branch=main", r.Dir)
	if _, _, err := r.Read(bg, "issue/1"); !errors.Is(err, core.ErrStoreUnavailable) {
		t.Errorf("unreachable remote: %v", err)
	}
}

// The claim store runs on the git ref, and a second cache rebuilds from it.
func TestClaimStoreOnTheGitRefAcrossTwoCaches(t *testing.T) {
	fx := newFixture(t, true)
	s1, err := claim.OpenWithLimit(t.TempDir(), 5)
	if err != nil {
		t.Fatal(err)
	}
	s1.Attach(fx.ref())
	issue, _ := core.NewIssueRef(7)
	rec, err := s1.Acquire(issue, "holder-a", "d1", "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s1.Update(issue, "holder-a", rec.Epoch, func(r *claim.Record) error { r.State = core.StateImplementing; return nil }); err != nil {
		t.Fatal(err)
	}

	s2, _ := claim.OpenWithLimit(t.TempDir(), 5)
	s2.Attach(fx.ref())
	n, err := s2.Rebuild(bg)
	if err != nil || n != 1 {
		t.Fatalf("rebuild: %d %v", n, err)
	}
	got, err := s2.Load(issue)
	if err != nil || got.State != core.StateImplementing || got.Rev != 2 {
		t.Fatalf("restored: %+v %v", got, err)
	}
	// s1's cache is now one revision behind another writer's work: its write is refused.
	if _, err := s2.Update(issue, "holder-a", rec.Epoch, func(r *claim.Record) error { r.State = core.StateVerifying; return nil }); err != nil {
		t.Fatal(err)
	}
	_, err = s1.Update(issue, "holder-a", rec.Epoch, func(r *claim.Record) error { r.State = core.StateBlocked; return nil })
	if !errors.Is(err, core.ErrRevisionConflict) {
		t.Errorf("a stale cache must be refused: %v", err)
	}
}

// Each compare-and-swap costs a bounded number of git invocations; this records
// the figure the latency report quotes.
func TestInvocationCountPerWriteIsBounded(t *testing.T) {
	r := newFixture(t, true).ref()
	before := r.Invocations()
	if _, err := r.CompareAndSwap(bg, "issue/1", 0, []byte(`{}`)); err != nil {
		t.Fatal(err)
	}
	if n := r.Invocations() - before; n > 12 {
		t.Errorf("one uncontended write used %d git invocations", n)
	} else {
		t.Logf("uncontended write: %d git invocations", n)
	}
}

func TestUnknownSchemaVersionIsRefusedForReadsAndWrites(t *testing.T) {
	fx := newFixture(t, true)
	work := filepath.Join(t.TempDir(), "w")
	run(t, filepath.Dir(work), "clone", "--branch", "moai-control", fx.remote, work)
	if err := os.WriteFile(filepath.Join(work, "control", "schema.json"), []byte(`{"schema":"moai-control","schema_version":2}`), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, work, "commit", "-am", "bump schema")
	run(t, work, "push", "origin", "HEAD:refs/heads/moai-control")
	r := fx.ref()
	if _, _, err := r.Read(bg, "issue/1"); !errors.Is(err, core.ErrStoreUnavailable) {
		t.Errorf("read under an unknown schema: %v", err)
	}
	before := fx.head()
	if _, err := r.CompareAndSwap(bg, "issue/1", 0, []byte(`{}`)); !errors.Is(err, core.ErrStoreUnavailable) {
		t.Errorf("write under an unknown schema: %v", err)
	}
	if fx.head() != before {
		t.Error("a write reached a branch with an unknown schema")
	}
}

func TestMissingSchemaFileIsRefused(t *testing.T) {
	fx := newFixture(t, true)
	work := filepath.Join(t.TempDir(), "w")
	run(t, filepath.Dir(work), "clone", "--branch", "moai-control", fx.remote, work)
	run(t, work, "rm", "control/schema.json")
	run(t, work, "commit", "-m", "drop schema")
	run(t, work, "push", "origin", "HEAD:refs/heads/moai-control")
	if _, _, err := fx.ref().Read(bg, "issue/1"); !errors.Is(err, core.ErrStoreUnavailable) {
		t.Errorf("a branch without control/schema.json is not a control ref: %v", err)
	}
}

func TestCommitsCarryTheConfiguredAuthorAndTheIssueNumber(t *testing.T) {
	fx := newFixture(t, true)
	r := fx.ref()
	r.Author = "Ada Lovelace <ada@example.invalid>"
	if _, err := r.CompareAndSwap(bg, "issue/1", 0, []byte(`{}`)); err != nil {
		t.Fatal(err)
	}
	if got := run(t, fx.remote, "log", "-1", "--format=%an|%ae|%cn", "refs/heads/moai-control"); got != "Ada Lovelace|ada@example.invalid|Ada Lovelace" {
		t.Errorf("commit identity = %q", got)
	}
	if msg := run(t, fx.remote, "log", "-1", "--format=%B", "refs/heads/moai-control"); !strings.Contains(msg, "Issue #1948") || !strings.Contains(msg, "control: issue/1") {
		t.Errorf("commit message = %q", msg)
	}
}

func TestNonJSONDataAndBadDeleteKeysAreRefusedBeforeAnyGitCall(t *testing.T) {
	r := newFixture(t, true).ref()
	before := r.Invocations()
	if _, err := r.CompareAndSwap(bg, "issue/1", 0, []byte(`not json`)); !errors.Is(err, core.ErrIDInvalid) {
		t.Errorf("non-JSON data: %v", err)
	}
	if err := r.Delete(bg, "../issue/1", 1); !errors.Is(err, core.ErrIDInvalid) {
		t.Errorf("delete with a bad key: %v", err)
	}
	if r.Invocations() != before {
		t.Error("validation must happen before git is run")
	}
}

// A rejection that is not "the branch moved" is reported, not retried forever.
func TestPushDeclinedByTheRemoteIsStoreUnavailable(t *testing.T) {
	fx := newFixture(t, true)
	hook := filepath.Join(fx.remote, "hooks", "pre-receive")
	if err := os.WriteFile(hook, []byte("#!/bin/sh\necho declined >&2\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	r := fx.ref()
	before := fx.head()
	if _, err := r.CompareAndSwap(bg, "issue/1", 0, []byte(`{}`)); !errors.Is(err, core.ErrStoreUnavailable) {
		t.Fatalf("declined push: %v", err)
	}
	if fx.head() != before {
		t.Error("the remote moved despite the declined push")
	}
}

func TestContextCancellationStopsTheCall(t *testing.T) {
	r := newFixture(t, true).ref()
	ctx, cancel := context.WithCancel(bg)
	cancel()
	if _, _, err := r.Read(ctx, "issue/1"); !errors.Is(err, core.ErrStoreUnavailable) {
		t.Errorf("cancelled read: %v", err)
	}
}

// With no identity configured the commit is written as the noreply machine
// identity, never as whatever user.name/user.email the git configuration holds.
func TestDefaultCommitIdentityIsNoreplyNotThePersonalGitConfig(t *testing.T) {
	fx := newFixture(t, true)
	cfg := filepath.Join(t.TempDir(), "gitconfig")
	if err := os.WriteFile(cfg, []byte("[user]\n\tname = Real Person\n\temail = real.person@example.com\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	r := fx.ref()
	// A personal configuration, not the hermetic one, and no GIT_* identity variables.
	r.Env = []string{"GIT_CONFIG_GLOBAL=" + cfg, "GIT_CONFIG_NOSYSTEM=1",
		"GIT_AUTHOR_NAME=", "GIT_AUTHOR_EMAIL=", "GIT_COMMITTER_NAME=", "GIT_COMMITTER_EMAIL="}
	if _, err := r.CompareAndSwap(bg, "issue/1", 0, []byte(`{}`)); err != nil {
		t.Fatal(err)
	}
	got := run(t, fx.remote, "log", "-1", "--format=%an|%ae|%cn|%ce", "refs/heads/moai-control")
	if strings.Contains(got, "Real Person") || strings.Contains(got, "real.person") {
		t.Fatalf("the personal git identity leaked into the commit: %s", got)
	}
	want := DefaultIdentity
	name, email, _ := strings.Cut(strings.TrimSuffix(want, ">"), " <")
	if got != fmt.Sprintf("%s|%s|%s|%s", name, email, name, email) {
		t.Errorf("default identity = %q, want %q", got, want)
	}
	if !strings.Contains(email, "noreply") {
		t.Errorf("the default must be a noreply address, got %q", email)
	}
}

func TestAuthorAndCommitterAreConfigurableSeparately(t *testing.T) {
	fx := newFixture(t, true)
	r := fx.ref()
	r.Author = "Control Bot <bot@users.noreply.invalid>"
	r.Committer = "Control Service <svc@users.noreply.invalid>"
	if _, err := r.CompareAndSwap(bg, "issue/1", 0, []byte(`{}`)); err != nil {
		t.Fatal(err)
	}
	got := run(t, fx.remote, "log", "-1", "--format=%an|%ae|%cn|%ce", "refs/heads/moai-control")
	if got != "Control Bot|bot@users.noreply.invalid|Control Service|svc@users.noreply.invalid" {
		t.Errorf("identities = %q", got)
	}
}

func TestMalformedIdentitiesAreRefusedBeforeAnyGitCall(t *testing.T) {
	for _, bad := range []string{"just a name", "Name <>", "<a@b.invalid>", "Name <a@b.invalid", "Name <a@b.invalid>\nInjected: x", "Na<me <a@b.invalid>"} {
		r := newFixture(t, true).ref()
		r.Author = bad
		before := r.Invocations()
		if _, err := r.CompareAndSwap(bg, "issue/1", 0, []byte(`{}`)); !errors.Is(err, core.ErrIDInvalid) {
			t.Errorf("author %q: %v", bad, err)
		}
		if r.Invocations() != before {
			t.Errorf("author %q: git ran before the identity was validated", bad)
		}
		r2 := newFixture(t, true).ref()
		r2.Committer = bad
		if _, err := r2.CompareAndSwap(bg, "issue/1", 0, []byte(`{}`)); !errors.Is(err, core.ErrIDInvalid) {
			t.Errorf("committer %q: %v", bad, err)
		}
	}
}

// The keyed documents (intents, audit rounds, the context ledger) use the same
// compare-and-swap line as claims, and unknown kinds are still refused.
func TestDocumentKindsAreStoredAndUnknownKindsRefused(t *testing.T) {
	r := newFixture(t, true).ref()
	for _, key := range []string{"intent/9f2c0a", "audit/issue-1948", "context/issue-1948"} {
		if rev, err := r.CompareAndSwap(bg, key, 0, []byte(`{"n":1}`)); err != nil || rev != 1 {
			t.Errorf("%s: %d %v", key, rev, err)
		}
	}
	keys, _ := r.Keys(bg, "")
	if strings.Join(keys, ",") != "audit/issue-1948,context/issue-1948,intent/9f2c0a" {
		t.Errorf("keys = %v", keys)
	}
	for _, bad := range []string{"permit/issue-1", "docs/issue-1", "intent/Bad"} {
		if _, err := r.CompareAndSwap(bg, bad, 0, []byte(`{}`)); !errors.Is(err, core.ErrIDInvalid) {
			t.Errorf("%s: %v", bad, err)
		}
	}
}
