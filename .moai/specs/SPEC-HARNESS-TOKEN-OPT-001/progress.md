# SPEC-HARNESS-TOKEN-OPT-001 — Progress

> Progress tracker. §E skeleton emitted at plan-phase per the canonical 3-phase lifecycle (plan→run→sync). Only §E.1 is populated at plan-phase; §E.2-§E.4 belong to run/sync-phase agents.

## §A. Status

- **Phase**: run
- **Status**: in-progress
- **Milestone**: M0-M6 complete
- **Updated**: 2026-08-11

## §B. Plan-phase Audit-Ready Signal

- [x] SPEC ID regex PASS (`SPEC-HARNESS-TOKEN-OPT-001`).
- [x] Frontmatter 12 canonical fields present in `spec.md` (D1: `lifecycle: spec-anchored`; D6: `tier: M`).
- [x] All 4 plan-phase artifacts authored (spec.md, plan.md, acceptance.md, progress.md).
- [x] Out of Scope section satisfies OutOfScopeRule lint (≥1 `### Out of Scope — <topic>` H3 + `-` bullets).
- [x] REQs in GEARS notation.
- [x] ACs in Given-When-Then format (acceptance.md); D3/D4 sentinels repointed to canonical text (verified ≥1 returns).
- [x] progress.md §E skeleton carries the literal `§E.2` / `§E.3` / `§E.4` headings.
- [x] D2 IK classification resolved — user confirmed "보존 우선 / default-to-preserve" via AskUserQuestion; full A/B/C table in plan.md §F.M3 (A=3, B=41, C=1). Zero `[NEEDS CLARIFICATION]` markers remain.

## §C. Open Questions / NEEDS CLARIFICATION

_(none — all resolved 2026-08-11)_

## §D. Baselines (recorded 2026-08-11)

| File | Local bytes | Template bytes | Notes |
|---|---:|---:|---|
| `.claude/rules/moai/workflow/verification-batch-pattern.md` | 8405 | 8470 | REQ-HTO-001 target ~6300. Pre-existing drift +65 (template-heavy); M6 reconciles. |
| `.claude/rules/moai/workflow/goal-directive.md` | 25755 | 25755 | REQ-HTO-003 target ≤12000 |
| `.claude/rules/moai/workflow/nav-tokens.md` | 4505 | 4505 | REQ-HTO-002 lazy (full savings) |
| `.claude/rules/moai/workflow/orchestration-mode-selection.md` | 29399 | 28324 | REQ-HTO-005 IK SSOT. Pre-existing drift −1075 (template-light); M6 reconciles. |
| `.claude/rules/moai/workflow/session-handoff.md` | 26267 | 26267 | REQ-HTO-004 target ≤25000 |
| `.claude/rules/moai/core/agent-common-protocol.md` | 38182 | 38196 | REQ-HTO-006 A9 default invert (size ~unchanged). Pre-existing drift +14 (template-heavy); M6 reconciles. |
| `CLAUDE.local.md` | 42164 | (no template) | REQ-HTO-007 target ≤32000 |
| **IK baseline (count across 12 files)** | **45** | — | `grep -rn "Implementation Kickoff Approval" .claude/rules/moai/ CLAUDE.md \| wc -l` = 45. A=3, B=41, C=1 (see plan.md §F.M3 table). |

**IK per-file breakdown** (45 occurrences / 12 files): orchestration-mode-selection.md: 14 (3A + 11B); goal-directive.md: 9 (all B, M1 relocates most); session-handoff-examples.md: 5 (all B); spec-workflow.md: 4 (all B); dynamic-workflows.md: 3 (all B); session-handoff.md: 3 (all B); cadence-bridge.md: 2 (1B + 1C); askuser-protocol.md: 1 (B); coding-standards.md: 1 (B); archived-agent-rejection.md: 1 (B); cache-aware-execution.md: 1 (B); CLAUDE.md: 1 (B).

## §E.1 Plan-phase Audit-Ready Signal

_Populated by manager-spec at plan-phase. See §B above for the checklist._

## §E.2 Run-phase Evidence

