package web

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

// Card t1051 — SPEC-WEB-CONSOLE-017 run-phase instruments.
//
// WHAT THIS OBSERVES. A settings save that fails at a persistence seam must
// become observable on two independent surfaces (never merged into one):
//   - REQ-A (user): the failure phrase reaches the existing inline
//     save-failure slot (save__msg save__msg--error, role=alert, rendered by
//     the save cluster in shell.templ) through an htmx-swappable (2xx)
//     response — no full-page navigation required.
//   - REQ-B (maintainer): exactly one stderr line, prefixed `moai web: `,
//     naming the failed seam/layer, carrying no raw error value.
//
// The discriminator for the inline surface is the PHRASE BODY TEXT
// (spec.md §4 HARD-4) — never a CSS class (banner banner--warn is rendered
// pre-submit by fieldsets.templ, so class assertions pass vacuously) and
// never an err.Error() value. The harness reuses recordingSeams from
// partial_apply_repro_test.go (SPEC-WEB-CONSOLE-016 REQ-WC16-010 named this
// card as the dependent sibling): its seam list and signatures are extended,
// never rewritten (HARD-2).

// slotCarriesPhrase reports whether phrase appears inside the inline
// save-failure slot region of body — between the save__msg--error element and
// the save cluster's trailing submit button. The region bound keeps the
// assertion from being satisfied by the same phrase rendered as the page
// banner above the save cluster: the phrase must be inside the slot, not
// merely somewhere on the page.
func slotCarriesPhrase(body, phrase string) bool {
	slot := strings.Index(body, "save__msg--error")
	if slot < 0 {
		return false
	}
	at := strings.Index(body[slot:], phrase)
	if at < 0 {
		return false
	}
	end := strings.Index(body[slot:], "</button>")
	return end < 0 || at < end
}

// TestSaveFailureReasonReachesInlineSlot is AC-WC17-001: a save that fails at
// a persistence seam must deliver the seam's failure phrase to the inline
// slot through a 2xx (htmx-swappable) response, and the pristine page must
// carry no failure phrase before any submission (two-directional guard —
// absence-only would pass vacuously, presence-only would be a blind
// pattern, acceptance.md §D.2).
//
// RED today: the failure response is a 500 full page (handlers.go
// renderErrorPage) and the form is hx-boosted (root.templ), so htmx discards
// the non-2xx body — the phrase never reaches the slot.
func TestSaveFailureReasonReachesInlineSlot(t *testing.T) {
	a := newTestApp(t)
	var calls []saveStep
	recordingSeams(a, &calls, stepWritePreferences)
	h := a.routes()

	// Direction (a) — pre-submit absence.
	pre := serveGet(t, h, "/settings")
	if pre.Code != http.StatusOK {
		t.Fatalf("GET /settings = %d, want 200", pre.Code)
	}
	if got := pre.Body.String(); strings.Contains(got, "could not save profile preferences") {
		t.Errorf("pre-submit page already carries the seam-failure phrase:\n%s", got)
	}

	// Direction (b) — post-failure presence through an htmx-swappable
	// response.
	rec := servePost(t, h, "/save", reproForm())
	if rec.Code != http.StatusOK {
		t.Fatalf("failed-save status = %d, want 200 (htmx discards non-2xx boosted bodies, so the phrase never reaches the slot)", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "could not save profile preferences") {
		t.Fatalf("seam-failure phrase missing from the failure response entirely:\n%s", body)
	}
	if !slotCarriesPhrase(body, "could not save profile preferences") {
		t.Errorf("phrase present on the page but not inside the save__msg--error slot; body:\n%s", body)
	}
}

// captureStderr swaps os.Stderr for a pipe so the test can read what the
// production code writes through the established fmt.Fprintf(os.Stderr, ...)
// idiom. The returned func restores stderr and yields the captured bytes.
// Tests using it MUST NOT call t.Parallel(): os.Stderr is process-global.
func captureStderr(t *testing.T) func() string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stderr
	os.Stderr = w
	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()
	return func() string {
		os.Stderr = old
		_ = w.Close()
		captured := <-done
		_ = r.Close()
		return captured
	}
}

// saveFailureLines extracts the save-failure stderr lines from captured
// output: the lines the SPEC's prefix contract owns. Other stderr traffic
// (config-load WARN logs etc.) is not a save-failure line and is excluded
// from the exactly-one count — counting the whole stream would conflate the
// REQ's subject with ambient process logging.
func saveFailureLines(captured string) []string {
	var out []string
	for _, line := range strings.Split(captured, "\n") {
		if strings.HasPrefix(line, "moai web: ") {
			out = append(out, line)
		}
	}
	return out
}

