//! SPEC-20261120-worktree-cap-and-cleanup, the preview half: `optimize-plan` says how many
//! worktrees the project has against a cap (default 40) and lists the finished ones as cleanup
//! candidates with the reason and the allocated bytes, and the ones it keeps with every reason.
//!
//! Every command runs against a temporary repository and a temporary HOME; the real `~/.moai`
//! and the real worktrees are never read or written. Nothing is deleted here (the apply half is
//! `tests/optimize_apply_cli.rs`).
#![cfg(unix)]
mod common;

use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Child, Command, Output, Stdio};

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

/// A repository on `main` with the integration branch `integ` at the same commit, a profile
/// naming the project `demo`, and a ledger for it.
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

/// A linked worktree at `rel` (under the repository, or absolute) on a new branch cut from
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

/// Cargo-shaped output so a worktree holds bytes to measure.
fn cargo_output(wt: &Path) {
    let t = wt.join("target");
    fs::create_dir_all(t.join("debug/deps")).unwrap();
    fs::write(
        t.join("CACHEDIR.TAG"),
        "Signature: 8a477f597d28d172789f06886806bc55\n",
    )
    .unwrap();
    fs::write(t.join("debug/.cargo-lock"), "").unwrap();
    fs::write(t.join("debug/deps/lib.rlib"), vec![7u8; 65_536]).unwrap();
}

/// One card of the ledger reserving `wt` on `branch`: `state` is the card's, `status` and
/// `lease_until` the reservation's.
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

