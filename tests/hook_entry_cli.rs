//! SPEC-20261123-hook-entry-skeleton: the `moai hook <name>` entry point. Three names are
//! registered (`session-context`, `bash-boundary`, `write-boundary`) and every handler of this
//! slice answers no decision. The entry reads stdin once, stays silent outside a MoAI project,
//! exits 0 for any name it does not know, and records one receipt line when the input of a
//! registered name is bad. In a MoAI project the two `PreToolUse` names also deny an input they
//! cannot read (SPEC-20261126-hook-unreadable-input, the `unreadable_*` cases). Every case runs
//! the real binary in a temporary project and HOME.
#![cfg(unix)]
mod common;

use common::CleanEnv;
use serde_json::Value;
use std::collections::{BTreeMap, BTreeSet};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

const NAMES: [&str; 3] = ["session-context", "bash-boundary", "write-boundary"];
const PRE_TOOL_USE: [&str; 2] = ["bash-boundary", "write-boundary"];
const RECEIPT: &str = "hook-entry.jsonl";
const RECEIPT_CAP: usize = 262_144;
const STDIN_CAP: usize = 4 * 1024 * 1024;
const MARKER: &str = "SECRETMARKER";
/// The receipt reason word of an input over [`STDIN_CAP`].
const TOO_LARGE: &str = "stdin_too_large";
/// The reason of an over-limit `write-boundary` input, as SPEC-20261126-hook-write-boundary
/// REQ-13 prints it.
const WRITE_OVER_LIMIT: &str = "moai-boundary/protected-path: this call is over the 4 MiB the \
    check reads, so its path and content cannot be checked; write the file in calls under 4 MiB";

/// The state folder of a project and the projects folder a registered project has under HOME,
/// both read from the library's constants.
fn state(project: &Path) -> PathBuf {
    project.join(moai_cli::state_root::MOAI_DIR)
}

fn project_file(project: &Path) -> PathBuf {
    state(project).join("project.yaml")
}

fn registered_dir(home: &Path) -> PathBuf {
    moai_cli::home_workdir::projects_root(home).join("demo")
}

struct Run {
    code: i32,
    stdout: String,
    stderr: String,
}

/// A temporary project folder and HOME. `registered` makes the project folder `demo` under the
/// HOME's projects folder.
struct Fx {
    _tmp: tempfile::TempDir,
    project: PathBuf,
    home: PathBuf,
}

impl Fx {
    fn new(moai_project: bool, registered: bool) -> Fx {
        let tmp = tempfile::TempDir::new().unwrap();
        let root = fs::canonicalize(tmp.path()).unwrap();
        let project = root.join("project");
        let home = root.join("home");
        fs::create_dir_all(&project).unwrap();
        fs::create_dir_all(&home).unwrap();
        if moai_project {
            fs::create_dir_all(state(&project)).unwrap();
            fs::write(
                project_file(&project),
                "{\n  \"schema_version\": 2,\n  \"project_id\": \"demo\"\n}\n",
            )
            .unwrap();
        }
        if registered {
            fs::create_dir_all(registered_dir(&home)).unwrap();
        }
        Fx {
            _tmp: tmp,
            project,
            home,
        }
    }

    fn receipt(&self) -> PathBuf {
        registered_dir(&self.home).join(RECEIPT)
    }

    fn run(&self, args: &[&str], cwd: &Path, stdin: &[u8], env: &[(&str, &str)]) -> Run {
        let mut cmd = Command::new(common::MOAI_BIN);
        cmd.args(args)
            .current_dir(cwd)
            .env("HOME", &self.home)
            .env_remove("CLAUDE_PROJECT_DIR")
            .env_remove("MOAI_FORGE_GH_BIN")
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null");
        for (k, v) in env {
            cmd.env(k, v);
        }
        finish(&mut cmd, stdin)
    }

    fn hook(&self, name: &str, stdin: &[u8]) -> Run {
        self.run(&["hook", name], &self.project, stdin, &[])
    }
}

/// Runs `cmd` with `stdin` on its standard input and both output streams captured.
fn finish(cmd: &mut Command, stdin: &[u8]) -> Run {
    cmd.stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());
    let mut child = cmd.spawn().unwrap();
    common::feed_stdin(&mut child, stdin);
    let out = child.wait_with_output().unwrap();
    Run {
        code: out.status.code().unwrap(),
        stdout: String::from_utf8_lossy(&out.stdout).into_owned(),
        stderr: String::from_utf8_lossy(&out.stderr).into_owned(),
    }
}

/// Every file under `root` with its bytes; a link is listed by its target text.
fn snapshot(root: &Path) -> BTreeMap<String, Vec<u8>> {
    fn walk(base: &Path, dir: &Path, out: &mut BTreeMap<String, Vec<u8>>) {
        let Ok(entries) = fs::read_dir(dir) else {
            return;
        };
        for entry in entries.flatten() {
            let path = entry.path();
            let rel = path
                .strip_prefix(base)
                .unwrap()
                .to_string_lossy()
                .into_owned();
            let meta = fs::symlink_metadata(&path).unwrap();
            if meta.file_type().is_symlink() {
                let target = fs::read_link(&path).unwrap();
                out.insert(
                    format!("{rel} -> link"),
                    target.to_string_lossy().as_bytes().to_vec(),
                );
            } else if meta.is_dir() {
                out.insert(format!("{rel}/"), Vec::new());
                walk(base, &path, out);
            } else {
                out.insert(rel, fs::read(&path).unwrap());
            }
        }
    }
    let mut out = BTreeMap::new();
    walk(root, root, &mut out);
    out
}

