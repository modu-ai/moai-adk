//! C-105 slice 2 (SPEC-20261119-plan-remote-branch): `moai project plan` checks the
//! profile's `work_flow.integration_branch` on the project's remote — existence, the
//! behind count against the remote's default branch, and a refusal for names that must
//! never reach git as an argument. Temporary HOME, temporary repository, bare remote,
//! fake `git` logger for the injection probe; never the network or the real HOME.

mod common;

use common::MOAI_BIN;
use serde_json::{json, Value};
use std::fs;
use std::io::Write;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use std::sync::atomic::{AtomicUsize, Ordering};
use std::time::{Duration, Instant};

static NEXT: AtomicUsize = AtomicUsize::new(0);

fn git(dir: &Path, args: &[&str]) {
    // Audit B3: every fixture git call goes through the shared guard (gc.auto=0,
    // maintenance off, isolated config, LC_ALL=C) — no bare `git` spawns here.
    let ok = common::git::command()
        .arg("-C")
        .arg(dir)
        .args(args)
        .output()
        .expect("git");
    assert!(
        ok.status.success(),
        "git {args:?}: {}",
        String::from_utf8_lossy(&ok.stderr)
    );
}

/// A project repository on `main` with a bare remote `origin`, and a HOME for the run.
struct World {
    home: PathBuf,
    project: PathBuf,
    remote: PathBuf,
}

impl World {
    fn new() -> Self {
        let n = NEXT.fetch_add(1, Ordering::SeqCst);
        let base = std::env::temp_dir().join(format!("c105-plan-{}", std::process::id()));
        let root = base.join(format!("w{n}"));
        fs::create_dir_all(&root).unwrap();
        let home = root.join("home");
        let project = root.join("demo");
        let remote = root.join("remote.git");
        fs::create_dir_all(&home).unwrap();
        let w = World {
            home,
            project,
            remote,
        };
        let init_ok = common::git::command()
            .arg("init")
            .arg("-q")
            .arg("-b")
            .arg("main")
            .arg(&w.project)
            .output()
            .expect("git init");
        assert!(init_ok.status.success(), "git init");
        let bare_ok = common::git::command()
            .arg("init")
            .arg("-q")
            .arg("--bare")
            .arg("-b")
            .arg("main")
            .arg(&w.remote)
            .output()
            .expect("git init --bare");
        assert!(bare_ok.status.success(), "git init --bare");
        git(
            &w.project,
            &["remote", "add", "origin", w.remote.to_str().unwrap()],
        );
        fs::write(w.project.join("seed.txt"), "base\n").unwrap();
        git(&w.project, &["add", "-A"]);
        commit(&w.project, "base");
        git(&w.project, &["push", "-q", "origin", "main"]);
        w
    }

    /// Advance `main` by `n` commits on the remote.
    fn advance_main(&self, n: usize) {
        for i in 0..n {
            fs::write(self.project.join(format!("m{i}.txt")), "m\n").unwrap();
            git(&self.project, &["add", "-A"]);
            commit(&self.project, &format!("main {i}"));
        }
        git(&self.project, &["push", "-q", "origin", "main"]);
    }

    /// Push an integration branch at the current `main`.
    fn push_branch(&self, name: &str) {
        git(&self.project, &["branch", name]);
        git(&self.project, &["push", "-q", "origin", name]);
    }

    fn plan(&self, integration_branch: Option<&str>, git_log: Option<&Path>) -> Value {
        let mut profile = json!({
            "schema_version": 2, "project_id": "demo", "purpose": "Fixture",
            "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"],
            "decisions": [{"id": "d1", "topic": "tools", "value": "Rust",
                "source": "user", "scope": "project"}],
        });
        if let Some(b) = integration_branch {
            profile["work_flow"] = json!({"flow": "m-flow", "integration_branch": b});
        }
        let mut cmd = Command::new(MOAI_BIN);
        cmd.arg("project-plan")
            .arg(self.project.to_str().unwrap())
            .env("HOME", &self.home)
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped());
        if let Some(log) = git_log {
            cmd.env(
                "PATH",
                format!(
                    "{}:{}",
                    log.parent().unwrap().display(),
                    std::env::var("PATH").unwrap()
                ),
            );
            cmd.env("C105_GIT_LOG", log);
        }
        let mut child = cmd.spawn().unwrap();
        child
            .stdin
            .take()
            .unwrap()
            .write_all(profile.to_string().as_bytes())
            .unwrap();
        let out = child.wait_with_output().unwrap();
        assert!(
            out.status.success(),
            "plan failed: {} {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        );
        serde_json::from_slice(&out.stdout).unwrap()
    }
}

