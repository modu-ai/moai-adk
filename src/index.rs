//! Incremental index (C6 #2182): immutable generations under
//! `.moai-cli/index/generations/<generation_id>/manifest.json`, published by an atomic
//! pointer swap of `.moai-cli/index/current.json` (temp write + rename, so a reader sees
//! the old or the new pointer, never a partial one — AC-2: mixed generations 0).
//!
//! Design basis (design snapshot cache lifetime/concurrent writes, gap-analysis row 10):
//! immutable shards keyed by content digest; a partial update builds a new generation
//! and publishes it through an atomic pointer; an interrupted build stays behind as an
//! orphan candidate without a completion mark; cache loss is a performance loss only —
//! it never destroys source or receipt data. The index never writes outside
//! `.moai-cli/index/` and never modifies source files.
//!
//! REQ-C6-01: when a build fails mid-way, the previous generation is kept (the pointer
//! is never moved to a partial build) and the failure is recorded so queries report the
//! `partial` state until the next successful build clears it.
//!
//! Query freshness — the four states of design §10, never guessed:
//! - `unavailable`: no pointer, an unreadable pointer, or a missing manifest.
//! - `partial`: the last build attempt failed (the marker is present).
//! - `stale`: a complete generation exists but the source moved (branch switch, or
//!   added/deleted/modified files vs the manifest).
//! - `fresh`: the manifest matches the current scan exactly.
use crate::baseline::utc_now;
use cap_fs_ext::{DirExt, FollowSymlinks, OpenOptionsFollowExt};
#[cfg(unix)]
use cap_std::fs::OpenOptionsExt;
use cap_std::fs::{Dir, OpenOptions};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    collections::BTreeMap,
    ffi::OsStr,
    fs,
    io::{self, Read, Write},
    path::{Component, Path, PathBuf},
    process::{Command, Stdio},
};

type Result<T> = std::result::Result<T, String>;

pub const SCHEMA_VERSION: u32 = 1;
/// Scan guard: a tracked-file table beyond this is refused, not silently truncated.
pub const MAX_FILES: usize = 16384;
const GIT_OUTPUT_LIMIT: usize = 4 * 1024 * 1024;
/// A manifest file-count cap on read (a hand-edited generation must not blow memory).
const MAX_MANIFEST_FILES: usize = MAX_FILES;
// Covers the maximum Git path table even after JSON's six-byte control escaping.
const STATE_BYTE_LIMIT: u64 = 32 * 1024 * 1024;

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
enum ReadPoint {
    BeforeParentOpen,
    BeforeFileOpen,
}

/// One operation's authority. Paths below this directory are never reopened ambiently.
struct IndexIo {
    root: Dir,
    /// The project's state folder, read once when the handle opens (#2725 S2b).
    state: &'static str,
}

impl IndexIo {
    fn new(project: &Path) -> io::Result<Self> {
        Ok(Self {
            root: Dir::open_ambient_dir(project, cap_std::ambient_authority())?,
            state: crate::state_root::dir(project),
        })
    }

    fn parent<'a>(
        &self,
        relative: &'a Path,
        create: bool,
        hook: &mut impl FnMut(ReadPoint, &Path),
    ) -> io::Result<(Option<Dir>, &'a OsStr)> {
        let mut components = relative.components().peekable();
        let mut owned = None;
        while let Some(component) = components.next() {
            let Component::Normal(name) = component else {
                return Err(io::Error::new(
                    io::ErrorKind::InvalidInput,
                    "non-relative index path",
                ));
            };
            if components.peek().is_none() {
                return Ok((owned, name));
            }
            let current = owned.as_ref().unwrap_or(&self.root);
            hook(ReadPoint::BeforeParentOpen, relative);
            let mut opened = current.open_dir_nofollow(name);
            if create
                && opened
                    .as_ref()
                    .is_err_and(|e| e.kind() == io::ErrorKind::NotFound)
            {
                match current.create_dir(name) {
                    Ok(()) => (),
                    Err(e) if e.kind() == io::ErrorKind::AlreadyExists => (),
                    Err(e) => return Err(e),
                }
                opened = current.open_dir_nofollow(name);
            }
            let next = opened.map_err(|error| {
                // This is error classification, never authorization for a subsequent open.
                if current
                    .symlink_metadata(name)
                    .is_ok_and(|m| !m.is_dir() || m.file_type().is_symlink())
                {
                    io::Error::new(io::ErrorKind::NotFound, "source parent replaced")
                } else {
                    error
                }
            })?;
            owned = Some(next);
        }
        Err(io::Error::new(
            io::ErrorKind::InvalidInput,
            "empty index path",
        ))
    }

    fn regular_file(
        parent: &Dir,
        name: &OsStr,
        write: bool,
    ) -> io::Result<(fs::File, fs::Metadata)> {
        let mut options = OpenOptions::new();
        options
            .read(!write)
            .write(write)
            .create(write)
            .follow(FollowSymlinks::No);
        #[cfg(unix)]
        options.custom_flags(libc::O_NONBLOCK);
        let file = parent.open_with(name, &options)?.into_std();
        let metadata = file.metadata()?;
        if !metadata.is_file() {
            return Err(io::Error::new(
                io::ErrorKind::InvalidData,
                "index state/source is not a regular file",
            ));
        }
        Ok((file, metadata))
    }

    fn read_state(&self, relative: &Path) -> io::Result<Option<Vec<u8>>> {
        let (owned, name) = match self.parent(relative, false, &mut |_, _| {}) {
            Ok(parent) => parent,
            Err(e) if e.kind() == io::ErrorKind::NotFound => return Ok(None),
            Err(e) => return Err(e),
        };
        let parent = owned.as_ref().unwrap_or(&self.root);
        let (file, metadata) = match Self::regular_file(parent, name, false) {
            Ok(file) => file,
            Err(e) if e.kind() == io::ErrorKind::NotFound => return Ok(None),
            Err(e) => return Err(e),
        };
        if metadata.len() > STATE_BYTE_LIMIT {
            return Err(io::Error::new(
                io::ErrorKind::InvalidData,
                "index state byte limit exceeded",
            ));
        }
        let mut bytes = Vec::with_capacity(metadata.len() as usize + 1);
        file.take(STATE_BYTE_LIMIT + 1).read_to_end(&mut bytes)?;
        if bytes.len() as u64 > STATE_BYTE_LIMIT {
            return Err(io::Error::new(
                io::ErrorKind::InvalidData,
                "index state byte limit exceeded",
            ));
        }
        Ok(Some(bytes))
    }

    /// Reads a tracked source file through the same handle walk as the scan (no symlink
    /// is followed), bounded by `limit`.
    fn read_source(&self, relative: &Path, limit: u64) -> io::Result<Vec<u8>> {
        let (owned, name) = self.parent(relative, false, &mut |_, _| {})?;
        let parent = owned.as_ref().unwrap_or(&self.root);
        let (file, _) = Self::regular_file(parent, name, false)?;
        let mut bytes = Vec::new();
        file.take(limit + 1).read_to_end(&mut bytes)?;
        if bytes.len() as u64 > limit {
            return Err(io::Error::new(io::ErrorKind::InvalidData, "source grew"));
        }
        Ok(bytes)
    }

    fn write_marker(&self, bytes: &[u8]) -> io::Result<()> {
        if bytes.len() as u64 > STATE_BYTE_LIMIT {
            return Err(io::Error::new(
                io::ErrorKind::InvalidData,
                "index state byte limit exceeded",
            ));
        }
        let marker = format!("{}/index/last-build.json", self.state);
        let (owned, name) = self.parent(Path::new(&marker), true, &mut |_, _| {})?;
        let parent = owned.as_ref().unwrap_or(&self.root);
        let (mut file, _) = Self::regular_file(parent, name, true)?;
        // Never truncate until the opened handle itself is known to be regular.
        file.set_len(0)?;
        file.write_all(bytes)
    }
}

