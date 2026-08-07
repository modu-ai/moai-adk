package tiers

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"log/slog"
)

// decisionsDirRel is the ADR home (REQ-NS3-008 / design.md §1.D2). Pre-existing
// files in this directory are GRANDFATHERED — M4 does NOT rewrite them; the
// supersede operation creates NEW files only.
const decisionsDirRel = ".moai/decisions"

// adrParsed carries the best-effort four-field parse of an ADR file. A
// missing field degrades to empty string — the parse NEVER fails.
type adrParsed struct {
	DecisionDate string
	Status       string
	Supersedes   string // parsed from the optional `supersedes:` metadata line
	HasBody      bool   // true when the file is non-empty
}

// resolveADR looks up `.moai/decisions/<id>.md` by filename (case-insensitive)
// and returns a DecisionEnrichment + the parsed fields. Returns ok=false
// when no ADR file exists (the caller keeps the decision node with adr_path
// empty — graceful degrade per REQ-NS3-008).
//
// @MX:ANCHOR: [AUTO] ADR resolution entry; high fan_in (enumerateDecisions + supersedeDecision + tests)
// @MX:REASON: drives the Tier 2 enrichment of every M0 decision node; called once per decision identifier
// @MX:SPEC:SPEC-NAVIGATOR-SYNC-003
func resolveADR(projectRoot, decID string) (DecisionEnrichment, adrParsed, bool) {
	decDir := filepath.Join(projectRoot, decisionsDirRel)
	entries, err := os.ReadDir(decDir)
	if err != nil {
		return DecisionEnrichment{Identifier: decID}, adrParsed{}, false
	}
	target := strings.ToLower(decID) + ".md"
	var matchPath string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.ToLower(e.Name()) == target {
			matchPath = filepath.Join(decDir, e.Name())
			break
		}
	}
	if matchPath == "" {
		return DecisionEnrichment{Identifier: decID}, adrParsed{}, false
	}
	content, err := os.ReadFile(matchPath)
	if err != nil {
		slog.Debug("tiers: ADR read error", "path", matchPath, "error", err)
		return DecisionEnrichment{Identifier: decID}, adrParsed{}, false
	}
	parsed := parseADR(content)
	rel, _ := filepath.Rel(projectRoot, matchPath)
	if rel == "" {
		rel = matchPath
	}
	enr := DecisionEnrichment{
		Identifier: decID,
		AdrPath:    rel,
		Supersedes: parsed.Supersedes,
	}
	return enr, parsed, true
}

// parseADR performs a best-effort four-field parse of an ADR body. It
// recognizes the canonical fields by case-insensitive prefix:
//   - `Decision Date:` → DecisionDate
//   - `Status:` → Status
//   - `supersedes:` → Supersedes (optional metadata line, case-insensitive)
//
// A field absent from the body degrades to empty string. The parse never
// returns an error — it captures what is present.
func parseADR(content []byte) adrParsed {
	out := adrParsed{HasBody: len(bytes.TrimSpace(content)) > 0}
	scanner := bufio.NewScanner(bytes.NewReader(content))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if rest, ok := stripFieldPrefix(line, "Decision Date:"); ok {
			out.DecisionDate = strings.TrimSpace(rest)
			continue
		}
		if rest, ok := stripFieldPrefix(line, "Status:"); ok {
			out.Status = strings.TrimSpace(rest)
			continue
		}
		if rest, ok := stripFieldPrefix(line, "supersedes:"); ok {
			out.Supersedes = strings.TrimSpace(rest)
			continue
		}
	}
	return out
}

// stripFieldPrefix reports whether line begins (case-insensitively) with the
// given prefix and returns the remainder if so.
func stripFieldPrefix(line, prefix string) (string, bool) {
	if len(line) < len(prefix) {
		return "", false
	}
	if strings.EqualFold(line[:len(prefix)], prefix) {
		return line[len(prefix):], true
	}
	return "", false
}

// supersedeOp records the result of a supersede operation (REQ-NS3-009) so
// the caller can emit the superseded_by edge into the overlay.
type supersedeOp struct {
	PriorID string
	NewID   string
}

