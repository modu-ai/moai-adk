//! Everything the statusline reads outside stdin: environment, the project configuration
//! and state files, the HOME state files, and local git. The Go statusline also writes
//! `$HOME/.moai/state/last-model.txt` and `<anchor>/.moai/state/context-usage/*.json`; this
//! module writes neither. It has one write of its own, the per-session cache of the git facts
//! under `$TMPDIR/moai-statusline/` (SPEC-20261122-statusline-where REQ-5): one small file per
//! session id, a folder only the user may enter, an unwritable folder never failing a render.
//! The forge counts cache is written by the detached refresh child in [`super::counts`]
//! (#2588), never by the render path.
//!
//! HOME state (#2455, user decision 2026-11-11): `last-model.txt` is read from
//! `$HOME/.moai-cli/state/` only, never from the Go folder `$HOME/.moai/`, and `MOAI_HOME` is
//! not consulted. No Rust command writes it yet, so "nothing to read" is the normal state. The
//! session task (`last-session-state.json`, 📋) is not read at all since the three-line layout
//! (#2376, user decision 3: the task is never shown).
//!
//! The cards by stage are the one read under `$HOME/.moai/`: the project's local ledger in
//! `$HOME/.moai/projects/<project_id>/`, opened read-only through `crate::ledger` with a short
//! wait, in four statements (see [`read_flow`]).
//!
//! The only subprocess is `git`, at most one `git status` and, on a lane branch, one
//! `git rev-list --left-right --count` per render, and only when a `.git` entry is found above
//! the directory Go would open and the cache of the session does not answer; both run read-only
//! under the render deadline (SPEC-20261123-statusline-integration-distance). The branch name
//! and the short commit of a detached HEAD come from the `HEAD` file, never from a process. No
//! network here: the counts refresh child is started by `statusline::run` after the render
//! (#2588).

use super::config::{self, Node};
use super::icons;
use super::input::{self, Input};
use super::segments::{self, ContextReading, GitStatus};
use crate::ledger::partial::partial_landing_where;
use serde::{Deserialize, Serialize};
use std::collections::{BTreeMap, HashMap, HashSet};
use std::io::Read;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use std::time::{Duration, Instant};

/// Which midnight the count of cards closed today starts from (SPEC-20261124-statusline-counts-truth
/// REQ-9, REQ-10).
#[derive(Debug, Clone, Copy, Default, PartialEq, Eq)]
pub enum DayZone {
    /// The UTC day.
    #[default]
    Utc,
    /// The local day of the machine's time zone (`TZ`, else the system zone); the UTC day where
    /// the zone cannot be read.
    Machine,
    /// A fixed zone, in seconds east of UTC (tests).
    Offset(i64),
}

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
    /// The zone "done today" is counted in: the machine's outside tests, UTC in them so a
    /// test never depends on the zone of the machine it runs on.
    pub day_zone: DayZone,
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
            day_zone: DayZone::Machine,
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
            day_zone: DayZone::Utc,
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

    /// `HOME` as this render sees it (the Orca pass-through finds its wrapper below it).
    pub fn home_dir(&self) -> Option<PathBuf> {
        self.home()
    }

    /// Every variable of this render, for a child started with the render's environment.
    pub fn vars(&self) -> impl Iterator<Item = (&String, &String)> {
        self.vars.iter()
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
    /// The `CW` percentage as the host gives it, or its fallback (SPEC-20261122-statusline-limits
    /// D-1); `None` when the input names none.
    pub context: Option<f64>,
    /// `cost.total_cost_usd`, the host's estimate of the session's cost in US dollars.
    pub cost: Option<f64>,
    pub version: Option<Version>,
    pub claude_version: String,
    /// `Some` when `effort` is present, even with an empty level.
    pub effort: Option<String>,
    pub thinking: Option<bool>,
    pub five_hour: Option<input::RateWindow>,
    pub seven_day: Option<input::RateWindow>,
    pub git: Option<GitStatus>,
    pub repo: Option<(String, String)>,
    pub github: GitHubCounts,
    /// The cards of the project's local ledger by stage, for the flow line; `None` when there
    /// is no ledger to read (see [`read_flow`]).
    pub flow: Option<Flow>,
    /// #2756 AC-7: the open pull request of the current branch.
    pub pr: Option<PrInfo>,
    /// #2756 AC-9: the worktree name the host gave; empty in the main checkout. It is no item of
    /// the first line (REQ-10 of SPEC-20261123-statusline-order-bars) and no layout reads it; the
    /// card of the session is found from its working folder and its session id.
    pub worktree: String,
    /// SPEC-20261122-statusline-where: the integration branch of the project (`work_flow` of its
    /// profile, else the remote's default branch), `None` when neither names one.
    pub integration: Option<String>,
    /// SPEC-20261123-statusline-order-bars REQ-9: the first seven characters of the commit that the
    /// local ref `refs/heads/<integration>` names; empty when no integration branch is known or
    /// its ref cannot be read.
    pub integration_tip: String,
}

/// The `pr` fields the status line shows.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct PrInfo {
    pub number: i64,
    /// Empty when absent or not a plain `http(s)` URL; the number is then shown unlinked.
    pub url: String,
    /// Empty when absent.
    pub review_state: String,
    /// SPEC-20261122-statusline-where REQ-9: the host says the request is a GitLab merge
    /// request (`pr.kind` is `mr`), so the item is written `!<n>`.
    pub merge_request: bool,
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

/// The stage of an open card in its progress (SPEC-20261122-statusline-flow REQ-1): the nine
/// stages the ledger tells apart, each with the icon the flow line writes for a card of mine in
/// it (SPEC-20261122-statusline-limits-left REQ-6).
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Stage {
    /// Drafted, not yet ready.
    Draft,
    /// Ready to be taken.
    Ready,
    /// A lane holds it with a live lease and works on it.
    Working,
    /// It is worked on by nobody — the lease has ended or no reservation holds it.
    Stale,
    /// Pushed, waiting for an audit or the leader's integration (every `review` card).
    Submitted,
    /// An audit passed for the head and nobody landed it.
    Audited,
    /// Landed on the integration branch and not yet closed.
    Landed,
    /// The lane asked for a decision.
    Blocked,
    /// On hold.
    Hold,
}

impl Stage {
    /// The icon of the stage in the one icon set of the progress board, the status line and the
    /// lane reports ([`icons`], SPEC-20261124-card-flow-icons REQ-5): to do (a draft has the icon
    /// as the card in hand, but the flow line counts only the ready cards,
    /// SPEC-20261124-statusline-counts-truth), work in progress (a card whose lease has ended is
    /// still the card in hand and is counted with the working ones), in check, check passed and
    /// waiting to land or to be closed, blocked, and needs the user's pick.
    pub fn icon(self) -> &'static str {
        match self {
            Stage::Draft | Stage::Ready => icons::TODO,
            Stage::Working | Stage::Stale => icons::WORK,
            Stage::Submitted => icons::CHECK,
            Stage::Audited | Stage::Landed => icons::READY,
            Stage::Blocked => icons::BLOCKED,
            Stage::Hold => icons::QUESTION,
        }
    }
}

/// The card the session is on: its number and stage (SPEC REQ-6).
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct MyCard {
    pub card: i64,
    pub stage: Stage,
}

/// The cards of the project's local ledger by the stage of their progress, for the flow line
/// (SPEC-20261122-statusline-flow). The nine stage counts and the `partial` count add up to the
/// cards that are not `done` and not withdrawn (a hold card with a withdraw or fold event is in
/// no stage, SPEC-20261124-statusline-counts-truth REQ-8).
#[derive(Debug, Default, Clone, Copy, PartialEq, Eq)]
pub struct Flow {
    pub draft: i64,
    /// The ready cards that wait for a lane: one a lane handed back and released is `submitted`
    /// or `audited`.
    pub ready: i64,
    pub working: i64,
    pub stale: i64,
    pub submitted: i64,
    pub audited: i64,
    pub landed: i64,
    /// Landed cards with open criteria that would count as `ready` or `landed` (a partial
    /// landing); they are in neither of those two counts.
    pub partial: i64,
    /// The open criteria of the `partial` cards together.
    pub partial_open: i64,
    pub blocked: i64,
    /// The hold cards that were not withdrawn: the cards that wait for the user's pick.
    pub hold: i64,
    /// The hold cards with a withdraw or fold event: nobody owes them anything. They are in no
    /// stage and not in `open` (SPEC-20261124-card-flow-icons REQ-4).
    pub paused: i64,
    /// The cards closed at or after the start of the day of the render's clock in the zone of
    /// [`Context::day_zone`] (the machine's local day; SPEC-20261124-statusline-counts-truth).
    pub done_today: i64,
    /// The card this session holds, when it holds one.
    pub mine: Option<MyCard>,
    /// `false` when the notes, the landing events or the reservations could not be read in the
    /// budget: only `draft`, `ready`, `hold`, `done_today` and `working` (`active` + `review`,
    /// unsplit) are set then and the line writes the unsplit form, never a guess.
    pub split: bool,
}

/// Who asks: the stdin session id and the folder the session works in, for "my card"
/// (SPEC REQ-6).
#[derive(Debug, Clone, Copy, Default)]
pub struct Whose<'a> {
    pub session_id: &'a str,
    pub dir: Option<&'a Path>,
}

/// `statusline.yaml` as the Go CLI reads it: theme and the segments map.
#[derive(Debug, Default, Clone)]
pub struct DisplayConfig {
    pub theme: String,
    pub segments: BTreeMap<String, bool>,
    /// The `padding` of the host's `statusLine` block that this file mirrors, in characters
    /// (default 0): the host adds it to every line, so the width budget subtracts it.
    pub padding: usize,
}

/// Version this binary reports in the `🗿` segment.
pub const BINARY_VERSION: &str = crate::BUILD_VERSION;

pub fn collect(input: Option<&Input>, ctx: &Context) -> Status {
    let project_root = find_project_root(ctx);
    let anchor = input.and_then(resolve_anchor);
    let mut status = Status {
        model: collect_model(input, ctx),
        version: Some(collect_version(ctx, BINARY_VERSION)),
        ..Default::default()
    };
    // The branch and the commit come from the `HEAD` file and the numbers from the session
    // cache or the git processes of the render (SPEC-20261122-statusline-where REQ-2, REQ-6;
    // SPEC-20261123-statusline-integration-distance REQ-6); the profile of the project names the
    // integration branch.
    let git_dir = project_root.clone().or_else(|| ctx.cwd.clone());
    let profile_root = anchor.clone().or_else(|| project_root.clone());
    let session_id = input.map_or("", |i| input::text(&i.session_id));
    if let Some(read) = git_dir.and_then(|d| read_git(&d, ctx, profile_root.as_deref(), session_id))
    {
        status.git = Some(read.git);
        status.integration = read.integration;
        status.integration_tip = read.integration_tip;
    }
    let Some(input) = input else {
        return status;
    };
    // SPEC-20261122-statusline-limits REQ-11: only what a shown item needs is collected — not
    // the session duration, the output style, the cache hit ratio, the prompt cache timer or
    // the goal file.
    status.context = collect_context(input, ctx);
    status.cost = input.cost.as_ref().and_then(|c| c.total_cost_usd);
    // The version comes from stdin and is printed as text: a control character (an escape, a
    // newline) or an invisible format character (a direction mark, a zero-width one) could
    // move or hide what the line shows, so none reaches the screen.
    status.claude_version = input::text(&input.version)
        .chars()
        .filter(|c| !c.is_control() && !invisible_format_char(*c))
        .collect();
    status.effort = input
        .effort
        .as_ref()
        .map(|e| input::text(&e.level).to_string());
    status.thinking = input.thinking.as_ref().map(|t| t.enabled.unwrap_or(false));
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
    }
    let dir = session_dir(input);
    status.flow = anchor
        .as_deref()
        .or(project_root.as_deref())
        .and_then(|project| {
            let who = Whose {
                session_id: input::text(&input.session_id),
                dir: dir.as_deref(),
            };
            read_flow(project, ctx, &who)
        });
    status.pr = input.pr.as_ref().and_then(|p| {
        Some(PrInfo {
            number: p.number.filter(|n| *n > 0)?,
            url: safe_link(input::text(&p.url)).to_string(),
            review_state: input::text(&p.review_state).to_string(),
            // Official statusline page: `pr.kind` is `mr` for a GitLab merge request and absent
            // for a GitHub pull request; any other value is a pull request.
            merge_request: p.kind.as_deref() == Some("mr"),
        })
    });
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
    // A `padding` that is not a plain non-negative number reads as 0, never as a failed file.
    let padding = match doc.get(&["statusline", "padding"]) {
        Some(Node::Scalar(s)) => s.trim().parse().unwrap_or(0),
        _ => 0,
    };
    DisplayConfig {
        theme,
        segments,
        padding,
    }
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

/// Where the git common directory of `dir` is, found on disk: the main checkout's `.git`
/// folder (with that checkout's top), or the folder a linked worktree's `commondir` names.
enum GitCommon {
    Main { top: PathBuf },
    Linked(PathBuf),
}

fn find_common(dir: &Path) -> Option<GitCommon> {
    let dir = std::fs::canonicalize(dir).ok()?;
    let (top, dot_git) = find_dot_git(&dir)?;
    if dot_git.is_dir() {
        return Some(GitCommon::Main { top });
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
        GitCommon::Main { top } => Some(top),
        GitCommon::Linked(common) => common.parent().map(Path::to_path_buf),
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
    } else if lower.contains("fable") {
        // The glm launcher maps the fable alias too (SPEC-20261123-glm-54-followups D-5).
        "ANTHROPIC_DEFAULT_FABLE_MODEL"
    } else if !lower.starts_with("claude") {
        return cleaned.to_string();
    } else {
        return display.to_string();
    };
    // The glm launcher's value carries Claude Code's 1M marker (`glm-5.4[1m]`); the line shows
    // the model without it, as `shorten_model_name` does for an id.
    match ctx.var(key) {
        Some(glm) if !glm.starts_with("claude-") => {
            glm.strip_suffix("[1m]").unwrap_or(glm).trim().to_string()
        }
        _ => display.to_string(),
    }
}

// ---- context window ----

/// An invisible Unicode format character that changes how a line reads without showing: the
/// soft hyphen, the Arabic letter mark, the zero-width and direction marks, the direction
/// embeddings, overrides and isolates, the word joiner and invisible operators, and the byte
/// order mark.
fn invisible_format_char(c: char) -> bool {
    matches!(
        c,
        '\u{00AD}'
            | '\u{061C}'
            | '\u{180E}'
            | '\u{200B}'..='\u{200F}'
            | '\u{202A}'..='\u{202E}'
            | '\u{2060}'..='\u{2064}'
            | '\u{2066}'..='\u{206F}'
            | '\u{FEFF}'
    )
}

/// The tokens of `current_usage` the `CW` number falls back to: none when the object carries no
/// count at all (REQ-2: `CW n/a`, never a guessed `0%`), and a sum that stops at `i64::MAX`
/// instead of overflowing on absurd counts.
fn usage_tokens(u: &input::CurrentUsage) -> Option<i64> {
    let counts = [
        u.input_tokens,
        u.cache_creation_input_tokens,
        u.cache_read_input_tokens,
    ];
    counts.iter().any(Option::is_some).then(|| {
        counts
            .iter()
            .flatten()
            .fold(0i64, |sum, n| sum.saturating_add(*n))
    })
}

fn collect_context(input: &Input, ctx: &Context) -> Option<f64> {
    let reading = input.context_window.as_ref().map(|cw| ContextReading {
        used_percentage: cw.used_percentage,
        context_window_size: cw.context_window_size.unwrap_or(0),
        used: cw.used.unwrap_or(0),
        total: cw.total.unwrap_or(0),
        current_usage_tokens: cw.current_usage.as_ref().and_then(usage_tokens),
    });
    let size_override = if reading.is_some() {
        context_window_override(ctx)
    } else {
        0
    };
    segments::context_percent(reading, size_override)
}

