//! The verdict check of `card land` (SPEC-20261123-landing-reads-verdict): whether the exact
//! candidate head has an independent PASS in the verdict events of the card, read before the
//! landing writes anything.
//!
//! The rules, in the order a PASS meets them:
//!
//! 1. the card has verdict events at all (`no_verdict`) and some for this head
//!    (`verdict_other_head`);
//! 2. no reviewer's newest verdict on the head is a FAIL (`verdict_fail`; the rule of the pull
//!    request gate, SPEC D-4);
//! 3. the authors of the head are known (`authors_unknown`): a commit of the range that counts as
//!    authored carries a `Moai-Session` trailer. A merge commit whose tree is the clean merge of
//!    its two parents is a join and is no authorship (D-10);
//! 4. the reviewer is none of the authors and held no reservation of the card (`self_verdict`);
//! 5. the effective tier is the higher of the stated tier and the floor of the changed paths
//!    (`gate::tier_of`); at tier 1 and 2 the reviewer is not the lander, and a landing with no
//!    session cannot be told from the reviewer (`lander_review`, D-11);
//! 6. at tier 1 and 2 the PASS cites a receipt of this head with a step and every step at exit 0
//!    (`receipt_missing`).
//!
//! A reader that cannot read git or the ledger answers an error text, never a verdict: enforce
//! mode refuses with it, observe mode reports it as `would_refuse` with the code
//! `check_unreadable` and the landing goes on.

use crate::audit_gate::sessions_of;
use crate::gate;
use crate::ledger::workflow::{self, VerdictEvent};
use crate::trailers;
use rusqlite::Connection;
use serde_json::{json, Value};
use std::path::Path;
use std::process::{Command, Output, Stdio};

type Result<T> = std::result::Result<T, String>;

/// What `work_flow.land_verdict` asks for.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Mode {
    /// Report the refusal the landing would get and land as before.
    Observe,
    /// Refuse.
    Enforce,
}

/// The mode of the project: the profile's `work_flow.land_verdict`, `observe` when it names none.
pub fn mode(project: &Path) -> Result<Mode> {
    let (profile, _) = crate::profile::read_profile(project)
        .map_err(|_| "the project profile cannot be read".to_string())?;
    Ok(
        match profile.work_flow.and_then(|w| w.land_verdict).as_deref() {
            Some("enforce") => Mode::Enforce,
            _ => Mode::Observe,
        },
    )
}

/// The verdict a landing relied on, as the start row, the landing event and the answer name it.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Relied {
    pub event: i64,
    pub reviewer: String,
    /// The effective tier.
    pub tier: u8,
    pub lander: Option<String>,
}

impl Relied {
    pub fn json(&self) -> Value {
        json!({
            "event": self.event, "reviewer": self.reviewer, "tier": self.tier,
            "lander": self.lander,
        })
    }
}

/// A refusal: its code and why.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Refusal {
    pub code: &'static str,
    pub why: String,
}

impl Refusal {
    fn new(code: &'static str, why: impl Into<String>) -> Self {
        Refusal {
            code,
            why: why.into(),
        }
    }

    pub fn json(&self) -> Value {
        json!({"code": self.code, "why": self.why})
    }

    pub fn text(&self) -> String {
        format!("{}: {}", self.code, self.why)
    }
}

/// What the check decided.
#[derive(Debug, PartialEq, Eq)]
pub enum Outcome {
    Relied(Relied),
    Refused(Refusal),
}

/// What the check reads.
pub struct Input<'a> {
    pub project: &'a Path,
    pub conn: &'a Connection,
    pub card_id: i64,
    /// The candidate head, 40 lowercase hex.
    pub head: &'a str,
    /// The head the verdict events name: the candidate itself, or the audited head a clean join
    /// candidate was made of (SPEC-20261124-card-select-batch-land D-5).
    pub verdict_head: &'a str,
    /// The integration tip read for this landing.
    pub tip: &'a str,
    /// Whether the head stands on the tip (the landing moves the branch); false when the head is
    /// already inside it.
    pub on_tip: bool,
    /// The session that runs the landing, when it has one.
    pub lander: Option<&'a str>,
}

fn same_session(a: &str, b: &str) -> bool {
    trailers::same_identity(a, b)
}

/// What the landing reads before its first write.
#[derive(Debug, Default, PartialEq, Eq)]
pub struct Reading {
    /// The verdict the landing relies on (the `Relied` object of [`Relied::json`], or the object
    /// the start row of an earlier call names).
    pub relied: Option<Value>,
    /// In observe mode, the refusal the landing would have had: `{code, why}`.
    pub would_refuse: Option<Value>,
}

