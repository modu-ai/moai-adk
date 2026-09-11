# Progress — SPEC-UPDATE-SETTINGS-BASE-SNAPSHOT-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready (final revision under the operator-approved one-time extension)
plan_complete_at: 2026-09-11
card: t656
tier: M
artifact_count: 4 (spec.md, plan.md, acceptance.md, progress.md)
spec_version: 0.4.0
era: V3R6
base_tree: 04a8ab731 (internal/ identical to 81c1d58f9 — git diff --stat empty)
branch: WT-update-value-merge
depends_on:
  - SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001
counts:
  requirements: 16 (Tier M ceiling 16)
  acceptance_criteria: 16 (Tier M ceiling 16)
  mutant_rows: 27
  needs_clarification_markers: 0
audits:
  iter1: {verdict: FAIL, score: 0.73, report: .moai/reports/t656/plan-audit-iter1.md}
  iter2: {verdict: FAIL, score: 0.75, report: .moai/reports/t656/plan-audit-iter2.md}
  extension: operator-approved one-time third revision 2026-09-11
decisions:
  A1_B1_C1: confirmed (operator, before plan phase)
  F05_capture_only_when_deploy_wrote: confirmed 2026-09-11 — REQ-USB-016, AC-USB-014; D7 prefers manifest provenance + hash (N-11)
  D5_promotion_rule: decided 2026-09-11 — refined rule; REQ-USB-005 restated as three GEARS sentences (N-01); AC-USB-016 adopted with R3 != R2 (N-03)
  D5_record: option (a)'s purpose preserved; earlier abort explanation superseded (wrong premise)
  D5_implementation: "운영자 규칙의 구현 방식, 리드 수용 (2026-09-11)" — two-point judgement; leftover judged before any step of the next flow removes or rewrites the live .claude/settings.json (N-02, operator direction); positions update.go before :384, init.go before :867; supersedes "before the next flow writes its own staging copy"
  D5_normal_end_signal: merge preserve path taken (merge.go :197-204, :217-225, :229-237), not byte compare (N-10)
  D5_disclosed_deviation: an intervening write to the live file after an abort turns promote into discard (fail-safe) — spec §E, plan D5 (N-08)
  D6_sibling_relation: decided 2026-09-11 — sibling REQ-UMC-010 scoped; wording unchanged this round
  F17_user_deleted_key: accepted 2026-09-11 as a known limitation — spec §B.5, §E
run_phase_verification_items:
  - M1: whether RestoreMoaiConfig writes .claude/settings.json (read so far: restore_entry.go:47-79 calls only RestoreMoaiConfig; auditor read restore.go as .moai/config-only)
red_now_observed: none (plan phase forbids go test — recorded at run-phase M1)

## §E.2 Run-phase Evidence

Run phase started 2026-09-11 (Implementation Kickoff Approval granted by the operator via the lead). Worktree `.claude/worktrees/t656`, branch `WT-update-value-merge`, run base HEAD `41a470641` (local develop `0db675bed` absorbed). Long outputs live under `.moai/reports/t656/run/`.

### §E.2.0 Plan-audit iteration-3 minors (N3-01..N3-06) — disposition

The SPEC ownership matrix (`.claude/rules/moai/development/spec-frontmatter-schema.md` § Forbidden ownership crossings) forbids manager-develop from editing `spec.md` / `plan.md` / `acceptance.md` body content. Every N3 fix below is a body edit of `plan.md` or `acceptance.md`, so none is applied to those files in the run phase. Each is recorded as **debt for manager-spec** together with the run-phase handling that keeps the implementation and its evidence correct without the wording change.

| id | Where | Wording debt (manager-spec) | Run-phase handling (no SPEC body edit) |
|---|---|---|---|
| N3-01 | acceptance.md:124, :285 (M-D5g-w) | Split the row into M-D5g-wb (judgement moved to the backup step → killed only by `update_leftover_version_skip`) and M-D5g-wd (judgement moved after deploy → killed by `update_leftover_abort` `a == 2` and by `update_leftover_version_skip`) | The slot-request mutant list carries the two variants as separate mutants with their own killing cells |
| N3-02 | acceptance.md:115-116 | Add a retired deny entry to the `update_leftover_version_skip` cell and a mutant row M-D5g-s | The run-phase test gives the leftover R2 and the live file the retired entry `Write(./secrets/**)` (a strictly stronger fixture of the same cell); M-D5g-s is in the slot-request mutant list |
| N3-03 | plan.md:151 (D3) vs :207 (M3) | Name the preserve-path signal channel in D3 | Implemented as a sibling function `MergeUserFilesWithOutcome` that returns a per-path outcome; `MergeUserFiles` keeps its signature and becomes a thin wrapper, so D3 ("signature and base injection unchanged") still holds |
| N3-04 | plan.md:208 vs :217, acceptance.md:81 | State the pre-merge observation hook in one milestone | The hook is introduced once, in the `internal/cli` wiring (M4), next to its only consumer AC-USB-005 — consistent with acceptance.md:81. AC-USB-006/016 observe between flows and need no hook |
| N3-05 | acceptance.md:232, :235 | Relabel the c2/c5 next-flow cells: they are green only against the empty-promotion stub with base selection in place, not before implementation | The RED/GREEN evidence below records these cells against that stub, not against the pre-implementation tree |
| N3-06 | acceptance.md:135-137 | Optional: add a leftover-promotion-failure cell | A `backup` package test drives the leftover judgement with a directory planted at the canonical path and asserts a nil-free, non-blocking result with exactly one `settings-snapshot-promote-failed:` line; the `runUpdate` call-site cell is in the slot request |

### §E.2.1 M1 — baseline (this run, tree `41a470641`)

```
$ go test ./internal/cli/update/merge/... ./internal/cli/update/backup/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli/update/merge	0.656s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	1.323s
exit=0

$ go test -cover ./internal/cli/update/merge/ ./internal/cli/update/backup/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli/update/merge	0.581s	coverage: 92.1% of statements
ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	1.197s	coverage: 90.1% of statements
exit=0

$ grep -n "^\.moai/cache/$" .gitignore internal/template/templates/.gitignore
internal/template/templates/.gitignore:241:.moai/cache/
.gitignore:352:.moai/cache/
```

Coverage baseline for the DoD "no lower than M1": merge 92.1%, backup 90.1%.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
