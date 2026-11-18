//! #2359 (user decision 2026-11-10, comment 6760976883; resolves #2361), changed by #2971:
//! `-a` without `--name` is named `agent-N`. N is counted per project (one clone, all its
//! worktrees), shared across `cc`, `glm` and `codex`. A session is alive while its launcher holds
//! an exclusive lock on `<sessions dir>/agent-<n>.live` (the OS drops it on any exit, SIGKILL
//! included) or while a registry entry `agent-<n>.json` has a running lease. The next number is
//! the highest live one plus 1, or 1 when none is live; a gap below a live number is not
//! filled. Two concurrent launches never get the same number. `--dry-run`, `--name` and `-l`
//! take no number, and an unusable file is an error, never a silent `agent-1`. Real launches run
//! stub `claude` / `codex` executables from a temporary PATH in a temporary git project; no real
//! host starts. A stub started with `MOAI_TEST_HOLD_UNTIL` stays alive until that file appears.
#![cfg(unix)]
mod common;
use serde_json::Value;
use std::os::unix::fs::PermissionsExt;
use std::{
    fs,
    path::{Path, PathBuf},
    process::{Child, Command, Output, Stdio},
    time::{Duration, Instant},
};

struct Fixture {
    project: tempfile::TempDir,
    bin: tempfile::TempDir,
    home: tempfile::TempDir,
}

fn git(dir: &Path, args: &[&str]) {
    let ok = common::git::command()
        .args(["-c", "user.name=t", "-c", "user.email=t@example.invalid"])
        .args(args)
        .current_dir(dir)
        .stdout(Stdio::null())
        .stderr(Stdio::null())
        .status()
        .unwrap()
        .success();
    assert!(ok, "git {args:?}");
}

