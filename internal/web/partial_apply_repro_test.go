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
// MEASURED GAP — closed by card t1051 (SPEC-WEB-CONSOLE-017). Three steps
// previously carried NO injection seam because they were package-level
// calls: applyPerfTierEdits, glmcred.Save, jevcred.Save. SPEC-WEB-CONSOLE-017
// HARD-2 promoted the CALLS (not the implementations) to app fields and this
// harness was extended — never rewritten — to inject all of them, so the
// gap note above is historical. jevcred.Save was also missing from the
// original seven-step inventory.
//
// REUSE NOTE (card t1051): recordingSeams below is the injection harness for
// observing save-failure behaviour generally — the logging-surface probe in
// save_observability_test.go drives it for every one of the nine seams.

// saveStep names one persistence step of handleSave in execution order.
type saveStep string

const (
	stepWritePreferences saveStep = "1 writePreferences"
	stepRecordLast       saveStep = "2 recordLastProfile(advisory)"
	stepSyncToProject    saveStep = "3 syncToProject"
	stepWriteProjectCfg  saveStep = "4 writeProjectConfig"
	stepWriteNested      saveStep = "5 writeProjectNestedConfig"
	stepApplySchema      saveStep = "6 applySchemaEdits"
	stepApplyPerfTier    saveStep = "7 applyPerfTierEdits"
	stepPatchAgentFM     saveStep = "8 patchAgentFM"
	stepGlmcredSave      saveStep = "9 glmcred.Save"
	stepJevcredSave      saveStep = "10 jevcred.Save"
)

// injectableSteps is the execution-ordered list of steps this harness can
// fail. All nine persistence seams are injectable as of card t1051 — the
// three former package-level calls were promoted to app fields
// (SPEC-WEB-CONSOLE-017 HARD-2). The original seven-entry numbering is
// preserved verbatim; the new constants fill the gaps in the inventory.
var injectableSteps = []saveStep{
	stepWritePreferences,
	stepSyncToProject,
	stepWriteProjectCfg,
	stepWriteNested,
	stepApplySchema,
	stepApplyPerfTier,
	stepPatchAgentFM,
	stepGlmcredSave,
	stepJevcredSave,
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
	// Card t1051 (SPEC-WEB-CONSOLE-017 HARD-2): the three former package-level
	// calls, now app fields — same recorder, no signature change anywhere.
	a.applyPerfTierEdits = func(string, string) error { return record(stepApplyPerfTier) }
	a.glmcredSave = func(string) error { return record(stepGlmcredSave) }
	a.jevcredSave = func(string) error { return record(stepJevcredSave) }
}

// reproForm is one valid submission that reaches every persistence step. The
// development_mode value is the probe for the echo question: it is written at
// step 4, so a failure at step 4 means the value never landed on disk.
// Card t1051: the two credential keys are submitted (newline-free, so
// validation passes) so the form reaches glmcred.Save and jevcred.Save as
// well — recordingSeams always wires those, so the real credential writers
// are never reached by a test using this fixture.
func reproForm() url.Values {
	return url.Values{
		"__profile":        {"default"},
		"permission_mode":  {"acceptEdits"},
		"user_name":        {"REPRO-USER"},
		"development_mode": {"tdd"},
		"glm_api_key":      {"T1051-TEST-KEY"},
		"jev_api_key":      {"T1051-TEST-KEY"},
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
		stepWriteProjectCfg, stepWriteNested, stepApplySchema,
		stepApplyPerfTier, stepPatchAgentFM,
		stepGlmcredSave, stepJevcredSave,
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
