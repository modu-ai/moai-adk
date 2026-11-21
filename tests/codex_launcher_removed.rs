//! SPEC-20261121-codex-launcher-and-models-removed (REQ-1, REQ-2, REQ-3): `moai codex` is an
//! unknown command, the host word `codex` is refused like any other unknown host, and the
//! launcher builds no Codex command line. Every case runs the built binary in a temporary git
//! project with a temporary HOME; a stub `codex` on PATH proves that nothing starts it.
#![cfg(unix)]
mod common;
use common::CleanEnv;
use serde_json::{json, Value};
use std::os::unix::fs::PermissionsExt;
use std::process::{Command, Output};
use std::{fs, path::Path};

struct Fixture {
    project: tempfile::TempDir,
    bin: tempfile::TempDir,
    home: tempfile::TempDir,
}

fn fixture() -> Fixture {
    let project = tempfile::tempdir().unwrap();
    let init = common::git::command()
        .args(["init", "-q"])
        .current_dir(project.path())
        .status()
        .unwrap();
    assert!(init.success());
    // An initialized project, so no launch runs the first-launch initialization.
    fs::create_dir(project.path().join(".moai-cli")).unwrap();
    fs::write(project.path().join(".moai-cli/project.yaml"), "{}\n").unwrap();
    let bin = tempfile::tempdir().unwrap();
    // A stub `codex` and `claude` that leave a trace when they start.
    for host in ["codex", "claude"] {
        let path = bin.path().join(host);
        fs::write(
            &path,
            "#!/bin/sh\nprintf started > \"$MOAI_TEST_HOST_MARKER\"\nexit 0\n",
        )
        .unwrap();
        fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
    }
    Fixture {
        project,
        bin,
        home: tempfile::tempdir().unwrap(),
    }
}

fn run(f: &Fixture, dir: &Path, args: &[&str]) -> Output {
    Command::new(common::MOAI_BIN)
        .clean_env()
        .args(args)
        .current_dir(dir)
        .env("PATH", format!("{}:/usr/bin:/bin", f.bin.path().display()))
        .env("HOME", f.home.path())
        .env("USERPROFILE", f.home.path())
        .env("LANG", "C")
        .env("MOAI_TEST_HOST_MARKER", f.home.path().join("host-started"))
        .env("ANTHROPIC_AUTH_TOKEN", "moai-test-dummy-not-real")
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .output()
        .unwrap()
}

fn started(f: &Fixture) -> bool {
    f.home.path().join("host-started").exists()
}

fn out_text(out: &Output) -> String {
    format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

/// REQ-1: `moai codex` is an unknown command: the usage envelope, exit 2, nothing started.
#[test]
fn moai_codex_is_an_unknown_command() {
    let f = fixture();
    for args in [
        &["codex"][..],
        &["codex", "-l"][..],
        &["codex", "-a", "--dry-run"][..],
        &["codex", "--help"][..],
    ] {
        let out = run(&f, f.project.path(), args);
        assert_eq!(out.status.code(), Some(2), "{args:?}: {}", out_text(&out));
        let v: Value = serde_json::from_slice(&out.stdout)
            .unwrap_or_else(|e| panic!("{args:?}: stdout is not the envelope ({e}): {out:?}"));
        assert_eq!(
            v,
            json!({
                "protocol_version": 1,
                "operation": "cli",
                "ok": false,
                "error": {
                    "kind": "usage",
                    "message": "unknown command; see --help",
                    "retryable": false
                }
            }),
            "{args:?}"
        );
        assert!(!started(&f), "{args:?}: a host was started");
    }
}

/// REQ-2: the host word after `-a` (`-l` takes none) refuses `codex` exactly as it refuses any unknown
/// word: same exit code, same message shape, listing `cc|glm|omp` and no `codex`.
#[test]
fn the_codex_host_word_is_refused_like_any_unknown_host() {
    let f = fixture();
    let flag = "-a";
    let nope = run(&f, f.project.path(), &[flag, "nope", "--dry-run"]);
    let codex = run(&f, f.project.path(), &[flag, "codex", "--dry-run"]);
    assert_ne!(nope.status.code(), Some(0), "{flag} nope: {nope:?}");
    assert_eq!(codex.status.code(), nope.status.code(), "{flag}: {codex:?}");
    assert!(codex.stdout.is_empty(), "{flag}: {codex:?}");
    let text = out_text(&codex);
    assert!(text.contains("(cc|glm|omp)"), "{flag}: {text}");
    assert!(!text.contains("codex|"), "{flag}: {text}");
    assert!(!text.contains("|codex"), "{flag}: {text}");
    assert_eq!(
        text.replace("codex", "WORD"),
        out_text(&nope).replace("nope", "WORD"),
        "{flag}: the refusal differs from the unknown-host refusal"
    );
    assert!(!started(&f), "{flag}: a host was started");
    // Without `--dry-run` too: the stub `codex` on PATH is never started.
    let out = run(&f, f.project.path(), &["-a", "codex"]);
    assert_ne!(out.status.code(), Some(0), "{out:?}");
    assert!(!started(&f), "a host was started");
}

/// REQ-1: the usage text of a bare `moai` (no arguments) names the three launcher words.
#[test]
fn the_bare_usage_line_names_no_codex() {
    let f = fixture();
    let out = run(&f, f.project.path(), &[]);
    let text = out_text(&out);
    assert!(text.contains("cc|glm|omp"), "{text}");
    assert!(!text.contains("codex"), "{text}");
}

/// REQ-2: the legacy `session --host codex` is an unknown host like `--host nope`.
#[test]
fn the_legacy_session_launcher_refuses_codex_as_a_host() {
    let f = fixture();
    let nope = run(
        &f,
        f.project.path(),
        &["session", "--host", "nope", "--name", "x", "--dry-run"],
    );
    let codex = run(
        &f,
        f.project.path(),
        &["session", "--host", "codex", "--name", "x", "--dry-run"],
    );
    assert_ne!(nope.status.code(), Some(0), "{nope:?}");
    assert_eq!(codex.status.code(), nope.status.code(), "{codex:?}");
    let text = out_text(&codex);
    assert!(
        !text.contains("codex|") && !text.contains("|codex"),
        "{text}"
    );
    assert!(!started(&f));
}

/// REQ-3: a session request names no Codex program and the dry-run output of a real launch holds
/// no Codex argument note.
#[test]
fn a_session_request_has_no_codex_program_and_the_notes_name_no_codex() {
    let request: moai_cli::session::SessionRequest =
        serde_json::from_value(json!({"host": "claude"})).unwrap();
    let value = serde_json::to_value(&request).unwrap();
    let keys: Vec<&str> = value
        .as_object()
        .unwrap()
        .keys()
        .map(String::as_str)
        .collect();
    assert!(!keys.contains(&"codex_bin"), "{keys:?}");
    assert!(!value.to_string().contains("codex"), "{value}");

    let f = fixture();
    for args in [
        &["cc", "-l", "--dry-run"][..],
        &["cc", "-a", "--dry-run"][..],
        &["cc", "--dry-run"][..],
        &["omp", "-a", "--dry-run"][..],
        &["glm", "-l", "--dry-run"][..],
    ] {
        let out = run(&f, f.project.path(), args);
        assert!(out.status.success(), "{args:?}: {}", out_text(&out));
        let v: Value = serde_json::from_slice(&out.stdout).unwrap();
        let notes = v["notes"].to_string().to_ascii_lowercase();
        assert!(!notes.contains("codex"), "{args:?}: {notes}");
        assert!(!notes.contains("config.toml"), "{args:?}: {notes}");
    }
}
