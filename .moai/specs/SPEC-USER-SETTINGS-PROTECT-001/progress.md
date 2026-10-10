# SPEC-USER-SETTINGS-PROTECT-001 — Progress Record

Card t1630 (backlog 3.2-0-1, priority P0, change class C). Lane lane-18, run tmnboq, branch WT-3-2-0, worktree `.moai/worktrees/t1630`. Tier M (lane decision, §G; spec.md §0). Artifact set: spec.md, plan.md, acceptance.md, decision-index.md (required by the decision gate), and this progress record.

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: pending-reaudit
- plan_complete_at: 2026-10-10T07:02:43Z (first authoring)
- plan_tree: revision 2 is written on top of HEAD 5dc6c4530 (short SHA measured in the repair round). The RED observations stay pinned at 3975fe3cc. The lane commits revision 2, and the re-audit runs on that commit.
- revision: revision 2 is the repair round after the iteration-2 FAIL (aggregate 0.69; audited_sha 5dc6c45307d21e4b0d34c876631e2fa553fde57b). It applies the operator-delegated split under ruling d-20261010T091713Z-10fe (decision board, card t1630). Policy decisions Q1, Q2, and Q5 stay in scope with the text and evidence defects B1, B4, B6 (wording), B8, B9, B11, and B12. Moved to t1666: the test-isolation items, the clean --home run and db items, decision Q3, decision Q7, the ~/.moai/run growth-0 criterion, and the post-t1619 baseline. Audit defects B2 (the AC-009 run path), B3 (AC-007), B5 (the AC-006 cell 1 order), B7 (the REQ-011 and REQ-012 deletion clauses), and B10 (the AC-006 run-time verdict) move with their criteria.
- audit-ready signal: withdrawn until the re-audit of revision 2 passes. The re-audit is one run on this revision. The ceiling count restarts at this revision.
- counts (revision 2): in-scope normative requirements 4 (REQ-001, REQ-002, REQ-005, and REQ-013; REQ-003 and REQ-004 sit in the spec.md §2 note); in-scope acceptance criteria 6 (AC-001 to AC-005 and AC-013); both within the Tier M ceiling of 16. Moved and retained: REQ-006 to REQ-012 and AC-006 to AC-012 (7 of each).
- evidence_manifest_sha256: 628c676714fe4b3c040f7a6a14d84618e79d9e3b0bcd3e1f57e7ffca74fd1ba8 (`shasum -a 256 .moai/specs/SPEC-USER-SETTINGS-PROTECT-001/evidence-manifest.json`, observed in revision 2). The last commit that changed the manifest is 5dc6c4530, and revision 2 does not change it. The binding is by manifest hash; the originals stay local under `.moai/reports/t1630/` (plan.md Appendix A). The plan-artifact hash (the ComputeHash subject, internal/runtime/audit_cache.go:92-100) covers acceptance.md, decision-index.md, design.md, plan.md, research.md, spec.md, and tasks.md, and excludes progress.md and evidence-manifest.json.
- independent plan audit: not run by the author. The lane commits, then runs one re-audit (audit_multi with codex required, per the dispatch).
- decisions: Q8 DECIDED (anchor SPEC-INIT-WIZARD-REPAIR-001, HISTORY row 0.1.1). Q1, Q2, and Q5 DECIDED by the pinned board record `board:d-20261010T073547Z-07ae#9c93e47809f9` (decided_by 영실이 판단, operator-delegated). Q4 and Q6 EVIDENCE-NEEDED, closed at M2 without blocking run. Q3 and Q7 are Out of Scope - moved to t1666; their rows keep their text.
- Q7 and the Kickoff: Q7 (FOUNDER, implementation-level) has no subject once the sandbox items move, so it no longer gates this SPEC's Kickoff. The leader confirms this reading at re-audit.
- no default is applied at plan close.
- ordering gate REQ-013: open at plan close (AC-013 red, plan.md E-14); the run-phase gate is P-2.
- EC-1 to EC-4: observed in revision 2 by the EC observation command in acceptance.md (exit code 0 at HEAD 5dc6c4530).
- AC-003 part A RED: re-observed at HEAD 5dc6c4530 in revision 2 (verbatim in acceptance.md; exit code 1).
- open: G-18 (the digest covers the header copy only; disclosed).
- tier: the in-scope set is four files (plan.md §A.4 Basis), below the Tier M file band of 5 to 15. The tier is not re-decided in revision 2 (see §G, Tier decision).
- scope boundary: scope item S-7 (t1594) is blocked on card text (plan.md G-2). No requirement is authored for it in this revision.
- process note: HEAD moved during this repair round, from 5dc6c4530 to 169a780db (commit "docs(SPEC-USER-SETTINGS-PROTECT-001): revision 2 after split (card t1630)", author Goos Kim, 2026-10-10 18:52:13 +0900, which holds the revision-2 state as of the AC-013 header edit). The repair-round edits are uncommitted in the working tree; this pass did not commit or stage. Reported in the repair-round reply.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §G Lane design decisions (plan phase)

