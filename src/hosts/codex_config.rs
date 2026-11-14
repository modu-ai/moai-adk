//! Project `.codex/config.toml` plan and apply (Issue #2240, rc.2 I3, REQ-I3-01).
//!
//! Sources (read 2026-11-09):
//! - https://learn.chatgpt.com/docs/config-file/config-basic — user `~/.codex/config.toml`,
//!   project overrides in `.codex/config.toml`, "Codex loads project `.codex/` layers only when
//!   you trust the project", closest project file wins.
//! - https://learn.chatgpt.com/docs/config-file/config-reference — `tui.status_line`
//!   (`array<string> | null`, item identifiers not enumerated; #2928 S-5 checks them against the
//!   Codex 0.169.0 source list [`CODEX_STATUS_LINE_ITEMS`]), `skills.bundled.enabled`
//!   (`boolean`, default `true`, may be `false` in a trusted project), `project_doc_fallback_filenames`
//!   (`array<string>`), `project_doc_max_bytes` (`number`), `projects.<path>.trust_level`
//!   (stated only for the user config), and the keys Codex ignores in a project file.
//! - https://learn.chatgpt.com/docs/agent-configuration/agents-md — `project_doc_max_bytes`
//!   32 KiB by default.
//!
//! MoAI owns at most the four keys in [`MANAGED_CODEX_KEYS`] and decides per key with the
//! same baseline/current/desired three-way table as the Claude settings (`claude.rs`): a key
//! changes only when the user left it as MoAI owned it. Everything else in the file — other
//! keys, tables, comments, blank lines, order — is kept byte for byte (`toml_edit`). Trust is
//! never written: it lives in the user config, which is the home, and the home is blocked.
use serde_json::{json, Map, Value};
use std::{
    fs,
    path::{Path, PathBuf},
};
use toml_edit::{Array, DocumentMut, Item, Table, TableLike};

type Result<T> = std::result::Result<T, String>;

/// Dotted keys MoAI may own in a project `.codex/config.toml`.
pub const MANAGED_CODEX_KEYS: [&str; 4] = [
    "project_doc_fallback_filenames",
    "project_doc_max_bytes",
    "skills.bundled.enabled",
    "tui.status_line",
];

/// Top-level keys the config reference says Codex ignores in a project file (2026-11-09).
const PROJECT_IGNORED: [&str; 10] = [
    "openai_base_url",
    "chatgpt_base_url",
    "apps_mcp_product_sku",
    "model_provider",
    "model_providers",
    "notify",
    "profile",
    "profiles",
    "experimental_realtime_ws_base_url",
    "otel",
];

const CONFIG_FILE: &str = ".codex/config.toml";
/// Ownership receipt under the project state root, the same shape as the Claude settings one.
const CONFIG_RECEIPT: &str = "manifests/codex-config.json";
/// Size bound for the config file, its receipt and the rendered result.
pub const CONFIG_LIMIT: u64 = 262_144;
/// Upper bound MoAI accepts for `project_doc_max_bytes` (the reference states none).
pub const MAX_DOC_BYTES: i64 = 1024 * 1024;
const MAX_LIST: usize = 32;
const MAX_ITEM: usize = 128;

const TRUST_DECLARATION: &str = "Codex reads .codex/config.toml only for a trusted project; trust is projects.<path>.trust_level in the user config (~/.codex/config.toml), which this operation never writes";
const STATUS_LINE_UNVERIFIED: &str = "these tui.status_line items are not Codex 0.169.0 identifiers (openai/codex rust-v0.169.0 a41fe5b443652cbdbee57717fbb9fd78d8d8b32a, codex-rs/tui/src/bottom_pane/status_line_setup.rs:57-169); they are written as given, and Codex 0.169.0 ignores them with the warning \"Ignored invalid status line item\" (codex-rs/tui/src/chatwidget/status_surfaces.rs:121-138)";

