//! #2906 S2: `gate-schedule-plan` renders a launchd or cron entry for `gate-loop --once`
//! and writes nothing; `gate-schedule-apply` writes that one file (plus its receipt) only
//! for a fresh digest and a permission reference, and never runs `launchctl` or `crontab`.
//! Every run uses a temporary HOME, `GIT_CONFIG_GLOBAL=/dev/null` and
//! `GIT_CONFIG_SYSTEM=/dev/null`; the real HOME and the real schedulers are never touched.
//!
//! The goldens are `tests/fixtures/schedule/*`, with the temporary HOME as `<HOME>`. Regenerate
//! them with `MOAI_SCHEDULE_GOLDEN_WRITE=1 cargo test --test gate_schedule_cli` and review the
//! diff.
#![cfg(unix)]
mod common;

use common::CleanEnv;
use serde_json::Value;
use std::collections::BTreeMap;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Command;

const PERM: &str = "test-permission-2906";
const FIXED: [&str; 8] = [
    "--repo",
    "acme/widgets",
    "--moai-bin",
    "/opt/moai/bin/moai",
    "--dir",
    "/work/widgets",
    "--path",
    "/usr/local/bin:/usr/bin:/bin",
];

fn temp_home() -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::TempDir::new().unwrap();
    let home = fs::canonicalize(dir.path()).unwrap();
    (dir, home)
}

/// Runs `moai <op> <home> --os <os> <fixed inputs> <extra>`; returns the exit code and the
/// envelope. `bin_first` goes first on the process PATH (fake schedulers).
fn run_with(
    home: &Path,
    op: &str,
    os: &str,
    extra: &[&str],
    bin_first: Option<&Path>,
) -> (i32, Value) {
    let path = match bin_first {
        Some(b) => format!("{}:/usr/bin:/bin", b.display()),
        None => "/usr/bin:/bin".into(),
    };
    // A flag given in `extra` replaces the fixed one.
    let fixed: Vec<&str> = FIXED
        .chunks(2)
        .filter(|kv| !extra.contains(&kv[0]))
        .flatten()
        .copied()
        .collect();
    let out = Command::new(common::MOAI_BIN)
        .arg(op)
        .arg(home)
        .args(["--os", os])
        .args(fixed)
        .args(extra)
        .clean_env()
        .env("HOME", home)
        .env("PATH", path)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .current_dir(home)
        .output()
        .unwrap();
    let envelope: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "no envelope: {}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    (out.status.code().unwrap_or(-1), envelope)
}

fn run(home: &Path, op: &str, os: &str, extra: &[&str]) -> (i32, Value) {
    run_with(home, op, os, extra, None)
}

fn plan(home: &Path, os: &str, extra: &[&str]) -> Value {
    let (code, env) = run(home, "gate-schedule-plan", os, extra);
    assert_eq!(code, 0, "{env}");
    env["result"].clone()
}

fn apply(home: &Path, os: &str, digest: &str, perm: &str, extra: &[&str]) -> (i32, Value) {
    let mut args = vec!["--plan-digest", digest, "--permission-reference", perm];
    args.extend_from_slice(extra);
    run(home, "gate-schedule-apply", os, &args)
}

/// Every path under `root` with its kind and bytes (symlinks by target), sorted.
fn tree(root: &Path) -> BTreeMap<String, String> {
    let mut out = BTreeMap::new();
    let mut stack = vec![root.to_path_buf()];
    while let Some(dir) = stack.pop() {
        for e in fs::read_dir(&dir).unwrap() {
            let p = e.unwrap().path();
            let rel = p.strip_prefix(root).unwrap().display().to_string();
            let m = fs::symlink_metadata(&p).unwrap();
            let v = if m.file_type().is_symlink() {
                format!("link:{}", fs::read_link(&p).unwrap().display())
            } else if m.is_dir() {
                stack.push(p.clone());
                "dir".into()
            } else {
                format!("file:{}", moai_cli::digest(&fs::read(&p).unwrap()))
            };
            out.insert(rel, v);
        }
    }
    out
}

