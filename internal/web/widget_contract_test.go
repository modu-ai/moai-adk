package web

// widget_contract_test.go — behavior contracts for widgets and helpers that the
// currently-rendered tab set does not reach.
//
// Why these exist. The console renders five schema panels (llm, workflow,
// git-worktree, audit, report). None of them currently carries a TypeText field,
// and every ReadOnlyDisplayField / RawViewBlock belongs to a section that is not
// rendered at all (harness, observability, security, mx). So three widgets and
// four label helpers are live code with no live caller.
//
// They are not dead in the deletable sense: the moment a future SPEC renders one
// of those sections — which is exactly what M1 did for git_strategy — the widget
// is the render path. Testing the component directly pins the contract while it
// is unreachable, so the reachability change alone cannot silently break it.
// These assert rendered structure and escaping, not merely that the code runs.

import (
	"context"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings"
)

// textFieldFixture is a TypeText FieldDef with an EmptyLabel, standing in for
// the free-text fields that live on unrendered sections.
func textFieldFixture() settings.FieldDef {
	return settings.FieldDef{
		Name:       "observability.report_dir",
		Section:    settings.SectionObservability,
		Type:       settings.TypeText,
		I18nKey:    "f.observability.report_dir",
		EmptyLabel: "(runtime default)",
	}
}

// TestSchemaTextRowContract pins the free-text widget: the disk value round
// trips into the input, EmptyLabel becomes the placeholder (the "empty means
// runtime default" semantic), and the field name reaches both id and name so the
// label association and the form key agree.
func TestSchemaTextRowContract(t *testing.T) {
	f := textFieldFixture()
	var sb strings.Builder
	if err := schemaTextRow(f, "/tmp/reports", nil).Render(context.Background(), &sb); err != nil {
		t.Fatalf("render: %v", err)
	}
	html := sb.String()

	for _, want := range []string{
		`type="text"`,
		`name="observability.report_dir"`,
		`id="observability.report_dir"`,
		`value="/tmp/reports"`,
		`placeholder="(runtime default)"`,
		`autocomplete="off"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("schemaTextRow output missing %s\ngot: %s", want, html)
		}
	}
	if strings.Contains(html, "field__err") {
		t.Error("schemaTextRow marked as errored with no error supplied")
	}
}

// TestSchemaTextRowSurfacesFieldError pins the error half: a per-field error
// both marks the row and renders the message, so a rejected save is readable
// inline rather than silently discarded.
func TestSchemaTextRowSurfacesFieldError(t *testing.T) {
	f := textFieldFixture()
	errs := map[string]string{f.Name: "must be an absolute path"}
	var sb strings.Builder
	if err := schemaTextRow(f, "relative/path", errs).Render(context.Background(), &sb); err != nil {
		t.Fatalf("render: %v", err)
	}
	html := sb.String()
	if !strings.Contains(html, "field__err") {
		t.Error("a field with an error did not carry the field__err element")
	}
	if !strings.Contains(html, "must be an absolute path") {
		t.Error("the field error message was not rendered")
	}
}

// TestSchemaTextRowEscapesValue is the injection guard: a stored value is data,
// never markup. templ escapes by construction, so this pins that the widget did
// not opt out via templ.Raw or an unescaped attribute.
func TestSchemaTextRowEscapesValue(t *testing.T) {
	f := textFieldFixture()
	var sb strings.Builder
	if err := schemaTextRow(f, `"><script>alert(1)</script>`, nil).Render(context.Background(), &sb); err != nil {
		t.Fatalf("render: %v", err)
	}
	html := sb.String()
	if strings.Contains(html, "<script>") {
		t.Errorf("a stored value escaped its attribute and became markup:\n%s", html)
	}
}

// TestSchemaReadOnlyRowContract pins the read-only display row: the value is
// shown but carries NO input control, so it cannot be submitted back. That is
// the whole point of the read-only classification — a rendered <input> would
// invite an edit the write path refuses.
func TestSchemaReadOnlyRowContract(t *testing.T) {
	var sb strings.Builder
	if err := schemaReadOnlyRow("learning.auto_apply", "false", "ro.note.governance").
		Render(context.Background(), &sb); err != nil {
		t.Fatalf("render: %v", err)
	}
	html := sb.String()

	if strings.Contains(html, "<input") || strings.Contains(html, "<select") {
		t.Errorf("a read-only row rendered a submittable control:\n%s", html)
	}
	for _, want := range []string{
		"learning.auto_apply",
		"false",
		`data-i18n="ro.note.governance"`,
		"governance FROZEN",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("schemaReadOnlyRow output missing %q\ngot: %s", want, html)
		}
	}
}

