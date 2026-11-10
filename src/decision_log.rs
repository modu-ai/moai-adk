//! Decision-log row append: `decision-append` (rc.2 #2403 item 4).
//!
//! The user-decisions log (`docs/decisions-<date>.md`) is one markdown table that many pull
//! requests extend. Copying a whole edited file from a stale checkout silently drops rows merged
//! in the meantime (WF-OBS 6770614944: 21 lines nearly lost in PR #2398). This operation only
//! inserts new rows after the table's last row and refuses everything else:
//!
//! - the log is read from the HEAD commit, and the worktree file and the index entry must be
//!   byte-identical to it (no local or staged edit to carry over);
//! - HEAD must equal the `base` the caller names. The remote is not contacted: the caller
//!   fetches the integration tip and passes it as `base`; the result says `remote_checked: false`;
//! - the file holds exactly one table; existing rows are never parsed into cells, re-rendered,
//!   edited or removed (real rows carry a bare `|` inside code spans). Only their first cell is
//!   read, for the repeated-time-key warning;
//! - new cells are strict: the header's column count, no bare `|`, no line break;
//! - the default is a preview; `--apply` writes a temporary file beside the log, renames it over
//!   the log, and re-reads it to confirm `old[..k] + rows + old[k..]` byte for byte.
//!
//! Distinct from `decision-record` (`src/decision.rs`), which appends native judgement records
//! to `.moai-cli/decisions/<run>.jsonl`; this one writes the user-decisions markdown log.
use serde::Deserialize;
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    fs,
    io::Write,
    path::{Component, Path, PathBuf},
    process::Command,
};

type Result<T> = std::result::Result<T, String>;

#[derive(Deserialize)]
#[serde(deny_unknown_fields)]
pub struct AppendRequest {
    pub protocol_version: u32,
    /// 40-hex commit the caller fetched as the integration tip; HEAD must equal it.
    pub base: String,
    /// New rows, each a list of cell texts (without the surrounding `|`).
    pub rows: Vec<Vec<String>>,
}

fn git(project: &Path, args: &[&str]) -> Result<(bool, Vec<u8>)> {
    let mut cmd = Command::new("git");
    cmd.arg("--no-optional-locks")
        .arg("-C")
        .arg(project)
        .args(args);
    crate::identity::scrub_git_env(&mut cmd);
    let out = cmd
        .env("GIT_OPTIONAL_LOCKS", "0")
        .output()
        .map_err(|_| "decision append git failed to start")?;
    Ok((out.status.success(), out.stdout))
}

/// A relative path made of normal components only (no `..`, no root, no `.` tricks).
fn relative_inside(log: &Path) -> Result<()> {
    let ok =
        !log.as_os_str().is_empty() && log.components().all(|c| matches!(c, Component::Normal(_)));
    if ok {
        Ok(())
    } else {
        Err("decision append refused: --log must be a relative path inside the repository".into())
    }
}

/// No component of the path below the project root may be a symlink.
fn refuse_symlinks(root: &Path, log: &Path) -> Result<()> {
    let mut p = root.to_path_buf();
    for c in log.components() {
        p.push(c);
        if let Ok(m) = fs::symlink_metadata(&p) {
            if m.file_type().is_symlink() {
                return Err("decision append refused: --log passes through a symlink".into());
            }
        }
    }
    Ok(())
}

fn is_separator(line: &str) -> bool {
    let t = line.trim();
    t.starts_with('|')
        && t.ends_with('|')
        && t.len() > 2
        && t[1..t.len() - 1].split('|').all(|c| {
            let c = c.trim().trim_start_matches(':').trim_end_matches(':');
            c.len() >= 3 && c.chars().all(|ch| ch == '-')
        })
}

fn is_row(line: &str) -> bool {
    line.trim_start().starts_with('|')
}

/// Column count of the header line (the header carries no code spans with pipes).
fn header_columns(line: &str) -> usize {
    let t = line.trim();
    t.trim_start_matches('|')
        .trim_end_matches('|')
        .split('|')
        .count()
}

/// First cell of an existing row, read without splitting the rest of the row.
fn first_cell(line: &str) -> String {
    let t = line.trim_start().trim_start_matches('|');
    t.split('|').next().unwrap_or("").trim().to_string()
}

