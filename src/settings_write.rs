//! What a project settings write would be refused for once the checks enforce
//! (SPEC-20261122-settings-write-check). Observation only: nothing here changes a plan, its digest,
//! the written bytes or an existing refusal. This is the one place of the rules and of the entries
//! built from them. An entry names a rule and a key path or file, never a settings value (REQ-8).

use crate::claude::shown_value;
use crate::go_settings_clean::{numbers_that_change, repeats_a_key};
use serde_json::{json, Map, Value};
use std::fs;
use std::path::Path;
use std::time::{SystemTime, UNIX_EPOCH};

/// The field of a plan answer, an apply answer and an observation line that lists the entries.
pub(crate) const FIELD: &str = "would_refuse";
/// The settings observation log, in the `receipts` folder of the project state folder (REQ-9).
const LOG: &str = "settings-write.jsonl";

/// The fixed rule ids.
mod rule {
    // REQ-1: the documented shapes.
    pub const STATUS_LINE_NOT_OBJECT: &str = "status_line_not_object";
    pub const STATUS_LINE_TYPE: &str = "status_line_type";
    pub const STATUS_LINE_COMMAND: &str = "status_line_command";
    pub const STATUS_LINE_PADDING: &str = "status_line_padding";
    pub const STATUS_LINE_REFRESH_INTERVAL: &str = "status_line_refresh_interval";
    pub const STATUS_LINE_HIDE_VIM_MODE_INDICATOR: &str = "status_line_hide_vim_mode_indicator";
    pub const OUTPUT_STYLE_SHAPE: &str = "output_style_not_a_name";
    // REQ-2: the observed values.
    pub const PERMISSION_PATH_NOT_MATCHED: &str = "permission_path_not_matched";
    pub const BASH_WILDCARD_BEFORE_COMMAND: &str = "bash_wildcard_before_command";
    pub const ENV_VALUE_NOT_EXPANDED: &str = "env_value_not_expanded";
    pub const ENV_STORAGE_LOCATION: &str = "env_storage_location";
    pub const ENV_FEATURE_SWITCH: &str = "env_feature_switch";
    // REQ-6: content a rewrite loses.
    pub const REWRITE_REPEATED_KEY: &str = "rewrite_repeated_key";
    pub const REWRITE_NUMBER_SPELLING: &str = "rewrite_number_spelling";
}

/// Tools whose permission rules are accepted with a path but never match it.
const PATH_BLIND_TOOLS: [&str; 5] = ["Write", "NotebookEdit", "MultiEdit", "Glob", "Grep"];
/// The `env` names that move where files are stored, and the prefix of the XDG ones.
const STORAGE_NAMES: [&str; 6] = [
    "CLAUDE_CONFIG_DIR",
    "CLAUDE_CODE_TMPDIR",
    "HOME",
    "TMPDIR",
    "TMP",
    "TEMP",
];
const STORAGE_PREFIX: &str = "XDG_";
/// The `env` names that turn on agent teams or Code Mode.
const FEATURE_SWITCHES: [&str; 2] = [
    "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS",
    "CLAUDE_CODE_EXPERIMENTAL_CODE_MODE",
];

/// Whether the `env` entry turns a feature switch on: any value other than `0` or empty.
pub(crate) fn turns_on_switch(name: &str, value: &Value) -> bool {
    FEATURE_SWITCHES.contains(&name) && !matches!(value.as_str(), Some("" | "0"))
}

/// The entries of a write that turns the settings `file`, holding the bytes `before` (read as
/// `current`), into the bytes `after`. Empty when the file is not rewritten (`after` is `None`).
pub(crate) fn entries(
    file: &str,
    before: Option<&[u8]>,
    current: &Map<String, Value>,
    after: Option<&[u8]>,
) -> Vec<Value> {
    let Some(next) = after.and_then(|b| serde_json::from_slice::<Map<String, Value>>(b).ok())
    else {
        return Vec::new();
    };
    let mut out = Vec::new();
    status_line(current, &next, &mut out);
    output_style(current, &next, &mut out);
    env(current, &next, &mut out);
    permissions(current, &next, &mut out);
    if let Some(before) = before {
        lost_content(file, before, &mut out);
    }
    out
}

/// Appends the observation line of an apply that changed `file` to the log of `project`'s state
/// folder; never under HOME (REQ-9). The error is the receipt writer's fixed reason.
pub(crate) fn record(project: &Path, file: &str, entries: &[Value]) -> Result<(), &'static str> {
    let unix_ms = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map_or(0, |d| d.as_millis() as u64);
    let line = json!({"unix_ms": unix_ms, "file": file, FIELD: entries});
    create_private(project);
    crate::hooks::append_receipt_line_in(None, project, LOG, &line)
}

