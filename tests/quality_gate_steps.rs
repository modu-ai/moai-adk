//! SPEC-20261120-rust-quality-gate: language detection and step execution.
//! Test names start with `detection_` or `steps_` so the SPEC's verify filters pick them.

#[path = "../src/quality_gate/mod.rs"]
#[allow(dead_code)]
mod quality_gate;

use quality_gate::detect::{detect, Detected};
use quality_gate::{run, Invocation};
use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};
use tempfile::TempDir;

fn touch(root: &Path, rel: &str) {
    let p = root.join(rel);
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::write(p, "x").unwrap();
}

/// A Go module with one source file: the Go steps skip a module that holds no `.go` file.
fn go_project(root: &Path) {
    touch(root, "go.mod");
    touch(root, "main.go");
}

fn roots(found: &[Detected], base: &Path) -> Vec<String> {
    found
        .iter()
        .map(|d| d.root.strip_prefix(base).unwrap().display().to_string())
        .collect()
}

/// AC-3 (REQ-4): one root marker -> one entry in the root.
#[test]
fn detection_root_marker() {
    let p = TempDir::new().unwrap();
    touch(p.path(), "go.mod");
    let found = detect(p.path());
    assert_eq!(roots(&found, p.path()), vec![""]);
}

/// AC-3 (REQ-4): a glob marker (`*.csproj`, `*.cabal`) matches by file count.
#[test]
fn detection_glob_marker() {
    let p = TempDir::new().unwrap();
    touch(p.path(), "App.csproj");
    assert_eq!(detect(p.path()).len(), 1);
    let q = TempDir::new().unwrap();
    touch(q.path(), "demo.cabal");
    assert_eq!(detect(q.path()).len(), 1);
    let none = TempDir::new().unwrap();
    assert!(detect(none.path()).is_empty());
}

/// AC-3 (REQ-4): the root match comes first, nested modules of other languages follow, and a
/// language found at the root is not repeated below it.
#[test]
fn detection_monorepo_root_first_one_entry_per_language() {
    let p = TempDir::new().unwrap();
    touch(p.path(), "package.json");
    touch(p.path(), "svc/go.mod");
    touch(p.path(), "svc2/go.mod");
    touch(p.path(), "web/package.json");
    let found = detect(p.path());
    let r = roots(&found, p.path());
    assert_eq!(r[0], "", "root first: {r:?}");
    assert!(r.contains(&"svc".to_string()), "{r:?}");
    assert!(
        !r.contains(&"svc2".to_string()),
        "one entry per language: {r:?}"
    );
    assert!(
        !r.contains(&"web".to_string()),
        "node already found at the root: {r:?}"
    );
}

/// AC-3 (REQ-4): dependency and build folders are never walked; a module root is not descended.
#[test]
fn detection_skips_dependency_folders_and_module_interiors() {
    let p = TempDir::new().unwrap();
    touch(p.path(), "node_modules/dep/go.mod");
    touch(p.path(), "target/x/Cargo.toml");
    touch(p.path(), "vendor/y/go.mod");
    touch(p.path(), ".git/z/go.mod");
    touch(p.path(), "mod/Cargo.toml");
    touch(p.path(), "mod/inner/go.mod");
    let r = roots(&detect(p.path()), p.path());
    assert_eq!(r, vec!["mod"], "{r:?}");
}

/// AC-3 (REQ-4): the walk reaches four levels below the root and no further.
#[test]
fn detection_depth_limit_is_four() {
    let p = TempDir::new().unwrap();
    touch(p.path(), "a/b/c/d/go.mod");
    assert_eq!(roots(&detect(p.path()), p.path()), vec!["a/b/c/d"]);
    let q = TempDir::new().unwrap();
    touch(q.path(), "a/b/c/d/e/go.mod");
    assert!(detect(q.path()).is_empty());
}

/// AC-3 (REQ-4): the table order decides a root with two markers (Go before Node).
#[test]
fn detection_table_order_at_the_root() {
    let p = TempDir::new().unwrap();
    touch(p.path(), "go.mod");
    touch(p.path(), "package.json");
    let found = detect(p.path());
    assert_eq!(found.len(), 1, "the root keeps the first match only");
    let q = TempDir::new().unwrap();
    touch(q.path(), "go.mod");
    touch(q.path(), "sub/package.json");
    assert_eq!(detect(q.path()).len(), 2);
}

