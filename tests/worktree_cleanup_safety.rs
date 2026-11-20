//! SPEC-20261120-worktree-cap-and-cleanup, the safety half: what the preview and the apply must
//! never remove, probed on temporary repositories (the audit's findings, each as a test).
//!
//! - a finished card does not make unmerged commits of a detached worktree safe to lose;
//! - a path with a newline in it cannot turn a neighbour into a removable row;
//! - an empty worktree under a day old is kept, as the gc plan keeps it;
//! - the newest reservation names the card's state;
//! - an unreadable ledger and an install folder each keep the worktrees.
//!
//! Every command runs against a temporary repository and a temporary HOME.
#![cfg(unix)]
mod common;

use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output};

const BIN: &str = crate::common::MOAI_BIN;
const DAY: u64 = 86_400;

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

fn fx() -> Fx {
    let dir = tempfile::TempDir::new().unwrap();
    let base = fs::canonicalize(dir.path()).unwrap();
    let (root, home) = (base.join("repo"), base.join("home"));
    fs::create_dir_all(&root).unwrap();
    fs::create_dir_all(&home).unwrap();
    git(&root, &["init", "-q", "-b", "main"]);
    fs::write(
        root.join(".gitignore"),
        ".moai-cli/\n.claude/\ntarget/\n*.env\n",
    )
    .unwrap();
    fs::write(root.join("a.txt"), "a\n").unwrap();
    git(&root, &["add", ".gitignore", "a.txt"]);
    git(&root, &["commit", "-q", "-m", "init"]);
    git(&root, &["branch", "integ"]);
    fs::create_dir_all(root.join(".moai-cli")).unwrap();
    fs::write(
        root.join(".moai-cli/project.yaml"),
        r#"{"project_id":"demo"}"#,
    )
    .unwrap();
    moai_cli::ledger::register(&home, "demo").expect("register ledger");
    Fx {
        _dir: dir,
        root,
        home,
    }
}

/// A linked worktree at `rel` (relative to the repository, or absolute) on a new branch cut from
/// `integ`; with `own_commit` the branch carries one commit `integ` does not have.
fn add(fx: &Fx, rel: &str, branch: &str, own_commit: bool) -> PathBuf {
    let path = if Path::new(rel).is_absolute() {
        PathBuf::from(rel)
    } else {
        fx.root.join(rel)
    };
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    git(
        &fx.root,
        &[
            "worktree",
            "add",
            "-q",
            "-b",
            branch,
            path.to_str().unwrap(),
            "integ",
        ],
    );
    if own_commit {
        fs::write(path.join("own.txt"), format!("{branch}\n")).unwrap();
        git(&path, &["add", "own.txt"]);
        git(&path, &["commit", "-q", "-m", branch]);
    }
    fs::canonicalize(path).unwrap()
}

/// One card of the ledger reserving `wt` on `branch`.
fn seed_card(fx: &Fx, wt: &Path, branch: &str, state: &str, status: &str, lease_until: u64) {
    let l = moai_cli::ledger::open_write(&fx.home, "demo").expect("open ledger");
    let conn = l.conn();
    conn.execute(
        "INSERT INTO card (title, state, created_by, created_at, updated_at, closed_at) \
         VALUES (?1, ?2, 'fixture', 1, 1, ?3)",
        rusqlite::params![
            format!("card for {branch}"),
            state,
            (state == "done").then_some(1)
        ],
    )
    .unwrap();
    let id = conn.last_insert_rowid();
    conn.execute(
        "INSERT INTO reservation (card_id, generation, holder, status, acquired_at, lease_until, \
         updated_at, worktree, branch) VALUES (?1, 1, 'fixture-session', ?2, 1, ?3, 1, ?4, ?5)",
        rusqlite::params![id, status, lease_until as i64, wt.to_str().unwrap(), branch],
    )
    .unwrap();
    drop(l);
}

fn real_now() -> u64 {
    std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap()
        .as_secs()
}

fn run(fx: &Fx, args: &[&str]) -> Output {
    Command::new(BIN)
        .args(args)
        .current_dir(&fx.root)
        .env("HOME", &fx.home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .output()
        .unwrap()
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)))
}

