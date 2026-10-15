# SPEC-JEV-RESOLVED-MODEL-001 — Progress

Card: t1768 · Class C · Tier M · plan-phase artifacts authored 2026-10-15; repair round 1 applied the same day after plan-audit iteration 1 (FAIL, 0.65, audited commit `cfb5517962088b09936dfbb8a2ff5b80e63a8893`). Amends `SPEC-JEV-CORE-001` (v0.5.0) and `SPEC-JEV-OPTIN-MEASURE-001` (v0.2.0) in place.

## §E.1 Plan-phase Audit-Ready Signal

| Item | State |
|---|---|
| SPEC ID regex check | `PASS` (executed as a Bash command; verbatim output in plan.md §C) |
| SPEC ID collision | none (`ls -d .moai/specs/SPEC-JEV-RESOLVED-MODEL-001` → `No such file or directory` before the plan commit) |
| Tier | M — REQ 14 / ceiling 16; AC 16 / ceiling 16 |
| Artifact set | spec.md · plan.md · acceptance.md · progress.md |
| Requirements | 14 (REQ-JRM-001 … REQ-JRM-014) |
| Acceptance criteria | 16 (AC-JRM-001 … AC-JRM-016); AC-JRM-010 adopted (OQ-1 option A) |
| Clarification markers | none in the SPEC directory. OQ-1, OQ-2, and OQ-3 are decided (card t1768, board) and recorded in plan.md §G |
| Amendments applied | `SPEC-JEV-CORE-001` 0.4.0 → 0.5.0 (unchanged in the repair round) and `SPEC-JEV-OPTIN-MEASURE-001` 0.1.0 → 0.2.0 (the straddle clause of AC-JEVO-011 removed in the repair round). Each moves `completed → in-progress (amendment)`, with `amendment_of` set to itself and a `## Amendments` section |
| Predecessor | `SPEC-JEV-CORE-001`, `SPEC-JEV-OPTIN-MEASURE-001` (declared as `related_specs`, not `depends_on`; see plan.md D5) |
| Successor | card t1762 (dispatch order: t1687, t1768, t1762, t1767) |
| Vision revision | not triggered (`.moai/project/product.md` carries no statement of the pin or the alias) |
| Status transition | (none) → draft |
| plan_status | `pending-plan-audit`: repair round 1 is applied, and the plan-auditor re-audits the delta over the fix scope before the signal becomes audit-ready |

### Repair round 1 — status per audit item