#[derive(Debug, Serialize, Deserialize)]
pub struct IndexBuildRequest {
    pub protocol_version: u32,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct IndexQueryRequest {
    pub protocol_version: u32,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
pub struct IndexedFile {
    pub path: String,
    pub sha256: String,
    pub bytes: u64,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct IndexManifest {
    pub schema_version: u32,
    pub generation_id: String,
    pub base_commit: String,
    pub files: Vec<IndexedFile>,
    pub built_at: String,
    /// The Map/Graph views stored beside this manifest (C6b #2347); absent in a
    /// generation built before them.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub records: Option<RecordsRef>,
}

/// Digests and counts of `map.jsonl` / `graph.jsonl` — a reader rejects a view whose
/// bytes do not match, so Map and Graph always come from the same generation.
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
pub struct RecordsRef {
    pub version: String,
    pub file_set: String,
    pub observed_at: String,
    pub map_sha256: String,
    pub graph_sha256: String,
    pub map_records: usize,
    pub graph_records: usize,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct BuildFailure {
    /// Always `"partial"` — the marker exists only for a failed attempt.
    pub status: String,
    pub reason: String,
    pub at: String,
}

fn git(dir: &Path) -> Command {
    let mut cmd = Command::new("git");
    cmd.arg("--literal-pathspecs");
    cmd.arg("--no-optional-locks").arg("-C").arg(dir);
    crate::identity::scrub_git_env(&mut cmd);
    cmd.env("GIT_OPTIONAL_LOCKS", "0").stdin(Stdio::null());
    cmd
}

fn git_out(dir: &Path, args: &[&str]) -> Result<Vec<u8>> {
    let out = git(dir)
        .args(args)
        .stderr(Stdio::null())
        .output()
        .map_err(|_| "index source unavailable: git could not start")?;
    if !out.status.success() {
        return Err(format!("index source unavailable: git {} failed", args[0]));
    }
    if out.stdout.len() > GIT_OUTPUT_LIMIT {
        return Err(format!(
            "index source unavailable: git {} output limit exceeded ({GIT_OUTPUT_LIMIT} bytes)",
            args[0]
        ));
    }
    Ok(out.stdout)
}

fn git_line(dir: &Path, args: &[&str]) -> Option<String> {
    let s = String::from_utf8(git_out(dir, args).ok()?).ok()?;
    Some(s.trim_end_matches('\n').to_string()).filter(|s| !s.is_empty() && !s.contains('\n'))
}

fn file_sha(f: &mut fs::File) -> Option<(String, u64)> {
    let mut h = Sha256::new();
    let mut buf = [0u8; 65536];
    let mut total = 0u64;
    loop {
        let n = f.read(&mut buf).ok()?;
        if n == 0 {
            break;
        }
        h.update(&buf[..n]);
        total += n as u64;
        if total > 16 * 1024 * 1024 {
            return None;
        }
    }
    Some((format!("sha256:{:x}", h.finalize()), total))
}

fn index_root(project: &Path) -> PathBuf {
    crate::state_root::path(project, "index")
}

fn generations_dir(project: &Path) -> PathBuf {
    index_root(project).join("generations")
}

fn failure_marker_path(project: &Path) -> PathBuf {
    index_root(project).join("last-build.json")
}

/// Tracked files with content digests (sorted by path), and the `HEAD` commit the scan
/// is bound to. Untracked files are deliberately not part of the index content — they
/// are reported by query change detection instead. The index never indexes itself:
/// `.moai-cli/index/**` is excluded even when a repository tracks it, or every build
/// would flip the pointer it just wrote into a "modification". Symlinks are digested by
/// their target text (the same rule the 07-D snapshot uses).
fn scan(project: &Path, source: &IndexIo) -> Result<(String, Vec<IndexedFile>)> {
    scan_with_hook(project, source, |_, _| {})
}

fn scan_with_hook(
    project: &Path,
    source: &IndexIo,
    mut hook: impl FnMut(ReadPoint, &Path),
) -> Result<(String, Vec<IndexedFile>)> {
    let raw = git_out(project, &["ls-files", "-z"])?;
    let index_dir = format!("{}/index", source.state);
    let index_prefix = format!("{index_dir}/");
    let mut paths: Vec<String> = raw
        .split(|b| *b == 0)
        .filter(|e| !e.is_empty())
        .map(|e| String::from_utf8_lossy(e).into_owned())
        .filter(|p| !p.starts_with(&index_prefix) && *p != index_dir)
        .collect();
    paths.sort();
    paths.dedup();
    if paths.len() > MAX_FILES {
        return Err(format!(
            "index file limit exceeded ({MAX_FILES}); the scan is refused, not truncated"
        ));
    }
    let base_commit = git_line(project, &["rev-parse", "--verify", "HEAD^{commit}"])
        .ok_or("index source unavailable: no HEAD commit")?;
    let mut files = Vec::new();
    for p in &paths {
        let relative = Path::new(p);
        let (owned, name) = match source.parent(relative, false, &mut hook) {
            Ok(parent) => parent,
            Err(e) if e.kind() == io::ErrorKind::NotFound => continue,
            Err(_) => return Err(format!("index source unavailable: {p} parent unavailable")),
        };
        let parent = owned.as_ref().unwrap_or(&source.root);
        let meta = match parent.symlink_metadata(name) {
            Ok(meta) => meta,
            Err(e) if e.kind() == io::ErrorKind::NotFound => continue,
            Err(_) => return Err(format!("index source unavailable: {p} metadata unreadable")),
        };
        let (sha, bytes) = if meta.file_type().is_symlink() {
            let t: String = parent
                .read_link_contents(name)
                .ok()
                .and_then(|t| t.into_os_string().into_string().ok())
                .ok_or_else(|| format!("index source unavailable: {p} symlink unreadable"))?;
            let d = Sha256::digest(t.as_bytes());
            (format!("symlink-sha256:{d:x}"), t.len() as u64)
        } else if meta.is_file() {
            hook(ReadPoint::BeforeFileOpen, relative);
            let (mut file, _) = IndexIo::regular_file(parent, name, false)
                .map_err(|_| format!("index source unavailable: {p} read failed"))?;
            file_sha(&mut file)
                .ok_or_else(|| format!("index source unavailable: {p} read failed"))?
        } else if meta.is_dir() {
            // Only an actual gitlink is a submodule, not a file replaced by a directory.
            let entry = git_out(project, &["ls-files", "--stage", "-z", "--", p])?;
            if !entry.starts_with(b"160000 ") {
                continue;
            }
            let head = git_line(
                &project.join(p),
                &["rev-parse", "--verify", "HEAD^{commit}"],
            )
            .ok_or_else(|| format!("index source unavailable: {p} submodule unreadable"))?;
            (format!("commit:{head}"), 0)
        } else {
            return Err(format!(
                "index source unavailable: {p} unsupported file type"
            ));
        };
        files.push(IndexedFile {
            path: p.clone(),
            sha256: sha,
            bytes,
        });
    }
    Ok((base_commit, files))
}

/// Untracked files (`??` in `git status --porcelain=v1 -z --untracked-files=all
/// --no-renames`), capped, sorted — the "new file" detection input.
fn untracked(project: &Path) -> Result<Vec<String>> {
    let raw = git_out(
        project,
        &[
            "status",
            "--porcelain=v1",
            "-z",
            "--untracked-files=all",
            "--no-renames",
        ],
    )?;
    let mut out = Vec::new();
    let own = format!("{}/", crate::state_root::dir(project));
    for e in raw.split(|b| *b == 0) {
        if e.len() > 3 && e.starts_with(b"?? ") {
            let path = String::from_utf8_lossy(&e[3..]).into_owned();
            // The tool's own state directory (this index, receipts, claims) is never a
            // project change: the index writes there itself, so counting it would make
            // `fresh` unreachable.
            if path.starts_with(&own) {
                continue;
            }
            out.push(path);
        }
    }
    out.sort();
    if out.len() > MAX_FILES {
        return Err(format!(
            "index source unavailable: untracked file limit exceeded ({MAX_FILES})"
        ));
    }
    Ok(out)
}

fn canonical_manifest_bytes(m: &IndexManifest) -> Result<Vec<u8>> {
    let mut v = serde_json::to_vec(m).map_err(|_| "index manifest encode failed".to_string())?;
    v.push(b'\n');
    Ok(v)
}

/// Generation id: `gen-` + sha256 prefix over the indexed CONTENT only (base commit plus
/// the sorted file table) — same content, same id, so an identical rebuild dedupes into
/// the existing immutable directory. `built_at` and the carried id are excluded: a
/// timestamp is not content.
///
/// `history` is the digest of the Graph's history part (`build_history`): the same HEAD
/// and files can reach a different commit history (a shallow clone later unshallowed),
/// and that changes the Graph, so it must change the id — an immutable generation never
/// stands for two different Graphs.
fn content_id(base_commit: &str, files: &[IndexedFile], history: &str) -> String {
    let mut sorted: Vec<&IndexedFile> = files.iter().collect();
    sorted.sort_by(|a, b| a.path.cmp(&b.path));
    let mut h = Sha256::new();
    // v2: a generation also carries the Map/Graph views, so the record and extractor
    // versions are content — a generation built by older extractors is never reused.
    h.update(b"moai-index-generation-v2\0");
    let versions = crate::index_records::versions();
    for part in [versions.as_str(), history] {
        h.update((part.len() as u64).to_be_bytes());
        h.update(part.as_bytes());
    }
    h.update((base_commit.len() as u64).to_be_bytes());
    h.update(base_commit.as_bytes());
    for f in sorted {
        h.update((f.path.len() as u64).to_be_bytes());
        h.update(f.path.as_bytes());
        h.update((f.sha256.len() as u64).to_be_bytes());
        h.update(f.sha256.as_bytes());
        h.update(f.bytes.to_be_bytes());
    }
    let d = h.finalize();
    format!("gen-{d:x}")[..20].to_string()
}

fn orphan_candidates(project: &Path) -> Vec<String> {
    let mut out = Vec::new();
    if let Ok(entries) = fs::read_dir(generations_dir(project)) {
        for e in entries.flatten() {
            if let Some(name) = e.file_name().to_str() {
                if name.starts_with(".tmp-") {
                    out.push(name.to_string());
                }
            }
        }
    }
    out.sort();
    out
}

/// Writes the pointer through a temp file plus a rename in the same directory: a reader
/// either sees the previous pointer or the new one, never a partial write (AC-2).
fn swap_pointer(project: &Path, generation_id: &str) -> Result<()> {
    let pointer = crate::apply::target(
        project,
        &crate::state_root::rel(project, "index/current.json"),
    )?;
    let dir = index_root(project);
    fs::create_dir_all(&dir).map_err(|_| "index directory unavailable".to_string())?;
    let nanos = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_nanos())
        .unwrap_or(0);
    let tmp = dir.join(format!(".current.json.tmp-{}-{nanos}", std::process::id()));
    let body = serde_json::to_vec_pretty(
        &json!({"schema_version": SCHEMA_VERSION, "generation_id": generation_id}),
    )
    .map_err(|_| "index pointer encode failed".to_string())?;
    let mut staged = fs::OpenOptions::new()
        .write(true)
        .create_new(true)
        .open(&tmp)
        .map_err(|_| "index pointer write failed".to_string())?;
    staged
        .write_all(&body)
        .map_err(|_| "index pointer write failed".to_string())?;
    drop(staged);
    fs::rename(&tmp, &pointer).map_err(|_| "index pointer replace failed".to_string())
}

/// Same content and the same views: the only condition under which an existing
/// generation directory may stand for this build.
fn same_snapshot(stored: &IndexManifest, expected: &IndexManifest) -> bool {
    stored.base_commit == expected.base_commit
        && stored.files == expected.files
        && stored.records == expected.records
}

/// A simultaneous publisher may win the rename. Reuse only a validated identical
/// snapshot; keep this attempt's staging directory as an orphan, never delete data.
fn publish_staged(
    project: &Path,
    source: &IndexIo,
    staging: &Path,
    generation_id: &str,
    expected: &IndexManifest,
) -> Result<&'static str> {
    match fs::rename(staging, generations_dir(project).join(generation_id)) {
        Ok(()) => Ok("created"),
        Err(e) => {
            if load_manifest(source, generation_id).is_some_and(|stored| {
                same_snapshot(&stored, expected) && stored_views(source, &stored).is_ok()
            }) {
                Ok("unchanged")
            } else {
                Err(format!("index generation staging write failed: {e}"))
            }
        }
    }
}

/// The commit time of `base` in UTC — the records' `observed_at`. Time is content: the
/// same snapshot always yields the same records (INCREMENTAL EQUIVALENCE).
/// Read from the commit object itself (`committer <ident> <secs> <tz>`), so a damaged
/// history beyond the base commit cannot block the scan.
fn commit_time(dir: &Path, base: &str) -> Result<String> {
    let unreadable = || "index source unavailable: base commit time unreadable".to_string();
    let raw = git_out(dir, &["cat-file", "commit", base]).map_err(|_| unreadable())?;
    let text = String::from_utf8_lossy(&raw);
    let header = text.split("\n\n").next().unwrap_or("");
    let secs: i64 = header
        .lines()
        .find_map(|l| l.strip_prefix("committer "))
        .and_then(|l| l.rsplit(' ').nth(1))
        .and_then(|s| s.parse().ok())
        .ok_or_else(unreadable)?;
    Ok(crate::baseline::utc_from_epoch(secs))
}

/// Per-origin record lines of the published generation and its file digests, for
/// reuse. Anything unreadable or mismatched means no reuse — a cache loss is a
/// performance loss only.
type Reusable = (
    BTreeMap<String, String>,
    BTreeMap<Option<String>, Vec<Value>>,
    BTreeMap<Option<String>, Vec<Value>>,
);

fn reusable_records(source: &IndexIo) -> Option<Reusable> {
    let gen_id = read_pointer(source).ok()??;
    let manifest = load_manifest(source, &gen_id)?;
    if manifest.records.as_ref()?.version != crate::index_records::versions() {
        return None;
    }
    let (map, graph) = stored_views(source, &manifest).ok()?;
    let files = manifest
        .files
        .iter()
        .map(|f| (f.path.clone(), f.sha256.clone()))
        .collect();
    Some((
        files,
        crate::index_records::by_origin(map),
        crate::index_records::by_origin(graph),
    ))
}

/// Reads both views of a generation and checks them against its manifest: digests,
/// header generation id, base commit and file set. Returns the record lines.
fn stored_views(source: &IndexIo, manifest: &IndexManifest) -> Result<(Vec<Value>, Vec<Value>)> {
    let records = manifest
        .records
        .as_ref()
        .ok_or("generation has no map/graph records")?;
    let mut out = Vec::new();
    for (view, sha) in [
        ("map", &records.map_sha256),
        ("graph", &records.graph_sha256),
    ] {
        let bytes = source
            .read_state(Path::new(&format!(
                "{}/index/generations/{}/{view}.jsonl",
                source.state, manifest.generation_id
            )))
            .ok()
            .flatten()
            .ok_or("map/graph records unreadable")?;
        if &crate::digest(&bytes) != sha {
            return Err("map/graph records do not match the generation".into());
        }
        let (header, lines) = crate::index_records::parse_view(&bytes)?;
        let expected = crate::index_records::header(
            view,
            &manifest.generation_id,
            &manifest.base_commit,
            &records.file_set,
        );
        if header != expected["header"]
            || records.file_set != crate::index_records::file_set_digest(&manifest.files)
        {
            return Err("map/graph records do not match the generation".into());
        }
        out.push(lines);
    }
    let graph = out.pop().unwrap_or_default();
    let map = out.pop().unwrap_or_default();
    Ok((map, graph))
}

/// Builds both views for a scan: reuses the stored fragments of unchanged files, reads
/// every other file through the scan's handle walk and refuses the build when its bytes
/// no longer match the manifest digest (Map and Graph never describe a file set the
/// manifest does not). Returns the views, their reference and the reuse counts.
fn build_views(
    root: &Path,
    source: &IndexIo,
    gen_id: &str,
    base_commit: &str,
    files: &[IndexedFile],
    history: crate::spec_links::Builder,
) -> Result<(crate::index_records::Views, RecordsRef, usize, usize)> {
    use crate::index_records::{self as rec, Content};
    let observed_at = commit_time(root, base_commit)?;
    let file_set = rec::file_set_digest(files);
    let mut previous = reusable_records(source);
    let (mut reused, mut extracted) = (0usize, 0usize);
    let mut fragments = BTreeMap::new();
    for f in files {
        let key = Some(f.path.clone());
        let stored = previous.as_mut().and_then(|(old, map, graph)| {
            (old.get(&f.path) == Some(&f.sha256)).then(|| {
                let m = map.remove(&key).unwrap_or_default();
                let g = graph.remove(&key).unwrap_or_default();
                rec::fragment_from_lines(&m, &g).ok()
            })?
        });
        if let Some(frag) = stored {
            reused += 1;
            fragments.insert(f.path.clone(), frag);
            continue;
        }
        extracted += 1;
        let frag = if !f.sha256.starts_with("sha256:") {
            rec::extract_file(&f.path, &f.sha256, Content::NotRegular)
        } else if !rec::needs_bytes(&f.path, f.bytes) {
            rec::extract_file(&f.path, &f.sha256, Content::Oversize)
        } else {
            let bytes = source
                .read_source(Path::new(&f.path), f.bytes)
                .ok()
                .filter(|b| crate::digest(b) == f.sha256)
                .ok_or_else(|| format!("index source changed during build: {}", f.path))?;
            rec::extract_file(&f.path, &f.sha256, Content::Bytes(&bytes))
        };
        fragments.insert(f.path.clone(), frag);
    }
    let views = rec::assemble(
        fragments,
        history,
        (gen_id, base_commit, &file_set),
        &observed_at,
    )?;
    for bytes in [&views.map, &views.graph] {
        rec::within_limit(bytes.len() as u64, STATE_BYTE_LIMIT)?;
    }
    let records = RecordsRef {
        version: rec::versions(),
        file_set,
        observed_at,
        map_sha256: crate::digest(&views.map),
        graph_sha256: crate::digest(&views.graph),
        map_records: views.map_records,
        graph_records: views.graph_records,
    };
    Ok((views, records, reused, extracted))
}

/// The Graph's history part for `base` — commit → issue mentions reachable from the base
/// commit in this clone — and its digest. Coverage limits are recorded, never hidden: an
/// unreadable history (a git failure) is `history_unavailable` and a shallow clone is
/// `history_shallow`; neither fails a build that the file scan completed. The digest
/// enters the generation id (`content_id`).
fn build_history(root: &Path, base: &str) -> (crate::spec_links::Builder, String) {
    let mut history = crate::spec_links::Builder::default();
    if let Err(reason) = crate::spec_links::commit_links(&mut history, root, Some(base)) {
        history = crate::spec_links::Builder::default();
        history.miss("history_unavailable", json!({"reason": reason}));
    }
    if git_line(root, &["rev-parse", "--is-shallow-repository"]).as_deref() == Some("true") {
        history.miss(
            "history_shallow",
            json!({"reason": "commits beyond the shallow boundary are not read"}),
        );
    }
    let edges: Vec<_> = history.edges.iter().collect();
    let digest = serde_json::to_vec(&json!({
        "nodes": history.nodes,
        "edges": edges,
        "missing": history.missing,
    }))
    .map(|bytes| crate::digest(&bytes))
    .unwrap_or_default();
    (history, digest)
}

/// Builds a new generation from the current scan and publishes it with the atomic
/// pointer swap. On any failure the pointer stays where it was (REQ-C6-01) and the
/// failure is recorded in `last-build.json` so queries report `partial` until the next
/// successful build.
pub fn index_build(project: &Path, req: &IndexBuildRequest) -> Result<Value> {
    if req.protocol_version != SCHEMA_VERSION {
        return Err("unsupported protocol_version; supported: 1".into());
    }
    let root = fs::canonicalize(project).map_err(|_| "path unavailable".to_string())?;
    let source =
        IndexIo::new(&root).map_err(|_| "index directory handle unavailable".to_string())?;
    let build = || -> Result<Value> {
        for rel in [
            crate::state_root::rel(&root, "index/generations"),
            crate::state_root::rel(&root, "index/current.json"),
            crate::state_root::rel(&root, "index/last-build.json"),
        ] {
            crate::apply::target(&root, &rel)?;
        }
        // Three static progress lines (#2806); silent without a stderr sink.
        crate::progress::note("scanning tracked files (1/3)");
        let (base_commit, files) = scan(&root, &source)?;
        let total_bytes: u64 = files.iter().map(|f| f.bytes).sum();
        let (history, history_digest) = build_history(&root, &base_commit);
        let gen_id = content_id(&base_commit, &files, &history_digest);
        crate::progress::note("building the views (2/3)");
        let (views, records, reused, extracted) =
            build_views(&root, &source, &gen_id, &base_commit, &files, history)?;
        let manifest = IndexManifest {
            schema_version: SCHEMA_VERSION,
            generation_id: gen_id.clone(),
            base_commit: base_commit.clone(),
            files: files.clone(),
            built_at: utc_now(),
            records: Some(records.clone()),
        };
        let gens = generations_dir(&root);
        fs::create_dir_all(&gens).map_err(|_| "index directory unavailable".to_string())?;
        let gen_dir = gens.join(&gen_id);
        for name in ["manifest.json", "map.jsonl", "graph.jsonl"] {
            crate::apply::target(
                &root,
                &format!("{}/index/generations/{gen_id}/{name}", source.state),
            )?;
        }
        crate::progress::note("publishing the generation (3/3)");
        let action = if !gen_dir.is_dir() {
            let nanos = std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .map(|d| d.as_nanos())
                .unwrap_or(0);
            let tmp = gens.join(format!(".tmp-{}-{nanos}", std::process::id()));
            fs::create_dir(&tmp)
                .map_err(|_| "index generation staging write failed".to_string())?;
            let bytes = canonical_manifest_bytes(&manifest)?;
            // All three files land in the staging directory before the one rename, so a
            // published generation never holds a manifest without its views.
            for (name, body) in [
                ("map.jsonl", &views.map),
                ("graph.jsonl", &views.graph),
                ("manifest.json", &bytes),
            ] {
                let mut staged = fs::OpenOptions::new()
                    .write(true)
                    .create_new(true)
                    .open(tmp.join(name))
                    .map_err(|e| format!("index generation staging write failed: {e}"))?;
                staged
                    .write_all(body)
                    .map_err(|e| format!("index generation staging write failed: {e}"))?;
            }
            publish_staged(&root, &source, &tmp, &gen_id, &manifest)?
        } else {
            // Reuse is successful only when the immutable manifest still describes
            // this snapshot. Never publish a damaged cache or overwrite it in place.
            let stored = load_manifest(&source, &gen_id)
                .ok_or("index generation manifest invalid; refusing reuse")?;
            if !same_snapshot(&stored, &manifest) || stored_views(&source, &stored).is_err() {
                return Err("index generation manifest invalid; refusing reuse".into());
            }
            "unchanged"
        };
        swap_pointer(&root, &gen_id)?;
        let _ = fs::remove_file(failure_marker_path(&root));
        Ok(json!({
            "protocol_version": SCHEMA_VERSION,
            "schema_version": SCHEMA_VERSION,
            "generation_id": gen_id,
            "base_commit": base_commit,
            "files": files.len(),
            "bytes": total_bytes,
            "action": action,
            "records": {
                "map_records": records.map_records,
                "graph_records": records.graph_records,
                "reused_files": reused,
                "extracted_files": extracted,
            },
            "status": "complete",
            "swap": "atomic",
            "orphan_candidates": orphan_candidates(&root),
        }))
    };
    match build() {
        Ok(v) => Ok(v),
        Err(e) => {
            // REQ-C6-01: keep the previous generation (no swap happened on this path)
            // and mark the failed attempt so queries report `partial`.
            let marker = BuildFailure {
                status: "partial".into(),
                reason: e.clone(),
                at: utc_now(),
            };
            // The failure path must enforce the same target/ancestor rule as the
            // success path, or its marker could overwrite a symlinked source.
            if crate::apply::target(
                &root,
                &crate::state_root::rel(&root, "index/last-build.json"),
            )
            .is_ok()
            {
                if let Ok(bytes) = serde_json::to_vec_pretty(&marker) {
                    let _ = source.write_marker(&bytes);
                }
            }
            Err(format!("index build failed; previous generation kept: {e}"))
        }
    }
}

fn read_pointer(source: &IndexIo) -> Result<Option<String>> {
    let Some(bytes) = source
        .read_state(Path::new(&format!("{}/index/current.json", source.state)))
        .map_err(|_| "index pointer unreadable".to_string())?
    else {
        return Ok(None);
    };
    let v: Value =
        serde_json::from_slice(&bytes).map_err(|_| "index pointer unreadable".to_string())?;
    v["generation_id"]
        .as_str()
        .filter(|s| generation_id_valid(s))
        .map(|s| Some(s.to_string()))
        .ok_or_else(|| "index pointer unreadable".to_string())
}

/// A generation name is exactly one canonical component, never a caller-supplied path.
fn generation_id_valid(id: &str) -> bool {
    id.strip_prefix("gen-").is_some_and(|hex| {
        hex.len() == 16
            && hex
                .bytes()
                .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b))
    })
}

