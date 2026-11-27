//! The host-wake CLI (SPEC-20261122-session-messages): the registry read is
//! real (a temporary config root), a missing or unreadable sessions folder
//! answers with its path pinned in the refused shape (never an internal
//! error), each unreachable addressee answers its one-line refusal, a
//! `--idle` wake is refused (the notice rides the model-session tool), and a
//! resolved addressee answers queued with `socket_write: true` — one line to
//! its inbox socket and nothing else.

mod common;
use common::home::TempHome;
use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
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
/// The socket path names a harness-made folder (N8 of the first audit: never
/// the real /tmp/cc-socks folder); the tag gives each caller its own folder
/// (follow-up 3: one name per process races under threaded cargo test).
fn config_root(tag: &str) -> TempDir {
    let socks = SockFolder::new(tag);
    let dir = TempDir::new().unwrap();
    let sessions = dir.path().join("sessions");
    fs::create_dir_all(&sessions).unwrap();
    fs::write(
        sessions.join("1.json"),
        format!(
            r#"{{"name":"leader","messagingSocketPath":"{}"}}"#,
            socks.socket("1.sock")
        ),
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

/// A test-owned socket folder under /tmp with the registry folder's name
/// shape, with its mode set explicitly (the audit's item 12: never the real
/// /tmp/cc-socks folder, and never the umask's default); the Drop removes it,
/// panic or not.
struct SockFolder(std::path::PathBuf);
impl SockFolder {
    fn new(tag: &str) -> Self {
        let path = std::path::Path::new("/tmp")
            .join(format!("cc-socks-moai-test-{tag}-{}", std::process::id()));
        let _ = fs::remove_dir_all(&path);
        fs::create_dir_all(&path).unwrap();
        #[cfg(unix)]
        {
            use std::os::unix::fs::PermissionsExt as _;
            fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
        }
        Self(path)
    }
    fn socket(&self, name: &str) -> String {
        self.0.join(name).display().to_string()
    }
}
impl Drop for SockFolder {
    fn drop(&mut self) {
        let _ = fs::remove_dir_all(&self.0);
    }
}

#[test]
fn ac1_an_unknown_addressee_is_the_different_root_refusal() {
    let dir = config_root("ac1");
    let v = wake(&dir, "someone-else");
    assert_eq!(v["shape"], json!("refused"));
    assert!(
        v["reason"]
            .as_str()
            .unwrap()
            .contains("different config root"),
        "{v}"
    );
    assert_eq!(v["socket_write"], json!(false));
}

/// B1 of the N-slice audit: a missing sessions folder answers in the refused
/// shape with its path pinned, never an internal error. Follow-up 2: an
/// invalid wake answers its own reason even with the folder missing.
#[test]
fn ac2_a_missing_sessions_folder_is_refused_with_its_path() {
    let missing = TempDir::new().unwrap();
    let nowhere = missing.path().join("nowhere");
    let v = ok(&run(&[
        "host-wake",
        "leader",
        "--text",
        "a verdict landed",
        "--config-dir",
        nowhere.to_str().unwrap(),
    ]));
    assert_eq!(v["shape"], json!("refused"), "{v}");
    let reason = v["reason"].as_str().unwrap();
    // Follow-up 1 and 4 of the landed follow-ups: the sessions folder's own
    // path appears in the reason, not the bare config dir.
    assert!(reason.contains("missing or unreadable"), "{v}");
    assert!(
        reason.contains(&nowhere.join("sessions").display().to_string()),
        "the sessions folder path is pinned: {v}"
    );
    assert_eq!(v["socket_write"], json!(false), "{v}");
    // Follow-up 2: the idle flag answers its own line, not the folder's.
    let v = ok(&run(&[
        "host-wake",
        "leader",
        "--text",
        "a verdict landed",
        "--idle",
        "--config-dir",
        nowhere.to_str().unwrap(),
    ]));
    assert!(
        v["reason"]
            .as_str()
            .unwrap()
            .contains("idle notice rides the model-session tool"),
        "{v}"
    );
    // Follow-up 2: an invalid name answers its own line too.
    let v = ok(&run(&[
        "host-wake",
        "bad name",
        "--text",
        "a verdict landed",
        "--config-dir",
        nowhere.to_str().unwrap(),
    ]));
    assert!(
        v["reason"]
            .as_str()
            .unwrap()
            .contains("character outside letters"),
        "{v}"
    );
}

#[test]
fn ac2b_a_session_without_a_socket_is_the_no_inbox_refusal() {
    let dir = config_root("ac2b");
    let v = wake(&dir, "worker");
    assert_eq!(v["shape"], json!("refused"));
    assert!(
        v["reason"].as_str().unwrap().contains("binds no inbox"),
        "{v}"
    );
    assert_eq!(v["socket_write"], json!(false), "{v}");
}

#[cfg(unix)]
#[test]
fn ac3_a_resolved_addressee_answers_the_queued_shape() {
    // The addressee's inbox stands in as a bound listener the test owns, in a
    // folder with the registry socket folder's name shape; the wake writes
    // exactly the minimum user line to it and answers queued with
    // socket_write true.
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
    assert_eq!(v["shape"], json!("queued"), "{v}");
    assert_eq!(v["socket_write"], json!(true), "{v}");
    let line = reader.join().unwrap();
    // The audit's N2: the exact line, not a substring.
    assert_eq!(
        line,
        "{\"type\":\"user\",\"message\":{\"role\":\"user\",\"content\":\"a verdict landed; read the card timeline\"}}\n",
        "exactly the minimum user line and one newline"
    );
}

/// N10 of the first audit: a name answering to two live sessions is refused
/// outright; the wake never picks one.
#[test]
fn ac4_an_ambiguous_name_is_refused() {
    let socks = SockFolder::new("two");
    let dir = TempDir::new().unwrap();
    let sessions = dir.path().join("sessions");
    fs::create_dir_all(&sessions).unwrap();
    for (file, socket) in [("1.json", "a.sock"), ("2.json", "b.sock")] {
        fs::write(
            sessions.join(file),
            format!(
                r#"{{"name":"leader","messagingSocketPath":"{}"}}"#,
                socks.socket(socket)
            ),
        )
        .unwrap();
    }
    let v = wake(&dir, "leader");
    assert_eq!(v["shape"], json!("refused"), "{v}");
    assert!(
        v["reason"].as_str().unwrap().contains("more than one"),
        "{v}"
    );
    assert_eq!(v["socket_write"], json!(false), "{v}");
}

/// N7 of the first audit: the idle notice rides the model-session tool; the
/// minimum line carries none, so the flag is refused.
#[test]
fn ac5_the_idle_flag_is_refused_with_its_line() {
    let dir = config_root("ac5");
    let v = wake_with_idle(&dir);
    assert_eq!(v["shape"], json!("refused"), "{v}");
    assert!(
        v["reason"]
            .as_str()
            .unwrap()
            .contains("idle notice rides the model-session tool"),
        "{v}"
    );
    assert_eq!(v["socket_write"], json!(false), "{v}");
}

fn wake_with_idle(dir: &TempDir) -> Value {
    ok(&run(&[
        "host-wake",
        "leader",
        "--text",
        "a verdict landed",
        "--idle",
        "--config-dir",
        dir.path().to_str().unwrap(),
    ]))
}

/// The landed slice's follow-up (mutant 1 of the N-slice audit): a socket
/// folder with the group-write bit set — even alone, mode 0o775 — is
/// refused; the world-write bit alone refuses too.
#[cfg(unix)]
#[test]
fn ac6_a_socket_folder_open_to_group_or_world_writes_is_refused() {
    for mode in [0o777u32, 0o775, 0o720, 0o702] {
        let socks = SockFolder::new("mode");
        use std::os::unix::fs::PermissionsExt as _;
        fs::set_permissions(&socks.0, fs::Permissions::from_mode(mode)).unwrap();
        let socket_path = socks.0.join("wake.sock");
        let _listener = std::os::unix::net::UnixListener::bind(&socket_path).unwrap();
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
        let v = wake(&dir, "leader");
        assert_eq!(v["shape"], json!("refused"), "mode {mode:o}: {v}");
        assert!(
            v["reason"]
                .as_str()
                .unwrap()
                .contains("group or world writes"),
            "mode {mode:o}: {v}"
        );
    }
}

/// The follow-up of the landed follow-ups: a symlinked sessions folder — and
/// a symlinked config root — is followed on purpose (D-8): users link their
/// config root, the listing only reads, and the socket-path guard checks
/// every component between the socket and /tmp before any connect, so a link
/// cannot redirect a send.
#[cfg(unix)]
#[test]
fn ac7_a_linked_sessions_folder_and_a_linked_config_root_are_followed() {
    let socks = SockFolder::new("linkcfg");
    let socket_path = socks.0.join("wake.sock");
    let _listener = std::os::unix::net::UnixListener::bind(&socket_path).unwrap();

    // A real config root with a sessions folder, linked into place.
    let real_root = TempDir::new().unwrap();
    let real_sessions = real_root.path().join("sessions");
    fs::create_dir_all(&real_sessions).unwrap();
    fs::write(
        real_sessions.join("1.json"),
        format!(
            r#"{{"name":"leader","messagingSocketPath":"{}"}}"#,
            socket_path.display()
        ),
    )
    .unwrap();

    // Case 1: the sessions folder itself is a link.
    let linked_sessions = TempDir::new().unwrap();
    std::os::unix::fs::symlink(&real_sessions, linked_sessions.path().join("sessions")).unwrap();
    let v = wake(&linked_sessions, "leader");
    assert_eq!(v["shape"], json!("queued"), "sessions link: {v}");
    assert_eq!(v["socket_write"], json!(true), "sessions link: {v}");

    // Case 2: the whole config root is a link.
    let linked_root = TempDir::new().unwrap();
    std::os::unix::fs::symlink(real_root.path(), linked_root.path().join("root")).unwrap();
    let v = ok(&run(&[
        "host-wake",
        "leader",
        "--text",
        "a verdict landed",
        "--config-dir",
        linked_root.path().join("root").to_str().unwrap(),
    ]));
    assert_eq!(v["shape"], json!("queued"), "root link: {v}");
    assert_eq!(v["socket_write"], json!(true), "root link: {v}");
}

/// SPEC-20261124-home-env-readers AC-20 (REQ-16): without `--config-dir` the registry is looked
/// up under `CLAUDE_CONFIG_DIR` when it is set, else under HOME's `.claude`, and an unset or
/// empty HOME leaves the relative `.claude` of the working folder. The refusal names the folder
/// it looked in.
#[test]
fn host_wake_default_config_dir() {
    let home = TempHome::new();
    let cwd = TempDir::new().unwrap();
    let set = TempDir::new().unwrap();
    let looked_in = |config: Option<&Path>, home_value: Option<&Path>| -> String {
        let mut cmd = home.command(&["host-wake", "leader", "--text", "a verdict landed"]);
        cmd.current_dir(cwd.path());
        match home_value {
            Some(value) => cmd.env("HOME", value),
            None => cmd.env_remove("HOME"),
        };
        if let Some(dir) = config {
            cmd.env("CLAUDE_CONFIG_DIR", dir);
        }
        let v = ok(&cmd.output().unwrap());
        assert_eq!(v["shape"], json!("refused"), "{v}");
        v["reason"].as_str().unwrap().to_string()
    };
    let refusal = |sessions: PathBuf| {
        format!(
            "wake refused: the sessions registry {} is missing or unreadable",
            sessions.display()
        )
    };
    let under_set = refusal(set.path().join("sessions"));
    assert_eq!(looked_in(Some(set.path()), Some(home.path())), under_set);
    assert_eq!(looked_in(Some(set.path()), None), under_set);
    assert_eq!(
        looked_in(Some(Path::new("")), Some(home.path())),
        refusal(PathBuf::from("sessions"))
    );
    assert_eq!(
        looked_in(None, Some(home.path())),
        refusal(home.path().join(".claude").join("sessions"))
    );
    let relative = refusal(Path::new(".claude").join("sessions"));
    assert_eq!(looked_in(None, Some(Path::new(""))), relative);
    assert_eq!(looked_in(None, None), relative);
}
