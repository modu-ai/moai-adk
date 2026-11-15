//! #2871 U3c: `home-install-apply` and `home-install-revert` on temporary HOME folders only.
//! The real HOME is never read or written: every case passes `--home <tempdir>`, secrets are
//! synthetic names with placeholder or no content and a snapshot never reads a file under a
//! vault or with a secret-looking name. Faults are injected with the debug-only
//! `MOAI_CLI_TEST_INSTALL_FAULT` knob. The Rust sessions are those of a temporary git project.
#![cfg(unix)]

mod common;
use serde_json::{json, Value};
use std::fs;
use std::os::unix::fs::{symlink, PermissionsExt};
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

const BIN: &str = crate::common::MOAI_BIN;
const NOW: &str = "2026-11-14T12:00:00Z";
const STAMP: &str = "20261114T120000Z";
const LATER: &str = "2026-11-14T13:00:00Z";
const PERM: &str = "user approved in the fixture";
const FAULT: &str = "MOAI_CLI_TEST_INSTALL_FAULT";
/// Far in the future for any test run (year 2100), so no test depends on the clock.
const LEASE_RUNNING: i64 = 4_102_444_800;
/// Long past: the lease is over.
const LEASE_OVER: i64 = 1;
const MOVED: [&str; 13] = [
    "agents",
    "backups",
    "cleanup",
    "config",
    "home-routing",
    "home-style",
    "installations",
    "path",
    "receipts",
    "schedule",
    "state",
    "vault",
    "versions",
];

fn marker() -> &'static str {
    moai_cli::state_root::RUST_MARKER
}

fn write(root: &Path, rel: &str, body: &str) {
    let p = root.join(rel);
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::write(p, body).unwrap();
}

fn chmod(p: &Path, mode: u32) {
    fs::set_permissions(p, fs::Permissions::from_mode(mode)).unwrap();
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

/// A temporary HOME holding this generation's items in `~/.moai-cli`, and a temporary git
/// project whose session registry the apply reads.
struct Fx {
    _home: tempfile::TempDir,
    home: PathBuf,
    _project: tempfile::TempDir,
    project: PathBuf,
}

impl Fx {
    fn old(&self, rel: &str) -> PathBuf {
        self.home.join(".moai-cli").join(rel)
    }

    fn dest(&self, rel: &str) -> PathBuf {
        self.home.join(".moai").join(rel)
    }

    fn journal(&self) -> PathBuf {
        self.old("home-install")
    }

    fn sessions(&self) -> PathBuf {
        self.project.join(".git/.moai-cli/sessions")
    }
}

fn project_dir() -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::tempdir().unwrap();
    let project = fs::canonicalize(dir.path()).unwrap();
    git(&project, &["init", "-q"]);
    (dir, project)
}

/// Every item this generation writes sits in `~/.moai-cli`, no `~/.moai`, plus the entries that
/// stay (user data, the old worktrees, an unknown name, a link, the install journal) and a link
/// in `path/` into the earlier `versions`.
fn fx() -> Fx {
    let dir = tempfile::tempdir().unwrap();
    let home = fs::canonicalize(dir.path()).unwrap();
    let old = marker_free_rust_items(&home);
    symlink(
        old.join("versions/1.0.0/bin/moai-cli"),
        old.join("path/moai-cli"),
    )
    .unwrap();
    let (_p, project) = project_dir();
    Fx {
        _home: dir,
        home,
        _project: _p,
        project,
    }
}

fn marker_free_rust_items(home: &Path) -> PathBuf {
    for (rel, body) in [
        (".moai-cli/agents/a.json", "{}"),
        (".moai-cli/backups/b.txt", "x"),
        (".moai-cli/cleanup/c.json", "{}"),
        (".moai-cli/config/launcher.yaml", "a: 1\n"),
        (".moai-cli/home-routing/r.json", "{}"),
        (".moai-cli/home-style/s.json", "{}"),
        (".moai-cli/installations/i.json", "{\"install\":\"record\"}"),
        (".moai-cli/path/moai", "#!/bin/sh\n"),
        (
            ".moai-cli/receipts/r.json",
            "{\"receipt\":\"before the move\"}",
        ),
        (".moai-cli/schedule/s.json", "{}"),
        (".moai-cli/state/last-model.txt", "m\n"),
        (".moai-cli/vault/.master.key", "PLACEHOLDER-SECRET"),
        (".moai-cli/vault/anthropic", "PLACEHOLDER-SECRET"),
        (".moai-cli/versions/1.0.0/bin/moai-cli", "bin-1"),
        (".moai-cli/worktrees/projects/p/wt/README", "x\n"),
        (".moai-cli/projects/p/notes.md", "x\n"),
        (".moai-cli/bin/old-tool", "x"),
        (".moai-cli/mystery/m.txt", "x"),
        (".moai-cli/home-install/keep.txt", "x"),
    ] {
        write(home, rel, body);
    }
    write(home, &format!(".moai-cli/{}", marker()), "{}");
    symlink("agents", home.join(".moai-cli/link-x")).unwrap();
    home.join(".moai-cli")
}

fn call(args: &[&str], envs: &[(&str, &str)]) -> Value {
    let out = Command::new(BIN)
        .args(args)
        .env_remove("MOAI_HOME")
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .envs(envs.iter().copied())
        .stdin(Stdio::null())
        .output()
        .unwrap();
    serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)))
}

