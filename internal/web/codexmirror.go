package web

// codexmirror.go — SPEC-WEB-CODEX-PANEL-001: the row model behind the codex
// tab, a READ-ONLY MIRROR of the codex settings that live on the Audit and MCP
// tabs.
//
// Nothing here moves a field. partitionWorkflowFields, SectionFields(SectionMCP)
// and every existing panel are untouched: each mirrored field stays declared,
// rendered and editable on its owning tab, and this file only decides what the
// mirror displays and where it points (REQ-WCP-004).
//
// The rows are DERIVED by predicate over settings.AllFields() plus the shared
// MCP tool catalogue, never hand-listed — a hand list drifts the moment a codex
// field is added, and the drift is silent (REQ-WCP-005). The single exception is
// workflow.audit.model, which carries no codex token and which no predicate can
// therefore reach; it is named once, here, and asserted separately.

import (
	"strings"

	mcpcat "github.com/modu-ai/moai-adk/internal/mcp"
	"github.com/modu-ai/moai-adk/internal/settings"
)

const (
	// codexOwnerTabAudit / codexOwnerTabMCP are the panel ids of the tabs that
	// own the editing surface of a mirrored field. The mirror links to them; it
	// never renders their controls.
	codexOwnerTabAudit = "audit"
	codexOwnerTabMCP   = "mcp"

	// codexAuditModelField is the ONE declared exception (spec.md §C.1). It is
	// the shared audit backend selector — it answers "is codex the audit backend
	// on this project?", the first question a reader of this panel has — but it
	// carries no codex token, so no predicate reaches it. Naming it here, once,
	// keeps the derivation invariant testable apart from this judgement.
	codexAuditModelField = "workflow.audit.model"

	// codexSharedBackendI18nKey labels that exception in the panel as the shared
	// audit backend selector rather than as a codex-owned setting.
	codexSharedBackendI18nKey = "tab.codex.shared_backend"

	// codexMirrorUnsetI18nKey is the placeholder for a field with no value on
	// disk. The mirror shows what is on disk; it does not compute an effective
	// value, because computing one here would make the panel a second classifier.
	codexMirrorUnsetI18nKey = "tab.codex.value.unset"

	// codexPanelIcon must match an existing case in icons.templ. A name with no
	// case renders nothing and fails no test, so this is deliberately one of the
	// names that exists (the audit panel uses the same one).
	codexPanelIcon = "check-circle"
)

// codexPanelI18nKeys are the dictionary keys this panel introduces. Each must
// carry an entry in all four locale blocks of assets/i18n.js (REQ-WCP-008); the
// per-row labels reuse the existing f.<field>.title keys and add none.
var codexPanelI18nKeys = []string{
	"tab.codex.title",
	"tab.codex.desc",
	"tab.codex.readonly",
	"tab.codex.group.audit",
	"tab.codex.group.optin",
	"tab.codex.group.mcp",
	"tab.codex.group.mcp.help",
	"tab.codex.group.probe",
	"tab.codex.edit_on.audit",
	"tab.codex.edit_on.mcp",
	codexSharedBackendI18nKey,
	codexMirrorUnsetI18nKey,
}

// isCodexMCPToolField reports whether an MCP enablement field names a codex
// tool, resolved through the shared catalogue (internal/mcp MoaiMCPTools) rather
// than through a second tool list. A tool added to the catalogue is mirrored
// without an edit here.
func isCodexMCPToolField(fieldName string) bool {
	tool, ok := mcpToolNameFromField(fieldName)
	if !ok {
		return false
	}
	for _, t := range mcpcat.MoaiMCPTools() {
		if t.Name == tool {
			return strings.HasPrefix(t.Name, "codex_")
		}
	}
	return false
}

// isCodexMirrorField is the mirror predicate: the audit codex pins, the codex
// opt-ins, and the codex MCP tool toggles.
func isCodexMirrorField(fieldName string) bool {
	switch {
	case fieldName == "workflow.audit.gates.codex":
		return true
	case strings.HasPrefix(fieldName, "workflow.audit.codex."):
		return true
	case strings.HasPrefix(fieldName, "workflow.codex."):
		return true
	default:
		return isCodexMCPToolField(fieldName)
	}
}

