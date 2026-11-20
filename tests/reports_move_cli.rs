//! `reports_move` (C-126, SPEC-20261119-update-moves-reports): the update moves the untracked
//! report files an earlier generation left in the repository to the user-level reports folder.
//! Every test runs in a temporary HOME and a temporary repository; none touches the real HOME.
//! Git runs through the `common::git` helpers (environment scrubbed), never bare.

mod common;

use common::git::{git, git_ok};
use common::home::TempHome;
use moai_cli::reports_move::{apply, plan, MovePlan};
use std::fs;
use std::path::{Path, PathBuf};

const PERMISSION: &str = "test-approval";

/// A temporary repository with one commit. Returns the (canonical) root.
fn repo(parent: &Path) -> PathBuf {
    let root = parent.join("project");
    fs::create_dir_all(&root).unwrap();
    git_ok(&root, &["init", "-q"]);
    git_ok(&root, &["config", "user.email", "t@example.invalid"]);
    git_ok(&root, &["config", "user.name", "t"]);
    fs::write(root.join("README.md"), "x\n").unwrap();
    git_ok(&root, &["add", "README.md"]);
    git_ok(&root, &["commit", "-q", "-m", "init"]);
    fs::canonicalize(&root).unwrap()
}

fn write(root: &Path, rel: &str, body: &str) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, body).unwrap();
}

fn apply_plan(root: &Path, home: &Path, p: &MovePlan) -> Result<serde_json::Value, String> {
    apply(root, home, p.clone(), &p.plan_digest, PERMISSION)
}

#[test]
fn untracked_report_files_move_with_their_folders_and_tracked_ones_stay() {
    let h = TempHome::new();
    let tmp = tempfile::tempdir().unwrap();
    let root = repo(tmp.path());
    write(&root, ".moai-cli/reports/a.md", "a");
    write(&root, ".moai-cli/reports/sub/b.md", "bb");
    write(&root, ".moai/reports/c.md", "ccc");
    write(&root, ".moai-cli/reports/kept.md", "tracked");
    git_ok(&root, &["add", "-f", ".moai-cli/reports/kept.md"]);
    git_ok(&root, &["commit", "-q", "-m", "track one"]);

    let p = plan(&root, h.path()).unwrap();
    assert_eq!(p.state, "ready", "{}", p.reason);
    assert_eq!(p.moves.len(), 3);
    assert_eq!(p.tracked, 1);
    assert_eq!(p.skipped_existing.len(), 0);
    let dest = PathBuf::from(&p.destination);
    assert!(dest.starts_with(h.path().join(".moai").join("reports")));
    // The preview alone creates nothing in the HOME.
    assert!(!dest.exists());

    let done = apply_plan(&root, h.path(), &p).unwrap();
    assert_eq!(done["moved"], 3);
    assert_eq!(fs::read_to_string(dest.join("a.md")).unwrap(), "a");
    assert_eq!(fs::read_to_string(dest.join("sub/b.md")).unwrap(), "bb");
    assert_eq!(fs::read_to_string(dest.join("c.md")).unwrap(), "ccc");
    // Sources gone, the tracked file untouched and still tracked.
    assert!(!root.join(".moai-cli/reports/a.md").exists());
    assert!(!root.join(".moai/reports/c.md").exists());
    assert_eq!(
        fs::read_to_string(root.join(".moai-cli/reports/kept.md")).unwrap(),
        "tracked"
    );
    assert_eq!(
        git(&root, &["ls-files", ".moai-cli/reports/kept.md"]).trim(),
        ".moai-cli/reports/kept.md"
    );
}

#[cfg(unix)]
#[test]
fn moved_files_and_folders_are_owner_only() {
    use std::os::unix::fs::PermissionsExt;
    let h = TempHome::new();
    let tmp = tempfile::tempdir().unwrap();
    let root = repo(tmp.path());
    write(&root, ".moai-cli/reports/sub/a.md", "a");
    fs::set_permissions(
        root.join(".moai-cli/reports/sub/a.md"),
        fs::Permissions::from_mode(0o664),
    )
    .unwrap();
    let p = plan(&root, h.path()).unwrap();
    apply_plan(&root, h.path(), &p).unwrap();
    let dest = PathBuf::from(&p.destination);
    let mode = |path: PathBuf| fs::metadata(path).unwrap().permissions().mode() & 0o777;
    assert_eq!(mode(dest.join("sub/a.md")) & 0o077, 0);
    assert_eq!(mode(dest.join("sub")) & 0o077, 0);
    assert_eq!(mode(dest.clone()) & 0o077, 0);
}

