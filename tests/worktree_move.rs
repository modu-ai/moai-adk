//! #2723: `moai-cli worktree-move-plan` / `worktree-move-apply` on synthetic repositories in
//! a temporary folder with a temporary HOME only. No real worktree, branch or report of this
//! machine is moved or deleted here.
#![cfg(unix)]
mod common;

use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output};

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
    base: PathBuf,
    repo: PathBuf,
    home: PathBuf,
    root: PathBuf,
}

impl Fixture {
    fn wt(&self, name: &str) -> PathBuf {
        self.repo.join(".moai-cli/worktrees").join(name)
    }
    fn obs(&self, holds: Value) -> PathBuf {
        let p = self.base.join("obs.json");
        fs::write(
            &p,
            serde_json::to_vec(&json!({"schema_version": 1, "holds": holds})).unwrap(),
        )
        .unwrap();
        p
    }
}

/// Marks the current HEAD of `dir` as pushed (a remote-tracking ref, no network).
fn pushed(dir: &Path, name: &str) {
    git(
        dir,
        &["update-ref", &format!("refs/remotes/origin/{name}"), "HEAD"],
    );
}

fn add(fx_repo: &Path, name: &str) -> PathBuf {
    let wt = fx_repo.join(".moai-cli/worktrees").join(name);
    let branch = format!("issue/1-{name}");
    git(
        fx_repo,
        &["worktree", "add", "-q", "-b", &branch, wt.to_str().unwrap()],
    );
    wt
}

/// One clean worktree per blocking reason, plus `clean` that may move.
fn fixture(names: &[&str]) -> Fixture {
    let dir = tempfile::TempDir::new().unwrap();
    let base = fs::canonicalize(dir.path()).unwrap();
    let repo = base.join("proj");
    let home = base.join("home");
    let root = home.join("common-root");
    fs::create_dir_all(&repo).unwrap();
    fs::create_dir_all(&home).unwrap();
    git(&repo, &["init", "-q", "-b", "main"]);
    fs::write(repo.join(".gitignore"), ".moai-cli/\ntarget/\n").unwrap();
    fs::write(repo.join("a.txt"), "a\n").unwrap();
    git(&repo, &["add", ".gitignore", "a.txt"]);
    git(&repo, &["commit", "-q", "-m", "init"]);
    pushed(&repo, "main");
    for n in names {
        add(&repo, n);
    }
    Fixture {
        _dir: dir,
        base,
        repo,
        home,
        root,
    }
}

