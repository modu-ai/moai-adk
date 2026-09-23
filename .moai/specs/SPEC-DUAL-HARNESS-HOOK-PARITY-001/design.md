# Design — SPEC-DUAL-HARNESS-HOOK-PARITY-001

This file records the design decisions the run phase implements. Each decision names the
alternatives considered and the measured fact it rests on (research.md §R1). Decisions marked
**proposed** are confirmed at plan audit / Implementation Kickoff. The six clarifications Q1–Q6 were
resolved on 2026-09-23. plan.md §C records each decision and its source.

## §D1 Normalized decision model

Every decision-bearing handler result is normalized before harness translation (design §08):

| Normalized | Meaning | Claude rendering | Codex rendering |
|---|---|---|---|
| `allow` | proceed | `{}` / permissionDecision allow | `{}` |
| `deny` | refuse with reason | permissionDecision deny / decision block | permissionDecision deny + non-empty reason / decision block + reason |
| `needs_input` | a human must decide | permissionDecision ask | **fail-closed**: deny with reason naming the required input (§D5) |
| `retryable_error` | handler failed, retry is safe | advisory message, no decision | advisory record, no decision |
| `fatal_error` | handler failed on a decision-bearing event | event-specific fail-closed | event-specific fail-closed |

The translation table is data (one row per event × normalized decision × harness), so the
never-loosens property of REQ-HPR-007 is a table property test (AC-HPR-006), not prose.

## §D2 Stop-chain composition on Codex (proposed: single Go entry, ordered chain)

Options:

- **A. Mirror the Claude array** — render one Codex handler per chain member
  (`moai hook stop-goal --harness codex`, …).
  - For: the smallest code change.
  - Against: Codex's merge rule for several handlers returning decisions on one event is unmeasured
    (research H3). Each handler would also inherit MoAI's own render constant (10 s for every
    non-SessionEnd event today, R1.5), which is a code choice rather than a host limit. And the
    sync gate is a shell script that the `gpt` profile does not deploy (R1.8).
- **B. One Codex Stop handler that runs the ordered chain in Go** — `moai hook stop --harness codex`
  calls each member's existing Go entry in design §08 order (input validation → session/worktree
  attribution → dedup → required gates → goal → advisory), and merges decisions with a documented
  rule (cancellation first, then any `deny` wins, then goal continuation, then allow).
  - For: one merge rule that MoAI owns and can test; one budget; no dependency on host
    multi-handler semantics.
  - Against: the chain code must honour each member's budget inside one handler timeout (§D3).

**Chosen: B.** The Claude side keeps its shell registrations; effect equivalence is proven by the
goldens in AC-HPR-002, not by identical registration. The members that are already Go subcommands
(`stop-goal`, `security-turn`, `security-commit`, `codex-review-gate`, `multi-review-gate`,
`harness-observe-stop`) are called through extracted Go functions shared by both paths. The sync
gate's decision core (compile/vet per detected language, blocking opt-out, outcome record) needs a
Go entry the Codex chain can call. The Claude wrapper then becomes a shim over that entry or keeps
its script, provided AC-HPR-002 proves equal effect. Choosing between shim and parallel
implementation is run-phase M2d work.

The code stays out of `internal/hook` (SPEC-CODEX-HOOK-ADAPTER-001 REQ-7): the Codex chain runner
lives in `internal/cli` next to `hook_harness_codex.go`, or in `internal/codexadapter`.

## §D3 Stop timeout and receipts — the Codex Stop timeout is a design variable

### §D3.1 Premise (corrected after plan-audit iter-1)

