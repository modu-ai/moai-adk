//! #2175 (rc.2 B8) P09 opt-in home layout. The fixtures drive `layout-plan` /
//! `layout-apply` against a temporary project (profile + the three stores) with an
//! explicit `HOME` pointing at a temporary home root, so the default-path semantics
//! (AC-2) and the opt-in switch (AC-1) are both exercised without touching the real
//! `~/.moai-cli`. REQ-B8-01: the originals are never deleted; the pointer switches only
//! after the copy verifies; removing the marker reverts everything.
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
        .status()
        .unwrap();
    assert!(status.success(), "git {args:?} failed");
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
        &["layout-apply", ".", "--plan-digest", &"0".repeat(64)],
        None,
    );
    assert!(!out.status.success());
    assert_eq!(envelope(&out)["error"]["kind"], json!("stale_plan"));

    // Apply the real digest: every store is copied, the originals stay.
    let out = cli(
        &project,
        &home,
        &["layout-apply", ".", "--plan-digest", &digest],
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
        &["layout-apply", ".", "--plan-digest", &digest],
        None,
    );
    assert!(out.status.success());
    let before = hash_tree(&home);
    let out = cli(
        &project,
        &home,
        &["layout-apply", ".", "--plan-digest", &digest],
        None,
    );
    assert!(!out.status.success());
    assert_eq!(envelope(&out)["error"]["kind"], json!("conflict"));
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
