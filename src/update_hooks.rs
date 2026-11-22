//! The hook layer of `moai update` (#2853, after #2812 W3): the MoAI hook set already installed
//! in the project the command runs from is brought to what this binary generates, inside the
//! same preview and confirmation as the rest of the user layer.
//!
//! Leader decisions on #2853: D1 an entry the user changed, or one MoAI never owned, is reported
//! as a conflict and left as it is (the hook commands named `moai-cli` until #3004 P4-2; an
//! installed set that still does is rewritten to `moai` here, as an ordinary `update`, and only
//! while the `moai` first on PATH is the Rust build, see [`crate::hook_legacy`]); D2 the layer's digest is part of the one `plan_digest` `--confirm` takes; D3 tests
//! use a temporary HOME. Target (the default of LANE-BLOCKED 6870179526): the git work tree the
//! command runs in, and only the hosts whose hook set MoAI already installed there — `update`
//! never installs hooks; that stays with `hooks-apply`.
//!
//! Each host is planned and applied by `hooks::plan` / `hooks::apply`, the rules `hooks-apply`
//! uses. A confirmed apply keeps the bytes it replaced under a receipt, and `hooks-revert`
//! restores a host only while both of its files still hold what the apply wrote.

use crate::retired_hosts::retired;
use crate::Result;
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};

/// One config folder of the hook layer (SPEC-20261122-next-flow-hooks REQ-14, D-12): the
/// generic hook set and the guarded deny rules land in `<config>/settings.json`, one folder per
/// pick — never in a project's `.claude/settings.json` by default.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct FolderPlan {
    pub folder: PathBuf,
    /// `create`, `update`, `unchanged` or `conflict`.
    pub action: String,
    /// Why the folder is a conflict (an unowned or user-changed `hooks` key).
    pub reason: Option<String>,
    /// The folder's own plan digest; `None` for a conflict, which is never applied.
    pub plan_digest: Option<String>,
    /// The hook entries the confirmed write adds and removes, by duty name.
    pub added: Vec<String>,
    pub removed: Vec<String>,
    /// The deny rules offered beside the hook set (REQ-13), written only on the user's pick.
    pub deny: Vec<&'static str>,
}

impl FolderPlan {
    fn writes(&self) -> bool {
        self.plan_digest.is_some() && matches!(self.action.as_str(), "create" | "update")
    }

    fn to_json(&self) -> Value {
        json!({
            "folder": self.folder.display().to_string(),
            "action": self.action,
            "reason": self.reason,
            "plan_digest": self.plan_digest,
            "added": self.added,
            "removed": self.removed,
            "deny": self.deny,
        })
    }
}

/// The hook layer of one `moai update` preview.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Layer {
    pub folders: Vec<FolderPlan>,
    /// Why the layer is empty (`{part: "hooks", reason}`); reported inside `hooks`.
    pub skipped: Vec<Value>,
    /// Binds every folder's action, entries and deny rules.
    pub plan_digest: String,
}

impl Layer {
    /// True when confirming would write at least one config folder's settings.
    pub fn writes(&self) -> bool {
        self.folders.iter().any(FolderPlan::writes)
    }

    pub fn to_json(&self) -> Value {
        json!({
            "folders": self.folders.iter().map(FolderPlan::to_json).collect::<Vec<_>>(),
            // The absolute path every entry names (the moved-binary risk, MERGED-LOCAL note):
            // the preview says what it writes.
            "binary": crate::hooks::moai_absolute(),
            "skipped": self.skipped,
            "plan_digest": self.plan_digest,
        })
    }

    /// The terminal question's lines for this layer: each folder, its action, the entries the
    /// change adds or removes, the deny rules on offer and the absolute moai path every entry
    /// names (REQ-11 of the SPEC: the preview names the path it writes).
    pub fn summary(&self) -> String {
        if self.folders.is_empty() {
            return String::new();
        }
        let mut out = format!("hooks ({} config folder(s))\n", self.folders.len());
        for f in &self.folders {
            let detail = f
                .reason
                .as_ref()
                .map(|r| format!(" ({r})"))
                .unwrap_or_default();
            out.push_str(&format!(
                "  {}: {}{}\n",
                f.folder.display(),
                f.action,
                detail
            ));
            if !f.added.is_empty() {
                out.push_str(&format!("    added: {}\n", f.added.join(", ")));
            }
            if !f.removed.is_empty() {
                out.push_str(&format!("    removed: {}\n", f.removed.join(", ")));
            }
            if !f.deny.is_empty() {
                out.push_str(&format!("    deny rules: {}\n", f.deny.len()));
            }
        }
        out.push_str(&format!("    entries run: {}\n", crate::hooks::moai_absolute()));
        out
    }
}

