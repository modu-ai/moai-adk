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

### Run commits (M1-M3; M4 is outside this delegation)

| Commit | Content |
|---|---|
| `538af8d7c` | baseline characterization (D8-D11 cells GREEN on `6e729ba21`), pre-flight record, spec.md `draft -> in-progress` |
| `a022ada95` | M1: `launch_view.go` resolver, readers rewired to stage values, `TestLauncherAnchorFinalArgv/view_exec_byte_equality`, resolver unit tests |
| `87abe2ef7` | M2: guard (`factory_anchor_guard.go`), audit-log hygiene + nofollow build-tag split, cc/glm leader placement, five M2 test families, existing leader-fixture cwd adaptation |
| `f4676475b` | M3: `TestLaunchViewSingleParse` (source scan, parallel-reader absence, appended-flag neutrality) |
| `f08294446` | this M1-M3 evidence record |
| lint fix (next commit) | dropped the unused `requireShortWorktreeName` wrapper from `launch_view.go` (zero callers in the module, test or non-test; `golangci-lint --new-from-rev=6e729ba21` -> `unused`, now 0 issues, exit 0); `AC-LA-014` selector re-run exit 0 (180 top-level PASS, 0 FAIL), `TestLaunchViewSingleParse` exit 0 |

### Design decisions taken inside the plan's latitude (named, not hidden)

1. **Three stage reads, not one first-value field** (D9/D10/D11, measured above). `launchView` carries `StageA`, `StageB` and `Consumed.ChainTarget`, each computed by the reader function that always computed it, on the stage input it always read. The backend (last-wins) `Selection` is a separate field only the guard reads.
2. **`FinalArgv` is the backend argv minus the launcher-injected `--permission-mode` / `--model` / `--continue`.** Those three are resolved later from profile state, never from argv, and sit before the pass-through; AC-LA-005 compares exec argv with them stripped. For a leader launch the launcher also appends `--name` / `--settings` / `--effort` after the guard has judged; `appended_flags_are_worktree_neutral` pins that no such suffix changes any stage read or the selection on any §D.4 spelling, and the equality cell strips exactly those tail flags.
3. **The resolver is one function with two call sites per launch** (the guard before mutation, the shared tail), plus `consumeLaunchArgs` at the exec boundary. Equality of what the guard judged and what exec receives is pinned by AC-LA-005 and the neutrality cell, not by sharing one object across the mutation window (the mutations append to the argv and the seams are `[]string`-typed).
4. **Walk helpers moved into `launch_view.go` unchanged** (`worktreeFlagValue`, `normalizeWorktreeFlag`, `splitWorktreeBranchFlag`, `extractWorktreeLaunchTarget`); the old argv-taking entry points that the codex launcher and existing tests call (`resolveWorktreeL2Path`, `resolveWorktreeExistingBranch`, `seedAdmittedWorktreeHooks`, `injectChainNodeForLaunch`) live there as thin wrappers over value-taking cores. glm's second `normalizeWorktreeFlag` (idempotent) is gone with it.
5. **Claude option snapshot gained `-c`, `-h`, `-v` as booleans** (`claude --help`, 2.1.302): the short-cluster decomposition needs to know `-cw<tree>` is `-c` then `-w<tree>`. The resume interpreters treat boolean and unknown tokens identically, so no resume outcome changes (all `lane_resume*` tests re-run GREEN).
6. **Existing leader-entry fixtures now enter their project directory** (`netLeaderFixture`, `codexLedRun`, three cells in `factory_test.go`): the guard reads the working directory, and a lane runs the suite from a card worktree, so ten existing cells failed there with the guard's own refusal. Assertions are unchanged; same adaptation card t1604 made.
7. **`leaderAnchorGuardProbe` test seam** (no-op in production), same pattern as `leaderAnchorAuditOpenRace`: it lets the placement cell read the environment, run records and settings artifacts at the instant the guard starts.

### M1 results (tree `a022ada95`)

