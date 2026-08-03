# plan.md — SPEC-STOPCHAIN-TRIM-001

> Implementation plan. Order is decision-reversibility-first: the mode token contract (OQ-1, OQ-3) leads; mechanical shell-trim and per-edit consolidation follow.

## §A. Context

This SPEC is a **redesign codification**, not greenfield. The design authority is `.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §3.5 (A8 / A10 / A11) + §3.1 (mode token) + §1.4 (hook breakpoints). Per the report's P0 decomposition (§3.5 closeout row "SPEC-STOPCHAIN-TRIM"), this single SPEC bundles A8+A10+A11+the token because the mode-aware hooks (A11) are inert without the token, and the shell-trim (A10) + per-edit consolidation (A8) are the two biggest per-turn / per-edit latency hits the report identifies.

The report's §3.5 closeout note (P0 pair rationale): "A4 + A3 take sync time; A10 + A8 take run / general-session per-turn latency. Both P0s are mostly existing-knob rewiring, no new mechanism."

## §B. Known Issues

- **K-1** (RESOLVED via OQ-1, iter 1): the `MOAI_AUTONOMY_TIER` token's canonical home was env-key vs `workflow.yaml` key. Resolved to **env-key** (`internal/config/envkeys.go`), because shell hooks must read the token without the `moai` binary or YAML parsing. `workflow.yaml` is NOT a parallel source. This K is closed; the resolution is folded into REQ-003 + plan M1.
- **K-2**: Reading the mode token inside `internal/hook/pre_tool.go` requires the config to be loaded at PreToolUse time. If the config loader is not yet initialized when PreToolUse fires, the read returns empty (= `semi-auto`, the backward-compat default) — acceptable as a fail-open, but MUST be confirmed not to silently mis-classify a `fully-autonomous` session as `semi-auto` due to a load-timing bug.
- **K-3**: The shell-layer precondition in `handle-stop-goal.sh` needs the current session ID to locate `.moai/state/goal/<session-id>.json`. The wrapper today may or may not receive the session ID from the runtime; if not, the precondition check cannot fire and A10's "goal-absent → exit 0" becomes unimplementable without a runtime contract change.
- **K-4**: The `develop-completion` per-cycle Stop hook (A8) — OQ-2 asks whether it already exists. If it does NOT, A8 introduces a new Stop hook registration, expanding the blast radius from "edit existing" to "add new hook".
- **K-5**: The 4 observer/security Stop hooks' async transition (OQ-3) — if the runtime has no async/advisory hook channel today, A10's "async transition" requires a new hook-type enum value in the runtime, which is a runtime change, not a hook-script change.
- **K-6**: The deny/ask invariance (REQ-007) is the load-bearing safety constraint. Any refactor that moves the denylist evaluation relative to the mode-token read MUST be audited to confirm the deny path is reached BEFORE the mode-token branch, not after.

## §C. Pre-flight (read-only reconnaissance — before M1)

1. Read `.claude/hooks/moai/handle-stop-goal.sh` in full — confirm current exec pattern + whether session ID is available at the shell layer (K-3).
2. Read `.claude/hooks/moai/sync-phase-quality-gate.sh` — confirm the existing once-per-commit sentinel and its grep target (K-3 / A10 second clause).
3. Read `.claude/agents/moai/manager-develop.md` frontmatter — confirm PreToolUse + PostToolUse hook registration shape (A8).
4. Read `internal/hook/pre_tool.go` around the IsGitCommit branch (L429-441 cited) — confirm where the mode-token read would slot in.
5. Read `internal/config/envkeys.go` + `internal/config/defaults.go` — resolve OQ-1 (existing analog surface).
6. Grep `.claude/settings.json` for the 4 observer/security Stop hooks (K-5); check whether the runtime exposes an async hook channel.

## §D. Constraints (recap from spec.md §D — binding on the plan)

1. deny/ask rules tier-invariant (REQ-007). No tier weakens a deny.
2. Backward compat: unset/empty token = `semi-auto` = today.
3. Single source of truth for `MOAI_AUTONOMY_TIER`.
4. No new CLI surface.
5. Tier M.
6. Hook wrapper timeout stays at 5s.

## §E. Self-Verification (run-phase — what manager-develop must demonstrate)

- Shell-fixture test demonstrating `handle-stop-goal.sh` does NOT exec the `moai` binary when the goal state file is absent (AC-001).
- Shell-fixture test demonstrating `sync-phase-quality-gate.sh` does NOT run `go vet` / `go build` when HEAD subject is a non-sync commit (AC-002).
- Counting-wrapper fixture demonstrating 0 pre/post per-edit spawns across an N-edit TDD cycle (AC-003).
- Mode-token branch tests for sync-gate (AC-004), commit gate (AC-005), and subagent lifecycle (AC-006 cross-ref).
- Regression test: deny/ask rules still bind at `fully-autonomous` (AC-006).
- Backward-compat test: unset `MOAI_AUTONOMY_TIER` exercises `semi-auto` behavior end-to-end (AC-007).
- Lint / build / test clean.

## §F. Milestones

### Milestone M1 — Mode token introduction (OQ-1 resolution)

Highest reversibility: the token's canonical home determines every downstream hook read. K-1 hazard.

**Files (expected):**
- `internal/config/envkeys.go` — define the env-key constant `MOAI_AUTONOMY_TIER` (the single canonical source per OQ-1 resolution — env-key, NOT a `workflow.yaml` key, because shell hooks must read it without the `moai` binary).
- `internal/config/defaults.go` — record the 3-value enum (`semi-auto`, `automatic`, `fully-autonomous`) and the `semi-auto` default for unset/empty.
- A reader helper (e.g. `config.AutonomyTier() string` returning `semi-auto` on unset/empty) wrapping `os.Getenv(MoaiAutonomyTierKey)`.
- A unit test covering: unset → `semi-auto`; each of the 3 values round-trips; an invalid value → `semi-auto` (fail-safe to backward-compat default).

**Exit:** token read helper exists; backward-compat behavior verified (AC-007 partially exercisable from here).

### Milestone M2 — Stop-chain shell trim (A10)

Second reversibility tier: the shell preconditions are mechanical IF the wrapper has the inputs it needs (K-3).

**Files (expected):**
- `.claude/hooks/moai/handle-stop-goal.sh` — add shell precondition: `[ -f "$state_file" ] || exit 0` before the `exec moai hook stop-goal`. State-file path derived from session ID (K-3 precondition).
- `.claude/hooks/moai/sync-phase-quality-gate.sh` — add shell precondition: grep HEAD subject for the sync-commit sentinel; `exit 0` on mismatch.
- The 4 observer/security Stop hooks (per OQ-3 resolution): wrap in async dispatch OR transition to advisory `systemMessage` mode.

**Exit:** AC-001 + AC-002 green; observer/security async transition verified (no synchronous block on turn-end).

### Milestone M3 — Per-edit hook consolidation (A8)

Third reversibility tier: depends on OQ-2 (does `develop-completion` Stop hook exist?). If it exists, this is a content migration; if not, it adds a new Stop hook.

**Files (expected):**
- `.claude/agents/moai/manager-develop.md` frontmatter — remove the per-edit PreToolUse `develop-pre-implementation` + PostToolUse `develop-post-implementation` firing. Residual PreToolUse retains ONLY destructive-pattern + scope-discipline.
- `develop-completion` Stop hook (new or existing per OQ-2) — carries the verification workload that the per-edit hooks used to perform, firing once per cycle / milestone-commit boundary.
- Tests demonstrating 0 per-edit spawns over an N-edit TDD cycle (AC-003).

**Exit:** AC-003 green; TDD cycle latency drop demonstrated.

### Milestone M4 — Mode-aware hooks (A11)

Depends on M1 (token). The 3 hook surfaces (sync-gate, commit gate, subagent lifecycle) branch on the token.

**Files (expected):**
- `.claude/hooks/moai/sync-phase-quality-gate.sh` — read `MOAI_AUTONOMY_TIER`; branch advisory (fully-autonomous) / build-only-block (automatic) / full-block (semi-auto).
- `internal/hook/pre_tool.go` IsGitCommit branch (~L429-441) — read the token; gate OFF on automatic/fully-autonomous; retain deny/ask denylist evaluation BEFORE the tier branch (K-6).
- SubagentStop / TeammateIdle / TaskCompleted hooks — read the token; dormant (observe-only) on fully-autonomous.

**Exit:** AC-004 + AC-005 + AC-006 + AC-006b green; deny/ask regression test green at every tier; lifecycle-dormant regression green at `fully-autonomous`.

### Milestone M5 — Verify

Lowest reversibility.

- Full `go test ./...` + race.
- All AC matrix green with attributed evidence.
- Backward-compat sweep: run a session with `MOAI_AUTONOMY_TIER` unset and confirm zero behavior delta from today.

## §G. Anti-Patterns (specific to this SPEC)

- **AP-1**: Putting the denylist evaluation AFTER the mode-token branch in `pre_tool.go`, so a `fully-autonomous` session accidentally skips a deny rule. The deny path MUST be evaluated unconditionally; the tier branch only affects the verification gate.
- **AP-2**: Defining `MOAI_AUTONOMY_TIER` in BOTH `envkeys.go` AND `workflow.yaml` — parallel restatement; pick ONE in M1.
- **AP-3**: Defaulting an UNSET token to `fully-autonomous` "for convenience" — violates backward compat (REQ-003); unset MUST be `semi-auto`.
- **AP-4**: Moving the per-edit verification workload to the Stop hook but ALSO leaving the per-edit Pre/PostToolUse hooks in place "as a belt-and-suspenders" — the per-edit tax returns; A8 is a MOVE not a COPY.
- **AP-5**: Implementing the shell precondition in `handle-stop-goal.sh` as a `moai` binary call (`moai goal exists`) instead of a shell `[ -f ]` — defeats A10's purpose (the cold-start tax returns). The precondition MUST be shell-only.
- **AP-6**: Raising the hook wrapper timeout beyond 5s to accommodate the mode-token read — the read is a config lookup, well under 5s; the timeout stays (CLAUDE.local.md §7).

## §H. Cross-References

- spec.md: `.moai/specs/SPEC-STOPCHAIN-TRIM-001/spec.md`.
- acceptance.md: `.moai/specs/SPEC-STOPCHAIN-TRIM-001/acceptance.md`.
- Design report: `.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §3.5 (A8/A10/A11) + §1.4 + §3.1.
- Epic sibling P0: `SPEC-AUDIT-SNAPSHOT-001` (A1-A4).
- CLAUDE.local.md §7 (5s hook timeout invariant).
- `verification-claim-integrity.md` §1.1 (the deny/ask invariance is a safety claim — REQ-007).
