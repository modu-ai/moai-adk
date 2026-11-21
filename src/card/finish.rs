//! `moai card finish` (SPEC-20261121-card-finish): a leader closes a `ready`, `active` or
//! `review` card in one transaction, with no reservation of its own to take and no pull request
//! to wait for. What each kind of card closes on:
//!
//! - a card that changes the repository (every kind but `research` and `audit`) closes on a
//!   landing: `--head <sha>` (or the newest head `card landed` recorded, or a merged pull request
//!   link) that git proves is an ancestor of the profile's integration branch tip, and, when the
//!   card names a SPEC, that SPEC passes the same check `card done` makes;
//! - a `research` or `audit` card closes on exactly one of `--evidence-report PATH` (a regular
//!   file under the reports folder of the user level home or under the project folder) or
//!   `--evidence-note REF` (a report or decision note of the card), and refuses `--head`.
//!
//! Every open criterion needs one `--met N=REF` to close the card. A card that changes the
//! repository whose open criterion has no met option still has its proven landing recorded (once)
//! and the answer says `closed: false` and which criteria are open; a research or audit card
//! refuses instead. `--generation N` is checked against the card's newest reservation. All checks
//! come before the first write (the git proof and the files here, the criteria, the generation,
//! the reservation and the notes in [`workflow::finish`]), so a refusal leaves every ledger table
//! as it was.

use super::command::open_for_write;
use super::input::{self, CardRef};
use super::landed::{self, Containment};
use super::spec_link;
use crate::ledger::reservation::Reservation;
use crate::ledger::workflow::{self, FinishEvidence, WorkflowError, MAX_AUTHOR_CHARS};
use crate::ledger::LedgerError;
use rusqlite::OptionalExtension;
use serde_json::{json, Value};
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// The longest report path, in characters (the limit of every one-line value of a card).
pub const MAX_REPORT_PATH_CHARS: usize = workflow::MAX_LINE_CHARS;

/// What `card finish` was given.
pub struct FinishRequest<'a> {
    /// This session's id, when it has one: its own reservation passes the gate.
    pub session: Option<&'a str>,
    /// `--head`: the landed head (40 lowercase hex).
    pub head: Option<&'a str>,
    /// Each `--met N=REF`.
    pub met: &'a [String],
    /// `--evidence-report`: a report file.
    pub evidence_report: Option<&'a str>,
    /// `--evidence-note`: the `ref` of a note of the card.
    pub evidence_note: Option<&'a str>,
    /// `--generation`: the generation of the card's newest reservation the caller expects.
    pub generation: Option<i64>,
    /// `--by`: who finishes the card.
    pub by: &'a str,
}

/// The evidence the card layer has proven, owned.
enum Proven {
    Landing { head: String, tip: String },
    MergedPr,
    Report { path: String },
    Note { reference: String },
}

impl Proven {
    fn evidence(&self) -> FinishEvidence<'_> {
        match self {
            Proven::Landing { head, tip } => FinishEvidence::Landing { head, tip },
            Proven::MergedPr => FinishEvidence::MergedPr,
            Proven::Report { path } => FinishEvidence::Report { path },
            Proven::Note { reference } => FinishEvidence::Note { reference },
        }
    }
}

fn refused(why: impl std::fmt::Display) -> String {
    format!("card finish refused: {why}")
}

/// A refusal that comes before any write.
fn unchanged(why: impl std::fmt::Display) -> String {
    refused(format!("{why}; nothing was changed"))
}

/// The card's number; a host number (`#12`) needs a lookup, so finish refuses it.
fn card_id(reference: &str) -> Result<i64> {
    match input::parse_ref(reference).map_err(refused)? {
        CardRef::Id(n) => Ok(n),
        CardRef::SourceNumber(_) => Err(refused(
            "name the card by its C-number (`card show #N` finds it)",
        )),
    }
}

/// `--by`: one line of 1 to [`MAX_AUTHOR_CHARS`] characters that is not only blanks.
fn check_by(by: &str) -> Result<()> {
    if by.trim().is_empty()
        || by.chars().count() > MAX_AUTHOR_CHARS
        || by.chars().any(char::is_control)
    {
        return Err(refused(format!(
            "--by is required and must be one line of 1 to {MAX_AUTHOR_CHARS} characters"
        )));
    }
    Ok(())
}

