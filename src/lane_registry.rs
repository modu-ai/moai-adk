//! Lane name registry (#2592 S-C, leader decision 6856973444 = option (가), design 6850553121).
//!
//! Claim ownership is judged by the session name alone, so two lanes started under one name
//! cannot be told apart, and under decision D1 the second lane's claim line silently renews the
//! first lane's card. The launcher never hands out an `agent-<n>` that is alive (#2971; it
//! reads this folder with [`live_agent_numbers`]), but a session started without it can pick
//! any name. `session-register` closes that gap on one machine and one
//! clone: it keeps `<clone common dir>/.moai-cli/sessions/<name>.json` with a random
//! `instance` and a lease, and refuses a name that another instance holds under a live lease.
//! A lane registers once at start and renews with the same instance at every wake; the
//! MODE-AGENT `session-id` step calls it before claiming and stops the round on a refusal.
//!
//! The folder is the one `-a` numbering already uses (`session::agent_seq_dir`), so every
//! worktree of the clone shares it and git never lists it. Writes go through a temporary file
//! and a rename under an exclusive `File::lock` on `registry.lock`, so two registrations of one
//! name cannot both succeed. Symlinks, non-regular files and oversized or malformed entries are
//! refused, never followed or overwritten.
//!
//! An entry may also carry `mode` (`leader` or `agent`, #2928 decision D1 = (A)), written by
//! `session-register --mode`. The statusline reads it with [`mode_of`] to tag the session.

use std::fs::{self, OpenOptions};
use std::hash::{BuildHasher, Hasher};
use std::io::Write;
use std::path::{Path, PathBuf};

use serde_json::{json, Value};
use sha2::{Digest, Sha256};

/// Lease length when the caller gives none: four hours (#3015 C3; it was one hour). No hook
/// renews a lane's entry, so a lane that sits idle through a long CI or audit wait must not
/// lose its name; a dead lane's name is freed by the leader (`--release --by leader`), and the
/// fallback wake of 1,800 seconds (decision D6) renews it well inside this.
pub const DEFAULT_LEASE_SECS: u64 = 14_400;
/// Shortest accepted lease.
pub const MIN_LEASE_SECS: u64 = 60;
/// Longest accepted lease (one day).
pub const MAX_LEASE_SECS: u64 = 86_400;
/// Longest accepted session name.
pub const MAX_NAME_LEN: usize = 64;
/// Largest entry file read; a real entry is about 150 bytes.
const MAX_ENTRY_BYTES: u64 = 4_096;
/// Test-only: milliseconds to hold the registry lock between reading and writing.
pub const TEST_HOLD_VAR: &str = "MOAI_CLI_TEST_SESSION_REGISTRY_HOLD_MS";

/// A session name that is safe as a file name: 1 to 64 of `A-Z a-z 0-9 . _ -`, starting with
/// a letter or digit (so never `.`, `..` or a hidden file).
pub fn valid_name(name: &str) -> bool {
    let b = name.as_bytes();
    !b.is_empty()
        && b.len() <= MAX_NAME_LEN
        && b[0].is_ascii_alphanumeric()
        && b.iter()
            .all(|c| c.is_ascii_alphanumeric() || matches!(c, b'.' | b'_' | b'-'))
}

/// An instance id: exactly 32 lowercase hex digits.
pub fn valid_instance(instance: &str) -> bool {
    instance.len() == 32
        && instance
            .bytes()
            .all(|c| c.is_ascii_digit() || (b'a'..=b'f').contains(&c))
}

/// A fresh instance id. Uniqueness, not secrecy: the standard library's per-process random
/// hasher keys, the clock and the process id, hashed together.
pub fn new_instance() -> String {
    let mut h = std::collections::hash_map::RandomState::new().build_hasher();
    let nanos = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map_or(0, |d| d.as_nanos());
    h.write_u128(nanos);
    h.write_u32(std::process::id());
    let mut d = Sha256::new();
    d.update(h.finish().to_le_bytes());
    d.update(nanos.to_le_bytes());
    d.update(std::process::id().to_le_bytes());
    d.finalize()[..16]
        .iter()
        .map(|b| format!("{b:02x}"))
        .collect()
}

/// `<clone common dir>/.moai-cli/sessions/`, shared by every worktree of the clone.
pub fn registry_dir(project: &Path) -> Result<PathBuf, String> {
    Ok(registry_paths(project)?.1)
}

/// The clone's git common dir and its registry folder.
fn registry_paths(project: &Path) -> Result<(PathBuf, PathBuf), String> {
    let id = crate::identity::identity(project).map_err(|e| format!("session registry: {e}"))?;
    if id["identity_status"].as_str() != Some("ok") {
        return Err(format!(
            "session registry requires a git repository (identity_status {})",
            id["identity_status"].as_str().unwrap_or("unknown")
        ));
    }
    let common = PathBuf::from(
        id["clone_anchor"]
            .as_str()
            .ok_or("session registry requires a git repository")?,
    );
    let dir = crate::clone_root::path(&common).join("sessions");
    Ok((common, dir))
}

