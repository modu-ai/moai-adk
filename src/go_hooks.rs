//! The Go-generation project hooks and their replacement by the MoAI hook set (Issue #2877,
//! parent #2871 step I5).
//!
//! A project the Go generation set up carries hook entries in `.claude/settings.json` whose
//! handler runs a script under `.claude/hooks/moai/` (`bash -c '…' <script>`, or the script as
//! the command itself); the scripts forward to the Go `moai hook <event>`. Once `moai` on PATH
//! is this generation, that call is swallowed with exit 0, so every Go guard would switch off
//! with no replacement — the entries have to be replaced, not left.
//!
//! This module only reads: it tells Go-generation handlers from every other handler, splits a
//! `hooks` object into what goes and what stays, and lists the scripts. [`preview`] puts that
//! together for one project: the Go entries that would go, the other entries that stay, the
//! MoAI entries that would be added and the scripts that would move to an archive folder.
//! It writes nothing.

use crate::Result;
use serde_json::{json, Map, Value};
use std::fs;
use std::path::Path;

/// The folder of the Go-generation hook scripts, relative to the project.
pub const GO_HOOK_DIR: &str = ".claude/hooks/moai";

const SETTINGS: &str = ".claude/settings.json";

/// What may stand before [`GO_HOOK_DIR`] in a Go-generation handler: nothing (the project is
/// the hook's working folder) or the project folder variable, as the Go generation wrote it.
const PROJECT_PREFIXES: &[&str] = &[
    "${CLAUDE_PROJECT_DIR}/",
    "$CLAUDE_PROJECT_DIR/",
    "\"$CLAUDE_PROJECT_DIR\"/",
    "\"${CLAUDE_PROJECT_DIR}\"/",
    "",
];

/// The script name when one whole token is this project's Go hook script: an optional pair of
/// surrounding quotes, one of [`PROJECT_PREFIXES`], [`GO_HOOK_DIR`]`/` and a plain file name
/// (letters, digits, `-`, `_`, `.`; never `.` or `..`) up to the token's end.
fn script_token(token: &str) -> Option<String> {
    let t = ['"', '\'']
        .iter()
        .find_map(|q| {
            token
                .strip_prefix(*q)
                .and_then(|r| r.strip_suffix(*q))
                .filter(|inner| !inner.contains(*q))
        })
        .unwrap_or(token);
    let marker = format!("{GO_HOOK_DIR}/");
    PROJECT_PREFIXES.iter().find_map(|p| {
        let name = t.strip_prefix(p)?.strip_prefix(&marker)?;
        let plain = !name.is_empty()
            && name != "."
            && name != ".."
            && name
                .chars()
                .all(|c| c.is_ascii_alphanumeric() || matches!(c, '-' | '_' | '.'));
        plain.then(|| name.to_string())
    })
}

/// Whether a command runs a script given as its argument (`bash`, `sh`, by name or path).
fn is_shell(token: &str) -> bool {
    matches!(
        token,
        "bash" | "sh" | "/bin/bash" | "/bin/sh" | "/usr/bin/bash"
    )
}

/// The Go hook script a handler runs, when it runs one (#2884 audit B1): the command itself is
/// the script, or a shell (`bash`/`sh`) runs it — as the command's next word, or as one whole
/// `args` value. A path anywhere else, under another folder or with another prefix (another
/// project, HOME, `~/`), a name like `..`, or a command that only reads the folder is not a Go
/// hook: `None`.
pub fn go_script(handler: &Value) -> Option<String> {
    let command = handler["command"].as_str()?;
    let mut words = command.split_whitespace();
    let first = words.next()?;
    if let Some(name) = script_token(first) {
        return Some(name);
    }
    if !is_shell(first) {
        return None;
    }
    // `bash <script>` in the command, or `bash` with the script among its arguments.
    let rest: Vec<&str> = words.filter(|w| !w.starts_with('-')).collect();
    if let [only] = rest.as_slice() {
        if let Some(name) = script_token(only) {
            return Some(name);
        }
    }
    handler["args"]
        .as_array()?
        .iter()
        .filter_map(Value::as_str)
        .find_map(script_token)
}

