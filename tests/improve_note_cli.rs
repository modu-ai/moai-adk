//! `moai improve-note` (SPEC-20261121-self-improvement-loop, slice S2): the one write of the
//! slice, one line of the D7 shape appended to the project's `improve/notes.jsonl`.
//!
//! Every command runs against a temporary HOME and a project inside it; the notes path is made
//! with the product's own path function (the state folder is never spelled here). Acceptance
//! AC-15 and AC-16 of the SPEC are these tests, run with the real binary.
#![cfg(unix)]

mod common;

use common::home::TempHome;
use serde_json::{json, Value};
use std::ffi::CString;
use std::os::unix::ffi::OsStrExt;
use std::os::unix::fs::{symlink, FileTypeExt};
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};
use std::time::{Duration, Instant, SystemTime, UNIX_EPOCH};

const FILE_LIMIT: usize = 1024 * 1024;

struct World {
    home: TempHome,
    project: PathBuf,
}

fn world() -> World {
    let home = TempHome::new();
    let project = home.project("demo");
    World { home, project }
}

impl World {
    fn notes_path(&self) -> PathBuf {
        moai_cli::state_root::path(&self.project, "improve/notes.jsonl")
    }

    fn command(&self, kind: &str, key: &str, text: &str, extra: &[&str]) -> Command {
        let project = self.project.to_str().unwrap();
        let mut args = vec![
            "improve-note",
            project,
            "--kind",
            kind,
            "--key",
            key,
            "--text",
            text,
        ];
        args.extend_from_slice(extra);
        self.home.command(&args)
    }

    fn note(&self, kind: &str, key: &str, text: &str) -> Output {
        within(self.command(kind, key, text, &[]), Duration::from_secs(30))
    }

    fn lines(&self) -> Vec<Value> {
        std::fs::read_to_string(self.notes_path())
            .unwrap()
            .lines()
            .map(|l| serde_json::from_str(l).unwrap())
            .collect()
    }

    fn signals(&self) -> Value {
        let out = self
            .home
            .command(&["improve-signals", self.project.to_str().unwrap()])
            .output()
            .unwrap();
        result_of(&out, "improve.signals")
    }
}

fn within(mut cmd: Command, limit: Duration) -> Output {
    let mut child = cmd
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .expect("run moai");
    let start = Instant::now();
    loop {
        if child.try_wait().unwrap().is_some() {
            return child.wait_with_output().unwrap();
        }
        if start.elapsed() > limit {
            let _ = child.kill();
            let _ = child.wait();
            panic!("moai did not finish within {limit:?}: an open may be blocked");
        }
        std::thread::sleep(Duration::from_millis(20));
    }
}

