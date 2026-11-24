//! SPEC-20261124-repo-scans-precheck: `check affected` checks a top-level `tests/*.rs` file that a
//! range adds against the repository scan list `tools/repo_scans.json` (and the exemption lists of
//! `tests/repo_scans_table.rs`) before any heavy step, and stops there with the line to add.
//! Temporary git repositories and fake `cargo`, `python3` and `rustfmt` scripts only: the real
//! tools, the real HOME and the machine's heavy slots are never used.
//!
//! The words that mark a walking or reading test file are built at run time: this file spells none
//! of them, so the repository scan table asks nothing of it.
mod common;
use common::git::{git, git_ok};
use serde_json::Value;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

const BIN: &str = crate::common::MOAI_BIN;

/// The words of a test file that names the repository root and walks or reads it.
fn root_word() -> String {
    format!("CARGO_MANIFEST{}", "_DIR")
}

fn walk_call() -> String {
    format!("read{}", "_dir(")
}

/// `include_str!("<literal>")` or `include_bytes!(...)`, the quotes added here.
fn include(macro_name: &str, literal: &str) -> String {
    format!("{macro_name}!({q}{literal}{q})", q = '"')
}

/// A test file that walks the tree.
fn walker() -> String {
    format!(
        "fn t() {{ let _ = env!(\"{}\"); let _ = std::fs::{}\".\"); }}\n",
        root_word(),
        walk_call()
    )
}

/// A test file that opens the checked-in document `docs/guide.md` at run time.
fn reader() -> String {
    format!(
        "fn t() {{ let _ = std::fs::read_to_string(concat!(env!(\"{}\"), \"/docs/guide.md\")); }}\n",
        root_word()
    )
}

