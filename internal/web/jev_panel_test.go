package web

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/jevcred"
	"github.com/modu-ai/moai-adk/internal/settings"
)

// jev_panel_test.go — the `moai web` Jev opt-in section
// (SPEC-JEV-OPTIN-MEASURE-001 REQ-JEVO-001/002/003/004; AC-JEVO-001/004/006).
//
// The section is ONE sub-section on the existing /settings screen, inside the
// workflow panel: an enable toggle (a schema FieldDef, persisted through the
// shared internal/settings seam) and a credential field (hand-built,
// deliberately outside the schema so no generic schema-walking loop can render
// or write it).
//
// Fixture credentials below are deliberately NOT key-shaped. A realistic-looking
// literal trips the repository's hardcoded-credential scanner, and a test
// fixture is not worth an exemption.

// jevFixtureLongValue stands in for a stored credential longer than the
// four-character disclosure floor. Its last four characters are "1234".
const jevFixtureLongValue = "fixture-value-1234"

// TestJevPanel_CarriesToggleAndCredential is AC-JEVO-001.
func TestJevPanel_CarriesToggleAndCredential(t *testing.T) {
	html := renderConsolePage(t)

	// Exactly one Jev section.
	if n := strings.Count(html, `data-section="jev"`); n != 1 {
		t.Fatalf(`data-section="jev" appears %d times, want exactly 1`, n)
	}
	body := jevSectionHTML(t, html)

	for _, control := range []string{settings.JevEnabledField, jevAPIKeyFormField} {
		marker := `name="` + control + `"`
		if !strings.Contains(body, marker) {
			t.Errorf("the Jev section is missing control %q", control)
			continue
		}
		// A control that also renders elsewhere would be submitted twice, and
		// parseSchemaForm rejects duplicate submissions outright.
		if strings.Count(html, marker) != strings.Count(body, marker) {
			t.Errorf("control %q also renders outside the Jev section", control)
		}
	}

	// Positive control: the slice is a real panel body, not an empty string
	// that would pass every "missing" check above by vacuity if inverted.
	if strings.TrimSpace(body) == "" {
		t.Fatal("positive control failed: the Jev section body sliced empty")
	}
}

// TestJevCredentialInput_IsSecretClass: the credential control never echoes a
// stored key back into the form, and opts out of browser autofill.
func TestJevCredentialInput_IsSecretClass(t *testing.T) {
	body := jevSectionHTML(t, renderConsolePage(t))
	start := strings.Index(body, `name="`+jevAPIKeyFormField+`"`)
	if start < 0 {
		t.Fatal("the credential control is absent")
	}
	// Widen to the enclosing <input ...> tag.
	open := strings.LastIndex(body[:start], "<input")
	if open < 0 {
		t.Fatal("the credential control is not an <input>")
	}
	end := strings.Index(body[open:], ">")
	tag := body[open : open+end]

	for _, want := range []string{`type="password"`, `autocomplete="off"`, `value=""`} {
		if !strings.Contains(tag, want) {
			t.Errorf("credential input lacks %s\ntag: %s", want, tag)
		}
	}
}

// TestJevPanel_StatesThirdPartySend is AC-JEVO-004 for the console half: the
// section body carries the i18n key whose value states the privacy cost, and
// the inline English baseline states it too.
func TestJevPanel_StatesThirdPartySend(t *testing.T) {
	body := jevSectionHTML(t, renderConsolePage(t))
	if !strings.Contains(body, jevPrivacyNoteKey) {
		t.Errorf("the Jev section does not render the privacy note key %q", jevPrivacyNoteKey)
	}
	// The English baseline is rendered inline (data-i18n carries the key, the
	// element text carries the fallback), so the statement is present even
	// before the catalogue loads.
	if !strings.Contains(body, "third-party") {
		t.Error("the rendered Jev section does not state the third-party send in its inline baseline")
	}
}

