---
id: SPEC-FACTORY-LEASE-RENEW-001
version: "0.1.1"
updated: 2026-10-10
---

# Implementation plan: SPEC-FACTORY-LEASE-RENEW-001

## §A Context

Card t1663 (Class C, Tier M, lane-17) implements the lease renew verb, the lease configuration keys, and the watchdog renewal and stall window. The draft SPEC-FACTORY-RECORD-REPAIR-001 carried these as its items 1 and 3 before the card split. The draft's plan audit iteration 4 (`.moai/reports/t1621/plan-audit-iter4.md`) found D1 (the renewal-period bound), D3 (the operator verdict wording), and D4 (acceptance wording against the Definition of Done). The first audit of this SPEC (`.moai/reports/t1663/plan-audit.md`, iteration 1, verdict FAIL) found FLR-B1 to FLR-B8 and the lane's point (d); §B maps each finding to the revision that answers it.

The audited code is the code of this tree. The change set since the draft base `2aab5f797` is the five draft SPEC files only (`git diff --name-only 2aab5f797 HEAD`), so the code citations below were read at HEAD.

Current code:

- The card lease duration is a compiled constant, `DefaultFactoryLeaseDuration` (`internal/config/defaults.go:65`). It reaches the lease writes through the package variable `FactoryLeaseDuration` (`internal/homestate/card_record.go:93`). The writes read the variable at `card_transition.go:463` (renewal), `:542`, and `:601` (transition). The constant's comment says that no config key reads it.
- `FactoryDB.RenewLease` (`internal/homestate/card_transition.go:441`) renews the holder's lease. Its expired branch reaps the lease (`applyLeaseExpiry`) before it refuses (`:452-456`). Its only non-test caller is `factoryStageCard` (`internal/cli/factory_card.go:1737`), which renews right after a successful transition, so a lease is extended only by a stage transition.
- No factory verb renews a lease. The factory verbs are `next` (`factory_card.go:1529`), `stage` (`:1661`), `complete` (`:1754`), `status` (`:2617`, with `--json` at `:2675`), and `decide` (`:2757`). The factory command group is registered in `internal/cli/factory_handoff_recover.go:20`. The MCP factory tools are declared in `internal/mcp/catalog.go` and handled in `internal/cli/mcp_factory_card.go`.
- The queue claim uses the compiled 15-minute constant directly (`internal/factory/backlog_store.go:1244`) and is independent of the card lease.
- The standing watchdog wake is `CronCreate` with `7,27,47 * * * *` (`.claude/rules/moai/workflow/auto-semantics.md:108`). The watchdog skill (`.claude/skills/moai-lane-watchdog/SKILL.md`) opens with measurement (§1, line 51) and lists no renew command in `allowed-tools` (line 17).
- The watchdog snapshot is written by the lane following a markdown procedure. No Go code writes or reads `observed_at`, `window_state`, or `evidence_mtime_max`: a search of `internal` and `.claude` finds those fields only in the rule and skill text.
- The lane startup and `/clear` instruction is built by `internal/hook/session_start_factory_i18n.go` in four locales. `factoryLaneRuleForSource` (`internal/hook/session_start_factory.go:95`) takes no configuration argument; it reads the backend and dispatch values from the environment and selects one of the four locale strings. The same cadence is also restated in `.claude/rules/moai/workflow/factory-dispatch.md:138` (local and template copies), which the first revision did not list.
- `workflow.yaml` carries no lease key. Its shipped keys are inventoried in `internal/config/testdata/shipped_key_inventory.yaml` and checked by `internal/config/shipped_key_reader_test.go`. The loader symmetry test is `internal/config/audit_struct_yaml_symmetry_test.go`.
- `OpenFactory(projectRoot)` (`internal/homestate/factory.go:167`) is the factory open entry point, and the refusals of REQ-FLR-006 and REQ-FLR-007 sit on it.

### Assumptions

Each assumption carries a one-line rationale. An assumption that an operator could reasonably decide differently is also a row in decision-index.md, named in brackets.

