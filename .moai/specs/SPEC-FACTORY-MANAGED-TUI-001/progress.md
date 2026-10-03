# progress.md — SPEC-FACTORY-MANAGED-TUI-001

> Phase record. Only `§E.1` is filled by the plan phase (manager-spec); `§E.2` and `§E.3` belong to the run phase (manager-develop), `§E.4` to the sync phase (manager-docs).

## §E.1 Plan-phase Audit-Ready Signal

plan_status: PASS-WITH-DEBT by the leader's convergence rule (0.3.1; final delta audit FAIL 0.875 on two wording blockers, fixed without a new audit)
plan_complete_at: 2026-10-03
artifacts: spec.md (REQ 14) · plan.md · acceptance.md (AC 16) · design.md (D-1..D-10) · progress.md
tier: M (see plan.md §A; LOC total at the ceiling, tier-up path stated)
measured_tree: 2b9e4a4d0
open_clarifications: 0 NEEDS-CLARIFICATION markers; 6 leader/operator questions in the plan-phase report
card: t1408

### Ownership boundaries (leader constraint, t1440 lane-17)

Edits only in `managed_codex_factory.go`, `managed_factory_session.go`, new `managed_codex_tui.go` / `managed_codex_tui_test.go`, the fixture `testdata/codex-0.160.0/resume-help.txt`, `internal/config/defaults.go`, `envkeys.go`, one minimal separate bullet edit in the operator document, and CHANGELOG. Not touched: `codex_launcher.go`, `managed_operator_input.go`, `managed_card_child_test.go`, `store.go`.

### Plan-audit iteration 1 and its disposition (iteration 2 revision, version 0.2.0)

Iteration 1: FAIL 0.63 at `7287e64f3` (local report `.moai/reports/t1408/plan-audit-iter1.md`, not committed).

| Defect | Disposition |
|---|---|
| D1 REQ-MT-008 vs AC-MT-007(iii) | Fixed: REQ-MT-008 now carries the outstanding-`turn/start` exception; AC-MT-007 (iii) and known debt 9 use the same wording. |
| D2 false no-overlap with t1459 | Fixed: design.md D-9 rewritten from the t1459 draft (leading `ctx`, `ctx.Done()` in the select, interruption check, 14 occurrences in 6 files here vs 10 in 5 there); what each landing order owes is listed; "either order works" replaced by verified vs unverified. spec.md §D reworded. |
| D3 M1 not compile-able RED | Fixed: plan.md M1 is stub commit S plus RED commit R; RED = named failing assertion per AC; REQ-MT-014 and AC-MT-014 amended. |
| D4 M2 gate needs M4 | Fixed: minimal TUI wait/stop path moved into M2; M4 keeps status mapping, server-death monitor, blocked-call release; `Exit:` lines added for every milestone. |
| D5 AC-MT-010 nonexistent test | Fixed: AC-MT-010 is the four existing tests only; `TestManagedTUINeverReachedWithoutOptIn` is the new AC-MT-016. |
| D6 missing RED-now cells | Reclassified, not faked: test-based ACs are adoption-deferred; M2 cannot start until a delta check confirms every cell in `red-baseline.md` (acceptance.md header and §1.2). |
| D7 EXCL-syscall | Fixed: clause added to spec.md C.4, §D, §E and the pseudo-terminal exclusion (wording only). |
| D8 log-sink lifetime | Fixed: sink cleared at TUI reap; teardown lines go to the terminal; REQ-MT-006, D-6, D-8, AC-MT-005 (`post_reap_line_on_terminal`) agree. |
| D9 stderr after TUI start failure | Fixed: REQ-MT-004 exception; D-6; AC-MT-003 `tui_start_fails` checks it. |
| D10 stale-busy steering | Fixed by changing the rule: busy never clears by time, warn lines only; known debt 13; AC-MT-006 `TestManagedCodexBusyLongTurnKeepsDeferring`; mutant mu16. |
| D11 timing seams | Fixed: D-10 declares three overridable durations; D-7 amended. |
| D12 AC-MT-013 | Fixed: base counts for every row, same-line anchors, CHANGELOG rows, bare `t1459` row replaced, mutant mu17. |
| D13 "resolved" | Fixed: REQ-MT-013 and the DoD make "resolved" conditional on AC-MT-015 steps 1, 2, 3, 6. |
| D14 blocked `DeliverTurn` | Fixed: AC-MT-008 subtests `blocked_turn_start_released`, `blocked_wait_turn_released`; mutant mu15. |
| D15 §F vs D-9 | Fixed: §F says one line in `Start`; D-9 table gains the TUI-exit channel and the monitor goroutine rows. |
| D16 P10 undercount | Fixed: P10 names the three Codex writers and the out-of-scope `:179`. |
| D17 bundled REQ-MT-007, cross-refs | Cross-references fixed (D-8, §A.1); REQ-MT-007 not split (REQ count and AC traceability stay stable; optional). |
| D18 unnamed subtests, P6-false, AC-MT-014 placeholders | Fixed: AC-MT-003 subtests named; `frames_not_delivered` added; AC-MT-014 states one row per fix commit. |
| D19 ownership list drift | Fixed: CHANGELOG and the help-text fixture in plan.md and here. |

