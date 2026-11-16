//! Versioning schemes of a project (card C-69): SemVer, HeadVer and CalVer.
//!
//! Three parts. [`classify`] tells the shape of one version string, a pure function.
//! [`fact`] is what `project-detect` reports as `versioning`: the scheme the project's tag
//! names (or, without a tag that has a shape, the root manifests) show, with the evidence it
//! was read from. [`VersioningSettings`] is the `versioning` key of the project profile, the
//! answer of the `/moai project` interview, and [`render_line`] is the one line the
//! `AGENTS.md` managed section carries for it.
//!
//! How tag names are read (audit of 5433d7e007, leader decision): from files only. No program
//! is started, and no git object and no git configuration is read, so nothing a repository
//! declares can run and nothing is written.
//!
//! - `<project>/.git` is a real folder: the names are the regular files under `refs/tags`
//!   (names only, never their content; a link is no tag and is not followed; at most
//!   [`MAX_TAG_ENTRIES`] directory entries, names of at most [`MAX_TAG_PARTS`] parts) and the
//!   `<id> refs/tags/<name>` lines of `packed-refs` (a regular file of at most
//!   [`MAX_PACKED`] bytes; peeled `^` lines are skipped).
//! - `<project>/.git` is a file: it is followed only for a linked worktree, that is a
//!   `gitdir: <path>` line whose folder holds a regular `commondir` file and a `gitdir` file
//!   that names `<project>/.git` back. The tags are then read from the common folder
//!   `commondir` names, by the same rules. The folders the two pointers name are checked
//!   themselves (a real folder, not a link); the path leading to them is not.
//! - Anything else — `.git` as a link, a submodule's pointer, a pointer to a repository that
//!   does not name this folder back, `refs`, `refs/tags` or `packed-refs` as a link or a
//!   special file, a file or a folder over its limit — is `tags: unavailable` with a
//!   `tags_reason`, and no tag is read. No `.git` entry is `tags: no_repository`: a parent
//!   folder's repository is not the project's.
//!
//! "Highest" is version order, not date: the numeric segments compared left to right, a
//! release above its own pre-release. Only a name that is short, plain text ([`showable`]) is
//! ever considered or shown.
//!
//! The rules the skill follows are in `skills/moai/references/internal/versioning.md`.

use serde::{Deserialize, Serialize};
use serde_json::{json, Map, Value};
use std::{
    cmp::Ordering,
    fs,
    io::Read,
    path::{Path, PathBuf},
};

type Result<T> = std::result::Result<T, String>;

/// The shape of one version string.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Shape {
    Semver,
    Calver,
    Headver,
    /// Fits two readings; it is reported and never counted.
    Ambiguous,
    Unrecognised,
}

impl Shape {
    pub fn name(self) -> &'static str {
        match self {
            Shape::Semver => "semver",
            Shape::Calver => "calver",
            Shape::Headver => "headver",
            Shape::Ambiguous => "ambiguous",
            Shape::Unrecognised => "unrecognised",
        }
    }
}

/// Highest tags shown as evidence.
pub const MAX_TAGS: usize = 5;
/// Most directory entries looked at under `refs/tags`; one more and no tag is read.
pub const MAX_TAG_ENTRIES: usize = 2000;
/// Most `/`-separated parts of a loose tag name; deeper folders are not entered.
pub const MAX_TAG_PARTS: usize = 4;
/// Largest `packed-refs` read; one byte more and no tag is read.
pub const MAX_PACKED: u64 = 4 * 1024 * 1024;
/// Largest `.git`, `commondir` or `gitdir` pointer file read.
const MAX_POINTER: u64 = 4096;
/// Longest tag name or manifest version kept as evidence, in bytes.
const MAX_VALUE: usize = 64;
/// Longest `custom` text in the profile, in characters.
pub const MAX_CUSTOM_CHARS: usize = 128;
/// Largest HeadVer head the profile takes.
pub const MAX_HEAD: u64 = 9999;
/// Largest manifest read for its version.
const MAX_FILE: u64 = 256 * 1024;

const NOTE_SEMVER_OR_CALVER: &str =
    "valid SemVer and a possible CalVer (two-digit year, month); not counted";
const NOTE_TWO_PART_OR_CALVER: &str =
    "a possible CalVer (two-digit year, month) or a plain two-part version; not counted";
const NOTE_HEADVER_OR_CALVER: &str =
    "a HeadVer year-week segment behind a four-digit year; not counted";
const NOTE_HEADVER_ALSO_SEMVER: &str =
    "also valid SemVer; beside SemVer tags it is counted as SemVer";

/// A two-digit year is taken as one from 10 (2010): below that a first segment is far more
/// often a plain number.
const FIRST_SHORT_YEAR: u64 = 10;
const FIRST_FULL_YEAR: u64 = 2000;

/// The numeric segments of a version and whether a pre-release follows them. A leading `v`
/// and a prefix that ends in `/` or `@` (`release/v1.2.3`, `pkg@1.2.3`) are dropped; what
/// follows the first `-` or `+` is not read beyond telling a pre-release (`-`) apart.
/// `None` unless there are two to four all-digit segments of at most nine digits.
fn parts(text: &str) -> Option<(Vec<&str>, bool)> {
    let name = text.rsplit(['/', '@']).next().unwrap_or(text);
    let name = name.strip_prefix(['v', 'V']).unwrap_or(name);
    let end = name.find(['-', '+']).unwrap_or(name.len());
    let (core, tail) = name.split_at(end);
    let segments: Vec<&str> = core.split('.').collect();
    let numeric = |s: &&str| !s.is_empty() && s.len() <= 9 && s.bytes().all(|b| b.is_ascii_digit());
    ((2..=4).contains(&segments.len()) && segments.iter().all(numeric))
        .then(|| (segments, tail.starts_with('-')))
}

