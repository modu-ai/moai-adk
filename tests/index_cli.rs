//! #2182 (rc.2 C6) incremental index. The fixtures drive `index-build` / `index-query`
//! against a temporary git repository, so the freshness suite (AC-1: additions,
//! deletions, modifications, new files, branch switch), REQ-C6-01 (a failed build keeps
//! the previous generation and marks partial) and AC-2 (mixed generations 0 — the
//! pointer is the only resolution path, atomic swap, orphan temp dirs ignored) are all
//! exercised in isolation. No watcher: every case is command-driven.
use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};

fn git(dir: &Path, args: &[&str]) {
    let status = Command::new("git")
        .args(args)
        .current_dir(dir)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("GIT_AUTHOR_NAME", "t")
        .env("GIT_AUTHOR_EMAIL", "t@example.com")
        .env("GIT_COMMITTER_NAME", "t")
        .env("GIT_COMMITTER_EMAIL", "t@example.com")
        .status()
        .unwrap();
    assert!(status.success(), "git {args:?} failed");
}

fn setup_project() -> (tempfile::TempDir, PathBuf) {
    let tmp = tempfile::tempdir().unwrap();
    let project = tmp.path().join("project");
    fs::create_dir_all(&project).unwrap();
    git(&project, &["init", "-q", "--initial-branch=main"]);
    fs::write(project.join("a.txt"), b"alpha").unwrap();
    git(&project, &["add", "."]);
    git(&project, &["commit", "-q", "-m", "init"]);
    (tmp, project)
}

fn cli(project: &Path, args: &[&str], stdin: &str) -> std::process::Output {
    let mut child = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .args(args)
        .current_dir(project)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("GIT_AUTHOR_NAME", "t")
        .env("GIT_AUTHOR_EMAIL", "t@example.com")
        .env("GIT_COMMITTER_NAME", "t")
        .env("GIT_COMMITTER_EMAIL", "t@example.com")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    child
        .stdin
        .take()
        .unwrap()
        .write_all(stdin.as_bytes())
        .unwrap();
    child.wait_with_output().unwrap()
}

fn envelope(out: &std::process::Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON: {e}\n{}",
            String::from_utf8_lossy(&out.stdout)
        )
    })
}

