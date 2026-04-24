---
id: SPEC-V3R2-RT-002
title: "Permission Stack + Bubble Mode"
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: GOOS
priority: P0 Critical
phase: "v3.0.0 — Phase 2 — Runtime Hardening"
module: "internal/permission/"
dependencies:
  - SPEC-V3R2-CON-001
  - SPEC-V3R2-RT-001
  - SPEC-V3R2-RT-005
bc_id: []
related_principle: [P6 Permission Bubble, P7 Sandbox Default, P8 Hook JSON]
related_pattern: [S-1, S-2, T-5]
related_problem: [P-C01, P-C04]
related_theme: "Layer 3: Runtime"
breaking: false
lifecycle: spec-anchored
tags: "permission, bubble, provenance, v3r2, breaking, runtime, safety"
---

# SPEC-V3R2-RT-002: Permission Stack + Bubble Mode

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-04-23 | GOOS | Initial v3 Round-2 draft. New SPEC — no v3-legacy predecessor. Addresses P-C01 (no permission bubble, CRITICAL) head-on. |

---

## 1. Goal (목적)

Replace moai's flat `permissions.allow` list with a typed 8-source permission stack that resolves every tool invocation through an ordered precedence chain and carries provenance (which file set this rule?) to every downstream consumer. Introduce `bubble` as a first-class `PermissionMode` value — not a boolean special case — so fork agents that inherit parent context can escalate permission decisions to the parent terminal's user via AskUserQuestion rather than silently default-allow or deny in an isolated mailbox. This is the structural foundation for Safety Architecture: without a permission envelope, sandboxed execution (SPEC-V3R2-RT-003) cannot target the right processes, multi-source settings (SPEC-V3R2-RT-005) cannot route provenance, and hook-driven input mutation (SPEC-V3R2-RT-001) cannot be audited.

Master §4.3 Layer 3 type block names the 5 canonical `PermissionMode` values (`default`, `acceptEdits`, `bypassPermissions`, `plan`, `bubble`) and the 8 `Source` tiers (policy > user > project > local > plugin > skill > session > builtin). Master §5.2 declares this as the cross-layer concern threading Layer 3 (runtime enforcement), Layer 4 (agent `permissionMode`), and Layer 7 (plugin-contributed rules). Master §8 BC-V3R2-015 (multi-layer settings resolution) formalizes the reader change: v2's flat merge remains readable, but v3's resolution answer is tier-aware. RT-002 itself is additive (adds the bubble mode + tier reader) and is therefore non-breaking (bc_id: []).

## 2. Scope (범위)

In-scope:

- `Source` enum in `internal/permission/stack.go` with the 8 canonical tiers from master §4.3 Layer 3 type block.
- `PermissionMode` string enum including `bubble` as an equal peer to `default`, `acceptEdits`, `bypassPermissions`, `plan`.
- `PermissionRule` Go struct carrying `Pattern string`, `Action PermissionDecision`, `Source Source`, `Origin string` (file path for provenance).
- `PermissionResolver.Resolve(tool, input, ctx) (HookResponse, error)` walking the 8-source stack in priority order and returning the first non-empty decision, plus optional `updatedInput`.
- Bubble-mode semantics: when a fork agent (spawned by a parent terminal session) encounters a non-allowlisted tool, the resolver returns `PermissionDecision: "ask"` and enqueues the prompt via AskUserQuestion at the parent terminal, not the teammate's mailbox.
- PreToolUse hook integration: resolvers consult hook responses (SPEC-V3R2-RT-001) as the top priority within the `hookDecision` sub-tier that sits above `session` in the 8-source ordering per master §5.2 "policy > user > project > local > plugin > skill > session > builtin" — the hook contribution is overlaid on top of `session` by feeding into the session-scope rules.
- Pre-allowlist for common dev ops: shipped default list covering `Bash(go test:*)`, `Bash(golangci-lint run:*)`, `Bash(ruff check:*)`, `Bash(npm test:*)`, `Bash(pytest:*)`, `Read(*)`, `Glob(*)`, `Grep(*)` to avoid bubble-fatigue on read/verify operations.
- `moai doctor permission` subcommand that prints the merged stack with per-rule provenance and a worked example of how a candidate `(tool, input)` pair would resolve.
- Agent frontmatter `permissionMode` field becomes a strictly-validated enum; CI lint rejects unknown values.