/// One Go-generation handler found in a `hooks` object.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct GoEntry {
    pub event: String,
    pub matcher: Option<String>,
    pub script: String,
}

/// `hooks` split in two: the Go-generation handlers, and the object with them removed (a
/// matcher group left without handlers and an event left without groups are dropped; every
/// other key is kept as it is).
pub fn split(hooks: &Map<String, Value>) -> (Vec<GoEntry>, Map<String, Value>) {
    let mut go = Vec::new();
    let mut kept = Map::new();
    for (event, groups) in hooks {
        let Some(groups) = groups.as_array() else {
            kept.insert(event.clone(), groups.clone());
            continue;
        };
        let mut kept_groups = Vec::new();
        for group in groups {
            let Some(handlers) = group["hooks"].as_array() else {
                kept_groups.push(group.clone());
                continue;
            };
            let mut kept_handlers = Vec::new();
            for h in handlers {
                match go_script(h) {
                    Some(script) => go.push(GoEntry {
                        event: event.clone(),
                        matcher: group["matcher"].as_str().map(str::to_string),
                        script,
                    }),
                    None => kept_handlers.push(h.clone()),
                }
            }
            if kept_handlers.is_empty() {
                continue;
            }
            let mut g = group.clone();
            g["hooks"] = Value::Array(kept_handlers);
            kept_groups.push(g);
        }
        if !kept_groups.is_empty() {
            kept.insert(event.clone(), Value::Array(kept_groups));
        }
    }
    (go, kept)
}

/// The files under the project's [`GO_HOOK_DIR`], sorted; empty when the folder is absent.
/// A link anywhere on the way, or a folder inside it, is refused (nothing is followed).
pub fn scripts(project: &Path) -> Result<Vec<String>> {
    let mut dir = project.to_path_buf();
    for part in GO_HOOK_DIR.split('/') {
        dir.push(part);
        match fs::symlink_metadata(&dir) {
            Err(_) => return Ok(Vec::new()),
            Ok(m) if m.file_type().is_symlink() => {
                return Err(format!("{} is a link; nothing is followed", dir.display()))
            }
            Ok(m) if !m.is_dir() => return Err(format!("{} is not a folder", dir.display())),
            Ok(_) => {}
        }
    }
    let mut out = Vec::new();
    for e in fs::read_dir(&dir)
        .map_err(|e| format!("{} unreadable: {e}", dir.display()))?
        .flatten()
    {
        let name = e.file_name().to_string_lossy().into_owned();
        let kind = fs::symlink_metadata(e.path()).map_err(|e| e.to_string())?;
        if !kind.is_file() {
            return Err(format!("{GO_HOOK_DIR}/{name} is not a regular file"));
        }
        out.push(name);
    }
    out.sort();
    Ok(out)
}

/// The project's `.claude/settings.json` as an object; `None` when the file is absent. A link,
/// a file over the profile size limit or text that is not a JSON object is refused.
fn settings(project: &Path) -> Result<Option<Map<String, Value>>> {
    let claude = project.join(".claude");
    if fs::symlink_metadata(&claude).is_ok_and(|m| m.file_type().is_symlink()) {
        return Err(".claude is a link; nothing is followed".into());
    }
    let path = project.join(SETTINGS);
    match fs::symlink_metadata(&path) {
        Err(_) => return Ok(None),
        Ok(m) if m.file_type().is_symlink() => {
            return Err(format!("{SETTINGS} is a link; nothing is followed"))
        }
        Ok(_) => {}
    }
    let bytes = crate::profile::read(&path)?.ok_or(format!("{SETTINGS} vanished"))?;
    match serde_json::from_slice::<Value>(&bytes) {
        Ok(Value::Object(map)) => Ok(Some(map)),
        _ => Err(format!("{SETTINGS} is not a JSON object")),
    }
}

