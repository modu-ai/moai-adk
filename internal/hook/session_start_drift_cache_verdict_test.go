package hook

// session_start_drift_cache_verdict_test.go — AC-DCF-014 / REQ-DCF-016: the
// drift-cache probe carries a VERDICT, an empty sweep is not a pass, and the
// probe's own absence is visible.
//
// The defect being repaired is inside the instrument this card relies on. As
// shipped, TestSessionStart_DriftCacheProbe measures, prints, and t.Skips
// without its environment variables — so an ordinary `go test ./internal/hook/`
// prints `ok` whether the probe is healthy, broken, or deleted. That is the
// continued-firing defect of verification-completeness.md §1.3, in the very
// tool whose output the card's field measurement depends on: a check whose
// non-execution is indistinguishable from its success.
//
// Three things close it, and the first two run on EVERY package run:
//
//  1. the classifier below, extracted so the verdict is a VALUE that can be
//     asserted rather than a line that can be printed;
//  2. the liveness assertion, which reads the probe's own source and fails when
//     the probe function or its env gate is renamed or deleted;
//  3. the probe's asserting mode (MOAI_DRIFT_CACHE_PROBE_ASSERT), which turns a
//     non-pass verdict into t.Fatalf instead of a log line — the
//     report-not-verdict repair.

import (
	"os"
	"strings"
	"testing"
)

// driftProbeVerdict is the probe's answer about a swept set of runs.
type driftProbeVerdict string

const (
	// driftProbeFail — at least one run in the swept set left the cache absent.
	driftProbeFail driftProbeVerdict = "FAIL"
	// driftProbePass — every run in a NON-EMPTY swept set saw a HEAD-matching
	// cache.
	driftProbePass driftProbeVerdict = "PASS"
	// driftProbeEmpty — nothing was swept. This is deliberately DISTINCT from
	// PASS: a verification that selected nothing still reports success, and the
	// report is indistinguishable from one where everything passed.
	driftProbeEmpty driftProbeVerdict = "EMPTY"
	// driftProbeInconclusive — the set contains a state that is neither an
	// absence nor a HEAD match (a stale or corrupt cache). Not a pass.
	driftProbeInconclusive driftProbeVerdict = "INCONCLUSIVE"
)

// classifyDriftCacheProbeRuns reduces the probe's per-run cache states to one
// verdict. States are the strings the probe's own cacheState helper produces:
// "absent", "present(head-match)", "present(stale)", "corrupt".
//
// The EMPTY judgement is made AHEAD of every other signal, because the exit
// code of a run that swept nothing is the same zero a fully-passing run returns.
func classifyDriftCacheProbeRuns(states []string) (driftProbeVerdict, int) {
	if len(states) == 0 {
		return driftProbeEmpty, 0
	}
	sawOther := false
	for _, s := range states {
		switch s {
		case "absent":
			return driftProbeFail, len(states)
		case "present(head-match)":
		default:
			sawOther = true
		}
	}
	if sawOther {
		return driftProbeInconclusive, len(states)
	}
	return driftProbePass, len(states)
}

// TestClassifyDriftCacheProbeRuns is AC-DCF-014(a). It runs on EVERY package
// run, unlike the probe it classifies for.
func TestClassifyDriftCacheProbeRuns(t *testing.T) {
	cases := []struct {
		name   string
		states []string
		want   driftProbeVerdict
		swept  int
	}{
		{name: "one absent among hits", states: []string{"present(head-match)", "absent", "present(head-match)"}, want: driftProbeFail, swept: 3},
		{name: "all absent", states: []string{"absent", "absent"}, want: driftProbeFail, swept: 2},
		{name: "all head-match", states: []string{"present(head-match)", "present(head-match)"}, want: driftProbePass, swept: 2},
		{name: "empty set is NOT a pass", states: nil, want: driftProbeEmpty, swept: 0},
		{name: "stale is not a pass", states: []string{"present(head-match)", "present(stale)"}, want: driftProbeInconclusive, swept: 2},
		{name: "corrupt is not a pass", states: []string{"corrupt"}, want: driftProbeInconclusive, swept: 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, swept := classifyDriftCacheProbeRuns(tc.states)
			if got != tc.want {
				t.Errorf("verdict = %q, want %q", got, tc.want)
			}
			if swept != tc.swept {
				t.Errorf("swept count = %d, want %d", swept, tc.swept)
			}
		})
	}
	// The distinction the criterion exists for, stated as its own assertion so
	// a future edit collapsing EMPTY into PASS cannot pass silently.
	if driftProbeEmpty == driftProbePass {
		t.Fatal("an empty sweep and a pass are the same verdict — an empty sweep asserts nothing")
	}
}

// TestDriftCacheProbeIsLive is the §1.3 continued-firing carrier for this
// SPEC's on-demand check. The probe itself runs only when its environment
// variables are set, so nothing in an ordinary package run would notice its
// deletion. This test notices: it reads the probe's source and asserts the
// function and its env gates are still there, and that the asserting mode acts
// on the verdict rather than printing it.
//
// Deleting or renaming the probe turns the ordinary
// `go test ./internal/hook/...` red instead of silently removing a check
// nobody runs.
func TestDriftCacheProbeIsLive(t *testing.T) {
	const probeFile = "session_start_drift_cache_probe_test.go"
	raw, err := os.ReadFile(probeFile)
	if err != nil {
		t.Fatalf("the drift-cache probe file is gone: %v", err)
	}
	body := string(raw)

	for _, want := range []string{
		"func TestSessionStart_DriftCacheProbe(",
		"MOAI_DRIFT_CACHE_PROBE_BIN",
		"MOAI_DRIFT_CACHE_PROBE_SRC",
		"MOAI_DRIFT_CACHE_PROBE_ASSERT",
		"classifyDriftCacheProbeRuns(",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("%s no longer contains %q — the probe or its env gate was renamed or removed", probeFile, want)
		}
	}

	// report-not-verdict: the asserting mode must ACT on the verdict.
	if !strings.Contains(body, "t.Fatalf") {
		t.Error("the probe never fails on its own verdict — it prints findings and exits with the status of whatever ran last")
	}
	// And the swept count must be printed beside the verdict, so a green with
	// an empty sweep is visible rather than silent.
	if !strings.Contains(body, "swept=") {
		t.Error("the probe does not report its swept run count; an empty sweep would be indistinguishable from a pass")
	}
}