fn run(fx: &Fixture, args: &[&str]) -> Output {
    Command::new(BIN)
        .args(args)
        .current_dir(&fx.repo)
        .env("HOME", &fx.home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .output()
        .unwrap()
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "envelope on stdout; stderr {}",
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

fn plan(fx: &Fixture, obs: Option<&Path>) -> Value {
    let repo = fx.repo.to_str().unwrap().to_string();
    let root = fx.root.to_str().unwrap().to_string();
    let mut args = vec!["worktree-move-plan", repo.as_str(), "--root", root.as_str()];
    let o;
    if let Some(p) = obs {
        o = p.to_str().unwrap().to_string();
        args.extend(["--observations", o.as_str()]);
    }
    let out = run(fx, &args);
    let v = envelope(&out);
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn apply(fx: &Fixture, obs: &Path, digest: &str, permission: &str) -> Value {
    envelope(&run(
        fx,
        &[
            "worktree-move-apply",
            fx.repo.to_str().unwrap(),
            "--root",
            fx.root.to_str().unwrap(),
            "--observations",
            obs.to_str().unwrap(),
            "--plan-digest",
            digest,
            "--permission-reference",
            permission,
        ],
    ))
}

fn item<'a>(plan: &'a Value, path: &Path) -> &'a Value {
    let p = fs::canonicalize(path).unwrap_or_else(|_| path.to_path_buf());
    plan["items"]
        .as_array()
        .unwrap()
        .iter()
        .find(|it| it["path"] == json!(p.display().to_string()))
        .unwrap_or_else(|| panic!("no item for {} in {plan}", p.display()))
}

fn reasons(plan: &Value, path: &Path) -> Vec<String> {
    item(plan, path)["reasons"]
        .as_array()
        .unwrap()
        .iter()
        .map(|r| r.as_str().unwrap().to_string())
        .collect()
}

fn tree_digest(dir: &Path) -> Vec<(String, Vec<u8>)> {
    let mut out = Vec::new();
    let mut stack = vec![dir.to_path_buf()];
    while let Some(d) = stack.pop() {
        for e in fs::read_dir(&d).unwrap().flatten() {
            let p = e.path();
            if p.file_name().is_some_and(|n| n == "index") {
                continue;
            }
            if e.file_type().unwrap().is_dir() {
                stack.push(p);
            } else {
                out.push((p.display().to_string(), fs::read(&p).unwrap_or_default()));
            }
        }
    }
    out.sort();
    out
}

#[test]
fn every_blocking_reason_is_named_and_only_a_clean_worktree_moves() {
    let fx = fixture(&[
        "clean", "dirty", "untr", "ahead", "locky", "held", "pend", "stashy", "subby",
    ]);
    for n in ["clean", "dirty", "untr", "locky", "held", "pend", "stashy"] {
        pushed(&fx.wt(n), &format!("issue/1-{n}"));
    }
    fs::write(fx.wt("dirty").join("a.txt"), "changed\n").unwrap();
    fs::write(fx.wt("untr").join("new.txt"), "new\n").unwrap();
    fs::write(fx.wt("ahead").join("b.txt"), "b\n").unwrap();
    git(&fx.wt("ahead"), &["add", "b.txt"]);
    git(&fx.wt("ahead"), &["commit", "-q", "-m", "ahead"]);
    git(
        &fx.repo,
        &["worktree", "lock", fx.wt("locky").to_str().unwrap()],
    );
    fs::write(fx.wt("stashy").join("a.txt"), "stash me\n").unwrap();
    git(&fx.wt("stashy"), &["stash", "push", "-q", "-m", "wip"]);
    fs::write(
        fx.wt("subby").join(".gitmodules"),
        "[submodule \"x\"]\n\tpath = x\n\turl = ../x\n",
    )
    .unwrap();
    git(&fx.wt("subby"), &["add", ".gitmodules"]);
    git(&fx.wt("subby"), &["commit", "-q", "-m", "submodule file"]);
    pushed(&fx.wt("subby"), "issue/1-subby");

    let obs = fx.obs(json!({"source": "fixture holds", "branches": [
        {"branch": "issue/1-held", "reason": "hold"},
        {"branch": "issue/1-pend", "reason": "handoff_pending"}
    ]}));
    let p = plan(&fx, Some(&obs));
    assert_eq!(item(&p, &fx.repo)["action"], json!("blocked"));
    assert_eq!(reasons(&p, &fx.repo), ["main_worktree"]);
    assert_eq!(item(&p, &fx.wt("clean"))["action"], json!("move"), "{p}");
    assert_eq!(reasons(&p, &fx.wt("clean")), Vec::<String>::new());
    for (n, want) in [
        ("dirty", "uncommitted"),
        ("untr", "untracked"),
        ("ahead", "unpushed"),
        ("locky", "locked"),
        ("held", "hold"),
        ("pend", "handoff_pending"),
        ("stashy", "stash"),
        ("subby", "has_submodules"),
    ] {
        assert_eq!(item(&p, &fx.wt(n))["action"], json!("blocked"), "{n}: {p}");
        assert_eq!(reasons(&p, &fx.wt(n)), [want], "{n}");
    }
    assert_eq!(p["totals"], json!({"move": 1, "keep": 0, "blocked": 9}));
    assert_eq!(p["state"], json!("changes"));
    let key = p["project_key"].as_str().unwrap();
    assert!(
        key.starts_with("proj-") && key.len() == "proj-".len() + 12,
        "{key}"
    );
    assert_eq!(
        item(&p, &fx.wt("clean"))["destination"],
        json!(fx.root.join(key).join("clean").display().to_string())
    );
    assert_eq!(p["writes"], json!(false));
    assert!(!fx.root.exists(), "the plan creates nothing");

    // Without the holds observation the hold check has not run: nothing may move.
    let p0 = plan(&fx, None);
    assert_eq!(reasons(&p0, &fx.wt("clean")), ["hold_check_not_run"]);
    assert_eq!(p0["totals"]["move"], json!(0));
    assert_eq!(p0["state"], json!("unchanged"));
    // An empty list is an observation that nothing is held.
    let none_held = fx.obs(json!({"source": "fixture", "branches": []}));
    assert_eq!(
        reasons(&plan(&fx, Some(&none_held)), &fx.wt("held")),
        Vec::<String>::new()
    );
}

#[test]
fn apply_moves_on_the_same_digest_then_rerun_keeps_and_revert_restores() {
    let fx = fixture(&["clean", "dirty"]);
    pushed(&fx.wt("clean"), "issue/1-clean");
    pushed(&fx.wt("dirty"), "issue/1-dirty");
    fs::write(fx.wt("dirty").join("a.txt"), "changed\n").unwrap();
    let obs = fx.obs(json!({"source": "fixture", "branches": []}));
    let p = plan(&fx, Some(&obs));
    let digest = p["plan_digest"].as_str().unwrap().to_string();
    let head = item(&p, &fx.wt("clean"))["head"].clone();
    let dest = PathBuf::from(item(&p, &fx.wt("clean"))["destination"].as_str().unwrap());
    let src = fs::canonicalize(fx.wt("clean")).unwrap();

    let a = apply(&fx, &obs, &digest, "user approved moves in the fixture");
    assert_eq!(a["ok"], json!(true), "{a}");
    let r = &a["result"];
    assert_eq!(r["state"], json!("applied"), "{r}");
    assert_eq!(r["results"].as_array().unwrap().len(), 1);
    let moved = &r["results"][0];
    assert_eq!(moved["status"], json!("moved"));
    assert!(!src.exists() && dest.join("a.txt").exists());
    let list = git(&fx.repo, &["worktree", "list", "--porcelain"]);
    assert!(
        list.contains(&format!("worktree {}", dest.display())),
        "{list}"
    );
    assert_eq!(json!(git(&dest, &["rev-parse", "HEAD"]).trim()), head);
    assert_eq!(
        git(&dest, &["branch", "--show-current"]).trim(),
        "issue/1-clean"
    );
    // The dirty worktree was not touched.
    assert_eq!(
        fs::read_to_string(fx.wt("dirty").join("a.txt")).unwrap(),
        "changed\n"
    );

    // Re-run: the moved worktree is `keep`; nothing is left to move.
    let again = plan(&fx, Some(&obs));
    assert_eq!(item(&again, &dest)["action"], json!("keep"));
    assert_eq!(again["state"], json!("unchanged"));
    let a2 = apply(&fx, &obs, again["plan_digest"].as_str().unwrap(), "ok");
    assert_eq!(a2["result"]["state"], json!("unchanged"), "{a2}");

    // The way back the answer names puts it where it was.
    let revert = moved["revert"].as_str().unwrap();
    let parts: Vec<&str> = revert.split(' ').collect();
    assert_eq!(&parts[..2], ["git", "-C"]);
    git(Path::new(parts[2]), &parts[3..]);
    assert!(src.join("a.txt").exists() && !dest.exists());
    assert_eq!(item(&plan(&fx, Some(&obs)), &src)["action"], json!("move"));
}

#[test]
fn a_stale_digest_or_missing_permission_writes_nothing() {
    let fx = fixture(&["clean"]);
    pushed(&fx.wt("clean"), "issue/1-clean");
    let obs = fx.obs(json!({"source": "fixture", "branches": []}));
    let p = plan(&fx, Some(&obs));
    let digest = p["plan_digest"].as_str().unwrap().to_string();
    let before = tree_digest(&fx.repo.join(".git"));

    let a = apply(&fx, &obs, "sha256:0000", "approved");
    assert_eq!(a["ok"], json!(false), "{a}");
    assert!(a.to_string().contains("stale or altered plan"), "{a}");
    for perm in ["", "placeholder"] {
        let a = apply(&fx, &obs, &digest, perm);
        assert_eq!(a["ok"], json!(false), "{perm}: {a}");
    }
    // A change after the preview (a new untracked file) changes the digest.
    fs::write(fx.wt("clean").join("late.txt"), "late\n").unwrap();
    let a = apply(&fx, &obs, &digest, "approved");
    assert!(a.to_string().contains("stale or altered plan"), "{a}");
    assert!(!fx.root.exists(), "no folder was created");
    assert!(fx.wt("clean").join("a.txt").exists());
    assert_eq!(tree_digest(&fx.repo.join(".git")), before);
}

#[test]
fn a_process_that_starts_inside_after_the_preview_stops_the_move() {
    let fx = fixture(&["clean"]);
    pushed(&fx.wt("clean"), "issue/1-clean");
    let obs = fx.obs(json!({"source": "fixture", "branches": []}));
    let p = plan(&fx, Some(&obs));
    let digest = p["plan_digest"].as_str().unwrap().to_string();
    // Working inside it: the recomputed plan blocks it (`in_use`), so the digest differs.
    let mut child = Command::new("sleep")
        .arg("30")
        .current_dir(fx.wt("clean"))
        .spawn()
        .unwrap();
    let a = apply(&fx, &obs, &digest, "approved");
    let _ = child.kill();
    let _ = child.wait();
    assert_eq!(a["error"]["kind"], json!("stale_plan"), "{a}");
    assert!(fx.wt("clean").join("a.txt").exists() && !fx.root.exists());

    // Only a file held open (the preview reads working folders, not open files): the
    // re-check right before the move refuses it.
    let held = fs::File::open(fx.wt("clean").join("a.txt")).unwrap();
    let mut child = Command::new("sleep")
        .arg("30")
        .current_dir(&fx.base)
        .stdin(held)
        .spawn()
        .unwrap();
    let a = apply(&fx, &obs, &digest, "approved");
    let _ = child.kill();
    let _ = child.wait();
    let r = &a["result"];
    assert_eq!(r["state"], json!("partial"), "{a}");
    assert_eq!(r["results"][0]["status"], json!("refused"));
    assert!(
        r["results"][0]["detail"]
            .as_str()
            .unwrap()
            .contains("file_open_in_use"),
        "{r}"
    );
    assert_eq!(r["results"][0]["revert"], Value::Null);
    assert!(fx.wt("clean").join("a.txt").exists());
}

#[test]
fn the_root_is_checked_and_a_taken_destination_blocks() {
    let fx = fixture(&["clean"]);
    pushed(&fx.wt("clean"), "issue/1-clean");
    let obs = fx.obs(json!({"source": "fixture", "branches": []}));
    let repo = fx.repo.to_str().unwrap();
    for (root, why) in [
        ("relative/root", "absolute"),
        ("/tmp/../x", "`.` or `..`"),
        (
            fx.repo.join("inside").to_str().unwrap(),
            "inside the worktree",
        ),
        (
            fx.wt("clean").join("deeper").to_str().unwrap(),
            "inside the worktree",
        ),
    ] {
        let v = envelope(&run(&fx, &["worktree-move-plan", repo, "--root", root]));
        assert_eq!(v["ok"], json!(false), "{root}");
        assert!(v.to_string().contains(why), "{root}: {v}");
    }
    let key = plan(&fx, Some(&obs))["project_key"]
        .as_str()
        .unwrap()
        .to_string();
    fs::create_dir_all(fx.root.join(&key).join("clean")).unwrap();
    assert_eq!(
        reasons(&plan(&fx, Some(&obs)), &fx.wt("clean")),
        ["destination_exists"]
    );
    // An observation with an unknown reason is refused, not ignored.
    let bad = fx.obs(json!({"source": "x", "branches": [{"branch": "b", "reason": "maybe"}]}));
    let v = envelope(&run(
        &fx,
        &[
            "worktree-move-plan",
            repo,
            "--root",
            fx.root.to_str().unwrap(),
            "--observations",
            bad.to_str().unwrap(),
        ],
    ));
    assert_eq!(v["ok"], json!(false), "{v}");
}

#[test]
fn two_projects_with_one_name_get_two_keys() {
    let a = fixture(&[]);
    let b = fixture(&[]);
    let ka = plan(&a, None)["project_key"].as_str().unwrap().to_string();
    let kb = plan(&b, None)["project_key"].as_str().unwrap().to_string();
    assert!(ka.starts_with("proj-") && kb.starts_with("proj-"));
    assert_ne!(ka, kb);
    assert_eq!(
        plan(&a, None)["project_key"],
        json!(ka),
        "stable across runs"
    );
}

#[test]
fn a_worktree_moved_by_hand_is_missing_until_git_repairs_it() {
    let fx = fixture(&["clean"]);
    pushed(&fx.wt("clean"), "issue/1-clean");
    let elsewhere = fx.base.join("moved-by-hand");
    fs::rename(fx.wt("clean"), &elsewhere).unwrap();
    let p = plan(&fx, None);
    assert_eq!(reasons(&p, &fx.wt("clean")), ["missing"]);
    git(&elsewhere, &["worktree", "repair"]);
    let obs = fx.obs(json!({"source": "fixture", "branches": []}));
    let p2 = plan(&fx, Some(&obs));
    assert_eq!(item(&p2, &elsewhere)["action"], json!("move"), "{p2}");
}

#[test]
fn two_worktrees_with_one_name_block_each_other() {
    let fx = fixture(&["same"]);
    pushed(&fx.wt("same"), "issue/1-same");
    let other = fx.repo.join(".moai-cli/elsewhere/same");
    git(
        &fx.repo,
        &[
            "worktree",
            "add",
            "-q",
            "-b",
            "issue/2-same",
            other.to_str().unwrap(),
        ],
    );
    pushed(&other, "issue/2-same");
    let obs = fx.obs(json!({"source": "fixture", "branches": []}));
    let p = plan(&fx, Some(&obs));
    assert_eq!(
        reasons(&p, &fx.wt("same")),
        ["destination_collision"],
        "{p}"
    );
    assert_eq!(reasons(&p, &other), ["destination_collision"]);
    assert_eq!(p["totals"]["move"], json!(0));
}
