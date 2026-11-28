//! The managed locks and the cross session inbound value of `moai doctor`
//! (SPEC-20261127-doctor-moai-layer REQ-9..REQ-11). Both read settings files only: the managed
//! `managed-settings.json` and `managed-settings.d/*.json` of the platform folder, the user
//! settings of the Claude config folder and the project's `.claude/settings.json` and
//! `.claude/settings.local.json`. The scope rules are the documented ones
//! (https://code.claude.com/docs/en/managed-settings.md, settings-reference.md,
//! cross-session-messaging.md, read 2026-11-27). A source that is not a file (the macOS managed
//! preferences domain, the Windows registry, server-managed settings, `policyHelper`, a session's
//! `--settings`) is named as not read and never guessed.

use super::{Check, Gathered, Scope, Settings, State};
use serde_json::{json, Map, Value};

/// The label of the managed file that carries the base values.
pub(super) const MANAGED_BASE: &str = "managed-settings.json";

/// What `moai doctor` never reads as a file, named in the evidence of the checks that depend on
/// it (REQ-10, REQ-11).
fn not_read() -> Vec<&'static str> {
    let mut names = vec![
        "policyHelper (never run)",
        "server-managed settings",
        "--settings of a session",
    ];
    if cfg!(target_os = "macos") {
        names.push("macOS managed preferences domain com.anthropic.claudecode");
    }
    if cfg!(windows) {
        names.push("Windows registry HKLM\\SOFTWARE\\Policies\\ClaudeCode");
    }
    names
}

/// The value at `path` of a settings file; `null` counts as absent.
fn lookup<'a>(settings: &'a Settings, path: &[&str]) -> Option<&'a Value> {
    let Settings::Json(root) = settings else {
        return None;
    };
    let mut at = root;
    for step in path {
        at = at.get(*step)?;
    }
    (!at.is_null()).then_some(at)
}

/// The labels of the files that could not be read as JSON objects, among the managed files and
/// (when `all_scopes`) the user, project and local files.
fn unreadable(g: &Gathered, all_scopes: bool) -> Vec<String> {
    let mut labels: Vec<String> = g
        .managed
        .iter()
        .filter(|(_, s)| matches!(s, Settings::Unreadable))
        .map(|(label, _)| label.clone())
        .collect();
    if all_scopes {
        labels.extend(
            g.settings
                .iter()
                .filter(|(_, s)| matches!(s, Settings::Unreadable))
                .map(|(scope, _)| scope.label().to_string()),
        );
    }
    labels
}

/// The last managed file that carries `path` (the files merge in alphabetical order, a later one
/// wins), with its label.
fn managed_value<'a>(g: &'a Gathered, path: &[&str]) -> Option<(&'a str, &'a Value)> {
    g.managed
        .iter()
        .rev()
        .find_map(|(label, s)| lookup(s, path).map(|v| (label.as_str(), v)))
}

fn scope_value<'a>(g: &'a Gathered, scope: Scope, path: &[&str]) -> Option<&'a Value> {
    g.settings
        .iter()
        .find(|(s, _)| s.label() == scope.label())
        .and_then(|(_, settings)| lookup(settings, path))
}

/// A value that would lock something if it were in effect: anything but `false`, and an empty
/// array.
fn carries(v: &Value) -> bool {
    match v {
        Value::Bool(b) => *b,
        Value::Array(a) => !a.is_empty(),
        Value::Null => false,
        _ => true,
    }
}

const SURFACES: [&str; 4] = ["skills", "agents", "hooks", "mcp"];

/// The surfaces `strictPluginOnlyCustomization` locks: `true` all four, an array the ones it
/// names, and any other form all four (Claude Code 2.1.282 and later).
fn strict_surfaces(v: &Value) -> Vec<&'static str> {
    match v {
        Value::Bool(true) => SURFACES.to_vec(),
        Value::Array(items) => {
            let named: Option<Vec<&str>> = items.iter().map(Value::as_str).collect();
            match named {
                Some(names) if names.iter().all(|n| SURFACES.contains(n)) => SURFACES
                    .iter()
                    .copied()
                    .filter(|s| names.contains(s))
                    .collect(),
                _ => SURFACES.to_vec(),
            }
        }
        Value::Bool(false) | Value::Null => Vec::new(),
        _ => SURFACES.to_vec(),
    }
}

const USER_PROJECT_LOCAL: [Scope; 3] = [Scope::User, Scope::Project, Scope::Local];

