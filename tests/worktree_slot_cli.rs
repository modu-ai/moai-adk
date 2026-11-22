//! SPEC-20261123-worktree-slots slice 1: `worktree-slot plan|ensure`. Worktrees are fixed slots
//! `.claude/worktrees/lane-01..10`, `audit-1..3` and `seed`; a new slot is added detached at a
//! commit and, when it has no `target/` while the seed has one, warmed with an APFS clone of
//! the seed's `target/`; a slot is never added past the worktree cap or under the free-disk
//! floor. The binary runs on temporary repositories in a clean environment.
mod common;

use common::git::git;
use common::CleanEnv;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Command;
use std::time::{Duration, SystemTime};

const SLOTS: [&str; 14] = [
    "lane-01", "lane-02", "lane-03", "lane-04", "lane-05", "lane-06", "lane-07", "lane-08",
    "lane-09", "lane-10", "audit-1", "audit-2", "audit-3", "seed",
];

struct Repo {
    _dir: tempfile::TempDir,
    root: PathBuf,
    head: String,
}

/// A repository with one commit, its files given by `files`.
fn repo(files: &[(&str, &str)]) -> Repo {
    let dir = tempfile::tempdir().unwrap();
    let root = fs::canonicalize(dir.path()).unwrap().join("repo");
    fs::create_dir_all(&root).unwrap();
    git(&root, &["init", "-q", "-b", "main"]);
    for (rel, text) in files {
        let p = root.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, text).unwrap();
    }
    fs::write(root.join(".gitignore"), "/target/\n/.claude/\n").unwrap();
    git(&root, &["add", "-A"]);
    git(
        &root,
        &[
            "-c",
            "user.name=t",
            "-c",
            "user.email=t@e",
            "commit",
            "-q",
            "-m",
            "one",
        ],
    );
    let head = git(&root, &["rev-parse", "HEAD"]);
    Repo {
        _dir: dir,
        root,
        head,
    }
}

impl Repo {
    fn slot(&self, name: &str) -> PathBuf {
        self.root.join(".claude/worktrees").join(name)
    }

