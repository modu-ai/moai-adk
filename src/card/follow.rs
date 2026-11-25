//! The worktree that has the integration branch checked out follows a landing
//! (SPEC-20261125-land-checked-out-branch).
//!
//! `card land` moves the branch by one `git update-ref` compare-and-swap. A branch that a worktree
//! has checked out would leave that worktree's index and files at the old commit, so the landing
//! does two more things around the move, and nothing else:
//!
//! - [`plan`], before the first write: the worktree is found and every precondition is read (the
//!   state it is in, a tracked change, a path the new commit adds that something stands on, a
//!   submodule entry, and a dry run of the same two-tree merge git itself would make). A refusal
//!   carries its reason and the next thing to do; the landing adds `nothing was changed`.
//! - [`apply`], right after the move: `git read-tree -m -u <old tip> <new head>`, git's own
//!   carry-forward of `git switch`, then a check that the index equals the new commit.
//!
//! A call that stopped between the move and the follow is finished by the same call run again:
//! [`rerun`] reads the tip this card's start row recorded and, when the worktree's index is still
//! that old commit's tree, runs the same follow from it. Nothing is deleted or overwritten by
//! hand, and `git reset --hard` is never used (its `-u` form loses work).
//!
//! One window stays: git has no switch that makes `read-tree` refuse an ignored file at an added
//! path (it overwrites it), so [`apply`] reads the added paths again right before it runs
//! `read-tree`; a file that appears in the few milliseconds between that read and git's own scan
//! is not caught (SPEC-20261125-land-checked-out-branch, Risks).

use super::landed;
use rusqlite::{Connection, OptionalExtension};
use serde_json::{json, Value};
use std::collections::BTreeSet;
use std::path::{Path, PathBuf};

type Result<T> = std::result::Result<T, String>;

/// How many paths a refusal lists before it counts the rest.
const LISTED: usize = 5;

/// A worktree to bring along and the two commits it goes between.
#[derive(Debug, Clone)]
pub struct Plan {
    pub tree: PathBuf,
    pub from: String,
    pub to: String,
}

/// What a landing did to the worktree, as the answer's `worktree_follow` key.
pub fn answer(
    tree: Option<&Path>,
    status: &str,
    from: &str,
    to: &str,
    note: Option<&str>,
) -> Value {
    json!({
        "worktree": tree.map(|t| t.display().to_string()),
        "status": status, "from": from, "to": to, "note": note,
    })
}

/// What a refusal of the worktree step after the branch moved says: the reason, and that the
/// branch already moved and the same call run again finishes the step.
pub fn stopped(why: &str, branch: &str, head: &str) -> String {
    format!(
        "{why}; the integration branch {branch} already moved to {head} and the card is still \
         open: run the same call again after the cause is repaired"
    )
}

/// The answer of a landing with no worktree to follow.
pub fn none() -> Value {
    answer(None, "none", "", "", None)
}

/// A refusal: the reason and the next thing to do.
fn refusal(reason: impl std::fmt::Display, next: &str) -> String {
    format!("{reason}; {next}")
}

/// `git` in `tree`: the stdout on success, git's message on failure.
fn run(tree: &Path, args: &[&str]) -> Result<Vec<u8>> {
    let out = landed::git(tree, args)?;
    if out.status.success() {
        Ok(out.stdout)
    } else {
        Err(String::from_utf8_lossy(&out.stderr).trim().to_string())
    }
}

fn text(tree: &Path, args: &[&str]) -> Result<String> {
    Ok(String::from_utf8_lossy(&run(tree, args)?)
        .trim()
        .to_string())
}

/// Whether a git command that answers by its exit code (`--quiet`) exits 0.
fn quiet(tree: &Path, args: &[&str]) -> Result<bool> {
    Ok(landed::git(tree, args)?.status.success())
}

/// `a, b, c and 2 more` for the paths of a refusal.
fn listed(paths: &[String]) -> String {
    let mut shown: Vec<&str> = paths.iter().take(LISTED).map(String::as_str).collect();
    let more = paths.len().saturating_sub(LISTED);
    let tail = if more > 0 {
        format!(" and {more} more")
    } else {
        String::new()
    };
    shown.sort_unstable();
    format!("{}{tail}", shown.join(", "))
}

