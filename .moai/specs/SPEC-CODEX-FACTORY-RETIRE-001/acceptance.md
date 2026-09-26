# SPEC-CODEX-FACTORY-RETIRE-001 — acceptance

Every criterion is binary. `<base>` is the `run_base:` value the run lane records in
`progress.md` §E.2 (plan.md §C.1); the plan was authored on `553e224f3`. `<bin>` is a
binary built from the tree under test with `go build -o <scratch>/moai ./cmd/moai`.
Every grep-count criterion carries a positive control that must be non-zero, so a zero
cannot come from a wrong path or pattern. Every `go test` criterion runs with `-v` and
names the tests it expects, so an empty selection cannot pass (`[no tests to run]`
fails the criterion).

## §D.0 Red at arrival (measured on `553e224f3`)

| AC | Why it is red before the run | Milestone that flips it |
|---|---|---|
| 001, 002 | `-k`/`-f` are accepted and enter kanban/factory mode (`codex_launcher.go:711-745`) | M1 |
| 005, 006 | `codexChildEnv` drops only 2 keys; spawn forwards 9 lane keys (`:296-310`) | M1 |
| 007 | `codex_direct_posix.go:39` registers launch-pending under lane env | M1 |
| 008, 010 | No backend check on join/lead run selection (`factory.go:245`) | M1 |
| 011, 012, 016 | The symbols exist (positive controls below) | M2, M3 |
| 019 | 26 line matches + 5 folded matches on the file set | M4 |
| 025 | `registerFactoryHookPeer` is harness-agnostic (`factory_messages.go:53-116`) | M1 |
| 003, 004, 009, 013-015, 017, 018, 020-023 | Preservation guards: green before and must stay green | — (mutant probes stated in each) |
| 024 | Foreign files still present in the develop worktree (lead report) | M5 |

## §D AC Matrix

### M1 — refusal, environment, joins

### AC-CFR-001 — `moai codex -k` is refused (maps REQ-CFR-001, REQ-CFR-005)

**Given** `<bin>` and any working directory,
**When** each of `codex -k`, `codex --kanban`, `codex -k SPEC-X-001`,
`codex -k --name plan`, `codex --kanban=SPEC-X-001`, `codex -k cli`, `codex -k status`,
`codex -k -w t1`, `codex -k --spawn` is run,
**Then** each exits 1, stderr is exactly one line containing
`KANBAN_MODE_UNSUPPORTED_BACKEND` and `moai cc -k`, and stdout is empty.

### AC-CFR-002 — `moai codex -f` is refused (maps REQ-CFR-002, REQ-CFR-005)

**Given** `<bin>`,
**When** each of `codex -f`, `codex --factory`, `codex -f worker`, `codex -f worker-2`,
`codex -f agent`, `codex -f lane-3`, `codex --factory=worker`, `codex -f=worker-1`,
`codex --factory-run r1`, `codex -f app`, `codex -f status` is run,
**Then** each exits 1, stderr is exactly one line containing
`FACTORY_MODE_UNSUPPORTED_BACKEND`, `moai cc -f` and `moai glm -f`, and stdout is empty.

### AC-CFR-003 — refusal has no state effect (maps REQ-CFR-003)

**Given** a test in `./internal/cli` with an isolated `MOAI_HOME` and project root, the
codex launch seams captured, and the sha256 (or absence) of the factory state DB and the
worker registry file recorded,
**When** the AC-CFR-001 and AC-CFR-002 argument sets run through `runCodex`,
**Then** the captured launch count is 0, both files hash identically (or are still
absent), no `MOAI_KANBAN*`/`MOAI_FACTORY_*` key changed in `os.Environ()`, and no
directory appeared under `.claude/worktrees/`; the same harness records exactly 1
launch for the bare form (positive control). Mutant: a refusal placed after
`recordFactoryRunStart` must turn this red.

### AC-CFR-004 — tokens after `--` pass through (maps REQ-CFR-004)