The 10 s value that Codex handlers carry today is **MoAI's own render constant**, not a measured
host limit. `internal/codexwiring/codexwiring.go:58–60` calls it "the table constant for every
non-SessionEnd handler", and `moaiHandlerTimeout` (`hooks.go:33–39`) already renders a different
value for one event (SessionEnd, capped at 3 s by Codex's documented ceiling). A per-event Stop
timeout is therefore a code choice available to this SPEC. What is unknown is **the largest Stop
timeout Codex accepts and honours**. That is measured by AC-HPR-021 and stays `NOT_RUN` until the
measurement exists (research.md §R4).

Two readings from iter-1 are withdrawn. The goal loop is **not** unbounded: each mechanical
condition is bounded at 90 s (`hook_stop_goal.go:16–20`), and the Claude registration is 120 s.
What grows is the total, N conditions × 90 s. The two review gates are **not** fast: both are
registered at 900 s on Claude, and `codex_review_gate.go:83–87` bounds its RPC with
`config.DefaultCodexReviewGateTimeout` = 900 s.

### §D3.2 Two options for each member, and the variable that chooses between them

- **Per-event timeout option.** Render a Stop-specific timeout `T_stop` (a new branch in
  `moaiHandlerTimeout`, mirroring the SessionEnd branch) and run the member inside the hook.
- **Receipt option.** The member's expensive work runs outside the hook, during the turn, and
  leaves a receipt. The hook only compares the receipt with the current state.

`T_stop` must satisfy `T_stop ≤ T_codex_max`, where `T_codex_max` is the value AC-HPR-021 measures.
Until that measurement exists, `T_stop` stays at the current render constant (10 s). Only members
whose in-hook cost fits under it run in-hook. Once `T_codex_max` is measured, M2b may raise
`T_stop` and move a receipt member in-hook, provided the §D3.5 aggregate rule still holds. That
move is recorded in progress.md; it is not a new decision gate.

### §D3.3 Per-member budget and placement table (all 8 Claude Stop members)

Claude columns are measured from `settings.json.tmpl` and the named Go sources (research.md
§R1.2, §R1.5, §R1.11). Codex internal budgets are **proposals** that M2d validates by timing
measurement; they are the deadlines the chain runner sets per member, not host timeouts.

| # | Member | Claude registration timeout | Claude in-hook cost bound | Codex placement (default, `T_stop` = 10 s) | Codex internal budget | Receipt producer on Codex | Class |
|---|---|---|---|---|---|---|---|
| 1 | `moai hook stop` (turn state, telemetry prune, reflection, evidence gate, factory batch) | 5 s | handler work, no external command | in-hook | 2 s for the advisory steps; the factory-continuation step is exempt from the cut-off (see below) | — | advisory, except the factory-continuation path (`stop.go:80–81`), which is decision-bearing |
| 2 | sync-phase quality gate (compile/vet per detected language) | 60 s | compile/vet over the tree | **receipt** | 0.5 s (compare only) | the sync gate's decision core run out of hook through a `moai` entry decided in M2d; its record is the existing `.moai/state/sync-quality-gate.last` (`<head-sha> <outcome> <worktree-content-id>`), extended with the §D3.6 fields | required-gate |
| 3 | `moai hook stop-goal` | 120 s | 90 s per mechanical condition × N | **in-hook, lookup-only** | 2 s | `moai verify record` (`internal/cli/verify.go:88`), run by the working agent after it runs the condition command; the evaluator's existing snapshot source (`verify.Source.Lookup`, `internal/verify/source.go:44`) reads it | goal |
| 4 | `moai hook security-turn` | 5 s, async | observation | in-hook (MoAI's measured Codex handler whitelist is `{type, command, timeout}`, `hooks.go:16–18`, with no async key) | 0.5 s | — | advisory |
| 5 | `moai hook security-commit` | 5 s, async | observation | in-hook | 0.5 s | — | advisory |
| 6 | `moai hook codex-review-gate` | 900 s | codex review RPC ≤ 900 s (`codex_review_gate.go:87`) | **receipt** | 0.5 s (compare only) | a codex review result persisted during the turn (the `codex_audit` MCP tool or a `moai` CLI entry, decided in M2d). Invoking a codex review RPC from inside a Codex host's own Stop hook is nested Codex execution and is not measured, so it is not chosen | **fail-open-on-missing** (config-conditional) |
| 7 | `moai hook multi-review-gate` | 900 s | reads a persisted result only (`multi_review_gate.go:108–117`) | in-hook | 0.5 s | already receipt-shaped: `.moai/state/audit-multi/<session>.json`, written by the `audit_multi` MCP tool during the turn | **fail-open-on-missing** (config-conditional) |
| 8 | `moai hook harness-observe-stop` (inside a `{{ if .HookOptIn.Enabled }}` branch) | 5 s | observation | in-hook | 0.5 s | — | advisory |

Member 3 runs in-hook because comparing receipts is cheap. What changes on Codex is that a
condition whose command has no fresh receipt is **not re-executed** in the hook. That is the one
behavioural switch in the evaluator: today `Lookup` misses fall back to re-execution. On Codex the
evaluator runs in a lookup-only mode, and a miss yields the `unmeasured` class below. Model
conditions (transcript claims) are evaluated in-hook as today. They run no command.

On Claude every member keeps its current registration and in-hook behaviour. This SPEC does not
change the Claude Stop array.

**Class `fail-open-on-missing` (members 6 and 7).** Plan-audit iter-2 N1 corrected the v0.2.0
classing of the two review gates as `required-gate`. On Claude both gates allow the stop when their
result is missing, and the source states this on purpose:

- `internal/cli/multi_review_gate.go:47`: "no session id / missing state file → ALLOW (fail-open; no
  result yet)". The return is at `:79`.
- `internal/cli/multi_review_gate.go:110–111`: "The fail-open direction is load-bearing: a missing
  optional backend's evidence-of-absence must NOT trap the session."
- `internal/cli/codex_review_gate.go:59`: "codex missing → ALLOW (fail-open; can't trap the
  session)".

