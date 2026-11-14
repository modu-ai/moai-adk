//! The MoAI-Easy output style in each host's user layer (#2875, #2871 I3).
//!
//! The style text is one package file, `references/home/output-style.md`, rendered per host:
//! Claude Code reads it as the output style file `~/.claude/output-styles/MoAI-Easy.md`
//! (frontmatter added) and selects it by default through `outputStyle` in
//! `~/.claude/settings.json`; Codex reads it as a managed block with the id
//! `moai-output-style` in `~/.codex/AGENTS.md`. Hosts without a user-level file this module can
//! write are reported in `not_supported`, as for the routing section.
//!
//! `plan` writes nothing. Every target is three-way: the style file and the block are ours only
//! while they hold the bytes we last wrote (`~/.moai-cli/home-style/<host>.json`); anything else
//! is a `conflict`, never overwritten. `outputStyle` is set only when the key is absent — another
//! value is the user's choice and a `conflict`. A hand-written `<!-- moai:output-style` block in
//! `AGENTS.md` is a `conflict` too. A symlinked folder or file blocks the host. `apply` re-plans,
//! refuses any other digest and keeps the before bytes in a receipt; `revert` restores a file
//! only while it still holds the bytes the apply wrote, and removes the `outputStyle` key only
//! while it still names this style.

use crate::agents::{assert_single_block, markers, section_plan, splice};
use crate::apply::{publish_exclusive, publish_replace, read_opt};
use serde::{Deserialize, Serialize};
use serde_json::{json, Map, Value};
use std::{
    fs,
    path::{Path, PathBuf},
};

type Result<T> = std::result::Result<T, String>;

const SCHEMA_VERSION: u32 = 1;
const STATE_DIR: &str = "home-style";
/// The style source inside the skill package.
pub const SOURCE: &str = "references/home/output-style.md";
/// The managed block id in Codex's `AGENTS.md`.
pub const SECTION: &str = "moai-output-style";
/// The style name Claude Code selects (`outputStyle` is case-sensitive).
pub const STYLE_NAME: &str = "MoAI-Easy";
/// Hosts written.
pub const HOSTS: [&str; 2] = ["claude", "codex"];
const CLAUDE_STYLE_FILE: &str = ".claude/output-styles/MoAI-Easy.md";
const CLAUDE_SETTINGS: &str = ".claude/settings.json";
const CODEX_FILE: &str = ".codex/AGENTS.md";
/// The start of the block people pasted by hand before this module existed.
const HAND_BLOCK: &str = "<!-- moai:output-style";
const SOURCE_LIMIT: usize = 64 * 1024;
const SETTINGS_LIMIT: usize = 256 * 1024;
const DESCRIPTION: &str = "Plain-words pair programming: four steps (Understand, Plan, Do, Check), evidence before done, banners and a progress board.";

/// One write target of one host.
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct TargetPlan {
    pub host: String,
    /// style_file | setting | block
    pub kind: String,
    pub path: String,
    /// create | merge | append | set | unchanged | conflict | skipped | blocked
    pub action: String,
    /// The sha256 of the file's bytes the plan was made from; none when absent.
    pub expected_sha256: Option<String>,
    /// What this module owned before (style file or block sha256).
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
    pub targets: Vec<TargetPlan>,
    pub not_supported: Vec<Unsupported>,
    pub plan_digest: String,
}

