// Package template — Navigator regeneration script tests (TDD RED→GREEN driver).
//
// Test fixture driver for the Project Navigator regeneration shell script
// shipped at templates/.claude/skills/moai-workflow-project/scripts/navigator-regen.sh.
// The script is the deterministic core: it reads a SPEC registry + git log,
// emits three markdown files under .moai/project/navigator/, and writes a
// last-regen-commit sentinel. These tests exercise AC-PN-001..008, 013, 015, 016
// against fixture projects built under t.TempDir().
//
// The script is invoked as a shell subprocess so the tests are black-box over
// its bash internals. A project fixture is materialized by writing SPEC
// directories + a git repo (init + add + commit), then running the script with
// CLAUDE_PROJECT_DIR pointed at the temp dir.
package template

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// navigatorRegenScript is the template-relative path to the regeneration script.
const navigatorRegenScript = "templates/.claude/skills/moai-workflow-project/scripts/navigator-regen.sh"

// sha40Re matches a 40-char lowercase hex git commit SHA.
var sha40Re = regexp.MustCompile(`^[0-9a-f]{40}$`)

// iso8601Re matches an ISO-8601 timestamp (git log --format=%cI shape:
// 2026-08-05T12:34:56+00:00 or trailing Z).
var iso8601Re = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}([+-]\d{2}:\d{2}|Z)$`)

// initFixtureRepo creates a git repo at dir with one initial commit and returns
// the HEAD SHA. The repo is configured with a stable committer identity so
// regeneration output is deterministic.
func initFixtureRepo(t *testing.T, dir string) string {
	t.Helper()
	if err := gitRun(dir, "init", "-q"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	for _, kv := range []struct{ k, v string }{
		{"user.name", "Fixture Author"},
		{"user.email", "fixture@example.test"},
		{"commit.gpgsign", "false"},
	} {
		if err := gitRun(dir, "config", kv.k, kv.v); err != nil {
			t.Fatalf("git config %s: %v", kv.k, err)
		}
	}
	// An empty initial commit needs --allow-empty and an index; create a marker.
	if err := os.WriteFile(filepath.Join(dir, ".gitkeep"), []byte(""), 0o644); err != nil {
		t.Fatalf("write .gitkeep: %v", err)
	}
	if err := gitRun(dir, "add", ".gitkeep"); err != nil {
		t.Fatalf("git add: %v", err)
	}
	if err := gitRun(dir, "commit", "-m", "initial fixture commit"); err != nil {
		t.Fatalf("git commit: %v", err)
	}
	sha, err := gitHead(dir)
	if err != nil {
		t.Fatalf("gitHead: %v", err)
	}
	return sha
}

func gitRun(dir string, args ...string) error {
	c := exec.Command("git", args...)
	c.Dir = dir
	c.Env = append(os.Environ(), "GIT_AUTHOR_DATE=2026-01-15T12:00:00+00:00", "GIT_COMMITTER_DATE=2026-01-15T12:00:00+00:00")
	var buf bytes.Buffer
	c.Stderr = &buf
	c.Stdout = &buf
	if err := c.Run(); err != nil {
		return &gitErr{args: args, err: err, out: buf.String()}
	}
	return nil
}

type gitErr struct {
	args []string
	err  error
	out  string
}

func (e *gitErr) Error() string { return e.err.Error() + "\n" + e.out }

func gitHead(dir string) (string, error) {
	c := exec.Command("git", "rev-parse", "HEAD")
	c.Dir = dir
	out, err := c.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// writeSPEC writes a SPEC directory with spec.md frontmatter at dir/.moai/specs/<id>/.
func writeSPEC(t *testing.T, dir, id, title, status, phase, module string) {
	t.Helper()
	specDir := filepath.Join(dir, ".moai", "specs", id)
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", specDir, err)
	}
	body := "---\n" +
		"id: " + id + "\n" +
		"title: \"" + title + "\"\n" +
		"version: \"0.1.0\"\n" +
		"status: " + status + "\n" +
		"created: 2026-01-10\n" +
		"updated: 2026-01-12\n" +
		"author: Fixture\n" +
		"priority: P1\n" +
		"phase: \"" + phase + "\"\n" +
		"module: " + module + "\n" +
		"lifecycle: spec-anchored\n" +
		"tags: \"fixture, navigator-test\"\n" +
		"---\n\n# " + id + " — " + title + "\n\nBody text.\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(body), 0o644); err != nil {
		t.Fatalf("write spec.md: %v", err)
	}
}

// runRegen runs the navigator-regen.sh script against dir. It returns the
// combined stdout+stderr, the exit error (or nil), and the effective script
// absolute path.
func runRegen(t *testing.T, dir string, envExtra ...string) (output string, exitErr error) {
	t.Helper()
	scriptAbs, err := filepath.Abs(navigatorRegenScript)
	if err != nil {
		t.Fatalf("resolve script abs: %v", err)
	}
	if _, err := os.Stat(scriptAbs); err != nil {
		t.Fatalf("navigator-regen.sh not found at %s — has M1 landed the script?", scriptAbs)
	}
	c := exec.Command("bash", scriptAbs)
	c.Dir = dir
	env := append(os.Environ(), "CLAUDE_PROJECT_DIR="+dir)
	env = append(env, envExtra...)
	c.Env = env
	var buf bytes.Buffer
	c.Stdout = &buf
	c.Stderr = &buf
	err = c.Run()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return buf.String(), ee
		}
		t.Fatalf("invoke script: %v", err)
	}
	return buf.String(), nil
}

// navigatorFiles returns the three expected file paths under dir.
func navigatorFiles(dir string) (nav, cap, prog string) {
	base := filepath.Join(dir, ".moai", "project", "navigator")
	return filepath.Join(base, "navigator.md"),
		filepath.Join(base, "capability-map.md"),
		filepath.Join(base, "progress-map.md")
}

// TestACPN001_ThreeFilesNoExtras verifies AC-PN-001: regeneration produces
// exactly three files under .moai/project/navigator/ and no other top-level
// Navigator file.
func TestACPN001_ThreeFilesNoExtras(t *testing.T) {
	dir := t.TempDir()
	initFixtureRepo(t, dir)
	writeSPEC(t, dir, "SPEC-AUTH-001", "Auth", "in-progress", "v1.0.0", "internal/auth")
	writeSPEC(t, dir, "SPEC-BILLING-002", "Billing", "completed", "v1.0.0", "internal/billing")
	if err := gitRun(dir, "add", ".moai/specs"); err != nil {
		t.Fatalf("git add specs: %v", err)
	}
	if err := gitRun(dir, "commit", "-m", "feat: add two SPEC fixtures"); err != nil {
		t.Fatalf("git commit: %v", err)
	}

	if _, err := runRegen(t, dir); err != nil {
		t.Fatalf("regen exited non-zero: %v", err)
	}

	base := filepath.Join(dir, ".moai", "project", "navigator")
	entries, err := os.ReadDir(base)
	if err != nil {
		t.Fatalf("read navigator dir: %v", err)
	}
	names := map[string]bool{}
	for _, e := range entries {
		names[e.Name()] = true
	}
	want := []string{"navigator.md", "capability-map.md", "progress-map.md"}
	for _, w := range want {
		if !names[w] {
			t.Errorf("AC-PN-001: expected %s to exist under navigator/", w)
		}
	}
	if len(names) != len(want) {
		t.Errorf("AC-PN-001: expected exactly %d top-level files, got %d (%v)", len(want), len(names), entries)
	}
}

// TestACPN002_EveryRowCarriesProvenance verifies AC-PN-002: every row in
// capability-map.md and progress-map.md carries a 40-char commit-sha and an
// ISO-8601 captured-at.
func TestACPN002_EveryRowCarriesProvenance(t *testing.T) {
	dir := t.TempDir()
	initFixtureRepo(t, dir)
	writeSPEC(t, dir, "SPEC-AUTH-001", "Auth", "in-progress", "v1.0.0", "internal/auth")
	if err := gitRun(dir, "add", ".moai/specs"); err != nil {
		t.Fatal(err)
	}
	if err := gitRun(dir, "commit", "-m", "feat: auth spec"); err != nil {
		t.Fatal(err)
	}
	if _, err := runRegen(t, dir); err != nil {
		t.Fatalf("regen: %v", err)
	}

	_, capFile, progFile := navigatorFiles(dir)
	for _, f := range []string{capFile, progFile} {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		// Each non-header row MUST carry | <40hex> | <iso8601> | provenance fields.
		lines := strings.Split(string(data), "\n")
		dataRows := 0
		for _, ln := range lines {
			ln = strings.TrimSpace(ln)
			if ln == "" || strings.HasPrefix(ln, "|--") || strings.HasPrefix(ln, "| spec") || strings.HasPrefix(ln, "| capability") || strings.HasPrefix(ln, "| spec-id") || strings.HasPrefix(ln, "| #") {
				continue
			}
			if !strings.HasPrefix(ln, "|") {
				continue
			}
			dataRows++
			// AC-PN-002: the row must carry a 40-char hex commit-sha AND an
			// ISO-8601 captured-at somewhere in its cells (not fixed columns —
			// capability-map and progress-map carry different column shapes).
			cells := strings.Split(strings.Trim(ln, "|"), "|")
			hasSHA, hasISO := false, false
			for _, c := range cells {
				cv := strings.Trim(strings.TrimSpace(c), "`")
				if sha40Re.MatchString(cv) {
					hasSHA = true
				}
				if iso8601Re.MatchString(cv) {
					hasISO = true
				}
			}
			if !hasSHA {
				t.Errorf("AC-PN-002: row in %s missing 40-char commit-sha: %q", f, ln)
			}
			if !hasISO {
				t.Errorf("AC-PN-002: row in %s missing ISO-8601 captured-at: %q", f, ln)
			}
		}
		if dataRows == 0 {
			t.Errorf("AC-PN-002: %s has no data rows to verify", f)
		}
	}
}

