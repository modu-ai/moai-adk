//! Gate identity trailers (SDD-6, user decision U-B 2026-11-17; SPEC-20261117-session-trailer):
//! every commit an agent makes carries `Moai-Spec: <SPEC-id>` and `Moai-Session: <session id>`.
//! They replace the planned `Moai-Claim: C-<n>#<gen>`: a card number is personal and never goes
//! into a shared place.
//!
//! Beside them a commit says how it was verified (SPEC-20261119-verified-trailer, card C-98):
//! `Moai-Verified: <tests|mutants|e2e|smoke|none>[, …]`, `none` an honest value. `trailers
//! --verified` prints the line; `trailer-check` fails an unknown kind anywhere and a missing
//! line only on a descendant of the rule's start commit — a commit of the checked range that
//! carries the line, or `--verified-since <rev>` (that commit itself stays exempt: what is
//! "after the start" needs the line; SPEC-20261120-trailer-verified-ancestry: a commit listed
//! later but not descended from the start, such as one of an old merged side branch, is not
//! after it). Anywhere else a missing line reads as `none`. `trailers --list`
//! groups a range's commits by kind and marks the ones a later standard revert message names, so
//! `none` and reverted commits can be counted together.
//!
//! The session id is the one `session-current` reads (`crate::session_current`): the id the host
//! puts in the session's environment (`CLAUDE_CODE_SESSION_ID`, then `CODEX_THREAD_ID`), not a
//! secret and not a card number, accepted only in its id shape (8 to 64 hex digits and `-`). It
//! is not stable for a whole working stretch: Claude Code gives a new id after `/clear`
//! (its environment-variable reference: "updated on `/clear`"), so one author's commits can
//! carry several ids, and the lane-name comparison of the audit gate stays the fallback. The
//! value is self-declared — an author can write any value — and the gate trusts it the way it
//! trusts lane names. The audit gate reads the same `Moai-Session:` values as the authors
//! (`crate::audit_gate::sessions_of`).

use crate::session_current;
use crate::spec_doc::{self, IdForm};
use serde_json::{json, Value};
use std::collections::{HashMap, HashSet};
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

pub const SPEC_KEY: &str = "Moai-Spec";
pub const SESSION_KEY: &str = "Moai-Session";
pub const VERIFIED_KEY: &str = "Moai-Verified";

/// The accepted `Moai-Verified` kinds (SPEC-20261119-verified-trailer): how a commit was
/// verified, `none` included.
pub const VERIFIED_KINDS: [&str; 5] = ["tests", "mutants", "e2e", "smoke", "none"];

/// The kinds of one commit's `Moai-Verified` values (one line or several, comma-separated), in
/// order, without empties; an unknown kind stays as written for the finding that names it.
pub fn verified_kinds(values: &[String]) -> Vec<String> {
    values
        .iter()
        .flat_map(|v| v.split(','))
        .map(str::trim)
        .filter(|k| !k.is_empty())
        .map(String::from)
        .collect()
}

/// Whether the kinds are one or more of [`VERIFIED_KINDS`]; none at all is not a reading of
/// how a commit was verified.
pub fn kinds_accepted(kinds: &[String]) -> bool {
    !kinds.is_empty() && kinds.iter().all(|k| VERIFIED_KINDS.contains(&k.as_str()))
}

/// Hex digits of the worktree key.
const KEY_LEN: usize = 12;

/// The shape of one `Moai-Session` value (re-audit B3, leader decision 2026-11-17).
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Shape {
    /// `<session id>:<12 lowercase hex>`: the host session and the worktree.
    Full,
    /// A bare session id written before the worktree key: read for one transition.
    Legacy,
    /// Anything else.
    Bad,
}

/// The shape of a `Moai-Session` value.
pub fn value_shape(v: &str) -> Shape {
    match v.split_once(':') {
        Some((id, key))
            if session_current::well_formed(id)
                && key.len() == KEY_LEN
                && key
                    .bytes()
                    .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b)) =>
        {
            Shape::Full
        }
        None if session_current::well_formed(v) => Shape::Legacy,
        _ => Shape::Bad,
    }
}

/// Whether two `Moai-Session` values name the same identity: two full values match exactly; a
/// legacy bare id matches every value of its session, so the transition never makes an author
/// look independent (ASCII case ignored).
pub fn same_identity(a: &str, b: &str) -> bool {
    let (a, b) = (a.to_ascii_lowercase(), b.to_ascii_lowercase());
    match (value_shape(&a), value_shape(&b)) {
        (Shape::Full, Shape::Full) => a == b,
        _ => {
            let session = |v: &str| v.split(':').next().unwrap_or(v).to_string();
            session(&a) == session(&b)
        }
    }
}