This SPEC does not change Claude behaviour. It gives the two gates the same class on both
harnesses. When the gate's result is present, the gate decides on it (a block stays a block). When
the result is missing, the gate allows, and on Codex the allow is never silent: the chain writes a
discard record through `RecordDiscards` (`internal/codexadapter/diagnostics.go:24`) and puts
reason text on the output naming the missing result. The goal member (3) and the sync gate (2)
stay fail-closed: a missing receipt continues the turn and never allows.

**Member 1's factory-continuation path (plan-audit iter-2 N4).** `stop.go:80–81` returns a
`decision: block` when the factory batch asks the lane to keep working. That path is exempt from
the 2 s advisory cut-off. If it is interrupted by the handler's outer timeout, the chain treats the
member as `unmeasured` and continues the turn. It never turns a continuation into an allow. The
other steps of member 1 (telemetry prune, reflection, evidence gate, `stop.go:59–79`) stay advisory
under the 2 s budget.

### §D3.4 Continuation reason classes and the declared mapping (REQ-HPR-002)

| Situation | Claude class | Codex class | Treated as equivalent? |
|---|---|---|---|
| condition ran and failed | `unmet` | `unmet` (receipt present, failing exit) | yes (identical) |
| condition passed | `met` → allow | `met` → allow (receipt present, exit 0, all fields equal) | yes (identical) |
| condition not measured on this tree state | (does not arise: Claude runs it in-hook) | `unmeasured` → continuation naming the command to run | **yes, declared mapping `unmeasured` ↔ `unmet`**: both continue the turn, and neither may allow the stop |
| required gate (sync gate) failed | `gate_failed` → block | `gate_failed` → block (receipt shows fail) | yes (identical) |
| required gate (sync gate) not measured | (does not arise: Claude runs it in-hook) | `unmeasured` → continuation naming the gate command | yes, same declared mapping as above |
| review gate (member 6 or 7), result present and blocking | block | block | yes (identical) |
| review gate (member 6 or 7), result present and passing or inconclusive | allow | allow | yes (identical) |
| review gate (member 6 or 7), **result missing** | allow, fail-open (`multi_review_gate.go:47, :79`; `codex_review_gate.go:59`) | allow, fail-open, **plus** a discard record and reason text naming the missing result | yes (identical decision). The Codex-only diagnostic is an addition to the output, not a different decision |

The AC-HPR-002 goldens encode this table. Any other pairing fails the golden, including a Codex
`allow` against a Claude continuation, a Codex `unmeasured` from the goal or sync gate that allows
the stop, and a Codex review-gate missing-result allow that writes no discard record.

### §D3.5 Aggregate budget rule for the single Codex Stop handler (option B; resolves D11)

