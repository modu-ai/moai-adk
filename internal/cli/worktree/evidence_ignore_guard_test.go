package worktree

// Guard for the reports/worktrees ignore matrix
// (SPEC-REPORTS-LIFECYCLE-001 REQ-RLC-008 / AC-RLC-003 / AC-RLC-004).
//
// The matrix pins, via `git check-ignore` behavior on a hermetic temp git
// repo seeded with a copy of the real ruleset:
//   - `.moai/reports/*` excludes every new file under .moai/reports/
//     (including the relocated reports/ evidence now at
//     .moai/reports/historical/ — REQ-RLC-002: no re-inclusion);
//   - `.moai/worktrees/` excludes card-tree content (hoisted evidence lives
//     under .moai/reports/worktrees/, covered by the same reports rule);
//   - the defense-in-depth line `.moai/reports/*.md` sits AFTER the
//     negation block so a future re-inclusion pattern can never win
//     (gitignore semantics: last matching rule wins — spec.md §A.2);
//   - tracked files are unaffected by the rules (the relocation keeps the
//     511-file set tracked: git status reports no deletions).
//
// The guard's red is observed on a known failing input: mutation subtests
// delete one rule from the temp copy and assert the evaluation flips to
// failure. A green whose swept set were empty would assert nothing — the
// README.md negative control keeps the matrix from passing vacuously.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// ignoreCase is one row of the check-ignore action matrix.
type ignoreCase struct {
	path       string // path to probe, relative to the temp repo root
	wantIgnore bool   // whether git check-ignore must report it ignored
}

// evidenceIgnoreCases is the pinned action matrix. The negative control
// (README.md) must stay outside every rule — without it a ruleset that
// ignores everything would pass.
var evidenceIgnoreCases = []ignoreCase{
	{path: "README.md", wantIgnore: false},
	{path: ".moai/reports/newfile.md", wantIgnore: true},
	{path: ".moai/reports/t1320/evidence.md", wantIgnore: true},
	{path: ".moai/reports/historical/newfile.md", wantIgnore: true},
	{path: ".moai/reports/worktrees/wt-slug/evidence.md", wantIgnore: true},
	{path: ".moai/worktrees/t999/evidence.md", wantIgnore: true},
}

// newTempIgnoreRepo copies srcIgnore into a fresh temp git repo and returns
// the repo dir. removedRules (verbatim lines) are stripped from the copy —
// the mutation axis for the observed-failure cells. injectedRules are
// inserted immediately BEFORE the `.moai/reports/*.md` defense line — the
// shape the defense line exists to defeat (spec.md §A.2: a future
// re-inclusion pattern; gitignore semantics — last matching rule wins).
func newTempIgnoreRepo(t *testing.T, srcIgnore string, removedRules []string, injectedRules []string) string {
	t.Helper()
	repo := t.TempDir()
	data, err := os.ReadFile(srcIgnore)
	if err != nil {
		t.Fatalf("read source .gitignore: %v", err)
	}
	lines := strings.Split(string(data), "\n")
	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimRight(line, "\r")
		if trimmed == ".moai/reports/*.md" {
			kept = append(kept, injectedRules...)
		}
		dropped := false
		for _, rule := range removedRules {
			if trimmed == rule {
				dropped = true
				break
			}
		}
		if !dropped {
			kept = append(kept, line)
		}
	}
	if err := os.WriteFile(filepath.Join(repo, ".gitignore"), []byte(strings.Join(kept, "\n")), 0o644); err != nil {
		t.Fatalf("write temp .gitignore: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(repo, ".moai", "reports", "historical"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	runGit(t, repo, "init", "-q")
	runGit(t, repo, "add", ".gitignore")
	return repo
}

// runGit runs a git command in dir and fails the test on error.
func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s in %s: %v\n%s", strings.Join(args, " "), dir, err, out)
	}
	return string(out)
}

