//! SPEC-20261121-moai-name slice 1a: every reader of the `--version` first line accepts both
//! names, `moai <ver> (rust)` and `moai-cli <ver> (rust)`, before any binary prints the first
//! one. Three readers are covered here: the same-build check of `bin-names-plan`, the version
//! check of the release workflow (the real step script, run against stand-in binaries) and the
//! skill's fallback sentence. The install version read (`install::generation_of`) is pinned by its
//! own unit test. Every process runs in a temporary HOME or folder.

#![cfg(unix)]

mod common;

use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = common::MOAI_CLI_BIN;

fn repo_file(rel: &str) -> String {
    fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join(rel)).unwrap()
}

/// A temporary HOME with `~/.moai-cli/versions/a/bin/moai-cli`; returns (guard, home, exe).
fn layout() -> (TempDir, PathBuf, PathBuf) {
    let d = TempDir::new().unwrap();
    let home = fs::canonicalize(d.path()).unwrap();
    let bin = home.join(".moai-cli/versions/a/bin");
    fs::create_dir_all(&bin).unwrap();
    let exe = bin.join("moai-cli");
    common::copy_executable(Path::new(EXE), &exe);
    (d, home, exe)
}

fn plan(exe: &Path, home: &Path) -> Value {
    let out = Command::new(exe)
        .args(["bin-names-plan", "--home", home.to_str().unwrap()])
        .env("HOME", home)
        .env("PATH", "/usr/bin:/bin")
        .current_dir(home)
        .stdin(Stdio::null())
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)));
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

/// A separate `bin/moai` of this same build is moved aside for the link whichever of the two
/// names it prints; another build or another program is left alone.
#[test]
fn a_separate_moai_of_this_build_is_the_same_build_under_either_printed_name() {
    let sig = moai_cli::VERSION_SIGNATURE;
    for (printed, want) in [
        (format!("moai-cli {sig}"), "replace_copy"),
        (format!("moai {sig}"), "replace_copy"),
        ("moai 9.9.9 (rust)".to_string(), "blocked"),
        ("moai-cli 9.9.9 (rust)".to_string(), "blocked"),
        ("moai-adk 3.2.0".to_string(), "blocked"),
        (format!("other {sig}"), "blocked"),
    ] {
        let (_d, home, exe) = layout();
        let alias = exe.parent().unwrap().join("moai");
        common::write_executable(&alias, &format!("#!/bin/sh\necho '{printed}'\n"));
        let p = plan(&exe, &home);
        assert_eq!(p["state"], json!(want), "prints {printed:?}: {p}");
    }
}

/// The body of the `run: |` block of the workflow step with the given `id:`, dedented.
fn step_script(workflow: &str, id: &str) -> String {
    let lines: Vec<&str> = workflow.lines().collect();
    let id_line = format!("id: {id}");
    let at = lines
        .iter()
        .position(|l| l.trim() == id_line)
        .unwrap_or_else(|| panic!("no step `{id_line}`"));
    let run = at
        + lines[at..]
            .iter()
            .position(|l| l.trim() == "run: |")
            .expect("the step has a `run: |` block");
    let indent = lines[run].len() - lines[run].trim_start().len() + 2;
    let mut script = String::new();
    for l in &lines[run + 1..] {
        if l.trim().is_empty() {
            script.push('\n');
        } else if l.len() - l.trim_start().len() >= indent {
            script.push_str(&l[indent..]);
            script.push('\n');
        } else {
            break;
        }
    }
    script
}

const VERSION: &str = "4.2648.9";

/// Runs the release workflow's real "Binaries, tag and release note agree" script in a folder
/// whose `target/release/` holds two stand-in binaries that print the given first lines. Returns
/// (exit ok, stderr, the step summary the script wrote).
fn run_release_check(first_of_cli: &str, first_of_moai: &str) -> (bool, String, String) {
    let script = step_script(&repo_file(".github/workflows/release.yml"), "verify");
    let d = TempDir::new().unwrap();
    let root = fs::canonicalize(d.path()).unwrap();
    let release = root.join("target/release");
    fs::create_dir_all(&release).unwrap();
    for (name, first) in [("moai-cli", first_of_cli), ("moai", first_of_moai)] {
        common::write_executable(
            &release.join(name),
            &format!(
                "#!/bin/sh\nprintf '%s\\n%s\\n' '{first}' 'CI build 9 (MOAI_BUILD_NUMBER), built 2026-11-23'\n"
            ),
        );
    }
    let summary = root.join("summary.md");
    let out = Command::new("bash")
        .args(["-eo", "pipefail", "-c", &script])
        .current_dir(&root)
        .env("VERSION", VERSION)
        .env("BUILD", "9")
        .env("TAG", format!("moai-adk-v{VERSION}"))
        .env("SHA", "0".repeat(40))
        .env("PRERELEASE", "false")
        .env("GITHUB_STEP_SUMMARY", &summary)
        .stdin(Stdio::null())
        .output()
        .unwrap();
    (
        out.status.success(),
        String::from_utf8_lossy(&out.stderr).into_owned(),
        fs::read_to_string(&summary).unwrap_or_default(),
    )
}

/// The release check passes for a build printing either name and fails for another version, the
/// Go lines or another name; the note it writes says what the binaries printed.
#[test]
fn the_release_check_accepts_both_printed_names_and_nothing_else() {
    let old = format!("moai-cli {VERSION} (rust)");
    let new = format!("moai {VERSION} (rust)");
    for (cli, moai) in [(&old, &old), (&new, &new), (&old, &new), (&new, &old)] {
        let (ok, err, summary) = run_release_check(cli, moai);
        assert!(ok, "cli prints {cli:?}, moai prints {moai:?}: {err}");
        assert!(
            summary.contains(&format!("- --version: {moai};")),
            "the note names what the binary printed: {summary}"
        );
    }
    for bad in [
        "moai 1.0.0 (rust)".to_string(),
        "moai-adk 3.2.0".to_string(),
        format!("other {VERSION} (rust)"),
        format!("moai {VERSION}"),
    ] {
        let (ok, err, _) = run_release_check(&bad, &bad);
        assert!(!ok, "{bad:?} must fail the release check");
        assert!(err.contains("expected"), "{bad:?}: {err}");
    }
}

/// The skill's fallback sentence verifies the signature naming `moai`, still takes the earlier
/// printed name, and the path it falls back to is the earlier binary name.
#[test]
fn the_skill_fallback_sentence_names_both_signatures() {
    let skill = repo_file("skills/moai/SKILL.md");
    let line = skill
        .lines()
        .find(|l| l.contains("run `moai --version`"))
        .expect("the skill has the version check sentence");
    assert!(line.contains("`moai <version> (rust)`"), "{line}");
    assert!(line.contains("`moai-cli <version> (rust)`"), "{line}");
    assert!(line.contains("use `moai-cli` for every command"), "{line}");
    let first_new = line.find("`moai <version> (rust)`").unwrap();
    let first_old = line.find("`moai-cli <version> (rust)`").unwrap();
    assert!(first_new < first_old, "the new name comes first: {line}");
}