fn number(segment: &str) -> u64 {
    segment.parse().unwrap_or(u64::MAX)
}

/// The shape of one version string and the note its evidence row carries.
fn shape_of(text: &str) -> (Shape, Option<&'static str>) {
    let Some((segments, _)) = parts(text) else {
        return (Shape::Unrecognised, None);
    };
    // SemVer 2.0.0 rule 2: no leading zeros.
    let plain = |s: &&str| *s == "0" || !s.starts_with('0');
    let semver = segments.len() == 3 && segments.iter().all(plain);
    let full_year = segments[0].len() == 4 && number(segments[0]) >= FIRST_FULL_YEAR;
    let short_year = segments[0].len() == 2
        && number(segments[0]) >= FIRST_SHORT_YEAR
        && segments[1].len() <= 2
        && (1..=12).contains(&number(segments[1]));
    // With only two segments a short year is told from a plain `16.4` by the padded month
    // alone (`26.04`); `16.10` could be either.
    let padded_month = segments[1].len() == 2 && segments[1].starts_with('0');
    let headver = segments.len() == 3
        && plain(&segments[0])
        && plain(&segments[2])
        && segments[1].len() == 4
        && number(&segments[1][..2]) >= FIRST_SHORT_YEAR
        && (1..=53).contains(&number(&segments[1][2..]));
    match (headver, full_year, short_year, semver) {
        (true, true, ..) => (Shape::Ambiguous, Some(NOTE_HEADVER_OR_CALVER)),
        (true, ..) => (Shape::Headver, Some(NOTE_HEADVER_ALSO_SEMVER)),
        (_, true, ..) => (Shape::Calver, None),
        (_, _, true, true) => (Shape::Ambiguous, Some(NOTE_SEMVER_OR_CALVER)),
        (_, _, true, false) if segments.len() == 2 && !padded_month => {
            (Shape::Ambiguous, Some(NOTE_TWO_PART_OR_CALVER))
        }
        (_, _, true, false) => (Shape::Calver, None),
        (_, _, _, true) => (Shape::Semver, None),
        _ => (Shape::Unrecognised, None),
    }
}

/// The shape of one version string: a tag name or a manifest version.
pub fn classify(text: &str) -> Shape {
    shape_of(text).0
}

/// The scheme a set of shapes shows: the one scheme, `mixed` for two, `None` when no shape is
/// one of the three. Every HeadVer version is valid SemVer too, so HeadVer shapes beside
/// SemVer ones are SemVer (a minor that walks through `1001`), not a second scheme.
fn scheme_of(shapes: impl Iterator<Item = Shape>) -> Option<&'static str> {
    let mut found: Vec<Shape> = Vec::new();
    for shape in shapes {
        let definite = matches!(shape, Shape::Semver | Shape::Calver | Shape::Headver);
        if definite && !found.contains(&shape) {
            found.push(shape);
        }
    }
    if found.contains(&Shape::Semver) {
        found.retain(|s| *s != Shape::Headver);
    }
    match found.as_slice() {
        [] => None,
        [one] => Some(one.name()),
        _ => Some("mixed"),
    }
}

/// Where a tag stands among the evidence: versions of a definite shape first, then ambiguous
/// ones, then names that are no version; inside the first two the higher version first.
fn rank(name: &str) -> (u8, Vec<u64>, bool) {
    let tier = match classify(name) {
        Shape::Semver | Shape::Calver | Shape::Headver => 0,
        Shape::Ambiguous => 1,
        Shape::Unrecognised => 2,
    };
    match parts(name) {
        Some((segments, prerelease)) if tier < 2 => (
            tier,
            segments.iter().map(|s| number(s)).collect(),
            !prerelease,
        ),
        _ => (tier, Vec::new(), false),
    }
}

/// Highest first: by tier, then numeric segments, a release above its pre-release, and the
/// name itself last so the order never depends on how the names arrived.
fn by_rank(a: &&String, b: &&String) -> Ordering {
    let (ta, sa, ra) = rank(a);
    let (tb, sb, rb) = rank(b);
    ta.cmp(&tb)
        .then_with(|| sb.cmp(&sa))
        .then_with(|| rb.cmp(&ra))
        .then_with(|| if ta < 2 { b.cmp(a) } else { a.cmp(b) })
}

/// The versioning fact from tag names (in any order) and manifest versions
/// (`(file, version)`). The tags decide when one of them has a definite shape; the manifests
/// decide only without that. `evidence` holds the [`MAX_TAGS`] highest tags, then the
/// manifests. `latest` is the highest tag that is a version, else the first manifest version
/// that is one.
pub fn summarize(tags: &[String], manifests: &[(&str, String)]) -> Value {
    let row = |source: &str, value: &str| {
        let (shape, note) = shape_of(value);
        let mut row = json!({"source": source, "value": value, "shape": shape.name()});
        if let Some(note) = note {
            row["note"] = json!(note);
        }
        row
    };
    let mut ranked: Vec<&String> = tags.iter().collect();
    ranked.sort_by(by_rank);
    ranked.dedup();
    let evidence: Vec<Value> = ranked
        .iter()
        .take(MAX_TAGS)
        .map(|t| row("tag", t))
        .chain(manifests.iter().map(|(file, v)| row(file, v)))
        .collect();
    let scheme = scheme_of(tags.iter().map(|t| classify(t)))
        .or_else(|| scheme_of(manifests.iter().map(|(_, v)| classify(v))))
        .unwrap_or("none");
    let latest = ranked
        .first()
        .copied()
        .into_iter()
        .chain(manifests.iter().map(|(_, v)| v))
        .find(|v| classify(v) != Shape::Unrecognised);
    json!({"scheme": scheme, "latest": latest, "evidence": evidence})
}

