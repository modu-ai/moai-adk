//! SPEC-20261120-rust-quality-gate AC-6: the same fixture projects give the same exit code and
//! the same step rows and failure text as the Go build.
//!
//! The Go build is run once, on a developer machine, to record the table
//! (`.moai/specs/SPEC-20261120-rust-quality-gate/parity-recorded.txt`); CI compares the Rust
//! gate with the recorded table and does not need the Go build.
//!
//! Re-record:  `MOAI_GO_BIN=<go build> cargo test --test quality_gate_parity -- --ignored record`
//! Live check: `MOAI_GO_BIN=<go build> cargo test --test quality_gate_parity live`
//!
//! What the comparison leaves out (it is named, not hidden): the Go build's own log lines, its
//! graph-freshness and ast-grep lines and its closing error box (axes and decoration this build
//! does not print), blank lines, and the clock fields of an executed row.

#![cfg(unix)] // the gate runs process groups and flock; a Windows port is not part of this SPEC

#[path = "../src/quality_gate/mod.rs"]
#[allow(dead_code)]
mod quality_gate;

mod common;

use common::git::git_ok;
use common::CleanEnv;
use quality_gate::{run, Invocation};
use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::path::Path;
use std::process::Command;
use tempfile::TempDir;

struct Scenario {
    name: &'static str,
    files: &'static [(&'static str, &'static str)],
    /// Executables on PATH: `(name, sh body)`.
    tools: &'static [(&'static str, &'static str)],
    gate_yaml: Option<&'static str>,
}

const OK: &str = "exit 0";
const GO_MAIN: (&str, &str) = ("main.go", "package x\n");
const GO_MOD: (&str, &str) = ("go.mod", "module x\n");

