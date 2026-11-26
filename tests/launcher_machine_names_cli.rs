//! SPEC-20261127-launcher-session-names: the launcher names a lane `agent-<name>` and the
//! coordination window `leader-<project name>`, and no other session on the machine holds the
//! same name. A name is claimed first in `<HOME>/.moai/session-names/<name>.live` (an exclusive
//! lock the OS drops when the launcher or its host ends), then in the clone's own `.live` lock.
//! Real launches run a stub `claude` executable from a temporary PATH in temporary git projects
//! that share one temporary HOME; no real host starts and the real HOME is never touched.
#![cfg(unix)]
mod common;
use serde_json::Value;
use std::io::Write;
use std::os::unix::fs::PermissionsExt;
use std::{
    fs,
    path::{Path, PathBuf},
    process::{Child, Command, Output, Stdio},
    time::{Duration, Instant},
};

/// A host stub that leaves a mark, and waits for a release file when the test names one, so a
/// test can look at the locks while the host runs.
const STUB: &str = "#!/bin/sh
printf started > \"$MOAI_TEST_HOST_MARKER\"
if [ -n \"$MOAI_TEST_HOST_RELEASE\" ]; then
  n=0
  while [ ! -e \"$MOAI_TEST_HOST_RELEASE\" ] && [ \"$n\" -lt 1200 ]; do
    sleep 0.1
    n=$((n + 1))
  done
fi
exit 0
";

/// One temporary HOME and one host stub shared by every project of a test.
struct Env {
    home: tempfile::TempDir,
    bin: tempfile::TempDir,
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

fn env() -> Env {
    let bin = tempfile::tempdir().unwrap();
    let path = bin.path().join("claude");
    fs::write(&path, STUB).unwrap();
    fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
    Env {
        home: tempfile::tempdir().unwrap(),
        bin,
    }
}

/// An initialized git project (settings present, so no first-launch initialization runs).
fn project() -> tempfile::TempDir {
    let project = tempfile::tempdir().unwrap();
    git(project.path(), &["init", "-q"]);
    let settings = moai_cli::state_root::path(project.path(), "project.yaml");
    fs::create_dir_all(settings.parent().unwrap()).unwrap();
    fs::write(settings, "{}\n").unwrap();
    project
}

fn command(e: &Env, p: &Path, args: &[&str]) -> Command {
    let mut c = Command::new(crate::common::MOAI_BIN);
    c.args(args)
        .current_dir(p)
        .env("PATH", format!("{}:/usr/bin:/bin", e.bin.path().display()))
        .env("HOME", e.home.path())
        .env("USERPROFILE", e.home.path())
        .env("MOAI_TEST_HOST_MARKER", p.join("host-started"))
        .env_remove("MOAI_TEST_HOST_RELEASE")
        .env("ANTHROPIC_AUTH_TOKEN", "moai-test-dummy-not-real")
        .env_remove("CLAUDE_CODE_DISABLE_ADVISOR_TOOL");
    c
}

fn run(e: &Env, p: &Path, args: &[&str]) -> Output {
    command(e, p, args).output().unwrap()
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

/// A refused launch: exit 1, nothing on stdout, a message on stderr (returned).
fn refused(out: &Output) -> String {
    let stderr = String::from_utf8_lossy(&out.stderr).to_string();
    assert_eq!(out.status.code(), Some(1), "{stderr}");
    assert!(
        out.stdout.is_empty(),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    assert!(!stderr.is_empty());
    stderr
}

fn started(p: &Path) -> bool {
    p.join("host-started").exists()
}

fn machine_dir(e: &Env) -> PathBuf {
    moai_cli::home_workdir::root(e.home.path()).join("session-names")
}

fn machine_live(e: &Env, name: &str) -> PathBuf {
    machine_dir(e).join(format!("{name}.live"))
}

fn clone_sessions(p: &Path) -> PathBuf {
    moai_cli::clone_root::path(&p.join(".git")).join("sessions")
}

fn leader_name(p: &Path) -> String {
    format!("leader-{}", moai_cli::session::project_name(p))
}

/// What another launcher does: an exclusive lock on the machine-wide file of `name`, with a body
/// that says which project holds it.
fn hold_machine(e: &Env, name: &str, project: &str) -> fs::File {
    fs::create_dir_all(machine_dir(e)).unwrap();
    let f = fs::OpenOptions::new()
        .read(true)
        .write(true)
        .create(true)
        .truncate(false)
        .open(machine_live(e, name))
        .unwrap();
    f.lock().unwrap();
    (&f).write_all(format!("project={project}\n").as_bytes())
        .unwrap();
    f
}

/// Holds every name of the launcher's list except `except`, as another project's launchers.
fn hold_all_but(e: &Env, except: Option<&str>) -> Vec<fs::File> {
    moai_cli::session::AGENT_NAMES
        .iter()
        .filter(|n| Some(**n) != except)
        .map(|n| hold_machine(e, &format!("agent-{n}"), "other-project"))
        .collect()
}

/// Whether the lock of the file at `path` is held by someone else right now.
fn lock_held(path: &Path) -> bool {
    let f = fs::OpenOptions::new()
        .read(true)
        .write(true)
        .open(path)
        .unwrap();
    matches!(f.try_lock(), Err(fs::TryLockError::WouldBlock))
}

/// A bounded wait for a path to appear.
fn wait_for(path: &Path) {
    let start = Instant::now();
    while !path.exists() {
        // A hang guard, so the bound is generous.
        assert!(
            start.elapsed() < Duration::from_secs(60),
            "{} never appeared",
            path.display()
        );
        std::thread::sleep(Duration::from_millis(50));
    }
}

fn wait_exit(mut child: Child) -> Output {
    let start = Instant::now();
    while child.try_wait().unwrap().is_none() {
        if start.elapsed() > Duration::from_secs(120) {
            let _ = child.kill();
            let _ = child.wait();
            panic!("launcher did not exit");
        }
        std::thread::sleep(Duration::from_millis(50));
    }
    child.wait_with_output().unwrap()
}

/// AC-1 (REQ-1, REQ-3, REQ-5): another project's launchers hold 47 of the 48 names, so a lane of
/// this project can only be the free one, and a held `leader-<project>` refuses the second `-l`.
#[test]
fn collision_two_projects_never_share_a_name() {
    let e = env();
    let a = project();
    let b = project();
    let free = moai_cli::session::AGENT_NAMES[20];
    let held = hold_all_but(&e, Some(free));
    // Project A's launcher, `-a`: the only name left is the free one, whatever the seed.
    assert_eq!(
        name(&run(&e, a.path(), &["cc", "-a"])),
        format!("agent-{free}")
    );
    assert!(started(a.path()));
    // And the same for project B: the host of A has exited, so the name is claimable again.
    assert_eq!(
        name(&run(&e, b.path(), &["glm", "-a"])),
        format!("agent-{free}")
    );
    assert!(machine_live(&e, &format!("agent-{free}")).is_file());
    drop(held);
    // The leader of a project name another launcher holds is refused, with a hint to `--name`.
    let leader = leader_name(b.path());
    let other = hold_machine(&e, &leader, "some-other-project");
    let _ = fs::remove_file(b.path().join("host-started"));
    let err = refused(&run(&e, b.path(), &["cc", "-l"]));
    assert!(err.contains("some-other-project"), "{err}");
    assert!(err.contains("--name"), "{err}");
    assert!(!started(b.path()));
    drop(other);
    assert_eq!(name(&run(&e, b.path(), &["cc", "-l"])), leader);
}

/// AC-2 (REQ-2): with all 48 names held machine-wide, `-a` starts no host and names `--name`.
#[test]
fn used_up_every_name_held_machine_wide_refuses_a_lane() {
    let e = env();
    let p = project();
    let _held = hold_all_but(&e, None);
    let err = refused(&run(&e, p.path(), &["cc", "-a"]));
    assert!(err.contains("--name"), "{err}");
    assert!(!started(p.path()));
}

/// AC-3 (REQ-4, REQ-6): while the host runs both locks are held and the machine-wide file says
/// who holds it; when the host exits both are free.
#[test]
fn claim_holds_both_locks_while_the_host_runs_and_frees_them_after() {
    let e = env();
    let p = project();
    let leader = leader_name(p.path());
    let release = p.path().join("release");
    let child = command(&e, p.path(), &["cc", "-l"])
        .env("MOAI_TEST_HOST_RELEASE", &release)
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let pid = child.id().to_string();
    wait_for(&p.path().join("host-started"));
    let machine = machine_live(&e, &leader);
    let clone = clone_sessions(p.path()).join(format!("{leader}.live"));
    assert!(lock_held(&machine), "the machine-wide lock is held");
    assert!(lock_held(&clone), "the clone's lock is held");
    let body = fs::read_to_string(&machine).unwrap();
    assert!(
        body.contains(&moai_cli::session::project_name(p.path())),
        "{body}"
    );
    assert!(body.contains(&leader), "{body}");
    assert!(body.contains(&pid), "{body}");
    fs::write(&release, "go").unwrap();
    assert_eq!(name(&wait_exit(child)), leader);
    assert!(!lock_held(&machine), "the machine-wide lock is free again");
    assert!(!lock_held(&clone), "the clone's lock is free again");
}

/// AC-3 (REQ-4): when the clone's own lock cannot be taken, the machine-wide lock is let go.
#[test]
fn claim_clone_lock_refusal_frees_the_machine_lock() {
    let e = env();
    let p = project();
    let leader = leader_name(p.path());
    fs::create_dir_all(clone_sessions(p.path())).unwrap();
    let clone = fs::OpenOptions::new()
        .read(true)
        .write(true)
        .create(true)
        .truncate(false)
        .open(clone_sessions(p.path()).join(format!("{leader}.live")))
        .unwrap();
    clone.lock().unwrap();
    let err = refused(&run(&e, p.path(), &["cc", "-l"]));
    assert!(err.contains("already running"), "{err}");
    assert!(!started(p.path()));
    let machine = machine_live(&e, &leader);
    assert!(!machine.exists() || !lock_held(&machine));
}

/// AC-4 (REQ-7): a link or a missing HOME is refused with no host, and nothing is written
/// through the link.
#[test]
fn refuse_a_link_or_an_unresolved_home_starts_no_host() {
    // The folder `<HOME>/.moai/session-names` is a link.
    let e = env();
    let p = project();
    let elsewhere = e.home.path().join("elsewhere");
    fs::create_dir_all(&elsewhere).unwrap();
    fs::create_dir_all(machine_dir(&e).parent().unwrap()).unwrap();
    std::os::unix::fs::symlink(&elsewhere, machine_dir(&e)).unwrap();
    let err = refused(&run(&e, p.path(), &["cc", "-l"]));
    assert!(err.contains("link") || err.contains("symlink"), "{err}");
    assert!(!started(p.path()));
    assert_eq!(fs::read_dir(&elsewhere).unwrap().count(), 0);
    // The lock file of the name is a link.
    let e = env();
    let p = project();
    let real = e.home.path().join("real-file");
    fs::write(&real, "3\n").unwrap();
    fs::create_dir_all(machine_dir(&e)).unwrap();
    std::os::unix::fs::symlink(&real, machine_live(&e, &leader_name(p.path()))).unwrap();
    refused(&run(&e, p.path(), &["cc", "-l"]));
    assert_eq!(fs::read_to_string(&real).unwrap(), "3\n");
    assert!(!started(p.path()));
    // The lock file of the name is a directory.
    let e = env();
    let p = project();
    fs::create_dir_all(machine_live(&e, &leader_name(p.path()))).unwrap();
    refused(&run(&e, p.path(), &["cc", "-l"]));
    assert!(!started(p.path()));
    // No home folder can be resolved.
    let e = env();
    let p = project();
    let out = command(&e, p.path(), &["cc", "-l"])
        .env_remove("HOME")
        .env_remove("USERPROFILE")
        .output()
        .unwrap();
    refused(&out);
    assert!(!started(p.path()));
}

/// AC-4 (REQ-8): a dry run only looks. It names the session it would start, creates no folder
/// or file, and refuses what the real launch would refuse.
#[test]
fn dry_run_only_probes_and_creates_nothing() {
    let e = env();
    let p = project();
    let lane = name(&run(&e, p.path(), &["cc", "-a", "--dry-run"]));
    assert!(lane.starts_with("agent-"), "{lane}");
    assert!(moai_cli::session::AGENT_NAMES.contains(&&lane["agent-".len()..]));
    assert_eq!(
        name(&run(&e, p.path(), &["cc", "-l", "--dry-run"])),
        leader_name(p.path())
    );
    assert_eq!(fs::read_dir(e.home.path()).unwrap().count(), 0);
    assert!(!clone_sessions(p.path()).exists());
    // With names held, the dry run names the free one and leaves the folder as it was.
    let free = moai_cli::session::AGENT_NAMES[7];
    let _held = hold_all_but(&e, Some(free));
    let before = fs::read_dir(machine_dir(&e)).unwrap().count();
    assert_eq!(
        name(&run(&e, p.path(), &["cc", "-a", "--dry-run"])),
        format!("agent-{free}")
    );
    assert_eq!(fs::read_dir(machine_dir(&e)).unwrap().count(), before);
    // A held leader name is refused in the preview, as in the real launch.
    let _leader = hold_machine(&e, &leader_name(p.path()), "some-other-project");
    refused(&run(&e, p.path(), &["cc", "-l", "--dry-run"]));
    assert!(!started(p.path()));
}

/// AC-5 (REQ-5): a `--name` beside a role flag is claimed machine-wide and refused when another
/// project holds it; a `--name` alone takes no machine-wide lock and runs as before.
#[test]
fn named_beside_a_role_flag_is_claimed_and_alone_is_not() {
    let e = env();
    let p = project();
    let _held = hold_machine(&e, "agent-ben", "other-project");
    let err = refused(&run(&e, p.path(), &["cc", "-a", "--name", "agent-ben"]));
    assert!(err.contains("other-project"), "{err}");
    assert!(err.contains("--name"), "{err}");
    assert!(!started(p.path()));
    // `-l --name` is claimed the same way.
    let err = refused(&run(&e, p.path(), &["cc", "-l", "--name", "agent-ben"]));
    assert!(err.contains("other-project"), "{err}");
    // A plain `--name` runs as before and leaves no machine-wide file for it.
    let e2 = env();
    let q = project();
    assert_eq!(
        name(&run(&e2, q.path(), &["cc", "--name", "agent-ben"])),
        "agent-ben"
    );
    assert!(!machine_dir(&e2).exists());
}
