# SPEC-LANE-READ-COMPACT-001 — Implementation Plan

Version 0.2.0 (2026-10-19, card t1810 plan-phase premise refresh). Part (a) of the split of `SPEC-LANE-READ-VERBS-001` (card t1777) by board ruling `d-20261015T122748Z-6d6b`; card t1810. Tree pin: b0a37f759, the parent's pin; the carried cells keep their own pins. Re-anchored 2026-10-19 at tree a3e8d1409c: the line references main's four days of drift moved were re-pinned by content, and the size premises were re-measured (§ C.1; ledger L-14 to L-16).

## A. Context and Tier Judgment

- Scope: the todo compact form (parent REQ-LRV-001 to REQ-LRV-004), the factory status default (parent REQ-LRV-005 to REQ-LRV-007), the queue-and-factory doctrine sentence (from parent REQ-LRV-011), and the todo and factory parts of the parent's compatibility and help requirements (REQ-LRV-012, REQ-LRV-013). Id map: REQ-LRC-001 to REQ-LRC-007 are parent REQ-LRV-001 to REQ-LRV-007; REQ-LRC-008 to REQ-LRC-010 are new texts for the doctrine, compatibility, and help parts. AC-LRC-001 to AC-LRC-007 are parent AC-LRV-001 to AC-LRV-007, AC-LRC-008 is parent AC-LRV-010 with this SPEC's sentence, AC-LRC-009 is parent AC-LRV-011, and AC-LRC-010 and AC-LRC-011 are new texts for the help and lint parts.
- Requirements: 10 (Tier M ceiling 16). Acceptance criteria: 11 (ceiling 16).
- Files: 8 existing files modified and 3 new files, which makes 11. Modified: `internal/cli/todo.go`, `internal/cli/factory_card.go`, one `internal/homestate` file for the active-run query (§ D.2; the file is chosen at run-phase), the three `factory-dispatch.md` copies, and the two `gtd.md` copies. New: `internal/cli/lane_read_compact_test.go`, `internal/template/lane_read_compact_doctrine_test.go`, and `internal/cli/testdata/lane_read_compact/factory_full.golden.json`. If the run phase places `candidates_withheld` in `internal/cli/factory_quota.go`, the count is 12.
- Estimated change: about 400 to 600 lines, mostly tests and fixtures. Tier: **M**.
- Artifact set: `spec.md`, `plan.md`, `acceptance.md`, `progress.md`, `decision-index.md`, `decision-anchor.md`, and `measure-caps.py` (a byte-identical copy of the parent's script, which the ledger cells L-12 and L-13 run).

## B. Decisions and Their Routing

운영자 대리 standing 아래 카드 범위 브레인 ruling의 앵커 = 보드 id(d-20261015T031405Z-dfd8, d-20261015T113856Z-6f23); `moai decision read` 다이제스트 결함은 후속 카드 t1809

- Q1 (dfd8): the todo default stays the full record, and an opt-in `--compact` carries the quiet form. Q2 (6f23): compact row cap 6. Q3 (6f23): factory default cap 3 plus 2 quota candidates, full output through `--run <id>`. Q8 (6f23): `--run <id>` keeps its pre-change behavior. Q7 (6f23): the doctrine sentence goes in the role-core region of `factory-dispatch.md`. Q9 (6f23): the scope change, todo and factory half. Q6 stays EVIDENCE-NEEDED (§ B.5). The rows and their pinned records are in `decision-index.md`.

### B.5 Role-injection headroom (Q6)

- EVIDENCE-NEEDED. The measurement is `internal/hook/role_composite_budget_meas_test.go`, run before and after this SPEC's sentence lands (milestone M4). If there is no headroom, the placement is decided again, with the fallback `.claude/rules/moai/core/agent-common-protocol.md`.

## C. Measured Basis for the Bounds (Q2 and Q3)

### C.1 Method

- Each figure is the byte count of the verb's output shape, computed with the encoder rules the verb already uses: the todo default and the compact form use `json.Marshal` (compact, HTML-escaped, followed by a newline); the factory status uses `json.Encoder` with a two-space indent, HTML escaping, and a trailing newline; the session list uses `json.MarshalIndent` with a two-space indent.
- The emulation reproduces the current outputs exactly. The current todo default (2,626,934 bytes) and the current factory default (117,230 bytes) match byte for byte, and the current session default (43,135 bytes) matches byte for byte, on the snapshots taken in this run. The emulation is the planning estimate only. The command is `python3 -I .moai/specs/SPEC-LANE-READ-COMPACT-001/measure-caps.py <BIN>` (read-only; it runs `todo list --json`, `factory runs`, `factory status --json`, and `session list --json`, validates its emulation byte-for-byte against each live output, and then prints the candidate caps). Its verbatim output is in `acceptance.md` § D, ledger L-12. The run-phase bound tests (§ D.6) are the citable evidence for the caps.
- Operative premise, re-measured 2026-10-19 at tree 74ad197380 by the card lane under the resource rules (ledger L-14, L-15, L-16): the full todo default is 2,897,435 bytes (L-02 carried 2,629,774 on 10-15, about +10%), and the factory status default is 201,400 bytes listing 368 cards across 15 run ids (L-03 carried 117,251 · 218 · 14, about +72% in bytes). Both new flags are still absent — `--compact` and `--full` each still refuse as an unknown flag — so the L-01 and L-04 RED premises hold. The four days of drift strengthened the problem definition; the worst-case tables of § C.2 and § C.3 are fixture-based and are not moved by it.
- Worst case: each string field at its declared maximum JSON width. Todo: `id` 8 bytes, `state` 6 bytes (`queued` or `picked`), `prefix` at the maximum JSON width of 6 bytes per rune across 40 runes plus `...` (243 bytes), and every integer 10 digits. Factory, revised after plan audit iteration 1: `run_id` 6 characters; `card_id` up to 8 characters; `state` at the longest of the 21 enumerated states (`needs-decision`, 14 characters, `internal/homestate/card_record.go:33`, the `cardStates` slice at `:62-67` after the 2026-10-19 re-anchor); `stage` at the longest enumerated value (`merge-ready`, 11 characters); `mode` `parallelizable` (14) and `priority` `normal` (6) (`internal/factory/classification.go:34-39`); `lease_holder` cut to 16 bytes of encoded width, so that 16 raw `<` characters (96 encoded bytes) cannot reach the output; `lease_expired` `false`; counts 6 digits; durations 12 characters; and the quota block at its own worst case: pressure on, both windows `fresh` with every optional field set (`used_percentage` at the 25-byte longest float64 encoding, three 20-character RFC 3339 timestamps, `holds_claude_lane` `false`; the window set is fixed at two windows, `internal/cli/factory_quota.go:29-30`, `:77`, `:114-117`), two candidate lanes with labels at 16 encoded bytes, and `candidates_withheld` and `unknown_lanes` at 6 digits. The 0.2.0 basis counted the quota block as measured with pressure off, which is not its worst case: under pressure the block carries one candidate per live GLM or GPT lane (`internal/cli/factory_quota_lanes.go:96`).


### C.2 Todo compact row cap

| Cap | Worst case, bytes | Within 2048 | Note |
|---|---|---|---|
| 5 | 1,673 | yes | |
| **6** | **1,963** | **yes** | **chosen** |
| 7 | 2,253 | no | |
| 8 | 2,543 | no | |
| 10 | 3,123 | no | |

- Measured on the live queue at the observation (163 live items, at the L-12 run): cap 6 gives 817 bytes, cap 10 gives 1,273, cap 15 gives 1,833, and cap 20 gives 2,355 (over the bound on that queue).
- Correction to the working hypothesis. The figure of about 1,755 bytes at cap 10 assumed 3-byte CJK prefixes with no JSON escapes. The prefix function does not escape `<`, `>`, `&`, control characters, U+2028, or U+2029, and the JSON encoder writes each of them as a 6-byte escape. Counting those runes at their JSON width, cap 10 is 3,123 bytes and fails. For CJK-only text, cap 10 is 1,923 bytes (pass) and cap 11 is 2,093 bytes (fail). The bound must hold for any queue content, so the cap is 6.
- Operator option, not chosen: a prefix bounded by encoded width (the first 40 JSON-encoded bytes, then `...`) would allow cap 20, at a worst case of 2,023 bytes (cap 21 gives 2,113 and fails). This changes the prefix rule for the compact form only.


### C.3 Factory default row cap (worst case under quota pressure)

| Cap | Worst case, bytes | Within 2048 | Note |
|---|---|---|---|
| **3** | **1,899** | **yes** | **proposed** |
| 4 | 2,161 | no | |

- Measured in ledger L-13 (`acceptance.md` § D) with the encoder emulation of `measure-caps.py`, loaded unchanged and validated byte-exact against the three live outputs in the same run.
- Why the 0.2.0 table is withdrawn. It gave cap 6 = 1,924, cap 7 = 2,124, cap 8 = 2,324, and cap 10 = 2,724, with the quota block counted pressure off and the holder cut at 16 raw bytes. Measured again in L-13, the 0.2.0 shape at cap 6 reaches 2,113 bytes with live windows, pressure on, and two candidate lanes, and 2,404 bytes with a holder of 16 `<` characters. Both candidates in the card (8 and 10) fail by more.
- With two candidates, three rows is the largest cap that fits. With three rows, three candidates also fit (1,978 bytes) and four do not (2,057).
- Real active run (68 cards, `tmvgvn`) at cap 3: 1,338 bytes with the live quota block, and 1,557 bytes with live windows, pressure on, and two candidates.
- Operator options, not chosen (same worst case): (A1) candidates as a count only: cap 4 = 2,036 (pass), cap 5 = 2,298 (fail). (A2) rows without `mode` and `priority`, which would need a REQ-TCD-010 carve-out: cap 4 = 1,921 (pass), cap 5 = 2,123 (fail). (B) rows without `run_id`, valid only while exactly one run is active: cap 3 = 1,821, cap 4 = 2,057 (fail), so no gain. (C) a compact JSON encoding for the default: cap 6 = 1,926 (pass), cap 7 = 2,114 (fail); this changes the default's formatting, not only its rows.

## D. Planned Change by Surface

### D.2 `moai factory status --json` (REQ-LRC-005, REQ-LRC-006, REQ-LRC-007)

- Flag. Add `--full` (boolean) to `newFactoryStatusCommand` (`internal/cli/factory_card.go:3113` after the 2026-10-19 re-anchor; it was `:3009` at tree b0a37f759). `--run` and `--json` keep their meaning.
- Default (no `--run`, no `--full`). Cards come only from runs whose `runs.status` is `active`. That column is read by `classifyRuns` (`internal/homestate/factory_run_retire.go:172-178`), and `moai factory runs` reports the same column. The active-run set is read through a read-only query added to the existing `internal/homestate` factory database file. The query does not run the owner classification, which the default does not need.
- Rows. `run_id`, `card_id`, `state`, `stage`, `mode`, `priority`, `lease_holder`, and `lease_expired`. `mode` and `priority` keep REQ-TCD-010 on the default (`factoryCardView`, `internal/cli/factory_card.go:3044` after the 2026-10-19 re-anchor). `lease_holder` is cut at a rune boundary to the longest prefix whose JSON encoding, HTML escapes included, is at most 16 bytes. Row cap 3 (§ C.3). The rest of the active cards are counted in `withheld`.
- Envelope. `run` (the `--run` value, as today), `cards`, `withheld`, `unavailable_count` (an integer), `lease_duration`, `renew_interval`, `quota`, and `unavailable_skipped` when it is non-zero, as today.
- Quota block (`internal/cli/factory_quota.go:267`, `factoryQuotaStatusBlock`, after the 2026-10-19 re-anchor; the window view follows at `:300`). With pressure off the block is unchanged. With pressure on, `pressure`, `windows`, `unknown_lanes`, and `warning` are unchanged, and the steering view's `candidates`, one entry per live GLM or GPT lane (`internal/cli/factory_quota_lanes.go:96`, so unbounded), is cut to the first 2 in the registry's label order, each label cut like `lease_holder`, with a new `candidates_withheld` count. This departs from REQ-QAS-019, which asks the status JSON for the candidate lanes, on the default path only. The AC-QAS-019 test reads `status --run <id> --json` (`internal/cli/factory_quota_lanes_test.go:229-240`), which keeps the whole block, and so do `--full` and the text view's steering line. The unavailable entries (`internal/homestate/card_unavailable.go:28-36`) carry free-text errors and are unbounded, so the default reports their count, and `--full` lists them.
- Encoder. Unchanged: `json.Encoder`, two-space indent, HTML escaping, and a trailing newline (`internal/cli/factory_card.go:3174-3175` after the 2026-10-19 re-anchor).
- `--run <id>`. Keeps the pre-change behavior: every card field and no row cap.
- `--full` without `--run`. Emits the pre-change output for all runs, byte-identical for the same snapshot (§ E, the fixture golden).
- Scope. `factoryStatusReport` (`internal/cli/factory_card.go:3072` after the 2026-10-19 re-anchor) is used only by this verb. The MCP factory tools are not changed.


### D.3 `moai todo list --json --compact` (REQ-LRC-001, REQ-LRC-002, REQ-LRC-003)

- Flag. Add `--compact` (boolean) to `newTodoListCmd` (`internal/cli/todo.go:1283` after the 2026-10-19 re-anchor; it was `:1240` at tree b0a37f759). The default path is unchanged. `runTodoListRoot` keeps its JSON branch, `json.Marshal(todoJSONProjection(rec))` followed by a newline (`internal/cli/todo.go:1205`, the marshal at `:1224`, after the same re-anchor).
- Compact output. `json.Marshal` (compact, HTML-escaped, followed by a newline), with the keys in this order: `project_uuid`, `version`, `last_seq`, `live`, `withheld`, `items`, `findings_count`, `archived_count`, and `dropped_count`. `live` counts the items whose state is not `dropped`. `items` holds the first at most six live items in queue order, each as `{id, state, prefix}`, where `prefix` comes from `todoTextPrefix` (`internal/cli/todo.go:1824` after the 2026-10-19 re-anchor). `withheld` is `live` minus the rendered rows. The compact output carries no `findings` array and no `runtime` object.
- `--limit`. Ignored with `--compact`, as it is with `--json` (AC-TSP-021).
- Not added. `--state` and `--fields` (dropped from 0.1.0, § A).
- Row cap. Six, as a named constant in `internal/cli/todo.go` (the name is chosen at run-phase), with the bound test in § D.6.


### D.4 Lane-facing reads and the doctrine sentence (REQ-LRC-004, REQ-LRC-008)

- Doctrine sentence. One sentence, identical in the three copies: inside a role-core region of `.claude/rules/moai/workflow/factory-dispatch.md` and of its template copy `internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md`, and once in the shared-policy copy `.moai/policies/workflow/factory-dispatch.md`, which carries no role-core region. The role injection extracts regions from the deployed rule, falls back to the shared-policy copy only when the deployed rule is absent (`internal/hook/role_rules.go:150-162`), and takes the embedded template copy's core whenever the file it reads has no role-core marker (`internal/hook/role_rules.go:557-561`). A region in the shared-policy copy would therefore be read by nothing, and this plan adds no markers there (AC-LRC-008). The sentence:

  > Lanes read queue and factory state only through the quiet forms (`moai todo list --json --compact` and `moai factory status --json`); a full dump (`moai todo list --json` without `--compact`, or `moai factory status --json --full`) is an operator-requested exception, never a lane default.

- Parity. `TestAutoPickMirrorParity` (`internal/cli/todo_auto_pick_doc_test.go:279`) requires the local and template copies of `factory-dispatch.md` to differ only by the live-only line that names `moai worktree sweep`. The doctrine sentence is added at the same position in both copies.
- Lane-facing consumers in this tree, as measured. (a) `.claude/skills/moai/workflows/gtd.md` and its template copy name `moai todo list --json` on five lines (re-anchored 2026-10-19 at tree a3e8d1409c: the four lines of tree b0a37f759, re-measured at 9cf21f079 in ledger L-06, moved to lines 218, 261, 262, and 475, and card t1827 added a fifth between the authoring tree and today): the database-paths preference bullet (line 47 in both copies, `- Prefer the CLI: ...`, a lane-facing read instruction), the record description (line 218), two `jq` shape examples (lines 261 and 262), and the standing-source duplicate check (line 475, inside `### Not stacking duplicates`, whose heading is at line 470). The three `factory-dispatch.md` copies name it nowhere (re-checked 2026-10-19). (b) The `--auto` cycle reads the store in process (`internal/cli/todo_auto.go:278`, `store.LoadPure()`, after the 2026-10-19 re-anchor), so it has no CLI JSON call site for `--compact` to change. (c) No non-test Go file in `internal/hook` names `todo list`, so the lane join notice has no read to change and stays as it is; `internal/hook/session_start_factory_test.go:237` sits in `TestFactoryLeadNoticeCarriesDispatchDiscipline`, the leader-notice test (the re-anchor moved it from `:261`). (d) The foreman loop reads the full default (`internal/cli/todo_disclosure.go:18`), which is unchanged.
- `gtd.md` (both copies, kept byte-identical). One reading governs this plan and AC-LRC-004: a lane-facing read instruction names `--compact`, and the full form stays only where the subject is the full record itself. Accordingly: a paragraph after the record description's JSON block documents `moai todo list --json --compact` as the lane read (its shape, the row cap, and `withheld`) and states that the default is the full record; the database-paths preference bullet (line 47, added by card t1827) changes to name the compact form, because it is a read instruction and not a description of the record; the record description (line 218) and the two `jq` shape examples (lines 261 and 262) stay unchanged, because they describe the default record, whose JSON block `TestTodoListJSONShapeMatchesDoc` checks against the default output; and the standing-source duplicate check (line 475) names the full form and gives the reason in its paragraph: the check compares card text, and a compact row carries only a 40-rune prefix. That paragraph carries the reason with the literal `40-rune prefix`, which AC-LRC-004 asserts; neither `gtd.md` copy contains the literal today (0 for each at 2026-10-15T09:55:40Z, re-checked 2026-10-19). After the change, the four documented full-form lines are the record description, the two `jq` examples, and the duplicate check — the lines AC-LRC-004 lists — and no other line may name the full form.

- Ordering with part (b). `SPEC-LANE-READ-SESSION-001` adds a session sentence beside this one in the same role-core region. Whichever lands second places its sentence after the first; neither rewrites the other's sentence.

### D.5 Help text (REQ-LRC-010)

- `moai todo list --help` names `--compact` and the six-row cap. `moai factory status --help` names `--full` and the active-run row cap of three. The source is the cobra flag help strings in the two constructors.

### D.6 Tests and fixtures

- New test file `internal/cli/lane_read_compact_test.go`: the todo compact shape and byte bound, the factory default shape and bound, the factory `--run` and `--full` checks, the consumer check (REQ-LRC-004), and the help-text check.
- New test file `internal/template/lane_read_compact_doctrine_test.go`: the queue-and-factory sentence is present exactly once in each of the three copies, inside the role-core region of the two `.claude/rules` copies.
- Golden fixture (new, captured from the pre-change binary before any implementation commit): `internal/cli/testdata/lane_read_compact/factory_full.golden.json`. No existing test is edited.

## E. Frozen Contracts and Invariants

- `TestTodoListJSON_GoldenByteIdentity` (`internal/cli/todo_json_golden_test.go:65`) pins the default bytes; `internal/cli/testdata/ac_tst_012_golden.json` is not edited. `TestTodoListJSONIgnoresLimit` (`internal/cli/todo_list_limit_test.go:102`) is unchanged, and `--compact` also ignores `--limit`. `TestFactoryStatusReportsLeaseKeys` (`internal/cli/factory_status_lease_test.go:14`) asserts `lease_duration` and `renew_interval`, which stay in the default envelope.
- Baseline ordering (`verification-claim-integrity.md` §2.3): the golden fixture is committed in a commit that precedes the implementation commit.
- Mirror invariants: `TestAutoPickMirrorParity` (`internal/cli/todo_auto_pick_doc_test.go:279`) and `TestRuleTemplateMirrorDrift` (`internal/template/rule_template_mirror_test.go:378`).

## H. Risks

- R1 (High): the compact cap of 6 shows six of the live rows per call. Mitigation: the `withheld` count is printed; the full form stays available.
- R2 (Medium): the factory cap of 3 (worst case 1,899 bytes, ledger L-13) shows three of the active run's cards, and the caller's own card may fall outside them. Mitigation: the `withheld` count is printed; `--run <id>` and `--full` are the full forms.
- R3 (Medium): the role-injection budget (Q6) is unmeasured.
- R4 (Medium): live values move; green is judged on fixtures, never on the live store.

## I. Verification Plan for Run-Phase

- Run-phase reports E1 to E8 per `.claude/rules/moai/development/manager-develop-prompt-template.md` § Section E. RED before GREEN for each test-based criterion (E8). Targeted test runs only; no `go test ./...`. The judging binary is built from the implementation tree and invoked by its path.

### I.1 Milestones (M1 to M6)

- **M1 — Baseline golden.** Capture `factory_full.golden.json` from the pre-change build on the AC-LRC-007 fixture and commit it alone. Exit: the baseline commit precedes every implementation commit.
- **M2 — Todo compact form.** `--compact` and its row cap (§ D.3). Covers REQ-LRC-001 to REQ-LRC-003; flips AC-LRC-002 and AC-LRC-003 and keeps AC-LRC-001. Exit: `TestTodoListCompactShape` and `TestTodoListCompactByteBound` green, and `TestTodoListJSON_GoldenByteIdentity` green with no edit to its golden.
- **M3 — Factory status default.** Active-run rows, the field cut, and the quota candidate cap (§ D.2). Covers REQ-LRC-005 to REQ-LRC-007; flips AC-LRC-005 to AC-LRC-007. Exit: `TestFactoryStatusDefaultActiveOnly`, `TestFactoryStatusDefaultBoundedBytes`, `TestFactoryStatusFullByteIdentical`, and `TestFactoryStatusRunFlagUnchanged` green, with `TestQAS_AC019_AutoAndStatusRecommendNonClaudeLanes` still green.
- **M4 — Doctrine sentence.** The sentence in the three copies (§ D.4), with the Q6 measurement before and after. Covers REQ-LRC-008; flips AC-LRC-008. Exit: `TestLaneReadCompactDoctrinePresent` and `TestAutoPickMirrorParity` green, and the Q6 result in the progress record.
- **M5 — Lane-facing reads and help text.** Covers REQ-LRC-004 and REQ-LRC-010; flips AC-LRC-004 and AC-LRC-010. Exit: `TestLaneQueueReadsDefaultCompact` and `TestLaneReadCompactHelpNamesFlags` green.
- **M6 — Regression sweep and lint.** Covers REQ-LRC-009; flips AC-LRC-009 and closes AC-LRC-011. Exit: both AC-LRC-009 commands pass with non-empty swept sets, and `moai spec lint SPEC-LANE-READ-COMPACT-001 --json` exits 0.