fn golden(name: &str, os: &str) {
    let (_d, home) = temp_home();
    let p = plan(&home, os, &[]);
    let entry = p["entry"].as_str().expect("entry text");
    let got = entry.replace(home.to_str().unwrap(), "<HOME>");
    // cargo runs an integration test in the package root; only this named file is read.
    let path = Path::new("tests/fixtures/schedule").join(name);
    if std::env::var_os("MOAI_SCHEDULE_GOLDEN_WRITE").is_some() {
        fs::write(&path, &got).unwrap();
    }
    assert_eq!(
        got,
        fs::read_to_string(&path).unwrap(),
        "{name} differs from the golden"
    );
}

/// AC1: the launchd entry (StartInterval 90, dry-run, PATH only, log under the state root).
#[test]
fn launchd_entry_matches_the_golden() {
    golden("gate-loop-macos.plist", "macos");
    let (_d, home) = temp_home();
    let p = plan(&home, "macos", &[]);
    assert_eq!(
        p["target_path"].as_str().unwrap(),
        home.join("Library/LaunchAgents/moai.gate-loop.acme.widgets.plist")
            .to_str()
            .unwrap()
    );
    assert_eq!(p["label"], "moai.gate-loop.acme.widgets");
    let enable = p["enable"].to_string();
    assert!(enable.contains("launchctl bootstrap"), "{enable}");
    assert!(
        p["disable"].to_string().contains("launchctl bootout"),
        "{p}"
    );
}

/// AC2: the cron entry (`*/2`, flock, cd, log under the state root).
#[test]
fn cron_entry_matches_the_golden() {
    golden("gate-loop-linux.cron", "linux");
    let (_d, home) = temp_home();
    let p = plan(&home, "linux", &[]);
    assert_eq!(
        p["target_path"].as_str().unwrap(),
        home.join(".moai-cli/schedule/gate-loop.crontab")
            .to_str()
            .unwrap()
    );
    let enable = p["enable"].to_string();
    assert!(enable.contains("crontab -"), "{enable}");
}

/// AC3: the plan writes nothing in HOME, for both systems and both modes.
#[test]
fn plan_writes_nothing() {
    let (_d, home) = temp_home();
    fs::create_dir_all(home.join("Library/LaunchAgents")).unwrap();
    fs::write(home.join("Library/LaunchAgents/other.plist"), "x").unwrap();
    let before = tree(&home);
    for os in ["macos", "linux"] {
        plan(&home, os, &[]);
        plan(
            &home,
            os,
            &["--merge", "--merge-permission-reference", PERM],
        );
    }
    assert_eq!(tree(&home), before);
}

/// AC4: no write without a permission reference or with a stale or altered digest; the same
/// digest from a fresh plan writes exactly the planned bytes.
#[test]
fn apply_needs_a_fresh_digest_and_a_permission_reference() {
    let (_d, home) = temp_home();
    let p = plan(&home, "macos", &[]);
    let digest = p["plan_digest"].as_str().unwrap().to_string();
    let target = PathBuf::from(p["target_path"].as_str().unwrap());
    let before = tree(&home);

    let (code, env) = apply(&home, "macos", &digest, "", &[]);
    assert_ne!(code, 0, "{env}");
    assert!(
        env.to_string().contains("permission reference required"),
        "{env}"
    );
    assert_eq!(tree(&home), before);

    let mut altered = digest.clone();
    altered.pop();
    altered.push(if digest.ends_with('0') { '1' } else { '0' });
    let (code, env) = apply(&home, "macos", &altered, PERM, &[]);
    assert_ne!(code, 0, "{env}");
    assert!(env.to_string().contains("re-preview required"), "{env}");
    assert_eq!(tree(&home), before);

    // A digest from other inputs (the merge form) is stale for the dry-run entry.
    let merge = plan(
        &home,
        "macos",
        &["--merge", "--merge-permission-reference", PERM],
    );
    let (code, env) = apply(
        &home,
        "macos",
        merge["plan_digest"].as_str().unwrap(),
        PERM,
        &[],
    );
    assert_ne!(code, 0, "{env}");
    assert_eq!(tree(&home), before);

    let (code, env) = apply(&home, "macos", &digest, PERM, &[]);
    assert_eq!(code, 0, "{env}");
    assert_eq!(env["result"]["state"], "applied", "{env}");
    assert_eq!(
        fs::read_to_string(&target).unwrap(),
        p["entry"].as_str().unwrap()
    );

    // The file now exists: the old digest is stale, a fresh plan is `unchanged`.
    let (code, env) = apply(&home, "macos", &digest, PERM, &[]);
    assert_ne!(code, 0, "{env}");
    let again = plan(&home, "macos", &[]);
    assert_eq!(again["action"], "unchanged", "{again}");
}

