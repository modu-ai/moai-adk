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
//!
//! `projection` renders the Claude copy of the common package's `SKILL.md` (REQ-02B-01/03):
//! the body stays byte-identical to the canonical one and the frontmatter only gains fields
//! from the Claude allow-list (https://code.claude.com/docs/en/skills, read 2026-11-08).
use serde_json::{json, Map, Value};
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

/// Top-level `settings.json` keys MoAI may own (REQ-02B-06/07). `outputStyle` is managed only
/// when the user asks for it; it is never set by default. The RC1 hook set is empty (P04 = A),
/// so `hooks` is not managed.
pub const MANAGED_SETTINGS: [&str; 2] = ["outputStyle", "statusLine"];

/// Three-way plan for the managed keys of a Claude `settings.json` (SETTINGS-PRESERVE).
/// `baseline` holds what MoAI owned after its last apply, `current` the file now, `desired`
/// what MoAI wants. A key changes only when the user left it as MoAI owned it; a key the user
/// changed, or one MoAI never owned that differs, is a conflict and nothing is planned. Every
/// other key, and the order of keys, stays as in `current`. Missing, `null` and other values
/// are all distinct.
pub fn plan_settings(
    baseline: &Map<String, Value>,
    current: &Map<String, Value>,
    desired: &Map<String, Value>,
) -> Result<(Option<Map<String, Value>>, Vec<String>)> {
    if baseline
        .keys()
        .chain(desired.keys())
        .any(|k| !MANAGED_SETTINGS.contains(&k.as_str()))
    {
        return Err("unmanaged claude settings key refused".into());
    }
    let mut keys: Vec<&String> = baseline.keys().chain(desired.keys()).collect();
    keys.sort();
    keys.dedup();
    let mut candidate = current.clone();
    let mut conflicts = Vec::new();
    for key in keys {
        let (old, now, new) = (baseline.get(key), current.get(key), desired.get(key));
        if now == new || new == old {
            continue;
        }
        if now != old {
            conflicts.push(key.clone());
            continue;
        }
        match new {
            // Insert keeps an existing key's position and appends a new one.
            Some(v) => {
                candidate.insert(key.clone(), v.clone());
            }
            // Filtering keeps the order of the remaining keys with or without preserve_order.
            None => candidate = candidate.into_iter().filter(|(k, _)| k != key).collect(),
        }
    }
    Ok(((conflicts.is_empty()).then_some(candidate), conflicts))
}

/// Claude-only frontmatter fields a projection may add, with the values each accepts.
/// Fields that grant tools or register hooks (`allowed-tools`, `hooks`, ...) are not offered:
/// package content never widens permissions, and the RC1 hook set is empty (P04 = A).
const CLAUDE_FIELDS: [(&str, &[&str]); 9] = [
    ("when_to_use", &[]),
    ("argument-hint", &[]),
    ("model", &[]),
    ("agent", &[]),
    ("disable-model-invocation", &["true", "false"]),
    ("user-invocable", &["true", "false"]),
    ("effort", &["low", "medium", "high", "xhigh", "max"]),
    ("context", &["fork"]),
    ("shell", &["bash", "powershell"]),
];

/// The YAML scalar for one added field: an enumerated word as is, free text as a JSON
/// (YAML double-quoted) string on a single line.
fn field_value(key: &str, value: &Value) -> Result<String> {
    let (_, allowed) = CLAUDE_FIELDS
        .iter()
        .find(|(k, _)| *k == key)
        .ok_or("unknown claude frontmatter field")?;
    let word = match value {
        Value::Bool(b) => b.to_string(),
        Value::String(s) => s.clone(),
        _ => return Err("invalid claude frontmatter value".into()),
    };
    if !allowed.is_empty() {
        return if allowed.contains(&word.as_str()) {
            Ok(word)
        } else {
            Err("invalid claude frontmatter value".into())
        };
    }
    if word.trim().is_empty() || word.len() > 1024 || word.chars().any(char::is_control) {
        return Err("invalid claude frontmatter value".into());
    }
    Ok(serde_json::to_string(&word).expect("string"))
}

/// Splits `---\n<frontmatter>\n---\n<body>`; the body is returned as raw bytes.
fn split_frontmatter(bytes: &[u8]) -> Result<(&str, &[u8])> {
    let rest = bytes
        .strip_prefix(b"---\n")
        .ok_or("SKILL.md frontmatter missing")?;
    let end = rest
        .windows(5)
        .position(|w| w == b"\n---\n")
        .ok_or("SKILL.md frontmatter missing")?;
    let fm = std::str::from_utf8(&rest[..end]).map_err(|_| "SKILL.md is not UTF-8")?;
    Ok((fm, &rest[end + 5..]))
}

