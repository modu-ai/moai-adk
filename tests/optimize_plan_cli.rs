//! #2724: `moai-cli optimize-plan` on a synthetic repository, a temporary HOME and a temporary
//! build root — a scaled-down copy of the leader's measurement (#2677 comment 6850581277:
//! 185 build folders under /tmp, 152 merged worktrees still holding `target/`). Nothing real
//! is deleted; the only removal in this file is `worktree-gc-apply` inside the fixture.
#![cfg(unix)]
mod common;

use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output};

const BIN: &str = crate::common::MOAI_BIN;
/// Build folders in the scaled-down measurement.
const OLD_BUILD_FOLDERS: usize = 12;

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
    root: PathBuf,
    home: PathBuf,
    builds: PathBuf,
}

fn cargo_output(dir: &Path) {
    fs::create_dir_all(dir.join("debug/deps")).unwrap();
    fs::write(
        dir.join("CACHEDIR.TAG"),
        "Signature: 8a477f597d28d172789f06886806bc55\n",
    )
    .unwrap();
    fs::write(dir.join("debug/.cargo-lock"), "").unwrap();
    fs::write(dir.join("debug/deps/lib.rlib"), vec![7u8; 32_768]).unwrap();
}

fn wt(fx: &Fixture, n: &str) -> PathBuf {
    fx.root.join(".moai-cli/worktrees").join(n)
}

/// Worktrees (branch `issue/<n>-<name>`), integration ref `integ`:
/// - `a` (10) ancestor-merged, clean, cargo output — audit PASS for its head
/// - `s` (11) squash-merged: its one change sits on `integ` as a different commit — PASS
/// - `p` (12) partly on `integ`: one of its two changed files differs — not merged
/// - `n` (13) ancestor-merged, cargo output, no audit verdict
/// - `c` (14) ancestor-merged, cargo output, live claim on #14
/// - `h` (15) ancestor-merged, cargo output, #15 on HOLD
///
/// Build root: `OLD_BUILD_FOLDERS` cargo folders, one folder of notes, one cargo folder with
/// a note beside it. Reports: one file under `.moai-cli/reports`.
fn fixture() -> Fixture {
    let dir = tempfile::TempDir::new().unwrap();
    let base = fs::canonicalize(dir.path()).unwrap();
    let (root, home, builds) = (base.join("repo"), base.join("home"), base.join("builds"));
    for d in [&root, &home, &builds] {
        fs::create_dir_all(d).unwrap();
    }
    git(&root, &["init", "-q", "-b", "main"]);
    fs::write(root.join(".gitignore"), ".moai-cli/\ntarget/\n").unwrap();
    fs::write(root.join("a.txt"), "a\n").unwrap();
    git(&root, &["add", ".gitignore", "a.txt"]);
    git(&root, &["commit", "-q", "-m", "init"]);
    git(&root, &["branch", "integ"]);
    let fx = Fixture {
        _dir: dir,
        root,
        home,
        builds,
    };
    for (n, name) in [
        (10, "a"),
        (11, "s"),
        (12, "p"),
        (13, "n"),
        (14, "c"),
        (15, "h"),
    ] {
        let path = wt(&fx, name);
        git(
            &fx.root,
            &[
                "worktree",
                "add",
                "-q",
                "-b",
                &format!("issue/{n}-{name}"),
                path.to_str().unwrap(),
                "integ",
            ],
        );
    }
    for n in ["a", "s", "n", "c", "h"] {
        cargo_output(&wt(&fx, n).join("target"));
    }
    // s: one change, then the same content lands on integ as a squash commit
    fs::write(wt(&fx, "s").join("s.txt"), "squashed\n").unwrap();
    git(&wt(&fx, "s"), &["add", "s.txt"]);
    git(&wt(&fx, "s"), &["commit", "-q", "-m", "s part 1"]);
    fs::write(wt(&fx, "s").join("a.txt"), "a\nfrom s\n").unwrap();
    git(&wt(&fx, "s"), &["commit", "-q", "-am", "s part 2"]);
    // p: two changes, only one lands on integ
    fs::write(wt(&fx, "p").join("p1.txt"), "one\n").unwrap();
    fs::write(wt(&fx, "p").join("p2.txt"), "two\n").unwrap();
    git(&wt(&fx, "p"), &["add", "p1.txt", "p2.txt"]);
    git(&wt(&fx, "p"), &["commit", "-q", "-m", "p"]);
    // the squash commits on integ, written from a scratch worktree of integ
    let scratch = fx.root.join(".moai-cli/scratch");
    git(
        &fx.root,
        &["worktree", "add", "-q", scratch.to_str().unwrap(), "integ"],
    );
    fs::write(scratch.join("s.txt"), "squashed\n").unwrap();
    fs::write(scratch.join("a.txt"), "a\nfrom s\n").unwrap();
    fs::write(scratch.join("p1.txt"), "one\n").unwrap();
    git(&scratch, &["add", "s.txt", "a.txt", "p1.txt"]);
    git(
        &scratch,
        &["commit", "-q", "-m", "squash of s and part of p"],
    );
    git(
        &fx.root,
        &["worktree", "remove", "--force", scratch.to_str().unwrap()],
    );
    // n follows integ to its new tip: still an ancestor, but a head of its own (a verdict
    // for `a`'s head must not cover it)
    git(&wt(&fx, "n"), &["merge", "-q", "--ff-only", "integ"]);

    for i in 0..OLD_BUILD_FOLDERS {
        cargo_output(&fx.builds.join(format!("moai-rc-target-agent-{i}")));
    }
    fs::create_dir_all(fx.builds.join("notes-only")).unwrap();
    fs::write(fx.builds.join("notes-only/n.md"), "note\n").unwrap();
    cargo_output(&fx.builds.join("mixed"));
    fs::write(fx.builds.join("mixed/pr-body.md"), "lane note\n").unwrap();

    fs::create_dir_all(fx.root.join(".moai-cli/reports")).unwrap();
    fs::write(fx.root.join(".moai-cli/reports/r.json"), "{}\n").unwrap();
    fx
}

