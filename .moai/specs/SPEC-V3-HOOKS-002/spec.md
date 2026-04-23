---
id: SPEC-V3-HOOKS-002
title: "Hook Type System — 4 Types (command, prompt, agent, http)"
version: "0.1.0"
status: draft
created: 2026-04-22
updated: 2026-04-22
author: GOOS
priority: P1 High
phase: "v3.0.0 — Phase 6a Tier 2 Strategic Differentiators"
module: "internal/hook/types.go, internal/hook/exec_*.go"
dependencies:
  - SPEC-V3-HOOKS-001
  - SPEC-V3-SCH-001
related_gap:
  - gm#1
  - gm#2
  - gm#3
related_theme: "Theme 1: Hook Protocol v2 — Type System Expansion"
breaking: false
bc_id: null
lifecycle: spec-anchored
tags: "hook, v3, type-system, prompt, agent, http"
---

# SPEC-V3-HOOKS-002: Hook Type System — 4 Types

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-04-22 | GOOS | Initial v3 draft from Wave 4 bundle (Hooks/Commands) |

---

## 1. Goal (목적)

moai-adk-go supports only `type: 'command'` hooks today (shell wrappers). Claude Code defines four user-facing hook types — `command`, `prompt` (LLM-evaluated Haiku-class gate), `agent` (multi-turn subagent verifier), `http` (SSRF-guarded webhook) — all sharing the same rich-JSON reply surface defined in SPEC-V3-HOOKS-001. Adding the other three types unlocks low-cost LLM-gated quality checks (e.g., "does this SPEC make sense?"), subagent verifiers ("did tests actually run?"), and webhook integrations (notify Slack/Discord on commit). This SPEC introduces the type discriminator, per-type executors, and the SSRF guard port for HTTP hooks.

## 2. Scope (범위)

In-scope:
- Expand `HookDeclaration` (settings schema) with `type: command | prompt | agent | http` discriminator.
- Per-type fields:
  - `prompt` type: `prompt` (string), `model` (string default Haiku), `timeout` (seconds), `cost_budget_usd` (optional).
  - `agent` type: `prompt` (string), `model` (string), `timeout` (default 60s), `tools` (allowlist CSV), respects `ALL_AGENT_DISALLOWED_TOOLS`.
  - `http` type: `url` (validated URL), `headers` (map with env-var interpolation), `allowedEnvVars` (allowlist), `timeout`.
- Per-type executors: `internal/hook/exec_command.go` (existing, hardened), `internal/hook/exec_prompt.go`, `internal/hook/exec_agent.go`, `internal/hook/exec_http.go`.
- Port CC's `utils/hooks/ssrfGuard.ts` as `internal/hook/ssrf_guard.go` for `http` type.
- Schema validation for type-specific fields (required/forbidden combinations).
- Depth-cap of 2 for agent-type hooks (prevents recursion per master-v3 Theme 3.4).

Out-of-scope:
- Hook output protocol details — see SPEC-V3-HOOKS-001.
- `if` condition + matcher upgrade — see SPEC-V3-HOOKS-004.
- Async / once flags — see SPEC-V3-HOOKS-003.
- Plugin-sourced hooks — see SPEC-V3-PLG-001.
- Prompt/agent hook cost tracking beyond per-call `cost_budget_usd` (global budget deferred to a future cost SPEC, not in v3.0 scope).

## 3. Environment (환경)

Current moai-adk state:
- Shell-command-only hooks (findings-wave1-moai-current.md §5.5). 26 shell wrappers in `.claude/hooks/moai/`.
- `internal/hook/registry.go` dispatches via `EventType`; no `type` field is consulted today.
- No LLM-invocation primitives embedded in the hook path today; LLM calls happen only through Claude Code runtime.
- No HTTP client instrumented in `internal/hook/`; general HTTP calls use `net/http` stdlib in `internal/cli/` for feedback and updates.