/// The preview as of `now` (seconds); every file of a fixture was just written, so `now` is also
/// how old the worktrees look.
fn plan_at(fx: &Fx, now: u64) -> Value {
    let v = plan_raw(fx, now);
    assert_eq!(v["ok"], true, "{v}");
    v["result"]["cleanup"].clone()
}

fn plan_raw(fx: &Fx, now: u64) -> Value {
    let now = now.to_string();
    envelope(&run(
        fx,
        &[
            "optimize-plan",
            fx.root.to_str().unwrap(),
            "--integration-ref",
            "integ",
            "--now-unix",
            &now,
        ],
    ))
}

fn apply_at(fx: &Fx, now: u64, picks: &[String]) -> Value {
    let mut args = vec![
        "optimize-apply".to_string(),
        fx.root.to_str().unwrap().into(),
        "--integration-ref".into(),
        "integ".into(),
        "--now-unix".into(),
        now.to_string(),
        "--permission-reference".into(),
        "test-permission".into(),
    ];
    for p in picks {
        args.push("--pick".into());
        args.push(p.clone());
    }
    let refs: Vec<&str> = args.iter().map(String::as_str).collect();
    envelope(&run(fx, &refs))
}

fn row<'a>(list: &'a Value, tail: &str) -> Option<&'a Value> {
    list.as_array()?
        .iter()
        .find(|r| r["path"].as_str().is_some_and(|p| p.ends_with(tail)))
}

fn reasons(row: &Value) -> Vec<String> {
    row["reasons"]
        .as_array()
        .map(|a| {
            a.iter()
                .filter_map(|r| r.as_str().map(String::from))
                .collect()
        })
        .unwrap_or_default()
}

fn pick(cleanup: &Value, tail: &str) -> String {
    let c = row(&cleanup["candidates"], tail)
        .unwrap_or_else(|| panic!("{tail} is not a candidate: {cleanup}"));
    format!(
        "{}@{}",
        c["path"].as_str().unwrap(),
        c["item_digest"].as_str().unwrap()
    )
}

fn removed(v: &Value) -> Vec<String> {
    v["result"]["removed"]
        .as_array()
        .map(|a| {
            a.iter()
                .filter_map(|r| r["path"].as_str().map(String::from))
                .collect()
        })
        .unwrap_or_default()
}

fn commit_exists(dir: &Path, sha: &str) -> bool {
    Command::new("git")
        .args(["-C", dir.to_str().unwrap(), "cat-file", "-e"])
        .arg(format!("{sha}^{{commit}}"))
        .status()
        .unwrap()
        .success()
}

/// A finished card makes a worktree a candidate even when its branch is not merged, because the
/// branch stays after the worktree goes. A detached worktree has no branch: its commits are
/// safe only while some ref reaches them.
#[test]
fn a_finished_cards_detached_worktree_with_unreachable_commits_is_kept() {
    let fx = fx();
    let later = real_now() + 30 * DAY;
    let w = add(&fx, ".moai-cli/worktrees/det", "b-det", false);
    git(&w, &["checkout", "-q", "--detach"]);
    fs::write(w.join("lost.txt"), "only here\n").unwrap();
    git(&w, &["add", "lost.txt"]);
    git(&w, &["commit", "-q", "-m", "only on this worktree"]);
    let head = git(&w, &["rev-parse", "HEAD"]).trim().to_string();
    git(&fx.root, &["branch", "-D", "b-det"]);
    seed_card(&fx, &w, "b-det", "done", "released", 1);

    let cleanup = plan_at(&fx, later);
    assert!(
        row(&cleanup["candidates"], "worktrees/det").is_none(),
        "{cleanup}"
    );
    let kept = row(&cleanup["kept"], "worktrees/det").expect("kept");
    assert!(
        reasons(kept).contains(&"unmerged_detached_commits".to_string()),
        "{kept}"
    );
    // asked by name with a made-up digest, the apply refuses and the commit is still there
    let res = apply_at(
        &fx,
        later,
        &[format!("{}@made-up", kept["path"].as_str().unwrap())],
    );
    assert_eq!(removed(&res), Vec::<String>::new(), "{res}");
    assert!(w.is_dir(), "the worktree must stay");
    assert!(commit_exists(&fx.root, &head));

    // One ref that reaches the commit is enough: the worktree is then a finished-card candidate,
    // and removing it leaves the commit on that ref.
    git(&fx.root, &["branch", "kept-commit", &head]);
    let cleanup = plan_at(&fx, later);
    let c = row(&cleanup["candidates"], "worktrees/det").expect("candidate once a ref has it");
    assert_eq!(c["remove_because"], "card_done", "{c}");
    let res = apply_at(&fx, later, &[pick(&cleanup, "worktrees/det")]);
    assert_eq!(removed(&res).len(), 1, "{res}");
    assert!(!w.exists());
    assert!(commit_exists(&fx.root, &head));
}