/// The scopes (labels) among user, project and local that carry `path` with a locking value.
fn ignored_in(g: &Gathered, path: &[&str]) -> Vec<&'static str> {
    USER_PROJECT_LOCAL
        .iter()
        .filter(|scope| scope_value(g, **scope, path).is_some_and(carries))
        .map(|scope| scope.label())
        .collect()
}

/// The answer for one key: `state`, the `source` file, the `surfaces` it locks and the files
/// that carry it where it has no effect, could not be read or were discarded.
struct Key {
    state: &'static str,
    source: Option<String>,
    surfaces: Vec<&'static str>,
    ignored_in: Vec<&'static str>,
    unreadable: Vec<String>,
    discarded: Vec<String>,
}

impl Key {
    fn to_json(&self) -> Value {
        json!({
            "state": self.state,
            "source": self.source,
            "surfaces": self.surfaces,
            "ignored_in": self.ignored_in,
            "unreadable": self.unreadable,
            "discarded": self.discarded,
        })
    }
}

/// A key that only a managed source can set: the managed value counts, a value in the user,
/// project or local files has no effect.
fn managed_only(
    g: &Gathered,
    path: &[&str],
    surfaces: impl Fn(&Value) -> Vec<&'static str>,
) -> Key {
    let unreadable = unreadable(g, false);
    let ignored = ignored_in(g, path);
    if !unreadable.is_empty() {
        return Key {
            state: "unknown",
            source: None,
            surfaces: Vec::new(),
            ignored_in: ignored,
            unreadable,
            discarded: Vec::new(),
        };
    }
    if let Some((label, value)) = managed_value(g, path) {
        let locked = surfaces(value);
        if carries(value) && !locked.is_empty() {
            return Key {
                state: "effective",
                source: Some(label.to_string()),
                surfaces: locked,
                ignored_in: ignored,
                unreadable,
                discarded: Vec::new(),
            };
        }
    }
    Key {
        state: if ignored.is_empty() {
            "unset"
        } else {
            "ignored"
        },
        source: None,
        surfaces: Vec::new(),
        ignored_in: ignored,
        unreadable,
        discarded: Vec::new(),
    }
}

/// `disableAllHooks`: any file may set it; the managed file wins, then local, project and user.
/// A managed value that is not a boolean is discarded by Claude Code with a warning.
fn disable_all_hooks(g: &Gathered) -> Key {
    let path = ["disableAllHooks"];
    let unreadable = unreadable(g, true);
    if !unreadable.is_empty() {
        return Key {
            state: "unknown",
            source: None,
            surfaces: Vec::new(),
            ignored_in: Vec::new(),
            unreadable,
            discarded: Vec::new(),
        };
    }
    let mut discarded = Vec::new();
    let mut winner: Option<(String, bool, bool)> = None; // label, value, managed
    match managed_value(g, &path) {
        Some((label, Value::Bool(b))) => winner = Some((label.to_string(), *b, true)),
        Some((label, _)) => discarded.push(label.to_string()),
        None => {}
    }
    if winner.is_none() {
        for scope in [Scope::Local, Scope::Project, Scope::User] {
            if let Some(Value::Bool(b)) = scope_value(g, scope, &path) {
                winner = Some((scope.label().to_string(), *b, false));
                break;
            }
        }
    }
    match winner {
        Some((label, true, managed)) => {
            let mut surfaces = vec!["hooks", "statusline", "file_suggestion_command"];
            if managed {
                surfaces.push("managed_hooks");
            }
            Key {
                state: "effective",
                source: Some(label),
                surfaces,
                ignored_in: Vec::new(),
                unreadable,
                discarded,
            }
        }
        _ => Key {
            state: "unset",
            source: None,
            surfaces: Vec::new(),
            ignored_in: Vec::new(),
            unreadable,
            discarded,
        },
    }
}

