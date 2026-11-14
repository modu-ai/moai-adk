//! #2934 S4b (#2930 R-4): `hygiene-apply` moves only what a matching `hygiene-plan` holds, by
//! a same-file-system rename into `<HOME>/.moai/holding/<UTC date>/`, records each item in
//! `manifest.tsv` before it moves, re-checks use right before each move, prunes the moved
//! worktrees and refuses a stale digest or a linked holding folder before any write.
//! Temporary HOME, repository and temporary folder only.
#![cfg(unix)]
mod common;

use serde_json::Value;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use std::time::UNIX_EPOCH;

const BIN: &str = crate::common::MOAI_BIN;
const TAG: &str = "Signature: 8a477f597d28d172789f06886806bc55\n";
const HEADER: &str = "original\tholding\tkind\tbytes\treason\n";

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
    repo: PathBuf,
    home: PathBuf,
    tmp: PathBuf,
}

impl Fx {
    fn build(&self) -> PathBuf {
        self.tmp.join("build-idle")
    }
    fn wt(&self) -> PathBuf {
        self.root.join("wt-merged")
    }
}

/// A repository with one merged, clean worktree `wt-merged`, and a temporary folder with one
/// tagged build output `build-idle`.
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
    let wt = root.join("wt-merged");
    git(
        &repo,
        &[
            "worktree",
            "add",
            "-q",
            "-b",
            "merged",
            wt.to_str().unwrap(),
        ],
    );
    let b = tmp.join("build-idle");
    fs::create_dir_all(b.join("debug/deps")).unwrap();
    fs::write(b.join("CACHEDIR.TAG"), TAG).unwrap();
    fs::write(b.join("debug/deps/lib.rlib"), vec![7u8; 4096]).unwrap();
    Fx {
        _dir: dir,
        root,
        repo,
        home,
        tmp,
    }
}

/// Every path under `root` with its length (files) or link target (links).
fn snapshot(root: &Path) -> Vec<(PathBuf, u64, Option<PathBuf>)> {
    let mut out = Vec::new();
    let mut stack = vec![root.to_path_buf()];
    while let Some(d) = stack.pop() {
        for e in fs::read_dir(&d).unwrap().flatten() {
            let m = fs::symlink_metadata(e.path()).unwrap();
            out.push((e.path(), m.len(), fs::read_link(e.path()).ok()));
            if m.is_dir() {
                stack.push(e.path());
            }
        }
    }
    out.sort();
    out
}

