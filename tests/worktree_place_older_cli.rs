//! SPEC-20261122-worktree-place REQ-8 / AC-6, the part the classifier test cannot show: a
//! worktree in an old place (the home folder, the Go generation folder, the earlier project
//! state folder) keeps working for the commands that read where a worktree lies. "Every
//! command" is read as the three read-only ones that list a project's worktrees,
//! `worktree-report`, `worktree-gc-plan` and `worktree-move-plan`; the two commands that write
//! are held by `new_place_worktrees_are_cleaned_and_go_place_ones_never_touched`
//! (`tests/worktree_gc_apply.rs`) and `new_old_root_and_go_place_worktrees_are_kept`
//! (`tests/worktree_move.rs`). Everything runs in a temporary project with a temporary HOME.

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

/// A repository `proj` with one commit that ignores the state folders, and a HOME beside it.
fn world() -> World {
    let dir = tempfile::tempdir().unwrap();
    let base = std::fs::canonicalize(dir.path()).unwrap();
    let (repo, home) = (base.join("proj"), base.join("home"));
    std::fs::create_dir_all(&repo).unwrap();
    std::fs::create_dir_all(&home).unwrap();
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

/// One worktree on a new branch whose HEAD a remote-tracking ref holds, so nothing but its
/// place decides what a command says about it.
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

/// The answer of one run in the project with the temporary HOME; a run that is not `ok` fails.
fn run(w: &World, args: &[&str]) -> Value {
    let mut c = std::process::Command::new(MOAI_BIN);
    c.args(args)
        .current_dir(&w.repo)
        .clean_env()
        .env("HOME", &w.home)
        .env("USERPROFILE", &w.home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null");
    let out = c.output().unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "{args:?}: stdout is not JSON: {e}\n{}\n{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
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

/// How the report and the clean-up plan name a worktree: relative to the project when it lies
/// inside it, as written otherwise (the move preview always writes the whole path).
fn shown(w: &World, path: &Path) -> String {
    path.strip_prefix(&w.repo)
        .unwrap_or(path)
        .to_str()
        .unwrap()
        .to_string()
}

/// AC-6 (REQ-8): a worktree in each old place is listed by the space report, planned by the
/// clean-up with its place named (the Go generation's one kept as `preserved_location`) and
/// judged by the move preview without a root (the home and Go ones kept, the earlier project
/// state one still a move candidate), and none of the three commands changes anything.
#[test]
fn worktree_place_old_places_keep_working_for_the_listing_commands() {
    let w = world();
    let home_wt = w.home.join(".moai/worktrees/proj-home");
    let go_wt = w.home.join(".moai/worktrees/projects/proj/feat");
    let project_wt = w.repo.join(".moai-cli/worktrees/clean");
    for (path, branch) in [
        (&home_wt, "at-home"),
        (&go_wt, "feat"),
        (&project_wt, "clean"),
    ] {
        worktree(&w, path, branch);
    }
    let listed = git(&w.repo, &["worktree", "list", "--porcelain"]);
    let tmp = w.base.join("tmp");
    std::fs::create_dir_all(&tmp).unwrap();
    let observations = w.base.join("observations.json");
    std::fs::write(
        &observations,
        json!({"schema_version": 1, "holds": {"source": "fixture", "branches": []}}).to_string(),
    )
    .unwrap();
    let project = w.repo.to_str().unwrap();

    let report = run(
        &w,
        &[
            "worktree-report",
            project,
            "--tmp-root",
            tmp.to_str().unwrap(),
        ],
    );
    for (path, branch) in [
        (&home_wt, "at-home"),
        (&go_wt, "feat"),
        (&project_wt, "clean"),
    ] {
        assert_eq!(
            found(&report, "worktrees", &shown(&w, path))["branch"],
            json!(branch),
            "{report}"
        );
    }

    let plan = run(
        &w,
        &["worktree-gc-plan", project, "--integration-ref", "main"],
    );
    for (path, location) in [(&home_wt, "home"), (&go_wt, "go"), (&project_wt, "project")] {
        let c = found(&plan, "candidates", &shown(&w, path));
        assert_eq!(c["location"], json!(location), "{plan}");
    }
    let go = found(&plan, "candidates", &shown(&w, &go_wt));
    assert_eq!(go["action"], json!("skip"), "{go}");
    assert!(
        go["reasons"]
            .as_array()
            .unwrap()
            .contains(&json!("preserved_location")),
        "{go}"
    );

    let moves = run(
        &w,
        &[
            "worktree-move-plan",
            project,
            "--observations",
            observations.to_str().unwrap(),
        ],
    );
    for (path, location) in [(&home_wt, "home"), (&go_wt, "go")] {
        let item = found(&moves, "items", path.to_str().unwrap());
        assert_eq!(item["action"], json!("keep"), "{moves}");
        assert_eq!(item["location"], json!(location), "{moves}");
    }
    assert_eq!(
        found(&moves, "items", project_wt.to_str().unwrap())["action"],
        json!("move"),
        "{moves}"
    );

    for path in [&home_wt, &go_wt, &project_wt] {
        assert!(path.join("README.md").is_file(), "{}", path.display());
    }
    assert_eq!(git(&w.repo, &["worktree", "list", "--porcelain"]), listed);
    assert_eq!(git(&w.repo, &["status", "--porcelain"]), "");
}
