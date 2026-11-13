//! `home-state-plan` (#2723, design 6860356009, user decision #1975 6859326988): whose is
//! `~/.moai`, and what moving the Go generation's `~/.moai` aside to `~/.moai-backup` would do.
//! Reads only; the apply is a later step.
//!
//! The verdict table is the project's (`state_root::classify_with`, #2725 S2a) with the HOME's
//! Go markers: `user-assets.json` (a JSON object with an integer `schema_version` and a `files`
//! object, the Go user-asset manifest), `state/active-sessions.json` and `credentials.yaml`.
//! Only a `go` verdict offers the move; `mixed`, `unknown`, `refuse` and `interrupted` move
//! nothing and say why (leader condition of 6854854683).
//!
//! The move is one rename of `~/.moai` to `~/.moai-backup` beside it (atomic on one file
//! system, whatever the file count). An existing backup is never overwritten: the plan names
//! `~/.moai-backup-<UTC>` instead, and refuses when that is taken too. The move is blocked
//! while a Go session may be live (the `state/active-sessions.json.lock` lock is held, or a
//! process works inside `~/.moai`), while `~/.moai` holds a nested checkout (a git worktree
//! there would lose its registration), and while the folder is too large to check within the
//! entry limit (fail-closed). The leader's answers 6860443004: copying anything back is the
//! user's choice in the preview (default none), and a machine with a live Go session refuses.
//!
//! Secrets (`credentials*`, `.env*`, `secrets`) are listed but never offered for copying: the
//! user moves them by hand. `MOAI_HOME` (the Go root override, `internal/config/envkeys.go`) is
//! reported when it points elsewhere; this plan handles `~/.moai` only. Writes nothing.

use crate::state_root::{classify_with, read_limited, regular_file, MOAI_DIR};
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};

/// The backup folder's name beside `~/.moai` (user decision #1975 6859326988).
pub const BACKUP_NAME: &str = ".moai-backup";
/// The Go user-asset manifest is a list of files with hashes; larger is not one.
const USER_ASSETS_LIMIT: u64 = 16 * 1024 * 1024;
/// Top-level entries that hold a Go marker; never copied into the Rust folder.
const GO_MARKER_ENTRIES: &[&str] = &["user-assets.json", "state"];
/// Top-level entries listed as copy candidates; more are counted, not listed.
const CANDIDATE_LIMIT: usize = 500;

/// The HOME's Go markers, as paths relative to HOME (#2723 design §2).
pub fn home_go_markers(moai: &Path) -> Vec<String> {
    let mut found = Vec::new();
    let assets_ok = read_limited(moai, &["user-assets.json"], USER_ASSETS_LIMIT)
        .and_then(|b| serde_json::from_slice::<Value>(&b).ok())
        .is_some_and(|v| v["schema_version"].is_u64() && v["files"].is_object());
    if assets_ok {
        found.push(format!("{MOAI_DIR}/user-assets.json"));
    }
    if regular_file(moai, &["state", "active-sessions.json"]).is_some() {
        found.push(format!("{MOAI_DIR}/state/active-sessions.json"));
    }
    if regular_file(moai, &["credentials.yaml"]).is_some() {
        found.push(format!("{MOAI_DIR}/credentials.yaml"));
    }
    found
}

/// A name that holds secrets by convention; never offered for copying. Case is ignored: the
/// default macOS file system does not tell `Credentials.yaml` from `credentials.yaml` (#2771
/// audit N1).
pub fn is_secret(name: &str) -> bool {
    let n = name.to_ascii_lowercase();
    n.starts_with(".env") || n.starts_with("credentials") || n == "secrets"
}

/// The Go session registry is a list of rows with a `pid`; a few KiB in practice.
const REGISTRY_LIMIT: u64 = 4 * 1024 * 1024;

/// How many sessions in `state/active-sessions.json` are alive (the Go census counts them
/// the same way, `internal/homestate/runtime_census.go`): the Go tool holds the registry lock
/// only while it rewrites the file, so the lock alone misses a live session (#2771 audit F1).
/// `None` when the file cannot be read as rows with positive `pid`s or a pid cannot be judged.
fn live_sessions(moai: &Path) -> Option<usize> {
    if fs::symlink_metadata(moai.join("state").join("active-sessions.json")).is_err() {
        return Some(0);
    }
    let bytes = read_limited(moai, &["state", "active-sessions.json"], REGISTRY_LIMIT)?;
    let rows: Vec<Value> = serde_json::from_slice(&bytes).ok()?;
    let mut live = 0;
    for row in rows {
        let pid = row["pid"]
            .as_i64()
            .filter(|p| *p > 0 && *p <= i64::from(i32::MAX))?;
        // SAFETY: signal 0 only checks that the process exists and may be signalled.
        let rc = unsafe { libc::kill(pid as libc::pid_t, 0) };
        if rc == 0 {
            live += 1;
            continue;
        }
        match std::io::Error::last_os_error().raw_os_error() {
            Some(libc::EPERM) => live += 1,
            Some(libc::ESRCH) => {}
            _ => return None,
        }
    }
    Some(live)
}

