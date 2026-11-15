//! `lane-safe-to-reset` (Issue #2669): may a lane's session be cleared (`/clear`) or its work
//! dispatched to another lane without losing anything? Read-only.
//!
//! The leader must not clear or redispatch a lane before checking, for that lane: which
//! worktrees it works in, whether any of them holds uncommitted, unpushed or stashed work,
//! whether its screen still runs a background job, and whether every open claim has a resume
//! note. This module gathers what the local repository can tell and takes the rest — the
//! lane's open claims with their resume notes, its session folder, its screen — from an
//! observations file the caller fills from the forge and the terminal, each with its source.
//!
//! The answer is `safe` only when every check ran and found nothing; any finding makes it
//! `not_safe`; a check that could not run (no observation, git failed, an observation too old)
//! makes it `unknown` with the reason. Nothing is written: git runs with
//! `--no-optional-locks` (a plain `status` may rewrite another lane's index) and a scrubbed
//! environment, and no fetch is made — "unpushed" means "not on any local remote-tracking
//! ref", which a stale clone can overstate but never understate.

use crate::forge::transport::Transport;
use crate::identity::scrub_git_env;
use crate::worktree_report::{parse_worktree_list, WorktreeEntry};
use serde::Deserialize;
use serde_json::{json, Value};
use std::collections::BTreeSet;
use std::path::{Path, PathBuf};
use std::process::Command;

type Result<T> = std::result::Result<T, String>;

/// Longest age, in seconds, of a screen observation the verdict may rely on.
pub const SCREEN_MAX_AGE_SECS: i64 = 600;

/// The observations file (typed JSON, unknown keys rejected). Every part is optional; a
/// missing part leaves its check `unknown`.
#[derive(Clone, Debug, Default, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Observations {
    pub schema_version: u32,
    #[serde(default)]
    pub claims: Option<ClaimsObservation>,
    #[serde(default)]
    pub session_cwd: Option<CwdObservation>,
    #[serde(default)]
    pub screen: Option<ScreenObservation>,
}

/// The lane's open claims as read from the forge.
#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct ClaimsObservation {
    pub source: String,
    pub items: Vec<OpenClaim>,
}

#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct OpenClaim {
    pub issue: u64,
    pub claim_comment: u64,
    /// The comment holding the resume note for this claim, if one exists.
    #[serde(default)]
    pub resume_note: Option<u64>,
}

#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct CwdObservation {
    pub source: String,
    pub path: PathBuf,
}

#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct ScreenObservation {
    pub source: String,
    /// `YYYY-MM-DDTHH:MM:SSZ`.
    pub observed_at: String,
    pub background_jobs: u64,
}

/// An observations file that holds every part, as `--help` and the error text show it (#3015
/// C1). A test parses it, so the text cannot drift from the schema.
pub const OBSERVATIONS_EXAMPLE: &str = r#"{"schema_version":1,"claims":{"source":"github","items":[{"issue":7,"claim_comment":70,"resume_note":71}]},"session_cwd":{"source":"orca","path":"/path/to/the/lane/worktree"},"screen":{"source":"orca","observed_at":"2026-11-15T02:30:00Z","background_jobs":0}}"#;

/// The field types the example does not show by itself.
pub const OBSERVATIONS_TYPES: &str = "schema_version: 1; claims.items[].issue, claim_comment, resume_note: numbers (resume_note may be left out); session_cwd.path: text; screen.observed_at: YYYY-MM-DDTHH:MM:SSZ; screen.background_jobs: a number (not a list); no other keys. Every part except schema_version may be left out";

/// Reads an observations file's text; a refusal names the schema and the example.
pub fn parse_observations(text: &str) -> Result<Observations> {
    serde_json::from_str(text).map_err(|e| {
        format!(
            "invalid observations: {e}. Types: {OBSERVATIONS_TYPES}. Example: {OBSERVATIONS_EXAMPLE}"
        )
    })
}

/// One comment of an Issue thread as the claim reading needs it.
#[derive(Clone, Debug)]
pub struct ClaimComment {
    pub id: u64,
    pub author: String,
    pub body: String,
}

/// An Issue number with its comments, oldest first.
#[derive(Clone, Debug)]
pub struct IssueComments {
    pub issue: u64,
    pub comments: Vec<ClaimComment>,
}

/// The lane's open claims in the threads: the live `moai-claim: session=<lane>` comment (not
/// released or handed off, the rule of `lane-next`) with the newest `moai-resume:` note that
/// names this lane, this Issue and this claim and was written by the claim's author.
pub fn claims_from_threads(lane: &str, threads: &[IssueComments]) -> Vec<OpenClaim> {
    let mut items: Vec<OpenClaim> = threads
        .iter()
        .filter_map(|t| {
            let list: Vec<crate::lane_next::Comment> = t
                .comments
                .iter()
                .map(|c| crate::lane_next::Comment {
                    id: c.id,
                    body: c.body.clone(),
                    at: String::new(),
                })
                .collect();
            let claim = crate::lane_next::live_claim(&list, lane)?;
            let author = t.comments.iter().find(|c| c.id == claim)?.author.as_str();
            let resume_note = t
                .comments
                .iter()
                .rev()
                .filter(|c| c.author == author)
                .find(|c| {
                    crate::lane_loop::parse(&c.body)
                        .is_ok_and(|n| n.lane == lane && n.claim == claim && n.issue == t.issue)
                })
                .map(|c| c.id);
            Some(OpenClaim {
                issue: t.issue,
                claim_comment: claim,
                resume_note,
            })
        })
        .collect();
    items.sort_by_key(|i| i.issue);
    items
}

/// Most threads the search returns in its one page (GitHub's per-page limit is 100; the
/// lane holds a handful of claims, so 30 is plenty and a larger count is reported as partial).
const SEARCH_PAGE: u64 = 30;

