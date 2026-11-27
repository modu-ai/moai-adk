//! The release audit summary (SPEC-20261125-release-audit-summary): one comment on a release pull
//! request that lists, for every first-parent commit of `base..head`, the verdict the ledger
//! holds. This file is the grammar the command that builds the comment and the `audit-gate` job
//! that reads it share: [`render`] writes it, [`parse`] reads it, and both live here so the two
//! sides cannot drift apart.
//!
//! ```text
//! RELEASE-AUDIT-SUMMARY head=<40hex> base=<40hex> landings=<N>
//! S s1 <session id>                      (one line per session, numbered from 1)
//! L <sha40> p=<first parent 12hex> PASS v=<card>:<head12>:<tier>:<sN>:<sN.sN|->:<event> [v=...]
//! L <sha40> p=<12hex> ABS <target sha12>
//! L <sha40> p=<12hex> NONE c=<card,card|-> | <title, at most 40 bytes>
//! L <sha40> p=<12hex> FAIL why=<code> c=<card,card|-> | <title>
//!
//! <!-- moai-release-audit-summary <head 40hex> -->
//! ```
//!
//! The first line never starts with `AUDIT #`, so the older reader of verdict comments never takes
//! the summary for one of its own lines. The body holds only hex, ids and numbers except the
//! cleaned title of a line without a passing verdict.

use crate::card::land::check;
use crate::forge::connector::{
    summary_comment_body, HostConnector, HostKind, ProjectRef, SummaryOutcome, SummaryTarget,
};
use crate::forge::public_text::clean_public_text;
use crate::ledger::{self, workflow, LedgerError};
use crate::trailers;
use rusqlite::Connection;
use serde_json::{json, Value};
use std::collections::{BTreeMap, HashMap, HashSet};
use std::path::Path;

/// The first word of the first line.
pub const HEADER: &str = "RELEASE-AUDIT-SUMMARY";

/// The most bytes the posted comment may have. Beyond it the command refuses, it never cuts
/// (GitHub's own limit is 65536 characters; the official wording was not read).
pub const MAX_COMMENT_BYTES: usize = 60_000;

/// The most bytes of a line's title.
pub const TITLE_MAX_BYTES: usize = 40;

const MARKER_PREFIX: &str = "<!-- moai-release-audit-summary ";

/// What the comment says.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Summary {
    /// The pull request head, 40 lowercase hex.
    pub head: String,
    /// The merge base of the pull request, 40 lowercase hex.
    pub base: String,
    /// The session table; session `sN` is `sessions[N - 1]`.
    pub sessions: Vec<String>,
    /// One line per first-parent commit of `base..head`, oldest first.
    pub lines: Vec<Line>,
}

/// One first-parent commit.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Line {
    pub sha: String,
    /// The first 12 hex digits of the first parent.
    pub parent: String,
    pub kind: Kind,
}

/// What the ledger says of the commit.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum Kind {
    /// Every card the commit closed has a passing verdict.
    Pass(Vec<Verdict>),
    /// The commit belongs to a later landing (a lane that was fast-forwarded in): the first 12
    /// hex digits of that landing's commit.
    Abs { target: String },
    /// No passing verdict: a commit with no landing row, a card with no usable verdict, or a
    /// reviewer who is no independent one.
    None { cards: Vec<i64>, title: String },
    /// A FAIL verdict stands against the commit.
    Fail {
        why: String,
        cards: Vec<i64>,
        title: String,
    },
}

/// One card's passing verdict on a `PASS` line.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Verdict {
    pub card: i64,
    /// The first 12 hex digits of the audited head.
    pub head: String,
    pub tier: u8,
    /// Index into [`Summary::sessions`].
    pub reviewer: usize,
    /// Indexes into [`Summary::sessions`]: the sessions that wrote or held the card.
    pub excluded: Vec<usize>,
    /// The verdict event of the ledger.
    pub event: i64,
}

impl Verdict {
    /// Whether the audit gate counts this verdict (REQ-14): the list of sessions that wrote or
    /// held the card is not empty and the reviewer is none of them, compared with
    /// [`trailers::same_identity`].
    pub fn counts(&self, summary: &Summary) -> bool {
        let Some(reviewer) = summary.sessions.get(self.reviewer) else {
            return false;
        };
        !self.excluded.is_empty()
            && !self.excluded.iter().any(|&i| {
                summary
                    .sessions
                    .get(i)
                    .is_none_or(|other| trailers::same_identity(other, reviewer))
            })
    }
}

/// The one-line marker the posted comment ends with.
pub fn marker(head: &str) -> String {
    format!("{MARKER_PREFIX}{head} -->")
}

/// A title for a line: the first line of `raw` made safe to post ([`clean_public_text`]), cut to
/// [`TITLE_MAX_BYTES`] bytes on a character boundary; `-` when nothing is left.
pub fn clean_title(raw: &str) -> String {
    let first = raw.lines().next().unwrap_or("").trim();
    let (cleaned, _) = clean_public_text(first);
    let one_line = cleaned.replace('\n', " ");
    let mut end = one_line.len().min(TITLE_MAX_BYTES);
    while !one_line.is_char_boundary(end) {
        end -= 1;
    }
    let cut = one_line[..end].trim();
    if cut.is_empty() {
        "-".to_string()
    } else {
        cut.to_string()
    }
}

fn session_ref(index: usize) -> String {
    format!("s{}", index + 1)
}

fn cards_text(cards: &[i64]) -> String {
    if cards.is_empty() {
        "-".to_string()
    } else {
        cards
            .iter()
            .map(i64::to_string)
            .collect::<Vec<_>>()
            .join(",")
    }
}

/// The summary as text, without the marker.
pub fn render(summary: &Summary) -> String {
    let mut out = format!(
        "{HEADER} head={} base={} landings={}",
        summary.head,
        summary.base,
        summary.lines.len()
    );
    for (i, id) in summary.sessions.iter().enumerate() {
        out.push_str(&format!("\nS {} {id}", session_ref(i)));
    }
    for line in &summary.lines {
        out.push_str(&format!("\nL {} p={} ", line.sha, line.parent));
        match &line.kind {
            Kind::Pass(verdicts) => {
                out.push_str("PASS");
                for v in verdicts {
                    let excluded = if v.excluded.is_empty() {
                        "-".to_string()
                    } else {
                        v.excluded
                            .iter()
                            .map(|&i| session_ref(i))
                            .collect::<Vec<_>>()
                            .join(".")
                    };
                    out.push_str(&format!(
                        " v={}:{}:{}:{}:{excluded}:{}",
                        v.card,
                        v.head,
                        v.tier,
                        session_ref(v.reviewer),
                        v.event
                    ));
                }
            }
            Kind::Abs { target } => out.push_str(&format!("ABS {target}")),
            Kind::None { cards, title } => {
                out.push_str(&format!("NONE c={} | {title}", cards_text(cards)))
            }
            Kind::Fail { why, cards, title } => {
                out.push_str(&format!("FAIL why={why} c={} | {title}", cards_text(cards)))
            }
        }
    }
    out
}

