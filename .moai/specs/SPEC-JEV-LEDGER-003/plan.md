---
id: SPEC-JEV-LEDGER-003
title: "Plan — Jev decision-ledger cost report"
version: "0.1.0"
created: 2026-11-06
updated: 2026-11-06
---

# Plan — SPEC-JEV-LEDGER-003

Tree measured: `9cf8a1fc147605a8221206616d04fb43569961d8` (card worktree `t1971`, branch `WT-spec-jev-ledger-3`). Every claim about existing code below comes from a command run against that tree; the commands are listed in §I.

Card context: factory card t1971, Class C, Tier S. This is milestone M4 of the original SPEC-JEV-LEDGER-001 plan. It depends on nothing from SPEC-JEV-LEDGER-002 (card t1970, not landed).

## §A. Scope and files

| File | New/edited | Content | Expected lines |
|---|---|---|---|
| `internal/cli/jev_cost.go` | new | the `cost` subcommand: reader, de-dup, labels, `--json` | ~80 |
| `internal/cli/jev_cost_test.go` | new | the AC tests (§F test design) | ~100 |

Total estimate: 2 new files, ~80 production lines + ~100 test lines ≈ 180 changed lines — inside the Tier S card ceiling. `internal/cli/doctor_jev_test.go` is NOT edited: the reader needs no allow-list entry (decision D1), and the migrated AC-JEVL-014 part is carried as a non-importer proof instead of a dead allow-list row.