/// [`registry_dir`] for a reader that must not be misled through a link (C-80 audit N1): every
/// path component below the git common dir — the clone state folder and `sessions` — is read
/// with `lstat`, and a symbolic link on any of them is an error. Without this, a linked state
/// folder pointing at an empty registry reads as "no session".
pub fn registry_dir_unlinked(project: &Path) -> Result<PathBuf, String> {
    let (common, dir) = registry_paths(project)?;
    let below = dir
        .strip_prefix(&common)
        .map_err(|_| format!("session registry {} is not below its clone", dir.display()))?;
    let mut at = common.clone();
    for part in below.components() {
        at.push(part);
        refuse_symlink(&at)?;
    }
    Ok(dir)
}

/// Whether `name` has a live entry in the registry folder `dir`: the entry exists and its lease
/// runs past `now` (`lease_until > now`; an entry that ends exactly at `now` is over). A
/// missing entry is `false`; an unreadable, malformed or oversized one is an error (#3015 C2).
pub(crate) fn is_live(dir: &Path, name: &str, now: i64) -> Result<bool, String> {
    if !valid_name(name) {
        return Ok(false);
    }
    let file = dir.join(format!("{name}.json"));
    refuse_symlink(&file)?;
    Ok(matches!(read_entry(&file, name)?, Some((_, lease, _)) if lease > now))
}

/// Whether `name` holds a live lease in the session registry of the clone `project` belongs
/// to, read without a lock or a write; `None` when the registry cannot be read (not a clone, a
/// link, an unreadable entry) — never "no" (#3015 C3b).
pub fn session_live(project: &Path, name: &str, now: i64) -> Option<bool> {
    let dir = registry_dir(project).ok()?;
    match fs::symlink_metadata(&dir) {
        Ok(m) if m.is_dir() => is_live(&dir, name, now).ok(),
        Ok(_) => None,
        // No registry folder yet: no session has ever registered in this clone.
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => Some(false),
        Err(_) => None,
    }
}

/// Takes the registry's exclusive lock (the one `register` holds) in the existing folder `dir`;
/// the lock is held until the returned file is dropped. The folder is not created (#3015 C2).
pub(crate) fn lock_existing(dir: &Path) -> Result<fs::File, String> {
    refuse_symlink(dir)?;
    let lock_path = dir.join("registry.lock");
    refuse_symlink(&lock_path)?;
    if !dir.is_dir() {
        return Err("session registry folder is missing".into());
    }
    OpenOptions::new()
        .read(true)
        .write(true)
        .create(true)
        .truncate(false)
        .open(&lock_path)
        .and_then(|f| f.lock().map(|()| f))
        .map_err(|_| "session registry lock unavailable".to_string())
}

pub(crate) fn refuse_symlink(path: &Path) -> Result<(), String> {
    match fs::symlink_metadata(path) {
        Ok(m) if m.file_type().is_symlink() => Err(format!(
            "session registry path {} is a symlink; refused",
            path.display()
        )),
        _ => Ok(()),
    }
}

/// The current entry — `instance`, `lease_until` and the `mode` text as written (#2928, read
/// through [`accepted_mode`] only) — or `None` when there is none. A non-regular file, an
/// oversized file or one that is not a well-formed entry for `name` is an error, never treated
/// as absent. Unknown keys are ignored, so an entry with `mode` is still schema 1.
type Entry = (String, i64, Option<String>);

fn read_entry(file: &Path, name: &str) -> Result<Option<Entry>, String> {
    let broken = || {
        format!(
            "session registry entry {} is invalid; remove it after checking that no session uses that name",
            file.display()
        )
    };
    match fs::symlink_metadata(file) {
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(None),
        Err(_) => return Err(broken()),
        Ok(m) if !m.file_type().is_file() || m.len() > MAX_ENTRY_BYTES => return Err(broken()),
        Ok(_) => {}
    }
    let text = fs::read_to_string(file).map_err(|_| broken())?;
    let v: Value = serde_json::from_str(&text).map_err(|_| broken())?;
    let instance = v["instance"].as_str().filter(|i| valid_instance(i));
    let lease = v["lease_until"].as_i64();
    match (
        v["schema_version"].as_u64(),
        v["name"].as_str(),
        instance,
        lease,
    ) {
        (Some(1 | 2), Some(n), Some(i), Some(l)) if n == name => Ok(Some((
            i.to_string(),
            l,
            v["mode"].as_str().map(str::to_string),
        ))),
        _ => Err(broken()),
    }
}

/// The liveness anchor of a registry entry (C-135, SPEC-20261120-reclaim-liveness REQ-3): the
/// process that wrote the first registration, for the reclaim gate to probe. A `None` pid
/// means the entry names a session but records no process, which the gate reads as
/// undeterminable.
#[derive(Debug, PartialEq, Eq)]
pub struct LivenessAnchor {
    pub pid: Option<u32>,
    pub process_start: Option<String>,
}

