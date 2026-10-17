# SPEC-LAUNCHER-ARGV-001 진행 기록

카드 t1748 (lane-22, run tmvgvn) — plan-phase 산출물. 라이프사이클 상태는 spec.md frontmatter의 `status: draft`가 유일한 소리다.

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-10-15
audit_ready: true
tier: M
artifacts: [spec.md, plan.md, acceptance.md]
requirements: 13
acceptance_criteria: 15
direction: "synthesis anchored on Direction 1 — judge the final exec-boundary argv produced by ONE canonical resolver; the guard and the exec consume the same launch view"
evidence: "보존 트리 .moai/worktrees/t1604/ sync-audit-iter8.md (rcpt-649bc5e11cf68dfd88d64c14) + 본 트리 실측 (cc.go/launcher.go/worktree_branch_flag.go 좌표, 선행 가드 부재 확인)"
```

## §E.2 Run-phase Evidence

Run-phase evidence for card t1748, tree `.moai/worktrees/t1748`, branch `WT-launcher-argv-redesign`. Pre-flight measured on HEAD `6e729ba21` (tree SHA). No moai binary was used to judge anything in this section; every measurement is a Go toolchain or `claude --help` invocation.

### Pre-flight (plan §C)

**C-1 backend `--worktree` / `-w` declaration (re-measured).** Command `claude --version` -> `2.1.302 (Claude Code)`; command `claude --help` (exit 0). Verbatim lines carrying the worktree-relevant declarations:

```
  -p, --print                           Print response and exit (useful for
  -w, --worktree [name]                 Create a new git worktree for this
                                        session (optionally specify a name)
  --tmux                                Create a tmux session for the worktree
                                        (requires --worktree). Uses iTerm2
  -c, --continue                        Continue the most recent conversation in
  -h, --help                            Display help for command
  -v, --version                         Output the version number
  -n, --name <name>                     Set a display name for this session
```

Reading: `-w, --worktree [name]` is an optional-value option (the `[name]` synopsis), next to the boolean `-p, --print`, `-c, --continue`, `-h`, `-v` and the required-value `-n, --name <name>`. This is the commander synopsis layout the repo's option model already parses (`lane_resume_model.go`), and it is consistent with the audited model of REQ-LA-003 (optional value, attached `-w<value>`, a short boolean followed by the value option in one cluster `-pw<value>`). It contradicts nothing, so no blocker report.

Gap (named, not substituted): the help text does not state last-wins for a repeated `--worktree`, the `--worktree=` empty-value behaviour, or the cluster decomposition; those rest on commander's documented parsing and on the audited iter8 record, not on a measurement of a running backend. A running `claude -w ...` would create a worktree, so it was not executed. The resolver's selection model follows the declaration above; the empty equals form `--worktree=` is read as an empty value that names no tree (fail-open, REQ-LA-008).

**C-2 coordinate re-observation (HEAD `6e729ba21`).** plan.md was not edited; the new coordinates:

| Symbol | plan.md | observed |
|---|---|---|
| `runLaunchClaude` | launcher.go:601 | launcher.go:601 (consumption loop 658-695, `buildArgs` 744-761, `injectChainNodeForLaunch` 879, exec 880) |
| `normalizeWorktreeFlag` | launcher.go:917 | launcher.go:932 |
| `resolveWorktreeL2Path` | launcher.go:992 | launcher.go:1007 |
| `worktreeFlagValue` | launcher.go:1030 | launcher.go:1045 |
| cc.go leader / lane cases | 194-231 / 232-280 | 195-231 / 232-275 (`laneJoinChildArgv` 273) |
| cc.go common tail | 282-312 | `resolveWorktreeL2Path` 282, `resolveWorktreeExistingBranch` 289, `ccWorktreeWriterPrecheck` 296, `seedAdmittedWorktreeHooks` 302, `normalizeWorktreeFlag` 304, `launch` 312 |
| glm.go common tail | 334/339/347/348 | 334/339/347/348, plus a second `normalizeWorktreeFlag` at 367, `unifiedLaunch` 373; leader 261-294, lane 295-331 |
| `ccWorktreeWriterPrecheck` | session_worktree.go:1012 | session_worktree.go:1014 (`worktreeFlagValue` 1015) |
| `resolveWorktreeExistingBranch` / `splitWorktreeBranchFlag` / `requireShortWorktreeName` | worktree_branch_flag.go:57,86,121 | 57 / 86 / 120 (`worktreeFlagValue` 121) |
| `injectChainNodeForLaunch` / `extractWorktreeLaunchTarget` | launch_chain.go:34,58 | 34 / 58 |
| `seedAdmittedWorktreeHooks` | codex_hooks_seed.go:62 | 62 (`worktreeFlagValue` 63) |
| `laneJoinChildArgv` | lane_resume.go | lane_resume.go:95 (0 `worktreeFlagValue` calls) |
| `cardScopeFromBranch` | codex_review_scope.go:170 | 170 |
| claude option model | - | `activeClaudeOptionModel` / `claudeOptionModelSnapshot`, lane_resume_model.go:80 / 103 |

Two observations the plan did not carry: every leader-case argv mutation (`appendLeaderName`, `prepareFactorySettings`, `appendCrossSessionSettings`) appends at the END of the argv (plain `append`), while the lane join inserts `--name` before the separator only when no name is supplied (never, for `-l`); and `internal/gitenv` exists for the scrubbed git environment the anchor lookup needs.

**C-3 baselines (HEAD `6e729ba21`, before any change).**

- `go build ./...` -> exit 0, no output.
- `GOOS=windows GOARCH=amd64 go build ./...` -> exit 0, no output.
- `go test ./internal/cli/ -run 'Worktree' -count=1` -> `ok  github.com/modu-ai/moai-adk/internal/cli  266.901s` (exit 0, zero failures).
- `go test ./internal/cli/ -run 'TestLeader|TestLauncher' -count=1` -> `ok  github.com/modu-ai/moai-adk/internal/cli  37.891s` (exit 0, zero failures).
- No pre-existing failure was found, so no new-versus-baseline classification was needed.

**C-4** the preserved tree `.moai/worktrees/t1604/` was only read (`factory.go` guard region 440-870, `cc.go`/`glm.go` placement); nothing was written there.

HEAD/branch re-read: `git rev-parse --short HEAD` -> `6e729ba21`; `git branch --show-current` -> `WT-launcher-argv-redesign`.

### Coordinator addendum: D9 / D10 / D11 baseline (orchestrator message, measured before any refactor)

Sibling card t1783's claims were treated as claims. They were measured by driving the real `runCC` / `runGLM` entry (non-leader) through `runLaunchClaude` to the exec seam, recording per spelling: the launch error, the `--branch` materializations, the argv at the launch seam and at the exec seam, the chain-ledger nodes, whether `MOAI_CHAIN_NODE_ID` reached the child env, the tree the cc writer precheck read (observed through its refusal), and the tree the hook seeder resolved (observed at the `sessionWorktreeSeedCodexHooks` seam). Harness: `internal/cli/launcher_worktree_stage_test.go` (helper) and `launcher_worktree_stage_cases_test.go` (14 rows x cc/glm arms x 3 tree shapes, pinned). Tree SHA of the measurement: `6e729ba21` (the two new test files are the only additions; no production file differs).

First measurement (exploratory probe, since deleted; command `go test ./internal/cli/ -run '^TestLauncherWorktreeStageProbe$' -count=1 -v`, exit 0). Verbatim probe lines for the claim-bearing spellings (cc arm; the glm arm differs only by having no writer precheck, so its dirs-nongit run launches):

```
PROBE arm=cc kind=bare args=["-w" "--branch" "main" "plain"] => {Err: Materialize:[plain@main] LaunchArgs:[--worktree plain] ExecTail:[--worktree plain] ChainTrees:[<root>/.claude/worktrees/plain] EnvChainNode:true PrecheckTree: SeedTree:}
PROBE arm=cc kind=dirs-nongit args=["-w" "--branch" "main" "plain"] => {Err:writer-anchored Materialize:[plain@main] LaunchArgs:[] ExecTail:[] ChainTrees:[] EnvChainNode:false PrecheckTree:.claude/worktrees/plain SeedTree:}
PROBE arm=cc kind=dirs-git args=["-w" "--branch" "main" "plain"] => {Err: Materialize:[plain@main] LaunchArgs:[--worktree plain] ExecTail:[--worktree plain] ChainTrees:[.claude/worktrees/plain] EnvChainNode:true PrecheckTree: SeedTree:.claude/worktrees/plain}
PROBE arm=cc kind=bare args=["--worktree=" "plain"] => {Err: Materialize:[] LaunchArgs:[--worktree plain] ExecTail:[--worktree plain] ChainTrees:[<root>/.claude/worktrees/plain] EnvChainNode:true PrecheckTree: SeedTree:}
PROBE arm=cc kind=dirs-nongit args=["--worktree=" "plain"] => {Err: Materialize:[] LaunchArgs:[--worktree plain] ExecTail:[--worktree plain] ChainTrees:[.claude/worktrees/plain] EnvChainNode:true PrecheckTree: SeedTree:}
PROBE arm=cc kind=dirs-git args=["--worktree=" "plain"] => {Err: Materialize:[] LaunchArgs:[--worktree plain] ExecTail:[--worktree plain] ChainTrees:[.claude/worktrees/plain] EnvChainNode:true PrecheckTree: SeedTree:}
PROBE arm=cc kind=bare args=["--" "--worktree" "plain" "PROMPT"] => {Err: Materialize:[] LaunchArgs:[-- --worktree plain PROMPT] ExecTail:[--worktree plain PROMPT] ChainTrees:[<root>/.claude/worktrees/plain] EnvChainNode:true PrecheckTree: SeedTree:}
PROBE arm=cc kind=bare args=["-w" "-b" "plain"] => {Err: Materialize:[] LaunchArgs:[--worktree -b plain] ExecTail:[--worktree plain] ChainTrees:[<root>/.claude/worktrees/plain] EnvChainNode:true PrecheckTree: SeedTree:}
PROBE arm=cc kind=bare args=["--" "-w" "plain" "PROMPT"] => {Err: Materialize:[] LaunchArgs:[-- -w plain PROMPT] ExecTail:[-w plain PROMPT] ChainTrees:[] EnvChainNode:false PrecheckTree: SeedTree:}
PROBE arm=cc kind=bare args=["--" "-w=plain" "PROMPT"] => {Err: Materialize:[] LaunchArgs:[-- -w=plain PROMPT] ExecTail:[-w=plain PROMPT] ChainTrees:[] EnvChainNode:false PrecheckTree: SeedTree:}
PROBE arm=cc kind=bare args=["-w" "plain" "--worktree" "<elsewhere>"] => {Err: Materialize:[] LaunchArgs:[--worktree plain --worktree <elsewhere>] ExecTail:[--worktree plain --worktree <elsewhere>] ChainTrees:[<root>/.claude/worktrees/plain] EnvChainNode:true PrecheckTree: SeedTree:}
```

Pinned form (the same measurements as asserted cells), command `go test ./internal/cli/ -run 'TestLauncherWorktreeStageCharacterization|TestLauncherWorktreeDuplicateOptionFirstValue' -count=1 -v`, exit 0, tree `6e729ba21` + the two new test files:

```
--- PASS: TestLauncherWorktreeStageCharacterization (1.27s)   [28 subtests: 14 rows x cc, glm; every one --- PASS]
--- PASS: TestLauncherWorktreeDuplicateOptionFirstValue (0.09s)
    --- PASS: TestLauncherWorktreeDuplicateOptionFirstValue/cc (0.06s)
    --- PASS: TestLauncherWorktreeDuplicateOptionFirstValue/glm (0.03s)
ok  	github.com/modu-ai/moai-adk/internal/cli	2.581s
```

Verdict per claim:

- **D9 REPRODUCED.** `-w --branch main plain`: stage A (raw argv) reads no value, stage B (after the `--branch` strip) and stage C read `plain` -> the precheck refuses on `plain`, the seeder resolves `plain`, the chain node is on `plain`, materialize is `plain@main`. `--worktree= plain`: stages A and B read the empty value (the precheck does not refuse, the seeder resolves nothing) while the normalized stream carries `plain` -> a chain node on `plain` and `MOAI_CHAIN_NODE_ID` in the child env. Cells: rows "bare -w before --branch" and "empty equals form" of `TestLauncherWorktreeStageCharacterization`, both arms.
- **D10 REPRODUCED.** `-- --worktree plain PROMPT` and `-w -b plain`: stage C reads the consumption loop's passThrough (the separator gone, `-b` consumed), so a chain node on tree `plain` is created and `MOAI_CHAIN_NODE_ID` reaches the child env, while stages A and B read nothing / a bare flag. Cells: rows "tail --worktree after separator" and "consumed option between flag and value", both arms.
- **D11 REPRODUCED.** `-- -w plain PROMPT` and `-- -w=plain PROMPT`: no chain node, no `MOAI_CHAIN_NODE_ID`, and the tail reaches the exec seam unmodified (`-w plain PROMPT`, `-w=plain PROMPT`). The attached `-wplain` and the tail cluster `-pwplain` are likewise invisible to every non-guard reader. Cells: rows "tail short -w after separator", "tail short equals after separator", "attached short form", "tail short cluster", both arms.
- **D8 row 14 (already planned) REPRODUCED on the baseline:** `-w plain --worktree <out-of-prefix>` is allowed, all five readers read `plain`, both options reach the exec seam. Cell: `TestLauncherWorktreeDuplicateOptionFirstValue`.

Consequence for the design (SPEC §2 outranks the one-field wording of plan §F-M1 / §A.4): the five readers do not read one value. The view therefore carries the three stage values the readers consume today — stage A (`resolveWorktreeL2Path`), stage B (`requireShortWorktreeName`, `ccWorktreeWriterPrecheck`, `seedAdmittedWorktreeHooks`), stage C (`extractWorktreeLaunchTarget`) — each computed by the same reader function applied to the same stage input, inside the resolver. No SPEC body edit and no blocker: the REQ text (REQ-LA-001/002) and §2 are compatible with a view that carries per-stage consumer fields; only the plan's "single first-value field" phrasing is narrower than the measured behaviour, and the plan was not edited.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §G Override and Refusal Record

- 2026-10-17T14:48:03Z SPEC-LAUNCHER-ARGV-001 ceiling-refusal outcome=hold reasons="plan-audit ceiling reached (round count 4 >= tier ceiling 2); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1748/.moai/reports/t1748/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1748/.moai/reports/t1748/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1748/.moai/reports/t1748/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1748/.moai/reports/t1748/plan-audit.md
- 2026-10-17T14:50:06Z SPEC-LAUNCHER-ARGV-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 4 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1748/.moai/reports/t1748/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1748/.moai/reports/t1748/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1748/.moai/reports/t1748/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1748/.moai/reports/t1748/plan-audit.md
- 2026-10-17T14:50:23Z SPEC-LAUNCHER-ARGV-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 4 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1748/.moai/reports/t1748/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1748/.moai/reports/t1748/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1748/.moai/reports/t1748/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1748/.moai/reports/t1748/plan-audit.md
- 2026-10-17T14:52:04Z SPEC-LAUNCHER-ARGV-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 4 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1748/.moai/reports/t1748/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1748/.moai/reports/t1748/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1748/.moai/reports/t1748/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1748/.moai/reports/t1748/plan-audit.md
