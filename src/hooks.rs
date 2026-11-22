//! Host hook generation (rc.2 I1 #2238): one fixed MoAI hook set rendered for Claude Code
//! (`.claude/settings.json` key `hooks`), with a dry-run plan, a digest-bound apply and an
//! ownership receipt; plus `hook-receipt`, the handler every generated hook runs, and the
//! `claim-verify --hook` guard (rc.2 A9 #2255) before Bash calls. Codex and OMP are retired
//! hosts: they are named and refused, never planned, applied or reverted.
//!
//! Source (read 2026-11-09): https://code.claude.com/docs/en/hooks — `hooks` is event name ->
//! matcher groups -> handlers; a `command` handler takes `type`, `command` and an optional
//! `timeout` in seconds; hooks in `.claude/settings.json` apply to that project.
//!
//! Events `SessionStart`, `PreToolUse` and `PostToolUse` (`Stop` was removed in #2936 R9 S1;
//! matcher `Bash`, stdin `tool_name`, `tool_input.command` and `cwd`, exit 2 blocks with the
//! reason on stderr), handler fields `type: "command"`, `command` and an explicit `timeout`.
//! Nothing here writes a home directory: a project equal to the given home, or inside its host
//! configuration directories, is refused. No host session runs here; whether a host loaded the
//! hooks is `host_loaded: not_run`.

use crate::Result;
use serde_json::{json, Map, Value};
use std::fs;
use std::io::Write;
use std::path::{Path, PathBuf};

/// Seconds every generated handler may run. Explicit on every handler; `hook-receipt` returns
/// in well under a second, and a host cancelling it loses one receipt line, nothing else.
pub const HOOK_TIMEOUT_SECS: u64 = 10;

/// What a generated hook runs.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Handler {
    /// `hook-receipt`: records allow-listed fields, never blocks (always exit 0).
    Receipt,
    /// `claim-verify . --hook`: the claim write boundary (rc.2 A9 #2255). Allows every call
    /// except a Bash `git commit` / `git push` in a lane worktree whose claim does not verify,
    /// which it refuses with exit 2 — the code both hosts treat as blocking.
    ClaimGuard,
    /// `hook-dispatch --event tool-preflight`: the hook engine's write and command guards
    /// (dangerous Bash, `.git` internals, credential folders, private keys and tokens). Claude
    /// Code only for now (#2877 leader Q1): it keeps the protection the Go-generation hooks gave
    /// once those are replaced. Answers with the documented PreToolUse deny object, never allow.
    Preflight,
    /// `hook-dispatch --event session-check`: answers with the session context (project state
    /// folder, lane branch, local claim record) the model reads at session start (#2936 R9 S1).
    /// It replaces the record-only receipt at `SessionStart`: nothing but `hooks-status` read
    /// that receipt.
    SessionCheck,
    /// `hook-dispatch --event edit-diagnostics` after a `Write|Edit|MultiEdit` (SPEC-20261117-
    /// edit-diagnostics S1): Claude Code only, generated only while `[hooks.edit_diagnostics]`
    /// is enabled ([`edit_diagnostics_timeout`]); never blocks.
    EditDiagnostics,
}

/// The handler timeout of the edit-diagnostics hook when it is generated: the checker's own
/// `timeout_secs` plus 5 seconds, so the host does not cancel the hook before MoAI stops the
/// checker. `None` (no hook line) for any host but Claude Code, for the legacy binary name, and
/// while the HOME `hook-config.toml` does not enable a well-formed `[hooks.edit_diagnostics]`
/// (leader decision Q4: generated only when enabled).
pub fn edit_diagnostics_timeout(binary: &str, host: &str) -> Option<u64> {
    if host != "claude" || binary != HOOK_BINARY {
        return None;
    }
    crate::hook_config::load()
        .config
        .edit
        .map(|r| r.timeout_secs + 5)
}

/// The tools the Claude preflight guard sees: the hook engine's Bash and write guards
/// (`hook_engine::capabilities("claude")`).
pub const CLAUDE_PREFLIGHT_MATCHER: &str = "Bash|Write|Edit|MultiEdit|NotebookEdit";

/// The MoAI hook set: (event, matcher, handler). The session context at session start, one
/// blocking guard before each Bash tool call (it refuses only an unverified claim's commit or
/// push, #2255 AC-2) and a receipt after each Bash tool call. `Stop` is not in the set: its
/// receipt had no reader but `hooks-status`, and the turn-validation rules are empty
/// (`hook_engine::STOP_RULES`) (#2936 R9 S1, selection table of Issue #2936 comment
/// 6881107273).
pub const HOOK_SET: &[(&str, Option<&str>, Handler)] = &[
    ("SessionStart", None, Handler::SessionCheck),
    ("PreToolUse", Some("Bash"), Handler::ClaimGuard),
    ("PostToolUse", Some("Bash"), Handler::Receipt),
];

/// The hosts MoAI writes hooks for.
pub const HOSTS: &[&str] = &["claude"];

/// Names a retired host as retired (the words `project-skill-plan` uses, plus the one host that
/// is served, which also makes the refusal an `invalid_input`) and any other host as unknown.
fn check_host(host: &str) -> Result<()> {
    if HOSTS.contains(&host) {
        Ok(())
    } else if crate::retired_hosts::retired(host, HOSTS) {
        Err(format!(
            "host {host} is retired; not managed; must be claude"
        ))
    } else {
        Err("unknown hook host; must be claude".into())
    }
}

/// The command a generated hook runs: the bare binary name, never an absolute path. The guard
/// names the project as `.`: both hosts run a command hook in the session's working directory.
pub fn hook_command(host: &str, handler: Handler) -> String {
    command_with(HOOK_BINARY, host, handler)
}

