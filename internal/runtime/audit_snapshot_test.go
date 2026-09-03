package runtime

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestStickyCacheSurvivesPast24h verifies AC-AUDIT-SNAPSHOT-001 (A1):
// a cached PASS verdict whose plan-artifact hash is unchanged SHALL remain
// valid regardless of the elapsed time since the verdict was recorded. The
// legacy 24h time-window condition is RETIRED; cache validity is hash-only.
func TestStickyCacheSurvivesPast24h(t *testing.T) {
	t.Parallel()

	specID := "SPEC-AUDIT-SNAPSHOT-001"
	hash := "h1-unchanged"
	recordedAt := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)

	ages := []struct {
		name string
		now  time.Time
	}{
		{"T0+1h (within legacy window)", recordedAt.Add(1 * time.Hour)},
		{"T0+25h (past legacy 24h window)", recordedAt.Add(25 * time.Hour)},
		{"T0+30d (long past legacy window)", recordedAt.Add(30 * 24 * time.Hour)},
	}

	for _, tt := range ages {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cache := NewInMemoryCache()
			cache.Store(specID, hash, &AuditResult{
				Verdict:        VerdictPass,
				AuditAt:        recordedAt,
				AuditorVersion: "plan-auditor/v1",
			})

			entry, hit := cache.Lookup(specID, hash, tt.now)
			if !hit {
				t.Fatalf("sticky cache MUST hit at %s (hash unchanged) — the 24h time-window condition is retired", tt.name)
			}
			if entry == nil || entry.PlanArtifactHash != hash {
				t.Fatalf("hit returned wrong entry: %+v", entry)
			}
		})
	}
}

// TestStickyCacheStillInvalidatesOnHashChange verifies AC-AUDIT-SNAPSHOT-001's
// invariant complement: dropping the time-window condition does NOT weaken the
// hash-match invariant. A changed plan-artifact hash still yields a miss.
func TestStickyCacheStillInvalidatesOnHashChange(t *testing.T) {
	t.Parallel()

	cache := NewInMemoryCache()
	specID := "SPEC-AUDIT-SNAPSHOT-001b"
	hashBefore := "hash-before"
	hashAfter := "hash-after"
	t0 := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)

	cache.Store(specID, hashBefore, &AuditResult{Verdict: VerdictPass, AuditAt: t0})

	if _, hit := cache.Lookup(specID, hashBefore, t0.Add(30*24*time.Hour)); !hit {
		t.Error("sticky cache should still hit on hash match even far in the future")
	}
	if _, hit := cache.Lookup(specID, hashAfter, t0.Add(1*time.Hour)); hit {
		t.Error("sticky cache must still MISS on hash change (the hash invariant is preserved)")
	}
}

// TestTierLHashIncludesDesignAndResearch verifies AC-AUDIT-SNAPSHOT-001b (A1 Tier L):
// for a Tier L SPEC directory containing design.md and research.md, ComputeHash
// includes both files — mutating either changes the hash. The Tier L extension
// is implemented via the "skip if missing" rule: design.md/research.md are in
// the subject list and contribute only when present (Tier S/M dirs lack them).
func TestTierLHashIncludesDesignAndResearch(t *testing.T) {
	t.Parallel()

	base := func() string {
		dir := t.TempDir()
		for _, name := range []string{"spec.md", "plan.md", "acceptance.md", "design.md", "research.md"} {
			if err := os.WriteFile(filepath.Join(dir, name), []byte("initial "+name), 0o644); err != nil {
				t.Fatalf("WriteFile %s: %v", name, err)
			}
		}
		return dir
	}

	cache := NewInMemoryCache()

	hashBefore, err := cache.ComputeHash(base())
	if err != nil {
		t.Fatalf("ComputeHash baseline: %v", err)
	}

	// Mutate design.md → hash must change.
	mutatedDesign := base()
	if err := os.WriteFile(filepath.Join(mutatedDesign, "design.md"), []byte("mutated design content"), 0o644); err != nil {
		t.Fatalf("WriteFile design.md: %v", err)
	}
	hashAfterDesign, err := cache.ComputeHash(mutatedDesign)
	if err != nil {
		t.Fatalf("ComputeHash after design mutation: %v", err)
	}
	if hashAfterDesign == hashBefore {
		t.Error("mutating design.md did NOT change the Tier L hash — design.md must be a hash subject for Tier L")
	}

	// Mutate research.md → hash must change.
	mutatedResearch := base()
	if err := os.WriteFile(filepath.Join(mutatedResearch, "research.md"), []byte("mutated research content"), 0o644); err != nil {
		t.Fatalf("WriteFile research.md: %v", err)
	}
	hashAfterResearch, err := cache.ComputeHash(mutatedResearch)
	if err != nil {
		t.Fatalf("ComputeHash after research mutation: %v", err)
	}
	if hashAfterResearch == hashBefore {
		t.Error("mutating research.md did NOT change the Tier L hash — research.md must be a hash subject for Tier L")
	}
}

