//! Issue #2588: the statusline `⬇️ issues / 🔀 PRs` counts are refreshed by a detached
//! child the render starts, and the render never waits for it.
//!
//! Every case runs the real binary with a cleared environment, a temporary HOME and a
//! temporary checkout whose `origin` is on github.com. `gh` is a fake script on PATH that
//! logs each call, so no case reaches the network.

mod common;
use common::CleanEnv;

use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use std::time::{Duration, Instant};

struct Fixture {
    _tmp: tempfile::TempDir,
    home: PathBuf,
    repo: PathBuf,
    bin: PathBuf,
    log: PathBuf,
}

const GO_CACHE: &str = ".moai/state/github/counts.json";
const CACHE: &str = ".moai-cli/state/github/counts.json";
const LOCK: &str = ".moai-cli/state/github/refresh.lock";

/// A checkout with `origin` at `origin_url`. With `gh_delay` the PATH holds a fake `gh`
/// (see [`gh_script`]) beside `git`; without it, only `git`.
fn fixture(origin_url: &str, gh_delay: Option<u32>) -> Fixture {
    let tmp = tempfile::tempdir().unwrap();
    let home = tmp.path().join("home");
    let repo = tmp.path().join("repo");
    let bin = tmp.path().join("bin");
    for d in [&home, &repo, &bin] {
        std::fs::create_dir_all(d).unwrap();
    }
    common::git::git_ok(&repo, &["init", "-q", "-b", "main"]);
    common::git::git_ok(&repo, &["remote", "add", "origin", origin_url]);
    let log = tmp.path().join("gh.log");
    // Only git is linked next to the fake gh, so a real gh elsewhere is never found.
    let real_git = String::from_utf8(
        Command::new("sh")
            .args(["-c", "command -v git"])
            .output()
            .unwrap()
            .stdout,
    )
    .unwrap();
    #[cfg(unix)]
    std::os::unix::fs::symlink(real_git.trim(), bin.join("git")).unwrap();
    if let Some(delay) = gh_delay {
        common::write_executable(&bin.join("gh"), &gh_script(&log, delay));
    }
    Fixture {
        _tmp: tmp,
        home,
        repo,
        bin,
        log,
    }
}

/// A fake gh: logs one line per call (the call shape, the inherited `GH_REPO`/`GH_HOST`,
/// the owner and name fields), sleeps `delay` seconds, answers 12 issues / 5 PRs.
fn gh_script(log: &Path, delay: u32) -> String {
    format!(
        "#!/bin/sh\necho \"$1 $2 $3 $4 GH_REPO=${{GH_REPO-unset}} GH_HOST=${{GH_HOST-unset}} $8 ${{10}}\" >> '{log}'\n/bin/sleep {delay}\ncase \"$1 $2\" in 'api graphql') echo '12 5';; *) exit 2;; esac\n",
        log = log.display()
    )
}

/// The one call a refresh makes: explicit host and repository, with the inherited
/// `GH_REPO`/`GH_HOST` the render sets (see [`render`]) removed.
const CALL: &str =
    "api graphql --hostname github.com GH_REPO=unset GH_HOST=unset owner=acme name=widgets";

fn with_gh(origin_url: &str, delay: u32) -> Fixture {
    fixture(origin_url, Some(delay))
}

fn render(f: &Fixture) -> (String, Duration) {
    let stdin = serde_json::json!({
        "workspace": {
            "project_dir": f.repo,
            "repo": { "host": "github.com", "owner": "acme", "name": "widgets" }
        }
    })
    .to_string();
    let started = Instant::now();
    let mut child = Command::new(crate::common::MOAI_BIN)
        .args(["statusline", "--render-budget-ms", "3600000"])
        .current_dir(&f.repo)
        .clean_env()
        .env("PATH", &f.bin)
        .env("HOME", &f.home)
        .env("NO_COLOR", "1")
        // An inherited gh target must not redirect the refresh (Codex review on #2604).
        .env("GH_REPO", "elsewhere/other")
        .env("GH_HOST", "ghe.example.com")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::null())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, stdin.as_bytes());
    let out = child.wait_with_output().unwrap();
    assert!(out.status.success());
    (String::from_utf8(out.stdout).unwrap(), started.elapsed())
}

fn write(base: &Path, rel: &str, text: &str) {
    let path = base.join(rel);
    std::fs::create_dir_all(path.parent().unwrap()).unwrap();
    std::fs::write(path, text).unwrap();
}