/// [`hook_command`] for any binary name. The legacy set (`moai-cli`, [`LEGACY_BINARY`]) is this
/// generator called with the old name, so what P4 rewrites cannot drift from what was written.
pub fn command_with(binary: &str, host: &str, handler: Handler) -> String {
    match handler {
        Handler::Receipt => format!("{binary} hook-receipt --host {host}"),
        Handler::ClaimGuard => format!("{binary} claim-verify . --hook"),
        Handler::SessionCheck => {
            format!("{binary} hook-dispatch --host {host} --event session-check")
        }
        Handler::Preflight => {
            format!("{binary} hook-dispatch --host {host} --event tool-preflight")
        }
        Handler::EditDiagnostics => {
            format!("{binary} hook-dispatch --host {host} --event edit-diagnostics")
        }
    }
}

/// The binary every generated hook command names (#3004 P4-2; `moai-cli` before, #2812 W3 and
/// leader decision 6868439066 (나)). A Go-generation `moai` first on PATH would receive the claim
/// guard, exit 1 and so switch the guard off silently; the gate in [`crate::hook_legacy`] refuses
/// to write such a hook set where the `moai` first on PATH is not the Rust build.
pub const HOOK_BINARY: &str = "moai";

/// The binary name the hook sets of every release before #3004 P4-2 call. Only the recognition
/// of those sets ([`crate::hook_legacy`]) uses it.
pub const LEGACY_BINARY: &str = "moai-cli";

/// Whether the generated command's binary resolves on `path_env` (#2348). `hooks-plan` and
/// `hooks-apply` report it next to the plan, outside the plan digest: the PATH a host gives
/// its hook process may differ from the shell's, so `resolved` only says this PATH finds the
/// Rust build, and the set is never reported active before a host receipt shows a hook ran.
///
/// `hook_binary` is the first `moai` on `path_env` (#3004), told apart by its `--version` line
/// (see [`crate::install::moai_on_path_state`]): `resolved` for the Rust build, `wrong_generation`
/// for the Go one or anything unrecognised, `binary_unresolved` when there is none, `not_probed`
/// (reason `project_tree`) for a relative path or one inside `project`, which is never run.
/// `hooks_active` is `"not_run"` for `resolved` and `false` for every other state. `home` is the
/// one variable the probed `moai` is given.
pub fn binary_state(project: &Path, path_env: &str, home: Option<&Path>) -> Value {
    let hook_binary = moai_state(project, path_env, home);
    let active = if hook_binary["state"] == json!("resolved") {
        json!("not_run")
    } else {
        json!(false)
    };
    json!({"hook_binary": hook_binary, "hooks_active": active})
}

/// The state object of the hook binary on `path_env`: [`binary_state`]'s `hook_binary`.
pub fn moai_state(project: &Path, path_env: &str, home: Option<&Path>) -> Value {
    crate::install::moai_on_path_state(project, path_env, home)
}

/// The hook set of one host: [`HOOK_SET`], and for Claude Code the preflight guard after the
/// claim guard (#2877 Q1).
pub fn hook_set(host: &str) -> Vec<(&'static str, Option<&'static str>, Handler)> {
    let mut set = HOOK_SET.to_vec();
    if host == "claude" {
        let at = set
            .iter()
            .position(|(e, _, _)| *e == "PreToolUse")
            .map_or(set.len(), |i| i + 1);
        set.insert(
            at,
            (
                "PreToolUse",
                Some(CLAUDE_PREFLIGHT_MATCHER),
                Handler::Preflight,
            ),
        );
    }
    set
}

/// The hook set as a host `hooks` object (event -> matcher groups -> handlers); two entries of
/// one event are two matcher groups, in [`hook_set`] order.
pub fn hooks_object(host: &str) -> Result<Map<String, Value>> {
    hooks_object_with(HOOK_BINARY, host)
}

/// [`hooks_object`] for any binary name (the legacy set is `hooks_object_with(LEGACY_BINARY, …)`).
pub fn hooks_object_with(binary: &str, host: &str) -> Result<Map<String, Value>> {
    check_host(host)?;
    let mut out = Map::new();
    for (event, matcher, handler) in hook_set(host) {
        let mut group = Map::new();
        if let Some(m) = matcher {
            group.insert("matcher".into(), json!(m));
        }
        group.insert(
            "hooks".into(),
            json!([{"type": "command", "command": command_with(binary, host, handler), "timeout": HOOK_TIMEOUT_SECS}]),
        );
        let groups = out
            .entry(event.to_string())
            .or_insert_with(|| Value::Array(Vec::new()));
        if let Some(list) = groups.as_array_mut() {
            list.push(Value::Object(group));
        }
    }
    if let Some(timeout) = edit_diagnostics_timeout(binary, host) {
        let group = json!({
            "matcher": crate::edit_diagnostics::MATCHER,
            "hooks": [{"type": "command",
                       "command": command_with(binary, host, Handler::EditDiagnostics),
                       "timeout": timeout}],
        });
        if let Some(list) = out
            .entry("PostToolUse".to_string())
            .or_insert_with(|| Value::Array(Vec::new()))
            .as_array_mut()
        {
            list.push(group);
        }
    }
    Ok(out)
}

/// The hooks object of `host` as the releases before #3004 P4-2 wrote it (the old binary name).
/// The same generator as the current set, so the two cannot drift.
pub fn legacy_hooks(host: &str) -> Result<Map<String, Value>> {
    hooks_object_with(LEGACY_BINARY, host)
}

/// The hooks object [`plan`] generates now for `host`; the counterpart of [`legacy_hooks`].
pub fn current_hooks(host: &str) -> Result<Map<String, Value>> {
    hooks_object_with(HOOK_BINARY, host)
}

