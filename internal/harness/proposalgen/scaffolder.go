// SPEC: SPEC-V3R6-HARNESS-PROPOSAL-GEN-001 REQ-PGN-006..007.
//
// scaffolder.go renders each ProposalCandidate into a self-contained directory
// at .moai/proposals/<draft-id>/ containing two files:
//
//   - spec.md: language-neutral draft SPEC body with EARS-style placeholders,
//     downstream manager-spec authoring target.
//   - proposal.json: structured metadata for orchestrator + downstream tools.
//
// The scaffolder is strictly no-op on an empty candidate slice — the
// .moai/proposals/ directory is NOT created. This preserves the V3R4
// graceful no-op contract on current data.
package proposalgen

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// stamps carries the wall-clock-derived fields of a rendered draft. When a
// draft already exists on disk its stamps are reused verbatim, so re-running
// the scaffolder on an unchanged candidate is byte-idempotent (REQ-PGN-006)
// instead of churning .moai/proposals/ with a fresh timestamp on every run.
type stamps struct {
	created     string // YYYY-MM-DD
	updated     string // YYYY-MM-DD
	generatedAt string // RFC3339
}

// freshStamps derives a stamp set from the current wall clock. It is used only
// when no readable draft exists yet at the target path.
func freshStamps(now time.Time) stamps {
	return stamps{
		created:     now.Format("2006-01-02"),
		updated:     now.Format("2006-01-02"),
		generatedAt: now.Format(time.RFC3339),
	}
}

// existingStamps reads back the stamps of a previously written draft. spec.md
// is the single source for all three values, so the sibling proposal.json is
// re-rendered with the same generated_at rather than its own. A missing or
// unparseable file yields ok=false and the caller falls back to freshStamps —
// a corrupt draft must not fail the write.
func existingStamps(draftDir string) (stamps, bool) {
	body, err := os.ReadFile(filepath.Join(draftDir, "spec.md"))
	if err != nil {
		return stamps{}, false
	}
	var s stamps
	for _, line := range strings.Split(string(body), "\n") {
		switch {
		case s.created == "" && strings.HasPrefix(line, "created: "):
			s.created = strings.TrimSpace(strings.TrimPrefix(line, "created: "))
		case s.updated == "" && strings.HasPrefix(line, "updated: "):
			s.updated = strings.TrimSpace(strings.TrimPrefix(line, "updated: "))
		case s.generatedAt == "" && strings.HasPrefix(line, "- generated_at: "):
			s.generatedAt = strings.TrimSpace(strings.TrimPrefix(line, "- generated_at: "))
		}
	}
	if s.created == "" || s.updated == "" || s.generatedAt == "" {
		return stamps{}, false
	}
	return s, true
}

// WriteProposals renders each candidate under outDir/<draft-id>/. On an
// empty input the function returns immediately without creating outDir.
//
// Return value is the slice of draft IDs successfully written, in the same
// order as the input. Errors are wrapped with the draft ID that failed so
// the CLI can surface a precise diagnostic.
func WriteProposals(outDir string, candidates []ProposalCandidate) ([]string, error) {
	if len(candidates) == 0 {
		return nil, nil
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, fmt.Errorf("proposalgen scaffolder: mkdir %q: %w", outDir, err)
	}

	written := make([]string, 0, len(candidates))
	for _, c := range candidates {
		if c.DraftID == "" {
			return written, errors.New("proposalgen scaffolder: candidate has empty DraftID")
		}
		draftDir := filepath.Join(outDir, c.DraftID)
		if err := os.MkdirAll(draftDir, 0o755); err != nil {
			return written, fmt.Errorf("proposalgen scaffolder: mkdir %q: %w", draftDir, err)
		}
		st, ok := existingStamps(draftDir)
		if !ok {
			st = freshStamps(time.Now().UTC())
		}
		if err := os.WriteFile(filepath.Join(draftDir, "spec.md"), []byte(renderSpecMd(c, st)), 0o644); err != nil {
			return written, fmt.Errorf("proposalgen scaffolder: write spec.md for %s: %w", c.DraftID, err)
		}
		propBytes, err := marshalProposalJSON(c, st)
		if err != nil {
			return written, fmt.Errorf("proposalgen scaffolder: marshal proposal.json for %s: %w", c.DraftID, err)
		}
		if err := os.WriteFile(filepath.Join(draftDir, "proposal.json"), propBytes, 0o644); err != nil {
			return written, fmt.Errorf("proposalgen scaffolder: write proposal.json for %s: %w", c.DraftID, err)
		}
		written = append(written, c.DraftID)
	}
	return written, nil
}

