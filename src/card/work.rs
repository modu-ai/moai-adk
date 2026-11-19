//! What `moai card next | note | criterion | pr | review | reopen | done | renew | check` do (card C-53
//! S1): a lane finds the next card, works it and closes it, with no GitHub Issue.
//!
//! The rules are the ledger's: `ledger::workflow` for the finish path and `ledger::reservation`
//! for the lease and the fence. This module checks what a user typed, opens the ledger and
//! prints what the ledger returned. `next` and `check` open it read-only; the others write it,
//! and none creates it.
//!
//! As in `dispatch`, the holder is never typed: a fenced action acts as **this session**
//! (`session_holder`) and needs the generation `take` or `handoff` answered with. A stale
//! generation or another session is refused with the ledger's own sentence (`conflict`). A note
//! is not fenced: a decision is answered by someone who does not hold the card.
//!
//! `pr` is the one action that reads a code host, and it does so here, through a
//! [`HostConnector`] the caller hands in: `src/ledger` never reaches a host. The holder is
//! checked before the host is asked, so a refused caller causes no request.

use super::command::{actor, open_for_write, ABSENT_HINT};
use super::dispatch::{reservation_json, DEFAULT_LEASE_SECS};
use super::input::{self, CardRef};
use super::spec_link;
use crate::forge::connector::{HostConnector, HostKind, PrPhase, ProjectRef};
use crate::ledger::reservation;
use crate::ledger::workflow::{self, CriterionChange, PrTarget, SkipReason, WorkflowError};
use crate::ledger::{self, LedgerError};
use serde_json::{json, Value};
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// The card's number; a host number (`#12`) needs a lookup, so these actions refuse it.
fn card_id(action: &str, reference: &str) -> Result<i64> {
    match input::parse_ref(reference).map_err(|e| format!("card {action} refused: {e}"))? {
        CardRef::Id(n) => Ok(n),
        CardRef::SourceNumber(_) => Err(format!(
            "card {action} refused: name the card by its C-number (`card show #N` finds it)"
        )),
    }
}

/// A ledger refusal as the line the envelope classifies: a rejected value is
/// `card <action> refused: …`, everything else keeps the ledger's own sentence.
fn refusal(action: &str, error: WorkflowError) -> String {
    match error {
        WorkflowError::InvalidInput(why) => format!("card {action} refused: {why}"),
        other => other.to_string(),
    }
}