/// The generated hook list a plan reports (no secrets: fixed strings only).
fn hook_list_with(binary: &str, host: &str) -> Value {
    let mut list: Vec<Value> = hook_set(host)
        .into_iter()
        .map(|(event, matcher, handler)| {
            json!({"event": event, "matcher": matcher, "command": command_with(binary, host, handler),
                   "timeout": HOOK_TIMEOUT_SECS})
        })
        .collect();
    if let Some(timeout) = edit_diagnostics_timeout(binary, host) {
        list.push(
            json!({"event": "PostToolUse", "matcher": crate::edit_diagnostics::MATCHER,
                         "command": command_with(binary, host, Handler::EditDiagnostics),
                         "timeout": timeout}),
        );
    }
    Value::Array(list)
}

/// Refuses a project that is the home directory or lies inside a host's home configuration.
pub(crate) fn project_root(project: &Path, home: Option<&Path>) -> Result<PathBuf> {
    let root = fs::canonicalize(project).map_err(|_| "project directory unavailable")?;
    if !root.is_dir() {
        return Err("project must be a directory".into());
    }
    if let Some(home) = home.and_then(|h| fs::canonicalize(h).ok()) {
        let inside_config = [
            ".claude",
            ".codex",
            ".omp",
            ".agents",
            crate::HOME_STATE_DIR,
        ]
        .iter()
        .any(|d| root.starts_with(home.join(d)));
        // #2696 C4b: `~/.moai` is refused only when it is this generation's HOME state folder
        // (marker present), and even then not under its worktree folder, where W2 creates
        // worktrees and the Go generation keeps its own (`home_workdir::worktrees_root`). A Go
        // `~/.moai` without the marker is not ours and is not refused here (audit #2848 B1).
        let inside_state = crate::home_root::dir(&home) == crate::home_root::MOAI_DIR
            && root.starts_with(home.join(crate::home_root::MOAI_DIR))
            && !root.starts_with(crate::home_workdir::worktrees_root(&home));
        if root == home || inside_config || inside_state {
            return Err("home directory refused; hooks are written to a project tree only".into());
        }
    }
    // C-79 (audit B1): the folders `CLAUDE_CONFIG_DIR` and `CODEX_HOME` name are user folders
    // wherever they are, so a project inside one, or whose own `.claude`/`.codex` is one, would
    // write the user-scope hooks; refused like `~/.claude`. An empty or relative value is
    // refused by name (the same guard as `claude.rs` and `codex_config.rs`).
    for (host, dir) in [
        (crate::host_home::Host::Claude, ".claude"),
        (crate::host_home::Host::Codex, ".codex"),
    ] {
        let named = crate::host_home::variable_dir(host, |k| std::env::var_os(k))
            .map_err(|e| format!("user config folder refused: {e}"))?;
        let Some(user) = named.and_then(|p| fs::canonicalize(p).ok()) else {
            continue;
        };
        let own = fs::canonicalize(root.join(dir)).ok();
        if root.starts_with(&user) || own.as_deref() == Some(user.as_path()) {
            return Err(format!(
                "user config folder refused: the project or its {dir} is the folder {} names; \
                 hooks are written to a project tree only",
                host.variable().unwrap_or_default()
            ));
        }
    }
    Ok(root)
}

fn check_permission(permission: &str) -> Result<()> {
    // Authorization stays with the host; the reference is caller metadata, not a verifier.
    if !crate::contract::permission_reference_ok(permission) {
        return Err("permission reference required".into());
    }
    Ok(())
}

/// Claude desired settings: the keys MoAI owns now with `hooks` set to the hook set, so the
/// plan never drops another owned key.
fn claude_desired(root: &Path) -> Result<Map<String, Value>> {
    let mut desired = crate::claude::owned_settings(root)?;
    desired.insert("hooks".into(), Value::Object(hooks_object("claude")?));
    Ok(desired)
}

/// Previews generating the MoAI hook set for `host` in PROJECT. Never writes. The returned
/// `plan_digest` binds the current files and the result; `hooks-apply` requires it.
pub fn plan(project: &Path, host: &str, home: Option<&Path>) -> Result<Value> {
    check_host(host)?;
    let root = project_root(project, home)?;
    let mut out = crate::claude::hooks_settings_plan(&root, &claude_desired(&root)?)?;
    out["path"] = out["settings_path"].take();
    out.as_object_mut().map(|m| m.remove("settings_path"));
    // What the binary_migration preview reads: the bytes the plan digest binds.
    let before = crate::claude::read_existing(&root.join(host_config_path(host)))?;
    if before.as_deref().map(crate::digest) != out["expected_sha256"].as_str().map(str::to_string) {
        return Err("the settings file changed while the plan was made; plan again".into());
    }
    out["host"] = json!(host);
    out["hooks"] = hook_list_with(HOOK_BINARY, host);
    // Outside the plan digest: what the hook engine covers on this host (#2391 stage 3).
    out["engine_capabilities"] = crate::hook_engine::capabilities(host);
    // Outside the plan digest too: which entries of the old `moai-cli` set the plan rewrites.
    let action = out["action"].as_str().unwrap_or("").to_string();
    out["binary_migration"] = crate::hook_legacy::migration(&crate::hook_legacy::Subject {
        host,
        file: host_config_path(host),
        before: before.as_deref(),
        action: &action,
    });
    Ok(out)
}