/// The worktree key of a git dir: the first 12 hex digits of the sha256 of its path.
pub fn worktree_key_of(git_dir: &Path) -> String {
    use sha2::{Digest, Sha256};
    let hex = format!("{:x}", Sha256::digest(git_dir.to_string_lossy().as_bytes()));
    hex[..KEY_LEN].to_string()
}

/// The worktree key of the worktree that holds `dir`: from `git rev-parse --absolute-git-dir`
/// (scrubbed git environment), made canonical. `None` outside a git worktree.
pub fn worktree_key(dir: &Path) -> Option<String> {
    let out = crate::audit_delta::command(dir)
        .args(["rev-parse", "--absolute-git-dir"])
        .output()
        .ok()?;
    if !out.status.success() {
        return None;
    }
    let raw = String::from_utf8(out.stdout).ok()?;
    let git_dir = std::fs::canonicalize(raw.trim()).ok()?;
    Some(worktree_key_of(&git_dir))
}

/// [`format_verified`] with no `Moai-Verified` line: the identity trailers only.
pub fn format(
    spec: &str,
    get: impl Fn(&str) -> Option<String>,
    key: Option<&str>,
) -> Result<Value> {
    format_verified(spec, &[], get, key)
}

/// The trailer lines for a commit of `spec`, with `Moai-Session: <session id>:<worktree key>`
/// when `get` (an environment reader) gives this session's id and `key` names the worktree
/// ([`worktree_key`]), and one `Moai-Verified` line when `verified` names kinds (each must be
/// one of [`VERIFIED_KINDS`]). Only a new-form SPEC id is written.
pub fn format_verified(
    spec: &str,
    verified: &[String],
    get: impl Fn(&str) -> Option<String>,
    key: Option<&str>,
) -> Result<Value> {
    if spec_doc::id_form(spec) != Some(IdForm::New) {
        return Err(
            "the SPEC id must be SPEC-<yyyymmdd>-<slug>; a card number is never a trailer".into(),
        );
    }
    let kinds = verified_kinds(verified);
    if !verified.is_empty() && !kinds_accepted(&kinds) {
        return Err(format!(
            "the verified kinds must be one or more of {} (comma-separated)",
            VERIFIED_KINDS.join(", ")
        ));
    }
    let found = session_current::resolve(get);
    let mut text = format!("{SPEC_KEY}: {spec}\n");
    let mut out = json!({"spec": spec});
    match (found["source_session_id"].as_str(), key) {
        (Some(id), Some(key)) => {
            let value = format!("{id}:{key}");
            text.push_str(&format!("{SESSION_KEY}: {value}\n"));
            out["session"] = json!(value);
            out["session_source"] = found["source"].clone();
            out["session_missing"] = json!(false);
        }
        (Some(_), None) => {
            out["session_missing"] = json!(true);
            out["reason"] = json!(format!(
                "this folder is not in a git worktree, so no worktree key can be read; the commit carries {SPEC_KEY} only"
            ));
        }
        (None, _) => {
            out["session_missing"] = json!(true);
            out["reason"] = json!(format!(
                "no host variable gives this session's id ({}); the commit carries {SPEC_KEY} only and the audit gate reads its authors as unknown",
                session_current::SOURCES.join(", ")
            ));
        }
    }
    if !kinds.is_empty() {
        text.push_str(&format!("{VERIFIED_KEY}: {}\n", kinds.join(", ")));
        out["verified"] = json!(kinds);
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
    /// The `Moai-Verified` values of the commit, as written (one per line).
    pub verified: Vec<String>,
    /// The commit's subject and body, read for the card-number scan (the trailers are not part
    /// of either).
    pub subject: String,
    pub body: String,
    /// The log line had more fields than it should: a trailer value held the log's own field
    /// separator (audit N3), so its values cannot be read apart.
    pub garbled: bool,
}

/// One finding of [`check`]: `missing_spec`, `malformed_spec`, `missing_session`,
/// `malformed_session`, `missing_verified`, `malformed_verified`, or `malformed_trailers` for a
/// value holding the log's separator. `under_rule` says the commit sits after the `Moai-Verified`
/// rule's start, where a missing line is no longer read as `none`.
fn findings(c: &CommitTrailers, spec_branch: bool, under_rule: bool) -> Vec<Value> {
    let mut out = Vec::new();
    let mut add = |kind: &str, fails: bool, detail: String| {
        out.push(json!({"sha": c.sha, "kind": kind, "fails": fails, "detail": detail}));
    };
    if c.garbled {
        add(
            "malformed_trailers",
            true,
            "the commit record cannot be read apart (a raw separator byte in a trailer or the \
             body split it); its values are not trusted"
                .into(),
        );
        return out;
    }
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
        [one] => match value_shape(one) {
            Shape::Full => {}
            Shape::Legacy => add(
                "legacy_session",
                false,
                format!(
                    "{SESSION_KEY} is a bare session id from before the worktree key; read for the transition (it counts its whole session as one author)"
                ),
            ),
            Shape::Bad => add(
                "malformed_session",
                true,
                format!(
                    "{SESSION_KEY} must be <session id>:<12 hex digits> (8 to 64 hex digits and `-`, then the worktree key)"
                ),
            ),
        },
        _ => add(
            "malformed_session",
            true,
            format!("{SESSION_KEY} must appear once"),
        ),
    }
    let kinds = verified_kinds(&c.verified);
    if kinds.is_empty() {
        if under_rule {
            add(
                "missing_verified",
                true,
                format!(
                    "no {VERIFIED_KEY} trailer after the rule's start commit; `none` is allowed — say it"
                ),
            );
        }
    } else if !kinds_accepted(&kinds) {
        add(
            "malformed_verified",
            true,
            format!(
                "{VERIFIED_KEY} kinds must be one or more of {}",
                VERIFIED_KINDS.join(", ")
            ),
        );
    }
    out
}