### Plan-audit iteration 2 (FAIL 0.86) and the delta revision 0.3.0 (iteration 3 = final delta audit)

Local report `.moai/reports/t1408/plan-audit-iter2.md` (not committed). Leader decisions baked in: AC-MT-015 is operator-held (keep-set), not release-blocking, recorded as "operator confirmation pending, no terminal designated" in the sync-phase deliverable; t1408 lands before t1459 (decided, no hedging); the opt-out `MOAI_FACTORY_MANAGED_TUI` stays, default ON inside the managed gate.

| Defect | Disposition |
|---|---|
| D1 blocked write not releasable by a channel | Fixed: design.md D-8 "Releasing a blocked call" — the wait goroutine closes the WebSocket connection on TUI exit; a per-write deadline is the secondary device; `Close` stays single-goroutine; AC-MT-008 `blocked_turn_start_released` blocks the write and proves it. D-9 row and plan.md M4 follow. |
| D2 CN-4 CONFLICT from the H1 | Fixed: the AC range is removed from the acceptance.md H1. Exit/AC ordering re-read by hand: M1 exit AC-MT-014 (M1 part), M2 001-005 and 016, M3 006-007, M4 008-009, M5 010-012 (and the closing part of 014), M6 013; no acceptance clause orders a criterion across those exits. |
| D3 M1 gate unmechanized | Fixed: the orchestrator or the M1 lane runs the stated three-line command (acceptance.md §1.1) and records it in §E.2; AC-MT-014 split into an M1-evaluable part and a closing part re-run at M5. |
| D4 AC-MT-014 claims more than it checks | Fixed: 11 rows, 11 or more `--- FAIL: Test` lines, zero `undefined:`/`build failed`; "tree SHA" is S's commit SHA with the R tests applied. |
| D5 E1/E4 flips | Fixed: E1 gets `--exclude='*_test.go'`; E4 is noted to flip at S. |
| D7 plan.md §F | Fixed: "AC-MT-001..014 and 016 (015 operator-held)". |
| D6 and other optional | Not touched (D6 residual stays a Gap: the sync audit reads the anchored lines). |

### Final delta audit (FAIL 0.875) and 0.3.1

Local report `.moai/reports/t1408/plan-audit-iter3-delta.md` (not committed). Fixed in 0.3.1: B1 (`blocked_turn_start_released` now has an entered-and-not-returned barrier, a write deadline longer than the watchdog, a queued second sender, and mutant mu19 for a select-only release); B2 (the M1 delta check counts 11 distinct AC ids and 15 distinct named failing tests, excludes `panic:` and `test timed out`, and states that only re-running one cited RED command at the M2 advance stops a fabricated cell); O1 (plan.md Exit lines split AC-MT-014 into M1 and M5 parts); O2 (TUI status recorded before the connection closes; blocked subtests use exit code 7); O8 (spec.md §E range). Left as debt: the other optional findings of that report, and D6 (AC-MT-013 stays satisfiable by anchored search tokens; the sync audit reads the anchored lines).

## §E.2 Run-phase Evidence

Tree: branch `WT-managed-codex-tui-attach`, base `2b9e4a4d0` (merge-base with develop, read in this run). Toolchain: go1.26.8 darwin/arm64, golangci-lint v2.1.6 (equals the CI pin), codex-cli 0.160.0 (probe fixture only; no automated test starts a real codex). The lane environment was scrubbed in every `go test` (`unset MOAI_KANBAN ... && go test ...`). Raw logs are local scratch and are not cited; the deciding lines are quoted here.

