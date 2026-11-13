//! Read-only clone/workspace identity and state-root resolver (Issue #1988, 01-A).
//!
//! Boundaries come from git itself, never from whether `.git` is a directory or a gitfile:
//! `git rev-parse` reports the common dir (clone boundary) and the git dir (workspace
//! boundary). Nothing is written. The Go generation's `.moai/`, `~/.moai/` and `MOAI_HOME`
//! are never returned as a state root.
//!
//! Provisional decisions (leader RC1, decisions log 18:1x, pending P05 and P12): git runs as a
//! fixed-argv subprocess, ids are sha256 of canonical paths, the home root is fixed at
//! `~/.moai-cli`, and the output keeps the current contract (no envelope, exit 0/1).
use crate::{digest, read_profile, HOME_STATE_DIR};
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

/// Caller variables that would redirect git to another repository, inject configuration
/// (`core.bare`, `core.worktree`, ...) or change how it discovers one (Issue #2033).
const GIT_ENV_REMOVED: &[&str] = &[
    "GIT_DIR",
    "GIT_WORK_TREE",
    "GIT_COMMON_DIR",
    "GIT_INDEX_FILE",
    "GIT_NAMESPACE",
    "GIT_CONFIG_PARAMETERS",
    "GIT_CONFIG_COUNT",
    "GIT_CEILING_DIRECTORIES",
    "GIT_DISCOVERY_ACROSS_FILESYSTEM",
];
/// Under `LC_ALL=C` git starts its "no repository here" message with this text.
pub(crate) const NOT_A_REPOSITORY: &str = "not a git repository";

/// Removes every variable in [`GIT_ENV_REMOVED`] and the numbered
/// `GIT_CONFIG_KEY_<n>`/`GIT_CONFIG_VALUE_<n>` pairs from `cmd`, and fixes the locale so
/// failures can be classified. Shared by every git subprocess (`plan`, Issue #2043).
pub(crate) fn scrub_git_env(cmd: &mut Command) {
    for key in GIT_ENV_REMOVED {
        cmd.env_remove(key);
    }
    // GIT_CONFIG_KEY_<n>/GIT_CONFIG_VALUE_<n> are numbered, so they are matched by prefix.
    for (key, _) in std::env::vars_os() {
        let k = key.to_string_lossy();
        if k.starts_with("GIT_CONFIG_KEY_") || k.starts_with("GIT_CONFIG_VALUE_") {
            cmd.env_remove(&key);
        }
    }
    cmd.env("LC_ALL", "C");
}

enum Git {
    Unavailable,
    /// git ran and failed; carries a path-free summary of its message.
    Failed(String),
    Lines(Vec<String>),
}

/// First line of git's stderr without its `fatal:`/`error:` prefix, cut before the first
/// quote or path separator so no repository path reaches the caller.
pub(crate) fn failure_summary(stderr: &[u8]) -> String {
    let text = String::from_utf8_lossy(stderr);
    let line = text.lines().next().unwrap_or("").trim();
    let line = line
        .strip_prefix("fatal:")
        .or_else(|| line.strip_prefix("error:"))
        .unwrap_or(line)
        .trim();
    let cut = line.find(['\'', '"', '/', '\\']).unwrap_or(line.len());
    let summary: String = line[..cut].trim_end().chars().take(160).collect();
    if summary.is_empty() {
        "git rev-parse failed".into()
    } else {
        summary
    }
}

/// Runs `git -C <dir> rev-parse <args>` with a fixed argv. Variables that would redirect or
/// reconfigure git are removed and the locale is fixed. stderr is read only to classify a
/// failure; what reaches the caller is `failure_summary`, never a path.
fn rev_parse(dir: &Path, args: &[&str]) -> Git {
    let mut cmd = Command::new("git");
    cmd.arg("--no-optional-locks")
        .arg("-C")
        .arg(dir)
        .arg("rev-parse")
        .args(args);
    scrub_git_env(&mut cmd);
    let out = cmd
        .env("GIT_OPTIONAL_LOCKS", "0")
        .stdin(Stdio::null())
        .stderr(Stdio::piped())
        .output();
    let Ok(out) = out else {
        return Git::Unavailable;
    };
    if !out.status.success() {
        return Git::Failed(failure_summary(&out.stderr));
    }
    if out.stdout.len() > GIT_OUTPUT_LIMIT {
        return Git::Failed("git output exceeded the size limit".into());
    }
    match String::from_utf8(out.stdout) {
        Ok(text) => Git::Lines(text.lines().map(str::to_string).collect()),
        Err(_) => Git::Failed("git output was not UTF-8".into()),
    }
}

/// `no_git` status only when git said so; any other git failure is an error whose message
/// (`git failed: <summary>`) the contract classifies as `git_failed` (Issue #2033).
fn failed(summary: &str, home: Value) -> Result<Value> {
    if summary.starts_with(NOT_A_REPOSITORY) {
        Ok(empty("no_git", "not inside a git repository", home))
    } else {
        Err(format!("git failed: {summary}"))
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
/// path, a refused symlink or a git failure (`git failed: ...`); "not a repository" and "git
/// missing" are typed statuses.
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
        Git::Failed(summary) => return failed(&summary, home),
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
            Git::Failed(summary) => return failed(&summary, home),
            Git::Lines(_) => return Err("git did not report a work tree".into()),
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
            let root = crate::state_root::path(top, "");
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
