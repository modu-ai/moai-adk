package web

import (
	"net/http"
	"strings"
	"testing"
)

// SPEC-WEB-TRANSPORT-001 M2 (REQ-TR400-001, AC-TR400-001) — D-400a server-side
// characterization. httptest-submits a validation-failing POST /save and
// measures what the 400 response body actually carries: the error banner and
// at least one per-field error message.
//
// This is a CHARACTERIZATION test: it pins the behavior the validation-reject
// path exhibits today (handlers.go renders the full page at status 400 with
// the banner + merged field errors). It makes no judgment about whether htmx
// delivers this body under hx-boost — that is D-400b (M1) and the browser
// axis (card t1081). If a remediation card later changes the 400 path, this
// test converts to a specification test at that time.
//
// FAIL classification (REQ-TR400-003 truth table): a body-without-feedback
// observation here is the measurement-invalid (server-contract divergence)
// grade — re-measure, escalate; it is never evidence of defect-absence.

// TestCharacterizeValidation400BodyFeedback characterizes the 400 body of the
// validation-reject path: status 400 AND the banner phrase AND at least one
// per-field error marker present in the response body.
func TestCharacterizeValidation400BodyFeedback(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	// Atomic-reject guard: no write seam may fire on a validation failure —
	// a write here would mean the request did not traverse the reject path.
	writeCalled := false
	a.writeProjectConfig = func(string, string, string) error {
		writeCalled = true
		return nil
	}

	// permission_mode "" is submitted as "bogus" -> exactly one failing
	// validator (validatePrefs), yielding the per-field error
	// "unrecognized permission mode: bogus". permission_mode is chosen
	// because its field error RENDERS in the page (fieldsets fieldErr); the
	// development_mode field error exists in the view but its project render
	// surface was retired (SPEC-DESIGN-MOAIWEBV2-001 M1), so it never reaches
	// the response body — measured here on this tree (first probe attempt).
	form := projectSaveForm("", "angular")
	form.Set("permission_mode", "bogus")
	rec := servePost(t, a.routes(), "/save", form)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("validation-reject status = %d, want 400 (body: %.400s)", rec.Code, rec.Body.String())
	}
	if writeCalled {
		t.Fatal("write seam invoked on validation failure — request did not traverse the atomic-reject path")
	}

	body := rec.Body.String()

	const banner = "Validation failed — no changes were saved."
	if !strings.Contains(body, banner) {
		t.Errorf("400 body does not carry the banner phrase %q (D-400a divergence candidate — re-measure per REQ-TR400-003 before any judgment)", banner)
	}

	const fieldErr = "unrecognized permission mode: bogus"
	if !strings.Contains(body, fieldErr) {
		t.Errorf("400 body does not carry the per-field error marker %q (D-400a divergence candidate — re-measure per REQ-TR400-003 before any judgment)", fieldErr)
	}

	if strings.Contains(body, banner) && strings.Contains(body, fieldErr) {
		t.Logf("D-400a MEASURED: 400 body carries banner %q AND per-field error %q (%d bytes)", banner, fieldErr, len(body))
	}
}