/// What replacing the project's Go-generation hooks would do (#2877). Reads only.
///
/// `go_entries` would go, `kept_entries` (every handler that is not Go-generation) stay as
/// they are, `added` is the MoAI hook set for Claude Code ([`crate::hooks::hooks_object`]),
/// and `scripts` would move to an archive folder, not be deleted. `hooks_after` is the `hooks`
/// value the replacement would write. `action` is `replace` when there is a Go entry or a Go
/// script, `none` otherwise.
pub fn preview(project: &Path) -> Result<Value> {
    let settings = settings(project)?;
    let hooks = settings
        .as_ref()
        .and_then(|s| s.get("hooks"))
        .and_then(Value::as_object)
        .cloned()
        .unwrap_or_default();
    let (go, kept) = split(&hooks);
    let scripts = scripts(project)?;
    let added = crate::hooks::hooks_object("claude")?;
    let mut after = kept.clone();
    for (event, groups) in &added {
        let list = after
            .entry(event.clone())
            .or_insert_with(|| Value::Array(Vec::new()));
        if let (Some(list), Some(groups)) = (list.as_array_mut(), groups.as_array()) {
            for g in groups {
                if !list.contains(g) {
                    list.push(g.clone());
                }
            }
        }
    }
    let referenced: Vec<&str> = go.iter().map(|g| g.script.as_str()).collect();
    let unreferenced: Vec<&String> = scripts
        .iter()
        .filter(|s| !referenced.contains(&s.as_str()))
        .collect();
    let missing: Vec<&str> = {
        let mut m: Vec<&str> = referenced
            .iter()
            .copied()
            .filter(|r| !scripts.iter().any(|s| s == r))
            .collect();
        m.sort();
        m.dedup();
        m
    };
    let action = if go.is_empty() && scripts.is_empty() {
        "none"
    } else {
        "replace"
    };
    Ok(json!({
        "settings": SETTINGS,
        "settings_present": settings.is_some(),
        "action": action,
        "go_entries": go.iter().map(|g| json!({
            "event": g.event, "matcher": g.matcher, "script": g.script,
        })).collect::<Vec<_>>(),
        "kept_entries": kept,
        "added": added,
        "hooks_after": after,
        "scripts": scripts,
        "scripts_unreferenced": unreferenced,
        "scripts_missing": missing,
        "script_dir": GO_HOOK_DIR,
        "writes": false,
    }))
}

// ---- replacement: plan, apply, revert (#2877, after #2868) ----
//
// The apply moves the script folder into a receipt folder, takes the Go entries out of
// `.claude/settings.json` (the `hooks` key goes; every other key stays byte for byte), then
// installs the MoAI hook set through `hooks::plan` / `hooks::apply` — the writer and ownership
// receipt `hooks-apply` uses, so `moai update` sees the hooks as installed afterwards. A failure
// after the first write puts everything back. The receipt keeps the replaced bytes, and
// `go-hooks-revert` restores them while the files still hold what the apply wrote.

const RECEIPT_PREFIX: &str = "go-hooks-";
const CLAUDE_RECEIPT: &str = "manifests/claude-settings.json";

/// The replacement preview of `project` with what an apply needs: `action` is `replace`,
/// `none` (nothing of the Go generation) or `blocked`; `blocks` says why. Entries of other tools
/// in `hooks` block for now: the MoAI writer owns the whole `hooks` key, so keeping them needs a
/// decision first (fail-closed). `plan_digest` binds the settings bytes, the scripts and the
/// result.
pub fn plan(project: &Path, home: Option<&Path>) -> Result<Value> {
    let root = crate::hooks::project_root(project, home)?;
    let mut p = preview(&root)?;
    let mut blocks = Vec::new();
    if p["kept_entries"].as_object().is_some_and(|m| !m.is_empty()) {
        blocks.push("other_hooks_present");
    }
    if p["action"] == json!("replace") && !blocks.is_empty() {
        p["action"] = json!("blocked");
    }
    let settings = crate::claude::read_existing(&root.join(SETTINGS))?;
    let mut script_digests = Vec::new();
    for s in p["scripts"].as_array().into_iter().flatten() {
        let name = s.as_str().unwrap_or_default();
        let bytes = fs::read(root.join(GO_HOOK_DIR).join(name))
            .map_err(|_| format!("{GO_HOOK_DIR}/{name} unreadable"))?;
        script_digests.push(json!([name, crate::digest(&bytes)]));
    }
    let bound = json!({
        "settings_sha256": settings.as_deref().map(crate::digest),
        "scripts": script_digests,
        "go_entries": p["go_entries"],
        "hooks_after": p["hooks_after"],
        "action": p["action"],
    });
    p["blocks"] = json!(blocks);
    p["project"] = json!(root.display().to_string());
    p["plan_digest"] = json!(crate::digest(
        &serde_json::to_vec(&bound).map_err(|_| "encode failed")?
    ));
    p["next_actions"] = json!([
        "review the entries that go, stay and are added, then run `moai-cli go-hooks-apply --plan-digest <plan_digest> --permission-reference <ref> <project>`"
    ]);
    Ok(p)
}