/// The style file or block we last wrote for one host.
#[derive(Clone, Debug, Default, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
struct HostState {
    schema_version: u32,
    owned_sha256: Option<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct ReceiptEntry {
    pub host: String,
    pub kind: String,
    pub action: String,
    pub before_sha256: Option<String>,
    pub after_sha256: String,
    /// What this module owned before the apply; a revert restores it.
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

fn rel_path(host: &str, kind: &str) -> &'static str {
    match (host, kind) {
        ("claude", "style_file") => CLAUDE_STYLE_FILE,
        ("claude", "setting") => CLAUDE_SETTINGS,
        ("codex", "block") => CODEX_FILE,
        _ => unreachable!("unknown style target {host}/{kind}"),
    }
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

/// Every folder from HOME (exclusive) down to the target's parent must be plain or absent.
fn folders_plain(home: &Path, rel: &str) -> Result<()> {
    let mut dir = home.to_path_buf();
    let parts: Vec<&str> = rel.split('/').collect();
    for part in &parts[..parts.len() - 1] {
        dir.push(part);
        plain_dir_or_absent(&dir)?;
    }
    Ok(())
}

/// The current bytes of a target file: `Ok(None)` when absent; a symlink or a folder in its
/// place is an error (the target is blocked, never followed).
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
            owned_sha256: None,
        }),
        Some(bytes) => {
            let state: HostState = serde_json::from_slice(&bytes)
                .map_err(|_| format!("style state for {host} is unreadable; inspect it first"))?;
            if state.schema_version != SCHEMA_VERSION {
                return Err(format!("style state for {host} has an unknown schema"));
            }
            Ok(state)
        }
    }
}

fn write_state(home: &Path, host: &str, owned: Option<String>) -> Result<()> {
    write_json(
        &state_root(home).join(format!("{host}.json")),
        &HostState {
            schema_version: SCHEMA_VERSION,
            owned_sha256: owned,
        },
    )
}

fn write_json<T: Serialize>(path: &Path, value: &T) -> Result<()> {
    let mut bytes = serde_json::to_vec_pretty(value).map_err(|_| "encode failed")?;
    bytes.push(b'\n');
    match read_opt(path)? {
        Some(b) => publish_replace(path, Some(&b), &bytes),
        None => publish_exclusive(path, &bytes),
    }
}

/// The style source, checked: UTF-8, at most 64 KiB, no marker tokens, no frontmatter of its
/// own, ends with a newline.
fn source(package: &Path) -> Result<String> {
    let path = package.join(SOURCE);
    let text = match fs::symlink_metadata(&path) {
        Ok(m) if m.is_file() => {
            fs::read_to_string(&path).map_err(|_| format!("{SOURCE} is not readable UTF-8"))?
        }
        _ => return Err(format!("{SOURCE} missing from the package")),
    };
    if text.len() > SOURCE_LIMIT
        || text.starts_with("---")
        || ["<!--", "-->", "moai:managed:"]
            .iter()
            .any(|t| text.contains(t))
        || !text.ends_with('\n')
    {
        return Err(format!(
            "{SOURCE} must be at most 64 KiB, start without frontmatter, hold no marker tokens and end with a newline"
        ));
    }
    Ok(text)
}

/// Claude Code's output style file: frontmatter, then the source.
pub fn render_style_file(source: &str) -> String {
    format!(
        "---\nname: {STYLE_NAME}\ndescription: {DESCRIPTION}\nkeep-coding-instructions: true\n---\n\n{source}"
    )
}

/// Codex's managed block, markers included.
pub fn render_block(source: &str) -> String {
    let (begin, end) = markers(SECTION, "html");
    format!("{begin}\n{source}{end}\n")
}

fn target(home: &Path, host: &str, kind: &str) -> TargetPlan {
    TargetPlan {
        host: host.into(),
        kind: kind.into(),
        path: home
            .join(rel_path(host, kind))
            .to_string_lossy()
            .into_owned(),
        action: String::new(),
        expected_sha256: None,
        baseline_sha256: None,
        detail: None,
    }
}

fn blocked(mut t: TargetPlan, reason: String) -> TargetPlan {
    t.action = "blocked".into();
    t.detail = Some(reason);
    t
}

