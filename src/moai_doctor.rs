//! `moai doctor` (SPEC-20261127-doctor-moai-layer): a read-only diagnosis of the moai layer of a
//! Claude Code setup. Each check has its own `id`, so a later slice adds checks without changing
//! the others. The checks are pure functions over what [`gather`] read (variable names, files,
//! the `claude --version` probe, an optional `claude auth status` run), so each is testable
//! without a process.
//!
//! Secrets: a check prints variable names and states only, never a value, a key prefix, a
//! length, the output of `apiKeyHelper` or the contents of `.credentials.json` (REQ-14).
//! Writes: none. The only processes are the `claude --version` probe and, when asked, one
//! `claude auth status` (REQ-2, REQ-7).
//!
//! The minimum-version table lists the rows the research of the SPEC read as `verified` in
//! Claude Code's changelog (https://code.claude.com/docs/en/changelog.md, read 2026-11-27); a
//! row it could not verify is not listed (D-5).

mod capabilities;
mod locks;

use crate::claude::{claude_version, VersionProbe};
use crate::host_home::{self, Host};
use serde_json::{json, Map, Value};
use std::io::Read;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use std::sync::mpsc;
use std::time::{Duration, Instant};

/// The names the login check reports, in the order of the SPEC (REQ-6).
pub const LOGIN_NAMES: [&str; 7] = [
    "ANTHROPIC_API_KEY",
    "ANTHROPIC_AUTH_TOKEN",
    "CLAUDE_CODE_OAUTH_TOKEN",
    "ANTHROPIC_PROFILE",
    "CLAUDE_CODE_USE_BEDROCK",
    "CLAUDE_CODE_USE_VERTEX",
    "CLAUDE_CODE_USE_FOUNDRY",
];
const CLOUD_NAMES: [&str; 3] = [
    "CLAUDE_CODE_USE_BEDROCK",
    "CLAUDE_CODE_USE_VERTEX",
    "CLAUDE_CODE_USE_FOUNDRY",
];

/// How long `claude auth status` may run (REQ-7).
const AUTH_TIMEOUT: Duration = Duration::from_secs(10);
const AUTH_OUTPUT_LIMIT: usize = 64 * 1024;
const SETTINGS_LIMIT: u64 = 4 * 1024 * 1024;

type Version = (u32, u32, u32);

/// One row of the minimum-version table.
pub struct Row {
    pub feature: &'static str,
    pub min: Version,
    /// `Some("windows")` for a row that only applies on that platform.
    pub platform: Option<&'static str>,
    /// Where the minimum was read (page of https://code.claude.com/docs/en/).
    pub source: &'static str,
}

const CHANGELOG: &str = "changelog.md";

/// The `verified` rows of the SPEC's evidence table (read 2026-11-27, Claude Code 2.1.332).
pub const ROWS: &[Row] = &[
    row("worktree.baseRef", (2, 1, 133), None, CHANGELOG),
    row(
        "SessionStart hook sessionTitle",
        (2, 1, 152),
        None,
        CHANGELOG,
    ),
    row("agent view (claude agents)", (2, 1, 139), None, CHANGELOG),
    row("notify_when_idle", (2, 1, 236), None, CHANGELOG),
    row("hook onFailure", (2, 1, 295), None, CHANGELOG),
    row("--permission-prompts none", (2, 1, 259), None, CHANGELOG),
    row(
        "permissions.denyInsteadOfPrompting",
        (2, 1, 319),
        None,
        CHANGELOG,
    ),
    row("mods", (2, 1, 287), None, CHANGELOG),
    row("plugin engines", (2, 1, 323), None, CHANGELOG),
    row(
        "cross-session messaging (macOS, Linux, WSL)",
        (2, 1, 224),
        None,
        "cross-session-messaging.md",
    ),
    // D-5: the availability page says 2.1.234, the changelog line 2.1.239; the higher one is used.
    row(
        "cross-session messaging (native Windows)",
        (2, 1, 239),
        Some("windows"),
        CHANGELOG,
    ),
    row("crossSessionInbound", (2, 1, 224), None, CHANGELOG),
    row("agents file direct reading", (2, 1, 277), None, CHANGELOG),
    row("claude auth status", (2, 1, 41), None, CHANGELOG),
    row(
        "claude auth status configDirectory",
        (2, 1, 268),
        None,
        CHANGELOG,
    ),
    row(
        "claude auth status loginExpiresAt",
        (2, 1, 318),
        None,
        CHANGELOG,
    ),
    row(
        "strictPluginOnlyCustomization invalid value locks all",
        (2, 1, 282),
        None,
        "managed-settings.md",
    ),
];

