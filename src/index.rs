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
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    fs,
    io::{self, Read, Write},
    path::{Path, PathBuf},
    process::{Command, Stdio},
};

type Result<T> = std::result::Result<T, String>;

pub const SCHEMA_VERSION: u32 = 1;
/// Scan guard: a tracked-file table beyond this is refused, not silently truncated.
pub const MAX_FILES: usize = 16384;
const GIT_OUTPUT_LIMIT: usize = 4 * 1024 * 1024;
/// A manifest file-count cap on read (a hand-edited generation must not blow memory).
const MAX_MANIFEST_FILES: usize = MAX_FILES;

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
    cmd.arg("--no-optional-locks").arg("-C").arg(dir);
    crate::identity::scrub_git_env(&mut cmd);
    cmd.env("GIT_OPTIONAL_LOCKS", "0").stdin(Stdio::null());
    cmd
}

fn git_out(dir: &Path, args: &[&str]) -> Option<Vec<u8>> {
    let out = git(dir).args(args).stderr(Stdio::null()).output().ok()?;
    (out.status.success() && out.stdout.len() <= GIT_OUTPUT_LIMIT).then_some(out.stdout)
}

fn git_line(dir: &Path, args: &[&str]) -> Option<String> {
    let s = String::from_utf8(git_out(dir, args)?).ok()?;
    Some(s.trim_end_matches('\n').to_string()).filter(|s| !s.is_empty() && !s.contains('\n'))
}

fn file_sha(path: &Path) -> Option<(String, u64)> {
    let mut h = Sha256::new();
    let mut f = fs::File::open(path).ok()?;
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
    project.join(crate::PROJECT_STATE_DIR).join("index")
}

fn generations_dir(project: &Path) -> PathBuf {
    index_root(project).join("generations")
}