/// The same finished card on a worktree that has a branch: the branch keeps the unmerged commit
/// after the removal, so the worktree is a candidate and the commit survives.
#[test]
fn a_finished_cards_branch_worktree_is_removed_and_its_branch_keeps_the_commit() {
    let fx = fx();
    let later = real_now() + 30 * DAY;
    let w = add(&fx, ".moai-cli/worktrees/br", "b-br", true);
    let head = git(&w, &["rev-parse", "HEAD"]).trim().to_string();
    seed_card(&fx, &w, "b-br", "done", "released", 1);
    let cleanup = plan_at(&fx, later);
    let res = apply_at(&fx, later, &[pick(&cleanup, "worktrees/br")]);
    assert_eq!(removed(&res).len(), 1, "{res}");
    assert!(!w.exists());
    assert!(commit_exists(&fx.root, &head));
    assert_eq!(
        git(&fx.root, &["rev-parse", "b-br"]).trim(),
        head,
        "the branch stays on the commit"
    );
}

/// A worktree whose path holds a newline breaks plain `git worktree list --porcelain` into a
/// second record that names a neighbour; with the neighbour's unmerged commit on the line, the
/// fake record would have it removed. Neither the neighbour nor the odd path may be a candidate.
#[test]
fn a_path_with_a_newline_cannot_make_a_neighbour_a_candidate() {
    let fx = fx();
    let later = real_now() + 30 * DAY;
    let neighbour = add(&fx, ".moai-cli/worktrees/neighbour", "b-neighbour", true);
    let head = git(&neighbour, &["rev-parse", "HEAD"]).trim().to_string();
    let evil = format!(
        "{}/.moai-cli/worktrees/evil\nworktree {}",
        fx.root.display(),
        neighbour.display()
    );
    add(&fx, &evil, "b-evil", false);

    let cleanup = plan_at(&fx, later);
    assert!(
        row(&cleanup["candidates"], "worktrees/neighbour").is_none(),
        "the neighbour must not be a candidate: {cleanup}"
    );
    assert!(
        cleanup["candidates"]
            .as_array()
            .unwrap()
            .iter()
            .all(|c| !c["path"].as_str().unwrap().contains('\n')),
        "{cleanup}"
    );
    let res = apply_at(&fx, later, &[format!("{}@made-up", neighbour.display())]);
    assert_eq!(removed(&res), Vec::<String>::new(), "{res}");
    assert!(neighbour.is_dir(), "the neighbour must stay");
    assert!(commit_exists(&fx.root, &head));
    // the odd path is kept: the plain listing cuts it at the newline, so the cut path is a path
    // the NUL-separated listing does not know (or the path itself holds a control character)
    let odd = cleanup["kept"]
        .as_array()
        .unwrap()
        .iter()
        .find(|k| k["path"].as_str().unwrap().contains("evil"))
        .expect("the odd path is listed as kept");
    let why = reasons(odd);
    assert!(
        why.contains(&"listing_mismatch".to_string())
            || why.contains(&"path_not_plain".to_string()),
        "{odd}"
    );
}

