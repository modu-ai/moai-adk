package web

// fieldsets_states_test.go — TG-3 (SPEC-WEB-CONSOLE-018 REQ-007).
//
// Defect answer: these tests catch field-rendering regressions — an error state
// rendering without its message, a select losing its selected option, a toggle
// emitting the wrong state (absent-default polarity), a secret-class control
// echoing a stored key, or a panel omitting its note/read-only/raw sub-surfaces.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	mcpcat "github.com/modu-ai/moai-adk/internal/mcp"
	"github.com/modu-ai/moai-adk/internal/settings"
	"github.com/modu-ai/moai-adk/internal/settings/agentfm"
)

func tg3Errs(pairs ...string) map[string]string {
	m := map[string]string{}
	for i := 0; i+1 < len(pairs); i += 2 {
		m[pairs[i]] = pairs[i+1]
	}
	return m
}

// TestSchemaTextRowErrorState pins the text row's rejected-save branch: the
// errored render carries aria-invalid and the error span with the message; the
// clean render carries neither.
func TestSchemaTextRowErrorState(t *testing.T) {
	f := settings.FieldDef{Name: "fixture.text", Section: settings.SectionWorkflow, Type: settings.TypeText, I18nKey: "f.fixture.text"}
	clean := renderTempl(t, schemaTextRow(f, "stored", nil))
	if strings.Contains(clean, "aria-invalid") || strings.Contains(clean, "field__err") {
		t.Errorf("clean text row rendered error chrome:\n%s", clean)
	}
	if !strings.Contains(clean, `value="stored"`) {
		t.Errorf("clean text row lost its value:\n%s", clean)
	}
	errored := renderTempl(t, schemaTextRow(f, "submitted", tg3Errs("fixture.text", "too long")))
	for _, want := range []string{`aria-invalid="true"`, `aria-describedby="err_fixture.text"`, `field__err`, `too long`} {
		if !strings.Contains(errored, want) {
			t.Errorf("errored text row missing %q:\n%s", want, errored)
		}
	}
}

// TestSchemaSelectUnlistedCurrentValue pins the RC2 passthrough contract: a
// persisted value outside the offered set gets ONE synthetic selected option
// carrying the exact raw value — otherwise the browser auto-selects the first
// option and any save silently rewrites the stored value.
func TestSchemaSelectUnlistedCurrentValue(t *testing.T) {
	f := settings.FieldDef{
		Name: "fixture.closed", Section: settings.SectionWorkflow, Type: settings.TypeSelect, I18nKey: "f.fixture.closed",
		Options: []settings.OptionDef{{Value: "a", I18nKey: "f.fixture.closed.opt.a"}, {Value: "b", I18nKey: "f.fixture.closed.opt.b"}},
	}
	html := renderTempl(t, schemaSelectRow(f, "zzz-legacy", nil))
	if !strings.Contains(html, `<option value="zzz-legacy" selected>zzz-legacy (saved)</option>`) {
		t.Errorf("an unlisted persisted value lost its synthetic selected option:\n%s", html)
	}
	if n := strings.Count(html, `value="zzz-legacy"`); n != 1 {
		t.Errorf("synthetic option rendered %d times, want 1:\n%s", n, html)
	}
}

// TestSchemaSelectEffortHintSuffix pins the effort-field hint: a field whose
// name ends in .effort declares that its value resolves from the performance
// tier — dropping the hint invites per-field edits that look broken.
func TestSchemaSelectEffortHintSuffix(t *testing.T) {
	f := settings.FieldDef{
		Name: "agentfm.manager-spec.effort", Section: settings.SectionWorkflow, Type: settings.TypeSelect, I18nKey: "f.x",
		Options: []settings.OptionDef{{Value: "high", I18nKey: "f.x.opt.high"}},
	}
	with := renderTempl(t, schemaSelectRow(f, "high", nil))
	if !strings.Contains(with, `data-i18n="hint.effort.go_unbound"`) {
		t.Errorf("an effort field lost its tier-resolution hint:\n%s", with)
	}
	other := renderTempl(t, schemaSelectRow(settings.FieldDef{Name: "fixture.plain", Type: settings.TypeSelect, Options: f.Options}, "high", nil))
	if strings.Contains(other, "hint.effort.go_unbound") {
		t.Errorf("a non-effort field rendered the effort hint:\n%s", other)
	}
}

