//! SPEC-20261120-check-affected: `check affected <base>..<head>` picks the tests a change can
//! break and runs them. Temporary git repositories and fake `cargo`, `python3` and `rustfmt`
//! scripts only: the real tools, the real HOME and the machine's heavy slots are never used.
mod common;
use common::git::{git, git_ok};
use serde_json::Value;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

const BIN: &str = crate::common::MOAI_BIN;

const SCANS: &str = r#"{"about":"fixture","scans":[
  {"id":"home","checks":["cargo test --offline --locked --test home_isolation_scan"]},
  {"id":"lib","checks":["cargo test --offline --locked --lib state_dir"]}]}"#;

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
    /// A repository with the tests of the integration-day misses and one commit on top of
    /// the base that changes `changed`.
    fn new(scans: bool, changed: &[(&str, &str)]) -> Fixture {
        let dir = tempfile::tempdir().unwrap();
        let p = dir.path();
        git_ok(p, &["init", "-q", "-b", "main"]);
        git_ok(p, &["config", "user.email", "t@example.invalid"]);
        git_ok(p, &["config", "user.name", "t"]);
        put(p, "README.md", "x\n");
        put(p, "src/update_models.rs", "// a\n");
        put(p, "src/ops/alpha.rs", "// a\n");
        put(p, "tests/update_command_cli.rs", "fn main() {}\n");
        put(p, "tests/bundle_body_cli.rs", "mod common;\nfn main() {}\n");
        put(p, "tests/card_cli.rs", "mod common;\nfn main() {}\n");
        put(p, "tests/solo_cli.rs", "fn main() {}\n");
        put(p, "tests/home_isolation_scan.rs", "fn main() {}\n");
        put(p, "tests/help_layout.rs", "fn main() {}\n");
        put(p, "tests/update_tux_tty_test.py", "#\n");
        put(p, "tests/bundle_ops_test.py", "#\n");
        put(p, "tests/common/mod.rs", "// helper\n");
        put(p, "tests/fixtures/help/all.txt", "a\n");
        if scans {
            put(p, "tools/repo_scans.json", SCANS);
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

/// A folder of fake tools. Every fake appends `<tool> <args>` to `calls.log`; one named in
/// `FAKE_FAIL` (a word of the arguments) exits 3. `git` is the real one.
struct Tools {
    dir: tempfile::TempDir,
}

impl Tools {
    fn new(names: &[&str]) -> Tools {
        let dir = tempfile::tempdir().unwrap();
        for n in names {
            let script = format!(
                "#!/bin/sh\necho \"{n} $*\" >> \"$FAKE_CALLS\"\ncase \"$FAKE_FAIL\" in\n  '') ;;\n  *) case \" $* \" in *\" $FAKE_FAIL \"*) exit 3 ;; esac ;;\nesac\nexit 0\n"
            );
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

    fn calls(&self) -> PathBuf {
        self.dir.path().join("calls.log")
    }

    fn calls_text(&self) -> String {
        fs::read_to_string(self.calls()).unwrap_or_default()
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
    slot_dir_with(4)
}

/// A slot folder whose limits file allows `max` heavy runs; 0 makes every acquire refuse (125).
fn slot_dir_with(max: u32) -> tempfile::TempDir {
    let d = tempfile::tempdir().unwrap();
    fs::write(
        d.path().join("limits.json"),
        format!(r#"{{"max":{max},"min_free_gb":0,"max_swap_pct":100,"min_mem_free_pct":0}}"#),
    )
    .unwrap();
    d
}

/// The whole answer: `(ok, result or error)`.
fn affected(fx: &Fixture, extra: &[&str], tools: Option<(&Tools, &str)>) -> (bool, Value) {
    let mut cmd = Command::new(BIN);
    cmd.args(["check", "affected", &fx.range(), "--dir"])
        .arg(fx.root())
        .args(extra)
        .env_remove("MOAI_HOME")
        .stdin(Stdio::null());
    if let Some((t, fail)) = tools {
        cmd.env("PATH", t.path())
            .env("FAKE_CALLS", t.calls())
            .env("FAKE_FAIL", fail);
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

fn names(v: &Value) -> Vec<String> {
    v.as_array()
        .unwrap()
        .iter()
        .map(|s| s.as_str().unwrap().to_string())
        .collect()
}

fn labels(v: &Value) -> Vec<String> {
    v["steps"]
        .as_array()
        .unwrap()
        .iter()
        .map(|s| s["label"].as_str().unwrap().to_string())
        .collect()
}

#[test]
fn the_misses_of_the_integration_day_are_selected() {
    let fx = Fixture::new(true, &[("src/update_models.rs", "// b\n")]);
    let (ok, r) = affected(&fx, &["--list"], None);
    assert!(ok, "{r}");
    let rust = names(&r["rust_targets"]);
    let python = names(&r["python_modules"]);
    for want in ["update_command_cli", "home_isolation_scan"] {
        assert!(rust.contains(&want.to_string()), "{want} in {rust:?}");
    }
    assert!(
        python.contains(&"update_tux_tty_test".to_string()),
        "{python:?}"
    );
    assert!(!rust.contains(&"bundle_body_cli".to_string()), "{rust:?}");
    assert!(
        !python.contains(&"bundle_ops_test".to_string()),
        "{python:?}"
    );
    assert_eq!(r["scans"], "listed:1");
}

#[test]
fn areas_changed_tests_fixtures_and_the_helper_select_their_targets() {
    let fx = Fixture::new(
        false,
        &[
            ("tests/solo_cli.rs", "fn main() { /* b */ }\n"),
            ("tests/fixtures/help/all.txt", "b\n"),
            ("tests/common/mod.rs", "// helper b\n"),
        ],
    );
    let (ok, r) = affected(&fx, &["--list"], None);
    assert!(ok, "{r}");
    let rust = names(&r["rust_targets"]);
    for want in ["solo_cli", "help_layout", "bundle_body_cli", "card_cli"] {
        assert!(rust.contains(&want.to_string()), "{want} in {rust:?}");
    }
    assert!(
        !rust.contains(&"update_command_cli".to_string()),
        "{rust:?}"
    );
}

#[test]
fn the_fixed_steps_and_the_host_targets_are_always_listed() {
    let host = std::env::consts::OS;
    let fx = Fixture::new(
        false,
        &[
            ("README.md", "y\n"),
            ("src/ops/alpha.rs", "// b\n"),
            (
                "tests/solo_cli.rs",
                &format!("#[cfg(target_os = \"{host}\")]\nfn main() {{}}\n"),
            ),
            ("tools/a.sh", "#!/bin/sh\n"),
        ],
    );
    let (ok, r) = affected(&fx, &["--list"], None);
    assert!(ok, "{r}");
    let l = labels(&r);
    for want in [
        "diff_check",
        "fmt",
        "ops_rustfmt",
        "bash_n:tools/a.sh",
        "clippy",
        "test_lib",
        "test_bins",
        "test_targets",
    ] {
        assert!(l.contains(&want.to_string()), "{want} in {l:?}");
    }
    let ops = r["steps"]
        .as_array()
        .unwrap()
        .iter()
        .find(|s| s["label"] == "ops_rustfmt")
        .unwrap();
    assert_eq!(
        ops["command"],
        serde_json::json!([
            "rustfmt",
            "--edition",
            "2021",
            "--check",
            "src/ops/alpha.rs"
        ])
    );
    assert!(names(&r["rust_targets"]).contains(&"solo_cli".to_string()));
    assert_eq!(r["host_os"], host);
}

#[test]
fn a_head_behind_its_base_tip_counts_only_its_own_changes() {
    // The head's history left the base before the tip changed `src/card.rs`: the changed paths
    // are those since the merge base (three dots), so the tip's own change selects nothing.
    let fx = Fixture::new(false, &[("src/update_models.rs", "// b\n")]);
    git_ok(fx.root(), &["checkout", "-q", "-b", "tip", &fx.base]);
    put(fx.root(), "src/card.rs", "// the tip's own change\n");
    git_ok(fx.root(), &["add", "-A"]);
    git_ok(fx.root(), &["commit", "-q", "-m", "tip"]);
    let tip = git(fx.root(), &["rev-parse", "HEAD"]).trim().to_string();
    git_ok(fx.root(), &["checkout", "-q", "main"]);
    let mut cmd = Command::new(BIN);
    let out = cmd
        .args([
            "check",
            "affected",
            &format!("{tip}..{}", fx.head),
            "--list",
            "--dir",
        ])
        .arg(fx.root())
        .env_remove("MOAI_HOME")
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], true, "{v}");
    let rust = names(&v["result"]["rust_targets"]);
    assert!(rust.contains(&"update_command_cli".to_string()), "{rust:?}");
    assert!(!rust.contains(&"card_cli".to_string()), "{rust:?}");
    assert_eq!(v["result"]["changed"], 1, "{v}");
}

#[test]
fn a_project_without_a_scan_list_says_scans_none() {
    let fx = Fixture::new(false, &[("src/update_models.rs", "// b\n")]);
    let (ok, r) = affected(&fx, &["--list"], None);
    assert!(ok, "{r}");
    assert_eq!(r["scans"], "none");
    assert!(!names(&r["rust_targets"]).contains(&"home_isolation_scan".to_string()));
}

#[test]
fn list_reads_only_and_shows_the_selection() {
    let fx = Fixture::new(true, &[("src/update_models.rs", "// b\n")]);
    let before = (
        git(fx.root(), &["status", "--porcelain", "--ignored"]),
        git(fx.root(), &["for-each-ref"]),
    );
    let (ok, r) = affected(&fx, &["--list"], None);
    assert!(ok, "{r}");
    assert_eq!(r["writes"], false);
    let after = (
        git(fx.root(), &["status", "--porcelain", "--ignored"]),
        git(fx.root(), &["for-each-ref"]),
    );
    assert_eq!(before, after);
    assert!(!fx.root().join("target").exists());
}

#[test]
fn a_failed_step_does_not_stop_the_rest_and_the_summary_says_fail() {
    let fx = Fixture::new(true, &[("src/update_models.rs", "// b\n")]);
    let tools = Tools::new(&["cargo", "python3", "rustfmt"]);
    let slots = slot_dir();
    let slot = slots.path().to_string_lossy().to_string();
    let args = ["--slot-dir", slot.as_str()];
    let (ok, r) = affected(&fx, &args, Some((&tools, "fmt")));
    assert!(ok, "{r}");
    let line = r["summary"].as_str().unwrap();
    assert!(line.starts_with("check-affected base="), "{line}");
    assert!(line.contains(" fmt=3 "), "{line}");
    assert!(line.ends_with(" result=FAIL"), "{line}");
    assert_eq!(r["result"], "FAIL");
    // The steps after the failed one still ran.
    let calls = tools.calls_text();
    for later in [
        "clippy",
        "test --locked --offline --no-fail-fast --lib",
        "--bins",
        "--test update_command_cli",
    ] {
        assert!(calls.contains(later), "{later} in {calls}");
    }
    assert!(
        calls.contains("python3 -m unittest tests.update_tux_tty_test"),
        "{calls}"
    );

    // Every step passing gives PASS and exit 0 for each key.
    let tools = Tools::new(&["cargo", "python3", "rustfmt"]);
    let (ok, r) = affected(&fx, &args, Some((&tools, "")));
    assert!(ok, "{r}");
    let line = r["summary"].as_str().unwrap();
    assert!(line.ends_with(" result=PASS"), "{line}");
    assert!(!line.contains("=3"), "{line}");
    assert_eq!(r["steps"].as_array().unwrap().len(), labels(&r).len());
}

#[test]
fn a_program_that_cannot_start_is_127_and_a_refused_slot_keeps_its_code() {
    let fx = Fixture::new(false, &[("src/ops/alpha.rs", "// b\n")]);
    // No fake `rustfmt` on the path: the step cannot start.
    let tools = Tools::new(&["cargo"]);
    let slots = slot_dir();
    let slot = slots.path().to_string_lossy().to_string();
    let (ok, r) = affected(&fx, &["--slot-dir", slot.as_str()], Some((&tools, "")));
    assert!(ok, "{r}");
    let line = r["summary"].as_str().unwrap();
    assert!(line.contains(" ops_rustfmt=127 "), "{line}");
    assert!(line.ends_with(" result=FAIL"), "{line}");

    // A limits file with `max` 0 makes the heavy slot refuse (125): no cargo step runs.
    let tools = Tools::new(&["cargo", "rustfmt"]);
    let refusing = slot_dir_with(0);
    let refusing = refusing.path().to_string_lossy().to_string();
    let (ok, r) = affected(&fx, &["--slot-dir", refusing.as_str()], Some((&tools, "")));
    assert!(ok, "{r}");
    let line = r["summary"].as_str().unwrap();
    assert!(line.contains(" clippy=125 "), "{line}");
    assert!(line.ends_with(" result=FAIL"), "{line}");
    let calls = tools.calls_text();
    assert!(
        !calls.contains("clippy") && !calls.contains("--lib"),
        "{calls}"
    );
}

#[test]
fn a_checkout_that_is_not_the_head_runs_nothing() {
    let fx = Fixture::new(false, &[("src/update_models.rs", "// b\n")]);
    let tools = Tools::new(&["cargo", "python3", "rustfmt"]);
    let slots = slot_dir();
    let slot = slots.path().to_string_lossy().to_string();
    let args = ["--slot-dir", slot.as_str()];

    // A modified tracked file.
    put(fx.root(), "README.md", "dirty\n");
    let (ok, e) = affected(&fx, &args, Some((&tools, "")));
    assert!(
        !ok && e["message"]
            .as_str()
            .unwrap()
            .contains("a tracked file is modified"),
        "{e}"
    );
    git_ok(fx.root(), &["checkout", "-q", "--", "README.md"]);

    // A checkout that is not the range's head.
    git_ok(fx.root(), &["checkout", "-q", "--detach", &fx.base]);
    let (ok, e) = affected(&fx, &args, Some((&tools, "")));
    assert!(
        !ok && e["message"]
            .as_str()
            .unwrap()
            .contains("check out the head first"),
        "{e}"
    );

    // A base that cannot be read.
    git_ok(fx.root(), &["checkout", "-q", "main"]);
    let mut cmd = Command::new(BIN);
    let out = cmd
        .args(["check", "affected", "no-such-ref..HEAD", "--dir"])
        .arg(fx.root())
        .env("PATH", tools.path())
        .env("FAKE_CALLS", tools.calls())
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], false, "{v}");
    assert!(
        v["error"]["message"]
            .as_str()
            .unwrap()
            .contains("cannot be read"),
        "{v}"
    );
    assert_eq!(tools.calls_text(), "", "no step may have run");
}
