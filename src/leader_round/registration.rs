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

use std::path::Path;

use serde_json::{json, Value};

use crate::lane_registry::{valid_instance, DEFAULT_LEASE_SECS};

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

/// Every entry of mode `leader` in the clone's registry, ended leases included, sorted by name.
/// Both file forms are read (the file named by the lane name and the file named by the host
/// session id), through the lane registry's own listing, so an entry that cannot be read is an
/// error naming the file, never "absent". A missing registry folder has none. Reads only;
/// nothing is created.
pub fn leader_entries(project: &Path) -> Result<Vec<LeaderEntry>, String> {
    Ok(crate::lane_registry::session_entries(project)?
        .into_iter()
        .filter(|e| e.mode.as_deref() == Some("leader"))
        .map(|e| LeaderEntry {
            name: e.name,
            instance: e.instance,
            lease_until: e.lease_until,
            session: e.session,
        })
        .collect())
}

/// The live leader entry that registers `session` and has no launcher lock file (SPEC REQ-6):
/// a window opened without the launcher has no `.live` file, so its lease is the only sign it
/// is there. A `.live` file that exists is the launcher's word and is never overridden here
/// (a free lock means the launcher exited). Reads only; `None` for anything that cannot be
/// read or does not match.
pub fn leader_alive_by_lease(project: &Path, session: &str, now: i64) -> Option<LeaderEntry> {
    let trace = crate::lane_registry::launcher_trace_of_session(project, session).ok()?;
    if trace.live_file {
        return None;
    }
    leader_entries(project).ok()?.into_iter().find(|e| {
        e.name == trace.name && e.session.as_deref() == Some(session) && e.lease_until > now
    })
}

/// SPEC REQ-2: a leader registration names the host session of the window that registers, so
/// one without a host session id is refused before the registry folder is touched.
pub(crate) fn require_session(own: Option<&str>) -> Result<(), String> {
    if own.is_some() {
        return Ok(());
    }
    Err(
        "leader registration refused: this session has no host session id to record; run it in the window that opens the coordination (CLAUDE_CODE_SESSION_ID or CODEX_THREAD_ID)"
            .into(),
    )
}

/// Whether the entry the name holds under a lease that still runs is of mode `leader`. The
/// rules below apply to such an entry whatever mode the caller asks for, because the entry,
/// not the request, is what another window reads (SPEC REQ-4).
pub(crate) fn holds_live_leader(
    current: Option<&crate::lane_registry::SessionEntry>,
    previous: &str,
) -> bool {
    previous == "same_instance" && current.is_some_and(|c| c.mode.as_deref() == Some("leader"))
}

/// SPEC REQ-3 and REQ-4, judged under the registry lock after the registration's own conflict
/// check passed. `previous` is that check's answer (`none`, `expired`, `same_instance` or
/// `other_session`), `current` the entry the name holds now and `requested` whether the caller
/// asked for mode `leader`.
///
/// - A renewal that touches mode `leader` — the entry is a live leader entry, or the caller
///   asks for the mode — comes from the session the entry names, and that session must be
///   known; an entry that names none is filled in by the caller.
/// - A registration that asks for mode `leader`, and is not the renewal of a live leader entry,
///   is refused while an entry of another name is mode `leader` under a lease that runs past
///   `now`: the refusal names that entry and its lease end, never its instance. A renewal of a
///   lane entry into mode `leader` is such a registration.
/// - Any other registration is not touched by these rules.
///
/// An entry the scan cannot read fails the call with its message: whether a second window is
/// live cannot be judged, so nothing is written.
pub(crate) fn check_leader_rules(
    project: &Path,
    name: &str,
    current: Option<&crate::lane_registry::SessionEntry>,
    previous: &str,
    own: Option<&str>,
    now: i64,
    requested: bool,
) -> Result<(), String> {
    let live_leader = holds_live_leader(current, previous);
    if previous == "same_instance" && (live_leader || requested) {
        require_session(own)?;
        let held = current.and_then(|c| c.session.as_deref());
        if held.is_some() && held != own {
            return Err(
                "leader renewal refused: the entry names another host session; renew from the window that registered it"
                    .into(),
            );
        }
    }
    if live_leader || !requested {
        return Ok(());
    }
    match leader_entries(project)?
        .into_iter()
        .find(|e| e.name != name && e.lease_until > now)
    {
        Some(other) => Err(format!(
            "conflict in session registry: the leader entry `{}` is live until {}; one coordination window registers as leader at a time, so wait for that lease to end or have it released",
            other.name,
            crate::forge::policy::format_utc(other.lease_until)
        )),
        None => Ok(()),
    }
}

/// The note every `registration` part carries: the renewal is the leader's own act, and the
/// lease is no proof that the window lives.
const NOTE: &str = "the leader registration lease is the window's own to renew every round: run `moai leader-round --renew-instance <instance> …` (the round renews, then lists), or `moai session-register --name <name> --mode leader --instance <instance> <project>`; an ended lease means a round was missed, and a lease that still runs is not a proof that the window lives";

/// The `registration` part of the round (SPEC REQ-7 to REQ-9, REQ-12): every leader entry with
/// its lease, its session id and either the seconds left or the seconds since the lease ended;
/// the status is `absent`, `duplicate` (two or more leases still run), `expired`, `no_session`,
/// `ok` or `error` (the message alone), in that order of precedence.
pub fn registration(project: &Path, now: i64) -> Value {
    let entries = match leader_entries(project) {
        Ok(e) => e,
        Err(message) => return json!({"status": "error", "message": message}),
    };
    let ended = entries.iter().any(|e| e.lease_until <= now);
    let unnamed = entries.iter().any(|e| e.session.is_none());
    let live = entries.iter().filter(|e| e.lease_until > now).count();
    let status = if entries.is_empty() {
        "absent"
    } else if live > 1 {
        "duplicate"
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
