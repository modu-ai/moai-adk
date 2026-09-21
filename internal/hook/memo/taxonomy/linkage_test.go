// linkage_test.go — index-linkage and topic-count audits.
//
// The pre-existing audits validate a memory file's own shape (frontmatter,
// type, body structure, duplicate descriptions) and the index's line count.
// None of them reads the index's links, so a topic file that was written but
// never indexed passes every check while being unreachable: only MEMORY.md is
// loaded into a session, so an unlinked file is stored and never recalled.
// These tests pin the two checks that close that gap.
package taxonomy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeMemoryFixture lays out a memory directory: an index carrying one link
// line per name in linked, and a topic file per name in present.
func writeMemoryFixture(t *testing.T, linked, present []string) string {
	t.Helper()
	dir := t.TempDir()

	var idx strings.Builder
	idx.WriteString("# Memory Index\n\n")
	for _, name := range linked {
		idx.WriteString("- [Title](" + name + ") — hook\n")
	}
	if err := os.WriteFile(filepath.Join(dir, "MEMORY.md"), []byte(idx.String()), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}

	for _, name := range present {
		body := "---\nname: x\ndescription: d\nmetadata:\n  type: feedback\n---\n\nbody\n"
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	return dir
}

// writeIndexFile overwrites name with an index body: one link line per target.
// Archiving folds an entry out of MEMORY.md into a secondary index that sits
// beside the topic files, so a fixture needs to produce that shape directly.
func writeIndexFile(t *testing.T, dir, name string, targets []string) {
	t.Helper()
	var b strings.Builder
	b.WriteString("# " + name + "\n\n")
	for _, target := range targets {
		b.WriteString("- [Title](" + target + ") — hook\n")
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(b.String()), 0o644); err != nil {
		t.Fatalf("write index %s: %v", name, err)
	}
}

func codesOf(findings []AuditFinding) map[AuditCode]int {
	out := map[AuditCode]int{}
	for _, f := range findings {
		out[f.Code]++
	}
	return out
}

// TestAuditLinkageFindsOrphans is the case that actually bit this project:
// files on disk that no index line points at.
func TestAuditLinkageFindsOrphans(t *testing.T) {
	t.Parallel()
	dir := writeMemoryFixture(t,
		[]string{"feedback_a.md"},
		[]string{"feedback_a.md", "feedback_b.md", "project_c.md"})

	findings, err := AuditLinkage(dir)
	if err != nil {
		t.Fatalf("AuditLinkage: %v", err)
	}
	if got := codesOf(findings)[WarnOrphanNotIndexed]; got != 2 {
		t.Errorf("orphan findings = %d, want 2 (feedback_b, project_c): %+v", got, findings)
	}
	for _, f := range findings {
		if f.Code == WarnOrphanNotIndexed && strings.Contains(f.Path, "feedback_a.md") {
			t.Error("indexed file reported as an orphan")
		}
	}
}

// TestAuditLinkageFindsDanglingLinks covers the other direction: an index
// line whose target was moved or removed.
func TestAuditLinkageFindsDanglingLinks(t *testing.T) {
	t.Parallel()
	dir := writeMemoryFixture(t,
		[]string{"feedback_a.md", "feedback_gone.md"},
		[]string{"feedback_a.md"})

	findings, err := AuditLinkage(dir)
	if err != nil {
		t.Fatalf("AuditLinkage: %v", err)
	}
	if got := codesOf(findings)[WarnDanglingIndexLink]; got != 1 {
		t.Errorf("dangling findings = %d, want 1: %+v", got, findings)
	}
}

// TestAuditLinkageCleanIsSilent keeps the audit quiet when the index and the
// directory agree — a check that always fires teaches people to ignore it.
func TestAuditLinkageCleanIsSilent(t *testing.T) {
	t.Parallel()
	dir := writeMemoryFixture(t,
		[]string{"feedback_a.md", "project_b.md"},
		[]string{"feedback_a.md", "project_b.md"})

	findings, err := AuditLinkage(dir)
	if err != nil {
		t.Fatalf("AuditLinkage: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("clean fixture produced findings: %+v", findings)
	}
}

// TestAuditLinkageIgnoresArchive pins that archived files are out of scope:
// archiving is the remedy the cap prescribes, so counting archived files as
// orphans would make the remedy trip the alarm.
func TestAuditLinkageIgnoresArchive(t *testing.T) {
	t.Parallel()
	dir := writeMemoryFixture(t, []string{"feedback_a.md"}, []string{"feedback_a.md"})
	arch := filepath.Join(dir, "_archive")
	if err := os.MkdirAll(arch, 0o755); err != nil {
		t.Fatalf("mkdir archive: %v", err)
	}
	if err := os.WriteFile(filepath.Join(arch, "project_old.md"), []byte("---\n---\n"), 0o644); err != nil {
		t.Fatalf("write archived: %v", err)
	}

	findings, err := AuditLinkage(dir)
	if err != nil {
		t.Fatalf("AuditLinkage: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("archived file produced findings: %+v", findings)
	}
}

// TestAuditLinkageSecondaryIndexPreventsOrphan pins the reachability rule that
// actually governs this store: archiving folds an entry out of MEMORY.md into a
// secondary index file sitting beside the topic files, so a file reachable only
// through that secondary index is still reachable. Reading MEMORY.md alone made
// every folded entry look orphaned.
func TestAuditLinkageSecondaryIndexPreventsOrphan(t *testing.T) {
	t.Parallel()
	dir := writeMemoryFixture(t,
		[]string{"archive_index.md"},
		[]string{"archive_index.md", "feedback_folded_a.md", "feedback_folded_b.md", "feedback_folded_c.md"})
	writeIndexFile(t, dir, "archive_index.md",
		[]string{"feedback_folded_a.md", "feedback_folded_b.md", "feedback_folded_c.md"})

	findings, err := AuditLinkage(dir)
	if err != nil {
		t.Fatalf("AuditLinkage: %v", err)
	}
	if got := codesOf(findings)[WarnOrphanNotIndexed]; got != 0 {
		t.Errorf("orphan findings = %d, want 0 (all three files are reachable via archive_index.md): %+v", got, findings)
	}
}

// TestAuditLinkageOrphanSurvivesSecondaryIndex keeps the orphan check honest in
// the other direction: a secondary index makes the files it links reachable and
// nothing else.
//
// The fixture originally omitted feedback_folded_b.md, leaving the index with a
// link nothing answered — an unnoticed instance of the very defect the dangling
// check now reports. It is written out here so the test exercises what its
// comment claims rather than a half-resolved index.
func TestAuditLinkageOrphanSurvivesSecondaryIndex(t *testing.T) {
	t.Parallel()
	dir := writeMemoryFixture(t,
		[]string{"archive_index.md"},
		[]string{"archive_index.md", "feedback_folded_a.md", "feedback_folded_b.md", "feedback_folded_c.md", "feedback_lonely.md"})
	writeIndexFile(t, dir, "archive_index.md",
		[]string{"feedback_folded_a.md", "feedback_folded_b.md", "feedback_folded_c.md"})

	findings, err := AuditLinkage(dir)
	if err != nil {
		t.Fatalf("AuditLinkage: %v", err)
	}
	if got := codesOf(findings)[WarnOrphanNotIndexed]; got != 1 {
		t.Errorf("orphan findings = %d, want 1 (feedback_lonely.md): %+v", got, findings)
	}
	for _, f := range findings {
		if f.Code == WarnOrphanNotIndexed && !strings.Contains(f.Path, "feedback_lonely.md") {
			t.Errorf("unexpected orphan: %+v", f)
		}
	}
}

// TestAuditLinkageReportsDoubleIndexedFile pins the malfunction that is silent
// today: reviving a folded entry by restoring its MEMORY.md line without
// removing the archive line leaves it in two indexes. Reachability is satisfied
// twice over, so neither the orphan nor the dangling direction says anything.
func TestAuditLinkageReportsDoubleIndexedFile(t *testing.T) {
	t.Parallel()
	dir := writeMemoryFixture(t,
		[]string{"archive_index.md", "feedback_revived.md"},
		[]string{"archive_index.md", "feedback_revived.md", "feedback_folded.md", "feedback_folded_b.md"})
	writeIndexFile(t, dir, "archive_index.md",
		[]string{"feedback_revived.md", "feedback_folded.md", "feedback_folded_b.md"})

	findings, err := AuditLinkage(dir)
	if err != nil {
		t.Fatalf("AuditLinkage: %v", err)
	}
	if got := codesOf(findings)[WarnIndexDuplicateEntry]; got != 1 {
		t.Fatalf("duplicate-entry findings = %d, want 1: %+v", got, findings)
	}
	for _, f := range findings {
		if f.Code != WarnIndexDuplicateEntry {
			continue
		}
		if !strings.Contains(f.Detail, "feedback_revived.md") {
			t.Errorf("detail should name the double-indexed file: %q", f.Detail)
		}
		if !strings.Contains(f.Detail, "archive_index.md") {
			t.Errorf("detail should name the secondary index: %q", f.Detail)
		}
	}
}

// TestAuditLinkageSingleIndexIsNotDuplicate keeps the new check quiet for the
// ordinary case: one index line, one memory.
func TestAuditLinkageSingleIndexIsNotDuplicate(t *testing.T) {
	t.Parallel()
	dir := writeMemoryFixture(t,
		[]string{"archive_index.md", "feedback_live.md"},
		[]string{"archive_index.md", "feedback_live.md", "feedback_folded_a.md", "feedback_folded_b.md", "feedback_folded_c.md"})
	writeIndexFile(t, dir, "archive_index.md",
		[]string{"feedback_folded_a.md", "feedback_folded_b.md", "feedback_folded_c.md"})

	findings, err := AuditLinkage(dir)
	if err != nil {
		t.Fatalf("AuditLinkage: %v", err)
	}
	if got := codesOf(findings)[WarnIndexDuplicateEntry]; got != 0 {
		t.Errorf("duplicate-entry findings = %d, want 0: %+v", got, findings)
	}
}

// TestAuditLinkageIndexDiscriminatorBoundary pins the measured separation
// between a real index and a topic file that happens to cite a sibling: an
// index carries several link lines, an ordinary memory cites at most one.
func TestAuditLinkageIndexDiscriminatorBoundary(t *testing.T) {
	t.Parallel()

	// Two links: still a citation, not an index — this is the case measured on
	// the live store, where a card record citing two siblings on one line was
	// read as an index and hid both of them from the orphan finding.
	below := writeMemoryFixture(t,
		[]string{"feedback_citing.md"},
		[]string{"feedback_citing.md", "feedback_cited.md", "feedback_other.md"})
	writeIndexFile(t, below, "feedback_citing.md",
		[]string{"feedback_cited.md", "feedback_other.md"})

	findings, err := AuditLinkage(below)
	if err != nil {
		t.Fatalf("AuditLinkage(below): %v", err)
	}
	if got := codesOf(findings)[WarnOrphanNotIndexed]; got != 2 {
		t.Errorf("two-link file treated as an index: orphan findings = %d, want 2: %+v", got, findings)
	}

	// Three links: an index — every target becomes reachable.
	at := writeMemoryFixture(t,
		[]string{"feedback_citing.md"},
		[]string{"feedback_citing.md", "feedback_cited.md", "feedback_other.md", "feedback_third.md"})
	writeIndexFile(t, at, "feedback_citing.md",
		[]string{"feedback_cited.md", "feedback_other.md", "feedback_third.md"})

	findings, err = AuditLinkage(at)
	if err != nil {
		t.Fatalf("AuditLinkage(at): %v", err)
	}
	if got := codesOf(findings)[WarnOrphanNotIndexed]; got != 0 {
		t.Errorf("three-link file not treated as an index: orphan findings = %d, want 0: %+v", got, findings)
	}
}

// TestAuditTopicCountOverCap pins the constitution's per-project ceiling,
// which had no checker at all.
func TestAuditTopicCountOverCap(t *testing.T) {
	t.Parallel()
	names := []string{"feedback_a.md", "feedback_b.md", "feedback_c.md", "feedback_d.md"}
	dir := writeMemoryFixture(t, names, names)

	findings, err := AuditTopicCount(dir, 3)
	if err != nil {
		t.Fatalf("AuditTopicCount: %v", err)
	}
	if got := codesOf(findings)[WarnTopicCountOverCap]; got != 1 {
		t.Errorf("over-cap findings = %d, want 1: %+v", got, findings)
	}
	if len(findings) == 1 && !strings.Contains(findings[0].Detail, "4") {
		t.Errorf("detail should name the observed count: %q", findings[0].Detail)
	}

	under, err := AuditTopicCount(dir, 10)
	if err != nil {
		t.Fatalf("AuditTopicCount under cap: %v", err)
	}
	if len(under) != 0 {
		t.Errorf("under cap produced findings: %+v", under)
	}
}

// TestAuditMissingDirIsNotAnError keeps both audits fail-open: a project with
// no memory directory yet is not a defect.
func TestAuditMissingDirIsNotAnError(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "nope")

	if f, err := AuditLinkage(missing); err != nil || len(f) != 0 {
		t.Errorf("AuditLinkage(missing) = %+v, %v; want nil, nil", f, err)
	}
	if f, err := AuditTopicCount(missing, 50); err != nil || len(f) != 0 {
		t.Errorf("AuditTopicCount(missing) = %+v, %v; want nil, nil", f, err)
	}
}

// writeCitingFile overwrites name with an ordinary topic file whose body cites
// targets. Same bytes as an index — which is the point: what separates the two
// is what the links resolve to, not how they are written.
func writeCitingFile(t *testing.T, dir, name string, targets []string) {
	t.Helper()
	writeIndexFile(t, dir, name, targets)
}

// TestAuditLinkagePromotionIgnoresMissingTargets pins F1: links pointing at
// nothing must not buy index status. A file citing three absent siblings met
// the link threshold, became a secondary index, and every real memory it also
// linked went silent — reachable, on the audit's account, through an index
// that reaches nothing.
func TestAuditLinkagePromotionIgnoresMissingTargets(t *testing.T) {
	t.Parallel()
	dir := writeMemoryFixture(t,
		[]string{"feedback_citing.md"},
		[]string{"feedback_citing.md", "feedback_real_orphan.md"})
	writeCitingFile(t, dir, "feedback_citing.md",
		[]string{"ghost_a.md", "ghost_b.md", "ghost_c.md", "feedback_real_orphan.md"})

	findings, err := AuditLinkage(dir)
	if err != nil {
		t.Fatalf("AuditLinkage: %v", err)
	}
	if got := codesOf(findings)[WarnOrphanNotIndexed]; got != 1 {
		t.Errorf("orphan findings = %d, want 1 (feedback_real_orphan.md is hidden behind a file promoted by three absent targets): %+v", got, findings)
	}
}

// TestAuditLinkageSecondaryIndexDanglingIsReported pins F2, and with it the
// self-concealment the two defects compose into. The link set is byte-identical
// in both halves; only its carrier moves. In MEMORY.md the three absent targets
// are reported; inside a promoted secondary index the same three are silent —
// so the links that bought the promotion are exactly the links nothing checks.
func TestAuditLinkageSecondaryIndexDanglingIsReported(t *testing.T) {
	t.Parallel()
	ghosts := []string{"ghost_a.md", "ghost_b.md", "ghost_c.md"}

	// Control: the same three links carried by the index a session loads.
	inMemory := writeMemoryFixture(t,
		append([]string{"feedback_carrier.md"}, ghosts...),
		[]string{"feedback_carrier.md", "feedback_r1.md", "feedback_r2.md", "feedback_r3.md"})
	writeIndexFile(t, inMemory, "feedback_carrier.md",
		[]string{"feedback_r1.md", "feedback_r2.md", "feedback_r3.md"})

	control, err := AuditLinkage(inMemory)
	if err != nil {
		t.Fatalf("AuditLinkage(control): %v", err)
	}
	if got := codesOf(control)[WarnDanglingIndexLink]; got != 3 {
		t.Fatalf("control: dangling findings = %d, want 3 — the fixture does not reproduce the reported direction: %+v", got, control)
	}

	// Subject: the same three links moved into the secondary index.
	inSecondary := writeMemoryFixture(t,
		[]string{"feedback_carrier.md"},
		[]string{"feedback_carrier.md", "feedback_r1.md", "feedback_r2.md", "feedback_r3.md"})
	writeIndexFile(t, inSecondary, "feedback_carrier.md",
		append([]string{"feedback_r1.md", "feedback_r2.md", "feedback_r3.md"}, ghosts...))

	subject, err := AuditLinkage(inSecondary)
	if err != nil {
		t.Fatalf("AuditLinkage(subject): %v", err)
	}
	if got := codesOf(subject)[WarnDanglingIndexLink]; got != 3 {
		t.Errorf("subject: dangling findings = %d, want 3 — a secondary index's broken links are reported nowhere: %+v", got, subject)
	}
	for _, f := range subject {
		if f.Code == WarnDanglingIndexLink && !strings.Contains(f.Path, "feedback_carrier.md") {
			t.Errorf("dangling finding should name the carrying index, not %s: %+v", f.Path, f)
		}
	}
}

// TestAuditLinkageResolvedIndexStillPromotes is the other direction of the same
// change: a file whose links all resolve keeps its index status at the existing
// threshold. Without this, "count only resolved links" is indistinguishable
// from demoting every index — which would pass the two tests above while
// reporting every folded memory as an orphan.
func TestAuditLinkageResolvedIndexStillPromotes(t *testing.T) {
	t.Parallel()
	dir := writeMemoryFixture(t,
		[]string{"archive_index.md"},
		[]string{"archive_index.md", "feedback_a.md", "feedback_b.md", "feedback_c.md"})
	writeIndexFile(t, dir, "archive_index.md",
		[]string{"feedback_a.md", "feedback_b.md", "feedback_c.md"})

	findings, err := AuditLinkage(dir)
	if err != nil {
		t.Fatalf("AuditLinkage: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("a fully-resolved three-link index produced findings: %+v", findings)
	}
}
