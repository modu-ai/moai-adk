# SPEC-USER-SETTINGS-PROTECT-001 — Progress Record

Card t1630 (backlog 3.2-0-1, priority P0, change class C). Lane lane-18, run tmnboq, branch WT-3-2-0, worktree `.moai/worktrees/t1630`. Tier M (lane decision, §G; spec.md §0). Artifact set: spec.md, plan.md, acceptance.md, decision-index.md (required by the decision gate), and this progress record.

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: pending-reaudit
- plan_complete_at: 2026-10-10T07:02:43Z (first authoring)
- plan_tree: 3975fe3cc on WT-3-2-0 (the repair-round edits are uncommitted)
- revision: the artifacts were revised after the 3975fe3cc audit (FAIL 0.69). The audit-ready signal is withdrawn until one re-audit runs after the lane commits, as the leader approved.
- counts: 13 requirements (ceiling 16), 13 acceptance criteria (ceiling 16)
- independent plan audit: not run by the author. The lane commits, then runs one re-audit (audit_multi with codex required, per the dispatch).
- decisions: Q8 DECIDED (anchor SPEC-INIT-WIZARD-REPAIR-001, HISTORY row 0.1.1). Q1, Q2, Q3, and Q5 DECIDED by the pinned board record `board:d-20261010T073547Z-07ae#9c93e47809f9` (decided_by 영실이 판단, operator-delegated). Q4 and Q6 EVIDENCE-NEEDED, closed at M2 without blocking run. Q7 FOUNDER (implementation-level), open with no default applied: the autonomous Kickoff waits for an operator verdict.
- no default is applied at plan close. The earlier statement that Q7's default was applied at plan close is withdrawn.
- lane design decisions (§G): the tier (option b), the template value `"default"` (G-17), diff-based managed-block detection (G-20), and the run and db categories in the `--force` allowlist (G-23).
- ruling d-20261010T081810Z-49e5 (board, kind=ruling) resolves G-19: the record's `--yes` is the existing `--force` flag, with no new flag.
- open: G-18 (the digest covers the header copy only; disclosed) and G-21 (AC-011 clause iii, a regression-guard whose RED the run phase authors before its GREEN); the run-phase deletion criterion (acceptance.md Definition of Done) is authored with its RED first.
- ordering gate REQ-013: open at plan close (AC-013 red, plan.md E-14).
- open gaps: plan.md §A.4, G-1 to G-23.
- scope boundary: scope item S-7 (t1594) is blocked on card text (plan.md G-2). No requirement is authored for it in this revision.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §G Lane design decisions (plan phase)

### Tier decision (spec.md §0)

Decision: Tier M kept at 16 files, option (b). The file band of 5 to 15 files is a guide, and the 16th file is a one-line template edit. The decision overrides the tier-up trigger. Ruling d-20261010T081810Z-49e5 confirms option (b).

decision record: decided_by=lane-18 evidence_refs=.moai/specs/SPEC-USER-SETTINGS-PROTECT-001/spec.md §0; .moai/specs/SPEC-USER-SETTINGS-PROTECT-001/plan.md §A.4 Basis; board:d-20261010T081810Z-49e5 ladder_path=ladder ⑤ (lane judgment; the leader's ruling confirms option (b))

### G-17 — template defaultMode value

Decision: the template carries `permissions.defaultMode` = `"default"`. `acceptEdits` is not chosen: it would auto-approve edits, a behaviour change beyond ruling d-20261010T073547Z-07ae.

decision record: decided_by=lane-18 evidence_refs=.claude/skills/moai-foundation-cc/reference/claude-code-settings-official.md:89; internal/cli/launcher.go:737-741; board:d-20261010T081810Z-49e5 ladder_path=ladder ⑤ (lane judgment; ruling item 2 places this choice in the lane's ladder)

### G-20 — managed-block detection

Decision: diff-based. The writer keeps the last generated managed allow list in `.moai/state/tool-policy/managed-allow.json` as `last_generated`. user_added = existing allow entries minus `last_generated`; result = regenerated ∪ user_added. With no record, every existing entry is kept, so nothing is removed. Only the user removes user-added entries (ruling d-20261010T073547Z-07ae).

decision record: decided_by=lane-18 evidence_refs=board:d-20261010T073547Z-07ae (marker- or diff-based); internal/config/toolpolicy/tier_render.go:77; .moai/reports/t1630/evidence/probe-ac005_test.go.txt; .moai/reports/t1630/evidence/probe-ac005b_test.go.txt ladder_path=ladder ⑤ (lane judgment within the record's "marker- or diff-based" allowance)

### G-23 — run and db in the --force allowlist

Decision: the `run` and `db` categories join the `--force` allowlist of `moai clean --home`, so `--force` can delete run and db candidates. Run items are candidates only under the REQ-011 rule, and `--force` deletes only candidates. Without `--force`, run and db items are listed only. The ruling 07ae reads db as list-only, with deletion only by explicit `--yes`, which ruling 49e5 maps to `--force`. No live-record exclusion is added for db, because the ruling does not state one. plan.md's G-22 (the tier_render.go:106 statement) is a separate gap.

decision record: decided_by=lane-18 evidence_refs=board:d-20261010T073547Z-07ae; .moai/specs/SPEC-USER-SETTINGS-PROTECT-001/spec.md REQ-011 and REQ-012 ladder_path=ladder ⑤ (lane judgment; the ruling's explicit-deletion wording)
