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

/// The opt-in deny rules of stage 7 (B3, B9; decision 6779829826), off unless
/// `~/.moai-cli/config` turns them on (see `hook_config`).
pub const OPT_IN_RULES: &str = "[hooks.outside_project_writes] enabled + allow: file writes outside the project (after symlinks and ..) are refused unless under allow; [hooks.primary_checkout] enabled + repos + allow_commands: switch, checkout <branch>, reset, stash (not list/show) and rebase are refused in a listed primary checkout that has linked worktrees. A missing or broken config turns both off";

/// The confirmation class of tool-preflight (B2, Issue #2391 classification row B2).
pub const CONFIRM_COMMANDS: &str =
    "git reset --hard, a forced git clean (not a dry run), a force push to a branch other than main or master";

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
        stderr: format!("moai-cli hook-dispatch: engine error: {reason}"),
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
        return engine_error("unknown hook host; must be claude or codex");
    };
    let Some(event) = Event::parse(event) else {
        return engine_error(
            "unknown hook event; must be session-check, tool-preflight, tool-observe or turn-validation",
        );
    };
    let Ok(Value::Object(input)) = serde_json::from_slice::<Value>(stdin) else {
        // A PostToolUse input carries the whole tool result; one past the stdin bound is cut
        // off and no longer parses. Observation is best-effort and the tool already ran, so
        // that is a quiet skip, not an error the host would surface on every large result.
        if event == Event::ToolObserve {
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
    let verdict = match event {
        Event::SessionCheck => session_check(cwd),
        Event::ToolPreflight => tool_preflight(&input, cwd),
        Event::ToolObserve => tool_observe(&input, cwd, now_unix_ms, host),
        Event::TurnValidation => turn_validation_with(STOP_RULES, &input, cwd, now_unix_ms),
    };
    if let (Event::ToolPreflight, Verdict::Deny(_)) = (event, &verdict) {
        record_deny(&input, cwd, now_unix_ms, host);
    }
    match host {
        Host::Claude => render_claude(event, verdict),
        Host::Codex => render_codex(event, verdict),
    }
}

/// The hosts with a native adapter.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Host {
    Claude,
    Codex,
}

impl Host {
    pub fn parse(word: &str) -> Option<Host> {
        match word {
            "claude" => Some(Host::Claude),
            "codex" => Some(Host::Codex),
            _ => None,
        }
    }

    pub fn name(self) -> &'static str {
        match self {
            Host::Claude => "claude",
            Host::Codex => "codex",
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
fn record_deny(input: &Value, cwd: &Path, now_unix_ms: u128, host: Host) {
    let line = json!({
        "schema_version": 1,
        "kind": "deny",
        "host": host.name(),
        "session": session_key(input),
        "call": call_key(input),
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
                                   "Write/Edit/MultiEdit/NotebookEdit: .git internals, credential directories, private-key file names, private-key or GitHub-token content"],
                               "opt_in": OPT_IN_RULES,
                               "not_guarded": ["writes outside the project unless [hooks.outside_project_writes] is enabled", "shell redirections into protected paths", "MCP tools"],
                               "answer": "hookSpecificOutput.permissionDecision=deny",
                               "confirm": {"answer": "hookSpecificOutput.permissionDecision=ask", "commands": CONFIRM_COMMANDS,
                                           "note": "a deny anywhere on the line wins; ask in a bypass-permissions or -p session is not documented"},
                               "never_uses": ["allow", "defer", "decision/reason (deprecated)"]},
            "tool-observe": {"events": ["PostToolUse", "PostToolUseFailure"], "records": "observe line", "blocks": false},
            "turn-validation": {"event": "Stop", "rules": STOP_RULES.len(), "answer": "decision=block + reason",
                                "host_cap": "8 consecutive continuations (CLAUDE_CODE_STOP_HOOK_BLOCK_CAP)", "moai_cap": MOAI_STOP_BLOCK_CAP},
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
                                   "apply_patch: .git internals, credential directories, private-key file names (paths from '*** Add/Update/Delete File:' and '*** Move to:' lines, a format the docs do not define; an unreadable patch is not path-checked), private-key or GitHub-token content in added lines"],
                               "opt_in": OPT_IN_RULES,
                               "not_guarded": ["writes outside the project unless [hooks.outside_project_writes] is enabled", "shell redirections into protected paths", "MCP tools"],
                               "answer": "hookSpecificOutput.permissionDecision=deny",
                               "confirm": {"answer": "hookSpecificOutput.permissionDecision=deny (Codex has no ask)", "commands": CONFIRM_COMMANDS},
                               "never_uses": ["ask", "approve", "continue=false", "stopReason", "allow+updatedInput"]},
            "tool-observe": {"events": ["PostToolUse"], "records": "observe line", "blocks": false,
                             "note": "Codex has no PostToolUseFailure; apply_patch has no file_path, so its path is not recorded",
                             "outcome": "PostToolUse also runs after a failed Bash command, so the event never means success: outcome is unknown unless a Bash tool_response carries an integer exit_code (a field the docs do not name)"},
            "turn-validation": {"event": "Stop", "rules": STOP_RULES.len(), "answer": "decision=block + reason",
                                "host_cap": "not documented", "moai_cap": MOAI_STOP_BLOCK_CAP,
                                "unverified": "empty stdout on exit 0 for Stop (stated generally, not for Stop)"},
            "on_engine_error": "exit 1: not documented for command hooks (unverified)",
        }),
        None => json!({"error": "unknown hook host; must be claude or codex"}),
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
/// tool-preflight with the confirmation class (B2): any deny — a guard or the claim token, in
/// any segment of the line — wins; only a call that would be allowed is turned into `Ask`
/// (Claude's own order is deny > defer > ask > allow).
fn tool_preflight(input: &Value, cwd: &Path) -> Verdict {
    let verdict = tool_preflight_deny(input, cwd);
    if verdict != Verdict::Allow || input["tool_name"] != json!("Bash") {
        return verdict;
    }
    let command = input["tool_input"]["command"].as_str().unwrap_or("");
    match crate::hook_guards::bash_confirm(command) {
        Some(reason) => Verdict::Ask(reason.to_string()),
        None => Verdict::Allow,
    }
}