/// Text short and plain enough to be considered and shown: a tag name or a manifest version
/// of ASCII letters, digits and `. - _ + / @`, at most [`MAX_VALUE`] bytes.
fn showable(value: &str) -> bool {
    !value.is_empty()
        && value.len() <= MAX_VALUE
        && value.bytes().all(|b| {
            b.is_ascii_alphanumeric() || matches!(b, b'.' | b'-' | b'_' | b'+' | b'/' | b'@')
        })
}

/// The bytes of the regular file at `path`, at most `max`: `Ok(None)` when nothing is there,
/// `Err` with the reason for a link, a folder, a pipe or device, a larger file or a failed
/// read. The file is opened without following a link and without waiting (a pipe put there
/// between the check and the open cannot hold the read), and what was opened is checked
/// again.
fn read_regular(path: &Path, max: u64) -> Result<Option<Vec<u8>>> {
    let meta = match fs::symlink_metadata(path) {
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(None),
        Err(_) => return Err("could not be read".into()),
        Ok(meta) => meta,
    };
    if !meta.file_type().is_file() {
        return Err("is not a regular file".into());
    }
    if meta.len() > max {
        return Err(format!("is larger than {max} bytes"));
    }
    let mut options = fs::OpenOptions::new();
    options.read(true);
    #[cfg(unix)]
    {
        use std::os::unix::fs::OpenOptionsExt;
        options.custom_flags(libc::O_NOFOLLOW | libc::O_NONBLOCK);
    }
    let file = options.open(path).map_err(|_| "could not be read")?;
    if !file.metadata().is_ok_and(|m| m.is_file()) {
        return Err("is not a regular file".into());
    }
    let mut bytes = Vec::new();
    file.take(max + 1)
        .read_to_end(&mut bytes)
        .map_err(|_| "could not be read")?;
    if bytes.len() as u64 > max {
        return Err(format!("is larger than {max} bytes"));
    }
    Ok(Some(bytes))
}

/// The first line of a small pointer file, trimmed; `None` when the file is missing, is not
/// a small regular file or is not UTF-8.
fn pointer(path: &Path) -> Option<String> {
    let bytes = read_regular(path, MAX_POINTER).ok()??;
    let text = String::from_utf8(bytes).ok()?;
    Some(text.lines().next().unwrap_or("").trim().to_string())
}

/// `true` for a real folder; a link to one is not.
fn real_dir(path: &Path) -> bool {
    fs::symlink_metadata(path).is_ok_and(|m| m.is_dir())
}

/// The folder that holds the project's refs: `<root>/.git` when it is a real folder, the
/// common folder of the repository when `<root>/.git` is the pointer of a linked worktree
/// (module comment). `Ok(None)` when the project has no `.git` entry.
fn refs_home(root: &Path) -> Result<Option<PathBuf>> {
    let dot_git = root.join(".git");
    let meta = match fs::symlink_metadata(&dot_git) {
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(None),
        Err(_) => return Err(".git could not be read".into()),
        Ok(meta) => meta,
    };
    if meta.file_type().is_symlink() {
        return Err(".git is a symbolic link".into());
    }
    if meta.is_dir() {
        return Ok(Some(dot_git));
    }
    let not_a_worktree = ".git is a file that is not the pointer of a linked worktree";
    let target = pointer(&dot_git)
        .and_then(|line| line.strip_prefix("gitdir:").map(|p| p.trim().to_string()))
        .filter(|p| !p.is_empty())
        .ok_or(not_a_worktree)?;
    let admin = root.join(target);
    if !real_dir(&admin) {
        return Err("the folder the .git file points at is missing or a link".into());
    }
    let common = pointer(&admin.join("commondir"))
        .filter(|p| !p.is_empty())
        .ok_or(not_a_worktree)?;
    // The repository must name this folder as its worktree, or the pointer could borrow
    // the tags of any repository.
    let back = pointer(&admin.join("gitdir")).unwrap_or_default();
    let back = Path::new(&back);
    let same_folder = |a: &Path, b: &Path| match (a.canonicalize(), b.canonicalize()) {
        (Ok(a), Ok(b)) => a == b,
        _ => false,
    };
    let names_back = back.file_name().is_some_and(|n| n == ".git")
        && back.parent().is_some_and(|p| same_folder(p, root));
    if !names_back {
        return Err(
            "the repository the .git file points at does not name this folder as its worktree"
                .into(),
        );
    }
    let common = admin.join(common);
    if !real_dir(&common) {
        return Err("the common folder of the linked worktree is missing or a link".into());
    }
    Ok(Some(common))
}

