//! Hook engine (rc.2 #2391): one policy core for the four MoAI hook duties (session-check,
//! tool-preflight, tool-observe, turn-validation) with a native adapter per host. Stage 1:
//! `session-check` and `tool-preflight`; stage 2: `tool-observe` (observation record only)
//! and `turn-validation` (Stop machinery with an empty rule set); stage 3: the Codex adapter
//! for all four (`render_codex`, from the Codex hooks page) and [`capabilities`], the per-host
//! coverage `hooks-plan` reports.
//!
//! Sources (read 2026-11-10, https://code.claude.com/docs/en/hooks):
//! - PreToolUse decision control: exit 0 and print
//!   `{"hookSpecificOutput": {"hookEventName": "PreToolUse", "permissionDecision": "deny",
//!   "permissionDecisionReason": "..."}}`; `deny` prevents the tool call. The top-level
//!   `decision`/`reason` fields are deprecated for this event.
//! - SessionStart: `hookSpecificOutput.additionalContext` is "added to Claude's context at the
//!   start of the conversation"; SessionStart cannot block (exit 2 only shows stderr).
//! - Exit codes: 0 success (JSON fields apply); 2 blocking error; any other code, a failure to
//!   start or invalid JSON is a non-blocking error and the action goes ahead.
//!
//! Output contract (one deny channel, no auto-allow):
//! - allow: exit 0, empty stdout. The engine never prints `permissionDecision: "allow"`, so the
//!   host's own permission prompt still applies.
//! - policy deny: exit 0 and the documented PreToolUse deny object on stdout.
//! - session context: exit 0 and the documented SessionStart object on stdout.
//! - Stop block (turn-validation): exit 0 and the documented top-level
//!   `{"decision": "block", "reason": "..."}`; Claude Code caps stop-hook continuations at 8 in
//!   a row and sets `stop_hook_active` while continuing. PostToolUse (tool-observe) never
//!   blocks: the tool already ran.
//! - tool-observe input that does not parse (a tool result past the stdin bound is cut off) is
//!   a quiet skip: exit 0, nothing printed, nothing recorded.
//! - engine error (unsupported host or event, malformed input, wrong event name): exit 1 with
//!   the reason on stderr and nothing on stdout — a non-blocking error by the host's rules.
//!   Whether an engine error should block (fail-closed, the host's `onFailure: "block"`) is a
//!   user decision and is not taken here.
//!
//! Cost: no network on any path. The reads are the hook stdin, `git symbolic-ref` /
//! `git rev-parse HEAD` in the session's working directory, this clone's local claim journal
//! and the bounded engine log. Writes: tool-observe and a turn-validation block append one
//! line to `.moai-cli/receipts/hook-engine.jsonl`, only when `.moai-cli/` already exists
//! (the shared `hooks::append_receipt_line` guards). Text that reaches the
//! model (deny reasons, session context) is built from fixed strings, enumerated states and
//! Issue numbers only; tool input, branch names and file content are never echoed.

use serde_json::{json, Value};
use std::path::Path;

/// Bound on the hook stdin read by the CLI (tool input can be large).
pub const STDIN_LIMIT: u64 = 4 * 1024 * 1024;

/// The hook duties. Stage 1 implements the first two.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Event {
    SessionCheck,
    ToolPreflight,
    ToolObserve,
    TurnValidation,
    /// SPEC-20261117-edit-diagnostics S1: the user's checker after a file edit (Claude Code,
    /// opt-in `[hooks.edit_diagnostics]`).
    EditDiagnostics,
}

impl Event {
    pub fn parse(word: &str) -> Option<Event> {
        match word {
            "session-check" => Some(Event::SessionCheck),
            "tool-preflight" => Some(Event::ToolPreflight),
            "tool-observe" => Some(Event::ToolObserve),
            "turn-validation" => Some(Event::TurnValidation),
            "edit-diagnostics" => Some(Event::EditDiagnostics),
            _ => None,
        }
    }

    /// The host event name each duty answers (identical in Claude Code and Codex).
    fn host_event(self) -> &'static str {
        match self {
            Event::SessionCheck => "SessionStart",
            Event::ToolPreflight => "PreToolUse",
            Event::ToolObserve | Event::EditDiagnostics => "PostToolUse",
            Event::TurnValidation => "Stop",
        }
    }
}

/// The opt-in deny rules of stage 7 (B3, B9; decision 6779829826), off unless
/// `~/.moai-cli/hook-config.toml` turns them on (see `hook_config`).
pub const OPT_IN_RULES: &str = "[hooks.outside_project_writes] enabled + allow: file writes outside the project (after symlinks and ..) are refused unless under allow; [hooks.primary_checkout] enabled + repos + allow_commands: switch, checkout <branch>, reset, stash (not list/show) and rebase are refused in a listed primary checkout that has linked worktrees. A missing or broken config turns both off";

/// How edit-diagnostics is turned on (SPEC-20261117-edit-diagnostics S1).
pub const EDIT_OPT_IN: &str = "[hooks.edit_diagnostics] enabled + checkers (extension -> argv, the edited file appended, no shell) in the HOME hook-config.toml; timeout_secs 1-30 (10), max_lines 1-200 (20), max_chars 200-9000 (4000). Off when absent; a malformed table is off and logged as config_error";

/// The confirmation class of tool-preflight (B2, Issue #2391 classification row B2).
pub const CONFIRM_COMMANDS: &str =
    "git reset --hard, a forced git clean (not a dry run), a force push to a branch other than main or master, and in a MoAI project a direct pull request merge (gh pr merge, gh api PUT .../merge, the GraphQL merge mutations) outside `moai gate-merge`";

/// How `[hooks.modes]` changes the direct-merge confirmation (card C-86, decision note 2346).
pub const CONFIRM_MODES: &str = "[hooks.modes] merge_confirm = off | observe | enforce in the HOME hook-config.toml; absent or any other word is observe: the call goes through and one would_block line says it would have been confirmed; off does nothing; enforce confirms (Claude asks, Codex and OMP refuse). Where enforce refuses, the third refusal of the rule in one session goes through and one rule_release line records why, except that a call naming no session (OMP sends none, Codex may send an empty one) is refused every time; a refusal that cannot be written to the engine log goes through (hook cwd in a subfolder, the log at its size limit, a project without a .moai-cli folder)";

/// Appended to a confirmation on Codex, which cannot ask.
pub const CODEX_ASK_SUFFIX: &str =
    "Codex cannot ask for confirmation in a hook, so it is refused here; run it yourself outside the agent if you mean it.";

/// What the policy core decided, before any host rendering.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum Verdict {
    Allow,
    Deny(String),
    /// Confirmation needed (decision B2 = A, Issue #2389 c6774899518 row 25): Claude Code asks
    /// the user; Codex has no ask, so it is rendered as a deny there. Never written as a deny
    /// line: the user's answer is unknown to the engine.
    Ask(String),
    Context(String),
}

/// A rendered host answer: what the CLI prints and how it exits.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Answer {
    pub stdout: String,
    pub stderr: String,
    pub exit: u8,
}

/// Exit code of an engine error: non-blocking in both hosts.
pub const ENGINE_ERROR_EXIT: u8 = 1;

fn engine_error(reason: &str) -> Answer {
    Answer {
        stdout: String::new(),
        stderr: format!("moai hook-dispatch: engine error: {reason}"),
        exit: ENGINE_ERROR_EXIT,
    }
}

/// Runs one hook: parses the host and duty, decides, renders the host's native answer.
/// `cwd` is the hook process's working directory (the host runs hooks in the session's
/// directory); it is also the project whose local claim journal is read and, for
/// tool-observe, whose `.moai-cli/receipts/` receives the observation line. `now_unix_ms` is
/// injected so no decision reads the wall clock.
pub fn dispatch(host: &str, event: &str, stdin: &[u8], cwd: &Path, now_unix_ms: u128) -> Answer {
    let cwd = std::fs::canonicalize(cwd).unwrap_or_else(|_| cwd.to_path_buf());
    let cwd = cwd.as_path();
    let Some(host) = Host::parse(host) else {
        return engine_error("unknown hook host; must be claude, codex or omp");
    };
    let Some(event) = Event::parse(event) else {
        return engine_error(
            "unknown hook event; must be session-check, tool-preflight, tool-observe, turn-validation or edit-diagnostics",
        );
    };
    // OMP's turn-end event is `session_stop` (OMP 18.18.7, extensions.md), but the generated
    // extension module does not subscribe to it, so no Stop rule of `STOP_RULES` can run for
    // OMP (the rules run for Claude and Codex only). A call is a configuration error, never a
    // guess.
    if host == Host::Omp && event == Event::TurnValidation {
        return engine_error(
            "turn-validation is not wired for OMP (session_stop exists; no rule uses it yet)",
        );
    }
    // Slice S1 of SPEC-20261117-edit-diagnostics is Claude Code only (leader decision Q6).
    if event == Event::EditDiagnostics && host != Host::Claude {
        return engine_error("edit-diagnostics is Claude Code only in this release");
    }
    let Ok(Value::Object(input)) = serde_json::from_slice::<Value>(stdin) else {
        // A PostToolUse input carries the whole tool result; one past the stdin bound is cut
        // off and no longer parses. Observation is best-effort and the tool already ran, so
        // that is a quiet skip, not an error the host would surface on every large result.
        if matches!(event, Event::ToolObserve | Event::EditDiagnostics) {
            return allow_quietly();
        }
        return engine_error("hook input is not a JSON object");
    };
    let input = Value::Object(input);
    // A present event name must match the duty; unknown extra fields are ignored.
    match input.get("hook_event_name") {
        None => {}
        Some(Value::String(name)) if name == event.host_event() => {}
        // tool-observe also records failed tool calls on Claude Code (`PostToolUseFailure`);
        // Codex has no such event, so there the name is a configuration error.
        Some(Value::String(name))
            if host == Host::Claude
                && event == Event::ToolObserve
                && name == "PostToolUseFailure" => {}
        Some(_) => return engine_error("hook_event_name does not match the requested duty"),
    }
    if event == Event::EditDiagnostics {
        return edit_diagnostics(&input, cwd, now_unix_ms, host);
    }
    // The sentence of a guard whose text is fixed (set by tool-preflight); see `record_deny`.
    let mut fixed_reason: Option<&'static str> = None;
    let verdict = match event {
        Event::SessionCheck => session_check(host, &input, cwd),
        Event::ToolPreflight => {
            // Stage 7 opt-in rules (B3, B9; decision 6779829826) come from ~/.moai-cli/hook-config.toml,
            // read once per call; an unusable file leaves them off and one line says why (N1).
            let loaded = crate::hook_config::load();
            if let Some(problem) = loaded.problem {
                record_config_error(&input, cwd, now_unix_ms, host, problem);
            }
            tool_preflight(
                &input,
                cwd,
                &loaded.config,
                host,
                now_unix_ms,
                &mut fixed_reason,
            )
        }
        Event::ToolObserve => tool_observe(&input, cwd, now_unix_ms, host),
        Event::TurnValidation => turn_validation_with(STOP_RULES, &input, cwd, now_unix_ms),
        // Answered above, before any verdict.
        Event::EditDiagnostics => Verdict::Allow,
    };
    if let (Event::ToolPreflight, Verdict::Deny(_)) = (event, &verdict) {
        record_deny(&input, cwd, now_unix_ms, host, fixed_reason);
    }
    match host {
        Host::Claude => render_claude(event, verdict),
        Host::Codex => render_codex(event, verdict),
        Host::Omp => render_omp(event, verdict),
    }
}