fn text(out: &Output) -> String {
    format!(
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

fn result_of(out: &Output, operation: &str) -> Value {
    assert_eq!(out.status.code(), Some(0), "{}", text(out));
    let v: Value =
        serde_json::from_slice(&out.stdout).unwrap_or_else(|e| panic!("{e}: {}", text(out)));
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(v["operation"], json!(operation), "{v}");
    v["result"].clone()
}

fn appended(out: &Output) -> Value {
    result_of(out, "improve.note")
}

/// A refusal: the error kind and the message.
fn refusal(out: &Output) -> (String, String) {
    assert_eq!(out.status.code(), Some(1), "{}", text(out));
    let v: Value =
        serde_json::from_slice(&out.stdout).unwrap_or_else(|e| panic!("{e}: {}", text(out)));
    assert_eq!(v["ok"], json!(false), "{v}");
    let message = v["error"]["message"].as_str().unwrap().to_string();
    assert!(message.starts_with("improve-note refused: "), "{message}");
    (v["error"]["kind"].as_str().unwrap().to_string(), message)
}

fn fifo(path: &Path) {
    let c = CString::new(path.as_os_str().as_bytes()).unwrap();
    // SAFETY: mkfifo with a valid NUL-terminated path.
    assert_eq!(unsafe { libc::mkfifo(c.as_ptr(), 0o600) }, 0, "{path:?}");
}

#[test]
fn a_note_is_one_d7_line_made_when_the_folder_and_file_are_absent() {
    let w = world();
    assert!(!w.notes_path().exists());
    let before = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap()
        .as_secs() as i64;
    let r = appended(&w.note("failure", "cargo", "check was red"));
    let after = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap()
        .as_secs() as i64;
    assert_eq!(r["action"], json!("note"), "{r}");
    assert_eq!(r["writes"], json!(true), "{r}");
    assert_eq!(r["appended"], json!(1), "{r}");
    let lines = w.lines();
    assert_eq!(lines.len(), 1);
    let l = &lines[0];
    assert_eq!(l["v"], json!(1));
    assert_eq!(l["kind"], json!("failure"));
    assert_eq!(l["key"], json!("cargo"));
    assert_eq!(l["text"], json!("check was red"));
    assert!(l["by"].is_string(), "{l}");
    let at = l["at"].as_str().unwrap();
    assert!(
        at.ends_with('Z') && at.len() == 20,
        "UTC RFC 3339 to the second: {at}"
    );
    let secs = moai_cli::forge::policy::parse_utc(at).expect("improve-signals reads it back");
    assert!(before - 1 <= secs && secs <= after + 1, "{at}");
    // The key set is the D7 set and nothing else.
    let mut keys: Vec<&str> = l.as_object().unwrap().keys().map(String::as_str).collect();
    keys.sort_unstable();
    assert_eq!(keys, ["at", "by", "key", "kind", "text", "v"]);
}

#[test]
fn a_note_is_appended_after_the_earlier_ones_and_improve_signals_reads_it_as_its_class() {
    let w = world();
    appended(&w.note("failure", "cargo", "first"));
    appended(&w.note("correction", "naming", "second"));
    appended(&w.note("failure", "cargo", "third"));
    let lines = w.lines();
    assert_eq!(
        lines
            .iter()
            .map(|l| l["text"].as_str().unwrap())
            .collect::<Vec<_>>(),
        ["first", "second", "third"]
    );
    let s = w.signals();
    let class = |key: &str| {
        s["classes"]
            .as_array()
            .unwrap()
            .iter()
            .find(|c| c["key"] == json!(key))
            .cloned()
    };
    let cargo = class("failure:cargo").expect("class failure:cargo");
    assert_eq!(
        (cargo["count"].clone(), cargo["status"].clone()),
        (json!(2), json!("candidate"))
    );
    let naming = class("correction:naming").expect("class correction:naming");
    assert_eq!(
        (naming["count"].clone(), naming["status"].clone()),
        (json!(1), json!("observed_once"))
    );
    assert_eq!(s["skipped"]["bad_line"], json!(0), "{s}");
    for kind in ["false_claim"] {
        appended(&w.note(kind, "claim-1", "said fixed"));
    }
    assert!(w.signals()["classes"]
        .as_array()
        .unwrap()
        .iter()
        .any(|c| c["key"] == json!("false_claim:claim-1")));
}

#[test]
fn a_file_that_does_not_end_with_a_newline_gets_one_before_the_new_line() {
    let w = world();
    std::fs::create_dir_all(w.notes_path().parent().unwrap()).unwrap();
    std::fs::write(w.notes_path(), "{\"v\":1}").unwrap();
    appended(&w.note("failure", "k", "t"));
    let body = std::fs::read_to_string(w.notes_path()).unwrap();
    let mut it = body.lines();
    assert_eq!(it.next(), Some("{\"v\":1}"));
    assert!(it.next().unwrap().starts_with("{\"v\":1,\"at\""), "{body}");
    assert!(body.ends_with('\n'));
}

#[test]
fn the_kind_the_key_the_text_and_the_observer_are_held_to_their_limits() {
    let w = world();
    // N is accepted.
    appended(&w.note("correction", &"k".repeat(80), &"t".repeat(200)));
    appended(&w.note("false_claim", "k", "t"));
    let by80 = "b".repeat(80);
    appended(&within(
        w.command("failure", "k", "t", &["--by", &by80]),
        Duration::from_secs(30),
    ));
    let n = w.lines().len();
    // N + 1 and anything outside the shape is refused with nothing written.
    let long_key = "k".repeat(81);
    let long_text = "t".repeat(201);
    let long_by = "b".repeat(81);
    let before = w.home.snapshot();
    for out in [
        w.note("praise", "k", "t"),
        w.note("", "k", "t"),
        w.note("failure", &long_key, "t"),
        w.note("failure", "", "t"),
        w.note("failure", "k", &long_text),
        w.note("failure", "k\nx", "t"),
        w.note("failure", "k", "two\nlines"),
        w.note("failure", "k", "tab\there"),
        within(
            w.command("failure", "k", "t", &["--by", &long_by]),
            Duration::from_secs(30),
        ),
        within(
            w.command("failure", "k", "t", &["--by", "a\nb"]),
            Duration::from_secs(30),
        ),
    ] {
        let (kind, message) = refusal(&out);
        assert_eq!(kind, "invalid_input", "{message}");
    }
    assert_eq!(w.home.snapshot(), before, "a refusal writes nothing");
    assert_eq!(w.lines().len(), n);
    // The limits count characters, not bytes.
    appended(&w.note("failure", &"한".repeat(80), &"글".repeat(200)));
}

#[test]
fn an_append_that_would_take_the_file_over_one_mebibyte_is_refused() {
    let w = world();
    // The length of one line is fixed (the time has a fixed width), so measure it once.
    appended(&w.note("failure", "k", "t"));
    let one = std::fs::read(w.notes_path()).unwrap().len();
    let pad = |room: usize| {
        // A file of valid-looking filler that leaves `room` bytes under the limit.
        let mut body = vec![b'x'; FILE_LIMIT - room - 1];
        body.push(b'\n');
        std::fs::write(w.notes_path(), body).unwrap();
    };
    // The file ends exactly at the limit: accepted.
    pad(one);
    appended(&w.note("failure", "k", "t"));
    assert_eq!(
        std::fs::metadata(w.notes_path()).unwrap().len() as usize,
        FILE_LIMIT
    );
    // One byte over: refused and nothing written.
    pad(one - 1);
    let before = std::fs::read(w.notes_path()).unwrap();
    let (kind, message) = refusal(&w.note("failure", "k", "t"));
    assert_eq!(kind, "invalid_input", "{message}");
    assert_eq!(std::fs::read(w.notes_path()).unwrap(), before);
    // A file that already is over the limit stays refused.
    std::fs::write(w.notes_path(), vec![b'x'; FILE_LIMIT + 1]).unwrap();
    refusal(&w.note("failure", "k", "t"));
}

#[test]
fn a_secret_shaped_word_is_written_as_redacted_in_the_key_the_text_and_the_observer() {
    let w = world();
    let out = within(
        w.command(
            "correction",
            "password=hunter2",
            "do not use token=abc123def456 here",
            &["--by", "key=sk-live-0123456789abcdef"],
        ),
        Duration::from_secs(30),
    );
    let all = text(&out);
    assert!(
        !all.contains("hunter2")
            && !all.contains("abc123def456")
            && !all.contains("0123456789abcdef"),
        "{all}"
    );
    appended(&out);
    let body = std::fs::read_to_string(w.notes_path()).unwrap();
    assert!(
        !body.contains("hunter2")
            && !body.contains("abc123def456")
            && !body.contains("0123456789abcdef"),
        "{body}"
    );
    assert!(body.contains("[redacted]"), "{body}");
    let l = &w.lines()[0];
    assert!(l["key"].as_str().unwrap().contains("[redacted]"), "{l}");
    assert!(l["text"].as_str().unwrap().contains("[redacted]"), "{l}");
    // Improve-signals still reads the line as a class and never keys a secret.
    assert!(!w.signals().to_string().contains("hunter2"));
}

#[test]
fn a_link_or_a_special_file_on_the_way_is_refused_and_never_followed() {
    let w = world();
    let notes = w.notes_path();
    let improve = notes.parent().unwrap().to_path_buf();
    let state = improve.parent().unwrap().to_path_buf();
    let outside = w.home.path().join("outside");
    std::fs::create_dir_all(&outside).unwrap();
    let sentinel = outside.join("notes.jsonl");
    std::fs::write(&sentinel, "keep\n").unwrap();

    let check = |what: &str| {
        let before = w.home.snapshot();
        let out = within(w.command("failure", "k", "t", &[]), Duration::from_secs(10));
        let (kind, message) = refusal(&out);
        assert_eq!(kind, "invalid_input", "{what}: {message}");
        assert_eq!(
            w.home.snapshot(),
            before,
            "{what}: nothing under the HOME or the project changed"
        );
        assert_eq!(
            std::fs::read_to_string(&sentinel).unwrap(),
            "keep\n",
            "{what}: the link target is not written"
        );
        message
    };

    // 1. `improve/` is a link to a folder elsewhere.
    symlink(&outside, &improve).unwrap();
    assert!(check("improve/ is a link").contains("symbolic link"));
    std::fs::remove_file(&improve).unwrap();

    // 2. The notes file is a link, to a file and to nothing.
    std::fs::create_dir_all(&improve).unwrap();
    symlink(&sentinel, &notes).unwrap();
    assert!(check("notes file is a link").contains("symbolic link"));
    std::fs::remove_file(&notes).unwrap();
    symlink(outside.join("missing"), &notes).unwrap();
    check("notes file is a dangling link");
    assert!(
        !outside.join("missing").exists(),
        "a dangling link is not created through"
    );
    std::fs::remove_file(&notes).unwrap();

    // 3. The notes file is a named pipe: refused before the open can wait.
    // (`TempHome::snapshot` reads every file and would wait on a pipe, so this case compares
    // the folder listing and the pipe itself instead.)
    fifo(&notes);
    let out = within(w.command("failure", "k", "t", &[]), Duration::from_secs(10));
    let (kind, message) = refusal(&out);
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("not a regular file"), "{message}");
    assert_eq!(
        TempHome::names_in(&improve),
        ["notes.jsonl"],
        "nothing else was made"
    );
    assert!(std::fs::symlink_metadata(&notes)
        .unwrap()
        .file_type()
        .is_fifo());
    assert_eq!(std::fs::read_to_string(&sentinel).unwrap(), "keep\n");
    std::fs::remove_file(&notes).unwrap();

    // 4. The notes path is a folder.
    std::fs::create_dir_all(&notes).unwrap();
    assert!(check("notes path is a folder").contains("not a regular file"));
    std::fs::remove_dir(&notes).unwrap();
    std::fs::remove_dir(&improve).unwrap();

    // 5. The state folder above them is a link (a project with no profile, so the link is the
    //    only thing the command can read about the folder).
    let bare = w.home.path().join("work").join("bare");
    std::fs::create_dir_all(&bare).unwrap();
    let real_state = w.home.path().join("real-state");
    std::fs::create_dir_all(&real_state).unwrap();
    symlink(&real_state, moai_cli::state_root::path(&bare, "")).unwrap();
    let before = w.home.snapshot();
    let out = within(
        w.home.command(&[
            "improve-note",
            bare.to_str().unwrap(),
            "--kind",
            "failure",
            "--key",
            "k",
            "--text",
            "t",
        ]),
        Duration::from_secs(10),
    );
    let (kind, message) = refusal(&out);
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("symbolic link"), "{message}");
    assert_eq!(
        w.home.snapshot(),
        before,
        "the folder the state folder points at is not written"
    );
    assert!(std::fs::read_dir(&real_state).unwrap().next().is_none());
    let _ = &state;
}

#[test]
fn the_command_is_listed_in_the_help_and_has_its_operations_document() {
    let w = world();
    let help = w
        .home
        .command(&["improve-note", "--help"])
        .output()
        .unwrap();
    let help = String::from_utf8_lossy(&help.stdout).to_string();
    assert!(
        help.contains("--kind") && help.contains("--key") && help.contains("--text"),
        "{help}"
    );
    assert!(Path::new(env!("CARGO_MANIFEST_DIR"))
        .join("docs/design-review/appendix/operations/improve_note.md")
        .is_file());
}
