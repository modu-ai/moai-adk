//! `reports-move-plan` / `reports-move-apply` (C-126, SPEC-20261119-update-moves-reports): the
//! project update moves the untracked report files an earlier generation left inside the
//! repository (`.moai-cli/reports`, `.moai/reports`) to the user-level reports folder
//! `<HOME>/.moai/reports/<project>` that [`crate::home_workdir::reports_dir`] names.
//!
//! The plan only reads. A file git tracks is never moved (removing it from the repository is a
//! deletion the user chooses); a destination name that exists is never overwritten; a link or a
//! special file is never followed or moved. Every moved file keeps its folder structure under the
//! destination, which is owner-only. Running the apply again after a finished move moves nothing.

use crate::{contract, home_workdir};
use serde::{Deserialize, Serialize};
use std::collections::BTreeSet;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

type Result<T> = std::result::Result<T, String>;

/// The earlier report folders of a project, relative to the project, in the order they are read.
/// A name both hold goes to the first (the earlier state folder), the second is skipped.
const SOURCES: [&str; 2] = [".moai-cli/reports", ".moai/reports"];
/// More files than this make the plan `blocked`: a folder that large is not a report folder.
const MAX_FILES: usize = 20_000;
/// How many moved files a printed preview names.
pub const EXAMPLES: usize = 10;

/// One file to move.
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq)]
pub struct Item {
    /// The path relative to the project, e.g. `.moai-cli/reports/issue-bodies/1.md`.
    pub from: String,
    /// The path relative to the destination folder, e.g. `issue-bodies/1.md`.
    pub to: String,
    pub bytes: u64,
}

/// The preview `reports-move-apply` takes back on stdin.
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq)]
pub struct MovePlan {
    pub protocol_version: u32,
    pub project: String,
    /// `<HOME>/.moai/reports/<project>`.
    pub destination: String,
    /// `ready` (files to move), `unchanged` (nothing to move) or `blocked` (see `reason`).
    pub state: String,
    pub reason: String,
    /// The earlier folders that exist, relative to the project.
    pub sources: Vec<String>,
    pub moves: Vec<Item>,
    /// Destination-relative names that already exist there: kept in place, never overwritten.
    pub skipped_existing: Vec<String>,
    /// Files git tracks under the earlier folders: left alone.
    pub tracked: u64,
    /// Links and special files: neither followed nor moved.
    pub unmovable: u64,
    pub plan_digest: String,
}

impl MovePlan {
    fn digest(&self) -> String {
        let mut bound = self.clone();
        bound.plan_digest = String::new();
        crate::digest(serde_json::to_string(&bound).unwrap_or_default().as_bytes())
    }

    /// The counts and a few names, for the update preview (the whole plan goes to the apply).
    pub fn preview(&self) -> serde_json::Value {
        serde_json::json!({
            "destination": self.destination,
            "sources": self.sources,
            "to_move": self.moves.len(),
            "skipped_existing": self.skipped_existing.len(),
            "tracked_left_alone": self.tracked,
            "unmovable": self.unmovable,
            "examples": self.moves.iter().take(EXAMPLES).map(|m| m.from.as_str()).collect::<Vec<_>>(),
        })
    }
}

/// Runs git in `dir` with the caller's repository redirects removed; stdout on success.
fn git(dir: &Path, args: &[&str]) -> Result<Vec<u8>> {
    let mut cmd = Command::new("git");
    crate::identity::scrub_git_env(&mut cmd);
    let out = cmd
        .env("GIT_TERMINAL_PROMPT", "0")
        .arg("--no-optional-locks")
        .arg("-C")
        .arg(dir)
        .args(args)
        .stdin(Stdio::null())
        .output()
        .map_err(|e| format!("cannot run git: {e}"))?;
    if out.status.success() {
        Ok(out.stdout)
    } else {
        Err(String::from_utf8_lossy(&out.stderr).trim().to_string())
    }
}

/// The project-relative paths git tracks under `pathspecs`, or why git cannot tell.
fn tracked_files(root: &Path, pathspecs: &[String]) -> Result<BTreeSet<String>> {
    let mut args = vec!["ls-files", "-z", "--"];
    args.extend(pathspecs.iter().map(String::as_str));
    let out = git(root, &args)?;
    Ok(out
        .split(|b| *b == 0)
        .filter(|p| !p.is_empty())
        .map(|p| String::from_utf8_lossy(p).into_owned())
        .collect())
}

