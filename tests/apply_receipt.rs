//! Apply receipt and recovery fixtures (Issue #1989, 01-B). Kept in their own file so
//! concurrent lanes merging tests/product_cli.rs do not collide on its tail. Receipt
//! paths, statuses and recovery semantics are asserted through the CLI envelope only.

mod common;
use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};
use tempfile::TempDir;

fn profile() -> Value {
    // v2 input (03-A REQ-03A-08): project-plan accepts schema_version 2 only.
    json!({"schema_version":2,"project_id":"demo","purpose":"Build a small editor","tools":["Rust"],"development_methods":["TDD"],"hosts":["claude","codex"],"interview_revision":"rc1-fixture-1"})
}

fn cli(args: &[&str], input: &Value) -> std::process::Output {
    let mut child = Command::new(crate::common::MOAI_BIN)
        .args(args)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    // The child may validate its arguments and exit before reading stdin; only a
    // broken pipe is tolerated here (see tests/product_cli.rs).
    match child
        .stdin
        .take()
        .unwrap()
        .write_all(input.to_string().as_bytes())
    {
        Ok(()) => {}
        Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
        Err(e) => panic!("stdin write failed: {e}"),
    }
    child.wait_with_output().unwrap()
}

fn ok_result(out: &std::process::Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn agents_request(language: &str, local_file: &str) -> Value {
    json!({"protocol_version":1,"language":language,"workflow_reference":"skills/moai/workflows/","local_file":local_file})
}

fn agents_plan(root: &TempDir, request: &Value) -> std::process::Output {
    cli(&["agents-plan", root.path().to_str().unwrap()], request)
}

fn agents_apply(root: &TempDir, plan: &Value) -> std::process::Output {
    cli(
        &[
            "agents-apply",
            root.path().to_str().unwrap(),
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        plan,
    )
}

fn apply_recover(root: &TempDir, receipt_id: &str) -> std::process::Output {
    cli(
        &[
            "apply-recover",
            root.path().to_str().unwrap(),
            "--receipt",
            receipt_id,
            "--permission-reference",
            "fixture-approved",
        ],
        &json!({}),
    )
}

/// A project with `project.yaml` in place; the caller runs agents-plan/apply itself.
fn agents_project_root(root: &TempDir) {
    fs::create_dir_all(root.path().join("skills/moai/workflows")).unwrap();
    let p = root.path().to_str().unwrap();
    let preview = cli(&["project-plan", p], &profile());
    assert!(preview.status.success(), "{:?}", preview.stderr);
    let plan: Value = ok_result(&preview);
    let result = cli(
        &[
            "project-apply",
            p,
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        &plan,
    );
    assert!(result.status.success(), "{:?}", result.stderr);
}

fn digest_head(plan_digest: &str) -> String {
    plan_digest
        .strip_prefix("sha256:")
        .unwrap_or(plan_digest)
        .chars()
        .take(12)
        .collect()
}

/// The receipts of one request, identified by the plan digest head in their names.
fn find_receipts(root: &Path, head: &str) -> Vec<PathBuf> {
    let dir = root.join(".moai-cli/receipts");
    let mut found = Vec::new();
    for e in fs::read_dir(&dir).unwrap() {
        let p = e.unwrap().path();
        let name = p.file_name().unwrap().to_str().unwrap().to_string();
        if name.starts_with(&format!("agents-apply-{head}-")) && name.ends_with(".json") {
            found.push(p);
        }
    }
    found
}

fn read_receipt(root: &Path, head: &str) -> Value {
    let mut found = find_receipts(root, head);
    assert_eq!(found.len(), 1, "expected exactly one receipt for {head}");
    serde_json::from_slice(&fs::read(found.remove(0)).unwrap()).unwrap()
}

fn entry<'a>(receipt: &'a Value, path: &str) -> &'a Value {
    receipt["entries"]
        .as_array()
        .unwrap()
        .iter()
        .find(|e| e["path"] == json!(path))
        .unwrap_or_else(|| panic!("no entry for {path} in {}", receipt["request_id"]))
}

#[cfg(unix)]
fn temp_files_left(root: &Path) -> Vec<PathBuf> {
    fn walk(dir: &Path, found: &mut Vec<PathBuf>) {
        for e in fs::read_dir(dir).unwrap() {
            let p = e.unwrap().path();
            if p.is_dir() {
                walk(&p, found);
            } else if p
                .file_name()
                .and_then(|n| n.to_str())
                .is_some_and(|n| n.starts_with(".moai-apply-") && n.ends_with(".tmp"))
            {
                found.push(p);
            }
        }
    }
    let mut found = Vec::new();
    walk(root, &mut found);
    found
}

#[test]
fn receipt_records_per_file_results() {
    // REQ-01B-06: path, decision, before/after digests, published — and the receipt id
    // on the apply result, so `apply-recover --receipt` has one source of truth.
    let root = TempDir::new().unwrap();
    agents_project_root(&root);
    // "ignored" exercises the widest file set: AGENTS.md, AGENTS.local.md, .gitignore
    // and the generation manifest.
    let plan: Value = ok_result(&agents_plan(&root, &agents_request("en", "ignored")));
    let applied = agents_apply(&root, &plan);
    assert!(
        applied.status.success(),
        "{:?}",
        String::from_utf8_lossy(&applied.stderr)
    );
    let result = ok_result(&applied);
    let receipt_id = result["receipt"].as_str().unwrap();
    let head = digest_head(plan["plan_digest"].as_str().unwrap());
    assert!(
        receipt_id.starts_with(&format!("agents-apply-{head}-")),
        "{receipt_id}"
    );

    let receipt = read_receipt(root.path(), &head);
    assert_eq!(receipt["schema_version"], json!(1));
    assert_eq!(receipt["request_id"], json!(receipt_id));
    assert_eq!(receipt["operation"], json!("agents.apply"));
    assert_eq!(receipt["status"], json!("complete"));

    let agents = entry(&receipt, "AGENTS.md");
    assert_eq!(agents["decision"], json!("create"));
    assert_eq!(agents["before_sha256"], Value::Null);
    assert!(agents["after_sha256"].is_string());
    assert_eq!(agents["published"], json!(true));
    assert!(agents["error"].is_null());

    let ignore = entry(&receipt, ".gitignore");
    assert_eq!(ignore["decision"], json!("create"));
    assert_eq!(ignore["published"], json!(true));

    let local = entry(&receipt, "AGENTS.local.md");
    assert_eq!(local["published"], json!(true));

    let manifest = entry(&receipt, ".moai-cli/manifests/generation.json");
    assert_eq!(manifest["published"], json!(true));
    // Receipt JSON carries digests only; the strict fields of the 01-B schema.
    for key in [
        "schema_version",
        "request_id",
        "operation",
        "entries",
        "status",
        "clone_id",
        "workspace_id",
    ] {
        assert!(receipt.get(key).is_some(), "missing {key}");
    }
}

#[cfg(unix)]
#[test]
fn apply_receipt_records_partial_apply_and_leaves_no_temp() {
    // AC-01B-2/AC-01B-3: a forced failure mid-apply keeps the published files, records
    // the failed file in the receipt with its error, marks the receipt partial — and no
    // temporary file survives anywhere under the project.
    let root = TempDir::new().unwrap();
    agents_project_root(&root);
    let manifests = root.path().join(".moai-cli/manifests");
    fs::create_dir_all(&manifests).unwrap();
    use std::os::unix::fs::PermissionsExt;
    fs::set_permissions(&manifests, fs::Permissions::from_mode(0o555)).unwrap();
    let plan: Value = ok_result(&agents_plan(&root, &agents_request("en", "tracked")));
    let result = agents_apply(&root, &plan);
    assert!(!result.status.success());
    let stderr = String::from_utf8_lossy(&result.stderr);
    assert!(
        stderr.contains("partial apply") && stderr.contains("AGENTS.md"),
        "{stderr}"
    );
    assert!(root.path().join("AGENTS.md").is_file());

    let receipt = read_receipt(
        root.path(),
        &digest_head(plan["plan_digest"].as_str().unwrap()),
    );
    assert_eq!(receipt["status"], json!("partial"));
    let manifest = entry(&receipt, ".moai-cli/manifests/generation.json");
    assert_eq!(manifest["published"], json!(false));
    assert!(manifest["error"].is_string(), "{}", manifest);
    let agents = entry(&receipt, "AGENTS.md");
    assert_eq!(agents["published"], json!(true));
    assert!(
        temp_files_left(root.path()).is_empty(),
        "{:?}",
        temp_files_left(root.path())
    );
    fs::set_permissions(&manifests, fs::Permissions::from_mode(0o755)).unwrap();
}

#[test]
fn apply_recover_restores_only_unchanged_targets() {
    // AC-01B-4/AC-01B-5: recovery reverts exactly the files the request published and
    // still owns; a file the user edited since keeps the user's bytes.
    let root = TempDir::new().unwrap();
    agents_project_root(&root);
    let plan: Value = ok_result(&agents_plan(&root, &agents_request("en", "ignored")));
    let applied = agents_apply(&root, &plan);
    let result = ok_result(&applied);
    let receipt_id = result["receipt"].as_str().unwrap().to_string();

    // The user edits the managed file's non-managed prefix after the apply.
    let agents_md = root.path().join("AGENTS.md");
    let user_bytes = fs::read(&agents_md).unwrap();
    let user_text = String::from_utf8(user_bytes.clone()).unwrap().replacen(
        "# demo",
        "# demo — my own header",
        1,
    );
    assert_ne!(user_text.as_bytes(), user_bytes);
    fs::write(&agents_md, &user_text).unwrap();

    let recovered = apply_recover(&root, &receipt_id);
    let out = ok_result(&recovered);
    let reverted: Vec<&str> = out["reverted"]
        .as_array()
        .unwrap()
        .iter()
        .map(|v| v.as_str().unwrap())
        .collect();
    let skipped: Vec<&str> = out["skipped"]
        .as_array()
        .unwrap()
        .iter()
        .map(|v| v["path"].as_str().unwrap())
        .collect();

    // AGENTS.md changed since the apply: preserved byte for byte, reported as skipped.
    assert_eq!(skipped, vec!["AGENTS.md"]);
    assert_eq!(fs::read(&agents_md).unwrap(), user_text.as_bytes());
    // Created files are removed again; nothing else keeps apply state.
    assert!(!root.path().join("AGENTS.local.md").exists());
    assert!(!root.path().join(".gitignore").exists());
    assert!(!root
        .path()
        .join(".moai-cli/manifests/generation.json")
        .exists());
    assert!(reverted.contains(&"AGENTS.local.md"));
    assert!(reverted.contains(&".gitignore"));
    assert!(reverted.contains(&".moai-cli/manifests/generation.json"));
    // A second recovery finds nothing left to revert.
    let again = ok_result(&apply_recover(&root, &receipt_id));
    assert_eq!(again["state"], json!("nothing_to_revert"));
}

#[test]
fn reapply_records_zero_published_entries() {
    // REQ-01B-10 / AC-01B-6: the second apply publishes nothing and the files keep
    // their bytes; the receipt shows the unchanged files with zero published entries.
    let root = TempDir::new().unwrap();
    agents_project_root(&root);
    let first: Value = ok_result(&agents_plan(&root, &agents_request("en", "ignored")));
    assert!(agents_apply(&root, &first).status.success());
    let snapshot = |root: &TempDir| -> Vec<(String, Vec<u8>)> {
        let mut v = Vec::new();
        for rel in ["AGENTS.md", "AGENTS.local.md", ".gitignore"] {
            v.push((rel.into(), fs::read(root.path().join(rel)).unwrap()));
        }
        v
    };
    let before = snapshot(&root);

    let second: Value = ok_result(&agents_plan(&root, &agents_request("en", "ignored")));
    let applied = agents_apply(&root, &second);
    assert!(
        applied.status.success(),
        "{:?}",
        String::from_utf8_lossy(&applied.stderr)
    );
    let result = ok_result(&applied);
    assert_eq!(result["state"], json!("unchanged"), "{result}");
    assert_eq!(snapshot(&root), before, "re-run changed file bytes");

    let receipt = read_receipt(
        root.path(),
        &digest_head(second["plan_digest"].as_str().unwrap()),
    );
    assert_eq!(receipt["status"], json!("complete"));
    let published: Vec<&Value> = receipt["entries"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|e| e["published"] == json!(true))
        .collect();
    assert!(published.is_empty(), "{receipt}");
    assert!(!receipt["entries"].as_array().unwrap().is_empty());
}

#[test]
fn recover_refuses_go_paths_and_malformed_receipts() {
    // REQ-01B-09: a receipt naming Go-generation state or the receipt store is refused;
    // unknown or malformed receipts fail with a typed kind, not a crash.
    let root = TempDir::new().unwrap();
    agents_project_root(&root);
    let receipts = root.path().join(".moai-cli/receipts");
    fs::create_dir_all(&receipts).unwrap();

    let go_path_receipt = json!({
        "schema_version":1,
        "request_id":"manual-go-1",
        "operation":"agents.apply",
        "clone_id":null,"workspace_id":null,
        "entries":[{"path":".moai/state/x.json","decision":"create","before_sha256":null,
                    "after_sha256":"sha256:0000000000000000000000000000000000000000000000000000000000000000",
                    "published":true,"error":null}],
        "status":"complete"
    });
    fs::write(
        receipts.join("manual-go-1.json"),
        serde_json::to_vec(&go_path_receipt).unwrap(),
    )
    .unwrap();
    let out = apply_recover(&root, "manual-go-1");
    assert!(!out.status.success());
    let stderr = String::from_utf8_lossy(&out.stderr);
    assert!(stderr.contains("go generation state path"), "{stderr}");
    // The named path was never touched.
    assert!(!root.path().join(".moai").exists());

    // The receipt store itself is not recoverable.
    let store_receipt = {
        let mut v = go_path_receipt.clone();
        v["request_id"] = json!("manual-store-1");
        v["entries"] = json!([{"path":".moai-cli/receipts/manual-go-1.json","decision":"create",
            "before_sha256":null,"after_sha256":"sha256:0000000000000000000000000000000000000000000000000000000000000000",
            "published":true,"error":null}]);
        v
    };
    fs::write(
        receipts.join("manual-store-1.json"),
        serde_json::to_vec(&store_receipt).unwrap(),
    )
    .unwrap();
    let out = apply_recover(&root, "manual-store-1");
    assert!(!out.status.success());
    let stderr = String::from_utf8_lossy(&out.stderr);
    assert!(stderr.contains("receipt store path"), "{stderr}");

    // Unknown id and an id with path syntax are typed refusals.
    let out = apply_recover(&root, "manual-not-there");
    assert!(!out.status.success());
    let envelope: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(envelope["error"]["kind"], json!("not_found"), "{envelope}");
    let out = apply_recover(&root, "../escape");
    assert!(!out.status.success());
    let envelope: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(
        envelope["error"]["kind"],
        json!("invalid_input"),
        "{envelope}"
    );

    // deny_unknown_fields: an extra field makes the receipt unreadable, not plausibly valid.
    let extra_field = {
        let mut v = go_path_receipt.clone();
        v["request_id"] = json!("manual-extra-1");
        v["surprise"] = json!(true);
        v
    };
    fs::write(
        receipts.join("manual-extra-1.json"),
        serde_json::to_vec(&extra_field).unwrap(),
    )
    .unwrap();
    let out = apply_recover(&root, "manual-extra-1");
    assert!(!out.status.success());
    let stderr = String::from_utf8_lossy(&out.stderr);
    assert!(stderr.contains("receipt unreadable"), "{stderr}");
}

#[test]
fn project_apply_writes_receipt_and_rerun_is_unchanged() {
    // The project profile path runs through the same engine: created once with a
    // receipt, then a re-run records zero publishes and touches nothing.
    let root = TempDir::new().unwrap();
    agents_project_root(&root);
    let receipts = root.path().join(".moai-cli/receipts");
    let names: Vec<String> = fs::read_dir(&receipts)
        .unwrap()
        .map(|e| e.unwrap().file_name().to_str().unwrap().to_string())
        .filter(|n| n.starts_with("project-apply-") && n.ends_with(".json"))
        .collect();
    assert_eq!(names.len(), 1, "{names:?}");
    let receipt: Value =
        serde_json::from_slice(&fs::read(receipts.join(&names[0])).unwrap()).unwrap();
    assert_eq!(receipt["operation"], json!("project.apply"));
    assert_eq!(receipt["status"], json!("complete"));
    let entry = &receipt["entries"][0];
    assert_eq!(entry["path"], json!(".moai-cli/project.yaml"));
    assert_eq!(entry["decision"], json!("create"));
    assert_eq!(entry["published"], json!(true));

    // A second apply of the same profile is unchanged with a fresh zero-publish receipt.
    let p = root.path().to_str().unwrap();
    let preview = cli(&["project-plan", p], &profile());
    let plan: Value = ok_result(&preview);
    assert_eq!(plan["action"], json!("unchanged"));
    let result = cli(
        &[
            "project-apply",
            p,
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        &plan,
    );
    let applied = ok_result(&result);
    assert_eq!(applied["state"], json!("unchanged"));
    let receipt_id = applied["receipt"].as_str().unwrap().to_string();
    let second: Value =
        serde_json::from_slice(&fs::read(receipts.join(format!("{receipt_id}.json"))).unwrap())
            .unwrap();
    assert_eq!(second["status"], json!("complete"));
    assert!(
        second["entries"]
            .as_array()
            .unwrap()
            .iter()
            .all(|e| e["published"] == json!(false)),
        "{second}"
    );
}

#[test]
fn project_update_recover_restores_profile_and_baseline() {
    // Audit #1995 round 3 (leader decision a): the replaced baseline record keeps its
    // before bytes like any other published file, so recovering an update restores
    // project.yaml AND the baseline together, and the re-planned input is a real
    // update again instead of a silent unchanged.
    let root = TempDir::new().unwrap();
    agents_project_root(&root);
    let p = root.path().to_str().unwrap();
    let mut next = profile();
    next["purpose"] = json!("Second purpose");
    let plan: Value = ok_result(&cli(&["project-plan", p], &next));
    assert_eq!(plan["action"], json!("update"));
    let applied = ok_result(&cli(
        &[
            "project-apply",
            p,
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        &plan,
    ));
    assert_eq!(applied["state"], json!("updated"));
    assert!(String::from_utf8_lossy(
        &fs::read(root.path().join(".moai-cli/project.yaml")).unwrap()
    )
    .contains("Second purpose"));
    let receipt_id = applied["receipt"].as_str().unwrap().to_string();
    // Recovery restores both published entries from their before bytes.
    let recovered = ok_result(&apply_recover(&root, &receipt_id));
    assert_eq!(recovered["state"], json!("reverted"), "{recovered}");
    assert_eq!(
        recovered["reverted"],
        json!([
            ".moai-cli/project.yaml",
            ".moai-cli/manifests/profile-baseline.json"
        ])
    );
    let profile_after: Value =
        serde_json::from_slice(&fs::read(root.path().join(".moai-cli/project.yaml")).unwrap())
            .unwrap();
    assert_eq!(profile_after["purpose"], json!("Build a small editor"));
    let baseline_after: Value = serde_json::from_slice(
        &fs::read(
            root.path()
                .join(".moai-cli/manifests/profile-baseline.json"),
        )
        .unwrap(),
    )
    .unwrap();
    assert_eq!(
        baseline_after["profile"]["purpose"],
        json!("Build a small editor")
    );
    // The same input plans a real update again; nothing was silently absorbed.
    let replan: Value = ok_result(&cli(&["project-plan", p], &next));
    assert_eq!(replan["action"], json!("update"));
}

#[test]
fn agents_ignore_block_excludes_the_receipt_directory() {
    // HOLD-B1: the managed .gitignore block renders the receipts exclusion, so the
    // per-request receipts never become tracked user content.
    let root = TempDir::new().unwrap();
    agents_project_root(&root);
    let plan: Value = ok_result(&agents_plan(&root, &agents_request("en", "ignored")));
    assert!(agents_apply(&root, &plan).status.success());
    let text = String::from_utf8(fs::read(root.path().join(".gitignore")).unwrap()).unwrap();
    assert!(text.contains(".moai-cli/receipts/"), "{text}");
    assert!(text.contains("AGENTS.local.md"));
}
