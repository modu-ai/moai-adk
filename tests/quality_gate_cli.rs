//! SPEC-20261120-rust-quality-gate: the command's own behaviour — the pre-commit marker, the
//! enabled switch, the summary, the run lock. Names start with the SPEC's verify words
//! (`summary_`, `lock_`) so the filters pick them.

#![cfg(unix)] // the gate runs process groups and flock; a Windows port is not part of this SPEC

#[path = "../src/quality_gate/mod.rs"]
#[allow(dead_code)]
mod quality_gate;

mod common;

use common::git::git_ok;
use quality_gate::lock;
use quality_gate::{run, Invocation};
use std::ffi::OsString;
use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};
use tempfile::TempDir;

fn bin_dir(tools: &[(&str, &str)]) -> TempDir {
    let d = TempDir::new().unwrap();
    for (name, body) in tools {
        let p = d.path().join(name);
        fs::write(&p, format!("#!/bin/sh\n{body}\n")).unwrap();
        fs::set_permissions(&p, fs::Permissions::from_mode(0o755)).unwrap();
    }
    d
}

fn path_of(bin: &TempDir) -> OsString {
    std::env::join_paths([bin.path(), Path::new("/usr/bin"), Path::new("/bin")]).unwrap()
}

fn go_project() -> TempDir {
    let p = TempDir::new().unwrap();
    fs::write(p.path().join("go.mod"), "module x\n").unwrap();
    fs::write(p.path().join("main.go"), "package x\n").unwrap();
    p
}

fn config(project: &Path, text: &str) {
    let d = project.join(".moai/config/sections");
    fs::create_dir_all(&d).unwrap();
    fs::write(d.join("gate.yaml"), text).unwrap();
}

fn gate(project: &Path, bin: &TempDir, marker: bool) -> (i32, String) {
    let inv = Invocation {
        project_dir: project.to_path_buf(),
        precommit_marker: marker,
        path: path_of(bin),
    };
    let mut err = Vec::new();
    let code = run(&inv, &mut err);
    (code, String::from_utf8_lossy(&err).into_owned())
}

const SKIP_NOTICE: &str = "[moai gate] pre-commit context: heavy gate not opted in — skipping the project-wide steps (opt in via gate.pre_commit.enabled: true in .moai/config/sections/gate.yaml)\n";

/// AC-1 (REQ-1): with the marker and the default config no step runs and the notice is exact;
/// opting in runs the steps again.
#[test]
fn the_precommit_marker_skips_unless_opted_in() {
    let p = go_project();
    let bin = bin_dir(&[("go", "echo ran >&2; exit 3")]);
    let (code, err) = gate(p.path(), &bin, true);
    assert_eq!(code, 0, "{err}");
    assert_eq!(err, SKIP_NOTICE);
    config(p.path(), "gate:\n  pre_commit:\n    enabled: true\n");
    let (code, err) = gate(p.path(), &bin, true);
    assert_eq!(
        code, 1,
        "opted in: the steps run and the failing one fails: {err}"
    );
    // Without the marker the gate runs whatever the opt-in says.
    let q = go_project();
    let (code, err) = gate(q.path(), &bin, false);
    assert_eq!(code, 1, "{err}");
}

/// AC-1 (REQ-10): `enabled: false` prints nothing and exits 0 even for a project whose step
/// would fail.
#[test]
fn a_disabled_gate_passes_silently() {
    let p = go_project();
    config(p.path(), "gate:\n  enabled: false\n");
    let bin = bin_dir(&[("go", "exit 3")]);
    let (code, err) = gate(p.path(), &bin, false);
    assert_eq!(code, 0);
    assert_eq!(err, "", "no output at all");
}

/// AC-5 (REQ-8): the passing run prints the summary: header with the configured count, one row
/// per step in run order, the executed rows with duration, start time and command.
#[test]
fn summary_lists_every_configured_step_in_order() {
    let p = go_project();
    let bin = bin_dir(&[("go", "exit 0")]);
    let (code, err) = gate(p.path(), &bin, false);
    assert_eq!(code, 0, "{err}");
    let at = err
        .find("quality gate steps (4 configured):")
        .unwrap_or_else(|| panic!("{err}"));
    let rows: Vec<&str> = err[at..].lines().skip(1).take(4).collect();
    assert!(rows[0].starts_with("  - go vet: executed in "), "{rows:?}");
    assert!(rows[0].ends_with(" — go vet ./..."), "{rows:?}");
    assert!(
        rows[0].contains(" at 20"),
        "start time is an RFC 3339 stamp: {rows:?}"
    );
    assert_eq!(rows[1], "  - typecheck: skipped — no default for this language; set gate.typecheck.command to enable one");
    assert_eq!(
        rows[2],
        "  - golangci-lint: skipped — the optional tool \"golangci-lint\" was not found on PATH"
    );
    assert!(
        rows[3].starts_with("  - go test: executed in ") && rows[3].ends_with(" — go test ./..."),
        "{rows:?}"
    );
}