- `go test ./internal/cli/ -run 'Worktree|TestLaneJoinChildArgv' -v -count=1` -> exit 0, `ok ... 215.399s`, 180 top-level `--- PASS`, 0 `--- FAIL`; `--- PASS: TestLaneJoinChildArgv (0.00s)` and `--- PASS: TestLauncherWorktreeDuplicateOptionFirstValue (0.28s)` present (AC-LA-014).
- `go test ./internal/cli/ -run '^TestLauncherAnchorFinalArgv$/view_exec_byte_equality$' -v -count=1` -> exit 0; 14 rows x cc/glm, every row `--- PASS`.
- Mutants against the resolver / readers (the characterization cells must see each; every one reverted, `git status` clean afterwards). Command for all four: `go test ./internal/cli/ -run 'TestLauncherWorktreeStageCharacterization|TestLauncherWorktreeDuplicateOptionFirstValue|TestLaunchViewStageReads' -count=1`; tree `a022ada95` + the one-line mutant; each exit 1:
  - M1-mut1 `v.StageB = v.StageA` (one shared first-value field): `--- FAIL: TestLaunchViewStageReads/bare_-w_before_--branch`, `--- FAIL: TestLauncherWorktreeStageCharacterization/{cc,glm}/bare_-w_before_--branch`, message `[A=bare B=plain C=plain] launch refused: "branch-name"`. Red because stage A reads no value for `-w --branch main plain`, so the shared field loses the name the `--branch` path needs (D9).
  - M1-mut2 `Consumed.ChainTarget = StageB.Value`: `--- FAIL: TestLaunchViewStageReads/{empty_equals_form,tail_--worktree_after_separator,consumed_option_between_flag_and_value}`. Red because stage C differs from stage B on exactly the D9/D10 spellings.
  - M1-mut3 (`launcher.go`) chain populator fed `selectBackendWorktree(passThrough).Value`: `--- FAIL: .../cc/duplicate_options_first_value` (`chain nodes = [".../elsewhere"], want [".claude/worktrees/plain"]`), and `.../tail_short_-w_after_separator`, `.../tail_short_equals_after_separator`, `.../attached_short_form` (`chain nodes = [".claude/worktrees/plain"], want []`, `MOAI_CHAIN_NODE_ID in child env = true, want false`). Red because the chain reader would start reading the tail forms (D11) and the last option (D8).
  - M1-mut4 (`cc.go`) writer precheck reads the backend selection: `--- FAIL: .../cc/duplicate_options_first_value` (`precheck = err "" tree "", want writer-anchored on ".claude/worktrees/plain"`), `.../empty_equals_form`, `.../tail_--worktree_after_separator`, `.../consumed_option_between_flag_and_value` (`unexpected refusal "writer-anchored"`). Red because the precheck would read a different tree than today's first value (D8-D10).

### M2 results (tree `87abe2ef7`) — RED records per failure family (AC-LA-015)

Mutant protocol of acceptance.md §D.3: the named mutant is applied to the pinned tree, the cells observed RED, the mutant reverted (`git status --short` shows only the harness prune-state file), and the same command observed GREEN. Green-flip milestone for every cell below: **M2** (the guard + resolver at `87abe2ef7` flip them).

**Families (a)(b)(c), mutant: the resolver reads the selection BEFORE consumption** (`v.Selection = selectBackendWorktree(entry)` instead of `v.FinalArgv`).
- Command: `go test ./internal/cli/ -run '^TestLauncherAnchorFinalArgv$/(consumed_option_shadowing|multi_option_last_wins|bundled_short_cluster)$' -v -count=1`
- Tree SHA: `87abe2ef7` + the one-line mutant. Exit code: **1**.
- Verbatim stdout (filtered to the verdict lines):

```
--- FAIL: TestLauncherAnchorFinalArgv (13.55s)
    --- FAIL: TestLauncherAnchorFinalArgv/consumed_option_shadowing (3.89s)
        --- FAIL: TestLauncherAnchorFinalArgv/consumed_option_shadowing/cc (2.27s)
        --- FAIL: TestLauncherAnchorFinalArgv/consumed_option_shadowing/glm (1.62s)
    --- FAIL: TestLauncherAnchorFinalArgv/multi_option_last_wins (4.83s)
        --- FAIL: TestLauncherAnchorFinalArgv/multi_option_last_wins/cc (2.56s)
        --- FAIL: TestLauncherAnchorFinalArgv/multi_option_last_wins/glm (2.27s)
    --- FAIL: TestLauncherAnchorFinalArgv/bundled_short_cluster (4.83s)
        --- FAIL: TestLauncherAnchorFinalArgv/bundled_short_cluster/cc (2.54s)
        --- FAIL: TestLauncherAnchorFinalArgv/bundled_short_cluster/glm (2.29s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	14.861s
factory_anchor_guard_cells_test.go:48: cc: the anchored leader launch was not refused (launched=true)
factory_anchor_guard_cells_test.go:62: glm: the anchored leader launch was not refused (launched=true)
factory_anchor_guard_cells_test.go:85: cc -f -- -pwWT-card PROMPT: the anchored leader launch was not refused (launched=true)
```

