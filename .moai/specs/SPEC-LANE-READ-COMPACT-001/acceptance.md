# SPEC-LANE-READ-COMPACT-001 — Acceptance Criteria and Evidence

## A. Tree Pin, Attribution, and Carriers

- Document-level tree pin: **b0a37f759**. The cells carried from `SPEC-LANE-READ-VERBS-001` keep their own pins and observation times; their ids are renumbered (plan § A). L-05 is new to this SPEC.
- Re-anchored 2026-10-19 at tree **a3e8d1409c** (card t1810 plan-phase premise refresh): the carried cells' pins are unchanged; the line references in AC-LRC-004 were re-pinned by content after main's four days of drift (210→218, 253→261, 254→262, 455→475, plus a fifth naming line at `:47` added by card t1827), and the operative size premises were re-measured as L-14, L-15, and L-16. The parent's `measure-caps.py` is not on this card tree (`SPEC-LANE-READ-VERBS-001` rides the t1777 branch); byte-identity of this SPEC's copy was re-verified against the parent's copy at 9cf21f079, its latest recorded state (`cmp` exit 0, 7,488 bytes each).
- Judging binaries: `<BIN>` is the parent's judging binary (`SPEC-LANE-READ-VERBS-001` acceptance § A); `<BIN2>` is the attested build named in `SPEC-LANE-READ-VERBS-001` ledger L-06.
- The cells L-12 and L-13 name this SPEC's copy of `measure-caps.py`, which is byte-identical to the parent's copy (`cmp` exit 0); their outputs were recorded with the parent's copy.
- Carrier form: each RED cell is one read-only invocation; where stdout and the exit code are read together, a one-process Python carrier runs the command unmodified.

## B. Two-Cell Matrix

| AC | Covers | RED-now cell (ledger) | Green-path cell (milestone; passing output) | Status |
|----|--------|-----------------------|---------------------------------------------|--------|
| AC-LRC-001 | REQ-LRC-001 | L-02 (regression guard); L-08 (golden green) | M2 (unchanged); `TestTodoListJSON_GoldenByteIdentity` green | Regression guard |
| AC-LRC-002 | REQ-LRC-002 | L-01 (`--compact` is an unknown flag; exit 1) | M2; `TestTodoListCompactShape` green | Release-blocking; RED-now observed |
| AC-LRC-003 | REQ-LRC-003 | L-01; L-12 (worst case at the 0.2.0 id width: cap 6 = 1,963 bytes; cap 7 = 2,253; iter2 recomputation with the 16-byte id cut: cap 6 = 2,011; cap 7 = 2,309 — plan § C.2) | M2; `TestTodoListCompactByteBound`: stdout ≤ 2048 bytes | Release-blocking |
| AC-LRC-004 | REQ-LRC-004 | L-06 (no line names `todo list --json --compact` in any of the five surfaces; exit 1) | M5; `TestLaneQueueReadsDefaultCompact` green | Release-blocking; RED-now observed |
| AC-LRC-005 | REQ-LRC-005 | L-03 (live store: 218 cards across 14 runs; the default is not active-only) | M3; `TestFactoryStatusDefaultActiveOnly` green | Release-blocking; RED-now observed |
| AC-LRC-006 | REQ-LRC-006 | L-03 (117,251 bytes > 2048); L-13 (worst case under quota pressure at the 0.2.0 identifier widths: cap 3 = 1,899; cap 4 = 2,161; iter2 recomputation with the identifier cut: cap 3 = 1,953; cap 4 = 2,233 — plan § C.3) | M3; `TestFactoryStatusDefaultBoundedBytes` green | Release-blocking |
| AC-LRC-007 | REQ-LRC-007 | L-04 (`--full` is an unknown flag; exit 1) | M3; `TestFactoryStatusFullByteIdentical` (golden from M1); `TestFactoryStatusRunFlagUnchanged` | Release-blocking for `--full`; regression guard for `--run` |
| AC-LRC-008 | REQ-LRC-008 | L-05 (0 occurrences of the sentence in each of the three copies; exit 1) | M4; `TestLaneReadCompactDoctrinePresent` green | Release-blocking; RED-now observed |
| AC-LRC-009 | REQ-LRC-009 | none (regression guard; baselines L-08 and L-09) | M6; the named tests green with non-empty swept sets | Regression guard |
| AC-LRC-010 | REQ-LRC-010 | L-07 (neither help text names its new flag; each exits 0) | M5; `TestLaneReadCompactHelpNamesFlags` green | Release-blocking; RED-now observed |
| AC-LRC-011 | quality gate G1 | not a behavior criterion | M6; `moai spec lint` exit 0 | Gate |

## C. Acceptance Criteria (Given-When-Then)

### AC-LRC-001 — todo default is unchanged in bytes and keys (REQ-LRC-001)

- Maps REQ-LRC-001.
- Given the frozen golden `internal/cli/testdata/ac_tst_012_golden.json` and the store snapshot it was captured from
- When `moai todo list --json` runs without flags
- Then stdout is byte-identical to the golden, and `TestTodoListJSON_GoldenByteIdentity` (`internal/cli/todo_json_golden_test.go:65`) passes without any edit to its asserted bytes.


### AC-LRC-002 — todo compact shape (REQ-LRC-002)

