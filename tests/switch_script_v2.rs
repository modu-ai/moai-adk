//! SPEC-20261119-switch-script-v2 (C-33 S2a): the preview of the name-switch script v2.
//!
//! The installed `moai` script sends 63 words to the Go build. v2 sends `statusline`, `session`
//! and the Rust actions of `spec` to the Rust build and leaves the rest as v1 has it. These tests
//! run the two preview files (`moai.v1` is a byte copy of the installed script, `moai.v2` is the
//! preview) with `sh`, in a temporary HOME, against two fake builds that only print who was
//! called. Nothing here touches the real HOME and no real Go or Rust build is started.
#![cfg(unix)]

use std::collections::BTreeSet;
use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use tempfile::TempDir;

const PREVIEW: &str = concat!(
    env!("CARGO_MANIFEST_DIR"),
    "/.moai/specs/SPEC-20261119-switch-script-v2/preview"
);
const SPEC: &str = concat!(
    env!("CARGO_MANIFEST_DIR"),
    "/.moai/specs/SPEC-20261119-switch-script-v2/spec.md"
);

/// What `moai spec --help` lists on the Rust build (`new check list show take close help`).
const RUST_SPEC_ACTIONS: [&str; 6] = ["new", "check", "list", "show", "take", "close"];
/// The actions only the Go build has (read from `moai.old spec --help` in a temporary HOME).
const GO_ONLY_SPEC_ACTIONS: [&str; 7] = [
    "status", "drift", "view", "lint", "audit", "archive", "ceiling",
];
const LAUNCHER_WORDS: [&str; 8] = ["cc", "glm", "codex", "omp", "-l", "-a", "--leader", "--agent"];
/// The three words v2 moves from the Go build to the Rust build (`spec` by action).
const MOVED: [&str; 3] = ["spec", "statusline", "session"];

fn script(name: &str) -> PathBuf {
    Path::new(PREVIEW).join(name)
}

fn text(path: &Path) -> String {
    fs::read_to_string(path).unwrap_or_else(|e| panic!("{} is missing: {e}", path.display()))
}

/// Two fake builds in a temporary folder: each prints `<WHO> oneshot=<flag> <args…>` and exits
/// with `FAKE_EXIT` (default 0).
struct Fakes {
    dir: TempDir,
}

impl Fakes {
    fn new() -> Self {
        let dir = TempDir::new().unwrap();
        for who in ["RUST", "GO"] {
            let path = dir.path().join(who.to_lowercase());
            let body = format!(
                "#!/bin/sh\nprintf '%s oneshot=%s' {who} \"${{MOAI_MCP_ONESHOT-unset}}\"\n\
                 for a in \"$@\"; do printf ' %s' \"$a\"; done\nprintf '\\n'\nexit \"${{FAKE_EXIT:-0}}\"\n"
            );
            fs::write(&path, body).unwrap();
            fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
        }
        Fakes { dir }
    }
    fn rust(&self) -> PathBuf {
        self.dir.path().join("rust")
    }
    fn go(&self) -> PathBuf {
        self.dir.path().join("go")
    }
}

struct Out {
    code: i32,
    stdout: String,
    stderr: String,
}

fn run_with(script_path: &Path, fakes: &Fakes, go: &Path, args: &[&str], env: &[(&str, &str)]) -> Out {
    let mut cmd = Command::new("sh");
    cmd.arg(script_path)
        .args(args)
        .env_clear()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", fakes.dir.path())
        .env("MOAI_RUST_BIN", fakes.rust())
        .env("MOAI_GO_BIN", go)
        .stdin(Stdio::null());
    for (k, v) in env {
        cmd.env(k, v);
    }
    let out = cmd.output().expect("sh runs the script");
    Out {
        code: out.status.code().unwrap_or(-1),
        stdout: String::from_utf8_lossy(&out.stdout).trim_end().to_string(),
        stderr: String::from_utf8_lossy(&out.stderr).trim_end().to_string(),
    }
}

fn run(script_path: &Path, fakes: &Fakes, args: &[&str]) -> Out {
    run_with(script_path, fakes, &fakes.go(), args, &[])
}

/// The words of the script's `case` line that ends in `mcp-server`'s alternation.
fn go_words(script_text: &str) -> BTreeSet<String> {
    let line = script_text
        .lines()
        .map(str::trim)
        .find(|l| l.contains("|mcp-server|") && l.ends_with(')'))
        .unwrap_or_else(|| panic!("no Go word line found"));
    line.trim_end_matches(')').split('|').map(str::to_string).collect()
}

fn called(who: &str, oneshot: &str, args: &[&str]) -> String {
    let mut s = format!("{who} oneshot={oneshot}");
    for a in args {
        s.push(' ');
        s.push_str(a);
    }
    s
}