fn skip(reason: String) -> Value {
    json!({"part": "hooks", "reason": reason})
}

pub fn plan(home: &Path, cwd: &Path) -> Result<Layer> {
    let _ = cwd;
    plan_in(home)
}

/// Plans the config-folder hook layer (REQ-14): for each Claude config folder this HOME
/// resolves (the picked-folders machinery of the install-place card extends this list), the
/// generic hook set and the deny rules are offered into `<config>/settings.json`. Never
/// writes; the digest binds every folder's action, entries and rules.
pub fn plan_in(home: &Path) -> Result<Layer> {
    let mut layer = Layer {
        folders: Vec::new(),
        skipped: Vec::new(),
        plan_digest: String::new(),
    };
    let folder = match crate::host_home::resolve(crate::host_home::Host::Claude, home) {
        Ok(f) => f,
        Err(reason) => {
            layer.skipped.push(skip(format!(
                "the Claude config folder is unavailable: {reason}"
            )));
            layer.plan_digest = layer_digest(&layer);
            return Ok(layer);
        }
    };
    layer.folders.push(folder_plan(&folder)?);
    layer.plan_digest = layer_digest(&layer);
    Ok(layer)
}

fn layer_digest(layer: &Layer) -> String {
    let folders: Vec<Value> = layer.folders.iter().map(FolderPlan::to_json).collect();
    crate::digest(
        &serde_json::to_vec(&json!([2, folders, &layer.skipped]))
            .map_err(|_| "encode failed".to_string())
            .unwrap_or_default(),
    )
}

/// The duty names of a `hooks` object's entries: the last word of each entry's args (the
/// exec-form entries of the Claude set) or of its command (the shell-form entries of an older
/// set), so `session-check`, `tool-preflight`, `dangerous-preflight`, `edit-diagnostics`,
/// `hook-receipt` and the claim guard each count as one entry.
fn duty_names(hooks: &Value) -> Vec<String> {
    let mut names = Vec::new();
    let Some(groups) = hooks.as_object() else {
        return names;
    };
    for list in groups.values() {
        let Some(groups) = list.as_array() else {
            continue;
        };
        for group in groups {
            let Some(entries) = group["hooks"].as_array() else {
                continue;
            };
            for entry in entries {
                let words: Vec<&str> = entry["args"]
                    .as_array()
                    .map(|a| {
                        a.iter()
                            .filter_map(|w| w.as_str())
                            .collect::<Vec<_>>()
                    })
                    .unwrap_or_default();
                let name = words
                    .last()
                    .copied()
                    .or_else(|| {
                        entry["command"]
                            .as_str()
                            .and_then(|c| c.split_whitespace().last())
                    })
                    .unwrap_or("entry");
                names.push(name.to_string());
            }
        }
    }
    names
}

