//! GEARS requirements and legacy EARS import (Issue #2001, 05-A).
//!
//! Requirements are checked on four separate axes. Syntax and structure are decided here;
//! semantic only records whether a model review happened (`unknown` until one is recorded);
//! execution is always `not_run`, because running acceptance belongs to VERIFY. The enum types
//! carry no "passed" value for those two axes, so no code path can promote them.
//!
//! The canonical serialization behind the snapshot digest is HOLD (P06 / HOLD-00D-5). The
//! interim implementation, `SortedKeys`, sits behind the `Canonical` trait so the decided
//! form can replace it without touching callers.

use crate::digest;
use serde::{Deserialize, Serialize};
use serde_json::Value;

pub type Result<T> = std::result::Result<T, String>;

/// GEARS reference page revision the parser follows (sublang.ai, 2026-01-14).
pub const GEARS_GRAMMAR: &str = "gears@2026-01-14";
/// EARS as published at RE'09 (Mavin et al.), the legacy grammar imported as written.
pub const EARS_GRAMMAR: &str = "ears@2009";
const MAX_INPUT: usize = 64 * 1024;

#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Requirement {
    pub id: String,
    /// The original text, unchanged.
    pub text: String,
    /// `GEARS` for new authoring, `EARS-import` for imported legacy text.
    pub syntax: String,
    pub grammar_version: String,
    #[serde(rename = "where", default)]
    pub where_: Vec<String>,
    #[serde(rename = "while", default)]
    pub while_: Vec<String>,
    pub when: Option<String>,
    pub subject: Option<String>,
    pub behavior: Option<String>,
    /// `where`, `while`, `when` in the order they appear.
    #[serde(default)]
    pub condition_kinds: Vec<String>,
    /// Never filled with a default; absent until an author states it.
    pub observable_result: Option<String>,
    #[serde(default)]
    pub acceptance_refs: Vec<String>,
    #[serde(default)]
    pub source_refs: Vec<String>,
    #[serde(default)]
    pub unresolved: Vec<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub import: Option<EarsImport>,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct EarsImport {
    /// One of the six EARS patterns, identified only; nothing is converted.
    pub pattern: String,
    /// The label the source gave the requirement, kept verbatim.
    pub classification: Option<String>,
    pub source_range: SourceRange,
}

