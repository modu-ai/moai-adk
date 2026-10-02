# plan.md — SPEC-QUOTA-AWARE-SCHEDULING-001 (card t1347)

Tier L (reclassified from M at 0.3.0: 22 REQ / 22 AC and more than 15 files). Milestones are ordered by decision reversibility: the data model first (the decision most likely to change), then the policy keys, then the user-facing lane flow, then the mechanical surfaces. M0 is listed first for a different reason — commit order (REQ-QAS-016): the baseline must be its own commit before any implementation commit. No time estimates; priority and ordering only.

## §A Context

- Work tree: `.moai/worktrees/t1347`, branch `WT-quota-aware-scheduling`, planned at HEAD `c50da9c2f` (develop tip at planning time).
- Methodology: TDD (the run-phase default; every AC is decided by a Go test written RED first). The run phase starts only after the plan→run Kickoff gate is met (the autonomous form needs a plan-audit PASS at the Tier L threshold 0.85 and unchanged artifact hashes, `auto-semantics.md` §9.1).
- Artifacts (Tier L, 5 plus 2): `spec.md` (22 REQ, decisions D1-D8), `plan.md`, `acceptance.md` (22 AC, RED-now ledger), `design.md`, `research.md`, `decision-index.md` (open decisions, no recommendations), `progress.md` (§E skeleton). Decisions DO-1..DO-11 were resolved by the decision oracle, DO-3/-7/-8 by the leader's verdict (0.3.0); DO-12 (lane backend source) is open.
- Factory mode only. No kanban-specific code path is added or touched (REQ-QAS-015). Note for the removal of kanban mode: the Claude/GLM lane backend stamp is exported today by `internal/cli/kanban.go:565-567`; the gate's lane predicate therefore reads `MOAI_LAUNCH_PROVIDER` first (`launcher.go:231-248`, kanban-independent) and the kanban backend variable only as the fallback (`DO-10`).

## §B Known issues (injected, filtered to this SPEC)

- **Template-First (CLAUDE.local.md):** the shipped template `workflow.yaml` and the Go defaults change together; the local `.moai/config/sections/workflow.yaml` is the dogfood twin (gate on). Run `make build` before committing so the embedded templates match.
- **Config cache (`internal/config/cache.go:44`):** adding a field to the workflow config requires a cache schema bump (the observed AgentStopGuard / SettingsDriftGate / SlotLease defect); `configCacheSchemaVersion` goes 10 → 11 with a comment line in the established style.
- **Shipped-key guard (`shipped_key_reader_test.go:74`):** every shipped key needs a reader and an inventory row (`class: W`, `evidence: reader`) in `testdata/shipped_key_inventory.yaml`; a key with no production reader fails the guard.
- **Hardcoding prevention (§14):** thresholds and durations live in `internal/config/defaults.go` constants, env names in `envkeys.go` (this SPEC adds no env variable).
- **Subagent boundary (C-HRA-008):** CLI code never calls `AskUserQuestion`; the hold is a printed line and an exit status.
- **spec-lint heading convention (B6):** the SPEC carries `### Out of Scope — <topic>` H3 sub-headings with bullets (done).
- **Cross-platform:** `GOOS=windows GOARCH=amd64 go build ./...` must pass; the new files use only `filepath` and the standard library.
- **Test load:** `internal/cli` runs only with an anchored `-run`; heavy suites take a `moai slot` lease; background load is never spawned.
- **Throttle test coupling:** `TestWriteContextUsage_ThrottleSkipUnchanged` and `TestThrottleUnaffectedByModelAndEffort` pin today's throttle; the window-less path must keep them green untouched.

## §C Pre-flight (run phase, before the first change)

```bash
git branch --show-current
git rev-parse HEAD
go build ./...
GOOS=windows GOARCH=amd64 go build ./...
golangci-lint run --timeout=2m ./internal/statusline/... ./internal/config/... ./internal/cli/...
grep -rn "EnvMoaiLaunchProvider" internal/cli/launcher.go internal/cli/kanban.go
grep -rn "func resolveStateAnchor" internal/statusline
```

The last two close Unverified assumptions U3 and U6 (what token `moai cc -f lane` carries; whether a lane in a worktree writes its record under the project root the factory reads). Record the tokens found in `progress.md` §E.2 and fix the Claude predicate's accepted set from them.

