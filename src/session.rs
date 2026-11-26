//! Development launcher: starts a leader or agent session on a host CLI with the
//! model/effort options the operator decided (2026-11-08). The advisor is off by default and
//! is turned on by `--advisor` or the user setting `launcher.advisor` (user decision
//! 2026-11-12, #2476). The launcher composes argv from documented flags only and never
//! invents options.
//!
//! Hosts (00-F #2028; sources read 2026-11-08):
//! - `claude`: Claude Code. `--name`, `--model`, `--effort`, `--dangerously-skip-permissions`
//!   from https://code.claude.com/docs/en/cli-reference; `--advisor <model>` from
//!   https://code.claude.com/docs/en/advisor (alias `fable`, `opus`, `sonnet` or a full id,
//!   must rank at or above the main model, disabled when `CLAUDE_CODE_DISABLE_ADVISOR_TOOL=1`).
//! - `glm`: Claude Code pointed at z.ai through process environment only, per
//!   https://docs.z.ai/scenario-example/develop-tools/claude. The token is inherited from the
//!   caller and only its presence is checked; its value is never kept, printed or written.
//!
//! The `moai-cli cc|glm` launchers open a plain host session (Issue #2053) unless a role flag is
//! given: no role is guessed and a session name comes from `--name` only. The role flags
//! `-l|--leader` and `-a|--agent` are back (user decision 2026-11-25, SPEC-20261125-agent-team-
//! session-names, reversing D-1 of SPEC-20261124-launcher-retire-role-flags): a role flag names
//! the session ([`leader_session_name`], [`pick_agent_name`]), starts it in the bypass unless
//! `--permissions` is given and hands the host the first input `/moai leader` or `/moai agent`
//! ([`SessionRequest::first_input`]). Without a role flag the session is the plain one and the
//! role is entered inside it with `/moai agent` or `/moai leader`. `moai cc` starts the host in auto
//! mode (`--permission-mode auto`, user decision 2026-11-23, SPEC-20261123-launcher-auto-mode)
//! and `moai glm` in the bypass (`--dangerously-skip-permissions`, user decision 2026-11-24,
//! SPEC-20261124-launcher-retire-role-flags D-7), unless `--permissions` is given (no
//! permission flag) or the user setting `launcher.role_access` says `auto`, `full` or `prompt`. Codex and OMP are not hosts
//! any more (user decision 2026-11-21, SPEC-20261121-codex-launcher-and-models-removed,
//! SPEC-20261122-remove-omp-launcher).
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::fs::{self, OpenOptions};
use std::path::{Path, PathBuf};

const ROLES: [&str; 2] = ["leader", "agent"];
const HOSTS: [&str; 2] = ["claude", "glm"];
const EFFORTS: [&str; 6] = ["low", "medium", "high", "xhigh", "max", "ultracode"];
const ADVISOR_ALIASES: [&str; 3] = ["fable", "opus", "sonnet"];
/// Notes that tell the user an option they gave was dropped. The terminal launcher prints these
/// before the host starts (#3004 P9), so the text is shared with `launcher_terminal`.
pub const NOTE_ADVISOR_OMITTED_GLM: &str =
    "--advisor omitted on glm: the advisor is Anthropic API only";

/// The role words a launcher flag hands the host as its first input (SPEC-20261125-agent-team-
/// session-names REQ-7 and REQ-8). The text is typed into the session as the user would type it.
pub const FIRST_INPUT_LEADER: &str = "/moai leader";
pub const FIRST_INPUT_AGENT: &str = "/moai agent";

/// The Biblical names `-a` picks from when `--name` is not given: the session is called
/// `<project name>-agent-<name>` (SPEC-20261125-session-names-project-prefix D-1 for the shape,
/// decided 2026-11-25). The pick is random among the names that are free (see
/// [`pick_agent_name`]); SPEC-20261125-agent-names-bible replaces D-8 of SPEC-20261125-agent-team-
/// session-names, which listed 24 English names taken in order. Adding to the list is a card of
/// its own.
pub const AGENT_NAMES: [&str; 48] = [
    "adam",
    "eve",
    "noah",
    "abraham",
    "sarah",
    "isaac",
    "rebekah",
    "jacob",
    "rachel",
    "leah",
    "joseph",
    "moses",
    "aaron",
    "miriam",
    "joshua",
    "deborah",
    "gideon",
    "samson",
    "ruth",
    "naomi",
    "boaz",
    "hannah",
    "samuel",
    "david",
    "saul",
    "solomon",
    "elijah",
    "elisha",
    "isaiah",
    "jeremiah",
    "ezekiel",
    "daniel",
    "esther",
    "jonah",
    "mary",
    "martha",
    "lazarus",
    "peter",
    "andrew",
    "james",
    "john",
    "thomas",
    "matthew",
    "luke",
    "paul",
    "timothy",
    "lydia",
    "priscilla",
];

