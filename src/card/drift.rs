//! The integration branch against the ledger (SPEC-20261121-ref-guard D-13, SPEC-20261121-card-
//! land): where the ledger says the branch must be, the refusal `card land` makes when it is
//! somewhere else, the read-only view of the same comparison and the one way back.
//!
//! The rule is strict. The tip must equal the result commit of the newest successful landing of
//! the branch: a `card.land_started` row followed by a `card.landed` row of the same card and
//! head. The only other accepted tips are the two of an interrupted landing, the newest start row
//! that no landed row follows: the tip it started from (the move was refused or never done) and
//! its target (the move was done and the close is missing, which the rerun path closes). A
//! branch with no start row has no landing to compare with. Anything else is drift, a descendant
//! of the last landing included: a commit nobody landed is not a landing, and a rewind to where
//! the previous landing started would make the next landing drop a landed change.
//!
//! `git branch -C` writes a reference outside every reference transaction, so no hook sees it;
//! this module is the layer that notices it at the next landing and shows it without landing.
//!
//! The way back is `card drift --restore`: a preview with the move, the commits it drops and a
//! digest, and, only with the digest of a preview of the same state, a permission reference and
//! the leader word, the move itself under the guard (the same intent record `card land` writes)
//! and one `card.tip_restored` ledger row. Nothing here prints or runs a bare `git update-ref`.

use super::command::open_for_write;
use super::guard;
use super::input;
use super::landed::{self, RefMove};
use crate::ledger::workflow::{self, LEADER_BY};
use crate::ledger::{self, LedgerError};
use rusqlite::Connection;
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// The most dropped commits a preview lists.
const DROPS_LISTED: usize = 20;

/// Where the ledger says the branch must be.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Expected {
    /// The commit a restore moves the branch to: the result of the newest successful landing, or
    /// for a branch whose first landing was interrupted, the tip that landing started from.
    pub commit: String,
    /// The card that landing belongs to.
    pub card: i64,
    /// `landing` or `interrupted_landing`.
    pub source: &'static str,
    /// Every tip that is no drift.
    pub accepted: Vec<String>,
}

struct Row {
    card: i64,
    from: String,
    head: String,
    landed: bool,
}

fn sql_err(e: rusqlite::Error) -> String {
    LedgerError::from(e).to_string()
}

/// The start rows of `branch`, newest first, each with whether a `card.landed` row of the same
/// card and head follows it.
fn rows(conn: &Connection, branch: &str) -> Result<Vec<Row>> {
    let mut stmt = conn
        .prepare(
            "SELECT s.card_id, json_extract(s.payload, '$.tip'), json_extract(s.payload, '$.head'), \
                    EXISTS (SELECT 1 FROM event l WHERE l.card_id = s.card_id \
                              AND l.kind = 'card.landed' AND l.event_id > s.event_id \
                              AND json_extract(l.payload, '$.head') = json_extract(s.payload, '$.head')) \
             FROM event s WHERE s.kind = 'card.land_started' \
               AND json_extract(s.payload, '$.branch') = ?1 \
               AND json_type(s.payload, '$.head') = 'text' AND json_type(s.payload, '$.tip') = 'text' \
             ORDER BY s.event_id DESC",
        )
        .map_err(sql_err)?;
    let found = stmt
        .query_map([branch], |r| {
            Ok(Row {
                card: r.get(0)?,
                from: r.get(1)?,
                head: r.get(2)?,
                landed: r.get(3)?,
            })
        })
        .map_err(sql_err)?
        .collect::<std::result::Result<Vec<_>, _>>()
        .map_err(sql_err)?;
    Ok(found)
}

/// What the ledger says about `branch`, or `None` for a branch it never moved.
pub fn expected(conn: &Connection, branch: &str) -> Result<Option<Expected>> {
    let all = rows(conn, branch)?;
    let Some(newest) = all.first() else {
        return Ok(None);
    };
    let push = |list: &mut Vec<String>, commit: &str| {
        if !list.iter().any(|c| c == commit) {
            list.push(commit.to_string());
        }
    };
    let mut accepted = Vec::new();
    if let Some(done) = all.iter().find(|r| r.landed) {
        push(&mut accepted, &done.head);
        if !newest.landed {
            push(&mut accepted, &newest.from);
            push(&mut accepted, &newest.head);
        }
        return Ok(Some(Expected {
            commit: done.head.clone(),
            card: done.card,
            source: "landing",
            accepted,
        }));
    }
    push(&mut accepted, &newest.from);
    push(&mut accepted, &newest.head);
    Ok(Some(Expected {
        commit: newest.from.clone(),
        card: newest.card,
        source: "interrupted_landing",
        accepted,
    }))
}

/// The restore command a drift answer names. It is a preview until the user says yes.
const RESTORE_COMMAND: &str = "moai card drift --restore";

