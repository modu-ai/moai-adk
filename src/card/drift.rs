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
//! digest bound to a one-use nonce the preview writes into the git folder (the only file it
//! writes), and, only with the digest of the latest preview of the same state, a permission
//! reference and the leader word, the move itself under the guard (a restore intent that names the
//! ledger's recorded landing and its row: the one move that may be no fast-forward) and one
//! `card.tip_restored` ledger row. The plain view writes no file at all. Nothing here prints or runs a bare `git update-ref`.
//!
//! A `card land-sync` (SPEC-20261125-land-sync-merge) moves the branch too, and it has no card:
//! its two rows, `card.sync_started` before the move and `card.synced` after it, carry the merge
//! commit, the source tip and the old tip, and `card_id` NULL. They are a second pair beside the
//! landing pair and count the same way: the newest start row of the two kinds decides, a complete
//! sync expects its merge commit, an interrupted one accepts the merge (the move was done) and the
//! old tip (it was not). The two kinds are read by two queries, so a card-less row never reaches
//! the landing query, which reads `card_id` as a number.

use super::command::open_for_write;
use super::guard;
use super::input;
use super::landed::{self, RefMove};
use crate::ledger::workflow::{self, LEADER_BY};
use crate::ledger::{self, LedgerError, WriteLedger};
use rusqlite::{params, Connection, Transaction};
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// The most dropped commits a preview lists.
const DROPS_LISTED: usize = 20;

/// The ledger row `card land-sync` writes before it moves the branch (no card).
pub const SYNC_STARTED: &str = "card.sync_started";
/// The ledger row `card land-sync` writes after it moved the branch (no card).
pub const SYNCED: &str = "card.synced";
/// The ledger row `card drift --restore --apply` writes after the move back.
const TIP_RESTORED: &str = "card.tip_restored";

/// Where the ledger says the branch must be.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Expected {
    /// The commit a restore moves the branch to: the result of the newest successful landing or
    /// sync, or for a branch whose first move was interrupted, the tip that move started from.
    pub commit: String,
    /// The card that landing belongs to; `None` for a sync, which has no card.
    pub card: Option<i64>,
    /// The ledger row the commit was read from: the landed (or synced) row of the newest
    /// successful move, or the start row of an interrupted first move.
    pub row: i64,
    /// `landing`, `sync`, `interrupted_landing` or `interrupted_sync`.
    pub source: &'static str,
    /// Every tip that is no drift.
    pub accepted: Vec<String>,
}

/// One start row of a landing or of a sync. For a sync `from` is the old tip, `head` the merge
/// commit and `landed_row` the `card.synced` row.
struct Row {
    card: Option<i64>,
    sync: bool,
    from: String,
    head: String,
    /// The id of the `card.landed` (or `card.synced`) row that follows this start row, when there
    /// is one.
    landed_row: Option<i64>,
    /// The id of the start row itself.
    start_row: i64,
}

fn sql_err(e: rusqlite::Error) -> String {
    LedgerError::from(e).to_string()
}

/// The start rows of `branch`, landings and syncs together, newest first, each with whether its
/// closing row follows it.
fn rows(conn: &Connection, branch: &str) -> Result<Vec<Row>> {
    let mut found = landing_rows(conn, branch)?;
    found.extend(sync_rows(conn, branch)?);
    found.sort_by_key(|r| std::cmp::Reverse(r.start_row));
    Ok(found)
}

