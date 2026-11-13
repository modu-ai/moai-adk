//! Everything the statusline reads outside stdin: environment, the project configuration
//! and state files, the HOME state files, and local git. Read-only (#2350): the Go
//! statusline also writes `$HOME/.moai/state/last-model.txt` and
//! `<anchor>/.moai/state/context-usage/*.json`; this module writes nothing. The one write
//! of the statusline is the forge counts cache, done by the detached refresh child in
//! [`super::counts`] (#2588), never by the render path.
//!
//! HOME state (#2455, user decision 2026-11-11): `last-model.txt` is read from
//! `$HOME/.moai-cli/state/` only, never from the Go folder `$HOME/.moai/`, and `MOAI_HOME` is
//! not consulted. No Rust command writes it yet, so "nothing to read" is the normal state. The
//! session task (`last-session-state.json`, 📋) is not read at all since the three-line layout
//! (#2376, user decision 3: the task is never shown).
//!
//! The only subprocess is `git`, and only when a `.git` entry is found above the
//! directory Go would open; it runs read-only (`--no-optional-locks`) under the render
//! deadline. No network here: the counts refresh child is started by `statusline::run`
//! after the render (#2588).

use super::config::{self, Node};
use super::input::{self, Input};
use super::segments::{self, ContextReading, GitStatus, Memory};
use serde::Deserialize;
use std::collections::{BTreeMap, HashMap};
use std::io::Read;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use std::time::{Duration, Instant};

/// The process environment, captured once so collection is testable.
#[derive(Debug, Clone, Default)]
pub struct Context {
    vars: HashMap<String, String>,
    pub cwd: Option<PathBuf>,
    /// When the render must be finished; git is stopped at this point.
    pub deadline: Option<Instant>,
    /// The git executable; `git` from PATH outside tests.
    pub git_program: String,
    /// The current time in Unix seconds, for the cache timer (#2756 AC-8). The hidden
    /// `--now-unix` seam fixes it so goldens do not depend on the wall clock.
    pub now: i64,
}

impl Context {
    pub fn from_process(budget: Duration) -> Context {
        Context {
            vars: std::env::vars().collect(),
            cwd: std::env::current_dir().ok(),
            deadline: Some(Instant::now() + budget),
            git_program: "git".to_string(),
            now: std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .map(|d| d.as_secs() as i64)
                .unwrap_or(0),
        }
    }

    #[cfg(test)]
    pub fn for_test(vars: &[(&str, &str)], cwd: Option<&Path>) -> Context {
        Context {
            vars: vars
                .iter()
                .map(|(k, v)| (k.to_string(), v.to_string()))
                .collect(),
            cwd: cwd.map(Path::to_path_buf),
            deadline: None,
            git_program: "git".to_string(),
            now: 0,
        }
    }

    /// A variable as Go's `os.Getenv(k) != ""` sees it: unset and empty are both `None`.
    pub fn var(&self, key: &str) -> Option<&str> {
        self.vars
            .get(key)
            .map(String::as_str)
            .filter(|v| !v.is_empty())
    }

    fn home(&self) -> Option<PathBuf> {
        self.var("HOME").map(PathBuf::from)
    }

    /// The HOME state folder `$HOME/.moai-cli/state` (Issue #2455, user decision
    /// 2026-11-11). The Go folder `$HOME/.moai` is never read, and `MOAI_HOME` — the Go
    /// tool's variable — is not consulted, so a Go setting cannot point back at it.
    fn home_state(&self) -> Option<PathBuf> {
        self.home()
            .map(|h| h.join(crate::HOME_STATE_DIR).join("state"))
    }

    pub fn no_color(&self) -> bool {
        self.var("NO_COLOR").is_some() || self.var("MOAI_NO_COLOR").is_some()
    }
}

/// Everything one render needs (Go `StatusData`, the fields this port renders).
#[derive(Debug, Default)]
pub struct Status {
    pub model: String,
    pub session_ms: i64,
    pub memory: Memory,
    pub auto_compact: i64,
    pub version: Option<Version>,
    pub claude_version: String,
    pub output_style: String,
    pub session_name: String,
    /// `Some` when `effort` is present, even with an empty level.
    pub effort: Option<String>,
    pub thinking: Option<bool>,
    /// (cache read, cache creation) from `current_usage`.
    pub cache: Option<(i64, i64)>,
    pub five_hour: Option<input::RateWindow>,
    pub seven_day: Option<input::RateWindow>,
    pub git: Option<GitStatus>,
    pub repo: Option<(String, String)>,
    pub github: GitHubCounts,
    pub goal_armed: bool,
    /// #2756 AC-7: the open pull request of the current branch.
    pub pr: Option<PrInfo>,
    /// #2756 AC-8: how long the prompt cache stays warm.
    pub cache_timer: Option<CacheTimer>,
    /// #2756 AC-9: the worktree name; empty in the main checkout.
    pub worktree: String,
}

/// The `pr` fields the status line shows.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct PrInfo {
    pub number: i64,
    /// Empty when absent or not a plain `http(s)` URL; the number is then shown unlinked.
    pub url: String,
    /// Empty when absent.
    pub review_state: String,
}

