//! The hook layer of `moai update` (#2853, after #2812 W3): the MoAI hook set already installed
//! in the project the command runs from is brought to what this binary generates, inside the
//! same preview and confirmation as the rest of the user layer.
//!
//! Leader decisions on #2853: D1 the hook commands keep naming `moai-cli` (`hooks::HOOK_BINARY`)
//! and an entry the user changed, or one MoAI never owned, is reported as a conflict and left
//! as it is; D2 the layer's digest is part of the one `plan_digest` `--confirm` takes; D3 tests
//! use a temporary HOME. Target (the default of LANE-BLOCKED 6870179526): the git work tree the
//! command runs in, and only the hosts whose hook set MoAI already installed there — `update`
//! never installs hooks; that stays with `hooks-apply`.
//!
//! Each host is planned and applied by `hooks::plan` / `hooks::apply`, the rules `hooks-apply`
//! uses. A confirmed apply keeps the bytes it replaced under a receipt, and `hooks-revert`
//! restores a host only while both of its files still hold what the apply wrote.

use crate::Result;
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};

/// One host of the hook layer.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct HostPlan {
    pub host: &'static str,
    /// `create`, `update`, `unchanged` or `conflict`.
    pub action: String,
    /// The settings keys in conflict (Claude Code).
    pub conflicts: Vec<String>,
    /// Why a host is a conflict when it has no key list (an unowned or changed hook file).
    pub reason: Option<String>,
    /// The host's own plan digest; `None` for a conflict, which is never applied.
    pub plan_digest: Option<String>,
}

impl HostPlan {
    fn writes(&self) -> bool {
        self.plan_digest.is_some() && matches!(self.action.as_str(), "create" | "update")
    }

    fn to_json(&self) -> Value {
        json!({
            "host": self.host,
            "action": self.action,
            "conflicts": self.conflicts,
            "reason": self.reason,
            "plan_digest": self.plan_digest,
        })
    }
}

/// The hook layer of one `moai update` preview.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Layer {
    /// The project's work tree, when the command runs inside one.
    pub project: Option<PathBuf>,
    pub hosts: Vec<HostPlan>,
    /// Why the layer is empty (`{part: "hooks", reason}`); reported inside `hooks`, so the
    /// top-level `skipped` rows keep their `host`.
    pub skipped: Vec<Value>,
    /// Binds the project and every host's action, conflicts and plan digest.
    pub plan_digest: String,
}

impl Layer {
    /// True when confirming would write at least one host file.
    pub fn writes(&self) -> bool {
        self.hosts.iter().any(HostPlan::writes)
    }

    pub fn to_json(&self) -> Value {
        json!({
            "project": self.project.as_deref().map(|p| p.display().to_string()),
            "hosts": self.hosts.iter().map(HostPlan::to_json).collect::<Vec<_>>(),
            "skipped": self.skipped,
            "plan_digest": self.plan_digest,
        })
    }

    /// The terminal question's lines for this layer.
    pub fn summary(&self) -> String {
        let Some(project) = &self.project else {
            return String::new();
        };
        if self.hosts.is_empty() {
            return String::new();
        }
        let mut out = format!("hooks ({})\n", project.display());
        for h in &self.hosts {
            let detail = if !h.conflicts.is_empty() {
                format!(" ({})", h.conflicts.join(", "))
            } else {
                h.reason
                    .as_ref()
                    .map(|r| format!(" ({r})"))
                    .unwrap_or_default()
            };
            out.push_str(&format!("  {}: {}{detail}\n", h.host, h.action));
        }
        out
    }
}

fn skip(reason: String) -> Value {
    json!({"part": "hooks", "reason": reason})
}

/// The work tree `cwd` belongs to, or why there is none.
fn project_of(cwd: &Path) -> std::result::Result<PathBuf, String> {
    let id = crate::identity::identity(cwd).map_err(|e| format!("no project here: {e}"))?;
    match id["work_tree"].as_str() {
        Some(top) => Ok(PathBuf::from(top)),
        None => Err("not inside a git work tree".into()),
    }
}

/// Plans the hook layer for the project `cwd` belongs to. Never writes.
pub fn plan(home: &Path, cwd: &Path) -> Result<Layer> {
    let mut layer = Layer {
        project: None,
        hosts: Vec::new(),
        skipped: Vec::new(),
        plan_digest: String::new(),
    };
    match project_of(cwd) {
        Err(reason) => layer.skipped.push(skip(reason)),
        Ok(root) => {
            let same_as_home = fs::canonicalize(home).ok().as_deref() == Some(root.as_path());
            if same_as_home {
                layer.skipped.push(skip(
                    "the work tree is the home directory; hooks are written to a project only"
                        .into(),
                ));
            } else {
                for host in crate::hooks::HOSTS {
                    match crate::hooks::installed(&root, host) {
                        Ok(false) => {}
                        Ok(true) => layer.hosts.push(host_plan(&root, home, host)),
                        Err(e) => layer.hosts.push(conflict(host, e)),
                    }
                }
                if layer.hosts.is_empty() {
                    layer
                        .skipped
                        .push(skip("no MoAI hook set is installed in this project".into()));
                }
                layer.project = Some(root);
            }
        }
    }
    let hosts: Vec<Value> = layer.hosts.iter().map(HostPlan::to_json).collect();
    let project = layer.project.as_deref().map(|p| p.display().to_string());
    layer.plan_digest = crate::digest(
        &serde_json::to_vec(&json!([1, project, hosts])).map_err(|_| "encode failed")?,
    );
    Ok(layer)
}

