//! SPEC-20261125-check-fast-and-fix: `check fast <base>..<head>` runs the light checks and the
//! ledger tests, every one of them, and answers with every failure at once. Temporary git
//! repositories and a fake `cargo` script only: the real cargo, the real HOME and the machine's
//! heavy slots are never used.
mod common;
use common::git::{git, git_ok};
use moai_cli::check_fast::{self, Entry, Options};
use serde_json::Value;
use std::fs;
use std::path::{Path, PathBuf};

fn put(root: &Path, rel: &str, text: &[u8]) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}

/// A slot folder whose limits file allows four heavy runs and asks for no free disk.
fn slot_dir() -> tempfile::TempDir {
    let d = tempfile::tempdir().unwrap();
    fs::write(
        d.path().join("limits.json"),
        r#"{"max":4,"min_free_gb":0,"max_swap_pct":100,"min_mem_free_pct":0}"#,
    )
    .unwrap();
    d
}

/// A fake cargo: it logs its arguments beside itself, answers the nextest probe, accepts the format
/// check, and, when `fail_ledger` is set, fails a nextest run with one failure line on stderr, the
/// way nextest prints it.
fn fake_cargo(dir: &Path, fail_ledger: bool) -> PathBuf {
    let path = dir.join("fake-cargo");
    let script = "#!/bin/sh\n\
        here=$(dirname \"$0\")\n\
        echo \"$*\" >> \"$here/calls.log\"\n\
        case \"$1 $2\" in\n\
          \"nextest --version\") echo 'cargo-nextest 0.9.0'; exit 0 ;;\n\
          \"nextest run\")\n\
            if [ -e \"$here/fail-ledger\" ]; then\n\
              echo '        FAIL [   0.1s] repo_scans_table some_test' >&2\n\
              exit 1\n\
            fi\n\
            exit 0 ;;\n\
        esac\n\
        exit 0\n";
    fs::write(&path, script).unwrap();
    if fail_ledger {
        fs::write(dir.join("fail-ledger"), "").unwrap();
    }
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
    }
    path
}

/// A repository whose base commit holds every file of the ledger list (the files and functions the
/// list names, read from the list itself) and whose second commit changes `changed`.
struct Fixture {
    dir: tempfile::TempDir,
    base: String,
    head: String,
}

impl Fixture {
    fn new(changed: &[(&str, &[u8])]) -> Fixture {
        let dir = tempfile::tempdir().unwrap();
        let p = dir.path();
        git_ok(p, &["init", "-q", "-b", "main"]);
        git_ok(p, &["config", "user.email", "t@example.invalid"]);
        git_ok(p, &["config", "user.name", "t"]);
        put(p, "Cargo.toml", b"[package]\nname = \"fixture\"\n");
        put(p, "README.md", b"fixture\n");
        put(
            p,
            "tests/areas.toml",
            b"[areas]\ncard = [\"card\"]\nspec = [\"check\", \"repo\"]\n\n[solo]\nfiles = []\n",
        );
        let mut lib = String::new();
        for e in check_fast::LEDGER {
            match e {
                Entry::Target(name) => put(p, &format!("tests/{name}.rs"), b"fn main() {}\n"),
                Entry::Tests(name, fns) => {
                    let text: String = fns.iter().map(|f| format!("fn {f}() {{}}\n")).collect();
                    put(p, &format!("tests/{name}.rs"), text.as_bytes());
                }
                Entry::Lib(path) => {
                    lib.push_str(&format!(
                        "fn {}() {{}}\n",
                        path.rsplit("::").next().unwrap()
                    ));
                }
            }
        }
        put(p, "src/check_affected.rs", lib.as_bytes());
        git_ok(p, &["add", "-A"]);
        git_ok(p, &["commit", "-q", "-m", "base"]);
        let base = git(p, &["rev-parse", "HEAD"]);
        for (rel, text) in changed {
            put(p, rel, text);
        }
        git_ok(p, &["add", "-A"]);
        git_ok(p, &["commit", "-q", "-m", "change"]);
        let head = git(p, &["rev-parse", "HEAD"]);
        Fixture { dir, base, head }
    }

    fn root(&self) -> &Path {
        self.dir.path()
    }

    fn range(&self) -> String {
        format!("{}..{}", self.base, self.head)
    }
}

/// The options of a run that uses the fake cargo, the private slot folder and the given log folder.
fn options(tools: &Path, slots: &Path, out: &Path, fail_ledger: bool) -> Options {
    Options {
        slot_dir: Some(slots.to_path_buf()),
        max: 4,
        min_free_gb: 0,
        out: Some(out.to_path_buf()),
        integration_ref: None,
        cargo: Some(fake_cargo(tools, fail_ledger)),
    }
}

fn labels(v: &Value, key: &str) -> Vec<String> {
    v[key]
        .as_array()
        .unwrap()
        .iter()
        .map(|s| s["label"].as_str().unwrap().to_string())
        .collect()
}