/// Parses `git log --format=%x1e%H%x1f<specs>%x1f<sessions>%x1f<verified>%x1f<subject>%x1f<body>`
/// records (records separated by `\x1e`, trailer values by `\x03`).
pub fn parse_log(text: &str) -> Vec<CommitTrailers> {
    let split = |s: &str| -> Vec<String> {
        s.split('\x03')
            .map(str::trim)
            .filter(|v| !v.is_empty())
            .map(String::from)
            .collect()
    };
    text.split('\x1e')
        .filter_map(|record| {
            // The body is multiline, so a whole commit is one `\x1e` record and the fields
            // inside it split on `\x1f`; a trailer value holding the field separator still
            // shows up as too many fields, and a raw `\x1e` inside a message splits the
            // record, which shows up as a sha that is not a commit id — both garbled, both
            // loud, never silently scanned past.
            let parts: Vec<&str> = record.split('\x1f').collect();
            let sha = parts.first()?.trim().to_string();
            let real_sha = !sha.is_empty() && sha.bytes().all(|b| b.is_ascii_hexdigit());
            (!sha.is_empty()).then(|| CommitTrailers {
                sha,
                specs: split(parts.get(1).copied().unwrap_or("")),
                sessions: split(parts.get(2).copied().unwrap_or("")),
                verified: split(parts.get(3).copied().unwrap_or("")),
                subject: parts.get(4).copied().unwrap_or("").trim().to_string(),
                body: parts.get(5).copied().unwrap_or("").trim().to_string(),
                garbled: parts.len() > 6 || !real_sha,
            })
        })
        .collect()
}

/// The decision over parsed commits, oldest first: every finding, how many fail, and `ok`.
/// A missing SPEC fails only on a `spec/` branch; a missing session never fails; a missing
/// `Moai-Verified` fails only from `verified_from` on (index into `commits`, the first commit
/// the rule covers; `None` covers none of them, the default when the range names no start).
pub fn judge_from(commits: &[CommitTrailers], branch: &str, verified_from: Option<usize>) -> Value {
    let spec_branch = branch.starts_with("spec/");
    let from = verified_from.unwrap_or(usize::MAX);
    let all: Vec<Value> = commits
        .iter()
        .enumerate()
        .flat_map(|(i, c)| findings(c, spec_branch, i >= from))
        .collect();
    let failing = all.iter().filter(|f| f["fails"] == json!(true)).count();
    json!({"branch": branch, "spec_branch": spec_branch, "commits": commits.len(),
           "verified_from": verified_from, "findings": all, "failing": failing,
           "ok": failing == 0})
}

/// [`judge_from`] with no `Moai-Verified` rule: only a present, malformed line fails.
pub fn judge(commits: &[CommitTrailers], branch: &str) -> Value {
    judge_from(commits, branch, None)
}

/// The log fields of a range's commits, oldest first, merges skipped.
fn log_trailers(git: &impl Fn(&[&str]) -> Result<String>, range: &str) -> Result<String> {
    git(&[
        "log",
        "--no-merges",
        "--reverse",
        "--format=%x1e%H%x1f%(trailers:key=Moai-Spec,valueonly,separator=%x03)%x1f%(trailers:key=Moai-Session,valueonly,separator=%x03)%x1f%(trailers:key=Moai-Verified,valueonly,separator=%x03)%x1f%s%x1f%b",
        range,
        "--",
    ])
}

