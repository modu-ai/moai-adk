//! #2592 S-C: `moai-cli session-register` on synthetic repositories in a temporary folder with
//! a temporary HOME only. No real registry of this machine is read or written.
#![cfg(unix)]
mod common;

use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output};

const BIN: &str = crate::common::MOAI_BIN;
const T0: &str = "2026-11-13T10:00:00Z";

fn git(dir: &Path, args: &[&str]) -> String {
    let mut all = vec![
        "-c",
        "user.name=fixture",
        "-c",
        "user.email=fixture@example.invalid",
    ];
    all.extend_from_slice(args);
    common::git::git(dir, &all)
}

struct Fixture {
    _dir: tempfile::TempDir,
    repo: PathBuf,
    home: PathBuf,
}

impl Fixture {
    fn new() -> Self {
        let dir = tempfile::tempdir().unwrap();
        let base = fs::canonicalize(dir.path()).unwrap();
        let repo = base.join("repo");
        let home = base.join("home");
        fs::create_dir_all(&repo).unwrap();
        fs::create_dir_all(&home).unwrap();
        git(&repo, &["init", "-q", "-b", "main"]);
        git(&repo, &["commit", "-q", "--allow-empty", "-m", "base"]);
        Fixture {
            _dir: dir,
            repo,
            home,
        }
    }
    fn sessions(&self) -> PathBuf {
        self.repo.join(".git/.moai-cli/sessions")
    }
    fn entry(&self, name: &str) -> PathBuf {
        self.sessions().join(format!("{name}.json"))
    }
}

