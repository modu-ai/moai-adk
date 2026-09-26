# SPEC-CODEX-FACTORY-RETIRE-001 — research

Tree: `553e224f3` (branch `WT-codex-factory-retire`, base develop `553e224f3`).
Every coordinate here was read in this run on this tree. The design's coordinates came
from develop `35ab8cff3`; where they diverge, §R2 says so.

## R1 Verified coordinates

| Item | Coordinate (this tree) | Note |
|---|---|---|
| `moai codex -k` entry | `internal/cli/codex_kanban.go` (122 lines) | `stripCodexKanbanFlag`, `applyCodexKanbanEntry`, `codexKanbanEntry`, `codexKanbanUsageDiag` |
| `moai codex -f` entry | `internal/cli/codex_factory.go` (147 lines) | `stripCodexFactoryFlag`, `applyCodexFactoryEntry`, `codexFactoryBackend = "codex"`, `isFactoryLaneShape`, `isFactoryRoleToken`, `codexHeadTokenIsVerb` |
| `runCodex` | `codex_launcher.go:689-781` | `--factory-run` strip `:701-707`; `-f` strip `:711`; `-k` strip `:717-724`; factory apply + run select + `recordFactoryRunStart` `:725-745`; kanban apply `:762-764`; readout usage guard `:767` |
| spawn factory branch | `codex_launcher.go:218-273` (`defaultCodexSpawnLaunch`) | `factoryLaunchEnabled` `:228`; pane identity probe `:231`; **shared with the `-w` anchor** (`codexSpawnAnchorFn`, `:229-245`); launch-pending `:246-248`; `stampFactoryRunOwner` `:249-256`; `clearFactoryRunOwner` `:241`, `:266` |
| spawn forwarded env | `codex_launcher.go:296-310` (`codexSpawnForwardedEnv`) | 9 lane keys (`:298-308`) plus `MOAI_HOME`, `CLAUDE_PROJECT_DIR`; `buildCodexSpawnCommand` `:320-340` already blanks `CLAUDE_CODE_SESSION_ID=` and `MOAI_SESSION_PID=` at `:329` — the pattern REQ-CFR-007 extends |
| direct child env | `codex_launcher.go:613-627` (`codexChildEnv`) | drops only `CLAUDE_CODE_SESSION_ID`, `MOAI_SESSION_PID` |
| POSIX direct | `codex_direct_posix.go:39` launch-pending, `:44` `syscall.Exec`, `:45` rollback | keep `syscall.Exec` + `codexDirectAnchorPID` |
| Windows direct | `codex_direct_windows.go:23-50` | Start → probe → `clearFactoryRunOwner` `:33,:39` → launch-pending `:36` → `stampFactoryRunOwner` `:45` → Wait. All factory-only; `-w` anchor on Windows uses `codexDirectAnchorPID` before launch and does not depend on it |
| launch-pending helper | `factory_launch_pending.go` | **shared**: `launch_exec_posix.go:33,38`, `launch_exec_windows.go:58` (cc/glm) — keep |
| run owner helpers | `factory_run_owner.go` | **shared**: `launch_exec_windows.go`, `factory.go` — keep |
| run selection | `factory.go:245` `enterSelectedFactoryRun`, `:258` `recordFactoryRunStart`, `:278` `stripFactoryRunFlag` | shared with `cc.go:188,193,206`, `glm.go:234,242,255`; `stripFactoryRunFlag` has no other production caller than `runCodex` and `factory.go` itself — verify at run time before touching |
| run lead backend | `homestate` `runs.lead_backend` (`factory.go` schema `:36`, write `runtime.go:37-39`) | the column REQ-CFR-010 reads |
| existing refusal sentinels | `factory.go:48` `FACTORY_MODE_UNSUPPORTED_BACKEND`, `kanban.go:45` `KANBAN_MODE_UNSUPPORTED_BACKEND` | reused (D2) |
| handoff CLI | `factory_lane_handoff.go` (274), `_switch.go` (193), `_bind.go` (62), `_recover.go` (218) | production callers: none (§R3) |
| codex relocation | `mcp_codex.go:824-906` (`codexThreadRelocation`, `runCodexThreadRelocation`); consts `codexMethodThreadFork` `:77`, `codexNotifyThreadStarted` `:81` | `awaitCodexResponseObserving` `:1116` is also used by `awaitCodexResponse` `:1113` — keep |
| relocation timeout | `internal/config/defaults.go:498` `DefaultCodexHandoffRelocationTimeout` | only consumer is `_switch.go` |
| hook bind | `internal/hook/factory_handoff_bind.go`; called from `factory_messages.go:95,100,112` | backend-neutral, Claude per-prompt path — keep (D3) |
| handoff tables | `factorymsg/handoff.go:65-69` (4 tables), `handoff_bind.go:28-30` (3 tables), `dispatch.go:44` | schema unchanged (REQ-CFR-015) |
| MCP env allowlist | `internal/codexwiring/configtoml.go:21` (`mcpServerEnvVarsValue`), canonical check `:173`; generated `.codex/config.toml:4` | unchanged (D1) |
| env key constants | `internal/config/envkeys.go:182-287` | the eleven lane keys of REQ-CFR-006 |

