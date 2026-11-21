//! What the project's git and the ledger together say about a card whose change already landed
//! (SPEC-20261121-card-finish). `card next` uses [`find`] and [`take_landed`] to stop offering a
//! `ready` card that already landed; `card finish` uses [`integration_branch`], [`tip_of`] and
//! [`contained`] to prove a landing head.
//!
//! A proof names the card. Three are accepted, strongest first:
//! - a merged pull request link on the card (the ledger only, no git);
//! - the newest recorded landed head of the card is an ancestor of the integration branch tip;
//! - the `Moai-Spec` trailer of the card's linked SPEC sits on a commit reachable from the tip,
//!   read in one `git log` pass of at most [`COMMIT_LIMIT`] commits (`truncated` says the pass
//!   stopped at the limit).
//!
//! A card number in a commit message is never a proof: a text match is a hypothesis. A SPEC that
//! several cards share lets the trailer proof hide a sibling, so the skip entry names its proof
//! and one explicit `card take` still works.
//!
//! The git proofs need the profile's m-flow integration branch. Without it (or when git cannot
//! read it) only the recorded proofs are used and the answer says so in `landing`. Every git
//! command is read-only, runs through `scrub_git_env` and takes its program from `MOAI_GIT_BIN`
//! when that is set, like the other card commands.

use super::input;
use super::spec_link;
use crate::ledger::workflow::{self, Next, SkipReason};
use crate::ledger::LedgerError;
use rusqlite::Connection;
use serde_json::{json, Value};
use std::collections::{BTreeMap, BTreeSet};
use std::path::Path;
use std::process::{Command, Output, Stdio};

/// The most commits one trailer pass reads from the tip.
pub const COMMIT_LIMIT: usize = 20_000;

/// What proved that a card landed.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum Proof {
    /// The newest recorded landed head is inside the integration branch tip.
    Head { head: String },
    /// An active merged pull request link of the card.
    MergedPr { pr: String, merge_ref: String },
    /// The card's linked SPEC is the `Moai-Spec` trailer of this commit, reachable from the tip.
    SpecTrailer { spec: String, commit: String },
}

impl Proof {
    fn via(&self) -> &'static str {
        match self {
            Proof::Head { .. } => "landed_head",
            Proof::MergedPr { .. } => "merged_pull_request",
            Proof::SpecTrailer { .. } => "spec_trailer",
        }
    }
}

/// What [`find`] looked at and found.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Report {
    /// Whether the git proofs could be used: the profile names an integration branch and git
    /// read it. When false only the merged pull request proof applies.
    pub checked: bool,
    pub branch: Option<String>,
    /// The trailer pass stopped at [`COMMIT_LIMIT`] commits.
    pub truncated: bool,
    /// Why `checked` is false.
    pub note: Option<String>,
    /// The proof of each card that landed, by card number.
    pub proofs: BTreeMap<i64, Proof>,
}

impl Report {
    /// The `landing` key of the `card next` answer.
    pub fn json(&self) -> Value {
        json!({
            "checked": self.checked, "branch": self.branch,
            "truncated": self.truncated, "note": self.note,
        })
    }
}

/// The `landing` key when there is no ledger and so no card was looked at.
pub fn absent_json() -> Value {
    json!({
        "checked": false, "branch": null, "truncated": false,
        "note": "the project has no ledger, so no card was looked at",
    })
}

// ---- git ---------------------------------------------------------------------------------------

/// One read-only git command in `project`.
fn git(project: &Path, args: &[&str]) -> Result<Output, String> {
    // `MOAI_GIT_BIN` names the git program for a runner whose child carries no `PATH`.
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

/// The profile's integration branch, or the reason there is none: only an m-flow profile with
/// an `integration_branch` has one.
pub fn integration_branch(project: &Path) -> Result<String, String> {
    let Ok((profile, _)) = crate::profile::read_profile(project) else {
        return Err("the project profile cannot be read".to_string());
    };
    let Some(flow) = profile.work_flow.filter(|w| w.flow == "m-flow") else {
        return Err("the profile's work_flow is not m-flow".to_string());
    };
    flow.integration_branch
        .ok_or_else(|| "the profile's m-flow work_flow names no integration_branch".to_string())
}

/// The commit the integration branch points at now.
pub fn tip_of(project: &Path, branch: &str) -> Result<String, String> {
    let spec = format!("{branch}^{{commit}}");
    let out = git(project, &["rev-parse", "--verify", "--quiet", &spec])?;
    let tip = String::from_utf8_lossy(&out.stdout).trim().to_string();
    if out.status.success() && crate::gate::is_full_sha(&tip) {
        Ok(tip)
    } else {
        Err(format!(
            "the integration branch {branch} does not point at a commit of this repository"
        ))
    }
}

/// Where a head sits against the integration branch tip.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Containment {
    Inside,
    Outside,
    /// The head names no commit of the repository.
    NotACommit,
}

