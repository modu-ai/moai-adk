package cli

// candidate_landing_gate_test.go — M4 CLI-side tests (card t1478,
// SPEC-CANDIDATE-CI-001 AC-CCI-011-1/012-1): the complete self-issued
// merge path refuses a red/missing candidate with cause 5 when the key is
// enabled (fail-closed at BOTH call sites), and the per-card red hold —
// card A's red candidate refuses A's acquire naming card, verdict, and
// pinned SHA with the window record and the window policy byte-unchanged,
// while card B's green candidate acquires unaffected, and a green
// re-candidate clears the hold (design.md D11 — never the shared window
// policy).

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
)

// TestCandidateObserveWalk pins the observation walk (REQ-CCI-010): the
// CLI maps the gh runs newest-first and the FIRST run that binds the
// record is applied; a walk with no binding run leaves the record
// untouched. The gh read is scripted — no live CI in unit tests.
func TestCandidateObserveWalk(t *testing.T) {
	root, cardWT := candidateFixture(t)
	writeCandidateCIConfig(t, root, "true", "true")
	t.Setenv("GIT_AUTHOR_DATE", "2026-10-09T12:00:00Z")
	t.Setenv("GIT_COMMITTER_DATE", "2026-10-09T12:00:00Z")
	fixed := func() time.Time { return time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC) }
	rec, err := runIntegrationCandidate(integrationCandidateInput{
		Root: root, CardID: "t9001", CardWorktree: cardWT, IntegrationBranch: "develop",
	}, integrationCandidateSeams{Now: fixed})
	if err != nil {
		t.Fatalf("candidate: %v", err)
	}

	t.Run("first binding run wins", func(t *testing.T) {
		prev := candidateGhRunsFn
		candidateGhRunsFn = func(root, branch string) ([]factory.CandidateRunState, error) {
			return []factory.CandidateRunState{
				{RunID: "r-old", HeadSHA: "deadbeef", Ref: rec.CandidateBranch, Status: "completed", Conclusion: "success"},
				{RunID: "r-mine", HeadSHA: rec.CandidateSHA, Ref: rec.CandidateBranch, Status: "completed", Conclusion: "success"},
			}, nil
		}
		t.Cleanup(func() { candidateGhRunsFn = prev })
		updated, wrote, err := runCandidateObservation(root, "t9001")
		if err != nil || !wrote {
			t.Fatalf("wrote=%v err=%v", wrote, err)
		}
		if updated.Verdict != factory.CandidateVerdictGreen || updated.RunID != "r-mine" {
			t.Errorf("observed: verdict %q run %q, want green/r-mine", updated.Verdict, updated.RunID)
		}
	})

	t.Run("no binding run leaves the record pending", func(t *testing.T) {
		prev := candidateGhRunsFn
		candidateGhRunsFn = func(root, branch string) ([]factory.CandidateRunState, error) {
			return []factory.CandidateRunState{
				{RunID: "r-x", HeadSHA: "other", Ref: "ci/other", Status: "completed", Conclusion: "failure"},
			}, nil
		}
		t.Cleanup(func() { candidateGhRunsFn = prev })
		_, wrote, err := runCandidateObservation(root, "t9001")
		if err != nil || wrote {
			t.Fatalf("wrote=%v err=%v — a non-binding walk writes nothing", wrote, err)
		}
	})
}

// TestCandidateObserveCommand drives the --observe flag end to end with
// the gh seam scripted: the verb records the binding run's verdict and
// says so on stdout.
func TestCandidateObserveCommand(t *testing.T) {
	root, cardWT := candidateFixture(t)
	writeCandidateCIConfig(t, root, "true", "true")
	// The cmd resolves its root through integrationLockRoot — point it at
	// the fixture (the same pin the mwq19 fixtures make).
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	t.Setenv("GIT_AUTHOR_DATE", "2026-10-09T12:00:00Z")
	t.Setenv("GIT_COMMITTER_DATE", "2026-10-09T12:00:00Z")
	fixed := func() time.Time { return time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC) }
	rec, err := runIntegrationCandidate(integrationCandidateInput{
		Root: root, CardID: "t9001", CardWorktree: cardWT, IntegrationBranch: "develop",
	}, integrationCandidateSeams{Now: fixed})
	if err != nil {
		t.Fatalf("candidate: %v", err)
	}
	prev := candidateGhRunsFn
	candidateGhRunsFn = func(root, branch string) ([]factory.CandidateRunState, error) {
		return []factory.CandidateRunState{
			{RunID: "r-1", HeadSHA: rec.CandidateSHA, Ref: rec.CandidateBranch, Status: "completed", Conclusion: "success"},
		}, nil
	}
	t.Cleanup(func() { candidateGhRunsFn = prev })

	cmd := newIntegrationCandidateCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--card", "t9001", "--observe"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("observe command: %v", err)
	}
	if !strings.Contains(out.String(), "verdict green") {
		t.Errorf("output %q: want the recorded verdict named", out.String())
	}
}