/// `card next`: the cards this session holds, then the `ready` cards by priority and number,
/// then every card that cannot be taken with the reason. Each held and ready card carries the
/// SPEC it names and that SPEC's status (`spec`), and `warnings` names each card that changes
/// the repository without an approved SPEC (SDD-4). Reads only; a project without a ledger
/// has no cards (exit 0) and nothing is created.
pub fn next(project: &Path, home: &Path, holder: Option<&str>) -> Result<Value> {
    let id = ledger::command::project_id_of(project)?;
    let ledger = match ledger::open_read_only(home, &id) {
        Ok(l) => l,
        Err(LedgerError::Absent) => {
            return Ok(json!({
                "action": "next", "project_id": id, "ledger": "absent", "holder": holder,
                "next": null, "next_is": null,
                "mine": [], "ready": [], "skipped": [], "warnings": [], "writes": false,
            }));
        }
        Err(e) => return Err(e.to_string()),
    };
    let found = workflow::next(ledger.conn(), holder).map_err(|e| e.to_string())?;
    // The SPEC each card names and its status, with a warning for a card that changes the
    // repository without an approved SPEC (SDD-4), in the order the cards are listed.
    let mut warnings: Vec<String> = Vec::new();
    let mut spec_of = |card_id: i64| -> Result<Value> {
        let (kind, spec) = spec_link::kind_and_spec(ledger.conn(), card_id)
            .map_err(|e| LedgerError::from(e).to_string())?
            .unwrap_or_default();
        let (state, warning) =
            spec_link::status(project, &input::card_name(card_id), &kind, spec.as_deref());
        warnings.extend(warning);
        Ok(state)
    };
    let mut mine: Vec<Value> = Vec::new();
    for m in &found.mine {
        let mut v = reservation_json(&m.reservation);
        v["title"] = json!(m.title);
        v["state"] = json!(m.state);
        v["priority"] = json!(m.priority);
        v["lease_live"] = json!(m.lease_live);
        v["spec"] = spec_of(m.reservation.card_id)?;
        mine.push(v);
    }
    let mut ready: Vec<Value> = Vec::new();
    for c in &found.ready {
        ready.push(json!({
            "name": input::card_name(c.card_id), "card_id": c.card_id, "title": c.title,
            "kind": c.kind, "priority": c.priority, "write_scope": c.write_scope,
            "spec": spec_of(c.card_id)?,
        }));
    }
    let skipped: Vec<Value> = found
        .skipped
        .iter()
        .map(|s| {
            let mut v = json!({
                "name": input::card_name(s.card_id), "card_id": s.card_id, "title": s.title,
                "state": s.state, "priority": s.priority, "reason": s.reason.code(),
            });
            match &s.reason {
                SkipReason::Held {
                    holder,
                    generation,
                    lease_until,
                } => {
                    v["holder"] = json!(holder);
                    v["generation"] = json!(generation);
                    v["lease_until"] = json!(lease_until);
                }
                SkipReason::ScopeOverlap {
                    other_card,
                    mine,
                    theirs,
                } => {
                    v["other"] = json!(input::card_name(*other_card));
                    v["mine"] = json!(mine);
                    v["theirs"] = json!(theirs);
                }
                SkipReason::EmptyScope | SkipReason::ReviewNoHolder => {}
            }
            v
        })
        .collect();
    // The card to work on now: one this session already holds, else the first it could take.
    let (next, next_is) = match (found.mine.first(), found.ready.first()) {
        (Some(m), _) => (
            json!(input::card_name(m.reservation.card_id)),
            json!("mine"),
        ),
        (None, Some(c)) => (json!(input::card_name(c.card_id)), json!("ready")),
        (None, None) => (Value::Null, Value::Null),
    };
    Ok(json!({
        "action": "next", "project_id": id, "ledger": "ready", "holder": holder,
        "now": found.now, "next": next, "next_is": next_is,
        "mine": mine, "ready": ready, "skipped": skipped, "warnings": warnings,
        "writes": false,
    }))
}

/// The text of `card note` from `reader` (stdin), without the blanks and line breaks at its
/// end. The limit ([`workflow::MAX_NOTE_BYTES`]) is on that text, so a note of exactly the limit
/// is kept whatever a pipe adds after it, and reading stops at the first byte that makes the
/// text too large: nothing is ever cut short without a word, and no more than the limit is held
/// in memory.
pub fn read_note_text(mut reader: impl std::io::Read) -> Result<Vec<u8>> {
    let max = workflow::MAX_NOTE_BYTES;
    // `kept` ends at the last byte that is not a blank; `blanks` are the blanks since then,
    // stored only while they could still be followed by text that fits.
    let mut kept: Vec<u8> = Vec::new();
    let mut blanks: Vec<u8> = Vec::new();
    let mut buffer = [0u8; 64 * 1024];
    loop {
        let n = reader
            .read(&mut buffer)
            .map_err(|e| format!("card note refused: stdin could not be read ({e})"))?;
        if n == 0 {
            return Ok(kept);
        }
        for &byte in &buffer[..n] {
            if byte.is_ascii_whitespace() {
                if kept.len() + blanks.len() <= max {
                    blanks.push(byte);
                }
            } else {
                if kept.len() + blanks.len() + 1 > max {
                    return Err(format!(
                        "card note refused: the note text is larger than {max} bytes"
                    ));
                }
                kept.append(&mut blanks);
                kept.push(byte);
            }
        }
    }
}

/// What `card note` was given. `text` is what stdin held.
pub struct NoteRequest<'a> {
    pub kind: &'a str,
    pub text: &'a [u8],
    pub reference: Option<&'a str>,
    /// `--by`.
    pub by: Option<&'a str>,
    /// This session's id, when it has one: the author when `--by` is not given.
    pub session: Option<&'a str>,
}

