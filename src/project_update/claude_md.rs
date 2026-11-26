//! The instruction files a Go-generation project carries (`claude-md-plan`): whether the root
//! `CLAUDE.md` is a Go-generation file, the sources in their fixed order with their facts, the
//! lines of each that are the project's own, those lines as hunks of one heading section each,
//! the texts the `--keep` picks compose, and the digest and command of that plan. Everything here
//! reads only: no file is written, a link is never followed and only read-only git commands run.
//! Hunk text and the composed texts are part of the answer; a line that holds a
//! credential-shaped value is counted and printed nowhere.

use super::PERMISSION_PLACEHOLDER;
use crate::go_residue::{find_manifest, Recorded};
use crate::go_state::linked_component;
use crate::{agents, agents_local_merge, claude, identity, state_root};
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    collections::{BTreeMap, HashMap, HashSet},
    fs,
    io::{self, BufRead, BufReader, ErrorKind, Read, Write},
    path::Path,
    process::{Command, Stdio},
};

/// A source above this many bytes is refused.
pub const SOURCE_LIMIT: u64 = 1024 * 1024;
/// The history search compares at most this many versions of one path, newest first.
pub const HISTORY_LIMIT: usize = 200;
/// A resulting file with more lines than this is warned about.
pub const RESULT_LINE_LIMIT: usize = 200;
/// The root instruction file Claude Code loads.
const CLAUDE_MD: &str = "CLAUDE.md";
/// The personal file Claude Code loads.
const CLAUDE_LOCAL_MD: &str = "CLAUDE.local.md";
/// The root sources in the order they are listed; the local rules files follow them.
const ROOT_SOURCES: [&str; 4] = [CLAUDE_MD, "AGENTS.md", "AGENTS.local.md", CLAUDE_LOCAL_MD];
/// The personal files: one that is exactly a first text of moai holds nothing the user wrote.
const PERSONAL: [&str; 2] = ["AGENTS.local.md", CLAUDE_LOCAL_MD];
/// The local rules folder, under the state folder.
const LOCAL_RULES: &str = "docs/local-rules";
/// Every managed block of this generation starts with this text.
const MANAGED_BEGIN: &str = "<!-- moai:managed:begin";
// @cc:careful `go_state` keeps its own copy of the start prefixes: a new Go curator marker goes
// into both places.
const CURATOR_STARTS: [&str; 3] = [
    "<!-- moai:learned-start",
    "<!-- moai:harness-start",
    "<!-- moai:learned-local-start",
];
const CURATOR_ENDS: [&str; 3] = [
    "<!-- moai:learned-end -->",
    "<!-- moai:harness-end -->",
    "<!-- moai:learned-local-end -->",
];

/// The other instruction file Claude Code may find; this step neither reads nor moves it.
const CONFIG_CLAUDE_MD: &str = ".claude/CLAUDE.md";
/// Why a picked line is in no resulting file.
const IMPORTS_MOVED: &str = "imports_a_moved_file";
const CURATOR_MARKER: &str = "go_curator_marker";

const IS_LINK: &str = "is a link";
const BELOW_LINK: &str = "lies below a link folder";
const UNREADABLE: &str = "cannot be read";

/// What shows that the root `CLAUDE.md` is a Go-generation instruction file.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum GoProof {
    /// It holds a Go curator marker.
    Marker,
    /// It has no managed block of moai and the found Go manifest records it.
    Manifest,
}

impl GoProof {
    pub fn word(self) -> &'static str {
        match self {
            GoProof::Marker => "marker",
            GoProof::Manifest => "manifest",
        }
    }
}

/// Whether git tracks a source.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Tracking {
    Tracked,
    Untracked,
    /// No git repository, or git could not list its files.
    Unknown,
}

impl Tracking {
    pub fn word(self) -> &'static str {
        match self {
            Tracking::Tracked => "tracked",
            Tracking::Untracked => "untracked",
            Tracking::Unknown => "unknown",
        }
    }
}

/// Why no earlier text could be compared, so every line counts.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum NoBase {
    /// No found Go manifest records the path.
    NoRecord,
    NotARepository,
    GitFailed,
    /// None of the versions searched has the recorded bytes.
    NoMatchingVersion,
}

impl NoBase {
    pub fn word(self) -> &'static str {
        match self {
            NoBase::NoRecord => "no_record",
            NoBase::NotARepository => "not_a_repository",
            NoBase::GitFailed => "git_failed",
            NoBase::NoMatchingVersion => "no_matching_version",
        }
    }
}

/// What a source's lines were compared with.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum Comparison {
    /// The bytes equal the Go manifest's record: no line is the project's own.
    ManifestEqual,
    /// A personal file that is exactly a first text of moai: no line is the project's own.
    MoaiFirstText,
    /// The version in git history with the recorded bytes: the lines it lacks count.
    GitHistory {
        blob: String,
    },
    AllLines(NoBase),
}

impl Comparison {
    fn to_json(&self) -> Value {
        match self {
            Comparison::ManifestEqual => json!({"used": "manifest_equal"}),
            Comparison::MoaiFirstText => json!({"used": "moai_first_text"}),
            Comparison::GitHistory { blob } => json!({"used": "git_history", "blob": blob}),
            Comparison::AllLines(why) => json!({"used": "all_lines", "reason": why.word()}),
        }
    }
}

