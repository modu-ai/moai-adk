//! Putting a card's summary on its issue or pull request, once (#3071 H4, AC7).
//! `card-summary-plan` shows the exact bytes of the comment, the target and the marker and sends
//! nothing; `card-summary-apply` posts that comment only when the plan the caller holds is still
//! the plan, and a permission reference is given. The connector looks for the marker in a
//! member's comment, or one written by the token's own account, before it posts, so a second
//! run posts nothing.
//!
//! The comment is public, written with the user's own token, from text anyone could have put in
//! an issue. So the text is cleaned first ([`super::public_text`]): no line can act as a control
//! line, no `@name` notifies anyone, credential-shaped words are replaced, and the plan says how
//! many of each. The `text` the plan shows is the comment as it is posted, marker included
//! ([`super::connector::summary_comment_body`]), and the plan digest covers it.
//!
//! Nothing is posted automatically: the only caller of [`apply`] is the `card-summary-apply`
//! command. A card that did not come from a host has no target and is refused.
use super::card_import::sources_json;
use super::connector::{
    summary_comment_body, HostConnector, HostKind, ProjectRef, SummaryOutcome, SummaryTarget,
};
use super::public_text::{clean_public_text, Changes};
use crate::ledger::summary::{read_card_summary, sources_in_ledger, summary_marker, CardSummary};
use crate::ledger::{self, LedgerError};
use serde_json::{json, Value};
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// `C-12` or `12` → 12.
pub fn parse_card(text: &str) -> Result<i64> {
    let digits = text.strip_prefix("C-").unwrap_or(text);
    digits
        .parse::<i64>()
        .ok()
        .filter(|n| *n > 0 && !digits.starts_with('+'))
        .ok_or_else(|| format!("`{text}` is not a card id; write it C-<number>"))
}

/// Which card goes to which host target. `number` overrides the card's own issue (a pull request).
pub struct Request<'a> {
    pub project: &'a Path,
    pub home: &'a Path,
    pub card: &'a str,
    pub to: &'a str,
    pub number: Option<u64>,
}

/// What would be posted, to where.
struct Prepared {
    summary: CardSummary,
    target: SummaryTarget,
    marker: String,
    /// The cleaned text handed to the connector.
    cleaned: String,
    /// The comment exactly as the connector posts it: `cleaned` and the marker.
    posted: String,
    changes: Changes,
    /// `host project#number` the card came from.
    source: String,
    ledger_sources: serde_json::Value,
    ledger_notice: serde_json::Value,
    digest: String,
}

fn prepare(req: &Request) -> Result<(String, Prepared)> {
    let Request {
        project,
        home,
        card,
        to,
        number,
    } = *req;
    let host = HostKind::parse(to)
        .and_then(HostKind::require_implemented)
        .map_err(|e| e.to_string())?;
    let card_id = parse_card(card)?;
    let project_id = ledger::command::project_id_of(project)?;
    let read = ledger::open_read_only(home, &project_id).map_err(|error| match error {
        LedgerError::Absent => {
            "ledger not found: register it first (ledger-register-plan, then ledger-register-apply)"
                .to_string()
        }
        other => other.to_string(),
    })?;
    let summary = read_card_summary(&read, card_id)
        .map_err(|e| e.to_string())?
        .ok_or_else(|| format!("ledger: no card C-{card_id} (not found)"))?;
    let source = summary.source.clone().ok_or_else(|| {
        format!("card C-{card_id} did not come from a host, so there is no issue to post to")
    })?;
    if source.host != to {
        return Err(format!(
            "card C-{card_id} came from {}, not {to}; nothing was prepared",
            source.host
        ));
    }
    let host_project = ProjectRef::parse(host, &source.project).map_err(|e| e.to_string())?;
    let target = SummaryTarget {
        project: host_project,
        number: number.unwrap_or(source.number),
    };
    let marker = summary_marker(card_id);
    let (cleaned, changes) = clean_public_text(&summary.render());
    let posted = summary_comment_body(&cleaned, &marker);
    let digest = crate::digest(
        format!(
            "{}\n{}/{}#{}\n{posted}",
            source.host,
            target.project.owner(),
            target.project.repo(),
            target.number
        )
        .as_bytes(),
    );
    let sources = sources_in_ledger(&read).map_err(|e| e.to_string())?;
    let (ledger_sources, ledger_notice) = sources_json(&sources, &source.host, &source.project);
    let source_text = format!("{} {}#{}", source.host, source.project, source.number);
    Ok((
        project_id,
        Prepared {
            summary,
            target,
            marker,
            cleaned,
            posted,
            changes,
            source: source_text,
            ledger_sources,
            ledger_notice,
            digest,
        },
    ))
}

/// `card-summary-plan`: the text, the target and the digest. No network, no write.
pub fn plan(req: &Request) -> Result<Value> {
    let (project_id, p) = prepare(req)?;
    Ok(json!({
        "project_id": project_id,
        "card": format!("C-{}", p.summary.card_id),
        "target": {
            "host": req.to,
            "project": format!("{}/{}", p.target.project.owner(), p.target.project.repo()),
            "number": p.target.number,
        },
        "source": p.source,
        "ledger_sources": p.ledger_sources,
        "ledger_notice": p.ledger_notice,
        "marker": p.marker,
        "text": p.posted,
        "sanitized": {
            "control_characters": p.changes.control_characters,
            "html_comment_openers": p.changes.html_comment_openers,
            "mentions": p.changes.mentions,
            "secret_shapes": p.changes.secret_shapes,
            "control_lines": p.changes.control_lines,
        },
        "warnings": p.changes.warnings(),
        "plan_digest": p.digest,
        "writes": false,
        "network": "none",
        "apply": "card-summary-apply with the same arguments, this plan_digest and a permission reference, only after the user approves; a summary already on the target is not posted again",
    }))
}

/// `card-summary-apply`: posts the planned text once.
pub fn apply(
    req: &Request,
    connector: &dyn HostConnector,
    plan_digest: &str,
    permission_reference: &str,
) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission_reference) {
        return Err("permission reference required".into());
    }
    let (project_id, p) = prepare(req)?;
    if p.digest != plan_digest {
        return Err(
            "stale or altered plan: the card or the target changed since the preview; nothing was posted"
                .into(),
        );
    }
    if connector.host() != p.target.project.host() {
        return Err("the connector is for another host than the card".into());
    }
    let outcome = connector
        .post_summary(&p.target, &p.cleaned, &p.marker)
        .map_err(|e| e.to_string())?;
    let (state, comment_id) = match outcome {
        SummaryOutcome::Posted { comment_id } => ("posted", comment_id),
        SummaryOutcome::AlreadyPosted { comment_id } => ("already_posted", comment_id),
    };
    Ok(json!({
        "project_id": project_id,
        "card": format!("C-{}", p.summary.card_id),
        "state": state,
        "comment_id": comment_id,
        "marker": p.marker,
        "plan_digest": p.digest,
    }))
}
