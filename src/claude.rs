//! Claude Code adapter checks (Issue #1992, 02-B). Read-only: nothing here writes.
//!
//! `doctor` predicts whether Claude Code reads a project's `AGENTS.md`, following the rules
//! of https://code.claude.com/docs/en/memory (read 2026-11-08): direct reading needs v2.1.277
//! or later; a `CLAUDE.md`, `.claude/CLAUDE.md` or `CLAUDE.local.md` in the working directory
//! or above suppresses it under the default `claude-md-or-agents-md`; the
//! `pluginConfigs.<id>.options.instructionFiles` entry is honoured only in user, `--settings`
//! and managed settings, never in project settings; the plugin id is
//! `cc-plugin-agents-md@builtin` from v2.1.285 and `agents-md@builtin` before.
//! The result is a prediction from files and settings; whether a session actually loaded the
//! file is `host_loaded: not_run` until a session receipt exists.
use serde_json::{json, Value};
use std::{
    fs,
    io::Read,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};

type Result<T> = std::result::Result<T, String>;

pub const MIN_AGENTS_VERSION: (u32, u32, u32) = (2, 1, 277);
const PLUGIN_ID_SINCE: (u32, u32, u32) = (2, 1, 285);
/// Files that suppress direct `AGENTS.md` reading under the default setting.
const SUPPRESSING: [&str; 3] = ["CLAUDE.md", ".claude/CLAUDE.md", "CLAUDE.local.md"];
const AGENTS_FILES: [&str; 2] = ["AGENTS.md", ".claude/AGENTS.md"];
/// Upper bound on bytes read from one CLAUDE file or settings file.
const READ_LIMIT: u64 = 4 * 1024 * 1024;
const VERSION_OUTPUT_LIMIT: usize = 4096;

fn parse_version(text: &str) -> Option<(u32, u32, u32)> {
    let token = text.split_whitespace().next()?;
    let mut parts = token.split('.').map(|p| p.parse::<u32>().ok());
    let v = (parts.next()??, parts.next()??, parts.next()??);
    parts.next().is_none().then_some(v)
}

/// `claude --version` with a fixed argv; `None` when claude is absent or prints no version.
fn claude_version() -> Option<(u32, u32, u32)> {
    let out = Command::new("claude")
        .arg("--version")
        .stdin(Stdio::null())
        .stderr(Stdio::null())
        .output()
        .ok()?;
    if !out.status.success() || out.stdout.len() > VERSION_OUTPUT_LIMIT {
        return None;
    }
    parse_version(std::str::from_utf8(&out.stdout).ok()?)
}

fn read_bounded(path: &Path) -> Option<Vec<u8>> {
    let mut bytes = Vec::new();
    fs::File::open(path)
        .ok()?
        .take(READ_LIMIT + 1)
        .read_to_end(&mut bytes)
        .ok()?;
    (bytes.len() as u64 <= READ_LIMIT).then_some(bytes)
}

fn text(p: &Path) -> Result<String> {
    p.to_str()
        .map(str::to_string)
        .ok_or_else(|| "non UTF-8 path".into())
}

/// Every existing file named in `names` in `start` and each directory above it.
fn found_upward(start: &Path, names: &[&str]) -> Vec<PathBuf> {
    start
        .ancestors()
        .flat_map(|dir| names.iter().map(move |n| dir.join(n)))
        .filter(|p| p.is_file())
        .collect()
}

/// True when the file holds an `@...AGENTS.md` import outside code fences and code spans.
fn imports_agents(path: &Path) -> bool {
    let Some(bytes) = read_bounded(path) else {
        return false;
    };
    let mut fenced = false;
    String::from_utf8_lossy(&bytes).lines().any(|line| {
        if line.trim_start().starts_with("```") {
            fenced = !fenced;
            return false;
        }
        !fenced
            && line
                .split_whitespace()
                .any(|t| t.starts_with('@') && t.ends_with("AGENTS.md") && !t.contains('`'))
    })
}

/// The user-settings `instructionFiles` value: `Ok(None)` when absent, `Err` when the file
/// cannot be read as JSON.
fn user_instruction_files(home: &Path, plugin_id: &str) -> Result<Option<String>> {
    let path = home.join(".claude/settings.json");
    if fs::symlink_metadata(&path).is_err() {
        return Ok(None);
    }
    let bytes = read_bounded(&path).ok_or("user settings unreadable")?;
    let v: Value = serde_json::from_slice(&bytes).map_err(|_| "user settings are not JSON")?;
    match &v["pluginConfigs"][plugin_id]["options"]["instructionFiles"] {
        Value::Null => Ok(None),
        Value::String(s) => Ok(Some(s.clone())),
        _ => Err("instructionFiles is not a string".into()),
    }
}

fn paths(list: &[PathBuf]) -> Result<Value> {
    Ok(json!(list
        .iter()
        .map(|p| text(p))
        .collect::<Result<Vec<_>>>()?))
}

