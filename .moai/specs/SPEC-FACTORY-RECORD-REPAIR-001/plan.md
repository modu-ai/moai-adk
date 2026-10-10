---
id: SPEC-FACTORY-RECORD-REPAIR-001
version: "0.4.0"
updated: 2026-10-10
---

# Plan: SPEC-FACTORY-RECORD-REPAIR-001

## §A Context

- Card t1621, Class C, Tier M. Work root: worktree `.moai/worktrees/t1621`, branch `WT-factory-record-fix`, base HEAD `2aab5f797`. This plan is the plan phase only; the implementing lane writes the code and commits.
- Evidence is read-only and untracked. The lens-4 and lens-2 reports live in the primary checkout's `.moai/reports/audit-20261010/` and are not present in this worktree. Each code location below was re-read in this worktree. § Verified premises states, for each premise, whether it was read, run, or only reported by a lens.
- Problem: the factory record can strand a holder's card (expired lease, no renewal, no recovery), it has no pushed-to-done edge for a card with a remote, and it has no detection of writes that bypass the transition API. The narrowed batch covers all three items (card body items 1 to 3).
- Revision 0.2.0 narrows the plan to the three surviving items of the card re-scope, recorded as leader ruling `d-20261010T053435Z-006f`. It applies the operator verdicts of `d-20261010T053034Z-1a19` where their subjects survive. Q2, Q4, Q5, Q6, Q7, and Q13 are marked not applicable, because their subjects are retired (decision-index Routing). Revision 0.3.0 writes item (3) under leader ruling `d-20261010T060606Z-c5c4` (option a). Revision 0.4.0 applies operator-delegated decision `d-20261010T064450Z-68b2` (option b): it renumbers the requirements (B1), transfers item (3), the record-call placement, and the non-holder ordering to 3.2-2-2 (B2, B4, B5), and bounds the watchdog renewal interval (B6). Every decision is recorded in `decision-index.md`.

## §B Known issues and evidence map

| Item | Evidence (verified numbering) | Code location (re-read) | Status |
|---|---|---|---|
| R1 renewal (card body item 1) | lens-4 defects 1, 27 | `card_transition.go:311-316`, `:408-429`, `:441-463`; `factory_card.go:1737`; `defaults.go:61-65` | Active (item 1) |
| R2 recovery (card body item 2) | lens-4 defect 9 (lens-4 §7 case (a)) | `factory_card.go:1125-1128` (in-flight refusal); `factory_card.go:189-201` (next refuses non-parent checkouts) | Active (item 2) |
| R3 sync→run (card body item 7) | decision `d-20261009T235640Z-3653` | `card_transition.go:133-181` (no sync-audit→run row) | Retired (006f); spec.md HISTORY 0.2.0 |
| R4 re-stamp (card body item 7) | lens-4 defect 7 (holder part); decision `d-20261009T180009Z-ae0f` | `card_transition.go:133-181` (no same-stage audit row); `:564` | Retired (006f); spec.md HISTORY 0.2.0 |
| R5 worktree path (decision `d-20261009T182157Z-8ccf`; the re-lease of item 2) | decision `d-20261009T182157Z-8ccf` | `card_transition.go:520-545`; `internal/homestate/card_worktree.go:20-46`; `internal/cli/factory_card.go:569-602` | Active (item 2, re-lease); path source and timing closed (Q10) |
| R6 interim guard (card body item 4) | lens-2 D1 (P0); lens-4 defect 8 | `card_transition.go:138`, `:546-549`, `:751-756` | Retired (006f); lens-2 D1 stays open (spec.md Residual risk) |
| Item (3) pushed→done and out-of-band detection (card body item 3) | lens-4 defect 2 (pushed→done); card body (out-of-band claim, not re-measured) | `card_transition.go:156-158`, `:207-209`, `:668-670`; `card_evidence.go:64-73`; `internal/cli/factory_card.go:2803` (decide); premises (l) | Transferred to 3.2-2-2 (t1639) by `d-20261010T064450Z-68b2`; not active here (`progress.md` § G) |

### Corrections to the card brief's defect numbers

The brief's numbering was checked against the lens-4 defect table. This plan uses the lens-4 numbering.

1. R2 cites defects 2 and 9. Defect 2 is the remote done edge (lens-4 cites `card_transition.go:207-209, 668-670`). The lease-recovery defect is 9, so R2 cites 9.
2. R6 (retired) cited defects 4 and 8. Defect 4 is the stage-to-kickoff commit-then-error. The empty-stage bypass is 8. The card body's item (4) names the same interim guard, so R6 retires with that item.
3. The out-of-scope list has two labels swapped. "Run migration" is lens-4 defect 8. "Stage kickoff commit-then-error" is defect 4.
4. Defect 11 appears twice in the brief. Its table entry is the broker token rotation and null list. The stage help evidence table has no numbered entry; it is lens-4 §7 case (d) and design recommendation 5, so the spec lists it without a number.