    fn run(&self, args: &[&str]) -> (i32, Value) {
        let out = Command::new(common::MOAI_BIN)
            .clean_env()
            .current_dir(&self.root)
            .arg("worktree-slot")
            .args(args)
            .args(["--project", self.root.to_str().unwrap()])
            .output()
            .unwrap();
        let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
            panic!(
                "{args:?}: not JSON: {} / {}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        });
        (out.status.code().unwrap_or(-1), v)
    }

    fn ok(&self, args: &[&str]) -> Value {
        let (code, v) = self.run(args);
        assert_eq!((code, &v["ok"]), (0, &json!(true)), "{args:?}: {v}");
        v["result"].clone()
    }

    fn worktrees(&self) -> String {
        git(&self.root, &["worktree", "list", "--porcelain"])
    }
}

// ---- plan (AC-1: REQ-1, REQ-2)

#[test]
fn plan_lists_the_fourteen_slots_in_order_and_writes_nothing() {
    let r = repo(&[("a.txt", "a\n")]);
    let before = r.worktrees();
    let plan = r.ok(&["plan"]);
    let names: Vec<&str> = plan["slots"]
        .as_array()
        .unwrap()
        .iter()
        .map(|s| s["name"].as_str().unwrap())
        .collect();
    assert_eq!(names, SLOTS);
    for s in plan["slots"].as_array().unwrap() {
        assert_eq!(s["exists"], json!(false), "{s}");
        assert_eq!(
            s["path"],
            json!(format!(".claude/worktrees/{}", s["name"].as_str().unwrap())),
            "{s}"
        );
    }
    assert_eq!(plan["caps"]["worktrees"]["limit"], json!(25), "{plan}");
    assert_eq!(plan["caps"]["worktrees"]["count"], json!(0), "{plan}");
    assert_eq!(plan["caps"]["free_percent"]["limit"], json!(15), "{plan}");
    assert!(plan["caps"]["free_percent"]["value"].is_number(), "{plan}");
    assert_eq!(plan["caps"]["slots_gb"]["limit"], json!(300), "{plan}");
    assert_eq!(plan["caps"]["slots_gb"]["measured"], json!(false), "{plan}");
    assert_eq!(r.worktrees(), before);
    assert!(!r.root.join(".claude").exists());
}

#[test]
fn ensure_refuses_a_name_that_is_not_a_fixed_slot() {
    let r = repo(&[("a.txt", "a\n")]);
    for bad in [
        "lane-11",
        "lane-1",
        "audit-4",
        "seed2",
        "../x",
        "lane-01/x",
        "",
    ] {
        let (code, v) = r.run(&["ensure", bad, "--at", &r.head]);
        assert_ne!(code, 0, "{bad:?}: {v}");
        assert_eq!(v["error"]["kind"], json!("invalid_input"), "{bad:?}: {v}");
        assert!(
            v["error"]["message"]
                .as_str()
                .unwrap()
                .starts_with("invalid slot name:"),
            "{bad:?}: {v}"
        );
    }
    assert!(!r.root.join(".claude").exists());
}

// ---- ensure (AC-2: REQ-3, REQ-4, REQ-6)

fn mtime(p: &Path) -> SystemTime {
    fs::metadata(p).unwrap().modified().unwrap()
}

#[test]
fn ensure_adds_a_detached_slot_warms_it_from_the_seed_and_leaves_it_alone_after() {
    let r = repo(&[("a.txt", "a\n")]);
    // A seed with a target/ whose file carries an old modification time.
    let seed = r.ok(&["ensure", "seed", "--at", &r.head]);
    assert_eq!(
        (&seed["added"], &seed["warmed"]),
        (&json!(true), &json!("none")),
        "{seed}"
    );
    let built = r.slot("seed").join("target/debug/built.txt");
    fs::create_dir_all(built.parent().unwrap()).unwrap();
    fs::write(&built, "seed build\n").unwrap();
    let old = SystemTime::now() - Duration::from_secs(3 * 86_400);
    fs::File::options()
        .write(true)
        .open(&built)
        .unwrap()
        .set_modified(old)
        .unwrap();

    let lane = r.ok(&["ensure", "lane-01", "--at", &r.head]);
    assert_eq!(lane["added"], json!(true), "{lane}");
    let method = if cfg!(target_os = "macos") {
        "clone"
    } else {
        "copy"
    };
    assert_eq!(lane["warmed"], json!(method), "{lane}");
    let copied = r.slot("lane-01").join("target/debug/built.txt");
    assert_eq!(fs::read_to_string(&copied).unwrap(), "seed build\n");
    assert_eq!(mtime(&copied), mtime(&built), "the copy keeps the time");
    // Detached at the commit, with the checkout of that commit.
    assert_eq!(git(&r.slot("lane-01"), &["rev-parse", "HEAD"]), r.head);
    let symbolic = common::git::git_raw(&r.slot("lane-01"), &["symbolic-ref", "-q", "HEAD"]);
    assert!(!symbolic.success, "detached: {}", symbolic.stdout);
    assert!(r.slot("lane-01").join("a.txt").is_file());

    // Again: present, nothing changed, its own target/ kept.
    fs::write(r.slot("lane-01").join("target/debug/own.txt"), "own\n").unwrap();
    let before = r.worktrees();
    let again = r.ok(&["ensure", "lane-01", "--at", &r.head]);
    assert_eq!(
        (&again["added"], &again["present"], &again["warmed"]),
        (&json!(false), &json!(true), &json!("none")),
        "{again}"
    );
    assert_eq!(r.worktrees(), before);
    assert!(r.slot("lane-01").join("target/debug/own.txt").is_file());

    // The plan now shows both slots.
    let plan = r.ok(&["plan"]);
    let shown = |n: &str| {
        plan["slots"]
            .as_array()
            .unwrap()
            .iter()
            .find(|s| s["name"] == json!(n))
            .unwrap()
            .clone()
    };
    for n in ["seed", "lane-01"] {
        let s = shown(n);
        assert_eq!(
            (&s["exists"], &s["head"], &s["branch"], &s["target"]),
            (
                &json!(true),
                &json!(r.head),
                &json!("detached"),
                &json!(true)
            ),
            "{s}"
        );
    }
    assert_eq!(plan["caps"]["worktrees"]["count"], json!(2), "{plan}");
}

#[test]
fn ensure_without_a_seed_build_adds_the_slot_cold() {
    let r = repo(&[("a.txt", "a\n")]);
    let lane = r.ok(&["ensure", "audit-2", "--at", &r.head]);
    assert_eq!(
        (&lane["added"], &lane["warmed"]),
        (&json!(true), &json!("none")),
        "{lane}"
    );
    assert!(!r.slot("audit-2").join("target").exists());
}

// ---- caps (AC-4: REQ-5)

#[test]
fn ensure_refuses_past_the_worktree_cap_or_under_the_free_floor_and_adds_nothing() {
    let r = repo(&[("a.txt", "a\n")]);
    r.ok(&["ensure", "lane-01", "--at", &r.head]);
    let before = r.worktrees();
    let (code, v) = r.run(&["ensure", "lane-02", "--at", &r.head, "--max-worktrees", "1"]);
    assert_ne!(code, 0, "{v}");
    let m = v["error"]["message"].as_str().unwrap();
    assert!(
        m.contains("worktree cap") && m.contains("limit 1") && m.contains("count 1"),
        "{m}"
    );
    let (code, v) = r.run(&[
        "ensure",
        "lane-02",
        "--at",
        &r.head,
        "--min-free-percent",
        "100",
    ]);
    assert_ne!(code, 0, "{v}");
    let m = v["error"]["message"].as_str().unwrap();
    assert!(m.contains("free disk") && m.contains("limit 100"), "{m}");
    assert_eq!(r.worktrees(), before);
    assert!(!r.slot("lane-02").exists());
    // An existing slot is reported present even at the cap: nothing is added.
    let again = r.ok(&["ensure", "lane-01", "--at", &r.head, "--max-worktrees", "1"]);
    assert_eq!(again["present"], json!(true), "{again}");
}

// ---- a clone warms the dependencies (AC-3: REQ-4)

/// A crate with a dependency outside the repository (as a registry crate sits outside it),
/// built in the seed, then built again in a slot warmed from the seed: only the crate compiles.
#[cfg(target_os = "macos")]
#[test]
fn a_slot_warmed_from_the_seed_compiles_only_the_workspace_crate() {
    let dep_dir = tempfile::tempdir().unwrap();
    let dep = fs::canonicalize(dep_dir.path()).unwrap();
    fs::write(
        dep.join("Cargo.toml"),
        "[package]\nname = \"slotdep\"\nversion = \"0.1.0\"\nedition = \"2021\"\n[lib]\npath = \"lib.rs\"\n",
    )
    .unwrap();
    fs::write(dep.join("lib.rs"), "pub fn one() -> u32 { 1 }\n").unwrap();
    let manifest = format!(
        "[package]\nname = \"slotapp\"\nversion = \"0.1.0\"\nedition = \"2021\"\n[workspace]\n[dependencies]\nslotdep = {{ path = \"{}\" }}\n",
        dep.display()
    );
    let r = repo(&[
        ("Cargo.toml", &manifest),
        (
            "src/main.rs",
            "fn main() { println!(\"{}\", slotdep::one()); }\n",
        ),
    ]);
    let cargo = std::env::var("CARGO").unwrap_or_else(|_| "cargo".into());
    let build = |dir: &Path| {
        let out = Command::new(&cargo)
            .args(["build", "-v", "--offline"])
            .current_dir(dir)
            .env_remove("CARGO_TARGET_DIR")
            .env_remove("CARGO_BUILD_TARGET_DIR")
            .env_remove("RUSTC_WRAPPER")
            .env_remove("CARGO_BUILD_RUSTC_WRAPPER")
            .env_remove("CARGO_INCREMENTAL")
            .output()
            .unwrap();
        let err = String::from_utf8_lossy(&out.stderr).to_string();
        assert!(out.status.success(), "{err}");
        err
    };
    r.ok(&["ensure", "seed", "--at", &r.head]);
    let first = build(&r.slot("seed"));
    assert!(first.contains("Compiling slotdep"), "{first}");
    let lane = r.ok(&["ensure", "lane-03", "--at", &r.head]);
    assert_eq!(lane["warmed"], json!("clone"), "{lane}");
    let warm = build(&r.slot("lane-03"));
    assert!(warm.contains("Fresh slotdep"), "{warm}");
    assert!(!warm.contains("Compiling slotdep"), "{warm}");
    assert!(warm.contains("Compiling slotapp"), "{warm}");
}

/// The caps are exact: one under the worktree cap adds, and a floor equal to the measured free
/// share adds (N against N+1).
#[test]
fn ensure_adds_right_at_the_caps() {
    let r = repo(&[("a.txt", "a\n")]);
    r.ok(&["ensure", "lane-01", "--at", &r.head]);
    let added = r.ok(&["ensure", "lane-02", "--at", &r.head, "--max-worktrees", "2"]);
    assert_eq!(added["added"], json!(true), "{added}");
    let free = r.ok(&["plan"])["caps"]["free_percent"]["value"]
        .as_u64()
        .unwrap();
    let at = free.to_string();
    let added = r.ok(&[
        "ensure",
        "lane-03",
        "--at",
        &r.head,
        "--min-free-percent",
        &at,
    ]);
    assert_eq!(added["added"], json!(true), "{added}");
    let over = (free + 1).to_string();
    let (code, v) = r.run(&[
        "ensure",
        "lane-04",
        "--at",
        &r.head,
        "--min-free-percent",
        &over,
    ]);
    assert_ne!(code, 0, "{v}");
    let (code, v) = r.run(&["ensure", "lane-04", "--at", "no-such-commit"]);
    assert_ne!(code, 0, "{v}");
    assert!(
        v["error"]["message"]
            .as_str()
            .unwrap()
            .contains("names no commit"),
        "{v}"
    );
    assert!(!r.slot("lane-04").exists());
}