fn commit(dir: &Path, message: &str) {
    let ok = common::git::command()
        .arg("-C")
        .arg(dir)
        .args([
            "-c",
            "user.name=fixture",
            "-c",
            "user.email=fixture@example.invalid",
            "-c",
            "commit.gpgsign=false",
            "commit",
            "-q",
            "-m",
            message,
        ])
        .output()
        .unwrap();
    assert!(ok.status.success());
}

/// A fake `git` that logs its arguments (C105_GIT_LOG) and forwards to the real one.
fn make_fake_git(dir: &Path) {
    fs::write(
        dir.join("git"),
        "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$C105_GIT_LOG\"\nexec git \"$@\"\n",
    )
    .unwrap();
    make_executable(&dir.join("git"));
}

#[cfg(unix)]
fn make_executable(path: &Path) {
    use std::os::unix::fs::PermissionsExt;
    fs::set_permissions(path, fs::Permissions::from_mode(0o755)).unwrap();
}

#[cfg(not(unix))]
fn make_executable(_path: &Path) {}

/// The absolute path of the real `git` on this lane's PATH. A fake on PATH forwards to
/// it by path, so a regression can never exec the fake through PATH again (audit N2:
/// that self-exec loop hangs the suite instead of failing a test).
fn real_git() -> String {
    let out = Command::new("which")
        .arg("git")
        .output()
        .expect("which git");
    assert!(out.status.success(), "which git found nothing");
    String::from_utf8_lossy(&out.stdout).trim().to_string()
}

/// A fake `git` that records the `GIT_TERMINAL_PROMPT` value it was handed, then
/// forwards to the real git by absolute path.
fn make_env_git(dir: &Path) {
    let real = real_git();
    fs::write(
        dir.join("git"),
        format!(
            "#!/bin/sh\nprintf 'GIT_TERMINAL_PROMPT=%s\\n' \"$GIT_TERMINAL_PROMPT\" >> \"$C105_GIT_LOG\"\nexec \"{real}\" \"$@\"\n"
        ),
    )
    .unwrap();
    make_executable(&dir.join("git"));
}

/// A fake `git` that never answers `ls-remote` (it sleeps instead) and forwards every
/// other call to the real git by absolute path.
fn make_hanging_git(dir: &Path) {
    let real = real_git();
    fs::write(
        dir.join("git"),
        format!(
            "#!/bin/sh\ncase \" $* \" in *' ls-remote '*) sleep 120 ;; *) exec \"{real}\" \"$@\" ;; esac\n"
        ),
    )
    .unwrap();
    make_executable(&dir.join("git"));
}

#[test]
fn a_branch_on_the_remote_is_reported_present_without_warnings() {
    let w = World::new();
    w.push_branch("m-flow/integration");
    let v = w.plan(Some("m-flow/integration"), None);
    let wf = v["result"]["work_flow_remote"]
        .as_object()
        .expect("work_flow_remote");
    assert_eq!(wf["on_remote"], true, "{wf:?}");
    assert_eq!(
        wf.get("warnings").map(|x| x.as_array().unwrap().len()),
        Some(0),
        "{wf:?}"
    );
}

#[test]
fn a_branch_missing_on_the_remote_warns_and_names_it() {
    let w = World::new();
    let v = w.plan(Some("nope/absent"), None);
    let wf = v["result"]["work_flow_remote"]
        .as_object()
        .expect("work_flow_remote");
    assert_eq!(wf["on_remote"], false, "{wf:?}");
    let warnings = serde_json::to_string(&wf["warnings"]).unwrap();
    assert!(warnings.contains("nope/absent"), "{warnings}");
    assert!(warnings.to_lowercase().contains("remote"), "{warnings}");
}

#[test]
fn a_branch_behind_the_remote_default_warns_with_the_count() {
    let w = World::new();
    w.push_branch("m-flow/integration");
    w.advance_main(2);
    let v = w.plan(Some("m-flow/integration"), None);
    let wf = v["result"]["work_flow_remote"]
        .as_object()
        .expect("work_flow_remote");
    let warnings = serde_json::to_string(&wf["warnings"]).unwrap();
    assert!(warnings.contains('2'), "{warnings}");
    assert!(warnings.contains("behind"), "{warnings}");
}

