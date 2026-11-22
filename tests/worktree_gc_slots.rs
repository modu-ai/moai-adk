//! SPEC-20261123-worktree-slots slice 2 (option (b) of leader decision on 6852): rules in
//! `worktree-gc-plan` that only skip or report — nothing new becomes removable.
//! - a worktree holding another linked worktree inside it is skipped (`nested_worktree`);
//! - a worktree a held card reservation names is skipped (`live_claim`);
//! - a fixed slot under `<primary>/.claude/worktrees/` names `fixed_slot` beside the
//!   preservation it already has;
//! - a `cap` block orders the removal candidates oldest first and says how many bring the
//!   count back under the worktree cap.
//!
//! Every worktree but the slots sits in a folder of its own outside the project, where no
//! preservation rule applies, so the control worktree is a `remove` candidate.
mod common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Output;
use std::time::{SystemTime, UNIX_EPOCH};

const SESSION: &str = "aaaaaaaa-0000-4000-8000-00000000000a";

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

fn cargo_output(wt: &Path) {
    fs::create_dir_all(wt.join("target/debug/deps")).unwrap();
    fs::write(
        wt.join("target/CACHEDIR.TAG"),
        "Signature: 8a477f597d28d172789f06886806bc55\n",
    )
    .unwrap();
    fs::write(wt.join("target/debug/.cargo-lock"), "").unwrap();
    fs::write(wt.join("target/debug/deps/lib.rlib"), vec![7u8; 4096]).unwrap();
}

/// Three days after now: every change in the fixture is old.
fn later() -> u64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap()
        .as_secs()
        + 3 * 86_400
}

struct Fx {
    home: TempHome,
    project: PathBuf,
    wts: PathBuf,
}

/// A registered project that is a git repository with one commit and the branch `integ`.
fn fixture() -> Fx {
    let home = TempHome::new();
    let project = fs::canonicalize(home.project("demo")).unwrap();
    ledger::register(home.path(), "demo").unwrap();
    git(&project, &["init", "-q", "-b", "main"]);
    fs::write(
        project.join(".gitignore"),
        format!(
            "/{}/\n/target/\n/.claude/\n/inner/\n",
            moai_cli::PROJECT_STATE_DIR
        ),
    )
    .unwrap();
    fs::write(project.join("a.txt"), "a\n").unwrap();
    git(&project, &["add", ".gitignore", "a.txt"]);
    git(&project, &["commit", "-q", "-m", "init"]);
    git(&project, &["branch", "integ"]);
    let wts = fs::canonicalize(home.path()).unwrap().join("wts");
    fs::create_dir_all(&wts).unwrap();
    Fx { home, project, wts }
}

impl Fx {
    /// A merged, clean worktree holding only cargo output, detached at `integ`.
    fn worktree(&self, at: &Path) -> PathBuf {
        git(
            &self.project,
            &[
                "worktree",
                "add",
                "-q",
                "--detach",
                at.to_str().unwrap(),
                "integ",
            ],
        );
        cargo_output(at);
        at.to_path_buf()
    }

    fn moai(&self, args: &[&str]) -> Output {
        self.home
            .command(args)
            .env("CLAUDE_CODE_SESSION_ID", SESSION)
            // `lsof`, which the process check runs, is found on the caller's PATH.
            .env("PATH", std::env::var_os("PATH").unwrap_or_default())
            .output()
            .expect("run moai")
    }

    fn plan(&self, extra: &[&str]) -> Value {
        let later = later().to_string();
        let mut args = vec![
            "worktree-gc-plan",
            self.project.to_str().unwrap(),
            "--integration-ref",
            "integ",
            "--now-unix",
            &later,
        ];
        args.extend_from_slice(extra);
        let out = self.moai(&args);
        let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
            panic!(
                "not JSON: {} / {}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        });
        assert_eq!(v["ok"], json!(true), "{v}");
        v["result"].clone()
    }

    /// Takes a new card whose reservation names `worktree`.
    fn claim(&self, worktree: &Path) {
        let project = self.project.to_str().unwrap();
        let added = self.moai(&[
            "card",
            "add",
            "--title",
            "t",
            "--scope",
            "src/a.rs",
            "--project",
            project,
        ]);
        let added: Value = serde_json::from_slice(&added.stdout).unwrap();
        let name = added["result"]["name"].as_str().unwrap().to_string();
        let ready = self.moai(&["card", "ready", &name, "--project", project]);
        assert!(
            ready.status.success(),
            "{}",
            String::from_utf8_lossy(&ready.stdout)
        );
        let taken = self.moai(&[
            "card",
            "take",
            &name,
            "--lease-secs",
            "3600",
            "--worktree",
            worktree.to_str().unwrap(),
            "--project",
            project,
        ]);
        assert!(
            taken.status.success(),
            "{}",
            String::from_utf8_lossy(&taken.stdout)
        );
    }
}

fn entry<'a>(plan: &'a Value, path: &Path) -> &'a Value {
    let want = path.to_str().unwrap();
    plan["candidates"]
        .as_array()
        .unwrap()
        .iter()
        .find(|c| {
            let p = c["path"].as_str().unwrap();
            p == want || want.ends_with(&format!("/{p}"))
        })
        .unwrap_or_else(|| panic!("no candidate {want}: {plan}"))
}