const fn row(
    feature: &'static str,
    min: Version,
    platform: Option<&'static str>,
    source: &'static str,
) -> Row {
    Row {
        feature,
        min,
        platform,
        source,
    }
}

/// Claude Code's own checks, named for what `moai doctor` does not examine (REQ-13). `docs` is
/// `verified` only for a command the researched pages name.
const POINTERS: &[(&str, &str, &str)] = &[
    (
        "claude doctor",
        "installation and settings diagnostics, without starting a session",
        "verified",
    ),
    (
        "/status",
        "the session's account, model and settings sources",
        "verified",
    ),
    ("/context", "what fills the context window", "unverified"),
    (
        "/doctor prompt-audit",
        "the prompt files in use",
        "unverified",
    ),
    ("/skill-doctor", "skill loading problems", "unverified"),
    ("/usage", "plan usage and limits", "unverified"),
];

/// The five states of a check.
#[derive(Clone, Copy, PartialEq, Eq, Debug)]
pub enum State {
    Ok,
    Warn,
    Fail,
    Unknown,
    NotApplicable,
}

impl State {
    pub fn as_str(self) -> &'static str {
        match self {
            State::Ok => "ok",
            State::Warn => "warn",
            State::Fail => "fail",
            State::Unknown => "unknown",
            State::NotApplicable => "not_applicable",
        }
    }
}

/// One check of the report: every check has these fields, so one that cannot judge says
/// `unknown` with its reason instead of being absent (REQ-1).
pub struct Check {
    pub id: &'static str,
    pub state: State,
    pub summary: String,
    pub evidence: Value,
    pub next: String,
}

impl Check {
    fn to_json(&self) -> Value {
        json!({
            "id": self.id,
            "state": self.state.as_str(),
            "summary": self.summary,
            "evidence": self.evidence,
            "next": self.next,
        })
    }
}

/// What was read of a settings file: its absence, an unreadable or invalid state, or its JSON.
pub enum Settings {
    Absent,
    Unreadable,
    Json(Value),
}

/// The scope a settings file belongs to.
#[derive(Clone, Copy)]
pub enum Scope {
    User,
    Project,
    Local,
}

impl Scope {
    fn label(self) -> &'static str {
        match self {
            Scope::User => "user settings",
            Scope::Project => "project settings",
            Scope::Local => "local settings",
        }
    }
}

/// The result of the `claude --version` probe.
pub enum VersionResult {
    Found(Version),
    NotFound,
    Timeout,
    Unreadable,
}