fn plan_at(f: &Fx, now: &str) -> Value {
    let v = call(
        &[
            "home-install-plan",
            "--home",
            f.home.to_str().unwrap(),
            "--now",
            now,
            "--project",
            f.project.to_str().unwrap(),
        ],
        &[],
    );
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn plan(f: &Fx) -> Value {
    plan_at(f, NOW)
}

fn digest_of(p: &Value) -> String {
    p["plan_digest"].as_str().unwrap().to_string()
}

fn apply_with(f: &Fx, now: &str, digest: &str, perm: &str, envs: &[(&str, &str)]) -> Value {
    call(
        &[
            "home-install-apply",
            "--home",
            f.home.to_str().unwrap(),
            "--project",
            f.project.to_str().unwrap(),
            "--now",
            now,
            "--plan-digest",
            digest,
            "--permission-reference",
            perm,
        ],
        envs,
    )
}

fn apply(f: &Fx, digest: &str) -> Value {
    apply_with(f, NOW, digest, PERM, &[])
}

/// A preview and the apply of it; the apply must succeed.
fn applied(f: &Fx) -> Value {
    let p = plan(f);
    assert_eq!(p["action"], json!("install"), "{p}");
    let a = apply(f, &digest_of(&p));
    assert_eq!(a["ok"], json!(true), "{a}");
    a["result"].clone()
}

fn revert_with(f: &Fx, extra: &[&str]) -> Value {
    let mut args = vec![
        "home-install-revert",
        "--home",
        f.home.to_str().unwrap(),
        "--permission-reference",
        PERM,
    ];
    args.extend_from_slice(extra);
    call(&args, &[])
}

fn revert(f: &Fx) -> Value {
    revert_with(f, &[])
}

fn kind(v: &Value) -> &str {
    assert_eq!(v["ok"], json!(false), "{v}");
    v["error"]["kind"].as_str().unwrap()
}

fn message(v: &Value) -> String {
    v["error"]["message"]
        .as_str()
        .unwrap_or_default()
        .to_string()
}

fn outcome<'a>(result: &'a Value, name: &str) -> &'a str {
    result["items"]
        .as_array()
        .unwrap()
        .iter()
        .find(|i| i["name"] == name)
        .unwrap_or_else(|| panic!("{name} missing: {result}"))["outcome"]
        .as_str()
        .unwrap()
}

fn secret_looking(path: &Path) -> bool {
    let name = path
        .file_name()
        .unwrap()
        .to_string_lossy()
        .to_ascii_lowercase();
    path.components().any(|c| c.as_os_str() == "vault")
        || name.starts_with(".env")
        || name.starts_with("credentials")
        || name.contains("master")
}

/// Every path under `root`, the folders in `skip` left out: kind, mode and, for a regular file
/// that is not secret-looking, its bytes; for any other file its size. A link is not followed.
fn tree(root: &Path, skip: &[PathBuf]) -> Vec<String> {
    fn walk(p: &Path, skip: &[PathBuf], out: &mut Vec<String>) {
        if skip.iter().any(|s| s == p) {
            return;
        }
        let m = fs::symlink_metadata(p).unwrap();
        let mode = m.permissions().mode() & 0o7777;
        if m.file_type().is_symlink() {
            out.push(format!(
                "{} link {:o} -> {}",
                p.display(),
                mode,
                fs::read_link(p).unwrap().display()
            ));
        } else if m.is_dir() {
            out.push(format!("{} dir {mode:o}", p.display()));
            let mut names: Vec<_> = fs::read_dir(p)
                .unwrap()
                .map(|e| e.unwrap().path())
                .collect();
            names.sort();
            for n in names {
                walk(&n, skip, out);
            }
        } else if m.is_file() && !secret_looking(p) {
            out.push(format!(
                "{} file {mode:o} {:?}",
                p.display(),
                fs::read(p).unwrap()
            ));
        } else {
            out.push(format!("{} other {mode:o} {}", p.display(), m.len()));
        }
    }
    let mut out = Vec::new();
    walk(root, skip, &mut out);
    out
}

/// The tree of the HOME without the install journal folder, which the apply and the revert keep.
fn home_tree(f: &Fx) -> Vec<String> {
    tree(&f.home, &[f.journal()])
}

fn generation(f: &Fx) -> &'static str {
    moai_cli::state_root::classify_home(&f.home).generation
}

fn journal_state(f: &Fx) -> Value {
    let p = f.journal().join(format!("move-journal-{STAMP}.json"));
    let v: Value = serde_json::from_slice(&fs::read(p).unwrap()).unwrap();
    v["state"].clone()
}

/// The lock a running launcher holds on `agent-<n>.live`, or a registry entry as
/// `session-register` writes it.
fn register(f: &Fx, name: &str, lease: i64) {
    fs::create_dir_all(f.sessions()).unwrap();
    fs::write(
        f.sessions().join(format!("{name}.json")),
        format!(
            "{{\"schema_version\":1,\"name\":\"{name}\",\"instance\":\"{}\",\"lease_until\":{lease},\"registered_at\":1}}\n",
            "0123456789abcdef0123456789abcdef"
        ),
    )
    .unwrap();
}

// ---- the apply ----

#[test]
fn apply_moves_the_rust_items_and_writes_the_marker_last() {
    let f = fx();
    let before = home_tree(&f);
    let r = applied(&f);
    assert_eq!(r["state"], json!("applied"), "{r}");
    assert_eq!(r["resumed"], json!(false), "{r}");
    assert_eq!(r["generation"], json!("rust-first"), "{r}");
    for n in MOVED {
        assert!(
            fs::symlink_metadata(f.old(n)).is_err(),
            "{n} is still in ~/.moai-cli"
        );
        assert!(
            fs::symlink_metadata(f.dest(n)).is_ok(),
            "{n} is not in ~/.moai"
        );
        assert_eq!(outcome(&r, n), "moved");
    }
    // What stays, stays.
    for n in [
        "worktrees",
        "projects",
        "bin",
        "mystery",
        "link-x",
        "home-install",
    ] {
        assert!(fs::symlink_metadata(f.old(n)).is_ok(), "{n} left");
        assert!(fs::symlink_metadata(f.dest(n)).is_err(), "{n} moved");
    }
    // The earlier marker stays too; the one in ~/.moai is the new, schema-2 file.
    assert_eq!(fs::read_to_string(f.old(marker())).unwrap(), "{}");
    // Bytes arrive as they were.
    assert_eq!(
        fs::read_to_string(f.dest("state/last-model.txt")).unwrap(),
        "m\n"
    );
    // The marker is the schema-2 one, written by this install, with the mode of ~/.moai-cli.
    let m: Value = serde_json::from_slice(&fs::read(f.dest(marker())).unwrap()).unwrap();
    assert_eq!(m["schema_version"], json!(2), "{m}");
    assert_eq!(m["install_stamp"], json!(STAMP), "{m}");
    let old_mode = fs::metadata(f.old("")).unwrap().permissions().mode() & 0o777;
    assert_eq!(
        fs::metadata(f.dest("")).unwrap().permissions().mode() & 0o777,
        old_mode
    );
    assert_eq!(
        fs::metadata(f.dest(marker())).unwrap().permissions().mode() & 0o777,
        old_mode & 0o666
    );
    assert_eq!(generation(&f), "rust-first");
    assert_eq!(moai_cli::home_root::dir(&f.home), ".moai");
    // The manifest and the journal are there, and the journal is committed.
    assert!(f
        .journal()
        .join(format!("install-manifest-{STAMP}.json"))
        .is_file());
    assert_eq!(journal_state(&f), json!("committed"));
    assert_ne!(home_tree(&f), before);
}

