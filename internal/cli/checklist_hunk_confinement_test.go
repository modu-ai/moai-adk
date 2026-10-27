package cli

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// AC-TCL-022 (SPEC-TUI-CHECKLIST-001): every change this card makes to
// internal/cli/init.go sits in one of three regions of the region base
// cb30e05b3e. The test is a one-off coordination check for this card (it is
// removed at M6 after its last run is recorded in progress.md §E.2).

const hunkRegionBase = "cb30e05b3e"

// hunkRegions are the allowed regions of the region base's init.go, inclusive.
// R3 admits insertions only (b == 0).
var hunkRegions = []struct {
	name       string
	start, end int
	insertOnly bool
}{
	{"R1", 383, 389, false},
	{"R3", 919, 1075, true},
	{"R2", 1077, 1087, false},
}

var hunkHeaderRe = regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+\d+(?:,\d+)? @@`)

type hunkOld struct{ a, b int }

// parseHunks reads the old-side ranges of a -U0 diff.
func parseHunks(diff string) []hunkOld {
	var hunks []hunkOld
	for _, line := range strings.Split(diff, "\n") {
		m := hunkHeaderRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		a, _ := strconv.Atoi(m[1])
		b := 1
		if m[2] != "" {
			b, _ = strconv.Atoi(m[2])
		}
		hunks = append(hunks, hunkOld{a, b})
	}
	return hunks
}

// judgeHunk returns "" when the hunk is confined, else the reason.
func judgeHunk(h hunkOld) string {
	for _, r := range hunkRegions {
		if h.b == 0 {
			if h.a >= r.start && h.a <= r.end {
				return ""
			}
			continue
		}
		last := h.a + h.b - 1
		if h.a >= r.start && last <= r.end {
			if r.insertOnly {
				return fmt.Sprintf("hunk -%d,%d rewrites existing lines inside %s, which admits insertions only", h.a, h.b, r.name)
			}
			return ""
		}
	}
	return fmt.Sprintf("hunk -%d,%d lies outside R1 [383,389], R3 [919,1075] and R2 [1077,1087]", h.a, h.b)
}

func gitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed (git missing or history too shallow to hold the region base %s): %v\n%s", strings.Join(args, " "), hunkRegionBase, err, out)
	}
	return string(out)
}

func TestChecklistInitHunkConfinement(t *testing.T) {
	// The judge is not vacuous: the four mutants of the acceptance text.
	t.Run("judge_mutants", func(t *testing.T) {
		for name, tc := range map[string]struct {
			h    hunkOld
			want bool
		}{
			"reporter_line_728":       {hunkOld{728, 1}, false},
			"1070_to_1100":            {hunkOld{1070, 31}, false},
			"r3_rewrite":              {hunkOld{950, 2}, false},
			"r3_insertion":            {hunkOld{950, 0}, true},
			"r2_edit":                 {hunkOld{1081, 7}, true},
			"r1_edit":                 {hunkOld{388, 1}, true},
			"insertion_outside":       {hunkOld{700, 0}, false},
			"insertion_between_r3_r2": {hunkOld{1076, 0}, false},
		} {
			if got := judgeHunk(tc.h) == ""; got != tc.want {
				t.Errorf("%s: confined = %v, want %v", name, got, tc.want)
			}
		}
		if hs := parseHunks("@@ -388 +388,3 @@ x\n@@ -950,0 +951,2 @@\n@@ -1081,7 +1092 @@"); len(hs) != 3 || hs[0] != (hunkOld{388, 1}) || hs[1] != (hunkOld{950, 0}) || hs[2] != (hunkOld{1081, 7}) {
			t.Errorf("hunk header parsing is wrong: %v", hs)
		}
	})

	t.Run("card_changes", func(t *testing.T) {
		repo := filepath.Dir(filepath.Dir(inventoryRoot(t)))
		gitOutput(t, repo, "rev-parse", "--verify", hunkRegionBase+"^{commit}")
		diff := gitOutput(t, repo, "diff", "-U0", hunkRegionBase, "HEAD", "--", "internal/cli/init.go")
		hunks := parseHunks(diff)
		if len(hunks) < 1 {
			t.Fatalf("no init.go hunk since %s: the wiring is not committed (or the diff is empty)", hunkRegionBase)
		}
		for _, h := range hunks {
			if reason := judgeHunk(h); reason != "" {
				t.Errorf("%s", reason)
			}
		}
		if merges := strings.TrimSpace(gitOutput(t, repo, "log", "--merges", "--format=%h", hunkRegionBase+"..HEAD", "--", "internal/cli/init.go")); merges != "" {
			t.Errorf("a merge commit touches init.go since %s: %s", hunkRegionBase, merges)
		}
		t.Logf("measured %d hunks at HEAD %s", len(hunks), strings.TrimSpace(gitOutput(t, repo, "rev-parse", "HEAD")))
	})
}
