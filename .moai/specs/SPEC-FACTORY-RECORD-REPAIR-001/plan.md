---
id: SPEC-FACTORY-RECORD-REPAIR-001
version: "0.1.0"
updated: 2026-10-10
---

# Plan: SPEC-FACTORY-RECORD-REPAIR-001

## §A Context

- Card t1621, Class C, Tier M. Work root: worktree `.moai/worktrees/t1621`, branch `WT-factory-record-fix`, base HEAD `2aab5f797`. This plan is the plan phase only; the implementing lane writes the code and commits.
- Evidence is read-only and untracked. The lens-4 and lens-2 reports live in the primary checkout at `/Users/goos/MoAI/moai-adk-go/.moai/reports/audit-20261010/` and are not present in this worktree. Each code location below was re-read in this worktree. § 검증된 전제 states, for each premise, whether it was read, run, or only reported by a lens.
- Problem: the factory record can strand a holder's card (expired lease, no renewal, no recovery), can admit a SPEC-bearing card to run without a plan verdict, and has no re-entry from a failed sync audit. Details are in spec.md Background.

## §B Known issues and evidence map

| Item | Evidence (verified numbering) | Code location (re-read) | Status |
|---|---|---|---|
| R1 renewal | lens-4 defects 1, 27 | `card_transition.go:311-316`, `:408-429`, `:441-463`; `factory_card.go:1737`; `defaults.go:61-65` | In scope |
| R2 recovery | lens-4 defect 9 (lens-4 §7 case (a)) | `factory_card.go:1125-1128` (in-flight refusal); `factory_card.go:193-199` (next refuses non-parent checkouts) | In scope |
| R3 sync→run | decision `d-20261009T235640Z-3653` | `card_transition.go:133-181` (no sync-audit→run row) | In scope; discriminator open (D3) |
| R4 re-stamp | lens-4 defect 7; decision `d-20261009T180009Z-ae0f` | `card_transition.go:133-181` (no same-stage audit row); `:564` (verdict read against the stored hash) | In scope; holder only |
| R5 worktree | decision `d-20261009T182157Z-8ccf` | `card_transition.go:520-545`; `card_worktree.go:20-46` | In scope; path source open (D5) |
| R6 interim guard | lens-2 D1 (P0); lens-4 defect 8 | `card_transition.go:138`, `:546-549`, `:751-756` | In scope (interim) |

### Corrections to the card brief's defect numbers

The brief's numbering was checked against the lens-4 defect table. This spec uses the lens-4 numbering.

1. R2 cites defects 2 and 9. Defect 2 is the remote done edge (lens-4 cites `card_transition.go:207-209, 668-670`). The lease-recovery defect is 9, so R2 cites 9.
2. R6 cites defects 4 and 8. Defect 4 is the stage-to-kickoff commit-then-error (lens-4 cites `factory_card.go:1733-1739` and `card_transition.go:458`). The empty-stage bypass is 8, so R6 cites 8.
3. The out-of-scope list has two labels swapped. "Run migration" is lens-4 defect 8, whose minimal fix is the run migration and the T4b guard. "Stage kickoff commit-then-error" is defect 4.
4. Defect 11 appears twice in the brief. Its table entry is the broker token rotation and null list. The stage help evidence table has no numbered entry; it is lens-4 §7 case (d) and design recommendation 5, so the spec lists it without a number.

## § 검증된 전제 (Verified premises)

Each item names its source. "Read" means the line was opened in this worktree. "Run" means a read-only project command was executed. "Reported" means the claim comes from a lens report and was not re-measured.

### (a) Defect 5, stated precisely