Option B runs all members inside one Codex handler, one after another. The comparison target is:

```
Σ(internal budgets of in-hook members) + chain_overhead  ≤  T_stop  ≤  T_codex_max
```

- Receipt members contribute only their compare budget, never their out-of-hook cost.
- `chain_overhead` is a named constant fixed in M2b (input parsing, attribution, dedup, output
  write). It is declared, not implied.
- The proposed budgets sum to 2 + 0.5 + 2 + 0.5 + 0.5 + 0.5 + 0.5 + 0.5 = 7 s before overhead,
  leaving at most 3 s for `chain_overhead` under the current 10 s render constant. That is a
  statement about **declared** budgets. Whether each member's actual in-hook cost fits its budget
  is **not yet measured**, and this document does not claim that it does. Member 1 (telemetry
  prune plus `AnalyzeSessionAndLog`, `stop.go:59–73`) and member 3 (in-hook model-condition
  evaluation) are the likeliest to exceed their budgets.
- The measurement belongs to run-phase M2d: each in-hook member runs on the golden fixtures with
  its observed cost recorded (AC-HPR-016, timing leg). A member whose observed maximum exceeds its
  declared budget fails the leg. The budget is then raised within the aggregate, or the member moves
  to receipt placement.
- Operator decision Q5 (no live runs in this SPEC) means AC-HPR-021 stays `NOT_RUN`, so
  `T_codex_max` is unknown and `T_stop` stays at 10 s for this SPEC.
- A member that exceeds its internal budget at runtime is cut off. An advisory member is recorded
  failed (REQ-HPR-004). The goal member and the sync gate yield `unmeasured` and never allow. A
  `fail-open-on-missing` gate that is cut off is treated as result-missing: it allows, with the
  discard record. Member 1's factory-continuation path is not cut off (see §D3.3).

### §D3.6 Receipt fields (REQ-HPR-019)

| Field | Source | Reused primitive |
|---|---|---|
| `head` | `git rev-parse HEAD` | `verify.Key` prefix |
| `tree_digest` | porcelain-v2 + `git diff HEAD` + untracked content | `verify.Key` digest suffix |
| `config_digest` | hash of the configuration sections the check reads (e.g. quality, goal, the sync-gate blocking switch) | new — minimal |
| `command` | the exact command line run | already recorded by `moai verify record` |
| `tool_version` | version output of the tool that ran (go, linter, codex, …) | new — string |

Rules:

- Any mismatch, or an absent or partial receipt, reads as **not run** (`unmeasured`), never as
  passed. `verify.Fresh`'s TTL leg is kept only as an extra staleness bound.
- The hook compares only. It never starts a check whose budget exceeds `T_stop` (REQ-HPR-018).
- Storage reuses the existing verification-snapshot store (`internal/verify`) and the sync gate's
  existing outcome record, extended with the missing fields. There is no new store. Receipts are
  per worktree/run (design §14).

### §D3.7 What is decided now and what waits on measurement

- **Decided:** option B (one handler); the receipt method for members 2 and 6; lookup-only goal
  evaluation on Codex; the reason-class mapping; the aggregate rule.
- **Waits on AC-HPR-021 (not run in this SPEC, operator decision Q5):** the value of `T_stop` above
  10 s, and whether member 2 can later move in-hook. The probe is designed and kept opt-in; the
  follow-up live-certification card runs it. Until then `T_stop` stays at 10 s and the §D3.5 rule
  is met by the declared budgets.
- **Withdrawn:** the iter-1 claims that 10 s was a measured host budget and that the longest check
  is unbounded.

## §D4 Obligation registry (proposed)

A single data file plus a Go loader, not a new service:

- Location: `internal/template/obligations.yaml` (embedded), loader and coverage check in the
  existing `internal/template` package unless plan audit prefers a new package. `<registry-pkg>`
  in acceptance.md resolves to this decision.
- Row schema: `id`, `required` (bool), `claude_path`, `codex_path` (path or `UNSUPPORTED:<evidence-ref>`
  or `blocked:<reason>`), `check` (Go test name), `ac` (acceptance id).