/// The prompt cache state from `prompt_cache` (Unix-second timestamps).
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum CacheTimer {
    /// Warm, with whole minutes left until `expires_at` (rounded down).
    Warm(i64),
    /// `warm` is false, or `expires_at` has passed.
    Cold,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Version {
    pub current: String,
    pub latest: Option<String>,
}

#[derive(Debug, Default, Clone, Copy, PartialEq, Eq)]
pub struct GitHubCounts {
    pub available: bool,
    pub suppressed: bool,
    pub open_issues: i64,
    pub open_prs: i64,
}

/// `statusline.yaml` as the Go CLI reads it: theme and the segments map.
#[derive(Debug, Default, Clone)]
pub struct DisplayConfig {
    pub theme: String,
    pub segments: BTreeMap<String, bool>,
}

/// Version this binary reports in the `🗿` segment.
pub const BINARY_VERSION: &str = crate::BUILD_VERSION;

pub fn collect(input: Option<&Input>, ctx: &Context) -> Status {
    let project_root = find_project_root(ctx);
    let anchor = input.and_then(resolve_anchor);
    let auto_compact = segments::auto_compact_threshold(ctx.var("CLAUDE_AUTOCOMPACT_PCT_OVERRIDE"));
    let mut status = Status {
        model: collect_model(input, ctx),
        auto_compact,
        version: Some(collect_version(ctx, BINARY_VERSION)),
        ..Default::default()
    };
    let git_dir = project_root.clone().or_else(|| ctx.cwd.clone());
    status.git = git_dir.and_then(|d| collect_git(&d, ctx));
    let Some(input) = input else {
        return status;
    };
    status.session_ms = input
        .cost
        .as_ref()
        .and_then(|c| c.total_duration_ms)
        .unwrap_or(0);
    status.memory = collect_memory(input, ctx, auto_compact);
    status.session_name = input::text(&input.session_name).to_string();
    status.output_style = input
        .output_style
        .as_ref()
        .map_or("", |s| input::text(&s.name))
        .to_string();
    status.claude_version = input::text(&input.version).to_string();
    status.effort = input
        .effort
        .as_ref()
        .map(|e| input::text(&e.level).to_string());
    status.thinking = input.thinking.as_ref().map(|t| t.enabled.unwrap_or(false));
    status.cache = input
        .context_window
        .as_ref()
        .and_then(|cw| cw.current_usage.as_ref())
        .map(|u| {
            (
                u.cache_read_input_tokens.unwrap_or(0),
                u.cache_creation_input_tokens.unwrap_or(0),
            )
        });
    if let Some(limits) = &input.rate_limits {
        status.five_hour = limits.five_hour;
        status.seven_day = limits.seven_day;
    }
    status.repo = input
        .workspace
        .as_ref()
        .and_then(|w| w.repo.as_ref())
        .map(|r| {
            (
                input::text(&r.owner).to_string(),
                input::text(&r.name).to_string(),
            )
        });
    if let Some(anchor) = &anchor {
        status.github = github_counts(anchor);
        status.goal_armed = goal_armed(anchor, input::text(&input.session_id));
    }
    status.pr = input.pr.as_ref().and_then(|p| {
        Some(PrInfo {
            number: p.number.filter(|n| *n > 0)?,
            url: safe_link(input::text(&p.url)).to_string(),
            review_state: input::text(&p.review_state).to_string(),
        })
    });
    status.cache_timer = input
        .prompt_cache
        .as_ref()
        .and_then(|c| cache_timer(c, ctx.now));
    let worktree = input.worktree.as_ref().map_or("", |w| input::text(&w.name));
    status.worktree = if worktree.is_empty() {
        input
            .workspace
            .as_ref()
            .map_or("", |w| input::text(&w.git_worktree))
            .to_string()
    } else {
        worktree.to_string()
    };
    status
}

/// #2756 AC-8: `warm: false` is cold; `warm: true` with `expires_at` gives the whole minutes
/// left (a passed `expires_at` is cold). Anything else — no `warm`, a warm cache without
/// `expires_at`, a value of another type — shows no timer.
fn cache_timer(cache: &serde_json::Value, now: i64) -> Option<CacheTimer> {
    match cache.get("warm").and_then(serde_json::Value::as_bool)? {
        false => Some(CacheTimer::Cold),
        true => {
            // The decoder may hand numbers over as floats (Go `encoding/json` rules).
            let expires = cache.get("expires_at").and_then(|v| {
                v.as_i64()
                    .or_else(|| v.as_f64().filter(|f| f.is_finite()).map(|f| f as i64))
            })?;
            let left = expires - now;
            Some(if left <= 0 {
                CacheTimer::Cold
            } else {
                CacheTimer::Warm(left / 60)
            })
        }
    }
}

/// #2756 AC-7: a URL usable inside an OSC 8 hyperlink — `http://` or `https://` and no
/// control character, space or non-ASCII byte (which could end the escape sequence early or
/// smuggle another one). Anything else is dropped and the number is shown unlinked.
fn safe_link(url: &str) -> &str {
    let scheme = url.starts_with("https://") || url.starts_with("http://");
    let plain = url.bytes().all(|b| b.is_ascii_graphic());
    if scheme && plain && url.len() <= 2048 {
        url
    } else {
        ""
    }
}

/// The configuration root: the state anchor, else the cwd project root (cli/statusline.go).
pub fn display_config(input: Option<&Input>, ctx: &Context) -> DisplayConfig {
    let root = input
        .and_then(resolve_anchor)
        .or_else(|| find_project_root(ctx));
    root.map(|r| read_display_config(&config_path(&r, STATUSLINE_YAML)))
        .unwrap_or_default()
}

/// `statusline.yaml`: the RC location under the project's state folder, then the Go file.
const STATUSLINE_YAML: (&str, &str) = (
    "config/sections/statusline.yaml",
    ".moai/config/sections/statusline.yaml",
);
/// `statusline.local.yaml`: the RC location under the project's state folder, then the Go file.
const STATUSLINE_LOCAL_YAML: (&str, &str) =
    ("statusline.local.yaml", ".moai/statusline.local.yaml");

/// #2350 scope 3 (user decision 6774899518 row 42 = B): a statusline setting is read from
/// `.moai-cli/` (the Go relative path under the RC folder) when that file exists, else
/// from the Go file. The Go file is only read — never written, moved or copied — so the
/// Go state stays as it is. Existence decides, not parsing: an unreadable `.moai-cli/`
/// file means defaults (as Go treats an unparseable file), never a silent return to the
/// Go values.
fn config_path(root: &Path, (rc, go): (&str, &str)) -> PathBuf {
    // #2725 S2b: `rc` is relative to the project's state folder (`.moai-cli`, or `.moai`).
    let rc = crate::state_root::path(root, rc);
    if rc.exists() {
        rc
    } else {
        root.join(go)
    }
}

fn read_display_config(path: &Path) -> DisplayConfig {
    let Some(doc) = config::read(path) else {
        return DisplayConfig::default();
    };
    let theme = match doc.get(&["statusline", "theme"]) {
        None => String::new(),
        Some(Node::Scalar(s)) => s.clone(),
        Some(Node::Map(_)) => return DisplayConfig::default(),
    };
    let mut segments = BTreeMap::new();
    match doc.get(&["statusline", "segments"]) {
        None => {}
        Some(Node::Scalar(s)) if s.is_empty() || s == "null" || s == "~" => {}
        Some(Node::Scalar(_)) => return DisplayConfig::default(),
        Some(Node::Map(map)) => {
            for (key, value) in map {
                let Node::Scalar(v) = value else {
                    return DisplayConfig::default();
                };
                let Some(b) = config::boolean(v) else {
                    return DisplayConfig::default();
                };
                segments.insert(key.clone(), b);
            }
        }
    }
    DisplayConfig { theme, segments }
}

/// `findProjectRoot` (cli/glm.go): walk up from the cwd to a directory holding `.moai`,
/// stopping at HOME.
fn find_project_root(ctx: &Context) -> Option<PathBuf> {
    let mut dir = ctx.cwd.clone()?;
    if let Ok(resolved) = std::fs::canonicalize(&dir) {
        dir = resolved;
    }
    let home = ctx.home().map(|h| std::fs::canonicalize(&h).unwrap_or(h));
    loop {
        if home.as_deref() == Some(dir.as_path()) {
            return None;
        }
        if dir.join(".moai").exists() {
            return Some(dir);
        }
        if !dir.pop() {
            return None;
        }
    }
}

/// `stateanchor.Resolve`: an existing absolute `workspace.project_dir`, else
/// `worktree.original_cwd`, else the main checkout of the git repository holding
/// `workspace.current_dir` / `cwd`, found on disk instead of by running git.
pub fn resolve_anchor(input: &Input) -> Option<PathBuf> {
    let candidate = |p: &str| {
        let path = PathBuf::from(p);
        (!p.is_empty() && path.is_absolute() && path.is_dir()).then_some(path)
    };
    let ws = input.workspace.as_ref();
    if let Some(p) = ws.and_then(|w| candidate(input::text(&w.project_dir))) {
        return Some(p);
    }
    if let Some(p) = input
        .worktree
        .as_ref()
        .and_then(|w| candidate(input::text(&w.original_cwd)))
    {
        return Some(p);
    }
    let dir = ws.map_or("", |w| input::text(&w.current_dir));
    let dir = if dir.is_empty() {
        input::text(&input.cwd)
    } else {
        dir
    };
    if dir.is_empty() {
        return None;
    }
    main_checkout_of(Path::new(dir))
}

/// The parent of the git common directory for `dir`: `.git` as a directory, or a `.git`
/// file naming a worktree git dir whose `commondir` points back to the main repository.
fn main_checkout_of(dir: &Path) -> Option<PathBuf> {
    let dir = std::fs::canonicalize(dir).ok()?;
    let (top, dot_git) = find_dot_git(&dir)?;
    if dot_git.is_dir() {
        return Some(top);
    }
    let text = std::fs::read_to_string(&dot_git).ok()?;
    let git_dir = PathBuf::from(text.trim().strip_prefix("gitdir:")?.trim());
    let git_dir = if git_dir.is_absolute() {
        git_dir
    } else {
        top.join(git_dir)
    };
    let common = match std::fs::read_to_string(git_dir.join("commondir")) {
        Ok(c) => {
            let c = PathBuf::from(c.trim());
            if c.is_absolute() {
                c
            } else {
                git_dir.join(c)
            }
        }
        Err(_) => git_dir,
    };
    std::fs::canonicalize(common)
        .ok()?
        .parent()
        .map(Path::to_path_buf)
}

pub(crate) fn find_dot_git(start: &Path) -> Option<(PathBuf, PathBuf)> {
    let mut dir = start.to_path_buf();
    loop {
        let candidate = dir.join(".git");
        if candidate.exists() {
            return Some((dir, candidate));
        }
        if !dir.pop() {
            return None;
        }
    }
}

// ---- model ----

/// `CollectMetrics` model resolution: display name, shortened id or name, then the
/// `MOAI_LAST_MODEL` / `$HOME/.moai-cli/state/last-model.txt` fallback (Go read
/// `$HOME/.moai/state/`; #2455 moved it), then the GLM mapping.
fn collect_model(input: Option<&Input>, ctx: &Context) -> String {
    let mut name = String::new();
    if let Some(m) = input.and_then(|i| i.model.as_ref()) {
        name = if !m.display_name.is_empty() {
            m.display_name.clone()
        } else if !m.id.is_empty() {
            segments::shorten_model_name(&m.id)
        } else {
            segments::shorten_model_name(&m.name)
        };
    }
    if name.is_empty() {
        name = if let Some(env) = ctx.var("MOAI_LAST_MODEL") {
            segments::shorten_model_name(env)
        } else {
            ctx.home_state()
                .and_then(|s| std::fs::read_to_string(s.join("last-model.txt")).ok())
                .map(|s| s.trim().to_string())
                .unwrap_or_default()
        };
    }
    resolve_glm_model_name(&name, ctx)
}

/// `resolveGLMModelName`: a Claude slot name maps to the GLM model configured for it.
fn resolve_glm_model_name(display: &str, ctx: &Context) -> String {
    if display.is_empty() {
        return String::new();
    }
    let cleaned = display.strip_suffix("[1m]").unwrap_or(display).trim();
    let lower = cleaned.to_lowercase();
    let key = if lower.contains("opus") {
        "ANTHROPIC_DEFAULT_OPUS_MODEL"
    } else if lower.contains("sonnet") {
        "ANTHROPIC_DEFAULT_SONNET_MODEL"
    } else if lower.contains("haiku") {
        "ANTHROPIC_DEFAULT_HAIKU_MODEL"
    } else if !lower.starts_with("claude") {
        return cleaned.to_string();
    } else {
        return display.to_string();
    };
    match ctx.var(key) {
        Some(glm) if !glm.starts_with("claude-") => glm.to_string(),
        _ => display.to_string(),
    }
}

// ---- context window ----

fn collect_memory(input: &Input, ctx: &Context, auto_compact: i64) -> Memory {
    let reading = input.context_window.as_ref().map(|cw| ContextReading {
        used_percentage: cw.used_percentage,
        context_window_size: cw.context_window_size.unwrap_or(0),
        used: cw.used.unwrap_or(0),
        total: cw.total.unwrap_or(0),
        current_usage_tokens: cw.current_usage.as_ref().map(|u| {
            u.input_tokens.unwrap_or(0)
                + u.cache_creation_input_tokens.unwrap_or(0)
                + u.cache_read_input_tokens.unwrap_or(0)
        }),
    });
    let size_override = if reading.is_some() {
        context_window_override(ctx)
    } else {
        0
    };
    segments::collect_memory(reading, size_override, auto_compact)
}

/// Built-in GLM context windows (memory.go `glmContextWindows`).
const GLM_CONTEXT_WINDOWS: [(&str, i64); 2] =
    [("glm-5.3-flash", 1_000_000), ("glm-5.3", 1_000_000)];
const GLM_ENV_SLOTS: [&str; 3] = [
    "ANTHROPIC_DEFAULT_OPUS_MODEL",
    "ANTHROPIC_DEFAULT_SONNET_MODEL",
    "ANTHROPIC_DEFAULT_HAIKU_MODEL",
];

/// `resolveContextWindowOverride`: `MOAI_STATUSLINE_CONTEXT_SIZE`, then `llm.yaml`
/// `llm.glm.context_windows`, then the built-in table, for non-Claude slot models.
fn context_window_override(ctx: &Context) -> i64 {
    if let Some(n) = ctx
        .var("MOAI_STATUSLINE_CONTEXT_SIZE")
        .and_then(|v| v.parse::<i64>().ok())
    {
        if n > 0 {
            return n;
        }
    }
    let slots: Vec<String> = GLM_ENV_SLOTS
        .iter()
        .filter_map(|k| ctx.vars.get(*k))
        .map(|v| v.trim().to_lowercase())
        .filter(|v| !v.is_empty() && !v.starts_with("claude"))
        .collect();
    let user = llm_context_windows(ctx);
    for table in [
        user.as_slice(),
        &GLM_CONTEXT_WINDOWS.map(|(k, v)| (k.to_string(), v))[..],
    ] {
        for model in &slots {
            if let Some(size) = match_context_window(model, table) {
                return size;
            }
        }
    }
    0
}

fn match_context_window(model: &str, table: &[(String, i64)]) -> Option<i64> {
    table
        .iter()
        .filter(|(name, _)| model.contains(&name.to_lowercase()))
        .max_by_key(|(name, _)| name.len())
        .map(|(_, size)| *size)
        .filter(|size| *size > 0)
}

/// The folders searched upward for project configuration: the cwd and its parents,
/// stopping before HOME (#2455 AC-5, leader scope 6792437161). HOME's own `.moai` is the
/// Go tool's home folder, not a project, so it is never read as project configuration and
/// nothing above HOME is searched from a cwd below it — the same stop as
/// `find_project_root`. A cwd outside HOME walks up to the root as before. Folders are
/// compared canonically, so a HOME given through a symlink is still recognised.
fn project_config_dirs(ctx: &Context) -> Vec<PathBuf> {
    let Some(mut dir) = ctx.cwd.clone() else {
        return Vec::new();
    };
    let home = ctx.home().map(|h| std::fs::canonicalize(&h).unwrap_or(h));
    let mut out = Vec::new();
    loop {
        let here = std::fs::canonicalize(&dir).unwrap_or_else(|_| dir.clone());
        if home.as_deref() == Some(here.as_path()) {
            return out;
        }
        out.push(dir.clone());
        if !dir.pop() {
            return out;
        }
    }
}

/// `readLLMYAMLContextWindows`: the nearest `llm.yaml` above the cwd decides (below HOME
/// only, see `project_config_dirs`).
fn llm_context_windows(ctx: &Context) -> Vec<(String, i64)> {
    for dir in project_config_dirs(ctx) {
        let path = dir.join(".moai/config/sections/llm.yaml");
        if path.exists() {
            let Some(Node::Map(map)) = config::read(&path)
                .as_ref()
                .and_then(|d| d.get(&["llm", "glm", "context_windows"]).cloned())
            else {
                return Vec::new();
            };
            let mut out = Vec::new();
            for (k, v) in map {
                match v {
                    Node::Scalar(s) => match s.parse::<i64>() {
                        Ok(n) => out.push((k, n)),
                        Err(_) => return Vec::new(),
                    },
                    Node::Map(_) => return Vec::new(),
                }
            }
            return out;
        }
    }
    Vec::new()
}

// ---- version ----

/// `VersionCollector.CheckUpdate`: the nearest `system.yaml` above the cwd names the
/// template version; a different binary version is shown as the update target.
fn collect_version(ctx: &Context, binary: &str) -> Version {
    let trim = |v: &str| v.strip_prefix('v').unwrap_or(v).to_string();
    let bv = trim(binary);
    let Some(template) = template_version(ctx) else {
        return Version {
            current: bv,
            latest: None,
        };
    };
    let tv = trim(&template);
    let latest = (!bv.is_empty() && !tv.is_empty() && bv != tv).then_some(bv);
    Version {
        current: tv,
        latest,
    }
}

/// The nearest `system.yaml` above the cwd, below HOME only (see `project_config_dirs`).
fn template_version(ctx: &Context) -> Option<String> {
    for dir in project_config_dirs(ctx) {
        let path = dir.join(".moai/config/sections/system.yaml");
        if path.exists() {
            let doc = config::read(&path)?;
            let pick = |key| {
                doc.scalar(&["moai", key])
                    .filter(|v| !v.is_empty())
                    .map(str::to_string)
            };
            return pick("template_version").or_else(|| pick("version"));
        }
    }
    None
}

// ---- git ----

fn collect_git(dir: &Path, ctx: &Context) -> Option<GitStatus> {
    // No repository above `dir`: nothing to ask git, so no process is started.
    find_dot_git(&std::fs::canonicalize(dir).ok()?)?;
    let out = run_git(
        dir,
        &["--no-optional-locks", "status", "--porcelain", "--branch"],
        ctx,
    )?;
    let mut status = segments::parse_porcelain(&out);
    if status.branch.is_empty() {
        status.branch = run_git(dir, &["symbolic-ref", "--short", "HEAD"], ctx)
            .map(|b| b.trim().to_string())
            .unwrap_or_default();
    }
    Some(status)
}

/// Most bytes of one git call's stdout the render reads; the rest is dropped.
pub const GIT_OUTPUT_CAP: u64 = 1024 * 1024;

fn read_git_output(stdout: impl Read) -> std::io::Result<Vec<u8>> {
    let mut buf = Vec::new();
    stdout
        .take(GIT_OUTPUT_CAP)
        .read_to_end(&mut buf)
        .map(|_| buf)
}

/// One read-only git call, its environment scrubbed like every other git runner in the
/// crate, stopped at the render deadline.
fn run_git(dir: &Path, args: &[&str], ctx: &Context) -> Option<String> {
    // Past the deadline no git is started (Go: a command on an expired context never starts).
    if ctx.deadline.is_some_and(|d| Instant::now() >= d) {
        return None;
    }
    #[cfg(test)]
    tests::GIT_SPAWNS.with(|n| n.set(n.get() + 1));
    let mut cmd = Command::new(&ctx.git_program);
    cmd.args(args)
        .current_dir(dir)
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::null());
    crate::identity::scrub_git_env(&mut cmd);
    cmd.env("GIT_TERMINAL_PROMPT", "0");
    let mut child = cmd.spawn().ok()?;
    let stdout = child.stdout.take()?;
    let reader = std::thread::spawn(move || read_git_output(stdout));
    loop {
        match child.try_wait() {
            Ok(Some(status)) => {
                let out = reader.join().ok()?.ok()?;
                return status
                    .success()
                    .then(|| String::from_utf8_lossy(&out).into_owned());
            }
            Ok(None) if ctx.deadline.is_some_and(|d| Instant::now() >= d) => {
                let _ = child.kill();
                let _ = child.wait();
                return None;
            }
            Ok(None) => std::thread::sleep(Duration::from_millis(2)),
            Err(_) => return None,
        }
    }
}

