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
use crate::retired_hosts::retired;
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
/// (code.claude.com/docs/en/sub-agents). The host folder is [`crate::host_home`]'s
/// (`CLAUDE_CONFIG_DIR`, card C-79); a relocated `agents` link to the default folder is
/// followed there (decision "기본 폴더 링크만 허용").
fn host_dir(home: &Path, host: &str) -> Result<PathBuf> {
    Ok(host_place(home, host)?.path())
}

/// The agent folder's place: the folder the link checks start at and the path below it.
fn host_place(home: &Path, host: &str) -> Result<crate::host_home::Place> {
    let h = crate::host_home::Host::from_name(host)
        .unwrap_or_else(|| unreachable!("unknown agent host {host}"));
    crate::host_home::place(h, home, host_rel(host))
}

/// The agent folder below the host folder. Claude Code gets its own `moai` folder inside
/// `agents/`: the official page scans `~/.claude/agents/` recursively and names an agent by its
/// `name` field only, so the five base agents keep their names and stay apart from the user's
/// own agents (SPEC-20261121-base-agents-user-folder).
fn host_rel(host: &str) -> &'static str {
    match host {
        "claude" => "agents/moai",
        _ => "agents",
    }
}

/// Where the earlier layout kept the base agents: the folder above `moai/`. `None` for a host
/// whose folder never moved.
fn legacy_rel(host: &str) -> Option<&'static str> {
    (host == "claude").then_some("agents")
}

/// The earlier flat folder of a host, with the same relocation rules as [`host_dir`].
fn legacy_dir(home: &Path, host: &str) -> Result<Option<PathBuf>> {
    let Some(rel) = legacy_rel(host) else {
        return Ok(None);
    };
    let h = crate::host_home::Host::from_name(host)
        .unwrap_or_else(|| unreachable!("unknown agent host {host}"));
    Ok(Some(crate::host_home::place(h, home, rel)?.path()))
}

fn state_root(home: &Path) -> PathBuf {
    crate::home_root::path(home).join(STATE_DIR)
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
    /// What the earlier flat layout left for this agent (Claude Code only; left out when there
    /// is nothing): `remove` (ours, unedited: the apply removes it after writing the nested
    /// copy), `kept_edited` (we wrote it, the user changed it) or `kept_foreign` (never ours).
    /// The last two are never touched and block the nested copy.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub legacy: Option<String>,
    /// The digest of that flat file, so a change between preview and apply is a stale plan.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub legacy_sha256: Option<String>,
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
    /// Tree digest of the agent source folder (#2721); `None` when it is absent.
    pub agents_digest: Option<String>,
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
    /// Per host, how many folders at the end of the agent folder's path (below the host folder)
    /// this apply created; the revert removes them again while they are empty. Absent in a
    /// receipt that created none.
    #[serde(default, skip_serializing_if = "BTreeMap::is_empty")]
    pub created_folders: BTreeMap<String, usize>,
}

/// How many folders at the end of the agent folder's path do not exist yet, counted from the
/// agent folder up and never past the host folder (`place.base`) or above `place.rel`.
fn missing_levels(place: &crate::host_home::Place) -> usize {
    let dir = place.path();
    dir.ancestors()
        .take(place.rel.components().count())
        .take_while(|d| fs::symlink_metadata(d).is_err())
        .count()
}