// TestSchemaRawBlockContract pins the raw view: a structured block is displayed
// as collapsed text with no control, and the content is escaped.
func TestSchemaRawBlockContract(t *testing.T) {
	var sb strings.Builder
	if err := schemaRawBlock("harness.levels", "minimal:\n  evaluator_profile: <x>\n", "").
		Render(context.Background(), &sb); err != nil {
		t.Fatalf("render: %v", err)
	}
	html := sb.String()

	if strings.Contains(html, "<input") || strings.Contains(html, "<textarea") {
		t.Errorf("a raw view rendered an editable control:\n%s", html)
	}
	if !strings.Contains(html, "<details") || !strings.Contains(html, "<pre") {
		t.Errorf("raw view is not a collapsed <details>/<pre> block:\n%s", html)
	}
	if strings.Contains(html, "<x>") {
		t.Error("raw block content was not escaped — yaml content became markup")
	}
	if !strings.Contains(html, `data-i18n="raw.note"`) {
		t.Error("an empty NoteKey did not fall back to the generic raw.note label")
	}
}

// TestNoteKeyFallbacks pins the note-key resolution: an empty NoteKey means
// "generic label", a named one passes through, and each named key has its own
// baseline text. The baseline is what a reader sees before applyI18n runs, so a
// wrong fallback is visible to every user with the dictionary unavailable.
func TestNoteKeyFallbacks(t *testing.T) {
	t.Run("read-only keys", func(t *testing.T) {
		if got := roNoteKey(""); got != "ro.note" {
			t.Errorf("roNoteKey(\"\") = %q, want the generic ro.note", got)
		}
		if got := roNoteKey("ro.note.governance"); got != "ro.note.governance" {
			t.Errorf("roNoteKey passed through wrong: %q", got)
		}
		baselines := map[string]string{
			"":                    "read-only (runtime-managed)",
			"ro.note.governance":  "read-only — governance FROZEN (auto_apply stays false)",
			"ro.note.dead_config": "read-only — path fixed by the runtime (informational)",
		}
		seen := map[string]bool{}
		for key, want := range baselines {
			got := roNoteBaseline(key)
			if got != want {
				t.Errorf("roNoteBaseline(%q) = %q, want %q", key, got, want)
			}
			if seen[got] {
				t.Errorf("roNoteBaseline(%q) duplicates another baseline — the distinction is not visible", key)
			}
			seen[got] = true
		}
	})

	t.Run("raw view keys", func(t *testing.T) {
		if got := rawNoteKey(""); got != "raw.note" {
			t.Errorf("rawNoteKey(\"\") = %q, want the generic raw.note", got)
		}
		if got := rawNoteKey("raw.note.informational"); got != "raw.note.informational" {
			t.Errorf("rawNoteKey passed through wrong: %q", got)
		}
		generic := rawNoteBaseline("")
		informational := rawNoteBaseline("raw.note.informational")
		if generic == informational {
			t.Error("the informational raw-view baseline is identical to the generic one — the honesty label says nothing")
		}
		if !strings.Contains(informational, "not wired") {
			t.Errorf("the informational baseline does not state the value is unwired: %q", informational)
		}
	})
}

// TestNoteKeyBaselinesAreInDictionary closes the loop between the Go fallback
// and the shipped catalogue: every note key the widgets emit must resolve in all
// four locales, or the label renders as the English baseline forever.
func TestNoteKeyBaselinesAreInDictionary(t *testing.T) {
	dict := readEmbeddedAsset(t, "i18n.js")
	keys := []string{"ro.note", "raw.note"}
	for _, ro := range settings.ReadOnlyDisplayFields() {
		keys = append(keys, roNoteKey(ro.NoteKey))
	}
	for _, rb := range settings.RawViewBlocks() {
		keys = append(keys, rawNoteKey(rb.NoteKey))
	}
	for _, k := range keys {
		for _, loc := range []string{"en", "ko", "ja", "zh"} {
			if !i18nLocaleHasKey(t, dict, loc, k) {
				t.Errorf("note key %q is missing from locale %q", k, loc)
			}
		}
	}
}
