# SPEC-CODEX-FACTORY-RETIRE-001 — acceptance

Every criterion is binary. `<base>` is the run base recorded in plan.md §C.1 (the plan
was authored on `553e224f3`); `<bin>` is a binary built from the tree under test with
`go build -o <scratch>/moai ./cmd/moai`. Grep-count criteria carry a positive control
that must be non-zero, so a zero cannot come from a wrong path or pattern.

## §D AC Matrix

### M1 — refusal and environment

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
launch for the bare form (positive control).

### AC-CFR-004 — tokens after `--` pass through (maps REQ-CFR-004)

**Given** the AC-CFR-003 harness,
**When** `runCodex` receives `-- -f worker -k`,
**Then** exactly one launch is captured and its argv tail ends with `-f worker -k`.

### AC-CFR-005 — direct child carries no lane identity (maps REQ-CFR-006, REQ-CFR-009)

**Given** each of the eleven lane keys set to a distinct non-empty value, plus
`MOAI_HOME` and a sentinel variable `T1242_KEEP=1`,
**When** the direct-launch child environment is built,
**Then** for each of the eleven keys, separately asserted, the environment holds no
entry with a non-empty value; `MOAI_HOME`, `T1242_KEEP=1` and the resolved `CODEX_HOME`
are present; `CLAUDE_CODE_SESSION_ID` and `MOAI_SESSION_PID` are absent.

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
`lead_process_start` are unchanged; under the same environment the cc launcher path
still registers exactly one launch-pending peer (positive control, e.g. the existing
`TestFactoryLauncherRegistersLaunchPendingPeers` family with a claude fixture).

### AC-CFR-008 — joining a codex run is refused (maps REQ-CFR-010)

**Given** exactly one active run `rc` with `lead_backend = 'codex'`,
**When** `moai cc -f worker`, `moai cc -f worker-2` and `moai glm -f worker` entries are
driven through their launch seams,
**Then** each returns a non-zero exit, the message contains `rc`, `codex` and
`moai factory runs --retire`, the worker registry file hashes identically before and
after, and no launch is captured.

### AC-CFR-009 — a non-codex run is joined as before (maps REQ-CFR-011)

**Given** the AC-CFR-008 setup with `lead_backend = 'claude'`,
**When** `moai cc -f worker` is driven through the same seams,
**Then** exactly one launch is captured and one registry slot is claimed.

### AC-CFR-010 — a lead cannot adopt a codex run (maps REQ-CFR-010)

**Given** the AC-CFR-008 run `rc`,
**When** `moai cc -f --factory-run rc` is driven through the launch seams,
**Then** the entry is refused with the AC-CFR-008 message, and
`SELECT lead_backend, lead_pid FROM runs WHERE run_id='rc'` returns the values
recorded before the call.

### M2 — dead entry code

### AC-CFR-011 — codex entry code is gone (maps REQ-CFR-012)

