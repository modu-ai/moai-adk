package web

// schema_parity_guard_test.go — SPEC-INIT-UPDATE-CONSISTENCY-001 REQ-ICU-005
// (F14 residual gap): the schema-driven form parser consumes
// settings.AllFields(), so a schema field that no panel renders would still
// be PARSED on every form save — a parse surface wider than the visible one.
// This guard pins the invariant: every editable FieldDef has a render surface
// or an explicit, rationale-carrying exemption.
//
// The render surface inventory (schemaSectionMetas() panels + the dedicated
// fieldsetMCP / codexAuthBlock components) and the exempt map below are the
// run-phase census this test encodes (card t588, 2026-09-14).

import (
	"sort"
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings"
)

// schemaRenderedFieldNames enumerates every FieldDef name the web console
// renders: the generic schemaSectionMetas() panels plus dedicated components
// that own their field enumeration — fieldsetMCP (SectionMCP fields) and the
// MCP tab's codexAuthBlock (the workflow.codex.* toggles, partitioned out of
// every workflow tab by isCodexToggleFieldName so they render exactly once).
func schemaRenderedFieldNames() map[string]bool {
	rendered := map[string]bool{}
	for _, m := range schemaSectionMetas() {
		for _, f := range m.Fields {
			rendered[f.Name] = true
		}
	}
	for _, f := range settings.SectionFields(settings.SectionMCP) {
		rendered[f.Name] = true
	}
	for _, f := range settings.AllFields() {
		if isCodexToggleFieldName(f.Name) {
			rendered[f.Name] = true
		}
	}
	return rendered
}

// schemaRenderExemptFields declares editable FieldDefs that intentionally have
// no web render surface. Each entry MUST carry a rationale — an exemption
// without evidence is exactly the silent parse/render gap this guard exists
// to close. The map is a census, not a dumping ground:
// TestSchemaParity_ExemptFieldsExist prunes it from the other direction.
//
// Census (card t588, tree 35141e3e9, 2026-09-14): no web panel ever rendered
// these groups, and granting seven new panels for hand-edit config keys is
// the over-engineering REQ-ICU-005's exemption path exists to avoid. All stay
// declared so parseSchemaForm's uniform AllFields() iteration and the
// seam write routes keep working; each is hand-editable in its section yaml.
var schemaRenderExemptFields = map[string]string{
	// harness.* — harness.yaml is the FROZEN-validation harness surface
	// (internal/config LoadHarnessConfig, HRN-001: error on absent file);
	// evaluator.memory_scope is additionally a FROZEN single-member value
	// (ErrEvalMemoryFrozen). Profile/mode_defaults changes belong to harness
	// tooling, not the console.
	"harness.default_profile":            "harness.yaml is a FROZEN validation surface (HRN-001); managed via harness tooling, not the console",
	"harness.effort_mapping.minimal":     "FROZEN harness.yaml knob (HRN-001) — see harness.default_profile",
	"harness.effort_mapping.standard":    "FROZEN harness.yaml knob (HRN-001) — see harness.default_profile",
	"harness.effort_mapping.thorough":    "FROZEN harness.yaml knob (HRN-001) — see harness.default_profile",
	"harness.auto_detection.enabled":     "FROZEN harness.yaml knob (HRN-001) — see harness.default_profile",
	"harness.escalation.enabled":         "FROZEN harness.yaml knob (HRN-001) — see harness.default_profile",
	"harness.escalation.max_escalations": "FROZEN harness.yaml knob (HRN-001) — see harness.default_profile",
	"harness.evaluator.memory_scope":     "FROZEN single-member value (ErrEvalMemoryFrozen rejects anything else) — no editable surface is possible",
	"harness.mode_defaults.cg":           "FROZEN harness.yaml knob (HRN-001) — see harness.default_profile",
	"harness.mode_defaults.solo":         "FROZEN harness.yaml knob (HRN-001) — see harness.default_profile",
	"harness.mode_defaults.team":         "FROZEN harness.yaml knob (HRN-001) — see harness.default_profile",

	// quality.* extras — quality.yaml typed keys declared in the schema bridge
	// for the seam writer; no console panel has ever rendered them.
	"quality.ddd_settings.behavior_snapshots":      "quality.yaml typed key, hand-edit config — no console panel (run-phase census, card t588)",
	"quality.ddd_settings.characterization_tests":  "quality.yaml typed key, hand-edit config — no console panel (run-phase census, card t588)",
	"quality.ddd_settings.preserve_before_improve": "quality.yaml typed key, hand-edit config — no console panel (run-phase census, card t588)",
	"quality.quality_extras_enabled":               "quality.yaml typed key, hand-edit config — no console panel (run-phase census, card t588)",

	// Seam-only sections whose FieldDefs exist for the uniform write route
	// (RouteForSection / yamlpatch seam) but whose panels were never built —
	// hand-edit config keys, per the census.
	"learning.enabled":                                  "seam-declared only; feedback/observability-class hand-edit key — no console panel (run-phase census, card t588)",
	"learning.log_retention_days":                       "seam-declared only; hand-edit key — no console panel (run-phase census, card t588)",
	"observability.enabled":                             "seam-declared only; hand-edit key — no console panel (run-phase census, card t588)",
	"observability.hook_metrics.slow_hook_threshold_ms": "seam-declared only; hand-edit key — no console panel (run-phase census, card t588)",
	"observability.max_file_size_mb":                    "seam-declared only; hand-edit key — no console panel (run-phase census, card t588)",
	"observability.report_dir":                          "seam-declared only; hand-edit key — no console panel (run-phase census, card t588)",
	"observability.retention_days":                      "seam-declared only; hand-edit key — no console panel (run-phase census, card t588)",
	"observability.trace_dir":                           "seam-declared only; hand-edit key — no console panel (run-phase census, card t588)",
	"security.permission.strict_mode":                   "seam-declared only; security posture stays hand-edit — no console panel (run-phase census, card t588)",
	"security.sandbox.docker_image":                     "seam-declared only; security posture stays hand-edit — no console panel (run-phase census, card t588)",
	"security.sandbox.required":                         "seam-declared only; security posture stays hand-edit — no console panel (run-phase census, card t588)",
	"ralph.lint_as_instruction":                         "seam-declared only; Ralph loop hand-edit key — no console panel (run-phase census, card t588)",
	"ralph.warn_as_instruction":                         "seam-declared only; Ralph loop hand-edit key — no console panel (run-phase census, card t588)",
	"cacheStrategy.enabled":                             "seam-declared only; cache hand-edit key — no console panel (run-phase census, card t588)",
	"cacheStrategy.session_ttl":                         "seam-declared only; cache hand-edit key — no console panel (run-phase census, card t588)",
	"handoff.mode":                                      "seam-declared only; handoff hand-edit key — no console panel (run-phase census, card t588)",
	"handoff.guide":                                     "seam-declared only; handoff hand-edit key — no console panel (run-phase census, card t588)",
}

