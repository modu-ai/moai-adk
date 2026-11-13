//! Gathering a merge gate snapshot (Issue #2674 S1b): what `target/leader/gate_merge.sh` read
//! with `gh` and `git` before deciding, now read by the product. GitHub is read through a
//! [`Transport`] (GET only); git runs through the scrubbed `audit_delta::command`. Nothing is
//! fetched, posted or written in the repository: the pull request head and the tip must already
//! be present locally, and a missing one is an error that names what to fetch.

use crate::forge::policy::next_link;
use crate::forge::transport::{HttpRequest, Transport};
use crate::gate::{self, CheckRun, Comment, Snapshot};
use sha2::{Digest, Sha256};
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

const MANIFEST: &str = "skills/moai/references/source-manifest.json";
/// Comment pages read at most (100 per page).
const MAX_PAGES: usize = 30;

/// Where to read: the repository on disk, the GitHub repository, the pull request and the tip.
pub struct Target<'a> {
    pub dir: &'a Path,
    pub owner: &'a str,
    pub repo: &'a str,
    pub pr: u64,
    /// A local ref or SHA for the integration tip (for example `refs/remotes/origin/v2026.11.0`).
    pub tip: &'a str,
    /// PR numbers whose REVIEW.md draft the user confirmed.
    pub review_confirmed_prs: Vec<u64>,
}

fn get_json<T: Transport>(t: &T, url: &str) -> Result<(serde_json::Value, Option<String>)> {
    let resp = t.send(&HttpRequest::new("GET", url))?;
    if resp.status != 200 {
        return Err(format!("GET {url} returned HTTP {}", resp.status));
    }
    let next = next_link(resp.header("link"));
    Ok((resp.body.unwrap_or(serde_json::Value::Null), next))
}

fn git(dir: &Path, args: &[&str]) -> Result<(bool, String)> {
    let out = crate::audit_delta::command(dir)
        .args(args)
        .output()
        .map_err(|e| format!("cannot execute git: {e}"))?;
    Ok((
        out.status.success(),
        String::from_utf8_lossy(&out.stdout).into_owned(),
    ))
}

fn git_ok(dir: &Path, args: &[&str]) -> Result<String> {
    match git(dir, args)? {
        (true, out) => Ok(out),
        (false, _) => Err(format!("git {} failed", args.join(" "))),
    }
}

/// Claims named by `Moai-Claim` trailers read at most per pull request.
const MAX_CLAIMS: usize = 20;

/// The sessions of the claims the PR's own commits name (`Moai-Claim: <issue>#<comment id>`,
/// merge base to head, merges skipped): each claim comment's first line
/// `moai-claim: session=<s> …` (G9, #2674 R1).
fn pr_authors<T: Transport>(
    t: &T,
    base_url: &str,
    dir: &Path,
    merge_base: &str,
    head: &str,
) -> Result<Vec<String>> {
    let out = git_ok(
        dir,
        &[
            "log",
            "--no-merges",
            "--format=%(trailers:key=Moai-Claim,valueonly,separator=%x0A)",
            &format!("{merge_base}..{head}"),
        ],
    )?;
    let mut ids: Vec<u64> = Vec::new();
    for value in out.lines().map(str::trim).filter(|l| !l.is_empty()) {
        let id = value
            .split_once('#')
            .and_then(|(_, id)| id.trim().parse::<u64>().ok())
            .ok_or_else(|| format!("a Moai-Claim trailer does not read `<issue>#<id>`: {value}"))?;
        if !ids.contains(&id) {
            ids.push(id);
        }
    }
    if ids.len() > MAX_CLAIMS {
        return Err(format!(
            "more than {MAX_CLAIMS} claims in the PR's trailers; refused"
        ));
    }
    let mut sessions = Vec::new();
    for id in ids {
        let (c, _) = get_json(t, &format!("{base_url}/issues/comments/{id}"))?;
        let first = c["body"]
            .as_str()
            .unwrap_or("")
            .lines()
            .next()
            .unwrap_or("");
        let session = first
            .strip_prefix("moai-claim: ")
            .and_then(|rest| {
                rest.split_whitespace()
                    .find_map(|w| w.strip_prefix("session="))
            })
            .filter(|s| !s.is_empty())
            .ok_or_else(|| format!("claim comment {id} has no `moai-claim: session=` line"))?;
        if !sessions.iter().any(|s: &String| s == session) {
            sessions.push(session.to_string());
        }
    }
    Ok(sessions)
}