- Read: the T7 evidence reader refuses a card whose SPEC id is not valid. `internal/homestate/card_evidence_readers.go:185-186` returns `card carries no valid SPEC id (%q), so the plan-artifact hash cannot be checked` when `specIDPattern` does not match `cur.SpecID`. The pattern is `^SPEC-[A-Z0-9]+(?:-[A-Z0-9]+)*$` (`card_picked.go:15`).
- Read: the same reader serves the kickoff audit (`card_audit_kickoff.go:28`), so the audit-decided kickoff → run edge (T8a) also refuses an empty SPEC id.
- Read: the SPEC id is written only while the card is picked. The field-apply path refuses other states at `card_picked.go:184-186`, and the SPEC id is set at `card_picked.go:114`.
- Read: `card_transition.go` has no SPEC-id guard. Its only references to `SpecID` are the persistence column list (`:823-824`).
- Read: the empty-stage to run edge (T4b, `card_transition.go:138`) is decided by `stageAdmits(stage, to)` (`:751-756`), whose signature carries the stage and the target only. It cannot read the SPEC id.
- Reported (lens-4 §0 and §4, not re-measured): four plan-audit rows carry an empty SPEC id (t1579, t1582, t1584, t1613), and card t1582 entered run in run `tmnboq` with empty SPEC and evidence.
- Precise condition: T7 is refused whenever the row's SPEC id is empty or fails `specIDPattern`, and the SPEC id can be written only while the card is picked. Defect 5 is out of scope, and REQ-FRR-013 keeps the empty-SPEC route unchanged, so the residual is stated rather than assumed.

### (b) R6 and the leader's recorded settlement

- R6 guards only the empty-stage to run edge. It does not touch the sync-audit to run edge of R3.
- The leader's recorded settlement for sync to run is decision `d-20261009T235640Z-3653`, pinned in (d). It is a leader record, not a state edge. R6's allow-condition must keep that operator-decided route open rather than block it.
- Read (lens-2 D2 and recommendation 3): no operator edge exists at T7 today ("there is no operator edge"). The operator branch of R6 therefore has no state-changing carrier yet. See Decisions owed, D2.

### (c) Plan-audit model and backend

- Read: `.moai/config/sections/llm.yaml:9` sets `agent_overrides: {}`. `agent_overrides_consume` is absent from `llm.yaml`, so the Go field (`internal/config/types.go:334`) takes its zero value `false`. The plan audit's spawn therefore passes no model and runs on the session model.
- Read: `.moai/config/sections/workflow.yaml:36-39` sets the audit gates to `claude: off`, `codex: required`, `glm: off`. Codex is the required backend.

### (d) Decision records, pinned verbatim

Each line was obtained with `moai decision read --all` and copied unmodified.

R3, record `d-20261009T235640Z-3653`:

```text
decision record: decided_by=leader claude (run tmnboq, session 069be28e) evidence_refs=lane-2: t1578 sync-audit FAIL 65/100, F1 update_integrity_probe.go:67-76 (FIFO read never returns, AC-UMF-002), F2 :52, F3 :53 + CHANGELOG.md:26; lane-9: t1478 sync-audit FAIL, M6 and F7 need code; the state machine has no sync to run edge (both lanes, and lens-4 finding); factory.db backup /Users/goos/.moai/db/moai-adk-go-1bd3d038/factory/factory.pre-run-reentry-20261009T2356Z.db ladder_path=leader field repair (precedent: state transition repair with a database backup); defects in a card's own scope are repaired in that card id=d-20261009T235640Z-3653 scope=standing kind=standing-rule
```

R4, record `d-20261009T180009Z-ae0f` (card t1579, precedent for R4):

```text
decision record: decided_by=leader claude (run tmnboq, session 069be28e) evidence_refs=lane-6 blocker c6717ee9; factory.db cards row tmnboq/t1579 before: evidence_sha d36e97571a4d…, after: 0cccc0d310d7…; backup /Users/goos/.moai/db/moai-adk-go-1bd3d038/factory/factory.pre-t1579-evidence-20261009T1800Z.db; plan-audit-iter2.md audited_sha 0cccc0d31 ladder_path=leader field repair of a stamped-evidence mismatch (no re-stamp edge exists; precedent t1585) id=d-20261009T180009Z-ae0f scope=card:t1579 kind=ruling
```

R5, record `d-20261009T182157Z-8ccf` (card t1619, precedent for R5):

```text
decision record: decided_by=leader claude (run tmnboq, session 069be28e) evidence_refs=lane-14 blocker 2026-10-09T18:2xZ (T5 refused: 'commit 520e2e0df… does not resolve in' with an empty path); factory.db cards row tmnboq/t1619 worktree_path was empty, other leased cards carry their tree path; backup /Users/goos/.moai/db/moai-adk-go-1bd3d038/factory/factory.pre-t1619-worktree-20261009T1821Z.db; tree .moai/worktrees/t1619 HEAD 520e2e0df on WT-moai-update-repair ladder_path=leader field repair of a record field the lease did not stamp id=d-20261009T182157Z-8ccf scope=card:t1619 kind=ruling
```

### (e) The lease, renewal, recovery, and worktree path