/// `card note C-n --kind K`: appends the text as one note. The line breaks and blanks a pipe
/// leaves at the end are not part of the note.
pub fn note(project: &Path, home: &Path, reference: &str, req: &NoteRequest<'_>) -> Result<Value> {
    let card_id = card_id("note", reference)?;
    let text = std::str::from_utf8(req.text)
        .map_err(|_| "card note refused: the note text on stdin is not UTF-8 text".to_string())?
        .trim_end();
    // A decision is acted on by other sessions, so it is never signed by default.
    if req.kind == "decision" && req.by.is_none() {
        return Err(
            "card note refused: a decision note needs --by NAME: say who decided".to_string(),
        );
    }
    // The one card text screen (`input::secret_words`, SPEC-20261119-card-secret-screen): a note
    // is kept for good and `card export` writes it out. The text is not repeated, and a word
    // is counted once.
    let words = input::secret_words(text);
    if words > 0 {
        return Err(format!(
            "card note refused: the note text has {}",
            input::secret_refusal(words)
        ));
    }
    // C-68 AC-3: the one-line fields pass the same screen as the body.
    if let Some(r) = req.reference {
        let words = input::secret_words(r);
        if words > 0 {
            return Err(format!(
                "card note refused: the note ref has {}",
                input::secret_refusal(words)
            ));
        }
    }
    let author = match (req.by, req.session) {
        (Some(by), _) => actor(Some(by), "note")?,
        (None, Some(session)) => session.to_string(),
        (None, None) => actor(None, "note")?,
    };
    let (id, mut ledger) = open_for_write(project, home)?;
    let note = workflow::add_note(
        &mut ledger,
        card_id,
        req.kind,
        &author,
        req.session,
        text,
        req.reference,
    )
    .map_err(|e| refusal("note", e))?;
    Ok(json!({
        "action": "note", "project_id": id,
        "name": input::card_name(card_id), "card_id": card_id,
        "note_id": note.note_id, "kind": note.kind, "author": note.author,
        "session": req.session,
        "ref": note.reference, "bytes": note.bytes, "created_at": note.created_at,
        "events": 1, "writes": true,
    }))
}

/// `card criterion C-n N --met --evidence REF | --withdraw --reason TEXT`.
pub fn criterion(
    project: &Path,
    home: &Path,
    reference: &str,
    holder: &str,
    generation: i64,
    seq: i64,
    change: CriterionChange<'_>,
) -> Result<Value> {
    let card_id = card_id("criterion", reference)?;
    // C-68 AC-3: the one-line fields pass the same screen as the note body.
    let (what, words) = match &change {
        CriterionChange::Met { evidence } => ("the evidence", input::secret_words(evidence)),
        CriterionChange::Withdrawn { reason } => ("the reason", input::secret_words(reason)),
    };
    if words > 0 {
        return Err(format!(
            "card criterion refused: {what} has {}",
            input::secret_refusal(words)
        ));
    }
    let (id, mut ledger) = open_for_write(project, home)?;
    let c = workflow::set_criterion(&mut ledger, card_id, holder, generation, seq, change)
        .map_err(|e| refusal("criterion", e))?;
    Ok(json!({
        "action": "criterion", "project_id": id,
        "name": input::card_name(card_id), "card_id": card_id,
        "seq": c.seq, "from": c.from, "status": c.status,
        "evidence_ref": c.evidence_ref, "withdrawn_reason": c.withdrawn_reason,
        "events": 1, "writes": true,
    }))
}

/// What `card pr` was given.
pub struct PrRequest<'a> {
    pub holder: &'a str,
    pub generation: i64,
    pub number: u64,
    /// `--from`: the host (`github` when not given, or the host the card came from).
    pub from: Option<&'a str>,
    /// `--repo owner/repo`; without it the project the card came from.
    pub repo: Option<&'a str>,
}