#[test]
fn home_state_root_stays_moai_cli_until_the_marker() {
    let f = fx();
    let p = plan(&f);
    let a = apply_with(&f, NOW, &digest_of(&p), PERM, &[(FAULT, "before_marker")]);
    assert_eq!(kind(&a), "internal", "{a}");
    assert!(
        message(&a).contains("injected crash at before_marker"),
        "{a}"
    );
    // Everything has moved, and still every reader looks in ~/.moai-cli: ~/.moai holds the
    // items but no marker, so it is not a generation yet.
    for n in MOVED {
        assert!(fs::symlink_metadata(f.dest(n)).is_ok(), "{n}");
    }
    assert!(!f.dest(marker()).exists());
    assert_eq!(generation(&f), "unknown");
    assert_eq!(moai_cli::home_root::dir(&f.home), ".moai-cli");
    // The same command finishes it, and only then does the root flip.
    let b = apply(&f, &digest_of(&p));
    assert_eq!(b["result"]["state"], json!("applied"), "{b}");
    assert_eq!(b["result"]["resumed"], json!(true), "{b}");
    assert_eq!(moai_cli::home_root::dir(&f.home), ".moai");
}

/// Each stop leaves a state the same command finishes (or, for the two stops inside a rename,
/// rolls back, since whether the rename happened is not guessed).
#[test]
fn apply_stops_after_each_step_then_resume() {
    for point in [
        "after_manifest",
        "after_mkdir",
        "after_journal",
        "after_mark_moving",
        "after_rename",
        "after_move:1",
        "after_move:7",
        "before_commit",
        "after_commit",
        "after_stub:1",
        "before_marker",
        "after_marker",
    ] {
        let f = fx();
        let before = home_tree(&f);
        let p = plan(&f);
        let a = apply_with(&f, NOW, &digest_of(&p), PERM, &[(FAULT, point)]);
        assert!(
            message(&a).contains(&format!("injected crash at {point}")),
            "{point}: {a}"
        );
        // Nothing but the marker makes the HOME read as moved.
        if point != "after_marker" {
            assert_ne!(generation(&f), "rust-first", "{point}");
        }
        let b = apply(&f, &digest_of(&p));
        assert_eq!(b["ok"], json!(true), "{point}: {b}");
        let r = &b["result"];
        match point {
            "after_mark_moving" | "after_rename" => {
                assert_eq!(r["state"], json!("rolled_back"), "{point}: {r}");
                assert_eq!(home_tree(&f), before, "{point}");
            }
            "after_marker" => assert_eq!(r["state"], json!("unchanged"), "{point}: {r}"),
            _ => {
                assert_eq!(r["state"], json!("applied"), "{point}: {r}");
                assert_eq!(r["resumed"], json!(true), "{point}: {r}");
            }
        }
        if r["state"] != json!("rolled_back") {
            assert_eq!(generation(&f), "rust-first", "{point}");
            for n in MOVED {
                assert!(fs::symlink_metadata(f.dest(n)).is_ok(), "{point}: {n}");
            }
            // The link was re-pointed whichever step stopped.
            assert_eq!(
                fs::read_link(f.dest("path/moai-cli")).unwrap(),
                f.dest("versions/1.0.0/bin/moai-cli"),
                "{point}"
            );
            assert_eq!(journal_state(&f), json!("committed"), "{point}");
        }
    }
}

#[test]
fn revert_directly_after_a_crash_restores_the_tree() {
    for point in [
        "after_manifest",
        "after_mkdir",
        "after_journal",
        "after_mark_moving",
        "after_rename",
        "after_move:1",
        "after_move:7",
        "before_commit",
        "after_commit",
        "after_stub:1",
        "before_marker",
        "after_marker",
    ] {
        let f = fx();
        let before = home_tree(&f);
        let p = plan(&f);
        let a = apply_with(&f, NOW, &digest_of(&p), PERM, &[(FAULT, point)]);
        assert!(message(&a).contains("injected crash"), "{point}: {a}");
        let r = revert(&f);
        assert_eq!(r["ok"], json!(true), "{point}: {r}");
        assert_eq!(r["result"]["state"], json!("reverted"), "{point}: {r}");
        assert_eq!(home_tree(&f), before, "{point}");
        assert!(
            !f.home.join(".moai").exists(),
            "{point}: ~/.moai was created by the apply"
        );
        assert_eq!(generation(&f), "none", "{point}");
        // The stopped install is over: the same command does not pick it up again.
        let again = apply(&f, &digest_of(&p));
        assert!(
            message(&again).contains("rolled back or reverted"),
            "{point}: {again}"
        );
        assert_eq!(home_tree(&f), before, "{point}");
    }
}

