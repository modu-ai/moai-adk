//! #2835 (W2b, leader decisions D1–D2): `moai-cli init --new <name>` makes
//! `<HOME>/.moai/projects/<name>/` and initializes it through the launcher's first-launch path
//! (`git init`, the project settings with their marker, `AGENTS.md`); `init` with no `--new`
//! initializes the current folder the same way and still refuses HOME and the file-system root.
//! A taken name, a bad name and a symbolic-link parent are refused before any write.
//! Every run uses a temporary HOME; the real HOME is never read or written.
#![cfg(unix)]
mod common;

use serde_json::Value;
use std::fs;
use std::path::Path;
use std::process::{Command, Output};

fn run(home: &Path, cwd: &Path, args: &[&str]) -> Output {
    Command::new(crate::common::MOAI_BIN)
        .args(args)
        .current_dir(cwd)
        .env("HOME", home)
        .env("USERPROFILE", home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("GIT_AUTHOR_NAME", "t")
        .env("GIT_AUTHOR_EMAIL", "t@example.invalid")
        .env("GIT_COMMITTER_NAME", "t")
        .env("GIT_COMMITTER_EMAIL", "t@example.invalid")
        .output()
        .unwrap()
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "no envelope: {}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

fn temp_home() -> (tempfile::TempDir, std::path::PathBuf) {
    let dir = tempfile::tempdir().unwrap();
    let home = fs::canonicalize(dir.path()).unwrap().join("home");
    fs::create_dir_all(&home).unwrap();
    (dir, home)
}

/// Every file and folder under `root`, relative, for "nothing changed" checks.
fn listing(root: &Path) -> Vec<String> {
    let mut out = Vec::new();
    let mut stack = vec![root.to_path_buf()];
    while let Some(d) = stack.pop() {
        let Ok(rd) = fs::read_dir(&d) else { continue };
        for e in rd.flatten() {
            let p = e.path();
            out.push(p.strip_prefix(root).unwrap().display().to_string());
            if fs::symlink_metadata(&p).is_ok_and(|m| m.is_dir()) {
                stack.push(p);
            }
        }
    }
    out.sort();
    out
}

/// AC-1: `init --new demo` creates and initializes `<HOME>/.moai/projects/demo`.
#[test]
fn init_new_creates_the_project_under_the_home_projects_folder() {
    let (_d, home) = temp_home();
    let out = run(&home, &home, &["init", "--new", "demo"]);
    let v = envelope(&out);
    assert_eq!(v["ok"], true, "{v}");
    assert_eq!(v["operation"], "project.init");
    let dir = home.join(".moai/projects/demo");
    assert_eq!(v["result"]["project"], dir.to_str().unwrap(), "{v}");
    assert_eq!(v["result"]["created"], true, "{v}");
    assert!(dir.join(".git").is_dir(), "{v}");
    assert!(dir.join("AGENTS.md").is_file(), "{v}");
    assert!(dir.join(".moai/project.yaml").is_file(), "{v}");
    // Nothing else appeared in HOME besides the projects folder.
    assert_eq!(listing(&home.join(".moai")).first().unwrap(), "projects");
}

/// AC-1: the same name again is refused and the first project stays as it was.
#[test]
fn a_taken_project_name_is_refused() {
    let (_d, home) = temp_home();
    assert_eq!(
        envelope(&run(&home, &home, &["init", "--new", "demo"]))["ok"],
        true
    );
    let before = listing(&home);
    let v = envelope(&run(&home, &home, &["init", "--new", "demo"]));
    assert_eq!(v["ok"], false, "{v}");
    assert!(v.to_string().contains("already exists"), "{v}");
    assert_eq!(listing(&home), before);
}

/// AC-1: bad names are refused before anything is created.
#[test]
fn bad_project_names_are_refused() {
    let (_d, home) = temp_home();
    for bad in ["..", ".", "a/b", "", "-x", "a b", "../escape"] {
        // `--new=<name>` so a name led by `-` reaches the name check, not the flag parser.
        let arg = format!("--new={bad}");
        let v = envelope(&run(&home, &home, &["init", &arg]));
        assert_eq!(v["ok"], false, "{bad:?}: {v}");
        assert!(
            v.to_string().contains("invalid project name"),
            "{bad:?}: {v}"
        );
    }
    assert!(!home.join(".moai").exists(), "nothing created");
}

/// AC-1: a symbolic link as `~/.moai` or as `~/.moai/projects` is refused; the link's target
/// is not written.
#[test]
fn a_symlinked_parent_is_refused() {
    for link in [".moai", ".moai/projects"] {
        let (d, home) = temp_home();
        let elsewhere = d.path().join("elsewhere");
        fs::create_dir_all(&elsewhere).unwrap();
        let at = home.join(link);
        fs::create_dir_all(at.parent().unwrap()).unwrap();
        std::os::unix::fs::symlink(&elsewhere, &at).unwrap();
        let v = envelope(&run(&home, &home, &["init", "--new", "demo"]));
        assert_eq!(v["ok"], false, "{link}: {v}");
        assert!(v.to_string().contains("symbolic link"), "{link}: {v}");
        assert!(listing(&elsewhere).is_empty(), "{link}");
    }
}

/// `--dry-run` names the folder and creates nothing.
#[test]
fn dry_run_creates_nothing() {
    let (_d, home) = temp_home();
    let v = envelope(&run(&home, &home, &["init", "--new", "demo", "--dry-run"]));
    assert_eq!(v["ok"], true, "{v}");
    assert_eq!(v["result"]["created"], false, "{v}");
    assert_eq!(v["result"]["writes"], false, "{v}");
    assert!(
        v["result"]["note"].as_str().unwrap().contains("would"),
        "{v}"
    );
    assert!(!home.join(".moai").exists());
}

/// D1: `init` with no `--new` initializes the current folder like a first launch, and HOME
/// itself is still refused.
#[test]
fn plain_init_initializes_the_current_folder_and_refuses_home() {
    let (d, home) = temp_home();
    let here = d.path().join("work");
    fs::create_dir_all(&here).unwrap();
    let v = envelope(&run(&home, &here, &["init"]));
    assert_eq!(v["ok"], true, "{v}");
    assert!(
        here.join(".git").is_dir() && here.join("AGENTS.md").is_file(),
        "{v}"
    );
    let v = envelope(&run(&home, &home, &["init"]));
    assert_eq!(v["ok"], false, "{v}");
    assert!(v.to_string().contains("your HOME folder"), "{v}");
    assert!(!home.join(".git").exists());
}

/// #2846 audit B1: when initialization fails after the folder was made (no git on PATH), the
/// empty folder this run made is removed, so the retry the error message asks for works.
#[test]
fn a_failed_init_leaves_no_folder_and_the_retry_works() {
    let (_d, home) = temp_home();
    let out = Command::new(crate::common::MOAI_BIN)
        .args(["init", "--new", "demo"])
        .current_dir(&home)
        .env("HOME", &home)
        .env("USERPROFILE", &home)
        .env("PATH", "/nonexistent")
        .output()
        .unwrap();
    let v = envelope(&out);
    assert_eq!(v["ok"], false, "{v}");
    assert!(!home.join(".moai/projects/demo").exists(), "{v}");
    let v = envelope(&run(&home, &home, &["init", "--new", "demo"]));
    assert_eq!(v["ok"], true, "{v}");
    assert_eq!(v["result"]["created"], true, "{v}");
}

/// #2846 audit N2: Windows device names and over-long names are refused by the name check,
/// before `projects/` is created.
#[test]
fn device_and_overlong_names_are_refused() {
    let (_d, home) = temp_home();
    let long = "a".repeat(101);
    for bad in ["CON", "nul", "Com1", "LPT9", "aux.txt", long.as_str()] {
        let arg = format!("--new={bad}");
        let v = envelope(&run(&home, &home, &["init", &arg]));
        assert_eq!(v["ok"], false, "{bad}: {v}");
        assert!(v.to_string().contains("invalid project name"), "{bad}: {v}");
    }
    assert!(!home.join(".moai").exists());
    let ok = "a".repeat(100);
    let v = envelope(&run(&home, &home, &["init", "--new", &ok, "--dry-run"]));
    assert_eq!(v["ok"], true, "100 characters is allowed: {v}");
}
