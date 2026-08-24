package web

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings"
)

// widget_policy_test.go — M3 guards for SPEC-WEB-CONSOLE-REDESIGN-001
// (AC-WCR-020..023). The policy: a closed value domain gets a closed widget.
// A free-text input over a closed set lets the user type an invalid value, and
// a server-side rejection afterwards does not undo the lie the UI already told.

// countCheckboxInputs counts checkbox inputs in rendered markup. It is a pure
// function so the AC-WCR-020 zero-assertion can be shown falsifiable against a
// synthetic fixture (a bare zero-count on the real page would pass forever on a
// typo).
func countCheckboxInputs(html string) int {
	return strings.Count(html, `type="checkbox"`)
}

// TestCountCheckboxInputsIsFalsifiable is the negative control for
// TestBoolFieldsRenderAsRadio: the detector reports a checkbox when one exists.
func TestCountCheckboxInputsIsFalsifiable(t *testing.T) {
	fixture := `<label class="seg"><input type="checkbox" id="x" name="x" value="1"/></label>`
	if got := countCheckboxInputs(fixture); got != 1 {
		t.Fatalf("countCheckboxInputs(one-checkbox fixture) = %d, want 1 — the zero-assertion below would be vacuous", got)
	}
	if got := countCheckboxInputs(`<input type="radio" name="x" value="1"/>`); got != 0 {
		t.Fatalf("countCheckboxInputs(radio-only fixture) = %d, want 0", got)
	}
}

// renderedBoolFieldNames returns the TypeBool field names that actually reach
// the rendered console (a schema-only field on an unrendered section cannot be
// asserted against the HTML).
func renderedBoolFieldNames(t *testing.T, html string) []string {
	t.Helper()
	var out []string
	for _, f := range settings.AllFields() {
		if f.Type != settings.TypeBool {
			continue
		}
		if strings.Contains(html, `name="`+f.Name+`"`) {
			out = append(out, f.Name)
		}
	}
	if len(out) == 0 {
		t.Fatal("no TypeBool field renders in the console — the radio assertions would be vacuous")
	}
	return out
}

// TestBoolFieldsRenderAsRadio verifies AC-WCR-020: zero checkboxes render, and
// every rendered bool field carries exactly two radio inputs (used / not used).
func TestBoolFieldsRenderAsRadio(t *testing.T) {
	html := renderConsolePage(t)

	if n := countCheckboxInputs(html); n != 0 {
		t.Errorf("rendered console has %d checkbox input(s), want 0 — bool fields render as a 2-option radio group", n)
	}

	for _, name := range renderedBoolFieldNames(t, html) {
		if got := strings.Count(html, `name="`+name+`"`); got != 2 {
			t.Errorf("bool field %q has %d inputs named %q, want exactly 2 (used / not used)", name, got, name)
		}
	}
	// The shared option labels are i18n-keyed and shared by every bool field,
	// so the key count does not grow with the field count.
	for _, key := range []string{`data-i18n="opt.enabled"`, `data-i18n="opt.disabled"`} {
		if !strings.Contains(html, key) {
			t.Errorf("rendered console missing the shared bool option label %s", key)
		}
	}
}

// TestSchemaTogglePresentCompanion verifies the first half of AC-WCR-021: the
// hidden `<name>__present` companion survives the checkbox→radio change. It is
// what lets the parser distinguish "explicitly false" from "not submitted".
func TestSchemaTogglePresentCompanion(t *testing.T) {
	html := renderConsolePage(t)
	for _, name := range renderedBoolFieldNames(t, html) {
		if !strings.Contains(html, `name="`+name+`__present"`) {
			t.Errorf("bool field %q lost its hidden __present companion", name)
		}
	}
}

// TestParseSchemaFormBoolSemanticsUnchanged verifies the second half of
// AC-WCR-021 behaviorally: parseSchemaForm's three bool outcomes are exactly
// what they were under the checkbox widget. Asserting the BEHAVIOR rather than a
// zero-line diff keeps the guard meaningful if the function is ever reformatted.
func TestParseSchemaFormBoolSemanticsUnchanged(t *testing.T) {
	const field = "workflow.branch_guard.enabled"

	cases := []struct {
		name      string
		form      url.Values
		wantValue string
		wantSet   bool
	}{
		{
			name:      "radio on → true",
			form:      url.Values{field + "__present": {"1"}, field: {"1"}},
			wantValue: "true", wantSet: true,
		},
		{
			name:      "radio off (empty value) → false",
			form:      url.Values{field + "__present": {"1"}, field: {""}},
			wantValue: "false", wantSet: true,
		},
		{
			name:    "companion absent → preserve (not an edit)",
			form:    url.Values{},
			wantSet: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/save", strings.NewReader(tc.form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			edits, errs := parseSchemaForm(req)
			if len(errs) != 0 {
				t.Fatalf("parseSchemaForm returned errors: %v", errs)
			}
			got, ok := edits[field]
			if ok != tc.wantSet {
				t.Fatalf("edits[%q] present = %v, want %v (edits: %v)", field, ok, tc.wantSet, edits)
			}
			if tc.wantSet && got != tc.wantValue {
				t.Errorf("edits[%q] = %q, want %q", field, got, tc.wantValue)
			}
		})
	}
}

