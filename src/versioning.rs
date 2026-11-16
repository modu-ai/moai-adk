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
//!   [`MAX_TAG_ENTRIES`] directory entries; folders are entered for names of at most
//!   [`MAX_TAG_PARTS`] parts, and deeper names are skipped) and the `<id> refs/tags/<name>`
//!   lines of `packed-refs` (a regular file of at most [`MAX_PACKED`] bytes; peeled `^`
//!   lines are skipped).
//! - `<project>/.git` is a file: it is followed only for a linked worktree, that is a
//!   `gitdir: <path>` line (relative to the project) naming an admin folder that holds a
//!   `commondir` file and a `gitdir` file (both relative to the admin folder) where `gitdir`
//!   names `<project>/.git` back and the admin folder is `<common>/worktrees/<name>` of the
//!   common folder `commondir` names. The tags are then read from that common folder, by the
//!   same rules. It is outside the project: the admin and common folders are checked
//!   themselves (a real folder, not a link), and links in the path leading to them are
//!   followed.
//! - Anything else — `.git` as a link, a submodule's pointer, a pointer to a repository that
//!   does not name this folder back, an admin folder outside the common folder's
//!   `worktrees`, `refs`, `refs/tags` or `packed-refs` as a link or a special file, a file
//!   or a folder over its limit — is `tags: unavailable` with a `tags_reason`, and no tag is
//!   read. No `.git` entry is `tags: no_repository`: a parent folder's repository is not the
//!   project's.
//!
//! "Highest" is version order, not date: the numeric segments compared left to right, then
//! the pre-release in SemVer 2.0.0 order (a release above its pre-releases). Only a name that
//! is short, plain text ([`showable`]) is ever considered or shown.
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
    "also valid SemVer; beside a SemVer version of the same major it is counted as SemVer";

/// A two-digit year is taken as one from 10 (2010): below that a first segment is far more
/// often a plain number.
const FIRST_SHORT_YEAR: u64 = 10;
const FIRST_FULL_YEAR: u64 = 2000;

/// The numeric segments of a version and its pre-release, the text between a `-` that ends
/// the segments and the next `+`. A leading `v` and a prefix that ends in `/` or `@`
/// (`release/v1.2.3`, `pkg@1.2.3`) are dropped; build metadata after `+` is not read.
/// `None` unless there are two to four all-digit segments of at most nine digits.
fn parts(text: &str) -> Option<(Vec<&str>, Option<&str>)> {
    let name = text.rsplit(['/', '@']).next().unwrap_or(text);
    let name = name.strip_prefix(['v', 'V']).unwrap_or(name);
    let end = name.find(['-', '+']).unwrap_or(name.len());
    let (core, tail) = name.split_at(end);
    let segments: Vec<&str> = core.split('.').collect();
    let numeric = |s: &&str| !s.is_empty() && s.len() <= 9 && s.bytes().all(|b| b.is_ascii_digit());
    let prerelease = tail
        .strip_prefix('-')
        .map(|p| p.split('+').next().unwrap_or(p));
    ((2..=4).contains(&segments.len()) && segments.iter().all(numeric))
        .then_some((segments, prerelease))
}

fn number(segment: &str) -> u64 {
    segment.parse().unwrap_or(u64::MAX)
}

/// The first number of a version; `u64::MAX` for text that is none.
fn first_number(text: &str) -> u64 {
    parts(text).map_or(u64::MAX, |(segments, _)| number(segments[0]))
}

