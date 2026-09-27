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
| 1 | `moai hook stop` (turn state, telemetry prune, reflection, evidence gate, factory batch) | 5 s | handler work, no external command | in-hook | 2 s for the advisory steps, plus the factory step's own 0.2 s bound (`factoryHookInspectionDeadline`, `internal/hook/factory_messages.go:20`, applied at `:112` and `:122`); the factory step is exempt from the 2 s cut-off (see below) | — | advisory, except the factory-continuation path (`stop.go:80–81`), which is decision-bearing |
| 2 | sync-phase quality gate (compile/vet per detected language) | 60 s | compile/vet over the tree | **receipt, behind the in-hook self-gate** (see "Self-gates before receipts" below) | 0.5 s (self-gate + compare) | the sync gate's decision core run out of hook through a `moai` entry decided in M2d; its record is the existing `.moai/state/sync-quality-gate.last` (`<head-sha> <outcome> <worktree-content-id>`), extended with the §D3.6 fields | required-gate |
| 3 | `moai hook stop-goal` | 120 s | 90 s per mechanical condition × N | **in-hook, lookup-only** | 2 s | `moai verify record` (`internal/cli/verify.go:88`), run by the working agent after it runs the condition command; the evaluator's existing snapshot source (`verify.Source.Lookup`, `internal/verify/source.go:44`) reads it | goal |
| 4 | `moai hook security-turn` | 5 s, async | observation | in-hook (MoAI's measured Codex handler whitelist is `{type, command, timeout}`, `hooks.go:16–18`, with no async key) | 0.5 s | — | advisory |
| 5 | `moai hook security-commit` | 5 s, async | observation | in-hook | 0.5 s | — | advisory |
| 6 | `moai hook codex-review-gate` | 900 s | codex review RPC ≤ 900 s (`codex_review_gate.go:87–94`) | **receipt, behind the in-hook self-gates** (see "Member 6 on the receipt method" below) | 0.5 s (self-gates + compare) | the out-of-hook codex review runner: a `moai` CLI entry, named in M2d, that makes the same reviewer call as the Claude gate (`codex_review_gate.go:87–109`) and writes a receipt. The working agent runs it after its last edit of the turn, and again whenever the Stop chain's continuation reason names it. Running the review inside the Codex Stop handler is not chosen: it can take up to 900 s, far above `T_stop` (REQ-HPR-018), and it would be nested Codex execution | **required-gate** (config-conditional; fail-open only when the codex binary is missing) |
| 7 | `moai hook multi-review-gate` | 900 s | reads a persisted result only (`multi_review_gate.go:108–117`) | in-hook | 0.5 s | already receipt-shaped: `.moai/state/audit-multi/<session>.json`, written by the `audit_multi` MCP tool during the turn | **fail-open-on-missing** (config-conditional) |
| 8 | `moai hook harness-observe-stop` (inside a `{{ if .HookOptIn.Enabled }}` branch) | 5 s | observation | in-hook | 0.5 s | — | advisory |

Member 3 runs in-hook because comparing receipts is cheap. What changes on Codex is that a
condition whose command has no fresh receipt is **not re-executed** in the hook. That is the one
behavioural switch in the evaluator: today `Lookup` misses fall back to re-execution. On Codex the
evaluator runs in a lookup-only mode, and a miss yields the `unmeasured` class below. Model
conditions (transcript claims) are evaluated in-hook as today. They run no command.

On Claude every member keeps its current registration and in-hook behaviour. This SPEC does not
change the Claude Stop array.

**Class `fail-open-on-missing` (member 7 only).** Plan-audit iter-2 N1 put both review gates in
this class. Plan-audit iter-3 R1 showed that this holds for member 7 and is false for member 6, so
the class now covers member 7 only. On Claude the multi-review gate reads a persisted result and
allows the stop when that result is missing, and the source says so on purpose:

- `internal/cli/multi_review_gate.go:47`: "no session id / missing state file → ALLOW (fail-open; no
  result yet)". The return is at `:79`.
- `internal/cli/multi_review_gate.go:110–111`: "The fail-open direction is load-bearing: a missing
  optional backend's evidence-of-absence must NOT trap the session."

Claude already reads a persisted result here, so "missing" means the same on both harnesses. When
the result is present, the gate decides on it (a block stays a block). When it is missing, the gate
allows, and on Codex the allow is never silent: the chain writes a discard record through
`RecordDiscards` (`internal/codexadapter/diagnostics.go:24`) and puts reason text on the output
naming the missing result.

