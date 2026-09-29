# SPEC-CODEX-ROLE-AUDIT-ROOT-001 — Acceptance Criteria

Two-cell discipline (verification-completeness rule): every release-blocking criterion carries a RED-now cell (command, verbatim output, exit code, tree SHA) and a green-path cell naming the milestone that flips it. Non-release-blocking criteria are classified regression-guard explicitly.

**Tree SHA for all RED cells**: `7ca01df35` (worktree `.moai/worktrees/t1324`, branch `WT-roleaudit-root`).
**Baseline attribution key**: `[lead]` = card t1324 lead-verified live measurement (two lanes); `[local]` = measured first-hand in this plan phase against a binary built from this tree (`go build -o /tmp/t1324-moai ./cmd/moai`, invoked by path).

## §D AC Matrix

### AC-001 — MCP role audit accepts a registered linked worktree root (release-blocking, High)

- **RED-now** `[lead]` — MCP stdio `tools/call` `codex_role_audit` with `worktree_root` set to a registered linked worktree of the same repository, server started in the primary checkout: tool result `isError: true`, diagnostic carrying the verbatim fragment `<root> is not the caller's own worktree (/Users/goos/MoAI/moai-adk-go)`. Source anchor at this SHA: `internal/cli/codex_audit_launch.go:308-309`. Run-phase MUST re-observe RED before flipping (M0 pre-flight); the plan-phase probe environment hit a lower-layer refusal (`the server did not start inside a git worktree`) and could not reach this clause — recorded as a measurement-environment gap, not as absence of the defect.
- **Green path** — M1 flips: the same call returns `isError: false` with a `job_id`, and no refusal diagnostic.

### AC-002 — CLI role-audit verb exists (release-blocking, High)

- **RED-now** `[local]` — command: `/tmp/t1324-moai codex role-audit`; verbatim output:
  `unknown verb - usage: moai codex [cli] [-w <worktree>] [-f [lane|lane-<n>]] [--factory-run <id>] [-- codex-args...] | moai codex status | moai codex app`; exit code `1`; tree `7ca01df35`.
- **Green path** — M3 flips: the verb launches the role against the caller's own tree from the lane cwd, exiting with the audit's exit code.

### AC-003 — codex_task accepts project_root and refuses its absence (release-blocking, High)

- **RED-now** `[local]` — command: MCP stdio `tools/list` against `/tmp/t1324-moai mcp-server`; verbatim property lists:
  `codex_task -> ['background', 'prompt', 'resume_last', 'thread_id', 'work_key', 'write']` (no `project_root`); exit code `0`; tree `7ca01df35`.
- **Green path** — M2 flips: `project_root` appears in the schema and an argument-less request returns `isError: true` naming the missing argument. The gating arm carries its own observable (defeats the schema-only mutant that adds the argument without wiring the validator): a `codex_task` request whose `project_root` resolves to a path outside the serving checkout's repository returns `isError: true` carrying the verbatim refusal fragment `belongs to a different repository` (`codex_audit_launch.go:322`).

### AC-004 — Same-repository refusals survive (regression-guard)

The out-of-repo refusal ("belongs to a different repository", `codex_audit_launch.go:322`) and the unregistered-worktree refusal ("is not a registered worktree of this repository", line 337) must remain observable after M1 with unchanged semantics.

### AC-005 — Primary-checkout callers unaffected (regression-guard)

A server started in the primary checkout with an explicitly presented primary toplevel is accepted — the primary is a registered worktree of its own repository.

### AC-006 — Missing worktree_root refused, no silent fallback (regression-guard)

An empty or absent `worktree_root` on `codex_role_audit` yields a structured refusal (today: the required-argument check at `codex_audit_mcp.go:91-92`); M1 must not replace it with a default-tree action.

### AC-007 — Destination containment and record format unchanged (regression-guard)

Verdict and launch-record destinations stay contained under the accepted root's report tree (`codexAuditValidateDest`), and the launch-record format is byte-compatible with today's.

## §D.1 Severity

AC-001..AC-003: release-blocking (High). AC-004..AC-007: regression-guard (Medium).

## §D.2 Traceability

REQ-001→AC-001 · REQ-002→AC-006 · REQ-003→AC-004/AC-005 · REQ-004→AC-002 · REQ-005→AC-003 · REQ-006→AC-007.

## §D.3 Indirect verification

AC-001's green path is verified indirectly through the job lifecycle (`codex_role_audit_status` → `codex_role_audit_result` returning `state: completed|failed` with a real launch record path) rather than by inspecting the codex process directly.

## §D.4 Closure gates

All release-blocking cells green; all regression-guard cells green; §E1 evidence recorded in progress.md before sync entry.

## §D.5 Forward-looking checks

At sync, `.claude/rules/moai/core/moai-mcp-tools.md` gains the role-audit surface in its `project_root`-family table; the tool description text no longer promises the server-started-in contract (E4 grep).
