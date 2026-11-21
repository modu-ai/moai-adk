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
//! Card C-65: the card counts are the one read under `$HOME/.moai/`: the project's local ledger
//! in `$HOME/.moai/projects/<project_id>/`, opened read-only through `crate::ledger` with a
//! short wait (see [`card_counts`]).
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
            .map(|h| crate::home_root::path(&h).join("state"))
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
    /// #2928 (decision D1 = (A)): the clone's lane registry holds a live `leader` entry for the
    /// stdin `session_name`. The tag shows `Leader` then, and also for the exact name `leader`
    /// (#2871); the name text itself is never copied (see `render::tag_text`).
    pub leader: bool,
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
    /// Card C-65: the card counts of the project's local ledger; `None` when there is no
    /// ledger to read (see [`card_counts`]).
    pub cards: Option<CardCounts>,
    pub goal_armed: bool,
    /// #2756 AC-7: the open pull request of the current branch.
    pub pr: Option<PrInfo>,
    /// #2756 AC-8: how long the prompt cache stays warm.
    pub cache_timer: Option<CacheTimer>,
    /// #2756 AC-9: the worktree name; empty in the main checkout.
    pub worktree: String,
    /// User instruction 2026-11-19: the card and the name of the held reservation that names
    /// the session's worktree folder, for `🌳 WT-c<card>-<name>`; `None` shows `worktree` as
    /// the host gave it.
    pub worktree_label: Option<WorktreeLabel>,
}

