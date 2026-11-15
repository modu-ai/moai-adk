//! What importing a host's open issues would make (#3071 H2a): the pure half of the preview.
//!
//! [`plan_import`] turns the issues a [`HostConnector`](super::connector::HostConnector) read
//! into the cards they would become, without a ledger, a file or a request: the card's title and
//! kind, the acceptance-criteria table read out of the body, the comments a member wrote (as
//! notes), what was left out, and a digest of the whole issue set. The apply step (a later slice)
//! reads the host again and writes only when the digest is the same.
//!
//! Issue text is data. The body is kept as written and never read for instructions; a comment
//! becomes a note only when `comment_authority` says a member wrote it, and the control lines in
//! those notes are counted by kind, not acted on. Non-member comments are counted and reported
//! (`dropped`), never kept.
use super::connector::{HostComment, HostIssue, OpenIssues};
use crate::comment_authority::{self, Authored, Dropped};
use sha2::{Digest, Sha256};
use std::collections::BTreeMap;

/// `card.kind` values (schema contract 6896940106 §3).
pub const KINDS: [&str; 7] = [
    "task", "feature", "fix", "docs", "research", "audit", "chore",
];
/// `card.title` is 1 to 300 characters.
pub const TITLE_MAX: usize = 300;
/// Most acceptance-criteria rows read from one body; more is "too many", not a cut-off list.
pub const MAX_AC_ROWS: usize = 200;
/// Longest cell of an acceptance-criteria row, in characters.
pub const MAX_CELL_CHARS: usize = 2000;

/// Where a card's kind came from.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum KindSource {
    /// Exactly one label named a kind.
    Label,
    /// No label named a kind, so `task`.
    Default,
    /// Labels named different kinds, so `task`; the preview says so.
    Ambiguous,
}

/// The kind a set of labels names: `kind:<kind>`, a bare kind name, or one of three common
/// aliases (`bug` is `fix`, `documentation` is `docs`, `enhancement` is `feature`).
pub fn kind_from_labels(labels: &[String]) -> (&'static str, KindSource) {
    let mut found: Vec<&'static str> = Vec::new();
    for label in labels {
        let name = label.trim().to_ascii_lowercase();
        let name = name.strip_prefix("kind:").unwrap_or(&name).trim();
        let kind = match name {
            "bug" => Some("fix"),
            "documentation" => Some("docs"),
            "enhancement" => Some("feature"),
            other => KINDS.iter().copied().find(|k| *k == other),
        };
        if let Some(kind) = kind {
            if !found.contains(&kind) {
                found.push(kind);
            }
        }
    }
    match found.as_slice() {
        [one] => (one, KindSource::Label),
        [] => ("task", KindSource::Default),
        _ => ("task", KindSource::Ambiguous),
    }
}

/// One row of an acceptance-criteria table as written.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct AcRow {
    /// The `#` cell (`AC1`, `1`, …); the ledger numbers the rows itself.
    pub label: String,
    pub text: String,
    pub verify: String,
}

/// Why a body gave no acceptance criteria.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum AcProblem {
    NoTable,
    /// Two tables have the criteria header: which one counts is not guessed.
    Ambiguous,
    /// The table has a header but no row with a criterion.
    Empty,
    TooManyRows,
    CellTooLong,
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub enum AcReading {
    /// `skipped` counts table rows that had no criterion text.
    Read {
        rows: Vec<AcRow>,
        skipped: usize,
    },
    Unreadable(AcProblem),
}

const CRITERION_HEADERS: [&str; 3] = ["기준", "criterion", "criteria"];
const VERIFY_HEADERS: [&str; 6] = [
    "확인 방법",
    "확인방법",
    "verify",
    "verification",
    "check",
    "how to verify",
];

/// The cells of one table line, or `None` when the line is not a table row. `\|` is a bar inside
/// a cell.
fn cells(line: &str) -> Option<Vec<String>> {
    let line = line.trim();
    let inner = line.strip_prefix('|')?;
    let mut out = Vec::new();
    let mut cell = String::new();
    let mut chars = inner.chars().peekable();
    while let Some(c) = chars.next() {
        match c {
            '\\' if chars.peek() == Some(&'|') => {
                cell.push('|');
                chars.next();
            }
            '|' => out.push(std::mem::take(&mut cell)),
            other => cell.push(other),
        }
    }
    // A last cell without a closing bar still counts.
    if !cell.trim().is_empty() {
        out.push(cell);
    }
    Some(out.into_iter().map(|c| c.trim().to_string()).collect())
}

