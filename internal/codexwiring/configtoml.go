package codexwiring

import (
	"regexp"
	"strings"
)

// Canonical MCP server registration constants — the SAME values the Claude-side
// .mcp.json provisioning writes (mcp_server.go moaiMCPServer* constants):
// PATH-resolved command, no absolute path, single fixed arg.
const (
	// mcpServerTableHeader is the [mcp_servers.moai] table header.
	mcpServerTableHeader = "[mcp_servers.moai]"
	// mcpServerCommandValue is the PATH-resolved server command.
	mcpServerCommandValue = "moai"
	// mcpServerArgValue is the single fixed server arg.
	mcpServerArgValue = "mcp-server"
	// mcpApprovalMode is the capability-based approval mode: `writes` prompts
	// for tools NOT marked read-only (MCP ReadOnlyHint annotation) — the
	// approval set therefore rides on server annotations, never on tool-name
	// enumeration (spec §A.4).
	mcpApprovalMode = "writes"
)

// StatusLineAllowlist is the fixed allowlist of the 29 canonical Codex TUI
// status-line item identifiers (openai/codex StatusLineItem enum, kebab-case)
// recorded in SPEC-CODEX-WIRING-001 §A.6. Emission and validation both read
// this single constant; the 7 parse-only legacy aliases are deliberately
// absent (canonical tokens only). Callers MUST NOT mutate this slice — the
// allowlist stays a repository-fixed constant by design (no blind upstream
// tracking; additions are an explicit judgment at documentation-refresh time).
var StatusLineAllowlist = []string{
	"model", "model-with-reasoning", "reasoning", "current-dir", "project-name",
	"hostname", "git-branch", "pull-request-number", "branch-changes", "run-state",
	"permissions", "approval-mode", "context-remaining", "context-used",
	"five-hour-limit", "weekly-limit", "codex-version", "context-window-size",
	"used-tokens", "total-input-tokens", "total-output-tokens", "thread-credits",
	"estimated-thread-cost", "thread-id", "fast-mode", "raw-output", "thread-title",
	"workspace-headline", "task-progress",
}

// defaultStatusLine is the fixed default configuration (operator directive
// 2026-08-24): the 5 canonical tokens, a superset of Codex's own 3-token
// default plus git-branch and thread-id.
var defaultStatusLine = []string{
	"model-with-reasoning", "context-remaining", "git-branch", "current-dir", "thread-id",
}

// DefaultStatusLine returns a copy of the default status_line configuration.
func DefaultStatusLine() []string {
	out := make([]string, len(defaultStatusLine))
	copy(out, defaultStatusLine)
	return out
}

// statusLineDefaultTOML is the rendered `status_line = [...]` assignment for
// the default configuration — the exact line branch (i)/(ii) inserts.
var statusLineDefaultTOML = renderStatusLineTOML(DefaultStatusLine())

// renderStatusLineTOML renders one status_line array assignment.
func renderStatusLineTOML(tokens []string) string {
	quoted := make([]string, len(tokens))
	for i, tok := range tokens {
		quoted[i] = `"` + tok + `"`
	}
	return "status_line = [" + strings.Join(quoted, ", ") + "]"
}

// Table-header detectors. Anchored so [mcp_servers.moai2] or [mcp_servers.moai-x]
// do not satisfy the moai-table match, and [profile.tui] / [[tui]] do not
// satisfy the tui-table match (dotted tables and array-of-tables are distinct
// surfaces).
var (
	mcpMoaiTableRe  = regexp.MustCompile(`^\[mcp_servers\.moai\]\s*(#.*)?$`)
	tuiTableRe      = regexp.MustCompile(`^\[tui\]\s*(#.*)?$`)
	anyTableRe      = regexp.MustCompile(`^\[\[?[^\]]*\]\]?\s*(#.*)?$`)
	statusLineKeyRe = regexp.MustCompile(`^status_line\s*=`)
)

// EnsureMCPTable returns content with the canonical [mcp_servers.moai] table
// ensured. Create-if-absent ONLY (plan D2): an existing table is user-owned
// and byte-invariant — divergence from the canonical shape is the doctor's to
// report, never the writer's to repair (REQ-CW-004/005). The table carries no
// tool-name enumeration: the approval policy rides on server annotations.
func EnsureMCPTable(content []byte) []byte {
	body := string(content)
	if tablePresent(body, mcpMoaiTableRe) {
		return content
	}
	table := mcpServerTableHeader + "\n" +
		"command = \"" + mcpServerCommandValue + "\"\n" +
		"args = [\"" + mcpServerArgValue + "\"]\n" +
		"default_tools_approval_mode = \"" + mcpApprovalMode + "\"\n"
	return []byte(appendSection(body, table))
}

