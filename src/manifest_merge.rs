//! The three-way merge of the skill source manifest, the work of
//! `moai manifest-merge-driver %O %A %B %P` (SPEC-20261126-manifest-merge-driver-rust).
//!
//! Git calls a merge driver before it writes any merged file to the work tree (and `merge-tree`
//! has no work tree), so nothing is read from the package on disk. Every manifest value is chosen
//! three-way (equal sides, or the side that changed from the common version; `origin` one key at
//! a time). The sha256 of a file both sides changed comes from git's histogram file merge of that
//! file's three versions, found by their sha256 in the commit history.
//!
//! Whatever cannot be decided, any error and any panic leave git's own text merge of the three
//! manifests in the `%A` file and refuse. The command trusts each side's sha256 values, and `-X`
//! options of `git merge` never reach a driver; `tests/source_manifest_facts.rs` fails on any
//! manifest of this repository that does not match its files.

use serde::de::{DeserializeSeed, Deserializer, Error as _, MapAccess, SeqAccess, Visitor};
use serde_json::{Map, Number, Value};
use sha2::{Digest, Sha256};
use std::collections::{BTreeSet, HashMap};
use std::ffi::OsStr;
use std::io::Write as _;
use std::panic::{self, AssertUnwindSafe};
use std::path::Path;
use std::process::{Command, Output, Stdio};

type Result<T> = std::result::Result<T, String>;

/// How the path of a package manifest ends; what stands before it is the package folder.
const MANIFEST_SUFFIX: &str = "/references/source-manifest.json";
/// Pseudo-refs that may name a side of the merge in progress, besides every ref (`--all`).
const HEADS: [&str; 6] = [
    "HEAD",
    "ORIG_HEAD",
    "MERGE_HEAD",
    "CHERRY_PICK_HEAD",
    "REVERT_HEAD",
    "REBASE_HEAD",
];

/// The exit status of a refused merge. The shared command entry gives every operation error this
/// status; git reads 1 to 127 as a conflict and anything above as a failed merge.
pub const REFUSED_STATUS: u8 = 1;

/// A merge the command did not decide.
#[derive(Debug)]
pub struct Refused {
    /// The exit status the command ends with ([`REFUSED_STATUS`]).
    pub status: u8,
    /// The reason and the repair, for standard error.
    pub message: String,
}

/// The driver: writes the merged manifest into `ours` (`%A`), or leaves git's text merge of the
/// three manifests there and refuses. `path` (`%P`) names the manifest in the repository.
pub fn drive(
    base: &Path,
    ours: &Path,
    theirs: &Path,
    path: &OsStr,
) -> std::result::Result<(), Refused> {
    guard(
        &path.to_string_lossy(),
        || merge_into(base, ours, theirs, path),
        || text_merge(base, ours, theirs),
    )
}

/// Runs `work`; when it fails or panics, runs `fallback` and refuses with the reason.
// @cc:careful: a panic that escaped would end the process with status 101 and leave `%A` with our
// side alone, which git shows as a conflict that silently lacks their side. Every path out of
// `work` other than success has to run `fallback`.
fn guard(
    path: &str,
    work: impl FnOnce() -> Result<()>,
    fallback: impl FnOnce() -> Result<()>,
) -> std::result::Result<(), Refused> {
    let reason = match panic::catch_unwind(AssertUnwindSafe(work)) {
        Ok(Ok(())) => return Ok(()),
        Ok(Err(reason)) => reason,
        Err(payload) => {
            let said = payload
                .downcast_ref::<&str>()
                .map(|s| s.to_string())
                .or_else(|| payload.downcast_ref::<String>().cloned())
                .unwrap_or_else(|| "no message".to_string());
            format!("the command failed unexpectedly: {said}")
        }
    };
    let held = match fallback() {
        Ok(()) => "The file holds git's text merge.".to_string(),
        Err(e) => format!("Git's text merge failed ({e}), so the file may hold our side only."),
    };
    Err(Refused {
        status: REFUSED_STATUS,
        message: format!(
            "manifest-merge-driver: {path} left as a conflict: {reason}. {held} Resolve the skill \
             files and the conflict in the manifest, commit, then run \
             `moai check fix <base>..<head>`."
        ),
    })
}

