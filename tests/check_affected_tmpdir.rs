//! SPEC-20261124-home-guard-flake revision 4: `check affected` makes a private temporary folder
//! for its `test` step and passes it as `TMPDIR` to that step only. The folder lies outside the
//! worktree, inside no git repository and with no session id shape in its path; a passed run
//! removes it, a failed run keeps it, and the summary line gets no new word. Temporary git
//! repositories and a fake `cargo` that writes the `TMPDIR` it received into a probe file.
mod common;
use common::git::{git, git_ok};
use serde_json::Value;
use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

const BIN: &str = crate::common::MOAI_BIN;

fn put(root: &Path, rel: &str, text: &str) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}

/// A repository with an area table, one `card_*` target and a changed `src/card/zeta.rs`, so a
/// `--card` run plans a `test` step.
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
        put(
            p,
            "tests/areas.toml",
            "[areas]\ncard = [\"card\"]\n\n[solo]\nfiles = []\n",
        );
        put(p, "tests/card_a_cli.rs", "fn main() {}\n");
        put(p, "src/lib.rs", "// a\n");
        put(p, "src/card/zeta.rs", "// a\n");
        git_ok(p, &["add", "-A"]);
        git_ok(p, &["commit", "-q", "-m", "base"]);
        let base = git(p, &["rev-parse", "HEAD"]).trim().to_string();
        put(p, "src/card/zeta.rs", "// b\n");
        git_ok(p, &["add", "-A"]);
        git_ok(p, &["commit", "-q", "-m", "change"]);
        let head = git(p, &["rev-parse", "HEAD"]).trim().to_string();
        Fixture { dir, base, head }
    }

    fn root(&self) -> &Path {
        self.dir.path()
    }
}