#[test]
fn a_name_the_destination_holds_is_kept_and_counted_never_overwritten() {
    let h = TempHome::new();
    let tmp = tempfile::tempdir().unwrap();
    let root = repo(tmp.path());
    write(&root, ".moai-cli/reports/a.md", "from the repository");
    write(&root, ".moai-cli/reports/b.md", "b");
    let first = plan(&root, h.path()).unwrap();
    let dest = PathBuf::from(&first.destination);
    fs::create_dir_all(&dest).unwrap();
    fs::write(dest.join("a.md"), "already there").unwrap();

    let p = plan(&root, h.path()).unwrap();
    assert_eq!(p.moves.len(), 1);
    assert_eq!(p.skipped_existing, vec!["a.md".to_string()]);
    let done = apply_plan(&root, h.path(), &p).unwrap();
    assert_eq!(done["moved"], 1);
    assert_eq!(done["skipped_existing"], 1);
    assert_eq!(
        fs::read_to_string(dest.join("a.md")).unwrap(),
        "already there"
    );
    assert_eq!(
        fs::read_to_string(root.join(".moai-cli/reports/a.md")).unwrap(),
        "from the repository"
    );
    assert!(!root.join(".moai-cli/reports/b.md").exists());
}

#[test]
fn a_second_run_moves_nothing_and_changes_nothing() {
    let h = TempHome::new();
    let tmp = tempfile::tempdir().unwrap();
    let root = repo(tmp.path());
    write(&root, ".moai-cli/reports/a.md", "a");
    let p = plan(&root, h.path()).unwrap();
    apply_plan(&root, h.path(), &p).unwrap();

    let before = h.snapshot();
    let again = plan(&root, h.path()).unwrap();
    assert_eq!(again.state, "unchanged");
    assert_eq!(again.moves.len(), 0);
    // An applied plan is stale once the files have moved.
    let err = apply_plan(&root, h.path(), &p).unwrap_err();
    assert!(err.contains("stale or altered plan"), "{err}");
    assert_eq!(h.snapshot(), before);
}

#[test]
fn the_same_name_in_both_earlier_folders_moves_the_first_and_keeps_the_second() {
    let h = TempHome::new();
    let tmp = tempfile::tempdir().unwrap();
    let root = repo(tmp.path());
    write(&root, ".moai-cli/reports/x.md", "state folder");
    write(&root, ".moai/reports/x.md", "older folder");
    let p = plan(&root, h.path()).unwrap();
    assert_eq!(p.moves.len(), 1);
    assert_eq!(p.skipped_existing, vec!["x.md".to_string()]);
    apply_plan(&root, h.path(), &p).unwrap();
    assert_eq!(
        fs::read_to_string(PathBuf::from(&p.destination).join("x.md")).unwrap(),
        "state folder"
    );
    assert_eq!(
        fs::read_to_string(root.join(".moai/reports/x.md")).unwrap(),
        "older folder"
    );
}

#[cfg(unix)]
#[test]
fn a_link_is_neither_followed_nor_moved() {
    let h = TempHome::new();
    let tmp = tempfile::tempdir().unwrap();
    let root = repo(tmp.path());
    let outside = tmp.path().join("outside.md");
    fs::write(&outside, "outside").unwrap();
    write(&root, ".moai-cli/reports/a.md", "a");
    std::os::unix::fs::symlink(&outside, root.join(".moai-cli/reports/link.md")).unwrap();
    let p = plan(&root, h.path()).unwrap();
    assert_eq!(p.moves.len(), 1);
    assert_eq!(p.unmovable, 1);
    apply_plan(&root, h.path(), &p).unwrap();
    assert!(fs::symlink_metadata(root.join(".moai-cli/reports/link.md")).is_ok());
    assert_eq!(fs::read_to_string(&outside).unwrap(), "outside");
    assert!(!PathBuf::from(&p.destination).join("link.md").exists());
}

