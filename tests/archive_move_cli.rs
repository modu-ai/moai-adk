//! D2 #2188 CLI wiring: one `archive-plan` → `archive-apply` → `archive-search` →
//! `archive-restore` round through the real binary, checking the envelope (operation,
//! ok, kind). The binary uses the real clock, so the fixture commit is dated far back.

mod common;
use serde_json::Value;
use std::{fs, path::Path, process::Command};
use tempfile::TempDir;

/// 2025-01-01T00:00:00Z — past the default 90-day window at any run time.
const OLD_SECS: u64 = 1_735_689_600;

fn git(dir: &Path, args: &[&str]) {
    let full = [
        [
            "-c",
            "user.name=fixture",
            "-c",
            "user.email=fixture@example.invalid",
            "-c",
            "commit.gpgsign=false",
        ]
        .as_slice(),
        args,
    ]
    .concat();
    common::git::git_pinned(dir, &full, Some(&OLD_SECS.to_string()));
}

fn cli(args: &[&str]) -> (i32, Value) {
    let out = Command::new(crate::common::MOAI_BIN)
        .args(args)
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "non-JSON stdout for {args:?}: {} / {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    (out.status.code().unwrap_or(-1), v)
}

#[test]
fn archive_ops_round_trip_through_the_cli() {
    let root = TempDir::new().unwrap();
    let project = root.path().join("project");
    let dir = project.join(".moai-cli/specs/SPEC-CLI");
    fs::create_dir_all(&dir).unwrap();
    fs::write(
        dir.join("spec.md"),
        "---\nstatus: completed\n---\n# SPEC-CLI\nunique-marker-2188\n",
    )
    .unwrap();
    git(&project, &["init", "-q", "-b", "main"]);
    git(&project, &["add", ".moai-cli/specs"]);
    git(&project, &["commit", "-q", "-m", "old"]);
    let p = project.to_str().unwrap();

    let (code, plan) = cli(&["archive-plan", p]);
    assert_eq!(code, 0, "{plan}");
    let digest = plan["result"]["plan_digest"].as_str().unwrap().to_string();
    assert!(digest.starts_with("sha256:"), "{digest}");

    // A stale digest is refused as stale_plan and names archive.plan as the next action.
    let (code, v) = cli(&[
        "archive-apply",
        p,
        "--plan-digest",
        "sha256:00",
        "--permission-reference",
        "decisions 00:5x",
    ]);
    assert_ne!(code, 0);
    assert_eq!(v["operation"], "archive.apply");
    assert_eq!(v["error"]["kind"], "stale_plan");
    assert_eq!(v["error"]["next_actions"][0], "archive.plan");

    let (code, v) = cli(&[
        "archive-apply",
        p,
        "--plan-digest",
        &digest,
        "--permission-reference",
        "decisions 00:5x",
    ]);
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["operation"], "archive.apply");
    assert_eq!(v["result"]["archived"][0]["spec_id"], "SPEC-CLI");
    assert!(!dir.exists());

    let (code, v) = cli(&["archive-search", p, "--query", "unique-marker-2188"]);
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["operation"], "archive.search");
    assert_eq!(v["result"]["hits"][0]["spec_id"], "SPEC-CLI");

    let (code, v) = cli(&[
        "archive-restore",
        p,
        "--spec",
        "SPEC-CLI",
        "--permission-reference",
        "decisions 00:5x",
    ]);
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["operation"], "archive.restore");
    assert!(dir.join("spec.md").is_file());
}
