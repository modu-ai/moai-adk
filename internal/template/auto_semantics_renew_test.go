package template

// auto_semantics_renew_test.go — SPEC-FACTORY-LEASE-RENEW-001 (card t1663, M3).
//
// AC-FLR-013 (REQ-FLR-012): the standing wake is armed as a minute step of the
// renewal period that starts at minute 0, so no two consecutive scheduled wakes
// are further apart than one period. AC-FLR-015 (REQ-FLR-013): the stall window
// is counted from the last observed channel change. Both criteria check shipped
// text and a pure model; neither observes a live watchdog turn.

import (
	"io/fs"
	"strings"
	"testing"
)

const (
	autoSemanticsPath   = ".claude/rules/moai/workflow/auto-semantics.md"
	factoryDispatchPath = ".claude/rules/moai/workflow/factory-dispatch.md"
	retiredArmLiteral   = "7,27,47 * * * *"
)

// embeddedText returns the text of one embedded template file.
func embeddedText(t *testing.T, path string) string {
	t.Helper()
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates(): %v", err)
	}
	raw, err := fs.ReadFile(fsys, path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(raw)
}

// maxFireGap returns the largest gap, in minutes, between consecutive scheduled
// fire minutes of the cron minute field start/step over one hour. The gap across
// the hour boundary is 60 plus the first fire minute minus the last one. A form
// with a single fire minute reads as a gap of 60.
func maxFireGap(start, step int) int {
	var fires []int
	for m := start; m < 60; m += step {
		fires = append(fires, m)
	}
	if len(fires) == 0 {
		return 60
	}
	gap := 60 + fires[0] - fires[len(fires)-1]
	for i := 1; i < len(fires); i++ {
		if d := fires[i] - fires[i-1]; d > gap {
			gap = d
		}
	}
	return gap
}

// armParagraph returns the standing-wake Arm bullet of auto-semantics §5.1.
func armParagraph(t *testing.T, doc string) string {
	t.Helper()
	sec, ok := sectionOf(doc, "### 5.1 The standing awaken carrier")
	if !ok {
		t.Fatalf("%s: §5.1 not found", autoSemanticsPath)
	}
	start := strings.Index(sec, "- **Arm.**")
	end := strings.Index(sec, "- **Keep.**")
	if start < 0 || end < start {
		t.Fatalf("%s: §5.1 has no Arm bullet bounded by the Keep bullet", autoSemanticsPath)
	}
	return sec[start:end]
}

// TestAutoSemanticsArmFollowsRenewInterval is AC-FLR-013 (REQ-FLR-012). The text
// subtest reads the Arm bullet and the factory-dispatch cadence sentence. The
// gap subtests run the largest-gap computation over every legal whole-minute
// period, and over the two mutant forms the criterion names.
func TestAutoSemanticsArmFollowsRenewInterval(t *testing.T) {
	t.Run("text", func(t *testing.T) {
		arm := armParagraph(t, embeddedText(t, autoSemanticsPath))
		for _, want := range []string{"*/N * * * *", "renew_interval", "no two consecutive scheduled wakes are further apart than one renewal period"} {
			if !strings.Contains(arm, want) {
				t.Errorf("the Arm bullet lacks %q", want)
			}
		}
		if strings.Contains(arm, retiredArmLiteral) {
			t.Errorf("the Arm bullet still carries the fixed literal %q", retiredArmLiteral)
		}
		if strings.Contains(embeddedText(t, factoryDispatchPath), retiredArmLiteral) {
			t.Errorf("%s still carries the fixed literal %q", factoryDispatchPath, retiredArmLiteral)
		}
	})
	t.Run("gap-every-N", func(t *testing.T) {
		for n := 1; n <= 59; n++ {
			if g := maxFireGap(0, n); g > n {
				t.Errorf("*/%d: largest gap %d minutes exceeds the period %d", n, g, n)
			}
		}
	})
	t.Run("mutant-2-59-refused", func(t *testing.T) {
		if g := maxFireGap(2, 59); g <= 59 {
			t.Errorf("2/59 passes the gap bound: largest gap %d, want a refusal above 59", g)
		} else {
			t.Logf("2/59 refused: largest gap %d minutes exceeds the period 59", g)
		}
	})
	t.Run("mutant-6-7-refused", func(t *testing.T) {
		if g := maxFireGap(6, 7); g <= 7 {
			t.Errorf("6/7 passes the gap bound: largest gap %d, want a refusal above 7", g)
		} else {
			t.Logf("6/7 refused: largest gap %d minutes exceeds the period 7", g)
		}
	})
}