/// Reads the three manifests, merges them and replaces `ours` with the result.
fn merge_into(base: &Path, ours: &Path, theirs: &Path, path: &OsStr) -> Result<()> {
    let read = |file: &Path, label: &str| {
        std::fs::read_to_string(file).map_err(|e| format!("{label} cannot be read: {e}"))
    };
    let texts = [
        read(base, "the common version")?,
        read(ours, "our side")?,
        read(theirs, "their side")?,
    ];
    let path = path.to_str().ok_or("the path is not UTF-8")?;
    let merged = merged_manifest(&texts, path)?;
    replace(ours, &merged)
}

/// Writes `text` into a temporary file beside `target` and renames it over `target`.
// @cc:careful: `target` is the `%A` file, the input of the fallback. It is replaced in one rename,
// so a failed write never leaves a cut-off file for git's text merge to read as our side.
fn replace(target: &Path, text: &str) -> Result<()> {
    let folder = target
        .parent()
        .filter(|p| !p.as_os_str().is_empty())
        .unwrap_or_else(|| Path::new("."));
    let failed = |e: std::io::Error| format!("the merged manifest cannot be written: {e}");
    let mut file = tempfile::NamedTempFile::new_in(folder).map_err(failed)?;
    file.write_all(text.as_bytes()).map_err(failed)?;
    file.persist(target).map_err(|e| failed(e.error))?;
    Ok(())
}

/// Git's own text merge of the three manifests, into the `ours` file.
fn text_merge(base: &Path, ours: &Path, theirs: &Path) -> Result<()> {
    let out = run(git()
        .args(["merge-file", "-L", "ours", "-L", "base", "-L", "theirs"])
        .args([ours, base, theirs]))?;
    // The status is the number of conflicts (at most 127); an error is negative, 255 here.
    match out.status.code() {
        Some(0..=127) => Ok(()),
        _ => Err(String::from_utf8_lossy(&out.stderr).trim().to_string()),
    }
}

// --- git -----------------------------------------------------------------------------------------

/// A git command with the environment git gave the driver.
// @cc:careful: the environment is inherited on purpose, so `identity::scrub_git_env` is not used:
// `GIT_DIR` names the repository being merged, `GITHEAD_*` its sides, and
// `GIT_CONFIG_PARAMETERS` carries `git -c diff.algorithm=...`, which the settings refusal has to
// see. `tests/git_env_scrub.rs` looks for the literal program name only and does not see this
// spawn; the `MOAI_GIT_BIN` form is the one the git hook of `card/guard.rs` uses.
fn git() -> Command {
    // `MOAI_GIT_BIN` names the git program for a runner whose child carries no `PATH`.
    let program = std::env::var_os("MOAI_GIT_BIN").unwrap_or_else(|| "git".into());
    let mut command = Command::new(program);
    command.stdin(Stdio::null());
    command
}

fn run(command: &mut Command) -> Result<Output> {
    command.output().map_err(|e| format!("git cannot run: {e}"))
}

/// The standard output of a git call that has to succeed; `what` names it in the error.
fn succeeded(what: &str, out: Output) -> Result<Vec<u8>> {
    if out.status.success() {
        return Ok(out.stdout);
    }
    Err(format!(
        "git {what} failed: {}",
        String::from_utf8_lossy(&out.stderr).trim()
    ))
}

/// The trimmed standard output of `git config <args>`; an unset key gives an empty answer.
fn config(args: &[&str]) -> Result<String> {
    let out = run(git().arg("config").args(args))?;
    Ok(String::from_utf8_lossy(&out.stdout).trim().to_string())
}

/// Refuses the settings that make git's own file merge differ from
/// `git merge-file --diff-algorithm=histogram`.
fn refuse_settings_git_merges_differently() -> Result<()> {
    let algorithm = config(&["--get", "diff.algorithm"])?.to_lowercase();
    if !matches!(algorithm.as_str(), "" | "histogram") {
        return Err(format!(
            "diff.algorithm is '{algorithm}'; the driver repeats only histogram"
        ));
    }
    if config(&["--type=bool", "--get", "merge.renormalize"])? == "true" {
        return Err(
            "merge.renormalize is on; the driver cannot repeat renormalized merges".to_string(),
        );
    }
    Ok(())
}

