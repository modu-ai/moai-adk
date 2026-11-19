//! SPEC-20261119-switch-script-v3 (card C-136): the name-switch script v3 and the table of every
//! word it routes.
//!
//! v2 still listed `bundle` among the Go-only words, so a plain `moai bundle build` went to the Go
//! build (whose `bundle` only has `add`, `remove` and `list`). v3 splits `bundle` by action like
//! `spec` and `session`. These tests run `install/moai-switch.sh` (the script the program installs)
//! and the v2 preview with `sh`, in a temporary HOME, against two fake builds that only print who
//! was called. The table test reads the Rust command list from the built `moai help --all`, so a
//! word that Rust gains later and that the script still sends to Go fails here. Nothing touches the
//! real HOME and no real Go build is started.
#![cfg(unix)]

mod common;
use common::CleanEnv;

use std::collections::BTreeSet;
use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use tempfile::TempDir;

const V3: &str = concat!(env!("CARGO_MANIFEST_DIR"), "/install/moai-switch.sh");
const PREVIEW: &str = concat!(
    env!("CARGO_MANIFEST_DIR"),
    "/.moai/specs/SPEC-20261119-switch-script-v2/preview"
);
const EXE: &str = crate::common::MOAI_BIN;

/// The actions the Rust `moai bundle` has (`moai bundle --help`).
const RUST_BUNDLE_ACTIONS: [&str; 5] = ["ready", "check", "body", "build", "verify"];
/// The actions only the Go `bundle` has (`moai.old bundle --help` in an empty temporary HOME:
/// `add <bundle-name>`, `remove <bundle-name>`, `list`).
const GO_ONLY_BUNDLE_ACTIONS: [&str; 3] = ["add", "remove", "list"];

/// How a word of the v1 Go list is routed by v3.
#[derive(Clone, Copy, PartialEq, Eq, Debug)]
enum Route {
    /// Only the Go build has the word.
    Go,
    /// The word moved to the Rust build.
    Rust,
    /// Both builds have the word: Go keeps the actions only it has, Rust gets the rest.
    Split,
}

use Route::{Go, Rust, Split};

/// Every word of the v1 script's Go list (63, S1 survey of card C-33) with its route in v3. The
/// Rust command list is read from the built program in the tests below; this table is what the
/// script has to say about each word.
const TABLE: [(&str, Route); 63] = [
    ("af", Go),
    ("agent", Go),
    ("ast-edit", Go),
    ("ast-grep", Go),
    ("bundle", Split),
    ("chain", Go),
    ("ci-verdict", Go),
    ("clean", Go),
    ("completion", Go),
    ("config", Go),
    ("constitution", Go),
    ("contract", Go),
    ("decision", Go),
    ("doctor", Go),
    ("epic", Go),
    ("factory", Go),
    ("feedback", Go),
    ("fetch", Go),
    ("gate", Go),
    ("github", Go),
    ("goal", Go),
    ("graph", Go),
    ("gtd", Go),
    ("handoff", Go),
    ("harness", Go),
    ("hook", Go),
    ("inbox", Go),
    ("init", Go),
    ("integration", Go),
    ("inventory", Go),
    ("jev", Go),
    ("jevmeasure", Go),
    ("loop", Go),
    ("lsp", Go),
    ("mcp", Go),
    ("mcp-server", Go),
    ("memory", Go),
    ("migrate", Go),
    ("migration", Go),
    ("moai", Go),
    ("mx", Go),
    ("paths", Go),
    ("plan", Go),
    ("pr", Go),
    ("preference", Go),
    ("profile", Go),
    ("research", Go),
    ("session", Split),
    ("slot", Go),
    ("spec", Split),
    ("state", Go),
    ("statusline", Rust),
    ("status", Go),
    ("telemetry", Go),
    ("that", Go),
    ("todo", Go),
    ("tokens", Go),
    ("tool-policy", Go),
    ("verify", Go),
    ("version", Go),
    ("web", Go),
    ("workflow", Go),
    ("worktree", Go),
];

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

fn run_with(script_path: &Path, fakes: &Fakes, go: &Path, args: &[&str]) -> Out {
    let out = Command::new("sh")
        .arg(script_path)
        .args(args)
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", fakes.dir.path())
        .env("MOAI_RUST_BIN", fakes.rust())
        .env("MOAI_GO_BIN", go)
        .stdin(Stdio::null())
        .output()
        .expect("sh runs the script");
    Out {
        code: out.status.code().unwrap_or(-1),
        stdout: String::from_utf8_lossy(&out.stdout).trim_end().to_string(),
        stderr: String::from_utf8_lossy(&out.stderr).trim_end().to_string(),
    }
}

fn run(script_path: &Path, fakes: &Fakes, args: &[&str]) -> Out {
    run_with(script_path, fakes, &fakes.go(), args)
}

fn v3() -> PathBuf {
    PathBuf::from(V3)
}

