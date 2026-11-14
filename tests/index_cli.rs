//! #2182 (rc.2 C6) incremental index. The fixtures drive `index-build` / `index-query`
//! against a temporary git repository, so the freshness suite (AC-1: additions,
//! deletions, modifications, new files, branch switch), REQ-C6-01 (a failed build keeps
//! the previous generation and marks partial) and AC-2 (mixed generations 0 — the
//! pointer is the only resolution path, atomic swap, orphan temp dirs ignored) are all
//! exercised in isolation. No watcher: every case is command-driven.
mod common;
use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};

fn git_command(dir: &Path) -> Command {
    let mut command = Command::new("git");
    command
        .args(["-c", "gc.auto=0", "-c", "maintenance.auto=false"])
        .current_dir(dir)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("GIT_AUTHOR_NAME", "t")
        .env("GIT_AUTHOR_EMAIL", "t@example.com")
        .env("GIT_COMMITTER_NAME", "t")
        .env("GIT_COMMITTER_EMAIL", "t@example.com");
    command
}

fn git(dir: &Path, args: &[&str]) {
    let status = git_command(dir).args(args).status().unwrap();
    assert!(status.success(), "git {args:?} failed");
}

fn git_output(dir: &Path, args: &[&str], input: &[u8]) -> Vec<u8> {
    let mut child = git_command(dir)
        .args(args)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    child.stdin.take().unwrap().write_all(input).unwrap();
    let output = child.wait_with_output().unwrap();
    assert!(
        output.status.success(),
        "{}",
        String::from_utf8_lossy(&output.stderr)
    );
    output.stdout
}

