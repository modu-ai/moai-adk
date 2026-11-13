//! Base agent install into the user's host folders (#2592 stage 2, S2).
//!
//! `plan` renders the package's agent sources (`agent_defs`) for every host and compares each
//! file three-way — baseline = what our last apply recorded, current = the bytes on disk,
//! new = the rendering — the same rule `update` uses (`apply::decide`). Only files named
//! `moai-<role>.<ext>` that the package renders are looked at; nothing else in the folder is
//! read, listed or touched. `apply` re-plans, refuses any other digest, writes only `create`
//! and `update` through the 01-B publish engine, and records a receipt with the before bytes;
//! `revert` undoes one receipt for files still holding the bytes it wrote. A user's edit is
//! never overwritten (`preserved` / `conflict`), nothing is ever deleted except a file this
//! tool created and the user has not changed since, and a symlinked folder or file blocks the
//! host instead of being followed.

use crate::agent_defs::{package_agents, render, HOSTS};
use crate::apply::{decide, publish_exclusive, publish_replace, read_opt, Decision};
use serde::{Deserialize, Serialize};
use std::{
    collections::BTreeMap,
    fs,
    path::{Path, PathBuf},
};

type Result<T> = std::result::Result<T, String>;
/// One rendered agent file: (file name, text, what the host cannot express).
type RenderedFile = (String, String, Vec<String>);

const SCHEMA_VERSION: u32 = 1;
const STATE_DIR: &str = "agents";

/// The user-level agent folder of a host: Claude Code `~/.claude/agents/`
/// (code.claude.com/docs/en/sub-agents), Codex `~/.codex/agents/` (learn.chatgpt.com
/// agent-configuration/subagents), OMP `~/.omp/agent/agents/` (github.com/can1357/oh-my-pi
/// `docs/task-agent-discovery.md`, read 2026-11-13; #2633).
fn host_dir(home: &Path, host: &str) -> PathBuf {
    match host {
        "claude" => home.join(".claude").join("agents"),
        "codex" => home.join(".codex").join("agents"),
        "omp" => home.join(".omp").join("agent").join("agents"),
        other => unreachable!("unknown agent host {other}"),
    }
}

fn state_root(home: &Path) -> PathBuf {
    home.join(crate::HOME_STATE_DIR).join(STATE_DIR)
}

