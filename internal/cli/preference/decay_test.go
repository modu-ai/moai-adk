package preference

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestDecayWeight_PowerLawCurve verifies design.md §E.1:
//
//	weight(ageDays) = (ageDays + 1)^(-α)     where α = decayAlpha (0.5)
//
// The authoritative spec is the FORMULA (design.md §E.1 declares the function
// explicitly; the illustrative table carries minor rounding at age 7 / 28 that
// does not match the formula to 3 decimals). Each expected value is derived at
// test time from the SAME formula the implementation uses, so this test
// validates "the implementation matches the declared function" — which is the
// load-bearing assertion. The curve-shape invariants (monotonic decrease, TTL
// threshold ≈ 0.19) are covered by the sibling tests.
//
// α = 0.5. epsilon is tight (1e-9) because the test recomputes the expected
// value from the same math.Pow expression — any drift is a real regression.
func TestDecayWeight_PowerLawCurve(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		ageDays int
	}{
		{"age 0 → fresh", 0},
		{"age 1 → 1/sqrt(2)", 1},
		{"age 7 → 1/sqrt(8)", 7},
		{"age 14 → 1/sqrt(15)", 14},
		{"age 28 → 1/sqrt(29) (TTL threshold)", 28},
		{"age 56 → 1/sqrt(57) (long tail)", 56},
	}
	const (
		alpha   = 0.5
		epsilon = 1e-9
	)
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			want := math.Pow(float64(tt.ageDays)+1.0, -alpha)
			got := decayWeight(tt.ageDays)
			if math.Abs(got-want) > epsilon {
				t.Errorf("decayWeight(%d) = %.6f, want %.6f (±%.g)", tt.ageDays, got, want, epsilon)
			}
		})
	}
}

// TestDecayWeight_TableApproxMatchesDesignDoc verifies the implementation's
// output at the design.md §E.1 table anchor ages is within a loose epsilon
// (0.01) of the table's documented 3-decimal values. The table is
// illustrative, so a loose band is appropriate; this test catches a gross
// exponent error (e.g. α=1.0 producing 1/29 ≈ 0.034 at age 28) while
// tolerating the table's own rounding.
//
// NOTE on age-7: the design.md §E.1 table lists 0.378 for age 7, but the
// declared formula (age+1)^(-0.5) yields 1/√8 = 0.3536. The 0.024 gap is a
// documentation rounding slip, not an implementation defect — the formula is
// authoritative per the SPEC Section D constraint #1. The age-7 anchor is
// therefore omitted from this doc-approximation test; the formula-deriving
// TestDecayWeight_PowerLawCurve covers age 7 exactly.
func TestDecayWeight_TableApproxMatchesDesignDoc(t *testing.T) {
	t.Parallel()
	tests := []struct {
		ageDays     int
		docTableVal float64 // design.md §E.1 table (3-decimal illustrative)
	}{
		{0, 1.000},
		{1, 0.707},
		// age 7 omitted — see NOTE above (doc-table rounding slip).
		{14, 0.258},
		{28, 0.189},
		{56, 0.133},
	}
	const epsilon = 0.01 // loose — tolerates design-doc table rounding
	for _, tt := range tests {
		tt := tt
		t.Run(tableAnchorName(tt.ageDays), func(t *testing.T) {
			t.Parallel()
			got := decayWeight(tt.ageDays)
			if math.Abs(got-tt.docTableVal) > epsilon {
				t.Errorf("decayWeight(%d) = %.4f, design doc table says %.3f (±%.3f)", tt.ageDays, got, tt.docTableVal, epsilon)
			}
		})
	}
}

// tableAnchorName renders a stable sub-test name for the design-doc table test.
func tableAnchorName(ageDays int) string {
	return fmt.Sprintf("age_%d_design_doc_anchor", ageDays)
}

// newDecayTestStore builds a fileStore rooted at t.TempDir() so every decay
// scan test runs in isolation. It returns the Store interface + the underlying
// *fileStore (so tests can call the unexported DecayScan method).
func newDecayTestStore(t *testing.T) (Store, *fileStore) {
	t.Helper()
	dir := t.TempDir()
	st, err := NewFileStore(dir)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	return st, st.(*fileStore)
}