- Coverage check (AC-HPR-018): each `check` must resolve to a test function that exists in the
  tree (static lookup), and each path must resolve to a rendered artifact or a registered
  subcommand.
- This is the minimal core of design §07's `features.yaml`. The capability-detection fields
  (`requires`, `implementations`, `on_unavailable`) are not built here. M1 extends the same file.

Scope boundary: this SPEC seeds the registry with the M2 obligations (Stop chain members,
decision preservation per event, goal continuation/cancellation/budget, receipt validity). It also
records two rows that keep the aggregate honest: the Claude user-interrupt cancellation source as
`UNSUPPORTED` (§D6), and the non-Stop multi-handler chains (spec.md §F) as `unverified`.

**Decided (Q1): AC-POL-01 closes over the whole obligation catalog.** M1 standing-policy
obligations are registered too, marked `blocked:M1`, and the aggregate reports them honestly as
FAIL until M1 lands. AC-POL-01 therefore does not read PASS from this SPEC alone.

## §D5 `needs_input` on Codex (DECIDED: fail-closed deny, surfaced visibly)

The current adapter drops PreToolUse `ask` to `{}` (R1.4). Under REQ-HPR-007 that is permitted
only if measurement shows Codex prompts under every supported approval policy. Proposed default:
translate `needs_input` to `deny` with a reason that names the approval needed and the MoAI command
to grant it. This loses the host's native approval UI on Codex but cannot become an allow.

**Decided (Q2): fail-closed deny, and the deny is surfaced visibly.** The deny reason reaches the
model through PreToolUse's working reason channel. The same event is also written as a discard
record through the adapter's existing no-silence sink (`RecordDiscards`), so an operator can see
that a `needs_input` was converted rather than learn it from a stalled turn. A silent allow and a
silent deny are both non-compliant. The pause-and-route alternative depends on the unmeasured
supervised-session path (design §09) and stays a follow-up.

## §D6 Cancellation in goal state (DECIDED: new `cancelled` status — operator, 2026-09-23)

A user interrupt is a different event from an explicit clear and carries a different audit
meaning, so this SPEC adds a new status `cancelled` to the goal status vocabulary. It is written
by the Codex Interrupt path (§D7) and takes precedence over an unmet goal (REQ-HPR-015). `cleared`
is not reused (operator decision Q3).

The second producer, `moai goal clear`, does not write a status at all. It removes the goal state
(`internal/goal/state.go:120–127`), so the next Stop finds no goal and does not block, and nothing
reads `satisfied`. `StatusCleared` is defined (`schema.go:73`) and read once
(`evaluate.go:294`), but no non-test code writes it (research.md §R1.10). A third site also removes
goal state: `internal/hook/stop_failure.go:111` calls `goal.ClearGoal` on an unrecoverable
StopFailure. That is a failure path, not a user cancellation, and it does not write `cancelled`.

Every non-test reader of goal status must handle `cancelled` explicitly, and an unrecognised
status must surface a diagnostic rather than be treated silently. Measured readers and writers
(research.md §R1.10; grep scope `internal/goal internal/cli internal/hook`):

| Reader | What it does with the status | Required handling of `cancelled` |
|---|---|---|
| `internal/goal/schema.go:65–78` | the status set | add the constant |
| `internal/goal/evaluate.go:294` | early return (no block) for `cleared` / `satisfied` | add `cancelled` to the early-return set |
| `internal/goal/evaluate.go:306/325/340/404/435` | writers of ceiling-exit / unsatisfiable / satisfied | must not overwrite `cancelled` |
| `internal/cli/launcher_blockcap_infinite.go:68` | acts only on `armed` | none (non-armed already excluded) — asserted by test |
| `internal/cli/handoff.go:85` | embeds only an `armed` goal | none — asserted by test |
| `internal/cli/goal.go:1282, 1331` | `goal status` output (list and single) | renders `cancelled` verbatim — asserted by test |
| `internal/goal/dashboard.go:141` | dashboard status string | renders `cancelled` verbatim — asserted by test |
| `internal/cli/hook_stop_goal.go` emission condition | emits only on block or exit transitions | a `cancelled` goal emits nothing and never blocks |
| `internal/cli/goal.go:1313` | `moai goal clear` → `goal.ClearGoal` (removes state) | none — asserted by test |
| `internal/hook/session_start_compact.go:88` | reader; acts only on `armed` | none; `internal/hook` stays untouched (REQ-7, Q6) — asserted by test |
| `internal/hook/handoff_inject.go:185` | writer of `armed` for an injected goal | none; `internal/hook` stays untouched — asserted by test |
| `internal/hook/stop_failure.go:111` | `goal.ClearGoal` on an unrecoverable StopFailure (failure path, not cancellation) | none; `internal/hook` stays untouched — asserted by test |