/// Each `--met N=REF` as (criterion number, reference). The reference is cut at the first `=`,
/// so it may hold another; its length and line rules are the ledger's, its secret screen is the
/// one every card text passes.
fn parse_met(entries: &[String]) -> Result<Vec<(i64, String)>> {
    let mut met = Vec::with_capacity(entries.len());
    for entry in entries {
        let Some((number, reference)) = entry.split_once('=') else {
            return Err(refused(
                "--met takes N=REF: a criterion number, an equals sign and the evidence",
            ));
        };
        let seq = number
            .trim()
            .parse::<i64>()
            .ok()
            .filter(|n| *n >= 1)
            .ok_or_else(|| refused("--met names a criterion by a number of 1 or more"))?;
        let words = input::secret_words(reference);
        if words > 0 {
            return Err(refused(format!(
                "the --met reference of criterion {seq} has {}",
                input::secret_refusal(words)
            )));
        }
        met.push((seq, reference.to_string()));
    }
    Ok(met)
}

/// The evidence of a card that changes the repository: a landing that git proves, or, off an
/// m-flow profile, the merged pull request link the card already has. An explicit `--head` comes
/// first, then a merged link, then the newest head `card landed` recorded. On an m-flow profile
/// the link alone is no evidence, the same as for `card done`: a landing git proves is needed.
fn landing_evidence(
    project: &Path,
    conn: &rusqlite::Connection,
    card_id: i64,
    head: Option<&str>,
) -> Result<Proven> {
    let sql_err = |e: rusqlite::Error| LedgerError::from(e).to_string();
    let head = match head {
        Some(head) => head.to_string(),
        None => {
            let mflow = landed::integration_branch(project).is_ok();
            let linked: i64 = conn
                .query_row(
                    "SELECT count(*) FROM card_link WHERE from_card = ?1 AND kind = 'pr' \
                     AND state = 'active' AND pr_state = 'merged'",
                    [card_id],
                    |r| r.get(0),
                )
                .map_err(sql_err)?;
            if linked > 0 && !mflow {
                return Ok(Proven::MergedPr);
            }
            let recorded: Option<String> = conn
                .query_row(
                    "SELECT json_extract(payload, '$.head') FROM event \
                     WHERE card_id = ?1 AND kind = 'card.landed' \
                       AND json_type(payload, '$.head') = 'text' \
                     ORDER BY event_id DESC LIMIT 1",
                    [card_id],
                    |r| r.get(0),
                )
                .optional()
                .map_err(sql_err)?;
            recorded.ok_or_else(|| {
                if mflow {
                    unchanged(
                        "no landing is recorded for the card (`card landed` after the merge, or \
                         give the landed commit as --head <sha>); a merged pull request link alone \
                         does not close a card on an m-flow profile",
                    )
                } else {
                    unchanged(
                        "no landing is recorded for the card: give the landed commit as --head \
                         <sha>, or link its merged pull request",
                    )
                }
            })?
        }
    };
    if !crate::gate::is_full_sha(&head) {
        return Err(refused("the head must be 40 lowercase hex"));
    }
    let branch = landed::integration_branch(project).map_err(|why| {
        unchanged(format!(
            "a landing is proven against the integration branch, and the profile gives none ({why})"
        ))
    })?;
    let tip = landed::tip_of(project, &branch).map_err(unchanged)?;
    match landed::contained(project, &head, &tip).map_err(unchanged)? {
        Containment::Inside => Ok(Proven::Landing { head, tip }),
        Containment::Outside => Err(unchanged(format!(
            "the head {head} is not contained in the integration branch {branch}"
        ))),
        Containment::NotACommit => Err(unchanged(format!(
            "the head {head} is not a commit of this repository"
        ))),
    }
}

/// The report path of `--evidence-report` as a canonical path: a regular file under the reports
/// folder of `home` or under the project folder. A relative path is read from the project
/// folder. A link, a `..` and a folder that lead out of both are refused after they are resolved.
fn report_path(project: &Path, home: &Path, given: &str) -> Result<String> {
    let chars = given.chars().count();
    if chars > MAX_REPORT_PATH_CHARS {
        return Err(unchanged(format!(
            "the report path is longer than {MAX_REPORT_PATH_CHARS} characters ({chars})"
        )));
    }
    if given.trim().is_empty() || given.chars().any(char::is_control) {
        return Err(unchanged(
            "the report path must be one line that is not blank",
        ));
    }
    let wanted = Path::new(given);
    let full = if wanted.is_absolute() {
        wanted.to_path_buf()
    } else {
        project.join(wanted)
    };
    let canonical = std::fs::canonicalize(&full).map_err(|e| {
        if e.kind() == std::io::ErrorKind::NotFound {
            unchanged(format!("the report {given} does not exist"))
        } else {
            unchanged(format!("the report path {given} cannot be resolved ({e})"))
        }
    })?;
    if !canonical.is_file() {
        return Err(unchanged(format!(
            "the report {given} is not a regular file"
        )));
    }
    let reports = crate::home_workdir::reports_root(home);
    let under =
        |root: &Path| std::fs::canonicalize(root).is_ok_and(|root| canonical.starts_with(root));
    if !under(&reports) && !under(project) {
        return Err(unchanged(format!(
            "the report {given} is not under the reports folder (<HOME>/.moai/reports) or the \
             project folder"
        )));
    }
    Ok(canonical.to_string_lossy().into_owned())
}