fn head(fx: &Fixture, n: &str) -> String {
    git(&wt(fx, n), &["rev-parse", "HEAD"]).trim().to_string()
}

fn now_plus_days(days: u64) -> u64 {
    std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap()
        .as_secs()
        + days * 86_400
}

fn run(fx: &Fixture, args: &[&str]) -> Output {
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

/// Full observations: audit PASS for `a` and `s`, a live claim on #14, #15 on HOLD.
fn observations(fx: &Fixture) -> Value {
    json!({
        "schema_version": 1,
        "audits": {"source": "fixture", "items": [
            {"pr": 100, "head": head(fx, "a"), "verdict_comment": 1, "verdict": "PASS"},
            {"pr": 101, "head": head(fx, "s"), "verdict_comment": 2, "verdict": "PASS"},
            {"pr": 103, "head": head(fx, "n"), "verdict_comment": 3, "verdict": "FAIL"},
        ]},
        "claims": {"source": "fixture", "items": [{"issue": 14, "claim_comment": 4}]},
        "holds": {"source": "fixture", "items": [{"issue": 15}]},
    })
}

fn optimize(fx: &Fixture, obs: Option<&Value>, now: u64, extra: &[&str]) -> Value {
    let obs_path = fx.home.join("obs.json");
    let mut args = vec![
        "optimize-plan".to_string(),
        fx.root.to_str().unwrap().into(),
        "--integration-ref".into(),
        "integ".into(),
        "--now-unix".into(),
        now.to_string(),
        "--build-root".into(),
        fx.builds.to_str().unwrap().into(),
    ];
    if let Some(o) = obs {
        fs::write(&obs_path, serde_json::to_vec(o).unwrap()).unwrap();
        args.extend(["--observations".into(), obs_path.to_str().unwrap().into()]);
    }
    args.extend(extra.iter().map(|s| s.to_string()));
    let refs: Vec<&str> = args.iter().map(String::as_str).collect();
    let v = envelope(&run(fx, &refs));
    assert_eq!(v["ok"], true, "{v}");
    assert_eq!(v["operation"], "optimize.plan");
    v["result"].clone()
}

fn row<'a>(r: &'a Value, kind: &str, path_end: &str) -> &'a Value {
    r["rows"]
        .as_array()
        .unwrap()
        .iter()
        .find(|x| x["kind"] == kind && x["path"].as_str().unwrap().ends_with(path_end))
        .unwrap_or_else(|| panic!("no {kind} {path_end}: {r}"))
}

