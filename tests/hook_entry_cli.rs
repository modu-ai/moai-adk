//! SPEC-20261123-hook-entry-skeleton: the `moai hook <name>` entry point. Three names are
//! registered (`session-context`, `bash-boundary`, `write-boundary`) and every handler of this
//! slice answers no decision. The entry reads stdin once, stays silent outside a MoAI project,
//! exits 0 for any name it does not know, and records one receipt line when the input of a
//! registered name is bad. Every case runs the real binary in a temporary project and HOME.
#![cfg(unix)]
mod common;

use serde_json::Value;
use std::collections::BTreeMap;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

const NAMES: [&str; 3] = ["session-context", "bash-boundary", "write-boundary"];
const RECEIPT: &str = "hook-entry.jsonl";
const RECEIPT_CAP: usize = 262_144;
const STDIN_CAP: usize = 4 * 1024 * 1024;

struct Run {
    code: i32,
    stdout: String,
    stderr: String,
}

/// A temporary project folder and HOME. `registered` makes `<HOME>/.moai/projects/demo/`.
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
            fs::create_dir_all(project.join(".moai")).unwrap();
            fs::write(
                project.join(".moai/project.yaml"),
                "{\n  \"schema_version\": 2,\n  \"project_id\": \"demo\"\n}\n",
            )
            .unwrap();
        }
        if registered {
            fs::create_dir_all(home.join(".moai/projects/demo")).unwrap();
        }
        Fx {
            _tmp: tmp,
            project,
            home,
        }
    }

    fn receipt(&self) -> PathBuf {
        self.home.join(".moai/projects/demo").join(RECEIPT)
    }

    fn run(&self, args: &[&str], cwd: &Path, stdin: &[u8], env: &[(&str, &str)]) -> Run {
        let mut cmd = Command::new(common::MOAI_BIN);
        cmd.args(args)
            .current_dir(cwd)
            .env("HOME", &self.home)
            .env_remove("CLAUDE_PROJECT_DIR")
            .env_remove("MOAI_FORGE_GH_BIN")
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped());
        for (k, v) in env {
            cmd.env(k, v);
        }
        let mut child = cmd.spawn().unwrap();
        common::feed_stdin(&mut child, stdin);
        let out = child.wait_with_output().unwrap();
        Run {
            code: out.status.code().unwrap(),
            stdout: String::from_utf8_lossy(&out.stdout).into_owned(),
            stderr: String::from_utf8_lossy(&out.stderr).into_owned(),
        }
    }

    fn hook(&self, name: &str, stdin: &[u8]) -> Run {
        self.run(&["hook", name], &self.project, stdin, &[])
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
            let rel = path.strip_prefix(base).unwrap().to_string_lossy().into_owned();
            let meta = fs::symlink_metadata(&path).unwrap();
            if meta.file_type().is_symlink() {
                let target = fs::read_link(&path).unwrap();
                out.insert(format!("{rel} -> link"), target.to_string_lossy().as_bytes().to_vec());
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
    for (name, event) in [
        ("session-context", "SessionStart"),
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
        for name in ["pre-tool", "nonsense", "SESSION-CONTEXT", "session-context-x", "session_context"] {
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
    let broken: [(&str, fn(&Path)); 3] = [
        ("no .moai folder", |_p| {}),
        ("project.yaml is a folder", |p| {
            fs::create_dir_all(p.join(".moai/project.yaml")).unwrap()
        }),
        ("project.yaml is a link to nothing", |p| {
            fs::create_dir_all(p.join(".moai")).unwrap();
            std::os::unix::fs::symlink(p.join("no-such-target"), p.join(".moai/project.yaml"))
                .unwrap();
        }),
    ];
    for (what, make) in broken {
        let fx = Fx::new(false, true);
        make(&fx.project);
        let before = (snapshot(&fx.project), snapshot(&fx.home));
        for name in NAMES {
            for input in [&b"SECRETMARKER not json"[..], &valid_input("SessionStart")[..]] {
                let r = fx.hook(name, input);
                assert_eq!((r.code, r.stdout.as_str(), r.stderr.as_str()), (0, "", ""), "{what} {name}");
            }
        }
        assert_eq!(before, (snapshot(&fx.project), snapshot(&fx.home)), "{what}");
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

/// AC-4: bad input of a registered name in a MoAI project adds exactly one receipt line with the
/// name, a reason word and a time, and none of the input bytes.
#[test]
fn ac4_bad_input_adds_one_receipt_line_without_the_input() {
    let big = {
        let mut v = br#"{"k":"SECRETMARKER"#.to_vec();
        v.extend(std::iter::repeat(b'x').take(STDIN_CAP));
        v.extend_from_slice(br#""}"#);
        v
    };
    let cases: Vec<(&str, Vec<u8>, &str)> = vec![
        ("empty", Vec::new(), "not_json"),
        ("text", b"SECRETMARKER not json".to_vec(), "not_json"),
        ("array", br#"["SECRETMARKER"]"#.to_vec(), "not_object"),
        ("scalar", b"\"SECRETMARKER\"".to_vec(), "not_object"),
        ("not utf-8", vec![b'{', 0xff, 0xfe, b'S', b'E', b'C', b'R', b'E', b'T', b'M', b'A', b'R', b'K', b'E', b'R'], "not_utf8"),
        ("too large", big, "stdin_too_large"),
    ];
    for name in NAMES {
        let fx = Fx::new(true, true);
        for (i, (what, input, reason)) in cases.iter().enumerate() {
            let r = fx.hook(name, input);
            assert_eq!(r.code, 0, "{name} {what}: {}", r.stderr);
            assert_eq!(r.stdout, "", "{name} {what}");
            assert_eq!(r.stderr, "", "{name} {what}: a recorded failure leaves stderr empty");
            let lines = receipt_lines(&fx.receipt());
            assert_eq!(lines.len(), i + 1, "{name} {what}");
            let v: Value = serde_json::from_str(lines.last().unwrap()).unwrap();
            assert_eq!(v["name"], name, "{what}");
            assert_eq!(v["reason"], *reason, "{what}");
            assert!(v["at_unix_ms"].as_u64().is_some_and(|t| t > 0), "{what}: {v}");
        }
        let text = fs::read_to_string(fx.receipt()).unwrap();
        assert!(!text.contains("SECRETMARKER"), "the input must not reach the receipt");
    }
}

/// AC-5: with no registered project folder the failure is one stderr line, nothing is created.
#[test]
fn ac5_without_a_project_folder_the_failure_is_one_stderr_line() {
    let fx = Fx::new(true, false);
    let before = snapshot(&fx.home);
    let r = fx.hook("write-boundary", b"SECRETMARKER not json");
    assert_eq!((r.code, r.stdout.as_str()), (0, ""), "{}", r.stderr);
    assert_eq!(r.stderr.lines().count(), 1, "{:?}", r.stderr);
    assert!(!r.stderr.contains("SECRETMARKER"));
    assert!(
        r.stderr.contains("write-boundary") && r.stderr.contains("not_json"),
        "the line names the hook and the reason: {:?}",
        r.stderr
    );
    assert_eq!(before, snapshot(&fx.home), "no folder and no file may be created in HOME");
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
    for (input, label) in [(&b"not json"[..], "bad"), (&valid_input("SessionStart")[..], "good")] {
        let a = plain.run(&["hook", "bash-boundary"], &plain.project, input, &[]);
        let b = with.run(&["hook", "bash-boundary"], &with.project, input, &orca);
        assert_eq!((a.code, &a.stdout, &a.stderr), (b.code, &b.stdout, &b.stderr), "{label}");
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

    let source = fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join("src/hook_entry.rs"))
        .expect("src/hook_entry.rs");
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
        let mut rest = source.as_str();
        while let Some(at) = rest.find(marker) {
            rest = &rest[at + marker.len()..];
            variables.push(rest[..rest.find('"').unwrap()].to_owned());
        }
    }
    variables.sort();
    variables.dedup();
    assert_eq!(variables, ["CLAUDE_PROJECT_DIR", "HOME"]);
}