/// AC-5 (REQ-8): a failed step prints the failure block, then the summary where the steps after
/// it read `not reached`.
#[test]
fn summary_marks_the_steps_after_a_failure_as_not_reached() {
    let p = go_project();
    let bin = bin_dir(&[("go", "echo vet-broke >&2; exit 3")]);
    let (code, err) = gate(p.path(), &bin, false);
    assert_eq!(code, 1, "{err}");
    let failed = err
        .find("quality gate failed: go vet\n\nvet-broke\n\nquality gate steps (4 configured):\n")
        .unwrap_or_else(|| panic!("{err}"));
    let tail = &err[failed..];
    assert!(tail.contains("\n  - typecheck: not reached\n"), "{tail}");
    assert!(
        tail.contains("\n  - golangci-lint: not reached\n"),
        "{tail}"
    );
    assert!(tail.contains("\n  - go test: not reached"), "{tail}");
}

/// AC-5 (REQ-8): the axes this build does not run yet are named, not left out (D-2).
#[test]
fn summary_names_the_axes_not_ported_yet() {
    let p = go_project();
    let bin = bin_dir(&[("go", "exit 0")]);
    let (_, err) = gate(p.path(), &bin, false);
    for axis in ["graph-freshness", "ast-grep"] {
        assert!(
            err.contains(&format!("  - {axis}: not run — not ported yet")),
            "{axis}: {err}"
        );
    }
    // A node project also names the node-only decisions (test and lint selection, typecheck).
    let n = TempDir::new().unwrap();
    fs::write(n.path().join("package.json"), "{}").unwrap();
    let (_, err) = gate(
        n.path(),
        &bin_dir(&[("npm", "exit 0"), ("npx", "exit 0")]),
        false,
    );
    assert!(err.contains("not run — not ported yet"), "{err}");
    assert!(err.contains("npm"), "the node steps are named: {err}");
}

/// AC-5 (REQ-8): a project with no recognised language has no step rows and passes, but the
/// axes this build does not run are still named (audit N3); with both disabled it is silent.
#[test]
fn summary_names_the_unported_axes_when_no_language_is_detected() {
    let p = TempDir::new().unwrap();
    let (code, err) = gate(p.path(), &bin_dir(&[]), false);
    assert_eq!(code, 0, "{err}");
    assert!(err.contains("quality gate steps (0 configured):"), "{err}");
    assert!(
        err.contains("  - graph-freshness: not run — not ported yet"),
        "{err}"
    );
    assert!(
        err.contains("  - ast-grep: not run — not ported yet"),
        "{err}"
    );
    assert!(!err.contains("  - vet:"), "{err}");
    let quiet = TempDir::new().unwrap();
    let dir = quiet.path().join(".moai/config/sections");
    fs::create_dir_all(&dir).unwrap();
    fs::write(
        dir.join("gate.yaml"),
        "gate:\n  ast_grep_gate:\n    enabled: false\n  graph_freshness:\n    enabled: false\n",
    )
    .unwrap();
    let (code, err) = gate(quiet.path(), &bin_dir(&[]), false);
    assert_eq!(code, 0);
    assert!(!err.contains("quality gate steps"), "{err}");
}

/// AC-5 (REQ-9): the lock file is created under `.moai/state/` and is free again after a run.
#[test]
fn lock_is_taken_and_released_by_a_run() {
    let p = go_project();
    let bin = bin_dir(&[("go", "exit 0")]);
    let (code, err) = gate(p.path(), &bin, false);
    assert_eq!(code, 0, "{err}");
    let lock_path = p.path().join(".moai/state/gate-run.lock");
    assert!(lock_path.exists());
    // flock belongs to the open file description, so a child that another test's thread forked
    // in the instant before its exec can hold it for a moment: retry briefly, never assume.
    let released = (0..40).any(|_| {
        let ok = lock::acquire(p.path()).is_ok();
        if !ok {
            std::thread::sleep(std::time::Duration::from_millis(50));
        }
        ok
    });
    assert!(released, "released after the run");
}