// EnsureStatusLine returns content with [tui].status_line ensured per the
// three-branch merge rule (plan D4 / REQ-CW-013):
//
//	(i)   no [tui] table      → new [tui] section appended at EOF
//	(ii)  [tui] without the key → assignment inserted right after the header
//	(iii) key present          → byte-invariant (user-owned)
//
// Everything else — other keys inside [tui], other tables, the whole rest of
// the file — is preserved untouched.
func EnsureStatusLine(content []byte) []byte {
	lines := splitLines(string(content))
	for i, line := range lines {
		if !tuiTableRe.MatchString(line) {
			continue
		}
		// Branch (ii) or (iii): a [tui] table exists. Scan its extent for the
		// key — the section runs until the next table header of any kind.
		for j := i + 1; j < len(lines); j++ {
			if anyTableRe.MatchString(lines[j]) {
				break // section ended without the key → branch (ii) below
			}
			if statusLineKeyRe.MatchString(lines[j]) {
				return content // branch (iii): user-owned, byte-invariant
			}
		}
		// Branch (ii): insert the assignment directly after the header line.
		out := make([]byte, 0, len(content)+len(statusLineDefaultTOML)+1)
		var sb strings.Builder
		for k := 0; k <= i; k++ {
			sb.WriteString(lines[k])
			sb.WriteString("\n")
		}
		sb.WriteString(statusLineDefaultTOML)
		sb.WriteString("\n")
		for k := i + 1; k < len(lines); k++ {
			sb.WriteString(lines[k])
			sb.WriteString("\n")
		}
		out = append(out, sb.String()...)
		return out
	}
	// Branch (i): append a new [tui] section at EOF.
	return []byte(appendSection(string(content), "[tui]\n"+statusLineDefaultTOML+"\n"))
}

// tablePresent reports whether any line of body is a table header matching re.
func tablePresent(body string, re *regexp.Regexp) bool {
	for _, line := range splitLines(body) {
		if re.MatchString(line) {
			return true
		}
	}
	return false
}

// MCPTableStatus reports what a config.toml carries for the
// [mcp_servers.moai] surface — the doctor's read-only view (the writer never
// repairs a user-owned table; divergence is REPORTED, REQ-CW-005).
type MCPTableStatus struct {
	// Present reports whether a [mcp_servers.moai] table header exists.
	Present bool
	// Canonical reports whether the table carries exactly the canonical
	// command/args/approval assignments.
	Canonical bool
}

// canonicalMCPAssignments are the three assignments EnsureMCPTable writes.
var canonicalMCPAssignments = []string{
	"command = \"" + mcpServerCommandValue + "\"",
	"args = [\"" + mcpServerArgValue + "\"]",
	"default_tools_approval_mode = \"" + mcpApprovalMode + "\"",
}

// InspectMCPTable reads the [mcp_servers.moai] table status of a config.toml
// body: whether the table is present and whether its assignment lines match
// the canonical shape. A present-but-non-canonical table is user-owned drift
// the doctor reports; the writer leaves it byte-invariant.
func InspectMCPTable(content []byte) MCPTableStatus {
	lines := splitLines(string(content))
	status := MCPTableStatus{}
	for i, line := range lines {
		if !mcpMoaiTableRe.MatchString(line) {
			continue
		}
		status.Present = true
		status.Canonical = true
		seen := 0
		for j := i + 1; j < len(lines); j++ {
			if anyTableRe.MatchString(lines[j]) {
				break
			}
			for _, want := range canonicalMCPAssignments {
				if strings.TrimSpace(lines[j]) == want {
					seen++
				}
			}
		}
		if seen != len(canonicalMCPAssignments) {
			status.Canonical = false
		}
		return status
	}
	return status
}

// splitLines splits body into trimmed-of-newline lines. A trailing newline
// yields no extra empty element.
func splitLines(body string) []string {
	if body == "" {
		return nil
	}
	lines := strings.Split(strings.TrimSuffix(body, "\n"), "\n")
	return lines
}

// appendSection appends a section block to body, guaranteeing exactly one
// blank-line separation from preceding content.
func appendSection(body, section string) string {
	if body == "" {
		return section
	}
	if !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	return body + "\n" + section
}
