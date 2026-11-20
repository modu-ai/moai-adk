//! Host hook generation (rc.2 I1 #2238): one fixed MoAI hook set rendered for Claude Code
//! (`.claude/settings.json` key `hooks`) and Codex (`.codex/hooks.json`), with a dry-run plan,
//! a digest-bound apply and an ownership receipt; plus `hook-receipt`, the handler every
//! generated hook runs, and the `claim-verify --hook` guard (rc.2 A9 #2255) before Bash calls.
//!
//! Sources (read 2026-11-09):
//! - https://code.claude.com/docs/en/hooks — `hooks` is event name -> matcher groups ->
//!   handlers; a `command` handler takes `type`, `command` and an optional `timeout` in
//!   seconds; hooks in `.claude/settings.json` apply to that project.
//! - https://learn.chatgpt.com/docs/hooks — Codex reads `<repo>/.codex/hooks.json` with the
//!   same event -> matcher group -> handler shape and an optional top-level `description`;
//!   project hooks load only when the project `.codex/` layer is trusted, and non-managed hooks
//!   are reviewed and trusted (by hash) before they run.
//!
//! Only the intersection both hosts accept is used: events `SessionStart`, `PreToolUse` and
//! `PostToolUse` (`Stop` was removed in #2936 R9 S1; both hosts: matcher `Bash`, stdin `tool_name`,
//! `tool_input.command` and `cwd`, exit 2 blocks with the reason on stderr), handler fields `type: "command"`, `command` and an explicit `timeout`, so neither
//! host's defaults matter. Nothing here writes a home directory: a project equal to the given
//! home, or inside its host configuration directories, is refused. No host session runs here;
//! whether a host loaded the hooks is `host_loaded: not_run`.

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
/// (`hook_engine::STOP_RULES`). The Codex lane Stop hook is added by [`codex_hooks`], not here
/// (#2936 R9 S1, selection table of Issue #2936 comment 6881107273).
pub const HOOK_SET: &[(&str, Option<&str>, Handler)] = &[
    ("SessionStart", None, Handler::SessionCheck),
    ("PreToolUse", Some("Bash"), Handler::ClaimGuard),
    ("PostToolUse", Some("Bash"), Handler::Receipt),
];

pub const HOSTS: &[&str] = &["claude", "codex"];
// The OMP project-hooks entry below stays until its own slice removes it (SPEC-20261120-drop-omp N1).

const CODEX_FILE: &str = ".codex/hooks.json";
/// Codex ownership receipt under the project state root.
const CODEX_RECEIPT: &str = "manifests/codex-hooks.json";
/// The `description` of the Codex hook file, naming the binary that generated it.
fn codex_description(binary: &str) -> String {
    format!("MoAI receipt hooks (generated by {binary} hooks-apply; edits make the next plan stop for review)")
}

/// A host whose whole hook file MoAI owns, recorded by an ownership receipt (Codex, OMP).
struct FileHost {
    host: &'static str,
    /// The hook file, relative to the project.
    file: &'static str,
    /// The receipt, relative to the project state root.
    receipt: &'static str,
    /// Folders under the project between it and the file, each refused when a link.
    dirs: &'static [&'static str],
}

const CODEX_HOST: FileHost = FileHost {
    host: "codex",
    file: CODEX_FILE,
    receipt: CODEX_RECEIPT,
    dirs: &[".codex"],
};

/// OMP loads project extensions from `<cwd>/.omp/extensions` (github.com/can1357/oh-my-pi
/// `docs/extension-loading.md`, read 2026-11-13; #2633).
const OMP_HOST: FileHost = FileHost {
    host: "omp",
    file: ".omp/extensions/moai-hooks.ts",
    receipt: "manifests/omp-hooks.json",
    dirs: &[".omp", ".omp/extensions"],
};

fn file_host(host: &str) -> &'static FileHost {
    if host == "omp" {
        &OMP_HOST
    } else {
        &CODEX_HOST
    }
}

/// The OMP extension module (#2633). OMP runs no command hooks, so this module calls
/// `hook-dispatch --host omp` with a Claude-shaped input and turns the answer into OMP's own
/// return values (`docs/extensions.md`, `extensibility/extensions/types.ts`, read 2026-11-13):
/// - `tool_call` on `bash`: `{block, reason}`; the reason reaches the model as the tool error.
/// - `before_agent_start`: `systemPrompt` replaces the prompt (runner.ts "Replace policy"), so
///   the context is appended to the prompt it received, never put in its place.
/// - `tool_result` on `bash`: observation only, `isError` as `tool_response.is_error`.
///
/// OMP treats a throwing or timed-out `tool_call` handler as a block; the module instead lets
/// the call through when `moai` is missing, fails or exceeds the timeout — the
/// non-blocking engine error of the other hosts. Only `bash` is guarded (see
/// `hook_engine::capabilities`).
///
/// The text is a template: [`omp_module`] puts the binary name in its two places, so the module
/// the releases before #3004 P4-2 wrote (`moai-cli`) can be rebuilt byte for byte.
const OMP_MODULE_TEMPLATE: &str = r#"// MoAI hooks for OMP (generated by @BINARY@ hooks-apply, Issue #2633; edits make the next plan stop for review).
import type { ExtensionAPI } from "@oh-my-pi/pi-coding-agent";

const TIMEOUT_MS = 10000;

async function dispatch(event: string, input: unknown, cwd: string): Promise<any> {
	try {
		const proc = Bun.spawn(["@BINARY@", "hook-dispatch", "--host", "omp", "--event", event], {
			cwd,
			stdin: "pipe",
			stdout: "pipe",
			stderr: "ignore",
		});
		proc.stdin.write(JSON.stringify(input));
		proc.stdin.end();
		const timer = setTimeout(() => proc.kill(), TIMEOUT_MS);
		const [out, code] = await Promise.all([new Response(proc.stdout).text(), proc.exited]);
		clearTimeout(timer);
		if (code !== 0 || out.trim() === "") return null;
		return JSON.parse(out);
	} catch {
		return null;
	}
}