// supersedeDecision performs the ADR supersede operation (REQ-NS3-009):
//
//   - Creates a NEW ADR file `<newID>.md` carrying a `supersedes: <priorID>`
//     metadata line. The new ADR's body is a minimal template the human/agent
//     refines later.
//   - Flips the prior ADR's `Status:` line from `Accepted` to `Superseded`.
//     The prior body is byte-identical EXCEPT that single line (immutability
//     characterization).
//
// Returns an error only when the prior ADR is missing or unreadable; the new
// ADR write is fail-open (logged at debug).
//
// @MX:WARN: [AUTO] supersede mutates an existing ADR's Status line only — body immutability is the load-bearing invariant (AC-NS3-009)
// @MX:REASON: a supersede that silently edits the prior body destroys the audit-trail property that makes ADRs valuable
// @MX:SPEC:SPEC-NAVIGATOR-SYNC-003
func supersedeDecision(projectRoot, priorID, newID string) (supersedeOp, error) {
	op := supersedeOp{PriorID: priorID, NewID: newID}
	decDir := filepath.Join(projectRoot, decisionsDirRel)
	priorPath := findADRPath(projectRoot, priorID)
	if priorPath == "" {
		return op, fmt.Errorf("tiers: prior ADR not found for %q", priorID)
	}
	content, err := os.ReadFile(priorPath)
	if err != nil {
		return op, fmt.Errorf("tiers: read prior ADR %q: %w", priorPath, err)
	}

	// (1) Flip the prior Status line. The body is otherwise byte-identical.
	flipped := flipStatusLine(content, "Accepted", "Superseded")
	if err := os.WriteFile(priorPath, flipped, 0o644); err != nil {
		return op, fmt.Errorf("tiers: write prior ADR %q: %w", priorPath, err)
	}

	// (2) Create the new ADR (fail-open if it already exists).
	newPath := filepath.Join(decDir, newID+".md")
	if _, statErr := os.Stat(newPath); statErr == nil {
		slog.Debug("tiers: new ADR already exists, not overwriting", "path", newPath)
		return op, nil
	}
	newBody := buildSupersedeTemplate(newID, priorID)
	if err := os.WriteFile(newPath, []byte(newBody), 0o644); err != nil {
		slog.Debug("tiers: new ADR write failed (fail-open)", "path", newPath, "error", err)
	}
	return op, nil
}

// flipStatusLine returns content with the FIRST `Status: <from>` line replaced
// by `Status: <to>`. If no Status line matches, the content is returned
// unchanged (best-effort — fail-open).
func flipStatusLine(content []byte, from, to string) []byte {
	lines := bytes.Split(content, []byte("\n"))
	prefix := []byte("Status:")
	wantValue := strings.TrimSpace(from)
	flipped := false
	for i, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if !bytes.HasPrefix(bytes.ToLower(trimmed), bytes.ToLower(prefix)) {
			continue
		}
		rest := strings.TrimSpace(string(trimmed[len(prefix):]))
		if rest != wantValue {
			continue
		}
		// Preserve leading indent of the original line.
		leading := line[:len(line)-len(bytes.TrimLeft(line, " \t"))]
		lines[i] = append(leading, []byte("Status: "+to)...)
		flipped = true
		break
	}
	if !flipped {
		return content
	}
	return bytes.Join(lines, []byte("\n"))
}

// buildSupersedeTemplate returns the new-ADR template carrying the
// `supersedes:` metadata line.
func buildSupersedeTemplate(newID, priorID string) string {
	return fmt.Sprintf(`# %s

Decision Date: <YYYY-MM-DD>
Status: Accepted
supersedes: %s

## Context

<Context for the new decision.>

## Decision

<The new decision.>

## Consequences

<Consequences of the supersession.>
`, newID, priorID)
}

// findADRPath returns the absolute path to <decID>.md under .moai/decisions/
// (case-insensitive match) or "" if absent.
func findADRPath(projectRoot, decID string) string {
	decDir := filepath.Join(projectRoot, decisionsDirRel)
	entries, err := os.ReadDir(decDir)
	if err != nil {
		return ""
	}
	target := strings.ToLower(decID) + ".md"
	for _, e := range entries {
		if !e.IsDir() && strings.ToLower(e.Name()) == target {
			return filepath.Join(decDir, e.Name())
		}
	}
	return ""
}

// enumerateDecisions emits DecisionEnrichment records for each identifier
// (sourced from M0 nav-graph.json decision nodes in production). Missing
// ADR → record emitted with adr_path empty (REQ-NS3-008 degrade). Output is
// sorted by Identifier for byte-stable emission (REQ-NS3-019).
func enumerateDecisions(projectRoot string, identifiers []string) []DecisionEnrichment {
	out := make([]DecisionEnrichment, 0, len(identifiers))
	for _, id := range identifiers {
		enr, _, ok := resolveADR(projectRoot, id)
		if ok {
			out = append(out, enr)
			continue
		}
		out = append(out, DecisionEnrichment{Identifier: id})
	}
	// Also scan the decisions dir for any superseded_by chains to populate
	// the SupersededBy field (parsed from the prior ADR's Status change OR
	// from the new ADR's supersedes: line).
	byNewSupersedes := map[string]string{} // priorID → newID
	for _, e := range out {
		if e.Supersedes != "" {
			byNewSupersedes[e.Supersedes] = e.Identifier
		}
	}
	for i, e := range out {
		if newID, found := byNewSupersedes[e.Identifier]; found && e.SupersededBy == "" {
			out[i].SupersededBy = newID
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Identifier < out[j].Identifier })
	return out
}