Claude Code reference:
- `schemas/hooks.ts:32-65` — `command` type fields (findings-wave1-hooks-commands.md §2.1).
- `schemas/hooks.ts:67-95` — `prompt` type fields (findings-wave1-hooks-commands.md §2.2). Run via `utils/hooks/execPromptHook.ts:21-80` with Haiku-class model.
- `schemas/hooks.ts:128-163` — `agent` type fields (findings-wave1-hooks-commands.md §2.3). Run via `utils/hooks/execAgentHook.ts` multi-turn.
- `schemas/hooks.ts:97-126` — `http` type fields (findings-wave1-hooks-commands.md §2.4). SSRF guard at `utils/hooks/ssrfGuard.ts` (8,732 bytes).
- `utils/hooks/execHttpHook.ts:43-58` — URL + env-var allowlist (findings-wave1-hooks-commands.md §2.4).

Affected modules:
- `internal/hook/types.go` — add `HookType` enum and `HookDeclaration` struct.
- `internal/hook/exec_*.go` — four new or hardened executors.
- `internal/hook/ssrf_guard.go` — new file porting CC's SSRF logic.
- `internal/config/schema/hooks_schema.go` — validation for type-specific fields.
- `.moai/config/sections/system.yaml` — optional `policy_settings.allowed_http_hook_urls` allowlist.

## 4. Assumptions (가정)

- An LLM invocation primitive (Haiku-class) is addressable from within moai-adk-go via a thin wrapper that respects the user's active Claude Code session credentials; exact wiring is an implementation detail deferred to SPEC-level plan.
- CC's SSRF guard implements a standard private-network blocklist (10.0.0.0/8, 127.0.0.0/8, 169.254.0.0/16, 172.16.0.0/12, 192.168.0.0/16, fc00::/7, fe80::/10, ::1) which is the authoritative reference for the Go port.
- Agent-type hooks cannot spawn further agent-type hooks (depth cap of 2 means: parent calls agent hook, agent hook may spawn one child; child is leaf).
- `ALL_AGENT_DISALLOWED_TOOLS` exists or will be introduced as a policy constant shared with the agent hook executor.

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

- REQ-HOOKS-002-001: The `HookDeclaration` schema SHALL include a required `type` field with allowed values `command | prompt | agent | http`.
- REQ-HOOKS-002-002: The `command` hook type SHALL preserve full behavioral compatibility with v2 shell wrappers (stdin = JSON input, stdout/stderr = protocol output per SPEC-V3-HOOKS-001).
- REQ-HOOKS-002-003: The `prompt` hook type SHALL execute a single-turn LLM call with system prompt enforcing JSON response shape `{"ok": boolean, "reason"?: string}`.
- REQ-HOOKS-002-004: The `agent` hook type SHALL execute a multi-turn subagent query bounded by the `timeout` field (default 60 seconds).
- REQ-HOOKS-002-005: The `http` hook type SHALL POST the hook input JSON to the target URL and interpret the HTTP response body as `HookOutput` (per SPEC-V3-HOOKS-001).
- REQ-HOOKS-002-006: The system SHALL port Claude Code's SSRF guard to `internal/hook/ssrf_guard.go` and apply it to every `http` type hook before network I/O.
- REQ-HOOKS-002-007: The `http` hook executor SHALL sanitize all header values by stripping `\r`, `\n`, and `\x00` characters to prevent CRLF injection.
- REQ-HOOKS-002-008: The `http` hook executor SHALL interpolate env-var references in header values only when the referenced variable is listed in the hook's `allowedEnvVars` array.

### 5.2 Event-driven Requirements

