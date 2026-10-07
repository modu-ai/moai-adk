# SPEC-MERGE-WINDOW-QUEUE-001 — Research

> Card t1479 · plan phase · measured on branch `WT-merge-window-queue`, HEAD `d7112d005`
> (tree `632f65b47aa5`), after `git merge develop` reported "Already up to date".

## §R1 Baseline cells (RED-now, measured in this plan run)

| Cell | Command | Observed | Reads as |
|---|---|---|---|
| E1 | `grep -c '"wait"' internal/cli/integration.go` | `0` | acquire has no `--wait` flag |
| E2 | `grep -n -i "queue\|ticket\|waiter" internal/kanban/integration_lock.go internal/kanban/integration_lock_mutation.go \| wc -l` | `0` | the record carries no queue |
| E3 | `grep -n 'Use:   "' internal/cli/integration.go` | `integration [command]`, `status`, `acquire`, `release` (lines 254/274/329/457) | no `push`, no policy verb |
| E4 | `grep -rn "LeadPushThreshold" internal \| grep -v _test.go \| grep -v internal/config/ \| wc -l` | `0` | threshold has no consumer — since v0.3.0 this cell belongs to SPEC-CANDIDATE-CI-001 (push verb moved), kept here as history |
| E5 | `grep -n 'strings.Contains(strings.ToLower(string(raw)), full\[:12\])' internal/homestate/card_evidence_readers.go` | line `255` | gate = SHA substring |
| E6 | `grep -n 'factoryWriteMergeRecord(root' internal/cli/factory_card.go` | line `1418` (def. `1616`) | complete writes its own stand-in |
| E7 | `grep -n "nominat\|지명" AGENTS.local.md` | line `221`: "리더의 창 지명만이 근거다." | nomination doctrine live |
| E8 | `grep -n "announcement to the lead" …/kanban-dispatch-mechanics.md` (template + local) | line `120` in both | distributed text keeps announcement layer |
| E9 | `grep -n -i -w "lease\|expires\|expiry" internal/kanban/integration_lock*.go \| wc -l` | `0` | no window lease today (a bare `lease` grep returns 29 — all inside "release"; word-bounded is the valid probe) |
| E10 | `grep -n "exec.Command\|\"test\"" internal/factorylane/merge.go` | only `86: exec.Command("git", args...)` | pre-merge triple runs no tests |
| E11 | no-`--wait` acquire against a live foreign holder (fixture) on a binary built from this branch (Go code = merge base `d7112d005`) | exit `1`, empty stdout, refusal on stderr, record unchanged — both human and `--json` | the baseline AC-MWQ-010 compares against; committed in `3bc274dac`, `.moai/reports/t1479/baseline-acquire-nowait/` |

The plan audit re-ran E1, E3, E5-E8, E10 on `1e1d0cc84` and reproduced them; it did not run the
piped cells E2, E4, E9 (single-invocation form) — they rest on this plan run's measurement only.

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
  `internal/hook/integration_lock_guard.go:84` (PreToolUse guard; REQ-MWQ-023).
- `internal/homestate/card_evidence_readers.go:205-258` — `verifyMerge`: two-parent check, merge
  tree == second-parent tree, reachability from the integration branch, then the substring gate.
  The tree-identity property means the candidate tree is exactly the card branch tip tree after
  absorbing the integration tip — the key REQ-MWQ-014 uses.
- `internal/factorylane/merge.go` — merge triple (sync-audit, conflict-free via
  `git merge-tree --write-tree`, tree identity), `VerifyRunBeforeAcquire`, `WindowCoversMerge`,
  merge-check store. REQ-MWQ-021 adds a fourth named condition here.
- `internal/cli/ci_verdict.go` and `internal/config/types.go:133-140` (`LeadPushThreshold`) — push
  inputs; since v0.3.0 they belong to SPEC-CANDIDATE-CI-001 (see §R6 for the `--limit 1` finding).
- `internal/config/envkeys.go:361` — `EnvFactoryRole` (`MOAI_FACTORY_ROLE`); lane value is the
  role claim the policy verb refuses (REQ-MWQ-012). A session with no value makes no claim.
- Process start time for ticket liveness (REQ-MWQ-003): `homestate.CurrentProcessFingerprint()`
  already pairs a pid with its start (used by the factory peer registry); reuse is the run phase's
  choice.

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
  asymmetry REQ-MWQ-006 carries through promotion by stamping the ticket's owner pid on the holder
  (tickets additionally require their waiter process, REQ-MWQ-003).
- SPEC-INTEGRATION-LOCK-TARGET-SOURCE-001 (completed) — additive-optional field precedent.
- SPEC-FACTORY-LANE-AUTONOMY-001 (completed) — merge triple and WAITING verdict.
- SPEC-FACTORY-SELF-DISPATCH-001 (completed) — REQ-SD-023 window phase in complete; REQ-SD-025
  Codex stop (out of scope here).
- SPEC-LEAD-AUTOPUSH-001 (completed) — chose a docs-only surface for the threshold; the mechanism
  it deferred is now SPEC-CANDIDATE-CI-001 REQ-CCI-012/013 (not this SPEC, since v0.3.0).

## §R5 Card t1478 coupling

t1478 (SPEC-CANDIDATE-CI-001, `moai integration candidate`, `ci/<card>` branches) is not landed
(`git log --oneline develop | grep -i t1478` → no output on this tree). Its draft (v0.4.0, read from
its card worktree `agent-a18f82893f35ed6c4`, uncommitted to develop) names the key
`workflow.candidate_ci.enabled` (REQ-CCI-023, template default false), a shared landing check
called by every develop-merging path (REQ-CCI-011), and the push verb (REQ-CCI-012/013).