/// The record of one replacement, under `<state>/receipts/go-hooks-<id>/`.
#[derive(serde::Serialize, serde::Deserialize)]
struct Receipt {
    schema_version: u32,
    id: String,
    plan_digest: String,
    permission_reference: String,
    status: String,
    /// sha256 of `.claude/settings.json` before and after (`None`: absent).
    settings_before: Option<String>,
    settings_after: Option<String>,
    /// sha256 of the Claude ownership receipt before and after.
    claude_receipt_before: Option<String>,
    claude_receipt_after: Option<String>,
    /// The script files moved to `archive/moai` in the receipt folder.
    scripts: Vec<String>,
}

fn receipts_dir(root: &Path) -> std::path::PathBuf {
    crate::state_root::path(root, "receipts")
}

fn new_receipt_dir(root: &Path) -> Result<(String, std::path::PathBuf)> {
    let receipts = receipts_dir(root);
    crate::claude::refuse_symlink_dir(&crate::state_root::path(root, ""))?;
    fs::create_dir_all(&receipts).map_err(|_| "receipt folder unavailable")?;
    crate::claude::refuse_symlink_dir(&receipts)?;
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
    Err("receipt id unavailable".into())
}

fn write_receipt(dir: &Path, r: &Receipt) -> Result<()> {
    let mut bytes = serde_json::to_vec_pretty(r).map_err(|_| "encode failed")?;
    bytes.push(b'\n');
    let path = dir.join("receipt.json");
    let before = crate::claude::read_existing(&path)?;
    crate::apply::publish_replace(&path, before.as_deref(), &bytes)
}

/// Puts a file back to `bytes` (or removes it when it did not exist), from `now`.
fn restore(path: &Path, now: Option<&[u8]>, bytes: Option<&[u8]>) -> Result<()> {
    match bytes {
        Some(b) => crate::apply::publish_replace(path, now, b),
        None if now.is_some() => {
            fs::remove_file(path).map_err(|_| format!("{} not removed", path.display()))
        }
        None => Ok(()),
    }
}

