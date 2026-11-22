//! The settings write check (SPEC-20261122-settings-write-check): one checked writer for
//! every settings file moai writes. The gate validates a candidate against the forbidden
//! list and the project-only keys before any byte is written, keeps a byte-exact backup of
//! the existing file (mode 0600 in a 0700 folder), re-reads the produced bytes, and swaps
//! atomically; a guarded file needs the user's pick of the previewed plan, and every
//! refusal and warning names the key only, never a value (SPEC REQ-12).

use serde_json::{Map, Value};
use std::path::{Path, PathBuf};
use std::time::{SystemTime, UNIX_EPOCH};

/// The settings file a write targets (project files cannot set what user or managed may).
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum WriteScope {
    User,
    Project,
    Local,
}

/// What the gate found: refusals stop the write, warnings ride along. Messages name keys
/// only, never values (SPEC REQ-12).
#[derive(Debug, Clone, Default, PartialEq, Eq)]
pub struct CheckReport {
    pub refusals: Vec<String>,
    pub warnings: Vec<String>,
}

impl CheckReport {
    fn refuse(&mut self, message: String) {
        self.refusals.push(message);
    }

    fn warn(&mut self, message: String) {
        self.warnings.push(message);
    }
}

/// One planned write the caller asks for.
#[derive(Debug)]
pub struct WriteInput<'a> {
    pub scope: WriteScope,
    pub settings_path: &'a Path,
    pub backup_dir: &'a Path,
    pub candidate: &'a Map<String, Value>,
    pub picked: bool,
}

/// What one checked write produced.
#[derive(Debug)]
pub enum CheckOutcome {
    Written { backup: PathBuf },
    Refused { report: CheckReport },
}

/// Validates a candidate (SPEC REQ-2, REQ-4, REQ-5, REQ-6): the forbidden list, the env
/// shape and, for a project-scope file, the keys a project file cannot set. Every message
/// names the key or entry; values never appear.
pub fn check_candidate(scope: WriteScope, candidate: &Map<String, Value>) -> CheckReport {
    let mut report = CheckReport::default();

    if let Some(hooks) = candidate.get("hooks").and_then(Value::as_object) {
        check_hooks(hooks, &mut report);
    }
    if let Some(permissions) = candidate.get("permissions").and_then(Value::as_object) {
        check_permissions(scope, permissions, &mut report);
    }
    if let Some(attribution) = candidate.get("attribution") {
        if !attribution.is_object() {
            report.refuse("attribution: must be an object".to_string());
        }
    }
    if let Some(env) = candidate.get("env") {
        check_env(scope, env, &mut report);
    }
    if scope == WriteScope::Project {
        check_project_only(candidate, &mut report);
    }
    report
}

/// Hook matchers must be one string (SPEC REQ-4; an array matcher under PreToolUse or
/// PermissionRequest stops that file's other hooks loading).
fn check_hooks(hooks: &serde_json::Map<String, Value>, report: &mut CheckReport) {
    for (event, entries) in hooks {
        let Some(entries) = entries.as_array() else {
            report.refuse(format!("hooks.{event}: must be an array of hook entries"));
            continue;
        };
        for (i, entry) in entries.iter().enumerate() {
            if let Some(matcher) = entry.get("matcher") {
                if !matcher.is_string() {
                    report.refuse(format!("hooks.{event}[{i}].matcher: must be one string"));
                }
            }
        }
    }
}

/// Permission rule shapes (SPEC REQ-4): `Write(path)`/`NotebookEdit(path)` never match, a
/// wildcard before the rest of the command allows more than it looks, and a committed
/// project allowlist carries no bare `Edit`.
fn check_permissions(
    scope: WriteScope,
    permissions: &serde_json::Map<String, Value>,
    report: &mut CheckReport,
) {
    for (list, entries) in permissions {
        if !matches!(list.as_str(), "allow" | "ask" | "deny") {
            continue;
        }
        let Some(entries) = entries.as_array() else {
            continue;
        };
        for entry in entries {
            let Some(rule) = entry.as_str() else {
                continue;
            };
            if rule.starts_with("Write(") || rule.starts_with("NotebookEdit(") {
                report.refuse(format!(
                    "permissions.{list} {rule}: path rules use Edit(path), which governs every file-writing tool"
                ));
            }
            for tool in ["Bash(", "PowerShell("] {
                if let Some(inner) = rule.strip_prefix(tool).and_then(|r| r.strip_suffix(')')) {
                    if inner.split_whitespace().next() == Some("*") {
                        report.refuse(format!(
                            "permissions.{list} {rule}: a wildcard before the rest of the command"
                        ));
                    }
                }
            }
            if scope == WriteScope::Project && list == "allow" && rule == "Edit" {
                report.refuse(
                    "permissions.allow Edit: a committed project allowlist carries no bare Edit"
                        .to_string(),
                );
            }
        }
    }
}

/// `env` shape (SPEC REQ-5): values are literal strings; an empty string means unset. The
/// refusal and the warning name the key; the value is never printed.
fn check_env(scope: WriteScope, env: &Value, report: &mut CheckReport) {
    let Some(env) = env.as_object() else {
        report.refuse("env: must be an object of names to string values".to_string());
        return;
    };
    for (name, value) in env {
        if !value.is_string() {
            report.refuse(format!(
                "env.{name}: value must be a string (settings env values are literal)"
            ));
            continue;
        }
        if scope == WriteScope::Project && project_cannot_set_env(name) {
            report.refuse(format!(
                "env.{name}: a project settings file cannot set this variable; user or managed settings only"
            ));
            continue;
        }
        let text = value.as_str().unwrap_or("");
        if text.starts_with('~') || text.contains('$') {
            report.warn(format!(
                "env.{name}: settings env values are literal (no shell expansion; an empty string means unset)"
            ));
        }
    }
}