## R2 Divergences from the design

1. **`factoryQueueCodexMessage` / the `to.Backend == "codex"` branch are absent** from
   the committed tree. Evidence, pinned to `553e224f3` and scoped to code so this
   SPEC's own body (which names the symbol) is not swept:
   `git grep -c factoryQueueCodexMessage 553e224f3 -- internal cmd pkg` → exit 1;
   `git grep -c 'to.Backend == "codex"' 553e224f3 -- internal` → exit 1;
   `git grep -c 'queue --thread' 553e224f3 -- internal` → exit 1. (An unpinned
   `git grep … HEAD -- .` on a later HEAD matches the SPEC artifacts themselves and
   is not evidence of presence in code — plan-audit iter-1 D9.)
   `internal/cli/mcp_factory_push_live_test.go` does not exist here;
   `internal/cli/mcp_factory_msg.go` here is 147 lines with no codex branch. The
   symbol lives only in the develop worktree's uncommitted copy (lead report). The
   REMOVE item therefore reduces to the foreign-file disposition (plan.md §M5).
2. **"§3 Codex lanes"** is actually the §0 capability-bindings row at `AGENTS.md:26`
   ("`Codex lanes: moai codex -w <worktree>`"). §3 itself carries no codex wording.
3. **`internal/template/templates/AGENTS.md` does not exist.** The mirror is
   `internal/template/templates/AGENTS.md.tmpl`; its row at `:32` carries "Codex lanes"
   and its §8 (`:272-277`) already has no `-f lead/agents` phrase. Root `AGENTS.md` §8
   (`:265-270`) does. The two copies diverge by design; both are edited only where the
   retired wording appears.
4. **`manager-lead` Role B has no codex-factory wording** in C1
   (`.claude/agents/moai/manager-lead.md`), C2
   (`internal/template/templates/.claude/agents/moai/manager-lead.md`) or C3
   (`internal/template/templates/.codex/agents/moai/manager-lead.toml`; there is no
   root `.codex/agents/` directory). The only "Codex" line (C2 `:46`, C3 `:39`) is the
   harness note on `subagent-spawn`, which stays. No agent edit, no `make agents-emit`.
5. **Handoff CLI has no production caller** (§R3), so removing it withdraws no command.
6. **Docs coordinates:** `docs-site/content/{ko,en,ja,zh}/advanced/codex-dual-harness.md:33`
   carries the `-f` lead/agents launch shape; `docs-site/content/{ko,en,ja,zh}/guides/mcp-server.md:180-184`
   carries "Codex lane orchestrator" / "Codex 레인 오케스트레이터" / "Codex レーンのオーケストレーター"
   / "Codex 泳道编排器" and a follow-on sentence about lanes.
7. **Rule pair is byte-identical today**: `cmp` of local vs template for
   `moai-mcp-tools.md` and `moai-mcp-tools-catalogue.md` → exit 0 for both.
   "Codex lane orchestrator" appears at `moai-mcp-tools.md:75` and
   `moai-mcp-tools-catalogue.md:103-105`.

## R3 Caller enumeration (symbols being removed)

Command form: `grep -rln "\b<sym>\b" internal cmd pkg --include='*.go'`, production
files only unless noted.

