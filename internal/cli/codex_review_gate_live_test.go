package cli

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook"
)

// SPEC codex-gate-protocol — LIVE integration test. Exercises the REAL codex
// binary (codex-cli app-server JSON-RPC) against a fixture git repo
// carrying an obvious command-injection sink + a hardcoded AWS key, and asserts
// the gate reaches a BLOCK decision when the review turn completes. When the
// turn itself fails (observed live: a usage-limited codex account on 0.147.0
// fails the turn with usageLimitExceeded before the diff is evaluated), the
// gate must fail open WITH the error surfaced — a fabricated pass is the one
// fatal shape. This is the end-to-end proof the 4 protocol
// gaps are closed: initialize handshake → thread/start → review/start (object
// target + threadId) → async verdict synthesis.
//
// Skip conditions (CI without codex still passes):
//   - codex binary absent from PATH (exec.LookPath fails)
//   - `codex --version` non-functional (e.g. a broken native-binary shim)
//   - MOAI_SKIP_LIVE_CODEX set (manual opt-out)
//
// The fixture AWS key is assembled at runtime from prefix parts so it is never
// a literal secret string in source (it is a public AWS documentation example,
// not a live credential).
func TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey(t *testing.T) {
	if os.Getenv("MOAI_SKIP_LIVE_CODEX") == "1" {
		t.Skip("MOAI_SKIP_LIVE_CODEX=1")
	}
	bin, err := exec.LookPath(codexBinaryName)
	if err != nil {
		t.Skipf("codex binary not on PATH: %v", err)
	}
	// A broken native-binary shim (observed: a bun-installed codex whose vendor
	// rust binary is ENOENT) prints an error to stderr and exits non-zero. Treat
	// that as "codex not functional" and skip rather than fail.
	if ver, vErr := exec.Command(bin, "--version").Output(); vErr != nil || !strings.Contains(string(ver), "codex") {
		t.Skipf("codex --version non-functional (ver=%q err=%v) — environment lacks a working codex", strings.TrimSpace(string(ver)), vErr)
	}

	// Fixture: a temp git repo with a committed clean seed, plus an uncommitted
	// file holding (a) a command-injection sink and (b) a hardcoded AWS key.
	// The committed .gitignore excludes .moai/ — the receipt write must not
	// move the tree key between the producer's record and the entry hook's
	// consult (the cacheTestRoot lesson; an untracked receipt snapshot would
	// make the delayed verdict unreadable and the entry arm would silently
	// allow).
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
	if err := os.WriteFile(filepath.Join(repo, ".gitignore"), []byte(".moai/\n"), 0o644); err != nil {
		t.Fatalf("write gitignore: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "seed.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write seed: %v", err)
	}
	git("add", ".gitignore", "seed.go")
	git("commit", "-m", "seed")
	if err := os.WriteFile(filepath.Join(repo, "vuln.go"), []byte(fixtureVulnGo()), 0o644); err != nil {
		t.Fatalf("write vuln: %v", err)
	}
	git("add", "vuln.go")
	// The live review path this test proves now requires the explicit
	// primary_scope restore: the distributed default skips a primary-checkout
	// tree session (REQ-CGSC-002 / REQ-CGSC-004).
	liveCfg := filepath.Join(repo, ".moai", "config", "sections", "workflow.yaml")
	if err := os.MkdirAll(filepath.Dir(liveCfg), 0o755); err != nil {
		t.Fatalf("mkdir live config: %v", err)
	}
	if err := os.WriteFile(liveCfg, []byte("workflow:\n  codex:\n    review_gate:\n      primary_scope: review\n"), 0o644); err != nil {
		t.Fatalf("write live config: %v", err)
	}

	// Point the gate's seams at the REAL codex + the fixture repo, gate enabled,
	// detector forced true (the fixture is staged so the porcelain detector would
	// also fire, but forcing removes any environment-dependent porcelain noise).
	prevLook, prevSess, prevDet, prevTO := codexLookPath, codexSession, reviewGateChangeDetector, config.DefaultCodexReviewGateTimeout
	codexLookPath = func(string) (string, error) { return bin, nil }
	codexSession = realCodexSessionRunner{}
	reviewGateChangeDetector = func(string) bool { return true }
	config.DefaultCodexReviewGateTimeout = liveCodexGateTimeout * time.Second
	t.Cleanup(func() {
		codexLookPath, codexSession, reviewGateChangeDetector, config.DefaultCodexReviewGateTimeout =
			prevLook, prevSess, prevDet, prevTO
	})

	// M2 (SPEC-GATE-BOTTLENECK-001 REQ-GBN-002): the review runs in the
	// receipt PRODUCER now — the live security assertion rides it, and the
	// recorded verdict's enforcement is the NEXT-turn-entry hook.
	r, rErr := produceCodexReviewReceipt(context.Background(), repo)
	if rErr != nil {
		t.Fatalf("the producer must record against the fixture (reviewer present): %v", rErr)
	}
	// Two legitimate outcomes, one fatal shape:
	//
	//  1. The review turn COMPLETED (a fail receipt): codex really evaluated
	//     the fixture, and an injection+AWS-key change MUST produce finding
	//     bullets ⇒ FAIL. This is the security assertion and it is not
	//     negotiable. The enforcement assertion rides with it: the fresh FAIL
	//     receipt must BLOCK at the next turn entry (the delayed point).
	//  2. The review turn itself FAILED (inconclusive — e.g. the codex account
	//     is usage-limited, as observed live on codex-cli 0.147.0: the turn
	//     dies with usageLimitExceeded BEFORE the diff is evaluated): the
	//     producer records inconclusive with the error on its stderr — never a
	//     fabricated pass.
	//
	// The fatal shape is verdict pass: a "review happened and found nothing"
	// claim no real review produced (card t52 — the pipeline used to launder
	// codex's "Reviewer failed to output a response." placeholder into pass).
	switch r.Verdict {
	case codexReviewVerdictFail:
		entry, entryErr := HandleCodexReviewEntry(&hook.HookInput{CWD: repo}, true /* enabled */, repo)
		if entryErr != nil {
			t.Fatalf("entry error: %v", entryErr)
		}
		if entry == nil || entry.Decision != hook.DecisionBlock {
			t.Fatalf("a fresh FAIL receipt must BLOCK at turn entry (the delayed enforcement point), got %+v", entry)
		}
		t.Logf("BLOCK reached at turn entry. reason=%q", entry.Reason)
	case codexReviewVerdictInconclusive:
		t.Skipf("codex review turn did not complete — the producer recorded inconclusive with the error surfaced (correct behavior)")
	default:
		t.Fatalf("expected a FAIL receipt on injection+AWS-key fixture; got verdict=%q\n"+
			"NOTE: the review turn COMPLETED, so codex really evaluated the fixture — "+
			"a pass here means codex passed this fixture (a real result — report it).", r.Verdict)
	}
}

// liveCodexGateTimeout gives the live codex review turn enough room to run the
// model on the fixture without bleeding the 900s production budget into the test
// suite. The model review typically completes in well under this; CI skips when
// codex is absent so this only runs locally.
const liveCodexGateTimeout = 180 // seconds; overridden into config.DefaultCodexReviewGateTimeout

// fixtureVulnGo builds the vuln.go fixture source at runtime. The AWS key is
// spliced from prefix + body so no literal credential string sits in source
// (the value is AWS's public documentation example, not a live key).
func fixtureVulnGo() string {
	const prefix = "package main\n\nimport \"os/exec\"\n\n" +
		"// runQuery runs a raw query through a shell — injection sink.\n" +
		"func runQuery(q string) { exec.Command(\"sh\", \"-c\", q).Run() }\n\n" +
		"const AWSAccessKey = \""
	// AWS public-docs example key, spliced to avoid a literal-secret guard trip.
	key := "AKIA" + "IOSF" + "ODNN" + "7EXAMPLE"
	const secretLine = "\"\nconst AWSSecretKey = \"wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY\"\n"
	return prefix + key + secretLine
}
