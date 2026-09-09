package cli

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func gitForCoverageTest(t *testing.T, root string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
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
	repo, _ := filepath.Abs(filepath.Join("..", ".."))
	files, _, err := changedProductionFiles(repo)
	if err != nil {
		t.Fatal(err)
	}
	coverage, err := measureChangedSurfaceCoverageWith(t.Context(), repo, func(_ context.Context, _ string, path string) error {
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
	})
	if err != nil || coverage != 100 {
		t.Fatalf("coverage=%.1f err=%v", coverage, err)
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

func TestChangedProductionFilesDerivesCurrentHeadDiffAndPlatformDisposition(t *testing.T) {
	repo, _ := filepath.Abs(filepath.Join("..", ".."))
	native, disposition, err := changedProductionFiles(repo)
	if err != nil {
		t.Fatal(err)
	}
	all := strings.Join(append(append([]string{}, native...), disposition...), "\n")
	for _, required := range []string{"internal/cli/launcher.go", "internal/cli/mcp_server.go", "internal/homestate/factory.go", "internal/homestate/handoff.go", "internal/hook/session_start.go", "internal/kanban/factory_slots.go", "internal/homestate/admission_lock_windows.go"} {
		if !strings.Contains(all, required) {
			t.Errorf("missing changed production file %s", required)
		}
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
	if err := os.Remove(p); err != nil {
		t.Fatal(err)
	}
	if _, _, err := changedProductionFiles(root); err == nil {
		t.Fatal("deleted production file accepted")
	}
}