/// The project name that prefixes the session names `-l` and `-a` give (SPEC-20261125-session-
/// names-project-prefix D-2, D-3): the `project_id` of the project profile when it is a name a
/// person chose, else the folder name of the clone's primary checkout (never the folder of a
/// worktree, whose name changes with every card). A random-form id (`project_id::is_random_form`)
/// names no project to a person and is not used. The text is reduced to the registry's characters
/// (`A-Z a-z 0-9 . _ -`, a letter or digit first and last) and cut so that the longest session
/// name, `<project name>-agent-<longest english name>`, fits the registry's 64 characters. A
/// folder outside a git repository uses its own name; an empty result is `project`.
pub fn project_name(dir: &Path) -> String {
    let identity = crate::identity::identity_in(dir, None).ok();
    let root = identity
        .as_ref()
        .and_then(|id| id["work_tree"].as_str())
        .map(PathBuf::from)
        .unwrap_or_else(|| dir.to_path_buf());
    // The primary checkout is the parent of the clone's common `.git` folder; any other layout
    // (a bare clone, a separate git dir) keeps the folder of the work tree.
    let primary = identity
        .as_ref()
        .and_then(|id| id["clone_anchor"].as_str())
        .map(PathBuf::from)
        .filter(|common| common.file_name().is_some_and(|n| n == ".git"))
        .and_then(|common| common.parent().map(Path::to_path_buf))
        .unwrap_or_else(|| root.clone());
    let id = crate::project_id::read(&root)
        .ok()
        .flatten()
        .filter(|id| !crate::project_id::is_random_form(id));
    let raw = id.unwrap_or_else(|| {
        primary
            .file_name()
            .map(|n| n.to_string_lossy().into_owned())
            .unwrap_or_default()
    });
    let is_edge = |c: char| !c.is_ascii_alphanumeric();
    let mut name: String = raw
        .chars()
        .map(|c| {
            if c.is_ascii_alphanumeric() || matches!(c, '.' | '_' | '-') {
                c
            } else {
                '-'
            }
        })
        .collect();
    name = name.trim_start_matches(is_edge).to_string();
    name.truncate(max_project_len());
    name = name.trim_end_matches(is_edge).to_string();
    if name.is_empty() {
        "project".into()
    } else {
        name
    }
}

/// The longest project name for which `<project name>-agent-<english name>` still fits the
/// registry's name length, whichever name of the list is taken.
fn max_project_len() -> usize {
    let longest = AGENT_NAMES.iter().map(|n| n.len()).max().unwrap_or(0);
    crate::lane_registry::MAX_NAME_LEN - "-agent-".len() - longest
}

/// `<project name>-leader`, the session name `-l` gives the coordination window.
pub fn leader_session_name(dir: &Path) -> String {
    format!("{}-leader", project_name(dir))
}

/// `<project name>-agent-`, the start of every lane name of this project.
pub fn agent_prefix(dir: &Path) -> String {
    format!("{}-agent-", project_name(dir))
}

/// `<project name>-agent-<name>`: a name of [`AGENT_NAMES`] chosen at random, with equal chance,
/// among the names that are free in both of their forms, `agent-<name>` and
/// `<project name>-agent-<name>` (SPEC-20261125-session-registry-session-key REQ-8). A form is
/// taken while its `<form>.live` lock is held or its registry entry `<form>.json` has a running
/// lease (REQ-2, REQ-3); the answer always carries the project prefix. A folder outside a git
/// repository has no sessions folder to read, so every name is free. When every name is taken
/// the answer is one message that names `--name` (REQ-2). The seed comes from the clock, the
/// process id, `now` and a process-wide counter, so two calls in one nanosecond differ;
/// [`pick_agent_name_seeded`] takes it from the caller.
pub fn pick_agent_name(dir: &Path, now: i64) -> Result<String, String> {
    pick_agent_name_seeded(dir, now, agent_pick_seed(now))
}

/// [`pick_agent_name`] with the seed given: the free names are collected in list order and the
/// one at `seed % free count` is returned, so a test reaches every free name and no other.
pub fn pick_agent_name_seeded(dir: &Path, now: i64, seed: u64) -> Result<String, String> {
    let sessions = agent_seq_dir(dir).ok();
    let prefix = agent_prefix(dir);
    let mut free: Vec<String> = Vec::new();
    for n in AGENT_NAMES {
        let name = format!("{prefix}{n}");
        if let Some(sessions) = &sessions {
            let mut taken = false;
            for form in [format!("agent-{n}"), name.clone()] {
                if live_lock_held(&sessions.join(format!("{form}.live")))?
                    || crate::lane_registry::is_live(sessions, &form, now)?
                {
                    taken = true;
                    break;
                }
            }
            if taken {
                continue;
            }
        }
        free.push(name);
    }
    if free.is_empty() {
        return Err(format!(
            "every name of the launcher's list ({} names) is taken: name this session with `--name <name>`",
            AGENT_NAMES.len()
        ));
    }
    let at = (seed % free.len() as u64) as usize;
    Ok(free.swap_remove(at))
}

/// A seed for [`pick_agent_name_seeded`] from the standard library only: the clock nanoseconds,
/// the process id, `now` and a counter that grows with every call in this process, mixed by the
/// splitmix64 finalizer so that neighbouring inputs give unrelated seeds.
fn agent_pick_seed(now: i64) -> u64 {
    use std::sync::atomic::{AtomicU64, Ordering};
    use std::time::{SystemTime, UNIX_EPOCH};
    static COUNTER: AtomicU64 = AtomicU64::new(0);
    let nanos = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map(|d| d.as_nanos() as u64)
        .unwrap_or(0);
    let count = COUNTER.fetch_add(1, Ordering::Relaxed);
    let mut z = nanos
        ^ (u64::from(std::process::id()) << 32)
        ^ (now as u64).rotate_left(17)
        ^ count.wrapping_mul(0x9E37_79B9_7F4A_7C15);
    z = z.wrapping_add(0x9E37_79B9_7F4A_7C15);
    z = (z ^ (z >> 30)).wrapping_mul(0xBF58_476D_1CE4_E5B9);
    z = (z ^ (z >> 27)).wrapping_mul(0x94D0_49BB_1331_11EB);
    z ^ (z >> 31)
}

