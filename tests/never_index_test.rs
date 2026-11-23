//! REQ-4 of SPEC-20261120-test-parallelism-cap: the test helpers mark the temporary HOME they
//! make with a `.metadata_never_index` file, so macOS Spotlight skips indexing it. This file
//! checks that marker. The lane build folder's half of the requirement is retired (decision D5
//! of that SPEC): `moai lane-env` makes no folder under the home, so there is none to mark, and
//! `tests/lane_env_cli.rs` `lane_env_makes_nothing_under_home` pins that it writes nothing,
//! the marker included. Best-effort by contract: the folders exist all the same, and these
//! tests check the marker only in the writable temporary folders.

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