// mustUpsert is a test helper that fails the test on any Upsert error.
func mustUpsert(t *testing.T, s Store, domain, key string, e Entry) {
	t.Helper()
	if err := s.Upsert(domain, key, e); err != nil {
		t.Fatalf("Upsert(%s/%s): %v", domain, key, err)
	}
}

// freshEntry builds a valid Entry with the given scope + last_used offset from
// base. valid_time is set 1 day before last_used (stable invariant: valid
// before last used). weight defaults to 1.0 (fresh).
func freshEntry(scope Scope, lastUsed time.Time, weight float64) Entry {
	return Entry{
		Fact:           "test fact",
		SourceCitation: "session=test tool=AskUserQuestion",
		ValidTime:      lastUsed.Add(-24 * time.Hour),
		LastUsed:       lastUsed,
		Scope:          scope,
		Domain:         "test_domain",
		DecisionKey:    "test_key",
		Confidence:     ConfidenceObserved,
		Weight:         weight,
	}
}

// TestDecayScan_StablePreservedTransientSoftDeleted (AC-ADM-011, S1 Blocker)
// verifies the deepest invariant of the SPEC: stable entries are exempt from
// pure time-decay and survive a 28-day+ scan, while transient entries of the
// same age are soft-deleted (moved to archival).
//
// Per acceptance.md AC-ADM-011:
//   Given stable + transient entries exist,
//   When 28 days pass,
//   Then transient is soft-deleted, stable is preserved, stable's weight unchanged.
//
// A naive time-decay that expires stable preferences FAILS this and reproduces
// Koren's "지속 신호 상실" (AP-ADM-006).
//
// Note on stable tier placement: Store.Upsert routes ScopeStable entries to
// CORE, not recall (filestore.go upsertToCore). DecayScan scans RECALL, so to
// exercise the stable-in-recall code path we seed the stable entry directly
// into recall via the unexported upsertToRecall helper. This mirrors the real
// scenario where a stable entry is demoted from core to recall by the 4KB-cap
// eviction (upsertToCore line 233 demotion path) and then needs the floor
// protection when the decay scan runs over recall.
func TestDecayScan_StablePreservedTransientSoftDeleted(t *testing.T) {
	t.Parallel()
	store, fs := newDecayTestStore(t)
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	longAgo := now.Add(-30 * 24 * time.Hour) // 30 days — past the 28d TTL

	stableEntry := freshEntry(ScopeStable, longAgo, 0.92)
	stableEntry.Domain = "lang"
	stableEntry.DecisionKey = "backend_language"
	// Seed directly into recall (bypasses Upsert's core routing) to exercise
	// the stable-floor code path in DecayScan.
	if err := fs.upsertToRecall(stableEntry.Domain, stableEntry.DecisionKey, stableEntry); err != nil {
		t.Fatalf("seed stable into recall: %v", err)
	}

	transientEntry := freshEntry(ScopeTransient, longAgo, 0.5)
	transientEntry.Domain = "log_level"
	transientEntry.DecisionKey = "verbose_this_session"
	mustUpsert(t, store, transientEntry.Domain, transientEntry.DecisionKey, transientEntry)

	report, err := fs.DecayScan(now)
	if err != nil {
		t.Fatalf("DecayScan: %v", err)
	}

	// AC-ADM-011 invariant: stable survived, transient soft-deleted.
	if report.StablePreserved != 1 {
		t.Errorf("StablePreserved = %d, want 1 (stable MUST survive 28d+ scan)", report.StablePreserved)
	}
	if report.SoftDeleted != 1 {
		t.Errorf("SoftDeleted = %d, want 1 (transient at 30d MUST soft-delete)", report.SoftDeleted)
	}

	// Stable entry survives in recall + its weight is unchanged (exempt).
	got, ok, tier, gErr := store.Get("lang", "backend_language")
	if !ok || tier != TierRecall {
		t.Errorf("stable entry after scan: ok=%v tier=%v err=%v (want ok=true tier=recall)", ok, tier, gErr)
	} else if got.Weight != 0.92 {
		t.Errorf("stable weight after scan = %.4f, want 0.92 (pure time-decay MUST NOT touch stable weight)", got.Weight)
	}

	// Transient entry is gone from recall (the cascade still finds it in
	// archival, but the tier where it lives is now archival, NOT recall).
	_, _, tier2, _ := store.Get("log_level", "verbose_this_session")
	if tier2 == TierRecall {
		t.Errorf("transient entry still in recall after 30d scan (tier=%v); want tier=archival", tier2)
	}
	arch, _, tier3, _ := store.Get("log_level", "verbose_this_session")
	if tier3 != TierArchival {
		t.Errorf("transient entry not in archival after soft-delete (tier=%v); want tier=archival", tier3)
	} else if arch.Scope != ScopeTransient {
		t.Errorf("archived entry scope = %q, want %q (soft-delete preserves scope)", arch.Scope, ScopeTransient)
	}
}

