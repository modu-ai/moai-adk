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
                "#!/bin/sh\necho \"{n} $*\" >> \"$FAKE_CALLS\"\necho \"{n} env CTD=$CARGO_TARGET_DIR MB=$MOAI_CLI_BIN\" >> \"$FAKE_CALLS\"\ncase \"$FAKE_FAIL\" in\n  '') ;;\n  *) case \" $* \" in *\" $FAKE_FAIL \"*) exit 3 ;; esac ;;\nesac\nexit 0\n"
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

/// Commit `change` on top of the fixture and return the fixture with the new head.
fn then(mut fx: Fixture, change: impl Fn(&Path)) -> Fixture {
    change(fx.root());
    git_ok(fx.root(), &["add", "-A"]);
    git_ok(fx.root(), &["commit", "-q", "-m", "more"]);
    fx.head = git(fx.root(), &["rev-parse", "HEAD"]).trim().to_string();
    fx
}

fn list_rust(fx: &Fixture, dir: &Path) -> Vec<String> {
    let out = Command::new(BIN)
        .args(["check", "affected", &fx.range(), "--list", "--dir"])
        .arg(dir)
        .env_remove("MOAI_HOME")
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], true, "{v}");
    names(&v["result"]["rust_targets"])
}

#[test]
fn a_renamed_file_counts_its_old_path_too() {
    // Git's default rename detection would print only `src/zzz.rs`; the old area's tests must
    // still be selected, whatever the user's `diff.renames` says.
    let fx = then(Fixture::new(false, &[("README.md", "y\n")]), |p| {
        git_ok(p, &["mv", "src/update_models.rs", "src/zzz.rs"]);
    });
    let rust = list_rust(&fx, fx.root());
    assert!(rust.contains(&"update_command_cli".to_string()), "{rust:?}");
}

#[test]
fn a_non_ascii_path_is_read_whatever_the_quote_setting() {
    let fx = Fixture::new(false, &[("README.md", "y\n")]);
    git_ok(fx.root(), &["config", "core.quotepath", "true"]);
    let fx = then(fx, |p| put(p, "src/update_한글.rs", "// a\n"));
    let rust = list_rust(&fx, fx.root());
    assert!(rust.contains(&"update_command_cli".to_string()), "{rust:?}");
}

#[test]
fn a_folder_inside_the_repository_reads_the_repository_root() {
    let fx = Fixture::new(false, &[("src/update_models.rs", "// b\n")]);
    let rust = list_rust(&fx, &fx.root().join("src"));
    assert!(rust.contains(&"update_command_cli".to_string()), "{rust:?}");
}

#[test]
fn the_python_step_uses_the_target_folder_the_build_step_builds_into() {
    let fx = Fixture::new(false, &[("src/update_models.rs", "// b\n")]);
    let tools = Tools::new(&["cargo", "python3", "rustfmt"]);
    let slots = slot_dir();
    let theirs = tempfile::tempdir().unwrap();
    let out = Command::new(BIN)
        .args(["check", "affected", &fx.range(), "--dir"])
        .arg(fx.root())
        .args(["--slot-dir"])
        .arg(slots.path())
        .env_remove("MOAI_HOME")
        .env("CARGO_TARGET_DIR", theirs.path())
        .env("PATH", tools.path())
        .env("FAKE_CALLS", tools.calls())
        .env("FAKE_FAIL", "")
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], true, "{v}");
    let calls = tools.calls_text();
    let want = format!(
        "python3 env CTD={0} MB={0}/debug/moai-cli",
        theirs.path().display()
    );
    assert!(calls.contains(&want), "{want} in {calls}");
    assert!(
        !fx.root().join("target").exists(),
        "logs and builds go to the caller's target folder"
    );
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
    // A heavy step fails (the cheap steps all pass): the rest still run (REQ-4).
    let (ok, r) = affected(&fx, &args, Some((&tools, "clippy")));
    assert!(ok, "{r}");
    let line = r["summary"].as_str().unwrap();
    assert!(line.starts_with("check-affected base="), "{line}");
    assert!(line.contains(" clippy=3 "), "{line}");
    assert!(line.ends_with(" result=FAIL"), "{line}");
    assert!(!line.contains("stopped_at"), "{line}");
    // The keys come in the order the steps run, then the result. The binary is built by the
    // test step, so no `build_cli` step is planned when a Rust target is selected (REQ-5).
    let keys: Vec<&str> = line
        .split(' ')
        .skip(1)
        .filter_map(|w| w.split_once('=').map(|(k, _)| k))
        .collect();
    assert_eq!(
        keys,
        [
            "base",
            "head",
            "targets",
            "py",
            "scans",
            "diff_check",
            "fmt",
            "clippy",
            "test_lib",
            "test_bins",
            "test_targets",
            "python",
            "result"
        ],
        "{line}"
    );
    assert_eq!(r["result"], "FAIL");
    // The steps after the failed one still ran.
    let calls = tools.calls_text();
    for later in [
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
    // A step key never carries 3 (a commit id can start with a 3, so look at whole words).
    assert!(
        !line
            .split(' ')
            .any(|w| w.split_once('=').is_some_and(|(_, v)| v == "3")),
        "{line}"
    );
    assert_eq!(r["steps"].as_array().unwrap().len(), labels(&r).len());
}