// closedSetFieldNames is the AC-WCR-022 list: every field whose value domain is
// a closed set. Each must render as a closed widget AND carry membership
// validation, so an out-of-set value is refused at the config-apply layer too.
var closedSetFieldNames = []string{
	"workflow.execution_mode",
	"workflow.default_mode",
	"workflow.audit.model",
	"workflow.audit.gates.claude",
	"workflow.audit.gates.codex",
	"workflow.audit.gates.glm",
	"harness.default_profile",
	"harness.effort_mapping.minimal",
	"harness.effort_mapping.standard",
	"harness.effort_mapping.thorough",
	"harness.evaluator.memory_scope",
	"harness.mode_defaults.cg",
	"harness.mode_defaults.solo",
	"harness.mode_defaults.team",
}

// TestClosedSetFieldsAreClosedWidgets verifies AC-WCR-022.
func TestClosedSetFieldsAreClosedWidgets(t *testing.T) {
	byName := map[string]settings.FieldDef{}
	for _, f := range settings.AllFields() {
		byName[f.Name] = f
	}
	for _, name := range closedSetFieldNames {
		f, ok := byName[name]
		if !ok {
			t.Errorf("closed-set field %q is missing from the schema", name)
			continue
		}
		if f.Type != settings.TypeSelect && f.Type != settings.TypeRadio {
			t.Errorf("%q Type = %q, want select or radio (a closed domain gets a closed widget)", name, f.Type)
		}
		if f.Validate == nil {
			t.Errorf("%q has no Validate — the closed set is not enforced on save", name)
		}
		if len(f.Options) == 0 {
			t.Errorf("%q has no Options — nothing for the widget to render", name)
		}
	}
}

// TestClosedSetRejectsOutOfSetValue is the behavioral half of AC-WCR-022: a
// submitted value outside the set is refused rather than written.
func TestClosedSetRejectsOutOfSetValue(t *testing.T) {
	req := httptest.NewRequest("POST", "/save",
		strings.NewReader(url.Values{"workflow.audit.model": {"definitely-not-a-backend"}}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	edits, errs := parseSchemaForm(req)
	if _, ok := edits["workflow.audit.model"]; ok {
		t.Error("an out-of-set audit model was collected as an edit — membership validation did not fire")
	}
	if _, ok := errs["workflow.audit.model"]; !ok {
		t.Errorf("no field error recorded for the out-of-set value; errs = %v", errs)
	}
}

// freeTextWhitelist is the AC-WCR-023 set: fields whose value domain is
// genuinely open, so a text input is the honest widget.
//
// The SPEC enumerates feedback.repository, the two observability paths, and the
// GLM API key. Two more genuinely-open values are included here that the SPEC's
// enumeration omitted: `user_name` (a person's display name) and
// `security.sandbox.docker_image` (any image reference). Neither has a closed
// domain, so converting them would be dishonest in the other direction.
//
// `workflow.audit.codex.model` (SPEC-V3R6-AUDIT-MODEL-PIN-001 M4) joins them:
// the codex-servable family is an OPEN prefix set (gpt-*, o1/o3/o4-*, codex-*),
// not an enumerable closed set — new codex model ids appear upstream and a pin
// like gpt-5.6-sol cannot be pre-listed. The servability filter at the resolver
// (codexServableModel) is the validation point, not the widget.
var freeTextWhitelist = map[string]bool{
	"user_name":                     true,
	"feedback.repository":           true,
	"observability.report_dir":      true,
	"observability.trace_dir":       true,
	"security.sandbox.docker_image": true,
	"workflow.audit.codex.model":    true,
}

// m4PendingFreeText held the GLM model tier fields while their conversion was
// still M4's to do. M4 landed: all four are selects over
// config.ValidGLMModels(), so the list is empty and TestFreeTextWhitelist now
// covers them with no hole. The map is retained (rather than deleted along with
// its guard) as the place a future deferred conversion is recorded — an empty
// map with a live guard is a cheaper contract than re-deriving the pattern.
var m4PendingFreeText = map[string]bool{}

// TestFreeTextWhitelist verifies AC-WCR-023: no field survives as free text
// unless its value domain is genuinely open (or it is the M4 debt above).
func TestFreeTextWhitelist(t *testing.T) {
	for _, f := range settings.AllFields() {
		if f.Type != settings.TypeText {
			continue
		}
		if freeTextWhitelist[f.Name] || m4PendingFreeText[f.Name] {
			continue
		}
		t.Errorf("field %q is still free text but its value domain is closed — render it as a select or radio (AC-WCR-023)", f.Name)
	}
}

// TestM4PendingFreeTextStillPending keeps the debt marker honest: when M4
// converts the GLM tiers, this test fails and forces the list to be emptied
// rather than left behind as a permanent hole in TestFreeTextWhitelist.
func TestM4PendingFreeTextStillPending(t *testing.T) {
	byName := map[string]settings.FieldDef{}
	for _, f := range settings.AllFields() {
		byName[f.Name] = f
	}
	for name := range m4PendingFreeText {
		f, ok := byName[name]
		if !ok {
			t.Errorf("m4PendingFreeText names %q which is no longer a field — drop it from the list", name)
			continue
		}
		if f.Type != settings.TypeText {
			t.Errorf("%q is no longer free text (M4 landed) — remove it from m4PendingFreeText so the whitelist guard covers it", name)
		}
	}
}
