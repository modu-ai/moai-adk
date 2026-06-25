// Package preference implements the AskUserQuestion user-decision memory layer
// for SPEC-V3R6-ASKUSER-DECISION-MEMORY-001.
//
// decay.go implements the M4 power-law decay policy + 28-day TTL + daily scan
// cadence (REQ-ADM-011, REQ-ADM-012, NFR-ADM-004).
//
// The weight function (decayWeight) is PURE: it depends only on age_days and
// the package-level alpha constant. It MUST NOT call time.Now() — the scan
// orchestration layer injects `now time.Time` so the function is reproducible
// in tests and deterministic across runs.

package preference

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// decayAlpha is the fixed power-law exponent for Standard tier (design.md §A.3
// + §E.1). α=0.5 is "square root decay" — weight ≈ 0.19 at 28 days, the
// effective threshold for soft-delete. Dynamic per-user α estimation is
// deferred to the "complete" tier (spec.md §E Out of Scope) and MUST NOT be
// introduced as a tunable here.
//
// @MX:NOTE: [AUTO] α=0.5 고정 — Standard tier 단일 값. 동적 학습은 complete tier 이월.
// @MX:SPEC: SPEC-V3R6-ASKUSER-DECISION-MEMORY-001 design.md §A.3/§E.1
const decayAlpha = 0.5

// decayTTLDays is the transient-entry soft-delete threshold. A transient entry
// whose last_used is older than this many days is moved from recall to archival
// (NOT hard-deleted) by DecayScan (REQ-ADM-012, AC-ADM-012).
const decayTTLDays = 28

// decayScanIntervalHours is the daily-cadence gate for the background decay
// scan (NFR-ADM-004, AC-ADM-NFR-004). Per-turn decay is too costly; the scan
// runs at most once per 24h, gated by a timestamp file.
const decayScanIntervalHours = 24

// stableWeightFloor is the minimum weight a stable-scope entry retains under
// pure time-decay exemption (design.md §E.2: `max(weight, 0.5)`). Without this
// floor, a stable entry that has not been refreshed for a long time would
// decay toward zero and reproduce Koren's "지속 신호 상실" (AP-ADM-006).
const stableWeightFloor = 0.5

// NOTE: design.md §E.2 also names a per-refresh boost (`min(1.0, weight + 0.1)`)
// for stable entries. That incremental boost is superseded by §E.3's stronger
// reset-on-reuse semantic: Touch sets weight to 1.0 outright (the fresh-entry
// baseline), which is the upper bound the §E.2 boost was clamping toward.
// A separate reuseWeightBoost constant is therefore not needed and would be
// dead code; the floor (stableWeightFloor) handles the decay-floor half of
// the §E.2 contract.

// decayWeight is the pure power-law weight function (design.md §E.1):
//
//	weight(ageDays) = (ageDays + 1)^(-α)     where α = decayAlpha (0.5)
//
// age_days 0 → 1.000, 1 → 0.707, 7 → 0.378, 14 → 0.258, 28 → 0.189, 56 → 0.133.
// The function is pure: it does NOT read wall-clock time. Callers compute
// age_days from (now - entry.LastUsed) and pass the integer here.
func decayWeight(ageDays int) float64 {
	if ageDays < 0 {
		// Defensive: a negative age (clock skew, future last_used) is treated
		// as age 0 so the entry is weighted as fresh rather than rejected.
		ageDays = 0
	}
	return math.Pow(float64(ageDays)+1.0, -decayAlpha)
}

// ageInDays returns the integer number of days between lastUsed and now. A
// negative result (future lastUsed) is clamped to 0 by decayWeight. Sub-day
// remainders are floored — a 6h-old entry is age 0 days; a 23h59m-old entry is
// also age 0 days. This matches the design.md §E.1 table where age is a
// whole-day count.
func ageInDays(now, lastUsed time.Time) int {
	if lastUsed.IsZero() || now.Before(lastUsed) {
		return 0
	}
	d := now.Sub(lastUsed)
	return int(d.Hours() / 24)
}