fn preview(name: &str) -> PathBuf {
    Path::new(PREVIEW).join(name)
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

/// The words of the script's split branch (`spec|session|bundle)`): every word that is routed by
/// its second word.
fn split_words(script_text: &str) -> BTreeSet<String> {
    let line = script_text
        .lines()
        .map(str::trim)
        .find(|l| l.starts_with("spec|") && l.ends_with(')'))
        .unwrap_or_else(|| panic!("no split line found"));
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

/// SPEC-20261119-switch-script-v3 AC-1 (REQ-1).
#[test]
fn v3_sends_the_rust_bundle_actions_and_its_help_to_rust() {
    let f = Fakes::new();
    let mut cases: Vec<Vec<&str>> = RUST_BUNDLE_ACTIONS
        .iter()
        .map(|a| vec!["bundle", *a, "--project", ".", "a b"])
        .collect();
    cases.extend([
        vec!["bundle"],
        vec!["bundle", "help"],
        vec!["bundle", "-h"],
        vec!["bundle", "--help"],
        vec!["bundle", "build", "--help"],
        vec!["bundle", "verify", "--help"],
    ]);
    for args in cases {
        let o = run(&v3(), &f, &args);
        assert_eq!(o.code, 0, "{args:?}: {}", o.stderr);
        assert_eq!(o.stdout, called("RUST", "unset", &args), "{args:?}");
    }
}

/// SPEC-20261119-switch-script-v3 AC-1 (REQ-2).
#[test]
fn v3_keeps_the_go_bundle_actions_on_go() {
    let f = Fakes::new();
    for action in GO_ONLY_BUNDLE_ACTIONS {
        let args = ["bundle", action, "design-pack", "--flag", "a b"];
        let o = run(&v3(), &f, &args);
        assert_eq!(o.stdout, called("GO", "unset", &args), "bundle {action}");
    }
}

/// SPEC-20261119-switch-script-v3 AC-1 (REQ-1, REQ-2): the same limits as `spec` and `session`.
#[test]
fn v3_routes_bundle_by_the_first_two_words_only() {
    let f = Fakes::new();
    for (args, who) in [
        (vec!["bundle", "--project", ".", "add"], "RUST"),
        (vec!["bundle", "Add"], "RUST"),
        (vec!["bundle", ""], "RUST"),
        (vec!["bundle", "help", "add"], "RUST"),
        (vec!["bundle", "build", "add"], "RUST"),
        (vec!["bundle", "add", "build"], "GO"),
        (vec!["bundle", "list", "ready"], "GO"),
    ] {
        let o = run(&v3(), &f, &args);
        assert_eq!(o.stdout, called(who, "unset", &args), "{args:?}");
    }
}

/// SPEC-20261119-switch-script-v3 AC-2 (REQ-3, REQ-4): missing Go build only fails what needs it.
#[test]
fn v3_without_a_go_build_fails_only_for_the_go_bundle_actions() {
    let f = Fakes::new();
    let nowhere = f.dir.path().join("no-such-go");
    for action in GO_ONLY_BUNDLE_ACTIONS {
        let o = run_with(&v3(), &f, &nowhere, &["bundle", action]);
        assert_eq!(o.code, 1, "bundle {action}");
        assert!(
            o.stderr.contains("'bundle'") && o.stderr.contains(&nowhere.display().to_string()),
            "{}",
            o.stderr
        );
    }
    for action in RUST_BUNDLE_ACTIONS {
        let o = run_with(&v3(), &f, &nowhere, &["bundle", action]);
        assert_eq!(o.code, 0, "bundle {action}: {}", o.stderr);
        assert_eq!(o.stdout, called("RUST", "unset", &["bundle", action]));
    }
    // And the exit status and arguments pass through for both builds.
    let o = {
        let mut cmd = Command::new("sh");
        cmd.arg(v3())
            .args(["bundle", "verify", "x"])
            .clean_env()
            .env("PATH", "/usr/bin:/bin")
            .env("HOME", f.dir.path())
            .env("MOAI_RUST_BIN", f.rust())
            .env("MOAI_GO_BIN", f.go())
            .env("FAKE_EXIT", "7")
            .stdin(Stdio::null());
        cmd.output().unwrap()
    };
    assert_eq!(
        o.status.code(),
        Some(7),
        "the called build's exit status is the script's"
    );
}

/// SPEC-20261119-switch-script-v3 AC-3 (REQ-5): v3 changes `bundle` and nothing else of v2.
#[test]
fn v3_changes_exactly_one_word_of_v2() {
    let v2_text = text(&preview("moai.v2"));
    let v3_text = text(&v3());
    let (g2, g3) = (go_words(&v2_text), go_words(&v3_text));
    let removed: BTreeSet<String> = g2.difference(&g3).cloned().collect();
    let added: BTreeSet<String> = g3.difference(&g2).cloned().collect();
    assert_eq!(
        removed,
        BTreeSet::from(["bundle".to_string()]),
        "v3 drops exactly bundle from the Go list"
    );
    assert!(
        added.is_empty(),
        "v3 adds no word to the Go list: {added:?}"
    );
    assert_eq!(
        split_words(&v3_text),
        BTreeSet::from(["spec", "session", "bundle"].map(str::to_string)),
        "v3 splits spec, session and bundle by action"
    );
    // Every other call behaves as v2 does: the other Go words, the launchers, an unknown word, an
    // empty line, `spec`, `session` and `statusline`.
    let f = Fakes::new();
    let (v2, v3) = (preview("moai.v2"), v3());
    let mut cases: Vec<Vec<&str>> = g2
        .iter()
        .filter(|w| *w != "bundle")
        .map(|w| vec![w.as_str(), "x", "a b"])
        .collect();
    cases.extend([
        vec!["cc", "-l"],
        vec!["glm"],
        vec!["-a"],
        vec!["no-such-word", "y"],
        vec![],
        vec!["statusline"],
        vec!["spec", "check", "--all"],
        vec!["spec", "lint", "SPEC-AUTH-001"],
        vec!["session"],
        vec!["session", "list", "--json"],
        vec!["mcp-server"],
    ]);
    for args in cases {
        let (a, b) = (run(&v2, &f, &args), run(&v3, &f, &args));
        assert_eq!(
            (a.code, a.stdout, a.stderr),
            (b.code, b.stdout, b.stderr),
            "{args:?}"
        );
    }
}

/// The Rust command words of the built program (`moai help --all`): the first column of every
/// line, split at commas, for the lines that start with two spaces and a lowercase letter.
fn rust_words() -> BTreeSet<String> {
    let home = TempDir::new().unwrap();
    let out = Command::new(EXE)
        .args(["help", "--all"])
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", home.path())
        .env("LC_ALL", "C")
        .current_dir(home.path())
        .stdin(Stdio::null())
        .output()
        .expect("run moai help --all");
    assert!(out.status.success(), "moai help --all failed");
    let listing = String::from_utf8_lossy(&out.stdout).into_owned();
    let mut words = BTreeSet::new();
    for line in listing.lines() {
        let Some(rest) = line.strip_prefix("  ") else {
            continue;
        };
        if !rest.starts_with(|c: char| c.is_ascii_lowercase()) {
            continue;
        }
        let first_column = rest.split("  ").next().unwrap_or("");
        for w in first_column.split(", ") {
            words.insert(w.trim().to_string());
        }
    }
    words
}

/// SPEC-20261119-switch-script-v3 AC-3: the parser really reads the Rust list.
#[test]
fn the_rust_command_list_is_read_from_the_built_program() {
    let words = rust_words();
    for known in [
        "spec",
        "session",
        "bundle",
        "statusline",
        "card",
        "update",
        "cc",
    ] {
        assert!(words.contains(known), "{known} missing from {words:?}");
    }
    assert!(words.len() > 100, "only {} words read", words.len());
}

/// SPEC-20261119-switch-script-v3 AC-4 (REQ-5, REQ-6): the table lists exactly the v1 Go list and
/// agrees with the v3 script: the Go words are the script's Go line, the split words are its split
/// branch, and the moved word is in neither.
#[test]
fn the_table_lists_every_word_of_the_v1_list_and_agrees_with_the_v3_script() {
    let v1 = go_words(&text(&preview("moai.v1")));
    let table: BTreeSet<String> = TABLE.iter().map(|(w, _)| w.to_string()).collect();
    assert_eq!(TABLE.len(), table.len(), "a word is in the table twice");
    assert_eq!(table, v1, "the table is the v1 list, word for word");
    let script = text(&v3());
    let of = |route: Route| -> BTreeSet<String> {
        TABLE
            .iter()
            .filter(|(_, r)| *r == route)
            .map(|(w, _)| w.to_string())
            .collect()
    };
    assert_eq!(
        of(Go),
        go_words(&script),
        "Go rows are the script's Go line"
    );
    assert_eq!(
        of(Split),
        split_words(&script),
        "Split rows are its split branch"
    );
    for w in of(Rust) {
        assert!(
            !go_words(&script).contains(&w) && !split_words(&script).contains(&w),
            "{w} moved to Rust but the script still names it"
        );
    }
    assert_eq!(
        (of(Go).len(), of(Split).len(), of(Rust).len()),
        (59, 3, 1),
        "59 stay Go, 3 are split, 1 moved to Rust"
    );
}

/// SPEC-20261119-switch-script-v3 AC-4 (REQ-6): every word the table sends to Go is checked
/// against the Rust command list. It fails when a word Rust now owns is still routed to Go
/// whole: the word has to move to Rust or be split by action.
#[test]
fn no_word_sent_to_go_whole_is_a_rust_command() {
    let rust = rust_words();
    let owned: Vec<&str> = TABLE
        .iter()
        .filter(|(w, r)| *r == Go && rust.contains(*w))
        .map(|(w, _)| *w)
        .collect();
    assert!(
        owned.is_empty(),
        "Rust now has {owned:?}, which the script still sends to Go whole: move it or split it by action"
    );
    // A split or moved word is one Rust has.
    for (w, r) in TABLE {
        if r != Go {
            assert!(
                rust.contains(w),
                "{w} is {r:?} but Rust has no such command"
            );
        }
    }
}
