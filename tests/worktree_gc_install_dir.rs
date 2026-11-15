//! #3010 P2a: `worktree-gc-plan` and `worktree-gc-apply` treat a real top-level `node_modules`
//! as rebuildable install output when a lockfile is tracked, so an installed, merged worktree
//! is a `remove` candidate (it used to be `skip` for ever: one install is far past the
//! ignored-file census limit). Anything the plan cannot vouch for keeps the worktree.
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

impl Fx {
    fn wt(&self, n: &str) -> PathBuf {
        self.root.join(".moai-cli/worktrees").join(n)
    }
}

fn write(p: &Path, body: &str) {
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::write(p, body).unwrap();
}

/// A primary checkout ignoring `node_modules/` (and `notes.md`), with `lockfile` committed
/// or not, and merged worktrees `names`, each holding cargo output so none is the empty kind.
fn fixture(lockfile: bool, names: &[&str]) -> Fx {
    let dir = tempfile::TempDir::new().unwrap();
    let base = fs::canonicalize(dir.path()).unwrap();
    let root = base.join("repo");
    let home = base.join("home");
    fs::create_dir_all(&root).unwrap();
    fs::create_dir_all(&home).unwrap();
    git(&root, &["init", "-q", "-b", "main"]);
    fs::write(
        root.join(".gitignore"),
        ".moai-cli/\ntarget/\nnode_modules/\nnotes.md\n",
    )
    .unwrap();
    fs::write(root.join("a.txt"), "a\n").unwrap();
    git(&root, &["add", ".gitignore", "a.txt"]);
    if lockfile {
        fs::write(root.join("package-lock.json"), "{}\n").unwrap();
        git(&root, &["add", "package-lock.json"]);
    }
    git(&root, &["commit", "-q", "-m", "init"]);
    git(&root, &["branch", "integ"]);
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

/// A small but real-looking install: nested packages, a `.bin` link into a package, a
/// dotfile at the top.
fn install(wt: &Path) {
    write(
        &wt.join("node_modules/pkg/index.js"),
        "module.exports = 1;\n",
    );
    write(&wt.join("node_modules/@scope/lib/package.json"), "{}\n");
    write(&wt.join("node_modules/.package-lock.json"), "{}\n");
    fs::create_dir_all(wt.join("node_modules/.bin")).unwrap();
    std::os::unix::fs::symlink("../pkg/index.js", wt.join("node_modules/.bin/tool")).unwrap();
}

fn later() -> u64 {
    std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap()
        .as_secs()
        + 86_400
}

fn run(fx: &Fx, args: &[&str]) -> Value {
    let out = Command::new(BIN)
        .args(args)
        .current_dir(&fx.root)
        .env("HOME", &fx.home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .output()
        .unwrap();
    serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)))
}

fn plan(fx: &Fx) -> Value {
    let now = later().to_string();
    let v = run(
        fx,
        &[
            "worktree-gc-plan",
            fx.root.to_str().unwrap(),
            "--integration-ref",
            "integ",
            "--now-unix",
            &now,
        ],
    );
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

fn row<'a>(r: &'a Value, n: &str) -> &'a Value {
    entry(r, &format!(".moai-cli/worktrees/{n}"))
}

/// With a tracked lockfile a real `node_modules` — nested packages, a dotfile and a `.bin`
/// link — does not keep a merged worktree; the install counts as safe to lose.
#[test]
fn an_install_with_a_tracked_lockfile_does_not_keep_a_merged_worktree() {
    let fx = fixture(true, &["inst"]);
    install(&fx.wt("inst"));
    let r = plan(&fx);
    let c = row(&r, "inst");
    assert_eq!(c["action"], "remove", "{c}");
    assert_eq!(c["ignored_outside_build"], 0, "{c}");
    assert!(c["ignored_safe"].as_u64().unwrap() >= 7, "{c}");
}

/// Without a tracked lockfile nothing says the install can be rebuilt: the files are kept.
#[test]
fn an_install_without_a_tracked_lockfile_keeps_the_worktree() {
    let fx = fixture(false, &["inst"]);
    install(&fx.wt("inst"));
    let r = plan(&fx);
    let c = row(&r, "inst");
    assert_ne!(c["action"], "remove", "{c}");
    assert!(c["ignored_outside_build"].as_u64().unwrap() >= 1, "{c}");
}

/// What the install proof does not cover keeps the worktree: a symlinked `node_modules`
/// (it may point anywhere), another ignored file next to it, a nested `node_modules`, a git
/// dependency's own `.git`.
#[test]
fn what_the_install_proof_does_not_cover_keeps_the_worktree() {
    let fx = fixture(true, &["linked", "beside", "nested", "gitdep"]);
    let outside = tempfile::tempdir().unwrap();
    write(&outside.path().join("keep.txt"), "mine\n");
    std::os::unix::fs::symlink(outside.path(), fx.wt("linked").join("node_modules")).unwrap();
    install(&fx.wt("beside"));
    write(&fx.wt("beside").join("notes.md"), "my notes\n");
    write(&fx.wt("nested").join("apps/web/node_modules/p/i.js"), "x\n");
    install(&fx.wt("gitdep"));
    write(&fx.wt("gitdep").join("node_modules/dep/.git/HEAD"), "ref\n");
    let r = plan(&fx);
    for n in ["linked", "beside", "nested", "gitdep"] {
        let c = row(&r, n);
        assert_ne!(c["action"], "remove", "{n}: {c}");
        if n == "linked" {
            // `node_modules/` in .gitignore matches folders only, so the link is an untracked
            // file: it keeps the worktree through the untracked census instead.
            assert_eq!(reasons(c), ["untracked_files"], "{n}: {c}");
        } else {
            assert!(
                c["ignored_outside_build"].as_u64().unwrap() >= 1,
                "{n}: {c}"
            );
        }
    }
    assert!(outside.path().join("keep.txt").exists());
}

/// `worktree-gc-apply` removes the installed worktree whole, including a link inside
/// `node_modules` that points out of it: the link goes, what it pointed at stays, and the
/// branch is kept.
#[test]
fn apply_removes_an_installed_worktree_and_never_follows_a_link_out() {
    let fx = fixture(true, &["inst"]);
    install(&fx.wt("inst"));
    let outside = tempfile::tempdir().unwrap();
    write(&outside.path().join("keep.txt"), "mine\n");
    std::os::unix::fs::symlink(outside.path(), fx.wt("inst").join("node_modules/escape")).unwrap();
    let r = plan(&fx);
    assert_eq!(row(&r, "inst")["action"], "remove", "{r}");
    let digest = r["plan_digest"].as_str().unwrap().to_string();
    let now = later().to_string();
    let v = run(
        &fx,
        &[
            "worktree-gc-apply",
            fx.root.to_str().unwrap(),
            "--integration-ref",
            "integ",
            "--now-unix",
            &now,
            "--plan-digest",
            &digest,
            "--permission-reference",
            "fixture-approved",
        ],
    );
    assert_eq!(v["ok"], true, "{v}");
    let res = &v["result"]["results"];
    let inst = res
        .as_array()
        .unwrap()
        .iter()
        .find(|x| x["path"] == ".moai-cli/worktrees/inst")
        .unwrap_or_else(|| panic!("{v}"));
    assert_eq!(inst["status"], "removed", "{v}");
    assert!(!fx.wt("inst").exists());
    // the link was removed as a link; what it pointed at is untouched
    assert!(outside.path().join("keep.txt").exists());
    // git kept the branch
    assert!(git(&fx.root, &["branch", "--list", "issue/1-inst"]).contains("issue/1-inst"));
}
