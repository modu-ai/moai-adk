package web

// widget_states_test.go — the display states each schema widget can be in.
//
// The console render tests exercise whichever state the live schema happens to
// produce, which leaves the states a user only reaches on a rejected save (field
// errors) or on an optional-value field (the empty option) unasserted. Those are
// the states that carry the most meaning: an unrendered error message is a save
// that silently did nothing, and a missing empty option removes the user's only
// way to say "leave this at the project default".

import (
	"context"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings"
)

func mustRender(t *testing.T, render func(*strings.Builder) error) string {
	t.Helper()
	var sb strings.Builder
	if err := render(&sb); err != nil {
		t.Fatalf("render: %v", err)
	}
	return sb.String()
}

// selectFixture builds a TypeSelect FieldDef, optionally with an empty option.
func selectFixture(emptyLabel, emptyKey string) settings.FieldDef {
	f := settings.FieldDef{
		Name:    "workflow.default_mode",
		Section: settings.SectionWorkflow,
		Type:    settings.TypeSelect,
		I18nKey: "f.workflow.default_mode",
		Options: []settings.OptionDef{
			{Value: "autopilot", I18nKey: "f.workflow.default_mode.opt.autopilot"},
			{Value: "loop", I18nKey: "f.workflow.default_mode.opt.loop"},
		},
		EmptyLabel:    emptyLabel,
		EmptyLabelKey: emptyKey,
	}
	return f
}

// TestSchemaSelectRowEmptyOptionSemantics pins the optional-value contract: a
// field carrying an EmptyLabel renders an empty-valued option, and that option
// is preselected when nothing is stored. Without it a user cannot express
// "unset" — the widget would force a value the config treats as a real choice.
func TestSchemaSelectRowEmptyOptionSemantics(t *testing.T) {
	f := selectFixture("(project default)", "opt.project_default")

	unset := mustRender(t, func(sb *strings.Builder) error {
		return schemaSelectRow(f, "", nil).Render(context.Background(), sb)
	})
	if !strings.Contains(unset, `<option value="" selected`) {
		t.Errorf("an unset optional field did not preselect the empty option:\n%s", unset)
	}
	if !strings.Contains(unset, `data-i18n="opt.project_default"`) {
		t.Error("the empty option lost its i18n key")
	}

	set := mustRender(t, func(sb *strings.Builder) error {
		return schemaSelectRow(f, "loop", nil).Render(context.Background(), sb)
	})
	if strings.Contains(set, `<option value="" selected`) {
		t.Error("a stored value did not take selection away from the empty option")
	}
	if !strings.Contains(set, `<option value="loop" selected`) {
		t.Errorf("the stored value was not preselected:\n%s", set)
	}
}

// TestSchemaSelectRowOmitsEmptyOptionWhenRequired is the other half: a field
// with no EmptyLabel is a required choice, so no empty option is offered.
func TestSchemaSelectRowOmitsEmptyOptionWhenRequired(t *testing.T) {
	f := selectFixture("", "")
	html := mustRender(t, func(sb *strings.Builder) error {
		return schemaSelectRow(f, "autopilot", nil).Render(context.Background(), sb)
	})
	if strings.Contains(html, `<option value=""`) {
		t.Errorf("a required select offered an empty option:\n%s", html)
	}
}

// TestSchemaSelectRowSurfacesFieldError pins the rejected-save display for the
// select widget.
func TestSchemaSelectRowSurfacesFieldError(t *testing.T) {
	f := selectFixture("", "")
	html := mustRender(t, func(sb *strings.Builder) error {
		return schemaSelectRow(f, "bogus", map[string]string{f.Name: "invalid option"}).
			Render(context.Background(), sb)
	})
	if !strings.Contains(html, "has-error") || !strings.Contains(html, "invalid option") {
		t.Errorf("select row did not surface the field error:\n%s", html)
	}
}

