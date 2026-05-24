// Package proposalgen scaffolder unit tests.
// SPEC: SPEC-V3R6-HARNESS-PROPOSAL-GEN-001 REQ-PGN-006..007.
package proposalgen

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestScaffolder_NoOpSkipsCreation verifies AC-PGN-003 / REQ-PGN-007: an
// empty candidate slice triggers ZERO filesystem writes — the .moai/proposals
// directory is not created.
func TestScaffolder_NoOpSkipsCreation(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	outDir := filepath.Join(root, ".moai", "proposals")
	written, err := WriteProposals(outDir, nil)
	if err != nil {
		t.Fatalf("WriteProposals(empty) error = %v, want nil", err)
	}
	if len(written) != 0 {
		t.Errorf("written = %v, want empty slice", written)
	}
	if _, statErr := os.Stat(outDir); !errors.Is(statErr, fs.ErrNotExist) {
		t.Errorf(".moai/proposals MUST NOT exist on no-op path; stat err = %v", statErr)
	}
}

// TestScaffolder_CreatesDirectoryAndFiles verifies REQ-PGN-006: a single
// candidate produces .moai/proposals/<draft-id>/spec.md and proposal.json.
func TestScaffolder_CreatesDirectoryAndFiles(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	outDir := filepath.Join(root, ".moai", "proposals")
	candidate := ProposalCandidate{
		PatternKey:       "code_change:func_extract:auth_module",
		ObservationCount: 7,
		Confidence:       0.85,
		Tier:             "recommendation",
		SourceTs:         time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC),
		DraftID:          "PROPOSAL-20260524-deadbeef",
	}
	written, err := WriteProposals(outDir, []ProposalCandidate{candidate})
	if err != nil {
		t.Fatalf("WriteProposals: %v", err)
	}
	if len(written) != 1 {
		t.Fatalf("len(written) = %d, want 1", len(written))
	}

	draftDir := filepath.Join(outDir, candidate.DraftID)
	info, err := os.Stat(draftDir)
	if err != nil {
		t.Fatalf("stat draftDir: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("draftDir is not a directory: %v", info.Mode())
	}
	mode := info.Mode().Perm()
	if mode != 0o755 {
		t.Errorf("draftDir permission = %o, want 0755", mode)
	}

	// spec.md exists and contains the Origin section with pattern_key
	// observation_count and confidence.
	specBytes, err := os.ReadFile(filepath.Join(draftDir, "spec.md"))
	if err != nil {
		t.Fatalf("read spec.md: %v", err)
	}
	specStr := string(specBytes)
	if !strings.Contains(specStr, "## Origin") {
		t.Errorf("spec.md missing `## Origin` section")
	}
	if !strings.Contains(specStr, candidate.PatternKey) {
		t.Errorf("spec.md does not reference pattern_key %q", candidate.PatternKey)
	}
	if !strings.Contains(specStr, "observation_count: 7") {
		t.Errorf("spec.md does not include observation_count: 7")
	}
	if !strings.Contains(specStr, "confidence: 0.85") {
		t.Errorf("spec.md does not include confidence: 0.85")
	}
	if !strings.Contains(specStr, "status: draft") {
		t.Errorf("spec.md frontmatter missing status: draft (REQ-PGN-006)")
	}
	if !strings.Contains(specStr, "lifecycle: exploratory") {
		t.Errorf("spec.md frontmatter missing lifecycle: exploratory (REQ-PGN-006)")
	}
	if !strings.Contains(specStr, "priority: P3") {
		t.Errorf("spec.md frontmatter missing priority: P3 (REQ-PGN-006)")
	}

	// proposal.json exists and parses with the expected schema.
	propBytes, err := os.ReadFile(filepath.Join(draftDir, "proposal.json"))
	if err != nil {
		t.Fatalf("read proposal.json: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(propBytes, &parsed); err != nil {
		t.Fatalf("proposal.json is not valid JSON: %v", err)
	}
	required := []string{
		"pattern_key", "observation_count", "confidence",
		"tier", "generated_at", "generator_version",
	}
	for _, k := range required {
		if _, ok := parsed[k]; !ok {
			t.Errorf("proposal.json missing required field %q (REQ-PGN-006)", k)
		}
	}
	if got, want := parsed["pattern_key"], candidate.PatternKey; got != want {
		t.Errorf("proposal.json pattern_key = %v, want %v", got, want)
	}
	if got, want := parsed["generator_version"], GeneratorVersion; got != want {
		t.Errorf("proposal.json generator_version = %v, want %v", got, want)
	}
}

// TestScaffolder_LanguageNeutralBody verifies REQ-PGN-006 + plan.md §C.1:
// the rendered spec.md body MUST be language-neutral (no Go/Python/TypeScript
// syntax).
func TestScaffolder_LanguageNeutralBody(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	outDir := filepath.Join(root, ".moai", "proposals")
	candidate := ProposalCandidate{
		PatternKey: "code_change:func_extract:auth_module",
		ObservationCount: 5, Confidence: 0.9, Tier: "recommendation",
		SourceTs: time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC),
		DraftID:  "PROPOSAL-20260524-cafebabe",
	}
	if _, err := WriteProposals(outDir, []ProposalCandidate{candidate}); err != nil {
		t.Fatalf("WriteProposals: %v", err)
	}

	specBytes, err := os.ReadFile(filepath.Join(outDir, candidate.DraftID, "spec.md"))
	if err != nil {
		t.Fatalf("read spec.md: %v", err)
	}
	body := string(specBytes)

	// Forbidden tokens — surface accidental language-specific code blocks.
	forbidden := []string{
		"```go", "```python", "```typescript", "```javascript",
		"```rust", "```java", "func main(", "def main(",
	}
	for _, token := range forbidden {
		if strings.Contains(body, token) {
			t.Errorf("spec.md body contains forbidden language-specific token %q", token)
		}
	}
}

