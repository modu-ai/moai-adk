//! The earlier project settings as merge candidates (#2936 R8).
//!
//! When `go-state-apply` moved the Go generation's `.claude/settings.json` into its backup, the
//! install step writes a new one. What the user had set there is not lost: it is still in the
//! backup, and this module reads that copy and says, key by key, what could be merged back.
//! It never writes. A merge is the ordinary `claude-settings-apply` with the chosen keys, so the
//! ownership receipt, the three-way plan and the digest all stay the ones that file already has.
//!
//! Only a few keys can be carried (`permissions`, `env`, `model`, a user `statusLine`, a user
//! `outputStyle`). The rest is shown and stays in the backup: `hooks` is MoAI's as a whole key
//! (`hooks-apply` writes it, a user hook there is a conflict, never a merge), and what the Go
//! generation set up itself (its status line script, its hooks, its styles) is not the user's.
//! A value that looks like a credential is never shown or carried, only its key.

use crate::{apply, claude, go_state};
use serde_json::{json, Map, Value};
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// The size of a status line object (as JSON) a merge carries.
const STATUS_LINE_BYTES_MAX: usize = 1024;
/// The size of an output style name a merge carries.
const STYLE_CHARS_MAX: usize = 128;
/// A path that marks something the Go generation set up itself.
const GO_HOOK_DIR: &str = ".claude/hooks/moai/";
const GO_STATE_DIR: &str = ".moai/";

/// The keys a merge may carry, in the order they are shown.
pub const MERGEABLE: [&str; 5] = ["permissions", "env", "model", "statusLine", "outputStyle"];

/// What was read from the backup copy.
#[derive(Debug)]
pub struct Candidates {
    /// The apply stamp of the backup that holds the copy.
    pub backup: String,
    /// The project-relative path of the copy.
    pub rel: String,
    pub sha256: String,
    /// One row per key of the old file: `key`, `state`, `reason`, `summary`.
    pub rows: Vec<Value>,
    /// The values a merge carries, by key (only keys whose row is `mergeable`).
    pub carry: Map<String, Value>,
}

struct Row {
    key: String,
    state: &'static str,
    reason: String,
    summary: Value,
    carry: Option<Value>,
}

impl Row {
    fn mergeable(key: &str, summary: Value, carry: Value, reason: &str) -> Row {
        Row {
            key: key.into(),
            state: "mergeable",
            reason: reason.into(),
            summary,
            carry: Some(carry),
        }
    }

    fn other(key: &str, state: &'static str, reason: &str, summary: Value) -> Row {
        Row {
            key: key.into(),
            state,
            reason: reason.into(),
            summary,
            carry: None,
        }
    }
}

fn one_line(s: &str, max: usize) -> bool {
    !s.trim().is_empty() && s.chars().count() <= max && !s.chars().any(char::is_control)
}

fn goes_back_to_go(text: &str) -> bool {
    text.contains(GO_STATE_DIR) || text.contains(GO_HOOK_DIR) || text.contains("status_line.sh")
}

/// Whether the writer accepts this `defaultMode` (only the narrowing modes).
fn mode_ok(mode: &Value) -> bool {
    let mut one = Map::new();
    one.insert("defaultMode".into(), mode.clone());
    claude::check_permissions(&Value::Object(one)).is_ok()
}

fn permissions_row(v: &Value) -> Row {
    let Value::Object(map) = v else {
        return Row::other(
            "permissions",
            "kept_in_backup",
            "permissions is not an object",
            json!({}),
        );
    };
    let mut carry = Map::new();
    let mut left_out: Vec<String> = Vec::new();
    for (k, val) in map {
        match k.as_str() {
            "allow" | "deny" | "ask" => {
                carry.insert(k.clone(), val.clone());
            }
            "defaultMode" if mode_ok(val) => {
                carry.insert(k.clone(), val.clone());
            }
            other => left_out.push(other.chars().filter(|c| !c.is_control()).take(64).collect()),
        }
    }
    left_out.sort();
    let count = |k: &str| carry.get(k).and_then(Value::as_array).map_or(0, Vec::len);
    let summary = json!({
        "allow": count("allow"), "deny": count("deny"), "ask": count("ask"),
        "default_mode": carry.get("defaultMode"), "left_in_backup": left_out,
    });
    if carry.is_empty() {
        return Row::other(
            "permissions",
            "kept_in_backup",
            "nothing in permissions can be carried (only allow, deny, ask and a narrowing \
             defaultMode are)",
            summary,
        );
    }
    let carry = Value::Object(carry);
    match claude::check_permissions(&carry) {
        Ok(()) => Row::mergeable(
            "permissions",
            summary,
            carry,
            "the rules would be written to the committed settings file; widening ones too, so \
             read them first",
        ),
        Err(e) => Row::other("permissions", "blocked", &e, summary),
    }
}