// TestSchemaFieldWidgetAbsentDefaultPolarity pins the bool dispatch: a field
// declared default-ON shows checked while its key is absent from the stored
// map — an == "true" reading here drew absent as OFF and recorded enabled:false
// on the untouched-rendered save.
func TestSchemaFieldWidgetAbsentDefaultPolarity(t *testing.T) {
	f := settings.FieldDef{Name: "fixture.defaulton", Section: settings.SectionWorkflow, Type: settings.TypeBool, I18nKey: "f.fixture.defaulton", AbsentDefault: "true"}
	html := renderTempl(t, schemaFieldWidget(pageView{SchemaValues: map[string]string{}}, f))
	if !strings.Contains(html, `id="fixture.defaulton--on" name="fixture.defaulton" value="1" checked`) {
		t.Errorf("an absent key on a default-ON field did not render checked:\n%s", html)
	}
	off := renderTempl(t, schemaFieldWidget(pageView{SchemaValues: map[string]string{"fixture.defaulton": "false"}}, f))
	if !strings.Contains(off, `id="fixture.defaulton--off" name="fixture.defaulton" value="" checked`) {
		t.Errorf("an explicit false did not render the off branch:\n%s", off)
	}
}

// TestSchemaReadOnlyRowAndRawBlock pins the two non-input sub-surfaces: a
// read-only row renders the value with no form control, and a raw block renders
// its collapsed <details> content.
func TestSchemaReadOnlyRowAndRawBlock(t *testing.T) {
	ro := renderTempl(t, schemaReadOnlyRow("fixture.ro", "on-disk value", "ro.note.runtime"))
	for _, want := range []string{`field--ro`, `fixture.ro`, `on-disk value`} {
		if !strings.Contains(ro, want) {
			t.Errorf("read-only row missing %q:\n%s", want, ro)
		}
	}
	if strings.Contains(ro, "<input") || strings.Contains(ro, "<select") {
		t.Errorf("a read-only row rendered a form control:\n%s", ro)
	}
	raw := renderTempl(t, schemaRawBlock("fixture.raw", "key: value", "raw.note"))
	if !strings.Contains(raw, `<pre class="raw__pre">key: value</pre>`) {
		t.Errorf("raw block lost its content:\n%s", raw)
	}
}

// TestSettingsPanelDensityAndCount pins the panel chrome: the density derives
// from the schema field count (≥7 → dense), and the count label renders the
// derived number.
func TestSettingsPanelDensityAndCount(t *testing.T) {
	dense := renderTempl(t, settingsPanel(panelChrome{Icon: "check", TitleKey: "sec.x.title", Title: "X", DescKey: "sec.x.desc", Desc: "d", Count: 7}))
	roomy := renderTempl(t, settingsPanel(panelChrome{Icon: "check", TitleKey: "sec.x.title", Title: "X", DescKey: "sec.x.desc", Desc: "d", Count: 3}))
	if !strings.Contains(dense, `data-density="dense"`) || !strings.Contains(roomy, `data-density="roomy"`) {
		t.Errorf("density did not follow the field count:\ndense:\n%s\nroomy:\n%s", dense, roomy)
	}
	if !strings.Contains(dense, ">7 <") || !strings.Contains(roomy, ">3 <") {
		t.Errorf("the count label lost its number:\n%s", dense)
	}
}

// TestFieldsetSchemaSectionNote pins the panel-level note: a meta carrying a
// note key renders the note once in the header; a meta without one renders no
// note paragraph.
func TestFieldsetSchemaSectionNote(t *testing.T) {
	meta := schemaPanelMeta("gate")
	meta.NoteKey = "sec.gate.note"
	meta.Note = "watch out"
	with := renderTempl(t, fieldsetSchemaSection(pageView{}, meta))
	if !strings.Contains(with, `data-i18n="sec.gate.note"`) || !strings.Contains(with, "watch out") {
		t.Errorf("a panel note key did not render:\n%s", with)
	}
	without := renderTempl(t, fieldsetSchemaSection(pageView{}, schemaPanelMeta("gate")))
	if strings.Contains(without, "watch out") {
		t.Errorf("a meta with no note rendered one:\n%s", without)
	}
}

