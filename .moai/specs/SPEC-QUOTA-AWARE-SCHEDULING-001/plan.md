# plan.md — SPEC-QUOTA-AWARE-SCHEDULING-001 (card t1347)

Tier M. Milestones are ordered by decision reversibility: the data model first (the decision most likely to change), then the policy keys, then the user-facing lane flow, then the mechanical surfaces. M0 is listed first for a different reason — commit order (REQ-QAS-016): the baseline must be its own commit before any implementation commit. No time estimates; priority and ordering only.

## §A Context

- Work tree: `.moai/worktrees/t1347`, branch `WT-quota-aware-scheduling`, planned at HEAD `c50da9c2f` (develop tip at planning time).
- Methodology: TDD (the run-phase default; every AC is decided by a Go test written RED first). The run phase starts only after the plan→run Kickoff gate is met (the autonomous form needs a plan-audit PASS at the Tier M threshold 0.80 and unchanged artifact hashes, `auto-semantics.md` §9.1).
- Artifacts: `spec.md` (16 REQ, decisions D1-D7), `acceptance.md` (16 AC, RED-now ledger), `decision-index.md` (open decisions, no recommendations), `progress.md` (§E skeleton).
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
| `internal/cli/factory_card.go` | gate call in the `next` verb, hold status constant, status block in text and JSON | M3, M4 |
| `internal/cli/mcp_factory_card.go` | gate call in `handleFactoryNext` | M3 |
| `internal/cli/factory_quota_test.go` (new) | AC-006b, -008..-013, -014 (cli side) | M3, M4 |
| `internal/cli/integration.go` | warn-only line in acquire (beside the warnings at `:387-406`) | M4 |

The two files AC-QAS-014 sweeps are `internal/statusline/quota.go` and `internal/cli/factory_quota.go`.

## §E Milestones

### M0 — Baseline (own commit, precedes every other commit; REQ-QAS-016, AC-QAS-015)

1. On the unmodified tree, serialize a window-less record from a fixed input and commit its bytes as `qas_baseline_windowless_record.golden.json` (capture time, writer pid, and schema version are normalized by the test, not stored as the baseline's identity).
2. On the unmodified tree, capture `moai factory status --json` for a fixed one-card fixture and for the empty fixture; commit as `qas_baseline_factory_status.golden.json`.
3. Commit these alone: `test(SPEC-QUOTA-AWARE-SCHEDULING-001): M0 baseline goldens measured pre-change (card t1347)`. Record the measuring commands, their verbatim outputs, and the tree SHA in `progress.md` §E.2. The commit graph, not the message, is the ordering witness.

### M1 — Record: windows, schema 3, throttle, exhausted time (REQ-QAS-001..004; most reversible, so first)

RED: `TestQAS_AC001..004` plus the two `AC002` tests fail (undefined symbols, then assertion failures); record the verbatim failing output (E8). GREEN: add the optional window fields (omitted when empty), bump the schema constant to 3, add window presence, integer-rounded percentage, and reset time to the throttle payload, add the heartbeat rewrite for window-carrying records only, add the sticky first-observed-exhausted time. The heartbeat (5m) and exhaustion percentage (100) are named constants in `internal/config/defaults.go`. Regression: the two pinned throttle tests and `TestReadsPreviousSchemaRecord` stay green.

### M2 — Aggregator and configuration (REQ-QAS-005..008)

RED: `TestQAS_AC005`, `AC006` (statusline side), `AC007`, `AC014` (statusline side). GREEN: the aggregator in `internal/statusline/quota.go` (parse only files changed within `max_age`; per-window freshest; reset when the reset time is not after now; unknown otherwise; no network, no process, no write); the `workflow.quota_gate` struct, defaults, single-block loader (every failure and every out-of-range value yields the default), cache schema 11, inventory rows, template block (`enabled: false`), local block (`enabled: true`). `make build` after the template edit.

### M3 — Lane gate: hold, hold line, in-wait latch (REQ-QAS-009..012; the user-facing flow)

RED: `TestQAS_AC006b`, `AC008`, `AC008b`, `AC009`, `AC010`, `AC011`, `AC011b`, `AC014` (cli side). GREEN: evaluate the gate before the lease in the `next` verb and in the MCP handler (one shared function, so the two surfaces cannot drift — the rule `factory_card.go:50-64` already follows for its predicates); a held lane prints one hold line and returns the hold status; the `--wait` loop carries the latch (release below `hold - margin` or at the reset); gate failures of any kind fall through to the existing lease path. Gate only leasing; do not touch `stage`, `complete`, or `factoryNextSkipForBackend`. Re-run the anchored guards `TestSD_AC008_NextSelectionOrderAndOutput`, `TestSD_AC009_NextWaitLeasesOrTimesOut`, `TestSD_AC023_CodexNextSkipsUnadvanceableCard`.

### M4 — Observability and the integration-window warning (REQ-QAS-013, -014, -015; mechanical, last)

RED: `TestQAS_AC012`, `AC013`. GREEN: the read-only quota block in `moai factory status` (text line, JSON key omitted when no data; compare to the M0 golden); the warn-only stderr line in `moai integration acquire` (record, exit status, stdout unchanged). Close: the AC-QAS-015 and AC-QAS-016 evidence commands, the whole-package runs of `./internal/statusline` and `./internal/config`, `moai spec lint`, and `moai spec audit` for this SPEC.

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
- **U11** — The numeric defaults (90 / 95 / 5 / 30m / 5m heartbeat) have no measurement behind them (`DO-3`).
- **U12** — Exit status of the zero-count `grep` rows in `acceptance.md` §E: the harness surfaced none (G1). The deciding signal is the stdout count; the Go test decides each AC.
- **U13** — The "manual quota-reset gate" is located only as operator practice (a kickoff option and a scheduled wakeup recorded in `SPEC-CI-FLAKE-SERIES-001/progress.md:9-10`, and an auto-memory note), found by grep over code, rules, skills, docs, and memory; absence of a code gate rests on those greps, which cannot rule out a gate under a name they did not match.

## §H Risks

| Risk | Effect | Handling |
|---|---|---|
| Throttle regression | window-less records write more often or less often than today | window-less path untouched by construction; AC-002b and the pinned throttle tests |
| Heartbeat churn | one write per 5 minutes per live session | bounded; the percentage bucket already bounds writes to ~100 per window |
| Wrong Claude predicate | gate never holds (fail open) or holds a non-Claude lane | AC-011 table; U3 closed in pre-flight; `DO-10` |
| Thresholds mis-set | early parking or no relief | status block makes readings visible; `DO-3` |
| Stale record at lease time | reading up to `max_age` old | heartbeat for live sessions; unknown fails open |
| Kanban removal removes the backend stamp | fallback variable disappears | primary variable is the launch provider (`DO-10`) |

## §I Cross-references

- `SPEC-SESSION-TELEMETRY-001` (the record and its schema history), `SPEC-FACTORY-SELF-DISPATCH-001` (the `next` verb, REQ-SD-008/-009/-025), `SPEC-STATE-ANCHOR-001` (where the record lands).
- `.claude/rules/moai/development/verification-completeness.md` §2 (RED-now / green path pair), `.claude/rules/moai/core/verification-claim-integrity.md` §2.3 (ordering attribution).
- Neighbours, not scope: H1 automatic class routing, card t1241 (controller), card t1375 (cross-host managed sessions), the M5 Jev engine-choice proposal.