/// One instruction file that can be read.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Source {
    /// Slash-separated path relative to the project.
    pub path: String,
    /// `sha256:<hex>` of the file's bytes.
    pub sha256: String,
    /// The whole text.
    pub text: String,
    pub git: Tracking,
    pub comparison: Comparison,
    /// The 1-based numbers of the project's own lines, ascending.
    pub project_lines: Vec<usize>,
}

impl Source {
    /// The source's row of the answer: facts and line numbers; its text appears in the hunks.
    fn to_json(&self) -> Value {
        json!({
            "path": self.path,
            "sha256": self.sha256,
            "line_count": self.text.lines().count(),
            "git": self.git.word(),
            "comparison": self.comparison.to_json(),
            "project_line_count": self.project_lines.len(),
            "project_lines": self.project_lines,
        })
    }
}

/// A source that must not be read, by name.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Refused {
    pub path: String,
    pub reason: &'static str,
}

/// One reading of the project's instruction files.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Preview {
    pub go_proof: Option<GoProof>,
    /// The found Go manifest, relative to the project.
    pub manifest: Option<String>,
    /// The readable sources: the root files in their order, then the local rules by path.
    pub sources: Vec<Source>,
    /// The sources that must not be read, in the same order.
    pub refused: Vec<Refused>,
}

/// Whether the root `CLAUDE.md` is a Go-generation instruction file: a regular file (never a
/// link) that holds a Go curator marker, or that has no managed block of moai while the found Go
/// manifest records `CLAUDE.md`. Reads only.
pub fn go_claude_md(project: &Path) -> Option<GoProof> {
    let path = project.join(CLAUDE_MD);
    if !fs::symlink_metadata(&path).ok()?.is_file() {
        return None;
    }
    let bytes = fs::read(&path).ok()?;
    let text = String::from_utf8_lossy(&bytes);
    if CURATOR_STARTS.iter().any(|marker| text.contains(marker)) {
        return Some(GoProof::Marker);
    }
    let recorded =
        || find_manifest(project).is_some_and(|(_, files)| files.contains_key(CLAUDE_MD));
    (!text.contains(MANAGED_BEGIN) && recorded()).then_some(GoProof::Manifest)
}

