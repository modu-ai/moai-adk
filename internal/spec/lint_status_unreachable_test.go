package spec

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// SPEC-SPECLINT-GITBLIND-001 (card t371) — StatusGitUnreachable observability
// tests. They pin the M1 contract: a lint run whose git-implied status signal
// is UNOBSERVED (as opposed to observed-and-matching) must surface an
// Info-severity StatusGitUnreachable finding, at most once per Lint() run.
//
// Fixture conventions follow the package precedent (acceptance.md 공통 픽스처
// 규약):
//   - repositories live under t.TempDir();
//   - tests chdir into the fixture via chdirForTest (the git walkers query
//     the process working directory, not a parameter);
//   - NO t.Parallel(): the per-run git-query cache and os.Chdir are
//     process-global (drift_characterization_test.go:53 precedent).
//
// The spec.md bodies here are NOT copies of fixtureSpecMD (frontmatter-only —
// its body would emit MissingExclusions). Each body carries a `## 4. Scope`
// section with a `### 4.1 Out of Scope — <suffix>` H3 subsection, matching
// the schema-valid form verified in .moai/reports/t371/repro/withscope/spec.md
// (AC-SLGB-001 [HARD] premise: the fixture must be finding-free
// pre-implementation so the `✓ No findings` short-circuit actually fires).

// unreachableSpecMD renders a schema-valid spec.md for the observability
// fixtures (see the file comment for why the Out of Scope subsection is
// mandatory here).
func unreachableSpecMD(id, status string) string {
	var b strings.Builder

	b.WriteString("---\n")
	fmt.Fprintf(&b, "id: %s\n", id)
	fmt.Fprintf(&b, "title: %q\n", id+" observability fixture")
	b.WriteString("version: \"0.1.0\"\n")
	fmt.Fprintf(&b, "status: %s\n", status)
	b.WriteString("created: 2026-09-02\n")
	b.WriteString("updated: 2026-09-02\n")
	b.WriteString("author: t371\n")
	b.WriteString("priority: P1\n")
	b.WriteString("phase: \"v3.1.4 target\"\n")
	b.WriteString("module: \"internal/spec\"\n")
	b.WriteString("lifecycle: spec-anchored\n")
	b.WriteString("tags: \"fixture\"\n")
	b.WriteString("---\n\n")
	fmt.Fprintf(&b, "# SPEC: %s\n\n", id)
	b.WriteString("## 1. Overview\n\n")
	b.WriteString("Fixture exercising the StatusGitConsistency observability path.\n\n")
	b.WriteString("## 2. Requirements\n\n")
	b.WriteString("- REQ-URO-001 (Ubiquitous): The linter shall classify git observability.\n\n")
	b.WriteString("## 3. Acceptance Criteria\n\n")
	b.WriteString("- AC-URO-001: Given the fixture, when lint runs, then observability is reported.\n\n")
	b.WriteString("## 4. Scope\n\n")
	b.WriteString("### 4.1 Out of Scope — unrelated surfaces\n\n")
	b.WriteString("- Anything outside this fixture.\n")

	return b.String()
}

// writeUnreachableSpecs materializes one spec.md per id under
// <root>/.moai/specs/<id>/ and returns the file paths in argument order.
func writeUnreachableSpecs(t *testing.T, root, status string, ids ...string) []string {
	t.Helper()

	var paths []string
	for _, id := range ids {
		dir := filepath.Join(root, ".moai", "specs", id)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s failed: %v", dir, err)
		}
		p := filepath.Join(dir, "spec.md")
		if err := os.WriteFile(p, []byte(unreachableSpecMD(id, status)), 0o644); err != nil {
			t.Fatalf("write %s failed: %v", p, err)
		}
		paths = append(paths, p)
	}
	return paths
}

// setupGitRepoAt is setupGitFixture with a caller-chosen initial branch. The
// observability fixtures need repositories whose refs sit OUTSIDE the base-ref
// resolution chain (an unborn `develop`), which the shared helper's
// hardcoded `-b main` cannot express. Commits are applied oldest → newest.
func setupGitRepoAt(t *testing.T, initialBranch string, commits []fixtureCommit) string {
	t.Helper()

	dir := t.TempDir()
	runGit := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v\noutput: %s", args, err, out)
		}
	}

	runGit("init", "-q", "-b", initialBranch, ".")
	runGit("config", "user.email", "t371@example.com")
	runGit("config", "user.name", "t371")

	for _, c := range commits {
		msg := c.title
		if c.body != "" {
			msg = fmt.Sprintf("%s\n\n%s", c.title, c.body)
		}
		runGit("commit", "--allow-empty", "-q", "-m", msg)
	}
	return dir
}

