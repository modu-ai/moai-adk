//! SPEC-20261124-new-test-file-precheck: `check affected` checks a top-level `tests/*.rs` file that
//! a range adds against `tests/areas.toml` and the state-folder allow list before any heavy step,
//! and stops there with the line to add. Temporary git repositories and fake `cargo`, `python3`
//! and `rustfmt` scripts only: the real tools, the real HOME and the machine's heavy slots are
//! never used.
//!
//! The state folder names are built at run time: this file spells neither name, so it needs no row
//! of its own in the allow list.
mod common;
use common::git::{git, git_ok};
use serde_json::Value;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

const BIN: &str = crate::common::MOAI_BIN;

/// The earlier state folder name (the new name plus `-cli`) and the new one.
fn earlier() -> String {
    format!(".{}-{}", "moai", "cli")
}

fn new_name() -> String {
    format!(".{}", "moai")
}

const AREAS: &str = "\
# fixture table
[areas]
card = [\"card\"]
spec = [\"spec\", \"check\"]
lane = [\"lane\"]

[solo]
files = [\"check_solo_cli\"]
";

const ALLOW: &str = "\
# fixture allow list
tests/check_listed_cli.rs\t0\t2
tests/check_over_cli.rs\t0\t1
";

fn put(root: &Path, rel: &str, text: &str) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}

struct Fixture {
    dir: tempfile::TempDir,
    base: String,
    head: String,
}

impl Fixture {
    /// A repository whose base holds the two tables (unless `tables` is false) and one commit on
    /// top of it that adds or changes `changed`.
    fn new(tables: bool, changed: &[(&str, String)]) -> Fixture {
        let dir = tempfile::tempdir().unwrap();
        let p = dir.path();
        git_ok(p, &["init", "-q", "-b", "main"]);
        git_ok(p, &["config", "user.email", "t@example.invalid"]);
        git_ok(p, &["config", "user.name", "t"]);
        put(p, "notes.txt", "x\n");
        put(p, "src/zz_stub.rs", "// a\n");
        put(p, "tests/card_cli.rs", "fn main() {}\n");
        put(p, "tests/lane/main.rs", "// a moved area\n");
        if tables {
            put(p, "tests/areas.toml", AREAS);
            put(p, "tests/fixtures/state_folder_literals/allow.tsv", ALLOW);
        }
        git_ok(p, &["add", "-A"]);
        git_ok(p, &["commit", "-q", "-m", "base"]);
        let base = git(p, &["rev-parse", "HEAD"]).trim().to_string();
        for (rel, text) in changed {
            put(p, rel, text);
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

fn keys(line: &str) -> Vec<&str> {
    line.split(' ')
        .skip(1)
        .filter_map(|w| w.split_once('=').map(|(k, _)| k))
        .collect()
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

fn plain(name: &str) -> (String, String) {
    (format!("tests/{name}.rs"), "fn main() {}\n".to_string())
}

/// SPEC REQ-1, REQ-2, REQ-5: an added file with an unlisted first word, a file of a moved area and
/// a file named like an area crate each get one line, and the run stops before the heavy steps.
#[test]
fn a_new_test_file_missing_from_the_area_table_stops_before_the_heavy_steps() {
    let added = [
        plain("zzfresh_cli"),
        plain("lane_extra_cli"),
        plain("spec"),
        plain("check_solo_cli"),
    ];
    let changed: Vec<(&str, String)> = added.iter().map(|(p, t)| (p.as_str(), t.clone())).collect();
    let fx = Fixture::new(true, &changed);
    let tools = fake_tools();
    let (ok, r) = affected(&fx, Some(&tools));
    assert!(ok, "{r}");
    let line = r["summary"].as_str().unwrap();
    assert!(line.contains(" test_tables=1 "), "{line}");
    assert!(
        line.ends_with(" stopped_at=test_tables result=FAIL"),
        "{line}"
    );
    assert_eq!(r["result"], "FAIL");
    assert_nothing_heavy_ran(&tools);
    let log = step_log(&r, "test_tables");
    let line_of = |needle: &str| {
        log.lines()
            .find(|l| l.contains(needle))
            .unwrap_or_else(|| panic!("no line for {needle}: {log}"))
            .to_string()
    };
    let fresh = line_of("tests/zzfresh_cli.rs");
    assert!(
        fresh.contains("tests/areas.toml") && fresh.contains("\"zzfresh\""),
        "{fresh}"
    );
    let moved = line_of("tests/lane_extra_cli.rs");
    assert!(
        moved.contains("area `lane` is moved") && moved.contains("tests/lane/main.rs"),
        "{moved}"
    );
    let krate = line_of("tests/spec.rs");
    assert!(krate.contains("area crate `spec`"), "{krate}");
    assert!(
        !log.contains("check_solo_cli"),
        "a solo file is no violation: {log}"
    );
    assert_eq!(log.lines().count(), 3, "{log}");
}

/// SPEC REQ-3, REQ-4: a new file that spells a state folder name needs a row; a row that allows
/// fewer than the file spells is no row; a row that allows enough passes.
#[test]
fn a_new_test_file_missing_from_the_state_folder_allow_list_stops_before_the_heavy_steps() {
    let literal = format!(
        "fn a() {{ let _ = \"{}/x\"; let _ = \"{}/y\"; }}\n",
        earlier(),
        new_name()
    );
    let twice = format!(
        "fn a() {{ let _ = \"{n}/x\"; let _ = \"{n}/y\"; }}\n",
        n = new_name()
    );
    let once = format!("fn a() {{ let _ = \"{}/x\"; }}\n", new_name());
    let changed = [
        ("tests/check_literal_cli.rs", literal),
        ("tests/check_over_cli.rs", twice),
        ("tests/check_listed_cli.rs", once),
    ];
    let fx = Fixture::new(true, &changed);
    let tools = fake_tools();
    let (ok, r) = affected(&fx, Some(&tools));
    assert!(ok, "{r}");
    let line = r["summary"].as_str().unwrap();
    assert!(
        line.ends_with(" test_tables=1 stopped_at=test_tables result=FAIL"),
        "{line}"
    );
    assert_nothing_heavy_ran(&tools);
    let log = step_log(&r, "test_tables");
    assert!(
        log.contains("tests/check_literal_cli.rs\t1\t1"),
        "the exact row: {log}"
    );
    let over = log
        .lines()
        .find(|l| l.contains("tests/check_over_cli.rs"))
        .unwrap_or_else(|| panic!("{log}"));
    assert!(
        over.contains(" 2 times")
            && over.contains("allows 1")
            && over.contains("tests/check_over_cli.rs\t0\t2"),
        "{over}"
    );
    assert!(!log.contains("check_listed_cli"), "{log}");
    assert_eq!(log.lines().count(), 2, "{log}");
}

/// SPEC REQ-5, REQ-6, REQ-7: a registered new file passes and the run goes on to the heavy steps,
/// and a range that adds no test file, or a repository without tables, plans no such step.
#[test]
fn a_registered_new_test_file_passes_and_other_ranges_plan_no_table_step() {
    let tools = fake_tools();
    let fx = Fixture::new(
        true,
        &[("tests/check_fine_cli.rs", "fn main() {}\n".to_string())],
    );
    let (ok, r) = affected(&fx, Some(&tools));
    assert!(ok, "{r}");
    let line = r["summary"].as_str().unwrap();
    let words = keys(line);
    let at = |k: &str| {
        words
            .iter()
            .position(|w| *w == k)
            .unwrap_or_else(|| panic!("{k}: {line}"))
    };
    assert_eq!(at("test_tables"), at("fmt") + 1, "{line}");
    assert!(at("test_tables") < at("clippy"), "{line}");
    assert!(line.contains(" test_tables=0 "), "{line}");
    assert!(line.ends_with(" result=PASS"), "{line}");
    assert!(tools.calls_text().contains("clippy"), "heavy steps ran");

    let src_only = Fixture::new(true, &[("src/zz_stub.rs", "// b\n".to_string())]);
    let (ok, r) = affected(&src_only, None);
    assert!(ok, "{r}");
    assert!(!labels(&r).contains(&"test_tables".to_string()), "{r}");

    let fresh = ("tests/zzfresh_cli.rs", "fn main() {}\n".to_string());
    let no_tables = Fixture::new(false, std::slice::from_ref(&fresh));
    let (ok, r) = affected(&no_tables, None);
    assert!(ok, "{r}");
    assert!(!labels(&r).contains(&"test_tables".to_string()), "{r}");
}

/// SPEC REQ-7: `--list` shows the step, cheap, before the first heavy step.
#[test]
fn list_shows_the_table_step_before_the_heavy_steps() {
    let fresh = ("tests/zzfresh_cli.rs", "fn main() {}\n".to_string());
    let fx = Fixture::new(true, std::slice::from_ref(&fresh));
    let (ok, r) = affected(&fx, None);
    assert!(ok, "{r}");
    let steps = r["steps"].as_array().unwrap();
    let at = steps
        .iter()
        .position(|s| s["label"] == "test_tables")
        .unwrap_or_else(|| panic!("{r}"));
    assert_eq!(steps[at]["heavy"], Value::Bool(false), "{r}");
    let clippy = steps.iter().position(|s| s["label"] == "clippy").unwrap();
    assert!(at < clippy, "{r}");
    assert_eq!(r["writes"], Value::Bool(false));
}
