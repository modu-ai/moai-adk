//! Issue #2588: the forge counts cache (`⬇️ issues / 🔀 PRs`) and its refresh.
//!
//! The render path stays a file read: it reads the cache and, when the cache is missing
//! or older than [`TTL`], starts one detached `moai-cli statusline --refresh-counts`
//! child and returns without waiting. The child asks `gh` for the counts within
//! [`FETCH_BUDGET`] and rewrites the cache. An exclusive lock file lets only one child
//! run per checkout however many sessions render at once.
//!
//! The approach follows the Go generation's stale-while-revalidate refresh
//! (`internal/statusline/github.go`, revision 4fa2aa9db2), adapted rather than ported:
//! the cache moves to `.moai-cli/state/github/` (the Go path `.moai/state/github/` is
//! still read, never written), the stampede guard is a lock file instead of an early
//! timestamp write, and a refresh starts only for a checkout whose `origin` remote is on
//! github.com, decided by reading `.git/config` (no subprocess on the render path).

use serde::{Deserialize, Serialize};
use std::fs::{self, OpenOptions};
use std::io::Write;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use std::time::{Duration, Instant, SystemTime, UNIX_EPOCH};

/// A cache younger than this is served without a refresh.
pub const TTL: Duration = Duration::from_secs(10 * 60);
/// A count fetched longer ago than this is hidden instead of shown as current.
pub const STALE: Duration = Duration::from_secs(60 * 60);
/// A refresh lock older than this belongs to a child that died; it is taken over.
pub const LOCK_TTL: Duration = Duration::from_secs(60);
/// Upper bound for each `gh` call in the refresh child.
pub const FETCH_BUDGET: Duration = Duration::from_secs(20);

/// The cache file, as written by this tool and by the Go generation.
#[derive(Debug, Default, Clone, Serialize, Deserialize, PartialEq)]
#[serde(default)]
pub struct Cache {
    pub open_issues: Option<i64>,
    pub open_prs: Option<i64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub suppressed: Option<bool>,
    /// When the counts were last fetched successfully (unix seconds).
    pub fetched_at: Option<i64>,
    /// When a refresh last ran, successful or not (unix seconds). Keeps a failing `gh`
    /// from being retried on every render.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub attempted_at: Option<i64>,
}

pub fn now_unix() -> i64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map_or(0, |d| d.as_secs() as i64)
}

fn dir(anchor: &Path) -> PathBuf {
    crate::state_root::path(anchor, "state/github")
}

/// Where this tool writes the cache.
pub fn cache_path(anchor: &Path) -> PathBuf {
    dir(anchor).join("counts.json")
}

/// The Go generation's cache, read while it is the only one.
pub fn go_cache_path(anchor: &Path) -> PathBuf {
    anchor.join(".moai/state/github/counts.json")
}

fn lock_path(anchor: &Path) -> PathBuf {
    dir(anchor).join("refresh.lock")
}

fn parse(path: &Path) -> Option<Cache> {
    fs::read(path)
        .ok()
        .and_then(|b| serde_json::from_slice::<Cache>(&b).ok())
}

/// The cache: this tool's file first, else the Go file.
pub fn read(anchor: &Path) -> Option<Cache> {
    parse(&cache_path(anchor)).or_else(|| parse(&go_cache_path(anchor)))
}

/// True when the cache is missing or its last refresh attempt is older than [`TTL`].
pub fn needs_refresh(cache: Option<&Cache>, now: i64) -> bool {
    let Some(c) = cache else { return true };
    let last = c.fetched_at.max(c.attempted_at).unwrap_or(0);
    now.saturating_sub(last) >= TTL.as_secs() as i64
}

/// True when the counts were fetched within [`STALE`]. A missing time is stale; a time
/// ahead of the clock (skew) counts as fresh.
pub fn is_fresh(cache: &Cache, now: i64) -> bool {
    cache
        .fetched_at
        .is_some_and(|t| now.saturating_sub(t) < STALE.as_secs() as i64)
}

