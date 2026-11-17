//! Edit-time diagnostics, slice S1 (SPEC-20261117-edit-diagnostics; leader decisions Q1-Q6 of
//! 2026-11-17): after Claude Code finishes a `Write`, `Edit` or `MultiEdit`, run the one checker
//! the user named in `[hooks.edit_diagnostics]` for the file's extension and return its
//! diagnostics as `additionalContext` next to the tool result. Off unless the user enabled it;
//! never blocks (blocking is slice S2); nothing is installed and MoAI opens no connection.
//!
//! - The file must lie inside the project (the git top level of the hook's working folder, else
//!   that folder) with no symbolic link on any component below the project root, and be a
//!   regular file. Otherwise no checker runs and nothing is printed.
//! - The checker runs without a shell, from the project root, with the project-relative path as
//!   its last argument and its own process group; past `timeout_secs` the whole group is
//!   killed. Each output stream is read up to [`READ_MAX`] bytes and drained past it; bytes that
//!   are not UTF-8 are replaced.
//! - Lines in the `path:line[:col]: [severity:] message` form become diagnostics: a
//!   project-relative path (absolute paths outside the project are not shown), line and column,
//!   a severity word, and the message cut to [`MESSAGE_CHARS`] characters. A line with a
//!   credential shape (`redact::contains_secret_shape`) is masked.
//! - One checker per project at a time (a lock file in the HOME state folder's
//!   `edit-diagnostics/`, mode 0700): a second edit while one runs gets a one-line "skipped".
//! - The checker gets `./<rel>`, never a path that starts with `-` (CWE-88), and its whole
//!   process group is stopped when it ends, by itself or by the timeout.
//! - ANSI escape sequences and control characters other than tab are removed from its output.
//! - A missing checker, a timeout, a cut or non-UTF-8 output each say so in one summary line.
//!   A clean run (exit 0, no diagnostic, nothing cut) prints nothing.

use crate::hook_config::EditRule;
use serde_json::Value;
use std::io::Read;
use std::path::{Component, Path, PathBuf};
use std::process::{Command, Stdio};
use std::sync::mpsc;
use std::time::{Duration, Instant};

/// Bytes read from each of the checker's output streams; the rest is drained unread.
pub const READ_MAX: usize = 64 * 1024;
/// Characters kept of one diagnostic message.
pub const MESSAGE_CHARS: usize = 200;
/// The tools whose edits are checked (Claude Code tool names).
pub const TOOLS: [&str; 3] = ["Write", "Edit", "MultiEdit"];
/// The hook matcher that runs this duty.
pub const MATCHER: &str = "Write|Edit|MultiEdit";
/// What a masked diagnostic shows instead of its text.
const MASKED: &str = "[masked: credential-shaped text]";

/// The context text for one finished edit, or `None` when there is nothing to say.
pub fn run(input: &Value, cwd: &Path, rule: &EditRule) -> Option<String> {
    let tool = input["tool_name"].as_str()?;
    if !TOOLS.contains(&tool) {
        return None;
    }
    let file = PathBuf::from(input["tool_input"]["file_path"].as_str()?);
    let root = project_root(cwd);
    let rel = inside(&root, &file)?;
    let ext = rel.extension()?.to_str()?.to_ascii_lowercase();
    let argv = &rule.checkers.iter().find(|(e, _)| *e == ext)?.1;
    let name = checker_name(&argv[0]);
    let shown = rel.to_string_lossy().into_owned();
    let _lock = match lock(&root) {
        Lock::Busy => {
            return Some(format!(
                "MoAI edit-diagnostics ({name}, {shown}): another check is running in this \
                 project; this edit was not checked"
            ))
        }
        Lock::Held(f) => Some(f),
        Lock::None => None,
    };
    let ran = run_checker(argv, &root, &rel, Duration::from_secs(rule.timeout_secs));
    Some(render(&ran, &name, &shown, &root, rule)).filter(|s| !s.is_empty())
}

/// The git top level of `cwd` (one scrubbed `git rev-parse`), else `cwd`; a real path.
fn project_root(cwd: &Path) -> PathBuf {
    let mut cmd = Command::new("git");
    cmd.args(["rev-parse", "--show-toplevel"])
        .current_dir(cwd)
        .stdin(Stdio::null())
        .stderr(Stdio::null());
    crate::identity::scrub_git_env(&mut cmd);
    let top = cmd
        .output()
        .ok()
        .filter(|o| o.status.success())
        .map(|o| PathBuf::from(String::from_utf8_lossy(&o.stdout).trim()));
    let root = top.unwrap_or_else(|| cwd.to_path_buf());
    std::fs::canonicalize(&root).unwrap_or(root)
}