fn is_commit(dir: &Path, sha: &str) -> Result<bool> {
    Ok(git(dir, &["cat-file", "-e", &format!("{sha}^{{commit}}")])?.0)
}

fn names(out: &str) -> Vec<String> {
    let mut v: Vec<String> = out
        .lines()
        .filter(|l| !l.is_empty())
        .map(String::from)
        .collect();
    v.sort();
    v
}

/// Lines `git diff -U0 from to -- file` adds (without the leading `+`), blank lines dropped.
fn added_lines(dir: &Path, from: &str, to: &str, file: &str) -> Result<Vec<String>> {
    let out = git_ok(
        dir,
        &[
            "diff",
            "--no-color",
            "--no-ext-diff",
            "-U0",
            from,
            to,
            "--",
            file,
        ],
    )?;
    Ok(out
        .lines()
        .filter(|l| l.starts_with('+') && !l.starts_with("+++"))
        .map(|l| l[1..].to_string())
        .filter(|l| !l.trim().is_empty())
        .collect())
}

/// Whether `file` in the merged `tree` still holds every non-blank line each `(from, to)` side
/// added to it — a documentation overlap the gate accepts only when nothing was lost.
pub fn lines_kept(dir: &Path, tree: &str, file: &str, sides: &[(&str, &str)]) -> Result<bool> {
    let merged = git_ok(dir, &["show", &format!("{tree}:{file}")])?;
    let merged_lines: Vec<&str> = merged.lines().collect();
    for (from, to) in sides {
        if !added_lines(dir, from, to, file)?
            .iter()
            .all(|l| merged_lines.contains(&l.as_str()))
        {
            return Ok(false);
        }
    }
    Ok(true)
}

/// Whether every skill file in `tree` has a manifest entry with its sha256, and every entry a file.
pub fn manifest_matches(dir: &Path, tree: &str) -> Result<bool> {
    let listing = git_ok(
        dir,
        &["ls-tree", "-r", "--name-only", tree, "--", "skills/moai"],
    )?;
    let files: Vec<&str> = listing.lines().filter(|l| *l != MANIFEST).collect();
    let manifest: serde_json::Value =
        serde_json::from_str(&git_ok(dir, &["show", &format!("{tree}:{MANIFEST}")])?)
            .map_err(|_| "merged source manifest is not JSON".to_string())?;
    let entries = manifest["files"]
        .as_object()
        .ok_or("manifest has no files")?;
    if entries.len() != files.len() {
        return Ok(false);
    }
    for f in files {
        let rel = f.strip_prefix("skills/moai/").unwrap_or(f);
        let bytes = crate::audit_delta::command(dir)
            .args(["show", &format!("{tree}:{f}")])
            .output()
            .map_err(|e| format!("cannot execute git: {e}"))?
            .stdout;
        let sha = format!("{:x}", Sha256::digest(&bytes));
        if entries.get(rel).and_then(|e| e["sha256"].as_str()) != Some(sha.as_str()) {
            return Ok(false);
        }
    }
    Ok(true)
}

/// The PR body Self-Audit check (`moai-cli pr-self-audit-check`, #2708). It used to run the
/// repository's `tools/check_pr_self_audit.py` and failed G4 where that file was missing; the
/// check is now this binary's own, the same in every repository.
fn body_check(body: &str) -> bool {
    crate::self_audit::check(body, false).is_ok()
}