/// AC-1: one failure planted in each of the four kinds of step; the single answer holds all four and
/// the run does not stop at the first. The fake cargo is never asked for clippy.
#[test]
fn check_fast_collects_every_failure() {
    let fx = Fixture::new(&[
        ("docs/spacey.md", b"a line with a trailing space \n"),
        ("docs/no_end.md", b"the last line has no newline"),
        ("tests/zzzz_new.rs", b"fn main() {}\n"),
    ]);
    let tools = tempfile::tempdir().unwrap();
    let slots = slot_dir();
    let out = tempfile::tempdir().unwrap();
    let opts = options(tools.path(), slots.path(), out.path(), true);
    let answer = check_fast::run(fx.root(), &fx.range(), &opts).expect("the run answers");
    assert_eq!(answer["action"], "fast");
    assert_eq!(answer["result"], "FAIL", "{answer}");
    assert_eq!(answer["writes"], true);
    assert_eq!(
        labels(&answer, "steps"),
        ["diff_check", "fmt", "test_tables", "endline", "ledger"],
        "every step ran, in plan order"
    );
    assert_eq!(
        labels(&answer, "failures"),
        ["diff_check", "test_tables", "endline", "ledger"],
        "{answer}"
    );
    let failures = answer["failures"].as_array().unwrap();
    let ledger = failures.iter().find(|f| f["label"] == "ledger").unwrap();
    assert_eq!(ledger["exit"], 1);
    assert_eq!(
        ledger["tests"],
        serde_json::json!(["repo_scans_table some_test"])
    );
    let endline = failures.iter().find(|f| f["label"] == "endline").unwrap();
    let log = fs::read_to_string(endline["log"].as_str().unwrap()).unwrap();
    assert!(
        log.contains("docs/no_end.md: no newline at end of file; add one"),
        "{log}"
    );
    assert!(!log.contains("docs/spacey.md"), "{log}");
    let summary = answer["summary"].as_str().unwrap();
    assert!(summary.starts_with("check-fast "), "{summary}");
    assert!(summary.ends_with("result=FAIL"), "{summary}");
    assert!(summary.contains(" fmt=0 ") && summary.contains(" ledger=1 "));
    let calls = fs::read_to_string(tools.path().join("calls.log")).unwrap();
    assert!(!calls.contains("clippy"), "{calls}");
    assert!(calls.contains("nextest run --locked --offline --no-fail-fast"));
    assert!(calls.contains("fmt --all -- --check"));
}

/// A range with nothing wrong passes, so the failures above are the planted ones and not standing
/// ones, and the answer has no failure to list.
#[test]
fn check_fast_passes_a_clean_range() {
    let fx = Fixture::new(&[("docs/ok.md", b"fine\n")]);
    let tools = tempfile::tempdir().unwrap();
    let slots = slot_dir();
    let out = tempfile::tempdir().unwrap();
    let opts = options(tools.path(), slots.path(), out.path(), false);
    let answer = check_fast::run(fx.root(), &fx.range(), &opts).expect("the run answers");
    assert_eq!(answer["result"], "PASS", "{answer}");
    assert!(
        answer["failures"].as_array().unwrap().is_empty(),
        "{answer}"
    );
    let summary = answer["summary"].as_str().unwrap();
    assert!(summary.ends_with("result=PASS"), "{summary}");
}

/// The run refuses a checkout that is not at the range head and one with a modified tracked file,
/// and starts nothing.
#[test]
fn check_fast_refuses_a_moved_or_dirty_checkout() {
    let fx = Fixture::new(&[("docs/ok.md", b"fine\n")]);
    let tools = tempfile::tempdir().unwrap();
    let opts = Options {
        out: Some(tools.path().join("out")),
        cargo: Some(fake_cargo(tools.path(), false)),
        ..Options::default()
    };
    put(fx.root(), "README.md", b"changed\n");
    let err = check_fast::run(fx.root(), &fx.range(), &opts).unwrap_err();
    assert!(err.contains("a tracked file is modified"), "{err}");
    git_ok(fx.root(), &["checkout", "-q", "--", "README.md"]);
    git_ok(fx.root(), &["checkout", "-q", &fx.base]);
    let err = check_fast::run(fx.root(), &fx.range(), &opts).unwrap_err();
    assert!(err.contains("the checkout is at"), "{err}");
    assert!(
        !tools.path().join("calls.log").exists(),
        "no program was started"
    );
}

/// AC-2: a file without a final newline is named; an empty file, a file with a NUL byte, a file
/// under `templates/bundle-packs/`, a link, a file that is gone and a file that ends in a newline
/// (or a CRLF) are not; a file that cannot be read is named as a failure of the check.
#[test]
fn the_endline_check_flags_a_missing_final_newline() {
    let dir = tempfile::tempdir().unwrap();
    let root = dir.path();
    put(root, "a.md", b"no newline");
    put(root, "ok.md", b"fine\n");
    put(root, "crlf.md", b"fine\r\n");
    put(root, "empty.md", b"");
    put(root, "bin.dat", b"ab\0cd");
    put(root, "templates/bundle-packs/p.md", b"no newline");
    #[cfg(unix)]
    std::os::unix::fs::symlink("a.md", root.join("link.md")).unwrap();
    // Where no link can be made, `link.md` is simply a path that is gone, which is skipped too.
    let present: Vec<String> = [
        "a.md",
        "ok.md",
        "crlf.md",
        "empty.md",
        "bin.dat",
        "templates/bundle-packs/p.md",
        "link.md",
        "gone.md",
    ]
    .iter()
    .map(|s| s.to_string())
    .collect();
    let lines = check_fast::endline_violations(root, &present);
    assert_eq!(lines, ["a.md: no newline at end of file; add one"]);
    assert!(check_fast::endline_violations(root, &[]).is_empty());

    // A file that cannot be read is a violation that names it (skipped where the process can read
    // it anyway, for example as the superuser).
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        put(root, "locked.md", b"x\n");
        fs::set_permissions(root.join("locked.md"), fs::Permissions::from_mode(0o000)).unwrap();
        if fs::read(root.join("locked.md")).is_err() {
            let lines = check_fast::endline_violations(root, &["locked.md".to_string()]);
            assert_eq!(lines.len(), 1, "{lines:?}");
            assert!(
                lines[0].starts_with("locked.md: cannot be read"),
                "{lines:?}"
            );
        }
        fs::set_permissions(root.join("locked.md"), fs::Permissions::from_mode(0o644)).unwrap();
    }
}