- A1. Tier M: spec.md, plan.md, and acceptance.md, plus decision-index.md (the decision gate is on) and progress.md. The change touches more than five files, and 13 requirements and 15 criteria fit the 16 / 16 ceilings. No design or research artifact is needed.
- A2. The card lease duration key keeps the draft's name, `workflow.factory_lease.duration`, so one setting has one name across the two SPECs. [Q1 ruled: d-20261010T095818Z-c501.]
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
- A14. The stall window is counted from the last observed channel change (REQ-FLR-013), because auto-semantics §3 states the window that way and the previous-snapshot comparison cannot fire below N. [Q6 ruled: d-20261010T095731Z-5ddc.]
- A15. The armed wake starts at minute 0 (`*/N`), because that form has the gap property for every N from 1 to 59 with no per-value computation; an off-minute start is possible only with a per-value gap check. [Q1 sub-question d.]
- A16. The snapshot gains exactly two additive fields, `renew_result` and `last_change_at`, the fewest that REQ-FLR-011 and REQ-FLR-013 need; the audit allows one additive field as a stated exception, and the stall window needs the second.
- A17. The renew verb gains a JSON output whose `result` field carries the refusal class that REQ-FLR-002 already requires the MCP tool to return, so the watchdog copies one machine-readable value and does not parse prose.
- A18. The hook renders the effective renewal period from the configuration; the mechanism (a read at hook time, or a value the launcher stamps the way it stamps the dispatch mode) is the run phase's choice, and AC-FLR-014 is written against the rendered output so that either mechanism passes it.

## §B Known issues and the audit mapping

### B. Code and configuration facts the implementation must respect

- B1. `RenewLease` reaps before it checks the holder. The expired branch must stop writing (REQ-FLR-004). `applyLeaseExpiry` stays for the transferred stage path. The stage path calls `RenewLease` only after a transition has extended the lease, so the expired branch is not reached there.
- B2. `FactoryLeaseDuration` is a package variable. The configured value must reach the lease writes through the factory handle that `OpenFactory` returns, and the process must not reassign the variable at start. The lease writes at `card_transition.go:463`, `:542`, and `:601` read the variable today.
- B3. The typed field, its default, and the template copy belong to this card, and so does the anti-rot triage entry for each new key in internal/config/testdata/shipped_key_inventory.yaml (class W, evidence reader). The config-registry entry belongs to t1636 (decision d-20261010T100853Z-98a1). The symmetry test's case list in `audit_struct_yaml_symmetry_test.go` gains the struct when one is added.
- B4. The MCP catalogue is pinned by `internal/mcp/catalog_test.go` (the catalogue size, the write-capable tool set, the count, and the tool-name list). `factory_renew` writes a lease row, so it joins the write-capable set. The MCP catalogue rule (`.claude/rules/moai/core/moai-mcp-tools-catalogue.md`) and its template mirror are updated in the same change.
- B5. The standing-wake paragraph at `auto-semantics.md` §5.1 belongs to a rule that is path-scoped to the watchdog skill, and it has a template mirror. Its arm sentence and its cost sentence both justify the 20-minute cadence against the stall window, and both are rewritten. The stall rule of §3 keeps its three channels and N = 15 minutes; §4 changes only its reference point (REQ-FLR-013) and gains the two snapshot fields. `factory-dispatch.md:138` ("every 20 minutes") is rewritten in the same change, in both copies; M3 names it.
- B6. The minute step must have the gap property: no two consecutive scheduled wakes further apart than the period. The form `*/N` (start 0) has it for every N from 1 to 59, because the last fire of an hour is at least 60 minus N. An off-minute start `o/N` has it only while the gap across the hour boundary stays within N: for N = 59 with o = 2 the gap is 60 minutes, and for N = 7 with o = 6 it is 11. AC-FLR-013 checks the form and computes the largest gap for every legal N.
- B7. The lane rule builder takes no configuration (`session_start_factory.go:95`). The four locale strings (`session_start_factory_i18n.go:132`, `:188`, `:242`, `:295`) carry the literal inside a constant string. The owned-card rule already takes a formatted argument (`laneOwnedCardRule`, `card id %[1]s`), which is the pattern for rendering the arm expression from the effective value.
- B8. The renew verb's JSON `result` is the only machine-readable refusal class, and the watchdog copies it into `renew_result`. No Go code writes the snapshot file, so AC-FLR-011 checks the class and the copy step, not a written file.
- B9. No Go code computes the stall verdict. AC-FLR-015 carries a model of the rule as §4 will state it, and a text pin on §4. Its limit is declared in spec.md Residual risk.

