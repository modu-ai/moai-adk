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

/// Why a key whose value looks like a credential is not carried. Its value is not shown anywhere:
/// it goes to a file that is usually committed (audit 6888156222 B1).
const CREDENTIAL_BLOCK: &str =
    "a value in it looks like a credential; the key stays in the backup and its value is not shown";

/// Why an Opus 4.8 model value beside a fast-mode key is not carried (SPEC-20261122-model-fast-pair
/// REQ-3, D-3, D-6): the API refuses that pair from 2026-11-30, so the value stays in the backup.
/// A fixed text: it names no saved value, and the date is words in it, never compared with a clock.
const FAST_PAIR_BLOCK: &str = "the model is Opus 4.8 and a fast-mode key sits beside it; the API refuses that pair from 2026-11-30, so the model stays in the backup; switch to Opus 5.5 or turn fast mode off";

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
    if claude::holds_credential(&carry) {
        return Row::other("permissions", "blocked", CREDENTIAL_BLOCK, summary);
    }
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

/// `env` names the Go generation set for its own `moai`: a merged `PATH` that lists the Go bin
/// folder would put the Go `moai` ahead of this one again, and `MOAI_CONFIG_SOURCE` points the
/// Go program at its own settings. They are never offered back (leader decision 6887271793, F29).
const GO_ENV_NAMES: [&str; 2] = ["PATH", "MOAI_CONFIG_SOURCE"];
const GO_ENV_WHY: &str = "set for the Go generation's moai; it stays in the backup";