- Maps REQ-LRC-002.
- Given a backlog fixture of eleven items in this queue order: live, dropped, live, live, live, dropped, then five live. The nine live items are in the queued, picked, and hold states, both dropped items sit among the first six queue positions (positions 2 and 6), and the first item's text is fifty runes of the two-byte rune U+00E9
- When `moai todo list --json --compact --limit 2` runs
- Then the top-level keys are exactly `project_uuid`, `version`, `last_seq`, `live`, `withheld`, `items`, `findings_count`, `archived_count`, and `dropped_count`; `items` holds exactly six rows, which are the live items at queue positions 1, 3, 4, 5, 7, and 8 in that order, and no dropped row; the first row's `prefix` is the first 40 runes of its text followed by `...` (forty U+00E9 runes, then `...`); `live` is 9; `withheld` is 3; `dropped_count` is 2; each row carries exactly `id`, `state`, and `prefix`; and `--limit 2` has no effect.
- Test: `TestTodoListCompactShape` in `internal/cli/lane_read_compact_test.go`.


### AC-LRC-003 — todo compact byte bound (REQ-LRC-003)

- Maps REQ-LRC-003.
- Given a fixture of at least seven live items whose texts open with at least 40 `<` characters, which the encoder writes as 6-byte escapes, with twenty-byte identifiers — wider than the declared 16-byte cut, so each row's `id` renders truncated at the declared width (ids are ASCII under `cardIDPattern`, so raw and encoded width coincide) — state `picked`, and integers of ten digits (set through the fixture store)
- When `moai todo list --json --compact` runs
- Then stdout is at most 2048 bytes and carries six rows, each `id` rendered as its first 16 bytes.
- Test: `TestTodoListCompactByteBound` in `internal/cli/lane_read_compact_test.go`.
- Truncation cell (plan-audit iter2 repair; supersedes the 0.3.0 out-of-scope long-input cell): the criterion verifies the bound for any store content, including the adversarial shapes plan-audit iter1 observed (receipt `rcpt-9aaec51f93f1b122971c9d8a`, `.moai/reports/t1810/plan-audit-iter1.md`) — six live rows with 202-byte identifiers, which rendered 3,039 bytes uncut. The § C.2 worst-case table is shape-complete: with the cut, the table's cap-6 worst case is 2,011 bytes for any store content, whatever the identifiers, and the iter1 shape lands inside it — its six 202-byte identifiers render at 16 encoded bytes each after the cut, so those six rows render 3,039 − 6 × (202 − 16) = 1,923 bytes (both source figures are iter1's attributed observations; the subtraction is deterministic, the overlay's identifiers being ASCII). The truncated-identifier collision is accepted: two live rows whose identifiers share the declared 16-byte encoded prefix render identically in the compact form, and the full form `moai todo list --json` carries the untruncated identifiers and disambiguates. The store enforces no identifier length (`cardIDPattern`, `internal/homestate/card_evidence_readers.go:19`); adding store-level caps is explicitly out of this SPEC's scope, and none is needed for the bound.


### AC-LRC-004 — lane-facing reads name the compact form (REQ-LRC-004)

