//! Project OMP agents (#2936 R9 S3): the base agents as `<project>/.omp/agents/moai-<role>.md`.
//!
//! OMP reads project agents from `.omp/agents` and does not read `.claude/agents`
//! (github.com/can1357/oh-my-pi `docs/task-agent-discovery.md`, read 2026-11-13), so a project
//! that works with OMP needs its own files. They are the same bytes the HOME installer writes
//! for OMP (`agent_defs::render(def, "omp")`). A project agent shadows a user agent of the same
//! name (`~/.omp/agent/agents`), which the preview says in words.
//!
//! The same rule as `agent_install`, applied to a project: `plan` compares each file
//! three-way — baseline = what our last apply recorded, current = the bytes on disk,
//! new = the rendering — and writes nothing. `apply` re-plans, refuses any other digest and
//! writes only `create` and `update`. A user's edit is never overwritten (`preserved` /
//! `conflict`); only `moai-<role>.md` names the package renders are looked at, and an agent
//! under another name is neither read nor listed. A link or a file in the place of `.omp`,
//! `.omp/agents`, the state folder or an agent file blocks the project instead of being followed.
//! Nothing is ever deleted.

use crate::agent_defs::{package_agents, render};
use crate::apply::{decide, publish_exclusive, publish_replace, read_opt, Decision};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::{
    collections::BTreeMap,
    fs,
    path::{Path, PathBuf},
};

type Result<T> = std::result::Result<T, String>;

const SCHEMA_VERSION: u32 = 1;
/// The agent folder, relative to the project.
const DIR: &str = ".omp/agents";
/// The folders between the project and the agent files, each refused when a link.
const DIRS: &[&str] = &[".omp", ".omp/agents"];
/// What our applies own, relative to the project state root.
const STATE: &str = "manifests/omp-agents.json";
const SHADOWS: &str =
    "a project agent shadows a user agent of the same name in ~/.omp/agent/agents; \
agents under other names are not touched and still load";

/// What our applies own in the folder: file name → sha256 we last wrote.
#[derive(Clone, Debug, Default, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
struct Owned {
    schema_version: u32,
    owned: BTreeMap<String, String>,
}

fn sha(bytes: &[u8]) -> String {
    crate::digest(bytes)
}

/// A path that must be a real folder (or absent): never a link, never a file.
fn plain_dir_or_absent(path: &Path) -> Result<()> {
    match fs::symlink_metadata(path) {
        Ok(m) if m.file_type().is_symlink() => {
            Err(format!("{} is a symlink; refused", path.display()))
        }
        Ok(m) if !m.is_dir() => Err(format!("{} is not a folder", path.display())),
        Ok(_) => Ok(()),
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(()),
        Err(_) => Err(format!("{} is unreadable", path.display())),
    }
}

/// The current bytes of one file: `Ok(None)` when absent; a link or a folder in its place is an
/// error (the project is blocked, never followed).
fn current_file(path: &Path) -> Result<Option<Vec<u8>>> {
    match fs::symlink_metadata(path) {
        Ok(m) if m.file_type().is_symlink() => {
            Err(format!("{} is a symlink; refused", path.display()))
        }
        Ok(m) if !m.is_file() => Err(format!("{} is not a regular file", path.display())),
        Ok(_) => read_opt(path),
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(None),
        Err(_) => Err(format!("{} is unreadable", path.display())),
    }
}

/// Every folder the apply writes through is a real folder or absent: `.omp`, `.omp/agents`,
/// the project state folder and its `manifests`. Plan, apply and every write check it.
fn folders_plain(root: &Path) -> Result<()> {
    let state = crate::state_root::path(root, "");
    let mut dirs: Vec<PathBuf> = DIRS.iter().map(|d| root.join(d)).collect();
    dirs.extend([state.clone(), state.join("manifests")]);
    dirs.iter().try_for_each(|d| plain_dir_or_absent(d))
}

fn state_path(root: &Path) -> PathBuf {
    crate::state_root::path(root, STATE)
}