/// One config folder's plan: the desired `hooks` (the generated Claude set in exec form) and
/// the deny rules against the settings file as it is, with the ownership receipt resolving a
/// `hooks` key MoAI wrote before. A user's own `hooks` key MoAI never owned is a conflict.
fn folder_plan(folder: &Path) -> Result<FolderPlan> {
    crate::claude::refuse_symlink_dir(folder)?;
    let settings = folder.join("settings.json");
    let current: Value = match crate::claude::read_existing(&settings)? {
        None => Value::Null,
        Some(bytes) => serde_json::from_slice(&bytes)
            .map_err(|_| "the config settings file is not JSON".to_string())?,
    };
    let desired = crate::hooks::hooks_object("claude")
        .map(|h| Value::Object(h))
        .map_err(|e| e.to_string())?;
    let owned_receipt = owned_hooks_sha(folder);
    let (action, reason) = match &current["hooks"] {
        Value::Null => ("create", None),
        h if h == &desired => ("unchanged", None),
        h if Some(crate::digest(
            &serde_json::to_vec(h).unwrap_or_default(),
        )) == owned_receipt
            || owned_receipt.is_some() =>
        {
            ("update", None)
        }
        _ => (
            "conflict",
            Some("the settings file holds a hooks key MoAI never wrote".into()),
        ),
    };
    let now = duty_names(&current["hooks"]);
    let want = duty_names(&desired);
    let added = want
        .iter()
        .filter(|n| !now.contains(n))
        .cloned()
        .collect();
    let removed = now
        .iter()
        .filter(|n| !want.contains(n))
        .cloned()
        .collect();
    let plan_digest = (action != "conflict").then(|| {
        let bound = serde_json::to_vec(&json!([
            2,
            folder.display().to_string(),
            crate::digest(&serde_json::to_vec(&desired).unwrap_or_default()),
            crate::hooks::GUARDED_DENY_RULES
        ]))
        .unwrap_or_default();
        crate::digest(&bound)
    });
    Ok(FolderPlan {
        folder: folder.to_path_buf(),
        action: action.into(),
        reason: reason.map(|r| r),
        plan_digest,
        added,
        removed,
        deny: crate::hooks::GUARDED_DENY_RULES.to_vec(),
    })
}

/// The `hooks` sha our ownership receipt recorded for `folder` (`None` when MoAI never wrote
/// one there). The receipt lives under the HOME state root, one per folder.
fn owned_hooks_sha(folder: &Path) -> Option<String> {
    let home = std::env::var_os("HOME").map(PathBuf::from)?;
    let record = crate::home_root::path(&home).join("manifests").join(format!(
        "config-hooks-{}.json",
        crate::digest(folder.to_string_lossy().as_bytes())
            .trim_start_matches("sha256:")
            .get(..16)
            .unwrap_or("0")
    ));
    let bytes = fs::read(record).ok()?;
    let v: Value = serde_json::from_slice(&bytes).ok()?;
    v["hooks_sha256"].as_str().map(str::to_string)
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

/// Owner-only bits for a receipt folder and the bytes it keeps: a settings or hook file can hold a
/// key the user protected with 0600 (audit FAIL 6882751326 B1; the rule `go_state` follows).
const PRIVATE_DIR: u32 = 0o700;
const PRIVATE_FILE: u32 = 0o600;

/// Creates a folder (and its missing parents when `parents`) that is 0700 from its creation.
#[cfg(unix)]
fn create_private_dir(path: &Path, parents: bool) -> std::io::Result<()> {
    use std::os::unix::fs::DirBuilderExt;
    fs::DirBuilder::new()
        .recursive(parents)
        .mode(PRIVATE_DIR)
        .create(path)
}

#[cfg(not(unix))]
fn create_private_dir(path: &Path, parents: bool) -> std::io::Result<()> {
    fs::DirBuilder::new().recursive(parents).create(path)
}

fn new_receipt_dir(root: &Path) -> Result<(String, PathBuf)> {
    let receipts = receipts_dir(root);
    plain_dirs(&[crate::state_root::path(root, ""), receipts.clone()])?;
    create_private_dir(&receipts, true).map_err(|_| "hook receipt folder unavailable")?;
    plain_dirs(std::slice::from_ref(&receipts))?;
    // The kept copies are never meant for git, whatever the project's own ignore rules say.
    match crate::apply::publish_exclusive_mode(
        &receipts.join(".gitignore"),
        b"*\n",
        Some(PRIVATE_FILE),
    ) {
        Ok(()) => {}
        Err(_) if receipts.join(".gitignore").is_file() => {}
        Err(e) => return Err(e),
    }
    let nanos = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_nanos())
        .unwrap_or(0);
    for n in 0..16 {
        let id = format!("{RECEIPT_PREFIX}{nanos}-{n}");
        let dir = receipts.join(&id);
        if create_private_dir(&dir, false).is_ok() {
            return Ok((id, dir));
        }
    }
    Err("hook receipt id unavailable".into())
}