/// The `config` file of the repository `anchor` belongs to, found without a subprocess:
/// a `.git` folder holds it; a `.git` file (linked worktree, `--separate-git-dir`
/// checkout, submodule) names the git directory, whose `commondir`, when present, points
/// at the shared directory that holds the remotes.
fn git_config_path(anchor: &Path) -> Option<PathBuf> {
    let (top, dot_git) = super::state::find_dot_git(&fs::canonicalize(anchor).ok()?)?;
    if dot_git.is_dir() {
        return Some(dot_git.join("config"));
    }
    let text = fs::read_to_string(&dot_git).ok()?;
    let git_dir = PathBuf::from(text.trim().strip_prefix("gitdir:")?.trim());
    let git_dir = if git_dir.is_absolute() {
        git_dir
    } else {
        top.join(git_dir)
    };
    let common = match fs::read_to_string(git_dir.join("commondir")) {
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
    Some(common.join("config"))
}

/// A git config value as git reads it: double quotes removed, `\"` and `\\` unescaped,
/// and an unquoted `;` or `#` starting a comment.
fn config_value(raw: &str) -> String {
    let mut out = String::new();
    let mut quoted = false;
    let mut chars = raw.trim().chars();
    while let Some(c) = chars.next() {
        match c {
            '"' => quoted = !quoted,
            '\\' => out.extend(chars.next()),
            ';' | '#' if !quoted => break,
            _ => out.push(c),
        }
    }
    out.trim().to_string()
}

/// `owner/name` of a github.com URL in any of git's forms: `scheme://[user@]host[:port]/path`
/// or the scp-like `[user@]host:path`, without a trailing `.git` or slash.
fn github_owner_name(url: &str) -> Option<(String, String)> {
    let (host, path) = match url.split_once("://") {
        Some((_, rest)) => {
            let (authority, path) = rest.split_once('/')?;
            let host = authority.rsplit('@').next()?;
            // An explicit port belongs to the authority, never to the path.
            (host.split(':').next()?, path)
        }
        None => {
            let (user_host, path) = url.split_once(':')?;
            (user_host.rsplit('@').next()?, path)
        }
    };
    if !host.eq_ignore_ascii_case("github.com") {
        return None;
    }
    let path = path.trim_matches('/');
    let path = path.strip_suffix(".git").unwrap_or(path);
    let (owner, name) = path.split_once('/')?;
    (!owner.is_empty() && !name.is_empty() && !name.contains('/'))
        .then(|| (owner.to_string(), name.to_string()))
}

/// The repository of the `origin` remote of `anchor` when it is on github.com, read
/// from the git config file alone (no subprocess on the render path).
pub fn github_repo(anchor: &Path) -> Option<(String, String)> {
    let text = fs::read_to_string(git_config_path(anchor)?).ok()?;
    let mut in_origin = false;
    for line in text.lines() {
        let line = line.trim();
        if line.starts_with('[') {
            in_origin = line == "[remote \"origin\"]";
            continue;
        }
        if in_origin {
            if let Some((key, value)) = line.split_once('=') {
                if key.trim() == "url" {
                    return github_owner_name(&config_value(value));
                }
            }
        }
    }
    None
}

/// The `origin` remote of `anchor` is on github.com.
pub fn origin_is_github(anchor: &Path) -> bool {
    github_repo(anchor).is_some()
}

/// Take the refresh lock: create it exclusively, or take over one older than
/// [`LOCK_TTL`]. False when a live refresh holds it or the folder cannot be written.
fn take_lock(anchor: &Path, now: i64) -> bool {
    let path = lock_path(anchor);
    if fs::create_dir_all(dir(anchor)).is_err() {
        return false;
    }
    for _ in 0..2 {
        match OpenOptions::new().write(true).create_new(true).open(&path) {
            Ok(mut f) => {
                let _ = writeln!(f, "{} {}", std::process::id(), now);
                return true;
            }
            Err(e) if e.kind() == std::io::ErrorKind::AlreadyExists => {
                // A lock whose holder has not written its time yet (or wrote something
                // unreadable) is aged by its modification time, so a fresh one is live.
                let started = fs::read_to_string(&path)
                    .ok()
                    .and_then(|s| s.split_whitespace().nth(1)?.parse::<i64>().ok())
                    .or_else(|| {
                        let modified = fs::metadata(&path).ok()?.modified().ok()?;
                        Some(modified.duration_since(UNIX_EPOCH).ok()?.as_secs() as i64)
                    });
                // Gone between the two reads: try to create it again.
                let Some(started) = started else { continue };
                if now.saturating_sub(started) < LOCK_TTL.as_secs() as i64 {
                    return false;
                }
                let _ = fs::remove_file(&path);
            }
            Err(_) => return false,
        }
    }
    false
}

/// Called after a render: start one detached refresh child when the cache needs it.
/// Never waits and never fails the render. `exe` must be the `moai-cli` binary; any
/// other executable (a test harness) is not re-invoked.
pub fn maybe_refresh(anchor: &Path, exe: &Path, suppressed: bool) -> bool {
    let is_cli = exe
        .file_stem()
        .and_then(|s| s.to_str())
        .is_some_and(|s| s == "moai-cli");
    if suppressed || !is_cli || !origin_is_github(anchor) {
        return false;
    }
    let now = now_unix();
    if !needs_refresh(read(anchor).as_ref(), now) || !take_lock(anchor, now) {
        return false;
    }
    let mut cmd = Command::new(exe);
    cmd.args(["statusline", "--refresh-counts", "--board-root"])
        .arg(anchor)
        .current_dir(anchor)
        .stdin(Stdio::null())
        .stdout(Stdio::null())
        .stderr(Stdio::null());
    // Its own process group, so a signal to the host's render group does not stop it.
    #[cfg(unix)]
    std::os::unix::process::CommandExt::process_group(&mut cmd, 0);
    let spawned = cmd.spawn();
    if spawned.is_err() {
        let _ = fs::remove_file(lock_path(anchor));
        return false;
    }
    true
}

/// Exact open Issue and PR totals of one repository in a single GraphQL call.
const COUNTS_QUERY: &str = "query($owner:String!,$name:String!){repository(owner:$owner,name:$name){issues(states:OPEN){totalCount} pullRequests(states:OPEN){totalCount}}}";

/// The open Issue and PR totals of `owner/name` on github.com, bounded by
/// [`FETCH_BUDGET`]. The repository and host are passed explicitly and `GH_REPO` /
/// `GH_HOST` are removed, so an inherited environment cannot redirect the query.
/// `Err(true)` when `gh` is not installed, `Err(false)` for any other failure.
fn gh_counts(anchor: &Path, owner: &str, name: &str) -> Result<(i64, i64), bool> {
    let mut child = Command::new("gh")
        .args(["api", "graphql", "--hostname", "github.com", "-f"])
        .arg(format!("query={COUNTS_QUERY}"))
        .arg("-f")
        .arg(format!("owner={owner}"))
        .arg("-f")
        .arg(format!("name={name}"))
        .args([
            "--jq",
            r#".data.repository | "\(.issues.totalCount) \(.pullRequests.totalCount)""#,
        ])
        .env_remove("GH_REPO")
        .env_remove("GH_HOST")
        .current_dir(anchor)
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::null())
        .spawn()
        .map_err(|e| e.kind() == std::io::ErrorKind::NotFound)?;
    let deadline = Instant::now() + FETCH_BUDGET;
    loop {
        match child.try_wait() {
            Ok(Some(status)) => {
                let mut out = String::new();
                if let Some(mut s) = child.stdout.take() {
                    let _ = std::io::Read::read_to_string(&mut s, &mut out);
                }
                if !status.success() {
                    return Err(false);
                }
                let mut it = out.split_whitespace().map(str::parse::<i64>);
                return match (it.next(), it.next(), it.next()) {
                    (Some(Ok(issues)), Some(Ok(prs)), None) => Ok((issues, prs)),
                    _ => Err(false),
                };
            }
            Ok(None) if Instant::now() < deadline => std::thread::sleep(Duration::from_millis(50)),
            _ => {
                let _ = child.kill();
                let _ = child.wait();
                return Err(false);
            }
        }
    }
}