#[test]
fn a_plan_made_stale_by_a_new_file_or_a_wrong_digest_or_reference_is_refused() {
    let h = TempHome::new();
    let tmp = tempfile::tempdir().unwrap();
    let root = repo(tmp.path());
    write(&root, ".moai-cli/reports/a.md", "a");
    let p = plan(&root, h.path()).unwrap();

    let err = apply(&root, h.path(), p.clone(), &p.plan_digest, "placeholder").unwrap_err();
    assert!(err.contains("permission reference"), "{err}");
    let err = apply(&root, h.path(), p.clone(), "0000", PERMISSION).unwrap_err();
    assert!(err.contains("stale or altered plan"), "{err}");
    let mut altered = p.clone();
    altered.moves[0].to = "elsewhere.md".into();
    let err = apply(
        &root,
        h.path(),
        altered.clone(),
        &altered.plan_digest,
        PERMISSION,
    )
    .unwrap_err();
    assert!(err.contains("stale or altered plan"), "{err}");

    write(&root, ".moai-cli/reports/new.md", "new");
    let err = apply_plan(&root, h.path(), &p).unwrap_err();
    assert!(err.contains("stale or altered plan"), "{err}");
    // Nothing moved in any refused case.
    assert!(root.join(".moai-cli/reports/a.md").exists());
    assert!(!PathBuf::from(&p.destination).exists());
}

#[test]
fn a_folder_git_cannot_read_blocks_the_move_and_nothing_is_created() {
    let h = TempHome::new();
    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path().join("not-a-repo");
    write(&root, ".moai-cli/reports/a.md", "a");
    let root = fs::canonicalize(&root).unwrap();
    let before = h.snapshot();
    let p = plan(&root, h.path()).unwrap();
    assert_eq!(p.state, "blocked", "{}", p.reason);
    assert!(p.moves.is_empty());
    let err = apply_plan(&root, h.path(), &p).unwrap_err();
    assert!(err.contains("blocked"), "{err}");
    assert!(root.join(".moai-cli/reports/a.md").exists());
    assert_eq!(h.snapshot(), before);
}

#[test]
fn no_earlier_folder_is_unchanged_and_preview_counts_are_reported() {
    let h = TempHome::new();
    let tmp = tempfile::tempdir().unwrap();
    let root = repo(tmp.path());
    let none = plan(&root, h.path()).unwrap();
    assert_eq!(none.state, "unchanged");
    assert!(none.sources.is_empty());

    write(&root, ".moai-cli/reports/a.md", "a");
    write(&root, ".moai-cli/reports/t.md", "t");
    git_ok(&root, &["add", "-f", ".moai-cli/reports/t.md"]);
    git_ok(&root, &["commit", "-q", "-m", "track"]);
    let p = plan(&root, h.path()).unwrap();
    let v = p.preview();
    assert_eq!(v["to_move"], 1);
    assert_eq!(v["tracked_left_alone"], 1);
    assert_eq!(v["skipped_existing"], 0);
    assert_eq!(v["examples"][0], ".moai-cli/reports/a.md");
}

/// `moai ARGS` in the temporary HOME with `stdin` on its input; the parsed success envelope.
fn run_op(h: &TempHome, args: &[&str], stdin: &[u8]) -> (bool, serde_json::Value) {
    use std::process::Stdio;
    let mut child = h
        .command(args)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, stdin);
    let out = child.wait_with_output().unwrap();
    let v: serde_json::Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|e| panic!("{e}: {}", String::from_utf8_lossy(&out.stderr)));
    (out.status.success(), v)
}

#[test]
fn update_plan_shows_the_reports_step() {
    let h = TempHome::new();
    let tmp = tempfile::tempdir().unwrap();
    let root = repo(tmp.path());
    let project = root.to_str().unwrap();
    // No earlier folder: the step is there and has nothing to do.
    let (ok, v) = run_op(&h, &["project-update-plan", project], b"");
    assert!(ok, "{v}");
    let steps = v["result"]["phases"][0]["steps"].as_array().unwrap();
    let names: Vec<&str> = steps.iter().map(|s| s["step"].as_str().unwrap()).collect();
    assert_eq!(names, ["go-settings", "go-state", "reports"], "{v}");
    assert_eq!(steps[2]["state"], "unchanged", "{}", steps[2]);
    assert!(steps[2]["apply"].is_null(), "{}", steps[2]);

    write(&root, ".moai-cli/reports/a.md", "a");
    write(&root, ".moai/reports/b.md", "bb");
    let (ok, v) = run_op(&h, &["project-update-plan", project], b"");
    assert!(ok, "{v}");
    let step = &v["result"]["phases"][0]["steps"][2];
    assert_eq!(step["step"], "reports", "{step}");
    assert_eq!(step["state"], "ready", "{step}");
    assert_eq!(step["preview"]["to_move"], 2, "{step}");
    assert_eq!(step["preview"]["tracked_left_alone"], 0, "{step}");
    let digest = step["plan_digest"].as_str().unwrap();
    let argv: Vec<&str> = step["apply"]["argv"]
        .as_array()
        .unwrap()
        .iter()
        .map(|a| a.as_str().unwrap())
        .collect();
    assert_eq!(argv[1], "reports-move-apply", "{step}");
    assert_eq!(argv[3], "--plan-digest", "{step}");
    assert_eq!(argv[4], digest, "{step}");
    assert_eq!(step["apply"]["stdin"]["plan_digest"], digest, "{step}");
    // The preview wrote nothing: the files are where they were and the HOME has no reports.
    assert!(root.join(".moai-cli/reports/a.md").exists());
    assert!(!h.path().join(".moai").join("reports").exists());
}

