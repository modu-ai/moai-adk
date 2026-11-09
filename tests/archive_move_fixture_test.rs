//! D2 #2188 archive move fixtures (REQ-D2-01) — AC-1: the success and failure
//! fixtures of `archive-apply` (hash before the move, copy, read back, index, then
//! remove the source; any mismatch cancels the move and keeps the original); AC-2:
//! the restore round trip; plus the non-vector archive search and its limits.

use moai_cli::archive::{
    apply, copy_verified_with, manifest, plan, restore, search, SEARCH_QUERY_CHARS_MAX,
    SEARCH_RESULTS_MAX,
};
use serde_json::Value;
use std::{
    fs,
    path::{Path, PathBuf},
    process::Command,
    time::{Duration, SystemTime, UNIX_EPOCH},
};
use tempfile::TempDir;

/// 2026-11-09T00:00:00Z — the pinned evaluation instant.
const NOW_SECS: u64 = 1_794_182_400;
/// 100 days before `now`: past the default 90-day cutoff.
const OLD_SECS: u64 = NOW_SECS - 100 * 86_400;
const PERMISSION: &str = "decisions 2026-11-09 00:5x";

fn now() -> SystemTime {
    UNIX_EPOCH + Duration::from_secs(NOW_SECS)
}

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

fn write_spec(project: &Path, id: &str, front: &str, body: &str) {
    let dir = project.join(".moai-cli/specs").join(id);
    fs::create_dir_all(dir.join("notes")).unwrap();
    fs::write(
        dir.join("spec.md"),
        format!("---\nname: {id}\n{front}---\n# {id}\n{body}\n"),
    )
    .unwrap();
    fs::write(dir.join("notes/plan.md"), format!("plan of {id}\n")).unwrap();
}

/// Two eligible SPECs (old, terminal) and one held SPEC, all committed at OLD_SECS.
fn fixture() -> (TempDir, PathBuf) {
    let root = TempDir::new().unwrap();
    let project = root.path().join("project");
    fs::create_dir_all(&project).unwrap();
    write_spec(
        &project,
        "SPEC-A",
        "status: completed\n",
        "Alpha retention body",
    );
    write_spec(
        &project,
        "SPEC-B",
        "status: superseded\n",
        "Bravo body line",
    );
    write_spec(
        &project,
        "SPEC-HOLD",
        "status: completed\nhold: true\n",
        "held",
    );
    git(&project, &["init", "-q", "-b", "main"], None);
    git(&project, &["add", ".moai-cli/specs"], None);
    git(
        &project,
        &["commit", "-q", "-m", "old specs"],
        Some(OLD_SECS.to_string()),
    );
    (root, project)
}

fn digest_of(project: &Path) -> String {
    plan(project, now(), None).unwrap().plan_digest
}

fn specs(project: &Path) -> PathBuf {
    project.join(".moai-cli/specs")
}

fn archived(project: &Path) -> PathBuf {
    project.join(".moai-cli/archive/specs")
}

fn index(project: &Path) -> Value {
    serde_json::from_slice(&fs::read(project.join(".moai-cli/archive/index.json")).unwrap())
        .unwrap()
}

#[test]
fn apply_moves_eligible_specs_with_hash_readback_and_index() {
    let (_root, project) = fixture();
    let before_a = manifest(&specs(&project).join("SPEC-A")).unwrap();
    let digest = digest_of(&project);

    let out = apply(&project, now(), None, &digest, PERMISSION).unwrap();
    assert_eq!(out["state"], "applied");
    let ids: Vec<&str> = out["archived"]
        .as_array()
        .unwrap()
        .iter()
        .map(|r| r["spec_id"].as_str().unwrap())
        .collect();
    assert_eq!(ids, ["SPEC-A", "SPEC-B"]);

    // The originals are gone, the held SPEC stays put.
    assert!(!specs(&project).join("SPEC-A").exists());
    assert!(!specs(&project).join("SPEC-B").exists());
    assert!(specs(&project).join("SPEC-HOLD/spec.md").is_file());

    // The archived copy carries the same bytes as the pre-move hash.
    let after_a = manifest(&archived(&project).join("SPEC-A")).unwrap();
    assert_eq!(before_a, after_a);
    assert!(archived(&project).join("SPEC-A/notes/plan.md").is_file());

    // The index is the pointer: one entry per archived SPEC with its manifest digest.
    let idx = index(&project);
    let entries = idx["entries"].as_array().unwrap();
    assert_eq!(entries.len(), 2);
    assert_eq!(entries[0]["spec_id"], "SPEC-A");
    assert_eq!(entries[0]["status"], "completed");
    assert_eq!(entries[0]["file_count"], 2);
    assert_eq!(entries[0]["manifest_digest"], before_a.digest());
    assert_eq!(out["archived"][0]["manifest_digest"], before_a.digest());

    // A second plan sees nothing left to move.
    assert!(plan(&project, now(), None).unwrap().eligible.is_empty());
}

