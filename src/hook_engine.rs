//! Hook engine, stage 1 (rc.2 #2391): one policy core for the four MoAI hook duties
//! (session-check, tool-preflight, tool-observe, turn-validation) with a native adapter per
//! host. Stage 1 implements `session-check` and `tool-preflight` for Claude Code; the Codex
//! adapter and the other two duties are later stages and answer as engine errors.
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
//! - engine error (unsupported host or event, malformed input, wrong event name): exit 1 with
//!   the reason on stderr and nothing on stdout — a non-blocking error by the host's rules.
//!   Whether an engine error should block (fail-closed, the host's `onFailure: "block"`) is a
//!   user decision and is not taken here.
//!
//! Cost: no network on any path. The only reads are the hook stdin, `git symbolic-ref` in the
//! session's working directory and this clone's local claim journal. Text that reaches the
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
}

impl Event {
    pub fn parse(word: &str) -> Option<Event> {
        match word {
            "session-check" => Some(Event::SessionCheck),
            "tool-preflight" => Some(Event::ToolPreflight),
            "tool-observe" => Some(Event::ToolObserve),
            "turn-validation" => Some(Event::TurnValidation),
            _ => None,
        }
    }

    /// The host event name each duty answers (identical in Claude Code and Codex).
    fn host_event(self) -> &'static str {
        match self {
            Event::SessionCheck => "SessionStart",
            Event::ToolPreflight => "PreToolUse",
            Event::ToolObserve => "PostToolUse",
            Event::TurnValidation => "Stop",
        }
    }
}

/// What the policy core decided, before any host rendering.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum Verdict {
    Allow,
    Deny(String),
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
        stderr: format!("moai-cli hook-dispatch: engine error: {reason}"),
        exit: ENGINE_ERROR_EXIT,
    }
}

/// Runs one hook: parses the host and duty, decides, renders the host's native answer.
/// `cwd` is the hook process's working directory (the host runs hooks in the session's
/// directory); it is also the project whose local claim journal is read.
pub fn dispatch(host: &str, event: &str, stdin: &[u8], cwd: &Path) -> Answer {
    if host == "codex" {
        return engine_error("the codex adapter is not implemented yet (#2391 stage 2)");
    }
    if host != "claude" {
        return engine_error("unknown hook host; must be claude or codex");
    }
    let Some(event) = Event::parse(event) else {
        return engine_error(
            "unknown hook event; must be session-check, tool-preflight, tool-observe or turn-validation",
        );
    };
    let Ok(Value::Object(input)) = serde_json::from_slice::<Value>(stdin) else {
        return engine_error("hook input is not a JSON object");
    };
    let input = Value::Object(input);
    // A present event name must match the duty; unknown extra fields are ignored.
    match input.get("hook_event_name") {
        None => {}
        Some(Value::String(name)) if name == event.host_event() => {}
        Some(_) => return engine_error("hook_event_name does not match the requested duty"),
    }
    let verdict = match event {
        Event::SessionCheck => session_check(cwd),
        Event::ToolPreflight => tool_preflight(&input, cwd),
        Event::ToolObserve | Event::TurnValidation => {
            return engine_error("this duty is not implemented yet (#2391 later stage)")
        }
    };
    render_claude(event, verdict)
}

/// Claude Code rendering of a verdict (documented shapes only).
fn render_claude(event: Event, verdict: Verdict) -> Answer {
    let stdout = match (event, verdict) {
        (_, Verdict::Allow) => String::new(),
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
        // A verdict the duty cannot carry is a bug in the policy core, reported, not guessed.
        _ => return engine_error("policy returned a verdict this duty cannot carry"),
    };
    Answer {
        stdout,
        stderr: String::new(),
        exit: 0,
    }
}

/// session-check: a short, fixed-vocabulary state summary for the model. Never blocks.
fn session_check(cwd: &Path) -> Verdict {
    let state_dir = match std::fs::symlink_metadata(cwd.join(crate::PROJECT_STATE_DIR)) {
        Ok(m) if m.is_dir() => "present",
        Ok(_) => "not a directory",
        Err(_) => "absent",
    };
    let lane = match crate::claim_verify::current_branch(cwd)
        .map(|b| crate::claim_verify::lane_branch(&b))
    {
        Some(crate::claim_verify::LaneBranch::Issue(n)) => format!("Issue #{n}"),
        Some(crate::claim_verify::LaneBranch::Unparseable) => "unparseable issue/ branch".into(),
        Some(crate::claim_verify::LaneBranch::NotLane) => "none".into(),
        None => "none (detached HEAD or not a git worktree)".into(),
    };
    let record = if crate::claim_verify::has_local_record(cwd, cwd) {
        "present"
    } else {
        "absent"
    };
    Verdict::Context(format!(
        "MoAI session-check (moai-cli {}): project state directory {}: {state_dir}; lane branch: \
         {lane}; local claim record: {record}. Read locally; no network was used.",
        env!("CARGO_PKG_VERSION"),
        crate::PROJECT_STATE_DIR,
    ))
}

/// tool-preflight, stage 1 policy: a Bash `git commit` on a lane branch `issue/<n>-…` that has
/// no local claim record must carry `Moai-Claim: <n>#<id>` in its command line (the decided
/// lane rule; AGENTS.local.md and the MODE-AGENT workflow step 6). Everything else is allowed
/// here; a forge check of the claim stays with `claim-verify --hook`.
fn tool_preflight(input: &Value, cwd: &Path) -> Verdict {
    let command = input["tool_input"]["command"].as_str().unwrap_or("");
    if input["tool_name"] != json!("Bash") || !crate::claim_verify::is_commit(command) {
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
    let Some(branch) = crate::claim_verify::current_branch(&dir) else {
        return Verdict::Allow;
    };
    let crate::claim_verify::LaneBranch::Issue(issue) = crate::claim_verify::lane_branch(&branch)
    else {
        return Verdict::Allow;
    };
    if crate::claim_verify::has_local_record(cwd, &dir) {
        return Verdict::Allow;
    }
    match crate::claim_verify::commit_claim_id(command, issue) {
        Ok(_) => Verdict::Allow,
        Err(gap) => Verdict::Deny(format!(
            "MoAI tool-preflight: {}; put `Moai-Claim: {issue}#<claim comment id>` in the -m \
             message.",
            gap.reason(issue)
        )),
    }
}