#[test]
fn the_two_operations_move_through_the_cli_and_refuse_a_wrong_digest() {
    let h = TempHome::new();
    let tmp = tempfile::tempdir().unwrap();
    let root = repo(tmp.path());
    let project = root.to_str().unwrap();
    write(&root, ".moai-cli/reports/sub/a.md", "a");
    let (ok, planned) = run_op(&h, &["reports-move-plan", project], b"");
    assert!(ok, "{planned}");
    let p = planned["result"].clone();
    assert_eq!(p["state"], "ready", "{p}");
    let digest = p["plan_digest"].as_str().unwrap().to_string();
    let stdin = p.to_string();

    let (ok, refused) = run_op(
        &h,
        &[
            "reports-move-apply",
            project,
            "--plan-digest",
            "0000",
            "--permission-reference",
            PERMISSION,
        ],
        stdin.as_bytes(),
    );
    assert!(!ok, "{refused}");
    assert!(root.join(".moai-cli/reports/sub/a.md").exists());

    let (ok, applied) = run_op(
        &h,
        &[
            "reports-move-apply",
            project,
            "--plan-digest",
            &digest,
            "--permission-reference",
            PERMISSION,
        ],
        stdin.as_bytes(),
    );
    assert!(ok, "{applied}");
    assert_eq!(applied["result"]["moved"], 1, "{applied}");
    assert!(!root.join(".moai-cli/reports/sub/a.md").exists());
    let dest = PathBuf::from(p["destination"].as_str().unwrap());
    assert_eq!(fs::read_to_string(dest.join("sub/a.md")).unwrap(), "a");

    let (ok, again) = run_op(&h, &["reports-move-plan", project], b"");
    assert!(ok, "{again}");
    assert_eq!(again["result"]["state"], "unchanged", "{again}");
}

// ---- tier 2 audit repair: links in the path, linked worktrees, the HOME, the limits ----

/// A second repository whose tracked report file a link must never reach.
fn other_repo_with_tracked_report(parent: &Path) -> (PathBuf, PathBuf) {
    let other = parent.join("other");
    fs::create_dir_all(&other).unwrap();
    git_ok(&other, &["init", "-q"]);
    git_ok(&other, &["config", "user.email", "t@example.invalid"]);
    git_ok(&other, &["config", "user.name", "t"]);
    write(&other, ".moai-cli/reports/tracked.md", "tracked");
    git_ok(&other, &["add", "-f", ".moai-cli/reports/tracked.md"]);
    git_ok(&other, &["commit", "-q", "-m", "report"]);
    write(&other, ".moai-cli/reports/untracked.md", "untracked");
    let other = fs::canonicalize(&other).unwrap();
    let tracked = other.join(".moai-cli/reports/tracked.md");
    (other, tracked)
}