fn env_row(v: &Value) -> Row {
    let Value::Object(map) = v else {
        return Row::other("env", "kept_in_backup", "env is not an object", json!({}));
    };
    let mut carry = Map::new();
    let mut refused: Vec<Value> = Vec::new();
    for (name, val) in map {
        match claude::env_entry_refusal(name, val) {
            None if carry.len() < claude::ENV_ENTRIES_MAX => {
                carry.insert(name.clone(), val.clone());
            }
            None => refused.push(json!({
                "name": shown(name), "why": "more entries than a merge carries",
            })),
            Some(why) => refused.push(json!({ "name": shown(name), "why": why })),
        }
    }
    let names: Vec<&String> = carry.keys().collect();
    let summary = json!({ "carried_names": names, "refused": refused });
    if carry.is_empty() {
        let state = if refused.is_empty() {
            "kept_in_backup"
        } else {
            "blocked"
        };
        return Row::other("env", state, "no env entry can be carried", summary);
    }
    Row::mergeable(
        "env",
        summary,
        Value::Object(carry),
        "the entries would be written to the committed settings file; the refused ones stay in \
         the backup",
    )
}

fn shown(name: &str) -> String {
    name.chars().filter(|c| !c.is_control()).take(64).collect()
}

fn model_row(v: &Value) -> Row {
    match v.as_str() {
        Some(s) if one_line(s, claude::MODEL_CHARS_MAX) => Row::mergeable(
            "model",
            json!({ "model": s }),
            v.clone(),
            "the model name would be written to the committed settings file",
        ),
        _ => Row::other(
            "model",
            "kept_in_backup",
            "model is not a one-line string within the size limit",
            json!({}),
        ),
    }
}

fn status_line_row(v: &Value) -> Row {
    let command = v["command"].as_str().unwrap_or("");
    if goes_back_to_go(command) {
        return Row::other(
            "statusLine",
            "go_generated",
            "the status line is the Go generation's script; the install step writes the new one",
            json!({}),
        );
    }
    let bytes = serde_json::to_vec(v).map_or(usize::MAX, |b| b.len());
    if v["type"] == "command"
        && one_line(command, STATUS_LINE_BYTES_MAX)
        && bytes <= STATUS_LINE_BYTES_MAX
    {
        return Row::mergeable(
            "statusLine",
            json!({ "command": command }),
            v.clone(),
            "replaces the MoAI status line the install step wrote with the user's own",
        );
    }
    Row::other(
        "statusLine",
        "kept_in_backup",
        "the status line is not a short command status line",
        json!({}),
    )
}

fn output_style_row(v: &Value) -> Row {
    let Some(name) = v.as_str().filter(|s| one_line(s, STYLE_CHARS_MAX)) else {
        return Row::other(
            "outputStyle",
            "kept_in_backup",
            "outputStyle is not a one-line name within the size limit",
            json!({}),
        );
    };
    if name.to_ascii_lowercase().starts_with("moai") {
        return Row::other(
            "outputStyle",
            "go_generated",
            "a MoAI style of the Go generation; its file is not installed by this generation",
            json!({ "name": name }),
        );
    }
    Row::mergeable(
        "outputStyle",
        json!({ "name": name }),
        v.clone(),
        "the style name would be written to the committed settings file",
    )
}

fn hooks_row(v: &Value) -> Row {
    let mut user_groups = 0usize;
    let mut events: Vec<(String, usize)> = Vec::new();
    if let Value::Object(map) = v {
        for (event, groups) in map {
            let n = groups.as_array().map_or(0, |gs| {
                gs.iter()
                    .filter(|g| !goes_back_to_go(&g.to_string()))
                    .count()
            });
            if n > 0 {
                events.push((shown(event), n));
                user_groups += n;
            }
        }
    }
    events.sort();
    if user_groups == 0 {
        return Row::other(
            "hooks",
            "go_generated",
            "only hooks of the Go generation; the hook set is written by hooks-apply",
            json!({}),
        );
    }
    let events: Map<String, Value> = events
        .into_iter()
        .map(|(e, n)| (e, Value::from(n)))
        .collect();
    Row::other(
        "hooks",
        "kept_in_backup",
        "hooks is MoAI's as a whole key (hooks-apply); the hooks the user wrote stay in the \
         backup",
        json!({ "user_groups": user_groups, "events": events }),
    )
}

fn rows_of(old: &Map<String, Value>) -> Vec<Row> {
    let mut keys: Vec<&String> = old.keys().collect();
    keys.sort();
    keys.into_iter()
        .map(|k| {
            let v = &old[k];
            match k.as_str() {
                "permissions" => permissions_row(v),
                "env" => env_row(v),
                "model" => model_row(v),
                "statusLine" => status_line_row(v),
                "outputStyle" => output_style_row(v),
                "hooks" => hooks_row(v),
                other => Row::other(
                    &shown(other),
                    "kept_in_backup",
                    "MoAI does not write this key; it stays in the backup",
                    json!({}),
                ),
            }
        })
        .collect()
}