const SCENARIOS: &[Scenario] = &[
    Scenario {
        name: "go_pass",
        files: &[GO_MOD, GO_MAIN],
        tools: &[("go", OK)],
        gate_yaml: None,
    },
    Scenario {
        name: "go_vet_fail",
        files: &[GO_MOD, GO_MAIN],
        tools: &[("go", "echo vet-broke >&2; exit 3")],
        gate_yaml: None,
    },
    Scenario {
        name: "go_vet_fail_no_output",
        files: &[GO_MOD, GO_MAIN],
        tools: &[("go", "exit 7")],
        gate_yaml: None,
    },
    Scenario {
        name: "go_test_fail",
        files: &[GO_MOD, GO_MAIN],
        tools: &[(
            "go",
            "case \"$1\" in test) echo FAIL >&2; exit 1;; esac; exit 0",
        )],
        gate_yaml: None,
    },
    Scenario {
        name: "go_lint_runs_with_config",
        files: &[GO_MOD, GO_MAIN, (".golangci.yml", "x: 1\n")],
        tools: &[("go", OK), ("golangci-lint", "echo lint-found >&2; exit 1")],
        gate_yaml: None,
    },
    Scenario {
        name: "go_lint_tool_missing",
        files: &[GO_MOD, GO_MAIN, (".golangci.yml", "x: 1\n")],
        tools: &[("go", OK)],
        gate_yaml: None,
    },
    Scenario {
        name: "go_no_sources",
        files: &[GO_MOD],
        tools: &[("go", "exit 9")],
        gate_yaml: None,
    },
    Scenario {
        name: "go_vet_disabled",
        files: &[GO_MOD, GO_MAIN],
        tools: &[("go", "case \"$1\" in vet) exit 5;; esac; exit 0")],
        gate_yaml: Some("gate:\n  disabled_steps:\n    go vet: false\n"),
    },
    Scenario {
        name: "go_skip_tests",
        files: &[GO_MOD, GO_MAIN],
        tools: &[("go", "case \"$1\" in test) exit 5;; esac; exit 0")],
        gate_yaml: Some("gate:\n  skip_tests: true\n"),
    },
    Scenario {
        name: "go_build_tags",
        files: &[
            GO_MOD,
            GO_MAIN,
            (".moai/config/build-tags", "# t\nfoo bar\n"),
        ],
        tools: &[("go", OK)],
        gate_yaml: None,
    },
    Scenario {
        name: "go_typecheck_override_runs",
        files: &[GO_MOD, GO_MAIN],
        tools: &[("go", OK), ("mytc", "exit 0")],
        gate_yaml: Some("gate:\n  typecheck:\n    command: mytc --strict\n"),
    },
    Scenario {
        name: "go_typecheck_override_fails",
        files: &[GO_MOD, GO_MAIN],
        tools: &[("go", OK), ("mytc", "echo type-error >&2; exit 2")],
        gate_yaml: Some("gate:\n  typecheck:\n    command: mytc\n"),
    },
    Scenario {
        name: "go_typecheck_override_tool_missing",
        files: &[GO_MOD, GO_MAIN],
        tools: &[("go", OK)],
        gate_yaml: Some("gate:\n  typecheck:\n    command: mytc\n"),
    },
    Scenario {
        name: "go_typecheck_off",
        files: &[GO_MOD, GO_MAIN],
        tools: &[("go", OK)],
        gate_yaml: Some("gate:\n  typecheck:\n    enabled: false\n"),
    },
    Scenario {
        name: "go_vet_timeout",
        files: &[GO_MOD, GO_MAIN],
        tools: &[("go", "sleep 30")],
        gate_yaml: Some("gate:\n  timeouts:\n    vet: 1\n"),
    },
    Scenario {
        name: "python_without_sources",
        files: &[("pyproject.toml", "[project]\nname='x'\n")],
        tools: &[("ruff", "exit 0"), ("mypy", "exit 2"), ("pytest", "exit 5")],
        gate_yaml: None,
    },
    Scenario {
        name: "python_pytest_exit_five_passes",
        files: &[("pyproject.toml", "[project]\n"), ("a.py", "x = 1\n")],
        tools: &[("pytest", "exit 5")],
        gate_yaml: None,
    },
    Scenario {
        name: "python_pytest_fails",
        files: &[("pyproject.toml", "[project]\n"), ("a.py", "x = 1\n")],
        tools: &[("pytest", "echo 1 failed >&2; exit 1")],
        gate_yaml: None,
    },
    Scenario {
        name: "python_uv_runner",
        files: &[
            ("pyproject.toml", "[project]\n"),
            ("uv.lock", ""),
            ("a.py", ""),
        ],
        tools: &[("uv", OK)],
        gate_yaml: None,
    },
    Scenario {
        name: "rust_pass",
        files: &[("Cargo.toml", "[package]\n")],
        tools: &[("cargo", OK)],
        gate_yaml: None,
    },
    Scenario {
        name: "rust_clippy_fails",
        files: &[("Cargo.toml", "[package]\n")],
        tools: &[(
            "cargo",
            "case \"$1\" in clippy) echo warn >&2; exit 101;; esac; exit 0",
        )],
        gate_yaml: None,
    },
    Scenario {
        // `mix test` is a required step and `mix` is not in the system directories of PATH.
        name: "elixir_required_tool_missing",
        files: &[("mix.exs", "# mix\n")],
        tools: &[],
        gate_yaml: None,
    },
    Scenario {
        name: "dotnet_restore_failure_passes",
        files: &[("App.csproj", "<Project/>\n"), ("a.cs", "")],
        tools: &[("dotnet", "echo 'error NU1202: not compatible' >&2; exit 1")],
        gate_yaml: None,
    },
    Scenario {
        name: "dotnet_other_failure",
        files: &[("App.csproj", "<Project/>\n"), ("a.cs", "")],
        tools: &[("dotnet", "echo boom >&2; exit 1")],
        gate_yaml: None,
    },
    Scenario {
        name: "flutter_project",
        files: &[(
            "pubspec.yaml",
            "dependencies:\n  flutter:\n    sdk: flutter\n",
        )],
        tools: &[("flutter", OK)],
        gate_yaml: None,
    },
    Scenario {
        name: "dart_project",
        files: &[("pubspec.yaml", "name: x\n")],
        tools: &[("dart", OK)],
        gate_yaml: None,
    },
    Scenario {
        name: "monorepo_go_and_rust",
        files: &[GO_MOD, GO_MAIN, ("svc/Cargo.toml", "[package]\n")],
        tools: &[("go", OK), ("cargo", OK)],
        gate_yaml: None,
    },
    Scenario {
        name: "go_second_step_not_reached_after_lint_fail",
        files: &[GO_MOD, GO_MAIN, (".golangci.yaml", "x: 1\n")],
        tools: &[("go", OK), ("golangci-lint", "exit 1")],
        gate_yaml: None,
    },
    Scenario {
        name: "no_language",
        files: &[("README.md", "x\n")],
        tools: &[],
        gate_yaml: None,
    },
    Scenario {
        name: "gate_disabled",
        files: &[GO_MOD, GO_MAIN],
        tools: &[("go", "exit 3")],
        gate_yaml: Some("gate:\n  enabled: false\n"),
    },
];

