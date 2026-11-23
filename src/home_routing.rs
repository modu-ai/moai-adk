//! The MoAI routing text in Claude Code's config folder (#2638, final design layer 2; user
//! decision 6845654776; SPEC-20261122-claude-md-home-routing, SPEC-20261122-claude-md-flow-rule).
//!
//! The routing text is one package file, `references/home/routing.md`, rendered for the host
//! (`{invoke}` becomes `/moai`) and written as the whole file `rules/moai.md` of a Claude config
//! folder (`CLAUDE_CONFIG_DIR`, else `~/.claude`); the flow rule `references/home/flow.md` is
//! written beside it as `rules/moai/flow.md`. Both are planned together for each config folder
//! (the resolved one, and every `--config-dir` the user picks) and shown in the preview. The
//! `moai-routing` block that an earlier version wrote into the folder's `CLAUDE.md` is no longer
//! refreshed; its removal is a separate pick (`remove_old_block`) and only happens while the block
//! still holds the bytes that version recorded and the routing file is in place. The Codex
//! section that an even earlier version wrote into `~/.codex/AGENTS.md` is no longer written
//! (SPEC-20261122-codex-home-update-removed); a receipt that names it is skipped.
//!
//! `plan` writes nothing. A whole file is moai's own only while it holds the bytes moai last
//! wrote (`home-routing/rules-<folder id>.json`); anything else, and any link or file that is
//! not UTF-8, is a `conflict` and is kept. Only a refused `CLAUDE_CONFIG_DIR` (or an unreadable
//! state file) blocks a folder. `apply` re-plans, refuses any other digest, keeps the before
//! bytes in a receipt whose entries name their kind and config folder; `revert` restores a file
//! only while it still holds the bytes the apply wrote, skips and names a file changed since,
//! and removes a folder the apply made only when it is empty. Hosts without a user-level
//! instruction file this module can write are reported in `not_supported` with the reason.

mod rules;

use crate::agents::{markers, FilePlan};
use crate::apply::{publish_exclusive, publish_replace, read_opt};
use crate::retired_hosts::retired;
use serde::{Deserialize, Serialize};
use std::{
    fs,
    path::{Path, PathBuf},
};

type Result<T> = std::result::Result<T, String>;

const SCHEMA_VERSION: u32 = 1;
const STATE_DIR: &str = "home-routing";
/// The managed block id the earlier version wrote into the global instruction file.
pub const SECTION: &str = "moai-routing";
/// The routing text inside the skill package.
pub const SOURCE: &str = "references/home/routing.md";
/// The flow rule text inside the skill package.
pub const FLOW_SOURCE: &str = "references/home/flow.md";
/// The routing file and the flow rule file, below a Claude config folder.
pub const RULES_FILE: &str = "rules/moai.md";
pub const FLOW_FILE: &str = "rules/moai/flow.md";
/// Hosts written: (host, the file of the earlier block under HOME when the host folder is not
/// relocated — see [`crate::host_home`] — and how the user invokes the skill there).
pub const HOSTS: [(&str, &str, &str); 1] = [("claude", ".claude/CLAUDE.md", "/moai")];
/// Hosts not written, each with the reason (#2635 6842183286).
pub const NOT_SUPPORTED: [(&str, &str); 2] = [
    (
        "kiro",
        "~/.kiro/steering is documented but not measured (Kiro not installed); not written",
    ),
    (
        "cursor",
        "Cursor CLI has no user-level instruction file (rules live in the app); the project AGENTS.md carries it",
    ),
];

/// The picks beyond the default config folder.
#[derive(Clone, Debug, Default)]
pub struct Options {
    /// Further Claude config folders (absolute), each planned like the default one.
    pub config_dirs: Vec<PathBuf>,
    /// Remove the earlier block from each folder's `CLAUDE.md` where it is unedited.
    pub remove_old_block: bool,
}

/// One config folder's plan. `section` is the routing file; when only the flow file changes it
/// carries that action, so a caller that reads this row alone still sees a write.
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct HostPlan {
    pub host: String,
    /// The config folder the row is about.
    pub config_dir: String,
    /// The routing file.
    pub file: String,
    /// ok | blocked (only a refused `CLAUDE_CONFIG_DIR` or an unreadable state file blocks)
    pub preflight: String,
    pub blocked_reason: Option<String>,
    /// The routing file's plan (create | merge | unchanged | conflict); none when blocked.
    pub section: Option<FilePlan>,
    /// The flow rule file's plan; none when blocked or the package has no flow text.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub flow: Option<FilePlan>,
    /// The earlier block in the folder's `CLAUDE.md`; none when there is no block to offer.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub old_block: Option<OldBlockPlan>,
}