/// Every `tui.status_line` identifier Codex 0.169.0 parses, aliases included (#2928 S-5).
/// The config reference lists none, so the source is the `StatusLineItem` enum of openai/codex
/// tag `rust-v0.169.0`, commit `a41fe5b443652cbdbee57717fbb9fd78d8d8b32a`,
/// `codex-rs/tui/src/bottom_pane/status_line_setup.rs:57-169`: each variant's kebab-case name
/// plus every `to_string`/`serialize` alias (read 2026-11-14). The picker's
/// `status-line-use-theme-colors` toggle is not an item and is not here.
pub const CODEX_STATUS_LINE_ITEMS: [&str; 41] = [
    "model",
    "model-name",
    "model-with-reasoning",
    "reasoning",
    "current-dir",
    "project-name",
    "project",
    "project-root",
    "hostname",
    "git-branch",
    "pull-request-number",
    "branch-changes",
    "run-state",
    "status",
    "permissions",
    "approval-mode",
    "approval",
    "context-remaining",
    "context-used",
    "context-usage",
    "five-hour-limit",
    "weekly-limit",
    "codex-version",
    "context-window-size",
    "used-tokens",
    "total-input-tokens",
    "total-output-tokens",
    "thread-credits",
    "estimated-thread-cost",
    "thread-id",
    "session-id",
    "fast-mode",
    "daybreak",
    "raw-output",
    "thread-name",
    "thread-title",
    "workspace-headline",
    "task-progress",
    "profile",
    "provider",
    "model-provider",
];

/// `"tui.status_line": "moai"` (#2928 S-5): the Claude Code status-line elements Codex 0.169.0
/// can show, in Claude Code order (line 1 model, effort, host version; line 2 repository, PR,
/// branch, changes; line 3 name tag, context, week). The mapping and what Codex cannot show:
/// `docs/codex-adapter-contract.md` "상태줄 (#2928 S-5)".
pub const MOAI_STATUS_LINE_PRESET: [&str; 10] = [
    "model",
    "reasoning",
    "codex-version",
    "project-name",
    "pull-request-number",
    "git-branch",
    "branch-changes",
    "thread-name",
    "context-remaining",
    "weekly-limit",
];

/// Replaces a preset name given as the `tui.status_line` value with its list. Returns the
/// request MoAI plans with and the preset used. Any other string is refused.
fn resolve_preset(desired: &Map<String, Value>) -> Result<(Map<String, Value>, Option<&str>)> {
    let mut out = desired.clone();
    let preset = match desired.get("tui.status_line") {
        Some(Value::String(name)) if name == "moai" => {
            out.insert("tui.status_line".into(), json!(MOAI_STATUS_LINE_PRESET));
            Some("moai")
        }
        Some(Value::String(_)) => {
            return Err("invalid codex config value for tui.status_line".into())
        }
        _ => None,
    };
    Ok((out, preset))
}

fn is_name(s: &str) -> bool {
    !s.is_empty() && s.len() <= MAX_ITEM && !s.chars().any(char::is_control) && s.trim() == s
}

/// Checks one desired key and value. Trust, ignored and unknown keys are refused.
fn validate(key: &str, value: &Value) -> Result<()> {
    let top = key.split('.').next().unwrap_or(key);
    if top == "projects" {
        return Err("projects.<path>.trust_level belongs to the user config; refused".into());
    }
    if PROJECT_IGNORED.contains(&top) {
        return Err(format!("codex ignores {top} in a project config; refused"));
    }
    let invalid = || format!("invalid codex config value for {key}");
    let list = |names: &Value, file: bool| -> Result<()> {
        let items = names.as_array().ok_or_else(invalid)?;
        if items.len() > MAX_LIST {
            return Err(invalid());
        }
        for item in items {
            let s = item.as_str().ok_or_else(invalid)?;
            let bad_file = file && (s.contains(['/', '\\']) || s == "." || s == "..");
            if !is_name(s) || bad_file {
                return Err(invalid());
            }
        }
        Ok(())
    };
    match key {
        "tui.status_line" => list(value, false),
        "project_doc_fallback_filenames" => list(value, true),
        "skills.bundled.enabled" => value.as_bool().map(|_| ()).ok_or_else(invalid),
        "project_doc_max_bytes" => match value.as_i64() {
            Some(n) if (1..=MAX_DOC_BYTES).contains(&n) => Ok(()),
            _ => Err(invalid()),
        },
        _ => Err("unmanaged codex config key refused".into()),
    }
}

fn review() -> String {
    format!("existing {CONFIG_FILE} requires preservation review")
}

/// The table that holds `key`'s last segment, or `None` when a parent is missing.
fn parent<'a>(doc: &'a DocumentMut, key: &str) -> Result<Option<&'a dyn TableLike>> {
    let mut table: &dyn TableLike = doc.as_table();
    let segments: Vec<&str> = key.split('.').collect();
    for seg in &segments[..segments.len() - 1] {
        match table.get(seg) {
            None => return Ok(None),
            Some(item) => table = item.as_table_like().ok_or_else(review)?,
        }
    }
    Ok(Some(table))
}