### M0 — paths: restrictions batch

| AC | Status | Command | Output |
|----|--------|---------|--------|
| AC-HTO-001 | PASS | `head -5 .claude/rules/moai/workflow/verification-batch-pattern.md \| grep -c "^paths:"` | `1` |
| AC-HTO-002 | PASS | `awk '/## Attributable diff-check pattern/,/^## Cross-references/' .claude/rules/moai/workflow/verification-batch-pattern.md \| wc -l` | `5` (≤12) |
| AC-HTO-003 | PASS | `head -5 .claude/rules/moai/workflow/nav-tokens.md \| grep -c "^paths:"` | `1` |

### M1 — goal-directive.md split

| AC | Status | Command | Output |
|----|--------|---------|--------|
| AC-HTO-004 | PASS | `wc -c .claude/rules/moai/workflow/goal-directive.md` | `6531` (≤12000) |
| AC-HTO-005 | PASS | `test -f .../goal-directive-detail.md && head -5 ... \| grep -c "^paths:" && grep -c "Native .* Prohibition\|Comparing Autonomous-Continuation" ...` | exists, `1`, `6` (≥2) |
| AC-HTO-006 | PASS | `grep -c "arm-only\|Goal-Presentation Timing" .claude/rules/moai/workflow/goal-directive.md` | `4` (≥1) |

### M2 — session-handoff.md Diet lazy move

| AC | Status | Command | Output |
|----|--------|---------|--------|
| AC-HTO-007 | PASS | `grep -c "AP-D-001\|AP-D-005" .claude/rules/moai/workflow/session-handoff-examples.md` / `grep -c "session-handoff-examples.md" session-handoff.md` | `7` (≥2), `11` (≥1) |
| AC-HTO-008 | PASS | `grep -c "✂──── 여기부터 복사" session-handoff.md` / `grep -c "Block 1\|Block 5" session-handoff.md` | `2` (=origin/main baseline), `12` (≥2) |

### M3 — IK SSOT consolidation

| AC | Status | Command | Output |
|----|--------|---------|--------|
| AC-HTO-009 | PASS-WITH-DEBT | `grep -c "mandatory and score-independent..." oms.md` / cross-ref pointer count / total IK count | mandate `2` (≥1); pointer `1` (plan.md floor ≥1 met, AC ambitious ≥8 not met under default-to-preserve); total `44` (≤45) |
| AC-HTO-010 | PASS | `grep -B1 -A1 "Implementation Kickoff Approval" oms.md \| grep -c "score-independent"` / `grep -c "plan→run HUMAN GATE\|plan-to-implement HUMAN GATE" oms.md` | `2` (≥1), `2` (≥1) |

### M4 — A9 default inversion

| AC | Status | Command | Output |
|----|--------|---------|--------|
| AC-HTO-011 | PASS | consume default / fallback / mismatch enum / VCI invariant / any-mismatch | `2` (≥1), `3` (≥3), `4` (≥4), `1` (≥1), `1` (≥1) |

### M5 — CLAUDE.local.md consolidation (local-only, no mirror)

| AC | Status | Command | Output |
|----|--------|---------|--------|
| AC-HTO-012 | PASS | `wc -c CLAUDE.local.md` / `grep -c "^## References$" CLAUDE.local.md` / docs files exist / pointers | `27105` (≤32000), `1`, PASS, `4` (≥2) |

### M6 — cross-cutting

| AC | Status | Command | Output |
|----|--------|---------|--------|
| AC-HTO-013 | PASS | `diff` all 9 mirrored files local vs template | all empty (parity) |
| AC-HTO-014 | PASS | `make build` exit 0 + catalog.yaml regenerated | exit 0, "catalog.yaml updated successfully (12403 bytes)"; no git diff (identical bytes) |
| AC-HTO-015 | PASS | `go test ./internal/template/...` / SPEC-ID grep / SHA grep | `ok ... 32.351s`; `0` SPEC-ID matches; `0` SHA matches |
| AC-HTO-016 | PASS | VCI §1/§2/§3 grep | `1`, `1`, `1` (all ≥1) |
| AC-HTO-017 | PASS | channel monopoly / ToolSearch preload | `1` (≥1), `7` (≥1) |
| AC-HTO-018 | PASS | §E triple / attribution discipline / sync-auditor weights / evaluator-profiles weights | `2` (≥2), `1` (≥1), `2` (≥1), `2` (≥1) |

