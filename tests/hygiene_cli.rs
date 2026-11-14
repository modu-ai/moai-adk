//! #2934 S4a (#2930 R-4): `hygiene-plan` holds only merged, clean, idle worktrees and idle,
//! unused cargo build outputs in the temporary folder; everything in use, dirty, unmerged,
//! recent or not a build output is kept or not listed. It writes nothing. Temporary HOME,
//! repository and temporary folder only.
#![cfg(unix)]
mod common;

use serde_json::Value;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Child, Command, Stdio};
use std::time::UNIX_EPOCH;

const BIN: &str = crate::common::MOAI_BIN;
const TAG: &str = "Signature: 8a477f597d28d172789f06886806bc55\n";

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
    repo: PathBuf,
    home: PathBuf,
    tmp: PathBuf,
}

fn fixture() -> Fx {
    let dir = tempfile::TempDir::new().unwrap();
    let root = fs::canonicalize(dir.path()).unwrap();
    let (repo, home, tmp) = (root.join("repo"), root.join("home"), root.join("tmp"));
    for d in [&repo, &home, &tmp] {
        fs::create_dir_all(d).unwrap();
    }
    git(&repo, &["init", "-q", "-b", "main"]);
    fs::write(repo.join("a.txt"), "a\n").unwrap();
    git(&repo, &["add", "a.txt"]);
    git(&repo, &["commit", "-q", "-m", "init"]);
    // `merged`: its HEAD is main's; `dirty`: the same, with an uncommitted file.
    for name in ["merged", "dirty"] {
        let wt = root.join(format!("wt-{name}"));
        git(
            &repo,
            &["worktree", "add", "-q", "-b", name, wt.to_str().unwrap()],
        );
    }
    fs::write(root.join("wt-dirty/new.txt"), "x\n").unwrap();
    // Temporary folder: tagged build outputs, an untagged folder, a link to a tagged one.
    for name in ["build-idle", "build-cwd", "build-env"] {
        let d = tmp.join(name);
        fs::create_dir_all(d.join("debug")).unwrap();
        fs::write(d.join("CACHEDIR.TAG"), TAG).unwrap();
    }
    fs::create_dir_all(tmp.join("plain/debug")).unwrap();
    std::os::unix::fs::symlink(tmp.join("build-idle"), tmp.join("link-to-build")).unwrap();
    Fx {
        _dir: dir,
        repo,
        home,
        tmp,
    }
}

fn snapshot(root: &Path) -> Vec<(PathBuf, u64)> {
    let mut out = Vec::new();
    let mut stack = vec![root.to_path_buf()];
    while let Some(d) = stack.pop() {
        for e in fs::read_dir(&d).unwrap().flatten() {
            let m = fs::symlink_metadata(e.path()).unwrap();
            out.push((e.path(), m.len()));
            if m.is_dir() {
                stack.push(e.path());
            }
        }
    }
    out.sort();
    out
}

fn mtime(p: &Path) -> u64 {
    fs::symlink_metadata(p)
        .unwrap()
        .modified()
        .unwrap()
        .duration_since(UNIX_EPOCH)
        .unwrap()
        .as_secs()
}

/// The newest change at the top of a build output, as the plan reads it.
fn newest_top(d: &Path) -> u64 {
    fs::read_dir(d)
        .unwrap()
        .flatten()
        .map(|e| mtime(&e.path()))
        .fold(mtime(d), u64::max)
}

fn plan(fx: &Fx, now: u64) -> Value {
    let out = Command::new(BIN)
        .args(["hygiene-plan", fx.repo.to_str().unwrap()])
        .args([
            "--integration-ref",
            "main",
            "--tmp",
            fx.tmp.to_str().unwrap(),
        ])
        .args(["--now", &now.to_string()])
        .env("HOME", &fx.home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)));
    assert_eq!(v["ok"], true, "{v}");
    assert_eq!(v["operation"], "hygiene.plan");
    v["result"].clone()
}

fn item<'a>(r: &'a Value, suffix: &str) -> Option<&'a Value> {
    r["items"]
        .as_array()
        .unwrap()
        .iter()
        .find(|i| i["path"].as_str().unwrap().ends_with(suffix))
}