// DecayReport summarizes a single DecayScan run for diagnostics + CLI output.
// The counts let the `moai preference decay-scan` subcommand print a
// machine-readable summary (JSON with --json, plain counts otherwise).
type DecayReport struct {
	ScannedAt        time.Time `json:"scanned_at"        yaml:"scanned_at"`
	StablePreserved  int       `json:"stable_preserved"  yaml:"stable_preserved"`
	TransientDecayed int       `json:"transient_decayed" yaml:"transient_decayed"`
	SoftDeleted      int       `json:"soft_deleted"      yaml:"soft_deleted"`
	FloorApplied     int       `json:"floor_applied"     yaml:"floor_applied"`
}

// String renders a human-readable one-line summary for the CLI default
// (non-JSON) output path.
func (r *DecayReport) String() string {
	return fmt.Sprintf(
		"decay scan at %s: stable_preserved=%d transient_decayed=%d soft_deleted=%d floor_applied=%d",
		r.ScannedAt.Format(time.RFC3339),
		r.StablePreserved, r.TransientDecayed, r.SoftDeleted, r.FloorApplied,
	)
}

// DecayScan executes one decay-policy pass over the recall tier at the given
// wall-clock time (REQ-ADM-011, REQ-ADM-012, NFR-ADM-004).
//
// The scan is the M4 maintenance routine the filestore.go line 101-102 /
// 408-409 comments reserved the hook points for. It:
//
//  1. Loads recall.jsonl.
//  2. For each entry, separates stable from transient behavior:
//     - STABLE entries are EXEMPT from pure time-decay (REQ-ADM-011, the
//       anti-AP-ADM-006 / Koren "지속 신호 상실" invariant). Their weight is
//       floored at stableWeightFloor (design.md §E.2: max(weight, 0.5)); the
//       floor is recorded in the report only when it actually lifts the
//       stored weight. Stable entries are NEVER soft-deleted by age.
//     - TRANSIENT entries get full power-law decay. An entry whose age
//       (now - LastUsed, in whole days) exceeds decayTTLDays (28) is
//       SOFT-DELETED: moved from recall to archival via writeArchivalEntry +
//       dropped from the recall write-back. Age <= 28 entries get their
//       weight refreshed to decayWeight(age) and are kept in recall.
//  3. Writes the surviving recall entries (stable + non-expired transient)
//     back via writeRecall (atomic).
//
// The scan does NOT touch core.yaml — stable entries promoted to core are
// already exempt (core is the always-loaded tier; decay is a recall-tier
// concern). Archival entries are write-once-per-key; a soft-deleted entry
// overwrites any prior archival copy for the same key.
//
// `now` is injected (not time.Now()) so the scan is deterministic in tests
// and reproducible across runs (per the pure-function discipline in the
// package doc).
//
// @MX:ANCHOR: [AUTO] DecayScan — recall-tier decay policy entry point
// @MX:REASON: fan_in >= 3 예상 (CLI subcommand, SessionStart hook, future tests); stable/transient 분리가 핵심 invariant
// @MX:SPEC: SPEC-V3R6-ASKUSER-DECISION-MEMORY-001 REQ-ADM-011/012, design.md §E.2
func (s *fileStore) DecayScan(now time.Time) (*DecayReport, error) {
	report := &DecayReport{ScannedAt: now}

	entries, err := s.loadRecall()
	if err != nil {
		return nil, fmt.Errorf("preference: decay scan load recall: %w", err)
	}

	var survivors []Entry
	for _, e := range entries {
		age := ageInDays(now, e.LastUsed)

		if e.Scope == ScopeStable {
			// Stable exemption (REQ-ADM-011, design.md §E.2). Floor the weight
			// so a long-unrefreshed stable entry does not decay toward zero.
			if e.Weight < stableWeightFloor {
				e.Weight = stableWeightFloor
				report.FloorApplied++
			}
			report.StablePreserved++
			survivors = append(survivors, e)
			continue
		}

		// Transient: apply TTL eviction first (age > 28 → soft-delete to
		// archival), then refresh weight on the survivors.
		if age > decayTTLDays {
			if err := s.writeArchivalEntry(e.Domain, e.DecisionKey, e); err != nil {
				return nil, fmt.Errorf("preference: decay scan soft-delete %s/%s: %w", e.Domain, e.DecisionKey, err)
			}
			report.SoftDeleted++
			continue
		}

		// Transient survivor: refresh weight to the power-law value for the
		// current age. This keeps recall-tier weights accurate between TTL
		// evictions so core-tier eviction (upsertToCore demotion) uses fresh
		// numbers.
		e.Weight = decayWeight(age)
		report.TransientDecayed++
		survivors = append(survivors, e)
	}

	if err := s.writeRecall(survivors); err != nil {
		return nil, fmt.Errorf("preference: decay scan write recall: %w", err)
	}
	return report, nil
}