// renderSpecMd produces the language-neutral spec.md body for a candidate.
// The body intentionally contains no Go/Python/TypeScript syntax — only
// natural-language EARS-style requirement placeholders. Language selection
// is downstream (manager-spec / manager-develop concern).
func renderSpecMd(c ProposalCandidate, st stamps) string {
	return "---\n" +
		"id: " + c.DraftID + "\n" +
		"title: \"Draft proposal — " + c.PatternKey + "\"\n" +
		"version: \"0.1.0\"\n" +
		"status: draft\n" +
		"created: " + st.created + "\n" +
		"updated: " + st.updated + "\n" +
		"author: proposalgen\n" +
		"priority: P3\n" +
		"phase: \"exploratory\"\n" +
		"module: \"TBD\"\n" +
		"lifecycle: exploratory\n" +
		"tags: \"proposal, harness-learning-loop, auto-generated\"\n" +
		"---\n\n" +
		"# " + c.DraftID + " — Draft Proposal\n\n" +
		"## Origin\n\n" +
		"This draft was generated automatically by the V3R4 self-evolving harness " +
		"learning loop from a recurring pattern observed in the local usage history.\n\n" +
		"- pattern_key: `" + c.PatternKey + "`\n" +
		"- observation_count: " + itoa(c.ObservationCount) + "\n" +
		"- confidence: " + ftoa(c.Confidence) + "\n" +
		"- tier: " + c.Tier + "\n" +
		"- source_ts: " + c.SourceTs.UTC().Format(time.RFC3339) + "\n" +
		"- generated_at: " + st.generatedAt + "\n" +
		"- generator_version: " + GeneratorVersion + "\n\n" +
		"## §2. Purpose & Background\n\n" +
		"_TBD by author. Describe the problem context that this proposal is meant " +
		"to address. The pattern observation above suggests a recurring code " +
		"change, error, tool failure, or repeated edit; convert that signal into a " +
		"first-class requirement statement._\n\n" +
		"## §3. EARS Requirements (placeholder)\n\n" +
		"- **REQ-001** (Ubiquitous): _TBD — the system SHALL..._\n" +
		"- **REQ-002** (Event-Driven): _TBD — WHEN <trigger> the system SHALL..._\n\n" +
		"## §4. Acceptance Reference\n\n" +
		"_TBD — link to acceptance.md once authored._\n\n" +
		"## §5. Out of Scope\n\n" +
		"### §5.1 Out of Scope — Implementation Language\n\n" +
		"This draft is language-neutral. Implementation language selection is a " +
		"downstream manager-spec / manager-develop concern. No Go, Python, " +
		"TypeScript, or other language-specific code appears in this body.\n\n" +
		"### §5.2 Out of Scope — Automatic Acceptance\n\n" +
		"This draft is NOT yet a committed SPEC. The orchestrator AskUserQuestion " +
		"gate (Approve / Modify / Reject) determines whether this draft enters the " +
		"plan-phase pipeline.\n"
}

// marshalProposalJSON encodes the proposal.json metadata payload.
func marshalProposalJSON(c ProposalCandidate, st stamps) ([]byte, error) {
	payload := map[string]any{
		"pattern_key":       c.PatternKey,
		"observation_count": c.ObservationCount,
		"confidence":        c.Confidence,
		"tier":              c.Tier,
		"source_ts":         c.SourceTs.UTC().Format(time.RFC3339),
		"generated_at":      st.generatedAt,
		"generator_version": GeneratorVersion,
		"draft_id":          c.DraftID,
	}
	out, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, err
	}
	// Trailing newline for POSIX-friendly file termination.
	return append(out, '\n'), nil
}

// itoa formats an integer for inclusion in the spec.md Origin section.
func itoa(n int) string {
	return strconv.Itoa(n)
}

// ftoa formats a float64 for inclusion in the spec.md Origin section.
// Uses JSON encoding to mirror the precision shown in proposal.json.
func ftoa(f float64) string {
	b, _ := json.Marshal(f)
	return string(b)
}