/// What reading the lane's claims from the forge gave.
#[derive(Clone, Debug)]
pub struct ClaimsRead {
    pub claims: ClaimsObservation,
    /// Why the reading may be incomplete (search result cut or flagged incomplete).
    pub partial: Option<String>,
}

/// `text` with everything but unreserved characters percent-encoded.
pub fn encode_query(text: &str) -> String {
    text.bytes()
        .map(|b| {
            if b.is_ascii_alphanumeric() || b"-._~".contains(&b) {
                (b as char).to_string()
            } else {
                format!("%{b:02X}")
            }
        })
        .collect()
}

/// Reads the lane's open claims from GitHub (GET only): one search for open threads that
/// mention `session=<lane>`, plus the threads in `extra` (the Issues the lane's local
/// worktree branches name, since the search index lags behind a fresh claim), then every
/// thread's comments. A failed read is an error with its reason; a result the search shows
/// only in part is returned with `partial` set.
pub fn read_claims<T: Transport, C: crate::forge::policy::Clock>(
    owner: &str,
    repo: &str,
    lane: &str,
    extra: &[u64],
    transport: &T,
    clock: &C,
) -> Result<ClaimsRead> {
    // Nothing is asked of GitHub with a name that would change the search or the URL
    // (audit #3023 N3).
    if !valid_lane(lane) {
        return Err(format!("invalid lane name `{lane}`"));
    }
    // `repo_ok` allows dots, so `.` and `..` (a path step in the URL) are refused here.
    if !crate::lane_loop::repo_ok(&format!("{owner}/{repo}"))
        || [owner, repo].iter().any(|p| p.chars().all(|c| c == '.'))
    {
        return Err(format!("invalid repository `{owner}/{repo}`"));
    }
    let client =
        crate::forge::Client::new(transport, clock, Default::default(), Default::default());
    let query = format!("repo:{owner}/{repo} is:open in:comments \"session={lane}\"");
    let url = format!(
        "https://api.github.com/search/issues?q={}&per_page={SEARCH_PAGE}",
        encode_query(&query)
    );
    let found = client
        .get(&url, None)
        .map_err(|e| format!("the search for the lane's claims failed: {}", e.message))?;
    let body = found.body.unwrap_or(Value::Null);
    let items = body["items"]
        .as_array()
        .ok_or("the search for the lane's claims answered without a list of items")?;
    let mut numbers: BTreeSet<u64> = items.iter().filter_map(|i| i["number"].as_u64()).collect();
    // A full page whose total is not said may hide more threads (audit #3023 B2): partial.
    let shown = items.len() as u64;
    let partial = match body["total_count"].as_u64() {
        Some(total) if total > shown => Some(format!(
            "the search found {total} threads and returned {shown}"
        )),
        None if shown >= SEARCH_PAGE => Some(format!(
            "the search returned a full page of {shown} threads and did not say how many it found"
        )),
        _ if body["incomplete_results"].as_bool() == Some(true) => {
            Some("GitHub's search flagged its result as incomplete".to_string())
        }
        _ => None,
    };
    numbers.extend(extra.iter().copied());
    let mut threads = Vec::new();
    for n in numbers {
        let url =
            format!("https://api.github.com/repos/{owner}/{repo}/issues/{n}/comments?per_page=100");
        let (raw, complete) = crate::lane_status::pages_with_evidence(&client, url, None)
            .map_err(|e| format!("the comments of #{n} could not be read: {e}"))?;
        // A page that is no list, or a `Link: rel="next"` that cannot be followed, is a read
        // that did not reach the end: never "no claim here" (audit #3023 B1).
        if !complete {
            return Err(format!(
                "the comments of #{n} could not be read to the end (an answer that is no list, or a next page that cannot be followed)"
            ));
        }
        threads.push(IssueComments {
            issue: n,
            comments: raw
                .iter()
                .filter_map(|c| {
                    Some(ClaimComment {
                        id: c["id"].as_u64()?,
                        author: c["user"]["login"].as_str().unwrap_or("").to_string(),
                        body: c["body"].as_str().unwrap_or("").to_string(),
                    })
                })
                .collect(),
        });
    }
    Ok(ClaimsRead {
        claims: ClaimsObservation {
            source: "github:search+issue-comments".into(),
            items: claims_from_threads(lane, &threads),
        },
        partial,
    })
}

/// The Issues named by the branches (`issue/<n>-…`) of the non-bare worktrees whose folder or
/// branch name carries the lane; `git worktree list --porcelain` text in, sorted numbers out.
pub fn lane_branch_issues(worktree_list: &str, lane: &str) -> Vec<u64> {
    let mut out: Vec<u64> = parse_worktree_list(worktree_list)
        .iter()
        .filter(|e| !e.bare)
        .filter(|e| {
            e.path
                .file_name()
                .is_some_and(|n| name_carries_lane(&n.to_string_lossy(), lane))
                || e.branch
                    .as_deref()
                    .is_some_and(|b| name_carries_lane(b, lane))
        })
        .filter_map(|e| e.branch.as_deref().and_then(branch_issue))
        .collect();
    out.sort_unstable();
    out.dedup();
    out
}

/// `OWNER`/`REPO` of the project's `origin` remote when it is a GitHub repository.
pub fn origin_repo(project: &Path) -> Option<(String, String)> {
    let url = git(project, &["remote", "get-url", "origin"])?;
    crate::claim_verify::github_owner_repo(url.trim())
}

/// The Issues the lane's local worktrees name, for [`read_claims`]; empty when git fails.
pub fn local_lane_issues(project: &Path, lane: &str) -> Vec<u64> {
    git(project, &["worktree", "list", "--porcelain"])
        .map(|t| lane_branch_issues(&t, lane))
        .unwrap_or_default()
}