/// `--all` and the commits the pseudo-refs and `GITHEAD_<id>` variables name.
fn candidate_revs() -> Result<Vec<String>> {
    let mut revs = vec!["--all".to_string()];
    let from_environment = std::env::vars_os().filter_map(|(key, _)| {
        key.to_str()
            .and_then(|k| k.strip_prefix("GITHEAD_"))
            .map(str::to_string)
    });
    for name in HEADS
        .map(str::to_string)
        .into_iter()
        .chain(from_environment)
    {
        let pseudo_ref =
            !name.is_empty() && name.bytes().all(|b| b.is_ascii_uppercase() || b == b'_');
        let object_id = matches!(name.len(), 40 | 64)
            && name.bytes().all(|b| matches!(b, b'0'..=b'9' | b'a'..=b'f'));
        if !pseudo_ref && !object_id {
            continue;
        }
        let out = run(git().args(["rev-parse", "--verify", "-q", &format!("{name}^{{commit}}")]))?;
        if out.status.success() {
            revs.push(String::from_utf8_lossy(&out.stdout).trim().to_string());
        }
    }
    Ok(revs)
}

/// `{sha256: blob id}` of every past version of `paths` reachable from any ref or head.
fn blobs_by_sha256(paths: &[String]) -> Result<HashMap<String, String>> {
    let mut log = git();
    log.arg("log").args(candidate_revs()?).args([
        "--diff-merges=separate",
        "--format=",
        "--raw",
        "--no-abbrev",
        "--no-renames",
        "--",
    ]);
    log.args(paths.iter().map(|p| format!(":(top,literal){p}")));
    let raw = succeeded("log", run(&mut log)?)?;
    let mut ids = BTreeSet::new();
    for line in String::from_utf8_lossy(&raw).lines() {
        if !line.starts_with(':') {
            continue;
        }
        // `:<mode> <mode> <old id> <new id> <status>\t<path>`; an id of zeros names no blob.
        let meta = line.split('\t').next().unwrap_or_default();
        let fields: Vec<&str> = meta.split_whitespace().collect();
        if fields.len() >= 4 {
            ids.extend(
                fields[2..4]
                    .iter()
                    .filter(|id| !id.trim_matches('0').is_empty())
                    .map(|id| id.to_string()),
            );
        }
    }
    if ids.is_empty() {
        return Ok(HashMap::new());
    }
    let asked: String = ids.iter().map(|id| format!("{id}\n")).collect();
    let data = fed("cat-file", git().args(["cat-file", "--batch"]), asked)?;
    let mut found = HashMap::new();
    let mut at = 0;
    for _ in 0..ids.len() {
        let cut_off = || "git cat-file answered less than it was asked".to_string();
        let rest = data.get(at..).ok_or_else(cut_off)?;
        let end = rest.iter().position(|b| *b == b'\n').ok_or_else(cut_off)?;
        let header = String::from_utf8_lossy(&rest[..end]).into_owned();
        let unread = || format!("git cat-file answered `{header}`");
        // `<id> <type> <size>`; a missing object answers `<id> missing`.
        let words: Vec<&str> = header.split_whitespace().collect();
        let [id, kind, size] = words[..] else {
            return Err(unread());
        };
        let size: usize = size.parse().map_err(|_| unread())?;
        let body = rest
            .get(end + 1..)
            .and_then(|b| b.get(..size))
            .ok_or_else(cut_off)?;
        if kind == "blob" {
            found.insert(sha256_hex(body), id.to_string());
        }
        // The body is followed by one newline.
        at += end + 1 + size + 1;
    }
    Ok(found)
}

/// The standard output of a git call that has to succeed, with `input` on its standard input.
fn fed(what: &str, command: &mut Command, input: String) -> Result<Vec<u8>> {
    let mut child = command
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .map_err(|e| format!("git cannot run: {e}"))?;
    let mut stdin = child
        .stdin
        .take()
        .ok_or_else(|| format!("git {what} takes no input"))?;
    // Written from a thread of its own: git answers while it is still being asked.
    let writer = std::thread::spawn(move || stdin.write_all(input.as_bytes()));
    let out = child
        .wait_with_output()
        .map_err(|e| format!("git {what} failed: {e}"))?;
    writer
        .join()
        .map_err(|_| format!("writing to git {what} failed"))?
        .map_err(|e| format!("writing to git {what} failed: {e}"))?;
    succeeded(what, out)
}

