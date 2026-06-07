package web

import (
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

// renderTempl renders a templ.Component to a string for markup-parity assertions.
// Additive M2 test harness for SPEC-WEB-CONSOLE-006 (does NOT modify any Class A
// or Class B test).
func renderTempl(t *testing.T, c templ.Component) string {
	t.Helper()
	var sb strings.Builder
	if err := c.Render(context.Background(), &sb); err != nil {
		t.Fatalf("render: %v", err)
	}
	return sb.String()
}

// TestLangSelectHelperMarkupParity verifies the ported langSelect Templ component
// emits the exact markup contract from the retired page.html.tmpl langSelect
// helper: field chrome, select--lang, (unset) empty option, name=/id=, data-i18n,
// the {{range}}->for option loop with `selected` on the current value, and the
// error span only when an error is present (REQ-WC6-002).
func TestLangSelectHelperMarkupParity(t *testing.T) {
	// Clean (no error) render with a current value of "ko".
	clean := renderTempl(t, langSelect(langSelectArgs{
		Name:    "conversation_lang",
		Title:   "Conversation language",
		Key:     "conversation_lang",
		Desc:    "Language the assistant replies in during chat.",
		Value:   "ko",
		Options: langOptions,
		Errors:  map[string]string{},
	}))

	for _, want := range []string{
		`class="field"`, // no has-error
		`class="field__title" data-i18n="f.conversation_lang.title"`,
		`<code class="field__key">conversation_lang</code>`,
		`class="field__desc" data-i18n="f.conversation_lang.desc"`,
		`class="select-wrap"`,
		`class="select select--lang"`,
		`id="conversation_lang"`,
		`name="conversation_lang"`,
		`<option value="">(unset)</option>`, // not selected (value=ko)
		`<option value="ko" selected>ko</option>`,
		`<option value="en">en</option>`,
	} {
		if !strings.Contains(clean, want) {
			t.Errorf("langSelect clean render missing %q\n--- rendered ---\n%s", want, clean)
		}
	}
	// No aria-invalid / field-error on a clean render.
	if strings.Contains(clean, "aria-invalid") {
		t.Errorf("langSelect clean render unexpectedly carries aria-invalid:\n%s", clean)
	}
	if strings.Contains(clean, "field-error") {
		t.Errorf("langSelect clean render unexpectedly carries a field-error span:\n%s", clean)
	}

	// Errored render: unset value + a validation error.
	errored := renderTempl(t, langSelect(langSelectArgs{
		Name:    "conversation_lang",
		Title:   "Conversation language",
		Key:     "conversation_lang",
		Desc:    "Language the assistant replies in during chat.",
		Value:   "",
		Options: langOptions,
		Errors:  map[string]string{"conversation_lang": "unrecognized language: xx"},
	}))
	for _, want := range []string{
		`class="field has-error"`,
		`aria-invalid="true"`,
		`<option value="" selected>(unset)</option>`, // unset value selected
		`class="field-error"`,
		`unrecognized language: xx`,
		`class="icon-alert-circle"`, // error icon
	} {
		if !strings.Contains(errored, want) {
			t.Errorf("langSelect errored render missing %q\n--- rendered ---\n%s", want, errored)
		}
	}
}

// TestOptSelectHelperMarkupParity verifies the ported optSelect Templ component
// emits the exact markup contract: field chrome, plain `select` (no select--lang),
// the caller-labelled `.Empty` empty option, name=/id=, data-i18n, the option loop
// with `selected` on the current value, and the error span only when errored.
func TestOptSelectHelperMarkupParity(t *testing.T) {
	clean := renderTempl(t, optSelect(optSelectArgs{
		Name:    "model",
		Title:   "Model",
		Key:     "model",
		Desc:    "The specific model to run, including 1M-context variants.",
		Value:   "sonnet",
		Empty:   "(project default)",
		Options: modelCanonical,
		Errors:  map[string]string{},
	}))
	for _, want := range []string{
		`class="field"`,
		`class="field__title" data-i18n="f.model.title"`,
		`<code class="field__key">model</code>`,
		`class="select"`, // plain select chrome (no select--lang)
		`id="model"`,
		`name="model"`,
		`<option value="">(project default)</option>`,
		`<option value="sonnet" selected>sonnet</option>`,
		`<option value="opus">opus</option>`,
	} {
		if !strings.Contains(clean, want) {
			t.Errorf("optSelect clean render missing %q\n--- rendered ---\n%s", want, clean)
		}
	}
	// optSelect must NOT carry the select--lang modifier.
	if strings.Contains(clean, "select--lang") {
		t.Errorf("optSelect render unexpectedly carries select--lang:\n%s", clean)
	}
}

// TestToggleHelperMarkupParity verifies the new toggle Templ component
// (SPEC-WEB-CONSOLE-007 §D, REQ-WC7-004) emits the boolean-checkbox markup
// contract: field chrome, a checkbox name=/value="1" carrying the field name, the
// hidden companion input (name+"__present"), `checked` only on the checked
// render, and the error span only when errored. Mirrors the
// TestOptSelectHelperMarkupParity pattern.
func TestToggleHelperMarkupParity(t *testing.T) {
	// Clean, unchecked render.
	unchecked := renderTempl(t, toggle(toggleArgs{
		Name:   "quality.enforce_quality",
		Title:  "Enforce quality gate",
		Key:    "quality.enforce_quality",
		Desc:   "Block on TRUST 5 failures.",
		Errors: map[string]string{},
	}))
	for _, want := range []string{
		`class="field"`, // no has-error
		`class="field__title" data-i18n="f.quality.enforce_quality.title"`,
		`<code class="field__key">quality.enforce_quality</code>`,
		`class="field__desc" data-i18n="f.quality.enforce_quality.desc"`,
		`<input type="hidden" name="quality.enforce_quality__present" value="1"`,
		`type="checkbox"`,
		`id="quality.enforce_quality"`,
		`name="quality.enforce_quality"`,
		`value="1"`,
	} {
		if !strings.Contains(unchecked, want) {
			t.Errorf("toggle unchecked render missing %q\n--- rendered ---\n%s", want, unchecked)
		}
	}
	// Unchecked render must NOT carry `checked`, has-error, or a field-error span.
	if strings.Contains(unchecked, " checked") {
		t.Errorf("toggle unchecked render unexpectedly carries `checked`:\n%s", unchecked)
	}
	if strings.Contains(unchecked, "has-error") {
		t.Errorf("toggle clean render unexpectedly carries has-error:\n%s", unchecked)
	}
	if strings.Contains(unchecked, "field-error") {
		t.Errorf("toggle clean render unexpectedly carries a field-error span:\n%s", unchecked)
	}

	// Checked render carries `checked`.
	checked := renderTempl(t, toggle(toggleArgs{
		Name:    "quality.enforce_quality",
		Title:   "Enforce quality gate",
		Key:     "quality.enforce_quality",
		Desc:    "Block on TRUST 5 failures.",
		Checked: true,
		Errors:  map[string]string{},
	}))
	if !strings.Contains(checked, `value="1" checked`) {
		t.Errorf("toggle checked render missing `value=\"1\" checked`:\n%s", checked)
	}

	// Errored render carries has-error + field-error span + alert-circle icon.
	errored := renderTempl(t, toggle(toggleArgs{
		Name:   "quality.enforce_quality",
		Title:  "Enforce quality gate",
		Key:    "quality.enforce_quality",
		Desc:   "Block on TRUST 5 failures.",
		Errors: map[string]string{"quality.enforce_quality": "boom"},
	}))
	for _, want := range []string{
		`class="field has-error"`,
		`class="field-error"`,
		`boom`,
		`class="icon-alert-circle"`,
	} {
		if !strings.Contains(errored, want) {
			t.Errorf("toggle errored render missing %q\n--- rendered ---\n%s", want, errored)
		}
	}
}

// TestNumberFieldHelperMarkupParity verifies the new numberField Templ component
// (SPEC-WEB-CONSOLE-007 §D, REQ-WC7-004) emits the number-input markup contract:
// field chrome, <input type="number"> with name=/value=/min=/max=/step=, and on an
// errored render aria-invalid="true" + an aria-describedby field-error span.
func TestNumberFieldHelperMarkupParity(t *testing.T) {
	clean := renderTempl(t, numberField(numberFieldArgs{
		Name:   "quality.test_coverage_target",
		Title:  "Coverage target",
		Key:    "quality.test_coverage_target",
		Desc:   "Minimum coverage percentage.",
		Value:  "85",
		Min:    "0",
		Max:    "100",
		Step:   "1",
		Errors: map[string]string{},
	}))
	for _, want := range []string{
		`class="field"`,
		`class="field__title" data-i18n="f.quality.test_coverage_target.title"`,
		`<code class="field__key">quality.test_coverage_target</code>`,
		`type="number"`,
		`id="quality.test_coverage_target"`,
		`name="quality.test_coverage_target"`,
		`value="85"`,
		`min="0"`,
		`max="100"`,
		`step="1"`,
	} {
		if !strings.Contains(clean, want) {
			t.Errorf("numberField clean render missing %q\n--- rendered ---\n%s", want, clean)
		}
	}
	if strings.Contains(clean, "aria-invalid") {
		t.Errorf("numberField clean render unexpectedly carries aria-invalid:\n%s", clean)
	}
	if strings.Contains(clean, "field-error") {
		t.Errorf("numberField clean render unexpectedly carries a field-error span:\n%s", clean)
	}

	errored := renderTempl(t, numberField(numberFieldArgs{
		Name:   "quality.test_coverage_target",
		Title:  "Coverage target",
		Key:    "quality.test_coverage_target",
		Desc:   "Minimum coverage percentage.",
		Value:  "150",
		Min:    "0",
		Max:    "100",
		Step:   "1",
		Errors: map[string]string{"quality.test_coverage_target": "must be between 0 and 100"},
	}))
	for _, want := range []string{
		`class="field has-error"`,
		`aria-invalid="true"`,
		`aria-describedby="err_quality.test_coverage_target"`,
		`class="field-error"`,
		`must be between 0 and 100`,
		`class="icon-alert-circle"`,
	} {
		if !strings.Contains(errored, want) {
			t.Errorf("numberField errored render missing %q\n--- rendered ---\n%s", want, errored)
		}
	}
}

// TestIconHelperParity verifies the ported icon Templ component emits the inline
// SVG markup for a representative set of icon names with no CDN reference, and an
// empty string for an unknown name (REQ-WC6-017 — templ.Raw on a closed switch).
func TestIconHelperParity(t *testing.T) {
	for _, name := range []string{"alert-circle", "save", "sun", "moon", "check", "user-round"} {
		out := renderTempl(t, icon(name))
		if !strings.Contains(out, "<svg class=\"icon-"+name+"\"") {
			t.Errorf("icon(%q) did not render the inline <svg class=\"icon-%s\">: %q", name, name, out)
		}
		for _, forbidden := range []string{"unpkg.com", "lucide@", "data-lucide", "cdn.jsdelivr"} {
			if strings.Contains(out, forbidden) {
				t.Errorf("icon(%q) contains forbidden CDN reference %q", name, forbidden)
			}
		}
	}
	// An unknown icon name renders nothing (no panic, no markup).
	if out := renderTempl(t, icon("definitely-not-an-icon")); out != "" {
		t.Errorf("icon(unknown) rendered non-empty output: %q", out)
	}
}
