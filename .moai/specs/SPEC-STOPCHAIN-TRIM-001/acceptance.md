# acceptance.md — SPEC-STOPCHAIN-TRIM-001

> Verification layer. Each AC is a binary-testable Given-When-Then. GEARS obligations live in `spec.md` (REQ-STOPCHAIN-TRIM-NNN); this file does NOT restate them as requirements.

## §D. AC Matrix

| AC ID | REQ | Subject | Severity | Traceability |
|-------|-----|---------|----------|--------------|
| AC-STOPCHAIN-TRIM-001 | REQ-001 | goal-absent session → no stop-goal exec | MUST | plan M2 |
| AC-STOPCHAIN-TRIM-002 | REQ-001 | non-sync HEAD → no vet/build from sync-gate | MUST | plan M2 |
| AC-STOPCHAIN-TRIM-003 | REQ-002 | N-edit TDD cycle → 0 per-edit pre/post spawns | MUST | plan M3 |
| AC-STOPCHAIN-TRIM-004 | REQ-004 | fully-autonomous → sync-gate advisory only | MUST | plan M4 |
| AC-STOPCHAIN-TRIM-005 | REQ-005 | automatic → IsGitCommit gate OFF | MUST | plan M4 |
| AC-STOPCHAIN-TRIM-006 | REQ-005 + REQ-007 | deny/ask still bind at every tier | MUST | plan M4 |
| AC-STOPCHAIN-TRIM-006b | REQ-006 | fully-autonomous → lifecycle hooks dormant | MUST | plan M4 |
| AC-STOPCHAIN-TRIM-007 | REQ-003 + REQ-004/005/006 | unset token = semi-auto (backward compat) | MUST | plan M1 |

### §D.1 Severity model

All eight ACs are MUST-pass. The deny/ask invariance (AC-006) is safety-critical: a regression there would let an unattended `fully-autonomous` session push to main / deploy / wipe a directory — exactly the outcome the mode token MUST NOT enable. AC-006b is the symmetric P0 safety guard for the subagent lifecycle surface: a lifecycle hook accidentally left active at `fully-autonomous` would block/reject/translate-to-AskUserQuestion and break the unattended autonomous loop the tier exists to enable. The other six are P0 latency / behavior ACs.

### §D.2 AC definitions (Given-When-Then)

#### AC-STOPCHAIN-TRIM-001 — Goal-absent session does not exec moai binary (A10)

**Given** a session with NO armed goal (no `.moai/state/goal/<session-id>.json` file exists for the current session ID), AND a `handle-stop-goal.sh` wrapper installed as a Stop hook.

**When** the Stop event fires at turn-end for this session.

**Then** the wrapper exits 0 at the shell layer WITHOUT invoking the `moai hook stop-goal` subprocess, AND no `moai` binary cold-start is paid for this turn-end.

**Test shape:** shell fixture — install a counting wrapper around the `moai` binary (or use `strace`/`dtrace` to count `execve` of the `moai` binary), ensure no goal state file exists for the test session, fire the Stop event, assert the `moai` binary `execve` count for this turn-end is 0; for the positive case, arm a goal (write the state file), fire Stop, assert `execve` count is 1.

#### AC-STOPCHAIN-TRIM-002 — Non-sync HEAD skips vet/build (A10)

**Given** a working tree whose current HEAD commit subject does NOT match the sync-commit sentinel (e.g. a run-phase `feat(scope): ...` commit, not a `docs(sync): ...` commit), AND `sync-phase-quality-gate.sh` installed as a Stop hook.

**When** the Stop event fires.

**Then** the wrapper exits 0 at the shell layer WITHOUT running `go vet`, `go build`, or `golangci-lint`, AND no diagnostic subprocess is spawned from this hook for this turn-end.

**Test shape:** shell fixture — set HEAD to a non-sync subject (e.g. via a test git dir), install counting wrappers around `go vet` / `go build` / `golangci-lint`, fire Stop, assert each command's invocation count is 0; flip HEAD to a sync-commit subject, fire Stop, assert the commands run as before (no false negative on the positive case).

#### AC-STOPCHAIN-TRIM-003 — Per-edit spawns eliminated on TDD cycle (A8)

**Given** a `manager-develop` cycle running a TDD RED-GREEN-REFACTOR loop that performs N (e.g. N=20) atomic Write|Edit operations, AND the per-edit PreToolUse `develop-pre-implementation` + PostToolUse `develop-post-implementation` hooks removed per A8, AND the per-cycle `develop-completion` Stop hook registered.

**When** the N atomic edits execute across the cycle.

**Then** the per-edit pre/post hook spawn count is 0 across the N edits (no per-edit spawn), AND the per-cycle `develop-completion` Stop hook fires exactly once for the cycle (or once per milestone commit boundary, whichever applies).

**Test shape:** counting-wrapper fixture — install a spawn counter around the per-edit hook entry points, run a 20-edit TDD fixture, assert the per-edit spawn count is 0; separately assert the per-cycle Stop hook fired the expected number of times (1 for a single-cycle fixture).

#### AC-STOPCHAIN-TRIM-004 — fully-autonomous makes sync-gate advisory (A11)

**Given** `MOAI_AUTONOMY_TIER=fully-autonomous` is set, AND `sync-phase-quality-gate.sh` evaluates a FAILING quality decision (e.g. the snapshot records a lint error).

**When** the Stop hook runs.

**Then** the hook emits an advisory `systemMessage` recording the failure, AND the hook's stdout JSON does NOT contain `"decision":"block"` (or the legacy exit-2 block), AND the turn-end is NOT blocked by this hook.

