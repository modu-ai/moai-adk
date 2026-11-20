//! The Go `moai bundle add | remove | list` port (SPEC-20261120-user-asset-bundles): the
//! opt-in user-asset packs. The optional-pack assets and their catalog data live in
//! `templates/bundle-packs/`, embedded by `build.rs`; the manifest, its lock and the
//! pending-install journal keep the Go paths and JSON shapes under `~/.moai/`, so a user
//! switches binaries without losing state. The output contract is the Go summary line and
//! its reason lines, carried in the result's `text` field (the Rust operations print
//! envelopes; `src/ops/bundle.rs` mounts the text).

pub mod embedded {
    include!(concat!(env!("OUT_DIR"), "/bundle_pack_files.rs"));
}

use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::io::Write;
use std::path::{Path, PathBuf};
use std::time::{Duration, SystemTime};

type Result<T> = std::result::Result<T, String>;

// ---------------------------------------------------------------------------
// catalog

/// The packs data: names, descriptions, entries (name, kind, deploy root, path,
/// depends_skills) and the base name set, parsed from the embedded JSON.
pub struct Catalog {
    pub base_names: Vec<String>,
    pub packs: Vec<(String, Pack)>,
}

pub struct Pack {
    pub entries: Vec<Entry>,
}

pub struct Entry {
    pub name: String,
    pub kind: String,
    pub root: String,
    pub path: String,
    pub depends_skills: Vec<String>,
}

pub fn catalog() -> Result<Catalog> {
    let data: Value =
        serde_json::from_str(embedded::PACKS_JSON).map_err(|e| format!("packs data: {e}"))?;
    let mut packs = Vec::new();
    for (name, p) in data["packs"]
        .as_object()
        .ok_or("packs data has no packs object")?
    {
        let entries = p["entries"]
            .as_array()
            .ok_or_else(|| format!("pack {name}: no entries array"))?
            .iter()
            .map(|e| Entry {
                name: e["name"].as_str().unwrap_or_default().to_string(),
                kind: e["kind"].as_str().unwrap_or_default().to_string(),
                root: e["root"].as_str().unwrap_or_default().to_string(),
                path: e["path"].as_str().unwrap_or_default().to_string(),
                depends_skills: e["depends_skills"]
                    .as_array()
                    .map(|a| {
                        a.iter()
                            .filter_map(|d| d.as_str().map(str::to_string))
                            .collect()
                    })
                    .unwrap_or_default(),
            })
            .collect();
        packs.push((name.clone(), Pack { entries }));
    }
    packs.sort_by(|a, b| a.0.cmp(&b.0));
    let base_names = data["base_names"]
        .as_array()
        .map(|a| {
            a.iter()
                .filter_map(|b| b.as_str().map(str::to_string))
                .collect()
        })
        .unwrap_or_default();
    Ok(Catalog { base_names, packs })
}

impl Catalog {
    /// The pack named `name`, or None.
    pub fn pack(&self, name: &str) -> Option<&Pack> {
        self.packs.iter().find(|(n, _)| n == name).map(|(_, p)| p)
    }

    /// The valid pack names in ascending order (the refusal's list).
    pub fn names(&self) -> Vec<&str> {
        self.packs.iter().map(|(n, _)| n.as_str()).collect()
    }

    /// The bytes of one file under the entry's folder, from the embedded table.
    pub fn file_bytes(&self, entry_path: &str, rel: &str) -> Option<&'static [u8]> {
        let prefix = format!("{entry_path}/");
        embedded::PACK_FILES
            .iter()
            .find(|(p, _)| p.strip_prefix(&prefix) == Some(rel))
            .map(|(_, bytes)| *bytes)
    }

    /// The file list under one entry's folder, in sorted order.
    pub fn entry_files(&self, entry_path: &str) -> Vec<&'static str> {
        let prefix = format!("{entry_path}/");
        embedded::PACK_FILES
            .iter()
            .filter_map(|(p, _)| p.strip_prefix(&prefix))
            .collect()
    }

    /// True when the entry path is itself one embedded file (an agent), not a folder of
    /// files; its install key and target are the whole path, with no extra segment.
    pub fn entry_is_file(&self, entry_path: &str) -> bool {
        embedded::PACK_FILES.iter().any(|(p, _)| *p == entry_path)
    }

    /// The bytes of one whole-path embedded file.
    pub fn file_bytes_at(&self, entry_path: &str) -> Option<&'static [u8]> {
        embedded::PACK_FILES
            .iter()
            .find(|(p, _)| *p == entry_path)
            .map(|(_, bytes)| *bytes)
    }

    /// Every installable file of one entry as (path under the deploy root, bytes): one item
    /// for a file entry (an agent), one per file for a folder entry (a skill).
    pub fn entry_items(&self, entry_path: &str) -> Vec<(String, &'static [u8])> {
        if self.entry_is_file(entry_path) {
            let under = entry_under_root(entry_path);
            return self
                .file_bytes_at(entry_path)
                .map(|bytes| vec![(under, bytes)])
                .unwrap_or_default();
        }
        let prefix = format!("{entry_path}/");
        let folder = entry_path.rsplit('/').next().unwrap_or_default();
        embedded::PACK_FILES
            .iter()
            .filter_map(|(p, bytes)| {
                p.strip_prefix(&prefix)
                    .map(|rel| (format!("{folder}/{rel}"), *bytes))
            })
            .collect()
    }
}

