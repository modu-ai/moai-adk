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
/// The actions only the Go build has for `session` (`moai.old session --help`; the Rust `session`
/// is the legacy launcher: `moai session --name <NAME>`, no actions).
const GO_ONLY_SESSION_ACTIONS: [&str; 7] = [
    "register",
    "heartbeat",
    "deregister",
    "list",
    "purge",
    "current",
    "doctor",
];
const LAUNCHER_WORDS: [&str; 8] = [
    "cc", "glm", "codex", "omp", "-l", "-a", "--leader", "--agent",
];
/// The three words v2 takes off the Go list (`spec` and `session` stay Go by action).
const MOVED: [&str; 3] = ["spec", "statusline", "session"];

fn script(name: &str) -> PathBuf {
    Path::new(PREVIEW).join(name)
}

fn text(path: &Path) -> String {
    fs::read_to_string(path).unwrap_or_else(|e| panic!("{} is missing: {e}", path.display()))
}

/// Two fake builds in a temporary folder: each prints `<WHO> oneshot=<flag>` and then every
/// argument in its own brackets, `[a b] [c]`, so one argument with a space differs from two
/// arguments, and exits with `FAKE_EXIT` (default 0).
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
                 for a in \"$@\"; do printf ' [%s]' \"$a\"; done\nprintf '\\n'\nexit \"${{FAKE_EXIT:-0}}\"\n"
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

fn run_with(
    script_path: &Path,
    fakes: &Fakes,
    go: &Path,
    args: &[&str],
    env: &[(&str, &str)],
) -> Out {
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
    line.trim_end_matches(')')
        .split('|')
        .map(str::to_string)
        .collect()
}

fn called(who: &str, oneshot: &str, args: &[&str]) -> String {
    let mut s = format!("{who} oneshot={oneshot}");
    for a in args {
        s.push_str(&format!(" [{a}]"));
    }
    s
}

#[test]
fn v2_sends_statusline_and_the_session_launcher_form_to_rust() {
    let f = Fakes::new();
    let v2 = script("moai.v2");
    for args in [
        vec!["statusline"],
        vec!["statusline", "--refresh-counts"],
        vec!["statusline", "a b", "*"],
        vec!["session"],
        vec!["session", "--name", "y z", "--role", "leader"],
        vec!["session", "-x", "y z"],
    ] {
        let o = run(&v2, &f, &args);
        assert_eq!(o.code, 0, "{args:?}: {}", o.stderr);
        assert_eq!(o.stdout, called("RUST", "unset", &args), "{args:?}");
    }
}

#[test]
fn v2_keeps_the_go_session_actions_on_go() {
    // Go-generation templates tell agents to run `moai session list --json` and
    // `moai session current`; the Rust `session` has no such actions (exit 2).
    let f = Fakes::new();
    let v2 = script("moai.v2");
    for action in GO_ONLY_SESSION_ACTIONS {
        let args = ["session", action, "--json", "a b"];
        let o = run(&v2, &f, &args);
        assert_eq!(o.stdout, called("GO", "unset", &args), "session {action}");
    }
}

#[test]
fn v2_routes_by_the_first_two_words_only() {
    // Documented limits (spec Risks): a flag before the action, a wrong case, an empty action and
    // `help <action>` all read as "not a Go-only action" and go to Rust.
    let f = Fakes::new();
    let v2 = script("moai.v2");
    for args in [
        vec!["spec", "--project", ".", "lint"],
        vec!["spec", "help", "lint"],
        vec!["spec", "Lint"],
        vec!["spec", ""],
        vec!["session", "--name", "x", "list"],
    ] {
        let o = run(&v2, &f, &args);
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
    rust_calls.extend([
        vec!["spec"],
        vec!["spec", "help"],
        vec!["spec", "-h"],
        vec!["spec", "--help"],
    ]);
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
    assert_eq!(
        removed, expected,
        "v2 drops exactly spec, statusline, session from the list"
    );
    assert!(
        added.is_empty(),
        "v2 adds no word to the Go list: {added:?}"
    );
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
        assert_eq!(
            (a.code, a.stdout),
            (b.code, b.stdout),
            "v1 and v2 agree on {args:?}"
        );
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
        assert_eq!(
            o.stdout,
            called("GO", "unset", &[word.as_str(), "x"]),
            "{word}"
        );
    }
    let o = run(&v2, &f, &["spec", "lint"]);
    assert_eq!(o.stdout, called("GO", "unset", &["spec", "lint"]));
}