- Why red (right reason): the raw entry argv shows `-w` followed by the not-yet-consumed `-b` (bare), `--worktree` followed by the not-yet-removed `--branch` (bare), and a leading `--` that ends the scan before the cluster — so the guard sees no card tree and the leader launches. This is exactly the iter8 shape: the check does not reflect the arguments the launch CONSUMES.
- GREEN pair (mutant reverted): the same command -> exit 0, `--- PASS: TestLauncherAnchorFinalArgv/consumed_option_shadowing`, `.../multi_option_last_wins`, `.../bundled_short_cluster` (each with cc and glm), `ok ... 12.594s`; and again at `f4676475b` in the AC matrix below.

**Family (d), mutant: the run-record counter reverted to the legacy probe** (`launcherRunCount` returns 0 when `os.Stat(factory.BacklogPathForRoot(root))` fails).
- Command: `go test ./internal/cli/ -run '^TestLauncherAnchorRunCountInstrument$' -v -count=1`. Tree: `87abe2ef7` + the test-helper mutant. Exit code: **1**.
- Verbatim stdout (verdict lines):

```
    factory_anchor_guard_cells_test.go:388: run records 0 -> 0 after a recorded leader start, want +1: the counter cannot see a record
    factory_anchor_guard_cells_test.go:388: run records 0 -> 0 after a recorded leader start, want +1: the counter cannot see a record
--- FAIL: TestLauncherAnchorRunCountInstrument (3.57s)
    --- FAIL: TestLauncherAnchorRunCountInstrument/cc/positive_control_counter_moves (1.29s)
    --- PASS: TestLauncherAnchorRunCountInstrument/cc/refusal_leaves_the_counter_unmoved (0.62s)
    --- FAIL: TestLauncherAnchorRunCountInstrument/glm/positive_control_counter_moves (0.99s)
    --- PASS: TestLauncherAnchorRunCountInstrument/glm/refusal_leaves_the_counter_unmoved (0.66s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	4.746s
```

- Why red: the absence cell (`refusal_leaves_the_counter_unmoved`) still PASSES under the mutant — that is the vacuity the family names — while its paired positive control fails because a recorded leader start does not move the legacy probe. The canonical reader (`LoadPure`) is what the real helper uses; the fixture also asserts no legacy `backlog.json` exists, so the two counters are separated by construction.
- GREEN pair: reverted -> exit 0, all four subtests `--- PASS` (`ok ... 5.387s`).

**Placement (AC-LA-006 / AP-3), mutant: the cc leader guard moved behind the session mutations.**
- Command: `go test ./internal/cli/ -run '^TestLauncherAnchorGuardSemantics$/refusal_before_mutation$' -v -count=1`. Tree `87abe2ef7` + mutant. Exit code: **1**. Verdict lines (the `cut -c1-260` filter truncated the first marker map):

```
factory_anchor_guard_cells_test.go:173: cc: factory markers stamped before the guard ran: map[MOAI_FACTORY_WORKERS:1 MOAI_KANBAN_BACKEND:claude MOAI_KANBAN_ID:tn274w MOAI_KANBAN_LEAD_ADDR:/tmp/moai-socket-factory/tn274w MOAI_KANBAN_LEAD_NAME:leader MOAI_KA
factory_anchor_guard_cells_test.go:173: cc: run records before=0 at guard=1 after=1, want all equal
factory_anchor_guard_cells_test.go:173: cc: 1 settings artifacts existed when the guard ran, want 0
--- FAIL: TestLauncherAnchorGuardSemantics/refusal_before_mutation/cc/named_tree (2.25s)
--- FAIL: TestLauncherAnchorGuardSemantics/refusal_before_mutation/cc/launch_directory (1.70s)
```

- Why red: all three indicators of AC-LA-006 trip (role/run markers in the environment, a run record, a settings artifact) when the guard stands behind the mutations. GREEN pair: reverted -> exit 0, all four subtests PASS.

**Weakening mutants (AP-6) and layer mutants — detection power of the preserved semantics** (tree `87abe2ef7` + mutant, each reverted):

