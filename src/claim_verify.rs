//! Claim ownership re-verification (rc.2 A9 #2255): the write boundary a lane crosses before
//! every commit and push.
//!
//! The claim contract is unchanged (AGENTS.md, 06-A): one `moai-claim:` comment per claimant,
//! the lowest comment id among open claims holds the Issue, a lease is 4 h, and an expired
//! lease never hands the Issue to someone else — a handoff comment or the operator does. The
//! comment is not an atomic lock; this module only re-reads every comment and decides, so a
//! competing claim that became visible late is judged on the re-read.
//!
//! What it adds:
//! - [`classify_comment`]: one parser for the claim line (the RC1 form plus the appended
//!   fields `render_claim_line` writes) and for two machine lines that close a claim,
//!   `moai-claim-release: comment=<id>` and `moai-claim-handoff: comment=<id> to=<session>`.
//! - [`evaluate`]: a pure verdict over all comments — `owner`, `not_owner`, `stale`, `closed`,
//!   `no_claim` or `uncertain` — with every ignored comment reported (`malformed`,
//!   `duplicate`, `author_mismatch`; REQ-A9-02).
//! - `claim-verify`: GET-only; succeeds only for `owner`, otherwise refuses with a classified
//!   message (REQ-A9-01). `--hook` reads a host `PreToolUse` input and guards only a Bash
//!   `git commit` / `git push`: in a worktree with an active claim record it verifies that
//!   record; without one, on a lane branch `issue/<n>-…` it verifies the `Moai-Claim: <n>#<id>`
//!   token of the commit message (commit) or the highest-id token on the branch's
//!   first-parent line, merges included (push), and refuses as `uncertain` when no usable
//!   token or GitHub origin is found.
//!
//! Limits: every lane posts through one shared GitHub account, so "author mismatch" means an
//! author outside `allowed_authors`; a forged session value from an allowed author cannot be
//! told apart from a real one.
//!
//! The `--hook` detector reads the command line as text, so it is a guardrail, not a complete
//! boundary (#2306). It sees past wrapper words, quoting, backslash escapes, one level of
//! `sh -c` / `bash -c`, and `git -c alias.<name>=commit|push` on the same line. It does not see
//! a git alias defined in config, a shell alias or function, `$VAR` / `$(...)` expansion,
//! `eval`, a script file, a second shell level, or another program that runs git itself. The
//! ANSI-C quoting `$'git' push` likewise stays invisible — it is `$` expansion to this text
//! detector (docs/hooks-generation.md §3, named limit; rc.2 A9 follow-up 2 #2323).

use crate::claim::{parse_claim_line, ClaimRecord};
use crate::forge::policy::{next_link, parse_utc, Clock};
use crate::forge::transport::Transport;
use crate::forge::Client;
use crate::Result;
use serde::Deserialize;
use serde_json::{json, Value};
use std::path::{Path, PathBuf};

const MAX_PAGES: u32 = 50;
const RELEASE_PREFIX: &str = "moai-claim-release:";
const HANDOFF_PREFIX: &str = "moai-claim-handoff:";

/// One Issue comment as the evaluator sees it.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct IssueComment {
    pub id: i64,
    pub author: String,
    pub body: String,
}

/// What one comment says about claims.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum CommentKind {
    Claim {
        session: String,
        epoch: u64,
        lease_until: String,
    },
    Release {
        target: i64,
    },
    Handoff {
        target: i64,
        to: String,
    },
    /// A line that starts like a claim line but cannot be read; the reason says why.
    Malformed(&'static str),
    /// No claim line at all.
    Other,
}

fn token<'a>(rest: &'a str, key: &str) -> Option<&'a str> {
    rest.split_whitespace()
        .find_map(|t| t.strip_prefix(key).and_then(|v| v.strip_prefix('=')))
}

/// Classifies a comment by its first claim-family line.
pub fn classify_comment(body: &str) -> CommentKind {
    for line in body.lines().map(str::trim) {
        if line.starts_with("moai-claim:") {
            return match parse_claim_line(line) {
                Some(c) if !c.session.is_empty() && parse_utc(&c.lease_until).is_some() => {
                    CommentKind::Claim {
                        session: c.session,
                        epoch: c.epoch,
                        lease_until: c.lease_until,
                    }
                }
                Some(_) => CommentKind::Malformed("claim lease_until is not YYYY-MM-DDTHH:MM:SSZ"),
                None => CommentKind::Malformed("claim line lacks session, epoch or lease_until"),
            };
        }
        if let Some(rest) = line.strip_prefix(RELEASE_PREFIX) {
            return match token(rest, "comment").and_then(|v| v.parse().ok()) {
                Some(target) => CommentKind::Release { target },
                None => CommentKind::Malformed("release line lacks comment=<id>"),
            };
        }
        if let Some(rest) = line.strip_prefix(HANDOFF_PREFIX) {
            let target = token(rest, "comment").and_then(|v| v.parse().ok());
            let to = token(rest, "to").filter(|v| !v.is_empty());
            return match (target, to) {
                (Some(target), Some(to)) => CommentKind::Handoff {
                    target,
                    to: to.to_string(),
                },
                _ => CommentKind::Malformed("handoff line lacks comment=<id> or to=<session>"),
            };
        }
    }
    CommentKind::Other
}

/// The release line a leader or operator posts to close a claim (no POST here).
pub fn render_release_line(claim_comment_id: i64, by: &str) -> String {
    format!("{RELEASE_PREFIX} comment={claim_comment_id} by={by}")
}