/// The refusal `card land` makes when `tip` is drift, or `None`: the tip is accepted or the
/// branch has no landing to compare with.
pub fn refusal(conn: &Connection, branch: &str, tip: &str) -> Result<Option<String>> {
    let Some(want) = expected(conn, branch)? else {
        return Ok(None);
    };
    if want.accepted.iter().any(|c| c == tip) {
        return Ok(None);
    }
    Ok(Some(format!(
        "drift: the integration branch {branch} points at {tip}, but the newest landing the \
         ledger recorded ({}) left it at {}: something moved the branch outside `moai card land` \
         (for example `git branch -C`, which no hook sees, or a commit added by hand, which is \
         no landing). `moai card drift` shows the comparison; after the \
         user's pick, `{RESTORE_COMMAND}` previews the move back to {} and applies it only with \
         the user's yes; then land again",
        input::card_name(want.card),
        want.commit,
        want.commit,
    )))
}

fn compare_json(branch: &str, tip: &str, want: &Option<Expected>) -> Value {
    let Some(w) = want else {
        return json!({
            "branch": branch, "tip": tip, "state": "no_landing", "expected": null,
            "accepted": [], "restore_command": null,
        });
    };
    let drift = !w.accepted.iter().any(|c| c == tip);
    json!({
        "branch": branch, "tip": tip,
        "state": if drift { "drift" } else { "ok" },
        "expected": {"commit": w.commit, "card": input::card_name(w.card), "source": w.source},
        "accepted": w.accepted,
        "restore_command": if drift { Value::from(RESTORE_COMMAND) } else { Value::Null },
    })
}

/// What `card drift --restore` is about to do.
struct Plan {
    branch: String,
    full: String,
    from: String,
    to: String,
    card: i64,
    drops: Vec<String>,
    drops_total: usize,
    digest: String,
}

fn digest_of(branch: &str, from: &str, to: &str, card: i64) -> String {
    format!(
        "{:x}",
        Sha256::digest(format!(
            "moai card drift --restore\n{branch}\n{from}\n{to}\n{card}\n"
        ))
    )
}

fn plan(project: &Path, conn: &Connection, branch: &str, tip: &str) -> Result<Plan> {
    let want = expected(conn, branch)?;
    let Some(w) = want else {
        return Err(format!(
            "card conflict in drift: no drift: the ledger never moved {branch}, so there is no \
             recorded commit to restore; nothing was changed"
        ));
    };
    if w.accepted.iter().any(|c| c == tip) {
        return Err(format!(
            "card conflict in drift: no drift: {branch} points at {tip}, which the ledger \
             accepts; nothing was changed"
        ));
    }
    let full = landed::branch_ref(branch)?;
    let all = landed::dropped_by(project, tip, &w.commit, DROPS_LISTED + 1)?;
    let drops_total = all.len();
    let drops: Vec<String> = all.into_iter().take(DROPS_LISTED).collect();
    Ok(Plan {
        digest: digest_of(branch, tip, &w.commit, w.card),
        branch: branch.to_string(),
        full,
        from: tip.to_string(),
        to: w.commit,
        card: w.card,
        drops,
        drops_total,
    })
}

fn plan_json(p: &Plan, moved: bool) -> Value {
    json!({
        "branch": p.branch, "from": p.from, "to": p.to, "card": input::card_name(p.card),
        "drops": p.drops, "drops_total": p.drops_total,
        "plan_digest": p.digest, "moved": moved,
        "apply": format!(
            "moai card drift --restore --apply --plan-digest {} --permission-reference <the \
             user's yes> --by {LEADER_BY}",
            p.digest
        ),
    })
}

/// What `card drift` was given.
pub struct DriftRequest<'a> {
    /// `--restore`: preview the move back (and with `--apply`, do it).
    pub restore: bool,
    /// `--apply`: do the move; needs the digest, the permission reference and the leader word.
    pub apply: bool,
    /// `--plan-digest`: the digest of a preview of the same state.
    pub plan_digest: Option<&'a str>,
    /// `--permission-reference`: the reference of the user's yes.
    pub permission_reference: Option<&'a str>,
    /// `--by`: who restores; only the leader word is accepted.
    pub by: Option<&'a str>,
}

fn refused(why: impl std::fmt::Display) -> String {
    format!("card drift refused: {why}")
}

/// A reference that names a person's yes: short, printable, with no whitespace or control.
fn reference_ok(r: &str) -> bool {
    (1..=80).contains(&r.chars().count())
        && r.chars()
            .all(|c| c.is_ascii_alphanumeric() || matches!(c, '.' | '_' | ':' | '/' | '-'))
}