/// 1-based inclusive lines and a byte range that slices `text` out of the source.
#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct SourceRange {
    pub start_line: usize,
    pub end_line: usize,
    pub start_byte: usize,
    pub end_byte: usize,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Acceptance {
    pub id: String,
    pub observable_result: String,
    pub verification: Verification,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Verification {
    pub kind: String,
    pub selector: String,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Snapshot {
    /// `<spec_id>@<revision>`.
    pub reference: String,
    pub digest: String,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct SpecDocument {
    pub spec_id: String,
    pub revision: u32,
    /// `proposed` or `approved`.
    pub status: String,
    pub syntax: String,
    pub source: Option<String>,
    pub requirements: Vec<Requirement>,
    pub acceptance: Vec<Acceptance>,
    #[serde(default)]
    pub scope: Vec<String>,
    pub policy_revision: Option<String>,
    #[serde(default)]
    pub verification_commands: Vec<String>,
    #[serde(default)]
    pub unresolved: Vec<String>,
    pub snapshot: Option<Snapshot>,
}

/// Where a GEARS sentence could not be split, and why. `reason` never quotes the input.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct SyntaxError {
    pub offset: usize,
    pub reason: String,
}

/// Size, control characters and secret-shaped strings, as the profile `portable()` rule
/// checks them; line breaks and tabs stay allowed and `/` is not refused (paths are normal
/// in requirement text). `sk-` counts only at the start of a word.
fn check_text(s: &str) -> std::result::Result<(), &'static str> {
    if s.len() > MAX_INPUT {
        return Err("input exceeds 64 KiB");
    }
    if s.chars()
        .any(|c| c.is_control() && !matches!(c, '\n' | '\r' | '\t'))
    {
        return Err("invalid spec text: control characters; no values were logged");
    }
    let lower = s.to_ascii_lowercase();
    let secret = [
        "ghp_",
        "github_pat_",
        "-----begin",
        "password=",
        "token=",
        "api_key=",
    ]
    .iter()
    .any(|x| lower.contains(x))
        || lower.match_indices("sk-").any(|(i, _)| {
            lower[..i]
                .chars()
                .next_back()
                .is_none_or(|c| !c.is_ascii_alphanumeric())
        });
    if secret {
        return Err("invalid spec text: secret-shaped value; no values were logged");
    }
    Ok(())
}

/// Byte offsets of `word` as a whole word, ignoring ASCII case.
fn word_positions(text: &str, word: &str) -> Vec<usize> {
    let lower = text.to_ascii_lowercase();
    let bytes = lower.as_bytes();
    let outside = |i: usize| bytes.get(i).is_none_or(|b| !b.is_ascii_alphanumeric());
    lower
        .match_indices(word)
        .map(|(i, _)| i)
        .filter(|&i| (i == 0 || outside(i - 1)) && outside(i + word.len()))
        .collect()
}

fn keyword(segment: &str) -> Option<(&'static str, &str)> {
    ["where", "while", "when"].into_iter().find_map(|k| {
        let head = segment.get(..k.len())?;
        let rest = &segment[k.len()..];
        (head.eq_ignore_ascii_case(k) && rest.starts_with(' ')).then(|| (k, rest.trim()))
    })
}

/// Splits `[Where …] [While …] [When …] <subject> shall <behavior>` into a typed record.
pub fn parse_gears(id: &str, text: &str) -> std::result::Result<Requirement, SyntaxError> {
    let fail = |offset: usize, reason: &str| SyntaxError {
        offset,
        reason: reason.into(),
    };
    check_text(text).map_err(|r| fail(0, r))?;
    let shall = word_positions(text, "shall");
    let at = match shall.as_slice() {
        [] => {
            return Err(fail(
                text.len(),
                "no `shall` separates subject and behavior",
            ))
        }
        [one] => *one,
        [_, second, ..] => {
            return Err(fail(
                *second,
                "more than one `shall`; split the requirement",
            ))
        }
    };
    let behavior = text[at + 5..].trim().trim_end_matches('.').trim();
    if behavior.is_empty() {
        return Err(fail(at + 5, "empty behavior after `shall`"));
    }
    // Prefix segments are comma-separated; a segment without a keyword continues the clause
    // before it, and the last segment is the subject.
    let prefix = &text[..at];
    let mut segments: Vec<(usize, &str)> = Vec::new();
    let mut start = 0;
    for (i, _) in prefix.match_indices(',') {
        segments.push((start, &prefix[start..i]));
        start = i + 1;
    }
    segments.push((start, &prefix[start..]));
    let (subject_at, subject) = segments.pop().expect("at least one segment");
    let subject = subject.trim();
    if subject.is_empty() || keyword(subject).is_some() {
        return Err(fail(subject_at, "empty subject before `shall`"));
    }
    let (mut where_, mut while_, mut when) = (Vec::new(), Vec::new(), None::<String>);
    let mut kinds: Vec<String> = Vec::new();
    let order = |k: &str| ["where", "while", "when"].iter().position(|x| *x == k);
    for (offset, raw) in segments {
        let lead = raw.len() - raw.trim_start().len();
        let seg = raw.trim();
        match keyword(seg) {
            Some((k, body)) => {
                if kinds.last().is_some_and(|last| order(last) > order(k))
                    || (k == "when" && when.is_some())
                {
                    return Err(fail(
                        offset + lead,
                        "condition clause out of order (Where, While, When)",
                    ));
                }
                match k {
                    "where" => where_.push(body.to_string()),
                    "while" => while_.push(body.to_string()),
                    _ => when = Some(body.to_string()),
                }
                if kinds.last().map(String::as_str) != Some(k) {
                    kinds.push(k.to_string());
                }
            }
            None => {
                let last = match kinds.last().map(String::as_str) {
                    Some("where") => where_.last_mut(),
                    Some("while") => while_.last_mut(),
                    Some(_) => when.as_mut(),
                    None => {
                        return Err(fail(
                            offset + lead,
                            "text before the subject is not a Where, While or When clause",
                        ))
                    }
                };
                let last = last.expect("clause exists");
                last.push_str(", ");
                last.push_str(seg);
            }
        }
    }
    Ok(Requirement {
        id: id.into(),
        text: text.into(),
        syntax: "GEARS".into(),
        grammar_version: GEARS_GRAMMAR.into(),
        where_,
        while_,
        when,
        subject: Some(subject.into()),
        behavior: Some(behavior.into()),
        condition_kinds: kinds,
        observable_result: None,
        acceptance_refs: Vec::new(),
        source_refs: Vec::new(),
        unresolved: Vec::new(),
        import: None,
    })
}

/// `- **REQ-X-001 (label)** — text` or `**REQ-X-001** — text`; returns (id, label).
fn requirement_head(line: &str) -> Option<(String, Option<String>)> {
    let rest = line.trim_start();
    let rest = rest
        .strip_prefix("- ")
        .or_else(|| rest.strip_prefix("* "))
        .unwrap_or(rest);
    let rest = rest.strip_prefix("**REQ-")?;
    let close = rest.find("**")?;
    let head = &rest[..close];
    let tail = rest[close + 2..].trim_start();
    if !(tail.starts_with('—') || tail.starts_with(':') || tail.starts_with('-')) {
        return None;
    }
    let (id, label) = match head.find(" (") {
        Some(p) if head.ends_with(')') => {
            (&head[..p], Some(head[p + 2..head.len() - 1].to_string()))
        }
        _ => (head, None),
    };
    let ok = !id.is_empty() && id.chars().all(|c| c.is_ascii_alphanumeric() || c == '-');
    ok.then(|| (format!("REQ-{id}"), label))
}

/// EARS unwanted-behaviour wording: `If … then`.
fn has_if_then(text: &str) -> bool {
    let plain = text.replace("**", "");
    !word_positions(&plain, "if").is_empty() && !word_positions(&plain, "then").is_empty()
}

fn ears_pattern(text: &str) -> &'static str {
    let plain = text.replace("**", "");
    let has = |w: &str| !word_positions(&plain, w).is_empty();
    let unwanted = has_if_then(text);
    let count =
        ["when", "while", "where"].iter().filter(|w| has(w)).count() + usize::from(unwanted);
    match (count, unwanted) {
        (0, _) => "ubiquitous",
        (1, true) => "unwanted-behaviour",
        (1, false) if has("when") => "event-driven",
        (1, false) if has("while") => "state-driven",
        (1, false) => "optional-feature",
        _ => "complex",
    }
}

