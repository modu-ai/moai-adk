# acceptance.md — SPEC-FACTORY-SELF-DISPATCH-001 (card t1240)

## §A Scope of verification

The criteria verify the F2 lane: the launch shapes on cc, glm, and codex, the lane verbs and their
selection order, the per-card worktree, the six MCP tools and their tree resolution, the lane
predicate and permission boundary, integration through the window, the Codex harness limits, the
SessionStart next-card rule and clear policy, and the invariants on schema, allowlist, parent checkout,
and vocabulary. F1 record behavior and F3 controller behavior are not verified here.

## §B Test-environment constraint (binds every AC)

- Every test builds its own project root with `t.TempDir()` and its own fixture git repository; the
  MoAI home is redirected to a temporary directory. No test reads or writes the real `~/.moai`.
- **Fixture layout (binds AC-SD-006, -013, -018, -025):** the parent checkout has `main` checked out;
  the integration branch is `develop`, configured as the project's integration branch and checked out
  in a provisioned integration worktree `.claude/worktrees/develop`; the repository has no remote.
- `./internal/cli` and `./internal/hook` run only with an anchored `-run` selector naming one AC test
  (`-run '^TestSD_AC0NN_Name$'`). Whole-package runs of those two packages are prohibited for this card.
- **Pass convention:** an AC passes only when every command it names exits 0 and each test's verbose
  output carries a PASS line for exactly the test the command names, the name followed by a space
  (never a prefix match). `[no tests to run]` is a gap.
- No test starts a real `claude` or `codex` binary; launcher tests substitute the binary lookup and
  capture the argv and environment the launcher would hand to it.
- Waits and lease expiry are driven by an injected clock, never by sleeping.
- "Lane environment" means the role marker set to the role-value constant plus a lane label; "label-only
  environment" means a lane label with the marker unset.

## §C Traceability

| REQ | AC |
|---|---|
| REQ-SD-001 | AC-SD-001 |
| REQ-SD-002 | AC-SD-002 |
| REQ-SD-003 | AC-SD-003 |
| REQ-SD-004 | AC-SD-004 |
| REQ-SD-005 | AC-SD-005 |
| REQ-SD-006 | AC-SD-006 |
| REQ-SD-007 | AC-SD-007 |
| REQ-SD-008 | AC-SD-008 |
| REQ-SD-009 | AC-SD-009 |
| REQ-SD-010 | AC-SD-010 |
| REQ-SD-011 | AC-SD-011 |
| REQ-SD-012 | AC-SD-012 |
| REQ-SD-013 | AC-SD-013 |
| REQ-SD-014 | AC-SD-014 |
| REQ-SD-015 | AC-SD-015 |
| REQ-SD-016 | AC-SD-016 |
| REQ-SD-017 | AC-SD-017 |
| REQ-SD-018 | AC-SD-018 |
| REQ-SD-019 | AC-SD-019 |
| REQ-SD-020 | AC-SD-020 |
| REQ-SD-021 | AC-SD-021 |
| REQ-SD-022 | AC-SD-022 |
| REQ-SD-023 | AC-SD-013, AC-SD-025 |
| REQ-SD-024 | AC-SD-010, AC-SD-014 |
| REQ-SD-025 | AC-SD-023, AC-SD-024 |

## §D Acceptance criteria (Given-When-Then)

### AC-SD-001 — run gate
- **Given** the commit that adds the `develop_sha:` and `t1256_landed:` lines to progress.md §E.2 (the
  gate commit), **When** the branch history before it is read, **Then** no commit of this branch between
  `ed506740b` and the gate commit, excluding commits absorbed from develop, touches a path outside
  `.moai/`; the recorded `develop_sha` equals the develop SHA the run read (the first parent of the
  absorbing merge, or `git rev-parse develop` at gate time); and `t1256_landed` is `yes`.
- Verify (production-path check, prints nothing on pass):
  `G=$(git log -n1 --format=%H -S'develop_sha:' -- .moai/specs/SPEC-FACTORY-SELF-DISPATCH-001/progress.md) && git log --first-parent --no-merges --format=%H ed506740b.."$G"^ -- . ':!.moai' | grep -c . | grep -qx 0 && echo OK`
  prints `OK`.
- Verify (SHA and landing lines):
  `grep -E '^- (develop_sha: [0-9a-f]{40}|t1256_landed: yes)$' .moai/specs/SPEC-FACTORY-SELF-DISPATCH-001/progress.md | wc -l`
  prints `2`, and the `develop_sha` value equals the output of `git rev-parse develop` recorded beside it
  in §E.2 at gate time.

