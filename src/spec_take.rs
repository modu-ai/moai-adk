//! `moai spec take <SPEC-ID>` (SDD-7, SPEC-20261118-spec-take): start work on a SPEC. It checks
//! the SPEC, creates or switches to its branch `spec/<yyyymmdd>-<slug>`, links a card on request
//! (SDD-4) and prints the next commands. Pushing the branch and opening a draft pull request are
//! external writes: they run only with `--push` and `--draft-pr`.
//!
//! Every git call removes the caller's `GIT_*` redirects (`identity::scrub_git_env`) and keeps
//! the user's own git configuration. The base a branch was created from is kept in
//! the repository's own configuration (`branch.<branch>.moai-base`), so a later `take` on that
//! branch still names the right base for the pull request.

use crate::spec_doc::{self, IdForm};
use serde_json::{json, Value};
use std::path::Path;
use std::process::{Command, Stdio};

type Result<T> = std::result::Result<T, String>;

const REFUSED: &str = "spec take refused: ";

/// What `take` was asked to do.
pub struct Request<'a> {
    pub project: &'a Path,
    pub id: &'a str,
    /// The base branch or commit; `None` takes the current branch (or the recorded base).
    pub base: Option<&'a str>,
    pub remote: &'a str,
    pub card: Option<&'a str>,
    pub home: Option<&'a Path>,
    pub push: bool,
    pub draft_pr: bool,
    /// The `gh` program for `--draft-pr` (`MOAI_FORGE_GH_BIN` in tests, else `gh`).
    pub gh: &'a str,
}

fn refuse(why: impl std::fmt::Display) -> String {
    format!("{REFUSED}{why}")
}

/// Runs git in `dir`; stdout on success, the trimmed stderr as the error. The caller's `GIT_*`
/// redirects are removed; the user's own configuration stays, since a switch and a push must
/// behave as the user's git does (line endings, credential helpers, URL rewrites). git never
/// asks on the terminal: a push that needs a password fails instead of waiting.
fn git(dir: &Path, args: &[&str]) -> Result<String> {
    git_raw(dir, args).map(|s| s.trim().to_string())
}

/// [`git`] without trimming: the output exactly as git wrote it.
fn git_raw(dir: &Path, args: &[&str]) -> Result<String> {
    let mut cmd = Command::new("git");
    crate::identity::scrub_git_env(&mut cmd);
    let out = cmd
        .env("GIT_TERMINAL_PROMPT", "0")
        .arg("-C")
        .arg(dir)
        .args(args)
        .stdin(Stdio::null())
        .output()
        .map_err(|e| refuse(format!("cannot run git: {e}")))?;
    if out.status.success() {
        Ok(String::from_utf8_lossy(&out.stdout).into_owned())
    } else {
        Err(String::from_utf8_lossy(&out.stderr).trim().to_string())
    }
}

/// The paths of `git status --porcelain=v1 -z` output, relative to the repository root. An
/// entry is `XY path`; a rename or copy (`R`/`C` in either column) is followed by its source
/// path, which is listed as well.
fn changed_paths(z: &str) -> Vec<String> {
    let mut out = Vec::new();
    let mut entries = z.split('\0').filter(|e| !e.is_empty());
    while let Some(entry) = entries.next() {
        let (xy, path) = (entry.get(..2).unwrap_or(""), entry.get(3..).unwrap_or(""));
        out.push(path.to_string());
        if xy.contains(['R', 'C']) {
            if let Some(source) = entries.next() {
                out.push(source.to_string());
            }
        }
    }
    out
}

/// The branch of a new-form SPEC id: `SPEC-20261118-spec-take` → `spec/20261118-spec-take`.
pub fn branch_of(id: &str) -> Option<String> {
    (spec_doc::id_form(id) == Some(IdForm::New))
        .then(|| format!("spec/{}", id.strip_prefix("SPEC-").unwrap_or(id)))
}

/// A value shown in a printed command: single-quoted when it holds anything but a plain word.
fn quoted(s: &str) -> String {
    if !s.is_empty()
        && s.bytes()
            .all(|b| b.is_ascii_alphanumeric() || b"-_./:".contains(&b))
    {
        s.to_string()
    } else {
        format!("'{}'", s.replace('\'', "'\\''"))
    }
}