fn env_row(v: &Value) -> Row {
    let Value::Object(map) = v else {
        return Row::other("env", "kept_in_backup", "env is not an object", json!({}));
    };
    let mut carry = Map::new();
    let mut refused: Vec<Value> = Vec::new();
    for (name, val) in map {
        if GO_ENV_NAMES.contains(&name.as_str()) {
            refused.push(json!({ "name": shown(name), "why": GO_ENV_WHY }));
            continue;
        }
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

/// An Opus 4.8 model id in any letter case (`opus-4-8`, `opus-4.8`, an alias suffix included). The
/// match ends there or goes on with `-`, `[` or `@` (a date, a context alias, a version tag); any
/// other next character, a digit above all, is another model (`opus-4-80`, `opus-4.85`), which
/// SPEC D-2 leaves alone. A space around the value is cut first (`one_line` accepts it, and
/// whether the host trims is unverified), so `opus-4-8 ` still ends the match; the carried value
/// is not touched (SPEC-20261124-model-value-trim).
fn is_opus_48(model: &str) -> bool {
    let lower = model.trim().to_ascii_lowercase();
    ["opus-4-8", "opus-4.8"].iter().any(|family| {
        lower.match_indices(family).any(|(at, _)| {
            matches!(
                lower[at + family.len()..].chars().next(),
                None | Some('-' | '[' | '@')
            )
        })
    })
}

fn model_row(v: &Value, old: &Map<String, Value>) -> Row {
    if claude::holds_credential(v) {
        return Row::other("model", "blocked", CREDENTIAL_BLOCK, json!({}));
    }
    let fast_key = old.contains_key("fastMode") || old.contains_key("fastModePerSessionOptIn");
    if fast_key && v.as_str().is_some_and(is_opus_48) {
        return Row::other("model", "blocked", FAST_PAIR_BLOCK, json!({}));
    }
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
    if claude::holds_credential(v) {
        return Row::other("statusLine", "blocked", CREDENTIAL_BLOCK, json!({}));
    }
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
    if claude::holds_credential(v) {
        return Row::other("outputStyle", "blocked", CREDENTIAL_BLOCK, json!({}));
    }
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
                "model" => model_row(v, old),
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
    let root =
        std::fs::canonicalize(project).map_err(|_| crate::contract::PROJECT_PATH_UNAVAILABLE)?;
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
    fn a_credential_blocks_its_key_whatever_the_key_is() {
        let secret = format!("{}{}", "ghp_", "a".repeat(30));
        let r = rows(json!({
            "permissions": {"allow": ["Bash(git status)", format!("Bash(curl {secret})")]},
            "model": format!("m-{secret}"),
            "statusLine": {"type": "command", "command": format!("s --t {secret}")},
            "outputStyle": format!("moai-{secret}"),
            "env": {"EDITOR": "vim"},
        }));
        for key in ["permissions", "model", "statusLine", "outputStyle"] {
            let row = r.iter().find(|r| r.key == key).unwrap();
            assert_eq!(row.state, "blocked", "{key}");
            assert!(row.carry.is_none(), "{key}");
            let shown = format!("{} {}", row.reason, row.summary);
            assert!(!shown.contains(&secret), "{key}: {shown}");
        }
        assert_eq!(state(&r, "env"), "mergeable");
        // A URL with a password counts too.
        let r = rows(json!({"model": "https://u:p@host/m"}));
        assert_eq!(state(&r, "model"), "blocked");
    }

    #[test]
    fn the_go_generations_path_is_never_carried_back() {
        let r = rows(json!({"env": {
            "EDITOR": "vim", "PATH": "/home/u/go/bin:/usr/bin", "MOAI_CONFIG_SOURCE": "x",
        }}));
        let env = r.iter().find(|r| r.key == "env").unwrap();
        assert_eq!(env.carry.as_ref().unwrap(), &json!({"EDITOR": "vim"}));
        let shown = env.summary.to_string();
        assert!(shown.contains("PATH") && shown.contains("MOAI_CONFIG_SOURCE"));
        assert!(!shown.contains("/home/u/go/bin"));

        let only = rows(json!({"env": {"PATH": "/home/u/go/bin:/usr/bin"}}));
        let env = only.iter().find(|r| r.key == "env").unwrap();
        assert_eq!(env.state, "blocked");
        assert!(env.carry.is_none());
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

    // SPEC-20261122-model-fast-pair REQ-3: the API refuses an Opus 4.8 request with fast mode from
    // 2026-11-30, so a backup that holds both is never offered for a merge.
    fn model_row_of(v: Value) -> Row {
        let mut r = rows(v);
        let at = r.iter().position(|r| r.key == "model").unwrap();
        r.swap_remove(at)
    }

    #[test]
    fn fast_pair_an_opus_48_model_beside_a_fast_mode_key_is_blocked_with_the_fixed_reason() {
        // An id ends, or goes on with `-`, `[` or `@`, right after the `4-8` / `4.8`.
        for model in [
            "claude-opus-4-8",
            "opus-4.8",
            "Claude-OPUS-4-8[1m]",
            "claude-opus-4-8-20261001",
            "claude-opus-4-8@latest",
            "opus-4.8-preview",
        ] {
            for key in ["fastMode", "fastModePerSessionOptIn"] {
                let mut old = Map::new();
                old.insert("model".into(), json!(model));
                old.insert(key.into(), json!(true));
                let row = model_row_of(Value::Object(old));
                assert_eq!(row.state, "blocked", "{model} + {key}");
                assert!(row.carry.is_none(), "{model} + {key}");
                assert_eq!(row.reason, FAST_PAIR_BLOCK, "{model} + {key}");
                // Card criterion 2: one plain notice that names the date and the fix.
                assert!(row.reason.contains("2026-11-30"), "{}", row.reason);
                assert!(row.reason.contains("Opus 5.5"), "{}", row.reason);
                assert!(row.reason.contains("fast mode"), "{}", row.reason);
                let shown = format!("{} {}", row.reason, row.summary);
                assert!(!shown.contains(model), "{model} + {key}: {shown}");
            }
        }
    }

    #[test]
    fn fast_pair_a_model_alone_or_another_family_still_merges() {
        let alone = model_row_of(json!({"model": "claude-opus-4-8"}));
        assert_eq!(alone.state, "mergeable");
        assert_eq!(alone.carry, Some(json!("claude-opus-4-8")));
        let other = model_row_of(json!({"model": "claude-opus-5-5", "fastMode": true}));
        assert_eq!(other.state, "mergeable");
        assert_eq!(other.carry, Some(json!("claude-opus-5-5")));
        // A near id is not the family (SPEC D-2): a digit right after `4-8` / `4.8` makes another
        // model, so the pair refusal, whose reason says "Opus 4.8", must not fire.
        for near in ["claude-opus-4-80", "claude-opus-4-7", "claude-opus-4.85"] {
            for key in ["fastMode", "fastModePerSessionOptIn"] {
                let mut old = Map::new();
                old.insert("model".into(), json!(near));
                old.insert(key.into(), json!(true));
                let row = model_row_of(Value::Object(old));
                assert_eq!(row.state, "mergeable", "{near} + {key}");
                assert_eq!(row.carry, Some(json!(near)), "{near} + {key}");
            }
        }
    }

    // SPEC-20261124-model-value-trim: `one_line` accepts a space around the model value, so the
    // pair check must see the family through it.
    fn pair_row(model: &str, key: &str) -> Row {
        let mut old = Map::new();
        old.insert("model".into(), json!(model));
        old.insert(key.into(), json!(true));
        model_row_of(Value::Object(old))
    }

    #[test]
    fn fast_pair_a_space_around_an_opus_48_value_does_not_hide_it() {
        for model in [
            " claude-opus-4-8",
            "claude-opus-4-8 ",
            " opus-4.8 ",
            "Claude-OPUS-4-8[1m] ",
        ] {
            for key in ["fastMode", "fastModePerSessionOptIn"] {
                let row = pair_row(model, key);
                assert_eq!(row.state, "blocked", "{model:?} + {key}");
                assert!(row.carry.is_none(), "{model:?} + {key}");
                assert_eq!(row.reason, FAST_PAIR_BLOCK, "{model:?} + {key}");
            }
        }
    }

    #[test]
    fn fast_pair_a_space_keeps_a_near_id_mergeable_and_the_carried_text_as_read() {
        for near in ["claude-opus-4-80 ", "claude-opus-4.85 ", " claude-opus-4-7 "] {
            for key in ["fastMode", "fastModePerSessionOptIn"] {
                let row = pair_row(near, key);
                assert_eq!(row.state, "mergeable", "{near:?} + {key}");
                assert_eq!(row.carry, Some(json!(near)), "{near:?} + {key}");
            }
        }
        // No fast-mode key: the value merges, and the spaces stay in the carried text.
        let alone = model_row_of(json!({"model": " claude-opus-4-8 "}));
        assert_eq!(alone.state, "mergeable");
        assert_eq!(alone.carry, Some(json!(" claude-opus-4-8 ")));
    }

    #[test]
    fn fast_pair_never_carries_a_fast_mode_key_itself() {
        let r = rows(json!({"model": "claude-opus-4-8", "fastMode": true}));
        assert!(r.iter().all(|r| r.carry.is_none()));
        assert_eq!(state(&r, "fastMode"), "kept_in_backup");
    }
}