| Mutant | Command | Exit | Verdict |
|---|---|---|---|
| E1 unrecordable bypass proceeds (`err != nil && false`) | `-run '^TestLauncherAnchorGuardSemantics$/override_audit_row$'` | 1 | `a bypass that cannot be recorded proceeded: err=<nil> launched=true` (cc and glm) |
| E2 descriptor file-type re-read removed | `-run '^TestLauncherAnchorAuditHygiene$'` | 1 | `descriptor_recheck_refuses_a_swapped_fifo`: `a FIFO swapped in after the Lstat check yielded handle=&{...} err=<nil>, want a refusal and no handle` |
| E3 Lstat regular-file refusal removed | `-run '^TestLauncherAnchorAuditHygiene$'` | **0** | **SURVIVED** — see Gaps: on Unix `O_NOFOLLOW|O_NONBLOCK` plus the descriptor re-read refuse the same shapes, so the Lstat layer is only observable where `auditLogOpenNoFollow` is 0 (Windows) |
| G existing tree judged by the `--branch` value | `-run '^TestLauncherAnchorBranchAnnotation$'` | 1 | `card tree with a plain --branch: the anchored leader launch was not refused` and `a plain tree with a card --branch value is reused unchanged and must launch: err=factory leader launch refused` |
| F selection takes the FIRST worktree option | `-run '^TestLauncherAnchorFinalArgv$/multi_option_last_wins$\|^TestLaunchViewBackendSelection$'` | 1 | `the anchored leader launch was not refused` and `control: a card tree FIRST and a plain tree LAST must launch (the backend takes the last)` |

### M3 results (tree `f4676475b`)

- Mutants against the scan (tree `87abe2ef7` + the then-uncommitted scan test, content identical to `f4676475b`): S1 a `worktreeFlagValue(nil)` call in `launch_chain.go`; S2 an `a == "-w"` match in `lane_resume.go`. Command `go test ./internal/cli/ -run '^TestLaunchViewSingleParse$/no_rederivation_on_launch_path$' -v -count=1`, exit **1**: `worktree selection is derived outside launch_view.go (2 findings): [lane_resume.go: matches tokens against the worktree spelling literal "-w" launch_chain.go: reference to worktreeFlagValue]`. Both reverted. The scan also found, on its first run, a false positive it was narrowed against: `glm.go` carries `"-w"`/`"--worktree"` as keys of the usage-redaction allowlist `glmFlagNames` (flag names, not a reader); the literal sweep now counts only token-matching contexts (`==`/`!=`, `case`, `strings.HasPrefix/TrimPrefix/CutPrefix`), with a control half proving `launch_view.go` itself matches every literal.

### AC matrix M1-M3 at tree `f4676475b` (E1; exact commands from acceptance.md, run with `-v -count=1`, one invocation each; `[no tests to run]` count 0 in every run)