/// OMP rendering (#2633). OMP runs no command hooks; the generated extension module calls
/// `hook-dispatch --host omp` and turns this answer into the extension's own return value
/// (github.com/can1357/oh-my-pi `docs/extensions.md`, `extensibility/extensions/types.ts`,
/// read 2026-11-13):
/// - tool-preflight deny: `{"block": true, "reason": …}` — the `tool_call` result shape; OMP
///   hands `reason` to the model as the tool's error. OMP has no ask, so a confirmation is a
///   deny with the same suffix as Codex.
/// - session-check context: `{"context": …}` — the module adds it to `before_agent_start`'s
///   `systemPrompt`, because `session_start` returns nothing.
/// - tool-observe: nothing printed; the observation line is the effect.
fn render_omp(event: Event, verdict: Verdict) -> Answer {
    let stdout = match (event, verdict) {
        (_, Verdict::Allow) => String::new(),
        (Event::ToolPreflight, Verdict::Ask(reason)) => {
            json!({"block": true, "reason": format!("{reason} {CODEX_ASK_SUFFIX}")}).to_string()
        }
        (Event::ToolPreflight, Verdict::Deny(reason)) => {
            json!({"block": true, "reason": reason}).to_string()
        }
        (Event::SessionCheck, Verdict::Context(text)) => json!({"context": text}).to_string(),
        _ => return engine_error("policy returned a verdict this duty cannot carry"),
    };
    Answer {
        stdout,
        stderr: String::new(),
        exit: 0,
    }
}

/// The hosts with a native adapter. OMP (#2633) has no command hooks: its extension module
/// calls `hook-dispatch --host omp` and reads the small answer [`render_omp`] prints.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Host {
    Claude,
    Codex,
    Omp,
}

impl Host {
    pub fn parse(word: &str) -> Option<Host> {
        match word {
            "claude" => Some(Host::Claude),
            "codex" => Some(Host::Codex),
            "omp" => Some(Host::Omp),
            _ => None,
        }
    }

    pub fn name(self) -> &'static str {
        match self {
            Host::Claude => "claude",
            Host::Codex => "codex",
            Host::Omp => "omp",
        }
    }
}

/// A short key for the host's `tool_use_id` (sha256, first 16 hex digits), so a deny and the
/// observation of the same call can be matched without logging the raw id.
fn call_key(input: &Value) -> Value {
    match input["tool_use_id"].as_str() {
        Some(id) if !id.is_empty() => {
            let d = crate::digest(id.as_bytes());
            json!(d.trim_start_matches("sha256:")[..16].to_string())
        }
        _ => Value::Null,
    }
}

/// A tool-preflight deny leaves one `deny` line (session, call key, host, time) in the engine
/// log when `.moai-cli/` exists — the evidence `hooks::status` reads as `action_blocked` when
/// no observation of the same call follows. Best effort, never changes the answer.
fn record_deny(
    input: &Value,
    cwd: &Path,
    now_unix_ms: u128,
    host: Host,
    reason: Option<&'static str>,
) {
    // `reason` is a guard's fixed sentence, or null when the sentence is built at run time
    // (the commit-identity rule names an email, the claim rule an Issue number, a line that
    // commits in more than one folder names the verbs): the line never carries the command,
    // a path or file content.
    let line = json!({
        "schema_version": 1,
        "kind": "deny",
        "host": host.name(),
        "session": session_key(input),
        "call": call_key(input),
        "unix_ms": now_unix_ms as u64,
        "reason": reason,
    });
    let _ = crate::hooks::append_receipt_line(cwd, ENGINE_LOG, &line);
}

/// edit-diagnostics (SPEC-20261117-edit-diagnostics S1, Claude Code): off unless
/// `[hooks.edit_diagnostics]` is enabled and well formed; an unusable file or a malformed table
/// leaves one `config_error` line and nothing runs. The answer is the documented PostToolUse
/// `hookSpecificOutput.additionalContext` (code.claude.com/docs/en/hooks, read 2026-11-17),
/// always exit 0: this slice never blocks.
fn edit_diagnostics(input: &Value, cwd: &Path, now_unix_ms: u128, host: Host) -> Answer {
    let loaded = crate::hook_config::load();
    if let Some(problem) = loaded.problem {
        record_config_error(input, cwd, now_unix_ms, host, problem);
        return allow_quietly();
    }
    if loaded.config.edit_invalid {
        record_config_error(
            input,
            cwd,
            now_unix_ms,
            host,
            crate::hook_config::ConfigProblem::InvalidEditDiagnostics,
        );
        return allow_quietly();
    }
    let Some(rule) = loaded.config.edit else {
        return allow_quietly();
    };
    match crate::edit_diagnostics::run(input, cwd, &rule) {
        Some(text) => Answer {
            stdout: json!({
                "hookSpecificOutput": {"hookEventName": "PostToolUse", "additionalContext": text}
            })
            .to_string(),
            stderr: String::new(),
            exit: 0,
        },
        None => allow_quietly(),
    }
}

/// An unusable `~/.moai-cli/hook-config.toml` leaves one `config_error` line (session, host, fixed
/// reason, time) in the engine log when `.moai-cli/` exists (#2391 N1): no path and no config
/// text. Best effort, never changes the answer.
fn record_config_error(
    input: &Value,
    cwd: &Path,
    now_unix_ms: u128,
    host: Host,
    problem: crate::hook_config::ConfigProblem,
) {
    let line = json!({
        "schema_version": 1,
        "kind": "config_error",
        "host": host.name(),
        "session": session_key(input),
        "reason": problem.reason(),
        "unix_ms": now_unix_ms as u64,
    });
    let _ = crate::hooks::append_receipt_line(cwd, ENGINE_LOG, &line);
}

/// Codex rendering of a verdict, from the Codex hooks page only
/// (https://learn.chatgpt.com/docs/hooks, read 2026-11-10), not copied from the Claude shapes:
/// - PreToolUse deny: `hookSpecificOutput` with `hookEventName: "PreToolUse"`,
///   `permissionDecision: "deny"` and `permissionDecisionReason`. `"ask"`, `"approve"`,
///   `continue: false` and `stopReason` are "parsed but not supported yet" there — the hook run
///   is marked failed and the tool call continues — so none of them is ever used to block.
/// - SessionStart context: `hookSpecificOutput.additionalContext` "is added as extra developer
///   context".
/// - Stop: "expects JSON on stdout when it exits 0"; `decision: "block"` with `reason` makes
///   Codex continue with a new continuation prompt. Allow prints nothing ("Exit 0 with no
///   output is treated as success" — stated generally, not for Stop; recorded as unverified).
/// - PostToolUse: "Plain text on stdout is ignored"; tool-observe prints nothing.
fn render_codex(event: Event, verdict: Verdict) -> Answer {
    let stdout = match (event, verdict) {
        (_, Verdict::Allow) => String::new(),
        // Codex parses `ask` but does not support it (the hook run fails and the call
        // continues), so a confirmation becomes a deny the user can act on outside the agent.
        (Event::ToolPreflight, Verdict::Ask(reason)) => json!({
            "hookSpecificOutput": {
                "hookEventName": "PreToolUse",
                "permissionDecision": "deny",
                "permissionDecisionReason": format!("{reason} {CODEX_ASK_SUFFIX}"),
            }
        })
        .to_string(),
        (Event::ToolPreflight, Verdict::Deny(reason)) => json!({
            "hookSpecificOutput": {
                "hookEventName": "PreToolUse",
                "permissionDecision": "deny",
                "permissionDecisionReason": reason,
            }
        })
        .to_string(),
        (Event::SessionCheck, Verdict::Context(text)) => json!({
            "hookSpecificOutput": {"hookEventName": "SessionStart", "additionalContext": text}
        })
        .to_string(),
        (Event::TurnValidation, Verdict::Deny(reason)) => {
            json!({"decision": "block", "reason": reason}).to_string()
        }
        _ => return engine_error("policy returned a verdict this duty cannot carry"),
    };
    Answer {
        stdout,
        stderr: String::new(),
        exit: 0,
    }
}

