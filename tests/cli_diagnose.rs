//! Product CLI tests of the Diagnose group (#2834 R1): one file per command group so a pull
//! request that adds or changes a command touches its own group file, not `tests/product_cli.rs`.
//! The identity tests below moved from `tests/product_cli.rs` unchanged.

mod common;

use common::cli::{create, profile, tree_state, write_file};
use common::git::git_in;
use serde_json::{json, Value};
use std::{fs, process::Command};
use tempfile::TempDir;

// ---- identity (Issue #1988, 01-A) ---------------------------------------------------------

/// Runs `moai-cli identity <path>` with a fixture HOME and no user git config. `path_env`
/// replaces PATH when given (used to make git unavailable).
fn ident(
    path: &std::path::Path,
    home: &std::path::Path,
    path_env: Option<&str>,
) -> (std::process::Output, Value) {
    ident_env(path, home, path_env, &[])
}
/// Like `ident`, with extra caller environment variables (Issue #2033).
fn ident_env(
    path: &std::path::Path,
    home: &std::path::Path,
    path_env: Option<&str>,
    envs: &[(&str, &str)],
) -> (std::process::Output, Value) {
    let mut cmd = Command::new(crate::common::MOAI_BIN);
    cmd.arg("identity")
        .arg(path)
        .env("HOME", home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_NOSYSTEM", "1")
        .env_remove("GIT_DIR")
        .env_remove("GIT_WORK_TREE")
        .envs(envs.iter().copied());
    if let Some(p) = path_env {
        cmd.env("PATH", p);
    }
    let out = cmd.output().unwrap();
    // The identity sits under the envelope's `result` (01-C #1990); a failure has none.
    let v = serde_json::from_slice::<Value>(&out.stdout)
        .map(|v| v["result"].clone())
        .unwrap_or(Value::Null);
    (out, v)
}
fn canon(p: &std::path::Path) -> String {
    fs::canonicalize(p).unwrap().to_str().unwrap().to_string()
}
fn committed_repo(dir: &std::path::Path) {
    git_in(dir, &["init", "-q", "-b", "main"]);
    write_file(dir, "a.txt", b"a\n");
    git_in(dir, &["add", "a.txt"]);
    git_in(dir, &["commit", "-q", "-m", "fixture"]);
}
/// REQ-01A-05: no returned string may point at the Go generation's state.
fn assert_no_go_state(v: &Value) {
    for (k, x) in v.as_object().unwrap() {
        if let Some(s) = x.as_str() {
            assert!(
                !s.ends_with("/.moai") && !s.contains("/.moai/") && !s.contains("MOAI_HOME"),
                "{k} points at Go state: {s}"
            );
        }
    }
}

#[test]
fn identity_primary_checkout() {
    let repo = TempDir::new().unwrap();
    let home = TempDir::new().unwrap();
    committed_repo(repo.path());
    fs::create_dir(repo.path().join("sub")).unwrap();
    let before = (tree_state(repo.path()), tree_state(home.path()));
    let (out, v) = ident(repo.path(), home.path(), None);
    assert!(out.status.success(), "{:?}", out.stderr);
    // AC-01C-1: identity answers with the same envelope as the other internal operations.
    let envelope: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(envelope["protocol_version"], json!(1));
    assert_eq!(envelope["operation"], json!("identity"));
    assert_eq!(envelope["ok"], json!(true));
    let git_dir = canon(&repo.path().join(".git"));
    assert_eq!(v["identity_status"], json!("ok"));
    assert_eq!(v["workspace_kind"], json!("primary"));
    assert_eq!(v["clone_anchor"], json!(git_dir));
    assert_eq!(v["workspace_git_dir"], json!(git_dir));
    assert_eq!(v["clone_id"], v["workspace_id"]);
    assert_eq!(v["clone_id"], json!(moai_cli::digest(git_dir.as_bytes())));
    assert_eq!(v["work_tree"], json!(canon(repo.path())));
    assert_eq!(
        v["project_state_root"],
        json!(format!("{}/.moai-cli", canon(repo.path())))
    );
    assert_eq!(
        v["home_state_root"],
        json!(home.path().join(".moai-cli").to_str().unwrap())
    );
    assert_eq!(v["profile_status"], json!("missing"));
    assert_eq!(v["project_id"], Value::Null);
    assert_eq!(v["writes"], json!(false));
    assert_no_go_state(&v);
    // A subdirectory resolves to the same workspace; two runs print identical bytes.
    let (sub, _) = ident(&repo.path().join("sub"), home.path(), None);
    assert_eq!(sub.stdout, out.stdout);
    assert_eq!(ident(repo.path(), home.path(), None).0.stdout, out.stdout);
    assert_eq!(
        before,
        (tree_state(repo.path()), tree_state(home.path())),
        "identity wrote to the repository or HOME"
    );
    assert!(!home.path().join(".moai-cli").exists());
    // The profile's project_id is reported separately from the path-derived ids.
    create(&repo, &profile(None));
    let (_, v) = ident(repo.path(), home.path(), None);
    assert_eq!(v["profile_status"], json!("present"));
    assert_eq!(v["project_id"], json!("demo"));
}

#[test]
fn identity_linked_worktree_shares_clone_id() {
    let repo = TempDir::new().unwrap();
    let side = TempDir::new().unwrap();
    let home = TempDir::new().unwrap();
    committed_repo(repo.path());
    let wt = side.path().join("wt");
    git_in(
        repo.path(),
        &["worktree", "add", "-q", "-b", "lane", wt.to_str().unwrap()],
    );
    let (o1, primary) = ident(repo.path(), home.path(), None);
    let (o2, linked) = ident(&wt, home.path(), None);
    assert!(o1.status.success() && o2.status.success());
    assert_eq!(primary["workspace_kind"], json!("primary"));
    assert_eq!(linked["workspace_kind"], json!("linked"));
    assert_eq!(primary["clone_id"], linked["clone_id"]);
    assert_eq!(primary["clone_anchor"], linked["clone_anchor"]);
    assert_ne!(primary["workspace_id"], linked["workspace_id"]);
    assert_ne!(linked["clone_anchor"], linked["workspace_git_dir"]);
    assert_eq!(linked["work_tree"], json!(canon(&wt)));
    assert_eq!(
        linked["project_state_root"],
        json!(format!("{}/.moai-cli", canon(&wt)))
    );
    assert_no_go_state(&linked);
}

#[test]
fn identity_bare_separate_gitdir_gitfile() {
    let base = TempDir::new().unwrap();
    let home = TempDir::new().unwrap();
    let b = base.path();
    // Bare repository: no work tree, so no project state root is offered.
    git_in(b, &["init", "-q", "--bare", "bare.git"]);
    let (out, v) = ident(&b.join("bare.git"), home.path(), None);
    assert!(out.status.success(), "{:?}", out.stderr);
    assert_eq!(v["workspace_kind"], json!("bare"));
    assert_eq!(v["clone_anchor"], json!(canon(&b.join("bare.git"))));
    assert_eq!(v["work_tree"], Value::Null);
    assert_eq!(v["project_state_root"], Value::Null);
    // --separate-git-dir: `.git` is a gitfile, the repository lives elsewhere.
    git_in(
        b,
        &[
            "init",
            "-q",
            "--separate-git-dir",
            b.join("sep.git").to_str().unwrap(),
            b.join("sep-wt").to_str().unwrap(),
        ],
    );
    assert!(b.join("sep-wt/.git").is_file());
    let (_, v) = ident(&b.join("sep-wt"), home.path(), None);
    assert_eq!(v["workspace_kind"], json!("primary"));
    assert_eq!(v["clone_anchor"], json!(canon(&b.join("sep.git"))));
    assert_eq!(v["work_tree"], json!(canon(&b.join("sep-wt"))));
    // Hand-written gitfile pointing at a moved repository.
    let m = b.join("moved");
    fs::create_dir(&m).unwrap();
    committed_repo(&m);
    fs::rename(m.join(".git"), b.join("moved.git")).unwrap();
    fs::write(
        m.join(".git"),
        format!("gitdir: {}\n", b.join("moved.git").to_str().unwrap()),
    )
    .unwrap();
    let (_, v) = ident(&m, home.path(), None);
    assert_eq!(v["workspace_kind"], json!("primary"));
    assert_eq!(v["clone_anchor"], json!(canon(&b.join("moved.git"))));
    assert_eq!(v["work_tree"], json!(canon(&m)));
    // Inside a non-bare git dir there is no work tree either.
    let (_, v) = ident(&b.join("moved.git"), home.path(), None);
    assert_eq!(v["workspace_kind"], json!("git_dir"));
    assert_eq!(v["project_state_root"], Value::Null);
}

#[test]
fn identity_outside_git_and_missing_git_binary() {
    let plain = TempDir::new().unwrap();
    let home = TempDir::new().unwrap();
    let (out, v) = ident(plain.path(), home.path(), None);
    assert!(out.status.success(), "{:?}", out.stderr);
    assert_eq!(v["identity_status"], json!("no_git"));
    for k in [
        "workspace_kind",
        "clone_anchor",
        "workspace_git_dir",
        "work_tree",
        "clone_id",
        "workspace_id",
        "project_id",
        "project_state_root",
    ] {
        assert_eq!(v[k], Value::Null, "{k} must not be guessed");
    }
    let repo = TempDir::new().unwrap();
    committed_repo(repo.path());
    let (out, v) = ident(repo.path(), home.path(), Some(""));
    assert!(out.status.success(), "{:?}", out.stderr);
    assert_eq!(v["identity_status"], json!("git_unavailable"));
    assert_eq!(v["clone_id"], Value::Null);
    assert_eq!(v["project_state_root"], Value::Null);
}

/// Issue #2033 AC-IDF-1: caller environment that would redirect git or change discovery
/// must not move any reported boundary.
#[test]
fn identity_ignores_injected_git_environment() {
    let repo = TempDir::new().unwrap();
    let elsewhere = TempDir::new().unwrap();
    let home = TempDir::new().unwrap();
    committed_repo(repo.path());
    fs::create_dir(repo.path().join("sub")).unwrap();
    let x = elsewhere.path().to_str().unwrap();
    let worktree_params = format!("'core.worktree'='{x}'");
    let ceiling = repo.path().to_str().unwrap();
    // Measured on git 2.54: injected core.bare=true turns a git dir into "bare" through
    // either config channel, and a ceiling at the repository hides it from a subdirectory.
    let cases: [&[(&str, &str)]; 6] = [
        &[
            ("GIT_CONFIG_COUNT", "1"),
            ("GIT_CONFIG_KEY_0", "core.worktree"),
            ("GIT_CONFIG_VALUE_0", x),
        ],
        &[
            ("GIT_CONFIG_COUNT", "1"),
            ("GIT_CONFIG_KEY_0", "core.bare"),
            ("GIT_CONFIG_VALUE_0", "true"),
        ],
        &[("GIT_CONFIG_PARAMETERS", &worktree_params)],
        &[("GIT_CONFIG_PARAMETERS", "'core.bare'='true'")],
        &[("GIT_CEILING_DIRECTORIES", ceiling)],
        &[("GIT_DISCOVERY_ACROSS_FILESYSTEM", "1")],
    ];
    for path in [repo.path().join("sub"), repo.path().join(".git")] {
        let (out, clean) = ident(&path, home.path(), None);
        assert!(out.status.success(), "{:?}", out.stderr);
        assert_eq!(clean["identity_status"], json!("ok"));
        for envs in cases {
            let (out, v) = ident_env(&path, home.path(), None, envs);
            assert!(out.status.success(), "{envs:?}: {:?}", out.stderr);
            assert_eq!(v, clean, "{path:?} {envs:?} moved the boundary");
        }
    }
}

/// Issue #2033 AC-IDF-2: git present but failing is a `git_failed` error envelope, not the
/// `no_git` status, and the message carries git's summary without any path.
#[test]
fn identity_reports_git_failure_separately_from_no_git() {
    let repo = TempDir::new().unwrap();
    let home = TempDir::new().unwrap();
    committed_repo(repo.path());
    // safe.directory refusal; GIT_CONFIG_GLOBAL=/dev/null leaves no exception configured.
    let (out, _) = ident_env(
        repo.path(),
        home.path(),
        None,
        &[("GIT_TEST_ASSUME_DIFFERENT_OWNER", "1")],
    );
    assert_eq!(out.status.code(), Some(1), "{:?}", out.stderr);
    let env: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(env["operation"], json!("identity"));
    assert_eq!(env["ok"], json!(false));
    assert!(env.get("result").is_none());
    let err = &env["error"];
    assert_eq!(err["kind"], json!("git_failed"));
    assert_eq!(err["retryable"], json!(false));
    let message = err["message"].as_str().unwrap();
    assert!(message.contains("dubious ownership"), "{message}");
    let stderr = String::from_utf8_lossy(&out.stderr);
    for text in [message, stderr.as_ref()] {
        assert!(
            !text.contains('/') && !text.contains('\''),
            "path echoed: {text}"
        );
    }
}

#[test]
fn identity_refuses_symlinks_and_bad_paths_without_echo() {
    let base = TempDir::new().unwrap();
    let home = TempDir::new().unwrap();
    let repo = base.path().join("repo");
    fs::create_dir(&repo).unwrap();
    committed_repo(&repo);
    #[cfg(unix)]
    let link = base.path().join("link");
    #[cfg(unix)]
    std::os::unix::fs::symlink(&repo, &link).unwrap();
    let missing = base.path().join("does-not-exist");
    let file = repo.join("a.txt");
    #[cfg(unix)]
    let bad_paths = [&link, &missing, &file];
    #[cfg(not(unix))]
    let bad_paths = [&missing, &file];
    for p in bad_paths {
        let (out, _) = ident(p, home.path(), None);
        assert!(!out.status.success(), "{p:?} must be refused");
        // A failure envelope on stdout (01-C #1990); neither stream echoes the path.
        let failure: Value = serde_json::from_slice(&out.stdout).unwrap();
        assert_eq!(failure["ok"], json!(false));
        assert_eq!(failure["operation"], json!("identity"));
        let expected = if p == &missing {
            "not_found"
        } else {
            "invalid_input"
        };
        assert_eq!(failure["error"]["kind"], json!(expected), "{p:?}");
        for stream in [&out.stdout, &out.stderr] {
            let text = String::from_utf8_lossy(stream);
            assert!(
                !text.contains(base.path().to_str().unwrap()),
                "path echoed: {text}"
            );
        }
    }
    #[cfg(unix)]
    {
        // A symlinked project state root is refused like target() refuses it.
        let elsewhere = TempDir::new().unwrap();
        std::os::unix::fs::symlink(elsewhere.path(), repo.join(".moai-cli")).unwrap();
        let (out, _) = ident(&repo, home.path(), None);
        assert!(!out.status.success());
        assert!(String::from_utf8_lossy(&out.stderr).contains("symlink"));
        fs::remove_file(repo.join(".moai-cli")).unwrap();
        // So is a symlinked home state root.
        std::os::unix::fs::symlink(elsewhere.path(), home.path().join(".moai-cli")).unwrap();
        let (out, _) = ident(&repo, home.path(), None);
        assert!(!out.status.success());
    }
}
