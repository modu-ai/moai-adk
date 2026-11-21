//! REQ-6·REQ-7 of SPEC-20261120-commit-message-card-numbers (repair slices): the project
//! update's git-hook step offers the shipped hook's install in a git project and never
//! replaces a user's own hook, a non-git folder gets no such step, and the hook file itself
//! passes a clean message, refuses a card number, and skips when the installed moai is
//! missing or answers a usage error. The real HOME and a real project are never touched.

mod common;

use common::git::git_in;
use common::home::TempHome;
use serde_json::Value;
use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};

const HOOK_TEXT: &str = include_str!("../install/commit-msg");

/// A git project under the temp HOME, with the generation marker and a complete profile:
/// the install phase plans only for a project that holds one.
fn git_project(home: &TempHome, name: &str) -> PathBuf {
    let p = home.path().join("work").join(name);
    fs::create_dir_all(p.join(".moai")).unwrap();
    fs::write(
        p.join(".moai/moai.json"),
        r#"{"schema_version":1,"generation":"moai","created_by":"test","project_id":"hook-proj"}"#,
    )
    .unwrap();
    fs::write(
        p.join(".moai/project.yaml"),
        r#"{"schema_version":2,"project_id":"hook-proj","purpose":"Hook fixture","tools":["Rust"],"development_methods":["TDD"],"hosts":["claude"]}"#,
    )
    .unwrap();
    git_in(&p, &["init", "-q"]);
    p
}

/// The `git_hook` steps of a project-update plan (usually exactly one).
fn plan_steps(home: &TempHome, project: &Path) -> Vec<Value> {
    let p = project.to_str().unwrap();
    let out = home.command(&["project-update-plan", p]).output().unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "{} / {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    assert_eq!(v["ok"], true, "{v}");
    let mut found = Vec::new();
    for phase in v["result"]["phases"].as_array().unwrap() {
        for s in phase["steps"].as_array().unwrap() {
            if s["step"] == "git_hook" {
                found.push(s.clone());
            }
        }
    }
    found
}

/// `(exit code, stdout, stderr)` of the hook run on `message`, with `MOAI_BIN` pointing at
/// `moai` when given.
fn run_hook(hook: &Path, message: &Path, moai: Option<&Path>) -> (i32, String, String) {
    let mut cmd = std::process::Command::new(hook);
    cmd.arg(message);
    if let Some(m) = moai {
        cmd.env("MOAI_BIN", m);
    }
    let out = cmd.output().unwrap();
    (
        out.status.code().unwrap_or(-1),
        String::from_utf8_lossy(&out.stdout).into_owned(),
        String::from_utf8_lossy(&out.stderr).into_owned(),
    )
}

/// A stub `moai` that answers `code` with `stderr` on stderr.
fn stub_moai(home: &TempHome, name: &str, code: &str, stderr: &str) -> PathBuf {
    let stub = home.path().join(name);
    common::write_executable(
        &stub,
        &format!("#!/bin/sh\nprintf '%s\\n' '{stderr}' >&2\nexit {code}\n"),
    );
    stub
}

#[test]
fn the_git_hook_step_offers_the_install_and_then_stays_unchanged() {
    let home = TempHome::new();
    let p = git_project(&home, "hook-proj");
    let steps = plan_steps(&home, &p);
    assert_eq!(steps.len(), 1, "one git_hook step");
    assert_eq!(steps[0]["state"], "ready", "{steps:?}");
    assert_eq!(steps[0]["apply"]["argv"][1], "git-hook-apply", "{steps:?}");
    // The apply installs the shipped bytes, executable, into git's own hook folder.
    let out = home
        .command(&["git-hook-apply", "--dir", p.to_str().unwrap()])
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], true, "{}", String::from_utf8_lossy(&out.stderr));
    let hook = PathBuf::from(v["result"]["hook"].as_str().unwrap());
    assert_eq!(fs::read(&hook).unwrap(), HOOK_TEXT.as_bytes());
    let mode = fs::metadata(&hook).unwrap().permissions().mode();
    assert_eq!(mode & 0o111, 0o111, "the hook is executable");
    // A second plan is `unchanged`; the apply refuses it.
    let steps = plan_steps(&home, &p);
    assert_eq!(steps[0]["state"], "unchanged", "{steps:?}");
    let out = home
        .command(&["git-hook-apply", "--dir", p.to_str().unwrap()])
        .output()
        .unwrap();
    assert_eq!(out.status.code(), Some(1), "a second apply is refused");
}