fn action(r: &Value, suffix: &str) -> (String, String) {
    let i = item(r, suffix).unwrap_or_else(|| panic!("{suffix} not listed: {r}"));
    (
        i["action"].as_str().unwrap().into(),
        i["reason"].as_str().unwrap_or("").into(),
    )
}

/// A process working in `dir`, and one naming `dir` in its environment.
fn users(fx: &Fx) -> (Child, Child) {
    let cwd = Command::new("/bin/sleep")
        .arg("60")
        .current_dir(fx.tmp.join("build-cwd"))
        .stdin(Stdio::null())
        .spawn()
        .unwrap();
    // Named in its command line (`sh -c '/bin/sleep 61; :' sh <dir>` — no `exec`, so the shell
    // keeps its arguments). On Linux the environment
    // counts too; macOS shows no other process's environment (`hygiene::process_text`).
    let named = Command::new("/bin/sh")
        .args(["-c", "/bin/sleep 61; :", "sh"])
        .arg(fx.tmp.join("build-env"))
        .env("CARGO_TARGET_DIR", fx.tmp.join("build-env"))
        .stdin(Stdio::null())
        .spawn()
        .unwrap();
    (cwd, named)
}

#[test]
fn only_idle_unused_merged_clean_leftovers_are_held_and_nothing_is_written() {
    let fx = fixture();
    let (mut a, mut b) = users(&fx);
    let before_tmp = snapshot(&fx.tmp);
    let before_home = snapshot(&fx.home);
    let far = newest_top(&fx.tmp.join("build-idle")) + 10 * 86_400;
    let r = plan(&fx, far);
    // The same state gives the same digest.
    let again = plan(&fx, far);
    let _ = (a.kill(), b.kill(), a.wait(), b.wait());
    assert_eq!(again["plan_digest"], r["plan_digest"]);
    assert_eq!(r["writes"], false);
    assert_eq!(
        action(&r, "/tmp/build-idle"),
        ("hold".into(), "idle_build_output".into())
    );
    assert_eq!(
        action(&r, "/tmp/build-cwd"),
        ("keep".into(), "in_use".into())
    );
    assert_eq!(
        action(&r, "/tmp/build-env"),
        ("keep".into(), "in_use".into())
    );
    assert!(
        item(&r, "/tmp/plain").is_none(),
        "untagged folders are not build outputs: {r}"
    );
    assert!(
        item(&r, "/tmp/link-to-build").is_none(),
        "links are not followed: {r}"
    );
    assert_eq!(action(&r, "/wt-merged").0, "hold", "{r}");
    assert_eq!(action(&r, "/wt-dirty").0, "keep", "{r}");
    let held = item(&r, "/tmp/build-idle").unwrap();
    assert!(held["bytes"].as_u64().unwrap() > 0, "{held}");
    let date = r["holding"].as_str().unwrap();
    assert!(
        date.starts_with(fx.home.join(".moai/holding/").to_str().unwrap()),
        "{date}"
    );
    assert_eq!(r["totals"]["hold"], 2, "{r}");
    // Reading wrote nothing anywhere.
    assert_eq!(snapshot(&fx.tmp), before_tmp);
    assert_eq!(snapshot(&fx.home), before_home);
}

/// The 2-day age limit, N/N+1: one second short of it the build output is `recent`.
#[test]
fn a_build_output_is_held_from_exactly_two_days_on() {
    let fx = fixture();
    let newest = newest_top(&fx.tmp.join("build-idle"));
    let limit = 2 * 86_400;
    let r = plan(&fx, newest + limit - 1);
    assert_eq!(
        action(&r, "/tmp/build-idle"),
        ("keep".into(), "recent".into())
    );
    let r = plan(&fx, newest + limit);
    assert_eq!(action(&r, "/tmp/build-idle").0, "hold");
    // A new item changes what an apply would act on, and so the digest.
    let before = r["plan_digest"].clone();
    let d = fx.tmp.join("build-new");
    fs::create_dir_all(&d).unwrap();
    fs::write(d.join("CACHEDIR.TAG"), TAG).unwrap();
    let r = plan(&fx, newest_top(&d) + limit);
    assert_ne!(r["plan_digest"], before);
}
