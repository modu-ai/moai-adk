//! The three commands of the ledger slice P1 (Issue #3068): `ledger-register-plan`,
//! `ledger-register-apply` and `ledger-status`. The `src/ops/ledger_*.rs` files only parse
//! arguments and call these.
//!
//! The project id is the one the project's profile already holds (`project_id::read`); the HOME
//! is an argument (default `$HOME`), so tests pass a temporary folder. The plan and the apply
//! follow the grammar of the other plan/apply pairs: the apply recomputes the plan and acts only
//! when its digest equals the one the caller was shown and a permission reference is given.

use super::{folders, location, open, LedgerError};
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
    crate::project_id::read(&real)?.ok_or_else(|| profile_missing_message(&real))
}

/// The error for a project with no profile, ending in the next step. A linked worktree reads its
/// primary checkout's profile, so the step names where the profile is expected.
fn profile_missing_message(real: &Path) -> String {
    format!(
        "project profile missing in {}: run `/moai project` in the project's main checkout to \
         create it (a linked worktree reads that profile), or pass `--project <folder>` for a \
         project that has one",
        real.display()
    )
}

/// `ledger-register-plan`: what registering would do. Writes nothing.
pub fn register_plan(project: &Path, home: &Path) -> Result<Value> {
    let id = project_id_of(project)?;
    plan_for(home, &id).map_err(|e| e.to_string())
}

fn plan_for(home: &Path, id: &str) -> std::result::Result<Value, LedgerError> {
    let path = location::ledger_path(home, id)?;
    let dir = location::project_dir(home, id)?;
    // A link at `.moai`, `projects` or `<id>` blocks the plan before anything is looked up
    // through it; so does a folder another user owns.
    let folders = folders::inspect(home, id);
    let (state, action, problem): (&str, &str, Option<LedgerError>) = match &folders {
        Err(e) => ("absent", "blocked", Some(e.clone())),
        Ok(f) if !f.not_ours.is_empty() => (
            "absent",
            "blocked",
            Some(LedgerError::FolderNotOurs(
                f.not_ours[0].display().to_string(),
            )),
        ),
        Ok(_) => match std::fs::symlink_metadata(&path) {
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
        },
    };
    let would_tighten: Vec<String> = match &folders {
        Ok(f) if action != "blocked" => f
            .would_tighten
            .iter()
            .map(|p| p.display().to_string())
            .collect(),
        _ => Vec::new(),
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
        "would_tighten": would_tighten,
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
    let (result, report) = open::register_with_report(home, &id).map_err(|e| e.to_string())?;
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
        "tightened": report.tightened.iter().map(|p| p.display().to_string()).collect::<Vec<_>>(),
        "schema_version": super::SCHEMA_VERSION,
        "plan_digest": plan_digest,
        "network": "disabled",
    }))
}

/// `ledger-status`: read-only report. An absent ledger is a normal answer (`exists: false`),
/// an unusable one is reported with its `problem` code; nothing is created, repaired or
/// written, and the file is opened read-only after its header was read. `stale_temp` lists the
/// `.creating-<pid>-<n>` files a dead `register` left in the project folder (their process is
/// gone); they are reported only, never removed.
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
        "folder_modes": {"projects": null, "project": null},
        "stale_temp": [],
        "writes": false,
    });
    // A link at `.moai`, `projects` or `<id>` is reported, never followed: nothing below it is
    // read.
    if let Err(e) = folders::refuse_links(home, &id) {
        out["exists"] = json!(true);
        out["state"] = json!("unavailable");
        out["problem"] = json!(e.code());
        out["message"] = json!(e.to_string());
        return Ok(out);
    }
    let (projects_mode, project_mode) = folders::modes(home, &id).map_err(|e| e.to_string())?;
    out["folder_modes"] = json!({"projects": projects_mode, "project": project_mode});
    out["stale_temp"] = json!(super::recover::stale_temp(home, &id).map_err(|e| e.to_string())?);
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
    let mut by_state = serde_json::Map::new();
    let mut total = 0;
    for (state, n) in state_counts(ledger)? {
        total += n;
        by_state.insert(state, json!(n));
    }
    Ok(json!({"total": total, "by_state": by_state}))
}

/// How many cards each state holds, ordered by state name; a state without a card is left out.
/// One `SELECT`, no write. `ledger-status` and the status line (card C-65) read it.
pub fn state_counts(
    ledger: &open::ReadLedger,
) -> std::result::Result<Vec<(String, i64)>, LedgerError> {
    let mut stmt = ledger
        .conn()
        .prepare("SELECT state, count(*) FROM card GROUP BY state ORDER BY state")?;
    let rows = stmt.query_map([], |r| Ok((r.get::<_, String>(0)?, r.get::<_, i64>(1)?)))?;
    let mut counts = Vec::new();
    for row in rows {
        counts.push(row?);
    }
    Ok(counts)
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::fs;

    /// A primary checkout holding a profile and a linked worktree of it with none (the profile
    /// is not tracked in the project), laid out the way git does it. No git process runs.
    fn primary_and_linked(root: &Path, with_profile: bool) -> (PathBuf, PathBuf) {
        let primary = root.join("primary");
        let linked = root.join("linked");
        fs::create_dir_all(primary.join(".git/worktrees/lane")).unwrap();
        fs::create_dir_all(&linked).unwrap();
        fs::write(
            linked.join(".git"),
            format!(
                "gitdir: {}\n",
                primary.join(".git/worktrees/lane").display()
            ),
        )
        .unwrap();
        if with_profile {
            let dir = primary.join(crate::PROJECT_STATE_DIR);
            fs::create_dir_all(&dir).unwrap();
            fs::write(
                dir.join("project.yaml"),
                "{\n  \"schema_version\": 2,\n  \"project_id\": \"canonical\"\n}\n",
            )
            .unwrap();
        }
        (primary, linked)
    }

    #[test]
    fn a_linked_worktree_and_its_primary_have_one_project_id() {
        let t = tempfile::tempdir().unwrap();
        let (primary, linked) = primary_and_linked(t.path(), true);
        assert_eq!(project_id_of(&primary).unwrap(), "canonical");
        assert_eq!(project_id_of(&linked).unwrap(), "canonical");
    }

    #[test]
    fn a_missing_profile_names_the_folder_and_the_next_step() {
        let t = tempfile::tempdir().unwrap();
        let (_primary, linked) = primary_and_linked(t.path(), false);
        let e = project_id_of(&linked).unwrap_err();
        let real = fs::canonicalize(&linked).unwrap();
        assert!(e.starts_with("project profile missing in "), "{e}");
        assert!(e.contains(&real.display().to_string()), "{e}");
        assert!(e.contains("run `/moai project`"), "{e}");
        assert!(e.contains("`--project <folder>`"), "{e}");
        // the contract still reads it as a missing thing
        assert_eq!(
            crate::contract::classify(&e),
            crate::contract::Kind::NotFound,
            "{e}"
        );
    }
}
