package settings

import (
	"slices"
	"sort"

	"testing"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/settings/yamlpatch"
	"github.com/modu-ai/moai-adk/internal/template"
)

// SPEC-V3R6-AUDIT-MODEL-PIN-001 M4 (REQ-AMP-009 / AC-AMP-008) — the four
// audit pin fields on the typed edit path: existence, widget closed sets,
// runtime-applied marking, and the seam write → config reload round-trip.

// auditPinField returns the named field, failing the test when absent.
func auditPinField(t *testing.T, name string) FieldDef {
	t.Helper()
	f, ok := Field(name)
	if !ok {
		t.Fatalf("field %q missing from the schema (the audit pin must be web-editable)", name)
	}
	return f
}

// optionValues returns the closed set of a select/radio field's option values
// (t1278 — the audit pins are radio groups now).
func optionValues(t *testing.T, f FieldDef) []string {
	t.Helper()
	if f.Type != TypeSelect && f.Type != TypeRadio {
		t.Fatalf("%s: Type = %v, want select or radio", f.Name, f.Type)
	}
	if len(f.Options) == 0 {
		t.Fatalf("%s: select carries no options", f.Name)
	}
	vals := make([]string, 0, len(f.Options))
	for _, o := range f.Options {
		vals = append(vals, o.Value)
	}
	sort.Strings(vals)
	return vals
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// TestAuditPinFields_ExistWithTypeAndPanel verifies the four fields exist with
// the intended widget types and closed sets (AC-AMP-008: glm effort select
// offers EXACTLY {low, high, max}; codex effort offers the 5-level Claude
// vocabulary; glm model offers the ValidGLMModels SSOT).
func TestAuditPinFields_ExistWithTypeAndPanel(t *testing.T) {
	t.Parallel()
	claudeModel := auditPinField(t, "workflow.audit.claude.model")
	if claudeModel.Type != TypeRadio || claudeModel.Section != SectionWorkflow {
		t.Errorf("workflow.audit.claude.model = {type:%v section:%v}, want radio/workflow (t1278)", claudeModel.Type, claudeModel.Section)
	}
	wantClaudeModels := append([]string{}, template.ModelAliasPickerValues()...)
	sort.Strings(wantClaudeModels)
	if gotClaudeModels := optionValues(t, claudeModel); !slices.Equal(gotClaudeModels, wantClaudeModels) {
		t.Errorf("workflow.audit.claude.model options = %v, want %v (launch model-alias SSOT)", gotClaudeModels, wantClaudeModels)
	}
	claudeEffort := auditPinField(t, "workflow.audit.claude.effort")
	wantClaudeEffort := append([]string{}, v4EffortValues()...)
	sort.Strings(wantClaudeEffort)
	if gotClaudeEffort := optionValues(t, claudeEffort); !slices.Equal(gotClaudeEffort, wantClaudeEffort) {
		t.Errorf("workflow.audit.claude.effort options = %v, want %v", gotClaudeEffort, wantClaudeEffort)
	}

	codexModel := auditPinField(t, "workflow.audit.codex.model")
	if codexModel.Type != TypeRadio {
		t.Errorf("workflow.audit.codex.model: Type = %v, want TypeRadio (t1278 — operator-confirmed closed set)", codexModel.Type)
	}
	wantCodexModels := append([]string{}, config.ValidCodexAuditModels()...)
	sort.Strings(wantCodexModels)
	if gotCodexModels := optionValues(t, codexModel); !slices.Equal(gotCodexModels, wantCodexModels) {
		t.Errorf("workflow.audit.codex.model options = %v, want %v (config.ValidCodexAuditModels SSOT)", gotCodexModels, wantCodexModels)
	}
	if codexModel.Section != SectionWorkflow {
		t.Errorf("workflow.audit.codex.model: Section = %v, want SectionWorkflow", codexModel.Section)
	}

	codexEffort := auditPinField(t, "workflow.audit.codex.effort")
	want := append([]string{}, v4EffortValues()...)
	sort.Strings(want)
	got := optionValues(t, codexEffort)
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("workflow.audit.codex.effort options = %v, want %v (the codex/Claude effort vocabulary)", got, want)
		}
	}

	glmModel := auditPinField(t, "workflow.audit.glm.model")
	wantGLM := append([]string{}, config.ValidGLMModels()...)
	sort.Strings(wantGLM)
	gotGLM := optionValues(t, glmModel)
	for i := range wantGLM {
		if gotGLM[i] != wantGLM[i] {
			t.Errorf("workflow.audit.glm.model options = %v, want %v (config.ValidGLMModels SSOT)", gotGLM, wantGLM)
		}
	}

	glmEffort := auditPinField(t, "workflow.audit.glm.effort")
	wantStates := append([]string{}, template.GLMReasoningStateNames()...)
	sort.Strings(wantStates) // {high, low, max}
	gotStates := optionValues(t, glmEffort)
	for i := range wantStates {
		if gotStates[i] != wantStates[i] {
			t.Errorf("workflow.audit.glm.effort options = %v, want EXACTLY %v (the z.ai state vocabulary, single reading)", gotStates, wantStates)
		}
	}
	// Negative: the Claude-only tokens must NOT be selectable on the GLM field.
	for _, forbidden := range []string{"medium", "xhigh"} {
		if containsString(gotStates, forbidden) {
			t.Errorf("workflow.audit.glm.effort offers %q — a Claude-only vocabulary value (REQ-AMP-006)", forbidden)
		}
	}

	// Effort fields are RUNTIME-applied (contrast with the stored-only tier
	// efforts, REQ-WCR-033 labeling discipline): StoreOnly must be false.
	for _, name := range []string{"workflow.audit.claude.effort", "workflow.audit.codex.effort", "workflow.audit.glm.effort"} {
		if f := auditPinField(t, name); f.StoreOnly {
			t.Errorf("%s: StoreOnly = true — audit pin efforts ARE runtime-applied (they ride the audit request builders); the stored-only marking belongs to the tier effort map, not this field", name)
		}
	}
}