| AC | Command (`go test ./internal/cli/ -run <selector>`) | Exit | Verbatim PASS lines (excerpt) |
|---|---|---|---|
| AC-LA-001 | `'^TestLauncherAnchorFinalArgv$/consumed_option_shadowing$'` | 0 | `--- PASS: TestLauncherAnchorFinalArgv/consumed_option_shadowing (1.40s)` (+ `/cc`, `/glm`) |
| AC-LA-002 | `'^TestLauncherAnchorFinalArgv$/multi_option_last_wins$'` | 0 | `--- PASS: TestLauncherAnchorFinalArgv/multi_option_last_wins (4.30s)` (+ `/cc`, `/glm`) |
| AC-LA-003 | `'^TestLauncherAnchorFinalArgv$/bundled_short_cluster$'` | 0 | `--- PASS: TestLauncherAnchorFinalArgv/bundled_short_cluster (4.04s)` (+ `/cc`, `/glm`) |
| AC-LA-004 | `'^TestLauncherAnchorRunCountInstrument$'` | 0 | `--- PASS: TestLauncherAnchorRunCountInstrument (2.37s)`, `.../cc/positive_control_counter_moves`, `.../glm/positive_control_counter_moves`, + the two `refusal_leaves_the_counter_unmoved` (5 PASS lines >= 2) |
| AC-LA-005 | `'^TestLauncherAnchorFinalArgv$/view_exec_byte_equality$'` | 0 | `--- PASS: TestLauncherAnchorFinalArgv/view_exec_byte_equality (12.98s)`, 30 PASS lines (14 rows x cc/glm + parents) |
| AC-LA-006 | `'^TestLauncherAnchorGuardSemantics$/refusal_before_mutation$'` | 0 | `--- PASS: .../refusal_before_mutation (1.86s)`, 6 PASS lines |
| AC-LA-007 | `'^TestLauncherAnchorGuardSemantics$/lane_and_nonfactory_exempt$'` | 0 | `--- PASS: .../lane_and_nonfactory_exempt (3.50s)`, 6 PASS lines |
| AC-LA-008 | `'^TestLauncherAnchorGuardSemantics$/override_audit_row$'` | 0 | `--- PASS: .../override_audit_row (4.28s)`, 8 PASS lines |
| AC-LA-009 | `'^TestLauncherAnchorAuditHygiene$'` | 0 | `--- PASS: TestLauncherAnchorAuditHygiene (0.01s)` + `symlink_at_the_audit_log_is_refused`, `symlink_for_the_audit_directory_is_refused`, `symlink_for_the_moai_directory_is_refused`, `regular_file_still_receives_its_row`, `fifo_refused_without_blocking`, `descriptor_recheck_refuses_a_swapped_fifo` (6 subtests >= 3) |
| AC-LA-010 | `'^TestLauncherAnchorGuardSemantics$/fail_open_on_unverifiable$'` | 0 | `--- PASS: .../fail_open_on_unverifiable (2.57s)`, 6 PASS lines |
| AC-LA-011 | `'^TestLauncherAnchorGuardSemantics$/discriminator_ignores_env$'` | 0 | `--- PASS: .../discriminator_ignores_env (2.45s)`, 4 PASS lines |
| AC-LA-012 | `'^TestLauncherAnchorBranchAnnotation$'` | 0 | `--- PASS: TestLauncherAnchorBranchAnnotation (4.89s)`, `.../cc/existing_tree_real_branch`, `.../cc/created_tree_annotation`, `.../glm/existing_tree_real_branch`, `.../glm/created_tree_annotation` (>= 2) |
| AC-LA-013 | `'^TestLaunchViewSingleParse$/no_rederivation_on_launch_path$'` | 0 | `--- PASS: TestLaunchViewSingleParse/no_rederivation_on_launch_path (0.01s)` |
| AC-LA-014 | `'Worktree\|TestLaneJoinChildArgv'` | **1** | 179 `--- PASS`, 1 failing test: `TestFactoryLeaseDriftLogVerbWorktreeWriteWaits/nominated` (`returned 633.191541ms after the release ... want ... within 500ms after it`). Present: `--- PASS: TestLaneJoinChildArgv (0.00s)`, `--- PASS: TestLauncherWorktreeDuplicateOptionFirstValue (0.09s)`, `--- PASS: TestLauncherWorktreeStageCharacterization (2.07s)`. Classification: timing-window flake under host load, not caused by this change — the same test re-run alone three times at the same tree: `ok ... 9.658s`, `ok ... 8.852s`, `ok ... 9.152s`; load average at the time 30.88 32.01 41.18. The same selector passed whole (exit 0, 180 PASS, 215.399s) at M1 `a022ada95`. |
| AC-LA-015 | this section | - | RED records above carry command, verbatim stdout, exit code, tree SHA and green-flip label (M2) for families (a)(b)(c) [one mutant, three cells], (d), and the placement; each paired with a GREEN observation. |

Further regression runs at `f4676475b` (all `-count=1`, scoped by test name): the 183 tests of every test file that drives a leader entry (`"-f"` / `"--factory"`) -> `ok ... 214.542s` exit 0; the 334 further tests of files that drive `runCC` / `runGLM` / `runClaudeEntry` -> `ok ... 83.006s` exit 0; `-run 'TestLeader|TestLauncher'` -> `ok ... 82.441s`; every test in the `lane_resume*_test.go` files after the snapshot change -> `ok ... 2.810s`.

### E2 / E4 / E5 / E6 (tree `f4676475b`)

