//! Issue #2588: the statusline `⬇️ issues 🔀 PRs` counts are refreshed by a detached
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
    std::fs::create_dir_all(repo.join(".moai-cli")).unwrap();
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
    // The render may keep a session cache under `$TMPDIR`: a folder of its own, so nothing
    // reaches the real temporary folder.
    let tmp = tempfile::tempdir().unwrap();
    let started = Instant::now();
    let mut child = Command::new(crate::common::MOAI_BIN)
        .args(["statusline", "--render-budget-ms", "3600000"])
        .current_dir(&f.repo)
        .clean_env()
        .env("PATH", &f.bin)
        .env("HOME", &f.home)
        .env("TMPDIR", tmp.path())
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
    // AC-3: a count fetched long ago is not shown as current; the unavailable window says
    // `n/a` (SPEC-20261121-statusline-detached REQ-3) — never the stale numbers.
    let (first, _) = render(&f);
    assert!(
        !first.contains("⬇️ 7") && !first.contains("🔀 3"),
        "stale counts shown: {first}"
    );
    assert!(first.contains("n/a"), "the failed read says so: {first}");
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
    assert!(second.contains("⬇️ 12 🔀 5"), "{second}");
    assert_eq!(std::fs::read_to_string(f.repo.join(GO_CACHE)).unwrap(), go);
    assert_eq!(
        gh_calls(&f),
        [CALL],
        "the fresh second render must not refetch"
    );
}