fn cache(f: &Fixture) -> Option<serde_json::Value> {
    std::fs::read(f.repo.join(CACHE))
        .ok()
        .and_then(|b| serde_json::from_slice(&b).ok())
}

/// Polls until `done` holds or a generous bound passes (a loaded machine is slow).
fn wait_for(mut done: impl FnMut() -> bool) -> bool {
    let deadline = Instant::now() + Duration::from_secs(30);
    while Instant::now() < deadline {
        if done() {
            return true;
        }
        std::thread::sleep(Duration::from_millis(50));
    }
    done()
}

fn gh_calls(f: &Fixture) -> Vec<String> {
    std::fs::read_to_string(&f.log)
        .unwrap_or_default()
        .lines()
        .map(str::to_string)
        .collect()
}

#[test]
fn a_stale_cache_is_refreshed_in_the_background_and_the_go_file_is_left_alone() {
    let f = with_gh("https://github.com/acme/widgets.git", 0);
    let go = r#"{"open_issues":7,"open_prs":3,"fetched_at":1}"#;
    write(&f.repo, GO_CACHE, go);
    // AC-3: a count fetched long ago is not shown as current.
    let (first, _) = render(&f);
    assert!(!first.contains('⬇'), "stale counts shown: {first}");
    assert!(
        wait_for(|| cache(&f).is_some_and(|c| c["open_issues"] == 12)),
        "no refresh: {:?}",
        cache(&f)
    );
    assert!(wait_for(|| !f.repo.join(LOCK).exists()), "lock left behind");
    let c = cache(&f).unwrap();
    assert_eq!(
        (c["open_issues"].as_i64(), c["open_prs"].as_i64()),
        (Some(12), Some(5))
    );
    let (second, _) = render(&f);
    assert!(second.contains("⬇️ 12 / 🔀 5"), "{second}");
    assert_eq!(std::fs::read_to_string(f.repo.join(GO_CACHE)).unwrap(), go);
    assert_eq!(
        gh_calls(&f),
        [CALL],
        "the fresh second render must not refetch"
    );
}

#[test]
fn the_render_does_not_wait_for_a_slow_gh() {
    let f = with_gh("git@github.com:acme/widgets.git", 15);
    let (_, took) = render(&f);
    assert!(took < Duration::from_secs(10), "render waited {took:?}");
    assert!(wait_for(|| gh_calls(&f).len() == 1), "refresh not started");
}

#[test]
fn concurrent_renders_start_a_single_refresh() {
    let f = with_gh("https://github.com/acme/widgets.git", 1);
    std::thread::scope(|s| {
        for _ in 0..8 {
            s.spawn(|| render(&f));
        }
    });
    assert!(wait_for(
        || cache(&f).is_some() && !f.repo.join(LOCK).exists()
    ));
    assert_eq!(gh_calls(&f), [CALL]);
}

#[test]
fn no_refresh_for_a_forge_override_or_a_non_github_origin() {
    let none = with_gh("https://github.com/acme/widgets.git", 0);
    write(
        &none.repo,
        ".moai/config/sections/statusline.yaml",
        "statusline:\n  forge: none\n",
    );
    let lab = with_gh("https://gitlab.com/acme/widgets.git", 0);
    for f in [&none, &lab] {
        render(f);
    }
    // Negative check: give a wrongly started child time to call gh.
    std::thread::sleep(Duration::from_millis(1500));
    for f in [&none, &lab] {
        assert!(gh_calls(f).is_empty(), "{}", f.repo.display());
        assert!(cache(f).is_none());
    }
}

#[test]
fn a_missing_gh_marks_the_counts_suppressed_without_a_retry_storm() {
    let f = fixture("https://github.com/acme/widgets.git", None);
    render(&f);
    assert!(
        wait_for(|| cache(&f).is_some_and(|c| c["suppressed"] == true)),
        "{:?}",
        cache(&f)
    );
    assert!(wait_for(|| !f.repo.join(LOCK).exists()));
    let attempted = cache(&f).unwrap()["attempted_at"].as_i64();
    assert!(attempted.is_some());
    let (out, _) = render(&f);
    assert!(!out.contains('⬇'), "{out}");
    std::thread::sleep(Duration::from_millis(500));
    assert_eq!(cache(&f).unwrap()["attempted_at"].as_i64(), attempted);
    assert!(!f.repo.join(LOCK).exists(), "a second refresh started");
}