fn is_header(cells: &[String]) -> bool {
    let lower = |i: usize| cells[i].to_lowercase();
    cells.len() >= 3
        && cells[0] == "#"
        && CRITERION_HEADERS.contains(&lower(1).as_str())
        && VERIFY_HEADERS.contains(&lower(2).as_str())
}

fn is_separator(line: &str) -> bool {
    cells(line).is_some_and(|c| {
        c.len() >= 3
            && c.iter()
                .all(|cell| !cell.is_empty() && cell.chars().all(|ch| ch == '-' || ch == ':'))
            && c.iter().all(|cell| cell.contains('-'))
    })
}

/// Reads the `| # | 기준 | 확인 방법 |` table of an issue body. Tables inside fenced code are
/// not read. Two such tables are [`AcProblem::Ambiguous`]; a limit that is passed is its own
/// problem, never a shorter list.
pub fn parse_ac_table(body: &str) -> AcReading {
    let lines: Vec<&str> = body.lines().collect();
    // The marker (``` or ~~~) of the open code fence: only the same marker closes it.
    let mut fenced: Option<&str> = None;
    let mut tables: Vec<(usize, usize)> = Vec::new();
    let mut i = 0;
    while i < lines.len() {
        let line = lines[i].trim_start();
        let marker = ["```", "~~~"].into_iter().find(|m| line.starts_with(m));
        if let Some(marker) = marker {
            fenced = match fenced {
                None => Some(marker),
                Some(open) if open == marker => None,
                keep => keep,
            };
            i += 1;
            continue;
        }
        let header = fenced.is_none()
            && cells(lines[i]).is_some_and(|c| is_header(&c))
            && lines.get(i + 1).is_some_and(|next| is_separator(next));
        if header {
            let start = i + 2;
            let mut end = start;
            while end < lines.len() && lines[end].trim_start().starts_with('|') {
                end += 1;
            }
            tables.push((start, end));
            i = end;
        } else {
            i += 1;
        }
    }
    let (start, end) = match tables.as_slice() {
        [] => return AcReading::Unreadable(AcProblem::NoTable),
        [one] => *one,
        _ => return AcReading::Unreadable(AcProblem::Ambiguous),
    };
    let mut rows = Vec::new();
    let mut skipped = 0;
    for line in &lines[start..end] {
        let row = cells(line).unwrap_or_default();
        if row.iter().any(|c| c.chars().count() > MAX_CELL_CHARS) {
            return AcReading::Unreadable(AcProblem::CellTooLong);
        }
        let text = row.get(1).cloned().unwrap_or_default();
        if text.is_empty() {
            skipped += 1;
            continue;
        }
        rows.push(AcRow {
            label: row.first().cloned().unwrap_or_default(),
            text,
            verify: row.get(2).cloned().unwrap_or_default(),
        });
    }
    if rows.len() > MAX_AC_ROWS {
        return AcReading::Unreadable(AcProblem::TooManyRows);
    }
    if rows.is_empty() {
        return AcReading::Unreadable(AcProblem::Empty);
    }
    AcReading::Read { rows, skipped }
}

/// A member's comment that would become a card note.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct PlannedNote {
    pub comment_id: u64,
    pub author: String,
    pub body: String,
}

impl Authored for HostComment {
    fn comment_id(&self) -> i64 {
        i64::try_from(self.id).unwrap_or(0)
    }
    fn body(&self) -> &str {
        &self.body
    }
    fn association(&self) -> &str {
        &self.association
    }
}

/// The card one open issue would become.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct PlannedCard {
    pub source_host: &'static str,
    /// Lower case `owner/repo`.
    pub source_project: String,
    pub source_number: u64,
    pub title: String,
    pub kind: &'static str,
    pub kind_source: KindSource,
    /// The issue body exactly as written. It is text, never an instruction.
    pub body: String,
    pub ac: AcReading,
    pub notes: Vec<PlannedNote>,
    /// Non-member comments that were left out, with the control lines among them.
    pub dropped: Dropped,
    /// Control lines found in member comments, by kind. Counted for the preview, not acted on.
    pub member_control_lines: BTreeMap<&'static str, usize>,
}

/// The cards a read of the open issues would make, and a digest of what was read.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct ImportPlan {
    pub source_host: &'static str,
    pub source_project: String,
    pub cards: Vec<PlannedCard>,
    pub pull_requests_skipped: usize,
    pub digest: String,
}