/// The result of an opt-in `claude auth status` run: the three kept fields and the exit code,
/// or why it could not run.
pub enum AuthResult {
    Ran {
        exit: i32,
        auth_method: Option<String>,
        config_directory: Option<String>,
        login_expires_at: Option<String>,
    },
    Unknown(&'static str),
}

/// Everything the checks judge, read once by [`gather`].
pub struct Gathered {
    /// The Claude config folder, or the refusal naming the variable (never its value).
    pub folder: Result<PathBuf, String>,
    /// Names of the seven login variables that are set (non-empty), from [`LOGIN_NAMES`].
    pub set_names: Vec<&'static str>,
    pub credentials_file: bool,
    pub settings: Vec<(Scope, Settings)>,
    /// The managed settings files that exist, as `(label, content)`: `managed-settings.json`
    /// first, then `managed-settings.d/*.json` in alphabetical order (a later file wins).
    pub managed: Vec<(String, Settings)>,
    /// The host column of the capability table the process runs under (`cli` or `gateway`).
    pub host: &'static str,
    /// Whether the whole capability table is printed (`--table`).
    pub table: bool,
    pub version: VersionResult,
    /// `None` unless `--auth-status` was given.
    pub auth: Option<AuthResult>,
}

/// Options of the command.
pub struct Options {
    pub project: PathBuf,
    pub auth_status: bool,
    /// The managed settings folder to read; `None` reads the platform's own folder.
    pub managed_dir: Option<PathBuf>,
    pub table: bool,
    /// The home directory, resolved once at the command edge (`src/ops/doctor.rs`) so `gather`
    /// reads no `HOME` itself; the error text is shown as the `config_folder` problem.
    pub home: Result<PathBuf, String>,
}

/// The folder of the managed settings files of this platform
/// (https://code.claude.com/docs/en/managed-settings.md, read 2026-11-27).
fn platform_managed_dir() -> PathBuf {
    if cfg!(target_os = "macos") {
        PathBuf::from("/Library/Application Support/ClaudeCode")
    } else if cfg!(windows) {
        PathBuf::from("C:\\Program Files\\ClaudeCode")
    } else {
        PathBuf::from("/etc/claude-code")
    }
}

/// `managed-settings.json` and the `managed-settings.d/*.json` files of `dir` that exist, in the
/// order Claude Code merges them.
fn read_managed(dir: &Path) -> Vec<(String, Settings)> {
    let mut files = Vec::new();
    let base = dir.join(locks::MANAGED_BASE);
    if base.exists() {
        files.push((locks::MANAGED_BASE.to_string(), read_settings(&base)));
    }
    let mut names: Vec<String> = std::fs::read_dir(dir.join("managed-settings.d"))
        .map(|entries| {
            entries
                .filter_map(Result::ok)
                .filter_map(|e| e.file_name().into_string().ok())
                .filter(|name| name.ends_with(".json"))
                .collect()
        })
        .unwrap_or_default();
    names.sort();
    for name in names {
        let path = dir.join("managed-settings.d").join(&name);
        if path.is_file() {
            files.push((format!("managed-settings.d/{name}"), read_settings(&path)));
        }
    }
    files
}

/// Reads the environment, the files and the probe (the only side effects are the reads and the
/// bounded `claude` runs).
pub fn gather(opts: &Options) -> Gathered {
    let get = |k: &str| std::env::var_os(k).filter(|v| !v.is_empty());
    let set_names = LOGIN_NAMES
        .iter()
        .copied()
        .filter(|n| get(n).is_some())
        .collect::<Vec<_>>();

    let folder = opts
        .home
        .clone()
        .and_then(|home| host_home::resolve(Host::Claude, &home))
        .and_then(|dir| match std::fs::symlink_metadata(&dir) {
            Ok(m) if m.file_type().is_symlink() => Err(
                "the Claude config folder is a link; moai doctor does not follow it".to_string(),
            ),
            _ => Ok(dir),
        });

    let mut settings = Vec::new();
    let mut credentials_file = false;
    if let Ok(dir) = &folder {
        credentials_file = dir.join(".credentials.json").is_file();
        settings.push((Scope::User, read_settings(&dir.join("settings.json"))));
    }
    let project_claude = opts.project.join(".claude");
    settings.push((
        Scope::Project,
        read_settings(&project_claude.join("settings.json")),
    ));
    settings.push((
        Scope::Local,
        read_settings(&project_claude.join("settings.local.json")),
    ));

    let version = match claude_version() {
        VersionProbe::Found(v) => VersionResult::Found(v),
        VersionProbe::TimedOut { .. } => VersionResult::Timeout,
        VersionProbe::NoVersion => {
            if on_path("claude") {
                VersionResult::Unreadable
            } else {
                VersionResult::NotFound
            }
        }
    };
    let auth = opts.auth_status.then(auth_status);
    let managed = read_managed(
        opts.managed_dir
            .as_deref()
            .unwrap_or(&platform_managed_dir()),
    );
    // Only the host of the base URL decides the column; the URL itself is never kept.
    let base_url = std::env::var("ANTHROPIC_BASE_URL").ok();
    Gathered {
        folder,
        set_names,
        credentials_file,
        settings,
        managed,
        host: capabilities::host_of(base_url.as_deref()),
        table: opts.table,
        version,
        auth,
    }
}

fn read_settings(path: &Path) -> Settings {
    if !path.exists() {
        return Settings::Absent;
    }
    let mut bytes = Vec::new();
    let read =
        std::fs::File::open(path).and_then(|f| f.take(SETTINGS_LIMIT + 1).read_to_end(&mut bytes));
    if read.is_err() || bytes.len() as u64 > SETTINGS_LIMIT {
        return Settings::Unreadable;
    }
    match serde_json::from_slice::<Value>(&bytes) {
        Ok(v) if v.is_object() => Settings::Json(v),
        _ => Settings::Unreadable,
    }
}

/// Whether a file named `name` is in a `PATH` folder.
fn on_path(name: &str) -> bool {
    std::env::var_os("PATH")
        .map(|p| std::env::split_paths(&p).any(|d| d.join(name).is_file()))
        .unwrap_or(false)
}

/// `claude auth status` once, bounded by [`AUTH_TIMEOUT`]; only three fields and the exit code
/// are kept (REQ-7).
fn auth_status() -> AuthResult {
    let mut cmd = Command::new("claude");
    cmd.args(["auth", "status"])
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::null());
    #[cfg(unix)]
    std::os::unix::process::CommandExt::process_group(&mut cmd, 0);
    let mut child = match cmd.spawn() {
        Ok(c) => c,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => {
            return AuthResult::Unknown("not_found")
        }
        Err(_) => return AuthResult::Unknown("not_started"),
    };
    let stdout = child.stdout.take();
    let (tx, rx) = mpsc::channel();
    std::thread::spawn(move || {
        let mut buf = Vec::new();
        if let Some(pipe) = stdout {
            let _ = pipe
                .take(AUTH_OUTPUT_LIMIT as u64 + 1)
                .read_to_end(&mut buf);
        }
        let _ = tx.send(buf);
    });
    let deadline = Instant::now() + AUTH_TIMEOUT;
    let out = match rx.recv_timeout(deadline.saturating_duration_since(Instant::now())) {
        Ok(out) => out,
        Err(_) => {
            // The group is signalled before the leader is reaped, so a descendant that holds the
            // pipe or keeps running does not outlive the time-out (note 10861 of C-338).
            crate::claude::stop_process_group(&mut child);
            return AuthResult::Unknown("timeout");
        }
    };
    let Ok(status) = child.wait() else {
        return AuthResult::Unknown("not_started");
    };
    let Some(exit) = status.code() else {
        return AuthResult::Unknown("signal");
    };
    let parsed: Option<Value> = (out.len() <= AUTH_OUTPUT_LIMIT)
        .then(|| serde_json::from_slice(&out).ok())
        .flatten();
    let field = |name: &str| {
        parsed
            .as_ref()
            .and_then(|v| v.get(name))
            .and_then(Value::as_str)
            .map(str::to_string)
    };
    AuthResult::Ran {
        exit,
        auth_method: field("authMethod"),
        config_directory: field("configDirectory"),
        login_expires_at: field("loginExpiresAt"),
    }
}