/// Applies a previewed hooks plan for `host`: re-plans against the files now, requires the
/// same digest, writes the host file only when it changes, then records the generated hook list
/// in the ownership receipt (REQ-I1-01/02).
pub fn apply(
    project: &Path,
    host: &str,
    home: Option<&Path>,
    plan_digest: &str,
    permission: &str,
) -> Result<Value> {
    check_host(host)?;
    check_permission(permission)?;
    let root = project_root(project, home)?;
    let mut out = crate::claude::hooks_settings_apply(
        &root,
        &claude_desired(&root)?,
        plan_digest,
        permission,
    )?;
    out["receipt_path"] = json!(crate::state_root::rel(
        &root,
        "manifests/claude-settings.json"
    ));
    out["host"] = json!(host);
    out["hooks_generated"] = hook_list_with(HOOK_BINARY, host);
    out["trust_required"] = json!("Claude Code runs project hooks only in a trusted workspace");
    Ok(out)
}

/// Where Claude Code keeps the project hook definitions MoAI generates.
fn host_config_path(_host: &str) -> &'static str {
    ".claude/settings.json"
}

/// Whether MoAI's hook set is installed for `host` in the project `root` (#2853): Claude Code
/// when its settings receipt owns `hooks`. An unreadable receipt is an error for the caller to
/// report, never "not installed".
pub fn installed(root: &Path, host: &str) -> Result<bool> {
    check_host(host)?;
    Ok(crate::claude::owned_settings(root)?.contains_key("hooks"))
}

/// The two files a hook apply for `host` writes in the project `root`, relative to it: the host
/// file and MoAI's ownership receipt for it (#2853 revert). Each folder between `root` and
/// these files is listed in `dirs`, so a caller can refuse one that became a link.
pub fn written_files(root: &Path, host: &str) -> Result<(Vec<String>, Vec<PathBuf>)> {
    check_host(host)?;
    let files = vec![
        host_config_path(host).to_string(),
        crate::state_root::rel(root, "manifests/claude-settings.json"),
    ];
    let state = crate::state_root::path(root, "");
    let mut dirs: Vec<PathBuf> = vec![root.join(".claude")];
    dirs.extend([state.clone(), state.join("manifests")]);
    Ok((files, dirs))
}

/// The JSON lines of `<root>/.moai-cli/receipts/<file>` (bounded by [`RECEIPT_LOG_LIMIT`];
/// a missing, oversized or non-regular log reads as empty).
fn receipt_lines(root: &Path, file: &str) -> Vec<Value> {
    receipt_lines_in(env_home().as_deref(), root, file)
}

/// The lines of one receipt log, oldest first: the project side (where every earlier hook
/// wrote) and then the HOME side (where a project with a fixed id writes, #3009 S4b). Each file
/// is bounded and read on its own; nothing is copied or removed.
pub(crate) fn receipt_lines_in(home: Option<&Path>, root: &Path, file: &str) -> Vec<Value> {
    let mut paths = vec![crate::state_root::path(root, "receipts").join(file)];
    paths.extend(home_receipts_dir(home, root).map(|d| d.join(file)));
    paths.iter().flat_map(|p| log_lines(p)).collect()
}

fn log_lines(path: &Path) -> Vec<Value> {
    match fs::symlink_metadata(path) {
        Ok(m) if m.is_file() && m.len() <= RECEIPT_LOG_LIMIT => {}
        _ => return Vec::new(),
    }
    fs::read_to_string(path)
        .unwrap_or_default()
        .lines()
        .filter_map(|l| serde_json::from_str(l).ok())
        .collect()
}

/// The user's HOME folder from `$HOME`; `None` when it is unset or empty.
pub fn env_home() -> Option<std::path::PathBuf> {
    std::env::var_os("HOME")
        .map(std::path::PathBuf::from)
        .filter(|h| !h.as_os_str().is_empty())
}

/// The HOME-side `receipts` folder of the project at `root` (see [`home_area_dir`]).
fn home_receipts_dir(home: Option<&Path>, root: &Path) -> Option<std::path::PathBuf> {
    home_area_dir(home, root, "receipts")
}

/// The HOME-side folder `area` (one of [`crate::project_home::AREAS`]) of the project at
/// `root`: `Some` only when `home` is given, its state root is a real folder, and the project's
/// key is its fixed id (a key made from the folder path is never written under, see
/// [`crate::project_home::ProjectHome::write_area`]). Nothing is created.
pub(crate) fn home_area_dir(
    home: Option<&Path>,
    root: &Path,
    area: &str,
) -> Option<std::path::PathBuf> {
    let home = home?;
    let home_state = crate::home_root::path(home);
    if !fs::symlink_metadata(&home_state).is_ok_and(|m| m.is_dir()) {
        return None;
    }
    crate::project_home::resolve(home, root)
        .ok()?
        .write_area(area)
        .ok()
        .flatten()
}

/// The hook state vocabulary of Issue #2391, each from evidence that exists locally, read only:
/// - `installed`: the hook binary (`moai`) resolves on `path_env`, whatever build it is.
/// - `binary` and `binary_generation` (#3004): the state and generation of the first `moai`
///   on `path_env` — `resolved` / `rust`, `wrong_generation` / `go` or `unknown`,
///   `binary_unresolved` / `absent`, `not_probed` / `not_probed` (see
///   [`crate::install::moai_on_path_state`]; `home` is the one variable its `--version` probe
///   gets). Only `resolved` is a `moai` the generated commands can use.
/// - `configured`: `yes` when the host file's `hooks` equal the generated hook set,
///   `legacy_binary` when they equal that set spelled with the `moai-cli` name of the releases
///   before #3004 P4-2 (it still runs; `hooks-plan` shows the rewrite), `no` when the file is
///   absent, `drifted` when it exists with other hooks (or does not parse).
/// - `trusted`: always `unknown`. Claude Code checks workspace trust before running any hook
///   from a settings file, but the page says nothing about where that is stored (read
///   2026-11-10), so it is never read here. No local line
///   proves a host ran the hooks either (see `verified`).
/// - `fired`: a receipt or engine line from this host stamped at or after the host file's
///   last modification (a changed definition makes older lines stale).
/// - `action_blocked`: since then, an engine `deny` line from this host, with a session key and
///   a call key, that no `observe` line of the same host, session and call at or after the deny
///   time follows — the denied tool did not run (audit H4-B1: another host's, another
///   session's or an earlier observation never cancels a deny).
/// - `verified`: always `false`. `fired` and `action_blocked` read local lines that any
///   `hook-dispatch` call writes, a direct command-line call included, so they show the engine
///   ran and answered — not that a host ran the hook or obeyed the deny (audit H4-B2). No
///   host-originated evidence is read; `evidence` names what the other two rest on.
pub fn status(project: &Path, host: &str, home: Option<&Path>, path_env: &str) -> Result<Value> {
    status_with(project, host, home, path_env, None)
}

