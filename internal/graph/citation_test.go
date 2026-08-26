package graph

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// citedFixture writes a file with a citable function and returns the tree
// root plus the excerpt to cite.
func citedFixture(t *testing.T, blankPrefix int) (string, string) {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "internal"), 0o755); err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	b.WriteString("package internal\n\n")
	for i := 0; i < blankPrefix; i++ {
		b.WriteString("\n")
	}
	b.WriteString("// Handler serves requests.\nfunc Handler() int {\n\treturn 42\n}\n")
	if err := os.WriteFile(filepath.Join(root, "internal", "svc.go"), []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	excerpt := "// Handler serves requests.\nfunc Handler() int {\n\treturn 42\n}"
	return root, excerpt
}

// RegionHash pins the exact hashing rule: sha256 of the excerpt with each
// line trimmed and blank lines dropped (normalized region content).
func regionHash(t *testing.T, excerpt string) string {
	t.Helper()
	var kept []string
	for _, line := range strings.Split(excerpt, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		kept = append(kept, line)
	}
	sum := sha256.Sum256([]byte(strings.Join(kept, "\n")))
	return hex.EncodeToString(sum[:])
}

// AC-GF-012 — the canon form carries excerpt + content hash; a line number,
// when present, is convenience notation only.
func TestCitationRenderCarriesCanon(t *testing.T) {
	c, err := NewCitation("internal/svc.go", "// Handler serves requests.\nfunc Handler() int {\n\treturn 42\n}", 7)
	if err != nil {
		t.Fatal(err)
	}

	rendered := c.Render()
	if !strings.Contains(rendered, "hash=") {
		t.Errorf("rendered citation lacks the content hash: %s", rendered)
	}
	if !strings.Contains(rendered, "internal/svc.go") {
		t.Errorf("rendered citation lacks the file path: %s", rendered)
	}
	if want := regionHash(t, c.Excerpt); !strings.Contains(rendered, want) {
		t.Errorf("rendered hash mismatch: want %s in %s", want, rendered)
	}
	if !strings.Contains(rendered, "L7") {
		t.Errorf("line number must appear as convenience notation: %s", rendered)
	}

	// A line-number-ONLY anchor is not the canon: NewCitation refuses to
	// build from an empty excerpt.
	if _, err := NewCitation("f.go", "", 7); err == nil {
		t.Error("NewCitation must reject an empty excerpt — line-only anchors are banned")
	}
}

// AC-GF-014 — two-tree identical-target guarantee: one citation resolved in
// two trees whose cited file differs only by blank lines inserted ABOVE the
// region resolves to the same target (line, file) in both. A line-anchored
// resolution would drift by N.
func TestCitationTwoTreeResolution(t *testing.T) {
	excerpt := "// Handler serves requests.\nfunc Handler() int {\n\treturn 42\n}"

	rootA, _ := citedFixture(t, 0)
	rootB, _ := citedFixture(t, 5) // 5 blank lines inserted above

	citeA, err := NewCitation("internal/svc.go", excerpt, 3)
	if err != nil {
		t.Fatal(err)
	}
	resA, err := ResolveCitation(rootA, citeA)
	if err != nil {
		t.Fatalf("resolve in tree A: %v", err)
	}
	resB, err := ResolveCitation(rootB, citeA)
	if err != nil {
		t.Fatalf("resolve in tree B: %v", err)
	}

	if resA.File != "internal/svc.go" || resB.File != "internal/svc.go" {
		t.Errorf("both resolutions must name the cited file: %q / %q", resA.File, resB.File)
	}
	if !resA.Matched || !resB.Matched {
		t.Fatalf("both resolutions must match by content: A=%+v B=%+v", resA, resB)
	}
	// Tree A: the region starts at its physical line. Tree B: the SAME
	// region sits exactly 5 lines lower. A line-anchored resolver would
	// return the same NUMBER in both (3) — pointing at the wrong place in
	// tree B. The content anchor finds each tree's true position.
	if resA.Line != 3 {
		t.Errorf("tree A resolution = line %d, want 3", resA.Line)
	}
	if resB.Line != resA.Line+5 {
		t.Errorf("tree B resolution = line %d, want the region's physical line %d (found by content, not by the stale hint)", resB.Line, resA.Line+5)
	}
	// The citation's convenience line was 3 in BOTH resolutions' input; the
	// resolver must never have trusted it (tree B proved this by moving).
	_ = citeA.Line
}