/// The handoff line that closes a claim and names the session expected to claim next.
pub fn render_handoff_line(claim_comment_id: i64, to: &str, by: &str) -> String {
    format!("{HANDOFF_PREFIX} comment={claim_comment_id} to={to} by={by}")
}

/// Who is verifying: the session, and the claim comment id when the journal knows it.
#[derive(Clone, Debug, Default)]
pub struct Verifier<'a> {
    pub session: &'a str,
    pub comment_id: Option<i64>,
    /// Logins allowed to post claim lines; empty means not checked.
    pub allowed_authors: &'a [String],
    pub now: i64,
}

#[derive(Clone, Debug)]
struct ValidClaim {
    id: i64,
    session: String,
    epoch: u64,
    live: bool,
}

/// The verdict over all comments of one Issue. `status` is one of `owner`, `not_owner`,
/// `stale`, `closed`, `no_claim`, `uncertain`; only `owner` may write.
pub fn evaluate(comments: &[IssueComment], who: &Verifier<'_>) -> Value {
    let mut sorted: Vec<&IssueComment> = comments.iter().collect();
    sorted.sort_by_key(|c| c.id);
    let mut claims: Vec<ValidClaim> = Vec::new();
    let mut invalid = Vec::new();
    let mut closed = Vec::new();
    let mut seen_claim_lines = 0;
    for c in sorted {
        let kind = classify_comment(&c.body);
        if kind == CommentKind::Other {
            continue;
        }
        seen_claim_lines += 1;
        if !who.allowed_authors.is_empty() && !who.allowed_authors.contains(&c.author) {
            invalid.push(json!({"comment_id": c.id, "reason": "author_mismatch"}));
            continue;
        }
        match kind {
            CommentKind::Claim {
                session,
                epoch,
                lease_until,
            } => {
                // The same session and epoch claimed again: the first stands.
                if claims
                    .iter()
                    .any(|v| v.session == session && v.epoch == epoch)
                {
                    invalid.push(json!({"comment_id": c.id, "reason": "duplicate"}));
                    continue;
                }
                let live = parse_utc(&lease_until).is_some_and(|t| t > who.now);
                claims.push(ValidClaim {
                    id: c.id,
                    session,
                    epoch,
                    live,
                });
            }
            CommentKind::Release { target } | CommentKind::Handoff { target, .. } => {
                // Only an earlier valid claim can be closed.
                if !claims.iter().any(|v| v.id == target) {
                    invalid.push(json!({"comment_id": c.id, "reason": "malformed",
                                        "detail": "closes no earlier claim comment"}));
                    continue;
                }
                let (by, to) = match &kind {
                    CommentKind::Handoff { to, .. } => ("handoff", Some(to.clone())),
                    _ => ("release", None),
                };
                closed.push(
                    json!({"comment_id": target, "by": by, "to": to, "closing_comment_id": c.id}),
                );
            }
            CommentKind::Malformed(detail) => {
                invalid.push(json!({"comment_id": c.id, "reason": "malformed", "detail": detail}));
            }
            CommentKind::Other => unreachable!(),
        }
    }
    let is_closed = |id: i64| closed.iter().any(|c| c["comment_id"] == json!(id));
    let mine = match who.comment_id {
        Some(id) => claims.iter().find(|v| v.id == id),
        None => claims.iter().find(|v| v.session == who.session),
    };
    let holder = claims.iter().find(|v| !is_closed(v.id));
    let (status, lost_to) = if let (Some(id), None) = (who.comment_id, mine) {
        let visible = comments.iter().any(|c| c.id == id);
        (if visible { "no_claim" } else { "uncertain" }, None)
    } else if mine.is_some_and(|m| is_closed(m.id)) {
        ("closed", None)
    } else {
        match holder {
            None => ("no_claim", None),
            // An expired holder stays the holder: no takeover by time alone (REQ-A9-03).
            Some(h) if !h.live => ("stale", None),
            Some(h) if mine.is_some_and(|m| m.id == h.id) => ("owner", None),
            Some(h) => ("not_owner", Some(h.id)),
        }
    };
    json!({
        "status": status,
        "holder": holder.map(|h| h.id),
        "holder_live": holder.map(|h| h.live),
        "mine": mine.map(|m| m.id),
        "lost_to": lost_to,
        "invalid": invalid,
        "closed": closed,
        "claim_lines": seen_claim_lines,
        "auto_assigned": false,
    })
}

/// The refusal message for a non-owner verdict; `None` for `owner`.
pub fn refusal(verdict: &Value) -> Option<String> {
    let id = |k: &str| {
        verdict[k]
            .as_i64()
            .map_or("none".to_string(), |v| v.to_string())
    };
    Some(match verdict["status"].as_str()? {
        "owner" => return None,
        "not_owner" => format!(
            "claim verify: not owner; the lowest open claim is comment {} (REQ-A9-01), so no commit or push",
            id("lost_to")
        ),
        "stale" => format!(
            "claim verify: stale; claim comment {} lease expired without a handoff, and the operator decides (REQ-A9-03)",
            id("holder")
        ),
        "closed" => format!(
            "claim verify: closed; claim comment {} was released or handed off",
            id("mine")
        ),
        "uncertain" => "claim verify: uncertain; this session's claim comment is not visible on the issue".into(),
        _ => "claim verify: no claim of this session on the issue".into(),
    })
}

/// `claim-verify` stdin (typed JSON, unknown keys rejected).
#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct VerifyRequest {
    pub schema_version: u32,
    pub owner: String,
    pub repo: String,
    pub issue_number: i64,
    pub session: String,
    #[serde(default)]
    pub comment_id: Option<i64>,
    #[serde(default)]
    pub allowed_authors: Vec<String>,
}

