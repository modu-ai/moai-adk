package settings

// SPEC-WEB-SAVE-LOSSLESS-001 — lossless-save contract tests (REQ-WSL-001/002/
// 003/004). RED-first against the pre-implementation tree, where typed section
// edits route through LoadRaw → SetSection → Save (a full re-marshal of the
// edited file) and an absent-key "" submission is persisted as an upsert.
//
// The lossless spine under test: a one-field edit changes exactly that field's
// line; unmodeled keys and comments survive every writer; an absent key plus an
// empty submission writes nothing.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings/yamlpatch"
	"gopkg.in/yaml.v3"
)

// seedLosslessFixture copies a testdata section fixture into the temp project
// and returns the seeded bytes. A postfix hook lets a test inject unmodeled
// content (the GitHub #1731 issue-named keys) into the seeded file.
func seedLosslessFixture(t *testing.T, root, name, postfix string) string {
	t.Helper()
	base := seedSectionFixture(t, root, name)
	if postfix == "" {
		return base
	}
	path := filepath.Join(root, ".moai", "config", "sections", name+".yaml")
	if err := os.WriteFile(path, []byte(base+postfix), 0o644); err != nil {
		t.Fatal(err)
	}
	return base + postfix
}

// changedLineIndexes returns the indexes of lines that differ between before
// and after. Both documents must have the same line count; a different count
// fails the test (a line splice never adds or removes lines).
func changedLineIndexes(t *testing.T, before, after string) []int {
	t.Helper()
	b := strings.Split(before, "\n")
	a := strings.Split(after, "\n")
	if len(a) != len(b) {
		t.Fatalf("line count changed: before=%d after=%d\n--- before ---\n%s\n--- after ---\n%s", len(b), len(a), before, after)
	}
	var changed []int
	for i := range b {
		if a[i] != b[i] {
			changed = append(changed, i)
		}
	}
	return changed
}

// assertExactlyOneLineChanged is the AC-WSL-002 line-splice predicate: the
// diff between before and after is exactly one line, and that line carries the
// wanted substring (the edited key: value).
func assertExactlyOneLineChanged(t *testing.T, before, after, wantInChangedLine string) {
	t.Helper()
	changed := changedLineIndexes(t, before, after)
	if len(changed) != 1 {
		t.Fatalf("expected exactly 1 changed line, got %d: %v\n--- before ---\n%s\n--- after ---\n%s", len(changed), changed, before, after)
	}
	line := strings.Split(after, "\n")[changed[0]]
	if !strings.Contains(line, wantInChangedLine) {
		t.Fatalf("changed line %q does not carry %q\n--- before ---\n%s\n--- after ---\n%s", line, wantInChangedLine, before, after)
	}
}

// TestApplySchemaEditsOneFieldEditSplicesOneLine carries AC-WSL-002 (line-splice
// variant) for a typed section: editing llm.glm.models.high must change ONLY
// that line of llm.yaml. RED on the pre-implementation tree: the typed
// LoadRaw → SetSection → Save path re-marshals the whole file.
func TestApplySchemaEditsOneFieldEditSplicesOneLine(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	before := seedLosslessFixture(t, root, "llm", "")
	if err := ApplySchemaEdits(root, map[string]string{"llm.glm.models.high": "glm-5.3-t"}); err != nil {
		t.Fatalf("ApplySchemaEdits: %v", err)
	}
	after := readSection(t, root, "llm")
	assertExactlyOneLineChanged(t, before, after, "high: glm-5.3-t")
}

// TestApplySchemaEditsGitStrategyEditPreservesUnmodeledKeysAndComments carries
// AC-WSL-002 + AC-WSL-003 for git-strategy.yaml: a mode edit changes one line;
// the fixture's Korean comments and unmodeled keys (github_username,
// required_reviews) survive verbatim. RED on the pre-implementation tree
// (typed Save re-marshal reorders keys and rewrites comments).
func TestApplySchemaEditsGitStrategyEditPreservesUnmodeledKeysAndComments(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	before := seedLosslessFixture(t, root, "git-strategy", "")
	if err := ApplySchemaEdits(root, map[string]string{"git_strategy.mode": "personal"}); err != nil {
		t.Fatalf("ApplySchemaEdits: %v", err)
	}
	after := readSection(t, root, "git-strategy")
	assertExactlyOneLineChanged(t, before, after, "mode: personal")
	if got, want := sectionCommentLines(after), sectionCommentLines(before); strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("comments not preserved:\n--- before ---\n%s\n--- after ---\n%s", strings.Join(want, "\n"), strings.Join(got, "\n"))
	}
	for _, key := range []string{"github_username", "required_reviews: 0"} {
		if !strings.Contains(after, key) {
			t.Errorf("unmodeled key %q lost by the edit:\n%s", key, after)
		}
	}
}