**Given** the tree under test,
**When** `grep -rnwE 'stripCodexKanbanFlag|applyCodexKanbanEntry|codexKanbanEntry|codexKanbanUsageDiag|stripCodexFactoryFlag|applyCodexFactoryEntry|codexFactoryBackend' internal cmd pkg --include='*.go'`
and `ls internal/cli/codex_kanban.go internal/cli/codex_factory.go` run,
**Then** the grep prints 0 lines and `ls` reports both files missing; the same grep
through `git grep -nwE '<pattern>' <base> -- internal` prints ≥ 1 line (positive
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
**When** `grep -c 'syscall.Exec' internal/cli/codex_direct_posix.go` and
`go test ./internal/cli -run 'TestCodex|TestWorktreeLaunchRejectsConcurrentWriter|TestPRMergeCleanupRefusesAnchoredCodexTree' -count=1`
run,
**Then** the grep prints ≥ 1 and the test run exits 0 with at least one `--- PASS` line
from each of the `-w` anchor, `--spawn`, readout (`status`), and local-instructions test
families.

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
file; `git grep -nwE '<pattern>' <base> -- internal` prints ≥ 1 line (positive control).

### AC-CFR-017 — factory schema is byte-identical (maps REQ-CFR-015)

**Given** the tree under test,
**When** `git grep -h 'CREATE TABLE' <base> -- internal/factorymsg internal/homestate ':!*_test.go' | sort | shasum -a 256`
and the same over the working tree (`grep -rh 'CREATE TABLE' internal/factorymsg internal/homestate --include='*.go' --exclude='*_test.go' | sort | shasum -a 256`) run,
**Then** the two digests are equal and the line count is ≥ 9 (positive control).

### AC-CFR-018 — shared factory core still passes (maps REQ-CFR-014)

**Given** the tree under test and a scratch `MOAI_HOME`,
**When** `go test ./internal/factorymsg ./internal/kanban -count=1`,
`MOAI_HOME=<scratch> go test ./internal/hook -run 'Factory' -count=1`,
`go test ./internal/cli -run 'Abandon|FactoryRuns|FactoryMsg|Kanban|Factory' -count=1`
and `<bin> factory handoff abandon-lane --help` run,
**Then** each test command exits 0 with ≥ 1 `ok` package line, the abandon-lane test
file reports ≥ 1 `--- PASS`, and the help command exits 0.

### M4 — docs, config, budget

### AC-CFR-019 — retired wording is gone from every user-facing surface (maps REQ-CFR-018)

**Given** the file set `AGENTS.md internal/template/templates/AGENTS.md.tmpl .claude/rules/moai/core/moai-mcp-tools.md .claude/rules/moai/core/moai-mcp-tools-catalogue.md internal/template/templates/.claude/rules/moai/core/moai-mcp-tools.md internal/template/templates/.claude/rules/moai/core/moai-mcp-tools-catalogue.md docs-site/content/*/advanced/codex-dual-harness.md docs-site/content/*/guides/mcp-server.md`,
**When** `grep -nE 'codex -f|codex -k|`-f` lead|-f lead/agents|Codex lanes?|Codex 레인|레인 셸|Codex レーン|レーンのシェル|Codex 泳道' <file set>`
runs,
**Then** it prints 0 lines; the same grep over `<base>` (`git grep`) prints ≥ 10 lines
(positive control); and `git diff --name-only <base>..HEAD -- docs-site/content` lists
all four locale copies of each edited page.

### AC-CFR-020 — rule mirror parity (maps REQ-CFR-018)

**Given** the tree under test,
**When** `cmp .claude/rules/moai/core/moai-mcp-tools.md internal/template/templates/.claude/rules/moai/core/moai-mcp-tools.md`
and the same for `moai-mcp-tools-catalogue.md` run,
**Then** both exit 0.

### AC-CFR-021 — generated codex MCP table unchanged (maps REQ-CFR-020)

**Given** the tree under test,
**When** `git diff <base> -- internal/codexwiring/configtoml.go .codex/config.toml` and
`grep -c MOAI_KANBAN_ID internal/codexwiring/configtoml.go` run,
**Then** the diff is empty and the grep prints 1.

### AC-CFR-022 — always-loaded budget holds on the merge tree (maps REQ-CFR-021)

**Given** the merge tree (the card branch merged with local develop) and the base figure
from plan.md §C.2,
**When** `go test ./internal/config -run 'TestAlwaysLoadedTokenBudget$' -count=1 -v`
runs on the merge tree,
**Then** it exits 0 and the logged `always-loaded surface` figure is ≤ the same test's
figure on the merge tree's develop parent, measured in the same run.

### AC-CFR-023 — no agent emission drift (maps REQ-CFR-016)

**Given** the tree under test,
**When** `make agents-emit-check` and
`git diff --stat <base> -- internal/template/templates/.codex/agents .codex/agents internal/template/templates/.claude/agents .claude/agents`
run,
**Then** the check exits 0 and the diff is empty (no agent edit is planned; a non-empty
diff requires a matching `make agents-emit` regeneration and a re-plan note).

### M5 — merge window

### AC-CFR-024 — foreign files preserved, then removed (maps REQ-CFR-019)

**Given** the lead's explicit confirmation recorded in the card's progress record,
**When** plan.md §M5 has been executed in the develop worktree,
**Then** `.moai/reports/t1242/foreign-6.patch` exists with its sha256 recorded in
progress.md, `grep -c factoryQueueCodexMessage .moai/reports/t1242/foreign-6.patch`
prints ≥ 1, both `factory-cross-host-push-20260924.{md,html}` exist under
`.moai/reports/t1242/`, and `git -C <develop-worktree> status --porcelain` lists none of
the six paths.

## §D.1 Edge cases

- `-f` value that is a verb (`-f cli`): refused (AC-CFR-002), not routed as `cli`.
- `--name` without `-k`: today a usage error; after retirement it remains a usage error
  (not a kanban refusal) — covered by the existing usage-diag tests in AC-CFR-013.
- A lane key set to whitespace only: counts as non-empty for AC-CFR-005 (must be
  removed, not passed).
- Two active runs, one codex: `ResolveActiveRun` fails `AMBIGUOUS_FACTORY` first; the
  refusal applies only once a single run is selected.

## §D.2 Quality gates

- `go vet ./internal/cli ./internal/config ./internal/factorymsg ./internal/hook`
  and `golangci-lint run ./internal/cli/... ./internal/config/...` exit 0 (catches
  unused constants left by M3).
- Coverage for the new refusal and env-scrub code ≥ 85% (`go test -cover` on the
  touched files' package).

## §D.3 Definition of Done

All 24 ACs pass on the merge tree with evidence recorded in progress.md §E.2; CI on the
develop push that carries the merge is green; the sync phase has marked the two
partially superseded SPECs (plan.md §I).