#[test]
fn apply_refuses_a_stale_plan_and_writes_nothing() {
    let f = fx();
    let p = plan(&f);
    let before = home_tree(&f);
    let all_before = tree(&f.home, &[]);
    // An entry appears after the preview: the digest is stale.
    write(&f.home, ".moai-cli/zz-new/x", "x");
    let after_entry = tree(&f.home, &[]);
    let a = apply(&f, &digest_of(&p));
    assert_eq!(kind(&a), "stale_plan", "{a}");
    assert_eq!(tree(&f.home, &[]), after_entry, "nothing was written");
    fs::remove_dir_all(f.old("zz-new")).unwrap();
    assert_eq!(tree(&f.home, &[]), all_before);
    // A digest that was never one, another clock and a missing reference.
    let a = apply(&f, "sha256:00");
    assert_eq!(kind(&a), "stale_plan", "{a}");
    let a = apply_with(&f, LATER, &digest_of(&p), PERM, &[]);
    assert_eq!(kind(&a), "stale_plan", "{a}");
    let a = apply_with(&f, NOW, &digest_of(&p), "", &[]);
    assert_eq!(kind(&a), "permission_reference_missing", "{a}");
    assert_eq!(home_tree(&f), before);
    assert!(
        !f.journal().join("apply.lock").exists(),
        "no lock for a refusal"
    );
    assert!(!f.home.join(".moai").exists());
    // The preview still stands.
    let ok = apply(&f, &digest_of(&p));
    assert_eq!(ok["ok"], json!(true), "{ok}");
}

#[test]
fn apply_refuses_a_plan_that_is_not_an_install() {
    let f = fx();
    // A name taken in ~/.moai blocks the preview, so the apply refuses it.
    fs::create_dir_all(f.dest("agents")).unwrap();
    let p = plan(&f);
    assert_eq!(p["action"], json!("blocked"), "{p}");
    let before = tree(&f.home, &[]);
    let a = apply(&f, &digest_of(&p));
    assert!(message(&a).contains("not `install`"), "{a}");
    assert_eq!(tree(&f.home, &[]), before);
}

#[test]
fn a_second_apply_is_unchanged() {
    let f = fx();
    let p = plan(&f);
    let first = apply(&f, &digest_of(&p));
    assert_eq!(first["result"]["state"], json!("applied"), "{first}");
    let after = tree(&f.home, &[]);
    let second = apply(&f, &digest_of(&p));
    assert_eq!(second["ok"], json!(true), "{second}");
    assert_eq!(second["result"]["state"], json!("unchanged"), "{second}");
    assert_eq!(second["result"]["writes"], json!(false), "{second}");
    assert_eq!(tree(&f.home, &[]), after, "nothing was written");
    // The shortcut still compares the digest with the one the install was made with.
    let wrong = apply(&f, "sha256:00");
    assert_eq!(kind(&wrong), "stale_plan", "{wrong}");
    assert_eq!(tree(&f.home, &[]), after);
}

#[test]
fn a_second_apply_is_refused_while_one_holds_the_lock() {
    let f = fx();
    let p = plan(&f);
    let lock = fs::OpenOptions::new()
        .write(true)
        .create(true)
        .truncate(false)
        .open(f.journal().join("apply.lock"))
        .unwrap();
    lock.lock().unwrap();
    let before = tree(&f.home, &[]);
    let a = apply(&f, &digest_of(&p));
    assert_eq!(kind(&a), "conflict", "{a}");
    assert!(message(&a).contains("install lock"), "{a}");
    assert_eq!(tree(&f.home, &[]), before, "nothing was written");
    drop(lock);
    let ok = apply(&f, &digest_of(&p));
    assert_eq!(ok["ok"], json!(true), "{ok}");
    // The revert takes the same lock.
    let lock = fs::OpenOptions::new()
        .write(true)
        .open(f.journal().join("apply.lock"))
        .unwrap();
    lock.lock().unwrap();
    let before = tree(&f.home, &[]);
    let r = revert(&f);
    assert_eq!(kind(&r), "conflict", "{r}");
    assert!(message(&r).contains("install lock"), "{r}");
    assert_eq!(tree(&f.home, &[]), before, "nothing was reverted");
    drop(lock);
    assert_eq!(revert(&f)["result"]["state"], json!("reverted"));
}

#[test]
fn vault_survives_the_move() {
    let dir = tempfile::tempdir().unwrap();
    let home = fs::canonicalize(dir.path()).unwrap();
    fs::create_dir(home.join("work")).unwrap();
    let (_p, project) = project_dir();
    let f = Fx {
        _home: dir,
        home: home.clone(),
        _project: _p,
        project,
    };
    let key = format!("moai-test-key-{}", std::process::id());
    let run = |args: &[&str], stdin: Option<&str>| {
        use std::io::Write;
        let mut child = Command::new(BIN)
            .args(args)
            .arg("--shared")
            .current_dir(home.join("work"))
            .env("HOME", &home)
            .env("GIT_CEILING_DIRECTORIES", &home)
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .unwrap();
        {
            let mut pipe = child.stdin.take().unwrap();
            if let Some(t) = stdin {
                pipe.write_all(t.as_bytes()).unwrap();
            }
        }
        let out = child.wait_with_output().unwrap();
        let text = String::from_utf8_lossy(&out.stdout).to_string();
        assert!(!text.contains(&key) && !String::from_utf8_lossy(&out.stderr).contains(&key));
        serde_json::from_str::<Value>(&text).unwrap()
    };
    let set = run(&["key-set", "glm"], Some(&key));
    assert_eq!(set["result"]["written"], json!(true), "{set}");
    let status = run(&["key-status", "glm"], None);
    assert_eq!(status["result"]["entry"], json!("present"), "{status}");
    let r = applied(&f);
    assert_eq!(r["vault"]["unchanged"], json!(true), "{r}");
    // The vault moved as a whole, and the key is found where the readers now look.
    assert!(f.old("vault").symlink_metadata().is_err());
    let status = run(&["key-status", "glm"], None);
    assert_eq!(status["result"]["entry"], json!("present"), "{status}");
    // And back.
    let rv = revert(&f);
    assert_eq!(rv["ok"], json!(true), "{rv}");
    let status = run(&["key-status", "glm"], None);
    assert_eq!(status["result"]["entry"], json!("present"), "{status}");
}