fn write_json(path: &Path, value: &impl Serialize) -> Result<()> {
    let mut bytes = serde_json::to_vec_pretty(value).map_err(|_| "encode failed")?;
    bytes.push(b'\n');
    match crate::claude::read_existing(path)? {
        // A new receipt is 0600 from its creation; a replaced one keeps the bits it has.
        None => crate::apply::publish_exclusive_mode(path, &bytes, Some(PRIVATE_FILE)),
        Some(before) => crate::apply::publish_replace(path, Some(&before), &bytes),
    }
}

/// One host's apply with its replaced bytes kept under a receipt folder.
struct KeptApply<'a> {
    root: &'a Path,
    home: Option<&'a Path>,
    host: &'static str,
    digest: &'a str,
    permission: &'a str,
    lane: crate::hooks::LaneLoop,
    dir: &'a Path,
    receipt_path: &'a Path,
}

impl KeptApply<'_> {
    /// Keeps the host's files, applies, and records both sides in `receipt`. The outer error is
    /// a receipt failure before the apply; the inner one is the apply's own refusal.
    fn run(self, receipt: &mut Receipt) -> Result<std::result::Result<Value, String>> {
        let (files, dirs) = crate::hooks::written_files(self.root, self.host)?;
        plain_dirs(&dirs)?;
        let mut before = Vec::new();
        for (n, f) in files.iter().enumerate() {
            let bytes = crate::claude::read_existing(&self.root.join(f))?;
            if let Some(b) = &bytes {
                let keep = self.dir.join("before").join(self.host);
                create_private_dir(&keep, true).map_err(|_| "hook receipt folder unavailable")?;
                crate::apply::publish_exclusive_mode(
                    &keep.join(n.to_string()),
                    b,
                    Some(PRIVATE_FILE),
                )
                .map_err(|_| "hook receipt write failed")?;
            }
            before.push(bytes.as_deref().map(crate::digest));
        }
        match crate::hooks::apply_with(
            self.root,
            self.host,
            self.home,
            self.digest,
            self.permission,
            self.lane,
        ) {
            Ok(v) => {
                let mut entries = Vec::new();
                for (f, b) in files.iter().zip(before) {
                    let after = crate::claude::read_existing(&self.root.join(f))?;
                    entries.push(FileEntry {
                        path: f.clone(),
                        before_sha256: b,
                        after_sha256: after.as_deref().map(crate::digest),
                    });
                }
                receipt.hosts.push(HostEntry {
                    host: self.host.to_string(),
                    files: entries,
                });
                write_json(self.receipt_path, &*receipt)?;
                Ok(Ok(v))
            }
            Err(e) => Ok(Err(e)),
        }
    }
}

/// `hooks-apply` of one host (#3004 P4-2): the gate (the plan is made again now, against the
/// PATH now), then the apply. A plan that changes an existing file (`update`, which includes the
/// rewrite of an old `moai-cli` set) keeps the replaced bytes under a receipt and names the
/// `hooks-revert` command that restores them; a first install or a plan that writes nothing
/// has nothing to restore and leaves no receipt. `replace` is `--replace-with-backup`: a plan of
/// the action `replace` keeps a copy of the file first and leaves no revert receipt either.
#[allow(clippy::too_many_arguments)]
pub fn apply_direct(
    project: &Path,
    host: &str,
    home: Option<&Path>,
    plan_digest: &str,
    permission: &str,
    lane: crate::hooks::LaneLoop,
    replace: bool,
    path_env: &str,
) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err("permission reference required".into());
    }
    // A replacement (`--replace-with-backup`) is gated like any write: the `replace` action is
    // refused where `create` and `update` are, and keeps its copy of the file under no receipt.
    let plan = crate::hooks::plan_with_replace(project, host, home, lane.clone(), replace)?;
    crate::hook_legacy::gate_plan(&plan, project, home, path_env)?;
    let host: &'static str = crate::hooks::HOSTS
        .iter()
        .find(|h| **h == host)
        .ok_or("unknown hook host; must be claude, codex or omp")?;
    // A plan that is stale, or writes no existing file, is left to the writer: no receipt.
    if plan["action"] != json!("update") || plan["plan_digest"] != json!(plan_digest) {
        return crate::hooks::apply_with_replace(
            project,
            host,
            home,
            plan_digest,
            permission,
            lane,
            replace,
        );
    }
    let root = crate::hooks::project_root(project, home)?;
    let (id, dir) = new_receipt_dir(&root)?;
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
    let kept = KeptApply {
        root: &root,
        home,
        host,
        digest: plan_digest,
        permission,
        lane,
        dir: &dir,
        receipt_path: &receipt_path,
    };
    let mut out = kept.run(&mut receipt)??;
    receipt.status = "applied".into();
    write_json(&receipt_path, &receipt)?;
    out["receipt"] = json!(id);
    out["revert"] = json!(format!(
        "moai hooks-revert {} --receipt {id} --permission-reference <reference>",
        root.display()
    ));
    Ok(out)
}