/// Renders the Claude projection of `package` (the common skill directory). `skills_root` is
/// the target `.claude/skills` directory (personal or project); when it does not exist yet a
/// session that is already running needs `/reload-skills` to see the skill. Never writes.
pub fn projection(package: &Path, skills_root: &Path, added: &Map<String, Value>) -> Result<Value> {
    let package_digest = crate::package_digest(package)?;
    let skill = read_bounded(&package.join("SKILL.md")).ok_or("package SKILL.md missing")?;
    let (fm, body) = split_frontmatter(&skill)?;
    let present: Vec<&str> = fm
        .lines()
        .filter(|l| !l.starts_with([' ', '\t', '#']))
        .filter_map(|l| l.split_once(':').map(|(k, _)| k.trim()))
        .collect();
    if !present.contains(&"name") || !present.contains(&"description") {
        return Err("SKILL.md frontmatter missing".into());
    }
    let mut keys: Vec<&String> = added.keys().collect();
    keys.sort();
    let mut lines = String::new();
    let mut recorded = Map::new();
    for key in keys {
        if present.contains(&key.as_str()) {
            return Err("canonical frontmatter field override refused".into());
        }
        let scalar = field_value(key, &added[key])?;
        lines.push_str(&format!("{key}: {scalar}\n"));
        recorded.insert(key.clone(), added[key].clone());
    }
    let mut projected = format!("---\n{fm}\n{lines}---\n").into_bytes();
    projected.extend_from_slice(body);
    let reload = fs::symlink_metadata(skills_root).is_err();
    Ok(json!({
        "skill_md": String::from_utf8(projected.clone()).expect("UTF-8 parts"),
        "manifest": {
            "canonical_package_digest": package_digest,
            "canonical_skill_sha256": crate::digest(&skill),
            "projected_skill_sha256": crate::digest(&projected),
            "body_sha256": crate::digest(body),
            "frontmatter_added": recorded,
            "supporting_files": "byte-identical copies of the package",
        },
        "target": text(&skills_root.join("moai"))?,
        "reload_skills_required": reload,
        "files_created": false,
        "host_loaded": "not_run",
        "execution_verified": "not_run",
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    fn package() -> PathBuf {
        Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai")
    }
    fn added(pairs: &[(&str, Value)]) -> Map<String, Value> {
        pairs
            .iter()
            .map(|(k, v)| (k.to_string(), v.clone()))
            .collect()
    }

    #[test]
    fn projection_keeps_the_canonical_body_and_records_additions() {
        let skills = tempfile::TempDir::new().unwrap();
        let canonical = fs::read(package().join("SKILL.md")).unwrap();
        let plain = projection(&package(), skills.path(), &Map::new()).unwrap();
        assert_eq!(plain["skill_md"].as_str().unwrap().as_bytes(), canonical);
        assert_eq!(plain["reload_skills_required"], json!(false));
        assert_eq!(plain["files_created"], json!(false));
        let v = projection(
            &package(),
            &skills.path().join("absent"),
            &added(&[
                ("user-invocable", json!(true)),
                ("effort", json!("high")),
                ("when_to_use", json!("say \"moai\": plan")),
            ]),
        )
        .unwrap();
        let text = v["skill_md"].as_str().unwrap();
        let (fm, body) = split_frontmatter(text.as_bytes()).unwrap();
        assert_eq!(body, split_frontmatter(&canonical).unwrap().1);
        assert_eq!(v["manifest"]["body_sha256"], json!(crate::digest(body)));
        assert!(fm.ends_with(
            "effort: high\nuser-invocable: true\nwhen_to_use: \"say \\\"moai\\\": plan\""
        ));
        assert_eq!(v["manifest"]["frontmatter_added"]["effort"], json!("high"));
        assert_eq!(v["reload_skills_required"], json!(true));
    }

    fn obj(v: Value) -> Map<String, Value> {
        v.as_object().unwrap().clone()
    }

    // Ported from tests/claude_settings_plan_test.py (the provisional Python planner).
    #[test]
    fn claude_settings_adds_only_selected_keys_and_preserves_unknown_values() {
        let current = obj(json!({"model": "sonnet", "hooks": {"Start": [{"user": true}]}}));
        let (result, conflicts) = plan_settings(
            &Map::new(),
            &current,
            &obj(json!({"outputStyle": "MoAI-Easy"})),
        )
        .unwrap();
        assert!(conflicts.is_empty());
        let mut expected = current.clone();
        expected.insert("outputStyle".into(), json!("MoAI-Easy"));
        assert_eq!(result.unwrap(), expected);
    }

    #[test]
    fn claude_settings_user_owned_or_modified_setting_is_a_whole_plan_conflict() {
        let current =
            obj(json!({"model": "sonnet", "statusLine": {"type": "command", "command": "orca"}}));
        let desired = obj(
            json!({"outputStyle": "MoAI-Easy", "statusLine": {"type": "command", "command": "moai"}}),
        );
        let (result, conflicts) = plan_settings(&Map::new(), &current, &desired).unwrap();
        assert_eq!(result, None);
        assert_eq!(conflicts, ["statusLine"]);
    }

    #[test]
    fn claude_settings_three_way_update_idempotence_and_user_drift() {
        let baseline = obj(json!({"outputStyle": "old"}));
        let new = obj(json!({"outputStyle": "new"}));
        let (result, c) =
            plan_settings(&baseline, &obj(json!({"outputStyle": "old"})), &new).unwrap();
        assert_eq!((result.clone(), c.len()), (Some(new.clone()), 0));
        // Re-running against the applied file plans no change.
        let again = plan_settings(&new, &result.clone().unwrap(), &new).unwrap();
        assert_eq!(again, (result, vec![]));
        let user = obj(json!({"outputStyle": "user"}));
        assert_eq!(
            plan_settings(&baseline, &user, &obj(json!({"outputStyle": "old"}))).unwrap(),
            (Some(user.clone()), vec![])
        );
        assert_eq!(
            plan_settings(&baseline, &user, &new).unwrap(),
            (None, vec!["outputStyle".to_string()])
        );
    }

    #[test]
    fn claude_settings_missing_null_and_boolean_are_distinct() {
        let p = |b, c, d| plan_settings(&obj(b), &obj(c), &obj(d)).unwrap();
        assert_eq!(
            p(
                json!({}),
                json!({"outputStyle": null}),
                json!({"outputStyle": "new"})
            ),
            (None, vec!["outputStyle".to_string()])
        );
        assert_eq!(
            p(
                json!({"outputStyle": null}),
                json!({"outputStyle": null}),
                json!({})
            ),
            (Some(Map::new()), vec![])
        );
        assert_eq!(
            p(
                json!({"outputStyle": true}),
                json!({"outputStyle": 1}),
                json!({"outputStyle": false})
            ),
            (None, vec!["outputStyle".to_string()])
        );
    }

    #[test]
    fn claude_settings_refuses_unmanaged_keys_and_keeps_order() {
        for (b, d) in [
            (json!({}), json!({"permissions": {}})),
            (json!({"hooks": {}}), json!({})),
        ] {
            let err = plan_settings(&obj(b), &Map::new(), &obj(d)).unwrap_err();
            assert_eq!(
                crate::contract::classify(&err),
                crate::contract::Kind::InvalidInput
            );
        }
        // Removing an owned key keeps the user's other keys where they were.
        let current: Map<String, Value> =
            serde_json::from_str(r#"{"z":1,"outputStyle":"MoAI","a":2,"m":3}"#).unwrap();
        let (result, _) =
            plan_settings(&obj(json!({"outputStyle": "MoAI"})), &current, &Map::new()).unwrap();
        let keys: Vec<&String> = result.as_ref().unwrap().keys().collect();
        let mut expected: Vec<&String> = current.keys().filter(|k| *k != "outputStyle").collect();
        assert_eq!(keys, expected);
        expected.sort();
        assert_eq!(result.unwrap().len(), expected.len());
    }

    #[test]
    fn projection_refuses_overrides_grants_and_bad_values() {
        let skills = tempfile::TempDir::new().unwrap();
        for (k, v, msg) in [
            (
                "name",
                json!("other"),
                "canonical frontmatter field override refused",
            ),
            (
                "description",
                json!("x"),
                "canonical frontmatter field override refused",
            ),
            (
                "allowed-tools",
                json!("Bash"),
                "unknown claude frontmatter field",
            ),
            ("hooks", json!("x"), "unknown claude frontmatter field"),
            ("effort", json!("ultra"), "invalid claude frontmatter value"),
            (
                "user-invocable",
                json!("yes"),
                "invalid claude frontmatter value",
            ),
            (
                "when_to_use",
                json!("two\nlines"),
                "invalid claude frontmatter value",
            ),
            ("model", json!(3), "invalid claude frontmatter value"),
        ] {
            let err = projection(&package(), skills.path(), &added(&[(k, v)])).unwrap_err();
            assert_eq!(err, msg, "{k}");
            assert_ne!(
                crate::contract::classify(&err),
                crate::contract::Kind::Internal
            );
        }
    }

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
