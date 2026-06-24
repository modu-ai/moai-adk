---
paths: "**/internal/spec/**,**/.moai/specs/**"
---

# Lifecycle Sync Gate Protocol — SSOT

> **Single Source of Truth** for the SPEC era classification heuristic table,
> grandfather clause policy, frontmatter `era:` field semantics, and the
> corresponding status-transition ownership cross-reference.
>
> Enforcement: `internal/spec/era.go` `ClassifyEra()` (canonical Go implementation),
> `internal/spec/audit.go` `Audit()` (drift detection engine).
> Cross-referenced by: `.claude/rules/moai/development/spec-frontmatter-schema.md`
> § Optional Fields, `internal/spec/CLAUDE.md`, `internal/spec/audit_test.go`.

---

## Era Classification Heuristic

[ZONE:Evolvable] [HARD] Every SPEC in `.moai/specs/` is classified into exactly
one of five era buckets. The classification determines whether the SPEC is subject
to lifecycle drift detection (V3R6 modern era only) or protected by the grandfather
clause (V2.x / V3R2-R4 / V3R5).

### Era Definitions

| Era | Period | Lifecycle Standard |
|-----|--------|-------------------|
| V2.x | Pre-2026-02 | No `progress.md`; SPEC implementation via direct commit |
| V3R2-R4 | 2026-02 ~ 2026-03 | `progress.md` introduced; no `sync_commit_sha` |
| V3R5 | 2026-03 ~ 2026-04 | Sync section emerges; `sync_commit_sha` not enforced |
| V3R6 | 2026-04 ~ present | 3-phase modern standard (plan / run / sync); `sync_commit_sha` required (the former `mx_commit_sha` / `§E.5 Mx-phase` signals were retired per SPEC-V3R6-LIFECYCLE-REDESIGN-001 — MX Tag is a cross-cutting sync concern, not a separate phase) |
| unclassified | — | Auto-detection ambiguous; no heuristic matched (H-6 fallback) |

### Heuristic Detection Table

The following H-1..H-6 heuristic rules are applied in order. **First match wins**,
except for H-override which takes absolute precedence when the optional frontmatter
`era:` field is present and valid.

| Heuristic ID | Rule | Era Inferred |
|--------------|------|--------------|
| H-override | `spec.md` frontmatter `era:` field non-empty and valid → returned verbatim (no auto-detection) | (explicit override) |
| H-1 | `.moai/specs/SPEC-*/progress.md` absent | V2.x |
| H-2 | `progress.md` present, no `§E.2` / `§E.3` / `§E.4` section markers | V3R2-R4 |
| H-3 | `progress.md` `§E.2` present, `sync_commit_sha` field absent or null | V3R5 |
| H-4 | `progress.md` `§E.2` present + `§E.4` present + `sync_commit_sha` SHA value (the new 3-phase predicate; per SPEC-V3R6-LIFECYCLE-REDESIGN-001 REQ-LR-005/006, a legacy-layout SPEC carrying the retired `§E.5 + mx_commit_sha` also classifies as V3R6 via a migration-window dual predicate) | V3R6 |
| H-5 | H-4 ambiguous; `spec.md` frontmatter `phase:` field references `"v3.0"` or `"v3R6"` OR `created:` date >= 2026-04-01 (tie-breaker) | V3R6 |
| H-6 | No heuristic matched | unclassified |

### Canonical Go Implementation

The classification logic is the authoritative implementation in
`internal/spec/era.go` `ClassifyEra(EraSignals)`. The function returns both the
classified era and a short rationale string (e.g., `"H-4 (§E.2 + §E.4 + sync_commit_sha present)"` for the new 3-phase predicate). The rationale string is included in `EraAutoDetected`
findings emitted by the audit engine.

#### `EraSignals` structure

The caller populates an `EraSignals` struct from disk (via `LoadEraSignalsFromDir`)
before calling `ClassifyEra`. The struct carries:

- `FrontmatterEra` — value of the optional `era:` frontmatter field (H-override)
- `ProgressMDExists` — whether `progress.md` is present in the SPEC directory (H-1)
- `ProgressMDContent` — raw content of `progress.md` (H-2..H-4 section/field detection)
- `FrontmatterPhase` — value of `phase:` frontmatter field (H-5 tie-breaker)
- `FrontmatterCreated` — value of `created:` frontmatter field, ISO-8601 YYYY-MM-DD (H-5 tie-breaker)

#### `modernEraThreshold`

The H-5 date tie-breaker uses the constant `modernEraThreshold = "2026-04-01"`.
SPECs created on or after this date are classified as V3R6 when H-4 is ambiguous.

#### `normalizeEra` — accepted string values

When the frontmatter `era:` field is present, its value is normalized by
`normalizeEra()` to a canonical `Era` constant. The following aliases are accepted
(case-sensitive match, leading/trailing space stripped):

| Accepted value | Canonical Era |
|----------------|---------------|
| `V2.x`, `v2.x`, `V2`, `v2` | `EraV2x` |
| `V3R2-R4`, `v3r2-r4`, `V3R2`, `V3R3`, `V3R4` | `EraV3R2R4` |
| `V3R5`, `v3r5` | `EraV3R5` |
| `V3R6`, `v3r6` | `EraV3R6` |
| `unclassified` | `EraUnclassified` |

Invalid values (unrecognized strings) fall through to auto-detection rather than
returning an error — this prevents a typo in the frontmatter from silently
suppressing drift detection on a modern-era SPEC.

---

## Grandfather Clause Policy

[ZONE:Evolvable] [HARD] SPECs classified as **V2.x**, **V3R2-R4**, or **V3R5** are
**grandfather-clause-protected**. This means:

1. `moai spec audit` classifies them as `era_final: true`.
2. **No drift findings** are emitted for grandfathered SPECs, regardless of their
   cross-tab pattern (missing sections, absent commit SHAs, etc.).
3. When `--include-grandfathered` is passed, grandfathered SPECs appear in the
   JSON output but with `severity: "INFO"` only — they are never promoted to
   `MUST-FIX`.

Only **V3R6 SPECs** (the modern era) are subject to lifecycle invariants and
drift detection. The `Era.IsModern()` method returns `true` exclusively for
`EraV3R6`.

### Rationale

Retroactive normalization of historical SPECs is operationally infeasible and
provides no production value. The grandfather clause acknowledges that era-specific
lifecycle standards were legitimate at their time of authorship. Applying V3R6
invariants retroactively would produce false findings on SPECs that were correctly
managed under their respective era's conventions.

### `modernEraThreshold` as boundary

The threshold `2026-04-01` acts as both the H-5 tie-breaker date and the conceptual
boundary between V3R5 (grandfathered) and V3R6 (subject to modern-era drift
detection). SPECs created before this date and lacking H-4 signals are assigned to
V3R5 or earlier eras and therefore receive grandfather protection.

### Audit output fields

When the audit engine processes a SPEC, it sets:

- `era_final: true` — for V2.x / V3R2-R4 / V3R5 (grandfather-protected)
- `IsModern() == true` — for V3R6 (subject to drift detection)
- `era: "unclassified"` + `severity: "INFO"` — when H-6 fallback fires

---

## Frontmatter Era Field Semantics

[ZONE:Evolvable] The optional `era:` field in `spec.md` YAML frontmatter provides
an **explicit override** of auto-detection. When present and valid, it supersedes
the H-1..H-5 heuristic chain entirely (H-override rule).

### Field specification

```yaml
---
# ...required 12 fields...
era: V3R6   # optional; explicit override of ClassifyEra auto-detection
---
```

- **Type**: enum string
- **Accepted values**: `V2.x` / `V3R2-R4` / `V3R5` / `V3R6` / `unclassified`
  (plus case-insensitive aliases; see `normalizeEra` table above)
- **Default when absent**: auto-detection via H-1..H-6 heuristic chain
- **Effect when present and valid**: `ClassifyEra` returns the specified era
  verbatim without inspecting `progress.md` or other signals
