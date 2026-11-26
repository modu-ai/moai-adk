//! The SessionStart handler (H1) of `moai hook session-context`
//! (SPEC-20261123-hook-session-start). It builds a short context text: the branch of the
//! checkout, the card reserved on it and, after a compaction, where to read the card's newest
//! RESUME note.
//!
//! What it reads: `.git` and its `HEAD` as plain files (no other program is started), and the
//! ledger once through one read-only connection. It writes nothing and never decides anything:
//! every failure only shortens the text, and the ledger failure is named to the caller so the
//! entry can record one receipt line.

use moai_cli::ledger::{self, LedgerError};
use serde_json::{json, Map, Value};
use std::fs;
use std::io::Read;
use std::os::unix::fs::OpenOptionsExt;
use std::path::{Path, PathBuf};
use std::time::Duration;

/// The ledger wait: a session start must not hang on a busy ledger.
const LEDGER_WAIT: Duration = Duration::from_millis(200);
/// `HEAD` is read up to this many bytes; one more byte reads as unknown.
const HEAD_LIMIT: u64 = 1024;
const BRANCH_LIMIT: usize = 100;
const TITLE_LIMIT: usize = 80;
/// How far up from the working folder the `.git` entry is looked for.
const CLIMB_LIMIT: usize = 64;

/// The answer: the JSON object for stdout, and the reason word of a ledger failure to record.
pub struct Context {
    pub output: Value,
    pub ledger_failure: Option<&'static str>,
}

/// The handler. `project_root` is the folder that holds `.moai/project.yaml`; `cwd` is the
/// working folder of the input, else of the process; `source` is the input's `source`.
pub fn context(
    input: &Map<String, Value>,
    project_root: &Path,
    cwd: &Path,
    home: Option<&Path>,
    project_id: Option<&str>,
) -> Context {
    let from = input
        .get("cwd")
        .and_then(Value::as_str)
        .filter(|c| c.starts_with('/'))
        .map_or_else(|| cwd.to_path_buf(), PathBuf::from);
    let checkout = find_checkout(&from);
    let root = checkout
        .as_ref()
        .map_or_else(|| project_root.to_path_buf(), |c| c.root.clone());
    let branch = checkout.as_ref().and_then(|c| read_head(&c.git_dir));
    let mut first = match &branch {
        Some(Head::Branch(name)) => format!("MoAI: branch {}", clean(name, BRANCH_LIMIT)),
        Some(Head::Detached(hex)) => format!("MoAI: detached {hex}"),
        None => "MoAI: branch unknown".to_owned(),
    };
    let branch_name = match &branch {
        Some(Head::Branch(name)) => Some(name.as_str()),
        _ => None,
    };
    let mut second = None;
    let mut ledger_failure = None;
    match (home, project_id) {
        (Some(home), Some(id)) => match reserved_card(home, id, &root, branch_name) {
            Ok(Some(card)) => {
                first.push_str(&format!(
                    " · card C-{} \"{}\" ({})",
                    card.id,
                    clean(&card.title, TITLE_LIMIT),
                    card.state
                ));
                if input.get("source").and_then(Value::as_str) == Some("compact") {
                    second = Some(match card.resume_note {
                        // A statement of fact, not an order: the host documents that context
                        // added at a session start is read as information.
                        Some(note) => format!(
                            "The newest RESUME note of C-{id} is #{note} (moai card show C-{id})",
                            id = card.id
                        ),
                        None => format!(
                            "C-{id} has no RESUME note yet (moai card show C-{id})",
                            id = card.id
                        ),
                    });
                }
            }
            Ok(None) => first.push_str(" · no card reserved here"),
            Err(reason) => ledger_failure = Some(reason),
        },
        _ => ledger_failure = Some("ledger_absent"),
    }
    let mut text = match second {
        Some(line) => format!("{first}\n{line}"),
        None => first,
    };
    if let Some(line) = checkout
        .as_ref()
        .and_then(|c| guard_line(&root, &c.git_dir))
    {
        text.push('\n');
        text.push_str(&line);
    }
    Context {
        output: json!({"hookSpecificOutput": {
            "hookEventName": "SessionStart",
            "additionalContext": text,
        }}),
        ledger_failure,
    }
}

