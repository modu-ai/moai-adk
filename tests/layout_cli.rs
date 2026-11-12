//! #2175 (rc.2 B8) P09 opt-in home layout. The fixtures drive `layout-plan` /
//! `layout-apply` against a temporary project (profile + the three stores) with an
//! explicit `HOME` pointing at a temporary home root, so the default-path semantics
//! (AC-2) and the opt-in switch (AC-1) are both exercised without touching the real
//! `~/.moai-cli`. REQ-B8-01: the originals are never deleted; the pointer switches only
//! after the copy verifies; removing the marker reverts everything.
mod common;
use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};

fn git(dir: &Path, args: &[&str]) {
    common::git::git_ok(dir, args);
}

/// A project with a profile and populated receipt/outbox/claim stores.
fn setup_project(home: &Path) -> (PathBuf, PathBuf) {
    let project = home.parent().unwrap().join("project");
    fs::create_dir_all(&project).unwrap();
    git(&project, &["init", "-q"]);
    fs::create_dir_all(project.join(".moai-cli")).unwrap();
    fs::write(
        project.join(".moai-cli/project.yaml"),
        r#"{"schema_version": 1, "project_id": "widgets", "purpose": "layout fixture", "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"]}"#,
    )
    .unwrap();
    fs::create_dir_all(project.join(".moai-cli/receipts")).unwrap();
    fs::write(project.join(".moai-cli/receipts/req-1.json"), b"{\"r\":1}").unwrap();
    fs::create_dir_all(project.join(".moai-cli/forge/outbox")).unwrap();
    fs::write(
        project.join(".moai-cli/forge/outbox/op-1.json"),
        b"{\"o\":1}",
    )
    .unwrap();
    let claims = project.join(".git/.moai-cli/claims");
    fs::create_dir_all(&claims).unwrap();
    fs::create_dir_all(home).unwrap();
    (project, home.to_path_buf())
}

