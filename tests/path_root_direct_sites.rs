//! SPEC-20261121-path-root-moai-default slice 2: the sites that spelled the earlier `.moai-cli`
//! folder directly follow the resolvers. The Go settings clean-up writes, shows and probes the
//! Rust launcher under the folder the HOME resolves to (REQ-11, REQ-12), and project detection
//! skips `.moai` as it skips `.moai-cli` (REQ-13). Every case uses temporary folders only.

use moai_cli::go_settings_clean::{
    launcher_shown, launcher_state, rust_status_line, LauncherState,
};
use moai_cli::project_detect::detect;
use std::fs;
use std::os::unix::fs::PermissionsExt as _;
use std::path::Path;

/// A launcher that runs but is not the Rust build: found, so never `missing`.
fn put_launcher(home: &Path, folder: &str) {
    let path = home.join(folder).join("path/moai-cli");
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(&path, "#!/bin/sh\necho other\n").unwrap();
    fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
}

/// A HOME that resolves to `.moai`: adopted the way a first launch adopts an empty HOME.
fn new_home() -> tempfile::TempDir {
    let home = tempfile::tempdir().unwrap();
    assert!(moai_cli::home_root::adopt_new(home.path(), "test").unwrap());
    assert_eq!(moai_cli::home_root::dir(home.path()), ".moai");
    home
}

/// A HOME that keeps the earlier `.moai-cli`.
fn earlier_home() -> tempfile::TempDir {
    let home = tempfile::tempdir().unwrap();
    fs::create_dir(home.path().join(".moai-cli")).unwrap();
    assert_eq!(moai_cli::home_root::dir(home.path()), ".moai-cli");
    home
}

#[test]
fn launcher_status_line_and_shown_path_name_the_folder_the_home_resolves_to() {
    let empty = tempfile::tempdir().unwrap();
    let cases = [
        (new_home(), "$HOME/.moai/path/moai-cli"),
        (earlier_home(), "$HOME/.moai-cli/path/moai-cli"),
        // A HOME with no state at all resolves to `.moai` too (the first launch adopts it).
        (empty, "$HOME/.moai/path/moai-cli"),
    ];
    for (home, shown) in cases {
        assert_eq!(launcher_shown(home.path()), shown);
        assert_eq!(
            rust_status_line(home.path())["command"],
            format!("\"{shown}\" statusline"),
            "{shown}"
        );
        assert_eq!(rust_status_line(home.path())["type"], "command");
    }
}

#[test]
fn launcher_is_looked_for_under_the_folder_the_home_resolves_to() {
    let project = tempfile::tempdir().unwrap();
    let root = project.path();

    let home = new_home();
    put_launcher(home.path(), ".moai");
    assert_ne!(launcher_state(root, home.path()), LauncherState::Missing);

    let home = new_home();
    put_launcher(home.path(), ".moai-cli");
    assert_eq!(launcher_state(root, home.path()), LauncherState::Missing);

    let home = earlier_home();
    put_launcher(home.path(), ".moai-cli");
    assert_ne!(launcher_state(root, home.path()), LauncherState::Missing);

    let home = earlier_home();
    put_launcher(home.path(), ".moai");
    assert_eq!(launcher_state(root, home.path()), LauncherState::Missing);
}

#[test]
fn detect_skips_the_state_folder_of_either_name() {
    for folder in [".moai", ".moai-cli"] {
        let project = tempfile::tempdir().unwrap();
        let dir = project.path().join(folder).join("tools");
        fs::create_dir_all(&dir).unwrap();
        fs::write(dir.join("helper.py"), "print(1)\n").unwrap();
        fs::write(dir.join("package.json"), "{}\n").unwrap();
        let d = detect(project.path()).unwrap();
        assert!(d.languages.is_empty(), "{folder}: {:?}", d.languages);
        assert!(d.manifests.is_empty(), "{folder}: {:?}", d.manifests);
        assert_eq!(d.project_kind, "new", "{folder}");

        // The same file beside the state folder counts.
        fs::write(project.path().join("helper.py"), "print(1)\n").unwrap();
        let d = detect(project.path()).unwrap();
        assert_eq!(d.languages, ["python"], "{folder}");
    }
}
