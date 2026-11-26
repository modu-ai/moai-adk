//! SPEC-20261123-worktree-slots slice 1: `worktree-slot plan|ensure`. Worktrees are fixed slots
//! `.claude/worktrees/lane-01..10`, `audit-1..3` and `seed`; a new slot is added detached at a
//! commit and, when it has no `target/` while the seed has one, warmed with an APFS clone of
//! the seed's `target/`; a slot is never added past the worktree cap or under the free-disk
//! floor. The binary runs on temporary repositories in a clean environment.
mod common;

use common::git::git;
use common::home::TempHome;
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
    home: TempHome,
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
            "user.email=t@example.invalid",
            "commit",
            "-q",
            "-m",
            "one",
        ],
    );
    let head = git(&root, &["rev-parse", "HEAD"]);
    Repo {
        _dir: dir,
        home: TempHome::new(),
        root,
        head,
    }
}

impl Repo {
    fn slot(&self, name: &str) -> PathBuf {
        self.root.join(".claude/worktrees").join(name)
    }

    fn run(&self, args: &[&str]) -> (i32, Value) {
        self.run_env(args, &[])
    }

    fn run_env(&self, args: &[&str], envs: &[(&str, &str)]) -> (i32, Value) {
        let out = self
            .home
            .command(&["worktree-slot"])
            .envs(envs.iter().copied())
            .current_dir(&self.root)
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
    let build = |dir: &Path| {
        let out = common::cargo::cargo_at(common::cargo::cargo_path())
            .args(["build", "-v", "--offline"])
            .current_dir(dir)
            .env_remove("CARGO_TARGET_DIR")
            .env_remove("CARGO_BUILD_TARGET_DIR")
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

// ---- audit repair: one ensure at a time, a copy that is complete or absent

fn seed_with_build(r: &Repo) {
    r.ok(&["ensure", "seed", "--at", &r.head]);
    let built = r.slot("seed").join("target/debug/built.txt");
    fs::create_dir_all(built.parent().unwrap()).unwrap();
    fs::write(&built, "seed build\n").unwrap();
}

fn leftovers(slot: &Path) -> Vec<String> {
    fs::read_dir(slot)
        .unwrap()
        .flatten()
        .map(|e| e.file_name().to_string_lossy().to_string())
        .filter(|n| n.starts_with(".target-warming-"))
        .collect()
}

/// Three ensure runs of one slot at once: one adds and warms it, the others find it present;
/// the slot is registered once and its `target/` holds no nested copy.
#[test]
fn concurrent_ensure_of_one_slot_adds_it_once_and_warms_it_once() {
    let r = repo(&[("a.txt", "a\n")]);
    seed_with_build(&r);
    for round in 0..3 {
        let name = ["lane-02", "lane-05", "audit-1"][round];
        let answers: Vec<(i32, Value)> = std::thread::scope(|s| {
            let runs: Vec<_> = (0..3)
                .map(|_| s.spawn(|| r.run(&["ensure", name, "--at", &r.head])))
                .collect();
            runs.into_iter().map(|h| h.join().unwrap()).collect()
        });
        for (code, v) in &answers {
            assert_eq!((*code, &v["ok"]), (0, &json!(true)), "{name}: {v}");
        }
        let added = answers
            .iter()
            .filter(|(_, v)| v["result"]["added"] == json!(true))
            .count();
        assert_eq!(added, 1, "{name}: {answers:?}");
        let registered = r
            .worktrees()
            .lines()
            .filter(|l| l.starts_with("worktree ") && l.ends_with(&format!("/{name}")))
            .count();
        assert_eq!(registered, 1, "{name}: {}", r.worktrees());
        assert!(r.slot(name).join("target/debug/built.txt").is_file());
        assert!(!r.slot(name).join("target/target").exists(), "{name}");
        assert!(leftovers(&r.slot(name)).is_empty(), "{name}");
    }
}

/// A slot whose checkout already has `target/` (here a tracked one) is never copied into.
#[test]
fn a_slot_that_already_has_target_is_not_copied_into() {
    let r = repo(&[("a.txt", "a\n")]);
    fs::create_dir_all(r.root.join("target")).unwrap();
    fs::write(r.root.join("target/keep.txt"), "kept\n").unwrap();
    git(&r.root, &["add", "-f", "target/keep.txt"]);
    git(
        &r.root,
        &[
            "-c",
            "user.name=t",
            "-c",
            "user.email=t@example.invalid",
            "commit",
            "-q",
            "-m",
            "t",
        ],
    );
    let head = git(&r.root, &["rev-parse", "HEAD"]);
    seed_with_build(&r);
    let lane = r.ok(&["ensure", "lane-04", "--at", &head]);
    assert_eq!(
        (&lane["added"], &lane["warmed"]),
        (&json!(true), &json!("none")),
        "{lane}"
    );
    assert!(r.slot("lane-04").join("target/keep.txt").is_file());
    assert!(!r.slot("lane-04").join("target/target").exists());
    assert!(!r.slot("lane-04").join("target/debug").exists());
}

#[test]
fn a_folder_at_the_slot_path_that_is_not_a_worktree_is_refused() {
    let r = repo(&[("a.txt", "a\n")]);
    fs::create_dir_all(r.slot("lane-05")).unwrap();
    let before = r.worktrees();
    let (code, v) = r.run(&["ensure", "lane-05", "--at", &r.head]);
    assert_ne!(code, 0, "{v}");
    assert!(
        v["error"]["message"]
            .as_str()
            .unwrap()
            .contains("lane-05 exists but is not a worktree"),
        "{v}"
    );
    assert_eq!(r.worktrees(), before);
}

/// A copy that fails leaves no `target/` and no temporary copy, and fails the command; the next
/// ensure warms the present slot.
#[cfg(unix)]
#[test]
fn a_failed_warm_fails_leaves_no_target_and_the_next_ensure_warms_it() {
    use std::os::unix::fs::PermissionsExt;
    let r = repo(&[("a.txt", "a\n")]);
    seed_with_build(&r);
    let locked = r.slot("seed").join("target/locked");
    fs::create_dir_all(&locked).unwrap();
    fs::write(locked.join("x"), "x\n").unwrap();
    fs::set_permissions(&locked, fs::Permissions::from_mode(0o000)).unwrap();
    let (code, v) = r.run(&["ensure", "lane-06", "--at", &r.head]);
    fs::set_permissions(&locked, fs::Permissions::from_mode(0o755)).unwrap();
    assert_ne!(code, 0, "{v}");
    assert!(
        v["error"]["message"]
            .as_str()
            .unwrap()
            .contains("lane-06 is added but warming failed"),
        "{v}"
    );
    assert!(!r.slot("lane-06").join("target").exists());
    assert!(leftovers(&r.slot("lane-06")).is_empty());
    let again = r.ok(&["ensure", "lane-06", "--at", &r.head]);
    let method = if cfg!(target_os = "macos") {
        "clone"
    } else {
        "copy"
    };
    assert_eq!(
        (&again["present"], &again["warmed"]),
        (&json!(true), &json!(method)),
        "{again}"
    );
    assert!(r.slot("lane-06").join("target/locked/x").is_file());
}

/// A registered slot whose folder is gone is reported missing, and ensure refuses it.
#[test]
fn a_registered_slot_whose_folder_is_gone_is_reported_missing() {
    let r = repo(&[("a.txt", "a\n")]);
    r.ok(&["ensure", "lane-07", "--at", &r.head]);
    fs::rename(r.slot("lane-07"), r.root.join("moved-away")).unwrap();
    let plan = r.ok(&["plan"]);
    let s = plan["slots"]
        .as_array()
        .unwrap()
        .iter()
        .find(|s| s["name"] == json!("lane-07"))
        .unwrap()
        .clone();
    assert_eq!(
        (&s["exists"], &s["missing"]),
        (&json!(true), &json!(true)),
        "{s}"
    );
    let (code, v) = r.run(&["ensure", "lane-07", "--at", &r.head]);
    assert_ne!(code, 0, "{v}");
    assert!(
        v["error"]["message"]
            .as_str()
            .unwrap()
            .contains("folder is missing"),
        "{v}"
    );
}

/// A seed `target` that is a link (a shared build folder) or a seed that is not a worktree is
/// refused before anything is added.
#[cfg(unix)]
#[test]
fn a_linked_seed_target_or_a_seed_that_is_not_a_worktree_is_refused() {
    let r = repo(&[("a.txt", "a\n")]);
    r.ok(&["ensure", "seed", "--at", &r.head]);
    let shared = r.root.join("shared-build");
    fs::create_dir_all(&shared).unwrap();
    std::os::unix::fs::symlink(&shared, r.slot("seed").join("target")).unwrap();
    let before = r.worktrees();
    let (code, v) = r.run(&["ensure", "lane-08", "--at", &r.head]);
    assert_ne!(code, 0, "{v}");
    assert!(
        v["error"]["message"]
            .as_str()
            .unwrap()
            .contains("not a real folder"),
        "{v}"
    );
    assert_eq!(r.worktrees(), before);

    let r = repo(&[("a.txt", "a\n")]);
    fs::create_dir_all(r.slot("seed").join("target")).unwrap();
    let (code, v) = r.run(&["ensure", "lane-09", "--at", &r.head]);
    assert_ne!(code, 0, "{v}");
    assert!(
        v["error"]["message"]
            .as_str()
            .unwrap()
            .contains("seed exists but is not a worktree"),
        "{v}"
    );
    assert!(!r.slot("lane-09").exists());
}

/// A `df` that fails refuses: the free share is unknown.
#[cfg(unix)]
#[test]
fn a_failing_df_refuses() {
    use std::os::unix::fs::PermissionsExt;
    let r = repo(&[("a.txt", "a\n")]);
    let bin = r.root.parent().unwrap().join("fake-bin");
    fs::create_dir_all(&bin).unwrap();
    let df = bin.join("df");
    fs::write(&df, "#!/bin/sh\nexit 1\n").unwrap();
    fs::set_permissions(&df, fs::Permissions::from_mode(0o755)).unwrap();
    let path = format!("{}:/usr/bin:/bin", bin.display());
    for args in [
        vec!["plan"],
        vec!["ensure", "lane-10", "--at", r.head.as_str()],
    ] {
        let (code, v) = r.run_env(&args, &[("PATH", &path)]);
        assert_ne!(code, 0, "{args:?}: {v}");
        assert!(
            v["error"]["message"]
                .as_str()
                .unwrap()
                .starts_with("df failed"),
            "{args:?}: {v}"
        );
    }
    assert!(!r.slot("lane-10").exists());
}

// ---- follow-ups of the slice 1 and 2 verdict

/// A slot path that is a link (here to another worktree) is refused, and nothing is written
/// through it.
#[cfg(unix)]
#[test]
fn a_slot_path_that_is_a_link_is_refused_and_nothing_is_written_through_it() {
    let r = repo(&[("a.txt", "a\n")]);
    seed_with_build(&r);
    let other = r.root.parent().unwrap().join("other");
    git(
        &r.root,
        &[
            "worktree",
            "add",
            "-q",
            "--detach",
            other.to_str().unwrap(),
            &r.head,
        ],
    );
    std::os::unix::fs::symlink(&other, r.slot("lane-01")).unwrap();
    let before = r.worktrees();
    let (code, v) = r.run(&["ensure", "lane-01", "--at", &r.head]);
    assert_ne!(code, 0, "{v}");
    assert_eq!(v["error"]["kind"], json!("invalid_input"), "{v}");
    let m = v["error"]["message"].as_str().unwrap();
    assert!(m.contains("lane-01") && m.contains("is a link"), "{m}");
    assert!(
        !other.join("target").exists(),
        "nothing was warmed through the link"
    );
    assert_eq!(r.worktrees(), before);
}

/// Warming a present slot that has no `target/` meets the free-disk floor like adding one does;
/// under the floor nothing is copied, and at the default floor the same slot is warmed.
#[test]
fn rewarming_a_present_slot_meets_the_free_disk_floor() {
    let r = repo(&[("a.txt", "a\n")]);
    r.ok(&["ensure", "lane-01", "--at", &r.head]);
    seed_with_build(&r);
    let (code, v) = r.run(&[
        "ensure",
        "lane-01",
        "--at",
        &r.head,
        "--min-free-percent",
        "100",
    ]);
    assert_ne!(code, 0, "{v}");
    let m = v["error"]["message"].as_str().unwrap();
    assert!(
        m.contains("free disk") && m.contains("was not warmed"),
        "{m}"
    );
    assert!(!r.slot("lane-01").join("target").exists());
    let warm = r.ok(&["ensure", "lane-01", "--at", &r.head]);
    assert_ne!(warm["warmed"], json!("none"), "{warm}");
    assert!(r.slot("lane-01").join("target/debug/built.txt").is_file());
}

/// The temporary copies a killed warm or inherit leaves are removed by the next ensure of the
/// slot, which holds the lock, and the answer names them.
#[test]
fn ensure_removes_the_temporary_copies_a_killed_run_left() {
    let r = repo(&[("a.txt", "a\n")]);
    r.ok(&["ensure", "lane-02", "--at", &r.head]);
    let slot = r.slot("lane-02");
    fs::create_dir_all(slot.join(".target-warming-99999/debug")).unwrap();
    fs::write(slot.join(".target-warming-99999/debug/x"), "x\n").unwrap();
    fs::create_dir_all(slot.join(".target-inherit-88888")).unwrap();
    fs::write(slot.join(".target-keep-me"), "not a temporary copy\n").unwrap();
    let again = r.ok(&["ensure", "lane-02", "--at", &r.head]);
    assert_eq!(
        again["cleaned"],
        json!([".target-inherit-88888", ".target-warming-99999"]),
        "{again}"
    );
    assert!(leftovers(&slot).is_empty(), "{:?}", leftovers(&slot));
    assert!(!slot.join(".target-inherit-88888").exists());
    assert!(slot.join(".target-keep-me").is_file());
}

/// An audit slot's leftover copy is left for audit-slot to name and discard, so `ensure` of an
/// audit slot removes nothing and names nothing.
#[test]
fn ensure_leaves_an_audit_slots_leftover_copy_to_audit_slot() {
    // Every audit slot, not the first alone: a rule that names one slot or one prefix of the
    // names would clean the others' copies (the audit's third mutant).
    for name in ["audit-1", "audit-2", "audit-3"] {
        let r = repo(&[("a.txt", "a\n")]);
        r.ok(&["ensure", name, "--at", &r.head]);
        let slot = r.slot(name);
        fs::create_dir_all(slot.join(".target-inherit-4242/debug")).unwrap();
        fs::write(slot.join(".target-inherit-4242/debug/x"), "x\n").unwrap();
        let again = r.ok(&["ensure", name, "--at", &r.head]);
        assert_eq!(again["cleaned"], json!([]), "{name}: {again}");
        assert!(
            slot.join(".target-inherit-4242/debug/x").is_file(),
            "{name}: its copy was removed"
        );
    }
}

/// A run that cannot get the slot lock waits for the configured time, then refuses and adds
/// nothing.
#[test]
fn ensure_waits_for_the_lock_then_refuses() {
    let r = repo(&[("a.txt", "a\n")]);
    let common = git(
        &r.root,
        &["rev-parse", "--path-format=absolute", "--git-common-dir"],
    );
    let held = fs::OpenOptions::new()
        .create(true)
        .truncate(false)
        .write(true)
        .open(Path::new(&common).join("moai-worktree-slot.lock"))
        .unwrap();
    held.lock().unwrap();
    let started = std::time::Instant::now();
    let (code, v) = r.run_env(
        &["ensure", "lane-03", "--at", &r.head],
        &[("MOAI_WORKTREE_SLOT_LOCK_WAIT_SECS", "1")],
    );
    let waited = started.elapsed();
    assert_ne!(code, 0, "{v}");
    let m = v["error"]["message"].as_str().unwrap();
    assert!(m.contains("held the slot lock for 1 seconds"), "{m}");
    assert!(waited >= Duration::from_secs(1), "{waited:?}");
    assert!(!r.slot("lane-03").exists());
    held.unlock().unwrap();
    let added = r.ok(&["ensure", "lane-03", "--at", &r.head]);
    assert_eq!(added["added"], json!(true), "{added}");
}
