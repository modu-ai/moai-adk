//! The instruction files a Go-generation project carries (`claude-md-plan`,
//! `claude-md-apply`, `claude-md-revert`): whether the root `CLAUDE.md` is a Go-generation file,
//! the sources in their fixed order with their facts, the lines of each that are the project's
//! own, those lines as hunks of one heading section each, the texts the `--keep` picks compose,
//! the digest and command of that plan, the apply of it and the undo of that apply.
//!
//! The preview reads only: no file is written, a link is never followed and only read-only git
//! commands run. The apply composes through the same code, refuses before its first write, saves
//! a receipt, moves every source into an owner-only backup folder and then writes the new files;
//! it overwrites none but `.gitignore`, which only gains a line, and a step that fails is undone
//! from that receipt. The undo, also run by the revert, removes only a file the apply wrote that
//! still holds those bytes, puts an original back only into an empty place, takes the
//! `.gitignore` line back only while no personal file would show, and never deletes an original.
//! Hunk text and the composed texts are part of the preview's answer; a line that holds a
//! credential-shaped value is counted and printed nowhere, and the receipt holds paths and
//! digests, never text.

use super::PERMISSION_PLACEHOLDER;
// By name: the crate root also has a function `apply`, which a module import would bring along.
use crate::apply::{
    publish_exclusive, publish_exclusive_mode, publish_replace, publish_replace_limited, read_opt,
};
use crate::go_residue::{find_manifest, Recorded};
use crate::go_state::{
    linked_component, make_dirs, make_private_dirs, move_exclusive, parent_of, plain_relative,
    read_bounded_text, sha256_file, still_as_previewed,
};
use crate::{agents, agents_local_merge, claude, contract, identity, state_root};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    collections::{BTreeMap, HashMap, HashSet},
    fs,
    io::{self, BufRead, BufReader, ErrorKind, Read, Write},
    path::{Path, PathBuf},
    process::{Command, Output, Stdio},
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
const BLOCK_MARKER: &str = "managed_block_marker";

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

    let mut sources = Vec::new();
    let mut refused = Vec::new();
    for (path, read) in read_candidates(project) {
        match read {
            Ok(Some(text)) => {
                let sha256 = digest(text.as_bytes());
                // The baseline is the text the Go tool would have written: its template hash,
                // never the deployed one, which holds a user's own pre-existing file as it was.
                let record = records.get(&path).and_then(|r| r.template_hash.as_deref());
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

/// One digest over the path and the bytes of every file the step would move, what a decline of
/// the step is bound to. `None` while the root `CLAUDE.md` is no Go-generation file or a source
/// cannot be read. Reads only, without git.
pub fn source_digest(project: &Path) -> Option<String> {
    go_claude_md(project)?;
    let mut bound = Vec::new();
    for (path, read) in read_candidates(project) {
        if let Some(text) = read.ok()? {
            bound.push(json!([path, digest(text.as_bytes())]));
        }
    }
    Some(digest(Value::Array(bound).to_string().as_bytes()))
}

/// What one read of a source gives: its text, `None` without such a file, or why it is refused.
type SourceRead = Result<Option<String>, &'static str>;

/// Every path that can be a source, read: the root files in their order, then the local rules by
/// path.
fn read_candidates(project: &Path) -> impl Iterator<Item = (String, SourceRead)> + '_ {
    ROOT_SOURCES
        .iter()
        .map(|name| (name.to_string(), None))
        .chain(local_rules(project))
        .map(move |(path, refusal)| {
            let read = match refusal {
                Some(reason) => Err(reason),
                None => read_source(project, &path),
            };
            (path, read)
        })
}

/// The answer of `claude-md-plan` for the `--keep` values `keep`. An error is a project that is
/// no folder or a `--keep` value that names no pick; anything else that stops the apply is listed
/// under `apply_blocked`, and then no digest and no command are printed.
pub fn plan(project: &Path, keep: &[String]) -> Result<Value, String> {
    Ok(compose(project, keep)?.answer)
}

/// One composition of the project and its picks: the preview's answer, and what an apply of that
/// answer would move and write. Both commands go through it, so the printed command applies
/// exactly what was previewed.
struct Composed {
    /// The canonical project folder.
    root: PathBuf,
    /// The state folder's name.
    state: &'static str,
    answer: Value,
    /// Why no apply may run.
    blocked: Vec<String>,
    /// `None` while anything blocks.
    plan_digest: Option<String>,
    /// The readable sources: what an apply moves.
    sources: Vec<Source>,
    /// The path and the whole text of each file an apply writes, `CLAUDE.md` first.
    files: Vec<(&'static str, String)>,
}

/// Reads the project once and composes it with the picks `keep`. Writes nothing.
fn compose(project: &Path, keep: &[String]) -> Result<Composed, String> {
    let preview = preview(project)?;
    let root = fs::canonicalize(project).map_err(|_| contract::PROJECT_PATH_UNAVAILABLE)?;
    let shown = root.to_str().ok_or("non UTF-8 project path")?;
    let state = state_root::dir(&root);
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

    let mut blocked: Vec<String> = Vec::new();
    if preview.go_proof.is_none() {
        blocked.push(format!(
            "the root {CLAUDE_MD} is not a Go-generation instruction file, so this step has \
             nothing to replace"
        ));
    }
    blocked.extend(
        preview
            .refused
            .iter()
            .map(|r| format!("a source is refused: {} {}", r.path, r.reason)),
    );
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
    let personal = picks.values().any(|to| *to == Destination::Personal);
    blocked.extend(ground(&root, state, PROBE_STAMP, &preview.sources, personal).blocked);

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
    let command = plan_digest.as_ref().map(|plan_digest| {
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

    let answer = json!({
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
        "apply": command,
    });
    let files = results
        .iter()
        .map(|file| (file.path, file.text()))
        .collect();
    Ok(Composed {
        state,
        answer,
        blocked,
        plan_digest,
        files,
        sources: preview.sources,
        root,
    })
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

/// Whether the line holds a comment that begins or ends a managed block of moai: copied into a
/// new file beside the block moai renders there, it could leave that file's markers unpaired.
fn holds_block_marker(line: &str) -> bool {
    line.split("<!--").skip(1).any(|comment| {
        let comment = comment.trim_start();
        ["moai:managed:begin", "moai:managed:end"]
            .iter()
            .any(|word| comment.starts_with(word))
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
    /// A line that imports a moved file (in its source, or where it would now stand), a line that
    /// holds a Go curator start or end marker and a line that holds a managed-block marker (also
    /// inside a code block, where the markers are found as well) are left out; blank lines at a
    /// hunk's ends are dropped.
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
                } else if holds_block_marker(line) {
                    Some(BLOCK_MARKER)
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
/// `template_hash` (`sha256:<hex>`) the found Go manifest holds for the path, `None` when it
/// holds none.
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

/// The backup folder under the state folder.
const BACKUPS: &str = "backups";
/// The folder under one apply's stamp folder that holds the moved sources at their paths.
const AREA: &str = "claude-md";
/// The receipt beside it. No other receipt has this name, so no other reader takes it for its own.
const RECEIPT_FILE: &str = "claude-md-receipt.json";
/// One size limit for writing the receipt, replacing it and reading it.
const RECEIPT_MAX: u64 = state_root::GO_STATE_RECEIPT_MAX;
/// The operation a receipt names.
const APPLY_OPERATION: &str = "claude_md.apply";
/// The receipt's states.
const PENDING: &str = "pending";
const COMMITTED: &str = "committed";
const RESTORED: &str = "restored";
const PARTIAL: &str = "partial";
const REVERTED: &str = "reverted";
const STATES: [&str; 5] = [PENDING, COMMITTED, RESTORED, PARTIAL, REVERTED];
/// The receipt and the personal file are their owner's alone.
const PRIVATE_FILE: u32 = 0o600;
/// How many following seconds a stamp may be moved on when its folder is taken already.
const STAMP_TRIES: i64 = 120;
/// A stamp of the real shape, for the preview's git question before any apply has a stamp.
const PROBE_STAMP: &str = "19700101T000000Z";
/// The file that gains the one line which keeps the personal file out of git.
const GITIGNORE: &str = ".gitignore";
/// What the apply's answer says about when the change takes effect.
const RESTART_NOTICE: &str = "settings apply at once, while running sessions and lanes read \
                              changed instructions only after a restart";
/// The refusal of a digest that is not the fresh composition's (`contract::classify` reads the
/// first words as `stale_plan`).
const STALE: &str = "stale or altered plan: the sources, the profile or the picks no longer \
                     match the preview that was confirmed; run `moai claude-md-plan` again, read \
                     it and confirm again";

fn backup_folder(state: &str) -> String {
    format!("{state}/{BACKUPS}")
}

fn receipt_rel(state: &str, stamp: &str) -> String {
    format!("{}/{stamp}/{RECEIPT_FILE}", backup_folder(state))
}

fn area_rel(state: &str, stamp: &str) -> String {
    format!("{}/{stamp}/{AREA}", backup_folder(state))
}

/// A refusal that comes before the first write.
fn refused(why: &str) -> String {
    format!("claude-md-apply refused: {why}; nothing was written")
}

/// What this disk and git say about the moves and writes of an apply under one stamp.
#[derive(Default)]
struct Ground {
    /// Why no apply may run.
    blocked: Vec<String>,
    /// The project is a git work tree.
    repository: bool,
    /// The line `.gitignore` gains so that git ignores the personal file.
    ignore: Option<IgnoreAdd>,
}

/// `.gitignore` as it is (`None`: no such file) and with the one line added.
struct IgnoreAdd {
    before: Option<Vec<u8>>,
    after: Vec<u8>,
}

/// Looks at the places an apply would move into and write, and asks git whether it would offer
/// any of them: the receipt and every backup file, and the personal file when `personal`. Outside
/// a git work tree git offers nothing; a git that cannot say counts as one that would. Reads only.
fn ground(root: &Path, state: &str, stamp: &str, sources: &[Source], personal: bool) -> Ground {
    let mut ground = Ground::default();
    for source in sources {
        let path = &source.path;
        if !plain_relative(path) {
            ground
                .blocked
                .push(format!("{path} is not a plain relative path"));
        }
        if path.rsplit('/').next() == Some(GITIGNORE) {
            ground.blocked.push(format!(
                "{path} would change which backup files git ignores once it is moved among \
                 them; rename it or move it away first"
            ));
        }
    }
    let backups = backup_folder(state);
    if linked_component(root, &backups) {
        ground.blocked.push(format!(
            "the backup folder {backups} is a link or lies below one"
        ));
        return ground;
    }
    if fs::symlink_metadata(root.join(&backups)).is_ok_and(|m| !m.is_dir()) {
        ground
            .blocked
            .push(format!("the backup folder {backups} is not a folder"));
        return ground;
    }
    let cannot = |why: String| {
        format!(
            "git could not say whether it ignores the backup folder ({why}), so it is taken as \
             not ignored"
        )
    };
    match in_work_tree(root) {
        Ok(false) => return ground,
        Ok(true) => ground.repository = true,
        Err(why) => {
            ground.blocked.push(cannot(why));
            return ground;
        }
    }
    let area = area_rel(state, stamp);
    let paths = std::iter::once(receipt_rel(state, stamp))
        .chain(sources.iter().map(|s| format!("{area}/{}", s.path)));
    for path in paths {
        match check_ignore(root, "-q", &path) {
            Ok(true) => continue,
            Ok(false) => ground.blocked.push(format!(
                "git would not ignore the backup folder: {path} is not ignored; add the line \
                 `{backups}/` to {GITIGNORE}"
            )),
            Err(why) => ground.blocked.push(cannot(why)),
        }
        // One path is enough to name.
        break;
    }
    if personal {
        match personal_ignore(root) {
            Ok(add) => ground.ignore = add,
            Err(reason) => ground.blocked.push(reason),
        }
    }
    ground
}

/// The line to add so that git ignores the personal file, `None` when a rule ignores it already;
/// an error says why no rule would keep it out of git.
fn personal_ignore(root: &Path) -> Result<Option<IgnoreAdd>, String> {
    let file = CLAUDE_LOCAL_MD;
    let cannot = |why: String| format!("git could not say whether it keeps {file} out ({why})");
    let tracked = tracked_paths(root).ok_or_else(|| cannot("its files cannot be listed".into()))?;
    if tracked.contains(file) {
        return Err(format!(
            "git tracks {file}, so no ignore rule keeps a personal pick out of the repository; \
             keep those lines as shared or leave them in the backup, or run `git rm --cached \
             {file}` first"
        ));
    }
    if check_ignore(root, "-q", file).map_err(cannot)? {
        return Ok(None);
    }
    // A rule that matches a path git does not ignore is a `!` rule.
    if check_ignore(root, "-v", file).map_err(cannot)? {
        return Err(format!(
            "a rule you wrote un-ignores {file}, and your rule wins: remove it first, or keep \
             those lines as shared or leave them in the backup"
        ));
    }
    if linked_component(root, GITIGNORE) {
        return Err(format!(
            "{GITIGNORE} is a link, so the rule that keeps {file} out cannot be added"
        ));
    }
    let before = read_opt(&root.join(GITIGNORE))
        .map_err(|e| format!("{GITIGNORE} cannot be changed ({e})"))?;
    let after = with_ignore_line(before.as_deref());
    Ok(Some(IgnoreAdd { before, after }))
}

/// `.gitignore` as an apply leaves it: the earlier bytes (`None`: no such file) and the one line
/// that names the personal file, in the line ends the file uses.
fn with_ignore_line(before: Option<&[u8]>) -> Vec<u8> {
    let mut after = before.unwrap_or_default().to_vec();
    let eol = if after.windows(2).any(|w| w == b"\r\n") {
        "\r\n"
    } else {
        "\n"
    };
    if !after.is_empty() && !after.ends_with(b"\n") {
        after.extend_from_slice(eol.as_bytes());
    }
    after.extend_from_slice(format!("{CLAUDE_LOCAL_MD}{eol}").as_bytes());
    after
}

/// A git command about ignore rules, built apart from [`git`]: `check-ignore` reads its path the
/// ordinary way, and git takes no optional lock.
fn ask_git(root: &Path, args: &[&str]) -> Result<Output, String> {
    let mut command = Command::new("git");
    identity::scrub_git_env(&mut command);
    command
        .arg("--no-optional-locks")
        .arg("-C")
        .arg(root)
        .args(args)
        .stdin(Stdio::null())
        .output()
        .map_err(|_| "git could not be run".into())
}

/// Whether the project is a git work tree. Git saying there is no repository counts only where
/// no `.git` entry stands at or above the project: one git rejects is a git that cannot say.
fn in_work_tree(root: &Path) -> Result<bool, String> {
    let out = ask_git(root, &["rev-parse", "--is-inside-work-tree"])?;
    if out.status.success() {
        return match String::from_utf8_lossy(&out.stdout).trim() {
            "true" => Ok(true),
            "false" => Ok(false),
            _ => Err("git gave an unexpected answer".into()),
        };
    }
    let why = identity::failure_summary(&out.stderr);
    let marked = root
        .ancestors()
        .any(|folder| fs::symlink_metadata(folder.join(".git")).is_ok());
    if why.starts_with(identity::NOT_A_REPOSITORY) && !marked {
        Ok(false)
    } else {
        Err(why)
    }
}

/// `git check-ignore` of one path, which need not exist: with `-q` whether git ignores it, with
/// `-v` whether any rule matches it, a `!` rule included.
fn check_ignore(root: &Path, how: &str, path: &str) -> Result<bool, String> {
    let out = ask_git(root, &["check-ignore", "--no-index", how, "--", path])?;
    match out.status.code() {
        Some(0) => Ok(true),
        Some(1) => Ok(false),
        _ => Err(identity::failure_summary(&out.stderr)),
    }
}

/// A moved source, as the receipt holds it.
#[derive(Clone, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
struct MovedFile {
    /// Where it was, relative to the project.
    path: String,
    /// Where it is, relative to the project.
    backup: String,
    sha256: String,
    /// The permission bits; the move keeps them.
    mode: u32,
    /// `tracked`, `untracked` or `unknown`.
    git: String,
    moved: bool,
}

/// A file the apply writes where no file is.
#[derive(Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
struct WrittenFile {
    path: String,
    sha256: String,
    mode: u32,
    written: bool,
}

/// The line `.gitignore` gains. The earlier bytes are the first `before_len` bytes of the new
/// ones, so no text is kept here; a file the apply creates has no `before_sha256`.
#[derive(Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
struct IgnoreRule {
    path: String,
    line: String,
    before_sha256: Option<String>,
    before_len: u64,
    after_sha256: String,
    written: bool,
}

/// Everything an undo of one apply needs: paths, digests and modes, never file text. It is saved
/// before the first move with every flag off, and again after each step, so a step that was done
/// and not saved is possible: a reader checks the disk against the digests, not the flags alone.
// @cc:careful `check_receipt` refuses what this apply does not write: a new field, path, state or
// word goes into both.
#[derive(Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
struct Receipt {
    schema_version: u32,
    operation: String,
    project: String,
    /// The receipt id: the name of the stamp folder.
    stamp: String,
    plan_digest: String,
    /// The caller's `--permission-reference`, recorded as given (not an authorization).
    permission_reference: String,
    /// One of [`STATES`]: `pending` until every step is done, then `committed`; `restored` or
    /// `partial` after an apply that failed and undid its steps, `reverted` or `partial` after a
    /// revert.
    state: String,
    moves: Vec<MovedFile>,
    writes: Vec<WrittenFile>,
    ignore: Option<IgnoreRule>,
}

fn encode(receipt: &Receipt) -> Result<Vec<u8>, String> {
    let bytes = serde_json::to_vec_pretty(receipt).map_err(|_| "encode failed".to_string())?;
    if bytes.len() as u64 > RECEIPT_MAX {
        return Err(format!(
            "the receipt is larger than {} MiB",
            RECEIPT_MAX / (1024 * 1024)
        ));
    }
    Ok(bytes)
}

/// The receipt and its file: every change is saved at once.
struct Journal {
    /// Relative to the project.
    rel: String,
    path: PathBuf,
    /// The bytes on disk.
    saved: Vec<u8>,
    receipt: Receipt,
}

impl Journal {
    /// Saves a new receipt where no file is, its owner's alone, and reads it back.
    fn create(root: &Path, rel: String, receipt: Receipt) -> Result<Journal, String> {
        let saved = encode(&receipt)?;
        let path = root.join(&rel);
        publish_exclusive_mode(&path, &saved, Some(PRIVATE_FILE))?;
        if fs::read(&path).ok().as_deref() != Some(saved.as_slice()) {
            return Err("it does not read back".into());
        }
        Ok(Journal {
            rel,
            path,
            saved,
            receipt,
        })
    }

    fn save(&mut self) -> Result<(), String> {
        if faulted("claude_md_save_fails") {
            return Err("fault".into());
        }
        let bytes = encode(&self.receipt)?;
        publish_replace_limited(&self.path, Some(&self.saved), &bytes, RECEIPT_MAX)?;
        self.saved = bytes;
        Ok(())
    }

    /// The paths the receipt marks as written, the `.gitignore` first.
    fn written(&self) -> Vec<&str> {
        let ignore = self.receipt.ignore.iter().filter(|rule| rule.written);
        let files = self.receipt.writes.iter().filter(|file| file.written);
        ignore
            .map(|rule| rule.path.as_str())
            .chain(files.map(|file| file.path.as_str()))
            .collect()
    }

    /// Undoes an apply a step of which failed after the pending receipt was saved, saves what
    /// the receipt then reads, and returns the apply's error: `restored` when nothing is left,
    /// else `partial` with every path that is left.
    fn restore(&mut self, root: &Path, at: &str, why: &str, area: &str) -> String {
        let done = undo(root, &mut self.receipt);
        let state = if done.clean() { RESTORED } else { PARTIAL };
        self.receipt.state = state.into();
        let saved = self.save();
        let rel = &self.rel;
        let receipt = match saved {
            Ok(()) => format!("the receipt {rel} reads {state}"),
            Err(e) => {
                format!("the receipt {rel} could not be updated ({e}) and may show an earlier step")
            }
        };
        if done.clean() {
            // `contract::classify` reads the first words as `conflict`, whatever the reason.
            let what = if done.restored.is_empty() && done.removed.is_empty() {
                "no project file was moved or changed"
            } else {
                "every change the apply made was undone"
            };
            return format!("conflict in claude-md-apply at {at}: {why}; {what}; {receipt}");
        }
        let undone: Vec<&str> = done
            .removed
            .iter()
            .chain(&done.restored)
            .map(String::as_str)
            .collect();
        let undone = if undone.is_empty() {
            "none".to_string()
        } else {
            undone.join(", ")
        };
        let left: Vec<String> = done
            .reasons
            .iter()
            .map(|(path, why)| format!("{path} ({why})"))
            .collect();
        format!(
            "partial apply: stopped at {at}: {why}; undone: {undone}; could not restore: {}; \
             {receipt}; an original that is not back stays under {area}; clear what is in the way \
             and run `moai claude-md-revert {} --receipt {} --permission-reference \
             {PERMISSION_PLACEHOLDER}`",
            left.join(", "),
            self.receipt.project,
            self.receipt.stamp,
        )
    }
}

/// What an undo of one receipt did, and what it had to leave.
#[derive(Default)]
struct Undone {
    /// Originals put back, and the `.gitignore` given its earlier bytes.
    restored: Vec<String>,
    /// Files of the apply removed.
    removed: Vec<String>,
    /// Paths left as they are.
    conflicts: Vec<String>,
    /// Originals whose recorded bytes are neither at their place nor in the backup.
    missing: Vec<String>,
    /// Why, for each path of `conflicts` and `missing`.
    reasons: BTreeMap<String, String>,
}

impl Undone {
    /// Nothing is left of the apply.
    fn clean(&self) -> bool {
        self.conflicts.is_empty() && self.missing.is_empty()
    }

    /// Leaves `path` as it is; the first reason given for it stays.
    fn keep(&mut self, path: &str, why: impl Into<String>) {
        if !self.conflicts.iter().any(|known| known == path) {
            self.conflicts.push(path.into());
        }
        self.reasons
            .entry(path.into())
            .or_insert_with(|| why.into());
    }

    fn lost(&mut self, path: &str) {
        self.missing.push(path.into());
        self.reasons.entry(path.into()).or_insert_with(|| {
            "the recorded bytes are neither at its place nor in the backup".into()
        });
    }
}

/// What stands at a path inside the project, read through no link.
#[derive(PartialEq, Eq)]
enum Standing {
    Absent,
    /// The path is a link or lies below one.
    Linked,
    /// A regular file with the `sha256:<hex>` of its bytes.
    File(String),
    /// Anything else, and what cannot be read.
    Other,
}

fn standing(root: &Path, rel: &str) -> Standing {
    if linked_component(root, rel) {
        return Standing::Linked;
    }
    let path = root.join(rel);
    match fs::symlink_metadata(&path) {
        Err(e) if is_absent(&e) => Standing::Absent,
        Ok(meta) if meta.is_file() => sha256_file(&path).map_or(Standing::Other, Standing::File),
        _ => Standing::Other,
    }
}

/// Whether a regular file with these bytes stands at `rel`.
fn holds(root: &Path, rel: &str, sha256: &str) -> bool {
    matches!(standing(root, rel), Standing::File(found) if found == sha256)
}

/// Undoes what an apply did, as far as the disk allows, and marks in the receipt what is undone.
/// Every step is decided by what stands on the disk against the recorded digests, so it serves an
/// apply that failed, one that was stopped between two steps and a revert that runs again. The
/// order keeps the text the apply wrote hidden: the files it wrote go first, then the originals
/// come back, and the `.gitignore` line is taken back last, only while the personal file's place
/// is empty or holds its original again.
fn undo(root: &Path, receipt: &mut Receipt) -> Undone {
    let mut done = Undone::default();
    remove_written(root, receipt, &mut done);
    move_back(root, receipt, &mut done);
    take_ignore_back(root, receipt, &mut done);
    done
}

/// Removes each file the apply wrote that still holds those bytes, the last written first, while
/// the original of its path, if it had one, can come back.
fn remove_written(root: &Path, receipt: &mut Receipt, done: &mut Undone) {
    for file in receipt.writes.iter_mut().rev() {
        let original = receipt.moves.iter().find(|m| m.path == file.path);
        let kept = original.is_some_and(|m| holds(root, &m.backup, &m.sha256));
        match standing(root, &file.path) {
            Standing::Absent => file.written = false,
            Standing::File(found) if found == file.sha256 && (original.is_none() || kept) => {
                match fs::remove_file(root.join(&file.path)) {
                    Ok(()) => {
                        done.removed.push(file.path.clone());
                        file.written = false;
                        test_stop("claude_md_stop_after_removal");
                    }
                    Err(e) => done.keep(&file.path, format!("it could not be removed ({e})")),
                }
            }
            // The original is at its place already.
            Standing::File(found) if original.is_some_and(|m| m.sha256 == found) => {
                file.written = false;
            }
            Standing::File(found) if found == file.sha256 => done.keep(
                &file.path,
                "it is kept because its original cannot be put back from the backup",
            ),
            // Something else stands there. Where the apply never wrote and an original had the
            // path, the original's own step decides.
            _ if file.written || original.is_none() => done.keep(
                &file.path,
                "it does not hold what the apply writes, so it is kept",
            ),
            _ => {}
        }
    }
}

/// Puts each original back where its place is empty, with its recorded bytes and mode. A backup
/// copy beside an original that is already at its place is left alone.
fn move_back(root: &Path, receipt: &mut Receipt, done: &mut Undone) {
    for file in &mut receipt.moves {
        let place = standing(root, &file.path);
        if matches!(&place, Standing::File(found) if *found == file.sha256) {
            file.moved = false;
            continue;
        }
        if !holds(root, &file.backup, &file.sha256) {
            // Lost: the receipt says it was moved, or its place is empty while something else
            // stands at its backup path. Otherwise it was never moved, or came back before and
            // whatever happened to it since is its owner's doing: nothing to undo.
            let emptied =
                place == Standing::Absent && standing(root, &file.backup) != Standing::Absent;
            if file.moved || emptied {
                done.lost(&file.path);
            }
            continue;
        }
        match place {
            Standing::Absent => {}
            Standing::Linked => {
                done.keep(
                    &file.path,
                    "a link is at or above its place; the original stays in the backup",
                );
                continue;
            }
            _ => {
                done.keep(
                    &file.path,
                    "its place is taken; the original stays in the backup",
                );
                continue;
            }
        }
        let (from, to) = (root.join(&file.backup), root.join(&file.path));
        let moved = parent_of(&file.path)
            .and_then(|parent| make_dirs(root, &parent))
            .and_then(|()| still_as_previewed(&from, &file.path, &file.sha256))
            .and_then(|()| move_exclusive(&from, &to));
        // A move that linked the file back and could not remove the copy put it back all the same.
        if let Err(e) = moved {
            if !holds(root, &file.path, &file.sha256) {
                done.keep(&file.path, format!("{e}; the original stays in the backup"));
                continue;
            }
        }
        file.moved = false;
        done.restored.push(file.path.clone());
        if let Err(e) = set_mode(&to, file.mode) {
            done.keep(&file.path, format!("it is back, but its mode is not ({e})"));
        }
        test_stop("claude_md_stop_after_restore");
    }
}

/// Why the `.gitignore` line stays although the file holds what the apply left.
const IGNORE_STAYS: &str =
    "CLAUDE.local.md is not the recorded original, and taking the line back \
     would show it to git; move it away and revert again, or remove the line yourself";

/// Takes the apply's line out of `.gitignore`, or removes the file the apply created, while that
/// file holds what the apply left and the line has nothing to hide: nothing stands at the
/// personal file's place, or the recorded original does, which git showed before the apply added
/// the line. The text the apply writes there stays hidden, even where a receipt calls it the
/// original.
fn take_ignore_back(root: &Path, receipt: &mut Receipt, done: &mut Undone) {
    let personal = match standing(root, CLAUDE_LOCAL_MD) {
        Standing::Absent => true,
        Standing::File(found) => {
            let original = receipt.moves.iter().find(|m| m.path == CLAUDE_LOCAL_MD);
            let written = receipt.writes.iter().find(|w| w.path == CLAUDE_LOCAL_MD);
            original.is_some_and(|m| m.sha256 == found) && written.is_none_or(|w| w.sha256 != found)
        }
        Standing::Linked | Standing::Other => false,
    };
    let Some(rule) = receipt.ignore.as_mut() else {
        return;
    };
    match standing(root, GITIGNORE) {
        Standing::Absent if rule.before_sha256.is_none() => rule.written = false,
        Standing::File(found) if rule.before_sha256.as_ref() == Some(&found) => {
            rule.written = false;
        }
        Standing::File(found) if found == rule.after_sha256 => {
            if !personal {
                done.keep(GITIGNORE, IGNORE_STAYS);
                return;
            }
            match restore_ignore(root, rule) {
                Ok(()) => {
                    let list = match rule.before_sha256 {
                        Some(_) => &mut done.restored,
                        None => &mut done.removed,
                    };
                    list.push(GITIGNORE.into());
                    rule.written = false;
                }
                Err(e) => done.keep(GITIGNORE, e),
            }
        }
        // Changed since, or never written: it is named while it may hold the apply's line.
        found => {
            let holds_line = matches!(found, Standing::File(_))
                && read_opt(&root.join(GITIGNORE))
                    .ok()
                    .flatten()
                    .is_some_and(|bytes| {
                        bytes
                            .split(|b| *b == b'\n')
                            .any(|line| line.trim_ascii_end() == CLAUDE_LOCAL_MD.as_bytes())
                    });
            if rule.written || holds_line {
                done.keep(GITIGNORE, "it changed after the apply, so it is kept");
            }
        }
    }
}

/// The write of [`take_ignore_back`]. Whatever the receipt says, only the exact line an apply
/// adds is taken out: the earlier bytes plus that line must be the file as it is.
fn restore_ignore(root: &Path, rule: &IgnoreRule) -> Result<(), String> {
    let path = root.join(GITIGNORE);
    let now = read_opt(&path)?.ok_or("it is gone")?;
    let other = || "it does not hold the one line an apply adds, so it is kept".to_string();
    let earlier = match &rule.before_sha256 {
        Some(sha256) => Some(
            usize::try_from(rule.before_len)
                .ok()
                .and_then(|length| now.get(..length))
                .filter(|earlier| digest(earlier) == *sha256)
                .ok_or_else(other)?,
        ),
        None => None,
    };
    if with_ignore_line(earlier) != now {
        return Err(other());
    }
    match earlier {
        Some(earlier) => publish_replace(&path, Some(&now), earlier),
        None => fs::remove_file(&path).map_err(|e| format!("it could not be removed ({e})")),
    }
}

/// Test seam: `MOAI_TEST_FAULT=<point>` appends one line to the file at `path`, standing in for
/// another writer that comes between two steps of the apply. Debug builds only.
fn test_fault(point: &str, path: &Path) {
    #[cfg(debug_assertions)]
    if std::env::var("MOAI_TEST_FAULT").as_deref() == Ok(point) {
        if let Ok(mut file) = fs::OpenOptions::new().append(true).create(true).open(path) {
            let _ = file.write_all(b"fault\n");
        }
    }
    #[cfg(not(debug_assertions))]
    let _ = (point, path);
}

/// Test seam: whether `MOAI_TEST_FAULT` names this point. Debug builds only.
#[cfg(debug_assertions)]
fn faulted(point: &str) -> bool {
    std::env::var("MOAI_TEST_FAULT").as_deref() == Ok(point)
}

#[cfg(not(debug_assertions))]
fn faulted(_point: &str) -> bool {
    false
}

/// Test seam: ends the process at this point, as an interrupted run does, with exit code 70.
/// Debug builds only.
fn test_stop(point: &str) {
    if faulted(point) {
        std::process::exit(70);
    }
}

#[cfg(unix)]
fn mode_bits(meta: &fs::Metadata) -> u32 {
    use std::os::unix::fs::PermissionsExt;
    meta.permissions().mode() & 0o777
}

#[cfg(not(unix))]
fn mode_bits(_meta: &fs::Metadata) -> u32 {
    PRIVATE_FILE
}

/// Gives a regular file the recorded permission bits when it has others.
#[cfg(unix)]
fn set_mode(path: &Path, mode: u32) -> io::Result<()> {
    use std::os::unix::fs::PermissionsExt;
    let meta = fs::symlink_metadata(path)?;
    if meta.is_file() && mode_bits(&meta) != mode {
        fs::set_permissions(path, fs::Permissions::from_mode(mode))?;
    }
    Ok(())
}

#[cfg(not(unix))]
fn set_mode(_path: &Path, _mode: u32) -> io::Result<()> {
    Ok(())
}

/// Creates a folder that must not exist yet, its owner's alone from the start.
fn create_private_dir(path: &Path) -> io::Result<()> {
    let mut builder = fs::DirBuilder::new();
    #[cfg(unix)]
    {
        use std::os::unix::fs::DirBuilderExt;
        builder.mode(0o700);
    }
    builder.create(path)
}

/// The first stamp, from now on, that names no entry of the backup folder.
fn free_stamp(root: &Path, state: &str) -> Result<String, String> {
    let now = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_secs() as i64)
        .unwrap_or(0);
    let backups = root.join(backup_folder(state));
    (0..STAMP_TRIES)
        .map(|i| crate::baseline::utc_from_epoch(now + i).replace(['-', ':'], ""))
        .find(|stamp| fs::symlink_metadata(backups.join(stamp)).is_err())
        .ok_or_else(|| refused("no free backup stamp in the next two minutes"))
}

/// Moves one source to its backup path: through no link, only with the previewed bytes, and
/// never onto an existing file. `faulted` marks the move the test seams act on.
fn move_source(root: &Path, file: &MovedFile, faulted: bool) -> Result<(), String> {
    let (from, to) = (root.join(&file.path), root.join(&file.backup));
    make_private_dirs(root, &parent_of(&file.backup)?)?;
    if faulted {
        test_fault("claude_md_source_edit", &from);
        test_fault("claude_md_backup_taken", &to);
    }
    if linked_component(root, &file.path) {
        return Err("a path component is a link".into());
    }
    still_as_previewed(&from, &file.path, &file.sha256)?;
    move_exclusive(&from, &to)
}

/// Adds the line to `.gitignore`: through no link, and only onto the bytes that were read.
fn add_ignore_rule(root: &Path, add: &IgnoreAdd) -> Result<(), String> {
    if linked_component(root, GITIGNORE) {
        return Err("it is a link".into());
    }
    let path = root.join(GITIGNORE);
    match &add.before {
        Some(before) => publish_replace(&path, Some(before), &add.after),
        None => publish_exclusive(&path, &add.after),
    }
}

/// Every step after the pending receipt: the moves, the read-back of each, the ignore rule
/// before any personal text is on disk, the new files, and the committed receipt. An error is
/// the step it stopped at and why.
fn perform(
    root: &Path,
    journal: &mut Journal,
    ignore: Option<&IgnoreAdd>,
    files: &[(&'static str, String)],
) -> Result<(), (String, String)> {
    let at = |step: &str| {
        let step = step.to_string();
        move |why: String| (step, why)
    };
    for i in 0..journal.receipt.moves.len() {
        let file = journal.receipt.moves[i].clone();
        move_source(root, &file, i == 1).map_err(at(&file.path))?;
        test_stop("claude_md_stop_after_move");
        journal.receipt.moves[i].moved = true;
        journal.save().map_err(at("saving the receipt"))?;
    }
    for file in &journal.receipt.moves {
        let kept = sha256_file(&root.join(&file.backup)).is_ok_and(|h| h == file.sha256);
        if !kept || fs::symlink_metadata(root.join(&file.path)).is_ok() {
            return Err(at(&file.path)(
                "it does not read back from the backup".into(),
            ));
        }
    }
    if let Some(add) = ignore {
        test_fault("claude_md_ignore_edit", &root.join(GITIGNORE));
        add_ignore_rule(root, add).map_err(at(GITIGNORE))?;
        test_stop("claude_md_stop_after_ignore");
        if let Some(rule) = journal.receipt.ignore.as_mut() {
            rule.written = true;
        }
        journal.save().map_err(at("saving the receipt"))?;
        if !check_ignore(root, "-q", CLAUDE_LOCAL_MD).map_err(at(GITIGNORE))? {
            return Err(at(GITIGNORE)(format!(
                "git still does not ignore {CLAUDE_LOCAL_MD}"
            )));
        }
    }
    for (i, (path, text)) in files.iter().enumerate() {
        let destination = root.join(path);
        if *path == CLAUDE_LOCAL_MD {
            test_fault("claude_md_destination_taken", &destination);
            if faulted("claude_md_personal_write_fails") {
                return Err(at(path)("fault".into()));
            }
        }
        let (mode, sha256) = {
            let file = &journal.receipt.writes[i];
            (file.mode, file.sha256.clone())
        };
        publish_exclusive_mode(&destination, text.as_bytes(), Some(mode)).map_err(at(path))?;
        test_stop("claude_md_stop_after_write");
        journal.receipt.writes[i].written = true;
        journal.save().map_err(at("saving the receipt"))?;
        if !sha256_file(&destination).is_ok_and(|h| h == sha256) {
            return Err(at(path)("it does not read back".into()));
        }
    }
    journal.receipt.state = COMMITTED.into();
    journal.save().map_err(at("marking the receipt committed"))
}

/// A permission reference of the accepted form that is neither the printed placeholder nor
/// shaped like a credential, since the receipt keeps it.
fn accepted_reference(reference: &str) -> bool {
    contract::permission_reference_ok(reference)
        && reference.trim() != PERMISSION_PLACEHOLDER
        && agents_local_merge::secret_kind(reference).is_none()
}

/// Applies the preview whose digest the caller confirmed. Everything that can refuse is asked
/// before the first write; after it, a step that fails is undone from the receipt
/// ([`Journal::restore`]) and the error says what is left.
pub fn apply(
    project: &Path,
    keep: &[String],
    plan_digest: &str,
    permission_reference: &str,
) -> Result<Value, String> {
    if !accepted_reference(permission_reference) {
        return Err(refused(
            "permission reference required: --permission-reference takes a short line that names \
             the user's approval, neither the printed placeholder nor a credential",
        ));
    }
    let composed = compose(project, keep).map_err(|e| refused(&e))?;
    if !composed.blocked.is_empty() {
        return Err(refused(&composed.blocked.join("; ")));
    }
    if composed.plan_digest.as_deref() != Some(plan_digest) {
        return Err(refused(STALE));
    }
    let Composed {
        root,
        state,
        sources,
        files,
        ..
    } = composed;
    let personal = files.iter().any(|(path, _)| *path == CLAUDE_LOCAL_MD);
    let stamp = free_stamp(&root, state)?;
    // Asked again for the paths that are really written.
    let ground = ground(&root, state, &stamp, &sources, personal);
    if !ground.blocked.is_empty() {
        return Err(refused(&ground.blocked.join("; ")));
    }

    let area = area_rel(state, &stamp);
    let mut moves = Vec::new();
    for source in &sources {
        let meta = fs::symlink_metadata(root.join(&source.path))
            .map_err(|_| refused(&format!("{} {UNREADABLE}", source.path)))?;
        moves.push(MovedFile {
            path: source.path.clone(),
            backup: format!("{area}/{}", source.path),
            sha256: source.sha256.clone(),
            mode: mode_bits(&meta),
            git: source.git.word().into(),
            moved: false,
        });
    }
    // `CLAUDE.md` keeps the mode it had; the personal file is its owner's alone.
    let claude_mode = moves
        .iter()
        .find(|m| m.path == CLAUDE_MD)
        .map_or(PRIVATE_FILE, |m| m.mode);
    let receipt = Receipt {
        schema_version: 1,
        operation: APPLY_OPERATION.into(),
        project: root.to_string_lossy().into_owned(),
        stamp: stamp.clone(),
        plan_digest: plan_digest.into(),
        permission_reference: permission_reference.into(),
        state: PENDING.into(),
        moves,
        writes: files
            .iter()
            .map(|(path, text)| WrittenFile {
                path: (*path).into(),
                sha256: digest(text.as_bytes()),
                mode: if *path == CLAUDE_MD {
                    claude_mode
                } else {
                    PRIVATE_FILE
                },
                written: false,
            })
            .collect(),
        ignore: ground.ignore.as_ref().map(|add| IgnoreRule {
            path: GITIGNORE.into(),
            line: CLAUDE_LOCAL_MD.into(),
            before_sha256: add.before.as_deref().map(digest),
            before_len: add.before.as_ref().map_or(0, |b| b.len() as u64),
            after_sha256: digest(&add.after),
            written: false,
        }),
    };

    let backups = backup_folder(state);
    let folder = format!("{backups}/{stamp}");
    make_private_dirs(&root, &backups)
        .and_then(|()| create_private_dir(&root.join(&folder)).map_err(|e| e.to_string()))
        .map_err(|e| {
            format!(
                "claude-md-apply refused: the backup folder {folder} cannot be made ({e}); no \
                 file was moved or written"
            )
        })?;
    let rel = receipt_rel(state, &stamp);
    let mut journal = Journal::create(&root, rel.clone(), receipt).map_err(|e| {
        format!(
            "claude-md-apply refused: the pending receipt {rel} could not be saved ({e}); no \
             file was moved or written"
        )
    })?;
    if let Err((at, why)) = perform(&root, &mut journal, ground.ignore.as_ref(), &files) {
        return Err(journal.restore(&root, &at, &why, &area));
    }

    let moves = &journal.receipt.moves;
    Ok(json!({
        "state": COMMITTED,
        "receipt": stamp,
        "receipt_path": rel,
        "backup": area,
        "plan_digest": plan_digest,
        "moved": moves.iter().map(|m| &m.path).collect::<Vec<_>>(),
        "tracked_moved": moves
            .iter()
            .filter(|m| m.git == Tracking::Tracked.word())
            .map(|m| &m.path)
            .collect::<Vec<_>>(),
        "written": journal.written(),
        "ignore_rule": journal
            .receipt
            .ignore
            .as_ref()
            .map(|rule| json!({"path": rule.path, "line": rule.line})),
        "backup_ignore": if ground.repository { "git" } else { "not_a_repository" },
        "notice": RESTART_NOTICE,
    }))
}

/// A refusal of the revert that comes before its first write.
fn revert_refused(why: &str) -> String {
    format!("claude-md-revert refused: {why}; nothing was written")
}

/// `sha256:` and 64 hex digits.
fn is_digest(text: &str) -> bool {
    text.strip_prefix("sha256:")
        .is_some_and(|hex| hex.len() == 64 && hex.bytes().all(|b| b.is_ascii_hexdigit()))
}

/// Whether an apply can move this path: a root source, or a file under the local rules folder.
fn movable(path: &str) -> bool {
    let folder = format!("{}/{LOCAL_RULES}/", state_root::MOAI_DIR);
    plain_relative(path)
        && path.rsplit('/').next() != Some(GITIGNORE)
        && (ROOT_SOURCES.contains(&path)
            || path
                .strip_prefix(&folder)
                .is_some_and(|rest| !rest.is_empty()))
}

/// Why the receipt is none an apply of this project and stamp writes. The receipt is a file
/// anyone may have changed, so every value an undo acts on is held to what an apply produces; the
/// reason names the field and repeats no value.
fn check_receipt(
    receipt: &Receipt,
    root: &Path,
    state: &str,
    stamp: &str,
) -> Result<(), &'static str> {
    if receipt.schema_version != 1 {
        return Err("its schema version");
    }
    if receipt.operation != APPLY_OPERATION {
        return Err("its operation");
    }
    if receipt.project != root.to_string_lossy() {
        return Err("it is another project's");
    }
    if receipt.stamp != stamp {
        return Err("it is another apply's");
    }
    if !STATES.contains(&receipt.state.as_str()) {
        return Err("its state");
    }
    let area = area_rel(state, stamp);
    let mut seen = HashSet::new();
    for file in &receipt.moves {
        if !movable(&file.path) || !seen.insert(file.path.as_str()) {
            return Err("a moved path");
        }
        if file.backup != format!("{area}/{}", file.path) {
            return Err("a backup path");
        }
        let tracking = [Tracking::Tracked, Tracking::Untracked, Tracking::Unknown];
        if !is_digest(&file.sha256)
            || file.mode > 0o777
            || !tracking.iter().any(|t| t.word() == file.git)
        {
            return Err("the facts of a moved file");
        }
    }
    if !seen.contains(CLAUDE_MD) {
        return Err("it moves no root instruction file");
    }
    let written: Vec<&str> = receipt.writes.iter().map(|w| w.path.as_str()).collect();
    if written != [CLAUDE_MD] && written != [CLAUDE_MD, CLAUDE_LOCAL_MD] {
        return Err("a written path");
    }
    if receipt
        .writes
        .iter()
        .any(|w| !is_digest(&w.sha256) || w.mode > 0o777)
    {
        return Err("the facts of a written file");
    }
    if let Some(rule) = &receipt.ignore {
        let before = match &rule.before_sha256 {
            Some(sha256) => is_digest(sha256),
            None => rule.before_len == 0,
        };
        if rule.path != GITIGNORE
            || rule.line != CLAUDE_LOCAL_MD
            || !is_digest(&rule.after_sha256)
            || !before
            || !written.contains(&CLAUDE_LOCAL_MD)
        {
            return Err("its ignore rule");
        }
    }
    Ok(())
}

/// Undoes the apply whose receipt is `stamp`. Everything that can refuse is asked before the
/// first write. What cannot be undone is left, named, and a later run tries it again; a receipt
/// with nothing left to undo is answered without a write.
pub fn revert(project: &Path, stamp: &str, permission_reference: &str) -> Result<Value, String> {
    if !accepted_reference(permission_reference) {
        return Err(revert_refused(
            "permission reference required: --permission-reference takes a short line that names \
             the user's approval, neither the printed placeholder nor a credential",
        ));
    }
    if !state_root::utc_stamp(stamp) {
        return Err(revert_refused(
            "invalid receipt name: --receipt takes the id the apply answered, YYYYMMDDTHHMMSSZ",
        ));
    }
    let root = fs::canonicalize(project).map_err(|_| contract::PROJECT_PATH_UNAVAILABLE)?;
    let state = state_root::dir(&root);
    let rel = receipt_rel(state, stamp);
    let text = read_bounded_text(&root, &rel, RECEIPT_MAX)
        .map_err(|e| revert_refused(&e))?
        .ok_or_else(|| revert_refused(&format!("receipt not found: {rel}")))?;
    let receipt: Receipt = serde_json::from_str(&text)
        .map_err(|_| revert_refused(&format!("{rel} is no receipt of this step")))?;
    check_receipt(&receipt, &root, state, stamp).map_err(|what| {
        revert_refused(&format!("{rel} is not a receipt an apply writes ({what})"))
    })?;

    let mut journal = Journal {
        path: root.join(&rel),
        rel,
        saved: text.into_bytes(),
        receipt,
    };
    let finished = [RESTORED, REVERTED].contains(&journal.receipt.state.as_str());
    let done = if finished {
        Undone::default()
    } else {
        let done = undo(&root, &mut journal.receipt);
        journal.receipt.state = if done.clean() { REVERTED } else { PARTIAL }.into();
        journal.save().map_err(|e| {
            format!(
                "partial apply: the revert put back {} paths, removed {} and left {}, but the \
                 receipt {} could not be updated ({e}); run the same revert again",
                done.restored.len(),
                done.removed.len(),
                done.reasons.len(),
                journal.rel,
            )
        })?;
        done
    };
    Ok(json!({
        "state": if done.clean() { REVERTED } else { "conflict" },
        "receipt": stamp,
        "receipt_path": journal.rel,
        "receipt_state": journal.receipt.state,
        "restored": done.restored,
        "removed": done.removed,
        "conflicts": done.conflicts,
        "missing": done.missing,
        "reasons": done.reasons,
        "note": finished.then_some("nothing was left to undo; nothing was written"),
        "notice": RESTART_NOTICE,
    }))
}