#[test]
fn v2_sends_statusline_and_session_to_rust() {
    let f = Fakes::new();
    let v2 = script("moai.v2");
    for args in [
        vec!["statusline"],
        vec!["statusline", "--refresh-counts"],
        vec!["session"],
        vec!["session", "-x", "y z"],
    ] {
        let o = run(&v2, &f, &args);
        assert_eq!(o.code, 0, "{args:?}: {}", o.stderr);
        assert_eq!(o.stdout, called("RUST", "unset", &args), "{args:?}");
    }
}

#[test]
fn v2_splits_spec_by_action() {
    let f = Fakes::new();
    let v2 = script("moai.v2");
    let mut rust_calls: Vec<Vec<&str>> = RUST_SPEC_ACTIONS
        .iter()
        .map(|a| vec!["spec", *a, "SPEC-20261119-x", "--project", "."])
        .collect();
    rust_calls.extend([vec!["spec"], vec!["spec", "help"], vec!["spec", "-h"], vec!["spec", "--help"]]);
    for args in rust_calls {
        let o = run(&v2, &f, &args);
        assert_eq!(o.stdout, called("RUST", "unset", &args), "{args:?}");
    }
    for action in GO_ONLY_SPEC_ACTIONS {
        let args = ["spec", action, "SPEC-AUTH-001", "--flag"];
        let o = run(&v2, &f, &args);
        assert_eq!(o.stdout, called("GO", "unset", &args), "spec {action}");
    }
}

#[test]
fn v1_still_sends_the_three_words_to_go() {
    // The baseline that shows the leak (F1): the installed script sends all three to Go.
    let f = Fakes::new();
    let v1 = script("moai.v1");
    for args in [
        vec!["statusline"],
        vec!["session"],
        vec!["spec", "check", "--all"],
        vec!["spec", "lint", "SPEC-AUTH-001"],
    ] {
        let o = run(&v1, &f, &args);
        assert_eq!(o.stdout, called("GO", "unset", &args), "{args:?}");
    }
}

#[test]
fn v2_changes_exactly_three_words_of_v1() {
    let v1 = go_words(&text(&script("moai.v1")));
    let v2 = go_words(&text(&script("moai.v2")));
    assert_eq!(v1.len(), 63, "v1 sends 63 words to Go (S1 survey)");
    let removed: BTreeSet<String> = v1.difference(&v2).cloned().collect();
    let added: BTreeSet<String> = v2.difference(&v1).cloned().collect();
    let expected: BTreeSet<String> = MOVED.iter().map(|s| s.to_string()).collect();
    assert_eq!(removed, expected, "v2 drops exactly spec, statusline, session from the list");
    assert!(added.is_empty(), "v2 adds no word to the Go list: {added:?}");
}

#[test]
fn v2_keeps_launcher_and_unknown_words_on_rust() {
    let f = Fakes::new();
    let (v1, v2) = (script("moai.v1"), script("moai.v2"));
    let mut cases: Vec<Vec<&str>> = LAUNCHER_WORDS.iter().map(|w| vec![*w, "x"]).collect();
    cases.extend([
        vec!["card", "list"],
        vec!["lane-env", "--lane", "agent-1"],
        vec!["claim-verify"],
        vec!["hook-dispatch"],
        vec!["zz-no-such-word"],
        vec![],
    ]);
    for args in cases {
        let (a, b) = (run(&v1, &f, &args), run(&v2, &f, &args));
        assert_eq!(b.stdout, called("RUST", "unset", &args), "{args:?}");
        assert_eq!((a.code, a.stdout), (b.code, b.stdout), "v1 and v2 agree on {args:?}");
    }
}

#[test]
fn v2_exports_the_mcp_oneshot_flag_for_mcp_server_only() {
    let f = Fakes::new();
    let v2 = script("moai.v2");
    let o = run(&v2, &f, &["mcp-server", "--stdio"]);
    assert_eq!(o.stdout, called("GO", "1", &["mcp-server", "--stdio"]));
    for word in go_words(&text(&v2)) {
        if word == "mcp-server" {
            continue;
        }
        let o = run(&v2, &f, &[word.as_str(), "x"]);
        assert_eq!(o.stdout, called("GO", "unset", &[word.as_str(), "x"]), "{word}");
    }
    let o = run(&v2, &f, &["spec", "lint"]);
    assert_eq!(o.stdout, called("GO", "unset", &["spec", "lint"]));
}