/// Every comment of the Issue (GET only, `Link` pagination).
pub fn read_comments<T: Transport, C: Clock>(
    client: &Client<'_, T, C>,
    owner: &str,
    repo: &str,
    issue: i64,
) -> Result<Vec<IssueComment>> {
    let mut url = Some(format!(
        "https://api.github.com/repos/{owner}/{repo}/issues/{issue}/comments?per_page=100"
    ));
    let mut out = Vec::new();
    let mut pages = 0;
    while let Some(u) = url {
        pages += 1;
        if pages > MAX_PAGES {
            return Err("claim comment page limit exceeded (50)".into());
        }
        let resp = client.get(&u, None).map_err(String::from)?;
        for c in resp
            .body
            .as_ref()
            .and_then(Value::as_array)
            .into_iter()
            .flatten()
        {
            if let (Some(id), Some(body)) = (c["id"].as_i64(), c["body"].as_str()) {
                out.push(IssueComment {
                    id,
                    author: c["user"]["login"].as_str().unwrap_or("").to_string(),
                    body: body.to_string(),
                });
            }
        }
        url = resp.header("link").and_then(|l| next_link(Some(l)));
    }
    Ok(out)
}

fn check_request(req: &VerifyRequest) -> Result<()> {
    if req.schema_version != 1 {
        return Err("unsupported request schema_version; supported: 1".into());
    }
    let name = |s: &str| {
        !s.is_empty()
            && s.len() <= 100
            && s.chars()
                .all(|c| c.is_ascii_alphanumeric() || "-_.".contains(c))
    };
    if !name(&req.owner) || !name(&req.repo) || req.issue_number <= 0 || req.session.is_empty() {
        return Err(
            "invalid claim verify request: owner, repo, issue_number and session are required"
                .into(),
        );
    }
    Ok(())
}

/// `claim-verify` with a typed request: the verdict when this session owns the Issue,
/// otherwise the refusal as an error (exit 1, classified).
pub fn verify_cli<T: Transport, C: Clock>(
    req: &VerifyRequest,
    transport: &T,
    clock: &C,
) -> Result<Value> {
    check_request(req)?;
    let client = Client::new(transport, clock, Default::default(), Default::default());
    let comments = read_comments(&client, &req.owner, &req.repo, req.issue_number)?;
    let verdict = evaluate(
        &comments,
        &Verifier {
            session: &req.session,
            comment_id: req.comment_id,
            allowed_authors: &req.allowed_authors,
            now: clock.now(),
        },
    );
    match refusal(&verdict) {
        None => Ok(verdict),
        Some(message) => Err(message),
    }
}

/// Words that run the command after them (#2306). Their options, `NAME=value` assignments and
/// a leading number or duration (`nice -n 10`, `timeout 30`) are skipped to reach the command.
const WRAPPERS: &[&str] = &[
    "builtin",
    "caffeinate",
    "chrt",
    "command",
    "doas",
    "env",
    "exec",
    "ionice",
    "nice",
    "nohup",
    "setsid",
    "stdbuf",
    "sudo",
    "time",
    "timeout",
    "xargs",
];

/// A wrapper's options that take the next word as their value (`sudo -u lane`, `xargs -I {}`).
/// Kept per wrapper: a flag misread as value-taking would swallow the `git` after it
/// (`time -p git commit`, `env -i git push`).
fn value_options(wrapper: &str) -> &'static [&'static str] {
    match wrapper {
        "caffeinate" => &["-t", "-w"],
        "doas" => &["-C", "-u"],
        "env" => &["-C", "-u"],
        "exec" => &["-a"],
        "ionice" => &["-c", "-n", "-P", "-p", "-u"],
        "nice" => &["-n"],
        "stdbuf" => &["-e", "-i", "-o"],
        "sudo" => &[
            "-C", "-D", "-g", "-h", "-p", "-R", "-r", "-T", "-t", "-U", "-u",
        ],
        "timeout" => &["-k", "-s"],
        "xargs" => &["-a", "-d", "-E", "-I", "-L", "-n", "-P", "-s"],
        _ => &[],
    }
}

/// Shells whose `-c` argument is itself a command line, judged one level deep.
const SHELLS: &[&str] = &["bash", "dash", "ksh", "sh", "zsh"];

/// Splits a shell command line into segments of words at unquoted `;`, `&`, `|`, newline and
/// parentheses, removing quotes and backslash escapes the way the shell does (#2306). This is
/// not a full shell parser: an unbalanced quote runs to the end of the line, and expansions
/// (`$(...)`, `$VAR`, globs, aliases) are left as literal text.
fn shell_segments(command: &str) -> Vec<Vec<String>> {
    let mut segments = Vec::new();
    let mut words = Vec::new();
    let mut word = String::new();
    let mut in_word = false;
    let mut chars = command.chars();
    while let Some(c) = chars.next() {
        match c {
            '\\' => {
                if let Some(next) = chars.next() {
                    word.push(next);
                }
                in_word = true;
            }
            '\'' => {
                for next in chars.by_ref() {
                    if next == '\'' {
                        break;
                    }
                    word.push(next);
                }
                in_word = true;
            }
            '"' => {
                while let Some(next) = chars.next() {
                    match next {
                        '"' => break,
                        '\\' => word.extend(chars.next()),
                        other => word.push(other),
                    }
                }
                in_word = true;
            }
            ';' | '&' | '|' | '\n' | '(' | ')' => {
                if in_word {
                    words.push(std::mem::take(&mut word));
                    in_word = false;
                }
                segments.push(std::mem::take(&mut words));
            }
            c if c.is_whitespace() => {
                if in_word {
                    words.push(std::mem::take(&mut word));
                    in_word = false;
                }
            }
            other => {
                word.push(other);
                in_word = true;
            }
        }
    }
    if in_word {
        words.push(word);
    }
    segments.push(words);
    segments
}