fn load_manifest(source: &IndexIo, gen_id: &str) -> Option<IndexManifest> {
    if !generation_id_valid(gen_id) {
        return None;
    }
    let bytes = source
        .read_state(Path::new(&format!(
            "{}/index/generations/{gen_id}/manifest.json",
            source.state
        )))
        .ok()??;
    let m: IndexManifest = serde_json::from_slice(&bytes).ok()?;
    (m.schema_version == SCHEMA_VERSION
        && m.generation_id == gen_id
        && m.files.len() <= MAX_MANIFEST_FILES)
        .then_some(m)
}

/// Change detection (AC-1): additions, deletions, modifications and untracked new files
/// against the manifest, plus the branch switch (the manifest's `HEAD` moved).
fn changes_vs(
    manifest: &IndexManifest,
    base_commit: &str,
    files: &[IndexedFile],
    new_files: &[String],
) -> Value {
    let old: std::collections::BTreeMap<&str, &IndexedFile> = manifest
        .files
        .iter()
        .map(|f| (f.path.as_str(), f))
        .collect();
    let now: std::collections::BTreeMap<&str, &IndexedFile> =
        files.iter().map(|f| (f.path.as_str(), f)).collect();
    let added: Vec<&str> = now
        .keys()
        .filter(|p| !old.contains_key(*p))
        .copied()
        .collect();
    let deleted: Vec<&str> = old
        .keys()
        .filter(|p| !now.contains_key(*p))
        .copied()
        .collect();
    let modified: Vec<&str> = now
        .iter()
        .filter(|(p, f)| old.get(*p).is_some_and(|o| o.sha256 != f.sha256))
        .map(|(p, _)| *p)
        .collect();
    let branch_switched = base_commit != manifest.base_commit;
    json!({
        "added": added,
        "deleted": deleted,
        "modified": modified,
        "new_files": new_files,
        "branch_switched": branch_switched,
    })
}

