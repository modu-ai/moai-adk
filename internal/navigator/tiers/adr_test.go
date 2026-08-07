package tiers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeADR writes a fixture ADR file under .moai/decisions/.
func writeADR(t *testing.T, root, id, body string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "decisions")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, id+".md")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestADR_ResolvePresent exercises AC-NS3-008: a @NAV:DEC-<id> whose ADR file
// exists resolves to adr_path and the four canonical fields parse
// best-effort.
func TestADR_ResolvePresent(t *testing.T) {
	root := t.TempDir()
	body := `# Title

Decision Date: 2026-08-06
Status: Accepted

## Context

Some context.

## Decision

The decision.

## Consequences

The consequences.
`
	writeADR(t, root, "AUTH-STRATEGY", body)

	enr, parsed, ok := resolveADR(root, "AUTH-STRATEGY")
	if !ok {
		t.Fatalf("resolveADR ok=false; want true (file present)")
	}
	if enr.Identifier != "AUTH-STRATEGY" {
		t.Errorf("Identifier = %q; want AUTH-STRATEGY", enr.Identifier)
	}
	if enr.AdrPath == "" {
		t.Errorf("AdrPath empty; want non-empty")
	}
	if !strings.HasSuffix(enr.AdrPath, "AUTH-STRATEGY.md") {
		t.Errorf("AdrPath = %q; want suffix AUTH-STRATEGY.md", enr.AdrPath)
	}
	if parsed.Status != "Accepted" {
		t.Errorf("Status = %q; want Accepted", parsed.Status)
	}
	if parsed.DecisionDate != "2026-08-06" {
		t.Errorf("DecisionDate = %q; want 2026-08-06", parsed.DecisionDate)
	}
}

// TestADR_ResolveCaseInsensitiveFilename verifies filename match is
// case-insensitive (REQ-NS3-008 / design.md §1.D2).
func TestADR_ResolveCaseInsensitiveFilename(t *testing.T) {
	root := t.TempDir()
	writeADR(t, root, "Auth-Strategy", "# x\nStatus: Accepted\n")
	enr, _, ok := resolveADR(root, "AUTH-STRATEGY")
	if !ok {
		t.Fatalf("case-insensitive lookup failed")
	}
	if !strings.HasSuffix(enr.AdrPath, "Auth-Strategy.md") {
		t.Errorf("AdrPath = %q; want Auth-Strategy.md", enr.AdrPath)
	}
}

// TestADR_ResolveAbsent_DegradesGracefully exercises AC-NS3-008: a token
// whose <id> has no ADR file degrades — ok=true (the identifier exists) but
// AdrPath is empty. No error returned.
func TestADR_ResolveAbsent_DegradesGracefully(t *testing.T) {
	root := t.TempDir()
	enr, _, ok := resolveADR(root, "ORPHAN-ID")
	// No file present → ok=false (no enrichment emitted). The CALLER keeps
	// the decision node in the graph with adr_path empty; this function
	// simply signals "no ADR".
	if ok {
		t.Errorf("ok = true; want false (no ADR file)")
	}
	if enr.AdrPath != "" {
		t.Errorf("AdrPath = %q; want empty", enr.AdrPath)
	}
}

// TestADR_GrandfatheredShape_ParsesBestEffort verifies a pre-existing
// ADR-shaped file (design.md §1.D2 grandfathering) parses without error and
// carries the substance even when the heading shape does not match the
// formal template verbatim.
func TestADR_GrandfatheredShape_ParsesBestEffort(t *testing.T) {
	root := t.TempDir()
	// Shape similar to the actual lsp-client-choice.md (no formal Status,
	// has Decision Date + prose sections).
	body := `# LSP Client Selection Rationale

SPEC: SPEC-LSP-CORE-002
Decision Date: 2026-04-12

## Selected Library

github.com/charmbracelet/x/powernap v0.1.4

## Why

Because it is purpose-built.
`
	writeADR(t, root, "lsp-client-choice", body)
	enr, parsed, ok := resolveADR(root, "LSP-CLIENT-CHOICE")
	if !ok {
		t.Fatalf("grandfathered ADR did not resolve")
	}
	if enr.AdrPath == "" {
		t.Errorf("AdrPath empty; want set")
	}
	if parsed.DecisionDate != "2026-04-12" {
		t.Errorf("DecisionDate = %q; want 2026-04-12", parsed.DecisionDate)
	}
	// Status field is absent in the grandfathered shape → empty (degrade).
	if parsed.Status != "" {
		t.Errorf("Status = %q; want empty (best-effort degrade)", parsed.Status)
	}
}

// TestADR_Parse_MissingField_Degrades ensures a missing field degrades to
// empty string, never an error (REQ-NS3-008 BEST-EFFORT).
func TestADR_Parse_MissingField_Degrades(t *testing.T) {
	root := t.TempDir()
	body := `# Bare

Decision Date: 2026-01-01
` // No Status, no Context, no Decision, no Consequences.
	writeADR(t, root, "bare", body)
	_, parsed, _ := resolveADR(root, "bare")
	if parsed.DecisionDate != "2026-01-01" {
		t.Errorf("DecisionDate = %q; want 2026-01-01", parsed.DecisionDate)
	}
	if parsed.Status != "" {
		t.Errorf("Status = %q; want empty", parsed.Status)
	}
}

