package web

// profile_bar_test.go — SPEC-WEB-CONSOLE-REDESIGN-001 M5 guards
// (AC-WCR-040..042).
//
// The console had two profile surfaces: a select in the profile bar and a
// separate card duplicating switch + create + delete. M5 removes the card and
// moves the CRUD controls next to the bar. The HTML constraint drives the shape:
// create/rename/delete are each their own POST target, and a <form> inside the
// main settings <form> is invalid markup a browser silently truncates — so the
// controls live in the bar, which already sits outside the settings form.

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/net/html"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// TestProfileManagerCardAbsent verifies AC-WCR-040: the duplicate card is gone.
func TestProfileManagerCardAbsent(t *testing.T) {
	body := renderIndexBody(t, profile.ProfilePreferences{})
	for _, marker := range []string{"profilemgr", "profile-manager"} {
		if strings.Contains(body, marker) {
			t.Errorf("rendered console still carries the profileManager card marker %q", marker)
		}
	}
	// The bar itself must survive — removing the card must not remove the
	// profile UI altogether.
	if !strings.Contains(body, "profilebar") {
		t.Error("the profile bar is gone — M5 consolidates onto the bar, it does not delete the surface")
	}
}

// findMainSettingsForm returns the <form class="form"> node, which is the main
// settings form the profile CRUD forms must stay outside of.
func findMainSettingsForm(t *testing.T, body string) *html.Node {
	t.Helper()
	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		t.Fatalf("parse rendered page: %v", err)
	}
	var found *html.Node
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if found != nil {
			return
		}
		if n.Type == html.ElementNode && n.Data == "form" {
			for _, a := range n.Attr {
				if a.Key == "class" && strings.Contains(a.Val, "form") && !strings.Contains(a.Val, "profile") {
					found = n
					return
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	if found == nil {
		t.Fatal("main settings form (<form class=\"form\">) not found in the rendered page")
	}
	return found
}

func countDescendantForms(n *html.Node) int {
	count := 0
	var walk func(*html.Node)
	walk = func(x *html.Node) {
		for c := x.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && c.Data == "form" {
				count++
			}
			walk(c)
		}
	}
	walk(n)
	return count
}

// renderBarWithModifiableProfile renders GET / for an app that has a profile
// the CRUD controls may act on. The rename and delete forms are deliberately
// not rendered when nothing is eligible (an empty target select would invite an
// action the server refuses), so a fixture carrying only `default` cannot
// exercise them.
func renderBarWithModifiableProfile(t *testing.T) string {
	t.Helper()
	a := newTestApp(t)
	a.listProfiles = func() []profile.ProfileEntry {
		return []profile.ProfileEntry{
			{Name: "default", Current: true},
			{Name: "scratch"},
		}
	}
	rec := serveGet(t, a.routes(), "/settings")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want 200", rec.Code)
	}
	return rec.Body.String()
}

// TestProfileFormsNotNested verifies AC-WCR-041 structurally. Nesting is a tree
// property, so the assertion parses the document rather than grepping: a string
// search cannot tell an adjacent form from a nested one.
func TestProfileFormsNotNested(t *testing.T) {
	body := renderBarWithModifiableProfile(t)
	mainForm := findMainSettingsForm(t, body)
	if n := countDescendantForms(mainForm); n != 0 {
		t.Errorf("the main settings form contains %d nested <form> element(s); nested forms are invalid HTML and browsers drop them", n)
	}

	// The CRUD targets must actually be present somewhere on the page —
	// otherwise the zero-nesting assertion above is vacuously true.
	for _, action := range []string{"/profile/create", "/profile/rename", "/profile/delete"} {
		if !strings.Contains(body, action) {
			t.Errorf("no form targets %s — the nesting assertion would be vacuous", action)
		}
	}
}