/// Every regular file under `rel` (relative to `root`, `/` separated), links and special files
/// counted in `unmovable`, folders descended but never links. Sorted.
fn walk(root: &Path, rel: &str, files: &mut Vec<(String, u64)>, unmovable: &mut u64) -> Result<()> {
    let mut stack = vec![rel.to_string()];
    while let Some(dir) = stack.pop() {
        let mut names: Vec<_> = fs::read_dir(root.join(&dir))
            .map_err(|e| format!("cannot read {dir}: {e}"))?
            .filter_map(|e| e.ok())
            .collect();
        names.sort_by_key(|e| e.file_name());
        for e in names {
            let Some(name) = e.file_name().to_str().map(str::to_string) else {
                *unmovable += 1;
                continue;
            };
            let path = format!("{dir}/{name}");
            let Ok(meta) = fs::symlink_metadata(e.path()) else {
                *unmovable += 1;
                continue;
            };
            if meta.is_dir() {
                stack.push(path);
            } else if meta.is_file() {
                files.push((path, meta.len()));
                if files.len() > MAX_FILES {
                    return Err(format!(
                        "more than {MAX_FILES} files under the report folders"
                    ));
                }
            } else {
                *unmovable += 1;
            }
        }
    }
    files.sort();
    Ok(())
}

fn blocked(root: &Path, destination: &Path, reason: String) -> MovePlan {
    let mut p = MovePlan {
        protocol_version: 1,
        project: root.display().to_string(),
        destination: destination.display().to_string(),
        state: "blocked".into(),
        reason,
        sources: Vec::new(),
        moves: Vec::new(),
        skipped_existing: Vec::new(),
        tracked: 0,
        unmovable: 0,
        plan_digest: String::new(),
    };
    p.plan_digest = p.digest();
    p
}

/// Whether `path` exists as anything at all (a link counts, a dangling one too).
fn exists(path: &Path) -> bool {
    fs::symlink_metadata(path).is_ok()
}

/// Preview the move. Reads only; creates nothing, in the project or under `home`.
pub fn plan(root: &Path, home: &Path) -> Result<MovePlan> {
    let root =
        fs::canonicalize(root).map_err(|_| contract::PROJECT_PATH_UNAVAILABLE.to_string())?;
    let destination = home_workdir::reports_dir(home, &root)?;
    let present: Vec<String> = SOURCES
        .iter()
        .filter(|s| fs::symlink_metadata(root.join(s)).is_ok_and(|m| m.is_dir()))
        .map(|s| s.to_string())
        .collect();
    let mut p = MovePlan {
        protocol_version: 1,
        project: root.display().to_string(),
        destination: destination.display().to_string(),
        state: "unchanged".into(),
        reason: "no earlier report folder in the project".into(),
        sources: present.clone(),
        moves: Vec::new(),
        skipped_existing: Vec::new(),
        tracked: 0,
        unmovable: 0,
        plan_digest: String::new(),
    };
    if present.is_empty() {
        p.plan_digest = p.digest();
        return Ok(p);
    }
    // The destination chain must be real folders (or absent): never write through a link.
    for dir in [
        home_workdir::root(home),
        home_workdir::reports_root(home),
        destination.clone(),
    ] {
        if fs::symlink_metadata(&dir).is_ok_and(|m| !m.is_dir()) {
            return Ok(blocked(
                &root,
                &destination,
                format!("{} is not a plain folder; nothing moves", dir.display()),
            ));
        }
    }
    // Which files git tracks: when git cannot tell, nothing moves (a tracked file must stay).
    let tracked = match tracked_files(&root, &present) {
        Ok(t) => t,
        Err(e) => {
            return Ok(blocked(
                &root,
                &destination,
                format!("git cannot tell which report files are tracked ({e}); nothing moves"),
            ));
        }
    };
    let mut files = Vec::new();
    for s in &present {
        if let Err(e) = walk(&root, s, &mut files, &mut p.unmovable) {
            return Ok(blocked(&root, &destination, e));
        }
    }
    let mut planned = BTreeSet::new();
    for (from, bytes) in files {
        if tracked.contains(&from) {
            p.tracked += 1;
            continue;
        }
        let source = present
            .iter()
            .find(|s| from.starts_with(&format!("{s}/")))
            .expect("a walked file sits under a present source");
        let to = from[source.len() + 1..].to_string();
        // A name the destination already holds, or an earlier source already planned, or one whose
        // parent there is not a folder, is kept where it is.
        let taken = exists(&destination.join(&to))
            || planned.contains(&to)
            || Path::new(&to).ancestors().skip(1).any(|a| {
                !a.as_os_str().is_empty()
                    && fs::symlink_metadata(destination.join(a)).is_ok_and(|m| !m.is_dir())
            });
        if taken {
            p.skipped_existing.push(to);
            continue;
        }
        planned.insert(to.clone());
        p.moves.push(Item { from, to, bytes });
    }
    if p.moves.is_empty() {
        p.reason = format!(
            "nothing to move ({} tracked left alone, {} already at the destination)",
            p.tracked,
            p.skipped_existing.len()
        );
    } else {
        p.state = "ready".into();
        p.reason = format!(
            "{} untracked report file(s) move to {}",
            p.moves.len(),
            p.destination
        );
    }
    p.plan_digest = p.digest();
    Ok(p)
}