/// SPEC-20261121-project-state-move REQ-10 (card criterion 2): a checkout with the marker and a
/// git-tracked earlier state folder keeps its counts cache under the folder the resolver names,
/// and nothing is created beside it under the earlier folder. The folder names come from the
/// library constants, so the test spells no state folder name of its own.
#[test]
fn a_marked_checkout_with_a_tracked_earlier_folder_writes_the_cache_under_the_new_folder() {
    let f = with_gh("https://github.com/acme/widgets.git", 0);
    let earlier = f.repo.join(moai_cli::PROJECT_STATE_DIR);
    let tracked = format!("{}/reports/tracked.md", moai_cli::PROJECT_STATE_DIR);
    write(&f.repo, &tracked, "tracked report\n");
    let marker = f
        .repo
        .join(moai_cli::state_root::MOAI_DIR)
        .join(moai_cli::state_root::RUST_MARKER);
    std::fs::create_dir_all(marker.parent().unwrap()).unwrap();
    std::fs::write(&marker, moai_cli::state_root::marker_bytes("test", &[])).unwrap();
    common::git::git_ok(&f.repo, &["add", "--force", "--", &tracked]);
    assert_eq!(
        moai_cli::state_root::dir(&f.repo),
        moai_cli::state_root::MOAI_DIR,
        "the marker puts the checkout on the new folder"
    );
    let new_cache = f
        .repo
        .join(moai_cli::state_root::MOAI_DIR)
        .join("state/github/counts.json");
    let _ = render(&f);
    assert!(
        wait_for(|| std::fs::read(&new_cache)
            .ok()
            .and_then(|b| serde_json::from_slice::<serde_json::Value>(&b).ok())
            .is_some_and(|c| c["open_issues"] == 12)),
        "no refresh under the new folder"
    );
    assert!(
        wait_for(|| !f
            .repo
            .join(moai_cli::state_root::MOAI_DIR)
            .join("state/github/refresh.lock")
            .exists()),
        "lock left behind"
    );
    assert!(
        !earlier.join("state").exists(),
        "the earlier folder got a state folder: {:?}",
        std::fs::read_dir(&earlier).map(|d| d.flatten().map(|e| e.file_name()).collect::<Vec<_>>())
    );
    assert!(earlier.join("reports/tracked.md").is_file());
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

// ---- SPEC-20261122-statusline-limits AC-6 (REQ-8, D-6): the Orca pass-through ----

/// The Orca script below the temporary HOME (D-6: the file Orca installs).
#[cfg(unix)]
const ORCA_SCRIPT: &str = ".orca/agent-hooks/claude-statusline.sh";
/// The three hook variables D-6 names, with fake values.
#[cfg(unix)]
const ORCA_VARS: [(&str, &str); 3] = [
    ("ORCA_AGENT_HOOK_PORT", "1"),
    ("ORCA_AGENT_HOOK_TOKEN", "fake"),
    ("ORCA_PANE_KEY", "tab:1"),
];

/// A stdin payload with `rate_limits` (or without it).
#[cfg(unix)]
fn orca_stdin(f: &Fixture, with_limits: bool) -> String {
    let mut v = serde_json::json!({
        "workspace": {"project_dir": f.repo},
        "context_window": {"used_percentage": 10},
    });
    if with_limits {
        v["rate_limits"] = serde_json::json!({
            "seven_day": {"used_percentage": 21, "resets_at": 4_102_444_800_i64}
        });
    }
    v.to_string()
}

/// A fake Orca script: keeps its stdin in `orca.in`, adds one line to `orca.calls`, then sleeps
/// `delay` seconds. Only absolute programs, since PATH holds `git` alone.
#[cfg(unix)]
fn orca_script(f: &Fixture, delay: u32) -> String {
    let dir = f.log.parent().unwrap().display().to_string();
    format!(
        "#!/bin/sh\n/bin/cat > '{dir}/orca.in'\necho call >> '{dir}/orca.calls'\n/bin/sleep {delay}\n"
    )
}

/// Writes the fake script the way Orca installs it: an executable file of mode 0755 in
/// folders of mode 0755 (set, not left to the umask).
#[cfg(unix)]
fn orca_install(path: &Path, text: &str) {
    use std::os::unix::fs::PermissionsExt;
    let hooks = path.parent().unwrap();
    std::fs::create_dir_all(hooks).unwrap();
    for dir in [hooks, hooks.parent().unwrap()] {
        std::fs::set_permissions(dir, std::fs::Permissions::from_mode(0o755)).unwrap();
    }
    common::write_executable(path, text);
    std::fs::set_permissions(path, std::fs::Permissions::from_mode(0o755)).unwrap();
}

#[cfg(unix)]
fn orca_calls(f: &Fixture) -> usize {
    std::fs::read_to_string(f.log.parent().unwrap().join("orca.calls"))
        .map(|s| s.lines().count())
        .unwrap_or(0)
}

/// One render with `stdin` and these extra variables; its output and how long it took.
#[cfg(unix)]
fn orca_render(f: &Fixture, stdin: &str, vars: &[(&str, &str)]) -> (String, Duration) {
    let tmp = tempfile::tempdir().unwrap();
    orca_render_in(f, tmp.path(), stdin, vars)
}

/// [`orca_render`] with a given `TMPDIR`, so several renders share the cache folder (and the
/// wrapper lock in it).
#[cfg(unix)]
fn orca_render_in(
    f: &Fixture,
    tmp: &Path,
    stdin: &str,
    vars: &[(&str, &str)],
) -> (String, Duration) {
    let started = Instant::now();
    let mut c = Command::new(crate::common::MOAI_BIN);
    c.args(["statusline", "--render-budget-ms", "3600000"])
        .current_dir(&f.repo)
        .clean_env()
        .env("PATH", &f.bin)
        .env("HOME", &f.home)
        .env("TMPDIR", tmp)
        .env("NO_COLOR", "1");
    for (k, v) in vars {
        c.env(k, v);
    }
    let mut child = c
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

/// AC-6: with the hook variables, an executable script and `rate_limits` in the stdin, the
/// script starts once and reads exactly the bytes the status line received.
#[cfg(unix)]
#[test]
fn orca_pass_through_starts_the_script_once_with_the_stdin_bytes() {
    let f = fixture("https://github.com/acme/widgets.git", None);
    orca_install(&f.home.join(ORCA_SCRIPT), &orca_script(&f, 0));
    let stdin = orca_stdin(&f, true);
    let (out, _) = orca_render(&f, &stdin, &ORCA_VARS);
    assert!(out.contains("CW █░░░░░░░░░ 10%"), "{out}");
    assert!(wait_for(|| orca_calls(&f) == 1), "the script did not start");
    std::thread::sleep(Duration::from_millis(300));
    assert_eq!(orca_calls(&f), 1, "started more than once");
    let got = std::fs::read(f.log.parent().unwrap().join("orca.in")).unwrap();
    assert_eq!(got, stdin.as_bytes());
}

/// AC-6 (D-6): the endpoint file variable alone is enough to start the script.
#[cfg(unix)]
#[test]
fn orca_pass_through_starts_with_the_endpoint_variable() {
    let f = fixture("https://github.com/acme/widgets.git", None);
    orca_install(&f.home.join(ORCA_SCRIPT), &orca_script(&f, 0));
    let endpoint = f.log.parent().unwrap().join("endpoint.env");
    std::fs::write(&endpoint, "").unwrap();
    let (_, _) = orca_render(
        &f,
        &orca_stdin(&f, true),
        &[("ORCA_AGENT_HOOK_ENDPOINT", endpoint.to_str().unwrap())],
    );
    assert!(wait_for(|| orca_calls(&f) == 1), "the script did not start");
}

/// AC-6: no `rate_limits`, no variable, only some of the three, a script that is not
/// executable, a folder in its place or no script at all start nothing; nor does a script
/// that others may change: a link in its place or in place of a folder, a script or folder
/// writable by group or others, or a script with only the group's execute bit. (A file of
/// another owner needs a second user and is not tested here.)
#[cfg(unix)]
#[test]
fn orca_pass_through_starts_nothing_without_limits_variables_or_an_executable() {
    let two = &ORCA_VARS[..2];
    // (case, stdin has rate_limits, hook variables set)
    type Case<'a> = (&'a str, bool, &'a [(&'a str, &'a str)]);
    let cases: [Case; 12] = [
        ("no rate_limits", false, &ORCA_VARS),
        ("no variable", true, &[]),
        ("two of three variables", true, two),
        ("not executable", true, &ORCA_VARS),
        ("a folder", true, &ORCA_VARS),
        ("no script", true, &ORCA_VARS),
        ("a link to the script", true, &ORCA_VARS),
        ("a linked hooks folder", true, &ORCA_VARS),
        ("a group-writable script", true, &ORCA_VARS),
        ("an other-writable hooks folder", true, &ORCA_VARS),
        ("a group-writable orca folder", true, &ORCA_VARS),
        ("only the group may execute", true, &ORCA_VARS),
    ];
    for (name, limits, vars) in cases {
        let f = fixture("https://github.com/acme/widgets.git", None);
        let script = f.home.join(ORCA_SCRIPT);
        match name {
            "not executable" => {
                std::fs::create_dir_all(script.parent().unwrap()).unwrap();
                std::fs::write(&script, orca_script(&f, 0)).unwrap();
            }
            "a folder" => std::fs::create_dir_all(&script).unwrap(),
            "no script" => {}
            "a link to the script" => {
                let real = f.home.join("real-statusline.sh");
                orca_install(&script, &orca_script(&f, 0));
                std::fs::rename(&script, &real).unwrap();
                std::os::unix::fs::symlink(&real, &script).unwrap();
            }
            "a linked hooks folder" => {
                let real = f.home.join("real-hooks/claude-statusline.sh");
                orca_install(&real, &orca_script(&f, 0));
                std::fs::create_dir_all(f.home.join(".orca")).unwrap();
                std::os::unix::fs::symlink(real.parent().unwrap(), script.parent().unwrap())
                    .unwrap();
            }
            _ => {
                orca_install(&script, &orca_script(&f, 0));
                let (target, mode) = match name {
                    "a group-writable script" => (script.clone(), 0o775),
                    "an other-writable hooks folder" => (script.parent().unwrap().into(), 0o757),
                    "a group-writable orca folder" => (f.home.join(".orca"), 0o775),
                    "only the group may execute" => (script.clone(), 0o654),
                    // The other cases keep the install as it is.
                    _ => (script.clone(), 0),
                };
                if mode != 0 {
                    use std::os::unix::fs::PermissionsExt;
                    std::fs::set_permissions(&target, std::fs::Permissions::from_mode(mode))
                        .unwrap();
                }
            }
        }
        let (out, _) = orca_render(&f, &orca_stdin(&f, limits), vars);
        assert!(out.contains("CW █░░░░░░░░░ 10%"), "{name}: {out}");
        std::thread::sleep(Duration::from_millis(700));
        assert_eq!(orca_calls(&f), 0, "{name}: the script started");
    }
}

/// AC-6: the render returns before a slow script ends.
#[cfg(unix)]
#[test]
fn orca_pass_through_does_not_wait_for_a_slow_script() {
    let f = fixture("https://github.com/acme/widgets.git", None);
    orca_install(&f.home.join(ORCA_SCRIPT), &orca_script(&f, 20));
    let (out, took) = orca_render(&f, &orca_stdin(&f, true), &ORCA_VARS);
    assert!(out.contains("CW █░░░░░░░░░ 10%"), "{out}");
    assert!(took < Duration::from_secs(10), "the render waited {took:?}");
    assert!(wait_for(|| orca_calls(&f) == 1), "the script did not start");
}

/// A stand-in that records its pid, its process group, its environment and the call, then
/// waits `hang` seconds in its own process (`exec`), the way a wrapper that hangs would.
#[cfg(unix)]
fn orca_hanging(f: &Fixture, hang: u32) -> String {
    let dir = f.log.parent().unwrap().display().to_string();
    format!(
        "#!/bin/sh\n/bin/cat > /dev/null\n/usr/bin/env > '{dir}/orca.env'\n\
         echo \"$$ $(/bin/ps -o pgid= -p $$)\" >> '{dir}/orca.pids'\n\
         echo call >> '{dir}/orca.calls'\nexec /bin/sleep {hang}\n"
    )
}

/// The pids and process groups the stand-in recorded, one pair per start.
#[cfg(unix)]
fn orca_pids(f: &Fixture) -> Vec<(i32, i32)> {
    std::fs::read_to_string(f.log.parent().unwrap().join("orca.pids"))
        .unwrap_or_default()
        .lines()
        .filter_map(|l| {
            let mut it = l.split_whitespace().map(|n| n.parse::<i32>());
            Some((it.next()?.ok()?, it.next()?.ok()?))
        })
        .collect()
}

#[cfg(unix)]
fn alive(pid: i32) -> bool {
    // SAFETY: signal 0 only asks whether the process exists; nothing is sent.
    unsafe { libc::kill(pid, 0) == 0 }
}

/// Ends what a test started, so no stand-in outlives it.
#[cfg(unix)]
fn orca_stop(f: &Fixture) {
    for (pid, _) in orca_pids(f) {
        // SAFETY: a plain kill of a pid this test's stand-in recorded.
        unsafe { libc::kill(pid, libc::SIGKILL) };
    }
}

/// While a wrapper still runs, renders start no other one (D-10): five renders against a
/// wrapper that hangs leave one live process and one call.
#[cfg(unix)]
#[test]
fn orca_pass_through_starts_no_second_wrapper_while_one_runs() {
    let f = fixture("https://github.com/acme/widgets.git", None);
    orca_install(&f.home.join(ORCA_SCRIPT), &orca_hanging(&f, 60));
    let tmp = tempfile::tempdir().unwrap();
    let stdin = orca_stdin(&f, true);
    orca_render_in(&f, tmp.path(), &stdin, &ORCA_VARS);
    assert!(wait_for(|| orca_calls(&f) == 1), "the script did not start");
    for _ in 0..4 {
        let (out, _) = orca_render_in(&f, tmp.path(), &stdin, &ORCA_VARS);
        assert!(out.contains("CW █░░░░░░░░░ 10%"), "{out}");
    }
    std::thread::sleep(Duration::from_millis(700));
    let pids = orca_pids(&f);
    let live = pids.iter().filter(|(p, _)| alive(*p)).count();
    orca_stop(&f);
    assert_eq!(orca_calls(&f), 1, "{pids:?}");
    assert_eq!(live, 1, "{pids:?}");
}

/// A wrapper that hangs ends at its time limit, and the next render starts one again (D-10).
#[cfg(unix)]
#[test]
fn a_hanging_wrapper_ends_at_its_time_limit_and_the_next_render_starts_one() {
    let limit = u64::from(moai_cli::statusline::orca::TIME_LIMIT_SECS);
    let f = fixture("https://github.com/acme/widgets.git", None);
    orca_install(&f.home.join(ORCA_SCRIPT), &orca_hanging(&f, 300));
    let tmp = tempfile::tempdir().unwrap();
    let stdin = orca_stdin(&f, true);
    orca_render_in(&f, tmp.path(), &stdin, &ORCA_VARS);
    assert!(wait_for(|| orca_calls(&f) == 1), "the script did not start");
    let (pid, _) = orca_pids(&f)[0];
    let deadline = Instant::now() + Duration::from_secs(limit + 10);
    while alive(pid) && Instant::now() < deadline {
        std::thread::sleep(Duration::from_millis(200));
    }
    let ended = !alive(pid);
    orca_render_in(&f, tmp.path(), &stdin, &ORCA_VARS);
    let again = wait_for(|| orca_calls(&f) == 2);
    orca_stop(&f);
    assert!(ended, "the wrapper outlived its time limit of {limit} s");
    assert!(again, "no wrapper started after the first one ended");
}

/// The wrapper gets the hook variables and none of the render's other variables, and runs in a
/// process group of its own (D-10).
#[cfg(unix)]
#[test]
fn the_wrapper_gets_the_hook_variables_only_and_its_own_process_group() {
    let f = fixture("https://github.com/acme/widgets.git", None);
    orca_install(&f.home.join(ORCA_SCRIPT), &orca_hanging(&f, 0));
    let mut vars = ORCA_VARS.to_vec();
    vars.push(("UNRELATED_SECRET", "s3cr3t"));
    orca_render(&f, &orca_stdin(&f, true), &vars);
    assert!(wait_for(|| orca_calls(&f) == 1), "the script did not start");
    let env = std::fs::read_to_string(f.log.parent().unwrap().join("orca.env")).unwrap();
    assert!(
        env.lines().any(|l| l.starts_with("ORCA_AGENT_HOOK_TOKEN=")),
        "{env}"
    );
    assert!(
        env.lines().any(|l| l.starts_with("ORCA_PANE_KEY=")),
        "{env}"
    );
    assert!(!env.contains("UNRELATED_SECRET"), "{env}");
    assert!(!env.contains("NO_COLOR"), "{env}");
    let pids = orca_pids(&f);
    assert_eq!(pids.len(), 1, "{pids:?}");
    assert_eq!(pids[0].0, pids[0].1, "not its own process group: {pids:?}");
}

/// A payload larger than a pipe holds starts nothing: it would reach the wrapper cut short
/// (D-10).
#[cfg(unix)]
#[test]
fn an_oversized_payload_starts_no_wrapper() {
    let f = fixture("https://github.com/acme/widgets.git", None);
    orca_install(&f.home.join(ORCA_SCRIPT), &orca_script(&f, 0));
    let mut v: serde_json::Value = serde_json::from_str(&orca_stdin(&f, true)).unwrap();
    v["padding"] = serde_json::json!("x".repeat(20 * 1024));
    let (out, _) = orca_render(&f, &v.to_string(), &ORCA_VARS);
    assert!(out.contains("CW █░░░░░░░░░ 10%"), "{out}");
    std::thread::sleep(Duration::from_millis(700));
    assert_eq!(orca_calls(&f), 0, "the script started");
}