/// AC-5 (REQ-9): a held lock makes the run wait its budget, say so, then run unserialized; the
/// lock never decides the exit code.
#[test]
fn lock_held_by_another_run_waits_then_runs_unserialized() {
    let p = go_project();
    config(p.path(), "gate:\n  timeouts:\n    lock_wait: 1\n");
    let held = lock::acquire(p.path()).expect("first acquire");
    // `acquire` records this process as the owner; a second acquire in the same process must
    // be refused as held (flock is per open file description).
    assert!(matches!(
        lock::acquire(p.path()),
        Err(lock::LockError::Held)
    ));
    let bin = bin_dir(&[("go", "exit 0")]);
    let started = std::time::Instant::now();
    let (code, err) = gate(p.path(), &bin, false);
    assert_eq!(code, 0, "{err}");
    assert!(started.elapsed().as_millis() >= 900, "waited the budget");
    let pid = std::process::id();
    assert!(
        err.contains(&format!(
            "gate-run lock: held by pid {pid} — waiting (budget 1s)"
        )),
        "{err}"
    );
    assert!(
        err.contains("gate-run lock: wait budget 1s expired — running unserialized"),
        "{err}"
    );
    drop(held);
}

/// AC-5 (REQ-9): the lock file's owner record is the Go build's shape, so a Go run and a Rust
/// run on one tree can read each other's holder.
#[test]
fn lock_owner_record_is_json_with_pid_and_created_at() {
    let p = go_project();
    let held = lock::acquire(p.path()).unwrap();
    let raw = fs::read_to_string(p.path().join(".moai/state/gate-run.lock")).unwrap();
    assert!(
        raw.contains(&format!("\"pid\": {}", std::process::id())),
        "{raw}"
    );
    assert!(raw.contains("\"created_at\": \"20"), "{raw}");
    drop(held);
    let _ = PathBuf::new();
}

// ---- moai hook pre-push (SPEC-20261120-rust-pre-push) -----------------------------------------

mod prepush_support {
    use super::quality_gate::pre_push::{run as pre_push, PrePushEnv, PrePushInvocation};
    use std::fs;
    use std::path::Path;

    pub fn config(project: &Path, file: &str, text: &str) {
        let d = project.join(".moai/config/sections");
        fs::create_dir_all(&d).unwrap();
        fs::write(d.join(file), text).unwrap();
    }

    pub fn env(enforce: Option<&str>, mode: Option<&str>, convention: Option<&str>) -> PrePushEnv {
        PrePushEnv {
            enforce_on_push: enforce.map(String::from),
            pre_push_mode: mode.map(String::from),
            convention: convention.map(String::from),
        }
    }

    /// `(exit, stdout, stderr)`
    pub fn check(project: &Path, env: PrePushEnv, stdin: &str) -> (i32, String, String) {
        let inv = PrePushInvocation {
            project_dir: project.to_path_buf(),
            env,
            stdin: stdin.as_bytes().to_vec(),
        };
        let (mut out, mut err) = (Vec::new(), Vec::new());
        let code = pre_push(&inv, &mut out, &mut err);
        (
            code,
            String::from_utf8_lossy(&out).into_owned(),
            String::from_utf8_lossy(&err).into_owned(),
        )
    }
}
use prepush_support::{check, config as pp_config, env as pp_env};

/// AC-2 (REQ-2): with no opt-in the check reads nothing and says nothing, whatever the input.
#[test]
fn prepush_is_off_by_default() {
    let p = TempDir::new().unwrap();
    let (code, out, err) = check(p.path(), pp_env(None, None, None), "bad message\n");
    assert_eq!((code, out.as_str(), err.as_str()), (0, "", ""));
    // An unrecognised value of the switch is off too.
    let (code, out, err) = check(
        p.path(),
        pp_env(Some("yes"), Some("enforce"), None),
        "bad message\n",
    );
    assert_eq!((code, out.as_str(), err.as_str()), (0, "", ""));
}

/// AC-2 (REQ-1): the default action is `warn` (team mode): the report is printed, the exit is 0.
#[test]
fn prepush_default_action_reports_but_exits_zero() {
    let p = TempDir::new().unwrap();
    let (code, out, err) = check(
        p.path(),
        pp_env(Some("1"), None, None),
        "bad message\nfeat: ok\n",
    );
    assert_eq!(code, 0);
    assert_eq!(out, "");
    assert!(err.starts_with("1 of 2 commit(s) violate conventional-commits convention:\n\nCommit message violates conventional-commits convention:\n  Message: \"bad message\"\n\n  [PATTERN] Pattern mismatch (expected: ^("), "{err}");
    assert!(err.contains("\n    Suggestion: chore: bad message\n\n  Examples of valid messages:\n    - feat(auth): add JWT token validation\n"), "{err}");
}