#[test]
fn readback_mismatch_cancels_the_move_and_keeps_the_original() {
    let (_root, project) = fixture();
    let src = specs(&project).join("SPEC-A");
    let dst = archived(&project).join("SPEC-A");
    let expected = manifest(&src).unwrap();

    // The staged copy is corrupted after the copy and before the readback: only the
    // readback comparison can notice it (the source itself is unchanged).
    let err = copy_verified_with(&src, &dst, &expected, |staged: &Path| {
        fs::write(staged.join("spec.md"), b"corrupted").unwrap();
    })
    .unwrap_err();
    assert!(err.contains("readback digest mismatch"), "{err}");
    assert_eq!(
        moai_cli::contract::classify(&err).as_str(),
        "conflict",
        "{err}"
    );
    assert!(!dst.exists(), "the corrupted copy must not be published");
    assert_eq!(manifest(&src).unwrap(), expected, "the original is kept");
    // No staging directory is left behind.
    let leftovers: Vec<_> = fs::read_dir(dst.parent().unwrap())
        .map(|d| d.filter_map(|e| e.ok()).collect())
        .unwrap_or_default();
    assert!(leftovers.is_empty(), "{leftovers:?}");
}

#[test]
fn source_change_during_the_copy_cancels_the_move() {
    let (_root, project) = fixture();
    let src = specs(&project).join("SPEC-A");
    let dst = archived(&project).join("SPEC-A");
    let expected = manifest(&src).unwrap();
    let err = copy_verified_with(&src, &dst, &expected, |_staged: &Path| {
        fs::write(src.join("late.md"), b"written mid-move").unwrap();
    })
    .unwrap_err();
    assert!(err.contains("changed during apply"), "{err}");
    assert!(!dst.exists());
    assert!(src.join("late.md").is_file());
    assert!(src.join("spec.md").is_file());
}

#[test]
fn apply_refuses_a_stale_plan_digest_and_moves_nothing() {
    let (_root, project) = fixture();
    let err = apply(&project, now(), None, "sha256:00", PERMISSION).unwrap_err();
    assert!(err.contains("stale or altered plan"), "{err}");
    assert!(specs(&project).join("SPEC-A/spec.md").is_file());
    assert!(!project.join(".moai-cli/archive").exists());
}

#[test]
fn apply_refuses_a_missing_permission_reference() {
    let (_root, project) = fixture();
    let digest = digest_of(&project);
    for bad in ["", "  ", "null", "placeholder"] {
        let err = apply(&project, now(), None, &digest, bad).unwrap_err();
        assert_eq!(err, "permission reference required");
    }
    assert!(specs(&project).join("SPEC-A/spec.md").is_file());
}

#[test]
fn apply_refuses_an_existing_destination_and_keeps_the_original() {
    let (_root, project) = fixture();
    fs::create_dir_all(archived(&project).join("SPEC-A")).unwrap();
    fs::write(archived(&project).join("SPEC-A/other.md"), b"someone else").unwrap();
    let digest = digest_of(&project);
    let err = apply(&project, now(), None, &digest, PERMISSION).unwrap_err();
    assert!(err.contains("exclusive create refused"), "{err}");
    assert!(specs(&project).join("SPEC-A/spec.md").is_file());
    assert_eq!(
        fs::read(archived(&project).join("SPEC-A/other.md")).unwrap(),
        b"someone else"
    );
}

#[cfg(unix)]
#[test]
fn manifest_refuses_a_symlink_inside_a_spec_directory() {
    let (_root, project) = fixture();
    let src = specs(&project).join("SPEC-A");
    std::os::unix::fs::symlink("/etc/hosts", src.join("link.md")).unwrap();
    let err = manifest(&src).unwrap_err();
    assert!(err.contains("symlink"), "{err}");
    assert!(err.contains("refused"), "{err}");
}

#[test]
fn restore_round_trips_the_archived_spec() {
    let (_root, project) = fixture();
    let original = manifest(&specs(&project).join("SPEC-A")).unwrap();
    let digest = digest_of(&project);
    apply(&project, now(), None, &digest, PERMISSION).unwrap();

    let out = restore(&project, "SPEC-A", PERMISSION).unwrap();
    assert_eq!(out["state"], "restored");
    assert_eq!(out["spec_id"], "SPEC-A");
    assert_eq!(
        manifest(&specs(&project).join("SPEC-A")).unwrap(),
        original
    );
    assert!(!archived(&project).join("SPEC-A").exists());
    let idx = index(&project);
    let left: Vec<&str> = idx["entries"]
        .as_array()
        .unwrap()
        .iter()
        .map(|e| e["spec_id"].as_str().unwrap())
        .collect();
    assert_eq!(left, ["SPEC-B"]);
}