/// A fake `cargo` (and `rustfmt`) in front of the system's `git`. Each call appends one line
/// `<arguments> | TMPDIR=<value or unset>` to `$PROBE_OUT`; a `nextest run` also writes a file
/// into its `TMPDIR` and exits with `$FAKE_TEST_EXIT` (0 when unset).
fn fake_path() -> (tempfile::TempDir, String) {
    let dir = tempfile::tempdir().unwrap();
    let cargo = "#!/bin/sh\n\
        echo \"$* | TMPDIR=${TMPDIR-unset}\" >> \"$PROBE_OUT\"\n\
        case \"$*\" in\n\
        \"nextest run\"*)\n\
          if [ -d \"$TMPDIR\" ]; then echo written > \"$TMPDIR/written\"; fi\n\
          exit \"${FAKE_TEST_EXIT:-0}\" ;;\n\
        esac\n\
        exit 0\n";
    for (n, text) in [("cargo", cargo), ("rustfmt", "#!/bin/sh\nexit 0\n")] {
        let path = dir.path().join(n);
        fs::write(&path, text).unwrap();
        fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
    }
    let git_dir = Command::new("sh")
        .args(["-c", "dirname \"$(command -v git)\""])
        .output()
        .unwrap();
    let path = format!(
        "{}:{}:/usr/bin:/bin",
        dir.path().display(),
        String::from_utf8_lossy(&git_dir.stdout).trim()
    );
    (dir, path)
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

/// What one `--card` run showed: the answer and the probe lines of the fake `cargo`.
struct Run {
    answer: Value,
    probe: Vec<String>,
}

impl Run {
    /// The `TMPDIR` the `nextest run` step received.
    fn test_tmpdir(&self) -> String {
        self.tmpdir_of(|l| l.starts_with("nextest run"))
            .pop()
            .expect("the test step ran")
    }

    /// The `TMPDIR` of every cargo step that is not the test run.
    fn other_tmpdirs(&self) -> Vec<String> {
        self.tmpdir_of(|l| !l.starts_with("nextest run"))
    }

    fn tmpdir_of(&self, pick: impl Fn(&str) -> bool) -> Vec<String> {
        self.probe
            .iter()
            .filter(|l| pick(l))
            .map(|l| l.rsplit_once(" | TMPDIR=").unwrap().1.to_string())
            .collect()
    }
}

fn affected(fx: &Fixture, caller_tmp: &Path, test_exit: &str) -> Run {
    let (_tools, path) = fake_path();
    let slots = slot_dir();
    let probe = tempfile::tempdir().unwrap();
    let probe_file = probe.path().join("probe.txt");
    let out = Command::new(BIN)
        .args([
            "check",
            "affected",
            &format!("{}..{}", fx.base, fx.head),
            "--dir",
        ])
        .arg(fx.root())
        .args(["--card", "--slot-dir"])
        .arg(slots.path())
        .env("PATH", &path)
        .env("TMPDIR", caller_tmp)
        .env("PROBE_OUT", &probe_file)
        .env("FAKE_TEST_EXIT", test_exit)
        .env_remove("MOAI_HOME")
        .env_remove("CARGO_TARGET_DIR")
        .stdin(Stdio::null())
        .output()
        .unwrap();
    let text = String::from_utf8_lossy(&out.stdout).to_string();
    let v: Value = serde_json::from_str(&text)
        .unwrap_or_else(|_| panic!("{text} / {}", String::from_utf8_lossy(&out.stderr)));
    let probe = fs::read_to_string(&probe_file)
        .unwrap_or_default()
        .lines()
        .map(String::from)
        .collect();
    Run {
        answer: v["result"].clone(),
        probe,
    }
}

/// Whether `text` holds a session id shape: hex groups 8-4-4-4-12.
fn session_shaped(text: &str) -> bool {
    let b = text.as_bytes();
    let groups = [8usize, 4, 4, 4, 12];
    (0..b.len()).any(|start| {
        let mut at = start;
        groups.iter().enumerate().all(|(i, n)| {
            let ok = b.len() >= at + n && b[at..at + n].iter().all(u8::is_ascii_hexdigit);
            at += n;
            if i < 4 {
                let dash = b.get(at) == Some(&b'-');
                at += 1;
                ok && dash
            } else {
                ok
            }
        })
    })
}

fn has_git_ancestor(path: &Path) -> bool {
    path.ancestors()
        .any(|a| a.join(".git").symlink_metadata().is_ok())
}

/// The three conditions of REQ-14 for the folder the test step received.
fn assert_meets_the_three(seen: &str, fx: &Fixture) {
    let folder = PathBuf::from(seen);
    assert!(folder.is_absolute(), "{seen}");
    let repo = fx.root().canonicalize().unwrap();
    assert!(!folder.starts_with(&repo), "inside the worktree: {seen}");
    assert!(
        !has_git_ancestor(&folder),
        "inside a git repository: {seen}"
    );
    assert!(!session_shaped(seen), "a session id shape: {seen}");
    let name = folder.file_name().unwrap().to_string_lossy().to_string();
    assert!(name.starts_with("moai-affected-"), "{seen}");
}

/// AC-15 (REQ-13, REQ-14, REQ-16): the `test` step gets a new `moai-affected-*` folder as
/// `TMPDIR` that meets the three conditions; the other steps see the caller's `TMPDIR`.
#[test]
fn the_test_step_gets_a_private_tmpdir_that_meets_the_three_conditions() {
    let fx = Fixture::new();
    let caller = tempfile::tempdir().unwrap();
    let caller_path = caller.path().canonicalize().unwrap();
    let run = affected(&fx, &caller_path, "0");
    assert_eq!(run.answer["result"], "PASS", "{}", run.answer);
    let seen = run.test_tmpdir();
    assert_meets_the_three(&seen, &fx);
    assert!(
        Path::new(&seen).starts_with(&caller_path),
        "the caller's folder is a root that fits, so it is the root: {seen}"
    );
    let others = run.other_tmpdirs();
    assert!(!others.is_empty(), "{:?}", run.probe);
    for other in others {
        assert_eq!(other, caller_path.to_string_lossy(), "{:?}", run.probe);
    }
    // A passed run removed it, and the answer says so.
    assert!(!Path::new(&seen).exists(), "{seen}");
    assert_eq!(
        run.answer["test_tmpdir"]["state"], "removed",
        "{}",
        run.answer
    );
}

/// AC-16 (REQ-14): a caller `TMPDIR` that lies inside the worktree, inside a git repository or
/// holds a session id shape is not the root; the step gets a folder that meets the conditions.
#[test]
fn a_caller_tmpdir_that_breaks_a_condition_is_not_the_root() {
    let fx = Fixture::new();
    let holder = tempfile::tempdir().unwrap();
    let repo_elsewhere = holder.path().join("other-repo");
    fs::create_dir_all(repo_elsewhere.join(".git")).unwrap();
    let shaped = holder
        .path()
        .join("0f4c1d2e-aaaa-bbbb-cccc-0123456789ab")
        .join("T");
    fs::create_dir_all(&shaped).unwrap();
    let inside_worktree = fx.root().join("target").join("t");
    fs::create_dir_all(&inside_worktree).unwrap();
    for (what, caller) in [
        ("inside the worktree", inside_worktree),
        ("inside a git repository", repo_elsewhere),
        ("a session id shape", shaped),
    ] {
        let run = affected(&fx, &caller, "0");
        assert_eq!(run.answer["result"], "PASS", "{what}: {}", run.answer);
        let seen = run.test_tmpdir();
        assert_meets_the_three(&seen, &fx);
        assert!(
            !Path::new(&seen).starts_with(caller.canonicalize().unwrap()),
            "{what}: {seen}"
        );
        assert_eq!(
            run.answer["test_tmpdir"]["state"], "removed",
            "{what}: {}",
            run.answer
        );
    }
}

/// AC-19 (REQ-20): a failing `test` step keeps the folder and the answer names it.
#[test]
fn a_failed_run_keeps_the_folder_and_names_it() {
    let fx = Fixture::new();
    let caller = tempfile::tempdir().unwrap();
    let caller_path = caller.path().canonicalize().unwrap();
    let run = affected(&fx, &caller_path, "100");
    assert_eq!(run.answer["result"], "FAIL", "{}", run.answer);
    let seen = run.test_tmpdir();
    assert_eq!(run.answer["test_tmpdir"]["state"], "kept", "{}", run.answer);
    assert_eq!(
        run.answer["test_tmpdir"]["path"],
        seen.as_str(),
        "{}",
        run.answer
    );
    assert_eq!(
        fs::read_to_string(Path::new(&seen).join("written")).unwrap(),
        "written\n",
        "the folder holds what the tests wrote"
    );
    // The folder is the only thing the run left under the caller's `TMPDIR`.
    let left: Vec<_> = fs::read_dir(&caller_path)
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().to_string())
        .collect();
    assert_eq!(left.len(), 1, "{left:?}");
}