fn is_empty_changes(c: &Value) -> bool {
    c["added"].as_array().is_some_and(|a| a.is_empty())
        && c["deleted"].as_array().is_some_and(|a| a.is_empty())
        && c["modified"].as_array().is_some_and(|a| a.is_empty())
        && c["new_files"].as_array().is_some_and(|a| a.is_empty())
        && c["branch_switched"] == json!(false)
}

/// The published generation id and its indexed file list, for read-only consumers such as
/// `search` (C7 #2183). `None` when no generation is published or it cannot be read; the
/// caller reports that through `index_query`'s freshness answer instead of guessing.
pub(crate) fn published_files(project: &Path) -> Option<(String, Vec<IndexedFile>)> {
    let root = fs::canonicalize(project).ok()?;
    let source = IndexIo::new(&root).ok()?;
    let gen_id = read_pointer(&source).ok()??;
    let manifest = load_manifest(&source, &gen_id)?;
    Some((manifest.generation_id, manifest.files))
}

/// Read-only query: resolves the published generation, scans the current state, and
/// reports one of the four freshness states with the detected changes. Never guesses a
/// state it cannot ground: anything unresolvable is `unavailable` with a reason — a
/// broken source (unreadable file, no `HEAD`) is itself a freshness answer, not an op
/// failure, because the state question is exactly what the caller is asking.
pub fn index_query(project: &Path, req: &IndexQueryRequest) -> Result<Value> {
    if req.protocol_version != SCHEMA_VERSION {
        return Err("unsupported protocol_version; supported: 1".into());
    }
    let root = fs::canonicalize(project).map_err(|_| "path unavailable".to_string())?;
    let orphans = orphan_candidates(&root);
    let unavailable = |reason: &str| {
        Ok(json!({
            "protocol_version": SCHEMA_VERSION,
            "schema_version": SCHEMA_VERSION,
            "state": "unavailable",
            "reason": reason,
            "generation_id": Value::Null,
            "changes": Value::Null,
            "orphan_candidates": orphans,
            "network": "disabled",
            "writes": false,
        }))
    };
    let source = match IndexIo::new(&root) {
        Ok(source) => source,
        Err(_) => return unavailable("index directory handle unavailable"),
    };
    // REQ-C6-01: a recorded failed attempt outranks the pointer — the index is known
    // incomplete until the next successful build clears the marker. The check comes
    // before the scan so the state stays readable even while the source is broken.
    let marker = match source.read_state(Path::new(&format!(
        "{}/index/last-build.json",
        source.state
    ))) {
        Ok(Some(bytes)) => match serde_json::from_slice::<BuildFailure>(&bytes) {
            Ok(marker) => Some(marker),
            Err(_) => return unavailable("index failure marker unreadable"),
        },
        Ok(None) => None,
        Err(_) => return unavailable("index failure marker unreadable"),
    };
    if let Some(marker) = marker {
        let pointer = read_pointer(&source).ok().flatten();
        let changes = scan(&root, &source)
            .and_then(|(base, files)| untracked(&root).map(|new| (base, files, new)))
            .ok()
            .and_then(|(base, files, new)| {
                pointer
                    .as_deref()
                    .and_then(|g| load_manifest(&source, g))
                    .map(|m| changes_vs(&m, &base, &files, &new))
            });
        return Ok(json!({
            "protocol_version": SCHEMA_VERSION,
            "schema_version": SCHEMA_VERSION,
            "state": "partial",
            "reason": marker.reason,
            "generation_id": pointer,
            "changes": changes,
            "orphan_candidates": orphans,
            "network": "disabled",
            "writes": false,
        }));
    }
    let scan = match scan(&root, &source).and_then(|(b, f)| untracked(&root).map(|n| (b, f, n))) {
        Ok(v) => v,
        Err(e) => return unavailable(&e),
    };
    let (base_commit, files, new_files) = scan;
    let gen_id = match read_pointer(&source) {
        Ok(Some(id)) => id,
        Ok(None) => return unavailable("no generation has been built"),
        Err(e) => return unavailable(&e),
    };
    let Some(manifest) = load_manifest(&source, &gen_id) else {
        return unavailable("generation manifest unavailable");
    };
    let changes = changes_vs(&manifest, &base_commit, &files, &new_files);
    let state = if is_empty_changes(&changes) {
        "fresh"
    } else {
        "stale"
    };
    Ok(json!({
        "protocol_version": SCHEMA_VERSION,
        "schema_version": SCHEMA_VERSION,
        "state": state,
        "reason": Value::Null,
        "generation_id": gen_id,
        "manifest_base_commit": manifest.base_commit,
        "changes": changes,
        "orphan_candidates": orphans,
        "network": "disabled",
        "writes": false,
    }))
}

