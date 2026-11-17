//! SPEC documents: `<project>/.moai/specs/<SPEC-ID>/spec.md` (SDD-1 of the SDD lifecycle plan,
//! 2026-11-17; SPEC-20261117-spec-tooling). Parses the frontmatter and the `##` sections and
//! reports what is wrong as findings with stable codes; never runs a program.
//!
//! - id: `SPEC-<yyyymmdd>-<slug>` (a real calendar date; slug `[a-z0-9]+(-[a-z0-9]+)*`, at most
//!   [`MAX_SLUG_CHARS`]). A Go-generation id (`SPEC-AUTH-001`) is read, never written, and its
//!   file is checked leniently: everything but the read, frontmatter-block, id and status codes
//!   is a warning there. The folder name must equal the id.
//! - status: [`STATUSES`]; a Go status ([`legacy_status`]) is read as its new word and reported.
//! - frontmatter: flat `key: value` lines between `---` lines, `null` read as absent;
//!   `approval:` holds `by`, `at` (a date, YYYY-MM-DD), `basis` and `digest` as indented lines
//!   or one `{…}` flow map. Not full YAML: only these shapes are read. The archive scanner's
//!   reader (`archive::frontmatter`) reads three marker keys and refuses indentation, so it
//!   cannot read `approval`; this one is separate and has no new crate.
//! - sections: `## Intent`, `## Requirements` (`- REQ-<n>: <GEARS sentence>`, parsed by
//!   [`crate::spec::parse_gears`], or the EARS unwanted form `- REQ-<n> (REQ-…): IF …, THEN …
//!   shall …`; every `When` requirement needs an IF/THEN requirement naming it or a trailing
//!   `· failure: none` note, `req_failure_path`), `## Acceptance` (`- AC-<n> (REQ-…): <result>
//!   · verify: <cmd>`, or `manual:` / `unknown:`, on the AC line or at the start of a later line
//!   of the same AC; refs read `REQ-1…REQ-3` ranges), `## Scope` (`write:`, `read-only:`,
//!   `forbidden:` relative paths; an untagged path is `write`) and `## Plan` always;
//!   `## Non-goals` and `## Risks` make a full SPEC, which then needs both; `## Decisions` and
//!   `## Result` are optional. Headings inside fenced code are text.
//! - digest: [`crate::spec::content_digest`] over the requirements, acceptance and scope; an
//!   approved SPEC records it as `approval.digest`, and a body that no longer matches it is an
//!   error (a change needs the next revision and a new approval). No `cards:` key: personal card
//!   numbers stay out of shared files.

use serde::Serialize;
use std::fs;
use std::io::Read;
use std::path::Path;

/// The largest `spec.md` read, in bytes (256 KiB).
pub const MAX_SPEC_BYTES: u64 = 256 * 1024;
/// The longest slug of a new id, in characters (all ASCII, so bytes too).
pub const MAX_SLUG_CHARS: usize = 48;
/// The longest title, in characters (not bytes: a Hangul letter is one).
pub const MAX_TITLE_CHARS: usize = 200;
/// The file name inside a SPEC folder.
pub const SPEC_FILE: &str = "spec.md";
/// The status words a SPEC is written with, in lifecycle order.
pub const STATUSES: [&str; 6] = [
    "draft",
    "approved",
    "completed",
    "superseded",
    "rejected",
    "archived",
];
/// The statuses that need an `approval` record.
const NEEDS_APPROVAL: [&str; 3] = ["approved", "completed", "superseded"];
/// Go-generation statuses and the word each reads as (read only, never written).
const LEGACY_STATUSES: [(&str, &str); 9] = [
    ("planned", "draft"),
    ("plan_ready", "draft"),
    ("ready-for-run", "approved"),
    ("in-progress", "approved"),
    ("implemented", "approved"),
    ("implemented-pending-sync", "approved"),
    ("sync-started", "approved"),
    ("backfilled", "completed"),
    ("retired", "archived"),
];
const KNOWN_KEYS: [&str; 7] = [
    "id", "title", "status", "revision", "created", "parent", "approval",
];
const APPROVAL_KEYS: [&str; 4] = ["by", "at", "basis", "digest"];
const SECTIONS: [&str; 9] = [
    "Intent",
    "Non-goals",
    "Requirements",
    "Acceptance",
    "Scope",
    "Plan",
    "Risks",
    "Decisions",
    "Result",
];
const REQUIRED: [&str; 5] = ["Intent", "Requirements", "Acceptance", "Scope", "Plan"];
const FULL_ONLY: [&str; 2] = ["Non-goals", "Risks"];
/// The codes that stay errors in a Go-generation SPEC; every other one is a warning there.
const LEGACY_KEEPS: [&str; 12] = [
    "read_missing",
    "read_not_file",
    "read_too_large",
    "read_utf8",
    "read_failed",
    "fm_missing",
    "fm_unclosed",
    "id_missing",
    "id_form",
    "id_folder",
    "status_missing",
    "status_unknown",
];

const FULL_TEMPLATE: &str = include_str!("../templates/spec/full.md");
const MINIMAL_TEMPLATE: &str = include_str!("../templates/spec/minimal.md");

#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize)]
#[serde(rename_all = "lowercase")]
pub enum Severity {
    Error,
    Warning,
}

/// One thing wrong with a SPEC. `message` names sections, keys and item ids, never the
/// author's text.
#[derive(Clone, Debug, PartialEq, Eq, Serialize)]
pub struct Finding {
    pub code: &'static str,
    pub severity: Severity,
    /// 1-based line in `spec.md`, when one line is at fault.
    pub line: Option<usize>,
    pub message: String,
}

#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize)]
#[serde(rename_all = "lowercase")]
pub enum IdForm {
    /// `SPEC-<yyyymmdd>-<slug>`, the form this generation writes.
    New,
    /// The Go generation's `SPEC-<WORD>(-<WORD>)*-<nnn>`, read only.
    Legacy,
}

#[derive(Clone, Copy, Debug, Default, PartialEq, Eq, Serialize)]
#[serde(rename_all = "lowercase")]
pub enum Form {
    /// Intent, Requirements, Acceptance, Scope and Plan.
    #[default]
    Minimal,
    /// The minimal sections plus Non-goals and Risks.
    Full,
}

#[derive(Clone, Debug, Default, PartialEq, Eq, Serialize)]
pub struct Approval {
    pub by: Option<String>,
    pub at: Option<String>,
    pub basis: Option<String>,
    /// [`SpecDoc::digest`] at approval.
    pub digest: Option<String>,
}

/// The grammar a requirement is written in.
#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize)]
#[serde(rename_all = "lowercase")]
pub enum ReqKind {
    /// `[Where …] [While …] [When …] <subject> shall <behavior>` (`spec::parse_gears`).
    Gears,
    /// EARS unwanted behaviour: `IF <condition>, THEN <subject> shall <behavior>`.
    Unwanted,
}

/// A `REQ-<n>` line.
#[derive(Clone, Debug, PartialEq, Eq, Serialize)]
pub struct Requirement {
    pub id: String,
    pub number: u32,
    pub line: usize,
    pub kind: ReqKind,
    /// The requirements an unwanted one covers: `REQ-2 (REQ-1): IF …`.
    pub refs: Vec<String>,
    /// A GEARS sentence with a `When` trigger.
    pub trigger: bool,
    /// The line says `failure: none`: its trigger has no failure path to cover.
    pub failure_none: bool,
    /// The sentence, without the refs and the `failure: none` note.
    pub text: String,
}