// TestACPN005_IdempotentRegeneration verifies AC-PN-005: two successive
// regenerations on the same commit produce byte-identical output.
func TestACPN005_IdempotentRegeneration(t *testing.T) {
	dir := t.TempDir()
	initFixtureRepo(t, dir)
	writeSPEC(t, dir, "SPEC-X-001", "X", "draft", "v1.0.0", "internal/x")
	if err := gitRun(dir, "add", ".moai/specs"); err != nil {
		t.Fatal(err)
	}
	if err := gitRun(dir, "commit", "-m", "feat: x"); err != nil {
		t.Fatal(err)
	}
	if _, err := runRegen(t, dir); err != nil {
		t.Fatal(err)
	}
	nav1, cap1, prog1 := navigatorFiles(dir)
	nav1b, _ := os.ReadFile(nav1)
	cap1b, _ := os.ReadFile(cap1)
	prog1b, _ := os.ReadFile(prog1)

	// Sleep past any sub-second granularity, then re-run.
	time.Sleep(1100 * time.Millisecond)
	if _, err := runRegen(t, dir); err != nil {
		t.Fatal(err)
	}
	nav2, _ := os.ReadFile(nav1)
	cap2, _ := os.ReadFile(cap1)
	prog2, _ := os.ReadFile(prog1)
	if !bytes.Equal(nav1b, nav2) {
		t.Errorf("AC-PN-005: navigator.md differs between two regens (idempotence broken)")
	}
	if !bytes.Equal(cap1b, cap2) {
		t.Errorf("AC-PN-005: capability-map.md differs between two regens (idempotence broken)")
	}
	if !bytes.Equal(prog1b, prog2) {
		t.Errorf("AC-PN-005: progress-map.md differs between two regens (idempotence broken)")
	}
}