fn reasons(e: &Value) -> Vec<String> {
    e["reasons"]
        .as_array()
        .unwrap()
        .iter()
        .map(|r| r.as_str().unwrap().to_string())
        .collect()
}

#[test]
fn a_clean_merged_worktree_outside_any_preserved_place_is_the_control_remove() {
    let fx = fixture();
    let free = fx.worktree(&fx.wts.join("free"));
    let plan = fx.plan(&[]);
    let e = entry(&plan, &free);
    assert_eq!(
        (&e["action"], reasons(e).len()),
        (&json!("remove"), 0),
        "{e}"
    );
}

// ---- REQ-8: nested worktree and live claim

#[test]
fn a_worktree_holding_another_linked_worktree_is_skipped_as_nested() {
    let fx = fixture();
    let outer = fx.worktree(&fx.wts.join("outer"));
    let inner = fx.worktree(&outer.join("inner"));
    let plan = fx.plan(&[]);
    let e = entry(&plan, &outer);
    assert_eq!(e["action"], json!("skip"), "{e}");
    assert!(reasons(e).contains(&"nested_worktree".to_string()), "{e}");
    // The inner worktree holds nothing nested: it is judged on its own facts.
    let i = entry(&plan, &inner);
    assert!(!reasons(i).contains(&"nested_worktree".to_string()), "{i}");
}

#[test]
fn a_worktree_a_held_card_reservation_names_is_skipped_as_a_live_claim() {
    let fx = fixture();
    let claimed = fx.worktree(&fx.wts.join("claimed"));
    let free = fx.worktree(&fx.wts.join("free"));
    fx.claim(&claimed);
    let plan = fx.plan(&[]);
    let e = entry(&plan, &claimed);
    assert_eq!(e["action"], json!("skip"), "{e}");
    assert_eq!(reasons(e), ["live_claim"], "{e}");
    assert_eq!(entry(&plan, &free)["action"], json!("remove"));
}

#[test]
fn a_released_reservation_no_longer_keeps_its_worktree() {
    let fx = fixture();
    let claimed = fx.worktree(&fx.wts.join("claimed"));
    fx.claim(&claimed);
    let project = fx.project.to_str().unwrap();
    let out = fx.moai(&[
        "card",
        "release",
        "C-1",
        "--generation",
        "1",
        "--project",
        project,
    ]);
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    let plan = fx.plan(&[]);
    assert_eq!(entry(&plan, &claimed)["action"], json!("remove"));
}

// ---- REQ-7: fixed slots

#[test]
fn a_fixed_slot_names_fixed_slot_and_another_name_does_not() {
    let fx = fixture();
    let slot = fx.worktree(&fx.project.join(".claude/worktrees/lane-01"));
    let other = fx.worktree(&fx.project.join(".claude/worktrees/lane-99"));
    let plan = fx.plan(&[]);
    let s = entry(&plan, &slot);
    assert_eq!(s["action"], json!("skip"), "{s}");
    assert!(reasons(s).contains(&"fixed_slot".to_string()), "{s}");
    assert!(
        reasons(s).contains(&"preserved_location".to_string()),
        "{s}"
    );
    let o = entry(&plan, &other);
    assert!(!reasons(o).contains(&"fixed_slot".to_string()), "{o}");
    // A slot name in a folder that is not this project's slot folder is not a slot.
    let away = fx.worktree(&fx.wts.join("lane-02"));
    let plan = fx.plan(&[]);
    assert_eq!(entry(&plan, &away)["action"], json!("remove"));
}