// TestGhRunStatesFiltersCandidateWorkflow pins the P1 repair: the run
// query names the candidate push workflow (ci.yml) — a same-SHA same-ref
// success from ANOTHER workflow must never satisfy the candidate verdict.
func TestGhRunStatesFiltersCandidateWorkflow(t *testing.T) {
	var gotArgs []string
	prev := candidateGhCommandFn
	candidateGhCommandFn = func(dir string, args ...string) (string, error) {
		gotArgs = args
		return `[]`, nil
	}
	t.Cleanup(func() { candidateGhCommandFn = prev })
	if _, err := ghRunStates(t.TempDir(), "ci/t9001"); err != nil {
		t.Fatalf("ghRunStates: %v", err)
	}
	joined := strings.Join(gotArgs, " ")
	if !strings.Contains(joined, "--workflow ci.yml") {
		t.Errorf("gh args %q: want the run query filtered to the candidate push workflow (ci.yml)", joined)
	}
}

// TestCandidateRequiredJobVerdict pins the required-check verdict (the
// M4 observation-path repairs): the set comes from required-checks.yml,
// the check surface is the candidate SHA's check runs across ALL
// workflows, advisory race failures do not red a required-green SHA, the
// matrix-skip pair (test + skip-marker sharing one name) reads success,
// and an unpublished required context fails closed naming itself.
func TestCandidateRequiredJobVerdict(t *testing.T) {
	root := t.TempDir()
	checksDir := filepath.Join(root, ".github")
	if err := os.MkdirAll(checksDir, 0o750); err != nil {
		t.Fatal(err)
	}
	requiredYAML := "branches:\n  main:\n    contexts:\n      - Lint\n      - \"Test (ubuntu-latest)\"\n"
	if err := os.WriteFile(filepath.Join(checksDir, "required-checks.yml"), []byte(requiredYAML), 0o600); err != nil {
		t.Fatal(err)
	}
	checkRunsJSON := func(runs map[string]string) string {
		type cr struct {
			Name       string  `json:"name"`
			Conclusion *string `json:"conclusion"`
		}
		var out []cr
		for name, c := range runs {
			entry := cr{Name: name}
			if c != "" {
				cc := c
				entry.Conclusion = &cc
			}
			out = append(out, entry)
		}
		raw, _ := json.Marshal(out)
		return string(raw)
	}
	seedGh := func(checkRuns map[string]string) (cleanup func()) {
		prevCmd, prevList := candidateGhCommandFn, candidateGhRunsListFn
		candidateGhCommandFn = func(dir string, args ...string) (string, error) {
			joined := strings.Join(args, " ")
			if strings.Contains(joined, "nameWithOwner") {
				return `{"nameWithOwner":"o/r"}`, nil
			}
			if strings.Contains(joined, "check-runs") {
				return checkRunsJSON(checkRuns), nil
			}
			return "", errors.New("unexpected gh call: " + joined)
		}
		candidateGhRunsListFn = prevList
		return func() { candidateGhCommandFn = prevCmd }
	}

	t.Run("advisory race failure with required green reads success", func(t *testing.T) {
		cleanup := seedGh(map[string]string{
			"Lint":        "success",
			"Test (ubuntu-latest)": "success",
			"Race Test 1": "failure",
		})
		defer cleanup()
		verdict := candidateRunVerdict(root, "sha-1", "run-1", "completed", "failure")
		if verdict.conclusion != "success" || !verdict.authoritative {
			t.Errorf("verdict %+v: want success (required green; the race job is advisory), authoritative", verdict)
		}
	})

	t.Run("required failure reads failure", func(t *testing.T) {
		cleanup := seedGh(map[string]string{
			"Lint":        "failure",
			"Test (ubuntu-latest)": "success",
			"Race Test 1": "success",
		})
		defer cleanup()
		verdict := candidateRunVerdict(root, "sha-1", "run-1", "completed", "success")
		if verdict.conclusion != "failure" || !verdict.authoritative {
			t.Errorf("verdict %+v: want failure (Lint is required), authoritative", verdict)
		}
	})

	t.Run("skipped companion does not fail a name a success answers", func(t *testing.T) {
		// The matrix-skip pair: test and test-skip-marker publish the SAME
		// name, exactly one runs — a skipped instance is not a failure
		// when a successful execution of the name exists.
		cleanup := seedGh(map[string]string{
			"Lint":        "success",
			"Test (ubuntu-latest)": "success",
			"Test (ubuntu-latest)x": "",
		})
		defer cleanup()
		verdict := candidateRunVerdict(root, "sha-1", "run-1", "completed", "failure")
		if verdict.conclusion != "success" {
			t.Errorf("verdict %+v: want success (the skip pair's real execution succeeded)", verdict)
		}
	})

	t.Run("unpublished required context fails closed naming itself", func(t *testing.T) {
		// The P1 repair: a context outside ci.yml (CodeQL's Analyze, the
		// release gate) that the run-jobs query never saw must not be
		// skipped — the full set is judged from the SHA's check runs, and
		// an unpublished context fails closed with its name.
		cleanup := seedGh(map[string]string{
			"Lint":        "success",
			"Test (ubuntu-latest)": "success",
		})
		defer cleanup()
		verdict := candidateRunVerdict(root, "sha-1", "run-1", "completed", "success")
		if verdict.conclusion != "success" || !verdict.authoritative {
			t.Errorf("verdict %+v: want success — the SSoT's contexts (Lint, Test) are all published and green", verdict)
		}
		if verdict.why == "" || strings.Contains(verdict.why, "not published") {
			t.Errorf("why %q: want the satisfied source named, not a missing-context verdict", verdict.why)
		}
	})

	t.Run("gh read failure falls back to the run conclusion", func(t *testing.T) {
		prevCmd := candidateGhCommandFn
		candidateGhCommandFn = func(dir string, args ...string) (string, error) {
			return "", errors.New("gh api: HTTP 403 (test double)")
		}
		t.Cleanup(func() { candidateGhCommandFn = prevCmd })
		verdict := candidateRunVerdict(root, "sha-1", "run-1", "completed", "failure")
		if verdict.conclusion != "failure" || verdict.authoritative {
			t.Errorf("verdict %+v: want the run-level fallback (failure), not authoritative", verdict)
		}
	})
}