#[test]
fn v2_without_a_go_build_fails_only_for_go_words() {
    let f = Fakes::new();
    let v2 = script("moai.v2");
    let missing = f.dir.path().join("no-go-here");
    for args in [
        vec!["hook", "pre-tool"],
        vec!["spec", "lint"],
        vec!["session", "list"],
        vec!["status"],
    ] {
        let o = run_with(&v2, &f, &missing, &args, &[]);
        assert_eq!(o.code, 1, "{args:?}");
        assert!(o.stdout.is_empty(), "{args:?}: {}", o.stdout);
        assert!(
            o.stderr.contains(&format!("'{}'", args[0])),
            "names the word: {}",
            o.stderr
        );
        assert!(
            o.stderr.contains(missing.to_str().unwrap()),
            "names the path: {}",
            o.stderr
        );
    }
    for args in [
        vec!["statusline"],
        vec!["session"],
        vec!["session", "--name", "x"],
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
    // Every route: the spec branch, the final Rust exec (statusline, an unknown word), the
    // launcher branch and the Go list. One argument with a space stays one `[…]`.
    for (args, who) in [
        (&["spec", "check", "a b", "*", "--x=y", "it's"][..], "RUST"),
        (&["spec", "lint", "a b", "*", "--x=y", "it's"][..], "GO"),
        (&["statusline", "a b", "*", "--x=y", "it's"][..], "RUST"),
        (&["session", "list", "a b", "*", "it's"][..], "GO"),
        (&["card", "list", "a b", "*", "it's"][..], "RUST"),
        (&["cc", "a b", "*", "it's"][..], "RUST"),
        (&["hook", "a b", "*", "it's"][..], "GO"),
    ] {
        let o = run_with(&v2, &f, &f.go(), args, &[("FAKE_EXIT", "7")]);
        assert_eq!(o.code, 7, "{who} exit status passes through");
        assert_eq!(
            o.stdout,
            called(who, "unset", args),
            "{who} arguments unchanged"
        );
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
    for word in MOVED
        .iter()
        .chain(GO_ONLY_SPEC_ACTIONS.iter())
        .chain(GO_ONLY_SESSION_ACTIONS.iter())
    {
        assert!(
            result.contains(&format!("`{word}`")),
            "the Result names `{word}`"
        );
    }
    assert!(
        result.contains("moai.before-"),
        "the Result names the saved copy the undo line restores"
    );
    let undo = result_line(&spec, "- undo line: ");
    assert!(
        undo.contains("$STAMP") && undo.contains("mv "),
        "undo line: {undo}"
    );
    let apply = result_line(&spec, "- apply lines (S4 only, after the user picks): ");
    assert!(
        apply.contains("$PREVIEW") && apply.contains("$STAMP"),
        "apply lines: {apply}"
    );
}

const STAMP: &str = "20261119T000000Z";

/// A temporary HOME holding v1 as `.moai-cli/path/moai` and `go/bin/moai` (both regular files,
/// mode 0755, like the real ones), and the apply and undo lines read from the SPEC Result.
struct Home {
    dir: TempDir,
    apply: String,
    undo: String,
    v1: Vec<u8>,
    v2: Vec<u8>,
}

impl Home {
    fn new() -> Self {
        let spec = text(Path::new(SPEC));
        let home = Home {
            dir: TempDir::new().unwrap(),
            undo: result_line(&spec, "- undo line: "),
            apply: result_line(&spec, "- apply lines (S4 only, after the user picks): "),
            v1: fs::read(script("moai.v1")).unwrap(),
            v2: fs::read(script("moai.v2")).unwrap(),
        };
        for d in home.dirs() {
            fs::create_dir_all(&d).unwrap();
            fs::write(d.join("moai"), &home.v1).unwrap();
            fs::set_permissions(d.join("moai"), fs::Permissions::from_mode(0o755)).unwrap();
        }
        home
    }
    fn dirs(&self) -> [PathBuf; 2] {
        [
            self.dir.path().join(".moai-cli/path"),
            self.dir.path().join("go/bin"),
        ]
    }
    /// Runs one line with `sh -c`; `stamp` None leaves STAMP unset. Returns (status ok, stderr).
    fn sh(&self, line: &str, stamp: Option<&str>) -> (bool, String) {
        let mut cmd = Command::new("sh");
        cmd.arg("-c")
            .arg(line)
            .env_clear()
            .env("PATH", "/usr/bin:/bin")
            .env("HOME", self.dir.path())
            .env("PREVIEW", PREVIEW);
        if let Some(s) = stamp {
            cmd.env("STAMP", s);
        }
        let out = cmd.output().unwrap();
        (
            out.status.success(),
            String::from_utf8_lossy(&out.stderr).to_string(),
        )
    }
    fn moai(&self, i: usize) -> Vec<u8> {
        fs::read(self.dirs()[i].join("moai")).unwrap()
    }
    fn saved(&self, i: usize, stamp: &str) -> PathBuf {
        self.dirs()[i].join(format!("moai.before-{stamp}"))
    }
}

#[test]
fn the_undo_line_restores_v1_byte_for_byte() {
    let h = Home::new();
    assert!(h.sh(&h.apply, Some(STAMP)).0, "apply succeeds");
    for i in 0..2 {
        assert_eq!(h.moai(i), h.v2, "apply put v2 in folder {i}");
        let mode = fs::metadata(h.dirs()[i].join("moai"))
            .unwrap()
            .permissions()
            .mode();
        assert_eq!(mode & 0o111, 0o111, "v2 is executable");
        assert_eq!(
            fs::read(h.saved(i, STAMP)).unwrap(),
            h.v1,
            "apply saved v1 beside it"
        );
    }
    assert!(h.sh(&h.undo, Some(STAMP)).0, "undo succeeds");
    for i in 0..2 {
        assert_eq!(h.moai(i), h.v1, "undo restored v1 in folder {i}");
        assert!(!h.saved(i, STAMP).exists(), "the saved copy was moved back");
    }
}

#[test]
fn a_second_apply_with_the_same_stamp_refuses_and_keeps_the_saved_v1() {
    // Audit M3: the old line copied the already-installed v2 over the saved v1.
    let h = Home::new();
    assert!(h.sh(&h.apply, Some(STAMP)).0);
    let (ok, err) = h.sh(&h.apply, Some(STAMP));
    assert!(!ok, "the second apply must fail");
    assert!(err.contains("apply stopped"), "it says so: {err}");
    for i in 0..2 {
        assert_eq!(fs::read(h.saved(i, STAMP)).unwrap(), h.v1, "v1 copy intact");
    }
    assert!(h.sh(&h.undo, Some(STAMP)).0);
    for i in 0..2 {
        assert_eq!(h.moai(i), h.v1, "undo still gives v1 in folder {i}");
    }
}

#[test]
fn apply_and_undo_report_failure_through_the_exit_status() {
    // Audit M4: a failed apply or undo must not look like success.
    let h = Home::new();
    // An empty or missing STAMP changes nothing and fails.
    for stamp in [None, Some("")] {
        let (ok, err) = h.sh(&h.apply, stamp);
        assert!(!ok, "apply with STAMP {stamp:?} fails");
        assert!(err.contains("STAMP is empty"), "it says why: {err}");
        for i in 0..2 {
            assert_eq!(h.moai(i), h.v1, "nothing changed in folder {i}");
        }
        let (ok, err) = h.sh(&h.undo, stamp);
        assert!(!ok, "undo with STAMP {stamp:?} fails");
        assert!(err.contains("STAMP is empty"), "it says why: {err}");
    }
    // The second folder has no script: apply stops there with a failing status; the first folder
    // is already v2 and the undo puts it back, then reports the folder that had nothing saved.
    fs::remove_file(h.dirs()[1].join("moai")).unwrap();
    let (ok, err) = h.sh(&h.apply, Some(STAMP));
    assert!(!ok, "apply fails when the second script is missing");
    assert!(err.contains("apply stopped"), "it says so: {err}");
    assert_eq!(h.moai(0), h.v2, "the first folder was applied");
    let (ok, err) = h.sh(&h.undo, Some(STAMP));
    assert!(!ok, "undo fails for the folder with nothing saved");
    assert!(err.contains("undo stopped"), "it says so: {err}");
    assert_eq!(h.moai(0), h.v1, "the first folder was still restored");
    // A repeated undo has nothing left to move and fails.
    assert!(!h.sh(&h.undo, Some(STAMP)).0, "a second undo fails");
}

#[test]
fn apply_refuses_a_folder_that_already_holds_v2() {
    // Audit M6: after a partial apply, a retry with a NEW stamp that skips the undo must not save
    // the installed v2 as "the old script".
    let h = Home::new();
    let second = h.dirs()[1].join("moai");
    fs::remove_file(&second).unwrap();
    assert!(
        !h.sh(&h.apply, Some(STAMP)).0,
        "the first apply stops at the second folder"
    );
    assert_eq!(h.moai(0), h.v2, "the first folder is v2 now");
    fs::write(&second, &h.v1).unwrap();
    fs::set_permissions(&second, fs::Permissions::from_mode(0o755)).unwrap();
    let (ok, err) = h.sh(&h.apply, Some("20261119T010000Z"));
    assert!(!ok, "the retry with a new stamp must fail");
    assert!(err.contains("already holds v2"), "it says why: {err}");
    assert!(
        !h.saved(0, "20261119T010000Z").exists(),
        "v2 was not saved as the old script"
    );
    assert!(
        !h.saved(1, "20261119T010000Z").exists(),
        "nothing happened in the second folder"
    );
    assert_eq!(h.moai(1), h.v1, "the second folder is untouched");
    // The documented order works: undo with the first stamp, then apply again.
    fs::write(h.saved(1, STAMP), &h.v1).unwrap();
    assert!(
        h.sh(&h.undo, Some(STAMP)).0,
        "the undo with the first stamp"
    );
    assert!(
        h.sh(&h.apply, Some("20261119T020000Z")).0,
        "a fresh apply after the undo"
    );
    for i in 0..2 {
        assert_eq!(h.moai(i), h.v2, "folder {i} is v2");
        assert_eq!(
            fs::read(h.saved(i, "20261119T020000Z")).unwrap(),
            h.v1,
            "the saved copy is v1"
        );
    }
}

#[test]
fn apply_never_overwrites_an_existing_saved_copy() {
    // The saved copy may hold something the apply did not write (a leftover, a hand-made copy).
    let h = Home::new();
    fs::write(h.saved(0, STAMP), b"sentinel").unwrap();
    let (ok, err) = h.sh(&h.apply, Some(STAMP));
    assert!(!ok, "apply must fail");
    assert!(err.contains("already exists"), "it says why: {err}");
    assert_eq!(
        fs::read(h.saved(0, STAMP)).unwrap(),
        b"sentinel",
        "kept as it was"
    );
    for i in 0..2 {
        assert_eq!(h.moai(i), h.v1, "nothing was installed in folder {i}");
    }
}

#[test]
fn undo_tries_the_second_folder_even_when_the_first_fails() {
    // Audit N5: an undo that stopped at the first failure would leave the second folder on v2.
    let h = Home::new();
    assert!(h.sh(&h.apply, Some(STAMP)).0);
    fs::remove_file(h.saved(0, STAMP)).unwrap();
    let (ok, err) = h.sh(&h.undo, Some(STAMP));
    assert!(!ok, "the undo reports the folder with nothing saved");
    assert!(err.contains("undo stopped"), "it says so: {err}");
    assert_eq!(
        h.moai(0),
        h.v2,
        "the first folder had no saved copy and stays v2"
    );
    assert_eq!(h.moai(1), h.v1, "the second folder was still restored");
}