/// [`status`] with the `moai` state already probed: `hooks-plan` reports it once as
/// `hook_binary` and hands that object here, so one command runs `moai --version` once
/// (#3004). `None` probes it.
pub fn status_with(
    project: &Path,
    host: &str,
    home: Option<&Path>,
    path_env: &str,
    moai: Option<&Value>,
) -> Result<Value> {
    check_host(host)?;
    let root = project_root(project, home)?;
    // Compared directly, not through the plan: the Claude settings plan preserves a user's
    // edit to an owned key and then reports `unchanged`, which is not "configured as generated".
    // A file with no `hooks` key at all is planned again there (C-55), never `unchanged`.
    let generated = Value::Object(hooks_object(host)?);
    // The set the releases before #3004 P4-2 wrote: the same generator, the old binary name.
    let legacy = Value::Object(hooks_object_with(LEGACY_BINARY, host)?);
    let configured = match crate::claude::read_existing(&root.join(host_config_path(host)))? {
        None => "no",
        Some(bytes) => match serde_json::from_slice::<Value>(&bytes) {
            Ok(v) if v["hooks"] == generated => "yes",
            Ok(v) if v["hooks"] == legacy => "legacy_binary",
            _ => "drifted",
        },
    };
    let config_ms = fs::metadata(root.join(host_config_path(host)))
        .and_then(|m| m.modified())
        .ok()
        .and_then(|t| t.duration_since(std::time::UNIX_EPOCH).ok())
        .map(|d| d.as_millis() as u64);
    let since = |line: &Value| {
        line["host"] == json!(host)
            && config_ms.is_some_and(|c| line["unix_ms"].as_u64().is_some_and(|t| t >= c))
    };
    let receipts = receipt_lines(&root, "hooks.jsonl");
    let engine = receipt_lines(&root, crate::hook_engine::ENGINE_LOG);
    let fired = receipts.iter().chain(engine.iter()).any(since);
    // A deny pairs only with an observation of the same host, session and call at or after it.
    let ran_later = |deny: &Value, deny_ms: u64| {
        engine.iter().any(|o| {
            o["kind"] == json!("observe")
                && o["host"] == deny["host"]
                && o["session"] == deny["session"]
                && o["call"] == deny["call"]
                && o["unix_ms"].as_u64().is_some_and(|t| t >= deny_ms)
        })
    };
    let action_blocked = engine.iter().any(|l| {
        let keyed = l["call"].is_string() && l["session"].as_str().is_some_and(|s| s != "unknown");
        l["kind"] == json!("deny")
            && since(l)
            && keyed
            && l["unix_ms"].as_u64().is_some_and(|t| !ran_later(l, t))
    });
    let installed = crate::install::resolve_on_path(path_env, HOOK_BINARY).is_some();
    let moai = match moai {
        Some(state) => state.clone(),
        None => moai_state(&root, path_env, home),
    };
    Ok(json!({
        "host": host,
        "installed": installed,
        "binary": moai["state"],
        "binary_generation": moai["generation"],
        "configured": configured,
        "trusted": "unknown",
        "fired": fired,
        "action_blocked": action_blocked,
        "verified": false,
        "evidence": "local engine and receipt lines; a direct hook-dispatch call writes the same lines, so no host run is proven",
        "config_path": host_config_path(host),
    }))
}

/// Size bound for a hook receipt log; past it lines are not appended (never truncated).
pub const RECEIPT_LOG_LIMIT: u64 = 4 * 1024 * 1024;

/// Appends one JSON line to `<cwd>/.moai-cli/receipts/<file>`, the write path every hook
/// handler shares (`hook-receipt`, and the hook engine's tool-observe, #2391). It writes only
/// when `<cwd>/.moai-cli` is a real directory (never through a symlink), creates `receipts/`
/// once, refuses a log that is not a regular file or has reached [`RECEIPT_LOG_LIMIT`], and
/// only appends — concurrent hooks never read-modify-write. The error is a fixed reason.
pub fn append_receipt_line(
    cwd: &Path,
    file: &str,
    line: &Value,
) -> std::result::Result<(), &'static str> {
    append_receipt_line_in(env_home().as_deref(), cwd, file, line)
}