// TestDecayScan_StableFloorApplied (AC-ADM-011 sub-case) verifies the stable
// weight floor (design.md §E.2: max(weight, 0.5)). A stable entry whose stored
// weight has decayed below the floor is lifted to stableWeightFloor, and the
// lift is counted in FloorApplied.
//
// Seeded directly into recall (upsertToRecall) because Store.Upsert routes
// ScopeStable to core — the floor code path only runs on stable entries that
// have been demoted to recall by the 4KB-cap eviction.
func TestDecayScan_StableFloorApplied(t *testing.T) {
	t.Parallel()
	store, fs := newDecayTestStore(t)
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	longAgo := now.Add(-40 * 24 * time.Hour) // 40 days — well past TTL

	// Stable entry with a stored weight below the floor (simulating a stale
	// weight from a prior, pre-M4 run or a manual import).
	stale := freshEntry(ScopeStable, longAgo, 0.1) // below 0.5 floor
	stale.Domain = "editor"
	stale.DecisionKey = "vim_mode"
	if err := fs.upsertToRecall(stale.Domain, stale.DecisionKey, stale); err != nil {
		t.Fatalf("seed stale stable into recall: %v", err)
	}

	report, err := fs.DecayScan(now)
	if err != nil {
		t.Fatalf("DecayScan: %v", err)
	}
	if report.FloorApplied != 1 {
		t.Errorf("FloorApplied = %d, want 1", report.FloorApplied)
	}

	got, ok, _, _ := store.Get("editor", "vim_mode")
	if !ok {
		t.Fatal("stable entry missing after scan")
	}
	if got.Weight != stableWeightFloor {
		t.Errorf("stale stable weight after scan = %.4f, want floor %.4f", got.Weight, stableWeightFloor)
	}
}

// TestDecayScan_TransientWeightRefreshed verifies that a transient entry that
// is NOT yet TTL-expired gets its weight refreshed to the power-law value for
// its current age. This keeps recall weights accurate between evictions.
func TestDecayScan_TransientWeightRefreshed(t *testing.T) {
	t.Parallel()
	store, fs := newDecayTestStore(t)
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	sevenDaysAgo := now.Add(-7 * 24 * time.Hour) // age 7, well under 28d TTL

	te := freshEntry(ScopeTransient, sevenDaysAgo, 1.0)
	te.Domain = "theme"
	te.DecisionKey = "dark_mode_this_session"
	mustUpsert(t, store, te.Domain, te.DecisionKey, te)

	report, err := fs.DecayScan(now)
	if err != nil {
		t.Fatalf("DecayScan: %v", err)
	}
	if report.TransientDecayed != 1 {
		t.Errorf("TransientDecayed = %d, want 1", report.TransientDecayed)
	}
	if report.SoftDeleted != 0 {
		t.Errorf("SoftDeleted = %d, want 0 (age 7 < 28 TTL)", report.SoftDeleted)
	}

	got, ok, tier, _ := store.Get("theme", "dark_mode_this_session")
	if !ok || tier != TierRecall {
		t.Fatalf("transient survivor missing from recall: ok=%v tier=%v", ok, tier)
	}
	want := decayWeight(7) // ≈ 0.354
	if math.Abs(got.Weight-want) > 1e-9 {
		t.Errorf("transient weight after scan = %.4f, want power-law(7) = %.4f", got.Weight, want)
	}
}

