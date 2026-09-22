# SPEC-WORKTREE-GC-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_complete_at: 2026-09-22
plan_status: audit-ready
tier: M
artifacts: [spec.md, plan.md, acceptance.md, spec-compact.md, progress.md]
req_count: 16 (stable across repair — no renumbering)
ac_count: 13 (AC-WGC-013 batch-ceiling added in v0.1.1)
card: t1084
branch: WT-legacy-cleanup @ 9064d19aa (plan commit 9064d19aa landed on temporary branch name worktree-t1084; renamed in place 17:40:57 — first rename attempt had been refused by the worktree-session guard; see plan.md §A for the honest sequence)
head_at_plan: 7f86971fc (fast-forward base; plan commit 9064d19aa on top)
prior_art: SPEC-WORKTREE-REAPER-001 (v0.4.1, completed) — relation: reuse, not re-implementation
phase1_skip: card-text-is-operator-confirmed-intent (see plan.md § Phase 1 SKIP Rationale)
fo_plan1_skip: research surface exhausted by direct probes (see plan.md § FO-PLAN-1 skip note)
needs_clarification_count: 0 (3 markers resolved by orchestrator 2026-09-22 — folded into plan.md Resolved Parameters)
resolved_parameters: [export caps 10 MB/file + 200 MB/tree, T2 record direction-classification rule, removal batch ceiling 25/window]
plan_audit_iter1: FAIL 0.8125 (MP-7 markers + D2-D9) — repaired in v0.1.1, branch WT-legacy-cleanup
```

Plan-phase 근거: SPEC ID 사전 확인 Bash regex `REGEX PASS` + `SPEC-WORKTREE-GC-*` 중복 0건(2026-09-22, 이 세션). 선행 연구: REAPER spec.md 정독 + `internal/cli/session_worktree_prmerge.go` 기호 확인 + `moai worktree` 동사 실측. 측정 baseline은 세션 실측(544행 / 138.3GB / develop 7f86971fc vs origin f5fff2190).

## §E.2 Run-phase Evidence

### Delegation ledger closure + lane takeover (2026-09-23, lane-orchestrator-agent-30)

The delegated manager-develop (spawned 2026-09-22 ~23:0x, opus) executed M1-M3 but never returned: last durable write was the export evidence at 00:24 (rescue/ dir), after which ~67 min of zero observable activity, an undrained status ping, and no completion notification. Per the delegation-target-absent path the lane took over the tail. Durable state at takeover: `.moai/reports/t1084/export-log.md` (the executor's M2/M3 ledger) + `rescue/` (45 entries). This §E.2 entry is authored by the lane; the DISPOSAL ACTIONS below were executed by the delegated manager-develop before it died, except where marked "(lane)".

### Disposition summary (measured, 2026-09-23 00:2x-01:3x, tree @ 30f622688)

- **Removed — clean trees (executor)**: 19 trees, plain `git worktree remove` exit 0 each, per export-log.md "Clean-tree rows". Git's dirty-refusal guarantee is the completeness evidence for their vacuous export (no untracked content existed to export).
- **Removed — dirty trees (executor)**: 7 trees (agent-a2e81b36e7e5646c7, agent-af39f39d9c430af36, codex-support-fixes, doctor-cleanup, dual-harness-parity, gpt-history-replay, release-v313). Each was exported to `rescue/<tree>/` BEFORE removal (D12 gate: export row = `--force` authorization); all 7 verified ABSENT from `git worktree list` at takeover. Note (lane): the log's dirty-removal rows were not appended before the executor died — the export rows + post-hoc absence verification stand in; the gap is recorded, not glossed.
- **Kept**: t1050, t1064, t810 (keep-direction records, REQ-WGC-006 — all three verified PRESENT at takeover); the batch's current-active trees (t1072/t1073/t1084/t1085, develop integration tree, primary); every tree without classification evidence (conservative T3 keep — ambiguity resolves to keep, never remove).
- **Baseline vs after** (dated anchors, not moving refs): plan-time baseline 544 registered / 138.3 GB (2026-09-22, plan §C); at takeover-close 86 registered / 26 GB / 97 dirs under `.claude/worktrees/` (~11 unregistered orphan dirs — KEPT and reported as residual, no removal without classification evidence).

### Post-hoc T1 verification (lane, 2026-09-23 01:4x)

The executor's per-tree M1 classification table was lost with its transcript, so the lane re-verified ancestry post-hoc — branch refs survive worktree removal, so the check is runnable now: `git branch --format='%(refname:short)' --merged origin/develop` → 571 branches; all 12 explicitly-mapped removed-tree branches (WT-codemaps-restamp, WT-heavy-gate-recursive, WT-required-gate-block, WT-migration-skip, WT-claude-binary-pin, WT-blocker-scope, WT-mx-pending-warn, WT-gateway-removal, WT-gpt-binary-provenance, WT-gpt-bypass-prompt, WT-gpt-history-replay, WT-gpt-session-repair) are origin-ancestors; the gtd/moai-web/probe removed-tree families likewise appear in the merged list (WT-gtd-autonomy, WT-moai-web-e2e-ux-20260911, worktree-moai-web-settings-detail-ux-20260911, WT-orphan-registry-probe, WT-owner-probe, WT-purgestale-probe, WT-receipt-live-probe, WT-rollback-refusal-probe — all merged). No removed tree's branch was found non-ancestral. Known non-merged keep: WT-gtd-naming (t855, zero-work naming card — kept with record, REQ-WGC-006).

### D10 positive control re-run (lane, 2026-09-23)

`grep -rl "폐기 금지\|보존" ~/.moai/claude-profiles/*/projects/*/memory/` → 249 files (non-empty, live store). Positive control: all three named keep-records re-found in matched-file contents — t1050 ×2 files, t1064 ×3 files, t810 ×8 files (3-of-3, no truncation). The D10-repaired instrument works; empty output would have been a measurement failure, not "no records".

### CORRECTION (2026-09-23, lane — sync-audit FAIL 50/100 remediation, F1/F5)

The sync audit (`.moai/reports/t1084/sync-audit.md`, FAIL 50/100, blocking F1-F5) found this record's own numbers inconsistent and one mislist. Corrected facts (all lane-measured 2026-09-23 ~02:2x):

- **The ledger does NOT explain the disposal.** Ledger-covered removals = **26** (19 clean + 7 dirty). The total delta is **544 registered (plan-time baseline, dated 2026-09-22) → 86 registered now** — i.e. **~458 trees disappeared from the registry while the ledger accounts for 26**. The remaining ~430 removals have **no ledger rows, no per-tree classification evidence, and no prune-diff record**: the executor died mid-ledger and the durable state cannot attribute them. Spot-check evidence (12 mapped + gtd/moai-web/probe branch families + the audit's own 16/16 sample, incl. WT-codex-support-fixes and release/v3.1.3) found every sampled removed branch an origin-ancestor — **zero observed loss** — but "no observed loss on sampled branches" is NOT the SPEC's no-loss proof for the unledgered set. **OPERATOR DISCLOSURE REQUIRED**: the unledgered ~430-tree disposal is disclosed to the lead/operator with this record; the trees are not revivable and the classification evidence for them no longer exists anywhere.
- **F5 corrections to the text above**: (1) kept_count below said 79 — wrong; measured **86 registered (incl. primary checkout) / 97 dirs** (audit measured 96 dirs — count-method variance, not load-bearing). (2) WT-receipt-live-probe was listed under "removed-tree families" — **mislist**: its tree (t850) survives; a merged BRANCH proves nothing about its tree's removal, and branch-to-tree over-matching is exactly how this error happened.
- **Delta decomposition supplement (2026-09-23, lead disclosure)**: the lead's own 09-22 first-pass cleanup removed **9 trees** with full 3-condition verification (merged+pushed+unoccupied) and its own ledger record. The 544→86 delta (458) therefore decomposes as: **26 executor-ledgered + 9 lead-ledgered (verified, ledgered) + ~423 attribution-unknown** (no durable ledger rows anywhere — the operator-disclosure segment is this ~423, not the full ~430+). The lead's 9-tree ledger lives in the lead's own record (local-only channel; citable by the lead on request — delta-audit R2).
- **AC failure states (honest, not retrofitted)**: AC-001 FAIL (16 removed trees' branch refs survived un-deleted — harmless here since all are origin-ancestors, but the procedure cell fails); AC-002/009 FAIL (the 7 dirty trees were removed without per-tree disposition-record path/direction citations — the D12 gate covered only the `--force` flag); AC-003/004/011/012/013 FAIL (verdict file, classification table, prune diff, window-boundary record: not persisted). AC-010 PASS (export evidence measured: rescue 45 dirs / 3.1 GB / 161,976 files, max 149 MB/tree ≤ 200 MB cap). The remaining ACs carry the same evidence limits.
- spec.md `status` returns to **in-progress** in this repair commit — the `completed` declaration on 79de5ad56 is not accepted while a blocking sync-audit FAIL stands; re-close follows the delta-scoped re-audit.

### Gaps

- The executor's M1 per-tree classification table and its §E verification matrix were NOT persisted (lost with the transcript) — see the CORRECTION above for what that costs (the ~430-tree unledgered disposal is the materialized form of this gap).
- The export-log dirty-removal rows are absent (executor died between removal and log update). The export rows + post-hoc absence stand in for the 7; the gap is a record incompleteness, not an unverified removal.
- ~11 unregistered orphan dirs under `.claude/worktrees/` remain (kept + reported; REQ-WGC-008 conservative direction).
- spec.md status flips in this takeover/repair series carry the lane trailer — the ownership-lint warning, if emitted, is the documented takeover shape (delegation target absent; see §F).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-23
run_status: complete (lane-takeover-completed) — SYNC AUDIT FAIL 50/100 on 79de5ad56; record repaired, status reverted to in-progress, delta-scoped re-audit pending
run_commit_sha: "df8af0621"   # backfilled (D3 window) — the takeover close commit
run_executed_by: manager-develop (M1-M3 actions) + lane-orchestrator-agent-30 (tail: post-hoc verification, census, record, corrections)
ledger_covered_removals: 26 (19 clean + 7 dirty, each dirty with a pre-removal export row)
unledgered_disposal_delta: "~430+ trees (544 registered at plan baseline -> 86 now; ledger covers 26) — NO classification evidence exists; OPERATOR DISCLOSURE REQUIRED (see §E.2 CORRECTION)"
kept_count: 86 registered now incl. primary + develop + current-active batch + keep-direction records (t1050/t1064/t810 verified present) + conservative keeps; 97 dirs, ~11 unregistered orphans kept
disposal_reduction: 544 registered → 86 registered; 138.3 GB → 26 GB (dated anchors: plan baseline 2026-09-22 vs close census 2026-09-23)
ac_status: "FAIL cells recorded honestly in §E.2 CORRECTION (AC-001, 002, 003, 004, 009, 011, 012, 013); AC-010 PASS on measured export evidence; re-judgment owned by the delta-scoped re-audit"
sync_audit_fail: ".moai/reports/t1084/sync-audit.md — FAIL 50/100, blocking F1-F5; repair = this record correction + operator disclosure; re-audit = defect delta only"
resume_or_followup: operator disclosure of the unledgered disposal (lead relays); delta-scoped re-audit; then re-close
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_commit_sha: "pending-backfill-sync"   # D3 backfill window — real SHA backfilled in a follow-up commit
sync_close_at: 2026-09-23
sync_status: complete
tier: M
takeover_note: >-
  Run phase completed via LANE TAKEOVER — the delegated manager-develop executed the disposal
  actions (M1-M3) but died before reporting; the lane recorded everything in progress.md §E.2.
  Evidence surface: §E.2 (this file) + .moai/reports/t1084/export-log.md + rescue/.
changelog_emitted: false
changelog_reason: >-
  maintainer-local operational cleanup (legacy worktree disposal) with zero user-facing product
  change; no CHANGELOG entry emitted per sync-phase emission discipline.
ac_verification_note: >-
  The executor's AC matrix was not persisted (lost with its transcript); AC-WGC-001..013
  re-judgment from durable evidence (export-log.md, rescue/, §E.2 post-hoc verification) is the
  sync audit's job.
```