// TestSaveFailureStderrLog is AC-WC17-003: a failed save appends EXACTLY one
// stderr line, prefixed `moai web: ` (REQ-WC-017-004), naming the failed
// seam (REQ-WC-017-003), carrying no raw error value (REQ-WC-017-005); a
// successful save appends zero (REQ-WC-017-006).
//
// RED today: no save path writes to stderr at all — the only os.Stderr
// writes in this package are the non-save sites (server.go file-watch and
// browser-open), so the failure count is 0, not 1.
func TestSaveFailureStderrLog(t *testing.T) {
	t.Run("failure emits exactly one prefixed line", func(t *testing.T) {
		a := newTestApp(t)
		var calls []saveStep
		recordingSeams(a, &calls, stepWritePreferences)
		h := a.routes()

		restore := captureStderr(t)
		rec := servePost(t, h, "/save", reproForm())
		lines := saveFailureLines(restore())

		if rec.Code != http.StatusOK {
			t.Fatalf("failed-save status = %d, want 200", rec.Code)
		}
		if len(lines) != 1 {
			t.Fatalf("save-failure stderr lines = %d, want exactly 1 (0 undercounts; 2+ violates REQ-WC-017-003's exactly-one); captured:\n%s",
				len(lines), strings.Join(lines, "\n"))
		}
		if !strings.HasPrefix(lines[0], "moai web: ") {
			t.Errorf("line %q does not start with the single `moai web: ` prefix", lines[0])
		}
		if !strings.Contains(lines[0], "writePreferences") {
			t.Errorf("line %q does not name the failed seam", lines[0])
		}
		// REQ-WC-017-005: the raw error value never reaches the line. The
		// injected failure's message text is the value here — its presence
		// would mean err.Error() leaked into the log.
		if strings.Contains(lines[0], "injected failure") {
			t.Errorf("stderr line carries the raw error value: %q", lines[0])
		}
	})

	t.Run("success emits zero lines", func(t *testing.T) {
		a := newTestApp(t)
		var calls []saveStep
		recordingSeams(a, &calls, "")
		h := a.routes()

		restore := captureStderr(t)
		rec := servePost(t, h, "/save", reproForm())
		lines := saveFailureLines(restore())

		if rec.Code != http.StatusOK {
			t.Fatalf("clean-save status = %d, want 200", rec.Code)
		}
		if len(lines) != 0 {
			t.Errorf("successful save emitted %d save-failure stderr lines, want 0: %v", len(lines), lines)
		}
	})
}

// seamTable is spec.md §5.1 as a test fixture: every persistence seam with
// its stderr layer token and its stable inline failure phrase. AC-WC17-002
// drives the save through EACH row and demands both surfaces name it — no
// seam may be skipped. If SPEC-WEB-CONSOLE-016 rewords the banner phrases,
// update the phrase column in the SAME commit (spec.md §6-3).
var seamTable = []struct {
	seam   saveStep
	token  string
	phrase string
}{
	{stepWritePreferences, "writePreferences", "could not save profile preferences"},
	{stepSyncToProject, "syncToProject", "profile preferences saved, but project config sync failed"},
	{stepWriteProjectCfg, "writeProjectConfig", "profile preferences saved, but project config write failed"},
	{stepWriteNested, "writeProjectNestedConfig", "profile preferences saved, but project nested config write failed"},
	{stepApplySchema, "applySchemaEdits", "profile preferences saved, but section config write failed"},
	{stepApplyPerfTier, "applyPerfTierEdits", "profile preferences saved, but performance_tier apply failed"},
	{stepPatchAgentFM, "patchAgentFM", "settings saved, but agent override write failed"},
	{stepGlmcredSave, "glmcred.Save", "settings saved, but GLM credential write failed"},
	{stepJevcredSave, "jevcred.Save", "settings saved, but Jev credential write failed"},
}

// TestSaveFailureSeamCoverage is AC-WC17-002: for EVERY one of the nine
// persistence seams, a forced failure names that seam on BOTH surfaces —
// the inline slot (REQ-A) and the stderr log (REQ-B). The pre-submit
// absence assertion rides along per seam so the two-directional guard of
// AC-WC17-001 holds row-wise, not just for the one seam its own test fails.
func TestSaveFailureSeamCoverage(t *testing.T) {
	for _, tc := range seamTable {
		t.Run(tc.token, func(t *testing.T) {
			a := newTestApp(t)
			var calls []saveStep
			recordingSeams(a, &calls, tc.seam)
			h := a.routes()

			// Pre-submit absence, per seam.
			pre := serveGet(t, h, "/settings")
			if pre.Code != http.StatusOK {
				t.Fatalf("GET /settings = %d, want 200", pre.Code)
			}
			if strings.Contains(pre.Body.String(), tc.phrase) {
				t.Errorf("pre-submit page already carries the %s failure phrase", tc.token)
			}

			// Both surfaces observed around the same submission: the stderr
			// line is written while handleSave runs, the inline phrase is in
			// the response it produces.
			restore := captureStderr(t)
			rec := servePost(t, h, "/save", reproForm())
			lines := saveFailureLines(restore())

			// The harness must have failed exactly the target seam (its last
			// call) — otherwise a pass here measures the wrong row.
			if len(calls) == 0 || calls[len(calls)-1] != tc.seam {
				t.Fatalf("last seam reached = %v, want %q", calls, tc.seam)
			}

			// REQ-A: the phrase sits inside the inline slot of a 2xx response.
			if rec.Code != http.StatusOK {
				t.Fatalf("failed-save status = %d, want 200", rec.Code)
			}
			if !slotCarriesPhrase(rec.Body.String(), tc.phrase) {
				t.Errorf("%s failure phrase missing from the inline slot", tc.token)
			}

			// REQ-B: exactly one stderr line, prefixed, naming the seam.
			if len(lines) != 1 {
				t.Fatalf("%s failure stderr lines = %d, want exactly 1; captured:\n%s",
					tc.token, len(lines), strings.Join(lines, "\n"))
			}
			if !strings.HasPrefix(lines[0], "moai web: ") {
				t.Errorf("line %q lacks the `moai web: ` prefix", lines[0])
			}
			if !strings.Contains(lines[0], tc.token) {
				t.Errorf("stderr line %q does not name the seam %q", lines[0], tc.token)
			}
		})
	}
}