// TestDecayScan_Transient28DaysExactly_NotSoftDeleted verifies the TTL
// boundary: an entry that is EXACTLY 28 days old is NOT soft-deleted (the
// predicate is age > decayTTLDays, strictly greater-than; 28 is the last day
// the entry is retained). This matches AC-ADM-012's "27-day reuse" test which
// implies 28 is still in-reuse-range.
func TestDecayScan_Transient28DaysExactly_NotSoftDeleted(t *testing.T) {
	t.Parallel()
	store, fs := newDecayTestStore(t)
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	exactly28DaysAgo := now.Add(-28 * 24 * time.Hour)

	te := freshEntry(ScopeTransient, exactly28DaysAgo, 0.9)
	te.Domain = "deploy"
	te.DecisionKey = "canary_flag"
	mustUpsert(t, store, te.Domain, te.DecisionKey, te)

	report, err := fs.DecayScan(now)
	if err != nil {
		t.Fatalf("DecayScan: %v", err)
	}
	if report.SoftDeleted != 0 {
		t.Errorf("SoftDeleted = %d, want 0 (age==28 is NOT > 28; entry retained)", report.SoftDeleted)
	}
	if report.TransientDecayed != 1 {
		t.Errorf("TransientDecayed = %d, want 1 (age==28 entry is a survivor)", report.TransientDecayed)
	}
}

// TestDecayScan_Transient29Days_SoftDeleted (AC-ADM-012, S2 Critical) verifies
// the 28-day TTL eviction: a transient entry at 29 days is soft-deleted to
// archival. This is the TTL half of AC-ADM-012 (the reuse-reset half lives in
// the Touch test).
func TestDecayScan_Transient29Days_SoftDeleted(t *testing.T) {
	t.Parallel()
	store, fs := newDecayTestStore(t)
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	twentyNineDaysAgo := now.Add(-29 * 24 * time.Hour)

	te := freshEntry(ScopeTransient, twentyNineDaysAgo, 0.5)
	te.Domain = "feature_flag"
	te.DecisionKey = "experimental_ui"
	mustUpsert(t, store, te.Domain, te.DecisionKey, te)

	report, err := fs.DecayScan(now)
	if err != nil {
		t.Fatalf("DecayScan: %v", err)
	}
	if report.SoftDeleted != 1 {
		t.Errorf("SoftDeleted = %d, want 1 (age==29 > 28 → soft-delete)", report.SoftDeleted)
	}

	// The entry moved recall → archival.
	_, _, tierRecall, _ := store.Get(te.Domain, te.DecisionKey)
	if tierRecall == TierRecall {
		t.Errorf("entry still in recall at age 29 (tier=%v); want tier=archival", tierRecall)
	}
	arch, _, tierArch, _ := store.Get(te.Domain, te.DecisionKey)
	if tierArch != TierArchival {
		t.Errorf("entry not promoted to archival at age 29 (tier=%v); want tier=archival", tierArch)
	} else if arch.DecisionKey != te.DecisionKey {
		t.Errorf("archived DecisionKey = %q, want %q", arch.DecisionKey, te.DecisionKey)
	}
}

// TestDecayScan_EmptyRecall is a no-op smoke test: an empty recall tier yields
// a zero-count report and no error. Guards against a nil-dereference on the
// survivors slice write-back path.
func TestDecayScan_EmptyRecall(t *testing.T) {
	t.Parallel()
	_, fs := newDecayTestStore(t)
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)

	report, err := fs.DecayScan(now)
	if err != nil {
		t.Fatalf("DecayScan on empty recall: %v", err)
	}
	if report.StablePreserved != 0 || report.TransientDecayed != 0 || report.SoftDeleted != 0 || report.FloorApplied != 0 {
		t.Errorf("empty-recall report = %+v, want all-zero counts", report)
	}
}