/// Creates `dir` and its missing parents owner-only.
fn create_private_dirs(dir: &Path) -> Result<()> {
    let mut builder = fs::DirBuilder::new();
    builder.recursive(true);
    #[cfg(unix)]
    {
        use std::os::unix::fs::DirBuilderExt;
        builder.mode(0o700);
    }
    builder
        .create(dir)
        .map_err(|e| format!("cannot create {}: {e}", dir.display()))
}

/// Moves one file without ever replacing a destination: a hard link fails when the name exists,
/// and a copy across file systems opens the destination with `create_new`. The source goes only
/// after the destination holds the same bytes. The moved file keeps no group or other bits.
/// `Ok(false)` when the destination name appeared meanwhile (skipped).
fn move_one(from: &Path, to: &Path) -> Result<bool> {
    match fs::hard_link(from, to) {
        Ok(()) => {}
        Err(e) if e.kind() == std::io::ErrorKind::AlreadyExists => return Ok(false),
        Err(_) => {
            // Another file system (or a link refused there): copy, verify, then remove.
            let bytes =
                fs::read(from).map_err(|e| format!("cannot read {}: {e}", from.display()))?;
            let mut out = fs::OpenOptions::new();
            out.write(true).create_new(true);
            #[cfg(unix)]
            {
                use std::os::unix::fs::OpenOptionsExt;
                out.mode(0o600);
            }
            let mut file = match out.open(to) {
                Ok(f) => f,
                Err(e) if e.kind() == std::io::ErrorKind::AlreadyExists => return Ok(false),
                Err(e) => return Err(format!("cannot create {}: {e}", to.display())),
            };
            std::io::Write::write_all(&mut file, &bytes)
                .map_err(|e| format!("cannot write {}: {e}", to.display()))?;
            drop(file);
            if fs::read(to).map_err(|e| format!("cannot read {}: {e}", to.display()))? != bytes {
                let _ = fs::remove_file(to);
                return Err(format!(
                    "{} differs after the copy; the source stays",
                    to.display()
                ));
            }
        }
    }
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        let mode = fs::metadata(to)
            .map_err(|e| format!("cannot read {}: {e}", to.display()))?
            .permissions()
            .mode();
        fs::set_permissions(to, fs::Permissions::from_mode(mode & 0o700))
            .map_err(|e| format!("cannot set the permissions of {}: {e}", to.display()))?;
    }
    fs::remove_file(from).map_err(|e| format!("cannot remove {}: {e}", from.display()))?;
    Ok(true)
}

/// Apply an approved plan: the exact plan, its digest and a permission reference. The plan is
/// made again and must match, so a file added, moved or tracked since the preview makes it stale.
pub fn apply(
    root: &Path,
    home: &Path,
    approved: MovePlan,
    confirmed: &str,
    permission: &str,
) -> Result<serde_json::Value> {
    if !contract::permission_reference_ok(permission) {
        return Err("permission reference required".into());
    }
    let current = plan(root, home)?;
    if approved.protocol_version != 1
        || confirmed != approved.plan_digest
        || approved.digest() != approved.plan_digest
        || current.plan_digest != approved.plan_digest
    {
        return Err("stale or altered plan; re-preview required".into());
    }
    if approved.state == "blocked" {
        return Err(format!("blocked: {}; nothing moved", approved.reason));
    }
    let root =
        fs::canonicalize(root).map_err(|_| contract::PROJECT_PATH_UNAVAILABLE.to_string())?;
    let destination = PathBuf::from(&approved.destination);
    let mut moved: Vec<serde_json::Value> = Vec::new();
    let mut skipped = approved.skipped_existing.len() as u64;
    for item in &approved.moves {
        let to = destination.join(&item.to);
        if let Some(parent) = to.parent() {
            create_private_dirs(parent).map_err(|e| partial(&moved, &item.from, e))?;
        }
        match move_one(&root.join(&item.from), &to) {
            Ok(true) => {
                moved.push(serde_json::json!({"from": item.from, "to": to.display().to_string()}))
            }
            Ok(false) => skipped += 1,
            Err(e) => return Err(partial(&moved, &item.from, e)),
        }
    }
    Ok(serde_json::json!({
        "destination": approved.destination,
        "moved": moved.len(),
        "skipped_existing": skipped,
        "tracked_left_alone": approved.tracked,
        "unmovable": approved.unmovable,
        "items": moved,
    }))
}

fn partial(moved: &[serde_json::Value], at: &str, error: String) -> String {
    format!(
        "partial apply: {} file(s) moved; failed at {at}: {error}; run the preview again to continue",
        moved.len()
    )
}
