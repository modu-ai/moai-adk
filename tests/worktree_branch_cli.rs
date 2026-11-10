//! #2131: the worktree identity records the branch checked out at claim time
//! (`WorktreeId.branch`, the #2130 additive field). Each test builds its own repository in a
//! temporary directory; no network, no `gh`, and the developer's repository is never touched.
use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
    process::Command,
};

const COMMENT_ID: i64 = 6_721_450_649;

fn git(dir: &Path, args: &[&str]) -> String {
    let out = Command::new("git")
        .args(["-c", "gc.auto=0", "-c", "maintenance.auto=false"])
        .args([
            "-c",
            "user.email=test@example.invalid",
            "-c",
            "user.name=test",
        ])
        .args(args)
        .current_dir(dir)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "git {args:?}: {}",
        String::from_utf8_lossy(&out.stderr)
    );
    String::from_utf8(out.stdout).unwrap().trim().to_string()
}

/// A primary repository with one commit; returns its HEAD.
fn make_repo(dir: &Path) -> String {
    fs::create_dir_all(dir).unwrap();
    git(dir, &["init", "-q", "-b", "main"]);
    fs::write(dir.join("README.md"), "x\n").unwrap();
    fs::write(dir.join(".gitignore"), "/.moai-cli/worktrees/\n").unwrap();
    git(dir, &["add", "README.md", ".gitignore"]);
    git(dir, &["commit", "-q", "-m", "init"]);
    git(dir, &["rev-parse", "HEAD"])
}

/// A pre-#2131 claim record: its worktree carries no `branch`.
fn record(base: &str) -> Value {
    json!({
        "schema_version": 1,
        "task_ref": {
            "provider": "github", "host": "github.com", "owner": "acme", "repo": "widgets",
            "repository_id": 101, "issue_id": 7001, "issue_number": 7
        },
        "task_key": "aaaaaaaaaaaaaaaa",
        "logical_key": "06-C",
        "owner": {"kind": "session", "id": "agent-1-0dd806"},
        "epoch": 1,
        "base_revision": base,
        "worktree": {"common_dir": "/x/.git", "git_dir": "/x/.git", "path": "/x", "head": base},
        "write_scope": ["src/worktree.rs"],
        "spec_ref": {"id": "SPEC-RC1-06-C", "revision": "proposed", "digest": "sha256:abc"},
        "comment_id": COMMENT_ID,
        "claim_token": format!("Moai-Claim: 7#{COMMENT_ID}"),
        "lease_until": "2026-11-08T15:25:00Z",
        "state": "running",
        "effects": [],
        "last_observed_at": "2026-11-08T12:00:00Z"
    })
}

fn run(cwd: &Path, args: &[&str]) -> Value {
    let out = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .args(args)
        .current_dir(cwd)
        .env("HOME", cwd.join("home"))
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("GIT_AUTHOR_NAME", "test")
        .env("GIT_AUTHOR_EMAIL", "test@example.invalid")
        .env("GIT_COMMITTER_NAME", "test")
        .env("GIT_COMMITTER_EMAIL", "test@example.invalid")
        .output()
        .unwrap();
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON: {e}\n{}\n{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

#[test]
fn worktree_apply_records_the_checked_out_branch() {
    // AC-1: after worktree-apply the identity (and the record update) names the branch.
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("primary");
    let base = make_repo(&repo);
    let rec = tmp.path().join("record.json");
    let com = tmp.path().join("comment.json");
    fs::write(&rec, record(&base).to_string()).unwrap();
    let body = "moai-claim: session=agent-1-0dd806 epoch=1 lease_until=2026-11-08T15:25:00Z";
    fs::write(&com, json!({"id": COMMENT_ID, "body": body}).to_string()).unwrap();
    let p = repo.to_str().unwrap();
    let source = [
        "--record",
        rec.to_str().unwrap(),
        "--claim-comment",
        com.to_str().unwrap(),
    ];
    let mut plan_args = vec!["worktree-plan", p, "--branch-prefix", "lane"];
    plan_args.extend(source);
    let plan = run(&repo, &plan_args);
    assert_eq!(plan["ok"], true, "{plan}");
    let digest = plan["result"]["plan_digest"].as_str().unwrap().to_string();
    let mut apply_args = vec![
        "worktree-apply",
        p,
        "--branch-prefix",
        "lane",
        "--plan-digest",
        &digest,
        // The host's free space must not decide this test (WT-5 floor).
        "--disk-floor-bytes",
        "0",
    ];
    apply_args.extend(source);
    let env = run(&repo, &apply_args);
    assert_eq!(env["ok"], true, "{env}");
    let r = &env["result"];
    assert_eq!(r["decision"], "created");
    let path = PathBuf::from(r["path"].as_str().unwrap());
    let checked_out = git(&path, &["rev-parse", "--abbrev-ref", "HEAD"]);
    assert_eq!(checked_out, "lane/7-06-C-e1");
    assert_eq!(r["identity"]["branch"], checked_out);
    assert_eq!(
        r["record_update_request"]["worktree"]["branch"],
        checked_out
    );
}

#[test]
fn detached_head_records_no_branch() {
    // AC-1: a detached HEAD is not an error; the branch is None and is not serialized.
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("primary");
    let base = make_repo(&repo);
    git(&repo, &["checkout", "-q", "--detach"]);
    let id = moai_cli::worktree::identity(&repo).unwrap();
    assert_eq!(id.head, base);
    assert_eq!(id.branch, None);
    let v = serde_json::to_value(&id).unwrap();
    assert!(v.get("branch").is_none(), "{v}");

    // Back on a branch, the same call names it.
    git(&repo, &["checkout", "-q", "main"]);
    let id = moai_cli::worktree::identity(&repo).unwrap();
    assert_eq!(id.branch.as_deref(), Some("main"));
}