### Commits (sha, subject)

| sha | subject |
|---|---|
| `d222d310e` | S: feat M1 behavior-neutral compile stubs (spec.md draft to in-progress rides this commit) |
| `fdcf28056` | R: test M1 RED baseline - reproduction tests and red-baseline.md |
| `c0d46d96f` | feat M2 attach, preconditions, fallback and clean terminal |
| `5e50715a8` | feat M3 shared-thread busy tracking and request scoping |
| `c1c17cdd5` | feat M4 TUI exit status, server-death monitor and blocked-call release |
| `82ae15be1` | fix M5 lint-clean the TUI attach and its tests |
| `1194c3415` | fix M5 owner re-checks the managed gate before probing |
| `c8ebb353c` | test M5 attach-order test able to see an early attach (mu3 killed) |
| `44b0d1822` | docs M6 operator document, CHANGELOG entry, manual check form |

S was amended once, before R existed (a `stopTUI` stub was missing); S and R are the shas above and nothing else was rewritten.

### AC-MT-014 M1 delta check (run on `red-baseline.md` at the end of M1, output verbatim)

```
$ grep -o -E '^[|] AC-MT-[0-9]+' .moai/specs/SPEC-FACTORY-MANAGED-TUI-001/red-baseline.md | sort -u | wc -l
11
$ grep -o -E -e '--- FAIL: [A-Za-z0-9_]+' .moai/specs/SPEC-FACTORY-MANAGED-TUI-001/red-baseline.md | sort -u | wc -l
15
$ grep -c -e 'undefined:' -e 'build failed' -e 'panic:' -e 'test timed out' .moai/specs/SPEC-FACTORY-MANAGED-TUI-001/red-baseline.md
0   (exit 1)
```

Ancestry: S is an ancestor of R (exit 0), R is not an ancestor of S (exit 1); `git ls-files` lists `red-baseline.md`; `git check-ignore -v` on it prints nothing (exit 1). Closing part: R `fdcf28056` is an ancestor of every commit that changed non-test files (`c0d46d96f`, `5e50715a8`, `c1c17cdd5`, `82ae15be1`, `1194c3415`: exit 0 each) and not the reverse (exit 1 for `c0d46d96f` and `1194c3415`). All 15 RED cells were produced by this lane running the commands itself; no second party re-ran one, which stays a Gap.

### Verification matrix (HEAD `44b0d1822` plus the progress.md commit that follows; exit codes observed)

Command form: the lane's worktree guard refuses a dollar sign in a command, so `-run` patterns are prefix-anchored (`'^Name'`) instead of ending in the dollar sign; `go test ./internal/cli -list TestManagedCodex` and `-list TestManagedTUI` show each of the 15 names once and no other test sharing a prefix. Alternation blocks are run one name per invocation.