fn build(project: &Path) -> std::process::Output {
    cli(project, &["index-build", "."], r#"{"protocol_version": 1}"#)
}

fn query(project: &Path) -> Value {
    let out = cli(project, &["index-query", "."], r#"{"protocol_version": 1}"#);
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    envelope(&out)["result"].clone()
}

fn pointer_gen(project: &Path) -> Option<String> {
    let bytes = fs::read(project.join(".moai-cli/index/current.json")).ok()?;
    let v: Value = serde_json::from_slice(&bytes).ok()?;
    v["generation_id"].as_str().map(str::to_string)
}

/// AC-1 freshness suite: unavailable → build → fresh → each tracked change (add,
/// delete, modify), untracked new file, branch switch — each detected with the right
/// category, no watcher involved.
#[test]
fn freshness_suite_detects_add_delete_modify_new_and_branch_switch() {
    let (_tmp, project) = setup_project();

    // Before any build the state is unavailable with a reason.
    let v = query(&project);
    assert_eq!(v["state"], json!("unavailable"));
    assert!(v["reason"].as_str().unwrap().contains("no generation"));
    assert_eq!(v["changes"], Value::Null);

    // First build: fresh, the pointer names the generation, and the same content
    // rebuild dedupes into the same immutable generation id.
    let out = build(&project);
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let v = envelope(&out)["result"].clone();
    assert_eq!(v["status"], json!("complete"));
    assert_eq!(v["action"], json!("created"));
    let gen1 = v["generation_id"].as_str().unwrap().to_string();
    let out = build(&project);
    assert_eq!(envelope(&out)["result"]["action"], json!("unchanged"));
    assert_eq!(envelope(&out)["result"]["generation_id"], json!(gen1));
    let v = query(&project);
    assert_eq!(v["state"], json!("fresh"));
    assert_eq!(v["generation_id"], json!(gen1));
    assert_eq!(v["changes"]["added"], json!([]));
    assert_eq!(v["changes"]["branch_switched"], json!(false));

    // Addition: a new tracked file is detected under "added".
    fs::write(project.join("b.txt"), b"beta").unwrap();
    git(&project, &["add", "b.txt"]);
    let v = query(&project);
    assert_eq!(v["state"], json!("stale"));
    assert_eq!(v["changes"]["added"], json!(["b.txt"]));

    // Publish the addition: fresh again on a two-file generation.
    let out = build(&project);
    let gen2 = envelope(&out)["result"]["generation_id"]
        .as_str()
        .unwrap()
        .to_string();
    assert_ne!(gen1, gen2);
    assert_eq!(query(&project)["state"], json!("fresh"));

    // Deletion: removing a tracked file is detected under "deleted".
    fs::remove_file(project.join("a.txt")).unwrap();
    git(&project, &["rm", "--cached", "a.txt"]);
    let v = query(&project);
    assert_eq!(v["state"], json!("stale"));
    assert!(v["changes"]["deleted"]
        .as_array()
        .unwrap()
        .contains(&json!("a.txt")));

    // Publish the deletion so the next delta is isolated.
    let out = build(&project);
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    assert_eq!(query(&project)["state"], json!("fresh"));

    // Modification: changed bytes of an indexed file are detected under "modified".
    fs::write(project.join("b.txt"), b"beta-2").unwrap();
    let v = query(&project);
    assert_eq!(v["state"], json!("stale"));
    assert_eq!(v["changes"]["modified"], json!(["b.txt"]));

    // New (untracked) file: reported under "new_files", still not part of the index.
    fs::write(project.join("scratch.txt"), b"untracked").unwrap();
    let v = query(&project);
    assert_eq!(v["changes"]["new_files"], json!(["scratch.txt"]));

    // Branch switch: committing moves HEAD past the manifest's base — the switch is
    // named even when no tracked delta remains.
    git(&project, &["add", "-A"]);
    git(&project, &["commit", "-q", "-m", "second"]);
    let v = query(&project);
    assert_eq!(v["state"], json!("stale"));
    assert_eq!(v["changes"]["branch_switched"], json!(true));
    assert_eq!(v["manifest_base_commit"], json!(gen1_base(&project, &gen1)));

    // Rebuild on the new head: fresh again, a second immutable generation exists
    // beside the first (never deleted — cache loss is a performance loss only).
    let out = build(&project);
    let gen2 = envelope(&out)["result"]["generation_id"]
        .as_str()
        .unwrap()
        .to_string();
    assert_ne!(gen1, gen2);
    assert!(project
        .join(".moai-cli/index/generations")
        .join(&gen1)
        .is_dir());
    let v = query(&project);
    assert_eq!(v["state"], json!("fresh"));
    assert_eq!(v["generation_id"], json!(gen2));
    assert_eq!(v["changes"]["branch_switched"], json!(false));
}

fn gen1_base(project: &Path, gen: &str) -> String {
    let bytes = fs::read(
        project
            .join(".moai-cli/index/generations")
            .join(gen)
            .join("manifest.json"),
    )
    .unwrap();
    let v: Value = serde_json::from_slice(&bytes).unwrap();
    v["base_commit"].as_str().unwrap().to_string()
}

/// REQ-C6-01: a build that cannot finish keeps the previous generation (the pointer is
/// never moved) and marks the attempt, so queries report `partial` with the reason;
/// the next successful build clears the marker.
#[test]
fn failed_build_keeps_previous_generation_and_marks_partial() {
    let (_tmp, project) = setup_project();
    let out = build(&project);
    let gen1 = envelope(&out)["result"]["generation_id"]
        .as_str()
        .unwrap()
        .to_string();

    // Make a tracked file unreadable so the scan cannot finish.
    let victim = project.join("a.txt");
    let mut perm = fs::metadata(&victim).unwrap().permissions();
    use std::os::unix::fs::PermissionsExt;
    perm.set_mode(0o000);
    fs::set_permissions(&victim, perm).unwrap();

    let out = build(&project);
    assert!(!out.status.success(), "the failing build must not succeed");
    let err = envelope(&out)["error"].clone();
    assert!(
        err["message"]
            .as_str()
            .unwrap()
            .contains("index build failed; previous generation kept"),
        "{}",
        err["message"]
    );

    // The pointer still names gen1 and the query reports partial with the reason.
    assert_eq!(pointer_gen(&project).as_deref(), Some(gen1.as_str()));
    let v = query(&project);
    assert_eq!(v["state"], json!("partial"));
    assert!(v["reason"].as_str().unwrap().contains("read failed"));
    assert_eq!(v["generation_id"], json!(gen1));

    // Repair the permission, rebuild: complete, fresh, marker cleared.
    let mut perm = fs::metadata(&victim).unwrap().permissions();
    perm.set_mode(0o644);
    fs::set_permissions(&victim, perm).unwrap();
    let out = build(&project);
    assert!(out.status.success());
    let v = query(&project);
    assert_eq!(v["state"], json!("fresh"));
    assert!(!project.join(".moai-cli/index/last-build.json").exists());
}

/// AC-2 mixed generations 0 + crash-mid-write: the pointer is the only resolution path.
/// A leftover `.tmp-*` staging directory (a crashed build's orphan candidate) is listed
/// but never resolved, and a corrupted pointer reports unavailable instead of guessing.
#[test]
fn orphan_temp_dirs_are_ignored_and_the_pointer_is_the_only_path() {
    let (_tmp, project) = setup_project();
    let out = build(&project);
    let gen1 = envelope(&out)["result"]["generation_id"]
        .as_str()
        .unwrap()
        .to_string();

    // Crash mid-write: a staging directory with a manifest that is NOT the published
    // generation. It must be reported as an orphan candidate and never resolved.
    let orphan = project
        .join(".moai-cli/index/generations")
        .join(".tmp-999-crashed");
    fs::create_dir_all(&orphan).unwrap();
    fs::write(
        orphan.join("manifest.json"),
        br#"{"schema_version": 1, "generation_id": "gen-not-published", "base_commit": "x", "files": [], "built_at": "2026-11-10T00:00:00Z"}"#,
    )
    .unwrap();
    let v = query(&project);
    assert_eq!(
        v["state"],
        json!("fresh"),
        "the orphan must not be resolved"
    );
    assert_eq!(v["generation_id"], json!(gen1));
    assert_eq!(v["orphan_candidates"], json!([".tmp-999-crashed"]));

    // A later build lists the orphan but still lands on its own generation.
    fs::write(project.join("b.txt"), b"beta").unwrap();
    git(&project, &["add", "b.txt"]);
    let out = build(&project);
    let gen2 = envelope(&out)["result"]["generation_id"]
        .as_str()
        .unwrap()
        .to_string();
    assert_ne!(gen1, gen2);
    assert!(orphan.is_dir(), "orphans are kept, not silently deleted");
    let v = query(&project);
    assert_eq!(v["generation_id"], json!(gen2));
    assert_eq!(v["orphan_candidates"], json!([".tmp-999-crashed"]));

    // A corrupted pointer (torn/damaged file) is unavailable with a reason — the
    // atomic swap makes this impossible mid-build, so seeing it means damage.
    fs::write(project.join(".moai-cli/index/current.json"), b"{truncated").unwrap();
    let v = query(&project);
    assert_eq!(v["state"], json!("unavailable"));
    assert!(v["reason"].as_str().unwrap().contains("unreadable"));
}