// t1051SentinelKey is a value standing in for real GLM/Jev credential material in
// the non-leak tests (AC-WC17-004). Newline-free so key validation accepts it
// and the persistence seams are reached.
const t1051SentinelKey = "T1051-SENTINEL-9f8e7d6c5b4a"

// TestSaveFailureNeverLeaksKeyMaterial is AC-WC17-004 (regression-guard,
// HARD-3 / REQ-WC-017-005; continues SPEC-GLM-KEY-INPUT-001 REQ-GKI-004-003):
// with sentinel key material submitted, a forced failure at any seam must
// leave the sentinel absent from BOTH new observability surfaces — the
// stderr stream and the inline failure response. Classification note
// (acceptance.md §D.1): before AC-WC17-003 landed there was no stderr
// surface to leak into, so this criterion could not be RED-adopted then; it
// became RED-capable once M2 shipped and is exercised here per seam.
func TestSaveFailureNeverLeaksKeyMaterial(t *testing.T) {
	for _, tc := range seamTable {
		t.Run(tc.token, func(t *testing.T) {
			a := newTestApp(t)
			var calls []saveStep
			recordingSeams(a, &calls, tc.seam)
			form := reproForm()
			form.Set("glm_api_key", t1051SentinelKey)
			form.Set("jev_api_key", t1051SentinelKey)

			restore := captureStderr(t)
			rec := servePost(t, a.routes(), "/save", form)
			captured := restore()

			// The forced failure must actually have happened — otherwise the
			// non-leakage assertions below would pass against a success.
			if !strings.Contains(rec.Body.String(), tc.phrase) {
				t.Fatalf("forced failure at %s did not surface its phrase; nothing to assert non-leakage against", tc.token)
			}
			if strings.Contains(rec.Body.String(), t1051SentinelKey) {
				t.Errorf("%s failure leaked sentinel key material into the inline response", tc.token)
			}
			if strings.Contains(captured, t1051SentinelKey) {
				t.Errorf("%s failure leaked sentinel key material into stderr:\n%s", tc.token, captured)
			}
		})
	}
}

// TestSaveSuccessSurfaceUnchanged is AC-WC17-005 (regression-guard,
// REQ-WC-017-006): a clean save keeps the success surface exactly as it was —
// the "Settings saved." banner, the saved inline state, no error slot, no
// seam-failure phrase, zero save-failure stderr lines. It is re-asserted at
// each landing to prove the two new surfaces changed nothing about success.
func TestSaveSuccessSurfaceUnchanged(t *testing.T) {
	a := newTestApp(t)
	var calls []saveStep
	recordingSeams(a, &calls, "")
	form := reproForm()
	form.Set("glm_api_key", t1051SentinelKey)
	form.Set("jev_api_key", t1051SentinelKey)

	restore := captureStderr(t)
	rec := servePost(t, a.routes(), "/save", form)
	captured := restore()

	if rec.Code != http.StatusOK {
		t.Fatalf("clean-save status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Settings saved.") {
		t.Errorf("clean save lost the success banner; body:\n%s", body)
	}
	if !strings.Contains(body, `data-save-state="saved"`) {
		t.Errorf("clean save did not render the saved inline state; body:\n%s", body)
	}
	if strings.Contains(body, "save__msg--error") {
		t.Errorf("clean save rendered an error slot; body:\n%s", body)
	}
	if strings.Contains(body, "could not save profile preferences") {
		t.Errorf("clean save rendered a seam-failure phrase; body:\n%s", body)
	}
	if lines := saveFailureLines(captured); len(lines) != 0 {
		t.Errorf("clean save emitted %d save-failure stderr lines, want 0: %v", len(lines), lines)
	}
}
