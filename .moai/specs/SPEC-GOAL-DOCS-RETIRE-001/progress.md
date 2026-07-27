---
id: SPEC-GOAL-DOCS-RETIRE-001
title: Retire native /goal emission references from public and internal documentation across four locales
version: 1.0.0
status: draft
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
priority: MEDIUM
phase: "v3.1.0"
module: docs
lifecycle: spec-anchored
tags: "goal, docs-site, i18n, locale-parity, split-surface"
tier: M
depends_on: [SPEC-GOAL-SURFACE-UNIFY-001]
---

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifact set complete (Tier M): `spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`, `progress.md`.

Origin: split from `SPEC-GOAL-SURFACE-UNIFY-001` after that SPEC's plan-audit iteration 2 returned FAIL 0.64 with a STOP signal (score regression from 0.71), and the user chose the auditor's scope-reduction proposal over iteration 3.

- SPEC ID regex check executed as Bash; observed output `SPEC-ID-CHECK: PASS`.
- ID confirmed unused before authoring (`ls .moai/specs/SPEC-GOAL-DOCS-RETIRE-001` → `No such file or directory`).
- **12** acceptance criteria; every judgment command executed and its verbatim baseline recorded. All 12 fail at baseline.
- **Every emission criterion carries a per-locale baseline**, and all five emission detectors measure symmetric (`distinct=1`). This is the requirement that closes the parent's audit finding N2.
- 11 REQs, all cited by ≥ 1 AC (`acceptance.md` §E). (10 at authoring; REQ-GDR-011 added closing audit iteration 2 finding B2-1.)
- 13 paths, single-owner across N1-N5 (`plan.md` §F.1). 41 retained files owned by no milestone.
- Four retention surfaces registered in this SPEC's own register (`plan.md` §A.2) — not inherited from the parent.

Design decisions recorded rather than deferred:

| Decision | Rationale |
|---|---|
| **Tier M**, not L | 13 files inside the 15-file ceiling; one layer; one verification regime (locale-symmetric grep repeated four times is the same regime, not four); fully reversible; no irreversibility surface. The split-surface difficulty is handled by plan-time detector design, not by added implementation layers |
| `.moai/docs/autonomous-workflow-strategy.md` belongs **here**, not in the parent | Its treatment (superseding note, sweep nothing) is a sync-phase documentation action and the parent has no sync phase post-split; its content is the same 3-engine comparison as `autonomous-loops.md`; its membership-test outcome matches the other retention rows. Counter-argument acknowledged: `.moai/docs/` is not public — but the retain-plus-note treatment does not depend on publication |
| `autonomous-loops.md` is a **split surface**, judged positively on both halves | A file-level `0` target would delete two sections, a comparison row, and the Axis-B justification in four locales — the parent's D4/N2 defect |
| AC-GDR-010 exists as a **meta-guard** | Per-detector locale diligence is not self-verifying; a criterion whose subject is the detectors makes the false-zero class mechanically impossible |

Carried-identifier provenance is recorded at `spec.md` §D; the parent's moved-identifier register is at `SPEC-GOAL-SURFACE-UNIFY-001/spec.md` §B.7.

### Plan-audit iteration 1 — FAIL 0.75 (Tier M threshold 0.80), remediated

The first audit confirmed Tier M correct, reproduced all 12 baselines with zero divergence, and found the retention register consistent and fully guarded. Both MUST-FIX were defects in AC-GDR-010's *specification*, not its measurement:

| Finding | Resolution |
|---|---|
| **B-1** — `distinct=1` admits a dead detector: a typo'd anchor reads `0/0/0/0` → `distinct=1` and passes, indistinguishable post-sweep from a genuine sweep | AC-GDR-010 gained a **liveness component (b)**: every detector must match non-zero content against the immutable base `e306e21a9`. New REQ-GDR-010. Dead-detector control executed: the typo'd anchor reads `live_min=0` and is rejected |
| **B-2** — the blanket symmetry rule contradicts REQ-GDR-008, with `hooks-reference.md` (`en:1 ja:0 ko:1 zh:0` → `distinct=2`) as a live example | REQ-GDR-004 split: per-locale recording stays unconditional; symmetry moved to new **REQ-GDR-009**, conditioned on content symmetry with a named-exemption carve-out. New §A.3 states both causes of asymmetry. Exemption list currently empty |
| **B-3** — §D provenance table mapped 4 of 7 rows to the wrong criterion | All rows re-derived from each criterion's actual subject; fixed jointly with the parent's A-1 so the two registers agree |
| **B-4** — `spec.md` stated 18 emission markers vs the measured 24 | Corrected to the command-derived `24`; the stale `18` used the disqualified `per-turn` detector's `2` and omitted L7's `4` |
| **B-5** | §A.1 now marks AC-GDR-004/005/006 as deliberate presence checks |
| **B-6** | Beyond-minimum artifact set documented as deliberate; tier stays M |
| **B-7** | AC-GDR-004's fallback re-specified as the locale-invariant triple co-occurrence, replacing the prose-adjacent `provides` (measured `1/1/1/1`) |

