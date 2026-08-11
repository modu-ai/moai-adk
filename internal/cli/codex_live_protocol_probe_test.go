package cli

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// Live protocol probe for SPEC-CODEX-PHASE2-001. It closes the run phase's
// standing gap: M0 read the generated JSON Schema and every wire-level claim
// after it was an inference from that schema, which cannot falsify a misreading
// of itself. This file executes the REAL codex app-server and records the
// verbatim NDJSON both directions.
//
// OPT-IN ONLY. Every turn spends real codex quota, so the probe is skipped
// unless MOAI_CODEX_LIVE_PROBE=1 is set explicitly. It never runs on a plain
// `go test ./...`.
//
// The probe drives the production session handle (openCodexSessionOn / runTurn /
// sendTurnInterrupt) rather than hand-rolling requests, so what it verifies is
// what ships. The only inserted seam is probeTap, a pass-through codexConn that
// records each line.

// probeLiveEnv is the explicit opt-in. Absent → skip.
const probeLiveEnv = "MOAI_CODEX_LIVE_PROBE"

// probeBinEnv overrides binary resolution. It exists because exec.LookPath —
// what production uses (codexLookPath) — does not necessarily find a WORKING
// codex: on the probe host it resolves to a bun shim whose vendored rust binary
// is ENOENT, while the functional 0.146.1 install sits under the npm prefix and
// is reachable interactively only through a shell function. Pointing the probe
// at the working binary is what lets the protocol be observed at all; it does
// not paper over the resolution finding, which is recorded in progress.md.
const probeBinEnv = "MOAI_CODEX_LIVE_BIN"

// probeTap wraps a codexConn and records every line in both directions. It
// forwards unchanged, so the session under test behaves exactly as in
// production; the tap only observes.
type probeTap struct {
	inner codexConn
	mu    sync.Mutex
	log   []string
}

func (t *probeTap) record(dir, line string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.log = append(t.log, dir+" "+line)
}

func (t *probeTap) send(line string) error {
	t.record("-->", line)
	return t.inner.send(line)
}

func (t *probeTap) recv() (string, bool) {
	line, ok := t.inner.recv()
	if ok {
		t.record("<--", line)
	}
	return line, ok
}

func (t *probeTap) close() error { return t.inner.close() }

// pid satisfies codexProcessConn so the tapped session still reports the pid of
// the process the runner spawned (REQ-CX2-012 shape is unchanged by the tap).
func (t *probeTap) pid() int { return codexConnPID(t.inner) }

func (t *probeTap) dump() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]string(nil), t.log...)
}

// probeRunner spawns the real subprocess and hands back a tapped connection,
// publishing the tap so the test can read the transcript afterwards.
type probeRunner struct{ out **probeTap }

func (r probeRunner) start(ctx context.Context, binaryPath string, args []string) (codexConn, error) {
	conn, err := (realCodexSessionRunner{}).start(ctx, binaryPath, args)
	if err != nil {
		return nil, err
	}
	tap := &probeTap{inner: conn}
	*r.out = tap
	return tap, nil
}

// probeLiveCodex resolves a working codex binary or skips.
func probeLiveCodex(t *testing.T) string {
	t.Helper()
	if os.Getenv(probeLiveEnv) != "1" {
		t.Skipf("%s != 1 — live protocol probe is opt-in (it spends real codex quota)", probeLiveEnv)
	}
	bin := os.Getenv(probeBinEnv)
	if bin == "" {
		var err error
		if bin, err = exec.LookPath(codexBinaryName); err != nil {
			t.Skipf("codex binary not on PATH: %v", err)
		}
	}
	ver, vErr := exec.Command(bin, "--version").Output()
	if vErr != nil || !strings.Contains(string(ver), "codex") {
		t.Skipf("codex --version non-functional (ver=%q err=%v)", strings.TrimSpace(string(ver)), vErr)
	}
	t.Logf("codex binary=%s version=%s", bin, strings.TrimSpace(string(ver)))
	return bin
}

// probeInstallRunner points the session seam at the tapped real runner.
func probeInstallRunner(t *testing.T) **probeTap {
	t.Helper()
	var tap *probeTap
	prev := codexSession
	codexSession = probeRunner{out: &tap}
	t.Cleanup(func() { codexSession = prev })
	return &tap
}