- Maps REQ-LRC-004.
- Given the five lane-facing surfaces: the three copies of `factory-dispatch.md` (under `.claude/rules/moai/workflow/`, `internal/template/templates/.claude/rules/moai/workflow/`, and `.moai/policies/workflow/`), `.claude/skills/moai/workflows/gtd.md`, and its template copy
- When the test scans every line that names `moai todo list --json` in those surfaces
- Then (a) each of the five surfaces carries at least one line naming `moai todo list --json --compact`; and (b) every line naming `moai todo list --json` without `--compact` is one of the four documented full-form lines of each `gtd.md` copy, and no other such line exists. The four lines, re-anchored 2026-10-19 at tree a3e8d1409c (the line numbers of tree b0a37f759 moved with main's drift: 210→218, 253→261, 254→262, 455→475): the record description (`gtd.md:218`, "`moai todo list --json` emits the file's records:"), whose JSON block `TestTodoListJSONShapeMatchesDoc` compares with the default output; the two `jq` shape examples (`gtd.md:261` and `:262`), whose subject is the top-level object of the default record; and the standing-source duplicate check (`gtd.md:475`, § Not stacking duplicates, whose heading is at `:470`), which compares card text that a compact row does not carry. The fifth naming line on this tree, the database-paths preference bullet (`gtd.md:47` in each copy, added by card t1827), is a read instruction and changes to name `--compact` (plan § D.4), so it is not a full-form line; and (c) the paragraph of § Not stacking duplicates that names the full form contains, in each `gtd.md` copy, the literal `40-rune prefix`, the reason plan § D.4 names. The test matches the first three lines by their text and the fourth by its section, so the compact paragraph that plan § D.4 inserts above them does not move the check. This reading is the one plan § D.4 follows.
- Test: `TestLaneQueueReadsDefaultCompact` in `internal/cli/lane_read_compact_test.go`.


### AC-LRC-005 — factory default is active-only, abbreviated, and capped (REQ-LRC-005)

- Maps REQ-LRC-005.
- Given a factory registry fixture with three runs (one `active` run holding nine cards, and two `retired` runs holding three cards each) plus one unavailable record; the first card of the active run leased to a holder of fifteen ASCII bytes followed by one three-byte rune (`lane-1234567890` then U+20AC, 18 bytes, so the 16-byte boundary falls inside the rune), and the second leased to `<<<` (3 raw bytes, 18 encoded bytes, because each `<` is a six-byte escape); and the quota gate enabled with pressure on, three live non-Claude lanes labelled `<<<`, `lane-b`, and `lane-c` (the registry's label order puts `<<<` first), and one live lane of unknown backend
- When `moai factory status --json` runs without flags
- Then every row's `run_id` is the active run's id; exactly three rows are rendered, in the store's order; each row carries exactly `run_id`, `card_id`, `state`, `stage`, `mode`, `priority`, `lease_holder`, and `lease_expired`, with `mode` and `priority` equal to what `--run <active id>` reports for the same card; the first two holders render as `lane-1234567890` and `<<`; `withheld` is 6; `unavailable_count` is 1; and the `quota` key carries `pressure: true`, the same two windows as the `--run <active id> --json` output for the same snapshot, the candidates `<<` (the label `<<<` cut to 12 encoded bytes, because a third escape would make 18) and `lane-b`, `candidates_withheld: 1`, and `unknown_lanes: 1`.
- Zero-candidate case: when the same command runs after the three non-Claude lanes are removed from the fixture registry, with the unknown-backend lane still live and pressure still on, the `quota` key carries `pressure: true`, `candidates: []`, `candidates_withheld: 0`, `warning: "no-non-claude-lane"`, and `unknown_lanes: 1`.
- Test: `TestFactoryStatusDefaultActiveOnly` in `internal/cli/lane_read_compact_test.go`.


### AC-LRC-006 — factory default byte bound under quota pressure (REQ-LRC-006)

- Maps REQ-LRC-006.
- Given the worst-case fixture: one `active` run holding ten cards, each with a `run_id` and a `card_id` of twenty characters each — wider than the declared 16-byte cut, so both render truncated at the declared width (ASCII, so raw and encoded width coincide) — state `needs-decision`, stage `merge-ready`, mode `parallelizable`, priority `normal`, and a lease holder of forty `<` characters (240 encoded bytes before the cut); five live non-Claude lanes with distinct labels of forty escapable characters each, namely thirty-nine `&` followed by one of `&`, `<`, `>`, U+2028, and U+2029, each of which the encoder writes as a six-byte escape (240 encoded bytes per label before the cut; distinct because `workers.label` is the registry's primary key, `internal/homestate/factory.go:25`); one unavailable record; and the quota gate enabled with both windows fresh, every optional window field set, and the five-hour window at its hold percentage, so quota pressure is on
- When `moai factory status --json` runs without flags
- Then stdout is at most 2048 bytes and carries three rows — each row's `run_id` and `card_id` rendered as their first 16 bytes — `withheld` 7, two candidates, each rendered as `&&` (the first two labels in label order, cut to 12 encoded bytes), and `candidates_withheld` 3; and a second check encodes, through the encoder the command uses, a report value assembled at the declared maxima of REQ-LRC-006 (counts of six digits, which no fixture of this size reaches) and asserts at most 2048 bytes. The planning figure for that report value is 1,953 bytes (plan § C.3, the iter2 recomputation at the identifier cut; L-13's 1,899 was measured at the 6- and 8-character identifier premise the cut supersedes).
- Test: `TestFactoryStatusDefaultBoundedBytes` in `internal/cli/lane_read_compact_test.go`.
- Truncation cell (plan-audit iter2 repair; supersedes the 0.3.0 out-of-scope long-input cell): the criterion verifies the bound for any store content, including the adversarial shape plan-audit iter1 observed (receipt `rcpt-9aaec51f93f1b122971c9d8a`, `.moai/reports/t1810/plan-audit-iter1.md`) — a factory store whose run identifier is set through the operator's `MOAI_KANBAN_ID` to 600 characters, which rendered 2,571 bytes uncut. The § C.3 worst-case table is shape-complete: with the cut, the table's cap-3 worst case is 1,953 bytes for any store content, whatever the identifiers, and the iter1 shape lands inside it — the three rows of the default carry the identifier once each, so the cut removes 3 × (600 − 16) = 1,752 bytes of the overlay's 2,571, leaving 819 bytes (both source figures are iter1's attributed observations; the subtraction is deterministic given that shape). The truncated-identifier collision is accepted: two cards whose identifiers share the declared 16-byte encoded prefix render identically in the default output, and the full forms (`--run <id>`, `--full`) carry the untruncated identifiers and disambiguate. `RecordRun` (`internal/homestate/runtime.go:28`) validates no run-identifier length and the store enforces no card-identifier cap (`cardIDPattern`, `internal/homestate/card_evidence_readers.go:19`); adding store-level caps is explicitly out of this SPEC's scope, and none is needed for the bound.


### AC-LRC-007 — `--run` keeps its pre-change output, and `--full` is byte-identical (REQ-LRC-007)

- Maps REQ-LRC-007.
- Given the factory fixture captured from the pre-change binary into `internal/cli/testdata/lane_read_compact/factory_full.golden.json`, committed in a baseline commit before the implementation commit (§ G, G4)
- When `moai factory status --json --full` runs against that fixture
- Then stdout is byte-identical to the golden; and `moai factory status --json --run <active id>` returns every pre-change card field with no cap.
- Tests: `TestFactoryStatusFullByteIdentical` and `TestFactoryStatusRunFlagUnchanged` in `internal/cli/lane_read_compact_test.go`. The existing callers of `status --json --run` (`internal/cli/factory_classify_test.go:418`, `internal/cli/factory_card_test.go:277` and `:670`, `internal/cli/factory_mirror_test.go:131`, `internal/cli/factory_quota_lanes_test.go:235`) stay green.


### AC-LRC-008 — the doctrine sentence in the three copies (REQ-LRC-008)

- Maps REQ-LRC-008.
- Given the three copies of `factory-dispatch.md`: the deployed rule `.claude/rules/moai/workflow/factory-dispatch.md`, its template copy `internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md`, and the shared-policy copy `.moai/policies/workflow/factory-dispatch.md`
- When `TestLaneReadCompactDoctrinePresent` reads each copy
- Then the sentence appears exactly once in each copy, and the three occurrences are byte-identical; in the deployed rule and in its template copy the occurrence lies inside a role-core region (after a role-core start marker and before its end marker); the shared-policy copy is checked for the single occurrence only; and `TestAutoPickMirrorParity` (`internal/cli/todo_auto_pick_doc_test.go:279`) passes.
- Why the role-core condition binds two copies only. A region is observable only where a reader extracts it. The SessionStart role injection reads the deployed rule, falls back to the shared-policy copy only when the deployed rule is absent (`internal/hook/role_rules.go:150-162`), and takes the embedded template copy's core whenever the file it reads carries no role-core marker (`internal/hook/role_rules.go:557-561`). The shared-policy copy carries none (`grep -c "role-core" .moai/policies/workflow/factory-dispatch.md` printed `0` at 9cf21f079; the file is unchanged since b0a37f759), and this SPEC adds none, so the 0.2.0 condition could never pass there, and a region in that copy would be read by nothing. The single occurrence is still required there, because `.moai/policies/` is the path from which Claude Code and Codex read the same policy (`internal/template/templates/.moai/policies/README.md`).
- Test: `TestLaneReadCompactDoctrinePresent` in `internal/template/lane_read_compact_doctrine_test.go`.


### AC-LRC-009 — frozen goldens, contract tests, and the mirror guard stay green (REQ-LRC-009)

- Maps REQ-LRC-009.
- Given the named tests of plan § E
- When `go test -count=1 -v -run '^(TestTodoListJSON_GoldenByteIdentity|TestTodoListJSONIgnoresLimit|TestTodoListJSONShapeMatchesDoc|TestFactoryStatusReportsLeaseKeys|TestSessionListDefaultNoProbe|TestAutoPickMirrorParity)$' ./internal/cli/` and `go test -count=1 -v -run '^TestRuleTemplateMirrorDrift$' ./internal/template/` run
- Then every named test passes, and each run has a non-empty swept set (the PASS lines are greater than zero).

### AC-LRC-010 — help text names the new flags (REQ-LRC-010)

- Maps REQ-LRC-010.
- Given the built binary
- When `moai todo list --help` and `moai factory status --help` run
- Then the todo help names `--compact` and the six-row cap, and the factory help names `--full` and the row cap of three.
- Test: `TestLaneReadCompactHelpNamesFlags` in `internal/cli/lane_read_compact_test.go`.

### AC-LRC-011 — the SPEC lints clean (quality gate G1)

- Given this SPEC directory
- When `moai spec lint SPEC-LANE-READ-COMPACT-001 --json` runs
- Then the exit status is 0 and the findings carry no error.

## D. Evidence Ledger (verbatim)

```
L-01 · RED (AC-LRC-002, AC-LRC-003)
command: python3 -I -c "from subprocess import run; r=run(['<BIN>','todo','list','--json','--compact'],capture_output=True); print(r.returncode, repr(r.stdout[:120]), repr(r.stderr[:100]))"
stdout (carrier print): 1 b'' b'          \n   ERROR  \n          \n  Unknown flag: --compact.                                         '
measured exit code: 1 (stderr carried "Unknown flag: --compact.")
tree: b0a37f759 · observed: 2026-10-15, this run
```

```
L-02 · regression guard (AC-LRC-001)
command: python3 -I -c "from subprocess import run; r=run(['<BIN>','todo','list','--json'],capture_output=True); print(r.returncode, len(r.stdout))"
stdout (carrier print): 0 2629774
measured exit code: 0
tree: b0a37f759 · observed: 2026-10-15, this run (the same command on the earlier snapshot printed 2626934)
```

```
L-03 · RED (AC-LRC-005, AC-LRC-006)
command: python3 -I -c "import json; from subprocess import run; r=run(['<BIN>','factory','status','--json'],capture_output=True); c=json.loads(r.stdout)['cards']; print(r.returncode, len(r.stdout), len(c), len(set(x['run_id'] for x in c)))"
stdout (carrier print): 0 117251 218 14
measured exit code: 0 (218 cards across 14 run ids; the earlier snapshot printed 117230 bytes and the same counts)
tree: b0a37f759 · observed: 2026-10-15, this run
```

```
L-04 · RED (AC-LRC-007)
command: python3 -I -c "from subprocess import run; r=run(['<BIN>','factory','status','--json','--full'],capture_output=True); print(r.returncode, repr(r.stdout[:120]), repr(r.stderr[:100]))"
stdout (carrier print): 1 b'' b'          \n   ERROR  \n          \n  Unknown flag: --full.                                            '
measured exit code: 1 (stderr carried "Unknown flag: --full.")
tree: b0a37f759 · observed: 2026-10-15, this run
```

```
L-05 · RED (AC-LRC-008): the queue-and-factory sentence is absent from the three copies
command: grep -c "Lanes read queue and factory state only through the quiet forms" .claude/rules/moai/workflow/factory-dispatch.md .moai/policies/workflow/factory-dispatch.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md
stdout: .claude/rules/moai/workflow/factory-dispatch.md:0
.moai/policies/workflow/factory-dispatch.md:0
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md:0
measured exit code: 1 (observed through the one-process carrier python3 -I -c "from subprocess import run; r=run(['grep','-c','Lanes read queue and factory state only through the quiet forms','.claude/rules/moai/workflow/factory-dispatch.md','.moai/policies/workflow/factory-dispatch.md','internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md'],capture_output=True); print(r.returncode, repr(r.stdout), repr(r.stderr))", which printed 1, the stdout above, and b'' for stderr)
red reason: the pattern is a literal of this SPEC's planned sentence, so the green-path run must print 1 for each copy. (The parent's L-07 pattern did not occur in its planned sentence.)
tree: 1e8a1e7b0 (code byte-identical to b0a37f759) · observed: 2026-10-15, about 12:32Z
```

```
L-06 · RED (AC-LRC-004)
command: grep -c "todo list --json --compact" .claude/skills/moai/workflows/gtd.md internal/template/templates/.claude/skills/moai/workflows/gtd.md .claude/rules/moai/workflow/factory-dispatch.md .moai/policies/workflow/factory-dispatch.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md
stdout: .claude/skills/moai/workflows/gtd.md:0
internal/template/templates/.claude/skills/moai/workflows/gtd.md:0
.claude/rules/moai/workflow/factory-dispatch.md:0
.moai/policies/workflow/factory-dispatch.md:0
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md:0
measured exit code: 1 (grep -c exits 1 when no line matches. Observed through the same one-process carrier as SPEC-LANE-READ-VERBS-001 ledger L-07, with this command's arguments, which printed 1, the stdout above, and b'' for stderr)
red reason: clause (a) of AC-LRC-004 fails in all five surfaces, and clause (c) fails in both gtd.md copies: python3 -I -c "from subprocess import run; r=run(['grep','-c','40-rune prefix','.claude/skills/moai/workflows/gtd.md','internal/template/templates/.claude/skills/moai/workflows/gtd.md'],capture_output=True); print(r.returncode, repr(r.stdout), repr(r.stderr))" printed 1 b'.claude/skills/moai/workflows/gtd.md:0\ninternal/template/templates/.claude/skills/moai/workflows/gtd.md:0\n' b'' at 2026-10-15T09:55:40Z (tree 9a5c2bebb). Clause (b) holds today: grep -n -F "moai todo list --json" over the same five files exits 0 and lists exactly lines 210, 253, 254, and 455 of each gtd.md copy, and nothing in the three factory-dispatch.md copies.
tree: b0a37f759 (the 0.2.0 cell covered three of the five surfaces); re-observed over all five at 9cf21f079, where the files are unchanged · observed: 2026-10-15T08:10:02Z
```

```
L-07 · RED (AC-LRC-010): the new flags are absent from the two help texts (carried from SPEC-LANE-READ-VERBS-001 ledger L-09, todo and factory lines)
command (todo): python3 -I -c "from subprocess import run; r=run(['<BIN>','todo','list','--help'],capture_output=True); print(r.returncode, '--compact' in r.stdout.decode(), '--full' in r.stdout.decode())"
stdout (carrier print): 0 False False
command (factory): python3 -I -c "from subprocess import run; r=run(['<BIN>','factory','status','--help'],capture_output=True); print(r.returncode, '--full' in r.stdout.decode(), '--json' in r.stdout.decode())"
stdout (carrier print): 0 False True
measured exit code: 0 for each help text
tree: b0a37f759 · observed: 2026-10-15, the parent's run
```

```
L-08 · baseline (AC-LRC-009): the named contract tests in internal/cli
command: python3 -I -c "from subprocess import run; r=run(['go','test','-count=1','-v','-run','^(TestTodoListJSON_GoldenByteIdentity|TestTodoListJSONIgnoresLimit|TestTodoListJSONShapeMatchesDoc|TestFactoryStatusReportsLeaseKeys|TestSessionListDefaultNoProbe|TestAutoPickMirrorParity)$','./internal/cli/'],capture_output=True); out=(r.stdout+r.stderr).decode(); print(r.returncode, 'PASS_lines', out.count('--- PASS'), 'FAIL_lines', out.count('--- FAIL'), 'no_tests', out.count('no tests to run')); print(out[-700:])"
stdout (carrier print, verbatim tail): 0 PASS_lines 15 FAIL_lines 0 no_tests 0 / "--- PASS: TestTodoListJSON_GoldenByteIdentity (16.89s)" / "--- PASS: TestTodoListJSONIgnoresLimit (12.75s)" / "PASS" / "ok  	github.com/modu-ai/moai-adk/internal/cli	74.769s"
measured exit code: 0
tree: b0a37f759 · observed: 2026-10-15, this run (the PASS lines include subtests of TestAutoPickMirrorParity)
```

```
L-09 · baseline (AC-LRC-009): the rule-template mirror guard
command: python3 -I -c "from subprocess import run; r=run(['go','test','-count=1','-v','-run','^TestRuleTemplateMirrorDrift$','./internal/template/'],capture_output=True); out=(r.stdout+r.stderr).decode(); print(r.returncode, 'PASS_lines', out.count('--- PASS'), 'FAIL_lines', out.count('--- FAIL'), 'no_tests', out.count('no tests to run')); print(out[-300:])"
stdout (carrier print, verbatim tail): 0 PASS_lines 51 FAIL_lines 0 no_tests 0 / "PASS" / "ok  	github.com/modu-ai/moai-adk/internal/template	0.746s"
measured exit code: 0
tree: b0a37f759 · observed: 2026-10-15, this run
```

```
L-10 · active-run observation (§ B.3 of plan; supports AC-LRC-005)
command: python3 -I -c "from subprocess import run; r=run(['<BIN>','factory','runs'],capture_output=True); print(r.returncode, [l.split()[0:2] for l in r.stdout.decode().splitlines() if ' active ' in l])"
stdout (carrier print): 0 [['tmvgvn', 'active']]
measured exit code: 0 (read-only classification; no --retire flag)
tree: b0a37f759 · observed: 2026-10-15, this run
```

```
L-11 · SUPERSEDED by L-12 (observation, not a release-blocking cell): worst-case byte emulation of the proposed shapes. Its todo envelope used 8-digit live and withheld counts, so its todo figures run 4 bytes low.
method: an encoder emulation (json.Marshal-compatible compact output for the todo shapes; json.Encoder-compatible two-space indent for the factory shape; both HTML-escaped with a trailing newline). The emulation reproduces the current outputs byte for byte: the todo default (2,626,934 bytes on the earlier snapshot) and the factory default (117,230 bytes on the earlier snapshot). The worst case uses the declared widths of plan § C.1. Superseded by L-12, which runs the same computation as the in-tree script `measure-caps.py`; the run-phase bound tests (AC-LRC-003, AC-LRC-006) remain the citable evidence for the caps.
todo --compact, worst case, cap: 5 → 1,669; 6 → 1,959; 7 → 2,249; 8 → 2,539; 10 → 3,119. CJK-only prefixes (3-byte runes, no escapes): cap 10 → 1,919; cap 11 → 2,089. Live queue (159 live items): cap 6 → 817; cap 10 → 1,273; cap 15 → 1,833; cap 20 → 2,355.
todo --compact, operator option (encoded-width prefix, not chosen): worst case cap 20 → 2,019; cap 21 → 2,109.
factory default, worst case, cap: 6 → 1,923; 7 → 2,123; 8 → 2,323; 10 → 2,723. Observed-width variant: cap 8 → 2,160. Real active run (68 cards): cap 6 → 1,692; cap 7 → 1,858; cap 8 → 2,039.
factory default, operator options (not chosen): run_id dropped from rows: cap 7 → 1,941; cap 8 → 2,115. Compact encoding: cap 10 → 1,939; cap 11 → 2,081.
measured exit code: not applicable (emulation)
tree: b0a37f759 · observed: 2026-10-15, this run
```

```
L-12 · measurement (Q2 and Q3; the command cited by decision-index.md and plan § C.1)
command: python3 -I .moai/specs/SPEC-LANE-READ-COMPACT-001/measure-caps.py <BIN>   (read-only; writes nothing)
stdout (verbatim):
== validation: encoder emulation vs live output (byte-exact)
  todo default    rc 0 bytes 2631808 exact True
  factory status  rc 0 bytes 117268 exact True
  session list    rc 0 bytes 43449 exact True
== todo --compact, live queue (163 live items)
  cap 6                                             817 bytes  PASS
  cap 7                                             917 bytes  PASS
  cap 8                                            1026 bytes  PASS
  cap 10                                           1273 bytes  PASS
  cap 15                                           1833 bytes  PASS
  cap 20                                           2355 bytes  FAIL
== todo --compact, worst case (declared widths; each prefix rune at 6 bytes)
  cap 5                                            1673 bytes  PASS
  cap 6                                            1963 bytes  PASS
  cap 7                                            2253 bytes  FAIL
  cap 8                                            2543 bytes  FAIL
  cap 10                                           3123 bytes  FAIL
== todo --compact, CJK-only prefixes (3 bytes per rune, no escapes)
  cap 10                                           1923 bytes  PASS
  cap 11                                           2093 bytes  FAIL
== todo --compact, option not chosen: prefix of at most 40 encoded bytes plus '...'
  cap 20                                           2023 bytes  PASS
  cap 21                                           2113 bytes  FAIL
== factory runs: active run ids ['tmvgvn'] (rc 0)
== factory default, worst case (declared widths; quota block as live)
  cap 6                                            1924 bytes  PASS
  cap 7                                            2124 bytes  FAIL
  cap 8                                            2324 bytes  FAIL
  cap 10                                           2724 bytes  FAIL
== factory default, live active run tmvgvn (68 cards)
  cap 6                                            1693 bytes  PASS
  cap 7                                            1859 bytes  PASS
  cap 8                                            2040 bytes  PASS
== factory options not chosen: B = run_id dropped from rows; C = compact encoding
  B cap 7 (indent)                                 1942 bytes  PASS
  B cap 8 (indent)                                 2116 bytes  FAIL
  C cap 8 (compact)                                1656 bytes  PASS
  C cap 9 (compact)                                1798 bytes  PASS
  C cap 10 (compact)                               1940 bytes  PASS
  C cap 11 (compact)                               2082 bytes  FAIL
== session list, live store: 135 entries {'ok': 135}
  session --all (every entry)                     43449 bytes  waived
  session default under S-2 (live entries)        43449 bytes  waived
measured exit code: 0 (the script's own exit status; read-only; it writes nothing)
tree: b0a37f759 · observed: 2026-10-15 between 05:35Z and 05:41Z, this run
```

```
L-13 · measurement (AC-LRC-005, AC-LRC-006; Q3): the factory default worst case with the quota block under pressure and with an escaped lease_holder, through the encoder of measure-caps.py
command (run from the worktree root; one invocation; measure-caps.py is loaded unchanged through runpy, its own output is captured, and only its validation lines are printed):
python3 -I -c "
import contextlib, copy, io, runpy, sys
sys.argv = ['measure-caps.py', '<BIN2>']
buf = io.StringIO()
with contextlib.redirect_stdout(buf):
    g = runpy.run_path('.moai/specs/SPEC-LANE-READ-COMPACT-001/measure-caps.py')
print(chr(10).join(buf.getvalue().splitlines()[:4]))
ib, cb, row, fac, W6, WROW, QL = g['indent_bytes'], g['compact_bytes'], g['row'], g['fac'], g['WIDE6'], g['WROW'], g['QUOTA']
def win(n):
    return {'window': n, 'state': 'fresh', 'used_percentage': 10 ** 24, 'resets_at': '9' * 20, 'captured_at': '9' * 20, 'holds_claude_lane': False, 'exhausted_at': '9' * 20}
def worst_quota(k):
    q = {'pressure': True, 'windows': [win('five_hour'), win('seven_day')], 'candidates': [{'label': 'h' * 16, 'backend': 'glm'}] * k, 'candidates_withheld': W6, 'unknown_lanes': W6}
    if k == 0:
        q['warning'] = 'no-non-claude-lane'
    return q
def live_pressure(labels, withheld=None):
    q = copy.deepcopy(QL)
    q['pressure'] = True
    q['candidates'] = [{'label': x, 'backend': 'glm'} for x in labels]
    if withheld is not None:
        q['candidates_withheld'] = withheld
    q['unknown_lanes'] = 0
    return q
def card(mp=True, runid=True):
    r = {'run_id': 'x' * 6} if runid else {}
    r.update({'card_id': 't' + '9' * 7, 'state': 'needs-decision', 'stage': 'merge-ready'})
    if mp:
        r.update({'mode': 'parallelizable', 'priority': 'normal'})
    r.update({'lease_holder': 'h' * 16, 'lease_expired': False})
    return r
def env(rows, q, withheld=W6, unavail=W6, skipped=W6, dur=('9' * 12, '9' * 12)):
    e = {'run': 'x' * 6, 'cards': rows, 'withheld': withheld, 'unavailable_count': unavail}
    if skipped:
        e['unavailable_skipped'] = skipped
    e['lease_duration'], e['renew_interval'] = dur
    e['quota'] = q
    return e
print('== 0.2.0 shape at cap 6 (six row fields; holder cut at 16 raw bytes), measured again')
row('quota as live, holder h*16', len(ib(env([WROW] * 6, QL))))
row('quota as live, holder <*16 (96 encoded bytes)', len(ib(env([dict(WROW, lease_holder='<' * 16)] * 6, QL))))
row('live windows, pressure on, 2 candidate lanes', len(ib(env([WROW] * 6, live_pressure(['lane-21', 'lane-22'])))))
print('== proposed shape: 8 row fields, holder and labels at 16 encoded bytes,')
print('   quota under pressure with every optional window field set, K candidates')
for cap, k in [(3, 2), (4, 2), (3, 3), (3, 4)]:
    row(f'cap {cap}, K {k}', len(ib(env([card()] * cap, worst_quota(k)))))
print('== alternatives not chosen (same worst case)')
for cap in (4, 5):
    row(f'A1 candidates as a count only (K 0), cap {cap}', len(ib(env([card()] * cap, worst_quota(0)))))
for cap in (4, 5):
    row(f'A2 rows without mode, priority, cap {cap}', len(ib(env([card(mp=False)] * cap, worst_quota(2)))))
for cap in (3, 4):
    row(f'B rows without run_id, cap {cap}', len(ib(env([card(runid=False)] * cap, worst_quota(2)))))
for cap in (6, 7):
    row(f'C compact encoding, cap {cap}', len(cb(env([card()] * cap, worst_quota(2)))))
active = g['active']
act = [c for c in fac['cards'] if c['run_id'] == active[0]]
real = [{k: c[k] for k in ['run_id', 'card_id', 'state', 'stage', 'mode', 'priority', 'lease_holder', 'lease_expired']} for c in act[:3]]
live = (len(act) - 3, len(fac['unavailable']), 0, (fac['lease_duration'], fac['renew_interval']))
print(f'== proposed shape on the live active run {active[0]} ({len(act)} cards), cap 3')
row('quota as live', len(ib(env(real, QL, *live))))
row('live windows, pressure on, 2 candidate lanes', len(ib(env(real, live_pressure(['lane-21', 'lane-22'], 0), *live))))
"
stdout (verbatim):
== validation: encoder emulation vs live output (byte-exact)
  todo default    rc 0 bytes 2635270 exact True
  factory status  rc 0 bytes 117305 exact True
  session list    rc 0 bytes 46423 exact True
== 0.2.0 shape at cap 6 (six row fields; holder cut at 16 raw bytes), measured again
  quota as live, holder h*16                       1924 bytes  PASS
  quota as live, holder <*16 (96 encoded bytes)    2404 bytes  FAIL
  live windows, pressure on, 2 candidate lanes     2113 bytes  FAIL
== proposed shape: 8 row fields, holder and labels at 16 encoded bytes,
   quota under pressure with every optional window field set, K candidates
  cap 3, K 2                                       1899 bytes  PASS
  cap 4, K 2                                       2161 bytes  FAIL
  cap 3, K 3                                       1978 bytes  PASS
  cap 3, K 4                                       2057 bytes  FAIL
== alternatives not chosen (same worst case)
  A1 candidates as a count only (K 0), cap 4       2036 bytes  PASS
  A1 candidates as a count only (K 0), cap 5       2298 bytes  FAIL
  A2 rows without mode, priority, cap 4            1921 bytes  PASS
  A2 rows without mode, priority, cap 5            2123 bytes  FAIL
  B rows without run_id, cap 3                     1821 bytes  PASS
  B rows without run_id, cap 4                     2057 bytes  FAIL
  C compact encoding, cap 6                        1926 bytes  PASS
  C compact encoding, cap 7                        2114 bytes  FAIL
== proposed shape on the live active run tmvgvn (68 cards), cap 3
  quota as live                                    1338 bytes  PASS
  live windows, pressure on, 2 candidate lanes     1557 bytes  PASS
measured exit code: 0 (read-only: the only commands it runs are the four read verbs of measure-caps.py)
worst-case widths: run_id 6 characters; card_id 8; state needs-decision (14, the longest of the 21 card states, internal/homestate/card_record.go:59-64 at the observation tree; :62-67 after the 2026-10-19 re-anchor); stage merge-ready (11); mode parallelizable (14) and priority normal (6), internal/factory/classification.go:34-39; lease_holder and candidate labels at 16 encoded bytes; counts of 6 digits; durations of 12 characters; run at 6 characters (an over-count, because the default leaves run empty); both windows fresh with every optional field set, used_percentage at the 25-byte longest float64 encoding (emulated by a 25-digit integer), three 20-character timestamps, and holds_claude_lane false; candidates_withheld and unknown_lanes at 6 digits. The live rows copy the live holders, which are at most 7 bytes, so no cut applies to them.
binary: <BIN2> (see SPEC-LANE-READ-VERBS-001 ledger L-06)
tree: 9cf21f079 (code byte-identical to b0a37f759) · observed: 2026-10-15, run ending 08:07:12Z
```

```
L-14 · premise re-measured 2026-10-19 (AC-LRC-001; refreshes the operative figure of L-02): the full todo default
command: the L-02 command form (`<BIN> todo list --json`, stdout byte count)
observed: 2,897,435 bytes (L-02 carried 2,629,774 on 2026-10-15; about +10% in four days)
tree: 74ad197380 · observed: 2026-10-19 by the card lane under the resource rules, carried into this ledger through the card dispatch record. Gap: the command's carrier print and exit code are not carried in that record, so this row cites the byte count only; the L-02 verbatim carrier print remains the pinned historical form.
```

```
L-15 · premise re-measured 2026-10-19 (AC-LRC-005, AC-LRC-006; refreshes the operative figure of L-03): the factory status default
command: the L-03 command form (`<BIN> factory status --json`; stdout byte count, card count, distinct run ids)
observed: 201,400 bytes · 368 cards · 15 run ids (L-03 carried 117,251 · 218 · 14 on 2026-10-15; about +72% in bytes in four days)
tree: 74ad197380 · observed: 2026-10-19 by the card lane under the resource rules, carried into this ledger through the card dispatch record. Gap: the command's carrier print and exit code are not carried in that record, so this row cites the counts only; the L-03 verbatim carrier print remains the pinned historical form.
```

```
L-16 · premise re-measured 2026-10-19 (AC-LRC-002, AC-LRC-003, AC-LRC-007; re-asserts the L-01 and L-04 RED premises): the new flags are still absent
observed: `moai todo list --json --compact` and `moai factory status --json --full` each still refuse with `Unknown flag` — the L-01 and L-04 red reason holds on today's tree
tree: 74ad197380 · observed: 2026-10-19 by the card lane under the resource rules, carried into this ledger through the card dispatch record. Gap: the verbatim refusal text of this run is not carried in that record; the L-01 and L-04 verbatim outputs remain the pinned forms.
```

## E. Edge Cases

- Empty queue: `todo list --json --compact` returns `items` as an empty array, `live` 0, `withheld` 0, and exit 0.
- Queue holding only dropped items: `--compact` returns an empty `items` array and `dropped_count` equal to the dropped total.
- No active factory run: the default returns `cards` as an empty array, `withheld` 0, and exit 0.
- `--full` with `--run <id>`: the output is the pre-change output for that run.

## G. Quality Gates

- G1 — `moai spec lint SPEC-LANE-READ-COMPACT-001 --json` exits 0.
- G3 — run-phase: `gofmt -l` and `go vet` report nothing for the touched packages; the new tests pass; coverage of the touched packages meets `.moai/config/sections/quality.yaml`.
- G4 — run-phase: `factory_full.golden.json` is committed before the implementation commit.
- G5 — run-phase: each commit stages by explicit pathspec.

## H. Definition of Done

- Decisions: Q1 is DECIDED on board record dfd8, and Q2, Q3, Q7, Q8, and Q9 on board record 6f23, each pinned in `decision-index.md`. Q6 has its measured result in the progress record before M4 closes.
- AC-LRC-001 to AC-LRC-011 each carry a green-path cell with its verbatim output, and each test-based criterion carries the RED output captured before its implementation (E8).
- G1, G3, G4, and G5 pass.