/// A host stub that exits at once, or, when `MOAI_TEST_HOLD_UNTIL` is set, writes its pid to
/// `MOAI_TEST_READY` and stays alive until the `MOAI_TEST_HOLD_UNTIL` file exists.
const STUB: &str = "#!/bin/sh
printf started > \"$MOAI_TEST_HOST_MARKER\"
if [ -n \"$MOAI_TEST_HOLD_UNTIL\" ]; then
  echo $$ > \"$MOAI_TEST_READY.tmp\" && mv \"$MOAI_TEST_READY.tmp\" \"$MOAI_TEST_READY\"
  while [ ! -e \"$MOAI_TEST_HOLD_UNTIL\" ]; do sleep 0.05; done
fi
exit 0
";

fn fixture() -> Fixture {
    let project = tempfile::tempdir().unwrap();
    git(project.path(), &["init", "-q"]);
    // An initialized project: with settings present no launch runs the #2481 first-launch
    // initialization, so these tests measure numbering only (tests/launcher_auto_init.rs).
    fs::create_dir(project.path().join(".moai-cli")).unwrap();
    fs::write(project.path().join(".moai-cli/project.yaml"), "{}\n").unwrap();
    let bin = tempfile::tempdir().unwrap();
    for host in ["claude", "codex"] {
        let path = bin.path().join(host);
        fs::write(&path, STUB).unwrap();
        fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
    }
    Fixture {
        project,
        bin,
        home: tempfile::tempdir().unwrap(),
    }
}

fn command(f: &Fixture, dir: &Path, args: &[&str]) -> Command {
    let mut c = Command::new(crate::common::MOAI_BIN);
    c.args(args)
        .current_dir(dir)
        .env("PATH", format!("{}:/usr/bin:/bin", f.bin.path().display()))
        .env("HOME", f.home.path())
        .env("USERPROFILE", f.home.path())
        .env("MOAI_TEST_HOST_MARKER", f.home.path().join("host-started"))
        .env("ANTHROPIC_AUTH_TOKEN", "moai-test-dummy-not-real")
        .env_remove("CLAUDE_CODE_DISABLE_ADVISOR_TOOL")
        .env_remove("MOAI_TEST_HOLD_UNTIL")
        .env_remove("MOAI_TEST_READY")
        .env_remove(moai_cli::session::AGENT_SEQ_TEST_HOLD_VAR);
    c
}

fn run_in(f: &Fixture, dir: &Path, args: &[&str]) -> Output {
    command(f, dir, args).output().unwrap()
}
fn run(f: &Fixture, args: &[&str]) -> Output {
    run_in(f, f.project.path(), args)
}

/// The session name the launcher reported, after asserting it succeeded.
fn name(out: &Output) -> String {
    assert!(
        out.status.success(),
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    v["name"].as_str().unwrap_or("").to_string()
}

/// The launcher's `notes` lines, after asserting it succeeded.
fn notes(out: &Output) -> Vec<String> {
    name(out);
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    v["notes"]
        .as_array()
        .unwrap()
        .iter()
        .map(|n| n.as_str().unwrap().to_string())
        .collect()
}

/// The leader's required warning for a number of 1 (#2971, leader decision 6880464413).
const CLAIM_WARNING: &str = "do not take over work from claim-verify alone";

fn warned(notes: &[String]) -> bool {
    notes
        .iter()
        .any(|n| n.contains(CLAIM_WARNING) && n.contains("earlier agent-1"))
}

fn sessions(f: &Fixture) -> PathBuf {
    f.project.path().join(".git/.moai-cli/sessions")
}
/// The old counter. It is no longer read or written; tests seed it to prove that.
fn store(f: &Fixture) -> PathBuf {
    sessions(f).join("agent-seq")
}
fn stored(f: &Fixture) -> Option<String> {
    fs::read_to_string(store(f)).ok()
}
fn seed(f: &Fixture, text: &str) {
    fs::create_dir_all(sessions(f)).unwrap();
    fs::write(store(f), text).unwrap();
}
fn live_path(f: &Fixture, n: u64) -> PathBuf {
    sessions(f).join(format!("agent-{n}.live"))
}
/// The `agent-<n>.live` files in the sessions folder, sorted; empty when there is no folder.
fn lives(f: &Fixture) -> Vec<String> {
    let mut names: Vec<String> = fs::read_dir(sessions(f))
        .map(|d| {
            d.map(|e| e.unwrap().file_name().to_string_lossy().to_string())
                .filter(|n| n.ends_with(".live"))
                .collect()
        })
        .unwrap_or_default();
    names.sort();
    names
}
/// A registry entry as `session-register` writes it, with an explicit lease end.
fn registry_entry(f: &Fixture, n: u64, lease_until: i64) {
    fs::create_dir_all(sessions(f)).unwrap();
    fs::write(
        sessions(f).join(format!("agent-{n}.json")),
        format!(
            "{{\"schema_version\":1,\"name\":\"agent-{n}\",\"instance\":\"{}\",\"lease_until\":{lease_until},\"registered_at\":1}}\n",
            "0123456789abcdef0123456789abcdef"
        ),
    )
    .unwrap();
}
/// Far in the future for any test run (year 2100), so no test depends on the clock.
const LEASE_RUNNING: i64 = 4_102_444_800;
/// Long past: the lease is over.
const LEASE_OVER: i64 = 1;

/// Takes the lock a running launcher would hold on `agent-<n>.live`, from the test process.
fn hold_live(f: &Fixture, n: u64) -> fs::File {
    fs::create_dir_all(sessions(f)).unwrap();
    let file = fs::OpenOptions::new()
        .read(true)
        .write(true)
        .create(true)
        .truncate(false)
        .open(live_path(f, n))
        .unwrap();
    file.lock().unwrap();
    file
}

/// A refused launch: exit 1, nothing on stdout, a message that names the way out.
fn refused(out: &Output) -> String {
    let stderr = String::from_utf8_lossy(&out.stderr).to_string();
    assert_eq!(out.status.code(), Some(1), "{stderr}");
    assert!(
        out.stdout.is_empty(),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    assert!(stderr.contains("pass --name"), "{stderr}");
    stderr
}

/// A launcher whose stub host stays alive until released: its number is live meanwhile. The
/// drop releases the stub and kills the launcher, so no test path leaves a process behind.
struct Held {
    child: Option<Child>,
    release: PathBuf,
    ready: PathBuf,
}

impl Held {
    fn spawn_in(f: &Fixture, dir: &Path, tag: &str, args: &[&str], hold_ms: Option<&str>) -> Held {
        let release = f.home.path().join(format!("release-{tag}"));
        let ready = f.home.path().join(format!("ready-{tag}"));
        let mut c = command(f, dir, args);
        c.env("MOAI_TEST_HOLD_UNTIL", &release)
            .env("MOAI_TEST_READY", &ready)
            .stdout(Stdio::piped())
            .stderr(Stdio::piped());
        if let Some(ms) = hold_ms {
            c.env(moai_cli::session::AGENT_SEQ_TEST_HOLD_VAR, ms);
        }
        Held {
            child: Some(c.spawn().unwrap()),
            release,
            ready,
        }
    }

    /// Started and ready: the stub host is running, so the launcher holds its number.
    fn hold(f: &Fixture, tag: &str, args: &[&str]) -> Held {
        let h = Held::spawn_in(f, f.project.path(), tag, args, None);
        h.wait_ready();
        h
    }

    fn wait_ready(&self) {
        let start = Instant::now();
        while fs::read_to_string(&self.ready).map_or(true, |p| p.trim().is_empty()) {
            // #2511: a hang guard, so the bound is generous on a loaded machine. The drop
            // releases and kills the launcher when this panics.
            assert!(
                start.elapsed() < Duration::from_secs(120),
                "the stub host never started"
            );
            std::thread::sleep(Duration::from_millis(25));
        }
    }

    /// Lets the stub exit and returns the launcher's output.
    fn finish(mut self) -> Output {
        fs::write(&self.release, "").unwrap();
        self.child.take().unwrap().wait_with_output().unwrap()
    }

    /// SIGKILL to the launcher alone: the stub host keeps running, the OS drops the lock.
    fn kill_launcher(&mut self) {
        let mut child = self.child.take().unwrap();
        child.kill().unwrap();
        child.wait().unwrap();
    }
}

impl Drop for Held {
    fn drop(&mut self) {
        let _ = fs::write(&self.release, "");
        if let Some(mut child) = self.child.take() {
            let _ = child.kill();
            let _ = child.wait();
        }
        // A stub whose launcher was killed is not our child: wait for it to see the release.
        if let Some(pid) = fs::read_to_string(&self.ready)
            .ok()
            .and_then(|p| p.trim().parse::<u32>().ok())
        {
            let start = Instant::now();
            while Command::new("kill")
                .args(["-0", &pid.to_string()])
                .stderr(Stdio::null())
                .status()
                .is_ok_and(|s| s.success())
            {
                if start.elapsed() > Duration::from_secs(10) {
                    let _ = Command::new("kill").args(["-9", &pid.to_string()]).status();
                    break;
                }
                std::thread::sleep(Duration::from_millis(25));
            }
        }
    }
}

const HOSTS: [&str; 3] = ["cc", "glm", "codex"];

// ---- AC1: no live session means agent-1 ------------------------------------------------------

#[test]
fn numbers_restart_at_1_after_every_session_ended_and_live_in_the_clone_common_dir() {
    let f = fixture();
    let names: Vec<String> = [
        &["cc", "-a"][..],
        &["glm", "-a"][..],
        &["codex", "-a"][..],
        &["-a", "codex"][..],
        &["cc", "-a"][..],
    ]
    .iter()
    .map(|args| name(&run(&f, args)))
    .collect();
    assert_eq!(names, ["agent-1"; 5]);
    // The old counter is neither created nor read.
    assert_eq!(stored(&f), None);
    // The lock files sit under `.git`, so git never lists them: only the fixture's settings show.
    let status = common::git::command()
        .args(["status", "--porcelain", "--ignored"])
        .current_dir(f.project.path())
        .output()
        .unwrap();
    assert_eq!(String::from_utf8_lossy(&status.stdout), "?? .moai-cli/\n");
    // codex carries the name in its first prompt, after the line that enters agent mode
    // (SPEC-20261118-launcher-agent-entry).
    let v: Value = serde_json::from_slice(&run(&f, &["codex", "-a"]).stdout).unwrap();
    assert_eq!(
        v["argv"].as_array().unwrap().last().unwrap(),
        "$moai --mode-agent\nMoAI session name: agent-1"
    );
}

#[test]
fn an_old_agent_seq_store_is_ignored_and_left_alone() {
    for text in ["41\n", "00041\n", "garbage", "18446744073709551615\n", ""] {
        let f = fixture();
        seed(&f, text);
        for host in HOSTS {
            assert_eq!(
                name(&run(&f, &[host, "-a", "--dry-run"])),
                "agent-1",
                "{text:?}"
            );
            assert_eq!(name(&run(&f, &[host, "-a"])), "agent-1", "{text:?}");
        }
        assert_eq!(stored(&f).as_deref(), Some(text));
    }
}

#[test]
fn a_leftover_live_file_without_a_lock_is_dead_and_reused() {
    for host in HOSTS {
        let f = fixture();
        fs::create_dir_all(sessions(&f)).unwrap();
        for n in [1, 2, 7] {
            fs::write(live_path(&f, n), "").unwrap();
        }
        assert_eq!(name(&run(&f, &[host, "-a", "--dry-run"])), "agent-1");
        assert_eq!(name(&run(&f, &[host, "-a"])), "agent-1", "{host}");
        // The dead files stay; nothing is cleaned up.
        assert_eq!(lives(&f), ["agent-1.live", "agent-2.live", "agent-7.live"]);
    }
}

#[test]
fn a_registry_entry_with_an_over_lease_is_dead() {
    for host in HOSTS {
        let f = fixture();
        registry_entry(&f, 1, LEASE_OVER);
        registry_entry(&f, 4, LEASE_OVER);
        assert_eq!(name(&run(&f, &[host, "-a", "--dry-run"])), "agent-1");
        assert_eq!(name(&run(&f, &[host, "-a"])), "agent-1", "{host}");
    }
}

// ---- AC2: live sessions are never overlapped --------------------------------------------------

#[test]
fn live_sessions_are_never_overlapped_and_the_next_number_follows_the_highest() {
    for (i, host) in HOSTS.iter().enumerate() {
        let f = fixture();
        let other = HOSTS[(i + 1) % 3];
        let h1 = Held::hold(&f, "one", &[host, "-a"]);
        let h2 = Held::hold(&f, "two", &[other, "-a"]);
        assert_eq!(lives(&f), ["agent-1.live", "agent-2.live"], "{host}");
        assert_eq!(name(&run(&f, &[host, "-a"])), "agent-3", "{host}");
        assert_eq!(name(&run(&f, &[other, "-a", "--dry-run"])), "agent-3");
        // agent-1 ends, agent-2 stays: 1 is a gap below a live number and is not filled.
        assert_eq!(name(&h1.finish()), "agent-1", "{host}");
        assert_eq!(name(&run(&f, &[host, "-a"])), "agent-3", "{host}");
        assert_eq!(name(&run(&f, &[host, "-a", "--dry-run"])), "agent-3");
        // Everything ended: back to 1.
        assert_eq!(name(&h2.finish()), "agent-2", "{host}");
        assert_eq!(name(&run(&f, &[host, "-a"])), "agent-1", "{host}");
    }
}

#[test]
fn only_agent_2_live_gives_agent_3_on_every_host() {
    for host in HOSTS {
        let f = fixture();
        let _lock = hold_live(&f, 2);
        assert_eq!(name(&run(&f, &[host, "-a"])), "agent-3", "{host}");
        assert_eq!(name(&run(&f, &[host, "-a", "--dry-run"])), "agent-3");
    }
}

#[test]
fn a_registry_only_session_with_a_running_lease_is_live() {
    for host in HOSTS {
        let f = fixture();
        registry_entry(&f, 4, LEASE_RUNNING);
        assert_eq!(name(&run(&f, &[host, "-a"])), "agent-5", "{host}");
        assert_eq!(name(&run(&f, &[host, "-a", "--dry-run"])), "agent-5");
        // The highest live number wins, whichever kind of proof it has.
        let _lock = hold_live(&f, 6);
        assert_eq!(name(&run(&f, &[host, "-a"])), "agent-7", "{host}");
        registry_entry(&f, 9, LEASE_RUNNING);
        assert_eq!(name(&run(&f, &[host, "-a"])), "agent-10", "{host}");
    }
}

#[test]
fn a_malformed_registry_entry_refuses_the_launch_instead_of_giving_agent_1() {
    for bad in [
        "not json",
        "{}",
        "{\"schema_version\":1,\"name\":\"agent-4\",\"instance\":\"short\",\"lease_until\":4102444800}",
        "{\"schema_version\":1,\"name\":\"agent-9\",\"instance\":\"0123456789abcdef0123456789abcdef\",\"lease_until\":4102444800}",
    ] {
        let f = fixture();
        fs::create_dir_all(sessions(&f)).unwrap();
        fs::write(sessions(&f).join("agent-4.json"), bad).unwrap();
        for args in [&["cc", "-a"][..], &["codex", "-a", "--dry-run"][..]] {
            let err = refused(&run(&f, args));
            assert!(err.contains("agent-4.json"), "{bad:?}: {err}");
        }
        assert_eq!(lives(&f), Vec::<String>::new(), "{bad:?}");
        assert!(!f.home.path().join("host-started").exists());
    }
}

// ---- AC3: concurrent launches ----------------------------------------------------------------

#[test]
fn concurrent_launches_never_share_a_number() {
    const N: usize = 8;
    let f = fixture();
    // Each launch holds the counter lock 150 ms between choosing its number and locking it, so
    // without the counter lock every launch would choose the same number. The stubs stay alive,
    // so every number stays live and the numbers must all differ.
    let held: Vec<Held> = (0..N)
        .map(|i| {
            Held::spawn_in(
                &f,
                f.project.path(),
                &format!("c{i}"),
                &[HOSTS[i % 3], "-a"],
                Some("150"),
            )
        })
        .collect();
    for h in &held {
        h.wait_ready();
    }
    let mut names: Vec<String> = held.into_iter().map(|h| name(&h.finish())).collect();
    names.sort_by_key(|n| n.trim_start_matches("agent-").parse::<u64>().unwrap());
    let want: Vec<String> = (1..=N).map(|n| format!("agent-{n}")).collect();
    assert_eq!(names, want);
    // They all ended: the next launch starts over.
    assert_eq!(name(&run(&f, &["cc", "-a"])), "agent-1");
}

#[test]
fn worktrees_of_one_clone_share_the_live_numbers() {
    let f = fixture();
    git(
        f.project.path(),
        &["commit", "-q", "--allow-empty", "-m", "base"],
    );
    let wt = f.project.path().join("wt");
    git(
        f.project.path(),
        &["worktree", "add", "-q", "-b", "lane", "wt"],
    );
    let held = Held::hold(&f, "main", &["cc", "-a"]);
    assert_eq!(name(&run_in(&f, &wt, &["codex", "-a"])), "agent-2");
    let held_wt = Held::spawn_in(&f, &wt, "wt", &["codex", "-a"], None);
    held_wt.wait_ready();
    assert_eq!(name(&run(&f, &["glm", "-a"])), "agent-3");
    assert_eq!(name(&held.finish()), "agent-1");
    assert_eq!(name(&held_wt.finish()), "agent-2");
    assert_eq!(name(&run(&f, &["glm", "-a"])), "agent-1");
}

// ---- Safety ----------------------------------------------------------------------------------

#[test]
fn killing_the_launcher_frees_its_number() {
    let f = fixture();
    let mut held = Held::hold(&f, "k", &["codex", "-a"]);
    assert_eq!(name(&run(&f, &["cc", "-a"])), "agent-2");
    held.kill_launcher();
    // The lock died with the launcher, whatever the stub host does next.
    assert_eq!(name(&run(&f, &["cc", "-a"])), "agent-1");
    assert_eq!(name(&run(&f, &["glm", "-a", "--dry-run"])), "agent-1");
}

#[test]
fn dry_run_creates_nothing_and_takes_no_number() {
    let f = fixture();
    let before = tree(&f.project.path().join(".git"));
    for host in HOSTS {
        assert_eq!(name(&run(&f, &[host, "-a", "--dry-run"])), "agent-1");
    }
    assert!(!sessions(&f).exists(), "a dry run creates no folder");
    assert_eq!(tree(&f.project.path().join(".git")), before);
    assert_eq!(name(&run(&f, &["glm", "-a"])), "agent-1");
    // With a live session: a dry run reads, takes nothing and leaves the files as they were.
    let _lock = hold_live(&f, 2);
    let before = tree(&sessions(&f));
    assert_eq!(name(&run(&f, &["codex", "-a", "--dry-run"])), "agent-3");
    assert_eq!(tree(&sessions(&f)), before);
}

/// Every path under `dir` with its length, sorted.
fn tree(dir: &Path) -> Vec<(String, u64)> {
    fn walk(dir: &Path, out: &mut Vec<(String, u64)>) {
        for e in fs::read_dir(dir).into_iter().flatten().flatten() {
            let meta = fs::symlink_metadata(e.path()).unwrap();
            out.push((e.path().display().to_string(), meta.len()));
            if meta.is_dir() {
                walk(&e.path(), out);
            }
        }
    }
    let mut out = Vec::new();
    walk(dir, &mut out);
    out.sort();
    out
}

#[test]
fn the_number_one_carries_the_claim_warning_and_other_numbers_do_not() {
    let f = fixture();
    for host in HOSTS {
        let preview = notes(&run(&f, &[host, "-a", "--dry-run"]));
        assert!(warned(&preview), "{host} dry run: {preview:?}");
        assert!(
            preview.iter().any(|n| n.contains("starts at agent-1")),
            "{host}: {preview:?}"
        );
        let real = notes(&run(&f, &[host, "-a"]));
        assert!(warned(&real), "{host}: {real:?}");
        // One line each: no line break inside a note.
        assert!(real.iter().all(|n| !n.contains('\n')), "{real:?}");
    }
    let _lock = hold_live(&f, 2);
    for host in HOSTS {
        for args in [&[host, "-a"][..], &[host, "-a", "--dry-run"][..]] {
            let out = notes(&run(&f, args));
            assert!(!warned(&out), "{args:?}: {out:?}");
            assert!(
                out.iter().any(|n| n.contains("next after live agent-2")),
                "{args:?}: {out:?}"
            );
        }
    }
    // A named session takes no number and gets no warning.
    assert!(!warned(&notes(&run(&f, &["cc", "-a", "--name", "x"]))));
}

#[test]
fn a_folder_outside_git_previews_agent_1_with_the_warning() {
    let f = fixture();
    let plain = tempfile::tempdir().unwrap();
    let preview = notes(&run_in(&f, plain.path(), &["codex", "-a", "--dry-run"]));
    assert!(warned(&preview), "{preview:?}");
    assert!(!plain.path().join(".git").exists());
}

#[test]
fn name_leader_and_roleless_take_no_number() {
    let f = fixture();
    assert_eq!(name(&run(&f, &["cc", "-a", "--name", "x"])), "x");
    assert_eq!(name(&run(&f, &["codex", "-l"])), "leader");
    assert_eq!(name(&run(&f, &["-l"])), "leader");
    assert_eq!(name(&run(&f, &["glm"])), "");
    assert!(!sessions(&f).exists());
    assert_eq!(name(&run(&f, &["cc", "-a"])), "agent-1");
}

#[test]
fn a_launch_refused_before_the_host_starts_takes_no_number() {
    let f = fixture();
    // glm without any token is blocked; an invalid model is refused.
    let out = command(&f, f.project.path(), &["glm", "-a"])
        .env_remove("ANTHROPIC_AUTH_TOKEN")
        .output()
        .unwrap();
    assert_eq!(out.status.code(), Some(1));
    assert_ne!(
        run(&f, &["cc", "-a", "--model", "two words"]).status.code(),
        Some(0)
    );
    assert_eq!(lives(&f), Vec::<String>::new());
    assert!(!f.home.path().join("host-started").exists());
    assert_eq!(name(&run(&f, &["cc", "-a"])), "agent-1");
}

#[test]
fn a_live_path_that_is_a_symlink_is_refused_and_nothing_is_written() {
    for n in [1, 3, 9] {
        let f = fixture();
        let real = f.project.path().join("elsewhere");
        fs::write(&real, "3\n").unwrap();
        fs::create_dir_all(sessions(&f)).unwrap();
        std::os::unix::fs::symlink(&real, live_path(&f, n)).unwrap();
        let before = tree(&sessions(&f));
        for args in [&["cc", "-a"][..], &["codex", "-a", "--dry-run"][..]] {
            let err = refused(&run(&f, args));
            assert!(err.contains("symlink"), "{n}: {err}");
        }
        assert_eq!(fs::read_to_string(&real).unwrap(), "3\n");
        assert!(!f.home.path().join("host-started").exists());
        // Only the lock file the launcher opens before it looks at anything may appear.
        let mut after = tree(&sessions(&f));
        after.retain(|(p, _)| !p.ends_with("/.lock"));
        assert_eq!(after, before, "{n}");
        assert_eq!(stored(&f), None);
    }
    // A dangling link is a link too.
    let f = fixture();
    fs::create_dir_all(sessions(&f)).unwrap();
    std::os::unix::fs::symlink(f.project.path().join("missing"), live_path(&f, 2)).unwrap();
    refused(&run(&f, &["cc", "-a"]));
}

#[test]
fn a_live_path_that_is_not_a_regular_file_is_refused_without_blocking() {
    // A directory.
    let f = fixture();
    fs::create_dir_all(live_path(&f, 2)).unwrap();
    refused(&run(&f, &["cc", "-a"]));
    // A FIFO: opening it for read and write would not fail, and locking it is no proof.
    let f = fixture();
    fs::create_dir_all(sessions(&f)).unwrap();
    assert!(Command::new("mkfifo")
        .arg(live_path(&f, 2))
        .status()
        .unwrap()
        .success());
    let mut child = command(&f, f.project.path(), &["cc", "-a"])
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let start = Instant::now();
    while child.try_wait().unwrap().is_none() {
        // #2511: a hang guard, so the bound is generous.
        if start.elapsed() > Duration::from_secs(120) {
            let _ = child.kill();
            let _ = child.wait();
            panic!("launcher blocked on a FIFO");
        }
        std::thread::sleep(Duration::from_millis(50));
    }
    refused(&child.wait_with_output().unwrap());
    // The same for the counter lock itself.
    let f = fixture();
    fs::create_dir_all(sessions(&f)).unwrap();
    std::os::unix::fs::symlink(f.project.path().join("missing"), sessions(&f).join(".lock"))
        .unwrap();
    let err = refused(&run(&f, &["cc", "-a"]));
    assert!(err.contains("symlink"), "{err}");
}

#[test]
fn outside_a_git_project_the_first_launch_initializes_it_and_a_given_name_works() {
    // #2481 (user decision 6810818827): the folder becomes a git repository on the first
    // launch, so numbering starts in its new clone common dir (tests/launcher_auto_init.rs).
    let f = fixture();
    let plain = tempfile::tempdir().unwrap();
    assert_eq!(
        name(&run_in(&f, plain.path(), &["codex", "-a", "--dry-run"])),
        "agent-1"
    );
    assert!(!plain.path().join(".git").exists());
    assert_eq!(name(&run_in(&f, plain.path(), &["cc", "-a"])), "agent-1");
    assert!(plain
        .path()
        .join(".git/.moai-cli/sessions/agent-1.live")
        .is_file());
    assert_eq!(
        name(&run_in(&f, plain.path(), &["cc", "-a", "--name", "n1"])),
        "n1"
    );
    assert_eq!(name(&run_in(&f, plain.path(), &["cc", "-l"])), "leader");
}

fn launcher_config(f: &Fixture) -> PathBuf {
    f.home.path().join(".moai-cli/config/launcher.yaml")
}

fn assert_config_refusal_before_side_effects(f: &Fixture) {
    for host in ["cc", "glm", "codex"] {
        for role in [&["-l"][..], &["-a"][..], &[][..]] {
            for cli in [false, true] {
                let mut args = vec![host];
                args.extend_from_slice(role);
                if cli {
                    args.push("--permissions");
                }
                let out = run(f, &args);
                assert_eq!(out.status.code(), Some(1), "{args:?}");
                assert!(out.stdout.is_empty(), "{args:?}");
                assert!(!String::from_utf8_lossy(&out.stderr).contains("SECRET"));
                assert!(!f.home.path().join("host-started").exists(), "{args:?}");
                // No number is taken, and the old store stays as it was.
                assert_eq!(lives(f), Vec::<String>::new(), "{args:?}");
                assert_eq!(stored(f).as_deref(), Some("00041\n"), "{args:?}");
            }
        }
    }
}

#[test]
fn invalid_launcher_config_never_executes_host_or_allocates_number() {
    for bad in [
        r#"{"launcher":{"role_access":"SECRET-invalid-access"}}"#,
        r#"{"launcher":{"role_access":null}}"#,
        r#"{"launcher":{"role_access":false}}"#,
        r#"{"launcher":{"role_access":42}}"#,
        r#"{"launcher":{"role_access":[]}}"#,
        r#"{"launcher":{"role_access":{}}}"#,
        r#"{"launcher":null}"#,
        r#"{"launcher":[]}"#,
        r#"{"launcher":"SECRET-invalid-structure"}"#,
        "null",
        "[]",
        "42",
        "{ SECRET-malformed",
    ] {
        let f = fixture();
        seed(&f, "00041\n");
        fs::create_dir_all(launcher_config(&f).parent().unwrap()).unwrap();
        fs::write(launcher_config(&f), bad).unwrap();
        assert_config_refusal_before_side_effects(&f);
    }
}

#[test]
fn nonregular_launcher_config_never_executes_host_or_allocates_number() {
    let f = fixture();
    seed(&f, "00041\n");
    fs::create_dir_all(launcher_config(&f)).unwrap();
    assert_config_refusal_before_side_effects(&f);
}

#[test]
fn unreadable_launcher_config_never_executes_host_or_allocates_number() {
    let f = fixture();
    seed(&f, "00041\n");
    fs::create_dir_all(launcher_config(&f).parent().unwrap()).unwrap();
    fs::write(
        launcher_config(&f),
        r#"{"launcher":{"role_access":"prompt"}}"#,
    )
    .unwrap();
    fs::set_permissions(launcher_config(&f), fs::Permissions::from_mode(0o000)).unwrap();
    // Elevated CI users can read mode-000 files; exercise refusal only when unreadable.
    let unreadable = fs::File::open(launcher_config(&f)).is_err();
    if unreadable {
        assert_config_refusal_before_side_effects(&f);
    }
    fs::set_permissions(launcher_config(&f), fs::Permissions::from_mode(0o600)).unwrap();
}

#[test]
fn valid_launcher_settings_execute_fake_hosts_and_allocate_normally() {
    for setting in ["full", "prompt"] {
        let f = fixture();
        seed(&f, "00041\n");
        fs::create_dir_all(launcher_config(&f).parent().unwrap()).unwrap();
        fs::write(
            launcher_config(&f),
            format!(r#"{{"launcher":{{"role_access":"{setting}"}}}}"#),
        )
        .unwrap();
        // Each session ends before the next starts, so every launch is agent-1; the old store
        // is neither read nor changed.
        for host in ["cc", "glm", "codex"] {
            assert_eq!(name(&run(&f, &[host, "-a"])), "agent-1");
            assert_eq!(stored(&f).as_deref(), Some("00041\n"));
            assert_eq!(
                fs::read(f.home.path().join("host-started")).unwrap(),
                b"started"
            );
        }
    }
}

#[test]
fn symlink_launcher_config_is_refused_without_following_it() {
    let f = fixture();
    seed(&f, "00041\n");
    fs::create_dir_all(launcher_config(&f).parent().unwrap()).unwrap();
    let target = f.home.path().join("synthetic-config");
    let contents = r#"{"launcher":{"role_access":"full"}}"#;
    fs::write(&target, contents).unwrap();
    std::os::unix::fs::symlink(&target, launcher_config(&f)).unwrap();
    assert_config_refusal_before_side_effects(&f);
    assert_eq!(fs::read_to_string(target).unwrap(), contents);
}

/// #2361 AC-5 as amended by the 06:45Z user decision (`-l`, `-a` = full-access mode): a
/// launch with `--name X` or `-l` takes no number and starts the host with exactly this argv,
/// pinned byte for byte per host. Real launches against the stub hosts, not `--dry-run`.
#[test]
fn name_and_leader_launch_the_decided_full_access_argv_and_take_no_number() {
    let f = fixture();
    let claude = |name: &'static str| -> Vec<&'static str> {
        vec![
            "claude",
            "--name",
            name,
            "--model",
            "opus",
            "--effort",
            "medium",
            "--dangerously-skip-permissions",
        ]
    };
    let codex = |prompt: &'static str| -> Vec<&'static str> {
        vec!["codex", "-s", "danger-full-access", "-a", "never", prompt]
    };
    let cases: [(&[&str], Vec<&str>); 6] = [
        (&["cc", "-l"], claude("leader")),
        (&["glm", "-l"], claude("leader")),
        (&["codex", "-l"], codex("MoAI session name: leader")),
        (&["cc", "-a", "--name", "x"], claude("x")),
        (&["glm", "-a", "--name", "x"], claude("x")),
        (
            &["codex", "-a", "--name", "x"],
            codex("$moai --mode-agent\nMoAI session name: x"),
        ),
    ];
    for (args, want) in cases {
        let out = run(&f, args);
        assert!(
            out.status.success(),
            "{args:?}: {}",
            String::from_utf8_lossy(&out.stderr)
        );
        let v: Value = serde_json::from_slice(&out.stdout).unwrap();
        let argv: Vec<&str> = v["argv"]
            .as_array()
            .unwrap_or_else(|| panic!("{args:?}: no argv in {v}"))
            .iter()
            .map(|a| a.as_str().unwrap())
            .collect();
        assert_eq!(argv, want, "{args:?}");
    }
    assert!(!sessions(&f).exists(), "--name and -l take no number");
}
