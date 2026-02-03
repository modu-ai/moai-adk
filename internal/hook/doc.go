// Package hook implements the Compiled Hook System for MoAI-ADK Go Edition.
//
// It replaces the Python-based hook scripts (~21,535 LOC across 46 files) with
// a single compiled binary subcommand (~1,500 LOC). Hooks are invoked via
// "moai hook <event>" and communicate with Claude Code through a JSON
// stdin/stdout protocol.
//
// Architecture:
//
//   - Protocol: Reads JSON from stdin, writes JSON to stdout (REQ-HOOK-010~013)
//   - Registry: Manages handler registration and event dispatch (REQ-HOOK-001~004)
//   - Contract: Validates the hook execution environment per ADR-012 (REQ-HOOK-020~022)
//   - Handlers: Six event handlers for SessionStart, PreToolUse, PostToolUse,
//     SessionEnd, Stop, and PreCompact events (REQ-HOOK-030~036)
//
// Exit Code Semantics:
//
//   - 0: Allow / Success - tool execution permitted
//   - 2: Block - tool execution denied, reason provided
//   - Other: Non-blocking error - logged and execution continues
//
// All diagnostic output is written to stderr via log/slog. stdout is reserved
// exclusively for JSON responses (REQ-HOOK-013).
//
// Reference: SPEC-HOOK-001, ADR-006 (Hooks as Binary Subcommands),
// ADR-012 (Hook Execution Contract).
package hook
