// Package cli — doctor_mcp_version.go
//
// `moai doctor` check: does a RUNNING `moai mcp-server` still match the
// installed binary?
//
// The host spawns the MCP server once per session and never respawns it on
// reinstall, so `make install` leaves the host talking to the previous build.
// Nothing surfaces that: tools/list simply lacks the new tools and the old
// handlers keep answering. This check reads the per-PID build stamps the
// server writes at startup (mcp_server_runtime.go), probes their liveness,
// and warns when a live server's commit differs from this binary's.
package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// mcpServerVersionCheckName is the doctor check identifier (also the value
// accepted by `moai doctor --check`).
const mcpServerVersionCheckName = "MCP Server Version"

// checkMCPServerVersion compares every live `moai mcp-server` process
// recorded for this project against the installed binary's build commit.
//
// Reported OK (never WARN) when:
//   - no live server is recorded — nothing to compare
//   - this binary carries no commit metadata ("", "none", "unknown") — a
//     dev build cannot attribute a mismatch, so a WARN would be noise
//
// WARN only on positive evidence: a live server whose commit differs from
// this binary's. Dead stamps left by a hard-killed server are pruned in
// passing so they cannot accumulate into a false positive later.
func checkMCPServerVersion(projectRoot string, verbose bool) DiagnosticCheck {
	return checkMCPServerVersionAgainst(projectRoot, version.GetCommit(), verbose)
}

// checkMCPServerVersionAgainst is checkMCPServerVersion with the installed
// binary's commit injected, so the comparison can be exercised without
// mutating the package-level build vars.
func checkMCPServerVersionAgainst(projectRoot, binaryCommit string, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: mcpServerVersionCheckName}

	live, stalePaths := liveMCPServerRuntimeRecords(projectRoot)
	for _, p := range stalePaths {
		_ = os.Remove(p)
	}

	if len(live) == 0 {
		check.Status = uikit.CheckOK
		check.Message = "no running moai MCP server recorded"
		if verbose {
			check.Detail = "the server stamps .moai/state/mcp-server/<pid>.json while it runs"
		}
		return check
	}

	binCommit := strings.TrimSpace(binaryCommit)
	if binCommit == "" || binCommit == "none" || binCommit == "unknown" {
		check.Status = uikit.CheckOK
		check.Message = fmt.Sprintf("%d running MCP server(s); development build (no commit metadata to compare)", len(live))
		return check
	}

	var stale []string
	for _, rec := range live {
		recCommit := strings.TrimSpace(rec.Commit)
		if recCommit == "" || recCommit == "none" || recCommit == "unknown" {
			// The server was itself a dev build; it cannot be attributed
			// either way, so it is not counted as a mismatch.
			continue
		}
		if !commitsMatch(recCommit, binCommit) {
			stale = append(stale, fmt.Sprintf("pid %d: %s", rec.PID, shortCommit(recCommit)))
		}
	}

	if len(stale) == 0 {
		check.Status = uikit.CheckOK
		check.Message = fmt.Sprintf("%d running MCP server(s) match the installed binary (%s)", len(live), shortCommit(binCommit))
		return check
	}

	sort.Strings(stale)
	check.Status = uikit.CheckWarn
	check.Message = fmt.Sprintf("running MCP server is stale (%s; binary: %s)", strings.Join(stale, ", "), shortCommit(binCommit))
	check.Detail = "Reconnect the MCP server so the host respawns it (Claude Code: /mcp -> reconnect, or restart the session). " +
		"A rebuilt binary does not replace an already-running server, so newly added tools stay absent until then."
	return check
}

// commitsMatch reports whether two commit strings identify the same commit,
// tolerating differing abbreviation lengths (one side may be a short hash).
func commitsMatch(a, b string) bool {
	if a == b {
		return true
	}
	return strings.HasPrefix(a, b) || strings.HasPrefix(b, a)
}