fn valid_input(event: &str) -> Vec<u8> {
    format!(r#"{{"hook_event_name":"{event}","session_id":"s1","cwd":"/x","source":"startup"}}"#)
        .into_bytes()
}

/// A valid JSON object of exactly `len` bytes whose text holds `SECRETMARKER`.
fn object_of_len(len: usize) -> Vec<u8> {
    let head = br#"{"k":"SECRETMARKER"#;
    let tail = br#""}"#;
    let mut v = head.to_vec();
    v.extend(std::iter::repeat_n(b'x', len - head.len() - tail.len()));
    v.extend_from_slice(tail);
    assert_eq!(v.len(), len);
    v
}

fn receipt_lines(path: &Path) -> Vec<String> {
    match fs::read_to_string(path) {
        Ok(text) => text.lines().map(str::to_owned).collect(),
        Err(_) => Vec::new(),
    }
}

/// AC-1: a registered name with a good input prints nothing and changes nothing.
#[test]
fn ac1_a_registered_name_with_good_input_prints_nothing_and_writes_nothing() {
    let fx = Fx::new(true, true);
    let before = (snapshot(&fx.project), snapshot(&fx.home));
    // `session-context` is no longer silent (SPEC-20261123-hook-session-start): its answers are
    // pinned in `tests/hook_session_start_cli.rs`.
    for (name, event) in [
        ("bash-boundary", "PreToolUse"),
        ("write-boundary", "PreToolUse"),
    ] {
        let r = fx.hook(name, &valid_input(event));
        assert_eq!(r.code, 0, "{name}: {}", r.stderr);
        assert_eq!(r.stdout, "", "{name}");
        assert_eq!(r.stderr, "", "{name}");
    }
    assert_eq!(before, (snapshot(&fx.project), snapshot(&fx.home)));
}

/// AC-2: a name the entry does not know exits 0 with stdout empty, in a project or not, and the
/// Go-generation `pre-tool` keeps its stderr line.
#[test]
fn ac2_an_unknown_name_exits_zero_with_nothing_on_stdout() {
    for moai in [true, false] {
        let fx = Fx::new(moai, true);
        let before = (snapshot(&fx.project), snapshot(&fx.home));
        for name in [
            "pre-tool",
            "nonsense",
            "SESSION-CONTEXT",
            "session-context-x",
            "session_context",
        ] {
            let r = fx.hook(name, &valid_input("PreToolUse"));
            assert_eq!(r.code, 0, "{name} (moai {moai}): {}", r.stderr);
            assert_eq!(r.stdout, "", "{name} (moai {moai})");
        }
        let go = fx.hook("pre-tool", &valid_input("PreToolUse"));
        assert!(
            go.stderr.contains("ignored a Go-generation hook call"),
            "{}",
            go.stderr
        );
        assert_eq!(before, (snapshot(&fx.project), snapshot(&fx.home)));
    }
}

/// AC-3: outside a MoAI project a registered name is silent, writes nothing and records nothing,
/// even for bad input; `CLAUDE_PROJECT_DIR` wins over the working folder.
#[test]
fn ac3_outside_a_moai_project_a_registered_name_is_silent() {
    type Break = fn(&Path);
    let broken: [(&str, Break); 3] = [
        ("no state folder", |_p| {}),
        ("project.yaml is a folder", |p| {
            fs::create_dir_all(project_file(p)).unwrap()
        }),
        ("project.yaml is a link to nothing", |p| {
            fs::create_dir_all(state(p)).unwrap();
            std::os::unix::fs::symlink(p.join("no-such-target"), project_file(p)).unwrap();
        }),
    ];
    for (what, make) in broken {
        let fx = Fx::new(false, true);
        make(&fx.project);
        let before = (snapshot(&fx.project), snapshot(&fx.home));
        for name in NAMES {
            for input in [
                &b"SECRETMARKER not json"[..],
                &valid_input("SessionStart")[..],
            ] {
                let r = fx.hook(name, input);
                assert_eq!(
                    (r.code, r.stdout.as_str(), r.stderr.as_str()),
                    (0, "", ""),
                    "{what} {name}"
                );
            }
        }
        assert_eq!(
            before,
            (snapshot(&fx.project), snapshot(&fx.home)),
            "{what}"
        );
    }
}

#[test]
fn ac3_claude_project_dir_wins_over_the_working_folder() {
    // The working folder is a MoAI project, CLAUDE_PROJECT_DIR is not: silent, no receipt.
    let fx = Fx::new(true, true);
    let plain = fx.project.parent().unwrap().join("plain");
    fs::create_dir_all(&plain).unwrap();
    let r = fx.run(
        &["hook", "session-context"],
        &fx.project,
        b"not json",
        &[("CLAUDE_PROJECT_DIR", plain.to_str().unwrap())],
    );
    assert_eq!((r.code, r.stdout.as_str(), r.stderr.as_str()), (0, "", ""));
    assert!(receipt_lines(&fx.receipt()).is_empty());
    // The working folder is not a project, CLAUDE_PROJECT_DIR is: bad input is recorded.
    let r = fx.run(
        &["hook", "session-context"],
        &plain,
        b"not json",
        &[("CLAUDE_PROJECT_DIR", fx.project.to_str().unwrap())],
    );
    assert_eq!((r.code, r.stdout.as_str()), (0, ""), "{}", r.stderr);
    assert_eq!(receipt_lines(&fx.receipt()).len(), 1);
}

/// The one deny object of SPEC-20261126-hook-unreadable-input REQ-1: one line on stdout holding
/// one object whose only key is `hookSpecificOutput`, with exactly its three fields. Returns the
/// reason text.
fn deny_reason(stdout: &str, what: &str) -> String {
    assert!(
        stdout.ends_with('\n') && stdout.lines().count() == 1,
        "{what}: no deny object on one line of stdout, got {stdout:?}"
    );
    let v: Value = serde_json::from_str(stdout.trim_end())
        .unwrap_or_else(|e| panic!("{what}: stdout is not one JSON object ({e}): {stdout:?}"));
    let reason = v["hookSpecificOutput"]["permissionDecisionReason"]
        .as_str()
        .unwrap_or_else(|| panic!("{what}: no reason text: {stdout:?}"))
        .to_owned();
    assert_eq!(
        v,
        serde_json::json!({"hookSpecificOutput": {
            "hookEventName": "PreToolUse",
            "permissionDecision": "deny",
            "permissionDecisionReason": reason,
        }}),
        "{what}"
    );
    reason
}

/// A deny object whose reason keeps the contract for this name and receipt word (REQ-3): the class
/// of the name, the cause, one allowed way and the request to tell the user, under 400 bytes and
/// without the marker of the input. The over-limit reason of `write-boundary` is the earlier one,
/// word for word (REQ-2). Returns the reason.
fn assert_reason(name: &str, word: &str, what: &str, stdout: &str) -> String {
    let reason = deny_reason(stdout, what);
    if (name, word) == ("write-boundary", TOO_LARGE) {
        assert_eq!(reason, WRITE_OVER_LIMIT, "{what}");
    } else {
        let class = match name {
            "bash-boundary" => "moai-boundary/unreadable:",
            "write-boundary" => "moai-boundary/protected-path:",
            other => panic!("{what}: {other} denies nothing"),
        };
        let (cause, way) = match word {
            "not_utf8" => ("not valid UTF-8", "send the call again"),
            "not_json" => ("not valid JSON", "send the call again"),
            "not_object" => ("not a JSON object", "send the call again"),
            TOO_LARGE => ("4 MiB", "shorter command"),
            other => panic!("{what}: no cause for the word {other}"),
        };
        assert!(reason.starts_with(class), "{what}: {reason:?}");
        for part in [cause, way, "tell the user"] {
            assert!(reason.contains(part), "{what}: {reason:?} lacks {part:?}");
        }
    }
    assert!(reason.len() < 400, "{what}: {} bytes", reason.len());
    assert!(!reason.contains(MARKER), "{what}: {reason:?}");
    reason
}

/// What bad input answers: nothing on stdout for `session-context`; the two `PreToolUse` names
/// deny it (SPEC-20261126-hook-unreadable-input REQ-1, REQ-2).
fn assert_answer(name: &str, reason: &str, stdout: &str) {
    if name == "session-context" {
        assert_eq!(stdout, "", "{name} {reason}");
    } else {
        assert_reason(name, reason, &format!("{name} {reason}"), stdout);
    }
}

/// AC-4: bad input of a registered name in a MoAI project adds exactly one receipt line with the
/// name, a reason word and a time, and none of the input bytes.
#[test]
fn ac4_bad_input_adds_one_receipt_line_without_the_input() {
    let big = object_of_len(STDIN_CAP + 1);
    let cases: Vec<(&str, Vec<u8>, &str)> = vec![
        ("empty", Vec::new(), "not_json"),
        ("text", b"SECRETMARKER not json".to_vec(), "not_json"),
        ("array", br#"["SECRETMARKER"]"#.to_vec(), "not_object"),
        ("scalar", b"\"SECRETMARKER\"".to_vec(), "not_object"),
        (
            "not utf-8",
            vec![
                b'{', 0xff, 0xfe, b'S', b'E', b'C', b'R', b'E', b'T', b'M', b'A', b'R', b'K', b'E',
                b'R',
            ],
            "not_utf8",
        ),
        ("too large", big, "stdin_too_large"),
    ];
    for name in NAMES {
        let fx = Fx::new(true, true);
        for (i, (what, input, reason)) in cases.iter().enumerate() {
            let r = fx.hook(name, input);
            assert_eq!(r.code, 0, "{name} {what}: {}", r.stderr);
            assert_answer(name, reason, &r.stdout);
            assert_eq!(
                r.stderr, "",
                "{name} {what}: a recorded failure leaves stderr empty"
            );
            let lines = receipt_lines(&fx.receipt());
            assert_eq!(lines.len(), i + 1, "{name} {what}");
            let v: Value = serde_json::from_str(lines.last().unwrap()).unwrap();
            assert_eq!(v["name"], name, "{what}");
            assert_eq!(v["reason"], *reason, "{what}");
            assert!(
                v["at_unix_ms"].as_u64().is_some_and(|t| t > 0),
                "{what}: {v}"
            );
        }
        let text = fs::read_to_string(fx.receipt()).unwrap();
        assert!(
            !text.contains("SECRETMARKER"),
            "the input must not reach the receipt"
        );
    }
}

/// AC-4: the stdin limit is the 4 MiB itself: an object of exactly 4194304 bytes is accepted
/// without a receipt, and one more byte is the "too large" failure, which the two `PreToolUse`
/// names deny.
#[test]
fn ac4_the_stdin_limit_takes_exactly_the_limit_and_refuses_one_more() {
    assert_eq!(STDIN_CAP, 4_194_304);
    for name in NAMES {
        let fx = Fx::new(true, true);
        let at_limit = fx.hook(name, &object_of_len(STDIN_CAP));
        // `session-context` answers a context text (SPEC-20261123-hook-session-start) and, in
        // this fixture without a ledger file, one `ledger_absent` receipt line; the other two
        // names stay silent. In every case the input of exactly the limit is no "too large".
        if name == "session-context" {
            assert_eq!(
                (at_limit.code, at_limit.stderr.as_str()),
                (0, ""),
                "{name} at the limit"
            );
            assert!(
                at_limit
                    .stdout
                    .contains("\"hookEventName\":\"SessionStart\""),
                "{name}: {}",
                at_limit.stdout
            );
        } else {
            assert_eq!(
                (
                    at_limit.code,
                    at_limit.stdout.as_str(),
                    at_limit.stderr.as_str()
                ),
                (0, "", ""),
                "{name} at the limit"
            );
        }
        let too_large = |fx: &Fx| {
            receipt_lines(&fx.receipt())
                .iter()
                .filter(|l| l.contains("stdin_too_large"))
                .count()
        };
        assert_eq!(
            too_large(&fx),
            0,
            "{name}: an input of exactly the limit is not a failure"
        );
        let before = receipt_lines(&fx.receipt()).len();
        let over = fx.hook(name, &object_of_len(STDIN_CAP + 1));
        assert_eq!(
            (over.code, over.stderr.as_str()),
            (0, ""),
            "{name} one byte over"
        );
        assert_answer(name, "stdin_too_large", &over.stdout);
        let lines = receipt_lines(&fx.receipt());
        assert_eq!(
            lines.len(),
            before + 1,
            "{name}: one byte over is one receipt line"
        );
        let v: Value = serde_json::from_str(lines.last().unwrap()).unwrap();
        assert_eq!(v["reason"], "stdin_too_large", "{name}");
        assert_eq!(too_large(&fx), 1, "{name}");
    }
}

/// AC-5: with no registered project folder the failure is one stderr line, nothing is created,
/// and the call is denied all the same (SPEC-20261126-hook-unreadable-input REQ-5).
#[test]
fn ac5_without_a_project_folder_the_failure_is_one_stderr_line() {
    let fx = Fx::new(true, false);
    let before = snapshot(&fx.home);
    let r = fx.hook("write-boundary", b"SECRETMARKER not json");
    assert_eq!(r.code, 0, "{}", r.stderr);
    assert_answer("write-boundary", "not_json", &r.stdout);
    assert_eq!(r.stderr.lines().count(), 1, "{:?}", r.stderr);
    assert!(!r.stderr.contains("SECRETMARKER"));
    assert!(
        r.stderr.contains("write-boundary") && r.stderr.contains("not_json"),
        "the line names the hook and the reason: {:?}",
        r.stderr
    );
    assert_eq!(
        before,
        snapshot(&fx.home),
        "no folder and no file may be created in HOME"
    );
}

/// AC-5: a receipt path that is a link is not followed, and a `project_id` that could leave the
/// projects folder is refused, each with one stderr line and nothing written outside.
#[test]
fn ac5_a_receipt_link_is_not_followed_and_a_hostile_project_id_is_refused() {
    let fx = Fx::new(true, true);
    let outside = fx.project.parent().unwrap().join("outside");
    fs::write(&outside, "original").unwrap();
    std::os::unix::fs::symlink(&outside, fx.receipt()).unwrap();
    let r = fx.hook("session-context", b"not json");
    assert_eq!((r.code, r.stdout.as_str()), (0, ""), "{}", r.stderr);
    assert_eq!(r.stderr.lines().count(), 1, "{:?}", r.stderr);
    assert_eq!(fs::read_to_string(&outside).unwrap(), "original");

    fs::remove_file(fx.receipt()).unwrap();
    fs::write(
        project_file(&fx.project),
        r#"{"project_id":"../../outside"}"#,
    )
    .unwrap();
    let root = fx.project.parent().unwrap();
    let before = snapshot(root);
    let r = fx.hook("session-context", b"not json");
    assert_eq!((r.code, r.stdout.as_str()), (0, ""), "{}", r.stderr);
    assert_eq!(r.stderr.lines().count(), 1, "{:?}", r.stderr);
    assert!(r.stderr.contains("portable"), "{:?}", r.stderr);
    assert_eq!(before, snapshot(root), "nothing may be written anywhere");
}

/// AC-6: a receipt file at 256 KiB takes no more lines and warns once; one byte below it still
/// takes the line.
#[test]
fn ac6_a_full_receipt_file_warns_and_appends_nothing() {
    let fx = Fx::new(true, true);
    fs::write(fx.receipt(), vec![b'\n'; RECEIPT_CAP]).unwrap();
    let r = fx.hook("session-context", b"not json");
    assert_eq!((r.code, r.stdout.as_str()), (0, ""));
    assert_eq!(r.stderr.lines().count(), 1, "{:?}", r.stderr);
    assert!(r.stderr.contains("262144"), "{}", r.stderr);
    assert_eq!(fs::read(fx.receipt()).unwrap().len(), RECEIPT_CAP);

    fs::write(fx.receipt(), vec![b'\n'; RECEIPT_CAP - 1]).unwrap();
    let r = fx.hook("session-context", b"not json");
    assert_eq!((r.code, r.stdout.as_str(), r.stderr.as_str()), (0, "", ""));
    assert!(fs::read(fx.receipt()).unwrap().len() > RECEIPT_CAP - 1);
}

/// AC-7: no Orca variable changes the answer, no settings file appears, and the source of the
/// entry names no process start, network or settings path and reads only two variables.
#[test]
fn ac7_orca_variables_change_nothing_and_no_settings_file_appears() {
    let orca = [
        ("ORCA_SESSION_ID", "s-1"),
        ("ORCA_PROFILE", "p"),
        ("ORCA_TERMINAL_ID", "t"),
        ("CLAUDE_CONFIG_DIR", "/nonexistent-config-dir"),
    ];
    let plain = Fx::new(true, true);
    let with = Fx::new(true, true);
    for (input, label) in [
        (&b"not json"[..], "bad"),
        (&valid_input("SessionStart")[..], "good"),
    ] {
        let a = plain.run(&["hook", "bash-boundary"], &plain.project, input, &[]);
        let b = with.run(&["hook", "bash-boundary"], &with.project, input, &orca);
        assert_eq!(
            (a.code, &a.stdout, &a.stderr),
            (b.code, &b.stdout, &b.stderr),
            "{label}"
        );
    }
    let lines = |fx: &Fx| -> Vec<Value> {
        receipt_lines(&fx.receipt())
            .iter()
            .map(|l| {
                let mut v: Value = serde_json::from_str(l).unwrap();
                v.as_object_mut().unwrap().remove("at_unix_ms");
                v
            })
            .collect()
    };
    assert_eq!(lines(&plain), lines(&with));
    for fx in [&plain, &with] {
        for root in [&fx.project, &fx.home] {
            let bad: Vec<String> = snapshot(root)
                .into_keys()
                .filter(|k| k.contains("settings"))
                .collect();
            assert!(bad.is_empty(), "{bad:?}");
        }
    }

    let source = include_str!("../src/hook_entry.rs");
    for word in [
        "Command::new",
        "std::process::Command",
        "TcpStream",
        "UdpSocket",
        "reqwest",
        "ureq",
        "hyper",
        "settings.json",
        "settings.local.json",
        "ORCA",
        "env::vars",
    ] {
        assert!(!source.contains(word), "src/hook_entry.rs names `{word}`");
    }
    let mut variables = Vec::new();
    for marker in ["var_os(\"", "var(\""] {
        let mut rest = source;
        while let Some(at) = rest.find(marker) {
            rest = &rest[at + marker.len()..];
            variables.push(rest[..rest.find('"').unwrap()].to_owned());
        }
    }
    variables.sort();
    variables.dedup();
    assert_eq!(variables, ["CLAUDE_PROJECT_DIR", "HOME"]);
}

/// AC-8 (SPEC-20261126-hook-write-boundary): a `hook_event_name` that differs from the event the
/// name belongs to is judged by nobody: nothing is printed and one `event_mismatch` receipt line
/// is written without any input byte; the matching name and an absent name write no such line.
#[test]
fn ac8_a_call_of_another_event_is_refused_with_one_receipt_line() {
    let event_of = |name: &str| match name {
        "session-context" => "SessionStart",
        _ => "PreToolUse",
    };
    for name in NAMES {
        let fx = Fx::new(true, true);
        let right = event_of(name);
        let wrong = if right == "PreToolUse" {
            "SessionStart"
        } else {
            "PreToolUse"
        };
        let make = |event: Option<Value>| {
            let mut object = serde_json::json!({
                "session_id": "s1",
                "cwd": fx.project,
                "tool_name": "Write",
                "tool_input": { "file_path": "/x/SECRETMARKER", "content": "SECRETMARKER" },
            });
            if let Some(event) = event {
                object["hook_event_name"] = event;
            }
            object.to_string().into_bytes()
        };
        // The matching and the absent event name write no mismatch line.
        for input in [make(Some(Value::from(right))), make(None)] {
            let r = fx.hook(name, &input);
            assert_eq!(r.code, 0, "{name}: {}", r.stderr);
            assert!(
                receipt_lines(&fx.receipt())
                    .iter()
                    .all(|l| !l.contains("event_mismatch")),
                "{name}"
            );
        }
        // Another event name, a number, a null and an array each write exactly one line.
        let wrong_inputs = [
            make(Some(Value::from(wrong))),
            make(Some(Value::from(7))),
            make(Some(Value::Null)),
            make(Some(serde_json::json!([right]))),
        ];
        for (i, input) in wrong_inputs.iter().enumerate() {
            let r = fx.hook(name, input);
            assert_eq!(r.code, 0, "{name} {i}: {}", r.stderr);
            assert_eq!(r.stdout, "", "{name} {i}");
            assert_eq!(r.stderr, "", "{name} {i}");
            // `session-context` may also write its own ledger line; only the mismatch lines count.
            let lines: Vec<String> = receipt_lines(&fx.receipt())
                .into_iter()
                .filter(|l| l.contains("event_mismatch"))
                .collect();
            assert_eq!(lines.len(), i + 1, "{name} {i}");
            let v: Value = serde_json::from_str(lines.last().unwrap()).unwrap();
            assert_eq!(v["name"], name);
            assert_eq!(v["reason"], "event_mismatch");
        }
        let text = fs::read_to_string(fx.receipt()).unwrap();
        assert!(
            !text.contains("SECRETMARKER"),
            "the input must not reach the receipt"
        );
    }
}

// ---------------------------------------------------- SPEC-20261126-hook-unreadable-input

/// An input the entry refuses in a MoAI project: a label, the bytes and the reason word of its
/// receipt line.
struct Refused {
    what: &'static str,
    input: Vec<u8>,
    word: &'static str,
}

fn refused_input(what: &'static str, input: Vec<u8>, word: &'static str) -> Refused {
    Refused { what, input, word }
}

fn empty_and_text() -> Vec<Refused> {
    vec![
        refused_input("empty input", Vec::new(), "not_json"),
        refused_input("plain text", b"SECRETMARKER not json".to_vec(), "not_json"),
    ]
}

/// A tool call whose command holds the first half of a surrogate pair as a JSON escape.
fn lone_surrogate() -> Refused {
    let call = br#"{"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":"SECRETMARKER \ud800"}}"#;
    refused_input("a lone surrogate escape", call.to_vec(), "not_json")
}

/// Objects `depth` deep with the marker at the bottom.
fn nested(depth: usize) -> Vec<u8> {
    format!(
        "{}\"{MARKER}\"{}",
        "{\"k\":".repeat(depth),
        "}".repeat(depth)
    )
    .into_bytes()
}

/// One object deeper than the 128 levels `serde_json` reads.
fn deep_nesting() -> Refused {
    refused_input("objects nested 129 deep", nested(129), "not_json")
}

fn array_and_scalar() -> Vec<Refused> {
    vec![
        refused_input(
            "a JSON array",
            br#"["SECRETMARKER"]"#.to_vec(),
            "not_object",
        ),
        refused_input("a JSON scalar", br#""SECRETMARKER""#.to_vec(), "not_object"),
    ]
}

fn invalid_utf8() -> Refused {
    let call = b"{\"k\":\"SECRETMARKER\xff\xfe\"}";
    refused_input("invalid UTF-8", call.to_vec(), "not_utf8")
}

/// Every input of AC-1: what the entry cannot read.
fn unreadable() -> Vec<Refused> {
    let mut cases = empty_and_text();
    cases.push(lone_surrogate());
    cases.push(deep_nesting());
    cases.extend(array_and_scalar());
    cases.push(invalid_utf8());
    cases
}

/// A readable object one byte over the limit.
fn one_over_the_limit() -> Refused {
    refused_input(
        "one byte over the limit",
        object_of_len(STDIN_CAP + 1),
        TOO_LARGE,
    )
}

/// What the two `PreToolUse` names refuse: the inputs of AC-1 and the over-limit input.
fn refused() -> Vec<Refused> {
    let mut cases = unreadable();
    cases.push(one_over_the_limit());
    cases
}

fn receipt_words(path: &Path) -> Vec<String> {
    receipt_lines(path)
        .iter()
        .map(|l| {
            let v: Value = serde_json::from_str(l).unwrap();
            v["reason"].as_str().unwrap().to_owned()
        })
        .collect()
}

/// Runs every case under every name of `names` in a registered MoAI project. Every run must be
/// answered: the silent ones are listed together, each with what the entry recorded, so one
/// failure names them all. Each answer then keeps the reason contract, with stderr empty.
/// Returns (name, receipt word, reason) per run.
fn denials(names: &[&'static str], cases: &[Refused]) -> Vec<(&'static str, &'static str, String)> {
    let fx = Fx::new(true, true);
    let mut runs = Vec::new();
    let mut silent = Vec::new();
    for &name in names {
        for case in cases {
            let _ = fs::remove_file(fx.receipt());
            let r = fx.hook(name, &case.input);
            if r.stdout.is_empty() {
                silent.push(format!(
                    "{name}, {}: exit {}, stdout empty, stderr {:?}, receipt {:?}",
                    case.what,
                    r.code,
                    r.stderr,
                    receipt_words(&fx.receipt())
                ));
            }
            runs.push((name, case, r));
        }
    }
    assert!(
        silent.is_empty(),
        "no denial, so the tool call would run:\n{}",
        silent.join("\n")
    );
    runs.into_iter()
        .map(|(name, case, r)| {
            let what = format!("{name}, {}", case.what);
            assert_eq!((r.code, r.stderr.as_str()), (0, ""), "{what}");
            let reason = assert_reason(name, case.word, &what, &r.stdout);
            (name, case.word, reason)
        })
        .collect()
}

/// SPEC-20261126-hook-unreadable-input AC-1.
#[test]
fn unreadable_ac1_empty_input_and_plain_text_are_denied_as_not_json() {
    denials(&PRE_TOOL_USE, &empty_and_text());
}

/// SPEC-20261126-hook-unreadable-input AC-1.
#[test]
fn unreadable_ac1_a_lone_surrogate_escape_is_denied_as_not_json() {
    let case = lone_surrogate();
    // The lone half alone makes the fixture unreadable: with its other half it reads.
    let paired = String::from_utf8(case.input.clone())
        .unwrap()
        .replace(r"\ud800", r"\ud800\udc00");
    assert!(serde_json::from_str::<Value>(&paired).is_ok(), "{paired}");
    denials(&PRE_TOOL_USE, &[case]);
}

/// SPEC-20261126-hook-unreadable-input AC-1.
#[test]
fn unreadable_ac1_nesting_past_the_parser_depth_is_denied_as_not_json() {
    // The depth alone makes the fixture unreadable: one the parser still takes reads.
    assert!(serde_json::from_slice::<Value>(&nested(127)).is_ok());
    assert!(serde_json::from_slice::<Value>(&nested(129)).is_err());
    denials(&PRE_TOOL_USE, &[deep_nesting()]);
}

/// SPEC-20261126-hook-unreadable-input AC-1.
#[test]
fn unreadable_ac1_a_json_array_and_a_scalar_are_denied_as_not_object() {
    denials(&PRE_TOOL_USE, &array_and_scalar());
}

/// SPEC-20261126-hook-unreadable-input AC-1.
#[test]
fn unreadable_ac1_invalid_utf8_is_denied_as_not_utf8() {
    denials(&PRE_TOOL_USE, &[invalid_utf8()]);
}

/// SPEC-20261126-hook-unreadable-input AC-1: the text is fixed per name and cause.
#[test]
fn unreadable_ac1_the_reason_is_one_text_per_cause_and_differs_between_causes() {
    let answers = denials(&PRE_TOOL_USE, &unreadable());
    for name in PRE_TOOL_USE {
        let mut by_word: BTreeMap<&str, BTreeSet<&str>> = BTreeMap::new();
        for (_, word, reason) in answers.iter().filter(|a| a.0 == name) {
            by_word.entry(word).or_default().insert(reason);
        }
        assert_eq!(by_word.len(), 3, "{name}: {by_word:?}");
        assert!(
            by_word.values().all(|texts| texts.len() == 1),
            "{name}: one text per cause: {by_word:?}"
        );
        let texts: BTreeSet<&str> = by_word.values().flatten().copied().collect();
        assert_eq!(texts.len(), 3, "{name}: causes share a text: {by_word:?}");
    }
}

/// SPEC-20261126-hook-unreadable-input AC-1: the denial needs no variable but HOME.
#[test]
fn unreadable_ac1_a_cleared_environment_denies_in_a_project_and_is_silent_outside() {
    let bare = |fx: &Fx, name: &str| {
        let mut cmd = Command::new(common::MOAI_BIN);
        cmd.args(["hook", name])
            .current_dir(&fx.project)
            .clean_env()
            .env("HOME", &fx.home);
        finish(&mut cmd, b"SECRETMARKER not json")
    };

    let plain = Fx::new(false, true);
    let root = plain.project.parent().unwrap();
    let before = snapshot(root);
    for name in PRE_TOOL_USE {
        let r = bare(&plain, name);
        assert_eq!(
            (r.code, r.stdout.as_str(), r.stderr.as_str()),
            (0, "", ""),
            "{name} outside a project"
        );
    }
    assert_eq!(before, snapshot(root), "nothing may be written");

    let moai = Fx::new(true, true);
    for name in PRE_TOOL_USE {
        let r = bare(&moai, name);
        assert_eq!((r.code, r.stderr.as_str()), (0, ""), "{name}");
        assert_reason(name, "not_json", name, &r.stdout);
    }
}

/// SPEC-20261126-hook-unreadable-input AC-2: the Bash name denies one byte over the limit.
#[test]
fn unreadable_ac2_a_bash_input_over_the_limit_is_denied_and_names_4_mib() {
    assert_eq!(STDIN_CAP + 1, 4_194_305);
    let cases = [
        one_over_the_limit(),
        refused_input(
            "a hundred bytes over the limit",
            object_of_len(STDIN_CAP + 100),
            TOO_LARGE,
        ),
        refused_input("plain text", b"SECRETMARKER not json".to_vec(), "not_json"),
    ];
    let answers = denials(&["bash-boundary"], &cases);
    assert_eq!(
        answers[0].2, answers[1].2,
        "one text for the same cause and name"
    );
    assert_ne!(answers[0].2, answers[2].2, "another cause has another text");
}

/// SPEC-20261126-hook-unreadable-input AC-2: an object of exactly the limit is judged as before.
#[test]
fn unreadable_ac2_a_bash_input_of_exactly_the_limit_gets_no_denial() {
    assert_eq!(STDIN_CAP, 4_194_304);
    let fx = Fx::new(true, true);
    let r = fx.hook("bash-boundary", &object_of_len(STDIN_CAP));
    assert_eq!((r.code, r.stdout.as_str(), r.stderr.as_str()), (0, "", ""));
    assert_eq!(receipt_words(&fx.receipt()), Vec::<String>::new());
}

/// SPEC-20261126-hook-unreadable-input AC-2: the write over-limit reason does not change.
#[test]
fn unreadable_ac2_the_write_over_limit_reason_is_todays() {
    let answers = denials(&["write-boundary"], &[one_over_the_limit()]);
    assert_eq!(answers[0].2, WRITE_OVER_LIMIT);
}

/// SPEC-20261126-hook-unreadable-input AC-3: each denial is recorded once, without input bytes.
#[test]
fn unreadable_ac3_each_denial_adds_one_receipt_line_with_the_reason_word() {
    for name in PRE_TOOL_USE {
        let fx = Fx::new(true, true);
        let mut answers = Vec::new();
        for (i, case) in refused().into_iter().enumerate() {
            let what = format!("{name}, {}", case.what);
            let r = fx.hook(name, &case.input);
            assert_eq!((r.code, r.stderr.as_str()), (0, ""), "{what}");
            let lines = receipt_lines(&fx.receipt());
            assert_eq!(lines.len(), i + 1, "{what}");
            let v: Value = serde_json::from_str(lines.last().unwrap()).unwrap();
            let keys: BTreeSet<&str> = v.as_object().unwrap().keys().map(String::as_str).collect();
            assert_eq!(
                keys,
                BTreeSet::from(["at_unix_ms", "name", "reason", "schema"]),
                "{what}"
            );
            assert_eq!(v["name"], name, "{what}");
            assert_eq!(v["reason"], case.word, "{what}");
            assert!(
                v["at_unix_ms"].as_u64().is_some_and(|t| t > 0),
                "{what}: {v}"
            );
            answers.push((what, case.word, r.stdout));
        }
        let text = fs::read_to_string(fx.receipt()).unwrap();
        assert!(
            !text.contains(MARKER),
            "the input must not reach the receipt"
        );
        for (what, word, stdout) in answers {
            assert_reason(name, word, &what, &stdout);
        }
    }
}

/// With no place for the receipt the call is still refused: the deny object on stdout, one stderr
/// line naming the hook, the reason word and `why`, and nothing changed under the temporary root
/// (the project, HOME and whatever lies beside them).
fn assert_denied_without_a_receipt(fx: &Fx, why: &str) {
    let root = fx.project.parent().unwrap();
    let before = snapshot(root);
    for name in PRE_TOOL_USE {
        for case in refused() {
            let what = format!("{name}, {}", case.what);
            let r = fx.hook(name, &case.input);
            assert_eq!(r.code, 0, "{what}: {}", r.stderr);
            assert_reason(name, case.word, &what, &r.stdout);
            assert_eq!(r.stderr.lines().count(), 1, "{what}: {:?}", r.stderr);
            for part in [name, case.word, why] {
                assert!(
                    r.stderr.contains(part),
                    "{what}: {:?} lacks {part:?}",
                    r.stderr
                );
            }
            assert!(!r.stderr.contains(MARKER), "{what}: {:?}", r.stderr);
        }
    }
    assert_eq!(before, snapshot(root), "nothing may be written anywhere");
}

/// SPEC-20261126-hook-unreadable-input AC-3.
#[test]
fn unreadable_ac3_without_a_registered_project_folder_the_call_is_denied_with_one_stderr_line() {
    assert_denied_without_a_receipt(&Fx::new(true, false), "not registered");
}

/// SPEC-20261126-hook-unreadable-input AC-3.
#[test]
fn unreadable_ac3_a_receipt_link_is_not_followed_and_the_call_is_denied() {
    let fx = Fx::new(true, true);
    let outside = fx.project.parent().unwrap().join("outside");
    fs::write(&outside, "original").unwrap();
    std::os::unix::fs::symlink(&outside, fx.receipt()).unwrap();
    assert_denied_without_a_receipt(&fx, "link");
}

/// SPEC-20261126-hook-unreadable-input AC-3.
#[test]
fn unreadable_ac3_a_full_receipt_log_still_denies_with_one_stderr_line() {
    let fx = Fx::new(true, true);
    fs::write(fx.receipt(), vec![b'\n'; RECEIPT_CAP]).unwrap();
    assert_denied_without_a_receipt(&fx, "262144");
}

type Prepare = fn(&Path);

/// The ways a folder is no MoAI project: each one prepares an empty folder.
fn not_projects() -> [(&'static str, Prepare); 3] {
    [
        ("no state folder", |_p| {}),
        ("project.yaml is a folder", |p| {
            fs::create_dir_all(project_file(p)).unwrap()
        }),
        ("project.yaml is a link to nothing", |p| {
            fs::create_dir_all(state(p)).unwrap();
            std::os::unix::fs::symlink(p.join("no-such-target"), project_file(p)).unwrap();
        }),
    ]
}

/// Every input of AC-1 and AC-2 gets nothing on either stream from any name.
fn assert_every_name_is_silent(fx: &Fx, env: &[(&str, &str)], what: &str) {
    let mut cases = refused();
    cases.push(refused_input(
        "exactly the limit",
        object_of_len(STDIN_CAP),
        "",
    ));
    for name in NAMES {
        for case in &cases {
            let r = fx.run(&["hook", name], &fx.project, &case.input, env);
            assert_eq!(
                (r.code, r.stdout.as_str(), r.stderr.as_str()),
                (0, "", ""),
                "{what}, {name}, {}",
                case.what
            );
        }
    }
}

/// SPEC-20261126-hook-unreadable-input AC-4: the working folder is no MoAI project.
#[test]
fn unreadable_ac4_outside_a_project_every_input_is_silent_and_writes_nothing() {
    for (what, make) in not_projects() {
        let fx = Fx::new(false, true);
        make(&fx.project);
        let root = fx.project.parent().unwrap();
        let before = snapshot(root);
        assert_every_name_is_silent(&fx, &[], what);
        assert_eq!(before, snapshot(root), "{what}");
    }
}

/// SPEC-20261126-hook-unreadable-input AC-4: `CLAUDE_PROJECT_DIR` names a folder that is no MoAI
/// project, though the working folder is one.
#[test]
fn unreadable_ac4_claude_project_dir_naming_such_a_folder_is_silent() {
    for (what, make) in not_projects() {
        let fx = Fx::new(true, true);
        let root = fx.project.parent().unwrap();
        let named = root.join("named");
        fs::create_dir_all(&named).unwrap();
        make(&named);
        let before = snapshot(root);
        let env = [("CLAUDE_PROJECT_DIR", named.to_str().unwrap())];
        assert_every_name_is_silent(&fx, &env, what);
        assert_eq!(before, snapshot(root), "{what}");
    }
}

/// SPEC-20261126-hook-unreadable-input AC-5: a SessionStart hook decides about no tool call.
#[test]
fn unreadable_ac5_session_context_prints_nothing_for_every_refused_input() {
    let fx = Fx::new(true, true);
    for case in refused() {
        let _ = fs::remove_file(fx.receipt());
        let r = fx.hook("session-context", &case.input);
        assert_eq!(
            (r.code, r.stdout.as_str(), r.stderr.as_str()),
            (0, "", ""),
            "{}",
            case.what
        );
        assert_eq!(receipt_words(&fx.receipt()), [case.word], "{}", case.what);
    }
    // The object of exactly the limit is readable, so it gets its context, never a decision.
    let r = fx.hook("session-context", &object_of_len(STDIN_CAP));
    assert_eq!((r.code, r.stderr.as_str()), (0, ""));
    assert!(!r.stdout.contains("permissionDecision"), "{}", r.stdout);
}

/// SPEC-20261126-hook-unreadable-input AC-5.
#[test]
fn unreadable_ac5_an_object_of_another_event_is_silent_for_every_name() {
    for name in NAMES {
        let fx = Fx::new(true, true);
        let other = if name == "session-context" {
            "PreToolUse"
        } else {
            "SessionStart"
        };
        let r = fx.hook(name, &valid_input(other));
        assert_eq!(
            (r.code, r.stdout.as_str(), r.stderr.as_str()),
            (0, "", ""),
            "{name}"
        );
        assert_eq!(receipt_words(&fx.receipt()), ["event_mismatch"], "{name}");
    }
}

/// SPEC-20261126-hook-unreadable-input AC-5: a readable object its handler cannot judge.
#[test]
fn unreadable_ac5_a_readable_object_without_tool_input_is_silent() {
    for (name, tool) in [("bash-boundary", "Bash"), ("write-boundary", "Write")] {
        let fx = Fx::new(true, true);
        let call = serde_json::json!({
            "hook_event_name": "PreToolUse",
            "session_id": "s1",
            "cwd": fx.project,
            "tool_name": tool,
        });
        for input in [call.to_string(), "{}".to_owned()] {
            let r = fx.hook(name, input.as_bytes());
            assert_eq!(
                (r.code, r.stdout.as_str(), r.stderr.as_str()),
                (0, "", ""),
                "{name}: {input}"
            );
        }
        assert_eq!(receipt_words(&fx.receipt()), Vec::<String>::new(), "{name}");
    }
}