/// The names of the regular files under `<home>/refs/tags`. Content is never read.
fn loose_tags(home: &Path, names: &mut Vec<String>) -> Result<()> {
    let refs = home.join("refs");
    let tags = refs.join("tags");
    for (dir, label) in [(&refs, "refs"), (&tags, "refs/tags")] {
        match fs::symlink_metadata(dir) {
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(()),
            Ok(meta) if meta.is_dir() => {}
            _ => return Err(format!("{label} is not a real folder")),
        }
    }
    let mut seen = 0usize;
    // Each folder with the name prefix of its files and how many parts those names have.
    let mut stack: Vec<(PathBuf, String, usize)> = vec![(tags, String::new(), 1)];
    while let Some((dir, prefix, name_parts)) = stack.pop() {
        let entries = fs::read_dir(&dir).map_err(|_| "refs/tags could not be read")?;
        for entry in entries {
            let entry = entry.map_err(|_| "refs/tags could not be read")?;
            seen += 1;
            if seen > MAX_TAG_ENTRIES {
                return Err(format!(
                    "refs/tags holds more than {MAX_TAG_ENTRIES} entries"
                ));
            }
            // The entry's own kind: a link is neither a file nor a folder here.
            let Ok(kind) = entry.file_type() else {
                continue;
            };
            let name = format!("{prefix}{}", entry.file_name().to_string_lossy());
            if kind.is_file() {
                names.push(name);
            } else if kind.is_dir() && name_parts < MAX_TAG_PARTS {
                stack.push((entry.path(), format!("{name}/"), name_parts + 1));
            }
        }
    }
    Ok(())
}

/// The tag names in `<home>/packed-refs`: lines `<id> refs/tags/<name>` with a 40- or
/// 64-digit hexadecimal id. Comment lines, peeled `^` lines and other refs are skipped.
fn packed_tags(home: &Path, names: &mut Vec<String>) -> Result<()> {
    let Some(bytes) = read_regular(&home.join("packed-refs"), MAX_PACKED)
        .map_err(|e| format!("packed-refs {e}"))?
    else {
        return Ok(());
    };
    for line in bytes.split(|b| *b == b'\n') {
        let line = String::from_utf8_lossy(line);
        let Some((id, name)) = line.split_once(' ') else {
            continue;
        };
        let hex = [40, 64].contains(&id.len()) && id.bytes().all(|b| b.is_ascii_hexdigit());
        if let (true, Some(tag)) = (hex, name.strip_prefix("refs/tags/")) {
            names.push(tag.to_string());
        }
    }
    Ok(())
}

/// The project's tag names that are [`showable`], sorted and without duplicates, and how the
/// read went: `read`, `no_repository` or `unavailable` with the reason (module comment).
fn tag_names(root: &Path) -> (Vec<String>, &'static str, Option<String>) {
    let home = match refs_home(root) {
        Ok(Some(home)) => home,
        Ok(None) => return (Vec::new(), "no_repository", None),
        Err(reason) => return (Vec::new(), "unavailable", Some(reason)),
    };
    let mut names = Vec::new();
    if let Err(reason) = loose_tags(&home, &mut names).and_then(|()| packed_tags(&home, &mut names))
    {
        return (Vec::new(), "unavailable", Some(reason));
    }
    // The filter comes before any pick, so a name that is not shown takes no one's place.
    names.retain(|n| showable(n));
    names.sort();
    names.dedup();
    (names, "read", None)
}

/// The regular file `root/<name>` as text; `None` for a link, a folder, a large or missing file.
fn root_file(root: &Path, name: &str) -> Option<String> {
    let path = root.join(name);
    let meta = fs::symlink_metadata(&path).ok()?;
    if !meta.file_type().is_file() || meta.len() > MAX_FILE {
        return None;
    }
    fs::read_to_string(path).ok()
}

/// The version each root manifest states, in a fixed order: `Cargo.toml` (`[package]`, else
/// `[workspace.package]`), `package.json`, `pyproject.toml` (`[project]`, else
/// `[tool.poetry]`). A value that is not plain version text is left out.
fn manifest_versions(root: &Path) -> Vec<(&'static str, String)> {
    let toml =
        |name: &str| root_file(root, name).and_then(|t| t.parse::<toml_edit::DocumentMut>().ok());
    let text = |item: Option<&toml_edit::Item>| {
        item.and_then(|i| i.get("version"))
            .and_then(toml_edit::Item::as_str)
            .map(String::from)
    };
    let cargo = toml("Cargo.toml").and_then(|doc| {
        text(doc.get("package"))
            .or_else(|| text(doc.get("workspace").and_then(|w| w.get("package"))))
    });
    let package_json = root_file(root, "package.json")
        .and_then(|t| serde_json::from_str::<Value>(&t).ok())
        .and_then(|v| v["version"].as_str().map(String::from));
    let pyproject = toml("pyproject.toml").and_then(|doc| {
        text(doc.get("project")).or_else(|| text(doc.get("tool").and_then(|t| t.get("poetry"))))
    });
    [
        ("Cargo.toml", cargo),
        ("package.json", package_json),
        ("pyproject.toml", pyproject),
    ]
    .into_iter()
    .filter_map(|(file, version)| Some((file, version.filter(|v| showable(v))?)))
    .collect()
}

/// The `versioning` fact of `project-detect`: [`summarize`] over the project's tag names and
/// root manifest versions, plus `tags`, how the tag read went, and `tags_reason` when it is
/// `unavailable`. Reads files only; starts no program.
pub fn fact(root: &Path) -> Value {
    let (tags, status, reason) = tag_names(root);
    let mut fact = summarize(&tags, &manifest_versions(root));
    fact["tags"] = json!(status);
    if let Some(reason) = reason {
        fact["tags_reason"] = json!(reason);
    }
    fact
}

