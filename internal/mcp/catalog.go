// Package mcp holds the single shared declaration of the moai self-hosted MCP
// server's tool surface (SPEC-MCP-CONSOLE-001 REQ-C-1 / C-C-5 / AP-C-4).
//
// Both the server-side registration (internal/cli/mcp_server.go
// registerMoaiMCPTools) AND the console-side settings schema
// (internal/settings schema_sections.go) consume MoaiMCPTools, so a tool added
// to registration cannot silently go unrepresented in the console: the guard
// test asserts the catalog equals the actual tools/list registration set, and
// the schema derives its per-tool enablement fields from the same list.
//
// This package is intentionally neutral: it imports nothing from internal/cli
// or internal/web, so internal/settings (itself neutral) can import it without
// breaking the import graph.
package mcp

// ToolDef describes one tool in the moai MCP server's declared surface.
type ToolDef struct {
	// Name is the tool identifier as it appears in tools/list and as the
	// console's per-tool enablement key is derived from
	// (mcp.tools.<name>.enabled).
	Name string
	// WriteCapable is true for the nine tools whose handler may mutate state
	// (goal_arm, verify_snapshot, codex_task, codex_job_cancel, glm_task,
	// glm_job_cancel, plus the session-messaging broker's three mutating
	// tools at the catalog tail) and false for the sixteen read-only tools.
	// The console renders this distinction (REQ-C-3 / AC-C-003); M1 carries
	// it so the declaration is complete.
	WriteCapable bool
}

// moaiMCPTools is the SINGLE shared declaration of the moai MCP server's tool
// surface. A tool registered by registerMoaiMCPTools MUST appear here — the
// guard test (internal/cli TestMoaiMCPServer_RegistrationMatchesCatalog) fails
// on any divergence.
//
// @MX:ANCHOR: [AUTO] single shared declaration of the moai MCP tool surface
// @MX:REASON: [AUTO] consumed by both internal/cli (registration) and internal/settings (schema); divergence = a tool added to the server but invisible to the console (AP-C-4).
// @MX:SPEC: SPEC-MCP-CONSOLE-001
var moaiMCPTools = []ToolDef{
	{Name: "session_list", WriteCapable: false},
	{Name: "goal_status", WriteCapable: false},
	{Name: "goal_arm", WriteCapable: true},
	{Name: "spec_progress", WriteCapable: false},
	{Name: "verify_snapshot", WriteCapable: true},
	{Name: "verify_trend", WriteCapable: false},
	{Name: "spec_audit", WriteCapable: false},
	{Name: "spec_drift", WriteCapable: false},
	{Name: "audit_cache", WriteCapable: false},
	{Name: "codex_audit", WriteCapable: false},
	{Name: "codex_setup", WriteCapable: false},
	{Name: "codex_task", WriteCapable: true},
	{Name: "codex_job_status", WriteCapable: false},
	{Name: "codex_job_result", WriteCapable: false},
	{Name: "codex_job_cancel", WriteCapable: true},
	{Name: "glm_task", WriteCapable: true},
	{Name: "glm_job_status", WriteCapable: false},
	{Name: "glm_job_result", WriteCapable: false},
	{Name: "glm_job_cancel", WriteCapable: true},
	{Name: "glm_audit", WriteCapable: false},
	{Name: "audit_multi", WriteCapable: false},
	{Name: "session_msg_register", WriteCapable: true},
	{Name: "session_msg_list", WriteCapable: false},
	{Name: "session_msg_send", WriteCapable: true},
	{Name: "session_msg_poll", WriteCapable: true},
}

// MoaiMCPTools returns the single shared declaration of the moai MCP server's
// tool surface. Callers MUST NOT mutate the returned slice.
func MoaiMCPTools() []ToolDef {
	return moaiMCPTools
}

// MoaiMCPToolNames returns just the tool names from the shared declaration, in
// declaration order. Convenience for callers that need only the identifier set
// (e.g. the schema field generator).
func MoaiMCPToolNames() []string {
	names := make([]string, len(moaiMCPTools))
	for i, t := range moaiMCPTools {
		names[i] = t.Name
	}
	return names
}