/// The whole report.
pub fn report(g: &Gathered, platform: &str) -> Value {
    let checks = [
        config_folder_check(g),
        version_check(&g.version, platform),
        login_check(g),
        api_key_check(g),
        locks::managed_locks_check(g),
        locks::cross_session_inbound_check(g),
        capabilities::host_check(g.host),
    ];
    let mut report = json!({
        "writes": false,
        "checks": checks.iter().map(Check::to_json).collect::<Vec<_>>(),
        "pointers": pointers(),
    });
    if g.table {
        report["capability_table"] = capabilities::table();
    }
    report
}

fn config_folder_check(g: &Gathered) -> Check {
    match &g.folder {
        Ok(_) => Check {
            id: "config_folder",
            state: State::Ok,
            summary: "the Claude config folder resolves".into(),
            evidence: json!({ "variable": Host::Claude.variable(), "default": Host::Claude.default_dir() }),
            next: String::new(),
        },
        Err(reason) => Check {
            id: "config_folder",
            state: State::Fail,
            summary: format!("the Claude config folder cannot be resolved: {reason}"),
            evidence: json!({ "variable": Host::Claude.variable() }),
            next: "set CLAUDE_CONFIG_DIR to an absolute folder or unset it; checks that read the folder are unknown until then".into(),
        },
    }
}

