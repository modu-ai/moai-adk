//! Read-only clone/workspace identity and state-root resolver (Issue #1988, 01-A).
//!
//! Boundaries come from git itself, never from whether `.git` is a directory or a gitfile:
//! `git rev-parse` reports the common dir (clone boundary) and the git dir (workspace
//! boundary). Nothing is written. The Go generation's `.moai/`, `~/.moai/` and `MOAI_HOME`
//! are never returned as a state root.
//!
//! Provisional decisions (leader RC1, decisions log 18:2x, pending P05 and P12): git runs as a
//! fixed-argv subprocess, ids are sha256 of canonical paths, the home root is fixed at
//! `~/.moai-cli`, and the output keeps the current contract (no envelope, exit 0/1).
use crate::{digest, read_profile, HOME_STATE_DIR, PROJECT_STATE_DIR};
use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};

type Result<T> = std::result::Result<T, String>;

/// Upper bound on what one rev-parse call may print; four absolute paths fit easily.
const GIT_OUTPUT_LIMIT: usize = 64 * 1024;
const ID_DERIVATION: &str = "sha256-canonical-path-v0";

enum Git {
    Unavailable,
    Failed,
    Lines(Vec<String>),
}

/// Runs `git -C <dir> rev-parse <args>` with a fixed argv. Variables that would redirect git
/// to another repository are removed, the locale is fixed, and stderr is discarded so no path
/// or git message reaches the caller.
fn rev_parse(dir: &Path, args: &[&str]) -> Git {
    let out = Command::new("git")
        .arg("--no-optional-locks")
        .arg("-C")
        .arg(dir)
        .arg("rev-parse")
        .args(args)
        .env_remove("GIT_DIR")
        .env_remove("GIT_WORK_TREE")
        .env_remove("GIT_COMMON_DIR")
        .env_remove("GIT_INDEX_FILE")
        .env_remove("GIT_NAMESPACE")
        .env("GIT_OPTIONAL_LOCKS", "0")
        .env("LC_ALL", "C")
        .stdin(Stdio::null())
        .stderr(Stdio::null())
        .output();
    let Ok(out) = out else {
        return Git::Unavailable;
    };
    if !out.status.success() || out.stdout.len() > GIT_OUTPUT_LIMIT {
        return Git::Failed;
    }
    match String::from_utf8(out.stdout) {
        Ok(text) => Git::Lines(text.lines().map(str::to_string).collect()),
        Err(_) => Git::Failed,
    }
}

fn canonical(raw: &str) -> Result<PathBuf> {
    fs::canonicalize(raw).map_err(|_| "git reported a path that cannot be resolved".into())
}
fn text(p: &Path) -> Result<String> {
    p.to_str()
        .map(str::to_string)
        .ok_or_else(|| "non UTF-8 path".into())
}
fn refuse_symlink(path: &Path, what: &str) -> Result<()> {
    match fs::symlink_metadata(path) {
        Ok(m) if m.file_type().is_symlink() => Err(format!("{what} is a symlink; refused")),
        _ => Ok(()),
    }
}

// @MX:DEBT Ids are sha256 of the canonical absolute path (leader RC1 provisional, HOLD-1 A).
// @MX:CEILING Moving or renaming the repository changes every id; two clones cannot share one.
// @MX:UPGRADE Re-evaluate when 06-A #2005 consumes workspace identity in the claim journal.
fn path_id(p: &Path) -> Result<String> {
    Ok(digest(text(p)?.as_bytes()))
}

fn empty(status: &str, reason: &str, home: Value) -> Value {
    json!({
        "identity_status": status,
        "reason": reason,
        "workspace_kind": null,
        "clone_anchor": null,
        "workspace_git_dir": null,
        "work_tree": null,
        "clone_id": null,
        "workspace_id": null,
        "project_id": null,
        "profile_status": null,
        "project_state_root": null,
        "home_state_root": home,
        "id_derivation": ID_DERIVATION,
        "writes": false,
    })
}

/// Resolves the identity of `path`. Never writes. Returns `Err` only for an unusable input
/// path or a refused symlink; "not a repository" and "git missing" are typed statuses.
pub fn identity(path: &Path) -> Result<Value> {
    let meta = fs::symlink_metadata(path).map_err(|_| "path unavailable")?;
    if meta.file_type().is_symlink() {
        return Err("path is a symlink; refused".into());
    }
    if !meta.is_dir() {
        return Err("path must be a directory".into());
    }
    let home = match std::env::var_os("HOME") {
        Some(h) if !h.is_empty() => {
            let root = PathBuf::from(h).join(HOME_STATE_DIR);
            refuse_symlink(&root, "home state root")?;
            json!(text(&root)?)
        }
        _ => Value::Null,
    };
    let lines = match rev_parse(
        path,
        &[
            "--path-format=absolute",
            "--git-common-dir",
            "--git-dir",
            "--is-bare-repository",
            "--is-inside-work-tree",
        ],
    ) {
        Git::Unavailable => return Ok(empty("git_unavailable", "git was not found on PATH", home)),
        Git::Failed => return Ok(empty("no_git", "not inside a git repository", home)),
        Git::Lines(l) if l.len() == 4 => l,
        Git::Lines(_) => return Err("unexpected git rev-parse output".into()),
    };
    let common = canonical(&lines[0])?;
    let git_dir = canonical(&lines[1])?;
    let bare = lines[2] == "true";
    let in_work_tree = lines[3] == "true";
    let work_tree = if in_work_tree {
        match rev_parse(path, &["--path-format=absolute", "--show-toplevel"]) {
            Git::Lines(l) if l.len() == 1 => Some(canonical(&l[0])?),
            Git::Unavailable => {
                return Ok(empty("git_unavailable", "git was not found on PATH", home))
            }
            _ => return Err("git did not report a work tree".into()),
        }
    } else {
        None
    };
    let kind = match (bare, &work_tree) {
        (true, _) => "bare",
        (false, None) => "git_dir",
        (false, Some(_)) if common == git_dir => "primary",
        (false, Some(_)) => "linked",
    };
    let (project_root, project_id, profile_status) = match &work_tree {
        Some(top) => {
            let root = top.join(PROJECT_STATE_DIR);
            refuse_symlink(&root, "project state root")?;
            let (id, status) = match fs::symlink_metadata(root.join("project.yaml")) {
                Err(_) => (Value::Null, "missing"),
                Ok(_) => match read_profile(top) {
                    Ok((p, _)) => (json!(p.project_id), "present"),
                    Err(_) => (Value::Null, "unreadable"),
                },
            };
            (json!(text(&root)?), id, json!(status))
        }
        None => (Value::Null, Value::Null, Value::Null),
    };
    Ok(json!({
        "identity_status": "ok",
        "reason": null,
        "workspace_kind": kind,
        "clone_anchor": text(&common)?,
        "workspace_git_dir": text(&git_dir)?,
        "work_tree": work_tree.as_deref().map(text).transpose()?,
        "clone_id": path_id(&common)?,
        "workspace_id": path_id(&git_dir)?,
        "project_id": project_id,
        "profile_status": profile_status,
        "project_state_root": project_root,
        "home_state_root": home,
        "id_derivation": ID_DERIVATION,
        "writes": false,
    }))
}