- **Effect when present but invalid**: falls through to auto-detection
  (invalid value is not accepted; the `normalizeEra` function returns false)

### When to set `era:` explicitly

The `era:` field is useful in the following scenarios:

1. **Ambiguous SPECs near era boundaries** — a SPEC created on exactly
   `2026-04-01` with incomplete `progress.md` sections might otherwise classify
   as `unclassified` (H-6). Setting `era: V3R5` or `era: V3R6` resolves the
   ambiguity explicitly.
2. **Historical SPECs with non-standard progress.md structure** — a SPEC that
   does not follow the standard `§E.2`/`§E.4` section naming but is known to
   belong to a specific era can be pinned via the field.
3. **Newly created SPECs before `progress.md` is populated** — at plan-phase
   creation, `progress.md` may be empty or minimal; setting `era: V3R6` avoids
   transient misclassification as V3R2-R4 (H-2 fires when no `§E.*` markers
   are present).

### When NOT to set `era:` explicitly

Do not set `era:` in the following cases:

- **Standard V3R6 SPECs with complete `progress.md`** — H-4 auto-detection is
  reliable and the explicit field is redundant.
- **Grandfathered SPECs that already have correct auto-classification** — adding
  `era: V2.x` to a SPEC that H-1 already classifies as V2.x adds noise without
  benefit.

### Relationship to `lint.skip`

The `era:` field is NOT a lint-skip mechanism. To suppress a specific lint
finding (e.g., `OwnershipTransitionInvalid`), use the separate `lint.skip`
optional frontmatter field:

```yaml
lint.skip: [OwnershipTransitionInvalid]
```

See `.claude/rules/moai/development/spec-frontmatter-schema.md` § Optional Fields
for the `lint.skip` definition.

---

## Status Transition Ownership Matrix Cross-Reference

[ZONE:Evolvable] The era classification interacts with the Status Transition
Ownership Matrix defined in `.claude/rules/moai/development/spec-frontmatter-schema.md`
§ Status Transition Ownership Matrix.

### How era affects ownership enforcement

The `OwnershipTransitionRule` (implemented in `internal/spec/lint_ownership.go`)
enforces that each canonical status transition is performed by the expected agent.
Era classification modulates this enforcement as follows:

| Era | Ownership enforcement | Notes |
|-----|-----------------------|-------|
| V2.x | **Not enforced** (grandfather-protected) | `era_final: true`; OwnershipTransitionRule emits no findings for these SPECs |
| V3R2-R4 | **Not enforced** (grandfather-protected) | Same as V2.x |
| V3R5 | **Not enforced** (grandfather-protected) | Same as V2.x |
| V3R6 | **Enforced** — transitions must match canonical owner | `IsModern() == true`; rule emits `OwnershipTransitionInvalid` on mismatch |
| unclassified | **Not enforced** (ambiguous era; conservative default) | INFO-only treatment; no MUST-FIX findings |

### Canonical ownership matrix (summary)