/// What each duty covers on `host`: the host event, the tools it acts on, how it answers and
/// what it never uses. `hooks-plan` reports it next to the plan so the coverage difference
/// between hosts is visible before anything is installed. Source pages as in the renderers.
pub fn capabilities(host: &str) -> Value {
    match Host::parse(host) {
        Some(Host::Claude) => json!({
            "host": "claude",
            "source": "https://code.claude.com/docs/en/hooks (read 2026-11-10)",
            "session-check": {"event": "SessionStart", "answer": "hookSpecificOutput.additionalContext", "blocks": false},
            "tool-preflight": {"event": "PreToolUse", "guards": [
                                   "Bash: git commit on an issue/<n> branch without a claim token",
                                   "Bash: dangerous rm targets, force push/delete of main or master, mkfs, dd onto a device, fork bomb, DROP DATABASE/SCHEMA or TRUNCATE TABLE through a database client, commit hook bypass",
                                   "Write/Edit/MultiEdit/NotebookEdit: .git internals, credential directories, private-key file names, private-key or GitHub-token content",
                                   BASH_LOCK_GUARD],
                               "opt_in": OPT_IN_RULES,
                               "not_guarded": ["writes outside the project unless [hooks.outside_project_writes] is enabled", "shell redirections into protected paths", BASH_LOCK_OPEN, "MCP tools"],
                               "answer": "hookSpecificOutput.permissionDecision=deny",
                               "confirm": {"answer": "hookSpecificOutput.permissionDecision=ask", "commands": CONFIRM_COMMANDS, "modes": CONFIRM_MODES,
                                           "note": "a deny anywhere on the line wins; ask in a bypass-permissions or -p session is not documented"},
                               "never_uses": ["allow", "defer", "decision/reason (deprecated)"]},
            "tool-observe": {"events": ["PostToolUse", "PostToolUseFailure"], "records": "observe line", "blocks": false},
            "turn-validation": {"event": "Stop", "rules": STOP_RULES.len(), "rule_ids": stop_rule_ids(), "answer": "decision=block + reason",
                                "modes": STOP_MODES, "limits": AC_VERIFY_LIMITS,
                                "host_cap": "8 consecutive continuations (CLAUDE_CODE_STOP_HOOK_BLOCK_CAP)", "moai_cap": MOAI_STOP_BLOCK_CAP},
            "edit-diagnostics": {"event": "PostToolUse", "matcher": crate::edit_diagnostics::MATCHER,
                                 "opt_in": EDIT_OPT_IN, "generated": "only while the setting is enabled",
                                 "answer": "hookSpecificOutput.additionalContext", "blocks": false,
                                 "not_covered": ["a blocking mode (slice S2)", "language servers (slice S3)", "files outside the project or behind a link", "edits made by Bash"]},
            "on_engine_error": "exit 1: non-blocking by the host's rules",
        }),
        Some(Host::Codex) => json!({
            "host": "codex",
            "source": "https://learn.chatgpt.com/docs/hooks (read 2026-11-10)",
            "trust": "non-managed hooks run only after the user reviews and trusts the exact hook definition",
            "session-check": {"event": "SessionStart", "answer": "hookSpecificOutput.additionalContext", "blocks": false},
            "tool-preflight": {"event": "PreToolUse", "guards": [
                                   "Bash: git commit on an issue/<n> branch without a claim token",
                                   "Bash: dangerous rm targets, force push/delete of main or master, mkfs, dd onto a device, fork bomb, DROP DATABASE/SCHEMA or TRUNCATE TABLE through a database client, commit hook bypass",
                                   "apply_patch: .git internals, credential directories, private-key file names (paths from '*** Add/Update/Delete File:' and '*** Move to:' lines, a format the docs do not define; an unreadable patch is not path-checked), private-key or GitHub-token content in added lines",
                                   BASH_LOCK_GUARD],
                               "opt_in": OPT_IN_RULES,
                               "not_guarded": ["writes outside the project unless [hooks.outside_project_writes] is enabled", "shell redirections into protected paths", BASH_LOCK_OPEN, "MCP tools"],
                               "answer": "hookSpecificOutput.permissionDecision=deny",
                               "confirm": {"answer": "hookSpecificOutput.permissionDecision=deny (Codex has no ask)", "commands": CONFIRM_COMMANDS, "modes": CONFIRM_MODES},
                               "never_uses": ["ask", "approve", "continue=false", "stopReason", "allow+updatedInput"]},
            "tool-observe": {"events": ["PostToolUse"], "records": "observe line", "blocks": false,
                             "note": "Codex has no PostToolUseFailure; apply_patch has no file_path, so its path is not recorded",
                             "outcome": "PostToolUse also runs after a failed Bash command, so the event never means success: outcome is unknown unless a Bash tool_response carries an integer exit_code (a field the docs do not name)"},
            "turn-validation": {"event": "Stop", "rules": STOP_RULES.len(), "rule_ids": stop_rule_ids(), "answer": "decision=block + reason",
                                "modes": STOP_MODES, "limits": AC_VERIFY_LIMITS,
                                "host_cap": "not documented", "moai_cap": MOAI_STOP_BLOCK_CAP,
                                "unverified": "empty stdout on exit 0 for Stop (stated generally, not for Stop)"},
            "edit-diagnostics": {"supported": false, "reason": "slice S1 is Claude Code only; Codex apply_patch is slice S2"},
            "on_engine_error": "exit 1: not documented for command hooks (unverified)",
        }),
        Some(Host::Omp) => json!({
            "host": "omp",
            "source": "https://github.com/can1357/oh-my-pi docs/extensions.md, docs/extension-loading.md, extensibility/extensions/types.ts (read 2026-11-13)",
            "adapter": "extension module .omp/extensions/moai-hooks.ts calling hook-dispatch --host omp (OMP runs no command hooks)",
            "session-check": {"event": "before_agent_start", "answer": "{context} added to systemPrompt", "blocks": false},
            "tool-preflight": {"event": "tool_call", "guards": [
                                   "bash: git commit on an issue/<n> branch without a claim token",
                                   "bash: dangerous rm targets, force push/delete of main or master, mkfs, dd onto a device, fork bomb, DROP DATABASE/SCHEMA or TRUNCATE TABLE through a database client, commit hook bypass",
                                   BASH_LOCK_GUARD],
                               "opt_in": OPT_IN_RULES,
                               "not_guarded": ["edit/write paths and content: OMP's edit input is not the Claude file_path/content shape, so it is not translated", "writes outside the project unless [hooks.outside_project_writes] is enabled", BASH_LOCK_OPEN, "MCP and extension tools"],
                               "answer": "{block: true, reason} (reason reaches the model as the tool error)",
                               "confirm": {"answer": "block (OMP has no ask)", "commands": CONFIRM_COMMANDS, "modes": CONFIRM_MODES},
                               "on_module_error": "the module lets the call through when moai is missing, times out or exits 1, matching the non-blocking engine error of the other hosts; OMP itself treats a throwing handler as a block"},
            "tool-observe": {"events": ["tool_result"], "records": "observe line", "blocks": false,
                             "outcome": "tool_result.isError"},
            "turn-validation": {"supported": false, "event": "session_stop", "answer": "{decision: \"block\", reason} (OMP 18.18.7 extensions.md)",
                                "reason": "not wired: the extension module does not subscribe to session_stop, so no turn-validation rule runs on OMP"},
            "edit-diagnostics": {"supported": false, "reason": "slice S1 is Claude Code only; OMP tool_result is slice S2"},
            "on_engine_error": "exit 1: the module does not block",
        }),
        None => json!({"error": "unknown hook host; must be claude, codex or omp"}),
    }
}

fn allow_quietly() -> Answer {
    Answer {
        stdout: String::new(),
        stderr: String::new(),
        exit: 0,
    }
}

/// Claude Code rendering of a verdict (documented shapes only).
fn render_claude(event: Event, verdict: Verdict) -> Answer {
    let stdout = match (event, verdict) {
        (_, Verdict::Allow) => String::new(),
        // "`ask` prompts the user to confirm" (code.claude.com/docs/en/hooks, read 2026-11-11);
        // how a bypass-permissions or `-p` session treats it is not documented.
        (Event::ToolPreflight, Verdict::Ask(reason)) => json!({
            "hookSpecificOutput": {
                "hookEventName": "PreToolUse",
                "permissionDecision": "ask",
                "permissionDecisionReason": reason,
            }
        })
        .to_string(),
        (Event::ToolPreflight, Verdict::Deny(reason)) => json!({
            "hookSpecificOutput": {
                "hookEventName": "PreToolUse",
                "permissionDecision": "deny",
                "permissionDecisionReason": reason,
            }
        })
        .to_string(),
        (Event::SessionCheck, Verdict::Context(text)) => json!({
            "hookSpecificOutput": {"hookEventName": "SessionStart", "additionalContext": text}
        })
        .to_string(),
        // Stop decision control is top-level: `decision: "block"` keeps Claude working and
        // `reason` (required) tells it why.
        (Event::TurnValidation, Verdict::Deny(reason)) => {
            json!({"decision": "block", "reason": reason}).to_string()
        }
        // A verdict the duty cannot carry is a bug in the policy core, reported, not guessed.
        _ => return engine_error("policy returned a verdict this duty cannot carry"),
    };
    Answer {
        stdout,
        stderr: String::new(),
        exit: 0,
    }
}

/// The sentences a session that was just compacted gets (#3015 C6). Both hosts run the
/// `SessionStart` hook again after a compaction with `source: "compact"` (Codex hooks page,
/// https://learn.chatgpt.com/docs/hooks; Claude Code hooks page,
/// https://code.claude.com/docs/en/hooks; both read 2026-11-15); `PreCompact` and `PostCompact`
/// cannot put text into the context on either host. The text is fixed: it names no instruction
/// the user gave and echoes nothing from the input. The hook reads no comment, claim or pull
/// request, so it cannot say that an instruction is done; it says how to find out (the rule of
/// `skills/moai/workflows/mode-leader.md`: run an instruction again only when no comment, claim
/// or pull request shows it done) and that a standing instruction still applies (audit B1 of
/// #3075).
const COMPACTED: &str =
    "This session was just compacted. User instructions you can still see may already be done: \
     before you act on one, check whether a comment, a claim or a pull request already shows it \
     done, and run it again only when none does. A standing instruction (a rule meant to stay \
     in force) still applies.";
/// Codex keeps the `/goal` as host state; the sentence points at it and says nothing the host
/// has not been shown to keep (whether `/goal` survives a compaction is not measured).
const COMPACTED_CODEX_GOAL: &str = "Check the active /goal before you continue.";