fn leaf(key: &str) -> &str {
    key.rsplit('.').next().unwrap_or(key)
}

/// JSON form of a TOML value for the three-way comparison. A value MoAI never writes (a
/// float, a date, a table) becomes a marker no desired value can equal.
fn to_json(v: &toml_edit::Value) -> Value {
    match v {
        toml_edit::Value::Boolean(b) => json!(*b.value()),
        toml_edit::Value::Integer(i) => json!(*i.value()),
        toml_edit::Value::String(s) => json!(s.value()),
        toml_edit::Value::Array(a) => Value::Array(a.iter().map(to_json).collect()),
        other => json!({"$toml": other.to_string().trim()}),
    }
}

/// Current values of `keys` in `doc`; a key whose item is not a plain value needs review.
fn read_keys(doc: &DocumentMut, keys: &[&String]) -> Result<Map<String, Value>> {
    let mut out = Map::new();
    for key in keys {
        let Some(table) = parent(doc, key)? else {
            continue;
        };
        if let Some(item) = table.get(leaf(key)) {
            let value = item.as_value().ok_or_else(review)?;
            out.insert((*key).clone(), to_json(value));
        }
    }
    Ok(out)
}

fn to_toml(v: &Value) -> toml_edit::Value {
    match v {
        Value::Bool(b) => (*b).into(),
        Value::Number(n) => n.as_i64().expect("validated integer").into(),
        Value::String(s) => s.as_str().into(),
        Value::Array(items) => {
            let mut a = Array::new();
            for i in items {
                a.push(to_toml(i));
            }
            toml_edit::Value::Array(a)
        }
        _ => unreachable!("validated value"),
    }
}

/// Sets `key` to `value`, creating missing parent tables. An existing key keeps its position
/// and the comments around it.
fn set_key(doc: &mut DocumentMut, key: &str, value: &Value) -> Result<()> {
    let segments: Vec<&str> = key.split('.').collect();
    let mut table: &mut dyn TableLike = doc.as_table_mut();
    for (i, seg) in segments[..segments.len() - 1].iter().enumerate() {
        if table.get(seg).is_none() {
            let mut t = Table::new();
            // Only the table that holds the value gets a header line.
            t.set_implicit(i + 2 < segments.len());
            table.insert(seg, Item::Table(t));
        }
        table = table
            .get_mut(seg)
            .and_then(Item::as_table_like_mut)
            .ok_or_else(review)?;
    }
    let mut new = to_toml(value);
    match table.get_mut(leaf(key)) {
        Some(item) => {
            let old = item.as_value().ok_or_else(review)?;
            *new.decor_mut() = old.decor().clone();
            *item = Item::Value(new);
        }
        None => {
            table.insert(leaf(key), Item::Value(new));
        }
    }
    Ok(())
}

fn remove_key(doc: &mut DocumentMut, key: &str) -> Result<()> {
    let segments: Vec<&str> = key.split('.').collect();
    let mut table: &mut dyn TableLike = doc.as_table_mut();
    for seg in &segments[..segments.len() - 1] {
        match table.get_mut(seg) {
            None => return Ok(()),
            Some(item) => table = item.as_table_like_mut().ok_or_else(review)?,
        }
    }
    table.remove(leaf(key));
    Ok(())
}

/// The per-key three-way table (same rule as `claude::plan_settings`): returns the changes to
/// make and the conflicting keys. Missing and present values are distinct.
type Changes = Vec<(String, Option<Value>)>;
fn three_way(
    baseline: &Map<String, Value>,
    current: &Map<String, Value>,
    desired: &Map<String, Value>,
) -> (Changes, Vec<String>) {
    let mut keys: Vec<&String> = baseline.keys().chain(desired.keys()).collect();
    keys.sort();
    keys.dedup();
    let (mut changes, mut conflicts) = (Vec::new(), Vec::new());
    for key in keys {
        let (old, now, new) = (baseline.get(key), current.get(key), desired.get(key));
        if now == new || new == old {
            continue;
        }
        if now != old {
            conflicts.push(key.clone());
        } else {
            changes.push((key.clone(), new.cloned()));
        }
    }
    (changes, conflicts)
}