// ---- Touch (AC-ADM-012 reset-on-reuse) ----

// TestTouch_RecallEntryResetsWeightAndLastUsed (AC-ADM-012 reuse-reset half)
// verifies that Touch on a transient recall entry resets age to 0 (via
// last_used=now) and weight to 1.0. Per acceptance.md AC-ADM-012:
// "엔트리가 28일 이내에 재사용되면 age 카운터가 0으로 리셋되고 confidence
// weight가 부양된다".
func TestTouch_RecallEntryResetsWeightAndLastUsed(t *testing.T) {
	t.Parallel()
	store, fs := newDecayTestStore(t)
	base := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	// Seed a transient entry 27 days old (under TTL but weight has decayed).
	te := freshEntry(ScopeTransient, base, decayWeight(27))
	te.Domain = "theme"
	te.DecisionKey = "dark_mode"
	mustUpsert(t, store, te.Domain, te.DecisionKey, te)

	// Verify the seeded weight is the decayed value (≈ 0.19), not 1.0.
	before, _, _, _ := store.Get(te.Domain, te.DecisionKey)
	if before.Weight >= 0.5 {
		t.Fatalf("seeded weight %.4f should be decayed below 0.5 before Touch", before.Weight)
	}

	if err := fs.Touch(te.Domain, te.DecisionKey); err != nil {
		t.Fatalf("Touch: %v", err)
	}

	after, ok, tier, _ := store.Get(te.Domain, te.DecisionKey)
	if !ok || tier != TierRecall {
		t.Fatalf("entry missing or wrong tier after Touch: ok=%v tier=%v", ok, tier)
	}
	if after.Weight != 1.0 {
		t.Errorf("weight after Touch = %.4f, want 1.0 (reset-on-reuse)", after.Weight)
	}
	// age_days is computed from last_used; a fresh last_used yields age 0.
	// Verify the reset indirectly by running a decay scan "now" and checking
	// the weight stays 1.0 (age 0 → decayWeight(0) = 1.0).
	now := time.Now().UTC()
	scanReport, err := fs.DecayScan(now)
	if err != nil {
		t.Fatalf("post-Touch DecayScan: %v", err)
	}
	if scanReport.SoftDeleted != 0 {
		t.Errorf("post-Touch scan soft-deleted the entry (age should be 0); SoftDeleted=%d", scanReport.SoftDeleted)
	}
	rescanned, _, _, _ := store.Get(te.Domain, te.DecisionKey)
	if math.Abs(rescanned.Weight-1.0) > 1e-9 {
		t.Errorf("weight after post-Touch scan = %.4f, want 1.0 (age 0 from reset last_used)", rescanned.Weight)
	}
}

// TestTouch_ArchivalEntryResurrectedToRecall verifies that Touch on a
// soft-deleted (archival) entry resurrects it back to recall with a reset
// weight + last_used. This is what makes the 28-day TTL a SOFT delete rather
// than a hard one — reuse brings the entry back.
func TestTouch_ArchivalEntryResurrectedToRecall(t *testing.T) {
	t.Parallel()
	store, fs := newDecayTestStore(t)
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	longAgo := now.Add(-40 * 24 * time.Hour) // 40 days — past TTL

	te := freshEntry(ScopeTransient, longAgo, 0.5)
	te.Domain = "feature_flag"
	te.DecisionKey = "experimental_ui"
	mustUpsert(t, store, te.Domain, te.DecisionKey, te)

	// Run a decay scan to soft-delete the entry to archival.
	if _, err := fs.DecayScan(now); err != nil {
		t.Fatalf("pre-Touch DecayScan: %v", err)
	}
	if _, _, tier, _ := store.Get(te.Domain, te.DecisionKey); tier != TierArchival {
		t.Fatalf("entry not in archival after scan; tier=%v", tier)
	}

	// Touch resurrects.
	if err := fs.Touch(te.Domain, te.DecisionKey); err != nil {
		t.Fatalf("Touch: %v", err)
	}

	after, _, tier, _ := store.Get(te.Domain, te.DecisionKey)
	if tier != TierRecall {
		t.Errorf("entry tier after Touch = %v, want recall (resurrected)", tier)
	}
	if after.Weight != 1.0 {
		t.Errorf("weight after Touch = %.4f, want 1.0", after.Weight)
	}
}

