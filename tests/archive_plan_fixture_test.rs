//! D1 #2187 (REQ-D1-01) archive-plan detection fixtures — AC-1: at least three
//! eligible and three excluded SPEC directories, detected deterministically on a
//! temporary project; AC-2: the report's JSON shape. The pinned `now` keeps every
//! fixture deterministic. The procedure writes nothing: the test also asserts
//! `writes: "none"` in the plan.

use moai_cli::archive::plan;
use serde_json::Value;
use std::{
    fs,
    path::{Path, PathBuf},
    process::Command,
    time::{Duration, UNIX_EPOCH},
};
use tempfile::TempDir;

/// 2026-11-09T00:00:00Z — the pinned evaluation instant.
const NOW_SECS: u64 = 1_794_182_400;
/// 2026-08-01T00:00:00Z — 100 days before `now`: past the 90-day cutoff.
const OLD_SECS: u64 = NOW_SECS - 100 * 86_400;
/// 2026-10-30T00:00:00Z — 10 days before `now`: inside the grace window.
const RECENT_SECS: u64 = NOW_SECS - 10 * 86_400;

fn git(dir: &Path, args: &[&str], date: Option<String>) {
    let mut cmd = Command::new("git");
    cmd.arg("-C")
        .arg(dir)
        .args([
            "-c",
            "user.name=fixture",
            "-c",
            "user.email=fixture@example.invalid",
            "-c",
            "commit.gpgsign=false",
        ])
        .args(args);
    if let Some(d) = date {
        cmd.env("GIT_AUTHOR_DATE", format!("{d} +0000"))
            .env("GIT_COMMITTER_DATE", format!("{d} +0000"));
    }
    let out = cmd.output().unwrap();
    assert!(
        out.status.success(),
        "git {args:?} failed: {}",
        String::from_utf8_lossy(&out.stderr)
    );
}

fn write_spec(project: &Path, id: &str, front: &str) {
    let dir = project.join(".moai-cli/specs").join(id);
    fs::create_dir_all(&dir).unwrap();
    fs::write(
        dir.join("spec.md"),
        format!("---\nname: {id}\n{front}---\n# {id}\nbody\n"),
    )
    .unwrap();
}

/// Builds the fixture project: six SPEC directories committed at two dates, plus one
/// directory without spec.md (git does not track empty directories — it exists on disk
/// only, which is exactly what the ambiguous-status fixture needs).
fn fixture() -> (TempDir, PathBuf) {
    let root = TempDir::new().unwrap();
    let project = root.path().join("project");
    fs::create_dir_all(&project).unwrap();
    write_spec(&project, "SPEC-ELIG-1", "status: completed\n");
    write_spec(&project, "SPEC-ELIG-2", "status: superseded\n");
    write_spec(&project, "SPEC-ELIG-3", "status: rejected\n");
    write_spec(&project, "SPEC-HOLD-1", "status: completed\nhold: true\n");
    write_spec(
        &project,
        "SPEC-EVID-1",
        "status: completed\nevidence: pending\n",
    );
    write_spec(&project, "SPEC-GRACE-1", "status: completed\n");
    fs::create_dir_all(project.join(".moai-cli/specs/SPEC-AMBIG-1")).unwrap();

    git(&project, &["init", "-q", "-b", "main"], None);
    for id in [
        "SPEC-ELIG-1",
        "SPEC-ELIG-2",
        "SPEC-ELIG-3",
        "SPEC-HOLD-1",
        "SPEC-EVID-1",
    ] {
        git(&project, &["add", &format!(".moai-cli/specs/{id}")], None);
    }
    git(
        &project,
        &["commit", "-q", "-m", "old specs"],
        Some(OLD_SECS.to_string()),
    );
    git(&project, &["add", ".moai-cli/specs/SPEC-GRACE-1"], None);
    git(
        &project,
        &["commit", "-q", "-m", "recent spec"],
        Some(RECENT_SECS.to_string()),
    );
    (root, project)
}

fn plan_at(project: &Path, grace_request: Option<u64>) -> Value {
    serde_json::to_value(
        plan(
            project,
            UNIX_EPOCH + Duration::from_secs(NOW_SECS),
            grace_request,
        )
        .unwrap(),
    )
    .unwrap()
}

fn ids<'a>(rows: &'a [Value], key: &str) -> Vec<&'a str> {
    rows.iter()
        .map(|r| r[key].as_str().expect("spec id"))
        .collect()
}

