//! `agents-merge-plan|apply|revert` (#2936 R10): when a project has an `AGENTS.local.md` with
//! text the user wrote, `/moai update` moves that text into `CLAUDE.local.md`, the personal file
//! Claude Code reads and the project keeps out of git, once. A missing `CLAUDE.local.md` is
//! created; its revert removes it again. The text never goes into a shared file.
//!
//! Where the text comes from: the file in place when it holds more than the text MoAI
//! generates (`agents::is_generated_local`); otherwise the copy a committed `go-state-apply`
//! moved into its backup's `root-files/`. Nothing is moved or deleted here: a file in place
//! stays and a copy goes into the receipt folder; the moved copy is already in its dated
//! backup. The text lands at the end of `CLAUDE.local.md` between two marker lines:
//!
//! ```text
//! <!-- moai:merged-local sha256=<sha256 of the source bytes> -->
//! ...the text...
//! <!-- /moai:merged-local -->
//! ```
//!
//! The markers make a second run `unchanged`. A block edited by hand, or a source that changed
//! since the merge, is a `conflict`: nothing is overwritten. The preview shows the whole text,
//! its position and that it goes into a committed file; a line shaped like a credential makes
//! the plan `blocked` (the value is never printed). `plan_digest` binds the source bytes, the
//! `CLAUDE.local.md` bytes and the action.
//!
//! Claude Code reads `AGENTS.md` only while no `CLAUDE.md`, `.claude/CLAUDE.md` or
//! `CLAUDE.local.md` exists. A merge that creates `CLAUDE.local.md` over an `AGENTS.md` hides that
//! file, so the plan carries a warning the user sees before they confirm.
//!
//! A receipt written before the personal file moved holds `AGENTS.md` as its target; it has no
//! `target_path`, and its revert still restores that file.
//!
//! The same engine has a second kind, [`Kind::Guidance`]: the text a person wrote in `AGENTS.md`
//! (outside MoAI's managed section and outside blocks an earlier merge wrote) goes into `CLAUDE.md`,
//! the team file, between `moai:merged-agents` markers, with the same preview, receipt and revert.
//! `AGENTS.md` is never written. A `CLAUDE.md` that already imports `AGENTS.md` is `unchanged`.

use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::{fs, path::Path};

type Result<T> = std::result::Result<T, String>;

const TARGET: &str = "CLAUDE.local.md";
/// The target of a receipt written before the personal file moved; also the source of the
/// guidance kind.
const LEGACY_TARGET: &str = "AGENTS.md";
const LOCAL: &str = "AGENTS.local.md";
const BEGIN: &str = "<!-- moai:merged-local sha256=";
const BEGIN_TAIL: &str = " -->";
const END: &str = "<!-- /moai:merged-local -->";
const GUIDANCE_TARGET: &str = "CLAUDE.md";
const GUIDANCE_BEGIN: &str = "<!-- moai:merged-agents sha256=";
const GUIDANCE_END: &str = "<!-- /moai:merged-agents -->";
const GUIDANCE_PREFIX: &str = "agents-md-migrate-";
/// The start of this generation's managed-section markers (`agents.rs`); a source that holds it
/// would be read as a second managed section.
const MANAGED_MARKER: &str = "<!-- moai:managed";
/// The most a source file may hold.
const SOURCE_LIMIT: usize = 128 * 1024;
const RECEIPT_PREFIX: &str = "agents-local-merge-";
/// The one line the preview shows about where the text goes. The key is still `commit_notice`:
/// `project_update` forwards it under that name.
pub const COMMIT_NOTICE: &str = "This text goes into CLAUDE.local.md, a personal file that is \
     not committed to the repository (keep it in .gitignore); AGENTS.local.md stays as it is.";
/// The same line for the guidance kind: `CLAUDE.md` is the team file everyone reads.
const GUIDANCE_COMMIT_NOTICE: &str =
    "This text goes into CLAUDE.md, a team file that is committed \
     to the repository and read by everyone who works on it; AGENTS.md stays as it is.";

/// The two moves this engine makes. `Local`: the text of `AGENTS.local.md` goes into the personal
/// file `CLAUDE.local.md`. `Guidance`: the text a person wrote in `AGENTS.md` goes into the team
/// file `CLAUDE.md`.
#[derive(Clone, Copy, PartialEq, Eq, Debug)]
pub enum Kind {
    Local,
    Guidance,
}

impl Kind {
    /// The project-relative file the text is written into.
    fn target(self) -> &'static str {
        match self {
            Kind::Local => TARGET,
            Kind::Guidance => GUIDANCE_TARGET,
        }
    }

    /// The project-relative file the text is read from.
    fn source_name(self) -> &'static str {
        match self {
            Kind::Local => LOCAL,
            Kind::Guidance => LEGACY_TARGET,
        }
    }

    fn begin(self) -> &'static str {
        match self {
            Kind::Local => BEGIN,
            Kind::Guidance => GUIDANCE_BEGIN,
        }
    }

    fn end(self) -> &'static str {
        match self {
            Kind::Local => END,
            Kind::Guidance => GUIDANCE_END,
        }
    }

    /// How a block of this kind is named in a message.
    fn label(self) -> &'static str {
        match self {
            Kind::Local => "merged-local",
            Kind::Guidance => "merged-agents",
        }
    }

    fn receipt_prefix(self) -> &'static str {
        match self {
            Kind::Local => RECEIPT_PREFIX,
            Kind::Guidance => GUIDANCE_PREFIX,
        }
    }

    fn commit_notice(self) -> &'static str {
        match self {
            Kind::Local => COMMIT_NOTICE,
            Kind::Guidance => GUIDANCE_COMMIT_NOTICE,
        }
    }
}

/// Where the text to merge was read.
struct Source {
    /// `in_place` or `backup`.
    kind: &'static str,
    /// Project-relative path of the file read.
    rel: String,
    /// The whole file read (the receipt keeps a copy of it).
    bytes: Vec<u8>,
    /// The part that is merged: the whole file for `Kind::Local`, the person's text of
    /// `AGENTS.md` for `Kind::Guidance`. The block's sha256 names these bytes.
    text: Vec<u8>,
    /// What the preview says about text that is not carried over.
    notes: Vec<String>,
}

impl Source {
    fn whole(kind: &'static str, rel: String, bytes: Vec<u8>) -> Source {
        Source {
            kind,
            rel,
            text: bytes.clone(),
            bytes,
            notes: Vec::new(),
        }
    }
}

enum Found {
    Source(Source),
    /// Nothing to merge, with the reason.
    Nothing(String),
    Blocked(String),
}

struct Decision {
    action: &'static str,
    reason: String,
    new_agents: Option<Vec<u8>>,
    starts_at_line: Option<usize>,
    body: Option<String>,
}

impl Decision {
    fn of(action: &'static str, reason: impl Into<String>) -> Decision {
        Decision {
            action,
            reason: reason.into(),
            new_agents: None,
            starts_at_line: None,
            body: None,
        }
    }
}

/// The text with exactly one trailing newline (a source without one gets it).
fn normalized(text: &str) -> String {
    format!("{}\n", text.trim_end_matches('\n'))
}

fn hex(bytes: &[u8]) -> String {
    crate::digest(bytes)
        .trim_start_matches("sha256:")
        .to_string()
}

fn block_of_for(kind: Kind, sha_hex: &str, body: &str) -> String {
    format!(
        "{}{sha_hex}{BEGIN_TAIL}\n{body}{}\n",
        kind.begin(),
        kind.end()
    )
}

#[cfg(test)]
fn block_of(sha_hex: &str, body: &str) -> String {
    block_of_for(Kind::Local, sha_hex, body)
}

/// The block already in the target: the sha256 its marker names and the text between the marks.
struct Existing {
    sha: String,
    body: String,
}

/// `Ok(None)`: no block. `Err`: a block that cannot be read as one (the plan is a conflict).
fn find_block_for(kind: Kind, agents: &str) -> std::result::Result<Option<Existing>, String> {
    let (begin, end_marker, label) = (kind.begin(), kind.end(), kind.label());
    let starts: Vec<usize> = agents.match_indices(begin).map(|(i, _)| i).collect();
    let start = match starts.as_slice() {
        // An end marker alone means the start marker was deleted: the text may still be
        // there, and merging again would double it (#2982 audit N5).
        [] if agents.contains(end_marker) => {
            return Err(format!("the {label} end marker has no start marker"))
        }
        [] => return Ok(None),
        [one] => *one,
        _ => {
            return Err(format!(
                "{} holds more than one {label} block",
                kind.target()
            ))
        }
    };
    if start != 0 && !agents[..start].ends_with('\n') {
        return Err(format!("the {label} marker is not at the start of a line"));
    }
    let after = &agents[start + begin.len()..];
    let damaged = || format!("the {label} marker is damaged");
    let close = after.find(BEGIN_TAIL).ok_or_else(damaged)?;
    let sha = &after[..close];
    if sha.len() != 64 || !sha.bytes().all(|b| b.is_ascii_hexdigit()) {
        return Err(damaged());
    }
    let rest = after[close + BEGIN_TAIL.len()..]
        .strip_prefix('\n')
        .ok_or_else(damaged)?;
    if rest.matches(end_marker).count() != 1 {
        return Err(format!("the {label} block needs exactly one end marker"));
    }
    let end = rest.find(end_marker).ok_or_else(damaged)?;
    let body = &rest[..end];
    if !(body.is_empty() || body.ends_with('\n')) {
        return Err(format!(
            "the {label} end marker is not at the start of a line"
        ));
    }
    Ok(Some(Existing {
        sha: sha.to_ascii_lowercase(),
        body: body.to_string(),
    }))
}