// ---------------------------------------------------------------------------
// paths (the Go names, userassets/paths.go)

/// `~/.moai/user-assets.json` — the Go manifest path.
pub fn manifest_path(home: &Path) -> PathBuf {
    home.join(".moai").join("user-assets.json")
}

/// `~/.moai/user-assets.lock` — the Go lock path.
pub fn lock_path(home: &Path) -> PathBuf {
    home.join(".moai").join("user-assets.lock")
}

/// `~/.moai/user-assets.acquire-guard.guard` — the Go flock guard's path.
pub fn guard_path(home: &Path) -> PathBuf {
    home.join(".moai").join("user-assets.acquire-guard.guard")
}

/// `~/.moai/user-assets-journal.json` — the Go journal path.
pub fn journal_path(home: &Path) -> PathBuf {
    home.join(".moai").join("user-assets-journal.json")
}

/// The deploy root of a slug under the user's home.
fn deploy_root(home: &Path, slug: &str) -> Result<PathBuf> {
    match slug {
        "claude-skills" => Ok(home.join(".claude").join("skills")),
        "claude-agents" => Ok(home.join(".claude").join("agents")),
        "agents-skills" => Ok(home.join(".agents").join("skills")),
        "codex-agents" => Ok(home.join(".codex").join("agents")),
        other => Err(format!("unknown deploy root {other}")),
    }
}

/// The path of an entry under its deploy root: the catalog path without the section's two
/// segments (`.claude/agents/moai/x.md` deploys as `moai/x.md`).
fn entry_under_root(entry_path: &str) -> String {
    entry_path
        .splitn(3, '/')
        .nth(2)
        .unwrap_or(entry_path)
        .to_string()
}

// ---------------------------------------------------------------------------
// summary (the Go output contract, cli/user_asset_phase.go writeInstallSummary)

/// The counts and reason lines of one add or remove run.
pub struct Summary {
    pub label: String,
    pub installed: usize,
    pub refreshed: usize,
    pub removed: usize,
    pub collision_skipped: usize,
    pub divergence_preserved: usize,
    pub lines: Vec<String>,
}

impl Summary {
    fn new(label: impl Into<String>) -> Self {
        Self {
            label: label.into(),
            installed: 0,
            refreshed: 0,
            removed: 0,
            collision_skipped: 0,
            divergence_preserved: 0,
            lines: Vec::new(),
        }
    }

    /// The Go summary line plus the reason lines, newline-joined.
    pub fn go_text(&self) -> String {
        let mut text = format!(
            "{}: {} installed, {} refreshed, {} removed, {} collision-skipped, {} divergence-preserved",
            self.label,
            self.installed,
            self.refreshed,
            self.removed,
            self.collision_skipped,
            self.divergence_preserved
        );
        for line in &self.lines {
            text.push('\n');
            text.push_str(line);
        }
        text
    }

    pub fn value(&self) -> Value {
        json!({
            "label": self.label,
            "installed": self.installed,
            "refreshed": self.refreshed,
            "removed": self.removed,
            "collision_skipped": self.collision_skipped,
            "divergence_preserved": self.divergence_preserved,
            "lines": self.lines,
            "text": self.go_text(),
        })
    }
}

// ---------------------------------------------------------------------------
// manifest (Go userassets/manifest.go: schema 1, unknown top fields preserved)

const SCHEMA_VERSION: u64 = 1;

fn fresh_manifest() -> Value {
    json!({
        "schema_version": SCHEMA_VERSION,
        "bundles": [],
        "files": {},
        "collisions": [],
    })
}

