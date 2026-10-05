package web

import (
	"net/http"
	"strings"
	"testing"
)

// SPEC-WEB-TRANSPORT-001 M2 (REQ-TR400-001, AC-TR400-001) — the D-400a body
// contract of the validation-reject path: a validation-failing POST /save
// renders the error banner and at least one per-field error message.
//
// This began as a CHARACTERIZATION test, pinning what the path did at the time
// (a full page rendered at status 400) while explicitly withholding judgment
// on whether htmx delivered that body under hx-boost. Its own closing note
// said: "If a remediation card later changes the 400 path, this test converts
// to a specification test at that time." Card t1105 is that card — it measured
// the join (the pinned htmx build answers 4xx with swap:false, so the body was
// rendered and then discarded) and moved the path to 2xx.
//
// So this is now a SPECIFICATION test. What it asserts is unchanged in
// substance — banner and per-field error present, nothing persisted — but the
// status it expects is the swappable one, and a failure here is a regression
// rather than a re-measurement. The swap side of the contract is asserted
// separately in transport400_swap_contract_test.go.

// TestCharacterizeValidation400BodyFeedback specifies the body of the
// validation-reject path: a swappable status AND the banner phrase AND at
// least one per-field error marker present in the response body.
//
// The name keeps its original spelling so the SPEC-WEB-TRANSPORT-001 evidence
// trail (t1080's verdict cites this test by name) still resolves.
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

	if rec.Code != http.StatusOK {
		t.Fatalf("validation-reject status = %d, want 200 (body: %.400s)", rec.Code, rec.Body.String())
	}
	if writeCalled {
		t.Fatal("write seam invoked on validation failure — request did not traverse the atomic-reject path")
	}

	body := rec.Body.String()

	const banner = "Validation failed — no changes were saved."
	if !strings.Contains(body, banner) {
		t.Errorf("validation-reject body does not carry the banner phrase %q — regression: the reject path renders no visible reason (card t1105)", banner)
	}

	const fieldErr = "unrecognized permission mode: bogus"
	if !strings.Contains(body, fieldErr) {
		t.Errorf("validation-reject body does not carry the per-field error marker %q — regression: the field that failed is not identified to the user (card t1105)", fieldErr)
	}

	if strings.Contains(body, banner) && strings.Contains(body, fieldErr) {
		t.Logf("D-400a SPECIFIED: validation-reject body carries banner %q AND per-field error %q (%d bytes)", banner, fieldErr, len(body))
	}
}