| AC | Command (all `go test -race ./internal/cli -count=1` unless noted) | Result |
|---|---|---|
| 001 | `-run '^TestManagedCodexTUIAttachCommand'` | exit 0, 7 subtests PASS |
| 002 | `-run '^TestManagedCodexTUIStartsAfterPrimingAndBind'` | exit 0, `order`, `priming_failure_never_attaches` PASS |
| 003 | `-run '^TestManagedCodexTUIPreconditionsAndFallback'` | exit 0, 11 subtests PASS |
| 003 | `-run '^TestManagedCodexRemoteSupportProbe'` | exit 0, 3 subtests PASS |
| 004 | `-run '^TestManagedCodexTUIOwnsOperatorStdin'` | exit 0 |
| 005 | `-run '^TestManagedCodexTUIKeepsTerminalClean'` | exit 0, 3 subtests PASS |
| 006 | `-run '^TestManagedCodexTUIDefersDeliveryWhileBusy'`, `'^TestManagedCodexReaderSurvivesOperatorTurns'`, `'^TestManagedCodexBusyLongTurnKeepsDeferring'` | exit 0 each |
| 007 | `-run '^TestManagedCodexServerRequestScopingWithTUI'` | exit 0, 6 subtests PASS |
| 008 | `-run '^TestManagedCodexTUIExitEndsSession'` | exit 0, 6 subtests PASS |
| 009 | `-run '^TestManagedCodexServerDeathStopsTUI'`, `'^TestManagedCodexSessionEndStopsTUI'` | exit 0 each |
| 010 | `go test ./internal/cli -run '^Name' -count=1 -v` for `TestFactoryManagedRequested`, `TestManagedLaunchRequiresOptIn`, `TestManagedCodexLaunchRequiresOptIn`, `TestManagedSwitchDoesNotReachCodexLaneLoop` (each listed once by `-list`) | exit 0, 4 PASS |  [Correction in the re-close: card t1440 removed `TestManagedSwitchDoesNotReachCodexLaneLoop` upstream (`a184aa89c`); AC-MT-010 now names the four surviving opt-in tests, `TestFactoryManagedRequested`, `TestManagedLaunchRequiresOptIn`, `TestManagedCodexLaunchRequiresOptIn` and `TestManagedTUINeverReachedWithoutOptIn`, per SPEC 0.3.2. The cell above is the original run record.]
| 011 | `-run '^TestManagedCodexTUILoopbackRoundTrip'` | exit 0 |
| 012 | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| 012 | `grep -rn 'syscall\.' internal/cli/managed_*.go` | no lines, exit 1 (positive control `grep -ln 'syscall\.' internal/cli/launch_exec_posix.go` prints that path, exit 0) |
| 012 | `git diff --stat 2b9e4a4d0..HEAD -- internal/factorymsg/store.go .moai/specs/SPEC-FACTORY-MANAGED-SESSION-001 .moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001` | no lines (positive control: the same range on `internal/cli` lists 5 files) |
| 013 | the section 1.3 grep list | operator document: the two old sentences 0 and 0; SPEC id 3; opt-out, log-file, quit, probe, signals, manual-check anchors 1 each; unobserved anchors 3; debt-status 1 (line says `addressed` and `unverified`, no `resolved`); CHANGELOG: SPEC id 1, `anchor:tui-` lines 5 |
| 014 | M1 delta check and ancestry above | 11 / 15 / 0, ancestry holds |
| 015 | operator-held manual check | NOT RUN (`manual-check.md`: operator confirmation pending, no terminal designated) |
| 016 | `go test ./internal/cli -run '^TestManagedTUINeverReachedWithoutOptIn' -count=1 -v` | exit 0, `no_switch`, `switch_only`, `stamps_only`, `opted_in_probes_once` PASS |

Repetition: every one of the 15 tests also ran `-race -count=5` as its own invocation (all exit 0, BAD=0), including the concurrency-sensitive 006-009; the 15-test single pass ran `-race -count=1` with BAD=0. Regression: `go test -race ./internal/cli -run Managed -count=1` exit 0 (`ok ... 84.991s`); managed top-level tests by `-list 'Managed|managed'`: 95 (base 79, plus 15 new, plus the fake codex helper). Existing managed suites were green after every milestone.

Static: `golangci-lint run ./internal/cli/... ./internal/config/...` v2.1.6 gives `0 issues.` (six findings of the first run, all in new code, were fixed in `82ae15be1`); `go vet ./internal/cli/... ./internal/config/...` is clean; `go build ./...` exit 0; `gofmt -l internal/cli internal/config` lists only `internal/cli/todo_classify_llm_test.go`, which this card did not touch (baseline state).

Coverage of `managed_codex_tui.go` by the managed set (`-run Managed -coverprofile`): 208 of 224 statements, 92.9 percent (measured before the gate re-check and the test-only fixes; the later commits add 5 statements, all exercised by AC-016).

[Correction in the re-close: the 208 of 224 figure above is stale for the current tree. The re-close measured 219 of 241 statements, 90.9 percent, after the F5 and F6 repairs added code and tests (see the re-close block in section E.4). The original line is kept as the run-phase record.]

### RED output (E8)

`red-baseline.md` holds the verbatim pre-GREEN failing output per AC (commit R, measured on S with the R tests applied). Example, AC-MT-001: `--- FAIL: TestManagedCodexTUIAttachCommand (10.50s)` with subtests `argv_exact`, `token_not_in_argv`, `token_in_env`, `no_generated_approval_args`, `operator_args_forwarded`, `cwd_matches_thread` failing; `app_server_keeps_token_file` passed on the base by design.

