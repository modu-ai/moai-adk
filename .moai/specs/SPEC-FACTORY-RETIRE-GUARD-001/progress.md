# SPEC-FACTORY-RETIRE-GUARD-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
audit_ready: true
- plan_complete_at: 2026-10-17T13:46:00Z
- Authoring: spec.md, plan.md, acceptance.md, decision-index.md written by manager-spec (Tier M, card t1852, Class C; part C only of the card line t1825 -> t1848). Tree `d2d9ac08fb1899440da911121ddf13eecc806a3f`, branch `WT-factory-retire-guard`. Not yet audited.
- Counts: 7 requirements (Tier M ceiling 16), 8 acceptance criteria (ceiling 16), 2 milestones (M1, M2).
- Tier choice: M. Tier S would fit the production LOC but not the file count (5 production files, 2 new test files, 8 sync-phase documents) and carries a plan-audit ceiling of 1 (`harness.plan_audit_tier_ceilings` S: 1, M: 2), which leaves no repair round.
- SPEC ID self-check, executed as Bash: `ID="SPEC-FACTORY-RETIRE-GUARD-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` printed `PASS`; `ls .moai/specs | grep -i -E "RETIRE-GUARD|RETIRE-TRANSFER|CARD-TRANSFER|RUN-SELECT|LEADER-ADOPT"` printed nothing (the ID is unique at this tree).
- Lint, two measurements: `moai spec lint --strict .moai/specs/SPEC-FACTORY-RETIRE-GUARD-001/spec.md` under the installed build (`moai version`: `v3.2.0-rc.36-t1822c`, commit `ee981a932`; `git merge-base --is-ancestor ee981a932 HEAD` exit 0, so a strict ancestor of the tree's HEAD) printed `✓ No findings — all SPEC documents are valid`. A binary built from this tree (`go build -o <scratchpad>/moai-tree ./cmd/moai` at HEAD `d2d9ac08f`, invoked by path; its version line reads `v3.1.3 none built unknown` because no ldflags were passed) printed the same line. The tree-built result is the one that is a measurement by a build of this tree (verification-claim-integrity §2.2); the lint ran on spec.md only, not on the other four files.
- Id consistency: `grep -ohE "REQ-RRG-[0-9]+"` over spec.md, plan.md and acceptance.md lists REQ-RRG-001..007 with no gap and no id outside that range; the same grep for `AC-RRG-[0-9]+` over plan.md and acceptance.md lists AC-RRG-001..008; `grep -n -E "RSC|RUNSEL"` over the five files printed nothing (no stale id prefix of the sibling SPECs); the sibling SPEC names appear only as prose citations in spec.md §F.
- RED-now evidence: `acceptance.md` ledger E1, E2 (empty sweeps, 5 + 4 named tests, exit 0), P1-P5 (behavioral probes observing the old behavior directly; probe sources deleted, never committed), E3 and E4 (existing guards GREEN). No behavioral RED of the named tests was observed (they do not exist yet).
- Decision rows: Q1-Q5 `DEFAULT-APPLIED` (all implementation-level); none open. Q3 carries a classification note: if the operator reads the added informational block as a change of a shipped command's default user-visible behavior, it becomes product-level.
- Inherited closed decisions (spec.md §F) and the code facts re-read at this tree are listed in the plan-phase return report; none was re-argued.
- Dispatch-fixed departure from the sibling SPEC: operator-facing prose of the `--retire` surface is Korean (plan §E); the sibling kept it English.

## Plan audit history

- Iteration 1: FAIL 0.88 (`.moai/reports/t1852/plan-audit.md`, audited commit `b26050d9b`; 3 blocking: D1 false "the one exit that still works" claim and an unscoped revival statement, D2 cards-before-owner ordering with a false "non-destructive exits" sentence, D3 abandon caveats missing from REQ-RRG-005, the stranded block and AC-RRG-006). Tier M ceiling 2, so this is the one repair round and the next audit is the last.
- Repair (version stays `0.1.0`; REQ 001-007 and AC 001-008 id sets unchanged), inside the audit's fix_scope: D1 spec REQ-RRG-005 and §C.7 (exclusivity removed; `--factory-run` statement scoped to that flag; revival routes `RecordRun` and `ResumeRun` recorded as a factual note and a residual risk, offered as no exit, no new requirement), plan §E stranded exit/note, AC-RRG-006; D2 plan §C.2 (order inside `retireRun` is status -> stamp -> cards -> flip -> event; the false ordering sentence replaced), AC-RRG-003 row (b) and mutant probes, Edge-cases third bullet; D3 REQ-RRG-005, plan §E (new stranded abandon-note row), AC-RRG-006 (abandon note asserted, mutant probe).
- Out-of-anchor changes, each the smallest edit that keeps a sentence true: `decision-index.md` Q3 Default (the phrase "the one exit that works for a non-active run" was the same false exclusivity claim). No other out-of-anchor edit. O1, O2 (AC-002, REQ-007) not taken; O3 (English error text duplicating the card list) not taken because the fix is the `Error()` sentence in plan §C.1, outside the anchors.
- Sentences re-checked for consistency with the new order and left unchanged: REQ-RRG-001 and REQ-RRG-002 (the owner gate still precedes the card read; "no row changed" stays true because the stamp check refuses before the flip), AC-RRG-002 (the hook still fires right after the in-transaction `SELECT`, after the stamp check), plan §C.1 (hook placement), plan §C.3 (the printed classification now always belongs to a still-unchanged owner).
- Repair-time lint: `moai spec lint --strict` on spec.md under the installed build and the tree-built binary, see the repair commit.
- Iteration 2: PASS 0.94 (.moai/reports/t1852/plan-audit-iter2.md, audited commit 9d171631a, plan_artifact_hash 0e68a9981361c7a407aa050f25c1614940ffd9cd07b668c04e2b17d1fd2e5110, codex receipt rcpt-dfdb489c540acf07a44604ff).
- Post-PASS edit (sync phase): acceptance.md was edited after the plan-audit PASS (audited commit 9d171631a, plan_artifact_hash 0e68a998…5110) only to add eight reserved `[REF]` tokens after the shorthand AC ids on the Requirement-to-criterion map line, for the CHANGELOG AC counter (live=8 excluded=8 ambiguous=0); no semantic change. The factory plan-hash check ran at T7/T8a before this edit.

## §E.2 Run-phase Evidence

Tree base fe761c8e9; run commits M1 f766f0b03, M2 5b03422ac (branch WT-factory-retire-guard). Env scrubbed of MOAI_KANBAN_*/MOAI_FACTORY_* per run. RED = first run of the named tests before the production change (M1: tests compiled against the new `factory_run_cards.go` seam, old `retireRun`; M2: tests against old CLI). Raw outputs are machine-local scratch; the deciding lines are quoted.

| AC | Test | RED (observed) | GREEN (observed) | Commit |
|---|---|---|---|---|
| AC-RRG-001 | TestRetireRunIfDeadRefusesInProgressCards / ...RetiresWhenOnlyTerminalCardsOrWorkers | `RetireRunIfDead retired a run holding non-terminal cards (class "dead")`; control test PASS by design (retires today) | PASS both | f766f0b03 |
| AC-RRG-002 | TestRetireRunIfDeadCardReadHoldsWriteLock | `card-read hook calls = 0, want 1`; mutant probe (card read via pool handle before BeginTx, hook with it): `second handle's insert committed while the card read ran: the read is outside the write-locked transaction`, then placement restored | PASS | f766f0b03 |
| AC-RRG-003 | TestRetireRunIfDeadFlipIsCASOnClassifiedStamp | 3 of 4 rows FAIL: `err = <nil>, want ErrRunOwnerNotDead stating the owner or status changed` (rows a, b), `err = <nil>, want an error wrapping ErrRunOwnerNotDead` (row c); row d (peer basis) PASS as regression | PASS 4 rows | f766f0b03 |
| AC-RRG-004 | TestRetireRunIfDeadAlreadyRetiredIsNoOp | `a no-op wrote:` (second event today); `err = <nil>, want the cards refusal` | PASS 3 subtests | f766f0b03 |
| AC-RRG-005 | TestFactoryRunsRetireRefusesInProgressCards | operator and lane subtests FAIL: output lacked the refusal header/exits/abandon note; no Hangul in output | PASS operator + lane | 5b03422ac |
| AC-RRG-006 | TestFactoryRunsRetireReportsStrandedCardsOfRetiredRun | subtest A FAIL: output lacked stranded header, cards, exits, notes; B and C PASS by design (first line only) | PASS A, B, C | 5b03422ac |
| AC-RRG-007 | TestRelaunchFromRunLeavesRunWithCardsActive | `outcome = "run runD: could not be retired (factory run holds in-progress cards: ...), left untouched"` | PASS | 5b03422ac |
| AC-RRG-008 | TestFactoryRunsHasNoForceFlag | regression guard, PASS before and after by design | PASS | 5b03422ac |

Existing guards GREEN after M1 and M2: TestRetireRunIfDeadRefusesLiveAndIndeterminate, TestEveryRetirementPathDeclinesUnenumeratedClassification, TestRetireRunIfDeadAcceptsBootProof, TestRetireRunPreservesRowAndAppendsEvent, TestFactoryRunsCommandReportsAndRefuses, TestFactoryRelaunchFromRunRetiresDeadOwnerOnly.

Cross-cutting (this tree, HEAD 5b03422ac, slot go-test-heavy held 15:30:11Z-released): `go vet ./internal/homestate ./internal/cli` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `gofmt -l` empty; `go test -cover ./internal/homestate` ok 81.9% (package, full); changed-function coverage: inProgressCards 66.7%, retireRun 76.7%, RetireRunIfDead 92.9%, InProgressCards 100%, text renderers 100%, newFactoryRunsCommand 88.2%, relaunchRetireFromRun 78.3% (the uncovered branches are error returns of DB reads). golangci-lint on ./internal/homestate/... ./internal/cli/...: exit 1 with 3 errcheck findings, all in untouched `internal/cli/factory_run_adopt.go:79,91,103` (pre-existing, none in touched files). No `\u` in the catalogue.

## §E.3 Run-phase Audit-Ready Signal

audit_ready: true
run_status: complete
run_commit_sha: 5b03422ac
ac_pass_count: 8
ac_fail_count: 0
cross_platform_build.windows_amd64: exit 0
m1_to_mN_commit_strategy: per-milestone (M1 f766f0b03, M2 5b03422ac) + this progress commit
gaps: sync-phase docs (8 files) not edited by design; new-code function coverage below 85% on error-return branches; pre-existing errcheck findings in factory_run_adopt.go; full internal/cli suite not run (CI owns it); the `retireAfterCardRead` hook in internal/homestate has no @MX:ANCHOR (fan_in 1)

## §E.4 Sync-phase Audit-Ready Signal

audit_ready: true
sync_status: complete
sync_complete_at: 2026-10-18
docs_commit_sha: c77e723c1
sync_commit_sha: 6029b3e9ae68c0762158bf28ff8398190b9f7819
ac_source: .moai/specs/SPEC-FACTORY-RETIRE-GUARD-001/acceptance.md
tier: M
ac_count: 8
changelog_entry_position: "[Unreleased] ### Added, first entry"
frontmatter_status_transitions: in-progress -> implemented -> completed (spec.md, single sync commit)

Docs obligation (spec §H/plan §I, 8 files) is done and verified; the CHANGELOG entry and the status transition were first held by a B12 self-test blocker (below, resolved by manager-spec commit d62505c83) and landed in the sync commit. The real SHA of that sync commit replaces `pending-backfill` in a following backfill commit.

- Docs edits (commit c77e723c1, tree base 05253e87a): the `--retire` sentence in README.ko.md (canonical), README.md, README.ja.md, README.zh.md and docs-site/content/{ko,en,ja,zh}/advanced/factory-mode.md. Each states, checked against `internal/cli/factory_retire_guard_text.go`, `factory_handoff_recover.go`, `factory_relaunch.go` and `internal/homestate/factory_run_retire.go`: refusal on non-terminal cards (not done/failed/abandoned) with the card list and the exits (`moai cc -l --factory-run`, `moai cc -f --factory-run`, `moai factory decide <card> --choice abandon --run`), join/adopt only while the run is active, abandon irreversible, no force flag, an already-retired run reported without a second write with its stranded cards listed, `relaunch --from-run` leaves an active run holding such cards active (`left active (N in-progress cards)`) and an already-retired run untouched (`not active (retired), left untouched`). Corrected by the commit that follows sync audit iteration 1 finding F1/F5; the first docs commit c77e723c1 had worded the relaunch clause as applying to retired runs too.
- Verify recipe (this tree, after the edits): URL blacklist grep over docs-site/content and README*.md exit 1 (no match); Mermaid LR/RL grep exit 1 (no match); README `grep -c '^## '` 12/12/12/12; factory-mode.md heading count `^#{2,} ` 10/10/10/10 (ko/en/ja/zh); file-existence loop printed no MISSING; locale-parity ratchet `comm -23` printed no NEW divergence; emoji scan limited to the added lines printed no hit and no `\u` escape in the added lines; `hugo --minify --gc` run with `--destination` in the scratchpad exit 0, 0 lines matching warn/error, sitemap.xml present. `moai slot status --resource go-test-heavy` read first: held by lane-6, not acquired (hugo is not a go-test-heavy workload).
- `moai spec lint --strict` on spec.md: `No findings — all SPEC documents are valid`.
- B12 CHANGELOG blocker: precheck `grep -c 'SPEC-FACTORY-RETIRE-GUARD-001' CHANGELOG.md` printed 0 (no duplicate). The AC count self-test (tier M, ac_source=.moai/specs/SPEC-FACTORY-RETIRE-GUARD-001/acceptance.md, file non-empty) printed `16` with `live=16 excluded=0 ambiguous=0`, against 8 declared criteria AC-RRG-001..008. The extra 8 are the shorthand ids `AC-001..AC-008` on the single line "Requirement-to-criterion map" (acceptance.md line 104), which the grammar counts as separate identifiers. A CHANGELOG count of 8 would not equal the counter, and writing 16 would be a false figure, so no CHANGELOG entry was written. Resolution: manager-spec adds `[REF]` immediately after each of the eight shorthand occurrences on that line (acceptance.md body, outside manager-docs' ownership), after which the counter prints 8 and the entry can be emitted with the status transition.
- B12 resolution (HEAD d62505c83): manager-spec added the eight `[REF]` tokens on acceptance.md line 104. Re-run exactly as B12 prescribes with `AC_FILE=.moai/specs/SPEC-FACTORY-RETIRE-GUARD-001/acceptance.md` (tier M, file non-empty): stdout `8`, stderr `live=8 excluded=8 ambiguous=0`, exit 0. Pre-emission `grep -c 'SPEC-FACTORY-RETIRE-GUARD-001' CHANGELOG.md` printed `0` again; the CHANGELOG entry carries 8 acceptance criteria AC-RRG-001..008. Paths named in the entry were checked with `ls`: internal/homestate/factory_run_retire.go, factory_run_cards.go, internal/cli/factory_handoff_recover.go, factory_retire_guard_text.go, factory_relaunch.go, docs-site/content/ko/advanced/factory-mode.md, README.ko.md all exist.
- Gaps (the three CHANGELOG/status items above are closed; the following remain): ja/zh/ko prose not read by a native human reviewer (no humanize skill pass; the optional pack was not invoked); installed `moai` build not re-compared with the tree HEAD for this lint run (spec.md unchanged since the plan-phase lint); the docs claims about exits are read from code, `decide`/`cc` were not run.

## §G Override and Refusal Record

- 2026-10-17T14:10:57Z SPEC-FACTORY-RETIRE-GUARD-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 2 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1852/.moai/reports/t1852/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1852/.moai/reports/t1852/plan-audit.md
- 2026-10-17T14:11:36Z SPEC-FACTORY-RETIRE-GUARD-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 2 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1852/.moai/reports/t1852/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1852/.moai/reports/t1852/plan-audit.md
- 2026-10-17T14:12:52Z SPEC-FACTORY-RETIRE-GUARD-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 2 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1852/.moai/reports/t1852/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1852/.moai/reports/t1852/plan-audit.md
- 2026-10-17T14:13:46Z SPEC-FACTORY-RETIRE-GUARD-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 2 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1852/.moai/reports/t1852/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1852/.moai/reports/t1852/plan-audit.md