Out-of-scope (addressed by other SPECs):

- Hook JSON protocol wire format — SPEC-V3R2-RT-001.
- Sandbox execution backends — SPEC-V3R2-RT-003.
- Settings file provenance reader + multi-layer merge — SPEC-V3R2-RT-005.
- Typed session state for resumable permission context — SPEC-V3R2-RT-004.
- Plugin-origin settings ingestion — deferred to v3.1+ per master §7 plugin system (X-4 NOT-NOW).
- UI prompt rendering for bubble (AskUserQuestion owns the UX).

## 3. Environment (환경)

Current moai-adk state:

- `.claude/settings.json` exposes a flat `permissions.allow` list with no provenance tracking per r3-cc-architecture-reread.md §4 Adopt 2 (moai-adk today has implicit trust).
- Agent frontmatter supports `permissionMode: default | acceptEdits | bypassPermissions | plan` — no `bubble` value exists per master §4.3 Layer 3 code block comparison.
- No resolver walks settings tiers; whatever file last wrote the key wins silently (r3 §4 Adopt 1).
- 6 agents missing `isolation: worktree` per problem-catalog.md P-A11; permission-without-sandbox is the v2.13.2 baseline.

Claude Code reference:

- CC permission modes: `default`, `acceptEdits`, `bypassPermissions`, `plan`, and crucially `bubble` for fork agents (r3 §2 Decision 15).
- CC settings tiers: `policySettings > userSettings > projectSettings > localSettings > pluginSettings > sessionRules > hookDecision → allow | ask | deny + updatedInput?` (r3 §1.3 hooks settings precedence).
- Bubble rationale: "a fork that inherits context SHOULD ask the parent-terminal's user for permission (bubble), not the teammate's mailbox" (r3 §2 Decision 15).

Affected modules:

- `internal/permission/stack.go` — new file, Source/PermissionRule types.
- `internal/permission/resolver.go` — new file, 8-source walker.
- `internal/permission/bubble.go` — new file, AskUserQuestion dispatch for bubble mode.
- `internal/cli/doctor.go` — add `permission` sub-subcommand.
- `.claude/agents/moai/*.md` frontmatter — accept `bubble` mode.
- `.moai/config/sections/security.yaml` — new section `permission.pre_allowlist` listing default dev-op patterns.
- CI lint — extend `internal/template/commands_audit_test.go` (or new test) to reject unknown permissionMode.

## 4. Assumptions (가정)

- Claude Code's bubble mode contract remains stable for 2.1.111+ (r3 §2 Decision 15 evidence).
- The 8 tiers cover moai's needs; no immediate requirement for a 9th tier (plugin marketplace is deferred X-4).
- AskUserQuestion is only invoked by the MoAI orchestrator per constitution agent-common-protocol.md; bubble mode dispatches back to the orchestrator context rather than the teammate context.
- Pre-allowlist of read/verify ops covers ~80% of bubble-fatigue scenarios per r3 §4 Adopt 2 "pre-allowlist for common dev ops".
- Hook JSON (SPEC-V3R2-RT-001) is available; `PermissionDecision` enum values `allow | ask | deny` are shared verbatim between protocol and resolver.
- Plugin-contributed permission rules (pluginSettings tier) exist as a schema slot only; no plugins actually contribute rules in v3.0 per master §7 plugin NOT-NOW.
- Existing `.claude/settings.json` flat `permissions.allow` lists upgrade cleanly to tier-annotated form via the migrator (SPEC-V3R2-MIG-001); all v2 rules inherit `Source: SrcProject` or `Source: SrcLocal` depending on file location.

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