fn build(s: &Scenario) -> (TempDir, TempDir) {
    let project = TempDir::new().unwrap();
    for (rel, body) in s.files {
        let p = project.path().join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }
    if let Some(y) = s.gate_yaml {
        let d = project.path().join(".moai/config/sections");
        fs::create_dir_all(&d).unwrap();
        fs::write(d.join("gate.yaml"), y).unwrap();
    }
    let bin = TempDir::new().unwrap();
    for (name, body) in s.tools {
        let p = bin.path().join(name);
        fs::write(&p, format!("#!/bin/sh\n{body}\n")).unwrap();
        fs::set_permissions(&p, fs::Permissions::from_mode(0o755)).unwrap();
    }
    (project, bin)
}

fn path_for(bin: &Path) -> String {
    format!("{}:/usr/bin:/bin", bin.display())
}

/// Keeps what the two builds have in common (see the module comment for what is left out).
fn normalize(raw: &str, project: &Path) -> String {
    let canon = project
        .canonicalize()
        .unwrap_or_else(|_| project.to_path_buf());
    let mut rows = Vec::new();
    let mut in_box = false;
    for line in raw.lines() {
        let line = line
            .replace(canon.to_string_lossy().as_ref(), "<P>")
            .replace(project.to_string_lossy().as_ref(), "<P>");
        let t = line.trim_end();
        if t.is_empty() || t.starts_with("time=") {
            continue;
        }
        if t.starts_with("graph-freshness:")
            || t.starts_with("ast-grep scan skipped")
            || t.starts_with("  - graph-freshness: not run")
            || t.starts_with("  - ast-grep: not run")
            // The header of a summary that holds only those two lines (no language detected).
            || t == "quality gate steps (0 configured):"
        {
            continue;
        }
        // The Go build ends a failed run with a decorated error box.
        if t.trim() == "ERROR" {
            in_box = true;
        }
        if in_box {
            continue;
        }
        rows.push(clock_free(t));
    }
    rows.join("\n")
}

/// `executed in 122ms at 2026-11-20T09:10:20.626+09:00 — cmd` -> `executed — cmd`.
fn clock_free(line: &str) -> String {
    match (line.find("executed in "), line.find(" — ")) {
        (Some(a), Some(b)) if line.starts_with("  - ") && a < b => {
            format!("{}executed{}", &line[..a], &line[b..])
        }
        (Some(a), None) if line.starts_with("  - ") => format!("{}executed", &line[..a]),
        _ => line.to_string(),
    }
}

fn rust_result(s: &Scenario) -> String {
    let (project, bin) = build(s);
    let inv = Invocation {
        project_dir: project.path().canonicalize().unwrap(),
        precommit_marker: false,
        path: path_for(bin.path()).into(),
    };
    let mut err = Vec::new();
    let code = run(&inv, &mut err);
    format!(
        "exit: {code}\n{}",
        normalize(&String::from_utf8_lossy(&err), project.path())
    )
    .trim_end()
    .to_string()
}