/// A lane name as the fleet writes it: 1 to 64 of ASCII letters, digits, `-`, `_`, `.`.
pub fn valid_lane(name: &str) -> bool {
    !name.is_empty()
        && name.len() <= 64
        && name
            .bytes()
            .all(|b| b.is_ascii_alphanumeric() || b"-_.".contains(&b))
}

fn git(dir: &Path, args: &[&str]) -> Option<String> {
    let mut cmd = Command::new("git");
    scrub_git_env(&mut cmd);
    let out = cmd
        .arg("--no-optional-locks")
        .arg("-C")
        .arg(dir)
        .args(args)
        .output()
        .ok()?;
    out.status
        .success()
        .then(|| String::from_utf8_lossy(&out.stdout).into_owned())
}

/// Whether a worktree folder or branch name carries the lane name as a whole token
/// (`agent-26-audit-2237` carries `agent-26`, not `agent-2`). The fleet writes the same lane
/// with and without its dashes (`2636-adapters-agent22` for `agent-22`; audit FAIL 6850218012
/// on PR #2689), so the dashless spelling counts too.
pub fn name_carries_lane(name: &str, lane: &str) -> bool {
    let dashless = lane.replace('-', "");
    let bytes = name.as_bytes();
    let carries = |l: &str| {
        !l.is_empty()
            && name.match_indices(l).any(|(i, _)| {
                let end = i + l.len();
                let before = i == 0 || !bytes[i - 1].is_ascii_alphanumeric();
                let after = end == bytes.len() || !bytes[end].is_ascii_alphanumeric();
                before && after
            })
    };
    carries(lane) || carries(&dashless)
}

/// The `issue/<n>-…` number of a branch name.
fn branch_issue(branch: &str) -> Option<u64> {
    let rest = branch.strip_prefix("issue/")?;
    let digits: String = rest.chars().take_while(|c| c.is_ascii_digit()).collect();
    (!digits.is_empty() && rest[digits.len()..].starts_with('-'))
        .then(|| digits.parse().ok())
        .flatten()
}

/// Why each worktree belongs to the lane; empty means it does not.
fn matched_by(
    entry: &WorktreeEntry,
    lane: &str,
    claimed: &BTreeSet<u64>,
    cwd_worktree: Option<&Path>,
) -> Vec<String> {
    let mut why = Vec::new();
    if let Some(n) = entry.branch.as_deref().and_then(branch_issue) {
        if claimed.contains(&n) {
            why.push(format!("claim #{n}"));
        }
    }
    if cwd_worktree == Some(entry.path.as_path()) {
        why.push("session_cwd".into());
    }
    if entry
        .path
        .file_name()
        .is_some_and(|n| name_carries_lane(&n.to_string_lossy(), lane))
    {
        why.push("folder name".into());
    }
    // A branch named for the lane (`issue/2670-worktreeinclude-secrets-agent22`).
    if entry
        .branch
        .as_deref()
        .is_some_and(|b| name_carries_lane(b, lane))
    {
        why.push("branch name".into());
    }
    why
}

/// Stash entries per branch name (`WIP on <b>:` / `On <b>:`); `(no branch)` for detached.
pub(crate) fn stash_counts(project: &Path) -> Option<Vec<String>> {
    let text = git(project, &["stash", "list", "--format=%gs"])?;
    Some(
        text.lines()
            .filter_map(|l| {
                let rest = l
                    .strip_prefix("WIP on ")
                    .or_else(|| l.strip_prefix("On "))?;
                Some(rest.split_once(':').map_or(rest, |(b, _)| b).to_string())
            })
            .collect(),
    )
}

fn count_lines(text: &str) -> u64 {
    text.lines().filter(|l| !l.is_empty()).count() as u64
}

/// The worktree part of the answer; findings and unknowns are appended to the lists.
fn check_worktree(
    entry: &WorktreeEntry,
    why: Vec<String>,
    stashes: Option<&[String]>,
    attributed: bool,
    findings: &mut Vec<String>,
    unknowns: &mut Vec<String>,
) -> Value {
    let shown = entry.path.display().to_string();
    let uncommitted = git(
        &entry.path,
        &["status", "--porcelain=v1", "--untracked-files=normal"],
    )
    .map(|t| count_lines(&t));
    // Commits reachable from HEAD and from no remote-tracking ref (no fetch is made).
    let unpushed = git(
        &entry.path,
        &["rev-list", "--count", "HEAD", "--not", "--remotes"],
    )
    .and_then(|t| t.trim().parse::<u64>().ok());
    let label = entry.branch.clone().unwrap_or_else(|| "(no branch)".into());
    let stash = stashes.map(|s| s.iter().filter(|b| **b == label).count() as u64);
    for (what, value) in [
        ("uncommitted change", uncommitted),
        ("unpushed commit", unpushed),
        ("stash entry", stash),
    ] {
        match value {
            None => unknowns.push(format!("{shown}: {what}s could not be read")),
            Some(0) => {}
            Some(n) => findings.push(format!("{shown}: {n} {what}(s)")),
        }
    }
    json!({
        "attributed": attributed,
        "path": shown,
        "branch": entry.branch,
        "head": entry.head,
        "matched_by": why,
        "uncommitted": uncommitted,
        "unpushed": unpushed,
        "stash": stash,
    })
}

/// Builds the answer for `lane` in the repository holding `project`, with `now` as Unix
/// seconds (injected so the screen-age check does not depend on the wall clock in tests).
pub fn check(project: &Path, lane: &str, obs: &Observations, now: i64) -> Result<Value> {
    check_with(project, lane, obs, now, None)
}