/// The `card.sync_started` rows of `branch`, newest first, each with whether a `card.synced` row
/// of the same merge and old tip follows it. Their `card_id` is NULL and is never selected.
fn sync_rows(conn: &Connection, branch: &str) -> Result<Vec<Row>> {
    let mut stmt = conn
        .prepare(&format!(
            "SELECT json_extract(s.payload, '$.old_tip'), json_extract(s.payload, '$.merge'), \
                    (SELECT min(l.event_id) FROM event l WHERE l.kind = '{SYNCED}' \
                              AND l.event_id > s.event_id \
                              AND json_extract(l.payload, '$.branch') = ?1 \
                              AND json_extract(l.payload, '$.merge') = json_extract(s.payload, '$.merge') \
                              AND json_extract(l.payload, '$.old_tip') = json_extract(s.payload, '$.old_tip')), \
                    s.event_id \
             FROM event s WHERE s.kind = '{SYNC_STARTED}' \
               AND json_extract(s.payload, '$.branch') = ?1 \
               AND json_type(s.payload, '$.merge') = 'text' AND json_type(s.payload, '$.old_tip') = 'text' \
             ORDER BY s.event_id DESC"
        ))
        .map_err(sql_err)?;
    let found = stmt
        .query_map([branch], |r| {
            Ok(Row {
                card: None,
                sync: true,
                from: r.get(0)?,
                head: r.get(1)?,
                landed_row: r.get(2)?,
                start_row: r.get(3)?,
            })
        })
        .map_err(sql_err)?
        .collect::<std::result::Result<Vec<_>, _>>()
        .map_err(sql_err)?;
    Ok(found)
}