### Mutants (one edit at a time on the committed tree, working files restored after each; the tree was clean afterwards)

| id | outcome | killed by |
|---|---|---|
| mu1 token value in TUI argv | KILLED | AC-001 `argv_exact`, `token_not_in_argv` |
| mu2 approval overrides in TUI argv | KILLED | AC-001 |
| mu3 attach before priming completes | first SURVIVED (the fake priming turn finished before a TUI process could start); test fixed with a turn-completion delay (`c8ebb353c`), then KILLED | AC-002 `order` |
| mu4 non-terminal stdin accepted | KILLED | AC-003 `stdin_not_terminal` |
| mu5 driver keeps stdin reader | KILLED | AC-004 |
| mu6 log lines stay on terminal | KILLED | AC-005 |
| mu7 App Server stderr on terminal | KILLED | AC-005 |
| mu8 claim ignores busy | KILLED | AC-006 `TestManagedCodexTUIDefersDeliveryWhileBusy` |
| mu9 operator-turn frames reach the event channel | KILLED | AC-006 `TestManagedCodexReaderSurvivesOperatorTurns` |
| mu10 operator-turn request answered | KILLED | AC-007 `operator_turn_unanswered` |
| mu11 scoping ignores the attach state | KILLED | AC-007 |
| mu12 TUI exit status dropped to 0 | KILLED | AC-008 `exit_seven`, `signaled` |
| mu13 TUI not stopped at session end | KILLED | AC-009 `TestManagedCodexSessionEndStopsTUI` |
| mu14 owner gate removed | KILLED | AC-016 negative subtests (direct owner call) |
| mu15 TUI exit does not release blocked calls | KILLED | AC-008 `blocked_turn_start_released`, `blocked_wait_turn_released` |
| mu16 busy flag self-clears | KILLED | AC-006 `TestManagedCodexBusyLongTurnKeepsDeferring` |
| mu17 stuffed documentation | KILLED (a scratch copy holding all tokens on separate lines fails every anchored same-line grep of section 1.3) | AC-013 |
| mu18 post-reap line still to the file | KILLED | AC-005 `post_reap_line_on_terminal` |
| mu19 release only through a select channel | KILLED, only `blocked_turn_start_released` fails | AC-008 |

Note on mu14: the launcher-gate form (editing `codex_launcher.go`) was not run because that file is outside this card's ownership; the owner-level gate was mutated.

### Deviations from the SPEC text and why

1. `-run` patterns have no trailing dollar sign and alternation blocks run per name (worktree guard refusal); the swept set is argued by `-list`.
2. The owner re-checks the managed gate in `planOperatorTUI` (REQ-MT-001 says the owner itself neither probes nor attaches outside the gate; calling the owner entry directly was otherwise ungated).
3. Operator document: besides the bullets, two table rows changed (the Codex row, whose old sentence AC-MT-013 requires gone, and the serial-queue row, which would otherwise misdescribe an attached session).
4. CHANGELOG was edited in the run phase because the leader's instruction and plan.md M6 list it; the manager-develop role definition reserves CHANGELOG to manager-docs, so the sync phase should treat the entry as a draft to confirm.
5. `manual-check.md` was added to the SPEC directory as the fill-in form for AC-MT-015.
6. The fake App Server needs the parent's MOAI_HOME sandbox marker, otherwise the re-executed test binary creates its own home and the broker store differs (harness fact).

### Known debt and unobserved

Real interactive TUI behavior (AC-MT-015) is NOT claimed. P6 (lifecycle frames of operator turns reach the launcher), P7 (request routing with two connections), P9 (real resume, rendering, `/quit`, terminal restoration), terminal detection on a real tty and Windows runtime are unobserved. The repository-wide test verdict belongs to CI on the integration branch and is PENDING; the whole `internal/cli` package suite was not run locally (SPEC constraint). No second party re-ran a RED cell.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-10-03
run_commit_sha: 44b0d1822   # last code/doc commit; the progress.md commit follows it
run_status: audit-ready (AC-MT-015 operator-held, not run)
ac_pass_count: 15   # AC-MT-001..014 and 016, automated
ac_fail_count: 0
preserve_list_post_run_count: 0   # store.go, both parent SPEC directories, codex_launcher.go unchanged
l44_pre_commit_fetch: not run (lane worktree; the leader owns the integration push)
l44_post_push_fetch: not applicable (no push by the run phase)
new_warnings_or_lints_introduced: 0   # golangci-lint v2.1.6: 0 issues
cross_platform_build:
  windows_amd64: pass
  native: pass
