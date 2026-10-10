---
id: SPEC-FACTORY-LEASE-RENEW-001
version: "0.1.0"
updated: 2026-10-10
---

# Implementation plan: SPEC-FACTORY-LEASE-RENEW-001

## §A Context

Card t1663 (Class C, Tier M, lane-17) implements the lease renew verb, the lease configuration keys, and the watchdog renewal. The draft SPEC-FACTORY-RECORD-REPAIR-001 carried these as its items 1 and 3 before the card split. The draft's plan audit iteration 4 (`.moai/reports/t1621/plan-audit-iter4.md`) found D1 (the renewal-period bound), D3 (the operator verdict wording), and D4 (acceptance wording against the Definition of Done). Section B maps each finding to the revision that builds it in.

The audited code is the code of this tree. The change set since the draft base `2aab5f797` is the five draft SPEC files only (`git diff --name-only 2aab5f797 HEAD`), so the code citations below were read at HEAD.

Current code:

- The card lease duration is a compiled constant, `DefaultFactoryLeaseDuration` (`internal/config/defaults.go:65`). It reaches the lease writes through the package variable `FactoryLeaseDuration` (`internal/homestate/card_record.go:93`). The constant's comment says that no config key reads it.
- `FactoryDB.RenewLease` (`internal/homestate/card_transition.go:441`) renews the holder's lease. Its expired branch reaps the lease (`applyLeaseExpiry`) before it refuses (`:451-456`). Its only non-test caller is `factoryStageCard` (`internal/cli/factory_card.go:1737`), so a lease is extended only by a successful stage transition.
- No factory verb renews a lease. The factory verbs are `next` (`factory_card.go:1529`), `stage` (`:1661`), `complete` (`:1754`), `status` (`:2617`), and `decide` (`:2757`). The MCP factory tools are declared in `internal/mcp/catalog.go` and handled in `internal/cli/mcp_factory_card.go`.
- The queue claim uses the compiled 15-minute constant directly (`internal/factory/backlog_store.go:1244`) and is independent of the card lease.
- The standing watchdog wake is `CronCreate` with `7,27,47 * * * *` (`.claude/rules/moai/workflow/auto-semantics.md:108`). The watchdog skill (`.claude/skills/moai-lane-watchdog/SKILL.md`) opens with measurement (§1, line 51) and lists no renew command in `allowed-tools` (line 17).
- `workflow.yaml` carries no lease key. Its shipped keys are inventoried in `internal/config/testdata/shipped_key_inventory.yaml` and checked by `internal/config/shipped_key_reader_test.go`. The loader symmetry test is `internal/config/audit_struct_yaml_symmetry_test.go`.
- `OpenFactory(projectRoot)` (`internal/homestate/factory.go:167`) is the factory open entry point, and the refusals of REQ-FLR-006 and REQ-FLR-007 sit on it.

### Assumptions

Each assumption carries a one-line rationale. An assumption that an operator could reasonably decide differently is also a row in decision-index.md, named in brackets.

- A1. Tier M: spec.md, plan.md, and acceptance.md, plus decision-index.md (the decision gate is on) and progress.md. The change touches more than five files, and 12 requirements and 13 criteria fit the 16 / 16 ceilings. No design or research artifact is needed.
- A2. The card lease duration key keeps the draft's name, `workflow.factory_lease.duration`, so one setting has one name across the two SPECs. [Q1 asks confirmation.]
- A3. The renewal period key is `workflow.factory_lease.renew_interval`, named in the same family as A2. [Q1.]
- A4. The renewal period default is 5 minutes. It lies inside the bound at the 15-minute default (7 minutes 30 seconds) and leaves a 10-minute late-pass margin. [Q1.]
- A5. A key outside its bound refuses to open the factory, because a silent clamp would hide a misconfiguration. [Q2 states the alternative.]
- A6. An expired renewal is refused with no write, the form the draft's REQ-FRR-004 states. [Q3 states the alternative.]
- A7. The standing wake follows the renewal period. A 20-minute wake cannot meet any bound at or below 7 minutes 30 seconds, so the 20-minute literal is replaced rather than kept.
- A8. The renew verb is lane-only, as the stage verb is. The caller is the lane label that the factory worker environment carries (`config.EnvMoaiFactoryWorker`, read by the MCP factory handlers at `internal/cli/mcp_factory_card.go:119-123`).
- A9. The transferred items (draft REQ-FRR-006 to REQ-FRR-008, and the non-holder expiry ordering) are authored nowhere in this SPEC, as the card body directs.
- A10. The draft SPEC-FACTORY-RECORD-REPAIR-001 is not modified. The overlap is decision-index Q5, a DEFAULT-APPLIED row that leaves the draft as it is.
- A11. Priority P0 is inherited from the draft batch, whose brief marks lens-4 defects 1 and 27 as P0.
- A12. The module field stays `internal/homestate`, the value the draft settled for its module.
- A13. The MCP tool is named `factory_renew`, following the existing `factory_next` and `factory_stage` names. This is a naming convention and carries no decision row.