// ---- local state files ----

/// `resolveGoalArmed`: `<anchor>/.moai/state/goal/<session_id>.json` has `status: armed`.
/// A session id carrying a path separator is refused rather than joined.
fn goal_armed(anchor: &Path, session_id: &str) -> bool {
    if session_id.is_empty() || session_id.contains(['/', '\\']) || session_id == ".." {
        return false;
    }
    #[derive(Deserialize, Default)]
    #[serde(default)]
    struct Goal {
        status: Option<String>,
    }
    let path = anchor
        .join(".moai/state/goal")
        .join(format!("{session_id}.json"));
    std::fs::read(path)
        .ok()
        .and_then(|b| serde_json::from_slice::<Goal>(&b).ok())
        .is_some_and(|g| g.status.as_deref() == Some("armed"))
}

/// `resolveGitHubCounts`: the cache only, never the network (#2588: the refresh runs
/// in a detached child, see [`super::counts`]); an explicit `statusline.forge` naming no
/// forge suppresses the pair, and counts fetched longer ago than
/// [`super::counts::STALE`] are hidden rather than shown as current.
fn github_counts(anchor: &Path) -> GitHubCounts {
    let now = super::counts::now_unix();
    let mut counts = super::counts::read(anchor)
        .filter(|c| c.suppressed == Some(true) || super::counts::is_fresh(c, now))
        .map_or(GitHubCounts::default(), |c| GitHubCounts {
            available: true,
            suppressed: c.suppressed.unwrap_or(false),
            open_issues: c.open_issues.unwrap_or(0),
            open_prs: c.open_prs.unwrap_or(0),
        });
    if forge_suppressed(anchor) {
        counts.suppressed = true;
    }
    counts
}