fn read_existing(path: &Path) -> Result<Option<Vec<u8>>> {
    match fs::symlink_metadata(path) {
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(None),
        Err(_) => Err("target metadata unavailable".into()),
        Ok(m) if m.file_type().is_symlink() => Err("symlink target refused".into()),
        Ok(m) if !m.is_file() || m.len() > CONFIG_LIMIT => {
            Err("target is not a bounded regular file".into())
        }
        Ok(_) => fs::read(path)
            .map(Some)
            .map_err(|_| "target read failed".into()),
    }
}

fn refuse_symlink_dir(path: &Path) -> Result<()> {
    match fs::symlink_metadata(path) {
        Ok(m) if m.file_type().is_symlink() => Err("symlink target refused".into()),
        _ => Ok(()),
    }
}

struct ConfigState {
    config: PathBuf,
    receipt: PathBuf,
    before: Option<Vec<u8>>,
    receipt_before: Option<Vec<u8>>,
    owned: Map<String, Value>,
    current: Map<String, Value>,
    conflicts: Vec<String>,
    after: Option<Vec<u8>>,
    digest: String,
}

fn config_state(project: &Path, home: &Path, desired: &Map<String, Value>) -> Result<ConfigState> {
    for (k, v) in desired {
        validate(k, v)?;
    }
    let root = fs::canonicalize(project).map_err(|_| "project directory unavailable")?;
    if !root.is_dir() {
        return Err("project must be a directory".into());
    }
    // `<home>/.codex/config.toml` is the user config: writing it is the blocked home write.
    if fs::canonicalize(home).is_ok_and(|h| h == root) {
        return Err("project is the home directory; the user codex config is refused".into());
    }
    let state_dir = crate::state_root::path(&root, "");
    for dir in [
        root.join(".codex"),
        state_dir.clone(),
        state_dir.join("manifests"),
    ] {
        refuse_symlink_dir(&dir)?;
    }
    let (config, receipt) = (root.join(CONFIG_FILE), state_dir.join(CONFIG_RECEIPT));
    let before = read_existing(&config)?;
    let receipt_before = read_existing(&receipt)?;
    let owned = match &receipt_before {
        None => Map::new(),
        Some(bytes) => resolve_owned(bytes, before.as_deref().map(crate::digest))?,
    };
    if owned
        .keys()
        .any(|k| !MANAGED_CODEX_KEYS.contains(&k.as_str()))
    {
        return Err("codex config receipt requires preservation review".into());
    }
    let mut doc = match &before {
        None => DocumentMut::new(),
        Some(bytes) => std::str::from_utf8(bytes)
            .ok()
            .and_then(|t| t.parse::<DocumentMut>().ok())
            .ok_or_else(review)?,
    };
    let mut keys: Vec<&String> = owned.keys().chain(desired.keys()).collect();
    keys.sort();
    keys.dedup();
    let current = read_keys(&doc, &keys)?;
    let (changes, conflicts) = three_way(&owned, &current, desired);
    let after = if conflicts.is_empty() && !changes.is_empty() {
        for (key, value) in &changes {
            match value {
                Some(v) => set_key(&mut doc, key, v)?,
                None => remove_key(&mut doc, key)?,
            }
        }
        let bytes = doc.to_string().into_bytes();
        (Some(&bytes) != before.as_ref()).then_some(bytes)
    } else {
        None
    };
    if after
        .as_ref()
        .is_some_and(|b| b.len() as u64 > CONFIG_LIMIT)
    {
        return Err("rendered codex config exceeds the 256 KiB limit; refused".into());
    }
    use crate::spec::Canonical;
    let canonical = crate::spec::SortedKeys.bytes(&Value::Object(desired.clone()));
    let digest = crate::digest(
        &serde_json::to_vec(&(
            1,
            crate::digest(&canonical),
            before.as_deref().map(crate::digest),
            receipt_before.as_deref().map(crate::digest),
            after.as_deref().map(crate::digest),
            &conflicts,
        ))
        .expect("plan tuple"),
    );
    Ok(ConfigState {
        config,
        receipt,
        before,
        receipt_before,
        owned,
        current,
        conflicts,
        after,
        digest,
    })
}

/// Requested `tui.status_line` items that are not Codex 0.169.0 identifiers. They are still
/// written (Codex ignores an unknown one with a warning, and a newer Codex may know it), but
/// reported one by one, never dropped.
fn unverified(desired: &Map<String, Value>) -> Value {
    let unknown: Vec<&Value> = desired
        .get("tui.status_line")
        .and_then(Value::as_array)
        .into_iter()
        .flatten()
        .filter(|id| {
            !id.as_str()
                .is_some_and(|s| CODEX_STATUS_LINE_ITEMS.contains(&s))
        })
        .collect();
    let reason = (!unknown.is_empty()).then_some(STATUS_LINE_UNVERIFIED);
    json!({"items": unknown, "reason": reason})
}