/// An `AC-<n>` line, the requirements it names and its `verify:` command, `manual:` reason or
/// `unknown:` reason.
#[derive(Clone, Debug, PartialEq, Eq, Serialize)]
pub struct Criterion {
    pub id: String,
    pub number: u32,
    pub line: usize,
    pub refs: Vec<String>,
    pub text: String,
    pub verify: Option<String>,
    pub manual: Option<String>,
    pub unknown: Option<String>,
}

/// `## Scope`, by kind.
#[derive(Clone, Debug, Default, PartialEq, Eq, Serialize)]
pub struct Scope {
    pub write: Vec<String>,
    pub read_only: Vec<String>,
    pub forbidden: Vec<String>,
}

/// An `AC-<n>` line of `## Result`: its first word is the outcome (`met`, `withdrawn`).
#[derive(Clone, Debug, PartialEq, Eq, Serialize)]
pub struct Outcome {
    pub ac: String,
    pub number: u32,
    pub outcome: String,
    pub line: usize,
    pub text: String,
}

/// What a `spec.md` says and what is wrong with it.
#[derive(Clone, Debug, Default, PartialEq, Eq, Serialize)]
pub struct SpecDoc {
    pub id: Option<String>,
    pub id_form: Option<IdForm>,
    pub title: Option<String>,
    /// The status as written.
    pub status: Option<String>,
    /// The status in this generation's words (a Go status mapped; `None` when unknown).
    pub status_as: Option<String>,
    /// The written status is a Go-generation word.
    pub status_legacy: bool,
    pub revision: Option<u32>,
    pub created: Option<String>,
    pub parent: Option<String>,
    pub approval: Option<Approval>,
    pub unknown_keys: Vec<String>,
    pub form: Form,
    /// The known `##` sections, in file order.
    pub sections: Vec<String>,
    pub requirements: Vec<Requirement>,
    pub acceptance: Vec<Criterion>,
    pub scope: Scope,
    pub result: Vec<Outcome>,
    /// `spec::content_digest` over requirements, acceptance and scope; `None` while a
    /// requirement is not a GEARS sentence.
    pub digest: Option<String>,
    pub findings: Vec<Finding>,
    #[serde(skip)]
    gears: Vec<crate::spec::Requirement>,
}

impl SpecDoc {
    pub fn errors(&self) -> usize {
        self.count(Severity::Error)
    }

    pub fn warnings(&self) -> usize {
        self.count(Severity::Warning)
    }

    fn count(&self, s: Severity) -> usize {
        self.findings.iter().filter(|f| f.severity == s).count()
    }

    fn error(&mut self, code: &'static str, line: Option<usize>, message: impl Into<String>) {
        self.push(code, Severity::Error, line, message);
    }

    fn warn(&mut self, code: &'static str, line: Option<usize>, message: impl Into<String>) {
        self.push(code, Severity::Warning, line, message);
    }

    fn push(
        &mut self,
        code: &'static str,
        severity: Severity,
        line: Option<usize>,
        message: impl Into<String>,
    ) {
        self.findings.push(Finding {
            code,
            severity,
            line,
            message: message.into(),
        });
    }
}

/// What [`parse`] checks besides the file itself.
#[derive(Clone, Debug, Default)]
pub struct Options {
    /// The SPEC folder's name; the id must equal it.
    pub folder: Option<String>,
    /// Closure: status `completed` and every AC `met` or `withdrawn` in `## Result`.
    pub closure: bool,
}

/// `YYYY-MM-DD`, a real calendar date (years 0001–9999).
pub fn valid_date(s: &str) -> bool {
    let b = s.as_bytes();
    b.len() == 10 && b[4] == b'-' && b[7] == b'-' && ymd(&s[..4], &s[5..7], &s[8..]).is_some()
}

/// Year, month and day digits as a real calendar date.
fn ymd(y: &str, m: &str, d: &str) -> Option<()> {
    let digits = |t: &str| !t.is_empty() && t.bytes().all(|c| c.is_ascii_digit());
    if !(digits(y) && digits(m) && digits(d)) {
        return None;
    }
    let (y, m, d): (u32, u32, u32) = (y.parse().ok()?, m.parse().ok()?, d.parse().ok()?);
    let leap = (y % 4 == 0 && y % 100 != 0) || y % 400 == 0;
    let days = match m {
        1 | 3 | 5 | 7 | 8 | 10 | 12 => 31,
        4 | 6 | 9 | 11 => 30,
        2 if leap => 29,
        2 => 28,
        _ => return None,
    };
    (y >= 1 && (1..=days).contains(&d)).then_some(())
}

/// A slug of a new id: lowercase letters and digits in words joined by single `-`, 1 to
/// [`MAX_SLUG_CHARS`] characters.
pub fn valid_slug(s: &str) -> bool {
    !s.is_empty()
        && s.len() <= MAX_SLUG_CHARS
        && s.split('-').all(|w| {
            !w.is_empty()
                && w.bytes()
                    .all(|c| c.is_ascii_lowercase() || c.is_ascii_digit())
        })
}

/// The form of a SPEC id, or `None` when it is neither.
pub fn id_form(id: &str) -> Option<IdForm> {
    if let Some(rest) = id.strip_prefix("SPEC-") {
        if let Some((date, slug)) = rest.split_once('-') {
            if date.len() == 8
                && ymd(&date[..4], &date[4..6], &date[6..]).is_some()
                && valid_slug(slug)
            {
                return Some(IdForm::New);
            }
        }
        let words: Vec<&str> = rest.split('-').collect();
        if let Some((num, names)) = words.split_last() {
            let word = |w: &&str| {
                w.bytes().next().is_some_and(|c| c.is_ascii_uppercase())
                    && w.bytes()
                        .all(|c| c.is_ascii_uppercase() || c.is_ascii_digit())
            };
            if num.len() == 3
                && num.bytes().all(|c| c.is_ascii_digit())
                && !names.is_empty()
                && names.iter().all(word)
            {
                return Some(IdForm::Legacy);
            }
        }
    }
    None
}

/// Whether an id has the new form's shape with a date that is no calendar date.
fn bad_date_id(id: &str) -> bool {
    id.strip_prefix("SPEC-")
        .and_then(|r| r.split_once('-'))
        .is_some_and(|(date, slug)| {
            date.len() == 8 && date.bytes().all(|c| c.is_ascii_digit()) && valid_slug(slug)
        })
}

/// `SPEC-<yyyymmdd>-<slug>` from `YYYY-MM-DD` and a slug.
pub fn new_id(date: &str, slug: &str) -> Result<String, String> {
    if !valid_date(date) {
        return Err("the date must be a real calendar date written YYYY-MM-DD".into());
    }
    if !valid_slug(slug) {
        return Err(format!(
            "the slug must be lowercase letters and digits in words joined by single `-`, \
             1 to {MAX_SLUG_CHARS} characters"
        ));
    }
    Ok(format!("SPEC-{}-{slug}", date.replace('-', "")))
}

/// The new word a Go-generation status reads as.
pub fn legacy_status(s: &str) -> Option<&'static str> {
    LEGACY_STATUSES
        .iter()
        .find(|(old, _)| *old == s)
        .map(|(_, new)| *new)
}

/// A title a new SPEC may carry: one line of 1 to [`MAX_TITLE_CHARS`] characters, no control
/// character, no `"` or `\` (the template writes it inside double quotes).
pub fn check_title(t: &str) -> Result<(), String> {
    if t.trim().is_empty() {
        return Err("the title must not be empty".into());
    }
    if t.chars().count() > MAX_TITLE_CHARS {
        return Err(format!(
            "the title must be at most {MAX_TITLE_CHARS} characters"
        ));
    }
    if t.chars().any(|c| c.is_control() || c == '"' || c == '\\') {
        return Err("the title must be one line without control characters, `\"` or `\\`".into());
    }
    Ok(())
}

