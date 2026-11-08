//! Codex adapter (02-C #1993): projection plan and read-only doctor. Nothing here writes.
//!
//! Sources (read 2026-11-08):
//! - https://learn.chatgpt.com/docs/build-skills — skill locations (`.agents/skills` from the
//!   working directory up to the repository root, `$HOME/.agents/skills`, `/etc/codex/skills`),
//!   required `name`/`description`, same-name skills are not merged, `agents/openai.yaml`
//!   fields, `allow_implicit_invocation` default `true`.
//! - https://learn.chatgpt.com/docs/agent-configuration/agents-md — global `AGENTS.override.md`
//!   else `AGENTS.md` (first non-empty), project root down to the working directory with at most
//!   one file per directory (override, `AGENTS.md`, fallback names), blank-line joins, stop at
//!   `project_doc_max_bytes` (32 KiB by default).
//! - https://learn.chatgpt.com/docs/config-file/config-reference — `projects.<path>.trust_level`;
//!   untrusted projects skip the project `.codex/` layer; `tui.status_line` item identifiers are
//!   not enumerated.
//!
//! RC1 interim decisions (leader, 2026-11-08): no TOML parser here (the TOML surface belongs to
//! 02-B #1992), so trust, inline hooks and the status line are `not_evaluated` and never
//! reported as trusted; no hooks are generated; the projection target is always explicit.
use crate::{digest, package_digest};
use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
};

type Result<T> = std::result::Result<T, String>;

/// Default of `project_doc_max_bytes` stated on the AGENTS.md page (32 KiB).
pub const DEFAULT_MAX_BYTES: u64 = 32 * 1024;
const FILE_LIMIT: u64 = 1024 * 1024;
const TOML_HOLD: &str =
    "not read: no TOML parser in RC1 (02-B #1992 owns the TOML surface; leader decision 2026-11-08)";
const STATUS_LINE_HOLD: &str = "tui.status_line item identifiers are not listed in the official config reference (read 2026-11-08); nothing is written";

/// Options for `agents/openai.yaml`; `None` keeps the documented default by omitting the key.
pub struct OpenAiYaml {
    pub display_name: String,
    pub short_description: String,
    pub default_prompt: String,
    pub allow_implicit_invocation: Option<bool>,
}

fn yaml_scalar(field: &str, value: &str) -> Result<String> {
    if value.is_empty() || value.chars().count() > 256 || value.chars().any(char::is_control) {
        return Err(format!(
            "invalid {field}: 1-256 characters without control characters"
        ));
    }
    Ok(format!(
        "\"{}\"",
        value.replace('\\', "\\\\").replace('"', "\\\"")
    ))
}

/// Renders `agents/openai.yaml` with only the explicitly chosen fields (REQ-02C-02).
pub fn openai_yaml(o: &OpenAiYaml) -> Result<String> {
    let mut text = format!(
        "interface:\n  display_name: {}\n  short_description: {}\n  default_prompt: {}\n",
        yaml_scalar("display_name", &o.display_name)?,
        yaml_scalar("short_description", &o.short_description)?,
        yaml_scalar("default_prompt", &o.default_prompt)?,
    );
    if let Some(allow) = o.allow_implicit_invocation {
        text.push_str(&format!("policy:\n  allow_implicit_invocation: {allow}\n"));
    }
    Ok(text)
}

fn package_files(root: &Path, dir: &Path, out: &mut Vec<(String, Vec<u8>)>) -> Result<()> {
    let mut entries: Vec<PathBuf> = fs::read_dir(dir)
        .map_err(|_| "package directory unavailable")?
        .map(|e| {
            e.map(|e| e.path())
                .map_err(|_| "package entry unavailable".to_string())
        })
        .collect::<Result<_>>()?;
    entries.sort();
    for path in entries {
        if path.is_dir() {
            package_files(root, &path, out)?;
        } else {
            let rel = path
                .strip_prefix(root)
                .map_err(|_| "package path invalid")?
                .to_str()
                .ok_or("nonportable package path")?
                .replace('\\', "/");
            out.push((rel, fs::read(&path).map_err(|_| "package read failed")?));
        }
    }
    Ok(())
}

