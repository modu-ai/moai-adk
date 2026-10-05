# SPEC-CODEX-FACTORY-RETIRE-001 — acceptance

Every criterion is binary. `<base>` is the `run_base:` value the run lane records in
`progress.md` §E.2 (plan.md §C.1); the plan was authored on `553e224f3`. `<bin>` is a
binary built from the tree under test with `go build -o <scratch>/moai ./cmd/moai`.
Every grep-count criterion carries a positive control that must be non-zero, so a zero
cannot come from a wrong path or pattern. Every `go test` criterion runs with `-v` and
names the tests it expects, so an empty selection cannot pass (`[no tests to run]` or
`[no test files]` fails the criterion).

**Card-scoped change guard.** "This card did not change path P" is measured with the
three-dot diff `git diff --exit-code develop...HEAD -- P`, evaluated on the card branch
**before** it is merged into develop (after any absorption of develop). Three-dot diff
compares HEAD with `merge-base(develop, HEAD)`, so commits other cards landed on develop
never enter the range, while the card's own commits and its absorb-merge resolution do.
It is never measured as `git diff <literal SHA>`. Controls are pinned to fixed SHAs —
the historical absorb merge `e48f290da` (card WT-worktree-moai-root absorbing develop
`acc1c2289`; fork point `35ab8cff3`), whose tree does not move:
- scope: `git diff --name-only 35ab8cff3 e48f290da` lists 43 files (card + absorbed
  foreign changes), while `git diff --name-only acc1c2289...e48f290da` lists 21 (the
  card's own contribution);
- foreign-commit exclusion: `git diff --exit-code --quiet 35ab8cff3 e48f290da -- internal/homestate`
  exits 1 (another card changed that path), while
  `git diff --exit-code --quiet acc1c2289...e48f290da -- internal/homestate` exits 0 —
  the literal-base form is polluted, the card-scoped form is not;
- card-commit detection: `git diff --exit-code --quiet acc1c2289...e48f290da -- internal/cli/mcp_worktree_root.go`
  exits 1 — the same form fires on a path the card did change.

The same property was also observed live on this card's tree (develop `c630de892`,
6 non-merge commits past `553e224f3`: literal base 19 foreign `internal/cli` files vs
card range 0), but live develop figures drift as other cards land; only the exit-code
property above is load-bearing, not any live count.

After the merge the range is empty by construction; post-merge evidence is tree identity
(`git rev-parse <merge>^{tree}` equal to the re-measured tree), not this guard.

## §D.0 Red at arrival (measured on `553e224f3`)

| AC | Why it is red before the run | Milestone that flips it |
|---|---|---|
| 001, 002 | `-k`/`-f` are accepted and enter kanban/factory mode (`codex_launcher.go:711-745`) | M1 |
| 005, 006 | `codexChildEnv` drops only 2 keys; spawn forwards 9 lane keys (`:296-310`) | M1 |
| 007 | `codex_direct_posix.go:39` registers launch-pending under lane env | M1 |
| 008, 010 | No backend check on join/lead run selection (`factory.go:245`) | M1 |
| 011, 012, 016 | The symbols exist (positive controls below) | M2, M3 |
| 014 | `internal/cli/codex_factory_test.go` still exists (the `ls … missing` clause is red) | M2 |
| 020 | The fixture carries no explicit target-existence assertion; `pathExists(target)` guards skip silently (`factory_handoff_abandon_test.go:140,159`) | M3 |
| 019 | 26 line matches + 5 folded matches on the file set | M4 |
| 025 | `registerFactoryHookPeer` is harness-agnostic (`factory_messages.go:53-116`) | M1 |
| 003, 004, 009, 013, 015, 017, 018, 021-023 | Preservation guards: green before and must stay green | — (mutant probes stated in each) |
| 024 | Foreign files still present in the develop worktree (lead report). Regression-guard, not release-blocking (see the AC) | M5 |

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
`lead_process_start` are unchanged; and, as the positive control,
`go test ./internal/cli -run '^TestFactoryLauncherRegistersLaunchPendingPeers$' -count=1 -v`
prints `--- PASS: TestFactoryLauncherRegistersLaunchPendingPeers` — that test calls the
shared helper `registerFactoryLaunchPending` directly (`factory_mixed_test.go:18-60`),
so it proves the **shared helper still registers** under a lane environment; it does not
exercise the cc launcher path, and its run fixture's backend is an opaque string (D4).

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