/// Predicts AGENTS.md loading for a session started in `project`. Never writes. `Err` only
/// for an unusable input path; everything else is a typed status.
pub fn doctor(project: &Path) -> Result<Value> {
    let meta = fs::symlink_metadata(project).map_err(|_| "path unavailable")?;
    if meta.file_type().is_symlink() {
        return Err("path is a symlink; refused".into());
    }
    if !meta.is_dir() {
        return Err("path must be a directory".into());
    }
    let dir = fs::canonicalize(project).map_err(|_| "path unavailable")?;
    let agents = found_upward(&dir, &AGENTS_FILES);
    let claude_files = found_upward(&dir, &SUPPRESSING);
    let importing: Vec<PathBuf> = claude_files
        .iter()
        .filter(|p| imports_agents(p))
        .cloned()
        .collect();
    let version = claude_version();
    let mut out = json!({
        "agents_md_status": "unknown",
        "reason": null,
        "load_path": null,
        "claude_version": version.map(|(a, b, c)| format!("{a}.{b}.{c}")),
        "minimum_version": "2.1.277",
        "agents_md_files": paths(&agents)?,
        "suppressing_files": paths(&claude_files)?,
        "importing_files": paths(&importing)?,
        "instruction_files": null,
        "instruction_files_source": null,
        "managed_settings": "not_checked",
        "suggestions": [],
        "basis": "predicted",
        "host_loaded": "not_run",
        "writes": false,
    });
    let set = |out: &mut Value, status: &str, reason: &str| {
        out["agents_md_status"] = json!(status);
        out["reason"] = json!(reason);
    };
    let Some(version) = version else {
        set(&mut out, "unknown", "claude --version gave no version");
        return Ok(out);
    };
    if version < MIN_AGENTS_VERSION {
        set(
            &mut out,
            "unsupported_version",
            "AGENTS.md reading needs 2.1.277 or later",
        );
        return Ok(out);
    }
    let plugin_id = if version >= PLUGIN_ID_SINCE {
        "cc-plugin-agents-md@builtin"
    } else {
        "agents-md@builtin"
    };
    let setting = match std::env::var_os("HOME").filter(|h| !h.is_empty()) {
        Some(home) => match user_instruction_files(Path::new(&home), plugin_id) {
            Ok(s) => s,
            Err(e) => {
                set(&mut out, "unknown", &e);
                return Ok(out);
            }
        },
        None => None,
    };
    out["instruction_files_source"] = json!(if setting.is_some() {
        "user_settings"
    } else {
        "default"
    });
    let setting = setting.unwrap_or_else(|| "claude-md-or-agents-md".into());
    out["instruction_files"] = json!(setting);
    if agents.is_empty() {
        set(
            &mut out,
            "unknown",
            "no AGENTS.md in the directory or above",
        );
        return Ok(out);
    }
    match setting.as_str() {
        "claude-md-and-agents-md" => {
            out["agents_md_status"] = json!("loaded");
            out["load_path"] = json!("direct");
        }
        "claude-md-or-agents-md" if claude_files.is_empty() => {
            out["agents_md_status"] = json!("loaded");
            out["load_path"] = json!("direct");
        }
        "claude-md-or-agents-md" if !importing.is_empty() => {
            out["agents_md_status"] = json!("loaded");
            out["load_path"] = json!("import");
        }
        "claude-md-or-agents-md" => {
            set(
                &mut out,
                "suppressed_by_claude_md",
                "a CLAUDE file in the directory or above takes precedence",
            );
            // Proposals only (REQ-02B-05): nothing is applied.
            out["suggestions"] = json!([
                "add an `@AGENTS.md` import to the CLAUDE.md you keep",
                "set Project instructions to claude-md-and-agents-md in /config or user settings",
            ]);
        }
        "claude-md" if !importing.is_empty() => {
            out["agents_md_status"] = json!("loaded");
            out["load_path"] = json!("import");
        }
        "claude-md" | "managed-only" => set(
            &mut out,
            "unknown",
            "instructionFiles excludes direct AGENTS.md reading",
        ),
        _ => set(&mut out, "unknown", "unrecognised instructionFiles value"),
    }
    Ok(out)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn version_parsing_is_strict() {
        assert_eq!(parse_version("2.1.319 (Claude Code)\n"), Some((2, 1, 319)));
        assert_eq!(parse_version("2.1.277"), Some((2, 1, 277)));
        for bad in [
            "",
            "Claude Code 2.1.319",
            "2.1",
            "2.1.x",
            "2.1.3.4",
            "v2.1.300",
        ] {
            assert_eq!(parse_version(bad), None, "{bad}");
        }
        assert!(MIN_AGENTS_VERSION <= (2, 1, 277) && (2, 1, 276) < MIN_AGENTS_VERSION);
    }
}