/// Two pre-releases in SemVer 2.0.0 order (section 11.4): a release (`None`) above any
/// pre-release; else the dot-separated identifiers left to right, numeric ones compared as
/// numbers and below alphanumeric ones, alphanumeric ones in ASCII order, and with an equal
/// prefix the shorter set lower.
fn prerelease_order(a: Option<&str>, b: Option<&str>) -> Ordering {
    let (a, b) = match (a, b) {
        (None, None) => return Ordering::Equal,
        (None, Some(_)) => return Ordering::Greater,
        (Some(_), None) => return Ordering::Less,
        (Some(a), Some(b)) => (a, b),
    };
    let numeric = |s: &str| !s.is_empty() && s.bytes().all(|b| b.is_ascii_digit());
    let identifier = |x: &str, y: &str| match (numeric(x), numeric(y)) {
        (true, true) => {
            let (x, y) = (x.trim_start_matches('0'), y.trim_start_matches('0'));
            x.len().cmp(&y.len()).then_with(|| x.cmp(y))
        }
        (true, false) => Ordering::Less,
        (false, true) => Ordering::Greater,
        (false, false) => x.cmp(y),
    };
    a.split('.')
        .zip(b.split('.'))
        .map(|(x, y)| identifier(x, y))
        .find(|o| o.is_ne())
        .unwrap_or_else(|| a.split('.').count().cmp(&b.split('.').count()))
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

/// The scheme each version is counted as: its shape, except that a HeadVer shape beside a
/// SemVer version of the same major is SemVer (every HeadVer version is valid SemVer, and a
/// minor that walks through `1001` is no second scheme). `None` for a shape that is not
/// counted.
fn counted(versions: &[&str]) -> Vec<Option<&'static str>> {
    let majors: Vec<u64> = versions
        .iter()
        .filter(|v| classify(v) == Shape::Semver)
        .map(|v| first_number(v))
        .collect();
    versions
        .iter()
        .map(|v| match classify(v) {
            Shape::Headver if majors.contains(&first_number(v)) => Some(Shape::Semver.name()),
            shape @ (Shape::Semver | Shape::Calver | Shape::Headver) => Some(shape.name()),
            Shape::Ambiguous | Shape::Unrecognised => None,
        })
        .collect()
}

/// The scheme a set of versions shows: the one scheme, `mixed` for two, `None` when none is
/// counted. SemVer and HeadVer together are a move to HeadVer, `headver`, when every head is
/// above every SemVer major (so every HeadVer version sorts above every SemVer one), and
/// `mixed` otherwise.
fn scheme_of(versions: &[&str]) -> Option<&'static str> {
    let counted = counted(versions);
    let firsts = |scheme: &'static str| {
        versions
            .iter()
            .zip(&counted)
            .filter(move |(_, c)| **c == Some(scheme))
            .map(|(v, _)| first_number(v))
    };
    let moved = matches!(
        (firsts("headver").min(), firsts("semver").max()),
        (Some(head), Some(major)) if head > major
    );
    let mut found: Vec<&'static str> = Vec::new();
    for scheme in counted.into_iter().flatten() {
        let scheme = if moved && scheme == "semver" {
            "headver"
        } else {
            scheme
        };
        if !found.contains(&scheme) {
            found.push(scheme);
        }
    }
    match found.as_slice() {
        [] => None,
        [one] => Some(one),
        _ => Some("mixed"),
    }
}

/// Where a tag stands among the evidence: versions of a definite shape first, then ambiguous
/// ones, then names that are no version; inside the first two the higher version first.
fn rank(name: &str) -> (u8, Vec<u64>, Option<&str>) {
    let tier = match classify(name) {
        Shape::Semver | Shape::Calver | Shape::Headver => 0,
        Shape::Ambiguous => 1,
        Shape::Unrecognised => 2,
    };
    match parts(name) {
        Some((segments, prerelease)) if tier < 2 => (
            tier,
            segments.iter().map(|s| number(s)).collect(),
            prerelease,
        ),
        _ => (tier, Vec::new(), None),
    }
}

/// Highest first: by tier, then numeric segments, then the pre-release in SemVer order (a
/// release above its pre-releases), and the name itself last so the order never depends on
/// how the names arrived.
fn by_rank(a: &&String, b: &&String) -> Ordering {
    let (ta, sa, ra) = rank(a);
    let (tb, sb, rb) = rank(b);
    ta.cmp(&tb)
        .then_with(|| sb.cmp(&sa))
        .then_with(|| prerelease_order(rb, ra))
        .then_with(|| if ta < 2 { b.cmp(a) } else { a.cmp(b) })
}

