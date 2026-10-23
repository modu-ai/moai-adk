---
id: SPEC-MCP-NEUTRALITY-001
title: "Research — verified code-path inventory (worktree t1632)"
version: "0.1.0"
created: 2026-10-14
updated: 2026-10-23
author: manager-spec
---

# SPEC-MCP-NEUTRALITY-001 — Research

Measured in this worktree at plan time (2026-10-14, HEAD ec193e4777, branch WT-3-2-0-3). Everything below is a verified file existence or grep observation from this tree, not carried over from the audit.

## §R1 Toolchain and project facts

- `quality.yaml`: `development_mode: tdd`, `test_coverage_target: 85`, `session_effort_default: xhigh`.
- `interview.yaml`: `decision_gate: on` — decision-index.md authored with this artifact set.
- Conversation language ko; SPEC artifacts follow the repo convention of English bodies (matches SPEC-ZONE-SHELL-PARSING-001 and siblings).

## §R2 Path-drift findings (audit numbers → this tree)

The 2026-10-10 audit cited paths that have moved since. Verified in this tree:

| Audit cited | Verified location | Evidence |
|---|---|---|
| `internal/cli/configtoml.go` (Codex env allowlist) | `internal/codexwiring/configtoml.go` | `ls` miss on the old path; grep hit `mcpServerEnvVarsValue` at the new path |
| `internal/hooks/stop.go` | `internal/hook/stop.go` (`internal/hooks` does not exist) | `ls` |
| `internal/messaging/store.go` (broker) | `internal/factorymsg/store.go` (`internal/messaging` does not exist) | `ls` |
| `internal/cli/factory_messages.go` | `internal/hook/factory_messages.go` | `ls` |

## §R3 Defect liveness in this tree (grep-verified)

- `todo_claim` registered in `internal/cli/mcp_todo.go` but ABSENT from `internal/mcp/catalog.go` (grep count 0 in the catalogue file) — dead-code defect live.
- `mcpServerEnvVarsValue` (internal/codexwiring/configtoml.go) forwards `MOAI_HOME, MOAI_KANBAN_ID, MOAI_SESSION_PID, MOAI_KANBAN_BACKEND, MOAI_FACTORY_WORKER, MOAI_FACTORY_WORKERS, CLAUDE_PROJECT_DIR, CLAUDE_CODE_SESSION_ID` — `MOAI_FACTORY_ROLE` and `MOAI_LAUNCH_PROVIDER` absent — allowlist defect live.
- `factoryLaneAdmission` / `factoryLaneRefusal` present in `internal/cli/factory_card.go` (absence-based role) — identity defect live.
- `codexAuditValidateRoot` in `internal/cli/codex_audit_launch.go` — the confinement pattern to generalize (REQ-MN-005).
- t1696 status: `git log -- internal/cli/factory_card.go` shows no t1696 commits; last touches are SPEC-CANDIDATE-CI-001 (t1478) and SPEC-MERGE-WINDOW-QUEUE-002 (t1582) merges. The D1 constraint is live.

## §R4 Files verified to exist (implementation surface)

`internal/cli/`: mcp_project_root.go, mcp_factory_msg.go, mcp_factory_card.go, factory_card.go, mcp_server.go, mcp_server_runtime.go, mcp_todo.go, mcp_claude.go, codex_jobs.go (superseded — surface removed by card t1909), glm_jobs.go, codex_audit_mcp.go, mcp_convergence.go, mcp_codex.go, mcp_glm.go, mcp_audit_multi_record.go, codex_review_scope.go, mcp_review_material.go, codex_audit_launch.go. `internal/`: auditreceipt/store.go, factorymsg/store.go, hook/stop.go, hook/factory_messages.go, mcp/catalog.go, codexwiring/configtoml.go, config/role_markers.go.

## §R5 Test-coverage state (from lens-7 §8, unverified in this tree)

Handler-level gaps reported by the audit (job control, `codex_role_audit*`, `glm_task`, graph tools, `jev_ask`, factory card verbs, `todo_*`, `factory_msg_list`/`body`/`receipt`) and protocol gaps (version negotiation, cancellation, wrong-typed arguments, stdout hygiene, schema lint) are the run phase's RED-first targets. Re-verify by grep before relying on any single claim from that section.