### AC-SD-002 — cc/glm lane launch
- **Given** a fixture factory run with a leader, **When** `moai cc -f lane` and `moai glm -f lane` are
  launched with the binary lookup substituted, **Then** each captured child environment carries the
  marker name constant equal to the value constant, the lane-label variable equal to `lane-<n>` (the next
  free number), and `MOAI_KANBAN_BACKEND` equal to `claude` and `glm` respectively, and the child's working
  directory is the parent checkout.
- Verify: `go test ./internal/cli -run '^TestSD_AC002_LaneLaunchStampsMarkerAndLabel$' -count=1 -v`

### AC-SD-003 — Codex per-card relaunch
- **Given** two operator-picked cards and a substituted `codex` binary that exits 0 after moving its card
  to `merge-ready`, **When** `moai codex -f lane` runs, **Then** the substitute is invoked twice, each time
  with its working directory set to that card's worktree and with the marker, the lane label, and
  `MOAI_KANBAN_BACKEND=gpt` in its environment, and the launcher exits 0 once `next` returns status 3.
- Verify: `go test ./internal/cli -run '^TestSD_AC003_CodexRelaunchPerCard$' -count=1 -v`

### AC-SD-004 — other Codex factory shapes
- **Given** each of `-f`, `--factory`, `--factory-run x`, `-f lane-2`, **When** passed to `moai codex`,
  **Then** stderr is exactly the one refusal line defined once for REQ-SD-004 (it carries
  `FACTORY_MODE_UNSUPPORTED_BACKEND` and names `moai codex -f lane`), the exit code is 1, and no child is
  started.
- Verify: `go test ./internal/cli -run '^TestSD_AC004_CodexOtherFactoryShapesRefused$' -count=1 -v`

### AC-SD-005 — git repository required
- **Given** a temporary directory that is not a git working tree, **When** a lane launch runs there,
  **Then** it exits non-zero with one line naming the git requirement, and the lane registry, factory
  database, and queue file are absent or byte-identical to before.
- Verify: `go test ./internal/cli -run '^TestSD_AC005_LaneLaunchRequiresGitRepo$' -count=1 -v`

### AC-SD-006 — no remote needed
- **Given** the §B fixture, **When** one card is carried by a Claude lane from `next` to `complete`,
  **Then** the card record reaches `merged-local`, and the recorded git command log contains no `fetch`
  and no `push`.
- Verify: `go test ./internal/cli -run '^TestSD_AC006_LaneCycleWithoutRemote$' -count=1 -v`

### AC-SD-007 — no headless engine
- **Given** every launch path this SPEC adds, **When** their captured argv are inspected, **Then** none
  contains `-p`/`--print` for `claude` or the `exec` subcommand for `codex`.
- Verify: `go test ./internal/cli -run '^TestSD_AC007_NoHeadlessEngineArgv$' -count=1 -v`

### AC-SD-008 — selection order, output, no-card status
- **Given**, in five sub-fixtures: (a) a card assigned to `lane-1` plus an unowned picked card; (b) an
  unowned picked card plus an older queued card; (c) only a card assigned to `lane-2`; (d) no picked card
  and two queued cards; (e) an empty queue; **When** `lane-1` runs `moai factory next`, **Then** (a) the
  assigned card is leased and the other stays `picked`; (b) the picked card is leased and the queued card
  stays `queued`; (c) nothing is leased and the exit status is 3; (d) the older queued card becomes
  `picked` in the queue and `leased` in the record and the newer stays `queued`; (e) stdout says no card
  is available and the exit status is 3. In (a), (b), (d) stdout carries the card id, its stage, its
  worktree name, and a PR/landed line equal to what `moai todo pr <id>` prints for that card.
  **And When** two lanes call `next` concurrently on fixture (b), **Then** exactly one leases the picked
  card and the other leases nothing from it.
- Verify: `go test ./internal/cli -run '^TestSD_AC008_NextSelectionOrderAndOutput$' -count=1 -v`

### AC-SD-009 — `--wait`
- **Given** an empty queue and an injected clock, **When** `next --wait` runs and a card is queued after
  two check intervals, **Then** that card is leased; **and When** no card arrives before the bound,
  **Then** exit status 3 is returned.
- Verify: `go test ./internal/cli -run '^TestSD_AC009_NextWaitLeasesOrTimesOut$' -count=1 -v`

### AC-SD-010 — `next` only on the parent checkout (CLI and MCP)
- **Given** a linked worktree of the fixture repository, **When** `moai factory next` runs with that
  worktree as working directory, and `factory_next` is called with `project_root` set to that worktree,
  **Then** both exit/return non-zero naming the parent checkout path, and the queue file and card rows are
  unchanged; with `project_root` set to the parent checkout, `factory_next` leases normally even though the
  server process's own working directory is the linked worktree.
- Verify: `go test ./internal/cli -run '^TestSD_AC010_NextRefusedOutsideParent$' -count=1 -v`