/// This session's own host id, as `session-current` resolves it (the environment only); `None`
/// when the host gives none.
fn own_session_id() -> Option<String> {
    crate::session_current::resolve(|name| std::env::var(name).ok())["source_session_id"]
        .as_str()
        .map(str::to_string)
}

/// The registering process's own start time (`ps`'s `lstart`), read through the one shared
/// probe: `probe_liveness` reports it as `observed_start` when no expected start was given.
fn own_process_start() -> Option<String> {
    let (_, evidence) = crate::handoff::probe_liveness(Some(std::process::id()), None);
    evidence["observed_start"].as_str().map(str::to_string)
}

/// The anchor fields (`session`, `pid`, `process_start`) of the entry at `file`, read
/// leniently: `None` when the file cannot be read as JSON at all; a field is absent when the
/// entry (schema 1, or a host without a session id) does not carry it.
fn anchor_fields(file: &Path) -> Option<(Option<String>, Option<u32>, Option<String>)> {
    let v: Value = serde_json::from_str(&fs::read_to_string(file).ok()?).ok()?;
    Some((
        v["session"].as_str().map(str::to_string),
        v["pid"].as_u64().map(|p| p as u32),
        v["process_start"].as_str().map(str::to_string),
    ))
}

/// REQ-3: the liveness anchor of the registry entry that names `session`, or `None` when no
/// entry names it. Reads the whole folder without a lock and without creating it; an entry
/// that cannot be read is skipped, never an error — the gate decides on what it can read.
/// When several entries name the session (one session under several names), the entry with
/// the longest lease wins.
pub fn anchor_of_session(project: &Path, session: &str) -> Option<LivenessAnchor> {
    if !crate::session_current::well_formed(session) {
        return None;
    }
    let dir = registry_dir(project).ok()?;
    let entries = fs::read_dir(&dir).ok()?;
    let mut best: Option<(i64, LivenessAnchor)> = None;
    for e in entries.flatten() {
        let path = e.path();
        if path.extension().and_then(|x| x.to_str()) != Some("json") {
            continue;
        }
        if fs::symlink_metadata(&path)
            .map(|m| m.file_type().is_symlink())
            .unwrap_or(true)
        {
            continue;
        }
        let Ok(text) = fs::read_to_string(&path) else {
            continue;
        };
        let Ok(v) = serde_json::from_str::<Value>(&text) else {
            continue;
        };
        if v["session"].as_str() != Some(session) {
            continue;
        }
        let anchor = LivenessAnchor {
            pid: v["pid"].as_u64().map(|p| p as u32),
            process_start: v["process_start"].as_str().map(str::to_string),
        };
        let lease = v["lease_until"].as_i64().unwrap_or(0);
        if best.as_ref().is_none_or(|(l, _)| lease >= *l) {
            best = Some((lease, anchor));
        }
    }
    best.map(|(_, anchor)| anchor)
}

/// The session modes an entry may carry (#2928, decision D1 = (A)).
pub const MODES: &[&str] = &["leader", "agent"];

/// An entry's `mode` when it is exactly `leader` or `agent`; any other text — a made-up word,
/// another case, escape bytes — is no mode at all, so it is never shown or trusted.
fn accepted_mode(raw: Option<&str>) -> Option<&'static str> {
    match raw {
        Some("leader") => Some("leader"),
        Some("agent") => Some("agent"),
        _ => None,
    }
}

/// The name text is not echoed: a refused name could carry words that change the error kind.
fn check_name(name: &str) -> Result<(), String> {
    if valid_name(name) {
        Ok(())
    } else {
        Err(format!(
            "session name is invalid: use 1 to {MAX_NAME_LEN} of A-Z a-z 0-9 . _ -, starting with a letter or digit"
        ))
    }
}

/// The registry folder, the entry path for `name`, and the held exclusive lock.
fn open_locked(project: &Path, name: &str) -> Result<(PathBuf, PathBuf, fs::File), String> {
    let dir = registry_dir(project)?;
    refuse_symlink(&dir)?;
    fs::create_dir_all(&dir)
        .map_err(|e| format!("session registry {} write failed: {e}", dir.display()))?;
    let lock_path = dir.join("registry.lock");
    refuse_symlink(&lock_path)?;
    let lock = OpenOptions::new()
        .read(true)
        .write(true)
        .create(true)
        .truncate(false)
        .open(&lock_path)
        .and_then(|f| f.lock().map(|()| f))
        .map_err(|_| format!("session registry lock {} unavailable", lock_path.display()))?;
    let file = dir.join(format!("{name}.json"));
    refuse_symlink(&file)?;
    Ok((dir, file, lock))
}