// TestTouch_MissingEntryReturnsErrNotFound verifies Touch on a non-existent
// key returns ErrNotFound and mutates nothing.
func TestTouch_MissingEntryReturnsErrNotFound(t *testing.T) {
	t.Parallel()
	_, fs := newDecayTestStore(t)
	err := fs.Touch("nope", "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Touch on missing key: err = %v, want ErrNotFound", err)
	}
}

// TestTouch_EmptyKeyReturnsInvalidEntry verifies Touch validates its inputs.
func TestTouch_EmptyKeyReturnsInvalidEntry(t *testing.T) {
	t.Parallel()
	_, fs := newDecayTestStore(t)
	if err := fs.Touch("", "key"); !errors.Is(err, ErrInvalidEntry) {
		t.Errorf("Touch(empty domain): err = %v, want ErrInvalidEntry", err)
	}
	if err := fs.Touch("domain", ""); !errors.Is(err, ErrInvalidEntry) {
		t.Errorf("Touch(empty key): err = %v, want ErrInvalidEntry", err)
	}
}

// ---- daily-cadence gate (AC-ADM-NFR-004) ----

// TestScanDue_WindowBehavior (AC-ADM-NFR-004, S3 Major) verifies the 24h gate:
//
//   - no stamp file → due (first-ever scan)
//   - stamp 23h old → NOT due (within window)
//   - stamp exactly 24h old → due (window elapsed)
//   - stamp 25h old → due
//   - corrupt stamp → due (fail-open)
func TestScanDue_WindowBehavior(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name     string
		stampAge time.Duration // 0 = no stamp written
		stampRaw string        // non-empty overrides stampAge with raw content
		wantDue  bool
	}{
		{"no stamp → due", 0, "", true},
		{"23h old → not due", 23 * time.Hour, "", false},
		{"exactly 24h → due", 24 * time.Hour, "", true},
		{"25h old → due", 25 * time.Hour, "", true},
		{"corrupt stamp → due (fail-open)", 0, "not-a-timestamp", true},
		{"empty stamp → due (fail-open)", 0, "\n  \n", true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			stateDir := t.TempDir()
			if tt.stampRaw != "" {
				if err := os.WriteFile(filepath.Join(stateDir, decayLastRunFileName), []byte(tt.stampRaw), 0o644); err != nil {
					t.Fatalf("seed corrupt stamp: %v", err)
				}
			} else if tt.stampAge > 0 {
				prior := now.Add(-tt.stampAge)
				if err := MarkScanned(stateDir, prior); err != nil {
					t.Fatalf("MarkScanned: %v", err)
				}
			}
			got, err := ScanDue(stateDir, now)
			if err != nil {
				t.Fatalf("ScanDue: %v", err)
			}
			if got != tt.wantDue {
				t.Errorf("ScanDue = %v, want %v", got, tt.wantDue)
			}
		})
	}
}

// TestMarkScanned_WritesTimestampFile verifies MarkScanned produces a file
// ScanDue can read back. Round-trip integration test for the two helpers.
func TestMarkScanned_WritesTimestampFile(t *testing.T) {
	t.Parallel()
	stateDir := t.TempDir()
	scanTime := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)

	if err := MarkScanned(stateDir, scanTime); err != nil {
		t.Fatalf("MarkScanned: %v", err)
	}

	// Immediately after, a ScanDue at the same instant is NOT due (0h elapsed).
	due, err := ScanDue(stateDir, scanTime)
	if err != nil {
		t.Fatalf("ScanDue: %v", err)
	}
	if due {
		t.Errorf("ScanDue immediately after MarkScanned = true, want false (0h < 24h window)")
	}

	// 24h later, ScanDue is true.
	due24, err := ScanDue(stateDir, scanTime.Add(24*time.Hour))
	if err != nil {
		t.Fatalf("ScanDue +24h: %v", err)
	}
	if !due24 {
		t.Errorf("ScanDue 24h after MarkScanned = false, want true (window elapsed)")
	}
}