const AREAS: &str = "\
# fixture table
[areas]
card = [\"card\"]
spec = [\"spec\", \"check\"]

[solo]
files = [\"check_solo_cli\"]
";

const ALLOW: &str = "# fixture allow list\n";

/// The scan list: one walking test registered by name, one reader covered by an `always` scan with
/// the check that runs it, and one reader named by an `always` scan that never runs it.
const SCANS: &str = r#"{"about":"x","scans":[
  {"id":"walk-listed","tests":["tests/check_listed_walk_cli.rs::t"],
   "checks":["cargo test --offline --locked --test check_listed_walk_cli"],
   "rules":["tests/check_listed_walk_cli.rs"],"scope":[{"prefix":"docs/"}]},
  {"id":"reads-docs","always":true,"tests":["tests/check_reader_ok_cli.rs::t"],
   "checks":["cargo test --offline --locked --test check_reader_ok_cli"],
   "rules":["tests/check_reader_ok_cli.rs"],"scope":[{"prefix":"docs/"}]},
  {"id":"reads-nocheck","always":true,"tests":["tests/check_nocheck_cli.rs::t"],
   "checks":["cargo test --offline --locked --test other"],
   "rules":["tests/check_nocheck_cli.rs"],"scope":[{"prefix":"docs/"}]}
]}"#;

/// The three exemption lists as `tests/repo_scans_table.rs` spells them.
const EXEMPT_FILE: &str = "\
const EXEMPT: &[(&str, &str)] = &[
    (\"tests/check_exempt_walk_cli.rs\", \"walks a TempDir (a \\\"quoted\\\" reason)\"),
];
const INCLUDE_STR_EXEMPT: &[(&str, &str)] = &[
    (\"tests/check_exempt_inc_cli.rs\", \"reads a fixture\"),
];
const FILE_READ_EXEMPT: &[(&str, &[&str], &str)] = &[
    (
        \"tests/check_exempt_read_cli.rs\",
        &[\"docs/guide.md\"],
        \"writes it into the temporary project\",
    ),
];
";

/// What the base repository holds besides its two older test files.
#[derive(Clone, Copy)]
struct Tables {
    /// `tests/areas.toml` and the state-folder allow list.
    area_tables: bool,
    /// `tools/repo_scans.json`.
    scan_list: bool,
    /// `tests/repo_scans_table.rs`, the file that holds the exemption lists.
    exempt_file: bool,
}

const ALL: Tables = Tables {
    area_tables: true,
    scan_list: true,
    exempt_file: true,
};

fn put(root: &Path, rel: &str, text: &str) {
    put_bytes(root, rel, text.as_bytes());
}

fn put_bytes(root: &Path, rel: &str, bytes: &[u8]) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, bytes).unwrap();
}

struct Fixture {
    dir: tempfile::TempDir,
    base: String,
    head: String,
}

impl Fixture {
    /// A repository whose base holds the tables `tables` names and the tracked document
    /// `docs/guide.md`, and one commit on top that adds or changes `changed`.
    fn new(tables: Tables, changed: &[(String, String)]) -> Fixture {
        Fixture::with_bytes(tables, &changed_bytes(changed), None)
    }

    /// Like `new` with byte contents, and the scan list of the commit's tree replaced by
    /// `scan_list` bytes when given (a list that cannot be read).
    fn with_bytes(
        tables: Tables,
        changed: &[(String, Vec<u8>)],
        scan_list: Option<&[u8]>,
    ) -> Fixture {
        let dir = tempfile::tempdir().unwrap();
        let p = dir.path();
        git_ok(p, &["init", "-q", "-b", "main"]);
        git_ok(p, &["config", "user.email", "t@example.invalid"]);
        git_ok(p, &["config", "user.name", "t"]);
        put(p, "notes.txt", "x\n");
        put(p, "docs/guide.md", "# a checked-in document\n");
        put(p, "src/zz_stub.rs", "// a\n");
        put(p, "tests/card_cli.rs", "fn main() {}\n");
        if tables.area_tables {
            put(p, "tests/areas.toml", AREAS);
            put(p, "tests/fixtures/state_folder_literals/allow.tsv", ALLOW);
        }
        if tables.scan_list {
            match scan_list {
                Some(bytes) => put_bytes(p, "tools/repo_scans.json", bytes),
                None => put(p, "tools/repo_scans.json", SCANS),
            }
        }
        if tables.exempt_file {
            put(p, "tests/repo_scans_table.rs", EXEMPT_FILE);
        }
        git_ok(p, &["add", "-A"]);
        git_ok(p, &["commit", "-q", "-m", "base"]);
        let base = git(p, &["rev-parse", "HEAD"]).trim().to_string();
        for (rel, bytes) in changed {
            put_bytes(p, rel, bytes);
        }
        git_ok(p, &["add", "-A"]);
        git_ok(p, &["commit", "-q", "-m", "change"]);
        let head = git(p, &["rev-parse", "HEAD"]).trim().to_string();
        Fixture { dir, base, head }
    }

    fn root(&self) -> &Path {
        self.dir.path()
    }

    fn range(&self) -> String {
        format!("{}..{}", self.base, self.head)
    }
}

fn changed_bytes(changed: &[(String, String)]) -> Vec<(String, Vec<u8>)> {
    changed
        .iter()
        .map(|(p, t)| (p.clone(), t.clone().into_bytes()))
        .collect()
}

/// A folder of fake tools. Every fake appends `<tool> <args>` to `calls.log` and exits 0.
struct Tools {
    dir: tempfile::TempDir,
}

impl Tools {
    fn new(names: &[&str]) -> Tools {
        let dir = tempfile::tempdir().unwrap();
        for n in names {
            let script = format!("#!/bin/sh\necho \"{n} $*\" >> \"$FAKE_CALLS\"\nexit 0\n");
            let path = dir.path().join(n);
            fs::write(&path, script).unwrap();
            #[cfg(unix)]
            {
                use std::os::unix::fs::PermissionsExt;
                fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
            }
        }
        Tools { dir }
    }

    fn calls_text(&self) -> String {
        fs::read_to_string(self.dir.path().join("calls.log")).unwrap_or_default()
    }

    /// PATH with the fakes first and only `git` and the system's shell tools after them.
    fn path(&self) -> String {
        let git_dir = Command::new("sh")
            .args(["-c", "dirname \"$(command -v git)\""])
            .output()
            .unwrap();
        format!(
            "{}:{}:/usr/bin:/bin",
            self.dir.path().display(),
            String::from_utf8_lossy(&git_dir.stdout).trim()
        )
    }
}

fn slot_dir() -> tempfile::TempDir {
    let d = tempfile::tempdir().unwrap();
    fs::write(
        d.path().join("limits.json"),
        r#"{"max":4,"min_free_gb":0,"max_swap_pct":100,"min_mem_free_pct":0}"#,
    )
    .unwrap();
    d
}

/// The whole answer: `(ok, result or error)`. With `tools` the run is the real run under fake
/// tools and a slot folder of its own; without them it is `--list`.
fn affected(fx: &Fixture, tools: Option<&Tools>) -> (bool, Value) {
    let slots = slot_dir();
    let mut cmd = Command::new(BIN);
    cmd.args(["check", "affected", &fx.range(), "--dir"])
        .arg(fx.root())
        .env_remove("MOAI_HOME")
        .env_remove("CARGO_TARGET_DIR")
        .env_remove("RUST_TEST_THREADS")
        .env_remove("NEXTEST_TEST_THREADS")
        .stdin(Stdio::null());
    match tools {
        Some(t) => {
            cmd.arg("--slot-dir")
                .arg(slots.path())
                .env("PATH", t.path())
                .env("FAKE_CALLS", t.dir.path().join("calls.log"));
        }
        None => {
            cmd.arg("--list");
        }
    }
    let out = cmd.output().unwrap();
    let text = String::from_utf8_lossy(&out.stdout).to_string();
    let v: Value = serde_json::from_str(&text)
        .unwrap_or_else(|_| panic!("{text} / {}", String::from_utf8_lossy(&out.stderr)));
    let ok = v["ok"] == Value::Bool(true);
    (
        ok,
        if ok {
            v["result"].clone()
        } else {
            v["error"].clone()
        },
    )
}

fn fake_tools() -> Tools {
    Tools::new(&["cargo", "python3", "rustfmt"])
}

fn labels(r: &Value) -> Vec<String> {
    r["steps"]
        .as_array()
        .unwrap()
        .iter()
        .map(|s| s["label"].as_str().unwrap().to_string())
        .collect()
}

/// The log of the step `label`.
fn step_log(r: &Value, label: &str) -> String {
    let step = r["steps"]
        .as_array()
        .unwrap()
        .iter()
        .find(|s| s["label"] == label)
        .unwrap_or_else(|| panic!("no step {label} in {r}"));
    fs::read_to_string(PathBuf::from(step["log"].as_str().unwrap())).unwrap()
}

/// No heavy cargo step ran and no Python did.
fn assert_nothing_heavy_ran(tools: &Tools) {
    let calls = tools.calls_text();
    for never in ["clippy", "--lib", "--bin moai-cli", "--test ", "python3"] {
        assert!(!calls.contains(never), "{never} ran: {calls}");
    }
}

fn file(name: &str, text: String) -> (String, String) {
    (format!("tests/{name}.rs"), text)
}

fn plain(name: &str) -> (String, String) {
    file(name, "fn main() {}\n".to_string())
}

/// The run under fake tools stops at `test_tables`, ran nothing heavy and left the one step log;
/// the log is returned.
fn stopped_log(fx: &Fixture) -> String {
    let tools = fake_tools();
    let (ok, r) = affected(fx, Some(&tools));
    assert!(ok, "{r}");
    let line = r["summary"].as_str().unwrap();
    assert!(line.contains(" test_tables=1 "), "{line}");
    assert!(
        line.ends_with(" stopped_at=test_tables result=FAIL"),
        "{line}"
    );
    assert_eq!(r["result"], "FAIL");
    assert_nothing_heavy_ran(&tools);
    step_log(&r, "test_tables")
}

fn line_of<'a>(log: &'a str, needle: &str) -> &'a str {
    log.lines()
        .find(|l| l.contains(needle))
        .unwrap_or_else(|| panic!("no line for {needle}: {log}"))
}