/// [`append_receipt_line`] with the HOME named by the caller (#3009 S4b). The log goes to the
/// project's folder in the HOME (`<HOME state root>/projects/<project_id>/receipts/`) when the
/// project has a fixed id and the HOME state root is a real folder, so every worktree of the
/// project writes one log; in every other case it goes to `<cwd>/.moai-cli/receipts/` as before.
/// The working directory must be a project with a state folder either way. The HOME side is
/// created private (folders 0700, the log 0600) and is never reached through a link.
pub(crate) fn append_receipt_line_in(
    home: Option<&Path>,
    cwd: &Path,
    file: &str,
    line: &Value,
) -> std::result::Result<(), &'static str> {
    let state_dir = crate::state_root::path(cwd, "");
    match fs::symlink_metadata(&state_dir) {
        Ok(m) if m.is_dir() => {}
        _ => return Err("no .moai-cli directory in the working directory"),
    }
    let home_dir = home_receipts_dir(home, cwd);
    let dir = match &home_dir {
        Some(dir) => {
            let state = home
                .map(crate::home_root::path)
                .ok_or("home folder unavailable")?;
            make_private_dirs(dir, &state)?;
            dir.clone()
        }
        None => {
            let dir = state_dir.join("receipts");
            if let Ok(m) = fs::symlink_metadata(&dir) {
                if !m.is_dir() {
                    return Err("receipts path is not a directory");
                }
            } else if fs::create_dir(&dir).is_err() {
                return Err("receipts directory unavailable");
            }
            dir
        }
    };
    let log = dir.join(file);
    match fs::symlink_metadata(&log) {
        Ok(m) if !m.is_file() => return Err("receipt log is not a regular file"),
        Ok(m) if m.len() >= RECEIPT_LOG_LIMIT => return Err("receipt log full"),
        _ => {}
    }
    let mut bytes = serde_json::to_vec(line).map_err(|_| "receipt line encode failed")?;
    bytes.push(b'\n');
    let mut open = fs::OpenOptions::new();
    open.create(true).append(true);
    #[cfg(unix)]
    if home_dir.is_some() {
        use std::os::unix::fs::OpenOptionsExt;
        open.mode(0o600);
    }
    open.open(&log)
        .and_then(|mut f| f.write_all(&bytes))
        .map_err(|_| "receipt log write failed")
}

/// Creates `dir` and the folders between it and the existing `base` (the HOME state root),
/// each private (0700). Every folder must end up a real folder: a link or a file in the way is
/// refused. `dir` must lie under `base`.
pub(crate) fn make_private_dirs(dir: &Path, base: &Path) -> std::result::Result<(), &'static str> {
    let Ok(rel) = dir.strip_prefix(base) else {
        return Err("receipts directory unavailable");
    };
    let mut cur = base.to_path_buf();
    for part in rel.components() {
        cur.push(part);
        match fs::symlink_metadata(&cur) {
            Ok(m) if m.is_dir() => continue,
            Ok(_) => return Err("receipts path is not a directory"),
            Err(_) => {}
        }
        let mut builder = fs::DirBuilder::new();
        #[cfg(unix)]
        {
            use std::os::unix::fs::DirBuilderExt;
            builder.mode(0o700);
        }
        // Another hook may create the folder between the check above and this call; a real
        // folder that is there afterwards is what was wanted (a link or a file is not).
        if builder.create(&cur).is_err() && !fs::symlink_metadata(&cur).is_ok_and(|m| m.is_dir()) {
            return Err("receipts directory unavailable");
        }
    }
    Ok(())
}
/// Hook stdin fields a receipt keeps. Everything else — `tool_input`, `tool_response`,
/// `prompt`, transcript paths — is never read into the log.
const RECEIPT_FIELDS: &[&str] = &["hook_event_name", "tool_name", "source"];
const FIELD_LIMIT: usize = 128;

/// `hook-receipt`: the handler the generated hooks run. Reads the host's hook stdin and appends
/// one line to `<cwd>/.moai-cli/receipts/hooks.jsonl` with the allow-listed fields, the host and
/// the time. It never fails the hook: every refusal is `recorded: false` with a reason, so the
/// host sees exit 0. It writes only when the working directory already has `.moai-cli/`.
pub fn receipt(host: &str, stdin: &[u8], cwd: &Path, now_unix_ms: u128) -> Value {
    let skip = |reason: &str| json!({"recorded": false, "reason": reason});
    if check_host(host).is_err() {
        return skip("unknown hook host");
    }
    let Ok(Value::Object(input)) = serde_json::from_slice::<Value>(stdin) else {
        return skip("hook input is not a JSON object");
    };
    let mut fields = Map::new();
    let mut dropped = Vec::new();
    for key in RECEIPT_FIELDS {
        match input.get(*key) {
            None => {}
            Some(Value::String(s))
                if s.len() <= FIELD_LIMIT
                    && !s.chars().any(char::is_control)
                    && !crate::redact::contains_secret_shape(s) =>
            {
                fields.insert((*key).into(), json!(s));
            }
            Some(_) => dropped.push(*key),
        }
    }
    let mut line = json!({"schema_version": 1, "host": host, "unix_ms": now_unix_ms as u64});
    for (k, v) in &fields {
        line[k] = v.clone();
    }
    if let Err(reason) = append_receipt_line(cwd, "hooks.jsonl", &line) {
        return skip(reason);
    }
    json!({
        "recorded": true,
        "path": crate::state_root::rel(cwd, "receipts/hooks.jsonl"),
        "fields": fields,
        "dropped": dropped,
    })
}

#[cfg(test)]
mod tests {
    use super::*;

    /// The legacy generator is the current one called with the old name: it rebuilds the
    /// commands of the releases before #3004 P4-2 exactly.
    #[test]
    fn the_legacy_generator_rebuilds_the_pre_p4_commands() {
        for host in HOSTS {
            for (event, groups) in current_hooks(host).unwrap() {
                let old = &legacy_hooks(host).unwrap()[&event];
                for (new_group, old_group) in groups
                    .as_array()
                    .unwrap()
                    .iter()
                    .zip(old.as_array().unwrap())
                {
                    let new_cmd = new_group["hooks"][0]["command"].as_str().unwrap();
                    let old_cmd = old_group["hooks"][0]["command"].as_str().unwrap();
                    assert!(
                        new_cmd.starts_with("moai ") && !new_cmd.contains("moai-cli"),
                        "{new_cmd}"
                    );
                    assert_eq!(old_cmd, new_cmd.replacen("moai ", "moai-cli ", 1));
                }
            }
        }
    }

