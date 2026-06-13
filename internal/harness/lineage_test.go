// Package harness — lineage.go + Apply()-lineage-integration tests (M6 auditable lineage).
// SPEC-HARNESS-LOOP-CLOSURE-001 REQ-HLC-001..010.
//
// All fixtures use t.TempDir() so no live .moai/harness/ runtime artifact is mutated.
// The manifest path is injected via newApplierWithManifest so the Apply() lineage
// write targets a temp file rather than the production
// .moai/harness/learning-history/manifest.jsonl path.
package harness

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ─────────────────────────────────────────────
// lineage.go unit tests (writer + loader)
// ─────────────────────────────────────────────

// TestLoadManifest_MissingFileReturnsEmpty verifies LoadManifest on a non-existent
// path returns an empty slice with no error (backward compat — REQ-HLC-005).
func TestLoadManifest_MissingFileReturnsEmpty(t *testing.T) {
	t.Parallel()

	missing := filepath.Join(t.TempDir(), "does-not-exist", "manifest.jsonl")
	entries, err := LoadManifest(missing)
	if err != nil {
		t.Fatalf("LoadManifest on missing file returned error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("LoadManifest on missing file len = %d, want 0", len(entries))
	}
}

// TestLineage_AppendOnly verifies WriteLineageEntry appends entries in write order
// and LoadManifest round-trips each entry's fields (REQ-HLC-001/002).
func TestLineage_AppendOnly(t *testing.T) {
	t.Parallel()

	manifestPath := filepath.Join(t.TempDir(), "learning-history", "manifest.jsonl")

	e1 := LineageEntry{
		ProposalID:     "p-001",
		TargetPath:     ".claude/skills/my-harness-a/SKILL.md",
		AppliedSurface: "description",
		Decision:       "approved",
		Reason:         "first transition",
	}
	e2 := LineageEntry{
		ProposalID: "p-002",
		TargetPath: ".claude/agents/moai/x.md",
		Decision:   "rejected",
		Reason:     "L1 frozen guard",
	}

	if err := WriteLineageEntry(manifestPath, e1); err != nil {
		t.Fatalf("WriteLineageEntry e1: %v", err)
	}
	if err := WriteLineageEntry(manifestPath, e2); err != nil {
		t.Fatalf("WriteLineageEntry e2: %v", err)
	}

	got, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("LoadManifest len = %d, want 2", len(got))
	}

	// Write order preserved.
	if got[0].ProposalID != "p-001" || got[1].ProposalID != "p-002" {
		t.Errorf("write order not preserved: got[0]=%q got[1]=%q", got[0].ProposalID, got[1].ProposalID)
	}
	// Round-trip fields on the approved entry.
	if got[0].Decision != "approved" {
		t.Errorf("got[0].Decision = %q, want approved", got[0].Decision)
	}
	if got[0].AppliedSurface != "description" {
		t.Errorf("got[0].AppliedSurface = %q, want description", got[0].AppliedSurface)
	}
	if got[0].TargetPath != ".claude/skills/my-harness-a/SKILL.md" {
		t.Errorf("got[0].TargetPath = %q", got[0].TargetPath)
	}
	if got[0].Timestamp.IsZero() {
		t.Error("got[0].Timestamp is zero — WriteLineageEntry should default to time.Now().UTC()")
	}
	// Round-trip fields on the rejected entry.
	if got[1].Decision != "rejected" {
		t.Errorf("got[1].Decision = %q, want rejected", got[1].Decision)
	}
	if got[1].AppliedSurface != "" {
		t.Errorf("got[1].AppliedSurface = %q, want empty for reject entry", got[1].AppliedSurface)
	}
	if got[1].Reason == "" {
		t.Error("got[1].Reason is empty — reject entry must carry a reason")
	}
}