## §D File map (indicative; confirmed in the run phase)

| File | Change | Milestone |
|---|---|---|
| `internal/statusline/testdata/qas_baseline_windowless_record.golden.json` | new (baseline) | M0 |
| `internal/cli/testdata/qas_baseline_factory_status.golden.json` | new (baseline) | M0 |
| `internal/cli/testdata/qas_baseline_todo_auto.golden.txt` | new (baseline: `--auto` output of a fixed fixture, Jev line stubbed) | M0 |
| `internal/statusline/context_usage.go` | record windows, schema 3, throttle payload, heartbeat, exhausted time | M1 |
| `internal/statusline/builder.go` | pass the stdin windows into the writer (call site at `:181`) | M1 |
| `internal/statusline/context_usage_quota_test.go` | AC-001..004 tests | M1 |
| `internal/statusline/quota.go` (new) | aggregator: freshest per window, rollover, fail-open, read-only | M2 |
| `internal/statusline/quota_test.go` (new) | AC-005, -006, -014 (statusline side) | M2 |
| `internal/config/types.go`, `defaults.go`, `cache.go` | `workflow.quota_gate` struct, defaults, schema bump | M2 |
| `internal/config/loader_quota_gate.go` (new) | single-block reader, default on every failure (pattern of `loader_slot_lease.go`) | M2 |
| `internal/config/testdata/shipped_key_inventory.yaml` | five inventory rows | M2 |
| `internal/config/workflow_quota_gate_test.go` (new) | AC-007 | M2 |
| `internal/template/templates/.moai/config/sections/workflow.yaml`, `.moai/config/sections/workflow.yaml` | template block (off) and local block (on) | M2 |
| `internal/cli/factory_quota.go` (new) | Claude-lane predicate, hold evaluation with in-wait latch, hold line, status block builder | M3 |
| `internal/cli/factory_card.go` | one boolean on the selection pass bypassing arms (b), (b2), (c) (single caller `:246`); gated variant of `factoryNextLeaseOnce` (the original keeps its signature); gate call in the `next` verb reusing `factoryNextNoCardExit`; status block in text and JSON | M3, M4 |
| `internal/cli/mcp_factory_card.go` | gate call in `handleFactoryNext` | M3 |
| `internal/cli/factory_quota_lanes.go` (new) | lane inventory (registry + liveness + newest session record), recommendation and warning formatter | M4 |
| `internal/cli/todo_auto.go`, `internal/cli/todo.go` | nil-inert quota seam on the `--auto` options, one printed line before each `accept`; production wiring at `todo.go:275-279` | M4 |
| `internal/cli/factory_quota_test.go`, `factory_quota_lanes_test.go` (new) | AC-006b, -008..-013, -014 (cli side), -017..-022 | M3-M5 |
| `internal/cli/integration.go` | warn-only line in acquire (beside the warnings at `:387-406`) | M5 |

The three files AC-QAS-014 sweeps are `internal/statusline/quota.go`, `internal/cli/factory_quota.go`, and `internal/cli/factory_quota_lanes.go`. Reads of `internal/kanban` (`LoadFactoryRegistry`, `FactoryProcessAlive`, `ReadAll`) are read-only uses of its API; no file under `internal/kanban/` is edited (AC-QAS-016).

## §E Milestones

### M0 — Baseline (own commit, precedes every other commit; REQ-QAS-016, AC-QAS-015)