// ---- REQ-10: the cap block

#[test]
fn the_cap_block_orders_removal_candidates_oldest_first_and_counts_what_is_needed() {
    let fx = fixture();
    let a = fx.worktree(&fx.wts.join("a"));
    let b = fx.worktree(&fx.wts.join("b"));
    let c = fx.worktree(&fx.wts.join("c"));
    // `b` is the oldest, `c` the newest: the cargo files carry the change times.
    // The plan reads a change's age from the worktree folder, `target/`, the cargo folders and
    // the worktree's git index, HEAD and HEAD log; each is set back by `days`.
    let old = |p: &Path, days: u64| {
        let gitdir = PathBuf::from(git(p, &["rev-parse", "--absolute-git-dir"]));
        let when = SystemTime::now() - std::time::Duration::from_secs(days * 86_400);
        for f in [
            p.to_path_buf(),
            p.join("target"),
            p.join("target/debug"),
            p.join("target/debug/deps"),
            p.join("target/CACHEDIR.TAG"),
            gitdir.join("index"),
            gitdir.join("HEAD"),
            gitdir.join("logs/HEAD"),
        ] {
            if f.exists() {
                fs::File::open(&f).unwrap().set_modified(when).unwrap();
            }
        }
    };
    old(&b, 5);
    old(&a, 4);
    let plan = fx.plan(&["--max-worktrees", "1"]);
    let cap = &plan["cap"];
    assert_eq!(cap["limit"], json!(1), "{cap}");
    assert_eq!(cap["count"], json!(3), "{cap}");
    assert_eq!(cap["over"], json!(2), "{cap}");
    let shown = |p: &Path| entry(&plan, p)["path"].clone();
    assert_eq!(
        cap["oldest_first"],
        json!([shown(&b), shown(&a), shown(&c)]),
        "{cap}"
    );
    assert_eq!(cap["needed"], json!(2), "{cap}");
    assert_eq!(cap["enough"], json!(true), "{cap}");
    // Exactly as many candidates as the count is over: still enough.
    let zero = fx.plan(&["--max-worktrees", "0"]);
    assert_eq!(
        (
            &zero["cap"]["over"],
            &zero["cap"]["needed"],
            &zero["cap"]["enough"]
        ),
        (&json!(3), &json!(3), &json!(true)),
        "{}",
        zero["cap"]
    );
    // Under the cap: nothing is over and nothing is needed. The default cap is 25.
    let plan = fx.plan(&[]);
    assert_eq!(
        (
            &plan["cap"]["limit"],
            &plan["cap"]["over"],
            &plan["cap"]["needed"]
        ),
        (&json!(25), &json!(0), &json!(0)),
        "{}",
        plan["cap"]
    );
}

#[test]
fn the_cap_block_says_when_the_candidates_are_not_enough() {
    let fx = fixture();
    fx.worktree(&fx.wts.join("a"));
    let kept = fx.worktree(&fx.wts.join("kept"));
    fs::write(kept.join("a.txt"), "changed\n").unwrap();
    let plan = fx.plan(&["--max-worktrees", "0"]);
    let cap = &plan["cap"];
    assert_eq!(
        (&cap["over"], &cap["needed"]),
        (&json!(2), &json!(1)),
        "{cap}"
    );
    assert_eq!(cap["enough"], json!(false), "{cap}");
    // The cap block is a report: the digest binds only the candidates.
    let again = fx.plan(&["--max-worktrees", "1"]);
    assert_eq!(plan["plan_digest"], again["plan_digest"]);
}

/// A ledger that exists but cannot be read rules no claim out: every linked worktree is kept.
#[test]
fn an_unreadable_ledger_keeps_every_linked_worktree() {
    let fx = fixture();
    let free = fx.worktree(&fx.wts.join("free"));
    let file = ledger::ledger_path(fx.home.path(), "demo").unwrap();
    fs::write(&file, b"not a ledger").unwrap();
    let plan = fx.plan(&[]);
    let e = entry(&plan, &free);
    assert_eq!(e["action"], json!("skip"), "{e}");
    assert_eq!(reasons(e), ["claim_check_unavailable"], "{e}");
    // The primary is never a candidate, so it does not name the unreadable ledger.
    assert!(!reasons(entry(&plan, Path::new("."))).contains(&"claim_check_unavailable".to_string()));
}