/// Plans a Codex projection of the common package into `target` without writing anything.
/// `status_line` and `disable_bundled_skills` are config requests: they are reported as HOLD.
pub fn projection_plan(
    package: &Path,
    target: &Path,
    yaml: &OpenAiYaml,
    status_line: &[String],
    disable_bundled_skills: bool,
) -> Result<Value> {
    let package_digest = package_digest(package)?;
    if !target.is_absolute() {
        return Err("invalid target: must be an absolute path".into());
    }
    let package_abs = package.canonicalize().map_err(|_| "package unavailable")?;
    if target.starts_with(&package_abs) {
        return Err("invalid target: must be outside the package".into());
    }
    if target.file_name().and_then(|n| n.to_str()) != Some("moai") {
        return Err("invalid target: the skill directory must be named moai".into());
    }
    let mut files = Vec::new();
    package_files(&package_abs, &package_abs, &mut files)?;
    if files.iter().any(|(rel, _)| rel.starts_with("agents/")) {
        return Err(
            "invalid package: agents/ belongs to the Codex projection, not the package".into(),
        );
    }
    let mut plan: Vec<Value> = files
        .iter()
        .map(|(rel, bytes)| {
            json!({"path": rel, "action": "copy", "source": "package", "sha256": digest(bytes)})
        })
        .collect();
    let rendered = openai_yaml(yaml)?;
    plan.push(
        json!({"path": "agents/openai.yaml", "action": "create", "source": "adapter",
        "sha256": digest(rendered.as_bytes()), "content": rendered}),
    );
    let unverified: Vec<&String> = status_line.iter().collect();
    let config = if status_line.is_empty() && !disable_bundled_skills {
        json!({"status": "none_requested"})
    } else {
        json!({"status": "hold", "reason": TOML_HOLD,
            "requested": {"tui.status_line": status_line, "skills.bundled.enabled": if disable_bundled_skills { json!(false) } else { Value::Null }},
            "unverified_items": unverified, "unverified_reason": STATUS_LINE_HOLD,
            "writes": []})
    };
    Ok(json!({
        "host": "codex",
        "target": target,
        "package_digest": package_digest,
        "files": plan,
        "skill_md_identical": true,
        "config_changes": config,
        "hooks": {"generated": [], "reason": "RC1 interim decision: no Codex hooks are generated"},
        "invocation": {"desktop": "$moai in the composer", "cli": "/skills or $moai", "ide": "/skills or $moai"},
        "notes": [
            "plan only: nothing is written; installing into a user home is EP-05",
            "target candidates ~/.agents/skills/moai and <repo>/.agents/skills/moai are a HOLD; the caller chooses",
            "a second skill named moai in another location is not merged; both can appear in selectors"
        ],
        "files_created": [],
        "host_loaded": false,
        "execution_verified": false
    }))
}

/// Reads a regular file of bounded size; `Ok(None)` when it is absent.
fn read_text(path: &Path) -> Result<Option<String>> {
    match fs::symlink_metadata(path) {
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(None),
        Err(_) => Err("target metadata unavailable".into()),
        Ok(m) if !m.is_file() || m.len() > FILE_LIMIT => {
            Err("target is not a bounded regular file".into())
        }
        Ok(_) => fs::read_to_string(path)
            .map(Some)
            .map_err(|_| "target read failed".into()),
    }
}

/// The first of `names` in `dir` that exists and is not empty ("Codex skips empty files").
fn first_doc(dir: &Path, names: &[String]) -> Result<Option<(PathBuf, String)>> {
    for name in names {
        let path = dir.join(name);
        if let Some(text) = read_text(&path)? {
            if text.trim().is_empty() {
                continue;
            }
            return Ok(Some((path, text)));
        }
    }
    Ok(None)
}

/// `roots` are skill roots (directories that hold `<name>/SKILL.md`).
fn moai_skills(roots: &[(PathBuf, &str)]) -> Vec<Value> {
    roots
        .iter()
        .map(|(root, scope)| (root.join("moai/SKILL.md"), scope))
        .filter(|(path, _)| path.is_file())
        .map(|(path, scope)| json!({"path": path, "scope": scope}))
        .collect()
}