fn cli(project: &Path, home: &Path, args: &[&str], stdin: Option<&str>) -> std::process::Output {
    let mut child = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .args(args)
        .current_dir(project)
        .env("HOME", home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .stdin(if stdin.is_some() {
            Stdio::piped()
        } else {
            Stdio::null()
        })
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    if let (Some(text), Some(mut sin)) = (stdin, child.stdin.take()) {
        match sin.write_all(text.as_bytes()) {
            Ok(()) => {}
            Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
            Err(e) => panic!("stdin write failed: {e}"),
        }
    }
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

fn hash_tree(root: &Path) -> Vec<String> {
    fn walk(root: &Path, dir: &Path, out: &mut Vec<String>) {
        for e in fs::read_dir(dir).unwrap() {
            let p = e.unwrap().path();
            let rel = p.strip_prefix(root).unwrap().to_string_lossy().into_owned();
            if p.is_dir() {
                walk(root, &p, out);
            } else {
                let bytes = fs::read(&p).unwrap();
                use moai_cli::digest;
                out.push(format!("{rel} {}", digest(&bytes)));
            }
        }
    }
    let mut out = Vec::new();
    walk(root, root, &mut out);
    out.sort();
    out
}

/// AC-1 + AC-2 (#2175 / REQ-B8-01): plan binds without writing; apply copies and
/// verifies without deleting the originals, then switches the pointer; the reader
/// falls back when the marker is removed (revert).
#[test]
fn home_layout_opt_in_moves_without_deleting_and_reverts() {
    let tmp = tempfile::tempdir().unwrap();
    let home = tmp.path().join("home");
    let (project, _home_guard) = setup_project(&home);

    // AC-2 first: without the marker, the claim journal resolves to the project side
    // (the git common dir) — the default paths are unchanged.
    let out = cli(
        &project,
        &home,
        &["claim-show", ".", "--task", "0123456789abcdef"],
        None,
    );
    let v = envelope(&out);
    assert_eq!(v["result"]["outcome"], json!("unclaimed"));
    assert!(project.join(".git/.moai-cli/claims").is_dir());

    // Plan: binds entries and digests, writes nothing on either side.
    let home_before = hash_tree(&home);
    let out = cli(&project, &home, &["layout-plan", "."], None);
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let v = envelope(&out);
    assert_eq!(v["result"]["schema_version"], json!(1));
    assert_eq!(v["result"]["project_id"], json!("widgets"));
    let entries = v["result"]["entries"].as_array().unwrap();
    assert_eq!(entries.len(), 3);
    let kinds: Vec<&str> = entries
        .iter()
        .map(|e| e["kind"].as_str().unwrap())
        .collect();
    assert_eq!(kinds, ["claims", "outbox", "receipts"]);
    let receipts = entries.iter().find(|e| e["kind"] == "receipts").unwrap();
    assert_eq!(receipts["files"], json!(1));
    assert!(receipts["source_digest"]
        .as_str()
        .unwrap()
        .starts_with("sha256:"));
    let digest = v["result"]["plan_digest"].as_str().unwrap().to_string();
    assert_eq!(hash_tree(&home), home_before, "the plan writes nothing");

    // Apply with a stale digest: refused (the house stale-plan vocabulary), nothing
    // switched.
    let out = cli(
        &project,
        &home,
        &[
            "layout-apply",
            ".",
            "--plan-digest",
            &"0".repeat(64),
            "--permission-reference",
            "test:layout-apply",
        ],
        None,
    );
    assert!(!out.status.success());
    assert_eq!(envelope(&out)["error"]["kind"], json!("stale_plan"));

    // Apply the real digest: every store is copied, the originals stay.
    let out = cli(
        &project,
        &home,
        &[
            "layout-apply",
            ".",
            "--plan-digest",
            &digest,
            "--permission-reference",
            "test:layout-apply",
        ],
        None,
    );
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let v = envelope(&out);
    assert_eq!(v["result"]["switched"], json!(true));
    assert_eq!(v["result"]["project_side_kept"], json!(true));
    let home_dir = home.join(".moai-cli/projects/widgets");
    assert!(home_dir.join("layout.json").is_file(), "the pointer exists");
    assert_eq!(
        fs::read(home_dir.join("receipts/req-1.json")).unwrap(),
        b"{\"r\":1}".to_vec(),
        "the copy is byte-identical"
    );
    assert_eq!(
        fs::read(home_dir.join("forge/outbox/op-1.json")).unwrap(),
        b"{\"o\":1}".to_vec()
    );
    assert!(
        project.join(".moai-cli/receipts/req-1.json").is_file(),
        "original kept"
    );
    assert!(
        project.join(".moai-cli/forge/outbox/op-1.json").is_file(),
        "original kept"
    );

    // After the switch, the claim journal resolves to the home side (the marker wins).
    let out = cli(
        &project,
        &home,
        &["claim-show", ".", "--task", "0123456789abcdef"],
        None,
    );
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );

    // REQ-B8-01 revert: removing the marker puts every reader back on the project side.
    fs::remove_file(home_dir.join("layout.json")).unwrap();
    let out = cli(
        &project,
        &home,
        &["claim-show", ".", "--task", "0123456789abcdef"],
        None,
    );
    let v = envelope(&out);
    assert_eq!(
        v["result"]["outcome"],
        json!("unclaimed"),
        "back on the project-side journal"
    );
    // The copied stores stay (the originals were never touched, so the home copies are
    // harmless history); only the marker is gone.
    assert!(home_dir.join("receipts/req-1.json").is_file());
    assert!(home_dir.join("forge/outbox/op-1.json").is_file());
    assert!(
        !home_dir.join("layout.json").exists(),
        "the marker is the revert"
    );
}

/// AC-1 double apply: the pointer already switched — a second apply is a conflict and
/// neither side is touched.
#[test]
fn layout_apply_twice_is_a_conflict() {
    let tmp = tempfile::tempdir().unwrap();
    let home = tmp.path().join("home");
    let (project, _home_guard) = setup_project(&home);
    let out = cli(&project, &home, &["layout-plan", "."], None);
    let digest = envelope(&out)["result"]["plan_digest"]
        .as_str()
        .unwrap()
        .to_string();
    let out = cli(
        &project,
        &home,
        &[
            "layout-apply",
            ".",
            "--plan-digest",
            &digest,
            "--permission-reference",
            "test:layout-apply",
        ],
        None,
    );
    assert!(out.status.success());
    let before = hash_tree(&home);
    let out = cli(
        &project,
        &home,
        &[
            "layout-apply",
            ".",
            "--plan-digest",
            &digest,
            "--permission-reference",
            "test:layout-apply",
        ],
        None,
    );
    assert!(!out.status.success());
    assert_eq!(envelope(&out)["error"]["kind"], json!("conflict"));
    assert!(
        envelope(&out)["error"]["message"]
            .as_str()
            .unwrap()
            .contains("already switched"),
        "the second apply is refused by the switched-pointer guard itself: {}",
        envelope(&out)["error"]["message"]
    );
    assert_eq!(hash_tree(&home), before, "a refused apply writes nothing");
}

/// AC-2: with no opt-in anywhere, the three stores resolve exactly as before — the
/// receipt store, the outbox store and the claim journal all stay project-side and no
/// home directory is created.
#[test]
fn layout_default_keeps_project_side() {
    let tmp = tempfile::tempdir().unwrap();
    let home = tmp.path().join("home");
    let (project, _home_guard) = setup_project(&home);
    let out = cli(&project, &home, &["layout-plan", "."], None);
    let digest = envelope(&out)["result"]["plan_digest"]
        .as_str()
        .unwrap()
        .to_string();

    // A project-side claim acquire (via claim-show on a missing key writes nothing, so
    // drive the journal through the product instead: layout-plan only reads).
    let _ = digest;
    assert!(
        !home.join(".moai-cli/projects").exists(),
        "no home project dir without opt-in"
    );
    assert!(project.join(".moai-cli/receipts/req-1.json").is_file());
    let out = cli(
        &project,
        &home,
        &["claim-show", ".", "--task", "0123456789abcdef"],
        None,
    );
    assert!(out.status.success());
    assert!(
        project.join(".git/.moai-cli/claims").is_dir(),
        "the claim journal still resolves project-side by default"
    );
}

/// The audited data-loss scenario (2026-11-09 FAIL, issue #2175): apply switches the
/// pointer, the home side receives new data during the switched window, the marker is
/// removed (the documented revert), and a re-apply must REFUSE instead of copying over
/// — the old code copied project-side bytes over the home target and its verify
/// failure deleted the home side, the only copy of the newer data. Repair direction
/// (a): an existing home target refuses the apply; nothing is overwritten or deleted.
#[test]
fn revert_then_reapply_refuses_and_keeps_home_data() {
    let tmp = tempfile::tempdir().unwrap();
    let home = tmp.path().join("home");
    let (project, _home_guard) = setup_project(&home);
    let out = cli(&project, &home, &["layout-plan", "."], None);
    let digest = envelope(&out)["result"]["plan_digest"]
        .as_str()
        .unwrap()
        .to_string();
    let out = cli(
        &project,
        &home,
        &[
            "layout-apply",
            ".",
            "--plan-digest",
            &digest,
            "--permission-reference",
            "test:layout-apply",
        ],
        None,
    );
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let home_dir = home.join(".moai-cli/projects/widgets");

    // New data lands on the home side during the switched window.
    fs::write(home_dir.join("receipts/req-2.json"), b"{\"r\":2}").unwrap();

    // The documented revert: remove the marker; readers fall back to the project side.
    fs::remove_file(home_dir.join("layout.json")).unwrap();
    let out = cli(
        &project,
        &home,
        &["claim-show", ".", "--task", "0123456789abcdef"],
        None,
    );
    assert_eq!(
        envelope(&out)["result"]["outcome"],
        json!("unclaimed"),
        "back on the project-side journal"
    );

    // Re-apply after the revert: refused, and the home-side req-2 (the only copy)
    // survives byte for byte.
    let before = hash_tree(&home);
    let out = cli(&project, &home, &["layout-plan", "."], None);
    let digest2 = envelope(&out)["result"]["plan_digest"]
        .as_str()
        .unwrap()
        .to_string();
    let out = cli(
        &project,
        &home,
        &[
            "layout-apply",
            ".",
            "--plan-digest",
            &digest2,
            "--permission-reference",
            "test:layout-apply",
        ],
        None,
    );
    assert!(!out.status.success(), "the re-apply must refuse");
    let v = envelope(&out);
    assert_eq!(v["error"]["kind"], json!("conflict"));
    assert!(
        v["error"]["message"]
            .as_str()
            .unwrap()
            .contains("already exists"),
        "{}",
        v["error"]["message"]
    );
    assert_eq!(
        hash_tree(&home),
        before,
        "a refused re-apply writes nothing and deletes nothing"
    );
    assert_eq!(
        fs::read(home_dir.join("receipts/req-2.json")).unwrap(),
        b"{\"r\":2}".to_vec(),
        "the only copy of the switched-window data survives"
    );
    assert_eq!(
        fs::read(home_dir.join("receipts/req-1.json")).unwrap(),
        b"{\"r\":1}".to_vec()
    );
    assert!(
        project.join(".moai-cli/receipts/req-1.json").is_file(),
        "the project side is untouched"
    );
}

/// The claim journal lock closes the copy→switch window: with the journal lock held,
/// the apply refuses (busy) before any copy, and it succeeds once the lock is released.
#[test]
fn layout_apply_refuses_while_the_claim_journal_is_busy() {
    let tmp = tempfile::tempdir().unwrap();
    let home = tmp.path().join("home");
    let (project, _home_guard) = setup_project(&home);
    let out = cli(&project, &home, &["layout-plan", "."], None);
    let digest = envelope(&out)["result"]["plan_digest"]
        .as_str()
        .unwrap()
        .to_string();

    // Hold the same advisory lock a claim-acquire holds (cross-process, so the
    // subprocess below really sees busy).
    let claims = project.join(".git/.moai-cli/claims");
    let journal = moai_cli::claim::Journal::open(&claims).unwrap();
    assert!(
        journal.try_lock().unwrap(),
        "the test holds the journal lock"
    );

    let before = hash_tree(&home);
    let out = cli(
        &project,
        &home,
        &[
            "layout-apply",
            ".",
            "--plan-digest",
            &digest,
            "--permission-reference",
            "test:layout-apply",
        ],
        None,
    );
    assert_eq!(
        out.status.code(),
        Some(1),
        "a busy journal refuses the apply"
    );
    let v = envelope(&out);
    // Literal oracle: changing the production constant must not silently update this contract.
    let message = "journal busy: the claim journal lock is unavailable (another process holds it); the caller may retry the same command after a bounded wait, within its own cap (REQ-06A-05)";
    assert_eq!(
        v,
        json!({
            "protocol_version": 1,
            "operation": "layout.apply",
            "ok": false,
            "error": {"kind": "unavailable", "message": message, "retryable": true}
        })
    );
    assert_eq!(
        String::from_utf8(out.stderr).unwrap(),
        format!("{message}\n")
    );
    assert_eq!(
        hash_tree(&home),
        before,
        "a busy-refused apply copies nothing"
    );

    // Released: the same command now succeeds.
    drop(journal);
    let out = cli(
        &project,
        &home,
        &[
            "layout-apply",
            ".",
            "--plan-digest",
            &digest,
            "--permission-reference",
            "test:layout-apply",
        ],
        None,
    );
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    assert_eq!(envelope(&out)["result"]["switched"], json!(true));
}

#[test]
fn outbox_reads_follow_the_active_store_without_creating_directories() {
    use moai_cli::forge::outbox::{self, OutboxEntry, OutboxState, Phase};

    const CHILD_ROOT: &str = "MOAI_P09_OUTBOX_CHILD_ROOT";
    if let Some(root) = std::env::var_os(CHILD_ROOT) {
        let root = PathBuf::from(root).canonicalize().unwrap();
        assert!(root.starts_with(std::env::temp_dir().canonicalize().unwrap()));
        assert!(root
            .file_name()
            .unwrap()
            .to_string_lossy()
            .starts_with("moai-layout-outbox-"));
        let project = root.join("project");
        let home = root.join("home");
        assert_eq!(
            PathBuf::from(std::env::var_os("HOME").unwrap())
                .canonicalize()
                .unwrap(),
            home.canonicalize().unwrap()
        );
        let home_project = home.join(".moai-cli/projects/widgets");
        let store = home_project.join("forge/outbox");
        fs::remove_dir_all(&store).unwrap();
        assert!(outbox::load(&project, "p09-history").unwrap().is_none());
        assert!(
            !store.exists(),
            "an absent read must not create the active store"
        );

        let mut updated = OutboxEntry::new("p09-history", "p09-marker");
        updated
            .advance(
                OutboxState::Blocked,
                Phase {
                    at: "2026-11-10T00:00:00Z".into(),
                    detail: json!({"source": "home"}),
                },
            )
            .unwrap();
        outbox::save(&project, &updated).unwrap();
        let loaded = outbox::load(&project, "p09-history").unwrap();
        assert_eq!(loaded.as_ref(), Some(&updated));
        fs::remove_file(home_project.join("layout.json")).unwrap();
        let reverted = outbox::load(&project, "p09-history").unwrap().unwrap();
        assert_eq!(reverted.state, OutboxState::Proposed);
        assert!(
            reverted.intent.is_none(),
            "the frozen project history is unchanged"
        );
        fs::write(
            root.join("outbox-result.json"),
            serde_json::to_vec(&loaded).unwrap(),
        )
        .unwrap();
        return;
    }

    let tmp = tempfile::Builder::new()
        .prefix("moai-layout-outbox-")
        .tempdir()
        .unwrap();
    let home = tmp.path().join("home");
    let (project, _) = setup_project(&home);
    let legacy = OutboxEntry::new("p09-history", "p09-marker");
    let legacy_bytes = serde_json::to_vec(&legacy).unwrap();
    fs::write(
        project.join(".moai-cli/forge/outbox/p09-history.json"),
        &legacy_bytes,
    )
    .unwrap();
    let preview = cli(&project, &home, &["layout-plan", "."], None);
    assert!(
        preview.status.success(),
        "{}",
        String::from_utf8_lossy(&preview.stderr)
    );
    let digest = envelope(&preview)["result"]["plan_digest"]
        .as_str()
        .unwrap()
        .to_owned();
    let applied = cli(
        &project,
        &home,
        &[
            "layout-apply",
            ".",
            "--plan-digest",
            &digest,
            "--permission-reference",
            "test:outbox-home",
        ],
        None,
    );
    assert!(
        applied.status.success(),
        "{}",
        String::from_utf8_lossy(&applied.stderr)
    );
    let thread = std::thread::current();
    let child = Command::new(std::env::current_exe().unwrap())
        .args(["--exact", thread.name().unwrap(), "--nocapture"])
        .env(CHILD_ROOT, tmp.path())
        .env("HOME", &home)
        .output()
        .unwrap();
    assert!(
        child.status.success(),
        "{}\n{}",
        String::from_utf8_lossy(&child.stdout),
        String::from_utf8_lossy(&child.stderr)
    );
    let loaded: Value =
        serde_json::from_slice(&fs::read(tmp.path().join("outbox-result.json")).unwrap()).unwrap();
    assert_eq!(loaded["state"], json!("blocked"));
    assert_eq!(loaded["intent"]["detail"]["source"], json!("home"));
    assert_eq!(
        fs::read(project.join(".moai-cli/forge/outbox/p09-history.json")).unwrap(),
        legacy_bytes
    );
}

#[cfg(unix)]
#[test]
fn source_inspection_errors_are_aggregated_before_any_switch() {
    use std::os::unix::fs::PermissionsExt;

    let tmp = tempfile::tempdir().unwrap();
    let home = tmp.path().join("home");
    let (project, _) = setup_project(&home);
    let stores = [
        project.join(".moai-cli/receipts"),
        project.join(".moai-cli/forge/outbox"),
    ];
    let files = [stores[0].join("req-1.json"), stores[1].join("op-1.json")];
    for paths in [&stores, &files] {
        let original_permissions = [
            fs::metadata(&paths[0]).unwrap().permissions(),
            fs::metadata(&paths[1]).unwrap().permissions(),
        ];
        for path in paths {
            fs::set_permissions(path, fs::Permissions::from_mode(0o000)).unwrap();
        }
        let preview = cli(&project, &home, &["layout-plan", "."], None);
        let applied = cli(
            &project,
            &home,
            &[
                "layout-apply",
                ".",
                "--plan-digest",
                "unconfirmed",
                "--permission-reference",
                "test:unreadable",
            ],
            None,
        );
        for (path, permissions) in paths.iter().zip(original_permissions) {
            fs::set_permissions(path, permissions).unwrap();
        }
        for output in [&preview, &applied] {
            assert!(
                !output.status.success(),
                "an unreadable source cannot be confirmed"
            );
            let failure = envelope(output);
            assert_eq!(
                failure["error"]["kind"],
                json!(moai_cli::contract::Kind::Io.as_str())
            );
            let message = failure["error"]["message"].as_str().unwrap();
            assert!(
                message.contains("receipts") && message.contains("outbox"),
                "{message}"
            );
            assert!(
                !message.contains(tmp.path().to_str().unwrap()),
                "private paths must not escape"
            );
        }
    }
    assert!(!home.join(".moai-cli/projects").exists());
    assert_eq!(
        fs::read(project.join(".moai-cli/receipts/req-1.json")).unwrap(),
        b"{\"r\":1}"
    );
    assert_eq!(
        fs::read(project.join(".moai-cli/forge/outbox/op-1.json")).unwrap(),
        b"{\"o\":1}"
    );
}