/// Read-only Map/Graph query (C6b #2347): the `index_query` freshness answer plus the
/// records of the same generation. Both views are verified against the manifest
/// (digest, header generation id, base commit, file set); any mismatch, a generation
/// built before the views, or a pointer that moved during the query is `unavailable`
/// — never one view from another snapshot.
pub fn index_records(project: &Path, req: &IndexQueryRequest) -> Result<Value> {
    let mut answer = index_query(project, req)?;
    let unavailable = |mut answer: Value, reason: &str| {
        answer["state"] = json!("unavailable");
        answer["reason"] = json!(reason);
        answer["map"] = Value::Null;
        answer["graph"] = Value::Null;
        Ok(answer)
    };
    let Some(gen_id) = answer["generation_id"].as_str().map(str::to_string) else {
        answer["map"] = Value::Null;
        answer["graph"] = Value::Null;
        return Ok(answer);
    };
    let root = fs::canonicalize(project).map_err(|_| "path unavailable".to_string())?;
    let Ok(source) = IndexIo::new(&root) else {
        return unavailable(answer, "index directory handle unavailable");
    };
    if read_pointer(&source).ok().flatten().as_deref() != Some(gen_id.as_str()) {
        return unavailable(answer, "generation changed during the query");
    }
    let Some(manifest) = load_manifest(&source, &gen_id) else {
        return unavailable(answer, "generation manifest unavailable");
    };
    let views = stored_views(&source, &manifest).and_then(|(map, graph)| {
        Ok((
            crate::index_records::document(&map)?,
            crate::index_records::document(&graph)?,
        ))
    });
    match views {
        Ok((map, graph)) => {
            answer["records"] = json!(manifest.records);
            answer["map"] = map;
            answer["graph"] = graph;
            Ok(answer)
        }
        Err(e) => unavailable(answer, &e),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn protocol_gates_refuse_other_versions() {
        let tmp = tempfile::tempdir().unwrap();
        let bad = IndexBuildRequest {
            protocol_version: 2,
        };
        let err = index_build(tmp.path(), &bad).unwrap_err();
        assert_eq!(err, "unsupported protocol_version; supported: 1");
        let badq = IndexQueryRequest {
            protocol_version: 0,
        };
        let err = index_query(tmp.path(), &badq).unwrap_err();
        assert_eq!(err, "unsupported protocol_version; supported: 1");
    }

    #[test]
    fn change_detection_reports_each_category() {
        let m = IndexManifest {
            schema_version: 1,
            generation_id: "gen-x".into(),
            base_commit: "aaaa".into(),
            files: vec![
                IndexedFile {
                    path: "kept.md".into(),
                    sha256: "s".into(),
                    bytes: 1,
                },
                IndexedFile {
                    path: "gone.md".into(),
                    sha256: "s".into(),
                    bytes: 1,
                },
                IndexedFile {
                    path: "changed.md".into(),
                    sha256: "old".into(),
                    bytes: 1,
                },
            ],
            built_at: "2026-11-10T00:00:00Z".into(),
            records: None,
        };
        let now = vec![
            IndexedFile {
                path: "kept.md".into(),
                sha256: "s".into(),
                bytes: 1,
            },
            IndexedFile {
                path: "changed.md".into(),
                sha256: "new".into(),
                bytes: 2,
            },
            IndexedFile {
                path: "added.md".into(),
                sha256: "s".into(),
                bytes: 1,
            },
        ];
        let c = changes_vs(&m, "bbbb", &now, &["scratch.txt".to_string()]);
        assert_eq!(c["added"], json!(["added.md"]));
        assert_eq!(c["deleted"], json!(["gone.md"]));
        assert_eq!(c["modified"], json!(["changed.md"]));
        assert_eq!(c["new_files"], json!(["scratch.txt"]));
        assert_eq!(c["branch_switched"], json!(true));
        assert!(!is_empty_changes(&c));

        // Same base and no deltas is the exact fresh shape.
        let none = changes_vs(&m, "aaaa", &m.files.clone(), &[]);
        assert!(is_empty_changes(&none));
    }

    #[test]
    fn generation_id_is_content_keyed_and_stable() {
        let files = |names: &[(&str, &str)]| {
            names
                .iter()
                .map(|(p, s)| IndexedFile {
                    path: (*p).to_string(),
                    sha256: (*s).to_string(),
                    bytes: 1,
                })
                .collect::<Vec<_>>()
        };
        // Same base and files give the same id regardless of order; any content or
        // base change moves the id.
        let a = content_id("base1", &files(&[("a.md", "s1"), ("b.md", "s2")]), "h");
        let b = content_id("base1", &files(&[("b.md", "s2"), ("a.md", "s1")]), "h");
        assert_eq!(a, b, "file order must not change identity");
        assert_ne!(
            a,
            content_id("base2", &files(&[("a.md", "s1"), ("b.md", "s2")]), "h")
        );
        assert_ne!(
            a,
            content_id("base1", &files(&[("a.md", "OTHER"), ("b.md", "s2")]), "h")
        );
        assert_ne!(a, content_id("base1", &files(&[("a.md", "s1")]), "h"));
        // Same base and files, different reachable history: a different Graph.
        assert_ne!(
            a,
            content_id("base1", &files(&[("a.md", "s1"), ("b.md", "s2")]), "h2")
        );
    }

    #[test]
    fn racing_publication_reuses_only_an_identical_winner() {
        // A winner is reusable only with the same base, files and views: a different
        // base, a missing or different records reference (its own views intact), or a
        // tampered view refuses it.
        for case in [
            "matching",
            "other_base",
            "no_records",
            "other_records",
            "tampered_view",
        ] {
            let matching = case == "matching";
            let tmp = tempfile::tempdir().unwrap();
            let gen_id = content_id("base", &[], "h");
            let file_set = crate::index_records::file_set_digest(&[]);
            let views = crate::index_records::assemble(
                Default::default(),
                Default::default(),
                (&gen_id, "base", &file_set),
                "2026-11-10T00:00:00Z",
            )
            .unwrap();
            let expected = IndexManifest {
                schema_version: SCHEMA_VERSION,
                generation_id: gen_id.clone(),
                base_commit: "base".into(),
                files: vec![],
                built_at: "2026-11-10T00:00:00Z".into(),
                records: Some(RecordsRef {
                    version: crate::index_records::versions(),
                    file_set,
                    observed_at: "2026-11-10T00:00:00Z".into(),
                    map_sha256: crate::digest(&views.map),
                    graph_sha256: crate::digest(&views.graph),
                    map_records: views.map_records,
                    graph_records: views.graph_records,
                }),
            };
            let winner = generations_dir(tmp.path()).join(&expected.generation_id);
            let staging = generations_dir(tmp.path()).join(".tmp-loser");
            fs::create_dir_all(&winner).unwrap();
            fs::create_dir(&staging).unwrap();
            let mut stored = expected.clone();
            let mut winner_map = views.map.clone();
            match case {
                "other_base" => stored.base_commit = "other".into(),
                "no_records" => stored.records = None,
                "other_records" => {
                    if let Some(r) = stored.records.as_mut() {
                        r.observed_at = "2026-11-11T00:00:00Z".into();
                    }
                }
                "tampered_view" => winner_map.extend_from_slice(b"{}\n"),
                _ => {}
            }
            for dir in [&winner, &staging] {
                fs::write(dir.join("graph.jsonl"), &views.graph).unwrap();
            }
            fs::write(winner.join("map.jsonl"), &winner_map).unwrap();
            fs::write(staging.join("map.jsonl"), &views.map).unwrap();
            let winner_bytes = canonical_manifest_bytes(&stored).unwrap();
            let staged_bytes = canonical_manifest_bytes(&expected).unwrap();
            fs::write(winner.join("manifest.json"), &winner_bytes).unwrap();
            fs::write(staging.join("manifest.json"), &staged_bytes).unwrap();
            let source = IndexIo::new(tmp.path()).unwrap();
            let result = publish_staged(
                tmp.path(),
                &source,
                &staging,
                &expected.generation_id,
                &expected,
            );
            if matching {
                assert_eq!(result.unwrap(), "unchanged");
            } else {
                assert!(
                    result.is_err(),
                    "{case}: a mismatched winner cannot be reused"
                );
            }
            assert_eq!(
                fs::read(winner.join("manifest.json")).unwrap(),
                winner_bytes
            );
            assert_eq!(
                fs::read(staging.join("manifest.json")).unwrap(),
                staged_bytes
            );
        }
    }
}

#[cfg(test)]
mod handle_io_tests {
    use super::*;

    #[cfg(unix)]
    fn fixture_git(project: &Path, args: &[&str]) {
        let output = git(project)
            .args([
                "-c",
                "gc.auto=0",
                "-c",
                "maintenance.auto=false",
                "-c",
                "user.name=t",
                "-c",
                "user.email=t@example.com",
            ])
            .args(args)
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .output()
            .unwrap();
        assert!(
            output.status.success(),
            "{}",
            String::from_utf8_lossy(&output.stderr)
        );
    }

    /// AC-3: the views are extracted only from bytes that still match the manifest
    /// digest. A file rewritten after the scan (same length, so only the digest can
    /// tell) refuses the build instead of describing content the manifest does not.
    #[cfg(unix)]
    #[test]
    fn source_rewritten_after_the_scan_refuses_the_views() {
        let tmp = tempfile::tempdir().unwrap();
        let project = tmp.path().join("project");
        fs::create_dir_all(&project).unwrap();
        fs::write(project.join("a.rs"), b"// @MOAI:REQ REQ-C6B-01\n").unwrap();
        fixture_git(&project, &["init", "-q"]);
        fixture_git(&project, &["add", "a.rs"]);
        fixture_git(&project, &["commit", "-q", "-m", "first"]);
        let root = fs::canonicalize(&project).unwrap();
        let source = IndexIo::new(&root).unwrap();
        let (base, files) = scan(&root, &source).unwrap();
        let views = |root: &Path| {
            let (history, digest) = build_history(root, &base);
            let id = content_id(&base, &files, &digest);
            build_views(root, &source, &id, &base, &files, history)
        };
        assert!(views(&root).is_ok());
        fs::write(project.join("a.rs"), b"// @MOAI:REQ REQ-C6B-09\n").unwrap();
        let err = views(&root)
            .err()
            .expect("a rewritten source must refuse the views");
        assert_eq!(err, "index source changed during build: a.rs");
    }

    #[cfg(unix)]
    #[test]
    fn parent_swap_at_open_seams_never_hashes_the_replacement() {
        for point in [ReadPoint::BeforeParentOpen, ReadPoint::BeforeFileOpen] {
            for replaced in ["dir", "dir/nested"] {
                for external in [false, true] {
                    let tmp = tempfile::tempdir().unwrap();
                    let project = tmp.path().join("project");
                    fs::create_dir_all(project.join("dir/nested")).unwrap();
                    fs::write(project.join("dir/nested/x.txt"), b"INSIDE!-SECRET-CONTENT").unwrap();
                    let outside = if external {
                        tmp.path().join("outside")
                    } else {
                        project
                            .join(Path::new(replaced).parent().unwrap())
                            .join("elsewhere")
                    };
                    fs::create_dir_all(outside.join("nested")).unwrap();
                    fs::write(outside.join("x.txt"), b"OUTSIDE-SECRET-CONTENT").unwrap();
                    fs::write(outside.join("nested/x.txt"), b"OUTSIDE-SECRET-CONTENT").unwrap();
                    fixture_git(&project, &["init", "-q", "--initial-branch=main"]);
                    fixture_git(&project, &["add", "dir/nested/x.txt"]);
                    fixture_git(&project, &["commit", "-q", "-m", "fixture"]);
                    let source = IndexIo::new(&project).unwrap();
                    let mut swapped = false;
                    let (_, files) = scan_with_hook(&project, &source, |at, relative| {
                        if !swapped && at == point && relative == Path::new("dir/nested/x.txt") {
                            fs::rename(project.join(replaced), tmp.path().join("kept")).unwrap();
                            let link = if external {
                                outside.as_path()
                            } else {
                                Path::new("elsewhere")
                            };
                            std::os::unix::fs::symlink(link, project.join(replaced)).unwrap();
                            swapped = true;
                        }
                    })
                    .unwrap();
                    assert!(swapped, "{point:?} {replaced}");
                    if point == ReadPoint::BeforeParentOpen {
                        assert_eq!(files, vec![], "{replaced} external={external}");
                    } else {
                        assert_eq!(
                            files,
                            vec![IndexedFile {
                                path: "dir/nested/x.txt".into(),
                                sha256: format!(
                                    "sha256:{:x}",
                                    Sha256::digest(b"INSIDE!-SECRET-CONTENT")
                                ),
                                bytes: 22,
                            }],
                            "{replaced} external={external}"
                        );
                    }
                }
            }
        }
    }

    #[cfg(unix)]
    #[test]
    fn leaf_replaced_after_metadata_is_not_followed() {
        let tmp = tempfile::tempdir().unwrap();
        fs::write(tmp.path().join("x.txt"), b"inside").unwrap();
        fs::write(tmp.path().join("sibling.txt"), b"outside").unwrap();
        fixture_git(tmp.path(), &["init", "-q", "--initial-branch=main"]);
        fixture_git(tmp.path(), &["add", "x.txt"]);
        fixture_git(tmp.path(), &["commit", "-q", "-m", "fixture"]);
        let source = IndexIo::new(tmp.path()).unwrap();
        let result = scan_with_hook(tmp.path(), &source, |point, relative| {
            if point == ReadPoint::BeforeFileOpen && relative == Path::new("x.txt") {
                fs::rename(tmp.path().join("x.txt"), tmp.path().join("kept.txt")).unwrap();
                std::os::unix::fs::symlink("sibling.txt", tmp.path().join("x.txt")).unwrap();
            }
        });
        assert!(
            result.is_err(),
            "a metadata/open replacement must not become a regular digest"
        );
    }

    #[test]
    fn marker_byte_budget_preserves_the_previous_failure_when_exceeded() {
        let tmp = tempfile::tempdir().unwrap();
        let source = IndexIo::new(tmp.path()).unwrap();
        let mut bytes = Vec::with_capacity(33_554_433);
        bytes.extend_from_slice(
            br#"{"status":"partial","reason":"original","at":"2026-11-10T00:00:00Z"}"#,
        );
        bytes.resize(33_554_432, b' ');
        source.write_marker(&bytes).unwrap();
        let request = IndexQueryRequest {
            protocol_version: 1,
        };
        let accepted = index_query(tmp.path(), &request).unwrap();
        assert_eq!(accepted["state"], "partial");
        assert_eq!(accepted["reason"], "original");
        let offset = bytes.windows(8).position(|v| v == b"original").unwrap();
        bytes[offset..offset + 8].copy_from_slice(b"replaced");
        bytes.push(b' ');
        assert_eq!(
            source.write_marker(&bytes).unwrap_err().kind(),
            io::ErrorKind::InvalidData
        );
        let retained = index_query(tmp.path(), &request).unwrap();
        assert_eq!(retained["state"], "partial");
        assert_eq!(retained["reason"], "original");
    }
}