**Test shape:** set the tier, force a failing lint in the snapshot, run the hook, assert stdout JSON has no `decision:block` (and exit code is not 2); flip tier to `semi-auto`, re-run, assert the block returns (regression guard).

#### AC-STOPCHAIN-TRIM-005 — automatic turns commit gate OFF (A11)

**Given** `MOAI_AUTONOMY_TIER=automatic` is set, AND a `git commit` PreToolUse event fires in `internal/hook/pre_tool.go` at the IsGitCommit branch (~L429-441).

**When** the gate evaluates.

**Then** the synchronous vet+lint+test verification is NOT invoked (the gate returns allow without calling those tools), AND the commit proceeds without the verification tax.

**Test shape:** Go unit test over the IsGitCommit branch — inject a tier of `automatic`, simulate a commit PreToolUse event, assert the verification-tool invocation count is 0; flip tier to `semi-auto`, assert the verification tools ARE invoked (regression guard).

#### AC-STOPCHAIN-TRIM-006 — deny/ask rules still bind at every tier (cross-cutting safety)

**Given** ANY value of `MOAI_AUTONOMY_TIER` ∈ {unset, `semi-auto`, `automatic`, `fully-autonomous`}, AND a tool call that matches a destructive-pattern deny rule (e.g. `git push origin main`, a secret in pending diff, `rm -rf`, a deploy command).

**When** the PreToolUse gate evaluates.

**Then** the gate DENIES the tool call regardless of tier — the deny decision is identical across all four tier values, AND no tier weakens or skips a deny rule.

**Test shape:** table-driven test — for each of the four tier values, exercise each denylist fixture (main push, secret, `rm -rf`, deploy), assert the gate decision is `deny` in every cell of the 4×N matrix. This is the load-bearing safety regression guard; any cell flipping to allow is a hard FAIL.

#### AC-STOPCHAIN-TRIM-006b — fully-autonomous makes subagent lifecycle hooks dormant (A11)

**Given** `MOAI_AUTONOMY_TIER=fully-autonomous` is set, AND any of the three subagent lifecycle events fires — SubagentStop, TeammateIdle, TaskCompleted.

**When** each lifecycle hook evaluates the event.

**Then** the hook records an audit-log entry (observe-only), AND the hook does NOT emit `decision:block` (or legacy exit-2), AND the hook does NOT emit the `continue:false` / `stopReason` reject contract for TaskCompleted, AND the orchestrator is NOT required to translate the hook's output into an `AskUserQuestion` round for this event. Each of the three lifecycle events exercises this path independently.

**Test shape:** three lifecycle-event fixtures (one per event type) at `fully-autonomous`; for each, assert exit 0 (or the equivalent non-blocking result for the event's contract — `decision` absent or `allow` for SubagentStop/TeammateIdle, `continue:true` or absent for TaskCompleted), assert no `AskUserQuestion`-translation trigger from this hook, AND assert an audit-log entry WAS written (observe-only is NOT silent — it records). Flip tier to `semi-auto` and re-run each fixture to confirm the active blocking/reject behavior returns (regression guard). This AC is P0 safety behavior for unattended autonomous loops — a lifecycle hook accidentally left active at `fully-autonomous` would block/reject/translate-to-AskUserQuestion and break the autonomous loop.

#### AC-STOPCHAIN-TRIM-007 — Unset token preserves semi-auto behavior (backward compat)

**Given** `MOAI_AUTONOMY_TIER` is unset (empty) in the environment and in `workflow.yaml`.

**When** any hook or gate reads the token.

**Then** the reader returns `semi-auto`, AND every hook's behavior is identical to today's pre-SPEC behavior — `sync-phase-quality-gate.sh` full-blocks, the IsGitCommit gate runs full verification, the subagent lifecycle hooks are active, the `handle-stop-goal.sh` shell precondition still fires when a goal IS armed.

**Test shape:** unset the token, run the full hook surface end-to-end (sync-gate blocking, commit gate verifying, lifecycle hooks active), assert zero behavior delta from a baseline captured before the SPEC lands. This AC MUST be green for the SPEC to be considered non-regressive.

### §D.3 Indirect verification

- The 5s hook wrapper timeout invariant (CLAUDE.local.md §7) is verified indirectly: every shell-layer precondition added in A10 is a single file-existence test or a single `git log -1 --format=%s` call. The shell-trim MUST NOT introduce a sequence that approaches 5s. Verified by timing the wrapper under a representative fixture.

### §D.4 Closure gates

- All eight MUST ACs green with attributed evidence (command + verbatim output) — AC-001 through AC-007 plus AC-006b.
- The 4×N deny/ask matrix (AC-006) is fully green — no tier weakens any deny.
- The 3 lifecycle events at `fully-autonomous` (AC-006b) are all dormant yet audit-logging.
- The backward-compat sweep (AC-007) confirms zero behavior delta on unset token.
- LSP gate: zero errors, zero type errors, lint clean.

### §D.5 Forward-looking checks (advisory, non-blocking for this SPEC)

- The `MOAI_AUTONOMY_TIER` token's UI surface (web console toggle, `moai init` interactive) — when that surface lands, it MUST write to the same canonical source picked in M1, not a parallel one.
- The stateful MCP layer (epic P2) will read the same token; the M1 reader helper should be the single re-used API.

### §D.6 Definition of Done

This SPEC is DONE when:

1. All eight MUST ACs pass with attributed evidence (AC-001 through AC-007 plus AC-006b).
- The deny/ask 4×N matrix is fully green (no regression).
- The backward-compat sweep is green.
- The project's standard quality gate is green.
- Frontmatter `status` transitions `draft → in-progress → implemented → completed` are owned by manager-develop and manager-docs; this plan-phase authoring only emits `draft`.
