package cli

// SPEC-CODEX-REVIEW-TARGET-001 AC-CRT-010 — the ROUND-TRIP layer.
//
// The contract layer (codex_review_target_test.go) proves the serialized
// request satisfies the measured schema. That the schema is what codex actually
// enforces is the schema document's claim, not this SPEC's measurement — so
// this test asks the real binary.
//
// The verdict is the POSITIVE fact: the review/start response is not a JSON-RPC
// error and the turn reaches turn/started. "not inconclusive" is never the
// judgment — counting inconclusive as a pass is the defect this card closes.
//
// The session is torn down at turn/started, exactly as the precedent
// TestCodexLive_ReviewStartEmitsTurnStarted does, so no full review turn is
// billed. A schema violation is rejected before the turn starts at all.
//
// Skip conditions are the three established in codex_review_gate_live_test.go:
// binary absent from PATH, `codex --version` non-functional, or
// MOAI_SKIP_LIVE_CODEX=1. A skip is UNOBSERVED, never a pass (acceptance.md
// §E DoD 1, plan.md anti-pattern 7).

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// liveCodexBinary applies the three-condition skip guard and returns a working
// codex binary path.
func liveCodexBinary(t *testing.T) string {
	t.Helper()
	if os.Getenv("MOAI_SKIP_LIVE_CODEX") == "1" {
		t.Skip("MOAI_SKIP_LIVE_CODEX=1 — AC-CRT-010 UNOBSERVED (a skip is not a pass)")
	}
	bin, err := exec.LookPath(codexBinaryName)
	if err != nil {
		t.Skipf("codex binary not on PATH (%v) — AC-CRT-010 UNOBSERVED (a skip is not a pass)", err)
	}
	ver, vErr := exec.Command(bin, "--version").Output()
	if vErr != nil || !strings.Contains(string(ver), "codex") {
		t.Skipf("codex --version non-functional (ver=%q err=%v) — AC-CRT-010 UNOBSERVED (a skip is not a pass)",
			strings.TrimSpace(string(ver)), vErr)
	}
	t.Logf("codex binary=%s version=%s", bin, strings.TrimSpace(string(ver)))
	return bin
}

// seedBaseBranchRepo builds a throwaway git repo shaped for a baseBranch review:
// a base branch (`main`, present both locally and as origin/main with
// origin/HEAD pointing at it) and a feature branch carrying one commit that
// diverged from it. probeSeedRepo plants UNCOMMITTED changes, which is the
// uncommittedChanges fixture — this is its baseBranch sibling, an extension of
// the kit rather than a replacement.
func seedBaseBranchRepo(t *testing.T) string {
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
	git("branch", "-M", "main")

	// origin/main + origin/HEAD so the server-side resolver's step 1 has a real
	// remote default head to read.
	out, err := exec.Command("git", "-C", repo, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatalf("rev-parse HEAD: %v", err)
	}
	git("update-ref", "refs/remotes/origin/main", strings.TrimSpace(string(out)))
	git("symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")

	// A feature branch with one commit, so `main...HEAD` is a non-empty range.
	git("checkout", "-b", "feature")
	if err := os.WriteFile(filepath.Join(repo, "add.go"),
		[]byte("package main\n\nfunc Add(a, b int) int { return a + b }\n"), 0o644); err != nil {
		t.Fatalf("write add: %v", err)
	}
	git("add", "add.go")
	git("commit", "-m", "add")
	return repo
}

// liveWriteTranscript persists the verbatim NDJSON both directions.
//
// It writes under .moai/state/verify/, NOT under .moai/reports/: this test runs
// on any plain `go test ./internal/cli/...` where a working codex is present, so
// a tracked destination would be rewritten on every run and show up as a dirty
// file that is never a change anyone made. The two transcripts the card cites as
// evidence — the pre-fix rejection and the post-fix acceptance — are copies
// pinned under .moai/reports/t399/, which is what keeps those paths resolvable
// at audit time while this one stays disposable.
func liveWriteTranscript(t *testing.T, name string, tap *probeTap) {
	t.Helper()
	if tap == nil {
		t.Logf("no transcript: session never started")
		return
	}
	lines := tap.dump()
	dir := filepath.Join("..", "..", ".moai", "state", "verify", "t399-live")
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

// TestCodexLive_ReviewStartBaseBranchIsNotRejected — AC-CRT-010 / REQ-CRT-002.
//
// Drives the PRODUCTION request assembler (runTurn → buildCodexReviewParams) so
// what is measured is what ships, and judges on two positive observations: the
// review/start ack carried no JSON-RPC error, and the turn reached turn/started.
func TestCodexLive_ReviewStartBaseBranchIsNotRejected(t *testing.T) {
	bin := liveCodexBinary(t)
	tap := probeInstallRunner(t)
	repo := seedBaseBranchRepo(t)

	// NOTE ON COST: unlike the opt-in probe in codex_live_protocol_probe_test.go,
	// this test's skip guard is the three conditions acceptance.md AC-CRT-010
	// names, so it RUNS on a plain `go test ./internal/cli/...` wherever a
	// working codex is installed. That is deliberate — a round trip nobody runs
	// verifies nothing — and it is bounded: the session is cut at turn/started,
	// so what is spent is a handshake and a turn start, not a review.

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	sess, err := openCodexSession(ctx, bin, map[string]any{"cwd": repo})
	if err != nil {
		liveWriteTranscript(t, "basebranch-roundtrip", *tap)
		t.Fatalf("handshake failed: %v", err)
	}
	defer func() {
		_ = sess.close()
		liveWriteTranscript(t, "basebranch-roundtrip", *tap)
	}()

	started := make(chan string, 1)
	sess.onTurnStarted = func(id string) {
		select {
		case started <- id:
		default:
		}
	}

	type turnResult struct {
		out ReviewOutput
		err error
	}
	done := make(chan turnResult, 1)
	go func() {
		out, tErr := sess.runTurn(ctx, codexMethodReviewStart, map[string]any{
			"cwd":    repo,
			"target": codexTargetBaseBranch,
		})
		done <- turnResult{out: out, err: tErr}
	}()

	select {
	case id := <-started:
		// The positive fact. Cut the session here: the question is answered and
		// a full review turn is not billed.
		t.Logf("AC-CRT-010 OBSERVED: live codex accepted the baseBranch review/start; turn.id=%q", id)
	case r := <-done:
		if r.err != nil {
			t.Fatalf("AC-CRT-010 RED: live codex did not accept the baseBranch review/start.\nerror: %v\nsummary: %s",
				r.err, r.out.Summary)
		}
		t.Fatalf("AC-CRT-010 UNRESOLVED: the turn ended without a turn/started notification; summary=%s", r.out.Summary)
	case <-time.After(150 * time.Second):
		t.Fatalf("AC-CRT-010 NOT OBSERVED within 150s: neither turn/started nor a rejection")
	}
}