/// The code `would_refuse` carries in observe mode when the check could not read git or the
/// ledger (REQ-14): the landing goes on as it did before the check existed.
pub const CHECK_UNREADABLE: &str = "check_unreadable";

/// The verdict step of a landing (REQ-6, REQ-14, REQ-15).
///
/// A head the integration branch already holds (`on_tip` is false) is closed on the verdict its
/// start row names, so a FAIL recorded after the move does not stop the close (REQ-17). A head that
/// still stands on the tip has not moved the branch, whatever an earlier call that was stopped
/// before the move wrote: its verdicts are judged again now, so a FAIL recorded in between refuses
/// (REQ-8).
///
/// A refusal is an `Err` in enforce mode. In observe mode it is reported in `would_refuse`, and so
/// is a git or ledger read that failed inside the check, as [`CHECK_UNREADABLE`].
pub fn read(input: &Input<'_>, mode: Mode) -> Result<Reading> {
    if !input.on_tip {
        if let Some(named) = workflow::land_started_verdict(input.conn, input.card_id, input.head)
            .map_err(|e| e.to_string())?
        {
            return Ok(Reading {
                relied: Some(named),
                would_refuse: None,
            });
        }
    }
    let refusal = match judge(input) {
        Ok(Outcome::Relied(relied)) => {
            return Ok(Reading {
                relied: Some(relied.json()),
                would_refuse: None,
            })
        }
        Ok(Outcome::Refused(refusal)) => refusal,
        Err(why) => match mode {
            Mode::Enforce => return Err(why),
            Mode::Observe => Refusal::new(CHECK_UNREADABLE, why),
        },
    };
    match mode {
        Mode::Enforce => Err(refusal.text()),
        Mode::Observe => Ok(Reading {
            relied: None,
            would_refuse: Some(refusal.json()),
        }),
    }
}

/// Decides the landing from the verdict events.
pub fn judge(input: &Input<'_>) -> Result<Outcome> {
    let events = workflow::verdict_events(input.conn, input.card_id).map_err(|e| e.to_string())?;
    let Some(newest_any) = events.last() else {
        return Ok(refused(
            "no_verdict",
            "the card has no verdict record (a session that audited the head records one with \
             `card verdict`; the text of a note is not read)",
        ));
    };
    let on_head: Vec<&VerdictEvent> = events
        .iter()
        .filter(|e| e.head == input.verdict_head)
        .collect();
    if on_head.is_empty() {
        return Ok(refused(
            "verdict_other_head",
            format!(
                "the card has verdicts only for other heads; the newest is for {} (a verdict \
                 names one exact head)",
                newest_any.head
            ),
        ));
    }
    // Each reviewer's newest verdict on the head, newest first.
    let mut newest: Vec<&VerdictEvent> = Vec::new();
    for event in on_head.iter().rev() {
        if !newest
            .iter()
            .any(|n| n.reviewer.eq_ignore_ascii_case(&event.reviewer))
        {
            newest.push(event);
        }
    }
    if let Some(fail) = newest.iter().find(|e| !e.pass) {
        return Ok(refused(
            "verdict_fail",
            format!(
                "the newest verdict of {} on this head is FAIL (event {}); the same reviewer \
                 passes it with a later verdict",
                fail.reviewer, fail.event_id
            ),
        ));
    }
    let mut first: Option<Refusal> = None;
    for event in newest {
        match judge_pass(input, event)? {
            Ok(relied) => return Ok(Outcome::Relied(relied)),
            Err(refusal) => {
                first.get_or_insert(refusal);
            }
        }
    }
    let refusal = first.unwrap_or_else(|| Refusal::new("no_verdict", "no PASS on this head"));
    Ok(Outcome::Refused(refusal))
}

fn refused(code: &'static str, why: impl Into<String>) -> Outcome {
    Outcome::Refused(Refusal::new(code, why))
}

