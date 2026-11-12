//! Claude Code adapter (Issue #1992, 02-B). `doctor`, `projection` and `settings_plan` only
//! read; `settings_apply` is the one function that writes (a project's `.claude/settings.json`
//! and its ownership receipt).
//!
//! `doctor` predicts whether Claude Code reads a project's `AGENTS.md`, following the rules
//! of https://code.claude.com/docs/en/memory (read 2026-11-08): direct reading needs v2.1.277
//! or later; a `CLAUDE.md`, `.claude/CLAUDE.md` or `CLAUDE.local.md` in the working directory
//! or above suppresses it under the default `claude-md-or-agents-md`, except the user's own
//! `~/.claude/CLAUDE.md`; the `pluginConfigs.<id>.options.instructionFiles` entry is honoured
//! only in user, `--settings` and managed settings, never in project settings; the plugin id
//! is `cc-plugin-agents-md@builtin` from v2.1.285, which still reads the older
//! `agents-md@builtin`, and only `agents-md@builtin` before.
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
    sync::mpsc,
    time::{Duration, Instant},
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

/// How long `claude --version` may run before the probe gives up (#2349).
const VERSION_TIMEOUT: Duration = Duration::from_secs(5);
const VERSION_GROUP_KILLER: &str = "/bin/kill";

/// Outcome of the `claude --version` probe.
enum VersionProbe {
    Found((u32, u32, u32)),
    NoVersion,
    TimedOut { group_signal_sent: bool },
}

/// The leader has not been reaped at any call site, keeping its group id reserved.
fn stop_version_probe(child: &mut std::process::Child, _group_kill_program: &Path) -> bool {
    #[cfg(unix)]
    let group_signal_sent = Command::new(_group_kill_program)
        .args(["-KILL", "--", &format!("-{}", child.id())])
        .stdin(Stdio::null())
        .stdout(Stdio::null())
        .stderr(Stdio::null())
        .status()
        .is_ok_and(|status| status.success());
    #[cfg(not(unix))]
    let group_signal_sent = false;
    let _ = child.kill();
    let _ = child.wait();
    group_signal_sent
}

fn timeout_reason(group_signal_sent: bool) -> &'static str {
    if group_signal_sent {
        "claude --version timed out after 5 s; process-group SIGKILL was acknowledged"
    } else {
        "claude --version timed out after 5 s; process-group termination was not confirmed"
    }
}

/// `claude --version` with a fixed argv, bounded by `VERSION_TIMEOUT`. The probe runs in
/// its own process group; timeout cleanup signals the group before reaping its
/// leader and reports whether the system signal command acknowledged the request.
fn claude_version() -> VersionProbe {
    let mut cmd = Command::new("claude");
    cmd.arg("--version")
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::null());
    #[cfg(unix)]
    std::os::unix::process::CommandExt::process_group(&mut cmd, 0);
    let Ok(mut child) = cmd.spawn() else {
        return VersionProbe::NoVersion;
    };
    let stdout = child.stdout.take();
    let (tx, rx) = mpsc::channel();
    // Read on its own thread so the wait below stays bounded.
    std::thread::spawn(move || {
        let mut buf = Vec::new();
        if let Some(pipe) = stdout {
            let _ = pipe
                .take(VERSION_OUTPUT_LIMIT as u64 + 1)
                .read_to_end(&mut buf);
        }
        let _ = tx.send(buf);
    });
    let deadline = Instant::now() + VERSION_TIMEOUT;
    // Wait for stdout before reaping the leader: a launcher may exit while a
    // descendant still holds the pipe, and its process group must remain reserved.
    let out = match rx.recv_timeout(deadline.saturating_duration_since(Instant::now())) {
        Ok(out) => out,
        Err(mpsc::RecvTimeoutError::Timeout) => {
            let group_signal_sent = stop_version_probe(&mut child, Path::new(VERSION_GROUP_KILLER));
            return VersionProbe::TimedOut { group_signal_sent };
        }
        Err(mpsc::RecvTimeoutError::Disconnected) => {
            stop_version_probe(&mut child, Path::new(VERSION_GROUP_KILLER));
            return VersionProbe::NoVersion;
        }
    };
    let status = loop {
        match child.try_wait() {
            Ok(Some(status)) => break Some(status),
            Ok(None) if Instant::now() < deadline => std::thread::sleep(Duration::from_millis(20)),
            Ok(None) => break None,
            Err(_) => {
                stop_version_probe(&mut child, Path::new(VERSION_GROUP_KILLER));
                return VersionProbe::NoVersion;
            }
        }
    };
    let Some(status) = status else {
        let group_signal_sent = stop_version_probe(&mut child, Path::new(VERSION_GROUP_KILLER));
        return VersionProbe::TimedOut { group_signal_sent };
    };
    if !status.success() || out.len() > VERSION_OUTPUT_LIMIT {
        return VersionProbe::NoVersion;
    }
    match std::str::from_utf8(&out).ok().and_then(parse_version) {
        Some(v) => VersionProbe::Found(v),
        None => VersionProbe::NoVersion,
    }
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