/// The index reported as `verified_from` in the oldest-first list of `commits`: the position of
/// the rule's start. `--verified-since` resolves to a commit — the index just after it, or 0 when
/// that commit is outside the list (a start outside the range, or a merge commit, which the list
/// leaves out); without it the first commit carrying the line is the start. The index only
/// reports the start: which commits the rule requires the line on is decided by ancestry in
/// `verified_cover`, not by this index.
fn verified_from(commits: &[CommitTrailers], since: Option<&str>, sha: &str) -> Option<usize> {
    if since.is_some() {
        return Some(match commits.iter().position(|c| c.sha == sha) {
            Some(at) => at + 1,
            None => 0,
        });
    }
    commits.iter().position(|c| !c.verified.is_empty())
}

/// Reads the commits of `range` in the repository at `dir` (merges skipped, oldest first) and
/// judges their identity trailers. Reads only.
pub fn check(
    dir: &Path,
    range: &str,
    branch: Option<&str>,
    verified_since: Option<&str>,
) -> Result<Value> {
    check_rules(dir, range, branch, verified_since, None)
}

/// [`check`] with the card-number rule: `card_numbers_since` records the rule's start (that
/// commit itself exempt; a start outside the range covers the whole range), `None` keeps the
/// rule off.
pub fn check_rules(
    dir: &Path,
    range: &str,
    branch: Option<&str>,
    verified_since: Option<&str>,
    card_numbers_since: Option<&str>,
) -> Result<Value> {
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
    let log = log_trailers(&git, range)?;
    let since_sha = match verified_since {
        None => None,
        Some(s) if s.is_empty() || s.starts_with('-') => {
            return Err("the verified start must be a revision, not an option".into())
        }
        Some(s) => Some(
            git(&["rev-parse", "--verify", &format!("{s}^{{commit}}")])?
                .trim()
                .to_string(),
        ),
    };
    let cards_sha = match card_numbers_since {
        None => None,
        Some(s) if s.is_empty() || s.starts_with('-') => {
            return Err("the card-number start must be a revision, not an option".into())
        }
        Some(s) => Some(
            git(&["rev-parse", "--verify", &format!("{s}^{{commit}}")])?
                .trim()
                .to_string(),
        ),
    };
    // Re-audit N4: a detached checkout (CI) names the branch it judges.
    let branch = match branch {
        Some(b) => b.to_string(),
        None => git(&["rev-parse", "--abbrev-ref", "HEAD"])?
            .trim()
            .to_string(),
    };
    let commits = parse_log(&log);
    let from = verified_from(&commits, verified_since, since_sha.as_deref().unwrap_or(""));
    let covered = verified_cover(&git, range, &commits, since_sha.as_deref(), from)?;
    let cards_from = match cards_sha.as_deref() {
        Some(sha) => card_numbers_from(&commits, sha),
        None => None,
    };
    Ok(judge_covered(&commits, &branch, from, &covered, cards_from))
}

/// Which of `commits` the `Moai-Verified` rule covers: those that descend from the rule's start,
/// not those listed after it (the listing mixes the commits of merged side branches in by date).
/// The start is every commit of the range that carries the line, or the one commit
/// `--verified-since` names (`since_sha`); a commit is covered when one of its ancestors inside the
/// range is a start or is itself covered, and merge commits pass the mark on without being
/// judged. A `--verified-since` commit outside the range covers the whole range. A commit the
/// ancestry listing does not show falls back to the listing index `from`.
fn verified_cover(
    git: &impl Fn(&[&str]) -> Result<String>,
    range: &str,
    commits: &[CommitTrailers],
    since_sha: Option<&str>,
    from: Option<usize>,
) -> Result<Vec<bool>> {
    let listing = git(&[
        "rev-list",
        "--parents",
        "--reverse",
        "--topo-order",
        range,
        "--",
    ])?;
    let mut order: Vec<&str> = Vec::new();
    let mut parents: HashMap<&str, Vec<&str>> = HashMap::new();
    for line in listing.lines() {
        let mut shas = line.split_whitespace();
        if let Some(sha) = shas.next() {
            order.push(sha);
            parents.insert(sha, shas.collect());
        }
    }
    if since_sha.is_some_and(|s| !parents.contains_key(s)) {
        return Ok(vec![true; commits.len()]);
    }
    let carrying: HashSet<&str> = commits
        .iter()
        .filter(|c| !c.verified.is_empty())
        .map(|c| c.sha.as_str())
        .collect();
    let is_start = |sha: &str| match since_sha {
        Some(s) => sha == s,
        None => carrying.contains(sha),
    };
    let mut under: HashSet<&str> = HashSet::new();
    for sha in &order {
        let below_start = parents[sha]
            .iter()
            .any(|p| parents.contains_key(p) && (is_start(p) || under.contains(p)));
        if below_start {
            under.insert(sha);
        }
    }
    Ok(commits
        .iter()
        .enumerate()
        .map(|(i, c)| {
            if parents.contains_key(c.sha.as_str()) {
                under.contains(c.sha.as_str())
            } else {
                from.is_some_and(|f| i >= f)
            }
        })
        .collect())
}