/// SPEC REQ-1: an added test file that walks the tree and is neither named in the scan list nor
/// exempt gets one line that names the file and both places to register it; a registered file, an
/// exempt file and a file that walks nothing get none.
#[test]
fn a_new_walking_test_file_missing_from_the_scan_list_stops_before_the_heavy_steps() {
    let changed = [
        file("check_walk_cli", walker()),
        file("check_listed_walk_cli", walker()),
        file("check_exempt_walk_cli", walker()),
        plain("check_plain_cli"),
    ];
    let fx = Fixture::new(ALL, &changed);
    let log = stopped_log(&fx);
    let line = line_of(&log, "tests/check_walk_cli.rs");
    assert!(
        line.contains("tools/repo_scans.json") && line.contains("EXEMPT"),
        "{line}"
    );
    assert_eq!(log.lines().count(), 1, "{log}");
}

/// SPEC REQ-2, REQ-3, REQ-5: a new file that reads a file outside `tests/` with `include_str!` or
/// `include_bytes!`, and one that opens a checked-in document by its path, each get a line with the
/// file and the path when no `always` scan covers it; a file an exemption list keeps out gets none.
#[test]
fn a_new_reader_without_an_always_scan_stops_before_the_heavy_steps() {
    let changed = [
        file(
            "check_inc_cli",
            format!(
                "const A: &str = {};\n",
                include("include_str", "../src/zz_new_table.rs")
            ),
        ),
        file(
            "check_incb_cli",
            format!(
                "const A: &[u8] = {};\n",
                include("include_bytes", "../docs/guide.md")
            ),
        ),
        file("check_read_cli", reader()),
        file(
            "check_exempt_inc_cli",
            format!(
                "const A: &str = {};\n",
                include("include_str", "../src/zz_new_table.rs")
            ),
        ),
        file("check_exempt_read_cli", reader()),
    ];
    let fx = Fixture::new(ALL, &changed);
    let log = stopped_log(&fx);
    let inc = line_of(&log, "tests/check_inc_cli.rs");
    assert!(inc.contains("src/zz_new_table.rs"), "{inc}");
    let incb = line_of(&log, "tests/check_incb_cli.rs");
    assert!(incb.contains("docs/guide.md"), "{incb}");
    let read = line_of(&log, "tests/check_read_cli.rs");
    assert!(
        read.contains("docs/guide.md") && read.contains("tools/repo_scans.json"),
        "{read}"
    );
    assert!(
        !log.contains("check_exempt_"),
        "an exempt file has no line: {log}"
    );
    assert_eq!(log.lines().count(), 3, "{log}");
}