fn conflict(host: &'static str, reason: String) -> HostPlan {
    HostPlan {
        host,
        action: "conflict".into(),
        conflicts: Vec::new(),
        reason: Some(reason),
        plan_digest: None,
    }
}

fn host_plan(root: &Path, home: &Path, host: &'static str) -> HostPlan {
    let v = match crate::hooks::plan(root, host, Some(home)) {
        Ok(v) => v,
        Err(e) => return conflict(host, e),
    };
    let action = v["action"].as_str().unwrap_or("conflict").to_string();
    let conflicts = v["conflicts"]
        .as_array()
        .into_iter()
        .flatten()
        .filter_map(|c| c.as_str().map(str::to_string))
        .collect();
    let plan_digest = (action != "conflict")
        .then(|| v["plan_digest"].as_str().map(str::to_string))
        .flatten();
    HostPlan {
        host,
        action,
        conflicts,
        reason: None,
        plan_digest,
    }
}

const RECEIPT_PREFIX: &str = "hooks-";

/// A revert receipt of one confirmed hook layer.
#[derive(Serialize, Deserialize)]
struct Receipt {
    schema_version: u32,
    id: String,
    plan_digest: String,
    permission_reference: String,
    status: String,
    hosts: Vec<HostEntry>,
}

#[derive(Serialize, Deserialize)]
struct HostEntry {
    host: String,
    files: Vec<FileEntry>,
}

#[derive(Serialize, Deserialize)]
struct FileEntry {
    /// Relative to the project; always one of `hooks::written_files` for the host.
    path: String,
    before_sha256: Option<String>,
    after_sha256: Option<String>,
}

fn receipts_dir(root: &Path) -> PathBuf {
    crate::state_root::path(root, "receipts")
}

fn receipt_id_ok(id: &str) -> bool {
    id.starts_with(RECEIPT_PREFIX)
        && id.len() <= 64
        && id.bytes().all(|b| b.is_ascii_alphanumeric() || b == b'-')
}

fn plain_dirs(dirs: &[PathBuf]) -> Result<()> {
    for d in dirs {
        crate::claude::refuse_symlink_dir(d)?;
    }
    Ok(())
}

fn new_receipt_dir(root: &Path) -> Result<(String, PathBuf)> {
    let receipts = receipts_dir(root);
    plain_dirs(&[crate::state_root::path(root, ""), receipts.clone()])?;
    fs::create_dir_all(&receipts).map_err(|_| "hook receipt folder unavailable")?;
    plain_dirs(std::slice::from_ref(&receipts))?;
    let nanos = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_nanos())
        .unwrap_or(0);
    for n in 0..16 {
        let id = format!("{RECEIPT_PREFIX}{nanos}-{n}");
        let dir = receipts.join(&id);
        if fs::create_dir(&dir).is_ok() {
            return Ok((id, dir));
        }
    }
    Err("hook receipt id unavailable".into())
}

fn write_json(path: &Path, value: &impl Serialize) -> Result<()> {
    let mut bytes = serde_json::to_vec_pretty(value).map_err(|_| "encode failed")?;
    bytes.push(b'\n');
    let before = crate::claude::read_existing(path)?;
    crate::apply::publish_replace(path, before.as_deref(), &bytes)
}