/// Refuses a comment of `bytes` bytes when it is longer than [`MAX_COMMENT_BYTES`].
pub fn check_size(bytes: usize) -> Result<(), String> {
    if bytes > MAX_COMMENT_BYTES {
        Err(format!(
            "the comment would be {bytes} bytes, over the {MAX_COMMENT_BYTES} byte cap; refused, \
             not cut"
        ))
    } else {
        Ok(())
    }
}

/// The exact comment the connector posts: the cleaned body, a blank line and the marker
/// ([`summary_comment_body`]). A comment over the cap is refused.
pub fn comment(summary: &Summary) -> Result<String, String> {
    let text = summary_comment_body(&render(summary), &marker(&summary.head));
    check_size(text.len())?;
    Ok(text)
}

/// The digest of the comment's bytes: what a person approved is what is posted.
pub fn plan_digest(comment: &str) -> String {
    crate::digest(comment.as_bytes())
}

/// Reads a comment: with or without the trailing marker, with LF or CRLF line ends. A comment
/// that does not form one unbroken chain from `base` to `head` is refused (REQ-11), so the
/// caller has to compare only [`Summary::head`] and [`Summary::base`] with what it knows.
pub fn parse(text: &str) -> Result<Summary, String> {
    if text.len() > MAX_COMMENT_BYTES {
        return Err(format!(
            "a summary comment is at most {MAX_COMMENT_BYTES} bytes; this one has {}",
            text.len()
        ));
    }
    let text = text.replace("\r\n", "\n");
    let mut lines: Vec<&str> = text.split('\n').collect();
    while lines.last() == Some(&"") {
        lines.pop();
    }
    let marker_line = match lines.last() {
        Some(last) if last.starts_with(MARKER_PREFIX) => {
            let last = *last;
            lines.pop();
            if lines.pop() != Some("") {
                return Err("the marker line must follow one blank line".to_string());
            }
            Some(last)
        }
        _ => None,
    };
    let mut rest = lines.into_iter().enumerate();
    let (_, first) = rest.next().ok_or("the comment is empty")?;
    let (head, base, landings) = parse_header(first)?;
    if let Some(found) = marker_line {
        if found != marker(&head) {
            return Err("the marker names another head than the first line".to_string());
        }
    }
    let mut sessions: Vec<String> = Vec::new();
    let mut parsed: Vec<Line> = Vec::new();
    for (at, line) in rest {
        let at = at + 1;
        if let Some(entry) = line.strip_prefix("S ") {
            if !parsed.is_empty() {
                return Err(format!("line {at}: a session line after a landing line"));
            }
            let (tag, id) = entry
                .split_once(' ')
                .ok_or_else(|| format!("line {at}: a session line is `S sN <id>`"))?;
            if tag != session_ref(sessions.len()) {
                return Err(format!(
                    "line {at}: the sessions are numbered s1, s2, ... in order"
                ));
            }
            if id.is_empty() || !id.bytes().all(|b| b.is_ascii_graphic()) {
                return Err(format!(
                    "line {at}: a session id has no blank or control character"
                ));
            }
            if sessions.iter().any(|s| s.eq_ignore_ascii_case(id)) {
                return Err(format!("line {at}: a session is listed twice"));
            }
            sessions.push(id.to_string());
        } else if let Some(entry) = line.strip_prefix("L ") {
            parsed
                .push(parse_landing(entry, sessions.len()).map_err(|e| format!("line {at}: {e}"))?);
        } else {
            return Err(format!(
                "line {at}: not a session line, a landing line or blank"
            ));
        }
    }
    if parsed.is_empty() || parsed.len() != landings {
        return Err(format!(
            "the first line says landings={landings} but the comment has {} landing lines",
            parsed.len()
        ));
    }
    check_chain(&parsed, &base, &head)?;
    Ok(Summary {
        head,
        base,
        sessions,
        lines: parsed,
    })
}

fn is_hex(text: &str, len: usize) -> bool {
    text.len() == len
        && text
            .bytes()
            .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b))
}

/// A decimal number without sign or leading zero, at least `min`.
fn number(text: &str, min: u64) -> Result<u64, String> {
    let canonical = !text.is_empty()
        && text.bytes().all(|b| b.is_ascii_digit())
        && (text == "0" || !text.starts_with('0'));
    text.parse::<u64>()
        .ok()
        .filter(|n| canonical && *n >= min)
        .ok_or_else(|| format!("`{text}` is not a number of at least {min}"))
}

fn parse_header(line: &str) -> Result<(String, String, usize), String> {
    let tokens: Vec<&str> = line.split(' ').collect();
    let [name, head, base, landings] = tokens.as_slice() else {
        return Err(
            "the first line is `RELEASE-AUDIT-SUMMARY head=<40hex> base=<40hex> landings=<N>`"
                .into(),
        );
    };
    if *name != HEADER {
        return Err(format!("the first word is not {HEADER}"));
    }
    let field = |token: &str, key: &str| -> Result<String, String> {
        let value = token
            .strip_prefix(key)
            .ok_or_else(|| format!("the first line lacks `{key}`"))?;
        Ok(value.to_string())
    };
    let head = field(head, "head=")?;
    let base = field(base, "base=")?;
    if !is_hex(&head, 40) || !is_hex(&base, 40) {
        return Err("head and base are 40 lowercase hex digits".into());
    }
    let landings = number(&field(landings, "landings=")?, 1)?;
    Ok((head, base, landings as usize))
}

fn parse_session(token: &str, sessions: usize) -> Result<usize, String> {
    let n = token
        .strip_prefix('s')
        .ok_or_else(|| format!("`{token}` is not a session reference"))
        .and_then(|digits| number(digits, 1))?;
    if n as usize > sessions {
        return Err(format!("`{token}` names a session the table does not list"));
    }
    Ok(n as usize - 1)
}

fn parse_cards(text: &str) -> Result<Vec<i64>, String> {
    if text == "-" {
        return Ok(Vec::new());
    }
    text.split(',')
        .map(|c| number(c, 1).map(|n| n as i64))
        .collect()
}

fn parse_title(title: &str) -> Result<String, String> {
    if title.is_empty() || title.len() > TITLE_MAX_BYTES || title.chars().any(char::is_control) {
        return Err(format!(
            "a title is one line of 1 to {TITLE_MAX_BYTES} bytes without control characters"
        ));
    }
    Ok(title.to_string())
}