### Plan-audit iteration 2 — PASS 0.86 (Tier M threshold 0.80), MUST-FIX B2-1 closed

Iteration 2 reproduced 12 of 12 baselines with zero divergence and scored Clarity 0.90 / Completeness 0.85 / Testability 0.80 / Traceability 0.90 (harmonic mean 0.8605). One MUST-FIX and two SHOULD-FIX were raised.

| Finding | Severity | Resolution |
|---|---|---|
| **B2-1** — AC-GDR-010 validated liveness and symmetry but not *aptness*: a detector aimed at an adjacent token (`ac_converge`) or a compound regex with its `` `/goal` `` half dropped passes (a) and (b) while matching no emission reference | MUST-FIX | **CLOSED.** New **REQ-GDR-011**; AC-GDR-010 gained component **(d)**, asserting each detector's pattern carries a literal `/goal` token from a single `p=` source shared by `w()` and the assertion. Three controls executed: the hyphenated dead detector is rejected by (b); `ac_converge` and the weakened `auto mode` half are rejected by (d) |
| **B2-2** — liveness against a frozen base can be stale-true; no criterion asserts the base is still current for the scoped pages | SHOULD-FIX | Open — deferred as follow-on |
| **B2-3** — the empty exemption list is guarded by prose, not by a check | SHOULD-FIX | Open — deferred as follow-on |

**Auditor's stronger alternative for B2-1 was executed and refuted.** The audit proposed asserting each detector's base-match line set is a subset of the base lines carrying `` `/goal` ``. Run against `e306e21a9`, both attack detectors yield `leak=0` and pass: `ac_converge` and `auto mode` occur on the *same line* as `` `/goal` `` (`en/advanced/autonomous-loops.md` line 7). Line granularity is precisely the granularity the adjacent-token attack exploits, so the subset form cannot discriminate. The pattern-literal form was adopted instead, with the single-`p=`-source discipline closing its own divergence risk.