// evaluateIgnoreMatrix runs `git check-ignore -v --non-matching` for every
// case in repo and returns the mismatch descriptions (empty = matrix holds).
//
// The verdict comes from the OUTPUT, not the exit code: on this git, `-v`
// exits 0 even when the deciding pattern is a negation (probe: negation-only
// match prints the `!` rule and exits 0; the same run without `-v` exits 1).
// With --non-matching every path prints a line — `<source>:<line>:<pattern>`
// when a pattern decided, `::` when nothing matched — so the decision is
// version-independent: ignored = matched pattern that does not start with '!'.
func evaluateIgnoreMatrix(t *testing.T, repo string) []string {
	t.Helper()
	var failures []string
	for _, tc := range evidenceIgnoreCases {
		cmd := exec.Command("git", "check-ignore", "-v", "--non-matching", tc.path)
		cmd.Dir = repo
		out, err := cmd.CombinedOutput()
		if err != nil && err.Error() != "exit status 1" {
			t.Fatalf("git check-ignore %s: %v\n%s", tc.path, err, out)
		}
		line := strings.TrimSpace(string(out))
		var ignored bool
		switch {
		case line == "":
			ignored = false // --non-matching should always print; treat as not ignored
		case strings.HasPrefix(line, "::"):
			ignored = false // no pattern matched
		default:
			// <source>:<line>:<pattern>\t<path> — a leading '!' on the
			// pattern means the path was re-included (not ignored).
			fields := strings.SplitN(line, ":", 3)
			pattern := ""
			if len(fields) == 3 {
				pattern = fields[2]
			}
			if idx := strings.IndexByte(pattern, '\t'); idx >= 0 {
				pattern = pattern[:idx]
			}
			ignored = pattern != "" && !strings.HasPrefix(pattern, "!")
		}
		if ignored != tc.wantIgnore {
			failures = append(failures,
				"check-ignore "+tc.path+": got ignored="+boolLabel(ignored)+", want "+boolLabel(tc.wantIgnore)+" (deciding: "+line+")")
		}
	}
	return failures
}

func boolLabel(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// repoRootFromPackageDir walks up from the package dir to the directory that
// holds the root .gitignore.
func repoRootFromPackageDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("abs package parent: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".gitignore")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repo root with .gitignore not found")
		}
		dir = parent
	}
}

// requireRule asserts srcIgnore contains each rule verbatim.
func requireRule(t *testing.T, content string, rule string, where string) {
	t.Helper()
	if !strings.Contains(content, rule+"\n") && !strings.HasSuffix(content, rule) {
		t.Errorf("%s: required ignore rule missing: %q", where, rule)
	}
}

