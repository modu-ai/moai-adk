//! `moai spec close <SPEC-ID>` (SDD-8, SPEC-20261118-spec-close): close an approved SPEC whose
//! work is merged. Every check reads only: the SPEC (new id form, no error, status `approved`),
//! the merge (the SPEC's local branch, when there is one, is reachable from the base, and a
//! commit reachable from the base carries `Moai-Spec: <id>`), and the closed text (the status
//! line set to `completed` passes `spec check --closure`). Then `--card` closes the card through
//! `card done` with its own refusals, and last the SPEC file is replaced only while it still
//! holds the bytes that were checked. Nothing is committed or pushed: the commit is printed.
//! git runs through the same helper as `spec take` (`scrub_git_env`, no terminal prompt).

use crate::spec_doc::{self, Options};
use crate::spec_take::{git, quoted, shown};
use serde_json::{json, Value};
use std::fs;
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

const REFUSED: &str = "spec close refused: ";

fn refuse(why: impl std::fmt::Display) -> String {
    format!("{REFUSED}{why}")
}

/// What `close` was asked to do.
pub struct Request<'a, G: Fn(&str) -> Option<String>> {
    pub project: &'a Path,
    pub id: &'a str,
    /// The base the work must be merged into; `None` takes the base `spec take` recorded, else
    /// the current branch.
    pub base: Option<&'a str>,
    pub card: Option<&'a str>,
    /// The generation the session holds the card at (`card take`).
    pub generation: Option<i64>,
    pub home: Option<&'a Path>,
    /// The environment (the session id for `--card` and the trailer preview).
    pub get: G,
}

/// The text with its frontmatter `status:` line replaced by `status: completed`, the line
/// ending kept; `None` when the frontmatter has no status line.
pub fn completed_text(text: &str) -> Option<String> {
    let mut out = String::with_capacity(text.len() + 2);
    let mut in_front = false;
    let mut done = false;
    for (i, line) in text.split_inclusive('\n').enumerate() {
        let bare = line.trim_end_matches(['\n', '\r']);
        if i == 0 {
            if bare != "---" {
                return None;
            }
            in_front = true;
        } else if in_front && bare == "---" {
            in_front = false;
        } else if in_front && !done && bare.starts_with("status:") {
            out.push_str("status: completed");
            out.push_str(&line[bare.len()..]);
            done = true;
            continue;
        }
        out.push_str(line);
    }
    done.then_some(out)
}

/// The commit reachable from `base` whose `Moai-Spec` trailer names `id`, if any.
fn trailer_commit(dir: &Path, base: &str, id: &str) -> Result<Option<String>> {
    let out = git(
        dir,
        &[
            "log",
            "--fixed-strings",
            &format!("--grep={id}"),
            "--format=%H%x1f%(trailers:key=Moai-Spec,valueonly,separator=%x1e)",
            "--end-of-options",
            base,
            "--",
        ],
    )
    .map_err(refuse)?;
    Ok(out.lines().find_map(|l| {
        let (sha, values) = l.split_once('\x1f')?;
        values
            .split('\x1e')
            .any(|v| v.trim() == id)
            .then(|| sha.to_string())
    }))
}

/// Checks `--card` read-only before anything is written: a C-number of a card in this
/// project's ledger that names this SPEC.
fn card_check(project: &Path, home: &Path, reference: &str, id: &str) -> Result<()> {
    use crate::card::input::{self, CardRef};
    use crate::ledger::{self, LedgerError};
    let card_id = match input::parse_ref(reference) {
        Ok(CardRef::Id(n)) => n,
        _ => return Err(refuse("--card must name a card by its C-number (C-<n>)")),
    };
    let name = input::card_name(card_id);
    let project_id = ledger::command::project_id_of(project).map_err(refuse)?;
    let read = ledger::open_read_only(home, &project_id).map_err(|e| match e {
        LedgerError::Absent => refuse("--card needs this project's ledger and there is none"),
        other => refuse(other),
    })?;
    match crate::card::spec_link::kind_and_spec(read.conn(), card_id) {
        Ok(Some((_, Some(spec)))) if spec == id => Ok(()),
        Ok(Some((_, Some(spec)))) => Err(refuse(format!(
            "{name} names {}, not {id}; nothing was changed",
            shown(&spec)
        ))),
        Ok(Some((_, None))) => Err(refuse(format!(
            "{name} names no SPEC; `moai card spec {name} {id}` links it; nothing was changed"
        ))),
        Ok(None) => Err(refuse(format!("there is no {name}; nothing was changed"))),
        Err(e) => Err(refuse(e)),
    }
}

