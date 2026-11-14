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

pub(crate) fn get_json<T: Transport>(
    t: &T,
    url: &str,
) -> Result<(serde_json::Value, Option<String>)> {
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
    claim_sessions(t, base_url, dir, &[&format!("{merge_base}..{head}")])
}

/// The sessions of the claims the commits `revs` select (a range, or `--no-walk` and shas)
/// name in `Moai-Claim` trailers, merges skipped.
fn claim_sessions<T: Transport>(
    t: &T,
    base_url: &str,
    dir: &Path,
    revs: &[&str],
) -> Result<Vec<String>> {
    let mut args = vec![
        "log",
        "--no-merges",
        "--format=%(trailers:key=Moai-Claim,valueonly,separator=%x0A)",
    ];
    args.extend_from_slice(revs);
    let out = git_ok(dir, &args)?;
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

/// The PR's own commits (merge base to head, merges skipped) whose message has no
/// `Moai-Claim` trailer as git reads trailers (G10). A claim line that is not in the
/// message's trailer block is not read, the same as `pr_authors` and the commit guard.
fn commits_without_claim(dir: &Path, merge_base: &str, head: &str) -> Result<Vec<String>> {
    unclaimed_in(dir, &[&format!("{merge_base}..{head}")])
}

/// [`commits_without_claim`] for the commits `revs` select (a range, or `--no-walk` and shas).
fn unclaimed_in(dir: &Path, revs: &[&str]) -> Result<Vec<String>> {
    let mut args = vec![
        "log",
        "--no-merges",
        "--format=%H%x09%(trailers:key=Moai-Claim,valueonly,separator=%x2C)",
    ];
    args.extend_from_slice(revs);
    let out = git_ok(dir, &args)?;
    Ok(out
        .lines()
        .filter_map(|l| {
            let (sha, claims) = l.split_once('\t')?;
            claims.trim().is_empty().then(|| sha.to_string())
        })
        .collect())
}

/// PR files a mechanical delta compares at most (#2833); four git calls per file.
pub const MAX_DELTA_FILES: usize = 200;

/// Pages of a pull request's commit list read for the stacked-PR rule (GitHub lists at most
/// 250 commits of a pull request, 100 per page).
const MAX_PR_COMMIT_PAGES: usize = 3;

/// Every page of a GitHub list, at most `max` pages.
fn get_all<T: Transport>(t: &T, url: &str, max: usize) -> Result<Vec<serde_json::Value>> {
    let mut items = Vec::new();
    let mut next = Some(url.to_string());
    let mut pages = 0;
    while let Some(u) = next {
        pages += 1;
        if pages > max {
            return Err(format!("more than {max} pages at {url}; refused"));
        }
        let (page, n) = get_json(t, &u)?;
        items.extend(
            page.as_array()
                .ok_or_else(|| format!("GET {u} is not a list"))?
                .iter()
                .cloned(),
        );
        next = n;
    }
    Ok(items)
}

/// #2859: the other open pull requests whose own commits this PR's head would bring into the
/// tip (`merge_base..head`), each with its last verdict and what G12 checks against it. A PR
/// below that already merged is part of the tip, so its commits are not brought in (D2).
fn stacked_below<T: Transport>(
    t: &T,
    base_url: &str,
    dir: &Path,
    pr: u64,
    merge_base: &str,
    head: &str,
) -> Result<Vec<gate::StackedBelow>> {
    let open = get_all(t, &format!("{base_url}/pulls?state=open&per_page=100"), 1)?;
    let others: Vec<u64> = open
        .iter()
        .filter_map(|p| p["number"].as_u64())
        .filter(|n| *n != pr)
        .collect();
    if others.is_empty() {
        return Ok(Vec::new());
    }
    let incoming = names(&git_ok(
        dir,
        &["rev-list", &format!("{merge_base}..{head}")],
    )?);
    let mut out = Vec::new();
    for n in others {
        let commits = get_all(
            t,
            &format!("{base_url}/pulls/{n}/commits?per_page=100"),
            MAX_PR_COMMIT_PAGES,
        )?;
        let theirs: Vec<&str> = commits.iter().filter_map(|c| c["sha"].as_str()).collect();
        let carried: Vec<&String> = incoming
            .iter()
            .filter(|c| theirs.contains(&c.as_str()))
            .collect();
        if carried.is_empty() {
            continue;
        }
        // The newest carried commits: those no other carried commit descends from.
        let mut brought = Vec::new();
        for c in &carried {
            let mut newest = true;
            for d in &carried {
                if c != d && is_ancestor(dir, c, d)? {
                    newest = false;
                    break;
                }
            }
            if newest {
                brought.push((*c).clone());
            }
        }
        let mut walk = vec!["--no-walk"];
        walk.extend(carried.iter().map(|c| c.as_str()));
        let authors = claim_sessions(t, base_url, dir, &walk)?;
        let unclaimed = unclaimed_in(dir, &walk)?;
        let comments = get_all(
            t,
            &format!("{base_url}/issues/{n}/comments?per_page=100"),
            MAX_PAGES,
        )?;
        let last = comments
            .iter()
            .filter_map(|c| {
                let body = c["body"].as_str()?;
                let id = c["id"].as_u64()?;
                body.lines()
                    .next()
                    .filter(|l| l.starts_with("AUDIT #"))
                    .map(|_| (id, body))
            })
            .max_by_key(|(id, _)| *id)
            .map(|(_, body)| body);
        let parsed = last.and_then(|b| b.lines().next().and_then(gate::parse_verdict));
        let verdict_head = parsed.as_ref().map(|(h, _, _)| h.clone());
        let verdict = match (&last, &parsed) {
            (_, Some((_, _, v))) => Some(v.clone()),
            (Some(_), None) => Some("unreadable".to_string()),
            _ => None,
        };
        let covered = match &verdict_head {
            Some(v) if is_commit(dir, v)? => {
                let mut all = true;
                for c in &brought {
                    if c != v && !is_ancestor(dir, c, v)? {
                        all = false;
                        break;
                    }
                }
                all
            }
            _ => false,
        };
        out.push(gate::StackedBelow {
            pr: n,
            brought,
            verdict_head,
            verdict,
            auditor: last.and_then(gate::auditor),
            authors,
            commits_without_claim: unclaimed,
            covered,
        });
    }
    Ok(out)
}

fn is_ancestor(dir: &Path, a: &str, b: &str) -> Result<bool> {
    Ok(git(dir, &["merge-base", "--is-ancestor", a, b])?.0)
}

/// `git patch-id --stable` of `file`'s change from `a` to `b` (line numbers and whitespace do
/// not count; the content of every added and removed line does). `--binary` keeps a binary
/// file's bytes in the patch, so a changed binary never reads as the same patch.
fn patch_id(dir: &Path, a: &str, b: &str, file: &str) -> Result<String> {
    let diff = crate::audit_delta::command(dir)
        .args(["diff", "--no-renames", "--binary", a, b, "--", file])
        .output()
        .map_err(|e| format!("cannot execute git: {e}"))?;
    if !diff.status.success() {
        return Err(format!("git diff {a} {b} -- {file} failed"));
    }
    let mut child = crate::audit_delta::command(dir)
        .args(["patch-id", "--stable"])
        .stdin(std::process::Stdio::piped())
        .stdout(std::process::Stdio::piped())
        .stderr(std::process::Stdio::null())
        .spawn()
        .map_err(|e| format!("cannot execute git: {e}"))?;
    {
        use std::io::Write;
        let mut stdin = child.stdin.take().ok_or("git patch-id has no stdin")?;
        stdin
            .write_all(&diff.stdout)
            .map_err(|e| format!("git patch-id stdin: {e}"))?;
    }
    let out = child
        .wait_with_output()
        .map_err(|e| format!("git patch-id failed: {e}"))?;
    Ok(String::from_utf8_lossy(&out.stdout)
        .split_whitespace()
        .next()
        .unwrap_or("")
        .to_string())
}

/// Whether `to` is the verdict head `from` with only the base branch merged in (#2833 scope
/// 1–3; scope 4, the CI on `to`, is G5's): `from` is an ancestor of `to` (AC-2); every commit
/// after `from` that the tip does not hold is a merge, so the PR's author added none; the PR's
/// file set against its merge base is the same on both heads, so a merge changed no file
/// outside it; and each file's patch-id is the same, so a resolution that changed a PR line
/// differs. `ok: false` carries the first reason.
pub fn mechanical_delta(dir: &Path, from: &str, to: &str, tip: &str) -> Result<gate::AutoDelta> {
    mechanical_delta_limited(dir, from, to, tip, MAX_DELTA_FILES)
}

/// [`mechanical_delta`] with the file limit as a parameter, so the limit's boundary can be
/// tested on a small repository.
#[doc(hidden)]
pub fn mechanical_delta_limited(
    dir: &Path,
    from: &str,
    to: &str,
    tip: &str,
    max_files: usize,
) -> Result<gate::AutoDelta> {
    let no = |detail: String| gate::AutoDelta {
        from: from.to_string(),
        to: to.to_string(),
        ok: false,
        detail,
        files: Vec::new(),
    };
    if !is_commit(dir, from)? {
        return Ok(no(format!("the verdict head {from} is not a local commit")));
    }
    if !git(dir, &["merge-base", "--is-ancestor", from, to])?.0 {
        return Ok(no(format!(
            "the verdict head {} is not an ancestor of the current head",
            short(from)
        )));
    }
    let own = git_ok(
        dir,
        &[
            "rev-list",
            "--no-merges",
            to,
            &format!("^{from}"),
            &format!("^{tip}"),
        ],
    )?;
    let own: Vec<&str> = own.lines().filter(|l| !l.is_empty()).collect();
    if !own.is_empty() {
        let list: Vec<&str> = own.iter().map(|c| short(c)).collect();
        return Ok(no(format!(
            "{} commit(s) after the verdict head are not merges of the base branch: {}",
            own.len(),
            list.join(", ")
        )));
    }
    let mb_from = git_ok(dir, &["merge-base", from, tip])?.trim().to_string();
    let mb_to = git_ok(dir, &["merge-base", to, tip])?.trim().to_string();
    let files_from = names(&git_ok(
        dir,
        &["diff", "--no-renames", "--name-only", &mb_from, from],
    )?);
    let files_to = names(&git_ok(
        dir,
        &["diff", "--no-renames", "--name-only", &mb_to, to],
    )?);
    if files_from != files_to {
        let extra: Vec<&String> = files_to
            .iter()
            .filter(|f| !files_from.contains(f))
            .collect();
        let gone: Vec<&String> = files_from
            .iter()
            .filter(|f| !files_to.contains(f))
            .collect();
        return Ok(no(format!(
            "the PR's files differ between the heads (only on the current head: {extra:?}; only on the verdict head: {gone:?})"
        )));
    }
    if files_from.is_empty() {
        return Ok(no("the PR changes no files".into()));
    }
    if files_from.len() > max_files {
        return Ok(no(format!(
            "{} PR files, more than the {max_files} a mechanical delta compares",
            files_from.len()
        )));
    }
    let mut files = Vec::new();
    for f in &files_from {
        let a = patch_id(dir, &mb_from, from, f)?;
        let b = patch_id(dir, &mb_to, to, f)?;
        if a.is_empty() || a != b {
            return Ok(no(format!(
                "{f}: the PR's change differs between the heads"
            )));
        }
        files.push(gate::DeltaFile {
            path: f.clone(),
            patch_id_from: a,
            patch_id_to: b,
        });
    }
    Ok(gate::AutoDelta {
        from: from.to_string(),
        to: to.to_string(),
        ok: true,
        detail: format!(
            "{} merges only the base branch into {}; {} PR file(s), patch-ids equal",
            short(to),
            short(from),
            files.len()
        ),
        files,
    })
}

fn short(sha: &str) -> &str {
    sha.get(..10).unwrap_or(sha)
}

/// The verdict first line for the head the auditor examined (#2833 D4), refused unless that
/// head is still the PR's head on GitHub right before printing. A verdict therefore never
/// names a head nobody audited (audit 6869947940 B1) nor an older one the gate would stop at
/// (#2821). When the head moved only by an update-branch merge of the base branch, the line
/// for the audited head is the right one: the gate carries it (AUTO-DELTA). The base is
/// `base` when given, else `git merge-base <head> <tip>` in the local repository (the head
/// must be fetched). Reads only.
pub fn verdict_line_for<T: Transport>(
    target: &Target<'_>,
    audited_head: &str,
    title: &str,
    verdict: &str,
    base: Option<&str>,
    t: &T,
) -> Result<serde_json::Value> {
    if !gate::is_full_sha(audited_head) {
        return Err("--head must be the audited head, 40 lowercase hex".into());
    }
    let base_url = format!(
        "https://api.github.com/repos/{}/{}",
        target.owner, target.repo
    );
    let (pr, _) = get_json(t, &format!("{base_url}/pulls/{}", target.pr))?;
    let now = pr["head"]["sha"]
        .as_str()
        .ok_or("pull request has no head sha")?
        .to_string();
    // The head moved during the audit. A line for the audited head is still right when the move
    // is a mechanical delta (the gate carries the PASS, AUTO-DELTA); anything else needs the
    // delta audited first, so no line is printed.
    let mut carried_to = None;
    if now != audited_head {
        let refuse = |why: String| {
            format!(
                "the pull request's head is now {now}, not the audited head {audited_head}, and {why}; no line printed — audit the delta {audited_head}..{now}, then run this again with --head {now}"
            )
        };
        if !is_commit(target.dir, &now)? {
            return Err(refuse(format!(
                "the new head is not in the local repository (fetch refs/pull/{}/head) to check whether only the base branch was merged in",
                target.pr
            )));
        }
        let d = mechanical_delta(target.dir, audited_head, &now, target.tip)?;
        if !d.ok {
            return Err(refuse(format!(
                "it is not a mechanical delta: {}",
                d.detail
            )));
        }
        carried_to = Some(now.clone());
    }
    let head = audited_head.to_string();
    let (base, from) = match base {
        Some(b) => (b.to_string(), "given"),
        None => {
            if !is_commit(target.dir, &head)? {
                return Err(format!(
                    "pull request head {head} is not in the local repository; fetch refs/pull/{}/head first, or pass --base",
                    target.pr
                ));
            }
            let tip = git_ok(
                target.dir,
                &[
                    "rev-parse",
                    "--verify",
                    &format!("{}^{{commit}}", target.tip),
                ],
            )
            .map_err(|_| format!("tip `{}` is not a local commit", target.tip))?;
            let mb = git_ok(target.dir, &["merge-base", &head, tip.trim()])?;
            (mb.trim().to_string(), "merge-base")
        }
    };
    let line = gate::format_verdict_line(target.pr, title, &head, &base, verdict)?;
    let mut out =
        serde_json::json!({ "line": line, "head": head, "base": base, "base_from": from });
    if let Some(to) = carried_to {
        // A FAIL is never carried; the gate carries only a PASS (G1).
        out["current_head"] = serde_json::json!(to);
        out["auto_delta"] = serde_json::json!(verdict == "PASS");
    }
    Ok(out)
}

/// Whether a PR body names a claim (G11): the marker the removed `flow-gate-verdict`
/// `pr_body_claim` check read, anywhere in the body.
pub fn has_body_claim(body: &str) -> bool {
    body.contains("Moai-Claim: ")
}

const LIB_RS: &str = "src/lib.rs";

/// A module declaration line, `mod x;` or `pub mod x;` (lowercase name).
fn is_mod_line(line: &str) -> bool {
    let rest = line.strip_prefix("pub ").unwrap_or(line);
    let Some(name) = rest.strip_prefix("mod ").and_then(|r| r.strip_suffix(';')) else {
        return false;
    };
    let mut chars = name.chars();
    chars
        .next()
        .is_some_and(|c| c.is_ascii_lowercase() || c == '_')
        && chars.all(|c| c.is_ascii_lowercase() || c.is_ascii_digit() || c == '_')
}

/// #2860 rule 2 (shell gate_merge.sh): the PR's own change to `src/lib.rs` is only added module
/// lines (`mod x;` / `pub mod x;`, any number, nothing removed or changed), and the merged tree
/// still has every one of them. Then a tip change to the same file is not a code overlap.
fn lib_mods_only_kept(dir: &Path, tree: &str, merge_base: &str, head: &str) -> Result<bool> {
    let diff = git_ok(
        dir,
        &["diff", "--no-color", "-U0", merge_base, head, "--", LIB_RS],
    )?;
    let changed: Vec<&str> = diff
        .lines()
        .filter(|l| {
            (l.starts_with('+') || l.starts_with('-'))
                && !l.starts_with("+++")
                && !l.starts_with("---")
        })
        .collect();
    let added: Vec<&str> = changed.iter().filter_map(|l| l.strip_prefix('+')).collect();
    if changed.is_empty() || added.len() != changed.len() || !added.iter().all(|l| is_mod_line(l)) {
        return Ok(false);
    }
    let (ok, merged) = git(dir, &["show", &format!("{tree}:{LIB_RS}")])?;
    Ok(ok && added.iter().all(|l| merged.lines().any(|m| m == *l)))
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

/// Earlier tip commits read at most when the tip's own push CI has not completed (G8).
pub const TIP_WALK: usize = 10;
/// The check the tip must pass (both spellings of the aggregate, `flow::CI_CHECK_NAMES`).
const TIP_REQUIRED: &str = "ci";

/// G8's judgment of one commit's check runs (every attempt): [`crate::flow::tip_ci_judgment`]
/// over the runs that completed with something other than `cancelled`. A cancelled run says
/// nothing about the code (a newer push cancels the older run), and a running attempt has no
/// result yet: a commit whose CI was only cancelled or is still running has no completed CI
/// and reads `pending`, while a completed failure stays `red` while its re-run runs.
pub fn completed_judgment(runs: &[serde_json::Value]) -> serde_json::Value {
    let kept: Vec<serde_json::Value> = runs
        .iter()
        .filter(|r| {
            r["conclusion"]
                .as_str()
                .is_some_and(|c| !c.is_empty() && c != "cancelled")
        })
        .cloned()
        .collect();
    crate::flow::tip_ci_judgment(&kept, &[TIP_REQUIRED.to_string()])
}

/// #2554 D8(a) (#2860, the judgment #2856 `lane-next` uses too): the aggregate check
/// (`ci`/`rc-ci`) is not judged on its own — a newer push cancels the older run's waiting jobs
/// and the aggregate then reads `failure` with every real job passed or cancelled — and a
/// cancelled job says nothing about the code. The remaining jobs are judged with
/// [`crate::flow::tip_ci_judgment`]: `red` only when a job itself failed with no success on the
/// same commit; no job left reads `pending`.
pub fn d8a_judgment(runs: &[serde_json::Value]) -> serde_json::Value {
    let kept: Vec<serde_json::Value> = runs
        .iter()
        .filter(|r| {
            let name = r["name"].as_str().unwrap_or("");
            !crate::flow::CI_CHECK_NAMES.contains(&name)
                && r["conclusion"].as_str() != Some("cancelled")
        })
        .cloned()
        .collect();
    let mut names: Vec<String> = kept
        .iter()
        .filter_map(|r| r["name"].as_str().map(str::to_string))
        .collect();
    names.sort();
    names.dedup();
    if names.is_empty() {
        return serde_json::json!({"state": "pending"});
    }
    crate::flow::tip_ci_judgment(&kept, &names)
}

/// G8's judgment of one commit: [`d8a_judgment`] whenever the commit has any job besides the
/// aggregate (a cancelled one included). A commit whose CI is the aggregate alone — a project
/// whose workflow has a single job named `ci` — is judged by that check ([`completed_judgment`]),
/// so D8(a) never reads such a project as forever pending.
fn g8_judgment(runs: &[serde_json::Value]) -> serde_json::Value {
    let has_jobs = runs.iter().any(|r| {
        r["name"]
            .as_str()
            .is_some_and(|n| !crate::flow::CI_CHECK_NAMES.contains(&n))
    });
    if has_jobs {
        d8a_judgment(runs)
    } else {
        completed_judgment(runs)
    }
}

/// The cancelled jobs on one commit that D8(a) does not excuse for the merge gate (#2860,
/// leader tool `green_tip.py`): a cancelled job with no success on the same commit counts only
/// when it is a macOS job that succeeded on a newer tip commit (`newer_ok`); anything else
/// leaves the commit undecided, and G8 reads the next older one.
fn unexcused_cancels(runs: &[serde_json::Value], newer_ok: &[String]) -> Vec<String> {
    let succeeded = |name: &str| {
        runs.iter().any(|r| {
            r["name"].as_str() == Some(name) && r["conclusion"].as_str() == Some("success")
        })
    };
    let mut out: Vec<String> = runs
        .iter()
        .filter(|r| r["conclusion"].as_str() == Some("cancelled"))
        .filter_map(|r| r["name"].as_str())
        .filter(|n| !crate::flow::CI_CHECK_NAMES.contains(n))
        .filter(|n| !succeeded(n))
        .filter(|n| !(n.contains("macos") && newer_ok.iter().any(|ok| ok == n)))
        .map(str::to_string)
        .collect();
    out.sort();
    out.dedup();
    out
}

/// Every attempt's check runs on `sha` (`filter=all`, so a failure a re-run turned green stays
/// visible), at most ten pages.
pub(crate) fn tip_runs<T: Transport>(
    t: &T,
    base_url: &str,
    sha: &str,
) -> Result<Vec<serde_json::Value>> {
    let mut out = Vec::new();
    let mut url = Some(format!(
        "{base_url}/commits/{sha}/check-runs?filter=all&per_page=100"
    ));
    let mut pages = 0;
    while let Some(u) = url {
        pages += 1;
        if pages > 10 {
            return Err("more than 10 check run pages".into());
        }
        let (page, next) = get_json(t, &u)?;
        out.extend(
            page["check_runs"]
                .as_array()
                .ok_or("check runs response has no list")?
                .iter()
                .cloned(),
        );
        url = next;
    }
    Ok(out)
}

/// G8 input: the tip's state from its last completed push CI. The tip is judged first; while
/// its CI runs or was only cancelled, the earlier tip commits (first parents, at most
/// [`TIP_WALK`]) are read in turn and the first with completed CI decides. A read that fails,
/// or no completed CI within the walk, is `unknown`.
pub(crate) fn tip_ci<T: Transport>(
    t: &T,
    base_url: &str,
    dir: &Path,
    tip: &str,
) -> Result<gate::TipCi> {
    let walk = git_ok(
        dir,
        &[
            "rev-list",
            "--first-parent",
            &format!("--max-count={TIP_WALK}"),
            tip,
        ],
    )?;
    let unknown = |detail: String| gate::TipCi {
        state: "unknown".into(),
        tip: tip.to_string(),
        decided_by: None,
        detail,
    };
    let mut waited = Vec::new();
    // Jobs that succeeded on a newer tip commit already read (newest first): a macOS job a
    // newer push cancelled here counts as passed when one of these is it (#2860, D8(a)).
    let mut newer_ok: Vec<String> = Vec::new();
    for sha in walk.lines().filter(|l| !l.is_empty()) {
        let runs = match tip_runs(t, base_url, sha) {
            Ok(runs) => runs,
            Err(e) => return Ok(unknown(format!("check runs of {sha} unreadable: {e}"))),
        };
        // #2860: D8(a) instead of the aggregate `ci` — the aggregate fails whenever a newer
        // push cancels a waiting macOS job, which would read a green tip as red.
        let j = g8_judgment(&runs);
        let mut state = j["state"].as_str().unwrap_or("unknown").to_string();
        if state != "red" && !unexcused_cancels(&runs, &newer_ok).is_empty() {
            state = "pending".into();
        }
        newer_ok.extend(
            runs.iter()
                .filter(|r| r["conclusion"].as_str() == Some("success"))
                .filter_map(|r| r["name"].as_str().map(str::to_string)),
        );
        if state == "pending" {
            waited.push(sha.get(..10).unwrap_or(sha).to_string());
            continue;
        }
        let detail = match state.as_str() {
            "red" => j["failed"]
                .as_array()
                .map(|f| {
                    f.iter()
                        .map(|r| {
                            format!(
                                "{} {} run {}",
                                r["check"].as_str().unwrap_or("?"),
                                r["conclusion"].as_str().unwrap_or("?"),
                                r["run_id"].as_u64().map_or("?".into(), |n| n.to_string())
                            )
                        })
                        .collect::<Vec<_>>()
                        .join(", ")
                })
                .unwrap_or_default(),
            other => format!("{other} on {}", sha.get(..10).unwrap_or(sha)),
        };
        return Ok(gate::TipCi {
            state,
            tip: tip.to_string(),
            decided_by: Some(sha.to_string()),
            detail,
        });
    }
    Ok(unknown(format!(
        "no completed push CI on the last {} tip commit(s) ({} running or only cancelled)",
        waited.len(),
        waited.join(", ")
    )))
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
    let unclaimed = commits_without_claim(dir, &merge_base, &head)?;

    // The verdict base is read from the last verdict the gate will judge; without one G1 stops.
    let last_line = comments
        .iter()
        .filter_map(|c| {
            let first = c.body.lines().next()?;
            first.starts_with("AUDIT #").then_some((c.id, first))
        })
        .max_by_key(|(id, _)| *id)
        .map(|(_, line)| line);
    // #2860 rule 3 (shell gate_merge.sh): a short verdict base counts when it names exactly one
    // commit here. `rev-parse --verify --quiet <b>^{commit}` refuses an ambiguous or unknown
    // prefix; the full SHA then stands for the base everywhere the gate reads it.
    let mut verdict_base_full = None;
    let mut last_verdict = last_line.and_then(gate::parse_verdict);
    if last_verdict.is_none() {
        if let Some((vh, short, v)) = last_line.and_then(gate::parse_verdict_short_base) {
            let (ok, out) = git(
                dir,
                &[
                    "rev-parse",
                    "--verify",
                    "--quiet",
                    &format!("{short}^{{commit}}"),
                ],
            )?;
            let full = out.trim().to_string();
            if ok && full.len() == 40 && full.starts_with(&short) {
                verdict_base_full = Some(full.clone());
                last_verdict = Some((vh, full, v));
            }
        }
    }
    let verdict_base = last_verdict.as_ref().map(|(_, base, _)| base.clone());
    // #2833: a PASS on an earlier head may still count when the current head only merged the
    // base branch into it. Measured only then; any other case leaves it absent.
    let auto_delta = match &last_verdict {
        Some((vhead, _, v)) if v == "PASS" && *vhead != head => {
            Some(mechanical_delta(dir, vhead, &head, &tip)?)
        }
        _ => None,
    };
    let verdict_base_known = match &verdict_base {
        Some(b) => is_commit(dir, b)?,
        None => false,
    };
    // What the tip changed since the verdict: measured from where the verdict's base meets the
    // tip. A verdict base that is an earlier PR head (not on the tip) would otherwise count the
    // PR's own changes, undone, as tip changes — a false overlap.
    // #2833: with a mechanical delta the current head already holds the tip it merged, so the
    // tip's changes count from where the current head meets the tip; otherwise the overlap G6
    // reads would be the very changes the update-branch merged in.
    let delta_ok = auto_delta.as_ref().is_some_and(|d| d.ok);
    let tip_since = match (&verdict_base, verdict_base_known) {
        (Some(_), true) if delta_ok => Some(merge_base.clone()),
        (Some(b), true) => Some(git_ok(dir, &["merge-base", b, &tip])?.trim().to_string()),
        _ => None,
    };
    let tip_files = match &tip_since {
        Some(from) => names(&git_ok(dir, &["diff", "--name-only", from, &tip])?),
        None => Vec::new(),
    };

    let mut docs_lines_kept = Vec::new();
    let mut merged_manifest_ok = false;
    let mut lib_mod_only_kept = false;
    if merge_clean {
        for f in pr_files.iter().filter(|f| tip_files.contains(f)) {
            if f == MANIFEST {
                merged_manifest_ok = manifest_matches(dir, &merged_tree)?;
            } else if f == LIB_RS {
                lib_mod_only_kept = lib_mods_only_kept(dir, &merged_tree, &merge_base, &head)?;
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

    // #2859: open pull requests below this one, read before the tip so G8 stays last.
    let stacked = stacked_below(t, &base_url, dir, target.pr, &merge_base, &head)?;

    // G8 is read last: every earlier read is unchanged, and a pull request the gate cannot read
    // locally never costs the tip reads.
    let tip_ci = tip_ci(t, &base_url, dir, &tip)?;

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
        lib_mod_only_kept,
        review_confirmed_prs: target.review_confirmed_prs.clone(),
        body_check_ok: body_check(&body),
        verdict_base_known,
        verdict_base_full,
        pr_authors: Some(authors),
        tip_ci: Some(tip_ci),
        tip_repair: crate::flow::is_tip_repair(&body),
        commits_without_claim: Some(unclaimed),
        body_claim: has_body_claim(&body),
        auto_delta,
        stacked_below: Some(stacked),
    })
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::process::Command;

    fn git_in(dir: &Path, args: &[&str]) -> String {
        // Every git subprocess in src/ is scrubbed (#2337; tests/git_env_scrub.rs).
        let mut cmd = Command::new("git");
        crate::identity::scrub_git_env(&mut cmd);
        let out = cmd
            .args([
                "-c",
                "user.name=t",
                "-c",
                "user.email=t@example.invalid",
                "-c",
                "gc.auto=0",
                "-c",
                "maintenance.auto=false",
                "-c",
                "commit.gpgsign=false",
            ])
            .args(args)
            .current_dir(dir)
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .output()
            .unwrap();
        assert!(out.status.success(), "git {args:?}");
        String::from_utf8_lossy(&out.stdout).trim().to_string()
    }

    fn commit_lib(dir: &Path, text: &str, msg: &str) -> String {
        std::fs::create_dir_all(dir.join("src")).unwrap();
        std::fs::write(dir.join("src/lib.rs"), text).unwrap();
        git_in(dir, &["add", "-A"]);
        git_in(dir, &["commit", "-q", "-m", msg]);
        git_in(dir, &["rev-parse", "HEAD"])
    }

    /// #2869 audit B2: the PR's added module lines count only while the tree they are checked
    /// against still has them. Checked against a tree without the PR's `pub mod b;` (the tip's,
    /// standing in for a merge that lost the line), the overlap is not excused.
    #[test]
    fn a_module_line_missing_from_the_tree_is_not_kept() {
        let tmp = tempfile::tempdir().unwrap();
        let d = tmp.path();
        git_in(d, &["init", "-q", "-b", "main"]);
        let base = commit_lib(d, "pub mod a;\n\npub mod z;\n", "base");
        git_in(d, &["checkout", "-q", "-b", "pr"]);
        let head = commit_lib(d, "pub mod a;\npub mod b;\n\npub mod z;\n", "pr");
        git_in(d, &["checkout", "-q", "main"]);
        let tip = commit_lib(d, "pub mod a;\n\npub mod z;\npub mod zz;\n", "tip");
        let tree_of = |c: &str| git_in(d, &["rev-parse", &format!("{c}^{{tree}}")]);
        // The head's own tree has the line: kept.
        assert!(lib_mods_only_kept(d, &tree_of(&head), &base, &head).unwrap());
        // The tip's tree lacks it: not kept, so G6 stops on the overlap.
        assert!(!lib_mods_only_kept(d, &tree_of(&tip), &base, &head).unwrap());
    }
}
