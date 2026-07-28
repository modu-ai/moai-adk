# plan.md — SPEC-PREEDIT-PARALLEL-SESSION-GUARD-001

> Tier M. Run-phase cycle_type: ddd (behavior-preserving doctrine extension — no production code logic changed unless REQ-PES-004 implements the hook).

## Design summary

Extend parallel-session detection from the spawn boundary to the direct-edit boundary, mirroring the proven Pre-Spawn Sync Check. Primary enforcement is procedural (orchestrator-run); a mechanical PreToolUse-on-Edit hook is evaluated and likely deferred on cost grounds; an ambient foreign-session signal gives cheap awareness.

## Milestones

### M1 — Pre-Edit Sync Check (doctrine) [REQ-PES-001]
- Add a `### Pre-Edit Sync Check` section to `.claude/rules/moai/core/agent-common-protocol.md`, immediately after the Pre-Spawn Sync Check section.
- Content: before a non-trivial direct edit (definition in spec.md §D) to shared paths, run the same 3-command batch as Pre-Spawn (`git fetch origin main` + `git rev-list --left-right origin/main...HEAD` + `moai session list --json`), interpret via the same matrix, and on a foreign-session/race signal isolate to a worktree or surface via AskUserQuestion.
- Mirror to `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md`.

### M2 — auto-isolation trigger broadening [REQ-PES-002, REQ-PES-003]
- Edit `.claude/rules/moai/workflow/worktree-integration.md` § Parallel-Session Branch Conflict Auto-Isolation: relax the "worktree entry is chosen" conjunct so the trigger also fires for direct-edit write work when ≥1 foreign active session is on the same checkout.
- Name the "direct edit bypasses the spawn gate" failure mode inline + cross-reference the new Pre-Edit Sync Check.
- Mirror to template.

### M3 — bypass failure-mode codification [REQ-PES-003]
- Add a one-line explicit note in the Pre-Spawn Sync Check section: "Direct main-session edits (Edit/Write/Bash) bypass this spawn gate; see Pre-Edit Sync Check." (cross-link both ways).
- Mirror to template.

### M4 — PreToolUse-on-Edit hook evaluation [REQ-PES-004]
- Evaluate: a PreToolUse hook on Edit/Write reading `.moai/state/active-sessions.json` + optional fetch. Record the per-edit cost finding (registry-read is cheap ~ms; a per-edit `git fetch` is expensive and would be gated to "first edit of a turn" or omitted).
- Decision: implement a **read-only advisory** hook (registry-read only, no fetch, no block — surfaces a `systemMessage` when foreign sessions live) OR **defer** the blocking hook to a follow-up SPEC with rationale. Record the decision + rationale in acceptance.md §F.

### M5 — advisory ambient signal [REQ-PES-005]
- If M4 does not add a hook, add an ambient signal: the orchestrator's session-start reads `active-sessions.json` and notes foreign-session count (cheap). Document the behavior in doctrine (no code, or minimal statusline note if a statusline segment is cheap to add).

### M6 — template mirror + build + verify + PR
- All `.claude/` edits mirrored to `internal/template/templates/`.
- `make build` (catalog regenerated if skills/rules enumerated).
- Verify: `go test ./internal/template/...` (Neutrality/Leak/Mirror/Parity), neutrality grep (no SPEC-ID/date/SHA leak into templates), `go build ./...`.
- Commit + push + PR from the isolated worktree.

## Files (working-tree, relative to worktree root)
- `.claude/rules/moai/core/agent-common-protocol.md` (M1, M3) + template mirror
- `.claude/rules/moai/workflow/worktree-integration.md` (M2) + template mirror
- (M4, conditional) `internal/hook/pre_tool.go` or a new `internal/hook/session_guard.go` + settings.json wiring — ONLY if the hook is implemented; if deferred, no Go change.
- (M5, optional) statusline/session-start note.

## Decisions to confirm at Implementation Kickoff
1. Hook: advisory-read-only (implemented) vs deferred (procedural-only this SPEC)? — **recommendation: implement advisory-read-only** (cheap, mechanical nudge) + keep the blocking hook as a follow-up.
2. Ambient signal: session-start note (no code) vs statusline segment (minimal code)?
