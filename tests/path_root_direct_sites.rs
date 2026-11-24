//! SPEC-20261121-path-root-moai-default slice 2: the sites that spelled the earlier `.moai-cli`
//! folder directly follow the resolvers. The Go settings clean-up writes, shows and probes the
//! Rust launcher under the folder the HOME resolves to (REQ-11, REQ-12), and project detection
//! skips `.moai` as it skips `.moai-cli` (REQ-13). Every case uses temporary folders only.
#![cfg(unix)]

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

// ---- SPEC-20261121-project-state-move REQ-10: no new direct spelling of the earlier folder ----

/// The earlier state folder's name and its constant, as product code spells them.
const SPELLINGS: [&str; 2] = ["PROJECT_STATE_DIR", ".moai-cli"];

/// The product source files that spell the earlier folder today, with how many lines each. A
/// file may spell it less often and still pass; a file that is not here, or spells it more
/// often, fails the scan. A spelling that is deliberate (a read of the earlier folder, the
/// resolver, the move) is added here by a card whose scope names this file.
const BASELINE: [(&str, usize); 17] = [
    ("agents.rs", 5),
    ("apply.rs", 1),
    ("bin_names.rs", 1),
    ("check_tree.rs", 1),
    ("forge/report.rs", 1),
    ("go_settings_clean.rs", 1),
    ("go_state.rs", 1),
    ("help.rs", 10),
    ("hook_guards.rs", 4),
    ("hooks.rs", 1),
    ("lib.rs", 19),
    ("main.rs", 1),
    ("ops/star_ask.rs", 1),
    ("project_detect.rs", 3),
    ("project_update.rs", 3),
    ("state_root.rs", 6),
    ("update.rs", 1),
];

/// Lines of product code that spell the earlier folder: everything before the first tests
/// module, comment lines skipped.
fn product_spellings(text: &str) -> usize {
    let mut n = 0;
    for line in text.lines() {
        let t = line.trim_start();
        if t.starts_with("mod tests")
            || t.starts_with("pub mod tests")
            || t.starts_with("mod test ")
        {
            break;
        }
        if t.starts_with("//") {
            continue;
        }
        if SPELLINGS.iter().any(|s| line.contains(s)) {
            n += 1;
        }
    }
    n
}

/// The files (name relative to `src/`, with `/`) that spell the earlier folder more often than
/// the baseline allows, with the count found.
fn over_baseline(files: &[(String, String)]) -> Vec<(String, usize)> {
    let mut over = Vec::new();
    for (name, text) in files {
        let found = product_spellings(text);
        let allowed = BASELINE
            .iter()
            .find(|(f, _)| f == name)
            .map_or(0, |(_, n)| *n);
        if found > allowed {
            over.push((name.clone(), found));
        }
    }
    over
}

fn rust_files(dir: &Path, base: &Path, out: &mut Vec<(String, String)>) {
    let mut entries: Vec<_> = fs::read_dir(dir)
        .unwrap()
        .map(|e| e.unwrap().path())
        .collect();
    entries.sort();
    for p in entries {
        if p.is_dir() {
            rust_files(&p, base, out);
        } else if p.extension().is_some_and(|e| e == "rs") {
            let name = p
                .strip_prefix(base)
                .unwrap()
                .to_string_lossy()
                .replace('\\', "/");
            out.push((name, fs::read_to_string(&p).unwrap()));
        }
    }
}

#[test]
fn the_repository_spells_the_earlier_folder_only_where_the_baseline_lists() {
    let src = Path::new(env!("CARGO_MANIFEST_DIR")).join("src");
    let mut files = Vec::new();
    rust_files(&src, &src, &mut files);
    assert!(
        files.len() > 50,
        "the scan found only {} files",
        files.len()
    );
    assert_eq!(over_baseline(&files), Vec::<(String, usize)>::new());
}

#[test]
fn a_new_spelling_outside_the_baseline_fails_the_scan_and_a_lower_count_passes() {
    let base = |name: &str, text: &str| (name.to_string(), text.to_string());
    // A file the baseline does not list.
    let added = [base("new_module.rs", "let d = \".moai-cli\";\n")];
    assert_eq!(over_baseline(&added), [("new_module.rs".to_string(), 1)]);
    // The constant counts too, and a spelling after the tests module or in a comment does not.
    let constant = [base("another.rs", "let d = PROJECT_STATE_DIR;\n")];
    assert_eq!(over_baseline(&constant).len(), 1);
    let quiet = [base(
        "another.rs",
        "// the .moai-cli folder\nfn f() {}\nmod tests {\n    const D: &str = \".moai-cli\";\n}\n",
    )];
    assert!(over_baseline(&quiet).is_empty());
    // A listed file over its count fails; at or under passes.
    let at = "let a = \".moai-cli\";\n".repeat(6);
    assert!(over_baseline(&[base("state_root.rs", &at)]).is_empty());
    let over = "let a = \".moai-cli\";\n".repeat(7);
    assert_eq!(
        over_baseline(&[base("state_root.rs", &over)]),
        [("state_root.rs".to_string(), 7)]
    );
    assert!(over_baseline(&[base("state_root.rs", "fn quiet() {}\n")]).is_empty());
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