/// The worktree that has `full_ref` checked out: none, one, or a refusal for several.
fn holder(project: &Path, full_ref: &str, branch: &str) -> Result<Option<PathBuf>> {
    let holders = landed::checked_out_at(project, full_ref)?;
    match holders.as_slice() {
        [] => Ok(None),
        [one] => Ok(Some(PathBuf::from(one))),
        many => Err(refusal(
            format!(
                "the integration branch {branch} is checked out in more than one worktree ({})",
                many.join(", ")
            ),
            "leave it checked out in one of them (switch the others to another branch) and run the call again",
        )),
    }
}

/// The marks of a merge, a rebase, a cherry-pick, a revert, a bisect or a lock in a git folder.
const PENDING: [&str; 9] = [
    "MERGE_HEAD",
    "CHERRY_PICK_HEAD",
    "REVERT_HEAD",
    "REBASE_HEAD",
    "rebase-merge",
    "rebase-apply",
    "sequencer",
    "BISECT_LOG",
    "index.lock",
];

/// The refusals about the state the worktree is in: its HEAD, an operation in progress, a sparse
/// checkout and entries git is told not to look at.
fn check_state(tree: &Path, tip: &str, branch: &str) -> Result<()> {
    let at = tree.display();
    let head = text(tree, &["rev-parse", "HEAD"]).map_err(|why| {
        refusal(
            format!("HEAD of {at} cannot be read ({why})"),
            "repair that worktree and run the call again",
        )
    })?;
    if head != tip {
        return Err(refusal(
            format!("HEAD of {at} is {head}, not the tip {tip} of {branch} the landing read"),
            "read the tip again and run the call again",
        ));
    }
    let git_dir = text(tree, &["rev-parse", "--absolute-git-dir"]).map_err(|why| {
        refusal(
            format!("the git folder of {at} cannot be read ({why})"),
            "repair that worktree and run the call again",
        )
    })?;
    let pending: Vec<&str> = PENDING
        .iter()
        .copied()
        .filter(|name| Path::new(&git_dir).join(name).exists())
        .collect();
    if !pending.is_empty() {
        return Err(refusal(
            format!("{at} has an operation in progress ({})", pending.join(", ")),
            "finish or abort it there (merge, rebase, cherry-pick, revert or bisect --abort/--reset, or remove a stale index.lock), then run the call again",
        ));
    }
    if text(tree, &["config", "--get", "core.sparseCheckout"]).is_ok_and(|v| v == "true") {
        return Err(refusal(
            format!("{at} uses a sparse checkout"),
            "the follow step does not support it: run `git sparse-checkout disable` there and run the call again",
        ));
    }
    check_entry_tags(tree)
}

/// The refusals about index entries `git ls-files -v` tags other than `H`: an unmerged entry (`M`),
/// a skip-worktree entry (`S`), an assume-unchanged entry (a lower-case tag) and any other state.
fn check_entry_tags(tree: &Path) -> Result<()> {
    let at = tree.display();
    let tagged = run(tree, &["ls-files", "-v", "-z"])?;
    let (mut unmerged, mut skipped, mut assumed, mut other) = (vec![], vec![], vec![], vec![]);
    for entry in tagged.split(|b| *b == 0).filter(|e| e.len() > 2) {
        let path = String::from_utf8_lossy(&entry[2..]).into_owned();
        match entry[0] {
            b'H' => {}
            b'M' | b'm' => unmerged.push(path),
            b'S' | b's' => skipped.push(path),
            tag if tag.is_ascii_lowercase() => assumed.push(path),
            tag => other.push(format!("{}: {path}", tag as char)),
        }
    }
    // An unmerged path has one entry per stage; name it once.
    unmerged.sort_unstable();
    unmerged.dedup();
    if !unmerged.is_empty() {
        return Err(refusal(
            format!("{at} has unmerged paths: {}", listed(&unmerged)),
            "finish or abort that merge there (resolve and `git add` them, or `git merge --abort`), then run the call again",
        ));
    }
    if !skipped.is_empty() {
        return Err(refusal(
            format!("{at} has skip-worktree entries: {}", listed(&skipped)),
            "clear them with `git update-index --no-skip-worktree <path>` and run the call again",
        ));
    }
    if !assumed.is_empty() {
        return Err(refusal(
            format!("{at} has assume-unchanged entries: {}", listed(&assumed)),
            "clear them with `git update-index --no-assume-unchanged <path>` and run the call again",
        ));
    }
    if !other.is_empty() {
        return Err(refusal(
            format!(
                "{at} has index entries in a state this step does not know: {}",
                listed(&other)
            ),
            "bring the index to a plain state (see `git status` there) and run the call again",
        ));
    }
    Ok(())
}