Residual (not in B2-1's required fix): AC-GDR-012's aggregate still carries its own inline copies of the five detector definitions, so the single-source discipline does not extend to the aggregate path.

N3 remains blocked on the parent's M7 landing (`acceptance.md` §C dependency gate) — the parent is now `completed`, so the gate is expected to clear at run-phase entry and must be re-measured there.

## §E.2 Run-phase Evidence

All 12 baselines were re-measured at run-phase entry against `24c84c56e` (the squash-merge commit for PR #1176; the original run-phase commit `d54ea108d` evaporated from git history via squash merge) and reproduced the values recorded in `acceptance.md` §B exactly (12/12), so every row below is a measured transition, not an assumed one.

### AC matrix

| AC | Judgment command (per `acceptance.md` §B) | Baseline | Actual output | Status |
|---|---|---|---|---|
| AC-GDR-001 | paired listing, `autonomous-loops.md`, per locale | `en:1 ja:1 ko:1 zh:1` | `en:0 ja:0 ko:0 zh:0` | PASS |
| AC-GDR-002 | paired listing, `self-evolving.md`, per locale | `en:2 ja:2 ko:2 zh:2` | `en:0 ja:0 ko:0 zh:0` | PASS |
| AC-GDR-003 | `` `/goal `` in `handoff.md`, per locale | `en:1 ja:1 ko:1 zh:1` | `en:0 ja:0 ko:0 zh:0` | PASS |
| AC-GDR-004 | line-7 mis-attribution, per locale | `en:1 ja:1 ko:1 zh:1` | `en:0 ja:0 ko:0 zh:0` | PASS |
| AC-GDR-005 | `auto mode` + `` `/goal` `` co-occurrence, per locale | `en:1 ja:1 ko:1 zh:1` | `en:0 ja:0 ko:0 zh:0` | PASS |
| AC-GDR-006 | split-surface structural pins, per locale | `h3=1,h2=1,row=1` ×4 | `en:h3=1,h2=1,row=1 ja:h3=1,h2=1,row=1 ko:h3=1,h2=1,row=1 zh:h3=1,h2=1,row=1` | PASS |
| AC-GDR-007 | superseding sentinel / content pin | `0` / `25` | `1` / `26` | **PASS-WITH-DEBT** — see below |
| AC-GDR-008 | five retention pins | `cc=80 goal.md=12 moai-goal=4 hooks=2 research=4` | `cc=80 goal.md=12 moai-goal=4 hooks=2 research=4` | PASS |
| AC-GDR-009 | four-locale file inventory | `autonomous-loops.md=4 handoff.md=4 self-evolving.md=4` | `autonomous-loops.md=4 handoff.md=4 self-evolving.md=4` | PASS |
| AC-GDR-010 | locale-invariance meta-guard (a/b/d) | all five `distinct=1,live_min>=1,apt=1` | `paired_al:distinct=1,live_min=1,apt=1 auto_mode:distinct=1,live_min=1,apt=1 l7:distinct=1,live_min=1,apt=1 paired_se:distinct=1,live_min=2,apt=1 handoff:distinct=1,live_min=1,apt=1` | PASS |
| AC-GDR-011 | docs-site build | `exit=0` | `exit=0` | PASS |
| AC-GDR-012 | aggregate emission | `total=24` | `total=0` | PASS |

### Invariants

| Invariant | Evidence | Status |
|---|---|---|
| Retention register row 2 intact (native H3, `Native /goal Details` H2, comparison row, factual statements) | `grep -n '`/goal`'` on `autonomous-loops.md` returns exactly the table row, the H3, and the four factual lines in every locale — `en:7 ja:7 ko:7 zh:7`, symmetric | HOLDS |
| No new locale pages (REQ-GDR-008) | AC-GDR-009 inventory unchanged at 4 each; `hooks=2` pin unchanged | HOLDS |
| Scope discipline (13 paths) | `git diff --stat 24c84c56e..HEAD` = 12 docs-site files + `.moai/docs/autonomous-workflow-strategy.md` + `spec.md` frontmatter only; zero changes under `docs-site/content/*/claude-code/`, `.moai/research/`, `.claude/`, `internal/template/templates/` | HOLDS |
| `--goal` CLI flag name unchanged | `grep -c '`--goal <condition>`'` = 1 per locale | HOLDS |
| `handoff.md` L51 `/cli-reference/goal` link path survives (D6) | `grep -c "\[moai goal\](/<locale>/cli-reference/goal)"` = 1 per locale | HOLDS |
| Locale parity (4-locale same-PR obligation) | Every swept marker reached its target in all four locales simultaneously; no detector shows a locale split (`distinct=1` on all five) | HOLDS |
| Build warning-free | `hugo --minify --gc` exit 0, `grep -cE 'WARN\|ERROR'` = 0, sitemap present | HOLDS |
| URL blacklist / Mermaid TD-only | both greps clean | HOLDS |

### AC-GDR-007 debt — criterion internal contradiction (off-by-one) — **CLOSED**

The two components cannot both hold as literally written. Component 1 requires the file to contain the literal sentinel `native `/goal` emission is retired`; that phrase itself contains a backticked `/goal`, so satisfying it necessarily raises the component-2 occurrence count from 25 to 26. Measured on a scratch copy before editing: adding the sentinel alone yields `sentinel=1, pin=26`.

The milestone's intent — annotate the record, sweep nothing — is met and verifiable: the diff is `1 insertion, 0 deletions` (`git diff --numstat`), and excluding the note line the count is exactly `25` (all historical occurrences byte-identical). The criterion's target value for component 2 is off by one; it should be `26` (25 historical + the sentinel) or the sentinel should be excluded from the count. `acceptance.md` was not modified during the run phase — the correction was manager-spec's to make.

**Resolution (manager-spec, 2026-07-27).** The exclusion form was adopted: component 2's detector now filters the sentinel line out before counting, leaving the pin at `25`. The raise-to-`26` alternative was declined — `26` is a composite (25 historical + 1 sentinel) that obscures what the pin asserts.

```bash
grep -ciF 'native `/goal` emission is retired' .moai/docs/autonomous-workflow-strategy.md
grep -vF 'native `/goal` emission is retired' .moai/docs/autonomous-workflow-strategy.md | grep -ohF '`/goal' | wc -l | tr -d ' '
```

Three validating measurements, all run before the edit was delegated:

| Measurement | Result |
|-------------|--------|
| Post-edit working tree, exclusion-form count | `25` (the target) |
| Sentinel component (component 1) | `1` (satisfied, target `>= 1`) |
| Pre-edit base — `git show origin/main:.moai/docs/autonomous-workflow-strategy.md`, exclusion form | `25` (recorded baseline unchanged) |

Neither the recorded baseline (`0` / `25`) nor the target (`>= 1` / `25`) changed; only the second detector's command did. The criterion's semantics ("all 25 historical occurrences survive") are preserved, and it is now satisfiable. Recorded in `spec.md` HISTORY `1.3.0`. AC-GDR-007 is now satisfiable and satisfied; the §E.3 signal below is the frozen run-phase snapshot and still records it as the one PASS-WITH-DEBT row.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-27
run_commit_sha: pending-backfill-run-final
run_status: audit-ready
ac_pass_count: 11
ac_fail_count: 0
ac_pass_with_debt_count: 1
preserve_list_post_run_count: 5
l44_pre_commit_fetch: not-run
l44_post_push_fetch: not-run
new_warnings_or_lints_introduced: 0
cross_platform_build:
  applicable: false
  reason: documentation-only SPEC; no Go source in scope
docs_build:
  command: hugo --minify --gc
  exit: 0
  warnings: 0
  sitemap: present
total_run_phase_files: 14
m1_to_mN_commit_strategy: one commit per milestone (N1..N4); N5 verification edits nothing
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-07-27
sync_commit_sha: pending-backfill-sync-2
sync_status: audit-ready
docs_build:
  command: hugo --minify --gc
  exit: 0
  warnings: 0
  sitemap: present
changelog_entry: yes
amendment_sync: true
amendment_version: "1.5.0"
run_phase_commits:
  - sha: "449c7cb28"
    description: "plan-phase amendment spec.md frontmatter + ## Amendments + plan.md rewrite"
  - sha: "f683675b3"
    description: "audit SHOULD-FIX D1/D2 + NIT D3-D7"
  - sha: "115b0b54e"
    description: "M1 AC-GDR-012 refactor (acceptance.md single-p= source + liveness + aptness guards)"
verification:
  independent_check: true
  tree_total: 0
  spec_lint_errors: 0
  e1_e7_pass: true
orphan_sha_mapping:
  - original: d54ea108d
    mapped_to: 24c84c56e
    rationale: "PR #1176 was squash-merged; d54ea108d evaporated from git history (verified via git merge-base --is-ancestor exit 1)"
```

**Sync-phase notes**: This is the amendment sync-phase (v1.5.0 D2 aggregate liveness/aptness guard). The run-phase M1 commit (`115b0b54e`) was independently verified this session with all E1-E7 checks passing. The sync_commit_sha will be backfilled in a follow-up commit per the D3 placeholder-backfill exemption (a commit cannot reference its own SHA).

**Mapping rationale**: The run-phase commit `d54ea108d` was squash-merged into PR #1176 (commit `24c84c56e`). Git no longer recognizes `d54ea108d` as an ancestor of the current tree (verified by `git merge-base --is-ancestor d54ea108d origin/main` exiting 1), confirming evaporation. All references in `progress.md` have been mapped to the squash-merge commit `24c84c56e`, which is the authoritative baseline for the run-phase work.

## §F Phase 4 Mode Selection

**Decision: sub-agent** (Mode 5, sequential per-milestone delegation)

**Input parameters**

| Signal | Value |
|---|---|
| tier | M (3-artifact set; 5 authored) |
| scope (file count) | 12 scoped locale files across 3 pages (sweep targets) + 1 annotate-only strategy record = 13 |
| domain count | 1 — docs-site markdown content |
| file language mix | 100% markdown (4 locales: en / ja / ko / zh) |
| concurrency benefit | LOW — per-locale prose differs, and 4-locale parity requires coordinated edits within one reasoning context |

**Mode evaluation**

| Mode | Selected | Rationale |
|---|---|---|
| 1 `trivial` | no | Multi-file semantic sweep with 12 acceptance criteria; not a typo class |
| 2 `background` | no | Work is write-heavy (Edit on scoped pages), not read-only analysis |
| 3 `agent-team` | no | RETIRED tombstone; never selected by the decision tree |
| 4 `parallel` | no | Single domain (< 3) and edit-heavy, not research-heavy; the coding-task parallelism caveat routes this to Mode 5 |
| 5 `sub-agent` | **YES** | Default fallback; sequential per-milestone delegation preserves 4-locale coordination |
| 6 `workflow` | no | Scope is ~12 files, below the ~30-file entry threshold, and the transform is not one uniform mechanical rule — each locale's surrounding prose differs |

**Decision: sub-agent**

Mode 5 is selected. The sweep is edit-heavy rather than research-heavy, touches one domain, and its hardest constraint is locale parity — four files must change consistently against a shared detector set, which a single sequential agent holds in one reasoning context. Anthropic's coding-task parallelism caveat ("most coding tasks involve fewer truly parallelizable tasks than research") applies directly, and the Mode 6 file-count threshold is not met.

**Boundary case** — none. No signal sat at a threshold ±1: domain count 1 (vs 3), file count ~12 (vs 30), and the transform-kind test fails Mode 6 independently of count.

**Gate provenance** — Implementation Kickoff Approval obtained from the user before this section was written. Progression mode selected at the same gate: **autonomous** (`/moai goal` armed after the gate, per `goal-directive.md` § Goal-Presentation Timing — arming pairs with the work-starting command, never replaces it).