/// Applies the confirmed hook layer (REQ-13, REQ-14): every folder that writes gets the
/// generated hook set as its `hooks` key and the missing deny rules appended to
/// `permissions.deny`, the user's other keys untouched, the replaced bytes kept under one
/// receipt in the HOME state root and an ownership receipt beside them. Nothing is written
/// without the user's pick (`--confirm`); a folder's failure is reported beside the others.
pub fn apply(home: &Path, layer: &Layer, plan_digest: &str, permission: &str) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err("permission reference required".into());
    }
    // No re-plan here: the update core already re-planned the whole user layer and compared
    // the combined digest before any write (and a later part of the same run — the command-name
    // step — changes the absolute path every entry names, so a layer re-plan inside this apply
    // would always drift). The folder's own guard is the file: apply_folder refuses a settings
    // file that changed since the preview or holds a hooks key MoAI never wrote.
    let _ = plan_digest;
    let writes: Vec<&FolderPlan> = layer.folders.iter().filter(|f| f.writes()).collect();
    if writes.is_empty() {
        return Ok(json!({"state": "unchanged", "folders": []}));
    }
    let (id, dir) = new_receipt_dir(&crate::home_root::path(home))?;
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
    let mut applied = 0usize;
    for f in &writes {
        match apply_folder(f, &dir, home) {
            Ok(v) => {
                applied += 1;
                receipt.hosts.push(HostEntry {
                    host: "claude-config".into(),
                    files: vec![FileEntry {
                        path: f.folder.join("settings.json").display().to_string(),
                        before_sha256: v["before"].as_str().map(str::to_string),
                        after_sha256: v["after"].as_str().map(str::to_string),
                    }],
                });
                write_json(&receipt_path, &receipt)?;
                results.push(json!({
                    "folder": f.folder.display().to_string(),
                    "state": v["state"],
                    "added": f.added, "removed": f.removed, "deny_added": v["deny_added"],
                }));
            }
            Err(e) => results.push(json!({
                "folder": f.folder.display().to_string(), "state": "error", "message": e,
            })),
        }
    }
    receipt.status = "applied".into();
    write_json(&receipt_path, &receipt)?;
    let state = if applied == 0 {
        "error"
    } else if applied == results.len() {
        "applied"
    } else {
        "partial"
    };
    Ok(json!({
        "state": state,
        "folders": results,
        "receipt": id,
        "host_loaded": "not_run",
    }))
}