Sync-phase scope: `spec.md` frontmatter transition (`in-progress → implemented → completed` merged
into this single sync commit; `updated: 2026-09-23`) + this §E.4 block ONLY. Zero code changes,
zero template-mirror edits across the whole SPEC.

## §F Phase 4 Mode Selection

Logged 2026-09-22 by the lane orchestrator (agent-30, re-dispatch; original lane session ended) before the first run-phase Agent() spawn, per the mode-logging contract.

**Input parameters**: tier=M; scope=disposal execution across cooled L1 worktrees (zero code, zero template mirrors); domains=1 (git-worktree lifecycle procedure); file language mix=markdown evidence only; concurrency benefit=LOW (serialized destructive-adjacent operations with export-before-removal ordering); agent-team prereqs=not requested.

**Mode evaluation**: direct — no (multi-tree procedural execution with evidence ledger); serial — SELECTED (one manager-develop delegation carries M1-M4 in order; destructive-adjacent steps serialize by nature); fanout — no (no independent read/write units); sweep — no (not mechanical-uniform bulk; judgment-laden per-tree disposition).

**Decision: serial**

**Justification**: per-tree disposal decisions with export-before-removal ordering and keep-direction overrides are sequential by nature; a single executor keeps the 3-tier predicate and the evidence ledger coherent. Kickoff Approval: PASSED (operator "전부 승인" 2026-09-22, relayed by lead). Plan-audit iter-2 PASS-WITH-DEBT 0.9375; the D10-D12 post-verdict repair (ce415cb01) is the audit-prescribed fix — recorded as the run-gate skip deviation, not a silent hash claim.