/// User decision 2026-11-08 (#2053 HOLD-L1 = A, #2062): a session without a role still gets the
/// launcher's defaults of model and effort. False would pass no model or effort flag so the
/// user's own Claude settings apply (option B, not chosen). The values are the host's own
/// (`model_effort_settings::launcher_default`: `opus`, effort `medium` on `claude` and `high`
/// on `glm`, SPEC-20261124-launcher-retire-role-flags D-8). The advisor is no longer a default
/// (user decision 2026-11-12, #2476, replacing the `--advisor fable` part of the 2026-11-08
/// decision): see [`launch_settings`].
pub const ROLELESS_LAUNCHER_DEFAULTS: bool = true;
/// User decision 2026-11-23 (SPEC-20261123-launcher-auto-mode, replacing the 2026-11-10 decision
/// of #2359): a launch on a Claude host starts in auto mode (`--permission-mode auto`, whatever
/// the project's `permissions` key says). The glm host starts in the bypass
/// (`--dangerously-skip-permissions`, user decision 2026-11-24, SPEC-20261124-launcher-retire-
/// role-flags REQ-13); see [`access_default`]. `--permissions` keeps the host's own mode (no
/// permission flag). The user-only `~/.moai-cli/config/launcher.yaml` key
/// `launcher.role_access` can choose `auto`, `full` (the bypass) or `prompt` for every launch
/// of both hosts; an absent file or key keeps the host default. Project settings are ignored.
pub const ACCESS_DEFAULT: Access = Access::Auto;

/// The access of a launch on `host` (`claude` or `glm`) when neither `--permissions` nor
/// `launcher.role_access` decides: auto mode on `claude`, the bypass on `glm`.
pub fn access_default(host: &str) -> Access {
    if host == "glm" {
        Access::Full
    } else {
        ACCESS_DEFAULT
    }
}
/// The environment that points Claude Code at z.ai GLM. The set follows the Go tool's
/// `setGLMEnv` (user decision 2026-11-10): endpoint, timeout, the model each Claude alias maps
/// to, the 1M context window of those models (Claude Code assumes 200K for an id it does not
/// know) and the switch that strips Anthropic beta headers the proxy rejects. The opus, sonnet
/// and fable aliases map to `glm-5.4[1m]` by the user decision of 2026-11-23
/// (SPEC-20261123-glm-54-model); haiku stays `glm-5.3-flash`, as no 5.4 flash model was named.
pub const GLM_ENV: [(&str, &str); 9] = [
    ("ANTHROPIC_BASE_URL", "https://api.z.ai/api/anthropic"),
    ("API_TIMEOUT_MS", "3000000"),
    ("ANTHROPIC_DEFAULT_OPUS_MODEL", "glm-5.4[1m]"),
    ("ANTHROPIC_DEFAULT_SONNET_MODEL", "glm-5.4[1m]"),
    ("ANTHROPIC_DEFAULT_HAIKU_MODEL", "glm-5.3-flash"),
    ("ANTHROPIC_DEFAULT_FABLE_MODEL", "glm-5.4[1m]"),
    ("CLAUDE_CODE_AUTO_COMPACT_WINDOW", "1000000"),
    ("CLAUDE_CODE_MAX_CONTEXT_TOKENS", "1000000"),
    ("CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS", "1"),
];
/// The z.ai MCP server reads the same key under this name; the launcher passes it along.
pub const GLM_MCP_KEY_VAR: &str = "Z_AI_API_KEY";
/// Taken from the caller's environment, else from a key file (#2359). The value is passed to
/// the host process only; it is never printed, logged or written.
pub const GLM_TOKEN_VAR: &str = "ANTHROPIC_AUTH_TOKEN";
/// Legacy key files under HOME (`~/.moai/.env.glm` is the store shared with the Go tool), one
/// list with `key-import-*`. The launcher never opens or reads them (user decision, Issue
/// #2393 comment 6778150969): only their presence is checked, to point at the move into the
/// vault (`legacy_key_only`). Only `key-import-plan`/`key-import-apply` read them, on request.
pub const GLM_KEY_FILES: [&str; 2] = crate::key_vault::LEGACY_GLM_FILES;

/// The first legacy key file present at all (any kind of entry). Only `symlink_metadata` is
/// used, so a FIFO, an unreadable file or a symlink is never opened.
fn legacy_key_file(home: &std::path::Path) -> Option<&'static str> {
    GLM_KEY_FILES
        .iter()
        .copied()
        .find(|f| std::fs::symlink_metadata(home.join(f)).is_ok())
}

/// The one rule for a token value from any source: surrounding whitespace is dropped, and an
/// empty, whitespace-only or control-character value counts as no token (#2359 AC4).
fn usable_token(value: &str) -> Option<String> {
    let value = value.trim();
    (!value.is_empty() && !value.chars().any(char::is_control)).then(|| value.to_string())
}

/// Where the GLM token came from, by name only (never the value).
#[derive(Clone, Debug, PartialEq, Eq)]
enum KeySource {
    Env,
    Vault(crate::key_vault::Scope),
}

impl KeySource {
    fn name(&self) -> &'static str {
        match self {
            KeySource::Env => "env:ANTHROPIC_AUTH_TOKEN",
            KeySource::Vault(crate::key_vault::Scope::Project(_)) => "vault:project:glm",
            KeySource::Vault(crate::key_vault::Scope::Shared) => "vault:shared:glm",
        }
    }
}

/// The token and its source, in this order (#2393 S2, leader decision on the Issue):
/// 1. the variable, when it holds a usable value (an empty or blank one does not hide the rest);
/// 2. the encrypted vault's `glm` entry: the current project's folder, then `_shared`
///    (decision 4 of Issue #2393 comment 6778671862) — an entry that is present but unsafe or
///    does not decrypt stops the launch with its code instead of falling through, so a broken
///    entry is not bypassed;
/// 3. with no vault entry, a legacy key file is not read (user decision, Issue #2393 comment
///    6778150969); its presence stops the launch with `legacy_key_only` and the way to move
///    the key, and the file is left as it is.
fn glm_token() -> Result<Option<(String, KeySource)>, crate::key_vault::Blocked> {
    let from_env = std::env::var(GLM_TOKEN_VAR)
        .ok()
        .and_then(|v| usable_token(&v));
    if let Some(token) = from_env {
        return Ok(Some((token, KeySource::Env)));
    }
    if let Ok(home) = crate::key_vault::home() {
        if let Some((token, scope)) = crate::key_vault::read_here(&home, "glm")? {
            return Ok(Some((token, KeySource::Vault(scope))));
        }
        if let Some(file) = legacy_key_file(&home) {
            return Err(crate::key_vault::Blocked::legacy_key_only(file));
        }
    }
    Ok(None)
}