/// An explicit `statusline.forge` naming neither github nor gitlab.
pub fn forge_suppressed(anchor: &Path) -> bool {
    [STATUSLINE_LOCAL_YAML, STATUSLINE_YAML]
        .into_iter()
        .filter_map(|pair| config::read(&config_path(anchor, pair)))
        .filter_map(|doc| {
            doc.scalar(&["statusline", "forge"])
                .map(|s| s.trim().to_string())
        })
        .find(|v| !v.is_empty())
        .is_some_and(|forge| !matches!(forge.to_lowercase().as_str(), "github" | "gitlab"))
}

#[cfg(test)]
mod tests {
    use super::*;

    thread_local! {
        /// `run_git` spawn attempts on this test thread.
        pub(super) static GIT_SPAWNS: std::cell::Cell<usize> = const { std::cell::Cell::new(0) };
    }

    #[test]
    fn glm_names_map_only_non_claude_slot_models() {
        let ctx = Context::for_test(&[("ANTHROPIC_DEFAULT_OPUS_MODEL", "glm-5.3")], None);
        assert_eq!(resolve_glm_model_name("Opus", &ctx), "glm-5.3");
        assert_eq!(resolve_glm_model_name("Sonnet", &ctx), "Sonnet");
        assert_eq!(resolve_glm_model_name("gpt-x[1m]", &ctx), "gpt-x");
        let claude =
            Context::for_test(&[("ANTHROPIC_DEFAULT_OPUS_MODEL", "claude-opus-5-5")], None);
        assert_eq!(resolve_glm_model_name("Opus", &claude), "Opus");
    }