- REQ-HOOKS-002-010: WHEN the hook type is `prompt`, the executor SHALL reject responses that are not valid JSON matching the `{ok, reason?}` contract and SHALL return an error with category `HookPromptInvalidResponse`.
- REQ-HOOKS-002-011: WHEN the hook type is `agent` and the agent tries to use a tool listed in `ALL_AGENT_DISALLOWED_TOOLS`, the executor SHALL deny the tool call and continue.
- REQ-HOOKS-002-012: WHEN the hook type is `http` and the SSRF guard rejects the URL, the executor SHALL abort the hook with `HookHttpSsrfRejected` and log the rejected URL (redacted to host portion only).
- REQ-HOOKS-002-013: WHEN the hook type is `http` and the response Content-Type is not `application/json`, the executor SHALL treat the hook as failed and surface a `HookHttpNonJsonResponse` error.

### 5.3 State-driven Requirements

- REQ-HOOKS-002-020: WHILE an `agent` type hook is running, the system SHALL count its depth in the agent spawn chain and SHALL refuse to spawn a further agent hook once depth reaches 2.
- REQ-HOOKS-002-021: WHILE `cost_budget_usd` is set on a `prompt` or `agent` type hook and the observed cost exceeds the budget, the system SHALL terminate the hook early with `HookCostBudgetExceeded`.

### 5.4 Optional Features

- REQ-HOOKS-002-030: WHERE the configuration key `policy_settings.allowed_http_hook_urls` is set in `.moai/config/sections/system.yaml`, the `http` hook executor SHALL refuse URLs not matching at least one entry in the allowlist, even if the URL passes the generic SSRF guard.
- REQ-HOOKS-002-031: WHERE a hook declaration specifies `model` for `prompt` or `agent` type, the system SHALL pass the value through to the LLM invocation layer; otherwise Haiku-class default is used.

### 5.5 Complex Requirements

- REQ-HOOKS-002-040: IF a `prompt` hook returns `{ok: false, reason: "..."}` on a PreToolUse event AND the user has not opted into policy-driven auto-block, THEN the system SHALL convert the response into a `HookSpecificOutput` with `permissionDecision: "ask"` and surface `reason` to the user; ELSE the response is converted to `permissionDecision: "deny"`.
- REQ-HOOKS-002-041: IF an `http` hook's target URL contains an env-var reference (e.g., `https://$NOTIFY_HOST/webhook`) AND the referenced variable is not in `allowedEnvVars`, THEN the executor SHALL reject the hook before any DNS resolution; ELSE the variable is interpolated before the SSRF guard runs.

## 6. Acceptance Criteria (수용 기준 요약)

- AC-HOOKS-002-01: Given a hook declaration with `type: prompt`, `prompt: "Is this SPEC valid? $ARGUMENTS"`, `model: haiku`, When a PreToolUse event fires, Then the executor issues a single LLM call and receives a `{ok, reason?}` JSON response. (maps REQ-HOOKS-002-003, REQ-HOOKS-002-010)
- AC-HOOKS-002-02: Given a hook declaration with `type: http`, `url: "http://127.0.0.1:8080/hook"`, When the SSRF guard runs, Then the hook is rejected with `HookHttpSsrfRejected` before any connection is attempted. (maps REQ-HOOKS-002-006, REQ-HOOKS-002-012)
- AC-HOOKS-002-03: Given a hook declaration with `type: http`, `headers: {Authorization: "Bearer $MY_TOKEN"}`, `allowedEnvVars: ["MY_TOKEN"]`, When executed, Then the header is interpolated before transmission and CRLF sanitization is applied. (maps REQ-HOOKS-002-007, REQ-HOOKS-002-008)
- AC-HOOKS-002-04: Given a hook declaration with `type: http`, `headers: {X-Custom: "$SECRET"}`, `allowedEnvVars: []`, When executed, Then the header value is transmitted literally as `$SECRET` (no interpolation) OR rejected if URL contains an un-allowlisted env-var reference. (maps REQ-HOOKS-002-008, REQ-HOOKS-002-041)
- AC-HOOKS-002-05: Given an `agent` hook running at depth 1, When it tries to dispatch another `agent` hook, Then the second spawn is refused with `HookAgentDepthExceeded`. (maps REQ-HOOKS-002-020)
- AC-HOOKS-002-06: Given `policy_settings.allowed_http_hook_urls: ["https://hooks.slack.com/*"]`, When a hook declares `url: "https://evil.example.com/x"`, Then the executor rejects the hook with `HookHttpUrlNotAllowlisted`. (maps REQ-HOOKS-002-030)
- AC-HOOKS-002-07: Given `cost_budget_usd: 0.01` on a `prompt` hook and observed cost reaches $0.01, When the next token is emitted, Then the executor terminates the hook with `HookCostBudgetExceeded`. (maps REQ-HOOKS-002-021)
- AC-HOOKS-002-08: Given a `prompt` hook returns `{"ok": false, "reason": "missing tests"}` on PreToolUse for a Write tool, When resolved, Then the system produces `HookSpecificOutput.permissionDecision = "ask"` with reason surfaced to the user. (maps REQ-HOOKS-002-040)