/// Reads the project's instruction files and compares each with its earlier text.
pub fn preview(project: &Path) -> Result<Preview, String> {
    if !fs::metadata(project).is_ok_and(|m| m.is_dir()) {
        return Err("the project is not a folder".into());
    }
    let (manifest, records) = match find_manifest(project) {
        Some((path, files)) => (Some(path), files),
        None => (None, BTreeMap::<String, Recorded>::new()),
    };
    let tracked = tracked_paths(project);
    let mut candidates: Vec<(String, Option<&'static str>)> = ROOT_SOURCES
        .iter()
        .map(|name| (name.to_string(), None))
        .collect();
    candidates.extend(local_rules(project));

    let mut sources = Vec::new();
    let mut refused = Vec::new();
    for (path, refusal) in candidates {
        let read = match refusal {
            Some(reason) => Err(reason),
            None => read_source(project, &path),
        };
        match read {
            Ok(Some(text)) => {
                let sha256 = digest(text.as_bytes());
                let record = records.get(&path).map(|r| r.deployed_hash.as_str());
                let (comparison, project_lines) = compare(project, &path, &text, &sha256, record);
                let git = match &tracked {
                    Some(set) if set.contains(&path) => Tracking::Tracked,
                    Some(_) => Tracking::Untracked,
                    None => Tracking::Unknown,
                };
                sources.push(Source {
                    path,
                    sha256,
                    text,
                    git,
                    comparison,
                    project_lines,
                });
            }
            Ok(None) => {}
            Err(reason) => refused.push(Refused { path, reason }),
        }
    }
    Ok(Preview {
        go_proof: go_claude_md(project),
        manifest,
        sources,
        refused,
    })
}

/// The answer of `claude-md-plan` for the `--keep` values `keep`. An error is a project that is
/// no folder or a `--keep` value that names no pick; anything else that stops the apply is listed
/// under `apply_blocked`, and then no digest and no command are printed.
pub fn plan(project: &Path, keep: &[String]) -> Result<Value, String> {
    let preview = preview(project)?;
    let root = fs::canonicalize(project).map_err(|_| crate::contract::PROJECT_PATH_UNAVAILABLE)?;
    let shown = root.to_str().ok_or("non UTF-8 project path")?;
    let readings: Vec<Reading> = preview
        .sources
        .iter()
        .map(|s| read_lines(&s.text))
        .collect();
    let moved = moved_paths(&preview.sources);
    let hunks = hunks(&preview.sources, &readings);
    let picks = parse_picks(keep, &hunks)?;
    // The picks as the command repeats them and the digest binds them: in one order.
    let mut words: Vec<(String, &str, Destination)> = picks
        .iter()
        .map(|(&i, &to)| {
            let id = hunks[i].id.as_str();
            (format!("{id}={}", to.word()), id, to)
        })
        .collect();
    words.sort();

    let mut blocked: Vec<String> = preview
        .refused
        .iter()
        .map(|r| format!("a source is refused: {} {}", r.path, r.reason))
        .collect();
    for (&i, &to) in &picks {
        let hunk = &hunks[i];
        let hidden = &readings[hunk.source].hidden;
        if to == Destination::Shared && hunk.lines.iter().any(|n| hidden[n - 1]) {
            blocked.push(format!(
                "{} holds a credential-shaped line and cannot be kept as shared, because \
                 {CLAUDE_MD} is shared through git; keep it as personal or leave it in the backup",
                hunk.id
            ));
        }
    }
    let fresh = fresh_texts(&root);
    let mut results = Vec::new();
    let mut profile_digest = None;
    match &fresh {
        Ok(fresh) => {
            let composer = Composer {
                sources: &preview.sources,
                readings: &readings,
                moved: &moved,
                hunks: &hunks,
                picks: &picks,
            };
            results.push(composer.file(CLAUDE_MD, &fresh.block, Destination::Shared));
            if picks.values().any(|to| *to == Destination::Personal) {
                results.push(composer.file(
                    CLAUDE_LOCAL_MD,
                    &fresh.personal,
                    Destination::Personal,
                ));
            }
            profile_digest = Some(digest(&fresh.profile));
        }
        Err(reason) => blocked.push(reason.clone()),
    }

    let mut warnings: Vec<String> = results
        .iter()
        .filter(|file| file.lines.len() > RESULT_LINE_LIMIT)
        .map(|file| {
            format!(
                "{} would have {} lines, more than {RESULT_LINE_LIMIT}",
                file.path,
                file.lines.len()
            )
        })
        .collect();
    // Only the entry is looked at: its text is no part of this step.
    if fs::symlink_metadata(root.join(CONFIG_CLAUDE_MD)).is_ok() {
        warnings.push(format!(
            "{CONFIG_CLAUDE_MD} exists; this step neither reads nor moves it"
        ));
    }

    let plan_digest = profile_digest
        .filter(|_| blocked.is_empty())
        .map(|profile| {
            let bound = json!({
                "operation": "claude_md.apply",
                "sources": preview.sources.iter().map(|s| [&s.path, &s.sha256]).collect::<Vec<_>>(),
                "profile": profile,
                "picks": words.iter().map(|(word, ..)| word).collect::<Vec<_>>(),
                "results": results
                    .iter()
                    .map(|file| [file.path.to_string(), digest(file.text().as_bytes())])
                    .collect::<Vec<_>>(),
            });
            digest(bound.to_string().as_bytes())
        });
    let apply = plan_digest.as_ref().map(|plan_digest| {
        let mut argv = vec!["moai", "claude-md-apply", shown];
        for (word, ..) in &words {
            argv.extend(["--keep", word.as_str()]);
        }
        argv.extend([
            "--plan-digest",
            plan_digest.as_str(),
            "--permission-reference",
            PERMISSION_PLACEHOLDER,
        ]);
        json!({"argv": argv, "stdin": null})
    });

    Ok(json!({
        "read_only": true,
        "go_claude_md": preview.go_proof.map(GoProof::word),
        "go_manifest": preview.manifest,
        "sources": preview.sources.iter().map(Source::to_json).collect::<Vec<_>>(),
        "refused": preview
            .refused
            .iter()
            .map(|r| json!({"path": r.path, "reason": r.reason}))
            .collect::<Vec<_>>(),
        "hunks": hunks
            .iter()
            .map(|h| h.to_json(&readings[h.source], &moved))
            .collect::<Vec<_>>(),
        "picks": words
            .iter()
            .map(|(_, id, to)| json!({"id": id, "destination": to.word()}))
            .collect::<Vec<_>>(),
        "results": results.iter().map(Resulting::to_json).collect::<Vec<_>>(),
        "warnings": warnings,
        "apply_blocked": blocked,
        "plan_digest": plan_digest,
        "apply": apply,
    }))
}

/// The managed block and the personal file's first text, with the profile bytes they came from.
struct Fresh {
    block: String,
    personal: String,
    profile: Vec<u8>,
}

/// [`Fresh`] for the project, or why its profile composes nothing.
fn fresh_texts(root: &Path) -> Result<Fresh, String> {
    let (profile, bytes) = crate::read_profile(root)
        .map_err(|e| format!("the project profile cannot be used ({e}); run /moai project"))?;
    if profile.is_placeholder() {
        return Err(
            "the project profile is still the placeholder a first launch writes; run \
             /moai project"
                .into(),
        );
    }
    let (block, personal) = agents::fresh_texts(&profile, root)?;
    Ok(Fresh {
        block,
        personal,
        profile: bytes,
    })
}

/// The fenced code block a reading of lines is inside: its mark and the length of its opening.
#[derive(Clone, Copy, Default)]
struct Fence(Option<(char, usize)>);

impl Fence {
    /// Reads the next line; `true` when it is code: a fence line, or a line inside a block.
    fn read(&mut self, line: &str) -> bool {
        let line = line.trim_start();
        let mark = line.chars().next().filter(|c| matches!(c, '`' | '~'));
        // Both marks are one byte, so the run's length is also where the rest begins.
        let run = mark.map_or(0, |mark| line.chars().take_while(|c| *c == mark).count());
        match (self.0, mark) {
            (Some((open, length)), _) => {
                if mark == Some(open) && run >= length && line[run..].trim().is_empty() {
                    self.0 = None;
                }
                true
            }
            // A backtick in what follows makes the run an inline code span, not a fence.
            (None, Some(mark)) if run >= 3 && !(mark == '`' && line[run..].contains('`')) => {
                self.0 = Some((mark, run));
                true
            }
            (None, _) => false,
        }
    }
}

/// A markdown heading line: at most three spaces, one to six `#`, then a blank or the end.
fn is_heading(line: &str) -> bool {
    let text = line.trim_start_matches(' ');
    let level = text.bytes().take_while(|b| *b == b'#').count();
    line.len() - text.len() <= 3
        && (1..=6).contains(&level)
        && matches!(text.as_bytes().get(level), None | Some(b' ' | b'\t'))
}

/// A URL whose user part before the host is shaped like a token and is no placeholder:
/// `scheme://<token>@host` holds the token even without the `:` a password follows.
fn url_user_part(line: &str) -> bool {
    line.match_indices("://").any(|(i, _)| {
        // The host part ends at the next `/`, which the next `://` also holds, so the parts of
        // one line never overlap and the scan stays linear.
        let authority = line[i + 3..]
            .split(|c: char| matches!(c, '/' | '?' | '#') || c.is_whitespace())
            .next()
            .unwrap_or_default();
        authority.rsplit_once('@').is_some_and(|(user, _)| {
            let secret = user.split_once(':').map_or(user, |(_, password)| password);
            // The shape the credential scan asks of an assigned value, so that the login name
            // of `ssh://git@host` is no secret.
            secret.len() >= 8
                && !secret.starts_with(['<', '$', '{', '*', '[', '('])
                && secret
                    .bytes()
                    .any(|b| b.is_ascii_digit() || b.is_ascii_uppercase())
        })
    })
}

/// For each line, whether its text must not be printed: a credential-shaped line, also when
/// invisible characters split the value, and every line of a private key from its begin line
/// through its end line, or through the last line.
fn hidden_lines(lines: &[&str]) -> Vec<bool> {
    let mut in_key = false;
    lines
        .iter()
        .map(|line| {
            let visible: String = line
                .chars()
                .filter(|c| !claude::invisible_char(*c))
                .collect();
            let shaped = [line, visible.as_str()]
                .into_iter()
                .any(|text| agents_local_merge::secret_kind(text).is_some() || url_user_part(text));
            let hidden = in_key || shaped;
            let lower = visible.to_ascii_lowercase();
            if lower.contains("private key") {
                match lower.rfind("-----begin") {
                    // A key may also end on the line it begins on.
                    Some(at) => in_key = !lower[at..].contains("-----end"),
                    None if lower.contains("-----end") => in_key = false,
                    None => {}
                }
            }
            hidden
        })
        .collect()
}

/// One source read line by line, over its whole text: what a line is can depend on lines that are
/// not the project's own.
struct Reading<'a> {
    /// The lines without their line ends, the first without a byte order mark.
    lines: Vec<&'a str>,
    /// A fence line or a line inside a fenced code block.
    code: Vec<bool>,
    hidden: Vec<bool>,
    /// The index of the heading line that opens the line's section; `None` before the first.
    section: Vec<Option<usize>>,
}

