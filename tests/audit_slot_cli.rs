//! SPEC-20261123-audit-slot: `audit-slot prep|done`. An audit takes a fixed audit slot
//! (`.claude/worktrees/audit-1..3`), switches it detached to the audited head and inherits the
//! lane's build when the lane's worktree is at that head and clean (else the seed's); `done`
//! leaves the slot clean for the next audit and keeps its `target/`. The binary runs on
//! temporary repositories in a clean environment.
mod common;

use common::git::git;
use common::home::TempHome;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::time::{Duration, SystemTime};

struct Repo {
    _dir: tempfile::TempDir,
    home: TempHome,
    root: PathBuf,
    /// The first commit, on `main`.
    first: String,
    /// The lane's commit on top of `first`, checked out in [`Repo::lane`].
    second: String,
    /// A linked worktree outside the slots, standing for a lane's READY worktree.
    lane: PathBuf,
}

fn commit(dir: &Path, message: &str) {
    git(
        dir,
        &[
            "-c",
            "user.name=t",
            "-c",
            "user.email=t@example.invalid",
            "commit",
            "-q",
            "-m",
            message,
        ],
    );
}

/// A repository with one commit holding `files`, and a lane worktree on branch `work` with a
/// second commit that adds `b.txt`.
fn repo(files: &[(&str, &str)]) -> Repo {
    let dir = tempfile::tempdir().unwrap();
    let base = fs::canonicalize(dir.path()).unwrap();
    let root = base.join("repo");
    fs::create_dir_all(&root).unwrap();
    git(&root, &["init", "-q", "-b", "main"]);
    for (rel, text) in files {
        let p = root.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, text).unwrap();
    }
    fs::write(root.join(".gitignore"), "/target/\n/.claude/\n").unwrap();
    git(&root, &["add", "-A"]);
    commit(&root, "one");
    let first = git(&root, &["rev-parse", "HEAD"]);
    let lane = base.join("lane");
    git(
        &root,
        &[
            "worktree",
            "add",
            "-q",
            "-b",
            "work",
            lane.to_str().unwrap(),
        ],
    );
    fs::write(lane.join("b.txt"), "b\n").unwrap();
    git(&lane, &["add", "b.txt"]);
    commit(&lane, "two");
    let second = git(&lane, &["rev-parse", "HEAD"]);
    Repo {
        _dir: dir,
        home: TempHome::new(),
        root,
        first,
        second,
        lane,
    }
}

fn mtime(p: &Path) -> SystemTime {
    fs::metadata(p).unwrap().modified().unwrap()
}

/// Writes `rel` under `dir` with a modification time three days old.
fn old_file(dir: &Path, rel: &str, text: &str) -> PathBuf {
    let p = dir.join(rel);
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::write(&p, text).unwrap();
    let old = SystemTime::now() - Duration::from_secs(3 * 86_400);
    fs::File::options()
        .write(true)
        .open(&p)
        .unwrap()
        .set_modified(old)
        .unwrap();
    p
}

fn leftovers(slot: &Path) -> Vec<String> {
    fs::read_dir(slot)
        .unwrap()
        .flatten()
        .map(|e| e.file_name().to_string_lossy().to_string())
        .filter(|n| n.starts_with(".target-"))
        .collect()
}

impl Repo {
    fn slot(&self, name: &str) -> PathBuf {
        self.root.join(".claude/worktrees").join(name)
    }

    fn run_op(&self, op: &str, args: &[&str]) -> (i32, Value) {
        self.run_op_env(op, args, &[])
    }