// TestLoadManifest_SkipsBlankLines verifies LoadManifest tolerates blank lines.
func TestLoadManifest_SkipsBlankLines(t *testing.T) {
	t.Parallel()

	manifestPath := filepath.Join(t.TempDir(), "manifest.jsonl")
	if err := WriteLineageEntry(manifestPath, LineageEntry{ProposalID: "x", Decision: "approved"}); err != nil {
		t.Fatalf("WriteLineageEntry: %v", err)
	}
	// Append a trailing blank line.
	f, err := os.OpenFile(manifestPath, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if _, err := f.WriteString("\n\n"); err != nil {
		t.Fatalf("write blank: %v", err)
	}
	_ = f.Close()

	got, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("LoadManifest len = %d, want 1 (blank lines skipped)", len(got))
	}
}

// TestWriteLineageEntry_ParentDirCreateFails verifies WriteLineageEntry returns an error
// when the parent directory cannot be created (a file occupies the parent path).
func TestWriteLineageEntry_ParentDirCreateFails(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Create a regular file where a parent directory would need to be.
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatalf("write blocker: %v", err)
	}
	// manifest path requires creating <blocker>/sub as a dir — impossible (blocker is a file).
	manifestPath := filepath.Join(blocker, "sub", "manifest.jsonl")

	err := WriteLineageEntry(manifestPath, LineageEntry{ProposalID: "x", Decision: "approved"})
	if err == nil {
		t.Error("WriteLineageEntry should return an error when parent dir cannot be created")
	}
}

// TestApply_RejectedSynthesizedReason verifies the reject lineage entry carries a synthesized
// reason (derived from RejectedBy) when the rejecting Decision has an empty Reason.
func TestApply_RejectedSynthesizedReason(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	snapshotBase := filepath.Join(dir, "snapshots")
	manifestPath := filepath.Join(dir, "learning-history", "manifest.jsonl")

	a := newApplierWithManifest(manifestPath)
	proposal := Proposal{
		ID:               "reject-noreason-001",
		TargetPath:       ".claude/agents/moai/x.md",
		FieldKey:         "description",
		NewValue:         "should not apply",
		PatternKey:       "test:test:ctx",
		Tier:             TierAutoUpdate,
		ObservationCount: 10,
	}
	// rejectedNoReasonEvaluator returns DecisionRejected with RejectedBy=2 and EMPTY Reason.
	rejector := &stubEvaluator{decision: Decision{Kind: DecisionRejected, RejectedBy: 2}}

	if err := a.Apply(proposal, rejector, snapshotBase, []Session{}); err == nil {
		t.Fatal("Apply on rejection returned nil")
	}

	entries, lerr := LoadManifest(manifestPath)
	if lerr != nil {
		t.Fatalf("LoadManifest: %v", lerr)
	}
	if len(entries) != 1 {
		t.Fatalf("manifest len = %d, want 1", len(entries))
	}
	if entries[0].Reason == "" {
		t.Error("synthesized reason is empty — rejectReason fallback should populate it from RejectedBy")
	}
	if !strings.Contains(entries[0].Reason, "L2") {
		t.Errorf("synthesized reason %q should reference layer L2", entries[0].Reason)
	}
}

// ─────────────────────────────────────────────
// Apply()-lineage integration tests
// ─────────────────────────────────────────────

// writeLineageFixture writes the shared skillFixture to a my-harness-* SKILL.md
// under a temp dir and returns the skill path.
func writeLineageFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	skillDir := filepath.Join(dir, ".claude", "skills", "my-harness-x")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("MkdirAll skillDir: %v", err)
	}
	skillPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(skillFixture), 0o644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}
	return skillPath
}