/// The versioning fact from tag names (in any order) and manifest versions
/// (`(file, version)`). The tags decide when one of them has a definite shape; the manifests
/// decide only without that. `evidence` holds the [`MAX_TAGS`] highest tags, with `mixed`
/// tags also the highest tag of each counted scheme not among them, then the manifests.
/// `latest` is the highest tag that is a version, else the first manifest version that is
/// one.
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
    let names: Vec<&str> = ranked.iter().map(|t| t.as_str()).collect();
    let tag_scheme = scheme_of(&names);
    let mut shown: Vec<&str> = names.iter().take(MAX_TAGS).copied().collect();
    if tag_scheme == Some("mixed") {
        // The highest tag of a scheme is its first in version order.
        let counted = counted(&names);
        let mut seen: Vec<&str> = counted.iter().take(MAX_TAGS).flatten().copied().collect();
        for (name, scheme) in names.iter().zip(&counted).skip(MAX_TAGS) {
            if let Some(scheme) = scheme.filter(|s| !seen.contains(s)) {
                seen.push(scheme);
                shown.push(name);
            }
        }
    }
    let evidence: Vec<Value> = shown
        .iter()
        .map(|t| row("tag", t))
        .chain(manifests.iter().map(|(file, v)| row(file, v)))
        .collect();
    let manifest_values: Vec<&str> = manifests.iter().map(|(_, v)| v.as_str()).collect();
    let scheme = tag_scheme
        .or_else(|| scheme_of(&manifest_values))
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

/// The file at `path`, opened for reading without following a link and without waiting (a
/// pipe put there after the caller's check cannot hold the open), and only when what was
/// opened is a regular file ([`check_opened`]).
fn open_regular(path: &Path) -> Result<fs::File> {
    let mut options = fs::OpenOptions::new();
    options.read(true);
    #[cfg(unix)]
    {
        use std::os::unix::fs::OpenOptionsExt;
        options.custom_flags(libc::O_NOFOLLOW | libc::O_NONBLOCK);
    }
    let file = options.open(path).map_err(|_| "could not be read")?;
    check_opened(&file)?;
    Ok(file)
}

/// `Err` unless the opened handle is a regular file: the path may have been swapped for a
/// pipe or a device between the caller's check and the open.
fn check_opened(file: &fs::File) -> Result<()> {
    if file.metadata().is_ok_and(|m| m.is_file()) {
        Ok(())
    } else {
        Err("is not a regular file".into())
    }
}