fn reasons(x: &Value) -> Vec<String> {
    x["reasons"]
        .as_array()
        .unwrap()
        .iter()
        .map(|v| v.as_str().unwrap().to_string())
        .collect()
}

fn gc_plan(fx: &Fixture, now: u64) -> Value {
    let v = envelope(&run(
        fx,
        &[
            "worktree-gc-plan",
            fx.root.to_str().unwrap(),
            "--integration-ref",
            "integ",
            "--now-unix",
            &now.to_string(),
        ],
    ));
    assert_eq!(v["ok"], true, "{v}");
    v["result"].clone()
}

fn gc_entry<'a>(r: &'a Value, name: &str) -> &'a Value {
    let path = format!(".moai-cli/worktrees/{name}");
    r["candidates"]
        .as_array()
        .unwrap()
        .iter()
        .find(|c| c["path"] == path.as_str())
        .unwrap_or_else(|| panic!("no {path}: {r}"))
}

/// AC: the squash-merge test. `s` is merged by squash (its changed paths match at the tip);
/// `p` is not (one of its two paths differs). `worktree-gc-plan`, which apply recomputes,
/// still counts only ancestors as merged.
#[test]
fn a_squash_merge_is_recognised_only_when_every_changed_path_matches() {
    let fx = fixture();
    let now = now_plus_days(8);
    let r = optimize(&fx, Some(&observations(&fx)), now, &[]);
    let s = row(&r, "worktree", "worktrees/s");
    assert_eq!(s["merged_by"], "squash", "{s}");
    assert_eq!(s["action"], "remove", "{s}");
    assert_eq!(s["audit_evidence"]["pr"], 101);
    let p = row(&r, "worktree", "worktrees/p");
    assert_eq!(p["merged_by"], Value::Null);
    assert_eq!(p["action"], "skip");
    assert!(reasons(p).contains(&"not_merged".to_string()), "{p}");
    assert_eq!(row(&r, "worktree", "worktrees/a")["merged_by"], "ancestor");

    let gc = gc_plan(&fx, now);
    assert_eq!(gc_entry(&gc, "s")["action"], "skip");
    assert_eq!(reasons(gc_entry(&gc, "s")), ["not_merged"]);
    assert_eq!(gc_entry(&gc, "s")["merged_by"], Value::Null);
    assert_eq!(gc_entry(&gc, "a")["merged_by"], "ancestor");
}

/// Remove needs an audit PASS for the exact head; a FAIL or no verdict keeps the worktree
/// (only its regenerable cargo output is offered); a live claim or a HOLD keeps it whole.
#[test]
fn a_worktree_is_a_remove_candidate_only_with_audit_evidence() {
    let fx = fixture();
    let r = optimize(&fx, Some(&observations(&fx)), now_plus_days(8), &[]);
    let a = row(&r, "worktree", "worktrees/a");
    assert_eq!(a["action"], "remove", "{a}");
    assert!(a["restore"].as_str().unwrap().contains("git worktree add"));
    let n = row(&r, "worktree", "worktrees/n");
    assert_eq!(n["action"], "trim", "{n}");
    assert_eq!(reasons(n), ["audit_unconfirmed"]);
    let c = row(&r, "worktree", "worktrees/c");
    assert_eq!(
        (c["action"].as_str(), reasons(c)),
        (Some("skip"), vec!["live_claim".to_string()])
    );
    let h = row(&r, "worktree", "worktrees/h");
    assert_eq!(
        (h["action"].as_str(), reasons(h)),
        (Some("skip"), vec!["hold".to_string()])
    );
    assert_eq!(r["totals"]["remove"], 2, "{r}");
}

