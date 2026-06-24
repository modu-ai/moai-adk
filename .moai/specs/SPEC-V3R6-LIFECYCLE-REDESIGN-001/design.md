---
id: SPEC-V3R6-LIFECYCLE-REDESIGN-001
design_version: "0.2.0"
spec_version: "0.2.0"
status: draft
created: 2026-06-18
updated: 2026-06-19
author: manager-spec
tier: L
---

# Design — SPEC-V3R6-LIFECYCLE-REDESIGN-001

## §A. Scope

This design document covers two architectural decisions that require explicit design rationale:
1. **H-4 reclassification strategy** — how `internal/spec/era.go` `ClassifyEra` is rewritten without losing V3R6 classification for the existing V3R6 SPECs (count re-measured at run-phase per D3).
2. **Epic taxonomy** — the final term set and the migration mapping from the legacy Sprint/cohort/Round/Wave vocabulary.

Both decisions are plan-phase artifacts; implementation is deferred to run-phase milestones (M1-M9 per plan.md).

### §A.1 Enumerated edit-scope (D5 — literal-instruction-following completeness)

Beyond the `ClassifyEra` function body and the drift branches, the following surfaces **also hardcode** the `§E.5` / `mx_commit_sha` / "4-phase" terminology and MUST be in the rewrite scope so a literal-instruction-following implementer does not miss them:

- **(a) `internal/spec/era.go` doc-comment block, lines ~86-101** — the `ClassifyEra` godoc heuristic table hardcodes `H-4: §E.2 + §E.5 present AND sync_commit_sha + mx_commit_sha non-empty → V3R6`. The doc-comment MUST be rewritten alongside the function body (M1) to describe the new `§E.2 + §E.4 + sync_commit_sha` predicate + the legacy fallback. (era.go scope: doc-comment lines ~86-101 + function body lines ~117-146 + the `hasMxSection`/`mxSHA` parse lines ~118/120.)
- **(b) `.claude/rules/moai/workflow/lifecycle-sync-gate.md` `## §E.5 Mx-phase Audit-Ready Signal` worked example, line ~303** (with `mx_commit_sha: "f6e5d4c3b2a1"` at ~304) — the Worked Example fixture demonstrates the 5-section layout. M4 MUST update this worked example (and the H-4 heuristic-table row at line ~43, and the era-definition row at line ~28) to reflect the 4-section layout + the new H-4 predicate.

These two items are added to the M1 (era.go) and M4 (lifecycle-sync-gate.md) milestone scopes in plan.md.

## §B. Design Decision 1 — H-4 Reclassification Strategy

### §B.1 Current State (the drift)

`internal/spec/era.go` `ClassifyEra` (lines 117-146):

```go
hasSyncSection := hasProgressMarker(signals.ProgressMDContent, "§E.2")
hasMxSection := hasProgressMarker(signals.ProgressMDContent, "§E.5")
syncSHA := extractProgressField(signals.ProgressMDContent, "sync_commit_sha")
mxSHA := extractProgressField(signals.ProgressMDContent, "mx_commit_sha")

// H-2: progress.md present but no §E.* markers → V3R2-R4
if signals.ProgressMDExists && !hasAnyProgressMarker(signals.ProgressMDContent) {
    return EraV3R2R4, "H-2 (progress.md without §E.* markers)"
}

// H-3: §E.2 present but sync_commit_sha empty → V3R5
if hasSyncSection && syncSHA == "" {
    return EraV3R5, "H-3 (§E.2 present, sync_commit_sha missing)"
}

// H-4: §E.2 + §E.5 + both *_commit_sha non-empty → V3R6
if hasSyncSection && hasMxSection && syncSHA != "" && mxSHA != "" {
    return EraV3R6, "H-4 (§E.2 + §E.5 + both commit_sha present)"
}
```

Key observations:
- `hasSyncSection` is a **misnomer** — it tests `§E.2` (the run-evidence START marker), not the sync phase (which lives at `§E.4`). This is documented in era.go lines 33-36.
- `hasMxSection` tests `§E.5` (the Mx-completion marker) — the drift to be removed.
- The H-4 predicate requires BOTH `§E.5` AND `mx_commit_sha` — these are the V3R6-era accretions that this redesign removes.