#[test]
fn a_user_hook_is_never_replaced() {
    let home = TempHome::new();
    let p = git_project(&home, "hook-user");
    let hook = p.join(".git/hooks/commit-msg");
    fs::create_dir_all(hook.parent().unwrap()).unwrap();
    fs::write(&hook, "#!/bin/sh\n# the user's own hook\n").unwrap();
    let steps = plan_steps(&home, &p);
    assert_eq!(steps[0]["state"], "needs_decision", "{steps:?}");
    let out = home
        .command(&["git-hook-apply", "--dir", p.to_str().unwrap()])
        .output()
        .unwrap();
    assert_eq!(out.status.code(), Some(1), "the apply refuses");
    assert_eq!(
        fs::read(&hook).unwrap(),
        b"#!/bin/sh\n# the user's own hook\n".to_vec(),
        "the user's hook stays"
    );
}

#[test]
fn a_non_git_folder_has_no_git_hook_step() {
    let home = TempHome::new();
    let p = git_project(&home, "no-git");
    fs::remove_dir_all(p.join(".git")).unwrap();
    let steps = plan_steps(&home, &p);
    assert!(
        steps.is_empty(),
        "a non-git folder gets no git_hook step: {steps:?}"
    );
}

#[test]
fn the_hook_passes_clean_refuses_a_card_number_and_skips() {
    let home = TempHome::new();
    let p = git_project(&home, "hook-run");
    // The hook file, installed by hand in git's own hook folder (the same bytes the
    // command installs), with a stub moai named by MOAI_BIN.
    let hook = p.join(".git/hooks/commit-msg");
    fs::create_dir_all(hook.parent().unwrap()).unwrap();
    fs::write(&hook, HOOK_TEXT).unwrap();
    let mut perms = fs::metadata(&hook).unwrap().permissions();
    use std::os::unix::fs::PermissionsExt as _;
    perms.set_mode(perms.mode() | 0o111);
    fs::set_permissions(&hook, perms).unwrap();
    let clean = p.join("clean-msg.txt");
    fs::write(&clean, "fix: the wording, without a card number\n").unwrap();
    let dirty = p.join("dirty-msg.txt");
    fs::write(&dirty, "fix: wording after C-150\n").unwrap();

    // A passing scan passes the commit.
    let passing = stub_moai(&home, "stub-pass", "0", "");
    let (code, _, err) = run_hook(&hook, &clean, Some(&passing));
    assert_eq!(code, 0, "{err}");
    // A scan that finds a card number refuses with the findings.
    let refusing = stub_moai(&home, "stub-refuse", "1", "card_number at line 1");
    let (code, _, err) = run_hook(&hook, &dirty, Some(&refusing));
    assert_eq!(code, 1, "the hook refuses");
    assert!(err.contains("card_number"), "{err}");
    // A usage error — an older installed moai — skips with a warning.
    let older = stub_moai(&home, "stub-older", "2", "unexpected argument");
    let (code, _, err) = run_hook(&hook, &dirty, Some(&older));
    assert_eq!(code, 0, "an older install never blocks");
    assert!(err.contains("skipping"), "{err}");
    // A missing install skips with a warning.
    let (code, _, err) = run_hook(&hook, &dirty, Some(&p.join("no-such-moai")));
    assert_eq!(code, 0, "a missing install never blocks");
    assert!(err.contains("skipped"), "{err}");
}