fn parse_verdict(token: &str, sessions: usize) -> Result<Verdict, String> {
    let body = token
        .strip_prefix("v=")
        .ok_or_else(|| format!("`{token}` is not a verdict (`v=card:head:tier:sN:sN.sN:event`)"))?;
    let parts: Vec<&str> = body.split(':').collect();
    let [card, head, tier, reviewer, excluded, event] = parts.as_slice() else {
        return Err(format!("`{token}` does not have six fields"));
    };
    if !is_hex(head, 12) {
        return Err(format!(
            "`{token}`: the audited head is 12 lowercase hex digits"
        ));
    }
    let tier = number(tier, 0)?;
    if tier > 2 {
        return Err(format!("`{token}`: the tier is 0, 1 or 2"));
    }
    let excluded = if *excluded == "-" {
        Vec::new()
    } else {
        excluded
            .split('.')
            .map(|s| parse_session(s, sessions))
            .collect::<Result<Vec<_>, _>>()?
    };
    Ok(Verdict {
        card: number(card, 1)? as i64,
        head: head.to_string(),
        tier: tier as u8,
        reviewer: parse_session(reviewer, sessions)?,
        excluded,
        event: number(event, 1)? as i64,
    })
}

fn parse_landing(entry: &str, sessions: usize) -> Result<Line, String> {
    let (sha, rest) = entry
        .split_once(' ')
        .ok_or("a landing line is `L <sha40> p=<12hex> <kind> ...`")?;
    if !is_hex(sha, 40) {
        return Err("the commit is 40 lowercase hex digits".into());
    }
    let (parent, rest) = rest
        .strip_prefix("p=")
        .and_then(|r| r.split_once(' '))
        .ok_or("the commit is followed by `p=<12hex>`")?;
    if !is_hex(parent, 12) {
        return Err("the parent is 12 lowercase hex digits".into());
    }
    let (word, tail) = match rest.split_once(' ') {
        Some((word, tail)) => (word, Some(tail)),
        None => (rest, None),
    };
    let kind = match (word, tail) {
        ("PASS", Some(tail)) => Kind::Pass(
            tail.split(' ')
                .map(|t| parse_verdict(t, sessions))
                .collect::<Result<Vec<_>, _>>()?,
        ),
        ("PASS", None) => return Err("a PASS line has at least one verdict".into()),
        ("ABS", Some(target)) if is_hex(target, 12) => Kind::Abs {
            target: target.to_string(),
        },
        ("NONE", Some(tail)) => {
            let (cards, title) = tail
                .split_once(" | ")
                .ok_or("a NONE line is `NONE c=<cards> | <title>`")?;
            Kind::None {
                cards: parse_cards(cards.strip_prefix("c=").ok_or("a NONE line has `c=`")?)?,
                title: parse_title(title)?,
            }
        }
        ("FAIL", Some(tail)) => {
            let (left, title) = tail
                .split_once(" | ")
                .ok_or("a FAIL line is `FAIL why=<code> c=<cards> | <title>`")?;
            let tokens: Vec<&str> = left.split(' ').collect();
            let [why, cards] = tokens.as_slice() else {
                return Err("a FAIL line has `why=` and `c=`".into());
            };
            let why = why.strip_prefix("why=").ok_or("a FAIL line has `why=`")?;
            if why.is_empty() || !why.bytes().all(|b| b.is_ascii_lowercase() || b == b'_') {
                return Err("a reason is lower-case letters and `_`".into());
            }
            Kind::Fail {
                why: why.to_string(),
                cards: parse_cards(cards.strip_prefix("c=").ok_or("a FAIL line has `c=`")?)?,
                title: parse_title(title)?,
            }
        }
        _ => return Err(format!("`{word}` is not a landing kind this reader knows")),
    };
    Ok(Line {
        sha: sha.to_string(),
        parent: parent.to_string(),
        kind,
    })
}

/// The lines form one unbroken first-parent chain from `base` to `head`, no commit twice, and
/// every `ABS` points at a later landing line that is no `ABS` itself.
fn check_chain(lines: &[Line], base: &str, head: &str) -> Result<(), String> {
    let mut previous = &base[..12];
    let mut seen = std::collections::HashSet::new();
    for (i, line) in lines.iter().enumerate() {
        if line.parent != previous {
            return Err(format!(
                "landing {} does not follow the one before it: its first parent is {}, not {previous}",
                i + 1,
                line.parent
            ));
        }
        if !seen.insert(line.sha.as_str()) {
            return Err(format!("landing {} repeats a commit", i + 1));
        }
        previous = &line.sha[..12];
    }
    if lines.last().map(|l| l.sha.as_str()) != Some(head) {
        return Err("the last landing is not the head".to_string());
    }
    for (i, line) in lines.iter().enumerate() {
        if let Kind::Abs { target } = &line.kind {
            let later = lines[i + 1..]
                .iter()
                .any(|l| l.sha[..12] == *target && !matches!(l.kind, Kind::Abs { .. }));
            if !later {
                return Err(format!(
                    "landing {} is absorbed by {target}, which is no later landing line",
                    i + 1
                ));
            }
        }
    }
    Ok(())
}

// ---- the builder -------------------------------------------------------------------------------

type Result<T, E = String> = std::result::Result<T, E>;

/// What the command reads: the project (git history and profile), the HOME that holds its
/// ledger, and the range.
pub struct Request<'a> {
    pub project: &'a Path,
    pub home: &'a Path,
    /// The merge base of the pull request, 40 lowercase hex.
    pub base: &'a str,
    /// The pull request head, 40 lowercase hex.
    pub head: &'a str,
}

/// The summary of a range and what the command shows of it.
pub struct Built {
    pub project_id: String,
    pub summary: Summary,
    /// The exact comment that would be posted, marker included.
    pub comment: String,
    pub digest: String,
    /// One entry per `NONE` or `FAIL` line.
    pub missing: Vec<Value>,
    /// `PASS`, `ABS`, `NONE` and `FAIL` lines.
    pub counts: [usize; 4],
}

impl Built {
    /// Whether every line has a passing verdict or is absorbed by one that has.
    pub fn complete(&self) -> bool {
        self.counts[2] + self.counts[3] == 0
    }

    fn json(&self, extra: Value) -> Value {
        let mut out = json!({
            "action": "release-audit-summary",
            "project_id": self.project_id,
            "base": self.summary.base,
            "head": self.summary.head,
            "landings": self.summary.lines.len(),
            "counts": {
                "pass": self.counts[0], "abs": self.counts[1],
                "none": self.counts[2], "fail": self.counts[3],
            },
            "complete": self.complete(),
            "missing": self.missing,
            "marker": marker(&self.summary.head),
            "text": self.comment,
            "bytes": self.comment.len(),
            "plan_digest": self.digest,
        });
        for (key, value) in extra.as_object().into_iter().flatten() {
            out[key] = value.clone();
        }
        out
    }
}

