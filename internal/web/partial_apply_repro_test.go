package web

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/profile"
)

// Card t1049 — reproduction instrument for the handleSave partial-apply order.
//
// WHAT THIS OBSERVES. handleSave persists through seven write steps in a fixed
// order (handlers.go:465-549 on this tree). Each step returns on error, so a
// failure at step N leaves steps 1..N-1 already on disk. This file injects a
// failure at each INJECTABLE step, records which steps ran, and captures what
// the response shows the user.
//
// WHAT IT DOES NOT OBSERVE. Step order is recorded as seam invocation, not as
// raw bytes on disk: the default seams write to the developer's real profile
// store (~/.moai/claude-profiles), which a test in this repository must not
// touch. Seam invocation is the faithful proxy — a seam that was called is a
// write that was attempted with the real implementation wired.
//
// MEASURED GAP — two of the seven steps carry NO injection seam:
//   - applyPerfTierEdits (handlers.go:523) is a package-level function
//   - glmcred.Save        (handlers.go:546) is a package-level function
//
// Failures at those two steps cannot be provoked here; they need a
// filesystem-level probe (making the target path unwritable). That is a gap,
// not an absence of the hazard.
//
// REUSE NOTE (card t1051): recordingSeams below is the injection harness for
// observing save-failure behaviour generally — a logging-surface probe can wrap
// the same seams without re-deriving them.

// saveStep names one persistence step of handleSave in execution order.
type saveStep string

const (
	stepWritePreferences saveStep = "1 writePreferences"
	stepRecordLast       saveStep = "2 recordLastProfile(advisory)"
	stepSyncToProject    saveStep = "3 syncToProject"
	stepWriteProjectCfg  saveStep = "4 writeProjectConfig"
	stepWriteNested      saveStep = "5 writeProjectNestedConfig"
	stepApplySchema      saveStep = "6 applySchemaEdits"
	stepPatchAgentFM     saveStep = "8 patchAgentFM"
)

// injectableSteps is the execution-ordered list of steps this harness can fail.
// Steps 7 (applyPerfTierEdits) and 9 (glmcred.Save) are absent by measurement,
// not by choice — see the file header.
var injectableSteps = []saveStep{
	stepWritePreferences,
	stepSyncToProject,
	stepWriteProjectCfg,
	stepWriteNested,
	stepApplySchema,
	stepPatchAgentFM,
}

// recordingSeams wires every injectable seam on a to record its invocation into
// *calls, failing at failAt (empty = fail nowhere: the positive control).
func recordingSeams(a *app, calls *[]saveStep, failAt saveStep) {
	record := func(s saveStep) error {
		*calls = append(*calls, s)
		if s == failAt {
			return assertErr(string(s) + " injected failure")
		}
		return nil
	}
	a.writePreferences = func(string, profile.ProfilePreferences) error { return record(stepWritePreferences) }
	a.recordLastProfile = func(string) error { return record(stepRecordLast) }
	a.syncToProject = func(string, profile.ProfilePreferences) error { return record(stepSyncToProject) }
	a.writeProjectConfig = func(string, string, string) error { return record(stepWriteProjectCfg) }
	a.writeProjectNestedConfig = func(string, projectNestedForm) error { return record(stepWriteNested) }
	a.applySchemaEdits = func(string, map[string]string) error { return record(stepApplySchema) }
	a.patchAgentFM = func(string, map[string]config.ModelEffort, []string) error { return record(stepPatchAgentFM) }
}

// reproForm is one valid submission that reaches every persistence step. The
// development_mode value is the probe for the echo question: it is written at
// step 4, so a failure at step 4 means the value never landed on disk.
func reproForm() url.Values {
	return url.Values{
		"__profile":        {"default"},
		"permission_mode":  {"acceptEdits"},
		"user_name":        {"REPRO-USER"},
		"development_mode": {"tdd"},
	}
}

// TestPartialApplyOrderPositiveControl establishes that the instrument fires:
// with no injected failure every injectable step runs, in the documented order,
// and the response is the success banner. Without this row, a later "steps 1..N
// ran" observation could be an artifact of a harness that never ran anything.
func TestPartialApplyOrderPositiveControl(t *testing.T) {
	a := newTestApp(t)
	var calls []saveStep
	recordingSeams(a, &calls, "")

	rec := servePost(t, a.routes(), "/save", reproForm())

	if rec.Code != http.StatusOK {
		t.Fatalf("clean save status = %d, want 200; body:\n%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Settings saved.") {
		t.Error("clean save did not render the success banner")
	}
	// stepRecordLast is absent by MEASUREMENT, not omission: handlers.go:478
	// guards it with `selected != "" && selected != "default"`, and this form
	// submits the default profile. The first run of this control asserted it
	// would fire and was refuted — the guard is the reason.
	want := []saveStep{
		stepWritePreferences, stepSyncToProject,
		stepWriteProjectCfg, stepWriteNested, stepApplySchema, stepPatchAgentFM,
	}
	if len(calls) != len(want) {
		t.Fatalf("clean save ran %d steps %v, want %d %v", len(calls), calls, len(want), want)
	}
	for i := range want {
		if calls[i] != want[i] {
			t.Errorf("step %d = %q, want %q (full order: %v)", i, calls[i], want[i], calls)
		}
	}
	t.Logf("positive control — step order: %v", calls)
}

// TestPartialApplyOrderOnInjectedFailure is the card's reproduction: for each
// injectable step, fail there and observe how far persistence got.
func TestPartialApplyOrderOnInjectedFailure(t *testing.T) {
	for _, failAt := range injectableSteps {
		t.Run(string(failAt), func(t *testing.T) {
			a := newTestApp(t)
			var calls []saveStep
			recordingSeams(a, &calls, failAt)

			rec := servePost(t, a.routes(), "/save", reproForm())

			// SPEC-WEB-CONSOLE-017 REQ-WC-017-001: a failed save answers 200 —
			// the boosted form discards non-2xx bodies, so the failure reason in
			// the inline slot only reaches the browser through a 2xx re-render.
			if rec.Code != http.StatusOK {
				t.Errorf("status = %d, want 200", rec.Code)
			}
			// The failing step must be the last one reached: every later step
			// is skipped by the early return, every earlier step already ran.
			if len(calls) == 0 || calls[len(calls)-1] != failAt {
				t.Fatalf("last step reached = %v, want %q", calls, failAt)
			}
			committed := calls[:len(calls)-1]
			t.Logf("failure at %q → steps already persisted: %v", failAt, committed)

			body := rec.Body.String()
			// Observation 1: the banner names the failing step only. It never
			// enumerates which of the earlier steps did land.
			for _, s := range committed {
				if strings.Contains(body, string(s)) {
					t.Errorf("banner unexpectedly names committed step %q", s)
				}
			}
			// Observation 2: the re-rendered form echoes the SUBMITTED value,
			// not the on-disk value. development_mode is written at step 4, so
			// when step 4 or anything before it failed, the page still shows
			// "tdd" selected although no write of it ever landed.
			echoesUnwritten := strings.Contains(body, `value="REPRO-USER"`)
			t.Logf("failure at %q → form echoes submitted user_name: %v", failAt, echoesUnwritten)
		})
	}
}