### Audit mapping (plan audit iteration 1 of SPEC-FACTORY-LEASE-RENEW-001, `.moai/reports/t1663/plan-audit.md`)

- **FLR-B1 (stall verdict at the new cadence).** REQ-FLR-013 counts the window from the last channel change; the snapshot gains `last_change_at`; auto-semantics §3 and §4 and the watchdog skill §1 state it (M3); AC-FLR-015 runs the model over 120 unchanged minutes at 5 and 7 minutes and checks equality with the previous-snapshot rule at 20 minutes. The choice is ruled at decision-index Q6 (d-20261010T095731Z-5ddc).
- **FLR-B2 (the 20-minute literal in the hook).** The four locale strings, `lane_recheck_rule_test.go:22`, `lane_recheck_doctrine_test.go:61`, and `factory-dispatch.md:138` are in M3. AC-FLR-014 checks the rendered rule for `*/7` and `*/5` and for the absence of the literal in four locales.
- **FLR-B3 (start minute and bound).** REQ-FLR-012 fixes the form at minute 0 with the gap property. AC-FLR-013 checks the form and computes the largest gap for every N from 1 to 59. The gap depends on N alone, so the legal N values cover every legal (duration, interval) pair; the duration only decides which N the open accepts.
- **FLR-B4 (REQ-FLR-005 clauses).** AC-FLR-005 has three subtests, acquisition, transition extension, and renewal, each with a 20-minute value.
- **FLR-B5 (M4 predicted RED).** The M4 RED is the output of three named mutants executed against the M2 code (§F, M4). The predicted "verb not registered" RED is withdrawn, because that cause belongs to M2.
- **FLR-B6 (snapshot field).** `renew_result` is named (REQ-FLR-011), and the format change is declared (spec.md, the snapshot paragraph after REQ-FLR-013; A16). AC-FLR-011 checks the verb's JSON class in Go and the procedure's copy step as text. The written snapshot is not observed, because no code writes it (B8). This is the one partial answer, recorded as such in the report of this revision.
- **FLR-B7 (`TestRenewLeaseRefusalsAndExpiry`).** M2 rewrites the reaped-state assertion (`f7_record_failure_test.go:36-42`) under REQ-FLR-004; the other assertions stay GREEN. The rewrite is made under Q3 option (a), ruled at decision-index Q3 (d-20261010T095731Z-5ddc).
- **FLR-B8 (both keys set).** AC-FLR-009 has a `configured` subtest with `20m` and `4m`.
- **Point (d).** AC-FLR-010 stays as a regression guard with its declared limit. AC-FLR-011 is strengthened (the verb's JSON class; the procedure's copy step and same-pass measurement). AC-FLR-013 is strengthened (the armed form and the computed gap).
- **FLR-O1, FLR-O2, FLR-O3.** AC-FLR-006 gains a negative-duration subtest; Q5 names the cadence contradiction; M4 names the third pair (20 minutes, 4 minutes).

### Audit mapping (plan audit iteration 4 of SPEC-FACTORY-RECORD-REPAIR-001)