total_run_phase_files: 13
m1_to_mN_commit_strategy: stub S, RED R, one commit per milestone M2..M4, two M5 fix commits, one M5 test fix, one M6 docs commit
```

## §E.4 Sync-phase Audit-Ready Signal

sync_status: audit-ready (AC-MT-015 operator-held, NOT run; real TUI behavior never observed)
sync_complete_at: 2026-10-03
superseded_sync_commit_sha: 40aa3aedf   # the first sync commit (backfill commit 3ca47fbe3); the independent sync audit returned FAIL 76.2 and the re-close block below supersedes it; the key was renamed from `sync_commit_sha` so tools read the current close
head_at_signal: 71befeeb9 (measured tree, = the card branch after absorbing local develop 1da5e4fc6; the sync edits were uncommitted when measured)
tree: .claude/worktrees/t1408
branch: WT-managed-codex-tui-attach
owner: manager-docs (sync-phase)
ac_source: .moai/specs/SPEC-FACTORY-MANAGED-TUI-001/acceptance.md (tier M, state `resolved`, non-empty)
docs_changed: `CHANGELOG.md` (the run-phase draft rewritten in place: complete run SHA list, card-child headless bullet, merge-hotspot bullet, honest AC-MT-015 and debt wording; still one entry) · `.moai/specs/SPEC-FACTORY-MANAGED-TUI-001/spec.md` (frontmatter `status:` line only) · this section
docs_not_changed: `.moai/docs/factory-managed-session.md` (read against the code after the merge, nothing found wrong; see below) · plan.md, acceptance.md, design.md, red-baseline.md, manual-check.md (no `status:` field; `updated:` already 2026-10-03) · every `.go` file, `internal/cli/codex_launcher.go`
changelog_entry_position: `[Unreleased]` -> `### Added`, first item, directly above SPEC-FACTORY-MANAGED-CARD-CHILD-001 (t1440)
frontmatter_status_transitions: spec.md `in-progress -> implemented -> completed` merged into this single sync commit (frontmatter records `status: completed`); `updated: 2026-10-03` unchanged (already the sync date)
b12_self_test_a: before the edit the run-phase draft was already present (count 1, so no duplicate emission was possible); after the edit `grep -c 'SPEC-FACTORY-MANAGED-TUI-001' CHANGELOG.md` -> `1`
b12_self_test_b: B12 counter on acceptance.md -> stdout `16`, stderr `live=16 excluded=0 ambiguous=0`, exit 0; `grep -oE '^- \*\*REQ-MT-[0-9]{3}\*\*' spec.md | sort -u | wc -l` -> `14`; the entry states 14 requirements and 16 acceptance criteria
b12_self_test_c: every path the entry cites exists (`internal/cli/managed_codex_tui.go`, `managed_codex_factory.go`, `managed_factory_session.go`, `managed_codex_tui_test.go`, `internal/config/defaults.go`, `internal/config/envkeys.go`, the operator document, `manual-check.md`, `progress.md`, `red-baseline.md`); the run SHAs were read from `git log` of this tree
canary_compliance_check: not applicable
ac_mt_013: operator-document and CHANGELOG anchors re-counted at this HEAD (table below)
ac_mt_015: NOT RUN. Record: "operator confirmation pending, no terminal designated" (`manual-check.md`). Steps 1, 2, 3 and 6 are unobserved.
known_debt_1_of_SESSION_001: addressed, unverified (fake codex only)

All attributions are `(this run, this tree, HEAD 71befeeb9 plus the uncommitted sync edits, before the sync commit)`.

### Evidence