/// `tui.status_line` in a config the plan already parsed, or `null`.
fn status_line_in(bytes: Option<&[u8]>) -> Value {
    let doc = bytes
        .and_then(|b| std::str::from_utf8(b).ok())
        .and_then(|t| t.parse::<DocumentMut>().ok());
    let key = "tui.status_line".to_string();
    doc.and_then(|d| read_keys(&d, &[&key]).ok())
        .and_then(|mut m| m.remove(&key))
        .unwrap_or(Value::Null)
}

/// Before and after of `tui.status_line` and each requested item's status (#2928 S-5).
fn status_line_plan(s: &ConfigState, desired: &Map<String, Value>, preset: Option<&str>) -> Value {
    let before = status_line_in(s.before.as_deref());
    let after = match &s.after {
        Some(bytes) => status_line_in(Some(bytes)),
        None => before.clone(),
    };
    let items: Vec<Value> = desired
        .get("tui.status_line")
        .and_then(Value::as_array)
        .into_iter()
        .flatten()
        .map(|id| {
            let known = id
                .as_str()
                .is_some_and(|s| CODEX_STATUS_LINE_ITEMS.contains(&s));
            json!({"id": id, "status": if known { "known" } else { "unknown_item" }})
        })
        .collect();
    json!({"before": before, "after": after, "preset": preset, "items": items})
}

/// Previews setting `desired` (managed dotted keys only) in `PROJECT/.codex/config.toml`.
/// Never writes. `home` is the user home; a project equal to it is refused.
/// `"tui.status_line": "moai"` names [`MOAI_STATUS_LINE_PRESET`].
pub fn config_plan(project: &Path, home: &Path, desired: &Map<String, Value>) -> Result<Value> {
    let (desired, preset) = resolve_preset(desired)?;
    let desired = &desired;
    let s = config_state(project, home, desired)?;
    let action = match (s.conflicts.is_empty(), &s.after, &s.before) {
        (false, _, _) => "conflict",
        (true, None, _) => "unchanged",
        (true, Some(_), None) => "create",
        (true, Some(_), Some(_)) => "update",
    };
    Ok(json!({
        "action": action,
        "config_path": CONFIG_FILE,
        "conflicts": s.conflicts,
        "expected_sha256": s.before.as_deref().map(crate::digest),
        "planned_sha256": s.after.as_deref().map(crate::digest),
        "plan_digest": s.digest,
        "trust": {"status": "not_evaluated", "declaration": TRUST_DECLARATION, "writes": false},
        "status_line": status_line_plan(&s, desired, preset),
        "status_line_unverified": unverified(desired),
        "writes": false,
    }))
}

fn receipt_bytes(
    owned: &Map<String, Value>,
    config_sha256: Option<&str>,
    pending: Option<(&Map<String, Value>, &str)>,
) -> Vec<u8> {
    let mut r = json!({
        "schema_version": 1,
        "config_path": CONFIG_FILE,
        "owned": owned,
        "config_sha256": config_sha256,
    });
    if let Some((owned, sha)) = pending {
        r["pending"] = json!({"owned": owned, "config_sha256": sha});
    }
    let mut bytes = serde_json::to_vec_pretty(&r).expect("receipt");
    bytes.push(b'\n');
    bytes
}

/// Ownership for the config whose sha256 is `current`; a pending receipt resolves by which
/// side the file is on (as `claude::resolve_owned`).
fn resolve_owned(bytes: &[u8], current: Option<String>) -> Result<Map<String, Value>> {
    let review = || "codex config receipt requires preservation review".to_string();
    let r: Value = serde_json::from_slice(bytes).map_err(|_| review())?;
    let Value::Object(owned) = &r["owned"] else {
        return Err(review());
    };
    if r["schema_version"].as_u64() != Some(1) {
        return Err(review());
    }
    let Some(pending) = r.get("pending") else {
        return Ok(owned.clone());
    };
    let (Value::Object(intended), Value::String(intended_sha)) =
        (&pending["owned"], &pending["config_sha256"])
    else {
        return Err(review());
    };
    if current.as_deref() == Some(intended_sha.as_str()) {
        Ok(intended.clone())
    } else if current.as_deref() == r["config_sha256"].as_str() {
        Ok(owned.clone())
    } else {
        Err(review())
    }
}