- **D1 (renewal-period bound).** The bound is stated in REQ-FLR-007 and enforced when the factory opens. Behavioural checks are AC-FLR-007 (refusal and boundary) and AC-FLR-012 (late-pass margin and expired refusal). The cadence that would have to meet the bound is changed in this SPEC (REQ-FLR-012, AC-FLR-013, AC-FLR-014), not left open. spec.md Residual risk, decision-index Q1, and progress.md §E.1 state the same period: a working value of 5 minutes, a bound of 7 minutes 30 seconds at the default duration, and a whole-minute maximum of 7 minutes.
- **D3 (operator verdict wording).** One sentence is used in decision-index.md (the "How to read" section) and in acceptance.md (the Definition of Done): "The `Operator verdict:` line is left empty at authoring; the operator writes it." A delegated verdict copied from a board record is recorded with the board decision id it rests on. The decision index carries no Status lines, so no status value is written outside the enum. Only Q5 takes a DEFAULT-APPLIED line, and it is the one implementation-level row with a ranked Default.
- **D4 (acceptance wording against the Definition of Done).** acceptance.md uses one vocabulary: at plan time there is a predicted RED or a mutant RED (not observed); in the run phase there is an observed RED before the GREEN output; and a criterion is recorded as PASS only from its GREEN output. The plan-time sentence "no criterion is recorded as a pass" and the Definition of Done's "no criterion is recorded as PASS before its GREEN output" agree.

## §C Pre-flight (for the implementing lane, before any edit)

1. Re-read the branch and commit state, because the staleness rule applies: `git rev-parse --short HEAD` and `git branch --show-current`. The branch is `WT-factory-lease-renew`. A difference from what the turn assumed stops the work.
2. Confirm the open path and the test seams before writing any test: `grep -rn "func OpenFactory" internal/homestate --include="*.go"` and `grep -rn "factoryCardNow" internal/cli --include="*_test.go"`.
3. Measure the baseline with the verification convention. `go test ./internal/homestate/ -run '^TestRenewLeaseRefusalsAndExpiry$' -v -count=1` must show `=== RUN` and `--- PASS` before M2 starts. After M2 it is GREEN with its reaped-state assertion rewritten under REQ-FLR-004 (M2 item 5).
4. Record the HEAD SHA beside every verbatim output, so that each claim is attributed to this run and this tree. Any tool measurement cited as evidence names the judging build's commit beside HEAD (verification-claim-integrity §2.2).

## §D Constraints

- The plan phase writes only the five artifacts in `.moai/specs/SPEC-FACTORY-LEASE-RENEW-001/`. No Go, template, or rule file changes at plan time.
- Run-phase edits follow Template-First: the template is edited, `make build` runs, and the local mirror follows (QG-4). The repository-local overrides in the rendered `workflow.yaml` survive the sync (decision-index Q4).
- The queue claim stays at 15 minutes (REQ-FLR-008). `internal/auditverdict/` is not changed.
- A test names no Go symbol that does not exist yet, apart from its own identifier. It reaches a new surface through the command tree, the MCP tool list, a public handle such as `OpenFactory`, or the rendered output of an existing rule builder. A test that names a missing symbol is TOOL_FAILURE and is never recorded as a predicted RED (acceptance.md, surface rule).
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
| REQ-FLR-005 | AC-FLR-005 (acquisition, transition extension, renewal) |
| REQ-FLR-006 | AC-FLR-006 |
| REQ-FLR-007 | AC-FLR-007, AC-FLR-012 |
| REQ-FLR-008 | AC-FLR-008 |
| REQ-FLR-009 | AC-FLR-009, AC-FLR-012 |
| REQ-FLR-010 | AC-FLR-010 |
| REQ-FLR-011 | AC-FLR-011 |
| REQ-FLR-012 | AC-FLR-013 (armed form and gap), AC-FLR-014 (injected instruction), AC-FLR-012 (margin consequence) |
| REQ-FLR-013 | AC-FLR-015 |

Criterion-to-requirement: AC-FLR-001 to AC-FLR-011 each map to the requirement of the same number (AC-FLR-001 to -004 to REQ-FLR-001 to -004). AC-FLR-012 maps to REQ-FLR-001, -004, -007, -009, and -012. AC-FLR-013 and AC-FLR-014 map to REQ-FLR-012. AC-FLR-015 maps to REQ-FLR-013.

## §F Milestones

Milestones are ordered by decision reversibility. The configuration decisions come first (Q1 and Q2), then the user-facing renew verb (Q3), then the watchdog doctrine that depends on both, then the margin check, and the mechanical mirrors come last.

### M1: Lease configuration keys and open-time refusal (Priority High; REQ-FLR-005, -006, -007, -008, -009)

