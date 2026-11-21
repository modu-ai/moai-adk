//! SPEC-20261121-check-receipts: `check affected` leaves a receipt for a passing tree and reuses
//! it for the same tree, planned steps and tool versions. Temporary git repositories and fake
//! `cargo`, `rustc`, `python3` and `rustfmt` scripts only: the real tools, the real HOME and the
//! machine's heavy slots are never used.
#![cfg(unix)]

mod common;
use common::git::{git, git_ok};
use serde_json::Value;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

const BIN: &str = crate::common::MOAI_BIN;

fn put(root: &Path, rel: &str, text: &str) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}

/// A repository with one commit on top of the base that changes a source file.
struct Fixture {
    dir: tempfile::TempDir,
    base: String,
    head: String,
}

impl Fixture {
    fn new() -> Fixture {
        let dir = tempfile::tempdir().unwrap();
        let p = dir.path();
        git_ok(p, &["init", "-q", "-b", "main"]);
        git_ok(p, &["config", "user.email", "t@example.invalid"]);
        git_ok(p, &["config", "user.name", "t"]);
        put(p, "README.md", "x\n");
        put(p, "src/update_models.rs", "// a\n");
        put(p, "tests/update_command_cli.rs", "fn main() {}\n");
        put(p, "tests/common/mod.rs", "// helper\n");
        git_ok(p, &["add", "-A"]);
        git_ok(p, &["commit", "-q", "-m", "base"]);
        let base = git(p, &["rev-parse", "HEAD"]).trim().to_string();
        put(p, "src/update_models.rs", "// b\n");
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

    /// One more commit that changes a tracked file, so the head has another tree.
    fn advance(&mut self) {
        put(self.root(), "src/update_models.rs", "// c\n");
        git_ok(self.root(), &["add", "-A"]);
        git_ok(self.root(), &["commit", "-q", "-m", "again"]);
        self.head = git(self.root(), &["rev-parse", "HEAD"]).trim().to_string();
    }

    /// The folder the receipts go to: under the state folder a project without a marker uses.
    fn receipts(&self) -> PathBuf {
        self.root().join(".moai-cli").join("check-receipts")
    }

    fn receipt_files(&self) -> Vec<PathBuf> {
        let Ok(entries) = fs::read_dir(self.receipts()) else {
            return Vec::new();
        };
        let mut files: Vec<PathBuf> = entries
            .flatten()
            .map(|e| e.path())
            .filter(|p| p.extension().is_some_and(|e| e == "json"))
            .collect();
        files.sort();
        files
    }
}

/// A folder of fake tools. A step appends `<tool> <args>` to `calls.log`; `--version` prints
/// `FAKE_VER` (default `1.0`) and exits `FAKE_VER_EXIT` (default 0) without a call line, so
/// `calls.log` holds the steps that ran and nothing else. One named in `FAKE_FAIL` exits 3.
struct Tools {
    dir: tempfile::TempDir,
}

impl Tools {
    fn new() -> Tools {
        let dir = tempfile::tempdir().unwrap();
        for n in ["cargo", "rustc", "python3", "rustfmt"] {
            let script = format!(
                "#!/bin/sh\ncase \"$1\" in --version) test -n \"$FAKE_VER_EXIT\" && exit \"$FAKE_VER_EXIT\"; echo \"{n} ${{FAKE_VER:-1.0}}\"; exit 0;; esac\n\
                 echo \"{n} $*\" >> \"$FAKE_CALLS\"\ncase \"$FAKE_FAIL\" in\n  '') ;;\n  *) case \" $* \" in *\" $FAKE_FAIL \"*) exit 3 ;; esac ;;\nesac\nexit 0\n"
            );
            let path = dir.path().join(n);
            fs::write(&path, script).unwrap();
            use std::os::unix::fs::PermissionsExt;
            fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
        }
        Tools { dir }
    }