fn run_in(fx: &Fixture, cwd: &Path, args: &[&str]) -> Output {
    Command::new(BIN)
        .arg("session-register")
        .args(args)
        .current_dir(cwd)
        .env("HOME", &fx.home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .output()
        .unwrap()
}

fn run(fx: &Fixture, args: &[&str]) -> Output {
    let repo = fx.repo.to_str().unwrap().to_string();
    let mut all = vec![repo.as_str()];
    all.extend_from_slice(args);
    run_in(fx, &fx.repo, &all)
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "envelope on stdout; stderr {}",
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

fn ok(out: &Output) -> Value {
    let v = envelope(out);
    assert_eq!(out.status.code(), Some(0), "{v}");
    assert_eq!(v["ok"], true, "{v}");
    assert_eq!(v["operation"], "session.register");
    v["result"].clone()
}

/// The error kind of a refused call, which also must exit 1.
fn refused(out: &Output) -> (String, String) {
    let v = envelope(out);
    assert_eq!(out.status.code(), Some(1), "{v}");
    assert_eq!(v["ok"], false, "{v}");
    (
        v["error"]["kind"].as_str().unwrap_or_default().to_string(),
        v["error"]["message"]
            .as_str()
            .unwrap_or_default()
            .to_string(),
    )
}

#[test]
fn a_name_is_registered_once_and_refused_to_another_instance() {
    let fx = Fixture::new();
    let first = ok(&run(&fx, &["--name", "agent-8", "--now", T0]));
    assert_eq!(first["status"], "registered");
    assert_eq!(first["previous"], "none");
    assert_eq!(first["lease_until"], "2026-11-13T11:00:00Z");
    let instance = first["instance"].as_str().unwrap().to_string();
    assert_eq!(instance.len(), 32);
    let stored: Value = serde_json::from_slice(&fs::read(fx.entry("agent-8")).unwrap()).unwrap();
    assert_eq!(stored["instance"], instance.as_str());
    let before = fs::read(fx.entry("agent-8")).unwrap();

    // A second session under the same name, without the instance, is refused and writes nothing.
    let (kind, message) = refused(&run(&fx, &["--name", "agent-8", "--now", T0]));
    assert_eq!(kind, "conflict");
    assert!(message.contains("until 2026-11-13T11:00:00Z"), "{message}");
    assert!(
        !message.contains(&instance),
        "the holder's instance is not shown"
    );
    assert_eq!(fs::read(fx.entry("agent-8")).unwrap(), before);

    // Another made-up instance is refused the same way.
    let other = "0".repeat(32);
    let (kind, _) = refused(&run(
        &fx,
        &["--name", "agent-8", "--instance", &other, "--now", T0],
    ));
    assert_eq!(kind, "conflict");

    // Another name is independent.
    assert_eq!(
        ok(&run(&fx, &["--name", "agent-9", "--now", T0]))["status"],
        "registered"
    );
}

#[test]
fn the_holder_renews_with_its_instance() {
    let fx = Fixture::new();
    let first = ok(&run(&fx, &["--name", "agent-8", "--now", T0]));
    let instance = first["instance"].as_str().unwrap();
    let renewed = ok(&run(
        &fx,
        &[
            "--name",
            "agent-8",
            "--instance",
            instance,
            "--now",
            "2026-11-13T10:30:00Z",
        ],
    ));
    assert_eq!(renewed["status"], "renewed");
    assert_eq!(renewed["previous"], "same_instance");
    assert_eq!(renewed["instance"], instance);
    assert_eq!(renewed["lease_until"], "2026-11-13T11:30:00Z");
}

/// Boundary: one second before the lease ends the name is still held; at the end it is free.
#[test]
fn a_lease_holds_until_its_end_second() {
    let fx = Fixture::new();
    ok(&run(&fx, &["--name", "agent-8", "--now", T0]));
    let (kind, _) = refused(&run(
        &fx,
        &["--name", "agent-8", "--now", "2026-11-13T10:59:59Z"],
    ));
    assert_eq!(kind, "conflict");
    let taken = ok(&run(
        &fx,
        &["--name", "agent-8", "--now", "2026-11-13T11:00:00Z"],
    ));
    assert_eq!(taken["status"], "registered");
    assert_eq!(taken["previous"], "expired");
}

#[test]
fn every_worktree_of_the_clone_sees_one_registry() {
    let fx = Fixture::new();
    let wt = fx.repo.join(".moai-cli/worktrees/lane");
    git(
        &fx.repo,
        &["worktree", "add", "-q", "--detach", wt.to_str().unwrap()],
    );
    ok(&run(&fx, &["--name", "agent-8", "--now", T0]));
    let (kind, _) = refused(&run_in(
        &fx,
        &wt,
        &[wt.to_str().unwrap(), "--name", "agent-8", "--now", T0],
    ));
    assert_eq!(kind, "conflict");
    assert!(!wt.join(".moai-cli/sessions").exists());
}

#[test]
fn only_the_holder_releases_and_then_the_name_is_free() {
    let fx = Fixture::new();
    let first = ok(&run(&fx, &["--name", "agent-8", "--now", T0]));
    let instance = first["instance"].as_str().unwrap();
    let other = "f".repeat(32);
    let (kind, _) = refused(&run(
        &fx,
        &[
            "--name",
            "agent-8",
            "--instance",
            &other,
            "--release",
            "--now",
            T0,
        ],
    ));
    assert_eq!(kind, "conflict");
    assert!(fx.entry("agent-8").exists());
    let released = ok(&run(
        &fx,
        &[
            "--name",
            "agent-8",
            "--instance",
            instance,
            "--release",
            "--now",
            T0,
        ],
    ));
    assert_eq!(released["status"], "released");
    assert!(!fx.entry("agent-8").exists());
    let again = ok(&run(
        &fx,
        &[
            "--name",
            "agent-8",
            "--instance",
            instance,
            "--release",
            "--now",
            T0,
        ],
    ));
    assert_eq!(again["status"], "already_free");
    assert_eq!(
        ok(&run(&fx, &["--name", "agent-8", "--now", T0]))["previous"],
        "none"
    );
}

/// Boundaries as literals: 59 refused / 60 accepted, 86400 accepted / 86401 refused.
#[test]
fn the_lease_is_bounded() {
    let fx = Fixture::new();
    for (secs, accepted) in [
        ("59", false),
        ("60", true),
        ("86400", true),
        ("86401", false),
    ] {
        let name = format!("lane-{secs}");
        let out = run(&fx, &["--name", &name, "--lease-secs", secs, "--now", T0]);
        if accepted {
            ok(&out);
        } else {
            assert_eq!(refused(&out).0, "invalid_input", "{secs}");
            assert!(!fx.entry(&name).exists(), "{secs}");
        }
    }
}

#[test]
fn unsafe_names_and_bad_instances_are_refused_before_any_write() {
    let fx = Fixture::new();
    for name in ["", ".hidden", "../x", "a/b", "a b", "x.json/.."] {
        let (kind, _) = refused(&run(&fx, &["--name", name, "--now", T0]));
        assert_eq!(kind, "invalid_input", "{name:?}");
    }
    let (kind, _) = refused(&run(
        &fx,
        &["--name", "agent-8", "--instance", "ABC", "--now", T0],
    ));
    assert_eq!(kind, "invalid_input");
    assert!(!fx.sessions().exists(), "nothing was created");
}

#[test]
fn a_symlinked_or_malformed_entry_is_refused_and_kept() {
    let fx = Fixture::new();
    fs::create_dir_all(fx.sessions()).unwrap();
    let target = fx.home.join("elsewhere.json");
    fs::write(&target, "keep").unwrap();
    std::os::unix::fs::symlink(&target, fx.entry("agent-8")).unwrap();
    let (kind, message) = refused(&run(&fx, &["--name", "agent-8", "--now", T0]));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("symlink"), "{message}");
    assert_eq!(fs::read_to_string(&target).unwrap(), "keep");

    fs::write(fx.entry("agent-9"), "{not json").unwrap();
    let (kind, message) = refused(&run(&fx, &["--name", "agent-9", "--now", T0]));
    assert_eq!(kind, "invalid_input", "{message}");
    assert_eq!(
        fs::read_to_string(fx.entry("agent-9")).unwrap(),
        "{not json"
    );

    // An entry recorded for another name is not taken as this name's.
    fs::write(
        fx.entry("agent-7"),
        format!(
            r#"{{"schema_version":1,"name":"agent-6","instance":"{}","lease_until":1,"registered_at":0}}"#,
            "0".repeat(32)
        ),
    )
    .unwrap();
    assert_eq!(
        refused(&run(&fx, &["--name", "agent-7", "--now", T0])).0,
        "invalid_input"
    );
}

#[test]
fn outside_a_repository_it_is_refused() {
    let fx = Fixture::new();
    let plain = fx.home.join("plain");
    fs::create_dir_all(&plain).unwrap();
    let out = run_in(
        &fx,
        &plain,
        &[plain.to_str().unwrap(), "--name", "agent-8", "--now", T0],
    );
    refused(&out);
    assert!(!plain.join(".moai-cli").exists());
}

/// Eight concurrent registrations of one name, with the read-to-write window widened: exactly
/// one is registered, the other seven are conflicts.
#[test]
fn concurrent_registrations_admit_one() {
    let fx = Fixture::new();
    let repo = fx.repo.to_str().unwrap().to_string();
    let children: Vec<_> = (0..8)
        .map(|_| {
            Command::new(BIN)
                .args(["session-register", &repo, "--name", "agent-8", "--now", T0])
                .current_dir(&fx.repo)
                .env("HOME", &fx.home)
                .env("GIT_CONFIG_GLOBAL", "/dev/null")
                .env("GIT_CONFIG_SYSTEM", "/dev/null")
                .env("MOAI_CLI_TEST_SESSION_REGISTRY_HOLD_MS", "150")
                .stdout(std::process::Stdio::piped())
                .stderr(std::process::Stdio::piped())
                .spawn()
                .unwrap()
        })
        .collect();
    let outs: Vec<Output> = children
        .into_iter()
        .map(|c| c.wait_with_output().unwrap())
        .collect();
    let wins = outs.iter().filter(|o| o.status.code() == Some(0)).count();
    let conflicts = outs
        .iter()
        .filter(|o| o.status.code() == Some(1) && envelope(o)["error"]["kind"] == "conflict")
        .count();
    assert_eq!((wins, conflicts), (1, 7));
}

/// A well-formed entry for `name` with `instance`, `lease_until` and `schema_version`, padded
/// with trailing spaces to exactly `len` bytes when `len` is given.
fn entry_text(
    name: &str,
    instance: &str,
    lease_until: i64,
    schema: u64,
    len: Option<usize>,
) -> String {
    let mut s = format!(
        r#"{{"schema_version":{schema},"name":"{name}","instance":"{instance}","lease_until":{lease_until},"registered_at":0}}"#
    );
    if let Some(n) = len {
        assert!(s.len() <= n);
        s.push_str(&" ".repeat(n - s.len()));
    }
    s
}

/// Audit B2 A1 (boundary): an entry of 4096 bytes is read; one of 4097 bytes is refused and
/// kept as it is.
#[test]
fn an_entry_over_four_kib_is_refused_and_kept() {
    let fx = Fixture::new();
    fs::create_dir_all(fx.sessions()).unwrap();
    let held = "1".repeat(32);
    fs::write(
        fx.entry("lane-a"),
        entry_text("lane-a", &held, 1, 1, Some(4096)),
    )
    .unwrap();
    assert_eq!(
        ok(&run(&fx, &["--name", "lane-a", "--now", T0]))["previous"],
        "expired",
        "4096 bytes are read"
    );
    let big = entry_text("lane-b", &held, 1, 1, Some(4097));
    fs::write(fx.entry("lane-b"), &big).unwrap();
    assert_eq!(
        refused(&run(&fx, &["--name", "lane-b", "--now", T0])).0,
        "invalid_input"
    );
    assert_eq!(fs::read_to_string(fx.entry("lane-b")).unwrap(), big);
}

/// Audit B2 A2: a registry folder that is a symlink is refused; nothing is written through it.
#[test]
fn a_symlinked_registry_folder_is_refused_and_its_target_untouched() {
    let fx = Fixture::new();
    let elsewhere = fx.home.join("elsewhere");
    fs::create_dir_all(&elsewhere).unwrap();
    fs::create_dir_all(fx.sessions().parent().unwrap()).unwrap();
    std::os::unix::fs::symlink(&elsewhere, fx.sessions()).unwrap();
    let (kind, message) = refused(&run(&fx, &["--name", "agent-8", "--now", T0]));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("symlink"), "{message}");
    assert_eq!(
        fs::read_dir(&elsewhere).unwrap().count(),
        0,
        "nothing written there"
    );
}