Files (run phase):

1. `internal/template/templates/.moai/config/sections/workflow.yaml` gains `factory_lease.duration` (15m) and `factory_lease.renew_interval` (5m), with a comment that names both bounds. The rendered `.moai/config/sections/workflow.yaml` gains the same keys, and its repository-local overrides stay intact (decision-index Q4).
2. `internal/config/types.go` (struct fields and YAML tags), `internal/config/defaults.go` (the defaults and an updated comment on `DefaultFactoryLeaseDuration`), and the loader wiring. `audit_struct_yaml_symmetry_test.go` gains the case.
3. internal/config/testdata/shipped_key_inventory.yaml gains an entry for each new key with class W and evidence reader. This is the existing anti-rot triage table of SPEC-CONFIG-KEY-HONESTY-001, and the entries land in the same commit as items 1 and 2. The config-registry entry belongs to t1636.
4. `internal/homestate`: `OpenFactory` reads both keys, refuses an invalid value with the named key (REQ-FLR-006, REQ-FLR-007), and sets the card lease duration on the handle. The lease writes read the handle's value (B2).
5. `internal/cli` factory status (`factory_card.go:2617`) reports the effective `lease_duration` and `renew_interval` in its JSON output (REQ-FLR-009).
6. The queue claim (`backlog_store.go:1244`, `:1322`) keeps its compiled constant (REQ-FLR-008). The mutant of AC-FLR-008 is executed in this milestone.

RED first (predicted, observed in the run phase before the edits): AC-FLR-005 (the three 20-minute subtests; the default subtests take a mutant with a different default), AC-FLR-006 (zero and negative), AC-FLR-007 (the refusal cases; the acceptance cases take a mutant that refuses at the bound), and AC-FLR-009 (both subtests).

Exit: AC-FLR-005, AC-FLR-006, AC-FLR-007, AC-FLR-008, and AC-FLR-009 are GREEN, and the AC-FLR-008 mutant is recorded.

### M2: Renew verb, MCP tool, and the expired refusal with no write (Priority High; REQ-FLR-001, -002, -003, -004, and the verb side of REQ-FLR-011)

Files (run phase):

1. `internal/cli/factory_card.go`: a `renew <card>` verb in the factory command group, lane-only, beside `stage` (`:1661`). It prints the new expiry (REQ-FLR-001). With `--json` it prints `result` (one of `renewed`, `refused-expired`, `refused-not-holder`, `error`) and `expires_at` (A17, REQ-FLR-011).
2. `internal/homestate/card_transition.go`, `RenewLease` (`:441-456`): the expired branch returns `ErrLeaseExpired` without calling `applyLeaseExpiry` and without writing. The non-holder branch is unchanged (REQ-FLR-003). This change is ruled at decision-index Q3 (d-20261010T095731Z-5ddc) as refusal with no write, as REQ-FLR-004 is written.
3. `internal/mcp/catalog.go` declares `factory_renew` as write-capable. `internal/cli/mcp_factory_card.go` handles it through the same function as the verb (REQ-FLR-002). The pins in `internal/mcp/catalog_test.go` are updated (B4).
4. `.claude/rules/moai/core/moai-mcp-tools-catalogue.md` and its template mirror gain the tool in the same change.
5. `TestRenewLeaseRefusalsAndExpiry` (`internal/homestate/f7_record_failure_test.go`), the reaped-state assertion at lines 36-42 (the state and the cleared holder). Under REQ-FLR-004 the card keeps its state and holder, and its version is unchanged. This assertion is rewritten in this milestone, and the test is GREEN with the rewritten assertion. Its other assertions are unchanged: the missing-card and unleased-card refusals (lines 19-27), the expired class (lines 33-35), and the zero-clock renewal (lines 44-61). Decision-index Q3 is ruled as refusal with no write (d-20261010T095731Z-5ddc), so this rewrite is made, and REQ-FLR-004 and AC-FLR-004 stand as written.

