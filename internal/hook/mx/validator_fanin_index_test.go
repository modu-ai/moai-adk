package mx

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// fanInFillerLines pads fixture files so byte-scanning cost is not negligible in
// the cost-shape guard below.
const fanInFillerLines = 40

// writeFanInFixture builds a small project tree used by the fan-in index tests.
// Contents are chosen to pin word-boundary semantics: a name must not be counted
// when it appears as a substring of a longer identifier, and must be counted when
// it is delimited by any non-identifier byte (dot, quote, paren, comment marker).
func writeFanInFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	files := map[string]string{
		"decl.go": `package p

// Target is the declaration site.
func Target() {}
`,
		"callers.go": `package p

func a() { Target() }
func b() { pkg.Target() }
func c() { _ = "Target" }
// d calls Target in a comment
`,
		"substrings.go": `package p

func e() {
	TargetX()
	XTarget()
	Target_()
	_ = my.TargetSuffix
}
`,
		"nested/deep.go": `package q

func f() { Target() }
`,
		"helpers.go": `package p

func g() { Target() }
`,
		"notgo.txt": `Target Target Target`,
	}

	for name, body := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", p, err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", p, err)
		}
	}

	// Excluded by defaultExcludeGlobs — must not contribute to counts.
	gen := filepath.Join(dir, "thing_generated.go")
	if err := os.WriteFile(gen, []byte("package p\n\nfunc h() { Target() }\n"), 0o644); err != nil {
		t.Fatalf("write generated: %v", err)
	}
	// Dot-directories are skipped by the walker — must not contribute.
	hidden := filepath.Join(dir, ".hidden")
	if err := os.MkdirAll(hidden, 0o755); err != nil {
		t.Fatalf("mkdir hidden: %v", err)
	}
	if err := os.WriteFile(filepath.Join(hidden, "x.go"), []byte("package p\n\nfunc i() { Target() }\n"), 0o644); err != nil {
		t.Fatalf("write hidden: %v", err)
	}

	return dir
}

// TestBuildFanInIndex_WordBoundarySemantics pins the exact per-file counts the
// index must produce. The expectations are hand-computed from the fixture, not
// read back from the implementation, so they hold across a counting-strategy change.
func TestBuildFanInIndex_WordBoundarySemantics(t *testing.T) {
	dir := writeFanInFixture(t)
	v := &mxValidator{projectRoot: dir, fanInThreshold: 3}

	idx := v.buildFanInIndex(context.Background(), []string{"Target"})
	got := idx.counts["Target"]

	want := map[string]int{
		"decl.go":        2, // doc comment + declaration
		"callers.go":     4, // call, qualified call, string literal, comment
		"substrings.go":  0, // TargetX / XTarget / Target_ / TargetSuffix are distinct identifiers
		"nested/deep.go": 1,
		"helpers.go":     1,
	}
	for rel, exp := range want {
		if c := got[filepath.Join(dir, rel)]; c != exp {
			t.Errorf("count for %s = %d, want %d", rel, c, exp)
		}
	}

	// Excluded and non-Go sources must be absent entirely.
	for _, rel := range []string{"thing_generated.go", ".hidden/x.go", "notgo.txt"} {
		if c, ok := got[filepath.Join(dir, rel)]; ok {
			t.Errorf("%s must not be counted, got %d", rel, c)
		}
	}
}

// TestFanInIndex_SubtractsDeclaration pins the fanIn() accounting: the total is
// project-wide minus one occurrence for the declaration in the current file.
func TestFanInIndex_SubtractsDeclaration(t *testing.T) {
	dir := writeFanInFixture(t)
	v := &mxValidator{projectRoot: dir, fanInThreshold: 3}
	idx := v.buildFanInIndex(context.Background(), []string{"Target"})

	// 2 + 4 + 0 + 1 + 1 = 8 project-wide, minus 1 for the declaration in decl.go.
	if got := idx.fanIn("Target", filepath.Join(dir, "decl.go")); got != 7 {
		t.Errorf("fanIn from declaring file = %d, want 7", got)
	}
	// A file that holds no occurrence subtracts nothing.
	if got := idx.fanIn("Target", filepath.Join(dir, "nowhere.go")); got != 8 {
		t.Errorf("fanIn from unrelated file = %d, want 8", got)
	}
}

// TestBuildFanInIndex_MultipleCandidates verifies that counting many names in one
// traversal yields exactly the same per-name results as counting each name alone.
// This is the correctness half of the single-traversal property.
func TestBuildFanInIndex_MultipleCandidates(t *testing.T) {
	dir := writeFanInFixture(t)
	v := &mxValidator{projectRoot: dir, fanInThreshold: 3}

	names := []string{"Target", "TargetX", "XTarget", "Target_", "TargetSuffix", "Absent"}
	together := v.buildFanInIndex(context.Background(), names)

	for _, n := range names {
		alone := v.buildFanInIndex(context.Background(), []string{n})
		if len(alone.counts[n]) != len(together.counts[n]) {
			t.Errorf("%s: %d files alone vs %d together", n, len(alone.counts[n]), len(together.counts[n]))
			continue
		}
		for p, c := range alone.counts[n] {
			if together.counts[n][p] != c {
				t.Errorf("%s in %s: %d alone vs %d together", n, filepath.Base(p), c, together.counts[n][p])
			}
		}
	}

	if len(together.counts["Absent"]) != 0 {
		t.Errorf("Absent must match no files, got %d", len(together.counts["Absent"]))
	}
}