/// A new cell may hold an escaped `\|` but no bare `|` and no line break or other control
/// character.
fn cell_ok(cell: &str) -> bool {
    let mut prev = '\0';
    for ch in cell.chars() {
        if ch.is_control() || (ch == '|' && prev != '\\') {
            return false;
        }
        prev = ch;
    }
    true
}

struct Table {
    columns: usize,
    /// Byte offset just after the last row's line terminator.
    insert_at: usize,
    /// 1-based line number of the last row.
    last_row_line: usize,
    first_cells: Vec<String>,
}

fn find_table(text: &str) -> Result<Table> {
    let lines: Vec<&str> = text.split_inclusive('\n').collect();
    let seps: Vec<usize> = (1..lines.len())
        .filter(|&i| is_separator(lines[i]) && is_row(lines[i - 1]))
        .collect();
    if seps.len() != 1 {
        return Err(format!(
            "decision append refused: the log must hold exactly one table (found {})",
            seps.len()
        ));
    }
    let sep = seps[0];
    let columns = header_columns(lines[sep - 1]);
    let mut last = sep;
    let mut first_cells = Vec::new();
    for (i, line) in lines.iter().enumerate().skip(sep + 1) {
        if !is_row(line) {
            break;
        }
        last = i;
        first_cells.push(first_cell(line));
    }
    if !lines[last].ends_with('\n') {
        return Err("decision append refused: the table's last line has no line terminator".into());
    }
    let insert_at = lines[..=last].iter().map(|l| l.len()).sum();
    Ok(Table {
        columns,
        insert_at,
        last_row_line: last + 1,
        first_cells,
    })
}

fn sha256_hex(bytes: &[u8]) -> String {
    format!("{:x}", Sha256::digest(bytes))
}