fn load_manifest(home: &Path) -> Result<Value> {
    let path = manifest_path(home);
    let bytes = match std::fs::read(&path) {
        Ok(bytes) => bytes,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(fresh_manifest()),
        Err(e) => {
            return Err(format!(
                "user-assets: read manifest {}: {e}",
                path.display()
            ))
        }
    };
    let value: Value = serde_json::from_slice(&bytes).map_err(|e| {
        format!(
            "user-assets: the manifest at {} is corrupt: {e}",
            path.display()
        )
    })?;
    Ok(value)
}

fn save_manifest(home: &Path, manifest: &Value) -> Result<()> {
    let path = manifest_path(home);
    if let Some(parent) = path.parent() {
        std::fs::create_dir_all(parent)
            .map_err(|e| format!("user-assets: mkdir {}: {e}", parent.display()))?;
    }
    let mut text = serde_json::to_string_pretty(manifest)
        .map_err(|e| format!("user-assets: manifest serialise: {e}"))?;
    text.push('\n');
    let temp = path.with_extension("json.tmp");
    std::fs::write(&temp, text)
        .map_err(|e| format!("user-assets: write {}: {e}", temp.display()))?;
    std::fs::rename(&temp, &path)
        .map_err(|e| format!("user-assets: save {}: {e}", path.display()))?;
    Ok(())
}

/// The manifest's `bundles` as a sorted Vec<String>.
fn bundles_of(manifest: &Value) -> Vec<String> {
    let mut bundles: Vec<String> = manifest["bundles"]
        .as_array()
        .map(|a| {
            a.iter()
                .filter_map(|b| b.as_str().map(str::to_string))
                .collect()
        })
        .unwrap_or_default();
    bundles.sort();
    bundles
}

fn set_bundles(manifest: &mut Value, bundles: Vec<String>) {
    if let Some(object) = manifest.as_object_mut() {
        object.insert(
            "bundles".into(),
            Value::Array(bundles.into_iter().map(Value::from).collect()),
        );
    }
}

/// One `{path, first_seen_at}` collision record, appended when the path is new.
fn record_collision(manifest: &mut Value, key: &str, now: &str) {
    let collisions = match manifest.get_mut("collisions").and_then(Value::as_array_mut) {
        Some(list) => list,
        None => {
            if let Some(object) = manifest.as_object_mut() {
                object.insert("collisions".into(), Value::Array(Vec::new()));
            }
            manifest
                .get_mut("collisions")
                .and_then(Value::as_array_mut)
                .expect("collisions array")
        }
    };
    let known = collisions
        .iter()
        .any(|c| c.get("path").and_then(Value::as_str) == Some(key));
    if !known {
        collisions.push(json!({"path": key, "first_seen_at": now}));
    }
}

fn set_file_entry(manifest: &mut Value, key: &str, sha: &str, bundle: &str, now: &str) {
    let entry = json!({
        "sha256": sha,
        "bundle": bundle,
        "installed_at": now,
        "moai_version": crate::BUILD_VERSION,
    });
    if let Some(files) = manifest.get_mut("files").and_then(Value::as_object_mut) {
        files.insert(key.to_string(), entry);
    }
}

// ---------------------------------------------------------------------------
// lock (Go userassets/lock.go: O_EXCL lock file + flock acquire guard)

/// The held user-level lock; the file is removed when the value drops.
pub struct UserLock {
    path: PathBuf,
    token: String,
}

impl Drop for UserLock {
    fn drop(&mut self) {
        // Remove the lock file only while it still names our token.
        if let Ok(text) = std::fs::read_to_string(&self.path) {
            if text.contains(&self.token) {
                let _ = std::fs::remove_file(&self.path);
            }
        }
    }
}

const LOCK_WAIT: Duration = Duration::from_secs(120);
const LOCK_STALE: Duration = Duration::from_secs(30 * 60);

#[cfg(unix)]
fn flock_exclusive(path: &Path, deadline: SystemTime) -> Result<std::fs::File> {
    use std::os::unix::io::AsRawFd;
    let file = std::fs::OpenOptions::new()
        .create(true)
        .truncate(false)
        .write(true)
        .read(true)
        .open(path)
        .map_err(|e| format!("user-assets: open guard {}: {e}", path.display()))?;
    loop {
        let code = unsafe { libc::flock(file.as_raw_fd(), libc::LOCK_EX | libc::LOCK_NB) };
        if code == 0 {
            // The guard holds for the File's lifetime; closing it releases the flock.
            return Ok(file);
        }
        if SystemTime::now() >= deadline {
            return Err(format!(
                "user-assets: the acquire guard at {} stayed held (2m)",
                path.display()
            ));
        }
        std::thread::sleep(Duration::from_millis(5));
    }
}