#[test]
fn apply_does_not_open_a_master_key_that_is_a_fifo() {
    use std::io::Read as _;
    let f = fx();
    let key = f.old("vault/.master.key");
    fs::remove_file(&key).unwrap();
    assert!(Command::new("mkfifo").arg(&key).status().unwrap().success());
    let p = plan(&f);
    let mut child = Command::new(BIN)
        .args([
            "home-install-apply",
            "--home",
            f.home.to_str().unwrap(),
            "--project",
            f.project.to_str().unwrap(),
            "--now",
            NOW,
            "--plan-digest",
            &digest_of(&p),
            "--permission-reference",
            PERM,
        ])
        .env_remove("MOAI_HOME")
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let start = std::time::Instant::now();
    loop {
        if child.try_wait().unwrap().is_some() {
            break;
        }
        if start.elapsed() > std::time::Duration::from_secs(60) {
            let _ = child.kill();
            let _ = child.wait();
            panic!("the apply waited on the FIFO: it opened the master key");
        }
        std::thread::sleep(std::time::Duration::from_millis(50));
    }
    let mut text = String::new();
    child
        .stdout
        .take()
        .unwrap()
        .read_to_string(&mut text)
        .unwrap();
    let v: Value = serde_json::from_str(&text).unwrap();
    assert_eq!(v["result"]["state"], json!("applied"), "{v}");
    assert_eq!(
        v["result"]["vault"]["after"]["kind"],
        json!("folder"),
        "{v}"
    );
    let moved = fs::symlink_metadata(f.dest("vault/.master.key")).unwrap();
    assert!(std::os::unix::fs::FileTypeExt::is_fifo(&moved.file_type()));
}

#[test]
fn apply_reports_a_loose_vault_without_chmod() {
    let f = fx();
    chmod(&f.old("vault"), 0o755);
    chmod(&f.old("vault/.master.key"), 0o644);
    let r = applied(&f);
    assert_eq!(r["vault"]["loose"], json!(true), "{r}");
    assert_eq!(r["vault"]["before"]["mode"], json!("755"), "{r}");
    assert_eq!(r["vault"]["unchanged"], json!(true), "{r}");
    let mode = |p: PathBuf| fs::metadata(p).unwrap().permissions().mode() & 0o777;
    assert_eq!(mode(f.dest("vault")), 0o755);
    assert_eq!(mode(f.dest("vault/.master.key")), 0o644);
}

#[test]
fn apply_repoints_the_path_stub() {
    let f = fx();
    let r = applied(&f);
    assert_eq!(r["stubs"]["repointed"], json!(["moai-cli"]), "{r}");
    assert_eq!(
        fs::read_link(f.dest("path/moai-cli")).unwrap(),
        f.dest("versions/1.0.0/bin/moai-cli")
    );
    assert_eq!(
        fs::read_to_string(f.dest("path/moai-cli")).unwrap(),
        "bin-1"
    );
    // A plain file in `path/` moves as it is, and no temporary link is left.
    assert_eq!(
        fs::read_to_string(f.dest("path/moai")).unwrap(),
        "#!/bin/sh\n"
    );
    let names: Vec<String> = fs::read_dir(f.dest("path"))
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .collect();
    assert_eq!(names.len(), 2, "{names:?}");
}

#[test]
fn revert_restores_the_old_stub_target() {
    let f = fx();
    let before = home_tree(&f);
    let old_target = fs::read_link(f.old("path/moai-cli")).unwrap();
    applied(&f);
    let rv = revert(&f);
    assert_eq!(rv["ok"], json!(true), "{rv}");
    assert_eq!(rv["result"]["state"], json!("reverted"), "{rv}");
    assert_eq!(
        rv["result"]["stubs"]["restored"],
        json!(["moai-cli"]),
        "{rv}"
    );
    assert_eq!(
        rv["result"]["marker"],
        json!("moved_into_install_journal"),
        "{rv}"
    );
    assert_eq!(rv["result"]["moai"], json!("removed"), "{rv}");
    assert_eq!(fs::read_link(f.old("path/moai-cli")).unwrap(), old_target);
    assert_eq!(home_tree(&f), before);
    // The marker was kept, not deleted.
    let kept = f.journal().join(format!("reverted-marker-{STAMP}.json"));
    assert!(kept.is_file());
    assert_eq!(generation(&f), "none");
    // A second revert changes nothing.
    let again = revert(&f);
    assert_eq!(again["result"]["state"], json!("reverted"), "{again}");
    assert_eq!(again["result"]["changed"], json!(false), "{again}");
    assert_eq!(home_tree(&f), before);
}

#[test]
fn revert_leaves_a_link_somebody_else_pointed_elsewhere() {
    let f = fx();
    applied(&f);
    let elsewhere = f.home.join("elsewhere-bin");
    fs::remove_file(f.dest("path/moai-cli")).unwrap();
    symlink(&elsewhere, f.dest("path/moai-cli")).unwrap();
    let rv = revert(&f);
    assert_eq!(rv["ok"], json!(true), "{rv}");
    assert_eq!(rv["result"]["stubs"]["left"], json!(["moai-cli"]), "{rv}");
    // `path` went back with the link as the person left it.
    assert_eq!(fs::read_link(f.old("path/moai-cli")).unwrap(), elsewhere);
}

#[test]
fn stub_is_not_repointed_when_versions_stays() {
    let f = fx();
    // `versions` becomes a link: it stays, so the old target stays valid.
    let parked = f.home.join("parked-versions");
    fs::rename(f.old("versions"), &parked).unwrap();
    symlink(&parked, f.old("versions")).unwrap();
    let link = f.old("path/moai-cli");
    fs::remove_file(&link).unwrap();
    symlink(f.old("versions/1.0.0/bin/moai-cli"), &link).unwrap();
    let want = fs::read_link(&link).unwrap();
    let r = applied(&f);
    assert_eq!(r["stubs"]["repointed"], json!([]), "{r}");
    assert_eq!(fs::read_link(f.dest("path/moai-cli")).unwrap(), want);
    assert!(fs::symlink_metadata(f.old("versions"))
        .unwrap()
        .file_type()
        .is_symlink());
    assert_eq!(
        fs::read_to_string(f.dest("path/moai-cli")).unwrap(),
        "bin-1"
    );
}

