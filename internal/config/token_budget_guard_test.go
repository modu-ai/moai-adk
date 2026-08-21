package config

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mustGetwd returns the current working directory or fails the test.
func mustGetwd(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	return wd
}

// countNoPathsRuleFiles independently counts the .claude/rules/moai/**/*.md files
// whose frontmatter carries NO `paths:` restriction, for the relative enumeration
// assertion in AC-TEF-002 (no hardcoded literal count).
func countNoPathsRuleFiles(t *testing.T, root string) int {
	t.Helper()
	rulesDir := filepath.Join(root, ".claude", "rules", "moai")
	n := 0
	err := filepath.WalkDir(rulesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		if !hasPathsRestriction(path) {
			n++
		}
		return nil
	})
	if err != nil {
		t.Fatalf("countNoPathsRuleFiles walk: %v", err)
	}
	return n
}

// TestAlwaysLoadedTokenBudget is the P0-1 regression tripwire (AC-TEF-001 fail-path
// mirror + AC-TEF-002 pass-on-current-tree). It measures the real repo's
// always-loaded context surface and FAILS when the estimated token sum exceeds
// AlwaysLoadedTokenBudget. Skips gracefully when the repo root cannot be located
// (test running outside the tree — AC edge case).
//
// # REQ-TEF-001, REQ-TEF-002
func TestAlwaysLoadedTokenBudget(t *testing.T) {
	root, ok := findRepoRoot(mustGetwd(t))
	if !ok {
		t.Skip("repo root (go.mod) not found; skipping always-loaded budget guard")
	}
	if _, err := os.Stat(filepath.Join(root, "CLAUDE.md")); err != nil {
		t.Skip("CLAUDE.md not found at repo root; skipping (not the real repo tree)")
	}

	total, surface, err := measureAlwaysLoaded(root)
	if err != nil {
		t.Fatalf("measureAlwaysLoaded: %v", err)
	}
	// Emit the achieved figure on the PASSING path too: every ratchet criterion in
	// SPEC-AGENTS-MD-CANON-001 quotes this line from `go test -v`, and a guard that
	// prints only on failure makes those criteria unsatisfiable.
	t.Logf("always-loaded surface = %d tokens (budget %d, headroom %d, %d entries)",
		total, AlwaysLoadedTokenBudget, AlwaysLoadedTokenBudget-total, len(surface))

	if total > AlwaysLoadedTokenBudget {
		t.Errorf("always-loaded surface = %d tokens, exceeds budget %d (overflow %d); surface has %d entries — trim always-loaded rules or raise AlwaysLoadedTokenBudget with justification",
			total, AlwaysLoadedTokenBudget, total-AlwaysLoadedTokenBudget, len(surface))
	}
}

// TestCodexContractByteCeiling is the Codex-cap byte guard on the real tree: the
// root AGENTS.md and its template mirror must each stay at or below
// CodexContractByteCeiling. It FAILS the build on breach rather than warning —
// codex truncates the overflow silently, so nothing else will ever signal it.
// The failure names the measured byte figure and the offending file.
func TestCodexContractByteCeiling(t *testing.T) {
	root, ok := findRepoRoot(mustGetwd(t))
	if !ok {
		t.Skip("repo root (go.mod) not found; skipping Codex contract byte guard")
	}
	if _, err := os.Stat(filepath.Join(root, "AGENTS.md")); err != nil {
		t.Skip("AGENTS.md not found at repo root; skipping (not the real repo tree)")
	}

	breaches, err := MeasureContractBytes(root)
	if err != nil {
		t.Fatalf("MeasureContractBytes: %v", err)
	}
	// Emit the measured figures on the passing path too, so a ratchet criterion
	// can quote them from `go test -v` without re-deriving them.
	docs, err := contractDocuments(root)
	if err != nil {
		t.Fatalf("contractDocuments: %v", err)
	}
	for _, p := range docs {
		if fi, statErr := os.Stat(p); statErr == nil {
			rel, relErr := filepath.Rel(root, p)
			if relErr != nil {
				rel = p
			}
			t.Logf("contract document %s = %d bytes (ceiling %d, headroom %d)",
				rel, fi.Size(), CodexContractByteCeiling, CodexContractByteCeiling-int(fi.Size()))
		}
	}
	for _, b := range breaches {
		t.Errorf("%s = %d bytes, exceeds the Codex contract ceiling %d (overflow %d) — codex truncates the tail SILENTLY, so trim the contract layer; raising the ceiling is not a substitute",
			b.Path, b.Bytes, CodexContractByteCeiling, b.Overflow)
	}
}