/// `YYYYMMDDTHHMMSSZ` from `YYYY-MM-DDTHH:MM:SSZ`.
fn stamp(now: &str) -> Result<String, String> {
    let ok = now.len() == 20
        && now.ends_with('Z')
        && now.as_bytes()[10] == b'T'
        && crate::forge::policy::parse_utc(now).is_some();
    if !ok {
        return Err(format!("now `{now}` is not YYYY-MM-DDTHH:MM:SSZ"));
    }
    Ok(now.replace(['-', ':'], ""))
}

fn exists(p: &Path) -> bool {
    fs::symlink_metadata(p).is_ok()
}

/// Whether the Go session lock is held by another process; `None` when it cannot be told.
fn session_lock_held(moai: &Path) -> Option<bool> {
    let lock = moai.join("state").join("active-sessions.json.lock");
    match fs::symlink_metadata(&lock) {
        Err(_) => return Some(false),
        Ok(m) if !m.is_file() => return None,
        Ok(_) => {}
    }
    let f = fs::File::open(&lock).ok()?;
    match f.try_lock_shared() {
        Ok(()) => Some(false),
        Err(fs::TryLockError::WouldBlock) => Some(true),
        Err(fs::TryLockError::Error(_)) => None,
    }
}

/// The preview for the HOME at `home`. `moai_home` is the `MOAI_HOME` value, `now` the
/// clock as `YYYY-MM-DDTHH:MM:SSZ` (names the alternative backup).
pub fn plan(home: &Path, moai_home: Option<&str>, now: &str) -> Result<Value, String> {
    if !home.is_absolute() {
        return Err("HOME is not an absolute path".into());
    }
    let stamp = stamp(now)?;
    let moai = home.join(MOAI_DIR);
    let c = classify_with(home, home_go_markers);

    let primary = home.join(BACKUP_NAME);
    let alternative = home.join(format!("{BACKUP_NAME}-{stamp}"));
    let backup: Option<PathBuf> = if !exists(&primary) {
        Some(primary)
    } else if !exists(&alternative) {
        Some(alternative)
    } else {
        None
    };

    let mut blocks: Vec<String> = Vec::new();
    if c.generation == "go" {
        if backup.is_none() {
            blocks.push("backup_name_taken".into());
        }
        let lock = session_lock_held(&moai);
        let registered = live_sessions(&moai);
        if lock == Some(true) || registered.is_some_and(|n| n > 0) {
            blocks.push("go_session_live".into());
        } else if lock.is_none() {
            blocks.push("session_lock_unknown".into());
        } else if registered.is_none() {
            blocks.push("session_registry_unknown".into());
        }
        let canon = fs::canonicalize(&moai).unwrap_or_else(|_| moai.clone());
        match crate::worktree_gc::process_cwds() {
            Some(cwds) => {
                if cwds
                    .into_iter()
                    .map(|p| fs::canonicalize(&p).unwrap_or(p))
                    .any(|p| p.starts_with(&canon))
                {
                    blocks.push("in_use".into());
                }
            }
            None => blocks.push("process_check_unavailable".into()),
        }
        // A process may hold a file inside open from elsewhere (design 6860356009 §3,
        // #2771 audit F2).
        match crate::worktree_gc::open_file_paths() {
            Some(paths) => {
                if paths.iter().any(|p| p.starts_with(&canon)) {
                    blocks.push("file_open_in_use".into());
                }
            }
            None => blocks.push("open_file_check_unavailable".into()),
        }
        match crate::go_state::nested_checkouts_under(
            home,
            MOAI_DIR,
            crate::go_state::unrecorded_limit(),
        ) {
            Some(0) => {}
            Some(n) => blocks.push(format!("nested_checkouts:{n}")),
            None => blocks.push("not_fully_checked".into()),
        }
    }

    let action = match c.generation {
        "go" if blocks.is_empty() => "migrate",
        "none" | "empty" | "rust" => "none",
        _ => "blocked",
    };

    let mut names: Vec<String> = if c.moai_dir == "folder" {
        fs::read_dir(&moai)
            .map(|rd| {
                rd.filter_map(|e| e.ok())
                    .filter_map(|e| e.file_name().into_string().ok())
                    .collect()
            })
            .unwrap_or_default()
    } else {
        Vec::new()
    };
    names.sort();
    let total = names.len();
    let candidates: Vec<Value> = names
        .iter()
        .take(CANDIDATE_LIMIT)
        .map(|n| {
            if is_secret(n) {
                json!({"name": n, "kind": "secret", "selected": false,
                       "note": "holds secrets; move it by hand if you need it"})
            } else if GO_MARKER_ENTRIES.contains(&n.as_str()) {
                // Copied into the Rust folder it would make the verdict `mixed`.
                json!({"name": n, "kind": "generation_marker", "selected": false,
                       "note": "marks the Go generation; it stays in the backup"})
            } else {
                json!({"name": n, "kind": "entry", "selected": false})
            }
        })
        .collect();

    let mut warnings = Vec::new();
    if let Some(m) = moai_home.filter(|m| !m.is_empty()) {
        let mp = Path::new(m);
        if mp != moai {
            warnings.push(format!(
                "MOAI_HOME is set to {m}: the Go generation's live state may be there; this plan \
                 handles {} only",
                moai.display()
            ));
        }
    }
    if total > CANDIDATE_LIMIT {
        warnings.push(format!(
            "{total} entries in {}; the first {CANDIDATE_LIMIT} are listed",
            moai.display()
        ));
    }

    let backup_shown = backup.as_ref().map(|b| b.display().to_string());
    let bound = json!({
        "generation": c.generation,
        "action": action,
        "backup": backup_shown,
        "blocks": blocks,
        "candidates": candidates,
    });
    let plan_digest = crate::digest(serde_json::to_string(&bound).unwrap_or_default().as_bytes());
    Ok(json!({
        "schema_version": 1,
        "home": home.display().to_string(),
        "state": c.to_json(),
        "action": action,
        "backup": backup_shown,
        "blocks": blocks,
        "copy": {"default": "none", "candidates": candidates, "total": total},
        "warnings": warnings,
        "plan_digest": plan_digest,
        "apply": "not yet: the apply is the next step of #2723; on a real HOME it runs only after the user confirms",
        "writes": false,
    }))
}