### §B.2 Target State (3-phase restoration)

The new H-4 predicate drops the `§E.5 + mx_commit_sha` requirement:

```go
hasRunEvidence := hasProgressMarker(signals.ProgressMDContent, "§E.2")
hasSyncMarker := hasProgressMarker(signals.ProgressMDContent, "§E.4")  // NEW: the actual sync section
syncSHA := extractProgressField(signals.ProgressMDContent, "sync_commit_sha")

// H-4 (NEW): §E.2 run-evidence + §E.4 sync marker + sync_commit_sha → V3R6
if hasRunEvidence && hasSyncMarker && syncSHA != "" {
    return EraV3R6, "H-4 (§E.2 + §E.4 + sync_commit_sha)"
}

// H-4 (LEGACY FALLBACK — migration window, REQ-LR-006):
// SPECs authored before the redesign still carry §E.5 + mx_commit_sha.
// Treat them as V3R6 during the migration window.
hasLegacyMx := hasProgressMarker(signals.ProgressMDContent, "§E.5")
mxSHA := extractProgressField(signals.ProgressMDContent, "mx_commit_sha")
if hasRunEvidence && hasLegacyMx && syncSHA != "" && mxSHA != "" {
    return EraV3R6, "H-4-legacy (§E.2 + §E.5 + both commit_sha — migration window)"
}
```

Variable renaming:
- `hasSyncSection` → `hasRunEvidence` (resolves the misnomer; §E.2 is run-evidence start, not sync).
- New `hasSyncMarker` tests `§E.4` (the actual sync section).

#### §B.2a Corrected fall-through mechanism (D1 — verified against `internal/spec/era.go` lines 102-146)

The original draft of this design (and research.md §D.3) claimed that a SPEC with `§E.2 present, §E.4 absent, sync_commit_sha="abc123"` would, after the H-4 rewrite, fall to **H-3 → V3R5 (regression)**. **This was FALSE.** The actual `ClassifyEra` control flow is:

1. **H-3** (`era.go:130`): `if hasSyncSection && syncSHA == ""` — fires ONLY when `sync_commit_sha` is **EMPTY**. A SPEC carrying a non-empty `sync_commit_sha="abc123"` does **NOT** match H-3. (`hasSyncSection` here tests the `§E.2` run-evidence marker, per the misnomer note above — it has nothing to do with `§E.4`.)
2. After the rewritten H-4 misses (no `§E.4`) and H-3 misses (sync_sha non-empty), control reaches **H-5** (`era.go:140`): `if matchesModernPhase(phase) || isAfterModernThreshold(created)` → **V3R6**.
3. Only if H-5 ALSO misses (neither a modern `phase:` matching `v3r6`/`v3.0*` NOR `created >= 2026-04-01` per `modernEraThreshold`) does control fall to **H-6 → unclassified**.

Therefore the genuine at-risk population after the H-4 rewrite is **NOT** "SPECs that regress to V3R5"; it is "SPECs that would fall to **H-6 unclassified**" — i.e., a V3R6 SPEC that (a) lacks `§E.4`, (b) lacks the legacy `§E.5`+`mx_sha` predicate, AND (c) has neither a modern `phase:` nor `created >= 2026-04-01`. The corrected worked-example trace lives in research.md §D.3.

#### §B.2b Re-derived at-risk set (D1 — empirical, plan-phase snapshot 2026-06-18)

A concrete measurement against the live catalog (`moai spec audit --json` + per-SPEC frontmatter inspection of `phase:`/`created:`) of the current V3R6 population produced:

| Disposition after H-4 rewrite | Count (illustrative, e.g. N≈53 as of plan-phase) |
|-------------------------------|------------------|
| Caught directly by NEW H-4 (`§E.2 + §E.4 + sync_sha`) | 29 |
| NOT caught by new H-4; rely on legacy-fallback OR H-5 (lack `§E.4`, carry legacy `§E.5`+`mx_sha`) | 11 |
| Caught by H-5 (modern `phase:` AND/OR `created >= 2026-04-01`) as a safety net | all 53 (45 phase+date, 8 date-only) |
| **Fall to H-6 unclassified (genuine at-risk)** | **0** |

