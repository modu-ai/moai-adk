//! `moai card repair`, and the absolute worktree path `moai card take` records
//! (SPEC-20261127-reservation-repair).
//!
//! A reservation records the `--worktree` and `--branch` it was taken with and, before this
//! module, had no way to correct them: a relative worktree path stopped `card submit`,
//! `card reclaim` and `card release` for good, and a mistyped branch name made the landing and
//! the cleanup look for the wrong branch. Now `card take` turns a relative `--worktree` into the
//! top of the checkout it names (resolved against the folder the command runs in), and the
//! holder of a live reservation corrects either value with one `card.repair` event.
//!
//! Only the holder at the current generation repairs: the fence is the one every holder call
//! uses, and a refusal writes nothing. Git is only read (`rev-parse`, `check-ref-format`); no
//! branch or worktree is created, moved or removed.

use super::command::{actor, open_for_write};
use super::dispatch::{card_id, checked_reason, reservation_json};
use super::input;
use super::land::check::git_text;
use crate::ledger::reservation::{self, RepairChange};
use serde_json::{json, Value};
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// The longest path a refusal prints, in characters.
const PATH_SHOWN_MAX: usize = 400;

/// A path as a refusal prints it: whole (the caller needs the full folder to act on it), with a
/// control character shown as `?` and a very long one cut.
fn show(path: &str) -> String {
    let mut out: String = path
        .chars()
        .take(PATH_SHOWN_MAX)
        .map(|c| if c.is_control() { '?' } else { c })
        .collect();
    if path.chars().count() > PATH_SHOWN_MAX {
        out.push('…');
    }
    out
}

/// How a folder is named in a refusal: as it was given, with the folder it resolved to when that
/// is a different text.
fn named(given: &str, resolved: &Path) -> String {
    let resolved = resolved.to_string_lossy();
    if given == resolved {
        format!("`{}`", show(given))
    } else {
        format!("`{}` (resolved to `{}`)", show(given), show(&resolved))
    }
}

/// The top of the git checkout that `folder` is, as git spells it. `given` is the text the user
/// typed and only names the folder in a refusal.
///
/// Refused, with a message that names the folder and the cause: the folder does not exist or is
/// not a folder, it is no git checkout at all, or it lies inside a checkout whose top is
/// another folder (the message names that top).
pub fn checkout_top(action: &str, given: &str, folder: &Path) -> Result<String> {
    let shown = named(given, folder);
    let Ok(canonical) = std::fs::canonicalize(folder) else {
        return Err(format!(
            "card {action} refused: --worktree {shown} is not an existing folder"
        ));
    };
    if !canonical.is_dir() {
        return Err(format!(
            "card {action} refused: --worktree {shown} is not an existing folder"
        ));
    }
    match crate::lane_registry::worktree_of(&canonical) {
        Some(top) if std::fs::canonicalize(&top).is_ok_and(|t| t == canonical) => Ok(top),
        Some(top) => Err(format!(
            "card {action} refused: --worktree {shown} is not the top of a git checkout; its top \
             is `{}`",
            show(&top)
        )),
        None => Err(format!(
            "card {action} refused: --worktree {shown} is not the top of a git checkout"
        )),
    }
}

/// What `card take` records for a given `--worktree` (SPEC-20261127-reservation-repair REQ-1): an
/// absolute path is recorded as it was given; a relative one is resolved against `cwd`, the
/// folder the command runs in, and recorded as the top of the checkout it names. Refused, with
/// nothing written, when `cwd` cannot be read or the resolved folder is not the top of a checkout.
pub fn absolute_worktree(given: &str, cwd: Option<&Path>) -> Result<String> {
    if given.trim().is_empty() || Path::new(given).is_absolute() {
        return Ok(given.to_string());
    }
    let Some(cwd) = cwd else {
        return Err(format!(
            "card take refused: --worktree `{}` is a relative path and the current folder cannot \
             be read; give an absolute path",
            show(given)
        ));
    };
    checkout_top("take", given, &cwd.join(given))
}

/// A short branch name git accepts: not empty, no leading `-`, not spelled as a ref (`refs/...`),
/// and `git check-ref-format` takes `refs/heads/<name>`.
fn check_branch(project: &Path, name: &str) -> Result<()> {
    let bad = |why: &str| {
        Err(format!(
            "card repair refused: --branch `{}` {why}",
            input::echo(name)
        ))
    };
    if name.trim().is_empty() {
        return bad("is empty");
    }
    if name != name.trim() || name.chars().any(char::is_control) {
        return bad("must be one line without a blank at either end");
    }
    if name.starts_with('-') {
        return bad("must not start with `-`");
    }
    if name.starts_with("refs/") {
        return bad("must be a short branch name, not a ref");
    }
    if git_text(
        project,
        &["check-ref-format", &format!("refs/heads/{name}")],
    )
    .is_err()
    {
        return bad("is not a branch name git accepts");
    }
    Ok(())
}

/// What `card repair` was given.
pub struct RepairRequest<'a> {
    pub holder: &'a str,
    pub generation: i64,
    pub branch: Option<&'a str>,
    pub worktree: Option<&'a str>,
    pub by: &'a str,
    pub reason: &'a str,
}

/// `card repair C-n --generation g [--branch B] [--worktree P] --by NAME --reason TEXT`: the
/// holder's correction of what its reservation recorded. The new worktree must be an absolute
/// path to the top of a checkout that exists (recorded as git spells that top), the new branch a
/// short name git accepts, the reason one line of 1 to 500 characters that holds no secret.
/// Everything is checked before the ledger is opened, and a refusal of the fence writes nothing.
pub fn repair(
    project: &Path,
    home: &Path,
    reference: &str,
    req: &RepairRequest<'_>,
) -> Result<Value> {
    let card_id = card_id("repair", reference)?;
    let by = actor(Some(req.by), "repair")?;
    let reason = checked_reason("repair", req.reason)?;
    if req.branch.is_none() && req.worktree.is_none() {
        return Err(
            "card repair refused: give --branch, --worktree or both; there is nothing to correct"
                .to_string(),
        );
    }
    if let Some(branch) = req.branch {
        check_branch(project, branch)?;
    }
    let top = match req.worktree {
        None => None,
        Some(given) => {
            if !Path::new(given).is_absolute() {
                return Err(format!(
                    "card repair refused: --worktree `{}` must be an absolute path to the top of \
                     a git checkout",
                    show(given)
                ));
            }
            Some(checkout_top("repair", given, Path::new(given))?)
        }
    };
    let (id, mut ledger) = open_for_write(project, home)?;
    let r = reservation::repair(
        &mut ledger,
        card_id,
        req.holder,
        req.generation,
        &RepairChange {
            branch: req.branch,
            worktree: top.as_deref(),
        },
        &by,
        &reason,
    )
    .map_err(|e| e.to_string())?;
    Ok(json!({
        "action": "repair", "project_id": id, "reservation": reservation_json(&r),
        "events": 1, "writes": true,
    }))
}
