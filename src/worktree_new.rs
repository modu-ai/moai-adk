//! A lane's worktree in one command (SPEC-20261124-lane-worktree-command).
//!
//! [`make`] adds `<primary checkout>/.claude/worktrees/<name>` on a new branch cut from the
//! integration branch (or a named base), then reads `git worktree list --porcelain` again and
//! answers the absolute path only when the list holds that path on the new branch at the base
//! commit and the folder exists. The primary checkout is read from the git common directory, so
//! a run from inside a linked worktree still makes the folder in the primary's place and never
//! below the folder it was run in; the name is one plain folder name, never a path.
//!
//! The command only makes and reads back. There is no option that forces, deletes, prunes or
//! moves a worktree, and a read-back that does not match leaves the folder where it is and says
//! so: removing it is for a person or for `worktree-gc-*`.

use crate::worktree_slot::{common_dir, linked, primary_of, same, SLOT_DIR};
use serde_json::{json, Value};
use std::fs;
use std::path::Path;
use std::process::{Command, Stdio};

type Result<T> = std::result::Result<T, String>;

/// The longest worktree name.
pub const MAX_NAME: usize = 64;

/// Why `name` is not one plain folder name, or `None` when it is: one to [`MAX_NAME`] characters
/// of lower-case letters, digits and `-`, not starting with `-`. A separator, a `.` or `..` part,
/// a drive or an absolute path all hold a character that is not allowed, so a relative path or a
/// nested place cannot be spelled.
pub fn name_problem(name: &str) -> Option<String> {
    if name.is_empty() {
        return Some("the name is empty".into());
    }
    if name.chars().count() > MAX_NAME {
        return Some(format!("the name is longer than {MAX_NAME} characters"));
    }
    if name.starts_with('-') {
        return Some("the name starts with `-`".into());
    }
    if let Some(c) = name
        .chars()
        .find(|c| !(c.is_ascii_lowercase() || c.is_ascii_digit() || *c == '-'))
    {
        return Some(format!(
            "the name holds {c:?}; only lower-case letters, digits and `-` are allowed, so a path, a `.` or `..` part and a nested place cannot be spelled"
        ));
    }
    None
}

/// One git call in `dir` with the inherited git variables removed: the exit success, stdout and
/// the line of stderr that says why it failed.
fn git(dir: &Path, args: &[&str]) -> (bool, String, String) {
    let mut cmd = Command::new("git");
    cmd.args(args)
        .current_dir(dir)
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());
    crate::identity::scrub_git_env(&mut cmd);
    match cmd.output() {
        Ok(out) => {
            let err = String::from_utf8_lossy(&out.stderr);
            let why = err
                .lines()
                .map(str::trim)
                .find(|l| l.starts_with("fatal:") || l.starts_with("error:"))
                .or_else(|| err.lines().map(str::trim).rfind(|l| !l.is_empty()))
                .unwrap_or("")
                .to_string();
            (
                out.status.success(),
                String::from_utf8_lossy(&out.stdout).trim().to_string(),
                why,
            )
        }
        Err(e) => (false, String::new(), format!("git did not start: {e}")),
    }
}

/// The commit `rev` names in `dir`, or `None`.
fn commit_of(dir: &Path, rev: &str) -> Option<String> {
    let spec = format!("{rev}^{{commit}}");
    let (ok, out, _) = git(
        dir,
        &[
            "rev-parse",
            "--verify",
            "--quiet",
            "--end-of-options",
            &spec,
        ],
    );
    (ok && crate::gate::is_full_sha(&out)).then_some(out)
}

/// Whether `branch` is a valid branch name that git would not rewrite.
fn branch_name_ok(dir: &Path, branch: &str) -> bool {
    // `refs/...` would make `refs/heads/refs/...`, a branch nobody means.
    if branch.is_empty() || branch.starts_with('-') || branch.starts_with("refs/") {
        return false;
    }
    let (ok, out, _) = git(dir, &["check-ref-format", "--branch", branch]);
    ok && out == branch
}

