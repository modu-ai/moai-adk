package runtime

import (
	"path/filepath"
	"strings"
	"testing"
)

// SPEC-AUDIT-ENGINE-CONVERGENCE-001 M4 RED, sub-step S2d — both plan-audit round
// counters skip INCONCLUSIVE iteration files (AC-AEC-002b; REQ-AEC-007; card
// t2017).
//
// An audit that did not run is not a round: CountAuditRounds and
// CountPlanAuditRounds keep counting FAIL and PASS-family files and skip files
// labelled INCONCLUSIVE, so a quota wall cannot spend the ceiling. An INCONCLUSIVE
// file with the highest iteration number is still the latest verdict (Latest and
// SelectLatestVerdict), so an older verdict never becomes the decision.

const roundsSpec = "SPEC-ROUNDS-001"

// roundDir writes plan-audit-iter<N>.md files, one per label, label i at iteration i+1.
func roundDir(t *testing.T, labels ...string) string {
	t.Helper()
	dir := t.TempDir()
	for i, label := range labels {
		writeAuditFixture(t, dir, "", roundsSpec, label, i+1)
	}
	return dir
}

// The acceptance fixture: FAIL (1), INCONCLUSIVE (2), FAIL (3) counts 2 on both
// counters.
func TestRoundCounters_SkipAnInconclusiveIteration(t *testing.T) {
	cases := []struct {
		name   string
		labels []string
		want   int
	}{
		{"FAIL, INCONCLUSIVE, FAIL", []string{"FAIL", "INCONCLUSIVE", "FAIL"}, 2},
		{"FAIL, INCONCLUSIVE", []string{"FAIL", "INCONCLUSIVE"}, 1},
		{"INCONCLUSIVE only", []string{"INCONCLUSIVE"}, 0},
		{"PASS, INCONCLUSIVE, PASS-WITH-DEBT", []string{"PASS", "INCONCLUSIVE", "PASS-WITH-DEBT"}, 2},
		{"two INCONCLUSIVE rounds in a row", []string{"FAIL", "INCONCLUSIVE", "INCONCLUSIVE", "FAIL"}, 2},
	}
	for _, c := range cases {
		t.Run("CountAuditRounds/"+c.name, func(t *testing.T) {
			ev, err := CountAuditRounds(roundsSpec, []string{roundDir(t, c.labels...)})
			if err != nil {
				t.Fatal(err)
			}
			if ev.Count != c.want {
				t.Errorf("CountAuditRounds counted %d, want %d", ev.Count, c.want)
			}
		})
		t.Run("CountPlanAuditRounds/"+c.name, func(t *testing.T) {
			got, err := CountPlanAuditRounds(roundDir(t, c.labels...), nil)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Errorf("CountPlanAuditRounds counted %d, want %d", got, c.want)
			}
		})
	}
}

// The skip is by label, not by file: a FAIL file after an INCONCLUSIVE one keeps
// counting, and a file that carries no readable label still counts (a count the
// counter cannot verify must not shrink). Controls: pass today.
func TestRoundCounters_OnlyTheInconclusiveLabelIsSkipped_Control(t *testing.T) {
	ev, err := CountAuditRounds(roundsSpec, []string{roundDir(t, "FAIL", "FAIL", "PASS")})
	if err != nil || ev.Count != 3 {
		t.Errorf("CountAuditRounds over FAIL, FAIL, PASS = %d (err %v), want 3", ev.Count, err)
	}
	if got, err := CountPlanAuditRounds(roundDir(t, "FAIL", "FAIL", "PASS"), nil); err != nil || got != 3 {
		t.Errorf("CountPlanAuditRounds over FAIL, FAIL, PASS = %d (err %v), want 3", got, err)
	}

	dir := t.TempDir()
	writeAuditFixture(t, dir, "", roundsSpec, "FAIL", 1)
	writeRoundFile(t, dir, "plan-audit-iter2.md") // empty: no label to read
	if got, err := CountPlanAuditRounds(dir, nil); err != nil || got != 2 {
		t.Errorf("a file without a readable label must still count: got %d (err %v), want 2", got, err)
	}
}

// The legacy stream (<SPEC-ID>-review-<N>.md) is evidence too, so an INCONCLUSIVE
// legacy file is skipped as well.
func TestCountAuditRounds_SkipsAnInconclusiveLegacyFile(t *testing.T) {
	dir := t.TempDir()
	writeAuditFixture(t, dir, roundsSpec+"-review-1.md", roundsSpec, "FAIL", 0)
	writeAuditFixture(t, dir, roundsSpec+"-review-2.md", roundsSpec, "INCONCLUSIVE", 0)
	ev, err := CountAuditRounds(roundsSpec, []string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if ev.Count != 1 {
		t.Errorf("legacy FAIL, INCONCLUSIVE counted %d, want 1", ev.Count)
	}
}

// An INCONCLUSIVE file with the highest iteration number is still the latest
// verdict: the count skips it, the selection does not. Controls: pass today (the
// selection does not read labels); the Latest label arm is the part that reads the
// file.
func TestLatestVerdict_StillSelectsAnInconclusiveFile(t *testing.T) {
	dir := roundDir(t, "FAIL", "INCONCLUSIVE")

	ev, err := CountAuditRounds(roundsSpec, []string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(filepath.ToSlash(ev.LatestPath), "plan-audit-iter2.md") {
		t.Errorf("Latest = %q, want iteration 2 (the INCONCLUSIVE file)", ev.LatestPath)
	}
	if ev.Latest == nil || ev.Latest.Label != "INCONCLUSIVE" {
		t.Errorf("the parsed latest verdict = %+v, want label INCONCLUSIVE", ev.Latest)
	}

	got, err := SelectLatestVerdict(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(filepath.ToSlash(got), "plan-audit-iter2.md") {
		t.Errorf("SelectLatestVerdict = %q, want iteration 2 (an older verdict must not become the decision)", got)
	}
}
