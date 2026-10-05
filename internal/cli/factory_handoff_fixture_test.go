package cli

// factory_handoff_fixture_test.go — the lane-handoff fixture that the kept
// operator command (`moai factory handoff abandon-lane`) is tested against.
//
// It moved here from the retired lane-handoff CLI tests
// (SPEC-CODEX-FACTORY-RETIRE-001 M3). The handoff states are now driven through
// the factorymsg Store API instead of the deleted preparation CLI, and the
// target is still a real git worktree on a real branch, so the abandon test's
// "target preserved" assertion measures something.

import (
	"context"
	"database/sql"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

const (
	handoffTestSpec   = "SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001"
	handoffTestSlot   = "lane-1"
	handoffTestCard   = "t1082"
	handoffTestBranch = "WT-lane-handoff"
)

// laneHandoffFixture is a real git repository (main + develop) registered as
// an active factory run, with lane-1 seeded through the production launcher
// path. The process cwd is the repository because the operator command
// resolves the project from it.
type laneHandoffFixture struct {
	primary, run, developPin, mainSHA string
	store                             *factorymsg.Store
	source                            factorymsg.Peer
}

func handoffGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// newLaneHandoffFixture builds the fixture. bindSource=false leaves lane-1 in
// the launcher provisional (launch-pending) state.
func newLaneHandoffFixture(t *testing.T, bindSource bool) *laneHandoffFixture {
	t.Helper()
	t.Setenv("MOAI_HOME", t.TempDir())
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(t.TempDir(), "gitconfig"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	for k, v := range map[string]string{
		"GIT_AUTHOR_NAME": "t1082", "GIT_AUTHOR_EMAIL": "t1082@example.invalid",
		"GIT_COMMITTER_NAME": "t1082", "GIT_COMMITTER_EMAIL": "t1082@example.invalid",
	} {
		t.Setenv(k, v)
	}
	repo := filepath.Join(t.TempDir(), "repo")
	if out, err := exec.Command("git", "init", "-q", "-b", "main", repo).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
	if err := os.WriteFile(filepath.Join(repo, "seed.txt"), []byte("seed\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	handoffGit(t, repo, "add", "seed.txt")
	handoffGit(t, repo, "commit", "-qm", "seed")
	f := &laneHandoffFixture{run: "run-handoff"}
	f.mainSHA = handoffGit(t, repo, "rev-parse", "HEAD")
	handoffGit(t, repo, "checkout", "-q", "-b", "develop")
	if err := os.WriteFile(filepath.Join(repo, "develop.txt"), []byte("develop\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	handoffGit(t, repo, "add", "develop.txt")
	handoffGit(t, repo, "commit", "-qm", "develop")
	f.developPin = handoffGit(t, repo, "rev-parse", "HEAD")
	handoffGit(t, repo, "checkout", "-q", "main")
	// Config and L1 trees are not source changes.
	if err := os.WriteFile(filepath.Join(repo, ".git", "info", "exclude"), []byte(".moai/\n.claude/worktrees/\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(repo)

	f.primary = homestate.CanonicalProjectRoot(repo)
	fdb, err := homestate.OpenFactory(f.primary)
	if err != nil {
		t.Fatal(err)
	}
	if err := fdb.RecordRun(context.Background(), homestate.FactoryRun{RunID: f.run, Backend: "codex"}); err != nil {
		t.Fatal(err)
	}
	_ = fdb.Close()
	f.store, err = factorymsg.Open(f.primary, f.run)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = f.store.Close() })
	pending, err := f.store.RegisterLaunchPending(context.Background(), factorymsg.Peer{
		ProjectKey: "project", RunID: f.run, Backend: "codex", Role: "worker", Slot: handoffTestSlot,
		PID: os.Getpid(), ProcessStart: "fake-source-start",
	})
	if err != nil {
		t.Fatal(err)
	}
	f.source = pending
	if bindSource {
		f.bindSource(t)
	}
	return f
}

func (f *laneHandoffFixture) bindSource(t *testing.T) {
	t.Helper()
	bound := f.source
	bound.SessionUUID = "src-uuid"
	got, ok, err := f.store.BindLaunchPending(context.Background(), bound)
	if err != nil || !ok {
		t.Fatalf("bind source ok=%v err=%v", ok, err)
	}
	f.source = got
}

func (f *laneHandoffFixture) target(card string) string {
	return filepath.Join(f.primary, ".claude", "worktrees", card)
}

// reserve commits the handoff reservation. No target exists yet, exactly as
// in production (the reservation is taken before the tree is created).
func (f *laneHandoffFixture) reserve(t *testing.T, mode string) factorymsg.Handoff {
	t.Helper()
	h, err := f.store.ReserveHandoff(context.Background(), factorymsg.HandoffReservation{
		Slot: handoffTestSlot, CardID: handoffTestCard, SpecID: handoffTestSpec, Mode: mode,
		DevelopPin: f.developPin, TargetPath: f.target(handoffTestCard), TargetBranch: handoffTestBranch,
	})
	if err != nil {
		t.Fatal(err)
	}
	return h
}

// materializeTarget creates the reserved target as a real git worktree on the
// reserved branch, based on the develop pin.
func (f *laneHandoffFixture) materializeTarget(t *testing.T, h factorymsg.Handoff) {
	t.Helper()
	handoffGit(t, f.primary, "worktree", "add", "-q", "-b", h.TargetBranch, h.TargetPath, h.DevelopPin)
}

// wtReady drives a handoff to WT_READY: reserve, create the real target
// worktree and branch, then mark it ready through the Store API.
func (f *laneHandoffFixture) wtReady(t *testing.T, mode string) factorymsg.Handoff {
	t.Helper()
	h := f.reserve(t, mode)
	f.materializeTarget(t, h)
	h, err := f.store.MarkHandoffWTReady(context.Background(), h)
	if err != nil || h.State != factorymsg.HandoffWTReady {
		t.Fatalf("mark WT_READY = %+v err=%v", h, err)
	}
	return h
}

// requireTargetMaterialized asserts the target worktree path exists and its
// branch ref resolves in the primary repository.
func (f *laneHandoffFixture) requireTargetMaterialized(t *testing.T, target, branch string) {
	t.Helper()
	if !pathExists(target) {
		t.Fatalf("target worktree %s does not exist", target)
	}
	cmd := exec.Command("git", "-C", f.primary, "rev-parse", "--verify", "-q", "refs/heads/"+branch)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("target branch %s does not resolve: %v %s", branch, err, out)
	}
}

func (f *laneHandoffFixture) storedHandoff(t *testing.T, id string) factorymsg.Handoff {
	t.Helper()
	hs, err := f.store.HandoffsForLane(context.Background(), handoffTestSlot)
	if err != nil {
		t.Fatal(err)
	}
	for _, h := range hs {
		if h.ID == id {
			return h
		}
	}
	t.Fatalf("handoff %s not stored", id)
	return factorymsg.Handoff{}
}

func (f *laneHandoffFixture) brokerDB(t *testing.T) *sql.DB {
	t.Helper()
	path, err := factorymsg.BrokerPath(f.primary, f.run)
	if err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path)+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

type handoffEndpointRow struct {
	Session, ProcessStart, UpdatedAt string
	Generation                       int64
	PID                              int
}

func (f *laneHandoffFixture) endpointRow(t *testing.T) handoffEndpointRow {
	t.Helper()
	var r handoffEndpointRow
	if err := f.brokerDB(t).QueryRow(`SELECT session_uuid,generation,pid,process_start,updated_at FROM peers WHERE slot=?`, handoffTestSlot).
		Scan(&r.Session, &r.Generation, &r.PID, &r.ProcessStart, &r.UpdatedAt); err != nil {
		t.Fatalf("read endpoint row: %v", err)
	}
	return r
}

func (f *laneHandoffFixture) count(t *testing.T, query string, args ...any) int {
	t.Helper()
	var n int
	if err := f.brokerDB(t).QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatalf("%s: %v", query, err)
	}
	return n
}

func requireHandoffNack(t *testing.T, err error, want string) {
	t.Helper()
	got, ok := factorymsg.HandoffNackReason(err)
	if !ok || got != want {
		t.Fatalf("result err=%v, want NACK %s", err, want)
	}
}

func pathExists(p string) bool {
	_, err := os.Lstat(p)
	return err == nil
}