fn read_lines(text: &str) -> Reading<'_> {
    // Split first, as the line numbers of `compare` are: a file of the mark alone has one line.
    let mut lines: Vec<&str> = text.lines().collect();
    if let Some(first) = lines.first_mut() {
        *first = first.strip_prefix('\u{FEFF}').unwrap_or(first);
    }
    let mut fence = Fence::default();
    let mut open = None;
    let (mut code, mut section) = (Vec::new(), Vec::new());
    for (i, line) in lines.iter().enumerate() {
        let is_code = fence.read(line);
        if !is_code && is_heading(line) {
            open = Some(i);
        }
        code.push(is_code);
        section.push(open);
    }
    Reading {
        hidden: hidden_lines(&lines),
        lines,
        code,
        section,
    }
}

/// The paths the step moves, as text may name them: every readable source in its order, then the
/// local rules folder.
fn moved_paths(sources: &[Source]) -> Vec<String> {
    sources
        .iter()
        .map(|s| s.path.clone())
        .chain([format!("{}/{LOCAL_RULES}", state_root::MOAI_DIR)])
        .collect()
}

/// Whether `line` names `path` itself, written from the project root: not as the end of a longer
/// path or name, nor the start of one. A folder is also named by a path below it.
fn names_path(line: &str, path: &str, folder: bool) -> bool {
    // Only ASCII continues a name, so a word attached in another script ends it.
    let part = |c: char| c.is_ascii_alphanumeric() || matches!(c, '_' | '-');
    line.match_indices(path).any(|(i, _)| {
        let before = &line[..i];
        let before = before.strip_suffix("./").unwrap_or(before);
        let starts = !before
            .chars()
            .next_back()
            .is_some_and(|c| part(c) || matches!(c, '.' | '/'));
        let mut after = line[i + path.len()..].chars();
        let ends = match after.next() {
            Some('/') => folder,
            Some('.') => !after.next().is_some_and(part),
            next => !next.is_some_and(part),
        };
        starts && ends
    })
}

