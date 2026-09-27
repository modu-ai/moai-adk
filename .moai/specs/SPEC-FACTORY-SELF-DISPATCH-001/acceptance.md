# acceptance.md — SPEC-FACTORY-SELF-DISPATCH-001 (card t1240)

## §A Scope of verification

The criteria verify the F2 lane: the launch shapes on cc, glm, and codex, the lane verbs and their
selection order, the per-card worktree, the six MCP tools, the lane permission boundary, the
SessionStart next-card rule and clear policy, and the invariants on schema, allowlist, parent checkout,
and vocabulary. F1 record behavior and F3 controller behavior are not verified here.

## §B Test-environment constraint (binds every AC)

- Every test builds its own project root with `t.TempDir()` and its own fixture git repository; the
  MoAI home is redirected to a temporary directory. No test reads or writes the real `~/.moai`.
- `./internal/cli` and `./internal/hook` run only with an anchored `-run` selector naming one AC test
  (`-run '^TestSD_AC0NN_Name$'`). Whole-package runs of those two packages are prohibited for this card.
- **Pass convention:** an AC passes only when its command exits 0 and the verbose output carries a
  PASS line for exactly the test the command names, the name followed by a space (never a prefix
  match). `[no tests to run]` is a gap.
- No test starts a real `claude` or `codex` binary; launcher tests substitute the binary lookup and
  capture the argv and environment the launcher would hand to it.
- Waits and lease expiry are driven by an injected clock, never by sleeping.

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
| REQ-SD-008 | AC-SD-008a, AC-SD-008b, AC-SD-008c |
| REQ-SD-009 | AC-SD-009 |
| REQ-SD-010 | AC-SD-010 |
| REQ-SD-011 | AC-SD-011a, AC-SD-011b |
| REQ-SD-012 | AC-SD-012 |
| REQ-SD-013 | AC-SD-013a, AC-SD-013b |
| REQ-SD-014 | AC-SD-014 |
| REQ-SD-015 | AC-SD-015 |
| REQ-SD-016 | AC-SD-016 |
| REQ-SD-017 | AC-SD-017 |
| REQ-SD-018 | AC-SD-018 |
| REQ-SD-019 | AC-SD-019a, AC-SD-019b |
| REQ-SD-020 | AC-SD-020 |
| REQ-SD-021 | AC-SD-021 |
| REQ-SD-022 | AC-SD-022 |

## §D Acceptance criteria (Given-When-Then)

### AC-SD-001 — run gate records develop and t1256
- **Given** the run phase has started, **When** progress.md §E.2 is read, **Then** it carries a
  `develop_sha:` line with a hex SHA and a `t1256_landed:` line whose value is `yes`, and no production
  file differs from `ed506740b` in any commit made before that line.
- Verify: `grep -cE '^- (develop_sha: [0-9a-f]{7,40}|t1256_landed: yes)$' .moai/specs/SPEC-FACTORY-SELF-DISPATCH-001/progress.md` prints `2`.

### AC-SD-002 — cc/glm lane launch
- **Given** a fixture factory run with a leader, **When** `moai cc -f lane` and `moai glm -f lane` are
  launched with the binary lookup substituted, **Then** each captured child environment carries the
  marker name constant equal to the value constant and the lane-label variable equal to `lane-<n>`
  (the next free number), and the child's working directory is the parent checkout.
- Verify: `go test ./internal/cli -run '^TestSD_AC002_LaneLaunchStampsMarkerAndLabel$' -count=1 -v`

### AC-SD-003 — Codex per-card relaunch
- **Given** two operator-picked cards and a substituted `codex` binary that exits 0, **When**
  `moai codex -f lane` runs, **Then** the substitute is invoked twice, each time with its working
  directory set to that card's worktree and with the marker and lane label in its environment, and the
  launcher exits 0 after the queue is empty.
- Verify: `go test ./internal/cli -run '^TestSD_AC003_CodexRelaunchPerCard$' -count=1 -v`

### AC-SD-004 — other Codex factory shapes stay refused
- **Given** each of `-f`, `--factory`, `--factory-run x`, `-f lane-2`, `-f worker`, `-f agent`, **When**
  passed to `moai codex`, **Then** stderr carries `FACTORY_MODE_UNSUPPORTED_BACKEND`, the exit code is 1,
  and no child is started.
- Verify: `go test ./internal/cli -run '^TestSD_AC004_CodexOtherFactoryShapesRefused$' -count=1 -v`

### AC-SD-005 — git repository required
- **Given** a temporary directory that is not a git working tree, **When** a lane launch runs there,
  **Then** it exits non-zero with one line naming the git requirement, and the lane registry, factory
  database, and queue file are absent or byte-identical to before.
- Verify: `go test ./internal/cli -run '^TestSD_AC005_LaneLaunchRequiresGitRepo$' -count=1 -v`

### AC-SD-006 — no remote needed
- **Given** a fixture repository with no remote, **When** one card is carried by a Claude lane from
  `next` to `complete`, **Then** the card record reaches `merged-local`, and the recorded git command log
  contains no `fetch` and no `push`.