/// Reads one pull request into a [`Snapshot`] for [`gate::check`].
pub fn collect<T: Transport>(target: &Target<'_>, t: &T) -> Result<Snapshot> {
    let base_url = format!(
        "https://api.github.com/repos/{}/{}",
        target.owner, target.repo
    );
    let (pr, _) = get_json(t, &format!("{base_url}/pulls/{}", target.pr))?;
    let head = pr["head"]["sha"]
        .as_str()
        .ok_or("pull request has no head sha")?
        .to_string();
    let body = pr["body"].as_str().unwrap_or("").to_string();

    let mut comments = Vec::new();
    let mut url = Some(format!(
        "{base_url}/issues/{}/comments?per_page=100",
        target.pr
    ));
    let mut pages = 0;
    while let Some(u) = url {
        pages += 1;
        if pages > MAX_PAGES {
            return Err(format!("more than {MAX_PAGES} comment pages; refused"));
        }
        let (page, next) = get_json(t, &u)?;
        for c in page.as_array().ok_or("comments page is not a list")? {
            comments.push(Comment {
                id: c["id"].as_u64().ok_or("comment without id")?,
                body: c["body"].as_str().unwrap_or("").to_string(),
            });
        }
        url = next;
    }

    let (runs, _) = get_json(
        t,
        &format!("{base_url}/commits/{head}/check-runs?per_page=100"),
    )?;
    let checks = runs["check_runs"]
        .as_array()
        .ok_or("check runs response has no list")?
        .iter()
        .map(|r| CheckRun {
            name: r["name"].as_str().unwrap_or("").to_string(),
            conclusion: r["conclusion"].as_str().map(String::from),
        })
        .collect();

    let dir = target.dir;
    if !is_commit(dir, &head)? {
        return Err(format!(
            "pull request head {head} is not in the local repository; fetch refs/pull/{}/head first",
            target.pr
        ));
    }
    let tip = git_ok(
        dir,
        &[
            "rev-parse",
            "--verify",
            &format!("{}^{{commit}}", target.tip),
        ],
    )
    .map_err(|_| format!("tip `{}` is not a local commit", target.tip))?
    .trim()
    .to_string();
    let (merge_clean, merged_tree) = match git(dir, &["merge-tree", "--write-tree", &tip, &head])? {
        (true, out) => (true, out.lines().next().unwrap_or("").to_string()),
        (false, _) => (false, String::new()),
    };
    let merge_base = git_ok(dir, &["merge-base", &tip, &head])?
        .trim()
        .to_string();
    let pr_files = names(&git_ok(dir, &["diff", "--name-only", &merge_base, &head])?);
    let authors = pr_authors(t, &base_url, dir, &merge_base, &head)?;

    // The verdict base is read from the last verdict the gate will judge; without one G1 stops.
    let verdict_base = comments
        .iter()
        .filter_map(|c| {
            let first = c.body.lines().next()?;
            first.starts_with("AUDIT #").then_some((c.id, first))
        })
        .max_by_key(|(id, _)| *id)
        .and_then(|(_, line)| gate::parse_verdict(line))
        .map(|(_, base, _)| base);
    let verdict_base_known = match &verdict_base {
        Some(b) => is_commit(dir, b)?,
        None => false,
    };
    // What the tip changed since the verdict: measured from where the verdict's base meets the
    // tip. A verdict base that is an earlier PR head (not on the tip) would otherwise count the
    // PR's own changes, undone, as tip changes — a false overlap.
    let tip_since = match (&verdict_base, verdict_base_known) {
        (Some(b), true) => Some(git_ok(dir, &["merge-base", b, &tip])?.trim().to_string()),
        _ => None,
    };
    let tip_files = match &tip_since {
        Some(from) => names(&git_ok(dir, &["diff", "--name-only", from, &tip])?),
        None => Vec::new(),
    };

    let mut docs_lines_kept = Vec::new();
    let mut merged_manifest_ok = false;
    if merge_clean {
        for f in pr_files.iter().filter(|f| tip_files.contains(f)) {
            if f == MANIFEST {
                merged_manifest_ok = manifest_matches(dir, &merged_tree)?;
            } else if f.starts_with("docs/") {
                let from = tip_since.clone().unwrap_or_default();
                let sides = [
                    (merge_base.as_str(), head.as_str()),
                    (from.as_str(), tip.as_str()),
                ];
                if lines_kept(dir, &merged_tree, f, &sides)? {
                    docs_lines_kept.push(f.clone());
                }
            }
        }
    }

    Ok(Snapshot {
        schema_version: 1,
        pr: target.pr,
        head,
        comments,
        checks,
        pr_files,
        tip_files_since_verdict_base: tip_files,
        merge_clean,
        docs_lines_kept,
        merged_manifest_ok,
        review_confirmed_prs: target.review_confirmed_prs.clone(),
        body_check_ok: body_check(&body),
        verdict_base_known,
        pr_authors: Some(authors),
    })
}