fn pointer_path(project: &Path) -> PathBuf {
    index_root(project).join("current.json")
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
fn scan(project: &Path) -> Result<(String, Vec<IndexedFile>)> {
    let raw = git_out(project, &["ls-files", "-z"])
        .ok_or("index source unavailable: git ls-files failed")?;
    let mut paths: Vec<String> = raw
        .split(|b| *b == 0)
        .filter(|e| !e.is_empty())
        .map(|e| String::from_utf8_lossy(e).into_owned())
        .filter(|p| !p.starts_with(".moai-cli/index/") && p != ".moai-cli/index")
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
        let full = project.join(p);
        let meta = match fs::symlink_metadata(&full) {
            Ok(meta) => meta,
            Err(e) if e.kind() == io::ErrorKind::NotFound => continue,
            Err(_) => return Err(format!("index source unavailable: {p} metadata unreadable")),
        };
        let (sha, bytes) = if meta.file_type().is_symlink() {
            let t: String = fs::read_link(&full)
                .ok()
                .and_then(|t| t.into_os_string().into_string().ok())
                .ok_or_else(|| format!("index source unavailable: {p} symlink unreadable"))?;
            let d = Sha256::digest(t.as_bytes());
            (format!("symlink-sha256:{d:x}"), t.len() as u64)
        } else if meta.is_file() {
            file_sha(&full).ok_or_else(|| format!("index source unavailable: {p} read failed"))?
        } else {
            // A tracked directory entry is a gitlink (submodule); its commit is the state.
            let head = git_line(&full, &["rev-parse", "--verify", "HEAD^{commit}"])
                .ok_or_else(|| format!("index source unavailable: {p} submodule unreadable"))?;
            (format!("commit:{head}"), 0)
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
    )
    .ok_or("index source unavailable: git status failed")?;
    let mut out = Vec::new();
    for e in raw.split(|b| *b == 0) {
        if e.len() > 3 && e.starts_with(b"?? ") {
            let path = String::from_utf8_lossy(&e[3..]).into_owned();
            // The tool's own state directory (this index, receipts, claims) is never a
            // project change: the index writes there itself, so counting it would make
            // `fresh` unreachable.
            if path.starts_with(".moai-cli/") {
                continue;
            }
            out.push(path);
        }
    }
    out.sort();
    out.truncate(MAX_FILES);
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
fn content_id(base_commit: &str, files: &[IndexedFile]) -> String {
    let mut sorted: Vec<&IndexedFile> = files.iter().collect();
    sorted.sort_by(|a, b| a.path.cmp(&b.path));
    let mut h = Sha256::new();
    h.update(b"moai-index-generation-v1\0");
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
    let pointer = crate::apply::target(project, ".moai-cli/index/current.json")?;
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

/// A simultaneous publisher may win the rename. Reuse only a validated identical
/// snapshot; keep this attempt's staging directory as an orphan, never delete data.
fn publish_staged(
    project: &Path,
    staging: &Path,
    generation_id: &str,
    expected: &IndexManifest,
) -> Result<&'static str> {
    match fs::rename(staging, generations_dir(project).join(generation_id)) {
        Ok(()) => Ok("created"),
        Err(e) => {
            if load_manifest(project, generation_id).is_some_and(|stored| {
                stored.base_commit == expected.base_commit && stored.files == expected.files
            }) {
                Ok("unchanged")
            } else {
                Err(format!("index generation staging write failed: {e}"))
            }
        }
    }
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
    let build = || -> Result<Value> {
        for rel in [
            ".moai-cli/index/generations",
            ".moai-cli/index/current.json",
            ".moai-cli/index/last-build.json",
        ] {
            crate::apply::target(&root, rel)?;
        }
        let (base_commit, files) = scan(&root)?;
        let total_bytes: u64 = files.iter().map(|f| f.bytes).sum();
        let gen_id = content_id(&base_commit, &files);
        let manifest = IndexManifest {
            schema_version: SCHEMA_VERSION,
            generation_id: gen_id.clone(),
            base_commit: base_commit.clone(),
            files: files.clone(),
            built_at: utc_now(),
        };
        let gens = generations_dir(&root);
        fs::create_dir_all(&gens).map_err(|_| "index directory unavailable".to_string())?;
        let gen_dir = gens.join(&gen_id);
        crate::apply::target(
            &root,
            &format!(".moai-cli/index/generations/{gen_id}/manifest.json"),
        )?;
        let action = if !gen_dir.is_dir() {
            let nanos = std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .map(|d| d.as_nanos())
                .unwrap_or(0);
            let tmp = gens.join(format!(".tmp-{}-{nanos}", std::process::id()));
            fs::create_dir(&tmp)
                .map_err(|_| "index generation staging write failed".to_string())?;
            let bytes = canonical_manifest_bytes(&manifest)?;
            let mut staged = fs::OpenOptions::new()
                .write(true)
                .create_new(true)
                .open(tmp.join("manifest.json"))
                .map_err(|e| format!("index generation staging write failed: {e}"))?;
            staged
                .write_all(&bytes)
                .map_err(|e| format!("index generation staging write failed: {e}"))?;
            drop(staged);
            publish_staged(&root, &tmp, &gen_id, &manifest)?
        } else {
            // Reuse is successful only when the immutable manifest still describes
            // this snapshot. Never publish a damaged cache or overwrite it in place.
            let stored = load_manifest(&root, &gen_id)
                .ok_or("index generation manifest invalid; refusing reuse")?;
            if stored.base_commit != base_commit || stored.files != files {
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
            if let Ok(path) = crate::apply::target(&root, ".moai-cli/index/last-build.json") {
                let dir = index_root(&root);
                if fs::create_dir_all(&dir).is_ok() {
                    if let Ok(bytes) = serde_json::to_vec_pretty(&marker) {
                        let _ = fs::write(path, bytes);
                    }
                }
            }
            Err(format!("index build failed; previous generation kept: {e}"))
        }
    }
}

fn read_pointer(project: &Path) -> Result<Option<String>> {
    let bytes = match fs::read(pointer_path(project)) {
        Ok(b) => b,
        Err(e) if e.kind() == io::ErrorKind::NotFound => return Ok(None),
        Err(_) => return Ok(None),
    };
    let v: Value = match serde_json::from_slice(&bytes) {
        Ok(v) => v,
        Err(_) => return Ok(None),
    };
    Ok(v["generation_id"]
        .as_str()
        .map(str::to_string)
        .filter(|s| generation_id_valid(s)))
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

fn load_manifest(project: &Path, gen_id: &str) -> Option<IndexManifest> {
    let path = crate::apply::target(
        project,
        &format!(".moai-cli/index/generations/{gen_id}/manifest.json"),
    )
    .ok()?;
    let bytes = fs::read(path).ok()?;
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
    // REQ-C6-01: a recorded failed attempt outranks the pointer — the index is known
    // incomplete until the next successful build clears the marker. The check comes
    // before the scan so the state stays readable even while the source is broken.
    if let Ok(bytes) = fs::read(failure_marker_path(&root)) {
        if let Ok(marker) = serde_json::from_slice::<BuildFailure>(&bytes) {
            let pointer = read_pointer(&root)?;
            let changes = match scan(&root) {
                Ok((base_commit, files)) => {
                    let new_files = untracked(&root).unwrap_or_default();
                    pointer
                        .as_deref()
                        .and_then(|g| load_manifest(&root, g))
                        .map(|m| changes_vs(&m, &base_commit, &files, &new_files))
                }
                // The source is broken; the partial state is still the answer.
                Err(_) => None,
            };
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
    }
    let scan = match scan(&root).and_then(|(b, f)| untracked(&root).map(|n| (b, f, n))) {
        Ok(v) => v,
        Err(e) => return unavailable(&e),
    };
    let (base_commit, files, new_files) = scan;
    let pointer_exists = pointer_path(&root).is_file();
    let Some(gen_id) = read_pointer(&root)? else {
        // Distinguish "never built" from "the pointer is damaged": both are
        // unavailable (cache loss is a performance loss), with honest reasons.
        return unavailable(if pointer_exists {
            "index pointer unreadable"
        } else {
            "no generation has been built"
        });
    };
    let Some(manifest) = load_manifest(&root, &gen_id) else {
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
        let a = content_id("base1", &files(&[("a.md", "s1"), ("b.md", "s2")]));
        let b = content_id("base1", &files(&[("b.md", "s2"), ("a.md", "s1")]));
        assert_eq!(a, b, "file order must not change identity");
        assert_ne!(
            a,
            content_id("base2", &files(&[("a.md", "s1"), ("b.md", "s2")]))
        );
        assert_ne!(
            a,
            content_id("base1", &files(&[("a.md", "OTHER"), ("b.md", "s2")]))
        );
        assert_ne!(a, content_id("base1", &files(&[("a.md", "s1")])));
    }

    #[test]
    fn racing_publication_reuses_only_an_identical_winner() {
        for matching in [true, false] {
            let tmp = tempfile::tempdir().unwrap();
            let expected = IndexManifest {
                schema_version: SCHEMA_VERSION,
                generation_id: content_id("base", &[]),
                base_commit: "base".into(),
                files: vec![],
                built_at: "2026-11-10T00:00:00Z".into(),
            };
            let winner = generations_dir(tmp.path()).join(&expected.generation_id);
            let staging = generations_dir(tmp.path()).join(".tmp-loser");
            fs::create_dir_all(&winner).unwrap();
            fs::create_dir(&staging).unwrap();
            let mut stored = expected.clone();
            if !matching {
                stored.base_commit = "other".into();
            }
            let winner_bytes = canonical_manifest_bytes(&stored).unwrap();
            let staged_bytes = canonical_manifest_bytes(&expected).unwrap();
            fs::write(winner.join("manifest.json"), &winner_bytes).unwrap();
            fs::write(staging.join("manifest.json"), &staged_bytes).unwrap();
            let result = publish_staged(tmp.path(), &staging, &expected.generation_id, &expected);
            if matching {
                assert_eq!(result.unwrap(), "unchanged");
            } else {
                assert!(result.is_err(), "a mismatched winner cannot be reused");
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