## §B Known issues and the audit mapping

### B. Code and configuration facts the implementation must respect

- B1. `RenewLease` reaps before it checks the holder. The expired branch must stop writing (REQ-FLR-004). `applyLeaseExpiry` stays for the transferred stage path.
- B2. `FactoryLeaseDuration` is a package variable. The configured value must reach the lease writes through the factory handle that `OpenFactory` returns, and the process must not reassign the variable at start. The lease writes at `card_transition.go:463`, `:542`, and `:601` read the variable today.
- B3. A configuration key needs four registrations: the template `workflow.yaml` and the rendered copy (compared by key, QG-4); the struct field and YAML tag in `internal/config/types.go`; the default in `internal/config/defaults.go`; and an entry with a class in `internal/config/testdata/shipped_key_inventory.yaml`. The symmetry test's case list in `audit_struct_yaml_symmetry_test.go` gains the struct when one is added.
- B4. The MCP catalogue is pinned by `internal/mcp/catalog_test.go` (the catalogue size, the write-capable tool set, the count, and the tool-name list). `factory_renew` writes a lease row, so it joins the write-capable set. The MCP catalogue rule (`.claude/rules/moai/core/moai-mcp-tools-catalogue.md`) and its template mirror are updated in the same change.
- B5. The standing-wake paragraph at `auto-semantics.md:108` belongs to a rule that is path-scoped to the watchdog skill, and it has a template mirror. Its sentence that justifies the 20-minute cadence (keeping each fire longer than the N-minute stall window) is rewritten. The stall verdict itself (N = 15 minutes; a previous observation at least N old) does not change.
- B6. The minute step must have the gap property: no two consecutive passes further apart than the interval. A start at minute 0 (`*/N`) has it for every N from 1 to 59, because the last fire of an hour is at least 60 minus N. An off-minute start (`o/N`, as the old `7,27,47` literal is) has it only while the gap across the hour boundary stays within N. For N = 59 with o = 2 the gap is 60 minutes. The armed expression is checked against this property for the value in force.

### Audit mapping (plan audit iteration 4 of SPEC-FACTORY-RECORD-REPAIR-001)

- **D1 (renewal-period bound).** The bound is stated in REQ-FLR-007 and enforced when the factory opens. Behavioural checks are AC-FLR-007 (refusal and boundary) and AC-FLR-012 (late-pass margin and expired refusal). The cadence that would have to meet the bound is changed in this SPEC (REQ-FLR-012, AC-FLR-013), not left open. spec.md Residual risk, decision-index Q1, and progress.md §E.1 state the same period: a working value of 5 minutes, a bound of 7 minutes 30 seconds at the default duration, and a whole-minute maximum of 7 minutes.
- **D3 (operator verdict wording).** One sentence is used in decision-index.md (the "How to read" section) and in acceptance.md (the Definition of Done): "The `Operator verdict:` line is left empty at authoring; the operator writes it." No Operator verdict value is copied from another SPEC or from a board record. The decision index carries no Status lines, so no status value is written outside the enum. Only Q5 takes a DEFAULT-APPLIED line, and it is the one implementation-level row with a ranked Default.
- **D4 (acceptance wording against the Definition of Done).** acceptance.md uses one vocabulary: at plan time there is a predicted RED (not observed); in the run phase there is an observed RED before the GREEN output; and a criterion is recorded as PASS only from its GREEN output. The plan-time sentence "no criterion is recorded as a pass" and the Definition of Done's "no criterion is recorded as PASS before its GREEN output" agree. The mutant rule covers the cases that are green on the unchanged tree (the default case of AC-FLR-005, the acceptance case of AC-FLR-007, and AC-FLR-008), so that no RED is claimed where none was observed.