#[test]
fn apply_refuses_a_linked_moai_folder() {
    let f = fx();
    let elsewhere = f.home.join("elsewhere");
    fs::create_dir(&elsewhere).unwrap();
    symlink(&elsewhere, f.home.join(".moai")).unwrap();
    let p = plan(&f);
    let before = tree(&f.home, &[]);
    let a = apply(&f, &digest_of(&p));
    assert_eq!(a["ok"], json!(false), "{a}");
    assert_eq!(tree(&f.home, &[]), before);
    assert!(
        fs::read_dir(&elsewhere).unwrap().next().is_none(),
        "nothing went through the link"
    );
    // The same after a stop: a resume does not preview again, so it judges the folder itself.
    let g = fx();
    let p = plan(&g);
    let a = apply_with(&g, NOW, &digest_of(&p), PERM, &[(FAULT, "after_manifest")]);
    assert!(message(&a).contains("injected crash"), "{a}");
    let elsewhere = g.home.join("elsewhere");
    fs::create_dir(&elsewhere).unwrap();
    symlink(&elsewhere, g.home.join(".moai")).unwrap();
    let before = tree(&g.home, &[]);
    let b = apply(&g, &digest_of(&p));
    assert!(message(&b).contains("is a link or not a folder"), "{b}");
    assert_eq!(tree(&g.home, &[]), before, "nothing moved");
    assert!(fs::read_dir(&elsewhere).unwrap().next().is_none());
    // The revert refuses it too, before it moves anything.
    let r = revert(&g);
    assert!(message(&r).contains("is a link or not a folder"), "{r}");
    assert_eq!(tree(&g.home, &[]), before);
}

#[test]
fn revert_refuses_a_changed_marker() {
    let f = fx();
    applied(&f);
    let path = f.dest(marker());
    let mut text = fs::read_to_string(&path).unwrap();
    text.push('\n');
    fs::write(&path, text).unwrap();
    let before = tree(&f.home, &[]);
    let rv = revert(&f);
    assert_eq!(kind(&rv), "conflict", "{rv}");
    assert!(message(&rv).contains("changed since the apply"), "{rv}");
    assert_eq!(tree(&f.home, &[]), before, "nothing was changed");
}

#[test]
fn revert_keeps_a_moai_folder_it_did_not_create() {
    let f = fx();
    fs::create_dir(f.home.join(".moai")).unwrap();
    let before = home_tree(&f);
    let r = applied(&f);
    assert_eq!(r["state"], json!("applied"), "{r}");
    let rv = revert(&f);
    assert_eq!(
        rv["result"]["moai"],
        json!("kept_not_created_by_install"),
        "{rv}"
    );
    assert!(f.home.join(".moai").is_dir());
    assert!(fs::read_dir(f.home.join(".moai")).unwrap().next().is_none());
    assert_eq!(home_tree(&f), before);
}

#[test]
fn revert_keeps_a_moai_folder_that_is_no_longer_empty() {
    let f = fx();
    applied(&f);
    write(&f.home, ".moai/later/x", "a Go item or somebody's file");
    let rv = revert(&f);
    assert_eq!(rv["result"]["moai"], json!("kept_not_empty"), "{rv}");
    assert_eq!(
        fs::read_to_string(f.dest("later/x")).unwrap(),
        "a Go item or somebody's file"
    );
}

#[test]
fn revert_needs_an_install_and_a_reference() {
    let f = fx();
    let r = revert(&f);
    assert!(message(&r).contains("install manifest not found"), "{r}");
    assert_eq!(kind(&r), "not_found");
    let r = call(
        &[
            "home-install-revert",
            "--home",
            f.home.to_str().unwrap(),
            "--permission-reference",
            "",
        ],
        &[],
    );
    assert_eq!(kind(&r), "permission_reference_missing", "{r}");
    let r = revert_with(&f, &["--stamp", "yesterday"]);
    assert!(message(&r).contains("is not YYYYMMDDTHHMMSSZ"), "{r}");
}

#[test]
fn both_commands_require_home() {
    for (cmd, rest) in [
        (
            "home-install-apply",
            vec![
                "--now",
                NOW,
                "--plan-digest",
                "x",
                "--permission-reference",
                PERM,
            ],
        ),
        ("home-install-revert", vec!["--permission-reference", PERM]),
    ] {
        let mut args = vec![cmd];
        args.extend(rest);
        let out = Command::new(BIN)
            .args(&args)
            .stdin(Stdio::null())
            .output()
            .unwrap();
        assert!(!out.status.success(), "{cmd}");
        let text = String::from_utf8_lossy(&out.stdout).to_string()
            + &String::from_utf8_lossy(&out.stderr);
        assert!(text.contains("--home"), "{cmd}: {text}");
    }
}

// ---- boundaries ----

/// A HOME whose `~/.moai-cli` holds a vault and `extra` fillers (names the table does not know,
/// so they stay).
fn bare(extra: usize) -> Fx {
    let dir = tempfile::tempdir().unwrap();
    let home = fs::canonicalize(dir.path()).unwrap();
    write(&home, ".moai-cli/vault/note", "x");
    for i in 0..extra {
        fs::create_dir_all(home.join(format!(".moai-cli/n{i:04}"))).unwrap();
    }
    let (_p, project) = project_dir();
    Fx {
        _home: dir,
        home,
        _project: _p,
        project,
    }
}

#[test]
fn apply_entry_limit_is_500() {
    // 500 entries including the vault. The install journal folder, which the apply creates for
    // its lock, is not counted: the entry that is the 501st on disk changes nothing.
    let f = bare(499);
    let p = plan(&f);
    assert_eq!(p["total"], json!(500), "{p}");
    let a = apply(&f, &digest_of(&p));
    assert_eq!(a["result"]["state"], json!("applied"), "{a}");
    assert!(f.dest("vault/note").is_file());
    let g = bare(500);
    let p = plan(&g);
    assert_eq!(p["action"], json!("blocked"), "{p}");
    let before = tree(&g.home, &[]);
    let a = apply(&g, &digest_of(&p));
    assert!(message(&a).contains("not `install`"), "{a}");
    assert_eq!(tree(&g.home, &[]), before);
}