pub fn append(
    project: &Path,
    log: &Path,
    request: &AppendRequest,
    apply: bool,
    permission: Option<&str>,
) -> Result<Value> {
    if request.protocol_version != crate::contract::PROTOCOL_VERSION {
        return Err("unsupported protocol_version; supported: 1".into());
    }
    if apply {
        let p = permission.unwrap_or("");
        if ["", "null", "placeholder"].contains(&p.trim())
            || p.len() > 128
            || p.chars().any(char::is_control)
        {
            return Err("permission reference required".into());
        }
    }
    if request.rows.is_empty() {
        return Err("decision append refused: give at least one row".into());
    }
    if request.base.len() != 40
        || !request
            .base
            .bytes()
            .all(|b| matches!(b, b'0'..=b'9' | b'a'..=b'f'))
    {
        return Err("decision append refused: base must be a 40-hex commit".into());
    }
    relative_inside(log)?;
    let root = fs::canonicalize(project).map_err(|_| "decision append project unreadable")?;
    refuse_symlinks(&root, log)?;
    let log_str = log
        .to_str()
        .ok_or("decision append refused: --log must be UTF-8")?;

    // `HEAD:<path>` is relative to the repository top, so the project must be that top.
    let (ok, top) = git(&root, &["rev-parse", "--show-toplevel"])?;
    let top = String::from_utf8_lossy(&top).trim().to_string();
    if !ok || fs::canonicalize(&top).ok().as_deref() != Some(root.as_path()) {
        return Err("decision append refused: PROJECT must be the repository top directory".into());
    }
    let (ok, head) = git(&root, &["rev-parse", "--verify", "HEAD^{commit}"])?;
    if !ok {
        return Err("decision append blocked: the project has no HEAD commit".into());
    }
    let head = String::from_utf8_lossy(&head).trim().to_string();
    if head != request.base {
        return Err(format!(
            "decision append blocked: HEAD is not the given base (HEAD {head}); fetch the tip and work from a worktree at it"
        ));
    }
    // The HEAD blob is the only source: a path absent from HEAD (untracked, or only staged)
    // has no blob and is refused here.
    let (in_head, old) = git(&root, &["show", &format!("HEAD:{log_str}")])?;
    if !in_head {
        return Err("decision append refused: --log is not tracked at HEAD".into());
    }
    let disk = fs::read(root.join(log)).map_err(|_| "decision append log unreadable")?;
    let (index_same, _) = git(
        &root,
        &["diff", "--cached", "--quiet", "HEAD", "--", log_str],
    )?;
    if disk != old || !index_same {
        return Err(
            "decision append blocked: the log differs from HEAD in the worktree or the index; commit or discard local edits first".into(),
        );
    }
    let text = String::from_utf8(old.clone())
        .map_err(|_| "decision append refused: the log is not UTF-8")?;
    let table = find_table(&text)?;

    let mut added = String::new();
    let mut warnings = Vec::new();
    for (n, row) in request.rows.iter().enumerate() {
        if row.len() != table.columns {
            return Err(format!(
                "decision append refused: row {} has {} cells; the table has {}",
                n + 1,
                row.len(),
                table.columns
            ));
        }
        if let Some(bad) = row.iter().position(|c| !cell_ok(c)) {
            return Err(format!(
                "decision append refused: row {} cell {} holds a bare | or a line break",
                n + 1,
                bad + 1
            ));
        }
        let key = row[0].trim();
        if table.first_cells.iter().any(|c| c == key) {
            warnings.push(format!("time key {key} already has a row; added anyway"));
        }
        added.push_str("| ");
        added.push_str(&row.iter().map(|c| c.trim()).collect::<Vec<_>>().join(" | "));
        added.push_str(" |\n");
    }
    let mut new = Vec::with_capacity(old.len() + added.len());
    new.extend_from_slice(&old[..table.insert_at]);
    new.extend_from_slice(added.as_bytes());
    new.extend_from_slice(&old[table.insert_at..]);

    let mut result = json!({
        "log": log_str,
        "base": request.base,
        "head": head,
        "insert_after_line": table.last_row_line,
        "rows_added": request.rows.len(),
        "lines_removed": 0,
        "sha256_before": sha256_hex(&old),
        "sha256_after": sha256_hex(&new),
        "warnings": warnings,
        "remote_checked": false,
        "note": "base equality is checked against the local HEAD only; the caller fetched the integration tip",
    });
    if !apply {
        result["mode"] = json!("preview");
        result["writes"] = json!([]);
        return Ok(result);
    }
    let target = root.join(log);
    let dir = target
        .parent()
        .map(Path::to_path_buf)
        .unwrap_or_else(|| root.clone());
    let tmp: PathBuf = dir.join(format!(".decision-append-{}.tmp", std::process::id()));
    let write = (|| -> std::io::Result<()> {
        let mut f = fs::OpenOptions::new()
            .write(true)
            .create_new(true)
            .open(&tmp)?;
        f.write_all(&new)?;
        f.sync_all()?;
        if let Ok(m) = fs::metadata(&target) {
            fs::set_permissions(&tmp, m.permissions())?;
        }
        fs::rename(&tmp, &target)
    })();
    if write.is_err() {
        let _ = fs::remove_file(&tmp);
        return Err("decision append log replace failed; the log is unchanged".into());
    }
    let back = fs::read(&target).map_err(|_| "decision append log unreadable after write")?;
    if back != new {
        return Err("decision append log replace failed: read-back differs".into());
    }
    result["mode"] = json!("applied");
    result["writes"] = json!([log_str]);
    Ok(result)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn separator_needs_three_dashes_per_column() {
        assert!(is_separator("| --- | --- |\n"));
        assert!(is_separator("| :---: | ---: |\n"));
        assert!(!is_separator("| -- | --- |\n"));
        assert!(!is_separator("| --- | x |\n"));
    }

    #[test]
    fn new_cells_allow_only_escaped_pipes() {
        assert!(cell_ok("plain"));
        assert!(cell_ok("`a\\|b`"));
        assert!(!cell_ok("a|b"));
        assert!(!cell_ok("`a|b`"));
        assert!(!cell_ok("a\nb"));
        assert!(!cell_ok("a\tb"));
    }

    #[test]
    fn first_cell_ignores_pipes_later_in_the_row() {
        assert_eq!(first_cell("| 19:5x | t | `cc|codex` | x |\n"), "19:5x");
    }

    #[test]
    fn relative_paths_only() {
        assert!(relative_inside(Path::new("docs/decisions.md")).is_ok());
        assert!(relative_inside(Path::new("../x.md")).is_err());
        assert!(relative_inside(Path::new("/x.md")).is_err());
        assert!(relative_inside(Path::new("")).is_err());
    }
}