// setupShallowCloneFixture builds a depth-1 clone of a source repository on
// `main`, writes the given SPEC fixtures into it, and chdirs into the clone.
// The clone is a REAL shallow repository (git rev-parse
// --is-shallow-repository → true) whose base ref resolves (the clone checks
// out a local main) — exactly the AC-SLGB-004 precondition: shapes ②/③ must
// fire because the window is truncated, not because the base ref is missing.
func setupShallowCloneFixture(t *testing.T, srcCommits []fixtureCommit, status string, ids ...string) []string {
	t.Helper()

	src := setupGitRepoAt(t, "main", srcCommits)
	dst := filepath.Join(t.TempDir(), "clone")
	cmd := exec.Command("git", "clone", "-q", "--depth", "1", "file://"+src, dst)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("shallow clone failed: %v\noutput: %s", err, out)
	}
	paths := writeUnreachableSpecs(t, dst, status, ids...)
	chdirForTest(t, dst)
	return paths
}

// runLintOnPaths runs one full Lint() pass (per-run git-query cache active)
// over the given spec paths.
func runLintOnPaths(t *testing.T, strict bool, paths ...string) *Report {
	t.Helper()

	l := NewLinter(LinterOptions{Strict: strict})
	report, err := l.Lint(paths)
	if err != nil {
		t.Fatalf("Lint failed: %v", err)
	}
	return report
}

// findingsByCode filters a report down to one finding code.
func findingsByCode(report *Report, code string) []Finding {
	var out []Finding
	for _, f := range report.Findings {
		if f.Code == code {
			out = append(out, f)
		}
	}
	return out
}

// AC-SLGB-001: with no base ref resolvable anywhere, the lint run must
// surface a StatusGitUnreachable finding. printTable emits its
// `✓ No findings — all SPEC documents are valid` line iff
// len(report.Findings) == 0 (internal/cli/spec_lint.go:115-118), so this
// finding's existence is exactly that line's absence (spec.md §1.2); the CLI
// surface itself is observed directly in progress.md §E.2.
func TestStatusGitUnreachable_NoBaseRef(t *testing.T) {
	root := setupGitRepoAt(t, "develop", []fixtureCommit{{title: "chore: seed"}})
	paths := writeUnreachableSpecs(t, root, "draft", "SPEC-URO-001")
	chdirForTest(t, root)

	report := runLintOnPaths(t, true, paths...)

	got := findingsByCode(report, "StatusGitUnreachable")
	if len(got) != 1 {
		t.Fatalf("StatusGitUnreachable findings = %d, want 1 (report held %d findings total)", len(got), len(report.Findings))
	}
}

// AC-SLGB-002 (REQ-SLGB-002): the finding message names the candidate refs
// whose resolution was attempted and states the repository-wide scope of the
// condition.
func TestStatusGitUnreachable_MessageNamesTriedRefs(t *testing.T) {
	root := setupGitRepoAt(t, "develop", []fixtureCommit{{title: "chore: seed"}})
	paths := writeUnreachableSpecs(t, root, "draft", "SPEC-URO-001")
	chdirForTest(t, root)

	report := runLintOnPaths(t, true, paths...)

	got := findingsByCode(report, "StatusGitUnreachable")
	if len(got) != 1 {
		t.Fatalf("StatusGitUnreachable findings = %d, want 1", len(got))
	}
	msg := got[0].Message
	for _, want := range []string{"main", "master"} {
		if !strings.Contains(msg, want) {
			t.Errorf("message %q does not name candidate ref %q", msg, want)
		}
	}
	if !strings.Contains(msg, "repository-wide") {
		t.Errorf("message %q does not state the repository-wide scope of the condition", msg)
	}
}

// AC-SLGB-003: the emission is capped at exactly one per Lint() run, not one
// per SPEC (spec.md §2.2 — the cause is repository-wide, and a per-SPEC flood
// would bury the real signal in the very CI state this card repairs).
func TestStatusGitUnreachable_EmittedOncePerRun(t *testing.T) {
	root := setupGitRepoAt(t, "develop", []fixtureCommit{{title: "chore: seed"}})
	ids := []string{
		"SPEC-URO-001", "SPEC-URO-002", "SPEC-URO-003", "SPEC-URO-004", "SPEC-URO-005",
		"SPEC-URO-006", "SPEC-URO-007", "SPEC-URO-008", "SPEC-URO-009", "SPEC-URO-010",
	}
	paths := writeUnreachableSpecs(t, root, "draft", ids...)
	chdirForTest(t, root)

	report := runLintOnPaths(t, true, paths...)

	if got := findingsByCode(report, "StatusGitUnreachable"); len(got) != 1 {
		t.Fatalf("StatusGitUnreachable findings = %d, want exactly 1 (10 non-terminal SPECs linted in one run)", len(got))
	}
}

