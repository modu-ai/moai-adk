//! SPEC-20261122-worktree-place REQ-8 / AC-6, the two places `worktree_place_older_cli.rs` leaves
//! out: the hooks command's home refusal (`src/hooks.rs`) has an exception for the worktree
//! folder of this generation's `~/.moai`, and the earlier common root
//! `<HOME>/.moai-cli/worktrees/<key>/<name>` (`old_root` in the move preview) is a fourth old
//! place. Everything runs in a temporary project with a temporary HOME marked as this generation's.

mod common;
use common::git::{git, git_ok};
use common::{CleanEnv, MOAI_BIN};
use serde_json::{json, Value};
use std::path::{Path, PathBuf};

struct World {
    _dir: tempfile::TempDir,
    base: PathBuf,
    repo: PathBuf,
    home: PathBuf,
}

/// A repository `proj` with one commit that ignores the state folders, and a HOME beside it
/// whose `~/.moai` carries this generation's marker (the shape `hooks_cli.rs` uses).
fn world() -> World {
    let dir = tempfile::tempdir().unwrap();
    let base = std::fs::canonicalize(dir.path()).unwrap();
    let (repo, home) = (base.join("proj"), base.join("home"));
    std::fs::create_dir_all(&repo).unwrap();
    std::fs::create_dir_all(home.join(".moai")).unwrap();
    std::fs::write(
        home.join(".moai/moai-cli.json"),
        r#"{"schema_version":1,"generation":"moai-cli"}"#,
    )
    .unwrap();
    git_ok(&repo, &["init", "-q", "-b", "main"]);
    git_ok(&repo, &["config", "core.excludesFile", "/dev/null"]);
    git_ok(&repo, &["config", "user.email", "t@example.invalid"]);
    git_ok(&repo, &["config", "user.name", "t"]);
    std::fs::write(repo.join("README.md"), "x\n").unwrap();
    std::fs::write(repo.join(".gitignore"), ".moai/\n.moai-cli/\ntarget/\n").unwrap();
    git_ok(&repo, &["add", "README.md", ".gitignore"]);
    git_ok(&repo, &["commit", "-q", "-m", "init"]);
    World {
        _dir: dir,
        base,
        repo,
        home,
    }
}

/// One worktree on a new branch whose HEAD a remote-tracking ref holds.
fn worktree(w: &World, path: &Path, branch: &str) {
    git_ok(
        &w.repo,
        &[
            "worktree",
            "add",
            "-q",
            "-b",
            branch,
            path.to_str().unwrap(),
        ],
    );
    git_ok(
        path,
        &[
            "update-ref",
            &format!("refs/remotes/origin/{branch}"),
            "HEAD",
        ],
    );
}

/// The whole envelope of one run in the project with the temporary HOME.
fn call(w: &World, args: &[&str]) -> Value {
    let mut c = std::process::Command::new(MOAI_BIN);
    c.args(args)
        .current_dir(&w.repo)
        .clean_env()
        .env("HOME", &w.home)
        .env("USERPROFILE", &w.home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null");
    let out = c.output().unwrap();
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "{args:?}: stdout is not JSON: {e}\n{}\n{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

/// The `result` of a run that must be `ok`.
fn run(w: &World, args: &[&str]) -> Value {
    let v = call(w, args);
    assert_eq!(v["ok"], json!(true), "{args:?}: {v}");
    v["result"].clone()
}

/// The entry of `list` (an array in `result`) whose `path` is `path`.
fn found<'a>(result: &'a Value, list: &str, path: &str) -> &'a Value {
    let want = json!(path);
    result[list]
        .as_array()
        .unwrap_or_else(|| panic!("no {list} in {result}"))
        .iter()
        .find(|e| e["path"] == want)
        .unwrap_or_else(|| panic!("no {list} entry for {path} in {result}"))
}

/// AC-6 (REQ-8), the two parts the listing test leaves out. (1) The hooks command accepts a
/// worktree in the home place and in the Go place of a HOME whose `~/.moai` is this generation's
/// (the exception of the home refusal in `src/hooks.rs`), and still refuses the state folder
/// beside them. (2) A worktree in the earlier common root `<HOME>/.moai-cli/worktrees/<key>/<name>`
/// is listed by the space report and the clean-up plan, kept with the place named `old_root` by
/// the move preview, and none of the commands changes anything. What the hooks command says about
/// the earlier root is not pinned here: it refuses it today (a leader question, see the card).
#[test]
fn worktree_place_old_places_hooks_exception_and_the_earlier_root() {
    let w = world();
    let home_wt = w.home.join(".moai/worktrees/proj-home");
    let go_wt = w.home.join(".moai/worktrees/projects/proj/feat");
    let old_wt = w.home.join(".moai-cli/worktrees/proj-0123456789ab/old");
    for (path, branch) in [(&home_wt, "at-home"), (&go_wt, "feat"), (&old_wt, "old")] {
        worktree(&w, path, branch);
    }
    let listed = git(&w.repo, &["worktree", "list", "--porcelain"]);

    for path in [&home_wt, &go_wt] {
        let v = call(
            &w,
            &["hooks-plan", path.to_str().unwrap(), "--host", "claude"],
        );
        assert_eq!(v["ok"], json!(true), "{}: {v}", path.display());
    }
    let state = w.home.join(".moai/state/p");
    std::fs::create_dir_all(&state).unwrap();
    let refused = call(
        &w,
        &["hooks-plan", state.to_str().unwrap(), "--host", "claude"],
    );
    assert_eq!(refused["ok"], json!(false), "{refused}");
    assert!(
        refused["error"]["message"]
            .as_str()
            .unwrap()
            .contains("home directory refused"),
        "{refused}"
    );

    let project = w.repo.to_str().unwrap();
    let old = old_wt.to_str().unwrap();
    let tmp = w.base.join("tmp");
    std::fs::create_dir_all(&tmp).unwrap();
    let report = run(
        &w,
        &[
            "worktree-report",
            project,
            "--tmp-root",
            tmp.to_str().unwrap(),
        ],
    );
    assert_eq!(
        found(&report, "worktrees", old)["branch"],
        json!("old"),
        "{report}"
    );
    let plan = run(
        &w,
        &["worktree-gc-plan", project, "--integration-ref", "main"],
    );
    assert_eq!(
        found(&plan, "candidates", old)["branch"],
        json!("old"),
        "{plan}"
    );
    let observations = w.base.join("observations.json");
    std::fs::write(
        &observations,
        json!({"schema_version": 1, "holds": {"source": "fixture", "branches": []}}).to_string(),
    )
    .unwrap();
    let moves = run(
        &w,
        &[
            "worktree-move-plan",
            project,
            "--observations",
            observations.to_str().unwrap(),
        ],
    );
    let item = found(&moves, "items", old);
    assert_eq!(item["action"], json!("keep"), "{moves}");
    assert_eq!(item["location"], json!("old_root"), "{moves}");
    assert_eq!(moves["totals"]["move"], json!(0), "{moves}");

    for path in [&home_wt, &go_wt, &old_wt] {
        assert!(path.join("README.md").is_file(), "{}", path.display());
    }
    assert_eq!(git(&w.repo, &["worktree", "list", "--porcelain"]), listed);
    assert_eq!(git(&w.repo, &["status", "--porcelain"]), "");
}