> Canonical: the full 7-row Status Transition Ownership Matrix (transition → owning agent → canonical commit subject pattern, including the `(none) → draft` / `draft → in-progress` / `in-progress → implemented → completed` / `* → superseded|archived|rejected` rows) lives in `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix. This file owns only the era-modulation table above and the close-subject-full-ID one-liner below — both are lifecycle-gate-local deltas not present in the schema SSOT.

### Close-subject full-ID mandate

Per SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001, every close commit (the sync commit carrying the `implemented → completed` transition above) MUST name exactly one individual full SPEC-ID in its subject scope — e.g. `chore(SPEC-CCSYNC-CLAUDEMD-001): … 3-phase close`. A **combined/abbreviated scope** that names only a shared prefix (e.g. `chore(SPEC-CCSYNC): … 3-phase close (CLAUDEMD + TOOLCAT)`) is **prohibited**: the drift detector's exact-token SPEC-ID extraction cannot map an abbreviated prefix to its sibling SPECs, so combined-scope close subjects regenerate lifecycle drift false-positives. When closing N sibling SPECs together, emit N separate close commits, one per full SPEC-ID — combined/abbreviated scope is disallowed in close subjects. The drift detector accommodates historical combined-scope closes via a secondary scope-prefix grep fallback (see `internal/spec/drift.go` `resolveCombinedScopeClose`), but this doctrine prevents recurrence in new closes.

> **D4 reconciliation note (SPEC-V3R6-LIFECYCLE-REDESIGN-001 REQ-LR-020/021)**: The close-subject convention (including the close infix literal) is owned by **SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001**. SPEC-V3R6-LIFECYCLE-REDESIGN-001 amends the close infix from the legacy `"4-phase close"` to the canonical `"3-phase close"` in this prose. The drift detector's close-infix matcher (`internal/spec/transitions.go` `closeInfixMatch`) has been extended (M2) to accept BOTH infixes — the legacy `"4-phase close"` is RETAINED in the matcher because historical close commits in git history carry it. A doc-only rename without the dual-infix matcher update was forbidden (it would silently break drift close-recognition). This note credits DRIFT-LEGACY-CONVENTION-001 as the convention owner; it does NOT silently override it.

### `Authored-By-Agent` trailer as the gating signal

The `OwnershipTransitionRule` uses the `Authored-By-Agent:` commit body trailer
as the mechanical WHO signal. Accepted trailer values:

- `manager-spec`
- `manager-develop`
- `manager-docs`
- `manager-git`
- `orchestrator-direct`

Commits **without** the `Authored-By-Agent:` trailer are treated as legacy /
non-MoAI commits and are NOT subject to ownership validation (silent SKIP).
This prevents false positives on historical commits written before this
convention was established.

### Lint finding codes

- **`OwnershipTransitionInvalid`** (Warning severity) — emitted when a V3R6
  SPEC commit has the `Authored-By-Agent:` trailer AND the agent performing
  the transition does NOT match the canonical owner for that transition.
- **`OwnershipTransitionSkipped`** (Info severity) — emitted when
  `lint.skip: [OwnershipTransitionInvalid]` is present in frontmatter.
- **`OwnershipTransitionUnreachable`** (Info severity) — emitted when
  `git log` is unavailable (non-git environment, CI sandbox without history).

---

## Worked Example: Era Auto-Detection

This section demonstrates the end-to-end flow for a V3R6 SPEC that does NOT have
an explicit `era:` frontmatter field, exercising the auto-detection path that
produces an `EraAutoDetected` INFO finding.

### Scenario

A SPEC at `.moai/specs/SPEC-EXAMPLE-001/` has the following files:

**`spec.md` frontmatter** (excerpt):
```yaml
---
id: SPEC-EXAMPLE-001
title: "Example Feature"
version: "0.1.2"
status: implemented
created: 2026-05-01
updated: 2026-05-20
author: Author Name
priority: P1
phase: "v3.0.0"
module: "internal/example"
lifecycle: spec-anchored
tags: "example, demo"
# Note: no 'era:' field present
---
```

**`progress.md`** (excerpt — 4-section layout per SPEC-V3R6-LIFECYCLE-REDESIGN-001; NO `§E.5 Mx-phase` section):
```yaml
# ...
## §E.1 Plan-phase Audit-Ready Signal
# ...
## §E.2 Run-phase Evidence
# (run-evidence start marker — the literal §E.2 heading is what hasAnyProgressMarker detects)
# ...
## §E.3 Run-phase Audit-Ready Signal
# ...
## §E.4 Sync-phase Audit-Ready Signal
sync_commit_sha: "a1b2c3d4e5f6"
# (§E.5 Mx-phase section is RETIRED — folded into §E.4 per SPEC-V3R6-LIFECYCLE-REDESIGN-001;
#  MX Tag validation is a cross-cutting sync concern, not a separate phase.)
```

### Auto-detection trace

1. `LoadEraSignalsFromDir("SPEC-EXAMPLE-001/")` is called.
2. `FrontmatterEra` is empty (no `era:` field in frontmatter).
3. H-override: skipped (field absent).
4. `ProgressMDExists = true` (H-1 does not fire).
5. `hasAnyProgressMarker()` returns `true` (`§E.2` present) (H-2 does not fire).
6. `hasSyncSection = true` (`§E.2` marker present — note: `hasSyncSection` is a misnomer; it tests the literal `§E.2` run-evidence start marker, not the sync phase, which lives at `§E.4`).
7. `syncSHA = "a1b2c3d4e5f6"` (non-empty — field extracted from the `§E.4 Sync-phase Audit-Ready Signal` section).
8. H-3 check: `hasSyncSection && syncSHA == ""` → **false** (H-3 does not fire).
9. H-4 check (new 3-phase predicate per REQ-LR-005): `hasSyncSection && hasMxSection4Phase && syncSHA != ""` where `hasMxSection4Phase` detects the `§E.4` sync marker → **true**. (Legacy-layout SPECs carrying the retired `§E.5 + mx_commit_sha` also match via the dual-predicate migration window REQ-LR-006.)
10. `ClassifyEra()` returns `(EraV3R6, "H-4 (§E.2 + §E.4 + sync_commit_sha present)")`.

### Audit output (JSON excerpt)

When `moai spec audit --json` runs against the project, the output includes:

```json
{
  "audited_at": "2026-05-30T10:00:00Z",
  "total_specs": 1,
  "grandfathered": 0,
  "modern_era_clean": 0,
  "drift_findings": [
    {
      "spec_id": "SPEC-EXAMPLE-001",
      "era": "V3R6",
      "finding_type": "EraAutoDetected",
      "severity": "INFO",
      "remediation": "Add 'era: V3R6' to spec.md frontmatter to suppress this finding",
      "details": {
        "heuristic_matched": "H-4 (§E.2 + §E.4 + sync_commit_sha present)"
      }
    }
  ]
}
```

The `EraAutoDetected` INFO finding notifies operators that the era was derived
from heuristics rather than an explicit frontmatter override. It carries the
`details.heuristic_matched` field — the human-readable rationale string from
`ClassifyEra()` — so operators can verify the inference is correct.

### Go test binding

This flow is tested in `internal/spec/audit_test.go::TestEraAutoDetection`:

```go
// TestEraAutoDetection: fixture without 'era:' field matching V3R6 heuristic,
// assert era=V3R6 + INFO finding with details.heuristic_matched set.
func TestEraAutoDetection(t *testing.T) { ... }
```

The test fixture uses a SPEC directory with a `progress.md` containing `§E.2` +
`§E.4` markers and a non-empty `sync_commit_sha` value (the 4-section layout), and
asserts that the audit result carries `era: "V3R6"` AND an `EraAutoDetected`
finding with a non-empty `details.heuristic_matched` field.

### How to suppress the finding

If the auto-detection result is correct and the operator wants to suppress the
INFO finding, add the `era:` field to `spec.md` frontmatter:

```yaml
era: V3R6
```

With the explicit field present, H-override fires, `ClassifyEra` skips
auto-detection entirely, and no `EraAutoDetected` finding is emitted.

---

## Cross-references

- `internal/spec/era.go` — canonical Go implementation of H-1..H-6 + H-override
- `internal/spec/audit.go` — drift detection engine consuming `ClassifyEra()`
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Optional Fields — `era:` field schema entry; § Status Transition Ownership Matrix — canonical ownership matrix
- `internal/spec/lint_ownership.go` — `OwnershipTransitionRule` enforcement (V3R6 only)
- `internal/spec/era_test.go` — unit tests for H-1..H-6 classification
- `internal/spec/audit_test.go` — `TestEraAutoDetection` + `TestEraClassification5Buckets`

---

Version: 1.0.0
Status: Active — canonical SSOT for era classification + grandfather clause + frontmatter era field semantics
Origin: SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 M5
