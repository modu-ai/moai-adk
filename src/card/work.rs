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
use super::dispatch::{checked_scope, reservation_json, DEFAULT_LEASE_SECS};
use super::input::{self, CardRef};
use super::landed;
use super::spec_link;
use crate::forge::connector::{HostConnector, HostKind, PrPhase, ProjectRef};
use crate::ledger::reservation;
use crate::ledger::scope_return;
use crate::ledger::workflow::{self, CriterionChange, PrTarget, SkipReason, WorkflowError};
use crate::ledger::{self, LedgerError};
use serde_json::{json, Value};
use std::path::Path;
use std::process::Command;

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
                "landing": landed::absent_json(),
            }));
        }
        Err(e) => return Err(e.to_string()),
    };
    let mut found = workflow::next(ledger.conn(), holder).map_err(|e| e.to_string())?;
    // A `ready` card whose landing is proven is not offered again: it leaves `ready` for
    // `skipped` with `reason: "already_landed"` and the proof (`card finish` closes it).
    let landing = landed::find(project, ledger.conn(), &landed::candidates(&found))?;
    let landed_skips = landed::take_landed(&mut found, &landing);
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
    let mut skipped: Vec<Value> = found
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
    skipped.extend(landed_skips);
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
        "landing": landing.json(), "writes": false,
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
    // The one card text screen (`input::secret_words`, SPEC-20261119-card-secret-screen and
    // SPEC-20261120-card-secret-followups): a note is kept for good and `card export` writes it
    // out. The text is not repeated, and a word is counted once.
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

/// `card review C-n`: `active` becomes `review`; the reservation stays, and every path of its
/// scope is given back (no `--keep`, SPEC-20261120-card-shared-line-lease).
pub fn review(
    project: &Path,
    home: &Path,
    reference: &str,
    holder: &str,
    generation: i64,
) -> Result<Value> {
    review_keeping(project, home, reference, holder, generation, &[])
}

/// `card review C-n [--keep PATH]…`: `active` becomes `review` and every path of the reservation
/// that no `--keep` names is given back in the same transaction, so another card can take it
/// the moment the review is recorded. A `--keep` path the reservation does not hold refuses the
/// call and writes nothing. The answer lists `given_back` and `kept`; the `card.review` event
/// names both. A card sent back with `card reopen` widens its scope again with `card scope`.
pub fn review_keeping(
    project: &Path,
    home: &Path,
    reference: &str,
    holder: &str,
    generation: i64,
    keep: &[String],
) -> Result<Value> {
    let card_id = card_id("review", reference)?;
    let keep = checked_scope("review", keep)?;
    let (id, mut ledger) = open_for_write(project, home)?;
    let (moved, given_back, kept) =
        scope_return::review_keeping(&mut ledger, card_id, holder, generation, &keep).map_err(
            |e| match e {
                scope_return::ReviewKeepError::Workflow(w) => refusal("review", w),
                scope_return::ReviewKeepError::Scope(s) => s.to_string(),
            },
        )?;
    let mut answer = moved_json("review", &id, &moved);
    answer["given_back"] = json!(given_back);
    answer["kept"] = json!(kept);
    Ok(answer)
}

/// `card give-back C-n --path P…`: the holder gives paths of its reservation back; they are free
/// for any card at once. A path the reservation does not hold, a stale generation or another
/// session is refused and nothing is written. One `card.scope` event names the paths.
pub fn give_back(
    project: &Path,
    home: &Path,
    reference: &str,
    holder: &str,
    generation: i64,
    paths: &[String],
) -> Result<Value> {
    let card_id = card_id("give-back", reference)?;
    let paths = checked_scope("give-back", paths)?;
    if paths.is_empty() {
        return Err("card give-back refused: give at least one --path".to_string());
    }
    let (id, mut ledger) = open_for_write(project, home)?;
    let given = scope_return::give_back(
        &mut ledger,
        card_id,
        holder,
        generation,
        &paths,
        "give-back",
    )
    .map_err(|e| e.to_string())?;
    Ok(json!({
        "action": "give-back", "project_id": id,
        "name": input::card_name(card_id), "card_id": card_id,
        "generation": generation, "holder": holder,
        "given_back": given.given_back, "from": given.from,
        "write_scope": given.reservation.write_scope,
        "reservation": reservation_json(&given.reservation),
        "events": 1, "writes": true,
    }))
}

/// The longest wait for another card's reservation `card line-lease` accepts, in seconds.
pub const MAX_LEASE_WAIT_SECS: u64 = 600;
/// The wait `card line-lease` uses without `--wait-secs`, in seconds.
pub const DEFAULT_LEASE_WAIT_SECS: u64 = 60;