| Audit item | Status | Where |
|---|---|---|
| D1 (open-question marker, change notice) | Decided: option A. Marker removed | plan §G; REQ-JRM-010; AC-JRM-010 |
| D2 (open-question marker, shadow terms) | Decided: option (a). Marker removed | plan §G; REQ-JRM-011; plan §D.4 |
| D3 (open-question marker, regression set) | Decided: option (b). Marker removed | plan §G; REQ-JRM-010; REQ-JRM-011 |
| D4 (AC-JRM-014 red-now cells) | Fixed. AC-JRM-014 is a guard with line-anchored GRN cells; EV-RED-13 and EV-RED-14 are kept as historical base-tree cells; AC-JRM-014 is out of the Definition of Done's red-to-green list | acceptance.md AC-JRM-014 and ledger |
| D5 (REQ-JRM-013 form) | Fixed. REQ-JRM-013 is one shall clause; the sync return is REQ-JRM-014, with AC-JRM-016 for the sync phase | spec.md §C.4 |
| D6 (ModelID allow-list) | Fixed. Explicit allow-list, and no attribution field receives the constant. Plan D1, the §D.3 Run step, and the M6 gate are aligned | acceptance.md AC-JRM-005; plan §B D1, §D.2, §E M6 |
| D7 (section title casing) | Fixed. `Jev Model Attribution` (Title Case, which matches the headings of the rule file) in REQ-JRM-011, AC-JRM-011, EV-RED-11, and plan §D.4 | spec.md; acceptance.md; plan §D.4 |
| D8 (AC-JRM-011 depth) | Fixed. Anchored content checks, with a positive control | acceptance.md AC-JRM-011 |
| D9 (non-discriminating fixture) | Fixed. `jev-9.8.7` in AC-JRM-002, 004, 007, and 009, and in every new fixture; a two-call case in AC-JRM-002 | acceptance.md; plan §D.6 and §D.7 |
| D10 (unadopted REQ-JRM-010) | Fixed. AC-JRM-010 is adopted with a RED-now cell and a green-path cell; the Definition of Done follows it | acceptance.md AC-JRM-010 and Definition of Done |
| D11 (live consumer omitted; selector; count) | Fixed. `internal/cli/todo_auto_rank_test.go` added (its fixture gains a versioned model, and it gains an unattributed case); `internal/cli/todo_auto_rank.go` gains the reason `jev-unattributed`; the §F selector is widened and anchored. Importer recount: 23 importers, of which 12 change; with the 5 non-importer files, 17 files change. The tier is not changed | plan §D.1, §D.6, §F |
| D12 (green-path cells absent) | Fixed. A green-path cell per release-blocking AC, naming the milestone and the `--- PASS:` line; the Definition of Done requires the line, not exit 0 alone | acceptance.md green-path cells and Definition of Done |
| D13 ("every usage record") | Fixed. The positive clause of REQ-JRM-005 is scoped to answered calls; "never the request alias" stays for every record | spec.md §C.1 |
| D14 (OPTIN straddle clause) | Fixed. The straddle clause is dropped from AC-JEVO-011. The OPTIN HISTORY row never named the straddle, so it needed no edit | SPEC-JEV-OPTIN-MEASURE-001 acceptance.md AC-JEVO-011 |
| D15 (§F misstates the open decisions) | Fixed. §F names the requirements each decision can change, and records the three decisions as settled | spec.md §F |
| D16 (AC-JRM-015 traces to no REQ) | Fixed (optional). Mapped to REQ-JRM-013 | acceptance.md AC-JRM-015 |
| D17 (unanchored amendment greps) | Fixed (optional). Line anchors. The prior_completed_sha key is a bullet of `## Amendments`, so its anchor is `^- prior_completed_sha: ` | acceptance.md AC-JRM-014 |
| D18 (sibling SPEC pin wording) | Not applied. It is outside the list for this round. Still open for the leader | — |
| D19 (lint warning not recorded) | Fixed (optional). The tree-built lint result, warning included, is recorded below | this section, lint record |
| D20 ("answered call"; REQ-JRM-001 scope) | Fixed (optional). "Answered call" is defined in the spec.md Definitions; REQ-JRM-001 is narrowed to the request's model | spec.md; REQ-JRM-001; AC-JRM-001 |

Decisions the author took in this round, for the leader to confirm: the reason `jev-unattributed` for an unattributed answer in the todo auto-rank surface (plan §G; the existing one-reason-per-condition invariant forbids sharing `jev-unreachable`), and the kickoff fallback `jev_call_failed` for an unattributed answer (plan §G; fail-closed).

### Tree-built lint record (AC-JRM-015, D19)

Tree: HEAD `cfb551796`, with the repair-round SPEC edits uncommitted. Binary: `go build -o <scratchpad>/moai-tree ./cmd/moai` (plain build, no ldflags), invoked by path. Command: `<scratchpad>/moai-tree spec lint .moai/specs/SPEC-JEV-RESOLVED-MODEL-001 .moai/specs/SPEC-JEV-CORE-001 .moai/specs/SPEC-JEV-OPTIN-MEASURE-001`. Verbatim output:

```text
SEVERITY  CODE                  FILE                                            LINE  MESSAGE
--------  ----                  ----                                            ----  -------
WARNING   StatusGitConsistency  .moai/specs/SPEC-JEV-OPTIN-MEASURE-001/spec.md  1     SPEC SPEC-JEV-OPTIN-MEASURE-001 frontmatter status 'in-progress' disagrees with git-implied status 'implemented'

0 error(s), 1 warning(s)
```

The warning is pre-existing: the plan audit's iteration-1 verdict reports the same row at `cfb551796`. It compares the OPTIN amendment's `in-progress` status with the git-implied status. Whether it clears at the sync phase is not measured in this round. AC-JRM-015's "no errors" holds: 0 errors.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