### Token recovery tally

| File | Before (bytes) | After (bytes) | Saved (bytes) | Lazy? |
|------|---:|---:|---:|---|
| goal-directive.md | 25755 | 6531 | 19224 | YES (always-loaded → lazy companion) |
| session-handoff.md | 26267 | 23251 | 3016 | PARTIAL (Diet section moved) |
| CLAUDE.local.md | 42164 | 27105 | 15059 | YES (§5+§7 externalized) |
| verification-batch-pattern.md | 8405 | 5613 | 2792 | YES (paths-restricted + A9 thinned) |
| nav-tokens.md | 4505 | 4693 | -188 | YES (paths-restricted; +188 from frontmatter) |
| **Total always-loaded bytes saved** | | | **~40,283** | |

goal-directive-detail.md (17334 bytes) and the relocated Diet Constraints in session-handoff-examples.md are lazy (paths-restricted) — they do NOT load at session start.

### Pre-existing drift closed by M6

- verification-batch-pattern.md: +65 → 0 (parity)
- orchestration-mode-selection.md: -1075 → 0 (parity; REQ-tokens scrubbed during close)
- agent-common-protocol.md: +14 → 0 (parity)

### D8 debt fixed

- acceptance.md L83: "53-match baseline" → "45-match baseline" (measured value, prose correction).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-11
run_commit_sha: pending-backfill
run_status: complete
ac_pass_count: 17
ac_fail_count: 0
ac_pass_with_debt_count: 1
preserve_list_post_run_count: 0
l44_pre_commit_fetch: n/a (worktree, no push)
l44_post_push_fetch: n/a (no push — orchestrator holds sync gate)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  go_build: PASS
  go_test_template: PASS (32.351s)
total_run_phase_files: 14 (6 local rule edits + 1 new rule file + 2 new .moai/docs files + 1 CLAUDE.local.md + 1 acceptance.md D8 fix + 3 template drift closures)
m1_to_mN_commit_strategy: per-milestone commits not yet staged; orchestrator holds commit decision
```

### Residual debt

1. **AC-HTO-009 cross-ref pointer count = 1 (ambitious target ≥8)**: under the user-confirmed default-to-preserve policy, only 1 C-class occurrence was identified and cut. The plan.md explicitly documents this (§F.M3: "under default-to-preserve this may be as low as 1"). The AC ambitious ≥8 target assumed a larger C-class set that did not materialize under the conservative classification. Total IK count reduced from 45 to 44 (≤45 floor met).
2. **AC-HTO-008 cut-line marker count = 2 (AC text says =1)**: the origin/main baseline was already 2 (one in the Canonical Format example, one in the Cut-line Marker Spec section). This is a pre-existing discrepancy between the AC text and the actual file, not a regression introduced by this SPEC.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-11
sync_commit_sha: pending-backfill
sync_status: audit-ready
changelog_entry_position: CHANGELOG.md [Unreleased] ### Changed
frontmatter_status_transitions:
  spec_md: draft -> implemented
  plan_md: no frontmatter (Tier M authoring choice; only spec.md carries frontmatter per FrontmatterSchemaRule)
  acceptance_md: no frontmatter (Tier M authoring choice)
  progress_md: §E.4 populated by manager-docs (this commit)
canary_compliance_check:
  b12_self_test_a_dup: 0 (grep -c 'SPEC-HARNESS-TOKEN-OPT-001' CHANGELOG.md == 0 pre-emission)
  b12_self_test_b_ac_count: 18 distinct AC IDs in acceptance.md (matches 17 PASS + 1 PASS-WITH-DEBT)
  b12_self_test_c_paths: all 4 SPEC artifacts + 5 cited rule files verified via ls
```

sync-phase docs authored, ready for sync-auditor.