// radioFixture builds a TypeRadio FieldDef, optionally with per-option
// descriptions (which switch the group to the stacked layout).
func radioFixture(withDesc bool, emptyLabel string) settings.FieldDef {
	opts := []settings.OptionDef{
		{Value: "manual", I18nKey: "f.git_strategy.mode.opt.manual"},
		{Value: "team", I18nKey: "f.git_strategy.mode.opt.team"},
	}
	if withDesc {
		opts[0].OptionDesc = "fieldDesc.git_strategy.mode.option.manual"
	}
	return settings.FieldDef{
		Name:       "git_strategy.mode",
		Section:    settings.SectionGitStrategy,
		Type:       settings.TypeRadio,
		I18nKey:    "f.git_strategy.mode",
		Options:    opts,
		EmptyLabel: emptyLabel,
	}
}

// TestSchemaRadioRowStackedLayoutOptIn pins the layout opt-in: a group whose
// options carry descriptions renders stacked so each description sits beside its
// option. Inline layout with descriptions would run them together.
func TestSchemaRadioRowStackedLayoutOptIn(t *testing.T) {
	plain := mustRender(t, func(sb *strings.Builder) error {
		return schemaRadioRow(radioFixture(false, ""), "manual", nil).Render(context.Background(), sb)
	})
	stacked := mustRender(t, func(sb *strings.Builder) error {
		return schemaRadioRow(radioFixture(true, ""), "manual", nil).Render(context.Background(), sb)
	})

	if strings.Contains(plain, "radio-group--stacked") {
		t.Error("a group with no option descriptions used the stacked layout")
	}
	if !strings.Contains(stacked, "radio-group--stacked") {
		t.Errorf("a group WITH option descriptions did not opt into the stacked layout:\n%s", stacked)
	}
}

// TestSchemaRadioRowEmptyOptionAndError covers the radio widget's empty option
// and error state, the same two contracts asserted for select above.
func TestSchemaRadioRowEmptyOptionAndError(t *testing.T) {
	f := radioFixture(false, "(project default)")
	unset := mustRender(t, func(sb *strings.Builder) error {
		return schemaRadioRow(f, "", nil).Render(context.Background(), sb)
	})
	if !strings.Contains(unset, `value="" checked`) {
		t.Errorf("an unset optional radio group did not check the empty option:\n%s", unset)
	}

	withErr := mustRender(t, func(sb *strings.Builder) error {
		return schemaRadioRow(f, "team", map[string]string{f.Name: "not allowed here"}).
			Render(context.Background(), sb)
	})
	if !strings.Contains(withErr, "has-error") || !strings.Contains(withErr, "not allowed here") {
		t.Errorf("radio row did not surface the field error:\n%s", withErr)
	}
	if !strings.Contains(withErr, `value="team" checked`) {
		t.Error("a rejected radio submission lost the user's selection")
	}
}

// TestSchemaNumberRowStatesAndStep pins the number widget: the step attribute
// distinguishes an integer field from a float one, and the error state renders.
func TestSchemaNumberRowStatesAndStep(t *testing.T) {
	f := settings.FieldDef{
		Name:    "workflow.loop_prevention.max_iterations",
		Section: settings.SectionWorkflow,
		Type:    settings.TypeInt,
		I18nKey: "f.workflow.loop_prevention.max_iterations",
	}
	intRow := mustRender(t, func(sb *strings.Builder) error {
		return schemaNumberRow(f, "10", "1", nil).Render(context.Background(), sb)
	})
	if !strings.Contains(intRow, `step="1"`) || !strings.Contains(intRow, `value="10"`) {
		t.Errorf("integer row lost its step or value:\n%s", intRow)
	}
	if !strings.Contains(intRow, `type="number"`) {
		t.Error("number row is not a number input")
	}

	floatRow := mustRender(t, func(sb *strings.Builder) error {
		return schemaNumberRow(f, "0.85", "0.01", nil).Render(context.Background(), sb)
	})
	if !strings.Contains(floatRow, `step="0.01"`) {
		t.Errorf("float row did not carry the fractional step:\n%s", floatRow)
	}

	errRow := mustRender(t, func(sb *strings.Builder) error {
		return schemaNumberRow(f, "abc", "1", map[string]string{f.Name: "must be an integer"}).
			Render(context.Background(), sb)
	})
	if !strings.Contains(errRow, "has-error") || !strings.Contains(errRow, "must be an integer") {
		t.Errorf("number row did not surface the field error:\n%s", errRow)
	}
}