/// The conversation language the user chose, as one sentence for session-check (card C-87 S1):
/// the project's `language.user_chat` (`ui_text::profile_chat`: the profile of the project that
/// holds `cwd`, below HOME), else the code the user recorded in HOME (`language-set --home`,
/// the installer's choice). Never guessed: with neither, or a profile that does not read (an
/// unsupported code makes it unreadable) and nothing in HOME, there is no sentence.
fn conversation_line(cwd: &Path, home: Option<&Path>) -> Option<String> {
    let (code, from) = match crate::ui_text::profile_chat(cwd, home) {
        Some(code) => (code, "the project's `language.user_chat`"),
        None => (
            crate::language::home_conversation_supported(home?)?,
            "the language recorded in HOME",
        ),
    };
    let name = crate::language::english_name(&code)?;
    Some(format!(
        "Conversation language: {name} (`{code}`), from {from}: talk with the user in it and \
         relay subagent reports to the user in it."
    ))
}

/// session-check: a short, fixed-vocabulary state summary for the model. Never blocks. After a
/// compaction (`source: "compact"`, exactly that string) it adds [`COMPACTED`] and, if the work
/// has a card, points at its newest `moai-resume:` comment; any other `source`, or none, gives
/// the text it always did.
fn session_check(host: Host, input: &Value, cwd: &Path) -> Verdict {
    let state_dir = match std::fs::symlink_metadata(crate::state_root::path(cwd, "")) {
        Ok(m) if m.is_dir() => "present",
        Ok(_) => "not a directory",
        Err(_) => "absent",
    };
    let mut card_issue = None;
    let lane = match crate::claim_verify::current_branch(cwd)
        .map(|b| crate::claim_verify::lane_branch(&b))
    {
        Some(crate::claim_verify::LaneBranch::Issue(n)) => {
            card_issue = Some(n);
            format!("Issue #{n}")
        }
        Some(crate::claim_verify::LaneBranch::Unparseable) => "unparseable issue/ branch".into(),
        Some(crate::claim_verify::LaneBranch::NotLane) => "none".into(),
        None => "none (detached HEAD or not a git worktree)".into(),
    };
    let record = if crate::claim_verify::has_local_record(cwd, cwd) {
        "present"
    } else {
        "absent"
    };
    let mut text = format!(
        "MoAI session-check (moai {}): project state directory {}: {state_dir}; lane branch: \
         {lane}; local claim record: {record}. Read locally; no network was used.",
        env!("CARGO_PKG_VERSION"),
        crate::state_root::dir(cwd),
    );
    // Card C-87 S1: on every start, resume and compaction, so a compacted session keeps the
    // conversation language the user chose.
    if let Some(line) = conversation_line(cwd, std::env::var_os("HOME").as_deref().map(Path::new)) {
        text.push(' ');
        text.push_str(&line);
    }
    if input.get("source").and_then(Value::as_str) == Some("compact") {
        text.push(' ');
        text.push_str(COMPACTED);
        // Conditional on purpose: the hook does not know that the branch's Issue (or any card)
        // is this session's work.
        let card = match card_issue {
            Some(n) => format!("Issue #{n}"),
            None => "your work has a card Issue".to_string(),
        };
        let lead = if card_issue.is_some() {
            format!("If {card} is your card")
        } else {
            format!("If {card}")
        };
        text.push_str(&format!(
            " {lead}, its newest `moai-resume:` comment says where you stopped (read it there; \
             this hook did not)."
        ));
        if host == Host::Codex {
            text.push(' ');
            text.push_str(COMPACTED_CODEX_GOAL);
        }
    }
    Verdict::Context(text)
}

/// tool-preflight, stage 1 policy: a Bash `git commit` on a lane branch `issue/<n>-…` that has
/// no local claim record must carry `Moai-Claim: <n>#<id>` in its command line (the decided
/// lane rule; AGENTS.local.md and the MODE-AGENT workflow step 6). Everything else is allowed
/// here; a forge check of the claim stays with `claim-verify --hook`.
/// tool-preflight with the confirmation class (B2): any deny — a guard or the claim token, in
/// any segment of the line — wins; only a call that would be allowed is turned into `Ask`
/// (Claude's own order is deny > defer > ask > allow). The direct-merge confirmation is a rule
/// with a mode (`[hooks.modes]`, [`apply_rule_mode`]); the older confirmations are not.
fn tool_preflight(
    input: &Value,
    cwd: &Path,
    config: &crate::hook_config::HookConfig,
    host: Host,
    now_unix_ms: u128,
    fixed_reason: &mut Option<&'static str>,
) -> Verdict {
    let verdict = tool_preflight_deny(input, cwd, config, fixed_reason);
    if verdict != Verdict::Allow || input["tool_name"] != json!("Bash") {
        return verdict;
    }
    let command = input["tool_input"]["command"].as_str().unwrap_or("");
    let dir = input["cwd"]
        .as_str()
        .map(std::path::PathBuf::from)
        .unwrap_or_else(|| cwd.to_path_buf());
    if let Some(reason) = crate::hook_guards::bash_confirm(command) {
        return Verdict::Ask(reason.to_string());
    }
    match crate::hook_guards::merge_confirm(command, &dir) {
        Some(reason) => apply_rule_mode(
            MERGE_CONFIRM_RULE,
            reason,
            input,
            cwd,
            config,
            host,
            now_unix_ms,
        ),
        None => Verdict::Allow,
    }
}

/// The rule id of the direct-merge confirmation in `[hooks.modes]` and in the engine log.
pub const MERGE_CONFIRM_RULE: &str = "merge_confirm";

/// How many times one session may be refused by one rule before the next call goes through
/// (leader decision, C-86 note 2346): the third refusal is released.
pub const RULE_BLOCK_CAP: usize = 2;

/// Applies the mode `[hooks.modes]` gives `rule` to a refusal whose fixed sentence is
/// `reason` (SPEC-20261119-enforce-merge-confirm REQ-6…REQ-11):
/// - `off` → allow, nothing written;
/// - `observe` (the default) → allow, one `would_block` line (rule, fixed reason, host,
///   session, call, time) and no command, path or content;
/// - `enforce` → on Claude the confirmation, as before, never counted (a person answers it);
///   where the host renders it as a refusal (Codex, OMP) one `rule_block` line is written
///   first — a refusal that cannot be written down goes through, because it could not be
///   counted — and once the session already has [`RULE_BLOCK_CAP`] such lines for the rule the
///   call goes through with one `rule_release` line that carries the reason. A call that names
///   no session ([`UNKNOWN_SESSION`]) is refused every time: there is no session to count in.
fn apply_rule_mode(
    rule: &'static str,
    reason: &'static str,
    input: &Value,
    cwd: &Path,
    config: &crate::hook_config::HookConfig,
    host: Host,
    now_unix_ms: u128,
) -> Verdict {
    use crate::hook_config::RuleMode;
    let line = |kind: &str| {
        json!({
            "schema_version": 1,
            "kind": kind,
            "rule": rule,
            "reason": reason,
            "host": host.name(),
            "session": session_key(input),
            "call": call_key(input),
            "unix_ms": now_unix_ms as u64,
        })
    };
    match config.mode_of(rule) {
        RuleMode::Off => Verdict::Allow,
        RuleMode::Observe => {
            let _ = crate::hooks::append_receipt_line(cwd, ENGINE_LOG, &line("would_block"));
            Verdict::Allow
        }
        RuleMode::Enforce if host == Host::Claude => Verdict::Ask(reason.to_string()),
        RuleMode::Enforce => {
            let home = crate::hooks::env_home();
            let session = session_key(input);
            // A host that names no session (OMP sends none, Codex may send an empty one) has
            // nothing to count refusals in: every such call would share one count across all
            // sessions of the project, so the refusal is never released there.
            let strikes = rule_blocks_in(home.as_deref(), cwd, &session, rule);
            if session != UNKNOWN_SESSION && strikes >= RULE_BLOCK_CAP {
                let _ = crate::hooks::append_receipt_line(cwd, ENGINE_LOG, &line("rule_release"));
                return Verdict::Allow;
            }
            match crate::hooks::append_receipt_line(cwd, ENGINE_LOG, &line("rule_block")) {
                Ok(()) => Verdict::Ask(reason.to_string()),
                Err(_) => Verdict::Allow,
            }
        }
    }
}

/// How many `rule_block` lines of `rule` the engine log holds for `session`. Reads the bounded
/// logs only (project side, then the HOME side); a missing or unreadable log counts as zero.
fn rule_blocks_in(home: Option<&Path>, cwd: &Path, session: &str, rule: &str) -> usize {
    crate::hooks::receipt_lines_in(home, cwd, ENGINE_LOG)
        .iter()
        .filter(|v| {
            v["kind"] == json!("rule_block")
                && v["rule"] == json!(rule)
                && v["session"] == json!(session)
        })
        .count()
}