fn style_file_plan(home: &Path, rendered: &str) -> TargetPlan {
    let mut t = target(home, "claude", "style_file");
    let read = folders_plain(home, CLAUDE_STYLE_FILE)
        .and_then(|()| current_file(&home.join(CLAUDE_STYLE_FILE)))
        .and_then(|cur| read_state(home, "claude").map(|s| (cur, s)));
    let (cur, state) = match read {
        Ok(v) => v,
        Err(e) => return blocked(t, e),
    };
    t.expected_sha256 = cur.as_deref().map(sha);
    t.baseline_sha256 = state.owned_sha256.clone();
    t.action = match &cur {
        None => "create",
        Some(b) if b.as_slice() == rendered.as_bytes() => "unchanged",
        Some(b) if state.owned_sha256.as_deref() == Some(sha(b).as_str()) => "merge",
        Some(_) => {
            t.detail = Some(
                "the style file is not the one MoAI last wrote (edited or someone else's); preserve and review".into(),
            );
            "conflict"
        }
    }
    .into();
    t
}

/// The settings object, or why it cannot be read as one.
fn settings_object(bytes: &[u8]) -> std::result::Result<Map<String, Value>, String> {
    if bytes.len() > SETTINGS_LIMIT {
        return Err("settings.json is larger than 256 KiB; not changed".into());
    }
    match serde_json::from_slice::<Value>(bytes) {
        Ok(Value::Object(map)) => Ok(map),
        _ => Err("settings.json is not a JSON object; not changed".into()),
    }
}

fn setting_plan(home: &Path, style: &TargetPlan) -> TargetPlan {
    let mut t = target(home, "claude", "setting");
    let cur = match folders_plain(home, CLAUDE_SETTINGS)
        .and_then(|()| current_file(&home.join(CLAUDE_SETTINGS)))
    {
        Ok(c) => c,
        Err(e) => return blocked(t, e),
    };
    t.expected_sha256 = cur.as_deref().map(sha);
    if !["create", "merge", "unchanged"].contains(&style.action.as_str()) {
        t.action = "skipped".into();
        t.detail = Some(format!(
            "the style file is {}; the default is not set to a file MoAI does not own",
            style.action
        ));
        return t;
    }
    let Some(bytes) = cur else {
        t.action = "set".into();
        return t;
    };
    t.action = match settings_object(&bytes) {
        Err(e) => {
            t.detail = Some(e);
            "conflict"
        }
        Ok(map) => match map.get("outputStyle") {
            None => "set",
            Some(Value::String(s)) if s == STYLE_NAME => "unchanged",
            Some(other) => {
                t.detail = Some(format!(
                    "outputStyle is {other}: the user's choice; not changed"
                ));
                "conflict"
            }
        },
    }
    .into();
    t
}

fn block_plan(home: &Path, rendered: &str) -> TargetPlan {
    let mut t = target(home, "codex", "block");
    let read = folders_plain(home, CODEX_FILE)
        .and_then(|()| current_file(&home.join(CODEX_FILE)))
        .and_then(|cur| read_state(home, "codex").map(|s| (cur, s)));
    let (cur, state) = match read {
        Ok(v) => v,
        Err(e) => return blocked(t, e),
    };
    t.expected_sha256 = cur.as_deref().map(sha);
    t.baseline_sha256 = state.owned_sha256.clone();
    if cur
        .as_deref()
        .is_some_and(|b| String::from_utf8_lossy(b).contains(HAND_BLOCK))
    {
        t.action = "conflict".into();
        t.detail = Some(format!(
            "a hand-written `{HAND_BLOCK}` block is in the file; remove it to let the managed block replace it"
        ));
        return t;
    }
    match section_plan(
        &t.path.clone(),
        false,
        SECTION,
        "html",
        cur.as_deref(),
        state.owned_sha256,
        rendered,
    ) {
        Ok(section) => {
            t.action = section.action;
            t.detail = section.detail;
            t
        }
        Err(e) => blocked(t, e),
    }
}

/// Read-only preview of every host; writes nothing.
pub fn plan(home: &Path, package: &Path) -> Result<Plan> {
    plan_for(home, package, &HOSTS)
}

