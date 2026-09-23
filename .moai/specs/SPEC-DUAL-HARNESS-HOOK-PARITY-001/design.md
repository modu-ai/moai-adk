# Design — SPEC-DUAL-HARNESS-HOOK-PARITY-001

This file records the design decisions the run phase implements. Each decision names the
alternatives considered and the measured fact it rests on (research.md §R1). Decisions marked
**proposed** are confirmed at plan audit / Implementation Kickoff; plan.md lists the ones that need
an operator answer.

## §D1 Normalized decision model (proposed)

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
    (research H3). Each handler would also get its own 10-second budget. And the sync gate is a
    shell script that the `gpt` profile does not deploy (R1.8).
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

## §D3 Receipt method for long-running checks — ADOPTED

Design §08 proposes that checks which cannot finish inside a synchronous hook run earlier and leave
a receipt, which the Stop hook compares. **Adopted**, because the measured budgets make the
alternative infeasible: Codex handlers are registered at 10 seconds (R1.5), while the goal
evaluator allows 90 seconds per mechanical condition and the sync gate runs compile/vet over the
whole tree.

Receipt fields (all five required; REQ-HPR-019):

| Field | Source | Reused primitive |
|---|---|---|
| `head` | `git rev-parse HEAD` | `verify.Key` prefix |
| `tree_digest` | porcelain-v2 + `git diff HEAD` + untracked content | `verify.Key` digest suffix |
| `config_digest` | hash of the configuration sections the check reads (e.g. quality, goal, sync-gate env) | new — minimal |
| `command` | the exact command line run | new — string |
| `tool_version` | version output of the tool that ran (go, linter, …) | new — string |

Rules:

- Any mismatch or an absent/partial receipt reads as **not run** (REQ-HPR-019). There is no TTL
  grace: `verify.Fresh`'s TTL leg is reused only as an additional staleness bound.
- The hook does comparison only. It never starts a check whose budget exceeds its registered
  timeout (REQ-HPR-018). Where no receipt exists, the chain emits `needs_input`/continuation
  naming the command to run, rather than running it.
- Storage reuses the existing verification-snapshot store (`internal/verify`) instead of a new
  file format, extending the record with the three new fields. Receipts are per worktree/run
  (design §14).

Rejected alternative: raise the Codex handler timeout to cover the longest check. Rejected because
the longest check is unbounded (project-defined goal commands), and a long synchronous Stop is the
hazard design §20 names ("장시간 검증의 훅 timeout").

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
decision preservation per event, goal continuation/cancellation/budget, receipt validity). Whether
AC-POL-01 must also close over M1 standing-policy obligations is plan.md clarification Q1.

## §D5 `needs_input` on Codex (proposed: fail-closed deny)

The current adapter drops PreToolUse `ask` to `{}` (R1.4). Under REQ-HPR-007 that is permitted
only if measurement shows Codex prompts under every supported approval policy. Proposed default:
translate `needs_input` to `deny` with a reason that names the approval needed and the MoAI command
to grant it. This loses the host's native approval UI on Codex but cannot become an allow. The
alternative — pause and route to MoAI's explicit user-input path — depends on the supervised
session path (design §09), which is not measured; it stays a follow-up. This is plan.md clarification Q2.

## §D6 Cancellation in goal state (proposed: new status)

The existing `cleared` status means an explicit `moai goal clear`. A user interrupt is a different
event with a different audit meaning. Proposed: add `cancelled` to the goal status vocabulary,
written by the Interrupt path (Codex) and the user-interrupt path (Claude), with precedence over an
unmet goal (REQ-HPR-015). This changes a persisted schema (`internal/goal/schema.go`), so it is
listed first in plan.md. Clarification Q3 asks whether reusing `cleared` is preferred.

## §D7 Interrupt without touching `internal/hook`

`codexadapter.CodexEventInterrupt` already exists outside `internal/hook` (SPEC-CODEX-EVENT-COVERAGE-001
REQ-CEV-002). The Interrupt row gets a dispatcher arg handled entirely in `internal/cli` (a new
`moai hook interrupt` subcommand scoped to `--harness codex`), which writes the cancellation record
through the goal package. No Claude-side registration is added.

## §D8 Live verification harness

- One env switch `MOAI_PARITY_LIVE=1` gates the live ACs; the axis is declared in
  `codex_live_axis_declaration_test.go` like the existing switches.
- Codex runs under `t.TempDir()` as `CODEX_HOME`; Claude runs via `claude -p` in a scratch project
  under the OS temp dir, never the development checkout.
- The verdict is read from `go test -json` with rule P (acceptance.md §B). A skipped live test is
  `NOT_RUN` by construction.
- Model-turn cost is bounded per run and declared up front (clarification Q5).
