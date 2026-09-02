// SPEC-SELECTOR-CENSUS-001 (card t341) — zero-execution veto tests.
//
// Sibling of evidence_writer_test.go: same package, same helpers (bashInput,
// mustRaw, newMoaiTempRoot), split out only because the block is large enough
// to read on its own.
//
// Every runner string below was captured from the runner itself, in this
// worktree, on 2026-08-30 — none is guessed. Runner versions: go1.25
// (`go test`), pytest 8.4.2, cargo 1.94.1, jest 30.4.2, vitest 3.2.7, node
// v22.14.0.
//
// Origin note, NOT evidence: the raw captures behind those versions were
// written to .moai/state/verify/t341/, which is gitignored. Card t341 chose
// not to export them, so they do not resolve outside the authoring worktree
// and no single file can be named for them. The claim above stands on the
// versions and runner strings recorded here, not on that path.
package hook

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/telemetry"
)

// The zero-execution samples are read from the production corpus rather than
// copied here, so a corpus edit cannot leave these tests asserting the old set.
var (
	// go, `go test ./internal/kanban -run TestT341NoSuchTestXYZ -count=1`, shell rc 0.
	sampleGoNoTestsToRun = zeroExecutionSamples["go test"][0]
	// go, `go test ./cmd/...` over a package carrying no _test.go file.
	sampleGoNoTestFiles = zeroExecutionSamples["go test"][1]
)

// Measured genuine-pass outputs — AC-SEC-003's non-firing direction.
const (
	sampleGoPass     = "ok  \tgithub.com/modu-ai/moai-adk/internal/hook\t0.603s\n"
	sampleCargoPass  = "running 2 tests\ntest tests::t2 ... ok\n\ntest result: ok. 2 passed; 0 failed; 0 ignored; 0 measured; 0 filtered out; finished in 0.00s\n"
	samplePytestPass = "...                                                                      [100%]\n3 passed in 0.00s\n"
	// The jest samples are the summary LINE alone, not the whole block. A
	// multi-line block would let a bare-substring "0 passed" veto survive on
	// the strength of its neighbouring "1 passed" line, which is the precise
	// mutant sample (e) exists to kill.
	sampleJestPass    = "Tests:       2 passed, 2 total\n"
	sampleJestPass10  = "Tests:       10 passed, 10 total\n"
	sampleNodeTestTAP = "TAP version 13\n# Subtest: a\nok 1 - a\n1..1\n# tests 1\n# suites 0\n# pass 1\n# fail 0\n"
)

// textPayload builds a wrapped Bash tool_response carrying stdout only — the
// live shape, with NO exit-code signal, so the text branch is what decides.
func textPayload(stdout string) []byte {
	return mustRaw(map[string]any{"stdout": stdout, "interrupted": false})
}

// exitZeroPayload builds the same wrapped shape WITH {"exit_code":0}.
func exitZeroPayload(stdout string) []byte {
	return mustRaw(map[string]any{"stdout": stdout, "exit_code": 0})
}

// AC-SEC-001 — a go run that selected zero tests is not an observed pass, and
// the ledger record it produces is not Outcome=success.
func TestZeroExecution_GoZeroMatchIsNotAnObservedPass(t *testing.T) {
	t.Parallel()
	const cmd = "go test ./internal/config -run TestNope"
	payload := textPayload(sampleGoNoTestsToRun)

	isTest, isPass, isFail := classifyTestCommand(cmd, payload)
	if !isTest {
		t.Fatalf("isTest = false; %q is a test command", cmd)
	}
	if isPass {
		t.Errorf("isPass = true; a run that executed zero tests must not read as an observed pass")
	}
	if isFail {
		t.Errorf("isFail = true; zero execution is absence of signal, not failure")
	}

	rec, ok := buildBashRecord(bashInput("ac-sec-001", cmd, payload), telemetry.UsageRecord{})
	if !ok {
		t.Fatalf("buildBashRecord declined a test command")
	}
	if rec.IsTestPass {
		t.Errorf("record.IsTestPass = true; want false")
	}
	if rec.Outcome == telemetry.OutcomeSuccess {
		t.Errorf("record.Outcome = %q; a zero-execution run must not be recorded as success", rec.Outcome)
	}
}

