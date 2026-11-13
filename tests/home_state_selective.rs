//! #2815: `moai-cli home-state-plan --selective` on temporary HOME folders only. The real HOME
//! is never read or written: every case passes `--home <tempdir>`, secrets are synthetic names
//! with placeholder or no content, and git runs with no global or system configuration.
#![cfg(unix)]

mod common;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Child, Command, Stdio};
use std::time::{Duration, Instant};

const BIN: &str = crate::common::MOAI_BIN;
const NOW: &str = "2026-11-13T23:20:00Z";

fn write(root: &Path, rel: &str, body: &str) {
    let p = root.join(rel);
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::write(p, body).unwrap();
}

/// A Go-generation `~/.moai` with an entry of every class: shared worktrees, Go run data,
/// secrets and a name the table does not know.
fn go_home() -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::tempdir().unwrap();
    let home = fs::canonicalize(dir.path()).unwrap();
    write(
        &home,
        ".moai/user-assets.json",
        r#"{"schema_version": 1, "files": {}}"#,
    );
    write(&home, ".moai/state/active-sessions.json", "[]");
    write(&home, ".moai/credentials.yaml", "token: PLACEHOLDER\n");
    write(&home, ".moai/worktrees/projects/other/wt1/README", "x\n");
    write(&home, ".moai/cache/a.bin", "x");
    write(&home, ".moai/logs/go.log", "x\n");
    write(&home, ".moai/run/x.pid", "1\n");
    write(&home, ".moai/tmp-123/f", "x");
    write(&home, ".moai/gateway-auth/token", "PLACEHOLDER\n");
    write(&home, ".moai/odd-entry/f", "x");
    (dir, home)
}

fn plan_raw(home: &Path, extra: &[&str]) -> Value {
    let mut c = Command::new(BIN);
    c.args([
        "home-state-plan",
        "--home",
        home.to_str().unwrap(),
        "--now",
        NOW,
    ])
    .args(extra)
    .env_remove("MOAI_HOME")
    .stdin(Stdio::null());
    let mut child = c
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    // A plan that opened a FIFO secret would block here; fail instead of hanging.
    let start = Instant::now();
    while child.try_wait().unwrap().is_none() {
        if start.elapsed() > Duration::from_secs(60) {
            let _ = child.kill();
            panic!("home-state-plan did not finish (did it open a secret?)");
        }
        std::thread::sleep(Duration::from_millis(50));
    }
    let out = child.wait_with_output().unwrap();
    let v: Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)));
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn selective(home: &Path) -> Value {
    plan_raw(home, &["--selective"])
}

fn entry<'a>(plan: &'a Value, name: &str) -> &'a Value {
    plan["entries"]
        .as_array()
        .unwrap()
        .iter()
        .find(|e| e["name"] == name)
        .unwrap_or_else(|| panic!("{name} missing: {plan}"))
}

fn class_of(plan: &Value, name: &str) -> (String, String) {
    let e = entry(plan, name);
    (
        e["class"].as_str().unwrap().to_string(),
        e["reason"].as_str().unwrap().to_string(),
    )
}