#[test]
fn restore_refuses_an_existing_destination_and_unknown_ids() {
    let (_root, project) = fixture();
    let digest = digest_of(&project);
    apply(&project, now(), None, &digest, PERMISSION).unwrap();

    fs::create_dir_all(specs(&project).join("SPEC-A")).unwrap();
    let err = restore(&project, "SPEC-A", PERMISSION).unwrap_err();
    assert!(err.contains("exclusive create refused"), "{err}");
    assert!(archived(&project).join("SPEC-A/spec.md").is_file());

    let err = restore(&project, "SPEC-NOPE", PERMISSION).unwrap_err();
    assert_eq!(moai_cli::contract::classify(&err).as_str(), "not_found");
    let err = restore(&project, "../SPEC-B", PERMISSION).unwrap_err();
    assert_eq!(moai_cli::contract::classify(&err).as_str(), "invalid_input");
    let err = restore(&project, "SPEC-B", "").unwrap_err();
    assert_eq!(err, "permission reference required");
}

#[test]
fn restore_refuses_a_tampered_archive_copy() {
    let (_root, project) = fixture();
    let digest = digest_of(&project);
    apply(&project, now(), None, &digest, PERMISSION).unwrap();
    fs::write(archived(&project).join("SPEC-A/spec.md"), b"tampered").unwrap();
    let err = restore(&project, "SPEC-A", PERMISSION).unwrap_err();
    assert!(err.contains("digest mismatch"), "{err}");
    assert!(!specs(&project).join("SPEC-A").exists());
    assert!(archived(&project).join("SPEC-A/spec.md").is_file());
}

#[test]
fn search_finds_archived_text_case_insensitively() {
    let (_root, project) = fixture();
    let digest = digest_of(&project);
    apply(&project, now(), None, &digest, PERMISSION).unwrap();

    let out = search(&project, "RETENTION").unwrap();
    assert_eq!(out["writes"], "none");
    assert_eq!(out["truncated"], false);
    let hits = out["hits"].as_array().unwrap();
    assert_eq!(hits.len(), 1, "{hits:?}");
    assert_eq!(hits[0]["spec_id"], "SPEC-A");
    assert_eq!(hits[0]["path"], "spec.md");
    assert_eq!(hits[0]["line"], 5);

    // The SPEC id itself is searchable through the index.
    let out = search(&project, "spec-b").unwrap();
    assert!(out["hits"]
        .as_array()
        .unwrap()
        .iter()
        .any(|h| h["spec_id"] == "SPEC-B" && h["source"] == "index"));

    // Not-archived SPECs are not searched.
    let out = search(&project, "held").unwrap();
    assert!(out["hits"].as_array().unwrap().is_empty());
}

#[test]
fn search_query_length_is_bounded_at_the_limit() {
    let (_root, project) = fixture();
    let ok = "q".repeat(SEARCH_QUERY_CHARS_MAX);
    assert!(search(&project, &ok).is_ok());
    let long = "q".repeat(SEARCH_QUERY_CHARS_MAX + 1);
    let err = search(&project, &long).unwrap_err();
    assert!(err.contains("must be"), "{err}");
    let err = search(&project, "").unwrap_err();
    assert!(err.contains("must be"), "{err}");
    // Characters, not bytes: a multi-byte query at the limit is accepted.
    assert!(search(&project, &"한".repeat(SEARCH_QUERY_CHARS_MAX)).is_ok());
}

#[test]
fn search_results_are_capped_at_the_limit() {
    let (_root, project) = fixture();
    // SPEC-A carries exactly the cap of matching lines; SPEC-B adds one more later.
    let lines = |n: usize| -> String { (0..n).map(|i| format!("needle {i}\n")).collect() };
    fs::write(
        specs(&project).join("SPEC-A/notes/plan.md"),
        lines(SEARCH_RESULTS_MAX),
    )
    .unwrap();
    git(&project, &["add", ".moai-cli/specs"], None);
    git(
        &project,
        &["commit", "-q", "-m", "needles"],
        Some(OLD_SECS.to_string()),
    );
    let digest = digest_of(&project);
    apply(&project, now(), None, &digest, PERMISSION).unwrap();

    let out = search(&project, "needle").unwrap();
    assert_eq!(out["hits"].as_array().unwrap().len(), SEARCH_RESULTS_MAX);
    assert_eq!(out["truncated"], false);

    fs::write(archived(&project).join("SPEC-B/needle.md"), "needle\n").unwrap();
    let out = search(&project, "needle").unwrap();
    assert_eq!(out["hits"].as_array().unwrap().len(), SEARCH_RESULTS_MAX);
    assert_eq!(out["truncated"], true);
}