**Given** an isolated `MOAI_HOME` and an active run `r1` (`lead_backend = 'claude'`),
and each of two environment shapes:
(a) worker shape — `MOAI_KANBAN_ID=r1`, `MOAI_FACTORY_WORKER=worker-1`;
(b) lead shape — `MOAI_KANBAN_ID=r1`, `MOAI_FACTORY_WORKERS=2`, `MOAI_FACTORY_WORKER`
unset (the branch that registers slot `lead`, `factory_messages.go:59-65`),
**When** for each shape the session-start and user-prompt-submit hooks run once with
`--harness codex` and once without it, each with a distinct session id,
**Then** after the `--harness codex` runs the peer row count is 0 for `(r1, worker-1)` in
shape (a) and 0 for `(r1, lead)` in shape (b), and no endpoint was rotated; after the
Claude runs each count is 1 (positive control). Mutants: removing the codex-harness
guard turns both first counts to 1; a guard that only fires when `MOAI_FACTORY_WORKER`
is set leaves shape (b) at 1 and fails.

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
`go test ./internal/cli -run '^(TestWorktreeLaunchRejectsConcurrentWriter|TestCodexWorktreeAnchorLockAndBase|TestCodexSpawnAnchorsToPanePID|TestCodexLaunchVerb_StatusStaysTheReadout|TestCodexLocalInstructions_DirectSpawnAndAppSharePrefix|TestCodexTask_ForegroundReturnsOutput|TestSessionMsgSendPollAckHandlers|TestCodexRoleLoadNegativeControl|TestCodexAuditVerbRunsInCallerWorktree)$' -count=1 -v`
run,
**Then** the grep prints ≥ 1; the build-tag lines are `//go:build !windows` and
`//go:build windows` respectively; the test run exits 0 and prints exactly one
top-level `--- PASS: <name>` line for each of the nine named tests (9 lines;
`TestCodexAuditVerbRunsInCallerWorktree` in `codex_audit_launch_test.go` covers
`moai codex audit` / `codex_audit`).

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

**Given** the card branch before its merge into develop (after any absorption),
**When** `git diff --exit-code develop...HEAD -- internal/factorymsg internal/homestate ':!*_test.go'`
runs (the card-scoped guard defined at the top of this file),
**Then** it exits 0 with empty output. Because it is a whole-file diff of every
production source in both packages, it covers every multi-line `CREATE TABLE` body
(including `runs.lead_backend`, `internal/homestate/factory.go:36`), every `ALTER TABLE`
(`factory.go:226-228`), and every index statement — and it is blind to other cards'
changes to those packages on develop. Controls: the two card-scoped-guard controls
above, plus `git grep -c 'lead_backend' 553e224f3 -- internal/homestate/factory.go`
prints ≥ 1 (the guarded column is inside the swept files).

### AC-CFR-018 — shared factory core still passes (maps REQ-CFR-014)

**Given** the tree under test and a scratch `MOAI_HOME`,
**When** `go test ./internal/factorymsg ./internal/kanban -count=1 -v`,
`MOAI_HOME=<scratch> go test ./internal/hook -run 'Factory' -count=1 -v`,
`go test ./internal/cli -run '^(TestFactoryLaneHandoffOperatorAbandon|TestFactoryLeadNoticeUsesOperationalStatus|TestFactoryRunSelectionAtomicSlotsAndArgv)$' -count=1 -v`
and `<bin> factory handoff abandon-lane --help` run,
**Then** the first command prints one `ok` line per package, at least one `--- PASS`
line, and no `[no test files]`; the hook run exits 0,
contains `--- PASS: TestFactoryLaneHandoffInteractiveStateMachine`, and does not print
`[no tests to run]`; the cli run prints `--- PASS:` for each of the three named tests;
and the help command exits 0.

### AC-CFR-020 — the rebuilt abandon fixture still builds a real target (maps REQ-CFR-017, REQ-CFR-014)

**Given** the tree under test after M3,
**When** `go test ./internal/cli -run '^TestFactoryLaneHandoffOperatorAbandon$' -count=1 -v`
runs,
**Then** it exits 0 and prints four `--- PASS: TestFactoryLaneHandoffOperatorAbandon/`
subtest lines, and the fixture asserts before each abandon call:
- for the three states that have a target — `WT_READY`, `SWITCH_PENDING_INTERACTIVE`,
  `SWITCH_PENDING_HEADLESS` — that the target worktree path exists and the target
  branch ref resolves; the existing `if pathExists(target)` guards
  (`factory_handoff_abandon_test.go:140,159`) become unconditional assertions for these
  three, so a missing target fails instead of skipping;