fn git(dir: &Path, home: &Path, args: &[&str]) {
    let st = Command::new("git")
        .args(["-c", "user.name=t", "-c", "user.email=t@example.invalid"])
        .args(args)
        .current_dir(dir)
        .env("HOME", home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_NOSYSTEM", "1")
        .env_remove("GIT_DIR")
        .env_remove("GIT_WORK_TREE")
        .stdout(Stdio::null())
        .stderr(Stdio::null())
        .status()
        .unwrap();
    assert!(st.success(), "git {args:?}");
}

struct Kill(Child);
impl Drop for Kill {
    fn drop(&mut self) {
        let _ = self.0.kill();
        let _ = self.0.wait();
    }
}

/// AC-1: every entry gets one class and a reason; the plan writes nothing.
#[test]
fn each_entry_gets_a_class_and_a_reason() {
    let (_d, home) = go_home();
    // A real git worktree nested in a Go run-data entry (registered from a repo outside).
    let repo = home.join("repo");
    fs::create_dir_all(&repo).unwrap();
    git(&repo, &home, &["init", "-q"]);
    git(&repo, &home, &["commit", "-q", "--allow-empty", "-m", "x"]);
    fs::create_dir_all(home.join(".moai/rescue")).unwrap();
    let wt = home.join(".moai/rescue/wt");
    git(
        &repo,
        &home,
        &["worktree", "add", "-q", wt.to_str().unwrap()],
    );
    let before = snapshot(&home);

    let p = selective(&home);
    assert_eq!(p["mode"], json!("selective"));
    assert_eq!(p["state"]["generation"], json!("go"), "{p}");
    assert_eq!(p["writes"], json!(false));
    let want = [
        ("worktrees", "keep-in-place", "worktree_root"),
        ("cache", "move-to-backup", "go_run_data"),
        ("logs", "move-to-backup", "go_run_data"),
        ("run", "move-to-backup", "go_run_data"),
        ("state", "move-to-backup", "go_run_data"),
        ("tmp-123", "move-to-backup", "go_run_data"),
        ("user-assets.json", "move-to-backup", "go_run_data"),
        ("rescue", "keep-in-place", "nested_checkout"),
        ("credentials.yaml", "secret-manual", "secret_name"),
        ("gateway-auth", "secret-manual", "secret_name"),
        ("odd-entry", "unknown", "not_in_table"),
    ];
    for (name, class, reason) in want {
        assert_eq!(
            class_of(&p, name),
            (class.to_string(), reason.to_string()),
            "{name}"
        );
    }
    assert_eq!(p["entries"].as_array().unwrap().len(), want.len(), "{p}");
    assert_eq!(p["total"], json!(want.len()));
    let backup = home.join(".moai-backup");
    assert_eq!(p["backup"], json!(backup.display().to_string()));
    assert_eq!(
        entry(&p, "cache")["destination"],
        json!(backup.join("cache").display().to_string())
    );
    assert_eq!(entry(&p, "odd-entry")["destination"], Value::Null);
    assert_eq!(entry(&p, "worktrees")["destination"], Value::Null);
    assert_eq!(p["counts"]["move-to-backup"], json!(6), "{p}");
    assert_eq!(p["counts"]["unknown"], json!(1));
    assert_eq!(p["action"], json!("select"), "{p}");
    assert_eq!(snapshot(&home), before, "the plan wrote something");
}

/// AC-2: `worktrees/` stays even when nothing uses it, and an entry a process uses stays
/// whatever the table says — through its working folder or its command line.
#[test]
fn worktrees_and_entries_in_use_stay_in_place() {
    let (_d, home) = go_home();
    let p = selective(&home);
    assert_eq!(class_of(&p, "worktrees").0, "keep-in-place");
    assert_eq!(class_of(&p, "logs").0, "move-to-backup");
    assert_eq!(class_of(&p, "run").0, "move-to-backup");

    let _cwd = Kill(
        Command::new("sleep")
            .arg("60")
            .current_dir(home.join(".moai/logs"))
            .spawn()
            .unwrap(),
    );
    let pidfile = home.join(".moai/run/x.pid");
    let _line = Kill(
        Command::new("sh")
            // Two commands, so the shell stays (a lone command is exec'd and loses the path).
            .args(["-c", "sleep 60; :", "watcher", pidfile.to_str().unwrap()])
            .current_dir(&home)
            .spawn()
            .unwrap(),
    );
    let _wt = Kill(
        Command::new("sleep")
            .arg("60")
            .current_dir(home.join(".moai/worktrees/projects/other/wt1"))
            .spawn()
            .unwrap(),
    );
    // Let the processes appear in the process table.
    std::thread::sleep(Duration::from_millis(300));
    let p = selective(&home);
    assert_eq!(
        class_of(&p, "logs"),
        ("keep-in-place".into(), "process_cwd".into())
    );
    assert_eq!(
        class_of(&p, "run"),
        ("keep-in-place".into(), "process_command_line".into())
    );
    assert_eq!(class_of(&p, "worktrees").0, "keep-in-place");
    assert_eq!(class_of(&p, "cache").0, "move-to-backup");
    assert_eq!(entry(&p, "logs")["destination"], Value::Null);
}

/// AC-3: secrets are judged by name only. FIFOs named like secrets block any reader, and a
/// secret folder without permissions fails any walk; the plan still finishes and lists them.
#[test]
fn secrets_are_listed_and_never_opened() {
    let (_d, home) = go_home();
    let moai = home.join(".moai");
    for fifo in [".env", ".env.glm", "credentials-extra"] {
        let c = std::ffi::CString::new(moai.join(fifo).to_str().unwrap()).unwrap();
        // SAFETY: mkfifo with a valid NUL-terminated path.
        assert_eq!(unsafe { libc::mkfifo(c.as_ptr(), 0o600) }, 0, "{fifo}");
    }
    fs::create_dir_all(moai.join("secrets/inner")).unwrap();
    fs::set_permissions(moai.join("secrets"), fs::Permissions::from_mode(0o000)).unwrap();
    let p = selective(&home);
    fs::set_permissions(moai.join("secrets"), fs::Permissions::from_mode(0o700)).unwrap();
    for name in [
        ".env",
        ".env.glm",
        "credentials-extra",
        "credentials.yaml",
        "secrets",
        "gateway-auth",
    ] {
        assert_eq!(
            class_of(&p, name),
            ("secret-manual".into(), "secret_name".into()),
            "{name}"
        );
        assert_eq!(entry(&p, name)["destination"], Value::Null);
    }
    // Upper case is the same name on the default macOS file system.
    let (_d2, home2) = go_home();
    // The lower-case folder would be the same name on a case-insensitive file system.
    fs::remove_dir_all(home2.join(".moai/gateway-auth")).unwrap();
    write(&home2, ".moai/Gateway-Auth", "PLACEHOLDER\n");
    write(&home2, ".moai/CACHE2", "x");
    let p2 = selective(&home2);
    assert_eq!(class_of(&p2, "Gateway-Auth").0, "secret-manual");
    assert_eq!(class_of(&p2, "CACHE2").0, "unknown");
}

use std::os::unix::fs::PermissionsExt;

/// AC-4: without `--selective` the plan keeps its whole-folder output (same keys, no
/// per-entry fields), and the selective mode moves nothing outside the Go generation.
#[test]
fn the_whole_folder_mode_is_unchanged() {
    let (_d, home) = go_home();
    let p = plan_raw(&home, &[]);
    let mut keys: Vec<&str> = p.as_object().unwrap().keys().map(String::as_str).collect();
    keys.sort();
    assert_eq!(
        keys,
        [
            "action",
            "apply",
            "backup",
            "blocks",
            "copy",
            "home",
            "plan_digest",
            "schema_version",
            "state",
            "warnings",
            "writes"
        ]
    );
    assert_eq!(p["action"], json!("migrate"), "{p}");

    // An empty or Rust HOME: nothing to select.
    let dir = tempfile::tempdir().unwrap();
    let empty = fs::canonicalize(dir.path()).unwrap();
    let p = selective(&empty);
    assert_eq!(p["action"], json!("none"));
    assert_eq!(p["entries"], json!([]));
}

/// A live Go session blocks the selective plan as it blocks the whole-folder one.
#[test]
fn a_live_go_session_blocks_the_selective_plan() {
    let (_d, home) = go_home();
    let me = std::process::id();
    write(
        &home,
        ".moai/state/active-sessions.json",
        &format!(r#"[{{"pid": {me}}}]"#),
    );
    let p = selective(&home);
    assert_eq!(p["action"], json!("blocked"), "{p}");
    assert!(p["blocks"]
        .as_array()
        .unwrap()
        .contains(&json!("go_session_live")));
}

fn snapshot(root: &Path) -> Vec<(String, u64)> {
    let mut out = Vec::new();
    let mut stack = vec![root.to_path_buf()];
    while let Some(d) = stack.pop() {
        for e in fs::read_dir(&d).unwrap().flatten() {
            let m = fs::symlink_metadata(e.path()).unwrap();
            let rel = e.path().strip_prefix(root).unwrap().display().to_string();
            if m.is_dir() {
                stack.push(e.path());
            }
            out.push((rel, if m.is_file() { m.len() } else { 0 }));
        }
    }
    out.sort();
    out
}
