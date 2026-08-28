package guardliveness

import (
	"strings"
	"testing"
	"time"
)

// snapshotAt pairs a result with the moment it was taken. Every age assertion
// below is stated against THIS timestamp rather than against the render moment,
// which is the whole of REQ-GDL-006.
func snapshotAt(r Result, takenAt time.Time) Snapshot {
	return Snapshot{TakenAt: takenAt, Result: r}
}

// AC-GDL-002 — the trigger is the clean/non-clean partition, never a list.
//
// The second fixture is deliberately shaped so that no implementation can pass
// it by special-casing: every entry is non-clean and no two share a
// classification, so a trigger enumerating values would have to name three
// unrelated ones to fire on all of them.
func TestAdvisoryRendersOnAnyNonCleanEntry(t *testing.T) {
	now := time.Now()
	allDistinct := Result{
		Clean: designate("alpha"),
		Entries: []Entry{
			entry("subject-a", "beta", "settled"),
			entry("subject-b", "gamma", "unsettled"),
			entry("subject-c", "delta", "settled"),
		},
	}

	for _, tc := range []struct {
		name   string
		result Result
		want   []string
	}{
		{"one non-clean entry among clean ones", resultA(), []string{"subject-2", "subject-3"}},
		{"every entry non-clean, no two sharing a classification", allDistinct, []string{"subject-a", "subject-b", "subject-c"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			text := Advisory(snapshotAt(tc.result, now.Add(-90*time.Second)), now)
			if text == "" {
				t.Fatalf("advisory was silent on a result carrying non-clean entries")
			}
			for _, subject := range tc.want {
				if !strings.Contains(text, subject) {
					t.Errorf("advisory omits non-clean subject %q:\n%s", subject, text)
				}
			}
		})
	}
}

// REQ-GDL-004's other half: a conforming result with nothing to report says
// nothing. Without this, "fires on any non-clean entry" is satisfied by an
// advisory that fires on everything.
func TestAdvisoryIsSilentOnAConformingAllCleanResult(t *testing.T) {
	now := time.Now()
	allClean := Result{
		Clean:   designate("alpha"),
		Entries: []Entry{entry("subject-1", "alpha", "settled")},
	}
	if text := Advisory(snapshotAt(allClean, now.Add(-time.Hour)), now); text != "" {
		t.Fatalf("advisory spoke on an all-clean result: %q", text)
	}
}

// AC-GDL-006 — the stated age is derived from the persisted result's OWN
// timestamp, not from the render moment.
//
// Two results at distinct times is what discriminates the two mutants at once:
// a renderer computing age from the render moment prints a zero age, and a
// constant-offset renderer prints the SAME age twice.
func TestAdvisoryAgeComesFromThePersistedTimestamp(t *testing.T) {
	now := time.Now()
	first := Advisory(snapshotAt(resultA(), now.Add(-90*time.Minute)), now)
	second := Advisory(snapshotAt(resultA(), now.Add(-30*time.Minute)), now)

	if !strings.Contains(first, "1h30m") {
		t.Errorf("advisory taken 90m ago does not state that age:\n%s", first)
	}
	if !strings.Contains(second, "30m") {
		t.Errorf("advisory taken 30m ago does not state that age:\n%s", second)
	}
	if first == second {
		t.Fatalf("two results persisted at distinct times rendered the same advisory — the age is a constant, not a derivation:\n%s", first)
	}
	for _, zero := range []string{"0ms", "0s ago"} {
		if strings.Contains(first, zero) || strings.Contains(second, zero) {
			t.Errorf("advisory states a zero age — the age is taken at the render moment")
		}
	}
}

// A snapshot with no recorded timestamp is not a measurement, and reporting one
// as though it were current is the failure REQ-GDL-006 exists to prevent.
func TestAdvisoryIsSilentWithoutARecordedTimestamp(t *testing.T) {
	if text := Advisory(Snapshot{Result: resultA()}, time.Now()); text != "" {
		t.Fatalf("advisory rendered from a snapshot carrying no timestamp: %q", text)
	}
}

// AC-GDL-013 — a contract-violating result is REPORTED, never rendered green.
//
// Five fixtures, exactly the negations of the contract's clauses. The two
// entry-shaped ones are not decorative: an entry carrying two classifications
// matches the clean value on comparison and would be treated as
// nothing-to-report, so it under-fires rather than failing loudly.
func TestAdvisoryNamesTheContractViolationAndNeverReportsAllClear(t *testing.T) {
	now := time.Now()
	cleanEntry := entry("subject-1", "alpha", "settled")

	for _, tc := range []struct {
		name   string
		result Result
		names  string
	}{
		{"designation absent", Result{Clean: nil, Entries: []Entry{cleanEntry}}, "no clean-value designation"},
		{"designation null", Result{Clean: &Designation{}, Entries: []Entry{cleanEntry}}, "designation is null"},
		{"designation multi-valued", Result{Clean: designate("alpha", "quiet"), Entries: []Entry{cleanEntry}}, "more than one value"},
		{"entry carries no classification", Result{
			Clean:   designate("alpha"),
			Entries: []Entry{{Subject: "subject-1", Surface: "settled"}},
		}, "exactly one classification"},
		{"entry carries two classifications", Result{
			Clean:   designate("alpha"),
			Entries: []Entry{{Subject: "subject-1", Classifications: []string{"alpha", "beta"}, Surface: "settled"}},
		}, "exactly one classification"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			text := Advisory(snapshotAt(tc.result, now.Add(-5*time.Minute)), now)
			if text == "" {
				t.Fatalf("silence on a contract-violating result — the partition had no referent and nothing rendered, which is an all-clear by omission")
			}
			if !strings.Contains(text, contractViolationMarker) {
				t.Errorf("advisory does not report a contract violation:\n%s", text)
			}
			if !strings.Contains(text, tc.names) {
				t.Errorf("advisory does not NAME the violation (want %q):\n%s", tc.names, text)
			}
			if !strings.Contains(text, "5m") {
				t.Errorf("advisory omits the measurement's age:\n%s", text)
			}
		})
	}
}