/// The host project of the card's pull request: what was typed, else where the card came from.
fn pr_project(
    conn: &rusqlite::Connection,
    card_id: i64,
    req: &PrRequest<'_>,
) -> Result<ProjectRef> {
    let refused = |why: String| format!("card pr refused: {why}");
    let source: Option<(Option<String>, Option<String>)> = conn
        .query_row(
            "SELECT source_host, source_project FROM card WHERE card_id = ?1",
            [card_id],
            |r| Ok((r.get(0)?, r.get(1)?)),
        )
        .map(Some)
        .or_else(|e| match e {
            rusqlite::Error::QueryReturnedNoRows => Ok(None),
            other => Err(LedgerError::from(other).to_string()),
        })?;
    let Some((source_host, source_project)) = source else {
        return Err(format!("card not found: there is no C-{card_id}"));
    };
    let host_word = req
        .from
        .map(str::to_string)
        .or_else(|| source_host.clone())
        .unwrap_or_else(|| HostKind::Github.as_str().to_string());
    let host = HostKind::parse(&host_word)
        .and_then(HostKind::require_implemented)
        .map_err(|e| refused(e.to_string()))?;
    let repo = match (req.repo, &source_project) {
        (Some(repo), _) => repo.to_string(),
        (None, Some(project)) => project.clone(),
        (None, None) => {
            return Err(refused(format!(
                "C-{card_id} did not come from a code host; name the project with --repo owner/repo"
            )))
        }
    };
    let named = ProjectRef::parse(host, &repo).map_err(|e| refused(e.to_string()))?;
    // A card that came from a host project is closed by a pull request of that project: a
    // merged pull request anywhere else is not this card's work.
    if let (Some(from_host), Some(from_project)) = (source_host, source_project) {
        if named.host().as_str() != from_host || named.source_project() != from_project {
            return Err(refused(format!(
                "C-{card_id} came from {from_host}:{from_project}; its pull request is in that \
                 project, not in {}:{}",
                named.host().as_str(),
                input::echo(&named.source_project())
            )));
        }
    }
    Ok(named)
}

/// `card pr C-n NUMBER`: reads the pull request's state from the host (one GET through the
/// connector) and stores the link with that state. The holder is checked first; the write
/// checks it again in its own transaction.
pub fn pr(
    project: &Path,
    home: &Path,
    reference: &str,
    req: &PrRequest<'_>,
    connector: &impl HostConnector,
) -> Result<Value> {
    let card_id = card_id("pr", reference)?;
    let number = i64::try_from(req.number)
        .ok()
        .filter(|n| *n >= 1)
        .ok_or("card pr refused: a pull request number is 1 or more")?;
    let (id, mut ledger) = open_for_write(project, home)?;
    let host_project = pr_project(ledger.conn(), card_id, req)?;
    workflow::holder_check(
        &mut ledger,
        "pr",
        card_id,
        req.holder,
        req.generation,
        false,
    )
    .map_err(|e| refusal("pr", e))?;
    let state = connector
        .pr_state(&host_project, req.number)
        .map_err(|e| e.to_string())?;
    let pr_state = match state.phase {
        PrPhase::Open => "open",
        PrPhase::Merged => "merged",
        PrPhase::Closed => "closed",
    };
    if state.phase == PrPhase::Merged && state.merge_commit.is_none() {
        return Err(format!(
            "card pr refused: the host says #{} is merged but gave no merge commit; nothing was recorded",
            req.number
        ));
    }
    let project_name = host_project.source_project();
    let link = workflow::link_pr(
        &mut ledger,
        card_id,
        req.holder,
        req.generation,
        &PrTarget {
            host: host_project.host().as_str(),
            project: &project_name,
            number,
            pr_state,
            merge_ref: state.merge_commit.as_deref(),
        },
    )
    .map_err(|e| refusal("pr", e))?;
    Ok(json!({
        "action": "pr", "project_id": id,
        "name": input::card_name(card_id), "card_id": card_id,
        "link_id": link.link_id,
        "to": format!("{}:{}#{}", link.host, link.project, link.number),
        "pr_state": link.pr_state, "merge_ref": link.merge_ref, "head": state.head,
        "created": link.created, "events": 1, "writes": true,
    }))
}

fn moved_json(action: &str, id: &str, moved: &workflow::Moved) -> Value {
    json!({
        "action": action, "project_id": id,
        "name": input::card_name(moved.card_id), "card_id": moved.card_id,
        "from": moved.from, "state": moved.state, "rev": moved.rev,
        "closed_at": moved.closed_at,
        "reservation": reservation_json(&moved.reservation),
        "events": 1, "writes": true,
    })
}