// stallObservation is one scheduled watchdog pass in the model: its minute, and
// whether a channel differs from the previous pass.
type stallObservation struct {
	at      int
	changed bool
}

// stallPasses returns the scheduled passes from minute 0 through horizon at the
// given period. changeAt is the minute of one channel change, observed at the
// first pass at or after it, or -1 for no change. The first pass is the first
// observation.
func stallPasses(period, horizon, changeAt int) []stallObservation {
	var obs []stallObservation
	for t := 0; t <= horizon; t += period {
		changed := changeAt >= 0 && t >= changeAt && (len(obs) == 0 || obs[len(obs)-1].at < changeAt)
		obs = append(obs, stallObservation{at: t, changed: changed})
	}
	return obs
}

// stallVerdictByChange is REQ-FLR-013 as auto-semantics §4 is to state it: a
// stall when no channel changed since the previous pass and at least n minutes
// have passed since the last observed channel change. The first observation
// records its own time as the last change and gives no verdict.
func stallVerdictByChange(obs []stallObservation, n int) []bool {
	verdicts := make([]bool, len(obs))
	lastChange := 0
	for i, o := range obs {
		if i == 0 || o.changed {
			lastChange = o.at
			continue
		}
		verdicts[i] = o.at-lastChange >= n
	}
	return verdicts
}

// stallVerdictByPrevious is the rule auto-semantics §4 states today: a stall
// when no channel changed since the previous snapshot and the previous
// observation is at least n minutes old. The first observation gives no verdict.
func stallVerdictByPrevious(obs []stallObservation, n int) []bool {
	verdicts := make([]bool, len(obs))
	for i := 1; i < len(obs); i++ {
		verdicts[i] = !obs[i].changed && obs[i].at-obs[i-1].at >= n
	}
	return verdicts
}

// checkStallReachedAt asserts that the first stall falls at wantFirst, and that
// every later scheduled pass in the observation is also a stall.
func checkStallReachedAt(t *testing.T, obs []stallObservation, verdicts []bool, wantFirst int) {
	t.Helper()
	first := -1
	for i, o := range obs {
		if verdicts[i] {
			first = o.at
			break
		}
	}
	if first != wantFirst {
		t.Errorf("first stall at minute %d, want minute %d", first, wantFirst)
		return
	}
	for i, o := range obs {
		if o.at >= first && !verdicts[i] {
			t.Errorf("pass at minute %d reports no stall although no channel changed", o.at)
		}
	}
}

// TestStallVerdictReachesAtRenewCadence is AC-FLR-015 (REQ-FLR-013). The stall
// window N is 15 minutes (auto-semantics §3). The model subtests run the rule as
// REQ-FLR-013 states it over a deterministic pass sequence; the rule-text subtest
// pins auto-semantics §4.
func TestStallVerdictReachesAtRenewCadence(t *testing.T) {
	const window = 15
	t.Run("period-5m", func(t *testing.T) {
		obs := stallPasses(5, 120, -1)
		checkStallReachedAt(t, obs, stallVerdictByChange(obs, window), 15)
	})
	t.Run("period-7m", func(t *testing.T) {
		obs := stallPasses(7, 120, -1)
		checkStallReachedAt(t, obs, stallVerdictByChange(obs, window), 21)
	})
	t.Run("old-cadence-equal", func(t *testing.T) {
		obs := stallPasses(20, 120, 40)
		byChange := stallVerdictByChange(obs, window)
		byPrevious := stallVerdictByPrevious(obs, window)
		for i, o := range obs {
			if byChange[i] != byPrevious[i] {
				t.Errorf("pass at minute %d: REQ-FLR-013 verdict %t, previous-snapshot verdict %t", o.at, byChange[i], byPrevious[i])
			}
		}
	})
	t.Run("rule-text", func(t *testing.T) {
		sec, ok := sectionOf(embeddedText(t, autoSemanticsPath), "## 4. The observation snapshot")
		if !ok {
			t.Fatalf("%s: §4 not found", autoSemanticsPath)
		}
		for _, want := range []string{"last_change_at", "renew_result"} {
			if !strings.Contains(sec, want) {
				t.Errorf("§4 does not name %q", want)
			}
		}
		if strings.Contains(sec, "the previous observation is at least N minutes old") {
			t.Errorf("§4 still states the previous-observation clause as the stall condition")
		}
	})
}