- Verify: `go test ./internal/cli -run '^TestSD_AC006_LaneCycleWithoutRemote$' -count=1 -v`

### AC-SD-007 — no headless engine
- **Given** every launch path this SPEC adds, **When** their captured argv are inspected, **Then** none
  contains `-p`/`--print` for `claude` or the `exec` subcommand for `codex`.
- Verify: `go test ./internal/cli -run '^TestSD_AC007_NoHeadlessEngineArgv$' -count=1 -v`

### AC-SD-008a — own assigned card first
- **Given** a card assigned to `lane-1` and an unowned picked card, **When** `lane-1` runs
  `moai factory next`, **Then** the assigned card is leased and the other stays `picked`.
- Verify: `go test ./internal/cli -run '^TestSD_AC008a_NextPrefersOwnAssigned$' -count=1 -v`

### AC-SD-008b — queue promotion in queue order
- **Given** no picked card and two queued cards, **When** a lane runs `moai factory next`, **Then** the
  older queued card becomes `picked` in the queue and `leased` in the record, and the newer stays
  `queued`.
- Verify: `go test ./internal/cli -run '^TestSD_AC008b_NextPromotesOldestQueued$' -count=1 -v`

### AC-SD-008c — nothing available
- **Given** an empty queue and no assignable card, **When** a lane runs `moai factory next`, **Then** it
  prints that no card is available and exits with the documented no-card status, distinct from 0 and 1.
- Verify: `go test ./internal/cli -run '^TestSD_AC008c_NextNoCardStatus$' -count=1 -v`

### AC-SD-009 — `--wait`
- **Given** an empty queue and an injected clock, **When** `next --wait` runs and a card is queued after
  two check intervals, **Then** that card is leased; **and When** no card arrives before the bound,
  **Then** the no-card status of AC-SD-008c is returned.
- Verify: `go test ./internal/cli -run '^TestSD_AC009_NextWaitLeasesOrTimesOut$' -count=1 -v`

### AC-SD-010 — `next` only from the parent checkout
- **Given** a working directory inside a linked worktree of the fixture repository, **When** `next`
  runs there, **Then** it exits non-zero, names the parent checkout path, and the queue file and card
  rows are unchanged.
- Verify: `go test ./internal/cli -run '^TestSD_AC010_NextRefusedOutsideParent$' -count=1 -v`

### AC-SD-011a — new tree per card
- **Given** a leased card with no recorded worktree, **When** `next` completes, **Then**
  `.claude/worktrees/<card-id>` exists, its branch starts with `WT-` and does not contain the card id,
  and the card record's worktree path equals that directory.
- Verify: `go test ./internal/cli -run '^TestSD_AC011a_NextCreatesCardWorktree$' -count=1 -v`

### AC-SD-011b — no reuse of a foreign tree
- **Given** `.claude/worktrees/<card-id>` already exists and no card record names it, **When** `next`
  leases that card, **Then** it refuses, and the card row version is unchanged; **and Given** a card
  whose record names its own tree, **Then** `next` reuses it without creating another.
- Verify: `go test ./internal/cli -run '^TestSD_AC011b_NextRefusesForeignTree$' -count=1 -v`

### AC-SD-012 — `stage`
- **Given** a card leased by `lane-1` in `plan`, **When** `lane-1` runs `moai factory stage <card>
  plan-audit` with a valid commit and artifact, **Then** the card is in `plan-audit`, the event actor is
  `lane-1`, and the lease expiry moved forward; **and When** `lane-2` runs the same, **Then** the F1
  refusal is printed and the exit is non-zero.
- Verify: `go test ./internal/cli -run '^TestSD_AC012_StageAppliesEdgeAndRenews$' -count=1 -v`

### AC-SD-013a — Claude `complete` reaches `merged-local`
- **Given** a card in `merge-ready` whose branch is merged `--no-ff` into the configured base branch
  with a re-measure file naming the merge commit, **When** a Claude lane runs `complete`, **Then** the
  card is `merged-local` with that merge SHA recorded.
- Verify: `go test ./internal/cli -run '^TestSD_AC013a_ClaudeCompleteMergedLocal$' -count=1 -v`

### AC-SD-013b — Codex `complete` stops at `merge-ready`
- **Given** a card in `sync-audit` with a PASS verdict, **When** a Codex lane runs `complete`, **Then**
  the card is `merge-ready` and a following request for `merging` from that lane is refused.
- Verify: `go test ./internal/cli -run '^TestSD_AC013b_CodexCompleteStopsAtMergeReady$' -count=1 -v`

### AC-SD-014 — MCP ↔ CLI equivalence
- **Given** twin fixtures, **When** each of the six tools is called on one and its CLI counterpart on
  the other with the same inputs (success and one refusal case each), **Then** the resulting card rows,
  queue files, and refusal text are equal.
- Verify: `go test ./internal/cli -run '^TestSD_AC014_MCPMatchesCLI$' -count=1 -v`

