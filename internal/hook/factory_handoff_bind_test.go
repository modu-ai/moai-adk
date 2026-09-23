package hook

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// interactiveHandoffFixture is a real git repository with a develop pin and a
// card worktree on its WT-* branch, registered as an active factory run whose
// lane-1 endpoint was bound through the production launcher path and whose
// interactive handoff was driven to SWITCH_PENDING_INTERACTIVE through the
// production store transitions.
type interactiveHandoffFixture struct {
	primary, target, run, pin string
	store                     *factorymsg.Store
	source                    factorymsg.Peer
	h                         factorymsg.Handoff
}

func handoffTestGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func newInteractiveHandoffFixture(t *testing.T, switchPending bool) *interactiveHandoffFixture {
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
	handoffTestGit(t, repo, "add", "seed.txt")
	handoffTestGit(t, repo, "commit", "-qm", "seed")
	handoffTestGit(t, repo, "branch", "develop")
	if err := os.WriteFile(filepath.Join(repo, ".git", "info", "exclude"), []byte(".moai/\n.claude/worktrees/\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	f := &interactiveHandoffFixture{run: "run-interactive-handoff"}
	f.primary = homestate.CanonicalProjectRoot(repo)
	f.pin = handoffTestGit(t, f.primary, "rev-parse", "develop")
	f.target = filepath.Join(f.primary, ".claude", "worktrees", "t1082")
	// Test fixture only: the production path creates this tree through the
	// MoAI materializer (AC-FLH-002), which lives in package cli.
	handoffTestGit(t, f.primary, "worktree", "add", "-q", "-b", "WT-lane-handoff", f.target, "develop")

	recordActiveFactoryRun(t, f.primary, f.run)
	t.Setenv(config.EnvMoaiKanbanID, f.run)
	t.Setenv(config.EnvMoaiFactoryWorkers, "1")
	t.Setenv(config.EnvMoaiFactoryWorker, "lane-1")
	t.Setenv(config.EnvMoaiKanbanBackend, "codex")
	owner, start := factoryHookOwnerIdentity(t)

	var err error
	f.store, err = factorymsg.Open(f.primary, f.run)
	if err != nil {
		t.Fatal(err)
	}
	closeOnCleanup(t, "factory message broker", f.store)
	ctx := context.Background()
	pending, err := f.store.RegisterLaunchPending(ctx, factorymsg.Peer{
		ProjectKey: homestate.ProjectKey(f.primary), RunID: f.run, Backend: "codex",
		Role: "worker", Slot: "lane-1", PID: owner, ProcessStart: start,
	})
	if err != nil {
		t.Fatal(err)
	}
	pending.SessionUUID = "src-uuid"
	var ok bool
	if f.source, ok, err = f.store.BindLaunchPending(ctx, pending); err != nil || !ok {
		t.Fatalf("bind source ok=%v err=%v", ok, err)
	}
	h, err := f.store.ReserveHandoff(ctx, factorymsg.HandoffReservation{
		Slot: "lane-1", CardID: "t1082", SpecID: "SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001",
		Mode: factorymsg.HandoffModeInteractive, DevelopPin: f.pin, TargetPath: f.target, TargetBranch: "WT-lane-handoff",
	})
	if err != nil {
		t.Fatal(err)
	}
	if h, err = f.store.MarkHandoffWTReady(ctx, h); err != nil {
		t.Fatal(err)
	}
	if switchPending {
		if h, err = f.store.MarkHandoffSwitchPendingInteractive(ctx, h); err != nil {
			t.Fatal(err)
		}
	}
	f.h = h
	return f
}

func (f *interactiveHandoffFixture) db(t *testing.T) *sql.DB {
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

func (f *interactiveHandoffFixture) count(t *testing.T, q string, args ...any) int {
	t.Helper()
	var n int
	if err := f.db(t).QueryRow(q, args...).Scan(&n); err != nil {
		t.Fatalf("%s: %v", q, err)
	}
	return n
}

func (f *interactiveHandoffFixture) state(t *testing.T) (string, string) {
	t.Helper()
	var st, reason string
	if err := f.db(t).QueryRow(`SELECT state,reason FROM lane_handoffs WHERE id=?`, f.h.ID).Scan(&st, &reason); err != nil {
		t.Fatal(err)
	}
	return st, reason
}

type hookEndpointRow struct {
	Session, ProcessStart, UpdatedAt string
	Generation                       int64
	PID                              int
}

func (f *interactiveHandoffFixture) row(t *testing.T) hookEndpointRow {
	t.Helper()
	var r hookEndpointRow
	if err := f.db(t).QueryRow(`SELECT session_uuid,generation,pid,process_start,updated_at FROM peers WHERE slot='lane-1'`).
		Scan(&r.Session, &r.Generation, &r.PID, &r.ProcessStart, &r.UpdatedAt); err != nil {
		t.Fatal(err)
	}
	return r
}

// requireNothingBound asserts the rebind wrote none of its five facts.
func (f *interactiveHandoffFixture) requireNothingBound(t *testing.T, before hookEndpointRow) {
	t.Helper()
	if after := f.row(t); after != before {
		t.Fatalf("endpoint moved without BOUND: before=%+v after=%+v", before, after)
	}
	for _, q := range []string{
		`SELECT count(*) FROM lane_endpoint_tombstones`,
		`SELECT count(*) FROM lane_handoff_receipts`,
		`SELECT count(*) FROM lane_dispatch_releases`,
		`SELECT count(*) FROM lane_handoff_events WHERE to_state='BOUND'`,
	} {
		if n := f.count(t, q); n != 0 {
			t.Fatalf("%s = %d, want 0", q, n)
		}
	}
}

// sessionStart delivers the SessionStart of a normal turn through the
// production hook path.
func sessionStart(session, cwd string) string {
	return registerFactorySessionStartPeer(context.Background(), &HookInput{SessionID: session, CWD: cwd})
}

// TestFactoryLaneHandoffInteractiveStateMachine is AC-FLH-003
// (REQ-FLH-002/006): an interactive handoff stays SWITCH_PENDING_INTERACTIVE
// until the SessionStart of the user's next normal turn after /cd, and only
// verified target cwd/branch/HEAD plus a new session UUID advance it to BOUND.
func TestFactoryLaneHandoffInteractiveStateMachine(t *testing.T) {
	t.Run("pending_until_next_turn_then_bound", func(t *testing.T) {
		f := newInteractiveHandoffFixture(t, true)
		before := f.row(t)

		// A SessionStart of the current session (no /cd happened) is not
		// evidence: the handoff stays pending and nothing is written.
		if notice := sessionStart("src-uuid", f.primary); notice != "" {
			t.Fatalf("current-session SessionStart notice = %q", notice)
		}
		if st, _ := f.state(t); st != factorymsg.HandoffSwitchPendingInteractive {
			t.Fatalf("state = %s, want SWITCH_PENDING_INTERACTIVE", st)
		}
		f.requireNothingBound(t, before)

		// The user ran /cd and sent the next normal message.
		notice := sessionStart("post-cd-uuid", f.target)
		if !strings.Contains(notice, "factory handoff bound") || !strings.Contains(notice, "t1082") {
			t.Fatalf("next-turn SessionStart notice = %q", notice)
		}
		if st, _ := f.state(t); st != factorymsg.HandoffBound {
			t.Fatalf("state = %s, want BOUND", st)
		}
		after := f.row(t)
		if after.Session != "post-cd-uuid" || after.Generation != f.source.Generation+1 || after.PID != f.source.PID || after.ProcessStart != f.source.ProcessStart {
			t.Fatalf("endpoint = %+v, want post-cd-uuid at generation %d on the same owner", after, f.source.Generation+1)
		}
		if n := f.count(t, `SELECT count(*) FROM lane_endpoint_tombstones WHERE session_uuid='src-uuid' AND generation=? AND replaced_by_session='post-cd-uuid'`, f.source.Generation); n != 1 {
			t.Fatalf("tombstones = %d, want 1", n)
		}
		if n := f.count(t, `SELECT count(*) FROM lane_handoff_receipts WHERE handoff_id=? AND session_uuid='post-cd-uuid'`, f.h.ID); n != 1 {
			t.Fatalf("BOUND receipts = %d, want 1", n)
		}
		// The binding took no model turn and sent no message.
		if n := f.count(t, `SELECT count(*) FROM messages`); n != 0 {
			t.Fatalf("messages = %d, want 0", n)
		}
		var path []string
		rows, err := f.db(t).Query(`SELECT to_state FROM lane_handoff_events WHERE handoff_id=? ORDER BY id`, f.h.ID)
		if err != nil {
			t.Fatal(err)
		}
		for rows.Next() {
			var s string
			if err := rows.Scan(&s); err != nil {
				t.Fatal(err)
			}
			path = append(path, s)
		}
		_ = rows.Close()
		if got := strings.Join(path, ">"); got != "RESERVED>WT_READY>SWITCH_PENDING_INTERACTIVE>BOUND" {
			t.Fatalf("state path = %s", got)
		}

		// Later turns of the bound session change nothing.
		bound := f.row(t)
		if notice := sessionStart("post-cd-uuid", f.target); notice != "" {
			t.Fatalf("repeat SessionStart notice = %q", notice)
		}
		in := &HookInput{SessionID: "post-cd-uuid", CWD: f.target, Prompt: "continue the card"}
		if notice := registerFactoryUserPromptPeer(context.Background(), in); notice != "" {
			t.Fatalf("bound UserPromptSubmit notice = %q", notice)
		}
		if again := f.row(t); again != bound {
			t.Fatalf("bound endpoint rewritten: before=%+v after=%+v", bound, again)
		}

		// The tombstoned source session resumes at the primary checkout: its
		// SessionStart gets the shared endpoint-replaced notice, byte-equal to
		// its UserPromptSubmit notice, and nothing moves.
		durable := f.durable(t)
		start := sessionStart("src-uuid", f.primary)
		prompt := registerFactoryUserPromptPeer(context.Background(), &HookInput{SessionID: "src-uuid", CWD: f.primary, Prompt: "resume"})
		want := fmt.Sprintf("factory endpoint replaced STALE_ENDPOINT: slot=lane-1 is current at post-cd-uuid generation %d; this session receives no factory messages", bound.Generation)
		if start != want || prompt != want {
			t.Fatalf("tombstoned session notices:\nSessionStart     = %q\nUserPromptSubmit = %q\nwant             = %q", start, prompt, want)
		}
		if after := f.row(t); after != bound {
			t.Fatalf("tombstoned SessionStart moved the endpoint: before=%+v after=%+v", bound, after)
		}
		if after := f.durable(t); after != durable {
			t.Fatal("tombstoned SessionStart changed tombstone or receipt rows")
		}
	})

	t.Run("launch_pending_notice_matches_user_prompt", func(t *testing.T) {
		f := newInteractiveHandoffFixture(t, true)
		ctx := context.Background()
		// The handoff-bound owner is a helper process so it can be made not
		// current; the rebind is the production store transaction.
		helper := startStoppableLiveOwner(t)
		b, err := f.store.BindHandoff(ctx, f.h, factorymsg.HandoffBindEvidence{
			Mode: factorymsg.HandoffModeInteractive, Nonce: f.h.Nonce, CardID: f.h.CardID, SpecID: f.h.SpecID,
			SessionUUID: "post-cd-uuid", PID: helper.pid, ProcessStart: helper.start,
			Cwd: f.target, WorktreeRoot: f.target, Branch: "WT-lane-handoff", Head: f.pin,
		})
		if err != nil {
			t.Fatal(err)
		}
		helper.stopAndWait(t)
		owner, start := factoryHookOwnerIdentity(t)
		pending, err := f.store.RegisterLaunchPending(ctx, factorymsg.Peer{
			ProjectKey: homestate.ProjectKey(f.primary), RunID: f.run, Backend: "codex",
			Role: "worker", Slot: "lane-1", PID: owner, ProcessStart: start,
		})
		if err != nil {
			t.Fatalf("launcher registration after the bound owner stopped: %v", err)
		}
		if pending.Generation <= b.New.Generation {
			t.Fatalf("launch-pending generation %d not above %d", pending.Generation, b.New.Generation)
		}
		row, durable := f.row(t), f.durable(t)
		prompt := registerFactoryUserPromptPeer(ctx, &HookInput{SessionID: "src-uuid", CWD: f.primary, Prompt: "resume"})
		sess := sessionStart("src-uuid", f.primary)
		want := fmt.Sprintf("factory endpoint replaced STALE_ENDPOINT: slot=lane-1 is current at  generation %d; this session receives no factory messages", pending.Generation)
		if sess != prompt || prompt != want {
			t.Fatalf("launch-pending notices:\nSessionStart     = %q\nUserPromptSubmit = %q\nwant             = %q", sess, prompt, want)
		}
		if after := f.row(t); after != row {
			t.Fatalf("launch-pending row changed: before=%+v after=%+v", row, after)
		}
		if after := f.durable(t); after != durable {
			t.Fatal("tombstone or receipt rows changed")
		}
	})

	t.Run("no_evidence_before_the_cd_guidance", func(t *testing.T) {
		f := newInteractiveHandoffFixture(t, false)
		before := f.row(t)
		if notice := sessionStart("post-cd-uuid", f.target); strings.Contains(notice, "factory handoff") {
			t.Fatalf("WT_READY handoff consumed SessionStart evidence: %q", notice)
		}
		if st, _ := f.state(t); st != factorymsg.HandoffWTReady {
			t.Fatalf("state = %s, want WT_READY", st)
		}
		f.requireNothingBound(t, before)
	})

	mismatches := map[string]func(t *testing.T, f *interactiveHandoffFixture) string{
		"cwd_not_target": func(t *testing.T, f *interactiveHandoffFixture) string { return f.primary },
		"cwd_subdirectory": func(t *testing.T, f *interactiveHandoffFixture) string {
			sub := filepath.Join(f.target, "sub")
			if err := os.MkdirAll(sub, 0o700); err != nil {
				t.Fatal(err)
			}
			return sub
		},
		"branch_moved": func(t *testing.T, f *interactiveHandoffFixture) string {
			handoffTestGit(t, f.target, "switch", "-q", "-c", "WT-other")
			return f.target
		},
		"head_moved": func(t *testing.T, f *interactiveHandoffFixture) string {
			handoffTestGit(t, f.target, "commit", "-q", "--allow-empty", "-m", "off pin")
			return f.target
		},
	}
	for name, cwdFor := range mismatches {
		t.Run("mismatch/"+name, func(t *testing.T) {
			f := newInteractiveHandoffFixture(t, true)
			before := f.row(t)
			cwd := cwdFor(t, f)
			notice := sessionStart("post-cd-uuid", cwd)
			if !strings.Contains(notice, "factory handoff NACK "+factorymsg.NackTargetReadbackMismatch) {
				t.Fatalf("mismatched SessionStart notice = %q", notice)
			}
			if st, reason := f.state(t); st != factorymsg.HandoffNack || reason != factorymsg.NackTargetReadbackMismatch {
				t.Fatalf("handoff = %s/%s, want NACK/%s", st, reason, factorymsg.NackTargetReadbackMismatch)
			}
			f.requireNothingBound(t, before)
		})
	}
}

// durable renders the tombstone and BOUND receipt rows as one comparable string.
func (f *interactiveHandoffFixture) durable(t *testing.T) string {
	t.Helper()
	var out []string
	for _, q := range []string{
		`SELECT slot||'|'||session_uuid||'|'||generation||'|'||replaced_by_session||'|'||replaced_by_generation||'|'||handoff_id||'|'||bound_at FROM lane_endpoint_tombstones ORDER BY session_uuid,generation`,
		`SELECT id||'|'||handoff_id||'|'||session_uuid||'|'||generation||'|'||pid||'|'||process_start||'|'||created_at FROM lane_handoff_receipts ORDER BY id`,
	} {
		rows, err := f.db(t).Query(q)
		if err != nil {
			t.Fatal(err)
		}
		for rows.Next() {
			var s string
			if err := rows.Scan(&s); err != nil {
				t.Fatal(err)
			}
			out = append(out, s)
		}
		_ = rows.Close()
		out = append(out, "#")
	}
	return strings.Join(out, ";")
}

// stoppableLiveOwner is the owner helper child with a stop that kills AND
// reaps it: a killed but unreaped child is a zombie the probe still reads as
// live, so the endpoint it owns would stay current.
type stoppableLiveOwner struct {
	pid   int
	start string
	stop  func()
}

func startStoppableLiveOwner(t *testing.T) stoppableLiveOwner {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^TestFactoryHandoffOwnerHelperProcess$", "-test.timeout=300s")
	cmd.Env = append(os.Environ(), ownerHelperEnv+"=1")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start owner helper: %v", err)
	}
	var once sync.Once
	stop := func() {
		once.Do(func() {
			_ = stdin.Close()
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		})
	}
	t.Cleanup(stop)
	deadline := time.Now().Add(5 * time.Second)
	for {
		start, state := homestate.ProbeProcessIdentity(cmd.Process.Pid)
		if state == homestate.ProcessIdentityLive && start != "" {
			return stoppableLiveOwner{pid: cmd.Process.Pid, start: start, stop: stop}
		}
		if time.Now().After(deadline) {
			t.Fatalf("owner helper %d never became probeable (state=%v)", cmd.Process.Pid, state)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// stopAndWait makes the owner not current under the t1074 live-owner rule.
func (o stoppableLiveOwner) stopAndWait(t *testing.T) {
	t.Helper()
	o.stop()
	deadline := time.Now().Add(5 * time.Second)
	for {
		start, state := homestate.ProbeProcessIdentity(o.pid)
		if state != homestate.ProcessIdentityLive || start != o.start {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("owner %d still current after kill and wait", o.pid)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
