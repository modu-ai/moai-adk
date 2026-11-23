//! `moai card verdict` (SPEC-20261123-landing-reads-verdict): a session that audited a head
//! records its verdict as one structured event of the card, which `card land` reads. The
//! reviewer is the host's session id, never a typed name, so a verdict cannot be written in
//! another session's name.

use super::command::open_for_write;
use super::input::{self, CardRef};
use super::land::check::git_text;
use crate::ledger::workflow::{self, NewVerdict, VerdictReceipt, VerdictStep, WorkflowError};
use serde_json::{json, Value};
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// What the caller typed.
pub struct VerdictRequest<'a> {
    /// The host's session id, when the session has one.
    pub session: Option<&'a str>,
    pub head: &'a str,
    pub base: &'a str,
    /// `pass` or `fail`.
    pub verdict: &'a str,
    pub tier: u8,
    pub evidence: &'a str,
    /// The head the receipt names, with `steps` (`NAME=EXIT` each).
    pub receipt_head: Option<&'a str>,
    pub steps: &'a [String],
}

fn refused(why: impl std::fmt::Display) -> String {
    format!("card verdict refused: {why}")
}

fn card_id(reference: &str) -> Result<i64> {
    match input::parse_ref(reference).map_err(refused)? {
        CardRef::Id(n) => Ok(n),
        CardRef::SourceNumber(_) => Err(refused(
            "name the card by its C-number (`card show #N` finds it)",
        )),
    }
}

/// 40 lowercase hex.
fn is_hash(text: &str) -> bool {
    text.len() == 40 && text.bytes().all(|b| matches!(b, b'0'..=b'9' | b'a'..=b'f'))
}

fn hash_option(name: &str, value: &str) -> Result<()> {
    if is_hash(value) {
        Ok(())
    } else {
        Err(refused(format!(
            "{name} must be a full commit id (40 lowercase hex), not `{value}`"
        )))
    }
}

/// `NAME=EXIT` into a step.
fn step_of(text: &str) -> Result<VerdictStep> {
    let bad = || {
        refused(format!(
            "--receipt-step wants NAME=EXIT (a name and an exit code), not `{text}`"
        ))
    };
    let (name, exit) = text.split_once('=').ok_or_else(bad)?;
    let name = name.trim();
    if name.is_empty() || name.chars().any(char::is_whitespace) {
        return Err(bad());
    }
    let exit: i64 = exit.trim().parse().map_err(|_| bad())?;
    Ok(VerdictStep {
        name: name.to_string(),
        exit,
    })
}

/// `card verdict C-n --head H --base B --verdict pass|fail --tier N --evidence TEXT
/// [--receipt-head H --receipt-step NAME=EXIT …]`.
pub fn verdict(
    project: &Path,
    home: &Path,
    reference: &str,
    req: &VerdictRequest<'_>,
) -> Result<Value> {
    let card = card_id(reference)?;
    let Some(reviewer) = req.session else {
        return Err(refused(
            "this session has no id, and the reviewer is always the session's own id; a host \
             sets CLAUDE_CODE_SESSION_ID or CODEX_THREAD_ID",
        ));
    };
    hash_option("--head", req.head)?;
    hash_option("--base", req.base)?;
    let pass = match req.verdict {
        "pass" => true,
        "fail" => false,
        other => {
            return Err(refused(format!(
                "--verdict is `pass` or `fail`, not `{other}`"
            )))
        }
    };
    if req.tier > 2 {
        return Err(refused("--tier is 0, 1 or 2"));
    }
    if req.evidence.trim().is_empty() {
        return Err(refused(
            "--evidence names what was read and run, and is empty",
        ));
    }
    // The one card text screen: a verdict is kept for good and `card export` writes it out.
    let words = input::secret_words(req.evidence);
    if words > 0 {
        return Err(refused(format!(
            "the evidence has {}",
            input::secret_refusal(words)
        )));
    }
    let receipt = match (req.receipt_head, req.steps.is_empty()) {
        (None, true) => None,
        (None, false) => {
            return Err(refused(
                "--receipt-step needs --receipt-head, the head the receipt names",
            ))
        }
        (Some(head), _) => {
            hash_option("--receipt-head", head)?;
            let steps = req
                .steps
                .iter()
                .map(|s| step_of(s))
                .collect::<Result<Vec<_>>>()?;
            Some(VerdictReceipt {
                head: head.to_string(),
                steps,
            })
        }
    };
    // Both ends are commits of this repository and the base is an ancestor of the head.
    for (name, value) in [("--head", req.head), ("--base", req.base)] {
        let spec = format!("{value}^{{commit}}");
        git_text(project, &["rev-parse", "--verify", "--quiet", &spec])
            .map_err(|_| refused(format!("{name} {value} is not a commit of this repository")))?;
    }
    git_text(
        project,
        &["merge-base", "--is-ancestor", req.base, req.head],
    )
    .map_err(|_| {
        refused(format!(
            "--base {} is not an ancestor of --head {}",
            req.base, req.head
        ))
    })?;

    let (id, mut ledger) = open_for_write(project, home)?;
    let recorded = workflow::record_verdict(
        &mut ledger,
        card,
        &NewVerdict {
            head: req.head,
            base: req.base,
            pass,
            tier: req.tier,
            evidence: req.evidence,
            receipt: receipt.as_ref(),
            reviewer,
        },
    )
    .map_err(|e| match e {
        WorkflowError::InvalidInput(why) => refused(why),
        other => other.to_string(),
    })?;
    Ok(json!({
        "action": "verdict", "project_id": id,
        "name": input::card_name(card), "card_id": card,
        "event": recorded.event_id, "written": recorded.written,
        "head": req.head, "base": req.base, "verdict": req.verdict, "tier": req.tier,
        "reviewer": reviewer,
        "events": u8::from(recorded.written), "writes": recorded.written,
    }))
}
