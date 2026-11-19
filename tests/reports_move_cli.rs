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