/// [`check`] with the reason the caller could not (or only in part) read the lane's claims
/// from the forge itself (#3015 C1); the reason becomes an `unknowns` entry.
pub fn check_with(
    project: &Path,
    lane: &str,
    obs: &Observations,
    now: i64,
    claims_note: Option<&str>,
) -> Result<Value> {
    check_full(project, lane, obs, now, claims_note, &Extras::default())
}

/// What the caller read besides the observations, for the lane's state (#3015 C3b). `None` is
/// "not read".
#[derive(Clone, Debug, Default)]
pub struct Extras {
    /// The open pull requests that lead to the lane's open claims.
    pub open_prs: Option<Vec<crate::lane_state::OpenPr>>,
    /// The pull request read was cut at one page.
    pub prs_partial: bool,
    /// Whether the lane's session lease still runs.
    pub session_live: Option<bool>,
}

/// [`check_with`] with [`Extras`]; the answer then carries `lane_state` as well.
pub fn check_full(
    project: &Path,
    lane: &str,
    obs: &Observations,
    now: i64,
    claims_note: Option<&str>,
    extras: &Extras,
) -> Result<Value> {
    if !valid_lane(lane) {
        return Err("invalid lane name: 1 to 64 of letters, digits, `-`, `_`, `.`".into());
    }
    if obs.schema_version != 1 {
        return Err("invalid observations: schema_version must be 1".into());
    }
    let list = git(project, &["worktree", "list", "--porcelain"])
        .ok_or("git worktree list unavailable for the project")?;
    let entries = parse_worktree_list(&list);
    let (mut findings, mut unknowns) = (Vec::new(), Vec::new());

    // (4) every open claim has a resume note.
    let claimed: BTreeSet<u64> = obs
        .claims
        .as_ref()
        .map(|c| c.items.iter().map(|i| i.issue).collect())
        .unwrap_or_default();
    let claims = match &obs.claims {
        None => {
            unknowns.push(match claims_note {
                Some(why) => format!("claims could not be read from GitHub: {why}"),
                None => "no claims observation: open claims and resume notes unchecked".into(),
            });
            Value::Null
        }
        Some(c) => {
            if let Some(why) = claims_note {
                unknowns.push(format!("claims read only in part: {why}"));
            }
            for i in c.items.iter().filter(|i| i.resume_note.is_none()) {
                findings.push(format!(
                    "open claim #{} (comment {}) has no resume note",
                    i.issue, i.claim_comment
                ));
            }
            json!({
                "source": c.source,
                "items": c.items.iter().map(|i| json!({
                    "issue": i.issue,
                    "claim_comment": i.claim_comment,
                    "resume_note": i.resume_note,
                })).collect::<Vec<_>>(),
            })
        }
    };

    // (1) the lane's worktrees, (2) their uncommitted, unpushed and stashed work.
    // git lists real paths (`/private/var/…` on macOS), so a symlinked spelling is resolved.
    let cwd_real = obs
        .session_cwd
        .as_ref()
        .map(|c| c.path.canonicalize().unwrap_or_else(|_| c.path.clone()));
    let cwd = cwd_real.as_deref();
    if obs.session_cwd.is_none() {
        unknowns.push("no session_cwd observation: the session's own folder unchecked".into());
    }
    let stashes = stash_counts(project);
    if stashes.is_none() {
        unknowns.push("git stash list unavailable".into());
    }
    // For the lane's state: unsaved work in the lane's own worktrees, `None` when a check
    // of one of them could not run.
    let (mut unsaved_found, mut unsaved_unread) = (false, stashes.is_none());
    // Linked worktrees usually live inside the primary checkout, so the session folder belongs
    // to the deepest worktree that contains it, not to every one.
    let cwd_worktree = cwd.and_then(|c| {
        entries
            .iter()
            .filter(|e| !e.bare && c.starts_with(&e.path))
            .max_by_key(|e| e.path.components().count())
            .map(|e| e.path.as_path())
    });
    // The primary checkout is shared: lanes start there and work in linked worktrees, so its
    // changes belong to the person and the leader, not to a lane whose only tie to it is the
    // session folder (leader's first use, #2669 comment 6851881955). Its counts are reported
    // under `unattributed` and do not decide the verdict.
    // Only a non-bare first entry is a primary checkout; a bare repository has none, and its
    // first linked worktree is a lane's like any other (Codex review on PR #2715).
    let primary = entries
        .first()
        .filter(|e| !e.bare)
        .map(|e| e.path.as_path());
    let mut worktrees = Vec::new();
    let mut unattributed = Vec::new();
    for entry in entries.iter().filter(|e| !e.bare) {
        let why = matched_by(entry, lane, &claimed, cwd_worktree);
        if why.is_empty() {
            continue;
        }
        let only_cwd = why.iter().all(|w| w == "session_cwd");
        let attributed = !(only_cwd && primary == Some(entry.path.as_path()));
        let (mut f, mut u) = (Vec::new(), Vec::new());
        worktrees.push(check_worktree(
            entry,
            why,
            stashes.as_deref(),
            attributed,
            &mut f,
            &mut u,
        ));
        if attributed {
            unsaved_found |= !f.is_empty();
            unsaved_unread |= !u.is_empty();
            findings.extend(f);
            unknowns.extend(u);
        } else {
            unattributed.extend(f.into_iter().chain(u));
        }
    }
    if let Some(c) = cwd {
        if cwd_worktree.is_none() {
            unknowns.push(format!(
                "session_cwd {} is in no worktree of this repository",
                c.display()
            ));
        }
    }
    for n in &claimed {
        let has = entries
            .iter()
            .any(|e| e.branch.as_deref().and_then(branch_issue) == Some(*n));
        if !has {
            unknowns.push(format!(
                "open claim #{n} has no local `issue/{n}-…` worktree to check"
            ));
        }
    }

    // (3) background jobs on the lane's screen.
    let mut fresh_jobs = None;
    let screen = match &obs.screen {
        None => {
            unknowns.push("no screen observation: background jobs unchecked".into());
            Value::Null
        }
        Some(s) => {
            match crate::forge::policy::parse_utc(&s.observed_at) {
                None => unknowns.push(format!(
                    "screen observed_at `{}` is not YYYY-MM-DDTHH:MM:SSZ",
                    s.observed_at
                )),
                Some(at) if now - at > SCREEN_MAX_AGE_SECS || at > now => unknowns.push(format!(
                    "screen observation is {} s old; the limit is {SCREEN_MAX_AGE_SECS} s",
                    now - at
                )),
                Some(_) if s.background_jobs > 0 => {
                    fresh_jobs = Some(s.background_jobs);
                    findings.push(format!(
                        "{} background job(s) still running on the lane's screen",
                        s.background_jobs
                    ));
                }
                Some(_) => fresh_jobs = Some(0),
            }
            json!({"source": s.source, "observed_at": s.observed_at, "background_jobs": s.background_jobs})
        }
    };

    let verdict = if !findings.is_empty() {
        "not_safe"
    } else if !unknowns.is_empty() {
        "unknown"
    } else {
        "safe"
    };
    let open_claims: Option<Vec<OpenClaim>> = obs.claims.as_ref().map(|c| c.items.clone());
    let lane_state = crate::lane_state::classify(&crate::lane_state::Facts {
        claims: open_claims.as_deref(),
        claims_partial: obs.claims.is_some() && claims_note.is_some(),
        open_prs: extras.open_prs.as_deref(),
        prs_partial: extras.prs_partial,
        unsaved: if unsaved_unread {
            None
        } else {
            Some(unsaved_found)
        },
        background_jobs: fresh_jobs,
        session_live: extras.session_live,
    });
    Ok(json!({
        "lane": lane,
        "verdict": verdict,
        "lane_state": lane_state,
        "findings": findings,
        "unknowns": unknowns,
        "unattributed": unattributed,
        "worktrees": worktrees,
        "claims": claims,
        "session_cwd": obs.session_cwd.as_ref().map(|c| json!({"source": c.source, "path": c.path})),
        "screen": screen,
        "unpushed_basis": "local remote-tracking refs; no fetch",
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn lane_names_match_as_whole_tokens() {
        assert!(name_carries_lane("agent-26-audit-2237", "agent-26"));
        assert!(name_carries_lane("audit-2491-agent-28", "agent-28"));
        assert!(name_carries_lane("agent-27", "agent-27"));
        assert!(!name_carries_lane("agent-26-audit-2237", "agent-2"));
        assert!(!name_carries_lane("xagent-27", "agent-27"));
        assert!(!name_carries_lane("2669-lane-safe-to-reset", "agent-27"));
        // The fleet also drops the dash (audit FAIL 6850218012 on PR #2689).
        assert!(name_carries_lane("2636-adapters-agent22", "agent-22"));
        assert!(name_carries_lane(
            "issue/2670-worktreeinclude-secrets-agent22",
            "agent-22"
        ));
        assert!(name_carries_lane("audit-2486-agent29", "agent-29"));
        assert!(!name_carries_lane("2636-adapters-agent220", "agent-22"));
        assert!(!name_carries_lane("2636-adapters-agent2", "agent-22"));
    }

    #[test]
    fn branch_numbers_need_the_issue_prefix_and_a_dash() {
        assert_eq!(branch_issue("issue/2669-lane-safe"), Some(2669));
        assert_eq!(branch_issue("issue/2669"), None);
        assert_eq!(branch_issue("feature/2669-x"), None);
        assert_eq!(branch_issue("issue/x-1"), None);
    }

    #[test]
    fn lane_name_limits() {
        assert!(valid_lane(&"a".repeat(64)));
        assert!(!valid_lane(&"a".repeat(65)));
        assert!(!valid_lane(""));
        assert!(!valid_lane("agent 27"));
        assert!(!valid_lane("../x"));
    }

    // --- #3015 C1: the tool reads the lane's claims itself ---------------------------------

    use crate::forge::transport::{HttpRequest, HttpResponse, Transport};
    use std::cell::RefCell;

    fn cc(id: u64, author: &str, body: &str) -> ClaimComment {
        ClaimComment {
            id,
            author: author.into(),
            body: body.into(),
        }
    }

    fn claim(id: u64, author: &str, lane: &str) -> ClaimComment {
        cc(
            id,
            author,
            &format!("moai-claim: session={lane} epoch=1 lease_until=2026-11-15T08:00:00Z"),
        )
    }

    fn note(id: u64, author: &str, lane: &str, issue: u64, claim: u64) -> ClaimComment {
        cc(
            id,
            author,
            &format!(
                "moai-resume: lane={lane} issue={issue} claim={claim} lease_until=2026-11-15T08:00:00Z \
branch=issue/{issue}-x head=0123456789abcdef0123456789abcdef01234567 uncommitted=0\nnext: push"
            ),
        )
    }

    fn thread(issue: u64, comments: Vec<ClaimComment>) -> IssueComments {
        IssueComments { issue, comments }
    }

    #[test]
    fn an_open_claim_with_its_resume_note_is_read_from_the_thread() {
        let t = thread(
            7,
            vec![
                claim(6000070, "Bot", "agent-9"),
                note(6000071, "Bot", "agent-9", 7, 6000070),
            ],
        );
        let items = claims_from_threads("agent-9", &[t]);
        assert_eq!(items.len(), 1);
        assert_eq!(
            (items[0].issue, items[0].claim_comment, items[0].resume_note),
            (7, 6000070, Some(6000071))
        );
    }

    #[test]
    fn a_released_or_handed_off_claim_is_not_open() {
        let released = thread(
            7,
            vec![
                claim(6000070, "Bot", "agent-9"),
                cc(6000072, "Bot", "moai-claim-release: comment=6000070"),
            ],
        );
        let handed = thread(
            8,
            vec![
                claim(6000080, "Bot", "agent-9"),
                cc(
                    6000082,
                    "Bot",
                    "moai-claim-handoff: comment=6000080 to=agent-3",
                ),
            ],
        );
        assert!(claims_from_threads("agent-9", &[released, handed]).is_empty());
    }

    #[test]
    fn another_lanes_claim_is_not_this_lanes() {
        let t = thread(7, vec![claim(6000070, "Bot", "agent-90")]);
        assert!(claims_from_threads("agent-9", &[t]).is_empty());
    }

    #[test]
    fn a_resume_note_must_name_this_lane_and_this_claim_and_come_from_the_claim_author() {
        let wrong_claim = thread(
            1,
            vec![
                claim(6000010, "Bot", "agent-9"),
                note(6000011, "Bot", "agent-9", 1, 6000099),
            ],
        );
        let wrong_lane = thread(
            2,
            vec![
                claim(6000020, "Bot", "agent-9"),
                note(6000021, "Bot", "agent-3", 2, 6000020),
            ],
        );
        let wrong_author = thread(
            3,
            vec![
                claim(6000030, "Bot", "agent-9"),
                note(6000031, "Other", "agent-9", 3, 6000030),
            ],
        );
        let plain_text = thread(
            4,
            vec![
                claim(6000040, "Bot", "agent-9"),
                cc(6000041, "Bot", "resume note - see claim"),
            ],
        );
        // A note copied from another Issue's thread names that Issue, not this one.
        let wrong_issue = thread(
            5,
            vec![
                claim(6000050, "Bot", "agent-9"),
                note(6000051, "Bot", "agent-9", 6, 6000050),
            ],
        );
        let items = claims_from_threads(
            "agent-9",
            &[
                wrong_claim,
                wrong_lane,
                wrong_author,
                plain_text,
                wrong_issue,
            ],
        );
        assert_eq!(items.len(), 5);
        assert!(items.iter().all(|i| i.resume_note.is_none()), "{items:?}");
    }

    #[test]
    fn the_newest_matching_resume_note_wins() {
        let t = thread(
            7,
            vec![
                claim(6000070, "Bot", "agent-9"),
                note(6000071, "Bot", "agent-9", 7, 6000070),
                note(6000075, "Bot", "agent-9", 7, 6000070),
            ],
        );
        assert_eq!(
            claims_from_threads("agent-9", &[t])[0].resume_note,
            Some(6000075)
        );
    }

    /// Answers GETs from a list of (url prefix, status, body); keeps what was sent.
    struct Fake {
        answers: Vec<(String, u16, serde_json::Value)>,
        sent: RefCell<Vec<(String, String)>>,
    }

    impl Transport for Fake {
        fn send(&self, req: &HttpRequest) -> std::result::Result<HttpResponse, String> {
            self.sent
                .borrow_mut()
                .push((req.method.clone(), req.url.clone()));
            let (_, status, body) = self
                .answers
                .iter()
                .find(|(prefix, _, _)| req.url.starts_with(prefix.as_str()))
                .ok_or_else(|| format!("unexpected request {} {}", req.method, req.url))?;
            Ok(HttpResponse::new(*status).with_body(body.clone()))
        }
    }

    struct Clock0;
    impl crate::forge::policy::Clock for Clock0 {
        fn now(&self) -> i64 {
            0
        }
        fn sleep(&self, _: u64) {}
    }

    const API: &str = "https://api.github.com/repos/acme/widgets";

    fn raw(id: u64, author: &str, body: &str) -> serde_json::Value {
        json!({"id": id, "user": {"login": author}, "body": body})
    }

    fn search(numbers: &[u64], total: u64, incomplete: bool) -> (String, u16, serde_json::Value) {
        (
            "https://api.github.com/search/issues?q=".into(),
            200,
            json!({"total_count": total, "incomplete_results": incomplete,
                   "items": numbers.iter().map(|n| json!({"number": n})).collect::<Vec<_>>()}),
        )
    }

    fn comments(n: u64, list: Vec<serde_json::Value>) -> (String, u16, serde_json::Value) {
        (
            format!("{API}/issues/{n}/comments?per_page=100"),
            200,
            json!(list),
        )
    }

    #[test]
    fn claims_are_read_by_one_search_and_one_comment_read_per_thread_with_gets_only() {
        let fake = Fake {
            answers: vec![
                search(&[7], 1, false),
                comments(
                    7,
                    vec![
                        raw(6000070, "Bot", "moai-claim: session=agent-9 epoch=1 lease_until=2026-11-15T08:00:00Z"),
                        raw(6000071, "Bot", "moai-resume: lane=agent-9 issue=7 claim=6000070 lease_until=2026-11-15T08:00:00Z branch=issue/7-x head=0123456789abcdef0123456789abcdef01234567 uncommitted=0\nnext: push"),
                    ],
                ),
            ],
            sent: RefCell::new(Vec::new()),
        };
        let got = read_claims("acme", "widgets", "agent-9", &[], &fake, &Clock0).unwrap();
        assert_eq!(got.partial, None);
        assert_eq!(got.claims.source, "github:search+issue-comments");
        assert_eq!(got.claims.items.len(), 1);
        assert_eq!(got.claims.items[0].resume_note, Some(6000071));
        let sent = fake.sent.borrow();
        assert!(sent.iter().all(|(m, _)| m == "GET"), "{sent:?}");
        assert_eq!(sent.len(), 2, "{sent:?}");
        // The search names the repository, the open state and the lane's session token.
        // The whole encoded query (audit #3023 N6): open threads, comments, the lane's token.
        assert_eq!(
            sent[0].1,
            "https://api.github.com/search/issues?q=repo%3Aacme%2Fwidgets%20is%3Aopen%20in%3Acomments%20%22session%3Dagent-9%22&per_page=30"
        );
        assert_eq!(sent[1].1, format!("{API}/issues/7/comments?per_page=100"));
    }

    #[test]
    fn a_thread_the_search_missed_is_read_when_a_local_branch_names_it() {
        // GitHub's search index lags: a fresh claim may be missing from it.
        let fake = Fake {
            answers: vec![
                search(&[], 0, false),
                comments(
                    9,
                    vec![raw(
                        6000090,
                        "Bot",
                        "moai-claim: session=agent-9 epoch=1 lease_until=2026-11-15T08:00:00Z",
                    )],
                ),
            ],
            sent: RefCell::new(Vec::new()),
        };
        let got = read_claims("acme", "widgets", "agent-9", &[9], &fake, &Clock0).unwrap();
        assert_eq!(got.claims.items.len(), 1);
        assert_eq!(got.claims.items[0].issue, 9);
        assert_eq!(got.claims.items[0].resume_note, None);
    }

    #[test]
    fn a_search_that_does_not_show_everything_is_reported_as_partial() {
        let full = Fake {
            answers: vec![search(&[7], 31, false), comments(7, vec![])],
            sent: RefCell::new(Vec::new()),
        };
        let got = read_claims("acme", "widgets", "agent-9", &[], &full, &Clock0).unwrap();
        assert!(
            got.partial.as_deref().unwrap().contains("31"),
            "{:?}",
            got.partial
        );
        let lagging = Fake {
            answers: vec![search(&[7], 1, true), comments(7, vec![])],
            sent: RefCell::new(Vec::new()),
        };
        let got = read_claims("acme", "widgets", "agent-9", &[], &lagging, &Clock0).unwrap();
        assert!(got.partial.is_some());
        // Exactly the page size is not partial.
        let exact = Fake {
            answers: vec![search(&[7], 1, false), comments(7, vec![])],
            sent: RefCell::new(Vec::new()),
        };
        let got = read_claims("acme", "widgets", "agent-9", &[], &exact, &Clock0).unwrap();
        assert_eq!(got.partial, None);
    }

    #[test]
    fn a_read_that_fails_is_an_error_with_the_reason() {
        let down = Fake {
            answers: vec![(
                "https://api.github.com/search/issues?q=".into(),
                403,
                json!({"message": "rate limit"}),
            )],
            sent: RefCell::new(Vec::new()),
        };
        let err = read_claims("acme", "widgets", "agent-9", &[], &down, &Clock0)
            .map(|_| ())
            .unwrap_err();
        assert!(err.contains("search"), "{err}");
        let no_comments = Fake {
            answers: vec![
                search(&[7], 1, false),
                (
                    format!("{API}/issues/7/comments?per_page=100"),
                    500,
                    json!({}),
                ),
            ],
            sent: RefCell::new(Vec::new()),
        };
        let err = read_claims("acme", "widgets", "agent-9", &[], &no_comments, &Clock0)
            .map(|_| ())
            .unwrap_err();
        assert!(err.contains("#7"), "{err}");
    }

    fn search_body(count: usize, total: Option<u64>) -> (String, u16, serde_json::Value) {
        let mut body = json!({"incomplete_results": false,
            "items": (1..=count as u64).map(|n| json!({"number": n})).collect::<Vec<_>>()});
        if let Some(t) = total {
            body["total_count"] = json!(t);
        }
        ("https://api.github.com/search/issues?q=".into(), 200, body)
    }

    /// Audit #3023 B2: a full search page is partial unless its total says it was all there.
    #[test]
    fn a_full_search_page_is_partial_unless_the_total_says_otherwise() {
        let any_comments = || (format!("{API}/issues/"), 200, json!([]));
        for (count, total, partial) in [
            (30, Some(30), false),
            (30, Some(31), true),
            (30, None, true),
            (29, None, false),
            (29, Some(29), false),
            (0, None, false),
        ] {
            let fake = Fake {
                answers: vec![search_body(count, total), any_comments()],
                sent: RefCell::new(Vec::new()),
            };
            let got = read_claims("acme", "widgets", "agent-9", &[], &fake, &Clock0).unwrap();
            assert_eq!(
                got.partial.is_some(),
                partial,
                "{count} threads, total {total:?}: {:?}",
                got.partial
            );
        }
    }

    /// Answers the search with thread 7 and thread 7's comments in `pages` pages chained by
    /// `Link`; `last_link` is the `Link` the last page carries (none when the chain ends).
    struct Chain {
        pages: u32,
        last_link: Option<&'static str>,
        body: Option<serde_json::Value>,
    }

    impl Transport for Chain {
        fn send(&self, req: &HttpRequest) -> std::result::Result<HttpResponse, String> {
            if req
                .url
                .starts_with("https://api.github.com/search/issues?q=")
            {
                return Ok(HttpResponse::new(200).with_body(
                    json!({"total_count": 1, "incomplete_results": false, "items": [{"number": 7}]}),
                ));
            }
            let base = format!("{API}/issues/7/comments?per_page=100");
            let page: u32 = req
                .url
                .strip_prefix(&format!("{base}&page="))
                .map_or(1, |p| p.parse().unwrap());
            let mut resp = HttpResponse::new(200);
            resp.body = Some(self.body.clone().unwrap_or_else(|| json!([])));
            if page < self.pages {
                resp =
                    resp.with_header("link", format!("<{base}&page={}>; rel=\"next\"", page + 1));
            } else if let Some(link) = self.last_link {
                resp = resp.with_header("link", link);
            }
            Ok(resp)
        }
    }

    /// Audit #3023 B2: fifty comment pages are read; a fifty-first that is still asked for is a
    /// read that failed, never a quiet "no claim".
    #[test]
    fn the_comment_page_limit_is_fifty() {
        let ok = Chain {
            pages: 50,
            last_link: None,
            body: None,
        };
        assert!(read_claims("acme", "widgets", "agent-9", &[], &ok, &Clock0).is_ok());
        let over = Chain {
            pages: 51,
            last_link: None,
            body: None,
        };
        let err = read_claims("acme", "widgets", "agent-9", &[], &over, &Clock0)
            .map(|_| ())
            .unwrap_err();
        assert!(err.contains("#7") && err.contains("limit"), "{err}");
    }

    /// Audit #3023 B1: an answer that is no list, or a next page that cannot be followed, is
    /// a read that did not reach the end.
    #[test]
    fn comments_that_were_not_read_to_the_end_are_an_error_not_no_claim() {
        let object = Chain {
            pages: 1,
            last_link: None,
            body: Some(json!({"message": "weird object"})),
        };
        let unfollowable = Chain {
            pages: 1,
            last_link: Some("garbage; rel=\"next\""),
            body: None,
        };
        let no_body = Fake {
            answers: vec![
                search(&[7], 1, false),
                // 200 with no JSON body at all (a truncated answer).
                (
                    format!("{API}/issues/7/comments?per_page=100"),
                    200,
                    json!(null),
                ),
            ],
            sent: RefCell::new(Vec::new()),
        };
        for (what, result) in [
            (
                "an object",
                read_claims("acme", "widgets", "agent-9", &[], &object, &Clock0),
            ),
            (
                "an unfollowable next link",
                read_claims("acme", "widgets", "agent-9", &[], &unfollowable, &Clock0),
            ),
            (
                "a null body",
                read_claims("acme", "widgets", "agent-9", &[], &no_body, &Clock0),
            ),
        ] {
            let err = result.map(|_| ()).unwrap_err();
            assert!(
                err.contains("#7") && err.contains("to the end"),
                "{what}: {err}"
            );
        }
    }

    /// Audit #3023 N3: a lane name or a repository that would change the search or the URL is
    /// refused before any request.
    #[test]
    fn a_name_that_would_change_the_request_is_refused_before_it_is_sent() {
        let fake = Fake {
            answers: vec![search(&[], 0, false)],
            sent: RefCell::new(Vec::new()),
        };
        for lane in ["x\" repo:evil/other \"y", "a b", "", "a/b"] {
            assert!(
                read_claims("acme", "widgets", lane, &[], &fake, &Clock0).is_err(),
                "{lane:?}"
            );
        }
        for (owner, repo) in [
            ("acme", "widgets?x=1#"),
            ("acme", ".."),
            ("", "w"),
            ("a b", "w"),
        ] {
            assert!(
                read_claims(owner, repo, "agent-9", &[], &fake, &Clock0).is_err(),
                "{owner}/{repo}"
            );
        }
        assert!(fake.sent.borrow().is_empty(), "{:?}", fake.sent.borrow());
    }

    #[test]
    fn the_search_text_is_percent_encoded() {
        assert_eq!(
            encode_query("repo:a/b is:open \"session=x\""),
            "repo%3Aa%2Fb%20is%3Aopen%20%22session%3Dx%22"
        );
        assert_eq!(encode_query("A-z_0.9~"), "A-z_0.9~");
    }

    #[test]
    fn the_documented_example_parses_and_wrong_shapes_name_the_schema() {
        let ok: Observations = serde_json::from_str(OBSERVATIONS_EXAMPLE).unwrap();
        assert_eq!(ok.schema_version, 1);
        assert!(ok.claims.is_some() && ok.session_cwd.is_some() && ok.screen.is_some());
        let msg = parse_observations(r#"{"schema_version":1,"screen":{"source":"x","observed_at":"2026-11-15T00:00:00Z","background_jobs":[]}}"#)
            .map(|_| ())
            .unwrap_err();
        assert!(msg.starts_with("invalid observations:"), "{msg}");
        assert!(msg.contains("screen.background_jobs: a number"), "{msg}");
        assert!(msg.contains(OBSERVATIONS_TYPES), "{msg}");
        assert!(msg.contains(OBSERVATIONS_EXAMPLE), "{msg}");
        // A well-formed file passes through unchanged.
        assert!(parse_observations(OBSERVATIONS_EXAMPLE).is_ok());
    }

    /// The command's `--help` text is its doc comment, copied by `build.rs`; the example and
    /// the field types it shows must be the ones the error text shows.
    #[test]
    fn the_help_text_of_the_command_carries_the_example_and_the_types() {
        let source = include_str!("ops/lane_safe_to_reset.rs");
        assert!(source.contains(&format!("/// {OBSERVATIONS_EXAMPLE}\n")));
        assert!(source.contains(&format!("/// Observations file: {OBSERVATIONS_TYPES}\n")));
    }

    #[test]
    fn local_worktrees_of_the_lane_name_the_issues_to_read() {
        let list = "worktree /r\nHEAD aaaa\nbranch refs/heads/main\n\n\
worktree /r/wt/agent-9-card\nHEAD bbbb\nbranch refs/heads/issue/7-card\n\n\
worktree /r/wt/other\nHEAD cccc\nbranch refs/heads/issue/8-other\n\n\
worktree /r/wt/x\nHEAD dddd\nbranch refs/heads/issue/9-card-agent9\n";
        assert_eq!(lane_branch_issues(list, "agent-9"), vec![7, 9]);
    }
}