## § Verified premises

Each item names its source. "Read" means the line was opened in this worktree. "Run" means a read-only project command was executed. "Reported" means the claim comes from a lens report and was not re-measured.

### (a) Retired with old REQ-FRR-012

The SPEC-id premise for the empty-stage edge (T7 refuses a card whose SPEC id is invalid, and the empty-stage edge cannot read the SPEC id) served the interim guard of old item (4). That guard is retired (spec.md HISTORY 0.2.0). The premise is recorded here only so that a reinstatement can re-read it.

### (b) Retired with old REQ-FRR-012

The leader's settlement of the sync-to-run edge and the operator carrier of the empty-stage edge served old REQ-FRR-008 and old REQ-FRR-012. Both are retired. The operator verdict Q6 (c) is therefore not applied (decision-index Q6).

### (c) Plan-audit model and backend

- Read: `.moai/config/sections/llm.yaml:9` sets `agent_overrides: {}`. `agent_overrides_consume` is absent from `llm.yaml`, so the Go field (`internal/config/types.go:334`) takes its zero value `false`. The plan audit's spawn therefore passes no model and runs on the session model.
- Read: `.moai/config/sections/workflow.yaml:36-39` sets the audit gates to `claude: off`, `codex: required`, `glm: off`. Codex is the required backend.

### (d) Decision records, pinned verbatim

Each line was obtained with `moai decision read --all` and copied unmodified. The board output carries no digest field, so these records are quoted, not cited with a digest prefix (decision-index.md, Routing).

R5, record `d-20261009T182157Z-8ccf` (card t1619, precedent for R5):

```text
decision record: decided_by=leader claude (run tmnboq, session 069be28e) evidence_refs=lane-14 blocker 2026-10-09T18:2xZ (T5 refused: 'commit 520e2e0df… does not resolve in' with an empty path); factory.db cards row tmnboq/t1619 worktree_path was empty, other leased cards carry their tree path; backup /Users/goos/.moai/db/moai-adk-go-1bd3d038/factory/factory.pre-t1619-worktree-20261009T1821Z.db; tree .moai/worktrees/t1619 HEAD 520e2e0df on WT-moai-update-repair ladder_path=leader field repair of a record field the lease did not stamp id=d-20261009T182157Z-8ccf scope=card:t1619 kind=ruling
```