/// The launcher's one-line refusal for a vault that cannot give the key: code, reason and the
/// next action, never the value.
fn blocked_line(b: &crate::key_vault::Blocked) -> String {
    format!("blocked: {}: {} (next: {})", b.code, b.reason, b.next)
}

/// How a launch asks the host for permissions (SPEC-20261123-launcher-auto-mode).
#[derive(Clone, Copy, Debug, Default, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "lowercase")]
pub enum Access {
    /// `--permission-mode auto`: the host's auto mode (the default of `cc`).
    Auto,
    /// `--dangerously-skip-permissions`: no permission prompt at all (the default of `glm`; on
    /// `cc` only the user's own choice).
    Full,
    /// No permission flag: the host starts in the mode its own settings choose.
    // Not ACCESS_DEFAULT: a serialized request without `access` adds no flag (the safe side).
    #[default]
    Prompt,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
pub struct SessionRequest {
    /// `leader` or `agent`; `None` is a plain host session (#2053).
    #[serde(default)]
    pub role: Option<String>,
    pub host: String,
    /// `None` adds no session name (no `--name`, no name prompt).
    #[serde(default)]
    pub name: Option<String>,
    /// `None` uses the host default: `opus` on Claude hosts.
    #[serde(default)]
    pub model: Option<String>,
    /// `None` uses the host default on Claude hosts: `medium` for `cc`, `high` for `glm`.
    #[serde(default)]
    pub effort: Option<String>,
    #[serde(default)]
    pub advisor: Option<String>,
    /// The permission flag of the Claude host. The launcher sets it from [`launch_settings`]; the
    /// legacy `session` command leaves it at `Prompt` (no flag).
    #[serde(default)]
    pub access: Access,
    /// Extra first-prompt text (no host takes it: the Claude host is told it was not passed).
    #[serde(default)]
    pub prompt: Option<String>,
    /// The first input a role flag hands the host: [`FIRST_INPUT_LEADER`] or [`FIRST_INPUT_AGENT`],
    /// the last argument of the host command (SPEC-20261125-agent-team-session-names REQ-7 and
    /// REQ-8). `None` hands the host nothing.
    #[serde(default)]
    pub first_input: Option<String>,
    #[serde(default = "default_claude_bin")]
    pub claude_bin: String,
}

fn default_claude_bin() -> String {
    "claude".into()
}

fn rank(model: &str) -> Option<u8> {
    let m = model.to_ascii_lowercase();
    ["haiku", "sonnet", "opus", "fable"]
        .iter()
        .position(|k| m.contains(k))
        .map(|p| p as u8)
}
/// The first input is one of the two role entries and nothing else: a request that carries any
/// other text (a serialized request of the legacy `session` command) is refused.
fn clean_input(input: &str) -> Result<(), String> {
    if input == FIRST_INPUT_LEADER || input == FIRST_INPUT_AGENT {
        Ok(())
    } else {
        Err("invalid first_input".into())
    }
}
fn clean(field: &str, value: &str) -> Result<(), String> {
    if value.is_empty() || value.len() > 128 || value.chars().any(|c| c.is_control() || c == ' ') {
        return Err(format!("invalid {field}"));
    }
    Ok(())
}

/// Whether the launcher's model/effort defaults apply: always with a role, and without one
/// while [`ROLELESS_LAUNCHER_DEFAULTS`] is true (user decision 2026-11-08: A).
pub fn launcher_defaults(role: Option<&str>) -> bool {
    role.is_some() || ROLELESS_LAUNCHER_DEFAULTS
}

/// Reads the `launcher` object of the user-only `~/.moai-cli/config/launcher.yaml`; `None` when
/// HOME, the file or the key is absent. Like the project profile's `.yaml` file, this file
/// accepts only the JSON subset, parsed by serde_json (not general YAML). Errors never quote
/// the file's content.
fn user_launcher() -> Result<Option<serde_json::Map<String, Value>>, String> {
    let Ok(home) = crate::key_vault::home() else {
        return Ok(None);
    };
    let path = crate::home_root::path(&home).join("config/launcher.yaml");
    // Check the entry before reading: directories, symlinks and special files (including
    // FIFOs, which could block a read) are not launcher configuration files.
    match fs::symlink_metadata(&path) {
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(None),
        Err(_) => return Err("user launcher configuration cannot be read".into()),
        Ok(metadata) if !metadata.file_type().is_file() => {
            return Err("user launcher configuration must be a regular file".into())
        }
        Ok(_) => {}
    }
    let bytes = fs::read(&path).map_err(|_| "user launcher configuration cannot be read")?;
    let value: Value = serde_json::from_slice(&bytes)
        .map_err(|_| "user launcher configuration must use valid JSON-subset syntax")?;
    let Value::Object(mut root) = value else {
        return Err("user launcher configuration must be an object".into());
    };
    match root.remove("launcher") {
        None => Ok(None),
        Some(Value::Object(launcher)) => Ok(Some(launcher)),
        Some(_) => Err("user launcher configuration launcher must be an object".into()),
    }
}

fn configured_access(
    host: &str,
    launcher: Option<&serde_json::Map<String, Value>>,
) -> Result<Access, String> {
    match launcher.and_then(|l| l.get("role_access")) {
        None => Ok(access_default(host)),
        Some(Value::String(access)) if access == "auto" => Ok(Access::Auto),
        Some(Value::String(access)) if access == "full" => Ok(Access::Full),
        Some(Value::String(access)) if access == "prompt" => Ok(Access::Prompt),
        Some(_) => Err(
            "user launcher configuration launcher.role_access must be auto, full or prompt".into(),
        ),
    }
}

/// `launcher.advisor`: absent or `none` is no advisor; `fable`, `opus`, `sonnet` or a full
/// `claude-*` id (the `--advisor` forms) turns it on. Anything else is refused without quoting it.
fn configured_advisor(
    launcher: Option<&serde_json::Map<String, Value>>,
) -> Result<Option<String>, String> {
    match launcher.and_then(|l| l.get("advisor")) {
        None => Ok(None),
        Some(Value::String(a)) if a == "none" => Ok(None),
        Some(Value::String(a))
            if ADVISOR_ALIASES.contains(&a.as_str())
                || (a.starts_with("claude-") && clean("advisor", a).is_ok()) =>
        {
            Ok(Some(a.clone()))
        }
        Some(_) => Err(
            "user launcher configuration launcher.advisor must be none, fable, opus, sonnet or a full claude-* model id"
                .into(),
        ),
    }
}

/// Where the advisor of a launch came from.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum AdvisorSource {
    Cli,
    UserSetting,
}

