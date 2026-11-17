//! Gate identity trailers (SDD-6, user decision U-B 2026-11-17; SPEC-20261117-session-trailer):
//! every commit an agent makes carries `Moai-Spec: <SPEC-id>` and `Moai-Session: <session id>`.
//! They replace the planned `Moai-Claim: C-<n>#<gen>`: a card number is personal and never goes
//! into a shared place.
//!
//! The session id is the one `session-current` reads (`crate::session_current`): the id the host
//! puts in the session's environment (`CLAUDE_CODE_SESSION_ID`, then `CODEX_THREAD_ID`), stable
//! for that host session, not a secret and not a card number, accepted only in its id shape
//! (8 to 64 hex digits and `-`). The audit gate reads the same `Moai-Session:` values as the
//! authors (`crate::audit_gate::sessions_of`).

use crate::session_current;
use crate::spec_doc::{self, IdForm};
use serde_json::{json, Value};
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

pub const SPEC_KEY: &str = "Moai-Spec";
pub const SESSION_KEY: &str = "Moai-Session";

/// The trailer lines for a commit of `spec`, with this session's id when `get` (an environment
/// reader) gives one. Only a new-form SPEC id is written.
pub fn format(spec: &str, get: impl Fn(&str) -> Option<String>) -> Result<Value> {
    if spec_doc::id_form(spec) != Some(IdForm::New) {
        return Err(
            "the SPEC id must be SPEC-<yyyymmdd>-<slug>; a card number is never a trailer".into(),
        );
    }
    let found = session_current::resolve(get);
    let mut text = format!("{SPEC_KEY}: {spec}\n");
    let mut out = json!({"spec": spec});
    match found["source_session_id"].as_str() {
        Some(id) => {
            text.push_str(&format!("{SESSION_KEY}: {id}\n"));
            out["session"] = json!(id);
            out["session_source"] = found["source"].clone();
            out["session_missing"] = json!(false);
        }
        None => {
            out["session_missing"] = json!(true);
            out["reason"] = json!(format!(
                "no host variable gives this session's id ({}); the commit carries {SPEC_KEY} only and the audit gate reads its authors as unknown",
                session_current::SOURCES.join(", ")
            ));
        }
    }
    out["text"] = json!(text);
    Ok(out)
}

/// One commit's identity trailers as `git log` reads its trailer block.
#[derive(Debug, PartialEq)]
pub struct CommitTrailers {
    pub sha: String,
    pub specs: Vec<String>,
    pub sessions: Vec<String>,
}

/// One finding of [`check`]: `missing_spec`, `malformed_spec`, `missing_session` or
/// `malformed_session`.
fn findings(c: &CommitTrailers, spec_branch: bool) -> Vec<Value> {
    let mut out = Vec::new();
    let mut add = |kind: &str, fails: bool, detail: String| {
        out.push(json!({"sha": c.sha, "kind": kind, "fails": fails, "detail": detail}));
    };
    match c.specs.as_slice() {
        [] => add(
            "missing_spec",
            spec_branch,
            format!("no {SPEC_KEY} trailer"),
        ),
        [one] if spec_doc::id_form(one).is_some() => {}
        _ => add(
            "malformed_spec",
            true,
            format!("{SPEC_KEY} must be one SPEC id (SPEC-<yyyymmdd>-<slug>)"),
        ),
    }
    match c.sessions.as_slice() {
        [] => add(
            "missing_session",
            false,
            format!("no {SESSION_KEY} trailer; the audit gate reads the authors as unknown"),
        ),
        [one] if session_current::well_formed(one) => {}
        _ => add(
            "malformed_session",
            true,
            format!("{SESSION_KEY} must be one session id (8 to 64 hex digits and `-`)"),
        ),
    }
    out
}

/// Parses `git log --format=%H%x1f<specs>%x1f<sessions>` lines (values separated by `\x1e`).
pub fn parse_log(text: &str) -> Vec<CommitTrailers> {
    let split = |s: &str| -> Vec<String> {
        s.split('\x1e')
            .map(str::trim)
            .filter(|v| !v.is_empty())
            .map(String::from)
            .collect()
    };
    text.lines()
        .filter_map(|l| {
            let mut parts = l.split('\x1f');
            let sha = parts.next()?.trim().to_string();
            (!sha.is_empty()).then(|| CommitTrailers {
                sha,
                specs: split(parts.next().unwrap_or("")),
                sessions: split(parts.next().unwrap_or("")),
            })
        })
        .collect()
}