/// Judges the commits with both rules: the `Moai-Verified` start index and the card-number
/// start index (`None` keeps that rule off).
pub fn judge_rules(
    commits: &[CommitTrailers],
    branch: &str,
    verified_from: Option<usize>,
    card_numbers_from: Option<usize>,
) -> Value {
    let from = verified_from.unwrap_or(usize::MAX);
    let covered: Vec<bool> = (0..commits.len()).map(|i| i >= from).collect();
    judge_covered(commits, branch, verified_from, &covered, card_numbers_from)
}

/// [`judge_rules`] with the `Moai-Verified` coverage given per commit (`covered[i]` for
/// `commits[i]`, see [`verified_cover`]); `verified_from` is only reported.
pub fn judge_covered(
    commits: &[CommitTrailers],
    branch: &str,
    verified_from: Option<usize>,
    covered: &[bool],
    card_numbers_from: Option<usize>,
) -> Value {
    let spec_branch = branch.starts_with("spec/");
    let card_at = card_numbers_from.unwrap_or(usize::MAX);
    let mut all: Vec<Value> = Vec::new();
    for (i, c) in commits.iter().enumerate() {
        all.extend(findings(
            c,
            spec_branch,
            covered.get(i).copied().unwrap_or(false),
        ));
        if i >= card_at {
            all.extend(card_number_findings(c));
        }
    }
    let failing = all.iter().filter(|f| f["fails"] == json!(true)).count();
    json!({"branch": branch, "spec_branch": spec_branch, "commits": commits.len(),
           "verified_from": verified_from, "card_numbers_from": card_numbers_from,
           "findings": all, "failing": failing, "ok": failing == 0})
}

/// Where the card-number rule starts among `commits`: the index after the recorded start (itself
/// exempt), the whole range when the start sits outside it, none without a recorded start.
fn card_numbers_from(commits: &[CommitTrailers], sha: &str) -> Option<usize> {
    commits
        .iter()
        .position(|c| c.sha == sha)
        .map(|at| at + 1)
        .or(Some(0))
}

/// The card-number shapes of one line: `C-<n>` or `c-<n>` with a boundary before (no leading
/// alphanumeric or hyphen, so `AC-150` and `SPEC-C-9` never match) and after (no trailing
/// alphanumeric, so `C-150b` never matches). A match with an odd number of double quotes before
/// it on the line is quoted ledger output and exempt.
fn card_numbers_in_line(line: &str) -> Vec<String> {
    let chars: Vec<char> = line.chars().collect();
    let mut out = Vec::new();
    for (i, c) in chars.iter().enumerate() {
        if *c != 'C' && *c != 'c' || chars.get(i + 1) != Some(&'-') {
            continue;
        }
        let mut digits = 0;
        while chars
            .get(i + 2 + digits)
            .is_some_and(|d| d.is_ascii_digit())
        {
            digits += 1;
        }
        if digits == 0
            || chars
                .get(i + 2 + digits)
                .is_some_and(|d| d.is_ascii_alphanumeric())
        {
            continue;
        }
        if i > 0 && (chars[i - 1].is_ascii_alphanumeric() || chars[i - 1] == '-') {
            continue;
        }
        if chars[..i].iter().filter(|q| **q == '"').count() % 2 == 1 {
            continue;
        }
        out.push(chars[i..i + 2 + digits].iter().collect());
    }
    out
}

/// The `card_number` findings of one commit: every shape in its subject or body, with the
/// message line it sits on. The number itself is named — it is not a secret.
fn card_number_findings(c: &CommitTrailers) -> Vec<Value> {
    let mut out = Vec::new();
    let lines = std::iter::once((1usize, c.subject.as_str()))
        .chain(c.body.lines().enumerate().map(|(n, l)| (n + 3, l)));
    for (no, line) in lines {
        for shape in card_numbers_in_line(line) {
            out.push(json!({
                "sha": c.sha, "kind": "card_number", "fails": true,
                "detail": format!(
                    "the commit message names card {shape} on line {no}; card numbers live in \
                     the ledger and the decision record, not the message"
                ),
            }));
        }
    }
    out
}

