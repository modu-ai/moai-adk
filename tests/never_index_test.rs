//! REQ-4 of SPEC-20261120-test-parallelism-cap: the folder creators mark the folders they
//! make with a `.metadata_never_index` file, so macOS Spotlight skips indexing them. This
//! file checks the temporary HOME of the test helpers; the lane build folder's own check
//! lives next to `lane_env::prepare` (src/lane_env.rs, `prepare_marks_the_build_folder_
//! never_index`). Best-effort by contract: the folders exist all the same, and these tests
//! check the marker only in the writable temporary folders.

mod common;

use common::home::TempHome;
use std::fs;
use std::path::Path;

fn has_marker(folder: &Path) -> bool {
    fs::symlink_metadata(folder.join(".metadata_never_index"))
        .map(|m| m.is_file())
        .unwrap_or(false)
}

#[test]
fn a_temp_home_carries_the_marker() {
    let home = TempHome::new();
    assert!(
        has_marker(home.path()),
        "{} has no .metadata_never_index",
        home.path().display()
    );
}

#[test]
fn a_project_folder_under_the_temp_home_is_in_a_marked_tree() {
    let home = TempHome::new();
    let p = home.project("never-index");
    assert!(
        p.starts_with(home.path()),
        "the project must stay inside the HOME"
    );
    assert!(has_marker(home.path()));
}