/// The last path component of a command word (`/usr/bin/git` → `git`).
fn command_name(word: &str) -> &str {
    word.rsplit('/').next().unwrap_or(word)
}

/// True when a word is a number or a duration (`10`, `-10`, `30s`, `1.5m`).
fn numeric_word(word: &str) -> bool {
    let digits = word.trim_start_matches(['-', '+']);
    let body = digits.trim_end_matches(['s', 'm', 'h', 'd']);
    !body.is_empty() && body.chars().all(|c| c.is_ascii_digit() || c == '.')
}

/// The index of the word a segment actually runs: past `NAME=value` assignments and past any
/// wrapper words together with their options and values.
fn command_index(words: &[String]) -> Option<usize> {
    let assignment = |w: &str| w.contains('=') && !w.starts_with('-');
    let mut i = 0;
    while i < words.len() && assignment(&words[i]) {
        i += 1;
    }
    while i < words.len() && WRAPPERS.contains(&command_name(&words[i])) {
        let takes_value = value_options(command_name(&words[i]));
        i += 1;
        while i < words.len() {
            let w = words[i].as_str();
            if takes_value.contains(&w) {
                i += 2;
            } else if w.starts_with('-') || assignment(w) || numeric_word(w) {
                i += 1;
            } else {
                break;
            }
        }
    }
    (i < words.len()).then_some(i)
}

/// The script a shell runs with `-c` (`sh -c '...'`, `bash -lc "..."`), if any. Options that
/// take a value (`bash -o pipefail -c …`, `sh -O extglob -c …`, `bash +o posix -c …`) consume
/// their next word so the scan reaches the `-c` script (rc.2 A9 follow-up 2 #2323). Bundled
/// short options ending in `o`/`O` take the value too (`bash -eo pipefail -c …`, `bash -euo
/// pipefail -c …`; rc.2 A9 follow-up 3 #2330).
fn shell_script(args: &[String]) -> Option<&str> {
    let mut iter = args.iter();
    while let Some(arg) = iter.next() {
        if matches!(arg.as_str(), "-o" | "-O" | "+o" | "+O") {
            // The next word is this option's value, not a non-option word.
            iter.next();
            continue;
        }
        // A bundled flag word (`-eo`, `-euo`) carrying `-` or `+`: when it ends in `o`/`O`
        // its last short option takes a value, so the next word is that value.
        let flags = arg.strip_prefix('-').or_else(|| arg.strip_prefix('+'));
        match flags {
            Some(flags) if !flags.starts_with('-') && flags.contains('c') => {
                return iter.next().map(String::as_str)
            }
            Some(flags)
                if !flags.starts_with('-') && (flags.ends_with('o') || flags.ends_with('O')) =>
            {
                iter.next();
                continue;
            }
            Some(_) => continue,
            None => return None,
        }
    }
    None
}

/// True when `git`'s subcommand is one of `verbs`, directly or through an alias defined on
/// the same command line (`git -c alias.ci=commit ci`). An alias defined in git config is not
/// visible here.
fn git_runs_verb(args: &[String], verbs: &[&str]) -> bool {
    let aliases = inline_aliases(args);
    // Only the subcommand position counts: `git log --grep commit` or `git log -- commit`
    // names the word as an argument of a read command, not as the command (#2399 review).
    git_subcommand_candidates(args).iter().any(|w| {
        verbs.contains(w)
            || aliases.iter().any(|(name, words)| {
                name == w && words.first().is_some_and(|v| verbs.contains(&v.as_str()))
            })
    })
}

/// Aliases defined on the same command line (`git -c alias.ci=commit ci`, also the attached
/// `-calias.ci=commit` form), as (name, expansion words). A `!` shell alias that runs `git`
/// expands to the words after `git`. An alias defined in git config is not visible here.
fn inline_aliases(args: &[String]) -> Vec<(String, Vec<String>)> {
    let mut out = Vec::new();
    let mut i = 0;
    while i < args.len() {
        let config = match args[i].as_str() {
            "-c" => {
                i += 1;
                args.get(i).map(String::as_str)
            }
            other => other.strip_prefix("-c").filter(|v| v.starts_with("alias.")),
        };
        if let Some((name, value)) = config
            .and_then(|c| c.strip_prefix("alias."))
            .and_then(|c| c.split_once('='))
        {
            let mut words: Vec<String> = value
                .trim_start_matches('!')
                .split_whitespace()
                .map(str::to_string)
                .collect();
            if words.first().is_some_and(|w| command_name(w) == "git") {
                words.remove(0);
            }
            out.push((name.to_string(), words));
        }
        i += 1;
    }
    out
}

/// git(1) OPTIONS (git 2.54.0, read 2026-11-10) that take their value as the next word.
const GIT_GLOBAL_OPTIONS_WITH_VALUE: &[&str] = &[
    "-C",
    "-c",
    "--config-env",
    "--git-dir",
    "--work-tree",
    "--namespace",
    "--list-cmds",
    "--attr-source",
];