/// `card review C-n`: `active` becomes `review`; the reservation stays.
pub fn review(
    project: &Path,
    home: &Path,
    reference: &str,
    holder: &str,
    generation: i64,
) -> Result<Value> {
    let card_id = card_id("review", reference)?;
    let (id, mut ledger) = open_for_write(project, home)?;
    let moved = workflow::review(&mut ledger, card_id, holder, generation)
        .map_err(|e| refusal("review", e))?;
    Ok(moved_json("review", &id, &moved))
}

/// `card reopen C-n`: `review` becomes `active` again; the reservation stays.
pub fn reopen(
    project: &Path,
    home: &Path,
    reference: &str,
    holder: &str,
    generation: i64,
) -> Result<Value> {
    let card_id = card_id("reopen", reference)?;
    let (id, mut ledger) = open_for_write(project, home)?;
    let moved = workflow::reopen(&mut ledger, card_id, holder, generation)
        .map_err(|e| refusal("reopen", e))?;
    Ok(moved_json("reopen", &id, &moved))
}

/// `card done C-n`: `review` becomes `done` when a criterion is met, none is open and a merged
/// pull request is linked; the reservation ends. A card that names a SPEC closes only while
/// that SPEC passes `spec check` and every one of its acceptance criteria has a `verify:`
/// command or a `manual:` check (card C-86 slice 2, [`spec_link::done_refusal`]).
pub fn done(
    project: &Path,
    home: &Path,
    reference: &str,
    holder: &str,
    generation: i64,
) -> Result<Value> {
    let card_id = card_id("done", reference)?;
    let (id, mut ledger) = open_for_write(project, home)?;
    let spec = spec_link::kind_and_spec(ledger.conn(), card_id)
        .map_err(|e| LedgerError::from(e).to_string())?
        .and_then(|(_, spec)| spec);
    if let Some(why) = spec_link::done_refusal(project, &input::card_name(card_id), spec.as_deref())
    {
        return Err(format!("card done refused: {why}; nothing was changed"));
    }
    let moved =
        workflow::done(&mut ledger, card_id, holder, generation).map_err(|e| refusal("done", e))?;
    Ok(moved_json("done", &id, &moved))
}

/// `card renew C-n`: the lease ends `lease_secs` from now, never earlier than it already did.
pub fn renew(
    project: &Path,
    home: &Path,
    reference: &str,
    holder: &str,
    generation: i64,
    lease_secs: Option<i64>,
) -> Result<Value> {
    let card_id = card_id("renew", reference)?;
    let (id, mut ledger) = open_for_write(project, home)?;
    let r = reservation::renew(
        &mut ledger,
        card_id,
        holder,
        generation,
        lease_secs.unwrap_or(DEFAULT_LEASE_SECS),
    )
    .map_err(|e| e.to_string())?;
    Ok(json!({
        "action": "renew", "project_id": id, "reservation": reservation_json(&r),
        "events": 1, "writes": true,
    }))
}

/// A path as `card check` accepts it, the way `card add --scope` reads a scope entry: relative
/// to the repository, written with `/`, one line, no blank at either end, not empty. The check
/// is on the text of the path: a symbolic link on disk is not followed.
fn check_path(path: &str) -> Result<()> {
    let bad = |why: &str| {
        Err(format!(
            "card check refused: --path {why} (`{}`)",
            input::echo(path)
        ))
    };
    if path.trim().is_empty() {
        bad("is empty")
    } else if path != path.trim() {
        bad("starts or ends with a blank")
    } else if path.contains('\\') {
        bad("must be written with `/`, not `\\`")
    } else if path.starts_with('/') {
        bad("must be relative to the repository")
    } else if path.chars().any(char::is_control) {
        bad("must be one line")
    } else {
        Ok(())
    }
}

