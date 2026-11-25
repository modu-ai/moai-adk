//! SPEC-20261121-project-state-move revision 5, REQ-11 and REQ-12: reports are never committed.
//! The repository tracks no file under a `reports` folder of either state folder (no exception),
//! and its `.gitignore` hides both folders. The scan runs on temporary git projects for its own
//! cases and on this repository for the real one; git runs through the shared fixture runner.
//! The folder names come from the library constants, so the test spells no state folder name.
mod common;

use common::git::{git_ok, git_raw};
use moai_cli::state_root::MOAI_DIR;
use moai_cli::PROJECT_STATE_DIR as EARLIER_DIR;
use std::fs;
use std::path::Path;

/// `<state folder>/reports` for both state folder names, earlier one first.
fn report_folders() -> [String; 2] {
    [
        format!("{EARLIER_DIR}/reports"),
        format!("{MOAI_DIR}/reports"),
    ]
}

/// The tracked paths under a report folder, sorted, or git's own complaint.
fn tracked_reports(root: &Path) -> Result<Vec<String>, String> {
    let folders = report_folders();
    let out = git_raw(
        root,
        &[
            "ls-files",
            "-z",
            "--",
            folders[0].as_str(),
            folders[1].as_str(),
        ],
    );
    if !out.success {
        return Err(out.stderr);
    }
    let mut paths: Vec<String> = out
        .stdout
        .split('\0')
        .filter(|p| folders.iter().any(|f| p.starts_with(&format!("{f}/"))))
        .map(String::from)
        .collect();
    paths.sort();
    Ok(paths)
}

fn project() -> tempfile::TempDir {
    let dir = tempfile::tempdir().unwrap();
    git_ok(dir.path(), &["init", "-q"]);
    dir
}

/// Writes the file and forces it into the index, the way a report gets past an ignore rule.
fn force_track(root: &Path, rel: &str) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(&path, "report\n").unwrap();
    git_ok(root, &["add", "-f", "--", rel]);
}

#[test]
fn the_repository_tracks_no_report_file() {
    let root = Path::new(env!("CARGO_MANIFEST_DIR"));
    if fs::symlink_metadata(root.join(".git")).is_err() {
        eprintln!("not a git work tree: the tracked-report scan has nothing to read");
        return;
    }
    let found = tracked_reports(root).expect("git ls-files");
    assert!(
        found.is_empty(),
        "reports are never committed (REQ-11); tracked: {:?} ({} in all)",
        &found[..found.len().min(5)],
        found.len()
    );
}

#[test]
fn the_scan_names_every_tracked_report_in_both_folders_without_an_exception() {
    let dir = project();
    let root = dir.path();
    let [earlier, current] = report_folders();
    let mut want = vec![
        format!("{earlier}/a.md"),
        format!("{earlier}/issue-bodies/1 two.md"),
        format!("{current}/b.md"),
        format!("{current}/plan-audit/.gitkeep"),
    ];
    for path in &want {
        force_track(root, path);
    }
    force_track(root, "docs/a.md");
    want.sort();
    assert_eq!(tracked_reports(root).unwrap(), want);
}

#[test]
fn the_scan_passes_a_project_with_no_report_file_under_the_state_folders() {
    let dir = project();
    let root = dir.path();
    force_track(root, &format!("{EARLIER_DIR}/manifests/a.json"));
    force_track(root, &format!("{MOAI_DIR}/specs/SPEC-1/spec.md"));
    // A folder that only looks alike is not a report folder of a state folder.
    force_track(root, "docs/reports/x.md");
    force_track(root, &format!("{EARLIER_DIR}/reports-index.md"));
    assert_eq!(tracked_reports(root).unwrap(), Vec::<String>::new());
}

#[test]
fn the_scan_reports_git_failing_instead_of_passing() {
    let dir = tempfile::tempdir().unwrap();
    let err = tracked_reports(dir.path()).unwrap_err();
    assert!(
        !err.is_empty(),
        "a folder that is no work tree must fail the scan"
    );
}

#[test]
fn the_gitignore_hides_every_file_under_both_report_folders() {
    let dir = project();
    let root = dir.path();
    let rules = fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join(".gitignore"))
        .expect("the repository .gitignore");
    fs::write(root.join(".gitignore"), rules).unwrap();
    let [earlier, current] = report_folders();
    for path in [
        format!("{earlier}/x.md"),
        format!("{earlier}/issue-bodies/1 two.md"),
        format!("{current}/x.md"),
    ] {
        let out = git_raw(root, &["check-ignore", "-v", "--no-index", "--", &path]);
        assert!(
            out.success && out.stdout.contains("reports"),
            "no .gitignore rule hides {path} (REQ-12); git said {:?} {:?}",
            out.stdout,
            out.stderr
        );
    }
}