- Read: the T3 lease (assigned → leased) sets the holder, heartbeat, and expiry (`card_transition.go:540-542`) and the worker heartbeat (`:543-545`). It does not write the worktree path. The path is written by `RecordCardWorktree` as a separate `card.fields` event, and that function refuses a different path once one is recorded (`card_worktree.go:20-46`). This is the premise of R5.
- Read: `RenewLease` (`card_transition.go:441-463`) checks expiry before the holder. For an expired lease it reaps the card (`applyLeaseExpiry`, `:452-457`) and returns `ErrLeaseExpired`, for any caller and not only the holder. Its only caller is the stage path (`factory_card.go:1737`). The renewal therefore carries the commit-then-error pattern of lens-4 defect 1.
- Read: the transition path reaps the same way before any other check (`card_transition.go:311-316`). The refusal text is `card %s lease held by %q expired at %s` (`:315`). `ErrLeaseExpired` itself reads `card lease expired` (`:22`). Neither names a recovery command.
- Read: the next verb refuses a card in flight, including the holder's own expired row (`factory_card.go:1125-1128`). The next verb also refuses to run from a non-parent checkout (`factory_card.go:197-199`), so a lane inside its worktree cannot reach R2 through next. The stage verb's shared implementation (`factoryStageCard`, `factory_card.go:1702-1716`) has no parent-checkout check, so the worktree-reachable recovery path is the stage verb.
- Read: `defaults.go:65` defines `DefaultFactoryLeaseDuration = 15 * time.Minute` with the comment "no config key reads it". `card_record.go:92-93` binds the package variable `FactoryLeaseDuration` to it. The queue claim lease uses the same constant (`internal/factory/backlog_store.go:1244, 1322`), and its tests assert 15 minutes (`internal/factory/backlog_claim_test.go:148-155, 275-293`). R1 must leave the constant and the queue lease unchanged (REQ-FRR-015).
- Run (search): no `factory renew` verb exists in `internal/cli`, and `workflow.yaml` has no lease-duration key.
- Reported (lens-4 §7 case (a) and defect 4): the reaping order and the stage path's commit-then-error are the lens-4 descriptions of the same code.

### (f) The transition table

- Read: the table (`card_transition.go:133-181`) has T10 run → sync, T11 sync → sync-audit, T12 sync-audit → sync, and T13 sync-audit → merge-ready. It has no sync-audit → run row, which matches decision `d-20261009T235640Z-3653`. It has no plan-audit → plan-audit row and no sync-audit → sync-audit row, which matches decision `d-20261009T180009Z-ae0f`.
- Read: `internal/homestate/fr_transition_test.go:124` states "exactly the 71 requested edges are accepted and the other 370 are", and `card_transition.go:130` repeats 71 in an `@MX:ANCHOR` reason.
- Reported (lens-4 §Gaps): a hand count of 73 that the lens did not verify. This plan has not measured the count either. Pre-flight item 2 measures it before any row is added.

### (g) The verdict parser has no own-scope field

- Read: `internal/auditverdict/verdict.go` reads the keys `verdict`, `overall_score`, `must_pass_failed`, `blocking_count`, `plan_artifact_hash`, `convergence_overall`, and `required_backend` (lines 114-209). None marks a defect as belonging to the card's own code. R3 therefore cannot separate own-scope blocking defects from other defects on the parsed fields. The discriminator is Decisions owed, D3. The verdict format belongs to t1618.

### (h) Watchdog cadence against the lease

- Read: the standing awaken arms a recurring check with `7,27,47 * * * *`, that is every 20 minutes (`.claude/rules/moai/workflow/auto-semantics.md` §5.1). The lease lasts 15 minutes, so a renewal at each watchdog pass cannot keep a lease alive between passes.
- Read: the watchdog's three measurement channels contain no lease channel (`.claude/skills/moai-lane-watchdog/SKILL.md` §1).
- Decisions owed, D4.

### (i) Tests and catalogue pins the change touches