/// Every regular file under `root`, relative, with its bytes.
fn contents(root: &Path) -> Vec<(PathBuf, Vec<u8>)> {
    let mut out = Vec::new();
    let mut stack = vec![root.to_path_buf()];
    while let Some(d) = stack.pop() {
        for e in fs::read_dir(&d).unwrap().flatten() {
            let m = fs::symlink_metadata(e.path()).unwrap();
            if m.is_dir() {
                stack.push(e.path());
            } else if m.is_file() {
                let rel = e.path().strip_prefix(root).unwrap().to_path_buf();
                out.push((rel, fs::read(e.path()).unwrap()));
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

fn newest_top(d: &Path) -> u64 {
    fs::read_dir(d)
        .unwrap()
        .flatten()
        .map(|e| mtime(&e.path()))
        .fold(mtime(d), u64::max)
}

fn far(fx: &Fx) -> u64 {
    newest_top(&fx.build()) + 10 * 86_400
}

fn run(fx: &Fx, cmd: &str, now: u64, extra: &[&str]) -> Value {
    let out = Command::new(BIN)
        .args([cmd, fx.repo.to_str().unwrap()])
        .args([
            "--integration-ref",
            "main",
            "--tmp",
            fx.tmp.to_str().unwrap(),
        ])
        .args(["--now", &now.to_string()])
        .args(extra)
        .current_dir(&fx.repo)
        .env("HOME", &fx.home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .stdin(Stdio::null())
        .output()
        .unwrap();
    serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        serde_json::json!({
            "ok": false,
            "no_envelope": String::from_utf8_lossy(&out.stderr),
        })
    })
}

fn plan(fx: &Fx, now: u64) -> Value {
    let v = run(fx, "hygiene-plan", now, &[]);
    assert_eq!(v["ok"], true, "{v}");
    v["result"].clone()
}

fn apply(fx: &Fx, now: u64, digest: &str) -> Value {
    run(
        fx,
        "hygiene-apply",
        now,
        &[
            "--plan-digest",
            digest,
            "--permission-reference",
            "fixture-approved",
        ],
    )
}

fn held(r: &Value) -> Vec<String> {
    r["items"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|i| i["action"] == "hold")
        .map(|i| i["path"].as_str().unwrap().to_string())
        .collect()
}

fn worktree_listed(fx: &Fx, wt: &Path) -> bool {
    git(&fx.repo, &["worktree", "list", "--porcelain"])
        .lines()
        .any(|l| l == format!("worktree {}", wt.display()))
}

#[test]
fn a_stale_digest_is_refused_and_nothing_is_written() {
    let fx = fixture();
    let now = far(&fx);
    let digest = plan(&fx, now)["plan_digest"].as_str().unwrap().to_string();
    // A new build output after the plan: the recomputed digest differs.
    let d = fx.tmp.join("build-new");
    fs::create_dir_all(&d).unwrap();
    fs::write(d.join("CACHEDIR.TAG"), TAG).unwrap();
    let before = snapshot(&fx.root);
    let v = apply(&fx, now, &digest);
    assert_eq!(v["ok"], false, "{v}");
    assert!(v.to_string().contains("digest"), "{v}");
    assert_eq!(snapshot(&fx.root), before, "{v}");
    // A wrong digest outright is refused the same way.
    let v = apply(&fx, now, &format!("sha256:{}", "0".repeat(64)));
    assert_eq!(v["ok"], false, "{v}");
    assert_eq!(snapshot(&fx.root), before, "{v}");
}

#[test]
fn held_items_move_whole_into_the_holding_folder_with_a_manifest_line_each() {
    let fx = fixture();
    let now = far(&fx);
    let r = plan(&fx, now);
    assert_eq!(
        held(&r),
        vec![
            fx.build().to_string_lossy().to_string(),
            fx.wt().to_string_lossy().to_string()
        ],
        "{r}"
    );
    let (build_before, wt_before) = (contents(&fx.build()), contents(&fx.wt()));
    assert!(worktree_listed(&fx, &fx.wt()));
    let v = apply(&fx, now, r["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], true, "{v}");
    assert_eq!(v["operation"], "hygiene.apply", "{v}");
    let res = &v["result"];
    assert_eq!(res["writes"], true, "{v}");
    assert_eq!(res["skipped"].as_array().unwrap().len(), 0, "{v}");
    let holding = PathBuf::from(r["holding"].as_str().unwrap());
    assert_eq!(res["holding"], r["holding"], "{v}");
    let manifest = holding.join("manifest.tsv");
    assert_eq!(res["manifest"], manifest.to_string_lossy().as_ref(), "{v}");
    let (hb, hw) = (holding.join("1-build-idle"), holding.join("2-wt-merged"));
    // Originals gone, holding contents byte-identical.
    assert!(!fx.build().exists() && !fx.wt().exists(), "{v}");
    assert_eq!(contents(&hb), build_before);
    assert_eq!(contents(&hw), wt_before);
    // The manifest names exactly what moved, in order.
    let build_bytes = r["items"]
        .as_array()
        .unwrap()
        .iter()
        .find(|i| i["kind"] == "tmp_build")
        .unwrap()["bytes"]
        .as_u64()
        .unwrap();
    let moved = res["moved"].as_array().unwrap();
    assert_eq!(moved.len(), 2, "{v}");
    let wt_bytes = moved[1]["bytes"].as_u64().unwrap();
    assert!(wt_bytes > 0, "{v}");
    let expect = format!(
        "{HEADER}{}\t{}\ttmp_build\t{build_bytes}\tidle_build_output\n{}\t{}\tworktree\t{wt_bytes}\tmerged_clean_idle\n",
        fx.build().display(),
        hb.display(),
        fx.wt().display(),
        hw.display(),
    );
    assert_eq!(fs::read_to_string(&manifest).unwrap(), expect);
    for (m, (orig, hold)) in moved.iter().zip([(fx.build(), &hb), (fx.wt(), &hw)]) {
        assert_eq!(m["original"], orig.to_string_lossy().as_ref(), "{m}");
        assert_eq!(m["holding"], hold.to_string_lossy().as_ref(), "{m}");
    }
    // The moved worktree is pruned from git's list.
    assert!(!worktree_listed(&fx, &fx.wt()), "{v}");
    assert_eq!(res["pruned"], true, "{v}");
}

/// A process that starts working in a held item after the plan: the recomputed plan sees it,
/// so the apply refuses before any write and nothing moves.
#[test]
fn an_item_in_use_after_the_plan_is_not_moved() {
    let fx = fixture();
    let now = far(&fx);
    let digest = plan(&fx, now)["plan_digest"].as_str().unwrap().to_string();
    let mut sleeper = Command::new("/bin/sleep")
        .arg("30")
        .current_dir(fx.build())
        .stdin(Stdio::null())
        .spawn()
        .unwrap();
    let before = snapshot(&fx.root);
    let v = apply(&fx, now, &digest);
    let _ = sleeper.kill();
    let _ = sleeper.wait();
    assert_eq!(v["ok"], false, "{v}");
    assert!(fx.build().exists() && fx.wt().exists(), "{v}");
    assert_eq!(snapshot(&fx.root), before, "{v}");
}

/// A file held open inside a held item by a process that works elsewhere and names it only
/// relatively is invisible to the plan, so the digest matches; the re-check right before the
/// move skips that item and the others still move.
#[test]
fn an_item_with_a_file_held_open_is_skipped_right_before_the_move() {
    let fx = fixture();
    let now = far(&fx);
    let digest = plan(&fx, now)["plan_digest"].as_str().unwrap().to_string();
    let mut holder = Command::new("/bin/sh")
        .args(["-c", "exec 3<CACHEDIR.TAG; cd /; exec /bin/sleep 30"])
        .current_dir(fx.build())
        .stdin(Stdio::null())
        .spawn()
        .unwrap();
    std::thread::sleep(std::time::Duration::from_millis(500));
    let v = apply(&fx, now, &digest);
    let _ = holder.kill();
    let _ = holder.wait();
    assert_eq!(v["ok"], true, "{v}");
    let res = &v["result"];
    let skipped = res["skipped"].as_array().unwrap();
    assert_eq!(skipped.len(), 1, "{v}");
    assert_eq!(skipped[0]["path"], fx.build().to_string_lossy().as_ref());
    assert_eq!(skipped[0]["reason"], "file_open_in_use", "{v}");
    assert!(fx.build().join("CACHEDIR.TAG").exists(), "{v}");
    assert!(!fx.wt().exists(), "{v}");
    let manifest = fs::read_to_string(res["manifest"].as_str().unwrap()).unwrap();
    assert_eq!(manifest.lines().count(), 2, "{manifest}");
    assert!(!manifest.contains("build-idle"), "{manifest}");
}

#[test]
fn a_linked_holding_folder_or_manifest_is_refused() {
    let fx = fixture();
    let now = far(&fx);
    let r = plan(&fx, now);
    let digest = r["plan_digest"].as_str().unwrap().to_string();
    let holding = PathBuf::from(r["holding"].as_str().unwrap());
    let elsewhere = fx.root.join("elsewhere");
    fs::create_dir_all(&elsewhere).unwrap();
    fs::create_dir_all(holding.parent().unwrap()).unwrap();
    std::os::unix::fs::symlink(&elsewhere, &holding).unwrap();
    let before = snapshot(&fx.root);
    let v = apply(&fx, now, &digest);
    assert_eq!(v["ok"], false, "{v}");
    assert!(v.to_string().contains("symlink"), "{v}");
    assert_eq!(snapshot(&fx.root), before, "{v}");
    // A real date folder whose manifest is a link.
    fs::remove_file(&holding).unwrap();
    fs::create_dir_all(&holding).unwrap();
    let other = elsewhere.join("other.tsv");
    fs::write(&other, HEADER).unwrap();
    std::os::unix::fs::symlink(&other, holding.join("manifest.tsv")).unwrap();
    let before = snapshot(&fx.root);
    let v = apply(&fx, now, &digest);
    assert_eq!(v["ok"], false, "{v}");
    assert!(v.to_string().contains("symlink"), "{v}");
    assert_eq!(snapshot(&fx.root), before, "{v}");
    assert_eq!(fs::read_to_string(&other).unwrap(), HEADER);
}

/// The manifest line comes first: when it cannot be written the item does not move.
#[test]
fn an_item_whose_manifest_line_cannot_be_written_does_not_move() {
    let fx = fixture();
    let now = far(&fx);
    let r = plan(&fx, now);
    let holding = PathBuf::from(r["holding"].as_str().unwrap());
    fs::create_dir_all(&holding).unwrap();
    let manifest = holding.join("manifest.tsv");
    fs::write(&manifest, HEADER).unwrap();
    use std::os::unix::fs::PermissionsExt;
    fs::set_permissions(&manifest, fs::Permissions::from_mode(0o444)).unwrap();
    let v = apply(&fx, now, r["plan_digest"].as_str().unwrap());
    fs::set_permissions(&manifest, fs::Permissions::from_mode(0o644)).unwrap();
    assert!(fx.build().exists() && fx.wt().exists(), "{v}");
    assert_eq!(fs::read_to_string(&manifest).unwrap(), HEADER);
    assert!(worktree_listed(&fx, &fx.wt()), "{v}");
}

/// The 2-day age limit, N/N+1, through the apply: one second short of it the build output
/// stays where it is; at exactly 2 days it moves.
#[test]
fn the_two_day_limit_holds_through_the_apply() {
    let fx = fixture();
    let newest = newest_top(&fx.build());
    let limit = 2 * 86_400;
    let r = plan(&fx, newest + limit - 1);
    assert_eq!(held(&r), vec![fx.wt().to_string_lossy().to_string()], "{r}");
    let v = apply(&fx, newest + limit - 1, r["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], true, "{v}");
    assert!(fx.build().exists(), "{v}");
    assert!(!fx.wt().exists(), "{v}");
    let r = plan(&fx, newest + limit);
    assert_eq!(
        held(&r),
        vec![fx.build().to_string_lossy().to_string()],
        "{r}"
    );
    let v = apply(&fx, newest + limit, r["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], true, "{v}");
    assert!(!fx.build().exists(), "{v}");
    let moved = v["result"]["moved"].as_array().unwrap();
    assert_eq!(moved.len(), 1, "{v}");
    assert_eq!(moved[0]["original"], fx.build().to_string_lossy().as_ref());
}
