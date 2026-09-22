package web

import (
	"net/http"
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