/// The project-relative path of `file`, when it is an absolute path to a regular file inside
/// `root` (a real path) with no symbolic link on any component below the root. The root is
/// matched by its outermost ancestor of `file` that resolves to it, so a link above the project
/// (`/var` -> `/private/var`) is fine and a link inside it is not.
pub fn inside(root: &Path, file: &Path) -> Option<PathBuf> {
    if !file.is_absolute()
        || file
            .components()
            .any(|c| matches!(c, Component::ParentDir | Component::CurDir))
    {
        return None;
    }
    let ancestors: Vec<&Path> = file.ancestors().skip(1).collect();
    let base = ancestors
        .iter()
        .rev()
        .find(|a| std::fs::canonicalize(a).ok().as_deref() == Some(root))?;
    let rel = file.strip_prefix(base).ok()?.to_path_buf();
    let mut at = base.to_path_buf();
    let parts: Vec<_> = rel.components().collect();
    for (i, part) in parts.iter().enumerate() {
        let Component::Normal(_) = part else {
            return None;
        };
        at.push(part);
        let meta = std::fs::symlink_metadata(&at).ok()?;
        let last = i + 1 == parts.len();
        if meta.file_type().is_symlink() || (last && !meta.is_file()) || (!last && !meta.is_dir()) {
            return None;
        }
    }
    (!rel.as_os_str().is_empty()).then_some(rel)
}

/// The checker's file name only, so no folder of the user's reaches the model.
fn checker_name(program: &str) -> String {
    let name = Path::new(program)
        .file_name()
        .map(|n| n.to_string_lossy().into_owned())
        .unwrap_or_default();
    if name.is_empty() || crate::redact::contains_secret_shape(&name) {
        "checker".into()
    } else {
        name
    }
}

enum Lock {
    Held(std::fs::File),
    Busy,
    /// No lock could be made (not unix, or the lock file could not be opened): run anyway.
    None,
}

/// The folder that holds the locks: `edit-diagnostics/` in the user's HOME state folder (S1
/// audit N3: not the shared temporary folder), made with mode 0700 when the state folder is
/// there. `None` when HOME is unset, the state folder is missing or not a real folder, or a link
/// or a non-folder stands at either place: then the check runs without a lock.
fn lock_dir() -> Option<PathBuf> {
    let home = PathBuf::from(std::env::var_os("HOME").filter(|h| !h.is_empty())?);
    let state = crate::home_root::path(&home);
    let real_dir = |p: &Path| {
        std::fs::symlink_metadata(p).is_ok_and(|m| m.is_dir() && !m.file_type().is_symlink())
    };
    if !real_dir(&state) {
        return None;
    }
    let dir = state.join("edit-diagnostics");
    if std::fs::symlink_metadata(&dir).is_err() {
        let mut b = std::fs::DirBuilder::new();
        #[cfg(unix)]
        std::os::unix::fs::DirBuilderExt::mode(&mut b, 0o700);
        let _ = b.create(&dir);
    }
    real_dir(&dir).then_some(dir)
}

/// One checker per project at a time: an exclusive, non-blocking `flock` on a file in
/// [`lock_dir`] named by the project's digest (opened without following a link).
fn lock(root: &Path) -> Lock {
    #[cfg(unix)]
    {
        use std::os::unix::fs::OpenOptionsExt;
        use std::os::unix::io::AsRawFd;
        let Some(dir) = lock_dir() else {
            return Lock::None;
        };
        let key = crate::digest(root.as_os_str().as_encoded_bytes());
        let path = dir.join(format!("{}.lock", &key.trim_start_matches("sha256:")[..16]));
        let Ok(file) = std::fs::OpenOptions::new()
            .read(true)
            .write(true)
            .create(true)
            .truncate(false)
            .mode(0o600)
            .custom_flags(libc::O_NOFOLLOW)
            .open(&path)
        else {
            return Lock::None;
        };
        // SAFETY: flock(2) on a descriptor this function owns.
        if unsafe { libc::flock(file.as_raw_fd(), libc::LOCK_EX | libc::LOCK_NB) } == 0 {
            Lock::Held(file)
        } else {
            Lock::Busy
        }
    }
    #[cfg(not(unix))]
    {
        let _ = root;
        Lock::None
    }
}