/// Makes the worktree `name` on the new branch `branch` and reads it back. `base` is the commit
/// the branch starts at; without it the integration branch's tip is used. Nothing is made when
/// any check before the add refuses.
pub fn make(project: &Path, name: &str, branch: &str, base: Option<&str>) -> Result<Value> {
    if let Some(problem) = name_problem(name) {
        return Err(format!(
            "worktree name {name:?} refused: {problem}; nothing was made"
        ));
    }
    let common = common_dir(project)?;
    let primary = primary_of(&common)?;
    let target = primary.join(SLOT_DIR).join(name);
    crate::home_workdir::place_verdict(&primary, &target).map_err(|r| r.detail)?;
    if fs::symlink_metadata(&target).is_ok() {
        return Err(format!(
            "the folder {} already exists; nothing was made",
            target.display()
        ));
    }
    if !branch_name_ok(&primary, branch) {
        return Err(format!(
            "{branch:?} is not a usable branch name; nothing was made"
        ));
    }
    if commit_of(&primary, &format!("refs/heads/{branch}")).is_some() {
        return Err(format!(
            "the branch {branch} already exists; nothing was made"
        ));
    }
    let (sha, from) = match base {
        Some(b) => (
            commit_of(&primary, b).ok_or_else(|| {
                format!("the base {b:?} names no commit of this repository; nothing was made")
            })?,
            format!("--base {b}"),
        ),
        None => {
            let integration = crate::card::landed::integration_branch(project).map_err(|e| {
                format!("no integration branch is known ({e}); give --base <ref>; nothing was made")
            })?;
            let tip = crate::card::landed::tip_of(project, &integration)
                .map_err(|e| format!("{e}; nothing was made"))?;
            (tip, format!("integration branch {integration}"))
        }
    };
    crate::home_workdir::ensure_place_ignored(&primary)
        .map_err(|e| format!("{e}; nothing was made"))?;
    let path_arg = target.to_string_lossy().to_string();
    let (added, _, why) = git(
        &primary,
        &["worktree", "add", "-b", branch, &path_arg, &sha],
    );
    if !added {
        return Err(format!("git worktree add failed: {why}; nothing was made"));
    }
    read_back(&primary, &target, branch, &sha)?;
    Ok(json!({
        "path": path_arg,
        "name": name,
        "branch": branch,
        "base": sha,
        "head": sha,
        "from": from,
        "primary": primary.to_string_lossy(),
    }))
}

/// The second read: `git worktree list --porcelain` must hold `target` on `refs/heads/<branch>`
/// at `sha`, and the folder must exist. A mismatch names the path it leaves in place.
fn read_back(primary: &Path, target: &Path, branch: &str, sha: &str) -> Result<()> {
    let listed = linked(primary)?;
    let entry = listed.iter().find(|e| same(&e.path, target));
    // The parser gives a branch without its `refs/heads/` prefix.
    let holds = entry
        .is_some_and(|e| e.branch.as_deref() == Some(branch) && e.head.as_deref() == Some(sha));
    if holds && target.is_dir() {
        return Ok(());
    }
    Err(format!(
        "the worktree at {} was added but `git worktree list` does not hold it on {branch} at {sha} \
         (or its folder is missing); it was left in place and nothing was deleted",
        target.display()
    ))
}

#[cfg(test)]
mod tests {
    use super::{name_problem, MAX_NAME};

    #[test]
    fn plain_names_pass_and_path_shaped_names_do_not() {
        for ok in ["lane-a", "a", "x1-y2", &"a".repeat(MAX_NAME)] {
            assert_eq!(name_problem(ok), None, "{ok}");
        }
        for bad in [
            "",
            "a/b",
            "..",
            ".",
            "A",
            "-a",
            "a b",
            "a.b",
            "/abs",
            &"a".repeat(MAX_NAME + 1),
        ] {
            assert!(name_problem(bad).is_some(), "{bad:?}");
        }
    }
}