fn go_result(go_bin: &str, s: &Scenario) -> String {
    let (project, bin) = build(s);
    let home = TempDir::new().unwrap();
    let out = Command::new(go_bin)
        .arg("gate")
        .current_dir(project.path().canonicalize().unwrap())
        .clean_env()
        .env("HOME", home.path())
        .env("PATH", path_for(bin.path()))
        .output()
        .expect("run the Go build");
    format!(
        "exit: {}\n{}",
        out.status.code().unwrap_or(-1),
        normalize(&String::from_utf8_lossy(&out.stderr), project.path())
    )
    .trim_end()
    .to_string()
}

fn recorded_path() -> std::path::PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR"))
        .join(".moai/specs/SPEC-20261120-rust-quality-gate/parity-recorded.txt")
}

fn parse_recorded(text: &str) -> Vec<(String, String)> {
    let mut out: Vec<(String, String)> = Vec::new();
    for block in text.split("=== ").skip(1) {
        let (name, body) = block.split_once('\n').unwrap_or((block, ""));
        out.push((name.trim().to_string(), body.trim_end().to_string()));
    }
    out
}

/// AC-6: every scenario of the recorded table matches the Rust gate.
#[test]
fn parity_with_the_recorded_go_table() {
    let text =
        fs::read_to_string(recorded_path()).expect("recorded table is committed with the SPEC");
    let recorded = parse_recorded(&text);
    assert_eq!(
        recorded.len(),
        SCENARIOS.len(),
        "recorded table and scenario list differ in size"
    );
    let mut mismatches = Vec::new();
    for s in SCENARIOS {
        let want = recorded
            .iter()
            .find(|(n, _)| n == s.name)
            .unwrap_or_else(|| panic!("scenario {} is not recorded", s.name));
        let got = rust_result(s);
        if got != want.1 {
            mismatches.push(format!(
                "--- {}\nGo (recorded):\n{}\nRust:\n{}",
                s.name, want.1, got
            ));
        }
    }
    assert!(mismatches.is_empty(), "{}", mismatches.join("\n\n"));
}

/// Writes the recorded table from the Go build (`MOAI_GO_BIN`), for a developer to commit.
#[test]
#[ignore = "needs the Go build: set MOAI_GO_BIN"]
fn record_from_go() {
    let go = std::env::var("MOAI_GO_BIN").expect("MOAI_GO_BIN");
    let mut text = String::new();
    for s in SCENARIOS {
        text.push_str(&format!("=== {}\n{}\n\n", s.name, go_result(&go, s)));
    }
    fs::write(recorded_path(), format!("{}\n", text.trim_end())).unwrap();
}

/// With `MOAI_GO_BIN` set, compares the Rust gate with the live Go build; without it, nothing
/// to compare (the recorded table test above covers CI).
#[test]
fn live_parity_when_the_go_build_is_available() {
    let Ok(go) = std::env::var("MOAI_GO_BIN") else {
        return;
    };
    for s in SCENARIOS {
        assert_eq!(rust_result(s), go_result(&go, s), "scenario {}", s.name);
    }
}

// ---- moai hook pre-push -----------------------------------------------------------------------

struct PrePushScenario {
    name: &'static str,
    /// Config files under `.moai/config/sections/`.
    config: &'static [(&'static str, &'static str)],
    /// `(MOAI_ENFORCE_ON_PUSH, MOAI_PRE_PUSH, MOAI_GIT_CONVENTION)`
    env: (
        Option<&'static str>,
        Option<&'static str>,
        Option<&'static str>,
    ),
    stdin: &'static str,
    /// Commit subjects to create in a fresh repository first.
    history: &'static [&'static str],
}