fn sha(bytes: &[u8]) -> String {
    crate::digest(bytes)
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct FilePlan {
    /// File name inside the host folder, e.g. `moai-impl.md`.
    pub file: String,
    /// create | update | unchanged | preserved | conflict | remove_blocked
    pub action: String,
    pub baseline_sha256: Option<String>,
    pub current_sha256: Option<String>,
    pub new_sha256: Option<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct HostPlan {
    pub host: String,
    pub dir: String,
    /// ok | blocked
    pub preflight: String,
    pub blocked_reason: Option<String>,
    pub files: Vec<FilePlan>,
    /// What this host cannot express, per file (`agent_defs::Rendered::not_supported`).
    pub not_supported: BTreeMap<String, Vec<String>>,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct Plan {
    pub schema_version: u32,
    pub home: String,
    pub package_digest: String,
    pub hosts: Vec<HostPlan>,
    pub plan_digest: String,
}

/// What our applies own in one host folder: file name → sha256 we last wrote.
#[derive(Clone, Debug, Default, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
struct HostState {
    schema_version: u32,
    owned: BTreeMap<String, String>,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct ReceiptEntry {
    pub host: String,
    pub file: String,
    /// create | update
    pub action: String,
    pub before_sha256: Option<String>,
    pub after_sha256: String,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct Receipt {
    pub schema_version: u32,
    pub id: String,
    pub plan_digest: String,
    pub permission_reference: String,
    /// pending | committed | reverted
    pub status: String,
    pub entries: Vec<ReceiptEntry>,
}

/// A path that must be a real folder (or absent): never a symlink, never a file.
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

/// Every folder from the host folder (`~/.claude`, `~/.omp`) down to its agent folder
/// (`~/.claude/agents`, `~/.omp/agent/agents`) is a real folder or absent — never a link,
/// never a file. OMP's agent folder sits two levels below `~/.omp`, so each level is checked
/// (#2633). Plan, apply and revert all check it.
fn host_folders_plain(home: &Path, host: &str) -> Result<()> {
    let dir = host_dir(home, host);
    let below_home = dir.strip_prefix(home).expect("host folder under HOME");
    let mut path = home.to_path_buf();
    for part in below_home.components() {
        path.push(part);
        plain_dir_or_absent(&path)?;
    }
    Ok(())
}

/// The current bytes of one agent file: `Ok(None)` when absent; a symlink or a folder in its
/// place is an error (the host is blocked, never followed).
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

fn read_state(home: &Path, host: &str) -> Result<HostState> {
    let root = state_root(home);
    plain_dir_or_absent(&home.join(crate::HOME_STATE_DIR))?;
    plain_dir_or_absent(&root)?;
    match read_opt(&root.join(format!("{host}.json")))? {
        None => Ok(HostState {
            schema_version: SCHEMA_VERSION,
            owned: BTreeMap::new(),
        }),
        Some(bytes) => {
            let state: HostState = serde_json::from_slice(&bytes)
                .map_err(|_| format!("agent state for {host} is unreadable; inspect it first"))?;
            if state.schema_version != SCHEMA_VERSION {
                return Err(format!("agent state for {host} has an unknown schema"));
            }
            Ok(state)
        }
    }
}

fn write_json<T: Serialize>(path: &Path, value: &T) -> Result<()> {
    let mut bytes = serde_json::to_vec_pretty(value).map_err(|_| "encode failed")?;
    bytes.push(b'\n');
    let before = read_opt(path)?;
    match before {
        Some(b) => publish_replace(path, Some(&b), &bytes),
        None => publish_exclusive(path, &bytes),
    }
}

fn host_plan(home: &Path, host: &str, rendered: &[RenderedFile]) -> HostPlan {
    let dir = host_dir(home, host);
    let mut plan = HostPlan {
        host: host.to_string(),
        dir: dir.to_string_lossy().into_owned(),
        preflight: "ok".into(),
        blocked_reason: None,
        files: Vec::new(),
        not_supported: BTreeMap::new(),
    };
    let blocked = |plan: &mut HostPlan, reason: String| {
        plan.preflight = "blocked".into();
        plan.blocked_reason = Some(reason);
        plan.files.clear();
    };
    if let Err(e) = host_folders_plain(home, host) {
        blocked(&mut plan, e);
        return plan;
    }
    let state = match read_state(home, host) {
        Ok(s) => s,
        Err(e) => {
            blocked(&mut plan, e);
            return plan;
        }
    };
    for (file, text, notes) in rendered {
        if !notes.is_empty() {
            plan.not_supported.insert(file.clone(), notes.clone());
        }
        let current = match current_file(&dir.join(file)) {
            Ok(c) => c.map(|b| sha(&b)),
            Err(e) => {
                blocked(&mut plan, e);
                return plan;
            }
        };
        let new = sha(text.as_bytes());
        let baseline = state.owned.get(file).cloned();
        let action = match decide(baseline.as_deref(), current.as_deref(), Some(&new)) {
            Decision::KeepCurrent if current.as_deref() == Some(new.as_str()) => "unchanged",
            // The rendering did not change since our last apply: keep the user's edit.
            Decision::KeepCurrent => "preserved",
            Decision::Apply if current.is_none() => "create",
            Decision::Apply => "update",
            Decision::Conflict => "conflict",
        };
        plan.files.push(FilePlan {
            file: file.clone(),
            action: action.into(),
            baseline_sha256: baseline,
            current_sha256: current,
            new_sha256: Some(new),
        });
    }
    // Files we wrote before that the package no longer renders: never deleted, reported.
    for (file, owned_sha) in &state.owned {
        if rendered.iter().any(|(f, _, _)| f == file) {
            continue;
        }
        let current = current_file(&dir.join(file))
            .ok()
            .flatten()
            .map(|b| sha(&b));
        plan.files.push(FilePlan {
            file: file.clone(),
            action: "remove_blocked".into(),
            baseline_sha256: Some(owned_sha.clone()),
            current_sha256: current,
            new_sha256: None,
        });
    }
    plan.files.sort_by(|a, b| a.file.cmp(&b.file));
    plan
}

/// Every host's rendered agent files: (file name, text, not_supported).
fn rendered_all(package: &Path) -> Result<BTreeMap<String, Vec<RenderedFile>>> {
    let defs = package_agents(package)?;
    let mut out = BTreeMap::new();
    for host in HOSTS {
        let mut files = Vec::new();
        for def in &defs {
            let r = render(def, host)?;
            let file = r
                .path
                .strip_prefix("agents/")
                .ok_or("rendered agent path outside agents/")?
                .to_string();
            files.push((file, r.text, r.not_supported));
        }
        out.insert(host.to_string(), files);
    }
    Ok(out)
}

/// Read-only preview of every host; writes nothing.
pub fn plan(home: &Path, package: &Path) -> Result<Plan> {
    plan_for(home, package, &HOSTS)
}

/// Read-only preview of the named hosts only (#2638: `moai-cli update` leaves out a host whose
/// skill copy is blocked); writes nothing. An unknown host is refused.
pub fn plan_for(home: &Path, package: &Path, hosts: &[&str]) -> Result<Plan> {
    if let Some(h) = hosts.iter().find(|h| !HOSTS.contains(h)) {
        return Err(format!("agent host not supported: {h:?}"));
    }
    let home = fs::canonicalize(home).map_err(|_| "home directory unavailable")?;
    let package_digest = crate::package_digest(package)?;
    let rendered = rendered_all(package)?;
    let hosts = HOSTS
        .iter()
        .filter(|h| hosts.contains(h))
        .map(|h| host_plan(&home, h, &rendered[*h]))
        .collect();
    let mut plan = Plan {
        schema_version: SCHEMA_VERSION,
        home: home.to_str().ok_or("non UTF-8 home path")?.to_string(),
        package_digest,
        hosts,
        plan_digest: String::new(),
    };
    plan.plan_digest = crate::digest(&serde_json::to_vec(&plan).map_err(|_| "encode failed")?);
    Ok(plan)
}

fn validate_permission(permission: &str) -> Result<()> {
    // One rule for every operation (#2642 `contract::permission_reference_ok`).
    if !crate::contract::permission_reference_ok(permission) {
        return Err("permission reference required".into());
    }
    Ok(())
}

fn new_receipt_dir(home: &Path) -> Result<(String, PathBuf)> {
    let receipts = state_root(home).join("receipts");
    fs::create_dir_all(&receipts).map_err(|_| "agent receipt folder unavailable")?;
    plain_dir_or_absent(&receipts)?;
    let nanos = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_nanos())
        .unwrap_or(0);
    for n in 0..16 {
        let id = format!("agents-{nanos}-{n}");
        let dir = receipts.join(&id);
        if fs::create_dir(&dir).is_ok() {
            return Ok((id, dir));
        }
    }
    Err("agent receipt id unavailable".into())
}

/// Applies a previewed plan. `plan_digest` must equal a fresh plan's digest; only `create` and
/// `update` are written, blocked hosts are skipped.
pub fn apply(
    home: &Path,
    package: &Path,
    plan_digest: &str,
    permission: &str,
) -> Result<serde_json::Value> {
    apply_for(home, package, &HOSTS, plan_digest, permission)
}

/// [`apply`] for the named hosts only; `plan_digest` must equal a fresh [`plan_for`] of the
/// same hosts.
pub fn apply_for(
    home: &Path,
    package: &Path,
    hosts: &[&str],
    plan_digest: &str,
    permission: &str,
) -> Result<serde_json::Value> {
    validate_permission(permission)?;
    let fresh = plan_for(home, package, hosts)?;
    if fresh.plan_digest != plan_digest {
        return Err("stale or altered plan; re-preview required".into());
    }
    let home = PathBuf::from(&fresh.home);
    let rendered = rendered_all(package)?;
    let writes: Vec<(&HostPlan, &FilePlan)> = fresh
        .hosts
        .iter()
        .filter(|h| h.preflight == "ok")
        .flat_map(|h| h.files.iter().map(move |f| (h, f)))
        .filter(|(_, f)| f.action == "create" || f.action == "update")
        .collect();
    if writes.is_empty() {
        return Ok(
            serde_json::json!({"state": "unchanged", "plan_digest": fresh.plan_digest, "written": []}),
        );
    }
    let (id, receipt_dir) = new_receipt_dir(&home)?;
    let mut receipt = Receipt {
        schema_version: SCHEMA_VERSION,
        id: id.clone(),
        plan_digest: fresh.plan_digest.clone(),
        permission_reference: permission.to_string(),
        status: "pending".into(),
        entries: Vec::new(),
    };
    let receipt_path = receipt_dir.join("receipt.json");
    write_json(&receipt_path, &receipt)?;
    let mut written = Vec::new();
    for (host, file) in writes {
        let dir = host_dir(&home, &host.host);
        // Checked before creating (create_dir_all would follow a linked parent) and after.
        host_folders_plain(&home, &host.host)?;
        fs::create_dir_all(&dir).map_err(|_| format!("{} unavailable", dir.display()))?;
        host_folders_plain(&home, &host.host)?;
        let target = dir.join(&file.file);
        let text = &rendered[&host.host]
            .iter()
            .find(|(f, _, _)| *f == file.file)
            .ok_or("rendered file missing")?
            .1;
        let before = current_file(&target)?;
        if before.as_deref().map(sha) != file.current_sha256 {
            return Err(format!(
                "{} changed since the plan; re-preview required",
                target.display()
            ));
        }
        if let Some(b) = &before {
            let keep = receipt_dir.join("before").join(&host.host);
            fs::create_dir_all(&keep).map_err(|_| "agent receipt folder unavailable")?;
            publish_exclusive(&keep.join(&file.file), b)?;
        }
        match &before {
            Some(b) => publish_replace(&target, Some(b), text.as_bytes())?,
            None => publish_exclusive(&target, text.as_bytes())?,
        }
        let after = sha(text.as_bytes());
        receipt.entries.push(ReceiptEntry {
            host: host.host.clone(),
            file: file.file.clone(),
            action: file.action.clone(),
            before_sha256: before.as_deref().map(sha),
            after_sha256: after.clone(),
        });
        write_json(&receipt_path, &receipt)?;
        let mut state = read_state(&home, &host.host)?;
        state.owned.insert(file.file.clone(), after);
        write_json(
            &state_root(&home).join(format!("{}.json", host.host)),
            &state,
        )?;
        written.push(format!("{}:{}", host.host, file.file));
    }
    receipt.status = "committed".into();
    write_json(&receipt_path, &receipt)?;
    Ok(serde_json::json!({
        "state": "applied",
        "plan_digest": fresh.plan_digest,
        "receipt": id,
        "written": written,
    }))
}

fn receipt_id_ok(id: &str) -> bool {
    id.starts_with("agents-")
        && id.len() <= 64
        && id.bytes().all(|b| b.is_ascii_alphanumeric() || b == b'-')
}

/// Undoes one apply for the files that still hold the bytes it wrote: a created file is
/// removed, an updated one gets its before bytes back. A file changed since is skipped.
pub fn revert(home: &Path, receipt_id: &str, permission: &str) -> Result<serde_json::Value> {
    validate_permission(permission)?;
    if !receipt_id_ok(receipt_id) {
        return Err("unknown agent receipt id".into());
    }
    let home = fs::canonicalize(home).map_err(|_| "home directory unavailable")?;
    let dir = state_root(&home).join("receipts").join(receipt_id);
    plain_dir_or_absent(&dir)?;
    let path = dir.join("receipt.json");
    let bytes = read_opt(&path)?.ok_or("agent receipt not found")?;
    let mut receipt: Receipt =
        serde_json::from_slice(&bytes).map_err(|_| "agent receipt is unreadable")?;
    if receipt.status == "reverted" {
        return Err("agent receipt already reverted".into());
    }
    let mut reverted = Vec::new();
    let mut skipped = Vec::new();
    for e in receipt.entries.iter().rev() {
        if !HOSTS.contains(&e.host.as_str()) || e.file.contains('/') || e.file.contains('\\') {
            return Err("agent receipt names a path outside the agent folders".into());
        }
        // Audit FAIL 6842034981 B1: the folders may have turned into links since the apply;
        // a blocked folder is reported, never written through.
        if let Err(reason) = host_folders_plain(&home, &e.host) {
            skipped.push(serde_json::json!({"host": e.host, "file": e.file, "reason": format!("agent_folder_not_plain: {reason}")}));
            continue;
        }
        let target = host_dir(&home, &e.host).join(&e.file);
        let current = current_file(&target)?;
        if current.as_deref().map(sha) != Some(e.after_sha256.clone()) {
            skipped.push(serde_json::json!({"host": e.host, "file": e.file, "reason": "user_changed_since_apply"}));
            continue;
        }
        let current = current.expect("checked above");
        let mut state = read_state(&home, &e.host)?;
        match &e.before_sha256 {
            None => {
                fs::remove_file(&target)
                    .map_err(|_| format!("{} not removed", target.display()))?;
                state.owned.remove(&e.file);
            }
            Some(before_sha) => {
                let before = read_opt(&dir.join("before").join(&e.host).join(&e.file))?
                    .ok_or("before bytes missing from the receipt")?;
                if &sha(&before) != before_sha {
                    return Err("before bytes do not match the receipt".into());
                }
                publish_replace(&target, Some(&current), &before)?;
                state.owned.insert(e.file.clone(), before_sha.clone());
            }
        }
        write_json(&state_root(&home).join(format!("{}.json", e.host)), &state)?;
        reverted.push(format!("{}:{}", e.host, e.file));
    }
    receipt.status = "reverted".into();
    write_json(&path, &receipt)?;
    Ok(
        serde_json::json!({"state": "reverted", "receipt": receipt_id, "reverted": reverted, "skipped": skipped}),
    )
}