/// What one checker run gave.
#[derive(Debug, Default)]
pub struct Ran {
    /// `Some(code)` when it ended by itself (`None` code: ended by a signal).
    pub exit: Option<Option<i32>>,
    pub missing: bool,
    pub start_failed: bool,
    pub timed_out: bool,
    pub output: Vec<u8>,
    pub cut: bool,
}

/// Runs `argv` + `rel` in `root` without a shell, bounded by `timeout`.
pub fn run_checker(argv: &[String], root: &Path, rel: &Path, timeout: Duration) -> Ran {
    let mut cmd = Command::new(&argv[0]);
    // S1 audit B1 (CWE-88): `./<rel>` can never start with `-`, so a file named `-n.c` or
    // `--config=x.py` reaches the checker as a path, not as an option.
    cmd.args(&argv[1..])
        .arg(Path::new(".").join(rel))
        .current_dir(root)
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());
    #[cfg(unix)]
    std::os::unix::process::CommandExt::process_group(&mut cmd, 0);
    let mut child = match cmd.spawn() {
        Ok(c) => c,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => {
            return Ran {
                missing: true,
                ..Ran::default()
            }
        }
        Err(_) => {
            return Ran {
                start_failed: true,
                ..Ran::default()
            }
        }
    };
    let (tx, rx) = mpsc::channel();
    for (i, stream) in [
        child
            .stdout
            .take()
            .map(|s| Box::new(s) as Box<dyn Read + Send>),
        child
            .stderr
            .take()
            .map(|s| Box::new(s) as Box<dyn Read + Send>),
    ]
    .into_iter()
    .enumerate()
    {
        let Some(stream) = stream else { continue };
        let tx = tx.clone();
        std::thread::spawn(move || {
            let _ = tx.send((i, bounded(stream)));
        });
    }
    drop(tx);
    let deadline = Instant::now() + timeout;
    let mut ran = Ran::default();
    loop {
        match child.try_wait() {
            Ok(Some(status)) => {
                ran.exit = Some(status.code());
                // S1 audit N2: a descendant the checker left in its group (`sleep 120 &`)
                // would linger and keep the output pipes open; the group goes with it.
                kill_group(&mut child);
                break;
            }
            Ok(None) if Instant::now() >= deadline => {
                kill_group(&mut child);
                ran.timed_out = true;
                break;
            }
            Ok(None) => std::thread::sleep(Duration::from_millis(5)),
            Err(_) => {
                kill_group(&mut child);
                ran.start_failed = true;
                break;
            }
        }
    }
    // The group is gone by now; a descendant that moved to another process group could still
    // keep a pipe open, so the wait stays bounded.
    let mut parts: Vec<(usize, Vec<u8>, bool)> = Vec::new();
    let until = Instant::now() + Duration::from_millis(500);
    while parts.len() < 2 {
        let left = until.saturating_duration_since(Instant::now());
        match rx.recv_timeout(left) {
            Ok((i, (bytes, cut))) => parts.push((i, bytes, cut)),
            Err(_) => break,
        }
    }
    parts.sort_by_key(|p| p.0);
    for (_, bytes, cut) in parts {
        ran.output.extend_from_slice(&bytes);
        if !ran.output.ends_with(b"\n") && !ran.output.is_empty() {
            ran.output.push(b'\n');
        }
        ran.cut |= cut;
    }
    ran
}

/// Reads up to [`READ_MAX`] bytes and drains the rest; `true` when something was left unread.
fn bounded(mut stream: impl Read) -> (Vec<u8>, bool) {
    let mut kept = Vec::new();
    let _ = (&mut stream).take(READ_MAX as u64).read_to_end(&mut kept);
    let mut sink = [0u8; 8192];
    let mut cut = false;
    while let Ok(n) = stream.read(&mut sink) {
        if n == 0 {
            break;
        }
        cut = true;
    }
    (kept, cut)
}

fn kill_group(child: &mut std::process::Child) {
    #[cfg(unix)]
    {
        if let Ok(pid) = i32::try_from(child.id()) {
            // SAFETY: kill(2) on the process group this function's spawn created.
            unsafe { libc::kill(-pid, libc::SIGKILL) };
        }
    }
    let _ = child.kill();
    let _ = child.wait();
}

/// One parsed diagnostic.
#[derive(Debug, PartialEq, Eq)]
pub struct Diagnostic {
    pub path: String,
    pub line: u64,
    pub col: Option<u64>,
    pub severity: String,
    pub message: String,
}