fn now() -> u64 {
    std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap()
        .as_secs()
        + 30 * 86_400
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

fn preview_raw(fx: &Fx, extra: &[&str]) -> Value {
    let now = now().to_string();
    let mut args = vec![
        "optimize-plan",
        fx.root.to_str().unwrap(),
        "--integration-ref",
        "integ",
        "--now-unix",
        &now,
    ];
    args.extend_from_slice(extra);
    envelope(&run(fx, &args))
}

fn preview(fx: &Fx, extra: &[&str]) -> Value {
    let v = preview_raw(fx, extra);
    assert_eq!(v["ok"], true, "{v}");
    v["result"].clone()
}

/// The row of `list` whose `path` ends with `tail`.
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

#[test]
fn the_preview_counts_worktrees_against_the_cap() {
    let fx = fx();
    for n in 0..3 {
        add(
            &fx,
            &format!(".moai-cli/worktrees/w{n}"),
            &format!("b{n}"),
            true,
        );
    }
    // 3 linked worktrees and the primary checkout
    let over = preview(&fx, &["--worktree-cap", "2"]);
    let cap = &over["worktree_cap"];
    assert_eq!(cap["count"], 4, "{cap}");
    assert_eq!(cap["cap"], 2, "{cap}");
    assert_eq!(cap["source"], "flag", "{cap}");
    assert_eq!(cap["over"], true, "{cap}");
    assert_eq!(cap["over_by"], 2, "{cap}");
    // at the cap is not over it
    let at = preview(&fx, &["--worktree-cap", "4"]);
    assert_eq!(at["worktree_cap"]["over"], false, "{}", at["worktree_cap"]);
    assert_eq!(at["worktree_cap"]["over_by"], 0);
    // no cap given: the default of 40
    let dflt = preview(&fx, &[]);
    let cap = &dflt["worktree_cap"];
    assert_eq!(cap["cap"], 40, "{cap}");
    assert_eq!(cap["source"], "default", "{cap}");
    assert_eq!(cap["over"], false, "{cap}");
}

#[test]
fn a_cap_that_is_not_a_positive_whole_number_is_refused() {
    let fx = fx();
    for bad in ["0", "-1", "abc", "1.5", "", "99999999999999999999999"] {
        let arg = format!("--worktree-cap={bad}");
        let v = preview_raw(&fx, &[&arg]);
        assert_eq!(v["ok"], false, "{bad:?}: {v}");
        let msg = v["error"]["message"].as_str().unwrap_or_default();
        assert!(msg.contains("worktree-cap"), "{bad:?}: {msg}");
    }
}

#[test]
fn finished_worktrees_are_candidates_with_reason_and_bytes() {
    let fx = fx();
    let merged = add(&fx, ".moai-cli/worktrees/merged", "b-merged", false);
    cargo_output(&merged);
    let done = add(&fx, ".moai-cli/worktrees/done", "b-done", true);
    seed_card(&fx, &done, "b-done", "done", "released", 1);
    let open = add(&fx, ".moai-cli/worktrees/open", "b-open", true);
    let res = preview(&fx, &[]);
    let list = &res["cleanup"]["candidates"];

    let m = row(list, "worktrees/merged").unwrap_or_else(|| panic!("merged missing: {list}"));
    assert_eq!(m["remove_because"], "merged", "{m}");
    assert!(m["allocated_bytes"].as_u64().unwrap() >= 65_536, "{m}");
    assert_eq!(m["location"], "project", "{m}");
    assert!(
        m["item_digest"].as_str().is_some_and(|d| d.len() >= 16),
        "{m}"
    );
    assert_eq!(m["branch"], "b-merged");

    let d = row(list, "worktrees/done").unwrap_or_else(|| panic!("done missing: {list}"));
    assert_eq!(d["remove_because"], "card_done", "{d}");

    // neither merged nor with a finished card: not a candidate, and kept with the reason
    assert!(row(list, "worktrees/open").is_none(), "{list}");
    let kept = row(&res["cleanup"]["kept"], "worktrees/open").expect("open is listed as kept");
    assert!(reasons(kept).contains(&"not_merged".to_string()), "{kept}");
    let _ = open;
    // the totals add the candidates' bytes and say what they are
    assert!(
        res["cleanup"]["reclaimable_allocated_bytes"]
            .as_u64()
            .unwrap()
            >= 65_536
    );
    assert!(res["cleanup"]["space_basis"]
        .as_str()
        .is_some_and(|s| s.contains("allocated")));
}

fn spawn_in(dir: &Path) -> Child {
    Command::new("sleep")
        .arg("60")
        .current_dir(dir)
        .stdout(Stdio::null())
        .stderr(Stdio::null())
        .spawn()
        .unwrap()
}

#[test]
fn an_unsafe_worktree_is_kept_with_every_reason() {
    let fx = fx();
    let tracked = add(&fx, ".moai-cli/worktrees/tracked", "b-tracked", false);
    fs::write(tracked.join("a.txt"), "changed\n").unwrap();
    let untracked = add(&fx, ".moai-cli/worktrees/untracked", "b-untracked", false);
    fs::write(untracked.join("new.txt"), "n\n").unwrap();
    let ignored = add(&fx, ".moai-cli/worktrees/ignored", "b-ignored", false);
    fs::write(ignored.join("secrets.env"), "K=V\n").unwrap();
    let locked = add(&fx, ".moai-cli/worktrees/locked", "b-locked", false);
    git(&fx.root, &["worktree", "lock", locked.to_str().unwrap()]);
    let leased = add(&fx, ".moai-cli/worktrees/leased", "b-leased", false);
    seed_card(&fx, &leased, "b-leased", "active", "held", now() + 3_600);
    let busy = add(&fx, ".moai-cli/worktrees/busy", "b-busy", false);
    let mut child = spawn_in(&busy);
    // one worktree with two problems says both
    let both = add(&fx, ".moai-cli/worktrees/both", "b-both", false);
    fs::write(both.join("a.txt"), "changed\n").unwrap();
    fs::write(both.join("new.txt"), "n\n").unwrap();

    let res = preview(&fx, &[]);
    let _ = child.kill();
    let _ = child.wait();
    let cleanup = &res["cleanup"];
    let want = [
        ("worktrees/tracked", "tracked_changes"),
        ("worktrees/untracked", "untracked_files"),
        ("worktrees/ignored", "ignored_files_outside_build_output"),
        ("worktrees/locked", "locked"),
        ("worktrees/leased", "live_claim"),
        ("worktrees/busy", "process_in_use"),
    ];
    for (tail, why) in want {
        assert!(
            row(&cleanup["candidates"], tail).is_none(),
            "{tail} must not be a candidate: {cleanup}"
        );
        let kept = row(&cleanup["kept"], tail).unwrap_or_else(|| panic!("{tail} not kept"));
        assert!(reasons(kept).contains(&why.to_string()), "{tail}: {kept}");
    }
    let both = row(&cleanup["kept"], "worktrees/both").expect("both kept");
    let r = reasons(both);
    assert!(r.contains(&"tracked_changes".to_string()), "{both}");
    assert!(r.contains(&"untracked_files".to_string()), "{both}");

    // the apply half refuses one of them even when asked by name
    let now = now().to_string();
    let out = run(
        &fx,
        &[
            "optimize-apply",
            fx.root.to_str().unwrap(),
            "--integration-ref",
            "integ",
            "--now-unix",
            &now,
            "--permission-reference",
            "test:cap",
            "--pick",
            "x@deadbeef",
        ],
    );
    let v = envelope(&out);
    assert_eq!(v["ok"], true, "{v}");
    assert_eq!(v["result"]["removed"], json!([]), "{v}");
    assert!(tracked.is_dir(), "a dirty worktree must stay");
}

#[test]
fn old_locations_and_nested_worktrees_are_listed_and_preserved_ones_stay() {
    let fx = fx();
    let old = add(&fx, ".claude/worktrees/old", "b-old", false);
    let parent = add(&fx, ".moai-cli/worktrees/parent", "b-parent", false);
    cargo_output(&parent);
    let inner = add(
        &fx,
        &format!("{}/target/inner", parent.display()),
        "b-inner",
        false,
    );
    // preserved by the user's earlier decisions: the Go generation's folder and Codex
    let go = add(
        &fx,
        &format!("{}/.moai/worktrees/projects/demo/go1", fx.home.display()),
        "b-go",
        false,
    );
    let codex = add(
        &fx,
        &format!("{}/.codex/worktrees/abcd/repo", fx.home.display()),
        "b-codex",
        false,
    );
    let res = preview(&fx, &[]);
    let cleanup = &res["cleanup"];

    let o = row(&cleanup["candidates"], ".claude/worktrees/old")
        .unwrap_or_else(|| panic!("old location missing: {cleanup}"));
    assert_eq!(o["location"], "claude", "{o}");
    assert_eq!(o["remove_because"], "merged", "{o}");
    let _ = old;

    let i = row(&cleanup["candidates"], "target/inner")
        .unwrap_or_else(|| panic!("nested missing: {cleanup}"));
    assert_eq!(
        i["nested_in"]
            .as_str()
            .map(|p| p.ends_with("worktrees/parent")),
        Some(true),
        "{i}"
    );
    let p = row(&cleanup["candidates"], "worktrees/parent")
        .unwrap_or_else(|| panic!("parent missing: {cleanup}"));
    assert_eq!(
        p["nested_children"].as_array().map(Vec::len),
        Some(1),
        "{p}"
    );
    let _ = inner;

    for (tail, dir) in [("go1", &go), ("abcd/repo", &codex)] {
        assert!(
            row(&cleanup["candidates"], tail).is_none(),
            "{tail}: {cleanup}"
        );
        let kept = row(&cleanup["kept"], tail).unwrap_or_else(|| panic!("{tail} not listed"));
        assert!(
            reasons(kept).contains(&"preserved_location".to_string()),
            "{kept}"
        );
        assert!(dir.is_dir());
    }
}