| Symbol | Production files | Shared with Claude paths? |
|---|---|---|
| `stripCodexKanbanFlag`, `applyCodexKanbanEntry`, `codexKanbanUsageDiag` | `codex_kanban.go`, `codex_launcher.go` | no |
| `stripCodexFactoryFlag`, `applyCodexFactoryEntry` | `codex_factory.go`, `codex_launcher.go` (tests: `codex_factory_test.go`, `factory_worker_naming_test.go:115-137`) | no |
| `codexFactoryBackend` | `codex_factory.go`, `codex_kanban.go`, `codex_launcher.go:739` | no — but REQ-CFR-010 needs a `"codex"` literal; `mcp_convergence.go:64` already defines `BackendCodex = "codex"` |
| `exportKanbanLaunchFacts`, `exportFactoryLaunchFacts`, `enterKanbanMode`, `resolveCompanionName`, `resolveFactoryWorkerName` | also `cc.go`, `glm.go`, `kanban.go` | **yes — keep** |
| `registerFactoryLaunchPending` / `rollbackFactoryLaunchPending` | also `launch_exec_posix.go`, `launch_exec_windows.go` | **yes — keep the helper, drop the codex call sites** |
| `stampFactoryRunOwner` / `clearFactoryRunOwner` | also `launch_exec_windows.go`, `factory.go` | **yes — keep** |
| `enterSelectedFactoryRun`, `recordFactoryRunStart` | also `cc.go`, `glm.go` | **yes — keep** |
| `prepareLaneHandoff` | `factory_lane_handoff.go` only; call sites: 5 test files (25 calls) | no |
| `switchLaneHandoffInteractive` / `Headless` | `_switch.go` only; call sites: tests | no |
| `bindLaneHandoffHeadless` | `_bind.go`, `_recover.go`; call sites: tests | no |
| `recoverLaneHandoff` | `_recover.go`; call sites: `factory_lane_handoff_recover_test.go` | no |
| `codexThreadRelocation`, `runCodexThreadRelocation` | `mcp_codex.go`, `factory_lane_handoff*.go` | no |
| `DefaultCodexHandoffRelocationTimeout` | `config/defaults.go`, `_switch.go` | no |

## R4 Test impact

- **Delete with their subject:** `codex_kanban_test.go` (`TestCodexKanbanEntryParity`),
  `codex_factory_test.go` (`TestApplyCodexFactoryEntryExportsBackendAndRestores`,
  `TestStripCodexFactoryFlag`), `factory_lane_handoff_{,create_,switch_,bind_,recover_,nowrite_,live_gate_,compat_}test.go`
  (8 files), `factory_worker_naming_test.go` `TestStripCodexFactoryFlagWorkerVocabulary`,
  `codex_launcher_test.go` `TestFactoryCodexSpawnRegistersLaunchPendingPeer` (`:914`),
  `codex_launcher_exec_posix_test.go` factory-owner cases (`:21`, `:143`, `:159` —
  re-read; the exec-replacement property itself must stay covered).
- **Move (shared code):** `TestNextFactoryWorkerNumber` (`codex_factory_test.go:82`)
  and its helper `NextFactoryWorkerNumberForTest` (`codex_factory_helper_test.go`) test
  `kanban.NextFactoryWorkerNumber` — move to a kept file (e.g. beside
  `factory_worker_naming_test.go`). `internal/kanban/factory_worker_label_test.go:71`
  covers a neighbouring case but not this one.
- **Rewrite fixture, keep the test:** `factory_handoff_abandon_test.go` (abandon-lane,
  a kept command) uses `laneHandoffFixture` / `newLaneHandoffFixture` / `f.wtReady`
  defined in `factory_lane_handoff_test.go:29,60`. The fixture must move into a kept
  file and its states must be built through the `factorymsg` Store API
  (`ReserveHandoff`, `MarkHandoffWTReady`, `MarkHandoffSwitchPending*`) instead of
  `prepareLaneHandoff`. This is the largest hidden dependency of M3.
- **Switch fixture (crosses REQ-CFR-010):** `factory_mixed_test.go:44,97,135`
  (`TestFactoryRunSelectionAtomicSlotsAndArgv` records `Backend: "codex"` and then
  joins), `factory_operational_fixture_test.go:103,113,160`, `factory_run_owner_test.go:187-256`
  (drives `defaultCodexSpawnLaunch` with a codex backend).
- **LIVE (env-gated) tests:** `factory_live_test.go:26-32` (cases `codex-codex`,
  `codex-claude`, `claude-codex`), `factory_operational_live_test.go:80,245`
  (`moai codex -f`, `-f worker`) — remove the codex cases.
- **Local-instructions shapes:** `codex_local_instructions_test.go:436` enumerates `-f`,
  `-f agent`, `-f agent-2`, `-f lane-3` among launch forms — drop those forms (they
  now refuse) and keep the others.