`TestGoalStatusConsumersHandleCancelled` (AC-HPR-013) enumerates this table. Its scan scope is the
package list `internal/goal`, `internal/cli`, and `internal/hook`, non-test files only. It fails
when a new site that reads or writes `goal.Status`, or calls `goal.ClearGoal`, appears in those
packages and the table does not list it. An unrecognised status
value yields a stderr diagnostic from the evaluator and no block; it is never mapped to
`satisfied`.

There is no Claude-side interrupt producer, and this SPEC does not invent one. The Claude Stop
hook does not fire on user interrupt, and the only Claude interrupt signal, `IsInterrupt`
(`internal/hook/types.go:249`), rides PostToolUseFailure, which spec.md §F excludes. The Claude
source is recorded in the obligation registry as `UNSUPPORTED` with that evidence (§D4). On
Claude the cancellation producer is `moai goal clear`.

The status addition changes a persisted schema, so it is milestone M2a in plan.md.

## §D7 Interrupt without touching `internal/hook`

`codexadapter.CodexEventInterrupt` already exists outside `internal/hook` (SPEC-CODEX-EVENT-COVERAGE-001
REQ-CEV-002). The Interrupt row gets a dispatcher arg handled entirely in `internal/cli` (a new
`moai hook interrupt` subcommand scoped to `--harness codex`), which writes the cancellation record
through the goal package. No Claude-side registration is added. With the arg in place the row is adapted, so the adapted count becomes 12 and `RenderHooks` installs an Interrupt handler. That reverses SPEC-CODEX-EVENT-COVERAGE-001 REQ-CEV-004, and its test is amended in M2e.

**Decided (Q6): SPEC-CODEX-HOOK-ADAPTER-001 REQ-7 is kept.** Nothing under `internal/hook` is
modified. Shared Stop-member functions are extracted into `internal/cli` or
`internal/codexadapter`, not into `internal/hook`.

## §D8 Live verification harness

- One env switch `MOAI_PARITY_LIVE=1` gates the live ACs; the axis is declared in
  `codex_live_axis_declaration_test.go` like the existing switches.
- Codex runs under `t.TempDir()` as `CODEX_HOME`; Claude runs via `claude -p` in a scratch project
  under the OS temp dir, never the development checkout.
- The verdict is read from `go test -json` with rule P (acceptance.md §B). A skipped live test is
  `NOT_RUN` by construction.
- **No live test may return normally without achieving its trigger** (acceptance.md §B rule 8).
  When the trigger is not achieved — no compaction, no approval request, no SIGINT delivered, a
  policy that could not be set — the test calls `t.Skipf` carrying the attempted command and the
  observed output, so rule P reads `NOT_RUN`. The test also writes the same verdict record the
  aggregate reads. The aggregate (AC-HPR-019) takes the **weaker** of the go-test action and the
  record, so a `pass` action paired with a `NOT_RUN` record is `NOT_RUN`.
- Every live Claude run asserts its working directory is under the OS temp dir and outside the
  repository root before it starts.
- **Decided (Q5): no live run is executed in this SPEC.** The live axis, the live tests, and their
  `t.Skipf` discipline are designed and built but stay opt-in and unexecuted. The live legs
  (AC-HPR-004, 007, the live legs of 008–012, 020, 021) therefore record `NOT_RUN`, which is not
  PASS. The SPEC closes as **partial (live-uncertified)**. Live certification is deferred to a
  separate follow-up card, which also sets the model-turn budget and credential source.