/// Imports legacy EARS requirements from a markdown document without changing their meaning.
/// A requirement starts at a bold `REQ-` id line and runs to the next blank line or id line.
pub fn import_ears(document: &str) -> Result<Vec<Requirement>> {
    check_text(document)?;
    let mut lines: Vec<(usize, usize, &str)> = Vec::new(); // (line number, byte start, text)
    let mut at = 0;
    for (n, line) in document.split_inclusive('\n').enumerate() {
        lines.push((n + 1, at, line.trim_end_matches(['\n', '\r'])));
        at += line.len();
    }
    let mut out = Vec::new();
    let mut i = 0;
    while i < lines.len() {
        let Some((id, label)) = requirement_head(lines[i].2) else {
            i += 1;
            continue;
        };
        let first = i;
        i += 1;
        while i < lines.len()
            && !lines[i].2.trim().is_empty()
            && requirement_head(lines[i].2).is_none()
        {
            i += 1;
        }
        let (start_line, start_byte, _) = lines[first];
        let (end_line, last_start, last) = lines[i - 1];
        let end_byte = last_start + last.len();
        let text = &document[start_byte..end_byte];
        let pattern = ears_pattern(text);
        // `If … then` is never rewritten to When: whether it is an event or a state is a
        // review question (design 04 §8), recorded instead of answered.
        let unresolved = if has_if_then(text) {
            vec!["EARS If…then: decide whether the condition is an event or a state before writing it in GEARS".to_string()]
        } else {
            Vec::new()
        };
        out.push(Requirement {
            id,
            text: text.to_string(),
            syntax: "EARS-import".into(),
            grammar_version: EARS_GRAMMAR.into(),
            where_: Vec::new(),
            while_: Vec::new(),
            when: None,
            subject: None,
            behavior: None,
            condition_kinds: Vec::new(),
            observable_result: None,
            acceptance_refs: Vec::new(),
            source_refs: Vec::new(),
            unresolved,
            import: Some(EarsImport {
                pattern: pattern.into(),
                classification: label,
                source_range: SourceRange {
                    start_line,
                    end_line,
                    start_byte,
                    end_byte,
                },
            }),
        });
    }
    Ok(out)
}

#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum Outcome {
    Passed,
    Failed,
}

/// Recorded, never judged here.
#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum Semantic {
    Unknown,
    Reviewed,
}

/// Acceptance runs in VERIFY; this module only ever reports that it has not run.
#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum Execution {
    NotRun,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize)]
pub struct DanglingRef {
    pub from: String,
    pub to: String,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize)]
pub struct StructureReport {
    pub duplicate_ids: Vec<String>,
    pub dangling_refs: Vec<DanglingRef>,
    pub orphan_acceptance: Vec<String>,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize)]
pub struct SyntaxFinding {
    pub requirement: String,
    pub offset: usize,
    pub reason: String,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize)]
pub struct Validation {
    pub syntax: Outcome,
    pub structure: Outcome,
    pub semantic: Semantic,
    pub semantic_source: &'static str,
    pub execution: Execution,
    pub syntax_findings: Vec<SyntaxFinding>,
    pub structure_report: StructureReport,
    pub semantic_findings: Vec<String>,
}