/// One first-parent commit of the range.
struct Entry {
    sha: String,
    parents: Vec<String>,
    subject: String,
}

/// One `card.landed` row of the ledger for a commit of the range.
struct Landed {
    /// The verdict event the landing relied on, when the row names one.
    relied: Option<i64>,
    /// The session that ran the landing.
    lander: String,
}

/// One card's passing verdict before the session table exists.
struct Draft {
    card: i64,
    head: String,
    tier: u8,
    reviewer: String,
    excluded: Vec<String>,
    event: i64,
}

/// What the land rules say of one card on one commit.
enum Judged {
    Pass(Draft),
    Fail { code: String, detail: String },
    None { code: String, detail: String },
}

fn none(code: &str, detail: impl Into<String>) -> Judged {
    Judged::None {
        code: code.to_string(),
        detail: detail.into(),
    }
}

fn check_commit(project: &Path, flag: &str, sha: &str) -> Result<()> {
    if !crate::gate::is_full_sha(sha) {
        return Err(format!("{flag} must be 40 lowercase hex"));
    }
    let spec = format!("{sha}^{{commit}}");
    let found = check::git_text(project, &["rev-parse", "--verify", "--quiet", &spec])
        .map_err(|_| format!("{flag} {sha} is not a commit of this repository"))?;
    if found.trim() != sha {
        return Err(format!("{flag} {sha} is not a commit of this repository"));
    }
    Ok(())
}

/// The first-parent commits of `base..head`, oldest first, as one unbroken chain from `base`.
fn read_entries(project: &Path, base: &str, head: &str) -> Result<Vec<Entry>> {
    let range = format!("{base}..{head}");
    let log = check::git_text(
        project,
        &[
            "log",
            "--first-parent",
            "--reverse",
            "--format=%H%x1f%P%x1f%s%x1e",
            &range,
        ],
    )?;
    let mut entries = Vec::new();
    for record in log.split('\u{1e}') {
        let record = record.trim_start_matches('\n');
        if record.trim().is_empty() {
            continue;
        }
        let mut fields = record.splitn(3, '\u{1f}');
        let (Some(sha), Some(parents), Some(subject)) =
            (fields.next(), fields.next(), fields.next())
        else {
            return Err("git log answered a record this command cannot read".to_string());
        };
        entries.push(Entry {
            sha: sha.to_string(),
            parents: parents.split_whitespace().map(str::to_string).collect(),
            subject: subject.to_string(),
        });
    }
    if entries.is_empty() {
        return Err(format!(
            "no first-parent commit between {} and {}: --base has to be an ancestor of --head",
            &base[..12],
            &head[..12]
        ));
    }
    let mut previous = base;
    for entry in &entries {
        if entry.parents.first().map(String::as_str) != Some(previous) {
            return Err(format!(
                "the first-parent line of --head does not pass through --base: the first parent \
                 of {} is not {}",
                &entry.sha[..12],
                &previous[..12]
            ));
        }
        previous = &entry.sha;
    }
    Ok(entries)
}

fn sql(e: rusqlite::Error) -> String {
    format!("the ledger could not be read: {e}")
}

/// The `card.landed` rows whose head is a commit of the range, by head and card (the newest row
/// of a card and head wins).
fn read_landed(
    conn: &Connection,
    in_range: &HashSet<&str>,
) -> Result<HashMap<String, BTreeMap<i64, Landed>>> {
    let mut st = conn
        .prepare(
            "SELECT actor, card_id, json_extract(payload, '$.head'), \
                    json_extract(payload, '$.verdict.event') \
             FROM event WHERE kind = 'card.landed' AND card_id IS NOT NULL \
               AND json_type(payload, '$.head') = 'text' ORDER BY event_id",
        )
        .map_err(sql)?;
    let rows = st
        .query_map([], |r| {
            Ok((
                r.get::<_, String>(0)?,
                r.get::<_, i64>(1)?,
                r.get::<_, String>(2)?,
                r.get::<_, Option<i64>>(3).unwrap_or(None),
            ))
        })
        .map_err(sql)?;
    let mut by_head: HashMap<String, BTreeMap<i64, Landed>> = HashMap::new();
    for row in rows {
        let (lander, card, head, relied) = row.map_err(sql)?;
        if in_range.contains(head.as_str()) {
            by_head
                .entry(head)
                .or_default()
                .insert(card, Landed { relied, lander });
        }
    }
    Ok(by_head)
}

/// The tip each landing started from, by card and head (the newest `card.land_started` row wins).
fn read_started(conn: &Connection) -> Result<HashMap<(i64, String), String>> {
    let mut st = conn
        .prepare(
            "SELECT card_id, json_extract(payload, '$.head'), json_extract(payload, '$.tip') \
             FROM event WHERE kind = 'card.land_started' AND card_id IS NOT NULL \
               AND json_type(payload, '$.head') = 'text' AND json_type(payload, '$.tip') = 'text' \
             ORDER BY event_id",
        )
        .map_err(sql)?;
    let rows = st
        .query_map([], |r| {
            Ok((
                r.get::<_, i64>(0)?,
                r.get::<_, String>(1)?,
                r.get::<_, String>(2)?,
            ))
        })
        .map_err(sql)?;
    let mut started = HashMap::new();
    for row in rows {
        let (card, head, tip) = row.map_err(sql)?;
        started.insert((card, head), tip);
    }
    Ok(started)
}

fn renderable(session: &str) -> bool {
    !session.is_empty() && session.bytes().all(|b| b.is_ascii_graphic())
}

/// Whether `head` is a head a landing of `entry` can have relied a verdict on: the landed commit
/// itself (a fast-forward), or the second parent of a two-parent commit that is a clean join.
fn relied_head_fits(project: &Path, entry: &Entry, head: &str) -> bool {
    head == entry.sha
        || (entry.parents.len() == 2
            && head == entry.parents[1]
            && check::is_clean_join(project, &entry.sha, &entry.parents[0], &entry.parents[1]))
}

/// The head the verdict of one card on one first-parent commit is read on. The head of the event
/// the landing row relied on counts only when [`relied_head_fits`]; otherwise, and when the row
/// names none, the verdict is read on the commit, else on the second parent of a clean join, and
/// last on the commit itself, so that [`check::judge`] answers `verdict_other_head` (fail closed).
fn verdict_head_of(
    project: &Path,
    entry: &Entry,
    events: &[workflow::VerdictEvent],
    row: &Landed,
) -> String {
    let second = entry.parents.get(1).filter(|_| entry.parents.len() == 2);
    row.relied
        .and_then(|id| events.iter().find(|e| e.event_id == id))
        .map(|e| e.head.clone())
        .filter(|head| relied_head_fits(project, entry, head))
        .or_else(|| {
            events
                .iter()
                .any(|e| e.head == entry.sha)
                .then(|| entry.sha.clone())
        })
        .or_else(|| {
            let second = second?;
            let on_second = events.iter().any(|e| &e.head == second);
            (on_second && check::is_clean_join(project, &entry.sha, &entry.parents[0], second))
                .then(|| second.clone())
        })
        .unwrap_or_else(|| entry.sha.clone())
}