// TestAlwaysLoadedTokenBudget_OverBudgetFails verifies the guard's over-budget
// detection: a synthetic surface exceeding the budget is reported as over-budget
// (would fire t.Errorf), while a small surface stays under (AC-TEF-001).
//
// # REQ-TEF-001
func TestAlwaysLoadedTokenBudget_OverBudgetFails(t *testing.T) {
	tests := []struct {
		name     string
		fileSize int
		want     string // "fail" = over budget (guard fires), "pass" = under budget
	}{
		{name: "over-budget", fileSize: AlwaysLoadedTokenBudget*4 + 4096, want: "fail"},
		{name: "under-budget", fileSize: 1024, want: "pass"},
	}
	// Codex-cap dimension, same table shape and the same repo-root helpers — a
	// parallel harness would be the second measurement path REQ-AMC-008 forbids.
	contractTests := []struct {
		name       string
		agentsSize int
		mirrorSize int
		wantPaths  int // expected breach count
	}{
		{name: "contract-at-ceiling", agentsSize: CodexContractByteCeiling, mirrorSize: CodexContractByteCeiling, wantPaths: 0},
		{name: "contract-one-byte-over", agentsSize: CodexContractByteCeiling + 1, mirrorSize: 1024, wantPaths: 1},
		{name: "mirror-over-live-under", agentsSize: 1024, mirrorSize: CodexContractByteCeiling + 1, wantPaths: 1},
		{name: "both-over", agentsSize: CodexContractByteCeiling + 1, mirrorSize: CodexContractByteCeiling + 1, wantPaths: 2},
	}
	for _, tt := range contractTests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			mirrorDir := filepath.Join(root, "internal", "template", "templates")
			if err := os.MkdirAll(mirrorDir, 0o755); err != nil {
				t.Fatal(err)
			}
			// The enumeration helper walks the rules tree, so the fixture needs it
			// to exist — contractDocuments reuses that helper by design.
			if err := os.MkdirAll(filepath.Join(root, ".claude", "rules", "moai", "core"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte(strings.Repeat("a", tt.agentsSize)), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(mirrorDir, "AGENTS.md"), []byte(strings.Repeat("m", tt.mirrorSize)), 0o644); err != nil {
				t.Fatal(err)
			}
			breaches, err := MeasureContractBytes(root)
			if err != nil {
				t.Fatalf("MeasureContractBytes: %v", err)
			}
			if len(breaches) != tt.wantPaths {
				t.Errorf("breaches = %d, want %d (%+v)", len(breaches), tt.wantPaths, breaches)
			}
			for _, b := range breaches {
				if b.Path == "" || b.Bytes <= CodexContractByteCeiling || b.Overflow != b.Bytes-CodexContractByteCeiling {
					t.Errorf("breach must name the file and the measured figure: %+v", b)
				}
			}
		})
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			rulesDir := filepath.Join(root, ".claude", "rules", "moai", "core")
			if err := os.MkdirAll(rulesDir, 0o755); err != nil {
				t.Fatal(err)
			}
			body := strings.Repeat("x", tt.fileSize)
			if err := os.WriteFile(filepath.Join(rulesDir, "big.md"), []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
			total, _, err := measureAlwaysLoaded(root)
			if err != nil {
				t.Fatalf("measureAlwaysLoaded: %v", err)
			}
			over := total > AlwaysLoadedTokenBudget
			if tt.want == "fail" && !over {
				t.Errorf("total=%d NOT over budget %d — guard would not fire on over-budget surface", total, AlwaysLoadedTokenBudget)
			}
			if tt.want == "pass" && over {
				t.Errorf("total=%d over budget %d — guard would fire on under-budget surface", total, AlwaysLoadedTokenBudget)
			}
		})
	}
}

