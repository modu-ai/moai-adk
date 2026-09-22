---
id: SPEC-FACTORY-MIXED-HOOK-001
document: research
created: 2026-09-22
updated: 2026-09-22
author: manager-spec
card: t1074
module: "internal/factorymsg"
---

# Research — SPEC-FACTORY-MIXED-HOOK-001

## Measured repository facts

- `internal/cli/factory.go`: lead sets run/address environment; worker currently sets worker labels without equivalent run binding.
- `internal/kanban/factory_slots.go`: worker claim persists label/PID/time but not the complete run/backend/session/generation identity already representable by homestate factory records.
- `internal/hook/session_start_record.go`: current role parsing recognizes `lane-N`; the `agent-N` worker label requires coverage.
- `internal/cli/mcp_session_msg.go`: legacy store root is project-directory-relative, so linked worktrees need a new canonical factory namespace rather than silent legacy migration.
- `internal/sessionmsg.Store.Poll`: legacy `ack_ids` deletion lacks recipient generation and claim-token comparison. It is not safe as the factory receipt contract.
- `.claude/settings.json` and its template currently mark Stop async; blocking continuation requires synchronous Stop wiring.
- Codex hook wiring already routes UserPromptSubmit and Stop through `moai hook --harness codex`; adapter behavior must be verified against the installed version before claiming parity.

## Codex worktree and cwd findings

- Official-source refresh in this run resolved `openai/codex` `main` to `d1092865f8ec63735006211f65ee109ce91c30b9`; the source observations below are pinned to that commit.
- Installed Codex CLI `0.155.1` exposes `/cd`, `/pwd`, and `/worktree` in the TUI. An isolated live probe changed the displayed cwd from `/Users/goos/MoAI/moai-adk-go` to `.claude/worktrees/t1074` and the statusline branch from `main` to `WT-factory-mixed-hook`.
- The same probe changed the Codex session UUID from `01a0c733-e96a-77c2-b38f-fab5f01ffb9e` to `01a0c734-8deb-7431-ba6b-8b578cfb7c1a`. Therefore the user-visible TUI/history can continue while the physical thread endpoint rotates; a factory lane cannot use the thread UUID as its stable address.
- Official `TurnStartParams.cwd` overrides the working directory for that turn and subsequent turns. The official Python SDK documents `cwd` for thread start/run, resume, and fork. Official worktree guidance also supports local↔worktree handoff, while managed Codex worktrees are typically detached-HEAD and therefore do not replace MoAI's card worktree policy.
- The official `thread/queue/add` request has no cwd field. Queueing a message is not a supported way to switch an existing TUI lane's work root.
- The resulting lifecycle belongs to `t1082`; this SPEC retains only stable-lane/current-endpoint addressing and stale-generation rejection.

## Official contract boundaries

- Codex hooks: UserPromptSubmit additional context is elevated developer context; async hook completion does not wake an idle turn.
- Claude hooks: Stop can request continuation and supplies loop-prevention state; continuation must be bounded.

Sources:

- https://learn.chatgpt.com/docs/hooks
- https://github.com/openai/codex/blob/main/codex-rs/app-server-protocol/schema/typescript/v2/TurnStartParams.ts
- https://github.com/openai/codex/blob/main/sdk/python/docs/api-reference.md
- https://github.com/openai/codex/blob/main/codex-rs/tui/src/slash_command.rs
- https://github.com/openai/codex/blob/main/codex-rs/tui/src/session_queue_commands.rs
- https://learn.chatgpt.com/docs/environments/git-worktrees
- https://code.claude.com/docs/en/hooks

## Baseline checks observed in this run

```text
GOCACHE=/tmp/t1074-sessionmsg-cache go test ./internal/sessionmsg -count=1
ok github.com/modu-ai/moai-adk/internal/sessionmsg 1.983s

MOAI_HOME=/tmp/t1074-baseline-home GOCACHE=/tmp/t1074-baseline-cache go test ./internal/cli -run 'TestResolveFactoryWorkerName|TestCC_FactoryEntryThroughRunCC|TestGLM_FactoryWorkerEntry|TestCodexFactory' -count=1
ok github.com/modu-ai/moai-adk/internal/cli 4.991s
```

The broader hook/package attempts without an isolated MoAI home touched the real-home path and failed setup or existing SessionStart tests; those attempts are environment-invalid baselines, not product defects and not PASS evidence.

## Open research gaps

- Installed Claude/Codex hook trust and exact emitted JSON require live measurement.
- Real model-context delivery, continuation, fault injection, and performance are NOT_RUN at plan time.
- Same-UID local processes are not a strong security boundary; identity binding prevents mistakes and stale ownership, not a hostile local-user sandbox.
- Official interfaces do not expose a hook/MCP operation that executes a TUI slash command. Interactive handoff still requires an idle `/cd`; unattended handoff must use an official headless thread/fork/start `cwd` path and register the resulting endpoint.
