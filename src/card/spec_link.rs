//! The link from a card to the SPEC it works on (SDD-4, SPEC-20261117-card-spec-link).
//!
//! A card is a personal todo in this machine's ledger; a SPEC is a shared, git-tracked file
//! under `<project>/.moai/specs` (user decisions U-A..U-D, 2026-11-17). So the link lives on the
//! card side only: the SPEC never names a card, and the ledger keeps the link as `card.spec`
//! events (payload `{"from": <old id or null>, "to": <new id or null>}`). The card's SPEC is the
//! `to` of its newest `card.spec` event. That needs no new column: a ledger of schema version 1
//! takes the link as it is, and `card export` carries it with the other events.
//!
//! `card spec` writes the link after the SPEC is found and parses without an error. `card next`
//! and `card take` print the linked SPEC's status ([`status`]) and warn when a card that changes
//! the repository has no approved SPEC; `take --require-approved-spec` refuses instead. The hard
//! gate (no merge without an approved SPEC) is a later slice.

use super::command::{actor, open_for_write};
use super::input::{self, CardRef};
use crate::ledger;
use crate::spec_doc;
use rusqlite::{params, Connection, OptionalExtension};
use serde_json::{json, Value};
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// The event kind that records a link or an unlink.
pub const EVENT: &str = "card.spec";

/// The card kinds whose work changes nothing in the repository (decision U-A: research and
/// questions use a card only). Every other kind changes the repository and needs a SPEC.
pub const NO_REPO_CHANGE: [&str; 2] = ["research", "audit"];

/// Whether a card of `kind` changes the repository.
pub fn changes_repository(kind: &str) -> bool {
    !NO_REPO_CHANGE.contains(&kind)
}

/// The SPEC id a card names now: the `to` of its newest `card.spec` event, as a column of a
/// query over `card`.
pub const LINKED_COLUMN: &str = "(SELECT json_extract(e.payload, '$.to') FROM event e \
     WHERE e.card_id = card.card_id AND e.kind = 'card.spec' \
     ORDER BY e.event_id DESC LIMIT 1)";

/// The card's kind and the SPEC id it names; `None` when there is no such card.
pub fn kind_and_spec(
    conn: &Connection,
    card_id: i64,
) -> rusqlite::Result<Option<(String, Option<String>)>> {
    conn.query_row(
        &format!("SELECT kind, {LINKED_COLUMN} FROM card WHERE card_id = ?1"),
        [card_id],
        |r| Ok((r.get(0)?, r.get(1)?)),
    )
    .optional()
}

/// The SPEC id each card that is not `done` names now, as (card number, SPEC id) in card order;
/// a card that names none is left out.
pub fn linked_specs(conn: &Connection) -> rusqlite::Result<Vec<(i64, String)>> {
    let mut stmt = conn.prepare(&format!(
        "SELECT card_id, {LINKED_COLUMN} FROM card WHERE state <> 'done' ORDER BY card_id"
    ))?;
    let rows = stmt.query_map([], |r| {
        Ok((r.get::<_, i64>(0)?, r.get::<_, Option<String>>(1)?))
    })?;
    let mut out = Vec::new();
    for row in rows {
        let (card_id, spec) = row?;
        out.extend(spec.map(|s| (card_id, s)));
    }
    Ok(out)
}

/// The cards that name `spec_id` now, by number: name, title, kind and state.
pub fn cards_of(conn: &Connection, spec_id: &str) -> rusqlite::Result<Vec<Value>> {
    let mut stmt = conn.prepare(&format!(
        "SELECT card_id, title, kind, state FROM card WHERE {LINKED_COLUMN} = ?1 ORDER BY card_id"
    ))?;
    let rows = stmt.query_map([spec_id], |r| {
        Ok(json!({
            "name": input::card_name(r.get(0)?),
            "card_id": r.get::<_, i64>(0)?,
            "title": r.get::<_, String>(1)?,
            "kind": r.get::<_, String>(2)?,
            "state": r.get::<_, String>(3)?,
        }))
    })?;
    rows.collect()
}