/// Judges one card on one first-parent commit by the rules `card land` applies
/// ([`check::judge`], SPEC D-2). The verdict is read on the head [`verdict_head_of`] names;
/// `tip` is the commit the card's work was added to.
fn judge_card(
    project: &Path,
    conn: &Connection,
    entry: &Entry,
    card: i64,
    row: &Landed,
    tip: &str,
) -> Judged {
    let events = match workflow::verdict_events(conn, card) {
        Ok(events) => events,
        Err(e) => return none("check_unreadable", e.to_string()),
    };
    let verdict_head = verdict_head_of(project, entry, &events, row);
    let input = check::Input {
        project,
        conn,
        card_id: card,
        head: &entry.sha,
        verdict_head: &verdict_head,
        tip,
        on_tip: true,
        lander: Some(&row.lander),
    };
    match check::judge(&input) {
        Ok(check::Outcome::Relied(relied)) => {
            let authors = match check::authors_of(project, tip, &verdict_head) {
                Ok(authors) => authors,
                Err(why) => return none("check_unreadable", why),
            };
            let holders = match workflow::reservation_holders(conn, card) {
                Ok(holders) => holders,
                Err(e) => return none("check_unreadable", e.to_string()),
            };
            let mut excluded: Vec<String> = Vec::new();
            for session in authors.sessions.into_iter().chain(holders) {
                if !excluded.iter().any(|s| s.eq_ignore_ascii_case(&session)) {
                    excluded.push(session);
                }
            }
            if !renderable(&relied.reviewer) || excluded.iter().any(|s| !renderable(s)) {
                return none(
                    "session_id_unrenderable",
                    "a session id has a blank or a control character",
                );
            }
            Judged::Pass(Draft {
                card,
                head: verdict_head,
                tier: relied.tier,
                reviewer: relied.reviewer,
                excluded,
                event: relied.event,
            })
        }
        Ok(check::Outcome::Refused(refusal)) if refusal.code == "verdict_fail" => Judged::Fail {
            code: refusal.code.to_string(),
            detail: refusal.why,
        },
        Ok(check::Outcome::Refused(refusal)) => none(refusal.code, refusal.why),
        Err(why) => none("check_unreadable", why),
    }
}

fn intern(table: &mut Vec<String>, session: &str) -> usize {
    match table.iter().position(|s| s.eq_ignore_ascii_case(session)) {
        Some(i) => i,
        None => {
            table.push(session.to_string());
            table.len() - 1
        }
    }
}

fn card_names(cards: &[i64]) -> Vec<String> {
    cards.iter().map(|c| format!("C-{c}")).collect()
}

/// Builds the summary of `base..head` from the git history and the ledger (read only).
pub fn build(req: &Request) -> Result<Built> {
    let project = req.project;
    check_commit(project, "--base", req.base)?;
    check_commit(project, "--head", req.head)?;
    let entries = read_entries(project, req.base, req.head)?;
    let project_id = ledger::command::project_id_of(project)?;
    let read = ledger::open_read_only(req.home, &project_id).map_err(|e| match e {
        LedgerError::Absent => "ledger not found: register it first".to_string(),
        other => other.to_string(),
    })?;
    let conn = read.conn();
    let in_range: HashSet<&str> = entries.iter().map(|e| e.sha.as_str()).collect();
    let by_head = read_landed(conn, &in_range)?;
    let started = read_started(conn)?;
    let index: HashMap<&str, usize> = entries
        .iter()
        .enumerate()
        .map(|(i, e)| (e.sha.as_str(), i))
        .collect();
    // The commit the landing at line `at` started from, when that is the base or an earlier
    // line of the range: the one place a tip counts (SPEC D-7).
    let start_of = |card: i64, at: usize| -> Option<isize> {
        let tip = started.get(&(card, entries[at].sha.clone()))?;
        if tip == req.base {
            return Some(-1);
        }
        index
            .get(tip.as_str())
            .filter(|&&i| i < at)
            .map(|&i| i as isize)
    };
    // A commit between the tip a landing started from and its head, with no landing of its own,
    // belongs to that landing (a fast-forwarded lane); the nearest later landing takes it. Only a
    // fast-forward absorbs: the audited head of a merge onto a moved tip is not the landed
    // commit, so the commits under it were not audited with it.
    let mut absorbed: Vec<Option<usize>> = vec![None; entries.len()];
    for (at, entry) in entries.iter().enumerate() {
        let Some(rows) = by_head.get(&entry.sha) else {
            continue;
        };
        for (&card, row) in rows {
            let Some(from) = start_of(card, at) else {
                continue;
            };
            let Ok(events) = workflow::verdict_events(conn, card) else {
                continue;
            };
            if verdict_head_of(project, entry, &events, row) != entry.sha {
                continue;
            }
            for k in (from + 1) as usize..at {
                if absorbed[k].is_none() && !by_head.contains_key(&entries[k].sha) {
                    absorbed[k] = Some(at);
                }
            }
        }
    }
    let mut sessions: Vec<String> = Vec::new();
    let mut lines = Vec::with_capacity(entries.len());
    let mut missing = Vec::new();
    let mut counts = [0usize; 4];
    for (at, entry) in entries.iter().enumerate() {
        let parent = entry.parents[0][..12].to_string();
        let title = clean_title(&entry.subject);
        let kind = if let Some(target) = absorbed[at] {
            counts[1] += 1;
            Kind::Abs {
                target: entries[target].sha[..12].to_string(),
            }
        } else if let Some(rows) = by_head.get(&entry.sha) {
            let mut judged: Vec<(i64, Judged)> = Vec::new();
            for (&card, row) in rows {
                let tip = match start_of(card, at) {
                    Some(-1) => req.base.to_string(),
                    Some(i) => entries[i as usize].sha.clone(),
                    None => entry.parents[0].clone(),
                };
                judged.push((card, judge_card(project, conn, entry, card, row, &tip)));
            }
            let cards: Vec<i64> = judged.iter().map(|(c, _)| *c).collect();
            let reasons: Vec<Value> = judged
                .iter()
                .filter_map(|(card, j)| match j {
                    Judged::Pass(_) => None,
                    Judged::Fail { code, detail } | Judged::None { code, detail } => {
                        Some(json!({"card": format!("C-{card}"), "why": code, "detail": detail}))
                    }
                })
                .collect();
            let failed = judged.iter().find_map(|(_, j)| match j {
                Judged::Fail { code, .. } => Some(code.clone()),
                _ => None,
            });
            if let Some(why) = failed {
                counts[3] += 1;
                missing.push(json!({
                    "commit": entry.sha, "state": "FAIL", "cards": card_names(&cards),
                    "why": why, "reasons": reasons,
                }));
                Kind::Fail { why, cards, title }
            } else if !reasons.is_empty() {
                counts[2] += 1;
                missing.push(json!({
                    "commit": entry.sha, "state": "NONE", "cards": card_names(&cards),
                    "why": reasons[0]["why"], "reasons": reasons,
                }));
                Kind::None { cards, title }
            } else {
                counts[0] += 1;
                let verdicts = judged
                    .into_iter()
                    .filter_map(|(_, j)| match j {
                        Judged::Pass(d) => Some(d),
                        _ => None,
                    })
                    .map(|d| Verdict {
                        card: d.card,
                        head: d.head[..12].to_string(),
                        tier: d.tier,
                        reviewer: intern(&mut sessions, &d.reviewer),
                        excluded: d
                            .excluded
                            .iter()
                            .map(|s| intern(&mut sessions, s))
                            .collect(),
                        event: d.event,
                    })
                    .collect();
                Kind::Pass(verdicts)
            }
        } else {
            counts[2] += 1;
            missing.push(json!({
                "commit": entry.sha, "state": "NONE", "cards": [],
                "why": "no_landing_row", "reasons": [],
            }));
            Kind::None {
                cards: Vec::new(),
                title,
            }
        };
        lines.push(Line {
            sha: entry.sha.clone(),
            parent,
            kind,
        });
    }
    let summary = Summary {
        head: req.head.to_string(),
        base: req.base.to_string(),
        sessions,
        lines,
    };
    let comment = comment(&summary).map_err(|why| {
        format!(
            "the summary for {} landings cannot be posted: {why}",
            summary.lines.len()
        )
    })?;
    Ok(Built {
        project_id,
        digest: plan_digest(&comment),
        comment,
        summary,
        missing,
        counts,
    })
}