/// The refusal for a tracked change in the worktree: a staged change, an unstaged change or a
/// deleted tracked file.
fn check_clean(tree: &Path) -> Result<()> {
    let out = run(
        tree,
        &[
            "status",
            "--porcelain=v1",
            "-z",
            "--untracked-files=no",
            "--no-renames",
        ],
    )?;
    let changed: Vec<String> = out
        .split(|b| *b == 0)
        .filter(|entry| entry.len() > 3)
        .map(|entry| String::from_utf8_lossy(&entry[3..]).into_owned())
        .collect();
    if changed.is_empty() {
        return Ok(());
    }
    Err(refusal(
        format!(
            "{} has {} tracked change(s): {}",
            tree.display(),
            changed.len(),
            listed(&changed)
        ),
        "commit, stash or discard them there (the landing never touches work it does not own), then run the call again",
    ))
}

/// One path a move changes.
struct Change {
    old_mode: String,
    new_mode: String,
    status: char,
    path: String,
}

/// The paths `from..to` changes, without rename detection.
fn changes(tree: &Path, from: &str, to: &str) -> Result<Vec<Change>> {
    let out = run(
        tree,
        &[
            "diff-tree",
            "-r",
            "--raw",
            "-z",
            "--no-renames",
            "--no-abbrev",
            from,
            to,
        ],
    )?;
    let mut fields = out.split(|b| *b == 0).filter(|f| !f.is_empty());
    let mut found = Vec::new();
    while let Some(meta) = fields.next() {
        let path = fields
            .next()
            .ok_or_else(|| "git diff-tree gave a record with no path".to_string())?;
        let meta = String::from_utf8_lossy(meta).into_owned();
        let parts: Vec<&str> = meta.trim_start_matches(':').split(' ').collect();
        let [old_mode, new_mode, _old, _new, status] = parts.as_slice() else {
            return Err(format!(
                "git diff-tree gave a record this command cannot read: {meta}"
            ));
        };
        let path = String::from_utf8_lossy(path).into_owned();
        if path.contains('\u{fffd}') {
            return Err(refusal(
                format!("the move changes a path that is not valid UTF-8 ({path})"),
                "the follow step does not support it: rename the path in the head and land again",
            ));
        }
        found.push(Change {
            old_mode: old_mode.to_string(),
            new_mode: new_mode.to_string(),
            status: status.chars().next().unwrap_or('?'),
            path,
        });
    }
    Ok(found)
}

/// Every path the commit `from` tracks.
fn tracked_paths(tree: &Path, from: &str) -> Result<Vec<String>> {
    let out = run(tree, &["ls-tree", "-r", "--name-only", "-z", from])?;
    Ok(out
        .split(|b| *b == 0)
        .filter(|p| !p.is_empty())
        .map(|p| String::from_utf8_lossy(p).into_owned())
        .collect())
}

/// Whether `a` and `b` are one file under two spellings (a case-only rename on a file system that
/// ignores case): names equal but for ASCII case, and on Unix the same device and inode.
fn same_file(a: &Path, b: &Path) -> bool {
    if !a
        .to_string_lossy()
        .eq_ignore_ascii_case(&b.to_string_lossy())
    {
        return false;
    }
    #[cfg(unix)]
    {
        use std::os::unix::fs::MetadataExt;
        match (std::fs::symlink_metadata(a), std::fs::symlink_metadata(b)) {
            (Ok(x), Ok(y)) => x.dev() == y.dev() && x.ino() == y.ino(),
            _ => false,
        }
    }
    #[cfg(not(unix))]
    {
        a.exists() && b.exists()
    }
}

/// Whether what stands at `path` is a tracked file this same move deletes: the path itself, or a
/// file the move deletes under another spelling of the same name.
fn deleted_by_the_move(tree: &Path, path: &str, removed: &BTreeSet<&str>) -> bool {
    removed.contains(path)
        || removed
            .iter()
            .any(|r| same_file(&tree.join(path), &tree.join(r)))
}