/// git(1) OPTIONS (git 2.54.0) that never take a separate value word. `--exec-path` takes its
/// value only as `--exec-path=<path>`.
const GIT_GLOBAL_FLAGS: &[&str] = &[
    "-v",
    "--version",
    "-h",
    "--help",
    "--exec-path",
    "--html-path",
    "--man-path",
    "--info-path",
    "-p",
    "--paginate",
    "-P",
    "--no-pager",
    "--bare",
    "--no-replace-objects",
    "--no-lazy-fetch",
    "--no-optional-locks",
    "--no-advice",
    "--literal-pathspecs",
    "--glob-pathspecs",
    "--noglob-pathspecs",
    "--icase-pathspecs",
];

/// The words that may be a `git` argument list's subcommand. Known value options skip their
/// value and known flags stand alone; the first other word is the subcommand. An option this
/// list does not know (a newer git, a typo) may or may not take the next word as its value,
/// so both that word and the subcommand found after it are candidates: the guard errs toward
/// seeing a commit or push rather than missing one (#2399 audit B5, `--attr-source`). A second
/// unknown option ends the scan, so the arguments of a subcommand that was already a
/// candidate (`git --new log --grep commit`) are not read as a subcommand.
fn git_subcommand_candidates(args: &[String]) -> Vec<&str> {
    git_subcommand_positions(args)
        .into_iter()
        .map(|i| args[i].as_str())
        .collect()
}

/// The indexes in `args` of [`git_subcommand_candidates`] (the one parser; the hook guards of
/// #2391 stage 5 read each subcommand's own arguments from these positions).
fn git_subcommand_positions(args: &[String]) -> Vec<usize> {
    let mut out = Vec::new();
    let mut seen_unknown = false;
    let mut i = 0;
    while i < args.len() {
        let w = args[i].as_str();
        if GIT_GLOBAL_OPTIONS_WITH_VALUE.contains(&w) {
            i += 2;
        } else if !w.starts_with('-') {
            out.push(i);
            return out;
        } else if GIT_GLOBAL_FLAGS.contains(&w) || w.contains('=') || w.starts_with("-c") {
            // `--opt=value` and the attached `-c<name>=<value>` form carry their own value.
            i += 1;
        } else if seen_unknown {
            return out;
        } else {
            // Unknown option: its next word is either its value or the subcommand.
            seen_unknown = true;
            if args.get(i + 1).is_some_and(|n| !n.starts_with('-')) {
                out.push(i + 1);
            }
            i += 2;
        }
    }
    out
}

/// Every command a shell line runs, as (name without its path, arguments), past assignments
/// and wrapper words, with a shell's `-c` script expanded one level deep — the same reading
/// `runs_git` uses. Quoted text stays one argument and is never a command.
pub(crate) fn shell_commands(command: &str) -> Vec<(String, Vec<String>)> {
    let mut out = Vec::new();
    collect_commands(command, 0, &mut out);
    out
}

fn collect_commands(command: &str, depth: u8, out: &mut Vec<(String, Vec<String>)>) {
    for words in shell_segments(command) {
        let Some(i) = command_index(&words) else {
            continue;
        };
        let name = command_name(&words[i]).to_string();
        let args = words[i + 1..].to_vec();
        if depth == 0 {
            if let Some(script) = script_run_by(&name, &args) {
                collect_commands(&script, depth + 1, out);
            }
        }
        out.push((name, args));
    }
}

/// The script a command runs as shell code: a shell's `-c` script, or `eval`'s arguments
/// joined with spaces. The one place that decides which quoted text is code, shared by the
/// command reading above and [`executed_scripts`].
fn script_run_by(name: &str, args: &[String]) -> Option<String> {
    if SHELLS.contains(&name) {
        return shell_script(args).map(str::to_string);
    }
    (name == "eval" && !args.is_empty()).then(|| args.join(" "))
}

/// The shell text a command line runs: the line itself, then every script it hands to a
/// shell (`sh -c '…'`) or to `eval`, one level deep — the same reading as
/// [`shell_commands`]. A check that reads raw text (a fork bomb is punctuation, not a
/// command word) looks at each of these, so quoting a script for `sh -c` does not hide it.
pub(crate) fn executed_scripts(command: &str) -> Vec<String> {
    let mut out = vec![command.to_string()];
    for words in shell_segments(command) {
        let Some(i) = command_index(&words) else {
            continue;
        };
        if let Some(script) = script_run_by(command_name(&words[i]), &words[i + 1..]) {
            out.push(script);
        }
    }
    out
}

/// One git subcommand a `git` argument list may run: the verb, its arguments, and git's own
/// arguments before it. An alias defined on the same line (`-c alias.ci=commit ci`) is
/// resolved to its verb, with the words the alias carries put before the call's arguments.
pub(crate) struct GitSub {
    pub verb: String,
    pub rest: Vec<String>,
    pub global: Vec<String>,
}

/// Each git subcommand a `git` argument list may run (see [`GitSub`]).
pub(crate) fn git_subcommands(args: &[String]) -> Vec<GitSub> {
    let aliases = inline_aliases(args);
    git_subcommand_positions(args)
        .into_iter()
        .map(|i| {
            let global = args[..i].to_vec();
            let rest = &args[i + 1..];
            match aliases.iter().find(|(name, _)| *name == args[i]) {
                Some((_, words)) if !words.is_empty() => GitSub {
                    verb: words[0].clone(),
                    rest: words[1..]
                        .iter()
                        .cloned()
                        .chain(rest.iter().cloned())
                        .collect(),
                    global,
                },
                _ => GitSub {
                    verb: args[i].clone(),
                    rest: rest.to_vec(),
                    global,
                },
            }
        })
        .collect()
}