The genuine H-6 at-risk set is **empty** at plan-phase: every current V3R6 SPEC is caught by H-5 even with NO migration window, because all 53 carry `created >= 2026-04-01` (the `modernEraThreshold`) and/or a modern `phase:`. The catalog is moving (a parallel session is authoring more V3R6 SPECs), so this count MUST be re-measured at run-phase M1 start (see plan.md M1 step "re-derive H-6 at-risk set") rather than frozen here.

### §B.3 Migration Strategy — S1 + Temporary S3 Window

**Selected strategy**: S1 (auto-migrate / backfill) + a temporary S3 (dual-predicate) window.

Rationale (from research.md §D.1):
- S1 produces the clean end state (4-section layout across all V3R6 SPECs).
- The temporary S3 window prevents misclassification between M1 (code rewrite) and M3 (backfill).
- S2 (grandfather the 48) is rejected — permanent bifurcation.
- Pure S3 (permanent dual-predicate) is rejected — defeats the redesign.

**Backfill script (M3)**: `internal/spec/migrate_3phase.go` (new file) iterates the 48 affected SPECs:
1. Reads `progress.md`.
2. Extracts `§E.5 Mx-phase Audit-Ready Signal` section content.
3. Appends it to `§E.4 Sync-phase Audit-Ready Signal` under a `### (Migrated from §E.5)` sub-heading.
4. Removes the `§E.5` section.
5. Records the migration in `.moai/state/lifecycle-redesign-migration.json`.

**Migration log schema**:
```json
{
  "migration_date": "2026-06-XX",
  "spec_id": "SPEC-EXAMPLE-001",
  "folded_from": "§E.5",
  "folded_to": "§E.4",
  "mx_commit_sha_preserved": true,
  "backfill_commit": "<sha>"
}
```

**Window retirement**: After M3 + 1 release cycle, the legacy fallback branch can be removed from `ClassifyEra`. Open question OQ-1 defers this decision.