/// `card line-lease C-n --path P… [--wait-secs N] -- COMMAND…`: a short lease of a shared file in
/// which a card needs one line. The paths the reservation does not hold are added (retried every
/// quarter second while another card's reservation covers them, up to `wait_secs`; then the call
/// is refused naming that card and writes nothing), `command` runs with its standard output sent
/// to standard error (the answer stays one JSON envelope), and the added paths are given back
/// whatever the command's status was. A command that exits non-zero or cannot start is refused
/// with its status named, after the paths were given back.
#[allow(clippy::too_many_arguments)]
pub fn line_lease(
    project: &Path,
    home: &Path,
    reference: &str,
    holder: &str,
    generation: i64,
    paths: &[String],
    wait_secs: Option<u64>,
    command: &[String],
) -> Result<Value> {
    use std::process::{Command, Stdio};
    use std::time::{Duration, Instant};
    let card_id = card_id("line-lease", reference)?;
    let paths = checked_scope("line-lease", paths)?;
    if paths.is_empty() {
        return Err("card line-lease refused: give at least one --path".to_string());
    }
    let Some((program, args)) = command.split_first() else {
        return Err("card line-lease refused: give the command to run after `--`".to_string());
    };
    let wait = wait_secs.unwrap_or(DEFAULT_LEASE_WAIT_SECS);
    if wait > MAX_LEASE_WAIT_SECS {
        return Err(format!(
            "card line-lease refused: --wait-secs is 0 to {MAX_LEASE_WAIT_SECS}, not {wait}"
        ));
    }
    let started = Instant::now();
    let (id, mut ledger) = open_for_write(project, home)?;
    let leased = loop {
        match scope_return::lease_take(&mut ledger, card_id, holder, generation, &paths) {
            Ok(leased) => break leased,
            Err(scope_return::LeaseError::Overlap(text)) => {
                if started.elapsed() >= Duration::from_secs(wait) {
                    return Err(format!("{text}; waited {wait} s for it and wrote nothing"));
                }
                std::thread::sleep(Duration::from_millis(250));
            }
            Err(other) => return Err(other.to_string()),
        }
    };
    let waited_ms = started.elapsed().as_millis() as i64;
    // No handle is kept while the command runs: it may take a while and others write the ledger.
    drop(ledger);
    let outcome = Command::new(program)
        .args(args)
        .stdin(Stdio::inherit())
        .stdout(Stdio::piped())
        .stderr(Stdio::inherit())
        .spawn()
        .and_then(|child| child.wait_with_output());
    let status = match &outcome {
        Ok(out) => {
            use std::io::Write;
            let _ = std::io::stderr().write_all(&out.stdout);
            Ok(out.status)
        }
        Err(e) => Err(e.to_string()),
    };
    let mut ledger = open_for_write(project, home)
        .map_err(|e| {
            format!(
                "{e}; the command has already run, so do not run it again, and the lease of {} may still be in the reservation: `card give-back` frees it",
                leased.added.join(", ")
            )
        })?
        .1;
    let released =
        scope_return::lease_release(&mut ledger, card_id, holder, generation, &leased.added)
            .map_err(|e| {
                format!(
                "{e}; the lease of {} may still be in the reservation: `card give-back` frees it",
                leased.added.join(", ")
            )
            })?;
    match status {
        Err(why) => Err(format!(
            "card line-lease refused: could not start `{program}`: {why}; the paths were given back"
        )),
        Ok(status) if !status.success() => Err(format!(
            "card line-lease refused: the command ended with status {}; the paths were given back",
            status
                .code()
                .map_or_else(|| "none (a signal ended it)".to_string(), |c| c.to_string())
        )),
        Ok(_) => {
            let reservation = released
                .as_ref()
                .map(|g| g.reservation.clone())
                .unwrap_or(leased.reservation);
            Ok(json!({
                "action": "line-lease", "project_id": id,
                "name": input::card_name(card_id), "card_id": card_id,
                "generation": generation, "holder": holder,
                "paths": paths, "added": leased.added,
                "given_back": released.map(|g| g.given_back).unwrap_or_default(),
                "exit_code": 0, "waited_ms": waited_ms,
                "reservation": reservation_json(&reservation),
                "writes": !leased.added.is_empty(),
            }))
        }
    }
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
    let (landing, hint) = mflow_landing(&mut ledger, project, card_id)?;
    let moved = workflow::done(&mut ledger, card_id, holder, generation, landing).map_err(|e| {
        let refused = refusal("done", e);
        // The hint explains the landing: it belongs on the landing-related refusal only
        // (the delta audit's N1) — a card refused for an open criterion or a missing
        // reservation reads its own reason, not the profile's.
        match (&hint, refused.contains("no merged pull request is linked")) {
            (Some(h), true) => format!("{refused}; {h}"),
            _ => refused,
        }
    })?;
    Ok(moved_json("done", &id, &moved))
}