**Given** the AC-CFR-003 harness,
**When** `runCodex` receives `-- -f worker -k`,
**Then** exactly one launch is captured and its argv tail ends with `-f worker -k`.

### AC-CFR-005 — direct child carries no lane identity (maps REQ-CFR-006, REQ-CFR-009)

**Given** each of the eleven lane keys set to a distinct non-empty value, plus
`MOAI_HOME`, `MOAI_AUTONOMY_TIER`, `CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS`, and a
sentinel `T1242_KEEP=1`,
**When** the direct-launch child environment is built,
**Then** for each of the eleven keys, separately asserted, the environment holds no
entry with a non-empty value; `MOAI_HOME`, `MOAI_AUTONOMY_TIER`,
`CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS`, `T1242_KEEP=1` and the resolved `CODEX_HOME`
are present; `CLAUDE_CODE_SESSION_ID` and `MOAI_SESSION_PID` are absent. Mutant:
dropping any single key from the scrub list turns exactly that key's assertion red.

### AC-CFR-006 — spawn command carries no lane identity (maps REQ-CFR-007)

**Given** the AC-CFR-005 environment,
**When** the spawn command string is built,
**Then** for each of the eleven keys it contains `KEY=` with an empty value and never
`KEY=<the set value>`; `MOAI_HOME=` carries its set value (positive control).

### AC-CFR-007 — no factory registration from a codex launch under lane env (maps REQ-CFR-008)

**Given** an active factory run `r1` recorded with lead backend `claude`, and
`MOAI_KANBAN_ID=r1`, `MOAI_FACTORY_WORKER=worker-1` in the environment,
**When** a codex direct launch (seam-captured) and a codex spawn launch (tmux seam and
pane-identity seam faked) run,
**Then** the run's launch-pending peer count is 0 and its `lead_pid` /
`lead_process_start` are unchanged; and, as the positive control under the same
environment, `go test ./internal/cli -run '^TestFactoryLauncherRegistersLaunchPendingPeers$' -count=1 -v`
prints `--- PASS: TestFactoryLauncherRegistersLaunchPendingPeers` (the cc path still
registers, with a `claude` run fixture).

### AC-CFR-008 — joining a codex run is refused (maps REQ-CFR-010)

**Given** exactly one active run `rc` with `lead_backend = 'codex'`,
**When** `moai cc -f worker`, `moai cc -f worker-2` and `moai glm -f worker` entries are
driven through their launch seams,
**Then** each returns a non-zero exit, the message contains `rc`, `codex` and
`moai factory runs --retire`, the worker registry file hashes identically before and
after, and no launch is captured.

### AC-CFR-009 — non-codex runs are joined and led as before (maps REQ-CFR-011)

**Given** the AC-CFR-008 setup with `lead_backend = 'claude'`,
**When** (a) `moai cc -f worker` and (b) `moai cc -f --factory-run rc` are driven through
the same seams,
**Then** (a) captures exactly one launch and claims one registry slot, and (b) is not
refused, captures exactly one launch, and leaves the `runs` row for `rc` with
`status = 'active'` and `lead_backend = 'claude'`.

### AC-CFR-010 — a lead cannot adopt a codex run (maps REQ-CFR-010)

**Given** the AC-CFR-008 run `rc`,
**When** `moai cc -f --factory-run rc` is driven through the launch seams,
**Then** the entry is refused with the AC-CFR-008 message, and
`SELECT lead_backend, lead_pid FROM runs WHERE run_id='rc'` returns the values
recorded before the call.

### AC-CFR-025 — codex-harness hooks register no factory peer (maps REQ-CFR-022)

**Given** an isolated `MOAI_HOME`, an active run `r1` (`lead_backend = 'claude'`), and
`MOAI_KANBAN_ID=r1`, `MOAI_FACTORY_WORKER=worker-1` in the environment,
**When** the session-start and user-prompt-submit hooks run once with `--harness codex`
and once without it, each with a distinct session id,
**Then** after the `--harness codex` runs the peer row count for `(r1, worker-1)` is 0
and no endpoint was rotated; after the Claude runs it is 1 (positive control). Mutant:
removing the codex-harness guard turns the first count to 1.