/// Refuses a skill file among `changed` that has a `merge` attribute of its own: git merges such
/// a file its own way, which the manifests do not show.
fn refuse_own_merge_attribute(package: &str, changed: &[&str]) -> Result<()> {
    if changed.is_empty() {
        return Ok(());
    }
    let top = succeeded(
        "rev-parse",
        run(git().args(["rev-parse", "--show-toplevel"]))?,
    )?;
    let top = String::from_utf8(top).map_err(|_| "the repository path is not UTF-8")?;
    // On standard input: a merge that renames a whole folder changes more paths than one command
    // line holds everywhere.
    let asked: String = changed
        .iter()
        .map(|rel| format!("{package}/{rel}\n"))
        .collect();
    let attributes = fed(
        "check-attr",
        git().args(["-C", top.trim(), "check-attr", "--stdin", "merge"]),
        asked,
    )?;
    if let Some(line) = String::from_utf8_lossy(&attributes)
        .lines()
        .find(|line| !line.ends_with(": merge: unspecified"))
    {
        return Err(format!("a skill file has its own merge attribute: {line}"));
    }
    Ok(())
}

/// A file both sides changed: its three versions by the sha256 each manifest names.
struct Todo {
    rel: String,
    ours: Option<Value>,
    base: Value,
    theirs: Option<Value>,
}

/// The sha256 of git's histogram merge of each file of `todo`, in the same order.
fn merged_hashes(package: &str, todo: &[Todo]) -> Result<Vec<String>> {
    let paths: Vec<String> = todo
        .iter()
        .map(|t| format!("{package}/{}", t.rel))
        .collect();
    let blobs = blobs_by_sha256(&paths)?;
    let mut hashes = Vec::new();
    for file in todo {
        let find = |sha: Option<&Value>| sha.and_then(Value::as_str).and_then(|s| blobs.get(s));
        let (Some(ours), Some(base), Some(theirs)) = (
            find(file.ours.as_ref()),
            find(Some(&file.base)),
            find(file.theirs.as_ref()),
        ) else {
            return Err(format!(
                "a version of `{}` named by the manifests is not in the history",
                file.rel
            ));
        };
        let out = run(git()
            .args([
                "merge-file",
                "-p",
                "--object-id",
                "--diff-algorithm=histogram",
            ])
            .args([ours, base, theirs]))?;
        if !out.status.success() {
            return Err(format!("`{}` does not merge cleanly", file.rel));
        }
        hashes.push(sha256_hex(&out.stdout));
    }
    Ok(hashes)
}

fn sha256_hex(bytes: &[u8]) -> String {
    Sha256::digest(bytes)
        .iter()
        .map(|b| format!("{b:02x}"))
        .collect()
}

// --- the manifest --------------------------------------------------------------------------------

/// A JSON value read so that a key written twice in any object is an error, never "the last one
/// wins" (a text merge of two entry orders can leave such a manifest).
#[derive(Clone, Copy)]
struct NoKeyTwice;

impl<'de> DeserializeSeed<'de> for NoKeyTwice {
    type Value = Value;

    fn deserialize<D: Deserializer<'de>>(self, d: D) -> std::result::Result<Value, D::Error> {
        d.deserialize_any(self)
    }
}

impl<'de> Visitor<'de> for NoKeyTwice {
    type Value = Value;

    fn expecting(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str("a JSON value")
    }

    fn visit_bool<E>(self, v: bool) -> std::result::Result<Value, E> {
        Ok(Value::Bool(v))
    }

    fn visit_i64<E>(self, v: i64) -> std::result::Result<Value, E> {
        Ok(Value::from(v))
    }

    fn visit_u64<E>(self, v: u64) -> std::result::Result<Value, E> {
        Ok(Value::from(v))
    }

    fn visit_f64<E>(self, v: f64) -> std::result::Result<Value, E> {
        Ok(Number::from_f64(v).map_or(Value::Null, Value::Number))
    }