- E2: `go build ./...` -> exit 0, no output. `GOOS=windows GOARCH=amd64 go build ./...` -> exit 0, no output. `GOOS=windows GOARCH=amd64 go vet ./internal/cli/` -> exit 0.
- E4: `grep -n 'AskUserQuestion' <launch_view.go factory_anchor_guard.go factory_audit_nofollow_{unix,other}.go cc.go glm.go launcher.go launch_chain.go session_worktree.go worktree_branch_flag.go codex_hooks_seed.go lane_resume_model.go internal/config/defaults.go>` -> no output, exit 1 (no hits). `grep -c syscall` on the four new non-test files -> 0 each (the Unix flags come from `golang.org/x/sys/unix` behind `//go:build unix`, the repo's existing pattern; `//go:build !unix` carries a zero constant).
- E5: `gofmt -l internal/cli internal/config` -> no output; `go vet ./internal/cli/ ./internal/config/` -> exit 0. New versus baseline: no vet or gofmt finding exists before or after. Whole-tree and package `golangci-lint` were not run (excluded from this delegation).
- E6: new commits `538af8d7c`, `a022ada95`, `87abe2ef7`, `f4676475b` (listed above); nothing pushed; `git status --short` after `f4676475b`: only ` M .moai/harness/usage-log.jsonl.prune-state` (runtime-managed, never staged).

### Gaps (not observed) and refusals

- **Package coverage (85%) and whole-package lint are not measured** — M4, excluded from this delegation. No coverage figure is claimed.
- **`/verify` (the MoAI gate) was not run before the commits**: it runs lint + type + the full test suite, which this delegation forbids for `./internal/cli`. Substituted, and named as a substitution: `gofmt -l`, `go vet` (native + windows), both `go build ./...`, and the scoped test families listed above.
- **Backend semantics not measured on a running claude**: last-wins for a repeated `--worktree`, the empty `--worktree=` value, the cluster decomposition, and the attached `-w=value` read as `value` (the launcher's own equals handling) rest on the help declaration plus commander's documented parsing and the audited iter8 record. A running `claude -w` would create a worktree.
- **Mutant E3 survived** (Lstat refusal removed, `AuditHygiene` still PASS): on Unix the open itself refuses a link (`O_NOFOLLOW`), a FIFO (`O_NONBLOCK` write open with no reader fails, plus the descriptor re-read) and a directory, so the Lstat is a redundant layer there; it is the only guard where `auditLogOpenNoFollow` is 0, and no Windows run was available (cross-build and `go vet` only).
- **The `appendCrossSessionSettings` step between the entry and `runLaunchClaude` is not exercised by the exec-seam cells** (`unifiedLaunchFunc` is substituted to call `runLaunchClaude` directly, as the existing `launchModel` harness does for the later stage). It appends `--settings` / `--effort` at the end of the argv, which `appended_flags_are_worktree_neutral` covers as a class, not as the live function.
- **Guard call sites are cc and glm only**, per plan M2. The codex launcher keeps its own L2 handling (`resolveWorktreeL2Path` wrapper) and is out of the SPEC.
- **Refused tool calls**: none of this run's commands was refused by the worktree-isolation guard. One shell invocation failed on `ugrep` not being a command; it was re-run as `grep`.
- **Moving coordinates**: every line number in this file is as observed at the SHA named beside it; none is a pin.

### Residual risk

- The guard judges a view computed from the entry argv; the exec argv additionally carries tail flags appended afterward. Neutrality of `--name` / `--settings` / `--effort` is pinned for the current launcher; a future launcher change that inserts such a token BEFORE a separator (as `laneJoinChildArgv` can for a name-less lane) would break it — the lane is exempt from the guard today.
- `selectBackendWorktree` depends on `activeClaudeOptionModel` (derived from the installed claude's help when available). A claude release that makes `-w` a required-value option or adds an option that swallows the next token would shift selections; the snapshot fallback and the 2.1.302 re-measure are the mitigation, and `TestLaunchViewBackendSelection` pins the snapshot behaviour.
- Existing tests outside the two scanned families that start a leader entry from the real working directory would now see the guard's refusal when the suite runs inside a card worktree; all such tests found by the two name-scoped sweeps were adapted, but the sweeps are by file content, not by coverage.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §G Override and Refusal Record

- 2026-10-17T14:48:03Z SPEC-LAUNCHER-ARGV-001 ceiling-refusal outcome=hold reasons="plan-audit ceiling reached (round count 4 >= tier ceiling 2); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1748/.moai/reports/t1748/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1748/.moai/reports/t1748/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1748/.moai/reports/t1748/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1748/.moai/reports/t1748/plan-audit.md
- 2026-10-17T14:50:06Z SPEC-LAUNCHER-ARGV-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 4 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1748/.moai/reports/t1748/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1748/.moai/reports/t1748/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1748/.moai/reports/t1748/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1748/.moai/reports/t1748/plan-audit.md
- 2026-10-17T14:50:23Z SPEC-LAUNCHER-ARGV-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 4 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1748/.moai/reports/t1748/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1748/.moai/reports/t1748/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1748/.moai/reports/t1748/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1748/.moai/reports/t1748/plan-audit.md
- 2026-10-17T14:52:04Z SPEC-LAUNCHER-ARGV-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 4 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1748/.moai/reports/t1748/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1748/.moai/reports/t1748/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1748/.moai/reports/t1748/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1748/.moai/reports/t1748/plan-audit.md