    #[test]
    fn context_override_prefers_env_then_longest_builtin_match() {
        let env = Context::for_test(&[("MOAI_STATUSLINE_CONTEXT_SIZE", "128000")], None);
        assert_eq!(context_window_override(&env), 128_000);
        let glm = Context::for_test(&[("ANTHROPIC_DEFAULT_SONNET_MODEL", "GLM-5.3-Flash")], None);
        assert_eq!(context_window_override(&glm), 1_000_000);
        assert_eq!(context_window_override(&Context::for_test(&[], None)), 0);
    }

    #[test]
    fn version_compares_template_with_binary() {
        let tmp = tempfile::tempdir().unwrap();
        let ctx = Context::for_test(&[], Some(tmp.path()));
        assert_eq!(
            collect_version(&ctx, "v3.2.1"),
            Version {
                current: "3.2.1".into(),
                latest: None
            }
        );
        let sections = tmp.path().join(".moai/config/sections");
        std::fs::create_dir_all(&sections).unwrap();
        std::fs::write(
            sections.join("system.yaml"),
            "moai:\n  template_version: v2.0.0\n",
        )
        .unwrap();
        assert_eq!(
            collect_version(&ctx, "3.2.1"),
            Version {
                current: "2.0.0".into(),
                latest: Some("3.2.1".into())
            }
        );
        std::fs::write(sections.join("system.yaml"), "moai:\n  version: 3.2.1\n").unwrap();
        assert_eq!(collect_version(&ctx, "v3.2.1").latest, None);
    }