// Honest staleness: when the cited REGION ITSELF was edited (not just lines
// above it), the hash no longer matches and the resolver reports the
// mismatch rather than force-resolving (spec.md §D.6 edge case).
func TestCitationRegionEditedReportsMismatch(t *testing.T) {
	root, excerpt := citedFixture(t, 0)
	cite, err := NewCitation("internal/svc.go", excerpt, 3)
	if err != nil {
		t.Fatal(err)
	}

	edited := strings.Replace(excerpt, "return 42", "return 7", 1)
	if err := os.WriteFile(filepath.Join(root, "internal", "svc.go"),
		[]byte("package internal\n\n"+edited+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := ResolveCitation(root, cite)
	if err != nil {
		t.Fatalf("resolve must not error on mismatch: %v", err)
	}
	if res.Matched {
		t.Error("edited region must NOT resolve — honest staleness beats force-resolution")
	}
	if res.Reason == "" {
		t.Error("mismatch must carry a reason string")
	}
}

// CR round-2 boundary regression (t261: failing input + observed red):
// a citation whose File escapes the root via .. is REJECTED, and the
// reason carries no host path.
func TestCitation_RejectsPathEscape_NoHostPathInReason(t *testing.T) {
	root, excerpt := citedFixture(t, 0)
	cite, err := NewCitation("../../../etc/passwd", excerpt, 1)
	if err != nil {
		t.Fatal(err)
	}
	res, err := ResolveCitation(root, cite)
	if err != nil {
		t.Fatalf("escape attempt must resolve unmatched, not error: %v", err)
	}
	if res.Matched {
		t.Error("a tree-escaping citation path must NOT resolve")
	}
	if !strings.Contains(res.Reason, "outside the tree root") {
		t.Errorf("reason must name the containment rejection: %q", res.Reason)
	}
	if strings.Contains(res.Reason, root) {
		t.Errorf("reason must not embed the host path: %q", res.Reason)
	}
	// Unreadable-file reason is likewise host-path-free.
	cite2, _ := NewCitation("internal/missing.go", excerpt, 1)
	res2, err := ResolveCitation(root, cite2)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(res2.Reason, root) {
		t.Errorf("unreadable reason must not embed the host path: %q", res2.Reason)
	}
}

// REQ-GFR-002 (CR round-2 3855001995) — the internal-consistency branch
// (citation.go:148-152): a citation whose RegionHash is populated but
// disagrees with the sha256 of its OWN excerpt is reported as such — never
// silently re-hashed, never resolved by the content search below the branch.
// The cited file here is intact, so only this branch distinguishes the
// outcome from an ordinary stale-region mismatch.
func TestCitationRegionHashMismatchReported(t *testing.T) {
	root, excerpt := citedFixture(t, 0)
	cite, err := NewCitation("internal/svc.go", excerpt, 3)
	if err != nil {
		t.Fatal(err)
	}
	// Corrupt ONLY the hash: the excerpt still matches the file verbatim, so
	// the content search WOULD match — the hash check reports first.
	cite.RegionHash = regionHash(t, "func Different() int {\n\treturn 7\n}")

	res, err := ResolveCitation(root, cite)
	if err != nil {
		t.Fatalf("inconsistent citation must resolve unmatched, not error: %v", err)
	}
	if res.Matched {
		t.Fatal("an internally-inconsistent citation must NOT resolve")
	}
	if res.Reason != "citation region hash does not cover its excerpt" {
		t.Errorf("reason = %q, want the hash-does-not-cover-its-excerpt branch reason", res.Reason)
	}

	// Control: the same citation with the CORRECT hash resolves — the branch
	// fires on hash disagreement, not on anything else in this fixture.
	ctl, err := NewCitation("internal/svc.go", excerpt, 3)
	if err != nil {
		t.Fatal(err)
	}
	resCtl, err := ResolveCitation(root, ctl)
	if err != nil || !resCtl.Matched {
		t.Fatalf("control citation must resolve matched: err=%v res=%+v", err, resCtl)
	}
}

// Missing file resolves as unmatched (not a crash, not a forced match).
func TestCitationMissingFile(t *testing.T) {
	root, _ := t.TempDir(), ""
	cite, err := NewCitation("internal/gone.go", "func Gone() {}", 1)
	if err != nil {
		t.Fatal(err)
	}
	res, err := ResolveCitation(root, cite)
	if err != nil {
		t.Fatalf("missing file must not error: %v", err)
	}
	if res.Matched {
		t.Error("missing file must resolve unmatched")
	}
}
