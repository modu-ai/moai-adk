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

#[path = "../src/quality_gate/mod.rs"]
#[allow(dead_code)]
mod quality_gate;

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
        .env_clear()
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