const CC_SUBJECTS: &str = "feat: x\nfix(core): x\nfeat(a b): x\nfeat!: x\nfeat(a)!: x\nrevert: x\nfeat(a): (b): c\nfeat(a)(b): x\nfeat:x\nfeat: \nFeat: x\nfeat(): x\nfeat (a): x\nfeature: x\nfeat(a) : x\nwip\nfeat!!: x\n";
const ANGULAR_SUBJECTS: &str = "feat(my-scope): x\nfix: x\nbuild(a): y\nchore: x\nfeat(Router): x\nfeat(a_b): x\nfeat!: x\nrevert: x\n";
const KARMA_SUBJECTS: &str = "chore: x\nstyle(a b): y\ntest: z\nrevert: x\nbuild: x\nfeat!: x\n";
const LONG: &str = "feat: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\nfeat: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\n";
const ENFORCE_TEAM: (&str, &str) = (
    "git-strategy.yaml",
    "git_strategy:\n  mode: team\n  team:\n    hooks:\n      pre_push: enforce\n",
);

const PREPUSH_SCENARIOS: &[PrePushScenario] = &[
    PrePushScenario { name: "pp_off_by_default", config: &[], env: (None, None, None), stdin: "bad message\n", history: &[] },
    PrePushScenario { name: "pp_switch_word_yes_is_off", config: &[], env: (Some("yes"), Some("enforce"), None), stdin: "bad message\n", history: &[] },
    PrePushScenario { name: "pp_default_action_warns", config: &[], env: (Some("1"), None, None), stdin: "bad message\nfeat: ok\n", history: &[] },
    PrePushScenario { name: "pp_enforce_env", config: &[], env: (Some("true"), Some("enforce"), None), stdin: "bad message\n", history: &[] },
    PrePushScenario { name: "pp_skip_env", config: &[], env: (Some("1"), Some("skip"), None), stdin: "bad message\n", history: &[] },
    PrePushScenario { name: "pp_bad_action_env_falls_back", config: &[], env: (Some("1"), Some("loud"), None), stdin: "bad message\n", history: &[] },
    PrePushScenario { name: "pp_config_switch_and_team_enforce", config: &[("git-convention.yaml", "git_convention:\n  validation:\n    enforce_on_push: true\n"), ENFORCE_TEAM], env: (None, None, None), stdin: "bad message\n", history: &[] },
    PrePushScenario { name: "pp_config_switch_only_warns", config: &[("git-convention.yaml", "git_convention:\n  validation:\n    enforce_on_push: true\n")], env: (None, None, None), stdin: "bad message\n", history: &[] },
    PrePushScenario { name: "pp_unknown_mode_enforces", config: &[("git-convention.yaml", "git_convention:\n  validation:\n    enforce_on_push: true\n"), ("git-strategy.yaml", "git_strategy:\n  mode: custom\n")], env: (None, None, None), stdin: "bad message\n", history: &[] },
    PrePushScenario { name: "pp_unparsable_hook_value_enforces", config: &[("git-convention.yaml", "git_convention:\n  validation:\n    enforce_on_push: true\n"), ("git-strategy.yaml", "git_strategy:\n  mode: manual\n  manual:\n    hooks:\n      pre_push: sometimes\n")], env: (None, None, None), stdin: "bad message\n", history: &[] },
    PrePushScenario { name: "pp_unknown_convention", config: &[], env: (Some("1"), None, Some("bogus")), stdin: "feat: x\n", history: &[] },
    PrePushScenario { name: "pp_empty_input", config: &[], env: (Some("1"), None, None), stdin: "\n  \n", history: &[] },
    PrePushScenario { name: "pp_all_valid", config: &[], env: (Some("1"), None, None), stdin: "feat: a\nfix(core): b\n  docs: c  \n", history: &[] },
    PrePushScenario { name: "pp_conventional_patterns", config: &[], env: (Some("1"), Some("warn"), Some("conventional-commits")), stdin: CC_SUBJECTS, history: &[] },
    PrePushScenario { name: "pp_angular_patterns", config: &[], env: (Some("1"), Some("warn"), Some("angular")), stdin: ANGULAR_SUBJECTS, history: &[] },
    PrePushScenario { name: "pp_karma_patterns", config: &[], env: (Some("1"), Some("warn"), Some("karma")), stdin: KARMA_SUBJECTS, history: &[] },
    PrePushScenario { name: "pp_header_length", config: &[], env: (Some("1"), Some("warn"), None), stdin: LONG, history: &[] },
    PrePushScenario { name: "pp_max_length_config", config: &[("git-convention.yaml", "git_convention:\n  validation:\n    max_length: 20\n")], env: (Some("1"), Some("warn"), None), stdin: "feat: a short subject here\n", history: &[] },
    PrePushScenario { name: "pp_suggestions", config: &[], env: (Some("1"), Some("warn"), None), stdin: "Fix the bug\nAdd a thing\nUpdate the README\nTest more\nClean up code\nRefactor it\nSomething else\n", history: &[] },
    PrePushScenario { name: "pp_multibyte_subject", config: &[], env: (Some("1"), Some("warn"), Some("angular")), stdin: "한글 제목\n", history: &[] },
    PrePushScenario { name: "pp_quote_escapes", config: &[], env: (Some("1"), Some("warn"), None), stdin: "say \"hi\" \\ there\n", history: &[] },
    PrePushScenario { name: "pp_auto_history_conventional", config: &[], env: (Some("1"), Some("warn"), None), stdin: "bad message\n", history: &["feat: a", "fix: b", "revert: c", "docs: d"] },
    PrePushScenario { name: "pp_auto_history_free_form_uses_fallback", config: &[("git-convention.yaml", "git_convention:\n  auto_detection:\n    fallback: angular\n")], env: (Some("1"), Some("warn"), None), stdin: "bad message\n", history: &["Update stuff", "More stuff", "Fixes", "wip"] },
    PrePushScenario { name: "pp_auto_detection_off", config: &[("git-convention.yaml", "git_convention:\n  auto_detection:\n    enabled: false\n    fallback: karma\n")], env: (Some("1"), Some("warn"), None), stdin: "bad message\n", history: &["feat: a", "fix: b"] },
    // Only conventional-commits accepts `revert`, so the winner is never a three-way tie (Go breaks a tie by map order, which is random).
    PrePushScenario { name: "pp_auto_threshold_met", config: &[("git-convention.yaml", "git_convention:\n  auto_detection:\n    confidence_threshold: 0.5\n    fallback: karma\n")], env: (Some("1"), Some("warn"), None), stdin: "bad message\n", history: &["revert: a", "wip"] },
    PrePushScenario { name: "pp_auto_threshold_missed", config: &[("git-convention.yaml", "git_convention:\n  auto_detection:\n    confidence_threshold: 0.5\n    fallback: karma\n")], env: (Some("1"), Some("warn"), None), stdin: "bad message\n", history: &["revert: a", "revert: b", "wip", "wip", "wip"] },
    // A value the Go loader refuses makes it ignore the whole config: defaults, and no profile, which enforces.
    PrePushScenario { name: "pp_invalid_config_convention", config: &[("git-convention.yaml", "git_convention:\n  convention: bogus\n")], env: (Some("1"), None, None), stdin: "bad message\n", history: &[] },
    PrePushScenario { name: "pp_invalid_config_max_length", config: &[("git-convention.yaml", "git_convention:\n  validation:\n    max_length: -1\n")], env: (Some("1"), None, None), stdin: "bad message\n", history: &[] },
    PrePushScenario { name: "pp_invalid_config_threshold", config: &[("git-convention.yaml", "git_convention:\n  auto_detection:\n    confidence_threshold: 2\n")], env: (Some("1"), None, None), stdin: "bad message\n", history: &[] },
    PrePushScenario { name: "pp_invalid_config_sample_size", config: &[("git-convention.yaml", "git_convention:\n  auto_detection:\n    sample_size: -1\n")], env: (Some("1"), None, None), stdin: "bad message\n", history: &[] },
    PrePushScenario { name: "pp_invalid_config_switch_is_lost", config: &[("git-convention.yaml", "git_convention:\n  convention: bogus\n  validation:\n    enforce_on_push: true\n")], env: (None, None, None), stdin: "bad message\n", history: &[] },
    PrePushScenario { name: "pp_invisible_characters_quoted", config: &[], env: (Some("1"), Some("warn"), None), stdin: "a\u{a0}b\u{200b}c\u{3000}d\nx\u{7f}y\u{85}z\u{feff}w\n", history: &[] },
    PrePushScenario { name: "pp_dotted_capital_i_suggestion", config: &[], env: (Some("1"), Some("warn"), None), stdin: "F\u{130}X the thing\n", history: &[] },
    PrePushScenario { name: "pp_control_escapes_quoted", config: &[], env: (Some("1"), Some("warn"), None), stdin: "bell\u{7}back\u{8}form\u{c}vtab\u{b}end\n", history: &[] },
    PrePushScenario { name: "pp_unknown_fallback", config: &[("git-convention.yaml", "git_convention:\n  auto_detection:\n    enabled: false\n    fallback: nope\n")], env: (Some("1"), None, None), stdin: "x\n", history: &[] },
];