#[test]
fn detects_the_fixture_eligible_and_excluded_sets() {
    let (_root, project) = fixture();
    let v = plan_at(&project, None);

    // Deterministic evaluation instant and policy defaults (AC-2: the report shape).
    assert_eq!(v["protocol_version"], 1);
    assert_eq!(
        v["grace_days"],
        moai_cli::archive::DEFAULT_ARCHIVE_GRACE_DAYS
    );
    assert_eq!(v["grace_source"], "default");
    assert_eq!(v["writes"], "none");
    assert_eq!(v["network"], "disabled");
    assert_eq!(v["scanned"], 7);
    assert!(v["cutoff"].as_str().is_some());
    assert!(v["now"].as_str().is_some());

    // AC-1: three eligible SPECs — the terminal ones with activity past the cutoff,
    // detected through the git activity pass.
    let eligible = v["eligible"].as_array().unwrap();
    assert_eq!(
        ids(eligible, "spec_id"),
        ["SPEC-ELIG-1", "SPEC-ELIG-2", "SPEC-ELIG-3"]
    );
    let expected_status = [
        ("SPEC-ELIG-1", "completed"),
        ("SPEC-ELIG-2", "superseded"),
        ("SPEC-ELIG-3", "rejected"),
    ];
    for row in eligible {
        assert!(row["last_activity"].as_str().is_some());
        assert_eq!(row["activity_source"], "git");
        let id = row["spec_id"].as_str().unwrap();
        let want = expected_status.iter().find(|(i, _)| *i == id).unwrap().1;
        assert_eq!(row["status"], want);
    }

    // AC-1: four excluded SPECs, each with its documented reason.
    let excluded = v["excluded"].as_array().unwrap();
    let hold = excluded
        .iter()
        .find(|r| r["spec_id"] == "SPEC-HOLD-1")
        .unwrap();
    assert_eq!(hold["reason"], "hold");
    let evid = excluded
        .iter()
        .find(|r| r["spec_id"] == "SPEC-EVID-1")
        .unwrap();
    assert_eq!(evid["reason"], "pending_evidence");
    let grace = excluded
        .iter()
        .find(|r| r["spec_id"] == "SPEC-GRACE-1")
        .unwrap();
    assert_eq!(grace["reason"], "in_grace_window");
    let ambig = excluded
        .iter()
        .find(|r| r["spec_id"] == "SPEC-AMBIG-1")
        .unwrap();
    assert_eq!(ambig["reason"], "ambiguous_status");
    assert_eq!(ambig["status"], Value::Null);
    assert_eq!(excluded.len(), 4);
}

/// AC-2 shape check plus the profile override: a profile `archive_grace_days: 5`
/// pulls the cutoff to `now − 5 days`, which makes the 10-day-old terminal SPEC
/// eligible — the value is profile-overridable (Issue #2187 범위).
#[test]
fn profile_grace_days_override_moves_the_cutoff() {
    let (_root, project) = fixture();
    fs::write(
        project.join(".moai-cli/project.yaml"),
        "{\"schema_version\":2,\"project_id\":\"demo\",\"purpose\":\"p\",\"tools\":[\"Rust\"],\"development_methods\":[\"TDD\"],\"hosts\":[\"claude\"],\"archive_grace_days\":5}",
    )
    .unwrap();
    let v = plan_at(&project, None);
    assert_eq!(v["grace_days"], 5);
    assert_eq!(v["grace_source"], "profile");
    let eligible = v["eligible"].as_array().unwrap();
    assert!(
        ids(eligible, "spec_id").contains(&"SPEC-GRACE-1"),
        "a 5-day window makes the 10-day-old terminal SPEC eligible"
    );
}

/// The CLI request overrides the profile (source reads `request`).
#[test]
fn request_grace_override_reports_request_source() {
    let (_root, project) = fixture();
    let v = plan_at(&project, Some(10));
    assert_eq!(v["grace_days"], 10);
    assert_eq!(v["grace_source"], "request");
}

#[test]
fn missing_specs_dir_is_a_clean_empty_plan() {
    let root = TempDir::new().unwrap();
    let project = root.path().join("empty");
    fs::create_dir_all(&project).unwrap();
    let v = plan_at(&project, None);
    assert_eq!(v["scanned"], 0);
    assert_eq!(v["eligible"], Value::Array(vec![]));
    assert_eq!(v["excluded"], Value::Array(vec![]));
}
