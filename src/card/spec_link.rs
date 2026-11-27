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
    link_at(project, home, reference, spec, by, None)
}

/// [`link`] with the `--generation` the caller read (card C-608): given, the card's live
/// reservation must be at it (`work::check_generation`, in the same write transaction) and
/// nothing is written otherwise; left out, nothing is checked. The link stays unfenced: neither
/// the holder nor the lease is asked.
pub fn link_at(
    project: &Path,
    home: &Path,
    reference: &str,
    spec: Option<&str>,
    by: Option<&str>,
    generation: Option<i64>,
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
            if let Some(want) = generation {
                if let Err(refused) =
                    super::work::check_generation(tx, card_id, want, "linking a SPEC")
                {
                    return Ok(Err(refused));
                }
            }
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

/// The `relation` a `card.spec` event payload carries for a follow-up card
/// (SPEC-20261124-followup-per-spec).
pub const RELATION_FOLLOW_UP: &str = "follow-up";
/// The `relation` a `card.spec` event payload carries for a card split from another card; the
/// payload's `of` names that card.
pub const RELATION_SPLIT_OF: &str = "split-of";

/// How a card added with `--spec` stands to the other cards of its SPEC.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Relation<'a> {
    /// No relation flag: the SPEC may have no other open card.
    Plain,
    /// `--follow-up`: a follow-up card; a SPEC may have any number of them open at once.
    FollowUp,
    /// `--follow-up-of <card>`: a follow-up card that names the card it follows; that card names
    /// the same SPEC. Never refused for open cards, like `--follow-up`.
    FollowUpOf(&'a str),
    /// `--split-of <card>`: that card names the same SPEC.
    SplitOf(&'a str),
}

impl Relation<'_> {
    /// The `relation` string the `card.spec` payload carries, `None` for a plain add.
    pub fn name(&self) -> Option<&'static str> {
        match self {
            Relation::Plain => None,
            Relation::FollowUp | Relation::FollowUpOf(_) => Some(RELATION_FOLLOW_UP),
            Relation::SplitOf(_) => Some(RELATION_SPLIT_OF),
        }
    }

    /// The card the payload's `of` names, `None` when the relation names none.
    pub fn origin(&self) -> Option<&str> {
        match self {
            Relation::FollowUpOf(o) | Relation::SplitOf(o) => Some(o),
            Relation::Plain | Relation::FollowUp => None,
        }
    }
}

/// What `card show` prints as `spec_link`: the `relation` and `of` of the card's newest
/// `card.spec` event, as `{"relation": <string|null>, "of": <string|null>}`; `null` when the card
/// names no SPEC (no such event, or the newest one unlinked it). A link made by `card spec` has
/// no relation, so both fields are `null`.
pub fn link_json(conn: &Connection, card_id: i64) -> rusqlite::Result<Value> {
    let newest: Option<String> = conn
        .query_row(
            "SELECT payload FROM event WHERE card_id = ?1 AND kind = 'card.spec' \
             ORDER BY event_id DESC LIMIT 1",
            [card_id],
            |r| r.get(0),
        )
        .optional()?;
    let Some(text) = newest else {
        return Ok(Value::Null);
    };
    let payload: Value = serde_json::from_str(&text).unwrap_or(Value::Null);
    if payload["to"].is_null() {
        return Ok(Value::Null);
    }
    let text = |key: &str| payload[key].as_str().map_or(Value::Null, |s| json!(s));
    Ok(json!({"relation": text("relation"), "of": text("of")}))
}

/// A card that names a SPEC and is not `done`, with the relation its newest link carries.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct OpenCard {
    pub card_id: i64,
    pub relation: Option<String>,
}

/// The cards that name `spec_id` now and are open: not `done` and not `hold` (a withdrawn or a
/// folded card is on hold and leaves its SPEC room; draft, ready, active and review count), by
/// number.
pub fn open_cards_of(conn: &Connection, spec_id: &str) -> rusqlite::Result<Vec<OpenCard>> {
    let mut stmt = conn.prepare(&format!(
        "SELECT card_id, (SELECT json_extract(e.payload, '$.relation') FROM event e \
         WHERE e.card_id = card.card_id AND e.kind = 'card.spec' \
         ORDER BY e.event_id DESC LIMIT 1) \
         FROM card WHERE state NOT IN ('done', 'hold') AND {LINKED_COLUMN} = ?1 ORDER BY card_id"
    ))?;
    let rows = stmt.query_map([spec_id], |r| {
        Ok(OpenCard {
            card_id: r.get(0)?,
            relation: r.get(1)?,
        })
    })?;
    rows.collect()
}