// TestAlwaysLoadedSurfaceEnumeration asserts the enumerated surface equals
// (count of no-`paths:` rule files) + 3 fixed slots, and that a known
// paths:-scoped rule is excluded — the load-bearing enumeration-correctness proof
// (AC-TEF-002 relative count + AC-TEF-004 exclusion). No hardcoded file count.
//
// # REQ-TEF-002, REQ-TEF-004
func TestAlwaysLoadedSurfaceEnumeration(t *testing.T) {
	root, ok := findRepoRoot(mustGetwd(t))
	if !ok {
		t.Skip("repo root not found")
	}
	surface, err := alwaysLoadedSurface(root)
	if err != nil {
		t.Fatalf("alwaysLoadedSurface: %v", err)
	}

	wantRuleCount := countNoPathsRuleFiles(t, root)
	wantTotal := wantRuleCount + 4 // + CLAUDE.md + AGENTS.md + moai.md + MEMORY.md fixed slots
	if len(surface) != wantTotal {
		t.Errorf("surface has %d entries, want %d (= %d no-paths: rules + 4 fixed surfaces)", len(surface), wantTotal, wantRuleCount)
	}

	// AC-TEF-004: a known paths:-scoped rule (languages/go.md carries a paths:
	// restriction) MUST NOT appear in the always-loaded surface.
	scoped := filepath.Join(root, ".claude", "rules", "moai", "languages", "go.md")
	for _, p := range surface {
		if p == scoped {
			t.Errorf("paths:-scoped file %s must be excluded from the always-loaded surface", scoped)
		}
	}

	// AC-AMC-017: the enumeration MUST contain the root AGENTS.md. It is an
	// @-import of CLAUDE.md and therefore always-loaded; an enumeration that omits
	// it scores the relocation of clauses INTO AGENTS.md as a reduction while the
	// always-loaded context is unchanged (REQ-AMC-013).
	agents := filepath.Join(root, "AGENTS.md")
	found := false
	for _, p := range surface {
		if p == agents {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("always-loaded surface omits %s; the Codex contract layer must be enumerated (REQ-AMC-008)", agents)
	}
}

// TestHasPathsRestriction covers the frontmatter predicate incl. the AC-TEF-004
// edge: a file with malformed/unterminated frontmatter is treated conservatively
// as always-loaded (NOT restricted → counted).
func TestHasPathsRestriction(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{name: "paths present", content: "---\npaths: \"**/*.go\"\ntitle: x\n---\nbody", want: true},
		{name: "no paths key", content: "---\ntitle: x\n---\nbody", want: false},
		{name: "no frontmatter", content: "# just a heading\nbody", want: false},
		{name: "malformed unterminated frontmatter → counts (not restricted)", content: "---\npaths not a key line\nno closing delimiter", want: false},
		{name: "paths after frontmatter body ignored", content: "---\ntitle: x\n---\npaths: fake\n", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			p := filepath.Join(dir, "rule.md")
			if err := os.WriteFile(p, []byte(tt.content), 0o644); err != nil {
				t.Fatal(err)
			}
			if got := hasPathsRestriction(p); got != tt.want {
				t.Errorf("hasPathsRestriction(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}

	// Unreadable / non-existent path → treated as NOT restricted (counted).
	if got := hasPathsRestriction(filepath.Join(t.TempDir(), "does-not-exist.md")); got {
		t.Errorf("hasPathsRestriction(nonexistent) = true, want false (conservative count)")
	}
}

// TestMeasureAlwaysLoaded_WithMemory verifies measureAlwaysLoaded sums the
// no-paths: rule + fixed surfaces, applies the MEMORY.md head cap, and treats a
// paths:-scoped rule as excluded — on a synthetic hermetic repo.
func TestMeasureAlwaysLoaded_WithMemory(t *testing.T) {
	root := t.TempDir()
	// Fixed surfaces.
	writeFile(t, filepath.Join(root, "CLAUDE.md"), strings.Repeat("a", 400))                                   // 100 tokens
	writeFile(t, filepath.Join(root, ".claude", "output-styles", "moai", "moai.md"), strings.Repeat("b", 800)) // 200 tokens
	// MEMORY.md with a body far exceeding the 25KB byte cap → head-capped.
	writeFile(t, filepath.Join(root, "MEMORY.md"), strings.Repeat("m", 40*1024))
	// One no-paths: rule (counted) + one paths:-scoped rule (excluded).
	writeFile(t, filepath.Join(root, ".claude", "rules", "moai", "core", "keep.md"), "---\ntitle: x\n---\n"+strings.Repeat("c", 400)) // ~100+ tokens
	writeFile(t, filepath.Join(root, ".claude", "rules", "moai", "languages", "scoped.md"), "---\npaths: \"**/*.go\"\n---\n"+strings.Repeat("d", 4000))

	total, surface, err := measureAlwaysLoaded(root)
	if err != nil {
		t.Fatalf("measureAlwaysLoaded: %v", err)
	}
	// Enumeration: 1 no-paths: rule + 4 fixed slots = 5. AGENTS.md is absent from this
	// temp tree and contributes 0 tokens (hermetic), but is still enumerated.
	if len(surface) != 5 {
		t.Errorf("surface len = %d, want 5 (1 no-paths: rule + 4 fixed)", len(surface))
	}
	// MEMORY.md contributes head-capped tokens (25KB/4 = 6400), NOT the full 40KB.
	memHeadTokens := memoryHeadByteCap / 4
	// Lower bound: CLAUDE(100) + moai(200) + memHead(6400) = 6700, plus the rule.
	if total < 100+200+memHeadTokens {
		t.Errorf("total = %d, want ≥ %d (MEMORY.md head cap applied)", total, 100+200+memHeadTokens)
	}
	// Upper bound guard: total must NOT include the full 40KB MEMORY.md (10240 tokens).
	if total >= 100+200+(40*1024/4) {
		t.Errorf("total = %d includes uncapped MEMORY.md; head cap not applied", total)
	}
}

// writeFile writes content to path, creating parent dirs.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestEstimateTokens verifies the named char/4 estimation function (AC-TEF-003).
func TestEstimateTokens(t *testing.T) {
	if got := estimateTokens(make([]byte, 400)); got != 100 {
		t.Errorf("estimateTokens(400 bytes) = %d, want 100 (char/4)", got)
	}
	if got := estimateTokens(nil); got != 0 {
		t.Errorf("estimateTokens(nil) = %d, want 0", got)
	}
	// AC-TEF-003: the budget is a named constant, not a bare magic number.
	if AlwaysLoadedTokenBudget <= 0 {
		t.Errorf("AlwaysLoadedTokenBudget = %d, want a positive named constant", AlwaysLoadedTokenBudget)
	}
}

// TestMemoryHead verifies the MEMORY.md head cap (200 lines OR 25KB, whichever
// first) matching the Claude Code auto-memory loader.
func TestMemoryHead(t *testing.T) {
	// 300 short lines → capped at 200 lines.
	var b strings.Builder
	for i := 0; i < 300; i++ {
		b.WriteString("line\n")
	}
	head := memoryHead([]byte(b.String()))
	if got := strings.Count(string(head), "\n"); got != memoryHeadLineCap {
		t.Errorf("memoryHead line count = %d, want %d (line cap)", got, memoryHeadLineCap)
	}

	// A single 30KB line (no newline before cap) → capped at 25KB bytes.
	big := strings.Repeat("y", 30*1024)
	head = memoryHead([]byte(big))
	if len(head) != memoryHeadByteCap {
		t.Errorf("memoryHead byte length = %d, want %d (byte cap)", len(head), memoryHeadByteCap)
	}
}

// TestFindRepoRoot verifies repo-root discovery + graceful miss (AC-TEF edge:
// repo path unavailable → guard skips).
func TestFindRepoRoot(t *testing.T) {
	// A bare temp dir has no go.mod ancestor → miss.
	bare := t.TempDir()
	if _, ok := findRepoRoot(bare); ok {
		t.Errorf("findRepoRoot(bare temp) = ok, want miss")
	}

	// A dir containing go.mod → hit.
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(root, "internal", "config")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	got, ok := findRepoRoot(sub)
	if !ok {
		t.Fatalf("findRepoRoot(sub) = miss, want hit")
	}
	// Resolve symlinks for macOS /var → /private/var normalization.
	gotResolved, _ := filepath.EvalSymlinks(got)
	rootResolved, _ := filepath.EvalSymlinks(root)
	if gotResolved != rootResolved {
		t.Errorf("findRepoRoot(sub) = %q, want %q", gotResolved, rootResolved)
	}
}