// TestFieldsetSchemaSectionLLMGlmKey pins the GLM key fieldset rendered through
// the LLM panel: the secret-class input is present with an unconditionally
// empty value, and the configured state adds the bounded hint plus the reveal
// control — which the unconfigured state must not render.
func TestFieldsetSchemaSectionLLMGlmKey(t *testing.T) {
	unconfigured := renderTempl(t, fieldsetSchemaSection(pageView{}, schemaPanelMeta("llm")))
	if !strings.Contains(unconfigured, `name="glm_api_key"`) {
		t.Errorf("the LLM panel lost the GLM key field:\n%s", unconfigured)
	}
	if !strings.Contains(unconfigured, `value=""`) {
		t.Errorf("the GLM key input must carry an unconditionally empty value:\n%s", unconfigured)
	}
	if strings.Contains(unconfigured, "glmKeyReveal") {
		t.Errorf("a reveal button rendered with no stored key:\n%s", unconfigured)
	}

	configured := renderTempl(t, fieldsetSchemaSection(pageView{GLMKeyConfigured: true, GLMKeyHint: "abcd"}, schemaPanelMeta("llm")))
	for _, want := range []string{`f.glm_api_key.configured`, `…abcd`, `data-reveal-url="/glm-key/reveal"`} {
		if !strings.Contains(configured, want) {
			t.Errorf("configured GLM key state missing %q:\n%s", want, configured)
		}
	}

	errored := renderTempl(t, fieldsetSchemaSection(pageView{FieldErrors: tg3Errs("glm_api_key", "line break")}, schemaPanelMeta("llm")))
	if !strings.Contains(errored, `aria-invalid="true"`) || !strings.Contains(errored, "line break") {
		t.Errorf("a GLM key error lost its message:\n%s", errored)
	}
}

// TestFieldsetSchemaSectionWorkflowJev pins the Jev sub-section inside the
// workflow panel: the privacy note, the credential input, and the section
// marker render — and the Jev credential has no reveal control by construction.
func TestFieldsetSchemaSectionWorkflowJev(t *testing.T) {
	html := renderTempl(t, fieldsetSchemaSection(pageView{JevKeyConfigured: true, JevKeyHint: "wxyz"}, schemaPanelMeta("workflow")))
	for _, want := range []string{
		`data-section="jev"`,          // the sub-section marker
		`data-i18n="sec.jev.note"`,    // the privacy statement
		`name="jev_api_key"`,          // the credential input
		`f.jev_api_key.configured`,    // the configured indicator
		`…wxyz`,                       // the bounded trailing-four hint
	} {
		if !strings.Contains(html, want) {
			t.Errorf("workflow panel Jev section missing %q:\n%s", want, html)
		}
	}
	if strings.Contains(html, "data-reveal-url") {
		t.Errorf("the Jev credential must not carry a reveal control:\n%s", html)
	}
}

// TestPermissionOptionStates pins the permission_mode option branches: the
// empty option renders "(project default)" and is selected exactly when the
// current value is empty; a real option is selected only on its own value.
func TestPermissionOptionStates(t *testing.T) {
	opts := []settings.OptionDef{{Value: "", I18nKey: "opt.project_default"}, {Value: "yolo", I18nKey: "f.permission_mode.opt.yolo"}}

	html := renderTempl(t, permissionOption(opts[0], ""))
	if !strings.Contains(html, `<option value="" selected data-i18n="opt.project_default">`) {
		t.Errorf("an empty current value did not select the project-default option:\n%s", html)
	}
	html = renderTempl(t, permissionOption(opts[0], "yolo"))
	if strings.Contains(html, "selected") {
		t.Errorf("the empty option stayed selected over a stored value:\n%s", html)
	}
	html = renderTempl(t, permissionOption(opts[1], "yolo"))
	if !strings.Contains(html, `<option value="yolo" selected`) {
		t.Errorf("the stored value was not marked selected:\n%s", html)
	}
}