/// The fields of a new SPEC.
pub struct NewSpec<'a> {
    pub id: &'a str,
    pub title: &'a str,
    /// `YYYY-MM-DD`.
    pub created: &'a str,
    pub parent: Option<&'a str>,
    pub minimal: bool,
}

/// The new `spec.md` from the built-in template: status `draft`, revision 1.
pub fn render(n: &NewSpec) -> Result<String, String> {
    if id_form(n.id) != Some(IdForm::New) {
        return Err("the id must be SPEC-<yyyymmdd>-<slug>".into());
    }
    check_title(n.title)?;
    if !valid_date(n.created) {
        return Err("the date must be a real calendar date written YYYY-MM-DD".into());
    }
    let parent = match n.parent {
        None => String::new(),
        Some(p) if id_form(p).is_none() => {
            return Err("the parent must be a SPEC id".into());
        }
        Some(p) if p == n.id => return Err("a SPEC cannot be its own parent".into()),
        Some(p) => format!("parent: {p}\n"),
    };
    let template = if n.minimal {
        MINIMAL_TEMPLATE
    } else {
        FULL_TEMPLATE
    };
    // The title goes in last, so a title that spells a placeholder stays as typed.
    Ok(template
        .replace("{{id}}", n.id)
        .replace("{{created}}", n.created)
        .replace("{{parent}}", &parent)
        .replace("{{title}}", n.title.trim()))
}

/// Reads one `spec.md`: a regular file (a link is never followed) of at most
/// [`MAX_SPEC_BYTES`], UTF-8.
pub fn read_spec(path: &Path) -> Result<String, Finding> {
    let fail = |code: &'static str, message: &str| Finding {
        code,
        severity: Severity::Error,
        line: None,
        message: message.to_string(),
    };
    let meta = match fs::symlink_metadata(path) {
        Ok(m) => m,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => {
            return Err(fail("read_missing", "the SPEC folder has no spec.md"))
        }
        Err(_) => return Err(fail("read_failed", "spec.md could not be examined")),
    };
    if !meta.file_type().is_file() {
        return Err(fail(
            "read_not_file",
            "spec.md is not a regular file (a link is never followed)",
        ));
    }
    let too_large = || {
        fail(
            "read_too_large",
            &format!("spec.md is larger than {} KiB", MAX_SPEC_BYTES / 1024),
        )
    };
    if meta.len() > MAX_SPEC_BYTES {
        return Err(too_large());
    }
    let file =
        open_no_follow(path).map_err(|_| fail("read_failed", "spec.md could not be opened"))?;
    let opened = file
        .metadata()
        .map_err(|_| fail("read_failed", "spec.md could not be examined"))?;
    if !opened.is_file() || !same_file(&meta, &opened) {
        return Err(fail("read_not_file", "spec.md changed while it was opened"));
    }
    let mut bytes = Vec::new();
    file.take(MAX_SPEC_BYTES + 1)
        .read_to_end(&mut bytes)
        .map_err(|_| fail("read_failed", "spec.md could not be read"))?;
    if bytes.len() as u64 > MAX_SPEC_BYTES {
        return Err(too_large());
    }
    String::from_utf8(bytes).map_err(|_| fail("read_utf8", "spec.md is not UTF-8"))
}

#[cfg(unix)]
fn open_no_follow(path: &Path) -> std::io::Result<fs::File> {
    use std::os::unix::fs::OpenOptionsExt;
    fs::OpenOptions::new()
        .read(true)
        .custom_flags(libc::O_NOFOLLOW | libc::O_NONBLOCK)
        .open(path)
}

#[cfg(not(unix))]
fn open_no_follow(path: &Path) -> std::io::Result<fs::File> {
    fs::File::open(path)
}

#[cfg(unix)]
fn same_file(a: &fs::Metadata, b: &fs::Metadata) -> bool {
    use std::os::unix::fs::MetadataExt;
    a.dev() == b.dev() && a.ino() == b.ino()
}

#[cfg(not(unix))]
fn same_file(_: &fs::Metadata, _: &fs::Metadata) -> bool {
    true
}

/// Reads and parses `<dir>/spec.md`; the folder's name is the id it must carry. A folder that
/// is a link is not entered.
pub fn load(dir: &Path, closure: bool) -> SpecDoc {
    let folder = dir.file_name().map(|n| n.to_string_lossy().into_owned());
    let real_dir = fs::symlink_metadata(dir).is_ok_and(|m| m.is_dir());
    let read = if real_dir {
        read_spec(&dir.join(SPEC_FILE))
    } else {
        Err(Finding {
            code: "read_not_file",
            severity: Severity::Error,
            line: None,
            message: "the SPEC folder is not a real folder (a link is never followed)".into(),
        })
    };
    match read {
        Ok(text) => parse(&text, &Options { folder, closure }),
        Err(f) => SpecDoc {
            findings: vec![f],
            ..SpecDoc::default()
        },
    }
}

/// One `key: value` of the frontmatter.
struct Field {
    value: String,
    line: usize,
}

/// One layer of matching double or single quotes removed.
fn unquote(v: &str) -> String {
    let t = v.trim();
    t.strip_prefix('"')
        .and_then(|s| s.strip_suffix('"'))
        .or_else(|| t.strip_prefix('\'').and_then(|s| s.strip_suffix('\'')))
        .unwrap_or(t)
        .to_string()
}

/// Records one `approval` pair; a repeated key is `fm_duplicate`, a `null` value is absent.
fn add_approval(doc: &mut SpecDoc, list: &mut Vec<(String, Field)>, k: &str, v: &str, n: usize) {
    if list.iter().any(|(seen, _)| seen == k) {
        doc.error(
            "fm_duplicate",
            Some(n),
            format!("`approval.{k}` is given twice"),
        );
        return;
    }
    if is_null(v) {
        return;
    }
    list.push((
        k.to_string(),
        Field {
            value: unquote(v),
            line: n,
        },
    ));
}

/// YAML's null spellings: the key reads as absent.
fn is_null(v: &str) -> bool {
    matches!(v.trim(), "null" | "Null" | "NULL" | "~")
}

/// A one-line YAML flow map `{k: v, k: "v, with comma"}`: its pairs in order, or `None` when
/// the text is not one (not closed, a nested `{`/`[`, a pair without `key:`). Quotes are
/// honoured; values are unquoted.
fn flow_map(v: &str) -> Option<Vec<(String, String)>> {
    let inner = v.trim().strip_prefix('{')?.strip_suffix('}')?;
    let mut pieces = Vec::new();
    let (mut quote, mut start) = (None::<char>, 0);
    for (i, c) in inner.char_indices() {
        match (quote, c) {
            (Some(q), c) if c == q => quote = None,
            (Some(_), _) => {}
            (None, '"' | '\'') => quote = Some(c),
            (None, '{' | '[' | '}' | ']') => return None,
            (None, ',') => {
                pieces.push(&inner[start..i]);
                start = i + 1;
            }
            _ => {}
        }
    }
    if quote.is_some() {
        return None;
    }
    pieces.push(&inner[start..]);
    let mut out = Vec::new();
    for p in pieces.into_iter().filter(|p| !p.trim().is_empty()) {
        let (k, val) = p.split_once(':')?;
        let k = k.trim();
        let ok = !k.is_empty()
            && k.bytes()
                .all(|c| c.is_ascii_alphanumeric() || c == b'_' || c == b'-');
        if !ok {
            return None;
        }
        out.push((k.to_string(), val.trim().to_string()));
    }
    Some(out)
}