/// Without an observations file nothing outside the build root is a candidate: every
/// worktree that gc would act on is kept with the unobserved parts named.
#[test]
fn without_observations_no_worktree_is_a_candidate() {
    let fx = fixture();
    let r = optimize(&fx, None, now_plus_days(8), &[]);
    for x in r["rows"].as_array().unwrap() {
        if x["kind"] == "worktree" {
            assert_eq!(x["action"], "skip", "{x}");
        }
    }
    let a = row(&r, "worktree", "worktrees/a");
    assert_eq!(reasons(a), ["claims_unobserved", "holds_unobserved"]);
    assert_eq!(r["observations"]["claims"], Value::Null);
}

/// Build folders: only cargo output past retention, nothing else beside it.
#[test]
fn build_folders_count_only_past_retention_and_only_cargo_output() {
    let fx = fixture();
    let obs = observations(&fx);
    let r = optimize(&fx, Some(&obs), now_plus_days(8), &[]);
    let trims = r["rows"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|x| x["kind"] == "build_folder" && x["action"] == "trim")
        .count();
    assert_eq!(trims, OLD_BUILD_FOLDERS, "{r}");
    assert_eq!(
        reasons(row(&r, "build_folder", "/notes-only")),
        ["not_cargo_output", "other_files_in_folder"]
    );
    assert_eq!(
        reasons(row(&r, "build_folder", "/mixed")),
        ["other_files_in_folder"]
    );
    let one = row(&r, "build_folder", "/moai-rc-target-agent-0");
    assert!(one["bytes"].as_u64().unwrap() >= 32_768, "{one}");
    assert!(one["retention_until_unix"].as_u64().is_some());

    // a day later instead of eight: every build folder is within the 7-day retention
    let young = optimize(&fx, Some(&obs), now_plus_days(1), &[]);
    assert_eq!(
        reasons(row(&young, "build_folder", "/moai-rc-target-agent-0")),
        ["within_retention"]
    );
    let r3 = optimize(
        &fx,
        Some(&obs),
        now_plus_days(1),
        &["--retention-days", "1"],
    );
    assert_eq!(
        row(&r3, "build_folder", "/moai-rc-target-agent-0")["action"],
        "trim"
    );

    let rep = row(&r, "report", "reports");
    assert_eq!(
        (rep["action"].as_str(), reasons(rep)),
        (Some("keep"), vec!["reports_kept".to_string()])
    );
    // AUDIT #2832 B1: a project still on the earlier root has one reports folder, listed once
    // (the state folder and the earlier root are the same `.moai-cli` there).
    let report_paths: Vec<&str> = r["rows"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|x| x["kind"] == "report")
        .map(|x| x["path"].as_str().unwrap())
        .collect();
    assert_eq!(report_paths, [".moai-cli/reports"], "{r}");
}

/// #2696 C4d: a project on `.moai/` (marker present) lists the reports of both roots — the
/// ones a migration left in `.moai-cli/reports` (#2725 D5) and new ones in `.moai/reports` —
/// each kept, none trimmed.
#[test]
fn a_migrated_project_keeps_the_reports_of_both_roots() {
    let fx = fixture();
    fs::create_dir_all(fx.root.join(".moai/reports")).unwrap();
    fs::write(
        fx.root.join(".moai/moai-cli.json"),
        r#"{"schema_version":1,"generation":"moai-cli"}"#,
    )
    .unwrap();
    fs::write(fx.root.join(".moai/reports/n.json"), "{}\n").unwrap();
    let r = optimize(&fx, None, now_plus_days(8), &[]);
    let paths: Vec<&str> = r["rows"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|x| x["kind"] == "report")
        .map(|x| x["path"].as_str().unwrap())
        .collect();
    assert_eq!(paths, [".moai/reports", ".moai-cli/reports"], "{r}");
    for x in r["rows"].as_array().unwrap() {
        if x["kind"] == "report" {
            assert_eq!(x["action"], "keep", "{x}");
            assert_eq!(x["trim_paths"], json!([]), "{x}");
        }
    }
}