// Index-only entries keep cardinality/output fixtures small without fabricating git output.
fn replace_index(project: &Path, paths: &[String]) {
    let hash = git_output(project, &["rev-parse", "HEAD:a.txt"], &[]);
    let hash = std::str::from_utf8(&hash).unwrap().trim();
    git(project, &["read-tree", "--empty"]);
    let mut entries = Vec::new();
    for path in paths {
        writeln!(entries, "100644 {hash}\t{path}").unwrap();
    }
    git_output(project, &["update-index", "--index-info"], &entries);
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

fn cli_command(project: &Path) -> Command {
    let mut command = Command::new(crate::common::MOAI_BIN);
    command
        .current_dir(project)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("GIT_AUTHOR_NAME", "t")
        .env("GIT_AUTHOR_EMAIL", "t@example.com")
        .env("GIT_COMMITTER_NAME", "t")
        .env("GIT_COMMITTER_EMAIL", "t@example.com");
    command
}

fn cli(project: &Path, args: &[&str], stdin: &str) -> std::process::Output {
    let mut child = cli_command(project)
        .args(args)
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

/// Only intentional special-file probes use a deadline. File-backed output cannot
/// deadlock a child whose diagnostic or manifest exceeds a pipe's capacity.
#[cfg(unix)]
fn cli_before_deadline(project: &Path, args: &[&str]) -> std::process::Output {
    use std::io::{Read, Seek};
    use std::time::{Duration, Instant};
    let mut stdout = tempfile::tempfile().unwrap();
    let mut stderr = tempfile::tempfile().unwrap();
    let mut child = cli_command(project)
        .args(args)
        .stdin(Stdio::piped())
        .stdout(stdout.try_clone().unwrap())
        .stderr(stderr.try_clone().unwrap())
        .spawn()
        .unwrap();
    child
        .stdin
        .take()
        .unwrap()
        .write_all(br#"{"protocol_version":1}"#)
        .unwrap();
    // #2511: a hang guard, not a speed limit — a blocking open would wait forever, so a
    // generous bound keeps the check and cannot fail a slow but correct run under load.
    let deadline = Instant::now() + Duration::from_secs(120);
    let status = loop {
        if let Some(status) = child.try_wait().unwrap() {
            break status;
        }
        if Instant::now() >= deadline {
            let _ = child.kill();
            let _ = child.wait();
            panic!("index CLI exceeded the special-file deadline: {args:?}");
        }
        std::thread::sleep(Duration::from_millis(10));
    };
    stdout.rewind().unwrap();
    stderr.rewind().unwrap();
    let mut output = std::process::Output {
        status,
        stdout: vec![],
        stderr: vec![],
    };
    stdout.read_to_end(&mut output.stdout).unwrap();
    stderr.read_to_end(&mut output.stderr).unwrap();
    output
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

/// The command left `src/main.rs` for `src/ops/index_query.rs` (#2834 R2c): its envelope
/// still names `index.query`, on an answer and on a refused request alike.
#[test]
fn index_query_answers_under_its_operation_name() {
    let (_tmp, project) = setup_project();
    let ok = cli(
        &project,
        &["index-query", "."],
        r#"{"protocol_version": 1}"#,
    );
    assert_eq!(envelope(&ok)["operation"], "index.query");
    let bad = cli(&project, &["index-query", "."], "nope");
    assert_eq!(bad.status.code(), Some(1));
    let v = envelope(&bad);
    assert_eq!(
        (v["operation"].clone(), v["error"]["kind"].clone()),
        (json!("index.query"), json!("invalid_input"))
    );
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

/// Audit F1: exercise the public build path across a whole-second boundary. A
/// timestamp mixed into the generation key must fail even at second precision.
#[test]
fn identical_rebuild_across_seconds_keeps_generation_and_manifest() {
    let (_tmp, project) = setup_project();
    let first = build(&project);
    assert!(first.status.success());
    let first = envelope(&first)["result"].clone();
    let generation = first["generation_id"].as_str().unwrap();
    let manifest = project
        .join(".moai-cli/index/generations")
        .join(generation)
        .join("manifest.json");
    let before = fs::read(&manifest).unwrap();

    std::thread::sleep(std::time::Duration::from_millis(1100));

    let second = build(&project);
    assert!(second.status.success());
    let second = envelope(&second)["result"].clone();
    assert_eq!(second["generation_id"], first["generation_id"]);
    assert_eq!(second["action"], json!("unchanged"));
    assert_eq!(pointer_gen(&project).as_deref(), Some(generation));
    assert_eq!(fs::read(manifest).unwrap(), before);
    assert_eq!(query(&project)["state"], json!("fresh"));
}

/// A pointer is one generation name, never a path. An actual external manifest
/// makes traversal observable instead of relying on a missing target to reject it.
#[test]
fn pointer_generation_id_is_validated_before_path_resolution() {
    let (tmp, project) = setup_project();
    let first = build(&project);
    assert!(first.status.success());
    let generation = pointer_gen(&project).unwrap();
    let manifest = project
        .join(".moai-cli/index/generations")
        .join(&generation)
        .join("manifest.json");
    let original: Value = serde_json::from_slice(&fs::read(manifest).unwrap()).unwrap();
    let outside = tmp.path().join("external-generation");
    fs::create_dir(&outside).unwrap();
    let ids = [
        "../../../../external-generation".to_string(),
        outside.to_str().unwrap().to_string(),
        "gen-0123456789abcdeF".to_string(),
        "gen-0123456789abcde".to_string(),
        "gen-0123456789abcdef0".to_string(),
        "gen-0123456789abcdeg".to_string(),
    ];
    for id in ids {
        let mut foreign = original.clone();
        foreign["generation_id"] = json!(id);
        let bytes = serde_json::to_vec(&foreign).unwrap();
        fs::write(outside.join("manifest.json"), &bytes).unwrap();
        let pointer =
            serde_json::to_vec(&json!({"schema_version": 1, "generation_id": id})).unwrap();
        fs::write(project.join(".moai-cli/index/current.json"), &pointer).unwrap();
        let result = query(&project);
        assert_eq!(result["state"], json!("unavailable"), "{id}");
        assert_eq!(result["reason"], json!("index pointer unreadable"), "{id}");
        assert_eq!(result["generation_id"], Value::Null);
        assert_eq!(result["writes"], json!(false));
        assert_eq!(fs::read(outside.join("manifest.json")).unwrap(), bytes);
        assert_eq!(
            fs::read(project.join(".moai-cli/index/current.json")).unwrap(),
            pointer
        );
    }
}

/// Reuse is a successful build only when the existing immutable manifest still
/// decodes and describes this snapshot. Refusal keeps the other published generation.
#[test]
fn damaged_reusable_manifest_keeps_previous_pointer_and_marks_partial() {
    let (_tmp, project) = setup_project();
    assert!(build(&project).status.success());
    let old = pointer_gen(&project).unwrap();
    let old_manifest = project
        .join(".moai-cli/index/generations")
        .join(&old)
        .join("manifest.json");
    let before = fs::read(&old_manifest).unwrap();
    let original: Value = serde_json::from_slice(&before).unwrap();
    fs::write(project.join("b.txt"), b"beta").unwrap();
    git(&project, &["add", "b.txt"]);
    assert!(build(&project).status.success());
    let published = pointer_gen(&project).unwrap();
    assert_ne!(old, published);
    git(&project, &["rm", "--cached", "b.txt"]);
    fs::remove_file(project.join("b.txt")).unwrap();

    let mut damaged = vec![b"{truncated".to_vec()];
    for (field, value) in [
        ("schema_version", json!(2)),
        ("generation_id", json!("gen-0000000000000000")),
        ("base_commit", json!("wrong-base")),
        ("files", json!([])),
    ] {
        let mut manifest = original.clone();
        manifest[field] = value;
        damaged.push(serde_json::to_vec(&manifest).unwrap());
    }
    for bytes in damaged {
        fs::write(&old_manifest, &bytes).unwrap();
        let attempt = build(&project);
        assert!(
            !attempt.status.success(),
            "damaged manifest must not be reused"
        );
        assert_eq!(pointer_gen(&project).as_deref(), Some(published.as_str()));
        assert_eq!(query(&project)["state"], json!("partial"));
        assert_eq!(fs::read(&old_manifest).unwrap(), bytes);
    }
    fs::write(&old_manifest, &before).unwrap();
    let repaired = build(&project);
    assert!(repaired.status.success());
    assert_eq!(envelope(&repaired)["result"]["action"], json!("unchanged"));
    assert_eq!(pointer_gen(&project).as_deref(), Some(old.as_str()));
    assert_eq!(query(&project)["state"], json!("fresh"));
}

/// REQ-C6-01: a build that cannot finish keeps the previous generation (the pointer is
/// never moved) and marks the attempt, so queries report `partial` with the reason;
/// the next successful build clears the marker.
/// The unreadable-file setup uses Unix permission bits; Windows runtime is not measured.
#[cfg(unix)]
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

#[cfg(unix)]
#[test]
fn source_file_to_symlink_transition_changes_freshness_and_generation() {
    let (_tmp, project) = setup_project();
    fs::write(project.join("a.txt"), b"target.txt").unwrap();
    fs::write(project.join("target.txt"), b"different target bytes").unwrap();
    git(&project, &["add", "a.txt", "target.txt"]);
    git(&project, &["commit", "-q", "-m", "file and symlink target"]);
    assert!(build(&project).status.success());
    let old = pointer_gen(&project).unwrap();
    fs::remove_file(project.join("a.txt")).unwrap();
    std::os::unix::fs::symlink("target.txt", project.join("a.txt")).unwrap();
    let result = query(&project);
    assert_eq!(result["state"], json!("stale"));
    assert_eq!(result["changes"]["modified"], json!(["a.txt"]));
    assert!(build(&project).status.success());
    assert_ne!(pointer_gen(&project).as_deref(), Some(old.as_str()));
    assert_eq!(query(&project)["state"], json!("fresh"));
}

#[cfg(unix)]
#[test]
fn symlinked_failure_marker_never_overwrites_source() {
    let (_tmp, project) = setup_project();
    assert!(build(&project).status.success());
    let old = pointer_gen(&project).unwrap();
    std::os::unix::fs::symlink(
        "../../a.txt",
        project.join(".moai-cli/index/last-build.json"),
    )
    .unwrap();
    fs::write(
        project.join("too-large.txt"),
        vec![b'x'; 16 * 1024 * 1024 + 1],
    )
    .unwrap();
    git(&project, &["add", "too-large.txt"]);
    let attempt = build(&project);
    assert!(!attempt.status.success());
    assert_eq!(fs::read(project.join("a.txt")).unwrap(), b"alpha");
    assert_eq!(pointer_gen(&project).as_deref(), Some(old.as_str()));
    assert!(
        fs::symlink_metadata(project.join(".moai-cli/index/last-build.json"))
            .unwrap()
            .file_type()
            .is_symlink()
    );
}

#[cfg(unix)]
#[test]
fn symlinked_index_ancestors_never_receive_build_writes() {
    for rel in [
        ".moai-cli",
        ".moai-cli/index",
        ".moai-cli/index/generations",
    ] {
        let (tmp, project) = setup_project();
        assert!(build(&project).status.success());
        let external = tmp.path().join("external");
        fs::create_dir(&external).unwrap();
        fs::write(external.join("sentinel"), b"keep").unwrap();
        let original = project.join(rel);
        fs::rename(&original, original.with_extension("kept")).unwrap();
        std::os::unix::fs::symlink(&external, &original).unwrap();
        let attempt = build(&project);
        assert!(!attempt.status.success(), "{rel}");
        assert_eq!(fs::read(external.join("sentinel")).unwrap(), b"keep");
        assert_eq!(fs::read_dir(&external).unwrap().count(), 1, "{rel}");
        assert_eq!(fs::read(project.join("a.txt")).unwrap(), b"alpha");
    }
}

#[cfg(unix)]
#[test]
fn tracked_parent_symlinks_are_deleted_without_indexing_outside_content() {
    for parent in ["dir", "dir/nested"] {
        let (tmp, project) = setup_project();
        fs::create_dir_all(project.join("dir/nested")).unwrap();
        fs::write(project.join("dir/nested/x.txt"), b"inside").unwrap();
        git(&project, &["add", "dir/nested/x.txt"]);
        assert!(build(&project).status.success());
        let old = pointer_gen(&project).unwrap();
        let outside = tmp.path().join("outside");
        fs::create_dir_all(outside.join("nested")).unwrap();
        fs::write(outside.join("x.txt"), b"OUTSIDE-SECRET-CONTENT").unwrap();
        fs::write(outside.join("nested/x.txt"), b"OUTSIDE-SECRET-CONTENT").unwrap();
        fs::rename(project.join(parent), tmp.path().join("kept-source")).unwrap();
        std::os::unix::fs::symlink(&outside, project.join(parent)).unwrap();
        let changed = query(&project);
        assert_eq!(changed["state"], json!("stale"));
        assert_eq!(changed["changes"]["deleted"], json!(["dir/nested/x.txt"]));
        assert_eq!(changed["changes"]["modified"], json!([]));
        let built = build(&project);
        assert!(
            built.status.success(),
            "{}",
            String::from_utf8_lossy(&built.stderr)
        );
        assert_eq!(envelope(&built)["result"]["files"], json!(1));
        let generation = pointer_gen(&project).unwrap();
        assert_ne!(generation, old);
        let manifest: Value = serde_json::from_slice(
            &fs::read(
                project
                    .join(".moai-cli/index/generations")
                    .join(generation)
                    .join("manifest.json"),
            )
            .unwrap(),
        )
        .unwrap();
        assert_eq!(
            manifest["files"]
                .as_array()
                .unwrap()
                .iter()
                .map(|f| f["path"].as_str().unwrap())
                .collect::<Vec<_>>(),
            vec!["a.txt"]
        );
        assert_eq!(
            fs::read(outside.join("x.txt")).unwrap(),
            b"OUTSIDE-SECRET-CONTENT"
        );
    }
}

#[test]
fn tracked_file_count_has_literal_n_and_n_plus_one_boundaries() {
    let (_tmp, project) = setup_project();
    let mut paths = vec!["a.txt".to_string()];
    paths.extend((1..16_384).map(|n| format!("missing-{n:05}")));
    replace_index(&project, &paths);
    let accepted = build(&project);
    assert!(
        accepted.status.success(),
        "{}",
        String::from_utf8_lossy(&accepted.stderr)
    );
    assert_eq!(envelope(&accepted)["result"]["files"], json!(1));
    let previous = pointer_gen(&project).unwrap();
    paths.push("missing-16384".into());
    replace_index(&project, &paths);
    let refused = build(&project);
    assert!(!refused.status.success());
    let error = envelope(&refused);
    assert_eq!(error["error"]["kind"], json!("invalid_input"));
    assert_eq!(error["error"]["retryable"], json!(false));
    assert_eq!(pointer_gen(&project).as_deref(), Some(previous.as_str()));
}

#[test]
fn tracked_file_size_has_literal_n_and_n_plus_one_boundaries() {
    let (_tmp, project) = setup_project();
    fs::File::create(project.join("a.txt"))
        .unwrap()
        .set_len(16_777_216)
        .unwrap();
    let accepted = build(&project);
    assert!(
        accepted.status.success(),
        "{}",
        String::from_utf8_lossy(&accepted.stderr)
    );
    assert_eq!(envelope(&accepted)["result"]["bytes"], json!(16_777_216));
    let previous = pointer_gen(&project).unwrap();
    fs::OpenOptions::new()
        .write(true)
        .open(project.join("a.txt"))
        .unwrap()
        .set_len(16_777_217)
        .unwrap();
    let refused = build(&project);
    assert!(!refused.status.success());
    assert_eq!(envelope(&refused)["ok"], json!(false));
    assert_eq!(pointer_gen(&project).as_deref(), Some(previous.as_str()));
    assert_eq!(query(&project)["state"], json!("partial"));
}

fn virtual_path(number: usize, length: usize) -> String {
    let mut path = format!("missing-{number:05}/");
    while length - path.len() > 200 {
        path.push_str(&"x".repeat(199));
        path.push('/');
    }
    path.push_str(&"x".repeat(length - path.len()));
    path
}

#[test]
fn git_output_has_literal_n_and_n_plus_one_boundaries() {
    let (_tmp, project) = setup_project();
    // Split padding so the literal 4 MiB stream stays below macOS PATH_MAX per path.
    let mut paths = vec![
        "a.txt".to_string(),
        virtual_path(1, 764),
        virtual_path(2, 764),
    ];
    paths.extend((3..8_192).map(|n| virtual_path(n, 511)));
    replace_index(&project, &paths);
    assert_eq!(
        git_output(&project, &["ls-files", "-z"], &[]).len(),
        4_194_304
    );
    let accepted = build(&project);
    assert!(
        accepted.status.success(),
        "{}",
        String::from_utf8_lossy(&accepted.stderr)
    );
    assert_eq!(envelope(&accepted)["result"]["files"], json!(1));
    let previous = pointer_gen(&project).unwrap();
    paths[1] = virtual_path(1, 765);
    replace_index(&project, &paths);
    assert_eq!(
        git_output(&project, &["ls-files", "-z"], &[]).len(),
        4_194_305
    );
    let refused = build(&project);
    assert!(!refused.status.success());
    assert!(envelope(&refused)["error"]["message"]
        .as_str()
        .unwrap()
        .contains("git ls-files output limit exceeded"));
    assert_eq!(pointer_gen(&project).as_deref(), Some(previous.as_str()));
}

#[test]
fn pointer_swap_preserves_existing_readers_of_the_previous_pointer() {
    let (tmp, project) = setup_project();
    assert!(build(&project).status.success());
    let pointer = project.join(".moai-cli/index/current.json");
    let before = fs::read(&pointer).unwrap();
    let held = tmp.path().join("held-pointer.json");
    fs::hard_link(&pointer, &held).unwrap();
    let old = pointer_gen(&project).unwrap();
    fs::write(project.join("a.txt"), b"replacement content").unwrap();
    let rebuilt = build(&project);
    assert!(
        rebuilt.status.success(),
        "{}",
        String::from_utf8_lossy(&rebuilt.stderr)
    );
    let new = pointer_gen(&project).unwrap();
    assert_ne!(new, old);
    assert_eq!(
        fs::read(&held).unwrap(),
        before,
        "an existing reader must retain the old pointer bytes"
    );
    assert_eq!(query(&project)["generation_id"], json!(new));
    assert_eq!(query(&project)["state"], json!("fresh"));
}

#[test]
fn tracked_regular_file_replaced_by_a_directory_is_deleted_not_a_gitlink() {
    let (_tmp, project) = setup_project();
    assert!(build(&project).status.success());
    fs::remove_file(project.join("a.txt")).unwrap();
    fs::create_dir(project.join("a.txt")).unwrap();
    let changed = query(&project);
    assert_eq!(changed["changes"]["deleted"], json!(["a.txt"]));
    assert_eq!(changed["changes"]["modified"], json!([]));
    assert!(build(&project).status.success());
    let generation = pointer_gen(&project).unwrap();
    let manifest: Value = serde_json::from_slice(
        &fs::read(
            project
                .join(".moai-cli/index/generations")
                .join(generation)
                .join("manifest.json"),
        )
        .unwrap(),
    )
    .unwrap();
    assert_eq!(manifest["files"], json!([]));
}

#[cfg(unix)]
#[test]
fn tracked_fifo_is_refused_as_a_special_file_without_reading_it() {
    let (_tmp, project) = setup_project();
    assert!(build(&project).status.success());
    let previous = pointer_gen(&project).unwrap();
    fs::remove_file(project.join("a.txt")).unwrap();
    assert!(Command::new("/usr/bin/mkfifo")
        .arg(project.join("a.txt"))
        .status()
        .unwrap()
        .success());
    let refused = build(&project);
    assert!(!refused.status.success());
    assert!(envelope(&refused)["error"]["message"]
        .as_str()
        .unwrap()
        .contains("unsupported file type"));
    assert_eq!(pointer_gen(&project).as_deref(), Some(previous.as_str()));
    assert_eq!(query(&project)["state"], json!("partial"));
}

#[test]
fn untracked_overflow_is_unavailable_instead_of_a_truncated_change_list() {
    let (_tmp, project) = setup_project();
    assert!(build(&project).status.success());
    for n in 0..16_384 {
        fs::write(project.join(format!("new-{n:05}")), b"").unwrap();
    }
    let at_limit = query(&project);
    assert_eq!(at_limit["state"], json!("stale"));
    assert_eq!(
        at_limit["changes"]["new_files"].as_array().unwrap().len(),
        16_384
    );
    fs::write(project.join("new-16384"), b"").unwrap();
    let overflow = query(&project);
    assert_eq!(overflow["state"], json!("unavailable"));
    assert_eq!(overflow["changes"], Value::Null);
    assert!(overflow["reason"]
        .as_str()
        .unwrap()
        .contains("untracked file limit exceeded"));
    fs::write(
        project.join(".moai-cli/index/last-build.json"),
        serde_json::to_vec(&json!({
            "status": "partial", "reason": "source read failed", "at": "2026-11-10T00:00:00Z"
        }))
        .unwrap(),
    )
    .unwrap();
    let partial = query(&project);
    assert_eq!(partial["state"], json!("partial"));
    assert_eq!(partial["changes"], Value::Null);
}

#[test]
fn actual_gitlink_keeps_the_submodule_commit_in_the_manifest() {
    let (_tmp, project) = setup_project();
    let module = project.join("module");
    fs::create_dir(&module).unwrap();
    git(&module, &["init", "-q", "--initial-branch=main"]);
    fs::write(module.join("inner.txt"), b"submodule content").unwrap();
    git(&module, &["add", "inner.txt"]);
    git(&module, &["commit", "-q", "-m", "module fixture"]);
    let commit = git_output(&module, &["rev-parse", "HEAD"], &[]);
    let commit = std::str::from_utf8(&commit).unwrap().trim();
    git(
        &project,
        &[
            "update-index",
            "--add",
            "--cacheinfo",
            &format!("160000,{commit},module"),
        ],
    );
    let built = build(&project);
    assert!(
        built.status.success(),
        "{}",
        String::from_utf8_lossy(&built.stderr)
    );
    assert_eq!(envelope(&built)["result"]["files"], json!(2));
    let generation = pointer_gen(&project).unwrap();
    let manifest: Value = serde_json::from_slice(
        &fs::read(
            project
                .join(".moai-cli/index/generations")
                .join(generation)
                .join("manifest.json"),
        )
        .unwrap(),
    )
    .unwrap();
    let entry = manifest["files"]
        .as_array()
        .unwrap()
        .iter()
        .find(|f| f["path"] == "module")
        .unwrap();
    assert_eq!(entry["sha256"], json!(format!("commit:{commit}")));
    assert_eq!(query(&project)["state"], json!("fresh"));
}

#[cfg(unix)]
fn fifo(path: &Path) {
    assert!(Command::new("mkfifo").arg(path).status().unwrap().success());
}

#[cfg(unix)]
#[test]
fn state_fifos_are_unavailable_and_never_wait_for_a_writer() {
    for state in ["current.json", "last-build.json", "manifest.json"] {
        let (_tmp, project) = setup_project();
        assert!(build(&project).status.success());
        let generation = pointer_gen(&project).unwrap();
        let path = if state == "manifest.json" {
            project
                .join(".moai-cli/index/generations")
                .join(&generation)
                .join(state)
        } else {
            project.join(".moai-cli/index").join(state)
        };
        if path.exists() {
            fs::remove_file(&path).unwrap();
        }
        fifo(&path);
        let queried = cli_before_deadline(&project, &["index-query", "."]);
        assert!(queried.status.success(), "{state}: {queried:?}");
        let result = envelope(&queried)["result"].clone();
        assert_eq!(result["state"], "unavailable", "{state}");
        assert_eq!(result["changes"], Value::Null, "{state}");
        if state == "manifest.json" {
            let failed = cli_before_deadline(&project, &["index-build", "."]);
            assert!(!failed.status.success());
            assert_eq!(pointer_gen(&project).as_deref(), Some(generation.as_str()));
            assert_eq!(query(&project)["state"], "partial");
        }
    }
}

#[cfg(unix)]
#[test]
fn failed_build_never_waits_when_its_failure_marker_is_a_fifo() {
    let (_tmp, project) = setup_project();
    assert!(build(&project).status.success());
    let previous = pointer_gen(&project).unwrap();
    let pointer = fs::read(project.join(".moai-cli/index/current.json")).unwrap();
    fifo(&project.join(".moai-cli/index/last-build.json"));
    fs::remove_file(project.join("a.txt")).unwrap();
    fifo(&project.join("a.txt"));
    let failed = cli_before_deadline(&project, &["index-build", "."]);
    assert!(!failed.status.success());
    assert_eq!(pointer_gen(&project).as_deref(), Some(previous.as_str()));
    assert_eq!(
        fs::read(project.join(".moai-cli/index/current.json")).unwrap(),
        pointer
    );
    let queried = cli_before_deadline(&project, &["index-query", "."]);
    assert_eq!(envelope(&queried)["result"]["state"], "unavailable");
    use std::os::unix::fs::FileTypeExt;
    assert!(
        fs::symlink_metadata(project.join(".moai-cli/index/last-build.json"))
            .unwrap()
            .file_type()
            .is_fifo()
    );
}

#[test]
fn manifest_file_count_has_literal_n_and_n_plus_one_read_boundaries() {
    let (_tmp, project) = setup_project();
    assert!(build(&project).status.success());
    let generation = pointer_gen(&project).unwrap();
    let path = project
        .join(".moai-cli/index/generations")
        .join(&generation)
        .join("manifest.json");
    let mut manifest: Value = serde_json::from_slice(&fs::read(&path).unwrap()).unwrap();
    let original = manifest["files"][0].clone();
    let missing: Vec<String> = (1..16_384).map(|n| format!("missing-{n:05}")).collect();
    let mut files = vec![original.clone()];
    files.extend(missing.iter().map(|name| {
        let mut entry = original.clone();
        entry["path"] = json!(name);
        entry
    }));
    manifest["files"] = json!(files);
    fs::write(&path, serde_json::to_vec(&manifest).unwrap()).unwrap();
    let accepted = query(&project);
    assert_eq!(accepted["state"], "stale");
    assert_eq!(accepted["generation_id"], generation);
    assert_eq!(accepted["changes"]["deleted"], json!(missing));
    let mut last = original;
    last["path"] = json!("missing-16384");
    manifest["files"].as_array_mut().unwrap().push(last);
    fs::write(&path, serde_json::to_vec(&manifest).unwrap()).unwrap();
    let refused = query(&project);
    assert_eq!(refused["state"], "unavailable");
    assert_eq!(refused["changes"], Value::Null);
}

#[test]
fn state_byte_budget_has_literal_n_and_n_plus_one_read_boundaries() {
    for state in ["current.json", "last-build.json", "manifest.json"] {
        let (_tmp, project) = setup_project();
        assert!(build(&project).status.success());
        let generation = pointer_gen(&project).unwrap();
        let path = if state == "manifest.json" {
            project
                .join(".moai-cli/index/generations")
                .join(generation)
                .join(state)
        } else {
            project.join(".moai-cli/index").join(state)
        };
        let mut bytes = if state == "last-build.json" {
            br#"{"status":"partial","reason":"known failure","at":"2026-11-10T00:00:00Z"}"#.to_vec()
        } else {
            fs::read(&path).unwrap()
        };
        bytes.reserve(33_554_433 - bytes.len());
        bytes.resize(33_554_432, b' ');
        fs::write(&path, &bytes).unwrap();
        let accepted = query(&project);
        assert_eq!(
            accepted["state"],
            if state == "last-build.json" {
                "partial"
            } else {
                "fresh"
            },
            "{state}"
        );
        bytes.push(b' ');
        fs::write(&path, &bytes).unwrap();
        let refused = query(&project);
        assert_eq!(refused["state"], "unavailable", "{state}");
        assert_eq!(refused["changes"], Value::Null, "{state}");
    }
}