/// The keys of a summary line, in order.
fn keys(line: &str) -> Vec<&str> {
    line.split(' ')
        .skip(1)
        .filter_map(|w| w.split_once('=').map(|(k, _)| k))
        .collect()
}

/// SPEC-20261122-check-light-stop REQ-1 to REQ-3: a cheap step that fails stops the run before
/// clippy, the build, the tests and the Python step; the other cheap steps still run; the
/// summary names the first failed one right before the result and has no word for a step that
/// did not run.
#[test]
fn a_failed_cheap_step_stops_before_the_heavy_steps() {
    let fx = Fixture::new(
        true,
        &[
            ("src/update_models.rs", "// b\n"),
            ("src/ops/alpha.rs", "// b\n"),
        ],
    );
    let tools = Tools::new(&["cargo", "python3", "rustfmt"]);
    let slots = slot_dir();
    let slot = slots.path().to_string_lossy().to_string();
    let (ok, r) = affected(&fx, &["--slot-dir", slot.as_str()], Some((&tools, "fmt")));
    assert!(ok, "{r}");
    let line = r["summary"].as_str().unwrap();
    assert!(line.contains(" fmt=3 "), "{line}");
    assert!(line.ends_with(" stopped_at=fmt result=FAIL"), "{line}");
    assert_eq!(
        keys(line),
        [
            "base",
            "head",
            "targets",
            "py",
            "scans",
            "diff_check",
            "fmt",
            "ops_rustfmt",
            "stopped_at",
            "result"
        ],
        "{line}"
    );
    assert_eq!(r["result"], "FAIL");
    // The other cheap steps ran; nothing heavy and no Python did.
    let calls = tools.calls_text();
    assert!(calls.contains("rustfmt --edition 2021 --check"), "{calls}");
    for never in ["clippy", "--lib", "--bins", "--test ", "python3"] {
        assert!(!calls.contains(never), "{never} ran: {calls}");
    }
    for line in calls.lines().filter(|l| l.starts_with("cargo ")) {
        assert!(
            line.starts_with("cargo fmt ") || line.starts_with("cargo env "),
            "a heavy cargo step ran: {line}"
        );
    }
    let ran: Vec<String> = r["steps"]
        .as_array()
        .unwrap()
        .iter()
        .map(|s| s["label"].as_str().unwrap().to_string())
        .collect();
    assert_eq!(ran, ["diff_check", "fmt", "ops_rustfmt"], "{r}");
}

/// REQ-1 with the three kinds of cheap step: a real trailing-space `git diff --check` failure,
/// an operation-file format check whose program cannot start (127) and a real `bash -n` failure.
#[test]
fn every_kind_of_cheap_failure_stops_the_run() {
    let slots = slot_dir();
    let slot = slots.path().to_string_lossy().to_string();
    let cases: [(&[(&str, &str)], &str, &[&str]); 3] = [
        (
            &[("src/update_models.rs", "// trailing \n")],
            "diff_check",
            &["cargo", "python3", "rustfmt"],
        ),
        (
            &[("src/ops/alpha.rs", "// b\n")],
            "ops_rustfmt",
            // No fake `rustfmt` on the path: the step cannot start.
            &["cargo", "python3"],
        ),
        (
            &[("tools/a.sh", "if true; then\n")],
            "bash_n:tools/a.sh",
            &["cargo", "python3", "rustfmt"],
        ),
    ];
    for (changed, stopped, fakes) in cases {
        let fx = Fixture::new(false, changed);
        let tools = Tools::new(fakes);
        let (ok, r) = affected(&fx, &["--slot-dir", slot.as_str()], Some((&tools, "")));
        assert!(ok, "{stopped}: {r}");
        let line = r["summary"].as_str().unwrap();
        // The summary key of a repeated label stops at its `:`.
        let key = stopped.split(':').next().unwrap();
        assert!(
            line.ends_with(&format!(" stopped_at={key} result=FAIL")),
            "{stopped}: {line}"
        );
        let calls = tools.calls_text();
        for never in ["clippy", "--lib", "--bins", "python3"] {
            assert!(!calls.contains(never), "{stopped}: {never} ran: {calls}");
        }
    }
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
