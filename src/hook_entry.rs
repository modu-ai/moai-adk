//! `moai hook <name>` (SPEC-20261123-hook-entry-skeleton): the entry point Claude Code's hook
//! settings call. Three names are registered, one per planned duty: `session-context` (H1),
//! `bash-boundary` (H2) and `write-boundary` (H3). `session-context` adds context to the start
//! of a session (SPEC-20261123-hook-session-start); `bash-boundary` judges in
//! `hook_bash_boundary` (SPEC-20261123-hook-bash-boundary); `write-boundary` judges in
//! `hook_write_boundary` (SPEC-20261126-hook-write-boundary).
//!
//! What the entry does, in order: it reads standard input once (bounded); it looks for
//! `.moai/project.yaml` in `CLAUDE_PROJECT_DIR`, else in the working folder, and without it
//! stays silent; with it, bad input is recorded as one receipt line, and the two `PreToolUse`
//! names also deny an input they cannot read: over the limit, not UTF-8, not JSON or not a JSON
//! object (SPEC-20261126-hook-unreadable-input). Every path ends in exit 0; a call is refused
//! only by a deny object on stdout, from a handler or from the entry for such an input. The entry
//! makes no network call, starts no other program and reads no variable beyond the two named in
//! [`env`]. It does not look at which terminal or orchestrator started the host.

use serde_json::{json, Map, Value};
use std::fs;
use std::io::{Read, Write};
use std::os::unix::fs::OpenOptionsExt;
use std::path::{Path, PathBuf};

/// The SessionStart handler (H1). Declared here, by path, so the module needs no line in the
/// crate root (SPEC-20261123-hook-session-start, D-6).
#[path = "hook_session_start.rs"]
mod session_start;

/// The registered hook names, in the order of the plan (H1, H2, H3).
pub const NAMES: [&str; 3] = ["session-context", "bash-boundary", "write-boundary"];
/// Standard input is read up to this many bytes; one more byte means "too large".
pub const STDIN_LIMIT: usize = 4 * 1024 * 1024;
/// The receipt reason of an input over [`STDIN_LIMIT`].
const TOO_LARGE: &str = "stdin_too_large";
/// The receipt log, inside `<HOME>/.moai/projects/<project_id>/`.
const RECEIPT_FILE: &str = "hook-entry.jsonl";
/// A receipt log of this many bytes takes no more lines.
const RECEIPT_LIMIT: u64 = 262_144;
const PROJECT_FILE: &str = ".moai/project.yaml";
const PROJECT_FILE_LIMIT: u64 = 64 * 1024;

pub fn is_registered(name: &str) -> bool {
    NAMES.contains(&name)
}

/// What the entry prints; it always exits 0.
#[derive(Debug, Default, PartialEq)]
pub struct Answer {
    pub stdout: String,
    pub stderr: String,
}

/// The two variables the entry reads.
pub struct Env {
    pub project_dir: Option<PathBuf>,
    pub home: Option<PathBuf>,
}

pub fn env() -> Env {
    Env {
        project_dir: std::env::var_os("CLAUDE_PROJECT_DIR")
            .filter(|v| !v.is_empty())
            .map(PathBuf::from),
        home: std::env::var_os("HOME")
            .filter(|v| !v.is_empty())
            .map(PathBuf::from),
    }
}

/// Reads standard input once: at most [`STDIN_LIMIT`] + 1 bytes are kept, and an input of more
/// is read to its end and dropped, so the writer is never cut off by a closed pipe. A read
/// error keeps what came.
pub fn read_input(stdin: &mut impl Read) -> Vec<u8> {
    let mut bytes = Vec::new();
    let _ = stdin
        .by_ref()
        .take(STDIN_LIMIT as u64 + 1)
        .read_to_end(&mut bytes);
    if bytes.len() > STDIN_LIMIT {
        let _ = std::io::copy(stdin, &mut std::io::sink());
    }
    bytes
}

/// Runs the entry for a registered `name` and writes its answer: the text of stdout and the
/// text of stderr, both ignoring a write error (a closed pipe never changes the exit code, and
/// the caller prints nothing more). `cwd` is the working folder, `now_ms` the clock in
/// milliseconds since 1970.
pub fn run(name: &str, input: &[u8], cwd: &Path, env: &Env, now_ms: u128) -> Answer {
    let answer = decide(name, input, cwd, env, now_ms);
    if !answer.stdout.is_empty() {
        let mut out = std::io::stdout();
        let _ = writeln!(out, "{}", answer.stdout);
        let _ = out.flush();
    }
    if !answer.stderr.is_empty() {
        let _ = std::io::stderr().write_all(answer.stderr.as_bytes());
    }
    Answer::default()
}