fn build_prepush(s: &PrePushScenario) -> TempDir {
    let project = TempDir::new().unwrap();
    for (file, text) in s.config {
        let d = project.path().join(".moai/config/sections");
        fs::create_dir_all(&d).unwrap();
        fs::write(d.join(file), text).unwrap();
    }
    if !s.history.is_empty() {
        git_ok(project.path(), &["init", "-q", "."]);
        for subject in s.history {
            git_ok(
                project.path(),
                &[
                    "-c",
                    "user.name=t",
                    "-c",
                    "user.email=t@t",
                    "commit",
                    "-q",
                    "--allow-empty",
                    "-m",
                    subject,
                ],
            );
        }
    }
    project
}

/// Both streams, without the Go build's own log lines; its decorated error box becomes the one
/// message line it carries (`Pre-Push: ... .` -> `pre-push: ...`).
fn normalize_prepush(raw: &[u8]) -> String {
    let text = String::from_utf8_lossy(raw);
    let mut out = Vec::new();
    for line in text.lines() {
        let t = line.trim();
        if line.starts_with("time=")
            || t == "ERROR"
            || (t.is_empty() && out.last().is_none_or(|l: &String| l.starts_with("error: ")))
        {
            continue;
        }
        if let Some(rest) = t
            .strip_prefix("Pre-Push:")
            .or_else(|| t.strip_prefix("pre-push:"))
        {
            out.push(format!("error: pre-push:{}", rest.trim_end_matches('.')));
            continue;
        }
        out.push(line.to_string());
    }
    out.join("\n").trim_end().to_string()
}

