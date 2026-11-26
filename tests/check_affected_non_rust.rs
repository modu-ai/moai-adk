//! SPEC-20261125-check-affected-non-rust: in a repository without Rust code `check affected`
//! plans no cargo step, runs the checks the project profile declares and answers `not_run`, never
//! `PASS`, when it declares none. Every case builds a temporary git repository and a temporary
//! slot folder; the real HOME, the real tools and the machine's heavy slots are not used.
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

/// A profile as the interview writes it, with the declared checks given as `(id, command)`.
fn profile(checks: &[(&str, &str)]) -> String {
    let rows: Vec<Value> = checks
        .iter()
        .map(|(id, command)| serde_json::json!({"id": id, "command": command}))
        .collect();
    serde_json::json!({
        "schema_version": 2, "project_id": "demo", "purpose": "A small web app",
        "tools": ["TypeScript"], "development_methods": ["TDD"], "hosts": ["claude"],
        "checks": rows,
    })
    .to_string()
}

/// A repository holding `base` files, one commit on top of it that writes `changed`, and the
/// range between them.
struct Repo {
    dir: tempfile::TempDir,
    range: String,
}

impl Repo {
    fn new(base: &[(&str, &str)], changed: &[(&str, &str)]) -> Repo {
        let dir = tempfile::tempdir().unwrap();
        let p = dir.path();
        git_ok(p, &["init", "-q", "-b", "main"]);
        git_ok(p, &["config", "user.email", "t@example.invalid"]);
        git_ok(p, &["config", "user.name", "t"]);
        put(p, "README.md", "x\n");
        for (rel, text) in base {
            put(p, rel, text);
        }
        git_ok(p, &["add", "-A"]);
        git_ok(p, &["commit", "-q", "-m", "base"]);
        let base_sha = git(p, &["rev-parse", "HEAD"]).trim().to_string();
        for (rel, text) in changed {
            put(p, rel, text);
        }
        git_ok(p, &["add", "-A"]);
        git_ok(p, &["commit", "-q", "-m", "change"]);
        let head = git(p, &["rev-parse", "HEAD"]).trim().to_string();
        Repo {
            dir,
            range: format!("{base_sha}..{head}"),
        }
    }

    /// The repository the interview's TypeScript project would be: no `Cargo.toml`, no `.rs`.
    fn web(profile_text: Option<&str>) -> Repo {
        let mut base = vec![
            ("package.json", "{}\n"),
            ("src/index.ts", "export const a = 1;\n"),
        ];
        if let Some(text) = profile_text {
            base.push((".moai/project.yaml", text));
        }
        Repo::new(&base, &[("src/index.ts", "export const a = 2;\n")])
    }