/// The M-Flow landing evidence for `card done` (SPEC-20261120-card-done-mflow): when the
/// project's profile names the m-flow work_flow with an integration branch, the newest
/// `card.landed` event's head must be an ancestor of that branch's tip — checked here with
/// git, because the ledger never spawns commands. `None` for any other flow, which keeps the
/// merged-pull-request rule. The profile problem is never swallowed silently
/// (SPEC-20261121-mixed-profile-read): the second value is a hint for the merged-pull-request
/// refusal of a card with a landing on record whose profile could not be read, or was read but
/// does not name the m-flow work_flow with an integration branch: which profile, and why. A
/// project whose profile is only a stub (the card commands write one) keeps the old rule and
/// closes through a merged pull request.
fn mflow_landing(
    ledger: &mut ledger::WriteLedger,
    project: &Path,
    card_id: i64,
) -> Result<(Option<workflow::Landing>, Option<String>)> {
    let newest_head = |ledger: &mut ledger::WriteLedger| -> Option<String> {
        ledger
            .conn()
            .query_row(
                "SELECT json_extract(payload, '$.head') FROM event \
                 WHERE card_id = ?1 AND kind = 'card.landed' \
                 ORDER BY event_id DESC LIMIT 1",
                [card_id],
                |r| r.get(0),
            )
            .ok()
    };
    let profile_rel = crate::profile::read_rel(project);
    let recorded = |detail: String, ledger: &mut ledger::WriteLedger| -> Option<String> {
        newest_head(ledger).map(|_| {
            format!(
                "a local landing is recorded for this card, but {detail}, so the landing was not \
                 used"
            )
        })
    };
    let profile = match crate::profile::read_profile(project) {
        Ok((p, _)) => p,
        Err(e) if e == crate::profile::PROFILE_MISSING => return Ok((None, None)),
        Err(e) => {
            let hint = recorded(
                format!("the project profile {profile_rel} could not be read ({e})"),
                ledger,
            );
            return Ok((None, hint));
        }
    };
    let not_used = |why: String, ledger: &mut ledger::WriteLedger| -> Option<String> {
        recorded(
            format!(
                "the project profile {profile_rel} does not name the m-flow work_flow with an \
                 integration branch ({why})"
            ),
            ledger,
        )
    };
    let wf = match profile.work_flow {
        Some(w) if w.flow == "m-flow" => w,
        Some(w) => {
            let hint = not_used(format!("its work_flow is {}", w.flow), ledger);
            return Ok((None, hint));
        }
        None => {
            let hint = not_used("it has no work_flow".to_string(), ledger);
            return Ok((None, hint));
        }
    };
    let Some(branch) = wf.integration_branch else {
        let hint = not_used(
            "its m-flow work_flow has no integration_branch".to_string(),
            ledger,
        );
        return Ok((None, hint));
    };
    let head: Option<String> = newest_head(ledger);
    let Some(head) = head else {
        return Err(
            "card done refused: cannot close: no local landing is linked (`card landed` after \
             the merge); nothing was changed"
                .to_string(),
        );
    };
    let git_out = |args: &[&str]| -> Result<String> {
        let mut cmd = Command::new("git");
        cmd.arg("-C").arg(project).args(args);
        crate::identity::scrub_git_env(&mut cmd);
        let o = cmd
            .output()
            .map_err(|e| format!("card done refused: git could not run: {e}"))?;
        if !o.status.success() {
            return Err(format!(
                "card done refused: git {} failed: {}",
                args.first().unwrap_or(&"git"),
                String::from_utf8_lossy(&o.stderr).trim()
            ));
        }
        Ok(String::from_utf8_lossy(&o.stdout).trim().to_string())
    };
    let tip = git_out(&["rev-parse", "--verify", &branch])?;
    let mut cmd = Command::new("git");
    cmd.arg("-C")
        .arg(project)
        .args(["merge-base", "--is-ancestor", &head, &tip]);
    crate::identity::scrub_git_env(&mut cmd);
    let o = cmd
        .output()
        .map_err(|e| format!("card done refused: git could not run: {e}"))?;
    if !o.status.success() {
        return Err(format!(
            "card done refused: cannot close: the landed head is not contained in the \
             integration branch {branch}; nothing was changed"
        ));
    }
    Ok((Some(workflow::Landing { head, tip }), None))
}

/// `card landed C-n --head <sha>`: record that this card's head landed on the local
/// integration branch (M-Flow, SPEC-20261120-card-done-mflow). The card must be in `review`;
/// one `card.landed` event names the head and the by, with the writing session as the actor.
pub fn landed(project: &Path, home: &Path, reference: &str, head: &str, by: &str) -> Result<Value> {
    let card_id = card_id("landed", reference)?;
    let (id, mut ledger) = open_for_write(project, home)?;
    let session = std::env::var("CLAUDE_CODE_SESSION_ID")
        .or_else(|_| std::env::var("CODEX_THREAD_ID"))
        .ok();
    workflow::land(&mut ledger, card_id, head, by, session.as_deref())
        .map_err(|e| format!("card landed refused: {e}"))?;
    Ok(json!({
        "action": "landed", "project_id": id,
        "card": input::card_name(card_id), "writes": true,
    }))
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
