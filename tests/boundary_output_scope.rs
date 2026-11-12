//! #2392: executable boundary reproducers. F5c (incomplete output) stays an ignored RED until
//! its classification is decided; F6 (scope symlink) is repaired and runs by default.
#![cfg(unix)]
mod common;

use moai_cli::verify::{run_check, CheckResult};
use serde_json::json;
use std::{
    fs,
    io::{Read, Write},
    os::unix::{fs::symlink, net::UnixListener},
    process::Command,
    time::Duration,
};

// Explicitly run this RED until the leader decides how incomplete evidence is classified.
#[test]
#[ignore = "#2392 F5c: policy decision pending; explicit RED reproducer"]
fn incomplete_output_must_not_be_passed() {
    // Keep the AF_UNIX pathname below the platform socket limit.
    let tmp = tempfile::tempdir_in("/tmp").unwrap();
    let listener = UnixListener::bind(tmp.path().join("control.sock")).unwrap();
    listener.set_nonblocking(true).unwrap();
    let (release, wait) = std::sync::mpsc::channel();
    let controller = std::thread::spawn(move || loop {
        match listener.accept() {
            Ok((mut stream, _)) => {
                stream.set_nonblocking(false).unwrap();
                wait.recv().unwrap();
                stream
                    .set_read_timeout(Some(Duration::from_secs(10)))
                    .unwrap();
                stream.write_all(b"R").unwrap();
                let mut done = [0];
                stream.read_exact(&mut done).unwrap();
                assert_eq!(done, *b"D");
                break true;
            }
            Err(e) if e.kind() == std::io::ErrorKind::WouldBlock => {
                match wait.recv_timeout(Duration::from_millis(10)) {
                    Err(std::sync::mpsc::RecvTimeoutError::Timeout) => {}
                    _ => break false,
                }
            }
            Err(e) => panic!("control listener: {e}"),
        }
    });
    // The holder leaves the command's process group, then synchronizes BEFORE its
    // parent exits. It cannot close either output pipe until run_check returns.
    // No sleep or elapsed-time comparison determines the test outcome.
    fs::write(
        tmp.path().join("holder.py"),
        r#"import os, socket
r, w = os.pipe()
pid = os.fork()
if pid:
    os.close(w)
    assert os.read(r, 1) == b'Y'
    os.close(r)
    os._exit(0)
os.close(r)
os.setsid()
s = socket.socket(socket.AF_UNIX)
s.settimeout(30)
s.connect('control.sock')
os.write(w, b'Y')
os.close(w)
assert s.recv(1) == b'R'
os.close(1)
os.close(2)
s.sendall(b'D')
s.close()
os._exit(0)
"#,
    )
    .unwrap();
    let check = run_check(tmp.path(), 0, "python3 holder.py", Duration::from_secs(30));
    release.send(()).unwrap();
    assert!(
        controller.join().unwrap(),
        "producer never connected to control socket"
    );
    println!("{}", serde_json::to_string(&check).unwrap());
    assert_eq!(check.exit_code, Some(0));
    assert!(!check.timed_out);
    assert_eq!(check.output_complete, Some(false));
    assert_eq!(check.stdout_sha256, None);
    assert_eq!(check.stderr_sha256, None);
    assert_ne!(
        check.result,
        CheckResult::Passed,
        "incomplete evidence accepted as Passed"
    );
}

// F6 repaired: scope-check blocks a link it reaches (tests/scope_symlink_cli.rs has the matrix).
#[test]
fn scope_symlink_escape_must_be_blocked() {
    let tmp = tempfile::tempdir().unwrap();
    let parent = fs::canonicalize(tmp.path()).unwrap();
    let root = parent.join("project");
    let outside = parent.join("outside");
    fs::create_dir_all(root.join("safe")).unwrap();
    fs::create_dir_all(&outside).unwrap();
    fs::write(outside.join("sentinel"), b"ORIGINAL").unwrap();
    symlink(&outside, root.join("safe/link")).unwrap();
    fs::create_dir_all(root.join("home")).unwrap();
    fs::write(root.join("README.md"), "fixture\n").unwrap();
    common::git::git_ok(&root, &["init", "-q", "-b", "main"]);
    common::git::git_ok(&root, &["add", "README.md"]);
    common::git::git_ok(
        &root,
        &[
            "-c",
            "user.name=fixture",
            "-c",
            "user.email=fixture@example.invalid",
            "commit",
            "-q",
            "-m",
            "fixture",
        ],
    );
    let head = common::git::git(&root, &["rev-parse", "HEAD"]);
    let record = json!({
        "schema_version": 1,
        "task_ref": {"provider":"github","host":"github.com","owner":"fixture","repo":"fixture","repository_id":101,"issue_id":23920,"issue_number":2392},
        "task_key":"aaaaaaaaaaaaaaaa","logical_key":"2392", "owner":{"kind":"session","id":"fixture"},
        "epoch":1,"base_revision":head,
        "worktree":{"common_dir":root.join(".git"),"git_dir":root.join(".git"),"path":root,"head":head},
        "write_scope":["safe/**"],"spec_ref":{"id":"S","revision":"r","digest":null},
        "comment_id":1,"claim_token":"Moai-Claim: 2392#1","lease_until":"2099-01-01T00:00:00Z",
        "state":"running","effects":[],"last_observed_at":"2026-11-11T21:00:00Z"
    });
    let journal = root.join(".git/.moai-cli/claims");
    fs::create_dir_all(&journal).unwrap();
    fs::write(
        journal.join("journal.log"),
        format!("{}\n", json!({"event":"claimed", "record":record})),
    )
    .unwrap();
    let output = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .arg("scope-check")
        .arg(&root)
        .args(["--task", "aaaaaaaaaaaaaaaa"])
        .env_clear()
        .env("PATH", std::env::var_os("PATH").unwrap())
        .env("HOME", root.join("home"))
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("LC_ALL", "C")
        .output()
        .unwrap();
    let envelope: serde_json::Value = serde_json::from_slice(&output.stdout).unwrap();
    assert!(
        output.status.success(),
        "{envelope}: {}",
        String::from_utf8_lossy(&output.stderr)
    );
    assert_eq!(envelope["ok"], true);
    let result = &envelope["result"];
    println!("{result}");
    // Simulate the authorized write only when the actual scope entry point permits it.
    if result["decision"] == "clear" {
        fs::write(root.join("safe/link/sentinel"), b"ESCAPED").unwrap();
    }
    assert_eq!(
        fs::read(outside.join("sentinel")).unwrap(),
        b"ORIGINAL",
        "scope-check permitted a symlink escape"
    );
    assert_eq!(result["decision"], "blocked");
}