/// AC-2 (REQ-1): `enforce` ends with 2; `skip` ends silently with 0.
#[test]
fn prepush_enforce_exits_two_and_skip_is_silent() {
    let p = TempDir::new().unwrap();
    let (code, _, err) = check(
        p.path(),
        pp_env(Some("true"), Some("enforce"), None),
        "bad message\n",
    );
    assert_eq!(code, 2);
    assert!(err.contains("1 of 1 commit(s) violate"), "{err}");
    let (code, out, err) = check(
        p.path(),
        pp_env(Some("true"), Some("skip"), None),
        "bad message\n",
    );
    assert_eq!((code, out.as_str(), err.as_str()), (0, "", ""));
}

/// AC-2 (REQ-1, REQ-2): the files decide when the environment says nothing: the switch in
/// `git-convention.yaml`, the action in the current mode of `git-strategy.yaml`; an unknown mode
/// or an unparsable value enforces.
#[test]
fn prepush_config_files_decide_the_switch_and_the_action() {
    let p = TempDir::new().unwrap();
    pp_config(
        p.path(),
        "git-convention.yaml",
        "git_convention:\n  validation:\n    enforce_on_push: true\n",
    );
    assert_eq!(
        check(p.path(), pp_env(None, None, None), "bad message\n").0,
        0,
        "team mode warns"
    );
    pp_config(
        p.path(),
        "git-strategy.yaml",
        "git_strategy:\n  mode: team\n  team:\n    hooks:\n      pre_push: enforce\n",
    );
    assert_eq!(
        check(p.path(), pp_env(None, None, None), "bad message\n").0,
        2
    );
    pp_config(
        p.path(),
        "git-strategy.yaml",
        "git_strategy:\n  mode: personal\n  team:\n    hooks:\n      pre_push: enforce\n",
    );
    assert_eq!(
        check(p.path(), pp_env(None, None, None), "bad message\n").0,
        0,
        "the personal profile keeps the warn default"
    );
    pp_config(
        p.path(),
        "git-strategy.yaml",
        "git_strategy:\n  mode: custom\n",
    );
    assert_eq!(
        check(p.path(), pp_env(None, None, None), "bad message\n").0,
        2,
        "unknown mode enforces"
    );
    pp_config(
        p.path(),
        "git-strategy.yaml",
        "git_strategy:\n  mode: manual\n  manual:\n    hooks:\n      pre_push: sometimes\n",
    );
    assert_eq!(
        check(p.path(), pp_env(None, None, None), "bad message\n").0,
        2,
        "unparsable value enforces"
    );
    pp_config(
        p.path(),
        "git-strategy.yaml",
        "git_strategy:\n  mode: manual\n  manual:\n    hooks:\n      pre_push: skip\n",
    );
    assert_eq!(
        check(p.path(), pp_env(None, None, None), "bad message\n"),
        (0, String::new(), String::new())
    );
}

/// AC-1 (REQ-3): an unknown convention name is a load error with exit 1.
#[test]
fn prepush_unknown_convention_is_a_load_error() {
    let p = TempDir::new().unwrap();
    let (code, out, err) = check(
        p.path(),
        pp_env(Some("1"), None, Some("bogus")),
        "feat: x\n",
    );
    assert_eq!(code, 1);
    assert_eq!(out, "");
    assert_eq!(err, "pre-push: load convention: load convention \"bogus\": unknown built-in convention: \"bogus\"\n");
}

/// AC-1 (REQ-5, REQ-6): nothing to validate and all-valid messages.
#[test]
fn prepush_empty_input_and_all_valid() {
    let p = TempDir::new().unwrap();
    let (code, out, err) = check(p.path(), pp_env(Some("1"), None, None), "\n  \n");
    assert_eq!(
        (code, out.as_str(), err.as_str()),
        (0, "No commit messages to validate.\n", "")
    );
    let (code, out, err) = check(
        p.path(),
        pp_env(Some("1"), None, None),
        "feat: a\nfix(core): b\n  docs: c  \n",
    );
    assert_eq!(
        (code, out.as_str(), err.as_str()),
        (
            0,
            "All 3 commit(s) follow conventional-commits convention.\n",
            ""
        )
    );
}