// TestBuildFanInIndex_ASCIIBoundarySemantics pins the treatment of identifiers
// outside the ASCII identifier alphabet. Go's regexp \b is an ASCII-only word
// boundary, so a wholly non-ASCII name has never been counted here. The behavior
// is preserved deliberately: changing it would alter reported MX violations, which
// is out of scope for a performance change. Names that merely sit ADJACENT to
// non-ASCII bytes are counted, because those bytes act as delimiters.
func TestBuildFanInIndex_ASCIIBoundarySemantics(t *testing.T) {
	dir := t.TempDir()
	src := "package p\n\n// Ünïcode is exported.\nfunc Ünïcode() {}\n\nfunc z() { Ünïcode() }\n\nfunc w() { 한Target(); Target한() }\n"
	if err := os.WriteFile(filepath.Join(dir, "u.go"), []byte(src), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	v := &mxValidator{projectRoot: dir, fanInThreshold: 3}

	idx := v.buildFanInIndex(context.Background(), []string{"Ünïcode", "Target"})
	p := filepath.Join(dir, "u.go")

	if got := idx.counts["Ünïcode"][p]; got != 0 {
		t.Errorf("wholly non-ASCII name count = %d, want 0 (ASCII \\b semantics)", got)
	}
	// 한Target and Target한 both delimit "Target" with non-ASCII bytes.
	if got := idx.counts["Target"][p]; got != 2 {
		t.Errorf("ASCII name beside non-ASCII bytes = %d, want 2", got)
	}
}

// TestBuildFanInIndex_EmptyInputs covers the two early-return guards.
func TestBuildFanInIndex_EmptyInputs(t *testing.T) {
	t.Run("no names", func(t *testing.T) {
		v := &mxValidator{projectRoot: t.TempDir(), fanInThreshold: 3}
		if idx := v.buildFanInIndex(context.Background(), nil); len(idx.counts) != 0 {
			t.Errorf("want empty index, got %d entries", len(idx.counts))
		}
	})
	t.Run("no project root", func(t *testing.T) {
		v := &mxValidator{projectRoot: "", fanInThreshold: 3}
		if idx := v.buildFanInIndex(context.Background(), []string{"X"}); len(idx.counts) != 0 {
			t.Errorf("want empty index, got %d entries", len(idx.counts))
		}
	})
}

// TestBuildFanInIndex_ContextCancelled verifies the traversal stops early and
// still returns a usable (partial) index rather than panicking.
func TestBuildFanInIndex_ContextCancelled(t *testing.T) {
	dir := writeFanInFixture(t)
	v := &mxValidator{projectRoot: dir, fanInThreshold: 3}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	idx := v.buildFanInIndex(ctx, []string{"Target"})
	if idx == nil {
		t.Fatal("index must not be nil on cancellation")
	}
	if len(idx.counts["Target"]) != 0 {
		t.Errorf("cancelled traversal must count nothing, got %d files", len(idx.counts["Target"]))
	}
}

// TestBuildFanInIndex_CostDoesNotScaleWithCandidates is the regression guard for
// the defect this change fixes: the traversal used to run one full byte scan PER
// candidate name, so cost grew linearly with the candidate count (measured on this
// repository: ~370ms per additional name over 19MB of sources). A single-pass
// counter is candidate-count independent.
//
// The assertion is a ratio between two measurements taken back-to-back in the same
// process, so machine load — which affects both halves equally — cancels out. The
// tolerated ratio is deliberately far looser than the defect it catches: the old
// strategy cost ~64x here, the new one ~1x, and the gate sits at 8x.
func TestBuildFanInIndex_CostDoesNotScaleWithCandidates(t *testing.T) {
	if testing.Short() {
		t.Skip("cost-shape guard is skipped in -short mode")
	}

	dir := t.TempDir()
	padding := ""
	for range fanInFillerLines {
		padding += "// padding line so the scanner has some bytes to chew on\n"
	}
	for i := range 150 {
		body := fmt.Sprintf("package p\n\nfunc f%d() { Target() }\n%s", i, padding)
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("f%d.go", i)), []byte(body), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	v := &mxValidator{projectRoot: dir, fanInThreshold: 3}

	many := []string{"Target"}
	for i := 1; i < 64; i++ {
		many = append(many, fmt.Sprintf("Name%d", i))
	}

	// Warm the page cache so the first measurement is not paying for cold I/O.
	v.buildFanInIndex(context.Background(), []string{"Target"})

	elapsed := func(names []string) time.Duration {
		start := time.Now()
		v.buildFanInIndex(context.Background(), names)
		return time.Since(start)
	}

	one := elapsed([]string{"Target"})
	sixtyFour := elapsed(many)

	// Guard against a degenerate baseline: if the single-candidate run is too fast
	// to time reliably, the ratio is meaningless.
	if one < 100*time.Microsecond {
		t.Skipf("baseline too fast to compare reliably (%v)", one)
	}
	if ratio := float64(sixtyFour) / float64(one); ratio > 8 {
		t.Errorf("64 candidates cost %.1fx one candidate (%v vs %v) — traversal is scanning per candidate, not once",
			ratio, sixtyFour, one)
	}
}