fn tool_preflight_deny(input: &Value, cwd: &Path) -> Verdict {
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
    } else {
        crate::hook_guards::write_guard(tool, &input["tool_input"], &write_dir)
    };
    if let Some(reason) = guard {
        return Verdict::Deny(reason.to_string());
    }
    // Stage 7 opt-in rules (B3, B9; decision 6779829826): off unless ~/.moai-cli/config turns
    // them on; a missing or broken config leaves them off.
    let config = crate::hook_config::load();
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
        return Verdict::Deny(reason.to_string());
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
    Verdict::Deny(format!(
        "MoAI tool-preflight: {}; put `Moai-Claim: {issue}#<claim comment id>` in the -m \
         message.",
        gap.reason(issue)
    ))
}

/// The log both stage-2 duties share, under `<cwd>/.moai-cli/receipts/`: tool-observe appends
/// `observe` lines, turn-validation appends `stop_block` lines and counts them per session.
pub const ENGINE_LOG: &str = "hook-engine.jsonl";

/// A short, stable key for a host session id: the first 16 hex digits of its sha256. The raw
/// id is host input and never becomes a path or a log value.
fn session_key(input: &Value) -> String {
    match input["session_id"].as_str() {
        Some(id) if !id.is_empty() => {
            let d = crate::digest(id.as_bytes());
            d.trim_start_matches("sha256:")[..16].to_string()
        }
        _ => "unknown".into(),
    }
}

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
    let path = Path::new(raw);
    let rel = if path.is_absolute() {
        match path.strip_prefix(cwd) {
            Ok(rel) => rel.to_path_buf(),
            Err(_) => return json!("outside"),
        }
    } else {
        path.to_path_buf()
    };
    let escapes = rel.components().any(|c| {
        !matches!(
            c,
            std::path::Component::Normal(_) | std::path::Component::CurDir
        )
    });
    if escapes {
        return json!("outside");
    }
    match rel.to_str() {
        Some(s) if loggable(s, 256) => json!(s),
        _ => json!("unrecorded"),
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
    }
}

fn tool_observe(input: &Value, cwd: &Path, now_unix_ms: u128, host: Host) -> Verdict {
    // Cost (audit #2400 N3): a project without a real `.moai-cli/` directory records nothing,
    // so return before spawning `git` for the HEAD commit.
    match std::fs::symlink_metadata(cwd.join(crate::PROJECT_STATE_DIR)) {
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
    Verdict::Allow
}

/// What a Stop rule sees.
pub struct StopContext<'a> {
    /// The session directory.
    pub cwd: &'a Path,
    /// The host's `stop_hook_active`: this turn is already continuing because of a stop hook.
    pub stop_hook_active: bool,
}

/// One turn-validation rule: `check` returns the reason to keep working, or `None`. A rule
/// may only ask for something the model can resolve itself; a condition waiting on an audit,
/// CI, the leader or the user must never be a rule (it would block every turn end).
pub struct StopRule {
    pub id: &'static str,
    pub check: fn(&StopContext) -> Option<String>,
}

/// The decided Stop rules. Empty: no rule is decided yet (leader decision, Issue #2391), so
/// turn-validation never blocks. A rule is added here only with its decision record.
pub const STOP_RULES: &[StopRule] = &[];

/// How many times in a row MoAI itself may keep one session working before it lets the turn
/// end. Claude Code's own cap (8 continuations, shared by every stop hook) stays the outer
/// bound; this one keeps MoAI's share small next to a user's `/goal` hook.
pub const MOAI_STOP_BLOCK_CAP: usize = 2;

/// MoAI's consecutive `stop_block` lines for `session` since that session's last `observe`
/// line (a tool call resets the count, as Claude Code's own cap does). Reads the bounded log
/// only; a missing or unreadable log counts as zero.
fn consecutive_blocks(cwd: &Path, session: &str) -> usize {
    let log = cwd
        .join(crate::PROJECT_STATE_DIR)
        .join("receipts")
        .join(ENGINE_LOG);
    let Ok(meta) = std::fs::symlink_metadata(&log) else {
        return 0;
    };
    if !meta.is_file() || meta.len() > crate::hooks::RECEIPT_LOG_LIMIT {
        return 0;
    }
    let Ok(text) = std::fs::read_to_string(&log) else {
        return 0;
    };
    let mut count = 0;
    for line in text.lines().rev() {
        let Ok(v) = serde_json::from_str::<Value>(line) else {
            continue;
        };
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
///   does not report a continuation);
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
    let blocks = consecutive_blocks(cwd, &session);
    if active || blocks >= MOAI_STOP_BLOCK_CAP {
        return Verdict::Allow;
    }
    let ctx = StopContext {
        cwd,
        stop_hook_active: active,
    };
    for rule in rules {
        if let Some(reason) = (rule.check)(&ctx) {
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
    fn shipped_rule_set_is_empty_and_never_blocks_or_writes() {
        assert_eq!(STOP_RULES.len(), 0);
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
}