// probeWriteTranscript persists the verbatim NDJSON transcript so the evidence
// path still resolves at audit time (it is cited from progress.md §E.2).
func probeWriteTranscript(t *testing.T, name string, tap *probeTap) {
	t.Helper()
	if tap == nil {
		t.Logf("no transcript: session never started")
		return
	}
	lines := tap.dump()
	dir := filepath.Join("..", "..", ".moai", "state", "verify", "live-probe")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Logf("transcript dir: %v", err)
		return
	}
	path := filepath.Join(dir, name+".ndjson")
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Logf("transcript write: %v", err)
		return
	}
	t.Logf("transcript (%d lines) → %s", len(lines), path)
	for _, l := range lines {
		t.Logf("%s", probeTrim(l))
	}
}

// probeTrim bounds a logged line so a large aggregatedOutput payload does not
// swamp the transcript. The full line is on disk either way.
func probeTrim(s string) string {
	const max = 600
	if len(s) <= max {
		return s
	}
	return s[:max] + "…<truncated, full line in transcript file>"
}

// probeShortPrompt is deliberately the cheapest possible turn: the probe tests
// the protocol envelope, not the model's output.
const probeShortPrompt = "say ok"

// TestCodexLive_ThreadReuseAndTurnInterrupt covers probe items 1, 2 and the
// readOnly half of item 3 in ONE session, so a single handshake pays for three
// answers:
//
//	item 1 — a second turn/start on an existing threadId after the first turn
//	         completed (REQ-CX2-001, REQ-CX2-008; M0 (b) was schema-only).
//	item 2 — turn/started actually arrives and its turn.id is accepted as
//	         turn/interrupt's turnId (REQ-CX2-003, REQ-CX2-011).
//	item 3a — the internally-tagged {"type":"readOnly"} sandboxPolicy envelope
//	         M3 sends on every turn is accepted (REQ-CX2-007).
func TestCodexLive_ThreadReuseAndTurnInterrupt(t *testing.T) {
	bin := probeLiveCodex(t)
	tap := probeInstallRunner(t)
	cwd := t.TempDir()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	params := map[string]any{
		"cwd":           cwd,
		"prompt":        probeShortPrompt,
		"sandboxPolicy": codexSandboxPolicy(false),
	}

	sess, err := openCodexSession(ctx, bin, params)
	if err != nil {
		probeWriteTranscript(t, "reuse-interrupt", *tap)
		t.Fatalf("handshake failed: %v", err)
	}
	defer func() {
		_ = sess.close()
		probeWriteTranscript(t, "reuse-interrupt", *tap)
	}()
	t.Logf("ITEM-0 handshake OK: threadId=%q pid=%d", sess.threadID, sess.pid())

	// --- turn 1 -------------------------------------------------------------
	out1, err1 := sess.runTurn(ctx, codexMethodTurnStart, params)
	t.Logf("ITEM-3a turn 1 (sandboxPolicy readOnly): err=%v summary=%q", err1, probeTrim(out1.Summary))
	if err1 != nil {
		t.Fatalf("turn 1 failed — readOnly sandboxPolicy envelope or turn/start shape rejected: %v", err1)
	}
	turn1ID := sess.turnID
	t.Logf("ITEM-2a turn/started observed on turn 1: turnId=%q", turn1ID)
	if turn1ID == "" {
		t.Errorf("no turn/started observed on turn 1 — turnId is unavailable, REQ-CX2-011 degrades")
	}

	// --- turn 2: the reuse question -----------------------------------------
	out2, err2 := sess.runTurn(ctx, codexMethodTurnStart, params)
	t.Logf("ITEM-1 turn 2 on the SAME threadId=%q: err=%v summary=%q", sess.threadID, err2, probeTrim(out2.Summary))
	if err2 != nil {
		t.Fatalf("turn 2 on an existing thread REJECTED — REQ-CX2-001/008 thread reuse does not hold: %v", err2)
	}
	turn2ID := sess.turnID
	if turn2ID == turn1ID {
		t.Errorf("turn 2 reported the same turnId %q as turn 1 — turn ids are not per-turn", turn2ID)
	}
	t.Logf("ITEM-1 CONFIRMED: two turns on one thread; turnIds %q then %q", turn1ID, turn2ID)

	// --- turn 3: interrupt an in-flight turn ---------------------------------
	// Production shape (M4): the turn runs on its own goroutine while the cancel
	// path sends turn/interrupt on the same session from another goroutine, the
	// moment the mid-flight onTurnStarted hook publishes the turnId.
	got := make(chan string, 1)
	sess.onTurnStarted = func(id string) {
		select {
		case got <- id:
		default:
		}
	}
	longParams := map[string]any{
		"cwd":           cwd,
		"prompt":        "count slowly from 1 to 200, one number per line",
		"sandboxPolicy": codexSandboxPolicy(false),
	}
	done := make(chan error, 1)
	go func() {
		_, e := sess.runTurn(ctx, codexMethodTurnStart, longParams)
		done <- e
	}()

	var turn3ID string
	select {
	case turn3ID = <-got:
	case <-time.After(90 * time.Second):
		t.Errorf("no turn/started within 90s on turn 3 — cannot exercise turn/interrupt")
	}

	if turn3ID != "" {
		t.Logf("ITEM-2b interrupting turn %q on thread %q", turn3ID, sess.threadID)
		if err := sess.sendTurnInterrupt(sess.threadID, turn3ID); err != nil {
			t.Errorf("sendTurnInterrupt write failed: %v", err)
		}
	}

	select {
	case e := <-done:
		t.Logf("ITEM-2b turn 3 ended after interrupt: err=%v", e)
	case <-time.After(120 * time.Second):
		t.Errorf("turn 3 still running 120s after turn/interrupt — the interrupt was not honoured in-flight")
	}

	// The interrupt response (if any) is in the transcript; find it by id so the
	// verdict does not rest on the turn simply ending.
	probeReportInterruptOutcome(t, *tap)
}

