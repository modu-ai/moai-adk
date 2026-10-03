# SPEC-MERGE-WINDOW-QUEUE-001 — Research

> Card t1479 · plan phase · measured on branch `WT-merge-window-queue`, HEAD `d7112d005`
> (tree `632f65b47aa5`), after `git merge develop` reported "Already up to date".

## §R1 Baseline cells (RED-now, measured in this plan run)

| Cell | Command | Observed | Reads as |
|---|---|---|---|
| E1 | `grep -c '"wait"' internal/cli/integration.go` | `0` | acquire has no `--wait` flag |
| E2 | `grep -n -i "queue\|ticket\|waiter" internal/kanban/integration_lock.go internal/kanban/integration_lock_mutation.go \| wc -l` | `0` | the record carries no queue |
| E3 | `grep -n 'Use:   "' internal/cli/integration.go` | `integration [command]`, `status`, `acquire`, `release` (lines 254/274/329/457) | no `push`, no policy verb |
| E4 | `grep -rn "LeadPushThreshold" internal \| grep -v _test.go \| grep -v internal/config/ \| wc -l` | `0` | threshold has no consumer |
| E5 | `grep -n 'strings.Contains(strings.ToLower(string(raw)), full\[:12\])' internal/homestate/card_evidence_readers.go` | line `255` | gate = SHA substring |
| E6 | `grep -n 'factoryWriteMergeRecord(root' internal/cli/factory_card.go` | line `1418` (def. `1616`) | complete writes its own stand-in |
| E7 | `grep -n "nominat\|지명" AGENTS.local.md` | line `221`: "리더의 창 지명만이 근거다." | nomination doctrine live |
| E8 | `grep -n "announcement to the lead" …/kanban-dispatch-mechanics.md` (template + local) | line `120` in both | distributed text keeps announcement layer |
| E9 | `grep -n -i -w "lease\|expires\|expiry" internal/kanban/integration_lock*.go \| wc -l` | `0` | no window lease today (a bare `lease` grep returns 29 — all inside "release"; word-bounded is the valid probe) |
| E10 | `grep -n "exec.Command\|\"test\"" internal/factorylane/merge.go` | only `86: exec.Command("git", args...)` | pre-merge triple runs no tests |

## §R2 Code map

- `internal/kanban/integration_lock.go` — `IntegrationLock` record (holder fields, PIDSource,
  BranchSource, settings-drift bypass), `AcquireIntegrationLock` (read→decide→write inside
  `withIntegrationLockMutation`), `ReleaseIntegrationLock`, `Stale()` (pid liveness; pid 0 = live),
  atomic `writeIntegrationLock`. Record path: primary checkout `.moai/state/integration-lock.json`.
- `internal/kanban/integration_lock_mutation.go` — the cross-process mutation section the queue
  must reuse (REQ-MWQ-002 ordering is decided inside it).
- Lock consumers (non-test): `internal/cli/integration.go` (status/acquire/release verbs),
  `internal/cli/factory_merge.go` (merge-readiness: WAITING path, AC-FLA-010),
  `internal/cli/factory_card.go` (complete: window phase 1325-1350, own acquire ~1380, merge
  1408-1425, stand-in record 1612-1628), `internal/cli/session_worktree_automerge.go:93-100`
  (session-end automerge; calls acquire with force=false — must stay unchanged),
  `internal/hook/integration_lock_guard.go:84` (PreToolUse guard; REQ-MWQ-051).
- `internal/homestate/card_evidence_readers.go:205-258` — `verifyMerge`: two-parent check, merge
  tree == second-parent tree, reachability from the integration branch, then the substring gate.
  The tree-identity property means the candidate tree is exactly the card branch tip tree after
  absorbing the integration tip — the key REQ-MWQ-021 uses.
- `internal/factorylane/merge.go` — merge triple (sync-audit, conflict-free via
  `git merge-tree --write-tree`, tree identity), `VerifyRunBeforeAcquire`, `WindowCoversMerge`,
  merge-check store. REQ-MWQ-033 adds a fourth named condition here.
- `internal/cli/ci_verdict.go` — existing `gh run list --commit <head> --limit 1 --json
  conclusion,databaseId` producer with an injectable `ghRunner` and `mapGHConclusion`; the push
  verb's CI read (REQ-MWQ-041/042) reuses this rather than adding a second gh path.
- `internal/config/types.go:133-140`, `defaults.go:1050` — `LeadPushThreshold` (manual mode; 0 =
  disabled). This repository's value is `20` (`.moai/config/sections/git-strategy.yaml:26`); the
  template ships `0`.
- `internal/config/envkeys.go:361` — `EnvFactoryRole` (`MOAI_FACTORY_ROLE`); lane value is the
  role claim the policy and push verbs refuse (REQ-MWQ-012/043). A session with no value makes no
  claim.

## §R3 Doctrine surfaces

- `AGENTS.local.md` §4.1 (tracked maintainer copy): rule 4 (serial window), line 221 (nomination),
  the lane window procedure line (acquire → absorb → re-measure → merge → release), rule 6 (batch
  push + green-conditional).
- `.claude/rules/local/gitflow-lane-protocol.md` §3 (serialization: "리더 공지가 여전히 첫 번째
  층"), §4 (threshold + green-conditional by hand), §6 (self-dispatch window exception), §7.
- `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch-mechanics.md:120` and
  its local mirror — "The announcement to the lead rides alongside it." The stub
  `kanban-dispatch.md:181` already says integration is self-served; no "nomination" token exists in
  any template file (grep, §R1 E8 context).
- `.moai/docs/gitflow-integration-chain.md` — window execution bash (relocated from §4.1); updated
  to the new order.

## §R4 Related SPECs and precedents

- SPEC-INTEGRATION-LOCK-ATOMIC-001 (completed) — the mutation section; queue mutations ride it.
- SPEC-INTEGRATION-LOCK-LIVENESS-001 (completed) — owner-pid anchor and the live-when-unknown
  asymmetry REQ-MWQ-004 preserves.
- SPEC-INTEGRATION-LOCK-TARGET-SOURCE-001 (completed) — additive-optional field precedent.
- SPEC-FACTORY-LANE-AUTONOMY-001 (completed) — merge triple and WAITING verdict.
- SPEC-FACTORY-SELF-DISPATCH-001 (completed) — REQ-SD-023 window phase in complete; REQ-SD-025
  Codex stop (out of scope here).
- SPEC-LEAD-AUTOPUSH-001 (completed) — chose a docs-only surface for the threshold; this SPEC
  supplies the mechanism that SPEC deferred.
- SPEC-CI-VERDICT-PRODUCER-001 — the gh conclusion mapping reused by the push verb.

## §R5 Card t1478 coupling

t1478 (`moai integration candidate`, `ci/<card>` branches) is queued and not landed
(`git log --oneline develop | grep -i t1478` → no output on this tree). The SPEC therefore defines
the candidate tree independently (absorbed card-branch tip) and admits a t1478 candidate CI run id
as an alternative evidence form (REQ-MWQ-020/021). Whichever card lands second adapts to the other;
see decision-index.md Q5.