// TestAgentFMRowStates pins the agentfm row: a parsed agent renders both
// selects with the resolved selection, a haiku-resolved agent disables its
// effort select, and a parse-failed agent renders the unavailable row instead
// of editable selects that would submit garbage.
func TestAgentFMRowStates(t *testing.T) {
	parsed := renderTempl(t, agentFMRow(agentfmAgentInfo("manager-spec", true), configLLMZero(), nil))
	for _, want := range []string{`data-agent-row="manager-spec"`, `agentfm.manager-spec.model`, `agentfm.manager-spec.effort`} {
		if !strings.Contains(parsed, want) {
			t.Errorf("parsed agent row missing %q:\n%s", want, parsed)
		}
	}
	if strings.Contains(parsed, `disabled`) {
		t.Errorf("a non-haiku agent's effort select rendered disabled:\n%s", parsed)
	}

	haiku := renderTempl(t, agentFMRow(agentfmAgentInfo("manager-spec", true),
		configLLMOverride("manager-spec", "haiku", "low"), nil))
	if !strings.Contains(haiku, `disabled`) || !strings.Contains(haiku, `data-haiku-hint`) {
		t.Errorf("a haiku-resolved agent lost its disabled effort select:\n%s", haiku)
	}

	failed := renderTempl(t, agentFMRow(agentfmAgentInfo("broken-agent", false), configLLMZero(), nil))
	if !strings.Contains(failed, "unavailable (frontmatter parse failed)") {
		t.Errorf("a parse-failed agent lost its unavailable row:\n%s", failed)
	}
	if strings.Contains(failed, "<select") {
		t.Errorf("a parse-failed agent rendered editable selects:\n%s", failed)
	}
}

// TestMCPToggleRowWriteCapableBadge pins the write-capable badge: a
// write-capable tool's row carries the text badge, a read-only tool's does not
// — the distinction rides on the literal string, not colour.
func TestMCPToggleRowWriteCapableBadge(t *testing.T) {
	var writeTool, readOnlyTool settings.FieldDef
	for _, tool := range mcpcat.MoaiMCPTools() {
		name := "mcp.tools." + tool.Name + ".enabled"
		f := settings.FieldDef{Name: name, Section: settings.SectionMCP, Type: settings.TypeBool, I18nKey: "f." + name}
		if tool.WriteCapable && writeTool.Name == "" {
			writeTool = f
		}
		if !tool.WriteCapable && readOnlyTool.Name == "" {
			readOnlyTool = f
		}
	}
	if writeTool.Name == "" || readOnlyTool.Name == "" {
		t.Skip("catalog has no write/read pair to assert against")
	}
	with := renderTempl(t, mcpToolRow(writeTool, true))
	if !strings.Contains(with, "Write-capable") {
		t.Errorf("a write-capable tool lost its badge:\n%s", with)
	}
	without := renderTempl(t, mcpToolRow(readOnlyTool, true))
	if strings.Contains(without, "Write-capable") {
		t.Errorf("a read-only tool rendered the write-capable badge:\n%s", without)
	}
}

// TestCodexAuthBlockStates pins the codex authentication surface: installed
// with an unknown provider names the login command, a not-installed probe
// renders the graceful state, and both carry the two opt-in toggles.
func TestCodexAuthBlockStates(t *testing.T) {
	unknown := renderTempl(t, codexAuthBlock(pageView{CodexState: CodexStateView{Installed: true, Binary: "/usr/bin/codex", Version: "1.2.3", AuthProvider: codexAuthUnknown}}))
	for _, want := range []string{`/usr/bin/codex`, `1.2.3`, `f.mcp.codex.login_hint`, `codex login`, `workflow.codex.review_gate.enabled`, `workflow.codex.task.allow_write`} {
		if !strings.Contains(unknown, want) {
			t.Errorf("installed-unknown codex state missing %q:\n%s", want, unknown)
		}
	}
	absent := renderTempl(t, codexAuthBlock(pageView{}))
	if !strings.Contains(absent, `f.mcp.codex.not_installed`) {
		t.Errorf("an absent probe lost the not-installed state:\n%s", absent)
	}
}

// TestCodexMirrorRowViewStates pins the read-only mirror row: an unset value
// renders the unset placeholder (not an empty control), a set value renders
// verbatim, and the declared exception carries its shared-backend note.
func TestCodexMirrorRowViewStates(t *testing.T) {
	unset := renderTempl(t, codexMirrorRowView(codexMirrorRow{Name: "workflow.codex.review_gate.enabled", OwnerTab: codexOwnerTabMCP, LabelKey: "f.workflow.codex.review_gate.enabled.title"}))
	if !strings.Contains(unset, `data-i18n="tab.codex.value.unset"`) {
		t.Errorf("an unset mirror row lost its unset placeholder:\n%s", unset)
	}
	set := renderTempl(t, codexMirrorRowView(codexMirrorRow{Name: "workflow.audit.model", Value: "codex", OwnerTab: codexOwnerTabAudit, LabelKey: "f.workflow.audit.model.title", NoteKey: codexSharedBackendI18nKey}))
	if !strings.Contains(set, `codex`) || !strings.Contains(set, codexSharedBackendI18nKey) {
		t.Errorf("the declared exception lost its value or shared-backend note:\n%s", set)
	}
	if strings.Contains(unset+set, "<input") {
		t.Errorf("the read-only mirror rendered a form control:\n%s\n%s", unset, set)
	}
}

