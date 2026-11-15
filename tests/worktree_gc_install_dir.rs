//! #3010 P2a (leader decision 6895403187, B): an install folder is shown as a fact, never a
//! reason to remove. `worktree-gc-plan` does not plan an installed worktree for removal and
//! `worktree-gc-apply` does not delete one, with or without a tracked lockfile, however large
//! the install, and whatever a person left inside it. Worktrees a host made (Codex) are never
//! removal candidates either. Synthetic repositories in temporary folders only.
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
    // a tracked file in `apps/web` makes git list a nested `apps/web/node_modules/` by its own
    // name; without it git would list just `apps/`
    write(&root.join("apps/web/package.json"), "{}\n");
    git(
        &root,
        &["add", ".gitignore", "a.txt", "apps/web/package.json"],
    );
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
    write(&wt.join("node_modules/pkg/package.json"), "{}\n");
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

/// `worktree-gc-apply` with the digest of `r`, on the same synthetic clock as `plan`.
fn apply(fx: &Fx, r: &Value) -> Value {
    let digest = r["plan_digest"].as_str().unwrap().to_string();
    let now = later().to_string();
    run(
        fx,
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
    )
}

/// The decision: an install folder is a fact in the preview, never a reason to remove. A
/// merged worktree with one is kept however small or large the install is and whatever a
/// person left in it, and `worktree-gc-apply` leaves it and the person's files in place. A
/// merged worktree without one is still a `remove` candidate and is removed, so the test
/// tells the two apart.
#[test]
fn an_installed_worktree_is_a_preview_fact_and_is_never_removed() {
    use std::os::unix::fs::PermissionsExt;
    let names = ["plain", "small", "human", "big", "locked"];
    let fx = fixture(true, &names);
    install(&fx.wt("small"));
    install(&fx.wt("human"));
    write(&fx.wt("human").join("node_modules/.env"), "SECRET=1\n");
    write(&fx.wt("human").join("node_modules/my-notes.md"), "mine\n");
    write(&fx.wt("human").join("node_modules/data.sqlite"), "rows\n");
    install(&fx.wt("big"));
    let pkg = fx.wt("big").join("node_modules/pkg");
    for i in 0..20_001 {
        fs::write(pkg.join(format!("f{i}.js")), "x").unwrap();
    }
    install(&fx.wt("locked"));
    let unreadable = fx.wt("locked").join("node_modules/pkg");
    fs::set_permissions(&unreadable, fs::Permissions::from_mode(0o000)).unwrap();

    let r = plan(&fx);
    assert_eq!(row(&r, "plain")["action"], "remove", "{r}");
    for n in ["small", "human", "big", "locked"] {
        let c = row(&r, n);
        assert_ne!(c["action"], "remove", "{n}: {c}");
        // the facts stay, and they do not claim the worktree is removable
        assert_eq!(c["lifecycle"]["install_dir"], true, "{n}: {c}");
    }
    let v = apply(&fx, &r);
    fs::set_permissions(&unreadable, fs::Permissions::from_mode(0o755)).unwrap();
    assert_eq!(v["ok"], true, "{v}");
    assert!(!fx.wt("plain").exists());
    for n in ["small", "human", "big", "locked"] {
        assert!(fx.wt(n).join("node_modules").exists(), "{n}");
    }
    assert!(fx.wt("human").join("node_modules/.env").exists());
    assert!(fx.wt("human").join("node_modules/my-notes.md").exists());
    assert!(fx.wt("human").join("node_modules/data.sqlite").exists());
}

/// Without a tracked lockfile nothing says the install can be rebuilt: the worktree is kept
/// as well (the fact `lockfile tracked` is false).
#[test]
fn an_install_without_a_tracked_lockfile_keeps_the_worktree() {
    let fx = fixture(false, &["inst"]);
    install(&fx.wt("inst"));
    let r = plan(&fx);
    let c = row(&r, "inst");
    assert_ne!(c["action"], "remove", "{c}");
    assert_eq!(c["lifecycle"]["install_lockfile_tracked"], false, "{c}");
}

/// A symlinked `node_modules` (it may point anywhere) keeps the worktree and what it points at
/// is never touched.
#[test]
fn a_symlinked_node_modules_keeps_the_worktree_and_its_target() {
    let fx = fixture(true, &["linked"]);
    let outside = tempfile::tempdir().unwrap();
    write(&outside.path().join("keep.txt"), "mine\n");
    std::os::unix::fs::symlink(outside.path(), fx.wt("linked").join("node_modules")).unwrap();
    let r = plan(&fx);
    assert_ne!(row(&r, "linked")["action"], "remove", "{r}");
    let v = apply(&fx, &r);
    assert_eq!(v["ok"], true, "{v}");
    assert!(fx.wt("linked").exists());
    assert!(outside.path().join("keep.txt").exists());
}

/// Audit B3 on #3058 (user decision 6892859628): a worktree Codex made is a preview row and is
/// never removed, with or without an install folder.
#[test]
fn a_codex_made_worktree_is_never_a_remove_candidate() {
    let fx = fixture(true, &[]);
    for (id, installed) in [("ab12", true), ("cd34", false)] {
        let wt = fx.home.join(format!(".codex/worktrees/{id}/repo"));
        fs::create_dir_all(wt.parent().unwrap()).unwrap();
        git(
            &fx.root,
            &[
                "worktree",
                "add",
                "-q",
                "-b",
                &format!("codex/{id}"),
                wt.to_str().unwrap(),
            ],
        );
        fs::create_dir_all(wt.join("target/debug")).unwrap();
        fs::write(
            wt.join("target/CACHEDIR.TAG"),
            "Signature: 8a477f597d28d172789f06886806bc55\n",
        )
        .unwrap();
        if installed {
            install(&wt);
        }
    }
    let r = plan(&fx);
    let rows: Vec<&Value> = r["candidates"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|c| c["location"] == "codex")
        .collect();
    assert_eq!(rows.len(), 2, "{r}");
    for c in rows {
        assert_eq!(c["action"], "skip", "{c}");
        assert!(
            reasons(c).contains(&"preserved_location".to_string()),
            "{c}"
        );
    }
    let v = apply(&fx, &r);
    assert_eq!(v["ok"], true, "{v}");
    for id in ["ab12", "cd34"] {
        assert!(fx.home.join(format!(".codex/worktrees/{id}/repo")).exists());
    }
}
