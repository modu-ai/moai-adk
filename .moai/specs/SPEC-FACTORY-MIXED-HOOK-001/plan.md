---
id: SPEC-FACTORY-MIXED-HOOK-001
document: plan
created: 2026-09-22
updated: 2026-09-23
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
- Keep the broker address keyed by stable logical lane ID. This card owns only the initial `launch-pending → bound` transition; `t1082` owns worktree creation and the later worktree-handoff transition/rebind through the same generation-safe seam.

## Milestones

### M1 — Identity, run selection, and canonical namespace

- Extend `internal/homestate` with a canonical factory-message path and run/session generation records.
- Add active-run resolver for 0/1/many plus `--factory-run` parsing in `internal/cli/factory.go`, `codex_factory.go`, and Claude launch path.
- Make worker slot claim run-scoped and persist backend/session/run/process identity.
- Register a `launch-pending` provisional endpoint from the launcher's real child PID/process-start plus stable lane/run identity before the first turn; accept `agent-N` as well as `lane-N`.
- Treat startup `SessionStart` as an idempotent best-effort early-bind opportunity only when it runs after provisional registration; do not depend on that ordering.
- On the first legitimate non-empty `UserPromptSubmit`, before inbox batch processing, atomically replace any remaining provisional endpoint with the actual hook `session_id` plus resolved owner PID/process-start. Empty or whitespace-only prompts do not bind.
- Once the same actual endpoint is bound, later `UserPromptSubmit` events continue inbox checks but perform a no-write bind no-op: generation, peer fields, and `updated_at` remain unchanged.
- Preserve the logical lane ID across session generations and reject delivery or receipt through a stale physical endpoint.
- Tests: worktree convergence, project/run isolation, concurrent join, ambiguous/no run, restart generation, PID reuse.

### M2 — Factory broker and explicit receipt

- Add factory-specific types/store alongside, not inside, legacy semantics in `internal/sessionmsg` or a narrowly named sibling package.
- Implement validated envelope, idempotent send, atomic claim token, body read, persisted disposition, explicit receipt/ack, lease redelivery, dead-letter, and bounds.
- Expose MCP register/list/send/body-read/receipt/status handlers through the existing MCP registration path.
- Tests: retry dedupe, stale ACK, crash before/after output, concurrent poll, TTL/poison/overflow, traversal/spoof/revision rejection.

### M3 — Hook delivery and wiring

- Add a common idempotent bind seam to SessionStart and UserPromptSubmit. SessionStart may bind early after provisional registration; the first non-empty UserPromptSubmit is the required fallback and performs binding before the same hook's inbox batch. Stop retains delivery integration only.
- Guard the bind seam so an already-bound identical endpoint returns without a registry write, while every UserPromptSubmit still runs its inbox read path.
- Compose bounded metadata-only additional context without overwriting existing context.
- Make Claude Stop wiring synchronous; enforce one continuation per chain and ledger-based duplicate suppression.
- Update Codex adapter/wiring only where measured output shape requires it. Preserve deny/stop precedence.
- Tests: empty/receipt-only zero turn, 2 KiB/16 bound, malicious body absent, Stop→UserPrompt duplicate prevention, incompatible/trust-disabled capability error.

### M4 — Live matrix, benchmark, and evidence

- Build an isolated binary and run distinct real sessions for all required pairs plus Claude↔Claude.
- Exchange per-direction nonce, fetch body as untrusted data, persist disposition, send explicit receipt, and read it back.
- Benchmark empty/16/1000 pending, 1/10 sessions, warm/cold, and lock contention. Record p50/p95, wall time, file reads/writes, process count, and injected bytes.
- Write `.moai/reports/t1074/verdict.md`; `NOT_RUN` or simulated model context is FAIL for live ACs.

### M5 — Operational lane roster/status extension (pending)

- Implement REQ-FMH-OPS-001 through REQ-FMH-OPS-008 according to the normative [`operational-lane-status-addendum.md`](./operational-lane-status-addendum.md), reusing the existing run-scoped `peers` registry and read-only `factory_msg_status` surface rather than introducing a second roster.
- Execute the addendum's M1–M4 sequence: store read model and endpoint truth table, non-mutating MCP response, launcher provisional registration plus required `UserPromptSubmit` rebind with optional SessionStart early bind, then installed-binary/MCP-restart live proof in a separate fixture project.
- Treat AC-FMH-OPS-001 through AC-FMH-OPS-006 as new pending criteria. Existing AC-FMH-001 through AC-FMH-015 results remain historically attributed and do not satisfy the OPS criteria.
- Run the exact non-empty unit and production-launcher LIVE gates in addendum §6 as one combined acceptance package. The unit selector `TestFactoryUserPromptSubmitRebindsLaunchPendingPeer` owns the empty/whitespace no-bind proof. The LIVE chain does not send empty/whitespace input; it must use one built-tree `moai codex -f` lead plus two built-tree `moai codex -f agent` processes, prove all three process-backed provisional lanes before any prompt, prove required actual-session rebind during each lane's legitimate non-empty first UserPromptSubmit before inbox processing, and prove a later normal prompt on the same bound endpoint does not rewrite peer state. Direct `codex exec`, manually injected `MOAI_SESSION_PID`, direct `RegisterPeer`, pre-seeding, trust bypass, mock, skip, `NOT_RUN`, or in-process-only evidence cannot close its real-session criteria.
- At M5 Definition of Done, execute the current-baseline regression gates in `acceptance.md` for all existing AC-FMH-001 through AC-FMH-015 against the post-M5 tree. Historical logs cannot substitute; a live row that cannot run remains FAIL/GAP and blocks inherited PASS.

## Verification commands

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/t1074-test-home GOCACHE=/tmp/t1074-go-cache go test ./internal/homestate ./internal/sessionmsg ./internal/kanban ./internal/cli ./internal/hook ./internal/codexadapter -count=1
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/t1074-race-home GOCACHE=/tmp/t1074-race-cache go test -race ./internal/sessionmsg ./internal/kanban ./internal/hook -count=1
git diff --check
```

Live and benchmark commands SHALL be authored by the implementation with fixed timeouts, isolated state roots, cleanup-guaranteed child processes, and verbatim logs under `.moai/reports/t1074/` or `/tmp/t1074-*` cited by the verdict.

## Follow-up dependency boundary

- `t1082` consumes the M1 logical-lane/current-endpoint seam after this card has established initial `launch-pending → bound`, then implements `reserve → worktree create → SWITCH_PENDING → /cd or headless cwd fork → endpoint rebind → BOUND`.
- `t1075` may wake only the endpoint that `t1082` has marked current and `BOUND`; it must reject the pre-handoff generation.
- Neither follow-up expands t1074 into a worktree launcher or an idle-session controller.