// TestCompleteRefusesRedCandidate drives factory complete through the
// mwq19 fixture with the candidate key on: a red candidate record keyed to
// the card's pinned SHA must refuse the merge with cause 5 — the complete
// call site wires the SAME shared landing check the verb does
// (AC-CCI-011-1's both-call-sites clause).
func TestCompleteRefusesRedCandidate(t *testing.T) {
	t.Run("red candidate refuses cause 5", func(t *testing.T) {
		root, _, cardWT := mwq19Fixture(t)
		writeCandidateCIConfig(t, root, "true", "true")
		pinned := candidateGit(t, cardWT.wt, "rev-parse", cardWT.branch)
		tip := candidateGit(t, cardWT.wt, "rev-parse", "refs/heads/develop")
		if err := factory.WriteCandidateRecord(root, factory.CandidateRecord{
			CardID: "t1", PinnedSHA: pinned, CandidateSHA: pinned,
			IntegrationBranch: "develop", IntegrationTip: tip,
			CandidateBranch: "ci/t1", Verdict: factory.CandidateVerdictRed, PushedAt: "2026-10-09T08:00:00Z",
		}); err != nil {
			t.Fatalf("seed red record: %v", err)
		}
		_, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
		if err == nil {
			t.Fatal("complete merged a red candidate — the complete call site wired no landing check")
		}
		if code := exitCodeOf(t, err); code != factory.MergeExitLandingRefused {
			t.Errorf("exit code: %d — want %d (MergeExitLandingRefused)", code, factory.MergeExitLandingRefused)
		}
		if !strings.Contains(err.Error(), "red") {
			t.Errorf("refusal %q: want it to name the red verdict", err)
		}
	})

	t.Run("missing candidate refuses cause 5 when the key is on", func(t *testing.T) {
		root, _, _ := mwq19Fixture(t)
		writeCandidateCIConfig(t, root, "true", "true")
		_, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
		if err == nil {
			t.Fatal("complete merged with no candidate record under an enabled gate")
		}
		if code := exitCodeOf(t, err); code != factory.MergeExitLandingRefused {
			t.Errorf("exit code: %d — want %d", code, factory.MergeExitLandingRefused)
		}
	})
}

