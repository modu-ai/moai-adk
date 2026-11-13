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

/// A name that holds secrets by convention; never offered for copying.
pub fn is_secret(name: &str) -> bool {
    name.starts_with(".env") || name.starts_with("credentials") || name == "secrets"
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
        match session_lock_held(&moai) {
            Some(false) => {}
            Some(true) => blocks.push("go_session_live".into()),
            None => blocks.push("session_lock_unknown".into()),
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
        for s in ["config", "state", "secrets-notes", "envy"] {
            assert!(!is_secret(s), "{s}");
        }
    }
}