**Member 6 on the receipt method (plan-audit iter-3 R1; operator decision R1, 09-23: match
Claude).** On Claude the codex review gate does not read a stored result. With the gate enabled,
a reviewable change, and the codex binary installed, it runs the review itself and blocks on FAIL
(decision order `internal/cli/codex_review_gate.go:56–61`; reviewer lookup `:78–81`; review call
`:87–94`; block `:103–107`). It fails open in exactly two places: the codex binary is absent
(`:79–80`), and the review call returns an error, which reaches the caller as an allow
(`:95–101`). A pass or inconclusive verdict allows (`:109`).

The review can take up to 900 s, so on Codex it runs out of hook and the Stop chain reads its
receipt. The receipt carries the §D3.6 fields — `head` and `tree_digest` (HEAD plus the digest of
the uncommitted diff, from `verify.Key`), `config_digest`, `command`, `tool_version`
(`codex --version`) — plus the reviewer's verdict: `pass`, `inconclusive`, or `fail`. The receipt
producer is the out-of-hook codex review runner named in the §D3.3 row. It writes `inconclusive`,
with the error text, when the review call errors, mirroring `:95–101`. The working agent runs it
after its last edit of the turn. When the Stop chain continues the turn because the receipt is
missing or stale, its reason names the runner command, so the next turn produces the receipt.

The Codex member 6 evaluates in the same order as Claude (`:56–61`):

1. gate disabled → allow, on both harnesses;
2. `stop_hook_active` set → allow, on both harnesses (Claude `:71–72`). Step 2 is evaluated
   before any receipt read, so it also precedes step 7. **Step 2 is the only bound on the step-7
   continuation that both harnesses share:** on Claude member 6 carries no turn counter, ceiling, or
   retry limit (unlike the goal member, which has its turn ceiling), so a continued turn gets one
   more Stop and then allows. On Codex the step-7 continuation also carries the Codex-only
   consecutive-`unmeasured` cap (§D3.8, operator 09-23). The field's presence in the Codex Stop payload is measured: the captured payload
   `.moai/specs/SPEC-CODEX-HOOK-ADAPTER-001/testdata/hook-payloads/Stop.json:9` carries
   `"stop_hook_active": false`, and `TestGoldenStopCarriesStopHookActive`
   (`internal/codexadapter/golden_test.go:118–128`) pins the field. What is **not** measured is
   whether Codex sets the value to `true` on the Stop that follows a hook-driven continuation; that
   needs a live run, which stays `NOT_RUN` under operator decision Q5. If Codex never sets it,
   step 2 does not bound the step-7 continuation on Codex — recorded as a live-unmeasured host fact,
   not assumed away. The Codex-only consecutive-`unmeasured` cap (§D3.8, operator 09-23) is
   therefore the Codex bound that holds whatever Codex does with the flag. AC-HPR-002 carries a
   `stop_hook_active: true` golden and a skip-step-2 mutation, plus the §D3.8 cap goldens;
3. no reviewable change (the same `reviewGateChangeDetector` predicate, `:48`, `:74–76`) → allow,
   on both harnesses;
4. codex binary missing (the same `codexLookPath` lookup, `internal/cli/mcp_codex.go:457`) → allow
   on both harnesses; on Codex the allow also writes a discard record and reason text naming the
   missing reviewer;
5. codex installed and a fresh receipt whose verdict is `pass` or `inconclusive` → allow;
6. codex installed and a fresh receipt whose verdict is `fail` → block;
7. codex installed and the receipt missing, partial, or stale (any §D3.6 field differs, for example
   HEAD moved) → continuation — a Stop `block` decision — with reason class `unmeasured`, naming
   the runner command. This does not allow (fail-closed) until the §D3.8 cap is reached; at the
   cap the stop is allowed and the gate is recorded `unverified`, never passed.

Step 7 is the Codex counterpart of Claude running the review in-hook. §D3.4 declares it as the same
mapping used for the goal member. The claim is narrower than "no allow without a verdict": Claude
does allow without a review verdict, at step 2 (`stop_hook_active`, `:71–72`) and when the review
call errors (`:95–101`). What holds is this: at `stop_hook_active: false`, with the gate enabled, a
reviewable change, and codex installed, neither harness allows the stop before the review call has
completed for the current tree, below the Codex-only §D3.8 cap — on Claude the call runs in-hook, and on Codex the runner's receipt
records its outcome (`inconclusive` on a call error, mirroring `:95–101`).

**Self-gates before receipts (plan-audit iter-3 R2).** A member that does not apply on Claude must
not apply on Codex. So the Codex chain evaluates each gate's own applicability predicates in-hook,
before any receipt compare, and allows when they do not hold:

