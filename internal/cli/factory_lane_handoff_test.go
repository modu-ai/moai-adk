package cli

import (
	"context"
	"database/sql"
	"io"
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
	handoffTestSlug   = "lane-handoff"
	handoffTestBranch = "WT-lane-handoff"
)

// laneHandoffFixture is a real git repository (main + develop) registered as
// an active factory run, with lane-1 seeded through the production launcher
// path. The process cwd is the repository because the MoAI materializer
// shells out without an explicit directory.
type laneHandoffFixture struct {
	primary, run, developPin, mainSHA string
	store                             *factorymsg.Store
	source                            factorymsg.Peer
	materializeCalls                  int
	appServer                         *spyHandoffAppServer
}

type spyHandoffAppServer struct{ calls int }

func (s *spyHandoffAppServer) ForkThread(context.Context, string, string) (codexThreadRelocation, error) {
	s.calls++
	return codexThreadRelocation{}, nil
}
func (s *spyHandoffAppServer) StartThread(context.Context, string) (codexThreadRelocation, error) {
	s.calls++
	return codexThreadRelocation{}, nil
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
func newLaneHandoffFixture(t *testing.T, baseBranch string, bindSource bool) *laneHandoffFixture {
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
	f := &laneHandoffFixture{run: "run-handoff", appServer: &spyHandoffAppServer{}}
	f.mainSHA = handoffGit(t, repo, "rev-parse", "HEAD")
	handoffGit(t, repo, "checkout", "-q", "-b", "develop")
	if err := os.WriteFile(filepath.Join(repo, "develop.txt"), []byte("develop\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	handoffGit(t, repo, "add", "develop.txt")
	handoffGit(t, repo, "commit", "-qm", "develop")
	f.developPin = handoffGit(t, repo, "rev-parse", "HEAD")
	handoffGit(t, repo, "checkout", "-q", "main")
	handoffGit(t, repo, "update-ref", "refs/remotes/origin/develop", f.developPin)
	// Mirror the real repository's ignores: config and L1 trees are not
	// source changes.
	if err := os.WriteFile(filepath.Join(repo, ".git", "info", "exclude"), []byte(".moai/\n.claude/worktrees/\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	writeWorktreeBaseConfig(t, repo, baseBranch)
	t.Chdir(repo)

	origSafe, origGlobal := sessionWorktreeGitSafeDirAdd, sessionWorktreeGitGlobalGet
	sessionWorktreeGitSafeDirAdd = func(string) error { return nil }
	sessionWorktreeGitGlobalGet = func(string) string { return "" }
	t.Cleanup(func() { sessionWorktreeGitSafeDirAdd, sessionWorktreeGitGlobalGet = origSafe, origGlobal })

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
	ctx := context.Background()
	pending, err := f.store.RegisterLaunchPending(ctx, factorymsg.Peer{
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

func (f *laneHandoffFixture) request(card, slug, mode string) laneHandoffRequest {
	return laneHandoffRequest{
		ProjectRoot: f.primary, Slot: handoffTestSlot, CardID: card, SpecID: handoffTestSpec,
		Slug: slug, Mode: mode, SourceCwd: f.primary, Activity: laneActivityIdle,
	}
}

// deps wires the REAL materializer behind a counting spy.
func (f *laneHandoffFixture) deps() laneHandoffDeps {
	return laneHandoffDeps{
		Materialize: func(name string, out io.Writer) (string, error) {
			f.materializeCalls++
			return materializeSessionWorktree(name, out)
		},
		AppServer: f.appServer,
		Out:       io.Discard,
	}
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

// requireNoSideEffects asserts a fail-closed admission touched nothing.
func (f *laneHandoffFixture) requireNoSideEffects(t *testing.T, before handoffEndpointRow, handoffsBefore, materializeBefore int, target string) {
	t.Helper()
	if after := f.endpointRow(t); after != before {
		t.Fatalf("endpoint row mutated: before=%+v after=%+v", before, after)
	}
	if n := f.count(t, `SELECT count(*) FROM lane_handoffs`); n != handoffsBefore {
		t.Fatalf("handoff rows = %d, want %d", n, handoffsBefore)
	}
	if n := f.count(t, `SELECT count(*) FROM lane_endpoint_tombstones`); n != 0 {
		t.Fatalf("tombstones = %d, want 0", n)
	}
	if n := f.count(t, `SELECT count(*) FROM messages`); n != 0 {
		t.Fatalf("message rows = %d, want 0 (no dispatch release)", n)
	}
	if f.materializeCalls != materializeBefore {
		t.Fatalf("materializer calls = %d, want %d", f.materializeCalls, materializeBefore)
	}
	if f.appServer.calls != 0 {
		t.Fatalf("app-server requests = %d, want 0", f.appServer.calls)
	}
	if target != "" && pathExists(target) {
		t.Fatalf("worktree %s exists after NACK", target)
	}
}

// TestFactoryLaneHandoffLaunchPendingSourceNack is AC-FLH-017 (REQ-FLH-016).
func TestFactoryLaneHandoffLaunchPendingSourceNack(t *testing.T) {
	f := newLaneHandoffFixture(t, "develop", false)
	ctx := context.Background()
	if _, err := f.store.ResolveLane(ctx, handoffTestSlot); err != factorymsg.ErrEndpointLaunchPending {
		t.Fatalf("fixture lane is not launch-pending: %v", err)
	}
	before := f.endpointRow(t)
	for _, mode := range []string{factorymsg.HandoffModeInteractive, factorymsg.HandoffModeHeadless} {
		_, err := prepareLaneHandoff(ctx, f.request(handoffTestCard, handoffTestSlug, mode), f.deps())
		requireHandoffNack(t, err, factorymsg.NackEndpointLaunchPending)
		f.requireNoSideEffects(t, before, 0, 0, f.target(handoffTestCard))
	}
	// Neither bound nor rolled back: the provisional row is still there,
	// still launch-pending, byte-identical including updated_at.
	if !strings.HasPrefix(f.endpointRow(t).Session, "launch-pending:") {
		t.Fatal("provisional endpoint was bound or rolled back")
	}

	// Control: after the production bind, a fresh reservation is admitted.
	f.bindSource(t)
	h, err := prepareLaneHandoff(ctx, f.request(handoffTestCard, handoffTestSlug, factorymsg.HandoffModeHeadless), f.deps())
	if err != nil {
		t.Fatalf("control reservation after bind: %v", err)
	}
	if h.State != factorymsg.HandoffWTReady {
		t.Fatalf("control state = %s, want %s", h.State, factorymsg.HandoffWTReady)
	}
}

// TestFactoryLaneHandoffAdmissionFailClosed is AC-FLH-001 (REQ-FLH-001/003).
func TestFactoryLaneHandoffAdmissionFailClosed(t *testing.T) {
	ctx := context.Background()
	nackCases := []struct {
		name, want string
		setup      func(t *testing.T, f *laneHandoffFixture, r *laneHandoffRequest)
	}{
		{"active_turn", factorymsg.NackLaneActiveTurn, func(_ *testing.T, _ *laneHandoffFixture, r *laneHandoffRequest) {
			r.Activity = laneActivityActiveTurn
		}},
		{"permission_wait", factorymsg.NackLanePermissionWait, func(_ *testing.T, _ *laneHandoffFixture, r *laneHandoffRequest) {
			r.Activity = laneActivityPermissionWait
		}},
		{"interrupt", factorymsg.NackLaneInterrupting, func(_ *testing.T, _ *laneHandoffFixture, r *laneHandoffRequest) {
			r.Activity = laneActivityInterrupting
		}},
		{"dirty_source", factorymsg.NackSourceDirty, func(t *testing.T, f *laneHandoffFixture, _ *laneHandoffRequest) {
			if err := os.WriteFile(filepath.Join(f.primary, "untracked.txt"), []byte("x"), 0o600); err != nil {
				t.Fatal(err)
			}
		}},
		{"symlink_escape_cwd", factorymsg.NackUntrustedCwd, func(t *testing.T, f *laneHandoffFixture, r *laneHandoffRequest) {
			outside := t.TempDir()
			link := filepath.Join(f.primary, "escape")
			if err := os.Symlink(outside, link); err != nil {
				t.Fatal(err)
			}
			r.SourceCwd = link
		}},
		{"occupied_branch", factorymsg.NackBranchCollision, func(t *testing.T, f *laneHandoffFixture, _ *laneHandoffRequest) {
			handoffGit(t, f.primary, "branch", handoffTestBranch, "main")
		}},
		{"occupied_path", factorymsg.NackTargetPathConflict, func(t *testing.T, f *laneHandoffFixture, _ *laneHandoffRequest) {
			if err := os.MkdirAll(f.target(handoffTestCard), 0o755); err != nil {
				t.Fatal(err)
			}
		}},
		{"stale_reservation", factorymsg.NackStaleReservation, func(_ *testing.T, f *laneHandoffFixture, r *laneHandoffRequest) {
			stale := f.source
			stale.Generation--
			r.ExpectedSource = &stale
		}},
	}
	for _, tc := range nackCases {
		t.Run(tc.name, func(t *testing.T) {
			f := newLaneHandoffFixture(t, "develop", true)
			r := f.request(handoffTestCard, handoffTestSlug, factorymsg.HandoffModeInteractive)
			tc.setup(t, f, &r)
			targetPreexists := pathExists(f.target(handoffTestCard))
			before := f.endpointRow(t)
			_, err := prepareLaneHandoff(ctx, r, f.deps())
			requireHandoffNack(t, err, tc.want)
			target := f.target(handoffTestCard)
			if targetPreexists {
				target = ""
			}
			f.requireNoSideEffects(t, before, 0, 0, target)
		})
	}

	t.Run("same_lane_in_flight", func(t *testing.T) {
		f := newLaneHandoffFixture(t, "develop", true)
		if _, err := prepareLaneHandoff(ctx, f.request(handoffTestCard, handoffTestSlug, factorymsg.HandoffModeInteractive), f.deps()); err != nil {
			t.Fatalf("first reservation: %v", err)
		}
		before := f.endpointRow(t)
		_, err := prepareLaneHandoff(ctx, f.request("t2000", "other-card", factorymsg.HandoffModeInteractive), f.deps())
		requireHandoffNack(t, err, factorymsg.NackHandoffInFlight)
		f.requireNoSideEffects(t, before, 1, 1, f.target("t2000"))
	})

	// N6/N10 re-entry (REQ-FLH-003 second sentence): a fresh reservation after
	// a NACKed handoff for the same card, with the lane already in the target.
	reentry := func(t *testing.T, mutate func(t *testing.T, f *laneHandoffFixture, target string)) (*laneHandoffFixture, factorymsg.Handoff, error) {
		t.Helper()
		f := newLaneHandoffFixture(t, "develop", true)
		first, err := prepareLaneHandoff(ctx, f.request(handoffTestCard, handoffTestSlug, factorymsg.HandoffModeInteractive), f.deps())
		if err != nil {
			t.Fatalf("first reservation: %v", err)
		}
		if _, err := f.store.NackHandoff(ctx, first, "EVIDENCE_MISMATCH"); err != nil {
			t.Fatalf("nack first handoff: %v", err)
		}
		target := f.target(handoffTestCard)
		if mutate != nil {
			mutate(t, f, target)
		}
		r := f.request(handoffTestCard, handoffTestSlug, factorymsg.HandoffModeInteractive)
		r.SourceCwd = target
		calls := f.materializeCalls
		h, err := prepareLaneHandoff(ctx, r, f.deps())
		if f.materializeCalls != calls {
			t.Fatalf("re-entry invoked the materializer: %d -> %d", calls, f.materializeCalls)
		}
		return f, h, err
	}
	t.Run("reentry_clean_reconstructs_wt_ready", func(t *testing.T) {
		f, h, err := reentry(t, nil)
		if err != nil {
			t.Fatalf("re-entry reservation: %v", err)
		}
		if h.State != factorymsg.HandoffWTReady || h.DevelopPin != f.developPin || h.TargetBranch != handoffTestBranch {
			t.Fatalf("re-entry handoff = %+v", h)
		}
	})
	t.Run("reentry_dirty_target", func(t *testing.T) {
		_, _, err := reentry(t, func(t *testing.T, _ *laneHandoffFixture, target string) {
			if err := os.WriteFile(filepath.Join(target, "dirty.txt"), []byte("x"), 0o600); err != nil {
				t.Fatal(err)
			}
		})
		requireHandoffNack(t, err, factorymsg.NackTargetDirty)
	})
	t.Run("reentry_wrong_branch", func(t *testing.T) {
		_, _, err := reentry(t, func(t *testing.T, _ *laneHandoffFixture, target string) {
			handoffGit(t, target, "checkout", "-q", "-b", "WT-elsewhere")
		})
		requireHandoffNack(t, err, factorymsg.NackBranchCollision)
	})
	t.Run("reentry_head_moved_off_pin", func(t *testing.T) {
		_, _, err := reentry(t, func(t *testing.T, _ *laneHandoffFixture, target string) {
			handoffGit(t, target, "commit", "-q", "--allow-empty", "-m", "moved")
		})
		requireHandoffNack(t, err, factorymsg.NackBaseDrift)
	})
}

// TestFactoryLaneHandoffDevelopPinAndTraceability is AC-FLH-002 (REQ-FLH-004/005).
func TestFactoryLaneHandoffDevelopPinAndTraceability(t *testing.T) {
	f := newLaneHandoffFixture(t, "develop", true)
	ctx := context.Background()
	primaryBranch := handoffGit(t, f.primary, "branch", "--show-current")
	primaryHead := handoffGit(t, f.primary, "rev-parse", "HEAD")

	h, err := prepareLaneHandoff(ctx, f.request(handoffTestCard, handoffTestSlug, factorymsg.HandoffModeInteractive), f.deps())
	if err != nil {
		t.Fatalf("prepare: %v", err)
	}
	if f.materializeCalls != 1 {
		t.Fatalf("materializer calls = %d, want 1", f.materializeCalls)
	}
	target := f.target(handoffTestCard)
	if h.State != factorymsg.HandoffWTReady || h.TargetPath != target || h.TargetBranch != handoffTestBranch || h.DevelopPin != f.developPin {
		t.Fatalf("handoff = %+v, want WT_READY at %s on %s pinned %s", h, target, handoffTestBranch, f.developPin)
	}
	if h.CardID != handoffTestCard || h.SpecID != handoffTestSpec || h.Slot != handoffTestSlot || h.RunID != f.run || h.Nonce == "" {
		t.Fatalf("handoff identity = %+v", h)
	}
	if h.Source.SessionUUID != f.source.SessionUUID || h.Source.Generation != f.source.Generation ||
		h.Source.PID != f.source.PID || h.Source.ProcessStart != f.source.ProcessStart {
		t.Fatalf("pinned source = %+v, want %+v", h.Source, f.source)
	}
	if got := handoffGit(t, target, "rev-parse", "HEAD"); got != f.developPin {
		t.Fatalf("target HEAD = %s, want pin %s", got, f.developPin)
	}
	if got := handoffGit(t, target, "branch", "--show-current"); got != handoffTestBranch {
		t.Fatalf("target branch = %s, want %s", got, handoffTestBranch)
	}
	if n := strings.Count(handoffGit(t, f.primary, "worktree", "list", "--porcelain")+"\n", "branch refs/heads/"+handoffTestBranch+"\n"); n != 1 {
		t.Fatalf("branch %s checked out %d times, want 1", handoffTestBranch, n)
	}
	if got := handoffGit(t, f.primary, "branch", "--list", handoffTestCard); got != "" {
		t.Fatalf("materializer branch %q survived the rename", got)
	}
	trace := laneHandoffTraceOf(h)
	want := laneHandoffTrace{
		Card: handoffTestCard, Spec: handoffTestSpec, Worktree: target, Branch: handoffTestBranch,
		Evidence: ".moai/reports/t1082/verdict.md",
	}
	if trace != want {
		t.Fatalf("trace = %+v, want %+v", trace, want)
	}
	if got := handoffGit(t, f.primary, "branch", "--show-current"); got != primaryBranch {
		t.Fatalf("primary branch moved: %s -> %s", primaryBranch, got)
	}
	if got := handoffGit(t, f.primary, "rev-parse", "HEAD"); got != primaryHead {
		t.Fatalf("primary HEAD moved: %s -> %s", primaryHead, got)
	}
	if got := handoffGit(t, f.primary, "rev-parse", "refs/heads/develop"); got != f.developPin {
		t.Fatalf("develop moved: %s -> %s", f.developPin, got)
	}
}

// TestFactoryLaneHandoffCreationBaseDriftRejected is AC-FLH-016 (REQ-FLH-004).
func TestFactoryLaneHandoffCreationBaseDriftRejected(t *testing.T) {
	ctx := context.Background()
	requireDriftPreserved := func(t *testing.T, f *laneHandoffFixture, err error) {
		t.Helper()
		requireHandoffNack(t, err, factorymsg.NackBaseDrift)
		hs, lerr := f.store.HandoffsForLane(ctx, handoffTestSlot)
		if lerr != nil || len(hs) != 1 || hs[0].State != factorymsg.HandoffNack || hs[0].Reason != factorymsg.NackBaseDrift {
			t.Fatalf("handoffs = %+v err=%v, want one NACK/%s", hs, lerr, factorymsg.NackBaseDrift)
		}
		if !pathExists(f.target(handoffTestCard)) {
			t.Fatal("drifted target was not preserved for inspection")
		}
		if n := f.count(t, `SELECT count(*) FROM lane_handoff_events WHERE to_state IN ('SWITCH_PENDING_INTERACTIVE','SWITCH_PENDING_HEADLESS','BOUND')`); n != 0 {
			t.Fatalf("switch/bound events = %d, want 0", n)
		}
		if n := f.count(t, `SELECT count(*) FROM messages`); n != 0 {
			t.Fatalf("dispatch rows = %d, want 0", n)
		}
	}

	t.Run("main_creation_mutant", func(t *testing.T) {
		// The t1082 incident: the materializer cut the tree from main while the
		// reservation pinned local develop.
		f := newLaneHandoffFixture(t, "", true)
		_, err := prepareLaneHandoff(ctx, f.request(handoffTestCard, handoffTestSlug, factorymsg.HandoffModeInteractive), f.deps())
		requireDriftPreserved(t, f, err)
		if got := handoffGit(t, f.target(handoffTestCard), "rev-parse", "HEAD"); got != f.mainSHA {
			t.Fatalf("mutant target HEAD = %s, want main %s", got, f.mainSHA)
		}
	})
	t.Run("develop_moves_during_create", func(t *testing.T) {
		f := newLaneHandoffFixture(t, "develop", true)
		deps := f.deps()
		real := deps.Materialize
		deps.Materialize = func(name string, out io.Writer) (string, error) {
			tree := handoffGit(t, f.primary, "rev-parse", "develop^{tree}")
			moved := handoffGit(t, f.primary, "commit-tree", tree, "-p", "develop", "-m", "moved during create")
			handoffGit(t, f.primary, "update-ref", "refs/heads/develop", moved)
			return real(name, out)
		}
		_, err := prepareLaneHandoff(ctx, f.request(handoffTestCard, handoffTestSlug, factorymsg.HandoffModeInteractive), deps)
		requireDriftPreserved(t, f, err)
	})
	t.Run("matching_base_control", func(t *testing.T) {
		f := newLaneHandoffFixture(t, "develop", true)
		h, err := prepareLaneHandoff(ctx, f.request(handoffTestCard, handoffTestSlug, factorymsg.HandoffModeInteractive), f.deps())
		if err != nil || h.State != factorymsg.HandoffWTReady {
			t.Fatalf("control = %+v err=%v, want WT_READY", h, err)
		}
	})
}