// TestADR_Supersede_Immutability exercises AC-NS3-009: supersede creates a
// new ADR carrying `supersedes:`, flips the prior Status: Accepted→Superseded,
// leaves the prior body BYTE-IDENTICAL except the Status line, and records a
// superseded_by edge.
func TestADR_Supersede_Immutability(t *testing.T) {
	root := t.TempDir()
	priorBody := `# AUTH-STRATEGY

Decision Date: 2026-01-01
Status: Accepted

## Context

Original context.

## Decision

Original decision.

## Consequences

Original consequences.
`
	writeADR(t, root, "AUTH-STRATEGY", priorBody)

	// Snapshot the prior body line-by-line BEFORE supersede.
	priorPath := filepath.Join(root, ".moai", "decisions", "AUTH-STRATEGY.md")
	beforeBytes, err := os.ReadFile(priorPath)
	if err != nil {
		t.Fatal(err)
	}
	beforeLines := strings.Split(string(beforeBytes), "\n")

	// Run the supersede.
	op, err := supersedeDecision(root, "AUTH-STRATEGY", "AUTH-STRATEGY-V2")
	if err != nil {
		t.Fatalf("supersedeDecision error: %v", err)
	}

	// (1) New ADR exists.
	newPath := filepath.Join(root, ".moai", "decisions", "AUTH-STRATEGY-V2.md")
	if _, err := os.Stat(newPath); err != nil {
		t.Fatalf("new ADR not created: %v", err)
	}
	newBytes, err := os.ReadFile(newPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(newBytes), "supersedes: AUTH-STRATEGY") {
		t.Errorf("new ADR missing `supersedes: AUTH-STRATEGY` line\n%s", newBytes)
	}

	// (2) Prior body byte-identical EXCEPT the Status line.
	afterBytes, err := os.ReadFile(priorPath)
	if err != nil {
		t.Fatal(err)
	}
	afterLines := strings.Split(string(afterBytes), "\n")
	if len(beforeLines) != len(afterLines) {
		t.Fatalf("prior line count changed: before=%d after=%d (immutability)", len(beforeLines), len(afterLines))
	}
	statusChanges := 0
	for i := range beforeLines {
		if beforeLines[i] == afterLines[i] {
			continue
		}
		// The ONLY acceptable change is the Status line.
		if strings.HasPrefix(strings.TrimSpace(beforeLines[i]), "Status:") &&
			strings.HasPrefix(strings.TrimSpace(afterLines[i]), "Status:") {
			if strings.TrimSpace(beforeLines[i]) == "Status: Accepted" &&
				strings.TrimSpace(afterLines[i]) == "Status: Superseded" {
				statusChanges++
				continue
			}
		}
		t.Errorf("prior line %d mutated beyond Status: before=%q after=%q",
			i, beforeLines[i], afterLines[i])
	}
	if statusChanges != 1 {
		t.Errorf("expected exactly 1 Status line change; got %d", statusChanges)
	}

	// (3) The op carries the superseded_by edge.
	if op.NewID != "AUTH-STRATEGY-V2" || op.PriorID != "AUTH-STRATEGY" {
		t.Errorf("op IDs wrong: prior=%q new=%q", op.PriorID, op.NewID)
	}
}

// TestADR_EnumerateDecisions_FromM0Graph exercises AC-NS3-010: given a set
// of decision identifiers (sourced from M0 nav-graph.json in production),
// enumerateDecisions emits DecisionEnrichment records carrying adr_path +
// superseded_by. Missing ADR → record still emitted, adr_path empty.
func TestADR_EnumerateDecisions_FromM0Graph(t *testing.T) {
	root := t.TempDir()
	writeADR(t, root, "A", "# A\nStatus: Accepted\n")
	writeADR(t, root, "B", "# B\nStatus: Accepted\n")
	// C has no ADR file (degrade).

	ids := []string{"A", "B", "C"}
	out := enumerateDecisions(root, ids)
	if len(out) != 3 {
		t.Fatalf("emitted %d records; want 3", len(out))
	}
	// Deterministic by identifier.
	byID := map[string]DecisionEnrichment{}
	for _, e := range out {
		byID[e.Identifier] = e
	}
	if byID["A"].AdrPath == "" {
		t.Errorf("A.AdrPath empty; want set")
	}
	if byID["C"].AdrPath != "" {
		t.Errorf("C.AdrPath = %q; want empty (missing ADR degrade)", byID["C"].AdrPath)
	}
}

// TestADR_SupersedeChain_ReflectedInEnrichment: after a supersede, the
// prior decision's enrichment carries SupersededBy and the new decision
// carries Supersedes.
func TestADR_SupersedeChain_ReflectedInEnrichment(t *testing.T) {
	root := t.TempDir()
	writeADR(t, root, "POLICY", "# POLICY\nStatus: Accepted\n")
	if _, err := supersedeDecision(root, "POLICY", "POLICY-V2"); err != nil {
		t.Fatal(err)
	}
	out := enumerateDecisions(root, []string{"POLICY", "POLICY-V2"})
	byID := map[string]DecisionEnrichment{}
	for _, e := range out {
		byID[e.Identifier] = e
	}
	if byID["POLICY"].SupersededBy != "POLICY-V2" {
		t.Errorf("POLICY.SupersededBy = %q; want POLICY-V2", byID["POLICY"].SupersededBy)
	}
	if byID["POLICY-V2"].Supersedes != "POLICY" {
		t.Errorf("POLICY-V2.Supersedes = %q; want POLICY", byID["POLICY-V2"].Supersedes)
	}
}