    /// #2455 AC-5 (leader scope 6792437161): the upward search for `llm.yaml` and
    /// `system.yaml` never treats HOME's own `.moai` as project configuration and stops
    /// there; a project folder below HOME and a cwd outside HOME keep their reads.
    #[test]
    fn project_config_walk_stops_at_home() {
        let tmp = tempfile::tempdir().unwrap();
        let root = std::fs::canonicalize(tmp.path()).unwrap();
        let home = root.join("home");
        let put = |dir: &Path, llm: &str, version: &str| {
            let s = dir.join(".moai/config/sections");
            std::fs::create_dir_all(&s).unwrap();
            std::fs::write(
                s.join("llm.yaml"),
                format!("llm:\n  glm:\n    context_windows:\n      glm-marker: {llm}\n"),
            )
            .unwrap();
            std::fs::write(
                s.join("system.yaml"),
                format!("moai:\n  template_version: {version}\n"),
            )
            .unwrap();
        };
        put(&home, "123456", "v9.9.9");
        let deep = home.join("a/b/c");
        std::fs::create_dir_all(&deep).unwrap();
        let link = root.join("home-link");
        #[cfg(unix)]
        std::os::unix::fs::symlink(&home, &link).unwrap();
        let at = |home_var: &Path, cwd: &Path| {
            Context::for_test(&[("HOME", home_var.to_str().unwrap())], Some(cwd))
        };

        // Below HOME with no nearer project, and HOME itself: HOME's `.moai` is not read.
        for cwd in [&deep, &home] {
            let ctx = at(&home, cwd);
            assert_eq!(llm_context_windows(&ctx), Vec::new(), "{cwd:?}");
            assert_eq!(template_version(&ctx), None, "{cwd:?}");
        }
        // The search stops at HOME rather than skipping it: a folder above HOME holding
        // `.moai` is not reached from below HOME either.
        put(&root, "999", "v8.8.8");
        let ctx = at(&home, &deep);
        assert_eq!(llm_context_windows(&ctx), Vec::new());
        assert_eq!(template_version(&ctx), None);
        // HOME is recognised through a symlink, given as HOME or as part of the cwd.
        #[cfg(unix)]
        for (home_var, cwd) in [(&link, deep.clone()), (&home, link.join("a/b/c"))] {
            let ctx = at(home_var, &cwd);
            assert_eq!(llm_context_windows(&ctx), Vec::new(), "{cwd:?}");
            assert_eq!(template_version(&ctx), None, "{cwd:?}");
        }

        // A project folder directly below HOME keeps the #2389 project read, from its
        // root and from below it.
        let proj = home.join("proj");
        put(&proj, "777", "v1.2.3");
        let sub = proj.join("sub");
        std::fs::create_dir_all(&sub).unwrap();
        for cwd in [&proj, &sub] {
            let ctx = at(&home, cwd);
            assert_eq!(
                llm_context_windows(&ctx),
                vec![("glm-marker".to_string(), 777)],
                "{cwd:?}"
            );
            assert_eq!(template_version(&ctx).as_deref(), Some("v1.2.3"), "{cwd:?}");
        }

        // A cwd outside HOME walks up as before.
        let out = root.join("outside");
        put(&out, "555", "v4.5.6");
        let below = out.join("x");
        std::fs::create_dir_all(&below).unwrap();
        let ctx = at(&home, &below);
        assert_eq!(
            llm_context_windows(&ctx),
            vec![("glm-marker".to_string(), 555)]
        );
        assert_eq!(template_version(&ctx).as_deref(), Some("v4.5.6"));
    }

