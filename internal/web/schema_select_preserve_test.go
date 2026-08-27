package web

// schema_select_preserve_test.go — RC2 regression guards for out-of-option-set
// select coercion (glm-settings-persist).
//
// selectedSchemaValue returns the raw persisted value with no guard that it
// exists in the rendered option set. An HTML <select> whose current value
// matches no <option> auto-selects the FIRST option, and selects always submit
// on form POST — so ANY console save from ANY tab silently rewrote an unknown
// value to the first option (e.g. a tier slot still holding glm-5.2, which the
// offered set no longer lists). The fix: render one synthetic option carrying
// the exact persisted value (marked selected) so the submission round-trips,
// and let validation accept a submitted value that EQUALS the previously
// persisted value (passthrough-preserve) while still rejecting genuinely new
// out-of-set submissions.

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/defs"
)

// newAppWithRoot builds the console app against a caller-owned project root so
// a test can seed section files before rendering.
func newAppWithRoot(t *testing.T, root string) *app {
	t.Helper()
	a := newApp(Config{ProjectRoot: root, ProfileName: "default"})
	a.recordLastProfile = func(string) error { return nil }
	return a
}

// seedSectionFile writes a raw section yaml under <root>/.moai/config/sections/.
func seedSectionFile(t *testing.T, root, name, content string) {
	t.Helper()
	dir := filepath.Join(root, defs.MoAIDir, "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("write %s.yaml: %v", name, err)
	}
}

// readSectionFile reads a seeded section file back.
func readSeededSectionFile(t *testing.T, root, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, defs.MoAIDir, "config", "sections", name+".yaml"))
	if err != nil {
		t.Fatalf("read %s.yaml: %v", name, err)
	}
	return string(data)
}

// seedExoticSelectValues seeds persisted values that are loadable but outside
// the running binary's offered option sets:
//   - llm.glm.models.high = glm-5.2 — withdrawn from ValidGLMModels, still a
//     legal stored value (select widget)
//   - workflow.audit.glm.model = glm-5.2 — the audit GLM pin select (same
//     closed set)
//   - workflow.audit.model = custom — the audit backend picker, a RADIO
//     (outside ValidAuditModels; an unlisted radio leaves every button
//     unchecked, so the browser submits nothing)
func seedExoticSelectValues(t *testing.T, root string) {
	t.Helper()
	seedSectionFile(t, root, "llm",
		"llm:\n  glm:\n    models:\n      high: "+config.DefaultGLM52+"\n")
	seedSectionFile(t, root, "workflow",
		"workflow:\n  audit:\n    model: custom\n    glm:\n      model: "+config.DefaultGLM52+"\n")
}

// selOptionRe matches one rendered <option> element, capturing its value
// attribute and the remaining attributes (to test for `selected`).
var selOptionRe = regexp.MustCompile(`<option value="([^"]*)"([^>]*)>`)

// browserSelectedValue returns the value a real browser would submit for the
// named select: the option carrying `selected`, or the FIRST option when none
// is marked (the browser's auto-select rule — the exact mechanism behind the
// silent revert).
func browserSelectedValue(t *testing.T, body, fieldName string) string {
	t.Helper()
	i := strings.Index(body, `name="`+fieldName+`"`)
	if i < 0 {
		t.Fatalf("select %q is not rendered on the page", fieldName)
	}
	seg := body[i:]
	if j := strings.Index(seg, "</select>"); j >= 0 {
		seg = seg[:j]
	}
	first := ""
	for _, m := range selOptionRe.FindAllStringSubmatch(seg, -1) {
		if first == "" {
			first = m[1]
		}
		if strings.Contains(m[2], "selected") {
			return m[1]
		}
	}
	if first == "" {
		t.Fatalf("select %q renders no options", fieldName)
	}
	return first
}

// TestSchemaSelectSyntheticCurrentOptionRendered pins the render half of RC2:
// when the persisted value is outside the offered option set, the select must
// render ONE synthetic option carrying that exact raw value, marked selected,
// so the browser submits it back unchanged instead of auto-selecting the first
// offered option.
func TestSchemaSelectSyntheticCurrentOptionRendered(t *testing.T) {
	root := t.TempDir()
	seedExoticSelectValues(t, root)
	body := renderSettingsGET(newAppWithRoot(t, root))

	// The two SELECT widgets carrying the out-of-set value. (The audit backend
	// picker workflow.audit.model is a radio — an unlisted value leaves every
	// button unchecked and submits nothing, so it needs no synthetic option.)
	for _, name := range []string{"llm.glm.models.high", "workflow.audit.glm.model"} {
		if got := browserSelectedValue(t, body, name); got != config.DefaultGLM52 {
			t.Errorf("%s: browser would submit %q, want the persisted %q — no selected synthetic option, the first offered option wins and silently rewrites the value", name, got, config.DefaultGLM52)
		}
		if !strings.Contains(body, config.DefaultGLM52+schemaSavedCurrentSuffix) {
			t.Errorf("synthetic option for %q lacks the (saved) label", config.DefaultGLM52)
		}
	}
}

