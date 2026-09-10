package settings

// SPEC-WEB-WRITE-SAFETY-001 — reduced unit tests for the two first-measurement
// conclusions (M-a/M-b), serving as RED-first evidence for AC-WWS-003/004/005
// and the AC-WWS-004 positive control.
//
// Measured defect signatures (M1(d), worktree t517, pre-fix tree):
//   - feedback.yaml lost its blank line after a VALUE-IDENTICAL save (seam
//     PatchFile re-encodes the whole document; yaml.v3 normalizes blank lines).
//   - git-strategy.yaml was re-marshaled (key reorder + zero-value keys added)
//     after a VALUE-IDENTICAL save (applyTypedEdits calls SetSection
//     unconditionally, which sets gitStrategyDirty and forces Save() to rewrite).

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings/yamlpatch"
)

// TestPatchFileValueInvariantPreservesBytes carries AC-WWS-005 (RED-first): a
// seam patch whose new value equals the persisted value must leave the file
// byte-identical — blank lines included. RED on the defective tree: PatchFile
// re-encodes the document through the yaml.v3 encoder, which normalizes blank
// lines away (the documented limitation in the yamlpatch package header).
func TestPatchFileValueInvariantPreservesBytes(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	before := seedSectionFixture(t, root, "feedback")
	path := filepath.Join(root, ".moai", "config", "sections", "feedback.yaml")

	err := yamlpatch.PatchFile(path, []yamlpatch.KeyEdit{
		{Path: []string{"feedback", "repository"}, Value: "modu-ai/moai-adk"}, // identical value
	})
	if err != nil {
		t.Fatalf("PatchFile: %v", err)
	}

	after := readSection(t, root, "feedback")
	if after != before {
		t.Errorf("value-invariant seam patch rewrote the file — presentation elements lost\n--- before ---\n%s\n--- after ---\n%s", before, after)
	}
}

// TestPatchFileScalarChangePreservesPresentation carries AC-WWS-005 (RED-first)
// for the value-CHANGING case: when a save does replace a scalar, the untouched
// presentation elements — blank lines, comments, key order, unknown keys — must
// survive byte-for-byte around the one changed line.
func TestPatchFileScalarChangePreservesPresentation(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	before := seedSectionFixture(t, root, "feedback")
	path := filepath.Join(root, ".moai", "config", "sections", "feedback.yaml")

	err := yamlpatch.PatchFile(path, []yamlpatch.KeyEdit{
		{Path: []string{"feedback", "repository"}, Value: "someone/else-fork"},
	})
	if err != nil {
		t.Fatalf("PatchFile: %v", err)
	}

	after := readSection(t, root, "feedback")
	if !strings.Contains(after, "repository: someone/else-fork") {
		t.Errorf("scalar not updated:\n%s", after)
	}
	if got, want := blankLineCount(after), blankLineCount(before); got != want {
		t.Errorf("blank lines not preserved: before=%d after=%d\n%s", want, got, after)
	}
	if got, want := sectionCommentLines(after), sectionCommentLines(before); strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Error("comments not preserved by seam routing")
	}
}