/// Audit B2 A3: a `registry.lock` that is a symlink is refused; its target is not opened for
/// writing and no entry is written.
#[test]
fn a_symlinked_lock_is_refused_and_its_target_untouched() {
    let fx = Fixture::new();
    fs::create_dir_all(fx.sessions()).unwrap();
    let target = fx.home.join("lock-target");
    fs::write(&target, "keep").unwrap();
    std::os::unix::fs::symlink(&target, fx.sessions().join("registry.lock")).unwrap();
    let (kind, message) = refused(&run(&fx, &["--name", "agent-8", "--now", T0]));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("symlink"), "{message}");
    assert_eq!(fs::read_to_string(&target).unwrap(), "keep");
    assert!(!fx.entry("agent-8").exists());
}

/// Audit B2 A5 and A9: an entry whose `instance` is not 32 lowercase hex, or whose
/// `schema_version` is not 1, is refused and kept.
#[test]
fn an_entry_with_a_bad_instance_or_schema_is_refused_and_kept() {
    let fx = Fixture::new();
    fs::create_dir_all(fx.sessions()).unwrap();
    let cases = [
        ("lane-i", entry_text("lane-i", &"Z".repeat(32), 1, 1, None)),
        ("lane-s", entry_text("lane-s", &"1".repeat(32), 1, 2, None)),
    ];
    for (name, text) in &cases {
        fs::write(fx.entry(name), text).unwrap();
        assert_eq!(
            refused(&run(&fx, &["--name", name, "--now", T0])).0,
            "invalid_input",
            "{name}"
        );
        assert_eq!(&fs::read_to_string(fx.entry(name)).unwrap(), text, "{name}");
    }
}