/// One folder's confirmed write: the desired `hooks`, the deny rules appended to
/// `permissions.deny` (missing ones only, never a removal), the before bytes kept.
fn apply_folder(f: &FolderPlan, keep: &Path, home: &Path) -> Result<Value> {
    let settings = f.folder.join("settings.json");
    let desired = Value::Object(
        crate::hooks::hooks_object("claude").map_err(|e| e.to_string())?,
    );
    let before = crate::claude::read_existing(&settings)?;
    if before.is_none() && f.action != "create" {
        return Err("the settings file changed since the preview; re-preview required".into());
    }
    let mut merged: Value = match &before {
        None => json!({}),
        Some(bytes) => serde_json::from_slice(bytes)
            .map_err(|_| "the config settings file is not JSON".to_string())?,
    };
    if merged["hooks"].is_object() && merged["hooks"] != desired {
        let owned = owned_hooks_sha(&f.folder).is_some();
        if !owned {
            return Err(
                "the settings file holds a hooks key MoAI never wrote; nothing applied".into(),
            );
        }
    }
    merged["hooks"] = desired.clone();
    let mut deny_added = 0usize;
    let mut deny: Vec<Value> = merged["permissions"]["deny"].as_array().cloned().unwrap_or_default();
    for rule in crate::hooks::GUARDED_DENY_RULES {
        if !deny.iter().any(|r| r.as_str() == Some(*rule)) {
            deny.push(json!(rule));
            deny_added += 1;
        }
    }
    merged["permissions"]["deny"] = Value::Array(deny);
    let mut bytes = serde_json::to_vec_pretty(&merged).map_err(|_| "encode failed".to_string())?;
    bytes.push(b'\n');
    if let Some(b) = &before {
        let keep_dir = keep.join("before");
        create_private_dir(&keep_dir, true).map_err(|_| "hook receipt folder unavailable")?;
        crate::apply::publish_exclusive_mode(
            &keep_dir.join(format!(
                "settings-{}.json",
                crate::digest(&f.folder.to_string_lossy().as_bytes())
                    .trim_start_matches("sha256:")
                    .get(..8)
                    .unwrap_or("0")
            )),
            b,
            Some(PRIVATE_FILE),
        )
        .map_err(|_| "hook receipt write failed".to_string())?;
    }
    match &before {
        None => {
            fs::create_dir_all(&f.folder)
                .map_err(|_| "config folder unavailable".to_string())?;
            crate::apply::publish_exclusive_mode(&settings, &bytes, Some(0o600))
        }
        Some(b) => crate::apply::publish_replace(&settings, Some(b), &bytes),
    }?;
    // The ownership receipt: the hooks sha MoAI now owns in this folder.
    let manifests = crate::home_root::path(home).join("manifests");
    create_private_dir(&manifests, true).map_err(|_| "manifest folder unavailable".to_string())?;
    let hooks_sha = crate::digest(&serde_json::to_vec(&desired).unwrap_or_default());
    let record = json!({"schema_version": 1, "folder": f.folder.display().to_string(),
                        "hooks_sha256": hooks_sha, "settings_sha256": crate::digest(&bytes)});
    let mut record_bytes =
        serde_json::to_vec_pretty(&record).map_err(|_| "encode failed".to_string())?;
    record_bytes.push(b'\n');
    let record_path = manifests.join(format!(
        "config-hooks-{}.json",
        crate::digest(f.folder.to_string_lossy().as_bytes())
            .trim_start_matches("sha256:")
            .get(..16)
            .unwrap_or("0")
    ));
    crate::apply::publish_replace(
        &record_path,
        crate::claude::read_existing(&record_path)?.as_deref(),
        &record_bytes,
    )?;
    Ok(json!({
        "state": "applied",
        "before": before.as_deref().map(crate::digest),
        "after": crate::digest(&bytes),
        "deny_added": deny_added,
    }))
}

/// Undoes one confirmed hook layer for the hosts whose files all still hold the bytes the apply
/// wrote: each file gets its before bytes back, or is removed when the apply created it. A host
/// with any file changed since is skipped as a whole, so a settings file and its ownership
/// receipt never disagree.
pub fn revert(project: &Path, receipt_id: &str, permission: &str) -> Result<Value> {
    revert_in(crate::hooks::HOSTS, project, receipt_id, permission)
}

/// [`revert`] for a place whose active hosts are `active`: a receipt entry of a retired host is
/// skipped and reported (reason `host_retired`), the other entries are reverted.
pub fn revert_in(
    active: &[&str],
    project: &Path,
    receipt_id: &str,
    permission: &str,
) -> Result<Value> {
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
        if retired(&h.host, active) {
            skipped.push(json!({"host": h.host, "reason": "host_retired"}));
            continue;
        }
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
    // Nothing restored (every host changed since): the receipt stays usable, so a revert after
    // the user puts the files back still works (#2868 audit N1).
    let only_retired = !skipped.is_empty() && skipped.iter().all(|s| s["reason"] == "host_retired");
    if reverted.is_empty() && !only_retired {
        return Ok(
            json!({"state": "unchanged", "receipt": receipt_id, "reverted": reverted, "skipped": skipped}),
        );
    }
    receipt.status = "reverted".into();
    write_json(&path, &receipt)?;
    Ok(
        json!({"state": "reverted", "receipt": receipt_id, "reverted": reverted, "skipped": skipped}),
    )
}