// TestAuditPinFields_SeamRoundTrip saves values through the typed seam path
// and reloads them through the config.AuditConfig interpreter — the same
// struct the MCP audit resolvers unmarshal (no fork).
func TestAuditPinFields_SeamRoundTrip(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	seedSectionFixture(t, root, "workflow")

	edits := []yamlpatch.KeyEdit{
		{Path: []string{"workflow", "audit", "claude", "model"}, Value: "claude-opus-5-5"},
		{Path: []string{"workflow", "audit", "claude", "effort"}, Value: "medium"},
		{Path: []string{"workflow", "audit", "codex", "model"}, Value: "gpt-6.1-sol"},
		{Path: []string{"workflow", "audit", "codex", "effort"}, Value: "high"},
		{Path: []string{"workflow", "audit", "glm", "model"}, Value: "glm-5.3"},
		{Path: []string{"workflow", "audit", "glm", "effort"}, Value: "max"},
	}
	if err := WriteSectionViaSeam(root, "workflow", edits); err != nil {
		t.Fatalf("WriteSectionViaSeam: %v", err)
	}

	// Reload through the config interpreter (the typed read path — the same
	// AuditConfig struct the audit resolvers consume).
	raw := readSection(t, root, "workflow")
	wrapper := struct {
		Workflow struct {
			Audit config.AuditConfig `yaml:"audit"`
		} `yaml:"workflow"`
	}{}
	if err := yaml.Unmarshal([]byte(raw), &wrapper); err != nil {
		t.Fatalf("unmarshal written workflow.yaml: %v\n%s", err, raw)
	}
	audit := wrapper.Workflow.Audit
	if audit.Claude.Model != "claude-opus-5-5" || audit.Claude.Effort != "medium" {
		t.Errorf("claude pin round-trip: got %+v, want {claude-opus-5-5 medium}", audit.Claude)
	}
	if audit.Codex.Model != "gpt-6.1-sol" || audit.Codex.Effort != "high" {
		t.Errorf("codex pin round-trip: got %+v, want {gpt-6.1-sol high}", audit.Codex)
	}
	if audit.GLM.Model != "glm-5.3" || audit.GLM.Effort != "max" {
		t.Errorf("glm pin round-trip: got %+v, want {glm-5.3 max}", audit.GLM)
	}

	// Empty-value save persists empty strings → resolver falls back
	// (indistinguishable from absent sub-keys, §D.2 edge).
	if err := WriteSectionViaSeam(root, "workflow", []yamlpatch.KeyEdit{
		{Path: []string{"workflow", "audit", "glm", "model"}, Value: ""},
	}); err != nil {
		t.Fatalf("WriteSectionViaSeam (clear glm model): %v", err)
	}
	raw = readSection(t, root, "workflow")
	wrapper = struct {
		Workflow struct {
			Audit config.AuditConfig `yaml:"audit"`
		} `yaml:"workflow"`
	}{}
	if err := yaml.Unmarshal([]byte(raw), &wrapper); err != nil {
		t.Fatalf("unmarshal cleared workflow.yaml: %v", err)
	}
	if wrapper.Workflow.Audit.GLM.Model != "" {
		t.Errorf("cleared glm model = %q, want empty (web save with empty value persists empty strings)", wrapper.Workflow.Audit.GLM.Model)
	}
	if wrapper.Workflow.Audit.Codex.Model != "gpt-6.1-sol" {
		t.Errorf("codex pin lost by the glm clear edit: %q", wrapper.Workflow.Audit.Codex.Model)
	}
}