// TestSchemaToggleRowBothStates pins the bool widget introduced in M3: both
// options render, the stored state is the checked one, and the hidden
// __present companion is emitted in either state (it is what lets the parser
// tell "explicitly false" from "not submitted").
func TestSchemaToggleRowBothStates(t *testing.T) {
	f := settings.FieldDef{
		Name:    "workflow.branch_guard.enabled",
		Section: settings.SectionWorkflow,
		Type:    settings.TypeBool,
		I18nKey: "f.workflow.branch_guard.enabled",
	}
	for _, tc := range []struct {
		name string
		on   bool
	}{{"on", true}, {"off", false}} {
		t.Run(tc.name, func(t *testing.T) {
			html := mustRender(t, func(sb *strings.Builder) error {
				return schemaToggleRow(f, tc.on).Render(context.Background(), sb)
			})
			if n := strings.Count(html, `type="radio"`); n != 2 {
				t.Errorf("bool field rendered %d radio inputs, want 2", n)
			}
			if !strings.Contains(html, `name="`+f.Name+`__present"`) {
				t.Error("the hidden __present companion is missing")
			}
			if strings.Contains(html, `type="checkbox"`) {
				t.Error("bool field regressed to a checkbox")
			}
			if n := strings.Count(html, "checked"); n != 1 {
				t.Errorf("%d options are checked, want exactly 1", n)
			}
		})
	}
}

// TestSchemaFieldWidgetDispatch pins the type→widget routing. The dispatch is
// the single place a new FieldType could silently fall through to the free-text
// default, which would render a closed set as an open input — the exact
// dishonesty M3 set out to remove.
func TestSchemaFieldWidgetDispatch(t *testing.T) {
	cases := []struct {
		typ      settings.FieldType
		wantMark string
	}{
		{settings.TypeBool, `type="radio"`},
		{settings.TypeSelect, "<select"},
		{settings.TypeRadio, `type="radio"`},
		{settings.TypeInt, `type="number"`},
		{settings.TypeFloat, `type="number"`},
		{settings.TypeText, `type="text"`},
	}
	for _, tc := range cases {
		t.Run(string(tc.typ), func(t *testing.T) {
			f := settings.FieldDef{
				Name: "fixture.field", Section: settings.SectionWorkflow,
				Type: tc.typ, I18nKey: "f.fixture.field",
				Options: []settings.OptionDef{{Value: "a", I18nKey: "f.fixture.field.opt.a"}},
			}
			html := mustRender(t, func(sb *strings.Builder) error {
				return schemaFieldWidget(pageView{}, f).Render(context.Background(), sb)
			})
			if !strings.Contains(html, tc.wantMark) {
				t.Errorf("FieldType %q did not route to a widget containing %s:\n%s", tc.typ, tc.wantMark, html)
			}
		})
	}
}

// TestFieldDescriptionRendersOnlyWhenSet pins the optional helper paragraph: a
// field with no Description must not emit an empty node the i18n pass would
// then try to fill.
func TestFieldDescriptionRendersOnlyWhenSet(t *testing.T) {
	with := mustRender(t, func(sb *strings.Builder) error {
		return fieldDescription(settings.FieldDef{Description: "fieldDesc.x.y"}).Render(context.Background(), sb)
	})
	if !strings.Contains(with, `data-i18n="fieldDesc.x.y"`) {
		t.Errorf("a field with a Description did not render it: %s", with)
	}
	without := mustRender(t, func(sb *strings.Builder) error {
		return fieldDescription(settings.FieldDef{}).Render(context.Background(), sb)
	})
	if strings.TrimSpace(without) != "" {
		t.Errorf("a field with no Description emitted markup: %q", without)
	}
}