/// The project lines of one heading section of one source.
struct Hunk {
    /// `<path>:<first project line>`.
    id: String,
    path: String,
    /// The index of its source.
    source: usize,
    /// The index of the section's heading line; `None` before the first heading.
    heading: Option<usize>,
    /// The 1-based numbers of the lines, ascending.
    lines: Vec<usize>,
    /// The earlier hunk with the same lines.
    same_as: Option<String>,
}

impl Hunk {
    /// The row of the answer. A hunk equal to an earlier one shows neither text nor heading.
    fn to_json(&self, reading: &Reading, moved: &[String]) -> Value {
        let first_text = self.same_as.is_none();
        let heading = self
            .heading
            .filter(|&i| first_text && !reading.hidden[i])
            .map(|i| reading.lines[i]);
        let text = first_text.then(|| {
            self.lines
                .iter()
                .filter(|&&n| !reading.hidden[n - 1])
                .map(|&n| reading.lines[n - 1])
                .collect::<Vec<_>>()
                .join("\n")
        });
        // The last of the moved paths is the folder.
        let named: Vec<&String> = moved
            .iter()
            .enumerate()
            .filter(|(i, path)| {
                let folder = i + 1 == moved.len();
                self.lines
                    .iter()
                    .any(|&n| names_path(reading.lines[n - 1], path, folder))
            })
            .map(|(_, path)| path)
            .collect();
        json!({
            "id": self.id,
            "path": self.path,
            "heading": heading,
            "line_count": self.lines.len(),
            "text": text,
            "same_as": self.same_as,
            "hidden_lines": self.lines.iter().filter(|&&n| reading.hidden[n - 1]).count(),
            "moved_paths": named,
        })
    }
}

/// One hunk per section that holds a project line that is not blank, in source order.
fn hunks(sources: &[Source], readings: &[Reading]) -> Vec<Hunk> {
    let mut found: Vec<Hunk> = Vec::new();
    let mut first_with: HashMap<Vec<&str>, String> = HashMap::new();
    for (index, (source, reading)) in sources.iter().zip(readings).enumerate() {
        let sections = source
            .project_lines
            .chunk_by(|a, b| reading.section[a - 1] == reading.section[b - 1]);
        for lines in sections {
            let text: Vec<&str> = lines.iter().map(|&n| reading.lines[n - 1]).collect();
            if text.iter().all(|line| line.trim().is_empty()) {
                continue;
            }
            let id = format!("{}:{}", source.path, lines[0]);
            let same_as = first_with.get(&text).cloned();
            if same_as.is_none() {
                first_with.insert(text, id.clone());
            }
            found.push(Hunk {
                id,
                path: source.path.clone(),
                source: index,
                heading: reading.section[lines[0] - 1],
                lines: lines.to_vec(),
                same_as,
            });
        }
    }
    found
}

/// Where a picked hunk goes.
#[derive(Clone, Copy, Debug, PartialEq, Eq, PartialOrd, Ord)]
enum Destination {
    /// The new `CLAUDE.md`.
    Shared,
    /// The personal file `CLAUDE.local.md`.
    Personal,
}

impl Destination {
    fn word(self) -> &'static str {
        match self {
            Destination::Shared => "shared",
            Destination::Personal => "personal",
        }
    }
}

/// The `--keep <id>=shared|personal` values as hunk index and destination, so in source order. A
/// refused value is named through [`claude::shown_value`], so no message carries a credential.
fn parse_picks(keep: &[String], hunks: &[Hunk]) -> Result<BTreeMap<usize, Destination>, String> {
    let by_id: HashMap<&str, usize> = hunks
        .iter()
        .enumerate()
        .map(|(i, hunk)| (hunk.id.as_str(), i))
        .collect();
    let mut picks = BTreeMap::new();
    for value in keep {
        let parsed = value.rsplit_once('=').and_then(|(id, word)| {
            [Destination::Shared, Destination::Personal]
                .into_iter()
                .find(|to| to.word() == word)
                .map(|to| (id, to))
        });
        let Some((id, to)) = parsed else {
            return Err(format!(
                "--keep {}: the destination must be shared or personal \
                 (--keep <id>=shared|personal)",
                claude::shown_value(value)
            ));
        };
        let Some(&hunk) = by_id.get(id) else {
            return Err(format!(
                "--keep {}: no hunk has this id",
                claude::shown_value(id)
            ));
        };
        if picks.insert(hunk, to).is_some() {
            return Err(format!(
                "--keep {}: the id is given twice",
                claude::shown_value(id)
            ));
        }
    }
    Ok(picks)
}

/// `line` with every inline code span replaced by one backtick: a run of backticks through the
/// next run of the same length. A run with no such partner is text.
fn without_code_spans(line: &str) -> String {
    // (start, length) of each run of backticks.
    let mut runs: Vec<(usize, usize)> = Vec::new();
    for (i, _) in line.match_indices('`') {
        match runs.last_mut() {
            Some((start, length)) if *start + *length == i => *length += 1,
            _ => runs.push((i, 1)),
        }
    }
    let mut later: HashMap<usize, usize> = HashMap::new();
    let mut partner = vec![None; runs.len()];
    for (k, &(_, length)) in runs.iter().enumerate().rev() {
        partner[k] = later.insert(length, k);
    }
    let (mut out, mut from, mut k) = (String::new(), 0, 0);
    while k < runs.len() {
        if let Some(close) = partner[k] {
            out.push_str(&line[from..runs[k].0]);
            out.push('`');
            from = runs[close].0 + runs[close].1;
            k = close;
        }
        k += 1;
    }
    out.push_str(&line[from..]);
    out
}