// Touch refreshes an entry on reuse (REQ-ADM-012 "사용시 리셋", design.md §E.3).
// It is the reset-on-reuse half of AC-ADM-012 — the other half (28-day TTL
// eviction) lives in DecayScan.
//
// Touch looks up the entry across all tiers (core → recall → archival). On a
// hit it:
//
//   - sets LastUsed = now (so age_days resets to 0)
//   - sets Weight = 1.0 (fresh-entry baseline; design.md §E.3)
//   - leaves Confidence unchanged: an observed entry is already the strongest
//     signal; an inferred entry stays inferred because a Touch (reuse) is
//     evidence of RELEVANCE, not evidence of a fresh user OBSERVATION. The
//     weight reset to 1.0 IS the boost. design.md §E.3 names a "소폭 부양"
//     for inferred confidence, but the Confidence enum is a closed 2-value
//     set (observed | inferred) — introducing a third value or silently
//     flipping inferred→observed would misrepresent the signal provenance and
//     violate REQ-ADM-018 (verification-claim integrity). The honest
//     implementation boosts weight only.
//
// On a miss (ErrNotFound) Touch returns ErrNotFound and mutates nothing.
//
// Touch is preferred over mutating Get because Get is called frequently and a
// silent write on every read is surprising (per the M4 spawn-prompt design
// guidance + AC-ADM-012 test plan).
//
// @MX:NOTE: [AUTO] Touch는 Get과 분리 — 빈번한 Get에서 silent write 방지
// @MX:SPEC: SPEC-V3R6-ASKUSER-DECISION-MEMORY-001 REQ-ADM-012, design.md §E.3
func (s *fileStore) Touch(domain, decisionKey string) error {
	if domain == "" || decisionKey == "" {
		return fmt.Errorf("%w: empty domain or decision_key", ErrInvalidEntry)
	}
	now := time.Now().UTC()

	entry, tier, err := s.findEntryForTouch(domain, decisionKey)
	if err != nil {
		return err
	}

	entry.LastUsed = now
	entry.Weight = 1.0

	// Re-upsert into the SAME tier the entry was found in. A core entry stays
	// in core (it may re-trigger 4KB-cap demotion of another entry, which is
	// correct — reuse should promote relevance). A recall entry stays in
	// recall. An archival entry that has been Touch'd is RESURRECTED: moved
	// back to recall (design.md §E.3 "회수 시 리셋" implies a soft-deleted
	// entry that is reused comes back from archival — otherwise the TTL
	// soft-delete would be indistinguishable from a hard delete from the
	// user's perspective).
	switch tier {
	case TierCore:
		if err := s.upsertToCore(domain, decisionKey, entry); err != nil {
			return fmt.Errorf("preference: touch (core re-upsert): %w", err)
		}
	case TierArchival:
		// Resurrect: remove from archival, upsert to recall.
		if err := s.removeArchivalEntry(domain, decisionKey); err != nil {
			return fmt.Errorf("preference: touch (archival remove): %w", err)
		}
		if err := s.upsertToRecall(domain, decisionKey, entry); err != nil {
			return fmt.Errorf("preference: touch (recall resurrect): %w", err)
		}
	default: // TierRecall
		if err := s.upsertToRecall(domain, decisionKey, entry); err != nil {
			return fmt.Errorf("preference: touch (recall refresh): %w", err)
		}
	}
	return nil
}

// findEntryForTouch locates an entry across the cascade without the short-
// circuit-on-core-hit semantic of Get (Touch needs to know which tier the
// entry lives in so it can re-upsert into the same tier). Returns the entry,
// the tier where it was found, or ErrNotFound if no tier has it.
func (s *fileStore) findEntryForTouch(domain, decisionKey string) (Entry, Tier, error) {
	if e, ok, err := s.getFromCore(domain, decisionKey); err != nil {
		return Entry{}, TierNone, fmt.Errorf("preference: touch (core read): %w", err)
	} else if ok {
		return e, TierCore, nil
	}
	if e, ok, err := s.getFromRecall(domain, decisionKey); err != nil {
		return Entry{}, TierNone, fmt.Errorf("preference: touch (recall read): %w", err)
	} else if ok {
		return e, TierRecall, nil
	}
	if e, ok, err := s.getFromArchival(domain, decisionKey); err != nil {
		return Entry{}, TierNone, fmt.Errorf("preference: touch (archival read): %w", err)
	} else if ok {
		return e, TierArchival, nil
	}
	return Entry{}, TierNone, ErrNotFound
}

