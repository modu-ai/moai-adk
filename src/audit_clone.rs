//! `audit.clone.plan` (WT-3, Issue #2380; design `docs/worktree-strategy-2026-11-10.md` §D4).
//! Read-only: builds the commands of an independent audit clone and where its evidence goes;
//! it runs none of them and touches no network.
//!
//! The clone takes only the pull request's head branch without a checkout, then fetches the
//! base branch for the merge base and checks out the exact head detached (§D4: about 5 MiB of
//! `.git` instead of the full history's ~470 MiB, still independent of the author's object
//! store). The clone names its remote `origin` explicitly and copies a local source with
//! `--no-local`, so no alternates are inherited; a relative local origin is resolved against
//! the project's top-level folder. The folder is `<tmp>/moai-audit-<pr>-<head8>`, the name
//! `worktree-report` (WT-1) counts as an audit clone. Evidence is kept in `~/.moai-cli/projects/<project_id>/audits/
//! <pr>-<head8>/` (user decision 2026-11-10, #2389 comment 6774899518 row 32, Q6 B); without a
//! project profile the evidence folder is unknown and the answer says why.

use crate::identity::scrub_git_env;
use serde_json::{json, Value};
use std::path::{Path, PathBuf};
use std::process::Command;

type Result<T> = std::result::Result<T, String>;

/// The inputs of one audit clone.
pub struct Request<'a> {
    pub pr: u64,
    pub head: &'a str,
    pub head_branch: &'a str,
    pub base_branch: &'a str,
    pub repo_url: Option<&'a str>,
    pub tmp_root: &'a Path,
}

fn is_sha(s: &str) -> bool {
    s.len() == 40
        && s.bytes()
            .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b))
}

/// Stdout of a successful git call, byte for byte.
fn git_bytes(dir: &Path, args: &[&str]) -> Option<Vec<u8>> {
    let mut cmd = Command::new("git");
    scrub_git_env(&mut cmd);
    let out = cmd
        .arg("--no-optional-locks")
        .arg("-C")
        .arg(dir)
        .args(args)
        .output()
        .ok()?;
    out.status.success().then_some(out.stdout)
}

/// Trimmed stdout, for git answers whose surrounding whitespace carries no meaning.
fn git_out(dir: &Path, args: &[&str]) -> Option<String> {
    git_bytes(dir, args).map(|b| String::from_utf8_lossy(&b).trim().to_string())
}

/// True when `name` is a valid branch name for git (`git check-ref-format --branch`, offline).
fn branch_ok(project: &Path, name: &str) -> bool {
    !name.starts_with('-') && git_out(project, &["check-ref-format", "--branch", name]).is_some()
}

/// True when git reads `url` as a local path: no `scheme://` and not the scp-like
/// `host:path` form (a colon before the first slash).
fn is_local_path(url: &str) -> bool {
    if url.contains("://") {
        return false;
    }
    match url.find(':') {
        Some(i) => url[..i].contains('/'),
        None => true,
    }
}

/// True when `raw`, a repository URL as the user or the remote config wrote it, can be
/// planned: not empty, no whitespace (space, tab, CR, LF, …), no other control character, not
/// starting with `-`. Only the written value is checked; a path this tool builds from it (see
/// `origin_url`) may contain the spaces of the project's own folder and is passed as one argv
/// element.
fn usable_url(raw: &str) -> bool {
    !raw.is_empty()
        && !raw.chars().any(|c| c.is_whitespace() || c.is_control())
        && !raw.starts_with('-')
}

const URL_NOT_USABLE: &str = "audit clone refused: the repository URL is not usable";

/// The default origin as (written value, URL for the clone command). The written value is
/// read without loss: `git remote get-url` prints the configured URL and one newline, and only
/// that newline is removed, so leading or trailing spaces stay and are refused by
/// `usable_url` instead of silently naming a different repository. Git resolves a relative
/// local path in a remote against the repository's top-level folder, but the planned commands
/// run from anywhere, so a relative path is made absolute against that folder here.
fn origin_url(project: &Path) -> Result<(String, String)> {
    let mut raw = git_bytes(project, &["remote", "get-url", "origin"])
        .ok_or("audit clone refused: no --repo-url and no origin remote")?;
    if raw.last() == Some(&b'\n') {
        raw.pop();
    }
    let url = String::from_utf8(raw).map_err(|_| format!("{URL_NOT_USABLE} (not UTF-8)"))?;
    if !usable_url(&url) || !is_local_path(&url) || Path::new(&url).is_absolute() {
        return Ok((url.clone(), url));
    }
    let top = git_out(project, &["rev-parse", "--show-toplevel"])
        .map(PathBuf::from)
        .unwrap_or_else(|| project.to_path_buf());
    let joined = top.join(&url);
    let resolved = std::fs::canonicalize(&joined).unwrap_or(joined);
    Ok((url, resolved.to_string_lossy().into_owned()))
}

/// The evidence folder for `project_id` under `home_state_root`.
pub fn evidence_dir(home_state_root: &Path, project_id: &str, pr: u64, head: &str) -> PathBuf {
    home_state_root
        .join("projects")
        .join(project_id)
        .join("audits")
        .join(format!("{pr}-{}", &head[..8]))
}