/// True when one segment of a shell command line runs `git <verb>`. The segment's command is
/// found past assignments and wrapper words (`env`, `command`, `nice`, `sudo`, `xargs`, ...),
/// after quotes and backslashes are removed (`\git`, `"git"`); a shell's `-c` script is judged
/// one level deep. A `git` that is only an argument (`echo git commit`) does not count.
fn runs_git(command: &str, verbs: &[&str]) -> bool {
    runs_git_at(command, verbs, 0)
}

fn runs_git_at(command: &str, verbs: &[&str], depth: u8) -> bool {
    shell_segments(command).iter().any(|words| {
        let Some(i) = command_index(words) else {
            return false;
        };
        let name = command_name(&words[i]);
        let args = &words[i + 1..];
        if name == "git" {
            return git_runs_verb(args, verbs);
        }
        depth == 0
            && SHELLS.contains(&name)
            && shell_script(args).is_some_and(|script| runs_git_at(script, verbs, depth + 1))
    })
}

/// True when a shell command line runs `git commit` or `git push` in any of its segments.
pub fn is_commit_or_push(command: &str) -> bool {
    runs_git(command, &["commit", "push"])
}

/// True when a shell command line runs `git commit` in any of its segments.
pub fn is_commit(command: &str) -> bool {
    runs_git(command, &["commit"])
}

/// Why a commit command line on a lane branch carries no usable claim token.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum CommitTokenGap {
    /// Tokens are present, but none names the branch's Issue.
    OtherIssue,
    /// No `Moai-Claim:` token at all.
    Missing,
}

impl CommitTokenGap {
    /// The refusal reason. Built from fixed text and the branch's Issue number only, so it
    /// never echoes the command line.
    pub fn reason(self, issue: i64) -> String {
        match self {
            CommitTokenGap::OtherIssue => {
                format!("the commit message names another Issue than the branch's #{issue}")
            }
            CommitTokenGap::Missing => format!(
                "the commit message carries no `Moai-Claim: {issue}#<id>` (a message from \
                 -F or an editor is not read)"
            ),
        }
    }
}

/// The claim comment id a commit command line names for `issue` (the highest when several),
/// read from the command text without any forge call. The single decider for the local part
/// of the commit guard: `hook_cli` and the hook engine's tool-preflight both use it (#2391).
pub fn commit_claim_id(command: &str, issue: i64) -> std::result::Result<i64, CommitTokenGap> {
    let tokens = claim_tokens(command);
    match tokens.iter().filter(|t| t.0 == issue).map(|t| t.1).max() {
        Some(id) => Ok(id),
        None if !tokens.is_empty() => Err(CommitTokenGap::OtherIssue),
        None => Err(CommitTokenGap::Missing),
    }
}

/// The current branch of the worktree at `cwd`, or `None` on a detached HEAD or outside git.
pub fn current_branch(cwd: &Path) -> Option<String> {
    git_out(cwd, &["symbolic-ref", "--short", "-q", "HEAD"])
}

/// True when this clone's local claim journal has an active record whose worktree contains
/// `cwd` (local read only; the record path of `hook_cli` then decides with the forge).
pub fn has_local_record(project: &Path, cwd: &Path) -> bool {
    let records = crate::worktree::active_records(project).unwrap_or_default();
    record_for(records, cwd).is_some()
}

/// How a branch name reads as a lane branch (#2336).
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum LaneBranch {
    /// `issue/<n>` or `issue/<n>-<key>`: the Issue number comes first.
    Issue(i64),
    /// Not under `issue/`: not a lane branch.
    NotLane,
    /// Under `issue/` but without a leading Issue number — a letter-led key, a leading zero
    /// (`issue/06-C-2007-e1`, the old `worktree-plan` shape), or digits running into text.
    Unparseable,
}

/// Classifies a branch name. Every lane branch and every branch `worktree-plan` names put the
/// Issue number first (`issue/<n>-<key>`, `<prefix>/<n>-<key>-e<epoch>`); an Issue number has
/// no leading zero, so `issue/06-…` is a key, not Issue 6.
pub fn lane_branch(branch: &str) -> LaneBranch {
    let Some(rest) = branch.strip_prefix("issue/") else {
        return LaneBranch::NotLane;
    };
    let digits: String = rest.chars().take_while(char::is_ascii_digit).collect();
    let tail = &rest[digits.len()..];
    if digits.is_empty() || digits.starts_with('0') || !(tail.is_empty() || tail.starts_with('-')) {
        return LaneBranch::Unparseable;
    }
    match digits.parse() {
        Ok(n) => LaneBranch::Issue(n),
        Err(_) => LaneBranch::Unparseable,
    }
}

/// The Issue number of a lane branch `issue/<n>-<key>` (or `issue/<n>`).
pub fn issue_from_branch(branch: &str) -> Option<i64> {
    match lane_branch(branch) {
        LaneBranch::Issue(n) => Some(n),
        LaneBranch::NotLane | LaneBranch::Unparseable => None,
    }
}

/// Owner and repository of a GitHub remote URL: `https://github.com/o/r(.git)`,
/// `git@github.com:o/r(.git)` or `ssh://git@github.com/o/r(.git)`.
pub fn github_owner_repo(url: &str) -> Option<(String, String)> {
    let url = url.trim();
    let path = url
        .strip_prefix("https://github.com/")
        .or_else(|| url.strip_prefix("git@github.com:"))
        .or_else(|| url.strip_prefix("ssh://git@github.com/"))?;
    let path = path.strip_suffix(".git").unwrap_or(path);
    let (owner, repo) = path.split_once('/')?;
    let ok = |s: &str| {
        !s.is_empty()
            && s.len() <= 100
            && s.chars()
                .all(|c| c.is_ascii_alphanumeric() || "-_.".contains(c))
    };
    (ok(owner) && ok(repo)).then(|| (owner.to_string(), repo.to_string()))
}