fn decide(name: &str, input: &[u8], cwd: &Path, env: &Env, now_ms: u128) -> Answer {
    let root = env.project_dir.as_deref().unwrap_or(cwd);
    let Some(project) = moai_project(root) else {
        return Answer::default();
    };
    let object = match parse(input) {
        // An input the entry cannot read reaches no handler: it is recorded, and a `PreToolUse`
        // name denies the call, whether or not the receipt line could be written
        // (SPEC-20261126-hook-unreadable-input REQ-1, REQ-2, REQ-4, REQ-5).
        Err(reason) => {
            return Answer {
                stdout: refusal(name, reason)
                    .map(|v| v.to_string())
                    .unwrap_or_default(),
                ..failure(name, reason, &project, env, now_ms)
            }
        }
        Ok(object) => object,
    };
    // A call that names another event than the registered name belongs to is judged by nobody.
    if let (Some(expected), Some(seen)) = (expected_event(name), object.get("hook_event_name")) {
        if seen.as_str() != Some(expected) {
            return failure(name, "event_mismatch", &project, env, now_ms);
        }
    }
    if name == "session-context" {
        let context = session_start::context(
            &object,
            root,
            cwd,
            env.home.as_deref(),
            project.id.as_deref(),
        );
        let stderr = context
            .ledger_failure
            .map(|reason| failure(name, reason, &project, env, now_ms).stderr)
            .unwrap_or_default();
        return Answer {
            stdout: context.output.to_string(),
            stderr,
        };
    }
    Answer {
        stdout: handle(name, &object, root, cwd, env)
            .map(|v| v.to_string())
            .unwrap_or_default(),
        stderr: String::new(),
    }
}

struct Project {
    id: Option<String>,
}

/// A MoAI project has a regular file `.moai/project.yaml` (a link to a regular file counts).
fn moai_project(root: &Path) -> Option<Project> {
    let file = root.join(PROJECT_FILE);
    if !fs::metadata(&file).ok()?.is_file() {
        return None;
    }
    let id = fs::File::open(&file).ok().and_then(|f| {
        let mut bytes = Vec::new();
        f.take(PROJECT_FILE_LIMIT).read_to_end(&mut bytes).ok()?;
        let value: Value = serde_json::from_slice(&bytes).ok()?;
        value.get("project_id")?.as_str().map(str::to_owned)
    });
    Some(Project { id })
}

fn parse(input: &[u8]) -> Result<Map<String, Value>, &'static str> {
    if input.len() > STDIN_LIMIT {
        return Err(TOO_LARGE);
    }
    let text = std::str::from_utf8(input).map_err(|_| "not_utf8")?;
    match serde_json::from_str::<Value>(text) {
        Err(_) => Err("not_json"),
        Ok(Value::Object(map)) => Ok(map),
        Ok(_) => Err("not_object"),
    }
}

/// The deny object of an input the entry could not read, `None` for `session-context`, which
/// decides about no tool call. `write-boundary` keeps its own text for an input over the limit
/// (SPEC-20261126-hook-write-boundary REQ-13).
// @cc:careful: this fails closed. A host that changes the form of its hook input gets every call
// of the two `PreToolUse` names denied in a MoAI project, and the model cannot repair that with
// those tools, so each text has to keep asking it to tell the user
// (SPEC-20261126-hook-unreadable-input R-1).
fn refusal(name: &str, reason: &str) -> Option<Value> {
    match name {
        "bash-boundary" => Some(bash_refusal(reason)),
        "write-boundary" if reason == TOO_LARGE => {
            Some(moai_cli::hook_write_boundary::over_limit())
        }
        "write-boundary" => Some(moai_cli::hook_write_boundary::unreadable(reason)),
        _ => None,
    }
}