    /// A retired host is named as retired, any other name as unknown, and neither is planned.
    #[test]
    fn a_retired_host_is_named_and_an_unknown_one_is_not() {
        let dir = tempfile::tempdir().unwrap();
        for host in ["codex", "omp"] {
            let want = format!("host {host} is retired; not managed; must be claude");
            assert_eq!(check_host(host).unwrap_err(), want);
            assert_eq!(plan(dir.path(), host, None).unwrap_err(), want);
            assert_eq!(installed(dir.path(), host).unwrap_err(), want);
            assert_eq!(written_files(dir.path(), host).unwrap_err(), want);
            assert_eq!(status(dir.path(), host, None, "").unwrap_err(), want);
            assert_eq!(hooks_object(host).unwrap_err(), want);
            assert_eq!(
                apply(dir.path(), host, None, "sha256:0", "test-permission").unwrap_err(),
                want
            );
        }
        for host in ["gemini", "", "Codex"] {
            assert_eq!(
                check_host(host).unwrap_err(),
                "unknown hook host; must be claude",
                "{host:?}"
            );
        }
        assert!(check_host("claude").is_ok());
        assert!(
            !dir.path().join(".codex").exists() && !dir.path().join(".omp").exists(),
            "nothing was written"
        );
    }
}

/// Where the hook logs go and how both places are read (#3009 S4b).
#[cfg(test)]
mod receipt_home_tests {
    use super::*;
    use std::path::PathBuf;

    const ID: &str = "abcdefghijklmnopqrstuvwxyz";

    /// A project with a state folder (and a profile when `id` is given) and a HOME with a state
    /// root (when `home_state` is true).
    fn world(id: Option<&str>, home_state: bool) -> (tempfile::TempDir, tempfile::TempDir) {
        let (home, project) = (tempfile::tempdir().unwrap(), tempfile::tempdir().unwrap());
        let state = project.path().join(crate::PROJECT_STATE_DIR);
        fs::create_dir_all(&state).unwrap();
        if let Some(id) = id {
            fs::write(
                state.join("project.yaml"),
                format!("{{\"project_id\":\"{id}\"}}"),
            )
            .unwrap();
        }
        if home_state {
            fs::create_dir(crate::home_root::path(home.path())).unwrap();
        }
        (home, project)
    }

    fn home_log(home: &Path, id: &str, file: &str) -> PathBuf {
        crate::home_root::path(home)
            .join("projects")
            .join(id)
            .join("receipts")
            .join(file)
    }

    fn project_log(project: &Path, file: &str) -> PathBuf {
        project
            .join(crate::PROJECT_STATE_DIR)
            .join("receipts")
            .join(file)
    }

    /// Every path under `dir`, so a test can show nothing was created.
    fn listing(dir: &Path) -> Vec<String> {
        let mut out = Vec::new();
        let mut stack = vec![dir.to_path_buf()];
        while let Some(d) = stack.pop() {
            for e in fs::read_dir(&d).unwrap().flatten() {
                out.push(e.path().strip_prefix(dir).unwrap().display().to_string());
                if e.path().is_dir() {
                    stack.push(e.path());
                }
            }
        }
        out.sort();
        out
    }

    fn line(n: u64) -> Value {
        json!({"kind": "t", "n": n})
    }