/// The same guard where there is no `flock`: the standard library's advisory file lock, retried
/// until the deadline like the unix one.
#[cfg(not(unix))]
fn flock_exclusive(path: &Path, deadline: SystemTime) -> Result<std::fs::File> {
    let file = std::fs::OpenOptions::new()
        .create(true)
        .truncate(false)
        .write(true)
        .read(true)
        .open(path)
        .map_err(|e| format!("user-assets: open guard {}: {e}", path.display()))?;
    loop {
        match file.try_lock() {
            Ok(()) => return Ok(file),
            Err(std::fs::TryLockError::WouldBlock) => {}
            Err(std::fs::TryLockError::Error(e)) => {
                return Err(format!("user-assets: lock guard {}: {e}", path.display()));
            }
        }
        if SystemTime::now() >= deadline {
            return Err(format!(
                "user-assets: the acquire guard at {} stayed held (2m)",
                path.display()
            ));
        }
        std::thread::sleep(Duration::from_millis(5));
    }
}

/// Takes the user-level lock exactly as Go does: the flock guard serialises acquirers, the
/// O_EXCL lock file is retried until the deadline, and a lock older than 30 minutes is
/// stale and reclaimed.
pub fn acquire_user_lock(home: &Path) -> Result<UserLock> {
    let path = lock_path(home);
    if let Some(parent) = path.parent() {
        std::fs::create_dir_all(parent)
            .map_err(|e| format!("user-assets: mkdir {}: {e}", parent.display()))?;
    }
    let deadline = SystemTime::now() + LOCK_WAIT;
    let guard = flock_exclusive(&guard_path(home), deadline)?;
    let token = format!("pid={}-{}", std::process::id(), now_rfc3339_nanos());
    let outcome = loop {
        match std::fs::OpenOptions::new()
            .create_new(true)
            .write(true)
            .open(&path)
        {
            Ok(mut file) => {
                writeln!(
                    file,
                    "pid={} token={token} acquired={}",
                    std::process::id(),
                    now_rfc3339()
                )
                .ok();
                break Ok(());
            }
            Err(e) if e.kind() == std::io::ErrorKind::AlreadyExists => {
                let stale = std::fs::metadata(&path)
                    .and_then(|m| m.modified())
                    .ok()
                    .and_then(|modified| SystemTime::now().duration_since(modified).ok())
                    .is_some_and(|age| age > LOCK_STALE);
                if stale {
                    let _ = std::fs::remove_file(&path);
                    continue;
                }
                if SystemTime::now() >= deadline {
                    break Err(format!(
                        "user-assets: the lock at {} stayed held (2m)",
                        path.display()
                    ));
                }
                std::thread::sleep(Duration::from_millis(5));
            }
            Err(e) => {
                break Err(format!("user-assets: lock {}: {e}", path.display()));
            }
        }
    };
    drop(guard);
    outcome.map(|_| UserLock { path, token })
}

/// The UTC RFC3339 second form (`2026-11-20T01:56:07Z`), the Go stamp.
fn now_rfc3339() -> String {
    crate::update::now_rfc3339()
}

/// The UTC RFC3339 form with nanoseconds, the Go lock token stamp.
fn now_rfc3339_nanos() -> String {
    let now = SystemTime::now()
        .duration_since(SystemTime::UNIX_EPOCH)
        .unwrap_or_default();
    format!(
        "{}.{:09}Z",
        now_rfc3339().trim_end_matches('Z'),
        now.subsec_nanos()
    )
}

// ---------------------------------------------------------------------------
// journal (Go userassets/journal.go: the intent record, reconciled on the next run)

fn load_journal(home: &Path) -> Result<Option<Value>> {
    let path = journal_path(home);
    match std::fs::read(&path) {
        Ok(bytes) => {
            let value: Value = serde_json::from_slice(&bytes).map_err(|e| {
                format!(
                    "user-assets: the journal at {} is corrupt: {e}",
                    path.display()
                )
            })?;
            Ok(Some(value))
        }
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(None),
        Err(e) => Err(format!("user-assets: read journal {}: {e}", path.display())),
    }
}