function bashInput(event: any, cwd: string, hookEventName: string): Record<string, unknown> {
	return {
		hook_event_name: hookEventName,
		tool_name: "Bash",
		tool_use_id: String(event.toolCallId ?? ""),
		tool_input: { command: String(event.input?.command ?? "") },
		cwd,
	};
}

export default function (pi: ExtensionAPI) {
	pi.on("before_agent_start", async (event, ctx) => {
		const answer = await dispatch("session-check", { hook_event_name: "SessionStart", cwd: ctx.cwd }, ctx.cwd);
		if (answer && typeof answer.context === "string") {
			return { systemPrompt: [...event.systemPrompt, answer.context] };
		}
	});
	pi.on("tool_call", async (event, ctx) => {
		if (event.toolName !== "bash") return;
		const answer = await dispatch("tool-preflight", bashInput(event, ctx.cwd, "PreToolUse"), ctx.cwd);
		if (answer && answer.block === true) {
			return { block: true, reason: String(answer.reason ?? "blocked by MoAI") };
		}
	});
	pi.on("tool_result", async (event, ctx) => {
		if (event.toolName !== "bash") return;
		const input = { ...bashInput(event, ctx.cwd, "PostToolUse"), tool_response: { is_error: typeof event.isError === "boolean" ? event.isError : null } };
		await dispatch("tool-observe", input, ctx.cwd);
	});
}
"#;

/// The OMP extension module for `binary` ([`HOOK_BINARY`] for what is generated now).
pub fn omp_module(binary: &str) -> String {
    OMP_MODULE_TEMPLATE.replace("@BINARY@", binary)
}

/// What the OMP module wires: (OMP event, engine duty).
const OMP_EVENTS: &[(&str, &str)] = &[
    ("before_agent_start", "session-check"),
    ("tool_call", "tool-preflight"),
    ("tool_result", "tool-observe"),
];

fn check_host(host: &str) -> Result<()> {
    if HOSTS.contains(&host) {
        Ok(())
    } else {
        Err("unknown hook host; must be claude, codex or omp".into())
    }
}

/// A host this version no longer plans or writes for, whose receipts an older binary may
/// have written. The revert paths still recognize it — the same decision the agents
/// receipts made (leader decision on the removal audit) — while the plan-and-write paths
/// keep refusing it (they answer a skip with a notice instead).
const RETIRED: [&str; 1] = ["omp"];