- **Sync gate (member 2).** The Claude script's trigger predicate is the last commit's subject: it
  reads `git log -1 --format='%s'` and continues only when the subject matches
  `*"docs("*"): sync-phase"*`, `*"chore("*"): sync-phase"*`, `*"docs: sync"*`, or
  `*"chore: sync"*`. Any other subject exits 0 with empty stdout, which is an allow
  (`internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh:207–218`; the header
  comment at `:4` summarises it as "HEAD is a sync-phase commit"). Two further early allows follow:
  no recognised language marker (`:227–231`) and no code-file delta in `HEAD~1..HEAD`
  (`:233–253`). The Codex member evaluates the same predicates, in the same order, before reading
  the receipt. Only when all of them hold does it require the receipt; a missing, stale, or failing
  receipt then continues or blocks exactly as the rows of §D3.4 say, subject to the
  `stop_hook_active` rule below. The blocking opt-out (`MOAI_SYNC_GATE_BLOCKING`, `:386–395`) is
  read on both harnesses (REQ-HPR-005).

  **`stop_hook_active` on the sync gate (plan-audit iter-4 S1).** A Codex receipt is the analogue
  of Claude's stored gate record, not of a run that executes the checks: the gate command already
  ran out of hook. Claude never re-delivers a stored block on a turn whose stdin carries
  `"stop_hook_active": true` — it exits 0 with no state change
  (`internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh:484–486`; header
  `:53–56`), and the next turn without the flag re-delivers it. So a fresh `fail` receipt with
  `stop_hook_active: true` → **allow** on Codex, with no receipt or state change; the next Stop
  without the flag blocks again. The same header states that "the flag never suppresses the
  output of a run that executes the checks" (`:55–56`). The Codex analogue of that run is the
  missing or stale receipt, so that case keeps its `unmeasured` continuation whatever the flag says.
  On Claude that case does not arise (the check runs in-hook), so the flag gives the Codex path no
  bound there; the Codex-only consecutive-`unmeasured` cap (§D3.8) is that bound.
- **Codex review gate (member 6).** Steps 1–4 above.
- **Goal (member 3).** "No goal armed" allows before any receipt lookup, on both harnesses (already
  a golden).

The goal member (3) and the sync gate (2) stay fail-closed once their self-gates hold: a missing
receipt continues the turn and never allows a pass. Member 6 behaves the same way once its
self-gates hold. For the sync gate and member 6 the continuation is bounded on Codex by the
consecutive-`unmeasured` cap (§D3.8): at the cap the stop is allowed and the gate is recorded
`unverified`, which no verdict reads as PASS. The goal member keeps its own turn ceiling and is not
under the cap.

**Member 1's factory-continuation path (plan-audit iter-2 N4; iter-3 A2).** `stop.go:80–81` returns
a `decision: block` when the factory batch asks the lane to keep working. That step is bounded by
its own deadline, `factoryHookInspectionDeadline = 200 * time.Millisecond`
(`internal/hook/factory_messages.go:20`, applied at `:112` and `:122`), and is exempt from the 2 s
advisory cut-off. Its 0.2 s is counted in the §D3.5 aggregate. If the chain runner's own deadline
(`T_stop` minus `chain_overhead`) expires before the step returns, the chain treats the member as
`unmeasured` and continues the turn. It never turns a continuation into an allow. The other steps
of member 1 (telemetry prune, reflection, evidence gate, `stop.go:59–79`) stay advisory under the
2 s budget.

### §D3.4 Continuation reason classes and the declared mapping (REQ-HPR-002)