/// The `env` names a project file cannot set (settings.md's ignored-variables list).
fn project_cannot_set_env(name: &str) -> bool {
    matches!(
        name,
        "HOME" | "TMPDIR" | "CLAUDE_CONFIG_DIR" | "CLAUDE_CODE_TMPDIR"
    ) || name.starts_with("OTEL_")
}

/// Keys a project settings file cannot set (SPEC REQ-6): the mode keys the CLI ignores or
/// reads from user/managed only, and the security keys a project file may not loosen.
fn check_project_only(candidate: &Map<String, Value>, report: &mut CheckReport) {
    if let Some(mode) = candidate
        .get("permissions")
        .and_then(|p| p.get("defaultMode"))
        .and_then(Value::as_str)
    {
        if matches!(mode, "auto" | "bypassPermissions") {
            report.refuse(format!(
                "permissions.defaultMode {mode}: user or managed settings only"
            ));
        }
    }
    for key in [
        "autoMode",
        "pluginConfigs",
        "skipDangerousModePermissionPrompt",
        "crossSessionInbound",
        "isolatePeerMachines",
        "processWrapper",
    ] {
        if candidate.contains_key(key) {
            report.refuse(format!("{key}: user or managed settings only"));
        }
    }
    if candidate.get("remoteControlAtStartup") == Some(&Value::Bool(true)) {
        report.refuse(
            "remoteControlAtStartup: a project file cannot turn this on; user or managed settings only"
                .to_string(),
        );
    }
    if candidate.get("enableArtifact") == Some(&Value::Bool(true)) {
        report.refuse(
            "enableArtifact: a project file cannot turn this on; user or managed settings only"
                .to_string(),
        );
    }
    let permissions = candidate.get("permissions").and_then(Value::as_object);
    if let Some(p) = permissions {
        for key in ["denyInsteadOfPrompting"] {
            if p.contains_key(key) {
                report.refuse(format!("permissions.{key}: user or managed settings only"));
            }
        }
        for key in [
            "blockReadsOutsideWorkingDirectories",
            "blockWritesOutsideWorkingDirectories",
        ] {
            if p.get(key) == Some(&Value::Bool(false)) {
                report.refuse(format!(
                    "permissions.{key}: a project file cannot loosen this; user or managed settings only"
                ));
            }
        }
    }
}

/// Runs one checked write (SPEC REQ-1, REQ-3, REQ-8): the user's pick of the previewed
/// plan comes first, then the gate, the re-read round-trip, the byte-exact backup, and the
/// atomic swap. A refusal leaves the previous file byte-identical.
pub fn write_checked(input: &WriteInput) -> Result<CheckOutcome, String> {
    if !input.picked {
        let mut report = CheckReport::default();
        report.refuse(
            "the user's pick of the previewed plan is required before any settings write"
                .to_string(),
        );
        return Ok(CheckOutcome::Refused { report });
    }
    let report = check_candidate(input.scope, input.candidate);
    if !report.refusals.is_empty() {
        return Ok(CheckOutcome::Refused { report });
    }

    // REQ-3: the produced bytes must parse back as one JSON object.
    let bytes = format!(
        "{}\n",
        serde_json::to_string_pretty(input.candidate).map_err(|e| e.to_string())?
    );
    if serde_json::from_str::<Value>(&bytes).is_err() {
        let mut report = CheckReport::default();
        report.refuse(
            "the produced settings text does not parse back; nothing is written".to_string(),
        );
        return Ok(CheckOutcome::Refused { report });
    }

    // REQ-1: a byte-exact backup first — 0600 in a 0700 folder; a file that does not exist
    // yet is recorded as absent rather than skipped.
    std::fs::create_dir_all(input.backup_dir).map_err(|e| format!("backup folder: {e}"))?;
    set_mode(input.backup_dir, 0o700);
    let stamp = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map(|d| d.as_millis())
        .unwrap_or_default();
    let existing = std::fs::read(input.settings_path).ok();
    let (backup, backup_bytes) = match &existing {
        Some(bytes) => (
            input.backup_dir.join(format!("settings-{stamp}.json")),
            bytes.clone(),
        ),
        None => (
            input.backup_dir.join(format!("settings-absent-{stamp}")),
            Vec::new(),
        ),
    };
    std::fs::write(&backup, &backup_bytes).map_err(|e| format!("backup write: {e}"))?;
    set_mode(&backup, 0o600);

    // The atomic swap through the shared writer: no swap, no change.
    crate::agents::write_replace(input.settings_path, existing.as_deref(), bytes.as_bytes())?;
    Ok(CheckOutcome::Written { backup })
}

#[cfg(unix)]
fn set_mode(path: &Path, mode: u32) {
    use std::os::unix::fs::PermissionsExt;
    let _ = std::fs::set_permissions(path, std::fs::Permissions::from_mode(mode));
}

#[cfg(not(unix))]
fn set_mode(_path: &Path, _mode: u32) {}