/// One PASS against rules 3 to 6: `Ok(Ok(..))` relies on it, `Ok(Err(..))` is its refusal.
fn judge_pass(
    input: &Input<'_>,
    event: &VerdictEvent,
) -> Result<std::result::Result<Relied, Refusal>> {
    // The range the head added: from the tip when the landing moves the branch, from the base
    // the reviewer named when the head is already inside it (D-12).
    let range_base = if input.on_tip {
        input.tip
    } else {
        event.base.as_str()
    };
    let authors = authors_of(input.project, range_base, input.verdict_head)?;
    if !authors.any_trailer {
        return Ok(Err(Refusal::new(
            "authors_unknown",
            format!(
                "no commit between {range_base} and the head that counts as authored carries a \
                 `Moai-Session` trailer, so the authors of the head are unknown"
            ),
        )));
    }
    let holders =
        workflow::reservation_holders(input.conn, input.card_id).map_err(|e| e.to_string())?;
    if authors
        .sessions
        .iter()
        .chain(holders.iter())
        .any(|a| same_session(a, &event.reviewer))
    {
        return Ok(Err(Refusal::new(
            "self_verdict",
            format!(
                "the PASS of {} comes from a session that wrote a commit of the head or held a \
                 reservation of the card",
                event.reviewer
            ),
        )));
    }
    let tier = floor_tier(input.project, range_base, input.verdict_head, event.tier)?;
    if tier >= 1
        && input
            .lander
            .is_none_or(|lander| same_session(lander, &event.reviewer))
    {
        return Ok(Err(Refusal::new(
            "lander_review",
            format!(
                "the head is tier {tier}: the PASS of {} counts only from a session other than \
                 the one that lands (a landing with no session cannot show that it is another)",
                event.reviewer
            ),
        )));
    }
    if tier >= 1 && !receipt_ok(event) {
        return Ok(Err(Refusal::new(
            "receipt_missing",
            format!(
                "the head is tier {tier}: the PASS must cite a receipt of this head with at \
                 least one step and every step at exit 0"
            ),
        )));
    }
    Ok(Ok(Relied {
        event: event.event_id,
        reviewer: event.reviewer.clone(),
        tier,
        lander: input.lander.map(str::to_string),
    }))
}

fn receipt_ok(event: &VerdictEvent) -> bool {
    event.receipt.as_ref().is_some_and(|r| {
        r.head == event.head && !r.steps.is_empty() && r.steps.iter().all(|s| s.exit == 0)
    })
}

/// A FAIL on the head recorded after the verdict the landing relied on: `{event, reviewer}` of
/// the newest one (REQ-17). The landing is not undone; the answer names it.
pub fn late_fail(conn: &Connection, card_id: i64, head: &str, after: i64) -> Result<Option<Value>> {
    let events = workflow::verdict_events(conn, card_id).map_err(|e| e.to_string())?;
    Ok(events
        .iter()
        .rev()
        .find(|e| e.head == head && !e.pass && e.event_id > after)
        .map(|e| json!({"event": e.event_id, "reviewer": e.reviewer})))
}

// ---- git ---------------------------------------------------------------------------------------

/// The authors of a range: the `Moai-Session` values of the commits that count as authored.
#[derive(Debug, Default, PartialEq, Eq)]
pub(crate) struct Authors {
    pub(crate) sessions: Vec<String>,
    /// Whether any commit that counts as authored carries a session trailer.
    pub(crate) any_trailer: bool,
}

/// One git command in `project`, through the program `MOAI_GIT_BIN` names and the environment
/// scrub every card command uses.
fn git(project: &Path, args: &[&str]) -> Result<Output> {
    let program = std::env::var("MOAI_GIT_BIN").unwrap_or_else(|_| "git".to_string());
    let mut cmd = Command::new(program);
    crate::identity::scrub_git_env(&mut cmd);
    cmd.arg("--no-optional-locks")
        .arg("-C")
        .arg(project)
        .args(args)
        .stdin(Stdio::null())
        .output()
        .map_err(|e| format!("git could not run: {e}"))
}

pub(crate) fn git_text(project: &Path, args: &[&str]) -> Result<String> {
    let out = git(project, args)?;
    if out.status.success() {
        Ok(String::from_utf8_lossy(&out.stdout).into_owned())
    } else {
        Err(format!(
            "git {} failed: {}",
            args.first().copied().unwrap_or(""),
            String::from_utf8_lossy(&out.stderr).trim()
        ))
    }
}

/// The commits of `base..head` that count as authored and their sessions. A merge commit counts
/// as authored unless it is a clean join ([`is_clean_join`]).
pub(crate) fn authors_of(project: &Path, base: &str, head: &str) -> Result<Authors> {
    let range = format!("{base}..{head}");
    let log = git_text(project, &["log", "--format=%H%x1f%P%x1f%B%x1e", &range])?;
    let mut authors = Authors::default();
    for record in log.split('\u{1e}') {
        let record = record.trim_start_matches('\n');
        if record.trim().is_empty() {
            continue;
        }
        let mut fields = record.splitn(3, '\u{1f}');
        let (Some(hash), Some(parents), Some(message)) =
            (fields.next(), fields.next(), fields.next())
        else {
            return Err("git log answered a record this check cannot read".to_string());
        };
        let parents: Vec<&str> = parents.split_whitespace().collect();
        if parents.len() == 2 && is_clean_join(project, hash, parents[0], parents[1]) {
            continue;
        }
        let sessions = sessions_of(message);
        authors.any_trailer |= !sessions.is_empty();
        authors.sessions.extend(sessions);
    }
    Ok(authors)
}