#[test]
fn apply_stub_limit_is_50() {
    for (links, ok) in [(50usize, true), (51, false)] {
        let f = fx();
        for i in 0..links - 2 {
            symlink(
                f.old("versions/1.0.0/bin/moai-cli"),
                f.old(&format!("path/p{i:03}")),
            )
            .unwrap();
        }
        // `path/` holds `moai` and `moai-cli` already: `links` entries in all.
        let p = plan(&f);
        let a = apply(&f, &digest_of(&p));
        if ok {
            assert_eq!(a["result"]["state"], json!("applied"), "{links}: {a}");
            assert_eq!(
                a["result"]["stubs"]["repointed"].as_array().unwrap().len(),
                links - 1
            );
            assert_eq!(
                fs::read_link(f.dest("path/p000")).unwrap(),
                f.dest("versions/1.0.0/bin/moai-cli")
            );
        } else {
            assert!(message(&a).contains("not `install`"), "{links}: {a}");
            assert!(f.old("path/p000").symlink_metadata().is_ok());
            assert!(!f.home.join(".moai").exists());
        }
    }
}

#[test]
fn apply_versions_limit_is_500() {
    for (count, ok) in [(500usize, true), (501, false)] {
        let f = bare(0);
        for i in 0..count {
            write(
                &f.home,
                &format!(".moai-cli/versions/v{i:04}/bin/moai-cli"),
                "bin",
            );
        }
        let p = plan(&f);
        let a = apply(&f, &digest_of(&p));
        if ok {
            assert_eq!(a["result"]["state"], json!("applied"), "{count}: {a}");
            // The install ids of all of them are in the journal for the readers.
            let aliases = moai_cli::home_install::committed_aliases(&f.home);
            assert_eq!(aliases.len(), count);
        } else {
            assert!(message(&a).contains("not `install`"), "{count}: {a}");
            assert!(f.old("versions").is_dir());
            assert!(!f.home.join(".moai").exists());
        }
    }
}

#[test]
fn apply_journal_limit_is_64() {
    for (existing, ok) in [(63usize, true), (64, false)] {
        let f = fx();
        for i in 0..existing {
            let name = format!("move-journal-20260101T{:02}{:02}00Z.json", i / 60, i % 60);
            fs::write(f.journal().join(name), "{}").unwrap();
        }
        let p = plan(&f);
        let a = apply(&f, &digest_of(&p));
        if ok {
            assert_eq!(a["result"]["state"], json!("applied"), "{existing}: {a}");
        } else {
            assert!(
                message(&a).contains("64 install journals"),
                "{existing}: {a}"
            );
            assert_eq!(kind(&a), "invalid_input");
            assert!(f.old("versions").is_dir());
            assert!(!f.home.join(".moai").exists());
        }
    }
}

// ---- what comes after, what was there before ----

#[test]
fn late_arrivals_are_reported_not_merged() {
    let f = fx();
    let p = plan(&f);
    applied(&f);
    // An older build writes into the earlier folder again.
    write(&f.home, ".moai-cli/state/late.txt", "late");
    write(&f.home, ".moai-cli/agents/late.json", "{}");
    let after = tree(&f.home, &[]);
    let later = plan(&f);
    assert_eq!(later["action"], json!("none"), "{later}");
    let names: Vec<&str> = later["late_arrivals"]
        .as_array()
        .unwrap()
        .iter()
        .map(|e| e["name"].as_str().unwrap())
        .collect();
    assert!(
        names.contains(&"state") && names.contains(&"agents"),
        "{later}"
    );
    // Nothing was merged or moved by looking, nor by applying again.
    assert_eq!(tree(&f.home, &[]), after);
    let again = apply(&f, &digest_of(&p));
    assert_eq!(again["result"]["state"], json!("unchanged"), "{again}");
    assert!(f.old("state/late.txt").is_file());
    assert!(!f.dest("state/late.txt").exists());
    assert_eq!(tree(&f.home, &[]), after);
}

#[test]
fn apply_keeps_a_receipt_recorded_before_the_move() {
    use sha2::{Digest, Sha256};
    let id_of = |p: &Path| format!("{:x}", Sha256::digest(p.to_str().unwrap().as_bytes()));
    let f = fx();
    let receipt = fs::read(f.old("receipts/r.json")).unwrap();
    let record = fs::read(f.old("installations/i.json")).unwrap();
    let old_id = id_of(&fs::canonicalize(f.old("versions/1.0.0/bin/moai-cli")).unwrap());
    applied(&f);
    // The bytes are the ones written before the move.
    assert_eq!(fs::read(f.dest("receipts/r.json")).unwrap(), receipt);
    assert_eq!(fs::read(f.dest("installations/i.json")).unwrap(), record);
    // The build moved, so its install id changed; the journal tells the readers the earlier one.
    let new_id = id_of(&fs::canonicalize(f.dest("versions/1.0.0/bin/moai-cli")).unwrap());
    assert_ne!(old_id, new_id);
    assert_eq!(
        moai_cli::home_install::earlier_ids_of(&f.home, &new_id),
        vec![old_id.clone()]
    );
    // A revert takes the ids back with it.
    revert(&f);
    assert!(moai_cli::home_install::earlier_ids_of(&f.home, &new_id).is_empty());
    assert_eq!(fs::read(f.old("receipts/r.json")).unwrap(), receipt);
}

// ---- rollback ----

#[test]
fn apply_rolls_back_when_the_process_table_cannot_be_read() {
    let f = fx();
    let before = home_tree(&f);
    let p = plan(&f);
    let a = apply_with(
        &f,
        NOW,
        &digest_of(&p),
        PERM,
        &[("MOAI_CLI_TEST_PROCESS_CHECK", "unavailable")],
    );
    assert_eq!(a["ok"], json!(true), "{a}");
    let r = &a["result"];
    assert_eq!(r["state"], json!("rolled_back"), "{r}");
    assert_eq!(r["partial"], json!(true), "{r}");
    assert_eq!(r["moai"], json!("removed"), "{r}");
    assert_eq!(outcome(r, "agents"), "skipped:in_use_unknown", "{r}");
    // The two that run in place were moved, then put back.
    assert_eq!(outcome(r, "versions"), "moved", "{r}");
    assert_eq!(home_tree(&f), before);
    assert_eq!(generation(&f), "none");
    assert_eq!(journal_state(&f), json!("reverted"));
}