/// Replaces the Go-generation hooks of `project` after the user confirmed the plan whose digest
/// is `plan_digest` (see the module notes for the order of writes).
pub fn apply(
    project: &Path,
    home: Option<&Path>,
    plan_digest: &str,
    permission: &str,
) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err("a permission reference is required; nothing was written".into());
    }
    let p = plan(project, home)?;
    if p["plan_digest"].as_str() != Some(plan_digest) {
        return Err("stale or altered plan; re-preview required".into());
    }
    match p["action"].as_str() {
        Some("replace") => {}
        Some("none") => return Ok(json!({"state": "unchanged", "writes": []})),
        _ => return Err(format!("the plan is blocked: {}", p["blocks"])),
    }
    let root = crate::hooks::project_root(project, home)?;
    let settings_path = root.join(SETTINGS);
    let receipt_path = crate::state_root::path(&root, CLAUDE_RECEIPT);
    let settings_before = crate::claude::read_existing(&settings_path)?;
    let claude_before = crate::claude::read_existing(&receipt_path)?;
    let scripts: Vec<String> = p["scripts"]
        .as_array()
        .into_iter()
        .flatten()
        .filter_map(|s| s.as_str().map(str::to_string))
        .collect();

    let (id, dir) = new_receipt_dir(&root)?;
    let keep = dir.join("before");
    fs::create_dir_all(&keep).map_err(|_| "receipt folder unavailable")?;
    if let Some(b) = &settings_before {
        fs::write(keep.join("settings.json"), b).map_err(|_| "receipt write failed")?;
    }
    if let Some(b) = &claude_before {
        fs::write(keep.join("claude-settings.json"), b).map_err(|_| "receipt write failed")?;
    }
    let mut receipt = Receipt {
        schema_version: 1,
        id: id.clone(),
        plan_digest: plan_digest.to_string(),
        permission_reference: permission.to_string(),
        status: "pending".into(),
        settings_before: settings_before.as_deref().map(crate::digest),
        settings_after: None,
        claude_receipt_before: claude_before.as_deref().map(crate::digest),
        claude_receipt_after: None,
        scripts: scripts.clone(),
    };
    write_receipt(&dir, &receipt)?;

    // 1. The scripts move into the receipt folder (nothing is deleted).
    let script_dir = root.join(GO_HOOK_DIR);
    let archive = dir.join("archive");
    let moved = if scripts.is_empty() {
        false
    } else {
        fs::create_dir_all(&archive).map_err(|_| "receipt folder unavailable")?;
        fs::rename(&script_dir, archive.join("moai"))
            .map_err(|e| format!("{GO_HOOK_DIR} not moved: {e}"))?;
        true
    };
    let undo_move = || {
        if moved {
            let _ = fs::rename(archive.join("moai"), &script_dir);
        }
    };

    // 2. The Go `hooks` key goes; every other key stays.
    let without = match &settings_before {
        None => None,
        Some(b) => {
            let Ok(Value::Object(mut map)) = serde_json::from_slice::<Value>(b) else {
                undo_move();
                return Err(format!("{SETTINGS} is not a JSON object"));
            };
            map.remove("hooks");
            let mut out = serde_json::to_vec_pretty(&map).map_err(|_| "encode failed")?;
            out.push(b'\n');
            Some(out)
        }
    };
    if let Some(w) = &without {
        if let Err(e) = crate::apply::publish_replace(&settings_path, settings_before.as_deref(), w)
        {
            undo_move();
            return Err(e);
        }
    }

    // 3. The MoAI hook set through the hooks-apply writer.
    let installed = crate::hooks::plan(&root, "claude", home).and_then(|hp| {
        if hp["action"] == json!("conflict") {
            return Err(format!("hooks-plan conflict: {}", hp["conflicts"]));
        }
        let digest = hp["plan_digest"].as_str().unwrap_or_default().to_string();
        crate::hooks::apply(&root, "claude", home, &digest, permission)
    });
    if let Err(e) = installed {
        let now = crate::claude::read_existing(&settings_path)?;
        restore(&settings_path, now.as_deref(), settings_before.as_deref())?;
        let now = crate::claude::read_existing(&receipt_path)?;
        restore(&receipt_path, now.as_deref(), claude_before.as_deref())?;
        undo_move();
        receipt.status = "rolled_back".into();
        write_receipt(&dir, &receipt)?;
        return Err(format!(
            "the MoAI hooks were not installed ({e}); everything was put back"
        ));
    }

    receipt.settings_after = crate::claude::read_existing(&settings_path)?
        .as_deref()
        .map(crate::digest);
    receipt.claude_receipt_after = crate::claude::read_existing(&receipt_path)?
        .as_deref()
        .map(crate::digest);
    receipt.status = "applied".into();
    write_receipt(&dir, &receipt)?;
    Ok(json!({
        "state": "applied",
        "project": root.display().to_string(),
        "removed_entries": p["go_entries"].as_array().map_or(0, Vec::len),
        "scripts_archived": scripts.len(),
        "archive": crate::state_root::rel(&root, &format!("receipts/{id}/archive/moai")),
        "receipt": id,
        "host_loaded": "not_run",
        "next_actions": [format!("to undo: moai-cli go-hooks-revert --receipt {id} --permission-reference <ref> <project>")],
    }))
}

