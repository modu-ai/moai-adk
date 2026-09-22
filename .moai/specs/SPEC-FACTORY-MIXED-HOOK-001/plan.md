---
id: SPEC-FACTORY-MIXED-HOOK-001
document: plan
status: in-progress
created: 2026-09-22
updated: 2026-09-22
author: manager-spec
card: t1074
module: "internal/factorymsg"
---

# Plan — SPEC-FACTORY-MIXED-HOOK-001

## Constraints

- Work only in card worktree `WT-factory-mixed-hook`; preserve shared primary checkout.
- Preserve legacy `internal/sessionmsg` behavior and data.
- No raw peer text in elevated hook context.
- No queue completion, shell execution, or authority escalation from messages.
- Template-managed hook/config changes follow template-first parity.
- Keep the broker address keyed by stable logical lane ID. `t1082` alone owns worktree creation and endpoint transition/rebind; this card only exposes the generation-safe binding seam it consumes.

## Milestones

### M1 — Identity, run selection, and canonical namespace

- Extend `internal/homestate` with a canonical factory-message path and run/session generation records.
- Add active-run resolver for 0/1/many plus `--factory-run` parsing in `internal/cli/factory.go`, `codex_factory.go`, and Claude launch path.
- Make worker slot claim run-scoped and persist backend/session/run/process identity.
- Bind actual hook `session_id` at SessionStart; accept `agent-N` as well as `lane-N`.
- Preserve the logical lane ID across session generations and reject delivery or receipt through a stale physical endpoint.
- Tests: worktree convergence, project/run isolation, concurrent join, ambiguous/no run, restart generation, PID reuse.

### M2 — Factory broker and explicit receipt

- Add factory-specific types/store alongside, not inside, legacy semantics in `internal/sessionmsg` or a narrowly named sibling package.
- Implement validated envelope, idempotent send, atomic claim token, body read, persisted disposition, explicit receipt/ack, lease redelivery, dead-letter, and bounds.
- Expose MCP register/list/send/body-read/receipt/status handlers through the existing MCP registration path.
- Tests: retry dedupe, stale ACK, crash before/after output, concurrent poll, TTL/poison/overflow, traversal/spoof/revision rejection.

### M3 — Hook delivery and wiring

- Add a common hook integration to SessionStart, UserPromptSubmit, and Stop.
- Compose bounded metadata-only additional context without overwriting existing context.
- Make Claude Stop wiring synchronous; enforce one continuation per chain and ledger-based duplicate suppression.
- Update Codex adapter/wiring only where measured output shape requires it. Preserve deny/stop precedence.
- Tests: empty/receipt-only zero turn, 2 KiB/16 bound, malicious body absent, Stop→UserPrompt duplicate prevention, incompatible/trust-disabled capability error.

### M4 — Live matrix, benchmark, and evidence

- Build an isolated binary and run distinct real sessions for all required pairs plus Claude↔Claude.
- Exchange per-direction nonce, fetch body as untrusted data, persist disposition, send explicit receipt, and read it back.
- Benchmark empty/16/1000 pending, 1/10 sessions, warm/cold, and lock contention. Record p50/p95, wall time, file reads/writes, process count, and injected bytes.
- Write `.moai/reports/t1074/verdict.md`; `NOT_RUN` or simulated model context is FAIL for live ACs.

## Verification commands

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/t1074-test-home GOCACHE=/tmp/t1074-go-cache go test ./internal/homestate ./internal/sessionmsg ./internal/kanban ./internal/cli ./internal/hook ./internal/codexadapter -count=1
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/t1074-race-home GOCACHE=/tmp/t1074-race-cache go test -race ./internal/sessionmsg ./internal/kanban ./internal/hook -count=1
git diff --check
```

Live and benchmark commands SHALL be authored by the implementation with fixed timeouts, isolated state roots, cleanup-guaranteed child processes, and verbatim logs under `.moai/reports/t1074/` or `/tmp/t1074-*` cited by the verdict.

## Follow-up dependency boundary

- `t1082` consumes the M1 logical-lane/current-endpoint seam to implement `reserve → worktree create → SWITCH_PENDING → /cd or headless cwd fork → SessionStart rebind → BOUND`.
- `t1075` may wake only the endpoint that `t1082` has marked current and `BOUND`; it must reject the pre-handoff generation.
- Neither follow-up expands t1074 into a worktree launcher or an idle-session controller.