/// `card drift`: see the module comment.
pub fn drift(project: &Path, home: &Path, req: &DriftRequest<'_>) -> Result<Value> {
    if req.apply && !req.restore {
        return Err(refused("--apply belongs to --restore"));
    }
    if !req.restore
        && (req.plan_digest.is_some() || req.permission_reference.is_some() || req.by.is_some())
    {
        return Err(refused(
            "--plan-digest, --permission-reference and --by belong to `--restore --apply`",
        ));
    }
    if !req.apply
        && (req.plan_digest.is_some() || req.permission_reference.is_some() || req.by.is_some())
    {
        return Err(refused(
            "--plan-digest, --permission-reference and --by belong to `--restore --apply`; the \
             preview takes none",
        ));
    }
    let branch = landed::integration_branch(project).map_err(|why| {
        refused(format!(
            "the profile gives no integration branch ({why}); nothing was changed"
        ))
    })?;
    let tip = landed::tip_of(project, &branch).map_err(refused)?;

    if !req.apply {
        let id = ledger::command::project_id_of(project)?;
        let reader = ledger::open_read_only(home, &id).map_err(|e| match e {
            LedgerError::Absent => super::command::ABSENT_HINT.to_string(),
            other => other.to_string(),
        })?;
        let want = expected(reader.conn(), &branch)?;
        let mut out = compare_json(&branch, &tip, &want);
        out["action"] = json!("drift");
        out["project_id"] = json!(id);
        out["writes"] = json!(false);
        if req.restore {
            let p = plan(project, reader.conn(), &branch, &tip)?;
            out["plan"] = plan_json(&p, false);
        }
        return Ok(out);
    }

    // The apply: every input first, then the plan again from the state now, then the move.
    if req.by != Some(LEADER_BY) {
        return Err(refused(format!(
            "only the leader restores the branch: give --by {LEADER_BY}; nothing was changed"
        )));
    }
    let Some(digest) = req.plan_digest else {
        return Err(refused(
            "--apply needs --plan-digest from a preview of the same state; nothing was changed",
        ));
    };
    let Some(reference) = req.permission_reference else {
        return Err(refused(
            "--apply needs --permission-reference, the reference of the user's yes; nothing \
             was changed",
        ));
    };
    if !reference_ok(reference) {
        return Err(refused(
            "--permission-reference holds 1 to 80 letters, digits or . _ : / -; nothing was \
             changed",
        ));
    }
    let (id, mut writer) = open_for_write(project, home)?;
    let p = plan(project, writer.conn(), &branch, &tip)?;
    if p.digest != digest {
        return Err(format!(
            "card conflict in drift: the plan digest is not the digest of the state now ({branch} \
             at {tip}, restore to {}): run the preview again; nothing was changed",
            p.to
        ));
    }
    let trees = landed::checked_out_at(project, &p.full).map_err(refused)?;
    if !trees.is_empty() {
        return Err(format!(
            "card conflict in drift: the integration branch {branch} is checked out in {}: a \
             move under it would leave that worktree behind the reference; nothing was changed",
            trees.join(", ")
        ));
    }
    let common = guard::common_dir_of(project).map_err(refused)?;
    guard::write_intent(&common, &p.full, &p.from, &p.to).map_err(refused)?;
    let message = format!("moai card drift --restore {}", &p.to[..12.min(p.to.len())]);
    let attempt = landed::move_branch(project, &p.full, &p.to, &p.from, &message);
    guard::remove_intent(&common);
    match attempt.map_err(refused)? {
        RefMove::Moved => {}
        RefMove::Lost(why) => {
            return Err(format!(
                "card conflict in drift: git refused to move {branch} back to {}: {why}; \
                 nothing was changed",
                p.to
            ));
        }
    }
    workflow::record_tip_restored(
        &mut writer,
        p.card,
        LEADER_BY,
        &branch,
        &p.from,
        &p.to,
        reference,
    )
    .map_err(|e| {
        format!(
            "card drift refused after the move: {e}; the integration branch {branch} already \
             moved back to {} and the ledger row is missing",
            p.to
        )
    })?;
    let mut out = compare_json(&branch, &p.to, &expected(writer.conn(), &branch)?);
    out["action"] = json!("drift");
    out["project_id"] = json!(id);
    out["writes"] = json!(true);
    out["plan"] = plan_json(&p, true);
    out["permission_reference"] = json!(reference);
    Ok(out)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn a_reference_names_a_yes_in_plain_characters() {
        assert!(reference_ok("user-pick-1"));
        assert!(reference_ok("note:6080/yes"));
        assert!(!reference_ok(""));
        assert!(!reference_ok("two words"));
        assert!(!reference_ok("tab\there"));
        assert!(!reference_ok(&"x".repeat(81)));
        assert!(reference_ok(&"x".repeat(80)));
    }

    #[test]
    fn the_digest_changes_with_every_part_of_the_move() {
        let base = digest_of("integ", "a", "b", 1);
        assert_ne!(base, digest_of("other", "a", "b", 1));
        assert_ne!(base, digest_of("integ", "x", "b", 1));
        assert_ne!(base, digest_of("integ", "a", "y", 1));
        assert_ne!(base, digest_of("integ", "a", "b", 2));
        assert_eq!(base, digest_of("integ", "a", "b", 1));
    }
}