// ---- step execution -------------------------------------------------------------------------

/// A fake toolchain directory: each `(name, script body)` becomes an executable on PATH.
fn bin_dir(tools: &[(&str, &str)]) -> TempDir {
    let d = TempDir::new().unwrap();
    for (name, body) in tools {
        let p = d.path().join(name);
        fs::write(&p, format!("#!/bin/sh\n{body}\n")).unwrap();
        fs::set_permissions(&p, fs::Permissions::from_mode(0o755)).unwrap();
    }
    d
}

fn path_of(bin: &TempDir) -> std::ffi::OsString {
    // The tool directory first, then a minimal system PATH so `sh` and `sleep` resolve.
    let mut v: Vec<PathBuf> = vec![bin.path().to_path_buf()];
    v.push(PathBuf::from("/usr/bin"));
    v.push(PathBuf::from("/bin"));
    std::env::join_paths(v).unwrap()
}

fn gate(project: &Path, bin: &TempDir) -> (i32, String) {
    let inv = Invocation {
        project_dir: project.to_path_buf(),
        precommit_marker: false,
        path: path_of(bin),
    };
    let mut err = Vec::new();
    let code = run(&inv, &mut err);
    (code, String::from_utf8_lossy(&err).into_owned())
}

fn config(project: &Path, text: &str) {
    let d = project.join(".moai/config/sections");
    fs::create_dir_all(&d).unwrap();
    fs::write(d.join("gate.yaml"), text).unwrap();
}

/// AC-4 (REQ-6): go vet failing stops the run at that step with exit 1 and names the step.
#[test]
fn steps_a_failing_vet_fails_the_gate() {
    let p = TempDir::new().unwrap();
    go_project(p.path());
    let bin = bin_dir(&[("go", "echo vet-broke >&2; exit 3")]);
    let (code, err) = gate(p.path(), &bin);
    assert_eq!(code, 1, "{err}");
    assert!(err.contains("quality gate failed: go vet"), "{err}");
    assert!(err.contains("vet-broke"), "step output is shown: {err}");
}

/// AC-4 (REQ-5, REQ-8): an optional tool that is missing is skipped with its reason in the
/// summary and does not fail the run; a passing required test step passes it.
#[test]
fn steps_a_missing_optional_tool_is_skipped_and_named() {
    let p = TempDir::new().unwrap();
    go_project(p.path());
    touch(p.path(), ".golangci.yml");
    // `go` passes; golangci-lint is absent from PATH (optional).
    let bin = bin_dir(&[("go", "exit 0")]);
    let (code, err) = gate(p.path(), &bin);
    assert_eq!(code, 0, "{err}");
    assert!(err.contains("quality gate steps ("), "{err}");
    assert!(err.contains("golangci-lint: skipped"), "{err}");
    assert!(
        err.contains("the optional tool \"golangci-lint\" was not found on PATH"),
        "{err}"
    );
    assert!(err.contains("go vet: executed in "), "{err}");
}

/// AC-4 (REQ-5): a step whose config files are all absent is skipped with the list of names.
#[test]
fn steps_missing_config_file_skips_the_linter() {
    let p = TempDir::new().unwrap();
    go_project(p.path());
    let bin = bin_dir(&[("go", "exit 0"), ("golangci-lint", "exit 9")]);
    let (code, err) = gate(p.path(), &bin);
    assert_eq!(code, 0, "{err}");
    assert!(
        err.contains("none of its config files exist in the project directory ("),
        "{err}"
    );
}

/// AC-4 (REQ-5): `disabled_steps: <name>: false` turns one step off and names the setting.
#[test]
fn steps_disabled_by_config() {
    let p = TempDir::new().unwrap();
    go_project(p.path());
    config(p.path(), "gate:\n  disabled_steps:\n    go vet: false\n");
    let bin = bin_dir(&[("go", "case \"$1\" in vet) exit 5;; esac; exit 0")]);
    let (code, err) = gate(p.path(), &bin);
    assert_eq!(code, 0, "{err}");
    assert!(
        err.contains("go vet: disabled — turned off by the gate.disabled_steps setting"),
        "{err}"
    );
}