// AC-SEC-002 — the veto stands ahead of the exit-code branch. A zero-match run
// exits 0, so a veto placed inside deriveFromOutputText alone would change
// nothing on any payload that carries a structured exit code.
func TestZeroExecution_ExitCodeZeroDoesNotCarryItPast(t *testing.T) {
	t.Parallel()
	const cmd = "go test ./internal/config -run TestNope"
	_, isPass, _ := classifyTestCommand(cmd, exitZeroPayload(sampleGoNoTestsToRun))
	if isPass {
		t.Errorf("isPass = true with exit_code 0; the veto must be evaluated before deriveFromExitCode")
	}
}

// AC-SEC-004 + AC-SEC-006 condition 2 — the shared corpus. Both criteria read
// this one variable; no second copy exists. Each sample must actually fire the
// veto, not merely fail to be a pass (an unrecognized string does the latter).
func TestZeroExecution_CorpusSamplesAreVetoed(t *testing.T) {
	t.Parallel()
	for sig, samples := range zeroExecutionSamples {
		if len(samples) == 0 {
			t.Errorf("signature %q has an empty sample list", sig)
			continue
		}
		for i, sample := range samples {
			if !detectZeroExecution(sample) {
				t.Errorf("%s[%d]: detectZeroExecution = false; the sample must fire the veto, not merely fail to be a pass", sig, i)
			}
			cmd := sig + " ."
			isTest, isPass, _ := classifyTestCommand(cmd, textPayload(sample))
			if !isTest {
				t.Errorf("%s[%d]: isTest = false for command %q", sig, i, cmd)
			}
			if isPass {
				t.Errorf("%s[%d]: isPass = true; want false", sig, i)
			}
			if _, gotPass, _ := classifyTestCommand(cmd, exitZeroPayload(sample)); gotPass {
				t.Errorf("%s[%d]: isPass = true with exit_code 0; want false", sig, i)
			}
		}
	}
}