// TestJevI18nKeys_PresentInFourLocales is AC-JEVO-004/005 for the console:
// every Jev key exists under en, ko, ja and zh, and the non-en values are not
// byte-identical to the English one (which would be an untranslated defect
// rather than a translation).
func TestJevI18nKeys_PresentInFourLocales(t *testing.T) {
	catalogue := readI18nCatalogue(t)

	keys := []string{
		"tab.jev.title",
		"tab.jev.desc",
		jevPrivacyNoteKey,
		"f." + settings.JevEnabledField + ".title",
		"f." + settings.JevEnabledField + ".desc",
		"f." + jevAPIKeyFormField + ".title",
		"f." + jevAPIKeyFormField + ".desc",
	}
	for _, key := range keys {
		en, ok := catalogue["en"][key]
		if !ok {
			t.Errorf("i18n key %q missing from locale en", key)
			continue
		}
		for _, loc := range []string{"ko", "ja", "zh"} {
			v, ok := catalogue[loc][key]
			if !ok {
				t.Errorf("i18n key %q missing from locale %s", key, loc)
				continue
			}
			if v == en {
				t.Errorf("i18n key %q is byte-identical to en in locale %s — untranslated", key, loc)
			}
		}
	}

	// Positive control: the reader found a real catalogue, and reports an
	// impossible key as absent rather than present.
	if _, ok := catalogue["en"]["tab.jev.absent-control"]; ok {
		t.Error("positive control failed: the catalogue reported an impossible key as present")
	}
	if len(catalogue["en"]) == 0 {
		t.Error("positive control failed: the en catalogue parsed empty")
	}
}

// TestJevCredential_AbsentFromSchema is the structural anti-leak guard, the
// sibling of TestGLMKeyField_AbsentFromSchema. The credential must never enter
// settings.AllFields(): a generic schema-walking loop (bulk value read, form
// state dump, diagnostics view) would otherwise pick it up.
func TestJevCredential_AbsentFromSchema(t *testing.T) {
	for _, f := range settings.AllFields() {
		lower := strings.ToLower(f.Name)
		if strings.Contains(f.Name, jevAPIKeyFormField) ||
			strings.Contains(lower, "typesafe") ||
			strings.Contains(lower, "api_key") {
			t.Errorf("credential-shaped field %q is in the settings schema", f.Name)
		}
	}
	// Positive control: the scan covers a non-empty schema, and the toggle —
	// which IS supposed to be there — is found by the same enumeration.
	found := false
	for _, f := range settings.AllFields() {
		if f.Name == settings.JevEnabledField {
			found = true
		}
	}
	if !found {
		t.Error("positive control failed: the enumeration did not find the Jev toggle either")
	}
}

// TestJevKeyHint_BoundedDisclosure: the view model surfaces at most the final
// four characters, and nothing at all for a value of four characters or fewer.
// A naive "last four, or the whole value if shorter" fallback would disclose a
// short credential entirely — the exact inverse of the requirement.
func TestJevKeyHint_BoundedDisclosure(t *testing.T) {
	cases := []struct {
		name           string
		stored         string
		wantConfigured bool
		wantHint       string
	}{
		{"absent", "", false, ""},
		{"four characters", "abcd", true, ""},
		{"three characters", "abc", true, ""},
		{"long", jevFixtureLongValue, true, "1234"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withStubbedJevCredential(t, tc.stored)
			h := computeJevKeyHint()
			if h.Configured != tc.wantConfigured {
				t.Errorf("Configured = %v, want %v", h.Configured, tc.wantConfigured)
			}
			if h.Hint != tc.wantHint {
				t.Errorf("Hint = %q, want %q", h.Hint, tc.wantHint)
			}
			if len(tc.stored) > 4 && h.Hint == tc.stored {
				t.Error("the hint carries the whole stored value")
			}
		})
	}
}

// TestJevKeyValidation mirrors the GLM credential rules: empty preserves, a
// body with surrounding whitespace is trimmed, an embedded line break is
// rejected.
func TestJevKeyValidation(t *testing.T) {
	cases := []struct {
		name      string
		submitted string
		wantErr   bool
		wantWrite string
	}{
		{"empty preserves", "", false, ""},
		{"whitespace preserves", "   \n ", false, ""},
		{"surrounding whitespace trimmed", "  value-abc  \n", false, "value-abc"},
		{"embedded line break rejected", "value-abc\ndef", true, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			errs := validateJevKey(tc.submitted)
			if got := len(errs) > 0; got != tc.wantErr {
				t.Errorf("validateJevKey error = %v, want %v (errs=%v)", got, tc.wantErr, errs)
			}
			if tc.wantErr {
				return
			}
			if got := normalizeJevKey(tc.submitted); got != tc.wantWrite {
				t.Errorf("normalizeJevKey = %q, want %q", got, tc.wantWrite)
			}
		})
	}
}

