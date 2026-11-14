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

/// The Go hook script a handler runs, when it runs one: the first `command` or `args` value
/// that names a path under [`GO_HOOK_DIR`] (with or without `${CLAUDE_PROJECT_DIR}/` or `$CLAUDE_PROJECT_DIR/`
/// in front). `None` for every other handler.
pub fn go_script(handler: &Value) -> Option<String> {
    let mut texts: Vec<&str> = Vec::new();
    if let Some(c) = handler["command"].as_str() {
        texts.push(c);
    }
    if let Some(args) = handler["args"].as_array() {
        texts.extend(args.iter().filter_map(Value::as_str));
    }
    let marker = format!("{GO_HOOK_DIR}/");
    texts.into_iter().find_map(|t| {
        let at = t.find(&marker)?;
        let rest = &t[at + marker.len()..];
        let name: String = rest
            .chars()
            .take_while(|c| c.is_ascii_alphanumeric() || matches!(c, '-' | '_' | '.'))
            .collect();
        (!name.is_empty()).then_some(name)
    })
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
        for other in [
            json!({"type": "command", "command": "moai-cli hook-receipt claude"}),
            json!({"type": "command", "command": "orca hook"}),
            json!({"type": "command", "command": "bash", "args": [".claude/hooks/mine/x.sh"]}),
            json!({"type": "command", "command": "echo .claude/hooks/moai/"}),
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