/// A column-0 `key: value` line: the key is one word of letters, digits, `_` and `-`.
fn key_value(line: &str) -> Option<(&str, &str)> {
    let (k, v) = line.split_once(':')?;
    let ok = !k.is_empty()
        && k.bytes()
            .all(|c| c.is_ascii_alphanumeric() || c == b'_' || c == b'-');
    (ok && (v.is_empty() || v.starts_with([' ', '\t']))).then(|| (k, v.trim()))
}

/// Parses a `spec.md` text and checks it.
pub fn parse(text: &str, opts: &Options) -> SpecDoc {
    let mut doc = SpecDoc::default();
    let lines: Vec<&str> = text
        .split('\n')
        .map(|l| l.strip_suffix('\r').unwrap_or(l))
        .collect();
    let (body_start, approval_line) = frontmatter(&lines, &mut doc, opts);
    if doc.findings.iter().any(|f| f.code == "fm_unclosed") {
        // An unclosed block swallowed the file: no body to check.
        return doc;
    }
    body(&lines, body_start, &mut doc);
    digest(&mut doc, approval_line);
    unfinished(&lines, &mut doc);
    budget(text.lines().count(), &mut doc);
    if opts.closure {
        closure(&mut doc);
    }
    if doc.id_form == Some(IdForm::Legacy) {
        for f in &mut doc.findings {
            if f.severity == Severity::Error
                && !LEGACY_KEEPS.contains(&f.code)
                && !f.code.starts_with("closure_")
            {
                f.severity = Severity::Warning;
            }
        }
    }
    doc
}

/// The marker of an open question an author left in a SPEC.
const NEEDS_CLARIFICATION: &str = "[NEEDS CLARIFICATION";
/// The size budget (B3): above these a SPEC is too big for one PR; a warning, never an error.
pub const MAX_REQUIREMENTS: usize = 12;
pub const MAX_LINES: usize = 400;

/// Every `<…>` placeholder of the built-in templates: text `spec new` writes for the author to
/// replace. Other `<…>` (such as `<state>` or `REQ-<n>`) is the author's own.
fn template_placeholders() -> Vec<&'static str> {
    let mut out: Vec<&'static str> = Vec::new();
    for t in [FULL_TEMPLATE, MINIMAL_TEMPLATE] {
        let mut rest = t;
        while let Some(open) = rest.find('<') {
            let Some(len) = rest[open..].find(['>', '\n']) else {
                break;
            };
            let span = &rest[open..open + len + 1];
            if span.ends_with('>') && !out.contains(&span) {
                out.push(span);
            }
            rest = &rest[open + 1..];
        }
    }
    out
}

/// B3: `[NEEDS CLARIFICATION…]` and unfilled template placeholders, outside fenced code. An
/// error once the SPEC is approved or completed (it would ship an open question), a warning
/// before.
fn unfinished(lines: &[&str], doc: &mut SpecDoc) {
    let severity = if matches!(doc.status_as.as_deref(), Some("approved" | "completed")) {
        Severity::Error
    } else {
        Severity::Warning
    };
    let placeholders = template_placeholders();
    let mut fence: Option<&str> = None;
    for (i, line) in lines.iter().enumerate() {
        let t = line.trim_start();
        if let Some(mark) = fence {
            if t.starts_with(mark) {
                fence = None;
            }
            continue;
        }
        if t.starts_with("```") || t.starts_with("~~~") {
            fence = Some(&t[..3]);
            continue;
        }
        if line.contains(NEEDS_CLARIFICATION) {
            doc.push(
                "needs_clarification",
                severity,
                Some(i + 1),
                "an open `[NEEDS CLARIFICATION …]` question; answer it before approval",
            );
        }
        if placeholders.iter().any(|p| line.contains(p)) {
            doc.push(
                "placeholder",
                severity,
                Some(i + 1),
                "template text `<…>` still to fill; fill it before approval",
            );
        }
    }
}

/// B3: the size budget, a warning above [`MAX_REQUIREMENTS`] requirements or [`MAX_LINES`]
/// lines (split the work into SPECs linked by `parent`, decision U-D).
fn budget(line_count: usize, doc: &mut SpecDoc) {
    if doc.requirements.len() > MAX_REQUIREMENTS {
        doc.warn(
            "size_requirements",
            None,
            format!(
                "more than {MAX_REQUIREMENTS} requirements; split the work into SPECs linked by \
                 `parent`"
            ),
        );
    }
    if line_count > MAX_LINES {
        doc.warn(
            "size_lines",
            None,
            format!("more than {MAX_LINES} lines; split the work into SPECs linked by `parent`"),
        );
    }
}