/// Removes up to `levels` empty folders at the end of the host's agent folder path, from the
/// agent folder up, stopping at the first folder that holds anything (the folders a revert
/// leaves behind when its apply created them).
fn remove_created_folders(home: &Path, host: &str, levels: usize) -> Result<Vec<String>> {
    let place = host_place(home, host)?;
    let mut dir = place.path();
    let mut removed = Vec::new();
    for _ in 0..levels.min(place.rel.components().count()) {
        let empty = fs::read_dir(&dir)
            .map(|mut entries| entries.next().is_none())
            .unwrap_or(false);
        if !empty {
            break;
        }
        fs::remove_dir(&dir).map_err(|_| format!("{} not removed", dir.display()))?;
        removed.push(dir.display().to_string());
        dir.pop();
    }
    Ok(removed)
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

/// Every folder from the host folder (`~/.claude`) down to its agent folder
/// (`~/.claude/agents`) is a real folder or absent — never a link, never a file. Plan, apply
/// and revert all check it.
fn host_folders_plain(home: &Path, host: &str) -> Result<()> {
    let place = host_place(home, host)?;
    let mut path = place.base;
    plain_dir_or_absent(&path)?;
    for part in place.rel.components() {
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
    plain_dir_or_absent(&crate::home_root::path(home))?;
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

fn host_plan(
    home: &Path,
    host: &str,
    rendered: &[RenderedFile],
    sources: &BTreeMap<String, String>,
) -> HostPlan {
    // An unnamed host folder (an empty or relative relocation variable) shows the folder
    // under HOME and is blocked below by `host_folders_plain`.
    let dir = host_dir(home, host).unwrap_or_else(|_| {
        let default = crate::host_home::Host::from_name(host).map_or(".", |h| h.default_dir());
        home.join(default).join(host_rel(host))
    });
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
    let flat_dir = match legacy_dir(home, host) {
        Ok(d) => d,
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
        let owned = state.owned.get(file).cloned();
        // The earlier flat layout: the same file name one folder up.
        let flat = match &flat_dir {
            None => None,
            Some(d) => match current_file(&d.join(file)) {
                Ok(c) => c.map(|b| sha(&b)),
                Err(e) => {
                    blocked(&mut plan, e);
                    return plan;
                }
            },
        };
        // Ours (SPEC-20261124-update-flat-agent-ownership REQ-1): the bytes this tool recorded,
        // the bytes it renders now, or the role's file in the package being applied. Anything
        // else is the person's: edited when we had a record for it, foreign when we had none.
        let legacy = flat.as_deref().map(|f| {
            let ours = owned.as_deref() == Some(f)
                || f == new
                || sources.get(file).map(String::as_str) == Some(f);
            match owned.as_deref() {
                _ if ours => "remove",
                Some(_) => "kept_edited",
                None => "kept_foreign",
            }
        });
        // A record made for the flat file is not the baseline of a nested file that is not
        // there yet: that one is a fresh create.
        let baseline = if current.is_none() && flat.is_some() {
            None
        } else {
            owned
        };
        let action = match decide(baseline.as_deref(), current.as_deref(), Some(&new)) {
            // A flat file that is not ours to remove blocks the nested copy: two files of one
            // scope with one name load only one of them.
            _ if matches!(legacy, Some("kept_edited" | "kept_foreign")) => "conflict",
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
            legacy: legacy.map(str::to_string),
            legacy_sha256: flat,
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
            legacy: None,
            legacy_sha256: None,
        });
    }
    plan.files.sort_by(|a, b| a.file.cmp(&b.file));
    plan
}

/// Every host's rendered agent files: (file name, text, not_supported).
fn rendered_all(package: &Path, agents: &Path) -> Result<BTreeMap<String, Vec<RenderedFile>>> {
    let defs = package_agents(agents, package)?;
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

/// The sha of each role's file in the package being applied (`agents/<role>.md`), keyed by the
/// rendered file name (`moai-<role>.md`). An earlier install copied these files as they were,
/// so a flat file with these bytes is ours. A role with no file in the folder has no entry.
fn package_file_shas(
    agents: &Path,
    rendered: &BTreeMap<String, Vec<RenderedFile>>,
) -> Result<BTreeMap<String, String>> {
    let mut out = BTreeMap::new();
    for files in rendered.values() {
        for (file, _, _) in files {
            let Some(role) = file
                .strip_prefix("moai-")
                .and_then(|r| r.strip_suffix(".md"))
            else {
                continue;
            };
            if let Some(bytes) = current_file(&agents.join(format!("{role}.md")))? {
                out.insert(file.clone(), sha(&bytes));
            }
        }
    }
    Ok(out)
}

/// Read-only preview of every host; writes nothing. `package` is the skill package the
/// agents' `reads` name, `agents` the source folder outside it (#2721, `agent_defs`).
pub fn plan(home: &Path, package: &Path, agents: &Path) -> Result<Plan> {
    plan_for(home, package, agents, &HOSTS)
}

/// Read-only preview of the named hosts only (#2638: `moai-cli update` leaves out a host whose
/// skill copy is blocked); writes nothing. An unknown host is refused.
pub fn plan_for(home: &Path, package: &Path, agents: &Path, hosts: &[&str]) -> Result<Plan> {
    if let Some(h) = hosts.iter().find(|h| !HOSTS.contains(h)) {
        return Err(format!("agent host not supported: {h:?}"));
    }
    let home = fs::canonicalize(home).map_err(|_| "home directory unavailable")?;
    let package_digest = crate::package_digest(package)?;
    let rendered = rendered_all(package, agents)?;
    // The source folder's own digest (#2721); `None` when there is no folder.
    let agents_digest = crate::agent_defs::source_digest(agents)?;
    let sources = package_file_shas(agents, &rendered)?;
    let hosts = HOSTS
        .iter()
        .filter(|h| hosts.contains(h))
        .map(|h| host_plan(&home, h, &rendered[*h], &sources))
        .collect();
    let mut plan = Plan {
        schema_version: SCHEMA_VERSION,
        home: home.to_str().ok_or("non UTF-8 home path")?.to_string(),
        package_digest,
        agents_digest,
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
    agents: &Path,
    plan_digest: &str,
    permission: &str,
) -> Result<serde_json::Value> {
    apply_for(home, package, agents, &HOSTS, plan_digest, permission)
}

/// [`apply`] for the named hosts only; `plan_digest` must equal a fresh [`plan_for`] of the
/// same hosts.
pub fn apply_for(
    home: &Path,
    package: &Path,
    agents: &Path,
    hosts: &[&str],
    plan_digest: &str,
    permission: &str,
) -> Result<serde_json::Value> {
    apply_with_choices(home, package, agents, hosts, plan_digest, permission, &[])
}

/// The `host:file` name a plan line goes by in a choice and in the result lists.
fn key(host: &str, file: &str) -> String {
    format!("{host}:{file}")
}

/// Whether a plan line is a flat-file conflict: a flat copy of the earlier layout that is not
/// ours (edited, or never written by this tool), which blocks the nested copy.
fn is_flat_conflict(f: &FilePlan) -> bool {
    f.action == "conflict" && matches!(f.legacy.as_deref(), Some("kept_edited" | "kept_foreign"))
}

/// The `host:file` name of every flat-file conflict of the plan, in plan order.
pub fn flat_conflicts(p: &Plan) -> Vec<String> {
    p.hosts
        .iter()
        .filter(|h| h.preflight == "ok")
        .flat_map(|h| h.files.iter().map(move |f| (h, f)))
        .filter(|(_, f)| is_flat_conflict(f))
        .map(|(h, f)| key(&h.host, &f.file))
        .collect()
}

/// What the apply does with the nested file of a flat-file conflict whose flat file the person
/// gave up: write it when it is absent or still what we wrote, leave it when it already holds
/// the render, and refuse when it was edited too (never overwritten by a choice about the
/// flat file).
fn chosen_action(f: &FilePlan) -> Result<&'static str> {
    let new = f.new_sha256.as_deref();
    match f.current_sha256.as_deref() {
        None => Ok("create"),
        current if current == new => Ok("unchanged"),
        current => match decide(f.baseline_sha256.as_deref(), current, new) {
            Decision::Apply => Ok("update"),
            _ => Err(format!(
                "{}: the nested copy was edited too; nothing written, resolve it by hand",
                f.file
            )),
        },
    }
}

/// [`apply_for`] that also uses MoAI's file for the named flat-file conflicts (`host:file`):
/// the person's flat file is kept in the receipt, the nested copy is written and the flat file
/// removed. A name that is no flat-file conflict of the fresh plan refuses the whole apply with
/// nothing written. A conflict left over makes the state `partial` with the names in `kept`.
pub fn apply_with_choices(
    home: &Path,
    package: &Path,
    agents: &Path,
    hosts: &[&str],
    plan_digest: &str,
    permission: &str,
    use_moais: &[String],
) -> Result<serde_json::Value> {
    validate_permission(permission)?;
    let fresh = plan_for(home, package, agents, hosts)?;
    if fresh.plan_digest != plan_digest {
        return Err("stale or altered plan; re-preview required".into());
    }
    let conflicts = flat_conflicts(&fresh);
    for c in use_moais {
        if !conflicts.contains(c) {
            return Err(format!(
                "{c}: not a flat-file conflict of the plan; nothing written"
            ));
        }
    }
    let chosen = |host: &str, file: &FilePlan| use_moais.contains(&key(host, &file.file));
    let kept: Vec<String> = conflicts
        .iter()
        .filter(|c| !use_moais.contains(c))
        .cloned()
        .collect();
    let home = PathBuf::from(&fresh.home);
    let rendered = rendered_all(package, agents)?;
    let writes: Vec<(&HostPlan, &FilePlan)> = fresh
        .hosts
        .iter()
        .filter(|h| h.preflight == "ok")
        .flat_map(|h| h.files.iter().map(move |f| (h, f)))
        .filter(|(h, f)| writes_anything(f) || chosen(&h.host, f))
        .collect();
    // Every choice is checked before the first write.
    for (h, f) in &writes {
        if chosen(&h.host, f) {
            chosen_action(f)?;
        }
    }
    // A host the plan blocked (a link at its folder, say) is never written; the result names it.
    let blocked: Vec<(&str, &str)> = fresh
        .hosts
        .iter()
        .filter(|h| h.preflight != "ok")
        .map(|h| {
            (
                h.host.as_str(),
                h.blocked_reason.as_deref().unwrap_or("blocked"),
            )
        })
        .collect();
    if writes.is_empty() {
        let mut v = serde_json::json!({"state": "unchanged", "plan_digest": fresh.plan_digest, "written": []});
        if !kept.is_empty() {
            v["state"] = "partial".into();
            v["kept"] = serde_json::json!(kept);
        }
        name_blocked(&mut v, &blocked);
        return Ok(v);
    }
    let (id, receipt_dir) = new_receipt_dir(&home)?;
    let mut receipt = Receipt {
        schema_version: SCHEMA_VERSION,
        id: id.clone(),
        plan_digest: fresh.plan_digest.clone(),
        permission_reference: permission.to_string(),
        status: "pending".into(),
        entries: Vec::new(),
        created_folders: BTreeMap::new(),
    };
    let receipt_path = receipt_dir.join("receipt.json");
    write_json(&receipt_path, &receipt)?;
    let mut written = Vec::new();
    for (host, file) in writes {
        let dir = host_dir(&home, &host.host)?;
        // Checked before creating (create_dir_all would follow a linked parent) and after.
        host_folders_plain(&home, &host.host)?;
        // Recorded before creating, so the revert can take back the folders it made (audit B1).
        let missing = missing_levels(&host_place(&home, &host.host)?);
        if missing > 0 && !receipt.created_folders.contains_key(&host.host) {
            receipt.created_folders.insert(host.host.clone(), missing);
            write_json(&receipt_path, &receipt)?;
        }
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
        // A chosen conflict is written like a move: its effective action is decided from the
        // nested file (checked above), its flat file goes after the nested copy.
        let replacing = chosen(&host.host, file);
        let action = if replacing {
            chosen_action(file)?
        } else {
            file.action.as_str()
        };
        // The nested copy first: a stop after it leaves a duplicate that the next preview shows
        // as a move again, never a missing agent (SPEC risk 2).
        if action == "create" || action == "update" {
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
                action: action.to_string(),
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
        if replacing || file.legacy.as_deref() == Some("remove") {
            let flat_dir =
                legacy_dir(&home, &host.host)?.ok_or("no earlier folder for the host")?;
            let flat = flat_dir.join(&file.file);
            let bytes = current_file(&flat)?;
            if bytes.as_deref().map(sha) != file.legacy_sha256 {
                return Err(format!(
                    "{} changed since the plan; re-preview required",
                    flat.display()
                ));
            }
            let bytes = bytes.ok_or("earlier file missing")?;
            let keep = receipt_dir.join("before").join(&host.host).join("legacy");
            fs::create_dir_all(&keep).map_err(|_| "agent receipt folder unavailable")?;
            publish_exclusive(&keep.join(&file.file), &bytes)?;
            fs::remove_file(&flat).map_err(|_| format!("{} not removed", flat.display()))?;
            receipt.entries.push(ReceiptEntry {
                host: host.host.clone(),
                file: file.file.clone(),
                // A file the person gave up is its own action: the revert puts it back without
                // making it ours (the record must not name the person's bytes).
                action: if replacing {
                    "replace_legacy".into()
                } else {
                    "remove_legacy".into()
                },
                before_sha256: Some(sha(&bytes)),
                after_sha256: sha(&bytes),
            });
            write_json(&receipt_path, &receipt)?;
            // The record follows the file: it names the nested bytes now.
            let mut state = read_state(&home, &host.host)?;
            if let Some(now) = current_file(&target)?.map(|b| sha(&b)) {
                if Some(&now) == file.new_sha256.as_ref() {
                    state.owned.insert(file.file.clone(), now);
                    write_json(
                        &state_root(&home).join(format!("{}.json", host.host)),
                        &state,
                    )?;
                }
            }
            written.push(format!("{}:{} (moved)", host.host, file.file));
        }
    }
    receipt.status = "committed".into();
    write_json(&receipt_path, &receipt)?;
    let mut v = serde_json::json!({
        "state": "applied",
        "plan_digest": fresh.plan_digest,
        "receipt": id,
        "written": written,
    });
    if !kept.is_empty() {
        v["state"] = "partial".into();
        v["kept"] = serde_json::json!(kept);
    }
    name_blocked(&mut v, &blocked);
    Ok(v)
}

/// Adds the blocked hosts (`host`, `reason`) to an apply result, with no key when none is blocked.
/// A result that wrote nothing becomes `blocked` with a one-line `reason`, so no screen reads it as
/// done; a result that wrote something for another host is `partial`.
fn name_blocked(v: &mut serde_json::Value, blocked: &[(&str, &str)]) {
    if blocked.is_empty() {
        return;
    }
    v["blocked"] = blocked
        .iter()
        .map(|(host, reason)| serde_json::json!({"host": host, "reason": reason}))
        .collect();
    v["reason"] = blocked
        .iter()
        .map(|(host, reason)| format!("{host}: {reason}"))
        .collect::<Vec<_>>()
        .join("; ")
        .into();
    v["state"] = if v["state"] == "unchanged" {
        "blocked".into()
    } else {
        "partial".into()
    };
}

/// Whether the apply writes anything for this file: a nested create or update, or the removal
/// of an earlier flat copy that is still ours and unedited.
pub fn writes_anything(f: &FilePlan) -> bool {
    f.action == "create" || f.action == "update" || f.legacy.as_deref() == Some("remove")
}

/// SPEC-20261124-update-flat-agent-ownership REQ-11: whether the update has something to ask
/// about for this plan: a write, or a flat-file conflict the person can choose for. A conflict
/// alone writes nothing, but without the question the person could never choose.
pub fn asks_anything(p: &Plan) -> bool {
    p.hosts.iter().any(|h| {
        h.preflight == "ok"
            && h.files
                .iter()
                .any(|f| writes_anything(f) || is_flat_conflict(f))
    })
}

/// The most lines of a difference that are printed; the rest is counted.
const DIFF_LINES: usize = 80;

/// The lines that differ between `mine` and `moai`, in order: `- ` for a line only in mine and
/// `+ ` for a line only in MoAI's, from the longest common run of lines. Every line goes through
/// [`crate::ui::visible`], so a control character of the person's file cannot move the cursor.
fn difference(mine: &str, moai: &str) -> Vec<String> {
    let a: Vec<&str> = mine.lines().collect();
    let b: Vec<&str> = moai.lines().collect();
    // `common[i][j]`: the length of the longest common run of a[i..] and b[j..].
    let mut common = vec![vec![0usize; b.len() + 1]; a.len() + 1];
    for i in (0..a.len()).rev() {
        for j in (0..b.len()).rev() {
            common[i][j] = if a[i] == b[j] {
                common[i + 1][j + 1] + 1
            } else {
                common[i + 1][j].max(common[i][j + 1])
            };
        }
    }
    let (mut i, mut j) = (0, 0);
    let mut out = Vec::new();
    while i < a.len() || j < b.len() {
        if i < a.len() && j < b.len() && a[i] == b[j] {
            i += 1;
            j += 1;
        } else if j == b.len() || (i < a.len() && common[i + 1][j] >= common[i][j + 1]) {
            out.push(format!("- {}", crate::ui::visible(a[i])));
            i += 1;
        } else {
            out.push(format!("+ {}", crate::ui::visible(b[j])));
            j += 1;
        }
    }
    out
}

/// Asks once per flat-file conflict of `plan` whether to keep mine (`k`), use MoAI's (`m`) or
/// show the difference (`d`, then the same question again), and answers the `host:file` names
/// the person chose MoAI's file for. Anything else, an empty line and the end of the input keep
/// the person's file. Reads only; the choice is applied by [`apply_with_choices`].
pub fn choose_flat_conflicts(
    plan: &Plan,
    package: &Path,
    agents: &Path,
    input: &mut dyn std::io::BufRead,
    out: &mut dyn std::io::Write,
    korean: bool,
) -> Result<Vec<String>> {
    let rendered = rendered_all(package, agents)?;
    let home = Path::new(&plan.home);
    let mut chosen = Vec::new();
    let write_err = |_| "could not write the question".to_string();
    for name in flat_conflicts(plan) {
        let (host, file) = name.split_once(':').ok_or("unreadable conflict name")?;
        let moai = rendered
            .get(host)
            .and_then(|files| files.iter().find(|(f, _, _)| f == file))
            .map(|(_, text, _)| text.as_str())
            .ok_or("rendered file missing")?;
        let flat = legacy_dir(home, host)?
            .ok_or("no earlier folder for the host")?
            .join(file);
        let mine = current_file(&flat)?.ok_or("earlier file missing")?;
        let mine = String::from_utf8_lossy(&mine).into_owned();
        loop {
            let (head, options) = if korean {
                (
                    format!("{file}: 내 파일이 MoAI 것과 달라요"),
                    "  내 것 유지(k) · MoAI 것 쓰기(m) · 차이 보기(d) [k]: ",
                )
            } else {
                (
                    format!("{file}: your file differs from MoAI's"),
                    "  keep mine (k) · use MoAI's (m) · show the difference (d) [k]: ",
                )
            };
            writeln!(out, "{}", crate::ui::visible(&head)).map_err(write_err)?;
            write!(out, "{options}").map_err(write_err)?;
            out.flush().map_err(write_err)?;
            let mut answer = String::new();
            input
                .read_line(&mut answer)
                .map_err(|_| "could not read the answer")?;
            // The answer ends the prompt line even when nothing echoed the person's Enter (a
            // scripted input, the end of the input); on a terminal it is a separating blank line.
            writeln!(out).map_err(write_err)?;
            match answer.trim().to_ascii_lowercase().as_str() {
                "m" => {
                    chosen.push(name.clone());
                    break;
                }
                "d" => {
                    let lines = difference(&mine, moai);
                    for l in lines.iter().take(DIFF_LINES) {
                        writeln!(out, "{l}").map_err(write_err)?;
                    }
                    if lines.len() > DIFF_LINES {
                        let more = lines.len() - DIFF_LINES;
                        writeln!(out, "… {more}").map_err(write_err)?;
                    }
                }
                _ => break,
            }
        }
    }
    Ok(chosen)
}

fn receipt_id_ok(id: &str) -> bool {
    id.starts_with("agents-")
        && id.len() <= 64
        && id.bytes().all(|b| b.is_ascii_alphanumeric() || b == b'-')
}

/// Undoes one apply for the files that still hold the bytes it wrote: a created file is
/// removed, an updated one gets its before bytes back. A file changed since is skipped.
pub fn revert(home: &Path, receipt_id: &str, permission: &str) -> Result<serde_json::Value> {
    revert_in(&HOSTS, home, receipt_id, permission)
}

/// [`revert`] for a place whose active hosts are `active`: a receipt entry of a retired host is
/// skipped and reported (reason `host_retired`), the other entries are reverted.
pub fn revert_in(
    active: &[&str],
    home: &Path,
    receipt_id: &str,
    permission: &str,
) -> Result<serde_json::Value> {
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
        if retired(&e.host, active) {
            skipped.push(
                serde_json::json!({"host": e.host, "file": e.file, "reason": "host_retired"}),
            );
            continue;
        }
        if !HOSTS.contains(&e.host.as_str()) || e.file.contains('/') || e.file.contains('\\') {
            return Err("agent receipt names a path outside the agent folders".into());
        }
        // Audit FAIL 6842034981 B1: the folders may have turned into links since the apply;
        // a blocked folder is reported, never written through.
        if let Err(reason) = host_folders_plain(&home, &e.host) {
            skipped.push(serde_json::json!({"host": e.host, "file": e.file, "reason": format!("agent_folder_not_plain: {reason}")}));
            continue;
        }
        let moved =
            |n: &ReceiptEntry| n.action == "remove_legacy" && n.host == e.host && n.file == e.file;
        if e.action == "remove_legacy" || e.action == "replace_legacy" {
            // The flat file goes back only when the nested copy of the same apply is going
            // away too (unchanged since); a changed nested copy stays, with the flat file
            // gone, so no two files of one scope share a name.
            let made = receipt.entries.iter().find(|n| {
                n.host == e.host
                    && n.file == e.file
                    && (n.action == "create" || n.action == "update")
            });
            if let Some(n) = made {
                let nested = host_dir(&home, &e.host)?.join(&e.file);
                if current_file(&nested)?.as_deref().map(sha) != Some(n.after_sha256.clone()) {
                    skipped.push(serde_json::json!({"host": e.host, "file": e.file, "reason": "nested_changed_since_apply"}));
                    continue;
                }
            }
            let flat = legacy_dir(&home, &e.host)?
                .ok_or("no earlier folder for the host")?
                .join(&e.file);
            if current_file(&flat)?.is_some() {
                skipped.push(serde_json::json!({"host": e.host, "file": e.file, "reason": "earlier_path_occupied"}));
                continue;
            }
            let kept = dir
                .join("before")
                .join(&e.host)
                .join("legacy")
                .join(&e.file);
            let bytes = read_opt(&kept)?.ok_or("before bytes missing from the receipt")?;
            if Some(sha(&bytes)) != e.before_sha256 {
                return Err("before bytes do not match the receipt".into());
            }
            publish_exclusive(&flat, &bytes)?;
            // A restored flat file we moved is ours again, so the record names it. A file the
            // person gave up is the person's: the record is left alone, so the next update
            // sees a conflict again and never moves it unasked.
            if e.action == "remove_legacy" {
                let mut state = read_state(&home, &e.host)?;
                state.owned.insert(e.file.clone(), sha(&bytes));
                write_json(&state_root(&home).join(format!("{}.json", e.host)), &state)?;
            }
            reverted.push(format!("{}:{} (restored)", e.host, e.file));
            continue;
        }
        let target = host_dir(&home, &e.host)?.join(&e.file);
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
                // A restored flat file owns the record now.
                if !receipt.entries.iter().any(moved) {
                    state.owned.remove(&e.file);
                }
            }
            Some(before_sha) => {
                let before = read_opt(&dir.join("before").join(&e.host).join(&e.file))?
                    .ok_or("before bytes missing from the receipt")?;
                if &sha(&before) != before_sha {
                    return Err("before bytes do not match the receipt".into());
                }
                publish_replace(&target, Some(&current), &before)?;
                if !receipt.entries.iter().any(moved) {
                    state.owned.insert(e.file.clone(), before_sha.clone());
                }
            }
        }
        write_json(&state_root(&home).join(format!("{}.json", e.host)), &state)?;
        reverted.push(format!("{}:{}", e.host, e.file));
    }
    // Audit B1: the folders this apply created go too, while they are empty.
    let mut removed_folders = Vec::new();
    for (host, levels) in &receipt.created_folders {
        if retired(host, active)
            || !HOSTS.contains(&host.as_str())
            || host_folders_plain(&home, host).is_err()
        {
            continue;
        }
        removed_folders.extend(remove_created_folders(&home, host, *levels)?);
    }
    receipt.status = "reverted".into();
    write_json(&path, &receipt)?;
    Ok(
        serde_json::json!({"state": "reverted", "receipt": receipt_id, "reverted": reverted, "skipped": skipped, "removed_folders": removed_folders}),
    )
}