### M2 — dead entry code

### AC-CFR-011 — codex entry code is gone (maps REQ-CFR-012)

**Given** the tree under test,
**When** `grep -rnwE 'stripCodexKanbanFlag|applyCodexKanbanEntry|codexKanbanEntry|codexKanbanUsageDiag|stripCodexFactoryFlag|applyCodexFactoryEntry|codexFactoryBackend' internal cmd pkg --include='*.go'`
and `ls internal/cli/codex_kanban.go internal/cli/codex_factory.go` run,
**Then** the grep prints 0 lines and `ls` reports both files missing; the same pattern
through `git grep -nwE '<pattern>' 553e224f3 -- internal` prints ≥ 1 line (positive
control).

### AC-CFR-012 — codex launch files hold no factory calls (maps REQ-CFR-008, REQ-CFR-012)

**Given** the tree under test,
**When** `grep -nE 'registerFactoryLaunchPending|rollbackFactoryLaunchPending|stampFactoryRunOwner|clearFactoryRunOwner|recordFactoryRunStart|enterSelectedFactoryRun|factoryLaunchEnabled|stripFactoryRunFlag' internal/cli/codex_launcher.go internal/cli/codex_direct_posix.go internal/cli/codex_direct_windows.go`
runs,
**Then** it prints 0 lines, and
`grep -c registerFactoryLaunchPending internal/cli/launch_exec_posix.go` prints ≥ 1
(positive control: the shared path survives).

### AC-CFR-013 — kept codex surfaces are unchanged (maps REQ-CFR-016)

**Given** the tree under test,
**When** `grep -c 'syscall.Exec' internal/cli/codex_direct_posix.go`,
`head -1 internal/cli/codex_direct_posix.go internal/cli/codex_direct_windows.go`, and
`go test ./internal/cli -run '^(TestWorktreeLaunchRejectsConcurrentWriter|TestCodexWorktreeAnchorLockAndBase|TestCodexSpawnAnchorsToPanePID|TestCodexLaunchVerb_StatusStaysTheReadout|TestCodexLocalInstructions_DirectSpawnAndAppSharePrefix|TestCodexTask_ForegroundReturnsOutput|TestSessionMsgSendPollAckHandlers|TestCodexRoleLoadNegativeControl)$' -count=1 -v`
run,
**Then** the grep prints ≥ 1; the build-tag lines are `//go:build !windows` and
`//go:build windows` respectively; the test run exits 0 and prints exactly one
top-level `--- PASS: <name>` line for each of the eight named tests (8 lines).

### AC-CFR-014 — shared-code test survives its host file (maps REQ-CFR-017)

**Given** the tree under test,
**When** `go test ./internal/cli -run '^TestNextFactoryWorkerNumber$' -count=1 -v`
runs,
**Then** the output holds exactly one `--- PASS: TestNextFactoryWorkerNumber` line and
`ls internal/cli/codex_factory_test.go` reports the file missing.

### AC-CFR-015 — cross-platform build (maps REQ-CFR-012, REQ-CFR-016)

**Given** the tree under test,
**When** `GOOS=darwin go build ./...`, `GOOS=linux go build ./...` and
`GOOS=windows go build ./...` run,
**Then** each exits 0.

### M3 — handoff CLI and relocation

### AC-CFR-016 — handoff CLI and relocation client are gone (maps REQ-CFR-013)

**Given** the tree under test,
**When** `grep -rnwE 'prepareLaneHandoff|switchLaneHandoffInteractive|switchLaneHandoffHeadless|bindLaneHandoffHeadless|recoverLaneHandoff|codexThreadRelocation|runCodexThreadRelocation|laneHandoffAppServer|DefaultCodexHandoffRelocationTimeout' internal cmd pkg --include='*.go'`
runs,
**Then** it prints 0 lines and `ls internal/cli/factory_lane_handoff*.go` matches no
file; `git grep -nwE '<pattern>' 553e224f3 -- internal` prints ≥ 1 line (positive
control).