fn clear_journal(home: &Path) -> Result<()> {
    match std::fs::remove_file(journal_path(home)) {
        Ok(()) => Ok(()),
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(()),
        Err(e) => Err(format!("user-assets: clear journal: {e}")),
    }
}

/// Reconciles an existing journal before any install or collision judgment: an absent
/// target installs, a hash-equal target is claimed as this run's own, a mismatched target
/// is preserved as a collision.
fn reconcile_journal(home: &Path, manifest: &mut Value, summary: &mut Summary) -> Result<()> {
    let Some(journal) = load_journal(home)? else {
        return Ok(());
    };
    let now = &now_rfc3339();
    let entries = journal["entries"].as_array().cloned().unwrap_or_default();
    for entry in entries {
        let key = entry["path"].as_str().unwrap_or_default().to_string();
        let expected = entry["expected_sha256"]
            .as_str()
            .unwrap_or_default()
            .to_string();
        let (root, rel) = key.split_once('/').unwrap_or(("", &key));
        let target = deploy_root(home, root)?.join(rel);
        match std::fs::read(&target) {
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => {
                // Install from the embedded tree when the bytes still match the intent.
                if let Some(bytes) = embedded_file_by_key(&key) {
                    if sha256_hex(bytes) == expected {
                        write_target(&target, bytes)?;
                        set_file_entry(
                            manifest,
                            &key,
                            &expected,
                            entry["bundle"].as_str().unwrap_or(""),
                            now,
                        );
                        summary.installed += 1;
                    }
                }
            }
            Ok(bytes) => {
                if sha256_hex(&bytes) == expected {
                    set_file_entry(
                        manifest,
                        &key,
                        &expected,
                        entry["bundle"].as_str().unwrap_or(""),
                        now,
                    );
                    summary.installed += 1;
                } else {
                    record_collision(manifest, &key, now);
                    summary.collision_skipped += 1;
                    summary.lines.push(format!(
                        "  collision (skipped, your file kept): {key} — rename it or remove it, then re-run to install"
                    ));
                }
            }
            Err(e) => return Err(format!("user-assets: read {}: {e}", target.display())),
        }
    }
    clear_journal(home)?;
    Ok(())
}

/// The embedded bytes of one manifest key (`claude-skills/<name>/<file>`).
fn embedded_file_by_key(key: &str) -> Option<&'static [u8]> {
    let (root, rest) = key.split_once('/')?;
    let (dir, file) = rest.split_once('/')?;
    let prefix = match root {
        "claude-skills" => format!(".claude/skills/{dir}/"),
        "claude-agents" => format!(".claude/agents/{dir}/"),
        "agents-skills" => format!(".agents/skills/{dir}/"),
        "codex-agents" => format!(".codex/agents/{dir}/"),
        _ => return None,
    };
    embedded::PACK_FILES
        .iter()
        .find(|(p, _)| p == &format!("{prefix}{file}"))
        .map(|(_, bytes)| *bytes)
}

fn sha256_hex(bytes: &[u8]) -> String {
    let mut hasher = Sha256::new();
    hasher.update(bytes);
    format!("{:x}", hasher.finalize())
}

fn write_target(target: &Path, bytes: &[u8]) -> Result<()> {
    if let Some(parent) = target.parent() {
        std::fs::create_dir_all(parent)
            .map_err(|e| format!("user-assets: mkdir {}: {e}", parent.display()))?;
    }
    std::fs::write(target, bytes)
        .map_err(|e| format!("user-assets: write {}: {e}", target.display()))
}

// ---------------------------------------------------------------------------
// the three commands

/// The Go refusal for a name absent from the catalog.
fn unknown_bundle(name: &str, catalog: &Catalog) -> String {
    format!(
        "unknown bundle {name:?} — valid bundles: {}",
        catalog.names().join(", ")
    )
}

/// `moai bundle add <pack>`: record the selection and install the pack's entries.
pub fn add(home: &Path, name: &str) -> Result<Value> {
    let catalog = catalog()?;
    add_with(home, &catalog, name).map(|summary| summary.value())
}