// ---- apply and revert (#2723) ----
//
// The apply acts only on an identical fresh preview whose action is `migrate`, with a
// permission reference and the user's copy selection. It renames `~/.moai` to the planned
// backup (refusing when that name exists right before), creates the Rust `~/.moai`, copies
// only the selected entries (plain files and folders; a link or special file refuses the
// whole apply before the rename), writes the receipt, and writes the Rust marker last
// (atomically), so a stop in the middle never looks finished. A failure after the rename puts
// the partial Rust folder aside as `~/.moai.partial-<UTC>` and the Go folder back. Nothing is
// deleted. The revert moves the Rust `~/.moai` aside to `~/.moai.rust-<UTC>` and the backup
// named in its receipt back to `~/.moai`.

/// The receipt inside the Rust `~/.moai`.
pub const RECEIPT: &str = "home-state-receipt.json";
/// Entries a copy may hold before it refuses (the copy is a convenience, not a full sync).
const COPY_ENTRY_LIMIT: u64 = 100_000;

/// Plain files and folders only, no links or special files, bounded; `Err` names the problem.
fn check_copyable(path: &Path, seen: &mut u64) -> Result<(), String> {
    *seen += 1;
    if *seen > COPY_ENTRY_LIMIT {
        return Err(format!(
            "the selection holds more than {COPY_ENTRY_LIMIT} entries; copy it by hand"
        ));
    }
    let m = fs::symlink_metadata(path).map_err(|e| format!("{}: {e}", path.display()))?;
    if m.is_dir() {
        for e in fs::read_dir(path).map_err(|e| format!("{}: {e}", path.display()))? {
            let e = e.map_err(|e| format!("{}: {e}", path.display()))?;
            check_copyable(&e.path(), seen)?;
        }
        Ok(())
    } else if m.is_file() {
        Ok(())
    } else {
        Err(format!(
            "{} is a link or a special file; it is never copied",
            path.display()
        ))
    }
}