// TestScaffolder_PreExistingProposalsPreserved verifies the §C edge case: a
// pre-existing .moai/proposals/ directory with sibling proposals is not
// disturbed by new candidate writes.
func TestScaffolder_PreExistingProposalsPreserved(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	outDir := filepath.Join(root, ".moai", "proposals")
	siblingDir := filepath.Join(outDir, "PROPOSAL-20260101-existing0")
	if err := os.MkdirAll(siblingDir, 0o755); err != nil {
		t.Fatalf("setup MkdirAll: %v", err)
	}
	siblingFile := filepath.Join(siblingDir, "preexisting.txt")
	if err := os.WriteFile(siblingFile, []byte("keep me"), 0o644); err != nil {
		t.Fatalf("setup WriteFile: %v", err)
	}

	candidate := ProposalCandidate{
		PatternKey: "error_pattern:nil_deref:payment_handler",
		ObservationCount: 4, Confidence: 0.9, Tier: "approval_required",
		SourceTs: time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC),
		DraftID:  "PROPOSAL-20260524-feedface",
	}
	if _, err := WriteProposals(outDir, []ProposalCandidate{candidate}); err != nil {
		t.Fatalf("WriteProposals: %v", err)
	}

	// Sibling untouched.
	siblingBytes, err := os.ReadFile(siblingFile)
	if err != nil {
		t.Fatalf("sibling file disappeared: %v", err)
	}
	if string(siblingBytes) != "keep me" {
		t.Errorf("sibling contents changed: %q", siblingBytes)
	}

	// New candidate written.
	if _, err := os.Stat(filepath.Join(outDir, candidate.DraftID, "spec.md")); err != nil {
		t.Errorf("new candidate spec.md missing: %v", err)
	}
}

// TestScaffolder_IdempotentOverwrite verifies that re-running the scaffolder
// on the same candidate produces consistent results (REQ-PGN-006 idempotency
// via deterministic DraftID).
func TestScaffolder_IdempotentOverwrite(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	outDir := filepath.Join(root, ".moai", "proposals")
	candidate := ProposalCandidate{
		PatternKey: "tool_failure:bash_timeout:db_migrate",
		ObservationCount: 6, Confidence: 0.82, Tier: "recommendation",
		SourceTs: time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC),
		DraftID:  "PROPOSAL-20260524-aaaaaaaa",
	}

	// First run.
	if _, err := WriteProposals(outDir, []ProposalCandidate{candidate}); err != nil {
		t.Fatalf("WriteProposals first: %v", err)
	}
	firstSpec, _ := os.ReadFile(filepath.Join(outDir, candidate.DraftID, "spec.md"))

	// Second run with identical input.
	if _, err := WriteProposals(outDir, []ProposalCandidate{candidate}); err != nil {
		t.Fatalf("WriteProposals second: %v", err)
	}
	secondSpec, _ := os.ReadFile(filepath.Join(outDir, candidate.DraftID, "spec.md"))

	if string(firstSpec) != string(secondSpec) {
		t.Errorf("spec.md content changed across idempotent runs")
	}
}

// TestScaffolder_RejectsEmptyDraftID guards against a downstream regression
// where mapper neglects to populate DraftID; the scaffolder must surface
// this as an error rather than write to .moai/proposals/.
func TestScaffolder_RejectsEmptyDraftID(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	outDir := filepath.Join(root, ".moai", "proposals")
	candidate := ProposalCandidate{
		PatternKey: "code_change:func_extract:auth_module",
		ObservationCount: 5, Confidence: 0.85, Tier: "recommendation",
		SourceTs: time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC),
		DraftID:  "", // empty — must be rejected
	}
	if _, err := WriteProposals(outDir, []ProposalCandidate{candidate}); err == nil {
		t.Errorf("WriteProposals accepted empty DraftID; want error")
	}
}
