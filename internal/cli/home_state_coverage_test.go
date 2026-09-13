package cli

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

func gitForCoverageTest(t *testing.T, root string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
}

func committedCoverageRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	gitForCoverageTest(t, root, "init", "-q")
	gitForCoverageTest(t, root, "config", "user.email", "test@example.com")
	gitForCoverageTest(t, root, "config", "user.name", "Test")
	if err := os.MkdirAll(filepath.Join(root, "internal", "x"), 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "internal", "x", "a.go")
	if err := os.WriteFile(path, []byte("package x\nfunc A() int { return 1 }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	gitForCoverageTest(t, root, "add", "internal/x/a.go")
	gitForCoverageTest(t, root, "commit", "-qm", "base")
	if err := os.WriteFile(path, []byte("package x\nfunc A() int { return 2 }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	gitForCoverageTest(t, root, "add", "internal/x/a.go")
	gitForCoverageTest(t, root, "commit", "-qm", "feat(state): add guarded home-state rollout (t592)")
	return root
}

func TestCommittedCoverageChangeSetWorksInCleanRepositoryAndAfterUnrelatedCommit(t *testing.T) {
	root := committedCoverageRepo(t)
	changeSet, err := resolveHomeStateCoverageChangeSet(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(changeSet.Native) != 1 || changeSet.Native[0] != "internal/x/a.go" || len(changeSet.Ranges["internal/x/a.go"]) == 0 {
		t.Fatalf("changeSet=%+v", changeSet)
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("unrelated\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	gitForCoverageTest(t, root, "add", "README.md")
	gitForCoverageTest(t, root, "commit", "-qm", "docs: unrelated")
	if _, err := resolveHomeStateCoverageChangeSet(root); err != nil {
		t.Fatalf("unrelated later commit rejected: %v", err)
	}
}

func TestCommittedCoverageChangeSetWorksAfterMergeCommit(t *testing.T) {
	root := committedCoverageRepo(t)
	gitForCoverageTest(t, root, "checkout", "-qb", "docs")
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("docs branch\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	gitForCoverageTest(t, root, "add", "README.md")
	gitForCoverageTest(t, root, "commit", "-qm", "docs: branch")
	gitForCoverageTest(t, root, "checkout", "-q", "-")
	if err := os.WriteFile(filepath.Join(root, "NOTICE"), []byte("main branch\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	gitForCoverageTest(t, root, "add", "NOTICE")
	gitForCoverageTest(t, root, "commit", "-qm", "chore: main")
	gitForCoverageTest(t, root, "merge", "--no-ff", "-qm", "chore: merge docs", "docs")
	if _, err := resolveHomeStateCoverageChangeSet(root); err != nil {
		t.Fatalf("merge preserving audited blobs rejected: %v", err)
	}
}

func TestCommittedCoverageChangeSetWorksCleanAfterRemediationCommit(t *testing.T) {
	root := committedCoverageRepo(t)
	path := filepath.Join(root, "internal", "x", "a.go")
	if err := os.WriteFile(path, []byte("package x\nfunc A() int { return 5 }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	gitForCoverageTest(t, root, "add", "internal/x/a.go")
	gitForCoverageTest(t, root, "commit", "-qm", homeStateCoverageRemediationCommitSubject)
	changeSet, err := resolveHomeStateCoverageChangeSet(root)
	if err != nil {
		t.Fatal(err)
	}
	head, err := gitCoverageOutput(root, "rev-parse", "HEAD")
	if err != nil || changeSet.Tip != head || len(changeSet.Ranges["internal/x/a.go"]) == 0 {
		t.Fatalf("changeSet=%+v head=%s err=%v", changeSet, head, err)
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("later\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	gitForCoverageTest(t, root, "add", "README.md")
	gitForCoverageTest(t, root, "commit", "-qm", "docs: after remediation")
	if _, err := resolveHomeStateCoverageChangeSet(root); err != nil {
		t.Fatalf("later unrelated commit rejected: %v", err)
	}
}

func TestCommittedCoverageChangeSetExcludesInterveningMergedProduction(t *testing.T) {
	root := committedCoverageRepo(t)
	gitForCoverageTest(t, root, "checkout", "-qb", "unrelated-production")
	codexPath := filepath.Join(root, "internal", "codexwiring", "configtoml.go")
	if err := os.MkdirAll(filepath.Dir(codexPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(codexPath, []byte("package codexwiring\nfunc Config() string { return \"unrelated\" }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	gitForCoverageTest(t, root, "add", "internal/codexwiring/configtoml.go")
	gitForCoverageTest(t, root, "commit", "-qm", "feat(codex): unrelated production")
	gitForCoverageTest(t, root, "checkout", "-q", "-")
	if err := os.WriteFile(filepath.Join(root, "NOTICE"), []byte("main advance\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	gitForCoverageTest(t, root, "add", "NOTICE")
	gitForCoverageTest(t, root, "commit", "-qm", "chore: advance main")
	gitForCoverageTest(t, root, "merge", "--no-ff", "-qm", "chore: merge unrelated production", "unrelated-production")
	a := filepath.Join(root, "internal", "x", "a.go")
	if err := os.WriteFile(a, []byte("package x\nfunc A() int { return 5 }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	gitForCoverageTest(t, root, "add", "internal/x/a.go")
	gitForCoverageTest(t, root, "commit", "-qm", homeStateCoverageRemediationCommitSubject)
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("later docs\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	gitForCoverageTest(t, root, "add", "README.md")
	gitForCoverageTest(t, root, "commit", "-qm", "docs: after remediation")

	changeSet, err := resolveHomeStateCoverageChangeSet(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := changeSet.Ranges["internal/codexwiring/configtoml.go"]; ok {
		t.Fatalf("intervening merged production included: %+v", changeSet)
	}
	if len(changeSet.Ranges["internal/x/a.go"]) == 0 {
		t.Fatalf("audited production missing: %+v", changeSet)
	}
}

func TestCommittedCoverageChangeSetSupportsVersionedRemediationChain(t *testing.T) {
	root := committedCoverageRepo(t)
	a := filepath.Join(root, "internal", "x", "a.go")
	for _, step := range []struct{ value, subject string }{
		{"5", homeStateCoverageRemediationCommitSubject},
		{"6", homeStateCoverageDeltaCommitSubject},
		{"7", homeStateCoverageCertificationSubject},
	} {
		if err := os.WriteFile(a, []byte("package x\nfunc A() int { return "+step.value+" }\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		gitForCoverageTest(t, root, "add", "internal/x/a.go")
		gitForCoverageTest(t, root, "commit", "-qm", step.subject)
	}
	tip, err := gitCoverageOutput(root, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("later docs\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	gitForCoverageTest(t, root, "add", "README.md")
	gitForCoverageTest(t, root, "commit", "-qm", "docs: after delta remediation")
	changeSet, err := resolveHomeStateCoverageChangeSet(root)
	if err != nil || changeSet.Tip != tip || len(changeSet.Ranges["internal/x/a.go"]) != 4 {
		t.Fatalf("changeSet=%+v tip=%s err=%v", changeSet, tip, err)
	}
}

func TestCommittedCoverageChangeSetMergesDirtyProductionDiff(t *testing.T) {
	root := committedCoverageRepo(t)
	dirty := filepath.Join(root, "internal", "x", "dirty.go")
	if err := os.WriteFile(dirty, []byte("package x\nfunc Dirty() int { return 1 }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	changeSet, err := resolveHomeStateCoverageChangeSet(root)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(changeSet.Native, "\n"), "internal/x/dirty.go") || len(changeSet.Ranges["internal/x/dirty.go"]) == 0 {
		t.Fatalf("dirty production diff not merged: %+v", changeSet)
	}
	tracked := filepath.Join(root, "internal", "x", "a.go")
	if err := os.WriteFile(tracked, []byte("package x\nfunc A() int { return 4 }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	changeSet, err = resolveHomeStateCoverageChangeSet(root)
	if err != nil || len(changeSet.Ranges["internal/x/a.go"]) < 2 {
		t.Fatalf("tracked dirty production diff not merged: %+v err=%v", changeSet, err)
	}
}

func TestCommittedCoverageChangeSetRejectsStaleOrAmbiguousEvidence(t *testing.T) {
	t.Run("covered path changed after audited tip", func(t *testing.T) {
		root := committedCoverageRepo(t)
		path := filepath.Join(root, "internal", "x", "a.go")
		if err := os.WriteFile(path, []byte("package x\nfunc A() int { return 3 }\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		gitForCoverageTest(t, root, "add", "internal/x/a.go")
		gitForCoverageTest(t, root, "commit", "-qm", "fix: mutate audited path")
		if _, err := resolveHomeStateCoverageChangeSet(root); err == nil {
			t.Fatal("stale audited production blob accepted")
		}
	})
	t.Run("duplicate audit marker", func(t *testing.T) {
		root := committedCoverageRepo(t)
		if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("duplicate\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		gitForCoverageTest(t, root, "add", "README.md")
		gitForCoverageTest(t, root, "commit", "-qm", "feat(state): add guarded home-state rollout (t592)")
		if _, err := resolveHomeStateCoverageChangeSet(root); err == nil {
			t.Fatal("ambiguous audit evidence accepted")
		}
	})
	t.Run("audit marker not reachable from head", func(t *testing.T) {
		root := committedCoverageRepo(t)
		gitForCoverageTest(t, root, "checkout", "-qb", "unrelated", "HEAD^1")
		if _, err := resolveHomeStateCoverageChangeSet(root); err == nil || !strings.Contains(err.Error(), "evidence commit not found") {
			t.Fatalf("unreachable audit evidence accepted: %v", err)
		}
	})
	t.Run("remediation marker is not a descendant", func(t *testing.T) {
		root := committedCoverageRepo(t)
		gitForCoverageTest(t, root, "checkout", "-qb", "forged-remediation", "HEAD^1")
		if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("forged\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		gitForCoverageTest(t, root, "add", "README.md")
		gitForCoverageTest(t, root, "commit", "-qm", homeStateCoverageRemediationCommitSubject)
		gitForCoverageTest(t, root, "checkout", "-q", "-")
		gitForCoverageTest(t, root, "merge", "--no-ff", "-qm", "merge forged remediation", "forged-remediation")
		if _, err := resolveHomeStateCoverageChangeSet(root); err == nil || !strings.Contains(err.Error(), "does not descend") {
			t.Fatalf("non-descendant remediation evidence accepted: %v", err)
		}
	})
	t.Run("duplicate remediation marker", func(t *testing.T) {
		root := committedCoverageRepo(t)
		for i := 0; i < 2; i++ {
			path := filepath.Join(root, "README"+strconv.Itoa(i)+".md")
			if err := os.WriteFile(path, []byte("remediation\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			gitForCoverageTest(t, root, "add", filepath.Base(path))
			gitForCoverageTest(t, root, "commit", "-qm", homeStateCoverageRemediationCommitSubject)
		}
		if _, err := resolveHomeStateCoverageChangeSet(root); err == nil || !strings.Contains(err.Error(), "ambiguous home-state coverage evidence marker") {
			t.Fatalf("duplicate remediation evidence accepted: %v", err)
		}
	})
	t.Run("versioned marker skips predecessor", func(t *testing.T) {
		root := committedCoverageRepo(t)
		if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("skip predecessor\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		gitForCoverageTest(t, root, "add", "README.md")
		gitForCoverageTest(t, root, "commit", "-qm", homeStateCoverageDeltaCommitSubject)
		if _, err := resolveHomeStateCoverageChangeSet(root); err == nil || !strings.Contains(err.Error(), "missing predecessor") {
			t.Fatalf("gapped evidence chain accepted: %v", err)
		}
	})
	t.Run("certification marker skips predecessor", func(t *testing.T) {
		root := committedCoverageRepo(t)
		path := filepath.Join(root, "internal", "x", "a.go")
		if err := os.WriteFile(path, []byte("package x\nfunc A() int { return 5 }\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		gitForCoverageTest(t, root, "add", "internal/x/a.go")
		gitForCoverageTest(t, root, "commit", "-qm", homeStateCoverageRemediationCommitSubject)
		if err := os.WriteFile(path, []byte("package x\nfunc A() int { return 7 }\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		gitForCoverageTest(t, root, "add", "internal/x/a.go")
		gitForCoverageTest(t, root, "commit", "-qm", homeStateCoverageCertificationSubject)
		if _, err := resolveHomeStateCoverageChangeSet(root); err == nil || !strings.Contains(err.Error(), "missing predecessor") {
			t.Fatalf("gapped certification evidence chain accepted: %v", err)
		}
	})
	t.Run("duplicate certification marker", func(t *testing.T) {
		root := committedCoverageRepo(t)
		path := filepath.Join(root, "internal", "x", "a.go")
		for _, step := range []struct{ value, subject string }{
			{"5", homeStateCoverageRemediationCommitSubject},
			{"6", homeStateCoverageDeltaCommitSubject},
			{"7", homeStateCoverageCertificationSubject},
			{"8", homeStateCoverageCertificationSubject},
		} {
			if err := os.WriteFile(path, []byte("package x\nfunc A() int { return "+step.value+" }\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			gitForCoverageTest(t, root, "add", "internal/x/a.go")
			gitForCoverageTest(t, root, "commit", "-qm", step.subject)
		}
		if _, err := resolveHomeStateCoverageChangeSet(root); err == nil || !strings.Contains(err.Error(), "ambiguous home-state coverage evidence marker") {
			t.Fatalf("duplicate certification evidence accepted: %v", err)
		}
	})
}

func TestCommittedCoverageChangeSetRejectsInvalidGitAndMalformedEvidence(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing")
	if _, err := resolveHomeStateCoverageChangeSet(missing); err == nil {
		t.Fatal("invalid git repository accepted")
	}
	if _, err := changedProductionLineRanges(missing, nil); err == nil {
		t.Fatal("invalid git repository accepted by line-range resolver")
	}
	t.Run("marker has no base parent", func(t *testing.T) {
		root := t.TempDir()
		gitForCoverageTest(t, root, "init", "-q")
		gitForCoverageTest(t, root, "config", "user.email", "test@example.com")
		gitForCoverageTest(t, root, "config", "user.name", "Test")
		if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("root marker\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		gitForCoverageTest(t, root, "add", "README.md")
		gitForCoverageTest(t, root, "commit", "-qm", homeStateCoverageCommitSubject)
		if _, err := resolveHomeStateCoverageChangeSet(root); err == nil || !strings.Contains(err.Error(), "coverage base") {
			t.Fatalf("parentless evidence accepted: %v", err)
		}
	})
	if _, err := productionFilesFromNameStatus("malformed"); err == nil {
		t.Fatal("malformed name-status accepted")
	}
	if _, err := productionFilesFromNameStatus("D\tinternal/x/gone.go"); err == nil {
		t.Fatal("production deletion accepted")
	}
	files, err := productionFilesFromNameStatus("D\tREADME.md\nM\tinternal/x/live.go\nM\tinternal/x/live_test.go")
	if err != nil || len(files) != 1 || !files["internal/x/live.go"] {
		t.Fatalf("files=%v err=%v", files, err)
	}
	root := committedCoverageRepo(t)
	ranges, err := changedProductionLineRanges(root, nil)
	if err != nil || len(ranges) != 0 {
		t.Fatalf("unwanted ranges not filtered: %v err=%v", ranges, err)
	}
	if _, err := changedProductionLineRanges(root, []string{"internal/x/missing.go"}); err == nil || !strings.Contains(err.Error(), "missing from diff") {
		t.Fatalf("requested missing file accepted: %v", err)
	}
	t.Run("unreadable untracked production", func(t *testing.T) {
		root := committedCoverageRepo(t)
		path := filepath.Join(root, "internal", "x", "dangling.go")
		if err := os.Symlink("missing-target", path); err != nil {
			t.Fatal(err)
		}
		if _, err := resolveHomeStateCoverageChangeSet(root); err == nil {
			t.Fatal("unreadable untracked production accepted")
		}
	})
}

func TestParseChangedSurfaceCoverageRejectsMissingZeroAndTamperedProfiles(t *testing.T) {
	files := []string{"internal/cli/a.go", "internal/homestate/b.go"}
	write := func(name, body string) string {
		t.Helper()
		path := filepath.Join(t.TempDir(), name)
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	var valid strings.Builder
	valid.WriteString("mode: set\n")
	for i, file := range files {
		count := 1
		if i == len(files)-1 {
			count = 0
		}
		valid.WriteString("github.com/modu-ai/moai-adk/" + file + ":1.1,1.2 10 " + strconv.Itoa(count) + "\n")
	}
	coverage, err := parseChangedSurfaceCoverage(write("valid.out", valid.String()), files)
	want := 100 * float64(len(files)-1) / float64(len(files))
	if err != nil || coverage != want {
		t.Fatalf("coverage=%.1f err=%v", coverage, err)
	}
	if _, err := parseChangedSurfaceCoverage(write("missing.out", "mode: set\n"), files); err == nil {
		t.Fatal("missing/zero profile accepted")
	}
	tampered := valid.String() + "github.com/modu-ai/moai-adk/" + files[0] + ":1.1,1.2 11 1\n"
	if _, err := parseChangedSurfaceCoverage(write("tampered.out", tampered), files); err == nil {
		t.Fatal("tampered duplicate accepted")
	}
	if _, err := parseChangedSurfaceCoverage(write("malformed.out", "not-a-profile\n"), files); err == nil {
		t.Fatal("malformed profile accepted")
	}
	for name, body := range map[string]string{
		"short-row.out":   "mode: set\nmissing-fields\n",
		"no-location.out": "mode: set\nno-colon 1 1\n",
		"bad-count.out":   "mode: set\ngithub.com/modu-ai/moai-adk/" + files[0] + ":1.1,1.2 nope 1\n",
		"negative.out":    "mode: set\ngithub.com/modu-ai/moai-adk/" + files[0] + ":1.1,1.2 1 -1\n",
	} {
		if _, err := parseChangedSurfaceCoverage(write(name, body), files); err == nil {
			t.Fatalf("%s accepted", name)
		}
	}
	if _, err := parseChangedSurfaceCoverage(write("zero.out", "mode: set\n"), nil); err == nil {
		t.Fatal("zero-statement profile accepted")
	}
	if _, err := parseChangedSurfaceCoverage(write("oversized.out", "mode: set\n"+strings.Repeat("x", 70*1024)), files); err == nil {
		t.Fatal("scanner failure accepted")
	}
}

func TestHomeStateChangedSurfaceCoverageConsumesFreshProfile(t *testing.T) {
	// A fixture repository keeps this test independent of the live history:
	// against the live tree it fails whenever a later commit touches an
	// audited file, which is the gate refusing rather than the measurement
	// being wrong.
	repo := committedCoverageRepo(t)
	files, _, err := changedProductionFiles(repo)
	if err != nil {
		t.Fatal(err)
	}
	fresh := func(_ context.Context, _ string, path string) error {
		var profile strings.Builder
		profile.WriteString("mode: set\n")
		ranges, rangeErr := changedProductionLineRanges(repo, files)
		if rangeErr != nil {
			return rangeErr
		}
		for _, file := range files {
			line := 1
			if len(ranges[file]) > 0 {
				line = ranges[file][0].Start
			}
			profile.WriteString("github.com/modu-ai/moai-adk/" + file + ":" + strconv.Itoa(line) + ".1," + strconv.Itoa(line) + ".2 1 1\n")
		}
		return os.WriteFile(path, []byte(profile.String()), 0o600)
	}
	coverage, err := measureChangedSurfaceCoverageWith(t.Context(), repo, fresh)
	if err != nil || coverage != 100 {
		t.Fatalf("coverage=%.1f err=%v", coverage, err)
	}
	// The live pre-apply gate reaches the audited-blob freeze through this
	// measurement, so a post-tip change to an audited file must still refuse.
	audited := filepath.Join(repo, "internal", "x", "a.go")
	if err := os.WriteFile(audited, []byte("package x\nfunc A() int { return 3 }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	gitForCoverageTest(t, repo, "add", "internal/x/a.go")
	gitForCoverageTest(t, repo, "commit", "-qm", "fix: mutate audited path")
	if _, err := measureChangedSurfaceCoverageWith(t.Context(), repo, fresh); err == nil || !strings.Contains(err.Error(), "audited production file changed after coverage tip: internal/x/a.go") {
		t.Fatalf("post-tip audited change accepted: %v", err)
	}
	if _, err := measureChangedSurfaceCoverageWith(t.Context(), t.TempDir(), func(context.Context, string, string) error { return context.Canceled }); err == nil {
		t.Fatal("coverage runner error accepted")
	}
	if _, err := measureChangedSurfaceCoverage(t.Context(), filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("missing repository executed coverage")
	}
	blockedTemp := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(blockedTemp, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TMPDIR", blockedTemp)
	if _, err := measureChangedSurfaceCoverageWith(t.Context(), t.TempDir(), func(context.Context, string, string) error { return nil }); err == nil {
		t.Fatal("unavailable temporary storage accepted")
	}
}

func TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite(t *testing.T) {
	if os.Getenv("MOAI_HOME_STATE_COVERAGE_CHILD") == "1" {
		return
	}
	// Opt-in: this runs the focused suite against the live repository and
	// resolves the audited evidence chain at HEAD, so it fails whenever an
	// audited production file changed after the last certification marker,
	// until the rollout is re-certified.
	if os.Getenv(config.EnvTestHomeStateLiveCoverage) != "1" {
		t.Skipf("set %s=1 to measure changed-surface coverage against the live repository", config.EnvTestHomeStateLiveCoverage)
	}
	t.Setenv("MOAI_HOME_STATE_COVERAGE_CHILD", "1")
	repo, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	result, err := measureChangedSurfaceCoverageResultWith(t.Context(), repo, runChangedSurfaceCoverageSuite)
	if err != nil || result.Percent <= 0 || result.Percent > 100 {
		t.Fatalf("coverage=%+v err=%v", result, err)
	}
	t.Logf("auto-diff changed production coverage: %d/%d = %.3f%%", result.Covered, result.Total, result.Percent)
}

func TestCoverageRunnerSetsRecursionGuardOnEveryChild(t *testing.T) {
	bin := t.TempDir()
	logPath := filepath.Join(bin, "guards.log")
	fakeGo := filepath.Join(bin, "go")
	script := `#!/bin/sh
profile=
for arg in "$@"; do
  case "$arg" in
    -coverprofile=*) profile=${arg#*=} ;;
  esac
done
printf '%s\n' "${MOAI_HOME_STATE_COVERAGE_CHILD:-missing}" >> "$MOAI_GUARD_LOG"
if [ "${MOAI_HOME_STATE_COVERAGE_CHILD:-}" != "1" ]; then
  echo "coverage child recursion guard missing" >&2
  exit 42
fi
printf 'mode: set\n' > "$profile"
`
	if err := os.WriteFile(fakeGo, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("MOAI_GUARD_LOG", logPath)
	t.Setenv("MOAI_HOME_STATE_COVERAGE_CHILD", "")
	profile := filepath.Join(t.TempDir(), "merged.out")
	if err := runChangedSurfaceCoverageSuite(t.Context(), t.TempDir(), profile); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	guards := strings.Fields(string(raw))
	if len(guards) != 5 {
		t.Fatalf("child invocations=%d guards=%q", len(guards), raw)
	}
	for i, guard := range guards {
		if guard != "1" {
			t.Fatalf("child %d recursion guard=%q", i, guard)
		}
	}
}

func TestChangedProductionFilesDerivesCurrentHeadDiffAndPlatformDisposition(t *testing.T) {
	root := committedCoverageRepo(t)
	crossCompiled := "only_windows.go"
	if runtime.GOOS == "windows" {
		crossCompiled = "only_unix.go"
	}
	for rel, body := range map[string]string{
		"internal/y/b.go":             "package y\nfunc B() int { return 1 }\n",
		"internal/x/" + crossCompiled: "package x\nfunc P() int { return 1 }\n",
		"internal/x/a_test.go":        "package x\n",
		"cmd/tool/main.go":            "package main\nfunc main() {}\n",
	} {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	gitForCoverageTest(t, root, "add", "internal", "cmd")
	gitForCoverageTest(t, root, "commit", "-qm", homeStateCoverageRemediationCommitSubject)
	native, disposition, err := changedProductionFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := strings.Join(native, ","), "internal/x/a.go,internal/y/b.go"; got != want {
		t.Errorf("native=%q, want %q", got, want)
	}
	if got, want := strings.Join(disposition, ","), "internal/x/"+crossCompiled+":cross-compile"; got != want {
		t.Errorf("disposition=%q, want %q", got, want)
	}
}

func TestParseChangedUnifiedZeroDiffHandlesRenameAndRejectsMalformedInput(t *testing.T) {
	ranges, err := parseUnifiedZeroDiff("diff --git a/old.go b/new.go\nrename from old.go\nrename to new.go\n--- a/old.go\n+++ b/new.go\n@@ -2 +2,2 @@\n-old\n+new\n+line\n")
	if err != nil || len(ranges["new.go"]) != 1 || ranges["new.go"][0] != (changedLineRange{Start: 2, End: 3}) {
		t.Fatalf("ranges=%v err=%v", ranges, err)
	}
	if _, err := parseUnifiedZeroDiff("--- a/gone.go\n+++ /dev/null\n@@ -1 +0,0 @@\n-x\n"); err == nil {
		t.Fatal("deletion accepted")
	}
	if _, err := parseUnifiedZeroDiff("+++ b/a.go\n@@ broken @@\n"); err == nil {
		t.Fatal("malformed hunk accepted")
	}
	for name, diff := range map[string]string{
		"target":         "+++ a.go\n",
		"no-target":      "@@ -1 +1 @@\n+x\n",
		"missing-end":    "+++ b/a.go\n@@ -1 +1\n+x\n",
		"bad-start":      "+++ b/a.go\n@@ -1 +x @@\n+x\n",
		"bad-count":      "+++ b/a.go\n@@ -1 +1,x @@\n+x\n",
		"extra-count":    "+++ b/a.go\n@@ -1 +1,2,3 @@\n+x\n",
		"negative-count": "+++ b/a.go\n@@ -1 +1,-1 @@\n+x\n",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := parseUnifiedZeroDiff(diff); err == nil {
				t.Fatal("malformed diff accepted")
			}
		})
	}
}

func TestParseChangedLineCoverageCountsOnlyChangedExecutableLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cover.out")
	body := "mode: set\nmodule/internal/cli/a.go:1.1,1.2 10 1\nmodule/internal/cli/a.go:5.1,5.2 3 0\nmodule/internal/cli/a.go:9.1,9.2 7 1\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := parseChangedLineCoverage(path, map[string][]changedLineRange{"internal/cli/a.go": {{Start: 5, End: 5}, {Start: 9, End: 9}}})
	if err != nil || result.Covered != 7 || result.Total != 10 || result.Percent != 70 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	if _, err := parseChangedLineCoverage(path, map[string][]changedLineRange{"internal/cli/missing.go": {{Start: 1, End: 1}}}); err == nil {
		t.Fatal("omitted changed file accepted")
	}
}

func TestChangedProductionLineRangesTracksModifiedRenamedAndUntracked(t *testing.T) {
	root := t.TempDir()
	gitForCoverageTest(t, root, "init", "-q")
	gitForCoverageTest(t, root, "config", "user.email", "test@example.com")
	gitForCoverageTest(t, root, "config", "user.name", "Test")
	if err := os.MkdirAll(filepath.Join(root, "internal", "x"), 0o700); err != nil {
		t.Fatal(err)
	}
	a := filepath.Join(root, "internal", "x", "a.go")
	base := "package x\nfunc A() int { return 1 }\nfunc C() int { return 3 }\nfunc D() int { return 4 }\nfunc E() int { return 5 }\n"
	if err := os.WriteFile(a, []byte(base), 0o600); err != nil {
		t.Fatal(err)
	}
	gitForCoverageTest(t, root, "add", "internal/x/a.go")
	gitForCoverageTest(t, root, "commit", "-qm", "base")
	base = strings.Replace(base, "return 1", "return 2", 1)
	if err := os.WriteFile(a, []byte(base), 0o600); err != nil {
		t.Fatal(err)
	}
	gitForCoverageTest(t, root, "add", "internal/x/a.go")
	gitForCoverageTest(t, root, "commit", "-qm", homeStateCoverageCommitSubject)
	gitForCoverageTest(t, root, "mv", "internal/x/a.go", "internal/x/b.go")
	b := filepath.Join(root, "internal", "x", "b.go")
	if err := os.WriteFile(b, []byte(base+"func B() int { return 6 }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	w := filepath.Join(root, "internal", "x", "only_windows.go")
	if err := os.WriteFile(w, []byte("package x\nfunc W() int { return 4 }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	native, disposition, err := changedProductionFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(disposition, "\n"), "only_windows.go:cross-compile") {
		t.Fatalf("disposition=%v", disposition)
	}
	ranges, err := changedProductionLineRanges(root, native)
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges["internal/x/b.go"]) == 0 {
		t.Fatalf("rename target ranges=%v", ranges)
	}
}

func TestChangedProductionFilesRejectsDeletion(t *testing.T) {
	root := t.TempDir()
	gitForCoverageTest(t, root, "init", "-q")
	gitForCoverageTest(t, root, "config", "user.email", "test@example.com")
	gitForCoverageTest(t, root, "config", "user.name", "Test")
	if err := os.MkdirAll(filepath.Join(root, "internal", "x"), 0o700); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(root, "internal", "x", "gone.go")
	if err := os.WriteFile(p, []byte("package x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	gitForCoverageTest(t, root, "add", "internal/x/gone.go")
	gitForCoverageTest(t, root, "commit", "-qm", "base")
	if err := os.WriteFile(p, []byte("package x\nfunc Present() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	gitForCoverageTest(t, root, "add", "internal/x/gone.go")
	gitForCoverageTest(t, root, "commit", "-qm", homeStateCoverageCommitSubject)
	if err := os.Remove(p); err != nil {
		t.Fatal(err)
	}
	if _, _, err := changedProductionFiles(root); err == nil || !strings.Contains(err.Error(), "deleted production coverage target") {
		t.Fatal("deleted production file accepted")
	}
}

// resolveRepeatedly runs the change-set resolver the given number of times and
// counts each distinct error message, so a caller can see whether the file a
// refusal names depends on map iteration order.
func resolveRepeatedly(t *testing.T, root string, runs int) map[string]int {
	t.Helper()
	seen := map[string]int{}
	for range runs {
		_, err := resolveHomeStateCoverageChangeSet(root)
		if err == nil {
			t.Fatal("change set accepted")
		}
		seen[err.Error()]++
	}
	return seen
}

func TestCommittedCoverageChangeSetNamesEveryPostTipChangeDeterministically(t *testing.T) {
	root := committedCoverageRepo(t)
	b := filepath.Join(root, "internal", "y", "b.go")
	if err := os.MkdirAll(filepath.Dir(b), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte("package y\nfunc B() int { return 1 }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	gitForCoverageTest(t, root, "add", "internal/y/b.go")
	gitForCoverageTest(t, root, "commit", "-qm", homeStateCoverageRemediationCommitSubject)
	for rel, body := range map[string]string{
		"internal/x/a.go": "package x\nfunc A() int { return 3 }\n",
		"internal/y/b.go": "package y\nfunc B() int { return 2 }\n",
	} {
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(rel)), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	gitForCoverageTest(t, root, "add", "internal")
	gitForCoverageTest(t, root, "commit", "-qm", "fix: mutate two audited paths")
	want := "audited production file changed after coverage tip: internal/x/a.go, internal/y/b.go"
	if seen := resolveRepeatedly(t, root, 20); len(seen) != 1 || seen[want] != 20 {
		t.Fatalf("messages across 20 runs = %v, want only %q", seen, want)
	}
}

func TestCommittedCoverageChangeSetNamesEveryFileMissingFromDiffDeterministically(t *testing.T) {
	root := committedCoverageRepo(t)
	for _, name := range []string{"b.go", "c.go"} {
		if err := os.WriteFile(filepath.Join(root, "internal", "x", name), []byte("package x\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	gitForCoverageTest(t, root, "add", "internal/x/b.go", "internal/x/c.go")
	gitForCoverageTest(t, root, "commit", "-qm", "chore: add unaudited files")
	// A mode-only change lists both files in name-status but produces no hunk,
	// so neither reaches the changed-line ranges.
	gitForCoverageTest(t, root, "update-index", "--chmod=+x", "internal/x/b.go", "internal/x/c.go")
	gitForCoverageTest(t, root, "commit", "-qm", homeStateCoverageRemediationCommitSubject)
	want := "changed production file missing from diff: internal/x/b.go, internal/x/c.go"
	if seen := resolveRepeatedly(t, root, 20); len(seen) != 1 || seen[want] != 20 {
		t.Fatalf("messages across 20 runs = %v, want only %q", seen, want)
	}
}
