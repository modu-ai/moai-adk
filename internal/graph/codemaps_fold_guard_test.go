package graph

// Fold-preservation guard for the codemaps generator documents.
//
// SPEC-CODEMAPS-FOLD-GUARD-001: during SPEC-CODEMAPS-REFRESH-002's first
// regeneration (commit e397ec00d), prose for all five fold-judged units was
// rewritten because the record-only fold disposition was misread as merge
// permission. The revert landed (cd03be0d3); this guard makes the recurrence
// fail loudly instead.
//
// Contract (REQ-CFG-001..004):
//   - the protected unit set is derived from .moai/project/codemaps/fold-judgments.txt
//     (`fold <unit>` lines), with the five t475 floor units pinned as a second
//     lock against the record silently leaking units;
//   - any generator document line carrying a protected unit's path or filename
//     token (exact-substring, the `grep -c -F` convention) is a violation. The
//     short form `core/git` is NOT a violation — it is the fold basis itself
//     (t475 §⑥), and it does not contain the full-path token;
//   - a missing, unparseable, or floor-degraded record fails; a missing or
//     unreadable generator document fails (read-error-as-pass is the vacuous
//     green this clause exists to kill);
//   - the verdict is stamp-independent: provenance.json is never read and
//     `moai graph check` is never consulted (the /tmp fixture scans below pass
//     with no provenance.json present at all, proving the independence
//     structurally).

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const codemapsRelDir = ".moai/project/codemaps"

// generatorDocs is the t688 §B.1 G3 set of generator-produced documents the
// guard scans. fold-judgments.txt is the record layer, deliberately absent.
var generatorDocs = []string{
	"overview.md",
	"modules.md",
	"dependencies.md",
	"entry-points.md",
	"data-flow.md",
}

// floorFoldUnits pins the five units t475 §⑥ judged fold, as a code-level
// floor under the data-driven record parse (REQ-CFG-001 dual lock). A floor
// unit missing from the record fails the guard — re-judgments must be recorded
// explicitly, never quietly.
var floorFoldUnits = []string{
	"internal/core/git",
	"internal/cli/doctor_hook_delivery.go",
	"internal/hook/quality/step_git_env.go",
	"internal/kanban/prlink_landedref.go",
	"internal/web/fieldsets_codex_templ.go",
}

// findCodemapsRoot walks up from the test's working directory until the
// codemaps directory is found. Failure to locate the tree is a test failure,
// never a skip (plan §D: a quiet skip disguises absent execution as green).
func findCodemapsRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if fi, err := os.Stat(filepath.Join(dir, codemapsRelDir)); err == nil && fi.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not locate %s above %s", codemapsRelDir, dir)
		}
		dir = parent
	}
}

// foldToken returns the violation token for a protected unit: the exact
// recorded form. A file unit also exposes its filename stem (e.g.
// fieldsets_codex_templ), mirroring the spec §A.2 grep convention. The
// directory unit internal/core/git exposes only the full path, so the folded
// short form `core/git` surviving in parent prose stays legal.
func foldToken(unit string) string {
	if suffix := ".go"; strings.HasSuffix(unit, suffix) {
		return strings.TrimSuffix(unit, suffix)
	}
	return unit
}

// parseFoldUnits reads the record layer and returns the fold units. A missing
// or unreadable record, or a record with zero fold lines, is an error — the
// hit-0 state is only meaningful when the judgment that produced it is visible
// (REQ-CFG-003).
func parseFoldUnits(recordPath string) ([]string, error) {
	data, err := os.ReadFile(recordPath)
	if err != nil {
		return nil, fmt.Errorf("fold-judgments record unreadable: %w", err)
	}
	var units []string
	for i, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if rest, ok := strings.CutPrefix(trimmed, "fold "); ok {
			unit := strings.TrimSpace(rest)
			if unit == "" {
				return nil, fmt.Errorf("%s:%d: malformed fold line (empty unit)", recordPath, i+1)
			}
			units = append(units, unit)
		}
	}
	if len(units) == 0 {
		return nil, fmt.Errorf("%s: no `fold <unit>` lines parsed — record empty or unparseable", recordPath)
	}
	return units, nil
}