/// The launcher's resolution of the user-only settings for one launch.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct LaunchSettings {
    /// See [`launch_settings`].
    pub access: Access,
    /// The advisor and its source; `None` passes no `--advisor`.
    pub advisor: Option<(String, AdvisorSource)>,
}

/// Resolves both user launcher settings from one read of the file (Issues #2413, #2476).
/// Access: CLI `--permissions` (no flag) > `launcher.role_access` > the host default
/// ([`access_default`]: auto on `claude`, the bypass on `glm`). Advisor (user decision 2026-11-12,
/// #2476, replacing
/// the 2026-11-08 default `fable`): CLI `--no-advisor` / `--advisor` > `launcher.advisor` >
/// none; `--advisor` with `--no-advisor` is refused, and the setting applies to Claude hosts
/// only. The file is validated for every launch, also when the CLI
/// decides, so a broken file is never silently skipped. Project settings do not matter.
pub fn launch_settings(
    host: &str,
    keep_prompts: bool,
    cli_advisor: Option<&str>,
    no_advisor: bool,
) -> Result<LaunchSettings, String> {
    let launcher = user_launcher()?;
    let configured_access = configured_access(host, launcher.as_ref())?;
    let configured = configured_advisor(launcher.as_ref())?;
    if cli_advisor.is_some() && no_advisor {
        return Err("--advisor and --no-advisor cannot be used together".into());
    }
    let advisor = match (cli_advisor, no_advisor) {
        (Some(a), _) => Some((a.to_string(), AdvisorSource::Cli)),
        (None, true) => None,
        (None, false) => configured.map(|a| (a, AdvisorSource::UserSetting)),
    };
    Ok(LaunchSettings {
        access: if keep_prompts {
            Access::Prompt
        } else {
            configured_access
        },
        advisor,
    })
}

/// Resolves CLI `--permissions` > user `launcher.role_access` > the host default
/// ([`access_default`]). The configuration is validated even for `--permissions`. Project
/// settings do not matter.
pub fn launch_access(host: &str, keep_prompts: bool) -> Result<Access, String> {
    Ok(launch_settings(host, keep_prompts, None, false)?.access)
}

// ---- session live locks (the launcher no longer numbers sessions: SPEC-20261124-launcher-
// retire-role-flags D-4; a name comes from `--name` only) ----

/// Where the session registry and the `.live` locks are kept: `<clone common dir>/.moai-cli/
/// sessions/`, beside the claim journal (`claim::journal_root`). Every worktree of one clone
/// resolves the same common dir, so they share it across hosts, and git never lists the files.
/// A session is alive while its launcher holds the lock on `<name>.live` or while the registry
/// entry `<name>.json` has a running lease.
pub fn agent_seq_dir(project: &Path) -> Result<PathBuf, String> {
    let id =
        crate::identity::identity_in(project, None).map_err(|e| format!("session folder: {e}"))?;
    if id["identity_status"].as_str() != Some("ok") {
        return Err(format!(
            "the session folder needs a git repository (identity_status {}); run from the project",
            id["identity_status"].as_str().unwrap_or("unknown")
        ));
    }
    let common = id["clone_anchor"]
        .as_str()
        .ok_or("the session folder needs a git repository")?;
    Ok(crate::clone_root::path(Path::new(common)).join("sessions"))
}

fn refuse_symlink(path: &Path) -> Result<(), String> {
    match fs::symlink_metadata(path) {
        Ok(m) if m.file_type().is_symlink() => Err(format!(
            "session lock path {} is a symlink; refused",
            path.display()
        )),
        _ => Ok(()),
    }
}

/// A path that is absent or a regular file. A symlink is refused, and so is anything else (a
/// FIFO would block an open), before the path is opened.
fn refuse_unless_regular(path: &Path) -> Result<(), String> {
    refuse_symlink(path)?;
    match fs::symlink_metadata(path) {
        Ok(m) if !m.file_type().is_file() => Err(format!(
            "session lock path {} is not a regular file; refused",
            path.display()
        )),
        _ => Ok(()),
    }
}

/// `n` of a canonical `agent-<n><suffix>` file name: decimal digits without a leading zero
/// that fit in 32 bits. Any other name is not an `agent-<n>` entry of the session registry.
pub(crate) fn agent_number_of(file_name: &str, suffix: &str) -> Option<u64> {
    let digits = file_name.strip_prefix("agent-")?.strip_suffix(suffix)?;
    let n: u32 = digits.parse().ok()?;
    (n.to_string() == digits).then_some(u64::from(n))
}