| Situation | Claude class | Codex class | Treated as equivalent? |
|---|---|---|---|
| condition ran and failed | `unmet` | `unmet` (receipt present, failing exit) | yes (identical) |
| condition passed | `met` → allow | `met` → allow (receipt present, exit 0, all fields equal) | yes (identical) |
| condition not measured on this tree state | (does not arise: Claude runs it in-hook) | `unmeasured` → continuation naming the command to run | **yes, declared mapping `unmeasured` ↔ `unmet`**: both continue the turn, and neither may allow the stop below the Codex-only §D3.8 cap |
| required gate (sync gate) failed, `stop_hook_active: false` | `gate_failed` → block (a fresh run, or the stored block re-delivered) | `gate_failed` → block (receipt shows fail) | yes (identical) |
| required gate (sync gate) failed and already recorded, `stop_hook_active: true` | allow; the stored block is not re-delivered on that turn and no state changes (`sync-phase-quality-gate.sh:484–486`, header `:53–56`) | allow (fresh `fail` receipt); the receipt is not changed, and the next Stop without the flag blocks | yes (identical) |
| required gate (sync gate) not measured, self-gate holds (HEAD is a sync-phase commit with a code delta) | (does not arise: Claude runs it in-hook) | `unmeasured` → continuation naming the gate command | yes, same declared mapping as above |
| sync gate, self-gate does not hold (HEAD is not a sync-phase commit, or no language marker, or no code delta) | allow, silent (`sync-phase-quality-gate.sh:207–218`, `:227–231`, `:233–253`) | allow; the receipt is not read | yes (identical) |
| codex review gate (member 6), gate disabled, `stop_hook_active`, or no reviewable change | allow (`codex_review_gate.go:68–76`) | allow; the receipt is not read | yes (identical) |
| codex review gate (member 6), **codex binary missing** | allow, fail-open (`codex_review_gate.go:78–80`) | allow, fail-open, **plus** a discard record and reason text naming the missing reviewer | yes (identical decision). The Codex-only diagnostic is an addition to the output, not a different decision |
| codex review gate (member 6), codex installed, review FAIL | block (`:103–107`) | block (fresh receipt, verdict `fail`) | yes (identical) |
| codex review gate (member 6), codex installed, review pass, inconclusive, or call error | allow (`:95–101`, `:109`) | allow (fresh receipt, verdict `pass` or `inconclusive`) | yes (identical) |
| codex review gate (member 6), codex installed, **receipt missing or stale** | (does not arise: Claude runs the review in-hook) | `unmeasured` → continuation naming the review runner command | **yes, declared mapping `unmeasured` ↔ review ran in-hook**: at `stop_hook_active: false`, neither harness allows the stop before the review call has completed for the current tree, below the Codex-only §D3.8 cap (Claude still allows without a verdict at step 2 and on a call error, `:95–101`; the Codex runner records that error as `inconclusive`) |
| codex review gate (member 6), codex installed, `stop_hook_active: true` (receipt missing, stale, `fail`, or `pass`) | allow (step 2, `codex_review_gate.go:71–72`, before the reviewer lookup and the review call) | allow (step 2, before any receipt read) | yes (identical). Step 2 is the only bound on the step-7 continuation shared by both harnesses (§D3.3); Codex adds the §D3.8 cap |
| sync gate (self-gate holds) or member 6 (codex installed), receipt missing or stale, **Nth consecutive `unmeasured` continuation** for the same gate, HEAD, and working-tree digest (N from §D3.8) | (does not arise: Claude runs the check in-hook) | allow, **plus** a discard record and reason text naming the gate as `unverified` and the command that was never run | **no — declared Codex-only parity deviation (§D3.8).** Not a PASS: the gate reads `unverified` (NOT_RUN class) in the verdict and the registry. These cap goldens are Codex-only and are excluded from the Claude/Codex equality comparison |
| multi review gate (member 7), result present and blocking | block | block | yes (identical) |
| multi review gate (member 7), result present and passing | allow | allow | yes (identical) |
| multi review gate (member 7), **result missing** | allow, fail-open (`multi_review_gate.go:47, :79`) | allow, fail-open, **plus** a discard record and reason text naming the missing result | yes (identical decision). The Codex-only diagnostic is an addition to the output, not a different decision |

The AC-HPR-002 goldens encode this table. Any other pairing fails the golden, including a Codex
`allow` against a Claude continuation, a Codex `unmeasured` from the goal that allows the stop, a
Codex `unmeasured` from the sync gate or member 6 that allows the stop below the §D3.8 cap, a Codex
sync gate that requires a receipt when HEAD is not a sync-phase commit, a member-6 allow below the
§D3.8 cap while codex is installed and the receipt is missing or stale, a member-6
continuation or block at `stop_hook_active: true`, a Codex sync gate that blocks on a fresh `fail`
receipt at `stop_hook_active: true`, and a missing-result or missing-reviewer allow that writes no
discard record.

### §D3.5 Aggregate budget rule for the single Codex Stop handler (option B; resolves D11)

Option B runs all members inside one Codex handler, one after another. The comparison target is:

```
Σ(internal budgets of in-hook members) + chain_overhead  ≤  T_stop  ≤  T_codex_max
```

- Receipt members contribute only their compare budget, never their out-of-hook cost.
- `chain_overhead` is a named constant fixed in M2b (input parsing, attribution, dedup, output
  write). It is declared, not implied.