Repair round 1 ruling, record `d-20261010T010044Z-57f8` (card t1621; the disposition of the audit's D1, D12, and D15 items; cited as an evidence note only):

```text
decision record: decided_by=leader session 069be28e (factory leader, run tmnboq) evidence_refs=.moai/worktrees/t1621/.moai/reports/t1621/plan-audit-iter1.md; lane-1 report 2026-10-10 ladder_path=leader ruling (verification-completeness.md 2.1; verification-claim-integrity.md 2.3) id=d-20261010T010044Z-57f8 scope=card:t1621 kind=ruling
```

The records `d-20261009T235640Z-3653` (R3) and `d-20261009T180009Z-ae0f` (R4) were pinned in revision 0.1.1. Their subjects are retired, so their pins are removed here.

### (e) The lease, renewal, recovery, and worktree path

- Read: `RenewLease` (`card_transition.go:441-463`) checks expiry before the holder. For an expired lease it reaps the card (`applyLeaseExpiry`, `:452-457`) and returns `ErrLeaseExpired`, for any caller and not only the holder (the holder check is at `:458-460`). Its only caller is the stage path (`internal/cli/factory_card.go:1737`), which runs after the transition succeeds (`:1733`). The renewal therefore carries the commit-then-error pattern of lens-4 defect 1.
- Read: `factoryEnsureCardWorktree` (`internal/cli/factory_card.go:581-602`, doc at `:569-576`) is the record step of a successful lease. It runs after the lease commits: the next path leases at `:1590` and records at `:1596`. It creates the tree and records the path through `RecordCardWorktree`. Its other callers are the MCP factory next (`internal/cli/mcp_factory_card.go:172`) and the auto lane (`internal/cli/todo_auto_lane.go:113`).
- Read: `RecordCardWorktree` (`internal/homestate/card_worktree.go:20-46`) refuses a different path once one is recorded (`:42-43`) and sets the path (`:46`). `PreviousCardWorktree` (`:64-90`, card t1521) carries a binding from an earlier run, called from `internal/cli/factory_card.go:989`.
- Read: the lease write (`card_transition.go:540-545`) sets the holder, heartbeat, and expiry and writes no worktree path. This is the premise of R5.
- Read: the transition path reaps the same way before any other check (`card_transition.go:311-316`). The refusal text is `card %s lease held by %q expired at %s` (`:315`). `ErrLeaseExpired` itself reads `card lease expired` (`:22`). Neither names a recovery command.
- Read: the next verb refuses a card in flight, including the holder's own expired row (`internal/cli/factory_card.go:1125-1128`). It also refuses to run from a non-parent checkout (`factoryAssertParentCheckout`, `:189-201`, refusal text at `:197-199`), so a lane inside its worktree cannot reach R2 through next. The stage verb's implementation (`factoryStageCard`, `:1702-1741`) has no parent-checkout check, so the worktree-reachable recovery path is the stage verb.
- Read: `internal/homestate/card_record.go:91-93` binds the package variable `FactoryLeaseDuration` to `config.DefaultFactoryLeaseDuration` (`internal/config/defaults.go:65`, 15 minutes). The queue claim lease uses the constant directly (`internal/factory/backlog_store.go:1244`), and its test asserts about 15 minutes (`internal/factory/backlog_claim_test.go:154-155`). R1 must leave the constant and the queue lease unchanged (REQ-FRR-010).
- Run (search, plan phase): no `factory renew` verb exists in `internal/cli`, and `workflow.yaml` has no lease-duration key.
- Reported (lens-4 §7 case (a) and defect 4): the reaping order and the stage path's commit-then-error are the lens-4 descriptions of the same code.

### (f) The transition table

- Read: the table (`card_transition.go:133-181`) has T10 run → sync (`:149`) and T11 to T13 among the sync and sync-audit states. T17 is merged-local → pushed (`:156`, guardPush); T18 is merged-local → done (`:157`, guardNoRemote); T20 is ci-green → done (`:158`, guardApprovalDone). It has no pushed → done row. The reserved pushed → ci-green edge is at `:203-208`. Item (3) would add a pushed → done edge.
- Read: the count-pin test `TestFR_AC005_TransitionTableEdgeCount` (`internal/homestate/fr_transition_test.go:146-166`) pins 71 accepted pairs and reports 73 table rows (`:165`). It stays in the code as a regression test. Its criterion (AC-FRR-016) is retired with the three rows it counted.

### (g) Retired with old REQ-FRR-008

The verdict parser has no own-scope field (`internal/auditverdict/verdict.go:114-209`). The sync-to-run ruling discriminator that depended on this premise is retired with old REQ-FRR-008 (spec.md HISTORY 0.2.0).

### (h) Watchdog cadence against the lease

- Read: the standing awaken arms a recurring check with `7,27,47 * * * *`, that is every 20 minutes (`.claude/rules/moai/workflow/auto-semantics.md:108`). The lease lasts 15 minutes, so a renewal at each watchdog pass cannot keep a lease alive between passes.
- Read: the watchdog's three measurement channels contain no lease channel (`.claude/skills/moai-lane-watchdog/SKILL.md` §1).
- Default applied (decision-index Q3): the doctrine cadence and the default lease duration do not change in this SPEC. Since 0.4.0, REQ-FRR-005 bounds the renewal interval at one half of the lease duration key, and the wake schedule that would have to meet that bound is an open question (`progress.md` § G; not decided here).

### (i) Tests and catalogue pins the change touches

- Read: `internal/mcp/catalog_test.go:27-28` pins the catalogue size, `:67-68` the write-capable tool set (which lists `factory_next` and `factory_stage`), `:79-80` its count, and `:106-107` the tool-name list. A `factory_renew` tool changes each of them. The tool declarations are in `internal/mcp/catalog.go`, and the MCP factory handlers are in `internal/cli/mcp_factory_card.go`.
- Read: `internal/config/shipped_key_reader_test.go:113` fails when a shipped config key is absent from `internal/config/testdata/shipped_key_inventory.yaml`, which carries a W/P/R/D class for each key.
- Read: the MCP catalogue is `.claude/rules/moai/core/moai-mcp-tools-catalogue.md`, with the template mirror at `internal/template/templates/.claude/rules/moai/core/moai-mcp-tools-catalogue.md`.

### (j) Schema impact

- Read: the lease expiry, heartbeat, holder, worktree path, and evidence hash are existing card columns, and `card_transition.go:823-824` persists the row. No migration is expected. Pre-flight item 2 confirms this against the table definition before M1 starts.

### (k) Interview gate

- Read: `.moai/config/sections/interview.yaml:6` sets `decision_gate: on`. The decision-index rule therefore owes `decision-index.md` alongside the tier artifacts. Revision 0.2.0 rewrites it (Q1 to Q19 and Q21).

### (l) Item (3), transferred to 3.2-2-2 by 68b2 (premises kept as evidence; revision 0.3.0)

- Read: the table rows T17 (`card_transition.go:156`, merged-local → pushed, guardPush), T18 (`:157`, merged-local → done, guardNoRemote), and T20 (`:158`, ci-green → done, guardApprovalDone). No row leads from `pushed` to `done`. The edges from `pushed` are T21 (to needs-decision, `:161`), T25 (to abandoned, `:172-175`), and the reserved `pushed` → `ci-green` (`:207-209`). `pushed` is not a lease-holding state (`internal/homestate/card_record.go:66-68`), so T26 does not apply.
- Read: `transitionTx` refuses the reserved edge with `ErrReservedEdge` before the table lookup (`card_transition.go:333-335`). The no-remote guard refuses merged-local → done while a remote is configured (`:668-670`). The card body's line numbers `:207-209` and `:668-670` match this tree.
- Read: the only edges into `ci-green` are the reserved edge and T22e (`card_transition.go:166`, generated for the fifth target of the loop), which returns a needs-decision card to its recorded resume state. A card's resume state is `ci-green` only if it left `ci-green` (`:161-164`, T21), so no path in this tree creates `ci-green`. A search of `internal/` for `ci-green` and `CardCIGreen` finds no write outside `card_transition.go` and `card_record.go`, only state lists.
- Read: `decideOne` (`internal/cli/factory_card.go:2803`): `--gate push` takes only merged-local cards (`want(homestate.CardMergedLocal)`), with a merged-PR case before it. A `pushed` card has no decide route to `done`. The approval uuid is read from the queue record for `done` and `pushed` requests.
- Read: `PushGateTarget` (`internal/homestate/card_evidence.go:64-73`) returns `pushed` when the repository has a remote, and `done` otherwise.
- Read: `verifyTransitionApproval` (`internal/homestate/leader_approval.go:530-543`) binds the receipt to the run, the card's version, the evidence SHA, and the owner label. `verifyPushed` (`internal/homestate/card_evidence_readers.go:360-384`) checks that the merge commit is an ancestor of the remote-tracking ref, and never fetches.
- Read: every write that bumps the card row version appends an event that carries the new version. Callers of `updateCardRow`: `card_worktree.go:49` with `card.fields` (`:52`); `card_transition.go:425` with `lease.expired` (`:428`); `card_transition.go:777` with `card.transition` (`:807`); `leader_approval.go:401` with `card.dispatch-owner` (`:404`); `card_picked.go:194` with `card.fields` (`:197`). The row insert at `card_picked.go:166` appends `card.transition` at version 1 (`:173`).
- Read: the lease renewal (`card_transition.go:465`) writes `heartbeat_at` and `lease_expires_at` with no version bump and no event. The state-and-version comparison of the transferred REQ-FRR-018 does not see those fields (`progress.md` § G).
- Read: the `record.drift` event (`internal/homestate/card_unavailable.go:186`) carries `card_id` and no version, so the version comparison excludes it.
- Read: the count pin `TestFR_AC005_TransitionTableEdgeCount` (`internal/homestate/fr_transition_test.go:127-183`) pins 71 accepted and 370 refused (`:166-167`). Its refused-pair list (`:179-181`) names merged-local → done, `pushed` → `ci-green`, and ci-green → done as T18, T19, and T20. The test therefore names the reserved `pushed` → `ci-green` edge T19, and `card_transition.go:184` calls T19 "still-reserved".
- Read: no `factory doctor` verb exists. The factory command group is defined at `internal/cli/factory_handoff_recover.go:20`, and no `doctor` command is registered under it. The existing `doctor` commands are top-level or in other packages. The MCP catalogue has no factory doctor tool.
- Card body claim, not measured: "기록의 done 76행 중 75행이 전이 API 밖에서 쓰임" (75 of the 76 done rows were written outside the transition API). This SPEC did not measure it.

## §C Pre-flight (for the implementing lane, before any edit)

1. Build provenance: build the tool from this tree and invoke that build by path. Every tool measurement cited in a verdict must name the judging build's commit beside HEAD (verification-claim-integrity §2.2). The installed binary is not a substitute.
2. Confirm the existing columns for the lease holder, lease expiry, heartbeat, worktree path, and evidence hash in the cards table definition. Expected: no migration.
3. Confirm that the dispatch's `wt` names the same tree that the lease's record step materializes (`.claude/rules/moai/workflow/factory-dispatch.md:111` and `:117`; `internal/cli/factory_card.go:581-602`), before M1 starts.
4. Confirm the names. `TestBacklogClaim_ReclaimsExpired` and `TestBacklogClaim_RenewAfterExpiry` (AC-FRR-015) are present; both were confirmed with `go test -list` at HEAD 35d242f34 in repair round 1. The new names are written in the run phase: `TestBacklogClaim_QueueLeaseIgnoresCardLeaseDuration` (AC-FRR-015) and the names in §F.
5. Transferred with B2 to 3.2-2-2 (`progress.md` § G): the retry question for a card worktree that was materialized but not recorded is not a pre-flight item of this SPEC.
6. Transferred with item (3) to 3.2-2-2 (`progress.md` § G); not a pre-flight item of this SPEC.

## §D Constraints

- Local landing only. No push, no pull request, and no remote operation from this lane (lane boundary). The card branch is `WT-factory-record-fix`.
- Out of scope and not touched here: t1616 (lane-13) owns the local main landing. t1618 (picked; this lane has not verified its progress) owns the ceiling engine and the verdict file format. It references this SPEC for the lease renewal verb and the TTL configuration key, which this SPEC owns. Old REQ-FRR-010 is retired, and this plan does not touch the verdict parser. The done-transition opening is transferred to 3.2-2-2 (`d-20261010T064450Z-68b2`); t1616 does not own it.
- Plan audit: runs on the session model, and codex is the required backend (§ Verified premises, item (c)).
- Template-First: edit under `internal/template/templates/` first, run `make build`, then mirror the rendered copy into `.claude/` and `.moai/`.
- Watchdog boundary amendment: the watchdog's Boundary 3 (SKILL.md) gains exactly one write, the lease heartbeat through the renew verb. Nothing else changes. REQ-FRR-005 states the same write.
- No schema migration (§ Verified premises, item (j)). No time estimates appear in this plan; priority labels only.
- Configuration key: `workflow.factory_lease.duration`, default `15m`, in the `workflow.*` namespace. Operator verdict (a), keep the name (decision-index Q18, `d-20261010T053034Z-1a19`).
- Repository-local overrides that a template sync must not overwrite are named in QG-4 (acceptance.md), and recorded as decision-index Q19.

## §E Self-verification (plan level)

Traceability, requirement to criterion to milestone, for the active set (0.4.0 numbering; criteria keep their identifiers):

| Requirement | Criterion | Milestone | Item |
|---|---|---|---|
| REQ-FRR-001 | AC-FRR-001 | M2 | renew verb |
| REQ-FRR-002 | AC-FRR-002 | M2 | renew verb, MCP tool |
| REQ-FRR-003 | AC-FRR-003 | M1 | lease-duration key |
| REQ-FRR-004 | AC-FRR-004 | M2 | holder's refused renewal of an expired lease |
| REQ-FRR-005 | AC-FRR-005 | M6 | watchdog renewal |
| REQ-FRR-006 | AC-FRR-006 | M3 | one-step recovery (holder) |
| REQ-FRR-007 | AC-FRR-007 | M3 | refusal names the recovery verb |
| REQ-FRR-008 | AC-FRR-011 | M1 | worktree path at acquisition, including the re-lease |
| REQ-FRR-009 | AC-FRR-014 | M2 | non-holder renewal of a live lease refused |
| REQ-FRR-010 | AC-FRR-015 | M1 | queue claim independent of the card key |

Every active requirement has at least one criterion. The active set has 10 requirements and 10 criteria, within the Tier M ceilings (16 and 16). Former milestone M8 is transferred to 3.2-2-2 and is not a milestone of this SPEC (see the note on M4 and M5 below). These checks are plan-level reads of the written artifacts; the code checks belong to the lane.

## §F Milestones

Ordered by decision reversibility: data model and configuration first, then the new interface, then the user-facing flow, then the mechanical sync. Priority labels only. Milestones M4 and M5 are retired (see the note at the end of this section); their numbers are not reused.

### M1: Lease data model, configuration, and acquisition path (Priority High; REQ-FRR-003, -008, -010)

Files:
1. `internal/template/templates/.moai/config/sections/workflow.yaml` and its rendered copy `.moai/config/sections/workflow.yaml`: add the lease-duration key (`workflow.factory_lease.duration`, default `15m`; Q18). Keep the repository-local overrides in the rendered file intact: `branch_guard.enabled`, `branch_guard.deny_commits_on`, and `commit_identity_guard.enabled` (Q19).
2. `internal/config/types.go` (struct field and YAML tag), `internal/config/defaults.go` (default `15m`; `DefaultFactoryLeaseDuration` keeps its value), and the loader wiring in `internal/config/`, including the symmetry test's case list.
3. `internal/config/testdata/shipped_key_inventory.yaml`: an entry for the new key with its class.
4. `internal/homestate/card_record.go:91-93`: the package variable stays as the default. The factory handle carries the configured duration, set when `OpenFactory` runs. The lease acquisition (`card_transition.go:542`) and the renewal (`:463`) read the handle's value.
5. Acquisition path (REQ-FRR-008): the card worktree is materialized first, by the factory's record step (`internal/cli/factory_card.go` `factoryEnsureCardWorktree`, `:581-602`), and the path is recorded as part of the acquisition. A failed materialization refuses the acquisition. A recorded path that differs from the materialized tree is refused (`card_worktree.go:42-43`). The transaction boundary of the record call, and the order of the expiry check against it, are transferred with B2 to 3.2-2-2 (`progress.md` § G) and are not decided here.

#### Closed before M1: R5 path source and timing (plan D5; decision-index Q10, operator verdict (ii) at T3)

- Source: the card worktree that the factory materializes and records through `RecordCardWorktree` in the record step `factoryEnsureCardWorktree` (`internal/cli/factory_card.go:581-602`). The dispatch's `wt` names the same tree (`.claude/rules/moai/workflow/factory-dispatch.md:111` and `:117`).
- Not chosen, first: a path the lane supplies. `factory next` runs from the parent checkout (`factoryAssertParentCheckout`, `factory_card.go:189-201`), and the MCP form derives the path internally (`internal/cli/mcp_factory_card.go:172-176`), so no lane-supplied path reaches the lease.
- Not chosen, second: a path recorded at picked alone. The field-apply path accepts a caller-supplied absolute path (`card_picked.go:91-92`, `:115`) that the factory never checks against the materialized tree.
- Timing: the lease acquisition itself, T3. The defect is that the lease can complete without a path: the lease write (`card_transition.go:540-545`) writes none, and the record step runs after the lease (`factory_card.go:1590-1600`). A later, different path is refused (`card_worktree.go:42-43`).
- Placement: transferred with B2 to 3.2-2-2 (`progress.md` § G). This SPEC does not choose the transaction boundary of the record call.
- A carried binding from an earlier run (`factory_card.go:989`, card t1521) passes through the same `RecordCardWorktree` and keeps its behavior.

RED first (predicted failures; not gates):
- Config: `TestFactoryLeaseDurationDefault` asserts the default is 15m and that a configured `20m` is read. On the unchanged tree the key does not exist, so the assertion fails. Record the failure output. A compile failure would be TOOL_FAILURE, not a predicted failure.
- Homestate: `TestLeaseRecordsWorktreePath` acquires a lease and asserts that the card row records the path. On the unchanged tree the path stays empty, because `card_transition.go:540-545` writes only the holder, heartbeat, and expiry. Record the failure output.
- CLI: `TestLeaseRefusedOnMaterializationFailure` (AC-FRR-011) forces the materialization to fail and asserts that the acquisition is refused with no lease and no path. On the unchanged tree the lease is acquired first (`factory_card.go:1590-1600`).
- Regression, not RED: the existing claim-lease tests in `internal/factory` stay green, and the new `TestBacklogClaim_QueueLeaseIgnoresCardLeaseDuration` is observed green before M1 starts (REQ-FRR-015).

### M2: Renewal verb and MCP tool (Priority High; REQ-FRR-001, -002, -004, -009)

Files:
1. `internal/homestate/card_transition.go` `RenewLease` (`:441-463`) already exists and is reached only by the stage path (`internal/cli/factory_card.go:1737`). Its checks move so that the holder and the expiry are judged before any write: for the holder on a live lease it extends from the request time (REQ-FRR-001); for the holder on an expired lease it refuses with no write (REQ-FRR-004; Q9); for a caller that is not the holder of a live lease it refuses with no write (REQ-FRR-009). The stage path's renewal call keeps its behavior, pinned by the existing stage tests.
2. `internal/cli/factory_card.go`: a `renew` subcommand built the way `stage` is built, with run and lane resolution as in `factoryStageCard` (`:1702`), plus help text and an output line that shows the new expiry.
3. `internal/cli/mcp_factory_card.go`: the MCP handler calls the same implementation the CLI calls, following the pattern documented at `factory_card.go:1696-1699`.
4. `internal/mcp/catalog.go` (tool declaration) and the pins in `internal/mcp/catalog_test.go:27-28, 67-68, 79-80, 106-107`.
5. Documentation of the tool is in M7.

RED first (predicted failures; not gates):
- `TestFactoryRenewExtendsLiveLease` (cli, AC-FRR-001): with a live lease of exactly 5 minutes, renewal sets the expiry to the request time plus the duration. Predicted: the verb is not registered.
- `TestRenewRefusesExpiredWithoutWrite` (homestate, AC-FRR-004): an expired lease held by the caller is refused, and the version, state, and event count are unchanged. Predicted on the unchanged tree: `RenewLease` runs `applyLeaseExpiry` (`:452-457`), which bumps the version and appends `lease.expired`.
- `TestNonHolderRenewRefusedCLIAndMCP` (cli, AC-FRR-014): the same refusal through the verb and through the MCP handler, on a live lease.
- `TestFactoryRenewMatchesMCP` (cli, AC-FRR-002): the same card and caller through the verb and through the MCP handler give equal results. Predicted: no tool exists.
- Catalogue pins: the expected count and the write-capable set change in the same edit. Predicted: the catalogue test fails until the tool is declared.

### M3: One-shot recovery and refusal text (Priority High; REQ-FRR-006, -007; REQ-FRR-008 for the re-lease)

Files:
1. `internal/homestate/card_transition.go` `transitionTx` (`:309-316`): when the caller is the holder of an expired lease, collect the lease, re-lease the card to that holder, and apply the requested edge in the same transaction. The re-lease is an acquisition under REQ-FRR-008; the placement of its materialization and path record is the B2 subject transferred to 3.2-2-2 (`progress.md` § G), so this item does not choose it. If the requested edge is refused, the whole step is refused and the expired row is unchanged. The refusal text (`:315`) names the factory stage verb (REQ-FRR-007; Q8). A caller that is not the holder is not specified by this SPEC (B5, transferred).
2. An expiry that returns the card to blocked (a merging card, `applyLeaseExpiry` `:416-420`) is not recovered by this step. It stays on the human unblock path.
3. `internal/cli/factory_card.go` `next --card` refusal (`:1125-1128`): for the holder's own expired row, collect and re-lease instead of refusing as in flight. The verb keeps its parent-checkout rule (`:189-201`); defect 26 is out of scope.
4. The renewal refusal text (`RenewLease`, `:456`) names the same stage verb (REQ-FRR-007; Q9).

RED first (predicted failures; not gates):
- `TestStageRecoversOwnExpiredLease` (cli, AC-FRR-006): the holder's expired lease at run, then `stage <card> run`. One invocation succeeds; the events read the expiry and then the transition; the card ends leased to the holder at run. Predicted: the transition returns `ErrLeaseExpired` (`:315`).
- `TestNextRecoversOwnExpiredLeaseFromParentCheckout` (cli, AC-FRR-006): the same through the next verb from the parent checkout. Predicted: refused as in flight (`factory_card.go:1125-1128`).
- `TestExpiredRefusalNamesRecoveryVerb` (homestate, AC-FRR-007): a non-holder request on an expired card, and the holder's own refused renewal, both return text that names the stage verb. Predicted: the texts at `:315` and `:456` name no verb.
- `TestRecoveryAllOrNothing` (homestate, AC-FRR-006): an illegal requested edge leaves the expired row unchanged. Predicted on the unchanged tree: the expired row is reaped (version bump, `lease.expired`) before the refusal, so the unchanged-row assertion fails.

### M4 and M5: retired (decision `d-20261010T053435Z-006f`)

M4 and M5 are retired (decision `d-20261010T053435Z-006f`): M4 (holder re-stamp and sync-audit re-entry: old REQ-FRR-008 to -010; AC-FRR-008 [RETIRED], AC-FRR-009 [RETIRED], AC-FRR-010 [RETIRED], and the count pin AC-FRR-016 [RETIRED]) and M5 (interim empty-stage guard: old REQ-FRR-012 and -013; AC-FRR-012 [RETIRED] and AC-FRR-013 [RETIRED]). Their file lists and RED lists are removed from this plan. Milestone numbers M1, M2, M3, M6, and M7 are kept so that cross-references stay valid. M8 (pushed-to-done edge and factory doctor) is transferred to 3.2-2-2 by `d-20261010T064450Z-68b2`; its number is not reused in this SPEC.

### M6: Watchdog renews first (Priority Medium; REQ-FRR-005)

Files:
1. `internal/template/templates/.claude/skills/moai-lane-watchdog/SKILL.md` (edit first): `allowed-tools` (line 17) gains `Bash(moai factory renew:*)`; §1 gains a first step that renews the lane's own card lease before the three-channel measurement, states the renewal interval bound of REQ-FRR-005 (no longer than one half of `workflow.factory_lease.duration`), and states that a refused renewal is recorded and the measurement still runs; Boundary 3 names the renewal heartbeat as the one write added.
2. `make build` to refresh the embedded template, then mirror the rendered file to `.claude/skills/moai-lane-watchdog/SKILL.md`.
3. `.claude/rules/moai/workflow/auto-semantics.md` §3 and its template mirror gain the renewal step. The cadence at line 108 is not changed by this SPEC; whether it should change is an open question (`progress.md` § G).

RED first (predicted failure; not a gate):
- `TestLaneWatchdogRenewsFirst` (AC-FRR-005): a template test asserts that the watchdog procedure names the renew verb before the measurement step, that `allowed-tools` lists the renew command, that the procedure states the renewal interval bound, and that it states a refused renewal is recorded while the measurement still runs. Predicted: the current SKILL.md has none of these (`allowed-tools` at line 17; §1 opens with measurement).

### M7: Catalogue and rule mirrors (Priority Low; mechanical; no behavior change)

Files:
1. `.claude/rules/moai/core/moai-mcp-tools-catalogue.md` and its template mirror: a `factory_renew` row.
2. `.claude/rules/moai/workflow/factory-dispatch-mechanics.md` and its template mirror `internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md`: one sentence on the renew verb in the lease discussion.
3. Rendered copies regenerated with `make build`; no rendered copy is edited first.

RED first: none for documentation. The M2 catalogue test is the check.

## §G Anti-patterns to avoid

- Reaping inside a call that then returns an error. A refusal leaves the row unchanged, except when the holder's own recovery is the requested action (M2, M3).
- Changing the value of `DefaultFactoryLeaseDuration`, which the queue claim lease and its tests share.
- Touching the verdict parser or the verdict file format (t1618).
- Widening `next` so that it runs from a card worktree (defect 26 is out of scope).
- Citing a count from a report instead of measuring it.
- Re-introducing the retired meanings of old REQ-FRR-008 to -010, REQ-FRR-012, or -013, or the M4 and M5 rows, without a new leader or operator decision. They are retired by `d-20261010T053435Z-006f`.
- Recording the worktree path outside the lease acquisition, or choosing the transaction boundary of the record call in this SPEC (transferred with B2).
- Reusing AC-FRR-016 [RETIRED] or any other retired criterion identifier.

## §H Cross-references

- `.claude/rules/moai/workflow/auto-semantics.md` §3, §5.1, and §9: the watchdog channels, the cadence, and the gate inventory. The cadence statement is at line 108.
- `.claude/skills/moai-lane-watchdog/SKILL.md`: the watchdog procedure and its boundaries.
- `.claude/rules/moai/development/verification-completeness.md` §1.1, §1.3, §2, and §2.1: the observed-failure rule, continued firing, the two-cell pairing, the wrong-reason red, and the undecidable disposition. The active criteria are regression guards, so the release-gate reading of §2 does not apply to them (decision-index Q1).
- `.claude/rules/moai/workflow/tdd-result-contract.md`: EXPECTED_RED and TOOL_FAILURE.
- `.claude/rules/moai/core/verification-claim-integrity.md` §2.2 and §2.3: build provenance for tool measurements, and the commit graph as the witness for ordering.
- `.claude/rules/moai/workflow/factory-dispatch-mechanics.md`: the lease discussion updated in M7.
- `decision-index.md`: every decision this revision carries, Q1 to Q19, Q21, and Q22 (Q20 is withdrawn), with its label, class, status, options, basis, and evidence.
- `progress.md`: the plan-phase §E skeleton, with §E.1 stating that the plan is awaiting re-audit.

## Decisions owed (plan numbering, mapped to decision-index.md)

The plan's own numbering of its open judgments is kept here for traceability. Each entry names the decision-index row that records it and its status.

- D1 (brief numbering, §B corrections): kept lens-4 numbering. Q15, default applied.
- D2 (R6 operator carrier): retired with old REQ-FRR-012. Q6 is not applicable (006f). The 1a19 value (c) is not written.
- D3 (R3 own-scope discriminator): retired with old REQ-FRR-008. Q4 is not applicable (006f). The 1a19 value (d) is not written.
- D4 (watchdog cadence against the 15-minute lease): cadence and default duration unchanged in this SPEC; since 0.4.0, REQ-FRR-005 bounds the renewal interval (`d-20261010T064450Z-68b2`, B6). The cadence question is open (`progress.md` § G). Q3, default applied.
- D5 (R5 path source and timing): source (ii) and timing T3 per 1a19 (Q10). Closed before M1. The placement of the record call is transferred with B2 (`progress.md` § G).
- D6 (recovery verb in the refusal text): `moai factory stage`. Q8, operator verdict (a), 1a19.
- D7 (renewal of an expired own lease): refused with no write; recovery stays on the transition path. Q9, operator verdict (a), 1a19.
- D8 (residual empty-SPEC route): retired with old REQ-FRR-013. Q13 is not applicable (006f). The residual is carried by the Residual risk of spec.md (lens-2 D1).
- D9 (lens-4 defects 10, 13, 14, 19, 20, 22, 24, 25, and 26): stay unassigned. Q14, default applied.
- D10 (decision-index.md and progress.md): decision-index.md rewritten (Q1 to Q19 and Q21); progress.md skeleton kept (Q12).
- D11 (module field): `internal/homestate` kept as the brief specifies. Q16, default applied.
- D12 (priority P0): kept. Q17, default applied.
- D13 (configuration key name `workflow.factory_lease.duration`): kept. Q18, operator verdict (a), 1a19.
- D14 (B1, criteria classification): the surviving criteria are regression guards and the release-gate wording is removed. Q1, operator verdict, 1a19.
- D15 (item (3), pushed-to-done with out-of-band detection): written under leader ruling `d-20261010T060606Z-c5c4` (option a) as old REQ-FRR-016 to -019 and milestone M8; transferred to 3.2-2-2 by `d-20261010T064450Z-68b2` (Q21; subject moved).
- D16 (audit iteration 3, B1; renumbering of the requirements): 0.4.0 renumbers to REQ-FRR-001 to -010 and reuses old REQ-FRR-008 to -010 with the map in spec.md § Requirements (Identifier map). Q22, decided by `d-20261010T064450Z-68b2`.
- D17 (B2, B4, and B5 transfer; B3 and B6 repairs): by `d-20261010T064450Z-68b2`; reasons and moved identifiers in `progress.md` § G. The 1a19 verdicts stay on their surviving rows.
