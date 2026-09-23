package web

import (
	"regexp"
	"strings"
	"testing"
)

// Card t1105 — the composite contract t1080 measured in two halves but never
// joined: the validation-reject response must carry a status the pinned,
// embedded htmx build actually swaps.
//
// t1080 measured the two halves separately and stopped there, by design:
//   - D-400a (transport400_characterization_test.go): the 400 body DOES carry
//     the banner and the per-field errors.
//   - D-400b (transport400_htmx_contract_test.go): the embedded htmx 2.0.4
//     default responseHandling table answers status 400 with swap:false.
//
// Neither half is a defect on its own. Joined, they are: a boosted form
// (root.templ hx-boost="true") posts to /save, the handler answers 400, htmx
// declines to swap it, and the feedback the server took the trouble to render
// never reaches the screen. The fix axis is the one SPEC-WEB-CONSOLE-017
// already took for the save-failure seam (handlers.go renderErrorPage answers
// 200, not 500, precisely so the boosted swap delivers the reason).
//
// This test joins the halves rather than asserting a literal status: it reads
// whatever status the handler actually returns and evaluates it against the
// table extracted from the committed asset. That keeps it honest if either
// side changes — a future htmx upgrade that starts swapping 4xx, or a handler
// that moves to a different 2xx, both keep this green for the right reason.

// extractHtmxResponseHandling parses the default responseHandling table out of
// the committed embedded htmx asset. It returns nil when the minified shape
// cannot be parsed, so the caller can record "not measured" rather than infer
// a verdict from an extraction failure.
func extractHtmxResponseHandling(t *testing.T) []htmxResponseEntry {
	t.Helper()

	js := readEmbeddedAsset(t, "htmx.min.js")
	tableRe := regexp.MustCompile(`responseHandling:(\[\{.*?\}\])`)
	m := tableRe.FindStringSubmatch(js)
	if m == nil {
		return nil
	}
	entryRe := regexp.MustCompile(`\{code:"([^"]+)",swap:(true|false)(,error:true)?\}`)
	matches := entryRe.FindAllStringSubmatch(m[1], -1)
	if len(matches) == 0 {
		return nil
	}
	entries := make([]htmxResponseEntry, 0, len(matches))
	for _, em := range matches {
		entries = append(entries, htmxResponseEntry{
			Code:  em[1],
			Swap:  em[2] == "true",
			Error: em[3] != "",
		})
	}
	return entries
}

// htmxSwapsStatus reports whether the pinned htmx build swaps a response with
// this status, replicating the asset's first-match-wins loop. The second
// return is false when no entry matches at all — a distinct outcome from
// "matched an entry that declines the swap".
func htmxSwapsStatus(entries []htmxResponseEntry, status string) (swap bool, matched bool) {
	for _, e := range entries {
		if htmxResponseHandlingMatches(e, status) {
			return e.Swap, true
		}
	}
	return false, false
}

// validationRejectBanner is the phrase handleSave renders into the page when a
// submission fails validation. It is the SAME literal the handler writes; the
// tests assert against it rather than against the status code because, since
// card t1105, the status no longer distinguishes a reject from a success.
const validationRejectBanner = "Validation failed — no changes were saved."

// assertValidationRejectBanner is the replacement signal for the status code
// the t1105 fix gave up.
//
// Before t1105 the validation-reject path answered 400, and ~17 assertions
// across this package read that status as "the submission was rejected". The
// fix moves the path to 2xx so the boosted htmx swap actually delivers the
// rendered feedback — which means the status alone no longer separates reject
// from accept. Those call sites keep a machine-readable reject signal by
// asserting the banner instead: it is rendered on the reject path and on no
// other, so it carries exactly what the status used to.
//
// Every such call site ALSO asserts its own atomic-reject condition (no write
// seam fired, the on-disk file is byte-unchanged). That pair — banner present
// AND nothing persisted — is a strictly stronger contract than the status
// check it replaces, which never observed the body at all.
func assertValidationRejectBanner(t *testing.T, body string) {
	t.Helper()
	if !strings.Contains(body, validationRejectBanner) {
		t.Errorf("validation-reject response does not carry the banner %q — since card t1105 the status is 2xx, so the banner is the reject signal; its absence means the response is indistinguishable from a successful save", validationRejectBanner)
	}
}

// TestValidationRejectStatusIsHtmxSwappable is the t1105 RED: it fails on the
// pre-fix tree because handleSave answers the validation-reject path with 400
// and the embedded htmx table answers 400 with swap:false.
func TestValidationRejectStatusIsHtmxSwappable(t *testing.T) {
	t.Parallel()

	entries := extractHtmxResponseHandling(t)
	if entries == nil {
		t.Skip("deferred-to-browser: responseHandling table not parseable from the embedded htmx asset — the swap contract is unmeasurable on this tree; this skip records \"not measured\", not a pass")
	}

	a := newTestApp(t)
	// Atomic-reject guard: a write here would mean the request never traversed
	// the validation-reject path, so the status measured below would be the
	// wrong one.
	writeCalled := false
	a.writeProjectConfig = func(string, string, string) error {
		writeCalled = true
		return nil
	}

	form := projectSaveForm("", "angular")
	form.Set("permission_mode", "bogus")
	rec := servePost(t, a.routes(), "/save", form)

	if writeCalled {
		t.Fatal("write seam invoked on validation failure — request did not traverse the atomic-reject path")
	}

	status := rec.Code

	swap, matched := htmxSwapsStatus(entries, itoa(status))
	if !matched {
		t.Fatalf("no responseHandling entry matches status %d — the pinned asset has no default for it, so the swap outcome is undetermined; record this rather than reading it as a pass", status)
	}
	if !swap {
		t.Errorf("validation-reject answers status %d, which the pinned htmx build handles with swap:false — the rendered banner and per-field errors never reach the screen under hx-boost", status)
	}

	// Feedback must still be IN the body: a status the client swaps is only
	// half the contract. Guards the degenerate fix of returning 200 with an
	// empty or success-looking page.
	body := rec.Body.String()
	const banner = "Validation failed — no changes were saved."
	if !strings.Contains(body, banner) {
		t.Errorf("swappable response does not carry the banner %q — the status was fixed without the feedback it is supposed to deliver", banner)
	}
	const fieldErr = "unrecognized permission mode: bogus"
	if !strings.Contains(body, fieldErr) {
		t.Errorf("swappable response does not carry the per-field error %q", fieldErr)
	}

	if swap && strings.Contains(body, banner) && strings.Contains(body, fieldErr) {
		t.Logf("t1105 contract MEASURED: status %d, htmx swap:true, banner and per-field error both present (%d bytes)", status, len(body))
	}
}