/// Reads and checks the frontmatter; returns the index of the first body line and the line of
/// `approval:`.
fn frontmatter(lines: &[&str], doc: &mut SpecDoc, opts: &Options) -> (usize, Option<usize>) {
    if lines.first().map(|l| l.trim_end()) != Some("---") {
        doc.error(
            "fm_missing",
            Some(1),
            "spec.md must start with a `---` frontmatter block",
        );
        return (0, None);
    }
    let Some(close) = lines
        .iter()
        .skip(1)
        .position(|l| l.trim_end() == "---")
        .map(|p| p + 1)
    else {
        doc.error(
            "fm_unclosed",
            Some(1),
            "the frontmatter block has no closing `---` line",
        );
        return (lines.len(), None);
    };
    let mut fields: Vec<(String, Field)> = Vec::new();
    let mut nulls: Vec<String> = Vec::new();
    let mut approval: Option<(usize, Vec<(String, Field)>)> = None;
    let mut current: Option<String> = None;
    let mut shaped: Vec<String> = Vec::new();
    for (i, raw) in lines[1..close].iter().enumerate() {
        let n = i + 2;
        let t = raw.trim();
        if t.is_empty() || t.starts_with('#') {
            continue;
        }
        let nested = raw.starts_with([' ', '\t']) || raw.starts_with("- ");
        if nested {
            match current.as_deref() {
                Some("approval") => {
                    let Some((k, v)) = key_value(t) else {
                        doc.error("fm_line", Some(n), "an `approval` line is not `key: value`");
                        continue;
                    };
                    let list = &mut approval.get_or_insert((n, Vec::new())).1;
                    add_approval(doc, list, k, v, n);
                }
                Some(k) if KNOWN_KEYS.contains(&k) => {
                    if !shaped.iter().any(|s| s == k) {
                        shaped.push(k.to_string());
                        doc.error(
                            "fm_shape",
                            Some(n),
                            format!("`{k}` takes one value on its own line, not nested lines"),
                        );
                    }
                }
                Some(_) => {}
                None => doc.error("fm_line", Some(n), "an indented line has no key above it"),
            }
            continue;
        }
        let Some((k, v)) = key_value(raw) else {
            doc.error("fm_line", Some(n), "a frontmatter line is not `key: value`");
            current = None;
            continue;
        };
        current = Some(k.to_string());
        if fields.iter().any(|(seen, _)| seen == k) {
            doc.error("fm_duplicate", Some(n), format!("`{k}` is given twice"));
            continue;
        }
        if k == "approval" && !is_null(v) {
            let list = &mut approval.get_or_insert((n, Vec::new())).1;
            if !v.is_empty() {
                match flow_map(v) {
                    Some(pairs) => {
                        for (pk, pv) in pairs {
                            add_approval(doc, list, &pk, &pv, n);
                        }
                    }
                    None => doc.error(
                        "fm_shape",
                        Some(n),
                        "`approval` holds `by`, `at`, `basis` and `digest`: indented lines or \
                         one `{…}` map",
                    ),
                }
            }
        } else if KNOWN_KEYS.contains(&k) && (v.starts_with('{') || v.starts_with('[')) {
            doc.error(
                "fm_shape",
                Some(n),
                format!("`{k}` takes one value, not a map or a list"),
            );
        }
        if is_null(v) {
            // `key: null` reads as absent; the key still counts for duplicates.
            fields.push((
                k.to_string(),
                Field {
                    value: String::new(),
                    line: n,
                },
            ));
            nulls.push(k.to_string());
            continue;
        }
        fields.push((
            k.to_string(),
            Field {
                value: unquote(v),
                line: n,
            },
        ));
    }
    let get = |k: &str| {
        fields
            .iter()
            .find(|(key, _)| key == k && !nulls.contains(key))
            .map(|(_, f)| f)
    };
    for (k, f) in &fields {
        if k == "cards" {
            doc.error(
                "fm_cards",
                Some(f.line),
                "`cards` is refused: personal card numbers stay on the card, not in a shared SPEC",
            );
        } else if !KNOWN_KEYS.contains(&k.as_str()) {
            doc.unknown_keys.push(k.clone());
            doc.warn(
                "fm_unknown_key",
                Some(f.line),
                format!("`{k}` is not a SPEC key; it is kept and not read"),
            );
        }
    }
    // id
    match get("id") {
        None => doc.error("id_missing", Some(1), "the frontmatter has no `id`"),
        Some(f) if f.value.is_empty() => doc.error("id_missing", Some(f.line), "`id` is empty"),
        Some(f) => {
            let id = f.value.clone();
            doc.id_form = id_form(&id);
            match doc.id_form {
                None if bad_date_id(&id) => doc.error(
                    "id_date",
                    Some(f.line),
                    "the id's date is not a real calendar date",
                ),
                None => doc.error(
                    "id_form",
                    Some(f.line),
                    "the id is not SPEC-<yyyymmdd>-<slug> (or a Go-generation SPEC-…-NNN id)",
                ),
                Some(IdForm::Legacy) => doc.warn(
                    "id_legacy",
                    Some(f.line),
                    "a Go-generation id: read only, checked leniently, never written",
                ),
                Some(IdForm::New) => {}
            }
            if let Some(folder) = &opts.folder {
                if *folder != id {
                    doc.error(
                        "id_folder",
                        Some(f.line),
                        "the id differs from the SPEC folder's name",
                    );
                }
            }
            doc.id = Some(id);
        }
    }
    // title
    match get("title") {
        Some(f) if !f.value.trim().is_empty() => {
            if f.value.chars().count() > MAX_TITLE_CHARS {
                doc.error(
                    "title_long",
                    Some(f.line),
                    format!("the title is longer than {MAX_TITLE_CHARS} characters"),
                );
            }
            doc.title = Some(f.value.clone());
        }
        f => doc.error(
            "title_missing",
            f.map(|f| f.line).or(Some(1)),
            "the frontmatter has no `title`",
        ),
    }
    // status
    match get("status") {
        Some(f) if !f.value.is_empty() => {
            let s = f.value.as_str();
            doc.status = Some(s.to_string());
            if STATUSES.contains(&s) {
                doc.status_as = Some(s.to_string());
            } else if let Some(new) = legacy_status(s) {
                doc.status_as = Some(new.to_string());
                doc.status_legacy = true;
                let message = format!("a Go-generation status, read as `{new}`; never written");
                if doc.id_form == Some(IdForm::Legacy) {
                    doc.warn("status_legacy", Some(f.line), message);
                } else {
                    doc.error("status_legacy", Some(f.line), message);
                }
            } else {
                doc.error(
                    "status_unknown",
                    Some(f.line),
                    format!("the status is not one of {}", STATUSES.join(", ")),
                );
            }
        }
        f => doc.error(
            "status_missing",
            f.map(|f| f.line).or(Some(1)),
            "the frontmatter has no `status`",
        ),
    }
    // revision
    match get("revision") {
        Some(f) => {
            let v = f.value.as_str();
            let ok = !v.starts_with('0') && v.bytes().all(|c| c.is_ascii_digit());
            match v.parse::<u32>() {
                Ok(r) if ok && r >= 1 => doc.revision = Some(r),
                _ => doc.error(
                    "revision_invalid",
                    Some(f.line),
                    "`revision` must be a whole number from 1",
                ),
            }
        }
        None => doc.error(
            "revision_invalid",
            Some(1),
            "the frontmatter has no `revision`",
        ),
    }
    // created
    match get("created") {
        Some(f) if valid_date(&f.value) => doc.created = Some(f.value.clone()),
        Some(f) => doc.error(
            "created_invalid",
            Some(f.line),
            "`created` must be a real calendar date written YYYY-MM-DD",
        ),
        None => doc.error(
            "created_invalid",
            Some(1),
            "the frontmatter has no `created`",
        ),
    }
    // parent
    if let Some(f) = get("parent") {
        if id_form(&f.value).is_none() {
            doc.error("parent_invalid", Some(f.line), "`parent` must be a SPEC id");
        } else if doc.id.as_deref() == Some(f.value.as_str()) {
            doc.error(
                "parent_self",
                Some(f.line),
                "a SPEC cannot be its own parent",
            );
        }
        doc.parent = Some(f.value.clone());
    }
    // approval
    if let Some((line, list)) = approval {
        let mut a = Approval::default();
        for (k, f) in &list {
            match k.as_str() {
                "by" => a.by = Some(f.value.clone()),
                "at" => a.at = Some(f.value.clone()),
                "basis" => a.basis = Some(f.value.clone()),
                "digest" => a.digest = Some(f.value.clone()),
                other => {
                    doc.unknown_keys.push(format!("approval.{other}"));
                    doc.warn(
                        "fm_unknown_key",
                        Some(f.line),
                        format!("`approval.{other}` is not an approval key"),
                    );
                }
            }
        }
        for k in APPROVAL_KEYS {
            let v = match k {
                "by" => &a.by,
                "at" => &a.at,
                "basis" => &a.basis,
                _ => &a.digest,
            };
            let good = v
                .as_deref()
                .is_some_and(|v| !v.trim().is_empty() && (k != "at" || valid_date(v)));
            if !good {
                doc.error(
                    "approval_field",
                    Some(line),
                    format!("`approval.{k}` is missing or empty (`at` is a date, YYYY-MM-DD)"),
                );
            }
        }
        doc.approval = Some(a);
        return (close + 1, Some(line));
    } else if doc
        .status_as
        .as_deref()
        .is_some_and(|s| NEEDS_APPROVAL.contains(&s))
    {
        doc.error(
            "approval_missing",
            get("status").map(|f| f.line),
            "an approved, completed or superseded SPEC needs `approval` with `by`, `at`, `basis` \
             and `digest`",
        );
    }
    (close + 1, None)
}

/// `REQ-<n>` or `AC-<n>` at the start of a line, after an optional list marker and `**`:
/// the number and the rest of the line.
fn item(line: &str, prefix: &str) -> Option<(u32, String)> {
    let mut t = line.trim_start();
    for marker in ["- ", "* ", "+ "] {
        if let Some(rest) = t.strip_prefix(marker) {
            t = rest.trim_start();
            break;
        }
    }
    let t = t.strip_prefix("**").unwrap_or(t);
    let rest = t.strip_prefix(prefix)?;
    let end = rest
        .find(|c: char| !c.is_ascii_digit())
        .unwrap_or(rest.len());
    if end == 0 {
        return None;
    }
    let after = &rest[end..];
    if !(after.is_empty() || after.starts_with([':', ' ', '.', ')', '*'])) {
        return None;
    }
    let n = rest[..end].parse().ok()?;
    let text = after
        .trim_start_matches(['*', ':', '.', ')'])
        .trim()
        .to_string();
    Some((n, text))
}