/// AC4: a replaced file's previous bytes stay in the receipt.
#[test]
fn apply_keeps_the_before_bytes() {
    for os in ["macos", "linux"] {
        let (_d, home) = temp_home();
        let first = plan(&home, os, &[]);
        let target = PathBuf::from(first["target_path"].as_str().unwrap());
        fs::create_dir_all(target.parent().unwrap()).unwrap();
        fs::write(&target, "previous bytes\n").unwrap();
        let p = plan(
            &home,
            os,
            &["--merge", "--merge-permission-reference", PERM],
        );
        assert_eq!(p["action"], "replace", "{p}");
        let (code, env) = apply(
            &home,
            os,
            p["plan_digest"].as_str().unwrap(),
            PERM,
            &["--merge", "--merge-permission-reference", PERM],
        );
        assert_eq!(code, 0, "{env}");
        let entry = fs::read_to_string(&target).unwrap();
        assert_eq!(entry, p["entry"].as_str().unwrap());
        assert!(entry.contains("--permission-reference"), "{entry}");
        assert!(!entry.contains("--dry-run"), "{entry}");
        let id = env["result"]["receipt"].as_str().unwrap();
        let dir = home.join(".moai-cli/schedule/receipts").join(id);
        assert_eq!(
            fs::read_to_string(dir.join("target.before")).unwrap(),
            "previous bytes\n"
        );
        let receipt: Value =
            serde_json::from_slice(&fs::read(dir.join("receipt.json")).unwrap()).unwrap();
        assert_eq!(receipt["status"], "committed", "{receipt}");
        assert_eq!(receipt["permission_reference"], PERM);
        assert_eq!(
            receipt["before_sha256"].as_str().unwrap(),
            moai_cli::digest(b"previous bytes\n")
        );
    }
}

/// AC5: plan and apply never start `launchctl` or `crontab`; they print the commands only.
#[test]
fn apply_never_runs_launchctl_or_crontab() {
    let (_d, home) = temp_home();
    let bin = home.join("fake-bin");
    let marks = home.join("marks");
    fs::create_dir_all(&bin).unwrap();
    fs::create_dir_all(&marks).unwrap();
    for tool in ["launchctl", "crontab"] {
        common::write_executable(
            &bin.join(tool),
            &format!("#!/bin/sh\ntouch '{}/{tool}-ran'\n", marks.display()),
        );
    }
    for os in ["macos", "linux"] {
        let (code, env) = run_with(&home, "gate-schedule-plan", os, &[], Some(&bin));
        assert_eq!(code, 0, "{env}");
        let digest = env["result"]["plan_digest"].as_str().unwrap().to_string();
        let (code, env) = run_with(
            &home,
            "gate-schedule-apply",
            os,
            &["--plan-digest", &digest, "--permission-reference", PERM],
            Some(&bin),
        );
        assert_eq!(code, 0, "{env}");
        assert!(env["result"]["enable"]
            .to_string()
            .contains(if os == "macos" {
                "launchctl"
            } else {
                "crontab"
            }));
    }
    assert_eq!(
        fs::read_dir(&marks).unwrap().count(),
        0,
        "a scheduler tool was run"
    );
}