// probeReportInterruptOutcome scans the transcript for the turn/interrupt
// request and any response carrying its id, so an accepted-vs-rejected verdict
// rests on the server's own line rather than on timing.
func probeReportInterruptOutcome(t *testing.T, tap *probeTap) {
	t.Helper()
	if tap == nil {
		return
	}
	var wantID json.RawMessage
	for _, l := range tap.dump() {
		if !strings.HasPrefix(l, "--> ") || !strings.Contains(l, codexMethodTurnInterrupt) {
			continue
		}
		var m struct {
			ID     json.RawMessage `json:"id"`
			Method string          `json:"method"`
		}
		if err := json.Unmarshal([]byte(strings.TrimPrefix(l, "--> ")), &m); err == nil && m.Method == codexMethodTurnInterrupt {
			wantID = m.ID
			t.Logf("ITEM-2b interrupt request sent verbatim: %s", strings.TrimPrefix(l, "--> "))
		}
	}
	if len(wantID) == 0 {
		t.Logf("ITEM-2b no turn/interrupt request found in transcript")
		return
	}
	for _, l := range tap.dump() {
		if !strings.HasPrefix(l, "<-- ") {
			continue
		}
		var m rpcMessage
		if err := json.Unmarshal([]byte(strings.TrimPrefix(l, "<-- ")), &m); err != nil {
			continue
		}
		if len(m.ID) == 0 || string(m.ID) != string(wantID) {
			continue
		}
		if m.Error != nil {
			t.Errorf("ITEM-2b turn/interrupt REJECTED: code=%d message=%q", m.Error.Code, m.Error.Message)
			return
		}
		t.Logf("ITEM-2b turn/interrupt ACCEPTED: %s", probeTrim(strings.TrimPrefix(l, "<-- ")))
		return
	}
	t.Logf("ITEM-2b no response line carrying the interrupt id %s was observed", string(wantID))
}

// TestCodexLive_SandboxPolicyStickiness is probe item 3b: does a policy set on
// one turn actually persist to the next turn on the same thread? The schema
// documents sandboxPolicy as overriding "for this turn and subsequent turns",
// and that sentence is the ENTIRE basis for REQ-CX2-007 requiring the field on
// every turn. If it is true, omitting the field on a later turn inherits the
// earlier write-enabled policy.
//
// Both turns run in a throwaway temp dir, so a granted write lands nowhere that
// matters.
func TestCodexLive_SandboxPolicyStickiness(t *testing.T) {
	bin := probeLiveCodex(t)
	tap := probeInstallRunner(t)
	cwd := t.TempDir()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	sess, err := openCodexSession(ctx, bin, map[string]any{"cwd": cwd})
	if err != nil {
		probeWriteTranscript(t, "sticky", *tap)
		t.Fatalf("handshake failed: %v", err)
	}
	defer func() {
		_ = sess.close()
		probeWriteTranscript(t, "sticky", *tap)
	}()

	// Turn 1 opts INTO writes explicitly.
	_, err1 := sess.runTurn(ctx, codexMethodTurnStart, map[string]any{
		"cwd":           cwd,
		"prompt":        "create a file named first.txt containing the word hi",
		"sandboxPolicy": codexSandboxPolicy(true),
	})
	first := probeFileExists(filepath.Join(cwd, "first.txt"))
	t.Logf("STICKY turn 1 (sandboxPolicy workspaceWrite): err=%v first.txt=%v", err1, first)

	// Turn 2 omits sandboxPolicy entirely — the shape REQ-CX2-007 forbids.
	_, err2 := sess.runTurn(ctx, codexMethodTurnStart, map[string]any{
		"cwd":    cwd,
		"prompt": "create a file named second.txt containing the word hi",
	})
	second := probeFileExists(filepath.Join(cwd, "second.txt"))
	t.Logf("STICKY turn 2 (sandboxPolicy OMITTED): err=%v second.txt=%v", err2, second)

	switch {
	case !first:
		t.Logf("STICKY INCONCLUSIVE: turn 1 did not write even WITH workspaceWrite, so turn 2 proves nothing about inheritance")
	case second:
		t.Logf("STICKY CONFIRMED: a policy set on turn 1 persisted to a turn that sent none — REQ-CX2-007's per-turn transmission is load-bearing")
	default:
		t.Logf("STICKY NOT REPRODUCED: turn 2 did not write. Either the policy is not sticky here, or the turn declined for an unrelated reason — read the transcript before concluding")
	}
}

func probeFileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// TestCodexLive_OmittedSandboxPolicyBaseline disambiguates the stickiness
// result. A turn that writes after OMITTING sandboxPolicy has two possible
// explanations, and they are not the same claim:
//
//	(a) the policy set on an EARLIER turn was inherited (sticky), or
//	(b) omitting the field simply defaults to a write-capable policy.
//
// The stickiness probe alone cannot tell them apart, because its turn 2 follows
// a workspaceWrite turn. This test removes that predecessor: it is the FIRST
// turn on a FRESH thread and sends no sandboxPolicy at all.
//
// Both outcomes leave REQ-CX2-007's per-turn transmission load-bearing — under
// (a) an omitted field inherits a write-enabled policy, under (b) it never
// meant readOnly in the first place — but only one of them is the reason the
// SPEC currently states, so the two are reported apart rather than merged.
func TestCodexLive_OmittedSandboxPolicyBaseline(t *testing.T) {
	bin := probeLiveCodex(t)
	tap := probeInstallRunner(t)
	cwd := t.TempDir()

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()

	sess, err := openCodexSession(ctx, bin, map[string]any{"cwd": cwd})
	if err != nil {
		probeWriteTranscript(t, "omitted-baseline", *tap)
		t.Fatalf("handshake failed: %v", err)
	}
	defer func() {
		_ = sess.close()
		probeWriteTranscript(t, "omitted-baseline", *tap)
	}()

	// The turn runs on its own goroutine behind a bounded wait. runTurn can block
	// past its ctx deadline — awaitCodexTurnReview checks ctx.Err() only between
	// lines, and conn.recv() blocks on a channel the reader goroutine only fills
	// while codex is emitting. A turn that goes silent (an approval request we
	// never answer, for instance) therefore hangs rather than deadlining, and an
	// unbounded wait here would take the transcript down with it.
	type turnResult struct{ err error }
	done := make(chan turnResult, 1)
	go func() {
		_, e := sess.runTurn(ctx, codexMethodTurnStart, map[string]any{
			"cwd":    cwd,
			"prompt": "create a file named baseline.txt containing the word hi",
		})
		done <- turnResult{err: e}
	}()

	var runErr error
	stalled := false
	select {
	case r := <-done:
		runErr = r.err
	case <-time.After(150 * time.Second):
		stalled = true
	}
	wrote := probeFileExists(filepath.Join(cwd, "baseline.txt"))
	t.Logf("BASELINE first turn on a fresh thread, sandboxPolicy OMITTED: stalled=%v err=%v baseline.txt=%v", stalled, runErr, wrote)

	if stalled && !wrote {
		t.Logf("BASELINE INCONCLUSIVE-BY-STALL: the turn neither completed nor wrote within 150s. Read the transcript: a pending server→client request (an approval ask) would explain it, and would itself mean an omitted sandboxPolicy is NOT silently write-capable")
		return
	}
	if wrote {
		t.Logf("BASELINE: omission alone is write-capable — the stickiness probe's turn 2 is explained by the DEFAULT, not proven to be inheritance. REQ-CX2-007's per-turn readOnly is still load-bearing (an omitted field does not mean readOnly), but the 'sticky' wording is not what this observation establishes")
		return
	}
	t.Logf("BASELINE: omission alone did NOT write — so the stickiness probe's turn 2 wrote because it INHERITED turn 1's workspaceWrite. Stickiness confirmed as the SPEC states it")
}

