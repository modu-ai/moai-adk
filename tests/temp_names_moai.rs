//! Slice 4b of SPEC-20261121-moai-name: the temporary names the product makes or checks say
//! `moai`. Here: the marker file of a prepared synthetic HOME (`migrate-*` refuses any other
//! home) and the command help that names it. The launcher's `git init` lock is pinned in
//! `tests/launcher_auto_init.rs`. Every child runs with a temporary HOME and TMPDIR; this file
//! reads no repository file.
mod common;
use moai_cli::migrate::{hash, SYNTHETIC_MARKER, SYNTHETIC_MARKER_TEXT};
use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    path::Path,
    process::{Command, Stdio},
};
use tempfile::TempDir;

const CONFIG: &[u8] = b"{\"language\":\"ko\"}";
/// The marker the product requires, spelled out so a renamed constant cannot hide a wrong name.
const MARKER: &str = ".moai-synthetic-home";
const MARKER_TEXT: &[u8] = b"moai synthetic home v1\n";
/// What an earlier build required. It is not a marker any more.
const EARLIER_MARKER: &str = ".moai-cli-synthetic-home";
const EARLIER_MARKER_TEXT: &[u8] = b"moai-cli synthetic home v1\n";

fn home_with(tmp: &Path, name: &str, marker: &str, text: &[u8]) -> std::path::PathBuf {
    let h = tmp.join(name);
    fs::create_dir_all(h.join(".moai/config")).unwrap();
    fs::write(h.join(".moai/config/language.json"), CONFIG).unwrap();
    fs::write(h.join(marker), text).unwrap();
    h
}

fn request() -> Value {
    json!({
        "hosts": ["Claude"],
        "mode": "Default",
        "go_release": {"version": "v3.2.1", "sha256": [hash(CONFIG)]},
        "selections": [{
            "source": ".moai/config/language.json",
            "target": moai_cli::project_state!("imports/language.json"),
            "recorded_sha256": hash(CONFIG),
            "project_current": null,
        }],
    })
}

/// Runs `moai <args>` with a temporary HOME and the given TMPDIR; returns the exit code and
/// the JSON the command printed.
fn run(args: &[&str], stdin: &Value, user: &Path, tmp: &Path) -> (i32, Value) {
    let mut child = Command::new(crate::common::MOAI_BIN)
        .args(args)
        .env("HOME", user)
        .env("TMPDIR", tmp)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    child
        .stdin
        .take()
        .unwrap()
        .write_all(&serde_json::to_vec(stdin).unwrap())
        .unwrap();
    let out = child.wait_with_output().unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "non-JSON stdout for {args:?}: {} / {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    (out.status.code().unwrap_or(-1), v)
}

fn digest(p: &Path) -> String {
    moai_cli::tree_digest(p, &[]).unwrap()
}

/// REQ-15, REQ-17: the marker is `.moai-synthetic-home` with `moai synthetic home v1`; a home
/// that carries only the earlier marker is refused with a sentence naming the new one, and
/// nothing in that home is read into a plan or written.
#[test]
fn the_synthetic_home_marker_is_moai_and_the_earlier_marker_is_refused() {
    assert_eq!(SYNTHETIC_MARKER, MARKER);
    assert_eq!(SYNTHETIC_MARKER_TEXT, MARKER_TEXT);

    let user = TempDir::new().unwrap();
    let tmp = TempDir::new().unwrap();

    // The marker the product names opens the home (the preview writes nothing).
    let good = home_with(tmp.path(), "good", MARKER, MARKER_TEXT);
    let before = digest(&good);
    let (code, v) = run(
        &["migrate-plan", good.to_str().unwrap()],
        &request(),
        user.path(),
        tmp.path(),
    );
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["result"]["writes"], 0, "{v}");
    assert_eq!(digest(&good), before, "the preview wrote");

    // The earlier marker with its own text is no marker: refused, naming the new marker.
    let old = home_with(tmp.path(), "old", EARLIER_MARKER, EARLIER_MARKER_TEXT);
    let before = digest(&old);
    let (code, v) = run(
        &["migrate-plan", old.to_str().unwrap()],
        &request(),
        user.path(),
        tmp.path(),
    );
    assert_ne!(code, 0, "{v}");
    let message = v["error"]["message"].as_str().unwrap_or_default();
    assert!(
        message.contains(".moai-synthetic-home marker missing"),
        "the refusal does not name the new marker: {v}"
    );
    assert!(
        !message.contains(EARLIER_MARKER),
        "the refusal still names the earlier marker: {v}"
    );
    assert_eq!(digest(&old), before, "the refused home changed");

    // The new file name with the earlier text, and the earlier name with the new text, are
    // refused too: both the name and the bytes are part of the marker.
    for (name, marker, text) in [
        ("name-new-text-old", MARKER, EARLIER_MARKER_TEXT),
        ("name-old-text-new", EARLIER_MARKER, MARKER_TEXT),
    ] {
        let h = home_with(tmp.path(), name, marker, text);
        let (code, v) = run(
            &["migrate-plan", h.to_str().unwrap()],
            &request(),
            user.path(),
            tmp.path(),
        );
        assert_ne!(code, 0, "{name}: {v}");
    }
}

/// REQ-15: the help of every `migrate-*` command that takes a home names the new marker and
/// never the earlier one.
#[test]
fn migrate_help_names_the_moai_marker_only() {
    for command in [
        "migrate-plan",
        "migrate-apply",
        "migrate-recovery-plan",
        "migrate-recover",
    ] {
        let out = Command::new(crate::common::MOAI_BIN)
            .args([command, "--help"])
            .env("HOME", TempDir::new().unwrap().path())
            .output()
            .unwrap();
        let help = String::from_utf8_lossy(&out.stdout);
        assert!(out.status.success(), "{command} --help failed: {help}");
        assert!(
            help.contains(MARKER),
            "{command} --help does not name {MARKER}: {help}"
        );
        assert!(
            !help.contains(EARLIER_MARKER),
            "{command} --help still names {EARLIER_MARKER}: {help}"
        );
    }
}