#[test]
fn dangerous_branch_names_never_reach_git_as_arguments() {
    let w = World::new();
    let logdir = w.home.join("fakebin");
    fs::create_dir_all(&logdir).unwrap();
    let log = logdir.join("git.log");
    make_fake_git(&logdir);
    for bad in ["-x", "a b", "a..b", "HEAD", &"l".repeat(101)] {
        fs::write(&log, b"").unwrap();
        // The profile validation refuses the name before any git call, so the plan
        // fails; the injected fake git proves nothing ever reached git.
        let mut cmd = Command::new(MOAI_BIN);
        cmd.arg("project-plan")
            .arg(w.project.to_str().unwrap())
            .env("HOME", &w.home)
            .env(
                "PATH",
                format!("{}:{}", logdir.display(), std::env::var("PATH").unwrap()),
            )
            .env("C105_GIT_LOG", &log)
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped());
        let profile = json!({
            "schema_version": 2, "project_id": "demo", "purpose": "Fixture",
            "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"],
            "decisions": [{"id": "d1", "topic": "tools", "value": "Rust",
                "source": "user", "scope": "project"}],
            "work_flow": {"flow": "m-flow", "integration_branch": bad},
        });
        let mut child = cmd.spawn().unwrap();
        child
            .stdin
            .take()
            .unwrap()
            .write_all(profile.to_string().as_bytes())
            .unwrap();
        let out = child.wait_with_output().unwrap();
        assert!(
            !out.status.success(),
            "{bad}: the plan should refuse a dangerous name, got {}",
            String::from_utf8_lossy(&out.stdout)
        );
        let stderr = String::from_utf8_lossy(&out.stderr);
        let _ = &profile;
        let calls = fs::read_to_string(&log).unwrap();
        assert!(
            !calls.contains("ls-remote"),
            "a git call ran for {bad}: {calls}"
        );
        assert!(!calls.contains(bad), "dangerous name reached git: {calls}");
        assert!(
            stderr.contains("invalid") || stderr.contains("refused"),
            "{bad}: {stderr}"
        );
    }
}

#[test]
fn without_work_flow_or_without_a_remote_the_key_is_absent() {
    let w = World::new();
    let v = w.plan(None, None);
    assert!(
        v["result"].get("work_flow_remote").is_none(),
        "{}",
        v["result"]
    );

    // A repository with no remote at all: the key stays absent too. The init goes
    // through the shared guarded helper (audit N1: this bare spawn had no fixture
    // guards and no exit-status check; the bundle's command lint caught it).
    let n = NEXT.fetch_add(1, Ordering::SeqCst);
    let root =
        std::env::temp_dir().join(format!("c105-plan-noremote-{}-{}", std::process::id(), n));
    fs::create_dir_all(root.join("home")).unwrap();
    common::git::git_ok(
        &root,
        &[
            "init",
            "-q",
            "-b",
            "main",
            root.join("demo").to_str().unwrap(),
        ],
    );
    let w2 = World {
        home: root.join("home"),
        project: root.join("demo"),
        remote: root.join("x"),
    };
    let v = w2.plan(Some("m-flow/integration"), None);
    assert!(
        v["result"].get("work_flow_remote").is_none(),
        "{}",
        v["result"]
    );
}

#[test]
fn a_stale_clone_warns_unavailable_instead_of_reading_healthy() {
    // Audit B1 (P6): the integration branch and the remote default exist on the
    // remote, but a second clone made those commits — this repository never
    // fetched them, so the behind measurement cannot run. It must warn, never
    // report a healthy branch.
    let w = World::new();
    w.push_branch("m-flow/integration");
    // A second clone advances the remote by two commits.
    let second = w.home.join("second");
    let clone_ok = common::git::command()
        .args([
            "clone",
            "-q",
            w.remote.to_str().unwrap(),
            second.to_str().unwrap(),
        ])
        .output()
        .expect("git clone");
    assert!(clone_ok.status.success(), "git clone");
    fs::write(second.join("late.txt"), "late\n").unwrap();
    git(&second, &["add", "-A"]);
    commit(&second, "late work");
    git(&second, &["push", "-q", "origin", "main"]);
    let v = w.plan(Some("m-flow/integration"), None);
    let wf = v["result"]["work_flow_remote"]
        .as_object()
        .expect("work_flow_remote");
    assert_eq!(wf["on_remote"], true, "{wf:?}");
    let warnings = serde_json::to_string(&wf["warnings"]).unwrap();
    assert!(
        warnings.to_lowercase().contains("unavailable"),
        "a stale clone must not read healthy: {warnings}"
    );
}

