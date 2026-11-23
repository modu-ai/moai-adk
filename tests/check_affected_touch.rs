//! SPEC-20261124-check-affected-touch (AC-1 to AC-5): `check affected` sets the modification time
//! of every changed file of the range to now before its first step, so a file restored with an old
//! time (`cp -p`) is rebuilt; it never follows a link out of the checkout and `--list` touches
//! nothing; the guard-mutant guidance names the restore rule. Temporary git repositories and a
//! temporary HOME; the real-cargo test builds one tiny crate in its own folder with the current
//! toolchain's own binaries, and the others use a fake `cargo`.
#![cfg(unix)]
mod common;

use common::git::{git, git_ok};
use serde_json::Value;
use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use std::time::{Duration, SystemTime, UNIX_EPOCH};

const BIN: &str = crate::common::MOAI_BIN;

/// Roomy machine numbers (debug-only seam), so the test machine's real disk and memory never
/// decide the outcome.
const ROOMY: &str = "free_gb=1000,swap_pct=0";

/// A file time long before any build of this test (2001-09-09).
fn old() -> SystemTime {
    UNIX_EPOCH + Duration::from_secs(1_000_000_000)
}

fn put(root: &Path, rel: &str, text: &str) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}

/// The way `cp -p` leaves a file: the time is set to `t`, the bytes stay. Follows a link.
fn set_time(path: &Path, t: SystemTime) {
    fs::File::open(path).unwrap().set_modified(t).unwrap();
}

fn mtime(path: &Path) -> SystemTime {
    fs::metadata(path).unwrap().modified().unwrap()
}

/// A link's own time, not its target's.
fn link_mtime(path: &Path) -> SystemTime {
    fs::symlink_metadata(path).unwrap().modified().unwrap()
}

fn git_init(p: &Path) {
    git_ok(p, &["init", "-q", "-b", "main"]);
    git_ok(p, &["config", "user.email", "t@example.invalid"]);
    git_ok(p, &["config", "user.name", "t"]);
}

fn commit_all(p: &Path, msg: &str) -> String {
    git_ok(p, &["add", "-A"]);
    git_ok(p, &["commit", "-q", "-m", msg]);
    git(p, &["rev-parse", "HEAD"])
}

/// The directory that holds `git`, for a PATH made of known folders only.
fn git_dir() -> String {
    let out = Command::new("sh")
        .args(["-c", "dirname \"$(command -v git)\""])
        .output()
        .unwrap();
    String::from_utf8_lossy(&out.stdout).trim().to_string()
}

/// A slot folder whose limits file allows four heavy runs.
fn slot_dir() -> tempfile::TempDir {
    let d = tempfile::tempdir().unwrap();
    fs::write(
        d.path().join("limits.json"),
        r#"{"max":4,"min_free_gb":0,"max_swap_pct":100,"min_mem_free_pct":0}"#,
    )
    .unwrap();
    d
}

/// A fake `cargo` that exits 0 for every call (a fake `nextest --version` answers 0 too).
fn fake_cargo() -> tempfile::TempDir {
    let d = tempfile::tempdir().unwrap();
    let path = d.path().join("cargo");
    fs::write(&path, "#!/bin/sh\nexit 0\n").unwrap();
    fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
    d
}

/// Everything one run needs besides the repository: the folders it writes to and a PATH.
struct Run {
    slots: tempfile::TempDir,
    out: tempfile::TempDir,
    home: tempfile::TempDir,
    path: String,
}

impl Run {
    /// `tools` first on PATH, then `git` and the system's shell tools.
    fn new(tools: &str) -> Run {
        Run {
            slots: slot_dir(),
            out: tempfile::tempdir().unwrap(),
            home: tempfile::tempdir().unwrap(),
            path: format!("{tools}:{}:/usr/bin:/bin", git_dir()),
        }
    }

    fn command(&self, root: &Path, range: &str, extra: &[&str]) -> Command {
        let mut cmd = Command::new(BIN);
        cmd.args(["check", "affected", range, "--dir"])
            .arg(root)
            .arg("--slot-dir")
            .arg(self.slots.path())
            .arg("--out")
            .arg(self.out.path())
            .args(extra)
            .env("HOME", self.home.path())
            .env("CARGO_HOME", self.home.path().join("cargo"))
            .env_remove("MOAI_HOME")
            .env_remove("XDG_CACHE_HOME")
            .env_remove("CARGO_TARGET_DIR")
            .env_remove("MOAI_HEAVY_SLOT_FD")
            .env_remove("RUSTC_WRAPPER")
            .env_remove("RUST_TEST_THREADS")
            .env_remove("NEXTEST_TEST_THREADS")
            .env("MOAI_TEST_MACHINE", ROOMY)
            .env("PATH", &self.path)
            .stdin(Stdio::null());
        cmd
    }
}

/// The `result` of a successful answer on stdout (anything before the JSON is skipped).
fn answer(out: &std::process::Output) -> Value {
    let text = String::from_utf8_lossy(&out.stdout);
    let at = text.find('{').unwrap_or(0);
    let v: Value = serde_json::from_str(&text[at..])
        .unwrap_or_else(|_| panic!("{text} / {}", String::from_utf8_lossy(&out.stderr)));
    assert_eq!(v["ok"], true, "{v}");
    v["result"].clone()
}