/// AC-18 (REQ-18): a sibling of the created folder under the same root survives a passed run.
#[test]
fn a_passed_run_removes_only_its_own_folder() {
    let fx = Fixture::new();
    let caller = tempfile::tempdir().unwrap();
    let caller_path = caller.path().canonicalize().unwrap();
    fs::create_dir(caller_path.join("moai-affected-sibling")).unwrap();
    fs::write(caller_path.join("moai-affected-sibling/keep"), "keep\n").unwrap();
    fs::write(caller_path.join("plain-file"), "keep\n").unwrap();
    let run = affected(&fx, &caller_path, "0");
    assert_eq!(
        run.answer["test_tmpdir"]["state"], "removed",
        "{}",
        run.answer
    );
    assert_eq!(
        fs::read_to_string(caller_path.join("moai-affected-sibling/keep")).unwrap(),
        "keep\n"
    );
    assert_eq!(
        fs::read_to_string(caller_path.join("plain-file")).unwrap(),
        "keep\n"
    );
    let left = fs::read_dir(&caller_path).unwrap().count();
    assert_eq!(left, 2, "only the sibling and the file remain");
}

/// AC-20 (REQ-21): the summary line has only the words it had before this revision.
#[test]
fn the_summary_line_has_no_new_word() {
    let fx = Fixture::new();
    let caller = tempfile::tempdir().unwrap();
    let run = affected(&fx, &caller.path().canonicalize().unwrap(), "0");
    let line = run.answer["summary"].as_str().unwrap();
    assert!(line.ends_with(" result=PASS"), "{line}");
    let known = [
        "base",
        "head",
        "targets",
        "scans",
        "selection",
        "card_by",
        "test-threads",
        "diff_check",
        "fmt",
        "ops_rustfmt",
        "test_tables",
        "bash_n",
        "clippy",
        "test",
        "result",
    ];
    for word in line.split(' ').skip(1) {
        let key = word.split_once('=').map_or(word, |(k, _)| k);
        assert!(known.contains(&key), "a new word `{key}` in {line}");
    }
    assert!(!line.contains("tmp"), "{line}");
}