/// Whether `head` (40 lowercase hex) is an ancestor of `tip` (or equal to it).
pub fn contained(project: &Path, head: &str, tip: &str) -> Result<Containment, String> {
    if !crate::gate::is_full_sha(head) {
        return Ok(Containment::NotACommit);
    }
    let spec = format!("{head}^{{commit}}");
    let known = git(project, &["rev-parse", "--verify", "--quiet", &spec])?;
    if !known.status.success() {
        return Ok(Containment::NotACommit);
    }
    let out = git(project, &["merge-base", "--is-ancestor", head, tip])?;
    match out.status.code() {
        Some(0) => Ok(Containment::Inside),
        Some(1) => Ok(Containment::Outside),
        _ => Err(format!(
            "git could not tell whether {head} is inside {tip}: {}",
            String::from_utf8_lossy(&out.stderr).trim()
        )),
    }
}

/// The `Moai-Spec` trailers of the commits reachable from `tip`, newest first, in one pass: each
/// SPEC id with the newest commit that names it, and whether the pass stopped at the limit.
fn trailer_commits(project: &Path, tip: &str) -> Result<(BTreeMap<String, String>, bool), String> {
    let limit = format!("--max-count={COMMIT_LIMIT}");
    let out = git(
        project,
        &[
            "log",
            &limit,
            "--format=%H%x1f%(trailers:key=Moai-Spec,valueonly,separator=%x1e)",
            "--end-of-options",
            tip,
            "--",
        ],
    )?;
    if !out.status.success() {
        return Err(format!(
            "git could not read the history of {tip}: {}",
            String::from_utf8_lossy(&out.stderr).trim()
        ));
    }
    Ok(parse_trailers(&String::from_utf8_lossy(&out.stdout)))
}

/// The lines of one trailer pass as SPEC id to newest commit, and whether there were
/// [`COMMIT_LIMIT`] of them.
fn parse_trailers(output: &str) -> (BTreeMap<String, String>, bool) {
    let mut found: BTreeMap<String, String> = BTreeMap::new();
    let mut commits = 0;
    for line in output.lines().filter(|l| !l.is_empty()) {
        commits += 1;
        let Some((sha, values)) = line.split_once('\x1f') else {
            continue;
        };
        for value in values
            .split('\x1e')
            .map(str::trim)
            .filter(|v| !v.is_empty())
        {
            found
                .entry(value.to_string())
                .or_insert_with(|| sha.to_string());
        }
    }
    (found, commits >= COMMIT_LIMIT)
}

// ---- card next ---------------------------------------------------------------------------------

/// The `ready` cards `card next` would offer or has skipped only for a scope reason: the ones a
/// landing proof can still remove.
pub fn candidates(found: &Next) -> Vec<i64> {
    let mut ids: Vec<i64> = found.ready.iter().map(|c| c.card_id).collect();
    ids.extend(
        found
            .skipped
            .iter()
            .filter(|s| is_scope_skip(s))
            .map(|s| s.card_id),
    );
    ids
}

fn is_scope_skip(s: &workflow::Skipped) -> bool {
    s.state == "ready"
        && matches!(
            s.reason,
            SkipReason::EmptyScope | SkipReason::ScopeOverlap { .. }
        )
}