pub fn plan(project: &Path, req: &Request) -> Result<Value> {
    if req.pr == 0 {
        return Err("audit clone refused: --pr must be a pull request number".into());
    }
    if !is_sha(req.head) {
        return Err("audit clone refused: --head must be a 40-hex lowercase commit".into());
    }
    for (what, name) in [
        ("--head-branch", req.head_branch),
        ("--base-branch", req.base_branch),
    ] {
        if !branch_ok(project, name) {
            return Err(format!(
                "audit clone refused: {what} is not a valid branch name"
            ));
        }
    }
    let (written, url) = match req.repo_url {
        Some(u) => (u.to_string(), u.to_string()),
        None => origin_url(project)?,
    };
    if !usable_url(&written) {
        return Err(URL_NOT_USABLE.into());
    }
    let head8 = &req.head[..8];
    let dir = req.tmp_root.join(format!("moai-audit-{}-{head8}", req.pr));
    let dir_s = dir.to_string_lossy().into_owned();
    let base_ref = format!("refs/remotes/origin/{}", req.base_branch);
    let id = crate::identity::identity(project)?;
    let (evidence, evidence_reason) = match (
        id["project_id"].as_str(),
        id["home_state_root"].as_str(),
    ) {
        (Some(pid), Some(home)) => (
            Value::String(
                evidence_dir(Path::new(home), pid, req.pr, req.head)
                    .to_string_lossy()
                    .into_owned(),
            ),
            Value::Null,
        ),
        (None, _) => (
            Value::Null,
            json!("project profile has no project_id; evidence stays in the verdict comment until it has one"),
        ),
        (_, None) => (Value::Null, json!("HOME unavailable")),
    };
    // `--origin origin`: the fetch and `refs/remotes/origin/…` below need that remote name,
    // whatever `clone.defaultRemoteName` says. `--no-local`: a local source is copied over the
    // git transport, so the clone never inherits its alternates or hardlinks its objects and
    // stays independent of the author's object store (§D4); for a remote URL it changes nothing.
    let commands = vec![
        vec![
            "git".to_string(),
            "clone".into(),
            "--origin".into(),
            "origin".into(),
            "--no-local".into(),
            "--single-branch".into(),
            "--branch".into(),
            req.head_branch.into(),
            "--no-checkout".into(),
            url.clone(),
            dir_s.clone(),
        ],
        vec![
            "git".into(),
            "-C".into(),
            dir_s.clone(),
            "fetch".into(),
            "origin".into(),
            format!("refs/heads/{}:{base_ref}", req.base_branch),
        ],
        vec![
            "git".into(),
            "-C".into(),
            dir_s.clone(),
            "checkout".into(),
            "--detach".into(),
            req.head.into(),
        ],
        vec![
            "git".into(),
            "-C".into(),
            dir_s.clone(),
            "rev-parse".into(),
            "HEAD".into(),
        ],
        vec![
            "git".into(),
            "-C".into(),
            dir_s.clone(),
            "merge-base".into(),
            req.head.into(),
            base_ref.clone(),
        ],
    ];
    Ok(json!({
        "schema_version": 1,
        "pr": req.pr,
        "head": req.head,
        "head_branch": req.head_branch,
        "base_branch": req.base_branch,
        "repo_url": url,
        "clone_dir": dir_s,
        "commands": commands,
        "expect": {
            "command_4_stdout": req.head,
            "command_5": "the merge base to quote as `base` in the verdict line",
        },
        "build": "inside the clone's own target/ (one build folder = one worktree or clone)",
        "evidence_dir": evidence,
        "evidence_reason": evidence_reason,
        "after_verdict": "post the verdict and read it back, copy the logs to evidence_dir, then remove clone_dir (the auditor's own clone)",
        "network": "none here; command 1 and 2 fetch when run",
        "writes": false,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn sha_must_be_exactly_40_lowercase_hex() {
        let ok = "0123456789abcdef0123456789abcdef01234567";
        assert!(is_sha(ok));
        assert!(!is_sha(&ok[..39]));
        assert!(!is_sha(&format!("{ok}8")));
        assert!(!is_sha(&ok.to_uppercase()));
        assert!(!is_sha("g123456789abcdef0123456789abcdef01234567"));
    }

    #[test]
    fn local_paths_are_told_apart_from_remote_urls() {
        for local in [
            "../forge.git",
            "forge.git",
            "/srv/forge.git",
            "./a:b/c",
            "sub/x:y",
        ] {
            assert!(is_local_path(local), "{local}");
        }
        for remote in [
            "https://github.com/o/r.git",
            "ssh://git@github.com/o/r.git",
            "file:///srv/forge.git",
            "git@github.com:o/r.git",
            "host:path",
        ] {
            assert!(!is_local_path(remote), "{remote}");
        }
    }

    #[test]
    fn usable_url_rejects_only_raw_whitespace_empty_and_leading_dash() {
        for bad in [
            "",
            " ",
            "a b",
            "../my forge.git",
            "-x",
            "\t",
            " ../forge.git",
            "../forge.git ",
            "../forge.git\r",
            "../a\nb",
            "../forge\u{1}.git",
            "../forge\0.git",
        ] {
            assert!(!usable_url(bad), "{bad:?}");
        }
        // Written values without whitespace pass. A path origin_url builds, such as
        // "/tmp/audit space/forge.git", is never given to this check.
        for good in [
            "../forge.git",
            "/srv/forge.git",
            "git@github.com:o/r.git",
            "x-",
        ] {
            assert!(usable_url(good), "{good:?}");
        }
    }

    #[test]
    fn evidence_lives_under_the_project_audits_folder() {
        let p = evidence_dir(
            Path::new("/h/.moai-cli"),
            "p1",
            2380,
            "0123456789abcdef0123456789abcdef01234567",
        );
        assert_eq!(
            p,
            Path::new("/h/.moai-cli/projects/p1/audits/2380-01234567")
        );
    }
}