/// The `card.land_started` rows of `branch`, newest first, each with whether a `card.landed` row
/// of the same card and head follows it.
fn landing_rows(conn: &Connection, branch: &str) -> Result<Vec<Row>> {
    let mut stmt = conn
        .prepare(
            "SELECT s.card_id, json_extract(s.payload, '$.tip'), json_extract(s.payload, '$.head'), \
                    (SELECT min(l.event_id) FROM event l WHERE l.card_id = s.card_id \
                              AND l.kind = 'card.landed' AND l.event_id > s.event_id \
                              AND json_extract(l.payload, '$.head') = json_extract(s.payload, '$.head')), \
                    s.event_id \
             FROM event s WHERE s.kind = 'card.land_started' \
               AND json_extract(s.payload, '$.branch') = ?1 \
               AND json_type(s.payload, '$.head') = 'text' AND json_type(s.payload, '$.tip') = 'text' \
             ORDER BY s.event_id DESC",
        )
        .map_err(sql_err)?;
    let found = stmt
        .query_map([branch], |r| {
            Ok(Row {
                card: Some(r.get(0)?),
                sync: false,
                from: r.get(1)?,
                head: r.get(2)?,
                landed_row: r.get(3)?,
                start_row: r.get(4)?,
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
    if let Some((done, landed_row)) = all.iter().find_map(|r| r.landed_row.map(|l| (r, l))) {
        push(&mut accepted, &done.head);
        if newest.landed_row.is_none() {
            push(&mut accepted, &newest.from);
            push(&mut accepted, &newest.head);
        }
        return Ok(Some(Expected {
            commit: done.head.clone(),
            card: done.card,
            row: landed_row,
            source: if done.sync { "sync" } else { "landing" },
            accepted,
        }));
    }
    push(&mut accepted, &newest.from);
    push(&mut accepted, &newest.head);
    Ok(Some(Expected {
        commit: newest.from.clone(),
        card: newest.card,
        row: newest.start_row,
        source: if newest.sync {
            "interrupted_sync"
        } else {
            "interrupted_landing"
        },
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
        card_label(want.card),
        want.commit,
        want.commit,
    )))
}

/// The card a drift answer names, or the sync that has none.
fn card_label(card: Option<i64>) -> String {
    card.map_or_else(
        || "a `moai card land-sync` merge".to_string(),
        input::card_name,
    )
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
        "expected": {
            "commit": w.commit, "card": w.card.map(input::card_name), "source": w.source,
        },
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
    card: Option<i64>,
    row: i64,
    drops: Vec<String>,
    drops_total: usize,
    digest: String,
}

/// The digest of a restore plan: the move, the card (`none` for a sync) and the nonce the preview
/// wrote into the git folder, so it cannot be computed without that preview and a used or older
/// one never matches.
fn digest_of(branch: &str, from: &str, to: &str, card: Option<i64>, nonce: &str) -> String {
    let card = card.map_or_else(|| "none".to_string(), |c| c.to_string());
    format!(
        "{:x}",
        Sha256::digest(format!(
            "moai card drift --restore\n{branch}\n{from}\n{to}\n{card}\n{nonce}\n"
        ))
    )
}

/// 128 random bits as hex: from the operating system when it has a source, otherwise from the
/// clock, the process id and the address of a local.
fn new_nonce() -> String {
    use std::io::Read;
    let mut bytes = [0u8; 16];
    let ok = std::fs::File::open("/dev/urandom")
        .and_then(|mut f| f.read_exact(&mut bytes))
        .is_ok();
    if !ok {
        let now = std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .map_or(0, |d| d.as_nanos());
        let seed = format!("{now}:{}:{:p}", std::process::id(), &bytes);
        bytes.copy_from_slice(&Sha256::digest(seed)[..16]);
    }
    bytes.iter().map(|b| format!("{b:02x}")).collect()
}

fn plan(project: &Path, conn: &Connection, branch: &str, tip: &str, nonce: &str) -> Result<Plan> {
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
        digest: digest_of(branch, tip, &w.commit, w.card, nonce),
        branch: branch.to_string(),
        full,
        from: tip.to_string(),
        to: w.commit,
        card: w.card,
        row: w.row,
        drops,
        drops_total,
    })
}

fn plan_json(p: &Plan, moved: bool) -> Value {
    json!({
        "branch": p.branch, "from": p.from, "to": p.to, "card": p.card.map(input::card_name),
        "drops": p.drops, "drops_total": p.drops_total,
        "plan_digest": p.digest, "moved": moved,
        "apply": format!(
            "moai card drift --restore --apply --plan-digest {} --permission-reference <the \
             user's yes> --by {LEADER_BY}",
            p.digest
        ),
    })
}

/// What a `card.sync_started` or `card.synced` row names.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct SyncRow {
    /// The integration branch, spelled as the landing rows spell it.
    pub branch: String,
    /// The merge commit the sync moves the branch to.
    pub merge: String,
    /// The source branch.
    pub from: String,
    /// The tip of the source branch the merge was made from.
    pub from_tip: String,
    /// The integration tip the merge was made on.
    pub old_tip: String,
    /// The approved SPEC the sync names.
    pub spec: String,
}

/// The card-less row `card land-sync` writes before the move; written once per branch, merge and
/// old tip, so a rerun after an interrupted sync adds none.
pub fn record_sync_started(w: &mut WriteLedger, row: &SyncRow) -> Result<()> {
    write_sync(w, row, SYNC_STARTED)
}

/// The card-less row `card land-sync` writes after the move; written once per branch, merge and
/// old tip, and only after the start row of the same sync.
pub fn record_synced(w: &mut WriteLedger, row: &SyncRow) -> Result<()> {
    write_sync(w, row, SYNCED)
}

/// A full lower-case commit id: what git prints and what the landing rows hold.
fn is_commit_id(text: &str) -> bool {
    text.len() == 40 && text.bytes().all(|b| matches!(b, b'0'..=b'9' | b'a'..=b'f'))
}

/// A card-less row with no `card_id`, in the transaction of the caller.
fn add_cardless_event(
    tx: &Transaction<'_>,
    now: i64,
    actor: &str,
    kind: &str,
    payload: &Value,
) -> std::result::Result<(), LedgerError> {
    let key = ledger::new_event_key(tx, kind, None)?;
    tx.execute(
        "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
         VALUES (?1, ?2, ?3, NULL, ?4, ?5)",
        params![key, now, actor, kind, payload.to_string()],
    )?;
    Ok(())
}

/// One `card.sync_started` or `card.synced` row, in one write transaction: refused for a value
/// that is no commit id or an empty name, a no-op when a row of that kind already names the branch,
/// the merge and the old tip, and, for the synced row, refused when no start row of that sync exists.
fn write_sync(w: &mut WriteLedger, row: &SyncRow, kind: &str) -> Result<()> {
    for (name, value) in [
        ("merge", &row.merge),
        ("from_tip", &row.from_tip),
        ("old_tip", &row.old_tip),
    ] {
        if !is_commit_id(value) {
            return Err(format!("card sync row refused: {name} is no commit id"));
        }
    }
    for (name, value) in [
        ("branch", &row.branch),
        ("from", &row.from),
        ("spec", &row.spec),
    ] {
        if value.is_empty() {
            return Err(format!("card sync row refused: {name} is empty"));
        }
    }
    let payload = json!({
        "branch": row.branch, "merge": row.merge, "from": row.from, "from_tip": row.from_tip,
        "old_tip": row.old_tip, "spec": row.spec, "by": LEADER_BY,
    });
    w.with_write_tx(|tx, now| {
        let exists = |of: &str| -> std::result::Result<bool, LedgerError> {
            Ok(tx.query_row(
                "SELECT EXISTS (SELECT 1 FROM event WHERE kind = ?1 \
                   AND json_extract(payload, '$.branch') = ?2 \
                   AND json_extract(payload, '$.merge') = ?3 \
                   AND json_extract(payload, '$.old_tip') = ?4)",
                params![of, row.branch, row.merge, row.old_tip],
                |r| r.get(0),
            )?)
        };
        if exists(kind)? {
            return Ok(Ok(()));
        }
        if kind == SYNCED && !exists(SYNC_STARTED)? {
            return Ok(Err(format!(
                "card sync row refused: no {SYNC_STARTED} row names the merge {} on {}",
                row.merge, row.old_tip
            )));
        }
        add_cardless_event(tx, now, LEADER_BY, kind, &payload)?;
        Ok(Ok(()))
    })
    .map_err(|e| e.to_string())?
}

/// The `card.tip_restored` row of a restore with no card (the move back of a sync): the same
/// payload as the card-bound row, `card_id` NULL.
pub fn record_tip_restored_cardless(
    w: &mut WriteLedger,
    by: &str,
    branch: &str,
    from: &str,
    to: &str,
    reference: &str,
) -> Result<()> {
    w.with_write_tx(|tx, now| {
        add_cardless_event(
            tx,
            now,
            by,
            TIP_RESTORED,
            &json!({"branch": branch, "from": from, "to": to, "reference": reference, "by": by}),
        )
    })
    .map_err(|e| e.to_string())
}

/// The ledger row of a restore: card-bound for a landing, card-less for a sync.
fn record_restore(
    w: &mut WriteLedger,
    card: Option<i64>,
    branch: &str,
    from: &str,
    to: &str,
    reference: &str,
) -> Result<()> {
    match card {
        Some(card) => {
            workflow::record_tip_restored(w, card, LEADER_BY, branch, from, to, reference)
                .map_err(|e| e.to_string())
        }
        None => record_tip_restored_cardless(w, LEADER_BY, branch, from, to, reference),
    }
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
            let nonce = new_nonce();
            let p = plan(project, reader.conn(), &branch, &tip, &nonce)?;
            let common = guard::common_dir_of(project).map_err(refused)?;
            guard::write_nonce(&common, &nonce).map_err(refused)?;
            out["plan"] = plan_json(&p, false);
            out["writes_file"] = json!(guard::NONCE_FILE);
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
    let common = guard::common_dir_of(project).map_err(refused)?;
    let Some(nonce) = guard::read_nonce(&common) else {
        return Err(refused(
            "no preview is waiting (the nonce of a preview is used once): run              `moai card drift --restore` again; nothing was changed",
        ));
    };
    let (id, mut writer) = open_for_write(project, home)?;
    let p = plan(project, writer.conn(), &branch, &tip, &nonce)?;
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
    // The digest matched: the nonce is spent whatever follows, and the one move that may be no
    // fast-forward is named by a restore intent that holds the recorded landing and its row.
    guard::remove_nonce(&common);
    guard::write_restore_intent(&common, &p.full, &p.from, &p.to, &p.to, p.row).map_err(refused)?;
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
    record_restore(&mut writer, p.card, &branch, &p.from, &p.to, reference).map_err(|e| {
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
        let base = digest_of("integ", "a", "b", Some(1), "n");
        assert_ne!(base, digest_of("other", "a", "b", Some(1), "n"));
        assert_ne!(base, digest_of("integ", "x", "b", Some(1), "n"));
        assert_ne!(base, digest_of("integ", "a", "y", Some(1), "n"));
        assert_ne!(base, digest_of("integ", "a", "b", Some(2), "n"));
        assert_ne!(base, digest_of("integ", "a", "b", None, "n"));
        assert_ne!(base, digest_of("integ", "a", "b", Some(1), "m"));
        assert_eq!(base, digest_of("integ", "a", "b", Some(1), "n"));
    }

    // ---- land-sync rows (SPEC-20261125-land-sync-merge REQ-6, REQ-9, D-4) -------------------

    use crate::ledger::schema;

    const BRANCH: &str = "integ";

    fn h(c: char) -> String {
        c.to_string().repeat(40)
    }

    /// An in-memory ledger with two cards.
    fn ledger() -> Connection {
        let mut conn = Connection::open_in_memory().unwrap();
        schema::create_v1(&mut conn, "unit", "test").unwrap();
        conn.execute(
            "INSERT INTO card (card_id, title, created_by, created_at, updated_at) \
             VALUES (1, 'a', 'x', 1, 1), (2, 'b', 'x', 1, 1)",
            [],
        )
        .unwrap();
        conn
    }

    /// One event row, returning its id.
    fn put(conn: &Connection, kind: &str, card: Option<i64>, payload: Value) -> i64 {
        let key = ledger::new_event_key(conn, kind, card).unwrap();
        conn.execute(
            "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
             VALUES (?1, 1, 'x', ?2, ?3, ?4)",
            params![key, card, kind, payload.to_string()],
        )
        .unwrap();
        conn.last_insert_rowid()
    }

    /// A landing of `head` from `from`: its start row and, when `done`, its landed row.
    fn landing(
        conn: &Connection,
        card: i64,
        from: &str,
        head: &str,
        done: bool,
    ) -> (i64, Option<i64>) {
        let payload = json!({"branch": BRANCH, "tip": from, "head": head, "by": "leader"});
        let start = put(conn, "card.land_started", Some(card), payload);
        let landed = done.then(|| put(conn, "card.landed", Some(card), json!({"head": head})));
        (start, landed)
    }

    fn sync_row(old: &str, merge: &str) -> SyncRow {
        SyncRow {
            branch: BRANCH.to_string(),
            merge: merge.to_string(),
            from: "hotfix".to_string(),
            from_tip: h('9'),
            old_tip: old.to_string(),
            spec: "SPEC-20261117-sync-fixture".to_string(),
        }
    }

    fn sync_payload(row: &SyncRow) -> Value {
        json!({
            "branch": row.branch, "merge": row.merge, "from": row.from,
            "from_tip": row.from_tip, "old_tip": row.old_tip, "spec": row.spec, "by": "leader",
        })
    }

    /// A sync of `merge` on `old`: its start row and, when `done`, its synced row, both written
    /// by hand with no card.
    fn syncing(conn: &Connection, old: &str, merge: &str, done: bool) -> (i64, Option<i64>) {
        let payload = sync_payload(&sync_row(old, merge));
        let start = put(conn, SYNC_STARTED, None, payload.clone());
        let synced = done.then(|| put(conn, SYNCED, None, payload));
        (start, synced)
    }

    fn want(conn: &Connection) -> Expected {
        expected(conn, BRANCH).unwrap().expect("a recorded move")
    }

    /// A ledger the real writers use: a temporary HOME and a registered project.
    fn write_ledger() -> (tempfile::TempDir, WriteLedger) {
        let home = tempfile::tempdir().unwrap();
        ledger::register(home.path(), "demo").unwrap();
        let w = ledger::open_write(home.path(), "demo").unwrap();
        (home, w)
    }

    /// The rows of `kind`: id, card and payload.
    fn events(conn: &Connection, kind: &str) -> Vec<(i64, Option<i64>, Value)> {
        let mut stmt = conn
            .prepare(
                "SELECT event_id, card_id, payload FROM event WHERE kind = ?1 ORDER BY event_id",
            )
            .unwrap();
        let rows = stmt
            .query_map([kind], |r| {
                let text: String = r.get(2)?;
                Ok((r.get(0)?, r.get(1)?, serde_json::from_str(&text).unwrap()))
            })
            .unwrap()
            .collect::<std::result::Result<Vec<_>, _>>()
            .unwrap();
        rows
    }

    /// REQ-9: a complete sync (start and synced row) expects its merge commit, whether it follows
    /// a landing or is the only move the ledger knows; the landing's own tip is drift after it.
    #[test]
    fn a_complete_sync_is_the_expected_tip() {
        let conn = ledger();
        landing(&conn, 1, &h('a'), &h('b'), true);
        let (_, synced) = syncing(&conn, &h('b'), &h('c'), true);
        let got = want(&conn);
        assert_eq!(got.commit, h('c'));
        assert_eq!(got.card, None);
        assert_eq!(got.source, "sync");
        assert_eq!(got.row, synced.unwrap());
        assert_eq!(got.accepted, vec![h('c')]);
        assert_eq!(refusal(&conn, BRANCH, &h('c')).unwrap(), None);
        let why = refusal(&conn, BRANCH, &h('b')).unwrap().expect("drift");
        assert!(why.starts_with("drift: "), "{why}");
        assert!(why.contains("a `moai card land-sync` merge"), "{why}");
        assert!(why.contains(&format!("left it at {}", h('c'))), "{why}");

        let alone = ledger();
        let (_, synced) = syncing(&alone, &h('b'), &h('c'), true);
        let got = want(&alone);
        assert_eq!(
            (got.commit, got.source, got.row),
            (h('c'), "sync", synced.unwrap())
        );
        assert!(refusal(&alone, BRANCH, &h('b')).unwrap().is_some());
    }

    /// REQ-9: an interrupted sync (start row, no synced row) accepts the merge (the move was done)
    /// and the old tip (it was not), and nothing else.
    #[test]
    fn an_interrupted_sync_accepts_the_merge_and_the_old_tip() {
        let conn = ledger();
        landing(&conn, 1, &h('a'), &h('b'), true);
        syncing(&conn, &h('b'), &h('c'), false);
        let got = want(&conn);
        assert_eq!(got.accepted, vec![h('b'), h('c')]);
        assert_eq!((got.commit.as_str(), got.card), (h('b').as_str(), Some(1)));
        assert_eq!(refusal(&conn, BRANCH, &h('c')).unwrap(), None);
        assert_eq!(refusal(&conn, BRANCH, &h('b')).unwrap(), None);
        assert!(refusal(&conn, BRANCH, &h('0')).unwrap().is_some());
        assert!(refusal(&conn, BRANCH, &h('a')).unwrap().is_some());

        // The first move of the branch, interrupted: it restores to the tip it started from.
        let first = ledger();
        let (start, _) = syncing(&first, &h('b'), &h('c'), false);
        let got = want(&first);
        assert_eq!(got.commit, h('b'));
        assert_eq!(got.card, None);
        assert_eq!(got.source, "interrupted_sync");
        assert_eq!(got.row, start);
        assert_eq!(got.accepted, vec![h('b'), h('c')]);
        assert!(refusal(&first, BRANCH, &h('0')).unwrap().is_some());
    }

    /// REQ-9: the newer of a landing and a sync decides, in both directions, and what the older
    /// one expected is drift once the newer one is complete.
    #[test]
    fn the_newer_of_a_landing_and_a_sync_decides() {
        let sync_then_landing = ledger();
        syncing(&sync_then_landing, &h('b'), &h('c'), true);
        let (_, landed) = landing(&sync_then_landing, 2, &h('c'), &h('d'), true);
        let got = want(&sync_then_landing);
        assert_eq!(got.commit, h('d'));
        assert_eq!((got.card, got.source), (Some(2), "landing"));
        assert_eq!(got.row, landed.unwrap());
        assert_eq!(got.accepted, vec![h('d')]);
        assert!(refusal(&sync_then_landing, BRANCH, &h('c'))
            .unwrap()
            .is_some());

        let landing_then_sync = ledger();
        landing(&landing_then_sync, 1, &h('a'), &h('b'), true);
        syncing(&landing_then_sync, &h('b'), &h('c'), true);
        let got = want(&landing_then_sync);
        assert_eq!((got.commit.as_str(), got.source), (h('c').as_str(), "sync"));
        assert!(refusal(&landing_then_sync, BRANCH, &h('b'))
            .unwrap()
            .is_some());

        // An interrupted landing after a complete sync accepts the merge and its own two tips.
        let interrupted = ledger();
        syncing(&interrupted, &h('b'), &h('c'), true);
        landing(&interrupted, 2, &h('c'), &h('d'), false);
        let got = want(&interrupted);
        assert_eq!((got.commit.as_str(), got.source), (h('c').as_str(), "sync"));
        assert_eq!(got.accepted, vec![h('c'), h('d')]);
    }

    /// REQ-9: a landing that follows a landing is refused with the same bytes as before the sync
    /// rows existed, and its digest is the digest of the same move with the card spelled as before.
    #[test]
    fn a_landing_after_a_landing_keeps_its_refusal_text_and_digest() {
        let conn = ledger();
        landing(&conn, 1, &h('a'), &h('b'), true);
        landing(&conn, 2, &h('b'), &h('d'), true);
        let tip = h('e');
        let why = refusal(&conn, BRANCH, &tip).unwrap().expect("drift");
        let pinned = format!(
            "drift: the integration branch integ points at {tip}, but the newest landing the \
             ledger recorded (C-2) left it at {d}: something moved the branch outside `moai card land` \
             (for example `git branch -C`, which no hook sees, or a commit added by hand, which is \
             no landing). `moai card drift` shows the comparison; after the \
             user's pick, `moai card drift --restore` previews the move back to {d} and applies it only with \
             the user's yes; then land again",
            d = h('d'),
        );
        assert_eq!(why, pinned);
        let got = want(&conn);
        assert_eq!((got.card, got.source), (Some(2), "landing"));
        let literal = format!(
            "{:x}",
            Sha256::digest("moai card drift --restore\ninteg\na\nb\n1\nn\n".as_bytes())
        );
        assert_eq!(digest_of("integ", "a", "b", Some(1), "n"), literal);
        assert_ne!(digest_of("integ", "a", "b", None, "n"), literal);
    }

    /// REQ-6: rows with no card never reach the landing query: a ledger that holds card-less
    /// rows of every kind the sync writes, a `claude.run` row and sync rows that are malformed or
    /// belong to another branch still answers, and only the sync rows of this branch count.
    #[test]
    fn card_less_rows_never_make_the_landing_query_error() {
        let conn = ledger();
        assert_eq!(expected(&conn, BRANCH).unwrap(), None);
        put(&conn, "claude.run", None, json!({"verdict": "ok"}));
        let restored = json!({"branch": BRANCH, "from": h('7'), "to": h('8')});
        put(&conn, TIP_RESTORED, None, restored);
        put(&conn, SYNC_STARTED, None, json!({"branch": BRANCH}));
        let numeric = json!({"branch": BRANCH, "merge": 7, "old_tip": h('b')});
        put(&conn, SYNC_STARTED, None, numeric);
        let mut other = sync_payload(&sync_row(&h('b'), &h('c')));
        other["branch"] = json!("other");
        put(&conn, SYNC_STARTED, None, other.clone());
        put(&conn, SYNCED, None, other);
        assert_eq!(expected(&conn, BRANCH).unwrap(), None);
        assert_eq!(refusal(&conn, BRANCH, &h('0')).unwrap(), None);
        landing(&conn, 1, &h('a'), &h('b'), true);
        let got = want(&conn);
        assert_eq!(
            (got.commit.as_str(), got.source),
            (h('b').as_str(), "landing")
        );
        // A synced row of another merge closes no start row.
        syncing(&conn, &h('b'), &h('c'), false);
        put(
            &conn,
            SYNCED,
            None,
            sync_payload(&sync_row(&h('b'), &h('f'))),
        );
        assert_eq!(want(&conn).accepted, vec![h('b'), h('c')]);
    }

    /// REQ-6: the writers add one card-less row each, the start row once per branch, merge and
    /// old tip (a rerun adds none), the synced row once and only after its start row, and a value
    /// that is no commit id is refused with nothing written.
    #[test]
    fn the_sync_writers_write_card_less_rows_once() {
        let (_home, mut w) = write_ledger();
        let row = sync_row(&h('b'), &h('c'));
        let err = record_synced(&mut w, &row).unwrap_err();
        assert!(err.contains(SYNC_STARTED), "{err}");
        assert!(events(w.conn(), SYNCED).is_empty());
        for bad in [
            SyncRow {
                merge: "abc".into(),
                ..row.clone()
            },
            SyncRow {
                old_tip: h('B'),
                ..row.clone()
            },
            SyncRow {
                from_tip: String::new(),
                ..row.clone()
            },
            SyncRow {
                branch: String::new(),
                ..row.clone()
            },
            SyncRow {
                spec: String::new(),
                ..row.clone()
            },
        ] {
            assert!(record_sync_started(&mut w, &bad).is_err(), "{bad:?}");
        }
        assert!(events(w.conn(), SYNC_STARTED).is_empty());

        record_sync_started(&mut w, &row).unwrap();
        record_sync_started(&mut w, &row).unwrap();
        let started = events(w.conn(), SYNC_STARTED);
        assert_eq!(started.len(), 1);
        assert_eq!(started[0].1, None);
        assert_eq!(started[0].2, sync_payload(&row));
        // The move may have happened: both tips are accepted until the close row is written.
        assert_eq!(want(w.conn()).accepted, vec![h('b'), h('c')]);

        record_synced(&mut w, &row).unwrap();
        record_synced(&mut w, &row).unwrap();
        let synced = events(w.conn(), SYNCED);
        assert_eq!(synced.len(), 1);
        assert_eq!(synced[0].1, None);
        assert_eq!(synced[0].2, sync_payload(&row));
        assert!(started[0].0 < synced[0].0);
        let got = want(w.conn());
        assert_eq!((got.commit.as_str(), got.source), (h('c').as_str(), "sync"));
        assert_eq!(got.row, synced[0].0);
        assert_eq!(got.accepted, vec![h('c')]);

        // A second sync of the same merge on another old tip is a sync of its own.
        record_sync_started(&mut w, &sync_row(&h('c'), &h('c'))).unwrap();
        assert_eq!(events(w.conn(), SYNC_STARTED).len(), 2);
    }

    /// D-4: the restore row of a sync-sourced expectation has no card and the payload of the
    /// card-bound row; a landing's restore row stays on its card.
    #[test]
    fn a_restore_from_a_sync_records_a_card_less_row() {
        let (_home, mut w) = write_ledger();
        let row = sync_row(&h('b'), &h('c'));
        record_sync_started(&mut w, &row).unwrap();
        record_synced(&mut w, &row).unwrap();
        let got = want(w.conn());
        assert_eq!(got.card, None);
        record_restore(
            &mut w,
            got.card,
            BRANCH,
            &h('0'),
            &got.commit,
            "user-pick-1",
        )
        .unwrap();
        let rows = events(w.conn(), TIP_RESTORED);
        assert_eq!(rows.len(), 1);
        assert_eq!(rows[0].1, None);
        assert_eq!(
            rows[0].2,
            json!({
                "branch": BRANCH, "from": h('0'), "to": h('c'),
                "reference": "user-pick-1", "by": LEADER_BY,
            })
        );

        let card = w
            .with_write_tx(|tx, now| {
                tx.execute(
                    "INSERT INTO card (title, state, created_by, created_at, updated_at) \
                     VALUES ('card', 'ready', 'test', ?1, ?1)",
                    [now],
                )?;
                Ok(tx.last_insert_rowid())
            })
            .unwrap();
        record_restore(&mut w, Some(card), BRANCH, &h('0'), &h('b'), "user-pick-2").unwrap();
        let rows = events(w.conn(), TIP_RESTORED);
        assert_eq!(rows.len(), 2);
        assert_eq!(rows[1].1, Some(card));
        assert_eq!(rows[1].2["to"], json!(h('b')));
    }

    /// The compare view of a sync-sourced expectation names no card and the `sync` source.
    #[test]
    fn the_compare_view_of_a_sync_names_no_card() {
        let conn = ledger();
        syncing(&conn, &h('b'), &h('c'), true);
        let ok = compare_json(BRANCH, &h('c'), &expected(&conn, BRANCH).unwrap());
        assert_eq!(ok["state"], json!("ok"));
        assert_eq!(ok["expected"]["card"], Value::Null);
        assert_eq!(ok["expected"]["source"], json!("sync"));
        let drift = compare_json(BRANCH, &h('b'), &expected(&conn, BRANCH).unwrap());
        assert_eq!(drift["state"], json!("drift"));
        assert_eq!(drift["restore_command"], json!(RESTORE_COMMAND));
    }
}