/// The Bash boundary's denial of an input the entry could not read, in its class `unreadable`.
/// The text is fixed per cause and holds no byte of the input
/// (SPEC-20261126-hook-unreadable-input REQ-3).
fn bash_refusal(reason: &str) -> Value {
    const AGAIN: &str = "send the call again with plain text";
    let (cause, way) = match reason {
        TOO_LARGE => (
            "this call is over the 4 MiB the check reads",
            "run a shorter command",
        ),
        "not_utf8" => ("the hook input is not valid UTF-8", AGAIN),
        "not_json" => ("the hook input is not valid JSON", AGAIN),
        _ => ("the hook input is not a JSON object", AGAIN),
    };
    let text = format!(
        "moai-boundary/unreadable: {cause}, so the command cannot be checked; {way}, and if \
         every call is refused, tell the user, who can remove or disable the moai hook entry"
    );
    json!({
        "hookSpecificOutput": {
            "hookEventName": "PreToolUse",
            "permissionDecision": "deny",
            "permissionDecisionReason": text,
        }
    })
}

/// The handlers of the other hooks: H1 (`session-context`) is `session_start::context`, H2
/// (`bash-boundary`) judges in `hook_bash_boundary`, and H3 (`write-boundary`) judges in `hook_write_boundary`.
fn handle(
    name: &str,
    input: &Map<String, Value>,
    root: &Path,
    cwd: &Path,
    env: &Env,
) -> Option<Value> {
    match name {
        "bash-boundary" => moai_cli::hook_bash_boundary::decide(
            input,
            &moai_cli::hook_bash_boundary::Env {
                project_root: root,
                project_dir: env.project_dir.as_deref(),
                home: env.home.as_deref(),
                process_cwd: cwd,
            },
        ),
        "write-boundary" => moai_cli::hook_write_boundary::decide(
            input,
            &moai_cli::hook_write_boundary::Env {
                project_dir: env.project_dir.as_deref(),
                process_cwd: cwd,
            },
        ),
        _ => None,
    }
}

/// The event a registered name belongs to, `None` for a name that is not one of the three.
fn expected_event(name: &str) -> Option<&'static str> {
    match name {
        "session-context" => Some("SessionStart"),
        "bash-boundary" | "write-boundary" => Some("PreToolUse"),
        _ => None,
    }
}

/// One receipt line, or one stderr line saying why there is none. The line holds the name, the
/// reason word and the time; no input byte goes into it.
fn failure(name: &str, reason: &str, project: &Project, env: &Env, now_ms: u128) -> Answer {
    let line = json!({
        "schema": 1,
        "name": name,
        "reason": reason,
        "at_unix_ms": u64::try_from(now_ms).unwrap_or(u64::MAX),
    })
    .to_string();
    match append_receipt(project, env, &line) {
        Ok(()) => Answer::default(),
        Err(why) => Answer {
            stdout: String::new(),
            stderr: format!("moai hook {name}: {reason}; no receipt: {why}\n"),
        },
    }
}

fn append_receipt(project: &Project, env: &Env, line: &str) -> Result<(), String> {
    let home = env.home.as_deref().ok_or("HOME is not set")?;
    let id = project
        .id
        .as_deref()
        .ok_or("project.yaml has no project_id")?;
    let dir = moai_cli::ledger::location::project_dir(home, id)
        .map_err(|_| "project_id is not a portable id")?;
    // `.moai`, `projects` and the project folder are never followed when they are links.
    moai_cli::ledger::folders::refuse_links(home, id)
        .map_err(|_| "a folder of the project state under HOME is a link")?;
    match fs::symlink_metadata(&dir) {
        Ok(m) if m.is_dir() => {}
        _ => return Err("the project is not registered under HOME".to_owned()),
    }
    // The receipt file is opened without following a link (and without waiting on a FIFO), then
    // judged by what was opened, so a swap after a check cannot lead the write elsewhere.
    let mut out = fs::OpenOptions::new()
        .create(true)
        .append(true)
        .mode(0o600)
        .custom_flags(libc::O_NOFOLLOW | libc::O_NONBLOCK)
        .open(dir.join(RECEIPT_FILE))
        .map_err(|e| match e.raw_os_error() {
            Some(libc::ELOOP) => "the receipt path is a link".to_owned(),
            _ => "the receipt log cannot be opened".to_owned(),
        })?;
    let meta = out
        .metadata()
        .map_err(|_| "the receipt path cannot be read".to_owned())?;
    if !meta.is_file() {
        return Err("the receipt path is not a regular file".to_owned());
    }
    if meta.len() >= RECEIPT_LIMIT {
        return Err(format!(
            "the receipt log is full ({RECEIPT_LIMIT} bytes), nothing appended"
        ));
    }
    out.write_all(format!("{line}\n").as_bytes())
        .map_err(|_| "the receipt line cannot be written".to_owned())
}
