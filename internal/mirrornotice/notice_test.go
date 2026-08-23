package mirrornotice_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/mirrornotice"
	"github.com/modu-ai/moai-adk/internal/template"
)

// entries builds n mirror entries in the given mode, each carrying a distinct
// warning so a per-item emission is distinguishable from a summary.
func entries(mode template.MirrorMode, n int) []template.SkillMirrorEntry {
	out := make([]template.SkillMirrorEntry, 0, n)
	for i := range n {
		out = append(out, template.SkillMirrorEntry{
			Skill:   fmt.Sprintf("moai-skill-%02d", i),
			Mode:    mode,
			Warning: fmt.Sprintf("cannot create moai-skill-%02d: permission denied", i),
		})
	}
	return out
}

func result(groups ...[]template.SkillMirrorEntry) *template.DeployResult {
	res := &template.DeployResult{}
	for _, g := range groups {
		res.SkillMirrors = append(res.SkillMirrors, g...)
	}
	return res
}

// AC-DRW-001 (notice + count): a run reporting copy-fallback entries produces a
// notice, and that notice names how many skills fell back. Asserting only that
// a notice exists would pass on an implementation that lost the count.
func TestLines_CopyFallbackCarriesCount(t *testing.T) {
	t.Parallel()

	lines := mirrornotice.Lines(result(entries(template.MirrorModeCopy, 12)))
	if len(lines) == 0 {
		t.Fatal("copy fallback produced no notice; the fallback is silent")
	}
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, mirrornotice.Token) {
		t.Errorf("notice does not carry the stable token %q: %s", mirrornotice.Token, joined)
	}
	if !strings.Contains(joined, "12") {
		t.Errorf("notice does not carry the fallback count 12: %s", joined)
	}
}

// AC-DRW-002: a run with no fallback and no failure emits nothing at all.
// Without this arm an implementation that emits unconditionally still passes
// AC-DRW-001.
func TestLines_NoFallbackEmitsNothing(t *testing.T) {
	t.Parallel()

	if lines := mirrornotice.Lines(result(entries(template.MirrorModeSymlink, 34))); len(lines) != 0 {
		t.Errorf("symlink-only run produced a notice: %v", lines)
	}
	if lines := mirrornotice.Lines(&template.DeployResult{}); len(lines) != 0 {
		t.Errorf("empty result produced a notice: %v", lines)
	}
	if lines := mirrornotice.Lines(nil); len(lines) != 0 {
		t.Errorf("nil result produced a notice: %v", lines)
	}

	var buf bytes.Buffer
	mirrornotice.WriteTo(&buf, result(entries(template.MirrorModeSymlink, 34)))
	if buf.Len() != 0 {
		t.Errorf("WriteTo wrote %q on a run with nothing to report", buf.String())
	}
}

// AC-DRW-004 arm 1 — the copy notice does not grow with the number of skills.
func TestLines_CopyNoticeIsNotProportional(t *testing.T) {
	t.Parallel()

	few := mirrornotice.Lines(result(entries(template.MirrorModeCopy, 2)))
	many := mirrornotice.Lines(result(entries(template.MirrorModeCopy, 34)))
	if len(few) != len(many) {
		t.Errorf("copy notice length grows with skill count: 2 skills → %d lines, 34 skills → %d lines\n2: %v\n34: %v",
			len(few), len(many), few, many)
	}
}

// AC-DRW-004 arm 2 — the failed notice does not grow either. The low sample is
// 4, ABOVE the 3-example cap: at 2 a correct implementation legitimately emits
// fewer lines than at 34 (1+2 vs 1+3), so a 2-vs-34 comparison would fail a
// correct implementation rather than a proportional one.
func TestLines_FailedNoticeIsNotProportional(t *testing.T) {
	t.Parallel()

	few := mirrornotice.Lines(result(entries(template.MirrorModeFailed, 4)))
	many := mirrornotice.Lines(result(entries(template.MirrorModeFailed, 34)))
	if len(few) != len(many) {
		t.Errorf("failed notice length grows with skill count: 4 skills → %d lines, 34 skills → %d lines\n4: %v\n34: %v",
			len(few), len(many), few, many)
	}
}

// AC-DRW-004 arm 3 — the cap itself. Kept separate from arm 2 so a broken cap
// and a broken non-proportionality are distinguishable.
func TestLines_FailedExamplesAreCappedAtThree(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct{ failed, wantExamples int }{
		{failed: 2, wantExamples: 2},
		{failed: 34, wantExamples: 3},
	} {
		lines := mirrornotice.Lines(result(entries(template.MirrorModeFailed, tc.failed)))
		got := 0
		for _, l := range lines {
			if strings.Contains(l, "permission denied") {
				got++
			}
		}
		if got != tc.wantExamples {
			t.Errorf("%d failed skills → %d example warnings, want %d\nlines: %v",
				tc.failed, got, tc.wantExamples, lines)
		}
	}
}

// AC-DRW-006 arms 1-2 — the failure count reaches the user and between one and
// three warnings come with it.
func TestLines_FailedCountAndWarningsReachTheUser(t *testing.T) {
	t.Parallel()

	lines := mirrornotice.Lines(result(entries(template.MirrorModeFailed, 5)))
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "5") {
		t.Errorf("failure count 5 absent from notice: %s", joined)
	}
	warnings := strings.Count(joined, "permission denied")
	if warnings < 1 || warnings > 3 {
		t.Errorf("got %d warning quotes, want between 1 and 3: %s", warnings, joined)
	}
}

// AC-DRW-008 — the skipped-mode warning misattributes the deployer's own prior
// copy to the user, so it must not reach the user at all.
func TestLines_SkippedWarningsAreNotForwarded(t *testing.T) {
	t.Parallel()

	skipped := make([]template.SkillMirrorEntry, 0, 34)
	for i := range 34 {
		skipped = append(skipped, template.SkillMirrorEntry{
			Skill: fmt.Sprintf("moai-skill-%02d", i),
			Mode:  template.MirrorModeSkipped,
			Warning: fmt.Sprintf(
				"a non-symlink entry already exists at .agents/skills/moai-skill-%02d — left untouched", i),
		})
	}

	lines := mirrornotice.Lines(result(skipped))
	joined := strings.Join(lines, "\n")
	if strings.Contains(joined, "a non-symlink entry already exists") {
		t.Errorf("misattributing skipped warning forwarded to the user: %s", joined)
	}
	if len(lines) != 0 {
		t.Errorf("skipped-only run produced a notice: %v", lines)
	}
}

// A mixed run reports both outcomes; skipped stays out of it.
func TestLines_MixedRunReportsCopyAndFailedOnly(t *testing.T) {
	t.Parallel()

	res := result(
		entries(template.MirrorModeSymlink, 20),
		entries(template.MirrorModeCopy, 7),
		entries(template.MirrorModeFailed, 2),
	)
	res.SkillMirrors = append(res.SkillMirrors, template.SkillMirrorEntry{
		Skill:   "moai-skipped",
		Mode:    template.MirrorModeSkipped,
		Warning: "a non-symlink entry already exists at .agents/skills/moai-skipped — left untouched",
	})

	joined := strings.Join(mirrornotice.Lines(res), "\n")
	if !strings.Contains(joined, "7") {
		t.Errorf("copy count 7 absent: %s", joined)
	}
	if !strings.Contains(joined, "2") {
		t.Errorf("failure count 2 absent: %s", joined)
	}
	if strings.Contains(joined, "a non-symlink entry already exists") {
		t.Errorf("skipped warning leaked into a mixed-run notice: %s", joined)
	}
}
