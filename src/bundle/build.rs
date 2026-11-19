//! `moai bundle build` (card C-119, SPEC-20261119-bundle-build): one new branch
//! `m-flow/bundle-<date>-<n>` and one new worktree `<HOME>/.moai/worktrees/<project>-bundle-<n>`
//! at the tip, one merge commit per head in the order given, a head whose merge conflicts
//! dropped and named. It writes a branch, a worktree and merge commits and nothing else (REQ-10):
//! no push, no fetch, no pull request, no card note, no ledger row. An existing branch, ref,
//! worktree or folder is never moved, reset, forced or deleted (REQ-6).

use super::{
    card_notes, newest_marker, parse_head_args, tip_json, tip_of, Git, HeadArg, Out, Ready, Result,
    Tip, TEXT_DRIVER,
};
use crate::{home_workdir, hygiene, worktree};
use serde_json::{json, Value};
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

/// `git` for the writes: the user's own author identity and hook settings apply, so the
/// environment is scrubbed of repository redirects only (`identity::scrub_git_env`), not of the
/// global configuration that `audit_delta::command` turns off (D-4).
struct Writer<'a> {
    dir: &'a Path,
}

impl Writer<'_> {
    fn run(&self, args: &[&str]) -> Result<Out> {
        let mut cmd = Command::new("git");
        crate::identity::scrub_git_env(&mut cmd);
        let o = cmd
            .args(args)
            .current_dir(self.dir)
            .env("GIT_TERMINAL_PROMPT", "0")
            .stdin(Stdio::null())
            .output()
            .map_err(|e| format!("cannot execute git: {e}"))?;
        Ok(Out {
            code: o.status.code(),
            out: String::from_utf8_lossy(&o.stdout).into_owned(),
            err: first_line(&String::from_utf8_lossy(&o.stderr)),
        })
    }
}

fn first_line(s: &str) -> String {
    s.lines()
        .map(str::trim)
        .find(|l| !l.is_empty())
        .unwrap_or("no message")
        .to_string()
}

fn short(sha: &str) -> &str {
    sha.get(..10).unwrap_or(sha)
}

/// The date part of the branch name: `--date` when given, else today (UTC).
fn date_part(date: Option<&str>) -> Result<String> {
    match date {
        Some(d) if d.len() == 8 && d.bytes().all(|b| b.is_ascii_digit()) => Ok(d.to_string()),
        Some(d) => Err(format!(
            "bundle build refused: --date `{d}` is eight digits, YYYYMMDD"
        )),
        None => Ok(
            hygiene::utc_date(crate::statusline::counts::now_unix().max(0) as u64).replace('-', ""),
        ),
    }
}

/// A head of the request: the card, its commit, and what the tool does with it.
struct Wanted {
    card: String,
    head: String,
    in_tip: bool,
}

/// REQ-2: every head a build cannot use, in one answer, before anything is created.
fn resolve(
    git: &Git,
    project: &Path,
    home: &Path,
    tip: &Tip,
    args: &[HeadArg],
) -> Result<Vec<Wanted>> {
    let mut problems: Vec<String> = Vec::new();
    let mut out: Vec<Wanted> = Vec::new();
    for a in args {
        let head = match &a.sha {
            Some(s) => s.clone(),
            None => match card_notes(project, home, &a.card) {
                Err(e) => {
                    problems.push(format!("{}: {e}", a.card));
                    continue;
                }
                Ok(notes) => match newest_marker(&a.card, &notes) {
                    Some(Ready { head: Some(h), .. }) => h,
                    _ => {
                        problems.push(format!(
                            "{}: no readable LANE-READY-INTEGRATION line",
                            a.card
                        ));
                        continue;
                    }
                },
            },
        };
        if git.commit(&head).is_none() {
            problems.push(format!(
                "{}: the head {} is not a commit here (fetch first)",
                a.card,
                short(&head)
            ));
            continue;
        }
        let in_tip = git.is_ancestor(&head, &tip.sha);
        out.push(Wanted {
            card: a.card.clone(),
            head,
            in_tip,
        });
    }
    if problems.is_empty() && out.iter().all(|w| w.in_tip) {
        problems.push("no head left to merge: every head is already in the tip".into());
    }
    if !problems.is_empty() {
        return Err(format!(
            "bundle build refused: {} problem(s), nothing was created: {}",
            problems.len(),
            problems.join("; ")
        ));
    }
    Ok(out)
}