/// Whether the merge commit `commit` is a clean join: its tree is the tree `git merge-tree
/// --write-tree` makes of its two parents. A conflicting merge, an edited one and a git that
/// cannot compute the tree are not joins.
pub(crate) fn is_clean_join(project: &Path, commit: &str, first: &str, second: &str) -> bool {
    let Ok(merged) = git_text(
        project,
        &["merge-tree", "--write-tree", "--no-messages", first, second],
    ) else {
        return false;
    };
    let Some(merged) = merged.lines().next().map(str::trim) else {
        return false;
    };
    let own = format!("{commit}^{{tree}}");
    git_text(project, &["rev-parse", &own]).is_ok_and(|tree| tree.trim() == merged)
}

/// The paths of a merge commit that [`join_of`] names before it adds "and N more".
const LISTED_PATHS: usize = 20;

/// How a merge commit stands against the clean merge of its two parents
/// (SPEC-20261125-land-closes-audited-merge).
#[derive(Debug, PartialEq, Eq)]
pub(crate) enum Join {
    /// Its tree is the tree `git merge-tree --write-tree` makes of the parents.
    Clean,
    /// The parents conflict, so no clean merge exists and the commit holds a resolution.
    Conflicts,
    /// Its tree differs from the clean merge: the first paths that differ and how many there are.
    Differs(Vec<String>, usize),
    /// git could not tell; the reason.
    Unreadable(String),
}

/// Reads how `commit`, a merge of `first` and `second`, stands against their clean merge.
pub(crate) fn join_of(project: &Path, commit: &str, first: &str, second: &str) -> Join {
    let out = match git(
        project,
        &["merge-tree", "--write-tree", "--no-messages", first, second],
    ) {
        Ok(out) => out,
        Err(why) => return Join::Unreadable(why),
    };
    match out.status.code() {
        Some(0) => {}
        Some(1) => return Join::Conflicts,
        _ => {
            return Join::Unreadable(format!(
                "git merge-tree failed: {}",
                String::from_utf8_lossy(&out.stderr).trim()
            ))
        }
    }
    let stdout = String::from_utf8_lossy(&out.stdout).into_owned();
    let Some(merged) = stdout
        .lines()
        .next()
        .map(str::trim)
        .filter(|l| !l.is_empty())
    else {
        return Join::Unreadable("git merge-tree named no tree".to_string());
    };
    let own = match git_text(project, &["rev-parse", &format!("{commit}^{{tree}}")]) {
        Ok(tree) => tree.trim().to_string(),
        Err(why) => return Join::Unreadable(why),
    };
    if own == merged {
        return Join::Clean;
    }
    match git_text(
        project,
        &[
            "diff-tree",
            "-r",
            "--name-only",
            "--no-renames",
            "-z",
            merged,
            &own,
        ],
    ) {
        Ok(text) => {
            let paths: Vec<String> = text
                .split('\0')
                .filter(|p| !p.is_empty())
                .map(str::to_string)
                .collect();
            let total = paths.len();
            Join::Differs(paths.into_iter().take(LISTED_PATHS).collect(), total)
        }
        Err(why) => Join::Unreadable(why),
    }
}

/// The tier a landing judges a head by: the tier the verdict states, raised by the floor of what
/// the head changed since it left `range_base` (the tip's own changes are not the head's). A
/// change that only grows the scan registry does not set the tier 2 floor by its path, the same
/// rule `moai bundle check` and the `audit-gate` job read
/// (SPEC-20261124-registry-line-add-tier REQ-7).
pub fn floor_tier(project: &Path, range_base: &str, head: &str, stated: u8) -> Result<u8> {
    let files_base = git_text(project, &["merge-base", range_base, head])
        .map(|s| s.trim().to_string())
        .unwrap_or_else(|_| range_base.to_string());
    let paths = changed_paths(project, &files_base, head)?;
    let files = gate::changed_files_with_text(paths, Some(&files_base), head, |rev, path| {
        gate::regular_file_text(|args| git_text(project, args).ok(), rev, path)
    });
    Ok(gate::tier_of_changes(Some(stated), &files))
}

/// The paths `base..head` changes, for the floor of the tier: without rename detection, so a
/// renamed file is listed by its old and its new name (REQ-10).
fn changed_paths(project: &Path, base: &str, head: &str) -> Result<Vec<String>> {
    let out = git_text(
        project,
        &["diff", "--name-only", "--no-renames", "-z", base, head],
    )?;
    Ok(out
        .split('\0')
        .filter(|p| !p.is_empty())
        .map(str::to_string)
        .collect())
}