    fn run_op_env(&self, op: &str, args: &[&str], env: &[(&str, &str)]) -> (i32, Value) {
        let mut cmd = self.home.command(&[op]);
        cmd.current_dir(&self.root)
            .args(args)
            .args(["--project", self.root.to_str().unwrap()]);
        // `prep` and `done` name a holder; the tests act as one auditor unless they name another.
        if op == "audit-slot"
            && matches!(args.first(), Some(&"prep") | Some(&"done"))
            && !args.contains(&"--holder")
        {
            cmd.args(["--holder", "auditor-a"]);
        }
        for (k, v) in env {
            cmd.env(k, v);
        }
        let out = cmd.output().unwrap();
        let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
            panic!(
                "{args:?}: not JSON: {} / {}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        });
        (out.status.code().unwrap_or(-1), v)
    }

    fn run(&self, args: &[&str]) -> (i32, Value) {
        self.run_op("audit-slot", args)
    }

    fn ok(&self, args: &[&str]) -> Value {
        let (code, v) = self.run(args);
        assert_eq!((code, &v["ok"]), (0, &json!(true)), "{args:?}: {v}");
        v["result"].clone()
    }

    /// A refusal: non-zero exit, `invalid_input`, and a message holding every word of `words`.
    fn refused(&self, args: &[&str], words: &[&str]) -> String {
        let (code, v) = self.run(args);
        assert_ne!(code, 0, "{args:?}: {v}");
        assert_eq!(v["error"]["kind"], json!("invalid_input"), "{args:?}: {v}");
        let m = v["error"]["message"].as_str().unwrap().to_string();
        for w in words {
            assert!(m.contains(w), "{args:?}: {w:?} not in {m}");
        }
        m
    }

    fn worktrees(&self) -> String {
        git(&self.root, &["worktree", "list", "--porcelain"])
    }

    fn lane_str(&self) -> &str {
        self.lane.to_str().unwrap()
    }
}

fn detached_at(slot: &Path, head: &str) {
    assert_eq!(git(slot, &["rev-parse", "HEAD"]), head);
    let symbolic = common::git::git_raw(slot, &["symbolic-ref", "-q", "HEAD"]);
    assert!(!symbolic.success, "detached: {}", symbolic.stdout);
    assert_eq!(
        git(slot, &["status", "--porcelain", "--untracked-files=all"]),
        ""
    );
}

// ---- names and heads (AC-1: REQ-1)

#[test]
fn prep_refuses_a_name_that_is_not_an_audit_slot() {
    let r = repo(&[("a.txt", "a\n")]);
    let before = r.worktrees();
    for bad in [
        "lane-01",
        "seed",
        "audit-4",
        "audit-0",
        "../x",
        "audit-1/x",
        "",
    ] {
        r.refused(&["prep", bad, "--head", &r.first], &["audit-1"]);
        r.refused(&["done", bad], &["audit-1"]);
    }
    // Not a full commit id: a short id, a branch, a tag, upper case, or 40 hex naming nothing.
    let short = &r.first[..12];
    let upper = r.first.to_uppercase();
    for bad in [short, "main", "HEAD", "no-such-commit", upper.as_str()] {
        r.refused(
            &["prep", "audit-1", "--head", bad],
            &["full 40-character commit id"],
        );
    }
    r.refused(
        &["prep", "audit-1", "--head", &"0".repeat(40)],
        &["names no commit"],
    );
    assert_eq!(r.worktrees(), before);
    assert!(!r.root.join(".claude").exists());
}

// ---- prep from the lane (AC-1: REQ-2; AC-2: REQ-4)

#[test]
fn prep_adds_the_slot_at_the_head_and_inherits_the_lane_build() {
    let r = repo(&[("a.txt", "a\n")]);
    let built = old_file(&r.lane, "target/debug/built.txt", "lane build\n");
    let prep = r.ok(&[
        "prep",
        "audit-1",
        "--head",
        &r.second,
        "--lane",
        r.lane_str(),
    ]);
    assert_eq!(
        (&prep["added"], &prep["head"], &prep["build"]["source"]),
        (&json!(true), &json!(r.second), &json!("lane")),
        "{prep}"
    );
    let method = if cfg!(target_os = "macos") {
        "clone"
    } else {
        "copy"
    };
    assert_eq!(prep["build"]["method"], json!(method), "{prep}");
    let slot = r.slot("audit-1");
    detached_at(&slot, &r.second);
    assert!(slot.join("b.txt").is_file());
    let copied = slot.join("target/debug/built.txt");
    assert_eq!(fs::read_to_string(&copied).unwrap(), "lane build\n");
    assert_eq!(mtime(&copied), mtime(&built), "the copy keeps the time");
    assert!(leftovers(&slot).is_empty(), "{:?}", leftovers(&slot));
    // The lane is untouched.
    assert_eq!(git(&r.lane, &["rev-parse", "HEAD"]), r.second);
    assert!(built.is_file());
}

/// The next audit in the same slot replaces the slot's `target/` with the lane's whole: a file
/// only the old slot build had is gone, and no temporary or old copy is left.
#[test]
fn prep_again_replaces_the_slot_target_with_the_lane_build() {
    let r = repo(&[("a.txt", "a\n")]);
    old_file(&r.lane, "target/debug/built.txt", "lane build\n");
    r.ok(&["prep", "audit-2", "--head", &r.first]);
    let slot = r.slot("audit-2");
    fs::create_dir_all(slot.join("target/debug")).unwrap();
    fs::write(slot.join("target/debug/stale.txt"), "old audit\n").unwrap();
    let prep = r.ok(&[
        "prep",
        "audit-2",
        "--head",
        &r.second,
        "--lane",
        r.lane_str(),
    ]);
    assert_eq!(
        (&prep["added"], &prep["build"]["source"]),
        (&json!(false), &json!("lane")),
        "{prep}"
    );
    detached_at(&slot, &r.second);
    assert!(!slot.join("target/debug/stale.txt").exists());
    assert!(slot.join("target/debug/built.txt").is_file());
    assert!(!slot.join("target/target").exists());
    assert!(leftovers(&slot).is_empty(), "{:?}", leftovers(&slot));
}

// ---- prep without a lane (AC-2: REQ-5)

#[test]
fn prep_without_a_lane_keeps_the_slot_build_or_warms_from_the_seed() {
    let r = repo(&[("a.txt", "a\n")]);
    // No seed, no lane: cold.
    let cold = r.ok(&["prep", "audit-1", "--head", &r.first]);
    assert_eq!(cold["build"]["source"], json!("none"), "{cold}");
    assert!(!r.slot("audit-1").join("target").exists());
    // A present slot with its own build switches head and keeps it.
    let own = r.slot("audit-1").join("target/debug/own.txt");
    fs::create_dir_all(own.parent().unwrap()).unwrap();
    fs::write(&own, "own\n").unwrap();
    let kept = r.ok(&["prep", "audit-1", "--head", &r.second]);
    assert_eq!(
        (&kept["added"], &kept["build"]["source"]),
        (&json!(false), &json!("kept")),
        "{kept}"
    );
    detached_at(&r.slot("audit-1"), &r.second);
    assert!(own.is_file());
    // With a seed build, a new slot is warmed from it.
    let (code, v) = r.run_op("worktree-slot", &["ensure", "seed", "--at", &r.first]);
    assert_eq!(code, 0, "{v}");
    old_file(&r.slot("seed"), "target/debug/seed.txt", "seed build\n");
    let warm = r.ok(&["prep", "audit-3", "--head", &r.second]);
    assert_eq!(warm["build"]["source"], json!("seed"), "{warm}");
    assert!(r.slot("audit-3").join("target/debug/seed.txt").is_file());
    detached_at(&r.slot("audit-3"), &r.second);
}

// ---- refusals that change nothing (AC-1: REQ-3, REQ-6)

#[cfg(unix)]
#[test]
fn prep_refuses_a_lane_it_cannot_inherit_from_and_changes_nothing() {
    let r = repo(&[("a.txt", "a\n")]);
    let before = r.worktrees();
    let lane = r.lane_str().to_string();
    let args = |h: &str| {
        vec![
            "prep".to_string(),
            "audit-1".into(),
            "--head".into(),
            h.into(),
            "--lane".into(),
            lane.clone(),
        ]
    };
    let run = |h: &str, words: &[&str]| {
        let a = args(h);
        let a: Vec<&str> = a.iter().map(String::as_str).collect();
        r.refused(&a, words)
    };
    // No target/ in the lane.
    run(&r.second, &["no target/"]);
    // A lane at another head.
    old_file(&r.lane, "target/debug/built.txt", "lane build\n");
    run(&r.first, &["is at", &r.second[..12]]);
    // A lane with a tracked change.
    fs::write(r.lane.join("a.txt"), "changed\n").unwrap();
    run(&r.second, &["tracked change", "a.txt"]);
    git(&r.lane, &["checkout", "--", "a.txt"]);
    // A lane whose target is a link to a shared build folder.
    let shared = r.root.parent().unwrap().join("shared-target");
    fs::create_dir_all(&shared).unwrap();
    fs::rename(r.lane.join("target"), r.lane.join("target-real")).unwrap();
    std::os::unix::fs::symlink(&shared, r.lane.join("target")).unwrap();
    run(&r.second, &["not a real folder"]);
    // A folder that is not a worktree of this repository.
    let plain = r.root.parent().unwrap().join("plain");
    fs::create_dir_all(plain.join("target")).unwrap();
    r.refused(
        &[
            "prep",
            "audit-1",
            "--head",
            &r.second,
            "--lane",
            plain.to_str().unwrap(),
        ],
        &["not a worktree of this repository"],
    );
    assert_eq!(r.worktrees(), before);
    assert!(!r.slot("audit-1").exists());
}

#[test]
fn prep_refuses_a_slot_with_changes_and_names_them() {
    let r = repo(&[("a.txt", "a\n")]);
    r.ok(&["prep", "audit-1", "--head", &r.first]);
    let slot = r.slot("audit-1");
    fs::write(slot.join("a.txt"), "mutant\n").unwrap();
    fs::write(slot.join("new.txt"), "probe\n").unwrap();
    r.refused(
        &["prep", "audit-1", "--head", &r.second],
        &["a.txt", "new.txt", "done audit-1 --discard"],
    );
    assert_eq!(git(&slot, &["rev-parse", "HEAD"]), r.first);
    assert_eq!(fs::read_to_string(slot.join("a.txt")).unwrap(), "mutant\n");
}

// ---- done (AC-4: REQ-7, REQ-8)

#[test]
fn done_refuses_changes_without_discard_then_discards_them_and_keeps_target() {
    let r = repo(&[("a.txt", "a\n"), ("src/lib.rs", "pub fn x() {}\n")]);
    old_file(&r.lane, "target/debug/built.txt", "lane build\n");
    r.ok(&[
        "prep",
        "audit-1",
        "--head",
        &r.second,
        "--lane",
        r.lane_str(),
    ]);
    let slot = r.slot("audit-1");
    // An audit's mutant, a staged new file and a probe left behind.
    fs::write(slot.join("src/lib.rs"), "pub fn x() { panic!() }\n").unwrap();
    fs::write(slot.join("staged.txt"), "s\n").unwrap();
    git(&slot, &["add", "staged.txt"]);
    fs::create_dir_all(slot.join("probe")).unwrap();
    fs::write(slot.join("probe/out.log"), "log\n").unwrap();

    r.refused(
        &["done", "audit-1"],
        &["src/lib.rs", "staged.txt", "probe/out.log", "--discard"],
    );
    assert_eq!(
        fs::read_to_string(slot.join("src/lib.rs")).unwrap(),
        "pub fn x() { panic!() }\n"
    );

    let done = r.ok(&["done", "audit-1", "--discard"]);
    let list = |k: &str| -> Vec<String> {
        done[k]
            .as_array()
            .unwrap()
            .iter()
            .map(|v| v.as_str().unwrap().to_string())
            .collect()
    };
    assert_eq!(list("reverted"), ["src/lib.rs", "staged.txt"], "{done}");
    assert_eq!(list("removed"), ["probe/out.log"], "{done}");
    assert_eq!(
        (&done["clean"], &done["target_kept"]),
        (&json!(true), &json!(true)),
        "{done}"
    );
    detached_at(&slot, &r.second);
    assert!(slot.join("target/debug/built.txt").is_file());
    assert!(!slot.join("staged.txt").exists());
    assert!(!slot.join("probe/out.log").exists());

    // A clean slot is done at once.
    let again = r.ok(&["done", "audit-1"]);
    assert_eq!(
        (&again["clean"], &again["reverted"], &again["removed"]),
        (&json!(true), &json!([]), &json!([])),
        "{again}"
    );
}

#[test]
fn done_refuses_a_slot_that_is_not_there() {
    let r = repo(&[("a.txt", "a\n")]);
    r.refused(&["done", "audit-2"], &["audit-2", "not a worktree"]);
}

/// Every file under `dir` (the `.git` link file left out) with its bytes, and its git status.
fn snapshot(dir: &Path) -> (Vec<(String, Vec<u8>)>, String) {
    fn walk(base: &Path, dir: &Path, out: &mut Vec<(String, Vec<u8>)>) {
        for e in fs::read_dir(dir).unwrap().flatten() {
            let p = e.path();
            let rel = p.strip_prefix(base).unwrap().to_string_lossy().to_string();
            if rel == ".git" {
                continue;
            }
            let m = fs::symlink_metadata(&p).unwrap();
            if m.is_dir() {
                walk(base, &p, out);
            } else {
                out.push((rel, fs::read(&p).unwrap_or_default()));
            }
        }
    }
    let mut files = Vec::new();
    walk(dir, dir, &mut files);
    files.sort();
    let status = git(dir, &["status", "--porcelain", "--untracked-files=all"]);
    (files, status)
}

/// A slot path that is a link to another registered worktree is refused by every writing action
/// (prep with and without a lane, done, done --discard) before its first write: the worktree it
/// points at keeps every byte, its change and its untracked file.
#[cfg(unix)]
#[test]
fn every_writing_action_refuses_a_slot_path_that_is_a_link() {
    let r = repo(&[("a.txt", "a\n")]);
    old_file(&r.lane, "target/debug/built.txt", "lane build\n");
    let victim = r.root.parent().unwrap().join("victim");
    git(
        &r.root,
        &[
            "worktree",
            "add",
            "-q",
            "--detach",
            victim.to_str().unwrap(),
            &r.first,
        ],
    );
    fs::write(victim.join("a.txt"), "work in progress\n").unwrap();
    fs::write(victim.join("notes.txt"), "mine\n").unwrap();
    // A seed build, so a slot step that ran before the refusal would warm the link's target.
    let (code, v) = r.run_op("worktree-slot", &["ensure", "seed", "--at", &r.first]);
    assert_eq!(code, 0, "{v}");
    old_file(&r.slot("seed"), "target/debug/seed.txt", "seed build\n");
    std::os::unix::fs::symlink(&victim, r.slot("audit-1")).unwrap();
    let before = snapshot(&victim);
    let head_before = git(&victim, &["rev-parse", "HEAD"]);
    let worktrees = r.worktrees();
    let lane = r.lane_str().to_string();
    for args in [
        vec!["prep", "audit-1", "--head", &r.second],
        vec!["prep", "audit-1", "--head", &r.second, "--lane", &lane],
        vec!["done", "audit-1"],
        vec!["done", "audit-1", "--discard"],
    ] {
        r.refused(&args, &["audit-1", "is a link"]);
        assert_eq!(snapshot(&victim), before, "{args:?}");
        assert_eq!(
            git(&victim, &["rev-parse", "HEAD"]),
            head_before,
            "{args:?}"
        );
    }
    assert!(!victim.join("target").exists());
    assert_eq!(r.worktrees(), worktrees);
}

// ---- the inherited build is fresh (AC-3: REQ-4)

/// A crate with a dependency outside the repository (as a registry crate sits outside it),
/// built in the lane worktree, then prepared into an audit slot from the lane: the dependency
/// is not compiled again, and a one-line change compiles only the crate.
#[cfg(target_os = "macos")]
#[test]
fn a_slot_prepared_from_the_lane_does_not_compile_the_dependency() {
    use std::process::Command;
    let dep_dir = tempfile::tempdir().unwrap();
    let dep = fs::canonicalize(dep_dir.path()).unwrap();
    fs::write(
        dep.join("Cargo.toml"),
        "[package]\nname = \"auditdep\"\nversion = \"0.1.0\"\nedition = \"2021\"\n[lib]\npath = \"lib.rs\"\n",
    )
    .unwrap();
    fs::write(dep.join("lib.rs"), "pub fn one() -> u32 { 1 }\n").unwrap();
    let manifest = format!(
        "[package]\nname = \"auditapp\"\nversion = \"0.1.0\"\nedition = \"2021\"\n[workspace]\n[dependencies]\nauditdep = {{ path = \"{}\" }}\n",
        dep.display()
    );
    let r = repo(&[
        ("Cargo.toml", &manifest),
        (
            "src/main.rs",
            "fn main() { println!(\"{}\", auditdep::one()); }\n",
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
    let first = build(&r.lane);
    assert!(first.contains("Compiling auditdep"), "{first}");
    let prep = r.ok(&[
        "prep",
        "audit-1",
        "--head",
        &r.second,
        "--lane",
        r.lane_str(),
    ]);
    assert_eq!(prep["build"]["method"], json!("clone"), "{prep}");
    let slot = r.slot("audit-1");
    let inherited = build(&slot);
    assert!(inherited.contains("Fresh auditdep"), "{inherited}");
    assert!(!inherited.contains("Compiling auditdep"), "{inherited}");
    // A mutant: one line of the crate changes, only the crate compiles.
    fs::write(
        slot.join("src/main.rs"),
        "fn main() { println!(\"{}\", auditdep::one() + 1); }\n",
    )
    .unwrap();
    let mutant = build(&slot);
    assert!(mutant.contains("Fresh auditdep"), "{mutant}");
    assert!(mutant.contains("Compiling auditapp"), "{mutant}");
    let done = r.ok(&["done", "audit-1", "--discard"]);
    assert_eq!(done["reverted"], json!(["src/main.rs"]), "{done}");
}

// ---- the slot is its own worktree (AC-6: REQ-11)

/// A registered slot path whose folder is not the slot's own worktree is refused by every
/// writing action before its first write, in four shapes: a plain folder (git would walk up to
/// the primary checkout), another repository, a `.git` file that borrows the lane's git dir, and
/// a clean primary that `prep` must not detach.
#[test]
fn a_registered_slot_that_is_not_its_own_worktree_is_refused_and_nothing_changes() {
    let r = repo(&[("a.txt", "a\n")]);
    r.ok(&["prep", "audit-1", "--head", &r.first]);
    r.ok(&["done", "audit-1"]);
    let slot = r.slot("audit-1");
    let moved = r.root.parent().unwrap().join("moved-slot");
    let primary_branch = || git(&r.root, &["symbolic-ref", "-q", "HEAD"]);
    let primary_head = git(&r.root, &["rev-parse", "HEAD"]);
    let refuse_all = |words: &[&str]| {
        for args in [
            vec!["prep", "audit-1", "--head", r.second.as_str()],
            vec!["done", "audit-1"],
            vec!["done", "audit-1", "--discard"],
        ] {
            r.refused(&args, words);
        }
    };

    // 1. A plain folder at the registered path, the primary clean: prep must not detach it.
    fs::rename(&slot, &moved).unwrap();
    fs::create_dir_all(&slot).unwrap();
    refuse_all(&["audit-1", "not its own worktree"]);
    assert_eq!(primary_branch(), "refs/heads/main");
    assert_eq!(git(&r.root, &["rev-parse", "HEAD"]), primary_head);
    // ... and with a tracked change in the primary, which done --discard must not revert.
    fs::write(r.root.join("a.txt"), "primary work\n").unwrap();
    refuse_all(&["audit-1", "not its own worktree"]);
    assert_eq!(
        fs::read_to_string(r.root.join("a.txt")).unwrap(),
        "primary work\n"
    );
    assert_eq!(primary_branch(), "refs/heads/main");
    fs::write(r.root.join("a.txt"), "a\n").unwrap();
    fs::remove_dir(&slot).unwrap();

    // 2. Another repository at the registered path, with a change and an untracked file.
    git(
        &r.root,
        &["init", "-q", "-b", "other", slot.to_str().unwrap()],
    );
    fs::write(slot.join("x.txt"), "x\n").unwrap();
    git(&slot, &["add", "x.txt"]);
    commit(&slot, "other");
    fs::write(slot.join("x.txt"), "other work\n").unwrap();
    fs::write(slot.join("keep.txt"), "untracked\n").unwrap();
    let other = snapshot(&slot);
    refuse_all(&["audit-1", "not its own worktree", "repository"]);
    assert_eq!(snapshot(&slot), other);
    fs::remove_dir_all(&slot).unwrap();

    // 3. A slot `.git` that borrows the lane's git dir, the lane with staged content.
    fs::rename(&moved, &slot).unwrap();
    let lane_gitdir = git(
        &r.lane,
        &["rev-parse", "--path-format=absolute", "--git-dir"],
    );
    let own_git = fs::read_to_string(slot.join(".git")).unwrap();
    fs::write(slot.join(".git"), format!("gitdir: {lane_gitdir}\n")).unwrap();
    fs::write(r.lane.join("staged.txt"), "staged\n").unwrap();
    git(&r.lane, &["add", "staged.txt"]);
    let lane_index = git(&r.lane, &["diff", "--cached", "--name-only"]);
    refuse_all(&["audit-1", "not its own worktree", "belongs to the worktree"]);
    assert_eq!(
        git(&r.lane, &["diff", "--cached", "--name-only"]),
        lane_index
    );
    assert_eq!(
        fs::read_to_string(r.lane.join("staged.txt")).unwrap(),
        "staged\n"
    );
    assert_eq!(git(&r.lane, &["rev-parse", "HEAD"]), r.second);

    // Put back, the slot works again.
    fs::write(slot.join(".git"), own_git).unwrap();
    r.ok(&["prep", "audit-1", "--head", &r.second]);
    detached_at(&slot, &r.second);
}

/// A folder above the slot that is a link (`.claude/worktrees` pointing elsewhere) is refused.
#[cfg(unix)]
#[test]
fn a_link_above_the_slot_is_refused() {
    let r = repo(&[("a.txt", "a\n")]);
    let elsewhere = r.root.parent().unwrap().join("elsewhere");
    fs::create_dir_all(&elsewhere).unwrap();
    fs::create_dir_all(r.root.join(".claude")).unwrap();
    std::os::unix::fs::symlink(&elsewhere, r.root.join(".claude/worktrees")).unwrap();
    let before = r.worktrees();
    r.refused(
        &["prep", "audit-1", "--head", &r.first],
        &[".claude/worktrees above it is a link"],
    );
    r.refused(&["done", "audit-1"], &["above it is a link"]);
    assert_eq!(r.worktrees(), before);
    assert_eq!(fs::read_dir(&elsewhere).unwrap().count(), 0);
}

// ---- one holder per slot (AC-7: REQ-12)

#[test]
fn a_slot_has_one_live_holder_and_done_releases_it() {
    let r = repo(&[("a.txt", "a\n")]);
    let a = r.ok(&[
        "prep",
        "audit-1",
        "--head",
        &r.first,
        "--holder",
        "auditor-a",
    ]);
    assert_eq!(a["holder"]["holder"], json!("auditor-a"), "{a}");
    assert_eq!(a["took_over"], json!(null), "{a}");
    // Another holder is refused, by prep and by done; the slot stays at the first head.
    let m = r.refused(
        &[
            "prep",
            "audit-1",
            "--head",
            &r.second,
            "--holder",
            "auditor-b",
        ],
        &["held by", "auditor-a"],
    );
    assert!(m.contains(&r.first), "{m}");
    r.refused(
        &["done", "audit-1", "--discard", "--holder", "auditor-b"],
        &["held by", "auditor-a"],
    );
    assert_eq!(git(&r.slot("audit-1"), &["rev-parse", "HEAD"]), r.first);
    // The same holder may prep again.
    r.ok(&[
        "prep",
        "audit-1",
        "--head",
        &r.second,
        "--holder",
        "auditor-a",
    ]);
    // auto takes the first slot without a live holder, until all are held.
    let b = r.ok(&["prep", "auto", "--head", &r.second, "--holder", "auditor-b"]);
    assert_eq!(b["slot"], json!("audit-2"), "{b}");
    let c = r.ok(&["prep", "auto", "--head", &r.second, "--holder", "auditor-c"]);
    assert_eq!(c["slot"], json!("audit-3"), "{c}");
    r.refused(
        &["prep", "auto", "--head", &r.second, "--holder", "auditor-d"],
        &[
            "all audit slots are held",
            "auditor-a",
            "auditor-b",
            "auditor-c",
        ],
    );
    // done releases: the next holder takes the slot.
    let done = r.ok(&["done", "audit-2", "--holder", "auditor-b"]);
    assert_eq!(done["released"], json!(true), "{done}");
    let d = r.ok(&["prep", "auto", "--head", &r.first, "--holder", "auditor-d"]);
    assert_eq!(d["slot"], json!("audit-2"), "{d}");
}

#[test]
fn a_dead_holder_is_taken_over_with_a_line() {
    let r = repo(&[("a.txt", "a\n")]);
    // A holder whose process has ended.
    let mut child = std::process::Command::new("true").spawn().unwrap();
    let pid = child.id().to_string();
    child.wait().unwrap();
    r.ok(&[
        "prep",
        "audit-1",
        "--head",
        &r.first,
        "--holder",
        "auditor-a",
        "--pid",
        &pid,
    ]);
    let b = r.ok(&[
        "prep",
        "audit-1",
        "--head",
        &r.second,
        "--holder",
        "auditor-b",
    ]);
    assert_eq!(b["took_over"]["holder"], json!("auditor-a"), "{b}");
    detached_at(&r.slot("audit-1"), &r.second);
    // A holder whose lease has ended.
    r.ok(&[
        "prep",
        "audit-2",
        "--head",
        &r.first,
        "--holder",
        "auditor-c",
        "--lease-secs",
        "1",
    ]);
    std::thread::sleep(Duration::from_millis(2100));
    let done = r.ok(&["done", "audit-2", "--holder", "auditor-d"]);
    assert_eq!(done["took_over"]["holder"], json!("auditor-c"), "{done}");
    assert_eq!(done["released"], json!(true), "{done}");
    for bad in ["0", "86401"] {
        r.refused(
            &["prep", "audit-3", "--head", &r.first, "--lease-secs", bad],
            &["--lease-secs"],
        );
    }
    r.refused(
        &["prep", "audit-3", "--head", &r.first, "--holder", "a b"],
        &["--holder"],
    );
}

// ---- the slot lock (REQ-9)

#[test]
fn prep_and_done_wait_for_the_slot_lock_then_refuse() {
    let r = repo(&[("a.txt", "a\n")]);
    r.ok(&["prep", "audit-1", "--head", &r.first]);
    let slot = r.slot("audit-1");
    fs::write(slot.join("a.txt"), "mutant\n").unwrap();
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
    let env = [("MOAI_WORKTREE_SLOT_LOCK_WAIT_SECS", "1")];
    for args in [
        vec!["prep", "audit-1", "--head", r.second.as_str()],
        vec!["prep", "audit-2", "--head", r.second.as_str()],
        vec!["done", "audit-1", "--discard"],
    ] {
        let (code, v) = r.run_op_env("audit-slot", &args, &env);
        assert_ne!(code, 0, "{args:?}: {v}");
        let m = v["error"]["message"].as_str().unwrap();
        assert!(
            m.contains("held the slot lock for 1 seconds"),
            "{args:?}: {m}"
        );
    }
    assert_eq!(fs::read_to_string(slot.join("a.txt")).unwrap(), "mutant\n");
    assert!(!r.slot("audit-2").exists());
    held.unlock().unwrap();
    r.ok(&["done", "audit-1", "--discard"]);
}

// ---- leftovers and refusal kinds

#[test]
fn a_copy_a_killed_prep_left_is_named_once_and_removed_whole() {
    let r = repo(&[("a.txt", "a\n")]);
    r.ok(&["prep", "audit-1", "--head", &r.first]);
    let slot = r.slot("audit-1");
    for f in ["debug/a", "debug/b", "debug/deps/c"] {
        let p = slot.join(".target-inherit-4242").join(f);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, "x").unwrap();
    }
    let m = r.refused(
        &["prep", "audit-1", "--head", &r.second],
        &[".target-inherit-4242", "--discard"],
    );
    assert!(!m.contains("debug/a"), "{m}");
    let done = r.ok(&["done", "audit-1", "--discard"]);
    assert_eq!(done["removed"], json!([".target-inherit-4242"]), "{done}");
    assert!(!slot.join(".target-inherit-4242").exists());
}

#[test]
fn a_cap_refusal_is_invalid_input_and_adds_nothing() {
    let r = repo(&[("a.txt", "a\n")]);
    let before = r.worktrees();
    r.refused(
        &[
            "prep",
            "audit-1",
            "--head",
            &r.first,
            "--max-worktrees",
            "1",
        ],
        &["audit-1", "cap reached"],
    );
    r.refused(
        &[
            "prep",
            "audit-1",
            "--head",
            &r.first,
            "--min-free-percent",
            "100",
        ],
        &["audit-1", "under the floor"],
    );
    assert_eq!(r.worktrees(), before);
}

// ---- the inherited build is the head's (AC-8: REQ-13)

/// A lane builds a mutant, then reverts the source: its HEAD is the audited head and it has no
/// tracked change, but its `target/` holds the mutant's binary, which Cargo would call `Fresh`
/// in a slot whose sources are older. After prep the slot's binary is the head's.
#[cfg(unix)]
#[test]
fn an_inherited_mutant_build_is_compiled_again_from_the_head() {
    use std::process::Command;
    let r = repo(&[
        (
            "Cargo.toml",
            "[package]\nname = \"auditbin\"\nversion = \"0.1.0\"\nedition = \"2021\"\n[workspace]\n",
        ),
        ("src/main.rs", "fn main() { println!(\"CLEAN\"); }\n"),
    ]);
    let cargo = std::env::var("CARGO").unwrap_or_else(|_| "cargo".into());
    let run = |dir: &Path| -> String {
        let out = Command::new(&cargo)
            .args(["run", "-q", "--offline"])
            .current_dir(dir)
            .env_remove("CARGO_TARGET_DIR")
            .env_remove("CARGO_BUILD_TARGET_DIR")
            .env_remove("RUSTC_WRAPPER")
            .env_remove("CARGO_BUILD_RUSTC_WRAPPER")
            .output()
            .unwrap();
        assert!(
            out.status.success(),
            "{}",
            String::from_utf8_lossy(&out.stderr)
        );
        String::from_utf8_lossy(&out.stdout).trim().to_string()
    };
    // The slot exists first, its sources checked out before the lane's mutant build.
    r.ok(&["prep", "audit-1", "--head", &r.second]);
    r.ok(&["done", "audit-1"]);
    std::thread::sleep(Duration::from_millis(1100));
    let main = r.lane.join("src/main.rs");
    fs::write(&main, "fn main() { println!(\"MUTANT\"); }\n").unwrap();
    assert_eq!(run(&r.lane), "MUTANT");
    git(&r.lane, &["checkout", "--", "src/main.rs"]);
    // No tracked change (Cargo.lock is new and untracked).
    assert_eq!(
        git(&r.lane, &["status", "--porcelain", "--untracked-files=no"]),
        ""
    );
    let prep = r.ok(&[
        "prep",
        "audit-1",
        "--head",
        &r.second,
        "--lane",
        r.lane_str(),
    ]);
    assert_eq!(prep["build"]["source"], json!("lane"), "{prep}");
    assert!(
        prep["build"]["sources_touched"].as_u64().unwrap() >= 3,
        "{prep}"
    );
    assert_eq!(run(&r.slot("audit-1")), "CLEAN");
}