/// `check_host` for a path that only reads what an older binary wrote ([`written_files`],
/// called from the revert): the retired hosts are recognized there, so their receipts
/// revert exactly as they did before.
fn check_host_written(host: &str) -> Result<()> {
    if HOSTS.contains(&host) || RETIRED.contains(&host) {
        Ok(())
    } else {
        Err("unknown hook host; must be claude, codex or omp".into())
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

/// The opt-in lane `Stop` hook of a Codex project (#2904 stage 2, #2899 D-O2): the only `Stop`
/// matcher group (the default set has none since #2936 R9 S1) that runs
/// `lane-stop-hook --lane L --repo O/R`.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum LaneLoop {
    /// What the Codex ownership receipt records under `lane_loop`; none recorded → the base
    /// set. Every caller that does not name a lane (`hooks-plan` without the flags, `update`).
    Keep,
    /// Add the lane's Stop hook.
    Set { lane: String, repo: String },
    /// The base set, without a lane Stop hook (the revert).
    Off,
}

/// A lane Stop hook as a plan renders it and the receipt records it.
#[derive(Clone, Debug, PartialEq, Eq)]
struct Lane {
    lane: String,
    repo: String,
}

impl Lane {
    fn checked(lane: &str, repo: &str) -> Option<Lane> {
        (crate::lane_loop::lane_ok(lane) && crate::lane_loop::repo_ok(repo)).then(|| Lane {
            lane: lane.into(),
            repo: repo.into(),
        })
    }

    fn command_with(&self, binary: &str) -> String {
        crate::lane_loop::stop_hook_command_with(binary, &self.lane, &self.repo)
    }

    fn json(&self) -> Value {
        json!({"lane": self.lane, "repo": self.repo})
    }
}

/// A lane is named for Codex only, and only a well-formed one.
fn check_lane(host: &str, lane: &LaneLoop) -> Result<()> {
    if host != "codex" && *lane != LaneLoop::Keep {
        return Err("--lane-loop and --lane-loop-off are for --host codex only".into());
    }
    match lane {
        LaneLoop::Set { lane, repo } if Lane::checked(lane, repo).is_none() => {
            Err("--lane-loop must be 1-64 of A-Z a-z 0-9 . _ - and --repo OWNER/REPO".into())
        }
        _ => Ok(()),
    }
}

/// The Codex hooks object: the hook set, plus the lane's Stop group. The base set has no `Stop`
/// event, so the lane group is the event's only group; an absent `Stop` is created, never
/// skipped (a skipped lane hook would vanish with no error while its receipt still names it).
fn codex_hooks(lane: Option<&Lane>) -> Result<Map<String, Value>> {
    codex_hooks_with(HOOK_BINARY, lane)
}

fn codex_hooks_with(binary: &str, lane: Option<&Lane>) -> Result<Map<String, Value>> {
    let mut out = hooks_object_with(binary, "codex")?;
    if let Some(l) = lane {
        let stop = out.entry("Stop").or_insert_with(|| json!([]));
        let Some(groups) = stop.as_array_mut() else {
            return Err("the Stop hook entry is not a list".into());
        };
        groups.push(
            json!({"hooks": [{"type": "command", "command": l.command_with(binary),
            "timeout": crate::lane_loop::STOP_HOOK_TIMEOUT_SECS}]}),
        );
    }
    Ok(out)
}

/// [`hook_list`] with the lane's Stop hook last, when there is one.
fn hook_list_for(host: &str, lane: Option<&Lane>) -> Value {
    hook_list_for_with(HOOK_BINARY, host, lane)
}

fn hook_list_for_with(binary: &str, host: &str, lane: Option<&Lane>) -> Value {
    let mut list = hook_list_with(binary, host);
    if let (Some(l), Some(items)) = (lane, list.as_array_mut()) {
        items.push(
            json!({"event": "Stop", "matcher": null, "command": l.command_with(binary),
            "timeout": crate::lane_loop::STOP_HOOK_TIMEOUT_SECS}),
        );
    }
    list
}

/// The hooks object of `host` as the releases before #3004 P4-2 wrote it (the old binary name),
/// with the Codex lane Stop group when `lane` (lane, repo) is given. The same generators as the
/// current set, so the two cannot drift.
pub fn legacy_hooks(host: &str, lane: Option<(&str, &str)>) -> Result<Map<String, Value>> {
    hooks_with(LEGACY_BINARY, host, lane)
}

/// The hooks object [`plan`] generates now for `host` (with the Codex lane Stop group when
/// `lane` is given); the counterpart of [`legacy_hooks`].
pub fn current_hooks(host: &str, lane: Option<(&str, &str)>) -> Result<Map<String, Value>> {
    hooks_with(HOOK_BINARY, host, lane)
}

fn hooks_with(binary: &str, host: &str, lane: Option<(&str, &str)>) -> Result<Map<String, Value>> {
    if host == "codex" {
        let lane = lane.and_then(|(l, r)| Lane::checked(l, r));
        codex_hooks_with(binary, lane.as_ref())
    } else {
        hooks_object_with(binary, host)
    }
}

/// The Codex hook file as the releases before #3004 P4-2 wrote it.
pub fn legacy_codex_file(lane: Option<(&str, &str)>) -> Result<Vec<u8>> {
    let lane = lane.and_then(|(l, r)| Lane::checked(l, r));
    codex_render_with(LEGACY_BINARY, lane.as_ref())
}

/// The hook list of an ownership receipt (and of a plan) for `binary`.
pub fn hook_list_of(binary: &str, host: &str, lane: Option<(&str, &str)>) -> Value {
    let lane = lane.and_then(|(l, r)| Lane::checked(l, r));
    hook_list_for_with(binary, host, lane.as_ref())
}

/// The lane a Codex receipt records; a malformed `lane_loop` stops for review.
fn receipt_lane(receipt: &[u8], fh: &FileHost) -> Result<Option<Lane>> {
    let review = || format!("{} receipt requires preservation review", fh.file);
    let r: Value = serde_json::from_slice(receipt).map_err(|_| review())?;
    match &r["lane_loop"] {
        Value::Null => Ok(None),
        v => match (v["lane"].as_str(), v["repo"].as_str()) {
            (Some(lane), Some(repo)) => Lane::checked(lane, repo).map(Some).ok_or_else(review),
            _ => Err(review()),
        },
    }
}

/// The generated hook list a plan and a receipt report (no secrets: fixed strings only).
fn hook_list_with(binary: &str, host: &str) -> Value {
    if host == "omp" {
        return OMP_EVENTS
            .iter()
            .map(|(event, duty)| {
                json!({"event": event, "command": format!("{binary} hook-dispatch --host omp --event {duty}"),
                       "timeout": HOOK_TIMEOUT_SECS})
            })
            .collect();
    }
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

/// One planned hooks-file change (Codex, OMP): everything the digest binds and apply re-checks.
struct CodexState {
    file: PathBuf,
    receipt: PathBuf,
    before: Option<Vec<u8>>,
    receipt_before: Option<Vec<u8>>,
    after: Option<Vec<u8>>,
    rendered: Vec<u8>,
    digest: String,
    /// The lane Stop hook the rendered file carries (Codex only).
    lane: Option<Lane>,
    /// A file no receipt matches is replaced (`--replace-with-backup`): the apply keeps `backup`
    /// first. Never set without the flag (SPEC-20261120-hooks-conflict-choices).
    replaced: bool,
    backup: Option<crate::hosts::codex_config::Backup>,
}

fn codex_render(lane: Option<&Lane>) -> Result<Vec<u8>> {
    codex_render_with(HOOK_BINARY, lane)
}

fn codex_render_with(binary: &str, lane: Option<&Lane>) -> Result<Vec<u8>> {
    let doc =
        json!({"description": codex_description(binary), "hooks": codex_hooks_with(binary, lane)?});
    let mut bytes = serde_json::to_vec_pretty(&doc).map_err(|_| "encode failed")?;
    bytes.push(b'\n');
    Ok(bytes)
}

fn file_render(fh: &FileHost, lane: Option<&Lane>) -> Result<Vec<u8>> {
    if fh.host == "omp" {
        Ok(omp_module(HOOK_BINARY).into_bytes())
    } else {
        codex_render(lane)
    }
}

/// The sha256 values a hooks-file receipt owns: the recorded one and, after an interrupted
/// apply, the pending one.
fn codex_owned(receipt: &[u8], fh: &FileHost) -> Result<(Option<String>, Option<String>)> {
    let review = || format!("{} receipt requires preservation review", fh.file);
    let r: Value = serde_json::from_slice(receipt).map_err(|_| review())?;
    if r["schema_version"].as_u64() != Some(1) {
        return Err(review());
    }
    let sha = |k: &str| r[k].as_str().map(str::to_string);
    Ok((sha("sha256"), sha("pending_sha256")))
}

fn codex_state(root: &Path, fh: &FileHost, want: &LaneLoop, replace: bool) -> Result<CodexState> {
    let state_dir = crate::state_root::path(root, "");
    let mut dirs: Vec<PathBuf> = fh.dirs.iter().map(|d| root.join(d)).collect();
    dirs.extend([state_dir.clone(), state_dir.join("manifests")]);
    for dir in dirs {
        crate::claude::refuse_symlink_dir(&dir)?;
    }
    let (file, receipt) = (root.join(fh.file), state_dir.join(fh.receipt));
    let before = crate::claude::read_existing(&file)?;
    let receipt_before = crate::claude::read_existing(&receipt)?;
    let mut replaced = false;
    if let Some(bytes) = &before {
        // A file MoAI never wrote, or one changed since, is not overwritten.
        let now = Some(crate::digest(bytes));
        let owned = match &receipt_before {
            None => false,
            Some(r) => {
                let (sha, pending) = codex_owned(r, fh)?;
                now == sha || now == pending
            }
        };
        if !owned {
            // Only the flag lets a file MoAI does not own be replaced, and only after its copy.
            if !replace {
                return Err(format!("existing {} requires preservation review", fh.file));
            }
            crate::claude::refuse_symlink_dir(
                &state_dir.join(crate::hosts::codex_config::BACKUP_DIR),
            )
            .map_err(|_| {
                format!(
                    "the backups folder {} is a link; replacement refused",
                    crate::state_root::rel(root, crate::hosts::codex_config::BACKUP_DIR)
                )
            })?;
            replaced = true;
        }
    }
    let lane = match want {
        LaneLoop::Keep => match &receipt_before {
            Some(r) if fh.host == "codex" => receipt_lane(r, fh)?,
            _ => None,
        },
        LaneLoop::Set { lane, repo } => Lane::checked(lane, repo),
        LaneLoop::Off => None,
    };
    let rendered = file_render(fh, lane.as_ref())?;
    let after = (before.as_deref() != Some(rendered.as_slice())).then(|| rendered.clone());
    let (b, r, a) = (
        before.as_deref().map(crate::digest),
        receipt_before.as_deref().map(crate::digest),
        after.as_deref().map(crate::digest),
    );
    // The lane joins the digest only when set, so a plan without one keeps its old digest. A
    // replacement binds its mode and the file it replaces, so no other plan's digest applies it.
    let bound = match (&lane, replaced) {
        (None, false) => serde_json::to_vec(&(1, fh.host, b, r, a)),
        (Some(l), false) => serde_json::to_vec(&(1, fh.host, b, r, a, l.json())),
        (None, true) => {
            serde_json::to_vec(&(1, fh.host, b, r, a, "replace-with-backup", [fh.file]))
        }
        (Some(l), true) => serde_json::to_vec(&(
            1,
            fh.host,
            b,
            r,
            a,
            "replace-with-backup",
            [fh.file],
            l.json(),
        )),
    };
    let digest = crate::digest(&bound.map_err(|_| "encode failed")?);
    // The copy is named by the bytes it holds, like the Codex config's: the first 16 hex digits.
    let backup = match (&before, replaced) {
        (Some(bytes), true) => {
            let sha = crate::digest(bytes);
            let ext = Path::new(fh.file)
                .extension()
                .and_then(|e| e.to_str())
                .unwrap_or("bak");
            let name = format!(
                "{}/{}-hooks-{}.{ext}",
                crate::hosts::codex_config::BACKUP_DIR,
                fh.host,
                sha.trim_start_matches("sha256:").get(..16).unwrap_or("0")
            );
            Some(crate::hosts::codex_config::Backup {
                rel: crate::state_root::rel(root, &name),
                path: state_dir.join(&name),
                bytes: bytes.clone(),
            })
        }
        _ => None,
    };
    Ok(CodexState {
        file,
        receipt,
        before,
        receipt_before,
        after,
        rendered,
        digest,
        lane,
        replaced,
        backup,
    })
}

fn codex_receipt_bytes(
    fh: &FileHost,
    sha256: Option<&str>,
    pending: Option<&str>,
    lane: Option<&Lane>,
) -> Vec<u8> {
    let mut r = json!({
        "schema_version": 1,
        "hooks_path": fh.file,
        "sha256": sha256,
        "hooks": hook_list_for(fh.host, lane),
    });
    if let Some(p) = pending {
        r["pending_sha256"] = json!(p);
    }
    if let Some(l) = lane {
        r["lane_loop"] = l.json();
    }
    let mut bytes = serde_json::to_vec_pretty(&r).expect("receipt");
    bytes.push(b'\n');
    bytes
}

/// Previews generating the MoAI hook set for `host` in PROJECT. Never writes. The returned
/// `plan_digest` binds the current files and the result; `hooks-apply` requires it.
pub fn plan(project: &Path, host: &str, home: Option<&Path>) -> Result<Value> {
    // R1 of the removal audit: `moai hooks-plan --host omp` answers the skip notice too, so a
    // profile naming omp never turns the plain plan into a refusal.
    if host == "omp" {
        return plan_with_replace(project, host, home, LaneLoop::Keep, false);
    }
    plan_with(project, host, home, LaneLoop::Keep)
}

/// [`plan`] with the Codex lane Stop hook set, kept or removed (#2904 stage 2). `Keep` is what
/// the ownership receipt records; a plan with a lane reports it under `lane_loop`.
pub fn plan_with(project: &Path, host: &str, home: Option<&Path>, lane: LaneLoop) -> Result<Value> {
    plan_with_replace(project, host, home, lane, false)
}

/// The flag `--replace-with-backup` is for the hosts whose file MoAI writes whole.
fn check_replace(host: &str, replace: bool) -> Result<()> {
    if replace && host == "claude" {
        return Err("--replace-with-backup must be used with --host codex or omp".into());
    }
    Ok(())
}

/// The most lines a replacement's `diff` lists on each side, and the most characters of a line.
const DIFF_LINES: usize = 40;
const DIFF_LINE_CHARS: usize = 200;

/// Words that name a credential. A line holding one is withheld from the diff whatever follows it:
/// the diff shows a person's own file, where hiding a line costs a marker and the file is still
/// theirs to read, and a leaked key costs far more.
const CREDENTIAL_WORDS: &[&str] = &[
    "authorization",
    "bearer ",
    "secret",
    "apikey",
    "api-key",
    "passwd",
    "credential",
    "private_key",
    "access_key",
];

/// True for a line the diff must not print: the shapes [`crate::secret_shaped`] knows, the key
/// prefixes [`crate::redact::redact_outgoing`] knows (cloud, chat and gh tokens, `eyJ` tokens) and
/// the credential words above.
fn diff_line_withheld(line: &str) -> bool {
    let lower = line.to_ascii_lowercase();
    crate::secret_shaped(line)
        || crate::redact::redact_outgoing(line).1 > 0
        || CREDENTIAL_WORDS.iter().any(|w| lower.contains(w))
}

/// The two sides of a replacement as lines: the lines only the file has and the lines only the
/// planned file has, in file order, at most [`DIFF_LINES`] each; a line shaped like a secret is
/// withheld, because a hook command may carry one.
fn hook_diff(file: &str, current: &[u8], desired: &[u8]) -> Value {
    let (cur, des) = (
        String::from_utf8_lossy(current),
        String::from_utf8_lossy(desired),
    );
    let side = |from: &str, other: &str| -> (Vec<String>, bool) {
        let other: std::collections::HashSet<&str> = other.lines().collect();
        let mut lines = Vec::new();
        for line in from
            .lines()
            .filter(|l| !l.trim().is_empty() && !other.contains(l))
        {
            if lines.len() == DIFF_LINES {
                return (lines, true);
            }
            lines.push(if diff_line_withheld(line) {
                "[withheld: secret-shaped line]".to_string()
            } else {
                line.chars().take(DIFF_LINE_CHARS).collect()
            });
        }
        (lines, false)
    };
    let ((removed, cut_removed), (added, cut_added)) = (side(&cur, &des), side(&des, &cur));
    json!({
        "path": file,
        "current": {"sha256": crate::digest(current), "bytes": current.len()},
        "desired": {"sha256": crate::digest(desired), "bytes": desired.len()},
        "removed": removed,
        "added": added,
        "truncated": {"removed": cut_removed, "added": cut_added},
    })
}

/// [`plan_with`] that may replace a hook file no receipt matches (`replace`, Codex and OMP
/// only): the plan then reports the action `replace`, the file, the copy it will keep first, a
/// `diff` and a digest of its own. Without a file to replace the flag changes nothing.
pub fn plan_with_replace(
    project: &Path,
    host: &str,
    home: Option<&Path>,
    lane: LaneLoop,
    replace: bool,
) -> Result<Value> {
    // R1 of the removal audit: OMP support was removed, so the plan for omp is not an error
    // but a skip answer — a project whose profile lists omp keeps finishing its update, with
    // one notice, and nothing is written.
    if host == "omp" {
        return Ok(serde_json::json!({
            "host": "omp",
            "action": "skip",
            "notice": "omp support was removed; the omp hook entry is ignored and nothing is written",
            "path": null,
            "skip": true,
        }));
    }
    check_host(host)?;
    check_replace(host, replace)?;
    check_lane(host, &lane)?;
    let root = project_root(project, home)?;
    let mut planned_lane = None;
    // What the binary_migration preview reads: the bytes the plan digest binds.
    let before: Option<Vec<u8>>;
    let mut old_lane: Option<Lane> = None;
    let mut after_sha256 = None;
    let mut out = if host == "claude" {
        let mut v = crate::claude::hooks_settings_plan(&root, &claude_desired(&root)?)?;
        v["path"] = v["settings_path"].take();
        v.as_object_mut().map(|m| m.remove("settings_path"));
        before = crate::claude::read_existing(&root.join(host_config_path(host)))?;
        if before.as_deref().map(crate::digest) != v["expected_sha256"].as_str().map(str::to_string)
        {
            return Err("the settings file changed while the plan was made; plan again".into());
        }
        v
    } else {
        let fh = file_host(host);
        let s = codex_state(&root, fh, &lane, replace)?;
        planned_lane = s.lane.clone();
        before = s.before.clone();
        after_sha256 = s.after.as_deref().map(crate::digest);
        old_lane = match &s.receipt_before {
            Some(r) if fh.host == "codex" => receipt_lane(r, fh).ok().flatten(),
            _ => None,
        };
        let action = match (&s.after, &s.before) {
            _ if s.replaced => "replace",
            (None, _) => "unchanged",
            (Some(_), None) => "create",
            (Some(_), Some(_)) => "update",
        };
        let mut v = json!({
            "action": action,
            "path": fh.file,
            "conflicts": [],
            "expected_sha256": s.before.as_deref().map(crate::digest),
            "planned_sha256": s.after.as_deref().map(crate::digest),
            "plan_digest": s.digest,
            "writes": false,
        });
        if let (true, Some(file), Some(b)) = (s.replaced, &s.before, &s.backup) {
            // A file already holding MoAI's bytes is adopted: the plan still names them.
            v["planned_sha256"] = json!(crate::digest(&s.rendered));
            v["replaced"] = json!([fh.file]);
            v["backup"] = json!({
                "path": b.rel,
                "sha256": crate::digest(&b.bytes),
                "bytes": b.bytes.len(),
            });
            v["diff"] = hook_diff(fh.file, file, &s.rendered);
        }
        v
    };
    out["host"] = json!(host);
    out["hooks"] = hook_list_for(host, planned_lane.as_ref());
    if let Some(l) = &planned_lane {
        out["lane_loop"] = l.json();
    }
    // Outside the plan digest: what the hook engine covers on this host (#2391 stage 3).
    out["engine_capabilities"] = crate::hook_engine::capabilities(host);
    // Outside the plan digest too: which entries of the old `moai-cli` set the plan rewrites.
    let action = out["action"].as_str().unwrap_or("").to_string();
    out["binary_migration"] = crate::hook_legacy::migration(&crate::hook_legacy::Subject {
        host,
        file: host_config_path(host),
        before: before.as_deref(),
        action: &action,
        lane: old_lane
            .as_ref()
            .map(|l| (l.lane.as_str(), l.repo.as_str())),
        after_sha256,
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
    apply_with(project, host, home, plan_digest, permission, LaneLoop::Keep)
}

/// [`apply`] of a [`plan_with`] plan: the same `lane` the plan was made with.
pub fn apply_with(
    project: &Path,
    host: &str,
    home: Option<&Path>,
    plan_digest: &str,
    permission: &str,
    lane: LaneLoop,
) -> Result<Value> {
    apply_with_replace(project, host, home, plan_digest, permission, lane, false)
}

/// [`apply_with`] of a [`plan_with_replace`] plan: the same `lane` and the same `replace` the
/// plan was made with. A replacement keeps the copy of the file first, then writes the pending
/// receipt, the file and the final receipt.
pub fn apply_with_replace(
    project: &Path,
    host: &str,
    home: Option<&Path>,
    plan_digest: &str,
    permission: &str,
    lane: LaneLoop,
    replace: bool,
) -> Result<Value> {
    check_host(host)?;
    check_replace(host, replace)?;
    check_lane(host, &lane)?;
    check_permission(permission)?;
    let root = project_root(project, home)?;
    let mut applied_lane = None;
    let mut out = if host == "claude" {
        let mut v = crate::claude::hooks_settings_apply(
            &root,
            &claude_desired(&root)?,
            plan_digest,
            permission,
        )?;
        v["receipt_path"] = json!(crate::state_root::rel(
            &root,
            "manifests/claude-settings.json"
        ));
        v
    } else {
        let fh = file_host(host);
        let s = codex_state(&root, fh, &lane, replace)?;
        if s.digest != plan_digest {
            return Err("stale or altered plan; re-preview required".into());
        }
        applied_lane = s.lane.clone();
        // The copy comes first: whatever stops the apply later, the user's bytes are kept.
        if let Some(backup) = &s.backup {
            crate::hosts::codex_config::write_backup(backup)?;
        }
        let rel = crate::state_root::rel(&root, fh.receipt);
        let rendered_sha = crate::digest(&s.rendered);
        let final_receipt = codex_receipt_bytes(fh, Some(&rendered_sha), None, s.lane.as_ref());
        let mut receipt_now = s.receipt_before.clone();
        let mut files = Vec::new();
        if let Some(after) = &s.after {
            let before_sha = s.before.as_deref().map(crate::digest);
            // The pending receipt already names the new lane: a re-plan after an interrupted
            // apply (Keep) renders the file this apply was writing.
            let pending = codex_receipt_bytes(
                fh,
                before_sha.as_deref(),
                Some(&rendered_sha),
                s.lane.as_ref(),
            );
            crate::agents::write_replace(&s.receipt, receipt_now.as_deref(), &pending)?;
            receipt_now = Some(pending);
            crate::agents::write_replace(&s.file, s.before.as_deref(), after)?;
            files.push(fh.file);
        }
        if receipt_now.as_deref() != Some(final_receipt.as_slice()) {
            if let Err(e) =
                crate::agents::write_replace(&s.receipt, receipt_now.as_deref(), &final_receipt)
            {
                return Err(if files.is_empty() {
                    e
                } else {
                    format!("partial apply: applied [{}]; failed at {rel}: {e}", fh.file)
                });
            }
        }
        let mut v = json!({
            "state": if files.is_empty() && !s.replaced { "unchanged" } else { "applied" },
            "files_changed": files,
            "files_created": if s.before.is_none() { files.clone() } else { vec![] },
            "hooks_sha256_before": s.before.as_deref().map(crate::digest),
            "hooks_sha256_after": rendered_sha,
            "receipt_path": rel,
            "host_loaded": "not_run",
            "execution_verified": "not_run",
        });
        if let Some(b) = &s.backup {
            v["backup"] = json!({
                "path": b.rel,
                "sha256": crate::digest(&b.bytes),
                "bytes": b.bytes.len(),
            });
        }
        v
    };
    out["host"] = json!(host);
    out["hooks_generated"] = hook_list_for(host, applied_lane.as_ref());
    if let Some(l) = &applied_lane {
        out["lane_loop"] = l.json();
    }
    out["trust_required"] = json!(trust_notice(host));
    Ok(out)
}

/// What the user must do before `host` runs the generated project hooks. Printed by
/// `hooks-apply` and by the `project-update-plan` host steps.
pub fn trust_notice(host: &str) -> &'static str {
    match host {
        "claude" => "Claude Code runs project hooks only in a trusted workspace",
        "omp" => "OMP loads project extensions from .omp/extensions with no trust prompt in its docs (docs/extension-loading.md, read 2026-11-13)",
        _ => "Codex runs project hooks only when the project .codex/ layer is trusted and each hook is reviewed (/hooks)",
    }
}

/// Where each host keeps the project hook definitions MoAI generates.
fn host_config_path(host: &str) -> &'static str {
    if host == "claude" {
        ".claude/settings.json"
    } else {
        file_host(host).file
    }
}

/// Whether MoAI's hook set is installed for `host` in the project `root` (#2853): Claude Code
/// when its settings receipt owns `hooks`, Codex and OMP when their ownership receipt exists.
/// An unreadable receipt is an error for the caller to report, never "not installed".
pub fn installed(root: &Path, host: &str) -> Result<bool> {
    check_host(host)?;
    if host == "claude" {
        return Ok(crate::claude::owned_settings(root)?.contains_key("hooks"));
    }
    let receipt = crate::state_root::path(root, file_host(host).receipt);
    match fs::symlink_metadata(&receipt) {
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(false),
        Err(_) => Err(format!("{} unavailable", receipt.display())),
        Ok(m) if m.file_type().is_symlink() => Err("symlink receipt refused".into()),
        Ok(_) => Ok(true),
    }
}

/// The two files a hook apply for `host` writes in the project `root`, relative to it: the host
/// file and MoAI's ownership receipt for it (#2853 revert). Each folder between `root` and
/// these files is listed in `dirs`, so a caller can refuse one that became a link.
pub fn written_files(root: &Path, host: &str) -> Result<(Vec<String>, Vec<PathBuf>)> {
    check_host_written(host)?;
    let receipt = if host == "claude" {
        "manifests/claude-settings.json"
    } else {
        file_host(host).receipt
    };
    let files = vec![
        host_config_path(host).to_string(),
        crate::state_root::rel(root, receipt),
    ];
    let host_dirs: &[&str] = if host == "claude" {
        &[".claude"]
    } else {
        file_host(host).dirs
    };
    let state = crate::state_root::path(root, "");
    let mut dirs: Vec<PathBuf> = host_dirs.iter().map(|d| root.join(d)).collect();
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
///   from a settings file, and Codex records trust against each hook's hash, but neither page
///   says where that is stored (read 2026-11-10), so it is never read here. No local line
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
    // R1 of the removal audit: the plan command asks for the host's status before planning,
    // so the retired host answers the skip notice here too instead of a refusal.
    if host == "omp" {
        return Ok(serde_json::json!({
            "host": "omp",
            "installed": false,
            "skip": true,
            "notice": "omp support was removed; the omp hook entry is ignored and nothing is written",
        }));
    }
    check_host(host)?;
    let root = project_root(project, home)?;
    // Compared directly, not through the plan: the Claude settings plan preserves a user's
    // edit to an owned key and then reports `unchanged`, which is not "configured as generated".
    // A file with no `hooks` key at all is planned again there (C-55), never `unchanged`.
    // Codex: the set its receipt records, so an installed lane Stop hook is `yes` (#2904). An
    // unreadable or malformed receipt compares with the base set.
    let lane = if host == "codex" {
        let receipt = crate::claude::read_existing(&crate::state_root::path(&root, CODEX_RECEIPT))
            .ok()
            .flatten();
        receipt.and_then(|r| receipt_lane(&r, &CODEX_HOST).ok().flatten())
    } else {
        None
    };
    let generated = Value::Object(if host == "codex" {
        codex_hooks(lane.as_ref())?
    } else {
        hooks_object(host)?
    });
    // The set the releases before #3004 P4-2 wrote: the same generators, the old binary name.
    let legacy = Value::Object(if host == "codex" {
        codex_hooks_with(LEGACY_BINARY, lane.as_ref())?
    } else {
        hooks_object_with(LEGACY_BINARY, host)?
    });
    let configured = match crate::claude::read_existing(&root.join(host_config_path(host)))? {
        None => "no",
        // The OMP module is code, not a hooks object: configured means byte-equal (#2633).
        Some(bytes) if host == "omp" => {
            if bytes == omp_module(HOOK_BINARY).as_bytes() {
                "yes"
            } else if bytes == omp_module(LEGACY_BINARY).as_bytes() {
                "legacy_binary"
            } else {
                "drifted"
            }
        }
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

    /// The bytes of a Codex install without a lane: the hook file, its receipt and a fresh
    /// project's plan digest. Captured from the code before #2904 stage 2 (a869647cb7) and
    /// captured again for the three-event set of #2936 R9 S1 (no `Stop`, `SessionStart` =
    /// session-check), so a later lane change still cannot move a no-lane install's bytes.
    /// Captured again on purpose for #3004 P4-2, when the generated command changed from
    /// `moai-cli` to `moai`: the three digests moved with it, the `LEGACY_*` ones below are the
    /// values before that change.
    const BASE_CODEX_SHA: &str =
        "sha256:746afebf5441f58f6e32a1c8080f2487ae98004e445865423ce3df03b91b2b44";
    const BASE_RECEIPT_SHA: &str =
        "sha256:6934026e4a437d05f40500bd3daab68d77e542cc5b934cf598b729fffa9ac34d";
    const BASE_PLAN_DIGEST: &str =
        "sha256:0af293e201ad53ad56e4e089db87540be920fe27f61db3b53d6f9b6e485a018d";
    /// The file digest of the Codex install before #3004 P4-2 (`moai-cli`); the legacy generator
    /// must still produce exactly these bytes, or the rewrite would miss every installed set.
    const LEGACY_CODEX_SHA: &str =
        "sha256:7539fda0dbab97970be03160c4c5f00ab7fcb0b406e5bad6fb6397fb5b9ebc8a";
    /// The OMP module before #3004 P4-2.
    const LEGACY_OMP_SHA: &str =
        "sha256:c12accc6f8d97d1f556f7e100f95bcd991db89f60443e75c231d0ba94433c794";

    #[test]
    fn a_codex_install_without_a_lane_keeps_its_bytes() {
        let file = codex_render(None).unwrap();
        assert_eq!(crate::digest(&file), BASE_CODEX_SHA);
        let receipt = codex_receipt_bytes(&CODEX_HOST, Some(&crate::digest(&file)), None, None);
        assert_eq!(crate::digest(&receipt), BASE_RECEIPT_SHA);
        let dir = tempfile::tempdir().unwrap();
        for lane in [LaneLoop::Keep, LaneLoop::Off] {
            let v = plan_with(dir.path(), "codex", None, lane).unwrap();
            assert_eq!(v["plan_digest"], json!(BASE_PLAN_DIGEST));
            assert!(v.get("lane_loop").is_none(), "{v}");
        }
    }

    /// The legacy generators are the current ones called with the old name: they rebuild the
    /// bytes of the releases before #3004 P4-2 exactly.
    #[test]
    fn the_legacy_generators_rebuild_the_pre_p4_bytes() {
        let codex = legacy_codex_file(None).unwrap();
        assert_eq!(crate::digest(&codex), LEGACY_CODEX_SHA);
        assert_eq!(
            crate::digest(omp_module(LEGACY_BINARY).as_bytes()),
            LEGACY_OMP_SHA
        );
        assert_ne!(omp_module(LEGACY_BINARY), omp_module(HOOK_BINARY));
        for text in [String::from_utf8(codex).unwrap(), omp_module(LEGACY_BINARY)] {
            assert!(text.contains("moai-cli"), "{text}");
        }
        // Every generated command, in every host and for the lane Stop hook, names `moai`, and
        // the legacy set names `moai-cli` in the same places.
        for host in HOSTS {
            for (event, groups) in current_hooks(host, Some(("agent-1", "o/r"))).unwrap() {
                let old = &legacy_hooks(host, Some(("agent-1", "o/r"))).unwrap()[&event];
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

    fn set(lane: &str, repo: &str) -> LaneLoop {
        LaneLoop::Set {
            lane: lane.into(),
            repo: repo.into(),
        }
    }

    fn install(root: &Path, lane: LaneLoop) -> Value {
        let v = plan_with(root, "codex", None, lane.clone()).unwrap();
        let digest = v["plan_digest"].as_str().unwrap().to_string();
        apply_with(root, "codex", None, &digest, "test-permission", lane).unwrap()
    }

    #[test]
    fn a_lane_adds_a_second_stop_group_that_keep_preserves_and_off_removes() {
        let dir = tempfile::tempdir().unwrap();
        let root = dir.path();
        install(root, LaneLoop::Keep);
        let file = root.join(CODEX_FILE);
        let base = fs::read(&file).unwrap();
        let receipt = crate::state_root::path(root, CODEX_RECEIPT);
        let base_receipt = fs::read(&receipt).unwrap();
        // The lane's plan differs from the base plan only by the lane.
        let with = plan_with(root, "codex", None, set("agent-1", "o/r")).unwrap();
        assert_eq!(with["action"], json!("update"));
        assert_eq!(with["lane_loop"], json!({"lane": "agent-1", "repo": "o/r"}));
        assert_eq!(fs::read(&file).unwrap(), base, "a plan writes nothing");
        install(root, set("agent-1", "o/r"));
        let v: Value = serde_json::from_slice(&fs::read(&file).unwrap()).unwrap();
        let stop = v["hooks"]["Stop"].as_array().unwrap();
        // The base set has no `Stop` event: the lane group is created and is the only one.
        assert_eq!(stop.len(), 1);
        assert_eq!(
            stop[0],
            json!({"hooks": [{"type": "command",
                "command": "moai lane-stop-hook --lane agent-1 --repo o/r", "timeout": 120}]})
        );
        let r: Value = serde_json::from_slice(&fs::read(&receipt).unwrap()).unwrap();
        assert_eq!(r["lane_loop"], json!({"lane": "agent-1", "repo": "o/r"}));
        assert_eq!(r["hooks"].as_array().unwrap().len(), 4);
        // Keep re-plans to the same file; the status reads the receipt's set.
        let keep = plan(root, "codex", None).unwrap();
        assert_eq!(keep["action"], json!("unchanged"), "{keep}");
        assert_eq!(keep["lane_loop"]["lane"], json!("agent-1"));
        let st = status(root, "codex", None, "").unwrap();
        assert_eq!(st["configured"], json!("yes"), "{st}");
        // Off returns both files to the base bytes.
        install(root, LaneLoop::Off);
        assert_eq!(fs::read(&file).unwrap(), base);
        assert_eq!(fs::read(&receipt).unwrap(), base_receipt);
        assert_eq!(
            plan(root, "codex", None).unwrap()["action"],
            json!("unchanged")
        );
    }

    #[test]
    fn a_lane_is_codex_only_and_well_formed() {
        let dir = tempfile::tempdir().unwrap();
        for host in ["claude"] {
            for lane in [set("agent-1", "o/r"), LaneLoop::Off] {
                let e = plan_with(dir.path(), host, None, lane).unwrap_err();
                assert!(e.contains("--host codex only"), "{e}");
            }
        }
        for (lane, repo) in [
            ("agent 1", "o/r"),
            ("", "o/r"),
            ("a", "o"),
            ("a", "o/r/x"),
            // A leading dash would read as a flag in the generated command (AUDIT #2917 N1).
            ("-a", "o/r"),
            ("a", "-o/r"),
        ] {
            let e = plan_with(dir.path(), "codex", None, set(lane, repo)).unwrap_err();
            assert!(e.contains("OWNER/REPO"), "{e}");
        }
        // A receipt whose lane_loop is malformed stops for review instead of dropping it.
        install(dir.path(), set("agent-1", "o/r"));
        let receipt = crate::state_root::path(dir.path(), CODEX_RECEIPT);
        let mut r: Value = serde_json::from_slice(&fs::read(&receipt).unwrap()).unwrap();
        r["lane_loop"]["repo"] = json!("no-slash");
        fs::write(&receipt, serde_json::to_vec(&r).unwrap()).unwrap();
        let e = plan(dir.path(), "codex", None).unwrap_err();
        assert!(e.contains("preservation review"), "{e}");
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