/// What the plan says about the earlier block.
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct OldBlockPlan {
    pub path: String,
    /// offered | remove | waiting | kept
    pub action: String,
    pub expected_sha256: Option<String>,
    /// The block's sha256 the earlier version recorded.
    pub baseline_sha256: Option<String>,
    pub detail: Option<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct Unsupported {
    pub host: String,
    pub reason: String,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct Plan {
    pub schema_version: u32,
    pub home: String,
    pub package_digest: String,
    pub source_sha256: String,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub flow_source_sha256: Option<String>,
    pub hosts: Vec<HostPlan>,
    pub not_supported: Vec<Unsupported>,
    pub plan_digest: String,
}

/// The block the earlier version last wrote into one host's file.
#[derive(Clone, Debug, Default, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
struct HostState {
    schema_version: u32,
    block_sha256: Option<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct ReceiptEntry {
    pub host: String,
    /// routing | flow | old_block; an entry without one (an earlier version's) is the block.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub kind: Option<String>,
    /// The config folder the entry is about; an entry without one is the host's resolved folder.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub config_dir: Option<String>,
    /// create | append | merge | remove
    pub action: String,
    pub before_sha256: Option<String>,
    pub after_sha256: String,
    /// What this module owned before the apply (a file's or the block's digest); a revert
    /// restores it.
    pub baseline_before: Option<String>,
    /// The folders the apply made for this file, top down; a revert removes the empty ones.
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub created_dirs: Vec<String>,
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

fn sha(bytes: &[u8]) -> String {
    crate::digest(bytes)
}

fn state_root(home: &Path) -> PathBuf {
    crate::home_root::path(home).join(STATE_DIR)
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

/// The current bytes of a target file: `Ok(None)` when absent; a symlink or a folder in its
/// place is an error (never followed).
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

/// The record of the earlier block (`home-routing/<host>.json`): the baseline of its removal.
fn read_state(home: &Path, host: &str) -> Result<HostState> {
    let root = state_root(home);
    plain_dir_or_absent(&crate::home_root::path(home))?;
    plain_dir_or_absent(&root)?;
    match read_opt(&root.join(format!("{host}.json")))? {
        None => Ok(HostState {
            schema_version: SCHEMA_VERSION,
            block_sha256: None,
        }),
        Some(bytes) => {
            let state: HostState = serde_json::from_slice(&bytes)
                .map_err(|_| format!("routing state for {host} is unreadable; inspect it first"))?;
            if state.schema_version != SCHEMA_VERSION {
                return Err(format!("routing state for {host} has an unknown schema"));
            }
            Ok(state)
        }
    }
}

fn write_json<T: Serialize>(path: &Path, value: &T) -> Result<()> {
    let mut bytes = serde_json::to_vec_pretty(value).map_err(|_| "encode failed")?;
    bytes.push(b'\n');
    match read_opt(path)? {
        Some(b) => publish_replace(path, Some(&b), &bytes),
        None => publish_exclusive(path, &bytes),
    }
}

/// The routing source, checked: UTF-8, one `{invoke}`, no marker tokens, ends with a newline.
fn source(package: &Path) -> Result<String> {
    let path = package.join(SOURCE);
    let text = match fs::symlink_metadata(&path) {
        Ok(m) if m.is_file() => {
            fs::read_to_string(&path).map_err(|_| format!("{SOURCE} is not readable UTF-8"))?
        }
        _ => return Err(format!("{SOURCE} missing from the package")),
    };
    if text.matches("{invoke}").count() != 1
        || ["<!--", "-->", "moai:managed:"]
            .iter()
            .any(|t| text.contains(t))
        || !text.ends_with('\n')
    {
        return Err(format!(
            "{SOURCE} must hold one {{invoke}}, no marker tokens and end with a newline"
        ));
    }
    Ok(text)
}

/// The routing file's text for one host: the source with `{invoke}` filled in.
pub fn render(source: &str, host: &str) -> Result<String> {
    let invoke = HOSTS
        .iter()
        .find(|(h, _, _)| *h == host)
        .map(|(_, _, i)| *i)
        .ok_or_else(|| format!("routing host not supported: {host:?}"))?;
    Ok(source.replacen("{invoke}", invoke, 1))
}

/// Read-only preview of every host; writes nothing.
pub fn plan(home: &Path, package: &Path) -> Result<Plan> {
    let all: Vec<&str> = HOSTS.iter().map(|(h, _, _)| *h).collect();
    plan_for(home, package, &all)
}

/// Read-only preview of the named hosts only (#2638: `moai-cli update` leaves out a host whose
/// skill copy is blocked); writes nothing. An unknown host is refused.
pub fn plan_for(home: &Path, package: &Path, only: &[&str]) -> Result<Plan> {
    plan_with(home, package, only, &Options::default())
}

/// [`plan_for`] with the picks: further config folders and the removal of the earlier block.
pub fn plan_with(home: &Path, package: &Path, only: &[&str], opts: &Options) -> Result<Plan> {
    plan_inner(home, package, only, opts).map(|(plan, _)| plan)
}

fn plan_inner(
    home: &Path,
    package: &Path,
    only: &[&str],
    opts: &Options,
) -> Result<(Plan, Vec<rules::Folder>)> {
    if let Some(h) = only.iter().find(|h| !HOSTS.iter().any(|(k, _, _)| k == *h)) {
        return Err(format!("routing host not supported: {h:?}"));
    }
    if opts.config_dirs.iter().any(|d| !d.is_absolute()) {
        return Err("--config-dir must be an absolute path".into());
    }
    let home = fs::canonicalize(home).map_err(|_| "home directory unavailable")?;
    // No host to write: the sources are not needed (a package from before #2638 has none).
    let (src, flow) = if only.is_empty() {
        (String::new(), None)
    } else {
        (
            rules::routing_source(package)?,
            rules::flow_source(package)?,
        )
    };
    let mut folders = Vec::new();
    if only.contains(&"claude") {
        let routing_text = render(&src, "claude")?;
        let mut dirs = vec![(crate::host_home::resolve_named("claude", &home), true)];
        for d in &opts.config_dirs {
            if !dirs.iter().any(|(r, _)| r.as_ref().ok() == Some(d)) {
                dirs.push((Ok(d.clone()), false));
            }
        }
        for (dir, is_default) in dirs {
            folders.push(rules::plan_folder(
                &home,
                dir,
                is_default,
                &routing_text,
                flow.as_deref(),
                opts.remove_old_block,
            ));
        }
    }
    let mut plan = Plan {
        schema_version: SCHEMA_VERSION,
        home: home.to_str().ok_or("non UTF-8 home path")?.to_string(),
        package_digest: crate::package_digest(package)?,
        source_sha256: sha(src.as_bytes()),
        flow_source_sha256: flow.as_ref().map(|f| sha(f.as_bytes())),
        hosts: folders.iter().map(|f| f.row.clone()).collect(),
        not_supported: NOT_SUPPORTED
            .iter()
            .map(|(h, r)| Unsupported {
                host: (*h).into(),
                reason: (*r).into(),
            })
            .collect(),
        plan_digest: String::new(),
    };
    plan.plan_digest = sha(&serde_json::to_vec(&plan).map_err(|_| "encode failed")?);
    Ok((plan, folders))
}

fn validate_permission(permission: &str) -> Result<()> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err("permission reference required".into());
    }
    Ok(())
}

fn new_receipt_dir(home: &Path) -> Result<(String, PathBuf)> {
    plain_dir_or_absent(&crate::home_root::path(home))?;
    let receipts = state_root(home).join("receipts");
    fs::create_dir_all(&receipts).map_err(|_| "routing receipt folder unavailable")?;
    plain_dir_or_absent(&state_root(home))?;
    plain_dir_or_absent(&receipts)?;
    let nanos = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_nanos())
        .unwrap_or(0);
    for n in 0..16 {
        let id = format!("routing-{nanos}-{n}");
        let dir = receipts.join(&id);
        if fs::create_dir(&dir).is_ok() {
            return Ok((id, dir));
        }
    }
    Err("routing receipt id unavailable".into())
}

/// Applies a previewed plan. `plan_digest` must equal a fresh plan's digest; only `create` and
/// `merge` files and a picked block removal are written; blocked folders and conflicts are
/// skipped.
pub fn apply(
    home: &Path,
    package: &Path,
    plan_digest: &str,
    permission: &str,
) -> Result<serde_json::Value> {
    let all: Vec<&str> = HOSTS.iter().map(|(h, _, _)| *h).collect();
    apply_for(home, package, &all, plan_digest, permission)
}

/// [`apply`] for the named hosts only; `plan_digest` must equal a fresh [`plan_for`] of the
/// same hosts.
pub fn apply_for(
    home: &Path,
    package: &Path,
    only: &[&str],
    plan_digest: &str,
    permission: &str,
) -> Result<serde_json::Value> {
    apply_with(
        home,
        package,
        only,
        &Options::default(),
        plan_digest,
        permission,
    )
}

/// [`apply_for`] with the picks the preview was made with.
pub fn apply_with(
    home: &Path,
    package: &Path,
    only: &[&str],
    opts: &Options,
    plan_digest: &str,
    permission: &str,
) -> Result<serde_json::Value> {
    validate_permission(permission)?;
    let (fresh, folders) = plan_inner(home, package, only, opts)?;
    if fresh.plan_digest != plan_digest {
        return Err("stale or altered plan; re-preview required".into());
    }
    let home = PathBuf::from(&fresh.home);
    if !folders.iter().any(rules::Folder::writes) {
        return Ok(
            serde_json::json!({"state": "unchanged", "plan_digest": fresh.plan_digest, "written": []}),
        );
    }
    let routing_text = render(&rules::routing_source(package)?, "claude")?;
    let flow_text = rules::flow_source(package)?;
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
    for folder in &folders {
        written.extend(rules::apply_folder(
            &home,
            &receipt_dir,
            &mut receipt,
            folder,
            &routing_text,
            flow_text.as_deref(),
        )?);
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
    id.starts_with("routing-")
        && id.len() <= 64
        && id.bytes().all(|b| b.is_ascii_alphanumeric() || b == b'-')
}

/// Undoes one apply for the files that still hold the bytes it wrote: a created file is
/// removed, a changed one gets its before bytes back. A file changed since is skipped and named.
pub fn revert(home: &Path, receipt_id: &str, permission: &str) -> Result<serde_json::Value> {
    revert_in(
        &HOSTS.iter().map(|(h, _, _)| *h).collect::<Vec<_>>(),
        home,
        receipt_id,
        permission,
    )
}

/// [`revert`] for a place whose active hosts are `active`: a receipt entry of a retired host is
/// skipped and reported (reason `host_retired`), the other entries are reverted last first.
/// `reverted` names each host once; `files` names each file restored.
pub fn revert_in(
    active: &[&str],
    home: &Path,
    receipt_id: &str,
    permission: &str,
) -> Result<serde_json::Value> {
    validate_permission(permission)?;
    if !receipt_id_ok(receipt_id) {
        return Err("unknown routing receipt id".into());
    }
    let home = fs::canonicalize(home).map_err(|_| "home directory unavailable")?;
    plain_dir_or_absent(&crate::home_root::path(&home))?;
    plain_dir_or_absent(&state_root(&home))?;
    let dir = state_root(&home).join("receipts").join(receipt_id);
    plain_dir_or_absent(&dir)?;
    let path = dir.join("receipt.json");
    let bytes = read_opt(&path)?.ok_or("routing receipt not found")?;
    let mut receipt: Receipt =
        serde_json::from_slice(&bytes).map_err(|_| "routing receipt is unreadable")?;
    if receipt.status == "reverted" {
        return Err("routing receipt already reverted".into());
    }
    let mut reverted: Vec<String> = Vec::new();
    let mut files = Vec::new();
    let mut skipped = Vec::new();
    for e in receipt.entries.iter().rev() {
        if retired(&e.host, active) {
            skipped.push(serde_json::json!({"host": e.host, "reason": "host_retired"}));
            continue;
        }
        if !HOSTS.iter().any(|(h, _, _)| *h == e.host) {
            return Err("routing receipt names an unknown host".into());
        }
        match rules::revert_entry(&home, &dir, e)? {
            Ok(file) => {
                if !reverted.contains(&e.host) {
                    reverted.push(e.host.clone());
                }
                files.push(serde_json::json!({
                    "host": e.host,
                    "kind": e.kind.as_deref().unwrap_or("old_block"),
                    "path": file,
                }));
            }
            Err(skip) => skipped.push(skip),
        }
    }
    receipt.status = "reverted".into();
    write_json(&path, &receipt)?;
    Ok(
        serde_json::json!({"state": "reverted", "receipt": receipt_id, "reverted": reverted, "files": files, "skipped": skipped}),
    )
}