fn copy_tree(from: &Path, to: &Path) -> Result<(), String> {
    let m = fs::symlink_metadata(from).map_err(|e| format!("{}: {e}", from.display()))?;
    if m.is_dir() {
        fs::create_dir(to).map_err(|e| format!("{}: {e}", to.display()))?;
        for e in fs::read_dir(from).map_err(|e| format!("{}: {e}", from.display()))? {
            let e = e.map_err(|e| format!("{}: {e}", from.display()))?;
            copy_tree(&e.path(), &to.join(e.file_name()))?;
        }
        Ok(())
    } else if m.is_file() {
        fs::copy(from, to)
            .map(|_| ())
            .map_err(|e| format!("{}: {e}", to.display()))
    } else {
        Err(format!("{} is not a plain file or folder", from.display()))
    }
}

fn write_new(path: &Path, bytes: &[u8]) -> Result<(), String> {
    use std::io::Write;
    let tmp = path.with_extension("tmp-write");
    let mut f = fs::OpenOptions::new()
        .write(true)
        .create_new(true)
        .open(&tmp)
        .map_err(|e| format!("{}: {e}", tmp.display()))?;
    f.write_all(bytes)
        .and_then(|()| f.sync_all())
        .map_err(|e| format!("{}: {e}", tmp.display()))?;
    fs::rename(&tmp, path).map_err(|e| format!("{}: {e}", path.display()))
}

/// Move the Go `~/.moai` aside and create the Rust one (see the section comment).
pub fn apply(
    home: &Path,
    moai_home: Option<&str>,
    now: &str,
    plan_digest: &str,
    permission_reference: &str,
    copy: &[String],
) -> Result<Value, String> {
    if !crate::contract::permission_reference_ok(permission_reference) {
        return Err("permission reference required".into());
    }
    let fresh = plan(home, moai_home, now)?;
    if fresh["plan_digest"].as_str() != Some(plan_digest) {
        return Err("stale or altered plan: the recomputed digest differs; re-preview required, nothing was written".into());
    }
    if fresh["action"] != "migrate" {
        return Err(format!(
            "the preview's action is `{}`, not `migrate`; nothing was written",
            fresh["action"].as_str().unwrap_or("?")
        ));
    }
    let moai = home.join(MOAI_DIR);
    let candidates = fresh["copy"]["candidates"]
        .as_array()
        .cloned()
        .unwrap_or_default();
    let mut seen = 0u64;
    let mut picked: Vec<String> = Vec::new();
    for name in copy {
        let c = candidates
            .iter()
            .find(|c| c["name"] == json!(name))
            .ok_or_else(|| {
                format!("`{name}` is not a copy candidate of the preview; nothing was written")
            })?;
        match c["kind"].as_str() {
            Some("entry") => {}
            Some("secret") => return Err(format!(
                "`{name}` holds secrets and is never copied; move it by hand. Nothing was written"
            )),
            _ => {
                return Err(format!(
                    "`{name}` marks the Go generation and stays in the backup; nothing was written"
                ))
            }
        }
        if !picked.contains(name) {
            check_copyable(&moai.join(name), &mut seen)
                .map_err(|e| format!("{e}; nothing was written"))?;
            picked.push(name.clone());
        }
    }
    let backup = PathBuf::from(
        fresh["backup"]
            .as_str()
            .ok_or("no backup name; nothing was written")?,
    );
    if exists(&backup) {
        return Err(format!(
            "{} appeared after the preview; nothing was written",
            backup.display()
        ));
    }
    fs::rename(&moai, &backup).map_err(|e| {
        format!(
            "moving {} aside failed: {e}; nothing was written",
            moai.display()
        )
    })?;

    let stamp = stamp(now)?;
    let finish = || -> Result<(), String> {
        fs::create_dir(&moai).map_err(|e| format!("{}: {e}", moai.display()))?;
        for name in &picked {
            copy_tree(&backup.join(name), &moai.join(name))?;
        }
        let receipt = json!({
            "schema_version": 1,
            "home": home.display().to_string(),
            "backup": backup.display().to_string(),
            "copied": picked,
            "created_at": now,
            "plan_digest": plan_digest,
            "permission_reference": permission_reference,
        });
        write_new(&moai.join(RECEIPT), format!("{receipt:#}\n").as_bytes())?;
        let marker = json!({
            "schema_version": crate::state_root::MARKER_SCHEMA,
            "generation": "moai-cli",
            "created_at": now,
            "go_backup": backup.file_name().map(|n| n.to_string_lossy().into_owned()),
        });
        write_new(
            &moai.join(crate::state_root::RUST_MARKER),
            format!("{marker}\n").as_bytes(),
        )
    };
    if let Err(e) = finish() {
        // Put things back without deleting anything.
        let aside = home.join(format!("{MOAI_DIR}.partial-{stamp}"));
        let mut back = Vec::new();
        if exists(&moai) {
            back.push(match fs::rename(&moai, &aside) {
                Ok(()) => format!("the partial Rust folder is at {}", aside.display()),
                Err(e2) => format!("moving the partial Rust folder aside failed: {e2}"),
            });
        }
        if !exists(&moai) {
            back.push(match fs::rename(&backup, &moai) {
                Ok(()) => format!("{} is back", moai.display()),
                Err(e2) => format!(
                    "putting the Go folder back failed: {e2}; it is at {}",
                    backup.display()
                ),
            });
        }
        return Err(format!("apply failed: {e}; {}", back.join("; ")));
    }
    Ok(json!({
        "schema_version": 1,
        "state": "applied",
        "home": home.display().to_string(),
        "backup": backup.display().to_string(),
        "copied": picked,
        "receipt": moai.join(RECEIPT).display().to_string(),
        "revert": "home-state-revert with the same --home and a permission reference",
        "rerun": "home-state-plan again: the verdict is `rust` and the action `none`",
    }))
}