fn dotted(v: Version) -> String {
    format!("{}.{}.{}", v.0, v.1, v.2)
}

/// The version check (REQ-4, REQ-5): each applicable row is `ok` at or above its minimum and
/// `below` otherwise; a row of another platform is `not_applicable`.
pub fn version_check(result: &VersionResult, platform: &str) -> Check {
    let unknown = |reason: &str, summary: &str, next: &str| Check {
        id: "claude_version",
        state: State::Unknown,
        summary: summary.into(),
        evidence: json!({ "reason": reason, "rows": [] }),
        next: next.into(),
    };
    let installed = match result {
        VersionResult::Found(v) => *v,
        VersionResult::NotFound => {
            return unknown(
                "not_found",
                "claude was not found on PATH, so no minimum version was judged",
                "install Claude Code or put it on PATH",
            )
        }
        VersionResult::Timeout => {
            return unknown(
                "timeout",
                "claude --version did not answer within 5 seconds",
                "run claude --version yourself",
            )
        }
        VersionResult::Unreadable => {
            return unknown(
                "unreadable",
                "claude --version printed no leading major.minor.patch",
                "run claude --version yourself",
            )
        }
    };
    let rows: Vec<Value> = ROWS
        .iter()
        .map(|r| {
            let applies = r.platform.is_none_or(|p| p == platform);
            let state = if !applies {
                "not_applicable"
            } else if installed >= r.min {
                "ok"
            } else {
                "below"
            };
            json!({
                "feature": r.feature,
                "min": dotted(r.min),
                "state": state,
                "source": r.source,
            })
        })
        .collect();
    let below: Vec<&str> = rows
        .iter()
        .filter(|r| r["state"] == "below")
        .filter_map(|r| r["feature"].as_str())
        .collect();
    let (state, summary, next) = if below.is_empty() {
        (
            State::Ok,
            format!(
                "Claude Code {} meets every listed minimum",
                dotted(installed)
            ),
            String::new(),
        )
    } else {
        (
            State::Warn,
            format!(
                "Claude Code {} is below the minimum of {} feature(s): {}",
                dotted(installed),
                below.len(),
                below.join(", ")
            ),
            "update Claude Code (claude update) before relying on those features".into(),
        )
    };
    Check {
        id: "claude_version",
        state,
        summary,
        evidence: json!({ "installed": dotted(installed), "rows": rows }),
        next,
    }
}

fn user_or_project_helper(g: &Gathered) -> Vec<&'static str> {
    g.settings
        .iter()
        .filter_map(|(scope, s)| match s {
            Settings::Json(v)
                if v.get("apiKeyHelper")
                    .and_then(Value::as_str)
                    .is_some_and(|h| !h.is_empty()) =>
            {
                Some(scope.label())
            }
            _ => None,
        })
        .collect()
}