### AC-SD-011 — worktree per card
- **Given** a leased card with no recorded worktree, **When** `next` completes, **Then**
  `.claude/worktrees/<card-id>` exists, its branch starts with `WT-` and does not contain the card id, and
  the card record's worktree path equals that directory; **Given** `.claude/worktrees/<card-id>` already
  exists and no card record names it, **Then** `next` refuses and the card row version is unchanged;
  **Given** a card whose record names its own tree, **Then** `next` reuses it without creating another.
- Verify: `go test ./internal/cli -run '^TestSD_AC011_CardWorktreeCreateReuseRefuse$' -count=1 -v`

### AC-SD-012 — `stage`
- **Given** a card leased by `lane-1` in `plan`, **When** `lane-1` runs `moai factory stage <card>
  plan-audit` with a valid commit and artifact, **Then** the card is in `plan-audit`, the event actor is
  `lane-1`, and the lease expiry moved forward; **and When** `lane-2` runs the same, **Then** the F1
  refusal is printed and the exit is non-zero.
- Verify: `go test ./internal/cli -run '^TestSD_AC012_StageAppliesEdgeAndRenews$' -count=1 -v`

### AC-SD-013 — Claude `complete` through the integration worktree
- **Given** the §B fixture and a Claude lane holding the integration window, with the card's branch merged
  `--no-ff` into `develop` inside `.claude/worktrees/develop` and a re-measure file naming the merge commit,
  **When** the lane runs `complete`, **Then** the card is `merged-local` with that merge SHA recorded and
  the window is released; **Given** the integration branch is checked out only in the parent checkout,
  **Then** `complete` refuses saying the integration worktree is not provisioned and the card row is
  unchanged.
- Verify: `go test ./internal/cli -run '^TestSD_AC013_ClaudeCompleteViaIntegrationWorktree$' -count=1 -v`

### AC-SD-014 — MCP ↔ CLI equivalence and `project_root`
- **Given** twin fixtures, **When** each of the six tools is called with `project_root` on one and its CLI
  counterpart run from the same tree on the other (a success and a refusal case each), **Then** the
  resulting card rows, queue files, and refusal text are equal; **and When** `factory_next`,
  `factory_stage`, or `factory_complete` is called without `project_root`, or with a path that is not a
  MoAI project root, **Then** it is rejected naming the argument or the path.
- Verify: `go test ./internal/cli -run '^TestSD_AC014_MCPMatchesCLIWithProjectRoot$' -count=1 -v`

### AC-SD-015 — lane predicate and queue allowlist
- **Given** a lane environment, **When** the test walks every subcommand of the `moai todo` cobra tree
  and calls `todo_add`, **Then** every subcommand outside {bare `todo`, `list`, `history`, `why`, `pr`,
  `triage`} and `todo_add` exit non-zero naming the lane boundary with the queue file's bytes unchanged,
  and every allowlisted one succeeds; the walked set equals the tree's actual subcommand set, so a
  subcommand added later is covered without editing the test.
- Verify: `go test ./internal/cli -run '^TestSD_AC015_LaneQueueAllowlistWalk$' -count=1 -v`
- **Given** a label-only environment, **When** `moai factory next`, `stage`, and `complete` are invoked,
  **Then** each is refused as not a lane session; and `moai todo add` is not refused by the lane guard.
- Verify: `go test ./internal/cli -run '^TestSD_AC015_LabelOnlyIsNotALane$' -count=1 -v`

### AC-SD-016 — lane cannot decide
- **Given** a card in `kickoff` and a lane environment, **When** `moai factory decide` and
  `factory_decide` are invoked, **Then** both are refused and the card row is unchanged; in a non-lane
  environment the same CLI call succeeds.
- Verify: `go test ./internal/cli -run '^TestSD_AC016_LaneDecideRefused$' -count=1 -v`

### AC-SD-017 — marker via constants; guard live
- **Given** the environment captured from a lane launch, **When** the contract guard classifies
  `moai contract sign SPEC-X --signer llm` under it, **Then** the guard denies; and a source scan of the
  production files that stamp or compare the marker finds the name and value only through the
  `internal/config` constants.
- Verify: `go test ./internal/cli -run '^TestSD_AC017_StampedMarkerArmsContractGuard$' -count=1 -v`

### AC-SD-018 — parent checkout untouched
- **Given** the §B fixture (parent on `main`), **When** the full lane cycle of AC-SD-006 ends, **Then** in
  the parent checkout `git status --porcelain` (excluding `.claude/worktrees/`) is empty, and `HEAD` and
  the checked-out branch (`main`) equal their values before the cycle.
- Verify: `go test ./internal/cli -run '^TestSD_AC018_ParentCheckoutUntouched$' -count=1 -v`