struct Checkout {
    root: PathBuf,
    git_dir: PathBuf,
}

/// The one line that names a missing integration branch guard (SPEC-20261125-integration-guard-
/// install REQ-4): only for a profile that names an integration branch, exactly when the hook
/// `reference-transaction` of the common git folder is absent or not executable or the record
/// `moai/integration-branch` is absent. It reads the profile and looks at two paths, starts no
/// program and decides nothing; a read that fails adds no line.
fn guard_line(root: &Path, git_dir: &Path) -> Option<String> {
    let branch = moai_cli::card::landed::profile_branch(root).ok()?;
    let common = match read_small(&git_dir.join("commondir"), 4096) {
        Some(text) => {
            let line = text.lines().next()?.trim().to_owned();
            git_dir.join(line)
        }
        None => git_dir.to_path_buf(),
    };
    let hook = fs::metadata(common.join("hooks/reference-transaction")).ok();
    let mut missing = Vec::new();
    match hook {
        None => missing.push("the hook reference-transaction is missing"),
        Some(meta) if !meta.is_file() => missing.push("the hook reference-transaction is missing"),
        Some(meta) => {
            use std::os::unix::fs::PermissionsExt;
            if meta.permissions().mode() & 0o111 == 0 {
                missing.push("the hook reference-transaction is not executable");
            }
        }
    }
    if !fs::metadata(common.join("moai/integration-branch")).is_ok_and(|m| m.is_file()) {
        missing.push("the record moai/integration-branch is missing");
    }
    if missing.is_empty() {
        return None;
    }
    Some(format!(
        "MoAI: integration branch {} is not guarded: {} (moai git-hooks-plan)",
        clean(&branch, BRANCH_LIMIT),
        missing.join(" and ")
    ))
}

/// The nearest folder at or above `from` that holds a `.git` entry: a folder (a plain checkout)
/// or a file `gitdir: <path>` (a linked worktree; the path may be relative to the folder).
fn find_checkout(from: &Path) -> Option<Checkout> {
    for dir in from.ancestors().take(CLIMB_LIMIT) {
        let entry = dir.join(".git");
        let Ok(meta) = fs::metadata(&entry) else {
            continue;
        };
        if meta.is_dir() {
            return Some(Checkout {
                root: dir.to_path_buf(),
                git_dir: entry,
            });
        }
        if meta.is_file() {
            let text = read_small(&entry, 4096)?;
            let target = text.lines().next()?.strip_prefix("gitdir:")?.trim();
            return Some(Checkout {
                root: dir.to_path_buf(),
                git_dir: dir.join(target),
            });
        }
    }
    None
}

/// Up to `limit` bytes of a regular file as text. A file of more bytes, a file that is not
/// regular (a FIFO opens without waiting and is refused) and a read error are `None`.
fn read_small(path: &Path, limit: u64) -> Option<String> {
    let file = fs::OpenOptions::new()
        .read(true)
        .custom_flags(libc::O_NONBLOCK)
        .open(path)
        .ok()?;
    if !file.metadata().ok()?.is_file() {
        return None;
    }
    let mut bytes = Vec::new();
    file.take(limit + 1).read_to_end(&mut bytes).ok()?;
    if bytes.len() as u64 > limit {
        return None;
    }
    String::from_utf8(bytes).ok()
}

enum Head {
    Branch(String),
    Detached(String),
}

fn read_head(git_dir: &Path) -> Option<Head> {
    let text = read_small(&git_dir.join("HEAD"), HEAD_LIMIT)?;
    let line = text.lines().next()?.trim();
    if let Some(name) = line.strip_prefix("ref: refs/heads/") {
        return (!name.is_empty()).then(|| Head::Branch(name.to_owned()));
    }
    let hex_digits = line.len() >= 12 && line.bytes().all(|b| b.is_ascii_hexdigit());
    hex_digits.then(|| Head::Detached(line[..12].to_ascii_lowercase()))
}