## §C Pre-flight (for the implementing lane, before any edit)

1. Re-read the branch and commit state, because the staleness rule applies: `git rev-parse --short HEAD` and `git branch --show-current`. The branch is `WT-factory-lease-renew`. A difference from what the turn assumed stops the work.
2. Confirm the open path and the test seams before writing any test: `grep -rn "func OpenFactory" internal/homestate --include="*.go"` and `grep -rn "factoryCardNow" internal/cli --include="*_test.go"`.
3. Measure the baseline with the verification convention. `go test ./internal/homestate/ -run '^TestRenewLeaseRefusalsAndExpiry$' -v -count=1` must show `=== RUN` and `--- PASS` before M2 starts, and it must stay GREEN after M2.
4. Record the HEAD SHA beside every verbatim output, so that each claim is attributed to this run and this tree.

## §D Constraints

- The plan phase writes only the five artifacts in `.moai/specs/SPEC-FACTORY-LEASE-RENEW-001/`. No Go, template, or rule file changes at plan time.
- Run-phase edits follow Template-First: the template is edited, `make build` runs, and the local mirror follows (QG-4). The repository-local overrides in the rendered `workflow.yaml` survive the sync (decision-index Q4).
- The queue claim stays at 15 minutes (REQ-FLR-008). `internal/auditverdict/` is not changed.
- A test names no Go symbol that does not exist yet. A test that does is TOOL_FAILURE and is never recorded as a predicted RED (acceptance.md, surface rule).
- The renew verb is lane-only, as the stage verb is (A8).
- Plans carry no time estimates. Order is by priority label and phase.
- Every commit names card t1663 (QG-5). The lane lands locally and does not push.

## §E Self-verification (plan level)

Requirement-to-criterion traceability. Every requirement maps to at least one criterion, and every criterion maps to at least one requirement.

| Requirement | Criteria |
|---|---|
| REQ-FLR-001 | AC-FLR-001, AC-FLR-012 |
| REQ-FLR-002 | AC-FLR-002 |
| REQ-FLR-003 | AC-FLR-003 |
| REQ-FLR-004 | AC-FLR-004, AC-FLR-012 |
| REQ-FLR-005 | AC-FLR-005 |
| REQ-FLR-006 | AC-FLR-006 |
| REQ-FLR-007 | AC-FLR-007, AC-FLR-012 |
| REQ-FLR-008 | AC-FLR-008 |
| REQ-FLR-009 | AC-FLR-009, AC-FLR-012 |
| REQ-FLR-010 | AC-FLR-010 |
| REQ-FLR-011 | AC-FLR-011 |
| REQ-FLR-012 | AC-FLR-012 (margin consequence), AC-FLR-013 (armed step) |

Criterion-to-requirement: AC-FLR-001 to AC-FLR-011 each map to the one requirement of the same number. AC-FLR-012 maps to REQ-FLR-001, -004, -007, -009, and -012. AC-FLR-013 maps to REQ-FLR-012.

## §F Milestones

Milestones are ordered by decision reversibility. The configuration decisions come first (Q1 and Q2), then the user-facing renew verb (Q3), then the watchdog doctrine that depends on both, then the margin check, and the mechanical mirrors come last.

### M1: Lease configuration keys and open-time refusal (Priority High; REQ-FLR-005, -006, -007, -008, -009)

