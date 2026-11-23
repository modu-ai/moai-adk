//! The coordination window's registration (SPEC-20261123-leader-registration-lease).
//!
//! A window that runs `/moai leader` registers in the clone's session registry with mode
//! `leader` (`lane_registry::register`). Its entry is the one place a lane can read that
//! window's host session id and how long its lease still runs. The lane registry's own readers
//! answer only for leases that still run, so an ended lease reads as `free` and nobody learns
//! that a leader stopped renewing (the design report's 3.1 finding: 38 hours). This module
//! reads every leader entry, ended ones included, and builds the `registration` part the
//! round shares; it also renews a leader entry from the window that registered it.
//!
//! Reads only, except [`renew`], which writes through `lane_registry::register`. A link, a
//! non-regular file, an oversized entry or one that is not a well-formed entry is an error
//! naming the file, never "absent".

use std::fs;
use std::path::{Path, PathBuf};

use serde_json::{json, Value};

use crate::lane_registry::{registry_dir, valid_instance, valid_name, DEFAULT_LEASE_SECS};

/// Largest entry file read; the lane registry's own bound.
const MAX_ENTRY_BYTES: u64 = 4_096;

/// One registry entry of mode `leader`, as stored.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct LeaderEntry {
    pub name: String,
    pub instance: String,
    pub lease_until: i64,
    /// The host session id of the window that registered; `None` for an entry written before
    /// the id was recorded.
    pub session: Option<String>,
}

/// This session's own host id, from the environment only (`session-current`).
fn own_session_id() -> Option<String> {
    crate::session_current::resolve(|name| std::env::var(name).ok())["source_session_id"]
        .as_str()
        .map(str::to_string)
}

fn broken(file: &Path) -> String {
    format!(
        "session registry entry {} is invalid; remove it after checking that no session uses that name",
        file.display()
    )
}

/// The entry at `file` for `name` when it is of mode `leader`; `None` for another mode.
fn read_leader(file: &Path, name: &str) -> Result<Option<LeaderEntry>, String> {
    match fs::symlink_metadata(file) {
        Err(_) => return Err(broken(file)),
        Ok(m) if !m.file_type().is_file() || m.len() > MAX_ENTRY_BYTES => return Err(broken(file)),
        Ok(_) => {}
    }
    let text = fs::read_to_string(file).map_err(|_| broken(file))?;
    let v: Value = serde_json::from_str(&text).map_err(|_| broken(file))?;
    let instance = v["instance"].as_str().filter(|i| valid_instance(i));
    match (
        v["schema_version"].as_u64(),
        v["name"].as_str(),
        instance,
        v["lease_until"].as_i64(),
    ) {
        (Some(1), Some(n), Some(i), Some(lease)) if n == name => {
            if v["mode"].as_str() != Some("leader") {
                return Ok(None);
            }
            Ok(Some(LeaderEntry {
                name: n.to_string(),
                instance: i.to_string(),
                lease_until: lease,
                session: v["session"].as_str().map(str::to_string),
            }))
        }
        _ => Err(broken(file)),
    }
}

/// Every entry of mode `leader` in the clone's registry, ended leases included, sorted by name.
/// A missing registry folder has none. Reads only; nothing is created.
pub fn leader_entries(project: &Path) -> Result<Vec<LeaderEntry>, String> {
    let dir = registry_dir(project)?;
    match fs::symlink_metadata(&dir) {
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(Vec::new()),
        Err(e) => {
            return Err(format!(
                "session registry {} cannot be read: {e}",
                dir.display()
            ))
        }
        Ok(m) if !m.is_dir() => {
            return Err(format!(
                "session registry path {} is a symlink or not a folder; refused",
                dir.display()
            ))
        }
        Ok(_) => {}
    }
    let read = fs::read_dir(&dir)
        .map_err(|e| format!("session registry {} cannot be read: {e}", dir.display()))?;
    let mut files: Vec<PathBuf> = read
        .flatten()
        .map(|e| e.path())
        .filter(|p| p.extension().and_then(|x| x.to_str()) == Some("json"))
        .collect();
    files.sort();
    let mut out = Vec::new();
    for path in files {
        let Some(stem) = path.file_stem().and_then(|s| s.to_str()) else {
            continue;
        };
        if !valid_name(stem) {
            continue;
        }
        if let Some(entry) = read_leader(&path, stem)? {
            out.push(entry);
        }
    }
    Ok(out)
}

/// The note every `registration` part carries: the renewal is the leader's own act, and the
/// lease is no proof that the window lives.
const NOTE: &str = "the leader registration lease is the window's own to renew every round: run `moai leader-round --renew-instance <instance> …` (the round renews, then lists), or `moai session-register --name <name> --mode leader --instance <instance> <project>`; an ended lease means a round was missed, and a lease that still runs is not a proof that the window lives";

/// The `registration` part of the round (SPEC REQ-7 to REQ-9, REQ-12): every leader entry with
/// its lease, its session id and either the seconds left or the seconds since the lease ended;
/// the status is `absent`, `ok`, `no_session`, `expired` or `error` (the message alone).
pub fn registration(project: &Path, now: i64) -> Value {
    let entries = match leader_entries(project) {
        Ok(e) => e,
        Err(message) => return json!({"status": "error", "message": message}),
    };
    let ended = entries.iter().any(|e| e.lease_until <= now);
    let unnamed = entries.iter().any(|e| e.session.is_none());
    let status = if entries.is_empty() {
        "absent"
    } else if ended {
        "expired"
    } else if unnamed {
        "no_session"
    } else {
        "ok"
    };
    let leaders: Vec<Value> = entries
        .iter()
        .map(|e| {
            let mut v = json!({
                "name": e.name,
                "lease_until": crate::forge::policy::format_utc(e.lease_until),
            });
            if let Some(s) = &e.session {
                v["session"] = json!(s);
            }
            if e.lease_until <= now {
                v["expired_for_secs"] = json!(now - e.lease_until);
            } else {
                v["remaining_secs"] = json!(e.lease_until - now);
            }
            v
        })
        .collect();
    json!({"status": status, "leaders": leaders, "note": NOTE})
}

/// Renew the leader entry that `instance` holds, for the default lease, from this window
/// (SPEC REQ-10, REQ-14). Refused with the reason when the instance is malformed or holds no
/// leader entry, when this session has no host id, or when the entry names another session.
/// The entry keeps its instance and mode; an ended lease is renewed too.
pub fn renew(project: &Path, instance: &str, now: i64) -> Result<Value, String> {
    if !valid_instance(instance) {
        return Err(
            "instance must be 32 lowercase hex digits (as returned by the leader registration)"
                .into(),
        );
    }
    let Some(own) = own_session_id() else {
        return Err("renewal refused: this session has no host session id".into());
    };
    let entries = leader_entries(project)?;
    let Some(entry) = entries.iter().find(|e| e.instance == instance) else {
        return Err("renewal refused: no leader entry holds this instance".into());
    };
    if entry.session.as_deref().is_some_and(|s| s != own) {
        return Err(
            "renewal refused: the entry names another host session; renew from the window that registered it"
                .into(),
        );
    }
    let out = crate::lane_registry::register(
        project,
        &entry.name,
        Some(instance),
        DEFAULT_LEASE_SECS,
        Some("leader"),
        now,
    )?;
    Ok(json!({
        "status": "renewed",
        "name": entry.name,
        "lease_until": out["lease_until"],
        "previous": out["previous"],
    }))
}
