//! #2879 (disk hygiene, D2): `worktree-gc-plan` removes merged worktrees whose only ignored
//! files are safe to lose, and a plan cut short by its budget is continued by the next run.
//!
//! Safe ignored files (the one table in `src/worktree_gc.rs`): regenerable caches
//! (`__pycache__/`, `.ruff_cache/`, `.pytest_cache/`, `*.pyc`), a lane's temporary files
//! (`target/<lane>/…` such as `target/agent-9/`, `*.log`) and byte-identical copies of the same
//! path in the primary checkout (what `.worktreeinclude` copied). Anything else — a file that
//! differs from the primary's copy or exists only in the worktree — keeps the worktree.
//! Synthetic repositories in temporary folders only.
#![cfg(unix)]
mod common;

use serde_json::Value;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Command;

const BIN: &str = crate::common::MOAI_BIN;

fn git(dir: &Path, args: &[&str]) -> String {
    let mut all = vec![
        "-c",
        "user.name=fixture",
        "-c",
        "user.email=fixture@example.invalid",
    ];
    all.extend_from_slice(args);
    common::git::git(dir, &all)
}

struct Fx {
    _dir: tempfile::TempDir,
    root: PathBuf,
    home: PathBuf,
}

fn write(p: &Path, body: &str) {
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::write(p, body).unwrap();
}

/// A primary checkout ignoring the usual caches and `.moai/`, with `.moai/shared.md` in the
/// primary (the file `.worktreeinclude` would copy) and merged worktrees `names`.
fn fixture(names: &[&str]) -> Fx {
    let dir = tempfile::TempDir::new().unwrap();
    let base = fs::canonicalize(dir.path()).unwrap();
    let root = base.join("repo");
    let home = base.join("home");
    fs::create_dir_all(&root).unwrap();
    fs::create_dir_all(&home).unwrap();
    git(&root, &["init", "-q", "-b", "main"]);
    fs::write(
        root.join(".gitignore"),
        ".moai-cli/\ntarget/\n.moai/\n__pycache__/\n.ruff_cache/\n.pytest_cache/\n*.pyc\n*.log\n",
    )
    .unwrap();
    fs::write(root.join("a.txt"), "a\n").unwrap();
    git(&root, &["add", ".gitignore", "a.txt"]);
    git(&root, &["commit", "-q", "-m", "init"]);
    git(&root, &["branch", "integ"]);
    write(&root.join(".moai/shared.md"), "shared report\n");
    for n in names {
        let wt = root.join(".moai-cli/worktrees").join(n);
        git(
            &root,
            &[
                "worktree",
                "add",
                "-q",
                "-b",
                &format!("issue/1-{n}"),
                wt.to_str().unwrap(),
            ],
        );
        // Build output, so the worktree is not the empty kind that is kept on purpose.
        fs::create_dir_all(wt.join("target/debug")).unwrap();
        fs::write(
            wt.join("target/CACHEDIR.TAG"),
            "Signature: 8a477f597d28d172789f06886806bc55\n",
        )
        .unwrap();
        fs::write(wt.join("target/debug/.cargo-lock"), "").unwrap();
    }
    Fx {
        _dir: dir,
        root,
        home,
    }
}

impl Fx {
    fn wt(&self, n: &str) -> PathBuf {
        self.root.join(".moai-cli/worktrees").join(n)
    }
}

fn plan(fx: &Fx, extra: &[&str]) -> Value {
    let later = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap()
        .as_secs()
        + 86_400;
    let out = Command::new(BIN)
        .args(["worktree-gc-plan"])
        .arg(&fx.root)
        .args([
            "--integration-ref",
            "integ",
            "--now-unix",
            &later.to_string(),
        ])
        .args(extra)
        .current_dir(&fx.root)
        .env("HOME", &fx.home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)));
    assert_eq!(v["ok"], true, "{v}");
    v["result"].clone()
}

