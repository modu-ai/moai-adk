//! The instruction files a Go-generation project carries (`claude-md-plan`): whether the root
//! `CLAUDE.md` is a Go-generation file, the sources in their fixed order with their facts, and
//! the lines of each that are the project's own. Everything here reads only: no file is
//! written, a link is never followed and only read-only git commands run. Line text stays in
//! [`Source::text`] and is never part of the answer.

use crate::go_residue::{find_manifest, Recorded};
use crate::go_state::linked_component;
use crate::{agents, identity, state_root};
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    collections::{BTreeMap, HashSet},
    fs,
    io::{self, BufRead, BufReader, ErrorKind, Read, Write},
    path::Path,
    process::{Command, Stdio},
};

/// A source above this many bytes is refused.
pub const SOURCE_LIMIT: u64 = 1024 * 1024;
/// The history search compares at most this many versions of one path, newest first.
pub const HISTORY_LIMIT: usize = 200;
/// The root instruction file Claude Code loads.
const CLAUDE_MD: &str = "CLAUDE.md";
/// The root sources in the order they are listed; the local rules files follow them.
const ROOT_SOURCES: [&str; 4] = [CLAUDE_MD, "AGENTS.md", "AGENTS.local.md", "CLAUDE.local.md"];
/// The personal files: one that is exactly a first text of moai holds nothing the user wrote.
const PERSONAL: [&str; 2] = ["AGENTS.local.md", "CLAUDE.local.md"];
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
    /// The row of the answer: facts and line numbers, never line text.
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

/// The answer of `claude-md-plan`.
pub fn plan(project: &Path) -> Result<Value, String> {
    let preview = preview(project)?;
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
    }))
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