/// The mode a live entry records for `name` in the clone whose git common dir is `common`
/// (#2928, decision D1 = (A)): `Some("leader")` or `Some("agent")` only when the entry is for
/// exactly this name, its lease runs past `now`, and its `mode` is exactly one of the two.
/// Read-only for the statusline: no lock, no folder created; an unsafe name, a symlinked
/// folder or entry, an oversized or malformed entry — anything unreadable — is `None`.
pub fn mode_of(common: &Path, name: &str, now: i64) -> Option<&'static str> {
    if !valid_name(name) {
        return None;
    }
    let dir = crate::clone_root::path(common).join("sessions");
    refuse_symlink(&dir).ok()?;
    let file = dir.join(format!("{name}.json"));
    match read_entry(&file, name).ok()?? {
        (_, lease, mode) if lease > now => accepted_mode(mode.as_deref()),
        _ => None,
    }
}

/// The numbers `n` of the `agent-<n>.json` entries in the registry folder `dir` whose lease
/// still runs past `now` (`lease_until > now`; an entry that ends exactly at `now` is over),
/// for the launcher's `agent-N` numbering (#2971). Read only: no lock, no write and no folder
/// created; a missing folder has no entries. Files that are not `agent-<n>.json` are ignored,
/// but an entry that is one and is a symlink, not a regular file, oversized or malformed is an
/// error, never "absent" (see [`read_entry`]). The caller has refused a symlinked `dir`.
pub fn live_agent_numbers(dir: &Path, now: i64) -> Result<Vec<u64>, String> {
    let entries = match fs::read_dir(dir) {
        Ok(entries) => entries,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(Vec::new()),
        Err(e) => {
            return Err(format!(
                "session registry {} cannot be read: {e}",
                dir.display()
            ))
        }
    };
    let mut live = Vec::new();
    for entry in entries {
        let entry =
            entry.map_err(|e| format!("session registry {} cannot be read: {e}", dir.display()))?;
        let file_name = entry.file_name();
        let Some(n) = file_name
            .to_str()
            .and_then(|f| crate::session::agent_number_of(f, ".json"))
        else {
            continue;
        };
        let name = format!("agent-{n}");
        if let Some((_, lease, _)) = read_entry(&entry.path(), &name)? {
            if lease > now {
                live.push(n);
            }
        }
    }
    Ok(live)
}

/// How many sessions of this clone are alive at `now` in the registry folder `dir` (#2871 U3c,
/// leader decision 6886405676): every `<name>.json` entry whose lease runs past `now`, and every
/// `agent-<n>` whose launcher holds the `.live` lock, each session counted once. This is the
/// liveness [`live_agent_numbers`] and the launcher already use, for every name instead of the
/// `agent-<n>` ones. Read only: no lock is kept, nothing is created, and a missing folder has no
/// session. A symlinked folder, an entry that is a link, not a regular file, oversized or
/// malformed, and a `.live` file that cannot be probed are errors, never "absent": the caller
/// that must not move state under a running session fails closed on them.
pub fn live_session_count(dir: &Path, now: i64) -> Result<usize, String> {
    refuse_symlink(dir)?;
    let entries = match fs::read_dir(dir) {
        Ok(entries) => entries,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(0),
        Err(e) => {
            return Err(format!(
                "session registry {} cannot be read: {e}",
                dir.display()
            ))
        }
    };
    let mut live = std::collections::BTreeSet::new();
    for entry in entries {
        let entry =
            entry.map_err(|e| format!("session registry {} cannot be read: {e}", dir.display()))?;
        let file_name = entry.file_name();
        let Some(file) = file_name.to_str() else {
            continue;
        };
        if let Some(n) = crate::session::agent_number_of(file, ".live") {
            if crate::session::live_lock_held(&entry.path())? {
                live.insert(format!("agent-{n}"));
            }
        } else if let Some(name) = file.strip_suffix(".json").filter(|n| valid_name(n)) {
            if let Some((_, lease, _)) = read_entry(&entry.path(), name)? {
                if lease > now {
                    live.insert(name.to_string());
                }
            }
        }
    }
    Ok(live.len())
}

/// Who may release on a lane's behalf (#2902 O3a): only the leader, which passes the instance
/// it read with `show`.
pub const RELEASED_BY: &[&str] = &["leader"];

/// The live holder of `name`, read without writing anything (#2902 O3a): `held` with its
/// `instance` and `lease_until`, or `free` when there is no entry or its lease has ended. No
/// lock is taken and no folder is created; a leader passes the instance to `release` with
/// `by = Some("leader")`, which compares it again under the lock.
pub fn show(project: &Path, name: &str, now: i64) -> Result<Value, String> {
    check_name(name)?;
    let dir = registry_dir(project)?;
    refuse_symlink(&dir)?;
    let file = dir.join(format!("{name}.json"));
    refuse_symlink(&file)?;
    Ok(match read_entry(&file, name)? {
        Some((instance, lease, mode)) if lease > now => {
            let mut out = json!({
                "status": "held",
                "name": name,
                "instance": instance,
                "lease_until": crate::forge::policy::format_utc(lease),
            });
            // #2928: only an accepted mode is shown; an entry without one keeps the output.
            if let Some(m) = accepted_mode(mode.as_deref()) {
                out["mode"] = json!(m);
            }
            out
        }
        _ => json!({"status": "free", "name": name}),
    })
}