fn add_with(home: &Path, catalog: &Catalog, name: &str) -> Result<Summary> {
    if catalog.pack(name).is_none() {
        return Err(unknown_bundle(name, catalog));
    }
    let lock = acquire_user_lock(home)?;
    let _ = &lock;
    let mut manifest = load_manifest(home)?;
    let mut summary = Summary::new(format!("bundle add {name}"));
    reconcile_journal(home, &mut manifest, &mut summary)?;

    let pack = catalog.pack(name).expect("checked above");
    let now = now_rfc3339();
    let mut intent = Vec::new();
    for entry in &pack.entries {
        for (under_root, bytes) in catalog.entry_items(&entry.path) {
            let key = format!("{}/{}", entry.root, under_root);
            intent.push((key, sha256_hex(bytes)));
        }
    }
    write_journal(home, name, &manifest, &intent)?;

    for entry in &pack.entries {
        for (under_root, bytes) in catalog.entry_items(&entry.path) {
            let key = format!("{}/{}", entry.root, under_root);
            let target = deploy_root(home, &entry.root)?.join(&under_root);
            match std::fs::read(&target) {
                Err(e) if e.kind() == std::io::ErrorKind::NotFound => {
                    write_target(&target, bytes)?;
                    set_file_entry(&mut manifest, &key, &sha256_hex(bytes), name, &now);
                    summary.installed += 1;
                }
                Ok(present) => {
                    if present == bytes {
                        set_file_entry(&mut manifest, &key, &sha256_hex(bytes), name, &now);
                        summary.refreshed += 1;
                    } else {
                        record_collision(&mut manifest, &key, &now);
                        summary.collision_skipped += 1;
                        summary.lines.push(format!(
                            "  collision (skipped, your file kept): {key} — rename it or remove it, then re-run to install"
                        ));
                    }
                }
                Err(e) => return Err(format!("user-assets: read {}: {e}", target.display())),
            }
        }
    }

    let mut bundles = bundles_of(&manifest);
    if !bundles.iter().any(|b| b == name) {
        bundles.push(name.to_string());
        // Go's addSelection appends; the list stays in insertion order.
        if let Some(object) = manifest.as_object_mut() {
            object.insert(
                "bundles".into(),
                Value::Array(bundles.iter().map(|b| Value::from(b.as_str())).collect()),
            );
        }
    }
    save_manifest(home, &manifest)?;
    clear_journal(home)?;
    Ok(summary)
}

fn write_journal(
    home: &Path,
    bundle: &str,
    manifest: &Value,
    intent: &[(String, String)],
) -> Result<()> {
    let now = now_rfc3339();
    let entries: Vec<Value> = intent
        .iter()
        .map(|(key, sha)| {
            json!({
                "path": key,
                "expected_sha256": sha,
                "bundle": bundle,
                "moai_version": crate::BUILD_VERSION,
                "installed_at": manifest["files"][key]["installed_at"].as_str().unwrap_or(now.as_str()),
                "complete": false,
            })
        })
        .collect();
    let selection = bundles_of(manifest);
    let mut selection = selection;
    if !selection.iter().any(|b| b == bundle) {
        selection.push(bundle.to_string());
    }
    let journal = json!({
        "schema_version": SCHEMA_VERSION,
        "bundles_selection": selection,
        "entries": entries,
    });
    if let Some(parent) = journal_path(home).parent() {
        std::fs::create_dir_all(parent).ok();
    }
    std::fs::write(
        journal_path(home),
        serde_json::to_string_pretty(&journal).unwrap_or_default(),
    )
    .map_err(|e| format!("user-assets: write journal: {e}"))
}

/// `moai bundle remove <pack>`: the complement rule — entries shared with the base set or a
/// remaining selection survive, a declared dependency of a preserved asset is deferred, the
/// rest is deleted; the manifest keeps the remaining selection and the journal is cleared.
pub fn remove(home: &Path, name: &str) -> Result<Value> {
    let catalog = catalog()?;
    remove_with(home, &catalog, name).map(|summary| summary.value())
}

