---
id: SPEC-V3-CMDS-003
title: "Command Inline Bash Execution — !`cmd` substitution"
version: "0.1.0"
status: draft
created: 2026-04-22
updated: 2026-04-22
author: GOOS
priority: P2 Medium
phase: "v3.0.0 — Phase 3 Agent Runtime v2"
module: "internal/command/shell_exec.go"
dependencies:
  - SPEC-V3-CMDS-001
  - SPEC-V3-CMDS-002
related_gap:
  - gm#103
related_theme: "Command inline shell execution parity"
breaking: false
bc_id: null
lifecycle: spec-anchored
tags: "command, shell, bash, substitution, v3"
---

# SPEC-V3-CMDS-003: Command Inline Bash Execution

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-04-22 | GOOS | Initial v3 draft from Wave 4 bundle (Hooks/Commands) |

---

## 1. Goal (목적)

Claude Code supports inline shell execution inside command/skill bodies via two syntaxes:
- Block: ` ```! command ``` ` — fenced block whose body is executed and output substituted in place.
- Inline: `` !`command` `` — single-line execution with output substituted where it appears (requires leading whitespace or line-start before `!`).

Execution goes through the standard BashTool permission flow (pre-check), honoring user allow/deny rules. PowerShell variant available via `shell: powershell` frontmatter. MCP-sourced skills are explicitly blocked from this mechanism (untrusted remote content). Built-in replacements `${CLAUDE_SKILL_DIR}` and `${CLAUDE_SESSION_ID}` apply before shell exec.

This SPEC adds the inline execution primitive to moai-adk so commands can, e.g., pull `git rev-parse HEAD` into their body dynamically.

## 2. Scope (범위)