// TestBrowserRoundTripPreservesExoticSelectValues pins the full RC2 chain:
// what the rendered page tells the browser to submit, when POSTed to /save,
// must leave the persisted exotic values intact on disk. The radio field
// (workflow.audit.model) is deliberately absent from the POST — an unchecked
// radio group submits nothing — and must survive via empty=preserve.
func TestBrowserRoundTripPreservesExoticSelectValues(t *testing.T) {
	root := t.TempDir()
	seedExoticSelectValues(t, root)
	a := newAppWithRoot(t, root)
	body := renderSettingsGET(a)

	form := url.Values{}
	form.Set("__profile", "default")
	for _, name := range []string{"llm.glm.models.high", "workflow.audit.glm.model"} {
		form.Set(name, browserSelectedValue(t, body, name))
	}
	rec := postGLMSave(a, form)
	if rec.Code != http.StatusOK {
		t.Fatalf("save status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}

	if got := readSeededSectionFile(t, root, "llm"); !strings.Contains(got, config.DefaultGLM52) {
		t.Errorf("llm.yaml lost the persisted exotic tier model after a browser-equivalent save:\n%s", got)
	}
	if got := readSeededSectionFile(t, root, "workflow"); !strings.Contains(got, "model: custom") {
		t.Errorf("workflow.yaml lost the persisted exotic audit backend (radio, unsubmitted) after a browser-equivalent save:\n%s", got)
	}
	if got := readSeededSectionFile(t, root, "workflow"); !strings.Contains(got, config.DefaultGLM52) {
		t.Errorf("workflow.yaml lost the persisted exotic audit GLM pin after a browser-equivalent save:\n%s", got)
	}
}

// TestParseSchemaFormPassthroughPreserve pins the validation half of RC2 at the
// unit level: a submitted value that EQUALS the previously persisted value is
// accepted even outside the offered closed set (so the synthetic-option
// round-trip can save), while a genuinely NEW out-of-set submission is still
// rejected.
func TestParseSchemaFormPassthroughPreserve(t *testing.T) {
	current := map[string]string{
		"llm.glm.models.high":  config.DefaultGLM52,
		"workflow.audit.model": "custom",
	}

	t.Run("value equal to persisted passes outside the closed set", func(t *testing.T) {
		edits, errs := parseSchemaForm(postSchemaForm(url.Values{
			"llm.glm.models.high":  {config.DefaultGLM52},
			"workflow.audit.model": {"custom"},
		}), current)
		if len(errs) != 0 {
			t.Fatalf("parseSchemaForm errors: %v", errs)
		}
		if got := edits["llm.glm.models.high"]; got != config.DefaultGLM52 {
			t.Errorf("edits[llm.glm.models.high] = %q, want the round-tripped %q", got, config.DefaultGLM52)
		}
		if got := edits["workflow.audit.model"]; got != "custom" {
			t.Errorf("edits[workflow.audit.model] = %q, want the round-tripped %q", got, "custom")
		}
	})

	t.Run("genuinely new out-of-set submission still rejected", func(t *testing.T) {
		edits, errs := parseSchemaForm(postSchemaForm(url.Values{
			"llm.glm.models.high": {"not-a-glm-model"},
		}), current)
		if _, ok := errs["llm.glm.models.high"]; !ok {
			t.Errorf("errs = %v, want an llm.glm.models.high entry (new out-of-set value rejected)", errs)
		}
		if _, ok := edits["llm.glm.models.high"]; ok {
			t.Errorf("out-of-set value became an edit — closed-set rejection must still hold")
		}
	})

	t.Run("nil current keeps the strict closed set", func(t *testing.T) {
		_, errs := parseSchemaForm(postSchemaForm(url.Values{
			"llm.glm.models.high": {config.DefaultGLM52},
		}), nil)
		if _, ok := errs["llm.glm.models.high"]; !ok {
			t.Errorf("errs = %v, want rejection when no persisted value backs the passthrough", errs)
		}
	})
}