    fn append(
        home: Option<&Path>,
        project: &Path,
        n: u64,
    ) -> std::result::Result<(), &'static str> {
        append_receipt_line_in(home, project, "hooks.jsonl", &line(n))
    }

    #[test]
    fn a_project_with_an_id_logs_in_the_home_and_leaves_its_own_folder_alone() {
        let (home, project) = world(Some(ID), true);
        let p0 = listing(project.path());
        append(Some(home.path()), project.path(), 1).unwrap();
        append(Some(home.path()), project.path(), 2).unwrap();
        let text = fs::read_to_string(home_log(home.path(), ID, "hooks.jsonl")).unwrap();
        assert_eq!(text.lines().count(), 2, "{text}");
        assert_eq!(
            listing(project.path()),
            p0,
            "the project folder is untouched"
        );
    }

    #[cfg(unix)]
    #[test]
    fn the_home_side_is_private() {
        use std::os::unix::fs::PermissionsExt;
        let (home, project) = world(Some(ID), true);
        append(Some(home.path()), project.path(), 1).unwrap();
        let mode = |p: &Path| fs::metadata(p).unwrap().permissions().mode() & 0o777;
        let log = home_log(home.path(), ID, "hooks.jsonl");
        assert_eq!(mode(&log), 0o600);
        let receipts = log.parent().unwrap();
        let key_dir = receipts.parent().unwrap();
        let projects = key_dir.parent().unwrap();
        for dir in [receipts, key_dir, projects] {
            assert_eq!(mode(dir), 0o700, "{}", dir.display());
        }
    }

    #[test]
    fn every_other_case_logs_in_the_project_as_before_and_leaves_the_home_alone() {
        // No profile, so the key would come from the path; no HOME state root; no HOME at all.
        let cases: [(Option<&str>, bool, bool); 3] = [
            (None, true, true),
            (Some(ID), false, true),
            (Some(ID), true, false),
        ];
        for (id, home_state, with_home) in cases {
            let (home, project) = world(id, home_state);
            let h0 = listing(home.path());
            let home_arg = with_home.then_some(home.path());
            append(home_arg, project.path(), 1).unwrap();
            let text = fs::read_to_string(project_log(project.path(), "hooks.jsonl")).unwrap();
            assert_eq!(text.lines().count(), 1, "{id:?} {home_state} {with_home}");
            assert_eq!(listing(home.path()), h0, "{id:?} {home_state} {with_home}");
        }
    }

    #[test]
    fn a_folder_that_is_not_a_project_gets_no_log_on_either_side() {
        let (home, _) = world(Some(ID), true);
        let bare = tempfile::tempdir().unwrap();
        let h0 = listing(home.path());
        assert!(append(Some(home.path()), bare.path(), 1).is_err());
        assert_eq!(listing(home.path()), h0);
        assert!(listing(bare.path()).is_empty());
    }

    #[cfg(unix)]
    #[test]
    fn a_link_or_a_file_in_the_way_on_the_home_side_is_refused() {
        let (home, project) = world(Some(ID), true);
        let elsewhere = tempfile::tempdir().unwrap();
        let projects = crate::home_root::path(home.path()).join("projects");
        std::os::unix::fs::symlink(elsewhere.path(), &projects).unwrap();
        assert!(append(Some(home.path()), project.path(), 1).is_err());
        assert!(
            listing(elsewhere.path()).is_empty(),
            "nothing is written through the link"
        );
        fs::remove_file(&projects).unwrap();
        // A file where `receipts` should be.
        let receipts = home_log(home.path(), ID, "x")
            .parent()
            .unwrap()
            .to_path_buf();
        fs::create_dir_all(receipts.parent().unwrap()).unwrap();
        fs::write(&receipts, "not a folder").unwrap();
        assert!(append(Some(home.path()), project.path(), 1).is_err());
    }

    #[test]
    fn a_full_home_log_refuses_at_the_limit_and_takes_a_line_just_below_it() {
        let (home, project) = world(Some(ID), true);
        append(Some(home.path()), project.path(), 1).unwrap();
        let log = home_log(home.path(), ID, "hooks.jsonl");
        let set_len = |n: u64| {
            fs::OpenOptions::new()
                .write(true)
                .open(&log)
                .unwrap()
                .set_len(n)
                .unwrap()
        };
        set_len(RECEIPT_LOG_LIMIT - 1);
        assert!(append(Some(home.path()), project.path(), 2).is_ok());
        set_len(RECEIPT_LOG_LIMIT);
        assert_eq!(
            append(Some(home.path()), project.path(), 3),
            Err("receipt log full")
        );
    }

    #[test]
    fn worktrees_with_one_id_share_one_home_log() {
        let (home, first) = world(Some(ID), true);
        let (_, second) = world(Some(ID), false);
        append(Some(home.path()), first.path(), 1).unwrap();
        append(Some(home.path()), second.path(), 2).unwrap();
        let text = fs::read_to_string(home_log(home.path(), ID, "hooks.jsonl")).unwrap();
        assert_eq!(text.lines().count(), 2, "{text}");
        for p in [&first, &second] {
            let lines = receipt_lines_in(Some(home.path()), p.path(), "hooks.jsonl");
            assert_eq!(lines.len(), 2, "each worktree sees both lines");
        }
    }

    #[test]
    fn reading_gives_the_project_side_first_and_then_the_home_side() {
        let (home, project) = world(Some(ID), true);
        let old = project_log(project.path(), "hooks.jsonl");
        fs::create_dir_all(old.parent().unwrap()).unwrap();
        fs::write(&old, "{\"n\":1}\nnot json\n{\"n\":2}\n").unwrap();
        let new = home_log(home.path(), ID, "hooks.jsonl");
        fs::create_dir_all(new.parent().unwrap()).unwrap();
        fs::write(&new, "{\"n\":3}\n").unwrap();
        let numbers = |home: Option<&Path>| -> Vec<u64> {
            receipt_lines_in(home, project.path(), "hooks.jsonl")
                .iter()
                .map(|v| v["n"].as_u64().unwrap())
                .collect()
        };
        assert_eq!(numbers(Some(home.path())), vec![1, 2, 3]);
        // Without a HOME only the project side is read.
        assert_eq!(numbers(None), vec![1, 2]);
    }

    #[test]
    fn an_oversized_log_on_one_side_does_not_hide_the_other_side() {
        let (home, project) = world(Some(ID), true);
        let old = project_log(project.path(), "hooks.jsonl");
        fs::create_dir_all(old.parent().unwrap()).unwrap();
        fs::write(&old, "{\"n\":1}\n").unwrap();
        fs::OpenOptions::new()
            .write(true)
            .open(&old)
            .unwrap()
            .set_len(RECEIPT_LOG_LIMIT + 1)
            .unwrap();
        let new = home_log(home.path(), ID, "hooks.jsonl");
        fs::create_dir_all(new.parent().unwrap()).unwrap();
        fs::write(&new, "{\"n\":3}\n").unwrap();
        let got = receipt_lines_in(Some(home.path()), project.path(), "hooks.jsonl");
        assert_eq!(got.len(), 1);
        assert_eq!(got[0]["n"], json!(3));
    }

    #[test]
    fn concurrent_hooks_never_break_a_line() {
        let (home, project) = world(Some(ID), true);
        let (h, p) = (home.path().to_path_buf(), project.path().to_path_buf());
        let threads: Vec<_> = (0..8u64)
            .map(|t| {
                let (h, p) = (h.clone(), p.clone());
                std::thread::spawn(move || {
                    for i in 0..40u64 {
                        append(Some(&h), &p, t * 100 + i).unwrap();
                    }
                })
            })
            .collect();
        for t in threads {
            t.join().unwrap();
        }
        let text = fs::read_to_string(home_log(home.path(), ID, "hooks.jsonl")).unwrap();
        assert_eq!(text.lines().count(), 320);
        assert!(text
            .lines()
            .all(|l| serde_json::from_str::<Value>(l).is_ok()));
    }
}