/// What stands where the move adds `path`: `None` when nothing does or only what the move itself
/// deletes (a tracked folder the move turns into a file, a tracked file renamed by case); else a
/// description. The file system is asked, so an ignored file counts too.
fn in_the_way(
    tree: &Path,
    path: &str,
    removed: &BTreeSet<&str>,
    tracked: &[String],
) -> Result<Option<String>> {
    if let Ok(meta) = std::fs::symlink_metadata(tree.join(path)) {
        if deleted_by_the_move(tree, path, removed) {
            return Ok(None);
        }
        if !meta.is_dir() {
            return Ok(Some(path.to_string()));
        }
        // A folder where the move adds a file: fine when it holds only tracked files the move
        // deletes (and nothing else), or nothing at all.
        let inside = format!("{path}/");
        let mine: Vec<&String> = tracked.iter().filter(|t| t.starts_with(&inside)).collect();
        let empty = std::fs::read_dir(tree.join(path)).is_ok_and(|mut d| d.next().is_none());
        if (mine.is_empty() && !empty) || mine.iter().any(|t| !removed.contains(t.as_str())) {
            return Ok(Some(format!(
                "{path}/ (a folder that holds files the move keeps)"
            )));
        }
        let mut strangers: Vec<String> = Vec::new();
        for flags in [&["--others"][..], &["--others", "--ignored"][..]] {
            let mut args = vec![
                "--literal-pathspecs",
                "ls-files",
                "-z",
                "--exclude-standard",
            ];
            args.extend_from_slice(flags);
            args.extend_from_slice(&["--", path]);
            let out = run(tree, &args)?;
            strangers.extend(
                out.split(|b| *b == 0)
                    .filter(|p| !p.is_empty())
                    .map(|p| String::from_utf8_lossy(p).into_owned()),
            );
        }
        if !strangers.is_empty() {
            return Ok(Some(format!(
                "{path}/ (a folder that holds files no commit tracks: {})",
                listed(&strangers)
            )));
        }
        return Ok(None);
    }
    let mut prefix = String::new();
    for part in path.split('/').take(path.split('/').count() - 1) {
        if !prefix.is_empty() {
            prefix.push('/');
        }
        prefix.push_str(part);
        let blocked = std::fs::symlink_metadata(tree.join(&prefix))
            .is_ok_and(|m| !m.file_type().is_dir())
            && !deleted_by_the_move(tree, &prefix, removed);
        if blocked {
            return Ok(Some(format!(
                "{prefix} (a file where the new commit needs a folder, for {path})"
            )));
        }
    }
    Ok(None)
}

/// The refusals about what the move would write: a submodule entry, and a path the new commit adds
/// that something stands on or under. What the same move deletes is not an obstacle, and the
/// obstacle is looked for by the file system, so an ignored file counts too.
fn check_paths(tree: &Path, from: &str, to: &str) -> Result<()> {
    let all = changes(tree, from, to)?;
    let submodules: Vec<String> = all
        .iter()
        .filter(|c| c.old_mode == "160000" || c.new_mode == "160000")
        .map(|c| c.path.clone())
        .collect();
    if !submodules.is_empty() {
        return Err(refusal(
            format!(
                "the move changes a submodule entry ({})",
                listed(&submodules)
            ),
            "the follow step does not support it: update that worktree by hand and land without it, or land the submodule change separately",
        ));
    }
    let stands = stands_on_added_paths(tree, &all, from)?;
    if stands.is_empty() {
        return Ok(());
    }
    Err(refusal(
        format!(
            "{} holds {} file(s) that no commit tracks at a path the new commit adds: {}",
            tree.display(),
            stands.len(),
            listed(&stands)
        ),
        "move or delete them (or add them to the commit), then run the call again",
    ))
}

/// What stands where the move adds a path and is no part of what the move deletes, one entry per
/// added path that has something in the way.
fn stands_on_added_paths(tree: &Path, all: &[Change], from: &str) -> Result<Vec<String>> {
    let removed: BTreeSet<&str> = all
        .iter()
        .filter(|c| c.status == 'D')
        .map(|c| c.path.as_str())
        .collect();
    let tracked = tracked_paths(tree, from)?;
    let mut stands: Vec<String> = Vec::new();
    for added in all.iter().filter(|c| c.status == 'A') {
        if let Some(found) = in_the_way(tree, &added.path, &removed, &tracked)? {
            stands.push(found);
        }
    }
    Ok(stands)
}