fn entry<'a>(r: &'a Value, path: &str) -> &'a Value {
    r["candidates"]
        .as_array()
        .unwrap()
        .iter()
        .find(|c| c["path"] == path)
        .unwrap_or_else(|| panic!("no {path}: {r}"))
}

fn reasons(c: &Value) -> Vec<String> {
    c["reasons"]
        .as_array()
        .unwrap()
        .iter()
        .map(|v| v.as_str().unwrap().to_string())
        .collect()
}

/// AC-1: a merged worktree whose ignored files are only caches, lane temporary files and a
/// byte-identical copy of the primary's `.moai/shared.md` is removed; one more file that
/// exists only in the worktree, or a copy that differs (in length, or only in its bytes), keeps it.
#[test]
fn only_safe_ignored_files_let_a_merged_worktree_go() {
    let fx = fixture(&["safe", "unique", "changed", "samelen", "dirty"]);
    for n in ["safe", "unique", "changed", "samelen", "dirty"] {
        let wt = fx.wt(n);
        write(&wt.join("tools/__pycache__/x.cpython-313.pyc"), "bytecode");
        write(&wt.join(".ruff_cache/0.1/cache"), "c");
        write(&wt.join("target/agent-9/notes.md"), "lane note\n");
        write(&wt.join("run.log"), "log\n");
        write(&wt.join(".moai/shared.md"), "shared report\n");
    }
    write(&fx.wt("unique").join(".moai/research/only-here.md"), "x\n");
    write(&fx.wt("changed").join(".moai/shared.md"), "edited\n");
    // Same length as the primary's copy, one byte different: only a byte compare tells.
    write(&fx.wt("samelen").join(".moai/shared.md"), "shared reporX\n");
    fs::write(fx.wt("dirty").join("a.txt"), "changed\n").unwrap();

    let r = plan(&fx, &[]);
    let safe = entry(&r, ".moai-cli/worktrees/safe");
    assert_eq!(safe["action"], "remove", "{safe}");
    assert_eq!(safe["ignored_outside_build"], 0, "{safe}");
    assert!(safe["ignored_safe"].as_u64().unwrap() >= 5, "{safe}");
    for n in ["unique", "changed", "samelen"] {
        let c = entry(&r, &format!(".moai-cli/worktrees/{n}"));
        assert_eq!(c["action"], "trim", "{n}: {c}");
        assert_eq!(reasons(c), ["ignored_files_outside_build_output"], "{n}");
        assert_eq!(c["ignored_outside_build"], 1, "{n}: {c}");
        assert_eq!(
            c["ignored_kept"][0],
            if n == "unique" {
                ".moai/research/only-here.md"
            } else {
                ".moai/shared.md"
            },
            "{n}: {c}"
        );
    }
    let d = entry(&r, ".moai-cli/worktrees/dirty");
    assert_eq!(d["action"], "skip", "{d}");
    assert!(reasons(d).contains(&"tracked_changes".to_string()), "{d}");
}

/// A symbolic link among the ignored files is never counted safe, even when its name looks
/// like a cache: removing the worktree would be fine, but the plan stays conservative.
#[test]
fn a_symlink_among_ignored_files_keeps_the_worktree() {
    let fx = fixture(&["linky"]);
    let wt = fx.wt("linky");
    fs::create_dir_all(wt.join("tools/__pycache__")).unwrap();
    std::os::unix::fs::symlink("/etc/hosts", wt.join("tools/__pycache__/x.pyc")).unwrap();
    let r = plan(&fx, &[]);
    let c = entry(&r, ".moai-cli/worktrees/linky");
    assert_ne!(c["action"], "remove", "{c}");
    assert_eq!(c["ignored_outside_build"], 1, "{c}");
}