fn remove_with(home: &Path, catalog: &Catalog, name: &str) -> Result<Summary> {
    let pack = catalog
        .pack(name)
        .ok_or_else(|| unknown_bundle(name, catalog))?;
    let lock = acquire_user_lock(home)?;
    let _ = &lock;
    let mut manifest = load_manifest(home)?;
    if manifest["schema_version"].as_u64() != Some(SCHEMA_VERSION) {
        return Err(format!(
            "user-assets: manifest schema_version unknown — manifest-driven removal refused (read {}, know {SCHEMA_VERSION})",
            manifest["schema_version"]
        ));
    }
    let mut summary = Summary::new(format!("bundle remove {name}"));
    reconcile_journal(home, &mut manifest, &mut summary)?;

    let bundles = bundles_of(&manifest);
    let mut remaining: Vec<String> = bundles.iter().filter(|b| b != &name).cloned().collect();
    remaining.sort();

    let mut preserved_names = catalog.base_names.clone();
    for other in &remaining {
        if let Some(other_pack) = catalog.pack(other) {
            preserved_names.extend(other_pack.entries.iter().map(|e| e.name.clone()));
        }
    }
    let mut deferred_names: Vec<String> = Vec::new();
    for other in &remaining {
        if let Some(other_pack) = catalog.pack(other) {
            for entry in &other_pack.entries {
                deferred_names.extend(entry.depends_skills.iter().cloned());
            }
        }
    }

    for entry in &pack.entries {
        if preserved_names.contains(&entry.name) {
            summary.lines.push(format!(
                "  kept (shared with L0 or another opted-in bundle): {}",
                entry.name
            ));
            continue;
        }
        if deferred_names.contains(&entry.name) {
            summary.lines.push(format!(
                "  kept (declared dependency of a preserved asset — re-evaluated at the next removal/update): {}",
                entry.name
            ));
            continue;
        }
        let folder = entry.path.rsplit('/').next().unwrap_or(&entry.name);
        let is_file = catalog.entry_is_file(&entry.path);
        let (exact, prefix) = if is_file {
            let under = entry_under_root(&entry.path);
            (
                Some(format!("{}/{}", entry.root, under)),
                format!("{}/{}/", entry.root, under),
            )
        } else {
            (None, format!("{}/{}/", entry.root, folder))
        };
        let root_prefix = format!("{}/", entry.root);
        let keys: Vec<String> = manifest["files"]
            .as_object()
            .map(|files| {
                files
                    .keys()
                    .filter(|key| {
                        exact.as_deref() == Some(key.as_str()) || key.starts_with(&prefix)
                    })
                    .cloned()
                    .collect()
            })
            .unwrap_or_default();
        for key in keys {
            // The path under the deploy root is everything after the root slug — for a file
            // entry the whole rest, for a folder entry the folder with the file.
            let under_root = key
                .strip_prefix(&root_prefix)
                .expect("a manifest key starts with its root slug")
                .to_string();
            let target = deploy_root(home, &entry.root)?.join(&under_root);
            match std::fs::remove_file(&target) {
                Ok(()) => {}
                Err(e) if e.kind() == std::io::ErrorKind::NotFound => {}
                Err(e) => return Err(format!("user-assets: remove {}: {e}", target.display())),
            }
            if let Some(files) = manifest.get_mut("files").and_then(Value::as_object_mut) {
                files.remove(&key);
            }
            summary.removed += 1;
        }
    }

    set_bundles(&mut manifest, remaining);
    save_manifest(home, &manifest)?;
    clear_journal(home)?;
    Ok(summary)
}

/// `moai bundle list`: one Go-format line per pack in ascending name order.
pub fn list() -> Result<Value> {
    let catalog = catalog()?;
    list_with(&catalog)
}