/// `fixed_reason` is set to the guard's sentence when a guard with fixed text refuses the
/// call; a refusal whose sentence is built at run time (commit identity, claim token, a line
/// that commits in more than one folder) leaves it unset.
fn tool_preflight_deny(
    input: &Value,
    cwd: &Path,
    config: &crate::hook_config::HookConfig,
    fixed_reason: &mut Option<&'static str>,
) -> Verdict {
    let command = input["tool_input"]["command"].as_str().unwrap_or("");
    let tool = input["tool_name"].as_str().unwrap_or("");
    // Stage 5 text guards (#2391 A1–A4) come first: they read only the input, so a command they
    // do not refuse still spawns no `git` unless it is a commit.
    let write_dir = input["cwd"]
        .as_str()
        .map(std::path::PathBuf::from)
        .unwrap_or_else(|| cwd.to_path_buf());
    let guard = if tool == "Bash" {
        crate::hook_guards::bash_guard(command)
            .or_else(|| crate::hook_guards::bash_lock_guard(command, &write_dir))
    } else {
        crate::hook_guards::write_guard(tool, &input["tool_input"], &write_dir)
    };
    if let Some(reason) = guard {
        *fixed_reason = Some(reason);
        return Verdict::Deny(reason.to_string());
    }
    // Stage 7 opt-in rules (B3, B9; decision 6779829826): off unless ~/.moai-cli/hook-config.toml turns
    // them on; a missing or broken config leaves them off.
    let opt_in = if tool == "Bash" {
        config
            .primary
            .as_ref()
            .and_then(|r| crate::hook_guards::primary_checkout_guard(command, &write_dir, r))
    } else {
        config.outside.as_ref().and_then(|r| {
            crate::hook_guards::outside_project_guard(tool, &input["tool_input"], &write_dir, r)
        })
    };
    if let Some(reason) = opt_in {
        *fixed_reason = Some(reason);
        return Verdict::Deny(reason.to_string());
    }
    // #2878 E3 (opt-in): no commit under a test identity.
    if tool == "Bash" {
        if let Some(reason) = config
            .identity
            .as_ref()
            .and_then(|r| crate::hook_guards::commit_identity_guard(command, &write_dir, r))
        {
            return Verdict::Deny(reason);
        }
    }
    if tool != "Bash" || !crate::claim_verify::is_commit(command) {
        return Verdict::Allow;
    }
    // The command runs in the session's directory unless the input names another one.
    let dir = input["cwd"]
        .as_str()
        .map(std::path::PathBuf::from)
        .unwrap_or_else(|| cwd.to_path_buf());
    let Ok(dir) = std::fs::canonicalize(dir) else {
        return Verdict::Allow;
    };
    // #2673: the commit runs in the folder `git -C` names, which may be a lane worktree even
    // when the session sits in the primary checkout.
    let dir = match crate::claim_verify::git_target_dir(command, &dir, &["commit"]) {
        Ok(d) => d.unwrap_or(dir),
        Err(why) => return Verdict::Deny(format!("MoAI tool-preflight: {why}.")),
    };
    let Some(branch) = crate::claim_verify::current_branch(&dir) else {
        return Verdict::Allow;
    };
    let crate::claim_verify::LaneBranch::Issue(issue) = crate::claim_verify::lane_branch(&branch)
    else {
        return Verdict::Allow;
    };
    // Cost: the token check reads only the command text, while the local-record lookup spawns
    // two `git rev-parse` calls to find the journal. Both outcomes allow, so the cheap check
    // goes first and a commit that carries its token never pays for the lookup.
    let gap = match crate::claim_verify::commit_claim_id(command, issue) {
        Ok(_) => return Verdict::Allow,
        Err(gap) => gap,
    };
    if crate::claim_verify::has_local_record(cwd, &dir) {
        return Verdict::Allow;
    }
    // SPEC-20261117-commit-guard-users: a lane safety only; no live lane session in this
    // clone means an ordinary user, whose commit is not refused.
    if !crate::claim_verify::lane_clone(&dir) {
        return Verdict::Allow;
    }
    Verdict::Deny(format!(
        "MoAI tool-preflight: {}; put `Moai-Claim: {issue}#<claim comment id>` in the -m \
         message.",
        gap.reason(issue)
    ))
}

/// The log both stage-2 duties share, under `<cwd>/.moai-cli/receipts/`: tool-observe appends
/// `observe` lines, turn-validation appends `stop_block` lines and counts them per session.
pub const ENGINE_LOG: &str = "hook-engine.jsonl";

/// SPEC-20261119-enforce-spec-hooks REQ-6: what the Bash side of the GREEN test lock covers,
/// and what it does not (shown in every host's `capabilities`).
const BASH_LOCK_GUARD: &str = "Bash: rm, unlink, git rm, mv or git mv of a test path the GREEN record locks, of a folder holding one or above the work tree, or of a glob (*, ?, [..]) or brace ({a,b}) pattern that names one; each word is also read literally, and the same commands are read behind the structure words ! { } if then elif else fi while until do done wherever a space, ; & | ( ), a quote or a backtick ends them, and inside backticks, busybox and git -C; ANSI-C ($'...') and locale ($ before a double quote) quoting read as plain quotes; a structure word as an argument refuses too (echo do rm tests/x)";
const BASH_LOCK_OPEN: &str = "Bash edits of a locked test other than that rm, unlink, git rm, mv and git mv on literal, glob or brace words: redirections, tee, sed -i, scripts, variables (rm $F), xargs, find -exec and find -delete, extended globs (@(..), +(..)), ~ expansion, more than 256 brace alternatives, braces nested deeper than 4, escape sequences inside $'...', coproc, cp, truncate, dd of=, install, ln -f, rsync --delete, git checkout --, git restore, git stash and git apply, and any command run after a cd";

/// A short, stable key for a host session id: the first 16 hex digits of its sha256. The raw
/// id is host input and never becomes a path or a log value.
fn session_key(input: &Value) -> String {
    match input["session_id"].as_str() {
        Some(id) if !id.is_empty() => {
            let d = crate::digest(id.as_bytes());
            d.trim_start_matches("sha256:")[..16].to_string()
        }
        _ => UNKNOWN_SESSION.into(),
    }
}

/// What [`session_key`] answers for a call that names no session.
const UNKNOWN_SESSION: &str = "unknown";

/// A host field value fit for a log line: bounded, no control characters, no secret shape.
fn loggable(s: &str, limit: usize) -> bool {
    !s.is_empty()
        && s.len() <= limit
        && !s.chars().any(char::is_control)
        && !crate::redact::contains_secret_shape(s)
}

/// The commit `HEAD` names in `dir` (40 hex), or `None` outside git or on an unborn branch.
/// One local `git rev-parse`; the working tree is not digested.
fn head_commit(dir: &Path) -> Option<String> {
    let mut cmd = std::process::Command::new("git");
    crate::identity::scrub_git_env(&mut cmd);
    let out = cmd
        .arg("-C")
        .arg(dir)
        .args(["rev-parse", "--verify", "-q", "HEAD"])
        .stderr(std::process::Stdio::null())
        .output()
        .ok()?;
    let sha = String::from_utf8_lossy(&out.stdout).trim().to_string();
    (out.status.success() && sha.len() == 40 && sha.bytes().all(|b| b.is_ascii_hexdigit()))
        .then_some(sha)
}

/// The tool's target path as the log keeps it: relative to the session directory, `outside`
/// when it leaves that directory, `unrecorded` when it is not a loggable string.
fn target_path(input: &Value, cwd: &Path) -> Value {
    let raw = input["tool_input"]["file_path"]
        .as_str()
        .or_else(|| input["tool_input"]["notebook_path"].as_str());
    let Some(raw) = raw else {
        return Value::Null;
    };
    if !loggable(raw, 1024) {
        return json!("unrecorded");
    }
    let Some(rel) = project_relative(raw, input["cwd"].as_str(), cwd) else {
        return json!("outside");
    };
    match rel.to_str() {
        Some(s) if loggable(s, 256) => json!(s),
        _ => json!("unrecorded"),
    }
}

/// `raw` as a path below the project folder `cwd` (the folder the hook runs in, canonical), or
/// `None` when it leaves it (SPEC-20261119-ac-verify-recorded-limits REQ-10). The host names
/// paths its own way, so an absolute path is placed against `cwd` first, then against the
/// host's own `cwd` field when that names the same folder (a link to it), then by the file
/// system (the path's existing part resolved, so a link or `/var` for `/private/var` does not
/// make a project path look outside). `.` and `..` are settled by name: `src/../x.rs` is
/// `x.rs`, and a path that climbs above the folder is outside.
fn project_relative(raw: &str, host_cwd: Option<&str>, cwd: &Path) -> Option<std::path::PathBuf> {
    use std::path::{Component, PathBuf};
    let path = Path::new(raw);
    let below: PathBuf = if path.is_absolute() {
        if let Ok(rel) = path.strip_prefix(cwd) {
            rel.to_path_buf()
        } else if let Some(rel) = host_cwd
            .map(Path::new)
            .filter(|h| h.is_absolute() && std::fs::canonicalize(h).is_ok_and(|c| c == cwd))
            .and_then(|h| path.strip_prefix(h).ok())
        {
            rel.to_path_buf()
        } else {
            canonical_below(path, cwd)?
        }
    } else {
        path.to_path_buf()
    };
    let mut clean = PathBuf::new();
    for part in below.components() {
        match part {
            Component::Normal(name) => clean.push(name),
            Component::CurDir => {}
            Component::ParentDir => {
                if !clean.pop() {
                    return None;
                }
            }
            Component::RootDir | Component::Prefix(_) => return None,
        }
    }
    Some(clean)
}

/// `path` below `cwd` by the file system: the longest existing part of `path` is resolved
/// (links followed), the rest is appended by name; `None` when the result is not below `cwd`.
fn canonical_below(path: &Path, cwd: &Path) -> Option<std::path::PathBuf> {
    let mut rest: Vec<std::ffi::OsString> = Vec::new();
    let mut head = path.to_path_buf();
    loop {
        if let Ok(resolved) = std::fs::canonicalize(&head) {
            let full = rest.iter().rev().fold(resolved, |acc, name| acc.join(name));
            return full.strip_prefix(cwd).ok().map(Path::to_path_buf);
        }
        rest.push(head.file_name()?.to_os_string());
        if !head.pop() {
            return None;
        }
    }
}