/// AC-2: with `--max-checked N` a plan checks at most N worktrees, marks the rest
/// `budget_spent`, says how many remain and where to go on; the next run with `--cursor`
/// checks the rest. N/N+1 at the edge: as many worktrees as N leaves none remaining.
#[test]
fn a_plan_cut_short_is_continued_by_the_next_run() {
    let fx = fixture(&["w1", "w2", "w3"]);
    for n in ["w1", "w2", "w3"] {
        write(&fx.wt(n).join("tools/__pycache__/x.pyc"), "b");
    }
    let checked = |r: &Value| -> Vec<String> {
        r["candidates"]
            .as_array()
            .unwrap()
            .iter()
            .filter(|c| c["path"] != ".")
            .filter(|c| !reasons(c).contains(&"budget_spent".to_string()))
            .map(|c| c["path"].as_str().unwrap().to_string())
            .collect()
    };
    let first = plan(&fx, &["--max-checked", "2"]);
    assert_eq!(first["remaining"], 1, "{first}");
    let a = checked(&first);
    assert_eq!(a.len(), 2, "{first}");
    let cursor = first["next_cursor"].as_str().unwrap().to_string();
    assert_eq!(cursor, a[1]);
    let second = plan(&fx, &["--max-checked", "2", "--cursor", &cursor]);
    let b = checked(&second);
    // Candidates are listed by path, so order is not visible; what matters is that the one
    // left out by the first run is checked by the second (it starts after the cursor).
    let left: Vec<&String> = ["w1", "w2", "w3"]
        .iter()
        .map(|n| format!(".moai-cli/worktrees/{n}"))
        .filter(|p| !a.contains(p))
        .collect::<Vec<_>>()
        .leak()
        .iter()
        .collect();
    assert_eq!(left.len(), 1);
    assert!(b.contains(left[0]), "{a:?} then {b:?}");
    assert_eq!(b.len(), 2, "{second}");
    let mut all: Vec<String> = a.iter().chain(b.iter()).cloned().collect();
    all.sort();
    all.dedup();
    assert_eq!(
        all.len(),
        3,
        "every worktree checked across two runs: {all:?}"
    );
    // The edge: N worktrees and N allowed leaves nothing remaining and no cursor.
    let all3 = plan(&fx, &["--max-checked", "3"]);
    assert_eq!(all3["remaining"], 0, "{all3}");
    assert_eq!(all3["next_cursor"], Value::Null, "{all3}");
    assert_eq!(all3["totals"]["remove"], 3, "{all3}");
    // Without the flags nothing is cut short.
    assert_eq!(plan(&fx, &[])["remaining"], 0);
}

/// AC-1 end to end: the apply removes the safe worktree (its caches, lane notes and the copy
/// go with it), keeps the one holding a file found nowhere else, and never touches the
/// primary's `.moai/`.
#[test]
fn apply_removes_only_the_safe_worktree() {
    let fx = fixture(&["safe", "unique"]);
    for n in ["safe", "unique"] {
        write(&fx.wt(n).join("tools/__pycache__/x.pyc"), "b");
        write(&fx.wt(n).join("target/agent-9/notes.md"), "lane note\n");
        write(&fx.wt(n).join(".moai/shared.md"), "shared report\n");
    }
    write(&fx.wt("unique").join(".moai/only-here.md"), "x\n");
    let r = plan(&fx, &[]);
    let later = r["now_unix"].as_u64().unwrap().to_string();
    let out = Command::new(BIN)
        .args(["worktree-gc-apply"])
        .arg(&fx.root)
        .args([
            "--integration-ref",
            "integ",
            "--now-unix",
            &later,
            "--plan-digest",
            r["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ])
        .current_dir(&fx.root)
        .env("HOME", &fx.home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], true, "{v}");
    assert!(!fx.wt("safe").exists(), "{v}");
    assert!(fx.wt("unique").join(".moai/only-here.md").is_file(), "{v}");
    assert_eq!(
        fs::read_to_string(fx.root.join(".moai/shared.md")).unwrap(),
        "shared report\n"
    );
}

/// #2879 D1: `--worktree <path>` limits the plan (and so the apply) to the lane's own card
/// worktrees; other lanes' merged worktrees are not even listed.
#[test]
fn worktree_option_limits_the_plan_to_the_named_worktrees() {
    let fx = fixture(&["mine", "other"]);
    let mine = fx.wt("mine");
    let r = plan(&fx, &["--worktree", mine.to_str().unwrap()]);
    let paths: Vec<&str> = r["candidates"]
        .as_array()
        .unwrap()
        .iter()
        .map(|c| c["path"].as_str().unwrap())
        .collect();
    assert_eq!(paths, [".moai-cli/worktrees/mine"], "{r}");
    assert_eq!(r["totals"]["remove"], 1, "{r}");
    // A path that is no worktree of this clone is refused, not silently ignored.
    let out = Command::new(BIN)
        .args(["worktree-gc-plan"])
        .arg(&fx.root)
        .args([
            "--integration-ref",
            "integ",
            "--worktree",
            "/nonexistent-wt",
        ])
        .current_dir(&fx.root)
        .env("HOME", &fx.home)
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], false, "{v}");
    assert!(
        v.to_string().contains("not a worktree of this clone"),
        "{v}"
    );
}