/// The login source (REQ-6, REQ-7) from names only.
fn login_check(g: &Gathered) -> Check {
    let is_set = |n: &str| g.set_names.contains(&n);
    let mut variables = Map::new();
    for name in LOGIN_NAMES {
        variables.insert(
            name.to_string(),
            json!(if is_set(name) { "set" } else { "unset" }),
        );
    }
    let helper = !user_or_project_helper(g).is_empty();
    // The documented precedence (authentication.md, "Authentication precedence").
    let would_pick = if CLOUD_NAMES.iter().any(|n| is_set(n)) {
        "cloud_provider"
    } else if is_set("ANTHROPIC_AUTH_TOKEN") {
        "ANTHROPIC_AUTH_TOKEN"
    } else if is_set("ANTHROPIC_API_KEY") {
        "ANTHROPIC_API_KEY"
    } else if helper {
        "api_key_helper"
    } else if is_set("CLAUDE_CODE_OAUTH_TOKEN") {
        "CLAUDE_CODE_OAUTH_TOKEN"
    } else if is_set("ANTHROPIC_PROFILE") {
        "ANTHROPIC_PROFILE"
    } else if g.credentials_file {
        "login_file"
    } else {
        "none_visible"
    };
    let mut evidence = json!({
        "variables": variables,
        "credentials_file": if g.credentials_file { "present" } else { "absent" },
        "would_pick": would_pick,
        "expiry": "unknown",
    });
    let mut next = String::new();
    match &g.auth {
        None => {
            next = "run `claude auth status --text` yourself, or `moai doctor --auth-status`, to see the login and its expiry".into();
        }
        Some(AuthResult::Unknown(reason)) => {
            evidence["auth_status"] = json!({ "state": "unknown", "reason": reason });
        }
        Some(AuthResult::Ran {
            exit,
            auth_method,
            config_directory,
            login_expires_at,
        }) => {
            let mut auth = json!({ "exit": exit });
            if let Some(v) = auth_method {
                auth["authMethod"] = json!(v);
            }
            if let Some(v) = config_directory {
                auth["configDirectory"] = json!(v);
            }
            if let Some(v) = login_expires_at {
                auth["loginExpiresAt"] = json!(v);
                evidence["expiry"] = json!(v);
            }
            evidence["auth_status"] = auth;
        }
    }
    let (state, summary) = if g.folder.is_err() {
        (
            State::Unknown,
            "the login file could not be looked for: the Claude config folder is unresolved"
                .to_string(),
        )
    } else if would_pick == "none_visible" {
        (
            State::Unknown,
            "no login source is visible from names and files (on macOS the login may be in the Keychain)".to_string(),
        )
    } else {
        (
            State::Ok,
            format!("the login Claude Code would use first is {would_pick}"),
        )
    };
    Check {
        id: "login",
        state,
        summary,
        evidence,
        next,
    }
}

/// The API key warning (REQ-8): names only.
fn api_key_check(g: &Gathered) -> Check {
    let mut present: Vec<String> = ["ANTHROPIC_API_KEY", "ANTHROPIC_AUTH_TOKEN"]
        .iter()
        .filter(|n| g.set_names.contains(n))
        .map(|n| n.to_string())
        .collect();
    present.extend(
        user_or_project_helper(g)
            .into_iter()
            .map(|label| format!("apiKeyHelper ({label})")),
    );
    let unreadable: Vec<&str> = g
        .settings
        .iter()
        .filter(|(_, s)| matches!(s, Settings::Unreadable))
        .map(|(scope, _)| scope.label())
        .collect();
    let evidence = json!({ "present": present, "unreadable_settings": unreadable });
    if present.is_empty() {
        return Check {
            id: "api_key",
            state: State::Ok,
            summary: "no API key variable or apiKeyHelper entry is present".into(),
            evidence,
            next: String::new(),
        };
    }
    Check {
        id: "api_key",
        state: State::Warn,
        summary: format!(
            "{} present: the key is used instead of a subscription login (always in -p, after one approval otherwise); Remote Control, /schedule and claude.ai connectors are off while it is set",
            present.join(", ")
        ),
        evidence,
        next: "unset the variable or remove the apiKeyHelper entry to use the subscription login".into(),
    }
}

fn pointers() -> Vec<Value> {
    POINTERS
        .iter()
        .map(
            |(command, covers, docs)| json!({ "command": command, "covers": covers, "docs": docs }),
        )
        .collect()
}

/// `moai doctor`: gather, then judge.
pub fn run(opts: &Options) -> Value {
    let platform = if cfg!(windows) { "windows" } else { "unix" };
    report(&gather(opts), platform)
}