- Read: `internal/mcp/catalog_test.go:27-28` pins the catalogue size, `:67-68` the write-capable tool set (which lists `factory_next` and `factory_stage`), `:79-80` its count, and `:106-107` the tool-name list. A `factory_renew` tool changes each of them. The tool declarations are in `internal/mcp/catalog.go`, and the MCP factory handlers are in `internal/cli/mcp_factory_card.go`.
- Read: `internal/config/shipped_key_reader_test.go:113` fails when a shipped config key is absent from `internal/config/testdata/shipped_key_inventory.yaml`, which carries a W/P/R/D class for each key.
- Read: the MCP catalogue is `.claude/rules/moai/core/moai-mcp-tools-catalogue.md`, with the template mirror at `internal/template/templates/.claude/rules/moai/core/moai-mcp-tools-catalogue.md`.

### (j) Schema impact

- Read: the lease expiry, heartbeat, holder, worktree path, and evidence hash are existing card columns, and `card_transition.go:823-824` persists the row. No migration is expected. Pre-flight item 4 confirms this against the table definition before M1 starts.

### (k) Interview gate

- Read: `.moai/config/sections/interview.yaml:6` sets `decision_gate: on`. The decision-index rule therefore owes `decision-index.md` alongside the tier artifacts. This batch writes the three files the lane specified; `decision-index.md` is Decisions owed, D10.

## §C Pre-flight (for the implementing lane, before any edit)

1. Build provenance: build the tool from this tree and invoke that build by path. Every tool measurement cited in a verdict must name the judging build's commit beside HEAD (verification-claim-integrity §2.2). The installed binary is not a substitute.
2. AC-005 baseline: run the pinned test in `internal/homestate/fr_transition_test.go` (the test at line 124) once on the unchanged tree, and record the command, the verbatim output, and the observed count. This settles 71 against the hand count of 73.
3. Confirm the baseline table has no sync-audit → run row and no same-stage audit rows (`card_transition.go:133-181`).
4. Confirm the existing columns for the lease holder, lease expiry, heartbeat, worktree path, and evidence hash in the cards table definition. Expected: no migration.
5. Confirm the names of the claim-lease test in `internal/factory/backlog_claim_test.go` (near lines 148-155 and 275-293) and of the AC-005 test in `fr_transition_test.go`. AC-FRR-015 and AC-FRR-016 cite them.

## §D Constraints

- Local landing only. No push, no pull request, and no remote done edge (lane boundary). The card branch is `WT-factory-record-fix`.
- Out of scope and not touched here: t1616 (lane-13) owns the done-transition opening and the local main landing; t1618 (not started) owns the ceiling engine and the verdict file format. REQ-FRR-010 touches the holder edge only, not the verdict parser.
- Plan audit: runs on the session model, and codex is the required backend (§ 검증된 전제, item (c)).
- Template-First: edit under `internal/template/templates/` first, run `make build`, then mirror the rendered copy into `.claude/` and `.moai/`.
- Watchdog boundary amendment: the watchdog's Boundary 3 (SKILL.md) gains exactly one write, the lease heartbeat through the renew verb. Nothing else changes.
- No schema migration (§ 검증된 전제, item (j)). No time estimates appear in this plan; priority labels only.
- Configuration key proposed: `workflow.factory_lease.duration`, default `15m`, mirroring the shape of the existing `workflow.slot_lease.default_max_duration`. The name is a naming choice (Decisions owed, D13).
- Not written in this batch: `decision-index.md` (Decisions owed, D10) and the `progress.md` §E skeleton, because the lane specified exactly three files.

## §E Self-verification (plan level)

Traceability, requirement to criterion to milestone:

| Requirement | Criterion | Milestone |
|---|---|---|
| REQ-FRR-001 | AC-FRR-001 | M2 |
| REQ-FRR-002 | AC-FRR-002 | M2 |
| REQ-FRR-003 | AC-FRR-003 | M1 |
| REQ-FRR-004 | AC-FRR-004 | M2 |
| REQ-FRR-005 | AC-FRR-005 | M6 |
| REQ-FRR-006 | AC-FRR-006 | M3 |
| REQ-FRR-007 | AC-FRR-007 | M3 |
| REQ-FRR-008 | AC-FRR-008, AC-FRR-016 | M4 |
| REQ-FRR-009 | AC-FRR-009 | M4 |
| REQ-FRR-010 | AC-FRR-010, AC-FRR-016 | M4 |
| REQ-FRR-011 | AC-FRR-011 | M1 |
| REQ-FRR-012 | AC-FRR-012 | M5 |
| REQ-FRR-013 | AC-FRR-013 | M5 |
| REQ-FRR-014 | AC-FRR-014 | M2, M3 |
| REQ-FRR-015 | AC-FRR-015 | M1 |

