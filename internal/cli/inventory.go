package cli

// @MX:ANCHOR: [AUTO] moai inventory — unified read-only composition of the three
// runtime inventory surfaces (sessions / worktrees / harnesses).
// @MX:REASON: SPEC-DIVECC-INVENTORY-VIEW-001 final candidate (Epic Dive-into-CC N6).
// Composes session.QueryActiveWork / WorktreeProvider.List / harness.ListHarnesses
// read-only; output stability (--json shape) is contract per REQ-INV-005.

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/cli/harness"
	wtroot "github.com/modu-ai/moai-adk/internal/cli/worktree"
	"github.com/modu-ai/moai-adk/internal/session"
)

// newInventoryCmd builds the `moai inventory [--json]` command — a single
// top-level command (NOT a subcommand group) that composes the three read-only
// inventory surfaces into one view (REQ-INV-001, REQ-INV-002). The default
// output is a compact human-readable summary; --json emits the structured
// UnifiedInventoryReport (REQ-INV-005/006). The command runs in CLI/subagent
// context and emits exit codes + JSON/text only; it never prompts the user
// (REQ-INV-013, C-HRA-008 — the orchestrator owns user interaction).
func newInventoryCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "inventory",
		Short: "Show a unified read-only view of sessions, worktrees, and harnesses",
		Long: `Compose the three runtime inventory surfaces — active sessions, git
worktrees, and user-owned harnesses — into one read-only view.

Each surface degrades independently: an absent backing yields a 0-count, and a
genuine backing error (e.g. invoked outside a git repository) annotates only
that surface while the others still render. The command exits non-zero only
when no surface could be rendered.`,
		GroupID: "tools",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			projectRoot, err := resolveProjectRoot(cmd)
			if err != nil {
				return fmt.Errorf("inventory: %w", err)
			}
			// Lazily wire the worktree provider the same way the `worktree`
			// subcommand does (deps.EnsureGit → worktree.WorktreeProvider).
			// This is a CALL into existing wiring, not a backing-package change
			// (REQ-INV-011). When the project root is not a git repository,
			// EnsureGit fails and the provider stays unset — collectWorktrees
			// then degrades the worktree surface per REQ-INV-008/010.
			ensureWorktreeProvider(projectRoot)
			report := collectInventory(projectRoot)

			if renderedSurfaceCount(report) == 0 {
				// No surface could be rendered — exit non-zero (REQ-INV-008).
				return fmt.Errorf("inventory: no surface could be rendered (sessions=%q worktrees=%q harnesses=%q)",
					report.Sessions.Error, report.Worktrees.Error, report.Harnesses.Error)
			}

			if jsonOutput {
				return renderInventoryJSON(cmd, report)
			}
			return renderInventoryText(cmd, report)
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "emit machine-readable JSON output (UnifiedInventoryReport)")
	cmd.Flags().String("project-root", "", "project root path (default: current directory)")
	return cmd
}

// renderInventoryJSON emits the structured UnifiedInventoryReport via
// json.MarshalIndent to stdout, mirroring the internal/cli/session.go
// convention (REQ-INV-005).
func renderInventoryJSON(cmd *cobra.Command, report UnifiedInventoryReport) error {
	out, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("inventory: marshal: %w", err)
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(out))
	return nil
}

// renderInventoryText emits a compact human-readable 3-surface summary using
// renderCard (one card per surface, with a count header + key-field rows). An
// all-empty report renders three (0) headers, not an error (REQ-INV-006). Each
// surface card carries the surface label so the three labels (Sessions /
// Worktrees / Harnesses) are always present (AC-INV-004).
func renderInventoryText(cmd *cobra.Command, report UnifiedInventoryReport) error {
	w := cmd.OutOrStdout()

	_, _ = fmt.Fprintln(w, renderCard(
		fmt.Sprintf("Sessions (%d)", report.Sessions.Count),
		sessionsCardBody(report.Sessions),
	))
	_, _ = fmt.Fprintln(w, renderCard(
		fmt.Sprintf("Worktrees (%d)", report.Worktrees.Count),
		worktreesCardBody(report.Worktrees),
	))
	_, _ = fmt.Fprintln(w, renderCard(
		fmt.Sprintf("Harnesses (%d)", report.Harnesses.Count),
		harnessesCardBody(report.Harnesses),
	))
	return nil
}