/// Release `name` held by `instance` (before a lane's session is cleared), so the next
/// session may register it at once. Only the holding instance releases a live entry; an
/// entry whose lease has ended, or none, is `already_free`. `by` (`leader`, #2902 O3a) is
/// recorded as `released_by` on a release; the compare-and-delete is the same.
pub fn release(
    project: &Path,
    name: &str,
    instance: &str,
    by: Option<&str>,
    now: i64,
) -> Result<Value, String> {
    if let Some(b) = by {
        if !RELEASED_BY.contains(&b) {
            return Err("release --by is invalid: only `leader` is accepted".into());
        }
    }
    check_name(name)?;
    if !valid_instance(instance) {
        return Err(
            "instance must be 32 lowercase hex digits (as returned by a previous registration)"
                .into(),
        );
    }
    let (_, file, lock) = open_locked(project, name)?;
    let status = match read_entry(&file, name)? {
        None => "already_free",
        Some((_, lease, _)) if lease <= now => "already_free",
        Some((held, _, _)) if held == instance => {
            fs::remove_file(&file).map_err(|e| {
                format!("session registry entry {} write failed: {e}", file.display())
            })?;
            "released"
        }
        Some((_, lease, _)) => {
            return Err(format!(
                "conflict in session registry: name `{name}` is held by another instance until {}; only that instance releases it",
                crate::forge::policy::format_utc(lease)
            ))
        }
    };
    drop(lock);
    let mut out = json!({"status": status, "name": name});
    if let (Some(b), "released") = (by, status) {
        out["released_by"] = json!(b);
    }
    Ok(out)
}