/// The preview: the comment, its digest and the missing landings. Reads, sends nothing.
pub fn plan(req: &Request) -> Result<Value> {
    Ok(build(req)?.json(json!({
        "writes": false,
        "network": "none",
        "post": "release-audit-summary with --post --pr N --repo OWNER/REPO, this plan_digest and a \
                 --permission-reference, only after the user approves; a summary already on the \
                 pull request is not posted again",
    })))
}

/// The error text of an incomplete summary: how many landings and the first of them by name.
pub fn incomplete_error(plan: &Value) -> String {
    let missing = plan["missing"].as_array().cloned().unwrap_or_default();
    let names: Vec<String> = missing
        .iter()
        .take(10)
        .filter_map(|m| m["commit"].as_str())
        .map(|c| c.chars().take(12).collect())
        .collect();
    let more = if missing.len() > names.len() {
        format!(" and {} more", missing.len() - names.len())
    } else {
        String::new()
    };
    format!(
        "the summary is incomplete: {} landing(s) have no passing verdict ({}{more})",
        missing.len(),
        names.join(", ")
    )
}

/// What `--post` was given.
pub struct PostArgs<'a> {
    pub pr: Option<u64>,
    pub repo: Option<&'a str>,
    pub plan_digest: Option<&'a str>,
    pub permission_reference: Option<&'a str>,
    pub allow_incomplete: bool,
    pub require_complete: bool,
}

/// A summary that passed every check of `--post`: nothing has been sent yet.
pub struct Approved {
    target: SummaryTarget,
    body: String,
    marker: String,
    digest: String,
    project_id: String,
    landings: usize,
}

/// Every refusal of `--post`, before anything is sent (REQ-9): a value that is missing, the plan
/// or the ledger that changed since the preview (the summary is built again and its digest is
/// compared), a summary over the cap, and an incomplete summary without `--allow-incomplete`.
pub fn approve(req: &Request, post: &PostArgs) -> Result<Approved> {
    let pr = post.pr.ok_or("--post needs --pr <number>")?;
    let repo = post.repo.ok_or("--post needs --repo OWNER/REPO")?;
    let digest = post
        .plan_digest
        .ok_or("--post needs --plan-digest <digest> from the preview")?;
    let reference = post
        .permission_reference
        .ok_or("permission reference required: --post needs --permission-reference <ref>")?;
    if !crate::contract::permission_reference_ok(reference) {
        return Err("permission reference required".to_string());
    }
    if pr == 0 {
        return Err("--pr is a pull request number, 1 or more".to_string());
    }
    let project = ProjectRef::parse(HostKind::Github, repo).map_err(|e| e.to_string())?;
    let built = build(req)?;
    if built.digest != digest {
        return Err(
            "stale or altered plan: the ledger, the range or the comment changed since the \
             preview; nothing was posted"
                .to_string(),
        );
    }
    if !built.complete() {
        let why = incomplete_error(&built.json(json!({})));
        if post.require_complete {
            return Err(why);
        }
        if !post.allow_incomplete {
            return Err(format!(
                "{why}; nothing was posted (--allow-incomplete posts it anyway)"
            ));
        }
    }
    Ok(Approved {
        target: SummaryTarget {
            project,
            number: pr,
        },
        body: render(&built.summary),
        marker: marker(&built.summary.head),
        digest: built.digest,
        project_id: built.project_id,
        landings: built.summary.lines.len(),
    })
}