fn write(anchor: &Path, cache: &Cache) {
    let path = cache_path(anchor);
    let tmp = path.with_extension(format!("json.{}.tmp", std::process::id()));
    let Ok(bytes) = serde_json::to_vec(cache) else {
        return;
    };
    if fs::create_dir_all(dir(anchor)).is_ok()
        && fs::write(&tmp, bytes).is_ok()
        && fs::rename(&tmp, &path).is_err()
    {
        let _ = fs::remove_file(&tmp);
    }
}

/// The refresh child (`statusline --refresh-counts --board-root <anchor>`): fetch both
/// totals, rewrite this tool's cache and release the lock. A failure keeps the previous
/// counts with a new `attempted_at`; a missing `gh` marks the pair suppressed.
pub fn refresh(anchor: &Path) {
    let mut cache = read(anchor).unwrap_or_default();
    let now = now_unix();
    cache.attempted_at = Some(now);
    let fetched = github_repo(anchor).map(|(owner, name)| gh_counts(anchor, &owner, &name));
    match fetched {
        Some(Ok((issues, prs))) => {
            cache.open_issues = Some(issues);
            cache.open_prs = Some(prs);
            cache.fetched_at = Some(now);
            cache.suppressed = None;
        }
        Some(Err(true)) => cache.suppressed = Some(true),
        _ => {}
    }
    write(anchor, &cache);
    let _ = fs::remove_file(lock_path(anchor));
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn refresh_is_due_at_the_ttl_and_not_before() {
        let ttl = TTL.as_secs() as i64;
        let at = |fetched: Option<i64>, attempted: Option<i64>| Cache {
            fetched_at: fetched,
            attempted_at: attempted,
            ..Cache::default()
        };
        assert!(needs_refresh(None, 1_000_000));
        assert!(!needs_refresh(
            Some(&at(Some(1_000_000), None)),
            1_000_000 + ttl - 1
        ));
        assert!(needs_refresh(
            Some(&at(Some(1_000_000), None)),
            1_000_000 + ttl
        ));
        // A failed attempt counts: no retry before the TTL even though the fetch is old.
        assert!(!needs_refresh(
            Some(&at(Some(1), Some(1_000_000))),
            1_000_000 + ttl - 1
        ));
        assert!(needs_refresh(Some(&at(None, None)), 1_000_000));
    }

    #[test]
    fn counts_are_fresh_below_the_stale_age_only() {
        let stale = STALE.as_secs() as i64;
        let c = |t: Option<i64>| Cache {
            fetched_at: t,
            ..Cache::default()
        };
        assert!(is_fresh(&c(Some(5_000)), 5_000 + stale - 1));
        assert!(!is_fresh(&c(Some(5_000)), 5_000 + stale));
        assert!(!is_fresh(&c(None), 5_000));
    }

    #[test]
    fn only_a_github_origin_qualifies() {
        let tmp = tempfile::tempdir().unwrap();
        let cfg = |text: &str| {
            fs::create_dir_all(tmp.path().join(".git")).unwrap();
            fs::write(tmp.path().join(".git/config"), text).unwrap();
            origin_is_github(tmp.path())
        };
        assert!(!origin_is_github(tmp.path()), "no .git");
        assert!(cfg(
            "[remote \"origin\"]\n\turl = https://github.com/acme/w.git\n"
        ));
        assert!(cfg(
            "[remote \"origin\"]\n\turl = git@github.com:acme/w.git\n"
        ));
        assert!(!cfg(
            "[remote \"origin\"]\n\turl = https://gitlab.com/acme/w.git\n"
        ));
        assert!(!cfg(
            "[remote \"fork\"]\n\turl = https://github.com/acme/w.git\n"
        ));
        assert!(!cfg("[core]\n\tbare = false\n"));
    }

    #[test]
    fn a_linked_worktree_reads_the_main_checkout_origin() {
        let tmp = tempfile::tempdir().unwrap();
        let main = tmp.path().join("main");
        let gitdir = main.join(".git/worktrees/wt");
        fs::create_dir_all(&gitdir).unwrap();
        fs::write(gitdir.join("commondir"), "../..\n").unwrap();
        fs::write(
            main.join(".git/config"),
            "[remote \"origin\"]\n\turl = https://github.com/acme/w.git\n",
        )
        .unwrap();
        let wt = tmp.path().join("wt");
        fs::create_dir_all(&wt).unwrap();
        fs::write(wt.join(".git"), format!("gitdir: {}\n", gitdir.display())).unwrap();
        assert!(origin_is_github(&wt));
    }

    #[test]
    fn github_urls_in_every_form_name_owner_and_repository() {
        let on = |u: &str| github_owner_name(u);
        let acme = Some(("acme".to_string(), "w".to_string()));
        assert_eq!(on("https://github.com/acme/w.git"), acme);
        assert_eq!(on("https://github.com/acme/w/"), acme);
        assert_eq!(on("git@github.com:acme/w.git"), acme);
        assert_eq!(on("ssh://git@github.com/acme/w"), acme);
        assert_eq!(on("https://user@github.com/acme/w.git"), acme);
        assert_eq!(on("https://github.com/acme"), None);
        assert_eq!(on("https://github.com/acme/w/extra"), None);
        assert_eq!(on("https://gitlab.com/acme/w.git"), None);
        // Codex review on #2608: an explicit port is part of the authority.
        assert_eq!(on("ssh://git@github.com:22/acme/w.git"), acme);
        assert_eq!(on("https://github.com:443/acme/w"), acme);
        assert_eq!(on("ssh://git@notgithub.com/acme/w.git"), None);
        assert_eq!(on("git@github.com.evil.example:acme/w.git"), None);
    }

    #[test]
    fn config_values_are_decoded_like_git() {
        assert_eq!(
            config_value(r#" "https://github.com/acme/widgets.git" "#),
            "https://github.com/acme/widgets.git"
        );
        assert_eq!(
            config_value("git@github.com:acme/w.git ; note"),
            "git@github.com:acme/w.git"
        );
        assert_eq!(
            config_value("git@github.com:acme/w.git # note"),
            "git@github.com:acme/w.git"
        );
        assert_eq!(config_value(r#""a;b#c""#), "a;b#c");
        assert_eq!(config_value(r#""a\"b""#), r#"a"b"#);
    }

    #[test]
    fn a_quoted_origin_url_names_the_repository() {
        let tmp = tempfile::tempdir().unwrap();
        fs::create_dir_all(tmp.path().join(".git")).unwrap();
        fs::write(
            tmp.path().join(".git/config"),
            "[remote \"origin\"]\n\turl = \"https://github.com/acme/widgets.git\"\n",
        )
        .unwrap();
        assert_eq!(
            github_repo(tmp.path()),
            Some(("acme".to_string(), "widgets".to_string()))
        );
    }

    #[test]
    fn a_separate_git_dir_checkout_reads_its_own_git_directory() {
        let tmp = tempfile::tempdir().unwrap();
        let git_dir = tmp.path().join("store/widgets.git");
        fs::create_dir_all(&git_dir).unwrap();
        fs::write(
            git_dir.join("config"),
            "[remote \"origin\"]\n\turl = https://github.com/acme/widgets.git\n",
        )
        .unwrap();
        let work = tmp.path().join("work");
        fs::create_dir_all(&work).unwrap();
        fs::write(
            work.join(".git"),
            format!("gitdir: {}\n", git_dir.display()),
        )
        .unwrap();
        assert_eq!(
            github_repo(&work),
            Some(("acme".to_string(), "widgets".to_string()))
        );
    }

    #[test]
    fn a_lock_whose_time_is_not_written_yet_is_live_until_its_file_ages() {
        let tmp = tempfile::tempdir().unwrap();
        let a = tmp.path();
        fs::create_dir_all(dir(a)).unwrap();
        // The window between create_new and the holder's write: an empty file.
        fs::write(lock_path(a), "").unwrap();
        let now = now_unix();
        assert!(!take_lock(a, now));
        assert!(!take_lock(a, now + LOCK_TTL.as_secs() as i64 - 2));
        assert!(take_lock(a, now + LOCK_TTL.as_secs() as i64 + 2));
    }

    #[test]
    fn the_lock_admits_one_until_it_is_released_or_too_old() {
        let tmp = tempfile::tempdir().unwrap();
        let a = tmp.path();
        let lock = LOCK_TTL.as_secs() as i64;
        assert!(take_lock(a, 100));
        assert!(!take_lock(a, 100));
        assert!(!take_lock(a, 100 + lock - 1));
        // A lock left by a dead child is taken over once it is LOCK_TTL old.
        assert!(take_lock(a, 100 + lock));
        fs::remove_file(lock_path(a)).unwrap();
        assert!(take_lock(a, 100 + lock));
    }

    #[test]
    fn a_non_cli_executable_is_never_reinvoked() {
        let tmp = tempfile::tempdir().unwrap();
        fs::create_dir_all(tmp.path().join(".git")).unwrap();
        fs::write(
            tmp.path().join(".git/config"),
            "[remote \"origin\"]\n\turl = https://github.com/acme/w.git\n",
        )
        .unwrap();
        // The test binary itself: re-invoking it would rerun this suite.
        let exe = std::env::current_exe().unwrap();
        assert!(!maybe_refresh(tmp.path(), &exe, false));
        assert!(!lock_path(tmp.path()).exists());
    }
}