/// #2879: the statusline's state files under `.moai/state/` (session context usage, GitHub
/// counts) are regenerable; any other `.moai/` file keeps the worktree.
#[test]
fn statusline_state_files_are_regenerable() {
    let fx = fixture(&["st", "rep"]);
    for n in ["st", "rep"] {
        write(
            &fx.wt(n).join(".moai/state/context-usage/1234-abcd.json"),
            "{}",
        );
        write(&fx.wt(n).join(".moai/state/github/counts.json"), "{}");
    }
    write(&fx.wt("rep").join(".moai/state/active-sessions.json"), "[]");
    let r = plan(&fx, &[]);
    assert_eq!(
        entry(&r, ".moai-cli/worktrees/st")["action"],
        "remove",
        "{r}"
    );
    let rep = entry(&r, ".moai-cli/worktrees/rep");
    assert_ne!(rep["action"], "remove", "{rep}");
    assert_eq!(rep["ignored_kept"][0], ".moai/state/active-sessions.json");
}

fn plan_at(fx: &Fx, now: u64) -> Value {
    let out = Command::new(BIN)
        .args(["worktree-gc-plan"])
        .arg(&fx.root)
        .args(["--integration-ref", "integ", "--now-unix", &now.to_string()])
        .current_dir(&fx.root)
        .env("HOME", &fx.home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)));
    assert_eq!(v["ok"], true, "{v}");
    v["result"].clone()
}

