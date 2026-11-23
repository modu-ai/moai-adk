//! `moai hook <name>` (SPEC-20261123-hook-entry-skeleton): the entry point Claude Code's hook
//! settings call. Three names are registered, one per planned duty: `session-context` (H1),
//! `bash-boundary` (H2) and `write-boundary` (H3). In this slice every handler answers no
//! decision, so a registered name prints nothing and exits 0; the judgments come in later slices.
//!
//! What the entry does, in order: it reads standard input once (bounded); it looks for
//! `.moai/project.yaml` in `CLAUDE_PROJECT_DIR`, else in the working folder, and without it
//! stays silent; with it, bad input is recorded as one receipt line. It never blocks: every path
//! ends in exit 0. The entry makes no network call, starts no other program and reads no
//! variable beyond the two named in [`env`]. It does not look at which terminal or
//! orchestrator started the host.

use serde_json::{json, Map, Value};
use std::fs;
use std::io::{Read, Write};
use std::path::{Path, PathBuf};

/// The registered hook names, in the order of the plan (H1, H2, H3).
pub const NAMES: [&str; 3] = ["session-context", "bash-boundary", "write-boundary"];
/// Standard input is read up to this many bytes; one more byte means "too large".
pub const STDIN_LIMIT: usize = 4 * 1024 * 1024;
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

/// Reads standard input once, at most [`STDIN_LIMIT`] + 1 bytes; a read error keeps what came.
pub fn read_input(stdin: &mut impl Read) -> Vec<u8> {
    let mut bytes = Vec::new();
    let _ = stdin.take(STDIN_LIMIT as u64 + 1).read_to_end(&mut bytes);
    bytes
}

/// Runs the entry for a registered `name`. `cwd` is the working folder, `now_ms` the clock in
/// milliseconds since 1970.
pub fn run(name: &str, input: &[u8], cwd: &Path, env: &Env, now_ms: u128) -> Answer {
    let root = env.project_dir.as_deref().unwrap_or(cwd);
    let Some(project) = moai_project(root) else {
        return Answer::default();
    };
    match parse(input) {
        Err(reason) => failure(name, reason, &project, env, now_ms),
        Ok(object) => Answer {
            stdout: handle(name, &object)
                .map(|v| v.to_string())
                .unwrap_or_default(),
            stderr: String::new(),
        },
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
        return Err("stdin_too_large");
    }
    let text = std::str::from_utf8(input).map_err(|_| "not_utf8")?;
    match serde_json::from_str::<Value>(text) {
        Err(_) => Err("not_json"),
        Ok(Value::Object(map)) => Ok(map),
        Ok(_) => Err("not_object"),
    }
}

/// The handlers of this slice: none decides anything.
fn handle(name: &str, input: &Map<String, Value>) -> Option<Value> {
    match name {
        "session-context" => session_context(input),
        "bash-boundary" => bash_boundary(input),
        "write-boundary" => write_boundary(input),
        _ => None,
    }
}

/// H1 (a later slice fills it).
fn session_context(_input: &Map<String, Value>) -> Option<Value> {
    None
}

/// H2 (a later slice fills it).
fn bash_boundary(_input: &Map<String, Value>) -> Option<Value> {
    None
}

/// H3 (a later slice fills it).
fn write_boundary(_input: &Map<String, Value>) -> Option<Value> {
    None
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
    match fs::symlink_metadata(&dir) {
        Ok(m) if m.is_dir() => {}
        _ => return Err("the project is not registered under HOME".to_owned()),
    }
    let file = dir.join(RECEIPT_FILE);
    match fs::symlink_metadata(&file) {
        Ok(m) if !m.is_file() => return Err("the receipt path is not a regular file".to_owned()),
        Ok(m) if m.len() >= RECEIPT_LIMIT => {
            return Err(format!(
                "the receipt log is full ({RECEIPT_LIMIT} bytes), nothing appended"
            ))
        }
        Ok(_) => {}
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => {}
        Err(_) => return Err("the receipt path cannot be read".to_owned()),
    }
    let mut out = fs::OpenOptions::new()
        .create(true)
        .append(true)
        .open(&file)
        .map_err(|_| "the receipt log cannot be opened".to_owned())?;
    out.write_all(format!("{line}\n").as_bytes())
        .map_err(|_| "the receipt line cannot be written".to_owned())
}