/// Audit B2 A7 (boundary): another instance's release is a conflict one second before the
/// lease ends and `already_free` at the end second.
#[test]
fn a_release_sees_the_lease_end_second_as_free() {
    let fx = Fixture::new();
    ok(&run(&fx, &["--name", "agent-8", "--now", T0]));
    let other = "e".repeat(32);
    let release_at = |now: &str| {
        run(
            &fx,
            &[
                "--name",
                "agent-8",
                "--instance",
                &other,
                "--release",
                "--now",
                now,
            ],
        )
    };
    assert_eq!(refused(&release_at("2026-11-13T10:59:59Z")).0, "conflict");
    assert!(fx.entry("agent-8").exists());
    assert_eq!(
        ok(&release_at("2026-11-13T11:00:00Z"))["status"],
        "already_free"
    );
}

/// #2902 O3a: `--show` reports the live holder (or `free`) and writes nothing, not even the
/// registry folder or its lock.
#[test]
fn show_reports_the_holder_and_writes_nothing() {
    let fx = Fixture::new();
    let free = ok(&run(&fx, &["--show", "--name", "agent-8", "--now", T0]));
    assert_eq!(free, json!({"status": "free", "name": "agent-8"}));
    assert!(!fx.sessions().exists(), "show creates no registry folder");

    let first = ok(&run(&fx, &["--name", "agent-8", "--now", T0]));
    let before = fs::read(fx.entry("agent-8")).unwrap();
    let held = ok(&run(&fx, &["--show", "--name", "agent-8", "--now", T0]));
    assert_eq!(
        held,
        json!({
            "status": "held",
            "name": "agent-8",
            "instance": first["instance"],
            "lease_until": "2026-11-13T11:00:00Z",
        })
    );
    assert_eq!(fs::read(fx.entry("agent-8")).unwrap(), before);
    // At the lease end second the name is free again, as for a registration.
    let ended = ok(&run(
        &fx,
        &[
            "--show",
            "--name",
            "agent-8",
            "--now",
            "2026-11-13T11:00:00Z",
        ],
    ));
    assert_eq!(ended["status"], "free");
    assert!(fx.entry("agent-8").exists(), "show removes nothing");
    // `--show` takes no instance and does not combine with a release.
    let zeros = "0".repeat(32);
    for extra in [
        vec!["--instance", zeros.as_str()],
        vec!["--release", "--instance", zeros.as_str()],
    ] {
        let mut args = vec!["--show", "--name", "agent-8", "--now", T0];
        args.extend_from_slice(&extra);
        assert_eq!(run(&fx, &args).status.code(), Some(2), "{args:?}");
    }
    let (kind, _) = refused(&run(&fx, &["--show", "--name", "../x", "--now", T0]));
    assert_eq!(kind, "invalid_input");
}