- REQ-V3R2-RT-002-001: The `Source` type SHALL be a typed enum with exactly 8 values in priority order: `SrcPolicy`, `SrcUser`, `SrcProject`, `SrcLocal`, `SrcPlugin`, `SrcSkill`, `SrcSession`, `SrcBuiltin`.
- REQ-V3R2-RT-002-002: Every `PermissionRule` SHALL carry a non-empty `Origin string` field identifying the file path that contributed the rule.
- REQ-V3R2-RT-002-003: The `PermissionMode` enum SHALL include `bubble` as a first-class value alongside `default`, `acceptEdits`, `bypassPermissions`, `plan`.
- REQ-V3R2-RT-002-004: `PermissionResolver.Resolve` SHALL walk the 8 tiers in priority order and return the first non-empty decision encountered.
- REQ-V3R2-RT-002-005: Resolution output SHALL include `PermissionDecision`, optional `UpdatedInput`, and a `ResolvedBy Source` field naming the tier that supplied the decision.
- REQ-V3R2-RT-002-006: The system SHALL ship a pre-allowlist covering `Bash(go test:*)`, `Bash(golangci-lint run:*)`, `Bash(ruff check:*)`, `Bash(npm test:*)`, `Bash(pytest:*)`, `Read(*)`, `Glob(*)`, `Grep(*)` at `SrcBuiltin` tier.
- REQ-V3R2-RT-002-007: `moai doctor permission` SHALL print the effective rule for a candidate `(tool, input)` pair along with the full chain of tiers inspected and the tier that supplied the answer.
- REQ-V3R2-RT-002-008: Agent frontmatter `permissionMode` SHALL be validated at CI time to be one of the 5 enum values; unknown values SHALL fail the build.

### 5.2 Event-Driven Requirements