// TestSchemaParity_EditableFieldsHaveRenderHome pins REQ-ICU-005: every
// editable FieldDef (seam or typed-section persist) is rendered by a panel or
// dedicated component, or is explicitly exempt with a rationale.
func TestSchemaParity_EditableFieldsHaveRenderHome(t *testing.T) {
	rendered := schemaRenderedFieldNames()

	var orphans []string
	for _, f := range settings.AllFields() {
		if !schemaEditableField(f) {
			continue
		}
		if rendered[f.Name] {
			continue
		}
		if _, exempt := schemaRenderExemptFields[f.Name]; exempt {
			continue
		}
		orphans = append(orphans, f.Name)
	}
	if len(orphans) > 0 {
		sort.Strings(orphans)
		t.Fatalf("editable schema fields with no render surface and no exemption (REQ-ICU-005): %v — grant a render home or declare an exemption with rationale", orphans)
	}
}

// TestSchemaParity_ExemptFieldsExist pins the census's other direction: an
// exemption entry whose field no longer exists (renamed, removed, or
// demoted to non-editable) is stale and must be pruned — the exempt map is a
// census, not a dumping ground.
func TestSchemaParity_ExemptFieldsExist(t *testing.T) {
	all := map[string]bool{}
	editable := map[string]bool{}
	for _, f := range settings.AllFields() {
		all[f.Name] = true
		if schemaEditableField(f) {
			editable[f.Name] = true
		}
	}
	rendered := schemaRenderedFieldNames()
	for name := range schemaRenderExemptFields {
		if !all[name] {
			t.Errorf("exemption entry %q names a field that no longer exists in AllFields() — prune the stale census entry", name)
			continue
		}
		if !editable[name] {
			t.Errorf("exemption entry %q names a non-editable field — the exemption is unnecessary (the guard skips non-editable fields anyway)", name)
		}
		if rendered[name] {
			t.Errorf("exemption entry %q is rendered anyway — remove the exemption or the render, not both", name)
		}
	}
}