/// The parts of `🌳 WT-c<card>-<name>`: the number of the card holding the worktree and the
/// name that says what the work is (empty when nothing says it).
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct WorktreeLabel {
    pub card: i64,
    pub name: String,
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

/// Card C-65 (user instructions 2026-11-16 and 2026-11-17): the cards of the project's local
/// ledger, as the repository line shows them (`🗂️ remaining ▶️ in_progress 🔃 waiting ✅ done`;
/// without `waiting` the three-number form `🗂️ ▶️ ✅`).
#[derive(Debug, Default, Clone, Copy, PartialEq, Eq)]
pub struct CardCounts {
    /// The cards being worked on: `active` ones whose newest marker note is not a waiting one.
    /// When `waiting` is `None` (the notes could not be read) it is `active` + `review`.
    pub in_progress: i64,
    /// The cards pushed and waiting for integration or an audit: every `review` card and each
    /// `active` card whose newest marker note is a waiting one. `None` when the notes could not
    /// be read, and the line then shows the three-number form.
    pub waiting: Option<i64>,
    /// `draft` + `ready` + `hold`.
    pub remaining: i64,
    /// `done`; counted in neither of the other two.
    pub done: i64,
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
    status.leader = is_leader(input, &status.session_name, ctx.now);
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
    status.cards = anchor
        .as_deref()
        .or(project_root.as_deref())
        .and_then(|project| card_counts(project, ctx));
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
    if !status.worktree.is_empty() {
        status.worktree_label = anchor
            .as_deref()
            .or(project_root.as_deref())
            .zip(session_dir(input))
            .and_then(|(project, dir)| worktree_label(project, &dir, ctx));
    }
    status
}

/// The folder the session works in: `workspace.current_dir`, else `cwd`; `None` when neither
/// is given.
fn session_dir(input: &Input) -> Option<PathBuf> {
    let dir = input
        .workspace
        .as_ref()
        .map_or("", |w| input::text(&w.current_dir));
    let dir = if dir.is_empty() {
        input::text(&input.cwd)
    } else {
        dir
    };
    (!dir.is_empty()).then(|| PathBuf::from(dir))
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
///
/// #2936 S1g: once `go-state-apply` moved the Go `.moai/` away, the Rust and the Go path are the
/// same place in the new `.moai/` and both are empty, so a project that set a theme or hid a
/// segment in Go got the defaults. With neither file present, the same Go path inside the newest
/// Go backup folder (`.moai-backup`, `.moai-backup-<UTC>`) is read — read only, never moved or
/// copied; a file in the new `.moai/` still wins whole. The backup is looked at only after such a
/// move (the state folder is `.moai`) and only while the new folder has no shared
/// `statusline.yaml` of its own (AUDIT #3125 S1, S2): a project that wrote its new settings has
/// said what it wants, and an old `forge: none` in the backup's local file must not outvote it.
fn config_path(root: &Path, (rc, go): (&str, &str)) -> PathBuf {
    // #2725 S2b: `rc` is relative to the project's state folder (`.moai-cli`, or `.moai`).
    let state = crate::state_root::dir(root);
    let rc = root.join(state).join(rc);
    if rc.exists() {
        return rc;
    }
    let go_file = root.join(go);
    if go_file.exists() {
        return go_file;
    }
    if state != crate::state_root::MOAI_DIR || root.join(state).join(STATUSLINE_YAML.0).exists() {
        return go_file;
    }
    // The backup is the Go `.moai/` itself, so the path inside it drops the leading `.moai/`.
    let inside = go.strip_prefix(".moai/").unwrap_or(go);
    crate::state_root::latest_go_backup(root)
        .map(|backup| backup.join(inside))
        .filter(|path| path.exists())
        .unwrap_or(go_file)
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
/// stopping at HOME — HOME's `.moai` is the tool's home folder, never a project — and at
/// the git repository (#2756 leader decision 6862179411): a folder with `.git` (a folder,
/// or the file of a linked worktree) is the last one searched, so a `.moai` above the
/// repository never becomes the root and the git segments read the repository itself.
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
        if dir.join(".git").exists() || !dir.pop() {
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

/// #2928: the session is the leader when its `session_name` is a valid lane name and the lane
/// registry of the clone holding the session's directory (`workspace.current_dir`, else `cwd`)
/// has a live entry for that name with mode `leader` (`lane_registry::mode_of`).
fn is_leader(input: &Input, session_name: &str, now: i64) -> bool {
    if !crate::lane_registry::valid_name(session_name) {
        return false;
    }
    let dir = input
        .workspace
        .as_ref()
        .map_or("", |w| input::text(&w.current_dir));
    let dir = if dir.is_empty() {
        input::text(&input.cwd)
    } else {
        dir
    };
    if dir.is_empty() {
        return false;
    }
    common_dir_of(Path::new(dir)).is_some_and(|common| {
        crate::lane_registry::mode_of(&common, session_name, now) == Some("leader")
    })
}

/// Where the git common directory of `dir` is, found on disk: the main checkout's `.git`
/// folder (with that checkout's top), or the folder a linked worktree's `commondir` names.
enum GitCommon {
    Main { top: PathBuf, dot_git: PathBuf },
    Linked(PathBuf),
}

fn find_common(dir: &Path) -> Option<GitCommon> {
    let dir = std::fs::canonicalize(dir).ok()?;
    let (top, dot_git) = find_dot_git(&dir)?;
    if dot_git.is_dir() {
        return Some(GitCommon::Main { top, dot_git });
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
    std::fs::canonicalize(common).ok().map(GitCommon::Linked)
}

/// The parent of the git common directory for `dir`: `.git` as a directory, or a `.git`
/// file naming a worktree git dir whose `commondir` points back to the main repository.
fn main_checkout_of(dir: &Path) -> Option<PathBuf> {
    match find_common(dir)? {
        GitCommon::Main { top, .. } => Some(top),
        GitCommon::Linked(common) => common.parent().map(Path::to_path_buf),
    }
}

/// #2928: the git common directory of `dir`, canonical — the folder the clone's lane registry
/// lives under (`identity`'s `clone_anchor`).
fn common_dir_of(dir: &Path) -> Option<PathBuf> {
    match find_common(dir)? {
        GitCommon::Main { dot_git, .. } => std::fs::canonicalize(dot_git).ok(),
        GitCommon::Linked(common) => Some(common),
    }
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
/// template version; a different binary version is shown as the update target, and only a
/// newer one when both versions are ones `headver::compare` reads.
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
    // Card C-67 S1: when both are versions the HeadVer comparison reads, the binary is the
    // update target only when it is newer; otherwise any difference shows, as before.
    let newer = crate::headver::compare(&bv, &tv).is_none_or(|o| o.is_gt());
    let latest = (!bv.is_empty() && !tv.is_empty() && bv != tv && newer).then_some(bv);
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
        // Detached HEAD (SPEC-20261121-statusline-detached): the short SHA stands in for
        // the branch name, so the repository and its forge counts stay on the line.
        if status.branch.is_empty() {
            status.head_sha = run_git(dir, &["rev-parse", "--short", "HEAD"], ctx)
                .map(|b| b.trim().to_string())
                .unwrap_or_default();
        }
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

/// How long the card counts wait for a ledger another process has locked (card C-65). The
/// status line is drawn again within seconds, so a locked ledger costs one render its segment,
/// never the render budget.
pub const CARDS_BUSY_TIMEOUT: Duration = Duration::from_millis(50);

// Card C-68 AC-2 (the C-65 audit's N1): the ledger wait stays a bounded part of the render
// budget — 50 ms today, 100 ms at the very most. A value above the bound fails this build.
const _: () = assert!(CARDS_BUSY_TIMEOUT.as_millis() <= 100);

/// Card C-65: the card counts of `project`'s local ledger
/// (`$HOME/.moai/projects/<project_id>/`), the one `moai card list` reads. Read-only like that
/// command (`ledger::open`: no folder, no file, no journal; no process and no network), with
/// [`CARDS_BUSY_TIMEOUT`] — or what is left of the render budget when that is less — in place of
/// the ledger's own wait. `None`, and so no segment, whenever the counts cannot be read: no
/// HOME, no project profile, no registered ledger, a file that is not this project's ledger
/// (corrupt, another project's, a link), a lock held past the wait, or a spent budget.
fn card_counts(project: &Path, ctx: &Context) -> Option<CardCounts> {
    let (ledger, _) = open_project_ledger(project, ctx)?;
    let mut counts = CardCounts::default();
    for (state, n) in crate::ledger::command::state_counts(&ledger).ok()? {
        match state.as_str() {
            "active" | "review" => counts.in_progress += n,
            "draft" | "ready" | "hold" => counts.remaining += n,
            "done" => counts.done += n,
            _ => {}
        }
    }
    // Split the `active` + `review` total; when the notes cannot be read the total stays whole
    // and the line shows the three-number form.
    counts.waiting = waiting_cards(&ledger, ctx.deadline);
    if let Some(waiting) = counts.waiting {
        counts.in_progress = split_in_progress(counts.in_progress, waiting);
    }
    Some(counts)
}

/// The project's ledger opened read-only the way [`card_counts`] reads it, with the project id.
/// `None` for every case [`card_counts`] documents (no HOME, no profile, no registered ledger,
/// a file that is not this project's ledger, a lock held past the wait, a spent budget).
fn open_project_ledger(
    project: &Path,
    ctx: &Context,
) -> Option<(crate::ledger::open::ReadLedger, String)> {
    let home = ctx.home()?;
    let busy = match ctx.deadline {
        Some(deadline) => CARDS_BUSY_TIMEOUT.min(deadline.checked_duration_since(Instant::now())?),
        None => CARDS_BUSY_TIMEOUT,
    };
    let id = crate::project_id::read(&std::fs::canonicalize(project).ok()?).ok()??;
    let ledger = crate::ledger::open::open_read_only_within(&home, &id, busy).ok()?;
    Some((ledger, id))
}

/// User instruction 2026-11-19: which card holds the worktree the session is in, and what the
/// work is called. `dir` is a folder inside the worktree (`workspace.current_dir`, else `cwd`).
/// The held reservation naming the worktree folder gives the card (the lowest number when more
/// than one names it) and its branch the name ([`label_name`]). `None` when no reservation names
/// the folder or the ledger cannot be read in the budget: the line then shows the host's name.
fn worktree_label(project: &Path, dir: &Path, ctx: &Context) -> Option<WorktreeLabel> {
    // The folder holding the nearest `.git` entry above `dir`, in the two spellings a holder
    // may have recorded: as found and canonical.
    let (found, _) = find_dot_git(dir)?;
    let canonical = std::fs::canonicalize(&found).unwrap_or_else(|_| found.clone());
    let (ledger, project_id) = open_project_ledger(project, ctx)?;
    let expired = || {
        ctx.deadline
            .is_some_and(|deadline| Instant::now() >= deadline)
    };
    let (card, branch) = holding_card(ledger.conn(), &found, &canonical, &expired)?;
    let folder = canonical
        .file_name()
        .map_or_else(String::new, |n| n.to_string_lossy().into_owned());
    Some(WorktreeLabel {
        card,
        name: label_name(branch.as_deref(), &folder, &project_id, card),
    })
}

/// The lowest card whose holding reservation names the folder `found` (as found) or `canonical`,
/// with that reservation's branch. The holder typed the path (`card take --worktree`), so it may
/// be a link spelling: each holding reservation is compared as written and as a canonical path.
/// A path that is not absolute names no folder from here (it would be read against this
/// process's working directory), so it never matches; a row that cannot be read is skipped.
/// `expired` is asked before every row, and `None` comes back as soon as it is true.
fn holding_card(
    conn: &rusqlite::Connection,
    found: &Path,
    canonical: &Path,
    expired: &dyn Fn() -> bool,
) -> Option<(i64, Option<String>)> {
    let mut stmt = conn
        .prepare(
            "SELECT card_id, branch, worktree FROM reservation \
             WHERE status IN ('held', 'held_for_review', 'expired_candidate') \
             AND worktree IS NOT NULL ORDER BY card_id",
        )
        .ok()?;
    let rows = stmt
        .query_map([], |r| {
            Ok((
                r.get::<_, i64>(0)?,
                r.get::<_, Option<String>>(1)?,
                r.get::<_, String>(2)?,
            ))
        })
        .ok()?;
    for row in rows {
        if expired() {
            return None;
        }
        let Ok((card, branch, held)) = row else {
            continue;
        };
        let named = Path::new(&held);
        if !named.is_absolute() {
            continue;
        }
        if named == found
            || named == canonical
            || std::fs::canonicalize(named).is_ok_and(|c| c == canonical)
        {
            return Some((card, branch));
        }
    }
    None
}

/// The name part of `WT-c<card>-<name>`, from the reservation's `branch`: the part after the
/// eight-digit date and a dash in its last path segment, else the worktree folder name without
/// its `<project>-` prefix and its card number, else nothing; control and invisible format
/// characters dropped ([`is_unseen`]) before the length is counted, a source nothing is left of
/// giving way to the next, and at most [`LABEL_NAME_CHARS`] characters.
fn label_name(branch: Option<&str>, folder: &str, project_id: &str, card: i64) -> String {
    // Each source is cleaned first; a slug that nothing is left of gives way to the folder name.
    let clean = |raw: String| -> String { raw.chars().filter(|c| !is_unseen(*c)).collect() };
    let raw = branch
        .and_then(branch_slug)
        .map(clean)
        .filter(|s| !s.is_empty())
        .or_else(|| folder_rest(folder, project_id, card).map(clean))
        .unwrap_or_default();
    let mut chars = raw.chars();
    let mut name: String = chars.by_ref().take(LABEL_NAME_CHARS).collect();
    if chars.next().is_some() {
        // One character too many: keep 31 and mark the cut.
        name = name.chars().take(LABEL_NAME_CHARS - 1).collect();
        name.push('…');
    }
    name
}

/// A character the terminal would not show as itself: a control character, or one of the
/// invisible format characters that reorder or hide text (zero-width and bidirectional marks,
/// the byte order mark). The name is free text from a branch or a folder.
fn is_unseen(c: char) -> bool {
    c.is_control()
        || matches!(
            c,
            '\u{200b}'..='\u{200f}'
                | '\u{202a}'..='\u{202e}'
                | '\u{2060}'..='\u{2064}'
                | '\u{2066}'..='\u{206f}'
                | '\u{feff}'
        )
}

/// `spec/20261119-statusline-waiting-cards` gives `statusline-waiting-cards`: the last path
/// segment is eight digits, a dash and something.
fn branch_slug(branch: &str) -> Option<String> {
    let last = branch.rsplit('/').next()?;
    let (date, rest) = last.split_at_checked(8)?;
    if !date.bytes().all(|b| b.is_ascii_digit()) {
        return None;
    }
    let rest = rest.strip_prefix('-')?;
    (!rest.is_empty()).then(|| rest.to_string())
}

/// The worktree folder name without its `<project>-` prefix and without the card number
/// (`demo-1-hooks` gives `hooks`, `demo-c1-hooks` too); `None` when nothing is left.
fn folder_rest(folder: &str, project_id: &str, card: i64) -> Option<String> {
    let rest = match folder.strip_prefix(project_id) {
        Some("") => "",
        Some(after) => after.strip_prefix('-').unwrap_or(folder),
        None => folder,
    };
    let mut rest = rest;
    for token in [format!("c{card}"), card.to_string()] {
        match rest.strip_prefix(token.as_str()) {
            Some("") => rest = "",
            Some(after) if after.starts_with('-') => rest = &after[1..],
            _ => continue,
        }
        break;
    }
    let rest = rest.trim_matches('-');
    (!rest.is_empty()).then(|| rest.to_string())
}

/// The longest `<name>` of the worktree label, in characters; one more is cut with `…`.
const LABEL_NAME_CHARS: usize = 32;

/// The working number once the waiting cards are taken out of the `active` + `review` total.
/// The state count and the note read are two reads of a ledger that may change between them, so
/// a card that moved can make `waiting` larger than the total: never below zero.
fn split_in_progress(total: i64, waiting: i64) -> i64 {
    (total - waiting).max(0)
}

/// How many characters of a note's body are read: enough for its first line, so a long report
/// costs no more than a short one.
const NOTE_HEAD_CHARS: usize = 1000;

/// What a note's first line says about its card.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
enum Marker {
    /// The card was pushed and waits for an audit or for the leader's integration.
    Waiting,
    /// The lane is working on the card again (took or resumed it, planned, or was told FAIL).
    Working,
}

/// The marker a note's first line carries, if any. Only the lines the lanes already write by
/// rule count; any other first line is no marker and changes nothing.
fn marker_of(body: &str) -> Option<Marker> {
    let first = body.lines().next().unwrap_or("");
    let audit = first.starts_with("AUDIT");
    if first.starts_with("LANE-READY-INTEGRATION")
        || first.starts_with("감사 요청")
        || (first.starts_with("LANE-DONE")
            && (first.contains("audit requested") || first.contains("감사 요청")))
        || (audit && first.contains("verdict: PASS"))
    {
        Some(Marker::Waiting)
    } else if first.starts_with("ACK ")
        || first.starts_with("PLAN ")
        || (audit && first.contains("verdict: FAIL"))
    {
        Some(Marker::Working)
    } else {
        None
    }
}

/// The cards pushed and waiting for integration: every `review` card and each `active` card
/// whose newest note with a marker (see [`marker_of`]) is a waiting one. One more `SELECT` over
/// the notes of the `active` cards, newest first, each body cut to [`NOTE_HEAD_CHARS`]; no
/// write. `None` when the render `deadline` has passed (checked before the read and on every
/// row) or the read fails, so the caller shows the three-number form instead of a guess.
fn waiting_cards(
    ledger: &crate::ledger::open::ReadLedger,
    deadline: Option<Instant>,
) -> Option<i64> {
    waiting_cards_with(ledger, &|| deadline.is_some_and(|d| Instant::now() >= d))
}

/// [`waiting_cards`] with the "is the render budget spent" question as a parameter, so a test
/// can spend it in the middle of the row loop.
fn waiting_cards_with(
    ledger: &crate::ledger::open::ReadLedger,
    spent: &dyn Fn() -> bool,
) -> Option<i64> {
    if spent() {
        return None;
    }
    let review: i64 = ledger
        .conn()
        .query_row(
            "SELECT count(*) FROM card WHERE state = 'review'",
            [],
            |r| r.get(0),
        )
        .ok()?;
    let mut stmt = ledger
        .conn()
        .prepare(&format!(
            "SELECT n.card_id, substr(n.body, 1, {NOTE_HEAD_CHARS}) FROM card_note n \
             JOIN card c ON c.card_id = n.card_id WHERE c.state = 'active' \
             ORDER BY n.card_id, n.note_id DESC"
        ))
        .ok()?;
    let rows = stmt
        .query_map([], |r| Ok((r.get::<_, i64>(0)?, r.get::<_, String>(1)?)))
        .ok()?;
    let mut decided: Option<i64> = None;
    let mut waiting = review;
    for row in rows {
        if spent() {
            return None;
        }
        let (card, body) = row.ok()?;
        if decided == Some(card) {
            continue;
        }
        match marker_of(&body) {
            Some(Marker::Waiting) => {
                waiting += 1;
                decided = Some(card);
            }
            Some(Marker::Working) => decided = Some(card),
            None => {}
        }
    }
    Some(waiting)
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

    /// Card C-67 S1 (SPEC-20261117-headver-version AC-5): when both versions are ones the
    /// HeadVer comparison reads, the binary is the update target only when it is newer; a
    /// binary older than the template is never shown as the update.
    #[test]
    fn the_binary_is_the_update_target_only_when_newer() {
        let tmp = tempfile::tempdir().unwrap();
        let ctx = Context::for_test(&[], Some(tmp.path()));
        let sections = tmp.path().join(".moai/config/sections");
        std::fs::create_dir_all(&sections).unwrap();
        for (template, binary, latest) in [
            ("2026.11.0-rc.2", "4.2647.5", Some("4.2647.5")),
            ("4.2647.4", "4.2647.5", Some("4.2647.5")),
            ("4.2647.5", "4.2647.5", None),
            ("4.2648.1", "4.2647.5", None),
            ("4.2647.5", "2026.11.0-rc.2+local.4", None),
        ] {
            std::fs::write(
                sections.join("system.yaml"),
                format!("moai:\n  template_version: {template}\n"),
            )
            .unwrap();
            let v = collect_version(&ctx, binary);
            assert_eq!(v.current, template, "{template} {binary}");
            assert_eq!(v.latest.as_deref(), latest, "{template} {binary}");
        }
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

    /// A project the Rust state folder (`.moai/`, marker last) replaced, the Go `.moai/` kept whole
    /// under the given backup folders: each `(folder, relative path, text)` is written.
    fn moved_project(files: &[(&str, &str, &str)]) -> tempfile::TempDir {
        let tmp = tempfile::tempdir().unwrap();
        std::fs::create_dir_all(tmp.path().join(".moai")).unwrap();
        std::fs::write(
            tmp.path().join(".moai/moai.json"),
            crate::state_root::marker_bytes("test", &[]),
        )
        .unwrap();
        for (folder, rel, text) in files {
            let path = tmp.path().join(folder).join(rel);
            std::fs::create_dir_all(path.parent().unwrap()).unwrap();
            std::fs::write(path, text).unwrap();
        }
        tmp
    }

    fn display_in(tmp: &tempfile::TempDir) -> DisplayConfig {
        display_config(None, &Context::for_test(&[], Some(tmp.path())))
    }

    const GO_YAML: &str = "config/sections/statusline.yaml";

    /// #2936 S1g: the Go values the move took away are read from the Go backup, newest first.
    #[test]
    fn a_moved_go_statusline_yaml_is_read_from_the_newest_go_backup() {
        let one = moved_project(&[(
            ".moai-backup",
            GO_YAML,
            "statusline:\n  theme: older\n  segments:\n    model: false\n",
        )]);
        let c = display_in(&one);
        assert_eq!(c.theme, "older");
        assert_eq!(c.segments.get("model"), Some(&false));

        let two = moved_project(&[
            (".moai-backup", GO_YAML, "statusline:\n  theme: older\n"),
            (
                ".moai-backup-20261116T050000Z",
                GO_YAML,
                "statusline:\n  theme: newer\n",
            ),
            (
                ".moai-backup-20261115T050000Z",
                GO_YAML,
                "statusline:\n  theme: middle\n",
            ),
        ]);
        assert_eq!(display_in(&two).theme, "newer");
        // The newest backup decides even where it lacks the file: nothing is merged across backups.
        let lacking = moved_project(&[
            (".moai-backup", GO_YAML, "statusline:\n  theme: older\n"),
            (".moai-backup-20261116T050000Z", "manifest.json", "{}"),
        ]);
        assert_eq!(display_in(&lacking).theme, "");
    }

    #[test]
    fn a_file_in_the_new_state_folder_still_wins_whole() {
        let tmp = moved_project(&[(
            ".moai-backup",
            GO_YAML,
            "statusline:\n  theme: go\n  segments:\n    git: false\n",
        )]);
        let rc = tmp.path().join(".moai/config/sections");
        std::fs::create_dir_all(&rc).unwrap();
        std::fs::write(rc.join("statusline.yaml"), "statusline:\n  theme: rust\n").unwrap();
        let c = display_in(&tmp);
        assert_eq!(c.theme, "rust");
        assert!(c.segments.is_empty());
        // An unreadable new file means defaults, not the backup's values.
        std::fs::write(rc.join("statusline.yaml"), "statusline:\n\ttheme: x\n").unwrap();
        let c = display_in(&tmp);
        assert_eq!(c.theme, "");
        assert!(c.segments.is_empty());
    }

    /// A Go `.moai/` that is still in place (nothing moved, or the folder stays) is the Go file the
    /// reader always used; an older backup beside it changes nothing.
    #[test]
    fn an_intact_go_folder_is_read_before_any_backup() {
        let tmp = tempfile::tempdir().unwrap();
        for (folder, theme) in [(".moai", "go"), (".moai-backup", "backup")] {
            let dir = tmp.path().join(folder).join("config/sections");
            std::fs::create_dir_all(&dir).unwrap();
            std::fs::write(
                dir.join("statusline.yaml"),
                format!("statusline:\n  theme: {theme}\n"),
            )
            .unwrap();
        }
        assert_eq!(display_in(&tmp).theme, "go");
    }

    #[test]
    fn only_a_real_go_backup_folder_is_read() {
        let tmp = moved_project(&[
            (".moai-backups", GO_YAML, "statusline:\n  theme: plural\n"),
            (".moai-backup-old", GO_YAML, "statusline:\n  theme: named\n"),
            (
                ".moai-backup-20261116T050000",
                GO_YAML,
                "statusline:\n  theme: short\n",
            ),
        ]);
        assert_eq!(display_in(&tmp).theme, "");
        // A file called `.moai-backup` is not a backup folder.
        let file = moved_project(&[]);
        std::fs::write(file.path().join(".moai-backup"), "not a folder").unwrap();
        assert_eq!(display_in(&file).theme, "");
    }

    #[cfg(unix)]
    #[test]
    fn a_linked_go_backup_folder_is_not_read() {
        let elsewhere = moved_project(&[("real", GO_YAML, "statusline:\n  theme: linked\n")]);
        let tmp = moved_project(&[]);
        std::os::unix::fs::symlink(
            elsewhere.path().join("real"),
            tmp.path().join(".moai-backup"),
        )
        .unwrap();
        assert_eq!(display_in(&tmp).theme, "");
    }

    #[test]
    fn the_local_file_and_the_forge_follow_the_same_backup_and_nothing_is_written() {
        let tmp = moved_project(&[
            (
                ".moai-backup",
                GO_YAML,
                "statusline:\n  theme: kept\n  forge: github\n",
            ),
            (
                ".moai-backup",
                "statusline.local.yaml",
                "statusline:\n  forge: none\n",
            ),
        ]);
        // The local file is read first, then the shared one: `none` suppresses the counts.
        assert!(github_counts(tmp.path()).suppressed);
        assert_eq!(display_in(&tmp).theme, "kept");
        let listing = |root: &Path| {
            let mut all = Vec::new();
            let mut stack = vec![root.to_path_buf()];
            while let Some(dir) = stack.pop() {
                for e in std::fs::read_dir(&dir).unwrap() {
                    let p = e.unwrap().path();
                    if p.is_dir() {
                        stack.push(p.clone());
                    }
                    all.push((p.clone(), std::fs::read(&p).ok()));
                }
            }
            all.sort();
            all
        };
        let before = listing(tmp.path());
        let _ = display_in(&tmp);
        let _ = github_counts(tmp.path());
        assert_eq!(listing(tmp.path()), before);
    }

    /// AUDIT #3125 S1: the backup's local file speaks only while the new folder has no shared file;
    /// a new `statusline.yaml` that says `forge: github` is not outvoted by an old `forge: none`.
    #[test]
    fn a_backup_local_file_never_outvotes_a_new_shared_file() {
        let tmp = moved_project(&[(
            ".moai-backup",
            "statusline.local.yaml",
            "statusline:\n  forge: none\n",
        )]);
        assert!(github_counts(tmp.path()).suppressed);
        let rc = tmp.path().join(".moai/config/sections");
        std::fs::create_dir_all(&rc).unwrap();
        std::fs::write(rc.join("statusline.yaml"), "statusline:\n  forge: github\n").unwrap();
        assert!(!github_counts(tmp.path()).suppressed);
    }

    /// AUDIT #3125 S2: a backup is read only when the state folder is the new `.moai`; a project that
    /// still keeps its state in `.moai-cli` and happens to have a backup folder reads nothing there.
    #[test]
    fn a_backup_is_read_only_after_the_move_to_the_new_state_folder() {
        let tmp = tempfile::tempdir().unwrap();
        // The project is found by its `.moai/` folder, which has no Rust marker: the state is still
        // in the earlier folder, and the backup beside it is nobody's business.
        std::fs::create_dir_all(tmp.path().join(".moai")).unwrap();
        let dir = tmp.path().join(".moai-backup/config/sections");
        std::fs::create_dir_all(&dir).unwrap();
        std::fs::write(
            dir.join("statusline.yaml"),
            "statusline:\n  theme: backup\n",
        )
        .unwrap();
        assert_eq!(display_in(&tmp).theme, "");
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
        assert_eq!(main_checkout_of(&main), Some(canon.clone()));
        // #2928: the common dir under both, the folder the lane registry lives in.
        assert_eq!(common_dir_of(&wt.join("sub")), Some(canon.join(".git")));
        assert_eq!(common_dir_of(&main), Some(canon.join(".git")));
        assert_eq!(common_dir_of(tmp.path()), None, "no repository");
    }

    /// #2811 AC-3: real `git worktree add` worktrees in the new home place and in the Go
    /// generation's folder, both outside the main checkout, anchor to the main checkout.
    #[test]
    fn anchor_follows_home_and_go_place_worktrees_to_the_main_checkout() {
        let tmp = tempfile::tempdir().unwrap();
        let root = std::fs::canonicalize(tmp.path()).unwrap();
        let main = root.join("repo");
        std::fs::create_dir_all(&main).unwrap();
        git_init(&main);
        let git = |args: &[&str]| {
            let mut cmd = Command::new("git");
            cmd.args([
                "-c",
                "user.name=t",
                "-c",
                "user.email=t@example.invalid",
                "-c",
                "commit.gpgsign=false",
            ])
            .args(args)
            .current_dir(&main);
            crate::identity::scrub_git_env(&mut cmd);
            assert!(cmd.status().unwrap().success(), "{args:?}");
        };
        git(&["commit", "-q", "--allow-empty", "-m", "init"]);
        let home = root.join("home");
        for (wt, branch) in [
            (home.join(".moai/worktrees/repo-1-x"), "issue/1-x"),
            (home.join(".moai/worktrees/projects/repo/feat"), "feat"),
        ] {
            git(&["worktree", "add", "-q", "-b", branch, wt.to_str().unwrap()]);
            std::fs::create_dir_all(wt.join("sub")).unwrap();
            assert_eq!(
                main_checkout_of(&wt.join("sub")),
                Some(main.clone()),
                "{wt:?}"
            );
        }
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

    /// #2928 S-2/S-3 (decision D1 = (A)): the session is the leader only when the stdin
    /// `session_name` is a valid lane name and the clone's registry holds a live entry for it
    /// with mode `leader`; the registry is found from the session's directory, in the main
    /// checkout or in a linked worktree. Every other case — no entry, an ended lease, mode
    /// `agent`, a title or escape text, no directory — is the agent.
    #[test]
    fn the_leader_tag_needs_a_live_leader_entry_for_a_valid_name() {
        let tmp = tempfile::tempdir().unwrap();
        let main = tmp.path().join("main");
        let sessions = main.join(".git/.moai-cli/sessions");
        std::fs::create_dir_all(&sessions).unwrap();
        let gitdir = main.join(".git/worktrees/wt");
        std::fs::create_dir_all(&gitdir).unwrap();
        std::fs::write(gitdir.join("commondir"), "../..\n").unwrap();
        let wt = tmp.path().join("wt");
        std::fs::create_dir_all(&wt).unwrap();
        std::fs::write(wt.join(".git"), format!("gitdir: {}\n", gitdir.display())).unwrap();
        let now = 1_000;
        let put = |name: &str, lease: i64, mode: &str| {
            let entry = json!({
                "schema_version": 1, "name": name, "instance": "1".repeat(32),
                "lease_until": lease, "registered_at": 0, "mode": mode,
            });
            std::fs::write(sessions.join(format!("{name}.json")), entry.to_string()).unwrap();
        };
        let leader = |name: &str, dir: &Path| {
            collect_json(
                json!({"session_name": name, "workspace": {"current_dir": dir}}),
                now,
            )
            .leader
        };
        put("leader", now + 1, "leader");
        put("agent-3", now + 1, "agent");
        put("old", now, "leader");
        assert!(leader("leader", &main), "live leader entry, main checkout");
        assert!(leader("leader", &wt), "live leader entry, linked worktree");
        assert!(
            collect_json(json!({"session_name": "leader", "cwd": main}), now).leader,
            "the cwd when there is no workspace"
        );
        assert!(!leader("agent-3", &main), "mode agent");
        assert!(!leader("old", &main), "the lease ended at this second");
        assert!(!leader("nobody", &main), "no entry");
        for title in [
            "\u{1b}]0;leader\u{7}",
            "leader\u{9b}",
            "#2388·#2403 Issue 닫기 대조",
            "/moai",
            "",
        ] {
            assert!(!leader(title, &main), "{title:?}");
        }
        assert!(
            !collect_json(json!({"session_name": "leader"}), now).leader,
            "no directory: no registry"
        );
        let elsewhere = tmp.path().join("elsewhere");
        std::fs::create_dir_all(&elsewhere).unwrap();
        assert!(!leader("leader", &elsewhere), "outside the repository");
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

/// Card C-65: the card counts read from a ledger in a temporary HOME; the real HOME is never
/// touched.
#[cfg(test)]
mod card_counts_tests {
    use super::*;

    struct Fixture {
        _tmp: tempfile::TempDir,
        home: PathBuf,
        project: PathBuf,
    }

    /// A HOME and a project `demo` with its profile; no ledger yet.
    fn fixture() -> Fixture {
        let tmp = tempfile::tempdir().unwrap();
        let home = tmp.path().join("home");
        let project = tmp.path().join("work/demo");
        let state = project.join(crate::PROJECT_STATE_DIR);
        std::fs::create_dir_all(&home).unwrap();
        std::fs::create_dir_all(&state).unwrap();
        std::fs::write(
            state.join("project.yaml"),
            "{\"schema_version\":2,\"project_id\":\"demo\"}\n",
        )
        .unwrap();
        Fixture {
            _tmp: tmp,
            home,
            project,
        }
    }

    impl Fixture {
        fn ctx(&self) -> Context {
            Context::for_test(&[("HOME", self.home.to_str().unwrap())], None)
        }

        fn counts(&self) -> Option<CardCounts> {
            card_counts(&self.project, &self.ctx())
        }

        fn ledger_file(&self) -> PathBuf {
            crate::ledger::ledger_path(&self.home, "demo").unwrap()
        }

        /// Adds `n` cards in `state`. An `INSERT` has no transition check, so every state can
        /// be written directly.
        fn add(&self, state: &str, n: usize) {
            let mut ledger = crate::ledger::open_write(&self.home, "demo").unwrap();
            ledger
                .with_write_tx(|tx, now| {
                    for _ in 0..n {
                        tx.execute(
                            "INSERT INTO card (title, state, created_by, created_at, updated_at, closed_at) \
                             VALUES ('t', ?1, 'test', ?2, ?2, iif(?1 = 'done', ?2, NULL))",
                            rusqlite::params![state, now],
                        )?;
                    }
                    Ok(())
                })
                .unwrap();
        }

        /// Every entry under the HOME with its bytes, sorted: equal before and after a read
        /// means no file, folder or journal was made and no byte changed.
        fn home_tree(&self) -> Vec<(PathBuf, Vec<u8>)> {
            let mut out = Vec::new();
            let mut stack = vec![self.home.clone()];
            while let Some(dir) = stack.pop() {
                for entry in std::fs::read_dir(&dir).unwrap() {
                    let path = entry.unwrap().path();
                    if path.is_dir() {
                        out.push((path.clone(), Vec::new()));
                        stack.push(path);
                    } else {
                        let bytes = std::fs::read(&path).unwrap();
                        out.push((path, bytes));
                    }
                }
            }
            out.sort();
            out
        }
    }

    /// Counts with no card waiting for integration.
    fn counts(in_progress: i64, remaining: i64, done: i64) -> Option<CardCounts> {
        counts_w(in_progress, 0, remaining, done)
    }

    fn counts_w(in_progress: i64, waiting: i64, remaining: i64, done: i64) -> Option<CardCounts> {
        Some(CardCounts {
            in_progress,
            waiting: Some(waiting),
            remaining,
            done,
        })
    }

    /// Adds one card in `state` and its notes oldest first, `(kind, body)` each; the note ids
    /// grow with the order, which is what "newest" means.
    fn add_with_notes(f: &Fixture, state: &str, notes: &[(&str, &str)]) {
        let mut ledger = crate::ledger::open_write(&f.home, "demo").unwrap();
        ledger
            .with_write_tx(|tx, now| {
                tx.execute(
                    "INSERT INTO card (title, state, created_by, created_at, updated_at, closed_at) \
                     VALUES ('t', ?1, 'test', ?2, ?2, iif(?1 = 'done', ?2, NULL))",
                    rusqlite::params![state, now],
                )?;
                let card_id = tx.last_insert_rowid();
                for (kind, body) in notes {
                    tx.execute(
                        "INSERT INTO card_note (card_id, kind, author, body, created_at) \
                         VALUES (?1, ?2, 'test', ?3, ?4)",
                        rusqlite::params![card_id, kind, body, now],
                    )?;
                }
                Ok(())
            })
            .unwrap();
    }

    /// The reading of one `active` card with the given notes: `(working, waiting)`.
    fn active_card(notes: &[(&str, &str)]) -> (i64, Option<i64>) {
        let f = fixture();
        crate::ledger::register(&f.home, "demo").unwrap();
        add_with_notes(&f, "active", notes);
        let c = f.counts().unwrap();
        (c.in_progress, c.waiting)
    }

    const READY: &str = "LANE-READY-INTEGRATION C-1 head=abc spec=SPEC-20261119-x";
    const ACK: &str = "ACK agent-1 C-1 2026-11-19T00:00:00Z\naccepted";
    const FAIL: &str = "AUDIT C-1 (x) head abc base def — verdict: FAIL";
    const PASS: &str = "AUDIT C-1 (x) head abc base def — verdict: PASS";

    #[test]
    fn a_review_card_is_waiting_and_an_active_card_without_notes_is_working() {
        let f = fixture();
        crate::ledger::register(&f.home, "demo").unwrap();
        add_with_notes(&f, "review", &[]);
        add_with_notes(&f, "active", &[]);
        assert_eq!(f.counts(), counts_w(1, 1, 0, 0));
    }

    #[test]
    fn an_active_card_is_waiting_by_its_newest_marker_note() {
        // The waiting markers, each as the first line of a note.
        for line in [
            READY,
            "LANE-DONE agent-1 C-1 PR 없음 — pushed head abc — audit requested (08:00Z)",
            "LANE-DONE agent-1 C-1 pushed head abc — 감사 요청 tier 2",
            "감사 요청 tier 2 head=abc spec=SPEC-20261119-x",
            PASS,
        ] {
            assert_eq!(active_card(&[("report", line)]), (0, Some(1)), "{line}");
        }
        // Newest wins: a working marker after a waiting one puts the card back to work, and a
        // waiting one after a working one sends it back to waiting.
        assert_eq!(
            active_card(&[("report", READY), ("comment", ACK)]),
            (1, Some(0))
        );
        assert_eq!(
            active_card(&[("report", READY), ("decision", FAIL)]),
            (1, Some(0))
        );
        assert_eq!(
            active_card(&[("report", "PLAN C-1\nscope"), ("report", PASS)]),
            (0, Some(1))
        );
        assert_eq!(
            active_card(&[("decision", FAIL), ("report", READY)]),
            (0, Some(1))
        );
    }

    #[test]
    fn notes_that_are_no_marker_do_not_change_the_reading() {
        // A neutral note newer than the waiting marker leaves the card waiting.
        assert_eq!(
            active_card(&[
                ("report", READY),
                ("report", "waiting: audit since 08:00Z"),
                ("decision", "LEADER 2026-11-19 — 감사 중, 받지 말 것"),
            ]),
            (0, Some(1))
        );
        // A neutral note newer than the working one leaves the card working.
        assert_eq!(
            active_card(&[("comment", ACK), ("report", "RESUME head abc")]),
            (1, Some(0))
        );
        // `LANE-DONE` without an audit request is no waiting marker.
        assert_eq!(
            active_card(&[("report", "LANE-DONE agent-1 C-1 slice 1 done")]),
            (1, Some(0))
        );
        // The words that make a `LANE-DONE` or an `AUDIT` line a marker must sit on that first
        // line: the same words on a later line change nothing.
        for body in [
            "LANE-DONE agent-1 C-1 slice 1 done\nnext: audit requested",
            "LANE-DONE agent-1 C-1 slice 1 done\n감사 요청 tier 2",
            "AUDIT C-1 (x) head abc base def\nverdict: PASS",
        ] {
            assert_eq!(active_card(&[("report", body)]), (1, Some(0)), "{body}");
        }
        assert_eq!(
            active_card(&[
                ("report", READY),
                ("report", "AUDIT C-1 head abc\nverdict: FAIL")
            ]),
            (0, Some(1)),
            "a FAIL verdict only on a later line does not put the card back to work"
        );
        // Only the first line counts: a marker on the second line is no marker.
        assert_eq!(
            active_card(&[("report", &format!("summary\n{READY}"))]),
            (1, Some(0))
        );
        // The marker must start the line, not sit inside it.
        assert_eq!(
            active_card(&[("report", &format!("see {READY}"))]),
            (1, Some(0))
        );
    }

    #[test]
    fn only_active_and_review_cards_are_split() {
        let f = fixture();
        crate::ledger::register(&f.home, "demo").unwrap();
        // A waiting marker on a card in another state changes no number but its own state's.
        for state in ["draft", "ready", "hold", "done"] {
            add_with_notes(&f, state, &[("report", READY)]);
        }
        assert_eq!(f.counts(), counts_w(0, 0, 3, 1));
    }

    #[test]
    fn the_split_adds_up_across_a_mix_and_the_read_changes_nothing() {
        let f = fixture();
        crate::ledger::register(&f.home, "demo").unwrap();
        add_with_notes(&f, "active", &[("report", READY)]);
        add_with_notes(&f, "active", &[("comment", ACK)]);
        add_with_notes(&f, "active", &[]);
        f.add("review", 2);
        f.add("draft", 20);
        f.add("ready", 25);
        f.add("hold", 5);
        f.add("done", 12);
        let before = f.home_tree();
        let digest = crate::digest(&std::fs::read(f.ledger_file()).unwrap());
        // 2 working (active), 1 + 2 waiting (active by marker, review), 50 remaining, 12 done.
        assert_eq!(f.counts(), counts_w(2, 3, 50, 12));
        assert_eq!(
            crate::digest(&std::fs::read(f.ledger_file()).unwrap()),
            digest,
            "the ledger file is byte for byte the same"
        );
        assert_eq!(f.home_tree(), before, "no journal, no new file");
    }

    #[test]
    fn a_failed_note_read_gives_the_three_number_form() {
        let f = fixture();
        crate::ledger::register(&f.home, "demo").unwrap();
        add_with_notes(&f, "active", &[("report", READY)]);
        f.add("review", 1);
        let ledger =
            crate::ledger::open::open_read_only_within(&f.home, "demo", Duration::from_millis(100))
                .unwrap();
        // The render deadline has passed: no reading of the notes.
        let spent = Instant::now() - Duration::from_millis(1);
        assert_eq!(waiting_cards(&ledger, Some(spent)), None);
        // With time left the same ledger reads: the active card and the review card wait.
        let left = Instant::now() + Duration::from_secs(60);
        assert_eq!(waiting_cards(&ledger, Some(left)), Some(2));
        assert_eq!(waiting_cards(&ledger, None), Some(2));
    }

    /// A deadline that passes in the middle of the row loop (not only before it) gives `None`:
    /// the first question is "no", the next ones are "yes".
    #[test]
    fn a_deadline_that_passes_between_rows_gives_the_three_number_form() {
        let f = fixture();
        crate::ledger::register(&f.home, "demo").unwrap();
        add_with_notes(&f, "active", &[("report", READY), ("comment", "plain")]);
        let ledger =
            crate::ledger::open::open_read_only_within(&f.home, "demo", Duration::from_millis(100))
                .unwrap();
        let asked = std::cell::Cell::new(0_u32);
        let spent = || {
            asked.set(asked.get() + 1);
            asked.get() > 1
        };
        assert_eq!(waiting_cards_with(&ledger, &spent), None);
        assert!(asked.get() >= 2, "the loop asked again after the first row");
        // Never spent: the same ledger reads.
        assert_eq!(waiting_cards_with(&ledger, &|| false), Some(1));
    }

    /// Several waiting notes on one card count that card once, and a waiting note keeps
    /// deciding even when an older working note sits under it.
    #[test]
    fn several_waiting_notes_on_one_card_count_it_once() {
        assert_eq!(
            active_card(&[("report", READY), ("report", READY), ("report", PASS)]),
            (0, Some(1))
        );
        assert_eq!(
            active_card(&[("comment", ACK), ("report", READY), ("report", READY)]),
            (0, Some(1))
        );
    }

    /// One line carrying both verdicts is read by the first word that decides it: `PASS` is
    /// tested before `FAIL`, so a FAIL line that quotes an earlier PASS reads as waiting. The
    /// real verdict line ends at the verdict word, so this only pins the rule.
    #[test]
    fn a_line_with_both_verdicts_reads_as_waiting() {
        let both = "AUDIT C-1 (x) head abc base def — verdict: FAIL (was verdict: PASS)";
        assert_eq!(marker_of(both), Some(Marker::Waiting));
        assert_eq!(marker_of(FAIL), Some(Marker::Working));
        assert_eq!(marker_of(PASS), Some(Marker::Waiting));
    }

    /// Only the first `NOTE_HEAD_CHARS` characters of a note are read: a `LANE-DONE` line whose
    /// `audit requested` ends at that character still waits, one character later it is cut off
    /// and the card keeps working. Korean text counts by characters, not bytes.
    #[test]
    fn the_note_head_cut_is_at_exactly_one_thousand_characters() {
        let line = |pad: usize, key: &str| {
            let mut s = String::from("LANE-DONE agent-1 C-1 ");
            s.push_str(&"x".repeat(pad));
            s.push(' ');
            s.push_str(key);
            s
        };
        for key in ["audit requested", "감사 요청"] {
            let base = line(0, key).chars().count();
            // The literal 1000 on purpose: the test must fail when the constant moves.
            let edge = line(1000 - base, key);
            assert_eq!(edge.chars().count(), 1000);
            assert_eq!(
                active_card(&[("report", &edge)]),
                (0, Some(1)),
                "{key} at the cut"
            );
            let past = line(1000 - base + 1, key);
            assert_eq!(
                active_card(&[("report", &past)]),
                (1, Some(0)),
                "{key} one past"
            );
        }
    }

    /// The working number never goes below zero: a card that moved between the state count and
    /// the note read must not show `▶️ -1`.
    #[test]
    fn the_working_number_never_goes_below_zero() {
        assert_eq!(split_in_progress(5, 2), 3);
        assert_eq!(split_in_progress(2, 2), 0);
        assert_eq!(split_in_progress(1, 3), 0);
    }

    #[test]
    fn no_registered_ledger_gives_no_counts_and_creates_nothing() {
        let f = fixture();
        assert_eq!(f.counts(), None);
        assert_eq!(f.home_tree(), Vec::new(), "the HOME stays empty");
    }

    #[test]
    fn an_empty_ledger_counts_zero_which_is_not_absent() {
        let f = fixture();
        crate::ledger::register(&f.home, "demo").unwrap();
        assert_eq!(f.counts(), counts(0, 0, 0));
    }

    /// Each state lands in exactly one number: `active` and `review` in progress, `draft`,
    /// `ready` and `hold` remaining, `done` only in done.
    #[test]
    fn each_state_is_counted_in_exactly_one_number() {
        for (state, want) in [
            ("active", counts(1, 0, 0)),
            ("review", counts_w(0, 1, 0, 0)),
            ("draft", counts(0, 1, 0)),
            ("ready", counts(0, 1, 0)),
            ("hold", counts(0, 1, 0)),
            ("done", counts(0, 0, 1)),
        ] {
            let f = fixture();
            crate::ledger::register(&f.home, "demo").unwrap();
            f.add(state, 1);
            assert_eq!(f.counts(), want, "{state}");
        }
    }

    #[test]
    fn a_mix_of_states_adds_up_and_the_read_changes_nothing() {
        let f = fixture();
        crate::ledger::register(&f.home, "demo").unwrap();
        for (state, n) in [
            ("active", 3),
            ("review", 2),
            ("draft", 20),
            ("ready", 25),
            ("hold", 5),
            ("done", 12),
        ] {
            f.add(state, n);
        }
        let before = f.home_tree();
        let digest = crate::digest(&std::fs::read(f.ledger_file()).unwrap());
        // 3 `active` without notes are working, the 2 `review` cards wait.
        assert_eq!(f.counts(), counts_w(3, 2, 50, 12));
        assert_eq!(
            crate::digest(&std::fs::read(f.ledger_file()).unwrap()),
            digest,
            "the ledger file is byte for byte the same"
        );
        assert_eq!(f.home_tree(), before, "no journal, no new file");
    }

    /// A ledger another process holds exclusively: no counts, and the wait is the short one,
    /// far below the ledger's own five seconds.
    #[test]
    fn a_locked_ledger_gives_no_counts_within_the_short_wait() {
        let f = fixture();
        crate::ledger::register(&f.home, "demo").unwrap();
        f.add("active", 1);
        let holder = rusqlite::Connection::open(f.ledger_file()).unwrap();
        holder.execute_batch("BEGIN EXCLUSIVE").unwrap();
        let started = Instant::now();
        assert_eq!(f.counts(), None);
        let waited = started.elapsed();
        assert!(
            waited < crate::ledger::open::BUSY_TIMEOUT / 2,
            "waited {waited:?}"
        );
        holder.execute_batch("ROLLBACK").unwrap();
        assert_eq!(f.counts(), counts(1, 0, 0), "readable again once released");
    }

    #[test]
    fn a_file_that_is_not_this_projects_ledger_gives_no_counts_and_is_left_alone() {
        // Not a database at all, a database cut short, and another project's ledger.
        let other = fixture();
        crate::ledger::register(&other.home, "other").unwrap();
        let other_bytes =
            std::fs::read(crate::ledger::ledger_path(&other.home, "other").unwrap()).unwrap();
        for bytes in [
            b"not a database".to_vec(),
            other_bytes[..200].to_vec(),
            other_bytes,
        ] {
            let f = fixture();
            crate::ledger::register(&f.home, "demo").unwrap();
            assert_eq!(f.counts(), counts(0, 0, 0));
            std::fs::write(f.ledger_file(), &bytes).unwrap();
            let before = f.home_tree();
            assert_eq!(f.counts(), None);
            assert_eq!(f.home_tree(), before);
        }
    }

    #[test]
    fn no_home_no_profile_or_a_spent_budget_gives_no_counts() {
        let f = fixture();
        crate::ledger::register(&f.home, "demo").unwrap();
        assert_eq!(f.counts(), counts(0, 0, 0));
        // HOME unset.
        let no_home = Context::for_test(&[], None);
        assert_eq!(card_counts(&f.project, &no_home), None);
        // A HOME that is not there.
        let gone = f.home.join("gone");
        let missing = Context::for_test(&[("HOME", gone.to_str().unwrap())], None);
        assert_eq!(card_counts(&f.project, &missing), None);
        // The render deadline has passed.
        let mut late = f.ctx();
        late.deadline = Some(Instant::now() - Duration::from_millis(1));
        assert_eq!(card_counts(&f.project, &late), None);
        late.deadline = Some(Instant::now() + Duration::from_secs(60));
        assert_eq!(card_counts(&f.project, &late), counts(0, 0, 0));
        // A project without a profile, and a folder that does not exist.
        std::fs::remove_file(
            f.project
                .join(crate::PROJECT_STATE_DIR)
                .join("project.yaml"),
        )
        .unwrap();
        assert_eq!(f.counts(), None);
        assert_eq!(card_counts(&f.project.join("absent"), &f.ctx()), None);
    }

    /// `collect` reads the counts for the project the input names (`workspace.project_dir`),
    /// whatever the working directory is, and falls back to the project above the cwd.
    #[test]
    fn collect_reads_the_counts_of_the_anchor_or_of_the_cwd_project() {
        let f = fixture();
        crate::ledger::register(&f.home, "demo").unwrap();
        f.add("ready", 2);
        f.add("done", 1);
        let elsewhere = f.home.parent().unwrap().to_path_buf();
        let home = f.home.to_str().unwrap();
        let named = input::parse(
            serde_json::json!({"workspace": {"project_dir": f.project}})
                .to_string()
                .as_bytes(),
        )
        .unwrap();
        let ctx = Context::for_test(&[("HOME", home)], Some(&elsewhere));
        assert_eq!(collect(Some(&named), &ctx).cards, counts(0, 2, 1));
        let empty = input::parse(b"{}").unwrap();
        assert_eq!(collect(Some(&empty), &ctx).cards, None, "no project here");
        std::fs::create_dir_all(f.project.join(".moai")).unwrap();
        let inside = Context::for_test(&[("HOME", home)], Some(&f.project));
        assert_eq!(collect(Some(&empty), &inside).cards, counts(0, 2, 1));
    }
}

/// #2756 leader decision 6862179411 item 3: the project root search stops at the git
/// repository, and HOME's `.moai` is never a project marker.
#[cfg(test)]
mod project_root_tests {
    use super::*;
    use std::fs;

    fn root_from(home: &Path, cwd: &Path) -> Option<PathBuf> {
        let home = home.to_str().unwrap();
        find_project_root(&Context::for_test(&[("HOME", home)], Some(cwd)))
    }

    #[test]
    fn a_moai_folder_above_the_git_repository_is_not_the_project() {
        // agent-29 case F (6862215957): outside HOME, a non-git ancestor holds `.moai`.
        let w = tempfile::tempdir().unwrap();
        let w = fs::canonicalize(w.path()).unwrap();
        fs::create_dir_all(w.join("home")).unwrap();
        fs::create_dir_all(w.join("outside/.moai")).unwrap();
        fs::create_dir_all(w.join("outside/repo/.git")).unwrap();
        fs::create_dir_all(w.join("outside/repo/src")).unwrap();
        assert_eq!(
            root_from(&w.join("home"), &w.join("outside/repo/src")),
            None
        );
    }

    #[test]
    fn homes_moai_is_never_the_project() {
        // agent-29 case E: `~/.moai` exists; a repository below HOME and a plain folder below
        // HOME both find no project.
        let w = tempfile::tempdir().unwrap();
        let home = fs::canonicalize(w.path()).unwrap();
        fs::create_dir_all(home.join(".moai")).unwrap();
        fs::create_dir_all(home.join("proj/repo/.git")).unwrap();
        fs::create_dir_all(home.join("notes")).unwrap();
        assert_eq!(root_from(&home, &home.join("proj/repo")), None);
        assert_eq!(root_from(&home, &home.join("notes")), None);
        assert_eq!(root_from(&home, &home), None);
    }

    #[test]
    fn the_repository_that_holds_moai_is_found_from_below() {
        let w = tempfile::tempdir().unwrap();
        let w = fs::canonicalize(w.path()).unwrap();
        fs::create_dir_all(w.join("home")).unwrap();
        fs::create_dir_all(w.join("repo/.git")).unwrap();
        fs::create_dir_all(w.join("repo/.moai")).unwrap();
        fs::create_dir_all(w.join("repo/a/b")).unwrap();
        assert_eq!(
            root_from(&w.join("home"), &w.join("repo/a/b")),
            Some(w.join("repo"))
        );
        // A linked worktree has a `.git` file; the search stops there too.
        fs::create_dir_all(w.join("repo/wt/sub")).unwrap();
        fs::write(w.join("repo/wt/.git"), "gitdir: ../.git/worktrees/wt\n").unwrap();
        assert_eq!(root_from(&w.join("home"), &w.join("repo/wt/sub")), None);
        fs::create_dir_all(w.join("repo/wt/.moai")).unwrap();
        assert_eq!(
            root_from(&w.join("home"), &w.join("repo/wt/sub")),
            Some(w.join("repo/wt"))
        );
    }
}

/// User instruction 2026-11-19: `🌳 WT-c<card>-<name>`. The name is a pure function of the
/// reservation's branch, the worktree folder name, the project id and the card number; the card
/// comes from the held reservation that names the worktree folder, read from the local ledger.
#[cfg(test)]
mod worktree_label_tests {
    use super::*;

    fn name(branch: Option<&str>, folder: &str, card: i64) -> String {
        label_name(branch, folder, "demo", card)
    }

    #[test]
    fn a_spec_branch_gives_its_slug() {
        assert_eq!(
            name(
                Some("spec/20261119-statusline-waiting-cards"),
                "demo-1-x",
                1
            ),
            "statusline-waiting-cards"
        );
        // Any kind before the slash, and no slash at all.
        assert_eq!(name(Some("fix/20261119-a"), "demo-1", 1), "a");
        assert_eq!(name(Some("20261119-a-b"), "demo-1", 1), "a-b");
        // The date must be eight digits followed by a dash and something.
        assert_eq!(name(Some("spec/2026111-a"), "demo-1-rest", 1), "rest");
        assert_eq!(name(Some("spec/20261119-"), "demo-1-rest", 1), "rest");
        assert_eq!(name(Some("spec/20261119"), "demo-1-rest", 1), "rest");
    }

    #[test]
    fn another_branch_takes_the_rest_of_the_folder_name() {
        let b = Some("m-flow/bundle-20261119-6");
        assert_eq!(name(b, "demo-1-hooks", 1), "hooks");
        assert_eq!(name(None, "demo-1-hooks", 1), "hooks");
        // The card number may carry its `c`.
        assert_eq!(name(b, "demo-c1-hooks", 1), "hooks");
        // The folder with nothing left gives nothing.
        assert_eq!(name(b, "demo-1", 1), "");
        assert_eq!(name(b, "demo", 1), "");
        // A number that is not this card's number stays; so does a folder of another project.
        assert_eq!(name(b, "demo-12-hooks", 1), "12-hooks");
        assert_eq!(name(b, "demo-audit-3230", 1), "audit-3230");
        assert_eq!(name(b, "other-hooks", 1), "other-hooks");
        // The prefix is the project id plus a dash, not any text that starts the same.
        assert_eq!(name(b, "demonstration-1-x", 1), "demonstration-1-x");
    }

    #[test]
    fn the_name_is_cut_at_thirty_two_characters_and_drops_control_characters() {
        let exact = "a".repeat(32);
        let long = "a".repeat(33);
        let branch = |s: &str| format!("spec/20261119-{s}");
        assert_eq!(name(Some(&branch(&exact)), "demo-1", 1), exact);
        assert_eq!(
            name(Some(&branch(&long)), "demo-1", 1),
            format!("{}…", "a".repeat(31))
        );
        // Characters, not bytes.
        let korean = "가".repeat(40);
        assert_eq!(
            name(Some(&branch(&korean)), "demo-1", 1),
            format!("{}…", "가".repeat(31))
        );
        assert_eq!(
            name(Some(&branch(&"가".repeat(32))), "demo-1", 1),
            "가".repeat(32)
        );
        assert_eq!(name(Some("spec/20261119-a\u{1b}b\nc"), "demo-1", 1), "abc");
        // Control characters are dropped before the length is counted.
        let padded = format!("{}\u{7}{}", "a".repeat(32), "");
        assert_eq!(name(Some(&branch(&padded)), "demo-1", 1), exact);
    }

    struct Fixture {
        _tmp: tempfile::TempDir,
        home: PathBuf,
        project: PathBuf,
        /// A worktree folder (a `.git` entry and a `src` folder), canonical.
        wt: PathBuf,
    }

    fn fixture() -> Fixture {
        let tmp = tempfile::tempdir().unwrap();
        let root = std::fs::canonicalize(tmp.path()).unwrap();
        let home = root.join("home");
        let project = root.join("work/demo");
        let state = project.join(crate::PROJECT_STATE_DIR);
        let wt = root.join("wt/demo-1-work");
        std::fs::create_dir_all(&home).unwrap();
        std::fs::create_dir_all(&state).unwrap();
        std::fs::create_dir_all(wt.join("src")).unwrap();
        std::fs::write(wt.join(".git"), "gitdir: nowhere\n").unwrap();
        std::fs::write(
            state.join("project.yaml"),
            "{\"schema_version\":2,\"project_id\":\"demo\"}\n",
        )
        .unwrap();
        crate::ledger::register(&home, "demo").unwrap();
        Fixture {
            _tmp: tmp,
            home,
            project,
            wt,
        }
    }

    impl Fixture {
        fn ctx(&self) -> Context {
            Context::for_test(&[("HOME", self.home.to_str().unwrap())], None)
        }

        fn label(&self, dir: &Path) -> Option<WorktreeLabel> {
            worktree_label(&self.project, dir, &self.ctx())
        }

        /// One card with no reservation; returns its number. Cards are numbered in the order
        /// they are made, so a later `reserve_card` can give the lower card its reservation
        /// after a higher one has had its own.
        fn add_card(&self) -> i64 {
            let mut ledger = crate::ledger::open_write(&self.home, "demo").unwrap();
            let mut card = 0;
            ledger
                .with_write_tx(|tx, now| {
                    tx.execute(
                        "INSERT INTO card (title, state, created_by, created_at, updated_at) \
                         VALUES ('t', 'active', 'test', ?1, ?1)",
                        rusqlite::params![now],
                    )?;
                    card = tx.last_insert_rowid();
                    Ok(())
                })
                .unwrap();
            card
        }

        /// A `held` reservation of the existing `card` naming `worktree` (any spelling).
        fn reserve_card(&self, card: i64, worktree: &str, branch: &str) {
            let mut ledger = crate::ledger::open_write(&self.home, "demo").unwrap();
            ledger
                .with_write_tx(|tx, now| {
                    tx.execute(
                        "INSERT INTO reservation (card_id, generation, holder, status, \
                         acquired_at, lease_until, updated_at, worktree, branch) \
                         VALUES (?1, 1, 'test', 'held', ?2, ?3, ?2, ?4, ?5)",
                        rusqlite::params![card, now, now + 14400, worktree, branch],
                    )?;
                    Ok(())
                })
                .unwrap();
        }

        /// One card in `card_state` and a reservation of it in `status` naming `worktree`.
        /// An `INSERT` has no transition check. Returns the card number.
        fn reserve(
            &self,
            card_state: &str,
            status: &str,
            worktree: &Path,
            branch: Option<&str>,
        ) -> i64 {
            let mut ledger = crate::ledger::open_write(&self.home, "demo").unwrap();
            let mut card = 0;
            ledger
                .with_write_tx(|tx, now| {
                    tx.execute(
                        "INSERT INTO card (title, state, created_by, created_at, updated_at) \
                         VALUES ('t', ?1, 'test', ?2, ?2)",
                        rusqlite::params![card_state, now],
                    )?;
                    card = tx.last_insert_rowid();
                    tx.execute(
                        "INSERT INTO reservation (card_id, generation, holder, status, \
                         acquired_at, lease_until, updated_at, worktree, branch) \
                         VALUES (?1, 1, 'test', ?2, ?3, ?4, ?3, ?5, ?6)",
                        rusqlite::params![
                            card,
                            status,
                            now,
                            now + 14400,
                            worktree.to_string_lossy(),
                            branch
                        ],
                    )?;
                    Ok(())
                })
                .unwrap();
            card
        }
    }

    fn label(card: i64, name: &str) -> Option<WorktreeLabel> {
        Some(WorktreeLabel {
            card,
            name: name.to_string(),
        })
    }

    #[test]
    fn a_held_reservation_gives_the_card_and_the_name_from_inside_the_folder() {
        let f = fixture();
        let card = f.reserve("active", "held", &f.wt, Some("spec/20261119-demo-work"));
        assert_eq!(f.label(&f.wt), label(card, "demo-work"));
        // From a folder inside the worktree, the way `workspace.current_dir` often is.
        assert_eq!(f.label(&f.wt.join("src")), label(card, "demo-work"));
    }

    #[test]
    fn only_a_reservation_that_still_holds_the_folder_counts() {
        for (status, counts) in [
            ("held", true),
            ("held_for_review", true),
            ("expired_candidate", true),
            ("released", false),
            ("reclaimed", false),
        ] {
            let f = fixture();
            let card = f.reserve("active", status, &f.wt, Some("spec/20261119-demo-work"));
            let want = counts.then(|| WorktreeLabel {
                card,
                name: "demo-work".into(),
            });
            assert_eq!(f.label(&f.wt), want, "status {status}");
        }
    }

    #[test]
    fn two_cards_holding_one_folder_give_the_lowest_number() {
        let f = fixture();
        let first = f.reserve("active", "held", &f.wt, Some("spec/20261119-first"));
        let _second = f.reserve("active", "held", &f.wt, Some("spec/20261119-second"));
        assert_eq!(f.label(&f.wt), label(first, "first"));
        // The lowest number among the ones that hold, not among all.
        let g = fixture();
        let _gone = g.reserve("ready", "released", &g.wt, Some("spec/20261119-gone"));
        let kept = g.reserve("active", "held", &g.wt, Some("spec/20261119-kept"));
        assert_eq!(g.label(&g.wt), label(kept, "kept"));
    }

    /// REQ-3 against the order the rows were written in: the lower card is reserved AFTER the
    /// higher one, so only the ordering of the read puts it first.
    #[test]
    fn the_lowest_card_wins_even_when_it_was_reserved_last() {
        let f = fixture();
        let low = f.add_card();
        let high = f.add_card();
        assert!(low < high);
        f.reserve_card(high, f.wt.to_str().unwrap(), "spec/20261119-high");
        f.reserve_card(low, f.wt.to_str().unwrap(), "spec/20261119-low");
        assert_eq!(f.label(&f.wt), label(low, "low"));
    }

    /// A recorded path that is not absolute names no folder from here: read against this
    /// process's working directory it could name the session's folder by accident. The path
    /// below is relative and, from this process, does resolve to the worktree.
    #[test]
    fn a_relative_recorded_path_never_matches() {
        let f = fixture();
        let here = std::env::current_dir().unwrap();
        let mut relative = PathBuf::new();
        for _ in here.components().skip(1) {
            relative.push("..");
        }
        relative.push(f.wt.strip_prefix("/").unwrap());
        assert!(relative.is_relative());
        assert_eq!(
            std::fs::canonicalize(&relative).unwrap(),
            f.wt,
            "the probe path resolves to the folder from this process"
        );
        let (by_walk, by_dot) = (f.add_card(), f.add_card());
        f.reserve_card(
            by_walk,
            relative.to_str().unwrap(),
            "spec/20261119-elsewhere",
        );
        f.reserve_card(by_dot, ".", "spec/20261119-dot");
        assert_eq!(f.label(&f.wt), None);
        // An absolute entry of a higher card is still found past the relative ones.
        let other = f.add_card();
        f.reserve_card(other, f.wt.to_str().unwrap(), "spec/20261119-real");
        assert_eq!(f.label(&f.wt), label(other, "real"));
    }

    /// The read stops at once when the budget runs out, even between two rows: the first row
    /// names another folder, the second would match.
    #[test]
    fn the_budget_is_asked_before_every_row() {
        let f = fixture();
        let elsewhere = f.wt.parent().unwrap().join("demo-2-other");
        std::fs::create_dir_all(&elsewhere).unwrap();
        let first = f.add_card();
        let second = f.add_card();
        f.reserve_card(first, elsewhere.to_str().unwrap(), "spec/20261119-other");
        f.reserve_card(second, f.wt.to_str().unwrap(), "spec/20261119-demo-work");
        let ledger = crate::ledger::open::open_read_only_within(
            &f.home,
            "demo",
            std::time::Duration::from_millis(50),
        )
        .unwrap();
        let asked = std::cell::Cell::new(0);
        let after_one_row = || {
            asked.set(asked.get() + 1);
            asked.get() > 1
        };
        assert_eq!(
            holding_card(ledger.conn(), &f.wt, &f.wt, &after_one_row),
            None
        );
        assert_eq!(asked.get(), 2, "asked again before the second row");
        let never = || false;
        assert_eq!(
            holding_card(ledger.conn(), &f.wt, &f.wt, &never).map(|(c, _)| c),
            Some(second)
        );
    }

    /// The folder the session works in: `workspace.current_dir`, else `cwd`, else none.
    #[test]
    fn the_session_folder_is_current_dir_then_cwd() {
        let parse = |json: &str| serde_json::from_str::<Input>(json).unwrap();
        let dir = |json: &str| session_dir(&parse(json));
        assert_eq!(
            dir(r#"{"cwd":"/c","workspace":{"current_dir":"/w"}}"#),
            Some(PathBuf::from("/w"))
        );
        assert_eq!(
            dir(r#"{"cwd":"/c","workspace":{"current_dir":""}}"#),
            Some(PathBuf::from("/c"))
        );
        assert_eq!(
            dir(r#"{"cwd":"/c","workspace":{}}"#),
            Some(PathBuf::from("/c"))
        );
        assert_eq!(dir(r#"{"cwd":"/c"}"#), Some(PathBuf::from("/c")));
        assert_eq!(
            dir(r#"{"workspace":{"current_dir":"/w"}}"#),
            Some(PathBuf::from("/w"))
        );
        assert_eq!(dir(r#"{"cwd":"","workspace":{"current_dir":""}}"#), None);
        assert_eq!(dir("{}"), None);
    }

    /// A name nothing visible is left of gives way to the folder name, and invisible format
    /// characters (zero-width, bidirectional marks) are dropped like control characters.
    #[test]
    fn a_name_with_nothing_visible_left_falls_to_the_folder_name() {
        let b = |slug: &str| Some(format!("spec/20261119-{slug}"));
        assert_eq!(
            label_name(b("\u{1b}\u{7}").as_deref(), "demo-1-hooks", "demo", 1),
            "hooks"
        );
        assert_eq!(
            label_name(b("\u{202e}\u{200b}").as_deref(), "demo-1-hooks", "demo", 1),
            "hooks"
        );
        assert_eq!(
            label_name(
                b("a\u{202e}b\u{200f}c\u{feff}").as_deref(),
                "demo-1",
                "demo",
                1
            ),
            "abc"
        );
        // Nothing visible in either source: no name, the card alone.
        assert_eq!(
            label_name(b("\u{7}").as_deref(), "demo-1-\u{202e}", "demo", 1),
            ""
        );
    }

    #[test]
    fn a_reservation_without_a_branch_takes_the_name_from_the_folder() {
        let f = fixture();
        let card = f.reserve("active", "held", &f.wt, None);
        // The folder is `demo-1-work` and the card number is 1: `work` is left.
        assert_eq!(f.label(&f.wt), label(card, "work"));
    }

    #[test]
    fn no_reservation_for_the_folder_gives_no_label() {
        let f = fixture();
        assert_eq!(f.label(&f.wt), None, "an empty ledger");
        let other = f.wt.parent().unwrap().join("demo-2-other");
        std::fs::create_dir_all(&other).unwrap();
        std::fs::write(other.join(".git"), "gitdir: nowhere\n").unwrap();
        f.reserve("active", "held", &other, Some("spec/20261119-other"));
        assert_eq!(f.label(&f.wt), None, "a reservation of another folder");
        // A folder with no `.git` entry anywhere above it is not a worktree.
        let bare = f._tmp.path().join("bare");
        std::fs::create_dir_all(&bare).unwrap();
        assert_eq!(f.label(&bare), None);
    }

    #[test]
    fn a_ledger_that_cannot_be_read_gives_no_label() {
        let f = fixture();
        f.reserve("active", "held", &f.wt, Some("spec/20261119-demo-work"));
        // No HOME, no ledger to read.
        let no_home = Context::for_test(&[], None);
        assert_eq!(worktree_label(&f.project, &f.wt, &no_home), None);
        // A project with no profile.
        assert_eq!(
            worktree_label(&f.project.join("absent"), &f.wt, &f.ctx()),
            None
        );
        // A budget that is spent.
        let mut spent = f.ctx();
        spent.deadline = Some(Instant::now() - Duration::from_millis(1));
        assert_eq!(worktree_label(&f.project, &f.wt, &spent), None);
        // With time left the same ledger reads.
        let mut left = f.ctx();
        left.deadline = Some(Instant::now() + Duration::from_secs(60));
        assert!(worktree_label(&f.project, &f.wt, &left).is_some());
    }

    #[cfg(unix)]
    #[test]
    fn a_folder_spelled_through_a_link_matches_in_either_spelling() {
        let f = fixture();
        let link = f.wt.parent().unwrap().join("link-to-work");
        std::os::unix::fs::symlink(&f.wt, &link).unwrap();
        // The session is in the folder through the link, the reservation holds the real path.
        let card = f.reserve("active", "held", &f.wt, Some("spec/20261119-demo-work"));
        assert_eq!(f.label(&link.join("src")), label(card, "demo-work"));
        // The reservation holds the link spelling, the session is in the real folder.
        let g = fixture();
        let glink = g.wt.parent().unwrap().join("link-to-work");
        std::os::unix::fs::symlink(&g.wt, &glink).unwrap();
        let gcard = g.reserve("active", "held", &glink, Some("spec/20261119-demo-work"));
        assert_eq!(g.label(&g.wt.join("src")), label(gcard, "demo-work"));
    }

    #[test]
    fn reading_the_label_leaves_the_ledger_and_the_home_as_they_were() {
        let f = fixture();
        f.reserve("active", "held", &f.wt, Some("spec/20261119-demo-work"));
        let ledger_file = crate::ledger::ledger_path(&f.home, "demo").unwrap();
        let before = std::fs::read(&ledger_file).unwrap();
        let listing = |dir: &Path| {
            let mut all = Vec::new();
            let mut stack = vec![dir.to_path_buf()];
            while let Some(d) = stack.pop() {
                for e in std::fs::read_dir(&d).unwrap() {
                    let p = e.unwrap().path();
                    all.push(p.clone());
                    if p.is_dir() {
                        stack.push(p);
                    }
                }
            }
            all.sort();
            all
        };
        let tree = listing(&f.home);
        assert!(f.label(&f.wt).is_some());
        assert_eq!(std::fs::read(&ledger_file).unwrap(), before);
        assert_eq!(listing(&f.home), tree, "no file or folder made in the HOME");
    }
}