RED first: the rewritten reaped-state assertion is an observed RED on the pre-M2 `RenewLease`, which returns the card as `assigned` with no holder. Record that output before the M2 edit. Predicted for AC-FLR-001 to AC-FLR-003 and the verb half of AC-FLR-011: the verb and the tool are absent, so each test fails on its assertion. AC-FLR-004: the test runs against the reaping branch and fails on the unchanged-state assertion.

Exit: AC-FLR-001, AC-FLR-002, AC-FLR-003, AC-FLR-004, and the verb half of AC-FLR-011 are GREEN. `TestRenewLeaseRefusalsAndExpiry` is GREEN after its state assertion is rewritten (item 5). The baseline recorded in §C.3 is the pre-M2 output.

### M3: Watchdog renewal, stall window, and standing wake (Priority Medium; REQ-FLR-010, -011, -012, -013)

Files (run phase):

1. `internal/template/templates/.claude/skills/moai-lane-watchdog/SKILL.md`, then the local mirror `.claude/skills/moai-lane-watchdog/SKILL.md`: `allowed-tools` (line 17) gains `Bash(moai factory renew:*)`. A first step renews the lane's card lease before the measurement in §1 and records the verb's `result` in `renew_result` (REQ-FLR-010, REQ-FLR-011). A refused renewal does not end the pass. §1 reads `last_change_at` for the stall verdict of REQ-FLR-013 in place of the previous-observation clause. Then `make build`.
2. `internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md`, then the local `.claude/rules/moai/workflow/auto-semantics.md`: §4 gains the fields `renew_result` and `last_change_at` and states the verdict of REQ-FLR-013. §5.1 "Arm" is a minute step of `workflow.factory_lease.renew_interval` written `*/N * * * *` (B6). The arm sentence and the cost sentence that justify the 20-minute cadence are rewritten (B5).
3. `.claude/rules/moai/workflow/factory-dispatch.md`, and its template mirror `internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md`: line 138, "every 20 minutes", becomes the renewal-period form (B5).
4. `.claude/rules/moai/workflow/factory-dispatch-mechanics.md`, and its template mirror: one sentence on the renew verb in the lease discussion.
5. `internal/hook/session_start_factory_i18n.go`: the four locale strings (`:132`, `:188`, `:242`, `:295`) render the arm expression `*/N * * * *` from the effective `renew_interval` (B7) instead of the literal. The tool names, the skill name, and the locale script stay as they are.
6. `internal/hook/lane_recheck_rule_test.go:22`: `recheckRuleTokens` drops `"7,27,47 * * * *"`. The arm-expression check moves to the test of AC-FLR-014.
7. `internal/template/lane_recheck_doctrine_test.go:61`: the need list of `## 5. The awaken rule` replaces `"7,27,47 * * * *"` with `"renew_interval"`.

RED first (predicted, observed in the run phase on the unchanged text): AC-FLR-010 (the procedure opens with measurement and lists no renew command), the procedure half of AC-FLR-011 (no refusal record), AC-FLR-013 (the literal is present in the paragraph), AC-FLR-014 (the rendered rule carries the literal in every locale), and the `rule-text` pin of AC-FLR-015 (§4 names no `last_change_at`).

Exit: AC-FLR-010, AC-FLR-011 (both halves), AC-FLR-013, AC-FLR-014, and AC-FLR-015 are GREEN.

### M4: Late-pass margin and expired refusal (Priority Medium; REQ-FLR-001, -004, -007, -009, -012)

Files (run phase):

1. An injected-clock test in `internal/cli` (`TestRenewCadenceKeepsLeaseLive`), using the factory test clock (`factoryCardNow`). It runs for three (D, I) pairs: (15 minutes, 5 minutes), the defaults; (10 minutes, 5 minutes), the boundary; and (20 minutes, 4 minutes), a non-default pair. On-schedule renewals at 0, I, 2I, and so on through 3D succeed. A renewal due at T plus I and requested at T plus I plus (D minus I minus one second) succeeds and extends the lease to its request time plus D. Renewals requested at T plus D and at T plus D plus one second are refused with `refused-expired`, with no change to state or version.

RED: the scenario is green on the M2 code when M2 is correct, so the criterion's RED comes from three mutants. Each is executed against the M2 code, its failing output is recorded, and then the code is restored.