/// Audit B1: only the last name of the earlier folder was checked, so a linked `.moai-cli` was
/// followed into another repository, git saw no tracked file there, and that repository's
/// tracked reports were "moved" and deleted. Every component of the path is now a plain folder
/// or the plan is blocked.
#[cfg(unix)]
#[test]
fn a_linked_state_folder_is_never_followed_into_another_repository() {
    for link_name in [".moai-cli", ".moai"] {
        let h = TempHome::new();
        let tmp = tempfile::tempdir().unwrap();
        let root = repo(tmp.path());
        let (other, tracked) = other_repo_with_tracked_report(tmp.path());
        if link_name == ".moai" {
            write(&other, ".moai/reports/tracked.md", "tracked");
            git_ok(&other, &["add", "-f", ".moai/reports/tracked.md"]);
            git_ok(&other, &["commit", "-q", "-m", "moai report"]);
        }
        std::os::unix::fs::symlink(other.join(link_name), root.join(link_name)).unwrap();
        let before = h.snapshot();
        let p = plan(&root, h.path()).unwrap();
        assert_eq!(p.state, "blocked", "{link_name}: {}", p.reason);
        assert!(p.moves.is_empty(), "{link_name}");
        assert!(p.reason.contains(link_name), "{link_name}: {}", p.reason);
        let err = apply_plan(&root, h.path(), &p).unwrap_err();
        assert!(err.contains("blocked"), "{err}");
        assert_eq!(fs::read_to_string(&tracked).unwrap(), "tracked");
        assert!(other.join(".moai-cli/reports/untracked.md").exists());
        assert_eq!(
            h.snapshot(),
            before,
            "{link_name}: the HOME stays untouched"
        );
    }
}

/// Audit B1, the same hole with a real folder: a nested repository's own tracked files are not
/// this repository's, so git run from the project lists none of them as tracked.
#[test]
fn a_report_folder_inside_a_nested_repository_is_never_moved() {
    let h = TempHome::new();
    let tmp = tempfile::tempdir().unwrap();
    let root = repo(tmp.path());
    let nested = root.join(".moai-cli");
    fs::create_dir_all(&nested).unwrap();
    git_ok(&nested, &["init", "-q"]);
    git_ok(&nested, &["config", "user.email", "t@example.invalid"]);
    git_ok(&nested, &["config", "user.name", "t"]);
    write(&root, ".moai-cli/reports/tracked.md", "tracked");
    git_ok(&nested, &["add", "-f", "reports/tracked.md"]);
    git_ok(&nested, &["commit", "-q", "-m", "nested"]);
    let p = plan(&root, h.path()).unwrap();
    assert_eq!(p.state, "blocked", "{}", p.reason);
    assert!(p.moves.is_empty());
    assert!(root.join(".moai-cli/reports/tracked.md").exists());
}

/// Audit B5: in a linked worktree the destination names the primary checkout, as
/// `optimize-plan` does, not the worktree's folder.
#[test]
fn a_linked_worktree_moves_to_the_primary_checkouts_folder() {
    let h = TempHome::new();
    let tmp = tempfile::tempdir().unwrap();
    let primary = repo(tmp.path());
    let wt = tmp.path().join("a-feature-worktree");
    git_ok(
        &primary,
        &[
            "worktree",
            "add",
            "-q",
            wt.to_str().unwrap(),
            "-b",
            "feature",
        ],
    );
    let wt = fs::canonicalize(&wt).unwrap();
    write(&wt, ".moai-cli/reports/a.md", "a");
    let p = plan(&wt, h.path()).unwrap();
    let want = moai_cli::home_workdir::reports_dir(h.path(), &primary).unwrap();
    assert_eq!(p.state, "ready", "{}", p.reason);
    assert_eq!(PathBuf::from(&p.destination), want, "{}", p.destination);
    apply_plan(&wt, h.path(), &p).unwrap();
    assert_eq!(fs::read_to_string(want.join("a.md")).unwrap(), "a");
    assert!(!wt.join(".moai-cli/reports/a.md").exists());
}

/// Audit N1: a reports folder inside the project would write the reports back into the
/// repository; it blocks the plan and creates nothing. The name still says "relative" because an
/// earlier SPEC's verify line runs it by that name; the relative HOME case is tried in a
/// temporary working folder by `every_blocked_apply_reports_invalid_input_through_the_cli`
/// (resolved here it would be taken against this repository).
#[test]
fn a_home_that_is_relative_or_inside_the_project_blocks_the_move() {
    let tmp = tempfile::tempdir().unwrap();
    let root = repo(tmp.path());
    write(&root, ".moai-cli/reports/a.md", "a");
    for home in [root.join("in-repo-home")] {
        let p = plan(&root, &home).unwrap();
        assert_eq!(p.state, "blocked", "{}: {}", home.display(), p.reason);
        assert!(p.moves.is_empty());
        assert!(
            p.reason.contains("HOME") || p.reason.contains("inside the project"),
            "{}",
            p.reason
        );
        let err = apply_plan(&root, &home, &p).unwrap_err();
        assert!(err.contains("blocked"), "{err}");
        assert!(root.join(".moai-cli/reports/a.md").exists());
        assert!(!root.join("in-repo-home").exists());
    }
}