    fn calls_text(&self) -> String {
        fs::read_to_string(self.dir.path().join("calls.log")).unwrap_or_default()
    }

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

/// What a run was given besides the fixture.
#[derive(Default)]
struct Env<'a> {
    fail: &'a str,
    ver: Option<&'a str>,
    ver_exit: Option<&'a str>,
    fresh: bool,
}

/// One `check affected` run with fresh fake tools: the answer and the tools it used.
fn run(fx: &Fixture, env: &Env) -> (Value, Tools) {
    let tools = Tools::new();
    let slots = slot_dir();
    let mut cmd = Command::new(BIN);
    cmd.args(["check", "affected", &fx.range(), "--dir"])
        .arg(fx.root())
        .args(["--slot-dir"])
        .arg(slots.path());
    if env.fresh {
        cmd.arg("--fresh");
    }
    cmd.env_remove("MOAI_HOME")
        .env("PATH", tools.path())
        .env("FAKE_CALLS", tools.dir.path().join("calls.log"))
        .env("FAKE_FAIL", env.fail)
        .stdin(Stdio::null());
    if let Some(v) = env.ver {
        cmd.env("FAKE_VER", v);
    }
    if let Some(c) = env.ver_exit {
        cmd.env("FAKE_VER_EXIT", c);
    }
    let out = cmd.output().unwrap();
    let text = String::from_utf8_lossy(&out.stdout).to_string();
    let v: Value = serde_json::from_str(&text)
        .unwrap_or_else(|_| panic!("{text} / {}", String::from_utf8_lossy(&out.stderr)));
    assert_eq!(v["ok"], Value::Bool(true), "{v}");
    (v["result"].clone(), tools)
}

fn state(r: &Value) -> &str {
    r["receipt"]["state"].as_str().unwrap_or("missing")
}

/// AC-1 (REQ-1, REQ-9): a passing run leaves one receipt and says `written`.
#[test]
fn a_passing_run_leaves_one_receipt() {
    let fx = Fixture::new();
    let (r, _) = run(&fx, &Env::default());
    assert_eq!(r["result"], "PASS", "{r}");
    assert_eq!(state(&r), "written", "{r}");
    assert!(
        r["receipt"]["key"].as_str().is_some_and(|k| !k.is_empty()),
        "{r}"
    );
    let files = fx.receipt_files();
    assert_eq!(files.len(), 1, "{files:?}");
    let stem = files[0].file_stem().unwrap().to_string_lossy().to_string();
    assert_eq!(Some(stem.as_str()), r["receipt"]["key"].as_str());
}

/// AC-2 (REQ-2, REQ-9): a run with a failing step leaves no receipt.
#[test]
fn a_failing_run_leaves_no_receipt() {
    let fx = Fixture::new();
    let (r, _) = run(
        &fx,
        &Env {
            fail: "fmt",
            ..Env::default()
        },
    );
    assert_eq!(r["result"], "FAIL", "{r}");
    assert_eq!(state(&r), "none", "{r}");
    assert!(fx.receipt_files().is_empty());
}

/// AC-3 (REQ-3, REQ-9): the same tree again runs no step and the summary says so.
#[test]
fn the_same_tree_reuses_the_receipt() {
    let fx = Fixture::new();
    let (first, _) = run(&fx, &Env::default());
    let (second, tools) = run(&fx, &Env::default());
    assert_eq!(state(&second), "reused", "{second}");
    assert_eq!(tools.calls_text(), "", "a step ran on a reused receipt");
    assert_eq!(second["result"], "PASS", "{second}");
    let line = second["summary"].as_str().unwrap();
    assert!(line.ends_with(" receipt=reused result=PASS"), "{line}");
    assert_eq!(second["steps"], first["steps"], "the stored step results");
    assert_eq!(second["receipt"]["key"], first["receipt"]["key"]);
}

/// AC-4 (REQ-4): another tree and another tool version each run every step.
#[test]
fn any_change_runs_every_step() {
    let mut fx = Fixture::new();
    let (first, _) = run(&fx, &Env::default());
    assert_eq!(state(&first), "written", "{first}");
    let (newer_tool, tools) = run(
        &fx,
        &Env {
            ver: Some("2.0"),
            ..Env::default()
        },
    );
    assert_eq!(state(&newer_tool), "written", "{newer_tool}");
    assert_ne!(newer_tool["receipt"]["key"], first["receipt"]["key"]);
    assert!(
        tools.calls_text().contains("cargo"),
        "{}",
        tools.calls_text()
    );
    fx.advance();
    let (other_tree, tools) = run(&fx, &Env::default());
    assert_eq!(state(&other_tree), "written", "{other_tree}");
    assert_ne!(other_tree["receipt"]["key"], first["receipt"]["key"]);
    assert!(
        tools.calls_text().contains("cargo"),
        "{}",
        tools.calls_text()
    );
}

/// AC-5 (REQ-5, REQ-9): `--fresh` runs every step and replaces the receipt of the key.
#[test]
fn fresh_runs_and_replaces() {
    let fx = Fixture::new();
    let (first, _) = run(&fx, &Env::default());
    let file = fx.receipt_files().remove(0);
    fs::write(&file, "{\"key\":\"stale\"}").unwrap();
    let (fresh, tools) = run(
        &fx,
        &Env {
            fresh: true,
            ..Env::default()
        },
    );
    assert_eq!(state(&fresh), "fresh", "{fresh}");
    assert!(
        tools.calls_text().contains("cargo"),
        "{}",
        tools.calls_text()
    );
    assert_eq!(fresh["receipt"]["key"], first["receipt"]["key"]);
    // The receipt was written again: the next plain run reuses it.
    let (again, tools) = run(&fx, &Env::default());
    assert_eq!(state(&again), "reused", "{again}");
    assert_eq!(tools.calls_text(), "");
}

/// AC-6 (REQ-6): a receipt that cannot be used is ignored and every step runs.
#[test]
fn a_broken_receipt_is_ignored() {
    let fx = Fixture::new();
    let (first, _) = run(&fx, &Env::default());
    let file = fx.receipt_files().remove(0);
    for broken in [
        "not json at all".to_string(),
        "{\"key\":\"another\",\"result\":\"PASS\"}".to_string(),
    ] {
        fs::write(&file, broken).unwrap();
        let (r, tools) = run(&fx, &Env::default());
        assert_eq!(state(&r), "written", "{r}");
        assert!(
            tools.calls_text().contains("cargo"),
            "{}",
            tools.calls_text()
        );
        assert_eq!(r["receipt"]["key"], first["receipt"]["key"]);
    }
    // An unreadable one: a folder in the place of the file.
    fs::remove_file(&file).unwrap();
    fs::create_dir(&file).unwrap();
    let (r, tools) = run(&fx, &Env::default());
    assert_eq!(r["result"], "PASS", "{r}");
    assert!(
        tools.calls_text().contains("cargo"),
        "{}",
        tools.calls_text()
    );
}

/// AC-7 (REQ-7): a receipt folder that is a symlink gets no receipt.
#[test]
fn a_symlinked_folder_gets_no_receipt() {
    let fx = Fixture::new();
    let elsewhere = tempfile::tempdir().unwrap();
    fs::create_dir_all(fx.root().join(".moai-cli")).unwrap();
    std::os::unix::fs::symlink(elsewhere.path(), fx.receipts()).unwrap();
    let (r, _) = run(&fx, &Env::default());
    assert_eq!(r["result"], "PASS", "{r}");
    assert_eq!(state(&r), "none", "{r}");
    assert_eq!(fs::read_dir(elsewhere.path()).unwrap().count(), 0);
}

/// AC-8 (REQ-8): unreadable tool versions run every step and write no receipt.
#[test]
fn unreadable_versions_leave_no_receipt() {
    let fx = Fixture::new();
    let env = Env {
        ver_exit: Some("1"),
        ..Env::default()
    };
    let (r, tools) = run(&fx, &env);
    assert_eq!(r["result"], "PASS", "{r}");
    assert_eq!(state(&r), "none", "{r}");
    assert!(fx.receipt_files().is_empty());
    assert!(tools.calls_text().contains("cargo"));
    let (again, tools) = run(&fx, &env);
    assert_eq!(state(&again), "none", "{again}");
    assert!(tools.calls_text().contains("cargo"));
}