// surfaceErrorOrEmpty returns the degradation line for a surface that errored,
// the "(empty)" marker for a 0-count surface, or "" when the caller should
// render rows instead. Centralizes the error/empty card-body branch so each
// surface body builder stays small.
func surfaceErrorOrEmpty(errMsg string, count int) string {
	if errMsg != "" {
		return "error: " + errMsg
	}
	if count == 0 {
		return "(empty)"
	}
	return ""
}

// sessionsCardBody builds the session card content (key fields per row).
func sessionsCardBody(s SessionInventory) string {
	if body := surfaceErrorOrEmpty(s.Error, s.Count); body != "" {
		return body
	}
	pairs := make([]kvPair, 0, len(s.Entries))
	for _, e := range s.Entries {
		pairs = append(pairs, kvPair{
			key:   e.SessionID,
			value: fmt.Sprintf("spec=%s phase=%s", e.SpecID, e.Phase),
		})
	}
	return renderKeyValueLines(pairs)
}

// worktreesCardBody builds the worktree card content (key fields per row).
func worktreesCardBody(wtv WorktreeInventory) string {
	if body := surfaceErrorOrEmpty(wtv.Error, wtv.Count); body != "" {
		return body
	}
	pairs := make([]kvPair, 0, len(wtv.Entries))
	for _, e := range wtv.Entries {
		pairs = append(pairs, kvPair{
			key:   e.Branch,
			value: fmt.Sprintf("head=%s path=%s", e.HEAD, e.Path),
		})
	}
	return renderKeyValueLines(pairs)
}

// harnessesCardBody builds the harness card content (key fields per row).
func harnessesCardBody(h HarnessInventory) string {
	if body := surfaceErrorOrEmpty(h.Error, h.Count); body != "" {
		return body
	}
	pairs := make([]kvPair, 0, len(h.Entries))
	for _, e := range h.Entries {
		val := fmt.Sprintf("domain=%s", e.Domain)
		if e.ManifestMissing {
			val += " manifest_missing=true"
		}
		pairs = append(pairs, kvPair{key: e.Name, value: val})
	}
	return renderKeyValueLines(pairs)
}

// UnifiedInventoryReport is the --json payload composing the three read-only
// inventory surfaces (REQ-INV-005). It owns NO data; every field is projected
// from an existing surface's exported function.
type UnifiedInventoryReport struct {
	Sessions  SessionInventory  `json:"sessions"`
	Worktrees WorktreeInventory `json:"worktrees"`
	Harnesses HarnessInventory  `json:"harnesses"`
}

// SessionInventory projects session.QueryActiveWork output (REQ-INV-004).
type SessionInventory struct {
	Count   int                 `json:"count"`
	Entries []SessionSummaryRow `json:"entries"`
	// Error carries a per-surface degradation message (REQ-INV-008). Empty when
	// the surface rendered successfully (including the legitimate 0-count case).
	Error string `json:"error,omitempty"`
}

// SessionSummaryRow is the key-field summary of one session entry (REQ-INV-004).
type SessionSummaryRow struct {
	SessionID string `json:"session_id"` // short form via shortID()
	SpecID    string `json:"spec_id"`
	Phase     string `json:"phase"`
}

// WorktreeInventory projects WorktreeProvider.List output (REQ-INV-004).
type WorktreeInventory struct {
	Count   int                  `json:"count"`
	Entries []WorktreeSummaryRow `json:"entries"`
	Error   string               `json:"error,omitempty"`
}

// WorktreeSummaryRow is the key-field summary of one worktree entry (REQ-INV-004).
type WorktreeSummaryRow struct {
	Branch string `json:"branch"`
	Path   string `json:"path"`
	HEAD   string `json:"head"` // short form (first 8) via shortID()
}

// HarnessInventory projects harness.ListHarnesses output (REQ-INV-004).
type HarnessInventory struct {
	Count   int                 `json:"count"`
	Entries []HarnessSummaryRow `json:"entries"`
	Error   string              `json:"error,omitempty"`
}

// HarnessSummaryRow is the key-field summary of one harness entry (REQ-INV-004).
type HarnessSummaryRow struct {
	Name            string `json:"name"`
	Domain          string `json:"domain"`
	ManifestMissing bool   `json:"manifest_missing"`
}