Every requirement has at least one criterion. The plan has 15 requirements and 16 criteria, within the Tier M ceilings. Each behavior change has a RED-first note in §F. These checks are plan-level reads of the written artifacts; the code checks belong to the lane.

## §F Milestones

Ordered by decision reversibility: data model and configuration first, then the new interface, then the user-facing flow, then the transition table, then the guard, then the mechanical sync. Priority labels only.

### M1: Lease data model and configuration (Priority High; REQ-FRR-003, -011, -015)

Files:
1. `internal/template/templates/.moai/config/sections/workflow.yaml` and its rendered copy `.moai/config/sections/workflow.yaml`: add `factory_lease.duration: 15m` (proposed key).
2. `internal/config/types.go` (struct field and YAML tag), `internal/config/defaults.go` (default `15m`; `DefaultFactoryLeaseDuration` keeps its value), and the loader wiring in `internal/config/`, including the symmetry test's case list.
3. `internal/config/testdata/shipped_key_inventory.yaml`: an entry for the new key with its class.
4. `internal/homestate/card_record.go:92-93`: the package variable stays as the default. The factory handle carries the configured duration, set when `OpenFactory` runs. The lease acquisition (`card_transition.go:542`) and the renewal read the handle's value.
5. `internal/homestate/card_transition.go` `guardLeaseAcquire` (`:520-545`): record the lane's worktree path when none is recorded, and refuse when a different path is recorded (REQ-FRR-011). The path source is Decisions owed, D5.

RED first:
- Config: a test asserts that the default is 15m and that a configured `20m` is read. On the unchanged tree the field does not exist, so the test fails to compile. Record the failure output.
- Homestate: `TestLeaseRecordsWorktreePath` acquires a lease and asserts that the card row records the path. On the unchanged tree the path stays empty, because `:540-545` write only the holder, heartbeat, and expiry. Record the failure output.
- Regression, not RED: the existing claim-lease tests in `internal/factory` stay green (REQ-FRR-015).

### M2: Renewal verb and MCP tool (Priority High; REQ-FRR-001, -002, -004, -014)

Files:
1. `internal/homestate/card_transition.go`: a renewal entry that checks the lease before any write. For the holder on a live lease it extends from the request time (REQ-FRR-001). For an expired lease it refuses with no write (REQ-FRR-004). For a caller that is not the holder it refuses with no write (REQ-FRR-014). The stage path's existing renewal call (`factory_card.go:1737`) keeps its behavior, pinned by the existing stage tests.
2. `internal/cli/factory_card.go`: a `renew` subcommand built the way `stage` is built, with run and lane resolution as in `factoryStageCard` (`:1702`), plus help text and an output line that shows the new expiry.
3. `internal/cli/mcp_factory_card.go`: the MCP handler calls the same implementation the CLI calls, following the pattern documented at `factory_card.go:1696-1699`.
4. `internal/mcp/catalog.go` (tool declaration) and the pins in `internal/mcp/catalog_test.go:27-28, 67-68, 79-80, 106-107`.
5. Documentation of the tool is in M7.

RED first:
- `TestFactoryRenewExtendsLiveLease` (CLI): with a live lease of about 5 minutes, renewal sets the expiry to the request time plus the duration. RED: the verb does not exist.
- `TestRenewRefusesExpiredWithoutWrite` (homestate): an expired lease held by the caller is refused, and the version, state, and event count are unchanged. RED on the unchanged tree: `RenewLease` runs `applyLeaseExpiry` (`:452-457`), which bumps the version and appends `lease.expired`.
- `TestRenewRefusesNonHolderWithoutWrite` (homestate): a different label is refused, and nothing changes. RED on the unchanged tree: the expiry branch runs before the holder check, so the caller reaps the lease.
- `TestFactoryRenewMatchesMCP` (CLI): the same card and caller through the verb and through the MCP handler give equal results. RED: no tool exists.
- Catalogue pins: the expected count and the write-capable set change in the same edit. RED: the catalogue test fails until the tool is declared.

### M3: One-shot recovery and refusal text (Priority High; REQ-FRR-006, -007, -014)