| Check | Command | Observed |
|---|---|---|
| SPEC lint | `moai spec lint .moai/specs/SPEC-FACTORY-MANAGED-TUI-001` | `No findings - all SPEC documents are valid`, exit 0 |
| AC count (B12) | counter from `manager-docs.md` on acceptance.md | `16`, `live=16 excluded=0 ambiguous=0`, exit 0 |
| REQ count | `grep -oE '^- \*\*REQ-MT-[0-9]{3}\*\*' spec.md` piped to `sort -u` and `wc -l` | `14` |
| CHANGELOG | `grep -c 'SPEC-FACTORY-MANAGED-TUI-001' CHANGELOG.md` / `grep -c 'anchor:tui-' CHANGELOG.md` | `1` / `6` (acceptance needs 1+ and 5+) |
| Merge-tree managed regression | lane env scrubbed, `go test -race -count=1 -run Managed ./internal/cli` | `ok  github.com/modu-ai/moai-adk/internal/cli  163.582s`, exit 0 |
| Public docs | `grep -rlE "MOAI_FACTORY_MANAGED|factory-managed|managed Codex|App Server" internal/template/templates docs-site/content README.md README.ko.md README.ja.md README.zh.md` | no file listed; the headless/no-TUI wording grep over the same trees (`headless|no TUI|헤드리스` filtered by `managed|codex`) hit only `e2e-tester.toml:128` (unrelated browser headless/headed note). No public statement to fix. |
| Operator-doc accuracy | read lines 46, 48, 75-86, 119 of `.moai/docs/factory-managed-session.md` against `operatorTUIPreconditions` (`managed_codex_tui.go:185-204`) | opt-out values, stdin/stdout terminal checks, `--remote` probe and the card-child headless bullet all match the code; nothing edited |
| Spec drift (repo-wide) | `moai spec drift --no-cache --count` | `177` (repo-wide baseline, not attributed to this SPEC) |

### Gaps

- AC-MT-015 and every real-TUI behavior: not run, no terminal designated.
- Codemaps freshness and the repository-wide test suite: not run here (CI on the integration branch owns the suite; codemaps regeneration was not requested).
- The `moai` binary used for `spec lint` / `spec drift` is the installed v3.2.0-rc.27; its commit was not compared with this tree's HEAD, so these two tool readings carry no judging-build attribution.
- The card-child headless statement is read from code, not observed on a lane terminal.

### Residual-risk

- The attach may not work with a real codex (frame delivery to the launcher, request routing with two connections, resume by thread id, terminal restoration).
- The attach touches the same lines as cards t1459 and t1410; the second to land must re-measure.

## §E.4 Sync-phase Audit-Ready Signal — re-close after the sync-audit iteration 1 FAIL