/// The schemes the profile accepts. `custom` is the user's own words, kept in `format`.
pub const SCHEMES: [&str; 4] = ["semver", "headver", "calver", "custom"];
/// The tokens of a CalVer format (calver.org, spec 26.0); `MM`, `WW` and `DD` are the
/// spelling the page used before that spec.
const CALVER_TOKENS: [&str; 16] = [
    "YYYY", "YY", "0Y", "M", "0M", "W", "0W", "D", "0D", "MM", "WW", "DD", "MAJOR", "MINOR",
    "MICRO", "MODIFIER",
];
const YEAR_TOKENS: [&str; 3] = ["YYYY", "YY", "0Y"];

/// The `versioning` key of the project profile: the scheme the project's releases follow, as
/// picked in the `/moai project` interview. It is read from a JSON object only: an array, a
/// string, an unknown field or a `null` value is refused.
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(try_from = "Map<String, Value>")]
pub struct VersioningSettings {
    /// One of [`SCHEMES`].
    pub scheme: String,
    /// The CalVer format (`YYYY.0M.MICRO`), or for `custom` the user's words in one line.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub format: Option<String>,
    /// The HeadVer head number.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub head: Option<u64>,
}

impl TryFrom<Map<String, Value>> for VersioningSettings {
    type Error = String;

    fn try_from(map: Map<String, Value>) -> Result<Self> {
        let refused = || "invalid versioning setting; no values were logged".to_string();
        let mut settings = VersioningSettings {
            scheme: String::new(),
            format: None,
            head: None,
        };
        let mut has_scheme = false;
        for (key, value) in map {
            match (key.as_str(), value) {
                ("scheme", Value::String(scheme)) => {
                    settings.scheme = scheme;
                    has_scheme = true;
                }
                ("format", Value::String(format)) => settings.format = Some(format),
                ("head", Value::Number(head)) => {
                    settings.head = Some(head.as_u64().ok_or_else(refused)?);
                }
                _ => return Err(refused()),
            }
        }
        has_scheme.then_some(settings).ok_or_else(refused)
    }
}

/// Two to four known tokens joined by dots, none twice, the first one the only year token.
fn calver_format(format: &str) -> bool {
    let tokens: Vec<&str> = format.split('.').collect();
    (2..=4).contains(&tokens.len())
        && tokens.iter().all(|t| CALVER_TOKENS.contains(t))
        && tokens
            .iter()
            .enumerate()
            .all(|(i, t)| !tokens[..i].contains(t))
        && YEAR_TOKENS.contains(&tokens[0])
        && tokens[1..].iter().all(|t| !YEAR_TOKENS.contains(t))
}

/// One line of the user's own words, at most [`MAX_CUSTOM_CHARS`] characters, that cannot be
/// read as markup or hide text where the managed section shows it: no control character, no
/// `<`, `>` or backtick, no `#` at the start, no invisible or direction-changing character.
fn custom_text(text: &str) -> bool {
    let hidden = |c: char| {
        matches!(c, '\u{2028}' | '\u{2029}' | '\u{202E}') || ('\u{200B}'..='\u{200F}').contains(&c)
    };
    !text.trim().is_empty()
        && text.chars().count() <= MAX_CUSTOM_CHARS
        && !text.trim_start().starts_with('#')
        && !text
            .chars()
            .any(|c| c.is_control() || matches!(c, '<' | '>' | '`') || hidden(c))
}

/// Accepts a `versioning` value: a known scheme, a `format` only on CalVer (a token format) or
/// `custom` (one bounded line), a `head` of at most [`MAX_HEAD`] only on HeadVer. No value is
/// echoed.
pub fn validate(v: &VersioningSettings) -> Result<()> {
    let scheme = v.scheme.as_str();
    let format_ok = match (scheme, v.format.as_deref()) {
        (_, None) => true,
        ("calver", Some(format)) => calver_format(format),
        ("custom", Some(text)) => custom_text(text),
        _ => false,
    };
    let head_ok = !v.head.is_some_and(|h| scheme != "headver" || h > MAX_HEAD);
    if !SCHEMES.contains(&scheme) || !format_ok || !head_ok {
        return Err("invalid versioning setting; no values were logged".into());
    }
    Ok(())
}

/// The line the `AGENTS.md` managed section carries for the choice, in the section's
/// language (`ko`, else English), newline included.
pub fn render_line(v: &VersioningSettings, lang: &str) -> String {
    let label = if lang == "ko" {
        "버전 규칙"
    } else {
        "Versioning"
    };
    let scheme = match (v.scheme.as_str(), v.format.as_deref(), v.head) {
        ("semver", ..) => "SemVer (MAJOR.MINOR.PATCH)".to_string(),
        ("headver", _, Some(head)) => {
            format!("HeadVer ({{head}}.{{yearweek}}.{{build}}), head {head}")
        }
        ("headver", ..) => "HeadVer ({head}.{yearweek}.{build})".to_string(),
        ("calver", Some(format), _) => format!("CalVer ({format})"),
        ("calver", ..) => "CalVer".to_string(),
        (_, Some(text), _) => format!("custom ({text})"),
        _ => "custom".to_string(),
    };
    format!("{label}: {scheme}.\n")
}

#[cfg(test)]
mod tests {
    use super::*;