/// Every `Moai-Claim: <issue>#<comment-id>` token in a text, in order. The token is found
/// anywhere (a commit command carries it inside a quoted `-m` argument), digits only.
pub fn claim_tokens(text: &str) -> Vec<(i64, i64)> {
    let mut out = Vec::new();
    for (at, _) in text.match_indices("Moai-Claim:") {
        let rest = text[at + "Moai-Claim:".len()..].trim_start_matches([' ', '\t']);
        let issue: String = rest.chars().take_while(char::is_ascii_digit).collect();
        let Some(after) = rest[issue.len()..].strip_prefix('#') else {
            continue;
        };
        let id: String = after.chars().take_while(char::is_ascii_digit).collect();
        if let (Ok(n), Ok(id)) = (issue.parse(), id.parse()) {
            out.push((n, id));
        }
    }
    out
}

fn git_out(cwd: &Path, args: &[&str]) -> Option<String> {
    let mut cmd = std::process::Command::new("git");
    crate::identity::scrub_git_env(&mut cmd);
    let out = cmd.arg("-C").arg(cwd).args(args).output().ok()?;
    out.status
        .success()
        .then(|| String::from_utf8_lossy(&out.stdout).trim().to_string())
}

/// The claim tokens a push accepts on the branch's commits: the verifier's own claim and
/// every earlier claim of the Issue closed by a handoff line, so commits written under a
/// previous epoch stay valid after the handoff.
fn accepted_tokens(issue: i64, verdict: &Value) -> Vec<String> {
    let mut ids: Vec<i64> = verdict["mine"].as_i64().into_iter().collect();
    for c in verdict["closed"].as_array().into_iter().flatten() {
        if c["by"] == json!("handoff") {
            ids.extend(c["comment_id"].as_i64());
        }
    }
    ids.iter()
        .map(|id| crate::claim::claim_token(issue, *id))
        .collect()
}

/// The active claim record whose worktree contains `cwd`.
fn record_for(records: Vec<ClaimRecord>, cwd: &Path) -> Option<ClaimRecord> {
    records
        .into_iter()
        .find(|r| std::fs::canonicalize(&r.worktree.path).is_ok_and(|p| cwd.starts_with(p)))
}