/// tool-observe (PostToolUse / PostToolUseFailure): appends one bounded observation line —
/// tool name and class, target path, outcome, the `HEAD` commit and the time. It never blocks
/// and never stores `tool_input` content, `tool_response` or messages. Marking earlier evidence
/// stale waits for a persisted verification receipt (#2392; leader decision, Issue #2391).
/// The observed call's outcome and what it was read from, per host (#2402 audit H3-B1):
/// - Claude Code: `PostToolUse` "runs immediately after a tool completes successfully" and
///   failures arrive as `PostToolUseFailure` — the event name decides.
/// - Codex: `PostToolUse` also runs "after commands that exit with a non-zero status", and the
///   page gives no shape for `tool_response` beyond "Tool-specific output" (read 2026-11-11).
///   The event name therefore never means success. Only a Bash `tool_response` carrying an
///   integer `exit_code` decides (0 success, anything else failure); that field is not named
///   in the docs, so its use is recorded in `outcome_source`. Everything else is `unknown`.
fn observe_outcome(input: &Value, host: Host) -> (&'static str, &'static str) {
    match host {
        Host::Claude if input["hook_event_name"] == json!("PostToolUseFailure") => {
            ("failure", "event")
        }
        Host::Claude => ("success", "event"),
        Host::Codex => match (
            input["tool_name"].as_str(),
            input["tool_response"]["exit_code"].as_i64(),
        ) {
            (Some("Bash"), Some(0)) => ("success", "tool_response.exit_code (undocumented)"),
            (Some("Bash"), Some(_)) => ("failure", "tool_response.exit_code (undocumented)"),
            _ => ("unknown", "none"),
        },
        // OMP `tool_result` carries `isError` (types.ts); the module passes it as
        // `tool_response.is_error`.
        Host::Omp => match input["tool_response"]["is_error"].as_bool() {
            Some(false) => ("success", "tool_result.isError"),
            Some(true) => ("failure", "tool_result.isError"),
            None => ("unknown", "none"),
        },
    }
}

fn tool_observe(input: &Value, cwd: &Path, now_unix_ms: u128, host: Host) -> Verdict {
    // Cost (audit #2400 N3): a project without a real `.moai-cli/` directory records nothing,
    // so return before spawning `git` for the HEAD commit.
    match std::fs::symlink_metadata(crate::state_root::path(cwd, "")) {
        Ok(m) if m.is_dir() => {}
        _ => return Verdict::Allow,
    }
    let tool = input["tool_name"].as_str().unwrap_or("");
    let tool_ok = !tool.is_empty()
        && tool.len() <= 64
        && tool
            .bytes()
            .all(|b| b.is_ascii_alphanumeric() || b == b'_' || b == b'-' || b == b'.');
    let class = match tool {
        "Write" | "Edit" | "MultiEdit" | "NotebookEdit" | "apply_patch" => "write",
        "Bash" => "bash",
        "Read" | "Glob" | "Grep" | "LS" | "WebFetch" | "WebSearch" => "read",
        _ => "other",
    };
    let (outcome, outcome_source) = observe_outcome(input, host);
    let line = json!({
        "schema_version": 1,
        "kind": "observe",
        "host": host.name(),
        "session": session_key(input),
        "call": call_key(input),
        "tool": if tool_ok { json!(tool) } else { json!("unrecorded") },
        "class": class,
        "path": target_path(input, cwd),
        "outcome": outcome,
        "outcome_source": outcome_source,
        "head": head_commit(cwd),
        "unix_ms": now_unix_ms as u64,
    });
    // Best effort: a project without `.moai-cli/`, or a refused log, is simply not recorded.
    let _ = crate::hooks::append_receipt_line(cwd, ENGINE_LOG, &line);
    if class == "bash" && outcome == "success" && ran_to_its_end(input, host) {
        record_ac_run(input, cwd, now_unix_ms);
    }
    Verdict::Allow
}

/// Whether a Bash call that [`observe_outcome`] calls a success also ran to its end, which is
/// what a recorded run of a verify command needs (SPEC-20261119-enforce-ac-verify-recorded REQ-1,
/// tier 2 audit B1). Claude's "success" is only an event name that is not `PostToolUseFailure`,
/// so for Claude the input must say it is a finished `PostToolUse`: a missing or other event name
/// is an unknown state, which is not a success. For every host a call started in the background
/// (`tool_input.run_in_background`, or a `tool_response.backgroundTaskId`, which says the command
/// is still running) and a call the host reports as `tool_response.interrupted` did not run to
/// its end. A flag that is absent, `null` or the plain `false` hides nothing.
///
/// SPEC-20261119-ac-verify-recorded-limits REQ-7 (card C-130, audit N-r1): the host documents
/// `run_in_background`, `interrupted` and `backgroundedByUser` as booleans, `backgroundTaskId` as
/// a string only for a background command and `timedOutAfterMs` as set when Claude Code moved a
/// command to the background itself (code.claude.com/docs/en/hooks.md, agent-sdk/typescript.md,
/// tools-reference.md, read 2026-11-19). A recorder must not count an unfinished call, so every
/// other value a mark can hold (the string "true", the number 1, an empty task id, a task id in
/// snake_case) reads as unfinished, and so does a `tool_response` that is present and not an
/// object (a shape the Bash tool does not document).
fn ran_to_its_end(input: &Value, host: Host) -> bool {
    if host == Host::Claude && input["hook_event_name"] != json!("PostToolUse") {
        return false;
    }
    let response = &input["tool_response"];
    if !response.is_null() && !response.is_object() {
        return false;
    }
    // A flag: present and not `false`/`null` marks the call.
    let flag = |value: &Value| !value.is_null() && value != &json!(false);
    // An id or a time: present and not `null` marks the call.
    let present = |value: &Value| !value.is_null();
    let started_in_background = flag(&input["tool_input"]["run_in_background"])
        || flag(&input["tool_input"]["runInBackground"])
        || present(&response["backgroundTaskId"])
        || present(&response["background_task_id"])
        || present(&response["timedOutAfterMs"])
        || flag(&response["backgroundedByUser"]);
    !started_in_background && !flag(&response["interrupted"])
}

/// The longest Bash command line [`record_ac_run`] compares with a SPEC's `verify:` commands.
const AC_RUN_COMMAND_LIMIT: usize = 4096;

/// SPEC-20261119-enforce-ac-verify-recorded REQ-1, REQ-2: a successful Bash call whose command,
/// white space collapsed, equals a `verify:` command of the SPEC of the branch it ran on leaves
/// one `ac_run` line per matching command (SPEC id, AC id, the command's position, the full
/// HEAD commit, the session, the time) and no command text. Anything unreadable records nothing.
fn record_ac_run(input: &Value, cwd: &Path, now_unix_ms: u128) {
    let Some(command) = input["tool_input"]["command"].as_str() else {
        return;
    };
    if command.len() > AC_RUN_COMMAND_LIMIT {
        return;
    }
    let wanted = crate::hook_guards::collapse_whitespace(command);
    if wanted.is_empty() {
        return;
    }
    let dir = input["cwd"]
        .as_str()
        .map(std::path::PathBuf::from)
        .unwrap_or_else(|| cwd.to_path_buf());
    let Some(work) = crate::hook_guards::ac_work(&dir) else {
        return;
    };
    for verify in work.commands.iter().filter(|v| v.command == wanted) {
        let line = json!({
            "schema_version": 1,
            "kind": "ac_run",
            "spec": work.spec,
            "ac": verify.ac,
            "cmd": verify.position,
            "head": work.head,
            "session": session_key(input),
            "unix_ms": now_unix_ms as u64,
        });
        let _ = crate::hooks::append_receipt_line(cwd, ENGINE_LOG, &line);
    }
}

/// What a Stop rule sees.
pub struct StopContext<'a> {
    /// The session directory.
    pub cwd: &'a Path,
    /// The host's `stop_hook_active`: this turn is already continuing because of a stop hook.
    pub stop_hook_active: bool,
    /// The key of the session whose turn ends ([`session_key`]).
    pub session: &'a str,
}

/// One turn-validation rule: `check` returns the reason to keep working, or `None`. A rule
/// may only ask for something the model can resolve itself; a condition waiting on an audit,
/// CI, the leader or the user must never be a rule (it would block every turn end).
pub struct StopRule {
    pub id: &'static str,
    pub check: fn(&StopContext) -> Option<String>,
    /// Whether `[hooks.modes]` sets how the rule acts (`off`, `observe` by default, `enforce`);
    /// a rule without a mode asks every time it finds something.
    pub moded: bool,
}

/// The decided Stop rules. A rule is added here only with its decision record (leader decision,
/// Issue #2391): `ac_verify_recorded` has note 2711 of card C-86 (2026-11-19).
pub const STOP_RULES: &[StopRule] = &[StopRule {
    id: AC_VERIFY_RULE,
    check: ac_verify_recorded,
    moded: true,
}];

/// The ids of [`STOP_RULES`], for `capabilities`.
fn stop_rule_ids() -> Vec<&'static str> {
    STOP_RULES.iter().map(|r| r.id).collect()
}

/// The rule id of the AC verify record in `[hooks.modes]` and in the engine log.
pub const AC_VERIFY_RULE: &str = "ac_verify_recorded";

/// How a Stop rule with a mode acts, shown in `capabilities` (SPEC-20261119-enforce-ac-verify-
/// recorded REQ-5).
const STOP_MODES: &str = "per rule in [hooks.modes] of the HOME hook-config.toml (for example ac_verify_recorded = enforce): off does nothing; observe, the default and what any other word reads as, writes one would_block line and lets the turn end; enforce keeps the session working with decision=block, at most moai_cap turns in a row";

/// What the rule `ac_verify_recorded` does not do, shown in `capabilities` (SPEC-20261119-
/// enforce-ac-verify-recorded Risks).
const AC_VERIFY_LIMITS: [&str; 12] = [
    "the generated host settings do not call tool-observe or turn-validation yet, so the rule runs only where hook-dispatch is called directly (a later slice wires them)",
    "a commit changes HEAD, so runs recorded before it no longer count and the verify commands run again",
    "shell writes (sed -i) are not writes, and a redirection is not one either: only the file-writing tools count as the session's last write, and of those a failed one and one outside the project folder do not count while one with an unknown outcome or no path does",
    "a run counts only when the called command must equal the verify command, white space collapsed: a pipe, an env prefix or an added word does not match",
    "the engine log is a plain file: a line can be written by hand, so the rule stops forgetting, not cheating (the auditor checks the rest)",
    "the log has a size limit: past it nothing more is recorded and the rule reads no lines, so it lets the turn end",
    "an AC with manual:, with unknown: or with a <...> placeholder in its verify value is not asked",
    "Stop reads the process folder while the observation reads the cwd the call names: a session whose calls run in another work tree than its process folder is asked about the process folder's SPEC",
    "a verify command over 4096 bytes is never recorded, but Stop still asks for it until the third turn end goes through (two blocks, then a rule_release line)",
    "a detached HEAD, sha256 objects, a reftable and a symbolic ref read as no commit, so the rule switches itself off without saying so",
    "a run from a subfolder of the work tree counts, because the SPEC is found from the top of the work tree the call's cwd is in",
    "on OMP the extension sends only isError, so a call OMP ran in the background (async, notify) or that timed out counts as a finished run",
];