/// Why `card done` must refuse card `card` (`C-n`) that names SPEC `spec`, or `None` (card C-86
/// slice 2, SPEC-20261117-ac-verify-required): the SPEC cannot be read, fails `spec check`, or
/// has an acceptance criterion with no `verify:` command or `manual:` check. A card that names
/// no SPEC is not judged here.
pub fn done_refusal(project: &Path, card: &str, spec: Option<&str>) -> Option<String> {
    let id = spec?;
    match spec_doc::find(project, id) {
        Err(e) => Some(format!("{card}: its SPEC {id} cannot be read: {e}")),
        Ok(doc) => judge(card, id, &doc),
    }
}

/// [`done_refusal`] for a card whose landing git proved: the SPEC is read from the integration
/// branch tip, not from the working tree, which a checkout that was not moved after the
/// fast-forward leaves stale. A SPEC the tip does not hold is refused.
pub fn done_refusal_at_tip(
    project: &Path,
    card: &str,
    spec: Option<&str>,
    tip: &str,
) -> Option<String> {
    let id = spec?;
    if spec_doc::id_form(id).is_none() {
        return Some(format!("{card}: its SPEC {id} is not a SPEC id"));
    }
    let path = format!("{}/{id}/spec.md", spec_doc::SPECS_DIR);
    match super::landed::show_at(project, tip, &path) {
        Err(e) => Some(format!(
            "{card}: its SPEC {id} cannot be read from the tip: {e}"
        )),
        Ok(None) => Some(format!(
            "{card}: its SPEC {id} is not in the integration branch tip {tip} ({path} is absent \
             there)"
        )),
        Ok(Some(text)) => {
            let opts = spec_doc::Options {
                folder: Some(id.to_string()),
                closure: false,
            };
            judge(card, id, &spec_doc::parse(&text, &opts))
        }
    }
}

/// What `card done` asks of a SPEC that was read: no error finding and a check for every
/// acceptance criterion.
fn judge(card: &str, id: &str, doc: &spec_doc::SpecDoc) -> Option<String> {
    if doc.errors() > 0 {
        return Some(format!(
            "{card}: its SPEC {id} has {} error(s); `moai spec check {id}` lists them",
            doc.errors()
        ));
    }
    let open = spec_doc::unverified(doc);
    (!open.is_empty()).then(|| {
        format!(
            "{card}: its SPEC {id} has no `verify:` command or `manual:` check for {}",
            open.join(", ")
        )
    })
}

/// What `next` and `take` say about the SPEC of card `card` (`C-n`) of `kind`: the SPEC's state
/// (`null` when the card names none) and, for a card that changes the repository, a warning
/// when that SPEC is not approved. Approved means a SPEC of this generation (`SPEC-<yyyymmdd>-
/// <slug>`) with status `approved` or `completed`, a recorded `approval` (`by`, `at`, `basis`,
/// `digest`) and no error finding (decision U-C). A Go-generation SPEC is never approved here,
/// and a SPEC that is gone or cannot be read is not approved.
pub fn status(
    project: &Path,
    card: &str,
    kind: &str,
    spec: Option<&str>,
) -> (Value, Option<String>) {
    let Some(id) = spec else {
        let warning = changes_repository(kind).then(|| {
            format!(
                "{card}: changes the repository (kind `{kind}`) and names no SPEC; link an \
                 approved one with `moai card spec {card} <SPEC-ID>`"
            )
        });
        return (Value::Null, warning);
    };
    let (state, problem) = match spec_doc::find(project, id) {
        Ok(doc) => {
            let status = doc.status_as.clone().or_else(|| doc.status.clone());
            let errors = doc.errors();
            // Decision U-C: approval is the user's selection recorded in the SPEC. A
            // Go-generation SPEC has no such record, whatever its status says.
            let go_generation = doc.id_form != Some(spec_doc::IdForm::New) || doc.status_legacy;
            let recorded = doc.approval.as_ref().is_some_and(|a| {
                a.by.is_some() && a.at.is_some() && a.basis.is_some() && a.digest.is_some()
            });
            let approved_word = matches!(doc.status.as_deref(), Some("approved" | "completed"));
            let approved = !go_generation && recorded && approved_word && errors == 0;
            let problem = if go_generation {
                Some(format!(
                    "{card}: its SPEC {id} is a Go-generation SPEC; record an approval to use it \
                     (a SPEC-<yyyymmdd>-<slug> SPEC with `approval`)"
                ))
            } else if errors > 0 {
                Some(format!(
                    "{card}: its SPEC {id} has {errors} error(s); `moai spec check {id}` lists them"
                ))
            } else if !approved_word {
                Some(format!(
                    "{card}: its SPEC {id} is `{}`, not approved",
                    status.as_deref().unwrap_or("unknown")
                ))
            } else if !recorded {
                Some(format!(
                    "{card}: its SPEC {id} has no recorded approval (by, at, basis, digest)"
                ))
            } else {
                None
            };
            (
                json!({"id": id, "found": true, "status": status, "errors": errors, "approved": approved}),
                problem,
            )
        }
        Err(e) => (
            json!({"id": id, "found": false, "status": null, "errors": null, "approved": false}),
            Some(format!("{card}: its SPEC {id} cannot be read: {e}")),
        ),
    };
    (state, problem.filter(|_| changes_repository(kind)))
}

