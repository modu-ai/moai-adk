//! SPEC-20261121-model-step-reachable AC-3: the model step of `moai update` on a person's
//! terminal, run through a pseudo terminal on a copy of the built binary in a temporary HOME.
//!
//! The installer applies the user layer with a scripted `--confirm`, so the next `moai update`
//! on a person's terminal has nothing to apply. The model questions must still open there when no
//! model setting is saved, and stay closed once one is saved. macOS only, like the other
//! terminal tests (BSD `script` gives the pseudo terminal); the real HOME is never touched.

#![cfg(target_os = "macos")]

mod common;
use common::CleanEnv;

use serde_json::{json, Value};
use std::{
    fs,
    io::{Read as _, Write as _},
    path::{Path, PathBuf},
    process::{Command, Stdio},
    sync::mpsc::{self, RecvTimeoutError},
    thread::{self, sleep},
    time::{Duration, Instant},
};
use tempfile::TempDir;

const EXE: &str = common::MOAI_BIN;
const TOOL_QUESTION: &str = "Which coding tools do you want to set up?";
const SAVE_QUESTION: &str = "Save these model settings? [y/N]";
/// The longest one terminal run may take; a debug binary hashes itself, so this is generous.
const RUN_LIMIT: Duration = Duration::from_secs(90);

struct Fx {
    _tmp: TempDir,
    home: PathBuf,
    exe: PathBuf,
}

/// A temporary HOME holding `~/.moai-cli/versions/a/bin/moai-cli` with its package beside it
/// (the release layout), after the installer's scripted `--confirm` has applied everything.
fn installed_like_the_installer() -> Fx {
    let tmp = TempDir::new().unwrap();
    let home = fs::canonicalize(tmp.path()).unwrap().join("home");
    let folder = home.join(".moai-cli/versions/a");
    fs::create_dir_all(folder.join("bin")).unwrap();
    fs::create_dir_all(folder.join("agents")).unwrap();
    let exe = folder.join("bin/moai-cli");
    common::copy_executable(Path::new(EXE), &exe);
    fs::hard_link(&exe, folder.join("bin/moai")).unwrap();
    let skill = folder.join("skills/moai");
    fs::create_dir_all(skill.join("workflows")).unwrap();
    fs::write(
        skill.join("SKILL.md"),
        "---\nname: moai\ndescription: fixture package\n---\n# MoAI\nbody\n",
    )
    .unwrap();
    fs::write(skill.join("workflows/update.md"), "# UPDATE\nv1\n").unwrap();
    let fx = Fx {
        _tmp: tmp,
        home,
        exe,
    };
    let preview = envelope(&piped(&fx, &["update"]));
    let digest = preview["plan_digest"].as_str().unwrap().to_string();
    let applied = envelope(&piped(&fx, &["update", "--confirm", &digest]));
    assert_eq!(applied["mode"], json!("applied"), "{applied}");
    sleep(Duration::from_millis(1100)); // receipts carry second-resolution timestamps
    let again = envelope(&piped(&fx, &["update"]));
    assert_eq!(again["nothing_to_apply"], json!(true), "{again}");
    fx
}

fn base_command(fx: &Fx, program: &str) -> Command {
    let mut cmd = Command::new(program);
    cmd.clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", &fx.home)
        .env("LC_ALL", "C")
        .env("TERM", "xterm-256color")
        .env("NO_COLOR", "1")
        .current_dir(&fx.home);
    cmd
}

fn piped(fx: &Fx, args: &[&str]) -> std::process::Output {
    let mut child = base_command(fx, fx.exe.to_str().unwrap())
        .args(args)
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let _ = child.stdin.take();
    child.wait_with_output().unwrap()
}