#[test]
fn v2_without_a_go_build_fails_only_for_go_words() {
    let f = Fakes::new();
    let v2 = script("moai.v2");
    let missing = f.dir.path().join("no-go-here");
    for args in [vec!["hook", "pre-tool"], vec!["spec", "lint"], vec!["status"]] {
        let o = run_with(&v2, &f, &missing, &args, &[]);
        assert_eq!(o.code, 1, "{args:?}");
        assert!(o.stdout.is_empty(), "{args:?}: {}", o.stdout);
        assert!(o.stderr.contains(&format!("'{}'", args[0])), "names the word: {}", o.stderr);
        assert!(o.stderr.contains(missing.to_str().unwrap()), "names the path: {}", o.stderr);
    }
    for args in [
        vec!["statusline"],
        vec!["session"],
        vec!["spec", "check"],
        vec!["card", "list"],
        vec![],
    ] {
        let o = run_with(&v2, &f, &missing, &args, &[]);
        assert_eq!(o.code, 0, "{args:?}: {}", o.stderr);
        assert_eq!(o.stdout, called("RUST", "unset", &args), "{args:?}");
    }
}

#[test]
fn v2_hands_over_arguments_and_exit_status() {
    let f = Fakes::new();
    let v2 = script("moai.v2");
    let args = ["spec", "check", "a b", "*", "--x=y", "it's"];
    for (args, who) in [
        (&args[..], "RUST"),
        (&["spec", "lint", "a b", "*", "--x=y", "it's"][..], "GO"),
    ] {
        let o = run_with(&v2, &f, &f.go(), args, &[("FAKE_EXIT", "7")]);
        assert_eq!(o.code, 7, "{who} exit status passes through");
        assert_eq!(o.stdout, called(who, "unset", args), "{who} arguments unchanged");
    }
}

/// The line of the Result section that starts with `prefix` and holds a backtick-quoted command.
fn result_line(spec: &str, prefix: &str) -> String {
    let line = spec
        .lines()
        .find(|l| l.starts_with(prefix))
        .unwrap_or_else(|| panic!("the Result has no line starting {prefix:?}"));
    let open = line.find('`').expect("opening backtick");
    let close = line.rfind('`').expect("closing backtick");
    assert!(close > open, "one backtick-quoted command on the line");
    line[open + 1..close].to_string()
}

#[test]
fn the_spec_names_every_word_it_changes_and_the_undo_line() {
    let spec = text(Path::new(SPEC));
    let result = spec.split("\n## Result").nth(1).expect("a Result section");
    for word in MOVED.iter().chain(GO_ONLY_SPEC_ACTIONS.iter()) {
        assert!(result.contains(&format!("`{word}`")), "the Result names `{word}`");
    }
    assert!(
        result.contains("moai.before-"),
        "the Result names the saved copy the undo line restores"
    );
    let undo = result_line(&spec, "- undo line: ");
    assert!(undo.contains("$STAMP") && undo.contains("mv "), "undo line: {undo}");
    let apply = result_line(&spec, "- apply lines (S4 only, after the user picks): ");
    assert!(apply.contains("$PREVIEW") && apply.contains("$STAMP"), "apply lines: {apply}");
}

#[test]
fn the_undo_line_restores_v1_byte_for_byte() {
    let spec = text(Path::new(SPEC));
    let undo = result_line(&spec, "- undo line: ");
    let apply = result_line(&spec, "- apply lines (S4 only, after the user picks): ");
    let home = TempDir::new().unwrap();
    let (v1, v2) = (fs::read(script("moai.v1")).unwrap(), fs::read(script("moai.v2")).unwrap());
    let dirs = [home.path().join(".moai-cli/path"), home.path().join("go/bin")];
    for d in &dirs {
        fs::create_dir_all(d).unwrap();
        fs::write(d.join("moai"), &v1).unwrap();
        fs::set_permissions(d.join("moai"), fs::Permissions::from_mode(0o755)).unwrap();
    }
    let sh = |line: &str| {
        let out = Command::new("sh")
            .arg("-c")
            .arg(line)
            .env_clear()
            .env("PATH", "/usr/bin:/bin")
            .env("HOME", home.path())
            .env("PREVIEW", PREVIEW)
            .env("STAMP", "20261119T000000Z")
            .output()
            .unwrap();
        assert!(out.status.success(), "{line}: {}", String::from_utf8_lossy(&out.stderr));
    };
    sh(&apply);
    for d in &dirs {
        assert_eq!(fs::read(d.join("moai")).unwrap(), v2, "apply put v2 in {}", d.display());
        assert_eq!(
            fs::metadata(d.join("moai")).unwrap().permissions().mode() & 0o111,
            0o111,
            "v2 is executable"
        );
        assert_eq!(
            fs::read(d.join("moai.before-20261119T000000Z")).unwrap(),
            v1,
            "apply saved v1 beside it"
        );
    }
    sh(&undo);
    for d in &dirs {
        assert_eq!(fs::read(d.join("moai")).unwrap(), v1, "undo restored v1 in {}", d.display());
        assert!(!d.join("moai.before-20261119T000000Z").exists(), "the saved copy was moved back");
    }
}