/// Read-only preview of the named hosts only (`moai update` leaves out a host whose skill copy
/// is blocked); writes nothing. An unknown host is refused.
pub fn plan_for(home: &Path, package: &Path, only: &[&str]) -> Result<Plan> {
    if let Some(h) = only.iter().find(|h| !HOSTS.contains(h)) {
        return Err(format!("style host not supported: {h:?}"));
    }
    let home = fs::canonicalize(home).map_err(|_| "home directory unavailable")?;
    // No host to write: the source is not needed (a package from before #2875 has none).
    let src = if only.is_empty() {
        String::new()
    } else {
        source(package)?
    };
    let mut targets = Vec::new();
    if only.contains(&"claude") {
        let style = style_file_plan(&home, &render_style_file(&src));
        let setting = setting_plan(&home, &style);
        targets.push(style);
        targets.push(setting);
    }
    if only.contains(&"codex") {
        targets.push(block_plan(&home, &render_block(&src)));
    }
    let mut plan = Plan {
        schema_version: SCHEMA_VERSION,
        home: home.to_str().ok_or("non UTF-8 home path")?.to_string(),
        package_digest: crate::package_digest(package)?,
        source_sha256: sha(src.as_bytes()),
        targets,
        not_supported: crate::home_routing::NOT_SUPPORTED
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

/// True when applying `plan` would write at least one file.
pub fn writes_anything(plan: &Plan) -> bool {
    plan.targets.iter().any(|t| is_write(&t.action))
}

/// Whether two plans write the same targets (#2875): `moai update` re-plans the style after the
/// routing section already went into the same `~/.codex/AGENTS.md`, so the file's sha256 — and
/// a `create` that is now an `append` — differ from the confirmed preview; which targets are
/// written must not.
pub fn same_writes(a: &Plan, b: &Plan) -> bool {
    let writes = |p: &Plan| -> Vec<(String, String, bool)> {
        p.targets
            .iter()
            .map(|t| (t.host.clone(), t.kind.clone(), is_write(&t.action)))
            .collect()
    };
    a.source_sha256 == b.source_sha256 && writes(a) == writes(b)
}

fn is_write(action: &str) -> bool {
    ["create", "append", "merge", "set"].contains(&action)
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
    fs::create_dir_all(&receipts).map_err(|_| "style receipt folder unavailable")?;
    plain_dir_or_absent(&state_root(home))?;
    plain_dir_or_absent(&receipts)?;
    let nanos = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_nanos())
        .unwrap_or(0);
    for n in 0..16 {
        let id = format!("style-{nanos}-{n}");
        let dir = receipts.join(&id);
        if fs::create_dir(&dir).is_ok() {
            return Ok((id, dir));
        }
    }
    Err("style receipt id unavailable".into())
}

/// The bytes a write leaves in the target file.
fn new_bytes(t: &TargetPlan, before: Option<&[u8]>, src: &str) -> Result<Vec<u8>> {
    match t.kind.as_str() {
        "style_file" => Ok(render_style_file(src).into_bytes()),
        "setting" => {
            let mut map = match before {
                None => Map::new(),
                Some(b) => settings_object(b)?,
            };
            map.insert("outputStyle".into(), json!(STYLE_NAME));
            let mut out =
                serde_json::to_vec_pretty(&Value::Object(map)).map_err(|_| "encode failed")?;
            out.push(b'\n');
            Ok(out)
        }
        _ => {
            let (begin, end) = markers(SECTION, "html");
            let rendered = render_block(src);
            let out = splice(before, &begin, &end, &rendered, "")?;
            assert_single_block(&out, &begin, &end, &rendered)?;
            Ok(out)
        }
    }
}

/// Applies a previewed plan. `plan_digest` must equal a fresh plan's digest; only `create`,
/// `append`, `merge` and `set` are written, blocked targets and conflicts are skipped.
pub fn apply(home: &Path, package: &Path, plan_digest: &str, permission: &str) -> Result<Value> {
    apply_for(home, package, &HOSTS, plan_digest, permission)
}

/// [`apply`] for the named hosts only; `plan_digest` must equal a fresh [`plan_for`] of the
/// same hosts.
pub fn apply_for(
    home: &Path,
    package: &Path,
    only: &[&str],
    plan_digest: &str,
    permission: &str,
) -> Result<Value> {
    validate_permission(permission)?;
    let fresh = plan_for(home, package, only)?;
    if fresh.plan_digest != plan_digest {
        return Err("stale or altered plan; re-preview required".into());
    }
    let home = PathBuf::from(&fresh.home);
    let writes: Vec<&TargetPlan> = fresh
        .targets
        .iter()
        .filter(|t| is_write(&t.action))
        .collect();
    if writes.is_empty() {
        return Ok(json!({"state": "unchanged", "plan_digest": fresh.plan_digest, "written": []}));
    }
    let src = source(package)?;
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
    for t in writes {
        let rel = rel_path(&t.host, &t.kind);
        let path = home.join(rel);
        let folder = path.parent().expect("target folder");
        // Checked before creating (create_dir_all would follow a linked parent) and after.
        folders_plain(&home, rel)?;
        fs::create_dir_all(folder).map_err(|_| format!("{} unavailable", folder.display()))?;
        folders_plain(&home, rel)?;
        let before = current_file(&path)?;
        if before.as_deref().map(sha) != t.expected_sha256 {
            return Err(format!(
                "{} changed since the plan; re-preview required",
                path.display()
            ));
        }
        let out = new_bytes(t, before.as_deref(), &src)?;
        if let Some(b) = &before {
            publish_exclusive(
                &receipt_dir.join(format!("{}-{}.before", t.host, t.kind)),
                b,
            )?;
        }
        match &before {
            Some(b) => publish_replace(&path, Some(b), &out)?,
            None => publish_exclusive(&path, &out)?,
        }
        receipt.entries.push(ReceiptEntry {
            host: t.host.clone(),
            kind: t.kind.clone(),
            action: t.action.clone(),
            before_sha256: before.as_deref().map(sha),
            after_sha256: sha(&out),
            baseline_before: t.baseline_sha256.clone(),
        });
        write_json(&receipt_path, &receipt)?;
        match t.kind.as_str() {
            "style_file" => write_state(&home, &t.host, Some(sha(&out)))?,
            "block" => write_state(&home, &t.host, Some(sha(render_block(&src).as_bytes())))?,
            _ => {}
        }
        written.push(format!("{}:{}", t.host, t.kind));
    }
    receipt.status = "committed".into();
    write_json(&receipt_path, &receipt)?;
    Ok(json!({
        "state": "applied",
        "plan_digest": fresh.plan_digest,
        "receipt": id,
        "written": written,
    }))
}

fn receipt_id_ok(id: &str) -> bool {
    id.starts_with("style-")
        && id.len() <= 64
        && id.bytes().all(|b| b.is_ascii_alphanumeric() || b == b'-')
}

/// Removes the `outputStyle` key the apply set, while it still names this style; the other keys
/// stay as they are. A file the apply created and nobody changed since is removed.
fn revert_setting(path: &Path, e: &ReceiptEntry) -> std::result::Result<(), String> {
    let current = current_file(path)
        .map_err(|r| format!("file_not_plain: {r}"))?
        .ok_or("user_changed_since_apply")?;
    if e.before_sha256.is_none() && sha(&current) == e.after_sha256 {
        return fs::remove_file(path).map_err(|_| format!("{} not removed", path.display()));
    }
    let mut map = settings_object(&current).map_err(|_| "user_changed_since_apply")?;
    if map.get("outputStyle") != Some(&json!(STYLE_NAME)) {
        return Err("user_changed_since_apply".into());
    }
    map.shift_remove("outputStyle");
    let mut out = serde_json::to_vec_pretty(&Value::Object(map)).map_err(|_| "encode failed")?;
    out.push(b'\n');
    publish_replace(path, Some(&current), &out)
}

/// Undoes one apply for the targets that still hold what it wrote: a created file is removed, a
/// changed one gets its before bytes back, the `outputStyle` key is removed. A target changed
/// since is skipped.
pub fn revert(home: &Path, receipt_id: &str, permission: &str) -> Result<Value> {
    validate_permission(permission)?;
    if !receipt_id_ok(receipt_id) {
        return Err("unknown style receipt id".into());
    }
    let home = fs::canonicalize(home).map_err(|_| "home directory unavailable")?;
    plain_dir_or_absent(&crate::home_root::path(&home))?;
    plain_dir_or_absent(&state_root(&home))?;
    let dir = state_root(&home).join("receipts").join(receipt_id);
    plain_dir_or_absent(&dir)?;
    let path = dir.join("receipt.json");
    let bytes = read_opt(&path)?.ok_or("style receipt not found")?;
    let mut receipt: Receipt =
        serde_json::from_slice(&bytes).map_err(|_| "style receipt is unreadable")?;
    if receipt.status == "reverted" {
        return Err("style receipt already reverted".into());
    }
    let mut reverted = Vec::new();
    let mut skipped = Vec::new();
    for e in receipt.entries.iter().rev() {
        if !matches!(
            (e.host.as_str(), e.kind.as_str()),
            ("claude", "style_file") | ("claude", "setting") | ("codex", "block")
        ) {
            return Err("style receipt names an unknown target".into());
        }
        let rel = rel_path(&e.host, &e.kind);
        let name = format!("{}:{}", e.host, e.kind);
        if let Err(reason) = folders_plain(&home, rel) {
            skipped.push(json!({"target": name, "reason": format!("folder_not_plain: {reason}")}));
            continue;
        }
        let target = home.join(rel);
        if e.kind == "setting" {
            match revert_setting(&target, e) {
                Ok(()) => reverted.push(name),
                Err(reason) => skipped.push(json!({"target": name, "reason": reason})),
            }
            continue;
        }
        let current = match current_file(&target) {
            Ok(c) => c,
            Err(reason) => {
                skipped
                    .push(json!({"target": name, "reason": format!("file_not_plain: {reason}")}));
                continue;
            }
        };
        if current.as_deref().map(sha) != Some(e.after_sha256.clone()) {
            skipped.push(json!({"target": name, "reason": "user_changed_since_apply"}));
            continue;
        }
        let current = current.expect("checked above");
        match &e.before_sha256 {
            None => {
                fs::remove_file(&target)
                    .map_err(|_| format!("{} not removed", target.display()))?;
            }
            Some(before_sha) => {
                let before = read_opt(&dir.join(format!("{}-{}.before", e.host, e.kind)))?
                    .ok_or("before bytes missing from the receipt")?;
                if &sha(&before) != before_sha {
                    return Err("before bytes do not match the receipt".into());
                }
                publish_replace(&target, Some(&current), &before)?;
            }
        }
        read_state(&home, &e.host)?;
        write_state(&home, &e.host, e.baseline_before.clone())?;
        reverted.push(name);
    }
    receipt.status = "reverted".into();
    write_json(&path, &receipt)?;
    Ok(
        json!({"state": "reverted", "receipt": receipt_id, "reverted": reverted, "skipped": skipped}),
    )
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn the_style_file_names_the_style_claude_selects() {
        let out = render_style_file("body\n");
        assert!(out.starts_with("---\nname: MoAI-Easy\n"));
        assert!(out.contains("\nkeep-coding-instructions: true\n---\n\nbody\n"));
    }

    #[test]
    fn the_block_carries_the_managed_markers() {
        let out = render_block("body\n");
        assert_eq!(
            out,
            "<!-- moai:managed:begin id=moai-output-style -->\nbody\n<!-- moai:managed:end id=moai-output-style -->\n"
        );
    }

    #[test]
    fn a_settings_file_that_is_not_an_object_is_refused() {
        assert!(settings_object(b"[]").is_err());
        assert!(settings_object(b"{").is_err());
        assert!(settings_object(&vec![b' '; SETTINGS_LIMIT + 1]).is_err());
        assert!(settings_object(b"{\"a\":1}").is_ok());
    }
}