/// Whether some launcher holds the lock on the `agent-<n>.live` file at `path`. The probe asks
/// for a shared lock and lets go at once: it conflicts with the holder's exclusive lock only,
/// and two probes never block each other.
pub(crate) fn live_lock_held(path: &Path) -> Result<bool, String> {
    refuse_unless_regular(path)?;
    let unusable = || format!("session lock path {} cannot be probed", path.display());
    let file = match OpenOptions::new().read(true).open(path) {
        Ok(f) => f,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(false),
        Err(_) => return Err(unusable()),
    };
    match file.try_lock_shared() {
        Ok(()) => Ok(false),
        Err(fs::TryLockError::WouldBlock) => Ok(true),
        Err(fs::TryLockError::Error(_)) => Err(unusable()),
    }
}

/// Takes the `<name>.live` lock for a named session (BOARD 03 (가),
/// SPEC-20261120-launcher-liveness REQ-4): the launcher holds this exclusive lock for as long
/// as the host runs, under the name `--name` gave, so a read-only probe ([`live_lock_held`])
/// can tell a live launcher from an exited one. A symlink or a non-regular file at the path is
/// refused; the name must be a safe registry name.
pub fn lock_named_live(dir: &Path, name: &str) -> Result<fs::File, String> {
    if !crate::lane_registry::valid_name(name) {
        return Err(
            "session live lock name is invalid: use 1 to 64 of A-Z a-z 0-9 . _ -".to_string(),
        );
    }
    // The first named launch in a clone makes the sessions folder.
    fs::create_dir_all(dir)
        .map_err(|e| format!("session live lock {} write failed: {e}", dir.display()))?;
    let path = dir.join(format!("{name}.live"));
    refuse_unless_regular(&path)?;
    let unavailable = || format!("session live lock {} unavailable", path.display());
    let file = OpenOptions::new()
        .read(true)
        .write(true)
        .create(true)
        .truncate(false)
        .open(&path)
        .map_err(|_| unavailable())?;
    // N2 of the audit: a session of this name is already running — say so at once, with no
    // wait; a named lock's only other holder is a full session.
    match file.try_lock() {
        Ok(()) => Ok(file),
        Err(fs::TryLockError::WouldBlock) => Err(format!(
            "{name} is already running: the session live lock {} is held; stop that session first",
            path.display()
        )),
        Err(fs::TryLockError::Error(_)) => Err(unavailable()),
    }
}

fn claude_argv(r: &SessionRequest, notes: &mut Vec<&'static str>) -> Result<Vec<String>, String> {
    let defaults = launcher_defaults(r.role.as_deref());
    let default = crate::model_effort_settings::launcher_default(&r.host);
    let model = r
        .model
        .as_deref()
        .or(defaults.then_some(default.model.as_str()));
    let effort = r
        .effort
        .as_deref()
        .or(defaults.then_some(default.effort.as_str()));
    if let Some(e) = effort {
        if !EFFORTS.contains(&e) {
            return Err("effort must be one of low|medium|high|xhigh|max|ultracode".into());
        }
    }
    if let Some(m) = model {
        clean("model", m)?;
    }
    clean("claude_bin", &r.claude_bin)?;
    let mut argv = vec![r.claude_bin.clone()];
    if let Some(n) = &r.name {
        argv.push("--name".into());
        argv.push(n.clone());
    }
    if let Some(m) = model {
        argv.push("--model".into());
        argv.push(m.into());
    }
    if let Some(e) = effort {
        argv.push("--effort".into());
        argv.push(e.into());
    }
    if let Some(a) = &r.advisor {
        clean("advisor", a)?;
        if !ADVISOR_ALIASES.contains(&a.as_str()) && !a.starts_with("claude-") {
            return Err("advisor must be fable|opus|sonnet or a full claude-* model id".into());
        }
        match (model.and_then(rank), rank(a)) {
            (Some(m), Some(x)) if x < m => {
                return Err("advisor must rank at or above the main model".into())
            }
            (None, _) | (_, None) => notes.push("advisor/model ranking not checked: unknown id"),
            _ => {}
        }
        if std::env::var("CLAUDE_CODE_DISABLE_ADVISOR_TOOL")
            .ok()
            .as_deref()
            == Some("1")
        {
            notes.push("CLAUDE_CODE_DISABLE_ADVISOR_TOOL=1 is set: the host ignores --advisor");
        }
        notes.push("--advisor is documented but hidden from `claude --help`; experimental, Anthropic API only");
        notes.push("--advisor may bill to usage credits; without them set up the host can exit at start (#2476)");
        argv.push("--advisor".into());
        argv.push(a.clone());
    }
    if r.prompt.is_some() {
        notes.push("PROMPT was not passed to the Claude host");
    }
    match r.access {
        Access::Auto => {
            notes.push("--permission-mode auto: Claude Code's auto mode checks each tool call (the default of `moai cc` since user decision 2026-11-23; --permissions keeps the host's own mode, launcher.role_access sets the mode of every launch: auto, full or prompt)");
            if r.host == "glm" {
                notes.push("glm runs Claude Code behind the z.ai gateway: the auto mode classifier may show a billing notice, see docs/troubleshooting.md section 5 (no setting is changed)");
            }
            argv.push("--permission-mode".into());
            argv.push("auto".into());
        }
        Access::Full => {
            notes.push("--dangerously-skip-permissions: every tool call runs without a permission prompt (the default of `moai glm` since user decision 2026-11-24, or launcher.role_access full for every launch; keep prompts with --permissions, pick auto with launcher.role_access auto)");
            argv.push("--dangerously-skip-permissions".into());
        }
        // A role session without a flag (`--permissions`, `prompt`, the legacy `session`
        // command) added no note before; only the roleless form did, and keeps doing so.
        Access::Prompt if r.role.is_none() => notes.push("no permission flag: Claude Code starts in the mode the user's and project's settings choose (--permissions, or launcher.role_access prompt for every launch)"),
        Access::Prompt => {}
    }
    // The first input is the last argument, after every option (REQ-7, REQ-8).
    if let Some(input) = &r.first_input {
        clean_input(input)?;
        argv.push(input.clone());
    }
    Ok(argv)
}