**Migration-window scope correction (D1)**: Because the re-derived genuine H-6 at-risk set is **empty** (§B.2b — every current V3R6 SPEC is caught by H-5's `created >= 2026-04-01` / modern-`phase:` test even with NO migration window), the dual-predicate window (REQ-LR-006) is **defense-in-depth + classification-rationale precision**, NOT a misclassification-prevention necessity. Its surviving value: the 11 legacy-layout SPECs (lacking `§E.4`, carrying `§E.5`+`mx_sha`) classify as "V3R6 via legacy predicate" — an explicit, precise rationale — rather than falling through to the weaker "V3R6 via H-5 date heuristic". REQ-LR-006 is therefore **NARROWED** (kept, but re-scoped from "prevent the 48/53 from regressing to V3R5" to "preserve precise V3R6 rationale for legacy-layout SPECs as a defensive belt while the catalog is migrated"). It is NOT removed, because (a) the catalog is moving and a future V3R6 SPEC authored with a pre-`modernEraThreshold` `created:` and no modern `phase:` COULD reach H-6, and (b) an explicit predicate is a stronger classification signal than the H-5 date heuristic.

### §B.4 Drift Detection Update (audit.go) — D2 corrected: all THREE §E.5-keyed findings

`internal/spec/audit.go` `checkV3R6Drift` (lines 224-300, verified) emits **THREE** §E.5-keyed MUST-FIX findings, NOT two. The original draft of this design addressed only the first two; the third (`Y_N_N_Y`) is the one the redesign's 4-section end-state actively triggers, and it MUST be in scope:

| Finding constant | audit.go line | Predicate | Post-redesign disposition |
|------------------|---------------|-----------|---------------------------|
| `Y_Y_Y_Y_StatusDrift` | 251-266 | `hasSyncSection(§E.2) && hasMxSection(§E.5) && syncSHA != "" && mxSHA != ""` and status != completed | **Re-anchored** to the new 3-marker predicate (`§E.2 + §E.4 + sync_commit_sha` present but status != completed) — `§E.5`/`mx_sha` dropped from the predicate. |
| `Y_Y_N_Y` | 268-282 | `hasSyncSection(§E.2) && hasMxSection(§E.5) && mxSHA == ""` and status != completed | **Retired** (REQ-LR-019) — `mx_commit_sha` is no longer a drift dimension. |
| `Y_N_N_Y` | 284-297 | `hasSyncSection(§E.2) && !hasMxSection(§E.5)` and status != completed | **Retired / re-anchored (D2 — the critical one)** — see below. |

**Why `Y_N_N_Y` is the blocking defect.** Its current message is "§E.2 sync section present but §E.5 Mx section absent" and it fires for any non-`completed`, non-terminal V3R6 SPEC where `§E.2` is present and `§E.5` is **absent**. After M3 folds away `§E.5` from every V3R6 SPEC, the mandated 4-section end-state means EVERY in-progress / implemented V3R6 SPEC has `§E.2` present and `§E.5` absent — so `Y_N_N_Y` would fire MUST-FIX on **every non-completed V3R6 SPEC in the catalog**. Leaving it untouched converts the redesign's intended clean end-state into a catalog-wide drift storm. M2 MUST retire or re-anchor `Y_N_N_Y` when `§E.5` is removed as a drift dimension.

Post-redesign target for `checkV3R6Drift`:
- The single surviving drift dimension is **status drift on the 3-marker predicate**: `§E.2 + §E.4 + sync_commit_sha` present (sync complete) but `spec.md status != completed`.
- The two `§E.5`/`mx_sha`-keyed branches (`Y_Y_N_Y`, the §E.5-absence branch `Y_N_N_Y`) are removed; the `Y_Y_Y_Y_StatusDrift` branch is re-anchored to the 3-marker predicate.

The `FindingType` constants at audit.go lines 54-65 are updated:
```go
// FindingSyncStatusDrift (renamed from Y_Y_Y_Y_StatusDrift):
// §E.2 + §E.4 + sync_commit_sha present but status != completed.
FindingSyncStatusDrift = "SyncStatusDrift"
// FindingY_Y_N_Y and FindingY_N_N_Y are REMOVED — §E.5/mx_commit_sha is no
// longer a drift dimension under the 3-phase lifecycle.
```

M2 scope explicitly includes `internal/spec/audit.go` lines **251-297** (all three finding branches) + lines 54-65 (the `FindingType` constants) + the corresponding `internal/spec/audit_test.go` fixtures for all three findings.

### §B.5 Test Strategy (M2)

`internal/spec/era_test.go` `TestClassifyEra_HeuristicTable` gains new fixtures:
- `H-4 new predicate`: §E.2 + §E.4 + sync_sha → V3R6.
- `H-4 legacy fallback`: §E.2 + §E.5 + sync_sha + mx_sha → V3R6 (migration window).
- `H-3 unchanged`: §E.2 + sync_sha missing → V3R5.
- `H-3 new edge case`: §E.2 + §E.4 but sync_sha missing → V3R5 (sync not yet complete).

`internal/spec/audit_test.go` gains (covering ALL three §E.5-keyed findings per D2):
- `SyncStatusDrift retired Y_Y_N_Y`: assert no `Y_Y_N_Y` finding for a SPEC with `§E.2`+`§E.5` but missing `mx_commit_sha`.
- `SyncStatusDrift retired Y_N_N_Y (D2)`: assert no `Y_N_N_Y` finding for a 4-section SPEC (`§E.2` present, `§E.5` ABSENT, status=in-progress) — this is the catalog-wide false-positive the redesign must NOT regenerate.
- `SyncStatusDrift new predicate`: assert `SyncStatusDrift` finding for `§E.2` + `§E.4` + sync_sha + status=in-progress (sync complete but status not advanced).
- `SyncStatusDrift completed clean`: assert NO finding for a 4-section SPEC with status=completed (the clean end state).

### §B.6 Close-Subject Convention Reconciliation (D4 — cross-SPEC ownership: DRIFT-LEGACY-CONVENTION-001)

The literal string **"4-phase close"** is NOT merely documentation prose — it is a load-bearing token in the drift detector's close-recognition logic, and it is OWNED by a different SPEC. The redesign's "4-phase close → 3-phase close" rename therefore requires explicit cross-SPEC reconciliation; a naive doc-only sweep would silently break the drift detector.

**Ownership (verified)**: The close-subject full-ID mandate that carries "4-phase close" lives in:
- `.claude/rules/moai/development/spec-frontmatter-schema.md` line ~71 (Status Transition Ownership Matrix `implemented → completed` row) + line ~78 (close-subject full-ID mandate prose).
- `.claude/rules/moai/workflow/lifecycle-sync-gate.md` line ~228 (matrix row) + line ~235 (close-subject mandate prose, attributed verbatim to **SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001**).

Both close-subject mandates are attributed to **SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001** as the canonical cross-SPEC owner of the close-subject convention. This redesign MUST NOT silently override that SPEC's convention; it must reconcile with it.

**The mechanical dependency (the load-bearing reason)**: `internal/spec/transitions.go` lines 73-83 define the close-recognition constants and matcher:
```go
const (
    closeInfix4Phase = "4-phase close"        // line 74
    closeInfixMx     = "mx-phase audit-ready" // line 75
)
func closeInfixMatch(lowerTitle string) bool { // line 80
    return strings.Contains(lowerTitle, closeInfix4Phase) ||
        strings.Contains(lowerTitle, closeInfixMx)
}
```
`closeInfixMatch` is, per its own `@MX:NOTE` (transitions.go:69), **"the walker's only positive `completed` signal"**. It is consumed by `internal/spec/drift.go` (`shouldSkipCommitTitle` line ~298, `resolveCombinedScopeClose` line ~457) and `ClassifyPRTitle` (transitions.go:110). If the redesign renames "4-phase close" → "3-phase close" in commit subjects WITHOUT updating `closeInfix4Phase`, the drift walker will stop recognizing future close commits → every newly-closed V3R6 SPEC would be mis-detected as an incomplete close (a drift false-positive). The two literals must move together.

**Reconciliation decision**: The "4-phase close" → "3-phase close" rename is split into a doc-axis and a code-axis that MUST ship together in the same milestone:
1. **Doc-axis (M4)**: the rename in the close-subject mandate prose of spec-frontmatter-schema.md + lifecycle-sync-gate.md. Because these two files are owned by SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001's convention, M4 MUST add a one-line reconciliation note in each crediting DRIFT-LEGACY-CONVENTION-001 as the close-subject convention owner and recording that LIFECYCLE-REDESIGN-001 amends the infix from "4-phase close" to "3-phase close".
2. **Code-axis (added to M2 scope)**: update `closeInfix4Phase` in `internal/spec/transitions.go` from `"4-phase close"` to accept the new `"3-phase close"` infix. **Backward-compatibility constraint**: historical close commits already in git history carry "4-phase close"; the matcher MUST accept BOTH `"3-phase close"` (new) AND `"4-phase close"` (legacy) so that `drift.go` does not regress on already-closed SPECs. Concretely, add a third constant `closeInfix3Phase = "3-phase close"` and OR it into `closeInfixMatch`; do NOT delete `closeInfix4Phase`. Update `internal/spec/drift_combined_scope_test.go` + `internal/spec/transitions_test.go` to assert both infixes are recognized.

**Verification the drift detector is unaffected**: AC-LR-012 (new) asserts `closeInfixMatch("...3-phase close...") == true` AND `closeInfixMatch("...4-phase close...") == true` (legacy still recognized), and that `go test ./internal/spec/...` (including `drift_combined_scope_test.go`) passes after the change.

## §C. Design Decision 2 — Epic Taxonomy

### §C.1 Final Term Set (4 terms)

| Term | Definition | Spec Kit alignment |
|------|------------|-------------------|
| **Epic** | A time-unit or thematic container for one or more SPECs, grouped by schedule, release, or theme. Replaces `Sprint`. | Not in Spec Kit; industry-standard Agile/SAFe term. |
| **SPEC** | A single work unit (feature, refactor, bugfix) with a unique `SPEC-{DOMAIN}-{NUM}` identifier. | Direct alignment with Spec Kit `spec`. |
| **Milestone** | An ordered within-SPEC work step (M1, M2, ... M6). The standard manager-develop delegation unit. | Aligns with Spec Kit `tasks` granularity. |
| **Constitution** | Project-level governing principles and development guidelines. Referenced for SDD vocabulary alignment. | Direct alignment with Spec Kit `constitution`. |

### §C.2 Migration Mapping Table

| Legacy term | New term | Semantic mapping | Disposition |
|-------------|----------|------------------|-------------|
| `Sprint` | `Epic` | Multi-SPEC grouping (time-unit or thematic) | **Renamed** |
| `Sprint N Lane A/B` | `Epic N Lane A/B` | Epic-internal parallel track sub-grouping | **Renamed** (Lane retained) |
| `cohort` | (folded into Epic) | Epic-internal grouping; use "Epic N cohort" or "Epic N Lane X" | **Removed** as standalone term |
| `Round` | (folded into Milestone) | Within-SPEC SSE-stall sub-division; use Milestone | **Removed** |
| `Wave` | (retired) | Legacy pre-Round term (per AP-SRN-004, already retired) | **Confirmed retired** |
| `Milestone` | `Milestone` (unchanged) | Within-SPEC ordered step | **Retained** |
| `SPEC` | `SPEC` (unchanged) | Single work unit | **Retained** (SDD-aligned) |
| (new) | `Constitution` | Project governance layer | **Added** |

### §C.3 Anti-Pattern Re-anchoring

The 4 anti-patterns in `sprint-round-naming.md` (AP-SRN-001..004) are re-anchored:

| AP ID | Old anchor | New anchor |
|-------|-----------|------------|
| AP-SRN-001 | "Calling a multi-SPEC group 'Round'" | "Calling a multi-SPEC group 'Round', 'Wave', or 'cohort' instead of 'Epic'" |
| AP-SRN-002 | "Calling a single-SPEC Milestone 'Round'" | "Calling a within-SPEC Milestone 'Round' instead of 'Milestone'" |
| AP-SRN-003 | "Mixing Sprint and Round in the same context" | "Mixing Epic and Round/Wave in the same context" |
| AP-SRN-004 | "Legacy 'Wave' terminology" | **Retired** (Wave is confirmed removed; no longer a live anti-pattern) |

A new anti-pattern is added:
- **AP-SRN-005 (new)**: "Using `cohort` as a standalone multi-SPEC grouping term instead of folding it into Epic-internal Lane/Cohort sub-naming."

### §C.4 Localization

Per the existing localization convention (Korean user-facing, English technical identifiers):

| Context | English form | Korean form |
|---------|-------------|-------------|
| SPEC ID, file name, rule cross-reference | Epic 1, Milestone M2 | (English required) |
| Korean documentation, paste-ready resume | Epic 1 / 에픽 1 | 에픽 1, 마일스톤 M2 |
| Commit message body (`git_commit_messages: ko`) | mixed allowed | "에픽 1 4/4 SPEC 완료" |
| Code comments (`code_comments: ko`) | English when `en` | "// 에픽 1 단계" when `ko` |

## §D. Alternatives Considered

### §D.1 Axis A — Alternative 1: Keep 4-phase, document Mx as canonical

**Rejected**: The user's context explicitly identifies the 4-phase model as drift from the original 3-phase intent. The redesign restores the original; documenting the drift as canonical would codify the error.

### §D.2 Axis A — Alternative 2: Remove Mx-phase but keep §E.5 as a sync sub-section

**Rejected**: §E.5 is structurally a top-level section (parallel to §E.1..§E.4). Keeping it as a sync sub-section would require renaming it `§E.4.1` or similar, adding complexity without value. The clean fold into §E.4 (S1 strategy) is simpler.

### §D.3 Axis B — Alternative 1: Adopt Spec Kit vocabulary wholesale (drop Epic)

**Rejected**: Spec Kit has no multi-feature grouping term beyond "project". MoAI genuinely needs a mid-level grouping (the former Sprint) for release planning and parallel-track coordination. Epic is the industry-standard term that fills this gap without conflicting with SDD vocabulary.

### §D.4 Axis B — Alternative 2: Keep Sprint but document it as MoAI-specific

**Rejected**: The user's context explicitly flags Sprint as legacy Agile terminology unaligned with SDD. Renaming to Epic aligns MoAI with the broader agentic-coding ecosystem.

### §D.5 Axis B — Alternative 3: Introduce `Constitution` slash command now

**Deferred (EX-6)**: This SPEC references `Constitution` for vocabulary alignment but does NOT introduce a new slash command. A follow-up SPEC MAY introduce `/moai constitution` if the user wants Spec Kit-style governance commands. Keeping this SPEC's scope tight avoids scope creep.

## §E. Risks & Mitigations (Design-Level)

- **DR-1 (HIGH)**: The era.go rewrite touches the most-loaded function in the spec subsystem. Mitigation: dual-predicate window (REQ-LR-006) + dedicated test fixtures (M2) + `moai spec audit --json` regression check at each milestone.
- **DR-2 (MEDIUM)**: The backfill script (M3) mutates 48 progress.md files. Mitigation: migration log, git checkpoint before backfill, and a rollback path (the legacy predicate remains valid during the window).
- **DR-3 (LOW)**: The Epic rename could confuse existing memory entries that reference "Sprint N". Mitigation: memory files are out of scope (EX-4); `[SUPERSEDED]` markers in new memory entries handle the transition.
- **DR-4 (LOW)**: The `Constitution` term introduction (without a command) could be perceived as incomplete. Mitigation: EX-6 explicitly defers the command to a follow-up SPEC; the term is documented for vocabulary alignment only.

## §F. Open Design Questions (Deferred)

- **ODQ-1**: Should the legacy fallback branch in `ClassifyEra` (REQ-LR-006) emit an `INFO` finding recommending backfill migration? **Recommendation: YES** — surfaces the 48 SPECs that still need M3 backfill. Deferred to M2 implementation.
- **ODQ-2**: Should the Epic term carry a sequence number (`Epic 1`, `Epic 2`) or a thematic name (`Epic Docs-v3`)? **Recommendation: BOTH allowed** — sequence for time-unit Epics, thematic name for release-cut Epics. Matches the former Sprint flexibility.
- **ODQ-3**: Should the backfill script also migrate the 270 grandfather-protected SPECs if they carry §E.5? **Recommendation: NO** (N4 + grandfather clause). The backfill is scoped to the 48 V3R6 SPECs only.

## §G. Cross-References

- `spec.md` §D (REQ-LR-001..019) — requirements this design operationalizes.
- `plan.md` §F (M1-M8) — milestones that implement this design.
- `acceptance.md` §C (AC-LR-001..011) — criteria that verify this design.
- `research.md` §C.3 (RQ-3) + §D.1 — era migration impact data + strategy selection rationale.
- `internal/spec/era.go` doc-comment lines ~86-101 + function body lines ~117-146 — the `ClassifyEra` doc + logic to be rewritten (M1).
- `internal/spec/audit.go` lines 251-297 (all three §E.5-keyed findings) + lines 54-65 (`FindingType` constants) — the `checkV3R6Drift` branches to be retired/re-anchored (M2).
- `internal/spec/transitions.go` lines 73-83 (`closeInfix4Phase` const + `closeInfixMatch`) — close-infix matcher to gain `"3-phase close"` while retaining `"4-phase close"` for legacy commits (M2, per §B.6 / D4).
- `.claude/rules/moai/workflow/lifecycle-sync-gate.md` lines ~28/~43/~228/~235/~303 — era-definition row, H-4 heuristic row, ownership-matrix row, close-subject mandate, and `## §E.5 Mx-phase` worked example (M4, per §A.1(b) / D5).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` lines ~71/~78 — ownership-matrix row + close-subject mandate (M4, per §B.6 / D4).
- `.claude/rules/moai/development/sprint-round-naming.md` — the SSOT to be rewritten (M6).