/// Applies the confirmed hook layer: each host that writes goes through `hooks::apply` with its
/// confirmed digest (which re-plans and refuses a changed file), its replaced bytes kept under
/// one receipt. A host's failure is reported beside the others, never hiding them.
pub fn apply(home: &Path, layer: &Layer, plan_digest: &str, permission: &str) -> Result<Value> {
    let writes: Vec<&HostPlan> = layer.hosts.iter().filter(|h| h.writes()).collect();
    let Some(root) = layer.project.as_deref().filter(|_| !writes.is_empty()) else {
        return Ok(json!({"state": "unchanged", "hosts": []}));
    };
    let (id, dir) = new_receipt_dir(root)?;
    let mut receipt = Receipt {
        schema_version: 1,
        id: id.clone(),
        plan_digest: plan_digest.to_string(),
        permission_reference: permission.to_string(),
        status: "pending".into(),
        hosts: Vec::new(),
    };
    let receipt_path = dir.join("receipt.json");
    write_json(&receipt_path, &receipt)?;
    let mut results = Vec::new();
    for h in writes {
        let (files, dirs) = crate::hooks::written_files(root, h.host)?;
        plain_dirs(&dirs)?;
        let mut before = Vec::new();
        for (n, f) in files.iter().enumerate() {
            let bytes = crate::claude::read_existing(&root.join(f))?;
            if let Some(b) = &bytes {
                let keep = dir.join("before").join(h.host);
                fs::create_dir_all(&keep).map_err(|_| "hook receipt folder unavailable")?;
                fs::write(keep.join(n.to_string()), b).map_err(|_| "hook receipt write failed")?;
            }
            before.push(bytes.as_deref().map(crate::digest));
        }
        let digest = h
            .plan_digest
            .as_deref()
            .expect("a writing host has a digest");
        match crate::hooks::apply(root, h.host, Some(home), digest, permission) {
            Ok(v) => {
                let mut entries = Vec::new();
                for (f, b) in files.iter().zip(before) {
                    let after = crate::claude::read_existing(&root.join(f))?;
                    entries.push(FileEntry {
                        path: f.clone(),
                        before_sha256: b,
                        after_sha256: after.as_deref().map(crate::digest),
                    });
                }
                receipt.hosts.push(HostEntry {
                    host: h.host.to_string(),
                    files: entries,
                });
                write_json(&receipt_path, &receipt)?;
                results.push(json!({"host": h.host, "state": v["state"], "files_changed": v["files_changed"]}));
            }
            Err(e) => results.push(json!({"host": h.host, "state": "error", "message": e})),
        }
    }
    receipt.status = "applied".into();
    write_json(&receipt_path, &receipt)?;
    let applied = results
        .iter()
        .filter(|r| r["state"] == json!("applied"))
        .count();
    let state = match applied {
        0 if receipt.hosts.is_empty() => "error",
        0 => "unchanged",
        n if n == results.len() => "applied",
        _ => "partial",
    };
    Ok(json!({
        "state": state,
        "project": root.display().to_string(),
        "hosts": results,
        "receipt": id,
        "host_loaded": "not_run",
    }))
}

/// Undoes one confirmed hook layer for the hosts whose files all still hold the bytes the apply
/// wrote: each file gets its before bytes back, or is removed when the apply created it. A host
/// with any file changed since is skipped as a whole, so a settings file and its ownership
/// receipt never disagree.
pub fn revert(project: &Path, receipt_id: &str, permission: &str) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err("permission reference required".into());
    }
    if !receipt_id_ok(receipt_id) {
        return Err("unknown hook receipt id".into());
    }
    let root = fs::canonicalize(project).map_err(|_| "project directory unavailable")?;
    let dir = receipts_dir(&root).join(receipt_id);
    plain_dirs(&[
        crate::state_root::path(&root, ""),
        receipts_dir(&root),
        dir.clone(),
    ])?;
    let path = dir.join("receipt.json");
    let bytes = crate::claude::read_existing(&path)?.ok_or("hook receipt not found")?;
    let mut receipt: Receipt =
        serde_json::from_slice(&bytes).map_err(|_| "hook receipt is unreadable")?;
    if receipt.status == "reverted" {
        return Err("hook receipt already reverted".into());
    }
    let mut reverted = Vec::new();
    let mut skipped = Vec::new();
    for h in &receipt.hosts {
        let host = crate::hooks::HOSTS
            .iter()
            .find(|x| **x == h.host)
            .ok_or("hook receipt names an unknown host")?;
        let (files, dirs) = crate::hooks::written_files(&root, host)?;
        if h.files.iter().map(|f| &f.path).ne(files.iter()) {
            return Err("hook receipt names a file outside the host's hook files".into());
        }
        if let Err(e) = plain_dirs(&dirs) {
            skipped.push(json!({"host": host, "reason": format!("folder_not_plain: {e}")}));
            continue;
        }
        let mut current = Vec::new();
        for f in &h.files {
            current.push(crate::claude::read_existing(&root.join(&f.path))?);
        }
        let unchanged = h
            .files
            .iter()
            .zip(&current)
            .all(|(f, now)| now.as_deref().map(crate::digest) == f.after_sha256);
        if !unchanged {
            skipped.push(json!({"host": host, "reason": "user_changed_since_apply"}));
            continue;
        }
        // Read and check every before copy first, so a host is restored whole or not at all.
        let mut restore = Vec::new();
        for (n, f) in h.files.iter().enumerate() {
            let before = match &f.before_sha256 {
                None => None,
                Some(sha) => {
                    let b = fs::read(dir.join("before").join(host).join(n.to_string()))
                        .map_err(|_| "before bytes missing from the hook receipt")?;
                    if &crate::digest(&b) != sha {
                        return Err("before bytes do not match the hook receipt".into());
                    }
                    Some(b)
                }
            };
            restore.push(before);
        }
        for ((f, now), before) in h.files.iter().zip(&current).zip(restore) {
            let target = root.join(&f.path);
            match before {
                Some(b) => crate::apply::publish_replace(&target, now.as_deref(), &b)?,
                None if now.is_some() => fs::remove_file(&target)
                    .map_err(|_| format!("{} not removed", target.display()))?,
                None => {}
            }
        }
        reverted.push(host.to_string());
    }
    receipt.status = "reverted".into();
    write_json(&path, &receipt)?;
    Ok(
        json!({"state": "reverted", "receipt": receipt_id, "reverted": reverted, "skipped": skipped}),
    )
}