/// Audit B2: the file limit is exact — 20,000 files plan, 20,001 block (boundary N and N+1).
#[test]
fn the_file_limit_holds_at_n_and_blocks_at_n_plus_one() {
    let h = TempHome::new();
    let tmp = tempfile::tempdir().unwrap();
    let root = repo(tmp.path());
    let limit = moai_cli::reports_move::MAX_FILES;
    assert_eq!(limit, 20_000);
    for i in 0..limit {
        let dir = root.join(format!(".moai-cli/reports/d{}", i % 100));
        if i < 100 {
            fs::create_dir_all(&dir).unwrap();
        }
        fs::write(dir.join(format!("{i}.md")), "x").unwrap();
    }
    let p = plan(&root, h.path()).unwrap();
    assert_eq!(p.state, "ready", "{}", p.reason);
    assert_eq!(p.moves.len(), limit);
    write(&root, ".moai-cli/reports/d0/one-more.md", "x");
    let p = plan(&root, h.path()).unwrap();
    assert_eq!(p.state, "blocked", "{}", p.reason);
    assert!(p.moves.is_empty());
    assert!(p.reason.contains("20000"), "{}", p.reason);
    assert!(root.join(".moai-cli/reports/d0/one-more.md").exists());
}

/// Audit B3: the destination chain — `<HOME>/.moai`, `<HOME>/.moai/reports` and the project's
/// own folder — must be plain folders; a link or a file there blocks the plan and nothing is
/// written through it.
#[cfg(unix)]
#[test]
fn a_destination_that_is_a_link_or_a_file_blocks_the_move() {
    for level in ["state", "reports", "project", "project-file"] {
        let h = TempHome::new();
        let tmp = tempfile::tempdir().unwrap();
        let root = repo(tmp.path());
        write(&root, ".moai-cli/reports/a.md", "a");
        let aside = tmp.path().join("aside");
        fs::create_dir_all(&aside).unwrap();
        let state = moai_cli::home_workdir::root(h.path());
        let reports = moai_cli::home_workdir::reports_root(h.path());
        let dest = moai_cli::home_workdir::reports_dir(h.path(), &root).unwrap();
        let bad = match level {
            "state" => state.clone(),
            "reports" => {
                fs::create_dir_all(&state).unwrap();
                reports.clone()
            }
            _ => {
                fs::create_dir_all(&reports).unwrap();
                dest.clone()
            }
        };
        if level == "project-file" {
            fs::write(&bad, "a file, not a folder").unwrap();
        } else {
            std::os::unix::fs::symlink(&aside, &bad).unwrap();
        }
        let p = plan(&root, h.path()).unwrap();
        assert_eq!(p.state, "blocked", "{level}: {}", p.reason);
        assert!(p.moves.is_empty(), "{level}");
        assert!(
            p.reason.contains("not a plain folder"),
            "{level}: {}",
            p.reason
        );
        let err = apply_plan(&root, h.path(), &p).unwrap_err();
        assert!(err.contains("blocked"), "{level}: {err}");
        assert!(root.join(".moai-cli/reports/a.md").exists(), "{level}");
        assert_eq!(
            fs::read_dir(&aside).unwrap().count(),
            0,
            "{level}: wrote through the link"
        );
    }
}

/// Audit N8: a printed preview names at most ten files (the boundary at ten and eleven).
#[test]
fn the_preview_names_at_most_ten_files() {
    let h = TempHome::new();
    let tmp = tempfile::tempdir().unwrap();
    let root = repo(tmp.path());
    for i in 0..10 {
        write(&root, &format!(".moai-cli/reports/{i:02}.md"), "x");
    }
    let v = plan(&root, h.path()).unwrap().preview();
    assert_eq!(v["to_move"], 10);
    assert_eq!(v["examples"].as_array().unwrap().len(), 10);
    write(&root, ".moai-cli/reports/10.md", "x");
    let v = plan(&root, h.path()).unwrap().preview();
    assert_eq!(v["to_move"], 11);
    assert_eq!(
        v["examples"].as_array().unwrap().len(),
        moai_cli::reports_move::EXAMPLES
    );
}