impl ImportPlan {
    /// Issue numbers whose body gave no acceptance criteria.
    pub fn unreadable_ac(&self) -> Vec<u64> {
        self.cards
            .iter()
            .filter(|c| matches!(c.ac, AcReading::Unreadable(_)))
            .map(|c| c.source_number)
            .collect()
    }
}

fn title_of(issue: &HostIssue) -> String {
    let title = issue.title.trim();
    if title.is_empty() {
        return format!("issue #{}", issue.number);
    }
    title.chars().take(TITLE_MAX).collect()
}

fn plan_card(source_host: &'static str, project: &str, issue: &HostIssue) -> PlannedCard {
    let (kind, kind_source) = kind_from_labels(&issue.labels);
    let mut comments = issue.comments.clone();
    comments.sort_by_key(|c| c.id);
    let dropped = comment_authority::retain_members(&mut comments);
    let mut member_control_lines: BTreeMap<&'static str, usize> = BTreeMap::new();
    for comment in &comments {
        if let Some(kind) = comment_authority::control_kind(&comment.body) {
            *member_control_lines.entry(kind).or_default() += 1;
        }
    }
    let notes = comments
        .into_iter()
        .filter(|c| !c.body.trim().is_empty())
        .map(|c| PlannedNote {
            comment_id: c.id,
            author: c.author,
            body: c.body,
        })
        .collect();
    PlannedCard {
        source_host,
        source_project: project.to_string(),
        source_number: issue.number,
        title: title_of(issue),
        kind,
        kind_source,
        body: issue.body.clone(),
        ac: parse_ac_table(&issue.body),
        notes,
        dropped,
        member_control_lines,
    }
}

/// Appends one field with its length, so no two field lists encode alike.
fn field(out: &mut Vec<u8>, value: &str) {
    out.extend_from_slice(value.len().to_string().as_bytes());
    out.push(b':');
    out.extend_from_slice(value.as_bytes());
}

/// Appends a list: its length, then each element as a field.
fn list(out: &mut Vec<u8>, items: &[String]) {
    field(out, &items.len().to_string());
    for item in items {
        field(out, item);
    }
}

fn hash(text: &str) -> String {
    format!("{:x}", Sha256::digest(text.as_bytes()))
}

/// The digest of an issue set: everything [`plan_import`] reads, so a change to any of it is a
/// different digest. Per project and per issue (ascending by number): number, update time, title,
/// the labels (sorted), a hash of the body, and per comment (ascending by id) its id, its
/// association and a hash of its body. `sha256:<hex>`.
pub fn issue_set_digest(issues: &OpenIssues) -> String {
    let mut sorted: Vec<&HostIssue> = issues.issues.iter().collect();
    sorted.sort_by_key(|i| i.number);
    let mut out = b"moai-import-digest-v2\n".to_vec();
    field(&mut out, issues.project.host().as_str());
    field(&mut out, &issues.project.source_project());
    field(&mut out, &sorted.len().to_string());
    for issue in sorted {
        field(&mut out, &issue.number.to_string());
        field(&mut out, &issue.updated_at);
        field(&mut out, &issue.title);
        let mut labels = issue.labels.clone();
        labels.sort();
        list(&mut out, &labels);
        field(&mut out, &hash(&issue.body));
        let mut comments: Vec<&HostComment> = issue.comments.iter().collect();
        comments.sort_by_key(|c| c.id);
        field(&mut out, &comments.len().to_string());
        for c in comments {
            field(&mut out, &c.id.to_string());
            field(&mut out, &c.association);
            field(&mut out, &hash(&c.body));
        }
    }
    format!("sha256:{:x}", Sha256::digest(&out))
}

/// The cards the open issues would become, in issue-number order.
pub fn plan_import(issues: &OpenIssues) -> ImportPlan {
    let source_host = issues.project.host().as_str();
    let project = issues.project.source_project();
    let mut sorted: Vec<&HostIssue> = issues.issues.iter().collect();
    sorted.sort_by_key(|i| i.number);
    ImportPlan {
        source_host,
        source_project: project.clone(),
        cards: sorted
            .into_iter()
            .map(|issue| plan_card(source_host, &project, issue))
            .collect(),
        pull_requests_skipped: issues.pull_requests_skipped,
        digest: issue_set_digest(issues),
    }
}