const SEVERITIES: [&str; 5] = ["error", "warning", "note", "info", "help"];

/// A `path:line[:col]: [severity:] message` line, or `None`.
pub fn parse_line(line: &str, root: &Path) -> Option<Diagnostic> {
    let mut parts = line.splitn(4, ':');
    let path = parts.next()?.trim();
    let line_no: u64 = parts.next()?.trim().parse().ok()?;
    let third = parts.next()?;
    let (col, rest) = match third.trim().parse::<u64>() {
        Ok(c) => (Some(c), parts.next().unwrap_or("").to_string()),
        Err(_) => (
            None,
            match parts.next() {
                Some(more) => format!("{third}:{more}"),
                None => third.to_string(),
            },
        ),
    };
    if path.is_empty() || line_no == 0 {
        return None;
    }
    let rest = rest.trim();
    let (severity, message) = match rest.split_once(':') {
        Some((word, msg))
            if SEVERITIES.contains(
                &word
                    .split('[')
                    .next()
                    .unwrap_or("")
                    .trim()
                    .to_ascii_lowercase()
                    .as_str(),
            ) =>
        {
            (
                word.split('[')
                    .next()
                    .unwrap_or("")
                    .trim()
                    .to_ascii_lowercase(),
                msg.trim(),
            )
        }
        _ => ("issue".to_string(), rest),
    };
    Some(Diagnostic {
        path: shown_path(path, root),
        line: line_no,
        col,
        severity,
        message: cut_chars(message, MESSAGE_CHARS),
    })
}

/// A project-relative path; an absolute path outside the project is not shown.
fn shown_path(path: &str, root: &Path) -> String {
    let p = Path::new(path);
    let rel = if p.is_absolute() {
        match p.strip_prefix(root) {
            Ok(rel) => rel,
            Err(_) => return "<outside the project>".into(),
        }
    } else {
        p
    };
    // The checker got `./<rel>` and may echo it, or a `/./` inside an absolute path.
    rel.components()
        .filter(|c| !matches!(c, Component::CurDir))
        .collect::<PathBuf>()
        .to_string_lossy()
        .into_owned()
}

/// S1 audit N1: the checker's text without ANSI escape sequences (`ESC [` … final byte, or
/// `ESC` and one more character) and without C0 control characters other than tab and line
/// feed (a carriage return included), so none of them reaches the context.
pub fn clean(text: &str) -> String {
    let mut out = String::with_capacity(text.len());
    let mut chars = text.chars().peekable();
    while let Some(c) = chars.next() {
        match c {
            '\u{1b}' => {
                if chars.peek() == Some(&'[') {
                    chars.next();
                    // Parameter and intermediate bytes, then one final byte (0x40-0x7E).
                    for d in chars.by_ref() {
                        if ('\u{40}'..='\u{7e}').contains(&d) {
                            break;
                        }
                    }
                } else {
                    chars.next();
                }
            }
            '\t' | '\n' => out.push(c),
            c if c.is_control() => {}
            c => out.push(c),
        }
    }
    out
}

fn cut_chars(s: &str, n: usize) -> String {
    if s.chars().count() <= n {
        s.to_string()
    } else {
        let mut out: String = s.chars().take(n.saturating_sub(1)).collect();
        out.push('…');
        out
    }
}

/// The context text for one run; empty for a clean run.
pub fn render(ran: &Ran, name: &str, shown: &str, root: &Path, rule: &EditRule) -> String {
    let head = format!("MoAI edit-diagnostics ({name}, {shown})");
    if ran.missing {
        return format!("{head}: the checker `{name}` was not found; this edit was not checked");
    }
    if ran.start_failed {
        return format!("{head}: the checker could not be started; this edit was not checked");
    }
    if ran.timed_out {
        return format!(
            "{head}: stopped after {} s with its whole process group; nothing is reported",
            rule.timeout_secs
        );
    }
    let text = clean(&String::from_utf8_lossy(&ran.output));
    let replaced = std::str::from_utf8(&ran.output).is_err();
    let found: Vec<Diagnostic> = text.lines().filter_map(|l| parse_line(l, root)).collect();
    let exit = ran.exit.flatten();
    if found.is_empty() && exit == Some(0) && !ran.cut && !replaced {
        return String::new();
    }
    let mut lines = Vec::new();
    for d in found.iter().take(rule.max_lines) {
        let at = match d.col {
            Some(c) => format!("{}:{c}", d.line),
            None => d.line.to_string(),
        };
        let whole = format!("{} {at} {} {}", d.path, d.severity, d.message);
        let line = if crate::redact::contains_secret_shape(&whole) {
            if crate::redact::contains_secret_shape(&d.path) {
                format!("[masked] {at} {} {MASKED}", d.severity)
            } else {
                format!("{} {at} {} {MASKED}", d.path, d.severity)
            }
        } else {
            whole
        };
        lines.push(line);
    }
    // The whole text, summary included, stays within `max_chars`: drop lines from the end
    // until it fits (the summary then counts them as not shown).
    loop {
        let out = compose(&head, ran, &found, exit, replaced, &lines);
        if out.chars().count() <= rule.max_chars || lines.is_empty() {
            return cut_chars(&out, rule.max_chars);
        }
        lines.pop();
    }
}