/// Audit B1, a link that stays inside this repository: git run from the project would list the
/// folder's files through the link as ordinary paths, so the link itself blocks the plan.
#[cfg(unix)]
#[test]
fn a_linked_state_folder_inside_the_same_repository_blocks_the_move() {
    let h = TempHome::new();
    let tmp = tempfile::tempdir().unwrap();
    let root = repo(tmp.path());
    write(&root, "stash/reports/a.md", "a");
    std::os::unix::fs::symlink(root.join("stash"), root.join(".moai-cli")).unwrap();
    let p = plan(&root, h.path()).unwrap();
    assert_eq!(p.state, "blocked", "{}", p.reason);
    assert!(p.moves.is_empty());
    assert!(p.reason.contains("is a link"), "{}", p.reason);
    assert!(root.join("stash/reports/a.md").exists());
}

/// Audit R1: a git repository of its own further down the report folder (its `.git` a folder, or
/// a file as a worktree or a submodule has) is not this project's report: the plan is blocked
/// and nothing of it moves, so the repository stays where it was.
#[test]
fn a_git_repository_below_the_report_folder_blocks_the_move() {
    for kind in ["folder", "file"] {
        let h = TempHome::new();
        let tmp = tempfile::tempdir().unwrap();
        let root = repo(tmp.path());
        write(&root, ".moai-cli/reports/a.md", "a");
        let sub = root.join(".moai-cli/reports/sub");
        write(&root, ".moai-cli/reports/sub/tracked.md", "tracked");
        if kind == "folder" {
            git_ok(&sub, &["init", "-q"]);
            git_ok(&sub, &["config", "user.email", "t@example.invalid"]);
            git_ok(&sub, &["config", "user.name", "t"]);
            git_ok(&sub, &["add", "-f", "tracked.md"]);
            git_ok(&sub, &["commit", "-q", "-m", "inner"]);
        } else {
            fs::write(sub.join(".git"), "gitdir: ../../../../elsewhere\n").unwrap();
        }
        let before = h.snapshot();
        let p = plan(&root, h.path()).unwrap();
        assert_eq!(p.state, "blocked", "{kind}: {}", p.reason);
        assert!(p.moves.is_empty(), "{kind}");
        assert!(
            p.reason.contains(".moai-cli/reports/sub") && p.reason.contains("repository"),
            "{kind}: {}",
            p.reason
        );
        let err = apply_plan(&root, h.path(), &p).unwrap_err();
        assert!(err.contains("blocked"), "{kind}: {err}");
        assert!(sub.join(".git").exists(), "{kind}: the repository moved");
        assert!(root.join(".moai-cli/reports/a.md").exists(), "{kind}");
        assert_eq!(h.snapshot(), before, "{kind}: the HOME stays untouched");
    }
}

/// Audit non-blocking: in a linked worktree a HOME that lies inside the primary checkout would
/// write the reports back into the repository, so it blocks too.
#[test]
fn a_linked_worktrees_home_inside_the_primary_checkout_blocks_the_move() {
    let tmp = tempfile::tempdir().unwrap();
    let primary = repo(tmp.path());
    let wt = tmp.path().join("a-feature-worktree");
    git_ok(
        &primary,
        &[
            "worktree",
            "add",
            "-q",
            wt.to_str().unwrap(),
            "-b",
            "feature",
        ],
    );
    let wt = fs::canonicalize(&wt).unwrap();
    write(&wt, ".moai-cli/reports/a.md", "a");
    let home = primary.join("in-primary-home");
    let p = plan(&wt, &home).unwrap();
    assert_eq!(p.state, "blocked", "{}", p.reason);
    assert!(p.reason.contains("inside the project"), "{}", p.reason);
    assert!(!home.exists());
    assert!(wt.join(".moai-cli/reports/a.md").exists());
}

/// `moai ARGS` in `h` with its HOME replaced by `home` and the working folder `cwd`; the parsed
/// envelope (success or error) and the exit status.
fn run_op_with(
    h: &TempHome,
    args: &[&str],
    stdin: &[u8],
    home: &std::ffi::OsStr,
    cwd: &Path,
) -> (bool, serde_json::Value) {
    use std::process::Stdio;
    let mut child = h
        .command(args)
        .env("HOME", home)
        .env("USERPROFILE", home)
        .current_dir(cwd)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, stdin);
    let out = child.wait_with_output().unwrap();
    let v: serde_json::Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|e| panic!("{e}: {}", String::from_utf8_lossy(&out.stderr)));
    (out.status.success(), v)
}