### AC-SD-015 — lane cannot mutate the queue
- **Given** a session environment carrying the lane role, **When** `moai todo add`, `next <n>`, `done`,
  `drop`, `edit`, `move`, `unpick`, `relate` and the `todo_add` tool are invoked, **Then** each exits
  non-zero naming the lane boundary and the queue file's bytes are unchanged; `moai todo list` succeeds.
- Verify: `go test ./internal/cli -run '^TestSD_AC015_LaneQueueMutationRefused$' -count=1 -v`

### AC-SD-016 — lane cannot decide
- **Given** a card in `kickoff` and a lane-role environment, **When** `moai factory decide` and
  `factory_decide` are invoked, **Then** both are refused and the card row is unchanged; without the lane
  role the same call succeeds.
- Verify: `go test ./internal/cli -run '^TestSD_AC016_LaneDecideRefused$' -count=1 -v`

### AC-SD-017 — marker via constants; guard live
- **Given** the environment captured from a lane launch, **When** the contract guard classifies
  `moai contract sign SPEC-X --signer llm` under it, **Then** the guard denies; and a source scan of the
  production files that stamp or compare the marker finds the name and value only through the
  `internal/config` constants.
- Verify: `go test ./internal/cli -run '^TestSD_AC017_StampedMarkerArmsContractGuard$' -count=1 -v`

### AC-SD-018 — parent checkout untouched
- **Given** the fixture of AC-SD-006, **When** the full lane cycle ends, **Then** in the parent checkout
  `git status --porcelain` (excluding `.claude/worktrees/`) is empty and `HEAD` and the current branch
  equal their values before the cycle.
- Verify: `go test ./internal/cli -run '^TestSD_AC018_ParentCheckoutUntouched$' -count=1 -v`

### AC-SD-019a — next-card rule on clear
- **Given** a lane-role environment, **When** SessionStart runs with source `clear` for each of en, ko,
  ja, zh, **Then** additionalContext carries the next-card rule in that language and names each of the
  six MCP tools together with its CLI equivalent.
- Verify: `go test ./internal/hook -run '^TestSD_AC019a_ClearInjectsNextCardRule$' -count=1 -v`

### AC-SD-019b — no rule for leader or non-factory sessions
- **Given** a leader environment and an environment with no factory keys, **When** SessionStart runs
  with source `clear`, **Then** no next-card rule is present.
- Verify: `go test ./internal/hook -run '^TestSD_AC019b_NoRuleOutsideLanes$' -count=1 -v`

### AC-SD-020 — clear policies
- **Given** each policy, **When** a card completes, **Then** `clear-each` prints exactly one `/clear`
  request; `clear-when-full` prints none below the threshold (injected usage) and one above it;
  `relaunch` has the supervising launcher start a new session for the next card; with no policy given,
  the behavior equals `clear-each`.
- Verify: `go test ./internal/cli -run '^TestSD_AC020_ClearPolicies$' -count=1 -v`

### AC-SD-021 — vocabulary
- **Given** the new verbs, MCP tools, and launch shapes, **When** given `worker-1`, `agent-1`, `worker`,
  `agent`, or `lead` where a lane label or role is expected, **Then** each is refused with an error naming
  the canonical form, and no new user-facing string contains `worker` or `agent` as a role or `lead` as
  the leader noun.
- Verify: `go test ./internal/cli -run '^TestSD_AC021_LegacySpellingsRefused$' -count=1 -v`

### AC-SD-022 — schemas and allowlist frozen
- **Given** the merge tree, **When** the `CREATE TABLE`, `ALTER TABLE`, and index statements of the
  factory, queue, and broker stores and `mcpServerEnvVarsValue` are compared with develop at run start,
  **Then** they are byte-identical.
- Verify: `go test ./internal/codexwiring -run '^TestSD_AC022_EnvVarsAllowlistFrozen$' -count=1 -v`
  and `go test ./internal/cli -run '^TestSD_AC022_SchemaStatementsFrozen$' -count=1 -v` (the test
  extracts every schema statement from the production files of `internal/homestate`,
  `internal/factorymsg`, and `internal/kanban` and compares them with a snapshot taken from the
  run-start develop tree)

## §E Edge cases

- Two lanes race `next` for one card: F1's version check lets one win; the loser re-selects (covered by
  AC-SD-008a's fixture with a second concurrent caller).
- A card returns from `kickoff` approved: it is `assigned` to its lane with stage `run`, and `next`
  re-enters its recorded tree (AC-SD-011b second arm).
- A lane label present with the marker unset (agent unset it): CLI refusals also accept the label as
  the lane signal (design.md §5).

## §F Quality gate and Definition of Done

- Every AC above passes under §B's convention on the merge tree.
- `moai spec lint` reports no error for this SPEC.
- golangci-lint (CI version) clean on changed packages; scoped tests per plan.md §E green.
- Sync phase: SPEC-CODEX-FACTORY-RETIRE-001 carries `partially_superseded_by`.