/// Judges one draft commit message (the commit-msg hook's input) with the card-number scan
/// always on. Reads the file only.
pub fn message_file_check(path: &Path) -> Result<Value> {
    let bytes = std::fs::read(path)
        .map_err(|e| format!("cannot read the message file {}: {e}", path.display()))?;
    let text = String::from_utf8_lossy(&bytes);
    let mut findings = Vec::new();
    for (n, line) in text.lines().enumerate() {
        for shape in card_numbers_in_line(line) {
            findings.push(json!({
                "kind": "card_number", "fails": true,
                "detail": format!(
                    "the commit message names card {shape} on line {}; card numbers live in \
                     the ledger and the decision record, not the message",
                    n + 1
                ),
            }));
        }
    }
    let failing = findings.len();
    Ok(json!({"message_file": path.display().to_string(),
              "findings": findings, "failing": failing, "ok": failing == 0}))
}

/// The commits a standard revert message (`This reverts commit <sha>`, the text `git revert`
/// writes) names among `commits`; a short sha counts when it is a prefix of a listed one.
fn reverted_of(commits: &[CommitTrailers], bodies: &str) -> Vec<String> {
    let mut reverted = Vec::new();
    for record in bodies.split('\x1e') {
        for (at, _) in record.match_indices("This reverts commit ") {
            let token: String = record[at + "This reverts commit ".len()..]
                .chars()
                .take_while(|c| c.is_ascii_hexdigit())
                .collect();
            if (7..=40).contains(&token.len())
                && commits
                    .iter()
                    .any(|c| c.sha == token || c.sha.starts_with(&token))
                && !reverted.iter().any(|r| r == &token)
            {
                reverted.push(token);
            }
        }
    }
    reverted
}

