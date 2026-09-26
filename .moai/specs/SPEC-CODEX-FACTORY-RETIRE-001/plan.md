# SPEC-CODEX-FACTORY-RETIRE-001 — plan

Tier L · card t1242 · class C · methodology per `quality.yaml` (TDD for the new refusal
and env-defense behavior; deletion milestones are behavior-preserving for the kept
surfaces). Milestones are ordered by decision reversibility: the user-facing refusal
and the environment contract come first, mechanical deletion last. Every milestone
leaves `go build ./...` green on darwin, linux, and windows.

## §A Context

See `spec.md` §A and `research.md` §R1-R4. Base tree `553e224f3`.

## §B Known issues carried in

- The spawn pane-identity probe in `defaultCodexSpawnLaunch` serves both the factory
  branch and the `-w` anchor. Removing the factory branch must keep the probe, the
  anchor call, and the pane cleanup on anchor failure.
- `factory_handoff_abandon_test.go` depends on `laneHandoffFixture` from a file M3
  deletes (research §R4).
- internal/hook tests must run under an isolated `MOAI_HOME` (card t1229 hazard:
  SessionStart tests can overwrite the lane's row in the real lease DB).

## §C Pre-flight (run phase)

The run lane records its measurements in `progress.md` §E.2 (the run lane may write
that section; it may not edit this plan body):

1. Absorb local `develop`; record `git rev-parse --short HEAD` as `run_base:` in
   `progress.md` §E.2. Every `<base>` in `acceptance.md` means this value.
2. Re-run `go test ./internal/config -run 'TestAlwaysLoadedTokenBudget$' -count=1 -v`
   on the run base and record the surface figure as `budget_base:` in §E.2 (context
   only; AC-CFR-022's verdict compares against the merge commit's develop parent,
   re-measured in the same run).
3. Re-run the R3 caller greps; a new production caller of a symbol slated for removal
   stops the milestone that removes it.

## §D Constraints

- No `factory.db` schema change (REQ-CFR-015).
- No change to `internal/codexwiring/configtoml.go` `mcpServerEnvVarsValue` (D1).
- No hand edit of `internal/template/templates/.codex/agents/moai/*.toml`; no agent
  edit is planned (research §R2.4).
- Rule edits land in both the local copy and the template mirror in the same commit.
- docs-site edits: ko first, then en/ja/zh in the same commit; Mermaid none added.
- Verification scope is the touched packages; the full suite is CI's (AGENTS.md §4).

## §E Self-verification per milestone

Each milestone closes with the ACs listed under it, the package tests of every
touched package, and `GOOS=windows go build ./...` plus `GOOS=linux go build ./...`.

## §F Milestones

### M1 — entry refusal, lane-env defense, codex-run join refusal (Priority High)

Highest change likelihood: this is the only operator-visible behavior the SPEC adds.

1. `runCodex`: replace the `-f`/`--factory-run`/`-k` stripping with one pre-verb scan
   of the head (before `--`) that refuses with the D2 sentinels, rc 1, before any env
   or state effect (REQ-CFR-001..005). The refusal message text is a named constant;
   one constant per mode.
2. `codexChildEnv`: drop the eleven lane keys (REQ-CFR-006). Put the eleven keys in one
   named list so the direct and spawn paths cannot drift; the test asserts each key
   separately (t1222 lesson, research §R4).
3. `buildCodexSpawnCommand` / `codexSpawnForwardedEnv`: remove the nine lane keys from
   the forwarded list and emit `KEY=` for all eleven, extending the existing
   `CLAUDE_CODE_SESSION_ID=` pattern (REQ-CFR-007). `MOAI_HOME` and
   `CLAUDE_PROJECT_DIR` stay forwarded.
4. Codex launch files: remove the launch-pending registration, rollback, run-owner
   stamp/clear from `codex_direct_posix.go`, `codex_direct_windows.go`, and
   `defaultCodexSpawnLaunch` (REQ-CFR-008), keeping `syscall.Exec`,
   `codexDirectAnchorPID`, and the `-w` anchor path.
5. `enterSelectedFactoryRun` (or the cc/glm call sites): after the run is resolved,
   refuse when its `runs.lead_backend` is `codex` (REQ-CFR-010/011). Use the existing
   `"codex"` constant rather than a new literal where one is reachable. Read the
   column from `internal/cli` through the already-exported `homestate.OpenFactory(...).DB`
   handle — **no production change to `internal/homestate` or `internal/factorymsg`**,
   which AC-CFR-017 guards byte-for-byte.
6. Codex hook peer refusal (REQ-CFR-022): at the CLI boundary where `--harness codex`
   is already parsed (`harnessModeIsCodex`), keep the factory peer registration from
   seeing a lane identity (e.g. run the handler with the eleven lane keys unset for
   that process). The Claude path (no flag) is unchanged.

ACs: AC-CFR-001..010, AC-CFR-025.

### M2 — dead entry code and test moves (Priority Medium)

1. Move `TestNextFactoryWorkerNumber` + `NextFactoryWorkerNumberForTest` into a kept test
   file **before** deleting `codex_factory_test.go` / `codex_factory_helper_test.go`
   (REQ-CFR-017).
2. Delete `codex_kanban.go`, `codex_factory.go`, `codex_kanban_test.go`,
   `codex_factory_test.go`, and the helper file once emptied.
3. Remove codex-path tests listed in research §R4 ("Delete with their subject"); keep
   one test that proves the POSIX exec replacement property without a factory run.
4. Switch the fixtures listed in research §R4 ("Switch fixture") to `claude`/`glm`;
   drop the codex cases of the env-gated LIVE tests; drop the `-f` forms from
   `codex_local_instructions_test.go:436`.
5. Remove `stripFactoryRunFlag` only if the R3 re-grep shows no remaining caller.

ACs: AC-CFR-011..015.

### M3 — lane handoff CLI and codex relocation (Priority Medium)

1. Move `laneHandoffFixture` and every helper `factory_handoff_abandon_test.go` uses
   (`wtReady`, `storedHandoff`, `brokerDB`, `count`, `endpointRow`, `target`,
   `handoffGit`, `targetSnapshot`) into a kept test file. Rebuild `wtReady` so it
   **still materializes a real target worktree and branch** (git through `handoffGit`)
   and then drives the handoff to `WT_READY` through the `factorymsg` Store API
   (`ReserveHandoff`, `MarkHandoffWTReady`, `MarkHandoffSwitchPending*`). The fixture
   asserts, before any abandon call, that the target path exists and the target
   branch ref resolves. Run the abandon-lane tests green **before** deleting anything,
   and run the AC-CFR-020 mutant once (fixture without the worktree step → test FAILs).
2. Delete `factory_lane_handoff.go`, `_switch.go`, `_bind.go`, `_recover.go` and their
   eight test files.
3. Delete `codexThreadRelocation`, `runCodexThreadRelocation` from `mcp_codex.go`, the
   now-unused `codexMethodThreadFork` / `codexNotifyThreadStarted` constants (confirm no
   other reader first), and `config.DefaultCodexHandoffRelocationTimeout`.
4. Keep `internal/hook/factory_handoff_bind.go`, the `factorymsg` handoff API, and every
   `CREATE TABLE` (D3, REQ-CFR-014/015).

ACs: AC-CFR-016..018, AC-CFR-020.

### M4 — docs, rules, AGENTS.md, token budget (Priority Low)

1. `AGENTS.md:26` and `AGENTS.md.tmpl:32`: "Codex lanes: `moai codex -w`" → "Codex:
   `moai codex -w`". `AGENTS.md` §8 (`:267`): drop "and `-f` lead/agents" from the launch
   shapes. Shrink, never grow (headroom 61 tokens on the base).
2. `moai-mcp-tools.md:75` and `moai-mcp-tools-catalogue.md:103-105`, local and template:
   "the Codex lane orchestrator" → "a Codex session" (shorter).
3. docs-site, ko first then en/ja/zh: `advanced/codex-dual-harness.md:33` drop the `-f`
   launch shape; `guides/mcp-server.md:180-184` replace the lane-orchestrator column value
   and the lane sentence with the Codex-session wording.
   The en sentence "inside a lane's shell" and the zh "泳道 shell" go with it.
4. Run `TestAlwaysLoadedTokenBudget` on the merge tree and on the merge commit's develop
   parent in the same run; record both figures in `progress.md` §E.2.

ACs: AC-CFR-019, AC-CFR-021..023.

### M5 — merge-window step: foreign-file disposition (lead-gated)

Executed only after the lead explicitly confirms, inside the lead's integration window.
This agent and the run lane never touch `.claude/worktrees/develop` before that
(REQ-CFR-023).

The six paths, as provided by the lead (research §R6):

- `internal/cli/mcp_factory_msg.go` (modified, +62)
- `internal/cli/mcp_factory_msg_test.go` (modified, +94)
- `internal/factorymsg/store.go` (modified, +4)
- `internal/cli/mcp_factory_push_live_test.go` (untracked)
- `reports/factory-cross-host-push-20260924.md` (untracked)
- `reports/factory-cross-host-push-20260924.html` (untracked)

Steps (by explicit pathspec only; the patch and reports land in the **primary
checkout's** `.moai/reports/t1242/`, which is gitignored):

0. Record the lead's confirmation in `progress.md` §E.2 as
   `m5_lead_confirmation: <ISO-8601 timestamp> <message or dispatch reference>`. No
   later step runs without this line.
1. `moai integration acquire --name <lane> --card t1242`.
2. Confirm no live writer holds the develop worktree (`moai session list --json`
   filtered to its cwd; liveness-probe each pid; a live or indeterminate entry stops
   the step and is reported to the lead).
3. `EnterWorktree(.claude/worktrees/develop)` — never `git -C` from another tree.
4. Re-read `git status --porcelain` there; stop if the set differs from the six above.
5. Preserve: write the diff of the three modified files plus the untracked test to
   `<primary>/.moai/reports/t1242/foreign-6.patch` (untracked content included, e.g.
   `git diff --no-index /dev/null <file>`); record its sha256 in `progress.md` §E.2
   after the confirmation line.
6. Revert the three modified files to `HEAD` and remove the untracked test, by pathspec.
7. Move the two reports into `<primary>/.moai/reports/t1242/`.
8. Re-read `git status --porcelain`: none of the six paths remain.
9. `ExitWorktree`, then `moai integration release`.

AC: AC-CFR-024.

## §G Anti-patterns

- Deleting `codex_factory_test.go` before moving `TestNextFactoryWorkerNumber`.
- Deleting `factory_lane_handoff_test.go` before moving the abandon-lane fixture.
- Removing the spawn pane-identity probe together with the factory branch (breaks the
  `-w --spawn` anchor).
- Asserting only `factoryLaunchEnabled(env) == false` for the env defense (conjunctive
  gate — a single-key regression passes it).
- Editing `.codex/config.toml` or `mcpServerEnvVarsValue` "for tidiness" (D1).
- Growing always-loaded text while replacing wording.

## §H Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| Token budget overflow from M4 wording (headroom 61) | Medium | Only shortening edits; AC-CFR-022 on merge tree |
| A hidden production caller of a handoff symbol appears on develop after `553e224f3` | Low | §C.3 re-grep; M3 stops on a hit |
| Windows direct launch regresses when factory code is removed | Medium | `GOOS=windows go build ./...`; keep Start→Wait semantics and exit-code propagation |
| hook tests touch the real lease DB | Medium | isolated `MOAI_HOME` in AC-CFR-018 |
| Operators with a live stale codex run cannot join | Low (intended) | REQ-CFR-010 message names `moai factory runs --retire` |
| Foreign patch lost before preservation | Low | M5 step 2 precedes step 3; sha256 recorded |

## §I Cross-references

- `spec.md` §C decisions D1-D4; `design.md`; `research.md`.
- Sync phase (manager-docs): mark SPEC-FACTORY-MIXED-HOOK-001,
  SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001, SPEC-DUAL-HARNESS-RECOVERY-001,
  SPEC-CODEX-LOCALMD-001 and SPEC-FACTORY-RUN-RETIRE-001 with `partially_superseded_by:
  [SPEC-CODEX-FACTORY-RETIRE-001]`; regenerate codemaps (`.moai/project/codemaps/modules.md:71`
  describes the retired `moai codex -f` entry); CHANGELOG entry.
- Follow-up card to suggest at sync: measure whether a directly started `codex`'s MCP
  server can attribute to a Claude lane (spec.md §E known limitation).