/// Undoes one replacement while `.claude/settings.json` and the Claude ownership receipt still
/// hold the bytes the apply wrote and the script folder is free: the settings file and the
/// receipt get their before bytes back and the scripts move back.
pub fn revert(project: &Path, receipt_id: &str, permission: &str) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err("a permission reference is required; nothing was written".into());
    }
    if !(receipt_id.starts_with(RECEIPT_PREFIX)
        && receipt_id.len() <= 64
        && receipt_id
            .bytes()
            .all(|b| b.is_ascii_alphanumeric() || b == b'-'))
    {
        return Err("unknown receipt id".into());
    }
    let root = fs::canonicalize(project).map_err(|_| "project directory unavailable")?;
    let dir = receipts_dir(&root).join(receipt_id);
    for d in [
        crate::state_root::path(&root, ""),
        receipts_dir(&root),
        dir.clone(),
    ] {
        crate::claude::refuse_symlink_dir(&d)?;
    }
    let bytes =
        crate::claude::read_existing(&dir.join("receipt.json"))?.ok_or("receipt not found")?;
    let mut r: Receipt = serde_json::from_slice(&bytes).map_err(|_| "receipt is unreadable")?;
    if r.status != "applied" {
        return Err(format!("the receipt is {}; nothing to revert", r.status));
    }
    let settings_path = root.join(SETTINGS);
    let receipt_path = crate::state_root::path(&root, CLAUDE_RECEIPT);
    let settings_now = crate::claude::read_existing(&settings_path)?;
    let claude_now = crate::claude::read_existing(&receipt_path)?;
    if settings_now.as_deref().map(crate::digest) != r.settings_after
        || claude_now.as_deref().map(crate::digest) != r.claude_receipt_after
    {
        return Err(format!(
            "{SETTINGS} or its ownership receipt changed since the apply; nothing was written"
        ));
    }
    let script_dir = root.join(GO_HOOK_DIR);
    let archived = dir.join("archive").join("moai");
    if !r.scripts.is_empty() && fs::symlink_metadata(&script_dir).is_ok() {
        return Err(format!("{GO_HOOK_DIR} exists again; nothing was written"));
    }
    // Read and check every before copy first, so the revert is whole or not at all.
    let load = |name: &str, want: &Option<String>| -> Result<Option<Vec<u8>>> {
        match want {
            None => Ok(None),
            Some(sha) => {
                let b = fs::read(dir.join("before").join(name))
                    .map_err(|_| "before bytes missing from the receipt")?;
                if &crate::digest(&b) != sha {
                    return Err("before bytes do not match the receipt".into());
                }
                Ok(Some(b))
            }
        }
    };
    let settings_before = load("settings.json", &r.settings_before)?;
    let claude_before = load("claude-settings.json", &r.claude_receipt_before)?;
    restore(
        &settings_path,
        settings_now.as_deref(),
        settings_before.as_deref(),
    )?;
    restore(
        &receipt_path,
        claude_now.as_deref(),
        claude_before.as_deref(),
    )?;
    if !r.scripts.is_empty() {
        if let Some(parent) = script_dir.parent() {
            fs::create_dir_all(parent).map_err(|_| ".claude/hooks unavailable")?;
        }
        fs::rename(&archived, &script_dir)
            .map_err(|e| format!("{GO_HOOK_DIR} not moved back: {e}"))?;
    }
    r.status = "reverted".into();
    write_receipt(&dir, &r)?;
    Ok(json!({"state": "reverted", "receipt": receipt_id, "scripts_restored": r.scripts.len()}))
}

#[cfg(test)]
mod tests {
    use super::*;