/// The four managed locks (REQ-9, REQ-10).
pub(super) fn managed_locks_check(g: &Gathered) -> Check {
    const MODS_PATH: [&str; 4] = [
        "pluginConfigs",
        "cc-plugin-sec-default@builtin",
        "options",
        "allowManagedModsOnly",
    ];
    let keys = [
        (
            "allowManagedHooksOnly",
            managed_only(g, &["allowManagedHooksOnly"], |_| vec!["hooks", "mods"]),
        ),
        (
            "strictPluginOnlyCustomization",
            managed_only(g, &["strictPluginOnlyCustomization"], strict_surfaces),
        ),
        (
            "allowManagedModsOnly",
            managed_only(g, &MODS_PATH, |v| {
                if *v == Value::Bool(true) {
                    vec!["mods"]
                } else {
                    Vec::new()
                }
            }),
        ),
        ("disableAllHooks", disable_all_hooks(g)),
    ];
    let effective: Vec<&str> = keys
        .iter()
        .filter(|(_, k)| k.state == "effective")
        .map(|(name, _)| *name)
        .collect();
    let unknown: Vec<&str> = keys
        .iter()
        .filter(|(_, k)| k.state == "unknown")
        .map(|(name, _)| *name)
        .collect();
    let (state, summary) = if !effective.is_empty() {
        (
            State::Warn,
            format!(
                "a lock is in effect: {}; the surfaces it names are off for what moai installs there",
                effective.join(", ")
            ),
        )
    } else if !unknown.is_empty() {
        (
            State::Unknown,
            format!(
                "a settings file could not be read as JSON, so {} cannot be judged",
                unknown.join(", ")
            ),
        )
    } else {
        (
            State::Ok,
            "no managed lock is in effect in the files that were read".to_string(),
        )
    };
    let mut map = Map::new();
    for (name, key) in &keys {
        map.insert((*name).to_string(), key.to_json());
    }
    Check {
        id: "managed_locks",
        state,
        summary,
        evidence: json!({ "keys": map, "not_read": not_read() }),
        next: "a managed lock is set by the organization: ask whoever deploys managed-settings.json before changing it".into(),
    }
}

/// How strict a `crossSessionInbound` value is: `accept` < `hold` < `refuse`.
fn strictness(value: &str) -> u8 {
    match value {
        "refuse" => 2,
        "hold" => 1,
        _ => 0,
    }
}

/// The documented reading of one value: the three strings, and any other value is `hold` in the
/// user and project files and `refuse` in the managed file.
fn inbound_value(v: &Value, managed: bool) -> &'static str {
    match v.as_str() {
        Some("accept") => "accept",
        Some("hold") => "hold",
        Some("refuse") => "refuse",
        _ if managed => "refuse",
        _ => "hold",
    }
}

/// `crossSessionInbound` by the documented precedence (REQ-11): the managed file, then the user
/// settings; a project or local value only when it is stricter than that, or when no trusted
/// source sets one.
pub(super) fn cross_session_inbound_check(g: &Gathered) -> Check {
    const PATH: [&str; 1] = ["crossSessionInbound"];
    let unreadable = unreadable(g, true);
    let mut chosen: Option<(String, &'static str)> =
        managed_value(g, &PATH).map(|(label, v)| (label.to_string(), inbound_value(v, true)));
    if chosen.is_none() {
        chosen = scope_value(g, Scope::User, &PATH)
            .map(|v| (Scope::User.label().to_string(), inbound_value(v, false)));
    }
    for scope in [Scope::Project, Scope::Local] {
        let Some(v) = scope_value(g, scope, &PATH) else {
            continue;
        };
        let value = inbound_value(v, false);
        let stricter = chosen
            .as_ref()
            .is_none_or(|(_, current)| strictness(value) > strictness(current));
        if value != "accept" && stricter {
            chosen = Some((scope.label().to_string(), value));
        }
    }
    let (value, source) = match &chosen {
        Some((label, value)) => (*value, Some(label.clone())),
        None => ("unset", None),
    };
    let (state, summary) = if !unreadable.is_empty() {
        (
            State::Unknown,
            format!(
                "a settings file could not be read as JSON ({}), so the inbound value is not final",
                unreadable.join(", ")
            ),
        )
    } else if value == "accept" {
        (
            State::Ok,
            format!(
                "cross session messages are accepted ({})",
                source.as_deref().unwrap_or("")
            ),
        )
    } else {
        (
            State::Warn,
            format!(
                "crossSessionInbound is {value}: a lane started with -l or -a runs with bypassed permissions and holds messages then"
            ),
        )
    };
    Check {
        id: "cross_session_inbound",
        state,
        summary,
        evidence: json!({
            "value": value,
            "source": source,
            "unreadable": unreadable,
            "not_read": not_read(),
        }),
        next: "set crossSessionInbound to \"accept\" in the user settings (or the managed file) for lanes that must receive messages".into(),
    }
}