fn read_owned(root: &Path) -> Result<Owned> {
    match current_file(&state_path(root))? {
        None => Ok(Owned {
            schema_version: SCHEMA_VERSION,
            owned: BTreeMap::new(),
        }),
        Some(bytes) => {
            let owned: Owned = serde_json::from_slice(&bytes)
                .map_err(|_| "omp agent state is unreadable; inspect it first".to_string())?;
            if owned.schema_version != SCHEMA_VERSION {
                return Err("omp agent state has an unknown schema".into());
            }
            Ok(owned)
        }
    }
}

fn write_state(root: &Path, owned: &Owned) -> Result<()> {
    let path = state_path(root);
    let mut bytes = serde_json::to_vec_pretty(owned).map_err(|_| "encode failed")?;
    bytes.push(b'\n');
    match current_file(&path)? {
        Some(b) => publish_replace(&path, Some(&b), &bytes),
        None => publish_exclusive(&path, &bytes),
    }
}

/// The OMP file of every base agent: (file name, text, what OMP cannot express).
fn rendered(package: &Path, agents: &Path) -> Result<Vec<(String, String, Vec<String>)>> {
    package_agents(agents, package)?
        .iter()
        .map(|def| {
            let r = render(def, "omp")?;
            let file = r
                .path
                .strip_prefix("agents/")
                .ok_or("rendered agent path outside agents/")?
                .to_string();
            Ok((file, r.text, r.not_supported))
        })
        .collect()
}

/// One planned file: what the digest binds.
fn file_entry(
    file: &str,
    action: &str,
    base: Option<String>,
    cur: Option<String>,
    new: Option<String>,
) -> Value {
    json!({
        "file": file, "action": action,
        "baseline_sha256": base, "current_sha256": cur, "new_sha256": new,
    })
}

struct Planned {
    root: PathBuf,
    value: Value,
}

fn plan_inner(project: &Path, package: &Path, agents: &Path) -> Result<Planned> {
    let root = fs::canonicalize(project).map_err(|_| "project directory unavailable")?;
    let files = rendered(package, agents)?;
    let mut value = json!({
        "schema_version": SCHEMA_VERSION,
        "host": "omp",
        "project": root.to_str().ok_or("non UTF-8 project path")?,
        "dir": DIR,
        "package_digest": crate::package_digest(package)?,
        "agents_digest": crate::agent_defs::source_digest(agents)?,
        "preflight": "ok",
        "blocked_reason": null,
        "files": [],
        "not_supported": {},
        "shadows_user_agents": SHADOWS,
        "writes": false,
        "plan_digest": "",
    });
    match file_plans(&root, &files) {
        Ok((list, not_supported)) => {
            value["files"] = Value::Array(list);
            value["not_supported"] = json!(not_supported);
        }
        Err(reason) => {
            value["preflight"] = json!("blocked");
            value["blocked_reason"] = json!(reason);
        }
    }
    value["plan_digest"] = json!("");
    let bytes = serde_json::to_vec(&value).map_err(|_| "encode failed")?;
    value["plan_digest"] = json!(sha(&bytes));
    Ok(Planned { root, value })
}