### AC-SD-019 — next-card rule
- **Given** a lane environment, **When** SessionStart runs with source `startup` under each of the three
  clear policies and with source `clear`, for each of en, ko, ja, zh, **Then** additionalContext carries
  the next-card rule in that language, names each of the six MCP tools together with its CLI equivalent,
  and states that lane queue promotion is operator-authorized; **Given** a leader environment and an
  environment with no factory keys, **Then** no next-card rule is present for either source.
- Verify: `go test ./internal/hook -run '^TestSD_AC019_NextCardRuleInjection$' -count=1 -v`

### AC-SD-020 — clear policies
- **Given** each policy and an injected context-usage record, **When** a card completes, **Then**
  `clear-each` prints exactly one `/clear` request; `clear-when-full` prints none below the threshold and
  one at or above it; `relaunch` prints one end-session request and the supervising launcher starts a new
  session for the next card after the child exits; with no policy given, the behavior equals
  `clear-each`.
- Verify: `go test ./internal/cli -run '^TestSD_AC020_ClearPolicies$' -count=1 -v`

### AC-SD-021 — legacy spellings
- **Given** `moai cc`, `moai glm`, `moai codex`, the new verbs, and the MCP tools, **When** given
  `worker-1`, `agent-1`, `worker`, `agent`, or `lead` where a lane label, role token, or leader label is
  expected, **Then** each is refused with the REQ-RNC-003/-005/-007 message naming the canonical form
  (on `moai codex` too, not the REQ-SD-004 line), and no new user-facing string contains `worker` or
  `agent` as a role or `lead` as the leader noun.
- Verify: `go test ./internal/cli -run '^TestSD_AC021_LegacySpellingsRefused$' -count=1 -v`

### AC-SD-022 — schemas and allowlist frozen
- **Given** the merge tree, **When** `mcpServerEnvVarsValue` and every schema statement (`CREATE TABLE`,
  `ALTER TABLE`, index) in the production files of `internal/homestate`, `internal/factorymsg`, and
  `internal/kanban` are compared with a snapshot taken from the run-start develop tree, **Then** they are
  byte-identical.
- Verify: `go test ./internal/codexwiring -run '^TestSD_AC022_EnvVarsAllowlistFrozen$' -count=1 -v`
- Verify: `go test ./internal/cli -run '^TestSD_AC022_SchemaStatementsFrozen$' -count=1 -v`

### AC-SD-023 — Codex lane never re-leases what it cannot advance
- **Given** a Codex lane that brought a card to `merge-ready` and an injected clock advanced past the lease
  expiry (the card is back in `assigned` to that lane with stage `merge-ready`), and no other card,
  **When** the lane runs `moai factory next`, **Then** the exit status is 3 and the card row is unchanged;
  a Claude lane owning the same card would lease it.
- Verify: `go test ./internal/cli -run '^TestSD_AC023_CodexNextSkipsUnadvanceableCard$' -count=1 -v`

### AC-SD-024 — Codex merge edge refused on every path
- **Given** a card in `merge-ready` leased by a lane whose `MOAI_KANBAN_BACKEND` is `gpt`, **When** the lane
  requests `merging` through each path, **Then** each is refused and the card row is unchanged:
- Verify (`complete`): `go test ./internal/cli -run '^TestSD_AC024_CodexMergeRefusedComplete$' -count=1 -v`
- Verify (`stage … merging`): `go test ./internal/cli -run '^TestSD_AC024_CodexMergeRefusedStage$' -count=1 -v`
- Verify (MCP `factory_complete` and `factory_stage`): `go test ./internal/cli -run '^TestSD_AC024_CodexMergeRefusedMCP$' -count=1 -v`

### AC-SD-025 — integration window serializes lanes
- **Given** the §B fixture and two Claude lanes each with a card in `merge-ready`, **When** `lane-1` holds
  the integration window and `lane-2` runs `complete`, **Then** `lane-2` is refused naming `lane-1` as
  holder and its card row is unchanged; after `lane-1` completes and releases, `lane-2`'s `complete`
  succeeds and both merges are on `develop` in the integration worktree.
- Verify: `go test ./internal/cli -run '^TestSD_AC025_IntegrationWindowSerializes$' -count=1 -v`

## §E Edge cases

- A card returns from `kickoff` approved: it is `assigned` to its lane with stage `run`, and `next`
  re-enters its recorded tree (AC-SD-011 reuse arm).
- Queue promotion writes the queue first and the record second; if the record write fails, the card stays
  `picked` and unowned and the next `next` takes it as an unowned picked card (design.md §3).

## §F Quality gate and Definition of Done

- Every AC above passes under §B's convention on the merge tree.
- `moai spec lint --strict` reports no finding for this SPEC.
- golangci-lint (CI version) clean on changed packages; scoped tests per plan.md §E green.
- Sync phase: SPEC-CODEX-FACTORY-RETIRE-001 carries `partially_superseded_by`.