    fn shapes(list: &[&str]) -> Vec<&'static str> {
        list.iter().map(|t| classify(t).name()).collect()
    }

    #[test]
    fn the_three_schemes_are_told_apart() {
        for semver in [
            "v3.2.1",
            "3.2.1",
            "0.1.0",
            "1.0.0-rc.1",
            "1.0.0+build.7",
            "pkg@1.4.0",
            "release/v2.0.1",
            "100.4.1",
            "1.999.0",
            "1.1000.0",
        ] {
            assert_eq!(classify(semver), Shape::Semver, "{semver}");
        }
        for calver in [
            "2026.11.0",
            "v2026.11.0-rc.1",
            "2026.03.17",
            "2026.3",
            "26.04",
            "26.04.1",
            "2026.11.17.1",
        ] {
            assert_eq!(classify(calver), Shape::Calver, "{calver}");
        }
        for headver in ["3.2647.59", "v0.2134.1", "12.2601.4+meta", "1.1001.0"] {
            assert_eq!(classify(headver), Shape::Headver, "{headver}");
        }
        for other in [
            "nightly",
            "1.2",
            "1.2.3.4",
            "01.2.3",
            "1..2",
            "",
            "v",
            "1.2.x",
            "1.2.3.4.5",
            "2026",
        ] {
            assert_eq!(classify(other), Shape::Unrecognised, "{other:?}");
        }
    }

    /// A shape that fits two readings is named ambiguous, never guessed.
    #[test]
    fn a_shape_that_fits_two_readings_is_ambiguous() {
        // Valid SemVer and a possible CalVer `YY.M.MICRO`.
        assert_eq!(shapes(&["26.4.1", "22.12.0", "10.1.0"]), ["ambiguous"; 3]);
        // A HeadVer middle segment behind a four-digit year.
        assert_eq!(classify("2026.2647.59"), Shape::Ambiguous);
        // Audit N1: two segments are CalVer only with a zero-padded month; `v16.4` and
        // `16.10` are as likely a plain two-part version.
        assert_eq!(
            shapes(&["v16.4", "26.4", "16.10", "26.12", "26.04", "26.09"]),
            [
                "ambiguous",
                "ambiguous",
                "ambiguous",
                "ambiguous",
                "calver",
                "calver"
            ]
        );
    }

    /// Each limit at its edge: the last value inside and the first outside.
    #[test]
    fn the_limits_hold_at_their_edges() {
        // A four-digit year starts at 2000.
        assert_eq!(shapes(&["2000.1", "1999.1"]), ["calver", "unrecognised"]);
        assert_eq!(shapes(&["2000.1.0", "1999.1.0"]), ["calver", "semver"]);
        // A two-digit year needs a month from 1 to 12 behind it.
        assert_eq!(
            shapes(&["26.12", "26.13", "26.0", "26.00"]),
            ["ambiguous", "unrecognised", "unrecognised", "unrecognised"]
        );
        assert_eq!(shapes(&["26.12.1", "26.13.1"]), ["ambiguous", "semver"]);
        // A two-digit year starts at 10.
        assert_eq!(shapes(&["10.04", "09.04"]), ["calver", "unrecognised"]);
        // The HeadVer week runs from 01 to 53, behind a two-digit year from 10.
        assert_eq!(
            shapes(&["3.2601.5", "3.2653.5", "3.2654.5", "3.2600.5"]),
            ["headver", "headver", "semver", "semver"]
        );
        assert_eq!(
            shapes(&["3.1001.5", "3.0901.5"]),
            ["headver", "unrecognised"]
        );
        // The middle segment has exactly four digits.
        assert_eq!(shapes(&["3.264.5", "3.26471.5"]), ["semver", "semver"]);
    }

    fn tags(list: &[&str]) -> Vec<String> {
        list.iter().map(|t| t.to_string()).collect()
    }

    fn values(v: &Value) -> Vec<&str> {
        v["evidence"]
            .as_array()
            .unwrap()
            .iter()
            .map(|e| e["value"].as_str().unwrap())
            .collect()
    }

    #[test]
    fn tags_of_one_scheme_name_it_and_the_highest_is_latest() {
        let v = summarize(&tags(&["v3.2.0", "nightly", "v3.2.1"]), &[]);
        assert_eq!(v["scheme"], "semver");
        assert_eq!(v["latest"], "v3.2.1");
        assert_eq!(
            v["evidence"],
            json!([
                {"source": "tag", "value": "v3.2.1", "shape": "semver"},
                {"source": "tag", "value": "v3.2.0", "shape": "semver"},
                {"source": "tag", "value": "nightly", "shape": "unrecognised"},
            ])
        );
        assert_eq!(summarize(&tags(&["2026.11.0"]), &[])["scheme"], "calver");
        assert_eq!(summarize(&tags(&["3.2647.59"]), &[])["scheme"], "headver");
    }

    /// The order is the version's, whatever order the names arrive in: numbers compare as
    /// numbers, a release stands above its pre-release, and the five highest are evidence.
    #[test]
    fn evidence_is_the_five_highest_versions() {
        let names = [
            "v1.9.0",
            "v1.10.0-rc.1",
            "v2.0.0",
            "v1.10.0",
            "v0.9.9",
            "nightly",
            "v1.2.3",
            "v1.2.10",
        ];
        let v = summarize(&tags(&names), &[]);
        assert_eq!(v["latest"], "v2.0.0");
        assert_eq!(
            values(&v),
            ["v2.0.0", "v1.10.0", "v1.10.0-rc.1", "v1.9.0", "v1.2.10"]
        );
        let mut reversed = names;
        reversed.reverse();
        assert_eq!(summarize(&tags(&reversed), &[]), v);
        // A name that is no version is shown only when there is room, after the versions.
        let v = summarize(&tags(&["zeta", "v1.0.0", "alpha", "26.4.1"]), &[]);
        assert_eq!(values(&v), ["v1.0.0", "26.4.1", "alpha", "zeta"]);
    }

    #[test]
    fn tags_of_two_schemes_are_mixed_and_nothing_is_none() {
        let v = summarize(&tags(&["v3.2.1", "2026.11.0"]), &[]);
        assert_eq!(v["scheme"], "mixed");
        assert_eq!(v["latest"], "2026.11.0");
        assert_eq!(
            summarize(&tags(&["3.2647.59", "2026.11.0"]), &[])["scheme"],
            "mixed"
        );
        let v = summarize(&[], &[]);
        assert_eq!(v, json!({"scheme": "none", "latest": null, "evidence": []}));
        // Tags that are no versions are evidence and decide nothing.
        let v = summarize(&tags(&["nightly", "latest"]), &[]);
        assert_eq!(v["scheme"], "none");
        assert_eq!(v["latest"], Value::Null);
        assert_eq!(v["evidence"].as_array().unwrap().len(), 2);
    }

    /// Audit N2: every HeadVer version is valid SemVer. Beside SemVer tags a HeadVer shape is
    /// SemVer (a minor that walks through 999, 1000, 1001), and its row says so.
    #[test]
    fn a_headver_shape_beside_semver_tags_is_semver() {
        let v = summarize(&tags(&["v1.999.0", "v1.1000.0", "v1.1001.0"]), &[]);
        assert_eq!(v["scheme"], "semver");
        assert_eq!(v["latest"], "v1.1001.0");
        let row = &v["evidence"][0];
        assert_eq!(row["shape"], "headver");
        assert!(
            row["note"].as_str().unwrap().contains("also valid SemVer"),
            "{row}"
        );
        // HeadVer shapes alone stay HeadVer, with the same note.
        let v = summarize(&tags(&["3.2646.58", "3.2647.59"]), &[]);
        assert_eq!(v["scheme"], "headver");
        assert!(v["evidence"][0]["note"]
            .as_str()
            .unwrap()
            .contains("also valid SemVer"));
    }

    /// An ambiguous shape says so in its evidence row and is not counted.
    #[test]
    fn an_ambiguous_version_is_reported_and_not_counted() {
        let v = summarize(&tags(&["26.4.1"]), &[]);
        assert_eq!(v["scheme"], "none");
        assert_eq!(v["latest"], "26.4.1");
        let row = &v["evidence"][0];
        assert_eq!(row["shape"], "ambiguous");
        assert!(
            row["note"].as_str().unwrap().contains("SemVer")
                && row["note"].as_str().unwrap().contains("CalVer"),
            "{row}"
        );
        let v = summarize(&tags(&["v16.4"]), &[]);
        assert_eq!(v["scheme"], "none");
        assert!(v["evidence"][0]["note"]
            .as_str()
            .unwrap()
            .contains("two-part version"));
        // Beside a definite shape the definite one decides and stands first.
        let v = summarize(&tags(&["26.4.1", "26.04.0"]), &[]);
        assert_eq!(v["scheme"], "calver");
        assert_eq!(v["latest"], "26.04.0");
        // Rows of CalVer and SemVer shape carry no note.
        assert!(v["evidence"][0].get("note").is_none());
    }

    /// Tags decide when one of them has a shape; the manifests decide only without that.
    #[test]
    fn manifests_decide_only_without_a_recognised_tag() {
        let manifests = [("Cargo.toml", "0.3.1".to_string())];
        let v = summarize(&[], &manifests);
        assert_eq!(v["scheme"], "semver");
        assert_eq!(v["latest"], "0.3.1");
        assert_eq!(
            v["evidence"],
            json!([{"source": "Cargo.toml", "value": "0.3.1", "shape": "semver"}])
        );
        let v = summarize(&tags(&["2026.11.0"]), &manifests);
        assert_eq!(v["scheme"], "calver");
        assert_eq!(v["latest"], "2026.11.0");
        assert_eq!(v["evidence"].as_array().unwrap().len(), 2);
        let two = [
            ("Cargo.toml", "0.3.1".to_string()),
            ("package.json", "2026.11.0".to_string()),
        ];
        let v = summarize(&tags(&["nightly"]), &two);
        assert_eq!(v["scheme"], "mixed");
        assert_eq!(v["latest"], "0.3.1");
    }

    fn settings(scheme: &str, format: Option<&str>, head: Option<u64>) -> VersioningSettings {
        VersioningSettings {
            scheme: scheme.into(),
            format: format.map(String::from),
            head,
        }
    }

    #[test]
    fn settings_accept_the_four_schemes_and_refuse_the_rest() {
        for ok in [
            settings("semver", None, None),
            settings("headver", None, None),
            settings("headver", None, Some(0)),
            settings("headver", None, Some(3)),
            settings("calver", None, None),
            settings("calver", Some("YYYY.0M.MICRO"), None),
            settings("calver", Some("YY.0M.MICRO"), None),
            settings("calver", Some("YYYY.0M.0D"), None),
            settings("calver", Some("YY.0M"), None),
            settings("calver", Some("YY.MINOR.MICRO"), None),
            settings("calver", Some("YYYY.0M.0D.MICRO"), None),
            // The spelling calver.org used before spec 26.0 is still read.
            settings("calver", Some("YYYY.MM.DD"), None),
            settings("custom", None, None),
            settings(
                "custom",
                Some("release train letter plus build number"),
                None,
            ),
            settings("custom", Some("build #42 of the week"), None),
        ] {
            assert!(validate(&ok).is_ok(), "{ok:?}");
        }
        for bad in [
            settings("", None, None),
            settings("SemVer", None, None),
            settings("romantic", None, None),
            // A format belongs to CalVer or custom, a head to HeadVer.
            settings("semver", Some("MAJOR.MINOR.PATCH"), None),
            settings("semver", None, Some(1)),
            settings("calver", None, Some(1)),
            settings("headver", Some("YYYY.0M.MICRO"), None),
            // A CalVer format is two to four known tokens.
            settings("calver", Some("YYYY"), None),
            settings("calver", Some("MAJOR.MINOR.MICRO"), None),
            settings("calver", Some("YYYY.0M.0D.MICRO.MODIFIER"), None),
            settings("calver", Some("YYYY.Q.MICRO"), None),
            settings("calver", Some("yyyy.0m"), None),
            settings("calver", Some(""), None),
            // Audit N7: no token twice, one year token, and it comes first.
            settings("calver", Some("YYYY.0M.0M"), None),
            settings("calver", Some("YYYY.MICRO.MICRO"), None),
            settings("calver", Some("YYYY.YY.MICRO"), None),
            settings("calver", Some("YYYY.0Y"), None),
            settings("calver", Some("MAJOR.YYYY.MICRO"), None),
            settings("calver", Some("0M.YYYY"), None),
            // Custom text is one line.
            settings("custom", Some(""), None),
            settings("custom", Some("   "), None),
            settings("custom", Some("a\nb"), None),
            // Audit N8: nothing that reads as markup or hides text.
            settings("custom", Some("x <b> y"), None),
            settings("custom", Some("x < y"), None),
            settings("custom", Some("x > y"), None),
            settings("custom", Some("run `id`"), None),
            settings("custom", Some("# Heading"), None),
            settings("custom", Some("  ## Heading"), None),
            settings("custom", Some("a\u{2028}b"), None),
            settings("custom", Some("a\u{2029}b"), None),
            settings("custom", Some("a\u{202e}b"), None),
            settings("custom", Some("a\u{200b}b"), None),
            settings("custom", Some("a\u{200d}b"), None),
            settings("custom", Some("a\u{200f}b"), None),
        ] {
            assert!(validate(&bad).is_err(), "{bad:?}");
        }
        // Audit N3: the custom limit counts characters, not bytes.
        let korean = |n: usize| "\u{d55c}".repeat(n);
        assert!(validate(&settings("custom", Some(&"x".repeat(128)), None)).is_ok());
        assert!(validate(&settings("custom", Some(&"x".repeat(129)), None)).is_err());
        assert!(validate(&settings("custom", Some(&korean(128)), None)).is_ok());
        assert!(validate(&settings("custom", Some(&korean(129)), None)).is_err());
        // Audit N7: the head stops at 9999.
        assert!(validate(&settings("headver", None, Some(9999))).is_ok());
        assert!(validate(&settings("headver", None, Some(10000))).is_err());
    }

    /// Audit N7: the value is read from a JSON object only.
    #[test]
    fn settings_are_read_from_an_object_only() {
        let read = |v: Value| serde_json::from_value::<VersioningSettings>(v);
        assert_eq!(
            read(json!({"scheme": "headver", "head": 3})).unwrap(),
            settings("headver", None, Some(3))
        );
        assert_eq!(
            read(json!({"format": "YYYY.0M", "scheme": "calver"})).unwrap(),
            settings("calver", Some("YYYY.0M"), None)
        );
        for bad in [
            json!(["semver"]),
            json!(["calver", "YYYY.0M.MICRO"]),
            json!(["headver", null, 3]),
            json!("semver"),
            json!(null),
            json!({}),
            json!({"format": "YYYY.0M"}),
            json!({"scheme": "semver", "format": null}),
            json!({"scheme": "headver", "head": null}),
            json!({"scheme": "headver", "head": -1}),
            json!({"scheme": "headver", "head": 1.5}),
            json!({"scheme": "headver", "head": "3"}),
            json!({"scheme": 1}),
            json!({"scheme": "semver", "extra": true}),
        ] {
            assert!(read(bad.clone()).is_err(), "{bad}");
        }
        // What is written is what is read.
        let written = serde_json::to_value(settings("semver", None, None)).unwrap();
        assert_eq!(written, json!({"scheme": "semver"}));
    }

    #[test]
    fn the_managed_line_names_the_scheme() {
        assert_eq!(
            render_line(&settings("semver", None, None), "en"),
            "Versioning: SemVer (MAJOR.MINOR.PATCH).\n"
        );
        assert_eq!(
            render_line(&settings("headver", None, Some(3)), "en"),
            "Versioning: HeadVer ({head}.{yearweek}.{build}), head 3.\n"
        );
        assert_eq!(
            render_line(&settings("headver", None, None), "en"),
            "Versioning: HeadVer ({head}.{yearweek}.{build}).\n"
        );
        assert_eq!(
            render_line(&settings("calver", Some("YYYY.0M.MICRO"), None), "en"),
            "Versioning: CalVer (YYYY.0M.MICRO).\n"
        );
        assert_eq!(
            render_line(&settings("calver", None, None), "en"),
            "Versioning: CalVer.\n"
        );
        assert_eq!(
            render_line(&settings("custom", Some("train plus build"), None), "en"),
            "Versioning: custom (train plus build).\n"
        );
        assert_eq!(
            render_line(&settings("semver", None, None), "ko"),
            "버전 규칙: SemVer (MAJOR.MINOR.PATCH).\n"
        );
    }
}