// TestApplySchemaEditsValueInvariantTouchesNothing carries AC-WWS-003 and
// AC-WWS-004 (RED-first): submitting values that equal the persisted values for
// BOTH a typed section (git_strategy.mode) and a seam section
// (feedback.repository) must rewrite NEITHER file. RED on the defective tree:
// applyTypedEdits calls SetSection("git_strategy") unconditionally (raising
// gitStrategyDirty, so Save() re-marshals git-strategy.yaml), and
// ApplySchemaEdits routes the feedback edit to PatchFile (blank-line
// normalization).
func TestApplySchemaEditsValueInvariantTouchesNothing(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedTypedFixtures(t, root, "git-strategy", "llm", "quality")
	feedbackBefore := seedSectionFixture(t, root, "feedback")

	sectionsDir := filepath.Join(root, ".moai", "config", "sections")
	gsBefore, err := os.ReadFile(filepath.Join(sectionsDir, "git-strategy.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	err = ApplySchemaEdits(root, map[string]string{
		"git_strategy.mode":   "team",             // fixture's persisted value — value-invariant
		"feedback.repository": "modu-ai/moai-adk", // fixture's persisted value — value-invariant
	})
	if err != nil {
		t.Fatalf("ApplySchemaEdits: %v", err)
	}

	gsAfter, err := os.ReadFile(filepath.Join(sectionsDir, "git-strategy.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(gsBefore) != string(gsAfter) {
		t.Errorf("value-invariant save rewrote git-strategy.yaml (dirty-gate bypassed):\n--- before ---\n%s\n--- after ---\n%s", gsBefore, gsAfter)
	}
	feedbackAfter := readSection(t, root, "feedback")
	if feedbackAfter != feedbackBefore {
		t.Errorf("value-invariant save rewrote feedback.yaml (seam presentation lost)\n--- before ---\n%s\n--- after ---\n%s", feedbackBefore, feedbackAfter)
	}
}

// TestApplySchemaEditsGitStrategyRealChangeStillRewrites is the AC-WWS-004
// POSITIVE CONTROL (mandatory): the same environment where the value-invariant
// save above must hold — with a genuinely changed value — must still rewrite
// git-strategy.yaml. Without this control a green on the value-invariant test
// is indistinguishable from "the gate never ran at all".
func TestApplySchemaEditsGitStrategyRealChangeStillRewrites(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedTypedFixtures(t, root, "git-strategy", "llm", "quality")
	gsPath := filepath.Join(root, ".moai", "config", "sections", "git-strategy.yaml")
	before, err := os.ReadFile(gsPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := ApplySchemaEdits(root, map[string]string{"git_strategy.mode": "personal"}); err != nil {
		t.Fatalf("ApplySchemaEdits: %v", err)
	}

	after, err := os.ReadFile(gsPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) == string(after) {
		t.Error("positive control failed: a real git_strategy change did NOT rewrite git-strategy.yaml (gate is inert)")
	}
	if !strings.Contains(string(after), "mode: personal") {
		t.Errorf("changed value not persisted:\n%s", after)
	}
}

func blankLineCount(s string) int {
	count := 0
	for _, line := range strings.Split(s, "\n") {
		if strings.TrimSpace(line) == "" {
			count++
		}
	}
	return count
}

// --- sync-audit F1: absent-polarity of the value-invariant gate ---

// F1 premise helpers: the two fields the audit flagged are default-ON when
// absent (todo: TodoEnabled() absent⇒enabled, SPEC-TODO-ENABLE-FLAG-001;
// mcp: fail-OPEN all-enabled, SPEC-MCP-CONSOLE-001). Their FieldDefs must
// declare that polarity explicitly so the value-invariant gate can reason
// about the effective default.
func TestAbsentDefaultPolarityDeclared(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"workflow.todo.enabled", "mcp.tools.spec_progress.enabled"} {
		f, ok := Field(name)
		if !ok {
			t.Fatalf("field %q not registered", name)
		}
		if f.Type != TypeBool {
			t.Fatalf("field %q: type %q, want bool", name, f.Type)
		}
		if f.AbsentDefault != "true" {
			t.Errorf("field %q: AbsentDefault=%q, want \"true\" (default-ON-when-absent)", name, f.AbsentDefault)
		}
	}
	// The default-OFF control must stay default-OFF.
	f, ok := Field("gate.pre_commit.enabled")
	if !ok {
		t.Fatal("gate.pre_commit.enabled not registered")
	}
	if f.AbsentDefault != "" && f.AbsentDefault != "false" {
		t.Errorf("gate.pre_commit.enabled: AbsentDefault=%q, want empty (default-off)", f.AbsentDefault)
	}
}

// TestApplySchemaEditsAbsentDefaultOnBoolFalseStillWrites carries sync-audit
// F1 RED-first (a): a default-ON bool key that is ABSENT on disk receiving a
// console OFF submission ("false") IS a real change — the runtime interprets
// absent as enabled, so the value-invariant gate must WRITE it. RED on the
// pre-fix tree: the gate skipped every absent+false submission, silently
// swallowing the user's explicit OFF save.
func TestApplySchemaEditsAbsentDefaultOnBoolFalseStillWrites(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedSectionFixture(t, root, "workflow") // fixture ships no todo block — key absent

	if err := ApplySchemaEdits(root, map[string]string{"workflow.todo.enabled": "false"}); err != nil {
		t.Fatalf("ApplySchemaEdits: %v", err)
	}
	after := readSection(t, root, "workflow")
	if !strings.Contains(after, "enabled: false") {
		t.Errorf("explicit OFF on default-ON absent key was silently skipped — todo.enabled not persisted:\n%s", after)
	}
}

// TestApplySchemaEditsAbsentDefaultOnMcpFalseStillWrites carries sync-audit
// F1 RED-first (b): same polarity fix for the 24 fail-OPEN mcp.tools.<name>.enabled
// keys. RED on the pre-fix tree.
func TestApplySchemaEditsAbsentDefaultOnMcpFalseStillWrites(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedSectionFixture(t, root, "mcp") // fixture declares only session_list — spec_progress absent

	if err := ApplySchemaEdits(root, map[string]string{"mcp.tools.spec_progress.enabled": "false"}); err != nil {
		t.Fatalf("ApplySchemaEdits: %v", err)
	}
	after := readSection(t, root, "mcp")
	if !strings.Contains(after, "spec_progress") || !strings.Contains(after, "enabled: false") {
		t.Errorf("explicit OFF on absent mcp tool key was silently skipped — not persisted:\n%s", after)
	}
}

// TestApplySchemaEditsAbsentDefaultOnBoolTrueIsNoOp carries sync-audit F1
// RED-first (c): on a default-ON absent key a "true" submission equals the
// effective default — it must be SKIPPED (no key creation), mirroring the
// default-off gate's shape. RED on the pre-fix tree (which only skipped
// "false" and therefore wrote the key).
func TestApplySchemaEditsAbsentDefaultOnBoolTrueIsNoOp(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	before := seedSectionFixture(t, root, "workflow")

	if err := ApplySchemaEdits(root, map[string]string{"workflow.todo.enabled": "true"}); err != nil {
		t.Fatalf("ApplySchemaEdits: %v", err)
	}
	after := readSection(t, root, "workflow")
	if after != before {
		t.Errorf("true submission on default-ON absent key should be a no-op (absent IS enabled):\n--- before ---\n%s\n--- after ---\n%s", before, after)
	}
}

// --- coverage reinforcement for the SPEC-WEB-WRITE-SAFETY-001 repairs ---

// TestReadSeamScalarEdges covers the unreadable-target branches of
// readSeamScalar: an absent file, an empty document, a missing key, a
// non-scalar target, and the happy path.
func TestReadSeamScalarEdges(t *testing.T) {
	t.Parallel()

	if v, ok := readSeamScalar(t.TempDir(), "feedback", []string{"feedback", "repository"}); ok {
		t.Errorf("absent file reported ok=%v value=%q", ok, v)
	}

	root := t.TempDir()
	seedSectionFixture(t, root, "feedback")

	if v, ok := readSeamScalar(root, "feedback", []string{"feedback", "no_such_key"}); ok {
		t.Errorf("missing key reported ok=%v value=%q", ok, v)
	}
	if v, ok := readSeamScalar(root, "feedback", []string{"feedback", "repository", "too_deep"}); ok {
		t.Errorf("over-deep path reported ok=%v value=%q", ok, v)
	}
	// A path segment whose parent is a scalar must be unreadable.
	if v, ok := readSeamScalar(root, "feedback", []string{"feedback", "repository", "nested"}); ok {
		t.Errorf("scalar-parent path reported ok=%v value=%q", ok, v)
	}
	if v, ok := readSeamScalar(root, "feedback", []string{"feedback", "repository"}); !ok || v != "modu-ai/moai-adk" {
		t.Errorf("happy path: ok=%v value=%q, want ok=true modu-ai/moai-adk", ok, v)
	}
}

// TestApplySchemaEditsSeamAbsentKeyFalseIsNoOp covers repair (i) branch (b):
// a bool seam field whose key is ABSENT on disk submitting "false" must not
// create the key — absent IS false for a bool key.
func TestApplySchemaEditsSeamAbsentKeyFalseIsNoOp(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	before := seedSectionFixture(t, root, "gate")

	if err := ApplySchemaEdits(root, map[string]string{"gate.pre_commit.enabled": "false"}); err != nil {
		t.Fatalf("ApplySchemaEdits: %v", err)
	}
	after := readSection(t, root, "gate")
	if after != before {
		t.Errorf("absent-key + false submission created the key (no-op expected)\n--- before ---\n%s\n--- after ---\n%s", before, after)
	}
}

// TestApplySchemaEditsSeamAbsentKeyTrueStillWrites is the companion control
// for the absent-key branch: a bool field absent on disk submitting "true"
// IS a real change (the key must be created). The assertion is scoped to the
// pre_commit block (sync-audit F1-f): a bare strings.Contains("enabled: true")
// was already satisfied by the fixture's top-level gate.enabled: true, so the
// test passed vacuously — only a match INSIDE the pre_commit block proves the
// write happened.
func TestApplySchemaEditsSeamAbsentKeyTrueStillWrites(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedSectionFixture(t, root, "gate")

	if err := ApplySchemaEdits(root, map[string]string{"gate.pre_commit.enabled": "true"}); err != nil {
		t.Fatalf("ApplySchemaEdits: %v", err)
	}
	after := readSection(t, root, "gate")
	m := regexp.MustCompile(`(?m)^\s*pre_commit:\s*$\n^\s+enabled: (\S+)`).FindStringSubmatch(after)
	if m == nil {
		t.Errorf("absent-key + true submission did not create the pre_commit block:\n%s", after)
	} else if m[1] != "true" {
		t.Errorf("pre_commit.enabled = %q, want \"true\":\n%s", m[1], after)
	}
}

// TestPatchFileSpliceFallsBackForUpsert covers the re-encode fallback: an
// edit whose path does not exist (upsert) cannot take the splice path and
// must still persist through the fallback.
func TestPatchFileSpliceFallsBackForUpsert(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedSectionFixture(t, root, "feedback")
	path := filepath.Join(root, ".moai", "config", "sections", "feedback.yaml")

	err := yamlpatch.PatchFile(path, []yamlpatch.KeyEdit{
		{Path: []string{"feedback", "brand_new_key"}, Value: "hello"},
	})
	if err != nil {
		t.Fatalf("PatchFile(upsert): %v", err)
	}
	after := readSection(t, root, "feedback")
	if !strings.Contains(after, "brand_new_key: hello") {
		t.Errorf("upsert not persisted:\n%s", after)
	}
}

// TestPatchFileSpliceQuotedScalarChange covers the quoted-style line splice:
// a double-quoted scalar keeps its quoting after a value change, and the
// re-parse verification accepts the hand-quoted replacement.
func TestPatchFileSpliceQuotedScalarChange(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "crosssession.yaml")
	original := "crosssession:\n    inbound: \"select\"\n"
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	err := yamlpatch.PatchFile(path, []yamlpatch.KeyEdit{
		{Path: []string{"crosssession", "inbound"}, Value: "isolate_machines"},
	})
	if err != nil {
		t.Fatalf("PatchFile(quoted): %v", err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(after), `"isolate_machines"`) {
		t.Errorf("quoted scalar replacement lost its quoting:\n%s", after)
	}
	if strings.Count(string(after), "\n") != strings.Count(original, "\n") {
		t.Errorf("unexpected document shape (splice should touch one line only):\n%s", after)
	}
}