- The proposed budgets sum to (2 + 0.2) + 0.5 + 2 + 0.5 + 0.5 + 0.5 + 0.5 + 0.5 = 7.2 s before
  overhead, leaving at most 2.8 s for `chain_overhead` under the current 10 s render constant. The
  0.2 s is member 1's factory step, bounded by `factoryHookInspectionDeadline`
  (`internal/hook/factory_messages.go:20`); the compare budgets of members 2 and 6 now include
  their in-hook self-gates (§D3.3). That is a
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
  failed (REQ-HPR-004). The goal member, the sync gate, and member 6 yield `unmeasured` and never
  allow — for member 6 this holds once its self-gates have held, since a cut-off before the codex
  lookup completes leaves no evidence that the reviewer is absent. Member 7, the one
  `fail-open-on-missing` gate, is treated as result-missing when cut off: it allows, with the
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

### §D3.8 Codex-only consecutive-`unmeasured` cap (DECIDED: operator, 2026-09-23)

**The gap.** Two Codex continuations have no proven bound. (1) The sync gate on a sync-phase
commit with a missing or stale receipt keeps its `unmeasured` continuation whatever
`stop_hook_active` says (§D3.3, "`stop_hook_active` on the sync gate"), because a missing receipt is
the Codex analogue of Claude's in-hook run, which the flag never suppresses. (2) Member 6's step-7
continuation has step 2 (`stop_hook_active`) as its only other bound, and whether Codex sets that
flag to `true` on a continued turn is live-unmeasured (`NOT_RUN` under Q5). If the working agent
never produces the receipt, either case can continue the turn without end.

**The rule.** On Codex, for the sync gate (member 2, self-gate holding) and member 6 (codex
installed), the Stop chain counts consecutive `unmeasured` continuations per gate, keyed by the
gate id, HEAD, and working-tree digest (the §D3.6 `head` and `tree_digest` fields, from
`verify.Key`):

- continuations 1 to N−1 for the same key → continue, exactly as §D3.4 says;
- the Nth → **allow** the stop, write a discard record through the existing sink (`RecordDiscards`,
  `internal/codexadapter/diagnostics.go:24`), and put reason text on the output naming the gate as
  `unverified` and the command that was not run;
- the count stays at N until it resets, so each further Stop on the same key also allows and also
  writes an `unverified` record — the gate is never silently dropped;
- the count **resets to 0** when a fresh receipt for the current key is read (a receipt was
  produced), or when HEAD or the working-tree digest differs from the stored key.
- "consecutive" means the count accumulated per gate and key within the session: only the two resets above clear it, so an intervening Stop that is not `unmeasured` (a step-2 allow, a self-gate that no longer holds, a different user prompt) leaves it unchanged.

`unverified` is NOT_RUN-class. The verdict record and the obligation registry read a capped gate as
`unverified`, never as PASS (REQ-HPR-022, REQ-HPR-023); AC-HPR-019 injects it. The cap bounds the
loop; it does not certify the gate.

**N.** A declared constant, finalized in run-phase M2d. **Proposed default: 3** — a proposal, not a
measurement; M2d records the final value and its reason in progress.md §E.2.

**State location.** Session-scoped state under `.moai/state/`, following the existing
per-session-file pattern `.moai/state/<name>/<session-id>.json` used by the goal state
(`internal/statusline/goal_armed.go:36`, `.moai/state/goal/<session-id>.json`) and the multi review
gate's result (`internal/cli/multi_review_gate.go:117`, `.moai/state/audit-multi/<session-id>.json`).
The counter lives at `.moai/state/codex-stop-cap/<session-id>.json` (directory name final in M2d),
one entry per gate holding the key and the count. Only the Codex Stop chain writes it. The gate's
receipt producer (the out-of-hook codex review runner and the sync gate's out-of-hook decision core)
never writes it; a produced receipt resets the counter only by being read at the next Stop. It is
not a receipt and never enters the receipt store (§D3.6).

**Declared parity deviation.** The cap is Codex-only and has no Claude counterpart. On Claude the
situation it bounds does not arise: Claude runs the sync gate and the codex review in-hook, so there
is no missing-receipt continuation to repeat, and a Claude continuation of member 6 ends at step 2
because Claude sets `stop_hook_active` on the Stop that follows a hook-driven continuation
(`internal/cli/codex_review_gate.go:71–72`; the sync-gate header `sync-phase-quality-gate.sh:53–56`
reads the same flag). This is a declared, justified deviation (§D3.4 row), not a silent one: its
goldens are Codex-only, excluded from the Claude/Codex equality comparison of REQ-HPR-002, and its
allow is always paired with an `unverified` record.

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