/// Syntax re-parses every GEARS requirement (imported EARS text is kept as written and is
/// not parsed). Structure checks unique ids, that `acceptance_refs` name acceptance criteria,
/// that `source_refs` name a requirement or criterion of this document, and that every
/// criterion is referenced by some requirement.
pub fn validate(doc: &SpecDocument) -> Validation {
    let syntax_findings: Vec<SyntaxFinding> = doc
        .requirements
        .iter()
        .filter(|r| r.syntax == "GEARS")
        .filter_map(|r| {
            parse_gears(&r.id, &r.text).err().map(|e| SyntaxFinding {
                requirement: r.id.clone(),
                offset: e.offset,
                reason: e.reason,
            })
        })
        .collect();
    let req_ids: Vec<&str> = doc.requirements.iter().map(|r| r.id.as_str()).collect();
    let ac_ids: Vec<&str> = doc.acceptance.iter().map(|a| a.id.as_str()).collect();
    let mut seen = std::collections::BTreeSet::new();
    let mut duplicate_ids: Vec<String> = req_ids
        .iter()
        .chain(&ac_ids)
        .filter(|id| !seen.insert(**id))
        .map(|id| id.to_string())
        .collect();
    duplicate_ids.dedup();
    let mut dangling_refs = Vec::new();
    for r in &doc.requirements {
        for to in &r.acceptance_refs {
            if !ac_ids.contains(&to.as_str()) {
                dangling_refs.push(DanglingRef {
                    from: r.id.clone(),
                    to: to.clone(),
                });
            }
        }
        for to in &r.source_refs {
            if !req_ids.contains(&to.as_str()) && !ac_ids.contains(&to.as_str()) {
                dangling_refs.push(DanglingRef {
                    from: r.id.clone(),
                    to: to.clone(),
                });
            }
        }
    }
    let orphan_acceptance: Vec<String> = ac_ids
        .iter()
        .filter(|ac| {
            !doc.requirements
                .iter()
                .any(|r| r.acceptance_refs.iter().any(|x| x == *ac))
        })
        .map(|ac| ac.to_string())
        .collect();
    let outcome = |ok: bool| if ok { Outcome::Passed } else { Outcome::Failed };
    Validation {
        syntax: outcome(syntax_findings.is_empty()),
        structure: outcome(
            duplicate_ids.is_empty() && dangling_refs.is_empty() && orphan_acceptance.is_empty(),
        ),
        semantic: Semantic::Unknown,
        semantic_source: "model-finding",
        execution: Execution::NotRun,
        syntax_findings,
        structure_report: StructureReport {
            duplicate_ids,
            dangling_refs,
            orphan_acceptance,
        },
        semantic_findings: Vec::new(),
    }
}

/// Records that a model reviewed the requirements. It changes the semantic axis only.
pub fn record_semantic_review(v: &mut Validation, findings: Vec<String>) {
    v.semantic = Semantic::Reviewed;
    v.semantic_findings.extend(findings);
}

/// Deterministic bytes for the snapshot digest (HOLD: P06 / HOLD-00D-5).
pub trait Canonical {
    fn bytes(&self, value: &Value) -> Vec<u8>;
}

// @MX:DEBT: [AUTO] interim canonical form: compact serde_json with sorted object keys
// @MX:CEILING: strings, booleans, null and integers; no float normalization (RFC 8785 differs there)
// @MX:UPGRADE: replace when P06 / HOLD-00D-5 decides the canonical serialization
pub struct SortedKeys;
impl Canonical for SortedKeys {
    fn bytes(&self, value: &Value) -> Vec<u8> {
        // serde_json is built without `preserve_order`, so object keys serialize sorted.
        serde_json::to_vec(value).expect("JSON value serializes")
    }
}

/// The approved content: requirements, acceptance, scope, policy revision and verification
/// commands. Identity, status and the snapshot itself are outside it.
pub fn content_digest(doc: &SpecDocument, canonical: &dyn Canonical) -> String {
    let content = serde_json::json!({
        "requirements": doc.requirements,
        "acceptance": doc.acceptance,
        "scope": doc.scope,
        "policy_revision": doc.policy_revision,
        "verification_commands": doc.verification_commands,
    });
    digest(&canonical.bytes(&content))
}

/// Returns the approved copy with its snapshot; the input is not modified.
pub fn approve(doc: &SpecDocument, canonical: &dyn Canonical) -> SpecDocument {
    let mut out = doc.clone();
    out.status = "approved".into();
    out.snapshot = Some(Snapshot {
        reference: format!("{}@{}", doc.spec_id, doc.revision),
        digest: content_digest(doc, canonical),
    });
    out
}

/// A body change after approval opens the next proposed revision; the approved revision and
/// its snapshot stay as they were.
pub fn amend(
    approved: &SpecDocument,
    requirements: Vec<Requirement>,
    acceptance: Vec<Acceptance>,
) -> SpecDocument {
    let mut out = approved.clone();
    out.revision = approved.revision + 1;
    out.status = "proposed".into();
    out.requirements = requirements;
    out.acceptance = acceptance;
    out.snapshot = None;
    out
}