    /// #2350 scope 3 (decision 6774899518 row 42 = B): `.moai-cli/` first, else the Go
    /// file, which is only read.
    #[test]
    fn statusline_config_reads_moai_cli_first_and_keeps_the_go_file() {
        let tmp = tempfile::tempdir().unwrap();
        let ctx = Context::for_test(&[], Some(tmp.path()));
        let go = tmp.path().join(".moai/config/sections");
        std::fs::create_dir_all(&go).unwrap();
        let go_file = go.join("statusline.yaml");
        let go_text =
            "statusline:\n  theme: catppuccin\n  forge: none\n  segments:\n    git: false\n";
        std::fs::write(&go_file, go_text).unwrap();

        // Only the Go file: it is read.
        let c = display_config(None, &ctx);
        assert_eq!(c.theme, "catppuccin");
        assert_eq!(c.segments.get("git"), Some(&false));
        assert!(github_counts(tmp.path()).suppressed);

        // A `.moai-cli/` file wins whole; nothing is merged from the Go file.
        let rc = tmp.path().join(crate::project_state!("config/sections"));
        std::fs::create_dir_all(&rc).unwrap();
        std::fs::write(
            rc.join("statusline.yaml"),
            "statusline:\n  theme: nord\n  forge: github\n",
        )
        .unwrap();
        let c = display_config(None, &ctx);
        assert_eq!(c.theme, "nord");
        assert!(c.segments.is_empty());
        assert!(!github_counts(tmp.path()).suppressed);

        // An unreadable `.moai-cli/` file means defaults, not the Go values.
        std::fs::write(rc.join("statusline.yaml"), "statusline:\n\ttheme: x\n").unwrap();
        let c = display_config(None, &ctx);
        assert_eq!(c.theme, "");
        assert!(c.segments.is_empty());

        // The local file follows the same order.
        std::fs::write(
            tmp.path().join(".moai/statusline.local.yaml"),
            "statusline:\n  forge: github\n",
        )
        .unwrap();
        std::fs::write(
            tmp.path()
                .join(crate::project_state!("statusline.local.yaml")),
            "statusline:\n  forge: none\n",
        )
        .unwrap();
        assert!(github_counts(tmp.path()).suppressed);

        // The Go file is never written, and nothing new appears under `.moai/`.
        assert_eq!(std::fs::read_to_string(&go_file).unwrap(), go_text);
        let mut names: Vec<_> = std::fs::read_dir(tmp.path().join(".moai"))
            .unwrap()
            .map(|e| e.unwrap().file_name())
            .collect();
        names.sort();
        assert_eq!(names, ["config", "statusline.local.yaml"]);
    }