    fn root(&self) -> PathBuf {
        self.dir.path().to_path_buf()
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

/// The command's whole answer: `(ok, result or error)`.
fn affected(repo: &Repo, extra: &[&str]) -> (bool, Value) {
    let home = tempfile::tempdir().unwrap();
    let out = Command::new(BIN)
        .args(["check", "affected", &repo.range, "--dir"])
        .arg(repo.root())
        .args(extra)
        .env("HOME", home.path())
        .env_remove("MOAI_HOME")
        .env_remove("CARGO_TARGET_DIR")
        .stdin(Stdio::null())
        .output()
        .unwrap();
    let text = String::from_utf8_lossy(&out.stdout).to_string();
    let v: Value = serde_json::from_str(text.lines().next().unwrap_or(""))
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

fn list(repo: &Repo) -> Value {
    let (ok, r) = affected(repo, &["--list"]);
    assert!(ok, "{r}");
    r
}

/// A run with its own slot folder; the slot folder lives as long as the call.
fn run(repo: &Repo) -> Value {
    let slots = slot_dir();
    let (ok, r) = affected(repo, &["--slot-dir", &slots.path().to_string_lossy()]);
    assert!(ok, "{r}");
    r
}

fn labels(r: &Value) -> Vec<String> {
    r["steps"]
        .as_array()
        .unwrap_or_else(|| panic!("no steps in {r}"))
        .iter()
        .map(|s| s["label"].as_str().unwrap().to_string())
        .collect()
}

fn summary_file(r: &Value) -> String {
    let out = Path::new(r["out"].as_str().unwrap());
    fs::read_to_string(out.join("summary.txt")).unwrap()
}

const DECLARED: &[(&str, &str)] = &[("lint", "echo lint"), ("test", "echo test")];

// AC-1 (REQ-1, REQ-2): no cargo step; one `sh -c` step per declared check, in the declared order.
#[test]
fn a_repository_without_rust_plans_no_cargo_step() {
    let repo = Repo::web(Some(&profile(DECLARED)));
    let r = list(&repo);
    assert_eq!(
        labels(&r),
        ["diff_check", "check:lint", "check:test"],
        "{r}"
    );
    assert_eq!(r["rust_targets"], serde_json::json!([]), "{r}");
    let steps = r["steps"].as_array().unwrap();
    assert_eq!(
        steps[1]["command"],
        serde_json::json!(["sh", "-c", "echo lint"])
    );
    assert_eq!(
        steps[2]["command"],
        serde_json::json!(["sh", "-c", "echo test"])
    );
    assert_eq!(steps[0]["heavy"], Value::Bool(false));
    assert_eq!(steps[1]["heavy"], Value::Bool(true));
    assert_eq!(steps[2]["heavy"], Value::Bool(true));
}

// AC-2 (REQ-2, REQ-6): the checks run in the repository's top folder and give a PASS receipt.
#[test]
fn the_declared_checks_run_in_order_and_give_the_receipt() {
    let repo = Repo::web(Some(&profile(&[
        ("where", "pwd -P"),
        ("then", "echo second"),
    ])));
    let r = run(&repo);
    assert_eq!(r["result"], "PASS", "{r}");
    let line = r["summary"].as_str().unwrap();
    assert!(line.ends_with(" check=0 result=PASS"), "{line}");
    assert_eq!(summary_file(&r), format!("{line}\n"));
    assert_eq!(
        labels(&r),
        ["diff_check", "check:where", "check:then"],
        "{r}"
    );
    let log = fs::read_to_string(r["steps"][1]["log"].as_str().unwrap()).unwrap();
    let top = fs::canonicalize(repo.root()).unwrap();
    assert_eq!(log.trim(), top.to_string_lossy(), "the check ran elsewhere");
}

// AC-3 (REQ-4): with `--keep-going` a failing check fails the run and the next check still runs.
#[test]
fn a_failed_declared_check_fails_the_run() {
    let repo = Repo::web(Some(&profile(&[
        ("lint", "exit 3"),
        ("test", "echo after"),
    ])));
    let slots = slot_dir();
    let (ok, r) = affected(
        &repo,
        &[
            "--slot-dir",
            &slots.path().to_string_lossy(),
            "--keep-going",
        ],
    );
    assert!(ok, "{r}");
    assert_eq!(r["result"], "FAIL", "{r}");
    let line = r["summary"].as_str().unwrap();
    assert!(line.contains(" check=3 "), "{line}");
    assert!(line.ends_with(" result=FAIL"), "{line}");
    assert_eq!(r["steps"][1]["exit"], 3, "{r}");
    assert_eq!(
        r["steps"][2]["exit"], 0,
        "the later check must still run: {r}"
    );
    assert_eq!(summary_file(&r), format!("{line}\n"));
}

// AC-3 (REQ-4): without `--keep-going` a failing check stops the run: the next check is skipped.
#[test]
fn a_failed_declared_check_stops_the_next_check_by_default() {
    let repo = Repo::web(Some(&profile(&[
        ("lint", "exit 3"),
        ("test", "echo after"),
    ])));
    let r = run(&repo);
    assert_eq!(r["result"], "FAIL", "{r}");
    assert_eq!(labels(&r), ["diff_check", "check:lint"], "{r}");
    assert_eq!(r["steps"][1]["exit"], 3, "{r}");
    assert_eq!(r["skipped"], serde_json::json!(["check:test"]), "{r}");
    let line = r["summary"].as_str().unwrap();
    assert!(line.ends_with(" stopped_at=check result=FAIL"), "{line}");
    assert_eq!(summary_file(&r), format!("{line}\n"));
}

// AC-4 (REQ-3, REQ-6): no profile, a profile that cannot be read and a profile without `checks`
// all answer `not_run`, never PASS, name the next step and start no cargo step.
#[test]
fn no_declared_check_answers_not_run_with_the_next_step() {
    let no_checks = serde_json::json!({
        "schema_version": 2, "project_id": "demo", "purpose": "A small web app",
        "tools": ["TypeScript"], "development_methods": ["TDD"], "hosts": ["claude"],
    })
    .to_string();
    let cases: [(&str, Option<String>); 3] = [
        ("no profile", None),
        ("a broken profile", Some("{ not json".to_string())),
        ("a profile without checks", Some(no_checks)),
    ];
    for (what, text) in cases {
        let repo = Repo::web(text.as_deref());
        let r = run(&repo);
        assert_eq!(r["result"], "not_run", "{what}: {r}");
        let line = r["summary"].as_str().unwrap();
        assert!(line.ends_with(" result=not_run"), "{what}: {line}");
        assert!(!line.contains("result=PASS"), "{what}: {line}");
        assert_eq!(summary_file(&r), format!("{line}\n"), "{what}");
        assert_eq!(labels(&r), ["diff_check"], "{what}: {r}");
        let next = r["not_run"]["next"].as_str().unwrap_or_default();
        assert!(
            next.contains("checks") && next.contains(".moai/project.yaml"),
            "{what}: the next step names the key and the file: {next}"
        );
        assert!(
            r["not_run"]["reason"]
                .as_str()
                .is_some_and(|s| !s.is_empty()),
            "{what}: {r}"
        );
        let l = list(&repo);
        assert_eq!(labels(&l), ["diff_check"], "{what}: {l}");
    }
}

// AC-5 (REQ-4): a cheap step that fails is FAIL even when no check is declared.
#[test]
fn a_failed_cheap_step_is_fail_not_not_run() {
    let repo = Repo::new(
        &[("package.json", "{}\n"), ("src/index.ts", "a\n")],
        &[("src/index.ts", "a \n")],
    );
    let r = run(&repo);
    assert_eq!(r["result"], "FAIL", "{r}");
    let line = r["summary"].as_str().unwrap();
    assert!(line.ends_with(" result=FAIL"), "{line}");
    assert!(line.contains(" diff_check=2 "), "{line}");
}

// AC-6 (REQ-5): `Cargo.toml` alone and a `.rs` file alone both keep the Rust steps, also with
// declared checks.
/// One case: what it shows, the base files `(path, text)` and the one file the change writes.
type RustCase<'a> = (&'a str, Vec<(&'a str, &'a str)>, (&'a str, &'a str));

#[test]
fn a_rust_repository_keeps_its_steps() {
    let declared = profile(DECLARED);
    let cases: [RustCase; 3] = [
        (
            "Cargo.toml only",
            vec![("Cargo.toml", "[package]\nname = \"a\"\n")],
            ("notes.md", "b\n"),
        ),
        (
            "a .rs file only",
            vec![("src/lib.rs", "// a\n")],
            ("src/lib.rs", "// b\n"),
        ),
        (
            "a nested .rs file only",
            vec![
                ("crates/a/src/lib.rs", "// a\n"),
                (".moai/project.yaml", &declared),
            ],
            ("crates/a/src/lib.rs", "// b\n"),
        ),
    ];
    for (what, base, change) in cases {
        let repo = Repo::new(&base, &[change]);
        let r = list(&repo);
        assert_eq!(
            labels(&r),
            ["diff_check", "fmt", "clippy", "test"],
            "{what}: {r}"
        );
    }
}