fn previous_json(previous: &Option<Reservation>) -> Value {
    match previous {
        Some(p) => json!({
            "holder": p.holder, "generation": p.generation, "lease_until": p.lease_until,
            "status": p.status,
        }),
        None => Value::Null,
    }
}

/// `card finish C-n`: see the module comment. Every refusal leaves the ledger as it was.
pub fn finish(
    project: &Path,
    home: &Path,
    reference: &str,
    req: &FinishRequest<'_>,
) -> Result<Value> {
    let card_id = card_id(reference)?;
    check_by(req.by)?;
    let met = parse_met(req.met)?;
    let (id, mut ledger) = open_for_write(project, home)?;
    let (kind, spec) = spec_link::kind_and_spec(ledger.conn(), card_id)
        .map_err(|e| LedgerError::from(e).to_string())?
        .ok_or_else(|| WorkflowError::CardNotFound(card_id).to_string())?;

    let proven = if spec_link::changes_repository(&kind) {
        for (option, given) in [
            ("--evidence-report", req.evidence_report),
            ("--evidence-note", req.evidence_note),
        ] {
            if given.is_some() {
                return Err(unchanged(format!(
                    "{option} is for a research or audit card; a `{kind}` card changes the \
                     repository and closes on a landing (--head)"
                )));
            }
        }
        let proven = landing_evidence(project, ledger.conn(), card_id, req.head)?;
        // The SPEC the card names must pass what `card done` asks of it: read from the tip git
        // proved the landing against, because the checkout may not have moved to it yet.
        let name = input::card_name(card_id);
        let refusal = match &proven {
            Proven::Landing { tip, .. } => {
                spec_link::done_refusal_at_tip(project, &name, spec.as_deref(), tip)
            }
            _ => spec_link::done_refusal(project, &name, spec.as_deref()),
        };
        if let Some(why) = refusal {
            return Err(unchanged(why));
        }
        proven
    } else {
        if req.head.is_some() {
            return Err(unchanged(format!(
                "a `{kind}` card does not close on a landing head (--head); give exactly one of \
                 --evidence-report or --evidence-note"
            )));
        }
        match (req.evidence_report, req.evidence_note) {
            (Some(path), None) => Proven::Report {
                path: report_path(project, home, path)?,
            },
            (None, Some(reference)) => Proven::Note {
                reference: reference.to_string(),
            },
            _ => {
                return Err(unchanged(format!(
                    "a `{kind}` card closes on exactly one of --evidence-report or --evidence-note"
                )))
            }
        }
    };

    let evidence = proven.evidence();
    let done = workflow::finish(
        &mut ledger,
        card_id,
        req.session,
        req.by,
        &evidence,
        &met,
        req.generation,
    )
    .map_err(|e| match e {
        WorkflowError::InvalidInput(why) => refused(why),
        other => other.to_string(),
    })?;
    let (via, head, tip, report, note) = match &proven {
        Proven::Landing { head, tip } => ("landing", Some(head), Some(tip), None, None),
        Proven::MergedPr => ("merged_pr", None, None, None, None),
        Proven::Report { path } => ("report", None, None, Some(path), None),
        Proven::Note { reference } => ("note", None, None, None, Some(reference)),
    };
    // A landing-only call applied no met option, so it reports none.
    let applied: &[(i64, String)] = if done.closed { &met } else { &[] };
    Ok(json!({
        "action": "finish", "project_id": id,
        "name": input::card_name(card_id), "card_id": card_id,
        "from": done.from, "state": done.state, "rev": done.rev, "closed_at": done.closed_at,
        "closed": done.closed, "open_criteria": done.open_criteria,
        "via": via, "head": head, "tip": tip, "report": report, "note": note,
        "met": applied.iter().map(|(seq, text)| json!({"seq": seq, "evidence": text})).collect::<Vec<_>>(),
        "by": req.by, "session": req.session,
        "previous": previous_json(&done.previous), "released": done.released,
        "events": done.events, "writes": done.events > 0,
    }))
}
