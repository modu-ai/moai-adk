//! SPEC-20261121-shared-build-root: worktrees build under the shared build root and a
//! worktree holding its own `target/` is reported. The plan names the build folder, the
//! apply creates it, and the gc plan counts in-worktree targets. Every test builds its own
//! git repository in a temporary directory and runs the real binary; the developer's
//! repository, its common dir and the real HOME are never touched.

#![cfg(unix)]

mod common;
use serde_json::{json, Value};
use std::{
    fs,
    os::unix::fs::PermissionsExt,
    path::{Path, PathBuf},
    process::{Command, Output},
};

const BIN: &str = crate::common::MOAI_BIN;
const ISSUE: i64 = 7;
const KEY_A: &str = "aaaaaaaaaaaaaaaa";
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

/// A primary repository with one commit, in a temporary folder with a HOME beside it.
struct Fx {
    _tmp: tempfile::TempDir,
    repo: PathBuf,
    home: PathBuf,
    source: Vec<String>,
}

fn fixture() -> Fx {
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("primary");
    let home = tmp.path().join("home");
    fs::create_dir_all(&repo).unwrap();
    fs::create_dir_all(&home).unwrap();
    git(&repo, &["init", "-q", "-b", "main"]);
    fs::write(repo.join("README.md"), "x\n").unwrap();
    fs::write(repo.join(".gitignore"), "/.moai-cli/worktrees/\n").unwrap();
    git(&repo, &["add", "README.md", ".gitignore"]);
    git(&repo, &["commit", "-q", "-m", "init"]);
    let rec = record(&git(&repo, &["rev-parse", "HEAD"]));
    let source = record_args(tmp.path(), &rec);
    Fx {
        _tmp: tmp,
        repo,
        home,
        source,
    }
}

fn record(base: &str) -> Value {
    json!({
        "schema_version": 1,
        "task_ref": {
            "provider": "github", "host": "github.com", "owner": "acme", "repo": "widgets",
            "repository_id": 101, "issue_id": 7001, "issue_number": ISSUE
        },
        "task_key": KEY_A,
        "logical_key": "06-C",
        "owner": {"kind": "session", "id": "agent-1-0dd806"},
        "epoch": 1,
        "base_revision": base,
        "worktree": {"common_dir": "/x/.git", "git_dir": "/x/.git", "path": "/x", "head": base},
        "write_scope": ["src/worktree.rs"],
        "spec_ref": {"id": "SPEC-RC1-06-C", "revision": "proposed", "digest": "sha256:abc"},
        "comment_id": COMMENT_ID,
        "claim_token": format!("Moai-Claim: {ISSUE}#{COMMENT_ID}"),
        "lease_until": "2026-11-08T15:25:00Z",
        "state": "running",
        "effects": [],
        "last_observed_at": "2026-11-08T12:00:00Z"
    })
}

fn record_args(dir: &Path, rec: &Value) -> Vec<String> {
    let r = dir.join("record.json");
    let c = dir.join("comment.json");
    fs::write(&r, rec.to_string()).unwrap();
    fs::write(
        &c,
        json!({
            "id": COMMENT_ID,
            "body": "moai-claim: session=agent-1-0dd806 epoch=1 lease_until=2026-11-08T15:25:00Z"
        })
        .to_string(),
    )
    .unwrap();
    vec![
        "--record".into(),
        r.to_string_lossy().into(),
        "--claim-comment".into(),
        c.to_string_lossy().into(),
    ]
}