fn release_by_leader(fx: &Fixture, instance: &str, now: &str) -> Output {
    run(
        fx,
        &[
            "--release",
            "--by",
            "leader",
            "--name",
            "agent-8",
            "--instance",
            instance,
            "--now",
            now,
        ],
    )
}

/// #2902 O3a: the leader releases the instance it read with `--show`; the release is still
/// compare-and-delete, so a name registered again since then is a conflict and is kept.
#[test]
fn the_leader_releases_the_instance_it_saw() {
    let fx = Fixture::new();
    ok(&run(&fx, &["--name", "agent-8", "--now", T0]));
    let seen = ok(&run(&fx, &["--show", "--name", "agent-8", "--now", T0]));
    let seen = seen["instance"].as_str().unwrap().to_string();
    let released = ok(&release_by_leader(&fx, &seen, T0));
    assert_eq!(
        released,
        json!({"status": "released", "name": "agent-8", "released_by": "leader"})
    );
    assert!(!fx.entry("agent-8").exists());

    // The leader reads the holder, then the lane's lease ends and it registers again.
    let stale = ok(&run(&fx, &["--name", "agent-8", "--now", T0]));
    let stale = stale["instance"].as_str().unwrap().to_string();
    let later = "2026-11-13T11:00:00Z";
    let fresh = ok(&run(&fx, &["--name", "agent-8", "--now", later]));
    assert_eq!(fresh["previous"], "expired");
    let fresh = fresh["instance"].as_str().unwrap().to_string();
    assert_ne!(fresh, stale);
    let before = fs::read(fx.entry("agent-8")).unwrap();
    let (kind, message) = refused(&release_by_leader(&fx, &stale, later));
    assert_eq!(kind, "conflict");
    assert!(!message.contains(&fresh), "{message}");
    assert_eq!(fs::read(fx.entry("agent-8")).unwrap(), before);

    // An ended lease, or no entry, is `already_free` and records no release.
    let ended = ok(&release_by_leader(&fx, &fresh, "2026-11-13T12:00:00Z"));
    assert_eq!(ended, json!({"status": "already_free", "name": "agent-8"}));
    let fx2 = Fixture::new();
    assert_eq!(
        ok(&release_by_leader(&fx2, &"0".repeat(32), T0)),
        json!({"status": "already_free", "name": "agent-8"})
    );
}

/// `--by` accepts only `leader`, and only with `--release`; a plain `--release` keeps its
/// output byte for byte.
#[test]
fn by_accepts_only_leader_and_plain_release_is_unchanged() {
    let fx = Fixture::new();
    let first = ok(&run(&fx, &["--name", "agent-8", "--now", T0]));
    let instance = first["instance"].as_str().unwrap();
    for by in ["other", "Leader", "lane", ""] {
        let out = run(
            &fx,
            &[
                "--release",
                "--by",
                by,
                "--name",
                "agent-8",
                "--instance",
                instance,
                "--now",
                T0,
            ],
        );
        assert_eq!(out.status.code(), Some(2), "--by {by:?}");
        assert_eq!(envelope(&out)["error"]["kind"], "usage");
    }
    let out = run(
        &fx,
        &[
            "--by",
            "leader",
            "--name",
            "agent-8",
            "--instance",
            instance,
            "--now",
            T0,
        ],
    );
    assert_eq!(out.status.code(), Some(2), "--by without --release");
    assert!(fx.entry("agent-8").exists());

    let out = run(
        &fx,
        &[
            "--release",
            "--name",
            "agent-8",
            "--instance",
            instance,
            "--now",
            T0,
        ],
    );
    assert_eq!(out.status.code(), Some(0));
    assert_eq!(
        String::from_utf8(out.stdout).unwrap(),
        "{\"protocol_version\":1,\"operation\":\"session.register\",\"ok\":true,\"result\":{\"status\":\"released\",\"name\":\"agent-8\"}}\n"
    );
}