/// Looks for a proof of landing for each of `candidates`. No git command runs unless a card has
/// a recorded head or a linked SPEC and no merged pull request link; git trouble never fails the
/// call, it turns `checked` off and says why.
pub fn find(project: &Path, conn: &Connection, candidates: &[i64]) -> Result<Report, String> {
    let wanted: BTreeSet<i64> = candidates.iter().copied().collect();
    let facts = workflow::landing_facts(conn).map_err(|e| e.to_string())?;
    let specs: BTreeMap<i64, String> = spec_link::linked_specs(conn)
        .map_err(|e| LedgerError::from(e).to_string())?
        .into_iter()
        .collect();
    let mut report = Report {
        checked: false,
        branch: None,
        truncated: false,
        note: None,
        proofs: BTreeMap::new(),
    };
    // The proof that needs no git.
    let mut pending: Vec<(i64, Option<String>, Option<String>)> = Vec::new();
    for fact in facts.iter().filter(|f| wanted.contains(&f.card_id)) {
        if let Some((pr, merge_ref)) = &fact.merged_pr {
            report.proofs.insert(
                fact.card_id,
                Proof::MergedPr {
                    pr: pr.clone(),
                    merge_ref: merge_ref.clone(),
                },
            );
            continue;
        }
        let head = fact.head.clone().filter(|h| crate::gate::is_full_sha(h));
        let spec = specs.get(&fact.card_id).cloned();
        if head.is_some() || spec.is_some() {
            pending.push((fact.card_id, head, spec));
        }
    }
    let branch = match integration_branch(project) {
        Ok(branch) => branch,
        Err(reason) => {
            report.note = Some(format!(
                "git proofs were not checked: {reason}; only recorded merged pull request links were used"
            ));
            return Ok(report);
        }
    };
    report.branch = Some(branch.clone());
    if pending.is_empty() {
        report.checked = true;
        return Ok(report);
    }
    let tip = match tip_of(project, &branch) {
        Ok(tip) => tip,
        Err(reason) => {
            report.note = Some(format!(
                "git proofs were not checked: {reason}; only recorded merged pull request links were used"
            ));
            return Ok(report);
        }
    };
    let mut problem: Option<String> = None;
    // One ancestor check per distinct head.
    let mut inside: BTreeMap<String, bool> = BTreeMap::new();
    for (card_id, head, _) in &pending {
        let Some(head) = head else { continue };
        if !inside.contains_key(head) {
            let verdict = match contained(project, head, &tip) {
                Ok(c) => c == Containment::Inside,
                Err(why) => {
                    problem.get_or_insert(why);
                    false
                }
            };
            inside.insert(head.clone(), verdict);
        }
        if inside[head] {
            report
                .proofs
                .insert(*card_id, Proof::Head { head: head.clone() });
        }
    }
    // One trailer pass for the cards that name a SPEC and are not proven yet.
    if pending
        .iter()
        .any(|(id, _, spec)| spec.is_some() && !report.proofs.contains_key(id))
    {
        match trailer_commits(project, &tip) {
            Ok((commits, truncated)) => {
                report.truncated = truncated;
                for (card_id, _, spec) in &pending {
                    let Some(spec) = spec else { continue };
                    if report.proofs.contains_key(card_id) {
                        continue;
                    }
                    if let Some(commit) = commits.get(spec) {
                        report.proofs.insert(
                            *card_id,
                            Proof::SpecTrailer {
                                spec: spec.clone(),
                                commit: commit.clone(),
                            },
                        );
                    }
                }
            }
            Err(why) => {
                problem.get_or_insert(why);
            }
        }
    }
    match problem {
        None => report.checked = true,
        Some(why) => {
            report.note = Some(format!(
                "git proofs were only partly checked: {why}; the proofs found are listed"
            ))
        }
    }
    Ok(report)
}

/// Moves every card of `found` that has a proof out of `ready` and out of the scope skips, and
/// returns its `skipped` entry (`reason: "already_landed"`, `via` and the proof's own key), in
/// card order. The answer's other keys stay as they were.
pub fn take_landed(found: &mut Next, report: &Report) -> Vec<Value> {
    let mut taken: Vec<(i64, Value)> = Vec::new();
    let mut entry = |card_id: i64, title: &str, priority: &str, proof: &Proof| {
        let mut v = json!({
            "name": input::card_name(card_id), "card_id": card_id, "title": title,
            "state": "ready", "priority": priority, "reason": "already_landed",
            "via": proof.via(),
        });
        match proof {
            Proof::Head { head } => v["head"] = json!(head),
            Proof::MergedPr { pr, merge_ref } => {
                v["pr"] = json!(pr);
                v["merge_ref"] = json!(merge_ref);
            }
            Proof::SpecTrailer { spec, commit } => {
                v["spec"] = json!(spec);
                v["commit"] = json!(commit);
            }
        }
        taken.push((card_id, v));
    };
    found.ready.retain(|c| match report.proofs.get(&c.card_id) {
        Some(proof) => {
            entry(c.card_id, &c.title, &c.priority, proof);
            false
        }
        None => true,
    });
    found
        .skipped
        .retain(|s| match report.proofs.get(&s.card_id) {
            Some(proof) if is_scope_skip(s) => {
                entry(s.card_id, &s.title, &s.priority, proof);
                false
            }
            _ => true,
        });
    taken.sort_by_key(|(card_id, _)| *card_id);
    taken.into_iter().map(|(_, v)| v).collect()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn a_trailer_pass_keeps_the_newest_commit_of_each_spec() {
        let out = "c3\x1fSPEC-1\x1eSPEC-2\nc2\x1f\nc1\x1fSPEC-1\n";
        let (found, truncated) = parse_trailers(out);
        assert_eq!(found["SPEC-1"], "c3");
        assert_eq!(found["SPEC-2"], "c3");
        assert_eq!(found.len(), 2);
        assert!(!truncated);
    }

    #[test]
    fn a_trailer_pass_of_the_limit_is_truncated_and_one_fewer_is_not() {
        let line = |n: usize| format!("{n:040x}\x1f\n");
        let full: String = (0..COMMIT_LIMIT).map(line).collect();
        let short: String = (0..COMMIT_LIMIT - 1).map(line).collect();
        assert!(parse_trailers(&full).1);
        assert!(!parse_trailers(&short).1);
        assert!(!parse_trailers("").1);
    }
}