/// Put the Go `~/.moai` back: the Rust folder named by a valid receipt and marker moves
/// aside to `~/.moai.rust-<UTC>`, the receipt's backup moves to `~/.moai`. Deletes nothing.
pub fn revert(home: &Path, now: &str, permission_reference: &str) -> Result<Value, String> {
    if !crate::contract::permission_reference_ok(permission_reference) {
        return Err("permission reference required".into());
    }
    if !home.is_absolute() {
        return Err("HOME is not an absolute path".into());
    }
    let stamp = stamp(now)?;
    let c = classify_with(home, home_go_markers);
    if c.generation != "rust" {
        return Err(format!(
            "~/.moai is `{}`, not the Rust folder an apply made; nothing was moved",
            c.generation
        ));
    }
    let moai = home.join(MOAI_DIR);
    let receipt: Value = read_limited(&moai, &[RECEIPT], 1024 * 1024)
        .and_then(|b| serde_json::from_slice(&b).ok())
        .ok_or("no readable home-state receipt in ~/.moai; nothing was moved")?;
    let backup = PathBuf::from(
        receipt["backup"]
            .as_str()
            .ok_or("the receipt names no backup; nothing was moved")?,
    );
    if backup.parent() != Some(home)
        || !backup
            .file_name()
            .and_then(|n| n.to_str())
            .is_some_and(|n| n.starts_with(BACKUP_NAME))
    {
        return Err("the receipt's backup is not a `.moai-backup*` folder beside ~/.moai; nothing was moved".into());
    }
    match fs::symlink_metadata(&backup) {
        Ok(m) if m.is_dir() => {}
        _ => {
            return Err(format!(
                "{} is not a folder; nothing was moved",
                backup.display()
            ))
        }
    }
    let aside = home.join(format!("{MOAI_DIR}.rust-{stamp}"));
    if exists(&aside) {
        return Err(format!("{} exists; nothing was moved", aside.display()));
    }
    fs::rename(&moai, &aside)
        .map_err(|e| format!("moving the Rust folder aside failed: {e}; nothing was moved"))?;
    if let Err(e) = fs::rename(&backup, &moai) {
        let undo = fs::rename(&aside, &moai);
        return Err(format!(
            "putting the Go folder back failed: {e}; the Rust folder {}",
            if undo.is_ok() {
                "is back in place"
            } else {
                "stays at its aside name"
            }
        ));
    }
    Ok(json!({
        "schema_version": 1,
        "state": "reverted",
        "home": home.display().to_string(),
        "rust_folder": aside.display().to_string(),
        "restored_from": backup.display().to_string(),
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn stamps_and_secrets() {
        assert_eq!(stamp("2026-11-13T18:05:09Z").unwrap(), "20261113T180509Z");
        for bad in ["2026-11-13 18:05:09Z", "2026-11-13T18:05:09", "x"] {
            assert!(stamp(bad).is_err(), "{bad}");
        }
        for s in [
            ".env.glm",
            ".env",
            "credentials",
            "credentials.yaml.bak",
            "secrets",
        ] {
            assert!(is_secret(s), "{s}");
        }
        for s in ["Credentials.yaml.bak", ".ENV", "SECRETS"] {
            assert!(is_secret(s), "{s}");
        }
        for s in ["config", "state", "secrets-notes", "envy"] {
            assert!(!is_secret(s), "{s}");
        }
    }
}