// TestACPN006_EmptyProjectResilience verifies AC-PN-006: a project with zero
// SPECs and zero commits still exits 0 and emits the placeholder.
func TestACPN006_EmptyProjectResilience(t *testing.T) {
	dir := t.TempDir()
	// Zero commits: init repo but do not commit.
	if err := gitRun(dir, "init", "-q"); err != nil {
		t.Fatal(err)
	}
	if err := gitRun(dir, "config", "user.name", "Fixture"); err != nil {
		t.Fatal(err)
	}
	if err := gitRun(dir, "config", "user.email", "fixture@example.test"); err != nil {
		t.Fatal(err)
	}
	out, err := runRegen(t, dir)
	if err != nil {
		t.Fatalf("AC-PN-006: regen on empty project exited non-zero: %v\noutput:\n%s", err, out)
	}
	nav, _, _ := navigatorFiles(dir)
	data, rerr := os.ReadFile(nav)
	if rerr != nil {
		t.Fatalf("read navigator.md: %v", rerr)
	}
	if !strings.Contains(string(data), "no features tracked yet") {
		t.Errorf("AC-PN-006: navigator.md missing placeholder; got:\n%s", data)
	}
}

// TestACPN007_MalformedFrontmatterTolerance verifies AC-PN-007: a SPEC with
// malformed frontmatter is skipped (warning logged), the rest are regenerated.
func TestACPN007_MalformedFrontmatterTolerance(t *testing.T) {
	dir := t.TempDir()
	initFixtureRepo(t, dir)
	writeSPEC(t, dir, "SPEC-GOOD-001", "Good", "completed", "v1.0.0", "internal/good")
	writeSPEC(t, dir, "SPEC-GOOD-002", "Good2", "draft", "v1.0.0", "internal/good2")

	// Malformed: no closing frontmatter delimiter, missing required fields.
	badDir := filepath.Join(dir, ".moai", "specs", "SPEC-BAD-001")
	if err := os.MkdirAll(badDir, 0o755); err != nil {
		t.Fatal(err)
	}
	badBody := "---\nid: SPEC-BAD-001\n# frontmatter truncated — no closing ---\n"
	if err := os.WriteFile(filepath.Join(badDir, "spec.md"), []byte(badBody), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := gitRun(dir, "add", ".moai/specs"); err != nil {
		t.Fatal(err)
	}
	if err := gitRun(dir, "commit", "-m", "specs"); err != nil {
		t.Fatal(err)
	}
	out, err := runRegen(t, dir)
	if err != nil {
		t.Fatalf("AC-PN-007: regen exited non-zero on malformed frontmatter: %v\n%s", err, out)
	}

	capFile := filepath.Join(dir, ".moai", "project", "navigator", "capability-map.md")
	data, _ := os.ReadFile(capFile)
	body := string(data)
	if strings.Contains(body, "SPEC-BAD-001") {
		t.Errorf("AC-PN-007: malformed SPEC-BAD-001 appeared in capability-map.md (should be skipped):\n%s", body)
	}
	if !strings.Contains(body, "SPEC-GOOD-001") || !strings.Contains(body, "SPEC-GOOD-002") {
		t.Errorf("AC-PN-007: good SPECs missing from capability-map.md:\n%s", body)
	}
	// Warning logged.
	warnLog := filepath.Join(dir, ".moai", "logs", "navigator-warnings.log")
	wdata, _ := os.ReadFile(warnLog)
	if !strings.Contains(string(wdata), "SPEC-BAD-001") {
		t.Errorf("AC-PN-007: warning log %s missing SPEC-BAD-001 mention; got:\n%s", warnLog, wdata)
	}
}

// TestACPN008_AtomicRenameDeterministic verifies AC-PN-008: the atomic-rename
// strategy is observed under a synchronized barrier. The script supports an env
// var NAVIGATOR_PRE_RENAME_BARRIER: a path. When set, after writing each <file>.tmp
// but before the mv, the script writes "ready" to the barrier path and blocks
// (read loop) until the barrier file is removed by the test. This lets the test
// read the target path mid-flight deterministically — no timing polling.
func TestACPN008_AtomicRenameDeterministic(t *testing.T) {
	dir := t.TempDir()
	initFixtureRepo(t, dir)
	writeSPEC(t, dir, "SPEC-A-001", "A", "completed", "v1.0.0", "internal/a")
	if err := gitRun(dir, "add", ".moai/specs"); err != nil {
		t.Fatal(err)
	}
	if err := gitRun(dir, "commit", "-m", "a"); err != nil {
		t.Fatal(err)
	}

	// Seed an OLD navigator.md so the barrier read observes the previous version.
	nav, _, _ := navigatorFiles(dir)
	if err := os.MkdirAll(filepath.Dir(nav), 0o755); err != nil {
		t.Fatal(err)
	}
	oldBody := []byte("# OLD navigator (pre-regen)\n")
	if err := os.WriteFile(nav, oldBody, 0o644); err != nil {
		t.Fatal(err)
	}

	// Barrier path: script writes "ready" here and blocks until the file is deleted.
	// A companion "<barrier>.armed" sentinel arms the barrier — the script fires
	// it exactly once (on the first atomic_write), then consumes the sentinel.
	barrier := filepath.Join(dir, ".moai", "state", "navigator-barrier.flag")
	armed := barrier + ".armed"
	if err := os.MkdirAll(filepath.Dir(barrier), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(armed, []byte("1"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Launch regen with the barrier armed.
	doneCh := make(chan error, 1)
	go func() {
		_, err := runRegen(t, dir, "NAVIGATOR_PRE_RENAME_BARRIER="+barrier)
		doneCh <- err
	}()

	// Wait for the script to signal "ready" (barrier file exists with content "ready").
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		if b, err := os.ReadFile(barrier); err == nil && strings.TrimSpace(string(b)) == "ready" {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	// Mid-flight read: the target must still be the OLD version (mv hasn't landed).
	mid, err := os.ReadFile(nav)
	if err != nil {
		t.Fatalf("AC-PN-008: read navigator.md mid-flight: %v", err)
	}
	if !bytes.Equal(mid, oldBody) {
		t.Errorf("AC-PN-008: mid-flight read observed a non-old version — atomic-rename leak:\nwant: %q\ngot:  %q", oldBody, mid)
	}

	// Release the barrier.
	if err := os.Remove(barrier); err != nil {
		t.Fatalf("AC-PN-008: remove barrier: %v", err)
	}
	// Wait for completion.
	select {
	case err := <-doneCh:
		if err != nil {
			t.Fatalf("AC-PN-008: regen exited non-zero after barrier release: %v", err)
		}
	case <-time.After(15 * time.Second):
		t.Fatalf("AC-PN-008: regen did not complete after barrier release (deadlock)")
	}

	// Post-mv read: the target must now be the NEW version (differs from OLD).
	post, err := os.ReadFile(nav)
	if err != nil {
		t.Fatalf("AC-PN-008: read navigator.md post-mv: %v", err)
	}
	if bytes.Equal(post, oldBody) {
		t.Errorf("AC-PN-008: post-mv read observed the OLD version — the mv never landed")
	}
}

// TestACPN013_NonDuplication verifies AC-PN-013: rows are references, not
// content copies — no row exceeds ~200 chars and no SPEC body content leaks.
func TestACPN013_NonDuplication(t *testing.T) {
	dir := t.TempDir()
	initFixtureRepo(t, dir)
	writeSPEC(t, dir, "SPEC-LONG-001", "LongFeature", "in-progress", "v1.0.0", "internal/long")
	// Inject a long body string that MUST NOT appear in capability-map.md.
	specDir := filepath.Join(dir, ".moai", "specs", "SPEC-LONG-001")
	longBody := "UNIQUE_LONG_BODY_TOKEN_" + strings.Repeat("x", 250)
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"),
		[]byte("---\nid: SPEC-LONG-001\ntitle: \"LongFeature\"\nstatus: in-progress\nphase: \"v1.0.0\"\nmodule: internal/long\n---\n\n# Body\n\n"+longBody+"\n"),
		0o644); err != nil {
		t.Fatal(err)
	}
	if err := gitRun(dir, "add", ".moai/specs"); err != nil {
		t.Fatal(err)
	}
	if err := gitRun(dir, "commit", "-m", "long"); err != nil {
		t.Fatal(err)
	}
	if _, err := runRegen(t, dir); err != nil {
		t.Fatal(err)
	}
	_, capFile, _ := navigatorFiles(dir)
	data, _ := os.ReadFile(capFile)
	body := string(data)
	if strings.Contains(body, "UNIQUE_LONG_BODY_TOKEN_") {
		t.Errorf("AC-PN-013: capability-map.md leaked SPEC body content (non-duplication violation)")
	}
	for _, ln := range strings.Split(body, "\n") {
		if strings.HasPrefix(strings.TrimSpace(ln), "|") && len(ln) > 200 {
			t.Errorf("AC-PN-013: row exceeds 200 chars (non-duplication): %q", ln)
		}
	}
}

// TestACPN015_NonGoProject verifies AC-PN-015: regeneration succeeds on a
// non-Go (Python-only) fixture project.
func TestACPN015_NonGoProject(t *testing.T) {
	dir := t.TempDir()
	initFixtureRepo(t, dir)
	// Materialize a Python-only project marker.
	if err := os.WriteFile(filepath.Join(dir, "pyproject.toml"),
		[]byte("[project]\nname = \"fixture-py\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.py"),
		[]byte("print('hello')\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeSPEC(t, dir, "SPEC-PY-001", "PyFeature", "draft", "v1.0.0", "src/feature")
	if err := gitRun(dir, "add", "."); err != nil {
		t.Fatal(err)
	}
	if err := gitRun(dir, "commit", "-m", "py fixture"); err != nil {
		t.Fatal(err)
	}
	out, err := runRegen(t, dir)
	if err != nil {
		t.Fatalf("AC-PN-015: regen failed on Python fixture: %v\n%s", err, out)
	}
	nav, _, _ := navigatorFiles(dir)
	if _, err := os.Stat(nav); err != nil {
		t.Errorf("AC-PN-015: navigator.md not produced for Python fixture: %v", err)
	}
}

// TestNavigatorRegen_NextTask_ExcludesImplemented_PrefersInProgress verifies
// SPEC-PROJECT-NAVIGATOR-004 AC-001/AC-002/AC-003: the "Next task" line must
// exclude implemented SPECs and prefer in-progress SPECs over draft SPECs,
// while the "Current frontier" display list stays inclusive of implemented.
//
// Fixture (alphabetically ordered so the pre-fix bug would misclassify):
//   SPEC-A-001 — status implemented (alphabetically first)
//   SPEC-B-001 — status in-progress
//   SPEC-C-001 — status draft
//
// Pre-fix, "Next task" = first non-terminal by alphabetical sort = SPEC-A-001
// (the misclassification). Post-fix, the positive status-tier predicate selects
// SPEC-B-001 (in-progress). SPEC-A-001 must remain in "Current frontier".
func TestNavigatorRegen_NextTask_ExcludesImplemented_PrefersInProgress(t *testing.T) {
	dir := t.TempDir()
	initFixtureRepo(t, dir)
	writeSPEC(t, dir, "SPEC-A-001", "alpha", "implemented", "v1.0.0", "internal/a")
	writeSPEC(t, dir, "SPEC-B-001", "beta", "in-progress", "v1.0.0", "internal/b")
	writeSPEC(t, dir, "SPEC-C-001", "gamma", "draft", "v1.0.0", "internal/c")
	if err := gitRun(dir, "add", ".moai/specs"); err != nil {
		t.Fatalf("git add specs: %v", err)
	}
	if err := gitRun(dir, "commit", "-m", "fixture: implemented + in-progress + draft"); err != nil {
		t.Fatalf("git commit: %v", err)
	}

	if _, err := runRegen(t, dir); err != nil {
		t.Fatalf("regen exited non-zero: %v", err)
	}

	nav, _, _ := navigatorFiles(dir)
	data, err := os.ReadFile(nav)
	if err != nil {
		t.Fatalf("read navigator.md: %v", err)
	}
	body := string(data)

	// AC-003: "Current frontier" remains inclusive of implemented SPECs.
	if !strings.Contains(body, "SPEC-A-001") {
		t.Errorf("AC-003: SPEC-A-001 (implemented) is MISSING from Current frontier (display regression — frontier must stay inclusive)")
	}

	// Slice the "## Next task" section so the assertions scope to the
	// recommendation, not the whole document (the frontier list also names
	// every SPEC). The section runs from its heading to the next "## " heading
	// or end of file.
	nextTaskSection := extractSection(body, "## Next task")

	// AC-001: implemented SPEC excluded from "Next task".
	if strings.Contains(nextTaskSection, "SPEC-A-001") {
		t.Errorf("AC-001: SPEC-A-001 (implemented) appeared in Next task section (must be excluded):\n%s", nextTaskSection)
	}
	// AC-002: in-progress preferred over draft.
	if !strings.Contains(nextTaskSection, "SPEC-B-001") {
		t.Errorf("AC-002: SPEC-B-001 (in-progress) is NOT the Next task (must be preferred over draft):\n%s", nextTaskSection)
	}
	if strings.Contains(nextTaskSection, "SPEC-C-001") {
		t.Errorf("AC-002: SPEC-C-001 (draft) appeared in Next task over the in-progress SPEC (tier order wrong):\n%s", nextTaskSection)
	}
}

// extractSection returns the body of the markdown section introduced by heading
// `## <name>` up to the next `## ` heading or end of text. Whitespace-only
// matching of the heading name.
func extractSection(md, name string) string {
	lines := strings.Split(md, "\n")
	var (
		out      strings.Builder
		capturing bool
	)
	for _, ln := range lines {
		trim := strings.TrimSpace(ln)
		if strings.HasPrefix(trim, "## ") {
			if capturing {
				break // next ## heading ends the section
			}
			if strings.TrimSpace(strings.TrimPrefix(trim, "##")) == strings.TrimSpace(strings.TrimPrefix(name, "##")) {
				capturing = true
				continue
			}
		}
		if capturing {
			out.WriteString(ln)
			out.WriteByte('\n')
		}
	}
	return out.String()
}

// TestACPN016_LSELBoundary verifies AC-PN-016: the script does NOT touch any
// LSEL surface. We assert by creating sentinel LSEL files and confirming the
// script neither reads nor deletes them (write-set is only navigator/ + state/navigator/).
func TestACPN016_LSELBoundary(t *testing.T) {
	dir := t.TempDir()
	initFixtureRepo(t, dir)
	writeSPEC(t, dir, "SPEC-Q-001", "Q", "draft", "v1.0.0", "internal/q")
	if err := gitRun(dir, "add", ".moai/specs"); err != nil {
		t.Fatal(err)
	}
	if err := gitRun(dir, "commit", "-m", "q"); err != nil {
		t.Fatal(err)
	}

	// LSEL sentinels: these MUST survive untouched.
	lselInbox := filepath.Join(dir, ".moai", "lessons-inbox.jsonl")
	lselState := filepath.Join(dir, ".moai", "state", "lsel", "clusters.json")
	for _, p := range []string{lselInbox, lselState} {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("LSEL_SENTINEL\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := runRegen(t, dir); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{lselInbox, lselState} {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Errorf("AC-PN-016: LSEL surface %s was touched/deleted by regen: %v", p, err)
			continue
		}
		if strings.TrimSpace(string(b)) != "LSEL_SENTINEL" {
			t.Errorf("AC-PN-016: LSEL surface %s was modified by regen", p)
		}
	}
}