sync_status: audit-ready (NOT audit-passed: no sync audit has passed on this close yet; AC-MT-015 operator-held, NOT run)
sync_complete_at: 2026-10-04
sync_commit_sha: 985cedfb7   # backfilled in the following progress.md-only commit (a commit cannot cite its own hash)
superseded_first_close: 40aa3aedf (the first sync commit; the independent sync audit `.moai/reports/t1408/sync-audit.md`, local-only, returned FAIL 76.2/100 against `3ca47fbe3`; its SHA is the `superseded_sync_commit_sha` line of the first section E.4 block above)
head_at_signal: 6f6d69349 (measured tree; the re-close edits were uncommitted when measured)
tree: .claude/worktrees/t1408
branch: WT-managed-codex-tui-attach
owner: manager-docs (sync-phase, re-close)
card_commits: run `d222d310e` `fdcf28056` `c0d46d96f` `5e50715a8` `c1c17cdd5` `82ae15be1` `1194c3415` `c8ebb353c` `44b0d1822` `d6b710690` · develop absorb `71befeeb9` · first sync `40aa3aedf` (superseded) · backfill `3ca47fbe3` · gofmt repair `b1ec7c725` · F5/F6 repair `42a952661` · SPEC amendment 0.3.2 `febfbeb0c` and `6f6d69349` · re-close (the commit carrying this block) · backfill (the next commit)
ac_source: .moai/specs/SPEC-FACTORY-MANAGED-TUI-001/acceptance.md (tier M, state `resolved`, non-empty)
docs_changed: `CHANGELOG.md` (the single existing entry edited in place; audit status, repair and amendment SHAs, F5/F6 fixed, F4 and the log-open race as known limits, measured coverage) · `.moai/docs/factory-managed-session.md` (two known-limit bullets for SPEC debts 14 and 15 so REQ-MT-013 holds; the card-child bullet now quotes the old sentence `Codex 관리 세션은 화면에 아무것도 보여 주지 않는다` scoped to card children, because the completed SPEC-FACTORY-MANAGED-CARD-CHILD-001 AC-CC-012 expects that sentence to be present and the first close had dropped it, a regression the amended AC-MT-013 no longer constrains) · `.moai/specs/SPEC-FACTORY-MANAGED-TUI-001/spec.md` (frontmatter `status:` line only) · this block, the key rename in the first block, and two correction notes in section E.2
docs_not_changed: spec.md body, plan.md, acceptance.md, design.md (manager-spec's amendment stands) · every `.go` file
frontmatter_status_transitions: spec.md `in-progress -> implemented -> completed` merged into this re-close commit (the amendment had set `in-progress`; `amendment_of` and `prior_completed_sha` are untouched); `updated: 2026-10-04` already the amendment date
changelog_entry_position: unchanged (single entry, first item under Added)
b12_self_test_a: re-close edits the single existing entry in place; `grep -c 'SPEC-FACTORY-MANAGED-TUI-001' CHANGELOG.md` -> `1`
b12_self_test_b: B12 counter on acceptance.md -> see the table below; the entry states 14 requirements and 16 acceptance criteria
b12_self_test_c: the entry adds no new file path beyond `internal/cli/managed_codex_tui_repair_test.go`, which exists; every SHA it cites was read from `git log`
canary_compliance_check: not applicable
audit_status: iteration 1 FAIL 76.2/100 (Functionality 72, must-pass); the cross-model gate (claude plus codex) was not met (codex failed twice, no independent claude backend run established). F1 and F7 were SPEC-text defects repaired by the amendment; F2 gofmt repaired in `b1ec7c725`; F5 and F6 fixed in `42a952661`; F4 deliberately not fixed (SPEC debt 14); the log-open race is SPEC debt 15. No re-audit has been run.

All attributions are `(this run, this tree, HEAD 6f6d69349 plus the uncommitted re-close edits, before the re-close commit)`.

### Evidence

| Check | Command | Observed |
|---|---|---|
| SPEC lint | `moai spec lint .moai/specs/SPEC-FACTORY-MANAGED-TUI-001` | no findings, exit 0 |
| AC count (B12) | counter from `manager-docs.md` on acceptance.md | `16`, `live=16 excluded=0 ambiguous=0` |
| CHANGELOG | `grep -c 'SPEC-FACTORY-MANAGED-TUI-001' CHANGELOG.md` / `grep -c 'anchor:tui-' CHANGELOG.md` | `1` / `7` |
| Operator document anchors | `grep -c 'anchor:tui-' .moai/docs/factory-managed-session.md` | `13`; `anchor:tui-signals.*t1459` 1; `anchor:tui-unobserved.*미관측` 3; old sentence `화면에는 아무것도 나타나지 않는다` 0 (AC-MT-013); `Codex 관리 세션은 화면에 아무것도 보여 주지 않는다` 1 (AC-CC-012 of CARD-CHILD-001); `런처는 시그널을 처리하지 않는다` 1 |
| Managed regression | lane env scrubbed, `timeout 900 go test -race -count=1 -run Managed -coverprofile=... ./internal/cli` | `ok  github.com/modu-ai/moai-adk/internal/cli  161.807s`, exit 0 |
| Coverage of `managed_codex_tui.go` | awk over the profile by statement count, cross-checked with `go tool cover -func` | `219/241 = 90.9%` (equals the repair lane's figure; the audit's 208/226 predates the repair) |
| Build | `go build ./...` | exit 0 |

### Gaps

- No sync audit has passed on this close. Iteration 1 was FAIL 76.2 and the cross-model gate (claude plus codex) was unmet; this re-close has not been re-audited.
- AC-MT-015 and every real-TUI behavior: not run, no terminal designated.
- Codemaps freshness and the repository-wide suite: not run (CI owns the suite).
- The installed `moai` binary's commit was not compared with this tree's HEAD, so the `spec lint` reading carries no judging-build attribution.
- The F5, F6 and gofmt repairs were not re-verified individually here beyond the managed-scope run passing.

### Residual-risk

- F4 (armed-window misattribution) and the log-open race remain, disclosed as SPEC debts 14 and 15.
- Real-TUI behavior may differ from the fake codex; the hotspot merge with t1459 and t1410 still needs a re-measure by whichever lands second.