// TestApply_LineageApproved (AC-HLC-001 / REQ-HLC-001/003): an approved apply appends
// exactly one decision:"approved" lineage entry with applied_surface == "description".
func TestApply_LineageApproved(t *testing.T) {
	t.Parallel()

	skillPath := writeLineageFixture(t)
	snapshotBase := filepath.Join(t.TempDir(), "snapshots")
	manifestPath := filepath.Join(t.TempDir(), "learning-history", "manifest.jsonl")

	a := newApplierWithManifest(manifestPath)
	proposal := Proposal{
		ID:               "approved-001",
		TargetPath:       skillPath,
		FieldKey:         "description",
		NewValue:         "harness frequently triggered",
		PatternKey:       "moai_subcommand:/moai plan:ctx001",
		Tier:             TierAutoUpdate,
		ObservationCount: 10,
	}

	if err := a.Apply(proposal, approvedEvaluator(), snapshotBase, []Session{}); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	entries, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("manifest len = %d, want 1", len(entries))
	}
	e := entries[0]
	if e.Decision != "approved" {
		t.Errorf("Decision = %q, want approved", e.Decision)
	}
	if e.AppliedSurface != "description" {
		t.Errorf("AppliedSurface = %q, want description", e.AppliedSurface)
	}
	if e.ProposalID != "approved-001" {
		t.Errorf("ProposalID = %q, want approved-001", e.ProposalID)
	}
	if e.TargetPath != skillPath {
		t.Errorf("TargetPath = %q, want %q", e.TargetPath, skillPath)
	}
	if e.Timestamp.IsZero() {
		t.Error("Timestamp is zero")
	}
}

// TestApply_FrozenRejectedLineage (AC-HLC-002 / REQ-HLC-004/008): a rejected apply
// appends exactly one decision:"rejected" entry with a non-empty reason, writes no
// SKILL.md and creates no snapshot directory.
func TestApply_FrozenRejectedLineage(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	snapshotBase := filepath.Join(dir, "snapshots")
	manifestPath := filepath.Join(dir, "learning-history", "manifest.jsonl")

	a := newApplierWithManifest(manifestPath)
	proposal := Proposal{
		ID:               "frozen-001",
		TargetPath:       ".claude/agents/moai/x.md", // FROZEN zone
		FieldKey:         "description",
		NewValue:         "should not apply",
		PatternKey:       "test:test:ctx",
		Tier:             TierAutoUpdate,
		ObservationCount: 10,
	}

	err := a.Apply(proposal, rejectedEvaluator(), snapshotBase, []Session{})
	if err == nil {
		t.Fatal("Apply on rejection returned nil — rejection error must still be returned")
	}
	if !strings.Contains(err.Error(), "rejected") && !strings.Contains(err.Error(), "거부") {
		t.Errorf("rejection error missing reject marker: %v", err)
	}

	// Exactly one rejected lineage entry.
	entries, lerr := LoadManifest(manifestPath)
	if lerr != nil {
		t.Fatalf("LoadManifest: %v", lerr)
	}
	if len(entries) != 1 {
		t.Fatalf("manifest len = %d, want 1 (one rejected entry)", len(entries))
	}
	e := entries[0]
	if e.Decision != "rejected" {
		t.Errorf("Decision = %q, want rejected", e.Decision)
	}
	if e.Reason == "" {
		t.Error("rejected entry Reason is empty — must derive from rejecting layer")
	}
	if e.ProposalID != "frozen-001" {
		t.Errorf("ProposalID = %q, want frozen-001", e.ProposalID)
	}

	// No snapshot directory created for the rejected proposal (active harness unchanged).
	if _, statErr := os.Stat(snapshotBase); statErr == nil {
		snapEntries, _ := os.ReadDir(snapshotBase)
		if len(snapEntries) > 0 {
			t.Error("snapshot directory created despite rejection")
		}
	}
}