fn snapshot(dir: &Path) -> Vec<(String, u64, std::time::SystemTime)> {
    let mut out = Vec::new();
    let mut stack = vec![dir.to_path_buf()];
    while let Some(p) = stack.pop() {
        let m = fs::symlink_metadata(&p).unwrap();
        if m.is_dir() {
            out.push((format!("{}/", p.display()), 0, m.modified().unwrap()));
            for e in fs::read_dir(&p).unwrap() {
                stack.push(e.unwrap().path());
            }
        } else {
            out.push((p.display().to_string(), m.len(), m.modified().unwrap()));
        }
    }
    out.sort();
    out
}

/// AC: removal without approval is zero. The preview writes nothing in the repository or the
/// build root, its digest is stable, and it has candidates — so a preview that deleted them
/// would show here. `worktree-gc-apply` with a digest that is not its own plan's (the
/// preview's) refuses before any write; with its own plan's digest it removes only the
/// ancestor-merged clean worktree, never the squash-merged one.
#[test]
fn no_removal_happens_without_the_gc_plan_digest() {
    let fx = fixture();
    let now = now_plus_days(8);
    let obs = observations(&fx);
    let (repo_before, builds_before) = (snapshot(&fx.root), snapshot(&fx.builds));
    let one = optimize(&fx, Some(&obs), now, &[]);
    let two = optimize(&fx, Some(&obs), now, &[]);
    assert!(one["totals"]["remove"].as_u64().unwrap() >= 1, "{one}");
    assert!(one["totals"]["trim"].as_u64().unwrap() >= 1, "{one}");
    assert_eq!(one["preview_digest"], two["preview_digest"]);
    assert_eq!(one["writes"], false);
    assert_eq!(snapshot(&fx.root), repo_before);
    assert_eq!(snapshot(&fx.builds), builds_before);

    let preview = one["preview_digest"].as_str().unwrap();
    let apply = |digest: &str| {
        envelope(&run(
            &fx,
            &[
                "worktree-gc-apply",
                fx.root.to_str().unwrap(),
                "--integration-ref",
                "integ",
                "--now-unix",
                &now.to_string(),
                "--plan-digest",
                digest,
                "--permission-reference",
                "fixture-approved",
            ],
        ))
    };
    let refused = apply(preview);
    assert_eq!(refused["ok"], false, "{refused}");
    assert_eq!(snapshot(&fx.root), repo_before);

    let gc = gc_plan(&fx, now);
    let done = apply(gc["plan_digest"].as_str().unwrap());
    assert_eq!(done["ok"], true, "{done}");
    assert!(
        !wt(&fx, "a").exists(),
        "the ancestor-merged clean worktree stays"
    );
    assert!(
        wt(&fx, "s").join("s.txt").is_file(),
        "the squash-merged worktree was removed"
    );
    assert_eq!(snapshot(&fx.builds), builds_before);
}

#[test]
fn a_relative_or_symlinked_build_root_and_bad_observations_are_refused() {
    let fx = fixture();
    let link = fx.home.join("link");
    std::os::unix::fs::symlink(&fx.builds, &link).unwrap();
    let base = |extra: &[&str]| {
        let mut a = vec![
            "optimize-plan",
            fx.root.to_str().unwrap(),
            "--integration-ref",
            "integ",
        ];
        a.extend_from_slice(extra);
        let out = run(&fx, &a);
        assert!(!out.status.success());
        envelope(&out)
    };
    for extra in [
        vec!["--build-root", "builds"],
        vec!["--build-root", link.to_str().unwrap()],
    ] {
        let v = base(&extra);
        assert_eq!(v["error"]["kind"], "invalid_input", "{v}");
    }
    let bad = fx.home.join("bad.json");
    fs::write(&bad, r#"{"schema_version":1,"extra":true}"#).unwrap();
    let v = base(&["--observations", bad.to_str().unwrap()]);
    assert_eq!(v["error"]["kind"], "invalid_input", "{v}");
    let v = base(&["--retention-days", "0"]);
    assert_eq!(v["error"]["kind"], "invalid_input", "{v}");
}