// TestEvidenceIgnoreMatrixGuard pins the reports/worktrees ignore matrix
// (REQ-RLC-008). Presence axis: the real local and template-mirror rulesets
// carry every required rule, with the defense-in-depth .md line ordered after
// the negation block. Behavior axis: check-ignore on a hermetic copy of the
// local ruleset satisfies the full action matrix, and deleting any single
// required rule flips the matrix to failure (observed red).
func TestEvidenceIgnoreMatrixGuard(t *testing.T) {
	root := repoRootFromPackageDir(t)
	localIgnore := filepath.Join(root, ".gitignore")
	mirrorIgnore := filepath.Join(root, "internal", "template", "templates", ".gitignore")

	localRules := []string{
		".moai/reports/*",
		".moai/worktrees/",
		".moai/reports/*.md",
	}
	mirrorRules := []string{
		".moai/reports/*",
		".moai/worktrees/",
	}

	localContent, err := os.ReadFile(localIgnore)
	if err != nil {
		t.Fatalf("read local .gitignore: %v", err)
	}
	mirrorContent, err := os.ReadFile(mirrorIgnore)
	if err != nil {
		t.Fatalf("read template mirror .gitignore: %v", err)
	}
	for _, rule := range localRules {
		requireRule(t, string(localContent), rule, "local .gitignore")
	}
	for _, rule := range mirrorRules {
		requireRule(t, string(mirrorContent), rule, "template mirror .gitignore")
	}

	// Defense-in-depth ordering (spec.md §A.2): the .moai/reports/*.md
	// re-exclusion must sit after the negation block so a future re-inclusion
	// pattern can never beat it (last matching rule wins).
	defenseIdx := strings.Index(string(localContent), ".moai/reports/*.md")
	negationIdx := strings.Index(string(localContent), "!.moai/reports/plan-audit/")
	if defenseIdx < 0 || negationIdx < 0 {
		t.Fatalf("ordering probe lines not found: defense=%d negation=%d", defenseIdx, negationIdx)
	}
	if defenseIdx < negationIdx {
		t.Errorf("defense-in-depth .moai/reports/*.md must come after the negation block (defense at %d, negation at %d)", defenseIdx, negationIdx)
	}

	// Behavior axis on a hermetic copy of the local ruleset.
	repo := newTempIgnoreRepo(t, localIgnore, nil, nil)
	if failures := evaluateIgnoreMatrix(t, repo); len(failures) > 0 {
		t.Errorf("ignore matrix violated on the current ruleset (%d case(s)):\n%s",
			len(failures), strings.Join(failures, "\n"))
	}

	// Tracked continuity (REQ-RLC-001/002 mechanism): a tracked file under
	// .moai/reports/historical/ is reported by check-ignore as ignored-path
	// material, yet stays tracked and shows no deletion in status.
	kept := filepath.Join(repo, ".moai", "reports", "historical", "tracked-kept.md")
	if err := os.WriteFile(kept, []byte("tracked evidence\n"), 0o644); err != nil {
		t.Fatalf("write tracked fixture: %v", err)
	}
	runGit(t, repo, "add", "-f", ".moai/reports/historical/tracked-kept.md")
	runGit(t, repo, "-c", "user.email=guard@test", "-c", "user.name=guard", "commit", "-q", "-m", "fixture")
	if status := strings.TrimSpace(runGit(t, repo, "status", "--porcelain")); status != "" {
		t.Errorf("tracked file under .moai/reports/historical/ polluted status:\n%s", status)
	}
	lsOut := runGit(t, repo, "ls-files", ".moai/reports/historical/")
	if !strings.Contains(lsOut, "tracked-kept.md") {
		t.Errorf("tracked file vanished from the index: ls-files=%q", lsOut)
	}
}

// TestEvidenceIgnoreMatrixGuard_MutationObservedRed is the completion cell
// for the guard: each required rule, when deleted from an otherwise identical
// copy, must flip the matrix to failure. If a mutant ruleset still passes,
// the guard cannot fail and is unfinished.
func TestEvidenceIgnoreMatrixGuard_MutationObservedRed(t *testing.T) {
	root := repoRootFromPackageDir(t)
	localIgnore := filepath.Join(root, ".gitignore")

	mutations := []struct {
		name     string
		rule     string
		injected []string
	}{
		{"delete reports rule", ".moai/reports/*", nil},
		{"delete worktrees rule", ".moai/worktrees/", nil},
		// The .md defense line is only load-bearing once a re-inclusion
		// pattern exists (spec.md §A.2) — simulate the future regression it
		// exists to defeat: with the defense line the matrix holds; without
		// it the re-inclusion wins and a .moai/reports/*.md file leaks.
		{
			name:     "md defense line absent under re-inclusion pressure",
			rule:     ".moai/reports/*.md",
			injected: []string{"!.moai/reports/**/*.md"},
		},
	}
	for _, mut := range mutations {
		t.Run(mut.name, func(t *testing.T) {
			repo := newTempIgnoreRepo(t, localIgnore, []string{mut.rule}, mut.injected)
			failures := evaluateIgnoreMatrix(t, repo)
			if len(failures) == 0 {
				t.Errorf("mutation %q did NOT flip the matrix — guard cannot detect its deletion", mut.rule)
			}
		})
	}

	// Positive control for the pressure scenario: the SAME injected
	// re-inclusion pattern WITH the defense line present must keep the
	// matrix green — the defense line, not luck, is what holds.
	repo := newTempIgnoreRepo(t, localIgnore, nil, []string{"!.moai/reports/**/*.md"})
	if failures := evaluateIgnoreMatrix(t, repo); len(failures) > 0 {
		t.Errorf("defense line failed to defeat the injected re-inclusion:\n%s", strings.Join(failures, "\n"))
	}
}