// TestConsoleRoutes_NoInitRoute is AC-JEVO-006: this SPEC adds no route, and
// in particular no init route. The console's route set is enumerated from the
// source registration block.
func TestConsoleRoutes_NoInitRoute(t *testing.T) {
	src, err := os.ReadFile("app.go")
	if err != nil {
		t.Fatalf("read app.go: %v", err)
	}
	text := string(src)
	for _, forbidden := range []string{`"/init"`, `"/jev"`, `jevKeyRevealPath`} { //nolint:gocritic // explicit list
		if strings.Contains(text, forbidden) {
			t.Errorf("app.go registers %s — this SPEC adds no route", forbidden)
		}
	}
	// Positive control: the search fires on a route that DOES exist.
	if !strings.Contains(text, `"/settings"`) {
		t.Error("positive control failed: the scan did not find the /settings route")
	}
}

// withStubbedJevCredential points the credential reader at a temp home holding
// the given value (empty means no file at all), restoring the seam afterwards.
// It never touches the developer's real home directory.
func withStubbedJevCredential(t *testing.T, stored string) {
	t.Helper()
	home := t.TempDir()
	if stored != "" {
		dir := filepath.Join(home, ".moai")
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		body := "TYPESAFE_API_KEY=" + jevcred.EscapeValue(stored) + "\n"
		if err := os.WriteFile(filepath.Join(dir, ".env.typesafe"), []byte(body), 0o600); err != nil {
			t.Fatalf("write credential: %v", err)
		}
	}
	prev := jevcred.HomeDirFn
	jevcred.HomeDirFn = func() (string, error) { return home, nil }
	t.Cleanup(func() { jevcred.HomeDirFn = prev })
}

// readI18nCatalogue parses internal/web/assets/i18n.js into locale -> key ->
// value. The file is a JS object literal with one block per locale; the parse
// is deliberately line-oriented rather than a JS engine, matching how the
// sibling i18n governance tests read it.
func readI18nCatalogue(t *testing.T) map[string]map[string]string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("assets", "i18n.js"))
	if err != nil {
		t.Fatalf("read i18n.js: %v", err)
	}
	out := map[string]map[string]string{}
	locale := ""
	for _, line := range strings.Split(string(b), "\n") {
		trimmed := strings.TrimSpace(line)
		for _, loc := range []string{"en", "ko", "ja", "zh"} {
			if trimmed == loc+": {" {
				locale = loc
				out[loc] = map[string]string{}
			}
		}
		if locale == "" || !strings.HasPrefix(trimmed, `"`) {
			continue
		}
		rest := trimmed[1:]
		sep := strings.Index(rest, `":`)
		if sep < 0 {
			continue
		}
		key := rest[:sep]
		val := strings.TrimSpace(rest[sep+2:])
		val = strings.TrimSuffix(val, ",")
		val = strings.Trim(val, `"`)
		out[locale][key] = val
	}
	return out
}

// TestJevPanel_RendersOnSettingsRoute guards that the section renders on the
// real route rather than only through the test helper.
func TestJevPanel_RendersOnSettingsRoute(t *testing.T) {
	a := newTestApp(t)
	rec := serveGet(t, a.routes(), "/settings")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /settings status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `data-section="jev"`) {
		t.Error("the Jev section does not render on /settings")
	}
}

// TestJevSection_LivesInsideWorkflowPanel pins the placement decision: the Jev
// section is a sub-section of the workflow panel, NOT a tab. A tab would
// couple this change to eight documentation surfaces that a run-phase change
// does not own (TestDocsTabContract enforces that coupling).
func TestJevSection_LivesInsideWorkflowPanel(t *testing.T) {
	html := renderConsolePage(t)
	if strings.Contains(html, `data-panel="jev"`) {
		t.Error("a Jev TAB was rendered; the section belongs inside the workflow panel")
	}
	workflow := panelHTML(t, html, "workflow")
	if !strings.Contains(workflow, `data-section="jev"`) {
		t.Error("the Jev section does not render inside the workflow panel")
	}
	for _, tb := range consoleTabs() {
		if tb.ID == "jev" {
			t.Error("consoleTabs() carries a jev tab")
		}
	}
}

// jevSectionHTML slices the rendered markup for the Jev sub-section: from its
// marker to the end of the enclosing panel.
func jevSectionHTML(t *testing.T, html string) string {
	t.Helper()
	marker := `data-section="jev"`
	start := strings.Index(html, marker)
	if start < 0 {
		t.Fatal("the Jev section is not present in the rendered console")
	}
	rest := html[start+len(marker):]
	if next := strings.Index(rest, `data-panel="`); next >= 0 {
		rest = rest[:next]
	}
	return rest
}