    fn visit_str<E>(self, v: &str) -> std::result::Result<Value, E> {
        Ok(Value::String(v.to_string()))
    }

    fn visit_string<E>(self, v: String) -> std::result::Result<Value, E> {
        Ok(Value::String(v))
    }

    fn visit_unit<E>(self) -> std::result::Result<Value, E> {
        Ok(Value::Null)
    }

    fn visit_seq<A: SeqAccess<'de>>(self, mut seq: A) -> std::result::Result<Value, A::Error> {
        let mut out = Vec::new();
        while let Some(value) = seq.next_element_seed(self)? {
            out.push(value);
        }
        Ok(Value::Array(out))
    }

    fn visit_map<A: MapAccess<'de>>(self, mut map: A) -> std::result::Result<Value, A::Error> {
        let mut out = Map::new();
        while let Some(key) = map.next_key::<String>()? {
            if out.contains_key(&key) {
                return Err(A::Error::custom(format!(
                    "has the key `{key}` twice; keep one entry"
                )));
            }
            let value = map.next_value_seed(self)?;
            out.insert(key, value);
        }
        Ok(Value::Object(out))
    }
}

/// One side of the merge: a JSON object with a `files` object, no key twice.
fn parse(text: &str, label: &str) -> Result<Map<String, Value>> {
    let mut reader = serde_json::Deserializer::from_str(text);
    let value = NoKeyTwice
        .deserialize(&mut reader)
        .and_then(|value| reader.end().map(|()| value))
        .map_err(|e| {
            if e.is_data() {
                format!("{label} of the manifest {e}")
            } else {
                format!("{label} of the manifest is not JSON: {e}")
            }
        })?;
    match value {
        Value::Object(map) if map.get("files").is_some_and(Value::is_object) => Ok(map),
        _ => Err(format!("{label} of the manifest has no `files` object")),
    }
}

/// The keys of `first`, then the keys only `second` has.
fn joined_keys<'a>(
    first: &'a Map<String, Value>,
    second: &'a Map<String, Value>,
) -> impl Iterator<Item = &'a String> {
    first
        .keys()
        .chain(second.keys().filter(|k| !first.contains_key(*k)))
}

/// What the common version holds under `key`: the outer `None` when it is not known (no common
/// object, or `null` there), the inner `None` when the key is absent.
fn common<'a>(base: Option<&'a Map<String, Value>>, key: &str) -> Option<Option<&'a Value>> {
    match base?.get(key) {
        Some(Value::Null) => None,
        found => Some(found),
    }
}

/// Three-way choice of one value (`None` is an absent key): equal sides, or the side that changed
/// from the common version. Without a known common version only equal sides are accepted.
fn pick<'a>(
    what: &str,
    base: Option<Option<&'a Value>>,
    ours: Option<&'a Value>,
    theirs: Option<&'a Value>,
) -> Result<Option<&'a Value>> {
    if ours == theirs {
        return Ok(ours);
    }
    if let Some(base) = base {
        if ours == base {
            return Ok(theirs);
        }
        if theirs == base {
            return Ok(ours);
        }
    }
    Err(format!("{what} differs between the sides"))
}

/// One `files` entry both sides hold: every key but `sha256` chosen three-way, an object value
/// (`origin`) one key at a time. The `sha256` is decided by [`merge_files`].
fn merge_entry(
    rel: &str,
    base: Option<&Value>,
    ours: &Value,
    theirs: &Value,
) -> Result<Map<String, Value>> {
    let not_an_object = || format!("`{rel}`: an entry is not an object");
    let ours = ours.as_object().ok_or_else(not_an_object)?;
    let theirs = theirs.as_object().ok_or_else(not_an_object)?;
    // An entry the common version holds as anything but an object, `null` included, is refused.
    let base = match base {
        None => None,
        Some(value) => Some(value.as_object().ok_or_else(not_an_object)?),
    };
    let mut out = Map::new();
    for key in joined_keys(ours, theirs) {
        let (o, t) = (ours.get(key), theirs.get(key));
        if key == "sha256" {
            if let Some(value) = o.or(t) {
                out.insert(key.clone(), value.clone());
            }
            continue;
        }
        let b = common(base, key);
        let value = match (o, t, b) {
            // Both sides hold an object, and the common version holds one or is not known: one
            // sub-key at a time, so one side's `kind` and the other's `reason` join.
            (
                Some(Value::Object(o)),
                Some(Value::Object(t)),
                b @ (None | Some(Some(Value::Object(_)))),
            ) => {
                let b = b.flatten().and_then(Value::as_object);
                let mut value = Map::new();
                for sub in joined_keys(o, t) {
                    let what = format!("`{rel}`: {key}.{sub}");
                    if let Some(chosen) = pick(&what, common(b, sub), o.get(sub), t.get(sub))? {
                        value.insert(sub.clone(), chosen.clone());
                    }
                }
                Some(Value::Object(value))
            }
            _ => pick(&format!("`{rel}`: {key}"), b, o, t)?.cloned(),
        };
        if let Some(value) = value {
            out.insert(key.clone(), value);
        }
    }
    Ok(out)
}