// AC-SLGB-004a (shape ②): shallow clone with the base ref resolved and no
// in-window commit naming the SPEC — "no git history found" is an artifact of
// the truncated window, so it must surface as unreachable.
func TestStatusGitUnreachable_ShallowNoHistory(t *testing.T) {
	paths := setupShallowCloneFixture(t,
		[]fixtureCommit{{title: "chore: initial unrelated commit"}},
		"draft", "SPEC-URO-001")

	report := runLintOnPaths(t, true, paths...)

	got := findingsByCode(report, "StatusGitUnreachable")
	if len(got) != 1 {
		t.Fatalf("StatusGitUnreachable findings = %d, want 1 (report held %d findings total)", len(got), len(report.Findings))
	}
}

// AC-SLGB-004b (shape ③): shallow clone whose only in-window commit naming
// the SPEC is unclassifiable (chore(spec) sweep) — window exhaustion as a
// shallow artifact must surface as unreachable (the iter-1 D1 blind spot).
func TestStatusGitUnreachable_ShallowWindowExhausted(t *testing.T) {
	paths := setupShallowCloneFixture(t,
		[]fixtureCommit{{title: "chore(spec): sweep SPEC-URO-001 residual findings"}},
		"draft", "SPEC-URO-001")

	report := runLintOnPaths(t, true, paths...)

	got := findingsByCode(report, "StatusGitUnreachable")
	if len(got) != 1 {
		t.Fatalf("StatusGitUnreachable findings = %d, want 1 (report held %d findings total)", len(got), len(report.Findings))
	}
}

// AC-SLGB-005 (REQ-SLGB-004): in a FULL repository with a resolved local
// main, shapes ②/③ are normal-harmless (a SPEC may simply have no lifecycle
// commits) and must stay silent. The fixture status is non-terminal ([HARD]:
// terminal statuses return before git is ever queried, which would make a
// zero here vacuous). This AC is vacuously green pre-implementation and is
// closed via the shallow-guard mutation recorded in progress.md §E.2.
func TestStatusGitUnreachable_FullRepoStaysSilent(t *testing.T) {
	t.Run("non_terminal_status_draft", func(t *testing.T) {
		root := setupGitRepoAt(t, "main", []fixtureCommit{{title: "chore: initial unrelated commit"}})
		paths := writeUnreachableSpecs(t, root, "draft", "SPEC-URO-001")
		chdirForTest(t, root)

		report := runLintOnPaths(t, true, paths...)

		if got := findingsByCode(report, "StatusGitUnreachable"); len(got) != 0 {
			t.Fatalf("StatusGitUnreachable findings = %d, want 0 in a full repository (findings: %+v)", len(got), got)
		}
	})

	// No-false-positive-on-closed-SPECs guard (internal/spec/CLAUDE.md sibling
	// convention): a terminal-status SPEC is never flagged, including when the
	// repository is unobservable — Check returns before touching git. The pair
	// (this subtest silent, AC-SLGB-001 emitting in the same repo shape) is
	// what keeps the terminal exemption from hiding a general silence.
	t.Run("terminal_status_completed_stays_silent", func(t *testing.T) {
		root := setupGitRepoAt(t, "develop", []fixtureCommit{{title: "chore: seed"}})
		paths := writeUnreachableSpecs(t, root, "completed", "SPEC-URO-001")
		chdirForTest(t, root)

		report := runLintOnPaths(t, true, paths...)

		if got := findingsByCode(report, "StatusGitUnreachable"); len(got) != 0 {
			t.Fatalf("StatusGitUnreachable findings = %d, want 0 for terminal-status SPEC (findings: %+v)", len(got), got)
		}
	})
}

// AC-SLGB-006 (REQ-SLGB-005): the finding is Info severity and does not move
// the --strict exit status (Report.HasErrors stays false).
func TestStatusGitUnreachable_InfoSeverityKeepsStrictGreen(t *testing.T) {
	root := setupGitRepoAt(t, "develop", []fixtureCommit{{title: "chore: seed"}})
	paths := writeUnreachableSpecs(t, root, "draft", "SPEC-URO-001")
	chdirForTest(t, root)

	report := runLintOnPaths(t, true, paths...) // strict

	got := findingsByCode(report, "StatusGitUnreachable")
	if len(got) != 1 {
		t.Fatalf("StatusGitUnreachable findings = %d, want 1", len(got))
	}
	if got[0].Severity != SeverityInfo {
		t.Errorf("severity = %q, want %q", got[0].Severity, SeverityInfo)
	}
	if report.HasErrors() {
		t.Error("HasErrors() = true under --strict, want false (Info must not change the exit code)")
	}
}