/// Read-only Codex doctor for `project` (the project root) and `cwd` inside it.
pub fn doctor(
    project: &Path,
    cwd: &Path,
    codex_home: &Path,
    home: &Path,
    fallback: &[String],
    max_bytes: Option<u64>,
) -> Result<Value> {
    let root = project
        .canonicalize()
        .map_err(|_| "project directory unavailable")?;
    let cwd = cwd.canonicalize().map_err(|_| "path unavailable")?;
    if !root.is_dir() {
        return Err("project must be a directory".into());
    }
    if !cwd.starts_with(&root) {
        return Err("invalid cwd: must be inside the project".into());
    }
    for name in fallback {
        if name.is_empty() || name.contains('/') || name.contains('\\') || name == ".." {
            return Err("invalid fallback name: a plain file name is required".into());
        }
    }
    let (budget, budget_source) = match max_bytes {
        Some(0) => return Err("invalid max-bytes: must be positive".into()),
        Some(n) => (n, "argument"),
        None => (
            DEFAULT_MAX_BYTES,
            "default 32 KiB stated on the AGENTS.md page",
        ),
    };
    // Global level: AGENTS.override.md if present, else AGENTS.md; the first non-empty file.
    let global_names = vec!["AGENTS.override.md".to_string(), "AGENTS.md".to_string()];
    let g_hit = first_doc(codex_home, &global_names)?;
    let global = json!({"dir": codex_home, "considered": global_names,
        "included": g_hit.as_ref().map(|(p, _)| p), "bytes": g_hit.as_ref().map(|(_, t)| t.len())});
    // Project level: root down to cwd, at most one file per directory.
    let mut dirs = vec![root.clone()];
    for part in cwd
        .strip_prefix(&root)
        .map_err(|_| "invalid cwd: must be inside the project")?
        .components()
    {
        let next = dirs.last().unwrap().join(part);
        dirs.push(next);
    }
    let mut names = vec!["AGENTS.override.md".to_string(), "AGENTS.md".to_string()];
    names.extend(fallback.iter().cloned());
    let mut chain = Vec::new();
    let mut included = Vec::new();
    let mut total: u64 = 0;
    let mut over_at: Option<PathBuf> = None;
    for dir in &dirs {
        let considered = &names;
        // "Codex skips empty files" (AGENTS.md page). Moving on to the next name in the same
        // directory after an empty one is inferred; the page does not say so for projects.
        let hit = first_doc(dir, &names)?;
        let entry = match &hit {
            None => json!({"dir": dir, "considered": considered, "included": null}),
            Some((path, text)) => {
                let bytes = text.len() as u64;
                let joined = if included.is_empty() {
                    bytes
                } else {
                    bytes + 2
                };
                if over_at.is_none() && total + joined > budget {
                    over_at = Some(path.clone());
                }
                total += joined;
                included.push(json!({"path": path, "bytes": bytes}));
                json!({"dir": dir, "considered": considered, "included": path})
            }
        };
        chain.push(entry);
    }
    // Skill locations from the build-skills page, for the duplicate-name warning (REQ-02C-04).
    let mut skill_dirs: Vec<(PathBuf, &str)> = dirs
        .iter()
        .rev()
        .map(|d| (d.join(".agents/skills"), "repository"))
        .collect();
    skill_dirs.push((home.join(".agents/skills"), "user"));
    skill_dirs.push((PathBuf::from("/etc/codex/skills"), "admin"));
    let found = moai_skills(&skill_dirs);
    let duplicate = found.len() > 1;
    let layer = root.join(".codex");
    Ok(json!({
        "host": "codex",
        "project_root": root,
        "cwd": cwd,
        "agents": {
            "global": global,
            "project_chain": chain,
            "included": included,
            "project_bytes": total,
            "max_bytes": budget,
            "max_bytes_source": budget_source,
            "fallback_names": fallback,
            "fallback_source": if fallback.is_empty() { "none given; project_doc_fallback_filenames is not read (TOML HOLD)" } else { "argument" },
            "over_budget": over_at.is_some(),
            "over_budget_at": over_at,
            "budget_note": "the project chain is counted with blank-line joins; whether Codex counts the global file in the same budget and how it cuts the crossing file are not stated on the page",
            "agents_override_created": false
        },
        "trust": {"status": "not_evaluated", "reason": TOML_HOLD,
            "consequence": "if the project is untrusted, Codex skips .codex/config.toml, hooks and rules (blocked_by_trust); this doctor cannot tell and never reports trusted"},
        "codex_layer": {
            "config_toml": layer.join("config.toml").is_file(),
            "hooks_json": layer.join("hooks.json").is_file(),
            "inline_hooks": "not_evaluated",
            "status_line": "not_evaluated",
            "reason": TOML_HOLD
        },
        "skills": {"moai": found, "duplicate_warning": if duplicate {
            json!("more than one skill named moai was found; Codex does not merge them and both can appear in selectors")
        } else { Value::Null }},
        "writes": [],
        "host_loaded": false,
        "execution_verified": false
    }))
}