    #[test]
    fn goal_and_forge_reads() {
        let tmp = tempfile::tempdir().unwrap();
        let goal = tmp.path().join(".moai/state/goal");
        std::fs::create_dir_all(&goal).unwrap();
        std::fs::write(goal.join("s1.json"), r#"{"status":"armed"}"#).unwrap();
        assert!(goal_armed(tmp.path(), "s1"));
        assert!(!goal_armed(tmp.path(), "s2"));
        assert!(!goal_armed(tmp.path(), "../goal/s1"));
        assert_eq!(github_counts(tmp.path()), GitHubCounts::default());
        let gh = tmp.path().join(".moai/state/github");
        std::fs::create_dir_all(&gh).unwrap();
        // #2588 AC-3: a count fetched more than an hour ago is hidden, not shown as current.
        std::fs::write(
            gh.join("counts.json"),
            r#"{"open_issues":7,"open_prs":3,"fetched_at":1}"#,
        )
        .unwrap();
        assert_eq!(github_counts(tmp.path()), GitHubCounts::default());
        let now = super::super::counts::now_unix();
        std::fs::write(
            gh.join("counts.json"),
            format!(r#"{{"open_issues":7,"open_prs":3,"fetched_at":{now}}}"#),
        )
        .unwrap();
        let c = github_counts(tmp.path());
        assert_eq!(
            (c.available, c.open_issues, c.open_prs, c.suppressed),
            (true, 7, 3, false)
        );
        // The cache this tool writes wins over the Go-generation file.
        let own = tmp.path().join(crate::project_state!("state/github"));
        std::fs::create_dir_all(&own).unwrap();
        std::fs::write(
            own.join("counts.json"),
            format!(r#"{{"open_issues":9,"open_prs":4,"fetched_at":{now}}}"#),
        )
        .unwrap();
        let c = github_counts(tmp.path());
        assert_eq!((c.open_issues, c.open_prs), (9, 4));
        let sections = tmp.path().join(".moai/config/sections");
        std::fs::create_dir_all(&sections).unwrap();
        std::fs::write(
            sections.join("statusline.yaml"),
            "statusline:\n  forge: none\n",
        )
        .unwrap();
        assert!(github_counts(tmp.path()).suppressed);
    }

    #[test]
    fn anchor_follows_a_linked_worktree_to_the_main_checkout() {
        let tmp = tempfile::tempdir().unwrap();
        let main = tmp.path().join("main");
        let gitdir = main.join(".git/worktrees/wt");
        std::fs::create_dir_all(&gitdir).unwrap();
        std::fs::write(gitdir.join("commondir"), "../..\n").unwrap();
        let wt = tmp.path().join("wt");
        std::fs::create_dir_all(wt.join("sub")).unwrap();
        std::fs::write(wt.join(".git"), format!("gitdir: {}\n", gitdir.display())).unwrap();
        let canon = std::fs::canonicalize(&main).unwrap();
        assert_eq!(main_checkout_of(&wt.join("sub")), Some(canon.clone()));
        assert_eq!(main_checkout_of(&main), Some(canon));
    }

    #[test]
    fn git_output_is_read_up_to_exactly_one_mebibyte() {
        let read = |n: usize| read_git_output(&vec![b'x'; n][..]).unwrap().len();
        assert_eq!(read(1_048_575), 1_048_575);
        assert_eq!(read(1_048_576), 1_048_576);
        assert_eq!(read(1_048_577), 1_048_576);
    }

    /// `git init -b main` for a test repository, scrubbed like every git call in `src/`
    /// (#2337), so an inherited `GIT_DIR` cannot redirect it.
    fn git_init(dir: &Path) {
        let mut cmd = Command::new("git");
        cmd.args(["init", "-q", "-b", "main"]).current_dir(dir);
        crate::identity::scrub_git_env(&mut cmd);
        assert!(cmd.status().unwrap().success());
    }

    #[test]
    fn git_is_stopped_at_the_render_deadline() {
        let tmp = tempfile::tempdir().unwrap();
        git_init(tmp.path());
        let mut ctx = Context::for_test(&[], Some(tmp.path()));
        assert!(run_git(tmp.path(), &["status", "--porcelain"], &ctx).is_some());
        ctx.deadline = Some(Instant::now());
        assert_eq!(run_git(tmp.path(), &["status", "--porcelain"], &ctx), None);
    }

    /// AC-2: git is started only when a `.git` entry is found, and then once.
    #[cfg(unix)]
    #[test]
    fn git_is_started_only_inside_a_repository() {
        use std::os::unix::fs::PermissionsExt;
        let tmp = tempfile::tempdir().unwrap();
        let log = tmp.path().join("git-calls.log");
        let wrapper = tmp.path().join("git-wrapper");
        std::fs::write(
            &wrapper,
            format!(
                "#!/bin/sh\necho \"$*\" >> '{}'\nexec git \"$@\"\n",
                log.display()
            ),
        )
        .unwrap();
        std::fs::set_permissions(&wrapper, std::fs::Permissions::from_mode(0o755)).unwrap();
        let calls = || std::fs::read_to_string(&log).map_or(0, |s| s.lines().count());

        let plain = tmp.path().join("plain");
        std::fs::create_dir(&plain).unwrap();
        let mut ctx = Context::for_test(&[], Some(&plain));
        ctx.git_program = wrapper.display().to_string();
        assert!(collect_git(&plain, &ctx).is_none());
        assert_eq!(calls(), 0);

        let repo = tmp.path().join("repo");
        std::fs::create_dir(&repo).unwrap();
        git_init(&repo);
        assert_eq!(
            collect_git(&repo, &ctx).map(|s| s.branch),
            Some("main".into())
        );
        assert_eq!(calls(), 1);

        // An expired deadline: nothing is started and the git status is absent. The
        // wrapper log cannot show this (a started child is killed before it writes), so the
        // spawn attempts of this thread are counted.
        ctx.deadline = Some(Instant::now());
        let before = GIT_SPAWNS.with(|n| n.get());
        assert!(collect_git(&repo, &ctx).is_none());
        assert_eq!(GIT_SPAWNS.with(|n| n.get()), before);
        ctx.deadline = None;
        assert!(collect_git(&repo, &ctx).is_some());
        assert_eq!(GIT_SPAWNS.with(|n| n.get()), before + 1);
    }
}

/// #2756 AC-7–AC-9: the pull request, prompt cache and worktree values the renderer shows.
#[cfg(test)]
mod extras_tests {
    use super::*;
    use serde_json::json;

    fn collect_json(value: serde_json::Value, now: i64) -> Status {
        let input = input::parse(value.to_string().as_bytes()).unwrap();
        let mut ctx = Context::for_test(&[], None);
        ctx.now = now;
        collect(Some(&input), &ctx)
    }

    #[test]
    fn the_cache_timer_counts_whole_minutes_and_turns_cold_at_expiry() {
        let at = |cache: serde_json::Value| cache_timer(&cache, 1000);
        assert_eq!(
            at(json!({"warm": true, "expires_at": 1000 + 299})),
            Some(CacheTimer::Warm(4))
        );
        assert_eq!(
            at(json!({"warm": true, "expires_at": 1000 + 60})),
            Some(CacheTimer::Warm(1))
        );
        assert_eq!(
            at(json!({"warm": true, "expires_at": 1000 + 59})),
            Some(CacheTimer::Warm(0))
        );
        assert_eq!(
            at(json!({"warm": true, "expires_at": 1000 + 1})),
            Some(CacheTimer::Warm(0))
        );
        assert_eq!(
            at(json!({"warm": true, "expires_at": 1000})),
            Some(CacheTimer::Cold)
        );
        assert_eq!(
            at(json!({"warm": false, "expires_at": null})),
            Some(CacheTimer::Cold)
        );
        // No value to show: no `warm`, a warm cache without `expires_at`, other types.
        assert_eq!(at(json!({"expires_at": 1300})), None);
        assert_eq!(at(json!({"warm": true, "expires_at": null})), None);
        assert_eq!(at(json!({"warm": "yes"})), None);
        assert_eq!(at(json!("warm")), None);
    }

    #[test]
    fn a_link_url_must_be_plain_http() {
        assert_eq!(
            safe_link("https://github.com/o/r/pull/1"),
            "https://github.com/o/r/pull/1"
        );
        assert_eq!(safe_link("http://example.test/1"), "http://example.test/1");
        for bad in [
            "",
            "javascript:alert(1)",
            "file:///etc/passwd",
            "https://x/\x1b]8;;evil\x1b\\",
            "https://x/a b",
            "https://x/\u{7}",
            "https://x/é",
        ] {
            assert_eq!(safe_link(bad), "", "{bad:?}");
        }
        let long = format!("https://x/{}", "a".repeat(2048));
        assert_eq!(safe_link(&long), "");
    }

    #[test]
    fn pr_worktree_and_cache_come_from_the_input_or_stay_empty() {
        let s = collect_json(
            json!({
                "pr": {"number": 12, "url": "https://github.com/o/r/pull/12", "review_state": "approved"},
                "worktree": {"name": "wt-session"},
                "workspace": {"git_worktree": "linked"},
                "prompt_cache": {"warm": true, "expires_at": 600}
            }),
            0,
        );
        assert_eq!(
            s.pr,
            Some(PrInfo {
                number: 12,
                url: "https://github.com/o/r/pull/12".into(),
                review_state: "approved".into()
            })
        );
        assert_eq!(
            s.worktree, "wt-session",
            "worktree.name wins over git_worktree"
        );
        assert_eq!(s.cache_timer, Some(CacheTimer::Warm(10)));
        let s = collect_json(
            json!({"workspace": {"git_worktree": "linked"}, "pr": {"number": 3}}),
            0,
        );
        assert_eq!(
            s.worktree, "linked",
            "outside a worktree session, git_worktree"
        );
        let pr = s.pr.unwrap();
        assert!(pr.url.is_empty() && pr.review_state.is_empty());
        let s = collect_json(json!({"pr": {"number": 0}, "workspace": {}}), 0);
        assert_eq!((s.pr, s.worktree.as_str(), s.cache_timer), (None, "", None));
        // A mistyped prompt_cache leaves the timer out but keeps the rest of the payload.
        let s = collect_json(json!({"session_name": "x", "prompt_cache": 7}), 0);
        assert_eq!((s.session_name.as_str(), s.cache_timer), ("x", None));
    }
}