/// Lists the commits of `range` (merges skipped, oldest first) by verification kind: each commit
/// with its kinds (`none` for a missing line), the commits grouped per kind (a commit with
/// several kinds appears in each), the reverted ones and per-kind counts. Reads only.
pub fn list(dir: &Path, range: &str) -> Result<Value> {
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
    let commits = parse_log(&log_trailers(&git, range)?);
    let bodies = git(&[
        "log",
        "--no-merges",
        "--reverse",
        "--format=%H%x1e%B%x1e",
        "-F",
        "--grep=This reverts commit",
        range,
        "--",
    ])?;
    let reverted = reverted_of(&commits, &bodies);
    let is_reverted = |sha: &str| {
        reverted.iter().any(|r| {
            sha == r || sha.starts_with(r) || r.starts_with(&sha[..r.len().min(sha.len())])
        })
    };
    let mut by_kind = serde_json::Map::new();
    let mut counts = serde_json::Map::new();
    for k in VERIFIED_KINDS {
        by_kind.insert(k.into(), json!([]));
        counts.insert(k.into(), json!(0));
    }
    let mut listed = Vec::new();
    for c in &commits {
        let kinds = verified_kinds(&c.verified);
        let kinds: Vec<String> = if kinds.is_empty() {
            vec!["none".into()]
        } else {
            kinds
        };
        let rev = is_reverted(&c.sha);
        for k in &kinds {
            if let Some(a) = by_kind.get_mut(k).and_then(|v| v.as_array_mut()) {
                a.push(json!(c.sha));
            }
            if let Some(n) = counts.get_mut(k).and_then(|v| v.as_i64()) {
                counts.insert(k.clone(), json!(n + 1));
            }
        }
        listed.push(json!({"sha": c.sha, "kinds": kinds, "reverted": rev}));
    }
    Ok(json!({
        "dir": dir.display().to_string(),
        "range": range,
        "commits": listed,
        "by_kind": Value::Object(by_kind),
        "counts": Value::Object(counts),
        "reverted": reverted,
        "writes": false,
    }))
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

    /// Re-audit B3 (leader decision): a Claude Code subagent inherits its parent's session id,
    /// so the value names the worktree too: `<session id>:<12 hex of sha256(git dir)>`.
    #[test]
    fn the_value_joins_the_session_and_the_worktree_key() {
        let key = worktree_key_of(Path::new("/w/a/.git"));
        assert_eq!(key.len(), 12);
        assert!(key
            .bytes()
            .all(|b| b.is_ascii_hexdigit() && !b.is_ascii_uppercase()));
        assert_ne!(key, worktree_key_of(Path::new("/w/b/.git/worktrees/b")));
        assert_eq!(key, worktree_key_of(Path::new("/w/a/.git")));
        let v = format!("{S}:{key}");
        assert_eq!(value_shape(&v), Shape::Full);
        assert_eq!(value_shape(S), Shape::Legacy);
        for bad in [
            format!("{S}:{}", &key[..11]),
            format!("{S}:{}x", &key[..11]),
            format!("{S}:{key}:x"),
            format!("{S}:{}", key.to_uppercase()),
            format!(":{key}"),
            "agent-7".to_string(),
        ] {
            assert_eq!(value_shape(&bad), Shape::Bad, "{bad}");
        }
    }

    /// Full values match exactly; a legacy bare id (before the worktree key) matches every value
    /// of that session, so the transition never makes an author look independent.
    #[test]
    fn identities_match_by_full_value_or_conservatively_by_session() {
        let a = format!("{S}:aaaaaaaaaaaa");
        let b = format!("{S}:bbbbbbbbbbbb");
        assert!(same_identity(&a, &a));
        assert!(!same_identity(&a, &b));
        assert!(same_identity(S, &b) && same_identity(&a, S));
        assert!(!same_identity("11111111-2222", &a));
        assert!(same_identity(&a.to_uppercase(), &a));
    }

    #[test]
    fn format_writes_the_spec_and_the_session_when_there_is_one() {
        const K: &str = "0123456789ab";
        let v = format("SPEC-20261117-x", env(&[("CODEX_THREAD_ID", S)]), Some(K)).unwrap();
        assert_eq!(
            v["text"],
            format!("Moai-Spec: SPEC-20261117-x\nMoai-Session: {S}:{K}\n")
        );
        let v = format("SPEC-20261117-x", env(&[]), Some(K)).unwrap();
        assert_eq!(v["text"], "Moai-Spec: SPEC-20261117-x\n");
        assert_eq!(v["session_missing"], true);
        // Outside a git worktree there is no key: the session line is left out too.
        let v = format("SPEC-20261117-x", env(&[("CODEX_THREAD_ID", S)]), None).unwrap();
        assert_eq!(v["text"], "Moai-Spec: SPEC-20261117-x\n");
        assert_eq!(v["session_missing"], true);
        assert!(format("SPEC-AUTH-001", env(&[]), Some(K)).is_err());
    }

    /// SPEC-20261119-verified-trailer: the kinds line is printed after the session, several
    /// kinds keep their order, `none` is accepted and an unknown kind is refused with no text.
    #[test]
    fn format_appends_the_verified_line_or_refuses_the_unknown_kind() {
        const K: &str = "0123456789ab";
        let kinds = ["tests".to_string(), "mutants".to_string()];
        let v = format_verified(
            "SPEC-20261119-x",
            &kinds,
            env(&[("CODEX_THREAD_ID", S)]),
            Some(K),
        )
        .unwrap();
        assert_eq!(
            v["text"],
            format!("Moai-Spec: SPEC-20261119-x\nMoai-Session: {S}:{K}\nMoai-Verified: tests, mutants\n")
        );
        assert_eq!(v["verified"], json!(["tests", "mutants"]));
        let v =
            format_verified("SPEC-20261119-x", &["none".to_string()], env(&[]), Some(K)).unwrap();
        assert!(v["text"]
            .as_str()
            .unwrap()
            .ends_with("Moai-Verified: none\n"));
        let e = format_verified(
            "SPEC-20261119-x",
            &["tests".to_string(), "tset".to_string()],
            env(&[]),
            Some(K),
        )
        .unwrap_err();
        assert!(e.contains("tests, mutants, e2e, smoke, none"), "{e}");
    }

    /// SPEC-20261119-verified-trailer: kinds parse from one line or several, commas and spaces,
    /// and only the five accepted values validate (`none` included).
    #[test]
    fn verified_kinds_parse_and_validate() {
        let one = ["tests, mutants".to_string()];
        assert_eq!(verified_kinds(&one), ["tests", "mutants"]);
        assert_eq!(
            verified_kinds(&["e2e".into(), " smoke ".into(), "".into()]),
            ["e2e", "smoke"]
        );
        assert!(kinds_accepted(&verified_kinds(&one)));
        assert!(kinds_accepted(&["none".into()]));
        assert!(!kinds_accepted(&["tset".into()]));
        assert!(!kinds_accepted(&[]));
        assert!(!kinds_accepted(&["tests".into(), "mutant".into()]));
    }

    /// The rule starts at the range's first carrying commit (or `verified_from`): before it a
    /// missing line reads as `none`, from it on the line is required, and a malformed value
    /// fails anywhere.
    #[test]
    fn judge_cuts_at_the_first_commit_carrying_the_line() {
        let log = "\x1ea\x1f\x1f\x1f\x1f\x1f\x1eb\x1f\x1f\x1ftests\x1f\x1f\x1ec\x1f\x1f\x1f\x1f\x1f\x1ed\x1f\x1f\x1ftset\x1f\x1f";
        let c = parse_log(log);
        assert_eq!(c.len(), 4);
        assert_eq!(c[1].verified, ["tests"]);
        // No start named: only the malformed value of d fails.
        let v = judge(&c, "main");
        assert_eq!(v["failing"], 1, "{v}");
        let failing: Vec<&str> = v["findings"]
            .as_array()
            .unwrap()
            .iter()
            .filter(|f| f["fails"] == json!(true))
            .map(|f| f["kind"].as_str().unwrap())
            .collect();
        assert_eq!(failing, ["malformed_verified"], "{v}");
        // The range's first carrying commit (b) starts the rule: c now misses the line.
        let v = judge_from(&c, "main", Some(2));
        assert_eq!(v["failing"], 2, "{v}");
        let kinds: Vec<&str> = v["findings"]
            .as_array()
            .unwrap()
            .iter()
            .map(|f| f["kind"].as_str().unwrap())
            .collect();
        assert!(kinds.contains(&"missing_verified"), "{v}");
        // With the rule from the very first commit, a and c miss the line.
        let v = judge_from(&c, "main", Some(0));
        assert_eq!(v["failing"], 3, "{v}");
    }

    /// Audit N3: a value that holds the log's own separators (or any control character)
    /// cannot be told apart from two values; the commit is malformed and fails, never misread.
    #[test]
    fn a_value_with_a_separator_or_control_character_is_malformed() {
        let garbled = format!("a\x1fSPEC-20261117-x\x1fextra\x1f{S}\x1f{S}\x1f\x1f");
        let c = parse_log(&garbled);
        assert_eq!(c.len(), 1);
        let v = judge(&c, "main");
        assert_eq!(v["failing"], 1, "{v}");
        assert_eq!(v["findings"][0]["kind"], "malformed_trailers", "{v}");
        let control = format!("b\x1fSPEC-20261117-x\x07\x1f{S}\x1f\x1f\x1f");
        let v = judge(&parse_log(&control), "main");
        assert_eq!(v["findings"][0]["kind"], "malformed_spec", "{v}");
        assert_eq!(
            judge(
                &parse_log(&format!("c\x1fSPEC-20261117-x\x1f{S}\x1f\x1f\x1f")),
                "main"
            )["ok"],
            true
        );
        // A raw record separator inside a message splits the record: the part before it stays
        // readable, the tail reads as a pseudo-commit whose sha is not a commit id and goes
        // garbled — the commit fails loudly, it is never scanned past silently.
        let split = format!(
            "\x1e{:040}\x1fSPEC-20261117-x\x1f{S}\x1ftests\x1fsubject\x1fbody\x1eraw byte",
            "a"
        );
        let c = parse_log(&split);
        assert_eq!(c.len(), 2);
        assert!(!c[0].garbled, "{c:?}");
        assert!(c[1].garbled, "{c:?}");
        let v = judge(&c, "main");
        assert_eq!(v["failing"], 1, "{v}");
        let failing: Vec<&str> = v["findings"]
            .as_array()
            .unwrap()
            .iter()
            .filter(|f| f["fails"] == json!(true))
            .map(|f| f["kind"].as_str().unwrap())
            .collect();
        assert_eq!(failing, ["malformed_trailers"], "{v}");
    }

    #[test]
    fn the_log_reads_several_values_and_empty_fields() {
        let log = format!(
            "\x1ea\x1fSPEC-20261117-x\x1f{S}\x1f\x1f\x1f\x1eb\x1f\x1f\x1f\x1f\x1f\x1ec\x1fX\x03Y\x1f{S}\x03{S}\x1f\x1f\x1f"
        );
        let c = parse_log(&log);
        assert_eq!(c.len(), 3);
        assert_eq!(c[1].specs, Vec::<String>::new());
        assert_eq!(c[2].specs, ["X", "Y"]);
        let v = judge(&c, "spec/x");
        // b: missing spec (fails on spec/) + missing session; c: two specs, two sessions.
        assert_eq!(v["failing"], 3, "{v}");
        // a: a bare id from before the worktree key is reported for the transition, not failed.
        let a = judge(&c[..1], "spec/x");
        assert_eq!(a["ok"], true);
        assert_eq!(a["findings"][0]["kind"], "legacy_session", "{a}");
        assert_eq!(a["findings"][0]["fails"], false, "{a}");
    }
}
