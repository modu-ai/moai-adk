//! The MoAI routing section in each host's global instruction file (#2638, final design
//! layer 2; user decision 6845654776).
//!
//! The section text is one package file, `references/home/routing.md`, rendered per host
//! (`{invoke}` becomes `/moai` or `$moai`) and kept between `moai:managed` markers with the id
//! `moai-routing` in Claude Code's `~/.claude/CLAUDE.md` and Codex's `~/.codex/AGENTS.md`. Codex
//! has no `@` import (#2635 6842183286, measured), so both files carry the text itself, from
//! the same source. `~/.agents/AGENTS.md` is never created (user decision 3).
//!
//! `plan` compares the block three-way with `agents::section_plan` — baseline = the block we
//! last wrote (`~/.moai-cli/home-routing/<host>.json`), current = the block on disk, new = the
//! rendering — and writes nothing. Bytes outside the markers are never changed; an edited
//! block or malformed markers are a `conflict`, never overwritten; a symlinked folder or file
//! blocks the host. `apply` re-plans, refuses any other digest, keeps the before bytes in a
//! receipt; `revert` restores a file only while it still holds the bytes the apply wrote.
//! Hosts without a user-level instruction file this module can write are reported in
//! `not_supported` with the reason — never served by a substitute.

use crate::agents::{assert_single_block, markers, section_plan, splice, FilePlan};
use crate::apply::{publish_exclusive, publish_replace, read_opt};
use serde::{Deserialize, Serialize};
use std::{
    fs,
    path::{Path, PathBuf},
};

type Result<T> = std::result::Result<T, String>;

const SCHEMA_VERSION: u32 = 1;
const STATE_DIR: &str = "home-routing";
/// The managed block id inside the global instruction files.
pub const SECTION: &str = "moai-routing";
/// The section source inside the skill package.
pub const SOURCE: &str = "references/home/routing.md";
/// Hosts written: (host, file under HOME, how the user invokes the skill there).
pub const HOSTS: [(&str, &str, &str); 2] = [
    ("claude", ".claude/CLAUDE.md", "/moai"),
    ("codex", ".codex/AGENTS.md", "$moai"),
];
/// Hosts not written, each with the reason (#2635 6842183286). The host restored by #2647 is
/// added there, together with its agent rendering.
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

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct HostPlan {
    pub host: String,
    pub file: String,
    /// ok | blocked
    pub preflight: String,
    pub blocked_reason: Option<String>,
    /// The section plan of the file (create | append | merge | unchanged | conflict); none when
    /// blocked.
    pub section: Option<FilePlan>,
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
    pub hosts: Vec<HostPlan>,
    pub not_supported: Vec<Unsupported>,
    pub plan_digest: String,
}

/// The block we last wrote into one host's file.
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
    /// create | append | merge
    pub action: String,
    pub before_sha256: Option<String>,
    pub after_sha256: String,
    /// The block this module owned before the apply; a revert restores it.
    pub baseline_before: Option<String>,
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

fn host_file(home: &Path, host: &str) -> PathBuf {
    let rel = HOSTS
        .iter()
        .find(|(h, _, _)| *h == host)
        .map(|(_, f, _)| *f)
        .unwrap_or_else(|| unreachable!("unknown routing host {host}"));
    home.join(rel)
}