/// One line of text, safe to show: whitespace runs become one space, every other control
/// character goes, and more than `limit` characters are cut to `limit` with the `…` included.
fn clean(text: &str, limit: usize) -> String {
    let mut out = String::new();
    let mut pending_space = false;
    for c in text.chars() {
        if c.is_whitespace() {
            pending_space = !out.is_empty();
        } else if !c.is_control() {
            if pending_space {
                out.push(' ');
                pending_space = false;
            }
            out.push(c);
        }
    }
    if out.chars().count() > limit {
        out = out.chars().take(limit - 1).collect::<String>() + "…";
    }
    out
}

struct Card {
    id: i64,
    title: String,
    state: String,
    resume_note: Option<i64>,
}

/// The card with an active reservation on this checkout (`worktree` is the root, preferred, or
/// `branch` is the branch) and the id of its newest RESUME note. The ledger is opened read-only,
/// once; a failure comes back as the reason word of the receipt.
fn reserved_card(
    home: &Path,
    project_id: &str,
    root: &Path,
    branch: Option<&str>,
) -> Result<Option<Card>, &'static str> {
    let ledger = ledger::open::open_read_only_within(home, project_id, LEDGER_WAIT).map_err(
        |e| match e {
            LedgerError::Absent => "ledger_absent",
            LedgerError::PathLink(_) => "ledger_link",
            LedgerError::Busy => "ledger_busy",
            _ => "ledger_unreadable",
        },
    )?;
    let conn = ledger.conn();
    let busy = |e: rusqlite::Error| match e.sqlite_error_code() {
        Some(rusqlite::ErrorCode::DatabaseBusy | rusqlite::ErrorCode::DatabaseLocked) => {
            "ledger_busy"
        }
        _ => "ledger_unreadable",
    };
    let mut stmt = conn
        .prepare(
            "SELECT c.card_id, c.title, c.state, r.worktree, r.branch \
             FROM reservation r JOIN card c ON c.card_id = r.card_id \
             WHERE r.status IN ('held', 'held_for_review', 'expired_candidate') \
             ORDER BY r.acquired_at DESC, r.reservation_id DESC",
        )
        .map_err(busy)?;
    let rows = stmt
        .query_map([], |row| {
            Ok((
                row.get::<_, i64>(0)?,
                row.get::<_, String>(1)?,
                row.get::<_, String>(2)?,
                row.get::<_, Option<String>>(3)?,
                row.get::<_, Option<String>>(4)?,
            ))
        })
        .map_err(busy)?;
    let mut by_branch = None;
    let mut found = None;
    for row in rows {
        let (id, title, state, worktree, row_branch) = row.map_err(busy)?;
        if worktree.as_deref().is_some_and(|w| same_folder(w, root)) {
            found = Some((id, title, state));
            break;
        }
        if by_branch.is_none() && branch.is_some() && row_branch.as_deref() == branch {
            by_branch = Some((id, title, state));
        }
    }
    let Some((id, title, state)) = found.or(by_branch) else {
        return Ok(None);
    };
    let resume_note = conn
        .query_row(
            "SELECT note_id FROM card_note WHERE card_id = ?1 AND substr(body, 1, 7) = 'RESUME ' \
             ORDER BY note_id DESC LIMIT 1",
            [id],
            |row| row.get::<_, i64>(0),
        )
        .map(Some)
        .or_else(|e| match e {
            rusqlite::Error::QueryReturnedNoRows => Ok(None),
            other => Err(busy(other)),
        })?;
    Ok(Some(Card {
        id,
        title,
        state,
        resume_note,
    }))
}

/// The reservation's folder text names the checkout root, as written or once links are resolved.
fn same_folder(stored: &str, root: &Path) -> bool {
    let stored = Path::new(stored);
    stored == root
        || fs::canonicalize(stored)
            .ok()
            .zip(fs::canonicalize(root).ok())
            .is_some_and(|(a, b)| a == b)
}