fn apply(fx: &Fx, r: &Value) -> Value {
    let out = Command::new(BIN)
        .args(["worktree-gc-apply"])
        .arg(&fx.root)
        .args([
            "--integration-ref",
            "integ",
            "--now-unix",
            &r["now_unix"].as_u64().unwrap().to_string(),
            "--plan-digest",
            r["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ])
        .current_dir(&fx.root)
        .env("HOME", &fx.home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .output()
        .unwrap();
    serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)))
}

fn kept(c: &Value) -> Vec<String> {
    c["ignored_kept"]
        .as_array()
        .unwrap()
        .iter()
        .map(|v| v.as_str().unwrap().to_string())
        .collect()
}

/// Audit B1 on PR #2896: a clone with an unpushed commit inside the lane's temporary folder
/// `target/agent-N/` keeps the worktree, and the apply leaves the clone and its commit as
/// they were. A lane folder of plain notes alone still lets the worktree go.
#[test]
fn a_nested_clone_in_the_lane_folder_keeps_the_worktree() {
    let fx = fixture(&["nest", "plain"]);
    write(&fx.wt("plain").join("target/agent-3/notes.md"), "n\n");
    let clone = fx.wt("nest").join("target/agent-3/clone");
    fs::create_dir_all(&clone).unwrap();
    git(&clone, &["init", "-q", "-b", "main"]);
    fs::write(clone.join("work.txt"), "unpushed\n").unwrap();
    git(&clone, &["add", "work.txt"]);
    git(&clone, &["commit", "-q", "-m", "unpushed"]);
    let head = git(&clone, &["rev-parse", "HEAD"]);

    let r = plan(&fx, &[]);
    let c = entry(&r, ".moai-cli/worktrees/nest");
    assert_ne!(c["action"], "remove", "{c}");
    assert!(
        kept(c).contains(&"target/agent-3/clone/.git".to_string()),
        "{c}"
    );
    assert_eq!(
        entry(&r, ".moai-cli/worktrees/plain")["action"],
        "remove",
        "{r}"
    );
    let v = apply(&fx, &r);
    assert_eq!(v["ok"], true, "{v}");
    assert!(!fx.wt("plain").exists(), "{v}");
    assert!(fx.wt("nest").is_dir(), "{v}");
    assert_eq!(git(&clone, &["rev-parse", "HEAD"]), head);
}

/// Audit B1 on PR #2896: a linked worktree of the same clone placed inside the lane folder
/// keeps the outer worktree; the plan names the nested worktree's path, and the apply leaves
/// its uncommitted change in place.
#[test]
fn a_nested_linked_worktree_in_the_lane_folder_keeps_the_worktree() {
    let fx = fixture(&["outer"]);
    let inner = fx.wt("outer").join("target/agent-3/wt2");
    fs::create_dir_all(inner.parent().unwrap()).unwrap();
    git(
        &fx.root,
        &[
            "worktree",
            "add",
            "-q",
            "-b",
            "issue/1-inner",
            inner.to_str().unwrap(),
        ],
    );
    fs::write(inner.join("a.txt"), "uncommitted\n").unwrap();

    let r = plan(&fx, &[]);
    let c = entry(&r, ".moai-cli/worktrees/outer");
    assert_ne!(c["action"], "remove", "{c}");
    assert!(kept(c).contains(&"target/agent-3/wt2".to_string()), "{c}");
    let v = apply(&fx, &r);
    assert_eq!(v["ok"], true, "{v}");
    assert!(fx.wt("outer").is_dir(), "{v}");
    assert_eq!(
        fs::read_to_string(inner.join("a.txt")).unwrap(),
        "uncommitted\n"
    );
}

/// Audit B2 on PR #2896: a file the plan counts safe (a lane folder note, a `*.log`) that was
/// written inside the recent window makes the worktree skip with `recent_change`; the same
/// worktree without it is removed. The files are dated two days ahead and the plan runs a
/// minute after that, so every other change-age path is two days old.
#[test]
fn a_recent_write_to_a_safe_file_keeps_the_worktree() {
    let fx = fixture(&["lane", "log", "old"]);
    let ahead = std::time::SystemTime::now() + std::time::Duration::from_secs(2 * 86_400);
    for (n, rel) in [
        ("lane", "target/agent-3/deep/evidence.md"),
        ("log", "logs/deep/run.log"),
    ] {
        let p = fx.wt(n).join(rel);
        write(&p, "fresh\n");
        fs::File::options()
            .write(true)
            .open(&p)
            .unwrap()
            .set_modified(ahead)
            .unwrap();
    }
    let at = ahead
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap()
        .as_secs()
        + 60;
    let r = plan_at(&fx, at);
    for n in ["lane", "log"] {
        let c = entry(&r, &format!(".moai-cli/worktrees/{n}"));
        assert_eq!(c["action"], "skip", "{n}: {c}");
        assert!(
            reasons(c).contains(&"recent_change".to_string()),
            "{n}: {c}"
        );
    }
    assert_eq!(
        entry(&r, ".moai-cli/worktrees/old")["action"],
        "remove",
        "{r}"
    );
}