fn violates(convention: &str, subject: &str) -> bool {
    let p = TempDir::new().unwrap();
    let (_, out, err) = check(
        p.path(),
        pp_env(Some("1"), Some("warn"), Some(convention)),
        &format!("{subject}\n"),
    );
    assert!(
        out.is_empty() != err.is_empty(),
        "exactly one of the two streams speaks: {out:?} {err:?}"
    );
    !err.is_empty()
}

/// AC-2 (REQ-1): the three built-in patterns, edge by edge.
#[test]
fn prepush_builtin_patterns() {
    let cc = "conventional-commits";
    for ok in [
        "feat: x",
        "fix(core): x",
        "feat(a b): x",
        "feat!: x",
        "feat(a)!: x",
        "revert: x",
        "feat(a): (b): c",
        "feat(a)(b): x",
    ] {
        assert!(!violates(cc, ok), "{ok}");
    }
    for bad in [
        "feat:x",
        "feat: ",
        "Feat: x",
        "feat(): x",
        "feat (a): x",
        "feature: x",
        "feat(a) : x",
        "feat(a): ",
        "wip",
        "feat!!: x",
    ] {
        assert!(violates(cc, bad), "{bad}");
    }
    let ang = "angular";
    for ok in ["feat(my-scope): x", "fix: x", "build(a): y"] {
        assert!(!violates(ang, ok), "{ok}");
    }
    for bad in [
        "chore: x",
        "feat(Router): x",
        "feat(a_b): x",
        "feat!: x",
        "revert: x",
    ] {
        assert!(violates(ang, bad), "{bad}");
    }
    let karma = "karma";
    for ok in ["chore: x", "style(a b): y", "test: z"] {
        assert!(!violates(karma, ok), "{ok}");
    }
    for bad in ["revert: x", "build: x", "feat!: x"] {
        assert!(violates(karma, bad), "{bad}");
    }
}

/// AC-2 (REQ-1): the header length limit is bytes, 100 by default, and `max_length` overrides it;
/// the report names both numbers.
#[test]
fn prepush_header_length_limit() {
    let p = TempDir::new().unwrap();
    let ok = format!("feat: {}", "a".repeat(94)); // 100 bytes
    let long = format!("feat: {}", "a".repeat(95)); // 101 bytes
                                                    // The default action warns, so the exit is 0 either way: the streams tell pass from report.
    assert_eq!(
        check(p.path(), pp_env(Some("1"), None, None), &format!("{ok}\n")),
        (
            0,
            "All 1 commit(s) follow conventional-commits convention.\n".to_string(),
            String::new()
        )
    );
    let (_, out, err) = check(
        p.path(),
        pp_env(Some("1"), None, None),
        &format!("{long}\n"),
    );
    assert_eq!(out, "");
    assert!(
        err.contains("  [LENGTH] Header too long (101 characters, max 100 characters)\n"),
        "{err}"
    );
    pp_config(
        p.path(),
        "git-convention.yaml",
        "git_convention:\n  validation:\n    max_length: 20\n",
    );
    let (_, _, err) = check(
        p.path(),
        pp_env(Some("1"), None, None),
        "feat: a short subject here\n",
    );
    assert!(
        err.contains("Header too long (26 characters, max 20 characters)"),
        "{err}"
    );
}

/// AC-2 (REQ-1): a subject that starts with a multi-byte character gets Go's garbled suggestion
/// (the first byte is sliced off), kept byte for byte.
#[test]
fn prepush_multibyte_subject_suggestion_is_byte_exact() {
    let p = TempDir::new().unwrap();
    let inv = quality_gate::pre_push::PrePushInvocation {
        project_dir: p.path().to_path_buf(),
        env: pp_env(Some("1"), Some("warn"), Some("angular")),
        stdin: "한글 제목\n".as_bytes().to_vec(),
    };
    let (mut out, mut err) = (Vec::new(), Vec::new());
    quality_gate::pre_push::run(&inv, &mut out, &mut err);
    let want: Vec<u8> = [
        b"    Suggestion: chore: ".as_slice(),
        "\u{FFFD}".as_bytes(),
        &"한글".as_bytes()[1..],
        " 제목\n\n".as_bytes(),
    ]
    .concat();
    assert!(
        err.windows(want.len()).any(|w| w == want.as_slice()),
        "{:?}",
        String::from_utf8_lossy(&err)
    );
}