/// The SPEC ids the `Moai-Spec` trailers of `range` name.
fn trailer_specs(dir: &Path, range: &str) -> Result<Vec<String>> {
    let out = git(
        dir,
        &[
            "log",
            "--no-merges",
            "--format=%(trailers:key=Moai-Spec,valueonly,separator=%x1e)",
            range,
            "--",
        ],
    )
    .map_err(refuse)?;
    Ok(out
        .split(['\n', '\x1e'])
        .map(str::trim)
        .filter(|s| !s.is_empty())
        .map(String::from)
        .collect())
}

/// `moai spec take`: see the module documentation. Every refusal happens before any write.
pub fn take(r: &Request<'_>) -> Result<Value> {
    if r.draft_pr && !r.push {
        return Err(refuse(
            "--draft-pr needs --push: a draft pull request needs the branch on the remote; nothing was changed",
        ));
    }
    let branch = branch_of(r.id).ok_or_else(|| {
        refuse(format!(
            "{} is not a SPEC id of the form SPEC-<yyyymmdd>-<slug> (a Go-generation SPEC has no branch); nothing was changed",
            shown(r.id)
        ))
    })?;
    let doc = spec_doc::find(r.project, r.id).map_err(refuse)?;
    let errors = doc.errors();
    if errors > 0 {
        return Err(refuse(format!(
            "{} has {errors} error(s); `moai spec check {}` lists them; nothing was changed",
            r.id, r.id
        )));
    }
    let title = doc.title.clone().unwrap_or_else(|| r.id.to_string());
    let dir = r.project;
    git(dir, &["rev-parse", "--show-toplevel"])
        .map_err(|_| refuse("the project is not in a git work tree; nothing was changed"))?;

    // A change outside the SPEC's own folder stops the switch (a fresh `spec new` is fine).
    let own = format!("{}/{}/", spec_doc::SPECS_DIR, r.id);
    let status = git_raw(
        dir,
        &["status", "--porcelain=v1", "-z", "--untracked-files=all"],
    )
    .map_err(refuse)?;
    let prefix = git(dir, &["rev-parse", "--show-prefix"]).map_err(refuse)?;
    let outside: Vec<String> = changed_paths(&status)
        .into_iter()
        .filter(|p| match p.strip_prefix(&prefix) {
            Some(rel) => !rel.starts_with(&own),
            None => true,
        })
        .collect();
    if !outside.is_empty() {
        let shown: Vec<String> = outside.iter().take(5).map(|p| shown(p)).collect();
        return Err(refuse(format!(
            "the work tree has changes outside {own}: {}; commit or set them aside first; nothing was changed",
            shown.join(", ")
        )));
    }

    let current = git(dir, &["rev-parse", "--abbrev-ref", "HEAD"]).map_err(refuse)?;
    let exists = git(
        dir,
        &[
            "rev-parse",
            "--verify",
            "--quiet",
            &format!("refs/heads/{branch}"),
        ],
    )
    .is_ok();
    let recorded = git(dir, &["config", &format!("branch.{branch}.moai-base")]).ok();
    let base = match (r.base, &recorded) {
        (Some(b), _) => b.to_string(),
        (None, Some(b)) if exists => b.clone(),
        (None, _) if current != "HEAD" && current != branch => current.clone(),
        (None, _) => {
            return Err(refuse(
                "no base: pass --base <branch> (the checkout is detached or already on the SPEC branch without a recorded base); nothing was changed",
            ))
        }
    };
    let (created, switched) = if exists {
        let others: Vec<String> = trailer_specs(dir, &format!("{base}..{branch}"))?
            .into_iter()
            .filter(|s| s != r.id)
            .collect();
        if let Some(other) = others.first() {
            return Err(refuse(format!(
                "{branch} already carries work of {} (a Moai-Spec trailer); one branch holds one SPEC; nothing was changed",
                shown(other)
            )));
        }
        if current == branch {
            (false, false)
        } else {
            git(dir, &["switch", "--quiet", &branch]).map_err(refuse)?;
            (false, true)
        }
    } else {
        git(
            dir,
            &[
                "rev-parse",
                "--verify",
                "--quiet",
                &format!("{base}^{{commit}}"),
            ],
        )
        .map_err(|_| {
            refuse(format!(
                "the base {} is not a commit; nothing was changed",
                shown(&base)
            ))
        })?;
        git(dir, &["switch", "--quiet", "-c", &branch, &base]).map_err(refuse)?;
        (true, true)
    };
    if recorded.is_none() || r.base.is_some() {
        git(
            dir,
            &["config", &format!("branch.{branch}.moai-base"), &base],
        )
        .map_err(refuse)?;
    }

    let card = match r.card {
        Some(c) => {
            let home = crate::ledger::command::resolve_home(r.home.map(Path::to_path_buf))
                .map_err(|e| refuse(e.to_string()))?;
            Some(crate::card::spec_link::link(
                r.project,
                &home,
                c,
                Some(r.id),
                None,
            )?)
        }
        None => None,
    };

    let pr_args = |base: &str| -> Vec<String> {
        vec![
            "pr".into(),
            "create".into(),
            "--draft".into(),
            "--base".into(),
            base.into(),
            "--head".into(),
            branch.clone(),
            "--title".into(),
            title.clone(),
            "--body".into(),
            format!("SPEC: {}", r.id),
        ]
    };
    let next = vec![
        format!("moai trailers --spec {}", r.id),
        format!("git push -u {} {branch}", quoted(r.remote)),
        format!(
            "gh {}",
            pr_args(&base)
                .iter()
                .map(|a| quoted(a))
                .collect::<Vec<_>>()
                .join(" ")
        ),
    ];

    let mut pushed = false;
    let mut draft_pr = Value::Null;
    if r.push {
        git(dir, &["push", "--quiet", "-u", r.remote, &branch])
            .map_err(|e| format!("spec take: the push failed after the branch was taken: {e}"))?;
        pushed = true;
        if r.draft_pr {
            let mut gh = Command::new(r.gh);
            crate::identity::scrub_git_env(&mut gh);
            let out = gh
                .args(pr_args(&base))
                .current_dir(dir)
                .stdin(Stdio::null())
                .output()
                .map_err(|e| format!("spec take: cannot run gh: {e}"))?;
            if !out.status.success() {
                return Err(format!(
                    "spec take: gh pr create failed after the push: {}",
                    String::from_utf8_lossy(&out.stderr).trim()
                ));
            }
            let url = String::from_utf8_lossy(&out.stdout)
                .lines()
                .rev()
                .find(|l| l.starts_with("https://"))
                .unwrap_or_default()
                .to_string();
            draft_pr = json!({"url": url});
        }
    }
    Ok(json!({
        "action": "take", "spec": r.id, "status": doc.status, "branch": branch, "base": base,
        "created": created, "switched": switched, "card": card, "next": next,
        "pushed": pushed, "draft_pr": draft_pr,
        "writes": created || switched || card.is_some() || pushed,
    }))
}

/// Text shown in a message: anything but ASCII letters, digits and `-_./` reads `?`.
fn shown(s: &str) -> String {
    s.chars()
        .map(|c| {
            if c.is_ascii_alphanumeric() || "-_./".contains(c) {
                c
            } else {
                '?'
            }
        })
        .collect()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn the_branch_follows_the_new_form_id_only() {
        assert_eq!(
            branch_of("SPEC-20261118-spec-take").as_deref(),
            Some("spec/20261118-spec-take")
        );
        assert_eq!(branch_of("SPEC-AUTH-001"), None);
        assert_eq!(branch_of("C-87"), None);
    }

    #[test]
    fn every_changed_path_is_read_whole_with_a_rename_source() {
        assert_eq!(
            changed_paths(" M a.txt\0?? dir/new file\0R  b.txt\0old.txt\0"),
            ["a.txt", "dir/new file", "b.txt", "old.txt"]
        );
        assert!(changed_paths("").is_empty());
    }

    #[test]
    fn printed_values_are_quoted_unless_plain() {
        assert_eq!(quoted("main"), "main");
        assert_eq!(quoted("Login fix"), "'Login fix'");
        assert_eq!(quoted("it's"), "'it'\\''s'");
    }
}
