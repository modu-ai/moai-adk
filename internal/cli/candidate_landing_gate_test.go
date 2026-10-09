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
		prevRuns, prevList := candidateGhRunsFn, candidateGhRunsListFn
		candidateGhRunsFn = func(root, branch string) ([]factory.CandidateRunState, error) {
			return []factory.CandidateRunState{
				{RunID: "r-old", HeadSHA: "deadbeef", Ref: rec.CandidateBranch, Status: "completed", Conclusion: "success"},
				{RunID: "7001", HeadSHA: rec.CandidateSHA, Ref: rec.CandidateBranch, Status: "completed", Conclusion: "success"},
			}, nil
		}
		// The required-check read is stubbed too: the run's jobs all
		// succeed, so the walk's verdict reflects the binding run (the
		// fixture root has no SSoT set — the fallback set is these jobs).
		candidateGhRunsListFn = func(dir string, args ...string) (string, error) {
			return `{"jobs":[{"name":"own-check","conclusion":"success"},{"name":"Guard Bundle","conclusion":"success"}]}`, nil
		}
		t.Cleanup(func() { candidateGhRunsFn, candidateGhRunsListFn = prevRuns, prevList })
		updated, wrote, err := runCandidateObservation(root, "t9001")
		if err != nil || !wrote {
			t.Fatalf("wrote=%v err=%v", wrote, err)
		}
		if updated.Verdict != factory.CandidateVerdictGreen || updated.RunID != "7001" {
			t.Errorf("observed: verdict %q run %q, want green/7001", updated.Verdict, updated.RunID)
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

// TestObserveCandidateRunsFirstBindingRunDecides pins card t1478 Finding 1
// (REQ-CCI-010): the FIRST run that binds the record decides, whether or not
// it yields a verdict. A newest binding run with an empty conclusion or still
// in progress must not let an older binding run's success turn the pending
// record green. A run that binds nothing is skipped without stopping the walk.
func TestObserveCandidateRunsFirstBindingRunDecides(t *testing.T) {
	const cardID, pinned, candidate = "t9002", "pin-2", "cand-2"
	now := time.Date(2026, 10, 10, 9, 0, 0, 0, time.UTC)
	newRoot := func(t *testing.T) string {
		t.Helper()
		root := t.TempDir()
		if err := factory.WriteCandidateRecord(root, factory.CandidateRecord{
			CardID: cardID, PinnedSHA: pinned, CandidateSHA: candidate, CandidateBranch: "ci/" + cardID,
			IntegrationBranch: "develop", IntegrationTip: "tip-2",
			Verdict: factory.CandidateVerdictPending, PushedAt: "2026-10-10T08:00:00Z",
		}); err != nil {
			t.Fatal(err)
		}
		return root
	}
	// Run 100 succeeds on every required job; every other run's job read
	// fails, so its run-level conclusion is the only verdict input.
	stubJobReads := func(t *testing.T) {
		t.Helper()
		prev := candidateGhRunsListFn
		candidateGhRunsListFn = func(dir string, args ...string) (string, error) {
			if len(args) > 2 && args[2] == "100" {
				return `{"jobs":[{"name":"own-check","conclusion":"success"},{"name":"Guard Bundle","conclusion":"success"}]}`, nil
			}
			return "", errors.New("gh run view: unavailable (test double)")
		}
		t.Cleanup(func() { candidateGhRunsListFn = prev })
	}
	binding := func(id, status, conclusion string) factory.CandidateRunState {
		return factory.CandidateRunState{RunID: id, HeadSHA: candidate, Ref: "ci/" + cardID, Status: status, Conclusion: conclusion}
	}

	t.Run("newest binding run with an empty conclusion decides: an older success writes nothing", func(t *testing.T) {
		root := newRoot(t)
		stubJobReads(t)
		runs := []factory.CandidateRunState{binding("200", "completed", ""), binding("100", "completed", "success")}
		_, wrote, err := observeCandidateRuns(root, cardID, pinned, "ci/**", runs, now)
		if err != nil {
			t.Fatalf("observe: %v", err)
		}
		stored, err := factory.ReadCandidateRecord(root, cardID, pinned)
		if err != nil {
			t.Fatal(err)
		}
		if wrote || stored.Verdict != factory.CandidateVerdictPending {
			t.Errorf("wrote=%v, stored verdict %q: the older run 100 must not decide over the newer binding run 200 — want wrote=false, pending", wrote, stored.Verdict)
		}
	})

	t.Run("newest binding run still in progress decides: an older success writes nothing", func(t *testing.T) {
		root := newRoot(t)
		stubJobReads(t)
		runs := []factory.CandidateRunState{binding("200", "in_progress", ""), binding("100", "completed", "success")}
		_, wrote, err := observeCandidateRuns(root, cardID, pinned, "ci/**", runs, now)
		if err != nil {
			t.Fatalf("observe: %v", err)
		}
		stored, err := factory.ReadCandidateRecord(root, cardID, pinned)
		if err != nil {
			t.Fatal(err)
		}
		if wrote || stored.Verdict != factory.CandidateVerdictPending {
			t.Errorf("wrote=%v, stored verdict %q: the in-progress binding run 200 decides — want wrote=false, pending", wrote, stored.Verdict)
		}
	})

	t.Run("newest binding run failed decides red; a non-binding run is skipped without stopping the walk", func(t *testing.T) {
		root := newRoot(t)
		stubJobReads(t)
		runs := []factory.CandidateRunState{
			{RunID: "300", HeadSHA: "other-sha", Ref: "ci/" + cardID, Status: "completed", Conclusion: "success"},
			binding("200", "completed", "failure"),
			binding("100", "completed", "success"),
		}
		rec, wrote, err := observeCandidateRuns(root, cardID, pinned, "ci/**", runs, now)
		if err != nil || !wrote {
			t.Fatalf("wrote=%v err=%v: the failed binding run 200 must record red", wrote, err)
		}
		if rec.Verdict != factory.CandidateVerdictRed || rec.RunID != "200" {
			t.Errorf("observed: verdict %q run %q, want red from run 200", rec.Verdict, rec.RunID)
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
	prevRuns, prevList := candidateGhRunsFn, candidateGhRunsListFn
	candidateGhRunsFn = func(root, branch string) ([]factory.CandidateRunState, error) {
		return []factory.CandidateRunState{
			{RunID: "1", HeadSHA: rec.CandidateSHA, Ref: rec.CandidateBranch, Status: "completed", Conclusion: "success"},
		}, nil
	}
	// The required-check read is stubbed too (the fixture root has no SSoT
	// set — the fallback set is these jobs, all green).
	candidateGhRunsListFn = func(dir string, args ...string) (string, error) {
		return `{"jobs":[{"name":"own-check","conclusion":"success"},{"name":"Guard Bundle","conclusion":"success"}]}`, nil
	}
	t.Cleanup(func() { candidateGhRunsFn, candidateGhRunsListFn = prevRuns, prevList })

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

// TestGhRunStatesMapsRunAttempt pins card t1478 Finding 2: the run list asks
// for each run's attempt and maps it into the run state, so the observation
// orders a re-run against the attempt it recorded (a re-run keeps its id).
func TestGhRunStatesMapsRunAttempt(t *testing.T) {
	var gotArgs []string
	prev := candidateGhCommandFn
	candidateGhCommandFn = func(dir string, args ...string) (string, error) {
		gotArgs = args
		return `[{"databaseId":200,"headSha":"c1","headBranch":"ci/t9001","status":"completed","conclusion":"success","attempt":2}]`, nil
	}
	t.Cleanup(func() { candidateGhCommandFn = prev })
	runs, err := ghRunStates(t.TempDir(), "ci/t9001")
	if err != nil {
		t.Fatalf("ghRunStates: %v", err)
	}
	if joined := strings.Join(gotArgs, " "); !strings.Contains(joined, "attempt") {
		t.Errorf("gh args %q: want the run query to ask for the attempt field", joined)
	}
	if len(runs) != 1 || runs[0].RunID != "200" || runs[0].Attempt != 2 {
		t.Errorf("runs %+v: want run 200 at attempt 2", runs)
	}
}

// TestCandidateRequiredJobVerdict pins the required-check verdict (the
// M4 landing-gate repairs): the set comes from required-checks.yml keyed
// by the candidate-run branch pattern (ci/** — main's set names checks a
// candidate push never publishes), the judged surface is the OBSERVED
// run's own jobs (another run's checks can never satisfy this run's set),
// advisory race failures do not red a required-green run, the matrix-skip
// pair (test + skip-marker sharing one name) reads success, an absent
// context fails closed naming itself, the Guard Bundle check admits by
// workflow.candidate_ci.guard_bundle_required, and a read failure never
// mints green.
func TestCandidateRequiredJobVerdict(t *testing.T) {
	root := t.TempDir()
	checksDir := filepath.Join(root, ".github")
	if err := os.MkdirAll(checksDir, 0o750); err != nil {
		t.Fatal(err)
	}
	writeSSoT := func(yaml string) {
		t.Helper()
		if err := os.MkdirAll(checksDir, 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(checksDir, "required-checks.yml"), []byte(yaml), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	writeSSoT("branches:\n  \"ci/**\":\n    contexts:\n      - Lint\n      - \"Test (ubuntu-latest)\"\n")
	jobsJSON := func(jobs map[string]string) string {
		type job struct {
			Name       string `json:"name"`
			Conclusion string `json:"conclusion"`
		}
		var out []job
		for name, c := range jobs {
			out = append(out, job{Name: name, Conclusion: c})
		}
		raw, _ := json.Marshal(map[string]any{"jobs": out})
		return string(raw)
	}
	seedJobs := func(jobs map[string]string) (cleanup func()) {
		prevList := candidateGhRunsListFn
		candidateGhRunsListFn = func(dir string, args ...string) (string, error) {
			return jobsJSON(jobs), nil
		}
		return func() { candidateGhRunsListFn = prevList }
	}

	t.Run("advisory race failure with required green reads success", func(t *testing.T) {
		cleanup := seedJobs(map[string]string{
			"Lint":                 "success",
			"Test (ubuntu-latest)": "success",
			"Guard Bundle":         "success",
			"Race Test 1":          "failure",
		})
		defer cleanup()
		verdict := candidateRunVerdict(root, "run-1", "completed", "failure", "ci/**")
		if verdict.conclusion != "success" || !verdict.authoritative {
			t.Errorf("verdict %+v: want success (required green; the race job is advisory), authoritative", verdict)
		}
	})

	t.Run("required failure reads failure", func(t *testing.T) {
		cleanup := seedJobs(map[string]string{
			"Lint":                 "failure",
			"Test (ubuntu-latest)": "success",
			"Guard Bundle":         "success",
			"Race Test 1":          "success",
		})
		defer cleanup()
		verdict := candidateRunVerdict(root, "run-1", "completed", "success", "ci/**")
		if verdict.conclusion != "failure" || !verdict.authoritative {
			t.Errorf("verdict %+v: want failure (Lint is required), authoritative", verdict)
		}
	})

	t.Run("another run's success cannot satisfy this run's failure", func(t *testing.T) {
		// The observed-run binding: the judged surface is THIS run's jobs
		// alone — a re-run elsewhere on the same SHA that succeeded can
		// never launder the observed run's required failure.
		cleanup := seedJobs(map[string]string{
			"Lint":                 "failure",
			"Test (ubuntu-latest)": "success",
			"Guard Bundle":         "success",
		})
		defer cleanup()
		verdict := candidateRunVerdict(root, "run-1", "completed", "success", "ci/**")
		if verdict.conclusion != "failure" {
			t.Errorf("verdict %+v: want failure (the observed run's Lint failed)", verdict)
		}
	})

	t.Run("ci/** set applies where main's set would hold every candidate red", func(t *testing.T) {
		// The P1 repair: main's contexts name checks a ci/** push never
		// publishes (codeql.yml's Analyze, the release gate). The ci/**
		// key's set is what the candidate is judged by — a main-keyed
		// judgment would red this green run.
		ssot := "branches:\n  main:\n    contexts:\n      - Lint\n      - \"Analyze (Go) (go)\"\n      - \"Release PR Multi-OS Gate\"\n  \"ci/**\":\n    contexts:\n      - Lint\n      - \"Test (ubuntu-latest)\"\n"
		writeSSoT(ssot)
		cleanup := seedJobs(map[string]string{
			"Lint":                 "success",
			"Test (ubuntu-latest)": "success",
			"Guard Bundle":         "success",
		})
		defer cleanup()
		verdict := candidateRunVerdict(root, "run-1", "completed", "failure", "ci/**")
		if verdict.conclusion != "success" || !verdict.authoritative {
			t.Errorf("verdict %+v: want success under the ci/** set (Analyze/release never publish on a candidate)", verdict)
		}
	})

	t.Run("red guard bundle reds the verdict while the key is true", func(t *testing.T) {
		// AC-CCI-008-2's true branch: guard_bundle_required defaults true —
		// a red bundle reds the candidate verdict (same gating as when the
		// guards rode the ordinary suite).
		cleanup := seedJobs(map[string]string{
			"Lint":                 "success",
			"Test (ubuntu-latest)": "success",
			"Guard Bundle":         "failure",
		})
		defer cleanup()
		verdict := candidateRunVerdict(root, "run-1", "completed", "success", "ci/**")
		if verdict.conclusion != "failure" || !verdict.authoritative {
			t.Errorf("verdict %+v: want failure (Guard Bundle red + key true)", verdict)
		}
		if !strings.Contains(verdict.why, "Guard Bundle") {
			t.Errorf("why %q: want the failed bundle named", verdict.why)
		}
	})

	t.Run("key false admits a red bundle without reding the verdict", func(t *testing.T) {
		// AC-CCI-008-2's false branch: the bundle's red is visible and
		// recorded but does not red the verdict.
		dir := filepath.Join(root, ".moai", "config", "sections")
		if err := os.MkdirAll(dir, 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte("workflow:\n  candidate_ci:\n    enabled: true\n    guard_bundle_required: false\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		cleanup := seedJobs(map[string]string{
			"Lint":                 "success",
			"Test (ubuntu-latest)": "success",
			"Guard Bundle":         "failure",
		})
		defer cleanup()
		verdict := candidateRunVerdict(root, "run-1", "completed", "success", "ci/**")
		if verdict.conclusion != "success" {
			t.Errorf("verdict %+v: want success (key false — the bundle does not gate)", verdict)
		}
	})

	t.Run("fallback set applies the same bundle admission", func(t *testing.T) {
		// The P2 repair: with no SSoT set for the branch, the fallback
		// (the run's own jobs) drops the Guard Bundle check under the same
		// key policy — a key-false bundle failure must not red through the
		// fallback either.
		if err := os.RemoveAll(filepath.Join(root, ".github")); err != nil {
			t.Fatal(err)
		}
		dir := filepath.Join(root, ".moai", "config", "sections")
		if err := os.MkdirAll(dir, 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte("workflow:\n  candidate_ci:\n    enabled: true\n    guard_bundle_required: false\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		cleanup := seedJobs(map[string]string{
			"Lint":         "success",
			"Guard Bundle": "failure",
		})
		defer cleanup()
		verdict := candidateRunVerdict(root, "run-1", "completed", "success", "release/v9.9")
		if verdict.conclusion != "success" {
			t.Errorf("verdict %+v: want success (fallback honors key false)", verdict)
		}
	})

	t.Run("skipped companion does not fail a name a success answers", func(t *testing.T) {
		// The matrix-skip pair: test and test-skip-marker live in the SAME
		// run and publish the SAME name, exactly one executed — a skipped
		// instance is not a failure when a successful execution of the
		// name exists in the run.
		prevList := candidateGhRunsListFn
		candidateGhRunsListFn = func(dir string, args ...string) (string, error) {
			return `{"jobs":[{"name":"Lint","conclusion":"success"},{"name":"Test (ubuntu-latest)","conclusion":"success"},{"name":"Test (ubuntu-latest)","conclusion":"skipped"},{"name":"Guard Bundle","conclusion":"success"}]}`, nil
		}
		t.Cleanup(func() { candidateGhRunsListFn = prevList })
		verdict := candidateRunVerdict(root, "run-1", "completed", "failure", "ci/**")
		if verdict.conclusion != "success" {
			t.Errorf("verdict %+v: want success (the skip pair's real execution succeeded)", verdict)
		}
	})

	t.Run("absent required context fails closed naming itself", func(t *testing.T) {
		// The P1 repair: a context the observed run never carried must not
		// be skipped — an absent context fails closed with its name.
		ssot := "branches:\n  \"ci/**\":\n    contexts:\n      - Lint\n      - \"Test (ubuntu-latest)\"\n      - \"Release PR Multi-OS Gate\"\n"
		writeSSoT(ssot)
		cleanup := seedJobs(map[string]string{
			"Lint":                 "success",
			"Test (ubuntu-latest)": "success",
			"Guard Bundle":         "success",
		})
		defer cleanup()
		verdict := candidateRunVerdict(root, "run-1", "completed", "success", "ci/**")
		if verdict.conclusion != "failure" || !verdict.authoritative {
			t.Errorf("verdict %+v: want failure — Release PR Multi-OS Gate is absent from the observed run", verdict)
		}
		if !strings.Contains(verdict.why, "Release PR Multi-OS Gate") {
			t.Errorf("why %q: want the absent context named", verdict.why)
		}
	})

	t.Run("gh read failure never mints green", func(t *testing.T) {
		// The P1 repair: a failed required-check read (403 injected) with a
		// run-level SUCCESS records NOTHING — read uncertainty is never
		// green. A failed read with a run-level failure still falls back to
		// red (the safe direction).
		prevList := candidateGhRunsListFn
		candidateGhRunsListFn = func(dir string, args ...string) (string, error) {
			return "", errors.New("gh api: HTTP 403 (test double)")
		}
		t.Cleanup(func() { candidateGhRunsListFn = prevList })
		verdict := candidateRunVerdict(root, "run-1", "completed", "success", "ci/**")
		if verdict.conclusion != "" || verdict.authoritative {
			t.Errorf("verdict %+v: want the empty conclusion (nothing recorded) — read uncertainty is never green", verdict)
		}
		verdictRed := candidateRunVerdict(root, "run-1", "completed", "failure", "ci/**")
		if verdictRed.conclusion != "failure" || verdictRed.authoritative {
			t.Errorf("verdict %+v: want the run-level fallback (failure), not authoritative", verdictRed)
		}
	})
}

// TestCandidateObserveWalkNeverGreenOnFailedRead pins the observe-level
// consequence of the never-green rule: a gh required-check read failure
// with a run-level success leaves the record pending — wrote=false.
func TestCandidateObserveWalkNeverGreenOnFailedRead(t *testing.T) {
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
	prevRuns, prevList := candidateGhRunsFn, candidateGhRunsListFn
	candidateGhRunsFn = func(root, branch string) ([]factory.CandidateRunState, error) {
		return []factory.CandidateRunState{
			{RunID: "1", HeadSHA: rec.CandidateSHA, Ref: rec.CandidateBranch, Status: "completed", Conclusion: "success"},
		}, nil
	}
	candidateGhRunsListFn = func(dir string, args ...string) (string, error) {
		return "", errors.New("gh run view: HTTP 403 (test double)")
	}
	t.Cleanup(func() { candidateGhRunsFn, candidateGhRunsListFn = prevRuns, prevList })

	_, wrote, err := runCandidateObservation(root, "t9001")
	if err != nil {
		t.Fatalf("observe walk: %v", err)
	}
	if wrote {
		t.Fatal("a failed required-check read recorded a verdict — read uncertainty is never green")
	}
	stored, err := factory.ReadCandidateRecord(root, "t9001", rec.PinnedSHA)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Verdict != factory.CandidateVerdictPending {
		t.Errorf("verdict after the failed read: %q, want pending", stored.Verdict)
	}
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

// seedWaitCandidate writes one card's candidate record with the given verdict;
// the wait and complete tests judge the per-card hold against it (card t1478
// Finding 4).
func seedWaitCandidate(t *testing.T, root, card, verdict string) {
	t.Helper()
	if err := factory.WriteCandidateRecord(root, factory.CandidateRecord{
		CardID: card, PinnedSHA: "pin-" + card, CandidateSHA: "cand-" + card,
		IntegrationBranch: "develop", IntegrationTip: "tip-" + card,
		CandidateBranch: "ci/" + card, Verdict: verdict, PushedAt: "2026-10-10T08:00:00Z",
	}); err != nil {
		t.Fatal(err)
	}
}

// waitForQueueLen polls the window record until n tickets are queued.
func waitForQueueLen(t *testing.T, root string, n int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		if lock, err := factory.ReadIntegrationLock(root); err == nil && lock != nil && len(lock.Queue) >= n {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("the queue never reached %d ticket(s)", n)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// TestWaitPromotionRefusesCardTurnedRedWhileQueued pins card t1478 Finding 4
// (a): a card that queued via --wait while its candidate was pending, and whose
// candidate turns red before the window is promoted, is refused at promotion.
// The window never ends up held by that card, and the refused waiter leaves the
// queue, so the ticket behind it is promoted in the same refresh.
func TestWaitPromotionRefusesCardTurnedRedWhileQueued(t *testing.T) {
	root := waitTestRoot(t)
	writeCandidateCIConfig(t, root, "true", "true")
	oldInterval := integrationWaitPollInterval
	integrationWaitPollInterval = 10 * time.Millisecond
	t.Cleanup(func() { integrationWaitPollInterval = oldInterval })
	oldClock := factory.WindowClock
	clockAt := time.Date(2026, 10, 10, 9, 0, 0, 0, time.UTC)
	factory.WindowClock = func() time.Time { return clockAt }
	t.Cleanup(func() { factory.WindowClock = oldClock })
	waitHolder(t, root, "sess-a")

	// tA's candidate is PENDING when it queues.
	seedWaitCandidate(t, root, "tA", factory.CandidateVerdictPending)
	doneA := make(chan error, 1)
	go func() {
		doneA <- integrationWaitInQueue(root, "sess-ta", factory.IntegrationTicket{
			SessionID: "sess-ta", SessionName: "lane-ta", Card: "tA",
			OwnerPID: os.Getpid(), WaiterPID: os.Getpid(),
		}, 60*time.Minute, nil)
	}()
	waitForQueueLen(t, root, 1)

	// tB queues behind tA; it has no candidate at all.
	doneB := make(chan error, 1)
	go func() {
		doneB <- integrationWaitInQueue(root, "sess-tb", factory.IntegrationTicket{
			SessionID: "sess-tb", SessionName: "lane-tb", Card: "tB",
			OwnerPID: os.Getpid(), WaiterPID: os.Getpid(),
		}, 60*time.Minute, nil)
	}()
	waitForQueueLen(t, root, 2)

	// tA's candidate turns RED before the window is promoted.
	seedWaitCandidate(t, root, "tA", factory.CandidateVerdictRed)

	// The holder releases: the promotion runs.
	if _, err := factory.ReleaseIntegrationLock(root, "sess-a", os.Getpid(), false); err != nil {
		t.Fatal(err)
	}

	// The refused waiter does not block the queue behind it: tB is promoted.
	select {
	case err := <-doneB:
		if err != nil {
			t.Fatalf("tB must be promoted past the refused tA: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("tB did not return: the refused waiter blocked the queue behind it")
	}
	lock, err := factory.ReadIntegrationLock(root)
	if err != nil {
		t.Fatal(err)
	}
	if lock.SessionID != "sess-tb" {
		t.Fatalf("the window is held by %q on card %q after the promotion, want sess-tb: a red card must never take the window", lock.SessionID, lock.Card)
	}

	// The refused waiter exits non-zero naming the hold.
	select {
	case err := <-doneA:
		if err == nil {
			t.Fatal("the refused waiter tA returned success: its card was granted the window while red")
		}
		if !strings.Contains(err.Error(), "red") || !strings.Contains(err.Error(), "tA") {
			t.Errorf("refusal %q: want the candidate hold naming card tA and its red verdict", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the refused waiter did not exit")
	}
}

// TestCompleteAcquisitionRefusesRedCandidate pins card t1478 Finding 4 (b):
// factory complete's own window acquisition for a red card is refused with the
// landing refusal's exit code, and the acquisition is never attempted.
func TestCompleteAcquisitionRefusesRedCandidate(t *testing.T) {
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
	attempts := 0
	factoryCompleteAcquireHook = func() { attempts++ }
	t.Cleanup(func() { factoryCompleteAcquireHook = nil })

	_, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
	if err == nil {
		t.Fatal("complete was not refused for a red card")
	}
	if code := exitCodeOf(t, err); code != factory.MergeExitLandingRefused {
		t.Errorf("exit code: %d — want %d (MergeExitLandingRefused)", code, factory.MergeExitLandingRefused)
	}
	if attempts != 0 {
		t.Errorf("complete attempted the window acquisition %d time(s) for a red card — the refusal must precede every grant", attempts)
	}
	if lock, readErr := factory.ReadIntegrationLock(root); readErr == nil && lock != nil && lock.Held() {
		t.Errorf("the window is held by %q after a refused complete — want no grant", lock.SessionID)
	}
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