### Tier decision (spec.md §0)

Decision: Tier M was kept at the pre-split scope of 16 files, option (b), as ratified by ruling d-20261010T081810Z-49e5. Revision 2 does not re-decide the tier. The in-scope set is four files (plan.md §A.4 Basis: items 1 to 3 and the Q2 template file), below the Tier M file band of 5 to 15, which is a guide and not a gate. The leader's decision, board record d-20261010T094229Z-74f2, keeps Tier M for the split scope.

In-scope files after revision 2: 4 (plan.md section A.4 Basis items 1 to 3, and the template file). Tier M is kept as recorded; the 5 to 15 file band is a guide, not a gate.

decision record: decided_by=lane-18 evidence_refs=.moai/specs/SPEC-USER-SETTINGS-PROTECT-001/spec.md §0; .moai/specs/SPEC-USER-SETTINGS-PROTECT-001/plan.md §A.4 Basis; board:d-20261010T081810Z-49e5 ladder_path=ladder ⑤ (lane judgment; the leader's ruling confirms option (b))

### G-17 — template defaultMode value

Decision: the template carries `permissions.defaultMode` = `"default"`. `acceptEdits` is not chosen: it would auto-approve edits, a behaviour change beyond ruling d-20261010T073547Z-07ae.

decision record: decided_by=lane-18 evidence_refs=.claude/skills/moai-foundation-cc/reference/claude-code-settings-official.md:89; internal/cli/launcher.go:737-741; board:d-20261010T081810Z-49e5 ladder_path=ladder ⑤ (lane judgment; ruling item 2 places this choice in the lane's ladder)

### G-20 — managed-block detection

Decision: diff-based. The writer keeps the last generated managed allow list in `.moai/state/tool-policy/managed-allow.json` as `last_generated`. user_added = existing allow entries minus `last_generated`; result = regenerated ∪ user_added. With no record, every existing entry is kept, so nothing is removed. Only the user removes user-added entries (ruling d-20261010T073547Z-07ae).

decision record: decided_by=lane-18 evidence_refs=board:d-20261010T073547Z-07ae (marker- or diff-based); internal/config/toolpolicy/tier_render.go:77; .moai/reports/t1630/evidence/probe-ac005_test.go.txt; .moai/reports/t1630/evidence/probe-ac005b_test.go.txt ladder_path=ladder ⑤ (lane judgment within the record's "marker- or diff-based" allowance)

### G-23 — run and db in the --force allowlist

Out of Scope - moved to t1666 (decision Q3). The text is retained unchanged.

Decision: the `run` and `db` categories join the `--force` allowlist of `moai clean --home`, so `--force` can delete run and db candidates. Run items are candidates only under the REQ-011 rule, and `--force` deletes only candidates. Without `--force`, run and db items are listed only. The ruling 07ae reads db as list-only, with deletion only by explicit `--yes`, which ruling 49e5 maps to `--force`. No live-record exclusion is added for db, because the ruling does not state one. plan.md's G-22 (the tier_render.go:106 statement) is a separate gap.

decision record: decided_by=lane-18 evidence_refs=board:d-20261010T073547Z-07ae; .moai/specs/SPEC-USER-SETTINGS-PROTECT-001/spec.md REQ-011 and REQ-012 ladder_path=ladder ⑤ (lane judgment; the ruling's explicit-deletion wording)