/// Why `card add --spec` must refuse a card of `relation` on SPEC `spec_id` when `open` are the
/// SPEC's cards that are not `done`, or `None`. The message names each open card and the way out;
/// its `conflict in ` prefix gives it the `conflict` kind (`contract::classify`).
pub fn add_conflict(spec_id: &str, relation: Relation<'_>, open: &[OpenCard]) -> Option<String> {
    let name = |c: &OpenCard| input::card_name(c.card_id);
    match relation {
        Relation::SplitOf(_) | Relation::FollowUp | Relation::FollowUpOf(_) => None,
        Relation::Plain if open.is_empty() => None,
        Relation::Plain => {
            let names: Vec<String> = open.iter().map(name).collect();
            Some(format!(
                "card add conflict in spec: {spec_id} already has an open card ({}); a plain add \
                 needs a SPEC with none. Add `--split-of <card>` for a card split from it, \
                 `--follow-up` or `--follow-up-of <card>` for a follow-up card, or add the work \
                 as a criterion of {} with `moai card criterion-add {} --ac <text>`",
                names.join(", "),
                names[0],
                names[0],
            ))
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn open(card_id: i64, relation: Option<&str>) -> OpenCard {
        OpenCard {
            card_id,
            relation: relation.map(str::to_string),
        }
    }

    #[test]
    fn a_plain_add_needs_a_spec_with_no_open_card() {
        let id = "SPEC-20261124-x";
        assert_eq!(add_conflict(id, Relation::Plain, &[]), None);
        let m = add_conflict(
            id,
            Relation::Plain,
            &[open(3, None), open(5, Some("follow-up"))],
        )
        .unwrap();
        for word in [
            "C-3, C-5",
            "--follow-up",
            "--split-of",
            "criterion-add C-3 --ac",
        ] {
            assert!(m.contains(word), "{word}: {m}");
        }
    }

    #[test]
    fn a_follow_up_is_never_blocked_by_open_cards() {
        let id = "SPEC-20261124-x";
        assert_eq!(add_conflict(id, Relation::FollowUp, &[]), None);
        let all = [
            open(3, None),
            open(4, Some("split-of")),
            open(5, Some("follow-up")),
            open(6, Some("follow-up")),
        ];
        for relation in [Relation::FollowUp, Relation::FollowUpOf("C-5")] {
            assert_eq!(add_conflict(id, relation, &all), None, "{relation:?}");
        }
    }

    #[test]
    fn a_plain_add_names_every_open_follow_up_and_the_ways_out() {
        let id = "SPEC-20261124-x";
        let all = [
            open(5, Some("follow-up")),
            open(6, Some("follow-up")),
            open(7, Some("follow-up")),
        ];
        let m = add_conflict(id, Relation::Plain, &all).unwrap();
        for word in [
            "C-5, C-6, C-7",
            "--follow-up-of <card>",
            "--split-of",
            "criterion-add C-5 --ac",
        ] {
            assert!(m.contains(word), "{word}: {m}");
        }
        assert!(!m.contains("one follow-up"), "{m}");
    }

    #[test]
    fn a_split_is_never_blocked_by_open_cards() {
        let id = "SPEC-20261124-x";
        let all = [open(3, None), open(4, Some("follow-up"))];
        assert_eq!(add_conflict(id, Relation::SplitOf("C-3"), &all), None);
    }

    #[test]
    fn a_relation_names_its_payload_fields() {
        assert_eq!(Relation::Plain.name(), None);
        assert_eq!(Relation::FollowUp.name(), Some(RELATION_FOLLOW_UP));
        assert_eq!(Relation::FollowUpOf("C-2").name(), Some(RELATION_FOLLOW_UP));
        assert_eq!(Relation::SplitOf("C-2").name(), Some(RELATION_SPLIT_OF));
        assert_eq!(Relation::Plain.origin(), None);
        assert_eq!(Relation::FollowUp.origin(), None);
        assert_eq!(Relation::FollowUpOf("C-2").origin(), Some("C-2"));
        assert_eq!(Relation::SplitOf("C-3").origin(), Some("C-3"));
    }

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