// TestTasksMdStillHashSubjectForGrandfathered verifies AC-AUDIT-SNAPSHOT-001b's
// backward-compat clause (K-2): tasks.md is retained in the subject set so
// grandfathered V3R4-era SPECs keep their hash behavior.
func TestTasksMdStillHashSubjectForGrandfathered(t *testing.T) {
	t.Parallel()

	base := func() string {
		dir := t.TempDir()
		for _, name := range []string{"spec.md", "plan.md", "tasks.md"} {
			if err := os.WriteFile(filepath.Join(dir, name), []byte("initial "+name), 0o644); err != nil {
				t.Fatalf("WriteFile %s: %v", name, err)
			}
		}
		return dir
	}

	cache := NewInMemoryCache()
	hashBefore, err := cache.ComputeHash(base())
	if err != nil {
		t.Fatalf("ComputeHash baseline: %v", err)
	}

	mutated := base()
	if err := os.WriteFile(filepath.Join(mutated, "tasks.md"), []byte("mutated tasks content"), 0o644); err != nil {
		t.Fatalf("WriteFile tasks.md: %v", err)
	}
	hashAfter, err := cache.ComputeHash(mutated)
	if err != nil {
		t.Fatalf("ComputeHash after tasks mutation: %v", err)
	}
	if hashAfter == hashBefore {
		t.Error("mutating tasks.md did NOT change the hash — tasks.md must remain a subject for grandfathered SPECs")
	}
}

// TestSkipEligibleByScorePerTier verifies AC-AUDIT-SNAPSHOT-002 (A2):
// the plan-audit skip-eligible score threshold equals the SPEC's per-tier PASS
// threshold (S 0.75 / M 0.80 / L 0.85), NOT the retired flat 0.90. A SPEC whose
// plan-phase audit verdict legitimately PASSED is skip-eligible by default.
func TestSkipEligibleByScorePerTier(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		tier        string
		score       float64
		eligible    bool
		noteFlat090 string
	}{
		// Per-tier PASS threshold alignment.
		{"Tier S 0.78 (>=0.75) eligible", "S", 0.78, true, "0.78 < 0.90 — eligible despite being below the retired flat 0.90"},
		{"Tier S 0.74 (<0.75) NOT eligible", "S", 0.74, false, ""},
		{"Tier M 0.81 (>=0.80) eligible", "M", 0.81, true, "0.81 < 0.90 — eligible despite being below the retired flat 0.90"},
		{"Tier M 0.78 (<0.80) NOT eligible", "M", 0.78, false, "0.78 against Tier M is below the 0.80 PASS threshold"},
		{"Tier L 0.86 (>=0.85) eligible", "L", 0.86, true, "0.86 < 0.90 — eligible despite being below the retired flat 0.90"},
		{"Tier L 0.82 (<0.85) NOT eligible", "L", 0.82, false, ""},
		// The retired flat 0.90 predicate MUST NOT be consulted: a 0.81 Tier M
		// SPEC is eligible even though 0.81 < 0.90.
		{"Tier M 0.81 eligible (flat 0.90 retired)", "M", 0.81, true, "explicit assertion: the flat >=0.90 predicate is retired"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := SkipEligibleByScore(tc.tier, tc.score)
			if got != tc.eligible {
				t.Errorf("SkipEligibleByScore(tier=%q, score=%.2f) = %v, want %v. %s",
					tc.tier, tc.score, got, tc.eligible, tc.noteFlat090)
			}
		})
	}
}

// TestSkipEligibleByScoreUnknownTierIsStrict verifies that an unrecognized tier
// (empty / typo) falls back to the strictest Tier L threshold (0.85), never to
// permissive behavior. Backward compat: a tier-absent SPEC is treated as Tier L
// per spec-frontmatter-schema.md § Optional Fields.
func TestSkipEligibleByScoreUnknownTierIsStrict(t *testing.T) {
	t.Parallel()

	// Unknown/empty tier → strict (Tier L) threshold.
	if !SkipEligibleByScore("", 0.86) {
		t.Error("empty tier should fall back to Tier L threshold (0.86 >= 0.85 → eligible)")
	}
	if SkipEligibleByScore("", 0.80) {
		t.Error("empty tier should fall back to Tier L threshold (0.80 < 0.85 → NOT eligible)")
	}
}