/// Posts an approved summary once through the connector: a member's comment with the marker
/// (or one by the token's own account) already on the pull request means nothing is posted.
pub fn apply(approved: &Approved, connector: &dyn HostConnector) -> Result<Value> {
    if connector.host() != approved.target.project.host() {
        return Err("the connector is for another host than the target".to_string());
    }
    let outcome = connector
        .post_summary(&approved.target, &approved.body, &approved.marker)
        .map_err(|e| e.to_string())?;
    let (state, comment_id) = match outcome {
        SummaryOutcome::Posted { comment_id } => ("posted", comment_id),
        SummaryOutcome::AlreadyPosted { comment_id } => ("already_posted", comment_id),
    };
    Ok(json!({
        "action": "release-audit-summary",
        "project_id": approved.project_id,
        "state": state,
        "comment_id": comment_id,
        "marker": approved.marker,
        "plan_digest": approved.digest,
        "landings": approved.landings,
        "writes": state == "posted",
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    fn sha(n: u64) -> String {
        // The number leads, so the twelve digits that name a commit differ from commit to commit.
        format!("{n:012x}{}", "f".repeat(28))
    }

    fn twelve(n: u64) -> String {
        sha(n)[..12].to_string()
    }

    const S_REVIEW: &str = "cccccccc-0000-4000-8000-00000000000c:0123456789ab";
    const S_AUTHOR: &str = "aaaaaaaa-0000-4000-8000-00000000000a:0123456789ab";
    const S_LEGACY: &str = "bbbbbbbb-0000-4000-8000-00000000000b";

    /// base = sha(0); lines sha(1)..sha(5); head = sha(5).
    fn sample() -> Summary {
        let line = |n: u64, kind: Kind| Line {
            sha: sha(n),
            parent: twelve(n - 1),
            kind,
        };
        Summary {
            head: sha(5),
            base: sha(0),
            sessions: vec![S_REVIEW.into(), S_AUTHOR.into(), S_LEGACY.into()],
            lines: vec![
                line(1, Kind::Abs { target: twelve(3) }),
                line(2, Kind::Abs { target: twelve(3) }),
                line(
                    3,
                    Kind::Pass(vec![
                        Verdict {
                            card: 12,
                            head: twelve(30),
                            tier: 2,
                            reviewer: 0,
                            excluded: vec![1, 2],
                            event: 77,
                        },
                        Verdict {
                            card: 13,
                            head: twelve(31),
                            tier: 0,
                            reviewer: 0,
                            excluded: vec![1],
                            event: 78,
                        },
                    ]),
                ),
                line(
                    4,
                    Kind::None {
                        cards: vec![],
                        title: "a direct commit".into(),
                    },
                ),
                line(
                    5,
                    Kind::Fail {
                        why: "verdict_fail".into(),
                        cards: vec![7, 9],
                        title: "fix | the pipe".into(),
                    },
                ),
            ],
        }
    }

    #[test]
    fn render_and_parse_round_trip() {
        let summary = sample();
        let body = render(&summary);
        assert_eq!(parse(&body), Ok(summary.clone()), "{body}");
        // The posted bytes carry the marker and a blank line; GitHub may hand back CRLF.
        let posted = comment(&summary).expect("a small summary is under the cap");
        assert_eq!(parse(&posted), Ok(summary.clone()), "{posted}");
        assert_eq!(parse(&posted.replace('\n', "\r\n")), Ok(summary));
        // The text is the one the connector would post for this body.
        assert_eq!(posted, summary_comment_body(&body, &marker(&sha(5))));
        assert!(posted.ends_with(&marker(&sha(5))), "{posted}");
    }

    #[test]
    fn the_first_line_never_starts_like_an_audit_verdict() {
        let body = render(&sample());
        let first = body.lines().next().unwrap();
        assert!(first.starts_with("RELEASE-AUDIT-SUMMARY head="), "{first}");
        assert!(!first.starts_with("AUDIT #"), "{first}");
        assert_eq!(crate::comment_authority::control_kind(first), None);
        assert!(!crate::comment_authority::looks_like_control_line(first));
    }

    /// The good body with one line replaced, or one line removed (`None`).
    fn with_line(prefix: &str, replace: Option<&str>) -> String {
        let body = render(&sample());
        let mut out = Vec::new();
        for l in body.lines() {
            if l.starts_with(prefix) {
                if let Some(r) = replace {
                    out.push(r.to_string());
                }
            } else {
                out.push(l.to_string());
            }
        }
        out.join("\n")
    }

    #[test]
    fn a_broken_comment_is_refused() {
        let good = render(&sample());
        assert!(parse(&good).is_ok());
        let l3 = format!("L {} p={}", sha(3), twelve(2));
        let cases: Vec<(&str, String)> = vec![
            ("empty", String::new()),
            (
                "another header",
                good.replacen(HEADER, "RELEASE-SUMMARY", 1),
            ),
            (
                "a count too high",
                good.replacen("landings=5", "landings=6", 1),
            ),
            (
                "a count too low",
                good.replacen("landings=5", "landings=4", 1),
            ),
            ("a zero count", good.replacen("landings=5", "landings=0", 1)),
            (
                "a signed count",
                good.replacen("landings=5", "landings=+5", 1),
            ),
            ("a short head", good.replacen(&sha(5), &twelve(5), 1)),
            (
                "an upper-case base",
                good.replacen(&sha(0), &sha(0).replace('0', "A"), 1),
            ),
            (
                "a head that is not the last line",
                good.replacen(&format!("head={}", sha(5)), &format!("head={}", sha(4)), 1),
            ),
            ("an unknown line", format!("{good}\nX something")),
            (
                "a blank line inside",
                good.replacen("\nS s2", "\n\nS s2", 1),
            ),
            ("a space too many", good.replacen("S s1 ", "S s1  ", 1)),
            (
                "a session table that skips s1",
                good.replacen("S s1 ", "S s2 ", 1),
            ),
            ("a session twice", good.replacen(S_AUTHOR, S_REVIEW, 1)),
            (
                "a session id with a space",
                good.replacen(S_LEGACY, "bbbb bbbb", 1),
            ),
            (
                "a gap in the chain",
                with_line(
                    &l3,
                    Some(&format!("L {} p={} ABS {}", sha(3), twelve(9), twelve(3))),
                ),
            ),
            (
                "a repeated commit",
                good.replacen(&format!("L {}", sha(2)), &format!("L {}", sha(1)), 1),
            ),
            (
                "the first line does not start at the base",
                good.replacen(&format!("p={}", twelve(0)), &format!("p={}", twelve(9)), 1),
            ),
            ("a missing last line", {
                let l5 = format!("L {} ", sha(5));
                with_line(&l5, None)
            }),
            ("a missing first line", {
                let l1 = format!("L {} ", sha(1));
                with_line(&l1, None)
            }),
            (
                "an ABS that points at an earlier line",
                good.replacen(
                    &format!("ABS {}", twelve(3)),
                    &format!("ABS {}", twelve(1)),
                    1,
                ),
            ),
            (
                "an ABS that points nowhere",
                good.replacen(
                    &format!("ABS {}", twelve(3)),
                    &format!("ABS {}", twelve(99)),
                    1,
                ),
            ),
            (
                "an ABS that points at an ABS",
                good.replacen(
                    &format!("ABS {}", twelve(3)),
                    &format!("ABS {}", twelve(2)),
                    1,
                ),
            ),
            (
                "a PASS with no verdict",
                with_line(
                    &format!("L {} ", sha(3)),
                    Some(&format!("L {} p={} PASS", sha(3), twelve(2))),
                ),
            ),
            (
                "a verdict with an undefined session",
                good.replacen(":s1:", ":s9:", 1),
            ),
            (
                "a verdict with session s0",
                good.replacen(":s1:", ":s0:", 1),
            ),
            (
                "a verdict with tier 3",
                good.replacen(":2:s1:", ":3:s1:", 1),
            ),
            (
                "a verdict with a short head",
                good.replacen(&twelve(30), &twelve(30)[..11], 1),
            ),
            (
                "a verdict with a card of zero",
                good.replacen("v=12:", "v=0:", 1),
            ),
            (
                "a verdict with a missing field",
                good.replacen(":77", "", 1),
            ),
            (
                "a NONE with no title",
                good.replacen(" | a direct commit", "", 1),
            ),
            (
                "a NONE with a bad card list",
                good.replacen("c=- |", "c=1,,2 |", 1),
            ),
            (
                "a FAIL with no reason",
                good.replacen("why=verdict_fail ", "", 1),
            ),
            (
                "a FAIL reason with a capital",
                good.replacen("why=verdict_fail", "why=Verdict", 1),
            ),
            (
                "a title over 40 bytes",
                good.replacen("a direct commit", &"x".repeat(41), 1),
            ),
            (
                "a marker for another head",
                format!("{good}\n\n{}", marker(&sha(4))),
            ),
            (
                "text after the marker",
                format!("{good}\n\n{}\nmore", marker(&sha(5))),
            ),
            (
                "a marker with no blank line before it",
                format!("{good}\n{}", marker(&sha(5))),
            ),
            (
                "a second marker",
                format!("{good}\n\n{}\n\n{}", marker(&sha(5)), marker(&sha(5))),
            ),
        ];
        for (what, text) in cases {
            let answer = parse(&text);
            assert!(answer.is_err(), "{what} was read: {text}");
        }
    }

    #[test]
    fn a_comment_over_the_cap_is_not_read() {
        let good = render(&sample());
        let padded = format!("{good}\n{}", "x".repeat(MAX_COMMENT_BYTES));
        assert!(parse(&padded).is_err());
    }

    #[test]
    fn an_empty_exclusion_list_is_read_and_not_counted() {
        let good = render(&sample());
        let none = good.replacen(":s2.s3:", ":-:", 1);
        let summary = parse(&none).expect("`-` is the spelling of an empty list");
        let Kind::Pass(v) = &summary.lines[2].kind else {
            panic!("line 3 is a PASS");
        };
        assert!(v[0].excluded.is_empty());
        assert!(!v[0].counts(&summary), "an empty list never counts");
        assert!(v[1].counts(&summary), "the other verdict is unchanged");
    }

    #[test]
    fn a_reviewer_among_the_excluded_is_not_counted_even_as_a_bare_id() {
        let mut summary = sample();
        let Kind::Pass(v) = &mut summary.lines[2].kind else {
            panic!("line 3 is a PASS");
        };
        // The author is the reviewer: the same full id.
        v[0].excluded = vec![0, 1];
        let v0 = v[0].clone();
        assert!(!v0.counts(&summary));
        // The excluded session is the bare id of the reviewer's session: a legacy author id
        // matches every value of that session, so the reviewer is not independent.
        summary.sessions[2] = "cccccccc-0000-4000-8000-00000000000c".into();
        let Kind::Pass(v) = &mut summary.lines[2].kind else {
            panic!("line 3 is a PASS");
        };
        v[0].excluded = vec![2];
        let v0 = v[0].clone();
        assert!(!v0.counts(&summary), "a bare id stands for the full one");
        // A different session counts.
        let Kind::Pass(v) = &mut summary.lines[2].kind else {
            panic!("line 3 is a PASS");
        };
        v[0].excluded = vec![1];
        let v0 = v[0].clone();
        assert!(v0.counts(&summary));
    }

    #[test]
    fn the_cap_holds_at_its_edge() {
        assert_eq!(check_size(0), Ok(()));
        assert_eq!(check_size(MAX_COMMENT_BYTES), Ok(()));
        let over = check_size(MAX_COMMENT_BYTES + 1).expect_err("one byte over is refused");
        assert!(over.contains("60000"), "{over}");
        assert!(over.contains("not cut") || over.contains("refus"), "{over}");
    }

    /// `n` lines of direct commits, a chain from sha(0).
    fn many(n: u64) -> Summary {
        let lines = (1..=n)
            .map(|i| Line {
                sha: sha(i),
                parent: twelve(i - 1),
                kind: Kind::None {
                    cards: vec![],
                    title: "a title of exactly forty bytes, no more".into(),
                },
            })
            .collect();
        Summary {
            head: sha(n),
            base: sha(0),
            sessions: vec![],
            lines,
        }
    }

    #[test]
    fn a_summary_over_the_cap_is_refused_at_the_edge_and_never_cut() {
        // Find the largest summary whose posted comment fits.
        let mut fits = 1;
        while comment(&many(fits + 1)).is_ok() {
            fits += 1;
        }
        let ok = comment(&many(fits)).expect("N lines fit");
        assert!(ok.len() <= MAX_COMMENT_BYTES, "{}", ok.len());
        assert!(
            ok.len() + 100 > MAX_COMMENT_BYTES,
            "the edge is one line away: {}",
            ok.len()
        );
        // N lines read back whole: nothing was cut.
        assert_eq!(parse(&ok).unwrap().lines.len() as u64, fits);
        // N + 1 lines are refused with the cap named, and no partial text comes back.
        let refused = comment(&many(fits + 1)).expect_err("N + 1 lines are over the cap");
        assert!(refused.contains("60000"), "{refused}");
    }

    #[test]
    fn a_title_is_cleaned_cut_on_a_character_and_never_empty() {
        assert_eq!(clean_title("fix the parser"), "fix the parser");
        assert_eq!(clean_title(""), "-");
        assert_eq!(clean_title("   "), "-");
        // Exactly 40 bytes stays, 41 is cut to 40.
        let forty = "a".repeat(TITLE_MAX_BYTES);
        assert_eq!(clean_title(&forty), forty);
        assert_eq!(clean_title(&format!("{forty}b")), forty);
        // A Korean title is cut on a character boundary: 13 characters of 3 bytes are 39 bytes.
        let korean = "한".repeat(20);
        let cut = clean_title(&korean);
        assert!(cut.len() <= TITLE_MAX_BYTES, "{}", cut.len());
        assert_eq!(cut, "한".repeat(13));
        // Only the first line is a title.
        assert_eq!(clean_title("first\nsecond"), "first");
        // A title that could act as a control line is defused, a mention does not notify.
        assert!(clean_title("audit: PASS").starts_with("· "));
        assert!(!clean_title("ping @octocat").contains("@octocat"));
        // What it returns is what the grammar reads back.
        let mut summary = many(2);
        summary.lines[1].kind = Kind::None {
            cards: vec![4],
            title: clean_title("ping @octocat 한글 — done"),
        };
        assert_eq!(parse(&render(&summary)), Ok(summary));
    }

    #[test]
    fn the_digest_covers_the_posted_bytes() {
        let summary = sample();
        let posted = comment(&summary).unwrap();
        assert_eq!(plan_digest(&posted), crate::digest(posted.as_bytes()));
        let mut other = summary;
        other.lines[3].kind = Kind::None {
            cards: vec![],
            title: "another".into(),
        };
        assert_ne!(plan_digest(&comment(&other).unwrap()), plan_digest(&posted));
    }
}