/// The decision over parsed commits, oldest first: every finding, how many fail, and `ok`.
/// A missing SPEC fails only on a `spec/` branch; a missing session never fails.
pub fn judge(commits: &[CommitTrailers], branch: &str) -> Value {
    let spec_branch = branch.starts_with("spec/");
    let all: Vec<Value> = commits
        .iter()
        .flat_map(|c| findings(c, spec_branch))
        .collect();
    let failing = all.iter().filter(|f| f["fails"] == json!(true)).count();
    json!({"branch": branch, "spec_branch": spec_branch, "commits": commits.len(),
           "findings": all, "failing": failing, "ok": failing == 0})
}

/// Reads the commits of `range` in the repository at `dir` (merges skipped, oldest first) and
/// judges their identity trailers. Reads only.
pub fn check(dir: &Path, range: &str) -> Result<Value> {
    if range.is_empty() || range.starts_with('-') {
        return Err(
            "the range must be a revision range such as <base>..HEAD, not an option".into(),
        );
    }
    let git = |args: &[&str]| -> Result<String> {
        let out = crate::audit_delta::command(dir)
            .args(args)
            .output()
            .map_err(|e| format!("cannot execute git: {e}"))?;
        if !out.status.success() {
            return Err(format!(
                "git {} failed: {}",
                args.first().unwrap_or(&""),
                String::from_utf8_lossy(&out.stderr).trim()
            ));
        }
        Ok(String::from_utf8_lossy(&out.stdout).into_owned())
    };
    let log = git(&[
        "log",
        "--no-merges",
        "--reverse",
        "--format=%H%x1f%(trailers:key=Moai-Spec,valueonly,separator=%x1e)%x1f%(trailers:key=Moai-Session,valueonly,separator=%x1e)",
        range,
        "--",
    ])?;
    let branch = git(&["rev-parse", "--abbrev-ref", "HEAD"])?
        .trim()
        .to_string();
    Ok(judge(&parse_log(&log), &branch))
}

#[cfg(test)]
mod tests {
    use super::*;

    const S: &str = "428e289a-1c67-492e-a23b-a5aca15b4d1f";

    fn env(pairs: &'static [(&'static str, &'static str)]) -> impl Fn(&str) -> Option<String> {
        move |k| {
            pairs
                .iter()
                .find(|(n, _)| *n == k)
                .map(|(_, v)| v.to_string())
        }
    }

    #[test]
    fn format_writes_the_spec_and_the_session_when_there_is_one() {
        let v = format("SPEC-20261117-x", env(&[("CODEX_THREAD_ID", S)])).unwrap();
        assert_eq!(
            v["text"],
            format!("Moai-Spec: SPEC-20261117-x\nMoai-Session: {S}\n")
        );
        let v = format("SPEC-20261117-x", env(&[])).unwrap();
        assert_eq!(v["text"], "Moai-Spec: SPEC-20261117-x\n");
        assert_eq!(v["session_missing"], true);
        assert!(format("SPEC-AUTH-001", env(&[])).is_err());
    }

    #[test]
    fn the_log_reads_several_values_and_empty_fields() {
        let log = format!("a\x1fSPEC-20261117-x\x1f{S}\nb\x1f\x1f\nc\x1fX\x1eY\x1f{S}\x1e{S}\n");
        let c = parse_log(&log);
        assert_eq!(c.len(), 3);
        assert_eq!(c[1].specs, Vec::<String>::new());
        assert_eq!(c[2].specs, ["X", "Y"]);
        let v = judge(&c, "spec/x");
        // b: missing spec (fails on spec/) + missing session; c: two specs, two sessions.
        assert_eq!(v["failing"], 3, "{v}");
        assert_eq!(judge(&c[..1], "spec/x")["ok"], true);
    }
}