Files:
1. `internal/homestate/card_transition.go` `transitionTx` (`:309-316`): when the caller is the holder of an expired lease, collect the lease, re-lease the card to that holder, and apply the requested edge in the same transaction. If the requested edge is refused, the whole step is refused and the expired row is unchanged. For any other caller, refuse with no write (REQ-FRR-014). The refusal text names the recovery verb (REQ-FRR-007); the exact wording is Decisions owed, D6.
2. An expiry that returns the card to blocked (a merging card, `applyLeaseExpiry` `:416-420`) is not recovered by this step. It stays on the human unblock path.
3. `internal/cli/factory_card.go` `next --card` refusal (`:1125-1128`): for the holder's own expired row, collect and re-lease instead of refusing as in flight. The verb keeps its parent-checkout rule (`:193-199`); defect 26 is out of scope.

RED first:
- `TestStageRecoversOwnExpiredLease`: the holder's expired lease at run, then `stage <card> run`. One invocation succeeds; the events read the expiry and then the transition; the card ends leased to the holder at run. RED: the transition returns `ErrLeaseExpired` (`:315`).
- `TestNextRecoversOwnExpiredLeaseFromParentCheckout`: the same through the next verb from the parent checkout. RED: refused as in flight (`factory_card.go:1125-1128`).
- `TestExpiredRefusalNamesRecoveryVerb`: a non-holder request on an expired card returns text that names the recovery verb. RED: the text at `:315` names no verb.
- `TestRecoveryAllOrNothing`: an illegal requested edge leaves the expired row unchanged. RED on the unchanged tree: the expired row is reaped (version bump, `lease.expired`) before the refusal, so the unchanged-row assertion fails.

### M4: Transition edges for sync-audit re-entry and holder re-stamp (Priority Medium; REQ-FRR-008, -009, -010; AC-FRR-016)

Blocked on Decisions owed, D3, for the R3 guard. The R4 rows can proceed independently.

Files:
1. `internal/homestate/card_transition.go`: the guard enum (`:100-121`) gains two values, one for the verdict-gated sync-audit → run edge and one for the holder re-stamp. The table (`:133-181`) gains the sync-audit → run row and the plan-audit and sync-audit same-stage rows. `holderGuards` (`:432-435`) lists the holder-only guards. The guard cases (`:515-747`) gain the new handlers. The `@MX:ANCHOR` reason at `:130` states the new count.
2. A re-stamp reads the verdict against the new hash. The current verdict guard reads against the stored hash (`card_transition.go:564`), which is the trap of lens-4 defect 7. The new guard must not inherit that read.
3. `internal/homestate/fr_transition_test.go:124`: the pinned count moves to the measured baseline plus the added rows (AC-FRR-016).

RED first:
- `TestSyncAuditFailOwnScopeReturnsToRun` (AC-FRR-008): RED: no row exists, so the request is an `ErrIllegalTransition`.
- `TestSyncAuditRefusesRunWithoutOwnScopeFail` (AC-FRR-009): on the unchanged tree this case is refused only because no edge exists. The test therefore pins the own-scope reason string, which is red until the guard exists. A bare refusal check would be green for the wrong reason.
- `TestHolderRestampsEvidenceAtPlanAudit` and `TestHolderRestampsEvidenceAtSyncAudit` (AC-FRR-010): RED: no same-stage row exists, so the request is an `ErrIllegalTransition`.
- AC-FRR-016: after the rows are added and before the pin moves, the count test fails. Record the output.

### M5: Interim guard on the empty-stage to run edge (Priority Medium; REQ-FRR-012, -013)

Files:
1. `internal/homestate/card_transition.go`: the stage guard (`guardStageResume`, `:546-549`) passes the card's SPEC id and an admission check to `stageAdmits` (`:751-756`). For an empty stage to run with a non-empty SPEC id, the guard requires an admitted plan verdict, read with the same admission the plan-to-kickoff edge applies (`admitCardVerdict`, called at `:569`), or a recorded operator decision (carrier: Decisions owed, D2).
2. Cards with an empty SPEC id, and the sync-audit → run row, pass through unchanged.

RED first:
- `TestEmptyStageToRunRefusedWithSpecID` (AC-FRR-012): RED: `stageAdmits` admits the empty-stage run unconditionally (`:752-753`).
- `TestEmptyStageToRunAdmittedWithPlanVerdict` (AC-FRR-012, allow side): green before the change, so it is adopted only together with the refuse side.
- `TestEmptyStageToRunUnchangedWithoutSpecID` (AC-FRR-013): a regression guard, green before and after.

### M6: Watchdog renews first (Priority Medium; REQ-FRR-005)