In-scope:
- Block syntax ` ```! ... ``` ` scanner: recognize fenced block with `!` language specifier; extract body; execute; replace block with output.
- Inline syntax `` !`...` ``: recognize after whitespace/line-start; execute; replace with output.
- Execution via the user's active BashTool permission layer (pre-approved commands run; disallowed commands trigger permission prompt or hook denial).
- `shell: powershell` frontmatter (SPEC-V3-CMDS-001 adds the discriminator; this SPEC adds the PowerShell path via `pwsh -NoProfile -NonInteractive -Command`).
- `${CLAUDE_SKILL_DIR}` body substitution (skill/command's own directory, forward-slash on Windows).
- `${CLAUDE_SESSION_ID}` body substitution.
- Hard block: commands/skills loaded from MCP (`loadedFrom: 'mcp'`) MUST NOT allow inline shell execution.

Out-of-scope:
- New hook events for inline exec (each exec already fires PreToolUse for the underlying Bash call).
- $ARGUMENTS substitution inside the shell block (covered by SPEC-V3-CMDS-002 — substitution happens BEFORE exec).
- Full fork-context isolation of inline exec (the exec runs in the session's current permission context by design).
- PowerShell on non-Windows platforms (if `pwsh` binary unavailable, emit error).

## 3. Environment (환경)

Current moai-adk state:
- Command/skill body is rendered verbatim (after `$ARGUMENTS` substitution per current moai support) — no shell execution (findings-wave1-moai-current.md §8).
- No `${CLAUDE_SKILL_DIR}` or `${CLAUDE_SESSION_ID}` substitution in bodies.
- PowerShell shell not supported in command bodies.

Claude Code reference:
- `utils/promptShellExecution.ts:48-143` — block and inline executor (findings-wave1-hooks-commands.md §9).
- `skills/loadSkillsDir.ts:373-396` — MCP skills blocked from shell exec (findings-wave1-hooks-commands.md §9).
- `skills/loadSkillsDir.ts:358-369` — `${CLAUDE_SKILL_DIR}` / `${CLAUDE_SESSION_ID}` substitution (findings-wave1-hooks-commands.md §9).
- `utils/hooks.ts:957-984` — PowerShell invocation pattern (`pwsh -NoProfile -NonInteractive -Command`) (findings-wave1-hooks-commands.md §4.5).

Affected modules:
- `internal/command/shell_exec.go` — new file: block and inline scanner + executor.
- `internal/command/loader.go` — integrate shell_exec after argument substitution.
- `internal/command/substitutions.go` — new file: `${CLAUDE_SKILL_DIR}` / `${CLAUDE_SESSION_ID}` resolver.

## 4. Assumptions (가정)

- BashTool permission layer is reachable from command rendering path; commands rendered at user-initiated slash invocation time (not at load time).
- Shell exec output is captured verbatim (stdout + stderr combined, or stdout-only per `2>&1` convention — implementation chooses stdout-only with a trailing note if stderr is non-empty).
- `${CLAUDE_SESSION_ID}` is available from the session context in the rendering path.
- Users understand inline exec happens in the session's current directory (not in the command's own dir) unless they explicitly `cd "${CLAUDE_SKILL_DIR}"` at the start of their block.

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

- REQ-CMDS-003-001: The command rendering pipeline SHALL support two shell-exec syntaxes: block ` ```! ... ``` ` and inline `` !`...` `` (with leading whitespace or line-start).
- REQ-CMDS-003-002: Every shell-exec invocation SHALL route through the session's existing BashTool permission flow; denied commands SHALL NOT execute silently.
- REQ-CMDS-003-003: Shell-exec output SHALL replace the matched syntax in the rendered body in-place.
- REQ-CMDS-003-004: The `${CLAUDE_SKILL_DIR}` placeholder SHALL substitute the absolute path to the command/skill's own directory before shell exec, normalized to forward slashes on Windows.
- REQ-CMDS-003-005: The `${CLAUDE_SESSION_ID}` placeholder SHALL substitute the current session ID before shell exec.
- REQ-CMDS-003-006: When frontmatter sets `shell: powershell`, the shell exec SHALL use `pwsh -NoProfile -NonInteractive -Command` instead of bash.
- REQ-CMDS-003-007: Commands/skills loaded from MCP sources (`loadedFrom: 'mcp'`) SHALL NOT be permitted to invoke shell-exec; the rendering pipeline SHALL strip exec syntax and emit a warn log.
- REQ-CMDS-003-008: Shell-exec SHALL honor a per-invocation timeout defaulting to 30 seconds; overridable via `shell_exec.timeout` frontmatter (out-of-scope field, no validation beyond positive integer).

### 5.2 Event-driven Requirements

- REQ-CMDS-003-010: WHEN a block or inline exec completes successfully, the output SHALL replace the syntax in the rendered body; stdout is preferred; if stdout is empty and stderr is non-empty, stderr SHALL be substituted with a trailing note `(from stderr)`.
- REQ-CMDS-003-011: WHEN a block or inline exec exits with non-zero code, the rendering pipeline SHALL substitute the combined output followed by `[exit N]` marker where N is the exit code.
- REQ-CMDS-003-012: WHEN the shell-exec times out (exceeds the configured timeout), the process SHALL be SIGKILLed and the substitution SHALL be `[timeout after Xs]` where X is the elapsed seconds.
- REQ-CMDS-003-013: WHEN the `pwsh` binary is not available at exec time on any platform, the rendering pipeline SHALL abort with `CommandPowerShellUnavailable`.

### 5.3 State-driven Requirements

- REQ-CMDS-003-020: WHILE the command/skill has `loadedFrom: 'mcp'`, all inline and block exec syntax SHALL be stripped from the rendered body and a warn log SHALL name the skill.
- REQ-CMDS-003-021: WHILE the session is in plan-mode (read-only), shell-exec SHALL be refused with `CommandShellExecInPlanMode` and the block SHALL be replaced with `[shell-exec skipped: plan mode]`.

### 5.4 Optional Features

- REQ-CMDS-003-030: WHERE a block specifies a language other than `!` (e.g., ` ```!python `), the executor SHALL treat the language as the interpreter (`python`, `node`, etc.) subject to BashTool permission for that interpreter.
- REQ-CMDS-003-031: WHERE the configuration key `command.shell_exec.enabled: false` is set, all shell-exec syntax SHALL be ignored (passed through as literal text).

### 5.5 Complex Requirements

- REQ-CMDS-003-040: IF a block exec body contains `$ARGUMENTS` or `$ARGUMENTS[N]` placeholders, THEN those SHALL be substituted BEFORE the shell exec (per SPEC-V3-CMDS-002 ordering); ELSE the exec runs with the unmodified body.
- REQ-CMDS-003-041: IF inline exec syntax `` !`cmd` `` appears at the very start of a line (no preceding whitespace), THEN it is still recognized; ELSE if `!` is mid-word (e.g., `foo!bar`), the syntax is NOT recognized and text passes through verbatim.

## 6. Acceptance Criteria (수용 기준 요약)

- AC-CMDS-003-01: Given a command body containing ` ```!\ngit rev-parse HEAD\n``` `, When rendered, Then the block is replaced with the stdout of `git rev-parse HEAD`. (maps REQ-CMDS-003-001, REQ-CMDS-003-003)
- AC-CMDS-003-02: Given a command body containing `Branch: !`git branch --show-current``, When rendered, Then the inline syntax is replaced with the output of `git branch --show-current`. (maps REQ-CMDS-003-001)
- AC-CMDS-003-03: Given a command body with `!`rm -rf /`` and BashTool permission denies the command, When rendered, Then no exec occurs and a permission error is surfaced. (maps REQ-CMDS-003-002)
- AC-CMDS-003-04: Given a command body with `cd "${CLAUDE_SKILL_DIR}" && ls`, When rendered, Then `${CLAUDE_SKILL_DIR}` substitutes to the command's absolute dir before exec. (maps REQ-CMDS-003-004)
- AC-CMDS-003-05: Given a command body with `echo "${CLAUDE_SESSION_ID}"`, When rendered in session `abc123`, Then the output is `abc123`. (maps REQ-CMDS-003-005)
- AC-CMDS-003-06: Given a command with frontmatter `shell: powershell` and body ` ```!\nGet-Location\n``` `, When rendered on a machine with `pwsh` available, Then the block is executed via `pwsh -NoProfile -NonInteractive -Command`. (maps REQ-CMDS-003-006)
- AC-CMDS-003-07: Given a skill with `loadedFrom: 'mcp'` and body containing `!`whoami``, When rendered, Then the syntax is stripped and a warn log names the skill. (maps REQ-CMDS-003-007, REQ-CMDS-003-020)
- AC-CMDS-003-08: Given a block exec that sleeps for 60 seconds and `shell_exec.timeout: 5`, When rendered, Then the process is killed after 5s and the body contains `[timeout after 5s]`. (maps REQ-CMDS-003-012)
- AC-CMDS-003-09: Given a block with exec `git status` that exits 0 with empty stdout, stderr `nothing to commit`, When rendered, Then the body contains `nothing to commit (from stderr)`. (maps REQ-CMDS-003-010)
- AC-CMDS-003-10: Given `command.shell_exec.enabled: false`, When a body with `!`date`` is rendered, Then the syntax passes through as literal text `!`date``. (maps REQ-CMDS-003-031)
- AC-CMDS-003-11: Given a session in plan-mode and a body with `!`rm foo``, When rendered, Then the syntax is replaced with `[shell-exec skipped: plan mode]` and no exec occurs. (maps REQ-CMDS-003-021)

## 7. Constraints (제약)

- Technical: Go 1.22+. `os/exec.CommandContext` with context-based timeout.
- Backward compat: Existing commands without exec syntax render unchanged.
- Platform: Bash via `/bin/bash -c` on POSIX; PowerShell via `pwsh -NoProfile -NonInteractive -Command` on any platform where available.
- Security: Exec always routes through BashTool permission layer; MCP-sourced commands strictly blocked.
- Performance: Exec runs synchronously; recommend async hooks (SPEC-V3-HOOKS-003) for long-running commands.

## 8. Risks & Mitigations (리스크 및 완화)

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| User commands inject shell metacharacters via `$ARGUMENTS[N]` substitution | M | H | Substitution happens at text level; user is responsible for quoting. Docs-site prominently documents this. `moai doctor command --validate` flags suspicious patterns like `rm $ARGUMENTS` unquoted. |
| MCP block bypass via clever YAML | L | H | `loadedFrom` is source-tracked in loader; syntax stripping is unconditional when `loadedFrom == 'mcp'`. Golden tests cover bypass attempts. |
| Timeout SIGKILL leaves zombie subprocesses | L | M | `CommandContext` + `exec.Cmd.Process.Kill()` on deadline; wait goroutine harvests zombie; integration tests cover. |
| PowerShell auto-detection assumes `pwsh` in PATH | L | M | Explicit error `CommandPowerShellUnavailable` at exec time (REQ-CMDS-003-013) rather than silent fallback; `moai doctor` surfaces pwsh availability. |
| Inline `!` clashes with user's literal `!` text | L | L | Strict grammar: `!` requires whitespace/line-start + backtick; mid-word `!` passes through verbatim (REQ-CMDS-003-041). |

## 9. Dependencies (의존성)

### 9.1 Blocked by

- SPEC-V3-CMDS-001 (command frontmatter includes `shell:` field).
- SPEC-V3-CMDS-002 (`$ARGUMENTS` substitution happens before shell exec).

### 9.2 Blocks

- None.

### 9.3 Related

- SPEC-V3-SKL-001 (shared substitution engine with skill bodies).
- SPEC-V3-HOOKS-002 (`type: command` hook executor reuses BashTool permission pattern).
- SPEC-V3-HOOKS-003 (async hook pattern recommended for long-running shell exec).

## 10. Traceability (추적성)

- Theme: master-v3 §8.4 (Skill SPECs — shared command+skill shell exec).
- Gap rows: gm#103 (Medium — `!`cmd` and `` ```! `` block execution inside skill body).
- BC-ID: None (purely additive).
- Wave 1 sources: findings-wave1-hooks-commands.md §9 (in-skill shell execution), §14.5 (command feature gap table), §12 (source references `utils/promptShellExecution.ts:48-143`, `skills/loadSkillsDir.ts:358-369, 373-396`).
- Priority: P2 Medium (ergonomic improvement for commands needing dynamic context; ships Phase 3).