// TestApplySchemaEditsQualityEditPreservesIssueNamedKey carries AC-WSL-003 and
// the AC-WSL-009 schema-path variant: editing a quality bool must not touch the
// issue-named unmodeled key `constitution.session_effort_default` (with its
// trailing comment) or any other quality.yaml byte. RED on the
// pre-implementation tree: the quality edit rode the typed Save re-marshal,
// which dropped the unmodeled key and added quality_extras_enabled.
func TestApplySchemaEditsQualityEditPreservesIssueNamedKey(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	// seedSectionFixture already carries the full constitution block; splice
	// the unmodeled key under the existing block rather than appending a
	// second constitution top-level key.
	base := seedSectionFixture(t, root, "quality")
	doctored := strings.Replace(base,
		"    development_mode: tdd\n",
		"    development_mode: tdd\n    session_effort_default: xhigh  # local note\n",
		1)
	if doctored == base {
		t.Fatalf("fixture anchor line not found; cannot seed the issue-named key")
	}
	path := filepath.Join(root, ".moai", "config", "sections", "quality.yaml")
	if err := os.WriteFile(path, []byte(doctored), 0o644); err != nil {
		t.Fatal(err)
	}
	before := doctored

	if err := ApplySchemaEdits(root, map[string]string{
		"quality.ddd_settings.characterization_tests": "false", // fixture true → real change
	}); err != nil {
		t.Fatalf("ApplySchemaEdits: %v", err)
	}
	after := readSection(t, root, "quality")

	assertExactlyOneLineChanged(t, before, after, "characterization_tests: false")
	if !strings.Contains(after, "session_effort_default: xhigh  # local note") {
		t.Errorf("issue-named unmodeled key lost by a quality edit:\n%s", after)
	}
	if strings.Contains(after, "quality_extras_enabled") {
		t.Errorf("unsubmitted quality_extras_enabled key was written (REQ-WSL-002 violation):\n%s", after)
	}
}

// TestApplySchemaEditsAbsentKeyEmptySubmissionIsNoOp carries AC-WSL-004: an
// EmptySubmits workflow audit pin whose key is ABSENT on disk, submitted as "",
// must not add a `model: ""` key — byte-identical. RED on the
// pre-implementation tree: the seam no-op gate only compared EXISTING keys, so
// absent + "" was upserted.
func TestApplySchemaEditsAbsentKeyEmptySubmissionIsNoOp(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	before := seedLosslessFixture(t, root, "workflow", "")

	if err := ApplySchemaEdits(root, map[string]string{"workflow.audit.claude.model": ""}); err != nil {
		t.Fatalf("ApplySchemaEdits: %v", err)
	}
	after := readSection(t, root, "workflow")
	if after != before {
		t.Errorf("absent-key empty submission rewrote workflow.yaml (no-op expected)\n--- before ---\n%s\n--- after ---\n%s", before, after)
	}
	if strings.Contains(after, "model: \"\"") {
		t.Errorf("absent-key empty submission created a model: \"\" key:\n%s", after)
	}
}