/// `claim-verify --hook`: the host `PreToolUse` guard. Allows everything except a Bash
/// `git commit` / `git push` inside a worktree with an active claim record, which it
/// verifies; a refusal is an error the CLI turns into exit 2 so the host blocks the call.
pub fn hook_cli<T: Transport, C: Clock>(
    project: &Path,
    stdin: &[u8],
    transport: &T,
    clock: &C,
) -> Result<Value> {
    let allow = |reason: &str| Ok(json!({"allowed": true, "reason": reason}));
    let input: Value = serde_json::from_slice(stdin).unwrap_or(Value::Null);
    let command = input["tool_input"]["command"].as_str().unwrap_or("");
    if input["tool_name"] != json!("Bash") || !is_commit_or_push(command) {
        return allow("not a git commit or push");
    }
    let cwd = input["cwd"]
        .as_str()
        .map(PathBuf::from)
        .or_else(|| std::env::current_dir().ok())
        .and_then(|p| std::fs::canonicalize(p).ok());
    let Some(cwd) = cwd else {
        return allow("working directory unavailable; no claimed worktree to guard");
    };
    let records = crate::worktree::active_records(project).unwrap_or_default();
    let push = runs_git(command, &["push"]);
    // A journal record, when this clone has one, decides: its comment id and base.
    if let Some(record) = record_for(records, &cwd) {
        let issue = record.task_ref.issue_number;
        let req = VerifyRequest {
            schema_version: 1,
            owner: record.task_ref.owner.clone(),
            repo: record.task_ref.repo.clone(),
            issue_number: issue,
            session: record.owner.id.clone(),
            comment_id: record.comment_id,
            allowed_authors: Vec::new(),
        };
        let mut verdict = verify_cli(&req, transport, clock)?;
        if push && record.claim_token.is_some() {
            let tokens = accepted_tokens(issue, &verdict);
            verdict["trailer_checked"] = json!(missing_trailers_since(
                &cwd,
                &record.base_revision,
                &tokens
            )?);
        }
        verdict["allowed"] = json!(true);
        verdict["issue"] = json!(issue);
        verdict["source"] = json!("journal");
        return Ok(verdict);
    }
    // No record (the claim was posted for the lane, not by `claim-acquire`): a lane branch
    // `issue/<n>-…` is still guarded, from the branch name and the `Moai-Claim` token.
    let Some(branch) = current_branch(&cwd) else {
        return allow("no claim record and no branch (detached HEAD); not a lane worktree");
    };
    let issue = match lane_branch(&branch) {
        LaneBranch::Issue(n) => n,
        LaneBranch::NotLane => {
            return allow("no claim record and the branch is not issue/<n>; not a lane worktree")
        }
        // Still allowed: fail-open versus fail-closed is decision D5 (HOLD). The verdict
        // only makes the unread branch visible instead of passing it as "not a lane".
        LaneBranch::Unparseable => {
            return Ok(json!({
                "allowed": true,
                "branch": "unparseable",
                "reason": format!(
                    "no claim record and the lane branch {branch} has no leading Issue number \
                     (issue/<n>-<key>); not guarded (fail-open, decision D5 HOLD)"
                ),
            }))
        }
    };
    let uncertain = |why: String| -> Result<Value> {
        Err(format!(
            "claim verify: uncertain; {why}; ask the leader for the claim comment id and put \
             `Moai-Claim: {issue}#<id>` in the commit message"
        ))
    };
    let Some((owner, repo)) =
        git_out(&cwd, &["remote", "get-url", "origin"]).and_then(|u| github_owner_repo(&u))
    else {
        return uncertain("the origin remote is not a GitHub repository".into());
    };
    let comment_id = if runs_git(command, &["commit"]) {
        // The commit being written: its message, as typed in the command line.
        match commit_claim_id(command, issue) {
            Ok(id) => id,
            Err(gap) => return uncertain(gap.reason(issue)),
        }
    } else {
        // A push: the highest claim comment id among this Issue's tokens on the branch's
        // first-parent line, merge commits included. A later epoch's claim always has a
        // higher comment id, and after a renewal its token may sit only on a merge commit;
        // taking the newest non-merge commit instead verified the handed-off epoch and
        // refused the owner as `closed`.
        let log = git_out(
            &cwd,
            &[
                "log",
                "--first-parent",
                "-n",
                "200",
                "--format=%B%x1e",
                "HEAD",
            ],
        )
        .unwrap_or_default();
        let found = claim_tokens(&log)
            .iter()
            .filter(|t| t.0 == issue)
            .map(|t| t.1)
            .max();
        match found {
            Some(id) => id,
            None => {
                return uncertain(format!(
                    "no commit on the branch carries `Moai-Claim: {issue}#<id>`"
                ))
            }
        }
    };
    let req = VerifyRequest {
        schema_version: 1,
        owner,
        repo,
        issue_number: issue,
        session: "unverified".into(),
        comment_id: Some(comment_id),
        allowed_authors: Vec::new(),
    };
    let mut verdict = verify_cli(&req, transport, clock)?;
    if push {
        // No recorded base. The lane's own commits are those not on any other `origin`
        // branch: the integration branch need not share history with the default branch
        // (this repository's `v2026.11.0` and `main` have none), so no single base is assumed.
        // With no other remote branch every commit is checked, and the push is refused.
        // With `--remotes`, `--exclude` matches the name without `refs/remotes/` (git
        // rev-parse docs), so the branch's own remote ref is `origin/<branch>`.
        let own = format!("--exclude=origin/{branch}");
        let range = ["HEAD", "--not", own.as_str(), "--remotes=origin"];
        let tokens = accepted_tokens(issue, &verdict);
        verdict["trailer_checked"] = json!(missing_trailers(&cwd, &range, &tokens)?);
    }
    verdict["allowed"] = json!(true);
    verdict["issue"] = json!(issue);
    verdict["source"] = json!("branch");
    Ok(verdict)
}

/// Before a push, every commit the lane wrote on the branch's own first-parent line since the
/// base must carry one of the accepted claim tokens. Merge commits are excluded (decision
/// 2026-11-09 06:02, #1975: the claim line is required on lane-written commits, a merge commit
/// is only a record), and so are the merged-in commits of other cards. Returns the number of
/// commits checked; refuses with the short SHAs that lack a token.
fn missing_trailers_since(cwd: &Path, base: &str, tokens: &[String]) -> Result<usize> {
    if base.len() != 40 || !base.bytes().all(|b| b.is_ascii_hexdigit()) {
        return Err(
            "claim verify: trailer check unavailable; the claim base is not a commit".into(),
        );
    }
    missing_trailers(cwd, &[&format!("{base}..HEAD")], tokens)
}

/// [`missing_trailers_since`] over a revision range given as `git log` arguments.
fn missing_trailers(cwd: &Path, range: &[&str], tokens: &[String]) -> Result<usize> {
    let mut cmd = std::process::Command::new("git");
    crate::identity::scrub_git_env(&mut cmd);
    let out = cmd
        .arg("-C")
        .arg(cwd)
        .args([
            "log",
            "--first-parent",
            "--no-merges",
            "--format=%H%x00%B%x1e",
        ])
        .args(range)
        .output()
        .map_err(|_| "claim verify: trailer check unavailable; git could not run")?;
    if !out.status.success() {
        return Err(
            "claim verify: trailer check unavailable; git log of the claim base failed".into(),
        );
    }
    let text = String::from_utf8_lossy(&out.stdout);
    let mut checked = 0;
    let mut lacking = Vec::new();
    for entry in text
        .split('\u{1e}')
        .map(str::trim)
        .filter(|e| !e.is_empty())
    {
        let (sha, body) = entry.split_once('\0').unwrap_or((entry, ""));
        checked += 1;
        if !body.lines().any(|l| tokens.iter().any(|t| l.trim() == t)) {
            lacking.push(sha.get(..9).unwrap_or(sha).to_string());
        }
    }
    if lacking.is_empty() {
        Ok(checked)
    } else {
        // At most 10 SHAs are named: a range with no other remote branch can span history.
        let more = lacking.len().saturating_sub(10);
        lacking.truncate(10);
        Err(format!(
            "claim verify: trailer missing; commits {}{} lack `{}`, so no push",
            lacking.join(", "),
            if more > 0 {
                format!(" and {more} more")
            } else {
                String::new()
            },
            tokens.join("` or `")
        ))
    }
}