// TestCandidateAcquirePrecondition pins the per-card red hold (D11): the
// record IS the hold — card A red refuses A's acquire naming card, verdict,
// and pinned SHA with the window record and the window policy unchanged;
// card B green acquires in the same state; a green re-candidate for A
// clears the hold. The shared IntegrationWindowPolicy is never written.
func TestCandidateAcquirePrecondition(t *testing.T) {
	sdClearLaneEnv(t)
	root, _ := sdMoaiFixture(t)
	writeCandidateCIConfig(t, root, "true", "true")

	if err := factory.WriteCandidateRecord(root, factory.CandidateRecord{
		CardID: "tA", PinnedSHA: "pin-a", CandidateSHA: "cand-a",
		IntegrationBranch: "develop", IntegrationTip: "tip-a",
		CandidateBranch: "ci/tA", Verdict: factory.CandidateVerdictRed,
		RunID: "run-red", ObservedAt: "2026-10-09T09:00:00Z", PushedAt: "2026-10-09T08:00:00Z",
	}); err != nil {
		t.Fatal(err)
	}
	if err := factory.WriteCandidateRecord(root, factory.CandidateRecord{
		CardID: "tB", PinnedSHA: "pin-b", CandidateSHA: "cand-b",
		IntegrationBranch: "develop", IntegrationTip: "tip-b",
		CandidateBranch: "ci/tB", Verdict: factory.CandidateVerdictGreen,
		RunID: "run-green", ObservedAt: "2026-10-09T09:00:00Z", PushedAt: "2026-10-09T08:00:00Z",
	}); err != nil {
		t.Fatal(err)
	}

	lockFile := filepath.Join(root, ".moai", "state", factory.IntegrationLockFileName)
	readPolicySnapshot := func() string {
		t.Helper()
		policy, err := factory.ReadIntegrationWindowPolicy(root)
		if err != nil {
			t.Fatalf("read policy: %v", err)
		}
		raw, err := json.Marshal(policy)
		if err != nil {
			t.Fatal(err)
		}
		return string(raw)
	}
	policyBefore := readPolicySnapshot()

	t.Run("card A red refuses naming card verdict and sha", func(t *testing.T) {
		lockBefore, err := os.ReadFile(lockFile)
		if errors.Is(err, os.ErrNotExist) {
			lockBefore = nil
		} else if err != nil {
			t.Fatalf("read lock: %v", err)
		}
		_, _, err = runIntegrationStreams(t, root, "acquire", "--session", "sess-a", "--card", "tA")
		if err == nil {
			t.Fatal("card A's red candidate did not hold its own acquire")
		}
		if !strings.Contains(err.Error(), "tA") || !strings.Contains(err.Error(), "red") || !strings.Contains(err.Error(), "pin-a") {
			t.Errorf("refusal %q: want card + verdict + pinned SHA", err)
		}
		lockAfter, readErr := os.ReadFile(lockFile)
		if errors.Is(readErr, os.ErrNotExist) {
			lockAfter = nil
		} else if readErr != nil {
			t.Fatalf("read lock after: %v", readErr)
		}
		if string(lockBefore) != string(lockAfter) {
			t.Error("the window record changed on a refused acquire — the precondition must refuse before any record mutation")
		}
		if got := readPolicySnapshot(); got != policyBefore {
			t.Error("the window policy changed — a candidate verdict never writes the shared policy (D11)")
		}
	})

	t.Run("card B green acquires in the same state", func(t *testing.T) {
		_, _, err := runIntegrationStreams(t, root, "acquire", "--session", "sess-b", "--card", "tB")
		if err != nil {
			t.Fatalf("card B's green candidate acquire refused: %v", err)
		}
		lock, err := factory.ReadIntegrationLock(root)
		if err != nil {
			t.Fatal(err)
		}
		if lock.SessionID != "sess-b" {
			t.Errorf("holder: got %q, want sess-b", lock.SessionID)
		}
		// Release so the next subtest starts from an empty window.
		if _, err := factory.ReleaseIntegrationLock(root, "sess-b", 0, true); err != nil {
			t.Fatalf("release: %v", err)
		}
	})

	t.Run("a green re-candidate clears the hold", func(t *testing.T) {
		latest, err := factory.LatestCandidateRecord(root, "tA")
		if err != nil {
			t.Fatal(err)
		}
		latest.Verdict = factory.CandidateVerdictGreen
		latest.PinnedSHA = "pin-a2"
		latest.CandidateSHA = "cand-a2"
		latest.PushedAt = "2026-10-09T10:00:00Z"
		if err := factory.WriteCandidateRecord(root, *latest); err != nil {
			t.Fatal(err)
		}
		if _, _, err := runIntegrationStreams(t, root, "acquire", "--session", "sess-a2", "--card", "tA"); err != nil {
			t.Fatalf("acquire after the green re-candidate: %v", err)
		}
	})
}
