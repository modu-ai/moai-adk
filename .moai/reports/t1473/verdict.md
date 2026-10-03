# t1473 verdict — factory_msg_send optional ids (GitHub #1737)

card: t1473 · class B (no SPEC) · cycle: tdd · branch: `WT-factory-msg-optional-ids`
base: local `develop` `b5815ca80` (fast-forwarded from `origin/develop` `3f3ebb763`)
commits: `00a77a839` (RED test) → `50068c75d` (GREEN fix)

## Claim

1. Before the fix, `factory_msg_send` called without `task_ref` / `correlation_id` failed with a tool error, although the schema registers both as optional.
2. After the fix, the same call delivers; an omitted or empty `task_ref` / `correlation_id` defaults to `idempotency_key`.
3. A same-key retry without the fields returns the original message id (no "idempotency key collision" error).
4. Caller-supplied values are preserved unchanged.
5. The store validation (`internal/factorymsg/store.go` `safeID` checks) is untouched; the invariant "every message carries both ids" still holds.

Cause: `internal/cli/mcp_factory_msg.go` forwarded `req.GetString(..., "")`; `Store.Send` rejects `""` via `safeID` (`^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$`) before the INSERT. The `task_ref` check runs first, so omitting both ids surfaces as "invalid task reference"; omitting only `correlation_id` surfaces as the reported "invalid correlation id".

Fix: `internal/cli/mcp_factory_msg.go` — `TaskRef: cmp.Or(req.GetString("task_ref", ""), idem)`, same for `CorrelationID`; `internal/cli/mcp_server.go` — both parameters now carry a description stating the default.

Choice for `task_ref`: its only non-store consumer is a display line (`internal/cli/managed_factory_session.go:403`, `task_ref=%s`); the columns are `NOT NULL` in the schema and migration DDL. Defaulting in the handler (rather than relaxing the store) keeps one rule for both fields and leaves the store's poison-input diagnostics intact.

## Evidence

RED (tree `00a77a839`, fix absent):

```
go test -count=1 ./internal/cli/ -run 'TestFactoryMsgSendOptionalIDsDefault'
--- FAIL: TestFactoryMsgSendOptionalIDsDefault (1.28s)
    mcp_factory_msg_test.go:405: factory_msg_send(map[body:ids omitted idempotency_key:optional-ids-once kind:status_request run_id:optional-ids-run to_slot:lane-1]) = [{Annotated:{Annotations:<nil>} Meta:<nil> Type:text Text:factory_msg_send: invalid task reference}], want delivery
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	2.655s
```

GREEN (tree `50068c75d`):

```
go test -count=1 ./internal/cli/ -run 'TestFactoryMsgSendOptionalIDsDefault' -v
--- PASS: TestFactoryMsgSendOptionalIDsDefault (3.40s)
ok  	github.com/modu-ai/moai-adk/internal/cli	5.119s

go test -count=1 ./internal/cli/ -run 'FactoryMsg' -v
--- PASS: TestFactoryMsgStatusReadOnlyRoster (5.88s)
--- PASS: TestFactoryMsgSendRejectsClaudeOnlyRun (2.00s)
--- PASS: TestFactoryMsgSendOptionalIDsDefault (2.39s)
ok  	github.com/modu-ai/moai-adk/internal/cli	11.862s

go test -count=1 ./internal/factorymsg/...
ok  	github.com/modu-ai/moai-adk/internal/factorymsg	203.496s

go vet ./internal/factorymsg/... ./internal/cli/
vet_exit=0

gofmt -l internal/cli/mcp_factory_msg.go internal/cli/mcp_factory_msg_test.go internal/cli/mcp_server.go
(no output)
```

The test arms: omitted ids → default to the key; same-key retry → same message id; explicit `""` → default to the key; explicit `t1473` / `corr-1` → preserved.

Catalogue check: `grep -rn "correlation_id\|task_ref"` over `.claude/rules/moai/core/moai-mcp-tools-catalogue.md` and its template mirror returned no lines, so neither documents the fields as required; no doc change was made.

## Baseline-attribution

All commands above ran in this session in worktree `.claude/worktrees/agent-a46522ff7ab6e3500`, RED against `00a77a839`, GREEN and package runs against the working tree that became `50068c75d`. The GREEN `-v` run of the single test was taken before the explicit-empty arm was added; the later `-run 'FactoryMsg'` run (2.39s PASS) includes that arm.

## Gaps

- The RED commit message states that omitting only `correlation_id` yields "invalid correlation id". That follows from the check order at `store.go:784-788`; it was not run as a separate case.
- No full `./internal/cli/` package run (machine load 120-255; whole-package verdict is CI's).
- A broader `-run MCP` run failed on `TestCodexAuditMCPTool` (`codex_audit_mcp_test.go:195: start blocked for 2.95s`, a 900 ms wall-clock assertion) and once on `TestCodexTaskMCPEOFAbortsStalledHandshake`. Neither test touches `factory_msg_send`. With the base versions of the two changed files restored, `TestCodexAuditMCPTool` passed once; with the fix, `-count=2` showed one fail and one pass. Under that load the run is not conclusive either way, so this is recorded as a gap, not as a pass.
- No push; the repository-wide test verdict belongs to the CI run on `origin/develop` after the leader's batch push, and is PENDING.

## Residual-risk

- A caller that relied on an empty-id send being refused (as a validation probe) now gets a delivered message. Nothing in-tree does this (grep above).
- `task_ref` defaulting to the idempotency key makes `managed_factory_session.go`'s `task_ref=` display show the key instead of a task id when the sender omits it.
- Timing-sensitive tests in `internal/cli` (`TestCodexAuditMCPTool`) may show red on loaded runners; this is unrelated to this change, but it could hide a real regression until CI runs on a clean runner.