/// The `receipts` folder also keeps the replaced bytes of hook files, so the folder and the log
/// are owner-only from their creation when this line is the first to need them; the receipt writer
/// alone would leave them the umask's bits. Nothing is made through a link, what is already there
/// keeps its bits, and a failure is left to the writer, which reports it.
#[cfg(unix)]
fn create_private(project: &Path) {
    use std::os::unix::fs::{DirBuilderExt, OpenOptionsExt};
    let real_dir = |p: &Path| fs::symlink_metadata(p).is_ok_and(|m| m.is_dir());
    let receipts = crate::state_root::path(project, "receipts");
    if !receipts.parent().is_some_and(real_dir) {
        return;
    }
    let _ = fs::DirBuilder::new().mode(0o700).create(&receipts);
    if real_dir(&receipts) {
        let _ = fs::OpenOptions::new()
            .write(true)
            .create_new(true)
            .mode(0o600)
            .open(receipts.join(LOG));
    }
}

#[cfg(not(unix))]
fn create_private(_project: &Path) {}

fn entry(rule: &str, key: &str) -> Value {
    json!({"rule": rule, "key": key})
}

/// The value of `key` when the write adds it or changes it.
fn changed<'a>(
    current: &Map<String, Value>,
    next: &'a Map<String, Value>,
    key: &str,
) -> Option<&'a Value> {
    next.get(key).filter(|v| current.get(key) != Some(*v))
}

fn status_line(current: &Map<String, Value>, next: &Map<String, Value>, out: &mut Vec<Value>) {
    let Some(value) = changed(current, next, "statusLine") else {
        return;
    };
    let Some(map) = value.as_object() else {
        out.push(entry(rule::STATUS_LINE_NOT_OBJECT, "statusLine"));
        return;
    };
    // A field the shape does not name is not judged.
    let optional = |name: &str, ok: fn(&Value) -> bool| map.get(name).is_none_or(ok);
    let broken = [
        (
            map.get("type").and_then(Value::as_str) == Some("command"),
            rule::STATUS_LINE_TYPE,
            "statusLine.type",
        ),
        (
            map.get("command")
                .and_then(Value::as_str)
                .is_some_and(|c| !c.is_empty()),
            rule::STATUS_LINE_COMMAND,
            "statusLine.command",
        ),
        (
            optional("padding", Value::is_number),
            rule::STATUS_LINE_PADDING,
            "statusLine.padding",
        ),
        (
            optional("refreshInterval", |v| v.as_f64().is_some_and(|n| n >= 1.0)),
            rule::STATUS_LINE_REFRESH_INTERVAL,
            "statusLine.refreshInterval",
        ),
        (
            optional("hideVimModeIndicator", Value::is_boolean),
            rule::STATUS_LINE_HIDE_VIM_MODE_INDICATOR,
            "statusLine.hideVimModeIndicator",
        ),
    ];
    out.extend(
        broken
            .into_iter()
            .filter(|(holds, ..)| !holds)
            .map(|(_, rule, key)| entry(rule, key)),
    );
}

fn output_style(current: &Map<String, Value>, next: &Map<String, Value>, out: &mut Vec<Value>) {
    if changed(current, next, "outputStyle")
        .is_some_and(|v| !v.as_str().is_some_and(|s| !s.is_empty()))
    {
        out.push(entry(rule::OUTPUT_STYLE_SHAPE, "outputStyle"));
    }
}

/// Each `env` entry the write adds or changes.
fn env(current: &Map<String, Value>, next: &Map<String, Value>, out: &mut Vec<Value>) {
    let Some(map) = next.get("env").and_then(Value::as_object) else {
        return;
    };
    for (name, value) in map {
        if current.get("env").and_then(|e| e.get(name)) == Some(value) {
            continue;
        }
        let key = format!("env.{}", shown_value(name));
        if value
            .as_str()
            .is_some_and(|s| s.contains('$') || s.starts_with('~'))
        {
            out.push(entry(rule::ENV_VALUE_NOT_EXPANDED, &key));
        }
        if STORAGE_NAMES.contains(&name.as_str()) || name.starts_with(STORAGE_PREFIX) {
            out.push(entry(rule::ENV_STORAGE_LOCATION, &key));
        }
        if turns_on_switch(name, value) {
            out.push(entry(rule::ENV_FEATURE_SWITCH, &key));
        }
    }
}