// TestApplySchemaEditsExistingKeyEmptySubmissionStillWrites carries the
// AC-WSL-004 (a) variant: the key EXISTS → a "" submission is the documented
// delete semantics and MUST be written (TestCrossSessionEmptySubmitsRoundTrip
// lineage). This is the negative control for the no-op gate above. The
// assertion parses the audit block (yaml renders an empty string scalar as a
// bare `model:`, so a literal `model: ""` grep cannot prove the write).
func TestApplySchemaEditsExistingKeyEmptySubmissionStillWrites(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedLosslessFixture(t, root, "workflow", "")

	if err := ApplySchemaEdits(root, map[string]string{"workflow.audit.claude.model": "sonnet"}); err != nil {
		t.Fatalf("ApplySchemaEdits(set): %v", err)
	}
	var wrapper struct {
		Workflow struct {
			Audit struct {
				Claude struct {
					Model string `yaml:"model"`
				} `yaml:"claude"`
			} `yaml:"audit"`
		} `yaml:"workflow"`
	}
	raw := readSection(t, root, "workflow")
	if err := yaml.Unmarshal([]byte(raw), &wrapper); err != nil {
		t.Fatalf("parse after set: %v", err)
	}
	if wrapper.Workflow.Audit.Claude.Model != "sonnet" {
		t.Fatalf("set step did not persist the value (positive precondition): %q", wrapper.Workflow.Audit.Claude.Model)
	}

	if err := ApplySchemaEdits(root, map[string]string{"workflow.audit.claude.model": ""}); err != nil {
		t.Fatalf("ApplySchemaEdits(clear): %v", err)
	}
	raw = readSection(t, root, "workflow")
	// A fresh wrapper: yaml renders the cleared scalar as a null node, and a
	// null leaves an already-populated struct field unchanged — reusing the
	// wrapper would misreport the clear as a no-op.
	var cleared struct {
		Workflow struct {
			Audit struct {
				Claude struct {
					Model string `yaml:"model"`
				} `yaml:"claude"`
			} `yaml:"audit"`
		} `yaml:"workflow"`
	}
	if err := yaml.Unmarshal([]byte(raw), &cleared); err != nil {
		t.Fatalf("parse after clear: %v", err)
	}
	if cleared.Workflow.Audit.Claude.Model != "" {
		t.Errorf("existing-key empty submission was not written (delete semantics lost): %q", cleared.Workflow.Audit.Claude.Model)
	}
	if !strings.Contains(raw, "audit:") {
		t.Errorf("audit block vanished after the clear:\n%s", raw)
	}
}

// TestApplySchemaEditsAbsentKeyRealValueStillUpserts carries the AC-WSL-004 (b)
// variant: absent key + a REAL value must still be written (upsert).
func TestApplySchemaEditsAbsentKeyRealValueStillUpserts(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedLosslessFixture(t, root, "workflow", "")

	if err := ApplySchemaEdits(root, map[string]string{"workflow.audit.claude.model": "sonnet"}); err != nil {
		t.Fatalf("ApplySchemaEdits: %v", err)
	}
	after := readSection(t, root, "workflow")
	if !strings.Contains(after, "model: sonnet") {
		t.Errorf("absent-key real-value submission was not persisted:\n%s", after)
	}
}

// TestNestedSeamEditPreservesIssueNamedKey carries AC-WSL-009 for the nested
// write seam: a coverage-target edit must change exactly one quality.yaml line
// and leave `constitution.session_effort_default` (unmodeled, commented)
// intact. RED on the pre-implementation tree: WriteProjectNestedConfig rode
// SetSection("quality") → Save (full re-marshal).
func TestNestedSeamEditPreservesIssueNamedKey(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	base := seedSectionFixture(t, root, "quality")
	doctored := strings.Replace(base,
		"    development_mode: tdd\n",
		"    development_mode: tdd\n    session_effort_default: xhigh  # local note\n",
		1)
	path := filepath.Join(root, ".moai", "config", "sections", "quality.yaml")
	if err := os.WriteFile(path, []byte(doctored), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := WriteProjectNestedConfig(root, NestedForm{CoverageTarget: 92, CoverageTargetSet: true}); err != nil {
		t.Fatalf("WriteProjectNestedConfig: %v", err)
	}
	after := readSection(t, root, "quality")

	assertExactlyOneLineChanged(t, doctored, after, "test_coverage_target: 92")
	if !strings.Contains(after, "session_effort_default: xhigh  # local note") {
		t.Errorf("issue-named unmodeled key lost by a nested edit:\n%s", after)
	}
}

// TestPatchFileBrokenFilePreflightFailsNoWrite carries AC-WSL-006 (a)
// pre-flight leg (REQ-WSL-006): a target file that exists but is UNPARSEABLE
// (broken/corrupted yaml) must fail the write BEFORE any bytes are touched —
// PatchFile tolerates absent files (greenfield seeding) but a present file it
// cannot parse is exactly the corrupt-file case, and writing over it would
// destroy the only recoverable copy.
func TestPatchFileBrokenFilePreflightFailsNoWrite(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	path := filepath.Join(root, ".moai", "config", "sections", "workflow.yaml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	broken := "workflow:\n  audit: [unclosed\n\tbroken: ::::\n"
	if err := os.WriteFile(path, []byte(broken), 0o644); err != nil {
		t.Fatal(err)
	}

	err := WriteSectionViaSeam(root, "workflow", []yamlpatch.KeyEdit{
		{Path: []string{"workflow", "audit"}, Value: "x"},
	})
	if err == nil {
		t.Fatal("broken target file: want pre-flight error, got nil")
	}
	after, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(after) != broken {
		t.Errorf("failed pre-flight mutated the broken file (bytes must be untouched):\n%s", after)
	}
}