// collectInventory composes all three read-only inventory surfaces into a
// single UnifiedInventoryReport (REQ-INV-003). Each surface degrades
// independently: an empty backing yields a 0-count (REQ-INV-007); a genuine
// backing error populates that surface's Error field while the others still
// render (REQ-INV-008). collectInventory itself never returns an error.
func collectInventory(projectRoot string) UnifiedInventoryReport {
	return UnifiedInventoryReport{
		Sessions:  collectSessions(),
		Worktrees: collectWorktrees(),
		Harnesses: collectHarnesses(projectRoot),
	}
}

// collectSessions projects session.QueryActiveWork("") into a SessionInventory.
// An absent registry file yields an empty slice (count 0, no error); a genuine
// read error populates Error (REQ-INV-007, REQ-INV-008).
func collectSessions() SessionInventory {
	entries, err := session.QueryActiveWork("")
	if err != nil {
		return SessionInventory{Error: err.Error()}
	}
	rows := make([]SessionSummaryRow, 0, len(entries))
	for _, e := range entries {
		rows = append(rows, SessionSummaryRow{
			SessionID: shortID(e.SessionID),
			SpecID:    e.SpecID,
			Phase:     e.Phase,
		})
	}
	return SessionInventory{Count: len(rows), Entries: rows}
}

// ensureWorktreeProvider lazily wires worktree.WorktreeProvider from the global
// deps composition root, mirroring the `worktree` subcommand's
// PersistentPreRunE wiring (root.go). It is best-effort: when deps is nil, the
// provider is already set, or EnsureGit fails (e.g. not a git repository), it
// leaves the provider unchanged so collectWorktrees can degrade gracefully
// (REQ-INV-008/010). It modifies NO backing package — it only CALLS
// deps.EnsureGit and SETS the existing worktree.WorktreeProvider var.
func ensureWorktreeProvider(projectRoot string) {
	if wtroot.WorktreeProvider != nil {
		return
	}
	if deps == nil {
		return
	}
	if err := deps.EnsureGit(projectRoot); err != nil {
		return // not a git repo (or other git error) → leave provider unset
	}
	wtroot.WorktreeProvider = deps.GitWorktree
}

// collectWorktrees projects WorktreeProvider.List() into a WorktreeInventory.
// A nil provider (git module unavailable) or a List() error (e.g. invoked
// outside a git repository → "fatal: not a git repository") populates Error;
// the worktree surface then degrades while sessions/harnesses still render
// (REQ-INV-008, REQ-INV-010). The provider is read from the existing
// worktree.WorktreeProvider var — no backing-package modification (REQ-INV-011).
func collectWorktrees() WorktreeInventory {
	provider := wtroot.WorktreeProvider
	if provider == nil {
		return WorktreeInventory{Error: "worktree provider unavailable (git module not initialized)"}
	}
	worktrees, err := provider.List()
	if err != nil {
		return WorktreeInventory{Error: err.Error()}
	}
	rows := make([]WorktreeSummaryRow, 0, len(worktrees))
	for _, w := range worktrees {
		rows = append(rows, WorktreeSummaryRow{
			Branch: w.Branch,
			Path:   w.Path,
			HEAD:   shortID(w.HEAD),
		})
	}
	return WorktreeInventory{Count: len(rows), Entries: rows}
}

// collectHarnesses projects harness.ListHarnesses(projectRoot) into a
// HarnessInventory. An absent harness directory yields (nil, nil) → count 0
// (REQ-INV-007); a genuine read error populates Error (REQ-INV-008).
func collectHarnesses(projectRoot string) HarnessInventory {
	entries, err := harness.ListHarnesses(projectRoot)
	if err != nil {
		return HarnessInventory{Error: err.Error()}
	}
	rows := make([]HarnessSummaryRow, 0, len(entries))
	for _, h := range entries {
		rows = append(rows, HarnessSummaryRow{
			Name:            h.Name,
			Domain:          h.Domain,
			ManifestMissing: h.ManifestMissing,
		})
	}
	return HarnessInventory{Count: len(rows), Entries: rows}
}

// renderedSurfaceCount reports how many of the three surfaces rendered without
// a per-surface error. Exit-code discipline (REQ-INV-008): the command exits
// non-zero only when this count is 0 (no surface could be rendered).
func renderedSurfaceCount(rep UnifiedInventoryReport) int {
	n := 0
	if rep.Sessions.Error == "" {
		n++
	}
	if rep.Worktrees.Error == "" {
		n++
	}
	if rep.Harnesses.Error == "" {
		n++
	}
	return n
}