// TestCodexLive_ExplicitReadOnlyApprovalStall checks whether the stall observed
// by the baseline probe is reachable through the envelope codex_task ACTUALLY
// sends, rather than only through omission.
//
// The baseline probe sent no sandboxPolicy at all and stalled: codex raised the
// server→client request item/fileChange/requestApproval and waited for an answer
// the session driver never sends. codex_task's non-write path instead sends
// {"type":"readOnly"} explicitly. Both should produce a read-only sandbox, but
// "should" is the inference this probe exists to replace, so the explicit form
// is exercised directly.
func TestCodexLive_ExplicitReadOnlyApprovalStall(t *testing.T) {
	bin := probeLiveCodex(t)
	tap := probeInstallRunner(t)
	cwd := t.TempDir()

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()

	sess, err := openCodexSession(ctx, bin, map[string]any{"cwd": cwd})
	if err != nil {
		probeWriteTranscript(t, "readonly-approval-stall", *tap)
		t.Fatalf("handshake failed: %v", err)
	}
	defer func() {
		_ = sess.close()
		probeWriteTranscript(t, "readonly-approval-stall", *tap)
	}()

	done := make(chan error, 1)
	go func() {
		_, e := sess.runTurn(ctx, codexMethodTurnStart, map[string]any{
			"cwd": cwd,
			// EXACTLY the envelope codex_task builds when write is not granted.
			"sandboxPolicy": codexSandboxPolicy(false),
			"prompt":        "create a file named blocked.txt containing the word hi",
		})
		done <- e
	}()

	select {
	case e := <-done:
		t.Logf("READONLY-STALL turn RETURNED (no stall): err=%v blocked.txt=%v",
			e, probeFileExists(filepath.Join(cwd, "blocked.txt")))
	case <-time.After(120 * time.Second):
		t.Logf("READONLY-STALL turn did NOT return within 120s — the explicit readOnly envelope reaches the same unanswered-approval stall the baseline hit")
	}
}

// TestCodexLive_ReviewStartEmitsTurnStarted is probe item 4: M2 assumed a
// review/start turn emits turn/started like a turn/start does. If it does not,
// a review-path job records an empty turn_id and M4's cancel degrades to
// process termination — a branch that is implemented but whose premise was
// never observed.
//
// The probe deliberately does NOT wait for the review verdict: turn/started is
// the FIRST notification of the turn, so the session is torn down as soon as the
// question is answered. That keeps a full review turn off the bill.
func TestCodexLive_ReviewStartEmitsTurnStarted(t *testing.T) {
	bin := probeLiveCodex(t)
	tap := probeInstallRunner(t)
	repo := probeSeedRepo(t)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	sess, err := openCodexSession(ctx, bin, map[string]any{"cwd": repo})
	if err != nil {
		probeWriteTranscript(t, "review-turnstarted", *tap)
		t.Fatalf("handshake failed: %v", err)
	}
	defer func() {
		_ = sess.close()
		probeWriteTranscript(t, "review-turnstarted", *tap)
	}()

	seen := make(chan string, 1)
	sess.onTurnStarted = func(id string) {
		select {
		case seen <- id:
		default:
		}
	}
	go func() {
		_, _ = sess.runTurn(ctx, codexMethodReviewStart, map[string]any{
			"cwd":    repo,
			"target": codexTargetUncommitted,
		})
	}()

	select {
	case id := <-seen:
		t.Logf("ITEM-4 CONFIRMED: review/start emitted turn/started, turn.id=%q — review-path jobs can record a turnId", id)
	case <-time.After(120 * time.Second):
		t.Logf("ITEM-4 NOT OBSERVED within 120s: no turn/started on the review/start path. If this is the steady behaviour, review-path jobs carry an empty turn_id and M4 cancel degrades to termination-only — which is the branch already implemented")
	}
}

// probeSeedRepo builds a throwaway git repo with one staged change, so
// review/start has something to review.
func probeSeedRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	git("init")
	git("config", "user.email", "t@t.test")
	git("config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(repo, "seed.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write seed: %v", err)
	}
	git("add", "seed.go")
	git("commit", "-m", "seed")
	if err := os.WriteFile(filepath.Join(repo, "add.go"), []byte("package main\n\nfunc Add(a, b int) int { return a + b }\n"), 0o644); err != nil {
		t.Fatalf("write add: %v", err)
	}
	git("add", "add.go")
	return repo
}