fn file_plans(
    root: &Path,
    files: &[(String, String, Vec<String>)],
) -> Result<(Vec<Value>, BTreeMap<String, Vec<String>>)> {
    folders_plain(root)?;
    let owned = read_owned(root)?;
    let dir = root.join(DIR);
    let mut list = Vec::new();
    let mut not_supported = BTreeMap::new();
    for (file, text, notes) in files {
        if !notes.is_empty() {
            not_supported.insert(file.clone(), notes.clone());
        }
        let current = current_file(&dir.join(file))?.map(|b| sha(&b));
        let new = sha(text.as_bytes());
        let baseline = owned.owned.get(file).cloned();
        let action = match decide(baseline.as_deref(), current.as_deref(), Some(&new)) {
            Decision::KeepCurrent if current.as_deref() == Some(new.as_str()) => "unchanged",
            // The rendering did not change since our last apply: keep the user's edit.
            Decision::KeepCurrent => "preserved",
            Decision::Apply if current.is_none() => "create",
            Decision::Apply => "update",
            Decision::Conflict => "conflict",
        };
        list.push(file_entry(file, action, baseline, current, Some(new)));
    }
    // Files we wrote before that the package no longer renders: never deleted, reported.
    for (file, owned_sha) in &owned.owned {
        if files.iter().any(|(f, _, _)| f == file) {
            continue;
        }
        let current = current_file(&dir.join(file))
            .ok()
            .flatten()
            .map(|b| sha(&b));
        list.push(file_entry(
            file,
            "remove_blocked",
            Some(owned_sha.clone()),
            current,
            None,
        ));
    }
    list.sort_by(|a, b| a["file"].as_str().cmp(&b["file"].as_str()));
    Ok((list, not_supported))
}

/// Read-only preview of `<project>/.omp/agents`; writes nothing. `package` is the skill package
/// the agents' `reads` name, `agents` the source folder outside it (#2721).
pub fn plan(project: &Path, package: &Path, agents: &Path) -> Result<Value> {
    Ok(plan_inner(project, package, agents)?.value)
}

/// Applies a previewed plan. `plan_digest` must equal a fresh plan's digest; only `create` and
/// `update` are written, and a blocked project is refused whole.
pub fn apply(
    project: &Path,
    package: &Path,
    agents: &Path,
    plan_digest: &str,
    permission: &str,
) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err("permission reference required".into());
    }
    let fresh = plan_inner(project, package, agents)?;
    if fresh.value["plan_digest"].as_str() != Some(plan_digest) {
        return Err("stale or altered plan; re-preview required".into());
    }
    if fresh.value["preflight"] != "ok" {
        return Err(fresh.value["blocked_reason"]
            .as_str()
            .unwrap_or("project blocked")
            .to_string());
    }
    let root = fresh.root;
    let texts = rendered(package, agents)?;
    let writes: Vec<&Value> = fresh.value["files"]
        .as_array()
        .into_iter()
        .flatten()
        .filter(|f| f["action"] == "create" || f["action"] == "update")
        .collect();
    if writes.is_empty() {
        return Ok(json!({"state": "unchanged", "plan_digest": plan_digest, "written": []}));
    }
    let dir = root.join(DIR);
    // Checked before creating (create_dir_all would follow a linked parent) and after.
    folders_plain(&root)?;
    fs::create_dir_all(&dir).map_err(|_| format!("{} unavailable", dir.display()))?;
    if let Some(parent) = state_path(&root).parent() {
        fs::create_dir_all(parent).map_err(|_| "omp agent state folder unavailable")?;
    }
    folders_plain(&root)?;
    let mut owned = read_owned(&root)?;
    let mut written = Vec::new();
    for w in writes {
        let file = w["file"].as_str().ok_or("plan file name missing")?;
        let text = &texts
            .iter()
            .find(|(f, _, _)| f == file)
            .ok_or("rendered file missing")?
            .1;
        let target = dir.join(file);
        let before = current_file(&target)?;
        if before.as_deref().map(sha).as_deref() != w["current_sha256"].as_str() {
            return Err(format!(
                "{} changed since the plan; re-preview required",
                target.display()
            ));
        }
        match &before {
            Some(b) => publish_replace(&target, Some(b), text.as_bytes())?,
            None => publish_exclusive(&target, text.as_bytes())?,
        }
        owned.owned.insert(file.to_string(), sha(text.as_bytes()));
        write_state(&root, &owned).map_err(|e| {
            format!(
                "partial apply: wrote [{}]; failed to record ownership: {e}",
                written.join(", ")
            )
        })?;
        written.push(file.to_string());
    }
    Ok(json!({
        "state": "applied",
        "plan_digest": plan_digest,
        "written": written,
        "state_path": crate::state_root::rel(&root, STATE),
    }))
}