#[test]
fn apply_rolls_back_when_a_program_uses_an_entry() {
    let f = fx();
    let before = home_tree(&f);
    let p = plan(&f);
    let mut holder = Command::new("sleep")
        .arg("120")
        .current_dir(f.old("state"))
        .stdin(Stdio::null())
        .spawn()
        .unwrap();
    let a = apply(&f, &digest_of(&p));
    let _ = holder.kill();
    let _ = holder.wait();
    let r = &a["result"];
    assert_eq!(r["state"], json!("rolled_back"), "{a}");
    assert_eq!(outcome(r, "state"), "skipped:process_cwd", "{a}");
    assert_eq!(home_tree(&f), before);
    assert!(!f.home.join(".moai").exists());
    // The same command does not pick the ended install up again; a new preview does.
    let again = apply(&f, &digest_of(&p));
    assert!(
        message(&again).contains("rolled back or reverted"),
        "{again}"
    );
    let p2 = plan_at(&f, LATER);
    assert_eq!(p2["action"], json!("install"), "{p2}");
    let ok = apply_with(&f, LATER, &digest_of(&p2), PERM, &[]);
    assert_eq!(ok["result"]["state"], json!("applied"), "{ok}");
}

// ---- Rust sessions and the journal folder ----

#[test]
fn a_live_rust_session_blocks_the_apply() {
    let f = fx();
    let p = plan(&f);
    register(&f, "leader", LEASE_RUNNING);
    let before = tree(&f.home, &[]);
    let a = apply(&f, &digest_of(&p));
    assert_eq!(kind(&a), "blocked", "{a}");
    assert!(message(&a).contains("rust_session_live:1"), "{a}");
    assert_eq!(tree(&f.home, &[]), before, "nothing was written");
    // Two sessions count as two, an agent among them.
    register(&f, "agent-3", LEASE_RUNNING);
    let a = apply(&f, &digest_of(&p));
    assert!(message(&a).contains("rust_session_live:2"), "{a}");
    // A lease that is over is not a session.
    register(&f, "leader", LEASE_OVER);
    register(&f, "agent-3", LEASE_OVER);
    let ok = apply(&f, &digest_of(&p));
    assert_eq!(ok["result"]["state"], json!("applied"), "{ok}");
    assert_eq!(ok["result"]["sessions"]["live"], json!(0), "{ok}");
}

#[test]
fn an_unreadable_session_registry_blocks_the_apply() {
    let f = fx();
    let p = plan(&f);
    fs::create_dir_all(f.sessions()).unwrap();
    fs::write(f.sessions().join("agent-1.json"), "{not json").unwrap();
    let before = tree(&f.home, &[]);
    let a = apply(&f, &digest_of(&p));
    assert_eq!(kind(&a), "blocked", "{a}");
    assert!(message(&a).contains("rust_session_registry_unknown"), "{a}");
    assert_eq!(tree(&f.home, &[]), before);
    // A project that is no clone has no registry to read: also closed.
    fs::remove_file(f.sessions().join("agent-1.json")).unwrap();
    let plain = tempfile::tempdir().unwrap();
    let out = call(
        &[
            "home-install-apply",
            "--home",
            f.home.to_str().unwrap(),
            "--project",
            plain.path().to_str().unwrap(),
            "--now",
            NOW,
            "--plan-digest",
            &digest_of(&p),
            "--permission-reference",
            PERM,
        ],
        &[],
    );
    assert_eq!(kind(&out), "blocked", "{out}");
    assert_eq!(tree(&f.home, &[]), before);
}

#[test]
fn the_preview_shows_the_sessions_it_can_see() {
    let f = fx();
    let before = tree(&f.home, &[]);
    let p = plan(&f);
    let s = &p["sessions"];
    assert_eq!(s["live"], json!(0), "{p}");
    assert_eq!(
        s["checked"],
        json!(f.sessions().display().to_string()),
        "{p}"
    );
    assert_eq!(s["other_clones"], json!("not visible from the HOME"), "{p}");
    assert_eq!(p["writes"], json!(false));
    register(&f, "agent-9", LEASE_RUNNING);
    let p = plan(&f);
    assert_eq!(p["sessions"]["live"], json!(1), "{p}");
    // The sessions are not an input of the digest.
    assert_eq!(p["plan_digest"], plan(&f)["plan_digest"]);
    fs::remove_file(f.sessions().join("agent-9.json")).unwrap();
    assert_eq!(tree(&f.home, &[]), before);
    // Where no registry can be found, the preview says so instead of claiming none.
    let plain = tempfile::tempdir().unwrap();
    let v = call(
        &[
            "home-install-plan",
            "--home",
            f.home.to_str().unwrap(),
            "--now",
            NOW,
            "--project",
            plain.path().to_str().unwrap(),
        ],
        &[],
    );
    let s = &v["result"]["sessions"];
    assert_eq!(s["checked"], Value::Null, "{v}");
    assert_eq!(s["live"], Value::Null, "{v}");
    assert!(s["error"].is_string(), "{v}");
    assert_eq!(s["other_clones"], json!("not visible from the HOME"), "{v}");
    assert_eq!(v["result"]["plan_digest"], p["plan_digest"]);
}

#[test]
fn apply_refuses_a_journal_folder_that_is_a_link_or_a_file() {
    for as_link in [true, false] {
        let f = fx();
        let p = plan(&f);
        let target = f.home.join("elsewhere");
        fs::create_dir(&target).unwrap();
        fs::remove_dir_all(f.journal()).unwrap();
        if as_link {
            symlink(&target, f.journal()).unwrap();
        } else {
            fs::write(f.journal(), "not a folder").unwrap();
        }
        let before = tree(&f.home, &[]);
        let a = apply(&f, &digest_of(&p));
        assert_eq!(a["ok"], json!(false), "{a}");
        assert!(message(&a).contains("is a link or not a folder"), "{a}");
        assert_eq!(
            tree(&f.home, &[]),
            before,
            "nothing was written (link {as_link})"
        );
        assert!(
            fs::read_dir(&target).unwrap().next().is_none(),
            "nothing went through the link"
        );
        assert!(!f.home.join(".moai").exists());
        // The revert reads the same folder and refuses as well.
        let r = revert(&f);
        assert_eq!(r["ok"], json!(false), "{r}");
        assert_eq!(tree(&f.home, &[]), before);
    }
}