/// The dry run of the two-tree merge the follow step makes: git's own refusal of a local file at
/// an added path or of a changed path that is not up to date.
fn dry_run(tree: &Path, from: &str, to: &str) -> Result<()> {
    run(tree, &["read-tree", "-m", "-u", "-n", from, to])
        .map(|_| ())
        .map_err(|why| {
            refusal(
                format!(
                    "git would refuse to bring {} from {from} to {to} ({})",
                    tree.display(),
                    why.lines().next().unwrap_or("no reason given")
                ),
                "repair that worktree (see git's reason) and run the call again",
            )
        })
}

/// Before the first write: the worktree that has the branch checked out (`None` when no worktree
/// has), every precondition, and the dry run. `Err` is the reason and the next step.
pub fn plan(
    project: &Path,
    full_ref: &str,
    branch: &str,
    tip: &str,
    head: &str,
) -> Result<Option<Plan>> {
    let Some(tree) = holder(project, full_ref, branch)? else {
        return Ok(None);
    };
    check_state(&tree, tip, branch)?;
    check_clean(&tree)?;
    check_paths(&tree, tip, head)?;
    dry_run(&tree, tip, head)?;
    Ok(Some(Plan {
        tree,
        from: tip.to_string(),
        to: head.to_string(),
    }))
}

/// After the move: `git read-tree -m -u <from> <to>`, then the index must equal `to`.
pub fn apply(plan: &Plan) -> Result<()> {
    let at = plan.tree.display();
    // `read-tree` overwrites an ignored file at an added path without a word, and the ledger
    // rows, the scans and the branch move came after the plan: read the added paths again now.
    check_paths(&plan.tree, &plan.from, &plan.to)?;
    run(&plan.tree, &["read-tree", "-m", "-u", &plan.from, &plan.to]).map_err(|why| {
        format!(
            "the worktree {at} could not follow the move from {} to {} ({why})",
            plan.from, plan.to
        )
    })?;
    if !quiet(&plan.tree, &["diff-index", "--cached", "--quiet", &plan.to])? {
        return Err(format!(
            "the worktree {at} was brought along but its index does not equal {}",
            plan.to
        ));
    }
    Ok(())
}

/// The tip this card's start row recorded for a landing of `head`.
fn started_from(conn: &Connection, card_id: i64, head: &str) -> Result<Option<String>> {
    conn.query_row(
        "SELECT json_extract(payload, '$.tip') FROM event \
         WHERE card_id = ?1 AND kind = 'card.land_started' \
           AND json_extract(payload, '$.head') = ?2 \
           AND json_type(payload, '$.tip') = 'text' \
         ORDER BY event_id DESC LIMIT 1",
        rusqlite::params![card_id, head],
        |r| r.get(0),
    )
    .optional()
    .map_err(|e| e.to_string())
}

/// What a rerun found: the worktree needs the follow, or it does not (with the reason).
pub enum Rerun {
    Needed(Plan),
    Nothing(Value),
}

/// A call that stopped between the move and the follow, run again: the branch already holds
/// `head`, so the worktree (when one has the branch checked out) is looked at against the tip this
/// card's start row recorded. An index that is still that old commit's tree is brought to the
/// tip's tree by the same follow; an index that is already the tip's tree needs nothing; an index
/// that is neither was changed by someone, and is left alone and said so.
pub fn rerun(
    project: &Path,
    conn: &Connection,
    card_id: i64,
    full_ref: &str,
    branch: &str,
    head: &str,
) -> Result<Rerun> {
    // A head the card's start row does not name was not moved by this card: nothing to finish.
    let Some(old) = started_from(conn, card_id, head)? else {
        return Ok(Rerun::Nothing(none()));
    };
    rerun_from(project, &old, full_ref, branch, head)
}

