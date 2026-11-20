//! SPEC-20261120-rust-quality-gate: `gate.yaml` over the built-in defaults, and the project
//! directory the gate runs in. The module under test is compiled into this test crate by path
//! (the `moai gate` word is wired into the binary in a later step).

#[path = "../src/quality_gate/mod.rs"]
#[allow(dead_code)]
mod quality_gate;

use quality_gate::config::{defaults, load, resolve_project_dir};
use std::fs;
use std::path::Path;
use std::time::Duration;
use tempfile::TempDir;

fn project_with(gate_yaml: Option<&str>) -> TempDir {
    let root = TempDir::new().unwrap();
    if let Some(text) = gate_yaml {
        let dir = root.path().join(".moai/config/sections");
        fs::create_dir_all(&dir).unwrap();
        fs::write(dir.join("gate.yaml"), text).unwrap();
    }
    root
}

/// AC-2 (REQ-2): the built-in defaults are the ones the Go build ships.
#[test]
fn defaults_match_the_go_build() {
    let d = defaults();
    assert!(d.enabled && !d.skip_tests && !d.pre_commit_enabled);
    assert_eq!(d.vet_timeout, Duration::from_secs(30));
    assert_eq!(d.lint_timeout, Duration::from_secs(60));
    assert_eq!(d.test_timeout, Duration::from_secs(120));
    assert_eq!(d.typecheck_timeout, Duration::from_secs(300));
    assert_eq!(d.lock_wait, Duration::from_secs(30));
    assert_eq!(d.total_timeout, Duration::ZERO);
    assert!(d.typecheck_enabled && d.typecheck_command.is_empty());
    assert!(d.ast_grep_enabled && d.graph_freshness_enabled);
    assert!(d.disabled_steps.is_empty());
}

/// AC-2 (REQ-2): an absent file is the defaults with nothing to warn about.
#[test]
fn an_absent_file_is_the_defaults() {
    let root = project_with(None);
    let (cfg, warnings) = load(root.path());
    assert_eq!(cfg.vet_timeout, defaults().vet_timeout);
    assert!(cfg.enabled);
    assert!(warnings.is_empty(), "{warnings:?}");
}

/// AC-2 (REQ-2): keys the file sets override; keys it omits keep the default (partial override).
#[test]
fn a_partial_file_overrides_only_what_it_sets() {
    let root = project_with(Some(
        "gate:\n  skip_tests: true\n  timeouts:\n    vet: 5\n    lock_wait: 2\n  pre_commit:\n    enabled: true\n",
    ));
    let (cfg, warnings) = load(root.path());
    assert!(warnings.is_empty(), "{warnings:?}");
    assert!(cfg.skip_tests && cfg.pre_commit_enabled && cfg.enabled);
    assert_eq!(cfg.vet_timeout, Duration::from_secs(5));
    assert_eq!(cfg.lock_wait, Duration::from_secs(2));
    assert_eq!(
        cfg.lint_timeout,
        Duration::from_secs(60),
        "omitted keys keep the default"
    );
    assert_eq!(cfg.test_timeout, Duration::from_secs(120));
}

/// AC-2 (REQ-2): a zero or negative timeout means the default; a total of zero or less means
/// "derive it".
#[test]
fn zero_and_negative_timeouts_fall_back() {
    let root = project_with(Some(
        "gate:\n  timeouts:\n    vet: 0\n    lint: -3\n    test: 7\n    typecheck: 0\n    total: -1\n    lock_wait: 0\n",
    ));
    let (cfg, _) = load(root.path());
    assert_eq!(cfg.vet_timeout, Duration::from_secs(30));
    assert_eq!(cfg.lint_timeout, Duration::from_secs(60));
    assert_eq!(cfg.test_timeout, Duration::from_secs(7));
    assert_eq!(cfg.typecheck_timeout, Duration::from_secs(300));
    assert_eq!(cfg.total_timeout, Duration::ZERO);
    assert_eq!(cfg.lock_wait, Duration::from_secs(30));
    let root = project_with(Some("gate:\n  timeouts:\n    total: 90\n"));
    assert_eq!(load(root.path()).0.total_timeout, Duration::from_secs(90));
}

/// AC-2 (REQ-2): `disabled_steps` keeps its verbatim values, spaced keys included; `false`
/// is the value that skips a step; YAML 1.1 words read as booleans the way Go reads them.
#[test]
fn disabled_steps_and_booleans_read_like_go() {
    let root = project_with(Some(
        "gate:\n  enabled: no\n  disabled_steps:\n    dotnet format: false\n    \"go vet\": true\n  typecheck:\n    enabled: off\n    command: tsc --noEmit\n",
    ));
    let (cfg, warnings) = load(root.path());
    assert!(warnings.is_empty(), "{warnings:?}");
    assert!(!cfg.enabled, "`no` is false");
    assert!(!cfg.typecheck_enabled);
    assert_eq!(cfg.typecheck_command, "tsc --noEmit");
    let get = |name: &str| {
        cfg.disabled_steps
            .iter()
            .find(|(k, _)| k == name)
            .map(|(_, v)| *v)
    };
    assert_eq!(get("dotnet format"), Some(false));
    assert_eq!(get("go vet"), Some(true));
}

/// AC-2 (REQ-3): a file outside the supported YAML subset is not guessed at: the defaults come
/// back with one warning that names the file.
#[test]
fn an_unreadable_file_warns_and_uses_the_defaults() {
    let root = project_with(Some("gate:\n  timeouts: {vet: 5}\n"));
    let (cfg, warnings) = load(root.path());
    assert_eq!(cfg.vet_timeout, Duration::from_secs(30));
    assert!(cfg.enabled);
    assert_eq!(warnings.len(), 1, "{warnings:?}");
    assert!(warnings[0].contains("gate.yaml"), "{warnings:?}");
}

/// AC-2 (REQ-2): the Go build tags come from the first non-comment line of
/// `.moai/config/build-tags`.
#[test]
fn go_build_tags_are_the_first_real_line() {
    let root = project_with(None);
    let dir = root.path().join(".moai/config");
    fs::create_dir_all(&dir).unwrap();
    fs::write(
        dir.join("build-tags"),
        "# tags\n\n  goolm sqlite_nocgo \nignored\n",
    )
    .unwrap();
    assert_eq!(load(root.path()).0.go_build_tags, "goolm sqlite_nocgo");
    let none = project_with(None);
    assert_eq!(load(none.path()).0.go_build_tags, "");
}

/// AC-2 (REQ-2): `CLAUDE_PROJECT_DIR` wins when set and not blank (made absolute); the working
/// directory otherwise.
#[test]
fn the_project_dir_follows_claude_project_dir_then_cwd() {
    let cwd = Path::new("/work/here");
    assert_eq!(resolve_project_dir(Some("/proj"), cwd), Path::new("/proj"));
    assert_eq!(
        resolve_project_dir(Some("  /proj  "), cwd),
        Path::new("/proj")
    );
    assert_eq!(
        resolve_project_dir(Some("sub"), cwd),
        Path::new("/work/here/sub")
    );
    assert_eq!(resolve_project_dir(Some("   "), cwd), cwd);
    assert_eq!(resolve_project_dir(Some(""), cwd), cwd);
    assert_eq!(resolve_project_dir(None, cwd), cwd);
}