/// The merge candidates from the settings the newest committed `go-state-apply` moved into its
/// backup; `Ok(None)` when none did (the settings stayed, or there was no such apply). Reads
/// only. An error means the copy cannot be trusted: a link on the way, a size over the limit,
/// bytes that no longer match the receipt, or something that is not a JSON object.
pub fn candidates(project: &Path) -> Result<Option<Candidates>> {
    let root = std::fs::canonicalize(project).map_err(|_| "project folder unavailable")?;
    let Some(found) = go_state::backed_up_settings(&root)? else {
        return Ok(None);
    };
    let path = apply::target(&root, &found.rel).map_err(|e| format!("{}: {e}", found.rel))?;
    let bytes = apply::read_opt(&path)
        .map_err(|e| format!("{}: {e}", found.rel))?
        .ok_or_else(|| format!("{} is missing", found.rel))?;
    if crate::digest(&bytes) != found.sha256 {
        return Err(format!(
            "{} changed since the apply moved it (sha256 differs from the receipt)",
            found.rel
        ));
    }
    let Ok(Value::Object(old)) = serde_json::from_slice::<Value>(&bytes) else {
        return Err(format!("{} is not a JSON object", found.rel));
    };
    let mut carry = Map::new();
    let mut rows = Vec::new();
    for row in rows_of(&old) {
        if let Some(v) = &row.carry {
            carry.insert(row.key.clone(), v.clone());
        }
        rows.push(json!({
            "key": row.key, "state": row.state, "reason": row.reason, "summary": row.summary,
        }));
    }
    Ok(Some(Candidates {
        backup: found.backup,
        rel: found.rel,
        sha256: found.sha256,
        rows,
        carry,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    fn rows(v: Value) -> Vec<Row> {
        let Value::Object(m) = v else { panic!() };
        rows_of(&m)
    }

    fn state<'a>(r: &'a [Row], key: &str) -> &'a str {
        r.iter()
            .find(|r| r.key == key)
            .map_or("absent", |r| r.state)
    }

    #[test]
    fn the_go_generation_is_told_apart_from_the_user() {
        let r = rows(json!({
            "statusLine": {"type": "command", "command": "bash \"$CLAUDE_PROJECT_DIR/.moai/status_line.sh\""},
            "outputStyle": "MoAI-Easy",
            "hooks": {"Stop": [{"hooks": [{"command": "bash", "args": ["${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/x.sh"]}]}]},
        }));
        assert_eq!(state(&r, "statusLine"), "go_generated");
        assert_eq!(state(&r, "outputStyle"), "go_generated");
        assert_eq!(state(&r, "hooks"), "go_generated");
    }

    #[test]
    fn the_users_own_values_are_offered() {
        let r = rows(json!({
            "statusLine": {"type": "command", "command": "my-status"},
            "outputStyle": "Explanatory",
            "model": "opus",
            "permissions": {"allow": ["Bash(git status)"], "deny": ["Read(./.env)"]},
            "env": {"EDITOR": "vim"},
        }));
        for key in MERGEABLE {
            assert_eq!(state(&r, key), "mergeable", "{key}");
        }
    }

    #[test]
    fn user_hooks_and_other_keys_are_shown_but_never_carried() {
        let r = rows(json!({
            "hooks": {"PreToolUse": [{"matcher": "Bash", "hooks": [{"type": "command", "command": "my-check"}]}]},
            "enabledPlugins": {"x": false},
        }));
        assert_eq!(state(&r, "hooks"), "kept_in_backup");
        assert_eq!(state(&r, "enabledPlugins"), "kept_in_backup");
        assert!(r.iter().all(|r| r.carry.is_none()));
    }

    #[test]
    fn a_credential_like_env_entry_is_named_but_its_value_is_not_kept() {
        let secret = format!("{}{}", "ghp_", "a".repeat(30));
        let r = rows(json!({"env": {"EDITOR": "vim", "MY_TOKEN": "x", "OTHER": secret}}));
        let env = r.iter().find(|r| r.key == "env").unwrap();
        assert_eq!(env.state, "mergeable");
        let carried = env.carry.as_ref().unwrap();
        assert_eq!(carried, &json!({"EDITOR": "vim"}));
        let shown = env.summary.to_string();
        assert!(shown.contains("MY_TOKEN") && shown.contains("OTHER"));
        assert!(!shown.contains(&secret));
    }

    #[test]
    fn permissions_carry_only_what_the_writer_accepts() {
        let r = rows(json!({"permissions": {
            "allow": ["Bash(ls)"], "defaultMode": "bypassPermissions",
            "additionalDirectories": ["/x"],
        }}));
        let p = r.iter().find(|r| r.key == "permissions").unwrap();
        assert_eq!(p.state, "mergeable");
        assert_eq!(p.carry.as_ref().unwrap(), &json!({"allow": ["Bash(ls)"]}));
        let left = p.summary["left_in_backup"].to_string();
        assert!(left.contains("additionalDirectories") && left.contains("defaultMode"));
    }
}