// AC-SEC-006 condition 1 — continued firing. Adding a runner to
// testCommandSignatures without a zero-execution sample turns this red, and the
// failure message names the runner that is missing.
func TestZeroExecution_CorpusCoversEveryRunnerSignature(t *testing.T) {
	t.Parallel()
	var missing []string
	for _, sig := range testCommandSignatures {
		if len(zeroExecutionSamples[sig]) == 0 {
			missing = append(missing, sig)
		}
	}
	if len(missing) > 0 {
		t.Errorf("testCommandSignatures entries with no zero-execution sample: %v", missing)
	}
	for sig := range zeroExecutionSamples {
		found := false
		for _, known := range testCommandSignatures {
			if known == sig {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("corpus key %q is not a testCommandSignatures entry", sig)
		}
	}
}

// AC-SEC-003 — the non-firing direction, across every runner axis the firing
// direction covers. Eleven payloads: five samples x two payload shapes, plus
// the node built-in runner, whose only pass signal today is the exit code.
func TestZeroExecution_GenuinePassIsUnchanged(t *testing.T) {
	t.Parallel()
	paired := []struct {
		name   string
		stdout string
	}{
		{"go", sampleGoPass},
		{"cargo", sampleCargoPass},
		{"pytest", samplePytestPass},
		{"jest_single_digit", sampleJestPass},
		{"jest_double_digit", sampleJestPass10},
	}
	shapes := []struct {
		name string
		wrap func(string) []byte
	}{
		{"text_only", textPayload},
		{"with_exit_zero", exitZeroPayload},
	}
	for _, tc := range paired {
		for _, shape := range shapes {
			t.Run(tc.name+"/"+shape.name, func(t *testing.T) {
				if detectZeroExecution(tc.stdout) {
					t.Fatalf("detectZeroExecution = true on a genuine pass output")
				}
				payload := shape.wrap(tc.stdout)
				isTest, isPass, isFail := classifyTestCommand("go test ./x", payload)
				if !isTest || !isPass || isFail {
					t.Errorf("got (isTest,isPass,isFail) = (%v,%v,%v), want (true,true,false)", isTest, isPass, isFail)
				}
				rec, ok := buildBashRecord(bashInput("ac-sec-003", "go test ./x", payload), telemetry.UsageRecord{})
				if !ok || rec.Outcome != telemetry.OutcomeSuccess {
					t.Errorf("record Outcome = %q, ok = %v; want success/true", rec.Outcome, ok)
				}
			})
		}
	}

	// Sample (f): the node built-in runner emits none of the four text pass
	// markers, so the exit-code branch is the only thing that makes it a pass.
	// It gets one payload, not two — without the exit code there is no pass
	// signal to preserve, and that is correct rather than a regression.
	t.Run("node_builtin/with_exit_zero", func(t *testing.T) {
		if detectZeroExecution(sampleNodeTestTAP) {
			t.Fatalf("detectZeroExecution = true on the node TAP pass output")
		}
		isTest, isPass, isFail := classifyTestCommand("npm test", exitZeroPayload(sampleNodeTestTAP))
		if !isTest || !isPass || isFail {
			t.Errorf("got (isTest,isPass,isFail) = (%v,%v,%v), want (true,true,false)", isTest, isPass, isFail)
		}
	})
}

// A `go test ./...` run over a repository containing one package with no test
// files prints `[no test files]` beside dozens of `ok` lines. That run did
// execute tests, so the veto must not swallow its pass — nor a failure
// reported alongside the same line.
func TestZeroExecution_MixedOutputKeepsItsSignal(t *testing.T) {
	t.Parallel()
	mixedPass := sampleGoNoTestFiles + sampleGoPass
	if detectZeroExecution(mixedPass) {
		t.Errorf("detectZeroExecution = true on a suite run that did execute tests")
	}
	if _, isPass, _ := classifyTestCommand("go test ./...", textPayload(mixedPass)); !isPass {
		t.Errorf("isPass = false; a green suite run containing a no-test-files package is still a pass")
	}

	mixedFail := sampleGoNoTestFiles + "--- FAIL: TestX (0.00s)\nFAIL\tgithub.com/x/y\t0.5s\n"
	if detectZeroExecution(mixedFail) {
		t.Errorf("detectZeroExecution = true on a run that reported a failure")
	}
	if _, _, isFail := classifyTestCommand("go test ./...", textPayload(mixedFail)); !isFail {
		t.Errorf("isFail = false; the veto must not suppress an observed failure")
	}
}

// AC-SEC-005 — the zero-execution run surfaces. The advisory rides the
// PostToolUse return payload (not stderr), carries a stable sentinel, records
// the run positively in the ledger, and never blocks.
func TestZeroExecution_SurfacesAsPostToolAdvisory(t *testing.T) {
	root := newMoaiTempRoot(t)
	h := NewPostToolHandler()
	in := bashInput("ac-sec-005", "go test ./internal/config -run TestNope", textPayload(sampleGoNoTestsToRun))
	in.CWD = root

	out, err := h.Handle(t.Context(), in)
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if out == nil {
		t.Fatal("Handle returned nil output")
	}
	if !strings.Contains(out.SystemMessage, zeroExecutionSentinel) {
		t.Errorf("SystemMessage = %q; want it to carry the %q sentinel", out.SystemMessage, zeroExecutionSentinel)
	}
	if out.Decision != "" {
		t.Errorf("Decision = %q; the advisory must not block", out.Decision)
	}

	recs, err := telemetry.LoadBySession(root, "ac-sec-005")
	if err != nil {
		t.Fatalf("LoadBySession: %v", err)
	}
	if len(recs) != 1 {
		t.Fatalf("got %d records, want 1", len(recs))
	}
	if !recs[0].IsZeroExecution {
		t.Errorf("IsZeroExecution = false; the run must be recorded positively, not merely left blank")
	}
	if recs[0].IsTestPass {
		t.Errorf("IsTestPass = true; want false")
	}
}

// The advisory appends; it never replaces an existing systemMessage.
func TestZeroExecution_AdvisoryAppendsToExistingMessage(t *testing.T) {
	t.Parallel()
	got := appendZeroExecutionAdvisory("prior message", "go test ./x")
	if !strings.Contains(got, "prior message") {
		t.Errorf("advisory dropped the existing message: %q", got)
	}
	if !strings.Contains(got, zeroExecutionSentinel) {
		t.Errorf("advisory missing sentinel: %q", got)
	}
}