- Mutant A: `Card.LeaseExpired` treats the expiry instant as live (strict `<` in place of the at-or-before test of `card_record.go:146`). Failing output: the renewal requested at T plus D succeeds where `refused-expired` is required.
- Mutant B: the renewal refuses any request made more than one renewal period after the previous renewal. Failing output: the renewal requested at T plus D minus one second is refused where success is required.
- Mutant C: the renewal leaves the expiry unchanged. Failing output: the on-schedule renewal at T plus D is refused as expired where success is required.

Exit: AC-FLR-012 is GREEN with the real code, and each mutant's output is recorded as its RED.

### M5: Mirrors, catalogue, and quality gates (Priority Low; mechanical; no behaviour change)

1. `make build`, then the key-scoped comparison of `workflow.yaml` and the byte comparison of each changed rule and skill mirror (QG-4), with the overrides preserved: `factory-dispatch.md`, `auto-semantics.md`, `factory-dispatch-mechanics.md`, the watchdog `SKILL.md`, and the MCP catalogue rule.
2. QG-1 to QG-3 (format, lint, race, coverage), each with its verbatim output and the HEAD SHA beside it.
3. QG-5: each commit message names card t1663.

Exit: QG-1 to QG-5 pass, or each shortfall is recorded as a Gap.

## §G Anti-patterns to avoid

- Clamping an out-of-bound value silently, or writing a default into a shipped configuration without its decision row (Q2 is ruled by d-20261010T095818Z-c501).
- Restating the 20-minute literal as a current cadence anywhere: auto-semantics §5.1, `factory-dispatch.md`, the watchdog skill, the hook, or a test.
- Measuring the stall window from the previous snapshot at a period below N, which can never give a verdict.
- Filling an `Operator verdict:` line with a value copied from the draft SPEC. A delegated verdict copied from a board record is recorded with its decision id. This plan writes the DEFAULT-APPLIED line of Q5 and the delegated verdicts of Q1, Q2, Q3, and Q6.
- Changing the queue claim to read the card key, or changing the audit-verdict parser.
- Naming in a test a Go symbol that does not exist yet. That is TOOL_FAILURE, not a predicted RED.
- Recording a mutant's output as the observed RED of the real code, or recording a green-before-change case as a pass.
- Running `git -C <worktree>`, `cd && git`, or a bare `git stash` in a worktree session.
- Reassigning `FactoryLeaseDuration` at process start instead of reading the handle (B2).
- Running `go test ./...` locally. Verification is lane-local and scoped to the change, and CI runs the full suite.

## §H Cross-references

- `spec.md` (requirements, exclusions, residual risk), `acceptance.md` (criteria and Definition of Done), `decision-index.md` (Q1 to Q6), and `progress.md` (§E.1).
- `.moai/specs/SPEC-FACTORY-RECORD-REPAIR-001/`: the draft that this card split from. It is not modified here.
- `.claude/rules/moai/workflow/auto-semantics.md` §3 (the stall window), §4 (the snapshot and the verdict), and §5.1 (the standing wake).
- `.claude/rules/moai/workflow/cache-aware-execution.md` directive 9 (session length and the cache window).
- `.claude/rules/moai/development/verification-completeness.md` §1.2 (the three-part check), §1.3 (continued firing), and §2 (the two-cell pair and the mutant probe).
- `.claude/rules/moai/workflow/tdd-result-contract.md` (RED and GREEN semantics).
- `.claude/rules/moai/core/verification-claim-integrity.md` §2 (attribution of each claim to a run and a tree) and §2.2 (the judging build of a tool measurement).

## Decisions owed

These map to decision-index.md. Q1 (the values, names, and armed form of the two keys) and Q2 (refusal or clamping) carry delegated operator verdicts with decision id d-20261010T095818Z-c501. Q3 (the expired-renewal form) and Q6 (the stall window reference) carry delegated operator verdicts with decision id d-20261010T095731Z-5ddc. Q4 is POLICY-COVERED. Q5 is DEFAULT-APPLIED at plan close.