/// The bytes of the regular file at `path`, at most `max`: `Ok(None)` when nothing is there,
/// `Err` with the reason for a link, a folder, a pipe or device, a larger file or a failed
/// read. The path is checked, then opened by [`open_regular`], which checks the handle again.
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
    let file = open_regular(path)?;
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
    // A relative pointer is relative to the folder that holds the file, as git reads it.
    let admin = root.join(target);
    if !real_dir(&admin) {
        return Err("the folder the .git file points at is missing or a link".into());
    }
    let common = pointer(&admin.join("commondir"))
        .filter(|p| !p.is_empty())
        .ok_or(not_a_worktree)?;
    // The repository must name this folder as its worktree, or the pointer could borrow
    // the tags of any repository. A relative back-pointer is relative to the admin folder.
    let back = pointer(&admin.join("gitdir"))
        .filter(|p| !p.is_empty())
        .map(|p| admin.join(p));
    let same_folder = |a: &Path, b: &Path| match (a.canonicalize(), b.canonicalize()) {
        (Ok(a), Ok(b)) => a == b,
        _ => false,
    };
    let names_back = back.as_deref().is_some_and(|back| {
        back.file_name().is_some_and(|n| n == ".git")
            && back.parent().is_some_and(|p| same_folder(p, root))
    });
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
    // The admin folder must be one of the common folder's own `worktrees/<name>` folders: an
    // admin folder made anywhere else could name this folder back and any repository as its
    // common folder.
    let owned = match (admin.canonicalize(), common.canonicalize()) {
        (Ok(admin), Ok(common)) => admin.parent() == Some(common.join("worktrees").as_path()),
        _ => false,
    };
    if !owned {
        return Err(
            "the folder the .git file points at is not a worktree folder of the repository it names"
                .into(),
        );
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
/// The unit of a CalVer token (calver.org, spec 26.0) and its place in a format: year, month
/// or week, day, `MINOR`, `MICRO`, `MODIFIER`. `MM`, `WW` and `DD` are the spelling the page
/// used before that spec. `MAJOR` has no place: calver.org makes it the first number, and the
/// profile's first token is the year.
#[derive(Clone, Copy, PartialEq, Eq, PartialOrd, Ord)]
enum CalverUnit {
    Year,
    Month,
    Week,
    Day,
    Minor,
    Micro,
    Modifier,
}

fn calver_unit(token: &str) -> Option<CalverUnit> {
    Some(match token {
        "YYYY" | "YY" | "0Y" => CalverUnit::Year,
        "M" | "0M" | "MM" => CalverUnit::Month,
        "W" | "0W" | "WW" => CalverUnit::Week,
        "D" | "0D" | "DD" => CalverUnit::Day,
        "MINOR" => CalverUnit::Minor,
        "MICRO" => CalverUnit::Micro,
        "MODIFIER" => CalverUnit::Modifier,
        _ => return None,
    })
}

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

/// Two to four known tokens joined by dots, at most one per unit, in the order of
/// [`CalverUnit`] and the year first; a day only behind a month (so never with a week), and
/// `MODIFIER` not right behind the year.
fn calver_format(format: &str) -> bool {
    let Some(units) = format
        .split('.')
        .map(calver_unit)
        .collect::<Option<Vec<CalverUnit>>>()
    else {
        return false;
    };
    let has = |unit| units.contains(&unit);
    (2..=4).contains(&units.len())
        && units[0] == CalverUnit::Year
        && !(has(CalverUnit::Month) && has(CalverUnit::Week))
        && (!has(CalverUnit::Day) || has(CalverUnit::Month))
        && units.windows(2).all(|w| w[0] < w[1])
        && !units
            .windows(2)
            .any(|w| w == [CalverUnit::Year, CalverUnit::Modifier])
}

/// A character that shows nothing or turns the text's direction where the managed section
/// shows it: the soft hyphen, the Arabic letter mark, the Mongolian vowel separator, the
/// zero-width and direction marks, the line and paragraph separators, the embeddings and
/// overrides, the word joiner and invisible operators and isolates, the byte order mark,
/// variation selectors and tag characters.
fn hidden(c: char) -> bool {
    matches!(
        u32::from(c),
        0x00AD
            | 0x061C
            | 0x180E
            | 0x200B..=0x200F
            | 0x2028..=0x202E
            | 0x2060..=0x206F
            | 0xFE00..=0xFE0F
            | 0xFEFF
            | 0xE0000..=0xE007F
            | 0xE0100..=0xE01EF
    )
}

/// One line of the user's own words, at most [`MAX_CUSTOM_CHARS`] characters, that cannot be
/// read as markup or hide text where the managed section shows it: no control character, no
/// `<`, `>`, backtick, `[`, `]`, `*` or `_`, no `#` at the start, no [`hidden`] character.
fn custom_text(text: &str) -> bool {
    !text.trim().is_empty()
        && text.chars().count() <= MAX_CUSTOM_CHARS
        && !text.trim_start().starts_with('#')
        && !text.chars().any(|c| {
            c.is_control() || matches!(c, '<' | '>' | '`' | '[' | ']' | '*' | '_') || hidden(c)
        })
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

    /// Re-audit B1: a HeadVer tag counts as SemVer only beside a SemVer tag of the same major.
    /// HeadVer heads above every SemVer major are a move to HeadVer; anything else is mixed.
    #[test]
    fn a_move_from_semver_to_headver_reads_headver() {
        let read = |list: &[&str]| {
            let v = summarize(&tags(list), &[]);
            (
                v["scheme"].as_str().unwrap().to_string(),
                v["latest"].clone(),
            )
        };
        // This repository's own plan: SemVer v3.x, then HeadVer with head 4. The kept head is
        // the first number of `latest`, 4, not the old major plus one.
        assert_eq!(
            read(&["v3.0.0", "v3.2.0", "4.2647.1"]),
            ("headver".into(), json!("4.2647.1"))
        );
        assert_eq!(
            read(&["v3.0.0", "v3.2.0", "4.2647.1", "4.2647.3"]),
            ("headver".into(), json!("4.2647.3"))
        );
        // A minor that walks through 1001 stays SemVer.
        assert_eq!(
            read(&["v1.999.0", "v1.1000.0", "v1.1001.0"]),
            ("semver".into(), json!("v1.1001.0"))
        );
        // The head at its edges beside major 3: one below is mixed, equal is SemVer, one
        // above is a move.
        assert_eq!(read(&["v3.2.0", "2.2647.1"]).0, "mixed");
        assert_eq!(read(&["v3.2.0", "3.2647.1"]).0, "semver");
        assert_eq!(read(&["v3.2.0", "4.2647.1"]).0, "headver");
        // A head between two SemVer majors, or heads on both sides, is no move.
        assert_eq!(read(&["v3.0.0", "v5.0.0", "4.2647.1"]).0, "mixed");
        assert_eq!(read(&["v3.0.0", "2.2647.1", "4.2647.1"]).0, "mixed");
        // A same-major HeadVer shape counts as SemVer, so the higher head is still a move.
        assert_eq!(read(&["v3.0.0", "3.2647.1", "4.2647.1"]).0, "headver");
        // CalVer beside a move is a second scheme.
        assert_eq!(read(&["v3.0.0", "4.2647.1", "2026.11.0"]).0, "mixed");
        // The manifests follow the same rule.
        let manifests = [
            ("Cargo.toml", "3.2.0".to_string()),
            ("package.json", "4.2647.1".to_string()),
        ];
        assert_eq!(summarize(&[], &manifests)["scheme"], "headver");
    }

    /// Re-audit N6: pre-releases compare per SemVer 2.0.0 section 11.4: dot-separated
    /// identifiers, numeric ones as numbers and below alphanumeric ones, a shorter set lower
    /// when the rest is equal.
    #[test]
    fn pre_releases_compare_by_their_identifiers() {
        let v = summarize(&tags(&["v2.0.0-rc.2", "v2.0.0-rc.10", "v1.0.0"]), &[]);
        assert_eq!(v["latest"], "v2.0.0-rc.10");
        assert_eq!(values(&v), ["v2.0.0-rc.10", "v2.0.0-rc.2", "v1.0.0"]);
        // The example of section 11.4, highest first, from a scrambled order.
        let highest_first = [
            "1.0.0",
            "1.0.0-rc.1",
            "1.0.0-beta.11",
            "1.0.0-beta.2",
            "1.0.0-beta",
            "1.0.0-alpha.beta",
            "1.0.0-alpha.1",
            "1.0.0-alpha",
        ];
        let names = tags(&[
            "1.0.0-beta.2",
            "1.0.0-alpha",
            "1.0.0",
            "1.0.0-alpha.beta",
            "1.0.0-beta.11",
            "1.0.0-rc.1",
            "1.0.0-alpha.1",
            "1.0.0-beta",
        ]);
        let mut sorted: Vec<&String> = names.iter().collect();
        sorted.sort_by(by_rank);
        assert_eq!(sorted, highest_first);
        // A numeric identifier with more digits is the higher number.
        let names = tags(&["1.0.0-9", "1.0.0-10", "1.0.0-x"]);
        let mut sorted: Vec<&String> = names.iter().collect();
        sorted.sort_by(by_rank);
        assert_eq!(sorted, ["1.0.0-x", "1.0.0-10", "1.0.0-9"]);
    }

    /// Re-audit N7: with `mixed`, the highest tag of each counted scheme is evidence even
    /// when it is not among the five highest.
    #[test]
    fn mixed_evidence_shows_the_highest_of_each_scheme() {
        let calver = [
            "2026.01.0",
            "2026.02.0",
            "2026.03.0",
            "2026.04.0",
            "2026.05.0",
            "2026.06.0",
        ];
        let five = [
            "2026.06.0",
            "2026.05.0",
            "2026.04.0",
            "2026.03.0",
            "2026.02.0",
        ];
        let with = |more: &[&str]| {
            let mut list = calver.to_vec();
            list.extend(more);
            summarize(&tags(&list), &[("Cargo.toml", "3.1.0".to_string())])
        };
        let v = with(&["v3.0.0", "v3.1.0"]);
        assert_eq!(v["scheme"], "mixed");
        let mut expected = five.to_vec();
        expected.extend(["v3.1.0", "3.1.0"]);
        assert_eq!(values(&v), expected);
        // Three schemes: the highest SemVer and the highest HeadVer, in version order.
        let v = with(&["v1.0.0", "0.2647.1", "0.2647.0"]);
        assert_eq!(v["scheme"], "mixed");
        let mut expected = five.to_vec();
        expected.extend(["v1.0.0", "0.2647.1", "3.1.0"]);
        assert_eq!(values(&v), expected);
        // A scheme already among the five adds nothing.
        let v = summarize(&tags(&["v3.0.0", "2026.01.0"]), &[]);
        assert_eq!(values(&v), ["2026.01.0", "v3.0.0"]);
        // One scheme: the five highest only.
        let v = with(&[]);
        assert_eq!(v["scheme"], "calver");
        let mut expected = five.to_vec();
        expected.push("3.1.0");
        assert_eq!(values(&v), expected);
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

    /// Re-audit N3: custom text refuses markup characters and every character of the hidden
    /// ranges, at both ends of each range; the visible neighbours of a range are taken.
    #[test]
    fn custom_text_refuses_markup_and_every_hidden_range() {
        let custom = |text: String| validate(&settings("custom", Some(&text), None));
        for c in ['[', ']', '*', '_', '<', '>', '`'] {
            assert!(custom(format!("a {c} b")).is_err(), "{c:?}");
        }
        let ranges: [(u32, u32); 10] = [
            (0x00AD, 0x00AD),
            (0x061C, 0x061C),
            (0x180E, 0x180E),
            (0x200B, 0x200F),
            (0x202A, 0x202E),
            (0x2060, 0x206F),
            (0xFEFF, 0xFEFF),
            (0xFE00, 0xFE0F),
            (0xE0000, 0xE007F),
            (0xE0100, 0xE01EF),
        ];
        for (first, last) in ranges {
            for code in [first, last] {
                let c = char::from_u32(code).unwrap();
                assert!(custom(format!("a{c}b")).is_err(), "U+{code:04X}");
            }
        }
        // The neighbours outside each range, where the neighbour is not itself refused for
        // another reason (U+180D and U+180F are variation selectors outside the list, U+2029
        // is a paragraph separator, U+FEFE is unassigned).
        for code in [
            0x00AC, 0x00AE, 0x061B, 0x061D, 0x200A, 0x2010, 0x202F, 0x205F, 0x2070, 0xFDFF, 0xFE10,
            0xE0080, 0xE00FF, 0xE01F0,
        ] {
            let c = char::from_u32(code).unwrap();
            assert!(custom(format!("a{c}b")).is_ok(), "U+{code:04X}");
        }
        // Plain text with other punctuation is taken.
        assert!(custom("release/train 4 (build #2), weekly; 50% + x".into()).is_ok());
    }

    /// Re-audit N4: a CalVer format has at most one token per unit, in the order year, month
    /// or week, day, MINOR, MICRO, MODIFIER; a week excludes a month and a day, a day follows
    /// a month, MODIFIER comes last and not right behind the year, and MAJOR is refused
    /// (calver.org: MAJOR is the first number, and the profile's first token is the year).
    #[test]
    fn calver_formats_follow_the_unit_order() {
        let calver = |f: &str| validate(&settings("calver", Some(f), None)).is_ok();
        for ok in [
            "YYYY.0M.MICRO",
            "YY.0M.MICRO",
            "YYYY.0M.0D",
            "YY.0M",
            "YYYY.0M.0D.MICRO",
            "YYYY.MINOR.MICRO",
            "YY.MINOR.MICRO",
            "YYYY.MM.DD",
            "YYYY.0W",
            "YYYY.0W.MICRO",
            "YYYY.0M.MODIFIER",
            "YYYY.0M.MICRO.MODIFIER",
            "0Y.M.D",
            "YYYY.MICRO",
        ] {
            assert!(calver(ok), "{ok}");
        }
        for bad in [
            // Two tokens of one unit.
            "YYYY.0M.MM",
            "YYYY.M.0M",
            "YYYY.0W.WW",
            "YYYY.0D.DD.MICRO",
            // Out of order.
            "YYYY.0D.0M",
            "YYYY.MICRO.0M",
            "YYYY.MICRO.MINOR",
            "YYYY.0M.MODIFIER.MICRO",
            // A week with a month or a day; a day without a month.
            "YYYY.0W.0D",
            "YYYY.0W.0M",
            "YYYY.0M.0W",
            "YYYY.0D",
            "YYYY.0D.MICRO",
            // MODIFIER right behind the year; MAJOR anywhere.
            "YYYY.MODIFIER",
            "YYYY.MAJOR",
            "YYYY.MAJOR.MINOR",
            "YYYY.0M.MAJOR",
        ] {
            assert!(!calver(bad), "{bad}");
        }
    }

    /// A named pipe nobody writes to, in a fresh temporary folder.
    #[cfg(unix)]
    fn named_pipe() -> (tempfile::TempDir, std::path::PathBuf) {
        let dir = tempfile::TempDir::new().unwrap();
        let pipe = dir.path().join("pipe");
        let made = std::process::Command::new("mkfifo")
            .arg(&pipe)
            .status()
            .unwrap();
        assert!(made.success());
        (dir, pipe)
    }

    /// Re-audit N5 (mutant M1): opening a named pipe does not wait for a writer, and the
    /// opened pipe is refused. Without `O_NONBLOCK` the open blocks and this times out.
    #[cfg(unix)]
    #[test]
    fn opening_a_named_pipe_returns_an_error_within_two_seconds() {
        let (_dir, pipe) = named_pipe();
        let (send, receive) = std::sync::mpsc::channel();
        std::thread::spawn(move || {
            let _ = send.send(open_regular(&pipe).map(|_| ()));
        });
        let opened = receive
            .recv_timeout(std::time::Duration::from_secs(2))
            .expect("opening a named pipe waited for a writer");
        assert_eq!(opened, Err("is not a regular file".to_string()));
    }

    /// Re-audit N5 (mutant M3): a path swapped for a pipe after the caller's metadata check
    /// is refused on the handle: the test above calls `open_regular` without that check, as
    /// if the swap came between the two, and here `check_opened` refuses an opened pipe and
    /// takes a regular file. (The pipe is not opened through `open_regular` here, so a
    /// missing `O_NONBLOCK` fails the test above instead of hanging this one.)
    #[cfg(unix)]
    #[test]
    fn the_opened_handle_is_checked_again() {
        use std::os::unix::fs::OpenOptionsExt;
        let (dir, pipe) = named_pipe();
        let handle = fs::OpenOptions::new()
            .read(true)
            .custom_flags(libc::O_NONBLOCK)
            .open(&pipe)
            .unwrap();
        assert_eq!(
            check_opened(&handle),
            Err("is not a regular file".to_string())
        );
        let regular = dir.path().join("regular");
        fs::write(&regular, b"x").unwrap();
        assert_eq!(check_opened(&fs::File::open(&regular).unwrap()), Ok(()));
        assert!(open_regular(&regular).is_ok());
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