fn list_with(catalog: &Catalog) -> Result<Value> {
    let mut lines = Vec::new();
    for (name, pack) in &catalog.packs {
        let skills = pack.entries.iter().filter(|e| e.kind == "skill").count();
        let agents = pack.entries.iter().filter(|e| e.kind == "agent").count();
        lines.push(format!("{name:<10} {skills:>3} skills {agents:>3} agents"));
    }
    Ok(json!({
        "packs": catalog.packs.iter().map(|(n, _)| n).collect::<Vec<_>>(),
        "lines": lines,
        "text": lines.join("\n"),
    }))
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::path::PathBuf;

    fn entry(name: &str, path: &str, deps: &[&str]) -> Entry {
        Entry {
            name: name.into(),
            kind: "skill".into(),
            root: "claude-skills".into(),
            path: path.into(),
            depends_skills: deps.iter().map(|d| d.to_string()).collect(),
        }
    }

    fn pack(entries: Vec<Entry>) -> Pack {
        Pack { entries }
    }

    fn synthetic() -> Catalog {
        Catalog {
            base_names: vec!["moai".into()],
            packs: vec![
                (
                    "pack-a".into(),
                    pack(vec![entry(
                        "moai-domain-backend",
                        ".claude/skills/moai-domain-backend",
                        &[],
                    )]),
                ),
                (
                    "pack-b".into(),
                    pack(vec![
                        entry(
                            "moai-domain-backend",
                            ".claude/skills/moai-domain-backend",
                            &[],
                        ),
                        entry(
                            "moai-domain-database",
                            ".claude/skills/moai-domain-database",
                            &["moai-domain-backend"],
                        ),
                    ]),
                ),
            ],
        }
    }

    fn tempdir() -> PathBuf {
        tempfile::Builder::new()
            .prefix("user-assets-unit-")
            .tempdir()
            .expect("temp dir")
            .keep()
    }

    fn file_present(home: &Path, name: &str) -> bool {
        home.join(".claude/skills")
            .join(name)
            .join("SKILL.md")
            .exists()
    }

    #[test]
    fn removing_one_pack_keeps_the_entry_another_selection_still_names() {
        let catalog = synthetic();
        let home = tempdir();
        add_with(&home, &catalog, "pack-a").expect("add pack-a");
        add_with(&home, &catalog, "pack-b").expect("add pack-b");
        let summary = remove_with(&home, &catalog, "pack-b").expect("remove pack-b");
        assert!(
            summary.lines.iter().any(|l| l
                .contains("kept (shared with L0 or another opted-in bundle): moai-domain-backend")),
            "the shared entry survives with the Go line: {:?}",
            summary.lines
        );
        assert_eq!(summary.removed, 1, "only the unshared entry is removed");
        assert!(
            file_present(&home, "moai-domain-backend"),
            "the shared entry's file stays"
        );
        assert!(
            !file_present(&home, "moai-domain-database"),
            "the unshared entry's file is gone"
        );
        let manifest = load_manifest(&home).expect("manifest");
        assert_eq!(bundles_of(&manifest), vec!["pack-a".to_string()]);
    }

    #[test]
    fn removing_a_pack_defers_an_entry_a_preserved_asset_declares() {
        let mut catalog = synthetic();
        // pack-a holds only the database entry, which declares the backend entry; pack-b
        // holds only the backend entry, so nothing is shared and only the dependency rule
        // can keep it.
        for (name, pack) in catalog.packs.iter_mut() {
            match name.as_str() {
                "pack-a" => {
                    pack.entries = vec![entry(
                        "moai-domain-database",
                        ".claude/skills/moai-domain-database",
                        &["moai-domain-backend"],
                    )];
                }
                "pack-b" => {
                    pack.entries = vec![entry(
                        "moai-domain-backend",
                        ".claude/skills/moai-domain-backend",
                        &[],
                    )];
                }
                _ => {}
            }
        }
        let home = tempdir();
        add_with(&home, &catalog, "pack-a").expect("add pack-a");
        add_with(&home, &catalog, "pack-b").expect("add pack-b");
        let summary = remove_with(&home, &catalog, "pack-b").expect("remove pack-b");
        assert!(
            summary
                .lines
                .iter()
                .any(|l| l.contains("kept (declared dependency of a preserved asset")),
            "the backend entry is deferred for the preserved database entry: {:?}",
            summary.lines
        );
        assert_eq!(
            summary.removed, 0,
            "nothing is deleted while it is deferred"
        );
        assert!(
            file_present(&home, "moai-domain-backend"),
            "the deferred entry's file stays"
        );
    }

    #[test]
    fn a_journal_of_an_interrupted_run_reconciles_to_an_install() {
        let catalog = synthetic();
        let home = tempdir();
        let key = ["claude-skills", "moai-domain-backend", "SKILL.md"].join("/");
        let bytes = embedded_file_by_key(&key).expect("embedded bytes");
        let journal = json!({
            "schema_version": SCHEMA_VERSION,
            "bundles_selection": ["pack-a"],
            "entries": [{
                "path": key,
                "expected_sha256": sha256_hex(bytes),
                "bundle": "pack-a",
                "moai_version": "test",
                "installed_at": now_rfc3339(),
                "complete": false,
            }],
        });
        std::fs::create_dir_all(home.join(".moai")).unwrap();
        std::fs::write(
            journal_path(&home),
            serde_json::to_string(&journal).unwrap(),
        )
        .unwrap();
        let summary = add_with(&home, &catalog, "pack-a").expect("add pack-a");
        assert!(
            summary.installed >= 1,
            "the journal's absent target installs: {:?}",
            summary.go_text()
        );
        assert!(
            file_present(&home, "moai-domain-backend"),
            "reconciled install lands"
        );
        assert!(
            !journal_path(&home).exists(),
            "a successful run clears the journal"
        );
    }
}