This SPEC therefore (v0.5.0 numbering):

- selects the record form by `workflow.candidate_ci.enabled`, absent = false (REQ-MWQ-014);
- has `moai integration merge --card <id>` resolve the card's branch by the REQ-CCI-004 contract
  (`git worktree list --porcelain` at the card's worktree path) and call the REQ-CCI-011 shared
  landing check, a no-op while the key is false (REQ-MWQ-017);
- acknowledges SPEC-CANDIDATE-CI-001 D7, which assigns to this card the rule "rebuild a stale
  candidate only when the card reaches the head of the merge-window queue, never eagerly": the
  merge verb detects staleness only while holding the window, and the re-measure follows that exit
  (REQ-MWQ-018);
- carries no push requirement (§E).

Two coordination notes from plan-audit iteration 3 (O3, O4), recorded rather than fixed here:

- **O3 — shared doctrine wording.** SPEC-CANDIDATE-CI-001 REQ-CCI-024 also writes a "lanes merge into
  develop only through the integration verbs" clause into the same two local files. AC-MWQ-013
  expects exactly one occurrence of this SPEC's sentence, so whichever card lands second reconciles
  the wording into one sentence rather than adding a second.
- **O4 — branch resolution reach.** Following REQ-CCI-004, the merge verb resolves a card's branch
  from `.claude/worktrees/<card>` or `.moai/worktrees/<card>` only. It has no `--branch` fallback
  (t1478's candidate verb has one) and does not read the factory record's `WorktreePath`, so a card
  tree with an agent-generated name (this card's own `agent-a1d44…`) does not resolve. Extending
  the resolver belongs to the REQ-CCI-004 contract owner, so both verbs keep one rule.

Whichever card lands second re-reads the other's committed key path and landing-check contract
and adapts its own text; the draft read above is a moving source, not a citation of record.

## §R6 Hand-off to SPEC-CANDIDATE-CI-001 (card t1478) — push-verb findings

The push verb left this SPEC in v0.3.0 (leader decision Q8). The plan audit of this SPEC
(`.moai/reports/t1479/plan-audit-iter1.md` D8) found three blocking defects in the push design
as it stood here. They are handed off, not dropped; t1478's v0.4.0 text already addresses (c) and
partly (a)/(b) — its owner confirms each:

- **(a) Repair push on a red remote tip.** A rule that refuses every push while the remote tip is
  red also refuses the push that would repair it. A repair path is needed — t1478 REQ-CCI-013 has a
  repair-card exception; confirm it is recorded and never force.
- **(b) Wrong CI run read.** `gh run list --commit <head> --limit 1` returns the newest run, which
  may be in progress while an earlier completed run on the same tip failed — read as "no completed
  run", the hold is bypassed. Read the most recent COMPLETED run (or all runs) for the tip SHA, and
  treat the mapped failure class (failure, timed_out, startup_failure — `mapGHConclusion`,
  `internal/cli/ci_verdict.go:167`) as red.
- **(c) Push the verified SHA, not the branch name.** Checking the window and then pushing the
  moving name `develop` lets a lane merge in between; push `<sha>:refs/heads/develop` for the pinned
  SHA that was verified (t1478 REQ-CCI-012 does this).
- Optional (d): "leader-only" was enforced only as "not lane role" (`MOAI_FACTORY_ROLE`); either
  narrow the wording or define a leader marker.
- **(e) Unreadable CI → refuse, never green** (added v0.5.0 after plan-audit iteration 2 N5).
  0.1.0-0.2.0 of this SPEC carried a fail-closed clause (then REQ-MWQ-042): when the CI state
  cannot be read — the CI client absent, unauthenticated, or the query failing — the push verb
  refuses, naming the fault, rather than treat the state as green. The move to t1478 dropped it:
  `grep -n -i "cannot be read|unauthenticated|unreadable"` over t1478's spec.md returned no output
  in the iteration-2 audit. The t1478 owner adopts or explicitly rejects this clause; either answer
  is recorded there. Note the distinction it rests on: "no completed run yet" is not red
  (decision-index Q7's policy), but "could not read" is unmeasured, which that policy does not
  cover.

Status of this hand-off: written for the leader to forward to card t1478; this SPEC makes no claim
that t1478 has accepted any item.

## §R7 Queue fairness — removed in v0.5.0, and what that leaves

Iterations 0.2.0-0.4.0 built fairness machinery around the requeue case: a reserved ticket kept for
a lane whose develop base moved, `--slice` waits with a between-slices state, a one-time front
promotion, a requeue counter, and a three-requeue bound. Plan-audit iteration 2
(`.moai/reports/t1479/plan-audit-iter2.md`) found that machinery self-contradictory:

- N1: the holder lost its owning-session pid;
- N2: the liveness drop removed reserved tickets before their bound applied;
- N3: between-slices tickets could neither survive the heartbeat window nor be promoted;
- N4: the requeue counter was unreachable by any defined transition.

The score regressed from 0.75 to 0.69.

The leader removed it (decision-index Q17). The reason is that the premise changed: once the
re-measure runs outside the window, the window is held for the seconds an identity check and a
`--no-ff` merge take. A lane sent to the tail waits behind a few seconds per earlier ticket, not
behind other lanes' re-measures, so fairness beyond FIFO buys little and its state cost was the
source of every iteration-2 defect.

What remains, stated plainly: a lane whose re-measure keeps being invalidated by other lanes' merges
re-enters at the tail each time, and nothing bounds how many times that happens. If the run phase
observes repeated re-acquires for one card, the measured count goes to the leader, who may revisit
fairness with that number in hand.