// TestApply_PendingNoLineage (AC-HLC-003 / REQ-HLC-005/006): a pending (human-gate)
// resolution returns *ApplyPendingError and writes NO lineage entry.
func TestApply_PendingNoLineage(t *testing.T) {
	t.Parallel()

	skillPath := writeLineageFixture(t)
	snapshotBase := filepath.Join(t.TempDir(), "snapshots")
	manifestPath := filepath.Join(t.TempDir(), "learning-history", "manifest.jsonl")

	a := newApplierWithManifest(manifestPath)
	proposal := Proposal{
		ID:               "pending-001",
		TargetPath:       skillPath,
		FieldKey:         "description",
		NewValue:         "pending note",
		PatternKey:       "moai_subcommand:/moai plan:ctx001",
		Tier:             TierAutoUpdate,
		ObservationCount: 10,
	}

	err := a.Apply(proposal, pendingEvaluator("pending-001"), snapshotBase, []Session{})
	if err == nil {
		t.Fatal("Apply on pending returned nil — *ApplyPendingError expected")
	}
	var pendingErr *ApplyPendingError
	if !isPendingError(err, &pendingErr) {
		t.Fatalf("error type is not *ApplyPendingError: %T", err)
	}
	if pendingErr.OversightPayload == nil {
		t.Error("OversightPayload is nil")
	}

	// Pending is not a transition — manifest must remain empty.
	entries, lerr := LoadManifest(manifestPath)
	if lerr != nil {
		t.Fatalf("LoadManifest: %v", lerr)
	}
	if len(entries) != 0 {
		t.Errorf("manifest len = %d, want 0 (pending writes no lineage)", len(entries))
	}
}

// TestApply_PreservesFrontmatterBody (AC-HLC-004 / REQ-HLC-007): an approved apply
// changes ONLY the description field; all other frontmatter lines and the body are
// preserved byte-for-byte.
func TestApply_PreservesFrontmatterBody(t *testing.T) {
	t.Parallel()

	skillPath := writeLineageFixture(t)
	snapshotBase := filepath.Join(t.TempDir(), "snapshots")
	manifestPath := filepath.Join(t.TempDir(), "learning-history", "manifest.jsonl")

	preBody := extractBody(skillFixture)

	a := newApplierWithManifest(manifestPath)
	proposal := Proposal{
		ID:               "preserve-001",
		TargetPath:       skillPath,
		FieldKey:         "description",
		NewValue:         "preservation check note",
		PatternKey:       "moai_subcommand:/moai plan:ctx001",
		Tier:             TierAutoUpdate,
		ObservationCount: 10,
	}

	if err := a.Apply(proposal, approvedEvaluator(), snapshotBase, []Session{}); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	after, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("read after: %v", err)
	}
	afterStr := string(after)

	// Body preserved byte-for-byte.
	postBody := extractBody(afterStr)
	if postBody != preBody {
		t.Errorf("body changed:\npre:  %q\npost: %q", preBody, postBody)
	}

	// Non-description frontmatter fields preserved.
	for _, mustKeep := range []string{
		"name: my-harness-test",
		`keyword: "harness test"`,
		`keyword: "test harness"`,
		`version: "1.0.0"`,
		`author: "tester"`,
	} {
		if !strings.Contains(afterStr, mustKeep) {
			t.Errorf("frontmatter line lost: %q", mustKeep)
		}
	}

	// The applied description enrichment is present.
	if !strings.Contains(afterStr, "preservation check note") {
		t.Error("description enrichment not applied")
	}
}