PRESERVE (untouched): `internal/jevledger/**` (the row shape is consumed as-is), `internal/cli/jev.go` (the cost subcommand attaches from `jev_cost.go`'s own `init()`), `internal/jev/**`, `internal/config/**`, `internal/template/**`, `.moai/config/**`.

## §B. Goal

Restated from spec.md §B: one read-only subcommand, `moai jev cost`, that reads the decision ledger, prints labelled aggregates, exits 0 on every ledger state — absent, empty, malformed — and makes no Jev call of any kind.

## §C. CORE-clause mapping (carried from the LEDGER-001 plan, scoped to the report)

| CORE clause | What it demands | How this SPEC satisfies it |
|---|---|---|
| REQ-JEVC-011, "a Jev answer shall not mutate ... a file" | display-only, with two named exceptions and the inert-sibling-record carve-out | The cost report is the carve-out's named reader: it reads ledger rows and prints aggregates, and writes nothing anywhere. Carve-out condition (i) holds — no code path reads a row back to write a field of anything it names; (ii) holds — rows render only through the labelled report, never as a mechanical measurement; (iii) holds — no decision, gate, verdict, mark, or disposition selects on a row (REQ-JEVL-014). SPEC-JEV-LEDGER-001 is the downstream SPEC that relied on the carve-out by name and named this reader; this SPEC is that reader arriving. |
| REQ-JEVC-012, no consultation *for* a listed act | the answer must not reach a verdict, gate, queue mutation, or user-surface behaviour change | The report consults no Jev at all (REQ-JEVL-014, REQ-JEVL-016). Its output changes no surface's behaviour: it is explicitly invoked, read-only, and its bytes reach nothing but the operator's terminal. |
| REQ-JEVC-013, labelling | any surfaced answer is labelled a model signal | The only surfaced answers are the aggregates themselves, under the literal heading `model signal aggregate`, with every token figure under the literal `estimated input tokens` unit (REQ-JEVL-015). |

## §D. Design notes — the reader

- **Command wiring.** `costCmd` is a cobra child of the package-level `jevCmd`, added from `jev_cost.go`'s own `init()` (`jevCmd.AddCommand(costCmd)`). `internal/cli/jev.go` is not edited. `Use: "cost"`, so the carried spelling `moai jev cost` is exact (decision D2).
- **Project root.** Resolved through the existing `resolveHookProjectRoot()` (`internal/cli/hook.go:790`, env-first `CLAUDE_PROJECT_DIR`, then `Getwd`) — the same resolver LEDGER-001's plan named for the recorder side. No new resolver.
- **Reading.** `os.Open` the ledger; a not-exist error is the absent case (REQ-JEVL-016). Scan lines (`bufio.Scanner`); decode each line into `jevledger.Row` directly (decision D5) — the row shape is consumed as-is and no parallel struct exists to drift. A line that fails to decode is malformed: skipped, counted. A line that decodes is a valid row, whatever its field values.
- **De-duplication.** A map keyed by `call_id` keeps the first row encountered per call (decision D4). calls = distinct `call_id` count; total estimated input tokens = the kept row's `input_tokens` summed over calls. The choice of kept row does not change any figure: the recorder repeats one call's `input_tokens` on every row of that call (LEDGER-001 plan §A/§F), and AC-JEVL-015's fixture is identical by construction. The per-row-sum mutant (1200 vs 600) is caught by the aggregates test.
- **Correct count and per-correct.** correct = valid rows (post-skip) whose `outcome_label` is exactly `correct` — rows, not calls, per the carried REQ text. Per-correct = total estimated input tokens ÷ correct. Zero correct → text prints `undefined`; JSON prints `null` (decision D3).
- **Output.** Plain text by default: the heading `model signal aggregate`, the figures each labelled `estimated input tokens`, and `skipped: N` printed only when N ≥ 1 (the carried REQ-JEVL-016 conditions the line on malformed lines existing). `--json` prints one flat object — `calls`, `estimated_input_tokens`, `correct`, `estimated_input_tokens_per_correct_decision`, `skipped` — always including `skipped` (decision D3). No figure is labelled charged or billed in either form.
- **Non-absent read failures.** An open or read error other than not-exist (a permission denial, a directory at the ledger path) takes the same fail-open path: zero calls, exit 0 (decision D6). The ledger is advisory local state and the report is display-only; a report command must not fail on it. Recorded in §H as the residual that an unreadable ledger is indistinguishable from an empty one in the output.
- **No client anywhere.** The file imports `internal/jevledger` and the standard library (plus cobra); it imports no `internal/jev`, constructs no client, and holds no URL literal (AC-JEVL-014 asserts this mechanically).

## §E. Milestones

One milestone. The label M4 is kept from the original LEDGER-001 numbering for continuity with its closure map ("AC-JEVL-015 final at M4"); LEDGER-001's landed M1/M2 and -002's M3 are other SPECs' milestones.

| M | Priority | Content | Files (prod / test) | Expected changed lines |
|---|---|---|---|---|
| M4 | High | Reader + report: command wiring, line scan, `jevledger.Row` decode, de-dup by `call_id`, correct count, undefined-when-zero, malformed-line skip with `skipped: N`, `estimated input tokens` labels, `--json` form | `internal/cli/jev_cost.go` (new), `internal/cli/jev_cost_test.go` (new) | ~80 prod + ~100 test |

### AC closure map (each AC closes only where its green path is satisfiable)

| AC | Final at | Why |
|---|---|---|
| AC-JEVL-015 | M4 | The report and its tests land together. |
| AC-JEVL-016 | M4 | The malformed-ledger behaviour needs the reader. |
| AC-JEVL-014 (part) | M4 | The boundary assertions need `jev_cost.go` to exist; before M4 the non-importer proof has no subject. |

No partial milestones: every green path is satisfiable only at M4.

## §F. Test design (green-path test names)

| Test | AC | Asserts |
|---|---|---|
| `TestJevCost_Aggregates` | AC-JEVL-015 | fixture ledger (3 calls × 2 rows, 2 correct): calls=3, tokens=600, correct=2, per-correct=300; literal heading `model signal aggregate`; literal label `estimated input tokens`; exit 0 |
| `TestJevCost_ZeroCorrectUndefined` | AC-JEVL-015 | zero correct rows → per-correct reads `undefined` (text) / `null` (`--json`) |
| `TestJevCost_AbsentAndMalformed` | AC-JEVL-016 | (i) absent, (ii) empty → calls=0, exit 0; (iii) malformed line between two valid rows → calls=2, tokens=300, correct=1, per-correct=300, `skipped: 1`, exit 0 |
| `TestJevCost_JsonMatchesTextAggregates` | AC-JEVL-015 | `--json` carries the same aggregates; `skipped` always present (0 included) |
| `TestJevCost_Boundary_NoClientOrEndpoint` | AC-JEVL-014 | source scan of `jev_cost.go`: no `internal/jev` import, no URL literal, no credential read (`jevcred`, credential-environment access); plus `TestJevCallPath_HasExactlyTheDeclaredConsumers` passing UNCHANGED with the file present — the allow-list maps files permitted to import `internal/jev`, and the guard passing with `jev_cost.go` absent from it is the by-name proof that the report imports none |

## §G. Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| A future caller of the report turns its output into a gate input | Low (later cards) | REQ-JEVL-014 binds the command itself; the CORE mapping (§C) names the clauses any such caller would have to satisfy or amend. |
| Concurrent -002 work edits `doctor_jev_test.go`'s allow-list (adding `hook_jev.go`) | Medium | This SPEC edits that file not at all (decision D1), so there is no merge surface on it from this card. |
| Ledger growth without rotation makes the scan slow | Low | Out of scope (spec §D). Rows are small; with no consumer in the shipped build, real ledgers are absent or empty today. |
| Valid-JSON rows with zero-valued fields (a truncated write that still parses) count as valid | Low | Carried semantics: malformed means "fails to decode". Recorded as a residual, not fixed. |

## §H. Residual risks recorded, not fixed

- A read failure other than not-exist (permission, a directory at the ledger path) is folded into the zero-calls fail-open path (decision D6): in the output it is indistinguishable from an empty ledger.
- A line that decodes but carries zero-valued fields counts as a valid row. "Malformed" is decode failure only.
- Concurrent writers cannot interleave a row (LEDGER-001 REQ-JEVL-011), so a torn line is not expected; if one ever appears it reads as one malformed line and counts in `skipped: N`.

## §I. Commands that grounded this plan (this run, tree `9cf8a1fc147605a8221206616d04fb43569961d8`, worktree `t1971`, branch `WT-spec-jev-ledger-3`)

- `git rev-parse HEAD` → `9cf8a1fc147605a8221206616d04fb43569961d8`; `git branch --show-current` → `WT-spec-jev-ledger-3`; `git status --short` → empty.
- `sh -c 'ls internal/cli/jev_cost.go; echo "exit=$?"'` → `ls: internal/cli/jev_cost.go: No such file or directory` / `exit=1`.
- Read `internal/jevledger/ledger.go`: the `Row` struct at `:46-67` (`call_id`, `input_tokens int`, `outcome_label`, `mode`, `action`, ...), the package doc ("never opens the ledger for reading"), and the `@MX:ANCHOR` on `Append` noting the -003 report fixtures depend on the fail-open contract.
- Read `internal/cli/jev.go`: `jevCmd` (package-level, `GroupID: "tools"`, `SetFlagErrorFunc`, bare `moai jev` is help + exit 0).
- Read `internal/cli/doctor_jev_test.go`: `jevImporters` (`:193`), `TestJevCallPath_HasExactlyTheDeclaredConsumers` (`:240`, the allow-list map), `TestJevCallPath_UnreachableFromDecisionSurfaces` (`:295`, decision-surface list and symbol scan; `jev_cost.go` is in neither list).
- `grep -rn "func resolveHookProjectRoot" internal/cli/` → `internal/cli/hook.go:790`.
- `test -d .moai/specs/SPEC-JEV-LEDGER-003` → absent before authoring.
- `git show 0c2d2aa4c2:.moai/specs/SPEC-JEV-LEDGER-001/{spec,acceptance,plan}.md` — the pre-split carried texts (REQ-JEVL-014 §C.6, REQ-JEVL-015/016 §C.7; AC-JEVL-014/015/016; the M4 milestone row, the CORE-clause mapping §C, closure map §E).