/// The path `target` names from the folder `base`, both relative to the project; `None` for an
/// absolute or home path and for one that leaves the project.
fn resolve(base: &str, target: &str) -> Option<String> {
    if target.starts_with(['/', '~']) {
        return None;
    }
    let mut parts = Vec::new();
    for part in base.split('/').chain(target.split('/')) {
        match part {
            "" | "." => {}
            ".." => {
                parts.pop()?;
            }
            part => parts.push(part),
        }
    }
    Some(parts.join("/"))
}

/// Whether a line of the source `source` holds a word `@<path>` outside inline code that names a
/// moved path, read from the source's folder (what it imports now) or from the project root (what
/// it would import in a resulting file).
fn imports_moved(line: &str, source: &str, moved: &[String]) -> bool {
    let folder = source.rsplit_once('/').map_or("", |(folder, _)| folder);
    without_code_spans(line)
        .split_whitespace()
        .filter_map(|word| word.strip_prefix('@'))
        .any(|target| {
            ["", folder]
                .into_iter()
                .filter_map(|base| resolve(base, target))
                .any(|path| moved.contains(&path))
        })
}

/// A file the apply would write: its lines, each with whether it must not be printed.
struct Resulting<'a> {
    path: &'static str,
    lines: Vec<(&'a str, bool)>,
    /// The ids of the hunks picked for it, in source order.
    hunks: Vec<&'a str>,
    /// The `<path>:<line>` of each picked line it leaves out, and why.
    left_out: Vec<(String, &'static str)>,
}

impl Resulting<'_> {
    /// The text as it would be written.
    fn text(&self) -> String {
        self.lines
            .iter()
            .flat_map(|(line, _)| [line, "\n"])
            .collect()
    }

    /// The row of the answer: `text` is the file without its hidden lines; the line count and
    /// the digest are the whole file's.
    fn to_json(&self) -> Value {
        json!({
            "path": self.path,
            "line_count": self.lines.len(),
            "sha256": digest(self.text().as_bytes()),
            "text": self
                .lines
                .iter()
                .filter(|(_, hidden)| !hidden)
                .flat_map(|(line, _)| [line, "\n"])
                .collect::<String>(),
            "hidden_lines": self.lines.iter().filter(|(_, hidden)| *hidden).count(),
            "hunks": self.hunks,
            "left_out": self
                .left_out
                .iter()
                .map(|(line, reason)| json!({"line": line, "reason": reason}))
                .collect::<Vec<_>>(),
        })
    }
}

struct Composer<'a> {
    sources: &'a [Source],
    readings: &'a [Reading<'a>],
    moved: &'a [String],
    hunks: &'a [Hunk],
    picks: &'a BTreeMap<usize, Destination>,
}

impl<'a> Composer<'a> {
    /// `head` followed by the hunks picked for `to` in source order, one blank line before each.
    /// A line that imports a moved file (in its source, or where it would now stand) and a line
    /// that holds a Go curator start or end marker are left out; blank lines at a hunk's ends
    /// are dropped.
    fn file(&self, path: &'static str, head: &'a str, to: Destination) -> Resulting<'a> {
        let head: Vec<&str> = head.lines().collect();
        let mut fence = Fence::default();
        for line in &head {
            fence.read(line);
        }
        let hidden = hidden_lines(&head);
        let mut file = Resulting {
            path,
            lines: head.into_iter().zip(hidden).collect(),
            hunks: Vec::new(),
            left_out: Vec::new(),
        };
        let picked = self.picks.iter().filter(|(_, picked)| **picked == to);
        for hunk in picked.map(|(&i, _)| &self.hunks[i]) {
            let reading = &self.readings[hunk.source];
            let source = &self.sources[hunk.source].path;
            let mut kept: Vec<(&str, bool)> = Vec::new();
            for &n in &hunk.lines {
                let line = reading.lines[n - 1];
                let mut next = fence;
                let code_here = next.read(line);
                let mut markers = CURATOR_STARTS.iter().chain(&CURATOR_ENDS);
                let reason = if markers.any(|marker| line.contains(marker)) {
                    Some(CURATOR_MARKER)
                } else if !(reading.code[n - 1] && code_here)
                    && imports_moved(line, source, self.moved)
                {
                    Some(IMPORTS_MOVED)
                } else {
                    None
                };
                match reason {
                    Some(reason) => file.left_out.push((format!("{source}:{n}"), reason)),
                    None => {
                        fence = next;
                        kept.push((line, reading.hidden[n - 1]));
                    }
                }
            }
            file.hunks.push(&hunk.id);
            let blank = |(line, _): &(&str, bool)| line.trim().is_empty();
            if let Some(from) = kept.iter().position(|line| !blank(line)) {
                let end = kept.iter().rposition(|line| !blank(line)).unwrap_or(from);
                file.lines.push(("", false));
                file.lines.extend_from_slice(&kept[from..=end]);
            }
        }
        file
    }
}

fn digest(bytes: &[u8]) -> String {
    format!("sha256:{:x}", Sha256::digest(bytes))
}