/// The line without its inline code spans: text between a backtick and the next one is
/// dropped, and an unclosed span drops the rest of the line.
fn outside_code_spans(line: &str) -> String {
    line.split('`').step_by(2).collect::<Vec<_>>().join(" ")
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
            && outside_code_spans(line)
                .split_whitespace()
                .any(|t| t.starts_with('@') && t.ends_with("AGENTS.md"))
    })
}

/// The user-settings `instructionFiles` value under the first of `plugin_ids` that has one:
/// `Ok(None)` when absent, `Err` when the file cannot be read as JSON.
fn user_instruction_files(home: &Path, plugin_ids: &[&str]) -> Result<Option<String>> {
    let path = home.join(".claude/settings.json");
    if fs::symlink_metadata(&path).is_err() {
        return Ok(None);
    }
    let bytes = read_bounded(&path).ok_or("user settings unreadable")?;
    let v: Value = serde_json::from_slice(&bytes).map_err(|_| "user settings are not JSON")?;
    for id in plugin_ids {
        match &v["pluginConfigs"][*id]["options"]["instructionFiles"] {
            Value::Null => continue,
            Value::String(s) => return Ok(Some(s.clone())),
            _ => return Err("instructionFiles is not a string".into()),
        }
    }
    Ok(None)
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
    let home = std::env::var_os("HOME")
        .filter(|h| !h.is_empty())
        .map(PathBuf::from);
    // The user's own ~/.claude/CLAUDE.md loads alongside AGENTS.md and does not suppress it.
    let user_claude = home
        .as_ref()
        .and_then(|h| fs::canonicalize(h).ok())
        .map(|h| h.join(".claude/CLAUDE.md"));
    let agents = found_upward(&dir, &AGENTS_FILES);
    let claude_files: Vec<PathBuf> = found_upward(&dir, &SUPPRESSING)
        .into_iter()
        .filter(|p| Some(p) != user_claude.as_ref())
        .collect();
    let importing: Vec<PathBuf> = claude_files
        .iter()
        .filter(|p| imports_agents(p))
        .cloned()
        .collect();
    let probe = claude_version();
    let version = match probe {
        VersionProbe::Found(v) => Some(v),
        VersionProbe::NoVersion | VersionProbe::TimedOut { .. } => None,
    };
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
        let reason = match probe {
            VersionProbe::TimedOut { group_signal_sent } => timeout_reason(group_signal_sent),
            _ => "claude --version gave no version",
        };
        set(&mut out, "unknown", reason);
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
    let plugin_ids: &[&str] = if version >= PLUGIN_ID_SINCE {
        &["cc-plugin-agents-md@builtin", "agents-md@builtin"]
    } else {
        &["agents-md@builtin"]
    };
    let setting = match &home {
        Some(home) => match user_instruction_files(home, plugin_ids) {
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
/// when the user asks for it; it is never set by default.
/// `permissions` (#2244) is managed only with narrowing values; see [`check_permissions`].
/// `hooks` (rc.2 I1 #2238) is owned as a whole key and written only by `hooks::apply`, from the
/// fixed MoAI hook set: a project whose `hooks` the user already set plans as a conflict, never
/// a merge. A length-free slice, so two branches that each add a key merge without a silent
/// length clash (PR #2044).
pub const MANAGED_SETTINGS: &[&str] = &["outputStyle", "statusLine", "permissions", "hooks"];

/// The managed keys a caller may set through `claude-settings-plan` / `-apply` (#2239, #2244):
/// every managed key except `hooks`.
pub const USER_SETTINGS: &[&str] = &["outputStyle", "statusLine", "permissions"];

/// REQ-I2-01: names every key outside `allowed` and refuses; nothing is planned.
fn refuse_unlisted<'a>(keys: impl Iterator<Item = &'a str>, allowed: &[&str]) -> Result<()> {
    let mut refused: Vec<&str> = keys.filter(|k| !allowed.contains(k)).collect();
    if refused.is_empty() {
        return Ok(());
    }
    refused.sort_unstable();
    refused.dedup();
    let shown: Vec<String> = refused
        .iter()
        .take(8)
        .map(|k| k.chars().filter(|c| !c.is_control()).take(64).collect())
        .collect();
    Err(format!(
        "unmanaged claude settings key refused: {}{}; allowed: {}",
        shown.join(", "),
        if refused.len() > 8 { ", …" } else { "" },
        allowed.join(", ")
    ))
}

/// The desired map for `claude-settings-plan` / `-apply` (#2239): the caller may set
/// `outputStyle`, `statusLine` and `permissions`, never `hooks` — only `hooks-apply` writes that key, from the
/// fixed MoAI hook set (#2238). A `hooks` key MoAI already owns is carried over unchanged, so a
/// settings change never drops the generated hook set.
pub fn user_settings_desired(
    project: &Path,
    desired: &Map<String, Value>,
) -> Result<Map<String, Value>> {
    refuse_unlisted(desired.keys().map(String::as_str), USER_SETTINGS)?;
    let mut out = desired.clone();
    if let Some(hooks) = owned_settings(project)?.get("hooks") {
        out.insert("hooks".into(), hooks.clone());
    }
    Ok(out)
}

/// Permission modes a managed `permissions.defaultMode` may take. `auto` and
/// `bypassPermissions` skip prompts and are refused: package content never widens permissions.
/// Mode list from https://code.claude.com/docs/en/permissions (read 2026-11-09).
const PERMISSION_MODES: [&str; 4] = ["default", "acceptEdits", "plan", "dontAsk"];
/// Bounds for one rule list (`allow`, `deny`, `ask`) and one rule.
const PERMISSION_RULES_MAX: usize = 256;
const PERMISSION_RULE_CHARS_MAX: usize = 256;

/// Validates a desired `permissions` value: an object with only `allow`, `deny`, `ask`
/// (arrays of non-empty one-line rules) and `defaultMode` (a narrowing mode). Any other
/// sub-key (`additionalDirectories`, `disableBypassPermissionsMode`, ...) is undecided and
/// refused.
fn check_permissions(value: &Value) -> Result<()> {
    let Value::Object(map) = value else {
        return Err("permissions must be an object".into());
    };
    for (key, v) in map {
        match key.as_str() {
            "allow" | "deny" | "ask" => {
                let Value::Array(rules) = v else {
                    return Err(format!("permissions.{key} must be an array of rules"));
                };
                if rules.len() > PERMISSION_RULES_MAX {
                    return Err(format!(
                        "permissions.{key} must be at most {PERMISSION_RULES_MAX} rules"
                    ));
                }
                for rule in rules {
                    let ok = rule.as_str().is_some_and(|s| {
                        !s.trim().is_empty()
                            && s.chars().count() <= PERMISSION_RULE_CHARS_MAX
                            && !s.chars().any(char::is_control)
                    });
                    if !ok {
                        return Err(format!(
                            "permissions.{key} rules must be non-empty one-line strings of at most {PERMISSION_RULE_CHARS_MAX} characters"
                        ));
                    }
                }
            }
            "defaultMode" => {
                if !v.as_str().is_some_and(|m| PERMISSION_MODES.contains(&m)) {
                    return Err(
                        "permissions.defaultMode must be default, acceptEdits, plan or dontAsk; widening modes are refused"
                            .into(),
                    );
                }
            }
            _ => return Err(format!("unknown permissions key {key} refused")),
        }
    }
    Ok(())
}

/// A planned settings object (`None` on conflict) and the conflicting keys.
pub type SettingsPlan = (Option<Map<String, Value>>, Vec<String>);

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
) -> Result<SettingsPlan> {
    // REQ-I2-01: only the managed keys are written; any other key is refused and named.
    refuse_unlisted(
        baseline.keys().chain(desired.keys()).map(String::as_str),
        MANAGED_SETTINGS,
    )?;
    if let Some(p) = desired.get("permissions") {
        check_permissions(p)?;
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

const SETTINGS_FILE: &str = ".claude/settings.json";
/// Ownership receipt under the project state root; 04-B UPDATE reads it as the baseline.
const SETTINGS_RECEIPT: &str = "manifests/claude-settings.json";

/// Size bound for the settings file and its receipt, the same as `agents::write_replace`
/// re-reads, so a plan this module offers is one the writer accepts.
const SETTINGS_LIMIT: u64 = 262_144;

pub(crate) fn read_existing(path: &Path) -> Result<Option<Vec<u8>>> {
    match fs::symlink_metadata(path) {
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(None),
        Err(_) => Err("target metadata unavailable".into()),
        Ok(m) if m.file_type().is_symlink() => Err("symlink target refused".into()),
        Ok(m) if !m.is_file() || m.len() > SETTINGS_LIMIT => {
            Err("target is not a bounded regular file".into())
        }
        Ok(_) => fs::read(path)
            .map(Some)
            .map_err(|_| "target read failed".into()),
    }
}

pub(crate) fn refuse_symlink_dir(path: &Path) -> Result<()> {
    match fs::symlink_metadata(path) {
        Ok(m) if m.file_type().is_symlink() => Err("symlink target refused".into()),
        _ => Ok(()),
    }
}

/// One planned settings change: everything the plan digest binds and apply re-checks.
struct SettingsState {
    settings: PathBuf,
    receipt: PathBuf,
    before: Option<Vec<u8>>,
    receipt_before: Option<Vec<u8>>,
    owned: Map<String, Value>,
    current: Map<String, Value>,
    candidate: Option<Map<String, Value>>,
    conflicts: Vec<String>,
    /// New settings bytes; `None` when the file stays as it is.
    after: Option<Vec<u8>>,
    digest: String,
}

/// The project's `.claude/settings.json` is the user's own settings file when the project is
/// the home directory, or when its `.claude` is `CLAUDE_CONFIG_DIR`. Writing there is a
/// home write, which is blocked (#2239); planning it is refused too, so no digest exists.
fn refuse_user_scope(root: &Path) -> Result<()> {
    let canon = |v: Option<std::ffi::OsString>| v.and_then(|p| fs::canonicalize(p).ok());
    let home = canon(std::env::var_os("HOME"));
    let config = canon(std::env::var_os("CLAUDE_CONFIG_DIR"));
    let claude_dir = fs::canonicalize(root.join(".claude")).ok();
    if home.as_deref() == Some(root)
        || (config.is_some() && claude_dir.is_some() && config == claude_dir)
    {
        return Err("user-scope claude settings refused: the project is the home directory or its .claude is CLAUDE_CONFIG_DIR; home writes are blocked".into());
    }
    Ok(())
}

fn settings_state(project: &Path, desired: &Map<String, Value>) -> Result<SettingsState> {
    let root = fs::canonicalize(project).map_err(|_| "project directory unavailable")?;
    if !root.is_dir() {
        return Err("project must be a directory".into());
    }
    refuse_user_scope(&root)?;
    let state_dir = root.join(crate::PROJECT_STATE_DIR);
    for dir in [
        root.join(".claude"),
        state_dir.clone(),
        state_dir.join("manifests"),
    ] {
        refuse_symlink_dir(&dir)?;
    }
    let (settings, receipt) = (root.join(SETTINGS_FILE), state_dir.join(SETTINGS_RECEIPT));
    let before = read_existing(&settings)?;
    let receipt_before = read_existing(&receipt)?;
    let owned = match &receipt_before {
        None => Map::new(),
        Some(bytes) => resolve_owned(bytes, before.as_deref().map(crate::digest))?,
    };
    let current = match &before {
        None => Map::new(),
        Some(bytes) => match serde_json::from_slice::<Value>(bytes) {
            Ok(Value::Object(m)) => m,
            _ => return Err("existing settings.json requires preservation review".into()),
        },
    };
    let (candidate, conflicts) = plan_settings(&owned, &current, desired)?;
    let after = candidate.as_ref().filter(|c| **c != current).map(|c| {
        let mut b = serde_json::to_vec_pretty(c).expect("JSON object");
        b.push(b'\n');
        b
    });
    if after
        .as_ref()
        .is_some_and(|b| b.len() as u64 > SETTINGS_LIMIT)
    {
        return Err("rendered settings exceed the 256 KiB limit; refused".into());
    }
    use crate::spec::Canonical;
    let canonical =
        |m: &Map<String, Value>| crate::spec::SortedKeys.bytes(&Value::Object(m.clone()));
    let digest = crate::digest(
        &serde_json::to_vec(&(
            1,
            crate::digest(&canonical(desired)),
            before.as_deref().map(crate::digest),
            receipt_before.as_deref().map(crate::digest),
            after.as_deref().map(crate::digest),
            &conflicts,
        ))
        .expect("plan tuple"),
    );
    Ok(SettingsState {
        settings,
        receipt,
        before,
        receipt_before,
        owned,
        current,
        candidate,
        conflicts,
        after,
        digest,
    })
}

/// The managed keys MoAI owns in `PROJECT/.claude/settings.json` now, from its receipt (empty
/// when there is none). A caller that changes one key starts from this, so its plan never drops
/// another key MoAI already owns.
pub fn owned_settings(project: &Path) -> Result<Map<String, Value>> {
    Ok(settings_state(project, &Map::new())?.owned)
}

/// Whether git ignores `PROJECT/.claude/settings.json` (#2581): `ignored`, `not_ignored`, or
/// `unknown` when PROJECT is not in a git work tree or git fails. Index-aware (the default of
/// `git check-ignore`): a tracked file that matches an ignore pattern is `not_ignored`. The
/// caller's git environment is scrubbed so PROJECT is the repository read (#2054).
fn settings_ignored(root: &Path) -> &'static str {
    let mut cmd = Command::new("git");
    cmd.arg("--no-optional-locks").arg("-C").arg(root).args([
        "check-ignore",
        "-q",
        "--",
        SETTINGS_FILE,
    ]);
    crate::identity::scrub_git_env(&mut cmd);
    let status = cmd
        .env("GIT_OPTIONAL_LOCKS", "0")
        .env("GIT_TERMINAL_PROMPT", "0")
        .stdin(Stdio::null())
        .stdout(Stdio::null())
        .stderr(Stdio::null())
        .status();
    match status.ok().and_then(|s| s.code()) {
        Some(0) => "ignored",
        Some(1) => "not_ignored",
        _ => "unknown",
    }
}

/// What a session started in a linked worktree of this repository sees of the settings file
/// (#2581). Claude Code reads the shared project settings from the session's working folder
/// (https://code.claude.com/docs/en/settings, read 2026-11-12), and a linked worktree carries
/// only the files git checks out, so an ignored settings file never reaches it.
fn worktree_sessions(ignored: &str) -> &'static str {
    match ignored {
        "ignored" => "not_reached",
        "not_ignored" => "reached_when_committed",
        _ => "unknown",
    }
}

/// Previews setting `desired` (managed keys only) in `PROJECT/.claude/settings.json`. Never
/// writes. The returned `plan_digest` binds the current file, the receipt and the result.
/// `settings_ignored` and `worktree_sessions` report whether the file reaches sessions started
/// in a linked worktree (#2581); they are facts for the caller and do not change the plan.
pub fn settings_plan(project: &Path, desired: &Map<String, Value>) -> Result<Value> {
    let s = settings_state(project, desired)?;
    let root = fs::canonicalize(project).map_err(|_| "project directory unavailable")?;
    let ignored = settings_ignored(&root);
    let action = match (&s.candidate, &s.after, &s.before) {
        (None, _, _) => "conflict",
        (Some(_), None, _) => "unchanged",
        (Some(_), Some(_), None) => "create",
        (Some(_), Some(_), Some(_)) => "update",
    };
    Ok(json!({
        "action": action,
        "settings_path": SETTINGS_FILE,
        "conflicts": s.conflicts,
        "expected_sha256": s.before.as_deref().map(crate::digest),
        "planned_sha256": s.after.as_deref().map(crate::digest),
        "plan_digest": s.digest,
        "settings_ignored": ignored,
        "worktree_sessions": worktree_sessions(ignored),
        "writes": false,
    }))
}

/// Receipt bytes. `pending` is the ownership MoAI intends once the settings file carries the
/// given sha256; it is written before the settings file and cleared after.
fn receipt_bytes(
    owned: &Map<String, Value>,
    settings_sha256: Option<&str>,
    pending: Option<(&Map<String, Value>, &str)>,
) -> Vec<u8> {
    let mut r = json!({
        "schema_version": 1,
        "settings_path": SETTINGS_FILE,
        "owned": owned,
        "settings_sha256": settings_sha256,
    });
    if let Some((owned, sha)) = pending {
        r["pending"] = json!({"owned": owned, "settings_sha256": sha});
    }
    let mut bytes = serde_json::to_vec_pretty(&r).expect("receipt");
    bytes.push(b'\n');
    bytes
}

/// The ownership a receipt grants for the settings file whose sha256 is `current`. A pending
/// receipt left by an interrupted apply resolves by which side the file is on: the intended
/// one (the settings write landed) or the recorded one (it did not). Anything else is a
/// change nobody recorded and needs review.
fn resolve_owned(bytes: &[u8], current: Option<String>) -> Result<Map<String, Value>> {
    let review = || "claude settings receipt requires preservation review".to_string();
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
        (&pending["owned"], &pending["settings_sha256"])
    else {
        return Err(review());
    };
    if current.as_deref() == Some(intended_sha.as_str()) {
        Ok(intended.clone())
    } else if current.as_deref() == r["settings_sha256"].as_str() {
        Ok(owned.clone())
    } else {
        Err(review())
    }
}

/// Applies a previewed settings plan: re-plans against the files now, requires the same
/// digest, replaces the settings file only when it changes, then records ownership. A key is
/// owned only when MoAI already owned it or set it now; an equal value the user already had is
/// left unowned. Before the settings file changes, the receipt records the intended ownership
/// as `pending`; an interruption after either write is resolved on the next plan by the file's
/// sha256, so a re-run completes it without adopting or losing a key.
pub fn settings_apply(
    project: &Path,
    desired: &Map<String, Value>,
    plan_digest: &str,
    permission: &str,
) -> Result<Value> {
    // Authorization stays with the host; the reference is caller metadata, not a verifier.
    if ["", "null", "placeholder"].contains(&permission.trim())
        || permission.len() > 128
        || permission.chars().any(char::is_control)
    {
        return Err("permission reference required".into());
    }
    let s = settings_state(project, desired)?;
    if !s.conflicts.is_empty() {
        return Err(format!(
            "conflict in {SETTINGS_FILE}: {} changed outside MoAI; nothing applied",
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
    let final_sha = final_bytes.map(crate::digest);
    let receipt = receipt_bytes(&owned, final_sha.as_deref(), None);
    let rel = format!("{}/{SETTINGS_RECEIPT}", crate::PROJECT_STATE_DIR);
    let mut files = Vec::new();
    let mut receipt_now = s.receipt_before.clone();
    if let Some(after) = &s.after {
        let after_sha = crate::digest(after);
        let pending = receipt_bytes(&s.owned, before_sha.as_deref(), Some((&owned, &after_sha)));
        crate::agents::write_replace(&s.receipt, receipt_now.as_deref(), &pending)?;
        receipt_now = Some(pending);
        crate::agents::write_replace(&s.settings, s.before.as_deref(), after)?;
        files.push(SETTINGS_FILE);
    }
    if receipt_now.as_deref() != Some(receipt.as_slice()) {
        if let Err(e) = crate::agents::write_replace(&s.receipt, receipt_now.as_deref(), &receipt) {
            return Err(if files.is_empty() {
                e
            } else {
                format!("partial apply: applied [{SETTINGS_FILE}]; failed at {rel}: {e}")
            });
        }
    }
    let created: Vec<&str> = if s.before.is_none() {
        files.clone()
    } else {
        vec![]
    };
    Ok(json!({
        "state": if files.is_empty() { "unchanged" } else { "applied" },
        "files_changed": files,
        "files_created": created,
        "owned_keys": owned.keys().collect::<Vec<_>>(),
        "settings_sha256_before": s.before.as_deref().map(crate::digest),
        "settings_sha256_after": final_bytes.map(crate::digest),
        "host_loaded": "not_run",
        "execution_verified": "not_run",
    }))
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
        "target": text(&skills_root.join(crate::update::SKILL_DIR))?,
        "reload_skills_required": reload,
        "files_created": false,
        "host_loaded": "not_run",
        "execution_verified": "not_run",
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    /// #2511: the documented five-second `claude --version` probe is pinned here, by value,
    /// so the CLI tests in `tests/product_cli.rs` can bound wall time generously under load.
    #[test]
    fn version_probe_deadline_is_five_seconds() {
        assert_eq!(VERSION_TIMEOUT, Duration::from_secs(5));
    }

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
            // `permissions` (#2244) and `hooks` (#2238) are managed; `env` stays unmanaged.
            (json!({}), json!({"env": {}})),
            (json!({"env": {}}), json!({})),
        ] {
            let err = plan_settings(&obj(b), &Map::new(), &obj(d)).unwrap_err();
            assert_eq!(
                crate::contract::classify(&err),
                crate::contract::Kind::InvalidInput
            );
        }
        // `hooks` is managed since #2238: a desired value plans, it is not refused.
        let (planned, conflicts) =
            plan_settings(&Map::new(), &Map::new(), &obj(json!({"hooks": {}}))).unwrap();
        assert!(conflicts.is_empty());
        assert_eq!(planned.unwrap()["hooks"], json!({}));
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

    // #2244 (I7): `permissions` is a managed key; values that widen access are refused.
    #[test]
    fn claude_settings_plans_a_narrowing_permissions_key() {
        let desired = obj(json!({"permissions": {
            "allow": ["Bash(cargo test *)"], "deny": ["Bash(git push *)"], "ask": [],
            "defaultMode": "plan"}}));
        let (result, conflicts) = plan_settings(&Map::new(), &Map::new(), &desired).unwrap();
        assert!(conflicts.is_empty());
        assert_eq!(result.unwrap()["permissions"], desired["permissions"]);
        for mode in ["default", "acceptEdits", "plan", "dontAsk"] {
            let d = obj(json!({"permissions": {"defaultMode": mode}}));
            assert!(
                plan_settings(&Map::new(), &Map::new(), &d).is_ok(),
                "{mode}"
            );
        }
        // A user-authored permissions key MoAI never owned is a conflict, not an overwrite.
        let current = obj(json!({"permissions": {"allow": ["Read"]}}));
        let (result, conflicts) = plan_settings(&Map::new(), &current, &desired).unwrap();
        assert_eq!((result, conflicts), (None, vec!["permissions".to_string()]));
    }

    #[test]
    fn claude_settings_refuses_permissions_that_widen_or_are_malformed() {
        let rule = |n: usize| "R".repeat(n);
        let many = |n: usize| vec!["Read"; n];
        let refused = [
            json!({"permissions": {"defaultMode": "bypassPermissions"}}),
            json!({"permissions": {"defaultMode": "auto"}}),
            json!({"permissions": {"defaultMode": "unknownMode"}}),
            json!({"permissions": {"additionalDirectories": ["/"]}}),
            json!({"permissions": {"disableBypassPermissionsMode": "disable"}}),
            json!({"permissions": {"allow": "Read"}}),
            json!({"permissions": {"allow": [1]}}),
            json!({"permissions": {"deny": [""]}}),
            json!({"permissions": {"ask": [rule(257)]}}),
            json!({"permissions": {"allow": many(257)}}),
            json!({"permissions": {"allow": ["Read\nBash"]}}),
            json!({"permissions": ["Read"]}),
        ];
        for d in refused {
            let err = plan_settings(&Map::new(), &Map::new(), &obj(d.clone())).unwrap_err();
            assert_eq!(
                crate::contract::classify(&err),
                crate::contract::Kind::InvalidInput,
                "{d}: {err}"
            );
        }
        // Edges: 256 characters per rule and 256 rules per list are accepted.
        for d in [
            json!({"permissions": {"ask": [rule(256)]}}),
            json!({"permissions": {"allow": many(256)}}),
        ] {
            assert!(plan_settings(&Map::new(), &Map::new(), &obj(d)).is_ok());
        }
    }

    // SETTINGS-PRESERVE: every case runs in a temporary project; no real settings are touched.
    fn project_with(settings: Option<&str>) -> tempfile::TempDir {
        let p = tempfile::TempDir::new().unwrap();
        if let Some(text) = settings {
            fs::create_dir(p.path().join(".claude")).unwrap();
            fs::write(p.path().join(SETTINGS_FILE), text).unwrap();
        }
        p
    }
    fn sha(p: &Path) -> Option<String> {
        fs::read(p).ok().map(|b| crate::digest(&b))
    }
    fn apply_now(p: &Path, desired: &Map<String, Value>) -> Result<Value> {
        let plan = settings_plan(p, desired)?;
        settings_apply(
            p,
            desired,
            plan["plan_digest"].as_str().unwrap(),
            "test-ref",
        )
    }
    const USER_FILE: &str = "{\n  \"model\": \"sonnet\",\n  \"hooks\": {\n    \"Stop\": [\n      {\n        \"user\": true\n      }\n    ]\n  },\n  \"zz-unknown\": 1\n}\n";

    #[test]
    fn claude_settings_apply_creates_preserves_and_is_idempotent() {
        // Created from nothing; the receipt owns what MoAI set.
        let p = project_with(None);
        let style = obj(json!({"outputStyle": "MoAI"}));
        let r = apply_now(p.path(), &style).unwrap();
        assert_eq!(r["files_created"], json!([SETTINGS_FILE]));
        assert_eq!(
            fs::read_to_string(p.path().join(SETTINGS_FILE)).unwrap(),
            "{\n  \"outputStyle\": \"MoAI\"\n}\n"
        );
        // An existing user file keeps every byte; the managed key is appended at the end.
        let p = project_with(Some(USER_FILE));
        let line = obj(json!({"statusLine": {"type": "command", "command": "moai-statusline"}}));
        let r = apply_now(p.path(), &line).unwrap();
        assert_eq!(r["state"], json!("applied"));
        assert_eq!(r["owned_keys"], json!(["statusLine"]));
        let after = fs::read_to_string(p.path().join(SETTINGS_FILE)).unwrap();
        let expected = USER_FILE.replacen(
            "  \"zz-unknown\": 1\n}",
            "  \"zz-unknown\": 1,\n  \"statusLine\": {\n    \"type\": \"command\",\n    \"command\": \"moai-statusline\"\n  }\n}",
            1,
        );
        assert_eq!(after, expected);
        // Re-running plans and writes nothing: same bytes, same receipt.
        let receipt = p.path().join(".moai-cli").join(SETTINGS_RECEIPT);
        let (s1, r1) = (sha(&p.path().join(SETTINGS_FILE)), sha(&receipt));
        assert_eq!(
            settings_plan(p.path(), &line).unwrap()["action"],
            json!("unchanged")
        );
        let again = apply_now(p.path(), &line).unwrap();
        assert_eq!(again["state"], json!("unchanged"));
        assert_eq!(
            (sha(&p.path().join(SETTINGS_FILE)), sha(&receipt)),
            (s1, r1)
        );
        // Dropping the key from `desired` removes only what MoAI owned.
        apply_now(p.path(), &Map::new()).unwrap();
        assert_eq!(
            fs::read_to_string(p.path().join(SETTINGS_FILE)).unwrap(),
            USER_FILE
        );
    }

    #[test]
    fn claude_settings_apply_refuses_conflicts_stale_plans_and_bad_targets() {
        // A user statusLine is a conflict; the file stays byte-identical.
        let user = "{\"statusLine\":{\"type\":\"command\",\"command\":\"orca\"}}";
        let p = project_with(Some(user));
        let path = p.path().join(SETTINGS_FILE);
        let before = sha(&path);
        let desired = obj(json!({"statusLine": {"type": "command", "command": "moai"}}));
        let plan = settings_plan(p.path(), &desired).unwrap();
        assert_eq!(plan["action"], json!("conflict"));
        let err = settings_apply(
            p.path(),
            &desired,
            plan["plan_digest"].as_str().unwrap(),
            "r",
        )
        .unwrap_err();
        assert_eq!(
            crate::contract::classify(&err),
            crate::contract::Kind::Conflict,
            "{err}"
        );
        assert_eq!(sha(&path), before);
        // A user edit after the preview makes the plan stale.
        let p = project_with(Some(USER_FILE));
        let style = obj(json!({"outputStyle": "MoAI"}));
        let plan = settings_plan(p.path(), &style).unwrap();
        fs::write(p.path().join(SETTINGS_FILE), "{\"model\":\"opus\"}\n").unwrap();
        let err = settings_apply(p.path(), &style, plan["plan_digest"].as_str().unwrap(), "r")
            .unwrap_err();
        assert_eq!(
            crate::contract::classify(&err),
            crate::contract::Kind::StalePlan
        );
        assert_eq!(
            fs::read_to_string(p.path().join(SETTINGS_FILE)).unwrap(),
            "{\"model\":\"opus\"}\n"
        );
        // An identical value the user already had is not adopted, and the file is untouched.
        let p = project_with(Some("{\"outputStyle\": \"MoAI\"}\n"));
        let r = apply_now(p.path(), &style).unwrap();
        assert_eq!(
            (r["state"].clone(), r["owned_keys"].clone()),
            (json!("unchanged"), json!([]))
        );
        // Placeholder permission, non-object JSON, and a symlinked settings file are refused.
        let p = project_with(Some("[1]"));
        let plan_err = settings_plan(p.path(), &style).unwrap_err();
        assert_eq!(
            crate::contract::classify(&plan_err),
            crate::contract::Kind::Conflict
        );
        let p = project_with(None);
        let plan = settings_plan(p.path(), &style).unwrap();
        let err = settings_apply(
            p.path(),
            &style,
            plan["plan_digest"].as_str().unwrap(),
            "placeholder",
        )
        .unwrap_err();
        assert_eq!(
            crate::contract::classify(&err),
            crate::contract::Kind::PermissionReferenceMissing
        );
        #[cfg(unix)]
        {
            let outside = tempfile::TempDir::new().unwrap();
            fs::write(outside.path().join("s.json"), "{}").unwrap();
            fs::create_dir(p.path().join(".claude")).unwrap();
            std::os::unix::fs::symlink(outside.path().join("s.json"), p.path().join(SETTINGS_FILE))
                .unwrap();
            assert!(settings_plan(p.path(), &style).is_err());
            assert_eq!(
                fs::read_to_string(outside.path().join("s.json")).unwrap(),
                "{}"
            );
        }
        assert!(!p.path().join(".moai-cli").exists());
    }

    /// Codex review #2044 (1): an apply interrupted between the settings write and the final
    /// receipt write is completed by a re-run, without losing or adopting ownership.
    #[test]
    fn claude_settings_interrupted_apply_resumes_from_the_pending_receipt() {
        let line = obj(json!({"statusLine": {"type": "command", "command": "moai-statusline"}}));
        let mut landed = obj(serde_json::from_str(USER_FILE).unwrap());
        landed.insert("statusLine".into(), line["statusLine"].clone());
        let mut landed_bytes = serde_json::to_vec_pretty(&landed).unwrap();
        landed_bytes.push(b'\n');
        let user_sha = crate::digest(USER_FILE.as_bytes());
        let pending = receipt_bytes(
            &Map::new(),
            Some(&user_sha),
            Some((&line, &crate::digest(&landed_bytes))),
        );
        let setup = |settings: &[u8]| {
            let p = project_with(None);
            fs::create_dir_all(p.path().join(".claude")).unwrap();
            fs::write(p.path().join(SETTINGS_FILE), settings).unwrap();
            let receipt = p.path().join(".moai-cli").join(SETTINGS_RECEIPT);
            fs::create_dir_all(receipt.parent().unwrap()).unwrap();
            fs::write(&receipt, &pending).unwrap();
            (p, receipt)
        };
        // The settings write landed: the re-run only finishes the receipt and owns the key.
        let (p, receipt) = setup(&landed_bytes);
        let r = apply_now(p.path(), &line).unwrap();
        assert_eq!(r["state"], json!("unchanged"));
        assert_eq!(r["owned_keys"], json!(["statusLine"]));
        let done: Value = serde_json::from_slice(&fs::read(&receipt).unwrap()).unwrap();
        assert!(done.get("pending").is_none());
        // ... so removing the key later removes MoAI's value and restores the user's bytes.
        apply_now(p.path(), &Map::new()).unwrap();
        assert_eq!(
            fs::read_to_string(p.path().join(SETTINGS_FILE)).unwrap(),
            USER_FILE
        );
        // The settings write did not land: the re-run applies it as if for the first time.
        let (p, _) = setup(USER_FILE.as_bytes());
        let r = apply_now(p.path(), &line).unwrap();
        assert_eq!(
            (r["state"].clone(), r["owned_keys"].clone()),
            (json!("applied"), json!(["statusLine"]))
        );
        assert_eq!(
            fs::read(p.path().join(SETTINGS_FILE)).unwrap(),
            landed_bytes
        );
        // Neither side: someone else changed the file meanwhile; nothing is guessed.
        let (p, _) = setup(b"{\"model\":\"opus\"}\n");
        let err = settings_plan(p.path(), &line).unwrap_err();
        assert_eq!(
            crate::contract::classify(&err),
            crate::contract::Kind::Conflict,
            "{err}"
        );
    }

    /// Codex review #2044 (3): the planner offers nothing the writer would refuse.
    #[test]
    fn claude_settings_size_limit_matches_the_writer() {
        let big = format!("{{\"pad\":\"{}\"}}", "x".repeat(SETTINGS_LIMIT as usize));
        let p = project_with(Some(&big));
        let style = obj(json!({"outputStyle": "MoAI"}));
        assert!(settings_plan(p.path(), &style).is_err());
        let p = project_with(None);
        let huge = obj(
            json!({"statusLine": {"type": "command", "command": "x".repeat(SETTINGS_LIMIT as usize)}}),
        );
        let err = settings_plan(p.path(), &huge).unwrap_err();
        assert!(err.contains("256 KiB"), "{err}");
        assert!(!p.path().join(SETTINGS_FILE).exists());
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

    #[cfg(unix)]
    #[test]
    fn version_cleanup_does_not_acknowledge_a_failed_group_signal() {
        let tmp = tempfile::tempdir().unwrap();
        for program in [
            tmp.path().join("missing-kill"),
            PathBuf::from("/usr/bin/false"),
        ] {
            let mut cmd = Command::new("/bin/sleep");
            cmd.arg("25")
                .stdin(Stdio::null())
                .stdout(Stdio::null())
                .stderr(Stdio::null());
            std::os::unix::process::CommandExt::process_group(&mut cmd, 0);
            let mut child = cmd.spawn().unwrap();
            assert!(!stop_version_probe(&mut child, &program));
            assert!(
                child.try_wait().unwrap().is_some(),
                "the leader must be reaped"
            );
        }
    }
}