/// `card spec C-n SPEC-ID` links the card to the SPEC, `card spec C-n --unlink` (`spec: None`)
/// removes the link. The SPEC must be in this project's `.moai/specs` and parse without an
/// error; that is checked before the ledger is opened. One transaction writes one `card.spec`
/// event and touches the card; a link that is already so writes nothing (`changed: false`).
pub fn link(
    project: &Path,
    home: &Path,
    reference: &str,
    spec: Option<&str>,
    by: Option<&str>,
) -> Result<Value> {
    let card_id =
        match input::parse_ref(reference).map_err(|e| format!("card spec refused: {e}"))? {
            CardRef::Id(n) => n,
            CardRef::SourceNumber(_) => {
                return Err(
                    "card spec refused: name the card by its C-number (`card show #N` finds it)"
                        .to_string(),
                )
            }
        };
    if let Some(id) = spec {
        if spec_doc::id_form(id).is_none() {
            return Err(
                "card spec refused: the SPEC id is not SPEC-<yyyymmdd>-<slug> (or a \
                        Go-generation SPEC-…-NNN id)"
                    .to_string(),
            );
        }
        let doc = spec_doc::find(project, id)?;
        let errors = doc.errors();
        if errors > 0 {
            return Err(format!(
                "card spec refused: {id} has {errors} error(s); `moai spec check {id}` lists \
                 them; nothing was written"
            ));
        }
    }
    let by = actor(by, "spec")?;
    let (pid, mut ledger) = open_for_write(project, home)?;
    let (from, changed, rev) = ledger
        .with_write_tx(|tx, now| {
            let Some((_, from)) = kind_and_spec(tx, card_id)? else {
                return Ok(Err(format!("card not found: there is no C-{card_id}")));
            };
            let rev = |tx: &Connection| {
                tx.query_row("SELECT rev FROM card WHERE card_id = ?1", [card_id], |r| {
                    r.get::<_, i64>(0)
                })
            };
            if from.as_deref() == spec {
                return Ok(Ok((from, false, rev(tx)?)));
            }
            tx.execute(
                "UPDATE card SET updated_at = ?2 WHERE card_id = ?1",
                params![card_id, now],
            )?;
            let key = ledger::new_event_key(tx, EVENT, Some(card_id))?;
            tx.execute(
                "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
                 VALUES (?1, ?2, ?3, ?4, 'card.spec', ?5)",
                params![
                    key,
                    now,
                    by,
                    card_id,
                    json!({"from": from, "to": spec}).to_string()
                ],
            )?;
            Ok(Ok((from, true, rev(tx)?)))
        })
        .map_err(|e| e.to_string())??;
    Ok(json!({
        "action": "spec", "project_id": pid,
        "name": input::card_name(card_id), "card_id": card_id,
        "from": from, "spec": spec, "changed": changed, "by": by, "rev": rev,
        "events": i32::from(changed), "writes": changed,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn only_research_and_audit_cards_leave_the_repository_alone() {
        for kind in ["task", "feature", "fix", "docs", "chore"] {
            assert!(changes_repository(kind), "{kind}");
        }
        for kind in ["research", "audit"] {
            assert!(!changes_repository(kind), "{kind}");
        }
    }
}