- for `RESERVED` — explicitly exempt from the existence assertion, because by design no
  target exists yet (`handoffPointReserved = "reserved" // reservation committed, target not created`,
  `factory_lane_handoff_recover.go:27` on `553e224f3`) — that the
  target path does **not** exist.
Mutant (run once in M3 and recorded in `progress.md` §E.2): a fixture that skips the
worktree/branch step makes this command exit non-zero.

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
`git diff --name-only develop...HEAD -- docs-site/content` (card-scoped) lists all four locale copies
of each edited page.

### AC-CFR-021 — generated codex MCP table unchanged (maps REQ-CFR-020)

**Given** the card branch before its merge into develop,
**When** `git diff --exit-code develop...HEAD -- internal/codexwiring/configtoml.go .codex/config.toml`
and `grep -c MOAI_KANBAN_ID internal/codexwiring/configtoml.go` run,
**Then** the diff exits 0 and the grep prints 1 (controls: the card-scoped-guard pair
at the top of this file).

### AC-CFR-022 — always-loaded budget holds on the merge tree (maps REQ-CFR-021)

**Given** the merge commit `<merge>` of the card branch into local develop and its
develop parent `<merge>^1`,
**When** `go test ./internal/config -run 'TestAlwaysLoadedTokenBudget$' -count=1 -v`
runs on a checkout of each, in the same run,
**Then** both exit 0 and the logged `always-loaded surface` figure on `<merge>` is ≤
the figure on `<merge>^1`; both figures are recorded in `progress.md` §E.2.

### AC-CFR-023 — no agent emission drift (maps REQ-CFR-016)

**Given** the card branch before its merge into develop,
**When** `make agents-emit-check` and
`git diff --exit-code --stat develop...HEAD -- internal/template/templates/.codex/agents internal/template/templates/.claude/agents .claude/agents`
run,
**Then** the check exits 0 and the diff exits 0 (no agent edit is planned; a non-zero
diff requires a matching `make agents-emit` regeneration and a re-plan note). Controls:
the card-scoped-guard pair at the top of this file.

### M5 — merge window

### AC-CFR-024 — foreign files preserved after recorded lead confirmation (maps REQ-CFR-019, REQ-CFR-023)

**Given** plan.md §M5 has been executed,
**When** `progress.md` §E.2, the primary checkout's `.moai/reports/t1242/`, and
`git -C <primary>/.claude/worktrees/develop status --porcelain` (run by the lead from
the primary checkout) are read,
**Then** all of the following hold:
1. the lead-authored confirmation file `<primary>/.moai/reports/t1242/m5-lead-confirm.md`
   exists and carries the lead's message id (or the dispatch reference) and a verbatim
   quote of the confirmation; §E.2 contains an `m5_lead_confirmation:` line that cites
   that file path and the same message id, recorded **before** plan.md §M5 step 1;
2. the patch's sha256 line in §E.2 appears after that confirmation line;
3. `<primary>/.moai/reports/t1242/foreign-6.patch` exists, its sha256 equals the
   recorded one, and `grep -c factoryQueueCodexMessage` on it prints ≥ 1;
4. both `factory-cross-host-push-20260924.{md,html}` exist under
   `<primary>/.moai/reports/t1242/`;
5. the develop worktree's porcelain status lists none of the six paths.
A patch with no preceding confirmation citation fails item 2 regardless of its content.

**Classification: regression-guard, not release-blocking**
(`verification-completeness.md` §2.1, undecidable disposition). The ordering in items 1-2
is written by the lane that performs the step, and a file's modification time can be
changed by `touch` or a copy, so the order of events is not mechanically provable after
the fact; mtime is advisory only and is not a pass condition. The binding evidence is
the lead-authored confirmation file and the lead's own read of the develop worktree
(item 5), which the lead performs before approving the merge.

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

All 25 ACs pass with evidence recorded in `progress.md` §E.2; CI on the develop push
that carries the merge is green; the sync phase has marked the five partially
superseded SPECs and regenerated codemaps (plan.md §I).

The card-scoped guards (`develop...HEAD`, used by AC-CFR-017, the locale listing in
AC-CFR-019, AC-CFR-021 and AC-CFR-023) are judged **pre-merge**, on the card branch after its last absorption
of develop. On the merge tree they are **not re-run** — after the merge the range is
empty and the guard would pass vacuously. Their verdict is carried to the merge tree by
tree identity: `git rev-parse <merge>^{tree}` equals `git rev-parse <absorbed WT tip>^{tree}`,
where the absorbed WT tip is the card commit the guards were judged on. Every other AC
is judged on the merge tree itself.