// TestApply_SnapshotCreatedFirstApply (AC-HLC-006 / REQ-HLC-009): the first approved
// apply creates the (previously absent) snapshot base dir with a per-apply snapshot
// dir containing a byte-for-byte backup + manifest.json.
func TestApply_SnapshotCreatedFirstApply(t *testing.T) {
	t.Parallel()

	skillPath := writeLineageFixture(t)
	preBytes, _ := os.ReadFile(skillPath)
	// snapshotBase deliberately does NOT exist yet (mirrors live tree).
	snapshotBase := filepath.Join(t.TempDir(), "snapshots-absent")
	manifestPath := filepath.Join(t.TempDir(), "learning-history", "manifest.jsonl")

	if _, statErr := os.Stat(snapshotBase); statErr == nil {
		t.Fatal("precondition failed: snapshotBase already exists")
	}

	a := newApplierWithManifest(manifestPath)
	proposal := Proposal{
		ID:               "first-apply-001",
		TargetPath:       skillPath,
		FieldKey:         "description",
		NewValue:         "first apply note",
		PatternKey:       "moai_subcommand:/moai plan:ctx001",
		Tier:             TierAutoUpdate,
		ObservationCount: 10,
	}

	if err := a.Apply(proposal, approvedEvaluator(), snapshotBase, []Session{}); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	// Base dir created.
	if _, statErr := os.Stat(snapshotBase); statErr != nil {
		t.Fatalf("snapshot base dir not created: %v", statErr)
	}
	entries, err := os.ReadDir(snapshotBase)
	if err != nil {
		t.Fatalf("ReadDir snapshotBase: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("snapshot base dir empty — no per-apply snapshot dir")
	}
	snapDir := filepath.Join(snapshotBase, entries[0].Name())

	// manifest.json present.
	if _, statErr := os.Stat(filepath.Join(snapDir, "manifest.json")); statErr != nil {
		t.Errorf("manifest.json missing in snapshot dir: %v", statErr)
	}

	// Byte-for-byte backup of the PRE-apply SKILL.md.
	backupPath := filepath.Join(snapDir, "SKILL.md")
	backupBytes, berr := os.ReadFile(backupPath)
	if berr != nil {
		t.Fatalf("read backup: %v", berr)
	}
	if !bytes.Equal(preBytes, backupBytes) {
		t.Error("snapshot backup is not byte-for-byte the pre-apply SKILL.md")
	}
}

// TestApply_Rollback (AC-HLC-007 / REQ-HLC-010): after an approved apply modifies the
// SKILL.md, RestoreSnapshot restores it byte-identically to its pre-apply content.
func TestApply_Rollback(t *testing.T) {
	t.Parallel()

	skillPath := writeLineageFixture(t)
	preBytes, _ := os.ReadFile(skillPath)
	snapshotBase := filepath.Join(t.TempDir(), "snapshots")
	manifestPath := filepath.Join(t.TempDir(), "learning-history", "manifest.jsonl")

	a := newApplierWithManifest(manifestPath)
	proposal := Proposal{
		ID:               "rollback-001",
		TargetPath:       skillPath,
		FieldKey:         "description",
		NewValue:         "rollback note",
		PatternKey:       "moai_subcommand:/moai plan:ctx001",
		Tier:             TierAutoUpdate,
		ObservationCount: 10,
	}

	if err := a.Apply(proposal, approvedEvaluator(), snapshotBase, []Session{}); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	// Sanity: the file was modified.
	afterApply, _ := os.ReadFile(skillPath)
	if bytes.Equal(preBytes, afterApply) {
		t.Skip("EnrichDescription made no change (fixture already contained value)")
	}

	entries, _ := os.ReadDir(snapshotBase)
	if len(entries) == 0 {
		t.Fatal("no snapshot dir to restore from")
	}
	snapDir := filepath.Join(snapshotBase, entries[0].Name())

	if err := RestoreSnapshot(snapDir); err != nil {
		t.Fatalf("RestoreSnapshot: %v", err)
	}

	restored, _ := os.ReadFile(skillPath)
	if !bytes.Equal(preBytes, restored) {
		t.Errorf("rollback not byte-identical:\ngot:  %q\nwant: %q", string(restored), string(preBytes))
	}
}

// TestApply_NoManifestPathSkipsLineage verifies backward compat: an Applier built via
// the existing NewApplier() (no manifest path) writes no lineage and apply still works.
func TestApply_NoManifestPathSkipsLineage(t *testing.T) {
	t.Parallel()

	skillPath := writeLineageFixture(t)
	snapshotBase := filepath.Join(t.TempDir(), "snapshots")

	a := NewApplier() // no manifest path → lineage write skipped
	proposal := Proposal{
		ID:               "nolineage-001",
		TargetPath:       skillPath,
		FieldKey:         "description",
		NewValue:         "no lineage note",
		PatternKey:       "moai_subcommand:/moai plan:ctx001",
		Tier:             TierAutoUpdate,
		ObservationCount: 10,
	}

	if err := a.Apply(proposal, approvedEvaluator(), snapshotBase, []Session{}); err != nil {
		t.Fatalf("Apply with no manifest path: %v", err)
	}
	// No panic, no error — the apply proceeded normally without a manifest target.
}