1. On the unmodified tree, serialize a window-less record from a fixed input and commit its bytes as `qas_baseline_windowless_record.golden.json` (capture time, writer pid, and schema version are normalized by the test, not stored as the baseline's identity).
2. On the unmodified tree, capture `moai factory status --json` for a fixed one-card fixture and for the empty fixture; commit as `qas_baseline_factory_status.golden.json`.
2b. On the unmodified tree, capture the `moai todo --auto` output of a fixed two-card fixture with the Jev line stubbed (the `autoOptions.jev` seam, `todo_auto.go:201`) and a fake clock; commit as `qas_baseline_todo_auto.golden.txt`.
3. Commit these alone: `test(SPEC-QUOTA-AWARE-SCHEDULING-001): M0 baseline goldens measured pre-change (card t1347)`. Record the measuring commands, their verbatim outputs, and the tree SHA in `progress.md` §E.2. The commit graph, not the message, is the ordering witness.

### M1 — Record: windows, schema 3, throttle, exhausted time (REQ-QAS-001..004; most reversible, so first)

RED: `TestQAS_AC001..004` plus the two `AC002` tests fail (undefined symbols, then assertion failures); record the verbatim failing output (E8). GREEN: add the optional window fields (omitted when empty), bump the schema constant to 3, add window presence, integer-rounded percentage, and reset time to the throttle payload, add the heartbeat rewrite for window-carrying records only, add the sticky first-observed-exhausted time. The heartbeat (5m) and exhaustion percentage (100) are named constants in `internal/config/defaults.go`. Regression: the two pinned throttle tests and `TestReadsPreviousSchemaRecord` stay green.

### M2 — Aggregator and configuration (REQ-QAS-005..008)

RED: `TestQAS_AC005`, `AC006` (statusline side), `AC007`, `AC014` (statusline side). GREEN: the aggregator in `internal/statusline/quota.go` (parse only files changed within `max_age`; per-window freshest; reset when the reset time is not after now; unknown otherwise; no network, no process, no write); the `workflow.quota_gate` struct, defaults, single-block loader (every failure and every out-of-range value yields the default), cache schema 11, inventory rows, template block (`enabled: false`), local block (`enabled: true`). `make build` after the template edit.

### M3 — Lane gate: hold, hold line, in-wait latch (REQ-QAS-009..012; the user-facing flow)

RED: `TestQAS_AC006b`, `AC008`, `AC008b`, `AC009`, `AC010`, `AC011`, `AC011b`, `AC014` (cli side). GREEN: evaluate the gate before the lease in the `next` verb and in the MCP handler (one shared function, so the two surfaces cannot drift — the rule `factory_card.go:50-64` already follows for its predicates); the hold is evaluated once per invocation and handed to the selection pass as one boolean, so arm (a) (a card already assigned to the lane) still leases and arms (b), (b2), (c) are bypassed without a second predicate (DO-6; seam verified in `factory_card.go:384-521`); a held lane with nothing from arm (a) prints one hold line on the error stream, nothing on standard output, and returns status 3 (`factoryNextNoCardExit`, DO-5); `factory_lane_relaunch.go:62` and `codex_launcher.go:968` keep calling the ungated function; the `--wait` loop carries the latch (release below `hold - margin` or at the reset); gate failures of any kind fall through to the existing lease path. Gate only leasing; do not touch `stage`, `complete`, or `factoryNextSkipForBackend`. Re-run the anchored guards `TestSD_AC008_NextSelectionOrderAndOutput`, `TestSD_AC009_NextWaitLeasesOrTimesOut`, `TestSD_AC023_CodexNextSkipsUnadvanceableCard`.

M3 also lands the single pressure evaluation of REQ-QAS-017 (`TestQAS_AC017` is written RED in M3 and goes GREEN as each surface adopts it in M4 and M5): the lane gate calls it with the caller's backend; every other surface calls it without.

### M4 — Steering: lane inventory, `--auto` and status recommendation (REQ-QAS-013, -018..-022; leader verdict, user-facing)

RED: `TestQAS_AC012`, `AC018`, `AC019`, `AC020`, `AC021`, `AC022`. GREEN: (1) the lane inventory in `internal/cli/factory_quota_lanes.go` — registry rows through `kanban.LoadFactoryRegistry`, liveness through `kanban.FactoryProcessAlive`, backend from the newest session record per lane number through `kanban.ReadAll`, ambiguous or absent record counted unknown, every read failure yielding no candidates, no write (`DO-12`); (2) one formatter for the recommendation and the no-candidate warning, used by both surfaces; (3) a nil-inert quota seam on the `--auto` options (the convention of `todo_auto.go:195-209`) wired at `todo.go:275-279` and called immediately before each `accept` print (`todo_auto.go:262`) — printing only, no queue write; (4) the read-only quota block in `moai factory status` (text line, JSON key omitted when no data, pressure flag, candidates, warning marker, unknown count) compared against the M0 goldens. Re-run the existing guards `TestTodoAutoSerialCycle`, `TestTodoAutoCreatesNoFactoryLease`, `TestTodoAutoClearGuidancePerCard`, `TestFactoryStatusShowsHolderModePriority`.

### M5 — Integration-window warning and closure (REQ-QAS-014, -015; mechanical, last)

RED: `TestQAS_AC013`. GREEN: the warn-only stderr line in `moai integration acquire` (record, exit status, stdout unchanged; never blocks, refuses, or delays; DO-7 final). Close: the AC-QAS-015 and AC-QAS-016 evidence commands, the whole-package runs of `./internal/statusline` and `./internal/config`, `moai spec lint`, and `moai spec audit` for this SPEC.

## §F Self-verification (run-phase report, five-section form)

Each completion report carries Claim / Evidence (command plus verbatim output) / Baseline-attribution (this run, this tree, the HEAD SHA) / Gaps / Residual-risk, per `verification-claim-integrity.md` §3, with the E1-E8 attribution triple of `manager-develop-prompt-template.md`. A tool measurement also names the judging build's commit next to the tree HEAD (`verification-claim-integrity.md` §2.2); at planning time `moai version` reported `gc50da9c2f` for tree `c50da9c2f`.

## §G Unverified assumptions (not established by a command run in this plan phase)

- **U1** — The 5h/7d quota is shared by every Claude-backend session of one account. Given by the card; not measured here.
- **U2** — Used percentage never falls inside a window before its reset. Unmeasured; the SPEC does not depend on it (`DO-9`: a stale high is treated as unknown). The Claude Code doc calls the five-hour window "rolling".
- **U3** — Which token a `moai cc -f lane` session carries in `MOAI_LAUNCH_PROVIDER` / `MOAI_KANBAN_BACKEND` (the constants name `claude`; `resolveMode` defaults an empty mode to `claude`, `launcher.go:55-60`; the stamp site for factory lanes was read, the value on a real lane was not). Closed by the M3 pre-flight.
- **U4** — Whether the StopFailure hook process inherits the session environment, and whether a 429 turn-end triggers a final statusline render at 100%. Unmeasured; both are why `DO-1` defers the StopFailure stamp.
- **U5** — Whether several Claude accounts/profiles can write the same project's `context-usage/` directory (the auto-memory path carries a `claude-profiles` segment; the state directory was not inspected for multi-account writes). `DO-2`.
- **U6** — Whether a lane running inside a worktree writes its record under the project root the factory reads. `resolveStateAnchor` is the resolver (`state_anchor.go:22-35`, delegating to `stateanchor.Resolve`); its worktree behaviour was not read, only its comment (`builder.go:178-181`). `factoryContextAtHandoffThreshold` (`factory_card.go:987-1005`) already depends on the same fact. Closed by the M1 pre-flight.
- **U7** — That statusline renders at the boundaries a lane's tool calls cross. The doc lists new assistant message, `/compact`, permission mode, vim toggle, `refreshInterval`, a window reaching `resets_at`, and a warm cache reaching `expires_at`; whether the render also follows a tool-call result was not measured.
- **U8** — That a gateway `spend_limit` object never co-occurs with the windows in a way that confuses the reader; the SPEC ignores `spend_limit` (out of scope).
- **U9** — That GLM and Codex sessions carry no `rate_limits` in their statusline stdin. Given by the card statement; not measured (a GLM session's record models the z.ai model, `builder.go:168-173`).
- **U10** — The 554-minute stall and the weekly-limit session end are the card's figures; not re-measured.
- **U11** — The numeric defaults (90 / 95 / 5 / 30m / 5m heartbeat) have no measurement behind them (`DO-3`: oracle confidence 0.32, accepted as provisional by the leader; the values are unmeasured defaults, all four tunable as configuration keys).
- **U15** — That a Codex lane session writes a kanban session record with backend `gpt`. The hook writes records for any session with a lane label (`session_start_record.go:60-116`, `kanbanRoleFromEnv` `:180-187`) and the Codex launcher stamps `MOAI_KANBAN_BACKEND=gpt` (`codex_launcher.go:1038`), but whether the SessionStart hook runs with that environment under Codex was not observed. If it does not, Codex lanes are counted unknown and the no-candidate warning fires while a Codex lane exists (`DO-12`).
- **U16** — The registered pid of a lane is alive exactly while the lane is: the Claude and GLM launchers exec into the backend, so the pid is the session's (`factory_slots.go:36-40` comment, not live-tested); the Codex supervising loop registers the launcher's pid, which outlives each per-card session. Liveness (not backend) is what the pid decides here, so a Codex lane between cards still counts as live, which is the intended reading.
- **U17** — That the kanban session record store survives the removal of kanban mode. The lane inventory depends on `kanban.ReadAll` and on the lane SessionStart hook writing records; the removal's scope is not known here. If the records go, every lane is unknown and steering degrades to the warning-only path; the lane gate (D4) is unaffected.
- **U18** — That writing the registry's unused `workers.backend` column at claim time is feasible without touching kanban mode: the claim function and both inserts were read (`factory_slots.go:106,329`), the launchers that would pass the backend were not. This is the `DO-12` alternative, not adopted.
- **U19** — That adding one line to `moai todo --auto` output only under pressure cannot break the doc-parity tests around it (`todo_auto_doc_test.go`, `todo_skill_doc_parity_test.go` exist; they were not read) or the `/moai todo --auto` skill documentation of the output shape.
- **U20** — That `moai factory status` is what a factory leader actually reads between dispatches. The status command is read-only and lists cards (`factory_card.go:1351`); the leader doctrine's own reading surface is the card evidence and the queue, and no code makes the status command the leader's required read.
- **U21** — That the newest-record-per-lane-number rule picks the right session when a lane label is reused by a new session that has not yet written its record (a startup race): the older record would then name the previous incarnation's backend.
- **U14** — The Claude relaunch supervising loop (`factory_lane_relaunch.go:62`) leases through the ungated function and is therefore not held; whether that loop should be gated is not decided here (an amendment candidate, `spec.md` §E).
- **U12** — Exit status of the zero-count `grep` rows in `acceptance.md` §E: the harness surfaced none (G1). The deciding signal is the stdout count; the Go test decides each AC.
- **U13** — The "manual quota-reset gate" is located only as operator practice (a kickoff option and a scheduled wakeup recorded in `SPEC-CI-FLAKE-SERIES-001/progress.md:9-10`, and an auto-memory note), found by grep over code, rules, skills, docs, and memory; absence of a code gate rests on those greps, which cannot rule out a gate under a name they did not match.

## §H Risks

| Risk | Effect | Handling |
|---|---|---|
| Throttle regression | window-less records write more often or less often than today | window-less path untouched by construction; AC-002b and the pinned throttle tests |
| Heartbeat churn | one write per 5 minutes per live session | bounded; the percentage bucket already bounds writes to ~100 per window |
| Wrong Claude predicate | gate never holds (fail open) or holds a non-Claude lane | AC-011 table; U3 closed in pre-flight; `DO-10` |
| Thresholds mis-set | early parking or no relief | status block makes readings visible; `DO-3` (unmeasured defaults, accepted as provisional) |
| Lane inventory unknown backend | the no-candidate warning fires while a usable non-Claude lane exists | unknown count printed in the warning; AC-018/-020; `DO-12`, U15, U17, U21 |
| Steering line shifts `--auto` output | doc or parity tests break | pressure-off output byte-identical (AC-022); U19 |
| Hold and empty queue share status 3 | a caller reading only the status cannot tell them apart | the stderr `quota hold:` line is the discriminator (AC-009); `DO-5` |
| Stale record at lease time | reading up to `max_age` old | heartbeat for live sessions; unknown fails open |
| Kanban removal removes the backend stamp | fallback variable disappears | primary variable is the launch provider (`DO-10`) |

## §I Cross-references

- `SPEC-SESSION-TELEMETRY-001` (the record and its schema history), `SPEC-FACTORY-SELF-DISPATCH-001` (the `next` verb, REQ-SD-008/-009/-025), `SPEC-STATE-ANCHOR-001` (where the record lands).
- `.claude/rules/moai/development/verification-completeness.md` §2 (RED-now / green path pair), `.claude/rules/moai/core/verification-claim-integrity.md` §2.3 (ordering attribution).
- Neighbours, not scope: H1 automatic class routing, card t1241 (controller), card t1375 (cross-host managed sessions), the M5 Jev engine-choice proposal.