/// AC-2 (REQ-1): the time of a changed file moves to a time after the run started, its bytes and
/// mode stay, and a file outside the range keeps its time.
#[test]
fn the_run_refreshes_only_the_changed_files_time() {
    let dir = tempfile::tempdir().unwrap();
    let p = dir.path();
    git_init(p);
    put(p, "README.md", "x\n");
    put(p, "src/probe.rs", "// a\n");
    put(p, "docs/keep.md", "k\n");
    let base = commit_all(p, "base");
    put(p, "README.md", "y\n");
    put(p, "src/probe.rs", "// b\n");
    fs::set_permissions(p.join("src/probe.rs"), fs::Permissions::from_mode(0o640)).unwrap();
    let head = commit_all(p, "change");
    for f in ["README.md", "src/probe.rs", "docs/keep.md"] {
        set_time(&p.join(f), old());
    }

    let tools = fake_cargo();
    let run = Run::new(&tools.path().to_string_lossy());
    let start = SystemTime::now();
    let out = run
        .command(p, &format!("{base}..{head}"), &[])
        .output()
        .unwrap();
    let v = answer(&out);

    assert_eq!(v["result"], "PASS", "{v}");
    assert_eq!(v["refresh"]["touched"], 2, "{v}");
    assert_eq!(v["refresh"]["skipped"], serde_json::json!([]), "{v}");
    for f in ["README.md", "src/probe.rs"] {
        assert!(mtime(&p.join(f)) >= start, "{f} keeps its old time");
    }
    assert_eq!(mtime(&p.join("docs/keep.md")), old(), "a file out of range");
    assert_eq!(fs::read_to_string(p.join("README.md")).unwrap(), "y\n");
    assert_eq!(
        fs::read_to_string(p.join("src/probe.rs")).unwrap(),
        "// b\n"
    );
    let mode = fs::metadata(p.join("src/probe.rs")).unwrap().permissions();
    assert_eq!(mode.mode() & 0o777, 0o640);
}

/// AC-3 (REQ-2): a changed link, and a changed path that runs through a linked folder, are
/// skipped and named; the targets outside the checkout keep their time.
#[test]
fn a_changed_link_is_skipped_and_named() {
    let outside = tempfile::tempdir().unwrap();
    let target = outside.path().join("target.md");
    let folder = outside.path().join("dir");
    fs::write(&target, "t\n").unwrap();
    put(&folder, "b", "1\n");
    for f in [target.clone(), folder.join("b")] {
        set_time(&f, old());
    }

    let dir = tempfile::tempdir().unwrap();
    let p = dir.path();
    git_init(p);
    put(p, "README.md", "x\n");
    put(p, "a/b", "1\n");
    let base = commit_all(p, "base");
    // The folder `a` becomes a link to a folder outside; `a/b` is still reachable through it.
    fs::remove_dir_all(p.join("a")).unwrap();
    std::os::unix::fs::symlink(&folder, p.join("a")).unwrap();
    std::os::unix::fs::symlink(&target, p.join("link.md")).unwrap();
    put(p, "README.md", "y\n");
    let head = commit_all(p, "change");
    let link_before = link_mtime(&p.join("link.md"));
    set_time(&p.join("README.md"), old());

    let tools = fake_cargo();
    let run = Run::new(&tools.path().to_string_lossy());
    let out = run
        .command(p, &format!("{base}..{head}"), &[])
        .output()
        .unwrap();
    let v = answer(&out);

    assert_eq!(v["result"], "PASS", "{v}");
    assert_eq!(v["refresh"]["touched"], 1, "{v}");
    let skipped: Vec<&str> = v["refresh"]["skipped"]
        .as_array()
        .unwrap()
        .iter()
        .filter_map(Value::as_str)
        .collect();
    assert!(skipped.contains(&"link.md"), "{v}");
    assert!(skipped.contains(&"a/b"), "{v}");
    assert_eq!(mtime(&target), old(), "the target of a link");
    assert_eq!(
        mtime(&folder.join("b")),
        old(),
        "a file behind a linked folder"
    );
    assert_eq!(link_mtime(&p.join("link.md")), link_before);
    assert!(mtime(&p.join("README.md")) > old(), "the plain file moved");
}

/// AC-4 (REQ-3): `--list` reads and writes nothing, so every file keeps its time.
#[test]
fn list_mode_touches_no_file() {
    let dir = tempfile::tempdir().unwrap();
    let p = dir.path();
    git_init(p);
    put(p, "README.md", "x\n");
    let base = commit_all(p, "base");
    put(p, "README.md", "y\n");
    let head = commit_all(p, "change");
    set_time(&p.join("README.md"), old());

    let tools = fake_cargo();
    let run = Run::new(&tools.path().to_string_lossy());
    let out = run
        .command(p, &format!("{base}..{head}"), &["--list"])
        .output()
        .unwrap();
    answer(&out);

    assert_eq!(mtime(&p.join("README.md")), old());
}