/// The rule `ac_verify_recorded` (SPEC-20261119-enforce-ac-verify-recorded REQ-3, REQ-4,
/// REQ-6): on a SPEC work branch, after a file-writing tool call of this session that did not fail
/// and is not outside the project folder (limits SPEC REQ-10), every runnable
/// `verify:` command of the branch's SPEC needs an `ac_run` line for the current HEAD commit
/// written at or after that write; the reason names the acceptance criteria that lack one.
/// Nothing readable, nothing written by the session or nothing to run gives `None`.
fn ac_verify_recorded(ctx: &StopContext) -> Option<String> {
    let work = crate::hook_guards::ac_work(ctx.cwd)?;
    if work.commands.is_empty() {
        return None;
    }
    let home = crate::hooks::env_home();
    let lines = crate::hooks::receipt_lines_in(home.as_deref(), ctx.cwd, ENGINE_LOG);
    let wrote = lines
        .iter()
        .filter(|l| {
            l["kind"] == json!("observe")
                && l["class"] == json!("write")
                && l["session"] == json!(ctx.session)
                // SPEC-20261119-ac-verify-recorded-limits REQ-10 (leader decision, C-130 note
                // 3387): a failed write and a write outside the project folder are no write to
                // ask about; an unknown outcome and a line with no path count (dropping them
                // would weaken the rule quietly).
                && l["outcome"] != json!("failure")
                && l["path"] != json!("outside")
        })
        .filter_map(|l| l["unix_ms"].as_u64())
        .max()?;
    let mut missing: Vec<&str> = Vec::new();
    for verify in &work.commands {
        let recorded = lines.iter().any(|l| {
            l["kind"] == json!("ac_run")
                && l["spec"] == json!(work.spec)
                && l["ac"] == json!(verify.ac)
                && l["cmd"] == json!(verify.position)
                && l["head"] == json!(work.head)
                && l["unix_ms"].as_u64().is_some_and(|t| t >= wrote)
        });
        if !recorded && !missing.contains(&verify.ac.as_str()) {
            missing.push(&verify.ac);
        }
    }
    if missing.is_empty() {
        return None;
    }
    let shown = missing
        .iter()
        .take(5)
        .copied()
        .collect::<Vec<_>>()
        .join(", ");
    let more = match missing.len() {
        n if n > 5 => format!(" and {} more", n - 5),
        _ => String::new(),
    };
    Some(format!(
        "{} has verify commands with no recorded run at commit {} since your last file write ({shown}{more}). Run each `verify:` command exactly as the SPEC writes it (one command line, nothing added), or tell the user which one cannot run and why.",
        work.spec,
        &work.head[..7],
    ))
}

/// How many times in a row MoAI itself may keep one session working before it lets the turn
/// end. Claude Code's own cap (8 continuations, shared by every stop hook) stays the outer
/// bound; this one keeps MoAI's share small next to a user's `/goal` hook.
pub const MOAI_STOP_BLOCK_CAP: usize = 2;

/// MoAI's consecutive `stop_block` lines for `session` since that session's last `observe`
/// line (a tool call resets the count, as Claude Code's own cap does). Reads the bounded logs
/// only (the project side first, then the HOME side where a project with a fixed id writes,
/// #3009 S4b); a missing or unreadable log counts as zero.
fn consecutive_blocks(cwd: &Path, session: &str) -> usize {
    consecutive_blocks_in(crate::hooks::env_home().as_deref(), cwd, session)
}

fn consecutive_blocks_in(home: Option<&Path>, cwd: &Path, session: &str) -> usize {
    let lines = crate::hooks::receipt_lines_in(home, cwd, ENGINE_LOG);
    let mut count = 0;
    for v in lines.iter().rev() {
        if v["session"] != json!(session) {
            continue;
        }
        match v["kind"].as_str() {
            Some("stop_block") => count += 1,
            Some("observe") => break,
            _ => {}
        }
    }
    count
}

/// turn-validation (Stop): runs `rules` and keeps the session working only when a rule asks
/// and MoAI's own bound allows it:
/// - no rule → allow, nothing read or written (the shipped default);
/// - `stop_hook_active` → allow, always (decision D2 = a, Issue #2389 c6774899518 row 24): the
///   turn is already being continued by a stop hook — a user's `/goal`, another hook, or MoAI's
///   own earlier block, which the field cannot tell apart — and MoAI never competes with it.
///   Both hosts send the field (Claude Code; Codex "Whether this turn was already continued by
///   `Stop`", learn.chatgpt.com/docs/hooks read 2026-11-11). The Go `/goal` Stop link is not
///   carried over (B5 rejected);
/// - MoAI has blocked [`MOAI_STOP_BLOCK_CAP`] times in a row → allow (backstop for a host that
///   does not report a continuation); a rule that would have blocked again writes a
///   `rule_release` line (rule, reason, session) so that the let-go leaves a trace — one per
///   turn end that goes through, so a session that keeps ending turns adds a line each time,
///   inside the log's size limit;
/// - a block that cannot be recorded (no `.moai-cli/`, refused log) → allow, because an
///   unrecorded block could not be bounded.
///
/// Timeout: the host's `timeout` on the generated hook bounds the process; the engine reads
/// only the bounded log and runs no subprocess here.
pub fn turn_validation_with(
    rules: &[StopRule],
    input: &Value,
    cwd: &Path,
    now_unix_ms: u128,
) -> Verdict {
    if rules.is_empty() {
        return Verdict::Allow;
    }
    let active = input["stop_hook_active"] == json!(true);
    let session = session_key(input);
    if active {
        return Verdict::Allow;
    }
    // At the cap the rules still run: a rule that would have blocked again is released with one
    // `rule_release` line (SPEC-20261119-ac-verify-recorded-limits REQ-2), so the silent let-go
    // after MoAI's own bound leaves a trace; a session no rule asks any more leaves none.
    let capped = consecutive_blocks(cwd, &session) >= MOAI_STOP_BLOCK_CAP;
    let ctx = StopContext {
        cwd,
        stop_hook_active: active,
        session: &session,
    };
    for rule in rules {
        if let Some(reason) = (rule.check)(&ctx) {
            // A rule with a mode acts by `[hooks.modes]` (read only now that it would ask):
            // `off` skips it, `observe` (the default) writes one `would_block` line and goes on.
            if rule.moded {
                use crate::hook_config::RuleMode;
                match crate::hook_config::load().config.mode_of(rule.id) {
                    RuleMode::Off => continue,
                    RuleMode::Observe => {
                        let line = json!({
                            "schema_version": 1,
                            "kind": "would_block",
                            "rule": rule.id,
                            "reason": reason,
                            "session": session,
                            "unix_ms": now_unix_ms as u64,
                        });
                        let _ = crate::hooks::append_receipt_line(cwd, ENGINE_LOG, &line);
                        continue;
                    }
                    RuleMode::Enforce => {}
                }
            }
            if capped {
                let line = json!({
                    "schema_version": 1,
                    "kind": "rule_release",
                    "rule": rule.id,
                    "reason": reason,
                    "session": session,
                    "unix_ms": now_unix_ms as u64,
                });
                let _ = crate::hooks::append_receipt_line(cwd, ENGINE_LOG, &line);
                return Verdict::Allow;
            }
            let line = json!({
                "schema_version": 1,
                "kind": "stop_block",
                "session": session,
                "rule": rule.id,
                "unix_ms": now_unix_ms as u64,
            });
            if crate::hooks::append_receipt_line(cwd, ENGINE_LOG, &line).is_err() {
                return Verdict::Allow;
            }
            return Verdict::Deny(format!("MoAI turn-validation ({}): {reason}", rule.id));
        }
    }
    Verdict::Allow
}

#[cfg(test)]
mod tests {
    use super::*;

    /// The fixture rule: always asks to keep working.
    const FAKE: &[StopRule] = &[StopRule {
        id: "fake-always",
        check: |_| Some("fixture rule asks for one more pass".into()),
        moded: false,
    }];

    fn project(with_state: bool) -> tempfile::TempDir {
        let dir = tempfile::TempDir::new().unwrap();
        if with_state {
            std::fs::create_dir(dir.path().join(crate::PROJECT_STATE_DIR)).unwrap();
        }
        dir
    }

    fn stop(active: bool) -> Value {
        json!({"hook_event_name": "Stop", "session_id": "sess-A", "stop_hook_active": active})
    }

    fn log(dir: &Path) -> std::path::PathBuf {
        dir.join(crate::PROJECT_STATE_DIR)
            .join("receipts")
            .join(ENGINE_LOG)
    }

    #[test]
    fn shipped_rule_set_is_the_ac_verify_rule_and_outside_a_spec_branch_never_blocks_or_writes() {
        assert_eq!(stop_rule_ids(), vec!["ac_verify_recorded"]);
        assert!(STOP_RULES.iter().all(|r| r.moded));
        let p = project(true);
        for active in [false, true] {
            assert_eq!(
                turn_validation_with(STOP_RULES, &stop(active), p.path(), 1),
                Verdict::Allow
            );
        }
        assert!(!log(p.path()).exists());
    }

