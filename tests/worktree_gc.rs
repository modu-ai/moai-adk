//! WT-4 (#2380): `moai-cli worktree-gc-plan` on a synthetic repository. Nothing is deleted;
//! the plan's actions are checked against the facts each fixture worktree was built with.
#![cfg(unix)]
mod common;

use serde_json::Value;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Command;

const BIN: &str = env!("CARGO_BIN_EXE_moai-cli");

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

struct Fixture {
    _dir: tempfile::TempDir,
    root: PathBuf,
    home: PathBuf,
}

fn cargo_output(wt: &Path) {
    fs::create_dir_all(wt.join("target/debug/deps")).unwrap();
    fs::write(
        wt.join("target/CACHEDIR.TAG"),
        "Signature: 8a477f597d28d172789f06886806bc55\n",
    )
    .unwrap();
    fs::write(wt.join("target/debug/.cargo-lock"), "").unwrap();
    fs::write(wt.join("target/debug/deps/lib.rlib"), vec![7u8; 65_536]).unwrap();
}

/// Worktrees under `.moai-cli/worktrees/`, integration ref `integ` = the base commit:
/// - `a` merged, clean, cargo output only -> remove
/// - `b` merged, cargo output plus a lane note `target/notes/pr.md` -> trim, note preserved
/// - `c` one commit not in `integ` -> skip not_merged
/// - `d` merged, a tracked file modified -> skip tracked_changes
/// - `e` merged, an untracked file, cargo output -> trim untracked_files
/// - `f` merged, locked -> skip locked
fn fixture() -> Fixture {
    let dir = tempfile::TempDir::new().unwrap();
    let base = fs::canonicalize(dir.path()).unwrap();
    let root = base.join("repo");
    let home = base.join("home");
    fs::create_dir_all(&root).unwrap();
    fs::create_dir_all(&home).unwrap();
    git(&root, &["init", "-q", "-b", "main"]);
    fs::write(root.join(".gitignore"), ".moai-cli/\ntarget/\n").unwrap();
    fs::write(root.join("a.txt"), "a\n").unwrap();
    git(&root, &["add", ".gitignore", "a.txt"]);
    git(&root, &["commit", "-q", "-m", "init"]);
    git(&root, &["branch", "integ"]);
    for (name, mode) in [
        ("a", "branch"),
        ("b", "branch"),
        ("c", "branch"),
        ("d", "branch"),
        ("e", "detach"),
        ("f", "branch"),
    ] {
        let wt = root.join(".moai-cli/worktrees").join(name);
        let wt_s = wt.to_str().unwrap().to_string();
        let branch = format!("issue/1-{name}");
        if mode == "detach" {
            git(&root, &["worktree", "add", "-q", "--detach", &wt_s, "HEAD"]);
        } else {
            git(&root, &["worktree", "add", "-q", "-b", &branch, &wt_s]);
        }
    }
    let wt = |n: &str| root.join(".moai-cli/worktrees").join(n);
    for n in ["a", "b", "e"] {
        cargo_output(&wt(n));
    }
    fs::create_dir_all(wt("b").join("target/notes")).unwrap();
    fs::write(wt("b").join("target/notes/pr.md"), "lane note\n").unwrap();
    fs::write(wt("c").join("c.txt"), "c\n").unwrap();
    git(&wt("c"), &["add", "c.txt"]);
    git(&wt("c"), &["commit", "-q", "-m", "unmerged"]);
    fs::write(wt("d").join("a.txt"), "changed\n").unwrap();
    fs::write(wt("e").join("loose.txt"), "u\n").unwrap();
    let f_s = wt("f").to_str().unwrap().to_string();
    git(
        &root,
        &["worktree", "lock", "--reason", "moai claim t e1", &f_s],
    );
    Fixture {
        _dir: dir,
        root,
        home,
    }
}

fn plan(fx: &Fixture, now: Option<u64>) -> Value {
    plan_from(fx, &fx.root, now)
}

/// The plan run with `cwd` as the process's working directory.
fn plan_from(fx: &Fixture, cwd: &Path, now: Option<u64>) -> Value {
    plan_with(fx, cwd, now, &[])
}

/// The plan run from `cwd` with extra command-line arguments.
fn plan_with(fx: &Fixture, cwd: &Path, now: Option<u64>, extra: &[&str]) -> Value {
    let mut cmd = Command::new(BIN);
    cmd.args(["worktree-gc-plan"])
        .arg(&fx.root)
        .args(["--integration-ref", "integ"])
        .args(extra)
        .current_dir(cwd)
        .env("HOME", &fx.home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null");
    if let Some(n) = now {
        cmd.args(["--now-unix", &n.to_string()]);
    }
    let out = cmd.output().unwrap();
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], true, "{v}");
    assert_eq!(v["operation"], "worktree.gc.plan");
    v["result"].clone()
}