/// Register `name` for `instance` (a fresh one when `None`) until `now + lease_secs`.
///
/// - No entry, or an entry whose lease has ended (`lease_until <= now`): `registered`, with
///   `previous` `none` or `expired`.
/// - A live entry with the same instance: `renewed` (`previous` `same_instance`).
/// - A live entry with another instance: refused as a conflict naming the lease end; the
///   other instance's id is not shown.
///
/// `mode` (#2928, `leader` or `agent`) is written into the entry and the result only when it is
/// given; every write replaces the whole entry, so a renew records the mode it is given.
pub fn register(
    project: &Path,
    name: &str,
    instance: Option<&str>,
    lease_secs: u64,
    mode: Option<&str>,
    now: i64,
) -> Result<Value, String> {
    if let Some(m) = mode {
        if !MODES.contains(&m) {
            return Err("session mode is invalid: only `leader` or `agent` is accepted".into());
        }
    }
    check_name(name)?;
    if !(MIN_LEASE_SECS..=MAX_LEASE_SECS).contains(&lease_secs) {
        return Err(format!(
            "lease must be {MIN_LEASE_SECS} to {MAX_LEASE_SECS} seconds, got {lease_secs}"
        ));
    }
    if let Some(i) = instance {
        if !valid_instance(i) {
            return Err(
                "instance must be 32 lowercase hex digits (as returned by a previous registration)"
                    .into(),
            );
        }
    }
    let (dir, file, lock) = open_locked(project, name)?;
    let current = read_entry(&file, name)?;
    let previous = match (&current, instance) {
        (None, _) => "none",
        (Some((_, lease, _)), _) if *lease <= now => "expired",
        (Some((held, _, _)), Some(mine)) if held == mine => "same_instance",
        (Some((_, lease, _)), _) => {
            return Err(format!(
                "conflict in session registry: name `{name}` is in use by another instance until {}; pick another name, or wait for that lease to end",
                crate::forge::policy::format_utc(*lease)
            ))
        }
    };
    // Test seam: widens the read-to-write window (at most 2 s) so the concurrency test can
    // show the lock is what keeps two registrations apart. Unset in normal use.
    if let Some(ms) = std::env::var(TEST_HOLD_VAR)
        .ok()
        .and_then(|v| v.parse::<u64>().ok())
    {
        std::thread::sleep(std::time::Duration::from_millis(ms.min(2000)));
    }
    let instance = instance.map_or_else(new_instance, str::to_string);
    let lease_until = now + lease_secs as i64;
    // C-135 (SPEC-20261120-reclaim-liveness REQ-1, REQ-2): the entry also carries the liveness
    // anchor of the registration — the host session id (when the host gives one) and the
    // registering process's id and start time — and a renewal by the same instance keeps the
    // first anchor, because the renewing process is not the session's process.
    let (session, pid, start) = if previous == "same_instance" {
        anchor_fields(&file).unwrap_or((None, None, None))
    } else {
        (
            own_session_id(),
            Some(std::process::id()),
            own_process_start(),
        )
    };
    let mut entry = json!({
        "schema_version": 2,
        "name": name,
        "instance": instance,
        "lease_until": lease_until,
        "registered_at": now,
    });
    if let Some(m) = mode {
        entry["mode"] = json!(m);
    }
    if let Some(s) = session {
        entry["session"] = json!(s);
    }
    if let Some(p) = pid {
        entry["pid"] = json!(p);
    }
    if let Some(s) = start {
        entry["process_start"] = json!(s);
    }
    let tmp = dir.join(format!("{name}.json.tmp-{}", std::process::id()));
    let written = OpenOptions::new()
        .write(true)
        .create_new(true)
        .open(&tmp)
        .and_then(|mut f| {
            f.write_all(format!("{entry}\n").as_bytes())?;
            f.sync_all()
        })
        .and_then(|()| fs::rename(&tmp, &file));
    if let Err(e) = written {
        let _ = fs::remove_file(&tmp);
        return Err(format!(
            "session registry entry {} write failed: {e}",
            file.display()
        ));
    }
    drop(lock);
    let mut out = json!({
        "status": if previous == "same_instance" { "renewed" } else { "registered" },
        "previous": previous,
        "name": name,
        "instance": instance,
        "lease_until": crate::forge::policy::format_utc(lease_until),
        "entry": file.to_string_lossy(),
    });
    if let Some(m) = mode {
        out["mode"] = json!(m);
    }
    Ok(out)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn names_are_file_safe() {
        for ok in [
            "agent-8",
            "a",
            "pr-1-hB",
            "agent_22.x",
            &"a".repeat(MAX_NAME_LEN),
        ] {
            assert!(valid_name(ok), "{ok}");
        }
        let long = "a".repeat(MAX_NAME_LEN + 1);
        for bad in [
            "",
            ".",
            "..",
            ".hidden",
            "-x",
            "a/b",
            "a b",
            "a\\b",
            "ä",
            long.as_str(),
        ] {
            assert!(!valid_name(bad), "{bad:?}");
        }
    }

    /// #2902 O3a: a release names only `leader` as `by`, refused before any path is read.
    #[test]
    fn release_by_accepts_only_leader() {
        let nowhere = Path::new("/nonexistent/moai-o3a");
        for by in ["other", "Leader", ""] {
            let e = release(nowhere, "agent-8", &"0".repeat(32), Some(by), 0).unwrap_err();
            assert!(e.contains("--by is invalid"), "{by:?}: {e}");
        }
        let e = release(nowhere, "agent-8", &"0".repeat(32), Some("leader"), 0).unwrap_err();
        assert!(!e.contains("--by"), "{e}");
    }

    /// #2928 S-2 (decision D1 = (A)): the statusline's leader reading is a live entry for exactly
    /// this name whose `mode` is exactly `leader` or `agent`. Everything else — an ended lease, an
    /// entry naming another lane, any other mode text, no entry, a symlink, an oversized file or an
    /// unsafe name — reads as no mode, and the read creates nothing.
    #[cfg(unix)]
    #[test]
    fn mode_of_reads_only_a_live_exact_leader_entry() {
        let tmp = tempfile::tempdir().unwrap();
        let common = tmp.path();
        let sessions = crate::clone_root::path(common).join("sessions");
        let now = 1_000;
        let entry = |name: &str, lease: i64, mode: Option<&str>| {
            let mut v = json!({
                "schema_version": 1,
                "name": name,
                "instance": "1".repeat(32),
                "lease_until": lease,
                "registered_at": 0,
            });
            if let Some(m) = mode {
                v["mode"] = json!(m);
            }
            v.to_string()
        };
        let put = |file: &str, text: &str| fs::write(sessions.join(file), text).unwrap();

        assert_eq!(mode_of(common, "leader", now), None, "no registry folder");
        assert!(!sessions.exists(), "the read creates no folder");
        fs::create_dir_all(&sessions).unwrap();
        assert_eq!(mode_of(common, "leader", now), None, "no entry");

        put("leader.json", &entry("leader", now + 1, Some("leader")));
        assert_eq!(mode_of(common, "leader", now), Some("leader"), "live");
        put("leader.json", &entry("leader", now, Some("leader")));
        assert_eq!(mode_of(common, "leader", now), None, "the lease end second");
        put("leader.json", &entry("leader", now + 1, Some("agent")));
        assert_eq!(mode_of(common, "leader", now), Some("agent"));
        for other in ["boss", "Leader", "leader ", "", "\u{1b}]0;leader\u{7}"] {
            put("leader.json", &entry("leader", now + 1, Some(other)));
            assert_eq!(mode_of(common, "leader", now), None, "{other:?}");
        }
        put("leader.json", &entry("leader", now + 1, None));
        assert_eq!(mode_of(common, "leader", now), None, "no mode");
        put("leader.json", &entry("agent-3", now + 1, Some("leader")));
        assert_eq!(
            mode_of(common, "leader", now),
            None,
            "the entry names another lane"
        );

        // 4096 bytes are read; 4097 are refused.
        let padded = |len: usize| {
            let mut s = entry("leader", now + 1, Some("leader"));
            s.push_str(&" ".repeat(len - s.len()));
            s
        };
        put("leader.json", &padded(4096));
        assert_eq!(mode_of(common, "leader", now), Some("leader"), "4096 bytes");
        put("leader.json", &padded(4097));
        assert_eq!(mode_of(common, "leader", now), None, "4097 bytes");

        // A symlinked entry or registry folder is never followed.
        let real = tmp.path().join("real.json");
        fs::write(&real, entry("leader", now + 1, Some("leader"))).unwrap();
        fs::remove_file(sessions.join("leader.json")).unwrap();
        std::os::unix::fs::symlink(&real, sessions.join("leader.json")).unwrap();
        assert_eq!(mode_of(common, "leader", now), None, "symlinked entry");
        fs::remove_file(sessions.join("leader.json")).unwrap();
        let elsewhere = tmp.path().join("elsewhere");
        fs::create_dir_all(&elsewhere).unwrap();
        fs::write(
            elsewhere.join("leader.json"),
            entry("leader", now + 1, Some("leader")),
        )
        .unwrap();
        fs::remove_dir(&sessions).unwrap();
        std::os::unix::fs::symlink(&elsewhere, &sessions).unwrap();
        assert_eq!(mode_of(common, "leader", now), None, "symlinked folder");

        for bad in ["", "../leader", ".leader", "a b"] {
            assert_eq!(mode_of(common, bad, now), None, "{bad:?}");
        }
    }

    fn numbered_entry(n: u64, lease: i64) -> String {
        format!(
            "{{\"schema_version\":1,\"name\":\"agent-{n}\",\"instance\":\"{}\",\"lease_until\":{lease}}}",
            "0".repeat(32)
        )
    }

    /// #2971: a lease that ends exactly at `now` is over; one second more is running.
    #[test]
    fn live_agent_numbers_split_at_the_lease_end() {
        let tmp = tempfile::tempdir().unwrap();
        let dir = tmp.path().join("sessions");
        fs::create_dir(&dir).unwrap();
        let end = 1_000;
        fs::write(dir.join("agent-4.json"), numbered_entry(4, end)).unwrap();
        fs::write(dir.join("agent-7.json"), numbered_entry(7, end + 1)).unwrap();
        let at = |now: i64| {
            let mut live = live_agent_numbers(&dir, now).unwrap();
            live.sort_unstable();
            live
        };
        assert_eq!(at(end - 1), [4, 7]);
        assert_eq!(at(end), [7]);
        assert_eq!(at(end + 1), Vec::<u64>::new());
    }

    /// #2971: only canonical `agent-<n>.json` names count; a folder that is missing is empty and
    /// stays missing; nothing is written.
    #[test]
    fn live_agent_numbers_ignore_other_names_and_write_nothing() {
        let tmp = tempfile::tempdir().unwrap();
        let dir = tmp.path().join("sessions");
        assert_eq!(live_agent_numbers(&dir, 0).unwrap(), Vec::<u64>::new());
        assert!(!dir.exists(), "no folder is created");
        fs::create_dir(&dir).unwrap();
        for ignored in [
            "leader.json",
            "agent-.json",
            "agent-x.json",
            "agent-01.json",
            "agent--1.json",
            "agent-4294967296.json",
            "agent-5.json.tmp-1",
            "agent-5.live",
            "agent-5",
        ] {
            fs::write(dir.join(ignored), "not json").unwrap();
        }
        fs::write(dir.join("agent-12.json"), numbered_entry(12, 10)).unwrap();
        let before = fs::read_dir(&dir).unwrap().count();
        assert_eq!(live_agent_numbers(&dir, 0).unwrap(), [12]);
        assert_eq!(fs::read_dir(&dir).unwrap().count(), before);
    }

    /// #2971: an entry that is one but cannot be trusted is an error, never "absent".
    #[test]
    fn live_agent_numbers_refuse_a_broken_entry() {
        for bad in [
            "not json".to_string(),
            "{}".to_string(),
            numbered_entry(9, 10),
            format!("{}{}", numbered_entry(4, 10), " ".repeat(5_000)),
        ] {
            let tmp = tempfile::tempdir().unwrap();
            fs::write(tmp.path().join("agent-4.json"), &bad).unwrap();
            let e = live_agent_numbers(tmp.path(), 0).unwrap_err();
            assert!(e.contains("agent-4.json"), "{e}");
        }
        let tmp = tempfile::tempdir().unwrap();
        fs::create_dir(tmp.path().join("agent-4.json")).unwrap();
        assert!(live_agent_numbers(tmp.path(), 0).is_err(), "directory");
    }

    /// #2971: a symlinked entry is refused too, never followed.
    #[cfg(unix)]
    #[test]
    fn live_agent_numbers_refuse_a_symlinked_entry() {
        let tmp = tempfile::tempdir().unwrap();
        fs::write(tmp.path().join("real"), numbered_entry(4, 10)).unwrap();
        std::os::unix::fs::symlink(tmp.path().join("real"), tmp.path().join("agent-4.json"))
            .unwrap();
        assert!(live_agent_numbers(tmp.path(), 0).is_err(), "symlink");
    }

    fn named_entry(name: &str, lease: i64) -> String {
        numbered_entry(0, lease).replace("agent-0", name)
    }

    /// #2871 U3c: every name counts once, not only `agent-<n>`; a lease that ends exactly at
    /// `now` is over; a missing folder has no session; other files are not sessions.
    #[test]
    fn live_session_count_counts_every_live_name_once() {
        let tmp = tempfile::tempdir().unwrap();
        let dir = tmp.path().join("sessions");
        assert_eq!(live_session_count(&dir, 0), Ok(0), "missing folder");
        fs::create_dir(&dir).unwrap();
        let now = 1_000;
        fs::write(dir.join("leader.json"), named_entry("leader", now + 1)).unwrap();
        fs::write(dir.join("agent-3.json"), named_entry("agent-3", now + 5)).unwrap();
        fs::write(dir.join("agent-4.json"), named_entry("agent-4", now)).unwrap();
        fs::write(dir.join("registry.lock"), "").unwrap();
        fs::write(dir.join("notes.txt"), "x").unwrap();
        fs::write(dir.join(".hidden.json"), "not an entry").unwrap();
        assert_eq!(live_session_count(&dir, now), Ok(2));
        assert_eq!(live_session_count(&dir, now + 5), Ok(0));
        assert_eq!(live_session_count(&dir, now - 1), Ok(3));
    }

    /// #2871 U3c: what cannot be judged is an error, so the install apply fails closed.
    #[cfg(unix)]
    #[test]
    fn live_session_count_fails_on_what_it_cannot_judge() {
        let tmp = tempfile::tempdir().unwrap();
        let dir = tmp.path().join("sessions");
        fs::create_dir(&dir).unwrap();
        fs::write(dir.join("agent-1.json"), "{not json").unwrap();
        assert!(live_session_count(&dir, 0).is_err(), "malformed entry");
        fs::remove_file(dir.join("agent-1.json")).unwrap();
        fs::write(dir.join("real"), named_entry("leader", 10)).unwrap();
        std::os::unix::fs::symlink(dir.join("real"), dir.join("leader.json")).unwrap();
        assert!(live_session_count(&dir, 0).is_err(), "symlinked entry");
        fs::remove_file(dir.join("leader.json")).unwrap();
        let linked = tmp.path().join("linked");
        std::os::unix::fs::symlink(&dir, &linked).unwrap();
        assert!(live_session_count(&linked, 0).is_err(), "symlinked folder");
    }

    /// #2871 U3c (audit of #3041, B1): a launcher's lock on `agent-<n>.live` is a session with no
    /// registry entry at all; a lock nobody holds is dead; and a folder that cannot be listed is
    /// an error, never "no sessions" (the install apply must fail closed).
    #[cfg(unix)]
    #[test]
    fn live_session_count_follows_the_launcher_lock_and_fails_on_an_unlistable_folder() {
        use std::os::unix::fs::PermissionsExt;
        let tmp = tempfile::tempdir().unwrap();
        let dir = tmp.path().join("sessions");
        fs::create_dir(&dir).unwrap();
        let live = fs::File::create(dir.join("agent-5.live")).unwrap();
        live.lock().unwrap();
        assert_eq!(live_session_count(&dir, 0), Ok(1), "a held lock alone");
        // The same session as a registry entry too: one session, not two.
        fs::write(dir.join("agent-5.json"), named_entry("agent-5", 10)).unwrap();
        assert_eq!(live_session_count(&dir, 0), Ok(1), "lock and entry");
        fs::remove_file(dir.join("agent-5.json")).unwrap();
        drop(live);
        // A lock nobody holds is a file nobody has locked. The file that was just released is not
        // read back for this: while other tests spawn processes in the same run, a child that has
        // forked but not yet exec'd holds a copy of the open file for a moment, and the released
        // lock then still looks held (observed on a loaded machine, `Ok(1)` instead of `Ok(0)`).
        fs::remove_file(dir.join("agent-5.live")).unwrap();
        fs::File::create(dir.join("agent-6.live")).unwrap();
        assert_eq!(live_session_count(&dir, 0), Ok(0), "a lock nobody holds");
        // A folder that cannot be listed. A user who can list it anyway (root) cannot make
        // this input, so the check is skipped there.
        fs::set_permissions(&dir, fs::Permissions::from_mode(0o000)).unwrap();
        let denied = fs::read_dir(&dir).is_err();
        let counted = live_session_count(&dir, 0);
        fs::set_permissions(&dir, fs::Permissions::from_mode(0o755)).unwrap();
        if denied {
            let err = counted.expect_err("an unlistable folder is not zero sessions");
            assert!(err.contains("cannot be read"), "{err}");
        }
    }

    #[test]
    fn instances_are_32_lowercase_hex() {
        let fresh = new_instance();
        assert!(valid_instance(&fresh), "{fresh}");
        assert_ne!(fresh, new_instance());
        assert!(valid_instance(&"0".repeat(32)));
        for bad in [
            "0".repeat(31),
            "0".repeat(33),
            "A".repeat(32),
            "g".repeat(32),
        ] {
            assert!(!valid_instance(&bad), "{bad}");
        }
    }
}