/// The note that says a value came from the choice `moai update` saved for the host and role
/// (SPEC-20261119-launcher-reads-models REQ-9): the key path and nothing else of the file.
fn saved_note(host: &str, role: &str) -> &'static str {
    match (host, role) {
        ("glm", "leader") => "model and effort not given on the command line come from launcher.models.glm.leader (saved by moai update)",
        ("glm", _) => "model and effort not given on the command line come from launcher.models.glm.agent (saved by moai update)",
        (_, "leader") => "model and effort not given on the command line come from launcher.models.claude.leader (saved by moai update)",
        _ => "model and effort not given on the command line come from launcher.models.claude.agent (saved by moai update)",
    }
}

/// The notes of a launch that read `launcher.models` for `host`: a flat entry says so when it
/// filled a value, and an old `leader` or `agent` entry says it is not read (one note each, key
/// paths only; SPEC-20261124-launcher-retire-role-flags REQ-4 and REQ-5).
fn flat_note(host: &str) -> &'static str {
    if host == "glm" {
        "model and effort not given on the command line come from launcher.models.glm (saved by moai update)"
    } else {
        "model and effort not given on the command line come from launcher.models.claude (saved by moai update)"
    }
}

fn old_role_note(host: &str) -> &'static str {
    if host == "glm" {
        "launcher.models.glm.leader and launcher.models.glm.agent are not read: a launch takes the flat launcher.models.glm entry; save one with moai update"
    } else {
        "launcher.models.claude.leader and launcher.models.claude.agent are not read: a launch takes the flat launcher.models.claude entry; save one with moai update"
    }
}

/// Fills the model and effort a person did not give. A launch (no role) takes them from the
/// flat entry `moai update` saved under `launcher.models.<host>` of the user launcher settings
/// file (`model` and `effort`), else from the defaults of its host
/// (`model_effort_settings::launcher_default`, applied in [`claude_argv`]); an old `leader` or
/// `agent` entry beside it is not read and adds one note (SPEC-20261124-launcher-retire-role-
/// flags REQ-4, REQ-5, REQ-15). The legacy `session --role` request still takes the role row of
/// the saved entry or of the default table (SPEC-20261119-launcher-reads-models, SPEC-20261121-
/// model-defaults-one-table). The file is read at every call, so a rewritten file is used by the
/// next launch; a saved entry outside the support table refuses the launch with key paths only.
/// The glm host is Claude Code on z.ai: a saved id carries the `zai/` provider prefix, which the
/// z.ai endpoint does not take, so the prefix is removed here (the alias mapping of the
/// environment is unchanged). A flag the person gave stays, and only the other field is filled.
fn apply_saved_choice(r: &SessionRequest) -> Result<(SessionRequest, Vec<&'static str>), String> {
    if !matches!(r.host.as_str(), "claude" | "glm") {
        return Ok((r.clone(), Vec::new()));
    }
    // REQ-28: `-a` reads the agent row, like a role session; `-l` and a launch without a role
    // flag read the flat entry as before.
    let agent_flag = r.role.is_none() && r.first_input.as_deref() == Some(FIRST_INPUT_AGENT);
    let Some(role) = r.role.as_deref().or(agent_flag.then_some("agent")) else {
        return apply_flat_choice(r);
    };
    let saved = match user_launcher()? {
        Some(launcher) => {
            let unsupported = crate::model_effort_settings::unsupported_entries(&launcher)?;
            if !unsupported.is_empty() {
                return Err(format!(
                    "saved model settings are not supported ({}): replace them with moai update",
                    unsupported.join(", ")
                ));
            }
            crate::model_effort_settings::saved_choice(&launcher, &r.host, role)
        }
        None => None,
    };
    let from_saved = saved.is_some();
    let Some(choice) =
        saved.or_else(|| crate::model_effort_settings::default_choice(&r.host, role))
    else {
        return Ok((r.clone(), Vec::new()));
    };
    let (filled, used) = fill_from(r, &choice);
    let notes = if used && from_saved {
        vec![saved_note(&r.host, role)]
    } else {
        Vec::new()
    };
    Ok((filled, notes))
}

/// `r` with the model and effort it lacks taken from `choice`; whether a value was taken.
fn fill_from(
    r: &SessionRequest,
    choice: &crate::model_effort_settings::Choice,
) -> (SessionRequest, bool) {
    let model = if r.host == "glm" {
        choice.model.strip_prefix("zai/").unwrap_or(&choice.model)
    } else {
        choice.model.as_str()
    };
    let mut filled = r.clone();
    let mut used = false;
    if filled.model.is_none() {
        filled.model = Some(model.to_string());
        used = true;
    }
    if filled.effort.is_none() {
        filled.effort = Some(choice.effort.clone());
        used = true;
    }
    (filled, used)
}

/// The launch path of [`apply_saved_choice`]: the flat `launcher.models.<host>` entry.
fn apply_flat_choice(r: &SessionRequest) -> Result<(SessionRequest, Vec<&'static str>), String> {
    let Some(launcher) = user_launcher()? else {
        return Ok((r.clone(), Vec::new()));
    };
    let read = crate::model_effort_settings::flat_for_launch(&launcher, &r.host)?;
    let mut notes = Vec::new();
    if read.old_role_entries {
        notes.push(old_role_note(&r.host));
    }
    let Some(choice) = read.choice else {
        return Ok((r.clone(), notes));
    };
    let (filled, used) = fill_from(r, &choice);
    if used {
        notes.push(flat_note(&r.host));
    }
    Ok((filled, notes))
}

/// Composes the host command without running it. Returns
/// `{host, role, name, argv, env_keys, notes, executed:false, host_loaded:false}`.
/// `env_keys` lists variable names only, never values.
pub fn compose(r: &SessionRequest) -> Result<Value, String> {
    let (resolved, saved) = apply_saved_choice(r)?;
    compose_resolved(&resolved, saved)
}