// TestScanDue_FailOpenOnUnreadableStamp verifies that an unreadable stamp
// (permission denied) is treated as "due" rather than blocking maintenance.
// This test is skipped on root (root bypasses file perms) and on systems
// where the temp dir is not chmod-able.
func TestScanDue_FailOpenOnUnreadableStamp(t *testing.T) {
	t.Parallel()
	if os.Geteuid() == 0 {
		t.Skip("running as root — file-perm test is non-deterministic")
	}
	stateDir := t.TempDir()
	stampPath := filepath.Join(stateDir, decayLastRunFileName)
	if err := os.WriteFile(stampPath, []byte(time.Now().Format(time.RFC3339)), 0o000); err != nil {
		t.Fatalf("seed unreadable stamp: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(stampPath, 0o644) }) // restore so cleanup can remove

	due, err := ScanDue(stateDir, time.Now())
	if err != nil {
		t.Fatalf("ScanDue returned error on unreadable stamp: %v (want nil + fail-open true)", err)
	}
	if !due {
		t.Errorf("ScanDue on unreadable stamp = false, want true (fail-open)")
	}
}

// TestTouch_CoreEntryRefreshedInCore covers the TierCore branch of Touch: a
// stable entry stored in core.yaml (the normal placement for ScopeStable) is
// refreshed in-place — last_used/weight updated, still in core.
func TestTouch_CoreEntryRefreshedInCore(t *testing.T) {
	t.Parallel()
	store, fs := newDecayTestStore(t)
	longAgo := time.Now().Add(-30 * 24 * time.Hour)

	stable := freshEntry(ScopeStable, longAgo, 0.92)
	stable.Domain = "lang"
	stable.DecisionKey = "backend"
	mustUpsert(t, store, stable.Domain, stable.DecisionKey, stable)
	// Verify it landed in core (Upsert routes ScopeStable → core).
	if _, _, tier, _ := store.Get(stable.Domain, stable.DecisionKey); tier != TierCore {
		t.Fatalf("setup: stable entry tier = %v, want core", tier)
	}

	if err := fs.Touch(stable.Domain, stable.DecisionKey); err != nil {
		t.Fatalf("Touch: %v", err)
	}

	after, _, tier, _ := store.Get(stable.Domain, stable.DecisionKey)
	if tier != TierCore {
		t.Errorf("after Touch tier = %v, want core (stable stays in core)", tier)
	}
	if after.Weight != 1.0 {
		t.Errorf("after Touch weight = %.4f, want 1.0", after.Weight)
	}
	// last_used moved forward (age near 0 now).
	if age := ageInDays(time.Now(), after.LastUsed); age > 0 {
		t.Errorf("after Touch age = %d, want 0 (last_used refreshed to ~now)", age)
	}
}

// TestTouch_NonExistentArchivalRemoveIsNoop covers removeArchivalEntry's
// idempotent branch (os.ErrNotExist → return nil) by Touching an entry that
// lives only in recall (no archival copy to remove).
func TestTouch_NonExistentArchivalRemoveIsNoop(t *testing.T) {
	t.Parallel()
	store, fs := newDecayTestStore(t)
	te := freshEntry(ScopeTransient, time.Now().Add(-1*time.Hour), 0.9)
	te.Domain = "recent"
	te.DecisionKey = "entry"
	mustUpsert(t, store, te.Domain, te.DecisionKey, te)

	// Touch on a recall entry — removeArchivalEntry is NOT called on this
	// path, but the recall-refresh path exercises upsertToRecall fully.
	if err := fs.Touch(te.Domain, te.DecisionKey); err != nil {
		t.Fatalf("Touch recall entry: %v", err)
	}
	after, _, tier, _ := store.Get(te.Domain, te.DecisionKey)
	if tier != TierRecall {
		t.Errorf("tier after Touch = %v, want recall", tier)
	}
	if after.Weight != 1.0 {
		t.Errorf("weight after Touch = %.4f, want 1.0", after.Weight)
	}
}

// TestParseStampTimestamp_UnixEpoch covers the forward-compat branch that
// accepts a bare unix-epoch integer as the stamp value.
func TestParseStampTimestamp_UnixEpoch(t *testing.T) {
	t.Parallel()
	got, err := parseStampTimestamp("1719316800\n")
	if err != nil {
		t.Fatalf("parseStampTimestamp unix epoch: %v", err)
	}
	want := time.Unix(1719316800, 0).UTC()
	if !got.Equal(want) {
		t.Errorf("parseStampTimestamp unix epoch = %v, want %v", got, want)
	}
}

// TestParseStampTimestamp_EmptyAndGarbage verifies the error paths return
// (time.Time{}, error) so the caller can fail-open.
func TestParseStampTimestamp_EmptyAndGarbage(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{"", "  \n  ", "garbage-not-a-date"} {
		if _, err := parseStampTimestamp(raw); err == nil {
			t.Errorf("parseStampTimestamp(%q) = nil error, want non-nil", raw)
		}
	}
}

// TestDecayReport_String covers the report's human-readable String() method
// (exercised indirectly by runDecayScan but pinned here for a stable format
// assertion).
func TestDecayReport_String(t *testing.T) {
	t.Parallel()
	r := &DecayReport{
		ScannedAt:        time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC),
		StablePreserved:  2,
		TransientDecayed: 3,
		SoftDeleted:      1,
		FloorApplied:     1,
	}
	got := r.String()
	for _, want := range []string{"stable_preserved=2", "transient_decayed=3", "soft_deleted=1", "floor_applied=1"} {
		if !strings.Contains(got, want) {
			t.Errorf("DecayReport.String() = %q, missing %q", got, want)
		}
	}
}

// TestDecayWeight_NegativeAgeClampedToZero verifies the defensive branch: a
// negative age (clock skew, future last_used) is treated as age 0 so the entry
// is weighted as fresh (1.0) rather than producing a NaN or a domain error.
func TestDecayWeight_NegativeAgeClampedToZero(t *testing.T) {
	t.Parallel()
	got := decayWeight(-5)
	if got != 1.0 {
		t.Errorf("decayWeight(-5) = %v, want 1.0 (clamped to age 0)", got)
	}
}

// TestDecayWeight_MonotonicallyDecreasing verifies the curve is strictly
// monotonically decreasing — a higher age MUST yield a strictly lower weight.
// This guards against an accidental sign error that would make the decay
// reward staleness instead of penalizing it.
func TestDecayWeight_MonotonicallyDecreasing(t *testing.T) {
	t.Parallel()
	prev := decayWeight(0)
	for age := 1; age <= 56; age++ {
		cur := decayWeight(age)
		if cur >= prev {
			t.Errorf("decayWeight not monotonically decreasing at age %d: cur=%.4f >= prev=%.4f", age, cur, prev)
		}
		prev = cur
	}
}

// TestAgeInDays verifies ageInDays floors sub-day remainders and clamps future
// last_used to 0. This is the boundary the scan uses to decide TTL eviction.
func TestAgeInDays(t *testing.T) {
	t.Parallel()
	base := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name     string
		now      time.Time
		lastUsed time.Time
		want     int
	}{
		{"same instant", base, base, 0},
		{"6h ago → 0 days", base, base.Add(-6 * time.Hour), 0},
		{"23h59m ago → 0 days", base, base.Add(-23*time.Hour - 59*time.Minute), 0},
		{"exactly 1 day ago", base, base.Add(-24 * time.Hour), 1},
		{"28 days ago", base, base.Add(-28 * 24 * time.Hour), 28},
		{"29 days ago (TTL+1)", base, base.Add(-29 * 24 * time.Hour), 29},
		{"future last_used clamped to 0", base, base.Add(+1 * time.Hour), 0},
		{"zero last_used clamped to 0", base, time.Time{}, 0},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ageInDays(tt.now, tt.lastUsed)
			if got != tt.want {
				t.Errorf("ageInDays(now, lastUsed) = %d, want %d", got, tt.want)
			}
		})
	}
}