/// Built-in GLM context windows (memory.go `glmContextWindows`); `glm-5.4` is the glm
/// launcher's main model since the user decision of 2026-11-23 (SPEC-20261123-glm-54-model).
const GLM_CONTEXT_WINDOWS: [(&str, i64); 3] = [
    ("glm-5.3-flash", 1_000_000),
    ("glm-5.3", 1_000_000),
    ("glm-5.4", 1_000_000),
];
const GLM_ENV_SLOTS: [&str; 4] = [
    "ANTHROPIC_DEFAULT_OPUS_MODEL",
    "ANTHROPIC_DEFAULT_SONNET_MODEL",
    "ANTHROPIC_DEFAULT_HAIKU_MODEL",
    "ANTHROPIC_DEFAULT_FABLE_MODEL",
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

// ---- where: the items of line 1 (SPEC-20261122-statusline-where) ----

/// What the `HEAD` file of the git directory says (REQ-2): a branch, a detached commit
/// (its first seven characters) or nothing usable.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Head<'a> {
    Branch(&'a str),
    Detached(&'a str),
    Unreadable,
}

/// The seconds a cache of the git facts is trusted (REQ-6, D-1 of the SPEC): a cache is fresh
/// while its age is below this.
pub const CACHE_MAX_AGE_SECS: i64 = 10;
/// A cache file older than this many seconds is removed by the render that writes a cache.
pub const CACHE_KEEP_SECS: i64 = 86_400;
/// The render that writes a cache removes at most this many old cache files.
pub const CACHE_REMOVE_PER_RENDER: usize = 10;
/// The folder under `$TMPDIR` that holds the cache files of the sessions.
const CACHE_FOLDER: &str = "moai-statusline";
/// A cache file larger than this is not one.
const CACHE_FILE_CAP: u64 = 64 * 1024;
/// A `HEAD` file, a ref file or a `.git` file larger than this is not read.
const SMALL_FILE_CAP: u64 = 4096;
/// `packed-refs` is searched only when the loose ref of the integration branch is missing.
const PACKED_REFS_CAP: u64 = 8 * 1024 * 1024;
/// How many characters of the integration commit the first line shows.
const INTEGRATION_TIP_CHARS: usize = 7;

/// The text of a `HEAD` file read as a branch, a detached commit or nothing. A branch name is
/// taken only when it is one git can be given safely (`profile::branch_name`: a portable subset,
/// no escape or control byte can reach the screen), and a commit is a 40 or 64 digit hash.
pub fn head_of(content: &str) -> Head<'_> {
    let text = content.strip_suffix('\n').unwrap_or(content);
    if let Some(name) = text.strip_prefix("ref: refs/heads/") {
        return if crate::profile::branch_name(name) {
            Head::Branch(name)
        } else {
            Head::Unreadable
        };
    }
    if matches!(text.len(), 40 | 64) && text.bytes().all(|b| b.is_ascii_hexdigit()) {
        Head::Detached(&text[..7])
    } else {
        Head::Unreadable
    }
}

/// The integration branch: the name `work_flow.integration_branch` of the project profile text
/// gives, else the remote's default branch from the text of `refs/remotes/origin/HEAD`, else
/// `None`. A name git cannot be given safely is no name.
pub fn integration_from(profile: Option<&str>, origin_head: Option<&str>) -> Option<String> {
    let named = profile
        .and_then(|p| serde_json::from_str::<serde_json::Value>(p).ok())
        .and_then(|v| {
            v.get("work_flow")?
                .get("integration_branch")?
                .as_str()
                .map(str::to_string)
        })
        .filter(|name| crate::profile::branch_name(name));
    named.or_else(|| {
        let text = origin_head?;
        let name = text
            .strip_suffix('\n')
            .unwrap_or(text)
            .strip_prefix("ref: refs/remotes/origin/")?;
        crate::profile::branch_name(name).then(|| name.to_string())
    })
}

/// How a render gets the commits ahead of and behind (SPEC-20261123-statusline-integration-distance).
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum Measure<'a> {
    /// git's own numbers from the `## branch...upstream` header of `git status --branch`: the
    /// integration branch itself, no integration branch known, a name git must not be given.
    Upstream,
    /// The branch against the local integration branch, by a second git call that reads no
    /// setting of the branch.
    Integration(&'a str),
    /// The integration branch has no local ref: nothing to measure, no second process, zeros.
    Missing,
}

/// What [`read_git`] measures: the lane branch against the integration branch, only when both
/// names are safe to hand to git and the branch is not the integration branch itself.
/// `have_commit` says that the local ref of the integration branch exists.
pub fn measure<'a>(
    branch: Option<&str>,
    integration: Option<&'a str>,
    have_commit: bool,
) -> Measure<'a> {
    match (branch, integration) {
        (Some(branch), Some(integration))
            if branch != integration
                && crate::profile::branch_name(branch)
                && crate::profile::branch_name(integration) =>
        {
            if have_commit {
                Measure::Integration(integration)
            } else {
                Measure::Missing
            }
        }
        _ => Measure::Upstream,
    }
}

/// The arguments of the `git status` call of a render, after the program name. There is no `-c`
/// override of a `branch.<name>.*` key: those keys can hold several values, and an override is
/// added to the ones the branch has. The `--branch` header is asked for only when its upstream
/// numbers are the ones shown.
pub fn status_args(measure: &Measure) -> Vec<String> {
    let mut args = ["--no-optional-locks", "status", "--porcelain"]
        .map(String::from)
        .to_vec();
    if *measure == Measure::Upstream {
        args.push("--branch".to_string());
    }
    args
}

/// The arguments of the second git call: the commits only the integration branch has (behind)
/// and only `HEAD` has (ahead), from the full ref of the integration branch. The `--` after the
/// range says that no path follows: a work tree file named like the range would otherwise stop
/// git with "ambiguous argument".
pub fn distance_args(integration: &str) -> Vec<String> {
    vec![
        "rev-list".to_string(),
        "--left-right".to_string(),
        "--count".to_string(),
        format!("refs/heads/{integration}...HEAD"),
        "--".to_string(),
    ]
}

/// `<behind>\t<ahead>` of a `rev-list --left-right --count`; `None` for anything else.
fn parse_distance(text: &str) -> Option<(i64, i64)> {
    let (behind, ahead) = text.strip_suffix('\n').unwrap_or(text).split_once('\t')?;
    let number = |s: &str| {
        (!s.is_empty() && s.bytes().all(|b| b.is_ascii_digit()))
            .then(|| s.parse::<i64>().ok())
            .flatten()
    };
    Some((number(behind)?, number(ahead)?))
}

/// What the first line needs to know about where the session stands.
#[derive(Debug, Clone, Copy)]
pub struct WhereInput<'a> {
    /// The host's worktree name; empty in the main checkout.
    pub worktree: &'a str,
    pub head: Head<'a>,
    pub integration: Option<&'a str>,
    /// The numbers of `git status --porcelain --branch`.
    pub ahead: i64,
    pub behind: i64,
    pub dirty: i64,
}

/// The `↑ ↓ +` marks of the place items; a zero is not written.
#[derive(Debug, Clone, Copy, Default, PartialEq, Eq)]
pub struct Marks {
    pub ahead: i64,
    pub behind: i64,
    pub dirty: i64,
}

/// The `🧭` item: the commits the branch is ahead of and behind the integration branch.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Distance {
    pub name: String,
    pub ahead: i64,
    pub behind: i64,
}

/// The place items of line 1. The marks follow the last of the worktree and branch items.
#[derive(Debug, Clone, Default, PartialEq, Eq)]
pub struct Where {
    /// `🌳 <worktree>`.
    pub worktree: Option<String>,
    /// `🅱️ <branch>` or the short commit of a detached HEAD.
    pub branch: Option<String>,
    pub marks: Marks,
    pub distance: Option<Distance>,
}

/// Which place items show and with which numbers (REQ-3, REQ-4). The worktree name comes from
/// the host, so its item stays when `HEAD` cannot be read, without marks; the branch item and the
/// distance need the `HEAD` file.
pub fn where_items(input: &WhereInput) -> Where {
    let worktree = (!input.worktree.is_empty()).then(|| input.worktree.to_string());
    let (name, detached) = match input.head {
        Head::Branch(name) => (name, false),
        Head::Detached(commit) => (commit, true),
        Head::Unreadable => {
            return Where {
                worktree,
                ..Where::default()
            }
        }
    };
    // A detached HEAD has no upstream: no ahead or behind count, only the dirty count.
    let (ahead, behind) = if detached {
        (0, 0)
    } else {
        (input.ahead, input.behind)
    };
    let integration = input.integration.filter(|n| !n.is_empty());
    let distance = integration
        .filter(|n| !detached && *n != name)
        .map(|n| Distance {
            name: n.to_string(),
            ahead,
            behind,
        });
    Where {
        branch: (worktree.as_deref() != Some(name) || detached).then(|| name.to_string()),
        worktree,
        // The behind count lives in the distance; it stays on the branch item only when git's
        // own upstream is all there is to measure against.
        marks: Marks {
            ahead,
            behind: if integration.is_none() { behind } else { 0 },
            dirty: input.dirty,
        },
        distance,
    }
}

/// The file name of a session's cache: the stdin session id and `.json`, only for an id of
/// 1 to 64 characters of `A-Z a-z 0-9 _ -` (REQ-5).
pub fn cache_file_name(session_id: &str) -> Option<String> {
    let plain = !session_id.is_empty()
        && session_id.len() <= 64
        && session_id
            .bytes()
            .all(|b| b.is_ascii_alphanumeric() || matches!(b, b'_' | b'-'));
    plain.then(|| format!("{session_id}.json"))
}

/// Whether a cache may stand in for git (REQ-6): the same `HEAD` content, the same integration
/// commit and an age below [`CACHE_MAX_AGE_SECS`]. A cache written in the future is not fresh.
pub fn cache_is_fresh(
    written_at: i64,
    now: i64,
    cached_head: &str,
    head: &str,
    cached_integration: &str,
    integration: &str,
) -> bool {
    (0..CACHE_MAX_AGE_SECS).contains(&now.saturating_sub(written_at))
        && cached_head == head
        && cached_integration == integration
}

// ---- git ----

/// The numbers of one render's git calls (`git status` and, on a lane branch, `git rev-list`),
/// the facts the cache keeps.
#[derive(Debug, Clone, Copy, Default, PartialEq, Eq, Serialize, Deserialize)]
struct Counts {
    ahead: i64,
    behind: i64,
    staged: i64,
    modified: i64,
    untracked: i64,
}

/// What the cache file holds (REQ-5): the key (the canonical git directory that holds the `HEAD`
/// file, the work state of [`work_state`], the `HEAD` text and the integration commit), the time
/// it was written, and the parsed git facts. A record without a folder or a work state (an older
/// file) never matches.
#[derive(Debug, Serialize, Deserialize)]
struct CacheRecord {
    #[serde(default)]
    folder: String,
    #[serde(default)]
    state: String,
    head: String,
    integration: String,
    written_at: i64,
    counts: Counts,
}

/// What a render reads about git: the facts of the status line, the integration branch and the
/// first seven characters of the commit its local ref names (empty when the ref cannot be read).
struct GitRead {
    git: GitStatus,
    integration: Option<String>,
    integration_tip: String,
}

/// Where the git directory of a checkout and the common directory of its repository are.
struct GitDirs {
    /// The folder holding `HEAD`: `.git`, or the gitdir a linked worktree's `.git` file names.
    git_dir: PathBuf,
    /// The folder holding `refs/` and `packed-refs`.
    common: PathBuf,
}

/// The git directories above the canonical `dir`, found on disk without a process.
fn git_dirs(dir: &Path) -> Option<GitDirs> {
    let (top, dot_git) = find_dot_git(dir)?;
    if std::fs::metadata(&dot_git).ok()?.is_dir() {
        return Some(GitDirs {
            git_dir: dot_git.clone(),
            common: dot_git,
        });
    }
    let text = read_small(&dot_git, SMALL_FILE_CAP)?;
    let named = PathBuf::from(text.lines().next()?.strip_prefix("gitdir:")?.trim());
    let git_dir = if named.is_absolute() {
        named
    } else {
        top.join(named)
    };
    let common = read_small(&git_dir.join("commondir"), SMALL_FILE_CAP)
        .map(|c| PathBuf::from(c.trim()))
        .map_or_else(
            || git_dir.clone(),
            |c| if c.is_absolute() { c } else { git_dir.join(c) },
        );
    Some(GitDirs { git_dir, common })
}

/// A small regular file as text: a link, a pipe, a folder, a file over `cap` bytes or text that
/// is not UTF-8 is not read (a named pipe would block the render).
fn read_small(path: &Path, cap: u64) -> Option<String> {
    let meta = std::fs::symlink_metadata(path).ok()?;
    if !meta.is_file() || meta.len() > cap {
        return None;
    }
    std::fs::read_to_string(path).ok()
}

/// The commit a hash text names: 40 or 64 hexadecimal digits.
fn is_commit(text: &str) -> bool {
    matches!(text.len(), 40 | 64) && text.bytes().all(|b| b.is_ascii_hexdigit())
}

/// The commit of the local branch `name` (a name that passed `profile::branch_name`): its loose
/// ref file, else its line of `packed-refs`; empty when neither says. Read for the cache key and
/// for the integration item of the first line.
fn ref_commit(common: &Path, name: &str) -> String {
    if let Some(text) = read_small(&common.join("refs/heads").join(name), SMALL_FILE_CAP) {
        if is_commit(text.trim()) {
            return text.trim().to_string();
        }
    }
    let want = format!("refs/heads/{name}");
    read_small(&common.join("packed-refs"), PACKED_REFS_CAP)
        .and_then(|packed| {
            packed.lines().find_map(|line| {
                let (hash, name) = line.split_once(' ')?;
                (name == want && is_commit(hash)).then(|| hash.to_string())
            })
        })
        .unwrap_or_default()
}

/// The git facts of the checkout above `dir`: the branch or the short commit from the `HEAD`
/// file, the numbers from the session cache or else from the git calls of the render.
/// `None` when no repository is found (no process is started then).
fn read_git(
    dir: &Path,
    ctx: &Context,
    profile_root: Option<&Path>,
    session_id: &str,
) -> Option<GitRead> {
    let dirs = git_dirs(&std::fs::canonicalize(dir).ok()?)?;
    let profile = profile_root
        .and_then(|root| crate::profile::chosen_profile(root).ok().flatten())
        .and_then(|bytes| String::from_utf8(bytes).ok());
    let origin_head = read_small(
        &dirs.common.join("refs/remotes/origin/HEAD"),
        SMALL_FILE_CAP,
    );
    let integration = integration_from(profile.as_deref(), origin_head.as_deref());
    // The integration commit is read before `HEAD` is, from the ref of the integration branch
    // alone (a loose ref file, else `packed-refs`): a detached or unreadable `HEAD` shows the same
    // seven characters as a branch does, and `HEAD` never stands in for it (REQ-9 of
    // SPEC-20261123-statusline-order-bars).
    let commit = integration
        .as_deref()
        .map_or_else(String::new, |name| ref_commit(&dirs.common, name));
    let integration_tip: String = commit.chars().take(INTEGRATION_TIP_CHARS).collect();
    let mut git = GitStatus::default();
    let head_text = read_small(&dirs.git_dir.join("HEAD"), SMALL_FILE_CAP).unwrap_or_default();
    let branch = match head_of(&head_text) {
        Head::Branch(name) => {
            git.branch = name.to_string();
            Some(name)
        }
        Head::Detached(commit) => {
            git.head_sha = commit.to_string();
            None
        }
        // Nothing to measure: no branch, no distance and no process.
        Head::Unreadable => {
            return Some(GitRead {
                git,
                integration,
                integration_tip,
            });
        }
    };
    // The per-checkout git directory, as the disk spells it, is the folder of the key.
    let folder = std::fs::canonicalize(&dirs.git_dir)
        .unwrap_or_else(|_| dirs.git_dir.clone())
        .to_string_lossy()
        .into_owned();
    let state = work_state(&dirs, branch);
    let counts = git_counts(
        dir,
        ctx,
        &CacheKey {
            folder: &folder,
            state: &state,
        },
        &head_text,
        &commit,
        &measure(branch, integration.as_deref(), !commit.is_empty()),
        SessionCache::open(ctx, session_id),
    );
    git.ahead = counts.ahead;
    git.behind = counts.behind;
    git.staged = counts.staged;
    git.modified = counts.modified;
    git.untracked = counts.untracked;
    Some(GitRead {
        git,
        integration,
        integration_tip,
    })
}

/// The parts of the cache key besides the `HEAD` text and the integration commit.
struct CacheKey<'a> {
    /// The canonical per-checkout git directory.
    folder: &'a str,
    /// [`work_state`]: the branch commit and the index size and time.
    state: &'a str,
}

/// The numbers for one render: the cache when it is fresh for this git directory, work state,
/// `HEAD` text and integration commit, else the output of the `git status` call and, for a lane
/// branch, the `rev-list` call of [`distance_args`], which then rewrite the cache. A git that
/// cannot answer (a spent budget, a failure) gives zeros for what it could not say and writes
/// nothing.
fn git_counts(
    dir: &Path,
    ctx: &Context,
    key: &CacheKey<'_>,
    head_text: &str,
    commit: &str,
    measure: &Measure<'_>,
    cache: Option<SessionCache>,
) -> Counts {
    if let Some(record) = cache.as_ref().and_then(SessionCache::read) {
        // Another folder with the same `HEAD` text is another key: a session that moves to
        // another checkout asks git once. A commit or a stage on the same branch changes the
        // work state, so it asks git once too.
        if record.folder == key.folder
            && record.state == key.state
            && cache_is_fresh(
                record.written_at,
                ctx.now,
                &record.head,
                head_text,
                &record.integration,
                commit,
            )
        {
            return record.counts;
        }
    }
    let status = status_args(measure);
    let status: Vec<&str> = status.iter().map(String::as_str).collect();
    let Some(out) = run_git(dir, &status, ctx) else {
        return Counts::default();
    };
    let parsed = segments::parse_porcelain(&out);
    let mut counts = Counts {
        ahead: parsed.ahead,
        behind: parsed.behind,
        staged: parsed.staged,
        modified: parsed.modified,
        untracked: parsed.untracked,
    };
    // The distance of a lane branch is its own call: it reads no setting of the branch, so an
    // upstream the branch has cannot change it.
    let mut answered = true;
    match measure {
        Measure::Upstream => {}
        Measure::Missing => (counts.ahead, counts.behind) = (0, 0),
        Measure::Integration(integration) => {
            let distance = distance_args(integration);
            let distance: Vec<&str> = distance.iter().map(String::as_str).collect();
            match run_git(dir, &distance, ctx)
                .as_deref()
                .and_then(parse_distance)
            {
                Some((behind, ahead)) => (counts.ahead, counts.behind) = (ahead, behind),
                None => {
                    (counts.ahead, counts.behind) = (0, 0);
                    answered = false;
                }
            }
        }
    }
    if !answered {
        return counts;
    }
    if let Some(cache) = cache {
        let record = CacheRecord {
            folder: key.folder.to_string(),
            state: key.state.to_string(),
            head: head_text.to_string(),
            integration: commit.to_string(),
            written_at: ctx.now,
            counts,
        };
        if cache.write(&record) {
            cache.remove_old(ctx.now);
        }
    }
    counts
}

/// The cache file of one session under `$TMPDIR/moai-statusline/` (`/tmp` when `TMPDIR` is
/// unset). The folder and the file name are one place: [`SessionCache::open`] makes the folder
/// private (0700, the user's own, never a link) and names the file only by the session id.
struct SessionCache {
    dir: PathBuf,
    file: PathBuf,
}