Files:
1. `internal/template/templates/.claude/skills/moai-lane-watchdog/SKILL.md` (edit first): `allowed-tools` (line 17) gains `Bash(moai factory renew:*)`; §1 gains a first step that renews the lane's own card lease before the three-channel measurement; Boundary 3 names the renewal heartbeat as the one write added.
2. `make build` to refresh the embedded template, then mirror the rendered file to `.claude/skills/moai-lane-watchdog/SKILL.md`.
3. `.claude/rules/moai/workflow/auto-semantics.md` §3 and §5.1, and the template mirror: the renewal step, and the cadence statement once Decisions owed, D4, is answered.

RED first:
- A template test asserts that the watchdog procedure names the renew verb before the measurement step, and that `allowed-tools` lists the renew command. RED: the current SKILL.md has neither (`allowed-tools` at line 17; §1 opens with measurement).

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
- Citing a count from a report instead of measuring it (§C item 2).
- Writing the R6 operator branch as an always-true allow. It must name the recorded decision it accepts.

## §H Cross-references

- `.claude/rules/moai/workflow/auto-semantics.md` §3, §5.1, and §9: the watchdog channels, the cadence, and the gate inventory.
- `.claude/skills/moai-lane-watchdog/SKILL.md`: the watchdog procedure and its boundaries.
- `.claude/rules/moai/development/verification-completeness.md` §1.1, §2, and §2.1: the observed-failure rule, the two-cell pairing, and the wrong-reason red.
- `.claude/rules/moai/core/verification-claim-integrity.md` §2.2: build provenance for tool measurements.
- `.claude/rules/moai/workflow/factory-dispatch-mechanics.md`: the lease discussion updated in M7.

## Decisions owed

These are judgment calls the plan does not settle. Each lists its options without a preferred answer; the lane decides.

- D1. Brief numbering (§B corrections): confirm that the spec uses the lens-4 numbering.
- D2 (lane judgment: option (c), R6 limited to the plan-verdict branch). R6 operator carrier. Lens-2 D2 says no operator edge exists. Options: (a) add an operator-decided edge at T4b in this batch; (b) keep the operator branch inert until a carrier lands in a later SPEC; (c) limit R6 to the plan-verdict branch only.
- D3 (lane judgment, 2026-10-10 00:4x Z): the R3 own-scope discriminator is the leader's recorded decision, so REQ-FRR-008 needs no verdict field and stays outside t1618. Options kept for the record: (a) defer REQ-FRR-008 until the verdict format carries a field (t1618); (b) add a field to the verdict format in this batch, which crosses t1618's boundary; (c) treat any FAIL with a blocking count of one or more as own-scope, which admits defects outside the card's scope.
- D4. Watchdog cadence against the 15-minute lease. Options: (a) keep the 20-minute cadence and renew at each fire and at the lane's own actions; (b) move the cadence to at most half the lease duration (a cost trade-off in auto-semantics §5.1); (c) raise the default lease duration above the cadence, which revises REQ-FRR-003.
- D5. R5 path source and timing. Source options: the path the lane supplies, checked against the path recorded at picked, or the path recorded at picked alone. Timing options: lease acquisition (T3, the proposal in M1) or assignment (T2). Decision record 8ccf says "a record field the lease did not stamp".
- D6. Recovery verb in the refusal text: `factory stage` (reachable from a worktree), `factory next --card` (parent checkout only), or both with their reachability stated.
- D7. Renewal of an expired own lease: refused with no write (as REQ-FRR-004 states) or recovered as REQ-FRR-006 does. The spec currently refuses.
- D8. Residual empty-SPEC route: accept it as stated in spec.md Residual risk, or move the narrow part of defect 5 (setting the SPEC id while leased) into scope.
- D9. Lens-4 defects 10, 13, 14, 19, 20, 22, 24, 25, and 26: confirm that they stay unassigned, or assign them.
- D10. `decision-index.md` (owed under `decision_gate: on`) and the `progress.md` §E skeleton: confirm that they are written in a later batch, since this batch writes exactly three files.
- D11. The `module` field is `internal/homestate`, as the brief specifies. The write set also touches `internal/cli`, `internal/config`, `internal/mcp`, and the templates. Confirm.
- D12. Priority P0, because the batch carries lens-4 defect 1 and lens-2 D1. Confirm.
- D13. The configuration key name `workflow.factory_lease.duration`. Confirm or rename.