/// A day from now: every fixture change is older than the 120-minute window.
fn later() -> u64 {
    std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap()
        .as_secs()
        + 86_400
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

#[test]
fn each_worktree_gets_the_action_its_facts_allow() {
    let fx = fixture();
    let r = plan(&fx, Some(later()));
    assert_eq!(r["writes"], false);
    assert_eq!(r["network"], "disabled");
    assert_eq!(r["integration_tip"], git(&fx.root, &["rev-parse", "integ"]));

    let primary = entry(&r, ".");
    assert_eq!(primary["action"], "skip");
    assert!(
        reasons(primary).contains(&"primary".to_string()),
        "{primary}"
    );

    let a = entry(&r, ".moai-cli/worktrees/a");
    assert_eq!(a["action"], "remove", "{a}");
    assert!(a["trim_bytes"].as_u64().unwrap() >= 65_536, "{a}");

    let b = entry(&r, ".moai-cli/worktrees/b");
    assert_eq!(b["action"], "trim", "{b}");
    assert_eq!(reasons(b), ["ignored_files_outside_build_output"]);
    // The lane note is never a trim path; only cargo-owned entries are.
    assert_eq!(b["preserved_in_target"], serde_json::json!(["notes"]));
    let trim: Vec<&str> = b["trim_paths"]
        .as_array()
        .unwrap()
        .iter()
        .map(|v| v.as_str().unwrap())
        .collect();
    assert_eq!(trim, ["target/CACHEDIR.TAG", "target/debug"]);

    assert_eq!(reasons(entry(&r, ".moai-cli/worktrees/c")), ["not_merged"]);
    assert_eq!(
        reasons(entry(&r, ".moai-cli/worktrees/d")),
        ["tracked_changes"]
    );
    let e = entry(&r, ".moai-cli/worktrees/e");
    assert_eq!(e["action"], "trim", "{e}");
    assert_eq!(reasons(e), ["untracked_files"]);
    assert_eq!(reasons(entry(&r, ".moai-cli/worktrees/f")), ["locked"]);

    assert_eq!(r["totals"]["remove"], 1);
    assert_eq!(r["totals"]["trim"], 2);
    assert_eq!(r["totals"]["skip"], 4);
}

#[test]
fn a_recent_change_keeps_every_worktree() {
    let fx = fixture();
    // The real clock: the fixtures were just made, inside the 120-minute window.
    let r = plan(&fx, None);
    let a = entry(&r, ".moai-cli/worktrees/a");
    assert_eq!(a["action"], "skip", "{a}");
    assert!(reasons(a).contains(&"recent_change".to_string()), "{a}");
    assert_eq!(r["totals"]["remove"], 0);
    assert_eq!(r["totals"]["trim"], 0);
}

/// Pre-audit FAIL 6777336338 B1 (X1): a worktree under `.claude/worktrees/` is preserved
/// state (Q8) — merged, clean and holding only cargo output, it is still `skip` with
/// `preserved_location` as its only reason.
#[test]
fn a_worktree_under_claude_worktrees_is_preserved() {
    let fx = fixture();
    let g = fx.root.join(".claude/worktrees/g");
    git(
        &fx.root,
        &[
            "worktree",
            "add",
            "-q",
            "-b",
            "issue/1-g",
            g.to_str().unwrap(),
        ],
    );
    cargo_output(&g);
    let r = plan(&fx, Some(later()));
    let e = entry(&r, ".claude/worktrees/g");
    assert_eq!(e["action"], "skip", "{e}");
    assert_eq!(reasons(e), ["preserved_location"]);
}

/// Pre-audit FAIL 6777336338 B1 (X2): the worktree the plan runs in is never a candidate.
/// The plan's own process also has its working directory there, so `process_in_use` holds
/// too; `current_directory` must be named on its own.
#[test]
fn the_worktree_the_plan_runs_in_is_kept() {
    let fx = fixture();
    let a = fx.root.join(".moai-cli/worktrees/a");
    let r = plan_from(&fx, &a, Some(later()));
    let e = entry(&r, ".moai-cli/worktrees/a");
    assert_eq!(e["action"], "skip", "{e}");
    assert!(reasons(e).contains(&"current_directory".to_string()), "{e}");
    // The primary is not the current directory here even though `a` is nested in it.
    assert!(!reasons(entry(&r, ".")).contains(&"current_directory".to_string()));
}

#[test]
fn a_process_working_in_a_worktree_keeps_it() {
    let fx = fixture();
    let a = fx.root.join(".moai-cli/worktrees/a");
    // #2511: the process must outlive `plan` even on a loaded machine; it is killed below.
    let mut child = Command::new("sleep")
        .arg("600")
        .current_dir(&a)
        .spawn()
        .unwrap();
    let r = plan(&fx, Some(later()));
    let _ = child.kill();
    let _ = child.wait();
    let e = entry(&r, ".moai-cli/worktrees/a");
    assert_eq!(e["action"], "skip", "{e}");
    assert_eq!(reasons(e), ["process_in_use"]);
}

/// Audit FAIL 6811382861 B1 (K1): a process whose working directory is a subfolder of a
/// linked worktree keeps it too, not only one at the worktree's top folder.
#[test]
fn a_process_working_in_a_worktree_subfolder_keeps_it() {
    let fx = fixture();
    let sub = fx.root.join(".moai-cli/worktrees/a/target/debug");
    // #2511: outlives `plan` under load; killed after the plan.
    let mut child = Command::new("sleep")
        .arg("600")
        .current_dir(&sub)
        .spawn()
        .unwrap();
    let r = plan(&fx, Some(later()));
    let _ = child.kill();
    let _ = child.wait();
    let e = entry(&r, ".moai-cli/worktrees/a");
    assert_eq!(e["action"], "skip", "{e}");
    assert_eq!(reasons(e), ["process_in_use"]);
}

/// Audit FAIL 6811382861 B1 (K6): the change age is the newest of the proxy paths. Every
/// fixture path is about a day older than `now`; only `target/debug` is set one minute
/// before it, and that one path keeps the worktree.
#[test]
fn one_recent_proxy_path_keeps_the_worktree() {
    let fx = fixture();
    let now = later();
    let debug = fx.root.join(".moai-cli/worktrees/a/target/debug");
    let one_minute_before = std::time::UNIX_EPOCH + std::time::Duration::from_secs(now - 60);
    fs::File::open(&debug)
        .unwrap()
        .set_modified(one_minute_before)
        .unwrap();
    let r = plan(&fx, Some(now));
    let e = entry(&r, ".moai-cli/worktrees/a");
    assert_eq!(e["action"], "skip", "{e}");
    assert_eq!(reasons(e), ["recent_change"]);
    assert_eq!(e["change_age_secs"], 60, "{e}");
}

/// Audit 6811382861 N4 (K8): with the time budget already spent, no worktree is a candidate;
/// every one is `skip` and names `budget_spent`.
#[test]
fn a_spent_budget_skips_every_worktree() {
    let fx = fixture();
    let r = plan_with(&fx, &fx.root, Some(later()), &["--budget-ms", "0"]);
    assert_eq!(r["totals"]["remove"], 0, "{r}");
    assert_eq!(r["totals"]["trim"], 0, "{r}");
    for c in r["candidates"].as_array().unwrap() {
        assert_eq!(c["action"], "skip", "{c}");
        assert!(reasons(c).contains(&"budget_spent".to_string()), "{c}");
    }
}

/// Every file and folder under `dir` with length and modification time (a folder's length is
/// recorded as 0, so removing an empty folder is seen too).
fn snapshot(dir: &Path) -> Vec<(String, u64, std::time::SystemTime)> {
    let mut out = Vec::new();
    let mut stack = vec![dir.to_path_buf()];
    while let Some(p) = stack.pop() {
        let m = fs::symlink_metadata(&p).unwrap();
        if m.is_dir() {
            out.push((
                format!("{}/", p.to_string_lossy()),
                0,
                m.modified().unwrap(),
            ));
            for e in fs::read_dir(&p).unwrap() {
                stack.push(e.unwrap().path());
            }
        } else {
            out.push((
                p.to_string_lossy().into_owned(),
                m.len(),
                m.modified().unwrap(),
            ));
        }
    }
    out.sort();
    out
}

/// What git itself reports about the worktrees, refs and the primary's status
/// (`--no-optional-locks`: reading the status must not refresh the index file).
fn git_state(root: &Path) -> Vec<String> {
    [
        &["worktree", "list", "--porcelain"][..],
        &["for-each-ref", "--format=%(refname) %(objectname)"],
        &["--no-optional-locks", "status", "--porcelain", "--ignored"],
    ]
    .iter()
    .map(|args| git(root, args))
    .collect()
}

#[test]
fn the_plan_writes_nothing_and_its_digest_is_stable() {
    let fx = fixture();
    let now = later();
    let before = snapshot(&fx.root);
    let git_before = git_state(&fx.root);
    let one = plan(&fx, Some(now));
    // The fixture must give the plan something it would remove and trim, or a plan that
    // deleted its candidates would leave nothing changed here.
    assert_eq!(one["totals"]["remove"], 1, "{one}");
    assert_eq!(one["totals"]["trim"], 2, "{one}");
    let two = plan(&fx, Some(now));
    assert_eq!(snapshot(&fx.root), before);
    assert_eq!(git_state(&fx.root), git_before);
    assert_eq!(one["writes"], false);
    assert_eq!(one["plan_digest"], two["plan_digest"]);
    assert!(one["plan_digest"].as_str().unwrap().starts_with("sha256:"));
}

#[test]
fn an_unknown_integration_ref_is_refused() {
    let fx = fixture();
    let out = Command::new(BIN)
        .args(["worktree-gc-plan"])
        .arg(&fx.root)
        .args(["--integration-ref", "no-such-ref"])
        .env("HOME", &fx.home)
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], false, "{v}");
    assert!(v.to_string().contains("integration ref not found"), "{v}");
}