### AC-CFR-017 — factory state sources are byte-identical (maps REQ-CFR-014, REQ-CFR-015)

**Given** the tree under test,
**When** `git diff --exit-code <base> -- internal/factorymsg internal/homestate ':!*_test.go'`
runs,
**Then** it exits 0 with empty output. This covers every multi-line `CREATE TABLE`
body (including `runs.lead_backend`, `internal/homestate/factory.go:36`), every
`ALTER TABLE` (`factory.go:226-228`), and every index statement. Positive control:
`git grep -c 'lead_backend' <base> -- internal/homestate/factory.go` prints ≥ 1, and
a scratch edit of one column name in that file makes the diff command exit 1.

### AC-CFR-018 — shared factory core still passes (maps REQ-CFR-014)

**Given** the tree under test and a scratch `MOAI_HOME`,
**When** `go test ./internal/factorymsg ./internal/kanban -count=1`,
`MOAI_HOME=<scratch> go test ./internal/hook -run 'Factory' -count=1 -v`,
`go test ./internal/cli -run '^(TestFactoryLaneHandoffOperatorAbandon|TestFactoryLeadNoticeUsesOperationalStatus|TestFactoryRunSelectionAtomicSlotsAndArgv)$' -count=1 -v`
and `<bin> factory handoff abandon-lane --help` run,
**Then** the first command prints one `ok` line per package; the hook run exits 0,
contains `--- PASS: TestFactoryLaneHandoffInteractiveStateMachine`, and does not print
`[no tests to run]`; the cli run prints `--- PASS:` for each of the three named tests;
and the help command exits 0.

### AC-CFR-020 — the rebuilt abandon fixture still builds a real target (maps REQ-CFR-017, REQ-CFR-014)

**Given** the tree under test after M3,
**When** `go test ./internal/cli -run '^TestFactoryLaneHandoffOperatorAbandon$' -count=1 -v`
runs,
**Then** it exits 0 and prints four `--- PASS: TestFactoryLaneHandoffOperatorAbandon/`
subtest lines; the fixture asserts, before each abandon call, that the target
worktree path exists and the target branch ref resolves. Mutant (run once in M3 and
recorded in `progress.md` §E.2): a fixture that skips the worktree/branch step makes
this command exit non-zero.

### M4 — docs, config, budget

### AC-CFR-019 — retired wording is gone from every user-facing surface (maps REQ-CFR-018)

**Given** the file set `AGENTS.md internal/template/templates/AGENTS.md.tmpl .claude/rules/moai/core/moai-mcp-tools.md .claude/rules/moai/core/moai-mcp-tools-catalogue.md internal/template/templates/.claude/rules/moai/core/moai-mcp-tools.md internal/template/templates/.claude/rules/moai/core/moai-mcp-tools-catalogue.md docs-site/content/*/advanced/codex-dual-harness.md docs-site/content/*/guides/mcp-server.md`,
**When** these three checks run:
1. line check — `grep -nE "codex -f|codex -k|Codex lanes?|lane's shell|Codex 레인|레인 셸|Codex レーン|レーンのシェル|Codex 泳道|泳道 shell" <file set>`;
2. line-break-folded check (catches `AGENTS.md:265-266`, where `` `-f` `` and
   `lead/agents` sit on two lines, and the ja "`` `-f` `` の lead/agents") —
   `cat AGENTS.md docs-site/content/{ko,en,ja,zh}/advanced/codex-dual-harness.md | tr '\n' ' ' | grep -oE '`-f`[^|]{0,6}lead/agent' | wc -l`;
3. mirror parity — `cmp` of the local and template copies of `moai-mcp-tools.md` and
   of `moai-mcp-tools-catalogue.md`,
**Then** check 1 prints 0 lines, check 2 prints `0`, and both `cmp` calls exit 0.
Positive controls on `553e224f3` (the same commands through `git show 553e224f3:<path>`):
check 1 prints 26 lines and check 2 prints `5` (AGENTS.md 1 + four locales 1 each).
`git diff --name-only <base>..HEAD -- docs-site/content` lists all four locale copies
of each edited page.