/// A line without its list marker (`- `, `* `, `+ `), if it has one.
fn unlisted(line: &str) -> (bool, &str) {
    let t = line.trim_start();
    ["- ", "* ", "+ "]
        .iter()
        .find_map(|m| t.strip_prefix(m))
        .map_or((false, t), |rest| (true, rest.trim_start()))
}

/// The verification tags of an AC.
const AC_TAGS: [&str; 3] = ["verify:", "manual:", "unknown:"];

/// The first verification tag in `text` that starts the text or follows a space or `·`: the
/// text before it (without a trailing `·`), the tag and its value (one `` `…` `` span
/// unwrapped). A tag with no value is not one.
fn split_tag(text: &str) -> (&str, Option<(&'static str, String)>) {
    let mut best: Option<(usize, &'static str)> = None;
    for tag in AC_TAGS {
        for (at, _) in text.match_indices(tag) {
            let ok = text[..at]
                .chars()
                .next_back()
                .is_none_or(|c| c.is_whitespace() || c == '·');
            if ok {
                if best.is_none_or(|(b, _)| at < b) {
                    best = Some((at, tag));
                }
                break;
            }
        }
    }
    let Some((at, tag)) = best else {
        return (text, None);
    };
    let value = text[at + tag.len()..].trim();
    let value = value
        .strip_prefix('`')
        .and_then(|v| v.strip_suffix('`'))
        .filter(|v| !v.contains('`'))
        .unwrap_or(value)
        .trim();
    let before = text[..at].trim_end().trim_end_matches('·').trim_end();
    if value.is_empty() {
        return (before, None);
    }
    (before, Some((tag, value.to_string())))
}

/// The scope kinds of a `## Scope` line: `write:`, `read-only:` and `forbidden:`.
const SCOPE_TAGS: [&str; 3] = ["write:", "read-only:", "forbidden:"];

/// The paths of a `## Scope` text: every `` `…` `` span, or else each comma-separated item's
/// first word.
fn scope_paths(text: &str) -> Vec<String> {
    let spans: Vec<String> = text
        .split('`')
        .skip(1)
        .step_by(2)
        .map(str::trim)
        .filter(|p| !p.is_empty())
        .map(str::to_string)
        .collect();
    if !spans.is_empty() || text.contains('`') {
        return spans;
    }
    text.split(',')
        .filter_map(|item| item.split_whitespace().next())
        .map(str::to_string)
        .collect()
}

fn absolute(p: &str) -> bool {
    let b = p.as_bytes();
    p.starts_with(['/', '\\', '~']) || (b.len() >= 2 && b[0].is_ascii_alphabetic() && b[1] == b':')
}

/// The numbered lines of one section.
type Numbered<'a> = Vec<(usize, &'a str)>;

/// Reads and checks the `##` sections from line index `start`.
fn body(lines: &[&str], start: usize, doc: &mut SpecDoc) {
    // (name, heading line, content lines)
    let mut sections: Vec<(String, usize, Numbered)> = Vec::new();
    let mut current: Option<usize> = None;
    let mut fence: Option<&str> = None;
    for (i, line) in lines.iter().enumerate().skip(start) {
        let n = i + 1;
        let t = line.trim_start();
        if let Some(mark) = fence {
            if t.starts_with(mark) {
                fence = None;
            }
        } else if t.starts_with("```") || t.starts_with("~~~") {
            fence = Some(&t[..3]);
        } else if let Some(name) = line.strip_prefix("## ") {
            let name = name.trim().to_string();
            if !SECTIONS.contains(&name.as_str()) {
                doc.warn(
                    "section_unknown",
                    Some(n),
                    "a `##` section this format does not name; it is kept and not read",
                );
                current = None;
                continue;
            }
            if sections.iter().any(|(s, _, _)| *s == name) {
                doc.error(
                    "section_duplicate",
                    Some(n),
                    format!("`## {name}` appears twice"),
                );
                current = None;
                continue;
            }
            sections.push((name, n, Vec::new()));
            current = Some(sections.len() - 1);
            continue;
        }
        if let Some(c) = current {
            sections[c].2.push((n, line));
        }
    }
    doc.sections = sections.iter().map(|(s, _, _)| s.clone()).collect();
    let has = |name: &str| sections.iter().any(|(s, _, _)| s == name);
    if FULL_ONLY.iter().any(|s| has(s)) {
        doc.form = Form::Full;
    }
    let mut needed: Vec<&str> = REQUIRED.to_vec();
    if doc.form == Form::Full {
        needed.extend(FULL_ONLY);
    }
    for name in needed {
        if !has(name) {
            doc.error(
                "section_missing",
                None,
                format!("the `## {name}` section is missing"),
            );
        }
    }
    for (name, at, content) in &sections {
        let blank = content.iter().all(|(_, l)| l.trim().is_empty());
        match name.as_str() {
            "Intent" | "Plan" | "Non-goals" | "Risks" if blank => doc.error(
                "section_empty",
                Some(*at),
                format!("the `## {name}` section is empty"),
            ),
            "Requirements" => requirements(doc, *at, content),
            "Acceptance" => acceptance(doc, *at, content),
            "Scope" => scope(doc, *at, content),
            "Result" => result(doc, content),
            _ => {}
        }
    }
}

/// `(REQ-1, REQ-2)` right after an item's id: the ids named and the rest of the line after
/// the `)` and an optional `:`.
fn refs_of(text: &str) -> (Vec<String>, &str) {
    match text.strip_prefix('(').and_then(|r| r.split_once(')')) {
        Some((inside, rest)) => (
            inside
                .split(',')
                .map(str::trim)
                .filter(|r| !r.is_empty())
                .flat_map(expand_range)
                .collect(),
            rest.trim_start_matches(':').trim(),
        ),
        None => (Vec::new(), text),
    }
}

/// The most ids one range may name.
const MAX_RANGE: u32 = 1000;

/// `REQ-1…REQ-3` or `REQ-1..REQ-3` as `REQ-1`, `REQ-2`, `REQ-3` (B2). Anything else — one id,
/// a backward range, mixed prefixes, more than [`MAX_RANGE`] ids — stays one literal ref, so
/// a ref that names nothing real is reported as unknown.
fn expand_range(r: &str) -> Vec<String> {
    let num = |s: &str, p: &str| -> Option<u32> {
        let d = s.trim().strip_prefix(p)?;
        (!d.is_empty() && d.bytes().all(|c| c.is_ascii_digit()))
            .then(|| d.parse().ok())
            .flatten()
    };
    let split = r.split_once('…').or_else(|| r.split_once(".."));
    if let Some((a, b)) = split {
        for p in ["REQ-", "AC-"] {
            if let (Some(x), Some(y)) = (num(a, p), num(b, p)) {
                if x <= y && y - x < MAX_RANGE {
                    return (x..=y).map(|n| format!("{p}{n}")).collect();
                }
            }
        }
    }
    vec![r.to_string()]
}

/// The note that a `When` requirement has no failure path to cover.
const FAILURE_NONE: &str = "failure: none";

/// A requirement sentence without its `failure: none` note, and whether it had one. The note
/// is only the trailing `· failure: none`, optionally followed by ` — <reason>` (N5); the same
/// words anywhere else are part of the sentence.
fn failure_note(text: &str) -> (&str, bool) {
    for (at, _) in text.match_indices('·') {
        let rest = text[at + '·'.len_utf8()..].trim_start();
        if let Some(after) = rest.strip_prefix(FAILURE_NONE) {
            let after = after.trim();
            if after.is_empty() || after.starts_with('—') {
                return (text[..at].trim_end(), true);
            }
        }
    }
    (text, false)
}

/// `text` starts with `word` (ASCII case ignored) followed by a space.
fn starts_with_word(text: &str, word: &str) -> bool {
    text.get(..word.len())
        .is_some_and(|h| h.eq_ignore_ascii_case(word))
        && text[word.len()..].starts_with(' ')
}

/// A GEARS sentence, or the EARS unwanted-behaviour form `IF <condition>, THEN <subject> shall
/// <behavior>` (the comma is optional; `IF` and `THEN` ignore ASCII case). The unwanted form is
/// checked by `spec::parse_gears` as `When <condition>, <subject> shall <behavior>`, so it gets
/// the same text, subject and `shall` rules; its record keeps the original text, syntax
/// `GEARS-unwanted` and the condition kind `if`. The error names the rule, never the text.
fn parse_requirement(
    id: &str,
    sentence: &str,
    kind: ReqKind,
) -> Result<crate::spec::Requirement, String> {
    if kind == ReqKind::Gears {
        return crate::spec::parse_gears(id, sentence).map_err(|e| e.reason);
    }
    let body = &sentence[2..];
    let lower = body.to_ascii_lowercase();
    let then = lower.match_indices("then").map(|(i, _)| i).find(|&i| {
        let before = lower[..i].chars().next_back();
        let after = lower[i + 4..].chars().next();
        before.is_some_and(|c| c == ' ' || c == ',') && after.is_some_and(|c| c == ' ')
    });
    let Some(then) = then else {
        return Err("an IF requirement needs THEN before its subject".into());
    };
    let condition = body[..then].trim().trim_end_matches(',').trim();
    if condition.is_empty() {
        return Err("empty condition after IF".into());
    }
    let rest = body[then + 4..].trim();
    let mut r =
        crate::spec::parse_gears(id, &format!("When {condition}, {rest}")).map_err(|e| e.reason)?;
    if r.condition_kinds.len() != 1 {
        return Err("an IF requirement takes no Where, While or When clause after THEN".into());
    }
    r.text = sentence.to_string();
    r.syntax = "GEARS-unwanted".into();
    r.condition_kinds = vec!["if".into()];
    Ok(r)
}

fn requirements(doc: &mut SpecDoc, at: usize, content: &[(usize, &str)]) {
    for (n, line) in content {
        if let Some((num, text)) = item(line, "REQ-") {
            if doc.requirements.iter().any(|r| r.number == num) {
                doc.error(
                    "req_duplicate",
                    Some(*n),
                    format!("REQ-{num} appears twice"),
                );
                continue;
            }
            let id = format!("REQ-{num}");
            let (refs, rest) = refs_of(&text);
            let (sentence, failure_none) = failure_note(rest);
            let kind = if starts_with_word(sentence, "if") {
                ReqKind::Unwanted
            } else {
                ReqKind::Gears
            };
            let mut trigger = false;
            match parse_requirement(&id, sentence, kind) {
                Ok(parsed) => {
                    trigger = kind == ReqKind::Gears && parsed.when.is_some();
                    doc.gears.push(parsed);
                }
                // The reason names the rule, never the text.
                Err(reason) => doc.error(
                    "req_gears",
                    Some(*n),
                    format!("{id} is not a GEARS or IF/THEN sentence: {reason}"),
                ),
            }
            doc.requirements.push(Requirement {
                id,
                number: num,
                line: *n,
                kind,
                refs,
                trigger,
                failure_none,
                text: sentence.to_string(),
            });
        }
    }
    if doc.requirements.is_empty() {
        doc.error(
            "req_missing",
            Some(at),
            "`## Requirements` has no `REQ-<n>` line",
        );
    }
}

fn acceptance(doc: &mut SpecDoc, at: usize, content: &[(usize, &str)]) {
    let mut open: Option<Criterion> = None;
    let mut done: Vec<Criterion> = Vec::new();
    for (n, line) in content {
        if let Some((num, text)) = item(line, "AC-") {
            done.extend(open.take());
            // `(REQ-1, REQ-2)` right after the id names the requirements it checks.
            let (refs, rest) = refs_of(&text);
            let (before, tag) = split_tag(rest);
            let mut c = Criterion {
                id: format!("AC-{num}"),
                number: num,
                line: *n,
                refs,
                text: before.to_string(),
                verify: None,
                manual: None,
                unknown: None,
            };
            set_tag(&mut c, tag);
            open = Some(c);
            continue;
        }
        if let Some(c) = open.as_mut() {
            let (_, t) = unlisted(line);
            if c.verify.is_none() && c.manual.is_none() && c.unknown.is_none() {
                if let (before, Some(tag)) = split_tag(t) {
                    // Only a line that starts with the tag continues the AC.
                    if before.is_empty() {
                        set_tag(c, Some(tag));
                    }
                }
            }
        }
    }
    done.extend(open);
    for c in done {
        if doc.acceptance.iter().any(|a| a.number == c.number) {
            doc.error(
                "ac_duplicate",
                Some(c.line),
                format!("{} appears twice", c.id),
            );
            continue;
        }
        // The text checks `parse_gears` makes on requirements (N2); the reason names the rule,
        // never the text.
        let refused = [
            Some(&c.text),
            c.verify.as_ref(),
            c.manual.as_ref(),
            c.unknown.as_ref(),
        ]
        .into_iter()
        .flatten()
        .find_map(|t| crate::spec::check_text(t).err());
        if let Some(reason) = refused {
            doc.error(
                "text_unsafe",
                Some(c.line),
                format!("{} is refused: {reason}", c.id),
            );
        }
        if c.verify.is_none() && c.manual.is_none() && c.unknown.is_none() {
            doc.error(
                "ac_no_verify",
                Some(c.line),
                format!(
                    "{} has no `verify:` command, `manual:` reason or `unknown:` reason",
                    c.id
                ),
            );
        } else if c.unknown.is_some() {
            doc.warn(
                "ac_unknown",
                Some(c.line),
                format!("{} has no way to verify it yet (`unknown:`)", c.id),
            );
        }
        doc.acceptance.push(c);
    }
    if doc.acceptance.is_empty() {
        doc.error(
            "ac_missing",
            Some(at),
            "`## Acceptance` has no `AC-<n>` line",
        );
    }
}

fn set_tag(c: &mut Criterion, tag: Option<(&'static str, String)>) {
    match tag {
        Some(("verify:", v)) => c.verify = Some(v),
        Some(("manual:", v)) => c.manual = Some(v),
        Some((_, v)) => c.unknown = Some(v),
        None => {}
    }
}

fn scope(doc: &mut SpecDoc, at: usize, content: &[(usize, &str)]) {
    // A tag line without paths (`write:`) sets the kind of the untagged list items below it;
    // before any such line an untagged item is `write`. A tag with paths names only its own.
    let mut heading = "write:";
    // A tag line with paths may wrap: indented lines below it that are not list items continue
    // it, until a list item or a blank line.
    let mut wrap: Option<&str> = None;
    for (n, line) in content {
        let (listed, t) = unlisted(line);
        let tag = SCOPE_TAGS.iter().find(|tag| t.starts_with(**tag));
        let (kind, paths) = match tag {
            Some(tag) => {
                let paths = scope_paths(&t[tag.len()..]);
                if paths.is_empty() {
                    heading = tag;
                    wrap = None;
                } else {
                    wrap = Some(tag);
                }
                (*tag, paths)
            }
            None if listed => {
                wrap = None;
                (heading, scope_paths(t).into_iter().take(1).collect())
            }
            None => match wrap {
                Some(kind) if line.starts_with([' ', '\t']) && !t.is_empty() => {
                    (kind, scope_paths(t))
                }
                _ => {
                    wrap = None;
                    continue;
                }
            },
        };
        for path in paths {
            if absolute(&path) {
                doc.error(
                    "scope_absolute",
                    Some(*n),
                    "a scope path must be relative to the repository",
                );
            } else if path.split(['/', '\\']).any(|seg| seg == "..") {
                doc.error(
                    "scope_parent",
                    Some(*n),
                    "a scope path must not climb with `..`",
                );
            } else if let Err(reason) = crate::spec::check_text(&path) {
                doc.error(
                    "text_unsafe",
                    Some(*n),
                    format!("a scope path is refused: {reason}"),
                );
            }
            match kind {
                "write:" => doc.scope.write.push(path),
                "read-only:" => doc.scope.read_only.push(path),
                _ => doc.scope.forbidden.push(path),
            }
        }
    }
    if doc.scope.write.is_empty() {
        doc.error("scope_empty", Some(at), "`## Scope` lists no `write:` path");
    }
}

/// The failure-path rule (user decision 2026-11-17): every GEARS requirement with a `When`
/// trigger is named by a parsed IF/THEN requirement of the same SPEC, or says `failure: none`.
/// An error, not a warning: a trigger without its failure path is a requirement nobody wrote,
/// so no AC can check it, and the exemption costs one explicit note. An IF/THEN requirement that
/// names a REQ the SPEC lacks is `req_ref_unknown`.
fn failure_paths(doc: &mut SpecDoc) {
    let parsed = |id: &str| doc.gears.iter().any(|g| g.id == id);
    let mut found = Vec::new();
    for r in doc
        .requirements
        .iter()
        .filter(|r| r.kind == ReqKind::Unwanted)
    {
        for to in &r.refs {
            if !doc.requirements.iter().any(|q| q.id == *to) {
                found.push((
                    "req_ref_unknown",
                    r.line,
                    format!("{} names {to}, which `## Requirements` lacks", r.id),
                ));
            }
        }
    }
    for r in doc
        .requirements
        .iter()
        .filter(|r| r.trigger && !r.failure_none)
    {
        let covered = doc
            .requirements
            .iter()
            .any(|u| u.kind == ReqKind::Unwanted && parsed(&u.id) && u.refs.contains(&r.id));
        if !covered {
            found.push((
                "req_failure_path",
                r.line,
                format!(
                    "{} has a `When` trigger and no failure path: add `REQ-<n> ({}): IF …, \
                     THEN … shall …` or end the line with `· {FAILURE_NONE}`",
                    r.id, r.id
                ),
            ));
        }
    }
    for (code, line, message) in found {
        doc.error(code, Some(line), message);
    }
}

/// The AC → REQ links, the failure-path rule and the content digest; an approved digest that no
/// longer matches is an error. The digest exists only when every requirement parses.
fn digest(doc: &mut SpecDoc, approval_line: Option<usize>) {
    let mut unknown = Vec::new();
    for c in &doc.acceptance {
        for r in &c.refs {
            if !doc.requirements.iter().any(|q| q.id == *r) {
                unknown.push((
                    c.line,
                    format!("{} names {r}, which `## Requirements` lacks", c.id),
                ));
            }
        }
    }
    for (line, message) in unknown {
        doc.error("ac_ref_unknown", Some(line), message);
    }
    failure_paths(doc);
    if doc.requirements.is_empty() || doc.gears.len() != doc.requirements.len() {
        return;
    }
    let mut requirements = doc.gears.clone();
    for r in &mut requirements {
        r.acceptance_refs = doc
            .acceptance
            .iter()
            .filter(|c| c.refs.contains(&r.id))
            .map(|c| c.id.clone())
            .collect();
    }
    let acceptance = doc
        .acceptance
        .iter()
        .map(|c| {
            let (kind, selector) = match (&c.verify, &c.manual, &c.unknown) {
                (Some(v), _, _) => ("verify", v.clone()),
                (_, Some(m), _) => ("manual", m.clone()),
                (_, _, Some(u)) => ("unknown", u.clone()),
                _ => ("none", String::new()),
            };
            crate::spec::Acceptance {
                id: c.id.clone(),
                observable_result: c.text.clone(),
                verification: crate::spec::Verification {
                    kind: kind.into(),
                    selector,
                },
            }
        })
        .collect();
    let s = &doc.scope;
    let scope = [
        ("write", &s.write),
        ("read-only", &s.read_only),
        ("forbidden", &s.forbidden),
    ]
    .iter()
    .flat_map(|(k, paths)| paths.iter().map(move |p| format!("{k}:{p}")))
    .collect();
    let content = crate::spec::SpecDocument {
        spec_id: doc.id.clone().unwrap_or_default(),
        revision: doc.revision.unwrap_or(0),
        status: doc.status_as.clone().unwrap_or_default(),
        syntax: "GEARS".into(),
        source: None,
        requirements,
        acceptance,
        scope,
        policy_revision: None,
        verification_commands: doc
            .acceptance
            .iter()
            .filter_map(|c| c.verify.clone())
            .collect(),
        unresolved: Vec::new(),
        snapshot: None,
    };
    let digest = crate::spec::content_digest(&content, &crate::spec::SortedKeys);
    let recorded = doc.approval.as_ref().and_then(|a| a.digest.clone());
    if let Some(recorded) = recorded.filter(|r| !r.trim().is_empty()) {
        if recorded != digest {
            doc.error(
                "approval_digest",
                approval_line,
                "the requirements, acceptance or scope changed after approval (`approval.digest` \
                 differs); raise `revision` and approve again",
            );
        }
    }
    doc.digest = Some(digest);
}

fn result(doc: &mut SpecDoc, content: &[(usize, &str)]) {
    for (n, line) in content {
        if let Some((num, text)) = item(line, "AC-") {
            let outcome = text
                .split_whitespace()
                .next()
                .unwrap_or("")
                .trim_matches(|c: char| !c.is_ascii_alphabetic())
                .to_ascii_lowercase();
            doc.result.push(Outcome {
                ac: format!("AC-{num}"),
                number: num,
                outcome,
                line: *n,
                text,
            });
        }
    }
}

/// The closure check: status `completed` and every AC `met` or `withdrawn` in `## Result`.
fn closure(doc: &mut SpecDoc) {
    if doc.status_as.as_deref() != Some("completed") {
        doc.error(
            "closure_status",
            None,
            "a closed SPEC has status `completed`",
        );
    }
    if !doc.sections.iter().any(|s| s == "Result") {
        doc.error(
            "closure_result",
            None,
            "a closed SPEC has a `## Result` section",
        );
        return;
    }
    let mut findings = Vec::new();
    for c in &doc.acceptance {
        let closed = doc
            .result
            .iter()
            .any(|r| r.number == c.number && (r.outcome == "met" || r.outcome == "withdrawn"));
        if !closed {
            findings.push((
                "closure_open",
                Some(c.line),
                format!("{} is not `met` or `withdrawn` in `## Result`", c.id),
            ));
        }
    }
    for r in &doc.result {
        if !doc.acceptance.iter().any(|c| c.number == r.number) {
            findings.push((
                "closure_unknown",
                Some(r.line),
                format!("`## Result` names {}, which `## Acceptance` lacks", r.ac),
            ));
        }
    }
    for (code, line, message) in findings {
        doc.error(code, line, message);
    }
}