Files (run phase):

1. `internal/template/templates/.moai/config/sections/workflow.yaml` gains `factory_lease.duration` (15m) and `factory_lease.renew_interval` (5m), with a comment that names both bounds. The rendered `.moai/config/sections/workflow.yaml` gains the same keys, and its repository-local overrides stay intact (decision-index Q4).
2. `internal/config/types.go` (struct fields and YAML tags), `internal/config/defaults.go` (the defaults and an updated comment on `DefaultFactoryLeaseDuration`), and the loader wiring. `audit_struct_yaml_symmetry_test.go` gains the case.
3. `internal/config/testdata/shipped_key_inventory.yaml` gains an entry for each key, with its class.
4. `internal/homestate`: `OpenFactory` reads both keys, refuses an invalid value with the named key (REQ-FLR-006, REQ-FLR-007), and sets the card lease duration on the handle. The lease writes read the handle's value (B2).
5. `internal/cli` factory status (`factory_card.go:2617`) reports the effective `lease_duration` and `renew_interval` in its JSON output (REQ-FLR-009).
6. The queue claim (`backlog_store.go:1244`, `:1322`) keeps its compiled constant (REQ-FLR-008). The mutant of AC-FLR-008 is executed in this milestone.

RED first (predicted, not observed until the run phase): AC-FLR-005 (the 20-minute override), AC-FLR-006, the refusal cases of AC-FLR-007, and AC-FLR-009. Each case that is green on the unchanged tree takes a mutant as its RED: the default case of AC-FLR-005 (a mutant with a different default), the acceptance case of AC-FLR-007 (a mutant that refuses at the bound), and AC-FLR-008 (a claim that reads the card key).

Exit: AC-FLR-005, AC-FLR-006, AC-FLR-007, AC-FLR-008, and AC-FLR-009 are GREEN.

### M2: Renew verb, MCP tool, and the expired refusal with no write (Priority High; REQ-FLR-001, -002, -003, -004)

Files (run phase):

1. `internal/cli/factory_card.go`: a `renew <card>` verb in the factory command group, lane-only, beside `stage` (`:1661`). It prints the new expiry (REQ-FLR-001).
2. `internal/homestate/card_transition.go`, `RenewLease` (`:441-456`): the expired branch returns the expired-lease class without calling `applyLeaseExpiry`. The non-holder branch is unchanged (REQ-FLR-003). This change is subject to decision-index Q3.
3. `internal/mcp/catalog.go` declares `factory_renew` as write-capable. `internal/cli/mcp_factory_card.go` handles it through the same function as the verb (REQ-FLR-002). The pins in `internal/mcp/catalog_test.go` are updated (B4).
4. `.claude/rules/moai/core/moai-mcp-tools-catalogue.md` and its template mirror gain the tool in the same change.

RED first (predicted): AC-FLR-001, AC-FLR-002, AC-FLR-003, and AC-FLR-004. The verb and the tool are absent, and `RenewLease` writes on the expired path.

Exit: AC-FLR-001, AC-FLR-002, AC-FLR-003, and AC-FLR-004 are GREEN. The existing `TestRenewLeaseRefusalsAndExpiry` (`internal/homestate/f7_record_failure_test.go`) is run to confirm it stays GREEN.

### M3: Watchdog renewal and standing wake (Priority Medium; REQ-FLR-010, -011, -012)

Files (run phase):

1. `internal/template/templates/.claude/skills/moai-lane-watchdog/SKILL.md`: `allowed-tools` gains `Bash(moai factory renew:*)`. A first step renews the lane's card lease before the measurement in §1 (REQ-FLR-010). The step states that a refused renewal is recorded in the watchdog snapshot and that measurement still runs (REQ-FLR-011). Then `make build`, and the local mirror `.claude/skills/moai-lane-watchdog/SKILL.md` follows.
2. `internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md`, §5.1 "Arm": the arm is a minute step of `workflow.factory_lease.renew_interval` with the gap property (B6). The sentence that justified the 20-minute cadence is rewritten (B5). The local rule `.claude/rules/moai/workflow/auto-semantics.md` follows.
3. `.claude/rules/moai/workflow/factory-dispatch-mechanics.md` and its template mirror `internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md`: one sentence on the renew verb in the lease discussion.