fn is_absent(error: &io::Error) -> bool {
    matches!(error.kind(), ErrorKind::NotFound | ErrorKind::NotADirectory)
}

/// Whether a folder above `rel`, inside the project, is a link.
fn linked_parent(project: &Path, rel: &str) -> bool {
    rel.rsplit_once('/')
        .is_some_and(|(parent, _)| linked_component(project, parent))
}

/// Every entry under the local rules folder that is not a folder, by path, each with the reason
/// when it is already refused: a folder that is a link, lies below one or cannot be listed is
/// named itself and never entered.
fn local_rules(project: &Path) -> BTreeMap<String, Option<&'static str>> {
    let top = format!("{}/{LOCAL_RULES}", state_root::MOAI_DIR);
    let mut found = BTreeMap::new();
    let meta = match fs::symlink_metadata(project.join(&top)) {
        Ok(meta) => meta,
        Err(e) if is_absent(&e) => return found,
        Err(_) => {
            found.insert(top, Some(UNREADABLE));
            return found;
        }
    };
    if linked_parent(project, &top) {
        found.insert(top, Some(BELOW_LINK));
        return found;
    }
    if meta.file_type().is_symlink() {
        found.insert(top, Some(IS_LINK));
        return found;
    }
    if !meta.is_dir() {
        return found;
    }
    let mut folders = vec![top];
    while let Some(folder) = folders.pop() {
        let listed = fs::read_dir(project.join(&folder))
            .and_then(|entries| entries.collect::<io::Result<Vec<_>>>());
        let Ok(entries) = listed else {
            found.insert(folder, Some("cannot be listed"));
            continue;
        };
        for entry in entries {
            let name = entry.file_name();
            let Some(name) = name.to_str() else {
                let lossy = format!("{folder}/{}", name.to_string_lossy());
                found.insert(lossy, Some("has a name that is not UTF-8 text"));
                continue;
            };
            let child = format!("{folder}/{name}");
            // A link to a folder is not a folder here: it is refused when it is read.
            if fs::symlink_metadata(entry.path()).is_ok_and(|m| m.is_dir()) {
                folders.push(child);
            } else {
                found.insert(child, None);
            }
        }
    }
    found
}

/// One source's text, `Ok(None)` when there is no such file. Nothing is opened before the entry
/// is known to be a regular file reached through no link, so a pipe is never read.
fn read_source(project: &Path, rel: &str) -> Result<Option<String>, &'static str> {
    let path = project.join(rel);
    let meta = match fs::symlink_metadata(&path) {
        Ok(meta) => meta,
        Err(e) if is_absent(&e) => return Ok(None),
        Err(_) => return Err(UNREADABLE),
    };
    if linked_parent(project, rel) {
        return Err(BELOW_LINK);
    }
    if meta.file_type().is_symlink() {
        return Err(IS_LINK);
    }
    if !meta.is_file() {
        return Err("is not a regular file");
    }
    let mut bytes = Vec::new();
    fs::File::open(&path)
        .and_then(|file| file.take(SOURCE_LIMIT + 1).read_to_end(&mut bytes))
        .map_err(|_| UNREADABLE)?;
    if bytes.len() as u64 > SOURCE_LIMIT {
        return Err("is larger than 1 MiB");
    }
    String::from_utf8(bytes)
        .map(Some)
        .map_err(|_| "is not UTF-8 text")
}

/// What the source is compared with, and the numbers of its project lines. `record` is the
/// `sha256:<hex>` the found Go manifest holds for the path.
fn compare(
    project: &Path,
    path: &str,
    text: &str,
    sha256: &str,
    record: Option<&str>,
) -> (Comparison, Vec<usize>) {
    if record == Some(sha256) {
        return (Comparison::ManifestEqual, Vec::new());
    }
    if PERSONAL.contains(&path) && agents::is_generated_local(text.as_bytes()) {
        return (Comparison::MoaiFirstText, Vec::new());
    }
    let earlier = record
        .ok_or(NoBase::NoRecord)
        .and_then(|record| recorded_version(project, path, record));
    match earlier {
        Ok((blob, bytes)) => {
            let earlier = String::from_utf8_lossy(&bytes);
            let known = earlier.lines().collect();
            (Comparison::GitHistory { blob }, project_lines(text, &known))
        }
        Err(why) => (
            Comparison::AllLines(why),
            project_lines(text, &HashSet::new()),
        ),
    }
}

/// The 1-based numbers of the lines that are absent from `known`, outside the `moai-workflow`
/// managed block and no Go curator marker line.
fn project_lines(text: &str, known: &HashSet<&str>) -> Vec<usize> {
    let lines: Vec<&str> = text.lines().collect();
    let block = managed_block_lines(&lines);
    lines
        .iter()
        .enumerate()
        .filter(|(i, line)| {
            !block.contains(i) && !is_curator_marker_line(line) && !known.contains(*line)
        })
        .map(|(i, _)| i + 1)
        .collect()
}