### AC-CFR-021 — generated codex MCP table unchanged (maps REQ-CFR-020)

**Given** the tree under test,
**When** `git diff <base> -- internal/codexwiring/configtoml.go .codex/config.toml` and
`grep -c MOAI_KANBAN_ID internal/codexwiring/configtoml.go` run,
**Then** the diff is empty and the grep prints 1.

### AC-CFR-022 — always-loaded budget holds on the merge tree (maps REQ-CFR-021)

**Given** the merge commit `<merge>` of the card branch into local develop and its
develop parent `<merge>^1`,
**When** `go test ./internal/config -run 'TestAlwaysLoadedTokenBudget$' -count=1 -v`
runs on a checkout of each, in the same run,
**Then** both exit 0 and the logged `always-loaded surface` figure on `<merge>` is ≤
the figure on `<merge>^1`; both figures are recorded in `progress.md` §E.2.

### AC-CFR-023 — no agent emission drift (maps REQ-CFR-016)

**Given** the tree under test,
**When** `make agents-emit-check` and
`git diff --stat <base> -- internal/template/templates/.codex/agents internal/template/templates/.claude/agents .claude/agents`
run,
**Then** the check exits 0 and the diff is empty (no agent edit is planned; a
non-empty diff requires a matching `make agents-emit` regeneration and a re-plan note).

### M5 — merge window

### AC-CFR-024 — foreign files preserved after recorded lead confirmation (maps REQ-CFR-019, REQ-CFR-023)

**Given** plan.md §M5 has been executed,
**When** `progress.md` §E.2, the primary checkout's `.moai/reports/t1242/`, and
`git -C <primary>/.claude/worktrees/develop status --porcelain` (run by the lead from
the primary checkout) are read,
**Then** all of the following hold:
1. §E.2 contains an `m5_lead_confirmation:` line with an ISO-8601 timestamp and a
   message or dispatch reference;
2. the patch's sha256 line in §E.2 appears **after** that confirmation line, and the
   patch file's modification time is not earlier than the confirmation timestamp;
3. `<primary>/.moai/reports/t1242/foreign-6.patch` exists, its sha256 equals the
   recorded one, and `grep -c factoryQueueCodexMessage` on it prints ≥ 1;
4. both `factory-cross-host-push-20260924.{md,html}` exist under
   `<primary>/.moai/reports/t1242/`;
5. the develop worktree's porcelain status lists none of the six paths.
A patch with no preceding confirmation line fails item 2 regardless of its content.

## §D.1 Edge cases

- `-f` value that is a verb (`-f cli`): refused (AC-CFR-002), not routed as `cli`.
- `--name` without `-k`: remains the existing usage error (not a kanban refusal).
- A lane key set to whitespace only counts as non-empty for AC-CFR-005 (must be
  removed, not passed).
- Two active runs, one codex: `ResolveActiveRun` fails `AMBIGUOUS_FACTORY` first; the
  refusal applies only once a single run is selected.

## §D.2 Quality gates

- `go vet ./internal/cli ./internal/config ./internal/factorymsg ./internal/hook` and
  `golangci-lint run ./internal/cli/... ./internal/config/...` exit 0 (catches unused
  constants left by M3).
- Coverage of the new code is measured on the changed functions by name, not by the
  package figure: `go test ./internal/cli -coverprofile=<scratch>/c.out -count=1` then
  `go tool cover -func=<scratch>/c.out` — the refusal scan, the lane-key scrub, the
  spawn-command builder, the codex-run join check, and the codex-harness hook guard
  each report ≥ 85%.

## §D.3 Definition of Done

All 25 ACs pass on the merge tree with evidence recorded in `progress.md` §E.2; CI on
the develop push that carries the merge is green; the sync phase has marked the five
partially superseded SPECs and regenerated codemaps (plan.md §I).