fn git_repo(subjects: &[&str]) -> TempDir {
    let p = TempDir::new().unwrap();
    git_ok(p.path(), &["init", "-q", "."]);
    for s in subjects {
        git_ok(
            p.path(),
            &[
                "-c",
                "user.name=t",
                "-c",
                "user.email=t@t",
                "commit",
                "-q",
                "--allow-empty",
                "-m",
                s,
            ],
        );
    }
    p
}

/// AC-2 (REQ-4): `auto` follows the history when it scores at least the threshold, else the
/// fallback; a disabled detection or an unreadable history goes to the fallback too.
#[test]
fn prepush_auto_detection() {
    let name_of = |p: &Path, env: quality_gate::pre_push::PrePushEnv| {
        let (_, _, err) = check(p, env, "bad message\n");
        err.lines().next().unwrap_or("").to_string()
    };
    let history = git_repo(&["feat: a", "fix: b", "chore: c", "docs: d"]);
    assert!(name_of(history.path(), pp_env(Some("1"), None, None))
        .contains("violate conventional-commits convention"));
    // Free-form history scores 0 for every convention: the configured fallback is used.
    let free = git_repo(&["Update stuff", "More stuff", "Fixes", "wip"]);
    pp_config(
        free.path(),
        "git-convention.yaml",
        "git_convention:\n  auto_detection:\n    fallback: angular\n",
    );
    assert!(
        name_of(free.path(), pp_env(Some("1"), None, None)).contains("violate angular convention")
    );
    // Detection off: the fallback even when the history is conventional.
    pp_config(
        history.path(),
        "git-convention.yaml",
        "git_convention:\n  auto_detection:\n    enabled: false\n    fallback: karma\n",
    );
    assert!(
        name_of(history.path(), pp_env(Some("1"), None, None)).contains("violate karma convention")
    );
    // No repository at all: fallback.
    let none = TempDir::new().unwrap();
    assert!(name_of(none.path(), pp_env(Some("1"), None, None))
        .contains("violate conventional-commits convention"));
}

/// AC-2 (REQ-1): a value the Go loader refuses voids the whole config: defaults apply, the
/// switch can then come from the environment only, and with no profile the action is `enforce`.
#[test]
fn prepush_invalid_config_fails_closed() {
    let p = TempDir::new().unwrap();
    pp_config(
        p.path(),
        "git-convention.yaml",
        "git_convention:\n  validation:\n    max_length: -1\n",
    );
    let (code, _, err) = check(p.path(), pp_env(Some("1"), None, None), "bad message\n");
    assert_eq!(code, 2, "{err}");
    assert!(
        err.contains("violate conventional-commits convention"),
        "{err}"
    );
    // The file's own switch is lost with it.
    pp_config(
        p.path(),
        "git-convention.yaml",
        "git_convention:\n  convention: bogus\n  validation:\n    enforce_on_push: true\n",
    );
    assert_eq!(
        check(p.path(), pp_env(None, None, None), "bad message\n"),
        (0, String::new(), String::new())
    );
    // An unknown fallback is refused the same way, so it never reaches the load error.
    pp_config(
        p.path(),
        "git-convention.yaml",
        "git_convention:\n  auto_detection:\n    enabled: false\n    fallback: nope\n",
    );
    assert_eq!(check(p.path(), pp_env(Some("1"), None, None), "x\n").0, 2);
}

/// Audit B2: the `custom:` block the Go build writes into `git-convention.yaml` (empty lists, a
/// list of names) must not turn the check off.
#[test]
fn prepush_reads_the_go_written_git_convention_file() {
    let p = TempDir::new().unwrap();
    pp_config(
        p.path(),
        "git-convention.yaml",
        "git_convention:\n  convention: auto\n  validation:\n    enforce_on_push: true\n  custom:\n    name: \"\"\n    types: []\n    scopes: []\n    examples: []\n",
    );
    pp_config(
        p.path(),
        "git-strategy.yaml",
        "git_strategy:\n  mode: team\n  team:\n    hooks:\n      pre_push: enforce\n",
    );
    let (code, _, err) = check(p.path(), pp_env(None, None, None), "bad message\n");
    assert_eq!(code, 2, "{err}");
    assert!(!err.contains("YAML subset"), "{err}");
    pp_config(
        p.path(),
        "git-convention.yaml",
        "git_convention:\n  validation:\n    enforce_on_push: true\n  custom:\n    types:\n      - feat\n      - fix\n",
    );
    assert_eq!(
        check(p.path(), pp_env(None, None, None), "bad message\n").0,
        2
    );
}