fn prepush_rust(s: &PrePushScenario) -> String {
    let project = build_prepush(s);
    let inv = quality_gate::pre_push::PrePushInvocation {
        project_dir: project.path().canonicalize().unwrap(),
        repo_dir: None,
        env: quality_gate::pre_push::PrePushEnv {
            enforce_on_push: s.env.0.map(String::from),
            pre_push_mode: s.env.1.map(String::from),
            convention: s.env.2.map(String::from),
        },
        stdin: s.stdin.as_bytes().to_vec(),
    };
    let (mut out, mut err) = (Vec::new(), Vec::new());
    let code = quality_gate::pre_push::run(&inv, &mut out, &mut err);
    format!(
        "exit: {code}\n--stdout\n{}\n--stderr\n{}",
        normalize_prepush(&out),
        normalize_prepush(&err)
    )
    .trim_end()
    .to_string()
}

fn prepush_go(go_bin: &str, s: &PrePushScenario) -> String {
    use std::io::Write;
    let project = build_prepush(s);
    let home = TempDir::new().unwrap();
    let mut cmd = Command::new(go_bin);
    cmd.args(["hook", "pre-push"])
        .current_dir(project.path().canonicalize().unwrap())
        .clean_env()
        .env("HOME", home.path())
        .env("PATH", "/usr/bin:/bin")
        .stdin(std::process::Stdio::piped())
        .stdout(std::process::Stdio::piped())
        .stderr(std::process::Stdio::piped());
    for (key, value) in [
        ("MOAI_ENFORCE_ON_PUSH", s.env.0),
        ("MOAI_PRE_PUSH", s.env.1),
        ("MOAI_GIT_CONVENTION", s.env.2),
    ] {
        if let Some(v) = value {
            cmd.env(key, v);
        }
    }
    let mut child = cmd.spawn().expect("run the Go build");
    child
        .stdin
        .take()
        .unwrap()
        .write_all(s.stdin.as_bytes())
        .unwrap();
    let out = child.wait_with_output().unwrap();
    format!(
        "exit: {}\n--stdout\n{}\n--stderr\n{}",
        out.status.code().unwrap_or(-1),
        normalize_prepush(&out.stdout),
        normalize_prepush(&out.stderr)
    )
    .trim_end()
    .to_string()
}