## 7. Constraints (제약)

- Technical: Go 1.22+. HTTP client uses `net/http` stdlib with custom `Transport.DialContext` for SSRF enforcement. No new top-level dependencies beyond SPEC-V3-SCH-001's validator/v10.
- Backward compat: v2 hooks are all implicit `type: command`. Schema default for missing `type` field SHALL remain `command` for one minor version.
- Platform: macOS / Linux / Windows. SSRF guard uses platform-neutral `net.IP` APIs.
- Security: No agent-type hook may exceed depth cap of 2. No http-type hook may bypass SSRF guard. No prompt-type hook may exceed per-call cost budget if set.
- Performance: Prompt-type hooks SHOULD complete within 5s p95 at Haiku class; agent-type hooks bounded by `timeout` field.

## 8. Risks & Mitigations (리스크 및 완화)

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Prompt/agent hooks incur surprise API spend | M | H | Per-hook `cost_budget_usd` enforced by REQ-HOOKS-002-021; documentation emphasizes budget-first defaults; `moai doctor hook --cost-estimate` (deferred SPEC) will surface projected costs. |
| SSRF guard port drifts from CC upstream | L | H | SSRF guard is faithful 1:1 port of `utils/hooks/ssrfGuard.ts`; unit tests against the same private-network blocklist fixture used upstream. |
| Agent hook runaway recursion | M | H | Depth cap 2 enforced by REQ-HOOKS-002-020; `moai doctor hook --validate` detects declarations that would exceed depth statically. |
| HTTP timeout leaks goroutines | L | M | Every HTTP hook uses `context.WithTimeout` bound to `timeout` field; executor closes response body in `defer`. |
| Header CRLF injection via env interpolation | L | H | REQ-HOOKS-002-007 mandates strict sanitization BEFORE transmission, even for allowlisted env vars. |

## 9. Dependencies (의존성)

### 9.1 Blocked by

- SPEC-V3-HOOKS-001 (Hook output protocol must be in place before type-specific executors consume it).
- SPEC-V3-SCH-001 (validator/v10 and schema registration for type-specific fields).

### 9.2 Blocks

- SPEC-V3-PLG-001 (plugin manifests may declare `prompt`/`agent`/`http` hooks; plugin validation relies on this schema).

### 9.3 Related

- SPEC-V3-HOOKS-004 (matcher + if condition applies uniformly to all four types).
- SPEC-V3-HOOKS-003 (async/once flags apply uniformly to all four types).

## 10. Traceability (추적성)

- Theme: master-v3 Section 3.1 (Theme 1 — Hook Protocol v2) and the Phase 6a "Tier 2 Strategic Differentiators" phase in §6.1.
- Gap rows: gm#1 (High — type:prompt), gm#2 (High — type:agent), gm#3 (Medium — type:http).
- BC-ID: None (purely additive).
- Wave 1 sources: findings-wave1-hooks-commands.md §2.1-2.4 (type taxonomy), §12 (SSRF guard reference), §14.2 (missing hook types table).
- Priority: P1 High (unlocks LLM-gated and webhook integrations; ships Phase 6a).