fn state_root(home: &Path) -> PathBuf {
    home.join(crate::HOME_STATE_DIR).join(STATE_DIR)
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

/// The current bytes of the instruction file: `Ok(None)` when absent; a symlink or a folder in
/// its place is an error (the host is blocked, never followed).
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

fn host_folder_plain(home: &Path, host: &str) -> Result<()> {
    let file = host_file(home, host);
    plain_dir_or_absent(file.parent().expect("host folder"))
}

fn read_state(home: &Path, host: &str) -> Result<HostState> {
    let root = state_root(home);
    plain_dir_or_absent(&home.join(crate::HOME_STATE_DIR))?;
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

/// The section source, checked: UTF-8, one `{invoke}`, no marker tokens, ends with a newline.
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

/// The full managed block for one host, markers included.
pub fn render(source: &str, host: &str) -> Result<String> {
    let invoke = HOSTS
        .iter()
        .find(|(h, _, _)| *h == host)
        .map(|(_, _, i)| *i)
        .ok_or_else(|| format!("routing host not supported: {host:?}"))?;
    let (begin, end) = markers(SECTION, "html");
    Ok(format!(
        "{begin}\n{}{end}\n",
        source.replacen("{invoke}", invoke, 1)
    ))
}

fn host_plan(home: &Path, host: &str, rendered: &str) -> HostPlan {
    let file = host_file(home, host);
    let mut plan = HostPlan {
        host: host.into(),
        file: file.to_string_lossy().into_owned(),
        preflight: "ok".into(),
        blocked_reason: None,
        section: None,
    };
    let current = host_folder_plain(home, host)
        .and_then(|()| current_file(&file))
        .and_then(|cur| read_state(home, host).map(|s| (cur, s)));
    match current {
        Ok((cur, state)) => match section_plan(
            &plan.file,
            false,
            SECTION,
            "html",
            cur.as_deref(),
            state.block_sha256,
            rendered,
        ) {
            Ok(section) => plan.section = Some(section),
            Err(e) => {
                plan.preflight = "blocked".into();
                plan.blocked_reason = Some(e);
            }
        },
        Err(e) => {
            plan.preflight = "blocked".into();
            plan.blocked_reason = Some(e);
        }
    }
    plan
}

/// Read-only preview; writes nothing.
pub fn plan(home: &Path, package: &Path) -> Result<Plan> {
    let home = fs::canonicalize(home).map_err(|_| "home directory unavailable")?;
    let src = source(package)?;
    let mut hosts = Vec::new();
    for (host, _, _) in HOSTS {
        hosts.push(host_plan(&home, host, &render(&src, host)?));
    }
    let mut plan = Plan {
        schema_version: SCHEMA_VERSION,
        home: home.to_str().ok_or("non UTF-8 home path")?.to_string(),
        package_digest: crate::package_digest(package)?,
        source_sha256: sha(src.as_bytes()),
        hosts,
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
    Ok(plan)
}

fn validate_permission(permission: &str) -> Result<()> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err("permission reference required".into());
    }
    Ok(())
}

fn new_receipt_dir(home: &Path) -> Result<(String, PathBuf)> {
    plain_dir_or_absent(&home.join(crate::HOME_STATE_DIR))?;
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

/// Applies a previewed plan. `plan_digest` must equal a fresh plan's digest; only `create`,
/// `append` and `merge` are written, blocked hosts and conflicts are skipped.
pub fn apply(
    home: &Path,
    package: &Path,
    plan_digest: &str,
    permission: &str,
) -> Result<serde_json::Value> {
    validate_permission(permission)?;
    let fresh = plan(home, package)?;
    if fresh.plan_digest != plan_digest {
        return Err("stale or altered plan; re-preview required".into());
    }
    let home = PathBuf::from(&fresh.home);
    let src = source(package)?;
    let writes: Vec<(&HostPlan, &FilePlan)> = fresh
        .hosts
        .iter()
        .filter(|h| h.preflight == "ok")
        .filter_map(|h| h.section.as_ref().map(|s| (h, s)))
        .filter(|(_, s)| ["create", "append", "merge"].contains(&s.action.as_str()))
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
    let (begin, end) = markers(SECTION, "html");
    let mut written = Vec::new();
    for (host, section) in writes {
        let target = host_file(&home, &host.host);
        let folder = target.parent().expect("host folder");
        // Checked before creating (create_dir_all would follow a linked parent) and after.
        host_folder_plain(&home, &host.host)?;
        fs::create_dir_all(folder).map_err(|_| format!("{} unavailable", folder.display()))?;
        host_folder_plain(&home, &host.host)?;
        let before = current_file(&target)?;
        if before.as_deref().map(sha) != section.expected_sha256 {
            return Err(format!(
                "{} changed since the plan; re-preview required",
                target.display()
            ));
        }
        let rendered = render(&src, &host.host)?;
        let out = splice(before.as_deref(), &begin, &end, &rendered, "")?;
        assert_single_block(&out, &begin, &end, &rendered)?;
        if let Some(b) = &before {
            publish_exclusive(&receipt_dir.join(format!("{}.before", host.host)), b)?;
        }
        match &before {
            Some(b) => publish_replace(&target, Some(b), &out)?,
            None => publish_exclusive(&target, &out)?,
        }
        receipt.entries.push(ReceiptEntry {
            host: host.host.clone(),
            action: section.action.clone(),
            before_sha256: before.as_deref().map(sha),
            after_sha256: sha(&out),
            baseline_before: section.baseline_sha256.clone(),
        });
        write_json(&receipt_path, &receipt)?;
        write_json(
            &state_root(&home).join(format!("{}.json", host.host)),
            &HostState {
                schema_version: SCHEMA_VERSION,
                block_sha256: Some(sha(rendered.as_bytes())),
            },
        )?;
        written.push(host.host.clone());
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
/// removed, a changed one gets its before bytes back. A file changed since is skipped.
pub fn revert(home: &Path, receipt_id: &str, permission: &str) -> Result<serde_json::Value> {
    validate_permission(permission)?;
    if !receipt_id_ok(receipt_id) {
        return Err("unknown routing receipt id".into());
    }
    let home = fs::canonicalize(home).map_err(|_| "home directory unavailable")?;
    plain_dir_or_absent(&home.join(crate::HOME_STATE_DIR))?;
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
    let mut reverted = Vec::new();
    let mut skipped = Vec::new();
    for e in receipt.entries.iter().rev() {
        if !HOSTS.iter().any(|(h, _, _)| *h == e.host) {
            return Err("routing receipt names an unknown host".into());
        }
        if let Err(reason) = host_folder_plain(&home, &e.host) {
            skipped.push(serde_json::json!({"host": e.host, "reason": format!("host_folder_not_plain: {reason}")}));
            continue;
        }
        let target = host_file(&home, &e.host);
        let current = match current_file(&target) {
            Ok(c) => c,
            Err(reason) => {
                skipped.push(serde_json::json!({"host": e.host, "reason": format!("file_not_plain: {reason}")}));
                continue;
            }
        };
        if current.as_deref().map(sha) != Some(e.after_sha256.clone()) {
            skipped.push(serde_json::json!({"host": e.host, "reason": "user_changed_since_apply"}));
            continue;
        }
        let current = current.expect("checked above");
        let mut state = read_state(&home, &e.host)?;
        match &e.before_sha256 {
            None => {
                fs::remove_file(&target)
                    .map_err(|_| format!("{} not removed", target.display()))?;
                state.block_sha256 = None;
            }
            Some(before_sha) => {
                let before = read_opt(&dir.join(format!("{}.before", e.host)))?
                    .ok_or("before bytes missing from the receipt")?;
                if &sha(&before) != before_sha {
                    return Err("before bytes do not match the receipt".into());
                }
                publish_replace(&target, Some(&current), &before)?;
                state.block_sha256 = e.baseline_before.clone();
            }
        }
        write_json(&state_root(&home).join(format!("{}.json", e.host)), &state)?;
        reverted.push(e.host.clone());
    }
    receipt.status = "reverted".into();
    write_json(&path, &receipt)?;
    Ok(
        serde_json::json!({"state": "reverted", "receipt": receipt_id, "reverted": reverted, "skipped": skipped}),
    )
}