    /// Decision D2 = a (Issue #2389 c6774899518 row 24): once the host is continuing the turn
    /// (`stop_hook_active`), MoAI never blocks — not even after its own block, since the field
    /// cannot tell whose block continued it. MoAI therefore blocks at most once per stop.
    #[test]
    fn fixture_rule_blocks_once_and_never_while_a_continuation_runs() {
        let p = project(true);
        let first = turn_validation_with(FAKE, &stop(false), p.path(), 1);
        assert!(matches!(first, Verdict::Deny(ref r) if r.contains("fake-always")));
        // The host now continues the turn: stop_hook_active is true from here on.
        for t in 2..5 {
            assert_eq!(
                turn_validation_with(FAKE, &stop(true), p.path(), t),
                Verdict::Allow
            );
        }
        let text = std::fs::read_to_string(log(p.path())).unwrap();
        assert_eq!(text.lines().count(), 1, "one recorded block: {text}");
    }

    /// The block cap stays as the backstop for a host that reports `stop_hook_active: false`
    /// again without a tool call in between: two blocks, then allow.
    #[test]
    fn the_cap_bounds_blocks_when_the_host_never_reports_a_continuation() {
        let p = project(true);
        for t in 1..3 {
            assert!(matches!(
                turn_validation_with(FAKE, &stop(false), p.path(), t),
                Verdict::Deny(_)
            ));
        }
        assert_eq!(
            turn_validation_with(FAKE, &stop(false), p.path(), 3),
            Verdict::Allow
        );
    }

    #[test]
    fn a_tool_call_resets_the_count() {
        let p = project(true);
        let _ = turn_validation_with(FAKE, &stop(false), p.path(), 1);
        let _ = turn_validation_with(FAKE, &stop(false), p.path(), 2);
        assert_eq!(
            turn_validation_with(FAKE, &stop(false), p.path(), 3),
            Verdict::Allow
        );
        let post = json!({"hook_event_name": "PostToolUse", "session_id": "sess-A",
                          "tool_name": "Read", "tool_input": {"file_path": "a"}});
        assert_eq!(
            tool_observe(&post, p.path(), 4, Host::Claude),
            Verdict::Allow
        );
        assert!(matches!(
            turn_validation_with(FAKE, &stop(false), p.path(), 5),
            Verdict::Deny(_)
        ));
    }

    #[test]
    fn another_sessions_blocks_do_not_count() {
        let p = project(true);
        let other = json!({"session_id": "sess-B", "stop_hook_active": false});
        let _ = turn_validation_with(FAKE, &other, p.path(), 1);
        let other = json!({"session_id": "sess-B", "stop_hook_active": true});
        let _ = turn_validation_with(FAKE, &other, p.path(), 2);
        assert!(matches!(
            turn_validation_with(FAKE, &stop(false), p.path(), 3),
            Verdict::Deny(_)
        ));
    }

    #[test]
    fn a_continuation_from_another_stop_hook_is_not_contested() {
        let p = project(true);
        assert_eq!(
            turn_validation_with(FAKE, &stop(true), p.path(), 1),
            Verdict::Allow
        );
        assert!(!log(p.path()).exists());
    }

    #[test]
    fn an_unrecordable_block_is_not_made() {
        let p = project(false);
        assert_eq!(
            turn_validation_with(FAKE, &stop(false), p.path(), 1),
            Verdict::Allow
        );
    }

    #[test]
    fn stop_block_renders_top_level_decision_and_reason_only() {
        let a = render_claude(Event::TurnValidation, Verdict::Deny("why".into()));
        assert_eq!(a.exit, 0);
        let v: Value = serde_json::from_str(&a.stdout).unwrap();
        assert_eq!(v, json!({"decision": "block", "reason": "why"}));
    }

    #[test]
    fn codex_renders_only_documented_codex_shapes() {
        let stop = render_codex(Event::TurnValidation, Verdict::Deny("why".into()));
        assert_eq!(
            serde_json::from_str::<Value>(&stop.stdout).unwrap(),
            json!({"decision": "block", "reason": "why"})
        );
        let deny = render_codex(Event::ToolPreflight, Verdict::Deny("no".into()));
        assert_eq!(
            serde_json::from_str::<Value>(&deny.stdout).unwrap(),
            json!({"hookSpecificOutput": {"hookEventName": "PreToolUse",
                   "permissionDecision": "deny", "permissionDecisionReason": "no"}})
        );
        let ctx = render_codex(Event::SessionCheck, Verdict::Context("c".into()));
        assert_eq!(
            serde_json::from_str::<Value>(&ctx.stdout).unwrap(),
            json!({"hookSpecificOutput": {"hookEventName": "SessionStart", "additionalContext": "c"}})
        );
        for event in [
            Event::SessionCheck,
            Event::ToolPreflight,
            Event::ToolObserve,
            Event::TurnValidation,
        ] {
            assert_eq!(render_codex(event, Verdict::Allow), allow_quietly());
        }
        // tool-observe never answers with a decision on Codex either.
        assert_eq!(
            render_codex(Event::ToolObserve, Verdict::Deny("x".into())).exit,
            ENGINE_ERROR_EXIT
        );
    }

    #[test]
    fn session_ids_are_hashed_never_logged_raw() {
        let p = project(true);
        let _ = turn_validation_with(FAKE, &stop(false), p.path(), 1);
        let text = std::fs::read_to_string(log(p.path())).unwrap();
        assert!(!text.contains("sess-A"), "{text}");
        assert_eq!(session_key(&json!({})), "unknown");
        assert_eq!(session_key(&stop(false)).len(), 16);
    }

    /// The block count reads the project-side log (older) and then the HOME-side log (newer),
    /// so an `observe` on the newer side resets blocks written on the older side (#3009 S4b).
    #[test]
    fn the_block_count_reads_the_project_log_and_then_the_home_log() {
        const ID: &str = "abcdefghijklmnopqrstuvwxyz";
        let p = project(true);
        std::fs::write(
            crate::state_root::path(p.path(), "project.yaml"),
            format!("{{\"project_id\":\"{ID}\"}}"),
        )
        .unwrap();
        let home = tempfile::tempdir().unwrap();
        std::fs::create_dir(crate::home_root::path(home.path())).unwrap();
        let session = "s";
        let put = |kind: &str, in_home: bool| {
            let line = json!({"kind": kind, "session": session});
            if in_home {
                crate::hooks::append_receipt_line_in(
                    Some(home.path()),
                    p.path(),
                    ENGINE_LOG,
                    &line,
                )
                .unwrap();
            } else {
                crate::hooks::append_receipt_line_in(None, p.path(), ENGINE_LOG, &line).unwrap();
            }
        };
        let count = || consecutive_blocks_in(Some(home.path()), p.path(), session);
        assert_eq!(count(), 0);
        put("stop_block", false);
        put("stop_block", false);
        assert_eq!(count(), 2, "blocks written before the switch still count");
        put("stop_block", true);
        assert_eq!(count(), 3);
        put("observe", true);
        assert_eq!(count(), 0, "a newer observe resets older blocks");
        put("stop_block", true);
        assert_eq!(count(), 1);
        // Another session's lines never count.
        assert_eq!(
            consecutive_blocks_in(Some(home.path()), p.path(), "other"),
            0
        );
        // Without the HOME the HOME-side lines are not seen.
        assert_eq!(consecutive_blocks_in(None, p.path(), session), 2);
    }

    /// SPEC-20261119-ac-verify-recorded-limits REQ-10 (audit of 0ca2f8892b, findings 1 and 2):
    /// the project is the folder the hook runs in (canonical), but the host names paths its own
    /// way: through a link to the project folder, in its own `cwd` field, or with `..` that stays
    /// inside. Those are project paths; a path that really leaves the folder is `outside`, and one
    /// the log cannot keep is `unrecorded`.
    #[cfg(unix)]
    #[test]
    fn target_path_places_a_project_path_the_host_names_through_a_link_or_dots() {
        let tmp = tempfile::TempDir::new().unwrap();
        let base = std::fs::canonicalize(tmp.path()).unwrap();
        let real = base.join("proj");
        std::fs::create_dir_all(real.join("src")).unwrap();
        std::fs::create_dir_all(base.join("other")).unwrap();
        let link = base.join("link");
        std::os::unix::fs::symlink(&real, &link).unwrap();
        let path = |file: &str, host_cwd: &Path| {
            target_path(
                &json!({"cwd": host_cwd, "tool_input": {"file_path": file}}),
                &real,
            )
        };
        let abs = |p: std::path::PathBuf| p.to_str().unwrap().to_string();
        // Plain.
        assert_eq!(path(&abs(real.join("src/x.rs")), &real), json!("src/x.rs"));
        assert_eq!(path("src/x.rs", &real), json!("src/x.rs"));
        assert_eq!(path("./src/x.rs", &real), json!("src/x.rs"));
        // Through a link to the project: the host's cwd is the link, or only the path is.
        assert_eq!(path(&abs(link.join("src/x.rs")), &link), json!("src/x.rs"));
        assert_eq!(path(&abs(link.join("src/x.rs")), &real), json!("src/x.rs"));
        assert_eq!(
            path(&abs(link.join("src/new.rs")), &real),
            json!("src/new.rs")
        );
        // `..` that stays inside.
        assert_eq!(path("src/../x.rs", &real), json!("x.rs"));
        assert_eq!(path(&abs(real.join("src/../x.rs")), &real), json!("x.rs"));
        // Really outside.
        assert_eq!(path("../x.rs", &real), json!("outside"));
        assert_eq!(path("src/../../x.rs", &real), json!("outside"));
        assert_eq!(
            path(&abs(real.join("src/../../x.rs")), &real),
            json!("outside")
        );
        assert_eq!(path(&abs(base.join("other/x.rs")), &real), json!("outside"));
        assert_eq!(path("/tmp/moai-c130-scratch.txt", &real), json!("outside"));
        // The host's own cwd field does not make a path outside the hook's folder inside.
        assert_eq!(
            path(&abs(base.join("other/x.rs")), &base.join("other")),
            json!("outside")
        );
        // What the log cannot keep.
        assert_eq!(path("", &real), json!("unrecorded"));
        assert_eq!(path("src/x\u{7}.rs", &real), json!("unrecorded"));
        assert_eq!(
            path(&format!("src/{}.rs", "a".repeat(1100)), &real),
            json!("unrecorded")
        );
        assert_eq!(
            target_path(&json!({"tool_input": {}}), &real),
            Value::Null,
            "no path"
        );
    }
}