fn compose(
    head: &str,
    ran: &Ran,
    found: &[Diagnostic],
    exit: Option<i32>,
    replaced: bool,
    lines: &[String],
) -> String {
    let mut out = format!(
        "{head}: {} diagnostic line(s), exit {}",
        found.len(),
        exit.map_or_else(|| "by signal".to_string(), |c| c.to_string())
    );
    if found.len() > lines.len() {
        out.push_str(&format!("; {} not shown", found.len() - lines.len()));
    }
    if found.is_empty() && exit != Some(0) {
        out.push_str("; no line in the path:line:col form");
    }
    if ran.cut {
        out.push_str(&format!(
            "; output cut at {} KiB per stream",
            READ_MAX / 1024
        ));
    }
    if replaced {
        out.push_str("; bytes that are not UTF-8 were replaced");
    }
    for l in lines {
        out.push('\n');
        out.push_str(l);
    }
    out
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn diagnostic_lines_parse() {
        let root = Path::new("/p");
        let d = parse_line("src/a.rs:3:7: error[E0425]: cannot find x", root).unwrap();
        assert_eq!(
            (d.path.as_str(), d.line, d.col, d.severity.as_str()),
            ("src/a.rs", 3, Some(7), "error")
        );
        assert_eq!(d.message, "cannot find x");
        let d = parse_line("a.py:1:8: F401 `os` imported but unused", root).unwrap();
        assert_eq!(
            (d.severity.as_str(), d.message.as_str()),
            ("issue", "F401 `os` imported but unused")
        );
        let d = parse_line("a.c:9: warning: unused", root).unwrap();
        assert_eq!((d.col, d.severity.as_str()), (None, "warning"));
        let d = parse_line("/p/src/x.ts:2:1: note: n", root).unwrap();
        assert_eq!(d.path, "src/x.ts");
        let d = parse_line("/q/x.ts:2:1: note: n", root).unwrap();
        assert_eq!(d.path, "<outside the project>");
        for not in ["plain text", "a:b:c", "x:0:1: error: e", ":3:1: error: e"] {
            assert_eq!(parse_line(not, root), None, "{not}");
        }
        let long = format!("a.rs:1:1: error: {}", "y".repeat(500));
        let msg = parse_line(&long, root).unwrap().message;
        assert_eq!(msg.chars().count(), MESSAGE_CHARS);
        assert!(msg.ends_with('…'));
        let exact = format!("a.rs:1:1: error: {}", "y".repeat(MESSAGE_CHARS));
        assert_eq!(
            parse_line(&exact, root).unwrap().message,
            "y".repeat(MESSAGE_CHARS)
        );
        // The `./<rel>` the checker was given comes back without `./`.
        assert_eq!(
            parse_line("./-x.rs:1:1: error: e", root).unwrap().path,
            "-x.rs"
        );
        assert_eq!(
            parse_line("/p/./src/a.rs:1:1: error: e", root)
                .unwrap()
                .path,
            "src/a.rs"
        );
    }

    #[test]
    fn control_characters_and_escapes_are_removed() {
        assert_eq!(
            clean("a\u{7}b\u{1b}[1;31mc\u{1b}[0m\r\nd\te\u{1b}Mf\u{0}g\u{7f}h"),
            "abc\nd\tefgh"
        );
        assert_eq!(clean("plain text\n"), "plain text\n");
        // An escape at the very end is dropped, not kept half.
        assert_eq!(clean("x\u{1b}"), "x");
        assert_eq!(clean("x\u{1b}[31"), "x");
    }
}