- REQ-V3R2-RT-002-010: WHEN a tool invocation arrives and no rule in tiers 1-8 matches the pattern, the resolver SHALL return `PermissionDecision: "ask"` for tiers that default to ask-on-miss and `"deny"` for tiers that default to deny-on-miss (per `default_action` field in each tier's settings file).
- REQ-V3R2-RT-002-011: WHEN a PreToolUse hook returns `PermissionDecision` in its HookResponse, the resolver SHALL treat it as overriding any tier at or below `SrcSession` priority for that single tool call.
- REQ-V3R2-RT-002-012: WHEN an agent with `permissionMode: bubble` is a fork (parent session exists) AND the resolver would return `PermissionDecision: "ask"`, the system SHALL route the prompt to the parent session's AskUserQuestion channel rather than the fork's own channel.
- REQ-V3R2-RT-002-013: WHEN a hook returns `UpdatedInput`, the resolver SHALL re-run pattern matching against the mutated input before finalizing the decision.
- REQ-V3R2-RT-002-014: WHEN a rule at `SrcPolicy` tier denies a tool, no lower-priority tier SHALL be able to override it.
- REQ-V3R2-RT-002-015: WHEN the user requests `/moai doctor permission --trace`, the resolver SHALL emit a machine-readable JSON trace listing every rule consulted and why it matched or missed.

### 5.3 State-Driven Requirements

- REQ-V3R2-RT-002-020: WHILE the agent's `permissionMode` is `plan`, the resolver SHALL return `PermissionDecision: "deny"` for every tool invocation that writes files (Write, Edit, Bash with known-write patterns) regardless of allowlist tiers.
- REQ-V3R2-RT-002-021: WHILE the agent's `permissionMode` is `bypassPermissions` AND the session is not a fork, the resolver SHALL short-circuit the tier walk and return `PermissionDecision: "allow"` with `ResolvedBy: SrcBuiltin` and `Origin: "bypassPermissions mode"`.
- REQ-V3R2-RT-002-022: WHILE `.moai/config/sections/security.yaml` sets `permission.strict_mode: true`, the `bypassPermissions` mode SHALL be rejected at agent-spawn time with error `PermissionModeRejected`.
- REQ-V3R2-RT-002-023: WHILE the fork depth exceeds 3 levels (fork of fork of fork), the resolver SHALL treat all permission modes except `plan` as degraded to `bubble` and raise a SystemMessage warning.

### 5.4 Optional Features

- REQ-V3R2-RT-002-030: WHERE a project defines `.claude/settings.local.json` with `permissions.session_rules`, those rules SHALL populate `SrcSession` tier and apply only during the current session.
- REQ-V3R2-RT-002-031: WHERE a plugin declares `permissions.rules` in its plugin.json (v3.1+ feature slot), the rules SHALL populate `SrcPlugin` tier tagged with the plugin name in `Origin`.
- REQ-V3R2-RT-002-032: WHERE `moai doctor permission --dry-run <tool> <input>` is invoked, the system SHALL print the resolution without executing the tool.

### 5.5 Unwanted Behavior

- REQ-V3R2-RT-002-040: IF a rule at any tier uses `bypassPermissions` action (legacy v2 form), THEN the loader SHALL migrate it to `acceptEdits` mode scoped to the rule's pattern and emit a deprecation warning identifying the origin file.
- REQ-V3R2-RT-002-041: IF the resolver returns `PermissionDecision: "ask"` AND no AskUserQuestion channel is reachable (non-interactive mode), THEN the system SHALL fail closed with `PermissionDecision: "deny"` and log the unreachable-prompt event to `.moai/logs/permission.log`.
- REQ-V3R2-RT-002-042: IF two rules at the same tier match the input with conflicting actions, THEN the more specific pattern SHALL win; if specificity is equal, the rule whose `Origin` file path comes later in filesystem scan order SHALL win; the conflict SHALL be logged.
- REQ-V3R2-RT-002-043: IF a fork agent attempts to register its own session-scoped rule that would allow an action denied at the parent's `SrcPolicy` tier, THEN the registration SHALL be rejected and logged.

### 5.6 Complex Requirements

- REQ-V3R2-RT-002-050: WHILE an agent runs with `permissionMode: bubble` AND the parent session has been closed, WHEN a non-allowlisted tool call arrives, THEN the resolver SHALL halt the tool call with `PermissionDecision: "deny"` and `SystemMessage: "Bubble target parent unavailable — decision deferred"`.
- REQ-V3R2-RT-002-051: WHILE `SrcSkill` tier contains a rule AND the same pattern appears at `SrcProject` tier with conflicting action, THEN `SrcProject` wins per the 8-tier priority ordering, and the override is surfaced in `moai doctor permission`.

## 6. Acceptance Criteria (수용 기준)

- AC-V3R2-RT-002-01: Given a `Bash(rm -rf /)` invocation and a project-level `SrcProject` rule denying `Bash(rm*:*)`, When resolved, Then the result is `{PermissionDecision: "deny", ResolvedBy: SrcProject, Origin: ".claude/settings.json"}`. (maps REQ-V3R2-RT-002-004, -005)
- AC-V3R2-RT-002-02: Given a `Bash(go test ./...)` invocation with only the pre-allowlist active, When resolved, Then the result is `{PermissionDecision: "allow", ResolvedBy: SrcBuiltin, Origin: "pre-allowlist"}`. (maps REQ-V3R2-RT-002-006)
- AC-V3R2-RT-002-03: Given an agent spawned with `permissionMode: bubble` under a parent terminal session, When it attempts `Write(path: /tmp/test.txt)`, Then the AskUserQuestion prompt appears in the parent session's channel, not the fork's mailbox. (maps REQ-V3R2-RT-002-012)
- AC-V3R2-RT-002-04: Given a PreToolUse hook returns `PermissionDecision: "allow"` via HookResponse, When the resolver is invoked, Then the hook-contributed decision overrides any `SrcSession`, `SrcSkill`, `SrcBuiltin` rules for that single call. (maps REQ-V3R2-RT-002-011)
- AC-V3R2-RT-002-05: Given `moai doctor permission --trace Bash "go build"`, When invoked, Then stdout contains a JSON trace enumerating all 8 tiers inspected with `matched: true|false` per tier and the final `ResolvedBy` value. (maps REQ-V3R2-RT-002-007, -015)
- AC-V3R2-RT-002-06: Given an agent in `plan` mode, When it invokes `Write(path: /tmp/x)`, Then the resolver returns `deny` with `Origin: "plan mode denies writes"`. (maps REQ-V3R2-RT-002-020)
- AC-V3R2-RT-002-07: Given `security.yaml` sets `permission.strict_mode: true`, When an agent with `permissionMode: bypassPermissions` is spawned, Then spawn fails with error `PermissionModeRejected`. (maps REQ-V3R2-RT-002-022)
- AC-V3R2-RT-002-08: Given a fork agent with `permissionMode: bubble` and parent session already closed, When a non-allowlisted tool fires, Then the call is denied with systemMessage "Bubble target parent unavailable". (maps REQ-V3R2-RT-002-050)
- AC-V3R2-RT-002-09: Given an agent frontmatter contains `permissionMode: ultra-bypass`, When CI lint runs, Then the build fails with error naming the agent file and the unknown value. (maps REQ-V3R2-RT-002-008)
- AC-V3R2-RT-002-10: Given a PreToolUse hook mutates input via `UpdatedInput: {file_path: "/safe/path"}`, When resolver re-matches, Then allowlist patterns against `/safe/path` are consulted (not the original path). (maps REQ-V3R2-RT-002-013)
- AC-V3R2-RT-002-11: Given a legacy v2 rule with `action: bypassPermissions`, When the loader reads it, Then it is migrated to `acceptEdits` scoped to the pattern and a deprecation warning naming the origin file is logged. (maps REQ-V3R2-RT-002-040)
- AC-V3R2-RT-002-12: Given two `SrcLocal` rules both match `Bash(git push*)` with actions `allow` and `deny`, When resolved, Then the more-specific pattern (e.g., `Bash(git push origin main)`) wins over `Bash(git push*)`. (maps REQ-V3R2-RT-002-042)
- AC-V3R2-RT-002-13: Given `SrcPolicy` contains `deny Bash(curl:*)` and `SrcProject` contains `allow Bash(curl:*)`, When resolved, Then policy tier wins and the tool is denied. (maps REQ-V3R2-RT-002-014)
- AC-V3R2-RT-002-14: Given a fork at depth 4, When any non-`plan` mode agent fires, Then the effective mode is treated as `bubble` and a warning SystemMessage is emitted. (maps REQ-V3R2-RT-002-023)
- AC-V3R2-RT-002-15: Given non-interactive mode (`CI=1`), When resolver would return `ask`, Then the result is instead `deny` with log entry in `.moai/logs/permission.log`. (maps REQ-V3R2-RT-002-041)

## 7. Constraints (제약)

- Technical: Go 1.22+; no new external dependencies (validator/v10 from SPEC-V3R2-SCH-001 covers struct tags).
- Backward compat: v2 `permissions.allow` flat lists MUST continue to parse for the v3.0 → v3.x window per master §8 BC-V3R2-015 (AUTO migration via reader layer).
- Platform: macOS / Linux / Windows. Bubble dispatch uses the same AskUserQuestion channel Claude Code 2.1.111+ provides.
- Performance: Tier walk MUST complete in under 500 µs p99 for rule sets up to 1000 patterns total across all tiers.
- Memory: Merged rule set in memory MUST NOT exceed 256 KiB for typical project (~100 rules per tier).
- Bubble UX: REQ-V3R2-RT-002-012 routes prompts to parent AskUserQuestion; the prompt itself uses the orchestrator-authored question text (not the fork's body) to avoid leaking teammate context.

## 8. Risks & Mitigations (리스크 및 완화)

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Bubble prompts fatigue users with frequent dev-op dialogs | H | M | Pre-allowlist covers ~80% common dev ops (REQ-V3R2-RT-002-006); telemetry post-beta.1 tunes the list (master §10 R7). |
| Provenance metadata bloats merged config in memory | L | L | Provenance is stored per-rule as a single string; worst case is O(N_rules) strings capped by 256 KiB constraint. |
| Forks at depth >3 confuse users on why permissions tightened | M | L | REQ-V3R2-RT-002-023 emits explicit SystemMessage naming depth limit; documentation publishes the design. |
| Priority order inverts user expectation when `SrcLocal` was assumed higher than `SrcUser` | M | M | `moai doctor permission` publishes the full chain; migration guide clarifies CC-parity ordering. |
| Plugin tier slot without actual plugin contributors yields dead code | L | L | `SrcPlugin` tier is wired but empty in v3.0; SPEC-V3R2-EXT-003 documents the plugin slot for v4. |
| Existing agents declaring `bypassPermissions` break under strict_mode | M | M | REQ-V3R2-RT-002-022 is opt-in via security.yaml flag; default remains compatible. |

## 9. Dependencies (의존성)

### 9.1 Blocked by

- SPEC-V3R2-RT-001 (provides `PermissionDecision` field and hook integration point for REQ-V3R2-RT-002-011).
- SPEC-V3R2-RT-005 (provides Source provenance reader for settings file tiers).
- SPEC-V3R2-CON-001 (FROZEN-zone codification declares the 8-source ordering as a constitutional clause).

### 9.2 Blocks

- SPEC-V3R2-RT-003 (sandbox launcher consults `PermissionMode` to decide whether to wrap the process in bwrap/seatbelt).
- SPEC-V3R2-ORC-001 (agent roster reduction consumes the validated `permissionMode` enum from REQ-V3R2-RT-002-008).
- SPEC-V3R2-ORC-004 (worktree MUST for implementers pairs with `permissionMode: acceptEdits` at project tier).

### 9.3 Related

- SPEC-V3R2-MIG-001 (v2→v3 migrator rewrites flat `permissions.allow` to tier-annotated form).
- SPEC-V3R2-SPC-004 (ACI `moai_lsp_*` commands must coexist with sandbox/permission layers — this SPEC gates their exposure via allowlist).
- SPEC-V3R2-CON-003 (constitution consolidation moves permission rule text into `.claude/rules/moai/core/settings-management.md`).

## 10. Traceability (추적성)

- Theme: master §4.3 Layer 3 Runtime; §5.2 Multi-Source Permission Resolution; §8 BC-V3R2-015.
- Principle: P6 (Permission Bubble Over Bypass); secondary P7 (Sandbox Default), P8 (Hook JSON).
- Pattern: S-1 (Multi-Source Permission Resolution priority 3), S-2 (Permission Bubble/Escalation).
- Problem: P-C01 (no permission bubble, CRITICAL), P-C04 (no config provenance, HIGH, prerequisite).
- Master Appendix A: Principle P6 → primary SPEC-V3R2-RT-002.
- Master Appendix C: Pattern S-1 → primary SPEC-V3R2-RT-002 (priority 3); S-2 → same SPEC.
- Wave 1 sources: r3-cc-architecture-reread.md §1.3 (8-source precedence), §2 Decision 15 (bubble as first-class mode), §4 Adopt 2 (permission envelope recommendation).
- Wave 2 sources: design-principles.md P6 (Permission Bubble); pattern-library.md S-1 (priority 3), S-2; problem-catalog.md Cluster 5 P-C01/P-C04.
- BC-ID: none (RT-002 is additive on top of BC-V3R2-015's multi-layer settings reader; bubble mode + tier-aware resolution introduce no breaking change on existing flat permission configs).
- Priority: P0 Critical — blocks sandbox (RT-003), settings provenance (RT-005), agent cleanup (ORC-001/004).
- **File:line anchors** (per D5 traceability requirement):
  - `docs/design/major-v3-master.md:L1043` (§11.3 RT-002 definition)
  - `docs/design/major-v3-master.md:L974` (§8 BC-V3R2-015 — multi-layer settings, on which RT-002 layers bubble mode)
  - `docs/design/major-v3-master.md:L989` (§9 Phase 2 Runtime Hardening)
  - `.moai/design/v3-redesign/synthesis/problem-catalog.md` Cluster 5 (P-C01, P-C04)