RED first (predicted): AC-FLR-010 (the procedure opens with measurement and lists no renew command), AC-FLR-011 (no refusal branch in the procedure), and AC-FLR-013 (the literal `7,27,47 * * * *` is present).

Exit: AC-FLR-010, AC-FLR-011, and AC-FLR-013 are GREEN.

### M4: Late-pass margin and expired refusal (Priority Medium; REQ-FLR-001, -004, -007, -009, -012)

Files (run phase):

1. An injected-clock test in `internal/cli` (`TestRenewCadenceKeepsLeaseLive`), using the factory test clock (`factoryCardNow`). It runs for the defaults (15 minutes and 5 minutes) and for the boundary pair (10 minutes and 5 minutes), across three durations. On-schedule renewals succeed. A renewal delayed by D minus I minus one second succeeds. A renewal requested at the expiry instant and at one second after it is refused, with no write.

RED first (predicted): the renew verb is absent, so the first on-schedule renewal fails.

Exit: AC-FLR-012 is GREEN.

### M5: Mirrors, catalogue, and quality gates (Priority Low; mechanical; no behaviour change)

1. `make build`, then the key-scoped comparison of `workflow.yaml` and the byte comparison of the skill and rule mirrors (QG-4), with the overrides preserved.
2. QG-1 to QG-3 (format, lint, race, coverage), each with its verbatim output and the HEAD SHA beside it.
3. QG-5: each commit message names card t1663.

Exit: QG-1 to QG-5 pass, or each shortfall is recorded as a Gap.

## §G Anti-patterns to avoid

- Clamping an out-of-bound value silently, or writing a default into a shipped configuration without its decision row (Q2 is open).
- Restating the 20-minute literal anywhere: auto-semantics §5.1, the watchdog skill, or the plan.
- Filling an `Operator verdict:` line with a value copied from the draft SPEC or from a board record. This plan writes only the DEFAULT-APPLIED line of Q5.
- Changing the queue claim to read the card key, or changing the audit-verdict parser.
- Naming in a test a Go symbol that does not exist yet. That is TOOL_FAILURE, not a predicted RED.
- Recording a mutant's output as the observed RED of the real code, or recording a green-before-change case as a pass.
- Running `git -C <worktree>`, `cd && git`, or a bare `git stash` in a worktree session.
- Reassigning `FactoryLeaseDuration` at process start instead of reading the handle (B2).
- Running `go test ./...` locally. Verification is lane-local and scoped to the change, and CI runs the full suite.

## §H Cross-references

- `spec.md` (requirements, exclusions, residual risk), `acceptance.md` (criteria and Definition of Done), `decision-index.md` (Q1 to Q5), and `progress.md` (§E.1).
- `.moai/specs/SPEC-FACTORY-RECORD-REPAIR-001/`: the draft that this card split from. It is not modified here.
- `.claude/rules/moai/workflow/auto-semantics.md` §4 (the stall window) and §5.1 (the standing wake).
- `.claude/rules/moai/workflow/cache-aware-execution.md` directive 9 (session length and the cache window).
- `.claude/rules/moai/development/verification-completeness.md` §1.2 (the three-part check), §1.3 (continued firing), and §2 (the two-cell pair and the mutant probe).
- `.claude/rules/moai/workflow/tdd-result-contract.md` (RED and GREEN semantics).
- `.claude/rules/moai/core/verification-claim-integrity.md` §2 (attribution of each claim to a run and a tree).

## Decisions owed

These map to decision-index.md. Q1 (the values and names of the two keys), Q2 (refusal or clamping), and Q3 (the expired-renewal form) are open FOUNDER rows, and the autonomous Kickoff stays blocked until their `Operator verdict:` lines are written. Q4 is POLICY-COVERED. Q5 is DEFAULT-APPLIED at plan close.