fn ref_exists(git: &Git, name: &str) -> bool {
    git.raw(&["show-ref", "--verify", "--quiet", name])
        .is_ok_and(|o| o.code == Some(0))
}

/// REQ-5: the lowest `n` whose branch (local and `origin`), registered worktree and folder are
/// all unused. Returns `(n, branch, folder)`.
fn number(git: &Git, dir: &Path, home: &Path, date: &str) -> Result<(u32, String, PathBuf)> {
    let entries = worktree::worktrees(dir)?;
    let primary = entries
        .first()
        .map(|e| PathBuf::from(&e.path))
        .ok_or("bundle build refused: the repository has no primary worktree")?;
    let registered: Vec<PathBuf> = entries
        .iter()
        .map(|e| worktree::resolve_existing(Path::new(&e.path)))
        .collect();
    for n in 1..=999u32 {
        let branch = format!("m-flow/bundle-{date}-{n}");
        let folder = home_workdir::worktree_dir(home, &primary, &format!("bundle-{n}"))?;
        let taken = ref_exists(git, &format!("refs/heads/{branch}"))
            || ref_exists(git, &format!("refs/remotes/origin/{branch}"))
            || registered.contains(&worktree::resolve_existing(&folder))
            || std::fs::symlink_metadata(&folder).is_ok();
        if !taken {
            return Ok((n, branch, folder));
        }
    }
    Err("bundle build refused: the numbers 1 to 999 of this date are all used".into())
}

/// REQ-6, REQ-7: create the branch and its worktree at `base`; `-b` (never `-B`) makes an
/// existing branch an error, and an existing folder is refused before git runs. Nothing that
/// exists is changed when it fails.
pub fn create(dir: &Path, branch: &str, folder: &Path, base: &str) -> Result<()> {
    if std::fs::symlink_metadata(folder).is_ok() {
        return Err(format!(
            "bundle build refused: {} already exists, nothing was changed",
            folder.display()
        ));
    }
    if let Some(parent) = folder.parent() {
        std::fs::create_dir_all(parent).map_err(|e| {
            format!(
                "bundle build refused: cannot create {}: {e}",
                parent.display()
            )
        })?;
    }
    let folder_s = folder.to_string_lossy().into_owned();
    let o = Writer { dir }.run(&["worktree", "add", "--quiet", "-b", branch, &folder_s, base])?;
    if o.code != Some(0) {
        return Err(format!(
            "bundle build refused: git worktree add failed, nothing was changed: {}",
            o.err
        ));
    }
    Ok(())
}

fn left_behind(branch: &str, folder: &Path, why: &str) -> String {
    format!(
        "bundle build stopped: {why}; the branch `{branch}` and the worktree {} are left as they are, nothing was deleted",
        folder.display()
    )
}

enum Merged {
    Commit(String),
    Conflict(Vec<String>),
}