/// The `files` map merged three-way, and the files both sides changed, whose `sha256` still has to
/// come from [`merged_hashes`].
fn merge_files(
    base: &Map<String, Value>,
    ours: &Map<String, Value>,
    theirs: &Map<String, Value>,
) -> Result<(Map<String, Value>, Vec<Todo>)> {
    let mut files = Map::new();
    let mut todo = Vec::new();
    for rel in joined_keys(ours, theirs) {
        let b = base.get(rel);
        let (o, t) = match (ours.get(rel), theirs.get(rel)) {
            (Some(o), Some(t)) => (o, t),
            (Some(side), None) | (None, Some(side)) => {
                match b {
                    // Added on one side.
                    None => {
                        files.insert(rel.clone(), side.clone());
                    }
                    Some(b) if b != side => {
                        return Err(format!(
                            "`{rel}` was deleted or renamed on one side and changed on the other"
                        ));
                    }
                    // Deleted on one side, unchanged on the other.
                    Some(_) => {}
                }
                continue;
            }
            (None, None) => continue,
        };
        let mut entry = merge_entry(rel, b, o, t)?;
        let sha = |entry: &Value| entry.get("sha256").filter(|v| !v.is_null()).cloned();
        let (bs, os, ts) = (b.and_then(sha), sha(o), sha(t));
        let chosen = if os == ts {
            os
        } else if bs.is_some() && os == bs {
            ts
        } else if bs.is_some() && ts == bs {
            os
        } else if let Some(bs) = bs {
            todo.push(Todo {
                rel: rel.clone(),
                ours: os,
                base: bs,
                theirs: ts,
            });
            None
        } else {
            return Err(format!(
                "`{rel}` was added on both sides with different contents"
            ));
        };
        entry.insert("sha256".to_string(), chosen.unwrap_or(Value::Null));
        files.insert(rel.clone(), Value::Object(entry));
    }
    refuse_folder_moves(base, ours, theirs)?;
    Ok((files, todo))
}

/// Every file whose entry either side changed from the common version: added, deleted or
/// different.
fn changed_files<'a>(
    base: &'a Map<String, Value>,
    ours: &'a Map<String, Value>,
    theirs: &'a Map<String, Value>,
) -> Vec<&'a str> {
    let every: BTreeSet<&str> = base
        .keys()
        .chain(ours.keys())
        .chain(theirs.keys())
        .map(String::as_str)
        .collect();
    every
        .into_iter()
        .filter(|rel| ours.get(*rel) != base.get(*rel) || theirs.get(*rel) != base.get(*rel))
        .collect()
}

/// Every folder above a file of `files`.
fn folders(files: &Map<String, Value>) -> BTreeSet<&str> {
    files
        .keys()
        .flat_map(|rel| rel.match_indices('/').map(move |(at, _)| &rel[..at]))
        .collect()
}