#[test]
fn a_failing_remote_check_warns_unavailable_and_the_plan_succeeds() {
    let mut w = World::new();
    // Point the remote at a path that cannot answer.
    let broken = w.home.join("broken.git");
    fs::write(&broken, b"not a repo").unwrap();
    git(
        &w.project,
        &["remote", "set-url", "origin", broken.to_str().unwrap()],
    );
    let v = w.plan(Some("m-flow/integration"), None);
    let wf = v["result"]["work_flow_remote"]
        .as_object()
        .expect("work_flow_remote");
    let warnings = serde_json::to_string(&wf["warnings"]).unwrap();
    assert!(
        warnings.to_lowercase().contains("unavailable"),
        "{warnings}"
    );
    let _ = &mut w; // keep the world alive to the end
}

#[test]
fn the_remote_check_never_lets_git_prompt_on_the_terminal() {
    // Audit B6: the git closure must hand git GIT_TERMINAL_PROMPT=0, so a credential
    // ask fails instead of blocking the plan on a hidden terminal prompt. The env
    // probe records the value the child actually received.
    let w = World::new();
    w.push_branch("m-flow/integration");
    let logdir = w.home.join("fakebin");
    fs::create_dir_all(&logdir).unwrap();
    let log = logdir.join("git.log");
    make_env_git(&logdir);
    let v = w.plan(Some("m-flow/integration"), Some(&log));
    let wf = v["result"]["work_flow_remote"]
        .as_object()
        .expect("work_flow_remote");
    assert_eq!(wf["on_remote"], true, "{wf:?}");
    let calls = fs::read_to_string(&log).unwrap();
    assert!(
        calls.contains("GIT_TERMINAL_PROMPT=0"),
        "git ran without GIT_TERMINAL_PROMPT=0: {calls}"
    );
}

#[test]
fn a_hanging_remote_check_is_killed_and_the_plan_still_succeeds() {
    // Audit B6: a remote that never answers must end in a bounded wait — the child's
    // process group is killed and the plan answers `unavailable … timed out` instead of
    // hanging. The test keeps its own watchdog, so a regression fails, never hangs.
    let w = World::new();
    w.push_branch("m-flow/integration");
    let logdir = w.home.join("fakebin");
    fs::create_dir_all(&logdir).unwrap();
    make_hanging_git(&logdir);
    let profile = json!({
        "schema_version": 2, "project_id": "demo", "purpose": "Fixture",
        "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"],
        "decisions": [{"id": "d1", "topic": "tools", "value": "Rust",
            "source": "user", "scope": "project"}],
        "work_flow": {"flow": "m-flow", "integration_branch": "m-flow/integration"},
    });
    let mut cmd = Command::new(MOAI_BIN);
    cmd.arg("project-plan")
        .arg(w.project.to_str().unwrap())
        .env("HOME", &w.home)
        .env("MOAI_GIT_TIMEOUT_SECS", "2")
        .env(
            "PATH",
            format!("{}:{}", logdir.display(), std::env::var("PATH").unwrap()),
        )
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());
    let mut child = cmd.spawn().unwrap();
    child
        .stdin
        .take()
        .unwrap()
        .write_all(profile.to_string().as_bytes())
        .unwrap();
    let deadline = Instant::now() + Duration::from_secs(30);
    let finished = loop {
        match child.try_wait().expect("try_wait") {
            Some(_) => break true,
            None if Instant::now() >= deadline => break false,
            None => std::thread::sleep(Duration::from_millis(50)),
        }
    };
    if !finished {
        let _ = child.kill();
        let _ = child.wait();
    }
    assert!(
        finished,
        "the plan must not wait forever on a hanging remote"
    );
    let out = child.wait_with_output().unwrap();
    assert!(
        out.status.success(),
        "plan failed: {} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    let wf = v["result"]["work_flow_remote"]
        .as_object()
        .expect("work_flow_remote");
    let warnings = serde_json::to_string(&wf["warnings"]).unwrap();
    assert!(warnings.contains("unavailable"), "{warnings}");
    assert!(warnings.contains("timed out"), "{warnings}");
}