/// REQ-3, REQ-8, REQ-9: merge one head with one merge commit under the text merge GitHub runs; a
/// conflict is aborted and the branch checked to be where it was.
fn merge_one(folder: &Path, branch: &str, card: &str, head: &str, before: &str) -> Result<Merged> {
    let w = Writer { dir: folder };
    let read = Git { dir: folder };
    let message = format!("Merge {card} head {} into {branch}", short(head));
    let o = w.run(&["-c", TEXT_DRIVER, "merge", "--no-ff", "-m", &message, head])?;
    match o.code {
        Some(0) => read
            .commit("HEAD")
            .map(Merged::Commit)
            .ok_or_else(|| left_behind(branch, folder, "HEAD cannot be read after a merge")),
        code => {
            let files = read
                .lines(&["diff", "--name-only", "--diff-filter=U"])
                .unwrap_or_default();
            // A failed merge may leave the index half merged (a conflict, or a hook or an
            // author identity refusing the commit): abort it, then check the branch.
            let abort = w.run(&["merge", "--abort"])?;
            let restored = read.commit("HEAD").as_deref() == Some(before);
            let conflict = code == Some(1) && !files.is_empty();
            if conflict && abort.code == Some(0) && restored {
                return Ok(Merged::Conflict(files));
            }
            let why = if !restored {
                format!(
                    "git merge of {card} head {} failed ({}) and the branch is no longer at {}",
                    short(head),
                    o.err,
                    short(before)
                )
            } else if conflict {
                format!(
                    "the merge of {card} head {} conflicts and `git merge --abort` failed: {}",
                    short(head),
                    abort.err
                )
            } else {
                format!("git merge of {card} head {} failed: {}", short(head), o.err)
            };
            Err(left_behind(branch, folder, &why))
        }
    }
}

/// `bundle build`: the answer is the branch, the worktree, the pinned SHA and what happened to
/// each head.
pub fn build(
    project: &Path,
    home: &Path,
    dir: &Path,
    tip: &str,
    heads: &[String],
    date: Option<&str>,
) -> Result<Value> {
    let args = parse_head_args(heads)?;
    let date = date_part(date)?;
    let git = Git { dir };
    let tip = tip_of(&git, tip)?;
    let wanted = resolve(&git, project, home, &tip, &args)?;
    let (n, branch, folder) = number(&git, dir, home, &date)?;
    create(dir, &branch, &folder, &tip.sha)?;

    let wt = Git { dir: &folder };
    let mut merged: Vec<Value> = Vec::new();
    let mut dropped: Vec<Value> = Vec::new();
    let mut skipped: Vec<Value> = Vec::new();
    for w in &wanted {
        if w.in_tip {
            skipped.push(json!({"card": w.card, "head": w.head, "reason": "in_tip"}));
            continue;
        }
        let before = wt
            .commit("HEAD")
            .ok_or_else(|| left_behind(&branch, &folder, "HEAD of the worktree cannot be read"))?;
        if wt.is_ancestor(&w.head, &before) {
            skipped.push(json!({"card": w.card, "head": w.head, "reason": "in_branch"}));
            continue;
        }
        match merge_one(&folder, &branch, &w.card, &w.head, &before)? {
            Merged::Commit(c) => {
                merged.push(json!({"card": w.card, "head": w.head, "merge": c}));
            }
            Merged::Conflict(files) => {
                dropped.push(json!({"card": w.card, "head": w.head, "files": files}));
            }
        }
    }
    // D-9: the pinned SHA is what git says for the branch now.
    let pinned = wt
        .commit(&format!("refs/heads/{branch}"))
        .ok_or_else(|| left_behind(&branch, &folder, "the branch cannot be read back"))?;
    let rebuild = if dropped.is_empty() {
        String::new()
    } else {
        let kept: Vec<String> = wanted
            .iter()
            .filter(|w| !dropped.iter().any(|d| d["card"] == w.card.as_str()))
            .map(|w| format!("{}={}", w.card, w.head))
            .collect();
        format!("moai bundle build {} --tip {}", kept.join(" "), tip.name)
    };
    Ok(json!({
        "action": "build", "number": n, "branch": branch,
        "worktree": folder.to_string_lossy(), "tip": tip_json(&tip), "pinned": pinned,
        "merged": merged, "dropped": dropped, "skipped": skipped,
        "rebuild": rebuild, "ok": dropped.is_empty(), "writes": true,
    }))
}