fn run(cwd: &Path, args: &[String]) -> Output {
    Command::new(BIN)
        .args(args)
        .current_dir(cwd)
        .env("HOME", cwd.parent().unwrap().join("home"))
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("GIT_AUTHOR_NAME", "test")
        .env("GIT_AUTHOR_EMAIL", "test@example.invalid")
        .env("GIT_COMMITTER_NAME", "test")
        .env("GIT_COMMITTER_EMAIL", "test@example.invalid")
        .output()
        .unwrap()
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON: {e}\n{}\n{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

fn plan(f: &Fx) -> Value {
    let out = run(
        &f.repo,
        &[
            "worktree-plan".to_string(),
            f.repo.to_string_lossy().into(),
            "--branch-prefix".into(),
            "lane".into(),
        ]
        .into_iter()
        .chain(f.source.iter().cloned())
        .collect::<Vec<_>>(),
    );
    let env = envelope(&out);
    assert_eq!(env["ok"], true, "{env}");
    env["result"].clone()
}

fn apply(f: &Fx, digest: &str) -> Value {
    let out = run(
        &f.repo,
        &[
            "worktree-apply".to_string(),
            f.repo.to_string_lossy().into(),
            "--branch-prefix".into(),
            "lane".into(),
            "--plan-digest".into(),
            digest.into(),
            "--disk-floor-bytes".into(),
            "0".into(),
        ]
        .into_iter()
        .chain(f.source.iter().cloned())
        .collect::<Vec<_>>(),
    );
    envelope(&out)
}

/// REQ-1: the plan names the worktree's build folder under the shared build root, and
/// planning creates nothing.
#[test]
fn worktree_plan_names_the_shared_build_folder() {
    let f = fixture();
    let r = plan(&f);
    let path = PathBuf::from(r["path"].as_str().unwrap());
    let name = path.file_name().unwrap().to_string_lossy().into_owned();
    assert_eq!(r["writes"], json!(false));
    assert_eq!(
        r["build_folder"],
        json!(f
            .home
            .join(".moai/build")
            .join(&name)
            .to_string_lossy()
            .as_ref()),
        "the plan names <HOME>/.moai/build/<worktree folder name>"
    );
    assert!(!path.exists(), "planning created nothing");
    assert!(
        !f.home.join(".moai/build").exists(),
        "planning created no build folder either"
    );
}

/// REQ-2: the apply creates the build folder with mode 0700.
#[test]
fn worktree_apply_creates_the_build_folder_0700() {
    let f = fixture();
    let p = plan(&f);
    let digest = p["plan_digest"].as_str().unwrap().to_string();
    let env = apply(&f, &digest);
    assert_eq!(env["result"]["decision"], json!("created"), "{env}");
    let folder = PathBuf::from(env["result"]["build_folder"].as_str().unwrap());
    let meta = fs::metadata(&folder).unwrap();
    assert!(meta.is_dir(), "{} is a directory", folder.display());
    #[cfg(unix)]
    assert_eq!(
        meta.permissions().mode() & 0o777,
        0o700,
        "the build folder is private to the owner"
    );
}

/// REQ-2: a link at the build folder's path is never followed; the apply stays created and
/// reports the refusal in the answer.
#[test]
fn worktree_apply_reports_a_link_at_the_build_folder() {
    let f = fixture();
    let elsewhere = f.home.parent().unwrap().join("elsewhere");
    fs::create_dir_all(f.home.join(".moai/build")).unwrap();
    fs::create_dir_all(&elsewhere).unwrap();
    #[cfg(unix)]
    std::os::unix::fs::symlink(&elsewhere, f.home.join(".moai/build/primary-7-06-C")).unwrap();

    let p = plan(&f);
    let digest = p["plan_digest"].as_str().unwrap().to_string();
    let env = apply(&f, &digest);
    assert_eq!(env["result"]["decision"], json!("created"), "{env}");
    let err = env["result"]["build_folder_error"]
        .as_str()
        .unwrap_or_else(|| panic!("the link is reported: {env}"));
    assert!(err.contains("is a link; it is not followed"), "{err}");
    assert!(
        fs::read_dir(&elsewhere).unwrap().count() == 0,
        "the link's target stayed empty"
    );
}

/// REQ-3 + REQ-4: the gc plan counts examined worktrees holding their own `target/`, with
/// example paths, and reports zero with an empty list when none holds one. Ignored until the
/// worktree_gc.rs slice lands: the leader sequenced C-275 (p0) onto that file first, and this
/// lane's summary patch waits as a saved patch outside the repository.
#[test]
#[ignore = "REQ-3/REQ-4 wait for the worktree_gc.rs slice (C-275 holds the file first per the leader's sequencing); the summary patch is saved under ~/.moai/reports/moai-adk/"]
fn gc_plan_counts_worktrees_holding_target() {
    let dir = tempfile::tempdir().unwrap();
    let root = dir.path().join("repo");
    let home = dir.path().join("home");
    fs::create_dir_all(&root).unwrap();
    fs::create_dir_all(&home).unwrap();
    git(&root, &["init", "-q", "-b", "main"]);
    fs::write(root.join(".gitignore"), ".moai-cli/\ntarget/\n").unwrap();
    fs::write(root.join("a.txt"), "a\n").unwrap();
    git(&root, &["add", ".gitignore", "a.txt"]);
    git(&root, &["commit", "-q", "-m", "init"]);
    git(&root, &["branch", "integ"]);
    for n in ["one", "two"] {
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
        fs::write(wt.join("target/debug/blob"), "x\n").unwrap();
    }

    let now = (std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap()
        .as_secs()
        + 86_400)
        .to_string();
    let out = Command::new(BIN)
        .args([
            "worktree-gc-plan",
            root.to_str().unwrap(),
            "--integration-ref",
            "integ",
            "--now-unix",
            &now,
        ])
        .current_dir(&root)
        .env("HOME", &home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .output()
        .unwrap();
    let env = envelope(&out);
    assert_eq!(env["ok"], true, "{}", String::from_utf8_lossy(&out.stderr));
    let r = &env["result"];
    let summary = &r["in_worktree_targets"];
    assert_eq!(summary["worktrees"], json!(2), "{r}");
    assert_eq!(
        summary["examples"].as_array().unwrap().len(),
        2,
        "both worktrees are named"
    );
    assert!(
        summary["trim_bytes"].as_u64().unwrap_or(0) > 0,
        "the measured bytes ride along: {summary}"
    );
    assert_eq!(summary["bytes_partial"], json!(false), "{summary}");

    // REQ-4: a repository without any in-worktree target reports the zero case.
    let clean = dir.path().join("clean");
    let clean_home = dir.path().join("clean-home");
    fs::create_dir_all(&clean).unwrap();
    fs::create_dir_all(&clean_home).unwrap();
    git(&clean, &["init", "-q", "-b", "main"]);
    fs::write(clean.join(".gitignore"), ".moai-cli/\ntarget/\n").unwrap();
    git(&clean, &["add", ".gitignore"]);
    git(&clean, &["commit", "-q", "-m", "init"]);
    git(&clean, &["branch", "integ"]);
    let out = Command::new(BIN)
        .args([
            "worktree-gc-plan",
            clean.to_str().unwrap(),
            "--integration-ref",
            "integ",
            "--now-unix",
            &now,
        ])
        .current_dir(&clean)
        .env("HOME", &clean_home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .output()
        .unwrap();
    let env = envelope(&out);
    assert_eq!(env["ok"], true, "{}", String::from_utf8_lossy(&out.stderr));
    let summary = &env["result"]["in_worktree_targets"];
    assert_eq!(summary["worktrees"], json!(0), "{}", env["result"]);
    assert_eq!(
        summary["examples"].as_array().unwrap().len(),
        0,
        "no examples without a target"
    );
}