#[cfg(test)]
fn find_block(agents: &str) -> std::result::Result<Option<Existing>, String> {
    find_block_for(Kind::Local, agents)
}

/// Lines of `text` shaped like a credential: `(line number, kind)`. The value is never kept.
fn secret_suspects(text: &str) -> Vec<(usize, &'static str)> {
    text.lines()
        .enumerate()
        .filter_map(|(i, line)| secret_kind(line).map(|kind| (i + 1, kind)))
        .collect()
}

/// A token of `prefix` followed by at least `min` token characters, the prefix not inside a word.
fn has_token(line: &str, prefix: &str, min: usize, upper_digit_only: bool) -> bool {
    line.match_indices(prefix).any(|(i, _)| {
        let word_before = line[..i]
            .chars()
            .next_back()
            .is_some_and(|c| c.is_ascii_alphanumeric());
        let run = line[i + prefix.len()..]
            .chars()
            .take_while(|c| {
                if upper_digit_only {
                    c.is_ascii_uppercase() || c.is_ascii_digit()
                } else {
                    c.is_ascii_alphanumeric() || *c == '-' || *c == '_'
                }
            })
            // Counting past `min` is never needed; without this one long line of repeated
            // prefixes costs time quadratic in its length (#2982 re-audit N6').
            .take(min)
            .count();
        !word_before && run >= min
    })
}

pub(crate) fn secret_kind(line: &str) -> Option<&'static str> {
    let lower = line.to_ascii_lowercase();
    if lower.contains("-----begin") && lower.contains("private key") {
        return Some("private_key");
    }
    for (prefix, min, upper, kind) in [
        ("sk-", 20, false, "api_key"),
        ("ghp_", 20, false, "github_token"),
        ("gho_", 20, false, "github_token"),
        ("github_pat_", 20, false, "github_token"),
        ("xoxb-", 10, false, "slack_token"),
        ("xoxp-", 10, false, "slack_token"),
        ("ghs_", 20, false, "github_token"),
        ("ghu_", 20, false, "github_token"),
        ("glpat-", 20, false, "gitlab_token"),
        ("sk_live_", 16, false, "api_key"),
        ("rk_live_", 16, false, "api_key"),
        ("xoxa-", 10, false, "slack_token"),
        ("xoxr-", 10, false, "slack_token"),
        ("xoxs-", 10, false, "slack_token"),
        ("xapp-", 10, false, "slack_token"),
        ("AIza", 35, false, "api_key"),
        ("hf_", 30, false, "api_key"),
        ("AKIA", 16, true, "aws_key"),
        ("ASIA", 16, true, "aws_key"),
    ] {
        if has_token(line, prefix, min, upper) {
            return Some(kind);
        }
    }
    // A JSON web token: a header and a payload, both starting `eyJ`.
    if line.contains(".eyJ") && has_token(line, "eyJ", 10, false) {
        return Some("jwt");
    }
    // `Authorization: Bearer <token>` or `Basic <base64>`.
    if lower.contains("authorization") {
        for word in ["bearer ", "basic "] {
            for (i, _) in lower.match_indices(word) {
                let run = line[i + word.len()..]
                    .chars()
                    .take(64)
                    .take_while(|c| c.is_ascii_alphanumeric() || "-._~+/=".contains(*c))
                    .count();
                if run >= 16 {
                    return Some("authorization_header");
                }
            }
        }
    }
    // A URL that carries a password: `scheme://user:password@host`.
    for (i, _) in line.match_indices("://") {
        // The host part ends at the next `/`, which the next `://` also holds, so the parts of
        // one line never overlap and the scan stays linear.
        let authority: String = line[i + 3..]
            .chars()
            .take_while(|c| *c != '/' && !c.is_whitespace())
            .collect();
        if let Some((userinfo, _)) = authority.split_once('@') {
            if let Some((_, password)) = userinfo.split_once(':') {
                if password.len() >= 3 && !password.starts_with(['<', '$', '{', '*', '[', '(']) {
                    return Some("url_credential");
                }
            }
        }
    }
    for key in [
        "password",
        "passwd",
        "secret",
        "token",
        "api_key",
        "apikey",
        "api-key",
        "secret_key",
        "secret_access_key",
        "access_key",
        "private_key",
    ] {
        // Every hit looks at a bounded number of value characters, so one huge line costs time
        // linear in its length, and a credential after many look-alikes is still found
        // (#2982 audit N6, re-audit N6'').
        for (i, _) in lower.match_indices(key) {
            let tail = lower[i + key.len()..].trim_start_matches(['"', '\'', '`', ' ', '\t']);
            let Some(tail) = tail.strip_prefix('=').or_else(|| tail.strip_prefix(':')) else {
                continue;
            };
            let at = line.len() - tail.len();
            let value: String = line[at..]
                .trim_start_matches([' ', '\t', '"', '\'', '`'])
                .chars()
                .take(256)
                .take_while(|c| !c.is_whitespace() && !matches!(c, '"' | '\'' | '`'))
                .collect();
            let looks_real = value.len() >= 8
                && value.is_ascii()
                && !value.starts_with(['<', '$', '{', '*', '[', '('])
                && value
                    .bytes()
                    .any(|b| b.is_ascii_digit() || b.is_ascii_uppercase());
            if looks_real {
                return Some("credential_assignment");
            }
        }
    }
    None
}

/// `true` when `text` imports `AGENTS.md` the way Claude Code reads an import: a word `@AGENTS.md`
/// (or `@./AGENTS.md`) outside a fenced code block and outside backticks.
fn imports_agents_md(text: &str) -> bool {
    let mut fenced = false;
    for line in text.lines() {
        let t = line.trim_start();
        if t.starts_with("```") || t.starts_with("~~~") {
            fenced = !fenced;
            continue;
        }
        if !fenced
            && line
                .split_whitespace()
                .any(|w| w == "@AGENTS.md" || w == "@./AGENTS.md")
        {
            return true;
        }
    }
    false
}