/// SPEC REQ-4, REQ-5, REQ-6: a registered or exempt new file passes and the heavy steps run; a
/// reader named by an `always` scan that runs no `--test` check of it stops the run; the step is
/// planned when only the scan list exists and is not planned without a table or without an added
/// test file.
#[test]
fn a_registered_or_exempt_new_test_file_passes_and_the_step_is_planned_only_with_a_table() {
    let tools = fake_tools();
    let good = [
        file("check_reader_ok_cli", reader()),
        file("check_listed_walk_cli", walker()),
        file("check_exempt_walk_cli", walker()),
        file("check_exempt_read_cli", reader()),
    ];
    let fx = Fixture::new(ALL, &good);
    let (ok, r) = affected(&fx, Some(&tools));
    assert!(ok, "{r}");
    let line = r["summary"].as_str().unwrap();
    assert!(line.contains(" test_tables=0 "), "{line}");
    assert!(line.ends_with(" result=PASS"), "{line}");
    assert!(tools.calls_text().contains("clippy"), "heavy steps ran");

    let nocheck = Fixture::new(ALL, &[file("check_nocheck_cli", reader())]);
    let log = stopped_log(&nocheck);
    let line = line_of(&log, "reads-nocheck");
    assert!(line.contains("--test check_nocheck_cli"), "{line}");
    assert_eq!(log.lines().count(), 1, "{log}");

    let only_scans = Tables {
        area_tables: false,
        scan_list: true,
        exempt_file: false,
    };
    let fresh = [file("check_walk_cli", walker())];
    let (ok, r) = affected(&Fixture::new(only_scans, &fresh), None);
    assert!(ok, "{r}");
    assert!(labels(&r).contains(&"test_tables".to_string()), "{r}");

    let none = Tables {
        area_tables: false,
        scan_list: false,
        exempt_file: false,
    };
    let (ok, r) = affected(&Fixture::new(none, &fresh), None);
    assert!(ok, "{r}");
    assert!(!labels(&r).contains(&"test_tables".to_string()), "{r}");

    let src_only = [("src/zz_stub.rs".to_string(), "// b\n".to_string())];
    let (ok, r) = affected(&Fixture::new(ALL, &src_only), None);
    assert!(ok, "{r}");
    assert!(!labels(&r).contains(&"test_tables".to_string()), "{r}");
}

/// SPEC REQ-7: a scan list that exists but cannot be read (invalid UTF-8), is no JSON, or has no
/// `scans` list stops the run with one line that names the file and the error; a readable list
/// gives no such line for the same range.
#[test]
fn a_scan_list_that_cannot_be_read_stops_before_the_heavy_steps() {
    let fresh = [file("check_plain_cli", "fn main() {}\n".to_string())];
    let broken: [(&str, &[u8]); 3] = [
        ("invalid UTF-8", &[0xff, 0xfe, b'{', b'}']),
        ("not JSON", b"{ not json"),
        ("no scans list", b"{\"about\":\"x\"}"),
    ];
    for (what, bytes) in broken {
        let fx = Fixture::with_bytes(ALL, &changed_bytes(&fresh), Some(bytes));
        let log = stopped_log(&fx);
        assert_eq!(log.lines().count(), 1, "{what}: {log}");
        assert!(
            log.contains("tools/repo_scans.json") && log.contains("cannot be read"),
            "{what}: {log}"
        );
    }
    let fx = Fixture::new(ALL, &fresh);
    let tools = fake_tools();
    let (ok, r) = affected(&fx, Some(&tools));
    assert!(ok, "{r}");
    let line = r["summary"].as_str().unwrap();
    assert!(line.contains(" test_tables=0 "), "{line}");
    assert!(line.ends_with(" result=PASS"), "{line}");
}