/// The gc plan keeps an empty merged worktree for a day (a lane may have just made it); the
/// cleanup keeps it too, and offers it after the day.
#[test]
fn an_empty_worktree_under_a_day_old_is_kept() {
    let fx = fx();
    add(&fx, ".moai-cli/worktrees/fresh", "b-fresh", false);
    let t = real_now();
    // three hours old: past the recent-change window, under the day
    let young = plan_at(&fx, t + 3 * 3_600);
    assert!(
        row(&young["candidates"], "worktrees/fresh").is_none(),
        "{young}"
    );
    let kept = row(&young["kept"], "worktrees/fresh").expect("kept");
    assert!(
        reasons(kept).contains(&"empty_worktree".to_string()),
        "{kept}"
    );
    // a day and an hour old
    let old = plan_at(&fx, t + DAY + 3_600);
    let c = row(&old["candidates"], "worktrees/fresh").expect("candidate after a day");
    assert_eq!(c["remove_because"], "merged", "{c}");
}

/// The newest reservation names the card's state, whatever the order the ledger lists them in:
/// an older reservation on the branch of an active card does not outvote a newer one on the
/// path of a finished card.
#[test]
fn the_newest_reservation_names_the_card_state() {
    let fx = fx();
    let later = real_now() + 30 * DAY;
    let w = add(&fx, ".moai-cli/worktrees/newest", "b-newest", true);
    // older: the branch, an active card (its reservation ended)
    seed_card(
        &fx,
        Path::new("/nowhere/else"),
        "b-newest",
        "active",
        "released",
        1,
    );
    // newer: the path, a finished card
    seed_card(&fx, &w, "b-other", "done", "released", 1);
    let cleanup = plan_at(&fx, later);
    let c = row(&cleanup["candidates"], "worktrees/newest")
        .unwrap_or_else(|| panic!("the newest reservation is a finished card: {cleanup}"));
    assert_eq!(c["remove_because"], "card_done", "{c}");
}

/// A ledger that is there and cannot be read leaves no candidate: no guess.
#[test]
fn an_unreadable_ledger_leaves_no_candidate() {
    let fx = fx();
    let later = real_now() + 30 * DAY;
    let w = add(&fx, ".moai-cli/worktrees/done", "b-done", true);
    seed_card(&fx, &w, "b-done", "done", "released", 1);
    let before = plan_at(&fx, later);
    assert!(
        row(&before["candidates"], "worktrees/done").is_some(),
        "{before}"
    );
    let ledger = fx.home.join(".moai/projects/demo/ledger.db");
    assert!(ledger.is_file(), "{}", ledger.display());
    fs::write(&ledger, b"this is not a database\n").unwrap();
    let v = plan_raw(&fx, later);
    if v["ok"] == json!(true) {
        let cleanup = &v["result"]["cleanup"];
        assert!(
            cleanup["candidates"].as_array().unwrap().is_empty(),
            "{cleanup}"
        );
    }
    let res = apply_at(&fx, later, &[pick(&before, "worktrees/done")]);
    assert_eq!(removed(&res), Vec::<String>::new(), "{res}");
    assert!(w.is_dir(), "the worktree must stay");
}

/// A real `node_modules` folder keeps the worktree, whatever else is true of it.
#[test]
fn an_install_folder_keeps_the_worktree() {
    let fx = fx();
    let later = real_now() + 30 * DAY;
    fs::create_dir_all(fx.root.join(".git/info")).unwrap();
    fs::write(fx.root.join(".git/info/exclude"), "node_modules/\n").unwrap();
    let w = add(&fx, ".moai-cli/worktrees/installed", "b-installed", false);
    fs::create_dir_all(w.join("node_modules/pkg")).unwrap();
    fs::write(w.join("node_modules/pkg/index.js"), "x\n").unwrap();
    let cleanup = plan_at(&fx, later);
    assert!(
        row(&cleanup["candidates"], "worktrees/installed").is_none(),
        "{cleanup}"
    );
    let kept = row(&cleanup["kept"], "worktrees/installed").expect("kept");
    assert!(
        reasons(kept).contains(&"install_folder".to_string()),
        "{kept}"
    );
}
