---
id: SPEC-MCP-NEUTRALITY-001
title: "Implementation plan — MCP refusal structuring and harness neutrality"
version: "0.1.0"
created: 2026-10-14
updated: 2026-10-23
author: manager-spec
---

# SPEC-MCP-NEUTRALITY-001 — Implementation Plan

## §A Context

- **Worktree**: the card worktree this SPEC lives in (branch `WT-3-2-0-3`, base = local `main` @ ec193e4777 at plan time). All work happens here; the lane commits at stage boundaries.
- **Artifacts**: `.moai/specs/SPEC-MCP-NEUTRALITY-001/{spec,plan,acceptance,design,research,decision-index,progress}.md` (Tier L).
- **Source of scope**: master plan stage-0 row 3.2-0-3 (queue card t1632, P0, Class C), absorbing cards t1622, t1624, t1625, and the t1597 MCP/state-path seam. Evidence: `.moai/reports/audit-20261010/lens-{3,4,6,7}-*.md`; design: the cross-harness redesign §6.3 (목표 MCP 모델).
- **Methodology**: TDD (`quality.yaml` `constitution.development_mode: tdd`) — RED-GREEN-REFACTOR per requirement cluster; coverage target 85% (`test_coverage_target`).
- **Existing infrastructure**: PRESERVE — the MCP registration gate and catalogue mechanism (`internal/cli/mcp_server.go` `add` + `internal/mcp/catalog.go`), `codexAuditValidateRoot` / `codexChildEnv` / `codexAuditMCPDisableArgs` (`internal/cli/codex_audit_launch.go`) as the containment patterns to reuse, `internal/auditreceipt/store.go` as the receipt store to extend, `internal/factorymsg/store.go` as the broker store to extend. EXTEND — the refusal path (`internal/cli/mcp_factory_card.go`, `mcp_todo.go`), the root resolver (`internal/cli/mcp_project_root.go`), the broker claim path (`internal/factorymsg/store.go`), the hook inbox path (`internal/hook/factory_messages.go`, `internal/hook/stop.go`).

## §B Known Issues (defect-to-requirement map)

Anchors verified in this worktree at plan time (`research.md` §R2 carries the full inventory). Line numbers omitted deliberately — they drift.