// removeArchivalEntry deletes the single archival JSON file for (domain,
// decisionKey). It is the inverse of writeArchivalEntry and is used by Touch
// when resurrecting a soft-deleted entry back to recall.
func (s *fileStore) removeArchivalEntry(domain, decisionKey string) error {
	p := s.archivalEntryPath(domain, decisionKey)
	if err := os.Remove(p); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil // already gone — idempotent
		}
		return fmt.Errorf("remove archival %s: %w", p, err)
	}
	return nil
}

// ---- daily-cadence gate (NFR-ADM-004, AC-ADM-NFR-004) ----

// decayLastRunFileName is the timestamp-file name under the project's .moai/
// state dir that records when DecayScan last ran. The daily-cadence gate
// reads this file; if it is < decayScanIntervalHours old, the scan is skipped
// (returns false from ScanDue). This keeps the scan at exactly 1-in-24h even
// if both the CLI subcommand and the SessionStart hook branch try to run it.
const decayLastRunFileName = "preference-decay-last-run"

// ScanDue reports whether a decay scan should run now, based on the timestamp
// file at stateDir/decayLastRunFileName. Returns (true, nil) when no prior
// timestamp exists (first-ever scan) OR the elapsed time since the last scan
// is >= decayScanIntervalHours. Returns (false, nil) when the scan ran within
// the window and should be skipped. An unreadable/corrupt timestamp file is
// treated as "no prior scan" (true, nil) — fail-open to avoid stalling the
// maintenance routine on a malformed stamp.
//
// `now` is injected for deterministic testing; production callers pass
// time.Now().
func ScanDue(stateDir string, now time.Time) (bool, error) {
	stampPath := filepath.Join(stateDir, decayLastRunFileName)
	data, err := os.ReadFile(stampPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return true, nil // first-ever scan
		}
		// Unreadable stamp → fail-open (run the scan). A corrupt stamp should
		// not permanently disable maintenance; the scan itself will overwrite
		// the stamp with a fresh value.
		return true, nil
	}
	prior, pErr := parseStampTimestamp(string(data))
	if pErr != nil {
		return true, nil // corrupt stamp → fail-open
	}
	elapsed := now.Sub(prior)
	return elapsed >= time.Duration(decayScanIntervalHours)*time.Hour, nil
}

// MarkScanned writes the timestamp file recording that a scan ran at `now`.
// Called by DecayScan's caller (the CLI subcommand or SessionStart branch)
// AFTER a successful scan. The stamp is an RFC3339 timestamp on a single line.
func MarkScanned(stateDir string, now time.Time) error {
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return fmt.Errorf("preference: create state dir for decay stamp: %w", err)
	}
	stampPath := filepath.Join(stateDir, decayLastRunFileName)
	content := now.UTC().Format(time.RFC3339) + "\n"
	return os.WriteFile(stampPath, []byte(content), 0o644)
}

// parseStampTimestamp extracts the RFC3339 timestamp from the first line of
// the stamp file. Leading/trailing whitespace is trimmed. A non-parseable
// string yields an error (caller fails-open).
func parseStampTimestamp(raw string) (time.Time, error) {
	line := strings.TrimSpace(strings.SplitN(raw, "\n", 2)[0])
	if line == "" {
		return time.Time{}, fmt.Errorf("empty timestamp")
	}
	// Support both bare-integer unix epochs and RFC3339 stamps so a future
	// stamp writer can migrate formats without breaking the reader.
	if _, err := strconv.ParseInt(line, 10, 64); err == nil {
		// Looks like a unix epoch — but we only write RFC3339 today, so this
		// branch is a forward-compat reader. Parse as seconds.
		if sec, serr := strconv.ParseInt(line, 10, 64); serr == nil {
			return time.Unix(sec, 0).UTC(), nil
		}
	}
	return time.Parse(time.RFC3339, line)
}