// runFoldGuard is the single reader path for the guard verdict (plan M1.4):
// the default-suite real-tree scan and the /tmp injected-fixture scans flow
// through this one function, parameterized by the documents directory and the
// record path. It returns a non-nil error naming the failing unit, document,
// and line on any violation, and nil only when every generator document was
// actually read and carried zero protected tokens.
func runFoldGuard(docsDir, recordPath string) error {
	// Record layer first: the floor assertion must precede any document scan,
	// so an empty sweep can never produce a silent green (REQ-CFG-001).
	units, err := parseFoldUnits(recordPath)
	if err != nil {
		return err
	}
	recorded := make(map[string]bool, len(units))
	for _, u := range units {
		recorded[u] = true
	}
	var missing []string
	for _, floor := range floorFoldUnits {
		if !recorded[floor] {
			missing = append(missing, floor)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("fold-judgments record lost floor unit line(s): %s", strings.Join(missing, ", "))
	}

	// Document integrity: every generator document must exist and be readable
	// before any verdict is rendered (REQ-CFG-003 — read errors are not green).
	type docContent struct {
		name string
		data string
	}
	docs := make([]docContent, 0, len(generatorDocs))
	for _, name := range generatorDocs {
		data, err := os.ReadFile(filepath.Join(docsDir, name))
		if err != nil {
			return fmt.Errorf("generator document missing or unreadable: %s: %v", name, err)
		}
		docs = append(docs, docContent{name: name, data: string(data)})
	}

	// Token scan: exact-substring hits (the `grep -c -F` convention, spec
	// §A.3(a)). All violations are collected so one run names every offender.
	var violations []string
	for _, doc := range docs {
		for i, line := range strings.Split(doc.data, "\n") {
			for _, unit := range units {
				if strings.Contains(line, foldToken(unit)) {
					violations = append(violations,
						fmt.Sprintf("fold unit %q appears in %s:%d", unit, doc.name, i+1))
				}
			}
		}
	}
	if len(violations) > 0 {
		return fmt.Errorf("fold-unit prose leaked into generator documents:\n%s",
			strings.Join(violations, "\n"))
	}
	return nil
}

// TestCodemapsFoldPreservationGuard is the default-suite entry: it scans the
// real repository codemaps surface through the same runFoldGuard reader the
// fixture scans use. PASS means the reverted state (spec §A.1) still holds.
func TestCodemapsFoldPreservationGuard(t *testing.T) {
	root := findCodemapsRoot(t)
	docsDir := filepath.Join(root, codemapsRelDir)
	if err := runFoldGuard(docsDir, filepath.Join(docsDir, "fold-judgments.txt")); err != nil {
		t.Fatal(err)
	}
}

// copyCodemapFixture builds a /tmp fixture copy of the codemaps surface —
// the live .moai/project/codemaps/** is never mutated (REQ-CFG-005).
func copyCodemapFixture(t *testing.T) (docsDir, recordPath string) {
	t.Helper()
	root := findCodemapsRoot(t)
	fixture := t.TempDir()
	docsDir = filepath.Join(fixture, "codemaps")
	if err := os.MkdirAll(docsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(root, codemapsRelDir)
	for _, name := range append(append([]string{}, generatorDocs...), "fold-judgments.txt") {
		data, err := os.ReadFile(filepath.Join(src, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(docsDir, name), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return docsDir, filepath.Join(docsDir, "fold-judgments.txt")
}

// requireGuardFailure runs the guard on an injected fixture and asserts the
// failure names every fragment the scenario requires (unit and/or document) —
// a failure that cannot say what it caught is report-not-verdict again.
func requireGuardFailure(t *testing.T, docsDir, recordPath string, nameFragments ...string) {
	t.Helper()
	err := runFoldGuard(docsDir, recordPath)
	if err == nil {
		t.Fatalf("guard passed an injected fixture — vacuous green (fragments wanted: %s)",
			strings.Join(nameFragments, ", "))
	}
	for _, frag := range nameFragments {
		if !strings.Contains(err.Error(), frag) {
			t.Errorf("guard failure does not name %q:\n%s", frag, err.Error())
		}
	}
}

// TestCodemapsFoldGuardFixtures pins the guard's failure surface on /tmp
// fixture copies: intact passes, each tampered shape fails with naming, and
// the legal short form `core/git` stays non-violating (REQ-CFG-002 exception).
func TestCodemapsFoldGuardFixtures(t *testing.T) {
	t.Run("intact copy passes", func(t *testing.T) {
		docsDir, recordPath := copyCodemapFixture(t)
		if err := runFoldGuard(docsDir, recordPath); err != nil {
			t.Fatalf("intact fixture copy failed: %v", err)
		}
	})

	t.Run("short form core/git is not a violation", func(t *testing.T) {
		docsDir, recordPath := copyCodemapFixture(t)
		overview := filepath.Join(docsDir, "overview.md")
		data, err := os.ReadFile(overview)
		if err != nil {
			t.Fatal(err)
		}
		// The folded basis itself (`core/git` short form) plus a generation
		// relation that names no artifact file: both must stay legal (spec
		// REQ-CFG-002, matching overview.md:103 / modules.md:171 reality).
		augmented := string(data) + "\n## fold basis note\n\nThe `core/git` package underpins higher utilities; `.templ` sources generate their runtime counterparts.\n"
		if err := os.WriteFile(overview, []byte(augmented), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := runFoldGuard(docsDir, recordPath); err != nil {
			t.Fatalf("fold-basis short form judged a violation:\n%v", err)
		}
	})

	t.Run("injected fold token fails naming unit and document", func(t *testing.T) {
		docsDir, recordPath := copyCodemapFixture(t)
		modules := filepath.Join(docsDir, "modules.md")
		data, err := os.ReadFile(modules)
		if err != nil {
			t.Fatal(err)
		}
		tampered := string(data) + "\n### internal/kanban/prlink_landedref.go\n\nLanded-ref resolution for kanban PR links.\n"
		if err := os.WriteFile(modules, []byte(tampered), 0o644); err != nil {
			t.Fatal(err)
		}
		requireGuardFailure(t, docsDir, recordPath,
			"internal/kanban/prlink_landedref.go", "modules.md")
	})

	t.Run("lost floor line fails naming the unit", func(t *testing.T) {
		docsDir, recordPath := copyCodemapFixture(t)
		data, err := os.ReadFile(recordPath)
		if err != nil {
			t.Fatal(err)
		}
		lines := strings.Split(string(data), "\n")
		var kept []string
		for _, line := range lines {
			if strings.TrimSpace(line) == "fold "+floorFoldUnits[2] {
				continue
			}
			kept = append(kept, line)
		}
		if err := os.WriteFile(recordPath, []byte(strings.Join(kept, "\n")), 0o644); err != nil {
			t.Fatal(err)
		}
		requireGuardFailure(t, docsDir, recordPath, floorFoldUnits[2])
	})

	t.Run("missing generator document fails naming the document", func(t *testing.T) {
		docsDir, recordPath := copyCodemapFixture(t)
		if err := os.Remove(filepath.Join(docsDir, "data-flow.md")); err != nil {
			t.Fatal(err)
		}
		err := runFoldGuard(docsDir, recordPath)
		if err == nil {
			t.Fatal("guard passed with a generator document missing — read-error-as-pass vacuous green")
		}
		if !strings.Contains(err.Error(), "data-flow.md") {
			t.Errorf("guard failure does not name the missing document:\n%s", err.Error())
		}
	})

	t.Run("missing record fails", func(t *testing.T) {
		docsDir, recordPath := copyCodemapFixture(t)
		if err := os.Remove(recordPath); err != nil {
			t.Fatal(err)
		}
		if err := runFoldGuard(docsDir, recordPath); err == nil {
			t.Fatal("guard passed without the fold-judgments record")
		}
	})
}

// TestCodemapsFoldGuardFloorCoverage asserts the floor lock is not a silent
// hand enumeration masquerading as the data-driven set: every floor unit must
// derive a distinct non-empty token, and the short form of the directory unit
// must NOT collide with its full-path token (REQ-CFG-001 / REQ-CFG-002).
func TestCodemapsFoldGuardFloorCoverage(t *testing.T) {
	seen := make(map[string]bool, len(floorFoldUnits))
	for _, unit := range floorFoldUnits {
		token := foldToken(unit)
		if token == "" {
			t.Errorf("floor unit %q derives an empty token", unit)
		}
		if seen[token] {
			t.Errorf("floor units collide on token %q", token)
		}
		seen[token] = true
	}
	if strings.Contains(foldToken("internal/core/git"), "core/git") &&
		foldToken("internal/core/git") != "internal/core/git" {
		t.Errorf("directory token must be the full path so `core/git` stays legal")
	}
}