/// The indexes of the lines of every complete `moai-workflow` block, its marker lines included.
/// A begin line with no end line before the next begin opens no block, so its lines count.
fn managed_block_lines(lines: &[&str]) -> HashSet<usize> {
    let (begin, end) = agents::markers(agents::AGENTS_SECTION, "html");
    let mut inside = HashSet::new();
    let mut open = None;
    for (i, line) in lines.iter().enumerate() {
        if line.trim() == begin {
            open = Some(i);
        } else if line.trim() == end {
            if let Some(from) = open.take() {
                inside.extend(from..=i);
            }
        }
    }
    inside
}

/// A line that is one Go curator marker comment and nothing else.
fn is_curator_marker_line(line: &str) -> bool {
    let line = line.trim();
    let one_comment = line.find("-->").is_some_and(|at| at + 3 == line.len());
    CURATOR_ENDS.contains(&line)
        || (one_comment && CURATOR_STARTS.iter().any(|start| line.starts_with(start)))
}

/// A read-only git command in the project: the scrubbed environment, the stored objects, paths
/// taken literally and no fetch of a missing object.
fn git(project: &Path) -> Command {
    let mut command = Command::new("git");
    command
        .args(["--no-replace-objects", "--literal-pathspecs", "-C"])
        .arg(project);
    identity::scrub_git_env(&mut command);
    command.env("GIT_NO_LAZY_FETCH", "1").stdin(Stdio::null());
    command
}

/// The paths git tracks below the project, or `None` when git cannot say.
fn tracked_paths(project: &Path) -> Option<HashSet<String>> {
    let out = git(project).args(["ls-files", "-z"]).output().ok()?;
    out.status.success().then(|| {
        String::from_utf8_lossy(&out.stdout)
            .split('\0')
            .filter(|path| !path.is_empty())
            .map(str::to_string)
            .collect()
    })
}

/// The version of `path` in git history, on any ref, whose bytes hash to `record`: its blob id
/// and bytes.
fn recorded_version(project: &Path, path: &str, record: &str) -> Result<(String, Vec<u8>), NoBase> {
    let ids = versions(project, path)?;
    find_blob(project, &ids, record)
        .map_err(|_| NoBase::GitFailed)?
        .ok_or(NoBase::NoMatchingVersion)
}

/// The blob ids of the latest [`HISTORY_LIMIT`] distinct versions of `path`, newest first, the
/// ones a merge wrote or dropped included.
fn versions(project: &Path, path: &str) -> Result<Vec<String>, NoBase> {
    let out = git(project)
        .args([
            "log",
            "--all",
            "--full-history",
            "--diff-merges=separate",
            "--no-renames",
            "--no-show-signature",
            "--raw",
            "--no-abbrev",
            "--format=",
            "--",
            path,
        ])
        .output()
        .map_err(|_| NoBase::GitFailed)?;
    if !out.status.success() {
        let message = String::from_utf8_lossy(&out.stderr);
        return Err(if message.contains(identity::NOT_A_REPOSITORY) {
            NoBase::NotARepository
        } else {
            NoBase::GitFailed
        });
    }
    let mut seen = HashSet::new();
    Ok(String::from_utf8_lossy(&out.stdout)
        .lines()
        .filter_map(|line| {
            // `:<old mode> <new mode> <old id> <new id> <status><tab><path>`
            let mut fields = line.strip_prefix(':')?.split_ascii_whitespace();
            let mode = fields.nth(1)?;
            let id = fields.nth(1)?;
            // A removed path or a link is no version of the file's text.
            (mode.starts_with("100") && seen.insert(id.to_string())).then(|| id.to_string())
        })
        .take(HISTORY_LIMIT)
        .collect())
}

/// The first of the objects `ids` whose bytes hash to `record`, read through one
/// `cat-file --batch` child.
fn find_blob(
    project: &Path,
    ids: &[String],
    record: &str,
) -> io::Result<Option<(String, Vec<u8>)>> {
    let mut child = git(project)
        .args(["cat-file", "--batch"])
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::null())
        .spawn()?;
    let found = match (child.stdin.take(), child.stdout.take()) {
        (Some(input), Some(output)) => compare_blobs(input, BufReader::new(output), ids, record),
        _ => Err(io::Error::other("git has no pipes")),
    };
    // Both pipes are closed by now, so the child ends.
    child.wait()?;
    found
}

/// Asks for one object at a time and reads its whole answer before the next request, so neither
/// pipe can fill while the other side waits.
fn compare_blobs(
    mut input: impl Write,
    mut output: impl BufRead,
    ids: &[String],
    record: &str,
) -> io::Result<Option<(String, Vec<u8>)>> {
    let broken = || io::Error::new(ErrorKind::InvalidData, "unreadable cat-file answer");
    for id in ids {
        writeln!(input, "{id}")?;
        input.flush()?;
        let mut header = String::new();
        output.read_line(&mut header)?;
        // `<id> <type> <size>`, or `<id> missing`.
        let size: u64 = match header.split_ascii_whitespace().collect::<Vec<_>>()[..] {
            [_, "missing"] => continue,
            [_, _, size] => size.parse().map_err(|_| broken())?,
            _ => return Err(broken()),
        };
        let mut bytes = Vec::new();
        (&mut output).take(size + 1).read_to_end(&mut bytes)?;
        if bytes.pop() != Some(b'\n') || bytes.len() as u64 != size {
            return Err(broken());
        }
        if digest(&bytes) == record {
            return Ok(Some((id.clone(), bytes)));
        }
    }
    Ok(None)
}