fn compose_resolved(r: &SessionRequest, saved: Vec<&'static str>) -> Result<Value, String> {
    if let Some(role) = &r.role {
        if !ROLES.contains(&role.as_str()) {
            return Err("role must be leader or agent".into());
        }
    }
    if !HOSTS.contains(&r.host.as_str()) {
        return Err(format!("unknown host {} (claude|glm)", r.host));
    }
    if let Some(n) = &r.name {
        clean("name", n)?;
    }
    let mut notes = Vec::new();
    notes.extend(saved);
    if r.role.is_none() {
        notes.push(if r.first_input.is_some() {
            "a role flag (-l or -a) names the session, starts it in the bypass unless --permissions is given and hands the host the first input /moai leader or /moai agent (user decision 2026-11-25, SPEC-20261125-agent-team-session-names)"
        } else {
            "no role: a plain host session; no role or session name is added (#2053)"
        });
        if r.first_input.as_deref() == Some(FIRST_INPUT_AGENT) {
            notes.push("-a model and effort: --model and --effort, else the choice saved for the agent role, else the agent row of the default table (user decision 2026-11-25); no advisor unless --advisor [model] or launcher.advisor asks for one, and none on glm (user decision 2026-11-26)");
        } else {
            notes.push(if ROLELESS_LAUNCHER_DEFAULTS {
            if r.host == "glm" {
                "launcher defaults of glm: model opus, effort high, unless --model, --effort or a choice saved by moai update say otherwise (user decisions 2026-11-08 and 2026-11-24); no advisor unless --advisor or launcher.advisor asks for one (user decision 2026-11-12, #2476)"
            } else {
                "launcher defaults of claude: model opus, effort medium, unless --model, --effort or a choice saved by moai update say otherwise (user decision 2026-11-08: A); no advisor unless --advisor or launcher.advisor asks for one (user decision 2026-11-12, #2476)"
            }
        } else {
            "no model or effort flag without a role; the user's Claude settings apply; no advisor unless --advisor or launcher.advisor asks for one (#2476)"
        });
        }
    }
    let mut env_keys: Vec<&str> = Vec::new();
    let mut key_source: Option<&str> = None;
    let argv = match r.host.as_str() {
        "glm" => {
            match glm_token() {
                Err(b) => return Err(blocked_line(&b)),
                Ok(None) => {
                    return Err(format!(
                        "blocked: key_absent: {GLM_TOKEN_VAR} is not set or is empty in this environment and the vault holds no glm key (store it with `moai key-set glm` reading the value from stdin, or set the variable; the launcher never prints or stores the value)"
                    ));
                }
                Ok(Some((_, source))) => {
                    key_source = Some(source.name());
                    match source {
                        KeySource::Vault(_) => notes.push("ANTHROPIC_AUTH_TOKEN comes from the moai key vault; it is passed to the host process only"),
                        KeySource::Env => {}
                    }
                }
            }
            let mut plain = r.clone();
            if plain.advisor.take().is_some() {
                notes.push(NOTE_ADVISOR_OMITTED_GLM);
            }
            notes.push("model aliases map to z.ai models for this session: opus, sonnet and fable to glm-5.4[1m], haiku to glm-5.3-flash, with the 1M context window declared");
            env_keys.extend(GLM_ENV.iter().map(|(k, _)| *k));
            env_keys.push(GLM_TOKEN_VAR);
            env_keys.push(GLM_MCP_KEY_VAR);
            claude_argv(&plain, &mut notes)?
        }
        _ => claude_argv(r, &mut notes)?,
    };
    let mut out = json!({"host": r.host, "role": r.role, "name": r.name, "argv": argv,
        "env_keys": env_keys, "notes": notes, "host_loaded": false, "executed": false});
    // Only a launch that reads a key names its source (#2393 S2); the other hosts' answer
    // keeps its recorded protocol shape.
    if let Some(source) = key_source {
        out["key_source"] = json!(source);
    }
    Ok(out)
}

/// Composes and, unless `dry_run`, starts the host and waits for it to exit.
pub fn launch(r: &SessionRequest, dry_run: bool) -> Result<Value, String> {
    // The same request as `compose` uses.
    let (resolved, saved) = apply_saved_choice(r)?;
    let r = &resolved;
    let mut v = compose_resolved(r, saved)?;
    if dry_run {
        return Ok(v);
    }
    let argv: Vec<String> = serde_json::from_value(v["argv"].clone()).map_err(|_| "argv")?;
    let mut command = std::process::Command::new(&argv[0]);
    command.args(&argv[1..]);
    if r.host == "glm" {
        command.envs(GLM_ENV);
        // The same rule as `compose`: both names carry the trimmed token, so an empty variable
        // is replaced by the key file value and never reaches the host.
        let (token, _) = glm_token()
            .map_err(|b| blocked_line(&b))?
            .ok_or("blocked: the key changed")?;
        command.env(GLM_TOKEN_VAR, &token);
        command.env(GLM_MCP_KEY_VAR, &token);
    }
    let status = command
        .status()
        .map_err(|e| format!("cannot start {}: {e}", argv[0]))?;
    v["executed"] = json!(true);
    v["exit_code"] = json!(status.code());
    Ok(v)
}

/// SPEC-20261124-identity-home-arg REQ-7: with a symlinked home state folder in the environment
/// HOME the session folder answers as it does with a sound HOME.
#[cfg(all(test, unix))]
mod home_arg_claim_group_tests {
    use super::agent_seq_dir;

    #[test]
    fn home_arg_claim_group_session_agent_seq_dir() {
        crate::identity::linked_home_test::run(
            module_path!(),
            "home_arg_claim_group_session_agent_seq_dir",
            |f| {
                let dir = agent_seq_dir(&f.project).unwrap();
                assert!(dir.ends_with("sessions"), "{}", dir.display());
            },
        );
    }
}