/// [`rerun`] with the old tip given: the same look at the worktree that has `full_ref` checked
/// out, for a caller that knows the tip the branch moved from and has no card to read it from
/// (`card land-sync`, SPEC-20261125-land-sync-merge D-3). `head` names the commit the branch moved
/// to; the tip is read from the branch, as `rerun` does, so `head` is not read again here.
pub fn rerun_from(
    project: &Path,
    old: &str,
    full_ref: &str,
    branch: &str,
    _head: &str,
) -> Result<Rerun> {
    let Some(tree) = holder(project, full_ref, branch)? else {
        return Ok(Rerun::Nothing(none()));
    };
    let tip = landed::tip_of(project, branch)?;
    if old == tip || quiet(&tree, &["diff-index", "--cached", "--quiet", &tip])? {
        return Ok(Rerun::Nothing(answer(Some(&tree), "none", old, &tip, None)));
    }
    if !quiet(&tree, &["diff-index", "--cached", "--quiet", old])? {
        // The branch moved and the landing is done, so the card closes; the worktree is left as
        // it is and the answer says how its owner brings it along (REQ-13).
        // `git reset --hard` also overwrites a file no commit tracks (ignored or not) that stands
        // where the new tip adds one, and the saved patch does not hold such a file: name them.
        let overwritten = match changes(&tree, old, &tip)
            .and_then(|all| stands_on_added_paths(&tree, &all, old))
        {
            Ok(stands) if stands.is_empty() => String::new(),
            Ok(stands) => format!(
                " Move these away before the reset, because the new commit adds a file at each of \
                 them and the reset overwrites what stands there, and the patch does not hold a \
                 file no commit tracks: {}.",
                listed(&stands)
            ),
            Err(_) => " The paths the new commit adds could not be read: before the reset, look \
                        for files of your own at them, because the reset overwrites those and the \
                        patch does not hold a file no commit tracks."
                .to_string(),
        };
        let note = format!(
            "the index of the worktree is neither the old tip's tree nor the new one's, so it was \
             left alone and shows the move backwards; `git diff {old}` there lists every change to \
             tracked files, staged or not, and `git reset --hard` discards all of them: save them \
             first (`git diff --binary {old} > keep.patch`; a file no commit tracks stays where it \
             is unless the new commit adds one at the same path), run `git reset --hard` there to \
             bring it to {tip}, and apply the patch again (`git apply --3way keep.patch`).{overwritten}"
        );
        return Ok(Rerun::Nothing(answer(
            Some(&tree),
            "none",
            old,
            &tip,
            Some(&note),
        )));
    }
    check_state(&tree, &tip, branch)?;
    // `git diff` compares content for an entry whose stat data differs (a file rewritten with the
    // same bytes); the plumbing `diff-files` would call it changed and never refreshes.
    if !quiet(&tree, &["diff", "--quiet", "--no-ext-diff"])? {
        return Err(refusal(
            format!(
                "{} has edited tracked files, and its index still lags at {old}",
                tree.display()
            ),
            "save a copy of the edited files elsewhere, discard the edits there (`git restore .`; a commit is refused on the integration branch and a stash would carry the lagging index too) and run the call again",
        ));
    }
    // The added paths are read by `apply` right before the write, so they are not read here.
    dry_run(&tree, old, &tip)?;
    Ok(Rerun::Needed(Plan {
        tree,
        from: old.to_string(),
        to: tip,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    const FULL: &str = "refs/heads/integ";
    const BRANCH: &str = "integ";

    fn git(dir: &Path, args: &[&str]) -> String {
        let mut cmd = std::process::Command::new("git");
        cmd.arg("-C").arg(dir).args([
            "-c",
            "user.name=fixture",
            "-c",
            "user.email=fixture@example.invalid",
            "-c",
            "commit.gpgsign=false",
        ]);
        cmd.args(args);
        crate::identity::scrub_git_env(&mut cmd);
        let out = cmd.output().unwrap();
        assert!(
            out.status.success(),
            "git {args:?}: {}",
            String::from_utf8_lossy(&out.stderr)
        );
        String::from_utf8_lossy(&out.stdout).trim().to_string()
    }

    /// A repository whose checked-out branch moved by `update-ref` only: the branch is at `tip`,
    /// the index and the files of its worktree are still at `old`.
    struct Lagging {
        dir: tempfile::TempDir,
        old: String,
        tip: String,
    }

    fn lagging() -> Lagging {
        let dir = tempfile::tempdir().unwrap();
        let at = dir.path();
        git(at, &["init", "-q", "-b", BRANCH]);
        std::fs::write(at.join("a.txt"), "one\n").unwrap();
        git(at, &["add", "a.txt"]);
        git(at, &["commit", "-q", "-m", "one"]);
        let old = git(at, &["rev-parse", "HEAD"]);
        git(at, &["checkout", "-q", "-b", "side"]);
        std::fs::write(at.join("b.txt"), "two\n").unwrap();
        git(at, &["add", "b.txt"]);
        git(at, &["commit", "-q", "-m", "two"]);
        let tip = git(at, &["rev-parse", "HEAD"]);
        git(at, &["checkout", "-q", BRANCH]);
        git(at, &["update-ref", FULL, &tip, &old]);
        Lagging { dir, old, tip }
    }

    /// A ledger that holds only the events `rerun` reads, with the start row of a landing of
    /// `head` that left the branch at `old`.
    fn ledger(start: Option<(&str, &str)>) -> Connection {
        let conn = Connection::open_in_memory().unwrap();
        conn.execute_batch(
            "CREATE TABLE event (event_id INTEGER PRIMARY KEY, card_id INTEGER, kind TEXT, \
             payload TEXT)",
        )
        .unwrap();
        if let Some((head, old)) = start {
            conn.execute(
                "INSERT INTO event (card_id, kind, payload) VALUES (7, 'card.land_started', ?1)",
                [json!({"head": head, "tip": old}).to_string()],
            )
            .unwrap();
        }
        conn
    }

    /// What a call found, as text to compare.
    fn shown(found: Result<Rerun>) -> String {
        match found {
            Ok(Rerun::Needed(p)) => format!("needed {} {} {}", p.tree.display(), p.from, p.to),
            Ok(Rerun::Nothing(v)) => format!("nothing {v}"),
            Err(why) => format!("refused {why}"),
        }
    }

    /// An explicit old tip follows a lagging checked-out worktree, as `rerun` does with the tip its
    /// start row recorded.
    #[test]
    fn rerun_from_follows_a_lagging_checked_out_worktree_like_rerun() {
        let f = lagging();
        let at = f.dir.path();
        let direct = shown(rerun_from(at, &f.old, FULL, BRANCH, &f.tip));
        assert!(direct.starts_with("needed "), "{direct}");
        assert!(
            direct.ends_with(&format!(" {} {}", f.old, f.tip)),
            "{direct}"
        );
        let conn = ledger(Some((&f.tip, &f.old)));
        let by_card = shown(rerun(at, &conn, 7, FULL, BRANCH, &f.tip));
        assert_eq!(direct, by_card);
        // The plan it gave brings the worktree along; the same call then finds nothing to do.
        let Ok(Rerun::Needed(plan)) = rerun_from(at, &f.old, FULL, BRANCH, &f.tip) else {
            panic!("the lagging worktree needs the follow");
        };
        apply(&plan).unwrap();
        let after = shown(rerun_from(at, &f.old, FULL, BRANCH, &f.tip));
        assert!(after.starts_with("nothing "), "{after}");
        assert!(after.contains("\"status\":\"none\""), "{after}");
        assert_eq!(after, shown(rerun(at, &conn, 7, FULL, BRANCH, &f.tip)));
    }

    /// A tracked file edited in the lagging worktree is refused, as `rerun` refuses it, and the
    /// file is left as it was.
    #[test]
    fn rerun_from_refuses_a_dirty_lagging_worktree_like_rerun() {
        let f = lagging();
        let at = f.dir.path();
        std::fs::write(at.join("a.txt"), "edited\n").unwrap();
        let direct = shown(rerun_from(at, &f.old, FULL, BRANCH, &f.tip));
        assert!(direct.starts_with("refused "), "{direct}");
        assert!(direct.contains("has edited tracked files"), "{direct}");
        assert!(direct.contains(&format!("lags at {}", f.old)), "{direct}");
        let conn = ledger(Some((&f.tip, &f.old)));
        assert_eq!(direct, shown(rerun(at, &conn, 7, FULL, BRANCH, &f.tip)));
        assert_eq!(
            std::fs::read_to_string(at.join("a.txt")).unwrap(),
            "edited\n"
        );
    }

    /// `rerun` still reads the old tip from the card's start row: a head no start row names finds
    /// nothing, while `rerun_from` with an explicit old tip needs no row.
    #[test]
    fn rerun_looks_the_old_tip_up_and_rerun_from_needs_no_row() {
        let f = lagging();
        let at = f.dir.path();
        let empty = ledger(None);
        assert_eq!(
            shown(rerun(at, &empty, 7, FULL, BRANCH, &f.tip)),
            format!("nothing {}", none())
        );
        let other = ledger(Some((&f.old, &f.old)));
        assert_eq!(
            shown(rerun(at, &other, 7, FULL, BRANCH, &f.tip)),
            format!("nothing {}", none())
        );
        let direct = shown(rerun_from(at, &f.old, FULL, BRANCH, &f.tip));
        assert!(direct.starts_with("needed "), "{direct}");
    }
}