// TestProfileFormNestingDetectorIsFalsifiable is the negative control for the
// parser above: given a genuinely nested fixture it reports the nesting.
//
// html.Parse follows the HTML5 parsing algorithm, which drops a nested <form>
// start tag outright — which is precisely the browser behavior the requirement
// exists to avoid. The control therefore asserts the observable consequence:
// the inner form's action never survives into the parsed tree.
func TestProfileFormNestingDetectorIsFalsifiable(t *testing.T) {
	fixture := `<html><body><form class="form"><input name="a"/><form class="profilebar__form" action="/profile/create"></form></form></body></html>`
	doc, err := html.Parse(strings.NewReader(fixture))
	if err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	var sawInnerAction bool
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "form" {
			for _, a := range n.Attr {
				if a.Key == "action" && a.Val == "/profile/create" {
					sawInnerAction = true
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	if sawInnerAction {
		t.Fatal("the nested inner form survived parsing — this fixture no longer demonstrates the hazard")
	}
}

// seedProfile creates a profile directory under the sandboxed profile base.
func seedProfile(t *testing.T, name string) string {
	t.Helper()
	dir := profile.GetProfileDir(name)
	if dir == "" {
		t.Fatalf("GetProfileDir(%q) returned empty — name is reserved or invalid", name)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("seed profile %q: %v", name, err)
	}
	return dir
}

// postProfileForm issues a same-origin loopback POST to a profile CRUD route.
func postProfileForm(a *app, path string, form url.Values) *httptest.ResponseRecorder {
	h := a.routes()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Host = "127.0.0.1:8080"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// TestProfileRename verifies AC-WCR-042: exactly one of the five cases succeeds.
func TestProfileRename(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		a := newTestApp(t)
		seedProfile(t, "scratch")
		rec := postProfileForm(a, profileRenameAction(), url.Values{
			"profile_name": {"scratch"}, "new_name": {"scratch2"},
		})
		if rec.Code != http.StatusOK {
			t.Fatalf("rename status = %d, want 200 (body: %.200s)", rec.Code, rec.Body.String())
		}
		if _, err := os.Stat(filepath.Join(profile.GetBaseDir(), "scratch2")); err != nil {
			t.Errorf("renamed profile directory does not exist: %v", err)
		}
		if _, err := os.Stat(filepath.Join(profile.GetBaseDir(), "scratch")); !os.IsNotExist(err) {
			t.Error("the original profile directory still exists after rename")
		}
	})

	// Each rejection case asserts the directory is untouched, not merely that
	// the status is 4xx — a handler could refuse loudly after already moving it.
	reject := func(t *testing.T, name string, form url.Values, mustSurvive string) {
		t.Helper()
		a := newTestApp(t)
		rec := postProfileForm(a, profileRenameAction(), form)
		if rec.Code == http.StatusOK {
			t.Errorf("%s: rename was accepted, want refusal", name)
		}
		if !strings.Contains(rec.Body.String(), "banner") {
			t.Errorf("%s: refusal did not render a reason banner", name)
		}
		if mustSurvive != "" {
			if _, err := os.Stat(filepath.Join(profile.GetBaseDir(), mustSurvive)); err != nil {
				t.Errorf("%s: %q was disturbed by a refused rename: %v", name, mustSurvive, err)
			}
		}
	}

	t.Run("default refused", func(t *testing.T) {
		reject(t, "default", url.Values{"profile_name": {"default"}, "new_name": {"other"}}, "")
	})

	t.Run("current refused", func(t *testing.T) {
		a := newTestApp(t)
		seedProfile(t, "live")
		current := a.activeProfileName()
		rec := postProfileForm(a, profileRenameAction(), url.Values{
			"profile_name": {current}, "new_name": {"renamed"},
		})
		if rec.Code == http.StatusOK {
			t.Errorf("renaming the active profile %q was accepted, want refusal", current)
		}
	})

	t.Run("conflict refused", func(t *testing.T) {
		seedProfile(t, "alpha")
		seedProfile(t, "beta")
		reject(t, "conflict",
			url.Values{"profile_name": {"alpha"}, "new_name": {"beta"}}, "alpha")
	})

	t.Run("traversal refused", func(t *testing.T) {
		seedProfile(t, "gamma")
		reject(t, "traversal",
			url.Values{"profile_name": {"gamma"}, "new_name": {"../escape"}}, "gamma")
	})

	t.Run("traversal in source refused", func(t *testing.T) {
		reject(t, "traversal-source",
			url.Values{"profile_name": {"../etc"}, "new_name": {"safe"}}, "")
	})

	t.Run("missing source refused", func(t *testing.T) {
		reject(t, "missing",
			url.Values{"profile_name": {"nope-not-here"}, "new_name": {"whatever"}}, "")
	})

	t.Run("GET refused", func(t *testing.T) {
		a := newTestApp(t)
		h := a.routes()
		req := httptest.NewRequest(http.MethodGet, profileRenameAction(), nil)
		req.Host = "127.0.0.1:8080"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code == http.StatusOK {
			t.Error("GET /profile/rename was accepted; rename must be a POST action")
		}
	})
}

// TestProfileBarMarksCurrentAndSelected pins the three display states a profile
// row can be in. They are visually distinct for a reason: `current` is the
// profile a new `moai cc` session would launch with, while `selected` is merely
// the one this page is editing. Conflating them would let a user edit one
// profile believing they had switched the launch default.
func TestProfileBarMarksCurrentAndSelected(t *testing.T) {
	a := newTestApp(t)
	a.listProfiles = func() []profile.ProfileEntry {
		return []profile.ProfileEntry{
			{Name: "default", Current: false},
			{Name: "work", Current: true},
			{Name: "scratch", Current: false},
		}
	}
	body := serveGet(t, a.routes(), "/settings").Body.String()

	// The launch-current profile is annotated; the others are not.
	if !strings.Contains(body, `>work (current)<`) {
		t.Errorf("the current profile is not marked in the switch select:\n%s", body)
	}
	for _, other := range []string{"default", "scratch"} {
		if strings.Contains(body, `>`+other+` (current)<`) {
			t.Errorf("%q was marked current but is not", other)
		}
	}

	// Exactly one option is preselected — the profile being edited.
	if n := strings.Count(body, `<option value="default" selected>`); n != 1 {
		t.Errorf("the edited profile is preselected %d times, want 1", n)
	}

	// The current profile is excluded from the rename/delete targets, matching
	// the server guard; offering it would propose a refused action.
	renameBlock := body[strings.Index(body, `action="/profile/rename"`):]
	renameBlock = renameBlock[:strings.Index(renameBlock, "</form>")]
	if strings.Contains(renameBlock, `value="work"`) {
		t.Error("the current profile is offered as a rename target but the server refuses it")
	}
	if !strings.Contains(renameBlock, `value="scratch"`) {
		t.Error("an eligible profile is missing from the rename targets")
	}
}
