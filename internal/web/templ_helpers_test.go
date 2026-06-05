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