    /// The shape the Go generation writes (settings.json of this repository, 2026-11-14).
    fn go_handler(script: &str) -> Value {
        json!({
            "type": "command",
            "command": "bash",
            "args": ["-c", "[ -f \"$0\" ] && exec bash \"$0\"; echo missing >> hook-missing.log; exit 0",
                     format!("${{CLAUDE_PROJECT_DIR}}/.claude/hooks/moai/{script}")],
            "timeout": 10
        })
    }

    #[test]
    fn a_go_handler_is_told_by_its_script_path() {
        assert_eq!(
            go_script(&go_handler("handle-pre-tool.sh")).as_deref(),
            Some("handle-pre-tool.sh")
        );
        let as_command = json!({"type": "command",
            "command": "\"$CLAUDE_PROJECT_DIR\"/.claude/hooks/moai/handle-stop.sh"});
        assert_eq!(go_script(&as_command).as_deref(), Some("handle-stop.sh"));
        // Every prefix the Go generation wrote, as a command and as `bash <script>`.
        for (command, name) in [
            ("${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/a.sh", "a.sh"),
            ("$CLAUDE_PROJECT_DIR/.claude/hooks/moai/b.sh", "b.sh"),
            ("\"${CLAUDE_PROJECT_DIR}\"/.claude/hooks/moai/c.sh", "c.sh"),
            (".claude/hooks/moai/d.sh", "d.sh"),
            (
                "bash \"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/e.sh\"",
                "e.sh",
            ),
            ("sh -e .claude/hooks/moai/f.sh", "f.sh"),
        ] {
            let h = json!({"type": "command", "command": command});
            assert_eq!(go_script(&h).as_deref(), Some(name), "{command}");
        }
        for other in [
            json!({"type": "command", "command": "moai-cli hook-receipt claude"}),
            json!({"type": "command", "command": "orca hook"}),
            json!({"type": "command", "command": "bash", "args": [".claude/hooks/mine/x.sh"]}),
            json!({"type": "command", "command": "echo .claude/hooks/moai/"}),
            // #2884 audit B1: another project, HOME, `..`, a reader, a glued prefix.
            json!({"type": "command", "command": "bash /opt/other-project/.claude/hooks/moai/x.sh"}),
            json!({"type": "command", "command": "bash ~/.claude/hooks/moai/x.sh"}),
            json!({"type": "command", "command": "bash .claude/hooks/moai/../../evil.sh"}),
            json!({"type": "command", "command": "bash .claude/hooks/moai/.."}),
            json!({"type": "command", "command": "cat .claude/hooks/moai/README >> log"}),
            json!({"type": "command", "command": "my.claude/hooks/moai/y.sh"}),
            json!({"type": "command", "command": "bash", "args": ["-c", "bash .claude/hooks/moai/x.sh"]}),
            json!({"type": "command", "command": "python3", "args": [".claude/hooks/moai/x.sh"]}),
            json!({"type": "command", "command": "bash .claude/hooks/moai/x.sh other.sh"}),
        ] {
            assert_eq!(go_script(&other), None, "{other}");
        }
    }

    #[test]
    fn split_takes_only_go_handlers_and_drops_what_is_left_empty() {
        let hooks = json!({
            "PreToolUse": [
                {"matcher": "Write|Edit|Bash", "hooks": [go_handler("handle-pre-tool.sh"),
                    {"type": "command", "command": "orca guard"}]},
                {"matcher": "Agent", "hooks": [go_handler("handle-pre-tool.sh")]}
            ],
            "Stop": [{"hooks": [go_handler("handle-stop.sh")]}],
            "Notification": [{"hooks": [{"type": "command", "command": "say done"}]}]
        });
        let (go, kept) = split(hooks.as_object().unwrap());
        assert_eq!(go.len(), 3);
        assert_eq!(go[0].matcher.as_deref(), Some("Write|Edit|Bash"));
        assert_eq!(go[2].event, "Stop");
        assert_eq!(
            Value::Object(kept),
            json!({
                "PreToolUse": [{"matcher": "Write|Edit|Bash",
                                "hooks": [{"type": "command", "command": "orca guard"}]}],
                "Notification": [{"hooks": [{"type": "command", "command": "say done"}]}]
            })
        );
    }
}