| # | Defect (source) | Anchor | REQ |
|---|-----------------|--------|-----|
| B1 | Omitted/mistyped `project_root` silently acts on the server's spawn tree; non-string coerced (lens-7 B1, t1624-1) | `resolveToolProjectRootWithSource`, `internal/cli/mcp_project_root.go` | MN-004/005 |
| B2 | Root validation accepts any `.moai`-bearing directory incl. `$HOME` (lens-7 B2, t1624-2) | `validateProjectRoot`, same file | MN-005 |
| B3 | ~23 tools take no `project_root`; no `_root` annotation outside catalogue+verify tools (lens-7 B1, §6.3 row 2) | `session_list`, `goal_*`, `audit_cache`, `codex_job_*` (superseded — surface removed by card t1909), `glm_*`, `session_msg_*`, `factory_msg_*`, `jev_ask` handlers | MN-004 |
| B4 | `todo_claim` registered but dead — absent from catalogue; gate drops silently (lens-7 B4, t1624-7) | `internal/cli/mcp_todo.go` + `internal/mcp/catalog.go` | MN-006 |
| B5 | Role decided by env ABSENCE → unmarked Codex session reads as operator (lens-7 B5, t1624-3) | `factoryLaneAdmission`/`factoryLaneRefusal`, `internal/cli/factory_card.go` | MN-003 |
| B6 | Codex env allowlist omits `MOAI_FACTORY_ROLE`, `MOAI_LAUNCH_PROVIDER` (lens-7 B6, t1624-4) | `mcpServerEnvVarsValue`, `internal/codexwiring/configtoml.go` | MN-003 |
| B7 | Policy refusals raised as `isError` (482/1,943); lessons observer wrote 409 refusal rows (§6.3 row 3) | refusal paths in `mcp_factory_card.go`, `mcp_todo.go`; failure observer | MN-001/002 |
| B8 | Stale run ids accepted; lease/state read from expired rows (§6.3 row 5) | factory verb run-id validation | MN-003 |
| B9 | Two claimers per endpoint, 30 s lease, token rotation on every claim → "claim identity mismatch" (lens-7 B3, lens-6 §5, t1622-2) | `Store.Claim`, `internal/factorymsg/store.go`; hook claim in `internal/hook/factory_messages.go`; `factory_msg_list` in `internal/cli/mcp_factory_msg.go` | MN-017 |
| B10 | Generation change orphans messages (9 real orphans) (lens-4 §5, t1622-1) | generation re-register path, `internal/factorymsg/store.go` | MN-016 |
| B11 | `factory_msg_list` zero targets → `"messages": null` (lens-7 B3, t1622-3) | handler serializing `nil, nil` | MN-017 |
| B12 | 2,048-byte cut creates claimed-but-invisible messages (lens-6 §5, t1622-4) | `factoryHookContextLimit`, `internal/hook/factory_messages.go` | MN-016 |
| B13 | Leader Stop blocks every natural turn (lens-6 §5, t1622-5) | `internal/hook/stop.go` block decision | MN-018 |
| B14 | 200 ms deadline + write-lock probe on the empty-inbox path → 1,619 deadline-exceeded (t1622-6) | `CheckWritable` path, hook inspection | MN-011 |
| B15 | Stale session id after `/clear` (t1622-7) | `currentFactoryPeer`, `internal/cli/mcp_factory_msg.go` | MN-017 |
| B16 | `structuredContent` filled but content is `"<tool>: ok"`; arrays null (lens-7 B8, t1624-6) | `toolJSON`, `internal/cli/mcp_server.go` | MN-008 |
| B17 | Stale server advertised only at initialize (lens-7 B9, t1624-5) | build stamp path, `internal/cli/mcp_server.go` | MN-013 |
| B18 | Stamp write fails closed; `.moai` created in any cwd (lens-7 B10, t1624-8) | `writeMCPServerRuntimeRecord`, `internal/cli/mcp_server_runtime.go` | MN-015 |
| B19 | Schemas: additionalProperties unset, 11 undocumented properties, undeclared `codex_task.model` (superseded — surface removed by card t1909), `goal_status` required mismatch, wrong hints (lens-7 schema section, t1624-9) | every tool registration | MN-007 |
| B20 | `notifications/message` without logging capability; fixed-Korean progress phrases (lens-7 B22, t1624-10) | `notifyMCPProgress` | MN-008 |
| B21 | Codex gate sessions: 38/38 MCP call failures, elicitation unimplemented, stdin double-write; foreign MCP servers active (§6.3 row 7, t1625 scope) | codex child launch paths | MN-009/010 |
| B22 | `claude_audit` denylist env scrub; `CLAUDE_CONFIG_DIR` + proxy pass through; 429 misread; binary pin bypassed (lens-3 #8, t1625-5) | `scrubClaudeAuditEnv`, `claudeLookPath`, `internal/cli/mcp_claude.go` + `claude_binary.go` | MN-012 |
| B23 | Codex audit/review legs: no sandbox policy, full env inheritance (lens-3 #9, t1625-6) | `buildCodexReviewParams`, `exec.CommandContext` sites, `internal/cli/mcp_codex.go` | MN-022 |
| B24 | `audit_multi` codex leg + adversarial mode ignore the target; second-review record claims baseBranch (lens-3 #2, t1625-1) | `performCodexAudit`, `appendSecondReviewRecord` | MN-019 |
| B25 | Receipts not tree-bound; codex FAIL reinforces PASS (lens-3 #3, t1625-2) | `internal/auditreceipt/store.go` | MN-020 |
| B26 | Transport/format failure → verdict `fail` (lens-3 #10, t1625-3) | `applyGateUnmet`, `enforceRequiredGateUnmet` | MN-021 |
| B27 | Converge passes on caller-submitted verdict when required backend inconclusive (lens-3 #13, t1625-4) | `claudeVerdictOrDefault` fallback, `internal/cli/mcp_convergence.go` | MN-021 |
| B28 | Diff truncation/untracked exclusion unmarked; no empty-material guard in `codex_audit` (lens-3 #11/12, t1625-7) | `collectReviewDiff`, `internal/cli/mcp_review_material.go` | MN-023 |
| B29 | Job lifetime: GLM/role jobs not cancelled on exit, no liveness, no role-audit cancel, foreground timeout loses thread (lens-3 #16/17, t1625-8; the codex foreground/background parts: superseded — surface removed by card t1909) | `codex_jobs.go` (superseded — surface removed by card t1909), `glm_jobs.go`, `codex_audit_mcp.go` | MN-014 |
| B30 | Gate state: per-session overwrite, non-atomic writes, unvalidated `session_id` path join, inconclusive reused as pass for 24 h (lens-3 #7/4, t1625-10) | `persistConvergenceResult`, review cache | MN-024 |
| B31 | `parseGLMReview` unmarshals whole model text; `glm_task` `Content[0]`; no model provenance on codex/GLM (lens-3 #21/23/20, t1625-11/12) | `parseGLMReview`, `internal/cli/glm_task.go` | MN-021/020 |
| B32 | `codex_audit` no deadline; 5-worker pool starvation (lens-3 #18/11, t1625-9) | `handleCodexAudit` request-context pass-through | MN-014 (bound via job lifetime) |
| B33 | State-root disagreement: receipt producer vs Codex Stop chain reader in nested sessions (t1597 reviewer overlay reproduction) | receipt writer root vs `codexReviewMember` read root | MN-025 |
| B34 | `job_id` path-concatenation unvalidated (lens-3 #26, t1625-11) | `codex_jobs.go` (superseded — surface removed by card t1909), `glm_jobs.go` join sites | MN-024 (same validation class) |
| B35 | Server logs stderr-only; no per-call record (§6.3 row 11) | server logging | MN-015 |
| B36 | Audit caches PASS with no verdict-file evidence; process-local only (§6.3 row 6) | `audit_cache` store path | MN-020 (disk persistence + verdict-file existence/SHA validation clause) → AC-MN-023 |

## §C Pre-flight

Run before any code change (all read-only; from the worktree root):

```bash
git branch --show-current
git rev-parse --short HEAD
go build ./...
GOOS=windows GOARCH=amd64 go build ./...
grep -c "todo_claim" internal/mcp/catalog.go
grep -n "MOAI_FACTORY_ROLE" internal/codexwiring/configtoml.go
git log --oneline -3 -- internal/cli/factory_card.go
```

Expected at plan time: branch `WT-3-2-0-3`; `todo_claim` catalogue count 0 (defect live); no `MOAI_FACTORY_ROLE` in the allowlist (defect live); `factory_card.go` log shows no t1696 commits yet. If t1696 has landed by run time, rebase the §D constraint accordingly.

## §D Constraints

- **D1 — t1696 serialization.** `internal/cli/factory_card.go` is being repaired in parallel by card t1696. Scope identity work to the MCP layer (`internal/cli/mcp_factory_card.go`, `internal/codexwiring/configtoml.go`) wherever possible. Any change that must touch `internal/cli/factory_card.go` is staged last and merges/absorbs t1696 after t1696 lands (lane board conditional task). Re-check `git log -- internal/cli/factory_card.go` before touching it.
- **D2 — No line-number citations.** Re-locate every anchor by symbol before editing; the audit's numbers predate this tree.
- **D3 — Closed refusal code set.** `{serial-slot, owned, raced, parent-checkout, stale-run}` is closed; a new code requires a SPEC amendment, not an ad-hoc string.
- **D4 — GEARS only.** No `IF/THEN` in new artifacts (lint emits `LegacyEARSKeyword`).
- **D5 — One writer per tree; lane commits at stage boundaries.** No `git add -A`; stage by explicit pathspec.
- **D6 — MCP reconnect.** After reinstalling the moai binary during run-phase verification, reconnect the host's MCP server before testing new tools; the running server never hot-reloads.
- **D7 — Security surfaces are exempt from the simplicity ladder.** Path-join validations, root confinement, env allowlists.
- **D8 — Verification scope.** Run the affected test families (`go test ./internal/cli/... -run '^<family>$'`, `./internal/factorymsg/...`, `./internal/hook/...`, `./internal/auditreceipt/...`), not the full suite (card-worktree full-suite runs are structurally red per recorded lesson; CI is the verdict surface). Concurrency tests judged with `-race` and `-count` repeats.

## §E Self-Verification

E1 AC PASS/FAIL matrix (25 ACs, command + verbatim output each). E2 cross-platform build (`go build ./...`, `GOOS=windows GOARCH=amd64 go build ./...`). E3 coverage on the touched packages (`go test -cover ./internal/cli/... ./internal/factorymsg/... ./internal/hook/... ./internal/auditreceipt/...` against the 85% target; per-change-scope, full-matrix judgment belongs to CI). E4 subagent-boundary grep (no AskUserQuestion in `internal/cli`, `internal/hook`, `internal/factorymsg`). E5 lint (`golangci-lint run --timeout=2m`, NEW vs baseline separated). E6 branch HEAD + commit list. E8 RED failure output per requirement cluster (TDD — verbatim pre-GREEN evidence). Each E-item carries the attribution triple: command, verbatim output, HEAD SHA.

Card completion criteria (the three terminal gates):
1. 정책 거부 is_error 0 — AC-MN-001 (plus every refusal-path test asserting `is_error` false).
2. lessons-inbox mcp 신규 0 — AC-MN-002.
3. Codex 게이트 호출 성공률 측정 — AC-MN-003 (a measurement exists and reports a number; zero measurement = FAIL).

## §F Milestones

Ordered by decision-reversibility and the stage-0 P0 ordering (the refusal/result contract shapes every later milestone, so it lands first; mechanical cleanups last). Each Milestone is independently verifiable.

- **M1 — Structured refusals + identity (P0)** (REQ-MN-001/002/003). Refusal result shape (`design.md` §D1) across factory/todo paths; observer split; actor-as-argument with env-as-default; stale-run refusal; positive operator mark. Files: `internal/cli/mcp_factory_card.go`, `mcp_todo.go`, the factory message handlers, the failure observer. Verify: refusal-path tests assert is_error false + codes (the `TestTodoClaimMCP_Mirror` refusal arm among them, updated to the structured shape); observer test with forced refusals; AC-MN-001..005. RED-first: the current paths raise `isError` — the EV-001 measurement is the pre-GREEN evidence.
- **M2 — Root contract + catalogue + schemas (P0)** (REQ-MN-004..008). Common `project_root` on the eight families; confinement via the `codexAuditValidateRoot` pattern; `_root` annotation everywhere; factory verbs' lockRoot through the shared CLI resolver; `todo_claim` catalogue row + fail-closed registration; schema strictness + lint test; wire honesty (content mirror, null-free arrays, capability honesty, neutral phrases). Files: `mcp_project_root.go`, `mcp_server.go`, `internal/mcp/catalog.go`, `mcp_todo.go`, the eight family handlers, `mcp_factory_msg.go` properties. Verify: root tests (42 → rejected; `$HOME` → rejected; worktree → accepted), tools/list contains `todo_claim`, schema lint green, AC-MN-006..010.
- **M3 — Gate session containment (P0)** (REQ-MN-009/010/011). `codexAuditMCPDisableArgs` applied to every codex gate/audit child; the gate-session read-only subset — enumerated: `spec_progress`, `spec_audit`, `spec_drift`, `verify_trend`, `session_list`, `audit_cache` (read op only), `goal_status`, the four graph read tools, `todo_list`, `jev_ask`, `session_msg_list`, `factory_msg_status`, `codex_setup`, and the job status/result tools — derived by the REQ-MN-009 rule and asserted equal by the exposure test; elicitation never-initiate guard + gate-client auto-decline; write mutex; empty-inbox counting path lock-free. The M3 measurement harness: a gate-session-shaped client issues the fixture call set (initialize, tools/list, one call per subset tool) and prints the AC-MN-003 metric line. Files: codex child launch paths, server request handling, `CheckWritable` path, hook inbox read. Verify: launch-argv test, exposure-set test, elicitation guard + decline test, mutex `-race` test, empty-path deadline test, the success-rate measurement reporting a number, AC-MN-011..013.
- **M4 — Env scrub + staleness + jobs + hygiene (P0/P1)** (REQ-MN-012..015). Allowlist child env for claude; per-call `_server` + `stale_server` refusal; job lifetime (cancel on exit, CommandContext, owner stamp, stale status, role-audit cancel, TTL, foreground thread + partial output [superseded — surface removed by card t1909]); stamp best-effort + no foreign `.moai`; rotated call log. Files: `mcp_claude.go`, `claude_binary.go`, `mcp_server.go`, `mcp_server_runtime.go`, `codex_jobs.go` (superseded — surface removed by card t1909), `glm_jobs.go`, `codex_audit_mcp.go`. Verify: env tests, staleness test (mutated mtime), job-lifetime tests, stamp test on a read-only/non-project cwd, AC-MN-014..017.
- **M5 — Broker + inbox repairs, absorbed t1622 (P0)** (REQ-MN-016/017/018). Generation migration/dead-letter + sender notify; token-stable re-claim; list zero-shape; post-clear live identity; line-boundary cut + visible-only claiming; Stop block scoping. Files: `internal/factorymsg/store.go`, `internal/hook/factory_messages.go`, `internal/hook/stop.go`, `internal/cli/mcp_factory_msg.go`. Verify: store tests (generation, token, zero-shape), hook tests (cut boundary, block scope), AC-MN-018..020.
- **M6 — Audit integrity + state-root seam, absorbed t1625 + t1597 (P0/P1)** (REQ-MN-019..025). Target honesty; receipt tree-binding + provenance parity; verdict truthfulness (could-not-run vs content-fail, one retry, stderr preserved; converge fail-closed; narrow GLM decode); codex child containment; material marking + empty guard; gate state integrity; state-root seam contract. Files: `mcp_convergence.go`, `mcp_codex.go`, `mcp_glm.go`, `glm_task.go`, `mcp_audit_multi_record.go`, `internal/auditreceipt/store.go`, `mcp_review_material.go`, the seam resolvers. Verify: receipt tests (FAIL non-reinforcement, digest mismatch), transport-failure test, converge test, containment tests, nested-worktree seam fixture (the reviewer overlay reproduction), AC-MN-021..025.

## §G Anti-Patterns

- Converting a policy refusal into an MCP error "because it is simpler" — the entire point of M1.
- Broadening the refusal code set ad hoc (D3).
- Touching `internal/cli/factory_card.go` without re-checking t1696 status (D1).
- Absorbing the 3.2-2-1 state unification or the two-base-resolver defect (spec §5).
- Claiming a verification without the command + verbatim output (verification-claim integrity); a green whose swept set is empty is not a pass.
- Editing `progress.md` §E.2–E.4 or SPEC bodies at run/sync phases (ownership matrix).
- Trusting a stale MCP server for verification — reconnect after reinstall (D6).

## §H Cross-References

- spec.md §3 (25 REQs), acceptance.md (25 ACs), design.md §D1–D6 (target contracts), research.md §R1–R5 (verified inventory), decision-index.md (2 rows).
- Cards: t1632 (this card), t1622, t1624, t1625, t1597 (absorbed), t1696 (parallel repair, D1).
- Evidence: `.moai/reports/audit-20261010/lens-{3,4,6,7}-*.md`; design: cross-harness redesign §6.3, master plan stage-0 row 3.2-0-3.
- Related SPECs: SPEC-MCP-WORKTREE-ROOT-001 (introduced `project_root` handling this SPEC tightens), SPEC-EVIDENCE-CLAIM-INVARIANT-001 (owns the lessons-inbox failure-row semantics this SPEC splits).