/// Each permission rule the write adds to a list; the key names its place in the new list.
fn permissions(current: &Map<String, Value>, next: &Map<String, Value>, out: &mut Vec<Value>) {
    for list in ["allow", "deny", "ask"] {
        let of = |m: &Map<String, Value>| {
            m.get("permissions")
                .and_then(|p| p.get(list))
                .and_then(Value::as_array)
                .cloned()
                .unwrap_or_default()
        };
        let held = of(current);
        for (i, value) in of(next).iter().enumerate() {
            if held.contains(value) {
                continue;
            }
            let Some((tool, argument)) = value
                .as_str()
                .and_then(|r| r.strip_suffix(')'))
                .and_then(|r| r.split_once('('))
            else {
                continue;
            };
            let key = format!("permissions.{list}[{i}]");
            if PATH_BLIND_TOOLS.contains(&tool) && argument != "*" {
                out.push(entry(rule::PERMISSION_PATH_NOT_MATCHED, &key));
            }
            if list == "allow" && tool == "Bash" && wildcard_before_command(argument) {
                out.push(entry(rule::BASH_WILDCARD_BEFORE_COMMAND, &key));
            }
        }
    }
}

/// Whether the first `*` of a Bash rule comes before a later word that determines the command
/// (https://code.claude.com/docs/en/permissions, "Wildcard patterns": `git * main`,
/// `git -C * status *`). An approximation from the words alone: the program and the subcommand
/// determine a command, so fewer than two plain words stand before the first `*` and a plain word
/// follows it. A flag, or a word with a `*`, is not a plain word.
fn wildcard_before_command(argument: &str) -> bool {
    let plain = |w: &&str| !w.starts_with('-') && !w.contains('*');
    let words: Vec<&str> = argument.split_whitespace().collect();
    let Some(star) = words.iter().position(|w| w.contains('*')) else {
        return false;
    };
    words[..star].iter().filter(|w| plain(w)).count() < 2 && words[star + 1..].iter().any(plain)
}

/// What writing the file again does not keep of the bytes it holds now.
fn lost_content(file: &str, before: &[u8], out: &mut Vec<Value>) {
    if repeats_a_key(before) {
        out.push(json!({"rule": rule::REWRITE_REPEATED_KEY, "file": file}));
    }
    // A key name is not a value; one shaped like a credential is still hidden.
    out.extend(numbers_that_change(before).iter().map(
        |key| json!({"rule": rule::REWRITE_NUMBER_SPELLING, "key": shown_value(key), "file": file}),
    ));
}

#[cfg(test)]
mod tests {
    use super::*;

    // SPEC-20261122-settings-write-check REQ-2
    #[test]
    fn a_wildcard_is_reported_only_before_a_word_that_determines_the_command() {
        for reported in ["git * main", "git -C * status *", "* install"] {
            assert!(wildcard_before_command(reported), "{reported:?}");
        }
        for silent in [
            "git status *",
            "ls *",
            "git log * main",
            "ls:*",
            "git *",
            "npm run *",
            "* --version",
            "* --help *",
            "*",
            "git status",
            "",
        ] {
            assert!(!wildcard_before_command(silent), "{silent:?}");
        }
    }

    // SPEC-20261122-settings-write-check REQ-9
    #[cfg(unix)]
    #[test]
    fn the_log_is_private_and_nothing_is_made_through_a_linked_receipts_folder() {
        use std::os::unix::fs::PermissionsExt;
        let bits = |p: &Path| fs::symlink_metadata(p).unwrap().permissions().mode() & 0o777;
        let project = tempfile::tempdir().unwrap();
        let state = project.path().join(crate::PROJECT_STATE_DIR);
        fs::create_dir(&state).unwrap();
        record(project.path(), "settings.json", &[]).unwrap();
        let receipts = state.join("receipts");
        assert_eq!(bits(&receipts) & 0o077, 0);
        assert_eq!(bits(&receipts.join(LOG)) & 0o077, 0);

        let linked = tempfile::tempdir().unwrap();
        let outside = tempfile::tempdir().unwrap();
        let state = linked.path().join(crate::PROJECT_STATE_DIR);
        fs::create_dir(&state).unwrap();
        std::os::unix::fs::symlink(outside.path(), state.join("receipts")).unwrap();
        assert!(record(linked.path(), "settings.json", &[]).is_err());
        assert_eq!(fs::read_dir(outside.path()).unwrap().count(), 0);
    }
}