/// Applies a previewed plan: re-plans against the files now, requires the same digest,
/// replaces the config only when it changes, then records ownership (pending first, as the
/// Claude settings apply does). A value the user already had is left unowned.
pub fn config_apply(
    project: &Path,
    home: &Path,
    desired: &Map<String, Value>,
    plan_digest: &str,
    permission: &str,
) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err("permission reference required".into());
    }
    let (desired, _) = resolve_preset(desired)?;
    let desired = &desired;
    let s = config_state(project, home, desired)?;
    if !s.conflicts.is_empty() {
        return Err(format!(
            "conflict in {CONFIG_FILE}: {} changed outside MoAI; nothing applied",
            s.conflicts.join(", ")
        ));
    }
    if s.digest != plan_digest {
        return Err("stale or altered plan; re-preview required".into());
    }
    let owned: Map<String, Value> = desired
        .iter()
        .filter(|(k, v)| s.owned.contains_key(*k) || s.current.get(*k) != Some(*v))
        .map(|(k, v)| (k.clone(), v.clone()))
        .collect();
    let before_sha = s.before.as_deref().map(crate::digest);
    let final_bytes = s.after.as_deref().or(s.before.as_deref());
    let receipt = receipt_bytes(&owned, final_bytes.map(crate::digest).as_deref(), None);
    let rel = crate::state_root::rel(project, CONFIG_RECEIPT);
    let mut files = Vec::new();
    let mut receipt_now = s.receipt_before.clone();
    if let Some(after) = &s.after {
        let after_sha = crate::digest(after);
        let pending = receipt_bytes(&s.owned, before_sha.as_deref(), Some((&owned, &after_sha)));
        crate::agents::write_replace(&s.receipt, receipt_now.as_deref(), &pending)?;
        receipt_now = Some(pending);
        crate::agents::write_replace(&s.config, s.before.as_deref(), after)?;
        files.push(CONFIG_FILE);
    }
    if receipt_now.as_deref() != Some(receipt.as_slice()) {
        if let Err(e) = crate::agents::write_replace(&s.receipt, receipt_now.as_deref(), &receipt) {
            return Err(if files.is_empty() {
                e
            } else {
                format!("partial apply: applied [{CONFIG_FILE}]; failed at {rel}: {e}")
            });
        }
    }
    Ok(json!({
        "state": if files.is_empty() { "unchanged" } else { "applied" },
        "files_changed": files,
        "files_created": if s.before.is_none() { files.clone() } else { vec![] },
        "owned_keys": owned.keys().collect::<Vec<_>>(),
        "config_sha256_before": before_sha,
        "config_sha256_after": final_bytes.map(crate::digest),
        "trust": {"status": "not_evaluated", "declaration": TRUST_DECLARATION, "writes": false},
        "host_loaded": "not_run",
        "execution_verified": "not_run",
    }))
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::contract::{classify, Kind};

    /// This module's error messages map to a deliberate kind (#2254: module-local, not a row
    /// in the shared `contract.rs` list). Validation messages come from `validate` itself.
    #[test]
    fn every_message_has_a_deliberate_kind() {
        let refused = |key: &str, value: Value| validate(key, &value).unwrap_err();
        let cases = [
            (review(), Kind::Conflict),
            (
                "codex config receipt requires preservation review".into(),
                Kind::Conflict,
            ),
            (
                format!("conflict in {CONFIG_FILE}: tui.status_line changed outside MoAI; nothing applied"),
                Kind::Conflict,
            ),
            (
                format!("partial apply: applied [{CONFIG_FILE}]; failed at .moai-cli/{CONFIG_RECEIPT}: x"),
                Kind::Partial,
            ),
            (refused("projects./x.trust_level", json!("trusted")), Kind::InvalidInput),
            (refused("model_provider", json!("x")), Kind::InvalidInput),
            (refused("project_doc_max_bytes", json!(0)), Kind::InvalidInput),
            (refused("model", json!("x")), Kind::InvalidInput),
            (
                "project is the home directory; the user codex config is refused".into(),
                Kind::InvalidInput,
            ),
            (
                "rendered codex config exceeds the 256 KiB limit; refused".into(),
                Kind::InvalidInput,
            ),
            ("symlink target refused".into(), Kind::InvalidInput),
            ("target is not a bounded regular file".into(), Kind::InvalidInput),
        ];
        for (message, kind) in cases {
            assert_eq!(classify(&message), kind, "{message}");
        }
    }
}