/// AC6: no entry carries `--review-confirmed`; G7 stays a human stop.
#[test]
fn entry_never_carries_review_confirmed() {
    let (_d, home) = temp_home();
    for os in ["macos", "linux"] {
        for extra in [
            &[][..],
            &["--merge", "--merge-permission-reference", PERM][..],
        ] {
            let p = plan(&home, os, extra);
            let entry = p["entry"].as_str().unwrap();
            assert!(entry.contains("gate-loop"), "{entry}");
            assert!(entry.contains("--once"), "{entry}");
            assert!(!entry.contains("review-confirmed"), "{entry}");
            assert!(!entry.contains("review_confirmed"), "{entry}");
        }
    }
    // The default is the dry run; --merge needs its own permission reference.
    let p = plan(&home, "linux", &[]);
    assert!(p["entry"].as_str().unwrap().contains("--dry-run"));
    let (code, env) = run(&home, "gate-schedule-plan", "linux", &["--merge"]);
    assert_ne!(code, 0, "{env}");
}

/// AC7: a token-like variable in the entry's environment or PATH is refused; a plain
/// variable passes.
#[test]
fn entry_refuses_token_variables() {
    let (_d, home) = temp_home();
    let ok = plan(&home, "macos", &["--env", "GH_CONFIG_DIR=/work/gh"]);
    assert!(
        ok["entry"].as_str().unwrap().contains("GH_CONFIG_DIR"),
        "{ok}"
    );
    for os in ["macos", "linux"] {
        for var in [
            "GH_TOKEN=x",
            "GITHUB_TOKEN=x",
            "MOAI_AUDIT_APP_TOKEN=x",
            "OPENAI_API_KEY=x",
            "AWS_SECRET=x",
            "gh_token=x",
        ] {
            let (code, env) = run(&home, "gate-schedule-plan", os, &["--env", var]);
            assert_ne!(code, 0, "{var}: {env}");
            assert!(env.to_string().contains("token_variable"), "{var}: {env}");
        }
        // A token assignment smuggled into PATH.
        let (code, env) = run(
            &home,
            "gate-schedule-plan",
            os,
            &["--path", "GH_TOKEN=ghp_x:/usr/bin"],
        );
        assert_ne!(code, 0, "{env}");
        assert!(env.to_string().contains("token_variable"), "{env}");
    }
}

/// AC19: Windows is refused with a named reason, and nothing is written.
#[test]
fn windows_is_refused_with_a_named_reason() {
    let (_d, home) = temp_home();
    let before = tree(&home);
    let (code, env) = run(&home, "gate-schedule-plan", "windows", &[]);
    assert_ne!(code, 0, "{env}");
    assert!(env.to_string().contains("windows_not_supported"), "{env}");
    let (code, env) = apply(&home, "windows", "sha256:00", PERM, &[]);
    assert_ne!(code, 0, "{env}");
    assert!(env.to_string().contains("windows_not_supported"), "{env}");
    assert_eq!(tree(&home), before);
}

/// A symlinked target folder (or state folder) is refused, and nothing is written through it.
#[test]
fn symlinked_target_folder_is_refused() {
    // macOS: ~/Library/LaunchAgents is a link to another folder.
    let (_d, home) = temp_home();
    let elsewhere = home.join("elsewhere");
    fs::create_dir_all(&elsewhere).unwrap();
    fs::create_dir_all(home.join("Library")).unwrap();
    std::os::unix::fs::symlink(&elsewhere, home.join("Library/LaunchAgents")).unwrap();
    let (code, env) = run(&home, "gate-schedule-plan", "macos", &[]);
    assert_ne!(code, 0, "{env}");
    assert!(env.to_string().contains("symlink"), "{env}");
    let (code, env) = apply(&home, "macos", "sha256:00", PERM, &[]);
    assert_ne!(code, 0, "{env}");
    assert_eq!(fs::read_dir(&elsewhere).unwrap().count(), 0);

    // Linux: the state root's schedule folder is a link.
    let (_d2, home) = temp_home();
    let elsewhere = home.join("elsewhere");
    fs::create_dir_all(&elsewhere).unwrap();
    fs::create_dir_all(home.join(".moai-cli")).unwrap();
    std::os::unix::fs::symlink(&elsewhere, home.join(".moai-cli/schedule")).unwrap();
    let (code, env) = run(&home, "gate-schedule-plan", "linux", &[]);
    assert_ne!(code, 0, "{env}");
    assert!(env.to_string().contains("symlink"), "{env}");
    assert_eq!(fs::read_dir(&elsewhere).unwrap().count(), 0);
}
