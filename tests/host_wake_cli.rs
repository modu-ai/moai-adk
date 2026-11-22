//! The host-wake CLI (SPEC-20261122-session-messages, slice (c) of the leader's
//! decision): the registry read is real (a temporary config root), the refusal
//! reasons answer per addressee, and a resolved addressee answers the D-8 open
//! item — the send seam — without anything being written.

mod common;
use common::home::TempHome;
use serde_json::{json, Value};
use std::{
    fs,
    process::{Output, Stdio},
};
use tempfile::TempDir;

fn run(args: &[&str]) -> Output {
    let home = TempHome::new();
    let mut child = home
        .command(args)
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let _ = child.stdin.take();
    child.wait_with_output().unwrap()
}

fn ok(out: &Output) -> Value {
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

/// A config root with one session that binds an inbox and one that does not.
fn config_root() -> TempDir {
    let dir = TempDir::new().unwrap();
    let sessions = dir.path().join("sessions");
    fs::create_dir_all(&sessions).unwrap();
    fs::write(
        sessions.join("1.json"),
        r#"{"name":"leader","messagingSocketPath":"/tmp/cc-socks/1.sock"}"#,
    )
    .unwrap();
    fs::write(sessions.join("2.json"), r#"{"name":"worker"}"#).unwrap();
    dir
}

fn wake(dir: &TempDir, name: &str) -> Value {
    ok(&run(&[
        "host-wake",
        name,
        "--text",
        "a verdict landed; read the card timeline",
        "--config-dir",
        dir.path().to_str().unwrap(),
    ]))
}

#[test]
fn ac1_an_unknown_addressee_is_the_different_root_refusal() {
    let dir = config_root();
    let v = wake(&dir, "someone-else");
    assert_eq!(v["shape"], json!("refused"));
    assert!(
        v["reason"]
            .as_str()
            .unwrap()
            .contains("different config root"),
        "{v}"
    );
    assert_eq!(v["writes"], json!(false));
}

#[test]
fn ac2_a_session_without_a_socket_is_the_no_inbox_refusal() {
    let dir = config_root();
    let v = wake(&dir, "worker");
    assert_eq!(v["shape"], json!("refused"));
    assert!(
        v["reason"].as_str().unwrap().contains("binds no inbox"),
        "{v}"
    );
}

/// A test-owned socket folder under /tmp with the registry folder's name
/// shape (the audit's N8: never the real /tmp/cc-socks folder); the Drop
/// removes it, panic or not.
struct SockFolder(std::path::PathBuf);
impl SockFolder {
    fn new(tag: &str) -> Self {
        let path = std::path::Path::new("/tmp")
            .join(format!("cc-socks-moai-test-{tag}-{}", std::process::id()));
        let _ = fs::remove_dir_all(&path);
        fs::create_dir_all(&path).unwrap();
        Self(path)
    }
}
impl Drop for SockFolder {
    fn drop(&mut self) {
        let _ = fs::remove_dir_all(&self.0);
    }
}

#[cfg(unix)]
#[test]
fn ac3_a_resolved_addressee_answers_the_queued_shape_and_writes_nothing() {
    // The addressee's inbox stands in as a bound listener the test owns, in a
    // folder with the registry socket folder's name shape; the wake writes
    // exactly the minimum user line to it and answers the queued shape.
    let socks = SockFolder::new("cli");
    let socket_path = socks.0.join("wake.sock");
    let listener = std::os::unix::net::UnixListener::bind(&socket_path).unwrap();
    let dir = TempDir::new().unwrap();
    let sessions = dir.path().join("sessions");
    fs::create_dir_all(&sessions).unwrap();
    fs::write(
        sessions.join("1.json"),
        format!(
            r#"{{"name":"leader","messagingSocketPath":"{}"}}"#,
            socket_path.display()
        ),
    )
    .unwrap();
    let reader = std::thread::spawn(move || {
        let (mut stream, _) = listener.accept().unwrap();
        use std::io::Read as _;
        let mut line = String::new();
        stream.read_to_string(&mut line).unwrap();
        line
    });
    let v = wake(&dir, "leader");
    assert_eq!(v["shape"], json!("delivered"), "{v}");
    let line = reader.join().unwrap();
    // The audit's N2: the exact line, not a substring.
    assert_eq!(
        line,
        "{\"type\":\"user\",\"message\":{\"role\":\"user\",\"content\":\"a verdict landed; read the card timeline\"}}\n",
        "exactly the minimum user line and one newline"
    );
    assert_eq!(v["writes"], json!(false));
}
