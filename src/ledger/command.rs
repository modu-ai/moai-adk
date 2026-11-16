//! The three commands of the ledger slice P1 (Issue #3068): `ledger-register-plan`,
//! `ledger-register-apply` and `ledger-status`. The `src/ops/ledger_*.rs` files only parse
//! arguments and call these.
//!
//! The project id is the one the project's profile already holds (`project_id::read`); the HOME
//! is an argument (default `$HOME`), so tests pass a temporary folder. The plan and the apply
//! follow the grammar of the other plan/apply pairs: the apply recomputes the plan and acts only
//! when its digest equals the one the caller was shown and a permission reference is given.

use super::{location, open, LedgerError};
use serde_json::{json, Value};
use std::path::{Path, PathBuf};

type Result<T> = std::result::Result<T, String>;

/// The HOME to use: `--home`, else `$HOME` (an empty value counts as unset).
pub fn resolve_home(home: Option<PathBuf>) -> Result<PathBuf> {
    match home {
        Some(h) => Ok(h),
        None => std::env::var_os("HOME")
            .filter(|h| !h.is_empty())
            .map(PathBuf::from)
            .ok_or_else(|| "HOME is not set; pass --home".to_string()),
    }
}

/// The profile's `project_id` of the project at `project`.
pub fn project_id_of(project: &Path) -> Result<String> {
    let real = std::fs::canonicalize(project)
        .map_err(|_| crate::contract::PROJECT_PATH_UNAVAILABLE.to_string())?;
    crate::project_id::read(&real)?
        .ok_or_else(|| "project profile missing: run the project setup first".to_string())
}

/// `ledger-register-plan`: what registering would do. Writes nothing.
pub fn register_plan(project: &Path, home: &Path) -> Result<Value> {
    let id = project_id_of(project)?;
    plan_for(home, &id).map_err(|e| e.to_string())
}

fn plan_for(home: &Path, id: &str) -> std::result::Result<Value, LedgerError> {
    let path = location::ledger_path(home, id)?;
    let dir = location::project_dir(home, id)?;
    let (state, action, problem): (&str, &str, Option<LedgerError>) =
        match std::fs::symlink_metadata(&path) {
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => {
                match open::check_home_ready(home) {
                    Ok(()) => ("absent", "register", None),
                    Err(e) => ("absent", "blocked", Some(e)),
                }
            }
            // A path that cannot be examined at all (for example `.moai` is a file, ENOTDIR) in
            // a HOME that is not ready is a blocked plan, not an `io` failure.
            Err(e) => match open::check_home_ready(home) {
                Err(not_ready) => ("absent", "blocked", Some(not_ready)),
                Ok(()) => return Err(LedgerError::Io(e.to_string())),
            },
            Ok(_) => match open::open_read_only(home, id) {
                Ok(_) => ("registered", "none", None),
                Err(e) => ("unavailable", "none", Some(e)),
            },
        };
    let would_create: Vec<String> = if action == "register" {
        let mut v = Vec::new();
        if !dir.exists() {
            v.push(dir.display().to_string());
        }
        v.push(path.display().to_string());
        v
    } else {
        Vec::new()
    };
    let core = json!({
        "project_id": id,
        "ledger_path": path.display().to_string(),
        "state": state,
        "action": action,
        "problem": problem.as_ref().map(LedgerError::code),
        "would_create": would_create,
        "schema_version": super::SCHEMA_VERSION,
    });
    let digest = crate::digest(core.to_string().as_bytes());
    let mut out = core;
    out["message"] = json!(problem.map(|e| e.to_string()));
    out["plan_digest"] = json!(digest);
    out["apply"] = json!(
        "ledger-register-apply with the same project and home, this plan_digest and a permission reference, only after the user approves"
    );
    out["network"] = json!("disabled");
    out["writes"] = json!(false);
    Ok(out)
}

/// `ledger-register-apply`: registers the ledger when the freshly computed plan has the digest
/// the caller holds. An existing valid ledger is left alone (`unchanged`).
pub fn register_apply(
    project: &Path,
    home: &Path,
    plan_digest: &str,
    permission_reference: &str,
) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission_reference) {
        return Err("permission reference required".into());
    }
    let id = project_id_of(project)?;
    let fresh = plan_for(home, &id).map_err(|e| e.to_string())?;
    if fresh["plan_digest"].as_str() != Some(plan_digest) {
        return Err("stale or altered plan: the recomputed digest differs; re-preview required, nothing was written".into());
    }
    let result = open::register(home, &id).map_err(|e| e.to_string())?;
    let path = location::ledger_path(home, &id).map_err(|e| e.to_string())?;
    Ok(json!({
        "project_id": id,
        "ledger_path": path.display().to_string(),
        "state": match result {
            open::Registered::Created => "applied",
            open::Registered::AlreadyExists => "unchanged",
        },
        "created": match result {
            open::Registered::Created => fresh["would_create"].clone(),
            open::Registered::AlreadyExists => json!([]),
        },
        "schema_version": super::SCHEMA_VERSION,
        "plan_digest": plan_digest,
        "network": "disabled",
    }))
}

/// `ledger-status`: read-only report. An absent ledger is a normal answer (`exists: false`),
/// an unusable one is reported with its `problem` code; nothing is created, repaired or
/// written, and the file is opened read-only after its header was read.
pub fn status(project: &Path, home: &Path) -> Result<Value> {
    let id = project_id_of(project)?;
    let path = location::ledger_path(home, &id).map_err(|e| e.to_string())?;
    let mut out = json!({
        "project_id": id,
        "ledger_path": path.display().to_string(),
        "exists": false,
        "state": "absent",
        "schema_version": null,
        "wal": false,
        "cards": null,
        "problem": null,
        "writes": false,
    });
    match open::open_read_only(home, &id) {
        Err(LedgerError::Absent) => {}
        Err(e) => {
            out["exists"] = json!(true);
            out["state"] = json!("unavailable");
            out["wal"] = json!(matches!(e, LedgerError::WalUnexpected));
            out["problem"] = json!(e.code());
            out["message"] = json!(e.to_string());
        }
        Ok(ledger) => {
            out["exists"] = json!(true);
            out["state"] = json!("ready");
            out["schema_version"] = json!(ledger.schema_version());
            out["cards"] = card_counts(&ledger).map_err(|e| e.to_string())?;
        }
    }
    Ok(out)
}

fn card_counts(ledger: &open::ReadLedger) -> std::result::Result<Value, LedgerError> {
    let mut stmt = ledger
        .conn()
        .prepare("SELECT state, count(*) FROM card GROUP BY state ORDER BY state")?;
    let rows = stmt.query_map([], |r| Ok((r.get::<_, String>(0)?, r.get::<_, i64>(1)?)))?;
    let mut by_state = serde_json::Map::new();
    let mut total = 0;
    for row in rows {
        let (state, n) = row?;
        total += n;
        by_state.insert(state, json!(n));
    }
    Ok(json!({"total": total, "by_state": by_state}))
}