/// A folder that disappears on one side while the other side adds or changes a file in it: git's
/// directory-rename detection may move that file, which the manifests cannot show.
fn refuse_folder_moves(
    base: &Map<String, Value>,
    ours: &Map<String, Value>,
    theirs: &Map<String, Value>,
) -> Result<()> {
    let before = folders(base);
    for (side, other) in [(ours, theirs), (theirs, ours)] {
        let kept = folders(side);
        for (rel, entry) in other {
            if base.get(rel) == Some(entry) {
                continue;
            }
            let moved = before.difference(&kept).find(|folder| {
                rel.strip_prefix(**folder)
                    .is_some_and(|rest| rest.starts_with('/'))
            });
            if let Some(folder) = moved {
                return Err(format!(
                    "folder `{folder}` disappears on one side while the other side adds or \
                     changes `{rel}` in it"
                ));
            }
        }
    }
    Ok(())
}

/// The `files` object of a parsed side.
fn files_of(side: &Map<String, Value>) -> Result<&Map<String, Value>> {
    side.get("files")
        .and_then(Value::as_object)
        .ok_or_else(|| "a side of the manifest has no `files` object".to_string())
}

/// The merged manifest text, from the texts of the common version, our side and their side.
fn merged_manifest(texts: &[String; 3], path: &str) -> Result<String> {
    let package = path
        .strip_suffix(MANIFEST_SUFFIX)
        .ok_or_else(|| format!("{path} is not a package manifest"))?;
    refuse_settings_git_merges_differently()?;
    let ours = parse(&texts[1], "our side")?;
    let theirs = parse(&texts[2], "their side")?;
    let base = parse(&texts[0], "the common version")?;

    // The top-level keys in our order, `files` in its place; their own keys follow.
    let mut data = Map::new();
    for key in joined_keys(&ours, &theirs) {
        if key == "files" {
            data.insert(key.clone(), Value::Null);
            continue;
        }
        let what = format!("top-level `{key}`");
        let value = pick(
            &what,
            common(Some(&base), key),
            ours.get(key),
            theirs.get(key),
        )?;
        if let Some(value) = value {
            data.insert(key.clone(), value.clone());
        }
    }

    let (bf, of, tf) = (files_of(&base)?, files_of(&ours)?, files_of(&theirs)?);
    let (mut files, todo) = merge_files(bf, of, tf)?;
    refuse_own_merge_attribute(package, &changed_files(bf, of, tf))?;
    if !todo.is_empty() {
        for (file, sha) in todo.iter().zip(merged_hashes(package, &todo)?) {
            if let Some(entry) = files.get_mut(&file.rel).and_then(Value::as_object_mut) {
                entry.insert("sha256".to_string(), Value::String(sha));
            }
        }
    }
    data.insert("files".to_string(), Value::Object(files));
    render(data)
}

/// The manifest text: `files` in sorted key order, every other key in place, two-space indent,
/// non-ASCII kept, one newline.
// @cc:careful: these are the bytes `check_fix::render_manifest` (`moai check fix`) writes for the
// same tree. Change both together; `tests/manifest_merge_driver_cli.rs` compares them.
fn render(mut data: Map<String, Value>) -> Result<String> {
    if let Some(Value::Object(files)) = data.get_mut("files") {
        let mut entries: Vec<(String, Value)> = std::mem::take(files).into_iter().collect();
        entries.sort_by(|a, b| a.0.cmp(&b.0));
        files.extend(entries);
    }
    serde_json::to_string_pretty(&Value::Object(data))
        .map(|text| text + "\n")
        .map_err(|e| format!("the merged manifest cannot be rendered: {e}"))
}

#[cfg(test)]
mod tests {
    use super::*;

    const PATH: &str = "skills/moai/references/source-manifest.json";

    #[test]
    fn a_panic_runs_the_fallback_and_refuses_with_a_status_below_128() {
        let mut fallbacks = 0;
        let refused = guard(
            PATH,
            || panic!("injected failure"),
            || {
                fallbacks += 1;
                Ok(())
            },
        )
        .expect_err("a panic is a refusal");
        assert_eq!(fallbacks, 1);
        assert!((1..=127).contains(&refused.status), "{}", refused.status);
        for said in [PATH, "injected failure", "moai check fix"] {
            assert!(refused.message.contains(said), "{}", refused.message);
        }
    }

    #[test]
    fn a_success_runs_no_fallback() {
        let mut fallbacks = 0;
        let done = guard(
            PATH,
            || Ok(()),
            || {
                fallbacks += 1;
                Ok(())
            },
        );
        assert!(done.is_ok());
        assert_eq!(fallbacks, 0);
    }
}