fn envelope(out: &std::process::Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "envelope JSON on stdout: {}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

/// `moai-cli update` on a pseudo terminal (stdin, stdout and stderr all on it). The answers are
/// typed only after `wait_for` is on screen, as a person would; `None` types nothing and waits for
/// the run to end by itself. A run that never shows `wait_for` fails with its output. A run that
/// is still waiting after `RUN_LIMIT` (a question nobody answers) is killed and fails with what it
/// showed, so a regression is a failing test and never a stuck one.
fn on_a_terminal(fx: &Fx, wait_for: Option<&str>, answers: &str) -> String {
    let mut child = base_command(fx, "/usr/bin/script")
        .args(["-q", "/dev/null"])
        .arg(&fx.exe)
        .arg("update")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::null())
        .spawn()
        .unwrap();
    let mut stdout = child.stdout.take().unwrap();
    let (tx, rx) = mpsc::channel::<Vec<u8>>();
    thread::spawn(move || {
        let mut buf = [0u8; 4096];
        while let Ok(n @ 1..) = stdout.read(&mut buf) {
            if tx.send(buf[..n].to_vec()).is_err() {
                break;
            }
        }
    });
    // The write end stays open until the run has ended: closing it early is an end of input
    // (Ctrl-D) at the question, which cancels the step.
    let mut stdin = child.stdin.take().unwrap();
    let deadline = Instant::now() + RUN_LIMIT;
    let mut seen = Vec::new();
    let mut typed = false;
    loop {
        let left = deadline.saturating_duration_since(Instant::now());
        match rx.recv_timeout(left) {
            Ok(chunk) => seen.extend_from_slice(&chunk),
            Err(RecvTimeoutError::Disconnected) => break,
            Err(RecvTimeoutError::Timeout) => {
                let _ = child.kill();
                let _ = child.wait();
                panic!(
                    "still waiting after {RUN_LIMIT:?}: {}",
                    String::from_utf8_lossy(&seen)
                );
            }
        }
        if let Some(question) = wait_for {
            if !typed && String::from_utf8_lossy(&seen).contains(question) {
                stdin.write_all(answers.as_bytes()).unwrap();
                typed = true;
            }
        }
    }
    drop(stdin);
    child.wait().unwrap();
    if let Some(question) = wait_for {
        assert!(
            typed,
            "no {question:?} on screen: {}",
            String::from_utf8_lossy(&seen)
        );
    }
    String::from_utf8_lossy(&seen).replace('\r', "")
}

fn saved_settings(fx: &Fx) -> Option<String> {
    fs::read_to_string(moai_cli::update_models::settings_path(&fx.home)).ok()
}

/// REQ-1: after the installer's `--confirm` there is nothing to apply and nothing is saved, and
/// the model questions still open on a terminal. A no at the save question writes nothing.
#[test]
fn nothing_to_apply_and_nothing_saved_opens_the_model_step() {
    let fx = installed_like_the_installer();
    assert_eq!(saved_settings(&fx), None, "the installer saves no models");
    // Claude only (the update installed Claude and Codex), Enter at its four model questions, a no.
    let text = on_a_terminal(&fx, Some(TOOL_QUESTION), "1\n\n\n\n\nn\n");
    assert!(text.contains(SAVE_QUESTION), "{text}");
    assert!(
        !text.contains("\"protocol_version\""),
        "no envelope on a terminal: {text}"
    );
    assert_eq!(saved_settings(&fx), None, "a no at the save writes nothing");
}

/// REQ-3: once a setting is saved, a run with nothing to apply asks nothing.
#[test]
fn a_saved_model_setting_keeps_the_model_step_closed() {
    let fx = installed_like_the_installer();
    let text = on_a_terminal(&fx, Some(TOOL_QUESTION), "1\n\n\n\n\ny\n");
    assert!(text.contains(SAVE_QUESTION), "{text}");
    let saved: Value = serde_json::from_str(&saved_settings(&fx).expect("saved")).unwrap();
    assert!(
        saved["launcher"]["models"]["claude"]["leader"].is_object(),
        "{saved}"
    );
    let again = on_a_terminal(&fx, None, "");
    assert!(!again.contains(TOOL_QUESTION), "{again}");
    assert!(!again.contains("[y/N]"), "{again}");
}

/// REQ-4: a pipe and a scripted run never open the step, and the preview keeps exactly its keys.
#[test]
fn a_pipe_and_a_scripted_run_never_open_the_step() {
    let fx = installed_like_the_installer();
    let preview = envelope(&piped(&fx, &["update"]));
    assert_eq!(preview["mode"], json!("preview"), "{preview}");
    assert!(preview.get("models").is_none(), "{preview}");
    assert_eq!(saved_settings(&fx), None);
}