impl SessionCache {
    /// `None` without a plain session id, with a relative `TMPDIR` (it would name a folder inside
    /// the checkout the render runs in), without a folder this user alone may enter, or on a
    /// platform without Unix modes: every render then asks git.
    fn open(ctx: &Context, session_id: &str) -> Option<SessionCache> {
        let name = cache_file_name(session_id)?;
        let dir = cache_folder(ctx.var("TMPDIR"))?;
        private_folder(&dir).then(|| SessionCache {
            file: dir.join(name),
            dir,
        })
    }

    /// The record in the file; a missing, unreadable, linked, oversized or malformed file is none.
    fn read(&self) -> Option<CacheRecord> {
        serde_json::from_str(&read_small(&self.file, CACHE_FILE_CAP)?).ok()
    }

    /// Writes the record through a temporary file and a rename, so a reader never sees half of
    /// it and a link in the file's place is replaced, not written through. `false` when it could
    /// not be written; nothing else changes then.
    fn write(&self, record: &CacheRecord) -> bool {
        use std::io::Write;
        let Ok(bytes) = serde_json::to_vec(record) else {
            return false;
        };
        let temp = self.dir.join(format!(".write-{}.tmp", std::process::id()));
        let mut options = std::fs::OpenOptions::new();
        options.write(true).create_new(true);
        #[cfg(unix)]
        {
            use std::os::unix::fs::OpenOptionsExt;
            options.mode(0o600);
        }
        let written = options
            .open(&temp)
            .and_then(|mut f| f.write_all(&bytes))
            .and_then(|()| std::fs::rename(&temp, &self.file));
        if written.is_err() {
            let _ = std::fs::remove_file(&temp);
        }
        written.is_ok()
    }

    /// Removes cache files (and temporary files a stopped render left) older than
    /// [`CACHE_KEEP_SECS`] by their modified time against the render's clock, at most
    /// [`CACHE_REMOVE_PER_RENDER`] of them. Other files and links stay.
    fn remove_old(&self, now: i64) {
        let Ok(entries) = std::fs::read_dir(&self.dir) else {
            return;
        };
        let mut removed = 0;
        for entry in entries.flatten() {
            if removed >= CACHE_REMOVE_PER_RENDER {
                break;
            }
            let path = entry.path();
            let cache_like = path.extension().is_some_and(|e| e == "json" || e == "tmp");
            let old = entry.metadata().ok().is_some_and(|m| {
                m.is_file()
                    && m.modified()
                        .ok()
                        .and_then(|t| t.duration_since(std::time::UNIX_EPOCH).ok())
                        .is_some_and(|t| now.saturating_sub(t.as_secs() as i64) > CACHE_KEEP_SECS)
            });
            if cache_like && path != self.file && old && std::fs::remove_file(&path).is_ok() {
                removed += 1;
            }
        }
    }
}

/// The folder the cache folder goes in: `TMPDIR` when it is an absolute path, `/tmp` when it is
/// unset or empty, and none when it is relative.
fn cache_base(tmpdir: Option<&str>) -> Option<PathBuf> {
    match tmpdir {
        None | Some("") => Some(PathBuf::from("/tmp")),
        Some(dir) => Some(PathBuf::from(dir)).filter(|p| p.is_absolute()),
    }
}

/// The status line's cache folder for a `TMPDIR` value: `moai-statusline` in [`cache_base`],
/// none for a relative `TMPDIR`. Makes nothing. The session caches and the usage record
/// ([`crate::usage_limit`]) both live here. With `TMPDIR` unset or empty the folder is in the
/// shared `/tmp`, so its name carries this user's id (`moai-statusline-<uid>`): another local
/// user who made `/tmp/moai-statusline` first would otherwise own the name, and every read would
/// be refused as not private.
pub(crate) fn cache_folder(tmpdir: Option<&str>) -> Option<PathBuf> {
    let base = cache_base(tmpdir)?;
    Some(match tmpdir {
        None | Some("") => base.join(shared_tmp_folder_name()),
        Some(_) => base.join(CACHE_FOLDER),
    })
}

/// [`CACHE_FOLDER`] with this user's id, for the shared `/tmp`.
#[cfg(unix)]
fn shared_tmp_folder_name() -> String {
    // SAFETY: geteuid has no preconditions and cannot fail.
    let me = unsafe { libc::geteuid() };
    format!("{CACHE_FOLDER}-{me}")
}

#[cfg(not(unix))]
fn shared_tmp_folder_name() -> String {
    CACHE_FOLDER.to_string()
}

/// What a commit or a stage on the same branch changes while the `HEAD` text stays: the commit
/// of the checked-out branch (none for a detached `HEAD`, whose text is the commit) and the size
/// and modified time of the checkout's index file. Read from disk, no process.
fn work_state(dirs: &GitDirs, branch: Option<&str>) -> String {
    let commit = branch.map_or_else(String::new, |name| ref_commit(&dirs.common, name));
    let index = std::fs::symlink_metadata(dirs.git_dir.join("index"))
        .ok()
        .filter(std::fs::Metadata::is_file)
        .map_or_else(
            || "none".to_string(),
            |m| {
                let at = m
                    .modified()
                    .ok()
                    .and_then(|t| t.duration_since(std::time::UNIX_EPOCH).ok())
                    .map_or(0, |d| d.as_nanos());
                format!("{}:{at}", m.len())
            },
        );
    format!("commit={commit};index={index}")
}

/// The cache folder exists as a plain folder of this user that nobody else may enter: made 0700
/// when missing, tightened to 0700 when it is the user's own with wider rights, refused when it
/// is a link, another kind of file or another user's.
#[cfg(unix)]
pub(crate) fn private_folder(dir: &Path) -> bool {
    use std::os::unix::fs::{DirBuilderExt, MetadataExt, PermissionsExt};
    if std::fs::symlink_metadata(dir).is_err() {
        // A render that made it first is as good: the checks below decide.
        let _ = std::fs::DirBuilder::new().mode(0o700).create(dir);
    }
    let Ok(meta) = std::fs::symlink_metadata(dir) else {
        return false;
    };
    // SAFETY: geteuid has no preconditions and cannot fail.
    let me = unsafe { libc::geteuid() };
    meta.file_type().is_dir()
        && meta.uid() == me
        && (meta.mode() & 0o077 == 0
            || std::fs::set_permissions(dir, std::fs::Permissions::from_mode(0o700)).is_ok())
}

#[cfg(not(unix))]
pub(crate) fn private_folder(_dir: &Path) -> bool {
    false
}

/// What [`private_folder`] demands, checked without making or changing anything, for a reader
/// that writes nothing: a plain folder (not a link) of this user that nobody else may enter.
#[cfg(unix)]
pub(crate) fn is_private_folder(dir: &Path) -> bool {
    use std::os::unix::fs::MetadataExt;
    let Ok(meta) = std::fs::symlink_metadata(dir) else {
        return false;
    };
    // SAFETY: geteuid has no preconditions and cannot fail.
    let me = unsafe { libc::geteuid() };
    meta.file_type().is_dir() && meta.uid() == me && meta.mode() & 0o077 == 0
}