/// AC-4 (REQ-5): `skip_tests` skips the test step with its own reason.
#[test]
fn steps_skip_tests_setting() {
    let p = TempDir::new().unwrap();
    go_project(p.path());
    config(p.path(), "gate:\n  skip_tests: true\n");
    let bin = bin_dir(&[("go", "case \"$1\" in test) exit 5;; esac; exit 0")]);
    let (code, err) = gate(p.path(), &bin);
    assert_eq!(code, 0, "{err}");
    assert!(
        err.contains("turned off by the gate.skip_tests setting"),
        "{err}"
    );
}

/// AC-4 (REQ-6): pytest exit 5 (no tests collected) passes; any other non-zero fails.
#[test]
fn steps_pytest_exit_five_passes() {
    let p = TempDir::new().unwrap();
    touch(p.path(), "pyproject.toml");
    let bin = bin_dir(&[("pytest", "exit 5")]);
    let (code, err) = gate(p.path(), &bin);
    assert_eq!(code, 0, "{err}");
    let bin2 = bin_dir(&[("pytest", "exit 1")]);
    let (code2, err2) = gate(p.path(), &bin2);
    assert_eq!(code2, 1, "{err2}");
    assert!(err2.contains("quality gate failed: pytest"), "{err2}");
}

/// AC-4 (REQ-7): a step that outlives its limit fails the run with a timeout message and the
/// whole process group is ended.
#[test]
fn steps_timeout_kills_the_group() {
    let p = TempDir::new().unwrap();
    go_project(p.path());
    config(p.path(), "gate:\n  timeouts:\n    vet: 1\n");
    let marker = p.path().join("grandchild-survived");
    let body = format!("(sleep 3; touch '{}') &\nsleep 30", marker.display());
    let bin = bin_dir(&[("go", &body)]);
    let started = std::time::Instant::now();
    let (code, err) = gate(p.path(), &bin);
    assert_eq!(code, 1, "{err}");
    assert!(err.contains("quality gate timed out: go vet exceeded 1s; termination was signalled to the step's whole process group"), "{err}");
    assert!(
        started.elapsed().as_secs() < 20,
        "the gate must not wait for the sleeper"
    );
    std::thread::sleep(std::time::Duration::from_secs(4));
    assert!(!marker.exists(), "the grandchild was left running");
}

/// AC-4 (REQ-11): steps run in the module's own directory, not in the project root.
#[test]
fn steps_run_in_the_module_root() {
    let p = TempDir::new().unwrap();
    touch(p.path(), "pkg/go.mod");
    touch(p.path(), "pkg/main.go");
    let out = p.path().join("cwd.txt");
    let bin = bin_dir(&[(
        "go",
        &format!(
            "case \"$1\" in vet) pwd > '{}';; esac; exit 0",
            out.display()
        ),
    )]);
    let (code, err) = gate(p.path(), &bin);
    assert_eq!(code, 0, "{err}");
    let seen = fs::read_to_string(&out).unwrap();
    assert!(seen.trim().ends_with("/pkg"), "{seen}");
}

/// AC-4 (REQ-11): the git repository-location variables never reach a step.
#[test]
fn steps_git_location_variables_are_removed() {
    let p = TempDir::new().unwrap();
    go_project(p.path());
    let out = p.path().join("env.txt");
    let bin = bin_dir(&[(
        "go",
        &format!(
            "case \"$1\" in vet) env | grep -c '^GIT_DIR=' > '{}';; esac; exit 0",
            out.display()
        ),
    )]);
    std::env::set_var("GIT_DIR", "/nonexistent-git-dir");
    let (code, err) = gate(p.path(), &bin);
    std::env::remove_var("GIT_DIR");
    assert_eq!(code, 0, "{err}");
    assert_eq!(fs::read_to_string(&out).unwrap().trim(), "0");
}

/// AC-4 (REQ-5): a project with no recognized marker passes with a one-line note, as Go does
/// (nothing to run is not a failure, and the summary says no step was configured).
#[test]
fn steps_no_toolchain_passes() {
    let p = TempDir::new().unwrap();
    let bin = bin_dir(&[]);
    let (code, err) = gate(p.path(), &bin);
    assert_eq!(code, 0, "{err}");
}