/// AC-5 (REQ-4, REQ-5): the guard-mutant line of the self-check page names the restore rule.
#[test]
fn the_guard_mutant_line_names_the_restore_rule() {
    let page = include_str!("../skills/moai/references/internal/self-check.md");
    let line = page
        .lines()
        .find(|l| l.contains("guard-mutant:") && l.starts_with("2. Guards:"))
        .expect("the Guards line of the Tests list");
    for needle in [
        "`git checkout -- <path>`",
        "plain `cp`",
        "never `cp -p`",
        "file time",
    ] {
        assert!(
            line.contains(needle),
            "the Guards line lacks {needle}: {line}"
        );
    }
}

const MANIFEST: &str = "[package]\nname = \"probe\"\nversion = \"0.1.0\"\nedition = \"2021\"\n\n\
[lib]\npath = \"src/lib.rs\"\n\n[[bin]]\nname = \"moai-cli\"\npath = \"src/main.rs\"\n";

/// A library whose only test passes when `answer()` is 1.
fn lib_with(answer: u32) -> String {
    format!(
        "pub fn answer() -> u32 {{\n    {answer}\n}}\n\n#[cfg(test)]\nmod tests {{\n    #[test]\n    fn answer_is_one() {{\n        assert_eq!(super::answer(), 1);\n    }}\n}}\n"
    )
}

/// The toolchain's own binary folder, so the crate builds with the same cargo, clippy and rustfmt
/// as this test and no `rustup` proxy reads the user's home.
fn toolchain_bin() -> PathBuf {
    let out = Command::new("rustc")
        .args(["--print", "sysroot"])
        .output()
        .unwrap();
    assert!(out.status.success());
    PathBuf::from(String::from_utf8_lossy(&out.stdout).trim()).join("bin")
}

/// AC-1 (REQ-1): the bug itself. A crate is built with the good source, then with a failing
/// mutant, then the good bytes come back with the old file time (what `cp -p` does). Cargo alone
/// keeps the mutant build; `check affected` on the range rebuilds the file and its tests pass.
#[test]
fn a_restore_with_the_old_file_time_is_rebuilt() {
    let dir = tempfile::tempdir().unwrap();
    let p = dir.path();
    git_init(p);
    put(p, "Cargo.toml", MANIFEST);
    put(p, "src/main.rs", "fn main() {}\n");
    put(p, "src/lib.rs", &lib_with(0));
    put(p, ".gitignore", "/target\n");
    let bin = toolchain_bin();
    // Cargo's own state lives beside the repository, never inside it (a tree that must stay clean).
    let aux = tempfile::tempdir().unwrap();
    let cargo = |args: &[&str]| {
        let mut cmd = Command::new(bin.join("cargo"));
        cmd.args(args)
            .current_dir(p)
            .env("PATH", format!("{}:/usr/bin:/bin", bin.display()))
            .env("HOME", aux.path().join("home"))
            .env("CARGO_HOME", aux.path().join("cargo-home"))
            .env_remove("CARGO_TARGET_DIR")
            .env_remove("RUSTC_WRAPPER")
            .env_remove("RUSTFLAGS")
            .env_remove("CARGO_ENCODED_RUSTFLAGS")
            .stdin(Stdio::null());
        cmd
    };
    let lock = cargo(&["generate-lockfile", "--offline"]).output().unwrap();
    assert!(
        lock.status.success(),
        "{}",
        String::from_utf8_lossy(&lock.stderr)
    );
    let base = commit_all(p, "base");

    // Head: the good source, with a time long before every build below.
    put(p, "src/lib.rs", &lib_with(1));
    let head = commit_all(p, "change");
    let lib = p.join("src/lib.rs");
    set_time(&lib, old());
    let build = |what: &str| {
        let out = cargo(&["test", "--offline", "--locked", "--lib", "--no-run"])
            .output()
            .unwrap();
        assert!(
            out.status.success(),
            "{what}: {}",
            String::from_utf8_lossy(&out.stderr)
        );
    };
    build("good build");
    // The mutant is newer than that build, so cargo builds it.
    put(p, "src/lib.rs", &lib_with(2));
    build("mutant build");
    // Restore the good bytes with the old time.
    put(p, "src/lib.rs", &lib_with(1));
    set_time(&lib, old());
    assert_eq!(git(p, &["status", "--porcelain"]), "", "the tree is clean");

    // The precondition that makes the test mean something: cargo alone runs the mutant build.
    let stale = cargo(&["test", "--offline", "--locked", "--lib"])
        .output()
        .unwrap();
    assert!(
        !stale.status.success(),
        "cargo alone must keep the mutant build: {}",
        String::from_utf8_lossy(&stale.stdout)
    );

    let run = Run::new(&bin.to_string_lossy());
    let out = run
        .command(p, &format!("{base}..{head}"), &[])
        .output()
        .unwrap();
    let v = answer(&out);
    assert_eq!(v["result"], "PASS", "{v}");
    assert!(
        v["summary"].as_str().unwrap().ends_with(" result=PASS"),
        "{v}"
    );
}