// codexMirrorFieldNames returns the registry fields the mirror derives, in
// registry order. It does NOT include the declared exception — that arrives
// through codexMirrorGroups by name, so the two remain separately falsifiable.
func codexMirrorFieldNames() []string {
	var out []string
	for _, f := range settings.AllFields() {
		if isCodexMirrorField(f.Name) {
			out = append(out, f.Name)
		}
	}
	return out
}

// codexMirrorOwner reports the panel that owns a mirrored field's editing
// surface. workflow.audit.* is edited on the Audit tab; the codex opt-ins and
// the MCP tool toggles are edited on the MCP tab (the opt-ins render there
// inside codexAuthBlock, which is why isCodexToggleFieldName keeps them out of
// the workflow partition).
func codexMirrorOwner(fieldName string) string {
	if strings.HasPrefix(fieldName, "workflow.audit.") {
		return codexOwnerTabAudit
	}
	return codexOwnerTabMCP
}

// codexMirrorRow is one displayed row: the dot-path identifier, the value as
// read from disk, and a route to where it is edited. It carries no control and
// no name, so it cannot be submitted (REQ-WCP-002).
type codexMirrorRow struct {
	Name     string // dot-path identifier
	Value    string // current value as read from disk ("" → the unset placeholder)
	OwnerTab string // codexOwnerTabAudit | codexOwnerTabMCP
	LabelKey string // existing f.<name>.title — no new per-row keys
	NoteKey  string // non-empty only on the declared exception
}

// Unset reports whether the field has no value on disk.
func (r codexMirrorRow) Unset() bool { return r.Value == "" }

// OwnerLinkKey is the i18n key of the row's owning-tab link label.
func (r codexMirrorRow) OwnerLinkKey() string { return "tab.codex.edit_on." + r.OwnerTab }

// OwnerHref is the row's route to the owning tab's editing surface.
func (r codexMirrorRow) OwnerHref() string { return "/settings?tab=" + r.OwnerTab }

// codexMirrorGroup is one titled block of mirror rows.
type codexMirrorGroup struct {
	TitleKey string
	Title    string
	HelpKey  string
	Help     string
	Rows     []codexMirrorRow
}

// codexMirrorGroups builds the panel's three row groups in reading order: the
// audit pins (led by the declared exception, since "which backend gates merges"
// frames everything below it), the codex opt-ins, then the MCP tool enablement.
func codexMirrorGroups(view pageView) []codexMirrorGroup {
	audit := codexMirrorGroup{
		TitleKey: "tab.codex.group.audit",
		Title:    "Audit backend and codex pins",
		Rows: []codexMirrorRow{{
			// The one declared exception, labelled as shared rather than as a
			// codex-owned setting (REQ-WCP-005).
			Name:     codexAuditModelField,
			Value:    view.SchemaValues[codexAuditModelField],
			OwnerTab: codexOwnerTabAudit,
			LabelKey: "f." + codexAuditModelField + ".title",
			NoteKey:  codexSharedBackendI18nKey,
		}},
	}
	optin := codexMirrorGroup{
		TitleKey: "tab.codex.group.optin",
		Title:    "Codex opt-ins",
	}
	mcp := codexMirrorGroup{
		TitleKey: "tab.codex.group.mcp",
		Title:    "MCP tool enablement",
		HelpKey:  "tab.codex.group.mcp.help",
		Help:     "An unset value reads as enabled; only an explicit false turns a tool off.",
	}

	for _, name := range codexMirrorFieldNames() {
		row := codexMirrorRow{
			Name:     name,
			Value:    view.SchemaValues[name],
			OwnerTab: codexMirrorOwner(name),
			LabelKey: "f." + name + ".title",
		}
		switch {
		case strings.HasPrefix(name, "mcp.tools."):
			mcp.Rows = append(mcp.Rows, row)
		case strings.HasPrefix(name, "workflow.codex."):
			optin.Rows = append(optin.Rows, row)
		default:
			audit.Rows = append(audit.Rows, row)
		}
	}
	return []codexMirrorGroup{audit, optin, mcp}
}