- **Opaque `"codex"` fixture values** in `internal/factorymsg/*_test.go`,
  `internal/hook/factory_*_test.go`, `internal/web/*` — keep (D4); the Store does not
  refuse by backend.
- **Existing t1222 pattern:** `factory_test.go:23-48` `clearFactoryTestEnv` +
  `requireFactoryLaunchDisabled`. The helper clears 6 of the eleven lane keys plus
  `CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS`. REQ-CFR-006 covers all eleven. The t1222
  lesson applies to the new env test: the factory gate is conjunctive (`KANBAN_ID`
  AND a worker key), so a single-key mutant does not flip it — assert the eleven keys
  individually, not only the gate.

- **Abandon-lane fixture depth (plan-audit iter-1 D5):** `TestFactoryLaneHandoffOperatorAbandon`
  (`factory_handoff_abandon_test.go:93`, 4 subtests) snapshots the real target
  worktree and branch through `targetSnapshot` (`:75-87`) and asserts they are
  preserved. The worktree and branch are created inside `wtReady`
  (`factory_lane_handoff_switch_test.go:133-140`) by `prepareLaneHandoff`. It also uses
  `f.storedHandoff`, `f.brokerDB`, `count`, `endpointRow`, `target` and `handoffGit`.
  A Store-API-only rebuild would create no target tree, making the preservation
  assertion vacuous. The rebuilt fixture must materialize the target worktree and
  branch itself (git, via the existing `handoffGit` helper) before marking the
  handoff `WT_READY`.
- **Pinned test names for kept surfaces (plan-audit iter-1 D4/D16):**
  `TestWorktreeLaunchRejectsConcurrentWriter`, `TestCodexWorktreeAnchorLockAndBase`,
  `TestCodexSpawnAnchorsToPanePID` (`codex_worktree_anchor_test.go`);
  `TestCodexLaunchVerb_StatusStaysTheReadout` (`codex_launch_verb_test.go`);
  `TestCodexLocalInstructions_DirectSpawnAndAppSharePrefix`
  (`codex_local_instructions_test.go`); `TestCodexTask_ForegroundReturnsOutput`
  (`codex_task_test.go`); `TestSessionMsgSendPollAckHandlers` (`mcp_session_msg_test.go`);
  `TestCodexRoleLoadNegativeControl` (`codex_role_load_control_test.go`). Each was
  located with `grep -rl "func <name>(" internal/cli` on this tree.
- **Codex hook path (plan-audit iter-1 D7):** `registerFactoryHookPeer`
  (`internal/hook/factory_messages.go:53-116`) reads `MOAI_KANBAN_ID` and
  `MOAI_FACTORY_WORKER(S)` only; `HookInput` carries no harness field. The harness is
  known only at the CLI boundary (`harnessModeIsCodex`, `internal/cli/hook_harness_codex.go:27-39`).
  Whether a directly started `codex` in a Claude lane actually overwrites the lane's
  slot was not run (code-path reading only).

## R5 Budget and generated-artifact measurements

- `go test ./internal/config -run 'TestAlwaysLoadedTokenBudget$' -count=1 -v` on this
  tree → `always-loaded surface = 77539 tokens (budget 77600, headroom 61, 16 entries)`.
- AC-count snapshot (`.moai/reports/t338/ac-count-baseline.txt`): adding a new
  `acceptance.md` is not a cascade trigger (`.moai/docs/ac-count-baseline-refresh.md`
  §2, last paragraph). No snapshot regeneration is part of this plan commit. Deleting
  test files in run does not touch any `acceptance.md`.

## R6 Foreign files in the develop worktree (lead-provided, not observed by this agent)

The lead reported (develop worktree `.claude/worktrees/develop`, HEAD `553e224f3`, last
written 09-24, no later author):

| Path | State | Size |
|---|---|---|
| `internal/cli/mcp_factory_msg.go` | modified | +62 (`factoryQueueCodexMessage`) |
| `internal/cli/mcp_factory_msg_test.go` | modified | +94 |
| `internal/factorymsg/store.go` | modified | +4 |
| `internal/cli/mcp_factory_push_live_test.go` | untracked | `TestFactoryMsgCodexWakeLive` |
| `reports/factory-cross-host-push-20260924.md` | untracked | report |
| `reports/factory-cross-host-push-20260924.html` | untracked | report |

This agent did not read that worktree (instruction: never touch it). The list is the
lead's observation; this tree's committed copies of the three modified files are the
base the revert returns to.