#[cfg(not(unix))]
pub(crate) fn is_private_folder(_dir: &Path) -> bool {
    false
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

/// The four statements of the flow read, in the order they run (SPEC-20261122-statusline-flow
/// REQ-11): the open cards with the done-today count, the notes of the cards a lane works on,
/// the landing events, and the reservations. Four short reads cost about 33 ms on the real
/// ledger against 115 ms for one statement joining them all; a test counts them here.
///
/// SPEC-20261124-statusline-counts-truth: a hold card with a withdraw or fold event is read as
/// `withdrawn` (REQ-8); the notes are those of the `ready`, `active` and `review` cards written at
/// or after the card's last reservation, with their kind and author (REQ-1 to REQ-5; a ready card
/// that was never taken has none, and an active or review card with no reservation row at all, an
/// older ledger's, keeps every note as it did); a landing counts when it is at or after that
/// reservation (REQ-6).
///
/// SPEC-20261126-statusline-stale-submit: the notes statement also selects each note's
/// `created_at` as its fifth column, so the age of the marker note can be told from the render's
/// clock; it stays one SELECT and the count stays four.
const FLOW_STATEMENTS: [&str; 4] = [
    "SELECT card_id, \
            CASE WHEN state = 'hold' AND card_id IN \
                      (SELECT card_id FROM event WHERE kind IN ('card.withdraw', 'card.fold')) \
                 THEN 'withdrawn' ELSE state END, 0 FROM card WHERE state <> 'done' \
     UNION ALL SELECT -1, 'done_today', count(*) FROM card \
     WHERE state = 'done' AND closed_at >= ?1",
    // `NOTE_HEAD_CHARS` is 1000: the literal keeps the statement a constant.
    "SELECT n.card_id, n.kind, n.author, substr(n.body, 1, 1000), n.created_at FROM card_note n \
     JOIN card c ON c.card_id = n.card_id WHERE c.state IN ('ready', 'active', 'review') \
     AND n.created_at >= coalesce((SELECT max(r.acquired_at) FROM reservation r \
                                   WHERE r.card_id = n.card_id), \
                                  CASE c.state WHEN 'ready' THEN 9223372036854775807 ELSE 0 END) \
     ORDER BY n.card_id, n.note_id DESC",
    // The cards landed since their last reservation (open count 0), then the partial landings
    // (SQL condition written once in `ledger::partial`) with their number of open criteria.
    concat!(
        "SELECT DISTINCT e.card_id, 0 FROM event e \
         WHERE e.kind = 'card.landed' AND e.card_id IS NOT NULL \
         AND e.at >= coalesce((SELECT max(r.acquired_at) FROM reservation r \
                               WHERE r.card_id = e.card_id), 0) \
         UNION ALL SELECT card.card_id, \
           (SELECT count(*) FROM card_ac a WHERE a.card_id = card.card_id AND a.status = 'open') \
         FROM card WHERE ",
        partial_landing_where!()
    ),
    "SELECT card_id, holder, status, lease_until, worktree, acquired_at, reservation_id \
     FROM reservation WHERE status IN ('held', 'held_for_review', 'expired_candidate')",
];

const _: () = assert!(NOTE_HEAD_CHARS == 1000);

/// How old, in seconds, the submit marker of a `ready` card may be before the card counts as
/// ready again (SPEC-20261126-statusline-stale-submit D-1): one day. A marker exactly this old is
/// still a submission; a card nobody audited for longer waits for a lane, not for a check.
const STALE_SUBMIT_SECS: i64 = 86_400;

/// The cards of `project`'s local ledger (`$HOME/.moai/projects/<project_id>/`, the one
/// `moai card list` reads) by the stage of their progress, for the flow line. Read-only like
/// that command (`ledger::open`: no folder, no file, no journal; no process and no network),
/// with [`CARDS_BUSY_TIMEOUT`] — or what is left of the render budget when that is less — in
/// place of the ledger's own wait. `None`, and so no flow line, whenever the ledger cannot be
/// read: no HOME, no project profile, no registered ledger, a file that is not this project's
/// ledger (corrupt, another project's, a link), a lock held past the wait, or a spent budget.
/// A budget spent after the open cards were read gives the unsplit [`Flow`] (`split` false).
pub fn read_flow(project: &Path, ctx: &Context, who: &Whose) -> Option<Flow> {
    let (ledger, _) = open_project_ledger(project, ctx)?;
    let spent = || {
        ctx.deadline
            .is_some_and(|deadline| Instant::now() >= deadline)
    };
    read_flow_in(&ledger, ctx.now, ctx.day_zone, who, &spent)
}

/// The first second of the UTC day that holds `now`.
fn utc_day_start(now: i64) -> i64 {
    now.div_euclid(86_400) * 86_400
}

/// The start of the day "done today" counts from (SPEC-20261124-statusline-counts-truth REQ-9
/// and REQ-10): `local` gives the first second of the local day that holds `now`, and a zone
/// that cannot be read (`None`) is the UTC day.
fn day_start_with(now: i64, local: impl FnOnce(i64) -> Option<i64>) -> i64 {
    local(now).unwrap_or_else(|| utc_day_start(now))
}

/// The first second of the day that holds `now` in `zone`.
fn day_start(now: i64, zone: DayZone) -> i64 {
    match zone {
        DayZone::Utc => utc_day_start(now),
        DayZone::Offset(east) => (now + east).div_euclid(86_400) * 86_400 - east,
        DayZone::Machine => day_start_with(now, machine_day_start),
    }
}

/// The local midnight of the machine's time zone before or at `now` (`TZ`, else the system zone):
/// the time of day of the local broken-down time set to 00:00:00 and read back by `mktime`, so a
/// day of 23 or 25 hours still starts at its own midnight. `None` when the zone cannot be read.
#[cfg(unix)]
fn machine_day_start(now: i64) -> Option<i64> {
    extern "C" {
        /// POSIX; the `libc` crate does not declare it for every unix target.
        fn tzset();
    }
    let at = libc::time_t::try_from(now).ok()?;
    // SAFETY: `tm` is plain data that `localtime_r` fills before it is read; `tzset`, `localtime_r`
    // and `mktime` touch only their arguments and the C library's zone tables. The status line
    // reads the zone on the render's own thread and starts no thread that reads it.
    unsafe {
        let mut tm: libc::tm = std::mem::zeroed();
        tzset();
        if libc::localtime_r(&at, &mut tm).is_null() {
            return None;
        }
        tm.tm_hour = 0;
        tm.tm_min = 0;
        tm.tm_sec = 0;
        tm.tm_isdst = -1;
        let start = libc::mktime(&mut tm);
        // `time_t` is `i64` on 64-bit unix and `i32` on 32-bit unix: the widening is needed there.
        #[allow(clippy::useless_conversion)]
        let start = i64::from(start);
        (start != -1).then_some(start)
    }
}

/// Where there is no `libc` the zone is not read (REQ-10): the day is the UTC day.
#[cfg(not(unix))]
fn machine_day_start(_now: i64) -> Option<i64> {
    None
}

/// [`read_flow_in`] with the UTC day, for the tests that spend the render budget.
#[cfg(test)]
fn read_flow_with(
    ledger: &crate::ledger::open::ReadLedger,
    now: i64,
    who: &Whose,
    spent: &dyn Fn() -> bool,
) -> Option<Flow> {
    read_flow_in(ledger, now, DayZone::Utc, who, spent)
}

/// [`read_flow`] on an open ledger with the render clock `now`, the zone of "done today" and the
/// "is the render budget spent" question as parameters, so a test can spend the budget in the
/// middle of a read.
fn read_flow_in(
    ledger: &crate::ledger::open::ReadLedger,
    now: i64,
    zone: DayZone,
    who: &Whose,
    spent: &dyn Fn() -> bool,
) -> Option<Flow> {
    if spent() {
        return None;
    }
    let conn = ledger.conn();
    // The local day of the render's clock (SPEC-20261124-statusline-counts-truth D-7): done
    // today is counted from its first second.
    let day_start = day_start(now, zone);
    let mut stmt = conn.prepare(FLOW_STATEMENTS[0]).ok()?;
    let rows = stmt
        .query_map([day_start], |r| {
            Ok((
                r.get::<_, i64>(0)?,
                r.get::<_, String>(1)?,
                r.get::<_, i64>(2)?,
            ))
        })
        .ok()?;
    let mut open: Vec<(i64, String)> = Vec::new();
    let mut flow = Flow::default();
    for row in rows {
        let (card, state, n) = row.ok()?;
        match state.as_str() {
            "done_today" => flow.done_today = n,
            "draft" => flow.draft += 1,
            "ready" => flow.ready += 1,
            "hold" => flow.hold += 1,
            // A withdrawn or folded card is on hold in the ledger and asks nothing of the user
            // (counts-truth REQ-8): it is in no stage the line draws, and it is counted as
            // paused (card-flow-icons REQ-4).
            "withdrawn" => {
                flow.paused += 1;
                continue;
            }
            // `active` and `review` stay whole until the notes are read.
            "active" | "review" => flow.working += 1,
            _ => continue,
        }
        if state != "done_today" {
            open.push((card, state));
        }
    }
    drop(stmt);
    // The split reads three more statements; when the budget is spent or a read fails the
    // unsplit counts above stand and the line writes the unsplit form.
    Some(split_flow(conn, now, who, &open, spent, flow).unwrap_or(flow))
}

/// What the three statements after the first add: the stage of every open card and "my card".
/// `None` when the budget is spent (checked before each statement and on every row) or a read
/// fails; the caller then keeps the unsplit counts.
fn split_flow(
    conn: &rusqlite::Connection,
    now: i64,
    who: &Whose,
    open: &[(i64, String)],
    spent: &dyn Fn() -> bool,
    unsplit: Flow,
) -> Option<Flow> {
    // Statement 2: the newest marker note of each card since its last reservation, newest first;
    // the first marker of a card decides it, and a block the leader's decision note answered
    // (D-3) is no marker.
    if spent() {
        return None;
    }
    let mut stmt = conn.prepare(FLOW_STATEMENTS[1]).ok()?;
    let rows = stmt
        .query_map([], |r| {
            Ok((
                r.get::<_, i64>(0)?,
                r.get::<_, String>(1)?,
                r.get::<_, String>(2)?,
                r.get::<_, String>(3)?,
                r.get::<_, i64>(4)?,
            ))
        })
        .ok()?;
    // The marker of each card with the time of the note that decided it.
    let mut markers: HashMap<i64, (Marker, i64)> = HashMap::new();
    let mut decided: Option<i64> = None;
    // The card the walk is on and whether a note of the leader's answer was seen above (newer
    // than) the notes still to read.
    let mut walking: Option<i64> = None;
    let mut answered = false;
    for row in rows {
        if spent() {
            return None;
        }
        let (card, kind, author, body, noted_at) = row.ok()?;
        if walking != Some(card) {
            walking = Some(card);
            answered = false;
        }
        if decided == Some(card) {
            continue;
        }
        if answers_a_block(&kind, &author) {
            answered = true;
            continue;
        }
        match marker_of(&body) {
            Some(Marker::Blocked) if answered => {}
            Some(marker) => {
                markers.insert(card, (marker, noted_at));
                decided = Some(card);
            }
            None => {}
        }
    }
    drop(stmt);
    // Statement 3: the cards with a landing event.
    if spent() {
        return None;
    }
    let mut stmt = conn.prepare(FLOW_STATEMENTS[2]).ok()?;
    let rows = stmt
        .query_map([], |r| Ok((r.get::<_, i64>(0)?, r.get::<_, i64>(1)?)))
        .ok()?;
    let mut landed: HashSet<i64> = HashSet::new();
    // The partial landings with their number of open criteria.
    let mut partial: HashMap<i64, i64> = HashMap::new();
    for row in rows {
        if spent() {
            return None;
        }
        let (card, open_criteria) = row.ok()?;
        if open_criteria > 0 {
            partial.insert(card, open_criteria);
        } else {
            landed.insert(card);
        }
    }
    drop(stmt);
    // Statement 4: the reservations that can hold a card.
    if spent() {
        return None;
    }
    let mut stmt = conn.prepare(FLOW_STATEMENTS[3]).ok()?;
    let rows = stmt
        .query_map([], |r| {
            Ok(Reservation {
                card: r.get(0)?,
                holder: r.get(1)?,
                status: r.get(2)?,
                lease_until: r.get(3)?,
                worktree: r.get(4)?,
                acquired_at: r.get(5)?,
                id: r.get(6)?,
            })
        })
        .ok()?;
    let mut reservations: Vec<Reservation> = Vec::new();
    for row in rows {
        if spent() {
            return None;
        }
        reservations.push(row.ok()?);
    }
    drop(stmt);

    let live: HashSet<i64> = reservations
        .iter()
        .filter(|r| r.is_live(now))
        .map(|r| r.card)
        .collect();
    // `ready` is counted again below: a ready card a lane handed back is no longer waiting for
    // one (REQ-1, REQ-2).
    let mut flow = Flow {
        ready: 0,
        working: 0,
        split: true,
        ..unsplit
    };
    let mut stages: HashMap<i64, Stage> = HashMap::new();
    for (card, state) in open {
        let (marker, noted_at) = markers
            .get(card)
            .map_or((None, 0), |(marker, at)| (Some(*marker), *at));
        let Some(stage) = stage_of(
            state,
            stage_marker(state, marker, noted_at, now),
            landed.contains(card),
            live.contains(card),
        ) else {
            continue;
        };
        // A landed card that keeps open criteria is counted apart where it would be counted as
        // ready or as landed; a card a lane works on again, hands in or is blocked on keeps its
        // stage, which says what is happening now.
        if matches!(stage, Stage::Ready | Stage::Landed) {
            if let Some(open_criteria) = partial.get(card) {
                flow.partial += 1;
                flow.partial_open += open_criteria;
                stages.insert(*card, stage);
                continue;
            }
        }
        match stage {
            // Counted by their ledger state already.
            Stage::Draft | Stage::Hold => {}
            Stage::Ready => flow.ready += 1,
            Stage::Working => flow.working += 1,
            Stage::Stale => flow.stale += 1,
            Stage::Submitted => flow.submitted += 1,
            Stage::Audited => flow.audited += 1,
            Stage::Landed => flow.landed += 1,
            Stage::Blocked => flow.blocked += 1,
        }
        stages.insert(*card, stage);
    }
    flow.mine = my_card(&reservations, who, now).and_then(|card| {
        stages.get(&card).map(|stage| MyCard {
            card,
            stage: *stage,
        })
    });
    Some(flow)
}

/// One row of the reservation read.
struct Reservation {
    card: i64,
    holder: String,
    status: String,
    lease_until: i64,
    worktree: Option<String>,
    acquired_at: i64,
    id: i64,
}

impl Reservation {
    /// Held (also for review) and with a lease that ends at or after the render's clock.
    fn is_live(&self, now: i64) -> bool {
        matches!(self.status.as_str(), "held" | "held_for_review") && self.lease_until >= now
    }

    fn is_held(&self) -> bool {
        matches!(self.status.as_str(), "held" | "held_for_review")
    }
}

/// The marker [`stage_of`] should see (SPEC-20261126-statusline-stale-submit REQ-1 to REQ-3): a
/// submit marker of a `ready` card whose note is strictly older than [`STALE_SUBMIT_SECS`] at the
/// render's clock is none, so the card counts as ready again; every other marker, and every other
/// state, passes unchanged. A note newer than the clock has a negative age and is fresh.
fn stage_marker(state: &str, marker: Option<Marker>, noted_at: i64, now: i64) -> Option<Marker> {
    match marker {
        Some(Marker::Submitted)
            if state == "ready" && now.saturating_sub(noted_at) > STALE_SUBMIT_SECS =>
        {
            None
        }
        other => other,
    }
}

/// The stage of one card by its ledger state, its newest marker, a landing event and a live
/// lease (D-1 and D-2 of the SPEC-20261122-statusline-flow, D-1 to D-5 of
/// SPEC-20261124-statusline-counts-truth). `marker` and `landed` are read at or after the card's
/// last reservation by the statements. `None` for a state that is not open. `draft` and `hold`
/// keep their own stage; a `ready` card is submitted or audited when a lane handed it back and
/// released it (the release moved it to `ready`) and ready otherwise; of the cards a lane works
/// on, a landing wins, then the newest marker (blocked, audited, submitted), then `review` itself
/// is submitted, and the rest are working with a live lease and stale without one.
fn stage_of(state: &str, marker: Option<Marker>, landed: bool, live: bool) -> Option<Stage> {
    match state {
        "draft" => Some(Stage::Draft),
        "ready" => Some(match marker {
            Some(Marker::Submitted) => Stage::Submitted,
            Some(Marker::Audited) => Stage::Audited,
            _ => Stage::Ready,
        }),
        "hold" => Some(Stage::Hold),
        "active" | "review" => Some(if landed {
            Stage::Landed
        } else {
            match marker {
                Some(Marker::Blocked) => Stage::Blocked,
                Some(Marker::Audited) => Stage::Audited,
                Some(Marker::Submitted) => Stage::Submitted,
                _ if state == "review" => Stage::Submitted,
                _ if live => Stage::Working,
                _ => Stage::Stale,
            }
        }),
        _ => None,
    }
}

/// My card (SPEC REQ-6, D-3): the lowest-numbered card whose held reservation names the folder
/// the session works in (the worktree), else the newest live held reservation of the stdin
/// session id (`acquired_at`, then the reservation id, decides which is newest). Never the
/// lowest number of a session's several cards.
fn my_card(reservations: &[Reservation], who: &Whose, now: i64) -> Option<i64> {
    if let Some((found, _)) = who.dir.and_then(find_dot_git) {
        let canonical = std::fs::canonicalize(&found).unwrap_or_else(|_| found.clone());
        let by_folder = reservations
            .iter()
            .filter(|r| r.is_held())
            .filter(|r| {
                r.worktree
                    .as_deref()
                    .is_some_and(|held| names_folder(Path::new(held), &found, &canonical))
            })
            .map(|r| r.card)
            .min();
        if by_folder.is_some() {
            return by_folder;
        }
    }
    if who.session_id.is_empty() {
        return None;
    }
    reservations
        .iter()
        .filter(|r| r.is_live(now) && r.holder == who.session_id)
        .max_by_key(|r| (r.acquired_at, r.id))
        .map(|r| r.card)
}

/// Whether the folder a holder recorded (`named`) is the folder `found` (as found) or
/// `canonical`. The holder typed the path (`card take --worktree`), so it may be a link
/// spelling: it is compared as written and as a canonical path. A path that is not absolute
/// names no folder from here (it would be read against this process's working directory).
fn names_folder(named: &Path, found: &Path, canonical: &Path) -> bool {
    named.is_absolute()
        && (named == found
            || named == canonical
            || std::fs::canonicalize(named).is_ok_and(|c| c == canonical))
}

/// The project's ledger opened read-only the way [`read_flow`] reads it, with the project id.
/// `None` for every case [`read_flow`] documents (no HOME, no profile, no registered ledger,
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

/// How many characters of a note's body are read: enough for its first line, so a long report
/// costs no more than a short one.
const NOTE_HEAD_CHARS: usize = 1000;

/// What a note's first line says about its card (D-1 of SPEC-20261122-statusline-flow).
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
enum Marker {
    /// The card was pushed and waits for an audit or for the leader's integration.
    Submitted,
    /// An audit passed for the head.
    Audited,
    /// The lane asked for a decision.
    Blocked,
    /// The lane is working on the card again (took or resumed it, planned, or was told FAIL).
    Working,
}

/// The marker a note's first line carries, if any. Only the lines the lanes already write by
/// rule count; any other first line is no marker and changes nothing (`RESUME`, `waiting:`,
/// `LANE-HOLD` and `PAIR-` among them). This is the one first-line parser of the status line and
/// a temporary one: typed submit, block and verdict rows will replace it. `REPAIRED` hands a head
/// back like `LANE-READY-INTEGRATION` does, as the receipt gate reads it
/// (SPEC-20261124-statusline-counts-truth REQ-4).
fn marker_of(body: &str) -> Option<Marker> {
    let first = body.lines().next().unwrap_or("");
    let audit = first.starts_with("AUDIT");
    if first.starts_with("LANE-BLOCKED") {
        Some(Marker::Blocked)
    } else if audit && first.contains("verdict: PASS") {
        Some(Marker::Audited)
    } else if matches!(
        crate::card::ready_receipt::marker_of_line(first),
        Some("LANE-READY-INTEGRATION" | "REPAIRED" | "READY")
    ) || first.starts_with("MERGE-READY")
        || first.starts_with("감사 요청")
        || (first.starts_with("LANE-DONE")
            && (first.contains("audit requested") || first.contains("감사 요청")))
    {
        Some(Marker::Submitted)
    } else if first.starts_with("ACK ")
        || first.starts_with("PLAN ")
        || (audit && first.contains("verdict: FAIL"))
    {
        Some(Marker::Working)
    } else {
        None
    }
}

/// Whether a note is the leader's answer to a block (SPEC-20261124-statusline-counts-truth D-3):
/// a decision note signed `leader`. The typed columns are read, never the words of the note, so
/// no text can end a block by accident.
fn answers_a_block(kind: &str, author: &str) -> bool {
    kind == "decision" && author == "leader"
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

    /// Absurd counts stop at `i64::MAX` (a plain `+`
    /// panics in a debug build), and an object with no count gives no number at all.
    #[test]
    fn usage_tokens_saturate_and_an_empty_object_has_none() {
        let usage = |a, b, c| input::CurrentUsage {
            input_tokens: a,
            cache_creation_input_tokens: b,
            cache_read_input_tokens: c,
        };
        assert_eq!(
            usage_tokens(&usage(Some(i64::MAX), Some(i64::MAX), Some(1))),
            Some(i64::MAX)
        );
        assert_eq!(usage_tokens(&usage(Some(1), None, Some(2))), Some(3));
        assert_eq!(usage_tokens(&usage(None, None, Some(0))), Some(0));
        assert_eq!(usage_tokens(&usage(None, None, None)), None);
    }

    #[test]
    fn glm_names_map_only_non_claude_slot_models() {
        let ctx = Context::for_test(&[("ANTHROPIC_DEFAULT_OPUS_MODEL", "glm-5.4")], None);
        assert_eq!(resolve_glm_model_name("Opus", &ctx), "glm-5.4");
        assert_eq!(resolve_glm_model_name("Sonnet", &ctx), "Sonnet");
        // The glm launcher's value carries Claude Code's 1M marker (SPEC-20261123-glm-54-model);
        // the line shows the model without it, from the slot name and from the id alike.
        let launcher = Context::for_test(&[("ANTHROPIC_DEFAULT_OPUS_MODEL", "glm-5.4[1m]")], None);
        assert_eq!(resolve_glm_model_name("Opus", &launcher), "glm-5.4");
        assert_eq!(resolve_glm_model_name("Opus[1m]", &launcher), "glm-5.4");
        assert_eq!(resolve_glm_model_name("glm-5.4[1m]", &launcher), "glm-5.4");
        assert_eq!(resolve_glm_model_name("gpt-x[1m]", &ctx), "gpt-x");
        let claude =
            Context::for_test(&[("ANTHROPIC_DEFAULT_OPUS_MODEL", "claude-opus-5-5")], None);
        assert_eq!(resolve_glm_model_name("Opus", &claude), "Opus");
    }

    /// SPEC-20261123-glm-54-followups REQ-8, REQ-9 (D-5): the launcher also maps the fable alias,
    /// so a Fable slot name shows the fable slot's GLM model without the `[1m]` marker, and a
    /// fable-only environment still gives the 1M window. The environment is the launcher's own
    /// (`GLM_ENV`), so a change of the launcher's values cannot drift from this line unseen.
    #[test]
    fn glm_fable_slot_names_and_sizes_its_model() {
        let launcher: Vec<(&str, &str)> = crate::session::GLM_ENV.to_vec();
        let ctx = Context::for_test(&launcher, None);
        assert_eq!(resolve_glm_model_name("Fable", &ctx), "glm-5.4");
        assert_eq!(resolve_glm_model_name("Fable[1m]", &ctx), "glm-5.4");
        assert_eq!(resolve_glm_model_name("Opus", &ctx), "glm-5.4");
        assert_eq!(resolve_glm_model_name("Haiku", &ctx), "glm-5.3-flash");
        // A claude model in the fable slot stays a claude name.
        let claude = Context::for_test(
            &[("ANTHROPIC_DEFAULT_FABLE_MODEL", "claude-fable-5-2")],
            None,
        );
        assert_eq!(resolve_glm_model_name("Fable", &claude), "Fable");
        // Only the fable slot holds a model: the window comes from it, with or without `[1m]`.
        for value in ["glm-5.4[1m]", "glm-5.4"] {
            let only = Context::for_test(&[("ANTHROPIC_DEFAULT_FABLE_MODEL", value)], None);
            assert_eq!(context_window_override(&only), 1_000_000, "{value}");
        }
        let only_claude = Context::for_test(
            &[("ANTHROPIC_DEFAULT_FABLE_MODEL", "claude-fable-5-2")],
            None,
        );
        assert_eq!(context_window_override(&only_claude), 0);
    }

    /// SPEC-20261123-glm-fable-slot-test REQ-1, REQ-2: the launcher's own environment gives the
    /// opus and fable slots one value, so it cannot tell the two slots apart. Here each slot holds
    /// a model of its own, and a Fable name must show the fable slot's model, an Opus name the
    /// opus slot's; a fable name never borrows the opus slot.
    #[test]
    fn glm_fable_and_opus_slots_resolve_their_own_models() {
        let both = Context::for_test(
            &[
                ("ANTHROPIC_DEFAULT_OPUS_MODEL", "glm-opus-x[1m]"),
                ("ANTHROPIC_DEFAULT_FABLE_MODEL", "glm-fable-y[1m]"),
            ],
            None,
        );
        assert_eq!(resolve_glm_model_name("Fable", &both), "glm-fable-y");
        assert_eq!(resolve_glm_model_name("Fable[1m]", &both), "glm-fable-y");
        assert_eq!(resolve_glm_model_name("Opus", &both), "glm-opus-x");
        // Only the opus slot holds a GLM model: a Fable name has no slot model to show.
        let only_opus =
            Context::for_test(&[("ANTHROPIC_DEFAULT_OPUS_MODEL", "glm-opus-x[1m]")], None);
        assert_eq!(resolve_glm_model_name("Fable", &only_opus), "Fable");
        assert_eq!(resolve_glm_model_name("Opus", &only_opus), "glm-opus-x");
    }

    #[test]
    fn context_override_prefers_env_then_longest_builtin_match() {
        let env = Context::for_test(&[("MOAI_STATUSLINE_CONTEXT_SIZE", "128000")], None);
        assert_eq!(context_window_override(&env), 128_000);
        let glm = Context::for_test(&[("ANTHROPIC_DEFAULT_SONNET_MODEL", "GLM-5.3-Flash")], None);
        assert_eq!(context_window_override(&glm), 1_000_000);
        // glm-5.4 is a 1M model too, with or without the launcher's `[1m]` marker, when it is
        // the only slot model (SPEC-20261123-glm-54-model AC-2).
        for value in ["glm-5.4[1m]", "glm-5.4"] {
            let glm = Context::for_test(&[("ANTHROPIC_DEFAULT_OPUS_MODEL", value)], None);
            assert_eq!(context_window_override(&glm), 1_000_000, "{value}");
        }
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
        std::fs::create_dir_all(tmp.path().join(crate::fixed_state!())).unwrap();
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
    fn forge_reads() {
        let tmp = tempfile::tempdir().unwrap();
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
        // SPEC-20261122-statusline-where REQ-2: the git folder (where `HEAD` is) and the common
        // folder (where the refs are) of a linked worktree and of the main checkout.
        let linked = git_dirs(&std::fs::canonicalize(wt.join("sub")).unwrap()).unwrap();
        assert_eq!(linked.git_dir, gitdir);
        assert_eq!(linked.common, gitdir.join("../.."));
        let primary = git_dirs(&canon).unwrap();
        assert_eq!(primary.git_dir, canon.join(".git"));
        assert_eq!(primary.common, canon.join(".git"));
        assert!(git_dirs(&std::fs::canonicalize(tmp.path()).unwrap()).is_none());
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
        assert!(read_git(&plain, &ctx, None, "").is_none());
        assert_eq!(calls(), 0);

        let repo = tmp.path().join("repo");
        std::fs::create_dir(&repo).unwrap();
        git_init(&repo);
        assert_eq!(
            read_git(&repo, &ctx, None, "").map(|r| r.git.branch),
            Some("main".into())
        );
        assert_eq!(calls(), 1);

        // An expired deadline: nothing is started, the numbers are zeros, and the branch still
        // comes from the `HEAD` file. The wrapper log cannot show this (a started child is
        // killed before it writes), so the spawn attempts of this thread are counted.
        ctx.deadline = Some(Instant::now());
        let before = GIT_SPAWNS.with(|n| n.get());
        let spent = read_git(&repo, &ctx, None, "").unwrap();
        assert_eq!(spent.git.branch, "main");
        assert_eq!(GIT_SPAWNS.with(|n| n.get()), before);
        ctx.deadline = None;
        assert!(read_git(&repo, &ctx, None, "").is_some());
        assert_eq!(GIT_SPAWNS.with(|n| n.get()), before + 1);
        // A `HEAD` that names nothing usable: no branch and no process at all.
        std::fs::write(repo.join(".git/HEAD"), "garbage\n").unwrap();
        let nothing = read_git(&repo, &ctx, None, "").unwrap();
        assert_eq!(nothing.git, GitStatus::default());
        assert_eq!(GIT_SPAWNS.with(|n| n.get()), before + 1);
    }
}

/// #2756 AC-7 and AC-9: the pull request and worktree values the renderer shows.
#[cfg(test)]
mod extras_tests {
    use super::*;
    use serde_json::json;

    fn collect_json(value: serde_json::Value, now: i64) -> Status {
        let input = input::parse(value.to_string().as_bytes());
        let mut ctx = Context::for_test(&[], None);
        ctx.now = now;
        collect(Some(&input), &ctx)
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
    fn pr_and_worktree_come_from_the_input_or_stay_empty() {
        let s = collect_json(
            json!({
                "pr": {"number": 12, "url": "https://github.com/o/r/pull/12", "review_state": "approved"},
                "worktree": {"name": "wt-session"},
                "workspace": {"git_worktree": "linked"}
            }),
            0,
        );
        assert_eq!(
            s.pr,
            Some(PrInfo {
                number: 12,
                url: "https://github.com/o/r/pull/12".into(),
                review_state: "approved".into(),
                merge_request: false,
            })
        );
        assert_eq!(
            s.worktree, "wt-session",
            "worktree.name wins over git_worktree"
        );
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
        assert_eq!((s.pr, s.worktree.as_str()), (None, ""));
    }
}

/// Card C-65: the card counts read from a ledger in a temporary HOME; the real HOME is never
/// touched.
/// AC-1 and AC-2 of SPEC-20261122-statusline-flow: the stage rule and the marker words, read
/// from a fixture ledger. The older tests of the card-count era stay as the regression of what
/// did not change: they read the flow through [`legacy`], the three numbers of that era
/// (working = working + stale, waiting = submitted + audited, remaining = draft + ready + hold).
#[cfg(test)]
mod stage_tests {
    use super::*;

    /// The card-count era's four numbers, rebuilt from a [`Flow`] for the older tests.
    #[derive(Debug, Default, Clone, Copy, PartialEq, Eq)]
    struct CardCounts {
        in_progress: i64,
        waiting: Option<i64>,
        remaining: i64,
        done: i64,
    }

    fn legacy(flow: Flow) -> CardCounts {
        CardCounts {
            in_progress: flow.working + flow.stale,
            waiting: flow.split.then_some(flow.submitted + flow.audited),
            remaining: flow.draft + flow.ready + flow.hold,
            done: flow.done_today,
        }
    }

    fn card_counts(project: &Path, ctx: &Context) -> Option<CardCounts> {
        read_flow(project, ctx, &Whose::default()).map(legacy)
    }

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

    /// REQ-11 and the read-only rule (audit N3): the connection the status line reads the ledger
    /// through is a read-only one, so a statement that writes is refused by the database itself,
    /// and a whole flow read leaves the ledger file as it was. A read-write open would fail
    /// the first check.
    #[test]
    fn the_status_line_opens_the_ledger_read_only() {
        let f = fixture();
        crate::ledger::register(&f.home, "demo").unwrap();
        f.add("active", 2);
        let before = std::fs::read(f.ledger_file()).unwrap();

        let (ledger, id) = open_project_ledger(&f.project, &f.ctx()).unwrap();
        assert_eq!(id, "demo");
        assert!(ledger.conn().is_readonly("main").unwrap());
        let refused = ledger
            .conn()
            .execute("UPDATE card SET state = 'done'", [])
            .unwrap_err();
        assert!(
            refused.to_string().to_lowercase().contains("readonly"),
            "{refused}"
        );
        drop(ledger);

        let flow = read_flow(&f.project, &f.ctx(), &Whose::default()).unwrap();
        assert!(flow.split);
        assert_eq!(std::fs::read(f.ledger_file()).unwrap(), before);
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
        let who = Whose::default();
        // The render budget is spent before the read: no flow at all.
        assert_eq!(read_flow_with(&ledger, 0, &who, &|| true), None);
        // With time left the same ledger reads: the active card and the review card wait.
        let flow = read_flow_with(&ledger, 0, &who, &|| false).unwrap();
        assert!(flow.split);
        assert_eq!((flow.submitted, flow.working, flow.stale), (2, 0, 0));
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
        // The first question (before the read) is "no", the next ones "yes": the open cards
        // were read, the notes were not, so the unsplit counts stand and say so.
        let flow = read_flow_with(&ledger, 0, &Whose::default(), &spent).unwrap();
        assert!(
            !flow.split,
            "a spent budget in the split gives the unsplit flow"
        );
        assert_eq!((flow.working, flow.submitted, flow.stale), (1, 0, 0));
        assert!(
            asked.get() >= 2,
            "the split asked again after the first read"
        );
        // Never spent: the same ledger reads and the card waits.
        let whole = read_flow_with(&ledger, 0, &Whose::default(), &|| false).unwrap();
        assert!(whole.split);
        assert_eq!((whole.submitted, whole.working, whole.stale), (1, 0, 0));
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
    /// tested before `FAIL`, so a FAIL line that quotes an earlier PASS reads as audited. The
    /// real verdict line ends at the verdict word, so this only pins the rule.
    #[test]
    fn a_line_with_both_verdicts_reads_as_audited() {
        let both = "AUDIT C-1 (x) head abc base def — verdict: FAIL (was verdict: PASS)";
        assert_eq!(marker_of(both), Some(Marker::Audited));
        assert_eq!(marker_of(FAIL), Some(Marker::Working));
        assert_eq!(marker_of(PASS), Some(Marker::Audited));
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
        );
        let ctx = Context::for_test(&[("HOME", home)], Some(&elsewhere));
        let counts_of = |status: Status| status.flow.map(legacy);
        assert_eq!(counts_of(collect(Some(&named), &ctx)), counts(0, 2, 1));
        let empty = input::parse(b"{}");
        assert_eq!(
            counts_of(collect(Some(&empty), &ctx)),
            None,
            "no project here"
        );
        std::fs::create_dir_all(f.project.join(".moai")).unwrap();
        let inside = Context::for_test(&[("HOME", home)], Some(&f.project));
        assert_eq!(counts_of(collect(Some(&empty), &inside)), counts(0, 2, 1));
    }
}

/// A fixture ledger for the stage tests: a HOME with the project `demo` and its registered
/// ledger, with helpers to add cards, notes, reservations and landing events.
#[cfg(test)]
mod flow_support {
    use super::*;

    pub(super) struct Fx {
        pub(super) _tmp: tempfile::TempDir,
        pub(super) home: PathBuf,
        pub(super) project: PathBuf,
    }

    pub(super) fn fx() -> Fx {
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
        crate::ledger::register(&home, "demo").unwrap();
        Fx {
            _tmp: tmp,
            home,
            project,
        }
    }

    impl Fx {
        /// One card in `state` with its notes oldest first; returns its number. A `done` card
        /// closes at the ledger's own clock.
        pub(super) fn card(&self, state: &str, notes: &[(&str, &str)]) -> i64 {
            let mut ledger = crate::ledger::open_write(&self.home, "demo").unwrap();
            let mut id = 0;
            ledger
                .with_write_tx(|tx, now| {
                    tx.execute(
                        "INSERT INTO card (title, state, created_by, created_at, updated_at, closed_at) \
                         VALUES ('t', ?1, 'test', ?2, ?2, iif(?1 = 'done', ?2, NULL))",
                        rusqlite::params![state, now],
                    )?;
                    id = tx.last_insert_rowid();
                    for (kind, body) in notes {
                        tx.execute(
                            "INSERT INTO card_note (card_id, kind, author, body, created_at) \
                             VALUES (?1, ?2, 'test', ?3, ?4)",
                            rusqlite::params![id, kind, body, now],
                        )?;
                    }
                    Ok(())
                })
                .unwrap();
            id
        }

        /// A `done` card closed at `closed_at`.
        pub(super) fn done_at(&self, closed_at: i64) {
            let mut ledger = crate::ledger::open_write(&self.home, "demo").unwrap();
            ledger
                .with_write_tx(|tx, now| {
                    tx.execute(
                        "INSERT INTO card (title, state, created_by, created_at, updated_at, closed_at) \
                         VALUES ('t', 'done', 'test', ?1, ?1, ?2)",
                        rusqlite::params![now, closed_at],
                    )?;
                    Ok(())
                })
                .unwrap();
        }

        /// A reservation of `card` (generation 1).
        pub(super) fn reserve(
            &self,
            card: i64,
            holder: &str,
            status: &str,
            acquired_at: i64,
            lease_until: i64,
            worktree: Option<&str>,
        ) {
            let mut ledger = crate::ledger::open_write(&self.home, "demo").unwrap();
            ledger
                .with_write_tx(|tx, _| {
                    tx.execute(
                        "INSERT INTO reservation (card_id, generation, holder, status, acquired_at, \
                         lease_until, updated_at, worktree, branch) \
                         VALUES (?1, 1, ?2, ?3, ?4, ?5, ?4, ?6, NULL)",
                        rusqlite::params![card, holder, status, acquired_at, lease_until, worktree],
                    )?;
                    Ok(())
                })
                .unwrap();
        }

        /// A `card.landed` event of `card`.
        pub(super) fn landed(&self, card: i64) {
            let mut ledger = crate::ledger::open_write(&self.home, "demo").unwrap();
            ledger
                .with_write_tx(|tx, now| {
                    tx.execute(
                        "INSERT INTO event (event_key, at, actor, card_id, kind) \
                         VALUES (?1, ?2, 'test', ?3, 'card.landed')",
                        rusqlite::params![format!("landed-{card}"), now, card],
                    )?;
                    Ok(())
                })
                .unwrap();
        }

        /// Criterion `seq` of `card` in `status`.
        pub(super) fn criterion(&self, card: i64, seq: i64, status: &str) {
            let mut ledger = crate::ledger::open_write(&self.home, "demo").unwrap();
            ledger
                .with_write_tx(|tx, now| {
                    tx.execute(
                        "INSERT INTO card_ac (card_id, seq, text, status, evidence_ref, \
                         withdrawn_reason, updated_at) VALUES (?1, ?2, 'c', ?3, \
                         iif(?3 = 'met', 'e', NULL), iif(?3 = 'withdrawn', 'r', NULL), ?4)",
                        rusqlite::params![card, seq, status, now],
                    )?;
                    Ok(())
                })
                .unwrap();
        }

        /// A reservation of `card` with its generation, for a card taken more than once.
        pub(super) fn reserve_gen(
            &self,
            card: i64,
            generation: i64,
            status: &str,
            acquired_at: i64,
            lease_until: i64,
        ) {
            let mut ledger = crate::ledger::open_write(&self.home, "demo").unwrap();
            ledger
                .with_write_tx(|tx, _| {
                    tx.execute(
                        "INSERT INTO reservation (card_id, generation, holder, status, \
                         acquired_at, lease_until, updated_at, worktree, branch) \
                         VALUES (?1, ?2, 'lane', ?3, ?4, ?5, ?4, NULL, NULL)",
                        rusqlite::params![card, generation, status, acquired_at, lease_until],
                    )?;
                    Ok(())
                })
                .unwrap();
        }

        /// One note of `card` written at `at` by `author`; the order of calls is the order of the
        /// note ids.
        pub(super) fn note_at(&self, card: i64, kind: &str, author: &str, body: &str, at: i64) {
            let mut ledger = crate::ledger::open_write(&self.home, "demo").unwrap();
            ledger
                .with_write_tx(|tx, _| {
                    tx.execute(
                        "INSERT INTO card_note (card_id, kind, author, body, created_at) \
                         VALUES (?1, ?2, ?3, ?4, ?5)",
                        rusqlite::params![card, kind, author, body, at],
                    )?;
                    Ok(())
                })
                .unwrap();
        }

        /// One event of `kind` of `card` at `at`.
        pub(super) fn event_at(&self, card: i64, kind: &str, at: i64) {
            let mut ledger = crate::ledger::open_write(&self.home, "demo").unwrap();
            ledger
                .with_write_tx(|tx, _| {
                    tx.execute(
                        "INSERT INTO event (event_key, at, actor, card_id, kind) \
                         VALUES (?1, ?2, 'test', ?3, ?4)",
                        rusqlite::params![format!("{kind}-{card}-{at}"), at, card, kind],
                    )?;
                    Ok(())
                })
                .unwrap();
        }

        /// The flow at the render clock `now` for `who`.
        pub(super) fn flow(&self, now: i64, who: &Whose) -> Flow {
            self.flow_in(now, who, DayZone::Utc)
        }

        /// The flow at `now` with "done today" counted in `zone`.
        pub(super) fn flow_in(&self, now: i64, who: &Whose, zone: DayZone) -> Flow {
            let mut ctx = Context::for_test(&[("HOME", self.home.to_str().unwrap())], None);
            ctx.now = now;
            ctx.day_zone = zone;
            read_flow(&self.project, &ctx, who).unwrap()
        }
    }
}

/// AC-1 of SPEC-20261122-statusline-flow: every edge of the stage rule has one answer, and the
/// nine counts add up to the open cards.
#[cfg(test)]
mod stage_tests_edges {
    use super::flow_support::fx;
    use super::*;

    const NOW: i64 = 1_800_000_000;

    /// SPEC-20261124-card-flow-icons REQ-1: every stage has its icon from the one icon set — the
    /// table of `statusline::icons` — five icons for the nine stages, each one in the table of
    /// icons the lines print, and none of the icons that left the set.
    #[test]
    fn every_stage_has_its_icon_from_the_one_icon_set() {
        use crate::statusline::icons::{BLOCKED, CHECK, QUESTION, READY, RETIRED, TODO, WORK};
        let table = [
            (Stage::Draft, TODO),
            (Stage::Ready, TODO),
            (Stage::Working, WORK),
            (Stage::Stale, WORK),
            (Stage::Submitted, CHECK),
            (Stage::Audited, READY),
            (Stage::Landed, READY),
            (Stage::Blocked, BLOCKED),
            (Stage::Hold, QUESTION),
        ];
        for (stage, icon) in table {
            assert_eq!(stage.icon(), icon, "{stage:?}");
            assert!(
                crate::statusline::width::ICONS.contains(&icon),
                "{icon} is not in the table"
            );
        }
        let mut icons: Vec<&str> = table.iter().map(|(s, _)| s.icon()).collect();
        icons.sort_unstable();
        icons.dedup();
        assert_eq!(icons.len(), 6, "{icons:?}");
        for gone in RETIRED {
            assert!(!icons.contains(&gone), "{gone} left the set");
        }
    }

    #[test]
    fn the_stage_rule_has_one_answer_for_every_edge() {
        use Marker::{Audited, Blocked, Submitted, Working};
        use Stage::{Draft, Hold, Landed, Ready, Stale};
        let s = Stage::Submitted;
        let a = Stage::Audited;
        let b = Stage::Blocked;
        let w = Stage::Working;
        // (state, newest marker, landing event, live lease) -> stage
        type Row = (&'static str, Option<Marker>, bool, bool, Option<Stage>);
        let table: [Row; 35] = [
            // draft and hold keep their own stage whatever else is true of them.
            ("draft", None, false, false, Some(Draft)),
            ("draft", Some(Submitted), true, true, Some(Draft)),
            // a ready card is ready, unless a lane handed it back (submitted) or an audit passed
            // for it (audited) after its last take; the other markers and a landing change nothing.
            ("ready", None, false, false, Some(Ready)),
            ("ready", Some(Blocked), true, false, Some(Ready)),
            ("ready", Some(Working), false, true, Some(Ready)),
            ("ready", Some(Submitted), false, false, Some(s)),
            ("ready", Some(Submitted), true, true, Some(s)),
            ("ready", Some(Audited), false, false, Some(a)),
            ("hold", Some(Audited), true, true, Some(Hold)),
            // active: working with a live lease, stale without one.
            ("active", None, false, true, Some(w)),
            ("active", None, false, false, Some(Stale)),
            ("active", Some(Working), false, true, Some(w)),
            ("active", Some(Working), false, false, Some(Stale)),
            // submitted, audited and blocked keep their stage whatever the lease says.
            ("active", Some(Submitted), false, false, Some(s)),
            ("active", Some(Submitted), false, true, Some(s)),
            ("active", Some(Audited), false, false, Some(a)),
            ("active", Some(Audited), false, true, Some(a)),
            ("active", Some(Blocked), false, true, Some(b)),
            ("active", Some(Blocked), false, false, Some(b)),
            // review is submitted unless a marker says audited or blocked.
            ("review", None, false, false, Some(s)),
            ("review", None, false, true, Some(s)),
            ("review", Some(Working), false, true, Some(s)),
            ("review", Some(Submitted), false, false, Some(s)),
            ("review", Some(Audited), false, false, Some(a)),
            ("review", Some(Blocked), false, false, Some(b)),
            // a landing event wins over every marker and lease of an active or review card.
            ("active", Some(Blocked), true, true, Some(Landed)),
            ("active", None, true, false, Some(Landed)),
            ("active", Some(Working), true, true, Some(Landed)),
            ("review", Some(Audited), true, false, Some(Landed)),
            ("review", None, true, true, Some(Landed)),
            // a closed card, and a state the ledger does not have, are no open card.
            ("done", None, false, false, None),
            ("done", Some(Submitted), true, true, None),
            ("bogus", None, false, false, None),
            ("", Some(Audited), false, true, None),
            ("Active", None, false, true, None),
        ];
        for (state, marker, landed, live, want) in table {
            assert_eq!(
                stage_of(state, marker, landed, live),
                want,
                "{state} {marker:?} landed={landed} live={live}"
            );
        }
    }

    #[test]
    fn the_nine_counts_add_up_to_the_open_cards_of_a_fixture_ledger() {
        let f = fx();
        f.card("draft", &[]);
        f.card("ready", &[]);
        f.card("hold", &[]);
        let working = f.card("active", &[]);
        f.reserve(working, "s-1", "held", 10, NOW + 60, None);
        f.card("active", &[]); // no reservation at all: stale
        f.card("active", &[("report", "READY (head abc)")]);
        f.card(
            "active",
            &[("report", "AUDIT C-1 head abc base def — verdict: PASS")],
        );
        f.card("active", &[("report", "LANE-BLOCKED kind=decision x")]);
        let landed = f.card("active", &[("report", "READY (head abc)")]);
        f.landed(landed);
        // Three cards closed today, and one closed long ago that still carries a marker note: a
        // closed card is never an open stage.
        f.done_at(NOW);
        f.done_at(NOW + 1);
        f.done_at(NOW + 2);
        f.card("done", &[("report", "READY (head abc)")]);
        let flow = f.flow(NOW, &Whose::default());
        assert!(flow.split);
        assert_eq!(
            (
                flow.draft,
                flow.ready,
                flow.working,
                flow.stale,
                flow.submitted,
                flow.audited,
                flow.landed,
                flow.blocked,
                flow.hold,
            ),
            (1, 1, 1, 1, 1, 1, 1, 1, 1),
            "one card in every stage"
        );
        let sum = flow.draft
            + flow.ready
            + flow.working
            + flow.stale
            + flow.submitted
            + flow.audited
            + flow.landed
            + flow.blocked
            + flow.hold;
        assert_eq!(sum, 9, "the nine counts are the open cards");
        assert_eq!(flow.done_today, 3, "closed cards are never an open stage");
    }

    /// A lease that ends at the render's clock is still live; one second less is over. Only
    /// `held` and `held_for_review` reservations count as a live holder.
    #[test]
    fn the_lease_edge_and_the_status_of_a_reservation_decide_working_or_stale() {
        let who = Whose::default();
        for (status, lease, want_working) in [
            ("held", NOW, true),
            ("held", NOW + 1, true),
            ("held", NOW - 1, false),
            ("held_for_review", NOW + 60, true),
            ("held_for_review", NOW - 1, false),
            ("expired_candidate", NOW + 60, false),
            ("released", NOW + 60, false),
            ("reclaimed", NOW + 60, false),
        ] {
            let f = fx();
            let card = f.card("active", &[]);
            f.reserve(card, "s-1", status, 10, lease, None);
            let flow = f.flow(NOW, &who);
            assert_eq!(
                (flow.working, flow.stale),
                if want_working { (1, 0) } else { (0, 1) },
                "{status} lease_until={lease}"
            );
        }
    }

    /// A landed card with open criteria is counted apart with its number of open criteria, in
    /// place of the `ready` or `landed` count; a landed card with every criterion met, a card
    /// that never landed and a card a lane works on again keep their stage.
    #[test]
    fn partial_landing_cards_are_counted_apart_with_their_open_criteria() {
        let f = fx();
        // Ready, landed once, two open and one met: partial.
        let given_back = f.card("ready", &[]);
        f.landed(given_back);
        f.criterion(given_back, 1, "open");
        f.criterion(given_back, 2, "open");
        f.criterion(given_back, 3, "met");
        // Active and landed since its reservation, one open and one withdrawn: partial.
        let held = f.card("active", &[]);
        f.reserve(held, "s-1", "held", 10, NOW + 60, None);
        f.landed(held);
        f.criterion(held, 1, "open");
        f.criterion(held, 2, "withdrawn");
        // Landed with nothing open: still the landed stage.
        let full = f.card("active", &[]);
        f.reserve(full, "s-2", "held", 10, NOW + 60, None);
        f.landed(full);
        f.criterion(full, 1, "met");
        // Never landed, with an open criterion: ready.
        let plain = f.card("ready", &[]);
        f.criterion(plain, 1, "open");
        // Landed before the reservation that works on the rest: working.
        let again = f.card("active", &[]);
        f.landed(again);
        f.reserve(again, "s-3", "held", i64::MAX / 2, NOW + 60, None);
        f.criterion(again, 1, "open");

        let flow = f.flow(NOW, &Whose::default());
        assert!(flow.split);
        assert_eq!(
            (flow.partial, flow.partial_open),
            (2, 3),
            "two partial landings with three open criteria: {flow:?}"
        );
        assert_eq!(
            (flow.ready, flow.landed, flow.working),
            (1, 1, 1),
            "{flow:?}"
        );
    }

    #[test]
    fn the_four_statements_read_and_never_write() {
        assert_eq!(FLOW_STATEMENTS.len(), 4);
        for sql in FLOW_STATEMENTS {
            assert!(sql.trim_start().starts_with("SELECT"), "{sql}");
            // Whole words: a column named `created_at` is no CREATE statement.
            let upper = sql.to_ascii_uppercase();
            let words: Vec<&str> = upper
                .split(|c: char| !c.is_ascii_alphanumeric() && c != '_')
                .collect();
            for word in [
                "INSERT", "UPDATE", "DELETE", "CREATE", "DROP", "PRAGMA", "ALTER",
            ] {
                assert!(!words.contains(&word), "{word} in {sql}");
            }
        }
    }

    /// Two live holders, a worktree: the folder wins over the newer card of the session id.
    #[test]
    fn my_card_is_the_worktree_card_then_the_newest_live_card_of_the_session() {
        let f = fx();
        let wt = tempfile::tempdir().unwrap();
        std::fs::create_dir(wt.path().join(".git")).unwrap();
        let wt_path = std::fs::canonicalize(wt.path()).unwrap();
        let folder = wt_path.to_str().unwrap();
        let old = f.card("active", &[]);
        let new = f.card("active", &[("report", "READY (head abc)")]);
        let expired = f.card("active", &[]);
        let in_folder = f.card("active", &[]);
        f.reserve(old, "s-1", "held", 100, NOW + 60, None);
        f.reserve(new, "s-1", "held", 200, NOW + 60, None);
        f.reserve(expired, "s-1", "held", 300, NOW - 1, None);
        f.reserve(in_folder, "s-2", "held", 50, NOW + 60, Some(folder));

        let by_session = Whose {
            session_id: "s-1",
            dir: None,
        };
        let mine = f.flow(NOW, &by_session).mine.unwrap();
        assert_eq!(
            mine.card, new,
            "the newest live card, not the lowest number"
        );
        assert_eq!(
            mine.stage,
            Stage::Submitted,
            "its stage is the stage rule's"
        );

        let in_worktree = Whose {
            session_id: "s-1",
            dir: Some(wt_path.as_path()),
        };
        assert_eq!(
            f.flow(NOW, &in_worktree).mine.map(|m| m.card),
            Some(in_folder),
            "the card whose reservation names the worktree comes first"
        );
        let stranger = Whose {
            session_id: "s-9",
            dir: None,
        };
        assert_eq!(
            f.flow(NOW, &stranger).mine,
            None,
            "a session that holds nothing"
        );
        let anonymous = Whose::default();
        assert_eq!(
            f.flow(NOW, &anonymous).mine,
            None,
            "no session id, no folder"
        );
        // Only the expired card is left to the session: never chosen.
        let g = fx();
        let gone = g.card("active", &[]);
        g.reserve(gone, "s-1", "held", 300, NOW - 1, None);
        assert_eq!(g.flow(NOW, &by_session).mine, None);
        // A card closed since the reservation is no open card, so no card of mine.
        let h = fx();
        let closed = h.card("done", &[]);
        h.reserve(closed, "s-1", "held", 300, NOW + 60, None);
        assert_eq!(h.flow(NOW, &by_session).mine, None);
    }

    #[test]
    fn two_cards_naming_one_folder_give_the_lowest_number_and_a_relative_path_names_none() {
        let f = fx();
        let wt = tempfile::tempdir().unwrap();
        std::fs::create_dir(wt.path().join(".git")).unwrap();
        let wt_path = std::fs::canonicalize(wt.path()).unwrap();
        let folder = wt_path.to_str().unwrap();
        let first = f.card("active", &[]);
        let second = f.card("active", &[]);
        let relative = f.card("active", &[]);
        // Each card holds one reservation; the second names the folder in a link spelling too.
        f.reserve(second, "a", "held", 1, NOW + 60, Some(folder));
        f.reserve(first, "b", "held_for_review", 1, NOW + 60, Some(folder));
        f.reserve(relative, "c", "held", 1, NOW + 60, Some("relative/path"));
        let here = Whose {
            session_id: "",
            dir: Some(wt_path.as_path()),
        };
        assert_eq!(f.flow(NOW, &here).mine.map(|m| m.card), Some(first));
        let elsewhere = Whose {
            session_id: "",
            dir: Some(Path::new("relative/path")),
        };
        assert_eq!(f.flow(NOW, &elsewhere).mine, None);
    }
}

/// AC-2 of SPEC-20261126-statusline-stale-submit: a submit marker of a ready card older than the
/// limit is no marker, at its edges; no other marker, state or age changes.
#[cfg(test)]
mod stale_submit_tests {
    use super::*;

    const NOW: i64 = 1_800_000_000;

    /// The stage of a card whose newest marker note is `age` seconds old at `NOW`, through the
    /// pure function and `stage_of`, as `split_flow` calls them.
    fn stage_at(state: &str, marker: Option<Marker>, age: i64) -> Option<Stage> {
        stage_of(
            state,
            stage_marker(state, marker, NOW - age, NOW),
            false,
            false,
        )
    }

    #[test]
    fn a_ready_submission_goes_stale_strictly_after_the_limit() {
        let table = [
            (
                STALE_SUBMIT_SECS - 1,
                Some(Marker::Submitted),
                Stage::Submitted,
            ),
            (STALE_SUBMIT_SECS, Some(Marker::Submitted), Stage::Submitted),
            (STALE_SUBMIT_SECS + 1, None, Stage::Ready),
            (30 * STALE_SUBMIT_SECS, None, Stage::Ready),
            // A note newer than the clock is fresh.
            (-1, Some(Marker::Submitted), Stage::Submitted),
            (
                -STALE_SUBMIT_SECS - 1,
                Some(Marker::Submitted),
                Stage::Submitted,
            ),
        ];
        for (age, kept, stage) in table {
            assert_eq!(
                stage_marker("ready", Some(Marker::Submitted), NOW - age, NOW),
                kept,
                "age {age}"
            );
            assert_eq!(
                stage_at("ready", Some(Marker::Submitted), age),
                Some(stage),
                "age {age}"
            );
        }
    }

    #[test]
    fn no_other_marker_state_or_age_changes() {
        let ages = [
            -1,
            0,
            STALE_SUBMIT_SECS,
            STALE_SUBMIT_SECS + 1,
            3 * STALE_SUBMIT_SECS,
        ];
        // The marker passes unchanged for every state and age but a ready card's submission.
        for state in ["draft", "ready", "active", "review", "hold", "done"] {
            for marker in [
                None,
                Some(Marker::Submitted),
                Some(Marker::Audited),
                Some(Marker::Blocked),
                Some(Marker::Working),
            ] {
                for age in ages {
                    if state == "ready"
                        && marker == Some(Marker::Submitted)
                        && age > STALE_SUBMIT_SECS
                    {
                        continue;
                    }
                    assert_eq!(
                        stage_marker(state, marker, NOW - age, NOW),
                        marker,
                        "{state} {marker:?} age {age}"
                    );
                }
            }
        }
        // The resulting stages: an audit PASS keeps its stage for three days (D-3), a block or a
        // working marker on a ready card is as before, and a held card is not aged (D-4).
        let old = 3 * STALE_SUBMIT_SECS;
        assert_eq!(
            stage_at("ready", Some(Marker::Audited), old),
            Some(Stage::Audited)
        );
        assert_eq!(
            stage_at("ready", Some(Marker::Blocked), old),
            Some(Stage::Ready)
        );
        assert_eq!(
            stage_at("ready", Some(Marker::Working), old),
            Some(Stage::Ready)
        );
        assert_eq!(stage_at("ready", None, old), Some(Stage::Ready));
        assert_eq!(
            stage_at("active", Some(Marker::Submitted), 2 * STALE_SUBMIT_SECS),
            Some(Stage::Submitted)
        );
        assert_eq!(
            stage_at("review", None, 2 * STALE_SUBMIT_SECS),
            Some(Stage::Submitted)
        );
        assert_eq!(
            stage_at("review", Some(Marker::Submitted), 2 * STALE_SUBMIT_SECS),
            Some(Stage::Submitted)
        );
        assert_eq!(stage_at("draft", None, old), Some(Stage::Draft));
        assert_eq!(stage_at("hold", None, old), Some(Stage::Hold));
        assert_eq!(stage_at("done", Some(Marker::Submitted), old), None);
    }
}

/// AC-2 of SPEC-20261122-statusline-flow: the marker words at their edges.
#[cfg(test)]
mod marker_tests {
    use super::flow_support::fx;
    use super::*;

    #[test]
    fn each_word_alone_is_the_marker_the_rule_names() {
        for (line, want) in [
            ("LANE-READY-INTEGRATION C-1 head=abc", Marker::Submitted),
            ("MERGE-READY (2026-11-20) lane x", Marker::Submitted),
            ("READY (SPEC revision 2) agent-17", Marker::Submitted),
            ("READY head abc", Marker::Submitted),
            ("감사 요청 tier 1", Marker::Submitted),
            (
                "LANE-DONE a C-1 pushed — audit requested",
                Marker::Submitted,
            ),
            ("LANE-DONE a C-1 pushed — 감사 요청", Marker::Submitted),
            ("AUDIT C-1 head abc — verdict: PASS", Marker::Audited),
            ("LANE-BLOCKED kind=decision question", Marker::Blocked),
            ("LANE-BLOCKED", Marker::Blocked),
            ("ACK agent-1 C-1", Marker::Working),
            ("PLAN C-1 steps", Marker::Working),
            ("AUDIT C-1 head abc — verdict: FAIL", Marker::Working),
        ] {
            assert_eq!(marker_of(line), Some(want), "{line}");
        }
    }

    #[test]
    fn the_words_that_are_no_marker_change_nothing() {
        for line in [
            "RESUME head abc",
            "RESUME (weekly usage limit reached)",
            "waiting: audit since 08:00Z",
            "LANE-HOLD C-1",
            "PAIR-REQUEST agent-2",
            "PAIR-OK",
            "MERGED-LOCAL (2026-11-22, leader): landed",
            "LANE-DONE a C-1 slice done",
            "AUDIT C-1 head abc",
            "LEADER 2026-11-19 — decision",
            "",
            " READY (leading space)",
            "ready (lower case)",
        ] {
            assert_eq!(marker_of(line), None, "{line:?}");
        }
    }

    /// `READY ` is the word and one space: the same letters inside a longer word, at the end of
    /// the line, or before a tab or a line break are no marker.
    #[test]
    fn ready_needs_its_space_and_is_not_a_prefix_of_a_longer_word() {
        assert_eq!(marker_of("READY x"), Some(Marker::Submitted));
        assert_eq!(marker_of("READY  two spaces"), Some(Marker::Submitted));
        for line in [
            "READYX",
            "READY",
            "READY\tx",
            "READY\nx",
            "READYNESS review",
            "READY-FOR-AUDIT",
        ] {
            assert_eq!(marker_of(line), None, "{line:?}");
        }
        // `MERGE-READY` is a prefix match by design (the rule's word is the first token).
        assert_eq!(marker_of("MERGE-READY"), Some(Marker::Submitted));
        assert_eq!(marker_of("MERGE-READYX"), Some(Marker::Submitted));
        assert_eq!(marker_of("XMERGE-READY"), None);
    }

    /// SPEC-20261124-ready-receipt-followup REQ-1: the status line counts a hand-back first line
    /// exactly when the one marker judge of the receipt gate says `LANE-READY-INTEGRATION`,
    /// `REPAIRED` or `READY` (`MERGE-READY` keeps its own prefix rule). `REPAIRED` was left out
    /// until SPEC-20261124-statusline-counts-truth REQ-4 counted it, as the gate does.
    #[test]
    fn status_line_marker_follows_marker_of_line() {
        for line in [
            "LANE-READY-INTEGRATION C-1 head=abc",
            "LANE-READY-INTEGRATION: C-1 head=abc",
            "  LANE-READY-INTEGRATION C-1",
            "READY C-1 head=abc",
            "  READY C-1 head=abc",
            "REPAIRED C-1 head=abc",
            "READYX",
            "see READY x",
            "",
        ] {
            let counted = matches!(
                crate::card::ready_receipt::marker_of_line(line),
                Some("LANE-READY-INTEGRATION" | "REPAIRED" | "READY")
            );
            assert_eq!(
                marker_of(line) == Some(Marker::Submitted),
                counted,
                "{line:?}"
            );
        }
    }

    /// A marker word only counts on the first line: each word on line two is no marker.
    #[test]
    fn a_marker_on_the_second_line_is_none() {
        for word in [
            "READY (head abc)",
            "MERGE-READY (x)",
            "LANE-READY-INTEGRATION C-1",
            "LANE-BLOCKED kind=decision",
            "AUDIT C-1 — verdict: PASS",
            "ACK agent-1",
            "PLAN C-1",
            "감사 요청 tier 2",
        ] {
            assert_eq!(marker_of(&format!("summary\n{word}")), None, "{word}");
        }
    }

    /// The first 1000 characters of a note are read, counted in characters (Korean too): a
    /// verdict that ends at the cut still reads and one character later does not.
    #[test]
    fn the_note_head_is_cut_at_exactly_one_thousand_characters() {
        let line = |pad: usize| {
            let mut s = String::from("AUDIT C-1 head abc ");
            s.push_str(&"감".repeat(pad));
            s.push_str(" verdict: PASS");
            s
        };
        let base = line(0).chars().count();
        for (pad, stage) in [
            (1000 - base, Stage::Audited),
            (1000 - base + 1, Stage::Stale),
        ] {
            let f = fx();
            f.card("active", &[("report", &line(pad))]);
            let flow = f.flow(0, &Whose::default());
            let expected = if stage == Stage::Audited {
                (1, 0)
            } else {
                (0, 1)
            };
            assert_eq!(
                (flow.audited, flow.stale),
                expected,
                "pad {pad}: first line of {} characters",
                line(pad).chars().count()
            );
        }
    }

    /// The newest marker wins, across kinds: every ordering of two notes reads as the newest.
    #[test]
    fn the_newest_marker_wins_over_an_older_one_of_another_kind() {
        let ready = "READY (head abc)";
        let blocked = "LANE-BLOCKED kind=decision";
        let pass = "AUDIT C-1 head abc — verdict: PASS";
        let ack = "ACK agent-1 C-1";
        for (notes, submitted, audited, blocked_n, stale) in [
            (vec![ready, blocked], 0, 0, 1, 0),
            (vec![blocked, ready], 1, 0, 0, 0),
            (vec![blocked, ack], 0, 0, 0, 1),
            (vec![ready, pass], 0, 1, 0, 0),
            (vec![pass, ready], 1, 0, 0, 0),
            (vec![pass, ack], 0, 0, 0, 1),
            // A note that is no marker, newer than the marker, changes nothing.
            (
                vec![blocked, "RESUME head abc", "waiting: a run"],
                0,
                0,
                1,
                0,
            ),
        ] {
            let f = fx();
            let all: Vec<(&str, &str)> = notes.iter().map(|n| ("report", *n)).collect();
            f.card("active", &all);
            let flow = f.flow(0, &Whose::default());
            assert_eq!(
                (flow.submitted, flow.audited, flow.blocked, flow.stale),
                (submitted, audited, blocked_n, stale),
                "{notes:?}"
            );
        }
    }

    /// Several markers on one card count that card once.
    #[test]
    fn several_markers_on_one_card_count_it_once() {
        let f = fx();
        f.card(
            "active",
            &[
                ("report", "READY (a)"),
                ("report", "READY (b)"),
                ("report", "MERGE-READY (c)"),
            ],
        );
        let flow = f.flow(0, &Whose::default());
        assert_eq!((flow.submitted, flow.draft, flow.stale), (1, 0, 0));
    }
}

/// AC-3 of SPEC-20261122-statusline-flow: done today changes at the UTC midnight.
#[cfg(test)]
mod done_today_tests {
    use super::flow_support::fx;
    use super::*;

    /// 20000 days after the epoch, at 00:00:00 UTC.
    const MIDNIGHT: i64 = 20_000 * 86_400;

    #[test]
    fn the_count_changes_at_the_utc_midnight_and_not_a_second_before() {
        let f = fx();
        f.done_at(MIDNIGHT - 1);
        f.done_at(MIDNIGHT);
        f.done_at(MIDNIGHT + 3_600);
        let who = Whose::default();
        // The last second of the day before: that day started a day earlier, so all three
        // (closed at or after its start) count.
        assert_eq!(f.flow(MIDNIGHT - 1, &who).done_today, 3);
        assert_eq!(f.flow(MIDNIGHT - 2, &who).done_today, 3);
        // At the midnight the new day starts: the card closed one second earlier is yesterday's.
        assert_eq!(f.flow(MIDNIGHT, &who).done_today, 2);
        assert_eq!(f.flow(MIDNIGHT + 86_399, &who).done_today, 2);
        // The next midnight: all three are older.
        assert_eq!(f.flow(MIDNIGHT + 86_400, &who).done_today, 0);
    }

    #[test]
    fn open_cards_never_count_as_done_today() {
        let f = fx();
        f.card("active", &[]);
        f.card("hold", &[]);
        f.done_at(MIDNIGHT + 5);
        assert_eq!(f.flow(MIDNIGHT + 10, &Whose::default()).done_today, 1);
    }

    /// SPEC-20261124-statusline-counts-truth REQ-9: in a zone `o` seconds east of UTC the day
    /// starts at `MIDNIGHT - o`; the count changes there and not a second before, in zones west
    /// of UTC, at UTC and east of it, and with a half-hour offset.
    #[test]
    fn the_count_changes_at_the_local_midnight_and_not_a_second_before() {
        for offset in [-8 * 3_600, 0, 9 * 3_600, 19_800] {
            let zone = DayZone::Offset(offset);
            let local_midnight = MIDNIGHT - offset;
            let f = fx();
            f.done_at(local_midnight - 1);
            f.done_at(local_midnight);
            f.done_at(local_midnight + 3_600);
            let who = Whose::default();
            let at = |now| f.flow_in(now, &who, zone).done_today;
            assert_eq!(at(local_midnight - 1), 3, "offset {offset}");
            assert_eq!(at(local_midnight), 2, "offset {offset}");
            assert_eq!(at(local_midnight + 86_399), 2, "offset {offset}");
            assert_eq!(at(local_midnight + 86_400), 0, "offset {offset}");
        }
    }

    /// The same ledger at the same instant counts differently in two zones: at 15:00Z Seoul
    /// starts a new day and UTC does not.
    #[test]
    fn two_zones_count_the_same_ledger_differently() {
        let f = fx();
        let seoul_midnight = MIDNIGHT - 9 * 3_600;
        f.done_at(seoul_midnight - 1);
        f.done_at(seoul_midnight);
        let who = Whose::default();
        assert_eq!(f.flow(seoul_midnight, &who).done_today, 2);
        assert_eq!(
            f.flow_in(seoul_midnight, &who, DayZone::Offset(9 * 3_600))
                .done_today,
            1
        );
    }

    /// REQ-10: a zone that cannot be read is the UTC day, and a readable one is its own day.
    #[test]
    fn an_unreadable_zone_is_the_utc_day() {
        assert_eq!(day_start_with(MIDNIGHT + 99, |_| None), MIDNIGHT);
        assert_eq!(day_start_with(MIDNIGHT - 1, |_| None), MIDNIGHT - 86_400);
        assert_eq!(day_start_with(MIDNIGHT + 99, |_| Some(7)), 7);
    }
}

/// SPEC-20261124-statusline-counts-truth AC-4 and AC-5: the stage of a card is read from the
/// notes and the landing at or after its last reservation, also for a ready card a lane released.
#[cfg(test)]
mod stage_tests_after_reserve {
    use super::flow_support::{fx, Fx};
    use super::*;

    /// The render clock, and the second a lane takes the card.
    const NOW: i64 = 1_800_000_000;
    const TAKE: i64 = 1_799_990_000;
    const LIVE: i64 = NOW + 60;

    const SUBMIT: [&str; 3] = [
        "LANE-READY-INTEGRATION C-1 head=abc",
        "REPAIRED C-1 head=abc",
        "READY C-1 head=abc",
    ];
    const PASS: &str = "AUDIT C-1 (scope) head abc base def — verdict: PASS";

    fn flow(f: &Fx) -> Flow {
        f.flow(NOW, &Whose::default())
    }

    /// A ready card a lane took at `TAKE`, handed back with `body` at `at` and released.
    fn released(f: &Fx, body: &str, at: i64) -> i64 {
        let card = f.card("ready", &[]);
        f.reserve_gen(card, 1, "released", TAKE, TAKE + 60);
        f.note_at(card, "report", "lane", body, at);
        card
    }

    #[test]
    fn a_ready_card_reads_a_submit_marker_at_or_after_its_last_take() {
        for body in SUBMIT {
            let f = fx();
            released(&f, body, TAKE - 1);
            released(&f, body, TAKE);
            released(&f, body, TAKE + 1);
            let flow = flow(&f);
            assert_eq!(flow.submitted, 2, "{body}");
            assert_eq!(flow.ready, 1, "{body}");
        }
    }

    #[test]
    fn a_ready_card_reads_an_audit_pass_at_or_after_its_last_take() {
        let f = fx();
        released(&f, PASS, TAKE - 1);
        released(&f, PASS, TAKE);
        released(&f, PASS, TAKE + 1);
        let flow = flow(&f);
        assert_eq!(flow.audited, 2);
        assert_eq!(flow.ready, 1);
    }

    #[test]
    fn a_ready_card_reads_only_the_hand_back_markers() {
        let f = fx();
        released(&f, "LANE-BLOCKED kind=decision C-1 which", TAKE + 1);
        released(&f, "ACK C-1 taking it", TAKE + 1);
        released(&f, "RESUME later", TAKE + 1);
        // A card nobody ever took has no round of work to read notes of.
        let never = f.card("ready", &[]);
        f.note_at(never, "report", "lane", SUBMIT[0], TAKE + 1);
        let flow = flow(&f);
        assert_eq!(
            (flow.ready, flow.submitted, flow.audited, flow.blocked),
            (4, 0, 0, 0)
        );
    }

    #[test]
    fn the_newest_marker_after_the_take_decides_a_ready_card() {
        let f = fx();
        let card = released(&f, SUBMIT[0], TAKE + 1);
        f.note_at(card, "report", "auditor", PASS, TAKE + 2);
        let flow = flow(&f);
        assert_eq!((flow.audited, flow.submitted, flow.ready), (1, 0, 0));
    }

    #[test]
    fn a_card_taken_again_does_not_keep_the_hand_back_of_the_round_before() {
        let f = fx();
        // Handed back, failed the audit, taken again: working until the lane says otherwise.
        let card = f.card("active", &[]);
        f.reserve_gen(card, 1, "released", TAKE - 5_000, TAKE - 4_000);
        f.note_at(card, "report", "lane", SUBMIT[0], TAKE - 3_000);
        f.reserve_gen(card, 2, "held", TAKE, LIVE);
        let flow = flow(&f);
        assert_eq!((flow.working, flow.submitted), (1, 0));
        // The same card with an ended lease is stale, not submitted.
        let g = fx();
        let card = g.card("active", &[]);
        g.reserve_gen(card, 1, "released", TAKE - 5_000, TAKE - 4_000);
        g.note_at(card, "report", "lane", SUBMIT[0], TAKE - 3_000);
        g.reserve_gen(card, 2, "held", TAKE, NOW - 1);
        let flow = g.flow(NOW, &Whose::default());
        assert_eq!((flow.stale, flow.submitted), (1, 0));
    }

    #[test]
    fn repaired_is_a_submit_marker_of_a_working_card() {
        let f = fx();
        let card = f.card("active", &[]);
        f.reserve_gen(card, 1, "held", TAKE, LIVE);
        f.note_at(card, "report", "lane", "REPAIRED C-1 head=abc", TAKE + 1);
        let flow = flow(&f);
        assert_eq!((flow.submitted, flow.working), (1, 0));
    }

    #[test]
    fn a_block_ends_when_the_leader_answers_it() {
        let f = fx();
        let blocked = "LANE-BLOCKED kind=decision C-1 which option";
        let answer = "LEADER 결정: 둘째 안";
        let card = |notes: &[(&str, &str, &str)]| {
            let id = f.card("active", &[]);
            f.reserve_gen(id, 1, "held", TAKE, LIVE);
            // One second for every note: the order is the order of the note ids.
            for (kind, author, body) in notes {
                f.note_at(id, kind, author, body, TAKE + 10);
            }
        };
        // answered by the leader: the lane works on
        card(&[("report", "lane", blocked), ("decision", "leader", answer)]);
        // answered by a lane, or by a leader's report: still blocked
        card(&[("report", "lane", blocked), ("decision", "lane", answer)]);
        card(&[("report", "lane", blocked), ("report", "leader", answer)]);
        // the answer came first: blocked
        card(&[("decision", "leader", answer), ("report", "lane", blocked)]);
        // handed back, blocked, answered: the older marker stands
        card(&[
            ("report", "lane", SUBMIT[0]),
            ("report", "lane", blocked),
            ("decision", "leader", answer),
        ]);
        let flow = flow(&f);
        assert_eq!(
            (flow.working, flow.blocked, flow.submitted),
            (1, 3, 1),
            "{flow:?}"
        );
    }

    #[test]
    fn a_landing_counts_only_at_or_after_the_last_take() {
        let f = fx();
        for landed_at in [TAKE - 5, TAKE, TAKE + 5] {
            let card = f.card("review", &[]);
            f.reserve_gen(card, 1, "held_for_review", TAKE, LIVE);
            f.event_at(card, "card.landed", landed_at);
        }
        let flow = flow(&f);
        // The card that landed before its take is a review card like any other: submitted.
        assert_eq!((flow.landed, flow.submitted), (2, 1));
    }

    #[test]
    fn a_withdrawn_hold_card_is_not_counted_as_waiting_for_the_user() {
        let f = fx();
        let withdrawn = f.card("hold", &[]);
        f.event_at(withdrawn, "card.withdraw", TAKE);
        let folded = f.card("hold", &[]);
        f.event_at(folded, "card.fold", TAKE);
        f.card("hold", &[]);
        let flow = flow(&f);
        assert_eq!(flow.hold, 1);
        // SPEC-20261124-card-flow-icons REQ-4: what was withdrawn or folded is counted as
        // paused, so the two counts together are every hold card.
        assert_eq!(flow.paused, 2);
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

/// SPEC-20261122-statusline-where AC-3: the worktree-name rule, the integration distance and
/// the readings behind them, as pure functions. `0123456` stands for a short commit and `next`
/// for the integration branch of the fixtures.
#[cfg(test)]
mod where_tests {
    use super::*;

    const FULL_SHA: &str = "0123456789abcdef0123456789abcdef01234567";

    fn at<'a>(
        worktree: &'a str,
        head: Head<'a>,
        integration: Option<&'a str>,
        numbers: (i64, i64, i64),
    ) -> Where {
        where_items(&WhereInput {
            worktree,
            head,
            integration,
            ahead: numbers.0,
            behind: numbers.1,
            dirty: numbers.2,
        })
    }

    fn marks(ahead: i64, behind: i64, dirty: i64) -> Marks {
        Marks {
            ahead,
            behind,
            dirty,
        }
    }

    fn distance(name: &str, ahead: i64, behind: i64) -> Option<Distance> {
        Some(Distance {
            name: name.into(),
            ahead,
            behind,
        })
    }

    /// REQ-3: the same name shows the worktree item alone, and the marks and the distance stay.
    #[test]
    fn a_worktree_named_like_its_branch_shows_the_worktree_item_alone() {
        let got = at(
            "statusline-flow",
            Head::Branch("statusline-flow"),
            Some("next"),
            (3, 0, 2),
        );
        assert_eq!(
            got,
            Where {
                worktree: Some("statusline-flow".into()),
                branch: None,
                marks: marks(3, 0, 2),
                distance: distance("next", 3, 0),
            }
        );
    }

    /// REQ-3: another name shows both items, the marks follow the branch, and the behind count
    /// lives in the distance only.
    #[test]
    fn unequal_names_show_the_worktree_item_and_the_branch_item() {
        let got = at(
            "c135",
            Head::Branch("spec-impl/20261122-statusline-where"),
            Some("next"),
            (1, 2, 0),
        );
        assert_eq!(
            got,
            Where {
                worktree: Some("c135".into()),
                branch: Some("spec-impl/20261122-statusline-where".into()),
                marks: marks(1, 0, 0),
                distance: distance("next", 1, 2),
            }
        );
    }

    /// The names are compared as they are: case and a longer name are two names.
    #[test]
    fn the_names_are_compared_exactly() {
        for branch in [
            "Statusline-Flow",
            "statusline-flow/x",
            "statusline-flo",
            "x",
        ] {
            let got = at(
                "statusline-flow",
                Head::Branch(branch),
                Some("next"),
                (0, 0, 0),
            );
            assert_eq!(got.worktree.as_deref(), Some("statusline-flow"), "{branch}");
            assert_eq!(got.branch.as_deref(), Some(branch), "{branch}");
        }
    }

    /// REQ-3: the main checkout (no worktree name) shows the branch item alone.
    #[test]
    fn the_main_checkout_shows_the_branch_item_alone() {
        let got = at("", Head::Branch("feature/login"), Some("next"), (1, 2, 3));
        assert_eq!(
            got,
            Where {
                worktree: None,
                branch: Some("feature/login".into()),
                marks: marks(1, 0, 3),
                distance: distance("next", 1, 2),
            }
        );
    }

    /// REQ-4: on the integration branch itself there is no distance and `↑` is the landings not
    /// yet pushed; in a worktree of that name the worktree item stands alone.
    #[test]
    fn on_the_integration_branch_the_ahead_count_is_the_landings_not_yet_pushed() {
        let got = at("", Head::Branch("next"), Some("next"), (2, 5, 1));
        assert_eq!(
            got,
            Where {
                worktree: None,
                branch: Some("next".into()),
                marks: marks(2, 0, 1),
                distance: None,
            }
        );
        let got = at("next", Head::Branch("next"), Some("next"), (2, 5, 1));
        assert_eq!(got.worktree.as_deref(), Some("next"));
        assert_eq!(got.branch, None);
        assert_eq!(got.distance, None);
        assert_eq!(got.marks, marks(2, 0, 1));
    }

    /// REQ-4: with no integration branch known the item is left out and the branch item keeps
    /// the numbers git gave against its own upstream, behind included.
    #[test]
    fn an_unknown_integration_branch_leaves_the_item_out() {
        let got = at("", Head::Branch("feature/login"), None, (2, 1, 4));
        assert_eq!(
            got,
            Where {
                worktree: None,
                branch: Some("feature/login".into()),
                marks: marks(2, 1, 4),
                distance: None,
            }
        );
        let got = at("lane", Head::Branch("other"), None, (2, 1, 4));
        assert_eq!(got.marks, marks(2, 1, 4));
        assert_eq!(got.distance, None);
    }

    /// REQ-2, REQ-4: a detached HEAD shows its short commit in the branch item, has no distance
    /// and no ahead or behind count whatever the input says, and keeps the dirty count.
    #[test]
    fn a_detached_head_shows_its_short_commit_and_no_distance() {
        let got = at("", Head::Detached("0123456"), Some("next"), (4, 9, 3));
        assert_eq!(
            got,
            Where {
                worktree: None,
                branch: Some("0123456".into()),
                marks: marks(0, 0, 3),
                distance: None,
            }
        );
        let got = at("c135", Head::Detached("0123456"), Some("next"), (0, 0, 0));
        assert_eq!(got.worktree.as_deref(), Some("c135"));
        assert_eq!(got.branch.as_deref(), Some("0123456"));
        assert_eq!(got.distance, None);
    }

    /// REQ-2: an unreadable `HEAD` leaves the branch item and the distance out.
    #[test]
    fn an_unreadable_head_leaves_the_branch_items_out() {
        let got = at("", Head::Unreadable, Some("next"), (3, 1, 2));
        assert_eq!(got.branch, None);
        assert_eq!(got.distance, None);
        assert_eq!(got.worktree, None);
        // The worktree name comes from the host: its item stays, without marks.
        let got = at("c135", Head::Unreadable, Some("next"), (3, 1, 2));
        assert_eq!(
            got,
            Where {
                worktree: Some("c135".into()),
                branch: None,
                marks: Marks::default(),
                distance: None,
            }
        );
    }

    /// REQ-4: the distance writes both numbers, a zero included.
    #[test]
    fn the_distance_keeps_both_numbers_at_zero() {
        let got = at("", Head::Branch("lane"), Some("next"), (0, 0, 0));
        assert_eq!(got.distance, distance("next", 0, 0));
        assert_eq!(got.marks, marks(0, 0, 0));
    }

    /// REQ-2: the `HEAD` file read as a branch, a short commit or nothing, at the edges of the
    /// commit length (a 40 and a 64 digit hash are commits, 39 and 41 are not).
    #[test]
    fn the_head_file_reads_as_a_branch_a_commit_or_nothing() {
        assert_eq!(head_of("ref: refs/heads/main\n"), Head::Branch("main"));
        assert_eq!(
            head_of("ref: refs/heads/feature/login"),
            Head::Branch("feature/login")
        );
        assert_eq!(
            head_of("ref: refs/heads/v2026.11.0\n"),
            Head::Branch("v2026.11.0")
        );
        assert_eq!(head_of(&format!("{FULL_SHA}\n")), Head::Detached("0123456"));
        assert_eq!(head_of(FULL_SHA), Head::Detached("0123456"));
        let long = FULL_SHA.repeat(2)[..64].to_string();
        assert_eq!(head_of(&long), Head::Detached("0123456"), "64 digits");
        assert_eq!(head_of(&FULL_SHA[..39]), Head::Unreadable, "39 digits");
        assert_eq!(
            head_of(&format!("{FULL_SHA}0")),
            Head::Unreadable,
            "41 digits"
        );
        for nothing in [
            "",
            "\n",
            "garbage",
            "ref: refs/heads/",
            "ref: refs/heads/\n",
            "ref: refs/tags/v1\n",
            "ref: refs/remotes/origin/main\n",
            "ref: refs/heads/a\u{1b}[31mb\n",
            "ref: refs/heads/a\nb\n",
            "xyz0123456789abcdef0123456789abcdef0123456",
        ] {
            assert_eq!(head_of(nothing), Head::Unreadable, "{nothing:?}");
        }
    }

    const PROFILE: &str = r#"{"schema_version":2,"project_id":"p","work_flow":{"flow":"m-flow","integration_branch":"v2026.11.0"}}"#;

    /// Risk 3 of the SPEC: the profile names the integration branch and else the remote's
    /// default branch; a repository with neither shows no item.
    #[test]
    fn the_integration_branch_comes_from_the_profile_then_the_remote_default() {
        let origin = "ref: refs/remotes/origin/main\n";
        assert_eq!(
            integration_from(Some(PROFILE), Some(origin)).as_deref(),
            Some("v2026.11.0")
        );
        assert_eq!(
            integration_from(Some(PROFILE), None).as_deref(),
            Some("v2026.11.0")
        );
        assert_eq!(
            integration_from(None, Some(origin)).as_deref(),
            Some("main")
        );
        assert_eq!(
            integration_from(None, Some("ref: refs/remotes/origin/release/1.2\n")).as_deref(),
            Some("release/1.2")
        );
        for profile in [
            r#"{"work_flow":{"flow":"trunk-based"}}"#,
            r#"{"work_flow":{"flow":"m-flow","integration_branch":null}}"#,
            r#"{"work_flow":null}"#,
            r#"{}"#,
            "not json",
            "",
        ] {
            assert_eq!(
                integration_from(Some(profile), Some(origin)).as_deref(),
                Some("main"),
                "{profile:?} falls to the remote default"
            );
            assert_eq!(integration_from(Some(profile), None), None, "{profile:?}");
        }
        assert_eq!(integration_from(None, None), None);
        for origin in [
            "",
            "garbage",
            "ref: refs/remotes/origin/\n",
            "ref: refs/heads/main\n",
            "0123456789abcdef0123456789abcdef01234567\n",
        ] {
            assert_eq!(integration_from(None, Some(origin)), None, "{origin:?}");
        }
    }

    /// A profile that names a branch git cannot be given safely names none.
    #[test]
    fn an_unsafe_integration_name_is_no_name() {
        for name in ["-x", "--upload-pack=x", "a b", "a..b", "HEAD", ""] {
            let profile = format!(
                r#"{{"work_flow":{{"flow":"m-flow","integration_branch":{}}}}}"#,
                serde_json::Value::String(name.into())
            );
            assert_eq!(integration_from(Some(&profile), None), None, "{name:?}");
        }
    }

    fn upstream_status_args() -> Vec<String> {
        ["--no-optional-locks", "status", "--porcelain", "--branch"]
            .map(String::from)
            .to_vec()
    }

    fn lane_status_args() -> Vec<String> {
        ["--no-optional-locks", "status", "--porcelain"]
            .map(String::from)
            .to_vec()
    }

    /// REQ-1, REQ-3, REQ-4: a lane branch is measured by the second call, the `status` call has
    /// no override of a `branch.<name>.*` key, and every other case keeps git's own numbers.
    #[test]
    fn a_lane_branch_is_measured_against_the_integration_branch_by_its_own_call() {
        for (branch, integration) in [
            ("lane", "next"),
            ("spec-impl/20261122-statusline-where", "v2026.11.0"),
        ] {
            let m = measure(Some(branch), Some(integration), true);
            assert_eq!(m, Measure::Integration(integration));
            let args = status_args(&m);
            assert_eq!(args, lane_status_args());
            assert!(
                !args.iter().any(|a| a == "-c" || a.contains("branch.")),
                "{args:?}"
            );
            assert_eq!(
                distance_args(integration),
                vec![
                    "rev-list".to_string(),
                    "--left-right".into(),
                    "--count".into(),
                    format!("refs/heads/{integration}...HEAD"),
                    "--".into(),
                ]
            );
        }
        // Without a local ref of the integration branch: nothing to measure, no second call.
        let missing = measure(Some("lane"), Some("next"), false);
        assert_eq!(missing, Measure::Missing);
        assert_eq!(status_args(&missing), lane_status_args());
        // On the integration branch, with no integration branch, detached or unreadable: git's
        // own numbers.
        for m in [
            measure(Some("next"), Some("next"), true),
            measure(Some("lane"), None, true),
            measure(None, Some("next"), true),
            measure(None, None, false),
        ] {
            assert_eq!(m, Measure::Upstream);
            assert_eq!(status_args(&m), upstream_status_args());
        }
    }

    /// The answer of the second call is `<behind>\t<ahead>`; anything else is no answer.
    #[test]
    fn the_distance_answer_is_two_numbers_or_nothing() {
        assert_eq!(parse_distance("1\t2\n"), Some((1, 2)));
        assert_eq!(parse_distance("0\t0\n"), Some((0, 0)));
        assert_eq!(parse_distance("10\t200"), Some((10, 200)));
        for bad in [
            "",
            "\n",
            "1\n",
            "1\t",
            "\t2",
            "x\ty",
            "1\t2\t3",
            "-1\t2",
            "+1\t2",
            "1 2",
            "1\t2\n\n",
            "99999999999999999999\t1",
        ] {
            assert_eq!(parse_distance(bad), None, "{bad:?}");
        }
    }

    /// A name that git could take as an option, as a revision range or as a config value is
    /// never handed to it; a name of exactly 100 bytes is, and one of 101 is not.
    #[test]
    fn a_name_git_could_misread_is_never_passed() {
        for bad in [
            "",
            "-x",
            "--upload-pack=x",
            "a b",
            "a\tb",
            "a..b",
            "a=b",
            "HEAD",
            "é",
            "a\nb",
            "-c core.pager=x",
        ] {
            assert_eq!(
                measure(Some(bad), Some("next"), true),
                Measure::Upstream,
                "{bad:?}"
            );
            assert_eq!(
                measure(Some("lane"), Some(bad), true),
                Measure::Upstream,
                "{bad:?}"
            );
        }
        let hundred = "a".repeat(100);
        assert_eq!(
            measure(Some(&hundred), Some("next"), true),
            Measure::Integration("next"),
            "100 bytes"
        );
        assert_eq!(
            measure(Some("lane"), Some(&hundred), true),
            Measure::Integration(&hundred),
            "100 bytes"
        );
        let over = "a".repeat(101);
        assert_eq!(
            measure(Some(&over), Some("next"), true),
            Measure::Upstream,
            "101 bytes"
        );
        assert_eq!(
            measure(Some("lane"), Some(&over), true),
            Measure::Upstream,
            "101 bytes"
        );
    }

    fn collect_pr(pr: serde_json::Value) -> Option<PrInfo> {
        let stdin = serde_json::json!({ "pr": pr }).to_string();
        let input = input::parse(stdin.as_bytes());
        collect(Some(&input), &Context::for_test(&[], None)).pr
    }

    /// REQ-9: a merge request is the one the stdin says is one, `pr.kind` of `mr` as the
    /// official statusline page documents it; anything else, including a value of another type,
    /// is a pull request and never fails the payload.
    #[test]
    fn a_merge_request_is_the_one_the_stdin_kind_says() {
        let pr = collect_pr(serde_json::json!({
            "number": 12, "review_state": "pending", "kind": "mr"
        }))
        .expect("a pr");
        assert!(pr.merge_request, "kind mr");
        assert_eq!((pr.number, pr.review_state.as_str()), (12, "pending"));
        for kind in [
            serde_json::json!("pull_request"),
            serde_json::json!("merge_request"),
            serde_json::json!("MR"),
            serde_json::json!(""),
            serde_json::json!(null),
            serde_json::json!(7),
            serde_json::json!({"a": 1}),
        ] {
            let pr = collect_pr(serde_json::json!({"number": 12, "kind": kind}))
                .unwrap_or_else(|| panic!("kind {kind} fails the payload"));
            assert!(!pr.merge_request, "kind {kind}");
        }
        let pr = collect_pr(serde_json::json!({"number": 12})).expect("a pr without a kind");
        assert!(!pr.merge_request);
    }
}

/// SPEC-20261122-statusline-where REQ-5 and REQ-6: the name of a session's cache file and the
/// rule that lets it stand in for git.
#[cfg(test)]
mod cache_key_tests {
    use super::*;

    #[test]
    fn the_age_limit_is_ten_seconds() {
        assert_eq!(CACHE_MAX_AGE_SECS, 10);
        assert_eq!(CACHE_KEEP_SECS, 86_400);
        assert_eq!(CACHE_REMOVE_PER_RENDER, 10);
    }

    /// Only the session id names the file, and only an id of 1 to 64 plain characters.
    #[test]
    fn only_a_plain_session_id_of_one_to_sixty_four_characters_names_a_file() {
        assert_eq!(
            cache_file_name("abc-DEF_123").as_deref(),
            Some("abc-DEF_123.json")
        );
        assert_eq!(cache_file_name("a").as_deref(), Some("a.json"));
        let n64 = "a".repeat(64);
        assert_eq!(cache_file_name(&n64), Some(format!("{n64}.json")), "64");
        for bad in [
            "".to_string(),
            "a".repeat(65),
            "a/b".into(),
            "../x".into(),
            "..".into(),
            "a.b".into(),
            "a b".into(),
            "a\0b".into(),
            "a\nb".into(),
            "é".into(),
            "a\\b".into(),
            "~".into(),
        ] {
            assert_eq!(cache_file_name(&bad), None, "{bad:?}");
        }
    }

    /// REQ-5: the cache goes under an absolute `TMPDIR`, under `/tmp` when it is unset or
    /// empty, and nowhere when it is relative.
    #[test]
    fn the_cache_base_refuses_a_relative_tmpdir() {
        assert_eq!(cache_base(None), Some(PathBuf::from("/tmp")));
        assert_eq!(cache_base(Some("")), Some(PathBuf::from("/tmp")));
        assert_eq!(
            cache_base(Some("/var/folders/x")),
            Some(PathBuf::from("/var/folders/x"))
        );
        for relative in [".", "./", "tmp", "./tmp", "../tmp"] {
            assert_eq!(cache_base(Some(relative)), None, "{relative:?}");
            assert_eq!(cache_folder(Some(relative)), None, "{relative:?}");
        }
    }

    /// In the shared `/tmp` (TMPDIR unset or empty) the cache folder's name carries this user's
    /// id, so a folder another user made first under the plain name is never this user's; with
    /// a TMPDIR of the user's own the plain name stays.
    #[cfg(unix)]
    #[test]
    fn the_shared_tmp_cache_folder_is_named_per_user() {
        // SAFETY: geteuid has no preconditions and cannot fail.
        let me = unsafe { libc::geteuid() };
        let own = PathBuf::from(format!("/tmp/moai-statusline-{me}"));
        assert_eq!(cache_folder(None), Some(own.clone()));
        assert_eq!(cache_folder(Some("")), Some(own));
        assert_eq!(
            cache_folder(Some("/var/folders/x")),
            Some(PathBuf::from("/var/folders/x/moai-statusline"))
        );
    }

    /// The reader's guard: a folder that others may enter (group or other rights), a link to a
    /// private folder and a file are all refused; only a 0700 folder of this user passes.
    #[cfg(unix)]
    #[test]
    fn the_reader_accepts_only_a_private_folder() {
        use std::os::unix::fs::PermissionsExt;
        let t = tempfile::tempdir().unwrap();
        let dir = t.path().join("c");
        std::fs::create_dir(&dir).unwrap();
        for (mode, want) in [
            (0o700, true),
            (0o750, false),
            (0o705, false),
            (0o770, false),
        ] {
            std::fs::set_permissions(&dir, std::fs::Permissions::from_mode(mode)).unwrap();
            assert_eq!(is_private_folder(&dir), want, "{mode:o}");
        }
        std::fs::set_permissions(&dir, std::fs::Permissions::from_mode(0o700)).unwrap();
        let link = t.path().join("link");
        std::os::unix::fs::symlink(&dir, &link).unwrap();
        assert!(!is_private_folder(&link));
        let file = t.path().join("f");
        std::fs::write(&file, "").unwrap();
        assert!(!is_private_folder(&file));
        assert!(!is_private_folder(&t.path().join("missing")));
    }

    /// REQ-6 at its edges: an age of 9 seconds is fresh, 10 and 11 are not; the same `HEAD`
    /// content and the same integration commit are both needed.
    #[test]
    fn a_cache_is_fresh_below_ten_seconds_with_the_same_head_and_integration_commit() {
        let fresh = |age: i64, cached_head: &str, head: &str, cached_i: &str, i: &str| {
            cache_is_fresh(1_000, 1_000 + age, cached_head, head, cached_i, i)
        };
        let (h, i) = ("ref: refs/heads/main\n", "0123456789abcdef");
        assert!(fresh(0, h, h, i, i), "age 0");
        assert!(fresh(9, h, h, i, i), "age 9");
        assert!(!fresh(10, h, h, i, i), "age 10");
        assert!(!fresh(11, h, h, i, i), "age 11");
        assert!(!fresh(3_600, h, h, i, i), "an hour");
        assert!(
            !fresh(1, h, "ref: refs/heads/other\n", i, i),
            "another HEAD"
        );
        assert!(!fresh(1, h, h, i, "fedcba9876543210"), "another commit");
        assert!(
            !fresh(1, h, h, "", i),
            "a cache without an integration commit"
        );
        assert!(
            fresh(1, h, h, "", ""),
            "no integration branch on both sides"
        );
    }
}