// TestCodexProbeBlockStates pins the probe block: an installed probe renders
// binary/version/provider, an absent probe renders the not-installed help.
func TestCodexProbeBlockStates(t *testing.T) {
	installed := renderTempl(t, codexProbeBlock(pageView{CodexState: CodexStateView{Installed: true, Binary: "/usr/bin/codex", Version: "1.2.3", AuthProvider: codexAuthChatGPT}}))
	for _, want := range []string{`/usr/bin/codex`, `1.2.3`, `ChatGPT`} {
		if !strings.Contains(installed, want) {
			t.Errorf("installed probe missing %q:\n%s", want, installed)
		}
	}
	absent := renderTempl(t, codexProbeBlock(pageView{}))
	if !strings.Contains(absent, `f.mcp.codex.not_installed`) {
		t.Errorf("an absent probe lost the not-installed state:\n%s", absent)
	}
}

// TestGlmKeyStateBlockStates pins the MCP-section key STATE surface: the
// configured state carries the bounded hint, the unconfigured state says so,
// and neither renders an input (the credential input lives in the LLM section).
func TestGlmKeyStateBlockStates(t *testing.T) {
	configured := renderTempl(t, glmKeyStateBlock(pageView{GLMKeyConfigured: true, GLMKeyHint: "abcd"}))
	for _, want := range []string{`f.mcp.glm_key.configured`, `…abcd`} {
		if !strings.Contains(configured, want) {
			t.Errorf("configured key state missing %q:\n%s", want, configured)
		}
	}
	unconfigured := renderTempl(t, glmKeyStateBlock(pageView{}))
	if !strings.Contains(unconfigured, `f.mcp.glm_key.not_configured`) {
		t.Errorf("unconfigured key state lost its message:\n%s", unconfigured)
	}
	if strings.Contains(configured+unconfigured, "<input") {
		t.Errorf("the key state block rendered an input:\n%s", configured)
	}
}

// TestCodexToggleRowCheckedBranches pins the codex opt-in toggle: checked and
// unchecked renders select the matching segment radio, and the hidden
// companion is present in both.
func TestCodexToggleRowCheckedBranches(t *testing.T) {
	on := renderTempl(t, codexToggleRow("workflow.codex.review_gate.enabled", true))
	off := renderTempl(t, codexToggleRow("workflow.codex.review_gate.enabled", false))
	if !strings.Contains(on, `id="workflow.codex.review_gate.enabled--on" name="workflow.codex.review_gate.enabled" value="1" checked`) {
		t.Errorf("checked toggle lost its on-segment:\n%s", on)
	}
	if !strings.Contains(off, `id="workflow.codex.review_gate.enabled--off" name="workflow.codex.review_gate.enabled" value="" checked`) {
		t.Errorf("unchecked toggle lost its off-segment:\n%s", off)
	}
	for _, html := range []string{on, off} {
		if !strings.Contains(html, `name="workflow.codex.review_gate.enabled__present"`) {
			t.Errorf("toggle lost its hidden companion:\n%s", html)
		}
	}
}

// tg3 helpers build the agentfm fixtures without touching disk: agentFMRow
// reads only the AgentInfo fields and the resolved profile matrix.

func agentfmAgentInfo(name string, parseOK bool) agentfm.AgentInfo {
	return agentfm.AgentInfo{Name: name, Path: "/agents/" + name + ".md", ParseOK: parseOK}
}

func configLLMZero() config.LLMConfig { return config.LLMConfig{} }

// configLLMOverride pins one agent's resolved model/effort through the
// override slot the profile matrix consults first.
func configLLMOverride(agent, model, effort string) config.LLMConfig {
	return config.LLMConfig{AgentOverrides: map[string]config.ModelEffort{agent: {Model: model, Effort: effort}}}
}