/// `moai spec close`: see the module documentation. Every refusal happens before any write.
pub fn close<G: Fn(&str) -> Option<String>>(r: &Request<'_, G>) -> Result<Value> {
    let Some(branch) = crate::spec_take::branch_of(r.id) else {
        return Err(refuse(format!(
            "{} is not a SPEC id of the form SPEC-<yyyymmdd>-<slug> (a Go-generation SPEC is \
             read, never closed); nothing was changed",
            shown(r.id)
        )));
    };
    if r.base.is_some_and(|b| b.is_empty() || b.starts_with('-')) {
        return Err(refuse(
            "--base must not be empty or start with `-`; nothing was changed",
        ));
    }
    let doc = spec_doc::find(r.project, r.id).map_err(refuse)?;
    if doc.errors() > 0 {
        return Err(refuse(format!(
            "{} has {} error(s); `moai spec check {}` lists them; nothing was changed",
            r.id,
            doc.errors(),
            r.id
        )));
    }
    let status = doc.status_as.clone().unwrap_or_default();
    match status.as_str() {
        "approved" => {}
        "draft" => {
            return Err(refuse(format!(
                "{} is a draft: approval is the user's to record before a SPEC can close; \
                 nothing was changed",
                r.id
            )))
        }
        "completed" => {
            return Err(refuse(format!(
                "{} is already completed; nothing was changed",
                r.id
            )))
        }
        other => {
            return Err(refuse(format!(
                "{} has status {}; only an approved SPEC closes; nothing was changed",
                r.id,
                shown(other)
            )))
        }
    }

    // The merge, read only.
    let dir = r.project;
    git(dir, &["rev-parse", "--show-toplevel"])
        .map_err(|_| refuse("the project is not in a git work tree; nothing was changed"))?;
    let recorded = git(dir, &["config", &format!("branch.{branch}.moai-base")]).ok();
    let current = git(dir, &["rev-parse", "--abbrev-ref", "HEAD"]).map_err(refuse)?;
    let base =
        match (r.base, recorded) {
            (Some(b), _) => b.to_string(),
            (None, Some(b)) => b,
            (None, None) if current != "HEAD" && current != branch => current,
            (None, None) => return Err(refuse(
                "no base: pass --base <branch> (the checkout is detached or on the SPEC branch); \
                 nothing was changed",
            )),
        };
    let base_commit = if base.starts_with('-') {
        None
    } else {
        git(
            dir,
            &[
                "rev-parse",
                "--verify",
                "--quiet",
                "--end-of-options",
                &format!("{base}^{{commit}}"),
            ],
        )
        .ok()
    };
    let Some(base_commit) = base_commit else {
        return Err(refuse(format!(
            "the base {} is not a commit; nothing was changed",
            shown(&base)
        )));
    };
    let branch_ref = format!("refs/heads/{branch}");
    let has_branch = git(dir, &["rev-parse", "--verify", "--quiet", &branch_ref]).is_ok();
    if has_branch
        && git(
            dir,
            &["merge-base", "--is-ancestor", &branch_ref, &base_commit],
        )
        .is_err()
    {
        return Err(refuse(format!(
            "{branch} is not merged into {}: a commit of it is not reachable from the base (a \
             squash merge leaves the branch behind: delete it once merged, or pass the base it \
             was merged into); nothing was changed",
            shown(&base)
        )));
    }
    let Some(evidence) = trailer_commit(dir, &base_commit, r.id)? else {
        return Err(refuse(format!(
            "no commit reachable from {} carries `Moai-Spec: {}`, so its work is not merged \
             there; nothing was changed",
            shown(&base),
            r.id
        )));
    };

    // The closed text, checked before anything is written.
    let (root, _) = spec_doc::specs_dir(r.project).map_err(refuse)?;
    let rel = format!("{}/{}/{}", spec_doc::SPECS_DIR, r.id, spec_doc::SPEC_FILE);
    let path = root.join(&rel);
    let before = fs::read(&path).map_err(|e| refuse(format!("cannot read {rel}: {e}")))?;
    let text = String::from_utf8(before.clone())
        .map_err(|_| refuse(format!("{rel} is not UTF-8; nothing was changed")))?;
    let after = completed_text(&text)
        .ok_or_else(|| refuse(format!("{rel} has no status line; nothing was changed")))?;
    let closed = spec_doc::parse(
        &after,
        &Options {
            folder: Some(r.id.to_string()),
            closure: true,
        },
    );
    if closed.id.as_deref() != Some(r.id) || closed.errors() > 0 {
        let codes: Vec<&str> = closed
            .findings
            .iter()
            .filter(|f| matches!(f.severity, spec_doc::Severity::Error))
            .map(|f| f.code)
            .collect();
        return Err(refuse(format!(
            "{} would not pass `spec check --closure` once completed ({}); fill `## Result` with \
             each AC `met` or `withdrawn`; nothing was changed",
            r.id,
            codes.join(", ")
        )));
    }

    // The card: checked read-only, then closed by `card done` with its own refusals.
    let card = match r.card {
        Some(c) => {
            let home = crate::ledger::command::resolve_home(r.home.map(Path::to_path_buf))
                .map_err(refuse)?;
            let holder = crate::card::dispatch::session_holder("done", &r.get).map_err(refuse)?;
            let generation = r
                .generation
                .ok_or_else(|| refuse("--card needs --generation"))?;
            card_check(r.project, &home, c, r.id)?;
            Some(
                crate::card::work::done(r.project, &home, c, &holder, generation)
                    .map_err(refuse)?,
            )
        }
        None => None,
    };

    crate::apply::publish_replace(&path, Some(&before), after.as_bytes()).map_err(|e| {
        if card.is_some() {
            format!("spec close: the card is done, but {rel} was not written: {e}")
        } else {
            refuse(format!("{rel} was not written: {e}; nothing was changed"))
        }
    })?;

    let key = crate::trailers::worktree_key(dir);
    let trailers = crate::trailers::format(r.id, &r.get, key.as_deref())
        .ok()
        .and_then(|v| v["text"].as_str().map(|s| s.trim_end().to_string()))
        .unwrap_or_else(|| format!("Moai-Spec: {}", r.id));
    let next = vec![
        format!("moai trailers --spec {}", r.id),
        format!(
            "git commit -m {} -m {} -- {rel}",
            quoted(&format!("docs(spec): close {}", r.id)),
            quoted(&trailers)
        ),
    ];
    Ok(json!({
        "action": "close", "spec": r.id, "from": status, "status": "completed",
        "path": rel, "base": base,
        "merged": {"branch": has_branch.then_some(&branch), "trailer_commit": evidence},
        "card": card, "next": next, "writes": true,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn only_the_frontmatter_status_line_changes() {
        let text = "---\nid: X\nstatus: approved\r\n---\n\nstatus: approved\n";
        assert_eq!(
            completed_text(text).unwrap(),
            "---\nid: X\nstatus: completed\r\n---\n\nstatus: approved\n"
        );
        assert_eq!(completed_text("---\nid: X\n---\nstatus: approved\n"), None);
        assert_eq!(completed_text("no frontmatter\n"), None);
    }
}