/// Audit follow-up: the same blocked apply reported `invalid_input` for a link but `internal`
/// for a nested repository and for a HOME problem. Every cause is one kind, the message keeps
/// the reason, and nothing moves. A relative HOME is tried with the working folder of the CLI
/// set to a temporary folder, so a broken guard cannot write the reports into this repository.
#[cfg(unix)]
#[test]
fn every_blocked_apply_reports_invalid_input_through_the_cli() {
    for cause in ["link", "repository", "home-in-project", "relative-home"] {
        let h = TempHome::new();
        let tmp = tempfile::tempdir().unwrap();
        let root = repo(tmp.path());
        let project = root.to_str().unwrap();
        let cwd = tmp.path().join("cwd");
        fs::create_dir_all(&cwd).unwrap();
        write(&root, ".moai-cli/reports/a.md", "a");
        let mut home = h.path().as_os_str().to_os_string();
        match cause {
            "link" => {
                let (other, _tracked) = other_repo_with_tracked_report(tmp.path());
                fs::remove_dir_all(root.join(".moai-cli")).unwrap();
                std::os::unix::fs::symlink(other.join(".moai-cli"), root.join(".moai-cli"))
                    .unwrap();
            }
            "repository" => write(&root, ".moai-cli/reports/sub/.git/HEAD", "ref: x\n"),
            "home-in-project" => home = root.join("in-repo-home").into_os_string(),
            _ => home = std::ffi::OsString::from("relative-home"),
        }
        let before = h.snapshot();
        let (ok, planned) = run_op_with(&h, &["reports-move-plan", project], b"", &home, &cwd);
        assert!(ok, "{cause}: {planned}");
        let p = planned["result"].clone();
        assert_eq!(p["state"], "blocked", "{cause}: {p}");
        let (ok, refused) = run_op_with(
            &h,
            &[
                "reports-move-apply",
                project,
                "--plan-digest",
                p["plan_digest"].as_str().unwrap(),
                "--permission-reference",
                PERMISSION,
            ],
            p.to_string().as_bytes(),
            &home,
            &cwd,
        );
        assert!(!ok, "{cause}: {refused}");
        assert_eq!(
            refused["error"]["kind"], "invalid_input",
            "{cause}: {refused}"
        );
        assert!(
            refused["error"]["message"]
                .as_str()
                .unwrap()
                .contains("blocked"),
            "{cause}: {refused}"
        );
        assert!(root.join(".moai-cli").exists(), "{cause}");
        assert!(!root.join("in-repo-home").exists(), "{cause}");
        assert!(
            !cwd.join("relative-home").exists(),
            "{cause}: a relative HOME was created in the working folder"
        );
        assert_eq!(
            h.snapshot(),
            before,
            "{cause}: the real HOME stays untouched"
        );
    }
}

/// Audit follow-up: on a case-insensitive file system (the default on macOS) a folder named
/// `.GIT` is the same folder as `.git`, so a repository below the report folder is found under
/// any letter case.
#[test]
fn a_git_folder_in_any_letter_case_below_the_report_folder_blocks_the_move() {
    for name in [".GIT", ".Git"] {
        let h = TempHome::new();
        let tmp = tempfile::tempdir().unwrap();
        let root = repo(tmp.path());
        write(&root, ".moai-cli/reports/a.md", "a");
        write(
            &root,
            &format!(".moai-cli/reports/sub/{name}/HEAD"),
            "ref: x\n",
        );
        let before = h.snapshot();
        let p = plan(&root, h.path()).unwrap();
        assert_eq!(p.state, "blocked", "{name}: {}", p.reason);
        assert!(p.moves.is_empty(), "{name}");
        assert!(
            p.reason.contains("repository of its own"),
            "{name}: {}",
            p.reason
        );
        let err = apply_plan(&root, h.path(), &p).unwrap_err();
        assert!(err.contains("blocked"), "{name}: {err}");
        assert!(
            root.join(".moai-cli/reports/sub").join(name).exists(),
            "{name}"
        );
        assert!(root.join(".moai-cli/reports/a.md").exists(), "{name}");
        assert_eq!(h.snapshot(), before, "{name}: the HOME stays untouched");
    }
}