fn prepush_recorded_path() -> std::path::PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR"))
        .join(".moai/specs/SPEC-20261120-rust-pre-push/parity-recorded.txt")
}

/// AC-1: every pre-push scenario of the recorded table matches the Rust check.
#[test]
fn prepush_parity_with_the_recorded_go_table() {
    let text = fs::read_to_string(prepush_recorded_path())
        .expect("recorded table is committed with the SPEC");
    let recorded = parse_recorded(&text);
    assert_eq!(
        recorded.len(),
        PREPUSH_SCENARIOS.len(),
        "recorded table and scenario list differ in size"
    );
    let mut mismatches = Vec::new();
    for s in PREPUSH_SCENARIOS {
        let want = recorded
            .iter()
            .find(|(n, _)| n == s.name)
            .unwrap_or_else(|| panic!("scenario {} is not recorded", s.name));
        let got = prepush_rust(s);
        if got != want.1 {
            mismatches.push(format!(
                "--- {}\nGo (recorded):\n{}\nRust:\n{}",
                s.name, want.1, got
            ));
        }
    }
    assert!(mismatches.is_empty(), "{}", mismatches.join("\n\n"));
}

/// Writes the recorded pre-push table from the Go build (`MOAI_GO_BIN`).
#[test]
#[ignore = "needs the Go build: set MOAI_GO_BIN"]
fn prepush_record_from_go() {
    let go = std::env::var("MOAI_GO_BIN").expect("MOAI_GO_BIN");
    let mut text = String::new();
    for s in PREPUSH_SCENARIOS {
        text.push_str(&format!("=== {}\n{}\n\n", s.name, prepush_go(&go, s)));
    }
    fs::write(prepush_recorded_path(), format!("{}\n", text.trim_end())).unwrap();
}

/// With `MOAI_GO_BIN` set, compares the Rust check with the live Go build.
#[test]
fn prepush_live_parity_when_the_go_build_is_available() {
    let Ok(go) = std::env::var("MOAI_GO_BIN") else {
        return;
    };
    for s in PREPUSH_SCENARIOS {
        assert_eq!(prepush_rust(s), prepush_go(&go, s), "scenario {}", s.name);
    }
}

/// Differential check over many subjects (`MOAI_PP_SUBJECTS`: one per line, `MOAI_GO_BIN`): the
/// hand-written matchers against the Go build's regular expressions, for every built-in name.
#[test]
#[ignore = "needs the Go build and a subject list: set MOAI_GO_BIN and MOAI_PP_SUBJECTS"]
fn prepush_differential_over_many_subjects() {
    let go = std::env::var("MOAI_GO_BIN").expect("MOAI_GO_BIN");
    let list = std::env::var("MOAI_PP_SUBJECTS").expect("MOAI_PP_SUBJECTS");
    let subjects = fs::read_to_string(list).unwrap();
    for convention in ["conventional-commits", "angular", "karma"] {
        // One subject per run, so a mismatch names the subject.
        for line in subjects.lines().filter(|l| !l.trim().is_empty()) {
            let stdin: &'static str = Box::leak(format!("{line}\n").into_boxed_str());
            let scenario = PrePushScenario {
                name: "diff",
                config: &[],
                env: (Some("1"), Some("warn"), Some(convention)),
                stdin,
                history: &[],
            };
            assert_eq!(
                prepush_rust(&scenario),
                prepush_go(&go, &scenario),
                "{convention}: {line:?}"
            );
        }
    }
}