/// `card check C-n --path P`: whether this session, at this generation, may write each path:
/// the fence (`fence_check`: current generation, this holder, a live lease) and the
/// reservation's write scope. Reads only; nothing is recorded.
pub fn check(
    project: &Path,
    home: &Path,
    reference: &str,
    holder: &str,
    generation: i64,
    paths: &[String],
) -> Result<Value> {
    let card_id = card_id("check", reference)?;
    if paths.is_empty() {
        return Err("card check refused: name at least one --path".to_string());
    }
    for path in paths {
        check_path(path)?;
    }
    let id = ledger::command::project_id_of(project)?;
    let ledger = ledger::open_read_only(home, &id).map_err(|e| match e {
        LedgerError::Absent => ABSENT_HINT.to_string(),
        other => other.to_string(),
    })?;
    reservation::fence_check(&ledger, card_id, holder, generation).map_err(|e| e.to_string())?;
    let held = reservation::current(ledger.conn(), card_id)
        .map_err(|e| e.to_string())?
        .ok_or("reservation conflict in card: the card has no reservation")?;
    let mut covered = Vec::with_capacity(paths.len());
    for path in paths {
        match workflow::covering_pattern(&held, path) {
            Some(pattern) => covered.push(json!({"path": path, "pattern": pattern})),
            None => {
                return Err(format!(
                    "reservation conflict in write scope: `{}` is not under what C-{card_id} reserved ({})",
                    input::echo(path),
                    held.write_scope.join(", ")
                ))
            }
        }
    }
    Ok(json!({
        "action": "check", "project_id": id,
        "name": input::card_name(card_id), "card_id": card_id,
        "holder": held.holder, "generation": held.generation,
        "lease_until": held.lease_until, "write_scope": held.write_scope,
        "allowed": true, "paths": covered, "writes": false,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn a_check_path_is_relative_one_line_and_not_empty() {
        for good in ["src/a.rs", "a", "./src/a.rs", "src/../a", "with space/a"] {
            assert!(check_path(good).is_ok(), "{good:?}");
        }
        for bad in [
            "",
            " ",
            "/abs",
            "/",
            "a\nb",
            "a\0b",
            "a\tb",
            "a\\b",
            "C:\\a",
            " a",
            "a ",
            "\\\\host\\a",
        ] {
            let err = check_path(bad).unwrap_err();
            assert!(err.starts_with("card check refused: "), "{err}");
        }
    }

    #[test]
    fn a_rejected_value_is_a_refusal_of_the_action_and_other_errors_keep_their_words() {
        assert_eq!(
            refusal(
                "note",
                WorkflowError::InvalidInput("the note text is empty".into())
            ),
            "card note refused: the note text is empty"
        );
        assert_eq!(
            refusal("done", WorkflowError::CardNotFound(4)),
            "card not found: C-4"
        );
    }
}

#[cfg(test)]
mod note_text_tests {
    use super::*;

    fn read(bytes: &[u8]) -> Result<Vec<u8>> {
        read_note_text(bytes)
    }

    #[test]
    fn the_limit_is_on_the_text_without_its_trailing_blanks() {
        let max = workflow::MAX_NOTE_BYTES;
        assert_eq!(read(b"").unwrap(), b"");
        assert_eq!(read(b" \n\t\r\n").unwrap(), b"");
        assert_eq!(read(b"a\n\n b \n").unwrap(), b"a\n\n b");
        let mut edge = vec![b'a'; max];
        assert_eq!(read(&edge).unwrap().len(), max);
        edge.extend(vec![b'\n'; 3 * max]);
        assert_eq!(
            read(&edge).unwrap().len(),
            max,
            "blanks after the text do not count"
        );
        let over = vec![b'a'; max + 1];
        assert!(read(&over).unwrap_err().starts_with("card note refused: "));
        // Blanks inside the text count: the last byte is the one that does not fit.
        let mut inside = vec![b'a'; max - 2];
        inside.extend_from_slice(b"\n\nb");
        assert!(read(&inside).is_err());
        let mut fits = vec![b'a'; max - 3];
        fits.extend_from_slice(b"\n\nb");
        assert_eq!(read(&fits).unwrap().len(), max);
        // Text after more blanks than the limit is not dropped in silence.
        let mut late = b"a".to_vec();
        late.extend(vec![b' '; 2 * max]);
        late.push(b'b');
        assert!(read(&late).is_err());
    }
}