/// The plan's decision from the two files and the credential scan. Pure.
fn decide_for(
    kind: Kind,
    agents: Option<&[u8]>,
    found: Found,
    suspects: &[(usize, &'static str)],
) -> Decision {
    let target = kind.target();
    let src = match found {
        Found::Nothing(reason) => return Decision::of("none", reason),
        Found::Blocked(reason) => return Decision::of("blocked", reason),
        Found::Source(s) => s,
    };
    let Ok(text) = std::str::from_utf8(&src.text) else {
        return Decision::of("blocked", format!("{} is not UTF-8 text", src.rel));
    };
    // The guidance kind also refuses the personal-file markers: a damaged one that stayed in the
    // text would be copied into a file that another merge reads.
    let foreign = kind == Kind::Guidance && (text.contains(BEGIN) || text.contains(END));
    if text.contains(kind.begin())
        || text.contains(kind.end())
        || text.contains(MANAGED_MARKER)
        || foreign
    {
        return Decision::of(
            "blocked",
            format!(
                "{} holds a MoAI marker line; merged, it would be read as a managed block",
                src.rel
            ),
        );
    }
    // A missing target is created: it holds the block alone.
    let creates = agents.is_none();
    let Ok(agents) = std::str::from_utf8(agents.unwrap_or_default()) else {
        return Decision::of("blocked", format!("{target} is not UTF-8 text"));
    };
    if kind == Kind::Guidance && imports_agents_md(agents) {
        return Decision::of(
            "unchanged",
            format!("{target} already reads {LEGACY_TARGET} through an @{LEGACY_TARGET} import"),
        );
    }
    let body = normalized(text);
    let sha = hex(&src.text);
    match find_block_for(kind, agents) {
        Err(reason) => Decision::of("conflict", format!("{reason}; nothing is written")),
        Ok(Some(e)) if e.sha == sha && e.body == body => {
            Decision::of("unchanged", format!("{target} already holds this text"))
        }
        Ok(Some(e)) if e.sha == sha => Decision::of(
            "conflict",
            "the merged block was edited by hand; it is kept and nothing is written",
        ),
        Ok(Some(_)) => Decision::of(
            "conflict",
            format!(
                "{} changed since it was merged; the block is kept and nothing is written",
                src.rel
            ),
        ),
        Ok(None) if !suspects.is_empty() => Decision::of(
            "blocked",
            format!(
                "{} line(s) of {} look like a credential (lines {}); remove them from the file \
                 and preview again",
                suspects.len(),
                src.rel,
                suspects
                    .iter()
                    .map(|(n, _)| n.to_string())
                    .collect::<Vec<_>>()
                    .join(", ")
            ),
        ),
        Ok(None) => {
            let separator = if agents.is_empty() || agents.ends_with("\n\n") {
                ""
            } else if agents.ends_with('\n') {
                "\n"
            } else {
                "\n\n"
            };
            let head = format!("{agents}{separator}");
            let starts_at_line = head.matches('\n').count() + 1;
            let new_agents = format!("{head}{}", block_of_for(kind, &sha, &body)).into_bytes();
            // The plan, the revert and every later read of the target go through
            // `apply::read_opt`, which refuses a file above this size (#2982 audit B1).
            if new_agents.len() as u64 > crate::apply::RECEIPT_LIMIT {
                return Decision::of(
                    "blocked",
                    format!(
                        "{target} would grow to {} bytes with this text; the limit is {} bytes",
                        new_agents.len(),
                        crate::apply::RECEIPT_LIMIT
                    ),
                );
            }
            let from = kind.source_name();
            Decision {
                action: "merge",
                reason: if creates {
                    format!("{target} does not exist; the text of {from} would create it")
                } else {
                    format!("the text of {from} is not in {target} yet")
                },
                new_agents: Some(new_agents),
                starts_at_line: Some(starts_at_line),
                body: Some(body),
            }
        }
    }
}

#[cfg(test)]
fn decide(agents: Option<&[u8]>, found: Found, suspects: &[(usize, &'static str)]) -> Decision {
    decide_for(Kind::Local, agents, found, suspects)
}

/// What the personal file in place holds: text the user wrote, nothing, or MoAI's own text.
enum Content {
    Personal,
    Blank,
    Generated,
}

fn content_of(bytes: &[u8]) -> Content {
    if bytes.iter().all(u8::is_ascii_whitespace) {
        Content::Blank
    } else if crate::agents::is_generated_local(bytes) {
        Content::Generated
    } else {
        Content::Personal
    }
}

fn read_checked(root: &Path, rel: &str) -> std::result::Result<Option<Vec<u8>>, String> {
    let path = crate::apply::target(root, rel).map_err(|e| format!("{rel}: {e}"))?;
    let bytes = crate::apply::read_opt(&path).map_err(|e| format!("{rel}: {e}"))?;
    match bytes {
        Some(b) if b.len() > SOURCE_LIMIT => Err(format!(
            "{rel} is larger than {} KiB; not merged",
            SOURCE_LIMIT / 1024
        )),
        other => Ok(other),
    }
}

fn find_source(root: &Path) -> Result<Found> {
    let mut why_not = format!("no {LOCAL} and no backup copy of one");
    match read_checked(root, LOCAL) {
        Err(reason) => return Ok(Found::Blocked(reason)),
        Ok(None) => {}
        Ok(Some(bytes)) => match content_of(&bytes) {
            Content::Personal => {
                return Ok(Found::Source(Source::whole(
                    "in_place",
                    LOCAL.into(),
                    bytes,
                )))
            }
            Content::Blank => why_not = format!("{LOCAL} is empty"),
            Content::Generated => {
                why_not = format!("{LOCAL} only holds the text MoAI generated");
            }
        },
    }
    let backed = match crate::go_state::backed_up_root_file(root, LOCAL) {
        Ok(b) => b,
        Err(reason) => return Ok(Found::Blocked(reason)),
    };
    let Some(backed) = backed else {
        return Ok(Found::Nothing(why_not));
    };
    let bytes = match read_checked(root, &backed.rel) {
        Err(reason) => return Ok(Found::Blocked(reason)),
        Ok(None) => {
            return Ok(Found::Blocked(format!(
                "{} is missing; the go-state receipt names it",
                backed.rel
            )))
        }
        Ok(Some(b)) => b,
    };
    if crate::digest(&bytes) != backed.sha256 {
        return Ok(Found::Blocked(format!(
            "{} changed since the go-state apply moved it; not merged",
            backed.rel
        )));
    }
    Ok(match content_of(&bytes) {
        Content::Personal => Found::Source(Source::whole("backup", backed.rel, bytes)),
        Content::Blank => Found::Nothing(format!("the backup copy of {LOCAL} is empty")),
        Content::Generated => Found::Nothing(format!(
            "the backup copy of {LOCAL} only holds the text MoAI generated"
        )),
    })
}

/// Removes the one block that opens with `begin` at the start of a line and closes on the line
/// that holds `end`, and joins what was around it with one blank line. `None` when the text has
/// no such block, a second one, or an end marker that is not its own (the caller keeps the text
/// whole and the marker check blocks the plan).
fn remove_block(text: &str, begin: &str, end: &str) -> Option<(String, usize)> {
    let mut starts = text.match_indices(begin).map(|(i, _)| i);
    let start = starts.next()?;
    if starts.next().is_some() || text.matches(end).count() != 1 {
        return None;
    }
    if start != 0 && !text[..start].ends_with('\n') {
        return None;
    }
    let finish = start + text[start..].find(end)?;
    let after = finish + end.len();
    let stop = match text[after..].find('\n') {
        Some(n) => after + n + 1,
        None => text.len(),
    };
    let removed = &text[start..stop];
    let lines = removed.matches('\n').count() + usize::from(!removed.ends_with('\n'));
    let before = text[..start].trim_end_matches('\n');
    let tail = text[stop..].trim_start_matches('\n');
    let joined = match (before.is_empty(), tail.is_empty()) {
        (true, _) => tail.to_string(),
        (_, true) => format!("{before}\n"),
        _ => format!("{before}\n\n{tail}"),
    };
    Some((joined, lines))
}

/// A text that is nothing but one `# title` line: the heading MoAI puts above its managed section
/// when it creates an `AGENTS.md`.
fn is_title_only(text: &str) -> bool {
    let mut lines = text.lines().filter(|l| !l.trim().is_empty());
    matches!((lines.next(), lines.next()), (Some(first), None) if first.starts_with("# "))
}

/// The guidance kind's source: the text of `AGENTS.md` outside MoAI's managed section and outside
/// the blocks a personal-file merge wrote into it. Nothing is read from a backup.
fn find_guidance(root: &Path) -> Result<Found> {
    let bytes = match read_checked(root, LEGACY_TARGET) {
        Err(reason) => return Ok(Found::Blocked(reason)),
        Ok(None) => {
            return Ok(Found::Nothing(format!(
                "no {LEGACY_TARGET} in this project"
            )))
        }
        Ok(Some(b)) => b,
    };
    let Ok(whole) = std::str::from_utf8(&bytes) else {
        return Ok(Found::Blocked(format!("{LEGACY_TARGET} is not UTF-8 text")));
    };
    let mut own = whole.to_string();
    let mut notes = Vec::new();
    let (managed_begin, managed_end) =
        crate::agents::markers(crate::agents::AGENTS_SECTION, "html");
    if let Some((rest, lines)) = remove_block(&own, &managed_begin, &managed_end) {
        notes.push(format!(
            "MoAI's managed section of {LEGACY_TARGET} ({lines} lines) is not copied; the \
             generation writes it into {GUIDANCE_TARGET}. Text you wrote inside that section is \
             not carried over."
        ));
        own = rest;
    }
    if let Some((rest, lines)) = remove_block(&own, BEGIN, END) {
        notes.push(format!(
            "A block an earlier personal-file merge put into {LEGACY_TARGET} ({lines} lines) is \
             not copied; it is personal text and belongs in CLAUDE.local.md."
        ));
        own = rest;
    }
    if own.trim().is_empty() || is_title_only(&own) {
        return Ok(Found::Nothing(format!(
            "{LEGACY_TARGET} holds no text of its own (only MoAI's section and a title line)"
        )));
    }
    Ok(Found::Source(Source {
        kind: "in_place",
        rel: LEGACY_TARGET.into(),
        text: normalized(&own).into_bytes(),
        bytes,
        notes,
    }))
}

fn find_source_for(kind: Kind, root: &Path) -> Result<Found> {
    match kind {
        Kind::Local => find_source(root),
        Kind::Guidance => find_guidance(root),
    }
}

/// What a project's `AGENTS.md` holds, for the guard that keeps a first `CLAUDE.md` or
/// `CLAUDE.local.md` from hiding it (Claude Code reads `AGENTS.md` only while neither exists).
#[derive(Debug)]
pub enum AgentsMdState {
    /// No `AGENTS.md`.
    Absent,
    /// Only MoAI's own section, with at most the one `# title` line MoAI writes above it.
    OnlyMoai,
    /// Text the person wrote: the migration has something to carry over.
    HasText,
    /// A link, a file that is not UTF-8 text or is too large, a credential-shaped line or a stray
    /// MoAI marker. The reason never carries a value.
    Blocked(String),
}

/// Reads `AGENTS.md` the way the guidance migration does and says what it holds. Writes nothing.
pub fn agents_md_state(root: &Path) -> AgentsMdState {
    if fs::symlink_metadata(root.join(LEGACY_TARGET)).is_err() {
        return AgentsMdState::Absent;
    }
    let found = match find_guidance(root) {
        Ok(found) => found,
        Err(reason) => return AgentsMdState::Blocked(reason),
    };
    let suspects = match &found {
        Found::Source(s) => std::str::from_utf8(&s.bytes)
            .map(secret_suspects)
            .unwrap_or_default(),
        _ => Vec::new(),
    };
    // A missing `CLAUDE.md` is the case that matters: the decision is what a create would do.
    let decision = decide_for(Kind::Guidance, None, found, &suspects);
    match decision.action {
        "merge" => AgentsMdState::HasText,
        "none" => AgentsMdState::OnlyMoai,
        _ => AgentsMdState::Blocked(decision.reason),
    }
}

/// The notice and the warning for a personal file git already tracks: it is committed whatever
/// `.gitignore` says, and this move does not untrack it.
const TRACKED_NOTICE: &str =
    "This text goes into CLAUDE.local.md, which git already tracks, so it \
     will be committed with your next commit; this move does not untrack it. AGENTS.local.md \
     stays as it is.";
const TRACKED_WARNING: &str = "CLAUDE.local.md is already tracked by git, so this text will be \
     committed and shared; MoAI does not untrack it. Run git rm --cached CLAUDE.local.md yourself \
     if you want it out of the repository.";

/// The one line the personal file adds to `.gitignore`, and the file it goes into.
const IGNORE_LINE: &str = "CLAUDE.local.md";
const GITIGNORE: &str = ".gitignore";

/// The `.gitignore` write a plan carries: the bytes the file holds now (`None`: it does not
/// exist) and the bytes after the one line is added.
struct IgnoreWrite {
    before: Option<Vec<u8>>,
    new_bytes: Vec<u8>,
}

struct Inner {
    json: Value,
    action: &'static str,
    reason: String,
    before: Option<Vec<u8>>,
    new_agents: Option<Vec<u8>>,
    source: Option<Source>,
    ignore: Option<IgnoreWrite>,
}

/// Claude Code reads `AGENTS.md` and `.claude/AGENTS.md` only while none of `CLAUDE.md`,
/// `.claude/CLAUDE.md` and `CLAUDE.local.md` exists (the memory page, "When Claude Code reads
/// AGENTS.md"). The warning applies when this merge would create the first of them over such a
/// file; it names the official setting that keeps both read.
fn hidden_agents_md_warning(root: &Path, action: &str, target_exists: bool) -> Vec<String> {
    let exists = |rel: &str| fs::symlink_metadata(root.join(rel)).is_ok();
    let hidden: Vec<&str> = [LEGACY_TARGET, ".claude/AGENTS.md"]
        .into_iter()
        .filter(|f| exists(f))
        .collect();
    if action == "merge"
        && !target_exists
        && !hidden.is_empty()
        && !exists("CLAUDE.md")
        && !exists(".claude/CLAUDE.md")
    {
        let files = hidden.join(" and ");
        return vec![format!(
            "Claude Code reads {files} only while no CLAUDE.md, .claude/CLAUDE.md or \
             CLAUDE.local.md exists; creating {TARGET} makes it stop reading {files} in this \
             project. Move that text into a CLAUDE.md first, or confirm that {files} will no \
             longer be read. To keep both read, set Project instructions to \
             claude-md-and-agents-md in /config (Claude Code's own setting)."
        )];
    }
    Vec::new()
}

/// The exit code of `git -C <root> <args>`, or `None` when git cannot run. Nothing is printed.
fn git_code(root: &Path, args: &[&str]) -> Option<i32> {
    use std::process::{Command, Stdio};
    let mut cmd = Command::new("git");
    crate::identity::scrub_git_env(&mut cmd);
    cmd.arg("-C")
        .arg(root)
        .args(args)
        .stdin(Stdio::null())
        .stdout(Stdio::null())
        .stderr(Stdio::null())
        .status()
        .ok()
        .and_then(|s| s.code())
}

/// What git says about the personal file, and the `.gitignore` write that follows from it.
struct IgnoreFacts {
    json: Value,
    write: Option<IgnoreWrite>,
    /// Git already tracks `CLAUDE.local.md`: it is committed whatever `.gitignore` says.
    tracked: bool,
    /// A sentence for the plan's warnings, when the file is neither ignored nor made ignored.
    warning: Option<String>,
}

fn plan_ignore(root: &Path) -> IgnoreFacts {
    let line = IGNORE_LINE;
    let skip = |why: &str, exists: bool, warning: Option<String>, tracked: bool| IgnoreFacts {
        json: json!({
            "path": GITIGNORE, "line": line, "action": "skip", "exists": exists, "note": why,
        }),
        write: None,
        tracked,
        warning,
    };
    match git_code(root, &["rev-parse", "--is-inside-work-tree"]) {
        None => {
            return skip(
                "git is not available, so .gitignore was not checked",
                false,
                None,
                false,
            )
        }
        Some(0) => {}
        Some(_) => {
            return skip(
                "this folder is not a git repository, so there is no .gitignore line to add",
                false,
                None,
                false,
            )
        }
    }
    let tracked = git_code(root, &["ls-files", "--error-unmatch", "--", line]) == Some(0);
    let ignored = git_code(root, &["check-ignore", "-q", "--", line]) == Some(0);
    let located = crate::apply::target(root, GITIGNORE).and_then(|p| crate::apply::read_opt(&p));
    let existing = match located {
        Ok(bytes) => bytes,
        Err(e) => {
            let why = format!("{GITIGNORE} cannot be changed ({e})");
            let warning = (!ignored).then(|| {
                format!("{line} is not ignored by git and {GITIGNORE} cannot be changed here; add the line yourself so the file is not committed")
            });
            return skip(&why, true, warning, tracked);
        }
    };
    let exists = existing.is_some();
    if ignored {
        return IgnoreFacts {
            json: json!({"path": GITIGNORE, "line": line, "action": "present", "exists": exists, "note": null}),
            write: None,
            tracked,
            warning: None,
        };
    }
    let old = existing.clone().unwrap_or_default();
    let eol = if old.windows(2).any(|w| w == b"\r\n") {
        "\r\n"
    } else {
        "\n"
    };
    let separator = if old.is_empty() || old.ends_with(b"\n") {
        ""
    } else {
        eol
    };
    let mut new_bytes = old;
    new_bytes.extend_from_slice(format!("{separator}{line}{eol}").as_bytes());
    if new_bytes.len() as u64 > crate::apply::RECEIPT_LIMIT {
        return skip(&format!("{GITIGNORE} is too large to change here"), exists, Some(format!("{line} is not ignored by git and {GITIGNORE} is too large to change here; add the line yourself so the file is not committed")), tracked);
    }
    IgnoreFacts {
        json: json!({"path": GITIGNORE, "line": line, "action": "add", "exists": exists, "note": null}),
        write: Some(IgnoreWrite {
            before: existing,
            new_bytes,
        }),
        tracked,
        warning: None,
    }
}

/// `true` when `bytes` holds a CRLF line ending.
fn has_crlf(bytes: &[u8]) -> bool {
    bytes.windows(2).any(|w| w == b"\r\n")
}

/// A text moved by an earlier generation into `AGENTS.md`: the same bytes in a `moai:merged-local`
/// block there. Merging them again would put the personal text in two files.
fn merged_into_agents_md(root: &Path, source: &Source) -> Option<String> {
    let bytes = read_checked(root, LEGACY_TARGET).ok()??;
    let agents = std::str::from_utf8(&bytes).ok()?;
    let text = std::str::from_utf8(&source.text).ok()?;
    let existing = find_block_for(Kind::Local, agents).ok()??;
    (existing.sha == hex(&source.text) && existing.body == normalized(text)).then(|| {
        format!(
            "the text of {LOCAL} is already in {LEGACY_TARGET}, in a block an earlier generation \
             merged there; moving that block belongs to the {GUIDANCE_TARGET} migration"
        )
    })
}

fn plan_inner(kind: Kind, root: &Path) -> Result<Inner> {
    let target = kind.target();
    // A target that cannot be read as a plain file (a link, a path through one) blocks the plan.
    let located = crate::apply::target(root, target).and_then(|p| crate::apply::read_opt(&p));
    let (before, target_error) = match located {
        Ok(bytes) => (bytes, None),
        Err(e) => (None, Some(format!("{target}: {e}"))),
    };
    let found = find_source_for(kind, root)?;
    let suspects = match &found {
        Found::Source(s) => std::str::from_utf8(&s.bytes)
            .map(secret_suspects)
            .unwrap_or_default(),
        _ => Vec::new(),
    };
    let mut decision = match target_error {
        Some(reason) => Decision::of("blocked", reason),
        None => decide_for(kind, before.as_deref(), clone_found(&found), &suspects),
    };
    // Text an earlier generation merged into `AGENTS.md` is not merged a second time.
    if kind == Kind::Local && decision.action == "merge" {
        if let Found::Source(s) = &found {
            if let Some(reason) = merged_into_agents_md(root, s) {
                decision = Decision::of("unchanged", reason);
            }
        }
    }
    // Creating the personal file over an `AGENTS.md` hides that file from Claude Code; the
    // guidance kind moves the text of `AGENTS.md` itself, so it has no such warning. The apply
    // refuses a plan with this warning unless the caller passes the token derived from it.
    let hidden = if kind == Kind::Local {
        hidden_agents_md_warning(root, decision.action, before.is_some())
    } else {
        Vec::new()
    };
    let confirm_token =
        (!hidden.is_empty()).then(|| hex(hidden.join("\n").as_bytes())[..16].to_string());
    let mut warnings = hidden;
    // What git says about the personal file: ignored or made ignored by one `.gitignore` line,
    // or tracked (committed whatever `.gitignore` says).
    let ignore = (kind == Kind::Local && decision.action == "merge").then(|| plan_ignore(root));
    let mut notice = kind.commit_notice();
    if let Some(facts) = &ignore {
        if facts.tracked {
            notice = TRACKED_NOTICE;
            warnings.push(TRACKED_WARNING.to_string());
        }
        warnings.extend(facts.warning.iter().cloned());
    }
    if decision.action == "merge" {
        if let Found::Source(s) = &found {
            warnings.extend(s.notes.iter().cloned());
        }
    }
    let crlf: Vec<String> = [
        before
            .as_deref()
            .filter(|b| has_crlf(b))
            .map(|_| target.to_string()),
        match &found {
            Found::Source(s) if has_crlf(&s.text) => Some(s.rel.clone()),
            _ => None,
        },
    ]
    .into_iter()
    .flatten()
    .collect();
    let line_endings = (decision.action == "merge" && !crlf.is_empty()).then(|| {
        format!(
            "{} use CRLF line endings; the marker lines are written with LF and the text keeps its own",
            crlf.join(" and ")
        )
    });
    let source = match found {
        Found::Source(s) => Some(s),
        _ => None,
    };
    let source_json = source.as_ref().map(|s| {
        json!({
            "kind": s.kind,
            "path": s.rel,
            "sha256": crate::digest(&s.bytes),
            "bytes": s.bytes.len(),
        })
    });
    let suspects_json: Vec<Value> = suspects
        .iter()
        .map(|(line, kind)| json!({"line": line, "kind": kind}))
        .collect();
    let ignore_json = ignore.as_ref().map(|f| f.json.clone());
    let ignore_before = ignore
        .as_ref()
        .and_then(|f| f.write.as_ref())
        .map(|w| w.before.as_deref().map(crate::digest));
    let bound = json!({
        "action": decision.action,
        "source": source_json,
        "agents_sha256": before.as_deref().map(crate::digest),
        "secret_suspects": suspects_json,
        "warnings": warnings,
        "gitignore": ignore_json,
        "gitignore_sha256": ignore_before,
        "confirm_token": confirm_token,
    });
    let digest = crate::digest(&serde_json::to_vec(&bound).map_err(|_| "encode failed")?);
    let mut json = json!({
        "project": root.display().to_string(),
        "action": decision.action,
        "reason": decision.reason,
        "source": source_json,
        "target": {
            "path": target,
            "exists": before.is_some(),
            "sha256_before": before.as_deref().map(crate::digest),
            "starts_at_line": decision.starts_at_line,
        },
        "gitignore": ignore_json,
        "commit_notice": notice,
        "warnings": warnings,
        "needs_confirmation": confirm_token.is_some(),
        "confirm_token": confirm_token,
        "line_endings": line_endings,
        "secret_suspects": suspects_json,
        "body": decision.body,
        "plan_digest": digest,
    });
    json["next_actions"] = if decision.action != "merge" {
        json!([])
    } else if kind == Kind::Local {
        json!([
            "show the user the body, the position, the commit_notice and any warnings; after \
             their yes run \
             `moai agents-merge-apply --plan-digest <plan_digest> \
             --permission-reference <ref> <project>`; when needs_confirmation is true, add \
             `--confirm-warnings <confirm_token>` only after the user has read the warning and \
             said yes to it"
        ])
    } else {
        // No command carries this kind yet; the plan is applied by its digest.
        json!([
            "show the user the body, the position, the commit_notice and any warnings; apply this \
             plan with its plan_digest only after their yes"
        ])
    };
    Ok(Inner {
        json,
        action: decision.action,
        reason: decision.reason,
        before,
        new_agents: decision.new_agents,
        source,
        ignore: ignore.and_then(|f| f.write),
    })
}

/// `decide` consumes its `Found`; the plan keeps the source for the receipt, so it decides on
/// a copy.
fn clone_found(found: &Found) -> Found {
    match found {
        Found::Source(s) => Found::Source(Source {
            kind: s.kind,
            rel: s.rel.clone(),
            bytes: s.bytes.clone(),
            text: s.text.clone(),
            notes: s.notes.clone(),
        }),
        Found::Nothing(r) => Found::Nothing(r.clone()),
        Found::Blocked(r) => Found::Blocked(r.clone()),
    }
}

/// The read-only preview of one kind. Writes nothing.
pub fn plan_for(kind: Kind, project: &Path) -> Result<Value> {
    let root = fs::canonicalize(project).map_err(|_| crate::contract::PROJECT_PATH_UNAVAILABLE)?;
    plan_inner(kind, &root).map(|i| i.json)
}

/// The read-only preview of the personal-file move. Writes nothing.
pub fn plan(project: &Path) -> Result<Value> {
    plan_for(Kind::Local, project)
}

#[derive(Serialize, Deserialize)]
struct Receipt {
    schema_version: u32,
    id: String,
    at: String,
    plan_digest: String,
    permission_reference: String,
    /// `pending`, `applied`, `failed` or `reverted`.
    status: String,
    agents_before: String,
    agents_after: Option<String>,
    block_sha256: String,
    source_kind: String,
    source_path: String,
    source_sha256: String,
    /// The project-relative file the apply wrote. A receipt of the earlier generation has no
    /// such field and names `AGENTS.md`.
    #[serde(default = "legacy_target")]
    target_path: String,
    /// False when the apply created the file; the revert then removes it.
    #[serde(default = "target_existed_default")]
    target_existed: bool,
    /// The `.gitignore` the apply wrote the one line into; absent when it wrote none.
    #[serde(default)]
    gitignore_path: Option<String>,
    /// False when the apply created `.gitignore`; the revert then removes it.
    #[serde(default = "target_existed_default")]
    gitignore_existed: bool,
    #[serde(default)]
    gitignore_before: Option<String>,
    #[serde(default)]
    gitignore_after: Option<String>,
}

fn legacy_target() -> String {
    LEGACY_TARGET.to_string()
}

fn target_existed_default() -> bool {
    true
}

fn receipts_dir(root: &Path) -> std::path::PathBuf {
    crate::state_root::path(root, "receipts")
}

/// Owner-only bits for a receipt folder and for what it holds: the receipt keeps a copy of a
/// personal file, which may hold a token (#2982 audit B2; the same bits as `go_state`'s backup).
const PRIVATE_DIR: u32 = 0o700;
const PRIVATE_FILE: u32 = 0o600;

/// Create one folder (not its parents) that only the owner can enter, from its creation.
fn create_private_dir(path: &Path) -> std::io::Result<()> {
    let mut builder = fs::DirBuilder::new();
    #[cfg(unix)]
    {
        use std::os::unix::fs::DirBuilderExt;
        builder.mode(PRIVATE_DIR);
    }
    builder.create(path)
}

/// Create `path` and the folders below `base` on the way to it, each one private. A link on
/// the way is refused.
fn create_private_dirs(base: &Path, rel: &[&str]) -> Result<std::path::PathBuf> {
    let mut cur = base.to_path_buf();
    for part in rel {
        cur.push(part);
        match fs::symlink_metadata(&cur) {
            Ok(m) if m.file_type().is_symlink() => return Err("symlink target refused".into()),
            Ok(m) if m.is_dir() => {}
            Ok(_) => return Err("receipt folder unavailable".into()),
            Err(_) => create_private_dir(&cur).map_err(|_| "receipt folder unavailable")?,
        }
    }
    Ok(cur)
}

/// A new file in a receipt folder: owner-only from its creation, never replacing a file.
fn write_private(path: &Path, bytes: &[u8]) -> Result<()> {
    crate::apply::publish_exclusive_mode(path, bytes, Some(PRIVATE_FILE))
        .map_err(|_| "receipt write failed".into())
}

fn new_receipt_dir(kind: Kind, root: &Path) -> Result<(String, std::path::PathBuf)> {
    let state = crate::state_root::path(root, "");
    crate::claude::refuse_symlink_dir(&state)?;
    fs::create_dir_all(&state).map_err(|_| "receipt folder unavailable")?;
    let receipts = create_private_dirs(&state, &["receipts"])?;
    let nanos = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_nanos())
        .unwrap_or(0);
    for n in 0..16 {
        let id = format!("{}{nanos}-{n}", kind.receipt_prefix());
        let dir = receipts.join(&id);
        if create_private_dir(&dir).is_ok() {
            return Ok((id, dir));
        }
    }
    Err("receipt id unavailable".into())
}

fn write_receipt(dir: &Path, r: &Receipt) -> Result<()> {
    let mut bytes = serde_json::to_vec_pretty(r).map_err(|_| "encode failed")?;
    bytes.push(b'\n');
    let path = dir.join("receipt.json");
    match crate::claude::read_existing(&path)? {
        // The first write creates the file owner-only; a later one keeps the bits it has.
        None => write_private(&path, &bytes),
        Some(before) => crate::apply::publish_replace(&path, Some(&before), &bytes),
    }
}

/// Merges the text of the personal file after the user confirmed the plan whose digest is
/// `plan_digest`.
pub fn apply(project: &Path, plan_digest: &str, permission: &str) -> Result<Value> {
    apply_for(Kind::Local, project, plan_digest, permission)
}

/// Puts a file this module wrote back: its earlier bytes, or no file when there were none. The
/// file must still hold `current`.
fn restore_file(path: &Path, current: &[u8], earlier: Option<&[u8]>) -> Result<()> {
    match earlier {
        Some(bytes) => crate::apply::publish_replace(path, Some(current), bytes),
        None => fs::remove_file(path).map_err(|_| "the created file could not be removed".into()),
    }
}

/// Merges the text of one kind after the user confirmed the plan whose digest is `plan_digest`.
/// A plan that carries the warning about `AGENTS.md` is refused here; `apply_confirmed` takes the
/// token that answers it.
pub fn apply_for(kind: Kind, project: &Path, plan_digest: &str, permission: &str) -> Result<Value> {
    apply_confirmed(kind, project, plan_digest, permission, None)
}

/// `apply_for` with the confirmation token of the plan: the user's yes to the warning that Claude
/// Code would stop reading `AGENTS.md`.
pub fn apply_confirmed(
    kind: Kind,
    project: &Path,
    plan_digest: &str,
    permission: &str,
    confirmation: Option<&str>,
) -> Result<Value> {
    let target = kind.target();
    if !crate::contract::permission_reference_ok(permission) {
        return Err("a permission reference is required; nothing was written".into());
    }
    let root = fs::canonicalize(project).map_err(|_| crate::contract::PROJECT_PATH_UNAVAILABLE)?;
    let inner = plan_inner(kind, &root)?;
    if inner.json["plan_digest"].as_str() != Some(plan_digest) {
        return Err("stale or altered plan; re-preview required".into());
    }
    match inner.action {
        "merge" => {}
        "none" | "unchanged" => return Ok(json!({"state": "unchanged", "writes": []})),
        other => return Err(format!("the plan is {other}: {}", inner.reason)),
    }
    if let Some(token) = inner.json["confirm_token"].as_str() {
        if confirmation != Some(token) {
            return Err(
                "the plan carries a warning that needs your explicit confirmation: read \
                        it, then pass --confirm-warnings with the confirmation token printed \
                        with the plan; nothing was written"
                    .into(),
            );
        }
    }
    let (Some(new_agents), Some(source)) = (inner.new_agents, inner.source) else {
        return Err("the plan holds no merge".into());
    };
    let ignore = inner.ignore;
    let ignore_path = match &ignore {
        Some(_) => Some(crate::apply::target(&root, GITIGNORE)?),
        None => None,
    };
    // `None`: the target does not exist and the apply creates it.
    let before = inner.before;
    if new_agents.len() as u64 > crate::apply::RECEIPT_LIMIT {
        return Err(format!(
            "the merged {target} would be larger than the limit; nothing was written"
        ));
    }
    let agents_path = crate::apply::target(&root, target)?;
    let (id, dir) = new_receipt_dir(kind, &root)?;
    let keep = create_private_dirs(&dir, &["before"])?;
    if let Some(before) = &before {
        write_private(&keep.join(target), before)?;
    }
    if let Some(Some(earlier)) = ignore.as_ref().map(|g| g.before.as_ref()) {
        write_private(&keep.join(GITIGNORE), earlier)?;
    }
    // A file in place stays; the receipt keeps the copy that serves as its backup. A moved
    // copy is already in its dated go-state backup.
    if source.kind == "in_place" {
        let copy = create_private_dirs(&dir, &["source"])?;
        write_private(&copy.join(kind.source_name()), &source.bytes)?;
    }
    let mut receipt = Receipt {
        schema_version: 1,
        id: id.clone(),
        at: crate::baseline::utc_now(),
        plan_digest: plan_digest.to_string(),
        permission_reference: permission.to_string(),
        status: "pending".into(),
        agents_before: crate::digest(before.as_deref().unwrap_or_default()),
        // Known before the write, so a revert can tell an interrupted apply that did write
        // from one that did not (#2982 audit N2).
        agents_after: Some(crate::digest(&new_agents)),
        block_sha256: hex(&source.text),
        source_kind: source.kind.into(),
        source_path: source.rel.clone(),
        source_sha256: crate::digest(&source.bytes),
        target_path: target.to_string(),
        target_existed: before.is_some(),
        gitignore_path: ignore.as_ref().map(|_| GITIGNORE.to_string()),
        gitignore_existed: ignore.as_ref().is_none_or(|g| g.before.is_some()),
        gitignore_before: ignore
            .as_ref()
            .and_then(|g| g.before.as_deref().map(crate::digest)),
        gitignore_after: ignore.as_ref().map(|g| crate::digest(&g.new_bytes)),
    };
    write_receipt(&dir, &receipt)?;
    // `.gitignore` first: when it cannot be written nothing else has changed.
    if let (Some(g), Some(path)) = (&ignore, &ignore_path) {
        if let Err(e) = crate::apply::publish_replace(path, g.before.as_deref(), &g.new_bytes) {
            receipt.status = "failed".into();
            receipt.agents_after = None;
            receipt.gitignore_after = None;
            write_receipt(&dir, &receipt)?;
            return Err(format!("{GITIGNORE} was not written ({e})"));
        }
    }
    if let Err(e) = crate::apply::publish_replace(&agents_path, before.as_deref(), &new_agents) {
        if let (Some(g), Some(path)) = (&ignore, &ignore_path) {
            // Best effort: the line goes back so the project is as it was.
            let _ = restore_file(path, &g.new_bytes, g.before.as_deref());
        }
        receipt.status = "failed".into();
        receipt.agents_after = None;
        receipt.gitignore_after = None;
        write_receipt(&dir, &receipt)?;
        return Err(format!("{target} was not written ({e})"));
    }
    receipt.status = "applied".into();
    write_receipt(&dir, &receipt)?;
    let undo = if kind == Kind::Local {
        format!(
            "to undo: moai agents-merge-revert --receipt {id} --permission-reference <ref> <project>"
        )
    } else {
        format!("to undo: revert receipt {id} with a permission reference")
    };
    Ok(json!({
        "state": "applied",
        "project": root.display().to_string(),
        "source": {"kind": source.kind, "path": source.rel},
        "starts_at_line": inner.json["target"]["starts_at_line"],
        "receipt": id,
        "writes": if ignore.is_some() { json!([GITIGNORE, target]) } else { json!([target]) },
        "next_actions": [undo],
    }))
}

/// Undoes one merge while the file it wrote still holds the bytes the apply wrote: the file
/// gets its earlier bytes back, or is removed when the apply created it. The source file and its
/// backup are not touched.
pub fn revert(project: &Path, receipt_id: &str, permission: &str) -> Result<Value> {
    revert_for(Kind::Local, project, receipt_id, permission)
}

/// Puts the `.gitignore` the apply changed back, only while it still holds the bytes the apply
/// wrote. `None`: the receipt wrote none. `"reverted"`, `"unchanged"` (it is already as it was) or
/// `"kept"` (it changed since, so it is left alone).
fn revert_gitignore(root: &Path, dir: &Path, r: &Receipt) -> Result<Option<&'static str>> {
    let Some(rel) = r.gitignore_path.as_deref() else {
        return Ok(None);
    };
    if rel != GITIGNORE {
        return Err("the receipt names a file this command never writes".into());
    }
    let path = crate::apply::target(root, rel)?;
    let now = crate::apply::read_opt(&path)?;
    let now_digest = now.as_deref().map(crate::digest);
    let as_before = if r.gitignore_existed {
        now_digest.as_deref() == r.gitignore_before.as_deref()
    } else {
        now.is_none()
    };
    if as_before {
        return Ok(Some("unchanged"));
    }
    let (Some(now), Some(after)) = (now, r.gitignore_after.as_deref()) else {
        return Ok(Some("kept"));
    };
    if crate::digest(&now) != after {
        return Ok(Some("kept"));
    }
    if r.gitignore_existed {
        crate::claude::refuse_symlink_dir(&dir.join("before"))?;
        let earlier = crate::claude::read_existing(&dir.join("before").join(GITIGNORE))?
            .ok_or_else(|| format!("the receipt has no copy of the earlier {GITIGNORE}"))?;
        if Some(crate::digest(&earlier).as_str()) != r.gitignore_before.as_deref() {
            return Err(format!(
                "the receipt's copy of the earlier {GITIGNORE} was altered; nothing was written"
            ));
        }
        restore_file(&path, &now, Some(&earlier))?;
    } else {
        restore_file(&path, &now, None)?;
    }
    Ok(Some("reverted"))
}

/// `revert` for a receipt of `kind`; a receipt id of the other kind is unknown to it.
pub fn revert_for(kind: Kind, project: &Path, receipt_id: &str, permission: &str) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err("a permission reference is required; nothing was written".into());
    }
    if !(receipt_id.starts_with(kind.receipt_prefix())
        && receipt_id.len() <= 64
        && receipt_id
            .bytes()
            .all(|b| b.is_ascii_alphanumeric() || b == b'-'))
    {
        return Err("unknown receipt id".into());
    }
    let root = fs::canonicalize(project).map_err(|_| "project directory unavailable")?;
    let dir = receipts_dir(&root).join(receipt_id);
    for d in [
        crate::state_root::path(&root, ""),
        receipts_dir(&root),
        dir.clone(),
    ] {
        crate::claude::refuse_symlink_dir(&d)?;
    }
    let bytes =
        crate::claude::read_existing(&dir.join("receipt.json"))?.ok_or("receipt not found")?;
    let mut r: Receipt = serde_json::from_slice(&bytes).map_err(|_| "receipt is unreadable")?;
    // The receipt names the file it wrote; only the files this kind ever wrote are valid. A
    // personal-file receipt of the earlier generation names `AGENTS.md`.
    let name: &str = match (kind, r.target_path.as_str()) {
        (Kind::Local, TARGET) => TARGET,
        (Kind::Local, LEGACY_TARGET) => LEGACY_TARGET,
        (Kind::Guidance, GUIDANCE_TARGET) => GUIDANCE_TARGET,
        _ => return Err("the receipt names a file this command never writes".into()),
    };
    let agents_path = crate::apply::target(&root, name)?;
    match r.status.as_str() {
        "applied" => {}
        "reverted" => return Ok(json!({"state": "unchanged", "writes": []})),
        // An apply that was interrupted: the file holds either the bytes it was about to write
        // (undo like an applied one) or its earlier state (it wrote nothing).
        "pending" => {
            let now = crate::apply::read_opt(&agents_path)?;
            let wrote_nothing = match &now {
                None => !r.target_existed,
                Some(b) => r.target_existed && crate::digest(b) == r.agents_before,
            };
            if wrote_nothing {
                // `.gitignore` is written first, so an interrupted apply may have changed it.
                let gitignore = revert_gitignore(&root, &dir, &r)?;
                r.status = "failed".into();
                r.agents_after = None;
                write_receipt(&dir, &r)?;
                let writes = if gitignore == Some("reverted") {
                    json!([GITIGNORE])
                } else {
                    json!([])
                };
                return Ok(json!({"state": "unchanged", "writes": writes, "gitignore": gitignore}));
            }
            // Any other bytes fall to the check below: the file changed; nothing is written.
        }
        _ => return Err("that apply wrote nothing; there is nothing to undo".into()),
    }
    let now = crate::apply::read_opt(&agents_path)?.ok_or_else(|| format!("{name} is missing"))?;
    if r.agents_after.as_deref() != Some(crate::digest(&now).as_str()) {
        return Err(format!(
            "{name} changed since the merge; nothing was written"
        ));
    }
    // The copy of the earlier file is checked before anything is written; `.gitignore` goes back
    // first (it is the one that may stay as it is), then the file itself.
    let gitignore;
    if r.target_existed {
        crate::claude::refuse_symlink_dir(&dir.join("before"))?;
        let before = crate::claude::read_existing(&dir.join("before").join(name))?
            .ok_or_else(|| format!("the receipt has no copy of the earlier {name}"))?;
        if crate::digest(&before) != r.agents_before {
            return Err(format!(
                "the receipt's copy of the earlier {name} was altered; nothing was written"
            ));
        }
        gitignore = revert_gitignore(&root, &dir, &r)?;
        crate::apply::publish_replace(&agents_path, Some(&now), &before)?;
    } else {
        gitignore = revert_gitignore(&root, &dir, &r)?;
        // The apply created the file and it still holds exactly those bytes: remove it, so the
        // project is as it was (an empty `CLAUDE.local.md` would still count for Claude Code).
        fs::remove_file(&agents_path).map_err(|_| format!("{name} could not be removed"))?;
    }
    r.status = "reverted".into();
    write_receipt(&dir, &r)?;
    let writes = if gitignore == Some("reverted") {
        json!([GITIGNORE, name])
    } else {
        json!([name])
    };
    Ok(json!({
        "state": "reverted",
        "project": root.display().to_string(),
        "receipt": receipt_id,
        "writes": writes,
        "gitignore": gitignore,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    fn src(text: &str) -> Found {
        Found::Source(Source::whole(
            "in_place",
            LOCAL.into(),
            text.as_bytes().to_vec(),
        ))
    }

    #[test]
    fn remove_block_joins_with_one_blank_line_and_refuses_odd_markers() {
        let (b, e) = ("<!-- b -->", "<!-- /b -->");
        let text = format!("one\n\n{b}\nx\n{e}\n\ntwo\n");
        assert_eq!(remove_block(&text, b, e), Some(("one\n\ntwo\n".into(), 3)));
        // Only the block: nothing is left to join.
        assert_eq!(
            remove_block(&format!("{b}\nx\n{e}\n"), b, e),
            Some((String::new(), 3))
        );
        // Text before the block only, and text after it only.
        assert_eq!(
            remove_block(&format!("one\n{b}\nx\n{e}\n"), b, e),
            Some(("one\n".into(), 3))
        );
        assert_eq!(
            remove_block(&format!("{b}\nx\n{e}\ntwo\n"), b, e),
            Some(("two\n".into(), 3))
        );
        // No block, two blocks, an end marker alone, a start that is not at a line start.
        assert_eq!(remove_block("one\n", b, e), None);
        assert_eq!(remove_block(&format!("{b}\n{e}\n{b}\n{e}\n"), b, e), None);
        assert_eq!(remove_block(&format!("one\n{e}\n"), b, e), None);
        assert_eq!(remove_block(&format!("a {b}\nx\n{e}\n"), b, e), None);
        // An end marker before the start marker is not a block.
        assert_eq!(remove_block(&format!("{e}\n{b}\n"), b, e), None);
    }

    #[test]
    fn a_title_alone_is_the_heading_moai_wrote() {
        assert!(is_title_only("# my-project\n"));
        assert!(is_title_only("\n# my-project\n\n"));
        assert!(!is_title_only("# my-project\n\nmore\n"));
        assert!(!is_title_only("## sub\n"));
        assert!(!is_title_only("plain\n"));
        assert!(!is_title_only(""));
    }

    #[test]
    fn an_import_counts_only_outside_code() {
        assert!(imports_agents_md("# Team\n\n@AGENTS.md\n"));
        assert!(imports_agents_md("See @./AGENTS.md for the rules.\n"));
        assert!(!imports_agents_md("```\n@AGENTS.md\n```\n"));
        assert!(!imports_agents_md("~~~\n@AGENTS.md\n~~~\n"));
        assert!(!imports_agents_md("Use `@AGENTS.md` to import.\n"));
        assert!(!imports_agents_md("@AGENTS.md.bak\n"));
        // A fence that closes ends the code.
        assert!(imports_agents_md("```\ncode\n```\n@AGENTS.md\n"));
    }

    fn action(agents: Option<&str>, found: Found) -> &'static str {
        let suspects = match &found {
            Found::Source(s) => secret_suspects(std::str::from_utf8(&s.bytes).unwrap()),
            _ => Vec::new(),
        };
        decide(agents.map(str::as_bytes), found, &suspects).action
    }

    /// Credential-shaped text is built here, not written out, so no fixture in the source is
    /// itself a secret.
    fn shaped(prefix: &str, run: usize) -> String {
        format!("{prefix}{}", "a".repeat(run))
    }

    #[test]
    fn the_text_lands_after_a_blank_line_at_the_end_with_its_hash() {
        let d = decide(Some(b"# Rules\n"), src("mine"), &[]);
        assert_eq!(d.action, "merge");
        let sha = hex(b"mine");
        assert_eq!(
            String::from_utf8(d.new_agents.unwrap()).unwrap(),
            format!("# Rules\n\n{BEGIN}{sha}{BEGIN_TAIL}\nmine\n{END}\n")
        );
        assert_eq!(d.starts_at_line, Some(3));
    }

    #[test]
    fn the_separator_adapts_to_how_agents_md_ends() {
        let at = |agents: &str| decide(Some(agents.as_bytes()), src("x\n"), &[]).starts_at_line;
        assert_eq!(at(""), Some(1));
        assert_eq!(at("a\n\n"), Some(3));
        assert_eq!(at("a\n"), Some(3));
        assert_eq!(at("a"), Some(3));
    }

    #[test]
    fn a_second_run_is_unchanged_and_an_edit_is_a_conflict() {
        let first = decide(Some(b"# Rules\n"), src("mine"), &[]);
        let merged = String::from_utf8(first.new_agents.unwrap()).unwrap();
        assert_eq!(action(Some(&merged), src("mine")), "unchanged");
        // A source that gained a trailing newline has the same normalized body but a new hash.
        assert_eq!(action(Some(&merged), src("mine\n")), "conflict");
        assert_eq!(action(Some(&merged), src("changed")), "conflict");
        let edited = merged.replace("mine", "edited");
        assert_eq!(action(Some(&edited), src("mine")), "conflict");
    }

    #[test]
    fn a_damaged_or_doubled_block_is_a_conflict() {
        let sha = hex(b"mine");
        let good = format!("{BEGIN}{sha}{BEGIN_TAIL}\nmine\n{END}\n");
        assert_eq!(action(Some(&good), src("mine")), "unchanged");
        assert_eq!(
            action(Some(&format!("{good}{good}")), src("mine")),
            "conflict"
        );
        let short = good.replace(&sha, "abc");
        assert_eq!(action(Some(&short), src("mine")), "conflict");
        let no_end = good.replace(END, "");
        assert_eq!(action(Some(&no_end), src("mine")), "conflict");
        let mid_line = format!("text {good}");
        assert_eq!(action(Some(&mid_line), src("mine")), "conflict");
        // Only the start marker deleted: merging again would double the text.
        let no_start: String = good.lines().skip(1).map(|l| format!("{l}\n")).collect();
        assert_eq!(action(Some(&no_start), src("mine")), "conflict");
    }

    /// The merged `AGENTS.md` must stay readable by `apply::read_opt` (#2982 audit B1).
    #[test]
    fn the_merged_file_may_not_pass_the_receipt_limit() {
        let limit = crate::apply::RECEIPT_LIMIT as usize;
        let block_len = block_of(&hex(b"x"), "x\n").len();
        let agents_of = |extra: usize| format!("{}\n\n", "a".repeat(limit - 2 - block_len + extra));
        let merged_len = |extra: usize| {
            decide(Some(agents_of(extra).as_bytes()), src("x"), &[])
                .new_agents
                .map(|b| b.len())
        };
        assert_eq!(merged_len(0), Some(limit));
        let over = decide(Some(agents_of(1).as_bytes()), src("x"), &[]);
        assert_eq!(over.action, "blocked");
        assert!(over.new_agents.is_none());
        assert!(
            over.reason.contains(&(limit + 1).to_string()),
            "{}",
            over.reason
        );
    }

    #[test]
    fn nothing_to_merge_and_waiting_cases() {
        assert_eq!(action(Some("a\n"), Found::Nothing("none".into())), "none");
        assert_eq!(action(Some("a\n"), Found::Blocked("why".into())), "blocked");
        assert_eq!(action(Some("a\n"), src("a\u{0}b")), "merge");
    }

    /// A missing target is created: the new file holds the block alone, from line 1.
    #[test]
    fn a_missing_target_is_created_with_the_block_alone() {
        let d = decide(None, src("mine"), &[]);
        assert_eq!(d.action, "merge");
        assert_eq!(d.starts_at_line, Some(1));
        let sha = hex(b"mine");
        assert_eq!(
            String::from_utf8(d.new_agents.unwrap()).unwrap(),
            format!("{BEGIN}{sha}{BEGIN_TAIL}\nmine\n{END}\n")
        );
        // Nothing to merge still means nothing, with or without the target.
        assert_eq!(action(None, Found::Nothing("none".into())), "none");
    }

    #[test]
    fn a_source_with_a_moai_marker_is_blocked() {
        for marker in [
            "<!-- moai:managed:begin x -->".to_string(),
            END.to_string(),
            format!("{BEGIN}{}{BEGIN_TAIL}", hex(b"x")),
        ] {
            assert_eq!(
                action(Some("a\n"), src(&format!("mine\n{marker}\n"))),
                "blocked"
            );
        }
    }

    #[test]
    fn credential_shaped_lines_block_the_merge() {
        let key = format!("key: {}", shaped("sk-", 24));
        assert_eq!(action(Some("a\n"), src(&key)), "blocked");
        let assignment = format!("a\npassword = {}", "Hunter2".repeat(2));
        assert_eq!(action(Some("a\n"), src(&assignment)), "blocked");
        let pem = format!("note\n{}{} RSA PRIVATE KEY{}", "-----", "BEGIN", "-----");
        assert_eq!(action(Some("a\n"), src(&pem)), "blocked");
        // Shapes that are not credentials: a short value, a placeholder, a prefix inside a word.
        assert_eq!(action(Some("a\n"), src("password: short")), "merge");
        assert_eq!(
            action(Some("a\n"), src("token: <your-token-here>")),
            "merge"
        );
        assert_eq!(action(Some("a\n"), src(&shaped("task-", 26))), "merge");
        // An existing block is judged without the scan: the text is already in the file.
        let sha = hex(key.as_bytes());
        let have = format!("{BEGIN}{sha}{BEGIN_TAIL}\n{key}\n{END}\n");
        assert_eq!(action(Some(&have), src(&key)), "unchanged");
    }

    #[test]
    fn suspects_report_the_line_and_kind_only() {
        let text = format!(
            "ok\n{}\n{}\n",
            shaped("ghp_", 30),
            "AKIA".to_string() + &"A".repeat(16)
        );
        assert_eq!(
            secret_suspects(&text),
            vec![(2, "github_token"), (3, "aws_key")]
        );
        assert_eq!(secret_suspects(&shaped("sk-", 19)), vec![]);
        assert_eq!(secret_suspects(&shaped("sk-", 20)), vec![(1, "api_key")]);
    }

    #[test]
    fn token_shapes_have_an_exact_minimum_length() {
        // (prefix, shortest run that counts, kind); one fewer is not a token.
        for (prefix, min, kind) in [
            ("AKIA", 16, "aws_key"),
            ("ASIA", 16, "aws_key"),
            ("ghs_", 20, "github_token"),
            ("ghu_", 20, "github_token"),
            ("glpat-", 20, "gitlab_token"),
            ("sk_live_", 16, "api_key"),
            ("xapp-", 10, "slack_token"),
            ("xoxa-", 10, "slack_token"),
            ("AIza", 35, "api_key"),
            ("hf_", 30, "api_key"),
        ] {
            let body = |n: usize| {
                if prefix.starts_with("AKIA") || prefix.starts_with("ASIA") {
                    format!("{prefix}{}", "A".repeat(n))
                } else {
                    shaped(prefix, n)
                }
            };
            assert_eq!(secret_suspects(&body(min)), vec![(1, kind)], "{prefix}");
            assert_eq!(secret_suspects(&body(min - 1)), vec![], "{prefix} short");
        }
    }

    #[test]
    fn headers_urls_and_web_tokens_are_credentials() {
        let bearer = |n: usize| format!("Authorization: {} {}", "Bearer", "a".repeat(n));
        assert_eq!(
            secret_suspects(&bearer(16)),
            vec![(1, "authorization_header")]
        );
        assert_eq!(secret_suspects(&bearer(15)), vec![]);
        // The word alone, without the header name, is prose.
        assert_eq!(
            secret_suspects(&format!("Bearer {}", "a".repeat(30))),
            vec![]
        );

        let url = |password: &str| format!("DB: {}://app:{password}@db.example/main", "postgres");
        assert_eq!(secret_suspects(&url("pw3")), vec![(1, "url_credential")]);
        assert_eq!(secret_suspects(&url("pw")), vec![]);
        assert_eq!(secret_suspects(&url("<password>")), vec![]);
        assert_eq!(
            secret_suspects("see https://example.com/a:b@c for details"),
            vec![]
        );

        let jwt = format!("{}.{}.sig", shaped("eyJ", 12), shaped("eyJ", 12));
        assert_eq!(secret_suspects(&jwt), vec![(1, "jwt")]);
        assert_eq!(secret_suspects(&shaped("eyJ", 12)), vec![]);
    }

    #[test]
    fn assignments_need_a_value_of_eight_characters_with_a_digit_or_capital() {
        for key in [
            "password",
            "SECRET_KEY",
            "AWS_SECRET_ACCESS_KEY",
            "private_key",
        ] {
            assert_eq!(
                secret_suspects(&format!("{key}=Abcdef12")),
                vec![(1, "credential_assignment")],
                "{key}"
            );
            assert_eq!(
                secret_suspects(&format!("{key}=Abcdef1")),
                vec![],
                "{key} short"
            );
        }
        assert_eq!(secret_suspects("password: abcdefghij"), vec![]);
    }

    /// A single huge line must not cost time quadratic in its length, whatever shape it repeats
    /// (#2982 audit N6 and re-audit N6'; each of these took 30 seconds or more in a debug build
    /// before its scan was bounded).
    #[test]
    fn one_huge_line_is_scanned_in_bounded_time() {
        let long = |unit: &str| unit.repeat(SOURCE_LIMIT / unit.len());
        let mut lines = vec![
            long("password:"),
            long("xsk-"),
            long("aghp_"),
            long("AKIA"),
            long("ghp_"),
            long("hf_"),
            long("://"),
            long("a://u:"),
            long("authorization bearer "),
            long("secret_key="),
            format!("password{}x", " ".repeat(SOURCE_LIMIT - 10)),
            format!("authorization: bearer {}", "a".repeat(SOURCE_LIMIT - 30)),
            format!("{}{}", "sk-", "a".repeat(SOURCE_LIMIT - 3)),
        ];
        lines.push(long("eyJ.eyJ"));
        for line in lines {
            let start = std::time::Instant::now();
            let _ = secret_suspects(&line);
            assert!(
                start.elapsed().as_secs() < 3,
                "{:?} for a line starting {:?}",
                start.elapsed(),
                &line[..12]
            );
        }
    }

    /// Look-alikes before a real credential on the same line must not hide it (re-audit N6'').
    #[test]
    fn a_credential_after_many_lookalikes_is_still_found() {
        let lookalikes = "password:x ".repeat(200);
        let real = format!("{lookalikes}password: Abcdef12");
        assert_eq!(secret_suspects(&real), vec![(1, "credential_assignment")]);
        let urls = "http://a/b ".repeat(200);
        let real = format!("{urls}{}://u:pw3@h", "postgres");
        assert_eq!(secret_suspects(&real), vec![(1, "url_credential")]);
        assert_eq!(secret_suspects(&lookalikes), vec![]);
        // The header words too (#2994 audit B1): look-alikes first, a real token last.
        for (word, token) in [
            ("bearer", "abcdefghijklmnopqrstuv"),
            ("basic", "QWxhZGRpbjpvcGVuIHNlc2FtZQ=="),
        ] {
            let alikes = format!("authorization: {}", format!("{word} x ").repeat(200));
            assert_eq!(secret_suspects(&alikes), vec![], "{word} look-alikes only");
            let real = format!("{alikes}{word} {token}");
            assert_eq!(
                secret_suspects(&real),
                vec![(1, "authorization_header")],
                "{word}"
            );
        }
    }
}
